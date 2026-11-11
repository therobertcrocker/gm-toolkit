# S-flag Effect Registry (A.001.3) — Effort 1 Plan: Foundation

## Context / Goal

Build the S-side model that Efforts 2–5 stand on, and prove it end to end with one
effect. After this effort an asset declares an `[[effects]]` list in its TOML; the
loader parses it strictly; the effects engine checks every entry against a registrar
table at construction; and a turn-start effect fires from a direct definition read in
the bookkeeping phase. Party Machine `SWN-C4-001` (`coin_gain`) is the proof.

- Overview plan: [`s-flag-effect-registry-plan.md`](./s-flag-effect-registry-plan.md)
  (Effort 1 row in the Effort Summary; scoping outline in Work Breakdown; Planning
  Decisions 5, 6, 7, 11, 13 and the Shared Context domain-shape contract are this
  effort's mandate).
- Discovery: [`s-flag-effect-registry-discovery.md`](../../discovery/completed/s-flag-effect-registry-discovery.md)
  (D-1…D-6 bind here).
- Architecture: [effect & mutation](../../../architecture/engine/effect-mutation.md) (the
  effects engine and its Key Decisions), [hooks](../../../architecture/engine/hooks.md),
  [ability & effect catalog](../../../architecture/engine/ability-catalog.md).

## Decisions Ratified in Planning

1. **A registrar is a spec struct; the registrar table is injected.** `effect.Registrar`
   carries `Triggers`, a map from each accepted trigger to an `effect.Accepts` that holds
   the `Owners`, `Scopes`, `Filters` and `Gates` accepted under that trigger (Decision
   Record 10), plus two function fields, `Register` (hook route) and `Fire` (phase
   route). A nil function
   means the effect has no such route. The table is a map literal keyed on
   `domain.SpecialEffectType`, returned fresh by `effects.Registrars()`, and handed to
   `effect.New`. The contract lives in `effect`; the implementations live in `effects`,
   which imports `effect` (never the reverse). Resolves the overview's first open
   question. Rejected: an interface with `Accepted*()` methods, which forces a no-op
   `Register` or `Fire` on every single-route effect.

2. **Rule 4 checks names, owner/scope values, routes, and registrar existence.** At
   engine construction every entry must name an effect with a registrar; its trigger
   must be accepted; its owner and scope must be in the sets accepted *for that
   trigger*; every optional filter and gate it sets must be accepted for that trigger;
   and the trigger's route function must be non-nil. The
   owner/scope value check closes a gap in the discovery's rule 4: `coin_gain` with
   `owner = "rival"` passes a names-only check and silently does nothing.

3. **The route is picked by the trigger.** `effect.IsPhaseTrigger` reports `turn_start`
   and `on_purchase`; those entries run through `Fire` from a direct definition read
   (D-5). Every other trigger runs through `Register` at cycle start.

4. **An effect with no registrar is a construction error.** *Refines effect-mutation's
   "Framework first, handlers incremental" for the S side.* An asset with no effects
   list stays inert; an entry naming a known effect the engine cannot run is wrong
   data, not absent data. Consequence for Efforts 2–5: each effect's TOML must land in
   the same commit as its registrar, which the overview already does.

5. **`turn_start` fires at the start of the bookkeeping phase, not in turn setup.**
   *Amends overview Planning Decision 5 (which names the turn setup).* `setupFactionTurn`
   re-runs on every resumed turn, has no idempotency flag, and cannot record mutations.
   `runBookkeepingPhase` fires turn-start effects before `ApplyBookkeeping`, gated on
   `!CurrentTurn.BookkeepingApplied` (the persisted guard), as their own
   `applyAndRecord` batch. Coin arrives before maintenance is paid, matching SWN's
   income-then-upkeep order. The observer's `OnBookkeepingApplied` receives the
   turn-start mutations ahead of the bookkeeping ones, so the GM sees them.

6. **Only active assets fire phase effects.** `FireTurnStart` skips an asset that is not
   `Ready` or not `Maintained` — the same test action eligibility uses
   (`eligibility.go:14`, `use_asset_ability.go:40`). `Maintained` reflects last turn's
   upkeep, because turn start fires before this turn's maintenance.

7. **`FireTurnStart` is a direct definition read on `EffectsEngine`.** It walks the
   acting faction's assets in ID order and calls `Fire` on `turn_start` entries. The
   composed handler serves the hook route only. The orchestrator still owns the call
   site, so effect-mutation's "the rule layer registers hooks, it never dispatches" holds
   in spirit: the effects engine resolves a phase event the orchestrator raises.
   Graduate this refinement at pre-merge.

8. **`EffectsEngine.Register` rejects a duplicate definition ID.** *Generalizes overview
   Planning Decision 6.* Instead of a `Transport`-and-`Effects` check in `core.go`, any
   second handler for one definition is an error. This covers the transport collision
   and any future one.

9. **`NewWithRulebook` grows its error return in the registrar commit.** *Amends overview
   Planning Decision 7's placement.* The duplicate guard needs the error path, and rule-4
   validation is one function beside it, so the overview's commits 3 and 4 merge.
   Effort 1 is **four commits**, not five.

10. **`ApplyAll` walks factions and assets in ID order.** It ranged over two maps, so hook
    registration order varied run to run. Every later effort's dispatch order rests on
    registration order.

11. **The whole asset file is strict.** `loadAssets` checks `MetaData.Undecoded()`; any
    key with no record field fails the load. This is what makes rules 1 and 2 enforceable
    for filter, gate and parameter-block names — BurntSushi drops unknown keys silently.
    It also catches typos in `attack`, `ability` and `transport`. The SWN data was read
    key by key against the records and is clean.

12. **Parameter blocks decode into named pointer fields.** `effectRecord` has one pointer
    per effect, tagged with the effect name, mirroring `abilityStepRecord.Relocate` /
    `Purchase` and the domain's typed pointers. Rule 2 is extended: an effect that takes
    parameters **must** carry its block. Rejected: `toml.Primitive` with deferred decode,
    which threads `MetaData` into `convertAsset` and hides the shape.

13. **The contract and the loader cover all 13 A.001.3 effects, and only A.001.3's
    vocabulary.** Thirteen effect names, twelve triggers (the four successor-only
    triggers — `own_defense_win`, `own_asset_dies`, `permission_query`,
    `ability_eligibility_query` — arrive with A.001.6). Parameter shapes come from the
    overview's authored entries. Later efforts write a registrar and TOML only.

14. **`RelocateParams` grows in Effort 1, and the A-side loader rejects the new fields.**
    *Amends the overview's Effort 5 outline, which gave the growth its own commit.*
    `grant_movement` parses in Commit 2, so `Categories`, `ExcludeTypes` and
    `CoinCostMode` must exist here. The action-spent relocate handler does not read them
    until Effort 5, so an `[ability]` relocate that sets any of them fails the load
    rather than silently doing less than its data says. Effort 5 wires the behaviour and
    lifts the rejection.

15. **Flat amounts parse through a new `parseAmount`.** *Corrects overview Planning
    Decision 11.* `parseDice` rejects `"1"` (its pattern is `NdM±K`). `parseAmount`
    turns a bare integer into `DiceRoll{Modifier: n}` — which `Roll` returns without
    rolling — and sends everything else to `parseDice`. `parseDice` now rejects a die
    with zero sides for every caller, `counter` and `attack.damage` included (Decision
    Record 11).

16. **S-side effect names use `SpecialEffectType` and a `Special*` prefix.** Mirrors the
    A-side's `AbilityEffectType`, and keeps the two enums apart in one package
    (`SpecialClearStealth` beside `EffectRevealStealth`).

17. **The `phase` filter is a typed `EffectRollPhase`.** The overview calls it a plain
    string so `domain` need not import `hooks`; a string-backed type meets the same
    constraint and lets the loader reject an unknown value. The `action` filter stays an
    unchecked string — action names live in the engine; Effort 3 gives it a source.

18. **The catalog's Effort-1 reconciliation items land in the `coin_gain` commit.** Scope:
    renumber the two `Cat 5 RuleModifier` labels to Cat 4 (per `hooks/doc.go`), mark
    Party Machine as built on the turn-start route, and bring the "latent vocabulary"
    table to as-built truth (A.001.2 removed `coin_steal`/`coin_drain` and wired the
    other rows). Mis-*classifications* in the S-flag table (Bank, tech-level uplift,
    attack restriction, free relocate) are left to the efforts that build those effects.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

1. **`EffectFilters.Present()` / `EffectGates.Present()` renamed to `Used()`** (Commit 1).
   "Present" reads as a verb or as "current". `List()` hints at the full vocabulary, not
   the set entries. `Active()` collides with "active asset" (Decision 6: `Ready` and
   `Maintained`), and suggests a runtime match, not declared data. `Used()` says "the
   filters this entry uses". Plan body reconciled: Commit 1 Tasks 1 and 5, Commit 1
   message, Commit 3 Task 1 `validateEntry`. The loader's `presentParamBlocks` (Commit 2)
   keeps its name — it is unexported and names parameter blocks, not filters or gates.

2. **`toEnum` errors list the accepted vocabulary** (Commit 2). The error reads
   `unknown <kind>: "<value>" (want one of [...])`. The `unknown <kind>: %q` prefix keeps
   the loader's error convention and the tests' `Contains` checks hold. The suffix tells a
   GM who is fixing a TOML file the valid values, with no need to read the source. The
   per-enum `switch` helpers in `rulebook.go` keep their shorter form; this is not
   back-ported. Plan body reconciled: Commit 2 Task 7 `toEnum`.

3. **Loud-failure tests cover the numeric parameter bounds** (Commit 2). The planned
   table checked names and blocks but no values, so the `> 0` checks for `bonus_die`,
   `reroll`, `tech_level_floor` and `cost_surcharge` had no test. Four rows were added,
   one per check, each set to 0. A zero `coin_gain` amount (`"0"`) still loads; it was
   left as it is (YAGNI). Plan body reconciled: Commit 2 Task 9.
   *Reversed by Decision Record 11:* a zero amount now fails the load.

4. **Rule-4 validation folds into `RegisterEffectsList`** (Commit 3). The plan had an
   exported `Validate(rulebook)` pass that `NewWithRulebook` had to call before any
   `RegisterEffectsList`. Nothing enforced that order, and the fence comment in
   `effectsListHandler.Apply` relied on it: an unchecked list with an unknown effect
   panics on a nil `Register` at cycle start, not at construction. Now
   `RegisterEffectsList` checks the definition's entries and then registers, so no list
   can be registered unchecked. `Validate` is removed. `NewWithRulebook` walks
   definitions in ID order, so the first error is the same every run. The error text and
   the tests' assertions are unchanged; `TestValidate` is renamed
   `TestRegisterEffectsList_Validates`. Plan body reconciled: Commit 3 Tasks 1, 2, 5 and
   10, Commit 3 message.

5. **Engine-construction errors are wrapped with the caller's prefix only** (Commit 3).
   `core.go` already prefixes `engine:`, so `faction: build engine: %w` read as
   `faction: build engine: engine: asset …`. `faction.go` now wraps with `faction: %w`
   and `dryrun.go` with `dryrun: %w`. Plan body reconciled: Commit 3 Tasks 6 and 9.

6. **`transport_lifecycle_test.go` handles the `Register` error** (Commit 3). Planning
   missed this fifth caller. A bare `Register(...)` statement still compiles after the
   error return lands, so the compiler gave no signal. The test registers a handler after
   the harness builds the engine, so a duplicate ID would otherwise fail silently. Plan
   body reconciled: Commit 3 Task 12 added.

7. **A recoverable turn-start failure falls through to bookkeeping** (Commit 4). The
   planned block copied the orchestrator's `return nil // continue past this
   sub-engine call`. Elsewhere each phase makes one sub-engine call, so `return nil`
   continues past that call. Here the turn-start call comes before `ApplyBookkeeping`,
   so `return nil` skipped the turn's income and maintenance with no signal. Now a
   recoverable error is logged and sent to `OnError`, `startMutations` is set to nil,
   and bookkeeping runs. That turn's payout is lost, not retried: a retry is the
   double payout the guard prevents. A fence comment marks the spot. Plan body
   reconciled: Commit 4 Task 4.
   *Amended after commit:* no current error reaches this branch. `recoverableErrors`
   lists only the three action sentinels, and neither `applyAndRecord` nor
   `ApplyBookkeeping` returns one, so every `IsRecoverable` branch in the bookkeeping
   phase is dead today. The fall-through is kept: it costs no more than `return nil`,
   and it is correct if a sentinel this path can return is ever added.
   *Amended at interim review:* the same dead branch has a worse outcome on the
   `ApplyBookkeeping` side. The turn-start batch is applied and recorded before
   `ApplyBookkeeping` sets `BookkeepingApplied`. If `ApplyBookkeeping` returned a
   recoverable error, the phase would return nil, `RunFactionTurn` would save the payout
   with `BookkeepingApplied` still false, and a resume would pay a second time. The
   observer would also not see the payout. No code change: no error reaches the branch
   today, and `recoverableErrors` already warns that a new sentinel changes the
   orchestrator's abort semantics. Whoever adds a sentinel this path can return must
   close this first.

8. **Catalog rows name `RollTable`, not `AbilityDefinition`** (Commit 4). The planned
   rows kept the labels `AbilityDefinition.Die` and `AbilityDefinition.Outcomes`, but
   `AbilityDefinition` now holds only `Steps`. The rows now name `RollTable.Die` and
   `RollTable.Outcomes`, and the outcome field list adds `DestroySelf`, which the
   struct has. Plan body reconciled: Commit 4 Task 12.

9. **`FireTurnStart` carries a fence comment on the unchecked `Fire` call** (Commit 4).
   It matches the comment on `Register` in `effectsListHandler.Apply`; both rely on
   `RegisterEffectsList` rejecting an entry with no route. Plan body reconciled:
   Commit 4 Task 1.
   *Amended by Decision Record 12:* the `FireTurnStart` fence now names its real
   assumption.

10. **Accepted sets are keyed per trigger** (interim-review fix-up). The flat
    `Triggers`, `Owners`, `Scopes`, `Filters` and `Gates` fields made rule 4 accept every
    combination of trigger and owner. Book of Secrets' `reroll` accepts `own_roll` with
    `owner = "own"` and `enemy_roll_at_me` with `owner = "rival"`, so `own_roll` with
    `rival` passed validation and would silently do nothing — the gap Decision 2 added
    the owner check to close. `Registrar.Triggers` is now
    `map[domain.EffectTrigger]Accepts`, and `Accepts` holds `Owners`, `Scopes`, `Filters`
    and `Gates`. `Register` and `Fire` stay on the registrar, because `IsPhaseTrigger`
    picks the route by the kind of trigger. Owner, scope, filter and gate errors now end
    `for trigger "<trigger>"`. `effect_test.go` adds a `reroll`-shaped registrar with an
    accepted second-trigger case and an "owner accepted only under another trigger"
    case; the "no registrar" case moves to `bonus_die`. Done before Effort 2, so
    `coin_gain` is the only registrar to migrate. Amends Decisions 1 and 2; plan prose
    reconciled, task blocks marked.

11. **Zero-sided dice and zero amounts fail the load** (interim-review fix-up).
    `"1d0"` loaded, and `RandRoller.Roll(0)` then panicked in `rand.IntN(0)` the first
    time the dice rolled. `parseDice` now rejects a die with zero sides when there is at
    least one die. The check is in `parseDice`, not only in `convertAmount`, so
    `attack.damage`, `counter` and roll tables get it too; the hole predates this
    branch for those fields. `convertAmount` also rejects an amount that can only be
    zero (`"0"`, `"0d6"`), which matches the `> 0` checks on every other parameter block.
    Reverses Decision Record 3's note on zero `coin_gain` amounts and amends Decision 15.
    Tests: `parseDice("1d0")`, and loud-failure rows for `amount = "0"` and
    `amount = "1d0"`. `TestLoad` confirms the SWN data has no zero-sided die.

12. **The `FireTurnStart` fence names its real assumption** (interim-review fix-up).
    `FireTurnStart` reads the live `rulebook.Assets`, not the lists
    `RegisterEffectsList` checked, so "RegisterEffectsList rejects it" was true only if
    the rulebook does not change after construction. The fence now says that. No nil
    check is added. The `Register` fence in `effectsListHandler.Apply` is unchanged: it
    reads the definition that was checked.

13. **Comments that only restate the code are removed** (interim-review fix-up). The
    plan spelled out doc comments that `CLAUDE.md`'s "No docstrings" rule does not
    allow. Removed: `OncePerKey`, both `Used()`, `coinGainRegistrar`, `Registrars` (its
    "fresh map, so no caller can mutate another's table" also overclaimed — the accepted
    sets' slices are shared), `validateEntry` and `toEnum`. `RegisterEffectsList` keeps
    only its "why" sentence. Comments that give a non-obvious reason or invariant stay,
    `convertEffectParams` and `presentParamBlocks` included. Comments outside this
    branch's diff are not touched.

---

## Shared Context

**Model.** This Plan session ran on Opus. Commit 1 (the domain contract every later
effort cites) warrants Opus; Commits 2–4 are specified literally and suit Sonnet.
Surface the recommendation at the Plan → Execution transition.

**Overview amendments.** Decisions 5, 8, 9, 14 and 15 above amend overview Planning
Decisions 5, 6, 7 and 11 and the Effort 5 outline's relocate-params commit. The overview
was reconciled with them in this Plan session; it cites this plan where they apply.

**Test patterns.** Loader rules are tested by writing a TOML body to a temp
`assets/` directory and calling `loadAssets` (`effectTestAsset` and
`loadAssetsFromTOML`, Commit 2 Task 9). Effects-engine tests inject a test-local
registrar table, which is what the injected table (Decision 1) buys. Scenario tests use
`testharness.NewHarness`, which loads the real `rulebooks/swn` data — so after Commit 4
every harness test also proves Party Machine validates at construction.

**Data copies.** `campaigns/test-camp/rulebook/assets/` holds byte-identical copies of
the SWN asset files. Data edits apply to both.
## Out of Scope

- Every effect other than `coin_gain` (Efforts 2–5), and every dispatch item (1–8).
- The `on_purchase` route. `IsPhaseTrigger` names it and Rule 4 enforces its route, but
  nothing reads a definition at purchase until Effort 5.
- Relocate behaviour for `Categories`, `ExcludeTypes` and `CoinCostMode` (Effort 5).
- Migrating `[transport]` onto the effects list (D-3).
- Updating `effect-mutation.md` / `hooks.md` Key Decisions — graduation runs at
  pre-merge (Decisions 4 and 7 are the candidates).

---

## Work Breakdown

### Phase 1 — Foundation

Single phase, four commits, executed in order. Each commit builds with green tests.

#### Commit 1 — `refactor: add the special effect model to the domain`

The complete contract Efforts 2–5 cite. Pure types; nothing reads them yet except one
test fix forced by `RelocateParams` growing slice fields.

##### Task 1 — `internal/faction/domain/effect.go` (new)

> Changed by the interim-review fix-up (Decision Record 13); the source is current.

Full contents:

```go
package domain

// SpecialEffectType names an S-side effect. The effects engine keys its registrar
// table on it, as the ability dispatch keys on AbilityEffectType.
type SpecialEffectType string

const (
	SpecialCoinGain       SpecialEffectType = "coin_gain"
	SpecialCoinSteal      SpecialEffectType = "coin_steal"
	SpecialIgnoreCoinLoss SpecialEffectType = "ignore_coin_loss"
	SpecialSelfDamage     SpecialEffectType = "self_damage"
	SpecialApplyStealth   SpecialEffectType = "apply_stealth"
	SpecialClearStealth   SpecialEffectType = "clear_stealth"
	SpecialBonusDie       SpecialEffectType = "bonus_die"
	SpecialRemoveTagDice  SpecialEffectType = "remove_tag_dice"
	SpecialReroll         SpecialEffectType = "reroll"
	SpecialTechLevelFloor SpecialEffectType = "tech_level_floor"
	SpecialCostSurcharge  SpecialEffectType = "cost_surcharge"
	SpecialGrantMovement  SpecialEffectType = "grant_movement"
	SpecialRestrictAttack SpecialEffectType = "restrict_attack"
)

type EffectTrigger string

const (
	TriggerTurnStart              EffectTrigger = "turn_start"
	TriggerOnPurchase             EffectTrigger = "on_purchase"
	TriggerOwnAttackHit           EffectTrigger = "own_attack_hit"
	TriggerOwnCounterHit          EffectTrigger = "own_counter_hit"
	TriggerOwnRoll                EffectTrigger = "own_roll"
	TriggerEnemyRollAtMe          EffectTrigger = "enemy_roll_at_me"
	TriggerStealthArrival         EffectTrigger = "stealth_arrival"
	TriggerCoinLoss               EffectTrigger = "coin_loss"
	TriggerTechLevelQuery         EffectTrigger = "tech_level_query"
	TriggerAssetCostQuery         EffectTrigger = "asset_cost_query"
	TriggerMovementOptionsQuery   EffectTrigger = "movement_options_query"
	TriggerAttackEligibilityQuery EffectTrigger = "attack_eligibility_query"
)

// EffectOwner names whose side of the trigger's event an entry matches. The
// reading differs per trigger; each registrar documents its own.
type EffectOwner string

const (
	OwnerOwn   EffectOwner = "own"
	OwnerRival EffectOwner = "rival"
	OwnerAny   EffectOwner = "any"
)

type EffectScope string

const (
	ScopeSelf    EffectScope = "self"
	ScopeWorld   EffectScope = "world"
	ScopeFaction EffectScope = "faction"
)

// EffectFilterName names an optional filter. owner and scope are not listed: every
// entry carries both, and a registrar accepts them by value.
type EffectFilterName string

const (
	FilterAttribute EffectFilterName = "attribute"
	FilterPhase     EffectFilterName = "phase"
	FilterAction    EffectFilterName = "action"
	FilterAssetType EffectFilterName = "asset_type"
	FilterCategory  EffectFilterName = "category"
	FilterStealthy  EffectFilterName = "stealthy"
)

type EffectGateName string

const (
	GateOncePer EffectGateName = "once_per"
	GateNoStack EffectGateName = "no_stack"
	GateContest EffectGateName = "contest"
)

// EffectRollPhase names a hooks.RollPhase; domain cannot import hooks.
type EffectRollPhase string

const (
	RollPhaseAttack      EffectRollPhase = "attack"
	RollPhaseDefense     EffectRollPhase = "defense"
	RollPhaseFactionTest EffectRollPhase = "faction_test"
	RollPhaseContested   EffectRollPhase = "contested"
)

// OncePerKey picks what a once-per-turn budget counts against.
type OncePerKey string

const (
	OncePerOwnerFaction  OncePerKey = "owner_faction"
	OncePerTargetFaction OncePerKey = "target_faction"
	OncePerAsset         OncePerKey = "asset"
)

// ContestRoller picks which side rolls as the contest's attacker: the declaring
// asset (Tripwire Cells), or the affected asset's faction testing to keep what it
// has (Panopticon Matrix).
type ContestRoller string

const (
	ContestRollerSelf   ContestRoller = "self"
	ContestRollerTarget ContestRoller = "target"
)

type RestrictSide string

const (
	RestrictDefense          RestrictSide = "defense"
	RestrictAttackAndCounter RestrictSide = "attack_and_counter"
)

type RestrictAllow string

const (
	AllowOnlyListed   RestrictAllow = "only_listed"
	AllowExceptListed RestrictAllow = "except_listed"
)

// EffectFilters narrows which events fire an entry. A zero optional field is an
// absent filter. Stealthy is a pointer because "stealthy = false" is a real
// filter, distinct from no filter.
type EffectFilters struct {
	Owner      EffectOwner
	Scope      EffectScope
	Attribute  FactionStat
	Phase      EffectRollPhase
	Action     string
	AssetTypes []AssetType
	Categories []FactionStat
	Stealthy   *bool
}

// Used lists the optional filters this entry sets, in declaration order.
func (filters EffectFilters) Used() []EffectFilterName {
	var present []EffectFilterName
	if filters.Attribute != "" {
		present = append(present, FilterAttribute)
	}
	if filters.Phase != "" {
		present = append(present, FilterPhase)
	}
	if filters.Action != "" {
		present = append(present, FilterAction)
	}
	if len(filters.AssetTypes) > 0 {
		present = append(present, FilterAssetType)
	}
	if len(filters.Categories) > 0 {
		present = append(present, FilterCategory)
	}
	if filters.Stealthy != nil {
		present = append(present, FilterStealthy)
	}
	return present
}

type EffectGates struct {
	OncePer OncePerKey
	NoStack bool
	Contest *EffectContest
}

// Used lists the gates this entry sets, in declaration order.
func (gates EffectGates) Used() []EffectGateName {
	var present []EffectGateName
	if gates.OncePer != "" {
		present = append(present, GateOncePer)
	}
	if gates.NoStack {
		present = append(present, GateNoStack)
	}
	if gates.Contest != nil {
		present = append(present, GateContest)
	}
	return present
}

type EffectContest struct {
	Contest
	Roller ContestRoller
}

type CoinGainParams struct {
	Amount DiceRoll
}

type CoinStealParams struct {
	Amount DiceRoll
}

type SelfDamageParams struct {
	Amount DiceRoll
}

type BonusDieParams struct {
	Sides int
	Count int
}

type RerollParams struct {
	Dice     int
	Elective bool
}

type TechLevelFloorParams struct {
	Level int
}

type CostSurchargeParams struct {
	Amount int
}

type RestrictAttackParams struct {
	Side  RestrictSide
	Allow RestrictAllow
}

type GrantMovementParams struct {
	Relocate RelocateParams
}

// EffectDefinition is one entry of an asset's [[effects]] list. At most one
// parameter pointer is set: the one named by Effect, when that effect takes
// parameters.
type EffectDefinition struct {
	Trigger        EffectTrigger
	Effect         SpecialEffectType
	Filters        EffectFilters
	Gates          EffectGates
	CoinGain       *CoinGainParams
	CoinSteal      *CoinStealParams
	SelfDamage     *SelfDamageParams
	BonusDie       *BonusDieParams
	Reroll         *RerollParams
	TechLevelFloor *TechLevelFloorParams
	CostSurcharge  *CostSurchargeParams
	RestrictAttack *RestrictAttackParams
	GrantMovement  *GrantMovementParams
}
```

##### Task 2 — `internal/faction/domain/asset.go`

`RelocateParams` grows the three fields Transit Web needs (Decision 14).

**Find:**
```go
// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per moved asset; AssetType
// filters "other" candidates (empty = any); MaxAssets caps the count moved (0 = unbounded).
type RelocateParams struct {
	Scope     RelocateScope
	Range     RelocateRange
	MaxHex    int
	CoinCost  int
	AssetType AssetType
	MaxAssets int
}
```
**Replace with:**
```go
// RelocateCostMode picks how a relocate charges CoinCost: once per moved asset, or once
// for the whole relocation (Transit Web: "for 1 Coin, any number of assets").
type RelocateCostMode string

const (
	RelocateCostPerAsset RelocateCostMode = "per_asset"
	RelocateCostFlat     RelocateCostMode = "flat"
)

// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per CoinCostMode; AssetType
// filters "other" candidates (empty = any), Categories narrows them by asset category
// (empty = any) and ExcludeTypes drops types; MaxAssets caps the count moved (0 = unbounded).
type RelocateParams struct {
	Scope        RelocateScope
	Range        RelocateRange
	MaxHex       int
	CoinCost     int
	CoinCostMode RelocateCostMode
	AssetType    AssetType
	Categories   []FactionStat
	ExcludeTypes []AssetType
	MaxAssets    int
}
```

##### Task 3 — `internal/faction/domain/asset.go`

**Find:**
```go
	Ability     *AbilityDefinition
	Transport   *TransportProfile
}
```
**Replace with:**
```go
	Ability     *AbilityDefinition
	Transport   *TransportProfile
	Effects     []EffectDefinition
}
```

##### Task 4 — `internal/faction/rulebook/rulebook_test.go`

`RelocateParams` now holds slices, so `!=` no longer compiles. Two subtests compare it.

**Find:**
```go
import (
	"slices"
	"strings"
	"testing"
```
**Replace with:**
```go
import (
	"reflect"
	"slices"
	"strings"
	"testing"
```

**Find:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1}
		if *step.Relocate != want {
```
**Replace with:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1}
		if !reflect.DeepEqual(*step.Relocate, want) {
```

**Find:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeRegion, AssetType: domain.TypeSpecialForces, MaxAssets: 0}
		if *step.Relocate != want {
```
**Replace with:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeRegion, AssetType: domain.TypeSpecialForces, MaxAssets: 0}
		if !reflect.DeepEqual(*step.Relocate, want) {
```

##### Task 5 — `internal/faction/domain/effect_test.go` (new)

Full contents:

```go
package domain_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestEffectFiltersUsed(t *testing.T) {
	stealthy := false
	cases := []struct {
		name    string
		filters domain.EffectFilters
		want    []domain.EffectFilterName
	}{
		{
			name:    "owner and scope only",
			filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			want:    nil,
		},
		{
			name: "every optional filter, stealthy false still counts",
			filters: domain.EffectFilters{
				Owner:      domain.OwnerRival,
				Scope:      domain.ScopeWorld,
				Attribute:  domain.StatCunning,
				Phase:      domain.RollPhaseDefense,
				Action:     "Expand Influence",
				AssetTypes: []domain.AssetType{domain.TypeStarship},
				Categories: []domain.FactionStat{domain.StatForce},
				Stealthy:   &stealthy,
			},
			want: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterAssetType, domain.FilterCategory, domain.FilterStealthy,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filters.Used(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Used() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEffectGatesUsed(t *testing.T) {
	cases := []struct {
		name  string
		gates domain.EffectGates
		want  []domain.EffectGateName
	}{
		{name: "no gates", gates: domain.EffectGates{}, want: nil},
		{
			name: "every gate",
			gates: domain.EffectGates{
				OncePer: domain.OncePerTargetFaction,
				NoStack: true,
				Contest: &domain.EffectContest{Roller: domain.ContestRollerSelf},
			},
			want: []domain.EffectGateName{domain.GateOncePer, domain.GateNoStack, domain.GateContest},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.gates.Used(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Used() = %v, want %v", got, tc.want)
			}
		})
	}
}
```

##### Commit message

```
refactor: add the special effect model to the domain

- EffectDefinition, EffectFilters, EffectGates, EffectContest and the
  per-effect parameter structs for all 13 A.001.3 effects
- trigger, owner, scope, filter, gate and roll-phase vocabularies
- Used() on filters and gates for registrar validation
- RelocateParams grows CoinCostMode, Categories and ExcludeTypes
- AssetDefinition.Effects
```

#### Commit 2 — `feat: parse the asset effects list`

The loader reads `[[effects]]`, enforces loud-failure rules 1–3 (with rule 2 extended to
missing blocks), and rejects every unknown key in an asset file.

##### Task 1 — `internal/faction/rulebook/dice.go`

> Changed by the interim-review fix-up (Decision Record 11); the source is current.

Append after `parseDice`:

```go
var flatAmountPattern = regexp.MustCompile(`^\d+$`)

// parseAmount converts an effect amount into a DiceRoll. A bare integer ("1") is a
// fixed amount — a zero-dice roll whose modifier is the value; anything else is dice
// notation. Returns nil without error for an empty string.
func parseAmount(s string) (*domain.DiceRoll, error) {
	if flatAmountPattern.MatchString(s) {
		value, _ := strconv.Atoi(s)
		return &domain.DiceRoll{Modifier: value}, nil
	}
	return parseDice(s)
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Add `strings` to the imports.

**Find:**
```go
import (
	"fmt"
	"path/filepath"
```
**Replace with:**
```go
import (
	"fmt"
	"path/filepath"
	"strings"
```

In `loadAssets`, fail on any key the records do not know (Decision 11).

**Find:**
```go
		var f assetFile
		if _, err := toml.DecodeFile(path, &f); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", path, err)
		}
```
**Replace with:**
```go
		var f assetFile
		metadata, err := toml.DecodeFile(path, &f)
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", path, err)
		}
		// deliberate: unknown keys fail the load, not skip, because the decoder drops them silently and a misspelt filter or gate would do nothing.
		if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
			keys := make([]string, len(undecoded))
			for i, key := range undecoded {
				keys[i] = key.String()
			}
			return nil, fmt.Errorf("unknown keys in %s: %s", path, strings.Join(keys, ", "))
		}
```

If `TestLoad` then fails with unknown keys on the existing SWN data, stop and report —
do not loosen the check.

##### Task 3 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
	Ability     *abilityRecord   `toml:"ability"`
	Transport   *transportRecord `toml:"transport"`
}
```
**Replace with:**
```go
	Ability     *abilityRecord   `toml:"ability"`
	Transport   *transportRecord `toml:"transport"`
	Effects     []effectRecord   `toml:"effects"`
}
```

##### Task 4 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
type relocateRecord struct {
	Scope     string `toml:"scope"`
	Range     string `toml:"range"`
	MaxHex    int    `toml:"max_hex"`
	CoinCost  int    `toml:"coin_cost"`
	AssetType string `toml:"asset_type"`
	MaxAssets int    `toml:"max_assets"`
}
```
**Replace with:**
```go
type relocateRecord struct {
	Scope        string   `toml:"scope"`
	Range        string   `toml:"range"`
	MaxHex       int      `toml:"max_hex"`
	CoinCost     int      `toml:"coin_cost"`
	CoinCostMode string   `toml:"coin_cost_mode"`
	AssetType    string   `toml:"asset_type"`
	Categories   []string `toml:"categories"`
	ExcludeTypes []string `toml:"exclude_types"`
	MaxAssets    int      `toml:"max_assets"`
}
```

##### Task 5 — `internal/faction/rulebook/rulebook.go`

In `convertAsset`, convert the effects list.

**Find:**
```go
	transport, err := convertTransport(r.Transport)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}
```
**Replace with:**
```go
	transport, err := convertTransport(r.Transport)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	effects, err := convertEffects(r.Effects)
	if err != nil {
		return nil, err
	}
```

**Find:**
```go
		Ability:     ability,
		Transport:   transport,
	}, nil
}
```
**Replace with:**
```go
		Ability:     ability,
		Transport:   transport,
		Effects:     effects,
	}, nil
}
```

##### Task 6 — `internal/faction/rulebook/rulebook.go`

Extract relocate conversion so the A-side ability and `grant_movement` share it, and
reject the Effort-5 fields on the A-side (Decision 14).

In `convertAbilityStep`:

**Find:**
```go
	if relocate := record.Relocate; relocate != nil {
		scope, err := toRelocateScope(relocate.Scope)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		rng, err := toRelocateRange(relocate.Range)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		params := &domain.RelocateParams{
			Scope:     scope,
			Range:     rng,
			MaxHex:    relocate.MaxHex,
			CoinCost:  relocate.CoinCost,
			MaxAssets: relocate.MaxAssets,
		}
		if relocate.AssetType != "" {
			assetType, err := toAssetType(relocate.AssetType)
			if err != nil {
				return domain.AbilityStep{}, fmt.Errorf("relocate asset_type: %w", err)
			}
			params.AssetType = assetType
		}
		step.Relocate = params
	}
```
**Replace with:**
```go
	if relocate := record.Relocate; relocate != nil {
		params, err := convertRelocate(relocate)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		// deliberate: reject, not ignore, because the action-spent relocate handler does not read these fields yet.
		if len(params.Categories) > 0 || len(params.ExcludeTypes) > 0 || params.CoinCostMode == domain.RelocateCostFlat {
			return domain.AbilityStep{}, fmt.Errorf("relocate categories, exclude_types and coin_cost_mode = %q are not supported on an [ability]", domain.RelocateCostFlat)
		}
		step.Relocate = params
	}
```

Add after `toRelocateRange`:

```go
func toRelocateCostMode(s string) (domain.RelocateCostMode, error) {
	switch s {
	case "", "per_asset":
		return domain.RelocateCostPerAsset, nil
	case "flat":
		return domain.RelocateCostFlat, nil
	default:
		return "", fmt.Errorf("unknown relocate coin_cost_mode: %q", s)
	}
}

func convertRelocate(record *relocateRecord) (*domain.RelocateParams, error) {
	scope, err := toRelocateScope(record.Scope)
	if err != nil {
		return nil, err
	}
	rng, err := toRelocateRange(record.Range)
	if err != nil {
		return nil, err
	}
	costMode, err := toRelocateCostMode(record.CoinCostMode)
	if err != nil {
		return nil, err
	}
	params := &domain.RelocateParams{
		Scope:        scope,
		Range:        rng,
		MaxHex:       record.MaxHex,
		CoinCost:     record.CoinCost,
		CoinCostMode: costMode,
		MaxAssets:    record.MaxAssets,
	}
	if record.AssetType != "" {
		assetType, err := toAssetType(record.AssetType)
		if err != nil {
			return nil, fmt.Errorf("relocate asset_type: %w", err)
		}
		params.AssetType = assetType
	}
	for i, value := range record.Categories {
		category, err := toFactionStat(value)
		if err != nil {
			return nil, fmt.Errorf("relocate categories[%d]: %w", i, err)
		}
		params.Categories = append(params.Categories, category)
	}
	for i, value := range record.ExcludeTypes {
		assetType, err := toAssetType(value)
		if err != nil {
			return nil, fmt.Errorf("relocate exclude_types[%d]: %w", i, err)
		}
		params.ExcludeTypes = append(params.ExcludeTypes, assetType)
	}
	return params, nil
}
```

##### Task 7 — `internal/faction/rulebook/effects.go` (new)

> Changed by the interim-review fix-up (Decision Records 11 and 13); the source is current.

Full contents:

```go
package rulebook

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type effectRecord struct {
	Trigger        string                `toml:"trigger"`
	Effect         string                `toml:"effect"`
	Filters        *effectFiltersRecord  `toml:"filters"`
	Gates          *effectGatesRecord    `toml:"gates"`
	CoinGain       *amountRecord         `toml:"coin_gain"`
	CoinSteal      *amountRecord         `toml:"coin_steal"`
	SelfDamage     *amountRecord         `toml:"self_damage"`
	BonusDie       *bonusDieRecord       `toml:"bonus_die"`
	Reroll         *rerollRecord         `toml:"reroll"`
	TechLevelFloor *techLevelFloorRecord `toml:"tech_level_floor"`
	CostSurcharge  *costSurchargeRecord  `toml:"cost_surcharge"`
	RestrictAttack *restrictAttackRecord `toml:"restrict_attack"`
	GrantMovement  *grantMovementRecord  `toml:"grant_movement"`
}

type effectFiltersRecord struct {
	Owner     string   `toml:"owner"`
	Scope     string   `toml:"scope"`
	Attribute string   `toml:"attribute"`
	Phase     string   `toml:"phase"`
	Action    string   `toml:"action"`
	AssetType []string `toml:"asset_type"`
	Category  []string `toml:"category"`
	Stealthy  *bool    `toml:"stealthy"`
}

type effectGatesRecord struct {
	OncePer string               `toml:"once_per"`
	NoStack bool                 `toml:"no_stack"`
	Contest *effectContestRecord `toml:"contest"`
}

type effectContestRecord struct {
	Roller       string `toml:"roller"`
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
}

type amountRecord struct {
	Amount string `toml:"amount"`
}

type bonusDieRecord struct {
	Sides int `toml:"sides"`
	Count int `toml:"count"`
}

type rerollRecord struct {
	Dice     int  `toml:"dice"`
	Elective bool `toml:"elective"`
}

type techLevelFloorRecord struct {
	Level int `toml:"level"`
}

type costSurchargeRecord struct {
	Amount int `toml:"amount"`
}

type restrictAttackRecord struct {
	Side  string `toml:"side"`
	Allow string `toml:"allow"`
}

type grantMovementRecord struct {
	Relocate *relocateRecord `toml:"relocate"`
}

var (
	effectTriggers = []domain.EffectTrigger{
		domain.TriggerTurnStart, domain.TriggerOnPurchase, domain.TriggerOwnAttackHit,
		domain.TriggerOwnCounterHit, domain.TriggerOwnRoll, domain.TriggerEnemyRollAtMe,
		domain.TriggerStealthArrival, domain.TriggerCoinLoss, domain.TriggerTechLevelQuery,
		domain.TriggerAssetCostQuery, domain.TriggerMovementOptionsQuery,
		domain.TriggerAttackEligibilityQuery,
	}
	specialEffects = []domain.SpecialEffectType{
		domain.SpecialCoinGain, domain.SpecialCoinSteal, domain.SpecialIgnoreCoinLoss,
		domain.SpecialSelfDamage, domain.SpecialApplyStealth, domain.SpecialClearStealth,
		domain.SpecialBonusDie, domain.SpecialRemoveTagDice, domain.SpecialReroll,
		domain.SpecialTechLevelFloor, domain.SpecialCostSurcharge, domain.SpecialGrantMovement,
		domain.SpecialRestrictAttack,
	}
	effectOwners     = []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival, domain.OwnerAny}
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction}
	effectRollPhases = []domain.EffectRollPhase{
		domain.RollPhaseAttack, domain.RollPhaseDefense, domain.RollPhaseFactionTest, domain.RollPhaseContested,
	}
	oncePerKeys    = []domain.OncePerKey{domain.OncePerOwnerFaction, domain.OncePerTargetFaction, domain.OncePerAsset}
	contestRollers = []domain.ContestRoller{domain.ContestRollerSelf, domain.ContestRollerTarget}
	restrictSides  = []domain.RestrictSide{domain.RestrictDefense, domain.RestrictAttackAndCounter}
	restrictAllows = []domain.RestrictAllow{domain.AllowOnlyListed, domain.AllowExceptListed}
)

func convertEffects(records []effectRecord) ([]domain.EffectDefinition, error) {
	if len(records) == 0 {
		return nil, nil
	}
	definitions := make([]domain.EffectDefinition, 0, len(records))
	for i, record := range records {
		definition, err := convertEffect(record)
		if err != nil {
			return nil, fmt.Errorf("effects[%d]: %w", i, err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func convertEffect(record effectRecord) (domain.EffectDefinition, error) {
	trigger, err := toEnum("effect trigger", record.Trigger, effectTriggers)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	effectType, err := toEnum("special effect", record.Effect, specialEffects)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	if record.Filters == nil {
		return domain.EffectDefinition{}, fmt.Errorf("missing [filters] block: owner and scope are required")
	}
	filters, err := convertEffectFilters(*record.Filters)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	gates, err := convertEffectGates(record.Gates)
	if err != nil {
		return domain.EffectDefinition{}, err
	}
	for _, block := range presentParamBlocks(record) {
		if block != effectType {
			return domain.EffectDefinition{}, fmt.Errorf("parameter block [%s] does not match effect %q", block, effectType)
		}
	}
	definition := domain.EffectDefinition{Trigger: trigger, Effect: effectType, Filters: filters, Gates: gates}
	if err := convertEffectParams(record, &definition); err != nil {
		return domain.EffectDefinition{}, err
	}
	return definition, nil
}

func convertEffectFilters(record effectFiltersRecord) (domain.EffectFilters, error) {
	if record.Owner == "" {
		return domain.EffectFilters{}, fmt.Errorf("filters.owner is required")
	}
	if record.Scope == "" {
		return domain.EffectFilters{}, fmt.Errorf("filters.scope is required")
	}
	owner, err := toEnum("effect owner", record.Owner, effectOwners)
	if err != nil {
		return domain.EffectFilters{}, err
	}
	scope, err := toEnum("effect scope", record.Scope, effectScopes)
	if err != nil {
		return domain.EffectFilters{}, err
	}
	filters := domain.EffectFilters{Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy}
	if record.Attribute != "" {
		attribute, err := toFactionStat(record.Attribute)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.attribute: %w", err)
		}
		filters.Attribute = attribute
	}
	if record.Phase != "" {
		phase, err := toEnum("roll phase", record.Phase, effectRollPhases)
		if err != nil {
			return domain.EffectFilters{}, err
		}
		filters.Phase = phase
	}
	for i, value := range record.AssetType {
		assetType, err := toAssetType(value)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.asset_type[%d]: %w", i, err)
		}
		filters.AssetTypes = append(filters.AssetTypes, assetType)
	}
	for i, value := range record.Category {
		category, err := toFactionStat(value)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.category[%d]: %w", i, err)
		}
		filters.Categories = append(filters.Categories, category)
	}
	return filters, nil
}

func convertEffectGates(record *effectGatesRecord) (domain.EffectGates, error) {
	if record == nil {
		return domain.EffectGates{}, nil
	}
	gates := domain.EffectGates{NoStack: record.NoStack}
	if record.OncePer != "" {
		key, err := toEnum("once_per key", record.OncePer, oncePerKeys)
		if err != nil {
			return domain.EffectGates{}, err
		}
		gates.OncePer = key
	}
	if contest := record.Contest; contest != nil {
		roller, err := toEnum("contest roller", contest.Roller, contestRollers)
		if err != nil {
			return domain.EffectGates{}, err
		}
		attacker, err := toFactionStat(contest.AttackerStat)
		if err != nil {
			return domain.EffectGates{}, fmt.Errorf("contest attacker_stat: %w", err)
		}
		defender, err := toFactionStat(contest.DefenderStat)
		if err != nil {
			return domain.EffectGates{}, fmt.Errorf("contest defender_stat: %w", err)
		}
		gates.Contest = &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender},
			Roller:  roller,
		}
	}
	return gates, nil
}

// presentParamBlocks names the parameter blocks set on record. Each block is named
// after the effect it configures, so a block for any other effect is a mismatch.
func presentParamBlocks(record effectRecord) []domain.SpecialEffectType {
	var present []domain.SpecialEffectType
	if record.CoinGain != nil {
		present = append(present, domain.SpecialCoinGain)
	}
	if record.CoinSteal != nil {
		present = append(present, domain.SpecialCoinSteal)
	}
	if record.SelfDamage != nil {
		present = append(present, domain.SpecialSelfDamage)
	}
	if record.BonusDie != nil {
		present = append(present, domain.SpecialBonusDie)
	}
	if record.Reroll != nil {
		present = append(present, domain.SpecialReroll)
	}
	if record.TechLevelFloor != nil {
		present = append(present, domain.SpecialTechLevelFloor)
	}
	if record.CostSurcharge != nil {
		present = append(present, domain.SpecialCostSurcharge)
	}
	if record.RestrictAttack != nil {
		present = append(present, domain.SpecialRestrictAttack)
	}
	if record.GrantMovement != nil {
		present = append(present, domain.SpecialGrantMovement)
	}
	return present
}

// convertEffectParams fills the parameter pointer for definition.Effect. An effect
// that takes parameters must carry its block; the four that take none
// (ignore_coin_loss, apply_stealth, clear_stealth, remove_tag_dice) fall through.
func convertEffectParams(record effectRecord, definition *domain.EffectDefinition) error {
	switch definition.Effect {
	case domain.SpecialCoinGain:
		amount, err := convertAmount(definition.Effect, record.CoinGain)
		if err != nil {
			return err
		}
		definition.CoinGain = &domain.CoinGainParams{Amount: amount}
	case domain.SpecialCoinSteal:
		amount, err := convertAmount(definition.Effect, record.CoinSteal)
		if err != nil {
			return err
		}
		definition.CoinSteal = &domain.CoinStealParams{Amount: amount}
	case domain.SpecialSelfDamage:
		amount, err := convertAmount(definition.Effect, record.SelfDamage)
		if err != nil {
			return err
		}
		definition.SelfDamage = &domain.SelfDamageParams{Amount: amount}
	case domain.SpecialBonusDie:
		if record.BonusDie == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.BonusDie.Sides <= 0 || record.BonusDie.Count <= 0 {
			return fmt.Errorf("bonus_die sides and count must be > 0, got %d and %d", record.BonusDie.Sides, record.BonusDie.Count)
		}
		definition.BonusDie = &domain.BonusDieParams{Sides: record.BonusDie.Sides, Count: record.BonusDie.Count}
	case domain.SpecialReroll:
		if record.Reroll == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.Reroll.Dice <= 0 {
			return fmt.Errorf("reroll dice must be > 0, got %d", record.Reroll.Dice)
		}
		definition.Reroll = &domain.RerollParams{Dice: record.Reroll.Dice, Elective: record.Reroll.Elective}
	case domain.SpecialTechLevelFloor:
		if record.TechLevelFloor == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.TechLevelFloor.Level <= 0 {
			return fmt.Errorf("tech_level_floor level must be > 0, got %d", record.TechLevelFloor.Level)
		}
		definition.TechLevelFloor = &domain.TechLevelFloorParams{Level: record.TechLevelFloor.Level}
	case domain.SpecialCostSurcharge:
		if record.CostSurcharge == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.CostSurcharge.Amount <= 0 {
			return fmt.Errorf("cost_surcharge amount must be > 0, got %d", record.CostSurcharge.Amount)
		}
		definition.CostSurcharge = &domain.CostSurchargeParams{Amount: record.CostSurcharge.Amount}
	case domain.SpecialRestrictAttack:
		if record.RestrictAttack == nil {
			return missingParamBlock(definition.Effect)
		}
		side, err := toEnum("restrict_attack side", record.RestrictAttack.Side, restrictSides)
		if err != nil {
			return err
		}
		allow, err := toEnum("restrict_attack allow", record.RestrictAttack.Allow, restrictAllows)
		if err != nil {
			return err
		}
		definition.RestrictAttack = &domain.RestrictAttackParams{Side: side, Allow: allow}
	case domain.SpecialGrantMovement:
		if record.GrantMovement == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.GrantMovement.Relocate == nil {
			return fmt.Errorf("grant_movement requires a [grant_movement.relocate] block")
		}
		relocate, err := convertRelocate(record.GrantMovement.Relocate)
		if err != nil {
			return fmt.Errorf("grant_movement: %w", err)
		}
		definition.GrantMovement = &domain.GrantMovementParams{Relocate: *relocate}
	}
	return nil
}

func convertAmount(effectType domain.SpecialEffectType, record *amountRecord) (domain.DiceRoll, error) {
	if record == nil {
		return domain.DiceRoll{}, missingParamBlock(effectType)
	}
	amount, err := parseAmount(record.Amount)
	if err != nil {
		return domain.DiceRoll{}, fmt.Errorf("%s amount: %w", effectType, err)
	}
	if amount == nil {
		return domain.DiceRoll{}, fmt.Errorf("%s amount is required", effectType)
	}
	return *amount, nil
}

func missingParamBlock(effectType domain.SpecialEffectType) error {
	return fmt.Errorf("effect %q requires a [%s] block", effectType, effectType)
}

// toEnum matches value against a string-backed vocabulary.
func toEnum[T ~string](kind, value string, allowed []T) (T, error) {
	for _, candidate := range allowed {
		if value == string(candidate) {
			return candidate, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("unknown %s: %q (want one of %v)", kind, value, allowed)
}
```

##### Task 8 — `internal/faction/rulebook/rulebook_test.go`

The A-side relocate now normalizes an absent `coin_cost_mode` to per-asset.

**Find:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1}
```
**Replace with:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, CoinCostMode: domain.RelocateCostPerAsset, AssetType: domain.TypeSpecialForces, MaxAssets: 1}
```

**Find:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeRegion, AssetType: domain.TypeSpecialForces, MaxAssets: 0}
```
**Replace with:**
```go
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeRegion, CoinCostMode: domain.RelocateCostPerAsset, AssetType: domain.TypeSpecialForces, MaxAssets: 0}
```

In the Seductress subtest, the relocate step is compared field by field, so it needs no
change.

##### Task 9 — `internal/faction/rulebook/effects_test.go` (new)

> Changed by the interim-review fix-up (Decision Record 11); the source is current.

Full contents:

```go
package rulebook

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const effectTestAsset = `
[assets.Machine]
id = "T-001"
name = "Machine"
category = "Cunning"
min_rating = 1
hp = 1
cost = 1
tech_level = 0
type = "Facility"
flags = ["S"]
counter = ""
description = "Test asset."
`

func loadAssetsFromTOML(t *testing.T, body string) (map[string]*domain.AssetDefinition, error) {
	t.Helper()
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("creating assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "test_assets.toml"), []byte(body), 0o644); err != nil {
		t.Fatalf("writing assets file: %v", err)
	}
	return loadAssets(dir)
}

func TestLoadEffects_ParsesEntries(t *testing.T) {
	assets, err := loadAssetsFromTOML(t, effectTestAsset+`
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"

[[assets.Machine.effects]]
trigger = "own_attack_hit"
effect  = "coin_steal"
[assets.Machine.effects.filters]
owner      = "rival"
scope      = "world"
attribute  = "Cunning"
phase      = "defense"
action     = "Expand Influence"
asset_type = ["Starship"]
category   = ["Force", "Wealth"]
stealthy   = true
[assets.Machine.effects.gates]
once_per = "target_faction"
no_stack = true
[assets.Machine.effects.gates.contest]
roller        = "target"
attacker_stat = "Cunning"
defender_stat = "Wealth"
[assets.Machine.effects.coin_steal]
amount = "1d4"

[[assets.Machine.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"
[assets.Machine.effects.filters]
owner = "own"
scope = "faction"
[assets.Machine.effects.grant_movement.relocate]
scope          = "other"
range          = "hexes"
max_hex        = 3
coin_cost      = 1
coin_cost_mode = "flat"
categories     = ["Cunning", "Wealth"]
exclude_types  = ["Starship"]

[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
`)
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	stealthy := true
	want := []domain.EffectDefinition{
		{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		},
		{
			Trigger: domain.TriggerOwnAttackHit,
			Effect:  domain.SpecialCoinSteal,
			Filters: domain.EffectFilters{
				Owner:      domain.OwnerRival,
				Scope:      domain.ScopeWorld,
				Attribute:  domain.StatCunning,
				Phase:      domain.RollPhaseDefense,
				Action:     "Expand Influence",
				AssetTypes: []domain.AssetType{domain.TypeStarship},
				Categories: []domain.FactionStat{domain.StatForce, domain.StatWealth},
				Stealthy:   &stealthy,
			},
			Gates: domain.EffectGates{
				OncePer: domain.OncePerTargetFaction,
				NoStack: true,
				Contest: &domain.EffectContest{
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
					Roller:  domain.ContestRollerTarget,
				},
			},
			CoinSteal: &domain.CoinStealParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}},
		},
		{
			Trigger: domain.TriggerMovementOptionsQuery,
			Effect:  domain.SpecialGrantMovement,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			GrantMovement: &domain.GrantMovementParams{Relocate: domain.RelocateParams{
				Scope:        domain.RelocateOther,
				Range:        domain.RangeHexes,
				MaxHex:       3,
				CoinCost:     1,
				CoinCostMode: domain.RelocateCostFlat,
				Categories:   []domain.FactionStat{domain.StatCunning, domain.StatWealth},
				ExcludeTypes: []domain.AssetType{domain.TypeStarship},
			}},
		},
		{
			Trigger: domain.TriggerOnPurchase,
			Effect:  domain.SpecialApplyStealth,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		},
	}
	if got := assets["T-001"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("effects:\n got  %+v\n want %+v", got, want)
	}
}

