# Data-Driven Tags (A.001.7) — Effort 5 Plan: A-Side Tag Holder

## In Plain Words

- **What:** the last four tag rules start to work, and so does one rule from the
  assets.
  - **Tags can carry abilities.** The Use Asset Ability action lists a faction's
    tag abilities next to its A-flag assets. A tag can carry any ability that needs
    no acting asset.
  - **Preceptor Archive** gets Teach Planetary Population: pay 2 Coin, roll 1d12,
    and on a 12 a world where the faction has a Base counts as TL4 for that
    faction's purchases from then on.
  - **Eugenics Cult** can buy the new Gengineered Slaves asset, which no other
    faction can buy, refit into or start with. Its Slaves get an extra die once
    per turn.
  - **Informers and Seductress** roll their contests as attacks, as the book
    words them, so attack dice and tie rules (Fanatical) reach them.
- **Why:** these rules do nothing in play today. With them, all 20 SWN tags run
  from rulebook data.
- **How it fits:** Effort 5 of five in `A.001.7` (arc `A.001`), on branch
  `feature/data-driven-tags`. It builds on Efforts 1 and 2. The pre-merge
  checklist follows.

## Context / Goal

- Overview: [`A.001.7-data-driven-tags-overview-plan.md`](./A.001.7-data-driven-tags-overview-plan.md).
  Decision 3 (one check in `CanPurchase`) gets `requires_tag`. Decision 7 (values
  and references in the data, widened by R-6) binds every new effect, field and
  filter here.
- Discovery: [`A.001.7-data-driven-tags-discovery.md`](../discovery/A.001.7-data-driven-tags-discovery.md),
  D-1, D-11, D-13.
- Hand-off from Effort 3 (overview Findings, "Ability contest ties"): the book
  words Informers' and Seductress' contests as "Cunning vs. Cunning attack".
  Decision 10 settles it.
- Architecture read: ability-catalog, actions, persistence.
- The discovery's three open questions were settled in the overview plan
  (Decisions 2, 3, 4). None reaches this effort.

## Decisions Ratified in Planning

1. **Tag abilities run inside Use Asset Ability.** Robert's call. The picker lists
   the faction's A-flag assets in ID order, then its tags that carry an ability, in
   ID order. The GM's pick order sets the resolve order, as today. A new
   `domain.AbilityHolder` (`Asset` or `Tag`, exactly one set) replaces
   `*domain.Asset` in the picker (`SelectAbilityAssets` becomes
   `SelectAbilityHolders`), in `SelectAbilityTarget` and in `ability.Dispatch`.
   It lives in `domain`, not as `effect.Holder`, because the collector and the TUI
   name it and must not import the effects engine. Rejected: a separate menu
   action per tag ability (the book's "special action" reading; Robert chose one
   action).

2. **A tag can carry any ability effect that needs no acting asset.** Robert's call,
   as R-1 did for S effects. Those effects are `coin_delta`, `purchase_discount`
   and the new `raise_tech_level`. The tag loader refuses every other effect, a
   contest, a `self`, `faction` or `asset` target, and an outcome row with
   `destroy_self` or a negative `coin`, because each one reads the acting asset or
   its world. The check is `checkTagAbility` in `rulebook.go`, beside the other
   per-effect ability checks. Rejected: only `raise_tech_level` accepts a tag (the
   next tag ability would need code).

3. **`raise_tech_level` is a new ability effect, named for what it does.** Planning
   call (overview Decision 7, Effort 4 Decision 3). Target `world`, a die, outcome
   rows, and `[raise_tech_level] coin_cost` and `tech_level`. The handler always
   pays `coin_cost`. When the roll matches an outcome row, it records the world at
   `tech_level` for the acting faction. Outcome rows carry no `coin` or
   `destroy_self`. An asset may carry it too.

4. **Teach's targets are the Base worlds where it helps.** Robert's call (rules
   reading). These are worlds where the acting faction has a Base of Influence and
   that do not yet count as the entry's `tech_level` for every purchase: the raw TL,
   the faction's record and the tech-level floors that serve every category,
   through `ResolveWorldTechLevel` with a nil definition. A category-limited floor
   (Pretech Researchers, R&D Department) does not count. Revised at the effort
   close; see the Decision Record. A
   faction that cannot pay `coin_cost` has no target, so the tag is not offered.
   When an earlier holder in the same action spends the Coin, the step fizzles
   (`errNoTargets`), as Seditionists does. With no world engine, no TL is known,
   so there is no target.

5. **The record is `Faction.WorldTechLevels`, read first by
   `ResolveWorldTechLevel`.** Planning call (D-11). `map[string]int`,
   `toml:"world_tech_levels,omitempty"`, so old saves load. The mutation is
   `WorldTechLevelRaised`, applied as the higher of the old and new value.
   `ResolveWorldTechLevel` starts from `max(baseTL, record)` and then runs the
   hooks. Only Buy Asset reads it. Pretech Logistics has its own cap, and the
   wizard's starting picker stays on the raw TL (Effort 4 Decision 12).

6. **The first world target.** Planning call. `stepCandidates` gets a world case
   through a `worldTargetRules` table keyed on the effect. The TUI gets
   `AskSelectAbilityWorldTarget` and a single-select overlay with world names,
   built like the Seize Planet picker. The dormant fence-signs on `TargetWorld`
   and `AbilityTargetCandidates.Worlds` come off.

7. **`requires_tag` reaches four purchase paths.** Robert's call. It is an asset
   field, read through `AssetDefinition.TagRequirementMet(tagIDs)`. It is checked
   in `CanPurchase` (Buy Asset, Pretech Logistics), in `validRefitReplacements`,
   and in the wizard's starting picker against the picked tags. The loader checks
   that the named tag exists.

8. **Gengineered Slaves is `SWN-F1-004`, a Military Unit.** Robert's call. The
   other values come from the tag text: Force 1, 6 HP, 2 Coin, TL4, Attack Force
   vs Force 1d6, Counter 1d4. No flags.

9. **The Eugenics Cult die uses a new `asset_definition` roll filter.** Planning
   call (overview Decision 7). The filter names the rolling asset's definition ID
   and accepts no value from Go. `bonus_die` accepts it.
   `checkGoalReferences` becomes `checkReferences`, which also checks
   `asset_definition` filters and `requires_tag` against the loaded data. The
   entry: `own_roll`, `bonus_die` 1d10, phases `attack` and `defense`, scope
   `faction`, `once_per = "owner_faction"`.

10. **Ability contests name their roll: `attack` or `test`, required.** Robert's call
    (rules reading, the Effort 3 hand-off). Informers and Seductress roll `attack`,
    Marketers `test`. With `attack`, `passesGate` rolls attack and defense phases,
    and a tie asks `dispatch.ResolveTie`, as `contestPasses` does. It still does no
    damage and triggers no counterattack. `Roll` moves from `EffectContest` into
    the shared `domain.Contest`, so one field serves both. The key is required on
    every ability contest, as on effect contests, so a missing key cannot silently
    default to `test`.

11. **Three phases, five commits.** Planning call. Phase 1: the holder refactor,
    then tag abilities. Phase 2: Teach Planetary Population. Phase 3: Gengineered
    Slaves and the Eugenics die, then the ability contests. Settled in
    conversation as two phases; the written Phase 1 ran about 2,200 lines, past
    the 1,500-line budget, so Teach became its own phase. Each phase runs about
    1,000 to 1,100 lines.

## Decision Record — Execution

- **Commit 3, a Buy Asset test the plan missed.** Task 18 tested that the record is
  written and that `ResolveWorldTechLevel` reads it, but nothing tested a purchase
  on a taught world. `TestPurchasableDefinitions_ReadsTheWorldTechLevelRecord` in
  `buy_asset_test.go` now does: a TL6 def is offered on a TL5 world with a record of
  6, and not without it. It also pins the key. Buy Asset passes the world IDs from
  `availableWorlds`, the same IDs `raiseTechLevelWorlds` records. The test router is
  fixed at TL 5, so the test uses 6 rather than Teach's 4.
- **Commit 3, no re-compose.** Still no `campaigns/sandbox-claude/`, so the Preceptor
  Archive data reached no campaign rulebook.
- **Commit 5, Task 4 anchor drift.** Commit 3 added `RaiseTechLevel` to
  `abilityStepRecord`, which widened the field columns, so the `Effect`/`Target`
  Find anchor no longer matched as written. The `Roll` field goes in the same place,
  after `Target`, at the new alignment.
- **Commit 5, Robert's sandbox no longer loads until re-composed.** `roll` is now
  required on every ability contest, and the git-tracked rulebook copy in
  `campaigns/sandbox/rulebook/` predates it. A load probe fails on
  `SWN-C2-004 ability: step 1: unknown contest roll: ""`. The plan's stance stands:
  the sandbox is Robert's and takes the new data when he re-composes it (D-4). Until
  then `--campaign sandbox` refuses to start. Commits 3 and 4 only add data, so they
  did not break the load.
- **Effort close, Decision 4 reversed: Teach's targets read the floors.** Robert's
  call. Before, a world was skipped when its raw TL reached the entry's `tech_level`
  or when it had any record at all. So Teach was offered on worlds that Technical
  Expertise, Colonists or Laboratory already lift to TL4, and the faction paid 2
  Coin for nothing. A world with a lower record could also never be raised. Now
  `raiseTechLevelWorlds` asks `ResolveWorldTechLevel` with a nil definition and skips
  the world when the answer reaches `tech_level`. A nil definition means "every
  purchase", so `techLevelFloorModifier` ignores a category-limited floor for it.
  Teach raises the world for every category, so a world that only Pretech
  Researchers covers is still worth teaching. `HasTargets`, `stepCandidates` and
  `resolveTarget` now take the hook registry.
- **Effort close, the handler's Coin check is an error.** The review said a faction
  left short by an earlier holder rolls the die and fizzles in silence. It does not:
  the target filter reads the view's Coin, so that case fizzles as `errNoTargets`
  before any roll (Decision 4). The handler's own `return nil, nil` could not be
  reached, so it is now an error that says the filter and the handler disagree.
- **Effort close, a clearer missing-`roll` error.** A contest with no `roll`, on an
  ability or an effect, now fails with `a contest needs roll (want one of [attack
  test])`, not `unknown contest roll: ""`. There is still no default, because a
  default to `test` would silently misroll an attack. The sandbox still needs a
  re-compose.
- **Effort close, stricter `raise_tech_level` load checks.** A `raise_tech_level`
  step needs at least one outcome row. A `[raise_tech_level]` block under another
  effect fails to load and is no longer silently ignored.
- **Initiative close: a picked world must be a candidate.** From the initiative
  review. The `TargetWorld` branch of `resolveTarget` took any world the collector
  returned, so a bad pick could pay 2 Coin and record TL4 on a world the filter
  (Decision 4) left out. It now refuses a world outside `candidates.Worlds`. Test:
  `TestResolveTarget_World`.
- **Initiative close: `Dispatch` reads the asset's definition itself.** From the
  initiative review. `Dispatch` took the holder and a separate `def` that callers
  had to keep in step with `holder.Asset`. It now reads the definition from the
  rulebook (`holderDefinition`; nil for a tag), and Use Asset Ability no longer
  looks it up. The task listings above keep the `def` parameter as executed.

---

## Shared Context

### Branch, cadence, and checks

- Branch `feature/data-driven-tags`. One commit per plan commit. At the effort
  boundary: `/gestalt:effort-close A.001.7.E5`, then the pre-merge checklist.
- Every commit builds and passes `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file.
- Edit files only with Edit/Write. `mockgen` is not installed, so Commit 1 edits
  `engine/action/actions/mocks/mock_collector.go` by hand, in the generated
  shape.
- A test-file edit given as "replace every" is an Edit with `replace_all: true`.
- Scenario tests set `h.Engine.Rand` to a `FixedRoller`. If the cycle rolls more
  dice than the test lists, the roller panics. Add the missing values in front and
  log the change in the Decision Record.
- No saved-state migration: `world_tech_levels` is `omitempty`. No
  `campaigns/sandbox-claude/` exists. Robert's `sandbox` takes the new tag data
  when he re-composes it (D-4).

### Pre-merge additions (beyond the overview's and Effort 4's)

- **Graduate to Key Decisions (actions):** tag abilities run inside Use Asset
  Ability through `domain.AbilityHolder`; a tag carries only effects that need no
  acting asset (Decisions 1, 2). Ability contests name `attack` or `test`, and an
  attack contest asks the tie resolvers (Decision 10).
- **Graduate to Key Decisions (persistence):** the per-faction world TL record
  (Decision 5); `requires_tag` and the `asset_definition` reference checked at
  load (Decisions 7, 9).
- **Ability catalog:** an as-built note for the first tag ability and the first
  world target.
- **`planned-work.md`:** a Backlog row: Refit Asset checks neither the world's TL
  nor the population floor, so a refit can reach an asset Buy Asset would refuse.
  This was found in planning; Effort 5 adds only `requires_tag` to Refit.

## Out of Scope

- Refit's missing TL and population checks (the Backlog row above).
- Teach's record in the wizard's starting picker (Effort 4 Decision 12).
- Tag abilities that hold a contest or target a faction or an asset (Decision 2).
- A display name for a tag ability. The picker shows the tag's name and the
  effect.

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Tag abilities**
  - Commit 1 — `refactor(E5): abilities run through a holder` (12 tasks)
  - Commit 2 — `feat(E5): tags carry abilities` (8 tasks)
- [x] **Phase 2 — Teach Planetary Population**
  - Commit 3 — `feat(E5): Teach Planetary Population` (18 tasks)
- [x] **Phase 3 — Gengineered Slaves and ability contests**
  - Commit 4 — `feat(E5): Gengineered Slaves and requires_tag` (14 tasks)
  - Commit 5 — `feat(E5): ability contests roll as the book words them` (9 tasks)

### Phase 1 — Tag abilities

**In plain words:** Use Asset Ability lists a faction's tags that carry an ability
beside its A-flag assets, and runs a tag's ability with no asset. No SWN tag
carries one yet.

#### Commit 1 — `refactor(E5): abilities run through a holder`

No behavior change. The picker, the target prompt and `Dispatch` take a
`domain.AbilityHolder` in place of an asset. Every holder is still an asset.

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
// AbilityTargetCandidates bundles the selectable options for a target stage; the
// collector reads the slice matching the step's TargetKind.
type AbilityTargetCandidates struct {
```
**Replace with:**
```go
// AbilityHolder is who uses an ability: an A-flag asset, or its faction through a
// tag (Tag set, Asset nil). Exactly one is set.
type AbilityHolder struct {
	Asset *Asset
	Tag   *Tag
}

// AbilityTargetCandidates bundles the selectable options for a target stage; the
// collector reads the slice matching the step's TargetKind.
type AbilityTargetCandidates struct {
```

