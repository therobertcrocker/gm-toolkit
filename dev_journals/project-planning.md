# GM Toolkit — Project Plan

Tracks the products within the gm-toolkit suite. Each product keeps its own tracker and initiatives under `dev_journals/<product>/`.

**Initiative IDs share one counter across all products.** A new row takes the next free number for its prefix across every tracker and every `_completed/` folder, so no two products ever hold the same ID.

<br/>

## Products

| # | Product | Status | Trigger | Tracker |
|---|---------|--------|---------|---------|
| 1 | `campaigns` | In progress | — | [`campaigns/planned-work.md`](campaigns/planned-work.md) |
| 2 | `faction-manager` | In progress | — | [`faction-manager/planned-work.md`](faction-manager/planned-work.md) |
| 3 | `codex` | Planned | `internal/spatial` package established; spatial model complete | [`codex/planned-work.md`](codex/planned-work.md) |

<br/>
<br/>

# Products

<br/>

## Campaigns

The umbrella product: the campaign as a whole, which every other product serves. Holds campaign-level tools that are not part of another product, such as the PF2e encounter simulator (`F.041`).

<br/>

## Codex

A campaign knowledge-base and wiki for GMs. Stores structured information about characters, Fragments, factions, game events, and world lore. Designed as the source of truth for spatial and world data shared with the faction tool.

**Package layout:** `internal/codex`, `cmd/codex`

**Shared dependencies:** `internal/spatial` — the Codex owns the canonical Fragment and Region TOML files; the faction tool reads from the same files via `SPATIAL_DATA_DIR`.

**Trigger:** Pick up once `internal/spatial` is established and the shared TOML schema is proven in the faction tool.

**Seed:** [`codex-assistant-seed.md`](codex/initiatives/seeds/codex-assistant-seed.md) — the GM Assistant concept the Codex is the first step of. Awareness, not a plan.
