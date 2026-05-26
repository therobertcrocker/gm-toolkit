# Spatial Map CLI — Implementation Plan

## Context / Goal

Implement the `gm-toolkit spatial build` CLI per the [Spatial Map CLI Discovery](../discovery/spatial-map-cli-discovery.md). The CLI consumes a GM-authored ASCII layout (`layout.txt`) and TOML data file (`data.toml`) under `<campaign>/spatial/source/` and emits the canonical `regions.toml` + `worlds.toml` that `internal/spatial.LoadRegionMap` already understands. One-way builder; no reverse conversion.

- Discovery doc: [`spatial-map-cli-discovery.md`](../discovery/spatial-map-cli-discovery.md)

## Decisions Ratified in Planning

1. **Builder logic lives in `internal/spatial/builder/`.** New sub-package alongside the engine. CLI in `cmd/gm-toolkit/spatial/` stays thin (cobra wiring only) and calls into `builder.Build(...)`. Rationale: mirrors the `cmd/gm-toolkit/campaign/` ↔ `internal/campaigns` pattern; builder doesn't pollute the engine package; testable in isolation. Builder declares its own canonical TOML structs (lock-step with engine; no shared schema package today — YAGNI).
2. **Fused validation + derivation, fail-fast errors.** A single `derive(layout, data)` pass walks regions/worlds/warps, building the model and returning the first source-level error encountered. No batched-error reporting. Rationale: simpler code, and the GM's edit/rerun loop is fast enough that one-error-at-a-time is fine. If multi-error reporting becomes valuable, revisit.
3. **Sentinels-plus-helpers error shape.** Two exported sentinels (`ErrMissingSource`, `ErrOutputExists`) — the only categories the CLI branches on. All source-attributed errors go through internal helpers `layoutErrorf(line, col, format, args)` and `dataErrorf(context, format, args)` that produce the canonical prefix (`layout.txt:<line>:<col>:` / `data.toml:<context>:`). Self-check failures are wrapped inline with the `internal: ... please file a bug` framing. Tests use `errors.Is` for sentinel categories and `strings.Contains` for attribution. Matches the `internal/campaigns` pattern.
4. **Hybrid test coverage.** One happy-path golden fixture (Helion's Reach from Discovery) plus 4–5 representative e2e error fixtures (layout-syntax, missing-region-entry, ambiguous-world, unknown-marker, warp-wrong-region). Inline-string unit tests cover parseLayout edge cases and derive region-inference detail. Self-check round-trip test calls `Build` then `spatial.LoadRegionMap` and asserts no error. `-update` flag regenerates goldens. Rationale: golden fixtures are visible and editable; inline strings stay tight for cases where the input is trivially small.
5. **Two commits.** Commit 1 lands the entire `internal/spatial/builder` package with all tests (the substantive work, driven through `builder.Build` directly). Commit 2 wires the cobra `spatial build` subcommand. Each commit is one execution session. Rationale: keeps the heavy work in a single substantive commit (per memory's no-padded-splits preference); CLI wiring is thin enough to deserve its own session for review clarity but not big enough to split further.

---

## Shared Context

### Package Layout

```
cmd/gm-toolkit/spatial/
  spatial.go     -- cobra group
  build.go       -- `spatial build` subcommand

internal/spatial/builder/
  builder.go     -- public Build entrypoint and option struct
  layout.go      -- ASCII layout parser
  data.go        -- data.toml parser
  derive.go      -- region/world/boundary derivation, adjacency scan
  emit.go        -- canonical regions.toml / worlds.toml writer
  errors.go      -- sentinels and source-attribution helpers
  testdata/      -- golden fixtures (see Test Strategy)
```

Flat package — no further sub-packaging within `builder/`. Files are organized by pipeline stage for readability, not by package boundary.

### Builder Pipeline

`Build` is the single public entrypoint. Internal stages run in fixed order; the first error short-circuits.

```go
// builder.go

package builder

type Options struct {
    SourceDir string // <campaign>/spatial/source/
    DestDir   string // <campaign>/spatial/
    Replace   bool
}

type Summary struct {
    Regions        []string // sorted region IDs ("A", "B", ...)
    WorldCount     int
    AdjacencyCount int
    WarpCount      int
}

func Build(opts Options) (Summary, error)
```

Internal flow:

```go
func Build(opts Options) (Summary, error) {
    if err := checkOutputs(opts); err != nil { return Summary{}, err }   // --replace gate
    layout, err := parseLayout(opts.SourceDir); if err != nil { return Summary{}, err }
    data,   err := parseData(opts.SourceDir);   if err != nil { return Summary{}, err }
    derived, err := derive(layout, data);       if err != nil { return Summary{}, err }
    if err := emit(opts.DestDir, derived);  err != nil { return Summary{}, err }
    if err := selfCheck(opts.DestDir);       err != nil { return Summary{}, err }
    return summarize(derived, len(data.Warps)), nil
}
```

#### Stage 1 — `parseLayout` (layout.go)

```go
type glyph rune // '.', 'A'-'Z', '1'-'9', 'a'-'z'

func (g glyph) isRegion() bool { return g >= 'A' && g <= 'Z' }
func (g glyph) isWorld() bool  { return (g >= '1' && g <= '9') || (g >= 'a' && g <= 'z') }
func (g glyph) isEmpty() bool  { return g == '.' }

type cell struct {
    Glyph    glyph
    Row      int // 0-indexed grid row (excludes comments/blank lines)
    Col      int // 0-indexed cell index within row (indent of odd rows is rendering, not a cell)
    FileLine int // 1-indexed line in layout.txt, for error attribution
    FileCol  int // 1-indexed column in layout.txt, for error attribution
}

type layoutFile struct {
    cells         [][]cell        // cells[row][col]
    regionLetters map[glyph]bool  // distinct region letters used in the layout
    markers       map[glyph]cell  // world glyph → cell (uniqueness enforced at parse time)
}

func parseLayout(srcDir string) (*layoutFile, error)
```

Mechanics:
- Read `<srcDir>/layout.txt`. Missing file → wrap `ErrMissingSource` (see Error Machinery).
- Walk lines; skip comments (`^\s*#`) and blank lines. Each surviving line becomes a grid row; each cell stores its `FileLine`/`FileCol` so post-parse code can report errors at file coordinates.
- Within each grid row, strip the single leading-space indent if the row is odd (row index 1, 3, 5, ...). Trailing whitespace tolerated. Cells are single non-space chars separated by single spaces; any deviation (multi-char cell, missing separator, invalid glyph) is an error attributed to `layout.txt:<fileLine>:<fileCol>`.
- Track world markers in `markers`; second occurrence of any glyph is a duplicate-marker error pointing at both source positions (file line + file col).

#### Stage 2 — `parseData` (data.go)

```go
type dataFile struct {
    Regions map[string]regionEntry `toml:"regions"` // key = "A", "B", ...
    Worlds  []worldEntry           `toml:"worlds"`
    Warps   []warpEntry            `toml:"warps"`
}

type regionEntry struct {
    Name string `toml:"name"`
}

type worldEntry struct {
    ID         string  `toml:"id"`
    Name       string  `toml:"name"`
    At         atGlyph `toml:"at"`
    TechLevel  int     `toml:"tech_level"`
    Population int     `toml:"population"`
    Region     string  `toml:"region"` // optional override
}

// atGlyph decodes TOML int (1-9) or string ("a"-"z") into a glyph rune.
type atGlyph struct{ Glyph glyph }

func (a *atGlyph) UnmarshalTOML(v interface{}) error // type-switch int64 / string; reject otherwise

type warpEntry struct {
    From hexAddr `toml:"from"`
    To   hexAddr `toml:"to"`
}

type hexAddr struct {
    Region string `toml:"region"`
    Row    int    `toml:"row"`
    Col    int    `toml:"col"`
}

func parseData(srcDir string) (*dataFile, error)
```

Mechanics:
- Read `<srcDir>/data.toml` via `BurntSushi/toml.DecodeFile`. Missing file → wrap `ErrMissingSource`. TOML syntax errors are wrapped with the data-file prefix via `dataErrorf("", ...)`.
- No semantic checks here — those happen in `derive`. parseData is parse-only.

#### Stage 3 — `derive` (derive.go)

Produces the in-memory model passed to emit.

```go
type derivedMap struct {
    Regions        []derivedRegion // sorted by ID
    Worlds         []derivedWorld  // sorted by ID
    AdjacencyCount int             // count of boundaries from adjacency (not warps), set by derive() between steps 5 and 6
}

type derivedRegion struct {
    ID         string
    Name       string
    Hexes      []spatial.HexCoord // sorted (Q, R)
    Boundaries []derivedBoundary  // sorted by (ToRegion, From.Q, From.R, To.Q, To.R)
}

type derivedBoundary struct {
    From     spatial.HexCoord
    ToRegion string
    To       spatial.HexCoord
}

type derivedWorld struct {
    ID         string
    Name       string
    TechLevel  int
    Population int
    Region     string
    Coord      spatial.HexCoord
}

func derive(layout *layoutFile, data *dataFile) (*derivedMap, error)
```

Derivation order (fail-fast, first error returned):

1. **Cross-check regions.** Every `glyph` in `layout.regionLetters` must have a `data.Regions[string(glyph)]` entry. Missing → error attributed to `data.toml`. Extra entries in `data.Regions` not referenced by the layout → error (defensive — likely a typo).
2. **Cross-check worlds.** Bijection between `layout.markers` keys and `data.Worlds[*].At.Glyph`. Each side surfaces a different error message:
   - Marker in layout, no matching `[[worlds]]` entry → error at the layout cell.
   - `[[worlds]]` entry with glyph not in layout → error at the data entry.
3. **Build region hex sets.** For each region-letter cell, translate `(row, col) → HexCoord` using the odd-r formula (Discovery's Coordinate Mapping section) and add to that region's hex set.
4. **Place worlds.** For each world marker:
   1. Walk the 6 offset-coord neighbors; collect distinct region letters from non-empty neighbors that are themselves region-letter cells (world markers and empty cells are transparent).
   2. Resolve region per Discovery Decision 5 (auto-infer / explicit override / error with candidate list).
   3. Add the world cell to the resolved region's hex set; add a `derivedWorld` with the world's TOML metadata.
5. **Derive adjacency boundaries.** Iterate every region-member hex (including world cells). For each of its 6 offset neighbors that belongs to a *different* region, emit one `derivedBoundary` on the alphabetically-first region's side. Deduplicate by `(fromRegion, From, toRegion, To)` so each shared edge appears exactly once.
6. **Process warps.** For each `warpEntry`: validate that each endpoint's `(region, row, col)` refers to an actual cell belonging to that region (look up the cell in `layout.cells[row][col]`, confirm membership in the resolved region's hex set). Translate to axial. Emit as a `derivedBoundary` on the alphabetically-first region's side. Warps and adjacency edges go into the same boundary list (Discovery Decision 7).
7. **Sort.** Regions by ID; each region's Hexes by `(Q, R)`; each region's Boundaries by `(ToRegion, From.Q, From.R, To.Q, To.R)`; Worlds by ID. Deterministic output for diff-friendliness.

#### Stage 4 — `emit` (emit.go)

```go
type emitRegion struct {
    ID         string         `toml:"id"`
    Name       string         `toml:"name"`
    Hexes      [][2]int       `toml:"hexes"`
    Boundaries []emitBoundary `toml:"boundary"`
}

type emitBoundary struct {
    FromQ    int    `toml:"from_q"`
    FromR    int    `toml:"from_r"`
    ToRegion string `toml:"to_region"`
    ToQ      int    `toml:"to_q"`
    ToR      int    `toml:"to_r"`
}

type emitWorld struct {
    ID         string `toml:"id"`
    Name       string `toml:"name"`
    TechLevel  int    `toml:"tech_level"`
    Population int    `toml:"population"`
    Region     string `toml:"region"`
    HexQ       int    `toml:"hex_q"`
    HexR       int    `toml:"hex_r"`
}

func emit(dstDir string, derived *derivedMap) error
```

Mechanics:
- Convert each `derivedRegion` → `emitRegion`; encode via `BurntSushi/toml` into `<dstDir>/regions.toml`.
- Convert each `derivedWorld` → `emitWorld`; encode into `<dstDir>/worlds.toml`.
- Field tags mirror `internal/spatial.tomlRegion` / `tomlBoundary` / `tomlWorld` exactly (lock-step coupling per Decision 1).
- Write order: `regions.toml` first, then `worlds.toml`. If the second write fails after the first succeeds, the canonical state is partially written; the next run requires `--replace`. Acceptable for v1.

#### Stage 5 — `selfCheck` (builder.go)

```go
func selfCheck(dstDir string) error {
    if _, err := spatial.LoadRegionMap(dstDir); err != nil {
        return fmt.Errorf("internal: canonical output failed self-check (%w); please file a bug", err)
    }
    return nil
}
```

Per Discovery Decision 13, self-check failures are attributed to the CLI, not the GM.

### Error Machinery

```go
// errors.go

package builder

import "errors"

var (
    ErrMissingSource = errors.New("missing source file")
    ErrOutputExists  = errors.New("canonical output exists")
)
```

These are the only two sentinels — the CLI branches on them to print hints. All other errors are unstructured detail strings carrying the source-attribution prefix.

**Attribution helpers** (unexported):

```go
// errors.go (continued)

func layoutErrorf(line, col int, format string, args ...any) error {
    return fmt.Errorf("layout.txt:%d:%d: "+format, append([]any{line, col}, args...)...)
}

func dataErrorf(context, format string, args ...any) error {
    return fmt.Errorf("data.toml:%s: "+format, append([]any{context}, args...)...)
}
```

`context` for `dataErrorf` is the human-readable selector for the offending entry, formed at the call site:
- For a world entry: `[[worlds]] id="helion"`
- For a region entry: `[regions.A]`
- For a warp entry: `[[warps]] #3` (1-indexed array position; warps have no id)
- For top-level errors (e.g. malformed toml): the context is the empty string and the prefix collapses to `data.toml:`

**Sentinel usage at call sites:**

```go
// missing layout.txt or data.toml
return fmt.Errorf("%w: %s", ErrMissingSource, path)

// regions.toml/worlds.toml already exists and Replace == false
return fmt.Errorf("%w: %s", ErrOutputExists, path)
```

**Self-check wrapping:**

```go
return fmt.Errorf("internal: canonical output failed self-check (%w); please file a bug", err)
```

No sentinel; the CLI prints as-is. The wrapped engine error is preserved (via `%w`) so detail isn't lost.

**CLI consumption shape** (mirrors `add_rules.go`):

```go
// cmd/gm-toolkit/spatial/build.go (excerpt)
summary, err := builder.Build(opts)
if err != nil {
    switch {
    case errors.Is(err, builder.ErrMissingSource):
        return fmt.Errorf("spatial build: %w\n  hint: create %s/spatial/source/layout.txt and data.toml", err, campRoot)
    case errors.Is(err, builder.ErrOutputExists):
        return fmt.Errorf("spatial build: %w\n  hint: pass --replace to overwrite", err)
    default:
        return fmt.Errorf("spatial build: %w", err)
    }
}
```

### Test Strategy

Fixture layout under `internal/spatial/builder/testdata/`:

```
testdata/
  valid/
    helions_reach/
      source/
        layout.txt
        data.toml
      golden/
        regions.toml
        worlds.toml
  errors/
    layout_syntax/         # malformed cell (multi-char or invalid glyph)
      source/{layout.txt,data.toml}
      err.txt              # expected substring of err.Error()
    missing_region_entry/  # layout uses region letter with no [regions.X] entry
      ...
    ambiguous_world/       # marker with multiple region candidates, no override
      ...
    unknown_marker/        # [[worlds]] entry whose `at` glyph isn't in layout
      ...
    warp_wrong_region/     # warp endpoint cell belongs to a different region
      ...
```

**Test files:**

- `builder_test.go`
  - `TestBuild_Valid` — for each `testdata/valid/<case>`, run `Build`, diff emitted output against `golden/`. `-update` flag regenerates goldens.
  - `TestBuild_SelfCheckRoundTrip` — run `Build` on `valid/helions_reach`, then `spatial.LoadRegionMap(dstDir)`; assert no error and a sanity-check (e.g. world count matches).
  - `TestBuild_MissingSource` — `Build` against an empty source dir; assert `errors.Is(err, ErrMissingSource)`.
  - `TestBuild_OutputExists` — pre-create `regions.toml` in dest; assert `errors.Is(err, ErrOutputExists)` without `--replace` and success with.
  - `TestBuild_Errors` — table-driven over `testdata/errors/<case>`; each row runs `Build`, asserts `strings.Contains(err.Error(), readErrFile(case))`.
- `layout_test.go` — table-driven unit tests with inline string literals: comment handling, blank-line tolerance, odd-row indent stripping, trailing-whitespace tolerance, duplicate-marker detection, invalid-glyph detection.
- `derive_test.go` — table-driven unit tests for region inference: auto-infer single-candidate, multi-candidate error, zero-candidate error, explicit override honored, override conflict (override names a region not adjacent).
- `emit_test.go` — small derivedMap; encode; re-decode; verify field round-trip. Confirms field tags match engine's expectations.

**`-update` flag:**

```go
var update = flag.Bool("update", false, "regenerate testdata/valid/*/golden files")
```

When `-update` is set, `TestBuild_Valid` writes the emitted output over the golden files instead of comparing. Standard Go golden-file pattern; invoked as `go test ./internal/spatial/builder/ -update`.

**Temp dest dirs.** Tests use `t.TempDir()` for destination paths so disk state is per-test and auto-cleaned.

**Helion's Reach fixture.** The layout from Discovery (`# Sector: Helion's Reach`) becomes the canonical happy-path fixture. Copied verbatim into `testdata/valid/helions_reach/source/`; goldens generated once via `-update` and reviewed before commit.

## Out of Scope

Mirrors Discovery's Out of Scope; called out here for plan-time clarity:

- Reverse conversion (canonical → intermediary).
- In-CLI editing.
- Maps beyond 26 regions / 35 worlds.
- Terrain / hex-type metadata.
- Multi-map per campaign.
- Faction-on-world placement.
- Adjacency-vs-warp cost distinction (filed as engine follow-up).

---

## Work Breakdown

### Commit 1 — `feat(spatial): builder package — parse, derive, emit`

Land the entire `internal/spatial/builder` package with unit tests, the happy-path golden fixture (Helion's Reach), and the e2e error fixtures. No CLI surface yet — the package is exercised end-to-end through `builder.Build(...)` in tests.

##### Task 1.1 — `internal/spatial/builder/errors.go`

Declare the two exported sentinels and the two unexported attribution helpers per the Error Machinery section.

```go
package builder

import (
    "errors"
    "fmt"
)

var (
    ErrMissingSource = errors.New("missing source file")
    ErrOutputExists  = errors.New("canonical output exists")
)

func layoutErrorf(line, col int, format string, args ...any) error {
    return fmt.Errorf("layout.txt:%d:%d: "+format, append([]any{line, col}, args...)...)
}

func dataErrorf(context, format string, args ...any) error {
    return fmt.Errorf("data.toml:%s: "+format, append([]any{context}, args...)...)
}
```

##### Task 1.2 — `internal/spatial/builder/layout.go`

Declare `glyph`/`cell`/`layoutFile` types per Shared Context, plus `parseLayout`. Implementation:

```go
package builder

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type glyph rune

func (g glyph) isRegion() bool { return g >= 'A' && g <= 'Z' }
func (g glyph) isWorld() bool  { return (g >= '1' && g <= '9') || (g >= 'a' && g <= 'z') }
func (g glyph) isEmpty() bool  { return g == '.' }

type cell struct {
    Glyph    glyph
    Row      int
    Col      int
    FileLine int
    FileCol  int
}

type layoutFile struct {
    cells         [][]cell
    regionLetters map[glyph]bool
    markers       map[glyph]cell
}

func parseLayout(srcDir string) (*layoutFile, error) {
    path := filepath.Join(srcDir, "layout.txt")
    contents, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("%w: %s", ErrMissingSource, path)
        }
        return nil, fmt.Errorf("read %s: %w", path, err)
    }

    layout := &layoutFile{
        regionLetters: make(map[glyph]bool),
        markers:       make(map[glyph]cell),
    }

    for fileLine0, rawLine := range strings.Split(string(contents), "\n") {
        fileLine := fileLine0 + 1
        line := strings.TrimRight(rawLine, " \t\r")
        if strings.TrimSpace(line) == "" {
            continue
        }
        if strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
            continue
        }

        row := len(layout.cells)
        cells, err := parseLayoutRow(line, row, fileLine, layout.markers)
        if err != nil {
            return nil, err
        }
        for _, c := range cells {
            if c.Glyph.isRegion() {
                layout.regionLetters[c.Glyph] = true
            }
        }
        layout.cells = append(layout.cells, cells)
    }

    if len(layout.cells) == 0 {
        return nil, fmt.Errorf("layout.txt: empty grid (no non-comment, non-blank rows)")
    }
    return layout, nil
}

