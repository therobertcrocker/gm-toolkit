# A-flag Ability Registry (A.001.2) — Effort 3 Plan: Relocate primitive

## Context / Goal

Build the `relocate` movement effect on top of Effort 1's pipeline: a world-engine
instant-relocate primitive, the world engine + collector threaded into the ability
dispatch path, and a `relocate` handler covering both **self** (Seductress moves
itself) and **other** (Covert Shipping / Covert Transit Net move friendly Special
Forces assets) scopes. Authoring the three assets makes Seductress the pipeline's
first **live two-step ability** (relocate then reveal) and the first real exercise of
Effort 1's halt-on-failure invariant.

- Overview plan: [`a-flag-ability-registry-plan.md`](./a-flag-ability-registry-plan.md)
  (Effort 3 row; Decisions 3–5; the overview's Commit outline is scoping guidance,
  superseded by this file's Work Breakdown).
- Effort 1 plan: [`-effort-1-plan.md`](./a-flag-ability-registry-effort-1-plan.md)
  (the `StepContext`/`EffectHandler` contract, halt-on-failure loop with the ungated-
  before-gated invariant, the deferred `Relocate *RelocateParams` sub-struct this
  effort first consumes).
- Architecture: [world & movement](../../architecture/engine/world-movement.md) (the
  `HexRouter` seam, `AssetMoved`, the instant-vs-drift-order split), [ability-catalog](../../architecture/engine/ability-catalog.md)
  (the relocate-other cluster).

## Decisions Ratified in Planning

1. **Threaded asset view for intra-ability location** (Robert, this session). Step
   mutations apply only after all steps run, so a later step otherwise reads the
   acting asset's *pre-move* location — Seductress's reveal would fire at the old
   world. Resolution: after each step, `Dispatch` scans the returned mutations for an
   `AssetMoved` on the acting asset and builds subsequent `StepContext`s with a
   location-updated **copy** of the asset (`applyAssetMovedView`). No shared-state
   write (engines never write state), the error path stays clean, and the real
   `AssetMoved` mutation still applies later. Rejected: an in-place `ctx.Asset.Location`
   write (partial-state smell on the error path) and deferring the combined semantics
   (leaves Seductress not end-to-end).

2. **Relocate owns all selection in its handler; the generic target stage is bypassed**
   (reversal of Effort 1 Decision 2 / overview Decision 7 for this effort). Because
   Transit Net moves **multiple** assets and every moved asset needs a paired
   destination, relocate does not fit the single-target `SelectAbilityTarget` stage.
   Relocate steps carry `Target = none`; a `RelocateParams.Scope` (`self`|`other`)
   drives the handler, which gathers candidates and prompts a dedicated order-slice
   collector method (`SelectRelocateOrders`). The `TargetAsset` case of
   `SelectAbilityTarget` stays unwired (still errors) — Effort 4 wires it for
   `coin_or_disable`'s single-asset-of-a-faction pick. (Logged as a reversal because
   Effort 1 Decision 2 anticipated Effort 3 adding the `TargetAsset` case; the
   multi-select + paired-destination requirement, surfaced by the Transit Net
   cardinality call this session, overrides it.)

3. **Transit Net moves multiple assets per action; Covert Shipping moves one** (Robert,
   this session). `RelocateParams.MaxAssets` governs cardinality: `1` for Covert
   Shipping, `0` (unbounded) for Transit Net. `SelectRelocateOrders` is a multi-select
   returning one `RelocateOrder{Asset, DestinationWorld}` per moved asset.

4. **The asset-type restriction is data** (Robert, this session). `RelocateParams.AssetType`
   (`asset_type` in TOML, empty = any) filters the "other" candidate assets. Covert
   Shipping / Transit Net set `asset_type = "Special Forces"`. Honors the catalog's
   data-driven thesis rather than hardcoding `TypeSpecialForces` in the handler.

5. **Hex-mode range is raw hex distance** (`crossingCost = 0`). Unlike `ChangeHomeworld`,
   which prices the drift journey via `DriftCost(3)`, relocate is an instant repositioning.
   *Refined by DR-1/DR-2:* relocate now has two range modes (`hexes` | `region`). Hex mode is
   raw hex distance — only `WorldsWithin` calls `Distance(..., 0)` (the `// deliberate:`
   fence-sign lives there). Region mode (`WorldsInRegion`) is pure `RegionID` equality, no
   distance. `BuildRelocateMutations` no longer calls `Distance` at all (membership check).

6. **Range is measured from the acting asset, not the moved asset.** SWN: "moved
   between any worlds within three hexes *of the Covert Shipping*." Both the candidate
   assets (their current world must be in range) and the destination are filtered
   against the **acting asset's** hex. For self-relocate the acting asset *is* the
   moved asset, so "within 1 hex of itself" falls out naturally.

