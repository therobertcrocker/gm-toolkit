#!/usr/bin/env bash
#
# Run: ./cmd/scripts/reset-campaign.sh <campaign-id>
#
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: reset-campaign.sh <campaign-id>" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
STATE_DIR="$REPO_ROOT/campaigns/$1/state"
SEED="$STATE_DIR/state.seed.toml"

if [[ ! -f "$SEED" ]]; then
  echo "error: no seed file found at $SEED" >&2
  exit 1
fi

cp "$SEED" "$STATE_DIR/state.toml"
> "$STATE_DIR/history.jsonl"
rm -f "$STATE_DIR/logs/"*.log

echo "$1 reset to cycle 1"