func TestLoadEffects_ParsesEveryParameterBlock(t *testing.T) {
	entry := func(trigger, effect, block string) string {
		return `
[[assets.Machine.effects]]
trigger = "` + trigger + `"
effect  = "` + effect + `"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
[assets.Machine.effects.` + effect + `]
` + block
	}
	assets, err := loadAssetsFromTOML(t, effectTestAsset+
		entry("own_attack_hit", "self_damage", `amount = "1d4"`)+
		entry("own_roll", "bonus_die", "sides = 10\ncount = 1")+
		entry("own_roll", "reroll", "dice = 1\nelective = true")+
		entry("tech_level_query", "tech_level_floor", "level = 4")+
		entry("asset_cost_query", "cost_surcharge", "amount = 1")+
		entry("attack_eligibility_query", "restrict_attack", `side = "defense"`+"\n"+`allow = "only_listed"`))
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	effects := assets["T-001"].Effects
	if len(effects) != 6 {
		t.Fatalf("effects: got %d, want 6", len(effects))
	}
	checks := []struct {
		name string
		got  any
		want any
	}{
		{"self_damage", effects[0].SelfDamage, &domain.SelfDamageParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}}},
		{"bonus_die", effects[1].BonusDie, &domain.BonusDieParams{Sides: 10, Count: 1}},
		{"reroll", effects[2].Reroll, &domain.RerollParams{Dice: 1, Elective: true}},
		{"tech_level_floor", effects[3].TechLevelFloor, &domain.TechLevelFloorParams{Level: 4}},
		{"cost_surcharge", effects[4].CostSurcharge, &domain.CostSurchargeParams{Amount: 1}},
		{"restrict_attack", effects[5].RestrictAttack, &domain.RestrictAttackParams{Side: domain.RestrictDefense, Allow: domain.AllowOnlyListed}},
	}
	for _, check := range checks {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("%s: got %+v, want %+v", check.name, check.got, check.want)
		}
	}
}