7. **Affordability is checked up front in the handler.** Covert Shipping costs 1 Coin
   per moved asset; the handler errors if `faction.Coin < CoinCost * len(orders)`
   before emitting, rather than emitting an unaffordable negative `CoinDelta`. Relocate
   is not a roll-table, so there is no insolvency/`DestroySelf` path here.

8. **Enumeration stays behind the world seam.** Rather than expose `AllWorlds` to the
   action layer, `WorldEngine.WorldsWithin(from, maxHex)` (hex) and — per DR-1 —
   `WorldsInRegion(from)` (region) gather in-range destination worlds. This requires adding
   `AllWorlds() []spatial.Location` to the `HexRouter` interface (`RegionMap` already
   implements it; per DR-3 six test stubs gain the method, five trivially).

## Decision Record — Execution

<!-- Execution-time decisions and reversals land here, logged before the commit they
ride. Append AND reconcile. -->

**DR-1 — Discriminated `range` field on relocate (new capability; Robert, Commit 1 session).**
Relocate gains a `RelocateRange` discriminator (`hexes` | `region`) on `RelocateParams`,
`range` in TOML (default `hexes`). `range = "region"` ignores `max_hex` and moves assets to
any world sharing the **acting asset's region** — pure `RegionID` equality, since every hex
already carries its region. Motivated by Robert wanting a region-scoped mover; **Covert
Transit Net** is authored with `range = "region"` (its region-wide-net theme fits the broad
reach), the other two assets stay hex-scoped. Refines Decision 5. Generalizing region limits
to the speed/drift movement engine is explicitly out of scope — backlogged as `F.023`.

**DR-2 — Membership-based relocate primitive (reverses the distance-recompute shape of
Decision 5 / Commit 1).** `BuildRelocateMutations` no longer recomputes hex distance. It
takes the caller's range-filtered `legalWorlds []string` and validates
`slices.Contains(legalWorlds, dest.WorldID)`. The hex-vs-region geometry lives entirely in
the enumerator the handler picks — `WorldsWithin` (hex radius) or the new `WorldsInRegion`
(same region). Rationale: a mode branch inside the primitive would grow per range mode;
membership keeps the primitive range-agnostic (a future mode = a new enumerator, zero
primitive change) and is a net simplification (~18 lines → ~4). Collector-injection defense
is preserved — a destination absent from `legalWorlds` is rejected exactly as a too-far hex
was (the `ChangeHomeworld` NoPartialState threat model). Signature change: dropped
`fromHex`/`maxHex`, added `legalWorlds`. The `crossingCost 0` fence-sign now lives **only**
in `WorldsWithin`.

**DR-3 — `HexRouter.AllWorlds` fan-out wider than planned (Commit 1 execution).** Widening
`HexRouter` broke six implementers, not the one (`testHexRouter`) Task 2 named:
`stubWorldMap` (orchestrator_test), `StubSpatialMap` (testharness/harness), `smokeHexRouter`
(tui/adapter/smoke), and `stubMap` (world_test — covers the embedded `stubSingleWorld` /
`stubMapWithDest`). Each pure satisfier got `AllWorlds() []spatial.Location { return nil }`;
`testHexRouter` got the real enumeration. Plan body (Commit 1 Task 2) reconciled below.

**DR-4 — `MaxAssets` enforced by a handler guard (Robert, Commit 3 session; refines Decision 2
and Decision 7).** Decision 2 made relocate own all selection and passed `MaxAssets` to the
collector as the cardinality mechanism; the handler itself did not enforce it. That leaves the
cap unprotected for any non-interactive caller (a scripted/mock collector, or a future
automated path) and for an over-clicking GM — a Covert Shipping (`max_assets = 1`) could move
2+ Special Forces for 2+ Coin, a rules violation the affordability check would not catch.
Resolution: `relocate` gains a one-line backstop after the empty-selection check —
`if params.MaxAssets > 0 && len(orders) > params.MaxAssets { error }` — mirroring the
up-front affordability guard (Decision 7) as an engine-level invariant independent of the
collector. The collector/overlay still drives the happy-path selection; the handler now
*guarantees* the cap. Unit-tested (`TestRelocate_ExceedsMaxAssets`).

