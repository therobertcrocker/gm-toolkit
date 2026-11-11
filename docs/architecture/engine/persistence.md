# Persistence & Static Data

> **Code:** `internal/campaigns/`, `internal/faction/state/`, `internal/faction/rulebook/`

## Purpose

This page covers everything durable: where the game's data lives on disk, who
owns it, and how it is loaded and saved. Three concerns share the page because
they share one principle — **the GM owns all of it, and the binary embeds none
of it**. A `campaign` is the addressing root that resolves every path; the
`state` package reads and writes the one mutable file (the factions and the
in-progress turn); the `rulebook` package loads the static, GM-authored rules.
There are no built-in defaults and no hard-coded data directory: an empty
`GM_TOOLKIT_HOME` is an error, not a fallback.

## Shape

### The campaign as addressing root

Nothing in the engine takes a bare file path. Instead a `Campaign` resolves a
campaign root into a typed **`Paths` bundle**, and that bundle is what engine
consumers receive. `Paths` is a flat struct whose fields are tagged with their
location and kind:

```go
type Paths struct {
	FactionDataDir string `path:"rulebook"            kind:"dir"`
	AssetsDir      string `path:"rulebook/assets"     kind:"dir"`
	SpatialDataDir string `path:"spatial"             kind:"dir"`
	StatePath      string `path:"state/state.toml"    kind:"file"`
	HistoryPath    string `path:"state/history.jsonl" kind:"file"`
	NarrativesDir  string `path:"narratives"          kind:"dir"`
	LogsDir        string `path:"state/logs"          kind:"dir"`
}
```

`Paths()` fills it by **reflecting over those tags** — joining each `path` to the
root — and `Scaffold()` reflects over the *same* tags to create the directory
tree. The two can never drift: adding a new campaign path is one struct field,
no other code touched. (A field missing its `path` tag panics at startup, by
design.)

Campaigns are discovered through a registry at
`$GM_TOOLKIT_HOME/campaigns-registry.toml` — a list of `{id, path}` entries plus
an `Active` pointer. `ResolveActive` takes an optional override id, falls back to
`Active`, and returns typed errors for the three failure modes a GM hits:
`ErrNoActiveCampaign`, `ErrCampaignNotRegistered`, and `ErrCampaignRootMissing`
(registered but the directory is gone). Creating a campaign is `Scaffold` (the
tree + a `campaign.toml` manifest) followed by `ComposeRulebook`. It walks a
source rulebook directory, plus each folder that the directory's `rulebook.toml`
manifest names in `includes` (one level deep), and copies every `.toml` except
the manifest into the campaign's `rulebook/`. Every check runs before any write:
a file that two folders both supply is always refused (`ErrRulebookDuplicate`),
and an existing campaign file is refused unless `replace` is set
(`ErrRulebookConflict`). A folder with no manifest copies as a plain folder.

### Mutable state: `state.toml`

`FactionState` is the entire mutable game: the `CampaignID`, the `CycleNumber`,
a `map[string]*Faction`, the government record (`Governments`, world ID → ruling
faction ID; a world with no entry is independent, governed by no tracked faction,
so a file without the table loads as "every world independent"), and the optional
`CurrentTurn` cursor (the
pause/resume state from the [turn pipeline](turn-pipeline.md)). It is one TOML file.