##### Task 2 — `internal/faction/engine/action/collector.go`

**Find:**
```go
	SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```
**Replace with:**
```go
	SelectAbilityTarget(holder domain.AbilityHolder, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```
**Find:**
```go
	SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
```
**Replace with:**
```go
	SelectAbilityHolders(faction *domain.Faction, candidates []domain.AbilityHolder, rulebook *rulebook.Rulebook) ([]domain.AbilityHolder, error)
```

##### Task 3 — `internal/faction/engine/action/actions/ability/dispatch.go`

**Find:**
```go
// Dispatch runs an asset's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations.
// Assets with no ability block, or whose effect has no registered handler, fall
// through to confirmApplied (the not-yet-built effects).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
```
**Replace with:**
```go
// Dispatch runs a holder's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations. def
// is an asset holder's definition. Assets with no ability block, or whose effect
// has no registered handler, fall through to confirmApplied (the not-yet-built
// effects).
func Dispatch(
	faction *domain.Faction,
	holder domain.AbilityHolder,
	def *domain.AssetDefinition,
```
**Find:**
```go
) ([]domain.Mutation, error) {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
```
**Replace with:**
```go
) ([]domain.Mutation, error) {
	asset := holder.Asset
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
```
**Find:**
```go
		target, err := resolveTarget(step, workingFaction, workingAsset, view, collector)
```
**Replace with:**
```go
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset}, view, collector)
```

##### Task 4 — `internal/faction/engine/action/actions/ability/stage.go`

**Find:**
```go
func resolveTarget(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	factionState *state.FactionState,
	collector action.Collector,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction.ID, asset, factionState)
	if picks && len(candidates.Factions)+len(candidates.Assets) == 0 {
		return domain.AbilityTarget{}, errNoTargets
	}
	switch step.Target {
	case domain.TargetNone:
		return domain.AbilityTarget{Kind: domain.TargetNone}, nil
	case domain.TargetSelf:
		return domain.AbilityTarget{Kind: domain.TargetSelf, Asset: asset, Faction: faction}, nil
	case domain.TargetFaction:
		target, err := collector.SelectAbilityTarget(asset, step, candidates)
```
**Replace with:**
```go
func resolveTarget(
	step domain.AbilityStep,
	faction *domain.Faction,
	holder domain.AbilityHolder,
	factionState *state.FactionState,
	collector action.Collector,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction.ID, holder.Asset, factionState)
	if picks && len(candidates.Factions)+len(candidates.Assets) == 0 {
		return domain.AbilityTarget{}, errNoTargets
	}
	switch step.Target {
	case domain.TargetNone:
		return domain.AbilityTarget{Kind: domain.TargetNone}, nil
	case domain.TargetSelf:
		return domain.AbilityTarget{Kind: domain.TargetSelf, Asset: holder.Asset, Faction: faction}, nil
	case domain.TargetFaction:
		target, err := collector.SelectAbilityTarget(holder, step, candidates)
```
**Find:**
```go
	case domain.TargetAsset:
		target, err := collector.SelectAbilityTarget(asset, step, candidates)
```
**Replace with:**
```go
	case domain.TargetAsset:
		target, err := collector.SelectAbilityTarget(holder, step, candidates)
```

##### Task 5 — `internal/faction/engine/action/actions/use_asset_ability.go`

Full contents:
```go
package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// UseAssetAbility activates the abilities of one or more holders. Holders are
// committed up front in resolution order; each is resolved in sequence.
type UseAssetAbility struct {
	collector action.Collector
	roller    domain.Roller
	registry  *hooks.Registry
	world     *world.WorldEngine
	selected  []domain.AbilityHolder
	mutations []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller, registry *hooks.Registry, worldEngine *world.WorldEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
		registry:  registry,
		world:     worldEngine,
	}
}

func (u *UseAssetAbility) Name() string { return "Use Asset Ability" }

func (u *UseAssetAbility) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	return len(abilityCandidates(faction, factionState, rulebook)) > 0
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := u.collector.SelectAbilityHolders(faction, abilityCandidates(faction, factionState, rulebook), rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selected = selected
	return nil
}

func (u *UseAssetAbility) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// deliberate: one view for the whole action, not one per asset, because the book lets one action use several assets and each must see what the earlier ones did (Effort 3 Decision 3).
	view := ability.NewView(factionState)
	for _, holder := range u.selected {
		def, ok := rulebook.Assets[holder.Asset.DefinitionID]
		if !ok {
			return fmt.Errorf("use asset ability: definition not found: %s", holder.Asset.DefinitionID)
		}
		mutations, err := ability.Dispatch(faction, holder, def, u.collector, u.roller, u.registry, factionState, view, rulebook, u.world)
		if err != nil {
			return fmt.Errorf("use asset ability: %w", err)
		}
		u.mutations = append(u.mutations, mutations...)
	}
	return nil
}

func (u *UseAssetAbility) Output() ([]domain.Mutation, error) {
	return u.mutations, nil
}

func hasActionFlag(def *domain.AssetDefinition) bool {
	for _, flag := range def.Flags {
		if flag == domain.FlagAction {
			return true
		}
	}
	return false
}

// abilityCandidates returns the holders that may use their ability now: the
// faction's A-flag assets, in ID order. A holder whose ability has no target now
// is not offered (Effort 3 Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.AbilityHolder {
	var candidates []domain.AbilityHolder
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, asset, def, factionState) {
			candidates = append(candidates, domain.AbilityHolder{Asset: asset})
		}
	}
	return candidates
}
```

##### Task 6 — `internal/faction/engine/action/actions/mocks/mock_collector.go`

Hand edit in the generated shape (Shared Context).

**Find:**
```go
// SelectAbilityAssets mocks base method.
func (m *MockCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, arg2 *rulebook.Rulebook) ([]*domain.Asset, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAbilityAssets", faction, candidates, arg2)
	ret0, _ := ret[0].([]*domain.Asset)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectAbilityAssets indicates an expected call of SelectAbilityAssets.
func (mr *MockCollectorMockRecorder) SelectAbilityAssets(faction, candidates, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAbilityAssets", reflect.TypeOf((*MockCollector)(nil).SelectAbilityAssets), faction, candidates, arg2)
}

// SelectAbilityTarget mocks base method.
func (m *MockCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAbilityTarget", asset, step, candidates)
	ret0, _ := ret[0].(domain.AbilityTarget)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectAbilityTarget indicates an expected call of SelectAbilityTarget.
func (mr *MockCollectorMockRecorder) SelectAbilityTarget(asset, step, candidates any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAbilityTarget", reflect.TypeOf((*MockCollector)(nil).SelectAbilityTarget), asset, step, candidates)
}
```
**Replace with:**
```go
// SelectAbilityHolders mocks base method.
func (m *MockCollector) SelectAbilityHolders(faction *domain.Faction, candidates []domain.AbilityHolder, arg2 *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAbilityHolders", faction, candidates, arg2)
	ret0, _ := ret[0].([]domain.AbilityHolder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectAbilityHolders indicates an expected call of SelectAbilityHolders.
func (mr *MockCollectorMockRecorder) SelectAbilityHolders(faction, candidates, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAbilityHolders", reflect.TypeOf((*MockCollector)(nil).SelectAbilityHolders), faction, candidates, arg2)
}

// SelectAbilityTarget mocks base method.
func (m *MockCollector) SelectAbilityTarget(holder domain.AbilityHolder, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAbilityTarget", holder, step, candidates)
	ret0, _ := ret[0].(domain.AbilityTarget)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectAbilityTarget indicates an expected call of SelectAbilityTarget.
func (mr *MockCollectorMockRecorder) SelectAbilityTarget(holder, step, candidates any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAbilityTarget", reflect.TypeOf((*MockCollector)(nil).SelectAbilityTarget), holder, step, candidates)
}
```

##### Task 7 — `internal/faction/engine/testharness/collector.go`

**Find:**
```go
	SelectAbilityAssetsFn         func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error)
```
**Replace with:**
```go
	SelectAbilityHoldersFn        func(*domain.Faction, []domain.AbilityHolder, *rulebook.Rulebook) ([]domain.AbilityHolder, error)
```
**Find:**
```go
	SelectAbilityTargetFn         func(*domain.Asset, domain.AbilityStep, domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```
**Replace with:**
```go
	SelectAbilityTargetFn         func(domain.AbilityHolder, domain.AbilityStep, domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```
**Find:**
```go
func (c *ScriptedCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error) {
	if c.SelectAbilityAssetsFn != nil {
		return c.SelectAbilityAssetsFn(faction, candidates, rulebook)
	}
	return nil, nil
}
```
**Replace with:**
```go
func (c *ScriptedCollector) SelectAbilityHolders(faction *domain.Faction, candidates []domain.AbilityHolder, rulebook *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
	if c.SelectAbilityHoldersFn != nil {
		return c.SelectAbilityHoldersFn(faction, candidates, rulebook)
	}
	return nil, nil
}
```
**Find:**
```go
func (c *ScriptedCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	if c.SelectAbilityTargetFn != nil {
		return c.SelectAbilityTargetFn(asset, step, candidates)
	}
```
**Replace with:**
```go
func (c *ScriptedCollector) SelectAbilityTarget(holder domain.AbilityHolder, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	if c.SelectAbilityTargetFn != nil {
		return c.SelectAbilityTargetFn(holder, step, candidates)
	}
```

##### Task 8 — `internal/faction/tui/adapter/channels.go`

**Find:**
```go
	AskSelectAbilityAssets
	AskConfirmAbilityApplied
```
**Replace with:**
```go
	AskSelectAbilityHolders
	AskConfirmAbilityApplied
```
**Find:**
```go
type SelectAbilityAssetsPayload struct{ Candidates []*domain.Asset }
```
**Replace with:**
```go
type SelectAbilityHoldersPayload struct{ Candidates []domain.AbilityHolder }
```

##### Task 9 — `internal/faction/tui/adapter/action_collector.go`

The faction and asset target prompts are only reached by asset holders (Decision 2),
so they keep reading `holder.Asset`.

**Find:**
```go
func (a *actionCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	switch step.Target {
	case domain.TargetFaction:
		raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
			Asset:      asset,
```
**Replace with:**
```go
func (a *actionCollector) SelectAbilityTarget(holder domain.AbilityHolder, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	switch step.Target {
	case domain.TargetFaction:
		raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
			Asset:      holder.Asset,
```
**Find:**
```go
		raw, err := a.adapter.ask(AskSelectAbilityAssetTarget, nil, SelectAbilityAssetTargetPayload{
			Asset:      asset,
```
**Replace with:**
```go
		raw, err := a.adapter.ask(AskSelectAbilityAssetTarget, nil, SelectAbilityAssetTargetPayload{
			Asset:      holder.Asset,
```
**Find:**
```go
func (a *actionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAbilityAssets, faction, SelectAbilityAssetsPayload{Candidates: candidates})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}
```
**Replace with:**
```go
func (a *actionCollector) SelectAbilityHolders(faction *domain.Faction, candidates []domain.AbilityHolder, rb *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
	raw, err := a.adapter.ask(AskSelectAbilityHolders, faction, SelectAbilityHoldersPayload{Candidates: candidates})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]domain.AbilityHolder)
	return chosen, nil
}
```

##### Task 10 — `internal/faction/tui/adapter/smoke.go`

**Find:**
```go
func (SmokeActionCollector) SelectAbilityTarget(_ *domain.Asset, _ domain.AbilityStep, _ domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
```
**Replace with:**
```go
func (SmokeActionCollector) SelectAbilityTarget(_ domain.AbilityHolder, _ domain.AbilityStep, _ domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
```
**Find:**
```go
func (SmokeActionCollector) SelectAbilityAssets(_ *domain.Faction, _ []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
```
**Replace with:**
```go
func (SmokeActionCollector) SelectAbilityHolders(_ *domain.Faction, _ []domain.AbilityHolder, _ *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
```

##### Task 11 — `internal/faction/tui/views/turn/overlay/ability.go` and `internal/faction/tui/views/turn/execution/execution.go`

In `overlay/ability.go`:

**Find:**
```go
// NewSelectAbilityAssets builds the multi-select of A-flagged assets whose
// abilities the faction activates (committed up front, resolved in order).
// Labels carry the ability effect. Answer: []*domain.Asset (uncapped). Esc cancels.
func NewSelectAbilityAssets(candidates []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(candidates))
	for _, asset := range candidates {
		opts = append(opts, huh.NewOption(abilityLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Activate which abilities?", "", opts, 0)
}
```
**Replace with:**
```go
// NewSelectAbilityHolders builds the multi-select of the holders whose abilities
// the faction activates (committed up front, resolved in order). Labels carry the
// ability effect. Answer: []domain.AbilityHolder (uncapped). Esc cancels.
func NewSelectAbilityHolders(candidates []domain.AbilityHolder, rb *rulebook.Rulebook) MultiSelectOverlay[domain.AbilityHolder] {
	opts := make([]huh.Option[domain.AbilityHolder], 0, len(candidates))
	for _, holder := range candidates {
		opts = append(opts, huh.NewOption(abilityLabel(holder, rb), holder))
	}
	return NewMultiSelect[domain.AbilityHolder]("Activate which abilities?", "", opts, 0)
}
```
**Find:**
```go
func abilityLabel(asset *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(asset, rb)
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && len(def.Ability.Steps) > 0 {
```
**Replace with:**
```go
func abilityLabel(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	asset := holder.Asset
	name := assetName(asset, rb)
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && len(def.Ability.Steps) > 0 {
```

In `execution/execution.go`:

**Find:**
```go
	adapter.AskSelectAbilityAssets: {"Ability Assets", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAbilityAssets(msg.Payload.(adapter.SelectAbilityAssetsPayload).Candidates, m.rulebook)
	}},
```
**Replace with:**
```go
	adapter.AskSelectAbilityHolders: {"Abilities", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAbilityHolders(msg.Payload.(adapter.SelectAbilityHoldersPayload).Candidates, m.rulebook)
	}},
```

##### Task 12 — tests

Mechanical, no new tests. Each row is an Edit with `replace_all: true`.

