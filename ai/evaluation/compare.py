"""Paired held-out evaluation and deterministic action replays."""
from __future__ import annotations
import argparse
import csv
import json
import os
from pathlib import Path
import time

for variable in ("OMP_NUM_THREADS", "MKL_NUM_THREADS", "OPENBLAS_NUM_THREADS"):
    os.environ[variable] = "1"
import numpy as np
from ai.artifacts import promotion_decision, sha256, verify_model, write_json
from ai.environments.cells import CellsEnv


def verify_replay(replay, binary=None):
    env = CellsEnv(opponents=replay["opponents"], horizon=replay["horizon"], binary=binary)
    try:
        env.reset(seed=replay["seed"])
        result = None
        for action in replay["actions"]:
            result = env.step(np.array(action, dtype=np.float32))
        if result is None or not (result[2] or result[3]) or result[4]["world_hash"] != replay["world_hash"]:
            raise ValueError("Replay does not reproduce its terminal world hash")
        return result[4]["world_hash"]
    finally:
        env.close()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--episodes", type=int, default=100)
    args = parser.parse_args()
    if not 1 <= args.episodes <= 1000:
        parser.error("Episode budget must be 1..1000 per policy")
    manifest = verify_model(args.model)
    if sha256(os.environ["CELLS_HEADLESS"]) != manifest["simulator_sha256"]:
        raise ValueError("Evaluation simulator must match training")
    import torch
    from stable_baselines3 import PPO
    torch.set_num_threads(1)
    torch.set_num_interop_threads(1)
    model = PPO.load(args.model, device="cpu")
    args.output.mkdir(parents=True, exist_ok=False)
    all_rows = []
    for policy in ("random", "heuristic", "ppo"):
        env = CellsEnv(opponents=7, horizon=300)
        try:
            for seed in range(100001, 100001 + args.episodes):
                observation, info = env.reset(seed=seed)
                actions, nanos, rewards = [], [], 0.
                started = time.monotonic()
                while True:
                    if policy == "ppo":
                        before = time.perf_counter_ns()
                        action, _ = model.predict(observation, deterministic=True)
                        cost = time.perf_counter_ns() - before
                        observation, reward, terminated, truncated, info = env.step(action)
                    else:
                        observation, reward, terminated, truncated, info = env.step_reference(policy)
                        cost = env.policy_ns
                    nanos.append(cost)
                    rewards += reward
                    actions.append(env.last_actions["agent_0"])
                    if terminated or truncated:
                        break
                row = {"policy": policy, "seed": seed, **info, "return": rewards, "steps": len(actions),
                       "inference_mean_us": float(np.mean(nanos)/1000),
                       "inference_p95_us": float(np.quantile(nanos, .95)/1000),
                       "elapsed_seconds": time.monotonic()-started}
                all_rows.append(row)
                if seed < 100004:
                    replay = {"seed": seed, "opponents": 7, "horizon": 300,
                              "simulator_sha256": manifest["simulator_sha256"], "actions": actions,
                              "world_hash": info["world_hash"], "policy": policy}
                    verify_replay(replay)
                    write_json(args.output / f"replay-{policy}-{seed}.json", replay)
                if (seed-100000) % 10 == 0:
                    write_json(args.output / "episodes.json", all_rows)
                    print(f"evaluated policy={policy} episodes={seed-100000}/{args.episodes}", flush=True)
        finally:
            env.close()
    write_json(args.output / "episodes.json", all_rows)
    with (args.output / "episodes.csv").open("w") as stream:
        writer = csv.DictWriter(stream, fieldnames=list(all_rows[0]))
        writer.writeheader()
        writer.writerows(all_rows)
    groups = {p: [r for r in all_rows if r["policy"] == p] for p in ("random", "heuristic", "ppo")}
    summary = {"model": manifest, "episode_count": len(all_rows), "opponents": 7, "horizon": 300,
               "policies": {p: {k: {"mean": float(np.mean([r[k] for r in rows])),
                                      "median": float(np.median([r[k] for r in rows]))} for k in
                                      ("mass", "peak", "mass_area", "survival_seconds", "kills", "win", "inference_p95_us")}
                            for p, rows in groups.items()},
               "promotion": promotion_decision(groups["random"], groups["heuristic"], groups["ppo"])}
    write_json(args.output / "summary.json", summary)
    print(json.dumps(summary["promotion"]), flush=True)


if __name__ == "__main__":
    main()