**DR-5 — TUI overlay shape and payload (Commit 3 execution; settles the OQ "TUI relocate
overlay fidelity").** Two adapter-side choices, both additive: (a) `SelectRelocateOrdersPayload`
carries an adapter-built `WorldNames map[string]string`, matching every other world-picker
payload (seize / change-homeworld / buy) rather than rendering raw world IDs — the plan's
sketch payload omitted it. (b) The overlay (`relocateorders.go`) is one destination
`Select[string]` per candidate asset with a `""` "— do not move —" skip option, folding asset
multi-select and per-row destination into a single static huh form (the plan's "model on
`repairorders.go`" steer). The cap is enforced engine-side (DR-4), so the overlay stays a
plain composite with no cross-field limit wiring.

---

## Open Questions — To Ratify at Implementation Time

- **TUI relocate overlay fidelity.** Commit 3 adds a `SelectRelocateOrders` overlay
  modeled on `repairorders.go` (multi-select with a per-row sub-choice) + `worldselect.go`
  (the destination picker). The exact bubbletea layout is an execution detail; the
  contract (candidates in, `[]RelocateOrder` out) is fixed here. Build against the live
  overlay patterns at execution.
- **Effort-1 carried question (multi-step built + unbuilt effects) — now partially
  fires.** Seductress is the first genuine multi-step ability, but both its steps have
  registered handlers (relocate, reveal), so the no-handler fallthrough's
  mutation-discard hazard still does not trigger. Remains deferred to the first
  authored *description-only* step.

## Shared Context

**Model.** This Plan session runs on Opus. For execution: Commit 1 (world primitive)
and Commit 2 (threading/wiring) are mechanical-to-moderate — Sonnet is appropriate.
Commit 3 is design-heavy (the handler's selection orchestration, the threaded-asset-view
loop change, the multi-step test, and a non-trivial TUI overlay) — suggest Opus, or
flag if Sonnet feels stretched. Surface the recommendation at the Plan → Execution
transition.

**Re-ground before coding (verified this session).** Effort 1/2's built API matches
their plans: `StepContext`/`EffectHandler` + the `handlers` map in `dispatch.go:13-31`,
the halt-on-failure loop with the fence-sign at `dispatch.go:63-68`, `resolveTarget`/
`passesGate`/`resolveRoll`/`outcomeFor` in `stage.go`. `resolveTarget` still errors on
`TargetAsset` (`stage.go:34-36`) and stays that way (Decision 2). `AssetMoved` applies
at `mutation/mutation.go:193` (sets `asset.Location = ToLocation`, idempotent).
`RegionMap.AllWorlds()` exists (`region_map.go:194`) but is not yet on `HexRouter`.
`register.go:38` builds `NewUseAssetAbility(c, resolveRoller())` with no world access;
`ChangeHomeworld` (`register.go:42`, `change_homeworld.go`) is the precedent for taking
the full `*world.WorldEngine` and gathering eligible worlds before prompting.

**Test patterns.** Follow `ability/dispatch_test.go`: `seqRoller{values}` for the
contest's d10s, gomock `MockCollector` for prompts. World-engine units use
`world.NewWithMap(testRouter, ...)` with a `HexRouter` stub (mirror `testHexRouter` in
`change_homeworld_test.go`). The multi-step Seductress test is the headline: it proves
(a) the threaded asset view (reveal fires at the destination world) and (b)
halt-on-failure preserves the relocate when the later contest loses.

**Mock regeneration.** `MockCollector` is generated (`//go:generate mockgen` on
`action/collector.go:3`). The interface change is applied by regenerating, not
hand-editing.

## Out of Scope

- `coin_or_disable`, `purchase_enabler`, `attach` (Effort 4); wiring the `TargetAsset`
  case of `SelectAbilityTarget` (Effort 4).
- The S-side *free* relocate harness (Transit Web), `GrantedMovementAbilities`, and any
  drift-order interaction — relocate is instant and distinct from the movement phase
  (A.001.3+).
- The D5 prose cleanup on Speed-based assets, unless a touched file makes it trivial.

---

## Work Breakdown

Single phase, three commits, executed in order. Each builds with green tests.

### Phase 1 — Relocate primitive

#### Commit 1 — `feat: world-engine instant relocate primitive`

World-package additions only; no ability consumers yet. Self-contained and unit-tested.

##### Task 1 — `internal/spatial/spatial.go` is unchanged; `internal/faction/engine/world/world.go`

Add `AllWorlds` to the `HexRouter` interface (Decision 8) and two methods to
`WorldEngine`.

**Find:**
```go
type HexRouter interface {
	spatial.SpatialMap
	Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
	Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
}
```
**Replace with:**
```go
type HexRouter interface {
	spatial.SpatialMap
	Distance(from, to spatial.RegionHex, crossingCost int) (int, error)
	Path(from, to spatial.RegionHex, crossingCost int) ([]spatial.RegionHex, int, error)
	AllWorlds() []spatial.Location
}
```

At the end of `world.go`, **add:**
```go
// WorldsWithin returns the IDs of every world within maxHex of the given hex,
// name-sorted, excluding the origin hex itself. Distance is raw hex count.
func (engine *WorldEngine) WorldsWithin(from spatial.RegionHex, maxHex int) ([]string, error) {
	var within []string
	for _, loc := range engine.spatialMap.AllWorlds() {
		hexLoc, ok := loc.(spatial.RegionLocation)
		if !ok {
			continue
		}
		// deliberate: crossingCost 0 — "within N hexes" is raw hex distance, not the
		// drift-priced journey ChangeHomeworld uses (Effort 3 Plan, Decision 5).
		dist, err := engine.spatialMap.Distance(from, hexLoc.RegionHex(), 0)
		if err != nil {
			continue
		}
		if dist > 0 && dist <= maxHex {
			within = append(within, hexLoc.ID())
		}
	}
	sort.Strings(within)
	return within, nil
}

// WorldsInRegion returns the IDs of every world sharing the origin hex's region,
// name-sorted, excluding the origin hex itself. Region membership is a pure equality
// check — no distance math, since every hex already carries its RegionID. (DR-1)
func (engine *WorldEngine) WorldsInRegion(from spatial.RegionHex) ([]string, error) {
	var within []string
	for _, loc := range engine.spatialMap.AllWorlds() {
		hexLoc, ok := loc.(spatial.RegionLocation)
		if !ok {
			continue
		}
		hex := hexLoc.RegionHex()
		if hex.RegionID != from.RegionID || hex == from {
			continue
		}
		within = append(within, hexLoc.ID())
	}
	sort.Strings(within)
	return within, nil
}

// BuildRelocateMutations validates that destination is one of the caller-supplied
// legalWorlds — the range-filtered candidates from WorldsWithin (hex radius) or
// WorldsInRegion (same region) — and emits the AssetMoved for the moved asset, plus a
// CoinDelta when coinCost > 0. Instant, distinct from the drift-order movement phase.
// Range geometry lives in the enumerators; this primitive is range-mode-agnostic. (DR-2)
func (engine *WorldEngine) BuildRelocateMutations(
	faction *domain.Faction,
	moved *domain.Asset,
	destination domain.Location,
	legalWorlds []string,
	coinCost int,
) ([]domain.Mutation, error) {
	if !slices.Contains(legalWorlds, destination.WorldID) {
		return nil, fmt.Errorf("relocate: destination %q is not a legal target", destination.WorldID)
	}
	mutations := []domain.Mutation{domain.AssetMoved{
		FactionID:         faction.ID,
		AssetID:           moved.ID,
		FromLocation:      moved.Location,
		ToLocation:        destination,
		Cause:             "ability",
		CausedByFactionID: faction.ID,
	}}
	if coinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{
			FactionID:         faction.ID,
			Delta:             -coinCost,
			Cause:             "ability",
			CausedByFactionID: faction.ID,
		})
	}
	return mutations, nil
}
```
Add `"fmt"`, `"slices"`, and `"sort"` to the `world.go` import block (currently `log/slog`,
domain, state, spatial).

##### Task 2 — `HexRouter` stubs gain `AllWorlds` (DR-3)

Widening `HexRouter` fails the `var _ world.HexRouter` assertions on every stub. Add the
method to all six (five trivial, one real):
- `change_homeworld_test.go` — `testHexRouter` gets the real enumeration:
  ```go
  func (r *testHexRouter) AllWorlds() []spatial.Location {
  	worlds := make([]spatial.Location, 0, len(r.worlds))
  	for id, hex := range r.worlds {
  		worlds = append(worlds, &testWorldLoc{id: id, hex: hex})
  	}
  	return worlds
  }
  ```
- `orchestrator_test.go` (`stubWorldMap`), `testharness/harness.go` (`StubSpatialMap`),
  `tui/adapter/smoke.go` (`smokeHexRouter`), `world_test.go` (`stubMap` — covers embedded
  `stubSingleWorld` / `stubMapWithDest`): each gets `AllWorlds() []spatial.Location { return nil }`.

##### Task 3 — `internal/faction/engine/world/relocate_test.go` (new)

Unit-test `BuildRelocateMutations`, `WorldsWithin`, and `WorldsInRegion` against a
`HexRouter` stub (`relocateRouter`: distance = |Q-diff| ignoring region; worlds carry a
`RegionID`). Cases:
- legal destination → one `AssetMoved` with the right `ToLocation`; `coinCost > 0` appends
  a `CoinDelta{Delta: -coinCost}`; `coinCost == 0` appends none.
- destination absent from `legalWorlds` → error, no mutations (collector-injection guard).
- `WorldsWithin` returns only worlds with `1 <= dist <= maxHex`, name-sorted, origin excluded.
- `WorldsInRegion` returns only same-`RegionID` worlds (any distance), name-sorted, origin
  and other-region worlds excluded.

##### Commit message
```
feat: world-engine instant relocate primitive

- WorldEngine.BuildRelocateMutations: validate dest is a caller-supplied legal world,
  emit AssetMoved + optional CoinDelta; instant, distinct from the drift-order path
- WorldEngine.WorldsWithin (hex radius) + WorldsInRegion (same region) gather candidate
  worlds behind the world seam; range geometry lives in the enumerators
- add AllWorlds() to the HexRouter interface (RegionMap already implements it)
```

#### Commit 2 — `feat: thread world engine into ability dispatch`

Give the dispatch path world-engine + collector access, mirroring how `ChangeHomeworld`
receives the engine. No new effect; existing reveal/roll tests still pass.

##### Task 1 — `internal/faction/engine/action/actions/ability/dispatch.go`

Add `Collector` and `WorldEngine` to `StepContext`, accept the world engine in
`Dispatch`, and pass both into the `StepContext`. (The threaded-asset-view loop change
lands in Commit 3 with its Seductress consumer.)

**Find:**
```go
import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

**Find:**
```go
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
	RollResult   int // resolved by the roll stage when Step.Roll != nil; 0 otherwise
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
}
```
**Replace with:**
```go
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
	RollResult   int // resolved by the roll stage when Step.Roll != nil; 0 otherwise
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
	Collector    action.Collector    // for handlers with a secondary prompt (relocate)
	WorldEngine  *world.WorldEngine  // for movement effects (relocate); nil for non-spatial
}
```

**Find:**
```go
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
) ([]domain.Mutation, error) {
```
**Replace with:**
```go
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
```

**Find:**
```go
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: factionState,
			Rulebook:     rb,
		})