func TestLoadEffects_LoudFailure(t *testing.T) {
	const filters = `
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
`
	cases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name: "unknown trigger",
			body: `
[[assets.Machine.effects]]
trigger = "dawn"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `unknown effect trigger: "dawn"`,
		},
		{
			name: "unknown effect",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_fountain"` + filters,
			wantErr: `unknown special effect: "coin_fountain"`,
		},
		{
			name: "unknown filter name",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `bogus_filter = "x"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "bogus_filter",
		},
		{
			name: "unknown gate name",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.gates]
bogus_gate = true
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "bogus_gate",
		},
		{
			name: "unknown parameter block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_fountain]
amount = "1"
`,
			wantErr: "coin_fountain",
		},
		{
			name:    "unknown key elsewhere in the asset",
			body:    "bogus_field = 1\n",
			wantErr: "bogus_field",
		},
		{
			name: "parameter block does not match the effect",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_steal]
amount = "1"
`,
			wantErr: `parameter block [coin_steal] does not match effect "coin_gain"`,
		},
		{
			name: "parameter block on an effect that takes none",
			body: `
[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `parameter block [coin_gain] does not match effect "apply_stealth"`,
		},
		{
			name: "missing parameter block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters,
			wantErr: `effect "coin_gain" requires a [coin_gain] block`,
		},
		{
			name: "missing amount",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
`,
			wantErr: "coin_gain amount is required",
		},
		{
			name: "missing filters block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "missing [filters] block",
		},
		{
			name: "missing owner",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "filters.owner is required",
		},
		{
			name: "missing scope",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "own"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "filters.scope is required",
		},
		{
			name: "unknown owner value",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "mine"
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `unknown effect owner: "mine"`,
		},
		{
			name: "contest without a roller",
			body: `
[[assets.Machine.effects]]
trigger = "stealth_arrival"
effect  = "clear_stealth"` + filters + `
[assets.Machine.effects.gates.contest]
attacker_stat = "Cunning"
defender_stat = "Cunning"
`,
			wantErr: `unknown contest roller: ""`,
		},
		{
			name: "grant_movement without a relocate block",
			body: `
[[assets.Machine.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"` + filters + `
[assets.Machine.effects.grant_movement]
`,
			wantErr: "grant_movement requires a [grant_movement.relocate] block",
		},
		{
			name: "bonus_die with zero sides",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "bonus_die"` + filters + `
[assets.Machine.effects.bonus_die]
sides = 0
count = 1
`,
			wantErr: "bonus_die sides and count must be > 0, got 0 and 1",
		},
		{
			name: "reroll with zero dice",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "reroll"` + filters + `
[assets.Machine.effects.reroll]
dice = 0
`,
			wantErr: "reroll dice must be > 0, got 0",
		},
		{
			name: "tech_level_floor with zero level",
			body: `
[[assets.Machine.effects]]
trigger = "tech_level_query"
effect  = "tech_level_floor"` + filters + `
[assets.Machine.effects.tech_level_floor]
level = 0
`,
			wantErr: "tech_level_floor level must be > 0, got 0",
		},
		{
			name: "cost_surcharge with zero amount",
			body: `
[[assets.Machine.effects]]
trigger = "asset_cost_query"
effect  = "cost_surcharge"` + filters + `
[assets.Machine.effects.cost_surcharge]
amount = 0
`,
			wantErr: "cost_surcharge amount must be > 0, got 0",
		},
		{
			name: "effort-5 relocate fields on an action-spent ability",
			body: `
[assets.Machine.ability]
effect = "relocate"
[assets.Machine.ability.relocate]
scope      = "other"
max_hex    = 3
categories = ["Wealth"]
`,
			wantErr: "not supported on an [ability]",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadAssetsFromTOML(t, effectTestAsset+tc.body)
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		input string
		want  *domain.DiceRoll
	}{
		{"1", &domain.DiceRoll{Modifier: 1}},
		{"1d4", &domain.DiceRoll{NumDice: 1, Sides: 4}},
		{"2d6+1", &domain.DiceRoll{NumDice: 2, Sides: 6, Modifier: 1}},
		{"", nil},
	}
	for _, tc := range cases {
		got, err := parseAmount(tc.input)
		if err != nil {
			t.Fatalf("parseAmount(%q): %v", tc.input, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseAmount(%q) = %+v, want %+v", tc.input, got, tc.want)
		}
	}
	if _, err := parseAmount("lots"); err == nil {
		t.Error(`parseAmount("lots"): expected an error`)
	}
}
```

##### Commit message

```
feat: parse the asset effects list

- [[effects]] entries with filters, gates and a named parameter block
- parameter blocks for all 13 A.001.3 effects
- loud failure: unknown names, mismatched or missing parameter block,
  missing owner or scope
- reject every unknown key in an asset file
- parseAmount accepts a bare integer as a fixed amount
- shared relocate conversion; the action-spent relocate rejects the
  fields its handler does not read yet
```

#### Commit 3 — `feat: effect registrar contract with construction-time validation`

The registrar spec, the composed per-definition handler, rule 4 at engine construction,
the duplicate-handler guard, and deterministic `ApplyAll`. The registrar table is empty
in this commit; `coin_gain` fills it in Commit 4.

##### Task 1 — `internal/faction/engine/effect/registrar.go` (new)

> Changed by the interim-review fix-up (Decision Records 10 and 13); the source is current.

Full contents:

```go
package effect

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// Registrar is one S-side effect's spec: what the effect accepts in the data, and
// the code that runs it. RegisterEffectsList checks every [[effects]] entry against
// the accepted sets. An entry runs through Fire when its trigger is a phase event
// (IsPhaseTrigger) and through Register otherwise; a nil function means the effect
// has no such route.
type Registrar struct {
	Triggers []domain.EffectTrigger
	Owners   []domain.EffectOwner
	Scopes   []domain.EffectScope
	Filters  []domain.EffectFilterName
	Gates    []domain.EffectGateName
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
}

type RegisterContext struct {
	Faction  *domain.Faction
	Asset    *domain.Asset
	Def      *domain.AssetDefinition
	Entry    domain.EffectDefinition
	Registry *hooks.Registry
}

type FireContext struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Def     *domain.AssetDefinition
	Entry   domain.EffectDefinition
	Roller  domain.Roller
}

// IsPhaseTrigger reports whether trigger fires from a direct definition read at a
// phase event, not from a hook registered at cycle start. Turn start is a phase
// event, not a mutation; a purchased asset has no hooks until the next cycle.
func IsPhaseTrigger(trigger domain.EffectTrigger) bool {
	return trigger == domain.TriggerTurnStart || trigger == domain.TriggerOnPurchase
}

// effectsListHandler holds a definition's whole [[effects]] list, because the
// engine keys handlers by definition and one definition may declare several effects.
type effectsListHandler struct {
	def        *domain.AssetDefinition
	registrars map[domain.SpecialEffectType]Registrar
}

func (handler *effectsListHandler) AssetDefinitionID() string { return handler.def.ID }

func (handler *effectsListHandler) Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry) {
	for _, entry := range handler.def.Effects {
		if IsPhaseTrigger(entry.Trigger) {
			continue
		}
		// deliberate: no nil check on Register, because RegisterEffectsList rejects a hook-route entry without one.
		handler.registrars[entry.Effect].Register(RegisterContext{
			Faction:  faction,
			Asset:    asset,
			Def:      handler.def,
			Entry:    entry,
			Registry: hookRegistry,
		})
	}
}

// validateEntry checks one [[effects]] entry against its registrar (loud-failure
// rule 4).
func (e *EffectsEngine) validateEntry(entry domain.EffectDefinition) error {
	registrar, ok := e.registrars[entry.Effect]
	if !ok {
		return fmt.Errorf("effect %q has no registrar", entry.Effect)
	}
	if !slices.Contains(registrar.Triggers, entry.Trigger) {
		return fmt.Errorf("effect %q does not accept trigger %q", entry.Effect, entry.Trigger)
	}
	if IsPhaseTrigger(entry.Trigger) && registrar.Fire == nil {
		return fmt.Errorf("effect %q has no phase route for trigger %q", entry.Effect, entry.Trigger)
	}
	if !IsPhaseTrigger(entry.Trigger) && registrar.Register == nil {
		return fmt.Errorf("effect %q has no hook route for trigger %q", entry.Effect, entry.Trigger)
	}
	if !slices.Contains(registrar.Owners, entry.Filters.Owner) {
		return fmt.Errorf("effect %q does not accept owner %q", entry.Effect, entry.Filters.Owner)
	}
	if !slices.Contains(registrar.Scopes, entry.Filters.Scope) {
		return fmt.Errorf("effect %q does not accept scope %q", entry.Effect, entry.Filters.Scope)
	}
	for _, filter := range entry.Filters.Used() {
		if !slices.Contains(registrar.Filters, filter) {
			return fmt.Errorf("effect %q does not accept filter %q", entry.Effect, filter)
		}
	}
	for _, gate := range entry.Gates.Used() {
		if !slices.Contains(registrar.Gates, gate) {
			return fmt.Errorf("effect %q does not accept gate %q", entry.Effect, gate)
		}
	}
	return nil
}
```

##### Task 2 — `internal/faction/engine/effect/effect.go` (rewritten)

> Changed by the interim-review fix-up (Decision Record 13); the source is current.

Full contents:

```go
package effect

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Handler interface {
	AssetDefinitionID() string
	Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry)
}

type EffectsEngine struct {
	handlers   map[string]Handler
	registrars map[domain.SpecialEffectType]Registrar
	log        *slog.Logger
}

func New(registrars map[domain.SpecialEffectType]Registrar, log *slog.Logger) *EffectsEngine {
	return &EffectsEngine{handlers: make(map[string]Handler), registrars: registrars, log: log}
}

// Register binds a handler to its asset definition. A second handler for the same
// definition is an error, not an overwrite: a definition carrying both a transport
// profile and an effects list would otherwise lose one silently.
func (e *EffectsEngine) Register(handler Handler) error {
	id := handler.AssetDefinitionID()
	if _, exists := e.handlers[id]; exists {
		return fmt.Errorf("effect handler for asset %q already registered", id)
	}
	e.handlers[id] = handler
	return nil
}

// RegisterEffectsList checks every entry in a definition's [[effects]] list against
// its registrar, then registers the composed handler for the list. Checking here,
// not in a separate pass, means no list can be registered unchecked.
func (e *EffectsEngine) RegisterEffectsList(def *domain.AssetDefinition) error {
	for i, entry := range def.Effects {
		if err := e.validateEntry(entry); err != nil {
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
	}
	return e.Register(&effectsListHandler{def: def, registrars: e.registrars})
}

// ApplyAll walks every asset across all factions — factions by ID, then assets by
// ID, so hook registration order is the same every run — looks up the handler for
// that asset's definition ID, and invokes it. Assets whose definitions are missing
// or whose def has no registered handler are silently skipped.
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, log *slog.Logger) {
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(faction) {
			def := rulebook.Assets[asset.DefinitionID]
			if def == nil {
				continue
			}
			handler, ok := e.handlers[def.ID]
			if !ok {
				continue
			}
			handler.Apply(faction, asset, hookRegistry)
		}
	}
}
```

##### Task 3 — `internal/faction/engine/effect/effects/registrars.go` (new)

> Changed by the interim-review fix-up (Decision Record 13); the source is current.

Full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// Registrars returns the S-side registrar table, keyed on effect name. A fresh map
// per call, so no caller can mutate another's table.
func Registrars() map[domain.SpecialEffectType]effect.Registrar {
	return map[domain.SpecialEffectType]effect.Registrar{}
}
```

##### Task 4 — `internal/faction/engine/effect/effects/registrars_test.go` (new)

> Changed by the interim-review fix-up (Decision Record 10); the source is current.

Guards the table itself, independent of any data. Vacuous until Commit 4 adds an entry.

Full contents:

```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)

func TestRegistrars_DeclareARouteForEveryTrigger(t *testing.T) {
	for name, registrar := range effects.Registrars() {
		if len(registrar.Triggers) == 0 || len(registrar.Owners) == 0 || len(registrar.Scopes) == 0 {
			t.Errorf("%s: a registrar must accept at least one trigger, owner and scope", name)
		}
		for _, trigger := range registrar.Triggers {
			if effect.IsPhaseTrigger(trigger) && registrar.Fire == nil {
				t.Errorf("%s: phase trigger %q needs Fire", name, trigger)
			}
			if !effect.IsPhaseTrigger(trigger) && registrar.Register == nil {
				t.Errorf("%s: hook trigger %q needs Register", name, trigger)
			}
		}
	}
}
```

##### Task 5 — `internal/faction/engine/core.go`

**Find:**
```go
import (
	"errors"
	"log/slog"
```
**Replace with:**
```go
import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
```

**Find:**
```go
func NewWithRulebook(rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine, log *slog.Logger) *Engine {
	e := &Engine{Rulebook: rulebook, Rand: NewRandRoller(log), Hooks: hooks.NewRegistry(), World: worldEngine, log: log}
	e.Turn = turn.New(e.Rand, e.Rulebook, log)
	e.Mutation = mutation.New(log)
	e.Action = action.New(log)
	e.Goal = goal.New(log)
	e.Tag = tag.New(log)
	e.Effect = effect.New(log)
	for _, def := range rulebook.Assets {
		if def.Transport != nil {
			e.Effect.Register(effects.NewTransportHandler(def))
		}
	}
	actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook)
	return e
}
```
**Replace with:**
```go
func NewWithRulebook(rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine, log *slog.Logger) (*Engine, error) {
	e := &Engine{Rulebook: rulebook, Rand: NewRandRoller(log), Hooks: hooks.NewRegistry(), World: worldEngine, log: log}
	e.Turn = turn.New(e.Rand, e.Rulebook, log)
	e.Mutation = mutation.New(log)
	e.Action = action.New(log)
	e.Goal = goal.New(log)
	e.Tag = tag.New(log)
	e.Effect = effect.New(effects.Registrars(), log)
	for _, defID := range slices.Sorted(maps.Keys(rulebook.Assets)) {
		def := rulebook.Assets[defID]
		if def.Transport != nil {
			if err := e.Effect.Register(effects.NewTransportHandler(def)); err != nil {
				return nil, fmt.Errorf("engine: %w", err)
			}
		}
		if len(def.Effects) > 0 {
			if err := e.Effect.RegisterEffectsList(def); err != nil {
				return nil, fmt.Errorf("engine: %w", err)
			}
		}
	}
	actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook)
	return e, nil
}
```

##### Task 6 — `cmd/gm-toolkit/faction.go`

**Find:**
```go
	eng := engine.NewWithRulebook(rb, world.NewWithMap(spatialMap, log), log)
```
**Replace with:**
```go
	eng, err := engine.NewWithRulebook(rb, world.NewWithMap(spatialMap, log), log)
	if err != nil {
		return fmt.Errorf("faction: %w", err)
	}
```

##### Task 7 — `internal/faction/engine/testharness/harness.go`

**Find:**
```go
	eng := engine.NewWithRulebook(rb, world.NewWithMap(spatialMap, log), log)
```
**Replace with:**
```go
	eng, err := engine.NewWithRulebook(rb, world.NewWithMap(spatialMap, log), log)
	if err != nil {
		t.Fatalf("engine.NewWithRulebook: %v", err)
	}
```

##### Task 8 — `internal/faction/engine/testharness/scenarios/spot_check_test.go`

**Find:**
```go
		eng := engine.NewWithRulebook(rb, world.NewWithMap(&testharness.StubSpatialMap{}, log), log)
```
**Replace with:**
```go
		eng, err := engine.NewWithRulebook(rb, world.NewWithMap(&testharness.StubSpatialMap{}, log), log)
		if err != nil {
			t.Fatalf("engine.NewWithRulebook: %v", err)
		}
```

##### Task 9 — `internal/faction/tui/dryrun.go`

**Find:**
```go
	eng := engine.NewWithRulebook(rb, adapter.NewSmokeWorldEngine(log), log)
```
**Replace with:**
```go
	eng, err := engine.NewWithRulebook(rb, adapter.NewSmokeWorldEngine(log), log)
	if err != nil {
		return fmt.Errorf("dryrun: %w", err)
	}
```

##### Task 10 — `internal/faction/engine/effect/effect_test.go` (rewritten)

> Changed by the interim-review fix-up (Decision Record 10); the source is current.

The transport test moves to the new `New` and `Register` signatures; new tests cover the
duplicate guard, sorted `ApplyAll`, the composed handler and rule 4.

Full contents:

```go
package effect_test

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var discardLog = slog.New(slog.DiscardHandler)

var (
	ownSelf    = domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf}
	rivalWorld = domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld}
)

