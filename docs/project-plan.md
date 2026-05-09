# GM Toolkit — Project Plan

Tracks planned tools within the gm-toolkit suite. Each tool is its own binary (`cmd/<tool>`) with domain logic under `internal/<tool>`. Feature-level planning for individual tools lives in their own `planned-work.md` within the relevant dev journal.

<br/>

## Tools

| # | Tool | Status | Trigger |
|---|------|--------|---------|
| 1 | `faction-manager` | In progress | — |
| 2 | `codex` | Planned | `internal/spatial` package established; spatial model complete |

<br/>
<br/>

# Planned Tools

<br/>

## Codex

A campaign knowledge-base and wiki for GMs. Stores structured information about characters, Fragments, factions, game events, and world lore. Designed as the source of truth for spatial and world data shared with the faction tool.

**Package layout:** `internal/codex`, `cmd/codex`

**Shared dependencies:** `internal/spatial` — the Codex owns the canonical Fragment and Region TOML files; the faction tool reads from the same files via `SPATIAL_DATA_DIR`.

**Trigger:** Pick up once `internal/spatial` is established and the shared TOML schema is proven in the faction tool.