```
**Replace with:**
```go
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: factionState,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
		})
```

##### Task 2 — `internal/faction/engine/action/actions/use_asset_ability.go`

Store the world engine and forward it to `Dispatch`.

**Find:**
```go
type UseAssetAbility struct {
	collector      action.Collector
	roller         domain.Roller
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
	}
}
```
**Replace with:**
```go
type UseAssetAbility struct {
	collector      action.Collector
	roller         domain.Roller
	world          *world.WorldEngine
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller, worldEngine *world.WorldEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
		world:     worldEngine,
	}
}
```

**Find:**
```go
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, factionState, rulebook)
```
**Replace with:**
```go
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, factionState, rulebook, u.world)
```
Add the `world` package to the import block.

##### Task 3 — `internal/faction/engine/action/actions/register.go`

Pass the world engine into the factory.

**Find:**
```go
	ae.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, resolveRoller())
	})
```
**Replace with:**
```go
	ae.Register(func(c action.Collector) action.Action {
		return NewUseAssetAbility(c, resolveRoller(), worldEngine)
	})
```

##### Task 4 — test call-site updates

`Dispatch(...)` and `NewUseAssetAbility(...)` gained a parameter. Update:
- `internal/faction/engine/action/actions/ability/dispatch_test.go` — every `Dispatch`
  call takes a trailing world-engine arg; pass `nil` (reveal/stub/roll-table tests use
  no world engine).
- `internal/faction/engine/action/actions/use_asset_ability_test.go` — `NewUseAssetAbility`
  call takes a trailing arg; pass `nil`.
Re-ground at execution: grep both files for the call sites; the reveal tests pass `nil`
safely because `revealStealth` never touches `WorldEngine`.

##### Commit message
```
feat: thread world engine into ability dispatch