| File | Replace every | With |
|---|---|---|
| `ability/dispatch_test.go` | `Dispatch(faction, asset, def,` | `Dispatch(faction, domain.AbilityHolder{Asset: asset}, def,` |
| `ability/dispatch_test.go`, `ability/levy_coin_test.go`, `ability/block_attack_test.go`, `ability/disable_asset_test.go` | `Dispatch(faction, acting, def,` | `Dispatch(faction, domain.AbilityHolder{Asset: acting}, def,` |
| `ability/stage_test.go`, `ability/levy_coin_test.go`, `ability/block_attack_test.go`, `ability/disable_asset_test.go`, `use_asset_ability_test.go`, `scenarios/seditionists_test.go`, `scenarios/marketers_test.go` | `func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates)` | `func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates)` |
| `ability/stage_test.go` | `collector.EXPECT().SelectAbilityTarget(acting, step, gomock.Any()).` | `collector.EXPECT().SelectAbilityTarget(domain.AbilityHolder{Asset: acting}, step, gomock.Any()).` |
| `ability/stage_test.go` | `resolveTarget(step, factionState.Factions["f1"], acting, factionState, collector)` | `resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector)` |
| `use_asset_ability_test.go` | `SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{actingAsset}, nil)` | `SelectAbilityHolders(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.AbilityHolder{{Asset: actingAsset}}, nil)` |
| `use_asset_ability_test.go` | `SelectAbilityAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{first, second}, nil)` | `SelectAbilityHolders(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.AbilityHolder{{Asset: first}, {Asset: second}}, nil)` |

The `ability/` paths are under `internal/faction/engine/action/actions/`;
`use_asset_ability_test.go` sits there too; `scenarios/` is under
`internal/faction/engine/testharness/`.

In `scenarios/seditionists_test.go`:

**Find:**
```go
	h.Collector.SelectAbilityAssetsFn = func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{seditionists}, nil
	}
```
**Replace with:**
```go
	h.Collector.SelectAbilityHoldersFn = func(*domain.Faction, []domain.AbilityHolder, *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
		return []domain.AbilityHolder{{Asset: seditionists}}, nil
	}
```
In `scenarios/marketers_test.go`:

**Find:**
```go
	h.Collector.SelectAbilityAssetsFn = func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{marketers}, nil
	}
```
**Replace with:**
```go
	h.Collector.SelectAbilityHoldersFn = func(*domain.Faction, []domain.AbilityHolder, *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
		return []domain.AbilityHolder{{Asset: marketers}}, nil
	}
```

##### Commit message

```
refactor(E5): abilities run through a holder

- domain.AbilityHolder: an A-flag asset or a tag, exactly one set
- SelectAbilityAssets becomes SelectAbilityHolders; SelectAbilityTarget and
  ability.Dispatch take the holder
- TUI picker, adapter, smoke collector, scripted collector and mock follow
- no behavior change: every holder is still an asset
```

#### Commit 2 — `feat(E5): tags carry abilities`

##### Task 1 — `internal/faction/domain/faction.go`

**Find:**
```go
	HomeworldGovernment bool
	Effects             []EffectDefinition
}
```
**Replace with:**
```go
	HomeworldGovernment bool
	Effects             []EffectDefinition
	// Ability is a tag's action, used through Use Asset Ability; nil for most tags.
	Ability *AbilityDefinition
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`, the tag record

**Find:**
```go
	HomeworldGovernment bool           `toml:"homeworld_government"`
	Effects             []effectRecord `toml:"effects"`
}
```
**Replace with:**
```go
	HomeworldGovernment bool           `toml:"homeworld_government"`
	Effects             []effectRecord `toml:"effects"`
	Ability             *abilityRecord `toml:"ability"`
}
```
**Find:**
```go
		effects, err := convertEffects(r.Effects)
		if err != nil {
			return nil, "", fmt.Errorf("tag %q in %s: %w", r.ID, path, err)
		}
		result[r.ID] = &domain.Tag{
			ID:                  r.ID,
			Name:                r.Name,
			Description:         r.Description,
			Effect:              r.Effect,
			Government:          r.Government,
			HomeworldGovernment: r.HomeworldGovernment,
			Effects:             effects,
		}
```
**Replace with:**
```go
		effects, err := convertEffects(r.Effects)
		if err != nil {
			return nil, "", fmt.Errorf("tag %q in %s: %w", r.ID, path, err)
		}
		ability, err := convertAbility(r.Ability)
		if err != nil {
			return nil, "", fmt.Errorf("tag %q in %s: ability: %w", r.ID, path, err)
		}
		if err := checkTagAbility(ability); err != nil {
			return nil, "", fmt.Errorf("tag %q in %s: %w", r.ID, path, err)
		}
		result[r.ID] = &domain.Tag{
			ID:                  r.ID,
			Name:                r.Name,
			Description:         r.Description,
			Effect:              r.Effect,
			Government:          r.Government,
			HomeworldGovernment: r.HomeworldGovernment,
			Effects:             effects,
			Ability:             ability,
		}
```

##### Task 3 — `internal/faction/rulebook/rulebook.go`, the tag ability check

Insert after `loadTags` (before the `// Goals` banner):

**Find:**
```go
	return result, government, nil
}
```
**Replace with:**
```go
	return result, government, nil
}

// tagAbilityEffects are the ability effects a tag may carry: those that read no
// acting asset (Effort 5 Decision 2). Every other effect reads the asset or its
// world.
var tagAbilityEffects = []domain.AbilityEffectType{domain.EffectCoinDelta, domain.EffectPurchaseDiscount}

// checkTagAbility refuses a tag ability that needs an acting asset: an effect
// outside tagAbilityEffects, a contest (it rolls from the asset's world), a target
// the asset picks from, or an outcome row that destroys or charges the asset.
func checkTagAbility(ability *domain.AbilityDefinition) error {
	if ability == nil {
		return nil
	}
	for i, step := range ability.Steps {
		if !slices.Contains(tagAbilityEffects, step.Effect) {
			return fmt.Errorf("ability step %d: a tag cannot carry effect %q", i, step.Effect)
		}
		if step.Contest != nil {
			return fmt.Errorf("ability step %d: a tag ability cannot hold a contest", i)
		}
		if step.Target != domain.TargetNone && step.Target != domain.TargetWorld {
			return fmt.Errorf("ability step %d: a tag ability cannot target %q", i, step.Target)
		}
		if step.Roll != nil && slices.ContainsFunc(step.Roll.Outcomes, func(outcome domain.AbilityOutcome) bool {
			return outcome.DestroySelf || outcome.Coin < 0
		}) {
			return fmt.Errorf("ability step %d: a tag has no asset to destroy or to lose for an unpaid loss", i)
		}
	}
	return nil
}
```

##### Task 4 — `internal/faction/engine/action/actions/ability/dispatch.go`

Replace the whole `Dispatch` function (as left by Commit 1 Task 3) with:
```go
// Dispatch runs a holder's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations. def
// is an asset holder's definition, nil for a tag. Assets with no ability block, or
// whose effect has no registered handler, fall through to confirmApplied (the
// not-yet-built effects). The loader lets a tag carry only built effects.
func Dispatch(
	faction *domain.Faction,
	holder domain.AbilityHolder,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	registry *hooks.Registry,
	factionState *state.FactionState,
	view *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
	asset := holder.Asset
	abilityDefinition := holderAbility(holder, def)
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	// Pre-flight: if any step's effect is not yet built, the whole ability is a
	// not-yet-built no-op. Bail before executing any step, so a built early step
	// never half-applies (moves an asset, spends Coin) only to be discarded when a
	// later unbuilt step short-circuits to confirmApplied.
	for _, step := range abilityDefinition.Steps {
		if _, ok := handlers[step.Effect]; !ok {
			return confirmApplied(asset, def, collector)
		}
	}
	workingFaction := view.Factions[faction.ID]
	if workingFaction == nil {
		return nil, fmt.Errorf("ability: acting faction %s not in the view", faction.ID)
	}
	var workingAsset *domain.Asset
	if asset != nil {
		workingAsset = workingFaction.Assets[asset.ID]
		if workingAsset == nil {
			// deliberate: fizzle, not act from real state, because an asset missing from the view was removed by an earlier asset in this action.
			return nil, nil
		}
	}
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)

	var mutations []domain.Mutation
	for _, step := range abilityDefinition.Steps {
		handler := handlers[step.Effect]
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector)
		// deliberate: fizzle this asset, not fail the action, because an earlier asset in the same action may have used up the targets and its result must stand (Effort 3 Decision Record 1).
		if errors.Is(err, errNoTargets) {
			break
		}
		if err != nil {
			return nil, err
		}
		// deliberate: the real faction, not workingFaction, because a roll hook's budget spent on the view would be lost (Effort 3 plan, Decision 12).
		passed, err := passesGate(step, faction, workingAsset, target, registry, collector, roller, factionState, rb)
		if err != nil {
			return nil, err
		}
		if !passed {
			// deliberate: halt the chain on a failed gate, not collect-and-continue,
			// because a missed contest fizzles the rest of the ability. Multi-step
			// abilities must therefore author ungated steps (relocate, roll-table)
			// before gated ones (Effort 1 Plan, Decision 1).
			break
		}
		stepMutations, err := handler(StepContext{
			Faction:      workingFaction,
			Asset:        workingAsset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: view,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
			Registry:     registry,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
		// deliberate: drop the projection's apply error, not return it, because Output() applies the same mutations against real state, where a genuine miss still surfaces.
		_ = mutationEngine.Apply(view, stepMutations, discardLog)
		if asset == nil {
			continue
		}
		if updated, ok := workingFaction.Assets[asset.ID]; ok {
			workingAsset = updated
		}
	}
	return mutations, nil
}

// holderAbility returns the ability a holder uses: its tag's, or its asset
// definition's.
func holderAbility(holder domain.AbilityHolder, def *domain.AssetDefinition) *domain.AbilityDefinition {
	if holder.Tag != nil {
		return holder.Tag.Ability
	}
	return def.Ability
}
```
In the `StepContext` doc, **Find:**
```go
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
```
**Replace with:**
```go
	Asset        *domain.Asset           // nil for a tag holder
	Def          *domain.AssetDefinition // nil for a tag holder
	Step         domain.AbilityStep
```

##### Task 5 — `internal/faction/engine/action/actions/ability/stage.go`

**Find:**
```go
// HasTargets reports whether the first step of def's ability can pick a target
// now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, factionState *state.FactionState) bool {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(def.Ability.Steps[0], faction.ID, asset, factionState)
```
**Replace with:**
```go
// HasTargets reports whether the first step of a holder's ability can pick a
// target now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, holder domain.AbilityHolder, abilityDefinition *domain.AbilityDefinition, factionState *state.FactionState) bool {
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(abilityDefinition.Steps[0], faction.ID, holder.Asset, factionState)
```

##### Task 6 — `internal/faction/engine/action/actions/use_asset_ability.go`

**Find:**
```go
// UseAssetAbility activates the abilities of one or more holders. Holders are
// committed up front in resolution order; each is resolved in sequence.
```
**Replace with:**
```go
// UseAssetAbility activates the abilities of one or more holders: A-flagged assets
// and the faction's tags (Effort 5 Decision 1). Holders are committed up front in
// resolution order; each is resolved in sequence.
```
**Find:**
```go
	for _, holder := range u.selected {
		def, ok := rulebook.Assets[holder.Asset.DefinitionID]
		if !ok {
			return fmt.Errorf("use asset ability: definition not found: %s", holder.Asset.DefinitionID)
		}
		mutations, err := ability.Dispatch(
```
**Replace with:**
```go
	for _, holder := range u.selected {
		var def *domain.AssetDefinition
		if holder.Asset != nil {
			found, ok := rulebook.Assets[holder.Asset.DefinitionID]
			if !ok {
				return fmt.Errorf("use asset ability: definition not found: %s", holder.Asset.DefinitionID)
			}
			def = found
		}
		mutations, err := ability.Dispatch(
```
**Find:**
```go
// abilityCandidates returns the holders that may use their ability now: the
// faction's A-flag assets, in ID order. A holder whose ability has no target now
// is not offered (Effort 3 Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.AbilityHolder {
	var candidates []domain.AbilityHolder
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, asset, def, factionState) {
			candidates = append(candidates, domain.AbilityHolder{Asset: asset})
		}
	}
	return candidates
}
```
**Replace with:**
```go
// abilityCandidates returns the holders that may use their ability now: the
// faction's A-flag assets in ID order, then its tags that carry an ability, in ID
// order. A holder whose ability has no target now is not offered (Effort 3
// Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.AbilityHolder {
	var candidates []domain.AbilityHolder
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		holder := domain.AbilityHolder{Asset: asset}
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, holder, def.Ability, factionState) {
			candidates = append(candidates, holder)
		}
	}
	for _, tagID := range slices.Sorted(slices.Values(faction.Tags)) {
		tag, ok := rulebook.Tags[tagID]
		if !ok || tag.Ability == nil {
			continue
		}
		holder := domain.AbilityHolder{Tag: tag}
		if ability.HasTargets(faction, holder, tag.Ability, factionState) {
			candidates = append(candidates, holder)
		}
	}
	return candidates
}
```
Add `"slices"` to the imports.

##### Task 7 — `internal/faction/tui/views/turn/overlay/ability.go`

**Find:**
```go
// abilityLabel renders "<name> · <effect>" (or just the name when the definition
// carries no ability effect — e.g. the structural-stub abilities).
func abilityLabel(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	asset := holder.Asset
```
**Replace with:**
```go
// abilityLabel renders "<name> · <effect>" (or just the name when the definition
// carries no ability effect — e.g. the structural-stub abilities). A tag's name
// is marked "(tag)".
func abilityLabel(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	if holder.Tag != nil {
		return fmt.Sprintf("%s (tag) · %s", holder.Tag.Name, holder.Tag.Ability.Steps[0].Effect)
	}
	asset := holder.Asset
```
Append at the end of the file:
```go
// holderName is an ability holder's display name: its tag's, or its asset's.
func holderName(holder domain.AbilityHolder, rb *rulebook.Rulebook) string {
	if holder.Tag != nil {
		return holder.Tag.Name
	}
	return assetName(holder.Asset, rb)
}
```
(`holderName` is first read by Commit 3 Task 13; it sits here beside the label it
belongs with.)

##### Task 8 — tests

In `internal/faction/engine/action/actions/ability/stage_test.go`, `TestHasTargets`,
replace every:

| Replace every | With |
|---|---|
| `HasTargets(factionState.Factions["f1"], acting, definition(` | `HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, definition(` |
| `HasTargets(faction, acting, definition(` | `HasTargets(faction, domain.AbilityHolder{Asset: acting}, definition(` |

