# Spatial Map CLI — Discovery

`F-012`. Sibling of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md); prerequisite to `F-005.2` (tui-manage) because Manage needs an authored spatial model to render. The CLI is content-authoring tooling — it produces the `regions.toml` + `worlds.toml` inputs that `internal/spatial.LoadRegionMap` consumes. It does not touch engine code.

## Problem

Authoring `regions.toml` and `worlds.toml` by hand is intractable beyond toy maps. The current schema expects axial hex coordinates per region cell (`hexes [][2]int`), explicit per-region boundary records (`from_q`, `from_r`, `to_region`, `to_q`, `to_r`), and worlds pinned to `(region, hex_q, hex_r)`. A modest sector — four regions, ~30 hexes each, a handful of worlds, a few warps — requires the GM to keep axial coordinates straight in their head, hand-enumerate every cross-region boundary cell, and maintain consistency across two files. The TUI cannot ship Manage against an unauthored or fragile spatial model.

The intermediary file approach: let the GM draw the map in an ASCII layout (regions as letters, worlds as digit markers) and supply a small TOML data file (region names, world metadata, non-adjacent warps). The CLI consumes both and emits the canonical `regions.toml` + `worlds.toml` that `LoadRegionMap` already understands. Adjacency-based boundaries are derived automatically from the ASCII grid; the GM never thinks in axial coordinates.

## Design Summary

The CLI is a **one-way builder**: it consumes two GM-authored intermediary files — an ASCII layout and a TOML data file — and emits the canonical `regions.toml` + `worlds.toml` that `internal/spatial.LoadRegionMap` already understands. No reverse direction. Canonical files are machine-output; the source of truth is the intermediary pair.

The layout file is a grid of single-character cells separated by single spaces, with odd rows indented one space to render the pointy-top hex stagger. Cell alphabet: `A`-`Z` for region membership, `1`-`9` then `a`-`z` for in-band world markers (35-world cap per map), `.` for empty hexes. The TOML data file carries region names, world metadata (id, name, tech_level, population, optional `region` override), and `[[warps]]` for non-adjacent connections.

The CLI is invoked as `gm-toolkit spatial build [--campaign <id>] [--replace]`. By convention, source files live at `<campaign>/spatial/source/layout.txt` and `<campaign>/spatial/source/data.toml`; outputs go to `<campaign>/spatial/regions.toml` and `<campaign>/spatial/worlds.toml`. No path flags — the locations are fixed.

