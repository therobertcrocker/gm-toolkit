# Migrate Data Files to TOML 1.1 — Exploration

> **Initiative:** `R.029`, a refactor whose Discovery runs as an exploration (Robert's call).
> It uses the sections of `E.002` and `E.004`, and it adds to the
> [`D.009` findings log](../discovery/D.009-exploration-type-findings.md) at close.
> **Feeds:** the `R.029` plan, and `A.001.8` D-17 (goal parts as TOML sub-tables).
> **Status:** complete (signed off by Robert, 2026-11-06). One session, on
> `feature/data-driven-goals`. Rows are filed.

## In Plain Words

- **What:** the hand-written data files move from TOML 1.0 to TOML 1.1, so the nested
  parts of an entry can be written in braces over several lines under one header. The
  exploration asks if the loader copes, what shape is best for a GM who edits by hand,
  and what else the change touches.
- **Short answer:** the loader already copes, with no code change. Each entry becomes
  one block (one header, every part inline, a short house style). Goals are written
  that way from the start, which amends `A.001.8` D-17. The machine-written
  `state.toml` stays TOML, but small types write themselves inline, so it loses its
  nested headers.
- **Why it matters:** an asset reads top to bottom like its rulebook page and moves as
  one unit. Today one asset can carry nine headers that repeat its key (Book of Secrets).
- **How it lands:** `R.029` goes straight to Plan; `A.001.8`'s row picks up the goal
  shape before its Plan. No order is forced between them.

<br/>

## Questions

Robert's prompt, from the tracker row: write nested parts (`attack`, `ability`,
`effects`, `filters`, `gates`) as multi-line inline tables, one header per entry.
`BurntSushi/toml` v1.6.0 already parses 1.1, and a trial of `cunning_assets.toml`
decoded identical (594 to 504 lines, 88 to 24 headers). Machine-written `state.toml`
stays as the encoder writes it. Run the real loader on a 1.1 file first.

1. **Loader.** Does every real loader (rulebook, spatial, seeds, scenarios) accept a
   1.1 file and decode it the same, strict unknown-key checks included?
2. **Inventory.** Which hand-written files exist, and which nested parts does each use?
3. **Shape.** Which parts go inline, and which (if any) keep their own header?
4. **D-17.** Do `A.001.8`'s goal parts become inline tables too, and how does this work
   sit in order with `A.001.8` Effort 1?
5. **Reach.** Does anything else read or write these files (compose, test fixtures,
   TOML strings in Go tests, scripts), and does it need to change?
6. **State.** Added after the audit, from Robert: should the machine-written files be
   1.1 too, and should state use a format other than TOML?

**Deciding criterion, from Robert:** best for a GM who edits the files by hand. Human
use comes first; line and header counts matter only where they serve it.

<br/>

## Current State

### The loader reads 1.1 today

Run on 2026-11-06 with a throwaway program (`scratch/loadcheck`, gitignored) that calls
the real `rulebook.Load` twice: once on `rulebooks/swn` plus `_core/goals.toml`, once
on the same folder with `cunning_assets.toml` swapped for the 1.1 trial
(`scratch/toml-format/cunning_assets.v11.toml`).

| Check | Result |
|---|---|
| 1.0 against 1.1, whole `Rulebook` | Identical (`reflect.DeepEqual`), 73 assets each |
| Misspelt key inside an inline table (`ownr`) | Refused by `refuseUnknownKeys`, with the full path (`assets.FalseFront.effects.filters.ownr`) |
| Missing comma in a multi-line inline table | Decode error with the line and the last key: "expected a comma or an inline table terminator" |
| A `[header]` that adds to an inline table | Decode error with the line: "Key 'assets.Smugglers.attack' has already been defined" |

- **TOML 1.1 is a superset of 1.0.** Every 1.0 file is a valid 1.1 file, so a file can
  mix both styles, and nothing marks a file's version.