// parseLayoutRow scans a single grid row. Odd rows have a one-space indent that
// is rendering, not a cell; it is stripped before scanning. Each emitted cell
// records its 1-indexed FileLine/FileCol in the original (unstripped) file line.
func parseLayoutRow(line string, row, fileLine int, markers map[glyph]cell) ([]cell, error) {
    indentOffset := 0
    if row&1 == 1 && strings.HasPrefix(line, " ") {
        line = line[1:]
        indentOffset = 1
    }

    var cells []cell
    cellCol := 0
    for i := 0; i < len(line); {
        ch := line[i]
        fileCol := i + indentOffset + 1
        g := glyph(ch)
        if !g.isRegion() && !g.isWorld() && !g.isEmpty() {
            return nil, layoutErrorf(fileLine, fileCol, "invalid glyph %q (expected '.', 'A'-'Z', '1'-'9', or 'a'-'z')", ch)
        }
        c := cell{Glyph: g, Row: row, Col: cellCol, FileLine: fileLine, FileCol: fileCol}
        cells = append(cells, c)

        if g.isWorld() {
            if prior, seen := markers[g]; seen {
                return nil, layoutErrorf(fileLine, fileCol, "world marker %q appears twice; first at line=%d col=%d", ch, prior.FileLine, prior.FileCol)
            }
            markers[g] = c
        }
        cellCol++
        i++
        if i >= len(line) {
            break
        }
        if line[i] != ' ' {
            return nil, layoutErrorf(fileLine, i+indentOffset+1, "expected single-space separator, got %q", line[i])
        }
        i++
    }
    if len(cells) == 0 {
        return nil, layoutErrorf(fileLine, 1, "empty row")
    }
    return cells, nil
}
```

##### Task 1.3 — `internal/spatial/builder/data.go`

Declare the data.toml struct types per Shared Context, plus `parseData` and `atGlyph.UnmarshalTOML`. Implementation:

```go
package builder

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