Then change its helper so it returns the ability, not a definition. **Find:**
```go
	definition := func(steps ...domain.AbilityStep) *domain.AssetDefinition {
		return &domain.AssetDefinition{ID: "fixture", Ability: &domain.AbilityDefinition{Steps: steps}}
	}
```
**Replace with:**
```go
	definition := func(steps ...domain.AbilityStep) *domain.AbilityDefinition {
		return &domain.AbilityDefinition{Steps: steps}
	}
```

Append to `internal/faction/rulebook/rulebook_test.go`:
```go
func TestLoadTags_Ability(t *testing.T) {
	load := func(ability string) (map[string]*domain.Tag, error) {
		path := filepath.Join(t.TempDir(), "tags.toml")
		body := "[tags.Alpha]\nid = \"T-001\"\nname = \"Alpha\"\n\n[tags.Alpha.ability]\n" + ability
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}
		tags, _, err := loadTags(path)
		return tags, err
	}

	tags, err := load("effect = \"purchase_discount\"\ndie = \"1d4\"\n")
	if err != nil {
		t.Fatalf("an asset-free effect: %v", err)
	}
	want := &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
		Effect: domain.EffectPurchaseDiscount,
		Target: domain.TargetNone,
		Roll:   &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 4}, Outcomes: []domain.AbilityOutcome{}},
	}}}
	if got := tags["T-001"].Ability; !reflect.DeepEqual(got, want) {
		t.Errorf("ability = %+v, want %+v", got, want)
	}

	refused := []struct {
		name    string
		ability string
		want    string
	}{
		{"an effect that reads the asset", "effect = \"levy_coin\"\ntarget = \"faction\"\n", `cannot carry effect "levy_coin"`},
		{"a target the asset picks from", "effect = \"coin_delta\"\ntarget = \"asset\"\ndie = \"1d6\"\n", `cannot target "asset"`},
		{"a loss row", "effect = \"coin_delta\"\ndie = \"1d6\"\n[[tags.Alpha.ability.outcomes]]\nfrom = 1\nto = 1\ncoin = -1\n", "no asset to destroy"},
		{"a destroy row", "effect = \"coin_delta\"\ndie = \"1d6\"\n[[tags.Alpha.ability.outcomes]]\nfrom = 1\nto = 1\ndestroy_self = true\n", "no asset to destroy"},
	}
	for _, tc := range refused {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load(tc.ability)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
```
(The contest refusal joins this test in Commit 5 Task 9, written with the `roll`
key that commit makes required, so the case is written once.)

Append to `internal/faction/engine/action/actions/use_asset_ability_test.go`:
```go
// TestUseAssetAbility_TagHolder: f1 holds a tag whose ability is a purchase
// discount and no A-flag asset. The tag is the only candidate, and using it grants
// the rolled discount from the faction, with no asset.
func TestUseAssetAbility_TagHolder(t *testing.T) {
	tag := &domain.Tag{ID: "T-1", Name: "Brokers", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
		Effect: domain.EffectPurchaseDiscount,
		Target: domain.TargetNone,
		Roll:   &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 4}},
	}}}}
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{}, Tags: map[string]*domain.Tag{"T-1": tag, "T-2": {ID: "T-2"}}}
	faction := &domain.Faction{ID: "f1", Tags: []string{"T-2", "T-1"}, Assets: map[string]*domain.Asset{}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityHolders(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Faction, candidates []domain.AbilityHolder, _ *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
			if want := []domain.AbilityHolder{{Tag: tag}}; !reflect.DeepEqual(candidates, want) {
				t.Errorf("candidates = %+v, want only the tag with an ability", candidates)
			}
			return candidates, nil
		})

	if !NewUseAssetAbility(collector, nil, nil, nil).Validate(faction, factionState, book) {
		t.Fatal("Validate = false, want true for a held tag ability")
	}
	mutations := runAbility(t, collector, &fixedRoller{values: []int{3}}, faction, factionState, book)

	want := []domain.Mutation{domain.PurchaseDiscountGranted{FactionID: "f1", Amount: 3, Cause: "ability", CausedByFactionID: "f1"}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v", mutations, want)
	}
}
```
The variable is `book`, not `rulebook`, because the closure names the
`*rulebook.Rulebook` type.

##### Commit message

```
feat(E5): tags carry abilities

- a tag record takes an [ability] block, in the asset shape
- the loader refuses a tag ability that needs an acting asset: an effect
  other than coin_delta or purchase_discount, a contest, a self, faction or
  asset target, or a loss or destroy row (Decision 2)
- Use Asset Ability lists held tags with an ability after the assets;
  Dispatch runs a tag holder with no asset (Decision 1)
- the picker marks a tag "(tag)"
```

### Phase 2 — Teach Planetary Population

**In plain words:** Preceptor Archive's Teach Planetary Population works end to end:
pick a world with a Base, pay 2 Coin, roll, and on a 12 the world counts as TL4 for
that faction's purchases.

#### Commit 3 — `feat(E5): Teach Planetary Population`

##### Task 1 — `internal/faction/domain/asset.go`, the effect and its parameters

**Find:**
```go
	EffectLevyCoin         AbilityEffectType = "levy_coin"
)
```
**Replace with:**
```go
	EffectLevyCoin         AbilityEffectType = "levy_coin"
	EffectRaiseTechLevel   AbilityEffectType = "raise_tech_level"
)
```
**Find:**
```go
type TargetKind string
```
**Replace with:**
```go
// RaiseTechLevelParams configures a raise_tech_level step (Teach Planetary
// Population): CoinCost is paid on every use; on a matching outcome row the target
// world counts as TechLevel for the acting faction's purchases (D-11).
type RaiseTechLevelParams struct {
	CoinCost  int
	TechLevel int
}

type TargetKind string
```
**Find:**
```go
	TargetAsset   TargetKind = "asset"
	// deliberate: TargetWorld is dormant vocab, not dead — no asset targets a world yet.
	TargetWorld TargetKind = "world"
```
**Replace with:**
```go
	TargetAsset   TargetKind = "asset"
	TargetWorld   TargetKind = "world"
```
**Find:**
```go
	Relocate *RelocateParams
	Purchase *PurchaseParams
}
```
**Replace with:**
```go
	Relocate       *RelocateParams
	Purchase       *PurchaseParams
	RaiseTechLevel *RaiseTechLevelParams
}
```
**Find:**
```go
	Assets   []*Asset
	// deliberate: Worlds is dormant, not dead. See TargetWorld.
	Worlds []string
}
```
**Replace with:**
```go
	Assets   []*Asset
	Worlds   []string
}

// Len counts every candidate, of any kind.
func (candidates AbilityTargetCandidates) Len() int {
	return len(candidates.Factions) + len(candidates.Assets) + len(candidates.Worlds)
}
```

##### Task 2 — `internal/faction/domain/faction.go`, the record

**Find:**
```go
	PopulationWaiver bool `toml:"population_waiver,omitempty"`
}
```
**Replace with:**
```go
	PopulationWaiver bool `toml:"population_waiver,omitempty"`
	// WorldTechLevels holds the worlds that count as a higher tech level for this
	// faction's purchases, by world ID (raise_tech_level, D-11).
	WorldTechLevels map[string]int `toml:"world_tech_levels,omitempty"`
}
```
In `Clone`, **Find:**
```go
	if faction.HookBudgets != nil {
		clone.HookBudgets = make(map[string]int, len(faction.HookBudgets))
		maps.Copy(clone.HookBudgets, faction.HookBudgets)
	}
	return &clone
```
**Replace with:**
```go
	if faction.HookBudgets != nil {
		clone.HookBudgets = make(map[string]int, len(faction.HookBudgets))
		maps.Copy(clone.HookBudgets, faction.HookBudgets)
	}
	if faction.WorldTechLevels != nil {
		clone.WorldTechLevels = maps.Clone(faction.WorldTechLevels)
	}
	return &clone
```

##### Task 3 — `internal/faction/domain/mutation.go`

**Find:**
```go
func (m PurchaseDiscountConsumed) Type() string { return "purchase_discount_consumed" }
```
**Replace with:**
```go
func (m PurchaseDiscountConsumed) Type() string { return "purchase_discount_consumed" }

// WorldTechLevelRaised records that WorldID counts as TechLevel for FactionID's
// purchases (raise_tech_level, D-11). A lower value never replaces a higher one.
type WorldTechLevelRaised struct {
	FactionID         string `json:"faction_id"`
	WorldID           string `json:"world_id"`
	TechLevel         int    `json:"tech_level"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m WorldTechLevelRaised) Type() string { return "world_tech_level_raised" }
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

**Find:**
```go
		case domain.PurchaseDiscountConsumed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = 0
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```
**Replace with:**
```go
		case domain.PurchaseDiscountConsumed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = 0
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.WorldTechLevelRaised:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.WorldTechLevels == nil {
					faction.WorldTechLevels = map[string]int{}
				}
				faction.WorldTechLevels[v.WorldID] = max(faction.WorldTechLevels[v.WorldID], v.TechLevel)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```

##### Task 5 — `internal/faction/rulebook/rulebook.go`, parsing

**Find:**
```go
	Relocate     *relocateRecord `toml:"relocate"`
	Purchase     *purchaseRecord `toml:"purchase"`
}
```
**Replace with:**
```go
	Relocate       *relocateRecord       `toml:"relocate"`
	Purchase       *purchaseRecord       `toml:"purchase"`
	RaiseTechLevel *raiseTechLevelRecord `toml:"raise_tech_level"`
}

type raiseTechLevelRecord struct {
	CoinCost  int `toml:"coin_cost"`
	TechLevel int `toml:"tech_level"`
}
```
(Re-run `gofmt`; the existing fields of `abilityStepRecord` realign.)

In `convertAbilityStep`, **Find:**
```go
	if step.Contest != nil && step.Roll != nil {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and roll; they are mutually exclusive")
	}
```
**Replace with:**
```go
	if raise := record.RaiseTechLevel; raise != nil {
		if raise.CoinCost < 0 || raise.TechLevel <= 0 {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level needs coin_cost >= 0 and tech_level > 0, got %d and %d", raise.CoinCost, raise.TechLevel)
		}
		step.RaiseTechLevel = &domain.RaiseTechLevelParams{CoinCost: raise.CoinCost, TechLevel: raise.TechLevel}
	}
	if step.Contest != nil && step.Roll != nil {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and roll; they are mutually exclusive")
	}
	if step.Effect == domain.EffectRaiseTechLevel {
		if step.Target != domain.TargetWorld || step.Roll == nil || step.RaiseTechLevel == nil {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level requires target = %q, a die and a [raise_tech_level] block", "world")
		}
		if slices.ContainsFunc(step.Roll.Outcomes, func(outcome domain.AbilityOutcome) bool { return outcome.Coin != 0 || outcome.DestroySelf }) {
			return domain.AbilityStep{}, fmt.Errorf("raise_tech_level outcome rows mark a success; they carry no coin or destroy_self")
		}
	}
```
In `toAbilityEffect`, **Find:**
```go
	case "levy_coin":
		return domain.EffectLevyCoin, nil
```
**Replace with:**
```go
	case "levy_coin":
		return domain.EffectLevyCoin, nil
	case "raise_tech_level":
		return domain.EffectRaiseTechLevel, nil
```
**Find:**
```go
var tagAbilityEffects = []domain.AbilityEffectType{domain.EffectCoinDelta, domain.EffectPurchaseDiscount}
```
**Replace with:**
```go
var tagAbilityEffects = []domain.AbilityEffectType{domain.EffectCoinDelta, domain.EffectPurchaseDiscount, domain.EffectRaiseTechLevel}
```

##### Task 6 — `internal/faction/engine/action/actions/ability/raise_tech_level.go` (new)

Full contents:
```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// raiseTechLevel is Teach Planetary Population: pay the step's cost, and on a
// matching outcome row the target world counts as the step's tech level for the
// acting faction's purchases (D-11). The cost is paid on a miss too. A faction
// that cannot pay does nothing.
func raiseTechLevel(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.RaiseTechLevel
	if params == nil || ctx.Step.Roll == nil {
		return nil, fmt.Errorf("raise_tech_level: step has no roll table or parameters")
	}
	if ctx.Faction.Coin < params.CoinCost {
		return nil, nil
	}
	var mutations []domain.Mutation
	if params.CoinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{
			FactionID:         ctx.Faction.ID,
			Delta:             -params.CoinCost,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		})
	}
	if _, ok := outcomeFor(ctx.RollResult, ctx.Step.Roll.Outcomes); ok {
		mutations = append(mutations, domain.WorldTechLevelRaised{
			FactionID:         ctx.Faction.ID,
			WorldID:           ctx.Target.World,
			TechLevel:         params.TechLevel,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		})
	}
	return mutations, nil
}
```

##### Task 7 — `internal/faction/engine/action/actions/ability/dispatch.go`

**Find:**
```go
	domain.EffectLevyCoin:         levyCoin,
}
```
**Replace with:**
```go
	domain.EffectLevyCoin:         levyCoin,
	domain.EffectRaiseTechLevel:   raiseTechLevel,
}
```
**Find:**
```go
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector)
```
**Replace with:**
```go
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector, worldEngine)
```

##### Task 8 — `internal/faction/engine/action/actions/ability/stage.go`, targets

