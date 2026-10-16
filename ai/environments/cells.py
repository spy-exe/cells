"""Gymnasium/PettingZoo adapters for the authoritative native Cells simulator."""
from __future__ import annotations

import json
import os
from pathlib import Path
import select
import subprocess
from typing import Any

import gymnasium as gym
from gymnasium import spaces
import numpy as np
from pettingzoo import ParallelEnv


class NativeBridge:
    def __init__(self, binary: str | None = None):
        path = binary or os.environ.get(
            "CELLS_HEADLESS", str(Path(__file__).resolve().parents[2] / "bin/cells-headless")
        )
        if not Path(path).is_file():
            raise FileNotFoundError("Build cells-headless and set CELLS_HEADLESS to its path")
        self.process = subprocess.Popen(
            [path], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
            stderr=subprocess.PIPE, text=True, bufsize=1,
            env={**os.environ, "GOMAXPROCS": "1"},
        )

    def request(self, payload: dict[str, Any]) -> dict[str, Any]:
        if self.process.poll() is not None:
            raise RuntimeError("Native simulator exited")
        assert self.process.stdin is not None and self.process.stdout is not None
        self.process.stdin.write(json.dumps(payload, allow_nan=False) + "\n")
        self.process.stdin.flush()
        if not select.select([self.process.stdout], [], [], 20)[0]:
            raise TimeoutError("Native simulator response exceeded 20 seconds")
        line = self.process.stdout.readline()
        if not line:
            raise RuntimeError("Native simulator closed its output")
        reply = json.loads(line)
        if reply.get("protocol") != 1:
            raise RuntimeError("Unsupported native simulator protocol")
        if reply.get("error"):
            raise RuntimeError(reply["error"])
        return reply

    def close(self):
        if self.process.poll() is None:
            try:
                self.request({"op": "close"})
                self.process.wait(timeout=2)
            except (OSError, RuntimeError, TimeoutError, subprocess.TimeoutExpired):
                self.process.kill()
                self.process.wait(timeout=2)
        for stream in (self.process.stdin, self.process.stdout, self.process.stderr):
            if stream:
                stream.close()


def action_values(action: np.ndarray | list[float]) -> list[float]:
    a = np.asarray(action, dtype=np.float32)
    if a.shape != (4,) or not np.isfinite(a).all() or (np.abs(a) > 1).any():
        raise ValueError("Action must contain four finite values in [-1, 1]")
    return a.tolist()


def info_for(result: dict[str, Any], state: dict[str, Any], seed: int) -> dict[str, Any]:
    return {**state["metrics"], "seed": seed, "tick": result["tick"],
            "world_hash": result.get("hash", "")}


class CellsEnv(gym.Env):
    metadata = {"render_modes": [], "name": "Cells-v0"}

    def __init__(self, opponents: int = 7, horizon: int = 300, binary: str | None = None):
        super().__init__()
        self.action_space = spaces.Box(-1, 1, shape=(4,), dtype=np.float32)
        self.observation_space = spaces.Box(-1, 1, shape=(80,), dtype=np.float32)
        self.bridge = NativeBridge(binary)
        self.horizon = horizon
        self.set_opponents(opponents)
        self.episode_seed = 0
        self.episode_opponents = opponents
        self.policy_ns = 0

    def set_opponents(self, opponents: int):
        if not 0 <= opponents <= 12:
            raise ValueError("Opponent count must be between zero and twelve")
        self.opponents = opponents

    def reset(self, *, seed=None, options=None):
        super().reset(seed=seed)
        self.episode_seed = int(seed if seed is not None else self.np_random.integers(0, 100000))
        self.episode_opponents = self.opponents
        reply = self.bridge.request({"op": "reset", "seed": self.episode_seed,
                                     "agents": 1, "opponents": self.opponents,
                                     "horizon": self.horizon})
        result = reply["result"]
        state = result["agents"]["agent_0"]
        return np.array(state["observation"], dtype=np.float32), {**info_for(result, state, self.episode_seed), "opponents": self.episode_opponents}

    def _transition(self, payload):
        reply = self.bridge.request(payload)
        self.policy_ns = reply.get("policy_ns", 0)
        result = reply["result"]
        state = result["agents"]["agent_0"]
        self.last_actions = result.get("actions", {})
        return (np.array(state["observation"], dtype=np.float32), float(state["reward"]),
                bool(state["terminated"]), bool(state["truncated"]),
                {**info_for(result, state, self.episode_seed), "opponents": self.episode_opponents})

    def step(self, action):
        return self._transition({"op": "step", "actions": {"agent_0": action_values(action)}})

    def step_reference(self, policy: str):
        return self._transition({"op": "step", "policy": policy})

    def close(self):
        self.bridge.close()


class CellsParallelEnv(ParallelEnv):
    metadata = {"render_modes": [], "name": "cells_parallel_v0", "is_parallelizable": True}

    def __init__(self, agents: int = 2, opponents: int = 7, horizon: int = 300,
                 binary: str | None = None):
        if not 1 <= agents <= 4:
            raise ValueError("One to four controlled agents required")
        self.possible_agents = [f"agent_{i}" for i in range(agents)]
        self.agents = []
        self.observation_spaces = {a: spaces.Box(-1, 1, shape=(80,), dtype=np.float32)
                                   for a in self.possible_agents}
        self.action_spaces = {a: spaces.Box(-1, 1, shape=(4,), dtype=np.float32)
                              for a in self.possible_agents}
        self.bridge = NativeBridge(binary)
        self.opponents, self.horizon = opponents, horizon
        self.episode_seed = 0
        self.np_random = np.random.default_rng()

    def observation_space(self, agent):
        return self.observation_spaces[agent]

    def action_space(self, agent):
        return self.action_spaces[agent]

    def reset(self, seed=None, options=None):
        if seed is not None:
            self.np_random = np.random.default_rng(seed)
        self.episode_seed = int(seed if seed is not None else self.np_random.integers(0, 100000))
        result = self.bridge.request({"op": "reset", "seed": self.episode_seed,
                                      "agents": len(self.possible_agents),
                                      "opponents": self.opponents,
                                      "horizon": self.horizon})["result"]
        self.agents = self.possible_agents.copy()
        return ({a: np.array(result["agents"][a]["observation"], dtype=np.float32) for a in self.agents},
                {a: info_for(result, result["agents"][a], self.episode_seed) for a in self.agents})

    def step(self, actions):
        if not self.agents:
            return {}, {}, {}, {}, {}
        if set(actions) != set(self.agents):
            raise ValueError("Exactly one action per live agent is required")
        result = self.bridge.request({"op": "step", "actions":
                                      {a: action_values(actions[a]) for a in self.agents}})["result"]
        previous = self.agents.copy()
        states = result["agents"]
        terms = {a: bool(states[a]["terminated"]) for a in previous}
        truncs = {a: bool(states[a]["truncated"]) for a in previous}
        observations = {a: np.array(states[a]["observation"], dtype=np.float32) for a in previous}
        rewards = {a: float(states[a]["reward"]) for a in previous}
        infos = {a: info_for(result, states[a], self.episode_seed) for a in previous}
        self.agents = [a for a in previous if not terms[a] and not truncs[a]]
        return observations, rewards, terms, truncs, infos

    def close(self):
        self.bridge.close()
