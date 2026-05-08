# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped — at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| # | Item | Size | Status | Trigger |
|---|------|------|--------|---------|
| 1 | [Spatial Model and World Graph](#spatial-model-and-world-graph) | Major | Planned | Hex distance, tech level, or Pirates tag becomes blocking |
| 2 | [`Faction.Assets` Slice → Map](#factionassets-slice--map) | Minor | Planned | Asset/faction counts make scan cost measurable, or during Spatial Model work |
| 3 | [Persistence Strategy: Decouple History Writes](#persistence-strategy-decouple-history-writes-from-state-checkpoints) | Minor | Planned | Next pipeline touch (TUI rebuild, stat raise phase, or Spatial Model) |

---
<br/>

## Deferred — Major

Items that will eventually warrant a full initiative entry. Each moves to Planned Initiatives when scoped.

| # | Item | Trigger |
|---|------|---------|
| 1 | TUI Rebuild | Write-up in progress |
| 2 | Programmatic tag handling | Scheduled tag milestone; 18 remaining tags + Effects Engine (`S`-flag assets); design in `tag-engine-discovery.md` |
| 3 | Campaign-scoped static data | No concrete trigger; revisit when distribution story is resolved |
| 4 | Command layer state mutation | Engine has sufficient substance to absorb `engine.CreateFaction` |
| 5 | Maintenance costs per asset | Cost data added to `AssetDefinition` in TOML |
| 6 | Starting Coin | Coin tracking designed and landed |
| 7 | Tag-granted assets | Engine action resolution complete |
| 8 | Tag reminder affordance | Alongside programmatic tag handling |

---
<br/>

## Deferred — Minor

Small, targeted fixes. No write-up needed — tracked here until scheduled.

| # | Item | Detail |
|---|------|--------|
| 1 | `BuyAsset` stealth ID hardcode (`"C3-002"`) | Add `stealth_applicator` flag or typed `TypeStealth` constant; silently breaks if the TOML ID changes |
| 2 | Goal/tag dispatch hardcodes display IDs | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only; one-time migration |
| 3 | `SeizePlanet` invisible in history | `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| 4 | `narrateUseAssetAbility` fallback text | Moot until TUI rebuild; resurfaces when narration is re-implemented against the new observer interface |

---
<br />
<br />

# Planned Initiatives
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

## Spatial Model and World Graph

### Problem

Worlds are bare strings (`asset.Location = "Tartarus"`). There is no registry of what worlds exist, where they are, or what properties they have. All spatial queries are derived at runtime by scanning faction state — finding assets on a world requires iterating every faction then every asset within each faction, filtering by `.Location`. This O(factions × assets) pattern appears in `attack.go` (`eligibleDefenders`, `liveDefenders`), `expand_influence.go` (`rivalsOnWorld`), `goal/progress.go` (`worldHasRivalPresence`, `rivalHasPlanetaryGovernmentOnWorld`), and `seize_planet.go`. At current scale it is immeasurable; as faction and asset counts grow it compounds.

The deeper cost is what the missing model prevents from being enforced:

- **Hex distance** — `Change Homeworld` sets `TurnsRemaining` to a fixed value that ignores distance; `UseAssetAbility` movement defers hex-distance enforcement to GM adjudication; the `MaxHex` field is parsed and stored but never checked
- **Tech level** — `BuyAsset.purchasableDefinitions` does not filter by `AssetDefinition.TechLevel`; there is nowhere to look up a world's tech level
- **Government permission (flag `P`)** — `BuyAsset.Validate` does not check the `P` flag; enforcing it requires knowing which faction holds the `Planetary Government` tag on the target world
- **Pirates tag** — the movement cost mechanic requires knowing which world an asset moved from and to; without a world graph this is unimplementable

### Approach

Introduce a `World` domain type with at minimum `{ ID string, TechLevel int, Population int }` and a hex-coordinate or adjacency structure for distance queries. Layer a derived spatial index on top — a `map[worldID][]*Asset` and `map[worldID][]*Base` — rebuilt at the start of each turn and treated as read-only during resolution. This index replaces the O(n) world scans throughout the engine.

Campaign state gains a `Worlds map[string]*World` field in `FactionState`. The `faction create` wizard gains a world-setup step. The `FACTION_DATA_DIR` data files gain a per-campaign `worlds.toml`.

### Unlocks

- Hex-distance enforcement for `Change Homeworld` goal timer and asset movement
- Tech level validation in `BuyAsset` and `ExpandInfluence`
- Government permission (`P` flag) enforcement in `BuyAsset`
- Pirates tag mechanic (movement cost)
- Full world-list for `SelectMoveDestination` (currently derived from live asset locations)

### Trigger

Pick up when the first of these becomes blocking: tech level enforcement is needed for a campaign, movement distance needs to be enforced (AI decision-making), or the Pirates tag is scheduled for implementation.

---

## `Faction.Assets` Slice → Map

### Problem

`Faction.Assets` is `[]*Asset`. Every lookup by asset ID requires an O(n) linear scan. This pattern appears 7–8 times in `mutation/mutation.go` alone — one per mutation type that targets an asset by ID (`AssetHPDelta`, `AssetMaintainedFlag`, `AssetStealthCleared`, `AssetMoved`, `AssetStealthApplied`, and others). It also appears in `goal/progress.go` (`findAsset`), `action/actions/attack.go` (`ownerFaction`, `liveDefenders`), and several action files. A mutation batch from a single action can trigger up to 10 of these scans before state is committed.

### Approach

Change `Faction.Assets` from `[]*Asset` to `map[string]*Asset` keyed by instance ID. All ID-based lookups in `MutationEngine.Apply`, the goal engine, and action resolution become O(1). Anywhere the codebase iterates assets and cares about order — bookkeeping maintenance, available worlds for `BuyAsset`, attack candidate lists — needs explicit sorting after the change.

### Tradeoffs

The TOML schema changes. The state file currently encodes assets as an array under each faction (`[[factions.alpha.assets]]`); a map produces `[factions.alpha.assets."alpha-F1-001-1"]` entries. This is a breaking change to existing campaign files and requires a one-time migration for any live campaigns. The precedent exists — `FactionState.Factions` was converted from slice to map (Decision 64) under the same pre-1.0 rationale.

### Trigger

Pick up when: faction or asset counts grow to where the scan cost is measurable, or when the Spatial Model work is underway and the codebase is already being restructured for a world index alongside the asset store.

---

## Persistence Strategy: Decouple History Writes from State Checkpoints

### Problem

`applyAndRecord` — the canonical write path — always does three things together: apply mutations to in-memory state, append a JSON line to the history file, and rewrite the full TOML state file. It is called up to four times per faction turn (goal lock phase, bookkeeping, stat raise, action resolution). The result is 3–4 full TOML rewrites per faction turn.

The history write and the state write serve different purposes and warrant different cadences. The history file is an audit log — every discrete state change should be recorded. The state file is a resume checkpoint — its purpose is "if the process dies, where do we restart from?" Writing it after the goal-lock phase (which may only tick a `TurnsRemaining` counter by 1) adds no resume value over writing it once after bookkeeping completes, because `TurnState.Phase` already encodes which phase the turn is in.

This also blurs the conceptual boundary: a reader of `applyAndRecord` cannot tell whether the state save after goal-lock mutations is meaningful or incidental — it looks the same as the save after action resolution, which genuinely matters for resume correctness.

### Approach

Separate the two write concerns. `applyAndRecord` always writes history (unchanged). `state.Save` is called explicitly at phase gates only — after bookkeeping and after action resolution — co-located with the `AwaitCheckpoint` calls that already mark those boundaries in `orchestrator.go`. The resume logic is identical; the write frequency halves; the intent of each write is explicit.

Update the architecture documentation to state plainly that the history file and the state file have different write cadences and different purposes, so future contributors don't re-merge them.

### Trigger

Pick up when: the turn pipeline is being touched for another reason (TUI rebuild, stat raise phase, or the Spatial Model work), making it low-cost to adjust the save sites at the same time.
