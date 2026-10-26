# Data-Driven Tags and Goals — Exploration

> **Initiative:** `E.002`, the pilot run of the Exploration (`E.`) type. `D.009` writes
> the type's flow and template from what worked in this doc.
> **Feeds:** the paused [`A.001.4` discovery](../discovery/A.001.4-goals-tags-core-discovery.md),
> which waits on the answer to "core, per-rulebook, or hybrid?".
> **Status:** in progress. Session 1 (audit) wrote Current State and Existing
> Patterns. Session 2 fills In Plain Words, Options, Recommendation and Follow-on Rows.

## In Plain Words

_Written after the audit._

<br/>

## Questions

The aim is one consistent way to handle data files, and a project-wide move away from
bespoke handlers. Cost is not a deciding factor.

1. **Tags.** Can tags reuse `A.001.3`'s `[[effects]]` model at faction scope?
2. **Goals.** What goal vocabulary is needed to state the twelve goals as data?
3. **IDs.** How do we stop bare-ID collisions across rulebooks?

<br/>

## Current State

### Tag handlers

`TagEngine` (`engine/tag/tag.go`) holds a `map[string]Handler` keyed on the bare tag
ID. `New` registers four handlers. At cycle start, `ApplyAll` walks each faction's
`Tags` and calls the handler for each ID it knows. An unknown ID is skipped with no
error. Each handler registers one or two hooks at `hooks.FactionScope`.

| ID | Tag | Hook category | Behavior |
|---|---|---|---|
| `T-005` | Fanatical | `RollResultHook` + `TieResolver` | Reroll every die that shows 1, not elective; the faction loses every attack tie |
| `T-013` | Preceptor Archive | `AssetCostModifier` | −1 Coin on the faction's own TL4+ purchases, floor 0 |
| `T-016` | Scavengers | `MutationReactor` | +1 Coin per `AssetRemoved` with cause `attack` or counterstrike, any owner |
| `T-020` | Warlike | `RollModifier` | Force attack: +1d10, keep highest; `BudgetKey` `tag:Warlike`; `FromTag: true` |

- **`T-011` Planetary Government** has no handler. `domain.TagPlanetaryGovernment`
  couples it to the world government record. Production readers:
  `planetary_seizure.go` (`G-004` grants and removes the tag) and the Manage wizard
  (`wizard.go:296`).
- **The other 15 tags are inert.** Their `effect` text is the only statement of the
  rule; the GM applies it. The effect-mutation page records why: the event-hooks
  effort shipped one proof-of-life handler per hook category, not a tag catalog.
- **Tag source is visible to other rules.** Warlike's offer sets
  `ModifierOffer.FromTag`, and the `remove_tag_dice` effect (Blackmail) strips offers
  that have it. A data-driven tag must keep this marker.
- **Tag hooks assume the roll is their own.** The roll dispatcher reads the roller's
  faction bucket and never the opponent's (hooks page, Key Decisions). Fanatical
  rerolls every 1 and does not check the actor.

### Goal handlers

`GoalEngine` (`engine/goal/goal.go`) holds a `map[string]Handler` keyed on the bare
goal ID. `New` registers twelve handlers. Each handler has two methods:
`CheckLock` (goal-lock phase, before the action) and `UpdateProgress` (after the
action, before `MutationEngine.Apply`). An unknown ID returns `LockNone` and no
progress.

| ID | Goal | Path | Counts | Threshold | XP |
|---|---|---|---|---|---|
| `G-001` | Military Conquest | progress | rival Force assets destroyed by own attack | own Force | progress / 2 |
| `G-002` | Commercial Expansion | progress | rival Wealth assets destroyed by own attack | own Wealth | progress / 2 |
| `G-003` | Intelligence Coup | progress | rival Cunning assets destroyed by own attack | own Cunning | progress / 2 |
| `G-004` | Planetary Seizure | lock + progress | phase 1: rival presence on the target world; phase 2: occupation turns | phase 2 countdown of 3 | half the previous ruler's average rating, min 1 |
| `G-005` | Expand Influence | progress | own `BaseAdded` on a world with no own base | 1 event | 1, or 2 if a rival is present |
| `G-006` | Blood the Enemy | progress | HP damage to rival assets and bases | Force + Cunning + Wealth | 2 |
| `G-007` | Peaceable Kingdom | progress | turns with no own attack; an attack resets to 0 | 4 | 1 |
| `G-008` | Destroy the Foe | progress | target faction HP after the batch | target HP ≤ 0 | 1 + the target's average rating |
| `G-009` | Inside Enemy Territory | progress | own `AssetStealthApplied` on a rival's ground | own Cunning | 2 |
| `G-010` | Invincible Valor | progress | own kill of a rival Force asset with `MinRating` above own Force | 1 event | 2 |
| `G-011` | Wealth of Worlds | progress | Coin spent on positive `InfluenceDelta` | 4 × own Wealth | 2 |
| `G-012` | Change Homeworld | lock | transit turns (`LockSkip` each turn) | countdown reaches 0 | 0 |