// TestEffectsEngine_RegistersTransportReactor verifies that ApplyAll registers
// exactly one MutationReactor at FactionScope for a faction whose assets include
// a transport-capable asset, and zero for a faction with no transport assets.
func TestEffectsEngine_RegistersTransportReactor(t *testing.T) {
	transportDefID := "test-transport"
	nonTransportDefID := "test-soldier"

	transportDef := &domain.AssetDefinition{
		ID:    transportDefID,
		Speed: 2,
		Transport: &domain.TransportProfile{
			MaxHex:     3,
			CoinCost:   1,
			CargoTypes: []domain.AssetType{domain.TypeSpecialForces},
			MaxCargo:   1,
		},
	}
	nonTransportDef := &domain.AssetDefinition{
		ID:   nonTransportDefID,
		Type: domain.TypeSpecialForces,
	}

	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			transportDefID:    transportDef,
			nonTransportDefID: nonTransportDef,
		},
	}

	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"alpha": {
				ID: "alpha",
				Assets: map[string]*domain.Asset{
					"transport-1": {ID: "transport-1", DefinitionID: transportDefID, OwnerID: "alpha"},
				},
			},
			"beta": {
				ID: "beta",
				Assets: map[string]*domain.Asset{
					"soldier-1": {ID: "soldier-1", DefinitionID: nonTransportDefID, OwnerID: "beta"},
				},
			},
		},
	}

	e := effect.New(nil, discardLog)
	if err := e.Register(effects.NewTransportHandler(transportDef)); err != nil {
		t.Fatalf("Register: %v", err)
	}

	registry := hooks.NewRegistry()
	e.ApplyAll(factionState, rb, registry, discardLog)

	wantSource := "transport:" + transportDefID

	// Faction alpha has a transport asset — one reactor should be registered.
	alphaReactors := registry.MutationReactorsFor("alpha", "")
	if len(alphaReactors) != 1 {
		t.Fatalf("alpha: got %d MutationReactors, want 1", len(alphaReactors))
	}
	if got := alphaReactors[0].Source; got != wantSource {
		t.Errorf("alpha reactor source = %q, want %q", got, wantSource)
	}

	// Faction beta has no transport asset — no reactor should be registered.
	betaReactors := registry.MutationReactorsFor("beta", "")
	if len(betaReactors) != 0 {
		t.Errorf("beta: got %d MutationReactors, want 0", len(betaReactors))
	}
}

