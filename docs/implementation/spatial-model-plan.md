# Spatial Model — Implementation Plan

> Discovery doc: [spatial-model-discovery.md](../discovery/spatial-model-discovery.md)

Four phases across two efforts. Each phase is its own execution session.

**Effort 1** — [`spatial-model-effort-1-plan.md`](spatial-model-effort-1-plan.md)
Builds `internal/spatial` as a standalone package with no faction engine dependency. Deliverable: a tested spatial library with interfaces, a full `HybridMap` implementation, and `HexMap`/`GraphMap` stubs.

**Effort 2** — [`spatial-model-effort-2-plan.md`](spatial-model-effort-2-plan.md)
Wires `HybridMap` into the faction engine across three phases: index + scan replacement (Phase 2), validation enforcement (Phase 3), distance-based mechanics and wizard (Phase 4).

<br/>
<br/>

# Cross-Cutting Notes

**`SPATIAL_DATA_DIR`** — a new required env var, separate from `FACTION_DATA_DIR`. Spatial data (Regions, Fragments) is campaign-world data that the Codex will eventually own; faction data (assets, tags, goals, drift costs) remains faction-tool-specific. Before Codex exists, the GM creates `regions.toml` and `fragments.toml` in a directory of their choosing and points `SPATIAL_DATA_DIR` at it.

**Drift cost resolution** — `internal/spatial.Distance()` takes `crossingCost int` (already resolved). The faction engine looks up `rulebook.DriftCost(asset.DriftRating)` and passes the result. The spatial package has no knowledge of drift ratings or `drift_costs.toml`.

**`asset.Location` migration** — `Asset.Location` remains a `string` but its semantics change from a bare display name (`"Tartarus"`) to a Fragment ID (`"tartarus"`). Existing campaign TOML files must be updated manually before Phase 2 can run. No automated migration.

**Spatial index lifetime** — rebuilt at turn start; read-only during resolution. Not persisted. Rebuilding is cheap (one scan of `FactionState`).
