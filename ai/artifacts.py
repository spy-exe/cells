"""Versioned, trusted local model artifacts; never accepts uploaded models."""
from __future__ import annotations
import hashlib
import json
from pathlib import Path
import subprocess


def sha256(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def write_json(path, value):
    path = Path(path)
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(json.dumps(value, indent=2, allow_nan=False) + "\n")
    temporary.replace(path)


def source_identity(root):
    root = Path(root)
    if (root / "source.json").exists():
        identity = json.loads((root / "source.json").read_text())
    else:
        identity = {"revision": subprocess.check_output(
            ["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip(),
            "dirty": bool(subprocess.check_output(
                ["git", "-C", str(root), "status", "--porcelain"], text=True))}
    if identity.get("dirty") or len(identity.get("revision", "")) != 40:
        raise ValueError("Training requires a clean, committed source revision")
    return identity


def verify_model(path):
    path = Path(path)
    manifest = json.loads((path.parent / "manifest.json").read_text())
    if manifest.get("model_sha256") != sha256(path):
        raise ValueError("Model digest differs from its manifest")
    if manifest.get("observation") != 80 or manifest.get("action") != 4:
        raise ValueError("Incompatible policy spaces")
    if manifest.get("algorithm") != "PPO" or manifest.get("protocol") != 1:
        raise ValueError("Unsupported model contract")
    return manifest


def promotion_decision(random_rows, heuristic_rows, learned_rows, inference_limit_us=2000):
    """A lab gate is evidence for further testing, never production authorization."""
    import numpy as np
    if len(learned_rows) < 100 or len(heuristic_rows) != len(learned_rows):
        return {"lab_candidate": False, "production_promoted": False,
                "reason": "At least 100 paired held-out episodes are required"}
    by_seed = {r["seed"]: r for r in heuristic_rows}
    if len(by_seed) != len(learned_rows) or {r["seed"] for r in learned_rows} != set(by_seed):
        raise ValueError("Evaluation seeds must be paired and unique")
    learned = np.array([r["mass_area"] for r in learned_rows])
    reference = np.array([by_seed[r["seed"]]["mass_area"] for r in learned_rows])
    rng = np.random.default_rng(20261016)
    samples = rng.integers(0, len(learned), size=(2000, len(learned)))
    ratios = learned[samples].mean(axis=1) / np.maximum(reference[samples].mean(axis=1), 1e-9)
    lower, upper = np.quantile(ratios, [0.025, 0.975])
    def mean(rows, key):
        return float(np.mean([r[key] for r in rows]))
    candidate = bool(lower > 1.05 and
                     mean(learned_rows, "survival_seconds") >= .95 * mean(heuristic_rows, "survival_seconds") and
                     mean(learned_rows, "kills") >= mean(heuristic_rows, "kills") and
                     mean(learned_rows, "win") >= mean(heuristic_rows, "win") and
                     max(r["inference_p95_us"] for r in learned_rows) <= inference_limit_us)
    return {"lab_candidate": candidate, "production_promoted": False,
            "auc_ratio": float(learned.mean() / max(reference.mean(), 1e-9)),
            "paired_bootstrap_95": [float(lower), float(upper)],
            "reason": ("Lab gate passed; production-size evaluation and native inference still required" if candidate else
                       "Learned policy did not satisfy all growth, survival, combat and inference gates; keep heuristics"),
            "random_episodes": len(random_rows), "episodes_per_policy": len(learned_rows)}
