#!/usr/bin/env bash
#
# Run: ./cmd/scripts/reset-campaign.sh <campaign-id> [scenario]
#
# With a scenario, state/scenarios/<scenario>.toml becomes the new seed, so a
# later reset without one restarts the same scenario.
#
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: reset-campaign.sh <campaign-id> [scenario]" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
STATE_DIR="$REPO_ROOT/campaigns/$1/state"
SEED="$STATE_DIR/state.seed.toml"

if [[ $# -eq 2 ]]; then
  SCENARIO="$STATE_DIR/scenarios/$2.toml"
  if [[ ! -f "$SCENARIO" ]]; then
    echo "error: no scenario file found at $SCENARIO" >&2
    if compgen -G "$STATE_DIR/scenarios/*.toml" > /dev/null; then
      echo "available scenarios:" >&2
      for file in "$STATE_DIR/scenarios/"*.toml; do
        echo "  $(basename "$file" .toml)" >&2
      done
    fi
    exit 1
  fi
  cp "$SCENARIO" "$SEED"
fi

if [[ ! -f "$SEED" ]]; then
  echo "error: no seed file found at $SEED" >&2
  exit 1
fi

cp "$SEED" "$STATE_DIR/state.toml"
> "$STATE_DIR/history.jsonl"
rm -f "$STATE_DIR/logs/"*.log

if [[ $# -eq 2 ]]; then
  echo "$1 reset to cycle 1 with scenario $2"
else
  echo "$1 reset to cycle 1"
fi