- StepContext carries Collector + WorldEngine for effects with a secondary prompt
  or movement (relocate); nil world engine tolerated for non-spatial effects
- Dispatch and NewUseAssetAbility accept the world engine; register.go wires it
  (mirrors ChangeHomeworld)
```

#### Commit 3 — `feat: relocate ability effect (self / other)`

The `relocate` effect, its params, parser, the order-slice collector method across all
five implementations, the threaded-asset-view loop change, the three authored assets,
and the multi-step Seductress test.

##### Task 1 — `internal/faction/domain/asset.go`

Add the effect constant, the relocate scope/params, and the `Relocate` sub-struct on
`AbilityStep` (Effort 1 Decision 4: added additively at first use).

**Find:**
```go
const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
)
```
**Replace with:**
```go
const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
	EffectRelocate         AbilityEffectType = "relocate"
)

type RelocateScope string

const (
	RelocateSelf  RelocateScope = "self"
	RelocateOther RelocateScope = "other"
)

// RelocateRange discriminates how far a relocate may reach (DR-1). RangeHexes uses MaxHex
// (raw hex distance from the acting asset); RangeRegion ignores MaxHex and allows any world
// in the acting asset's region.
type RelocateRange string

const (
	RangeHexes  RelocateRange = "hexes"
	RangeRegion RelocateRange = "region"
)

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

**Find:**
```go
type AbilityStep struct {
	Effect  AbilityEffectType
	Target  TargetKind
	Contest *Contest
	Roll    *RollTable
}
```
**Replace with:**
```go
type AbilityStep struct {
	Effect   AbilityEffectType
	Target   TargetKind
	Contest  *Contest
	Roll     *RollTable
	Relocate *RelocateParams
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Parse the effect string and a `[relocate]` sub-table on both the flat block and a step.

In `toAbilityEffect`'s switch, **add** the case:
```go
	case "relocate":
		return domain.EffectRelocate, nil
