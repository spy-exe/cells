"""Render figures from recorded episodes, without smoothing away poor results."""
import argparse
import csv
import json
from pathlib import Path
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--evaluation", type=Path, required=True)
    parser.add_argument("--monitor", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    rows = json.loads((args.evaluation / "episodes.json").read_text())
    with args.monitor.open() as stream:
        next(stream)
        training = list(csv.DictReader(stream))
    args.output.mkdir(parents=True, exist_ok=True)
    plt.rcParams.update({"font.family": "DejaVu Sans", "axes.spines.top": False,
                         "axes.spines.right": False, "axes.labelcolor": "#17251c",
                         "text.color": "#17251c", "axes.edgecolor": "#a5b19c"})
    fig, axes = plt.subplots(1, 3, figsize=(14, 4.5), layout="constrained")
    fig.set_facecolor("#f4f7ef")
    for ax in axes:
        ax.set_facecolor("#f4f7ef")
    steps = np.cumsum([int(r["l"]) for r in training])
    reward = np.array([float(r["r"]) for r in training])
    axes[0].scatter(steps, reward, s=9, alpha=.3, color="#557f21", label="Each finished episode")
    if len(reward) >= 10:
        axes[0].plot(steps[9:], np.convolve(reward, np.ones(10)/10, mode="valid"),
                     color="#254b18", linewidth=1.5, label="10-episode mean")
    for x in (8192, 16384):
        axes[0].axvline(x, color="#a5b19c", linestyle="--", linewidth=1)
    axes[0].set(title="PPO training · seed17", xlabel="Native simulation decisions", ylabel="Episode reward")
    axes[0].legend(fontsize=7)
    policies = ("random", "heuristic", "ppo")
    colors = ("#a5b19c", "#557f21", "#8c91d6")
    rng = np.random.default_rng(20261016)
    for ax, metric, title, label in (
        (axes[1], "mass_area", "Growth and survival", "Mass × seconds · episode area"),
        (axes[2], "survival_seconds", "Survival", "Seconds · maximum60")):
        means, errors = [], [[], []]
        for policy in policies:
            values = np.array([r[metric] for r in rows if r["policy"] == policy])
            mean = values.mean()
            bootstrap = values[rng.integers(0, len(values), size=(2000, len(values)))].mean(axis=1)
            lo, hi = np.quantile(bootstrap, [.025, .975])
            means.append(mean)
            errors[0].append(max(0, mean-lo)); errors[1].append(max(0, hi-mean))
        ax.bar(policies, means, yerr=errors, capsize=4, color=colors, width=.6)
        ax.set(title=title, ylabel=label)
    fig.suptitle("Cells · real bounded experiment · 100 held-out seeds per policy", fontsize=13)
    fig.savefig(args.output / "training-comparison.png", dpi=160)
    fig.savefig(args.output / "training-comparison.svg")
    plt.close(fig)


if __name__ == "__main__":
    main()