type dataFile struct {
    Regions map[string]regionEntry `toml:"regions"`
    Worlds  []worldEntry           `toml:"worlds"`
    Warps   []warpEntry            `toml:"warps"`
}

type regionEntry struct {
    Name string `toml:"name"`
}

type worldEntry struct {
    ID         string  `toml:"id"`
    Name       string  `toml:"name"`
    At         atGlyph `toml:"at"`
    TechLevel  int     `toml:"tech_level"`
    Population int     `toml:"population"`
    Region     string  `toml:"region"`
}

type atGlyph struct{ Glyph glyph }

func (a *atGlyph) UnmarshalTOML(v interface{}) error {
    switch x := v.(type) {
    case int64:
        if x < 1 || x > 9 {
            return fmt.Errorf("'at' integer must be in [1,9], got %d", x)
        }
        a.Glyph = glyph('0' + byte(x))
        return nil
    case string:
        if len(x) != 1 {
            return fmt.Errorf("'at' string must be a single character, got %q", x)
        }
        r := rune(x[0])
        if r < 'a' || r > 'z' {
            return fmt.Errorf("'at' string must be 'a'-'z', got %q", x)
        }
        a.Glyph = glyph(r)
        return nil
    default:
        return fmt.Errorf("'at' must be integer 1-9 or string 'a'-'z', got %T", v)
    }
}