```

Add `Relocate *relocateRecord` to **both** `abilityRecord` and `abilityStepRecord`
(`toml:"relocate"`), and add the record type:
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

`convertAbilityStep` gains a `relocate` parameter (thread it from both call sites in
`convertAbility` — flat: `r.Relocate`; steps: `sr.Relocate`). After the contest/roll
parsing, before the contest-xor-roll guard, **add:**
```go
	if relocate != nil {
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
And the converters (`range` defaults to `hexes` when omitted, so existing hex assets need no
`range` key):
```go
func toRelocateScope(s string) (domain.RelocateScope, error) {
	switch s {
	case "self":
		return domain.RelocateSelf, nil
	case "other":
		return domain.RelocateOther, nil
	default:
		return "", fmt.Errorf("unknown relocate scope: %q", s)
	}
}

func toRelocateRange(s string) (domain.RelocateRange, error) {
	switch s {
	case "", "hexes":
		return domain.RangeHexes, nil
	case "region":
		return domain.RangeRegion, nil
	default:
		return "", fmt.Errorf("unknown relocate range: %q", s)
	}
}
```
Re-ground at execution: update the `convertAbilityStep` signature and its two call
sites in `convertAbility` to pass the relocate record (the current signature is
`convertAbilityStep(effectStr, targetStr, attackerStr, defenderStr, dieStr string, outcomes []outcomeRecord)`).

##### Task 3 — `internal/faction/engine/action/collector.go`

Add the order type and the collector method.

After the `BuyOrder` type, **add:**
```go
// RelocateOrder describes one asset to move and the world it moves to.
type RelocateOrder struct {
	Asset            *domain.Asset
	DestinationWorld string
}
```
In the `Collector` interface, after `SelectAbilityTarget`, **add:**
```go
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]RelocateOrder, error)
```

##### Task 4 — `internal/faction/engine/action/actions/ability/relocate.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// relocate resolves a movement ability: gather the movable candidates (the acting
// asset for self-scope; range- and type-filtered friendly assets for other-scope),
// prompt the GM for the moved assets and their destinations, and emit the relocate
// mutations via the world engine. Range is measured from the acting asset.
func relocate(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.Relocate
	if params == nil {
		return nil, fmt.Errorf("relocate: step has no relocate params")
	}
	if ctx.WorldEngine == nil {
		return nil, fmt.Errorf("relocate: no world engine")
	}
	actingHex := ctx.Asset.Location.RegionHex
	// DR-1: the range mode picks the enumerator; the rest of the handler is mode-agnostic.
	var (
		candidateWorlds []string
		err             error
	)
	switch params.Range {
	case domain.RangeRegion:
		candidateWorlds, err = ctx.WorldEngine.WorldsInRegion(actingHex)
	default: // RangeHexes (and the zero value, for safety)
		candidateWorlds, err = ctx.WorldEngine.WorldsWithin(actingHex, params.MaxHex)
	}
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}

	var candidates []*domain.Asset
	maxAssets := params.MaxAssets
	switch params.Scope {
	case domain.RelocateSelf:
		candidates = []*domain.Asset{ctx.Asset}
		maxAssets = 1
	case domain.RelocateOther:
		candidates = eligibleRelocateAssets(ctx, params, candidateWorlds)
	default:
		return nil, fmt.Errorf("relocate: unknown scope %q", params.Scope)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("relocate: no movable assets in range")
	}

	orders, err := ctx.Collector.SelectRelocateOrders(ctx.Asset, candidates, candidateWorlds, maxAssets)
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("relocate: %w", action.ErrNoSelection)
	}
	if ctx.Faction.Coin < params.CoinCost*len(orders) {
		return nil, fmt.Errorf("relocate: insufficient Coin: need %d, have %d", params.CoinCost*len(orders), ctx.Faction.Coin)
	}

	var mutations []domain.Mutation
	for _, order := range orders {
		loc, ok := ctx.WorldEngine.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := ctx.WorldEngine.BuildRelocateMutations(ctx.Faction, order.Asset, dest, candidateWorlds, params.CoinCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
	}
	return mutations, nil
}

