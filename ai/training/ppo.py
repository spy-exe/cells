"""Bounded CPU PPO experiment with recoverable, fingerprinted checkpoints."""
from __future__ import annotations
import argparse
from datetime import datetime, timezone
import importlib.metadata
import os
from pathlib import Path
import platform
import random
import subprocess
import time

for variable in ("OMP_NUM_THREADS", "MKL_NUM_THREADS", "OPENBLAS_NUM_THREADS"):
    os.environ[variable] = "1"

import numpy as np
import torch
from stable_baselines3 import PPO
from stable_baselines3.common.callbacks import BaseCallback
from stable_baselines3.common.monitor import Monitor
from ai.artifacts import sha256, source_identity, verify_model, write_json
from ai.environments.cells import CellsEnv


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--steps", type=int, default=32768)
    parser.add_argument("--seed", type=int, default=17)
    parser.add_argument("--resume", type=Path)
    args = parser.parse_args()
    if not 512 <= args.steps <= 131072 or args.steps % 512:
        parser.error("Use 512..131072 steps in multiples of 512")
    identity = source_identity(Path(__file__).resolve().parents[2])
    args.output.mkdir(parents=True, exist_ok=False)
    torch.set_num_threads(1)
    torch.set_num_interop_threads(1)
    torch.use_deterministic_algorithms(True)
    random.seed(args.seed)
    np.random.seed(args.seed)
    binary = os.environ["CELLS_HEADLESS"]
    started = time.monotonic()
    base = {**identity, "algorithm": "PPO", "protocol": 1, "observation": 80, "action": 4,
            "seed": args.seed, "simulator_sha256": sha256(binary), "python": platform.python_version(),
            "dependencies": {name: importlib.metadata.version(name) for name in
                             ("torch", "stable-baselines3", "gymnasium", "pettingzoo", "numpy")},
            "arena": {"size": 3200, "food": 1200, "viruses": 12, "horizon": 300},
            "curriculum": [{"until": 8192, "opponents": 0}, {"until": 16384, "opponents": 3},
                           {"until": None, "opponents": 7}],
            "started_utc": datetime.now(timezone.utc).isoformat(), "production_promoted": False}
    env = CellsEnv(opponents=0, horizon=300)
    monitored = Monitor(env, str(args.output / "episodes.csv"),
                        info_keywords=("mass", "peak", "mass_area", "survival_seconds", "kills", "win", "opponents", "seed"))
    def save(model, directory, complete=False):
        directory.mkdir(parents=True, exist_ok=False)
        model.save(directory / "model.zip")
        write_json(directory / "manifest.json", {**base, "model_sha256": sha256(directory / "model.zip"),
                   "num_timesteps": model.num_timesteps, "updates": model._n_updates,
                   "complete": complete, "elapsed_seconds": time.monotonic() - started,
                   "hyperparameters": {key: getattr(model, key) for key in
                     ("n_steps", "batch_size", "n_epochs", "gamma", "gae_lambda", "ent_coef", "max_grad_norm")}})
    class Progress(BaseCallback):
        def _on_step(self):
            steps = self.num_timesteps
            env.set_opponents(0 if steps < 8192 else 3 if steps < 16384 else 7)
            if steps % 4096 == 0:
                save(self.model, args.output / f"checkpoint-{steps:06d}")
                print(f"checkpoint steps={steps} elapsed={time.monotonic()-started:.1f}s opponents={env.opponents}", flush=True)
            return True
    try:
        if args.resume:
            parent = verify_model(args.resume)
            if parent["simulator_sha256"] != base["simulator_sha256"] or parent["revision"] != identity["revision"]:
                raise ValueError("Resume requires the original simulator and source revision")
            base["parent_model_sha256"] = parent["model_sha256"]
            base["parent_timesteps"] = parent["num_timesteps"]
            model = PPO.load(args.resume, env=monitored, device="cpu")
            env.set_opponents(0 if model.num_timesteps < 8192 else 3 if model.num_timesteps < 16384 else 7)
        else:
            model = PPO("MlpPolicy", monitored, device="cpu", seed=args.seed, learning_rate=3e-4,
                        n_steps=512, batch_size=64, n_epochs=4, gamma=.995, gae_lambda=.95,
                        clip_range=.2, ent_coef=.005, max_grad_norm=.5,
                        policy_kwargs={"net_arch": {"pi": [64, 64], "vf": [64, 64]}, "log_std_init": -.7}, verbose=0)
        write_json(args.output / "experiment.json", base)
        (args.output / "requirements-resolved.txt").write_text(subprocess.check_output(
            [os.sys.executable, "-m", "pip", "freeze"], text=True))
        model.learn(total_timesteps=args.steps, callback=Progress(), reset_num_timesteps=not bool(args.resume))
        save(model, args.output / "final", complete=True)
        print(f"complete steps={model.num_timesteps} elapsed={time.monotonic()-started:.1f}s", flush=True)
    finally:
        monitored.close()


if __name__ == "__main__":
    main()
