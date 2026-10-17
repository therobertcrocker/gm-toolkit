# Contributing — Spatial Maps

> How to author a campaign's spatial map. Read the
> [contributing overview](overview.md) first for setup; this guide is a
> walkthrough — you build a small two-region map from nothing, one construct at a
> time, and run the builder after each step. Shape & why:
> [spatial](../architecture/engine/spatial.md).

## Orientation

A campaign's map is **generated, not written**. You author two source files — an
ASCII hex grid (`layout.txt`) and a TOML data file (`data.toml`) — and
`gm-toolkit spatial build` derives the canonical `regions.toml` and `worlds.toml`
the engine loads. You never write an axial coordinate by hand.

```
<campaign>/spatial/
├── source/
│   ├── layout.txt      you write this
│   └── data.toml       you write this
├── regions.toml        generated — do not edit
└── worlds.toml         generated — do not edit
```

The build is one-way. There is no reverse conversion, so a hand-edit to a
generated file is lost the next time you build. Edit the source pair and rebuild.

<br/>
<br/>

# The walkthrough

The finished map is two regions separated by a void, one world in each, and one
warp across the gap. Every command and every output below is real — run them in
order and you will see the same thing.

<br/>

## Step 1 — Scaffold the source directory

`campaign create` makes `spatial/`, but not `spatial/source/`. Make it yourself:

```sh
mkdir -p ~/games/drift/spatial/source
```

Both source files must exist before the builder runs. A missing one is an error,
not an empty default.

<br/>

## Step 2 — Draw the grid

`layout.txt` is a grid of single-character cells separated by **one space**. Odd
grid rows carry **one leading space** — that indent renders the pointy-top hex
stagger, so what you type has the topology of the map it describes. The parser
strips it.

```
# Sector: Tannhauser Drift
# A = Ember Reach   B = Cold Harbor
. A A . . . . .
 . A A . . . B B
. . A A . . B . .
```

- `.` is an empty hex — void, no region, not traversable.
- `A`–`Z` is region membership. Every cell with that letter belongs to that region.
- Lines starting with `#` are comments. Blank lines are ignored.

Comments and blank lines **do not count as grid rows**. Row parity — and so which
rows take the leading space — is counted over grid rows only. The two comment
lines above are invisible to it; the first `.` row is row 0.

Rows do not have to be the same length. A short row simply has fewer cells.

<br/>

## Step 3 — Name the regions

Every region letter in the layout needs an entry in `data.toml`:

```toml
[regions.A]
name = "Ember Reach"

[regions.B]
name = "Cold Harbor"
```

The check runs both ways. A letter in the layout with no entry is an error, and
an entry with no cells in the layout is an error too — a region cannot be
declared and unplaced.

That is enough to build:

```sh
gm-toolkit spatial build --campaign drift
```

```
Built spatial map for campaign drift:
  2 regions (A, B)
  0 worlds
  0 warps
Wrote ~/games/drift/spatial/regions.toml
Wrote ~/games/drift/spatial/worlds.toml
```

`regions.toml` now holds the axial coordinates the builder derived for you:

```toml
[[region]]
  id = "A"
  name = "Ember Reach"
  hexes = [[1, 0], [1, 1], [1, 2], [2, 0], [2, 1], [2, 2]]

[[region]]
  id = "B"
  name = "Cold Harbor"
  hexes = [[5, 2], [6, 1], [7, 1]]
```

<br/>

## Step 4 — Place the worlds

A world is a **marker glyph** in the layout — `1`–`9`, then `a`–`z` — and a
matching `[[worlds]]` entry keyed by that glyph. Replace one hex in each region:

```
# Sector: Tannhauser Drift
# A = Ember Reach   B = Cold Harbor
. A A . . . . .
 . A A . . . B B
. . A 1 . . B 2 .
```

```toml
[[worlds]]
id = "ember-prime"
name = "Ember Prime"
at = 1
tech_level = 4
population = 60000

[[worlds]]
id = "cold-harbor-station"
name = "Cold Harbor Station"
at = 2
tech_level = 3
population = 12000
```

`at` is the glyph: an integer for `1`–`9`, a single-character string for `a`–`z`.
Each glyph appears once in the layout and once in the data file — the pairing is
the only link between the two files.

**The region is inferred.** Notice that neither entry names a region. The builder
reads the marker's six neighbors and collects the distinct *region letters* among
them. Empty hexes and other world markers contribute nothing — they are
transparent, so two adjacent worlds never infer through each other.

- **Exactly one letter** — that region is assigned. Write nothing.
- **Two or more** — error. Add `region = "X"` to the entry to choose.
- **None** — error. The marker is floating in void; add `region = "X"`.

The marker's own hex joins whichever region wins, so a world is always inside its
region. Rebuild, and `worlds.toml` carries the placement:

```toml
[[world]]
  id = "ember-prime"
  name = "Ember Prime"
  tech_level = 4
  population = 60000
  region = "A"
  hex_q = 2
  hex_r = 2
```

<br/>

## Step 5 — Mark the warps

Movement between adjacent hexes is free geometry — the engine derives it and
nothing is stored. **Warps are the only stored edges**, and you do not write them
either. You mark endpoints and the builder connects them.

Add a `*` suffix to two hexes, one in each region:

```
# Sector: Tannhauser Drift
# A = Ember Reach   B = Cold Harbor
. A A . . . . .
 . A A* . . . B* B
. . A 1 . . B 2 .
```

