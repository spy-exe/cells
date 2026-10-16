#!/usr/bin/env bash
# A running instance must already occupy CELLS_ADDR; a second one must fail, not exit successfully.
set -uo pipefail
log=$(mktemp)
trap 'rm -f "$log"' EXIT
CELLS_CANDIDATE=1 timeout 10 "${CELLS_TEST_BINARY:-release/cells}" > "$log" 2>&1
status=$?
if [[ "$status" -ne 1 ]] || ! grep -q 'server stopped unexpectedly' "$log"; then
  echo 'Expected a nonzero bind failure from the second server.' >&2
  exit 1
fi
echo 'Duplicate listener rejected with process failure.'