func TestRegister_RejectsASecondHandlerForOneDefinition(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001", Transport: &domain.TransportProfile{MaxHex: 1}}
	e := effect.New(nil, discardLog)
	if err := e.Register(effects.NewTransportHandler(def)); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := e.Register(effects.NewTransportHandler(def))
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("second Register: got %v, want an already-registered error", err)
	}
}

// recordingHandler records the faction/asset pairs ApplyAll hands it.
type recordingHandler struct {
	defID string
	calls *[]string
}

func (handler recordingHandler) AssetDefinitionID() string { return handler.defID }

func (handler recordingHandler) Apply(faction *domain.Faction, asset *domain.Asset, _ *hooks.Registry) {
	*handler.calls = append(*handler.calls, faction.ID+"/"+asset.ID)
}

func TestApplyAll_WalksFactionsAndAssetsInIDOrder(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001"}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{}}
	for _, factionID := range []string{"delta", "alpha", "charlie", "bravo"} {
		faction := &domain.Faction{ID: factionID, Assets: map[string]*domain.Asset{}}
		for _, suffix := range []string{"3", "1", "2"} {
			assetID := factionID + "-" + suffix
			faction.Assets[assetID] = &domain.Asset{ID: assetID, DefinitionID: def.ID, OwnerID: factionID}
		}
		factionState.Factions[factionID] = faction
	}

	var calls []string
	e := effect.New(nil, discardLog)
	if err := e.Register(recordingHandler{defID: def.ID, calls: &calls}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	e.ApplyAll(factionState, rb, hooks.NewRegistry(), discardLog)

	want := []string{
		"alpha/alpha-1", "alpha/alpha-2", "alpha/alpha-3",
		"bravo/bravo-1", "bravo/bravo-2", "bravo/bravo-3",
		"charlie/charlie-1", "charlie/charlie-2", "charlie/charlie-3",
		"delta/delta-1", "delta/delta-2", "delta/delta-3",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("ApplyAll order:\n got  %v\n want %v", calls, want)
	}
}

func TestRegisterEffectsList_RegistersHookRouteEntriesOnly(t *testing.T) {
	var registered []domain.SpecialEffectType
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: []domain.EffectTrigger{domain.TriggerTurnStart},
			Owners:   []domain.EffectOwner{domain.OwnerOwn},
			Scopes:   []domain.EffectScope{domain.ScopeSelf},
			Fire: func(effect.FireContext) []domain.Mutation {
				t.Error("ApplyAll must not fire a phase-route entry")
				return nil
			},
		},
		domain.SpecialCoinSteal: {
			Triggers: []domain.EffectTrigger{domain.TriggerOwnAttackHit},
			Owners:   []domain.EffectOwner{domain.OwnerRival},
			Scopes:   []domain.EffectScope{domain.ScopeWorld},
			Register: func(ctx effect.RegisterContext) {
				if ctx.Faction.ID != "alpha" || ctx.Asset.ID != "alpha-1" || ctx.Def.ID != "T-001" || ctx.Registry == nil {
					t.Errorf("RegisterContext = %+v", ctx)
				}
				registered = append(registered, ctx.Entry.Effect)
			},
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinSteal, Filters: rivalWorld},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{
			"alpha-1": {ID: "alpha-1", DefinitionID: def.ID, OwnerID: "alpha"},
		}},
	}}

	e := effect.New(registrars, discardLog)
	if err := e.RegisterEffectsList(def); err != nil {
		t.Fatalf("RegisterEffectsList: %v", err)
	}
	e.ApplyAll(factionState, rb, hooks.NewRegistry(), discardLog)

	if want := []domain.SpecialEffectType{domain.SpecialCoinSteal}; !reflect.DeepEqual(registered, want) {
		t.Errorf("registered = %v, want %v", registered, want)
	}
}