**Find:**
```go
	collector action.Collector,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction.ID, holder.Asset, factionState)
	if picks && len(candidates.Factions)+len(candidates.Assets) == 0 {
		return domain.AbilityTarget{}, errNoTargets
	}
```
**Replace with:**
```go
	collector action.Collector,
	worldEngine *world.WorldEngine,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction, holder.Asset, factionState, worldEngine)
	if picks && candidates.Len() == 0 {
		return domain.AbilityTarget{}, errNoTargets
	}
```
**Find:**
```go
	default:
		// deliberate: world targets are not unbuilt-by-oversight — no asset targets a world yet. See domain.TargetWorld.
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
	}
```
**Replace with:**
```go
	case domain.TargetWorld:
		target, err := collector.SelectAbilityTarget(holder, step, candidates)
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.World == "" {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		return domain.AbilityTarget{Kind: domain.TargetWorld, World: target.World, Faction: faction}, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported", step.Target)
	}
```
**Find:**
```go
func HasTargets(faction *domain.Faction, holder domain.AbilityHolder, abilityDefinition *domain.AbilityDefinition, factionState *state.FactionState) bool {
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(abilityDefinition.Steps[0], faction.ID, holder.Asset, factionState)
	return !picks || len(candidates.Factions)+len(candidates.Assets) > 0
}

// stepCandidates returns the candidates for a step that picks a faction or an
// asset. picks is false for a step that picks neither.
func stepCandidates(step domain.AbilityStep, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) (candidates domain.AbilityTargetCandidates, picks bool) {
	switch step.Target {
	case domain.TargetFaction:
		return domain.AbilityTargetCandidates{Factions: factionCandidates(step.Effect, actingFactionID, acting, factionState)}, true
	case domain.TargetAsset:
		return domain.AbilityTargetCandidates{Assets: assetCandidates(step.Effect, actingFactionID, acting, factionState)}, true
	}
	return domain.AbilityTargetCandidates{}, false
}
```
**Replace with:**
```go
func HasTargets(faction *domain.Faction, holder domain.AbilityHolder, abilityDefinition *domain.AbilityDefinition, factionState *state.FactionState, worldEngine *world.WorldEngine) bool {
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(abilityDefinition.Steps[0], faction, holder.Asset, factionState, worldEngine)
	return !picks || candidates.Len() > 0
}

// stepCandidates returns the candidates for a step that picks a faction, an asset
// or a world. picks is false for a step that picks none of them.
func stepCandidates(step domain.AbilityStep, faction *domain.Faction, acting *domain.Asset, factionState *state.FactionState, worldEngine *world.WorldEngine) (candidates domain.AbilityTargetCandidates, picks bool) {
	switch step.Target {
	case domain.TargetFaction:
		return domain.AbilityTargetCandidates{Factions: factionCandidates(step.Effect, faction.ID, acting, factionState)}, true
	case domain.TargetAsset:
		return domain.AbilityTargetCandidates{Assets: assetCandidates(step.Effect, faction.ID, acting, factionState)}, true
	case domain.TargetWorld:
		rule := worldTargetRules[step.Effect]
		if rule == nil {
			return domain.AbilityTargetCandidates{}, true
		}
		return domain.AbilityTargetCandidates{Worlds: rule(step, faction, worldEngine)}, true
	}
	return domain.AbilityTargetCandidates{}, false
}

// worldTargetRules lists a step's world candidates by its effect. An effect with
// no rule has none.
var worldTargetRules = map[domain.AbilityEffectType]func(step domain.AbilityStep, faction *domain.Faction, worldEngine *world.WorldEngine) []string{
	domain.EffectRaiseTechLevel: raiseTechLevelWorlds,
}

// raiseTechLevelWorlds returns, sorted, the worlds where faction has a Base whose
// own tech level is below the step's and that it has no record for yet; none when
// faction cannot pay the cost (Effort 5 Decision 4). With no world engine no
// world's tech level is known, so there are none.
func raiseTechLevelWorlds(step domain.AbilityStep, faction *domain.Faction, worldEngine *world.WorldEngine) []string {
	params := step.RaiseTechLevel
	if worldEngine == nil || params == nil || faction.Coin < params.CoinCost {
		return nil
	}
	var worlds []string
	for _, base := range faction.Bases {
		worldID := base.Location.WorldID
		if slices.Contains(worlds, worldID) || faction.WorldTechLevels[worldID] > 0 {
			continue
		}
		location, ok := worldEngine.Location(worldID)
		if !ok || location.TechLevel() >= params.TechLevel {
			continue
		}
		worlds = append(worlds, worldID)
	}
	slices.Sort(worlds)
	return worlds
}
```
Add the `world` import (`.../engine/world`).

##### Task 9 — `internal/faction/engine/action/actions/use_asset_ability.go`

Replace every `abilityCandidates(faction, factionState, rulebook)` with
`abilityCandidates(faction, factionState, rulebook, u.world)` (two call sites,
in `Validate` and `Inputs`). Then **Find:**
```go
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.AbilityHolder {
```
**Replace with:**
```go
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine) []domain.AbilityHolder {
```
Replace every `ability.HasTargets(faction, holder, def.Ability, factionState)` with
`ability.HasTargets(faction, holder, def.Ability, factionState, worldEngine)`, and
`ability.HasTargets(faction, holder, tag.Ability, factionState)` with
`ability.HasTargets(faction, holder, tag.Ability, factionState, worldEngine)`.

##### Task 10 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
// ResolveWorldTechLevel runs faction's own WorldTechLevelModifiers, then those
// registered against each asset faction owns, in asset ID order. A tech-level
// floor serves its owner only, so a rival's assets are never asked; each
// asset-scoped hook decides whether its reach covers world and def.
func ResolveWorldTechLevel(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int {
	if registry == nil {
		return baseTL
	}
	tl := baseTL
```
**Replace with:**
```go
// ResolveWorldTechLevel starts from the higher of baseTL and faction's own record
// for world (raise_tech_level, D-11), then runs faction's own
// WorldTechLevelModifiers, then those registered against each asset faction owns,
// in asset ID order. A tech-level floor serves its owner only, so a rival's assets
// are never asked; each asset-scoped hook decides whether its reach covers world
// and def.
func ResolveWorldTechLevel(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int {
	tl := max(baseTL, faction.WorldTechLevels[world])
	if registry == nil {
		return tl
	}
```

##### Task 11 — `internal/faction/tui/adapter/channels.go`

**Find:**
```go
	AskConfirmPermissionContest
	// AskKindCount counts the kinds above; keep it last.
```
**Replace with:**
```go
	AskConfirmPermissionContest
	AskSelectAbilityWorldTarget
	// AskKindCount counts the kinds above; keep it last.
```
**Find:**
```go
type SelectFactionTestTargetPayload struct {
```
**Replace with:**
```go
type SelectAbilityWorldTargetPayload struct {
	Holder     domain.AbilityHolder
	Effect     domain.AbilityEffectType
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
type SelectFactionTestTargetPayload struct {
```

##### Task 12 — `internal/faction/tui/adapter/action_collector.go`

In `SelectAbilityTarget`, **Find:**
```go
		chosen, _ := raw.(*domain.Asset)
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: chosen}, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("SelectAbilityTarget: target kind %q not supported yet", step.Target)
	}
```
**Replace with:**
```go
		chosen, _ := raw.(*domain.Asset)
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: chosen}, nil
	case domain.TargetWorld:
		names := make(map[string]string, len(candidates.Worlds))
		for _, worldID := range candidates.Worlds {
			if loc, ok := a.adapter.engine.World.Location(worldID); ok {
				names[worldID] = loc.Name()
			}
		}
		raw, err := a.adapter.ask(AskSelectAbilityWorldTarget, nil, SelectAbilityWorldTargetPayload{
			Holder:     holder,
			Effect:     step.Effect,
			Worlds:     candidates.Worlds,
			WorldNames: names,
		})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		chosen, _ := raw.(string)
		return domain.AbilityTarget{Kind: domain.TargetWorld, World: chosen}, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("SelectAbilityTarget: target kind %q not supported", step.Target)
	}
```

##### Task 13 — `internal/faction/tui/views/turn/overlay/ability.go`

Append:
```go
// NewSelectAbilityWorldTarget picks the world an ability acts on (Teach Planetary
// Population). The header names the holder and the effect. Answer: string (world
// ID). Esc cancels.
func NewSelectAbilityWorldTarget(holder domain.AbilityHolder, effect domain.AbilityEffectType, worlds []string, worldNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[string] {
	opts := make([]huh.Option[string], 0, len(worlds))
	for _, worldID := range worlds {
		name := worldNames[worldID]
		if name == "" {
			name = worldID
		}
		opts = append(opts, huh.NewOption(name, worldID))
	}
	return NewSelect[string]("Target which world?", fmt.Sprintf("%s · %s", holderName(holder, rb), effect), opts)
}
```

##### Task 14 — `internal/faction/tui/views/turn/execution/execution.go`

**Find:**
```go
	adapter.AskSelectFactionTestTarget: {"Ability Target", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
```
**Replace with:**
```go
	adapter.AskSelectAbilityWorldTarget: {"Ability Target", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectAbilityWorldTargetPayload)
		return overlay.NewSelectAbilityWorldTarget(p.Holder, p.Effect, p.Worlds, p.WorldNames, m.rulebook)
	}},
	adapter.AskSelectFactionTestTarget: {"Ability Target", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
```

##### Task 15 — the digest, the wire and the turn list

`internal/faction/narrative/digest/types.go`, **Find:**
```go
	GoalWorldGoverned
	GoalTagLost
)
```
**Replace with:**
```go
	GoalWorldGoverned
	GoalTagLost
	GoalTechLevelRaised
)
```
**Find:**
```go
	ToWorld   string
	TagName   string
}
```
**Replace with:**
```go
	ToWorld   string
	TagName   string
	TechLevel int
}
```
`internal/faction/narrative/digest/build.go`, **Find:**
```go
			case "world_government_changed":
```
**Replace with:**
```go
			case "world_tech_level_raised":
				var m domain.WorldTechLevelRaised
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{Kind: GoalTechLevelRaised, ToWorld: m.WorldID, TechLevel: m.TechLevel})

			case "world_government_changed":
```
`internal/faction/narrative/wire_renderer.go`, **Find:**
```go
	case digest.GoalTagLost:
		tmpl := pick(rng, goalTagLostTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	}