- **One rule a GM must learn.** Once a part is an inline table, every key of that part
  goes inside its braces. A later header cannot add to it (row 4 above).

### Inventory

Six packages import `BurntSushi/toml`: `campaigns`, `faction/rulebook`, `faction/state`,
`spatial`, `spatial/builder` (production), and `faction/domain` (tests only).

| Files | Written by | Size | Nested parts | Headers |
|---|---|---|---|---|
| `rulebooks/swn/assets/*_assets.toml` | Hand | 3 files, 73 assets, 1,772 lines | `attack`, `transport`, `ability` (`steps`, `outcomes`, `relocate`, `target_roll`, `purchase`), `effects` (`filters`, `gates`, `gates.contest`, one parameter table per effect) | 253 |
| `rulebooks/swn/tags.toml` | Hand | 20 tags, 418 lines | `effects` (`filters`, `gates`, parameter tables), `ability` | 107 |
| `rulebooks/_core/goals.toml` | Hand | 12 goals, 71 lines | None today; `A.001.8` adds six parts and their filters | 12 |
| `rulebooks/swn/settings.toml` | Hand | 10 lines | `purchase_population` | 1 |
| `rulebooks/swn/rulebook.toml` | Hand | 2 lines | None | 0 |
| `campaigns/*/spatial/source/data.toml` | Hand | 80 lines | `regions.<letter>` (one key each), `[[worlds]]` (flat) | 14 |
| `campaigns/*/state/scenarios/*.toml` | Hand, in the state shape | 5 per campaign, 1,010 lines | `factions.<id>`, `assets`; `homeworld` and `bases` are already inline, on one line each | 12 to 22 |
| `campaigns/*/state/state.seed.toml` | `reset-campaign.sh` copies a scenario; `ensureSeed` writes one if none exists | 181 lines | As scenarios | 18 |
| `rulebook/testdata`, `spatial/builder/testdata/*/source` | Hand | Small | As their live files | Few |
| TOML strings in Go tests | Hand | `effects_test.go` (93 headers), `rulebook_test.go` (3) | Assets, tags | 96 |