Shapes that repeat:

- **Counter goals** (`G-001`–`G-003`, `G-006`, `G-009`, `G-011`): count matching
  mutations, add to `Progress`, complete at a threshold read from the faction's
  ratings. Only the event filter, the threshold formula and the XP differ.
- **One-event goals** (`G-005`, `G-010`): complete on the first matching mutation.
- **Streak goal** (`G-007`): progress on a quiet turn, reset on a matching event.
- **State goal** (`G-008`): reads target state after the batch, not a count.
- **Process goals** (`G-004`, `G-012`): multi-turn phases, locks, a countdown, and
  completion side effects (government change, `T-011`, homeworld change). An action
  starts each one (`Seize Planet`, `Change Homeworld`) with start data: a target
  world, a phase, a turn count.

### Tag and goal data

- **Loader.** `rulebook.Load` reads one `tags.toml` and one `goals.toml` from the
  rulebook folder (`rulebook.go:38`, `:43`). Records are keyed on `id`. The loaders do
  not check for unknown keys or duplicate IDs: a second record with the same `id`
  silently replaces the first.
- **Records carry text only.** Tag: `id`, `name`, `description`, `effect`. Goal: `id`,
  `name`, `description`, `difficulty`. `difficulty` is loaded and never read; XP is
  in Go (`B.005`).
- **State embeds full tag copies.** `TagAdded` writes the whole `domain.Tag` into
  the faction, so `state.toml` holds the name and text. A rulebook change does not
  reach tags that factions already hold (`A.001.4` audit).
- **How a goal starts.** Three writers set `ActiveGoal`: the Manage wizard at faction
  creation (`domain/faction.go:247`, goal ID only), the `GoalInitiated` mutation from
  `Seize Planet` and `Change Homeworld`, and nothing else. `GoalCompleted` and
  `GoalAbandoned` set it to nil. No turn step picks a new goal after one ends.

### Gaps found in the audit

Found by reading the code, not by a run. Filed in the session 1 audit: `B.039` (goal
selection), `B.040` (`G-008` target), `B.041` (wizard goal list).

- **`G-008` cannot complete in play.** Only `destroy_the_foe.go` reads
  `ActiveGoal.TargetFactionID`, and no production code writes it (two scenario tests
  set it by hand). With no target, `UpdateProgress` returns nil every turn.
- **The wizard offers goals that need start data.** `goalsToOptions` lists every
  goal. `G-004` from the wizard starts at phase 0 with no target world, so it never
  advances. `G-012` from the wizard starts with `TurnsRemaining` 0; `CheckLock`
  counts to −1 and never reaches 0, so the faction gets `LockSkip` every turn and
  cannot take `Abandon Goal`.
- **No new goal after completion (a regression).** After `GoalCompleted`, a faction
  has no goal unless it takes `Seize Planet` or `Change Homeworld`. The rules say it
  may select a new goal at the start of its next turn. The old TUI did this:
  `cmd/faction-manager/tui/phases/goal_select.go` (`0561d28`, 2026-04-29) prompted
  when `ActiveGoal == nil`, with start inputs for `G-012` and `G-004`. Commit
  `798cd45` (2026-05-03) deleted the old TUI before the orchestrator refactor, and
  the TUI rebuild did not bring the step back. That design also gave every other goal
  zero-value target fields, so the `G-008` gap is older than the rebuild.

### Related rows