```
**Replace with:**
```go
	case digest.GoalTagLost:
		tmpl := pick(rng, goalTagLostTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	case digest.GoalTechLevelRaised:
		tmpl := pick(rng, goalTechLevelTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		result = strings.ReplaceAll(result, "{world}", ge.ToWorld)
		return strings.ReplaceAll(result, "{level}", strconv.Itoa(ge.TechLevel))
	}
```
Add `"strconv"` to its imports if absent.

`internal/faction/narrative/wire_templates.go`, **Find:**
```go
var goalTagLostTemplates = []string{
```
**Replace with:**
```go
var goalTechLevelTemplates = []string{
	"{world} now counts as tech level {level} for {faction}'s purchases.",
	"{faction} can now buy on {world} as if it stood at tech level {level}.",
}

var goalTagLostTemplates = []string{
```
`internal/faction/tui/views/turn/execution/format.go`, **Find:**
```go
	case domain.WorldGovernmentChanged:
		if mut.ToFactionID == "" {
```
**Replace with:**
```go
	case domain.WorldTechLevelRaised:
		return fmt.Sprintf("%s counts as TL %d for purchases", mut.WorldID, mut.TechLevel), true
	case domain.WorldGovernmentChanged:
		if mut.ToFactionID == "" {
```

##### Task 16 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.PreceptorArchive.effects.cost_discount]
amount = 1
```
**Replace with:**
```toml
[tags.PreceptorArchive.effects.cost_discount]
amount = 1

[tags.PreceptorArchive.ability]
effect = "raise_tech_level"
target = "world"
die    = "1d12"
[[tags.PreceptorArchive.ability.outcomes]]
from = 12
to   = 12
[tags.PreceptorArchive.ability.raise_tech_level]
coin_cost  = 2
tech_level = 4
```

##### Task 17 — `internal/faction/engine/testharness/harness.go`

**Find:**
```go
	// Populations overrides a world's population; any other world has stubPopulation.
	Populations map[string]int
}
```
**Replace with:**
```go
	// Populations overrides a world's population; any other world has stubPopulation.
	Populations map[string]int
	// TechLevels overrides a world's tech level; any other world is TL 5.
	TechLevels map[string]int
}
```
**Find:**
```go
func (s *StubSpatialMap) Location(id string) (spatial.Location, bool) {
	return &stubLocation{id: id, population: s.populationOf(id)}, true
}
```
**Replace with:**
```go
func (s *StubSpatialMap) Location(id string) (spatial.Location, bool) {
	return &stubLocation{id: id, population: s.populationOf(id), techLevel: s.techLevelOf(id)}, true
}

func (s *StubSpatialMap) techLevelOf(id string) int {
	if techLevel, ok := s.TechLevels[id]; ok {
		return techLevel
	}
	return 5
}
```
**Find:**
```go
		locations = append(locations, &stubLocation{id: id, population: s.populationOf(id)})
```
**Replace with:**
```go
		locations = append(locations, &stubLocation{id: id, population: s.populationOf(id), techLevel: s.techLevelOf(id)})
```
**Find:**
```go
type stubLocation struct {
	id         string
	population int
}
```
**Replace with:**
```go
type stubLocation struct {
	id         string
	population int
	techLevel  int
}
```
**Find:**
```go
func (l *stubLocation) TechLevel() int               { return 5 }
```
**Replace with:**
```go
func (l *stubLocation) TechLevel() int               { return l.techLevel }
```
Also update the `StubSpatialMap` doc line "returning TL 5 and distance 1" to
"returning TL 5 (or TechLevels) and distance 1".

##### Task 18 — tests

In `internal/faction/engine/action/actions/ability/stage_test.go`, replace every:

| Replace every | With |
|---|---|
| `resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector)` | `resolveTarget(step, factionState.Factions["f1"], domain.AbilityHolder{Asset: acting}, factionState, collector, nil)` |
| `, definition(assetStep), factionState)` | `, definition(assetStep), factionState, nil)` |
| `, definition(factionStep), factionState)` | `, definition(factionStep), factionState, nil)` |
| `, definition(relocateStep, assetStep), factionState)` | `, definition(relocateStep, assetStep), factionState, nil)` |

In `internal/faction/engine/action/actions/ability/relocate_test.go`, give the test
router a tech level per world. **Find:**
```go
type relocRouter struct{ worlds map[string]spatial.RegionHex }

func (r *relocRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &relocLoc{id: id, hex: hex}, true
}
```
**Replace with:**
```go
type relocRouter struct {
	worlds     map[string]spatial.RegionHex
	techLevels map[string]int // a world with no entry is TL 5
}

func (r *relocRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &relocLoc{id: id, hex: hex, techLevel: r.techLevelOf(id)}, true
}

func (r *relocRouter) techLevelOf(id string) int {
	if techLevel, ok := r.techLevels[id]; ok {
		return techLevel
	}
	return 5
}
```
**Find:**
```go
		worlds = append(worlds, &relocLoc{id: id, hex: hex})
```
**Replace with:**
```go
		worlds = append(worlds, &relocLoc{id: id, hex: hex, techLevel: r.techLevelOf(id)})
```
**Find:**
```go
type relocLoc struct {
	id  string
	hex spatial.RegionHex
}
```
**Replace with:**
```go
type relocLoc struct {
	id        string
	hex       spatial.RegionHex
	techLevel int
}
```
**Find:**
```go
func (l *relocLoc) TechLevel() int               { return 5 }
```
**Replace with:**
```go
func (l *relocLoc) TechLevel() int               { return l.techLevel }
```

`internal/faction/engine/action/actions/ability/raise_tech_level_test.go` (new) —
full contents:
```go
package ability

import (
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

var teachStep = domain.AbilityStep{
	Effect:         domain.EffectRaiseTechLevel,
	Target:         domain.TargetWorld,
	Roll:           &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 12}, Outcomes: []domain.AbilityOutcome{{From: 12, To: 12}}},
	RaiseTechLevel: &domain.RaiseTechLevelParams{CoinCost: 2, TechLevel: 4},
}

func TestRaiseTechLevel(t *testing.T) {
	cases := []struct {
		name string
		coin int
		roll int
		want []domain.Mutation
	}{
		{"a 12 pays and records", 2, 12, []domain.Mutation{
			domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "ability", CausedByFactionID: "f1"},
			domain.WorldTechLevelRaised{FactionID: "f1", WorldID: "near", TechLevel: 4, Cause: "ability", CausedByFactionID: "f1"},
		}},
		{"a miss pays only", 2, 11, []domain.Mutation{
			domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "ability", CausedByFactionID: "f1"},
		}},
		{"too poor to pay does nothing", 1, 12, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "f1", Coin: tc.coin}
			got, err := raiseTechLevel(StepContext{
				Faction:    faction,
				Step:       teachStep,
				Target:     domain.AbilityTarget{Kind: domain.TargetWorld, World: "near", Faction: faction},
				RollResult: tc.roll,
			})
			if err != nil {
				t.Fatalf("raiseTechLevel: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestRaiseTechLevelWorlds(t *testing.T) {
	engine := world.NewWithMap(&relocRouter{
		worlds:     map[string]spatial.RegionHex{"origin": relocHex("r1", 0), "near": relocHex("r1", 1), "mid": relocHex("r1", 2), "far": relocHex("r1", 5)},
		techLevels: map[string]int{"origin": 2, "near": 3, "mid": 4, "far": 1},
	}, slog.New(slog.DiscardHandler))
	base := func(world string) *domain.Base { return &domain.Base{Location: domain.Location{WorldID: world}} }
	faction := func(coin int) *domain.Faction {
		return &domain.Faction{
			ID: "f1", Coin: coin,
			Bases:           []*domain.Base{base("near"), base("origin"), base("mid"), base("far")},
			WorldTechLevels: map[string]int{"far": 4},
		}
	}

	if got, want := raiseTechLevelWorlds(teachStep, faction(2), engine), []string{"near", "origin"}; !slices.Equal(got, want) {
		t.Errorf("worlds = %v, want %v (mid is TL4 already, far is recorded)", got, want)
	}
	if got := raiseTechLevelWorlds(teachStep, faction(1), engine); len(got) != 0 {
		t.Errorf("worlds for a faction that cannot pay = %v, want none", got)
	}
	if got := raiseTechLevelWorlds(teachStep, faction(2), nil); len(got) != 0 {
		t.Errorf("worlds with no world engine = %v, want none", got)
	}

	tag := &domain.Tag{ID: "T-013", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{teachStep}}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction(1)}}
	if HasTargets(factionState.Factions["f1"], domain.AbilityHolder{Tag: tag}, tag.Ability, factionState, engine) {
		t.Error("HasTargets = true for a faction that cannot pay, want false")
	}
}
```

Append to `internal/faction/engine/hooks/dispatch/rules_test.go`:
```go
func TestResolveWorldTechLevel_StartsFromTheFactionsRecord(t *testing.T) {
	faction := &domain.Faction{ID: "f1", WorldTechLevels: map[string]int{"Backwater": 4}}
	def := &domain.AssetDefinition{ID: "d"}
	if got := ResolveWorldTechLevel(nil, faction, def, "Backwater", 2); got != 4 {
		t.Errorf("recorded world, no registry = %d, want 4", got)
	}
	if got := ResolveWorldTechLevel(nil, faction, def, "Tartarus", 2); got != 2 {
		t.Errorf("unrecorded world = %d, want 2", got)
	}
	if got := ResolveWorldTechLevel(nil, faction, def, "Backwater", 5); got != 5 {
		t.Errorf("a world already above the record = %d, want 5", got)
	}
	registry := hooks.NewRegistry()
	registry.RegisterWorldTechLevelModifier(hooks.FactionScope("f1"), "floor", &stubWorldTechLevelModifier{delta: 1})
	if got := ResolveWorldTechLevel(registry, faction, def, "Backwater", 2); got != 5 {
		t.Errorf("record then a hook = %d, want 5 (the hooks run on the record)", got)
	}
}
```

Append to `internal/faction/engine/mutation/mutation_test.go`:
```go
func TestMutationEngine_WorldTechLevelRaised(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))
	apply := func(level int) {
		t.Helper()
		if err := me.Apply(s, []domain.Mutation{
			domain.WorldTechLevelRaised{FactionID: "f1", WorldID: "Backwater", TechLevel: level},
		}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply() unexpected error: %v", err)
		}
	}
	apply(4)
	if got := s.Factions["f1"].WorldTechLevels["Backwater"]; got != 4 {
		t.Errorf("WorldTechLevels[Backwater] = %d, want 4", got)
	}
	apply(3)
	if got := s.Factions["f1"].WorldTechLevels["Backwater"]; got != 4 {
		t.Errorf("WorldTechLevels[Backwater] after a lower value = %d, want 4 kept", got)
	}
}
```

Append to `internal/faction/narrative/digest/digest_test.go`:
```go
func TestWorldTechLevelRaised(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha"})
	records := []domain.EventRecord{
		eventRec("f1", mutRec(t, "world_tech_level_raised", domain.WorldTechLevelRaised{
			FactionID: "f1", WorldID: "Backwater", TechLevel: 4, Cause: "ability", CausedByFactionID: "f1",
		})),
	}
	cycleDigest := buildOne(t, records, factionState, nil)
	beat := findBeat(cycleDigest, "f1")
	if beat == nil || !slices.ContainsFunc(beat.GoalEvents, func(event GoalEvent) bool {
		return event.Kind == GoalTechLevelRaised && event.ToWorld == "Backwater" && event.TechLevel == 4
	}) {
		t.Errorf("f1 beat = %+v, want a GoalTechLevelRaised event for Backwater at 4", beat)
	}
}
```

In `internal/faction/narrative/wire_renderer_test.go`, `TestRenderGoalEvent_GovernmentAndTagLoss`,
**Find:**
```go
	lost := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalTagLost, TagName: "Planetary Government"}, "Bravo", rng)
```
**Replace with:**
```go
	raised := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalTechLevelRaised, ToWorld: "Backwater", TechLevel: 4}, "Alpha", rng)
	if !strings.Contains(raised, "Alpha") || !strings.Contains(raised, "Backwater") || !strings.Contains(raised, "4") {
		t.Errorf("tech level line = %q, want the faction, the world and the level", raised)
	}
	lost := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalTagLost, TagName: "Planetary Government"}, "Bravo", rng)
```

Append to `internal/faction/rulebook/rulebook_test.go`:
```go
func TestConvertAbilityStep_RaiseTechLevel(t *testing.T) {
	const header = "\n[assets.Machine.ability]\neffect = \"raise_tech_level\"\n"
	cases := []struct {
		name string
		body string
		want string
	}{
		{"no world target", "die = \"1d12\"\n[assets.Machine.ability.raise_tech_level]\ncoin_cost = 2\ntech_level = 4\n", "raise_tech_level requires"},
		{"no block", "target = \"world\"\ndie = \"1d12\"\n", "raise_tech_level requires"},
		{"no tech level", "target = \"world\"\ndie = \"1d12\"\n[assets.Machine.ability.raise_tech_level]\ncoin_cost = 2\n", "tech_level > 0"},
		{"a coin row", "target = \"world\"\ndie = \"1d12\"\n[[assets.Machine.ability.outcomes]]\nfrom = 12\nto = 12\ncoin = 1\n[assets.Machine.ability.raise_tech_level]\ncoin_cost = 2\ntech_level = 4\n", "carry no coin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadAssetsFromTOML(t, effectTestAsset+header+tc.body)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestLoad_PreceptorArchiveTeaches(t *testing.T) {
	rb := loadShippedRulebook(t)
	want := &domain.AbilityDefinition{Steps: []domain.AbilityStep{{
		Effect:         domain.EffectRaiseTechLevel,
		Target:         domain.TargetWorld,
		Roll:           &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 12}, Outcomes: []domain.AbilityOutcome{{From: 12, To: 12}}},
		RaiseTechLevel: &domain.RaiseTechLevelParams{CoinCost: 2, TechLevel: 4},
	}}}
	if got := rb.Tags["SWN-T-013"].Ability; !reflect.DeepEqual(got, want) {
		t.Errorf("Preceptor Archive ability = %+v, want %+v", got, want)
	}
}
```

Append to `internal/faction/engine/testharness/scenarios/tags_test.go`:
```go
// --- scenario N+9: Preceptor Archive — Teach Planetary Population ---

// TestPreceptorArchive_TeachPlanetaryPopulation: alpha holds Preceptor Archive and
// a Base on Backwater (TL 2); Tartarus, its homeworld, has no Base. Use Asset
// Ability offers the tag alone, and the world target offers Backwater alone. The
// action pays 2 Coin either way; on a 12, Backwater counts as TL4 for alpha.
func TestPreceptorArchive_TeachPlanetaryPopulation(t *testing.T) {
	cases := []struct {
		name   string
		roll   int
		wantTL int
	}{
		{"a 12 raises the world", 12, 4},
		{"an 11 pays and raises nothing", 11, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			h.SpatialMap.TechLevels = map[string]int{"Backwater": 2}
			alpha := h.AddFaction("alpha", "Tartarus", 2, 3, 2)
			alpha.Tags = []string{"SWN-T-013"}
			alpha.Coin = 4
			testharness.AddBase(alpha, "Backwater", 2)
			h.Engine.Rand = &testharness.FixedRoller{Values: []int{tc.roll}}
			var offered []domain.AbilityHolder
			var worlds []string
			h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Use Asset Ability")
			h.Collector.SelectAbilityHoldersFn = func(_ *domain.Faction, candidates []domain.AbilityHolder, _ *rulebook.Rulebook) ([]domain.AbilityHolder, error) {
				offered = candidates
				return candidates, nil
			}
			h.Collector.SelectAbilityTargetFn = func(_ domain.AbilityHolder, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
				worlds = candidates.Worlds
				return domain.AbilityTarget{Kind: domain.TargetWorld, World: "Backwater"}, nil
			}

			runOrderedCycle(t, h, alpha.ID)

			testharness.CheckStep(t, "the tag is the only holder offered", len(offered) == 1 && offered[0].Tag != nil && offered[0].Tag.ID == "SWN-T-013",
				fmt.Sprintf("offered = %+v", offered))
			testharness.CheckStep(t, "Backwater is the only world offered", len(worlds) == 1 && worlds[0] == "Backwater",
				fmt.Sprintf("worlds = %v", worlds))
			testharness.CheckStep(t, "the record matches the roll", alpha.WorldTechLevels["Backwater"] == tc.wantTL,
				fmt.Sprintf("WorldTechLevels[Backwater] = %d, want %d", alpha.WorldTechLevels["Backwater"], tc.wantTL))
			records := testharness.ReadHistory(t, h.Paths.HistoryPath)
			_, paid := testharness.FindMutationByTypeAndCause(records, "coin_delta", "ability")
			testharness.CheckStep(t, "the 2 Coin is paid", paid, "no coin_delta(cause=ability) in history")
		})
	}
}
```

##### Commit message

```
feat(E5): Teach Planetary Population

- raise_tech_level ability effect: target world, a die, outcome rows and
  coin_cost / tech_level from the entry; always pays, records on a match
  (Decision 3)
- world targets: Base worlds below the entry's TL with no record, none when
  the faction cannot pay (Decision 4); TUI world picker (Decision 6)
- Faction.WorldTechLevels and WorldTechLevelRaised; ResolveWorldTechLevel
  starts from the record (Decision 5)
- digest, wire and turn-list lines for the record
- Preceptor Archive carries the ability in tags.toml
```

### Phase 3 — Gengineered Slaves and ability contests

**In plain words:** only an Eugenics Cult faction can buy, refit into or start with
Gengineered Slaves, and its Slaves get the tag's once-per-turn die. Informers and
Seductress roll as attacks, so a Fanatical defender loses their ties.

#### Commit 4 — `feat(E5): Gengineered Slaves and requires_tag`

##### Task 1 — `internal/faction/domain/asset.go`, the field and the check

**Find:**
```go
	// PurchasePermission asks the government to raise this asset, though moving it
	// in needs none (Capital Fleet).
	PurchasePermission bool
}
```
**Replace with:**
```go
	// PurchasePermission asks the government to raise this asset, though moving it
	// in needs none (Capital Fleet).
	PurchasePermission bool
	// RequiresTag names the tag a faction must hold to buy, refit into or start
	// with this asset (Gengineered Slaves, D-13); empty for most assets.
	RequiresTag string
}

// TagRequirementMet reports whether a holder of tagIDs meets def's RequiresTag.
func (def *AssetDefinition) TagRequirementMet(tagIDs []string) bool {
	return def.RequiresTag == "" || slices.Contains(tagIDs, def.RequiresTag)
}
```
Add `"slices"` to the imports of `asset.go` if absent.

In `CanPurchase`, **Find:**
```go
// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, the def's required tech level is within TechCap, and
// the world's Population reaches PopulationFloor. Government permission is asked after the
// choice, not here (D-35).
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
```
**Replace with:**
```go
// CanPurchase reports whether the faction may buy def under terms: it holds any tag def
// requires, can afford Cost, meets the def's minimum category rating, the def's required
// tech level is within TechCap, and the world's Population reaches PopulationFloor.
// Government permission is asked after the choice, not here (D-35).
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case !def.TagRequirementMet(faction.Tags):
		return false
	case faction.Coin < terms.Cost:
		return false
```

##### Task 2 — `internal/faction/domain/effect.go`, the filter

**Find:**
```go
	FilterDetection         EffectFilterName = "detection"
)
```
**Replace with:**
```go
	FilterDetection         EffectFilterName = "detection"
	FilterAssetDefinition   EffectFilterName = "asset_definition"
)
```
**Find:**
```go
	Goal              string
	Detection         bool
}
```
**Replace with:**
```go
	Goal              string
	Detection         bool
	// AssetDefinition names the rolling asset's definition ID (the Eugenics Cult
	// die on Gengineered Slaves).
	AssetDefinition string
}
```
**Find:**
```go
	if filters.Detection {
		present = append(present, FilterDetection)
	}
	return present
```
**Replace with:**
```go
	if filters.Detection {
		present = append(present, FilterDetection)
	}
	if filters.AssetDefinition != "" {
		present = append(present, FilterAssetDefinition)
	}
	return present
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	Goal              string   `toml:"goal"`
	Detection         bool     `toml:"detection"`
}
```
**Replace with:**
```go
	Goal              string   `toml:"goal"`
	Detection         bool     `toml:"detection"`
	AssetDefinition   string   `toml:"asset_definition"`
}
```
**Find:**
```go
		Goal: record.Goal, Detection: record.Detection,
	}
```
**Replace with:**
```go
		Goal: record.Goal, Detection: record.Detection, AssetDefinition: record.AssetDefinition,
	}
```

##### Task 4 — `internal/faction/rulebook/rulebook.go`, the asset field and the references

**Find:**
```go
	Quality            bool             `toml:"quality"`
	PurchasePermission bool             `toml:"purchase_permission"`
}
```
**Replace with:**
```go
	Quality            bool             `toml:"quality"`
	PurchasePermission bool             `toml:"purchase_permission"`
	RequiresTag        string           `toml:"requires_tag"`
}
```
**Find:**
```go
		Quality:            r.Quality,
		PurchasePermission: r.PurchasePermission,
	}, nil
```
**Replace with:**
```go
		Quality:            r.Quality,
		PurchasePermission: r.PurchasePermission,
		RequiresTag:        r.RequiresTag,
	}, nil
```
**Find:**
```go
	if err := checkGoalReferences(assets, tags, goals); err != nil {
		return nil, fmt.Errorf("checking goal references: %w", err)
	}
```
**Replace with:**
```go
	if err := checkReferences(assets, tags, goals); err != nil {
		return nil, fmt.Errorf("checking references: %w", err)
	}
```
Replace the whole `checkGoalReferences` function (comment included) with:
```go
// checkReferences refuses a data-to-data reference to a record the rulebook does
// not have, checked once here: an [[effects]] entry's goal filter (D-5) or
// asset_definition filter (Effort 5 Decision 9), and an asset's requires_tag
// (D-13).
func checkReferences(assets map[string]*domain.AssetDefinition, tags map[string]*domain.Tag, goals map[string]*domain.Goal) error {
	check := func(holder string, entries []domain.EffectDefinition) error {
		for i, entry := range entries {
			if goal := entry.Filters.Goal; goal != "" && goals[goal] == nil {
				return fmt.Errorf("%s effects[%d]: unknown goal %q", holder, i, goal)
			}
			if definition := entry.Filters.AssetDefinition; definition != "" && assets[definition] == nil {
				return fmt.Errorf("%s effects[%d]: unknown asset %q", holder, i, definition)
			}
		}
		return nil
	}
	for _, id := range slices.Sorted(maps.Keys(assets)) {
		if err := check(fmt.Sprintf("asset %q", id), assets[id].Effects); err != nil {
			return err
		}
		if tag := assets[id].RequiresTag; tag != "" && tags[tag] == nil {
			return fmt.Errorf("asset %q: requires_tag names unknown tag %q", id, tag)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(tags)) {
		if err := check(fmt.Sprintf("tag %q", id), tags[id].Effects); err != nil {
			return err
		}
	}
	return nil
}
```

##### Task 5 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
```
**Replace with:**
```go
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
	if filters.AssetDefinition != "" && (ctx.Asset == nil || ctx.Asset.DefinitionID != filters.AssetDefinition) {
		return false
	}
```

##### Task 6 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal, domain.FilterDetection,
			},
```
**Replace with:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal, domain.FilterDetection,
				domain.FilterAssetDefinition,
			},
```

##### Task 7 — `internal/faction/engine/action/actions/refit_asset.go`

**Find:**
```go
// validRefitReplacements returns definitions the faction could refit oldAsset
// into: same category, different definition, attribute meets MinRating, and
// faction can afford the cost delta.
```
**Replace with:**
```go
// validRefitReplacements returns definitions the faction could refit oldAsset
// into: same category, different definition, attribute meets MinRating, any
// required tag held, and faction can afford the cost delta.
```
**Find:**
```go
		if statScore(faction, def.Category) < def.MinRating {
			continue
		}
		delta := max(0, def.Cost-oldDef.Cost)
```
**Replace with:**
```go
		if statScore(faction, def.Category) < def.MinRating {
			continue
		}
		if !def.TagRequirementMet(faction.Tags) {
			continue
		}
		delta := max(0, def.Cost-oldDef.Cost)
```

##### Task 8 — `internal/faction/tui/views/manage/wizard/wizard.go`

In `buildAssetsForm`, **Find:**
```go
	allDefs := make([]*domain.AssetDefinition, 0, len(rb.Assets))
	for _, def := range rb.Assets {
		if def.Quality {
			continue
		}
		allDefs = append(allDefs, def)
	}
```
**Replace with:**
```go
	pickedTags := tagIDs(data.selectedTags)
	allDefs := make([]*domain.AssetDefinition, 0, len(rb.Assets))
	for _, def := range rb.Assets {
		if def.Quality || !def.TagRequirementMet(pickedTags) {
			continue
		}
		allDefs = append(allDefs, def)
	}
```

##### Task 9 — `rulebooks/swn/assets/force_assets.toml`

**Find:**
```toml
[assets.MilitiaUnit.attack]
attacker_stat = "Force"
defender_stat = "Force"
damage = "1d6"
```
**Replace with:**
```toml
[assets.MilitiaUnit.attack]
attacker_stat = "Force"
defender_stat = "Force"
damage = "1d6"

[assets.GengineeredSlaves]
id = "SWN-F1-004"
name = "Gengineered Slaves"
category = "Force"
min_rating = 1
hp = 6
cost = 2
maintenance = 0
tech_level = 4
type = "Military Unit"
flags = []
counter = "1d4"
description = "Genetically engineered servitors bred by a eugenics cult for labor and war. Only a faction with the Eugenics Cult tag can buy them."
requires_tag = "SWN-T-003"

[assets.GengineeredSlaves.attack]
attacker_stat = "Force"
defender_stat = "Force"
damage = "1d6"
```

##### Task 10 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Can purchase the Gengineered Slaves asset (Force 1, 6HP, 2 cost, TL4, Attack Force vs Force 1d6, Counter 1d4). Once per turn, roll an extra d10 on an attack or defense by a Gengineered Slaves asset."
```
**Replace with:**
```toml
effect = "Can purchase the Gengineered Slaves asset (Force 1, 6HP, 2 cost, TL4, Attack Force vs Force 1d6, Counter 1d4). Once per turn, roll an extra d10 on an attack or defense by a Gengineered Slaves asset."

[[tags.EugenicsCult.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.EugenicsCult.effects.filters]
owner            = "own"
scope            = "faction"
phase            = ["attack", "defense"]
asset_definition = "SWN-F1-004"
[tags.EugenicsCult.effects.gates]
once_per = "owner_faction"
[tags.EugenicsCult.effects.bonus_die]
sides = 10
count = 1
```

##### Task 11 — `internal/faction/domain/asset_test.go`

Append:
```go
func TestCanPurchase_RequiresTag(t *testing.T) {
	def := &domain.AssetDefinition{ID: "slaves", Category: domain.StatForce, Cost: 2, TechLevel: 4, RequiresTag: "T-003"}
	terms := domain.PurchaseTerms{Cost: 2, TechCap: 5}
	if domain.CanPurchase(&domain.Faction{ID: "f1", Coin: 5, Force: 1}, def, terms) {
		t.Error("CanPurchase = true without the tag, want false")
	}
	if !domain.CanPurchase(&domain.Faction{ID: "f1", Coin: 5, Force: 1, Tags: []string{"T-003"}}, def, terms) {
		t.Error("CanPurchase = false with the tag, want true")
	}
}
```

##### Task 12 — rulebook tests

In `internal/faction/rulebook/rulebook_test.go`, replace the whole
`TestCheckGoalReferences` with:
```go
func TestCheckReferences(t *testing.T) {
	goals := map[string]*domain.Goal{"G-004": {ID: "G-004"}}
	tags := map[string]*domain.Tag{"T-1": {ID: "T-1"}}
	assets := map[string]*domain.AssetDefinition{"A-1": {ID: "A-1"}}
	naming := func(filters domain.EffectFilters) []domain.EffectDefinition {
		return []domain.EffectDefinition{{Filters: filters}}
	}
	withTag := func(entries []domain.EffectDefinition) map[string]*domain.Tag {
		return map[string]*domain.Tag{"T-1": {ID: "T-1", Effects: entries}}
	}

	if err := checkReferences(assets, withTag(naming(domain.EffectFilters{Goal: "G-004", AssetDefinition: "A-1"})), goals); err != nil {
		t.Errorf("known references: err = %v, want nil", err)
	}
	cases := []struct {
		name   string
		assets map[string]*domain.AssetDefinition
		tags   map[string]*domain.Tag
		want   string
	}{
		{"unknown goal on a tag", assets, withTag(naming(domain.EffectFilters{Goal: "G-404"})), `tag "T-1" effects[0]: unknown goal "G-404"`},
		{"unknown goal on an asset", map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", Effects: naming(domain.EffectFilters{Goal: "G-404"})}}, tags, `asset "A-1" effects[0]: unknown goal "G-404"`},
		{"unknown asset definition", assets, withTag(naming(domain.EffectFilters{AssetDefinition: "A-404"})), `tag "T-1" effects[0]: unknown asset "A-404"`},
		{"unknown required tag", map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", RequiresTag: "T-404"}}, tags, `asset "A-1": requires_tag names unknown tag "T-404"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkReferences(tc.assets, tc.tags, goals)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestLoad_EugenicsCult(t *testing.T) {
	rb := loadShippedRulebook(t)
	slaves := rb.Assets["SWN-F1-004"]
	if slaves == nil || slaves.RequiresTag != "SWN-T-003" || slaves.Type != domain.TypeMilitaryUnit {
		t.Fatalf("Gengineered Slaves = %+v, want a Military Unit that requires SWN-T-003", slaves)
	}
	want := []domain.EffectDefinition{{
		Trigger: domain.TriggerOwnRoll,
		Effect:  domain.SpecialBonusDie,
		Filters: domain.EffectFilters{
			Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
			Phases:          []domain.EffectRollPhase{domain.RollPhaseAttack, domain.RollPhaseDefense},
			AssetDefinition: "SWN-F1-004",
		},
		Gates:    domain.EffectGates{OncePer: domain.OncePerOwnerFaction},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}}
	if got := rb.Tags["SWN-T-003"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Eugenics Cult effects = %+v, want %+v", got, want)
	}
}
```

##### Task 13 — effect and action tests

In `internal/faction/engine/effect/effects/roll_match_test.go`, **Find:**
```go
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```
**Replace with:**
```go
		{"asset definition match", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha, Asset: primitive}, true},
		{"asset definition mismatch", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha, Asset: advanced}, false},
		{"asset definition, no rolling asset", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```
**Find:**
```go
	rivalHomeworld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeHomeworld}
```
**Replace with:**
```go
	rivalHomeworld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeHomeworld}
	slavesRoll := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetDefinition: "tl0"}
