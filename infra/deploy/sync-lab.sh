#!/usr/bin/env bash
set -euo pipefail
[[ "$(hostname)" == pve ]]
pct config 221 | grep -Fx 'hostname: cells-ai-lab-01' >/dev/null
repo=$(cd "$(dirname "$0")/../.." && pwd)
archive=$(mktemp /tmp/cells-lab-source.XXXXXX.tar.gz)
trap 'rm -f "$archive"' EXIT
tar --exclude=.git --exclude=node_modules --exclude=.runtime --exclude=dist \
 --exclude=artifacts --exclude=.env --exclude=.venv --exclude=coverage \
 --exclude='test-results*' --exclude=playwright-report --exclude=release \
 --exclude='*.dump' --exclude=__pycache__ -czf "$archive" -C "$repo" .
pct push 221 "$archive" /tmp/cells-lab-source.tar.gz
pct exec 221 -- bash -c 'mkdir -p /opt/cells-ai/src; tar -xzf /tmp/cells-lab-source.tar.gz -C /opt/cells-ai/src; rm /tmp/cells-lab-source.tar.gz'
source_meta=$(python3 - "$repo" <<'PY'
import json, subprocess, sys
repo=sys.argv[1]
revision=subprocess.check_output(['git','-C',repo,'rev-parse','HEAD'],text=True).strip()
dirty=bool(subprocess.check_output(['git','-C',repo,'status','--porcelain'],text=True))
print(json.dumps({'revision':revision,'dirty':dirty}))
PY
)
printf '%s\n' "$source_meta" | pct exec 221 -- tee /opt/cells-ai/src/source.json