// eligibleRelocateAssets returns the acting faction's assets that may be relocated by
// an other-scope step: matching AssetType (when set), currently in a world within
// range, and not the acting asset itself.
func eligibleRelocateAssets(ctx StepContext, params *domain.RelocateParams, candidateWorlds []string) []*domain.Asset {
	inRange := make(map[string]struct{}, len(candidateWorlds))
	for _, worldID := range candidateWorlds {
		inRange[worldID] = struct{}{}
	}
	var eligible []*domain.Asset
	for _, asset := range ctx.Faction.Assets {
		if asset.ID == ctx.Asset.ID {
			continue
		}
		if _, ok := inRange[asset.Location.WorldID]; !ok {
			continue
		}
		if params.AssetType != "" {
			def, ok := ctx.Rulebook.Assets[asset.DefinitionID]
			if !ok || def.Type != params.AssetType {
				continue
			}
		}
		eligible = append(eligible, asset)
	}
	return eligible
}
```
Re-ground at execution: confirm `ctx.Faction.Assets` keying, `action.ErrNoSelection`,
and `WorldEngine.Location` (`world.go:74`) return shape; sort `eligible` by ID/name to
match the codebase's deterministic-candidate convention (see `otherFactions`).

##### Task 5 — `internal/faction/engine/action/actions/ability/dispatch.go`

Register the handler and add the threaded-asset-view loop change (Decision 1).

**Find:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
}
```
**Replace with:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
	domain.EffectRelocate:         relocate,
}
```

Thread the moved location forward. The loop currently builds each `StepContext` with
`Asset: asset`; replace the fixed `asset` with a `workingAsset` that absorbs any
`AssetMoved` on the acting asset from prior steps, so a later step (Seductress's
reveal) reads the post-relocate world without writing shared state.

**Find:**
```go
	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		handler, ok := handlers[step.Effect]
		if !ok {
			return confirmApplied(asset, def, collector)
		}
		target, err := resolveTarget(step, faction, asset, factionState, collector)
		if err != nil {
			return nil, err
		}
		passed, err := passesGate(step, faction, target, roller)
		if err != nil {
			return nil, err
		}
		if !passed {
```
**Replace with:**
```go
	var mutations []domain.Mutation
	workingAsset := asset
	for _, step := range def.Ability.Steps {
		handler, ok := handlers[step.Effect]
		if !ok {
			return confirmApplied(asset, def, collector)
		}
		target, err := resolveTarget(step, faction, workingAsset, factionState, collector)
		if err != nil {
			return nil, err
		}
		passed, err := passesGate(step, faction, target, roller)
		if err != nil {
			return nil, err
		}
		if !passed {
```
In the same loop, change the `StepContext` to use `workingAsset`, and after appending
the step's mutations, absorb any move. **Find:**
```go
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: factionState,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
	}
	return mutations, nil
```
**Replace with:**
```go
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        workingAsset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: factionState,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
		workingAsset = applyAssetMovedView(workingAsset, stepMutations)
	}
	return mutations, nil
```
At the end of the file, **add:**
```go
// applyAssetMovedView returns a location-updated copy of asset if muts contains an
// AssetMoved targeting it, so later steps in the same ability read the post-move
// location without writing shared state (the real AssetMoved still applies later).
func applyAssetMovedView(asset *domain.Asset, muts []domain.Mutation) *domain.Asset {
	for _, mutation := range muts {
		if moved, ok := mutation.(domain.AssetMoved); ok && moved.AssetID == asset.ID {
			clone := *asset
			clone.Location = moved.ToLocation
			return &clone
		}
	}
	return asset
}
```

##### Task 6 — non-TUI collector implementations

- `internal/faction/engine/action/actions/mocks/mock_collector.go` — regenerate (Task:
  `go generate ./internal/faction/engine/action/...`).
- `internal/faction/engine/testharness/collector.go` — add a scripted hook:
  ```go
	SelectRelocateOrdersFn func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)
  ```
  and the method:
  ```go
  func (c *ScriptedCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
  	if c.SelectRelocateOrdersFn != nil {
  		return c.SelectRelocateOrdersFn(acting, candidates, candidateWorlds, maxAssets)
  	}
  	return nil, nil
  }
  ```
  Re-ground: confirm whether `testharness/collector.go` imports the `action` package
  for `RelocateOrder` (other order types there will show the pattern).
- `internal/faction/tui/adapter/smoke.go` — stub:
  ```go
  func (SmokeActionCollector) SelectRelocateOrders(_ *domain.Asset, _ []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
  	return nil, action.ErrActionUnavailable
  }
  ```

##### Task 7 — TUI relocate overlay (`internal/faction/tui/adapter/action_collector.go` + a new overlay)

Implement `SelectRelocateOrders` on `actionCollector` via a new ask channel + payload,
and add the overlay view. Model the multi-select-with-per-row-destination on
`repairorders.go` (multi-select + per-row sub-choice) and reuse the world picker shape
from `worldselect.go` for the destination. The adapter method shape:
```go
func (a *actionCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	raw, err := a.adapter.ask(AskSelectRelocateOrders, nil, SelectRelocateOrdersPayload{
		Acting:          acting,
		Candidates:      candidates,
		CandidateWorlds: candidateWorlds,
		MaxAssets:       maxAssets,
	})
	if err != nil {
		return nil, err
	}
	orders, _ := raw.([]action.RelocateOrder)
	return orders, nil
}
```
Add `AskSelectRelocateOrders` to the ask-message enum, `SelectRelocateOrdersPayload`,
and the overlay registration. This is UI work — Claude implements it (Robert does not
take UI slices); the contract is fixed (candidates + worlds in, `[]RelocateOrder` out).
Re-ground against the live overlay/registry patterns at execution.

##### Task 8 — author the three assets

`rulebooks/swn/assets/cunning_assets.toml`.

**Seductress** — restructure to two steps (relocate first, reveal second). **Find:**
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
effect = "relocate"
[assets.Seductress.ability.steps.relocate]
scope = "self"
max_hex = 1

[[assets.Seductress.ability.steps]]
effect = "reveal_stealth"
target = "faction"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```