```

Append to `internal/faction/engine/action/actions/refit_asset_test.go`:
```go
func TestValidRefitReplacements_RequiresTag(t *testing.T) {
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"guards": {ID: "guards", Category: domain.StatForce, MinRating: 1, Cost: 2},
		"slaves": {ID: "slaves", Category: domain.StatForce, MinRating: 1, Cost: 2, RequiresTag: "T-003"},
	}}
	asset := &domain.Asset{ID: "a1", DefinitionID: "guards", Location: domain.Location{WorldID: "Tartarus"}}
	without := &domain.Faction{ID: "f1", Force: 1, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
	if got := validRefitReplacements(without, asset, rulebook); len(got) != 0 {
		t.Errorf("replacements without the tag = %v, want none", got)
	}
	with := &domain.Faction{ID: "f1", Force: 1, Coin: 5, Tags: []string{"T-003"}, Assets: map[string]*domain.Asset{"a1": asset}}
	if got := validRefitReplacements(with, asset, rulebook); len(got) != 1 || got[0].ID != "slaves" {
		t.Errorf("replacements with the tag = %v, want slaves", got)
	}
}
```
Add the `rulebook` import to `refit_asset_test.go` if absent.

##### Task 14 — `internal/faction/engine/testharness/scenarios/tags_test.go`

Append:
```go
// --- scenario N+10: Eugenics Cult — Gengineered Slaves and their die ---

// TestEugenicsCult_OnlyTheTagBuysGengineeredSlaves: Buy Asset offers Gengineered
// Slaves on Tartarus (TL 5) to a faction with Eugenics Cult, and not to one
// without it.
func TestEugenicsCult_OnlyTheTagBuysGengineeredSlaves(t *testing.T) {
	cases := []struct {
		name  string
		tags  []string
		offer bool
	}{
		{"no tag", nil, false},
		{"Eugenics Cult", []string{"SWN-T-003"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			alpha := h.AddFaction("alpha", "Tartarus", 2, 2, 2)
			alpha.Tags = tc.tags
			alpha.Coin = 4
			offered := false
			h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Buy Asset")
			h.Collector.SelectBuyOrderFn = func(purchasable map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
				var fallback *domain.AssetDefinition
				for _, def := range purchasable["Tartarus"] {
					if def.ID == "SWN-F1-004" {
						offered = true
					}
					if def.ID == testharness.DefSecurityPersonnel {
						fallback = def
					}
				}
				return action.BuyOrder{World: "Tartarus", Definition: fallback}, nil
			}

			runOrderedCycle(t, h, alpha.ID)

			testharness.CheckStep(t, "Gengineered Slaves offered", offered == tc.offer,
				fmt.Sprintf("offered = %v, want %v", offered, tc.offer))
		})
	}
}
```
Append to `internal/faction/engine/effect/effects/bonus_die_test.go`:
```go
func TestBonusDie_EugenicsCultDieOnlyOnTheNamedAsset(t *testing.T) {
	slaves := &domain.Asset{ID: "s1", OwnerID: "alpha", DefinitionID: "SWN-F1-004"}
	guards := &domain.Asset{ID: "g1", OwnerID: "alpha", DefinitionID: "SWN-F1-001"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-003"}, Assets: map[string]*domain.Asset{"s1": slaves, "g1": guards}}
	beta := &domain.Faction{ID: "beta"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
	entry := bonusDieEntry(domain.EffectFilters{
		Owner: domain.OwnerOwn, Scope: domain.ScopeFaction,
		Phases:          []domain.EffectRollPhase{domain.RollPhaseAttack, domain.RollPhaseDefense},
		AssetDefinition: "SWN-F1-004",
	}, 1)
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialBonusDie].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-003", Name: "Eugenics Cult"}}, Entry: entry, Registry: registry,
	})
	hook := registry.RollModifiersAt(hooks.FactionScope("alpha"))[0].Hook
	roll := func(asset *domain.Asset, phase hooks.RollPhase) hooks.RollContext {
		return hooks.RollContext{Phase: phase, Actor: alpha, Opponent: beta, Attribute: "Force", Asset: asset, World: "Krylos"}
	}

	if offers := hook.OfferModifiers(roll(slaves, hooks.PhaseAttack), factionState, nil); len(offers) != 1 || !offers[0].FromTag {
		t.Errorf("Slaves attack offers = %+v, want one tag die", offers)
	}
	if offers := hook.OfferModifiers(roll(slaves, hooks.PhaseDefense), factionState, nil); len(offers) != 1 {
		t.Errorf("Slaves defense offers = %+v, want one", offers)
	}
	if offers := hook.OfferModifiers(roll(guards, hooks.PhaseAttack), factionState, nil); len(offers) != 0 {
		t.Errorf("another asset's attack offers = %+v, want none", offers)
	}
}
```

##### Commit message

```
feat(E5): Gengineered Slaves and requires_tag