- **`Load` is bootstrap-friendly.** A missing `state.toml` returns an *empty*
  state, not an error, so a freshly scaffolded campaign runs without a manual
  init step. Load also normalizes nil maps (`Factions`, each faction's `Assets`)
  so callers never have to nil-check.
- **`Save` seeds on first write.** After encoding the state, `ensureSeed` writes
  `state.seed.toml` beside it *if one doesn't already exist* — a one-time
  snapshot of the starting state that `reset-campaign.sh` restores. Every
  subsequent save leaves the seed untouched.
- **A location writes itself on one line.** `domain.Location` and
  `spatial.RegionHex` implement `toml.Marshaler`, so the encoder writes each
  location as one inline table (`location = { world_id = …, region_hex = { … } }`),
  not as three nested headers. `[[bases]]` and `current_order` keep their headers.

### Atomic faction CRUD

`CreateFaction` and `DeleteFaction` apply their change to the in-memory map,
then `Save`. If the save fails they **roll the in-memory change back** — the
created entry is deleted, the deleted entry restored — so `fs.Factions` always
matches what is on disk and the caller can safely retry. Both guard their
preconditions with sentinel errors (`ErrFactionAlreadyExists`,
`ErrFactionNotFound`, `ErrInvalidFactionID`, and `ErrWorldAlreadyGoverned` when
`CreateFaction` is given a world another faction governs).

### Static rules: the rulebook

`rulebook.Load` reads the GM's static data from the campaign's rulebook dir into
one in-memory `Rulebook` — the authoritative rules source the engine queries all
turn:

- **Assets** are **globbed**: every `assets/*_assets.toml` is decoded and merged
  into one `map[string]*AssetDefinition`, with a hard error on a duplicate asset
  ID across files. This lets a GM split assets across themed files
  (`core_assets.toml`, `homebrew_assets.toml`) that merge transparently.
- **Tags, goals, drift costs** load from single files (`tags.toml`, `goals.toml`,
  `drift_costs.toml`). The tag and goal loaders refuse unknown keys and a
  duplicate ID, as the asset loader does. A goal's parts are inline tables; the
  loader checks their keys and stat names, and the [goal](goals.md) engine checks
  every other name when it is built.
- Loading is also **validating conversion**: every TOML record is converted into
  a domain type, and an unknown faction stat, asset type, asset flag (`P`/`A`/`S`),
  or ability effect is a load-time error — bad rulebook data fails loudly at
  startup, never mid-turn.

`DriftCost(rating)` indexes the drift-cost table (1-indexed), falling back to the
worst cost (`DriftCosts[0]`) for an out-of-range rating. This is the value the
[world engine](world-movement.md) feeds spatial as a warp crossing cost.

### History and save cadence

The third on-disk artifact, `history.jsonl`, is an append-only event log written
by the [turn pipeline](turn-pipeline.md), not this package — `Paths` only owns
its location. And while `state.Save` is the writer, *when* it runs is the
[orchestrator](orchestrator.md)'s decision: saves happen only at phase gates, so
on-disk state always reflects a completed phase. Both are cross-linked here
because they live under `state/`, but their logic lives on those pages.

## Key Decisions

- **GMs own all data; the binary embeds none.** Every asset, tag, goal, drift
  cost, region, and world is GM-authored TOML under a campaign root. There are no
  compiled-in defaults and no hard-coded data path — `GM_TOOLKIT_HOME` is
  required. The toolkit is an engine for *your* campaign, not a game with a fixed
  ruleset.
- **`Paths` is tag-derived, so layout lives in one place.** `Paths()` and
  `Scaffold()` both reflect over the same struct tags, so the on-disk layout is
  declared once and a new path is a single field. (`campaigns.go` comment.)
- **TOML for static and state, JSONL for history.** Hand-editable TOML suits the
  rules and the current snapshot; line-delimited JSON suits an ever-growing
  append-only event log. The formats match their access patterns. (Frozen log
  1–18.)
- **Hand-written rulebook files are TOML 1.1, one block per entry.** An asset, tag
  or goal has one header, and every nested part (`attack`, `ability`, `effects`,
  `filters`, `gates`) is an inline table under it, so a GM reads and copies an
  entry as one unit. The loader needs no code for it: `BurntSushi/toml` reads 1.1,
  1.0 files still load, and `refuseUnknownKeys` works the same inside braces. The
  house style lives in `docs/contributing/engine.md`; TomlSmith's formatter does its
  layout. Rejected: a header per effect (each repeats the entry's key, so a typo
  makes a new, broken entry). (`R.029` exploration H-2; plan Decisions 1 and 4.)
- **State stays encoder-written TOML; small types write themselves inline.** The
  encoder cannot write 1.1, so `Location` and `RegionHex` implement `MarshalTOML`
  and write one-line inline tables, valid in 1.0 and 1.1. Rejected: a state writer
  of our own (a few hundred lines on the save path) and another format for state
  (it would reverse the TOML decision and hurt scenario authors, who edit state by
  hand). (`R.029` exploration S-1, S-2, S-3; plan Decision 9.)
- **The government record holds which worlds a faction rules; the government tag
  says whether it rules any.** `FactionState.Governments` maps a world ID to the
  ruling faction's ID. A world with no entry is independent, and the GM answers
  for its government. A nil map reads as empty, so an older `state.toml` loads
  with every world independent and needs no migration. The record and the tag hold
  different facts, so neither is a copy of the other, and every writer keeps them
  coupled: Planetary Seizure added the tag with a faction's first world and removed
  it with its last, and `F.040`'s Seize Planet action takes that writer over (until
  it lands, nothing in play changes a government); faction create writes the entry for a government-tag pick, or
  for the homeworld with a `homeworld_government` tag (Colonists), and refuses a
  world another faction governs; faction delete clears the entries. The government
  tag is the one marked `government = true` in the data (`Rulebook.GovernmentTagID`),
  not a Go constant; the loader refuses a second one, and a tag that sets both
  flags (A.001.7 Effort 1 Decision Record 4; Effort 4 Decision 7). A homeworld is
  not a government and never enters the record. Rejected: the tag as a label that
  nothing reads; a tag instance per world (a `Tag` stops being a rulebook copy).
  (A.001.6 overview Decisions 8 and 10; Effort 4 Decisions 4, 7 and 15, Decision
  Record 6.)
