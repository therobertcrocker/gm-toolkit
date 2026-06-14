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
tree + a `campaign.toml` manifest) followed by `CopyRulebook`, which walks a
source rulebook directory and copies every `.toml` into the campaign's
`rulebook/`, refusing to clobber existing files unless `replace` is set.

### Mutable state: `state.toml`

`FactionState` is the entire mutable game: the `CampaignID`, the `CycleNumber`,
a `map[string]*Faction`, and the optional `CurrentTurn` cursor (the pause/resume
state from the [turn pipeline](turn-pipeline.md)). It is one TOML file.

- **`Load` is bootstrap-friendly.** A missing `state.toml` returns an *empty*
  state, not an error, so a freshly scaffolded campaign runs without a manual
  init step. Load also normalizes nil maps (`Factions`, each faction's `Assets`)
  so callers never have to nil-check.
- **`Save` seeds on first write.** After encoding the state, `ensureSeed` writes
  `state.seed.toml` beside it *if one doesn't already exist* — a one-time
  snapshot of the starting state that `reset-campaign.sh` restores. Every
  subsequent save leaves the seed untouched.

### Atomic faction CRUD

`CreateFaction` and `DeleteFaction` apply their change to the in-memory map,
then `Save`. If the save fails they **roll the in-memory change back** — the
created entry is deleted, the deleted entry restored — so `fs.Factions` always
matches what is on disk and the caller can safely retry. Both guard their
preconditions with sentinel errors (`ErrFactionAlreadyExists`,
`ErrFactionNotFound`, `ErrInvalidFactionID`).

### Static rules: the rulebook

`rulebook.Load` reads the GM's static data from the campaign's rulebook dir into
one in-memory `Rulebook` — the authoritative rules source the engine queries all
turn:

- **Assets** are **globbed**: every `assets/*_assets.toml` is decoded and merged
  into one `map[string]*AssetDefinition`, with a hard error on a duplicate asset
  ID across files. This lets a GM split assets across themed files
  (`core_assets.toml`, `homebrew_assets.toml`) that merge transparently.
- **Tags, goals, drift costs** load from single files (`tags.toml`, `goals.toml`,
  `drift_costs.toml`).
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
- **Load degrades gracefully; mutations are atomic.** A missing state file is an
  empty campaign, not a crash, and faction CRUD rolls back its in-memory change
  if the save fails — disk and memory never diverge.
- **First save captures a seed.** `ensureSeed` snapshots the starting state once,
  enabling a clean reset without re-scaffolding.
- **Asset files glob and merge; rules validate at load.** Multiple
  `*_assets.toml` merge into one rulebook (duplicate IDs rejected), and every
  enum-like field is validated during conversion, so malformed rules fail at
  startup rather than during a turn.

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
