# Data-Driven Tags and Goals — Exploration

> **Initiative:** `E.002`, the pilot run of the Exploration (`E.`) type. `D.009` writes
> the type's flow and template from what worked in this doc.
> **Feeds:** the paused [`A.001.4` discovery](../_completed/arcs/A.001-data-driven-rulebook/initiatives/A.001.4-goals-tags-core/A.001.4-goals-tags-core-discovery.md),
> which waits on the answer to "core, per-rulebook, or hybrid?".
> **Status:** complete (reviewed by Robert). Session 1 (audit) wrote Current State
> and Existing Patterns. Session 2 wrote In Plain Words, Options, Recommendation and
> Follow-on Rows.

## In Plain Words

- **What:** tags and goals get their behavior from Go code written for each bare ID.
  This exploration asks how they can get it from rulebook data, as assets do now.
- **Short answer:** yes for both. Tags reuse the asset `[[effects]]` model: each
  registrar learns a second *holder*, the faction. Goals need a new vocabulary of
  named parts (start, progress, threshold, completion, XP, phases). The asset model
  gives goals the pattern, not the parts. When no Go code keys on a tag or goal ID, an
  ID collision is a load error, and a rulebook prefix makes it rare.
- **Why it matters:** a second rulebook (H.A.W.K) gets behavior from its own data, not
  by reusing an SWN ID. The goal vocabulary also closes the audit gaps by its shape:
  a goal states the start data it needs (`B.039`–`B.041`), its XP (`B.005`) and the
  action that resets it (`B.013`).
- **How it lands:** `A.001.4` unpauses with a scope answer (goals in `_core`, tags
  per rulebook), and two new initiatives do the conversion, tags first.

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