The CLI derives **dense adjacency boundaries** automatically from the grid (every shared hex-edge between regions becomes one canonical boundary entry on the alphabetically-first region's side). World markers are placed by glyph match between layout and data file; their region is auto-inferred from neighboring cells, with an optional explicit override for ambiguous positions. Warps are authored explicitly in the data file using grid `{ region, row, col }` endpoints.

The adjacency-vs-warp **cost distinction** is *not* implemented in this initiative — today's engine treats all `Region.Boundaries` uniformly. The CLI emits warps and adjacency edges into the same canonical list; a follow-up engine refactor (filed in Out of Scope) will introduce per-edge-type costs and the CLI will gain a matching output split at that point.

## Decisions Ratified in Discovery

1. **One-way conversion only.** CLI consumes intermediary files and emits canonical files; no reverse. Canonical `regions.toml`/`worlds.toml` are treated as machine-output. Rationale: YAGNI — round-trip adds a losslessness constraint with no current consumer. If a future tool generates canonical files directly, we'll revisit.
2. **Worlds are in-band in the layout file.** Single-char markers placed at the world's hex cell. Rationale: the visual map showing world-in-region context is the core authoring affordance. Out-of-band positioning loses what makes the layout file worth having.
3. **Cell alphabet:** `A`-`Z` = region membership; `1`-`9` then `a`-`z` = world markers (35 capacity per map); `.` = empty hex. Rationale: single-char cells keep grid alignment trivial; case partitioning gives 26 regions and 35 worlds without escaping or multi-char cells. If a map needs more than 35 worlds we revisit (likely splitting into multiple maps before that).
4. **Layout file mechanics.** Cells separated by single spaces. Odd rows indented one leading space to render the pointy-top hex stagger visually. Trailing whitespace on rows is tolerated. Lines beginning with `#` are comments and ignored by the parser. No required preamble / header — the file is just the grid (plus comments).
5. **World-region assignment: auto-infer with explicit override.** The CLI scans the 6 hex neighbors of each world marker. If exactly one distinct region letter appears among non-empty neighbors, that region is auto-assigned. If multiple region candidates appear OR all neighbors are empty, the world's TOML entry must include `region = "X"` and the CLI errors with the candidate list if missing. Rationale: zero friction in the common case (worlds sit inside their region); explicit only when the layout itself is ambiguous.
6. **Dense adjacency boundaries derived from the grid.** For every pair of touching hexes from different regions, the CLI emits one `[[region.boundary]]` entry on the alphabetically-first region's side (canonical, deduplicated). Rationale: matches the "A↔B and B↔C are auto-generated from the grid" framing; GM authors the layout, the CLI handles boundary enumeration.
7. **Adjacency-vs-warp cost distinction is deferred (Option C).** Today, the engine prices all `Region.Boundaries` at the same `crossingCost`. The intermediary file's `[[warps]]` section is a *semantic* distinction at the authoring layer, but on canonical output the CLI emits warps into the same `[[region.boundary]]` list as adjacency edges; the engine treats them identically. Rationale: keep this initiative CLI-only; the cost-distinction refactor (separate adjacency vs. warp edge types in the engine, differential cost in `neighbors()`) is filed as a follow-up. **Follow-up filed:** see Out of Scope.
8. **Warp endpoints addressed by grid `{ region, row, col }`; no per-warp cost.** Endpoints are hex positions (not world IDs); both endpoints are specified as `{ region = "X", row = N, col = M }` where row is the 0-indexed grid row (comments and blank lines do *not* count) and col is the 0-indexed cell index within that row (leftmost cell = 0, regardless of row indent). The CLI validates that the addressed cell exists and belongs to the named region; it does *not* enforce the "edge hex" convention (your call: GM responsibility). No per-warp cost field — crossing cost is asset-computed at traversal time, not inherent to the warp.
9. **Top-level `spatial` group with `build` subcommand.** Command shape: `gm-toolkit spatial build [--campaign <id>] [--replace]`. Rationale: leaves room for plausible future subcommands (`spatial validate`, `spatial show`, `spatial diff`) without re-rooting; "build" reads truer than "add" for an intermediary-to-canonical operation.
10. **Intermediary files live at `<campaign>/spatial/source/` by convention.** GM creates `<campaign>/spatial/source/layout.txt` and `<campaign>/spatial/source/data.toml`; CLI reads them from this fixed location and emits canonical outputs to `<campaign>/spatial/regions.toml` + `<campaign>/spatial/worlds.toml`. No `--layout` / `--data` path flags. Rationale: the CLI surface stays minimal; source files live with the campaign they describe and travel with it. The GM owns their authoring artifacts; the CLI just knows where to look.
11. **Overwrite policy follows `campaign add-rules`.** If `regions.toml` or `worlds.toml` already exist in `<campaign>/spatial/`, the CLI errors with a clear message and a hint to pass `--replace`. With `--replace`, both canonical files are overwritten without prompt. Rationale: consistency with the existing seed-the-campaign idiom; no novel destructive defaults.
12. **Coordinate mapping: odd-r offset → axial.** Internal-only; the GM never sees axial. The CLI translates each `(row, col)` cell index to `HexCoord{Q, R}` via the standard odd-r formula: `Q = col - (row - (row & 1)) / 2; R = row`. This matches the visual offset in the example layout (odd rows indented right) and produces axial coords compatible with the existing `region_map.go` adjacency math.
13. **Validation: re-run `LoadRegionMap` invariants as self-check; attribute errors to source files.** After emitting `regions.toml`/`worlds.toml`, the CLI re-parses them via `LoadRegionMap` to confirm the output is internally consistent. Any failure is reported with reference to the *source* file location (layout line/cell or data.toml entry), not the generated canonical file — the GM is editing the source. Rationale: the GM never debugs generated TOML; errors must point back to what they authored.

## Open Questions

None. Discovery resolved all branching decisions; the questions Plan addresses are implementation-shaped (package layout under `cmd/gm-toolkit/spatial/`, parser strategy, error-message phrasing, test fixture shape) rather than design-shaped.

## Per-Area Design Details

### ASCII Layout Syntax

The layout file is a grid of single-character cells separated by single spaces. Odd rows (counting from zero) are indented one leading space; this renders the pointy-top hex stagger so the GM authoring the map sees the actual hex topology, not a square grid.

**Cell alphabet:**

| Glyph         | Meaning                                                     |
|---------------|-------------------------------------------------------------|
| `.`           | Empty hex (no region, no world).                            |
| `A`-`Z`       | Region membership. Hex belongs to the named region.         |
| `1`-`9`, `a`-`z` | World marker. Hex contains the world identified by this glyph in the TOML data file. The hex's region is determined per the [World Placement](#world-placement) rules. |

**Mechanics:**

- Cells are separated by single spaces. Trailing whitespace on a row is tolerated by the parser.
- Lines beginning with `#` (optionally with leading whitespace) are comments. Useful for a sector name header, region legend, or annotations.
- No required preamble. The file is the grid plus optional comments.
- Empty lines are tolerated and ignored (they do not create blank rows in the grid).

**Example** (4 regions, 4 worlds):

```
# Sector: Helion's Reach
# Regions: A=Coreward Reach, B=Midline Sector, C=Outer Veil, D=Nova Cascade
. . . . . . . . . . . . . . . . . . .
 . A A A . . . . . . . . D 1 D . . .
. . A 2 A B B . . . . . . D D D . . .
 . . A A B B B . . . . . . D D D . . 
. . . . B 3 B . . . . . . . . . . . .
 . . . . . B B . . . . . . . . . . .
. . . . . . . . C C C . . . . . . . .
 . . . . . . C C 4 . C . . . . . . .
. . . . . . . C C . C C . . . . . . .
```

**Capacity:** 26 regions, 35 worlds per map. Beyond either limit the map should be split (or the alphabet revisited; deferred until a real map hits the cap).

**Companion `data.toml` for the example above:**

```toml
# Region metadata. Every region letter that appears in the layout
# needs an entry; the CLI errors on missing ones.
[regions.A]
name = "Coreward Reach"

[regions.B]
name = "Midline Sector"

[regions.C]
name = "Outer Veil"

[regions.D]
name = "Nova Cascade"

# Worlds. `at` is the marker glyph (integer for 1-9, string for a-z).
# `region` is omitted when inference is unambiguous.
[[worlds]]
id = "helion"
name = "Helion"
at = 1
tech_level = 8
population = 1200000

[[worlds]]
id = "anvil"
name = "Anvil"
at = 2
tech_level = 6
population = 80000

[[worlds]]
id = "bastion"
name = "Bastion"
at = 3
tech_level = 5
population = 200000

[[worlds]]
id = "veilport"
name = "Veilport"
at = 4
tech_level = 7
population = 950000

# Warps. Endpoints are { region, row, col } grid coords.
# Convention (not enforced): each endpoint is an edge hex of its region.
[[warps]]
from = { region = "A", row = 1, col = 1 }
to   = { region = "D", row = 1, col = 13 }
```

### Coordinate Mapping

Internal-only — the user never sees axial coordinates. The CLI converts each `(row, col)` cell index to `spatial.HexCoord{Q, R}` using the **odd-r offset** formula:

```go
q := col - (row - (row & 1)) / 2
r := row
```

Where `row` and `col` are the 0-indexed cell positions (per the [Warps](#warps) addressing rule). The odd-r convention matches the visual offset in the example layout (odd rows indented one space to the right).

The same formula is used everywhere the CLI needs to translate a layout position to an axial coord — region hex emission, world placement, warp endpoint validation, and adjacency neighbor scanning.

**Neighbor scan in offset terms.** Because the CLI works in offset space until the final emission, it uses the offset-coord neighbor table directly rather than translating each step through axial:

- **Even rows** (col, row): W (col-1, row), E (col+1, row), NW (col-1, row-1), NE (col, row-1), SW (col-1, row+1), SE (col, row+1).
- **Odd rows** (col, row): W (col-1, row), E (col+1, row), NW (col, row-1), NE (col+1, row-1), SW (col, row+1), SE (col+1, row+1).

Out-of-bounds neighbors are simply absent (edge cells have fewer than 6 neighbors).

### World Placement

The link between a layout marker and its TOML entry is the **marker glyph itself**. The world's TOML entry includes `at = "<glyph>"` (or `at = <digit>` for digits 1-9; strings for letters `a`-`z`). The CLI matches by glyph; each glyph appears at most once in the layout (a marker is unique per map).

**Region assignment.** The CLI walks the 6 hex neighbors of the marker cell and collects the set of distinct **region letters** (uppercase `A`-`Z`) among non-empty neighbors. World markers (digits / lowercase) and empty cells contribute nothing to the inference — they are transparent. So two adjacent world markers don't cycle through each other for inference; each is settled independently against its region-letter neighbors:

- **Exactly one region letter** → auto-assign that region. No `region` field in TOML needed.
- **Multiple region letters** → CLI errors: `world '2' at (row=4, col=5) has multiple region candidates [A, B]; add region = "X" to its TOML entry`. The world's TOML entry must include `region = "X"`; the override is honored.
- **Zero non-empty neighbors** (floating in empty space) → same error path as multi-candidate; the world's TOML entry must include `region = "X"`.

**Hex-membership emission.** Once the world's region is known, the CLI emits the marker cell's axial coords into that region's `hexes` list in `regions.toml`, AND emits the world to `worlds.toml` with that region and those coords. This satisfies `LoadRegionMap`'s invariant that a world's hex is in its region's hex set.

**Example TOML data file entry:**

```toml
[[worlds]]
id = "helion"
name = "Helion"
at = 1
tech_level = 8
population = 1200000
# region omitted -- inferred from layout neighbors

[[worlds]]
id = "frontier"
name = "Frontier"
at = "f"
tech_level = 4
population = 50000
region = "A"   # required: marker 'f' sits on the A/B border
```

### Adjacency and Boundary Derivation

After region membership is settled (including world-marker cells assigned to their inferred or explicit region), the CLI scans every region-member hex's 6 neighbors. For any neighbor that belongs to a *different* region, the pair is a cross-region adjacency.

**Canonicalization.** Each cross-region pair is emitted **once**, on the alphabetically-first region's side. So for an A-hex adjacent to a B-hex, the `[[region.boundary]]` entry lives under region A (`to_region = "B"`). This produces deterministic, diff-friendly output and avoids the double-listing the engine's `neighbors()` would also tolerate but doesn't need.

**Cost semantics today.** The current engine prices every `Region.Boundary` traversal at the parametric `crossingCost`. With dense adjacency emission, traversing a shared edge costs the same as a warp. That mismatch is acknowledged and **deferred** — see [Out of Scope](#out-of-scope) for the follow-up that introduces a per-edge-type cost distinction in the engine. For this initiative, the CLI emits both adjacency-derived edges and warp-declared edges into the same `[[region.boundary]]` list; the engine sees a unified set.

**Validation hook.** The CLI's output passes `LoadRegionMap`'s existing invariants (each boundary's `From` is in the source region's hex set; `ToRegion` exists; `To` is in target region's hex set). The deterministic alphabetical-side rule guarantees these by construction.

### Warps

A warp is a non-adjacent connection between two hexes in different regions (or, in principle, the same region; not disallowed). Authored in the TOML data file's `[[warps]]` array.

**Endpoint addressing.** Each endpoint is `{ region = "X", row = N, col = M }`:

- `region` — the region letter the cell belongs to. The CLI verifies the addressed cell is actually a member of this region; mismatch is an error with a clear message naming the actual region.
- `row` — 0-indexed grid row. **Comments and blank lines do not count.** The first non-comment line of the grid is row 0.
- `col` — 0-indexed cell index within the row. The leftmost cell is col 0, regardless of whether the row is indented. Odd rows' indent is rendering metadata, not a cell.

The GM counts rows and cells directly off the ASCII layout. No axial coordinates ever appear in user-facing files.

**Example:**

```toml
[[warps]]
from = { region = "A", row = 1, col = 6 }
to   = { region = "D", row = 5, col = 4 }

[[warps]]
from = { region = "B", row = 3, col = 8 }
to   = { region = "C", row = 7, col = 2 }
```

**No per-warp cost field.** Crossing cost is the caller's responsibility (assets compute their own cost at traversal time, passed into `Distance`/`Path` as the `crossingCost` parameter). The warp entry is pure topology.

**Convention (not enforced).** Warps are expected to connect edge-hexes of one region to edge-hexes of another, but the CLI does not validate this — it trusts the GM's authorship.

**Canonical emission.** Each warp is emitted as a `[[region.boundary]]` entry on the alphabetically-first region's side. This lumps warps into the same canonical list as adjacency-derived boundaries (per the deferred cost-distinction decision); the engine sees a unified set today.

### CLI Surface

```
gm-toolkit spatial build [--campaign <id>] [--replace]
```

A new top-level `spatial` cobra group registered in `cmd/gm-toolkit/root.go` alongside `campaign` and `faction`. The group's package lives at `cmd/gm-toolkit/spatial/` mirroring `cmd/gm-toolkit/campaign/`. The `build` subcommand is the only one in this initiative; `spatial validate` / `spatial show` are plausible follow-ups but not in scope.

**Inputs** — fixed locations under the resolved campaign:

- `<campaign>/spatial/source/layout.txt` — the ASCII layout.
- `<campaign>/spatial/source/data.toml` — the data file (region names, world metadata, warps).

Missing either file is an error (`spatial build: missing source file <path>; create <campaign>/spatial/source/layout.txt and data.toml`).

**Outputs** — written to `<campaign>/spatial/`:

- `regions.toml`
- `worlds.toml`

**Flags:**

- `--campaign <id>` — target campaign. Defaults to active campaign per `campaigns.ResolveActive`. Same semantics as `campaign add-rules`.
- `--replace` — overwrite existing `regions.toml` / `worlds.toml`. Without it, CLI errors if either output file exists, with a hint to pass `--replace`.

**Stdout on success** — a one-shot summary:

```
Built spatial map for campaign <id>:
  4 regions (A, B, C, D)
  12 worlds
  28 adjacency boundaries
  3 warps
Wrote <campaign>/spatial/regions.toml
Wrote <campaign>/spatial/worlds.toml
```

**Exit codes:** 0 on success, non-zero on any validation or I/O failure. Errors print to stderr.

### Validation

The CLI validates in two passes:

**Pass 1: Source-file validation (parse + semantic).** Before emitting any output:

- **Layout parse** — every non-comment line is a valid grid row (consistent cell count per row given the row's parity / indent; only valid glyphs); region letters reachable; world markers unique across the map.
- **Data file parse** — TOML syntactic validity; every region letter that appears in the layout has a `[regions.X]` entry (or default name); every world marker in the layout has a `[[worlds]]` entry; every world `region` override (when present) matches the layout neighbors' candidates or explicitly overrides ambiguity; every `[[warps]]` endpoint references a cell that exists and belongs to its declared region.

**Pass 2: Canonical-output validation.** After emitting `regions.toml` + `worlds.toml`, the CLI re-loads them via `spatial.LoadRegionMap` to confirm the output satisfies the engine's invariants (boundary `From` in source region's hex set; `ToRegion` exists; `To` in target region's hex set; worlds within their region). This catches CLI bugs (incorrect coord translation, missed boundary emission) before the GM sees a corrupt canonical file.

**Error attribution.** All errors reference the *source* file the GM is editing:

- Layout errors: `layout.txt:<line>:<col>: <message>` (e.g., `layout.txt:5:13: world marker '3' appears twice; first at line 3, col 8`).
- Data file errors: `data.toml:[[worlds]] id="helion": <message>` (e.g., `data.toml:[[worlds]] id="frontier": references marker 'q' but no such marker in layout.txt`).
- Pass-2 failures: surfaced as `internal: canonical output failed self-check (<message>); please file a bug` — these indicate a CLI defect, not GM error.

The CLI never reports an error citing the generated `regions.toml` or `worlds.toml` directly in the user-facing message.

## User-Facing Impact

After this initiative ships, the GM's spatial-map authoring loop is:

1. **Scaffold a campaign** (existing `campaign create`). The campaign root has an empty `spatial/` directory ready for content.
2. **Draft the source files** in `<campaign>/spatial/source/`:
   - `layout.txt` — the ASCII grid (regions as `A`-`Z`, worlds as `1`-`9` then `a`-`z`, `.` for empty hexes, odd rows indented one space).
   - `data.toml` — region names, world metadata, warps.
3. **Run `gm-toolkit spatial build`** (with `--campaign <id>` if not the active one). On success, `regions.toml` and `worlds.toml` are written to `<campaign>/spatial/`.
4. **Iterate.** Edit the source files in any text editor; re-run `spatial build --replace` to regenerate.

What the GM *no longer* does: write axial hex coords by hand, enumerate cross-region boundaries manually, keep `regions.toml` and `worlds.toml` consistent by eye. What the GM still does: hand-edit the source files (no in-CLI editor), reason about which regions touch which, and place worlds intentionally.

Concrete unlock: tui-manage (F-005.2) can be authored against a real (non-toy) spatial map for testing and demos. The full TUI authoring loop now has the spatial layer it needs.

## Out of Scope

- **Reverse conversion (canonical -> intermediary).** No round-trip. Canonical files are machine-output; the GM edits the intermediary pair and rebuilds.
- **In-CLI editing.** No interactive map editor. The intermediary files are edited in any text editor.
- **Maps larger than 26 regions or 35 worlds.** If a real campaign hits the cap, revisit the alphabet (or split into multiple maps); deferred until that materializes.
- **Terrain / hex-type metadata.** Hexes are either empty or a region member; no per-hex type (e.g., gas giant, asteroid field). If the engine grows hex-type semantics, the intermediary schema is extended at that point.
- **Multi-map authoring per campaign.** One spatial map per campaign for now, matching the current `LoadRegionMap` contract.
- **Faction-on-world placement.** The CLI authors the spatial topology; what factions hold which worlds is a state-layer concern handled elsewhere.
- **Adjacency vs. warp cost distinction in the engine.** Filed as a follow-up bugfix/refactor (to be added to `planned-work.md` Backlog when this initiative is promoted): split `Region.Boundary` into adjacency-edges (cost 1) and warp-edges (cost `crossingCost`), update `neighbors()` accordingly. Today's CLI emits both into the unified `[[region.boundary]]` list; the follow-up adds the distinction and the CLI gets a follow-on update to emit a `[[warps]]` (or `kind`-tagged) canonical form.

## Reference Exemplars

- **`internal/spatial/region_map.go`** — the canonical schema this CLI emits. `tomlRegion` / `tomlBoundary` / `tomlWorld` define the output shape.
- **`internal/spatial/region_map.go` — `LoadRegionMap`** — the validation rules the CLI must satisfy (boundary endpoints inside region hexes, world hex inside its region, unknown-region rejection).
- **`internal/spatial/region_map.go` — `neighbors`** — clarifies how boundaries function at pathfinding time (they are alternative edges with `crossingCost`, not opaque borders).
- **`cmd/gm-toolkit/campaign/`** — existing subcommand-group pattern under cobra root; the spatial CLI plugs in alongside.
