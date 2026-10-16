# Cells learning laboratory

Production uses five **programmed utility policies**. The laboratory trains PPO
against those policies through the authoritative Go simulator. Training never runs
in the production service. A saved model is an experiment, not a production bot.

The [first recorded experiment](../docs/ai/2026-10-16/README.md) trained33,280steps
including a real checkpoint resume, then evaluated300episodes. PPO lost to the
heuristic; no learned model was promoted. Figures use optional
`pip install -r ai/requirements-reports.txt`, followed by `python -m ai.evaluation.plot`.

## Reproduce

Python3.13, Go from go.mod. Install the official CPU Torch wheel first:

```sh
python3 -m venv .venv
.venv/bin/pip install torch==2.14.1+cpu --index-url https://download.pytorch.org/whl/cpu
.venv/bin/pip install -r ai/requirements-training.txt
mkdir -p bin
go build -trimpath -o bin/cells-headless ./apps/game-server/cmd/headless
export CELLS_HEADLESS="$PWD/bin/cells-headless"
.venv/bin/python -m unittest discover -s ai/tests -v
# Source must be committed and clean before training.
.venv/bin/python -m ai.training.ppo --output /tmp/cells-ppo-17 --steps 32768
.venv/bin/python -m ai.training.ppo --output /tmp/cells-ppo-17-resume --steps 512 --resume /tmp/cells-ppo-17/final/model.zip
.venv/bin/python -m ai.evaluation.compare --model /tmp/cells-ppo-17-resume/final/model.zip --output /tmp/cells-eval-17 --episodes 100
```

Training uses one CPU thread and seeds below100000. Evaluation uses seeds
100001..100100 with identical initial worlds for each policy. Six native30Hz ticks
per decision, 300decision/60second maximum episodes. See ADR0004 for exact features,
reward and bounded arena. Gymnasium and PettingZoo tests execute the native process.
No human chat or player data are used.

Every4096steps, a model and manifest record source revision, simulator digest,
dependency versions, optimizer state and timestep. Resume verifies digest, spaces,
revision and simulator. The environment resets on resume: this restores learning
state, not an identical interrupted rollout. Use only locally trusted SB3 files;
its serialized models are not an upload format for untrusted users.

Evaluation writes per-episode JSON/CSV, summary, and three command replays per
policy. Replays must reproduce the exact final world hash. Compute-only policy
latency excludes simulation/stdio and has different implementations (Go utility,
Python/Torch PPO); report that limitation. The paired bootstrap growth gate also
requires survival, combat and inference limits. Even passing that lab gate leaves
`production_promoted=false`: production-size testing and native inference are
separate requirements. A losing model stays an openly documented experiment.

## Personal laboratory operations

CT221 `cells-ai-lab-01` is isolated from CT220 production. Source sync:
`bash infra/deploy/sync-lab.sh` (personal pve host only). Build the native binary in
the Cells build area, copy to `/opt/cells-ai/bin/cells-headless`. Sync only while no
experiment is running. Source and venv are root-owned; outputs belong to cells-ai
under `/var/lib/cells-ai/runs`. Start the Python commands above with systemd-run,
User=cells-ai, WorkingDirectory=/opt/cells-ai/src, CPUQuota=65%, MemoryMax=1100M,
RuntimeMaxSec=1200, NoNewPrivileges=yes, ProtectSystem=strict,
ReadWritePaths=/var/lib/cells-ai/runs, PrivateTmp=yes and all thread variables1.
Set CELLS_HEADLESS explicitly and PYTHONDONTWRITEBYTECODE=1. Inspect journal and
exit status. Stop CT221 when idle. Do not modify production's CPU limits or run
large training there. A timeout preserves completed checkpoints; resume from a
verified checkpoint with its matching source. Never present a timeout as success.