**Machine-written, out of scope by the row (question 6 brings `state.toml`'s shape in):** `state.toml` (`state.Save`), `regions.toml`
and `worlds.toml` (`spatial/builder/emit.go`, golden files in builder testdata),
`campaigns-registry.toml` (`campaigns/registry.go`) and `campaign.toml`
(`campaigns/campaigns.go`). All five use `toml.NewEncoder`, which writes 1.0 style.
Scenarios become `state.toml` at reset, and the first save rewrites it in encoder style;
the scenario file keeps its own style.

**Campaign rulebooks are byte copies.** `ComposeRulebook` copies the source files as
they are, so a campaign gets the new style when it re-composes. `sandbox-claude`'s
`tags.toml` already differs from `rulebooks/swn/tags.toml`, so it is stale today.

### Findings from the audit

- **D-17's reason was a 1.0 limit.** `A.001.8` D-17 rejected inline tables as "hard to
  read once filters nest". Under 1.0 an inline table sits on one line. Under 1.1 it can
  span lines. No goal has a nested part in the files yet (Effort 1 writes them), so
  the goal shape costs nothing to change before Effort 1 starts.
- **`[[effects]]` is a name, not only syntax.** Twelve live docs and Go files (among
  them `effect-mutation.md`, `ability-catalog.md`, `orchestrator.md`) call the effects
  model "the `[[effects]]` list". In 1.1 style the files write `effects = [ … ]`.
- **The trial made style choices beyond the row.** It aligns `=` within a block, uses
  dotted keys inside inline tables (`gates.contest = { … }`) and at entry level
  (`ability.steps = [ … ]`), and keeps short parts on one line
  (`attack = { … }`). The longest line (`gates.contest`) runs to about 110 columns.
- **The encoder cannot write 1.1.** `BurntSushi/toml` v1.6.0's `Encoder` has one
  option (`Indent`). It writes every struct or map as a `[header]` and goes inline only
  for tables inside an array, always on one line. So `state.toml` gives each asset
  location three nested headers (`location`, `location.region_hex`,
  `location.region_hex.Coord`). The scenarios are hand-compacted copies of that shape.
  The one hook is `Marshaler`: `MarshalTOML` bytes are written raw as the value
  (`encode.go:236`), so a type can write itself as an inline table. It does not know
  its indent level.
- **The machine-written files differ in who reads them.** `regions.toml` and
  `worlds.toml` are builder output with golden tests; a GM edits `source/data.toml`.
  `campaigns-registry.toml` and `campaign.toml` are flat, so 1.1 changes nothing in
  them. Only `state.toml` is read by people, through scenarios and seeds.
- **Not checked:** how TomlSmith's formatter treats multi-line inline tables on save.

<br/>

## Existing Patterns

- **The 1.1 trial** (`scratch/toml-format/cunning_assets.v11.toml`, gitignored) — one
  header per asset, short parts inline on one line, longer parts and every `effects`
  element spread over lines with a trailing comma. 594 to 504 lines, 88 to 24 headers.
- **Inline tables in the state shape** — scenarios already write `homeworld` and every
  `bases` entry as one-line inline tables (up to 243 columns). This shows the
  1.0 limit from the reader's side.
- **Strict loading** — `refuseUnknownKeys` reads decoder metadata, not the file text,
  so it works unchanged for either style (Current State, row 2).
- **Format-neutral compose** — `ComposeRulebook` copies bytes and parses only
  `rulebook.toml`, so it needs no change.

<br/>

## Options

### State output (question 6)

Shared line: the builder output, the registry and `campaign.toml` stay as they are.
The builder files are read by the program and pinned by golden tests, and the other two
are flat. Only `state.toml` is read by people, through the scenarios and seeds that
`reset-campaign.sh` copies into place, so scenarios and state must share one format.

#### S-0 — No change

The encoder keeps writing a `[header]` per struct: three nested headers per asset
location. Scenario authors keep compacting by hand.

#### S-1 — `MarshalTOML` on the deep small types (chosen)

A type such as `domain.Location` implements `toml.Marshaler` and writes itself as a
one-line inline table: `location = { world_id = "mireholt", region_hex = { … } }`.
The encoder writes the bytes as they are (`encode.go:236`), and the decoder reads them
through the struct fields as today. The output is valid 1.0 and 1.1. It removes the
header depth, which is the real cost, at the cost of a small method and a round-trip
test. A scenario author can still write a scenario in full 1.1 style by hand.

#### S-2 — A state writer of our own

True multi-line 1.1, laid out as we like. A few hundred lines on the save path, which
runs at every phase gate, plus a round-trip test, kept for as long as state is TOML.

#### S-3 — Another format for state

Weighed against the persistence page's Key Decision "TOML for static and state, JSONL
for history" (frozen log D3: TOML is hand-editable; D195: state is a resume
checkpoint).

| Format | For | Against |
|---|---|---|
| JSON | Standard library; `MarshalIndent` is stable; matches `history.jsonl` | No comments, and every scenario opens with a comment block |
| YAML | Comments, compact nesting | Known traps (`no` reads as false, indentation), a new dependency, a third format |
| SQLite | Robust, queryable | Not hand-editable |

Rejected: it reverses a Key Decision and hurts scenario authors, the only people who
edit state, to fix a shape problem that S-1 fixes inside TOML. Worth a fresh look if
Edit Mode (`A.002`) makes state owned by the program alone.

**Robert's call (2026-11-06):** S-1.

### Terms

- **Entry** — one named record: an asset, a tag, a goal, a world.
- **Part** — a nested table inside an entry: `attack`, `ability`, `effects`, `filters`,
  `gates`, `outcomes`.
