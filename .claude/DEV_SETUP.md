# Development Setup

## Running the binary

Work from `cmd/faction-manager/` as the working directory:

```bash
cd cmd/faction-manager
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager <command>
```

- `bin/` holds the compiled binary
- `campaigns/` sits alongside `bin/` and holds campaign state files
- `FACTION_DATA_DIR` must point to `internal/faction/data` (contains all asset, tag, and goal TOML files)

## Test campaign

A test campaign with three pre-populated factions lives at:

```
cmd/faction-manager/campaigns/test/faction_state.toml
```

Run against it with `--campaign test`.

## Running tests

```bash
go test ./...
```

Tests live in:
- `internal/faction/loader/` — loader smoke test, `parseDice` table-driven, duplicate asset ID detection
- `internal/faction/state/` — `FactionState` TOML round-trip, missing file handling
