# Planned Work — Spatial Engine

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped — at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| # | Item | Size | Status | Trigger |
|---|------|------|--------|---------|

---
<br/>

## Deferred — Major

Items that will eventually warrant a full initiative entry. Each moves to Planned Initiatives when scoped.

| # | Item | Trigger |
|---|------|---------|

---
<br/>

## Deferred — Minor

Small, targeted fixes. No write-up needed — tracked here until scheduled.

| # | Item | Detail |
|---|------|--------|
| 1 | `internal/spatial` reverse-boundary index | `HybridMap.neighbors` step 4 scans every region's boundaries on each call (O(R × B per call)). Trivial for hand-authored campaign data; revisit if Codex generates large procedural maps and Dijkstra hot-loops surface in profiling |
| 2 | `Region.Hexes` storage shape | `map[HexCoord]bool` carries a 1-byte value per entry; `map[HexCoord]struct{}` is zero-byte. Pure micro-opt; not worth churning until map sizes are known |
| 3 | `internal/spatial` package documentation (`doc.go`) | TOML schema reference and concurrency model (read-only after `LoadHybrid`) currently live only in the implementation plan. Defer until Effort 2 lands and the public surface settles |

---
<br />
<br />

# Planned Initiatives

This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.