- **Inline table** — a part written in braces as a value, `attack = { … }`, not under
  its own `[header]`. In 1.1 it may span lines and end each line with a comma.
- **Dotted key** — `gates.contest = { … }`, a second way to write
  `gates = { contest = { … } }`.

### Entry shape (question 3)

#### H-0 — Stay on 1.0 headers

Today's files. Every part has its own header: 253 asset headers, 107 tag headers.

#### H-1 — Entry headers plus one header per effect

Each entry keeps its header, and each `effects` element keeps `[[assets.X.effects]]`.
Everything below an effect, and every other part, goes inline.

```toml
[[assets.FalseFront.effects]]
trigger = "lethal_damage"
effect  = "sacrifice_self"
filters = { owner = "own", scope = "world" }
```

- **For:** fewer braces and commas; the `[[effects]]` name stays literal.
- **Against:** each effect header repeats the entry's key, so a typo there makes a new,
  broken entry; an entry is not one block, so copying an asset means copying several.
  About 108 asset headers and 46 tag headers remain.

#### H-2 — One header per entry, every part inline (the trial's shape)

An entry is one block. Indentation shows the nesting, and a list part such as
`outcomes` reads as a table of rows.

```toml
[assets.VentureCapital]
id = "SWN-W6-001"
# … the entry's own keys, as today …
attack  = { attacker_stat = "Wealth", defender_stat = "Wealth", damage = "2d6" }
ability = {
  effect   = "coin_delta",
  die      = "1d8",
  outcomes = [
    { from = 1, to = 1, destroy_self = true },
    { from = 2, to = 3, coin = 1 },
    { from = 4, to = 7, coin = 2 },
    { from = 8, to = 8, coin = 3 },
  ],
}
```

- **For:** one block per entry, the same as the rulebook page; copy, move or delete an
  asset as one unit; no repeated keys; the row's stated aim.
- **Against:** braces and commas to keep right. A slip gives a clear error with a line
  number (Current State, rows 3 and 4).

#### House style under H-2

Each line is a separate call. The lean is mine.

| Call | Choices | Lean |
|---|---|---|
| Line rule | One line when the part fits a column limit; otherwise one key per line, each with a trailing comma | Yes |
| Column limit | 100 or 120 | 120: the longest gate (`gates = { contest = { … } }` in an effect) fits on one line, at 119 columns |
| List parts | `effects`, `steps`, `outcomes` always open over lines; each element follows the line rule | Yes |
| Dotted keys | Never, or allowed for a part with one key (`gates.contest`, `ability.steps`) | Never: one way to nest, and braces show it |
| Alignment | Align `=` inside each multi-line table, or never | Align, as the 1.0 files already do in `effects`; check that TomlSmith's formatter keeps it |
| Key order | The entry's own keys, then parts in a fixed order (`attack`, `transport`, `ability`, `effects`) | Yes, as the trial does |

**Robert's calls (2026-11-06):** H-2, with every lean above: a 120-column limit, no
dotted keys, aligned `=`.

### Goals and D-17 (question 4)

The goal file has no parts yet, and `A.001.8` Effort 1 writes them. TOML 1.0 and 1.1
mix in one file, and the loader reads both, so goals need the style decided, not
`R.029` executed.

#### G-0 — Keep D-17

Goals use sub-tables; every other rulebook file goes inline. The goal file is the one
that reads differently.

#### G-1 — Goals take the house style, written that way by `A.001.8` Effort 1 (lean)

```toml
[goals.MilitaryConquest]
id          = "CORE-G-001"
name        = "Military Conquest"
description = "Destroy a number of Force assets of rival factions equal to your faction's Force rating."
progress    = {
  event   = "kill",
  amount  = "one",
  filters = { owner = "rival", caused_by = "own", category = ["force"] },
}
threshold   = { ratings = ["force"] }
xp          = { formula = "half_progress" }
```

