# Decisions Log — Spatial Engine

A record of key decisions made during spatial engine development.

## Index

| Section | Decisions | Topic |
|---------|-----------|-------|
| [feat/spatial-effort-1](#featspatial-effort-1) | 1–7 | Sentinel errors, Fragment field shape, boundary validation, interface asserts |

<br />

### feat/spatial-effort-1

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Sentinel errors in spatial package: `ErrUnknownFragment`, `ErrNoPath`, `ErrNotImplemented`, `ErrInvalidCost`; wrap with `fmt.Errorf("%w: ...", ...)` | Spatial is positioned as a future shared library; callers (faction engine, eventual Codex) need `errors.Is` discrimination to handle missing-fragment vs no-path vs not-yet-implemented distinctly. String matching couples callers to error wording |
| 2 | `Fragment` locator fields (`id`, `name`, `techLevel`, `population`) are unexported and reached only through `Location` interface methods; `Region` and `Hex` remain exported | Two-tier surface: interface-backed identity goes through methods (real polymorphism across HybridMap/HexMap/GraphMap); spatial-mechanic fields stay direct-readable for the engine's hot-path scans. Resolves the `Fragment*`/`Frag*` prefix collision that arose from dual-export and makes the type immutable from outside the package |
| 3 | Compile-time interface satisfaction asserts (`var _ SpatialMap = (*HybridMap)(nil)`, etc.) declared in `spatial.go` | Future refactors of `SpatialMap`/`Location` interfaces must fail at compile time at the implementation site, not at downstream call sites. Free safety net for a shared library |
| 4 | `LoadHybrid` validates every boundary: `From` must be in declaring region's hexes; `ToRegion` must exist; `To` must be in target region's hexes | Without this, a typo in `regions.toml` produces a graph node with zero outgoing edges and Dijkstra silently returns "no path" instead of "your config is broken." Validation-at-boundary is mandatory for a library other tools will write authored data against |
| 5 | `Distance` rejects negative `crossingCost` with `ErrInvalidCost` | Negative edge weights violate Dijkstra's correctness invariant; cheaper to fail loudly at the call site than to debug a wrong path silently. Faction engine resolves `crossingCost` from `rulebook.DriftCost(...)`, so this catches bad TOML cost data before it corrupts results |
| 6 | `HybridMap.Location` returns explicit untyped `nil` on miss instead of returning the typed-nil `*Fragment` | Go's interface-nil gotcha: a `(Location, *Fragment, nil)` interface value is not equal to untyped `nil` at the call site; callers writing `if loc != nil` would be misled. Caught by `TestHybridMap_Location/unknown_id` after Location() was tested directly for the first time |
| 7 | `pqItem.idx` field removed; priority queue's `Swap`/`Push` no longer track index position | The field exists in canonical heap implementations to support `heap.Fix`/`heap.Remove`; this Dijkstra uses neither (relies on the `current.cost > dist[current.node]` skip-stale pattern instead). Dead state was misleading future readers about which heap operations were planned |