- **Load degrades gracefully; mutations are atomic.** A missing state file is an
  empty campaign, not a crash, and faction CRUD rolls back its in-memory change
  if the save fails — disk and memory never diverge.
- **First save captures a seed.** `ensureSeed` snapshots the starting state once,
  enabling a clean reset without re-scaffolding.
- **Asset files glob and merge; rules validate at load.** Multiple
  `*_assets.toml` merge into one rulebook (duplicate IDs rejected), and every
  enum-like field is validated during conversion, so malformed rules fail at
  startup rather than during a turn.
- **A rulebook names its shared folders, and one checked step composes them.** A
  rulebook folder's `rulebook.toml` lists the folders it includes
  (`includes = ["../_core"]`). Each path resolves from that folder, so the path is
  GM data, not part of the binary. Includes are one level deep. `ComposeRulebook`
  walks every folder, runs every check, and then copies, so a refused compose
  writes nothing. A file that two folders both supply is always refused
  (`ErrRulebookDuplicate`, which wraps `ErrRulebookConflict`), even with
  `--replace`. `--replace` applies only to a clash with a file that is already in
  the campaign. The manifest is never copied, so `rulebook.Load` knows nothing
  about includes. Goals live in the shared `rulebooks/_core/`; tags stay in each
  rulebook, because tags are setting identity. Rejected: a repeatable `--rules`
  (every GM and test site repeats `_core`, and a second pass with `--replace`
  silently overwrites the core file); a rulebook file that overrides an included
  file (no present user). (A.001.4 Discovery D-1 and D-7; plan Decisions 2 to 6.)
- **Factions hold tag IDs, and tag IDs carry the rulebook prefix.**
  `Faction.Tags` is a list of IDs (`tags = ["SWN-T-020"]`), and `TagAdded` and
  `TagRemoved` carry the ID. Every reader that needs a name, text or effects reads
  the rulebook, so a rulebook change reaches a tag a faction already holds, as a
  definition change reaches a held asset. A held ID the rulebook does not know is
  skipped. Saves of the old shape moved by a sandbox re-seed, with no migration
  code. Rejected: a full `Tag` copy in state (it froze the rules at pick time).
  (A.001.7 discovery D-4 and D-20; `E.002` I-1.)
- **Goal IDs carry the `_core` prefix (`CORE-G-001`).** A goal reference in tag
  data (`xp_roll`'s `goal` filter) is checked at load. `ActiveGoal` holds the goal
  ID, its progress, its target faction and the assets stealthed at the start; a
  held goal ID the rulebook does not know is skipped and replaced at the next
  turn-start pick. Saves moved by a sandbox re-seed, with no migration code, as
  the tag IDs did. Rejected: migration code that renames IDs in state. (A.001.8
  discovery D-24 and D-26; Effort 1 Decision Record 16.)
- **Per-setting rule values live in `settings.toml`.** A rulebook folder supplies
  `settings.toml`, composed like any other data file; `rulebook.Load` requires it.
  It holds the drift costs (once `drift_costs.toml`) and the least world
  population per asset type (`[purchase_population] Starship = 500000`).
  `drift_costs` is required; any other key left out turns its rule off, and an
  unknown key fails the load. `rulebook.toml` stays compose-only, because its include
  paths mean nothing inside a campaign. Older campaigns re-compose. (A.001.7
  discovery D-10; overview R-6.)
- **The tag and goal loaders are strict, as the asset loader is.** An unknown key
  fails the load and a duplicate ID is refused. All three loaders share
  `refuseUnknownKeys`. The TOML decoder drops unknown keys silently, so a misspelt
  key would otherwise do nothing. The duplicate-ID check is the load-time partner
  of the compose step's file-level refusal. (A.001.4 Discovery D-8.)

## Dependencies

**Depends on** [`domain`](effect-mutation.md) (the `Faction`, `Asset`, `Tag`,
`Goal`, and `AssetDefinition` types these files decode into) and the TOML
decoder. The `state` and `rulebook` packages depend on `campaigns` only
indirectly — they receive resolved paths, not a `Campaign`.

**Depended on by** nearly everything: the [orchestrator](orchestrator.md) holds
the `Rulebook` and drives `state.Save`; the [effect & mutation](effect-mutation.md)
apply layer writes the `FactionState` these files persist; the
[world engine](world-movement.md) reads drift costs and the [spatial](spatial.md)
data dir; and the CLI/[interface](../interface/overview.md) layer creates and
selects campaigns through the registry. The [narrative digest](narrative-digest.md)
reads the `history.jsonl` whose path this package owns.