D-17's reason ("hard to read once filters nest") was a 1.0 limit. The amendment goes
into `A.001.8`'s tracker row, for its Plan to apply, as `E.002`'s answer did for the
paused `A.001.4`; the discovery doc stays as written. No order is forced between the
two initiatives.

#### G-2 — Goals as sub-tables now, converted by `R.029` later

Effort 1 follows D-17 and `R.029` rewrites the goal file after. The goal file is
written twice.

**Robert's call (2026-11-06):** G-1.

### Reach (question 5)

| Item | Choices | Lean |
|---|---|---|
| TOML strings in Go tests (96 headers) | Convert, or leave as 1.0 | Leave: they test the loader, not the style, and they keep proving 1.0 still loads. Add one test that loads a multi-line inline table with trailing commas, so a library change cannot drop 1.1 unseen |
| `testdata` files | Convert, or leave | Leave, for the same reason |
| Spatial source `data.toml` | Leave, or write `[regions]` as one table of one-line entries (`A = { name = "Concord Core" }`) | One table: it reads as the map's legend. `[[worlds]]` stays as it is: one header per world already |
| `docs/contributing/spatial-maps.md` | Update its 5 TOML blocks to match | Yes, if `data.toml` changes |
| The name "`[[effects]]` list" (12 live docs and Go files) | Keep, or say "the `effects` list" | Rename under H-2, so a reader who searches a data file finds the name; the frozen and completed docs stay as written |
| Scenarios and seeds | Rewrite in 1.1 style, or leave | Leave: valid either way, already compacted by hand, and S-1 changes what the encoder writes |
| Campaign copies | Re-compose `sandbox` and `sandbox-claude` from `rulebooks/` | Yes; `sandbox` is Robert's to re-compose. This also fixes the stale `sandbox-claude` tags |
| `settings.toml`, `rulebook.toml` | Change, or leave | Leave: no parts to fold |

**Robert's calls (2026-11-06):** every lean above, the rename included. S-1 stays
inside `R.029`.

<br/>

## Recommendation

**H-2 with the house style, G-1, S-1 (chosen), and the Reach leans.** Each entry becomes
one block that a GM reads top to bottom and copies as one unit. Goals are born in that
shape, and state loses its nested headers without leaving TOML or the encoder.

### How it lands

- **`R.029` goes straight to Plan.** The audit is done and the target is chosen, so
  none of the refactor flow's reasons for a Discovery remain, as with `E.004` and
  `R.028`. The plan re-reads the code, as the refactor flow asks.
- **Units, one per phase or fewer:** the asset files, `tags.toml`, the spatial source
  and its doc, the state `MarshalTOML` with its round-trip test, and the doc rename.
  The plan opens with the TomlSmith formatter check.
- **`A.001.8` gets G-1 in its tracker row** before its Plan session, so Effort 1 writes
  goals in the house style.
- **Order:** none forced. `R.029` can run before, between or after `A.001.8`'s efforts
  on this branch.

### Calls made (Robert, 2026-11-06)

All taken through questions after the draft, and all as recommended:

- H-2, with a 120-column limit, no dotted keys and aligned `=`.
- G-1, which amends `A.001.8` D-17.
- The rename to "the `effects` list".
- S-1 stays inside `R.029`.
- Every Reach lean.
- File the `A.001.8` row change; do not file the state-format row.

<br/>

## Follow-on Rows

Filed after Robert's review, as marked.

| ID | Item | Type | Detail | Filed |
|---|---|---|---|---|
| `A.001.8` | Data-driven goals (row change) | `feature` | Add: goal parts are written as inline tables in `R.029`'s house style; amends D-17, whose reason was a TOML 1.0 limit (`R.029` exploration, G-1) | Yes |
| — | Revisit the state format | `refactor` | Trigger `when A.002 starts`: if Edit Mode makes state owned by the program alone, weigh JSON against TOML for `state.toml` (`R.029` exploration, S-3) | No, Robert's call |