type warpEntry struct {
    From hexAddr `toml:"from"`
    To   hexAddr `toml:"to"`
}

type hexAddr struct {
    Region string `toml:"region"`
    Row    int    `toml:"row"`
    Col    int    `toml:"col"`
}

func parseData(srcDir string) (*dataFile, error) {
    path := filepath.Join(srcDir, "data.toml")
    if _, err := os.Stat(path); err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("%w: %s", ErrMissingSource, path)
        }
        return nil, fmt.Errorf("stat %s: %w", path, err)
    }
    var data dataFile
    if _, err := toml.DecodeFile(path, &data); err != nil {
        return nil, dataErrorf("", "decode: %v", err)
    }
    return &data, nil
}
```

##### Task 1.4 — `internal/spatial/builder/derive.go`

Declare the derived-model types per Shared Context, plus the helpers and `derive`. Implementation:

```go
package builder

import (
    "fmt"
    "sort"

    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Offset-coord neighbor deltas, format {colDelta, rowDelta}. Pointy-top hexes, odd-r staggered.
var evenRowOffsets = [6][2]int{{-1, 0}, {1, 0}, {-1, -1}, {0, -1}, {-1, 1}, {0, 1}}
var oddRowOffsets = [6][2]int{{-1, 0}, {1, 0}, {0, -1}, {1, -1}, {0, 1}, {1, 1}}

func axialOf(row, col int) spatial.HexCoord {
    return spatial.HexCoord{Q: col - (row-(row&1))/2, R: row}
}

func offsetNeighbors(row, col int) [][2]int {
    deltas := evenRowOffsets
    if row&1 == 1 {
        deltas = oddRowOffsets
    }
    out := make([][2]int, 0, 6)
    for _, d := range deltas {
        out = append(out, [2]int{row + d[1], col + d[0]})
    }
    return out
}

// regionAcc is the in-flight region accumulator used during derivation.
type regionAcc struct {
    name  string
    hexes map[spatial.HexCoord]bool
}

func derive(layout *layoutFile, data *dataFile) (*derivedMap, error) {
    // 1. Cross-check regions: every layout letter has an entry; no extras; ids are 'A'-'Z'.
    for g := range layout.regionLetters {
        if _, ok := data.Regions[string(g)]; !ok {
            return nil, dataErrorf(fmt.Sprintf("[regions.%s]", string(g)), "missing entry; region %q is used in layout", string(g))
        }
    }
    for id := range data.Regions {
        if len(id) != 1 || !glyph(id[0]).isRegion() {
            return nil, dataErrorf(fmt.Sprintf("[regions.%s]", id), "region id must be a single uppercase letter 'A'-'Z'")
        }
        if !layout.regionLetters[glyph(id[0])] {
            return nil, dataErrorf(fmt.Sprintf("[regions.%s]", id), "region %q has no cells in layout", id)
        }
    }

    // 2. Cross-check worlds: bijection between layout markers and data.Worlds[*].At.Glyph.
    worldsByGlyph := make(map[glyph]*worldEntry, len(data.Worlds))
    for i := range data.Worlds {
        w := &data.Worlds[i]
        if _, dup := worldsByGlyph[w.At.Glyph]; dup {
            return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "duplicate `at` glyph %q", string(rune(w.At.Glyph)))
        }
        worldsByGlyph[w.At.Glyph] = w
        if _, ok := layout.markers[w.At.Glyph]; !ok {
            return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "references marker %q but no such marker in layout.txt", string(rune(w.At.Glyph)))
        }
    }
    for g, c := range layout.markers {
        if _, ok := worldsByGlyph[g]; !ok {
            return nil, layoutErrorf(c.FileLine, c.FileCol, "world marker %q has no [[worlds]] entry in data.toml", string(rune(g)))
        }
    }

    // 3. Build region hex sets from region-letter cells.
    regions := make(map[string]*regionAcc, len(data.Regions))
    for id, entry := range data.Regions {
        regions[id] = &regionAcc{name: entry.Name, hexes: make(map[spatial.HexCoord]bool)}
    }
    cellRegion := make(map[[2]int]string) // (row,col) -> region id
    for row, rowCells := range layout.cells {
        for col, c := range rowCells {
            if c.Glyph.isRegion() {
                regions[string(c.Glyph)].hexes[axialOf(row, col)] = true
                cellRegion[[2]int{row, col}] = string(c.Glyph)
            }
        }
    }

    // 4. Place worlds: infer region from non-empty region-letter neighbors, or honor explicit override.
    derivedWorlds := make([]derivedWorld, 0, len(data.Worlds))
    for _, w := range data.Worlds {
        c := layout.markers[w.At.Glyph]
        candidates := map[string]bool{}
        for _, n := range offsetNeighbors(c.Row, c.Col) {
            nrow, ncol := n[0], n[1]
            if nrow < 0 || nrow >= len(layout.cells) {
                continue
            }
            if ncol < 0 || ncol >= len(layout.cells[nrow]) {
                continue
            }
            nc := layout.cells[nrow][ncol]
            if nc.Glyph.isRegion() {
                candidates[string(nc.Glyph)] = true
            }
        }

        var regionID string
        switch {
        case w.Region != "":
            if _, ok := regions[w.Region]; !ok {
                return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID), "region override %q is not a declared region", w.Region)
            }
            regionID = w.Region
        case len(candidates) == 1:
            for id := range candidates {
                regionID = id
            }
        case len(candidates) == 0:
            return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID),
                "marker %q at (row=%d, col=%d) has no region-letter neighbors; add `region = \"X\"` to its entry",
                string(rune(w.At.Glyph)), c.Row, c.Col)
        default:
            return nil, dataErrorf(fmt.Sprintf("[[worlds]] id=%q", w.ID),
                "marker %q at (row=%d, col=%d) has multiple region candidates %v; add `region = \"X\"` to disambiguate",
                string(rune(w.At.Glyph)), c.Row, c.Col, sortedKeys(candidates))
        }

        coord := axialOf(c.Row, c.Col)
        regions[regionID].hexes[coord] = true
        cellRegion[[2]int{c.Row, c.Col}] = regionID
        derivedWorlds = append(derivedWorlds, derivedWorld{
            ID: w.ID, Name: w.Name, TechLevel: w.TechLevel, Population: w.Population,
            Region: regionID, Coord: coord,
        })
    }

    // 5. Derive adjacency boundaries, canonicalized to the alphabetically-first region's side.
    type edgeKey struct {
        fromRegion string
        from       spatial.HexCoord
        toRegion   string
        to         spatial.HexCoord
    }
    seen := map[edgeKey]bool{}
    boundariesByRegion := map[string][]derivedBoundary{}
    for ck, regionID := range cellRegion {
        row, col := ck[0], ck[1]
        for _, n := range offsetNeighbors(row, col) {
            nrow, ncol := n[0], n[1]
            if nrow < 0 || nrow >= len(layout.cells) {
                continue
            }
            if ncol < 0 || ncol >= len(layout.cells[nrow]) {
                continue
            }
            otherID, ok := cellRegion[[2]int{nrow, ncol}]
            if !ok || otherID == regionID {
                continue
            }
            ownerID, ownerCoord, neighborID, neighborCoord := regionID, axialOf(row, col), otherID, axialOf(nrow, ncol)
            if neighborID < ownerID {
                ownerID, neighborID = neighborID, ownerID
                ownerCoord, neighborCoord = neighborCoord, ownerCoord
            }
            key := edgeKey{fromRegion: ownerID, from: ownerCoord, toRegion: neighborID, to: neighborCoord}
            if seen[key] {
                continue
            }
            seen[key] = true
            boundariesByRegion[ownerID] = append(boundariesByRegion[ownerID], derivedBoundary{
                From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
            })
        }
    }
    adjacencyCount := 0
    for _, bs := range boundariesByRegion {
        adjacencyCount += len(bs)
    }

    // 6. Process warps. Validate endpoints; emit on alphabetically-first region's side.
    for i, w := range data.Warps {
        ctx := fmt.Sprintf("[[warps]] #%d", i+1)
        if err := validateWarpEndpoint(layout, regions, cellRegion, w.From, ctx, "from"); err != nil {
            return nil, err
        }
        if err := validateWarpEndpoint(layout, regions, cellRegion, w.To, ctx, "to"); err != nil {
            return nil, err
        }
        ownerID, ownerCoord := w.From.Region, axialOf(w.From.Row, w.From.Col)
        neighborID, neighborCoord := w.To.Region, axialOf(w.To.Row, w.To.Col)
        if neighborID < ownerID {
            ownerID, neighborID = neighborID, ownerID
            ownerCoord, neighborCoord = neighborCoord, ownerCoord
        }
        boundariesByRegion[ownerID] = append(boundariesByRegion[ownerID], derivedBoundary{
            From: ownerCoord, ToRegion: neighborID, To: neighborCoord,
        })
    }

    // 7. Assemble and sort.
    out := &derivedMap{AdjacencyCount: adjacencyCount}
    for _, id := range sortedKeys(mapKeysAsBoolSet(regions)) {
        acc := regions[id]
        hexes := make([]spatial.HexCoord, 0, len(acc.hexes))
        for h := range acc.hexes {
            hexes = append(hexes, h)
        }
        sort.Slice(hexes, func(i, j int) bool {
            if hexes[i].Q != hexes[j].Q {
                return hexes[i].Q < hexes[j].Q
            }
            return hexes[i].R < hexes[j].R
        })
        bs := boundariesByRegion[id]
        sort.Slice(bs, func(i, j int) bool {
            switch {
            case bs[i].ToRegion != bs[j].ToRegion:
                return bs[i].ToRegion < bs[j].ToRegion
            case bs[i].From.Q != bs[j].From.Q:
                return bs[i].From.Q < bs[j].From.Q
            case bs[i].From.R != bs[j].From.R:
                return bs[i].From.R < bs[j].From.R
            case bs[i].To.Q != bs[j].To.Q:
                return bs[i].To.Q < bs[j].To.Q
            default:
                return bs[i].To.R < bs[j].To.R
            }
        })
        out.Regions = append(out.Regions, derivedRegion{
            ID: id, Name: acc.name, Hexes: hexes, Boundaries: bs,
        })
    }
    sort.Slice(derivedWorlds, func(i, j int) bool { return derivedWorlds[i].ID < derivedWorlds[j].ID })
    out.Worlds = derivedWorlds
    return out, nil
}