func noopFire(effect.FireContext) []domain.Mutation { return nil }

func noopRegister(effect.RegisterContext) {}

func validationRegistrars() map[domain.SpecialEffectType]effect.Registrar {
	return map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: []domain.EffectTrigger{domain.TriggerTurnStart},
			Owners:   []domain.EffectOwner{domain.OwnerOwn},
			Scopes:   []domain.EffectScope{domain.ScopeSelf},
			Fire:     noopFire,
		},
		domain.SpecialCoinSteal: {
			Triggers: []domain.EffectTrigger{domain.TriggerOwnAttackHit},
			Owners:   []domain.EffectOwner{domain.OwnerRival},
			Scopes:   []domain.EffectScope{domain.ScopeWorld},
			Filters:  []domain.EffectFilterName{domain.FilterStealthy},
			Gates:    []domain.EffectGateName{domain.GateOncePer},
			Register: noopRegister,
		},
		// declares a hook trigger but has only a phase route
		domain.SpecialClearStealth: {
			Triggers: []domain.EffectTrigger{domain.TriggerStealthArrival},
			Owners:   []domain.EffectOwner{domain.OwnerRival},
			Scopes:   []domain.EffectScope{domain.ScopeWorld},
			Fire:     noopFire,
		},
		// declares a phase trigger but has only a hook route
		domain.SpecialApplyStealth: {
			Triggers: []domain.EffectTrigger{domain.TriggerOnPurchase},
			Owners:   []domain.EffectOwner{domain.OwnerOwn},
			Scopes:   []domain.EffectScope{domain.ScopeSelf},
			Register: noopRegister,
		},
	}
}