- AssetDefinition.RequiresTag and TagRequirementMet; checked in CanPurchase
  (Buy Asset, Pretech Logistics), Refit and the wizard's starting picker
  (Decision 7)
- asset_definition roll filter on bonus_die (Decision 9)
- checkGoalReferences becomes checkReferences: goal and asset_definition
  filters, and requires_tag
- Gengineered Slaves (SWN-F1-004, Military Unit) and the Eugenics Cult die in
  the SWN data (Decision 8)
```

#### Commit 5 — `feat(E5): ability contests roll as the book words them`

##### Task 1 — `internal/faction/domain/asset.go`, `Contest`

**Find:**
```go
type Contest struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
}
```
**Replace with:**
```go
// Contest is a stat contest: an ability's gate or an effect's. Roll names the roll
// phases it uses, as the book words it (ContestRoll).
type Contest struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
	Roll         ContestRoll
}
```

##### Task 2 — `internal/faction/domain/effect.go`

**Find:**
```go
// ContestRoll names the roll phases an effect contest rolls with, as the book
// words it: an attack and its defense (Tripwire Cells' "Cunning vs. Cunning
// attack"), or a faction test and its contested roll (Panopticon Matrix's and
// Lobbyists' "test"), as the A-side gate rolls.
```
**Replace with:**
```go
// ContestRoll names the roll phases a contest rolls with, as the book words it: an
// attack and its defense (Tripwire Cells', Informers' and Seductress' "Cunning vs.
// Cunning attack"), or a faction test and its contested roll (Panopticon Matrix's,
// Lobbyists' and Marketers' "test"). An attack contest's tie asks the tie
// resolvers; a test's never does.
```
**Find:**
```go
type EffectContest struct {
	Contest
	Roller ContestRoller
	Roll   ContestRoll
}
```
**Replace with:**
```go
type EffectContest struct {
	Contest
	Roller ContestRoller
}
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find:**
```go
		gates.Contest = &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender},
			Roller:  roller,
			Roll:    roll,
		}
```
**Replace with:**
```go
		gates.Contest = &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender, Roll: roll},
			Roller:  roller,
		}
```

##### Task 4 — `internal/faction/rulebook/rulebook.go`, the ability step

**Find:**
```go
	Effect       string          `toml:"effect"`
	Target       string          `toml:"target"`
	AttackerStat string          `toml:"attacker_stat"`
```
**Replace with:**
```go
	Effect       string          `toml:"effect"`
	Target       string          `toml:"target"`
	Roll         string          `toml:"roll"`
	AttackerStat string          `toml:"attacker_stat"`
```
In `convertAbility`, **Find:**
```go
		if r.Effect != "" || r.AttackerStat != "" || r.DefenderStat != "" || r.Die != "" {
```
**Replace with:**
```go
		if r.Effect != "" || r.Roll != "" || r.AttackerStat != "" || r.DefenderStat != "" || r.Die != "" {
```
In `convertAbilityStep`, **Find:**
```go
		step.Contest = &domain.Contest{AttackerStat: attacker, DefenderStat: defender}
	}
```
**Replace with:**
```go
		roll, err := toEnum("contest roll", record.Roll, contestRolls)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		step.Contest = &domain.Contest{AttackerStat: attacker, DefenderStat: defender, Roll: roll}
	} else if record.Roll != "" {
		return domain.AbilityStep{}, fmt.Errorf("roll %q needs a contest (attacker_stat and defender_stat)", record.Roll)
	}
```

##### Task 5 — `internal/faction/engine/action/actions/ability/stage.go`, `passesGate`

**Find:**
```go
	worldID := asset.Location.WorldID
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	attack := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:             hooks.PhaseFactionTest,
```
**Replace with:**
```go
	worldID := asset.Location.WorldID
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	attackPhase, defensePhase := hooks.PhaseFactionTest, hooks.PhaseContested
	if step.Contest.Roll == domain.ContestRollAttack {
		attackPhase, defensePhase = hooks.PhaseAttack, hooks.PhaseDefense
	}
	attack := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:             attackPhase,
```
**Find:**
```go
	defense := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:             hooks.PhaseContested,
```
**Replace with:**
```go
	defense := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:             defensePhase,
```
**Find:**
```go
		registry, collector, roller, defender, factionState, rulebook,
	)
	return attack.Sum > defense.Sum, nil
}
```
**Replace with:**
```go
		registry, collector, roller, defender, factionState, rulebook,
	)
	if attack.Sum == defense.Sum && step.Contest.Roll == domain.ContestRollAttack {
		tieCtx := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: faction, Opponent: defender, Asset: asset, World: worldID, Action: actionName}
		return dispatch.ResolveTie(registry, tieCtx, factionState) == hooks.TieAttackerWins, nil
	}
	return attack.Sum > defense.Sum, nil
}
```
Update the `passesGate` doc comment:

**Find:**
```go
// passesGate rolls the step's contest through the roll hooks. faction is the real
// acting faction, so a budget a hook spends persists past this step.
```
**Replace with:**
```go
// passesGate rolls the step's contest through the roll hooks: an attack and its
// defense for roll "attack", a faction test and its contested roll for "test".
// Ties go to the defender, unless an attack's tie resolver gives the tie to the
// attacker (Fanatical; Effort 5 Decision 10). faction is the real acting faction,
// so a budget a hook spends persists past this step.
```

##### Task 6 — `rulebooks/swn/assets/cunning_assets.toml` and `wealth_assets.toml`

`cunning_assets.toml`, **Find:**
```toml
[assets.Informers.ability]
attacker_stat = "Cunning"
defender_stat = "Cunning"
```
**Replace with:**
```toml
[assets.Informers.ability]
roll = "attack"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```
**Find:**
```toml
[[assets.Seductress.ability.steps]]
effect = "reveal_stealth"
target = "faction"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```
**Replace with:**
```toml
[[assets.Seductress.ability.steps]]
effect = "reveal_stealth"
target = "faction"
roll = "attack"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```
`wealth_assets.toml`, **Find:**
```toml
[assets.Marketers.ability]
attacker_stat = "Cunning"
defender_stat = "Wealth"
```
**Replace with:**
```toml
[assets.Marketers.ability]
roll = "test"
attacker_stat = "Cunning"
defender_stat = "Wealth"
```

##### Task 7 — tests that build an `EffectContest`

`Roll` moves into `Contest`. Each edit below:

`internal/faction/engine/effect/effects/contest_test.go`, **Find:**
```go
	return domain.EffectContest{
		Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
		Roller:  roller,
		Roll:    roll,
	}
```
**Replace with:**
```go
	return domain.EffectContest{
		Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning, Roll: roll},
		Roller:  roller,
	}
```
Same file, **Find:**
```go
			contest := domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
				Roller:  domain.ContestRollerSelf,
				Roll:    tc.roll,
			}
```
**Replace with:**
```go
			contest := domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth, Roll: tc.roll},
				Roller:  domain.ContestRollerSelf,
			}
```
`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_LobbyistsWithdrawPermission`, **Find:**
```go
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
			Roll:    domain.ContestRollTest,
```
**Replace with:**
```go
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning, Roll: domain.ContestRollTest},
			Roller:  domain.ContestRollerSelf,
```
`internal/faction/rulebook/effects_test.go`, **Find:**
```go
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
					Roller:  domain.ContestRollerTarget,
					Roll:    domain.ContestRollTest,
```
**Replace with:**
```go
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth, Roll: domain.ContestRollTest},
					Roller:  domain.ContestRollerTarget,
```

##### Task 8 — ability contest fixtures in `internal/faction/rulebook/rulebook_test.go`

In `TestConvertAbilityStep_DisableAssetNeedsAnAssetTarget`, **Find:**
```go
[assets.Machine.ability]
attacker_stat = "Cunning"
defender_stat = "Wealth"
effect = "disable_asset"
target = "faction"
```
**Replace with:**
```go
[assets.Machine.ability]
roll = "test"
attacker_stat = "Cunning"
defender_stat = "Wealth"
effect = "disable_asset"
target = "faction"
```
In `TestConvertAbilityStep_LevyCoinRefusesAContest`, **Find:**
```go
[assets.Machine.ability]
attacker_stat = "Wealth"
defender_stat = "Wealth"
effect = "levy_coin"
```
**Replace with:**
```go
[assets.Machine.ability]
roll = "test"
attacker_stat = "Wealth"
defender_stat = "Wealth"
effect = "levy_coin"
```

##### Task 9 — new tests

Append to `internal/faction/rulebook/rulebook_test.go`:
```go
func TestConvertAbilityStep_ContestRoll(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"a contest with no roll", "attacker_stat = \"Cunning\"\ndefender_stat = \"Cunning\"\neffect = \"reveal_stealth\"\ntarget = \"faction\"\n", `unknown contest roll: ""`},
		{"a roll with no contest", "roll = \"attack\"\neffect = \"coin_delta\"\ndie = \"1d6\"\n", "needs a contest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadAssetsFromTOML(t, effectTestAsset+"\n[assets.Machine.ability]\n"+tc.body)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}

	rb := loadShippedRulebook(t)
	rolls := map[string]domain.ContestRoll{"SWN-C1-002": domain.ContestRollAttack, "SWN-W5-001": domain.ContestRollTest}
	for id, want := range rolls {
		if got := rb.Assets[id].Ability.Steps[0].Contest.Roll; got != want {
			t.Errorf("%s contest roll = %q, want %q", id, got, want)
		}
	}
	if got := rb.Assets["SWN-C2-004"].Ability.Steps[1].Contest.Roll; got != domain.ContestRollAttack {
		t.Errorf("Seductress contest roll = %q, want attack", got)
	}
}
```
In `TestLoadTags_Ability` (Commit 2 Task 8), add a refused case. **Find:**
```go
		{"an effect that reads the asset", "effect = \"levy_coin\"\ntarget = \"faction\"\n", `cannot carry effect "levy_coin"`},
```
**Replace with:**
```go
		{"an effect that reads the asset", "effect = \"levy_coin\"\ntarget = \"faction\"\n", `cannot carry effect "levy_coin"`},
		{"a contest", "effect = \"purchase_discount\"\nroll = \"test\"\nattacker_stat = \"Cunning\"\ndefender_stat = \"Cunning\"\n", "cannot hold a contest"},
```

Append to `internal/faction/engine/action/actions/ability/detection_test.go`:
```go
// tieToAttacker is a Cat 5 hook that gives every tie to the attacker (Fanatical's
// lose_ties on the defender).
type tieToAttacker struct{}

func (tieToAttacker) ResolveTie(hooks.RollContext, *state.FactionState) hooks.TieOutcome {
	return hooks.TieAttackerWins
}

func TestPassesGate_AnAttackContestAsksTheTieResolvers(t *testing.T) {
	phases := &phaseRecorder{}
	cases := []struct {
		name       string
		roll       domain.ContestRoll
		wantPass   bool
		wantPhases []hooks.RollPhase
	}{
		{"attack: the tie goes to the attacker", domain.ContestRollAttack, true, []hooks.RollPhase{hooks.PhaseAttack, hooks.PhaseDefense}},
		{"test: the tie stays with the defender", domain.ContestRollTest, false, []hooks.RollPhase{hooks.PhaseFactionTest, hooks.PhaseContested}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			phases.seen = nil
			informers := &domain.Asset{ID: "inf", Location: domain.Location{WorldID: "Krylos"}}
			alpha := &domain.Faction{ID: "alpha", Cunning: 2, Assets: map[string]*domain.Asset{"inf": informers}}
			beta := &domain.Faction{ID: "beta", Cunning: 2}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "phases", phases)
			registry.RegisterTieResolver(hooks.FactionScope("beta"), "lose_ties", tieToAttacker{})
			step := domain.AbilityStep{
				Effect:  domain.EffectRevealStealth,
				Target:  domain.TargetFaction,
				Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning, Roll: tc.roll},
			}
			target := domain.AbilityTarget{Kind: domain.TargetFaction, Faction: beta}

			passed, err := passesGate(step, alpha, informers, target, registry, nil, &seqRoller{values: []int{5, 5}}, factionState, nil)
			if err != nil {
				t.Fatalf("passesGate: %v", err)
			}
			if passed != tc.wantPass {
				t.Errorf("passed = %v, want %v", passed, tc.wantPass)
			}
			if !reflect.DeepEqual(phases.seen, tc.wantPhases) {
				t.Errorf("phases = %v, want %v", phases.seen, tc.wantPhases)
			}
		})
	}
}

// phaseRecorder is a Cat 1 hook that records each roll's phase and offers nothing.
type phaseRecorder struct {
	seen []hooks.RollPhase
}

func (recorder *phaseRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.seen = append(recorder.seen, ctx.Phase)
	return nil
}
```

##### Commit message

```
feat(E5): ability contests roll as the book words them

- Roll moves from EffectContest into domain.Contest, one field for both
- an ability contest names roll = "attack" or "test", required
- an attack contest rolls attack and defense phases, and its tie asks the
  tie resolvers, so Fanatical loses Informers' and Seductress' ties
  (Decision 10)
- Informers and Seductress roll attack, Marketers test
```