func validateWarpEndpoint(layout *layoutFile, regions map[string]*regionAcc, cellRegion map[[2]int]string, addr hexAddr, ctx, side string) error {
    if addr.Row < 0 || addr.Row >= len(layout.cells) {
        return dataErrorf(ctx, "%s.row=%d is out of bounds (grid has %d rows)", side, addr.Row, len(layout.cells))
    }
    if addr.Col < 0 || addr.Col >= len(layout.cells[addr.Row]) {
        return dataErrorf(ctx, "%s.col=%d is out of bounds for row %d (has %d cells)", side, addr.Col, addr.Row, len(layout.cells[addr.Row]))
    }
    if _, ok := regions[addr.Region]; !ok {
        return dataErrorf(ctx, "%s.region=%q is not a declared region", side, addr.Region)
    }
    actual, ok := cellRegion[[2]int{addr.Row, addr.Col}]
    if !ok {
        return dataErrorf(ctx, "%s cell at (row=%d, col=%d) is empty, not a region member", side, addr.Row, addr.Col)
    }
    if actual != addr.Region {
        return dataErrorf(ctx, "%s cell at (row=%d, col=%d) belongs to region %q, not %q", side, addr.Row, addr.Col, actual, addr.Region)
    }
    return nil
}

func sortedKeys(m map[string]bool) []string {
    out := make([]string, 0, len(m))
    for k := range m {
        out = append(out, k)
    }
    sort.Strings(out)
    return out
}