Two words, kept apart in every option below (Robert's call, session 2):

- **holder** — the thing that declares an `[[effects]]` entry: an asset (through its
  definition) or a faction (through a tag it holds). New word; no Go identifier uses
  it.
- **scope** — unchanged: which events or rolls an entry matches, read relative to the
  holder. Today `scope = "faction"` on an asset entry means "an event anywhere in the
  owner's faction". It does not mean "the faction is the holder".

### Tags

#### T-1 — Shared effects, faction as a second holder

A tag record carries `[[tags.<Key>.effects]]` with the same `EffectDefinition`, the
same loader conversion, and the same registrars.

- **Registrar.** `Accepts` gains `Holders` (`asset`, `faction`), per trigger like the
  other sets. `validateEntry` refuses a tag entry for an effect that accepts only
  `asset`, at engine construction, as for assets.
- **Register route.** `RegisterContext` names its holder. For a faction holder,
  `Asset` and `Def` are nil and a tag field is set. The hook registers at
  `hooks.FactionScope`.
- **Fire-time check.** "The faction still holds the tag" replaces `active(asset)`.
  The reason is the same as for assets: tags change in the middle of a cycle (`G-004`
  removes `T-011`; Deep Rooted is lost on a homeworld change).
- **Scope per holder.** Each registrar documents what each scope means for a faction
  holder. `faction` reads as "any event of the holder's faction"; `self` has no
  asset to name. Deep Rooted and Colonists need a new value, `homeworld`.
- **`FromTag` comes from the holder.** The registrar sets the marker when the holder
  is a faction. No data flag, so Blackmail's `remove_tag_dice` keeps working.
- **Budget key.** `budgetKey` uses the definition ID; a tag entry uses the tag ID.
  Warlike's key changes from `tag:Warlike`, so a mid-cycle save from before allows one
  extra use (the `B.012` pattern; it self-heals).
- **Roll dispatch.** Data entries check the owner relation themselves
  (`rollMatches`), so the hooks Key Decision "never the opponent's faction bucket"
  loses its reason (Fanatical's missing actor check). It needs to change only for a
  tag that acts on a rival's roll (`T-014`, "whether or not they are involved").

Fit of the four tags with Go handlers:

| Tag | Data form | New vocabulary |
|---|---|---|
| Warlike | `bonus_die`, `own_roll`, `attribute = "force"`, `phase = "attack"` | `bonus_die` accepts no gates today; add `once_per` |
| Fanatical | `reroll`, `own_roll`, not elective | A die selector for "every die showing 1" (`reroll` picks the lowest *n*); a new `lose_ties` effect (Cat 5) |
| Scavengers | `coin_gain` on a new `asset_destroyed` trigger | The trigger, and an owner filter: own kills and own losses only (see `B.042` below) |
| Preceptor Archive | A new `cost_discount` (`asset_cost_query`, own buyer, floor 0) | The effect and a `min_tech_level` filter; a negative `cost_surcharge` would invert its owner and scope readings |

Fit of the 15 inert tags, by what they need:

- **Existing effects plus new filters or scopes:** Imperialists, Machiavellian,
  Plutocratic, Theocratic, Savage, Perimeter Agency, Deep Rooted (`bonus_die` with
  attribute, phase, action, tech-level or `homeworld` filters). Mercenary Group
  (`grant_movement`). Secretive (`apply_stealth` at `on_purchase`; `FirePurchase`
  must also read the buyer's tags).
- **New effects or triggers:** Pirates (a movement surcharge paid to the holder),
  Exchange Consulate (a `goal_completed` trigger with an XP roll), Deep Rooted's loss
  on a homeworld change (a tag that removes itself).
- **Outside the S-side model:** Eugenics Cult (grants a purchasable asset),
  Preceptor Archive's special action, Psychic Academy's mentor training, the
  population limits of Colonists and Technical Expertise. These are the rules calls
  that `A.001.4` already routes to Robert. A tag that grants an action would need an
  A-side harness.
- **`T-011` Planetary Government:** no effects. The government record finds it
  through `domain.TagPlanetaryGovernment`. A data marker (`government = true`) can
  replace the constant (see IDs).

#### T-2 — A separate tag vocabulary

A registrar table in `engine/tag`, with the same TOML shape and its own effect names.
The asset registrars do not change. One rule gets two names (`bonus_die` for assets,
another for tags), which breaks ACD-4 (one primitive, many harnesses) and the aim of
one way to handle data files.

#### T-3 — Tags as hidden assets

Each tag becomes an asset definition that the faction holds as an invisible instance,
so the asset model runs with no change. Every asset walk (attack targets,
maintenance, digest, asset-count goals, faction HP) must then skip it. The cost is
spread over every present and future reader.

#### T-4 — Keep the Go handlers (baseline)

Fix only the IDs. A rulebook still gets a behavior only by reusing an ID, and 15 tags
stay inert.

**For T-1 to T-3: factions hold tag IDs, not copies.** `TagAdded` writes the whole
`domain.Tag` into state. With effects on the record, each save would freeze the
rules that applied when the tag was granted. A faction holds the tag ID and the
engine reads the definition from the rulebook, as `Asset.DefinitionID` does. The two
sandbox campaigns need a migration.

### Goals

#### The vocabulary

Every goal in the table above can be stated with these parts. Each part name is a Go
primitive with parameters, checked at engine construction, as an effect name is.

| Part | Job | Used by |
|---|---|---|
| `start` | How the goal begins: `select` (the turn-start pick, `B.039`) or a named action (`seize_planet`, `change_homeworld`). The inputs it needs: `target_faction`, `target_world` with a candidate rule, a turn count from the action. A selection gate | All; inputs for `G-004`, `G-008`, `G-012`; gate for `G-011` ("Wealth must increase before this goal can be selected again", rules doc and `goals.toml`) |
| `progress` | An event source, filters and an amount. Events: `asset_destroyed`, `damage_dealt`, `stealth_applied`, `base_added`, `influence_bought`, `action_taken`. Filters: owner relation, category, world relation (rival-governed), rating compare. Amount: 1, damage, Coin | Counter and one-event goals |
| `tick` | Progress for each turn at the goal-lock beat | `G-007` |
| `reset` | Events that set progress to 0 | `G-007` (`action_taken = "attack"` fixes `B.013` by its shape) |
| `threshold` | A formula on own ratings: one rating, a sum, a multiplier, a constant | Counter goals; `G-007` (4) |
| `complete` | Progress reaches the threshold, or a state predicate is true | `G-008` (target HP ≤ 0); the predicate is `F.013`'s `Satisfied(state)` |
| `xp` | A named formula: a constant, half the progress, 1 + target average, half the previous ruler's average (min 1), +1 if contested | All (`B.005`) |
| `phases` | Per phase: a lock (`none`, `skip`, `restrict` plus actions), an advance (predicate or countdown), an abandon predicate, and named completion effects (`take_government`, `move_homeworld`) | `G-004`, `G-012` |

- **The two beats stay.** `tick`, countdowns and locks run in `CheckLock`; `progress`,
  `reset` and event advances run in `UpdateProgress`, before `MutationEngine.Apply`
  (goals Key Decisions).
- **`action_taken` needs a new input.** `UpdateProgress` receives the mutations and
  no action name (`orchestrator.go:545`). A missed attack by an unstealthed attacker
  emits no mutation, which is `B.013`.
- **XP checked against the rules.** The XP column above matches the
  `swn-faction-mechanics.md` goal table. A data formula keeps today's reading. Any
  change is a rules call (see Recommendation).

#### G-1 — Composed parts

Each goal composes the parts above. Counter, one-event, streak, state and process
goals use the same parts, and a process goal is a `phases` list. Example, `G-001`:

```toml
[goals.MilitaryConquest]
id = "SWN-G-001"
start = { kind = "select" }
progress = { event = "asset_destroyed", by = "own", owner = "rival", categories = ["force"], amount = "count" }
threshold = { rating = "force" }
xp = { formula = "half_progress" }
```

#### G-2 — One shape field

`shape = "counter" | "one_event" | "streak" | "state" | "process"`, each a Go kernel
with parameters. Fewer names. But one `process` parameter set cannot state both
`G-004` (two phases, a government change) and `G-012` (a transit countdown, a
homeworld change), so `process` splits into two kernels. Those are the two per-goal
handlers under a new name.

#### G-3 — Data for the simple shapes, Go for process goals

The counter, one-event, streak and state goals use G-1's parts. `G-004` and `G-012`
keep Go, keyed on a behavior name (`process = "seize_world"`), not on an ID. The
precedent is the `[transport]` profile. Smallest change. It leaves two kernels that
only these two goals use.

### IDs

In every option, **no Go code names a tag or goal ID**:

- **Roles, not IDs.** The government record finds its tag by a data marker
  (`government = true`). `Seize Planet` and `Change Homeworld` find their goal by its
  `start` action. This closes `B.001` and removes `domain.TagPlanetaryGovernment`.
- **Data-to-data references go by ID and are checked at load.** Exchange Consulate
  names Peaceable Kingdom. An unknown reference fails the load.
- **Strict loader.** The tag and goal loaders fail on unknown keys and refuse a
  duplicate ID, as `loadAssets` does.

#### I-1 — Rulebook prefix

`SWN-T-020`, and a prefix for each home folder (`CORE-` for `_core`). This is the
`A.001.1` asset convention. Saves need a migration (sandbox state holds `T-011` and
`G-004`).

#### I-2 — Table key as the identity (`B.002`)

`[tags.Warlike]` is the key, and `id` becomes display-only. Assets keep a prefixed
`id`, so the two conventions split. Table keys collide across rulebooks the same way
bare IDs do.

#### I-3 — Refusal only

Keep bare IDs. With behavior in data, a collision in a composed set is a load error,
not a silent handler. Smallest change. Tag and goal IDs then read unlike asset IDs.

<br/>

## Recommendation

- **Tags: T-1.** ACD-2 asks for behavior from data on every surface, and ACD-4 asks
  for one primitive with many harnesses. A faction holder is one more harness for the
  same registrars. Most inert tags need only new filters on effects that exist.
  Factions hold tag IDs, not copies.
- **Goals: G-1.** One way to state every goal. G-2 and G-3 keep two kernels that only
  `G-004` and `G-012` use, which are per-ID handlers under a behavior name. The
  `phases` part serves two goals today; cost is not a criterion, and a second
  rulebook can add process goals with no Go.
- **IDs: I-1, with roles and the strict loader.** It matches the asset convention.
  `B.002` is superseded.

### How it lands

- **`A.001.4` unpauses with a scope answer: goals in `_core`, tags per rulebook.**
  Goals are the objective structure of the faction game and are near-neutral. Tags
  are setting identity. Once behavior is data, an H.A.W.K tag states Warlike's
  behavior in its own entry and shares no record with SWN. This is narrower than the
  hybrid in `A.001.4`'s Open Questions: no tag goes to core. So the `T-003`, `T-013`
  and `T-014` neutralization rules calls leave scope, and a merge-by-record loader is
  not needed (each folder supplies a whole file). `T-011` need not be core, because
  the government role is a data marker. This amends ACD-6 (core holds goals only)
  and narrows ACD-7 (neutralization covers goals only), so the arc-plan needs an
  amendment. The strict loader and the I-1 prefixes also go into `A.001.4`, because
  it moves and edits the same files.
- **`A.001.7` — data-driven tags (T-1).** Holder on the registrars, faction holds IDs,
  the four handlers become data, then the inert tags that fit. It runs before
  H.A.W.K tag authoring.
- **`A.001.8` — data-driven goals (G-1).** The vocabulary, the turn-start goal pick as
  an engine step with a collector ask, and the twelve goals as data. It absorbs
  `B.001`, `B.005`, `B.013`, `B.039`–`B.041` and the goal half of `R.004`. The
  `complete` predicate keeps `F.013` open.
- **Order:** `A.001.4`, then `A.001.7` and `A.001.8` in either order. Both touch the
  rulebook loader; tags first is smaller and proves the holder before goals reuse
  the effect filters.

### Rules calls to carry (Robert)

- **Scavengers** — decided in session 2: own kills and own losses only (`B.042`).
- **`G-004` XP when no faction resists** — the `goals.toml` text says "counts as
  difficulty low"; Go gives 1 only when no faction rules the world.
- **`G-009` count** — the rules say "have Stealthed assets … equal to Cunning";
  Go counts stealth events, so an asset that loses stealth still counts.
- **`G-011` gate** — what "Wealth must increase" compares against.
- **`T-014` reach** — "whether or not they are involved" means any roll anywhere.

<br/>

## Follow-on Rows

Filed in `planned-work.md` (session 2): `A.001.7` and `A.001.8` in Up Next, `B.042`
under Bugfixes (Engine), the `A.001.4` row amended, `B.002` removed.

| ID | Item | Type | Detail |
|---|---|---|---|
| `A.001.7` | Data-driven tags | `feature` | Tags carry `[[effects]]`; registrars accept a faction holder; factions hold tag IDs; the four Go handlers and the inert tags that fit become data (`E.002` T-1) |
| `A.001.8` | Data-driven goals | `feature` | Goal vocabulary of named parts, a turn-start goal pick in the engine, the twelve goals as data; absorbs `B.001`, `B.005`, `B.013`, `B.039`–`B.041`, `R.004` goal half (`E.002` G-1) |
| `B.042` | Scavengers pays for other factions' fights | `bugfix` | `ScavengersReactor` pays 1 Coin on any combat `AssetRemoved`, and `allReactors` gives every faction every batch, so a third party's fight pays; the rules pay on own kills and own losses only (Robert's call, `E.002` session 2). From reading the code, not a run |
| `A.001.4` | Scope answer (amend the row) | -- | Goals in `_core`, tags per rulebook; add the strict tag/goal loader and the I-1 prefixes; arc-plan amendment for ACD-6 and ACD-7 |
| `B.002` | Close as superseded | -- | I-1 replaces the table-key approach |
