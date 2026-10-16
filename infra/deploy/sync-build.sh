#!/usr/bin/env bash
set -euo pipefail
repo=$(cd "$(dirname "$0")/../.." && pwd)
archive=$(mktemp /tmp/cells-source.XXXXXX.tar.gz)
trap 'rm -f "$archive"' EXIT
tar --exclude=.git --exclude=node_modules --exclude=.runtime --exclude=dist --exclude=artifacts --exclude=.env --exclude=.venv --exclude=coverage --exclude='test-results*' --exclude=playwright-report -czf "$archive" -C "$repo" .
pct push 220 "$archive" /tmp/cells-source.tar.gz
pct exec 220 -- bash -c 'mkdir -p /opt/cells/build; tar -xzf /tmp/cells-source.tar.gz -C /opt/cells/build; rm /tmp/cells-source.tar.gz'