// mapKeysAsBoolSet converts map[string]*regionAcc to map[string]bool for sortedKeys reuse.
func mapKeysAsBoolSet(m map[string]*regionAcc) map[string]bool {
    out := make(map[string]bool, len(m))
    for k := range m {
        out[k] = true
    }
    return out
}
```

##### Task 1.5 — `internal/spatial/builder/emit.go`

Declare the canonical emit types (parallel to engine's unexported `tomlRegion`/`tomlBoundary`/`tomlWorld`; lock-step per Decision 1), plus `emit`. Implementation:

```go
package builder

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

type emitRegion struct {
    ID         string         `toml:"id"`
    Name       string         `toml:"name"`
    Hexes      [][2]int       `toml:"hexes"`
    Boundaries []emitBoundary `toml:"boundary"`
}

type emitBoundary struct {
    FromQ    int    `toml:"from_q"`
    FromR    int    `toml:"from_r"`
    ToRegion string `toml:"to_region"`
    ToQ      int    `toml:"to_q"`
    ToR      int    `toml:"to_r"`
}

type emitWorld struct {
    ID         string `toml:"id"`
    Name       string `toml:"name"`
    TechLevel  int    `toml:"tech_level"`
    Population int    `toml:"population"`
    Region     string `toml:"region"`
    HexQ       int    `toml:"hex_q"`
    HexR       int    `toml:"hex_r"`
}