func TestRegisterEffectsList_Validates(t *testing.T) {
	stealthy := true
	cases := []struct {
		name    string
		entry   domain.EffectDefinition
		wantErr string
	}{
		{
			name:  "accepted phase-route entry",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		},
		{
			name: "accepted hook-route entry with a filter and a gate",
			entry: domain.EffectDefinition{
				Trigger: domain.TriggerOwnAttackHit,
				Effect:  domain.SpecialCoinSteal,
				Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld, Stealthy: &stealthy},
				Gates:   domain.EffectGates{OncePer: domain.OncePerTargetFaction},
			},
		},
		{
			name:    "no registrar",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialReroll, Filters: ownSelf},
			wantErr: `effect "reroll" has no registrar`,
		},
		{
			name:    "trigger not accepted",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinGain, Filters: ownSelf},
			wantErr: `does not accept trigger "own_attack_hit"`,
		},
		{
			name: "owner not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf}},
			wantErr: `does not accept owner "rival"`,
		},
		{
			name: "scope not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld}},
			wantErr: `does not accept scope "world"`,
		},
		{
			name: "filter not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf, Attribute: domain.StatCunning}},
			wantErr: `does not accept filter "attribute"`,
		},
		{
			name: "gate not accepted",
			entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain,
				Filters: ownSelf, Gates: domain.EffectGates{NoStack: true}},
			wantErr: `does not accept gate "no_stack"`,
		},
		{
			name:    "hook trigger without a hook route",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerStealthArrival, Effect: domain.SpecialClearStealth, Filters: rivalWorld},
			wantErr: "has no hook route",
		},
		{
			name:    "phase trigger without a phase route",
			entry:   domain.EffectDefinition{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: ownSelf},
			wantErr: "has no phase route",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{tc.entry}}
			err := effect.New(validationRegistrars(), discardLog).RegisterEffectsList(def)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("RegisterEffectsList: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("RegisterEffectsList: got %v, want an error containing %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), `asset "T-001" effects[0]`) {
				t.Errorf("error %q does not name the entry", err.Error())
			}
		})
	}
}
```

##### Task 11 — `internal/faction/engine/core_test.go` (new)

Proves construction runs rule 4. `coin_gain` never accepts `own_roll`, so this fails
both before Commit 4 (no registrar) and after it (trigger not accepted).

Full contents:

```go
package engine

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func TestNewWithRulebook_ValidatesEffects(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialCoinGain,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
	}}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), `asset "T-001" effects[0]`) {
		t.Fatalf("got %v, want a validation error naming the entry", err)
	}
}
```

##### Task 12 — `internal/faction/engine/testharness/scenarios/transport_lifecycle_test.go`

The one direct `Register` caller outside `core.go`. It still compiles when `Register`
grows its error return, so the compiler does not flag it.

**Find:**
```go
	h.Engine.Effect.Register(effects.NewTransportHandler(transportDef))
```
**Replace with:**
```go
	if err := h.Engine.Effect.Register(effects.NewTransportHandler(transportDef)); err != nil {
		t.Fatalf("Register: %v", err)
	}
```

##### Commit message

```
feat: effect registrar contract with construction-time validation

- Registrar spec struct with accepted sets and hook/phase routes
- composed per-definition handler for [[effects]] lists
- rule 4 at engine construction: RegisterEffectsList checks registrar,
  trigger, owner, scope, filters, gates and route before it registers
- Register rejects a second handler for one definition
- NewWithRulebook returns an error and registers in definition ID
  order; four callers updated
- ApplyAll walks factions and assets in ID order
```

#### Commit 4 — `feat: turn-start effect route and coin_gain`

The phase route, its first registrar, Party Machine's data, and the catalog
reconciliation.

##### Task 1 — `internal/faction/engine/effect/effect.go`

> Changed by the interim-review fix-up (Decision Record 12); the source is current.

Append after `ApplyAll` (end of file, as written in Commit 3 Task 2):

```go
// FireTurnStart runs every turn_start entry on faction's active assets, in asset ID
// order, through its registrar's phase route. A direct definition read, not a hook:
// turn start is a phase event the orchestrator raises once per turn.
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		if !asset.Ready || !asset.Maintained {
			continue
		}
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil {
			continue
		}
		for _, entry := range def.Effects {
			if entry.Trigger != domain.TriggerTurnStart {
				continue
			}
			// deliberate: no nil check on Fire, because RegisterEffectsList rejects a phase-route entry without one.
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Faction: faction,
				Asset:   asset,
				Def:     def,
				Entry:   entry,
				Roller:  roller,
			})...)
		}
	}
	return mutations
}
```

##### Task 2 — `internal/faction/engine/effect/effects/coin_gain.go` (new)

> Changed by the interim-review fix-up (Decision Records 10 and 13); the source is current.

Full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// coinGainRegistrar pays the owning faction a fixed or rolled amount at the start
// of its turn (Party Machine).
var coinGainRegistrar = effect.Registrar{
	Triggers: []domain.EffectTrigger{domain.TriggerTurnStart},
	Owners:   []domain.EffectOwner{domain.OwnerOwn},
	Scopes:   []domain.EffectScope{domain.ScopeSelf},
	Fire:     fireCoinGain,
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             ctx.Entry.CoinGain.Amount.Roll(ctx.Roller),
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Faction.ID,
	}}
}
```

##### Task 3 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
	return map[domain.SpecialEffectType]effect.Registrar{}
```
**Replace with:**
```go
	return map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: coinGainRegistrar,
	}