| Row | Overlap |
|---|---|
| `R.004` | Data-driven handlers: TOML-defined goal and ability dispatch with a typed primitive registry. The ability half shipped in `A.001`; the goal half is this exploration's question 2 |
| `B.001` | `"G-012"` hard-coded in the `Change Homeworld` action |
| `B.002` | Key `Rulebook.Goals` / `Rulebook.Tags` on the table key; `id` becomes display-only. Bears on question 3 |
| `B.005` | Retype goal `difficulty` so the engine can read XP from data |
| `B.013` | Peaceable Kingdom misses a missed attack; its event filter is wrong |
| `F.013` | Goal state predicates (`Satisfied(state) bool`) for AI; a data-driven goal shape must allow it |

<br/>

## Existing Patterns

### Asset model

Assets went from per-ID handlers to data in `A.001.1`–`A.001.6`. Each piece below is
a candidate for tags and goals to reuse.

- **Prefixed IDs.** Asset IDs carry the rulebook (`SWN-C1-003`, from `A.001.1`). A
  second rulebook cannot reuse an SWN asset ID by accident.
- **Strict loader.** `loadAssets` globs `assets/*_assets.toml`, fails on unknown keys
  (a misspelt filter would do nothing), and refuses a duplicate ID. The tag and goal
  loaders do neither.
- **Effects list on the definition.** An asset declares `[[assets.<Key>.effects]]`
  entries: `trigger`, `effect`, `filters` (owner, scope, optional filters), `gates`,
  and a params table named for the effect. Example: False Front,
  `trigger = "lethal_damage"`, `effect = "sacrifice_self"`, `owner = "own"`,
  `scope = "world"`.
- **Registrar per effect.** `effect.Registrar` (`engine/effect/registrar.go`) states
  which triggers, owners, scopes, filters and gates an effect accepts, and its route:
  `Register` (hook at cycle start), `Fire` (phase event), or `Matchup` (inside the
  Attack action). Engine construction validates every entry against its registrar and
  fails loud on a mismatch.
- **Dispatch keyed on the definition, not a handler per ID.** One
  `effectsListHandler` runs any definition's whole effects list.

**What the asset model assumes.** Every `Register` route receives
`RegisterContext{Faction, Asset, Def, …}` and builds its hook around the declaring
asset:

- It registers at `hooks.AssetScope(faction, asset)`, not `FactionScope`.
- At fire time it checks that the asset is still live and active
  (`liveAsset`, `active`).
- `owner` and `scope` are read from the asset: `scope = "world"` means the asset's
  world, and `scope = "faction"` (used by `ignore_coin_loss`, `grant_movement`,
  `tech_level_floor`) means "an event anywhere in the owner's faction". The hook is
  still asset-scoped and still needs the asset to be active.

A tag has no declaring asset. So `scope = "faction"` today is not the same as
"declared by the faction", which a tag needs. Question 1 turns on this gap.

**Nearest effect for each tag handler.**

| Tag | Nearest effect | Fit |
|---|---|---|
| Warlike | `bonus_die` (`own_roll`, filters `attribute`, `phase`) | Same dice shape (+1d10, keep highest). Missing: no asset; `FromTag` marker; once-per-turn budget |
| Fanatical | `reroll` (`own_roll`) | `reroll` picks the lowest *n* dice; Fanatical rerolls every 1. No effect resolves ties |
| Scavengers | none | No effect reacts to `AssetRemoved` with Coin; `coin_gain` fires at `turn_start` only |
| Preceptor Archive | `cost_surcharge` (`asset_cost_query`) | Opposite sign and owner: surcharge is +Coin on rival buyers on the asset's world; Archive is −Coin on own TL4+ buys anywhere |

**Goals have no model to reuse.** No asset effect tracks progress over turns, sets a
lock, or completes with XP. The goal-side vocabulary (event filters, thresholds read
from ratings, XP formulas, process phases) would be new. The closest existing pieces
are the effect filters (owner, attribute, category) for the event side, and the
`ProcessPhase` / `TurnsRemaining` fields already in `ActiveGoal`.

<br/>

## Options

### Tags

_Session 2._

### Goals

_Session 2._

### IDs

_Session 2._

<br/>

## Recommendation

_Session 2._

<br/>

## Follow-on Rows

_Session 2._