func emit(dstDir string, derived *derivedMap) error {
    if err := writeRegions(filepath.Join(dstDir, "regions.toml"), derived); err != nil {
        return err
    }
    if err := writeWorlds(filepath.Join(dstDir, "worlds.toml"), derived); err != nil {
        return err
    }
    return nil
}

func writeRegions(path string, derived *derivedMap) error {
    out := struct {
        Region []emitRegion `toml:"region"`
    }{}
    for _, r := range derived.Regions {
        em := emitRegion{ID: r.ID, Name: r.Name}
        for _, h := range r.Hexes {
            em.Hexes = append(em.Hexes, [2]int{h.Q, h.R})
        }
        for _, b := range r.Boundaries {
            em.Boundaries = append(em.Boundaries, emitBoundary{
                FromQ: b.From.Q, FromR: b.From.R,
                ToRegion: b.ToRegion,
                ToQ:      b.To.Q, ToR: b.To.R,
            })
        }
        out.Region = append(out.Region, em)
    }
    return writeTOML(path, out)
}

func writeWorlds(path string, derived *derivedMap) error {
    out := struct {
        World []emitWorld `toml:"world"`
    }{}
    for _, w := range derived.Worlds {
        out.World = append(out.World, emitWorld{
            ID: w.ID, Name: w.Name,
            TechLevel: w.TechLevel, Population: w.Population,
            Region: w.Region,
            HexQ:   w.Coord.Q, HexR: w.Coord.R,
        })
    }
    return writeTOML(path, out)
}

func writeTOML(path string, v any) error {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create %s: %w", path, err)
    }
    defer f.Close()
    if err := toml.NewEncoder(f).Encode(v); err != nil {
        return fmt.Errorf("encode %s: %w", path, err)
    }
    return nil
}
```

##### Task 1.6 — `internal/spatial/builder/builder.go`

Public entrypoint, orchestration, and self-check. Implementation:

```go
package builder

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Options struct {
    SourceDir string
    DestDir   string
    Replace   bool
}

type Summary struct {
    Regions        []string
    WorldCount     int
    AdjacencyCount int
    WarpCount      int
}

func Build(opts Options) (Summary, error) {
    if err := checkOutputs(opts); err != nil {
        return Summary{}, err
    }
    layout, err := parseLayout(opts.SourceDir)
    if err != nil {
        return Summary{}, err
    }
    data, err := parseData(opts.SourceDir)
    if err != nil {
        return Summary{}, err
    }
    derived, err := derive(layout, data)
    if err != nil {
        return Summary{}, err
    }
    if err := emit(opts.DestDir, derived); err != nil {
        return Summary{}, err
    }
    if err := selfCheck(opts.DestDir); err != nil {
        return Summary{}, err
    }
    return summarize(derived, len(data.Warps)), nil
}

func checkOutputs(opts Options) error {
    if opts.Replace {
        return nil
    }
    for _, name := range []string{"regions.toml", "worlds.toml"} {
        path := filepath.Join(opts.DestDir, name)
        if _, err := os.Stat(path); err == nil {
            return fmt.Errorf("%w: %s", ErrOutputExists, path)
        } else if !os.IsNotExist(err) {
            return fmt.Errorf("stat %s: %w", path, err)
        }
    }
    return nil
}

func selfCheck(dstDir string) error {
    if _, err := spatial.LoadRegionMap(dstDir); err != nil {
        return fmt.Errorf("internal: canonical output failed self-check (%w); please file a bug", err)
    }
    return nil
}

func summarize(derived *derivedMap, warpCount int) Summary {
    s := Summary{
        WorldCount:     len(derived.Worlds),
        WarpCount:      warpCount,
        AdjacencyCount: derived.AdjacencyCount,
    }
    for _, r := range derived.Regions {
        s.Regions = append(s.Regions, r.ID)
    }
    return s
}
```

##### Task 1.7 — `internal/spatial/builder/testdata/`

Create fixture directories:

- `valid/helions_reach/source/{layout.txt, data.toml}` — copy the example from Discovery's Per-Area Design Details section. `golden/{regions.toml, worlds.toml}` generated via `-update` flag after Tasks 1.1–1.6 land.
- `errors/layout_syntax/` — `layout.txt` with a multi-char cell or invalid glyph; `data.toml` minimal; `err.txt` containing `layout.txt:`.
- `errors/missing_region_entry/` — layout uses region `B` but `data.toml` only declares `[regions.A]`; `err.txt` containing `data.toml:` and `B`.
- `errors/ambiguous_world/` — marker `1` placed at a cell touching both `A` and `B`; `data.toml` omits the `region` override; `err.txt` containing `multiple region candidates`.
- `errors/unknown_marker/` — `[[worlds]]` entry with `at = "q"`, but glyph `q` never appears in layout; `err.txt` containing `data.toml:[[worlds]] id=` and `q`.
- `errors/warp_wrong_region/` — `[[warps]]` endpoint at `{region = "A", row = N, col = M}` where the addressed cell actually belongs to region `B`; `err.txt` containing `data.toml:[[warps]] #`.

##### Task 1.8 — `internal/spatial/builder/{builder,layout,derive,emit}_test.go`

- `builder_test.go`:
  - `TestBuild_Valid` — iterates `testdata/valid/*`; runs `Build`; compares emitted output against `golden/`; honors `-update` flag.
  - `TestBuild_SelfCheckRoundTrip` — runs `Build` on `valid/helions_reach`, then `spatial.LoadRegionMap(dstDir)`; asserts no error.
  - `TestBuild_MissingSource` — empty source dir; assert `errors.Is(err, ErrMissingSource)`.
  - `TestBuild_OutputExists` — pre-create `regions.toml`; assert `errors.Is(err, ErrOutputExists)` without `--replace`; assert success with.
  - `TestBuild_Errors` — table-driven over `testdata/errors/*`; assert `strings.Contains(err.Error(), readErrFile(case))`.