The suffix goes on a region glyph or a world glyph — `A*` and `2*` are both
valid. The rules the builder applies:

- **A mark must sit on a region boundary** — at least one of its six neighbors is
  a different region, void, or off-grid. An interior mark is a hard error.
- **Two marks connect when they are in different regions and the hexes between
  them are empty.** Line of sight is a straight hex line; any occupied hex on it
  blocks the pair.
- **Adjacent marks produce no warp.** They already connect by an ordinary step,
  so the edge would be redundant.
- **A mark that reaches nothing is a warning, not an error** — the build still
  succeeds. It usually means a region got in the way.

Both marks here sit on the void edge with three empty hexes between them, so they
pair:

```
Built spatial map for campaign drift:
  2 regions (A, B)
  2 worlds
  1 warps
```

```toml
[[warp]]
  from_region = "A"
  from_q = 2
  from_r = 1
  to_region = "B"
  to_q = 6
  to_r = 1
```

Because connection depends on what lies *between* marks, the drawing is
load-bearing. Widening a region across the gap can silently drop a warp — the
build still passes, with a warning.

<br/>

## Step 6 — Build and iterate

```
gm-toolkit spatial build [--campaign <id>] [--replace]
```

- `--campaign <id>` targets a campaign. Without it, the active campaign is used.
- `--replace` overwrites existing output. Without it, an existing `regions.toml`
  or `worlds.toml` stops the build — so the first build needs no flag and every
  rebuild needs `--replace`.

The loop is: edit the source pair, run `spatial build --replace`, read the
summary. Warnings go to stderr; the exit code is non-zero only on a real error.

Two properties worth trusting. The builder **reloads its own output** through the
engine's loader before reporting success, so a build that succeeds produces a map
the engine can read — a failure there is reported as a toolkit bug, not your
mistake. And the build is **deterministic**: regions, worlds, warps, and hexes
are all sorted before writing, so the same source always produces byte-identical
output. Generated files diff cleanly in version control.

<br/>
<br/>

# When it goes wrong

Every error names the *source* file you edit, with a line and column for layout
problems. Errors that name a generated file are toolkit bugs. (The CLI prints
command usage after the error; only the message is shown here.)

| Message | Cause | Fix |
|---------|-------|-----|
| `layout.txt:4:18: invalid glyph '?'` | A character that is not `.`, `A`–`Z`, `1`–`9`, or `a`–`z` | Use a valid glyph, or `.` for void |
| `layout.txt:3:14: expected single-space separator, got '2'` | Two cells written with no space between them | One space between every pair of cells |
| `layout.txt:1:13: world marker '1' appears twice; first at line=1 col=5` | The same marker glyph used in two places | A marker is unique per map — worlds do not share glyphs |
| `layout.txt:5:7: world marker "1" has no [[worlds]] entry in data.toml` | A marker in the grid with no data entry | Add the `[[worlds]]` entry, or make the cell a region letter |
| `data.toml:[regions.B]: missing entry; region "B" is used in layout` | A region letter with no `[regions.X]` block | Add the block, or remove the letter from the layout |
| `data.toml:[[worlds]] id="x": references marker "7" but no such marker in layout.txt` | A data entry whose `at` glyph is not in the grid | Fix `at`, or place the marker |
| `data.toml:[[worlds]] id="x": marker "3" ... has multiple region candidates [A B]` | The marker touches two regions, so inference is ambiguous | Add `region = "X"` to the entry |
| `data.toml:[[worlds]] id="x": marker ... has no region-letter neighbors` | The marker is surrounded by void | Add `region = "X"`, or move it beside its region |
| `data.toml: hex at (row=1, col=2) in region "A" is marked for warp but is not on a region boundary` | A `*` on a hex with six same-region neighbors | Move the mark to the region's edge |
| `canonical output exists: .../regions.toml` | Rebuilding without the flag | Pass `--replace` |
| `missing source file: .../layout.txt` | The source pair is absent or misplaced | Create both files under `<campaign>/spatial/source/` |

**Warning: `hex A(1,0) is marked for warp but reaches no other region`.** Not an
error — the build succeeded and simply produced no warp for that mark. Either an
occupied hex blocks the line, or there is no mark in another region to reach.
Check what sits between the two endpoints.

<br/>
<br/>

# Reference

**Layout glyphs**

| Glyph | Meaning |
|-------|---------|
| `.` | Empty hex — void |
| `A`–`Z` | Region membership |
| `1`–`9`, `a`–`z` | World marker |
| `*` | Warp anchor — a suffix on a region or world glyph |
| `#` | Comment line |

**`data.toml` fields**

| Field | Table | Required | Notes |
|-------|-------|----------|-------|
| `name` | `[regions.X]` | yes | Display name; `X` is a single uppercase letter |
| `id` | `[[worlds]]` | yes | Stable world identifier used by the engine |
| `name` | `[[worlds]]` | yes | Display name |
| `at` | `[[worlds]]` | yes | Marker glyph — integer `1`–`9` or string `"a"`–`"z"` |
| `tech_level` | `[[worlds]]` | yes | World metadata |
| `population` | `[[worlds]]` | yes | World metadata |
| `region` | `[[worlds]]` | no | Region override; needed only when inference is ambiguous |

**Capacity.** 26 regions and 35 world markers per map — the alphabet is the
limit. A campaign has one map.

**See also:** [spatial](../architecture/engine/spatial.md) for the derivation
pipeline and why warps are the only stored edges;
[world movement](../architecture/engine/world-movement.md) for what the engine
does with the map at runtime.