**Covert Shipping** — add a flat relocate ability after its `description`:
```toml
[assets.CovertShipping.ability]
effect = "relocate"
[assets.CovertShipping.ability.relocate]
scope = "other"
max_hex = 3
coin_cost = 1
asset_type = "Special Forces"
max_assets = 1
```

**Covert Transit Net** — add a flat relocate ability after its `description`. Uses
`range = "region"` (DR-1): a region-wide transit net, no `max_hex`.
```toml
[assets.CovertTransitNet.ability]
effect = "relocate"
[assets.CovertTransitNet.ability.relocate]
scope = "other"
range = "region"
asset_type = "Special Forces"
max_assets = 0
```
Re-ground at execution: confirm the TOML array-of-tables sub-table nesting
(`[assets.Seductress.ability.steps.relocate]` attaches to the preceding `[[...steps]]`
element) parses as intended; the `rulebook_test` assertion in Task 10 guards it.

##### Task 9 — `internal/faction/engine/action/actions/ability/relocate_test.go` (new)

Hand-build defs on the step shape (no parser), `world.NewWithMap(testRouter, ...)` with
a `HexRouter` stub, and a `ScriptedCollector` (or `MockCollector`) returning the orders.
Cases:
- **self (hex)**: one `RelocateOrder{acting, destWorld}` → one `AssetMoved`, no `CoinDelta`.
- **other, single (Covert Shipping, hex)**: candidates filtered to in-range Special Forces;
  one order → `AssetMoved` + `CoinDelta{-1}`; insufficient Coin → error.
- **other, multi (Transit Net, region)**: `Range = region` → candidates drawn from the whole
  region via `WorldsInRegion`; two orders → two `AssetMoved`, no `CoinDelta`.
- type/range filtering: a wrong-type asset, a hex-out-of-range asset (hex mode), and an
  out-of-region asset (region mode) are each excluded from candidates.

##### Task 10 — `internal/faction/engine/action/actions/ability/dispatch_test.go` — multi-step Seductress

The headline test (Effort 1 Decision 1 + this effort's Decision 1). Build Seductress's
two-step def (relocate self ≤1 hex, then reveal faction), a world engine whose router
places the Seductress and a target faction's stealthed asset at the **destination**
world, a `seqRoller` for the contest, and a collector scripting both
`SelectRelocateOrders` (→ the self asset + destination world) and `SelectAbilityTarget`
(→ the target faction). Assert:
- **contest win**: mutations contain the `AssetMoved` *and* an `AssetStealthCleared`
  for the stealthed asset at the destination world — proving the threaded asset view
  (`applyAssetMovedView`) let the reveal read the post-move location.
- **contest loss**: mutations contain only the `AssetMoved` — proving halt-on-failure
  preserves the ungated relocate when the later gated reveal fails.

##### Task 11 — `internal/faction/rulebook/rulebook_test.go`

Add a parse assertion guarding the authoring. After the existing Seductress reveal
subtest, extend/replace it to assert the two-step shape: step 0 is `EffectRelocate`
with `Relocate.Scope == RelocateSelf` and `MaxHex == 1`; step 1 is `EffectRevealStealth`
with the faction target + contest. Add a Covert Shipping subtest: one step,
`EffectRelocate`, `Relocate{Scope: other, Range: RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: Special Forces, MaxAssets: 1}`
(default `range` omitted in TOML parses to `RangeHexes`). Add a Covert Transit Net subtest:
one step, `EffectRelocate`, `Relocate{Scope: other, Range: RangeRegion, AssetType: Special Forces, MaxAssets: 0}`.

##### Commit message
```
feat: relocate ability effect (self / other)

- EffectRelocate + RelocateParams (scope, range, max_hex, coin_cost, asset_type, max_assets)
  on AbilityStep; parser case + [relocate] sub-table on flat block and steps
- relocate handler: pick the enumerator by range (hex radius or whole region), gather
  type-filtered candidates, prompt SelectRelocateOrders, emit mutations per order via the
  world engine; up-front affordability check
- threaded asset view: a step's AssetMoved updates the working asset for later steps
  (Seductress reveals at its destination) without writing shared state
- SelectRelocateOrders across all five collectors (TUI overlay + scripted/smoke/mock)
- author Covert Shipping, Covert Transit Net, and Seductress's self-relocate step
  (Seductress now live end-to-end; multi-step halt-on-failure test)
```