- `layout_test.go` — table-driven inline-string tests: comment lines, blank lines, odd-row indent strip, trailing whitespace tolerance, duplicate marker detection, invalid glyph rejection.
- `derive_test.go` — table-driven cases for region inference (single-candidate, multi-candidate error, zero-candidate error, explicit override honored, override-naming-non-adjacent error if applicable), adjacency canonicalization (A↔B emits on A side only).
- `emit_test.go` — small `derivedMap`; emit; re-decode via `toml.DecodeFile`; compare field round-trip.

`-update` flag at package init in `builder_test.go`:

```go
var update = flag.Bool("update", false, "regenerate testdata/valid/*/golden files")
```

---

### Commit 2 — `feat(spatial): build subcommand`

Wire the cobra `spatial` group and `build` subcommand. CLI is thin; all logic lives in `internal/spatial/builder`.

##### Task 2.1 — `cmd/gm-toolkit/spatial/spatial.go`

New cobra group, mirroring `cmd/gm-toolkit/campaign/campaign.go`:

```go
package spatial

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "spatial",
        Short: "Author and build spatial maps",
        Long:  "Build the canonical regions.toml / worlds.toml from GM-authored layout.txt + data.toml sources.",
    }
    cmd.AddCommand(newBuildCmd())
    return cmd
}
```

##### Task 2.2 — `cmd/gm-toolkit/spatial/build.go`

The `build` subcommand. Mirrors `add_rules.go` for campaign resolution and `--replace` handling. Fixed source/dest paths under `<campaign>/spatial/source/` and `<campaign>/spatial/` per Discovery Decision 10.

```go
package spatial

import (
    "errors"
    "fmt"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
    "github.com/therobertcrocker/gm-toolkit/internal/spatial/builder"
)

func newBuildCmd() *cobra.Command {
    var (
        campaignID string
        replace    bool
    )
    cmd := &cobra.Command{
        Use:   "build",
        Short: "Build canonical regions.toml + worlds.toml from spatial/source/ inputs",
        RunE: func(cmd *cobra.Command, args []string) error {
            reg, err := campaigns.LoadRegistry()
            if err != nil {
                return fmt.Errorf("spatial build: %w", err)
            }
            camp, err := campaigns.ResolveActive(reg, campaignID)
            if err != nil {
                if errors.Is(err, campaigns.ErrNoActiveCampaign) {
                    return fmt.Errorf("spatial build: no active campaign and no --campaign specified\n  hint: gm-toolkit campaign set-active --campaign <id>")
                }
                return fmt.Errorf("spatial build: %w", err)
            }
            opts := builder.Options{
                SourceDir: filepath.Join(camp.Root, "spatial", "source"),
                DestDir:   filepath.Join(camp.Root, "spatial"),
                Replace:   replace,
            }
            summary, err := builder.Build(opts)
            if err != nil {
                switch {
                case errors.Is(err, builder.ErrMissingSource):
                    return fmt.Errorf("spatial build: %w\n  hint: create %s/layout.txt and data.toml", err, opts.SourceDir)
                case errors.Is(err, builder.ErrOutputExists):
                    return fmt.Errorf("spatial build: %w\n  hint: pass --replace to overwrite", err)
                default:
                    return fmt.Errorf("spatial build: %w", err)
                }
            }
            printSummary(cmd, camp.ID, summary, opts.DestDir)
            return nil
        },
    }
    cmd.Flags().StringVar(&campaignID, "campaign", "", "target campaign id (defaults to active)")
    cmd.Flags().BoolVar(&replace, "replace", false, "overwrite existing regions.toml / worlds.toml")
    return cmd
}

func printSummary(cmd *cobra.Command, campaignID string, s builder.Summary, dstDir string) {
    out := cmd.OutOrStdout()
    fmt.Fprintf(out, "Built spatial map for campaign %s:\n", campaignID)
    fmt.Fprintf(out, "  %d regions (%s)\n", len(s.Regions), strings.Join(s.Regions, ", "))
    fmt.Fprintf(out, "  %d worlds\n", s.WorldCount)
    fmt.Fprintf(out, "  %d adjacency boundaries\n", s.AdjacencyCount)
    fmt.Fprintf(out, "  %d warps\n", s.WarpCount)
    fmt.Fprintf(out, "Wrote %s/regions.toml\n", dstDir)
    fmt.Fprintf(out, "Wrote %s/worlds.toml\n", dstDir)
}
```

##### Task 2.3 — `cmd/gm-toolkit/root.go`

Register the new group alongside `campaign` and the existing `faction` command:

```go
import (
    "github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/campaign"
    "github.com/therobertcrocker/gm-toolkit/cmd/gm-toolkit/spatial"
)

func NewRootCmd() *cobra.Command {
    root := &cobra.Command{ /* ... */ }
    root.AddCommand(campaign.NewCmd())
    root.AddCommand(spatial.NewCmd())
    root.AddCommand(newFactionCmd())
    return root
}
```

##### Task 2.4 — Manual smoke

After Commit 2 lands, manual smoke (not a checked-in test):

```
mkdir -p $GM_TOOLKIT_HOME/<camp>/spatial/source
# author layout.txt + data.toml (copy from testdata/valid/helions_reach if quick check)
gm-toolkit spatial build --campaign <camp>
# inspect <camp>/spatial/{regions.toml,worlds.toml}
gm-toolkit spatial build --campaign <camp>          # expect ErrOutputExists + --replace hint
gm-toolkit spatial build --campaign <camp> --replace # expect success
```

No automated CLI test; convention (per the rest of the project) is that the cobra wiring is exercised by hand and the logic is unit-tested inside `internal/`.

---

## Pre-Merge Checklist Additions

Standard checklist from CLAUDE.md applies. Initiative-specific notes:

- Verify the Helion's Reach golden fixture by hand once after first `-update` generation (sanity-check the canonical TOML for one or two boundaries and one world).
- Confirm `planned-work.md` Backlog gains the engine follow-up "adjacency vs warp cost distinction" (out-of-scope item from Discovery) if not already present.
- Confirm the initiative entry is removed from the Planned Initiatives table on merge.

## Model Selection

| Mode | Model |
|------|-------|
| Plan (this session) | Opus |
| Commit 1 execution | Sonnet (default for execution) — Opus only if the parser/derive design hits surprises |
| Commit 2 execution | Sonnet |
| Pre-merge checklist | Sonnet |