```

##### Task 4 — `internal/faction/engine/orchestrator.go`

In `runBookkeepingPhase`, fire turn-start effects before income and maintenance
(Decision 5).

**Find:**
```go
	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState, e.Hooks, phaseLog.With("engine", "turn"))
```
**Replace with:**
```go
	var startMutations []domain.Mutation
	// deliberate: fire here, gated on BookkeepingApplied, not in turn setup, because setup re-runs on every resume and would pay a turn-start effect twice.
	if e.Turn.InProgress(factionState) && !factionState.CurrentTurn.BookkeepingApplied {
		startMutations = e.Effect.FireTurnStart(faction, e.Rulebook, e.Rand)
		if err := e.applyAndRecord(factionState, faction, startMutations, paths, phaseLog); err != nil {
			if !factionErrors.IsRecoverable(err) {
				phaseLog.Error("failure", "err", err)
				observer.OnError(faction, err)
				return err
			}
			// deliberate: fall through to bookkeeping, not return, because returning here would skip this turn's income and maintenance.
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			startMutations = nil
		}
	}

	bookResult, bookMutations, err := e.Turn.ApplyBookkeeping(factionState, e.Hooks, phaseLog.With("engine", "turn"))
```

**Find:**
```go
	observer.OnBookkeepingApplied(faction, bookResult, bookMutations)
```
**Replace with:**
```go
	observer.OnBookkeepingApplied(faction, bookResult, append(startMutations, bookMutations...))
```

##### Task 5 — `rulebooks/swn/assets/cunning_assets.toml` and `campaigns/test-camp/rulebook/assets/cunning_assets.toml`

Apply the same edit to both files.

**Find:**
```toml
[assets.PartyMachine.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"
damage = "2d6"

[assets.VanguardCadres]
```
**Replace with:**
```toml
[assets.PartyMachine.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"
damage = "2d6"

[[assets.PartyMachine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.PartyMachine.effects.filters]
owner = "own"
scope = "self"
[assets.PartyMachine.effects.coin_gain]
amount = "1"

[assets.VanguardCadres]
```

##### Task 6 — `internal/faction/engine/testharness/harness.go`

**Find:**
```go
	DefSecurityPersonnel = "SWN-F1-001"
	DefHeavyDropAssets   = "SWN-F2-001"
```
**Replace with:**
```go
	DefSecurityPersonnel = "SWN-F1-001"
	DefHeavyDropAssets   = "SWN-F2-001"
	DefPartyMachine      = "SWN-C4-001"
```

##### Task 7 — `internal/faction/engine/effect/effect_test.go`

> Changed by the interim-review fix-up (Decision Record 10); the source is current.

Append to the end of the file:

```go
func TestFireTurnStart_FiresActiveTurnStartEntriesInAssetOrder(t *testing.T) {
	registrars := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialCoinGain: {
			Triggers: []domain.EffectTrigger{domain.TriggerTurnStart},
			Owners:   []domain.EffectOwner{domain.OwnerOwn},
			Scopes:   []domain.EffectScope{domain.ScopeSelf},
			Fire: func(ctx effect.FireContext) []domain.Mutation {
				return []domain.Mutation{domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: 1, Cause: ctx.Asset.ID}}
			},
		},
	}
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{
		{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialCoinGain, Filters: ownSelf},
		{Trigger: domain.TriggerOwnAttackHit, Effect: domain.SpecialCoinSteal, Filters: rivalWorld},
	}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{
		"a-2":           {ID: "a-2", DefinitionID: def.ID, Ready: true, Maintained: true},
		"a-1":           {ID: "a-1", DefinitionID: def.ID, Ready: true, Maintained: true},
		"a-not-ready":   {ID: "a-not-ready", DefinitionID: def.ID, Ready: false, Maintained: true},
		"a-unmaintained": {ID: "a-unmaintained", DefinitionID: def.ID, Ready: true, Maintained: false},
	}}

	mutations := effect.New(registrars, discardLog).FireTurnStart(faction, rb, nil)

	var causes []string
	for _, mutation := range mutations {
		causes = append(causes, mutation.(domain.CoinDelta).Cause)
	}
	if want := []string{"a-1", "a-2"}; !reflect.DeepEqual(causes, want) {
		t.Errorf("fired for %v, want %v", causes, want)
	}
}
```

##### Task 8 — `internal/faction/engine/effect/effects/coin_gain_test.go` (new)

Full contents:

```go
package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)

type fixedRoller struct{ value int }

func (roller fixedRoller) Roll(int) int { return roller.value }

func TestCoinGain_PaysTheOwnerTheAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount domain.DiceRoll
		want   int
	}{
		{"flat amount", domain.DiceRoll{Modifier: 1}, 1},
		{"rolled amount", domain.DiceRoll{NumDice: 1, Sides: 4}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "alpha"}
			entry := domain.EffectDefinition{
				Trigger:  domain.TriggerTurnStart,
				Effect:   domain.SpecialCoinGain,
				Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
				CoinGain: &domain.CoinGainParams{Amount: tc.amount},
			}
			got := effects.Registrars()[domain.SpecialCoinGain].Fire(effect.FireContext{
				Faction: faction,
				Entry:   entry,
				Roller:  fixedRoller{value: 3},
			})
			want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: tc.want, Cause: "coin_gain", CausedByFactionID: "alpha"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Fire = %+v, want %+v", got, want)
			}
		})
	}
}
```

##### Task 9 — `internal/faction/engine/core_test.go`

The collision guard, reachable now that `coin_gain` validates.

Append to the end of the file:

```go
func TestNewWithRulebook_RejectsTransportAndEffectsOnOneDefinition(t *testing.T) {
	def := &domain.AssetDefinition{
		ID:        "T-001",
		Transport: &domain.TransportProfile{MaxHex: 1},
		Effects: []domain.EffectDefinition{{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		}},
	}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("got %v, want an already-registered error", err)
	}
}
```

##### Task 10 — `internal/faction/engine/testharness/scenarios/effects_test.go` (new)

End to end against the real SWN data: Party Machine pays before maintenance, and a
resumed turn does not pay twice.

Full contents:

```go
package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func addPartyMachine(faction *domain.Faction) {
	assetID := faction.ID + "-party-machine"
	faction.Assets[assetID] = &domain.Asset{
		ID:           assetID,
		DefinitionID: testharness.DefPartyMachine,
		OwnerID:      faction.ID,
		Location:     faction.Homeworld,
		CurrentHP:    10,
		Ready:        true,
		Maintained:   true,
	}
}

func skipAction(*domain.Faction, []action.Action) (action.Action, error) { return nil, nil }

func TestTurnStartEffect_PartyMachinePaysAtTurnStart(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	addPartyMachine(alpha)
	h.Collector.SelectActionFn = skipAction
	initialCoin := alpha.Coin

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	// income: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 = 2; Party Machine: +1
	testharness.CheckStep(t, "alpha Coin includes the Party Machine payout", alpha.Coin == initialCoin+3,
		"alpha.Coin should be bookkeeping income plus 1")

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "coin_delta", "coin_gain")
	testharness.CheckStep(t, "coin_gain recorded in history", found, "no coin_delta with cause coin_gain")
}

func TestTurnStartEffect_ResumedTurnDoesNotPayTwice(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	addPartyMachine(alpha)
	h.Collector.SelectActionFn = skipAction
	initialCoin := alpha.Coin

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.BookkeepingApplied = true
	if _, err := h.Engine.RunFactionTurn(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunFactionTurn: %v", err)
	}

	testharness.CheckStep(t, "no payout on a turn whose bookkeeping already ran", alpha.Coin == initialCoin,
		"alpha.Coin changed on a resumed turn")
}
```

##### Task 11 — `internal/faction/rulebook/rulebook_test.go`

Insert before the `t.Run("goals populated"` subtest:

```go
	t.Run("party machine coin_gain effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C4-001"]
		if !ok {
			t.Fatal("asset SWN-C4-001 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

```

##### Task 12 — `docs/architecture/engine/ability-catalog.md`

The two Effort-1 reconciliation items (Decision 18). Every claim below was verified
against source in the Plan session; re-verify against the built code before committing.

**Find:**
```markdown
> [world & movement](world-movement.md) (relocate). The S-side and the remaining
> A-side effects (Marketers, Seditionists, Monopoly) are still future.
```
**Replace with:**
```markdown
> [world & movement](world-movement.md) (relocate). The remaining A-side effects
> (Marketers, Seditionists, Monopoly) are still future.
>
> **As-built (A.001.3 Effort 1):** the S-side model ships. An asset declares an
> `[[effects]]` list; the effects engine keys a registrar table on the effect name
> (`effect/registrar.go`, `effects/registrars.go`) and checks every entry against it
> at engine construction. Party Machine's `coin_gain` is the first S-side effect; the
> rest of the S-flag table below is still future.
```

**Find:**
```markdown
| `AbilityDefinition.Die *DiceRoll` | roll for an outcome-table ability | **dead** — `convertAbility` never reads it |
| `AbilityDefinition.Outcomes []AbilityOutcome{From,To,Coin}` | die-range → coin payout | **dead** — never parsed |
| `AbilityEffectType` = `reveal_stealth` | reveal effect | **wired** (Informers only) |
| `AbilityEffectType` = `coin_drain`, `coin_steal` | coin transfer effects | parse fine, **dispatched nowhere** |
| `[[ability.steps]]` in Seductress TOML | multi-step ability | **silently dropped** — `abilityRecord` has no `steps` field, so Seductress's reveal no-ops today |

The economic cluster (below) is the clearest payoff: `Die` + `Outcomes` already
models it, and only the parser and a handler are missing.
```
**Replace with:**
```markdown
| `RollTable.Die DiceRoll` | roll for an outcome-table ability | **wired** — parsed into `AbilityStep.Roll` |
| `RollTable.Outcomes []AbilityOutcome{From,To,Coin,DestroySelf}` | die-range → coin payout | **wired** — parsed into `AbilityStep.Roll` |
| `AbilityEffectType` = `reveal_stealth` | reveal effect | **wired** (Informers, Seductress) |
| `AbilityEffectType` = `coin_drain`, `coin_steal` | coin transfer effects | **removed** by A.001.2 — `AbilityEffectType` has five constants, none a coin transfer. The S-side `coin_steal` is a `SpecialEffectType` (`domain/effect.go`) |
| `[[ability.steps]]` in Seductress TOML | multi-step ability | **wired** — `abilityRecord.Steps` parses it (relocate, then reveal) |

A.001.2 built the parser path and handlers for every A-side row above; see
[actions](actions.md) for the as-built ability dispatch.
```

**Find:**
```markdown
The `S`-flag specials register [hooks](hooks.md) at cycle start through
`EffectsEngine.ApplyAll`. Each effect maps to one of the **five hook
categories**; the catalog's job on this side is to define the data shape that
```
**Replace with:**
```markdown
The `S`-flag specials register [hooks](hooks.md) at cycle start through
`EffectsEngine.ApplyAll`. Most effects map to one of the **five hook
categories**, numbered as in `hooks/doc.go` (Cat 4 is the `RuleModifier` family,
Cat 5 is `TieResolver`). Phase-event effects are the exception: a turn-start
effect fires from a direct read of the asset's definition at the start of the
bookkeeping phase, not from a registered hook. The catalog's job on this side is to define the data shape that
```

**Find:**
```markdown
| **passive income** (per-turn Coin) | Cat 3 `MutationReactor` (turn-start) | Party Machine `SWN-C4-001` |
```
**Replace with:**
```markdown
| **passive income** (per-turn Coin) | turn-start phase route (direct definition read, not a hook) | Party Machine `SWN-C4-001` ✅ |
```

**Find:**
```markdown
| **start-stealthed** | Cat 5 `RuleModifier` (on add) | Psychic Assassins `SWN-F5-003` |
| **grant-stealth** (quality bought for another asset) | Cat 5 `RuleModifier` | Stealth `SWN-C3-002` |
```
**Replace with:**
```markdown
| **start-stealthed** | Cat 4 `RuleModifier` (on add) | Psychic Assassins `SWN-F5-003` |
| **grant-stealth** (quality bought for another asset) | Cat 4 `RuleModifier` | Stealth `SWN-C3-002` |
```

##### Commit message

```
feat: turn-start effect route and coin_gain

- FireTurnStart runs turn_start entries on active assets, in asset order
- the bookkeeping phase fires them before income and maintenance, once
  per turn, guarded by BookkeepingApplied
- coin_gain registrar
- author Party Machine
- catalog: renumber hook-category labels, mark Party Machine built,
  bring the latent vocabulary table to as-built truth
```
