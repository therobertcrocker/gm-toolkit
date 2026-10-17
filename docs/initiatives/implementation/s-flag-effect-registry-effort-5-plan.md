# S-flag Effect Registry (A.001.3) — Effort 5 Plan: Movement and Purchase

## Session Cadence

This effort runs in four sessions, not one per commit. A commit that sets a contract
gets its own session, so the session after it can re-ground against the code that was
built. A commit that only uses a contract can share a session, with a check-in between
commits.

| Session | Commits | Why |
|---------|---------|-----|
| 1 | Commits 1, 2 | Commit 1 sets the shared relocate contract; Commit 2 grows it in the same package. |
| 2 | Commit 3 | Changes a hook type, the phase collector contract and the movement phase. |
| 3 | Commit 4 | Uses Commit 3's granter contract; registrar plus data. |
| 4 | Commit 5 | Adds a second phase route to the orchestrator and changes a built reactor. |

## Context / Goal

Build the free relocate harness and the purchase route, then author the last two
assets. Today `relocate` lives inside the A-side `ability` package. The movement phase
asks no `AssetMovementGranter`, and nothing fires an effect when an asset is bought.
Effort 5 closes dispatch items 5 and 6 and ships `grant_movement` and `apply_stealth`.

- Overview plan: [`s-flag-effect-registry-plan.md`](./s-flag-effect-registry-plan.md)
  (Effort 5 row and outline, Planning Decisions 2, 4 and 5, Decision Record 7 and 11,
  Reconciliation Items).
- Discovery: [`s-flag-effect-registry-discovery.md`](../discovery/s-flag-effect-registry-discovery.md)
  (dispatch items 5 and 6; Open Question 1).
- Builds on Effort 1's registrar contract (`effect/registrar.go`): the `Fire` route,
  `IsPhaseTrigger` (it already lists `on_purchase`), and `FireTurnStart` as the model
  for a phase read. Builds on Effort 2's `active()` rule (Decision Record 11) and
  Tripwire Cells' `clearStealthReactor`.
- The loader already parses both effects (Effort 1): `grant_movement` with its
  `[grant_movement.relocate]` block, and `apply_stealth` with no parameters.
  `RelocateParams` already carries `Categories`, `ExcludeTypes` and `CoinCostMode`.
- Assets: Transit Web `SWN-W7-003` (`grant_movement`), Psychic Assassins `SWN-F5-003`
  (`apply_stealth`). Tripwire Cells `SWN-C4-003` gains its purchase case.

The effort touches the A side in two places. Commit 1 moves the action-spent relocate
into the shared package. Commit 2 changes which assets it may move (the `CurrentOrder`
guard, Decision 4). Execution re-grounds against `ability/relocate.go` before it moves
it.

## Decisions Ratified in Planning

1. **A movement grant is faction-level, and carries its source.** The placeholder asks
   "what extra movement does *this asset* have?"
   (`GrantMovementAbilities(asset)`). Transit Web is one ability, centred on the Web,
   that moves many *other* assets for one Coin. With the per-asset query, the Web's
   asset-scope hook is found only when the asked asset is the Web itself. Registering
   at faction scope instead breaks the asset-scope rule (Decision Record 14) and
   returns one copy of the grant per asked asset. `{Name, Description}` also carries
   nothing the phase can run. The query is now `GrantMovementAbilities(faction)`, and
   `MovementAbility` is `{Source *domain.Asset, Relocate domain.RelocateParams}`. The
   resolver walks the faction's global and faction buckets, then each of its own assets'
   buckets in ID order. This is the `ResolveWorldTechLevel` shape: a grant serves its
   owner only. Nothing implements or calls the old shape, so nothing migrates. A
   per-asset free move is a grant whose `Source` is that asset, with
   `Relocate.Scope = self`. Resolves the overview's "reshape `MovementAbility`" item.

2. **`relocate` is a new `engine/relocate` package with one entry point, `Run`.**
   Planning Decision 2 ratified the package and the one-method `OrderCollector`. Two
   more choices: `RelocateOrder` stays in `action`, and `relocate` imports `action`.
   The engine already imports `action` (`engine/collector.go`), and `world` does not
   import it, so there is no cycle. Moving the type would change the action collector,
   its mock, the TUI overlay and the testharness, for no gain. A `Request` struct
   carries the inputs. The ability handler is renamed `relocateAssets`, because a
   function named `relocate` would shadow the package. Rejected: a type alias for
   `RelocateOrder` (no aliases, by project rule).

3. **`Run` reports "nothing to do" with sentinels, and each harness decides what they
   mean.** `relocate.ErrNoCandidates` (nothing in range may move),
   `relocate.ErrInsufficientCoin` (the faction cannot pay), and the existing
   `action.ErrNoSelection` (the GM moved nothing). The action-spent harness surfaces
   all three as errors, as today. The free harness skips the grant on all three, and
   also reports `ErrInsufficientCoin` to the observer, because there the GM chose a
   move the faction cannot pay for. It also skips the grant on
   `action.ErrTurnCanceled` (Esc), because the decision batch has already applied
   (Execution Decision Record 2). `Run` also checks the cost of a single move
   *before* it prompts, so a faction that cannot afford one move is never prompted.
   No sentinel joins `IsRecoverable`'s list. The free harness handles them itself, so
   the orchestrator's abort semantics do not change.

4. **The shared eligibility skips any asset with a `CurrentOrder`.** An asset on its
   way out on a drift order still has its origin `WorldID` until the next tick. A
   relocate would move it while its order keeps running from the old path. The
   guard matches `eligibleCargoForTransport`. It applies to both harnesses, so it also
   closes the same gap on the A side (Covert Shipping, Covert Transit Net). Assets in
   flight need no guard: they have no `WorldID`, so the range filter drops them.

5. **Flat cost is charged by the first order; the world primitive is unchanged.**
   `BuildRelocateMutations` takes `coinCost` per call. Under `flat`, the first order
   passes `CoinCost` and the rest pass `0`, so the history shows one `CoinDelta`. The
   overview outline says the geometry is untouched, and this keeps it so.

6. **Granted moves run after the decision batch is applied, one batch per grant.**
   A transport order puts its cargo in transit (no `WorldID`) only when its batch is
   applied (`TransportReactor`, `transport_issued`). If the grants were built in the
   same batch, a Wealth Facility just loaded onto a Freighter Contract would still show
   on its world, and Transit Web could move it too. After the apply, the range filter
   drops it. Each grant is its own `MutationReactors` + `applyAndRecord` batch, so a
   second Web sees the Coin and the locations that the first one changed. Reactors see
   every batch, so Tripwire Cells catch a Transit Web landing. `OnMovementResolved`
   receives the decision mutations and then every grant's mutations, in one call.
   Granted moves run even when the faction has no Speed assets.

7. **Each grant is offered once per movement phase; no budget.** Robert's call: once
   per Web per turn. One grant per Web, offered once, gives that with no counter. Two
   Webs give two grants. A resumed turn re-runs the movement phase and re-offers the
   grants, the same as it re-offers drift decisions today. This is not a regression
   that a budget should fix here.

8. **`PhaseCollector` grows `SelectRelocateOrders`, with the action collector's
   signature.** The TUI hoists the action collector's body into one
   `Adapter.selectRelocateOrders` method that both collectors call. The ask kind
   (`AskSelectRelocateOrders`) and the overlay are reused. The overlay's
   "— do not move —" on every asset gives the decline, so there is no new view. The
   testharness `ScriptedCollector` already has the method. The smoke collector returns
   `nil, nil`, and the orchestrator test stub gains a function field.

9. **The granter checks `active()` on its asset, read through its owner.** This is
   Decision Record 11 and Decision Record 19's pattern for a Cat 4 hook that gets no
   `FactionState`: keep the owner from `RegisterContext` and read
   `owner.Assets[assetID]` at query time. A grant asked for any faction other than its
   owner returns nothing.

10. **`on_purchase` fires from an orchestrator scan of the action batch.** Robert's call.
    After `e.Action.Run`, the action phase walks `actionMutations` for `AssetAdded` and
    calls `e.Effect.FirePurchase` for each bought asset. The scan covers both purchase
    paths (Buy Asset and the purchase enabler) at one site, and threads no
    `EffectsEngine` into the actions. Rejected: a call in each purchase path, which
    needs the effects engine in `BuyAsset` and in `StepContext`.

11. **A purchase is `AssetAdded` with cause `buy` or `ability`; a refit is not.**
    Robert's call: "when purchased" is literal. The scan whitelists the two purchase
    causes, so a future `AssetAdded` emitter does not fire purchase effects without a
    decision. Refit's `AssetAdded` (cause `refit`) does not fire.

12. **`apply_stealth` emits `AssetStealthApplied`, after the `AssetAdded`.** Robert's
    call. This fits the `Fire` contract unchanged, and the history shows "bought, then
    stealthed". `FireContext.Asset` is the asset carried in the `AssetAdded`, because it
    is not in state yet. `FirePurchase` does no active check: a bought asset is never
    ready, and its purchase effect is part of what was bought. The registrar emits
    nothing for an asset that is already stealthy.

13. **Tripwire Cells see a purchase as an `AssetStealthApplied` whose asset the same
    batch adds on their world.** The reactor reads that asset from the batch's
    `AssetAdded`, because reactors see state before the batch applies. A Stealth
    (`SWN-C3-002`) applied to an asset already on the world has no `AssetAdded` in its
    batch, so it is not a purchase and does not trip the Cells. This closes Decision
    Record 7.

14. **`grant_movement` accepts trigger `movement_options_query`, owner `own`, scope
    `faction`, with no filters or gates.** `owner = own` is "the grant serves the
    declaring asset's faction". `scope = faction` is "the faction's own assets move".
    The range and the asset filters come from the relocate block, measured from the
    declaring asset. Both relocate scopes run through the shared code, so the registrar
    accepts both and needs no `Validate`.

15. **`apply_stealth` accepts trigger `on_purchase`, owner `own`, scope `self`, with no
    filters or gates.** It has a `Fire` route only.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

1. **Commit 1: the ability test package keeps the full relocate fixture.** The plan
   kept only the router, the location and a small engine in `ability/relocate_test.go`.
   `dispatch_test.go` (Seductress, cross-step Coin projection) also uses
   `relocRulebook` and the full five-world `relocEngine`. The ability copy keeps both
   unchanged from the original, so the dispatch tests are not touched.

2. **Commit 3: Esc on a grant's relocate prompt declines that grant.** Robert's
   call. The TUI sends `action.ErrTurnCanceled` on Esc, and that sentinel is
   recoverable. Under the plan, it would leave `runGrantedMovement` as an error, and
   `runMovementPhase` would return early: the remaining grants are not offered, and
   `OnMovementResolved` and the checkpoint are skipped, although the decision batch
   has already applied. The runner now skips the grant on `ErrTurnCanceled` too,
   beside `ErrNoSelection` (a `deliberate:` fence-sign marks it). Decision 3 and
   Task 9 are corrected to match. Rejected: Esc ends the phase (the plan as
   written); Esc stops the grant loop but still resolves the phase.

3. **Commit 3: the pre-prompt cost check also changes an A-side test.**
   `TestDispatch_CrossStep_CoinProjection` (Courier, two self-relocate steps, Coin
   1) expected step 2 to prompt and then fail. Step 2 now fails before it prompts,
   so the test expects one prompt. It still proves that step 2 reads the projected
   Coin. `TestRun_InsufficientAfterSelection` is added, because Coin 0 tests no
   longer reach the post-selection cost check.

4. **Commit 3: the docs pass also covers `orchestrator.md` and corrects two
   world-movement Key Decisions.** `orchestrator.md` lists the `PhaseCollector`
   methods, so it gains `SelectRelocateOrders`. In `world-movement.md`, "dispatches
   reactors twice" and "share the world engine but not the movement phase" became
   false; both are corrected in place. `turn-pipeline.md` lists no movement steps
   and needs no edit.

5. **Commit 5: the purchase route is documented in `effect-mutation.md` and
   `orchestrator.md`, not `turn-pipeline.md`.** No page outside the catalog
   described the turn-start route, so Task 7 had no anchor. `turn-pipeline.md`
   covers only the turn engine and needs no edit. `effect-mutation.md`'s "Asset
   features" section still said only transport was wired and that the other
   S-flag features waited for handlers (stale since Effort 1), and it filed
   Psychic Assassins as a `RuleModifier`. That paragraph is rewritten to describe
   the effects list and both phase routes. `orchestrator.md`'s phase-beat note
   gains the purchase fold. The Tripwire scenario is its own test function,
   `TestEffects_TripwireCellsCatchAStealthedPurchase`, which shares a buy-script
   helper with `TestEffects_PsychicAssassinsBoughtStealthed`.

## Shared Context

**Re-grounding.** Commit 3 binds to Commit 1's `relocate.Run`, and Commit 4 binds to
Commit 3's granter. Each session re-reads the code it binds to before writing.

**Data copies.** `campaigns/test-camp/rulebook/assets/` holds copies of the SWN asset
files. Every data edit in this effort goes into both files (Effort 3 convention). Both
copies hold Transit Web and Psychic Assassins.

**Test patterns.** Registrar tests live in `package effects_test` and use the existing
helpers: `activeAsset`, `inactiveStates`, `fixedRoller`, `seqRoller`, and the
`krylos` / `tartarus` locations. Every filter and gate a registrar accepts gets a test
that proves it suppresses (overview Shared Context). These two registrars accept none,
so their tests cover the owner check and the active check. Relocate tests reuse the
`relocRouter` fixture, which moves with them in Commit 1. Scenario tests follow
`testharness/scenarios/effects_test.go`.

**Shipped rulebook.** `TestNewWithRulebook_ShippedRulebook` (`core_test.go`) builds the
engine from `rulebooks/swn`, so each authored entry is checked against its registrar's
`Accepts` (Decision Record 20). It needs no edit.

**Model.** Opus for every commit.

## Out of Scope

- **Asset description prose.** Capital Fleet and Freighter Contract both say "As an
  action…" but move by Speed. That is the same drift Planning Decision 15 names for
  Blockade Fleet and Surveyors. A later rewrite of the asset files covers it (Robert);
  it is not filed to planned-work.
- **`SWN-C3-002` Stealth.** Its hardcoded branch in `buy_asset.go` belongs to A.001.6.
  The purchase route sits beside it and does not migrate it.
- **A budget for granted moves on a resumed turn.** Decision 7.
- **Migrating `[transport]` onto the effects list.** Overview Out of Scope (D-3).
- **The `Name` / `Description` of a grant in the TUI.** The relocate overlay shows the
  candidates, not which grant offers them. Robert does not want UI work pulled in here;
  if it reads poorly in play, file it.

---

## Work Breakdown

### Phase 1 — Shared relocate

One copy of enumerate → filter → prompt → build, then the filters and cost mode that
Transit Web needs.

#### Commit 1 — `refactor: extract the shared relocate orchestration`

Behaviour-preserving move. The error texts stay the same.

##### Task 1 — `internal/faction/engine/relocate/relocate.go` (new)

```go
package relocate

import (
	"errors"
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// OrderCollector is the one prompt a relocation needs. The action collector and
// the phase collector both satisfy it, so the action-spent harness and the free
// harness share the prompt as well as the logic.
type OrderCollector interface {
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error)
}

// ErrNoCandidates reports that nothing in range may move. The action-spent
// harness surfaces it; the free harness skips the grant.
var ErrNoCandidates = errors.New("no movable assets in range")

// Request is one relocation. Acting is the asset range is measured from: the
// ability's own asset, or the asset that granted a free move.
type Request struct {
	Faction   *domain.Faction
	Acting    *domain.Asset
	Params    *domain.RelocateParams
	World     *world.WorldEngine
	Rulebook  *rulebook.Rulebook
	Collector OrderCollector
}

// Run resolves a relocation: gather the movable candidates (the acting asset for
// self-scope; range- and type-filtered friendly assets for other-scope), prompt
// for the moved assets and their destinations, and emit the relocate mutations
// via the world engine. Range is measured from the acting asset.
func Run(request Request) ([]domain.Mutation, error) {
	params := request.Params
	if request.World == nil {
		return nil, fmt.Errorf("relocate: no world engine")
	}
	actingHex := request.Acting.Location.RegionHex
	// other-scope may target the acting asset's own world (a friendly asset there is a
	// legal source, and the home world is a legal destination); self-scope must move off it.
	includeOrigin := params.Scope == domain.RelocateOther
	// DR-1: the range mode picks the enumerator; the rest of the handler is mode-agnostic.
	var (
		candidateWorlds []string
		err             error
	)
	switch params.Range {
	case domain.RangeRegion:
		candidateWorlds, err = request.World.WorldsInRegion(actingHex, includeOrigin)
	default: // RangeHexes (and the zero value, for safety)
		candidateWorlds, err = request.World.WorldsWithin(actingHex, params.MaxHex, includeOrigin)
	}
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}

	var candidates []*domain.Asset
	maxAssets := params.MaxAssets
	switch params.Scope {
	case domain.RelocateSelf:
		candidates = []*domain.Asset{request.Acting}
		maxAssets = 1
	case domain.RelocateOther:
		candidates = eligibleAssets(request, candidateWorlds)
	default:
		return nil, fmt.Errorf("relocate: unknown scope %q", params.Scope)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("relocate: %w", ErrNoCandidates)
	}

	orders, err := request.Collector.SelectRelocateOrders(request.Acting, candidates, candidateWorlds, maxAssets)
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("relocate: %w", action.ErrNoSelection)
	}
	if params.MaxAssets > 0 && len(orders) > params.MaxAssets {
		return nil, fmt.Errorf("relocate: %d assets selected, at most %d allowed", len(orders), params.MaxAssets)
	}
	if request.Faction.Coin < params.CoinCost*len(orders) {
		return nil, fmt.Errorf("relocate: insufficient Coin: need %d, have %d", params.CoinCost*len(orders), request.Faction.Coin)
	}

	var mutations []domain.Mutation
	for _, order := range orders {
		loc, ok := request.World.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, candidateWorlds, params.CoinCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
	}
	return mutations, nil
}

// eligibleAssets returns the faction's assets that an other-scope relocation may
// move: matching AssetType (when set), currently in a world within range, and not
// the acting asset itself. ID-sorted for deterministic prompts.
func eligibleAssets(request Request, candidateWorlds []string) []*domain.Asset {
	inRange := make(map[string]struct{}, len(candidateWorlds))
	for _, worldID := range candidateWorlds {
		inRange[worldID] = struct{}{}
	}
	var eligible []*domain.Asset
	for _, asset := range request.Faction.Assets {
		if asset.ID == request.Acting.ID {
			continue
		}
		if _, ok := inRange[asset.Location.WorldID]; !ok {
			continue
		}
		if request.Params.AssetType != "" {
			def, ok := request.Rulebook.Assets[asset.DefinitionID]
			if !ok || def.Type != request.Params.AssetType {
				continue
			}
		}
		eligible = append(eligible, asset)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ID < eligible[j].ID
	})
	return eligible
}
```

The only change from today's body: the "no movable assets" error now wraps
`ErrNoCandidates`. The text is the same.

##### Task 2 — `internal/faction/engine/action/actions/ability/relocate.go` (rewritten)

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/relocate"
)

// relocateAssets runs a relocate step through the shared relocate code, with the
// acting asset as the range origin. The caller spends the action.
func relocateAssets(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Relocate == nil {
		return nil, fmt.Errorf("relocate: step has no relocate params")
	}
	return relocate.Run(relocate.Request{
		Faction:   ctx.Faction,
		Acting:    ctx.Asset,
		Params:    ctx.Step.Relocate,
		World:     ctx.WorldEngine,
		Rulebook:  ctx.Rulebook,
		Collector: ctx.Collector,
	})
}
```

##### Task 3 — `internal/faction/engine/action/actions/ability/dispatch.go`

**Find:**
```go
	domain.EffectRelocate:         relocate,
```
**Replace with:**
```go
	domain.EffectRelocate:         relocateAssets,
```

##### Task 4 — tests move

- **New** `internal/faction/engine/relocate/relocate_test.go`, `package relocate`
  (internal, because `eligibleAssets` is tested directly). Move `relocRouter`,
  `relocLoc`, `relocHex`, `relocEngine`, `relocRulebook`, `assetIDs` and these tests
  from `ability/relocate_test.go`: `TestRelocate_OtherSingle_Hex`,
  `TestRelocate_OtherSingle_InsufficientCoin`, `TestRelocate_ExceedsMaxAssets`,
  `TestRelocate_OtherMulti_Region`, `TestEligibleRelocateAssets_Filtering` (renamed
  `TestEligibleAssets_Filtering`). Each test calls `Run(Request{...})` in place of
  `relocate(StepContext{...})`, with `Acting` in place of `Asset`. The gomock
  `mocks.NewMockCollector` still satisfies `OrderCollector`.
- **Add** `TestRun_NoCandidatesIsSentinel`: other-scope, nothing in range;
  `errors.Is(err, ErrNoCandidates)`, and the collector is not called.
- **Keep** in `ability/relocate_test.go`: `TestRelocate_Self_Hex`, renamed
  `TestRelocateAssets_Self_Hex`, as the wiring test. It needs its own copy of the small
  router fixture, so keep `relocRouter`, `relocLoc`, `relocHex`, `relocEngine` and
  `relocRulebook` there too; `dispatch_test.go` uses them (Execution Decision Record
  1). Fixture duplication between two test packages is acceptable. Add
  `TestRelocateAssets_NilParams` for the nil-params error.

##### Commit message

```
refactor: extract the shared relocate orchestration

- new relocate package: OrderCollector, Request, Run
- ErrNoCandidates sentinel for an empty candidate set
- the ability relocate handler becomes a call into it
- move the relocate tests with the code
```

#### Commit 2 — `feat: relocate honours category, excluded type and flat cost`

##### Task 1 — `internal/faction/engine/relocate/relocate.go`

Replace `eligibleAssets` and add `matchesFilters`:

```go
// eligibleAssets returns the faction's assets that an other-scope relocation may
// move: in a world within range, passing the definition filters, not the acting
// asset, and not under a movement order. ID-sorted for deterministic prompts.
func eligibleAssets(request Request, candidateWorlds []string) []*domain.Asset {
	inRange := make(map[string]struct{}, len(candidateWorlds))
	for _, worldID := range candidateWorlds {
		inRange[worldID] = struct{}{}
	}
	var eligible []*domain.Asset
	for _, asset := range request.Faction.Assets {
		// deliberate: skip an asset under orders, not just one in flight, because it keeps its origin WorldID until the next tick.
		if asset.ID == request.Acting.ID || asset.CurrentOrder != nil {
			continue
		}
		if _, ok := inRange[asset.Location.WorldID]; !ok {
			continue
		}
		if !matchesFilters(request.Rulebook.Assets[asset.DefinitionID], request.Params) {
			continue
		}
		eligible = append(eligible, asset)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ID < eligible[j].ID
	})
	return eligible
}

// matchesFilters reports whether def passes the relocation's type, category and
// excluded-type filters. An unknown definition passes only when no filter is set.
func matchesFilters(def *domain.AssetDefinition, params *domain.RelocateParams) bool {
	filtered := params.AssetType != "" || len(params.Categories) > 0 || len(params.ExcludeTypes) > 0
	if def == nil {
		return !filtered
	}
	if params.AssetType != "" && def.Type != params.AssetType {
		return false
	}
	if len(params.Categories) > 0 && !slices.Contains(params.Categories, def.Category) {
		return false
	}
	return !slices.Contains(params.ExcludeTypes, def.Type)
}
```

Add `"slices"` to the imports.

##### Task 2 — `internal/faction/engine/relocate/relocate.go`

Charge by `CoinCostMode`. Add after `matchesFilters`:

```go
// totalCost is what a relocation of moved assets charges: CoinCost per asset, or
// CoinCost once under flat mode (Transit Web: "for 1 Coin, any number").
func totalCost(params *domain.RelocateParams, moved int) int {
	if params.CoinCostMode == domain.RelocateCostFlat {
		return params.CoinCost
	}
	return params.CoinCost * moved
}
```

**Find:**
```go
	if request.Faction.Coin < params.CoinCost*len(orders) {
		return nil, fmt.Errorf("relocate: insufficient Coin: need %d, have %d", params.CoinCost*len(orders), request.Faction.Coin)
	}

	var mutations []domain.Mutation
	for _, order := range orders {
```
**Replace with:**
```go
	if cost := totalCost(params, len(orders)); request.Faction.Coin < cost {
		return nil, fmt.Errorf("relocate: insufficient Coin: need %d, have %d", cost, request.Faction.Coin)
	}

	var mutations []domain.Mutation
	for i, order := range orders {
		orderCost := params.CoinCost
		if params.CoinCostMode == domain.RelocateCostFlat && i > 0 {
			orderCost = 0
		}
```

and in the same loop pass `orderCost` in place of `params.CoinCost` to
`BuildRelocateMutations` (Decision 5).

##### Task 3 — `internal/faction/rulebook/rulebook.go`

The shared code now reads the three fields, so an `[ability]` relocate may carry them.

**Find:**
```go
		// deliberate: reject, not ignore, because the action-spent relocate handler does not read these fields yet.
		if len(params.Categories) > 0 || len(params.ExcludeTypes) > 0 || params.CoinCostMode == domain.RelocateCostFlat {
			return domain.AbilityStep{}, fmt.Errorf("relocate categories, exclude_types and coin_cost_mode = %q are not supported on an [ability]", domain.RelocateCostFlat)
		}
		step.Relocate = params
```
**Replace with:**
```go
		step.Relocate = params
```

##### Task 4 — `internal/faction/rulebook/effects_test.go`

The case `"effort-5 relocate fields on an action-spent ability"` now loads. Move it out
of the error table into a success test,
`TestConvertAbilityStep_RelocateCarriesEffortFiveFields`. It asserts
`Categories == [Wealth]` on the parsed step, plus `exclude_types` and
`coin_cost_mode = "flat"` in the same body.

##### Task 5 — `internal/faction/engine/relocate/relocate_test.go`

`relocRulebook` gains categories: `SF` is Force Special Forces, `MU` is Force Military
Unit. Add `WF` (Wealth Facility), `CS` (Cunning Starship) and `WS` (Wealth Starship).
Add:

- `TestMatchesFilters` — table: no filter passes anything (and an unknown def); type
  match and mismatch; category in and out of list; excluded type; Transit Web's set
  (`Categories [Cunning, Wealth]`, `ExcludeTypes [Starship]`) passes `WF` and refuses
  `CS`, `WS` and `SF`; an unknown def with any filter set fails.
- `TestEligibleAssets_SkipsAssetUnderOrders` — an in-range asset with a non-nil
  `CurrentOrder` is not a candidate.
- `TestRun_FlatCostChargesOnce` — two orders under `flat`, `CoinCost 1`, faction Coin
  1. One `CoinDelta -1` in total, after the first `AssetMoved`.
- `TestRun_FlatCostInsufficient` — Coin 0 under `flat`: error, no mutations.
- The existing per-asset tests stay green unchanged (zero `CoinCostMode` is per asset).

##### Commit message

```
feat: relocate honours category, excluded type and flat cost

- candidate filter reads categories and excluded types
- skip assets under a movement order
- flat cost mode charges the relocation once
- the loader accepts the three fields on an [ability] relocate
```

---

### Phase 2 — Free movement

Wire the granter into the movement phase, then author Transit Web.

#### Commit 3 — `feat: wire granted movement abilities into the movement phase`

##### Task 1 — `internal/faction/engine/hooks/rule_modifier.go`

**Find:**
```go
// MovementAbility describes an extra movement capability granted by a tag or
// asset effect.
type MovementAbility struct {
	Name        string
	Description string
}
```
**Replace with:**
```go
// MovementAbility is a free relocation a granter offers its faction this turn.
// Source is the granting asset; range is measured from it.
type MovementAbility struct {
	Source   *domain.Asset
	Relocate domain.RelocateParams
}
```

**Find:**
```go
// AssetMovementGranter reports additional movement abilities conferred on an
// asset by a tag or special effect (Cat 4).
type AssetMovementGranter interface {
	GrantMovementAbilities(asset *domain.Asset) []MovementAbility
}
```
**Replace with:**
```go
// AssetMovementGranter reports the free movement a tag or special effect offers
// faction this turn (Cat 4). A granter asked for a faction it does not serve
// returns nothing.
type AssetMovementGranter interface {
	GrantMovementAbilities(faction *domain.Faction) []MovementAbility
}
```

##### Task 2 — `internal/faction/engine/hooks/registry.go`

Add beside `WorldTechLevelModifiersForAsset`, matching its form:

```go
// AssetMovementGrantersForAsset returns the Cat 4 movement-granter hooks registered against one asset instance.
func (registry *Registry) AssetMovementGrantersForAsset(factionID, assetInstanceID string) []RegisteredAssetMovementGranter {
	return registry.assetMovementGranters[AssetScope(factionID, assetInstanceID)]
}
```

`AssetMovementGrantersFor` stays. The resolver calls it with `""` for the global and
faction buckets, as `ResolveWorldTechLevel` does.

##### Task 3 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
// GrantedMovementAbilities collects extra movement abilities granted to asset
// by all registered AssetMovementGranters (global + faction-scoped +
// asset-scoped). Returns nil when registry is nil or no granters match.
func GrantedMovementAbilities(registry *hooks.Registry, asset *domain.Asset) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(asset.OwnerID, asset.ID) {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(asset)...)
	}
	return abilities
}
```
**Replace with:**
```go
// GrantedMovementAbilities collects the free movement offered to faction: its
// global and faction granters, then those registered against each asset it owns,
// in asset ID order. A grant serves its owner only, so a rival's assets are never
// asked. Returns nil when registry is nil or no granter offers anything.
func GrantedMovementAbilities(registry *hooks.Registry, faction *domain.Faction) []hooks.MovementAbility {
	if registry == nil {
		return nil
	}
	var abilities []hooks.MovementAbility
	for _, registered := range registry.AssetMovementGrantersFor(faction.ID, "") {
		abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.AssetMovementGrantersForAsset(faction.ID, asset.ID) {
			abilities = append(abilities, registered.Hook.GrantMovementAbilities(faction)...)
		}
	}
	return abilities
}
```

##### Task 4 — `internal/faction/engine/relocate/relocate.go`

The free harness must not prompt a faction that cannot pay (Decision 3). Add the
sentinel beside `ErrNoCandidates`:

```go
// ErrInsufficientCoin reports that the faction cannot pay for the relocation.
var ErrInsufficientCoin = errors.New("insufficient Coin")
```

Before the `SelectRelocateOrders` call, add:

```go
	if request.Faction.Coin < params.CoinCost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, params.CoinCost, request.Faction.Coin)
	}
```

and change the post-selection cost error to wrap it:

```go
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, cost, request.Faction.Coin)
```

The text keeps "insufficient Coin: need N, have M". Add `TestRun_UnaffordableDoesNotPrompt`
(Coin 0, `CoinCost 1`: `errors.Is(err, ErrInsufficientCoin)`, collector not called).
Update `TestRelocate_OtherSingle_InsufficientCoin` to assert the sentinel.

##### Task 5 — `internal/faction/engine/collector.go`

**Find:**
```go
	SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
}
```
**Replace with:**
```go
	SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error)
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error)
}
```

Update the doc comment's "four orchestrator-owned input prompts" to name the prompts
without a count, or correct the count.

##### Task 6 — `internal/faction/tui/adapter/action_collector.go`, `phase_collector.go`

Hoist the body of `actionCollector.SelectRelocateOrders` into a method on `*Adapter`
(in `action_collector.go`, where the body is today):

```go
// selectRelocateOrders asks the relocate overlay. The action collector (a spent
// relocate ability) and the phase collector (a free movement grant) share it.
func (a *Adapter) selectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	// body moved verbatim from actionCollector.SelectRelocateOrders
}

func (a *actionCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	return a.adapter.selectRelocateOrders(acting, candidates, candidateWorlds, maxAssets)
}
```

In `phase_collector.go`:

```go
func (p *phaseCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	return p.adapter.selectRelocateOrders(acting, candidates, candidateWorlds, maxAssets)
}
```

##### Task 7 — `internal/faction/tui/adapter/smoke.go`

Add to `SmokePhaseCollector`, after `SelectTransportCargo`:

```go
func (s *SmokePhaseCollector) SelectRelocateOrders(_ *domain.Asset, _ []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
	return nil, nil
}
```

The testharness `ScriptedCollector` already has the method; no change.

##### Task 8 — `internal/faction/engine/orchestrator_test.go`

`minimalPhaseCollector` gains a field
`selectRelocateOrders func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)`
and the method that calls it, or returns `nil, nil`, matching `SelectTransportCargo`.

##### Task 9 — `internal/faction/engine/orchestrator.go`

Add the grant runner after `prepareMovementDecisions`'s helpers:

```go
// runGrantedMovement offers each free movement grant once, after the drift
// decisions apply: cargo loaded this phase is already in transit, so no grant can
// move it a second time. Each grant is its own batch, so the next grant sees the
// Coin and locations the last one changed. A grant with nothing to move, that the
// faction cannot pay for, or that the GM declines is skipped.
func (e *Engine) runGrantedMovement(
	faction *domain.Faction,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	collector PhaseCollector,
	observer TurnObserver,
	phaseLog *slog.Logger,
) ([]domain.Mutation, error) {
	var resolved []domain.Mutation
	for _, grant := range dispatch.GrantedMovementAbilities(e.Hooks, faction) {
		params := grant.Relocate
		mutations, err := relocate.Run(relocate.Request{
			Faction:   faction,
			Acting:    grant.Source,
			Params:    &params,
			World:     e.World,
			Rulebook:  e.Rulebook,
			Collector: collector,
		})
		if errors.Is(err, relocate.ErrInsufficientCoin) {
			phaseLog.Warn("granted movement skipped", "source", grant.Source.ID, "err", err)
			observer.OnError(faction, err)
			continue
		}
		// deliberate: Esc declines this grant, not the turn, because the decision batch has already applied and must still resolve.
		if errors.Is(err, relocate.ErrNoCandidates) || errors.Is(err, action.ErrNoSelection) || errors.Is(err, action.ErrTurnCanceled) {
			continue
		}
		if err != nil {
			return resolved, fmt.Errorf("granted movement from %q: %w", grant.Source.ID, err)
		}
		mutations, err = dispatch.MutationReactors(e.Hooks, faction, mutations, factionState, e.Rulebook)
		if err != nil {
			return resolved, err
		}
		if err := e.applyAndRecord(factionState, faction, mutations, paths, phaseLog); err != nil {
			return resolved, err
		}
		resolved = append(resolved, mutations...)
	}
	return resolved, nil
}
```

Add `"errors"` and the `relocate` import.

##### Task 10 — `internal/faction/engine/orchestrator.go`, `runMovementPhase`

**Find:**
```go
	observer.OnMovementResolved(faction, decisionMutations)
```
**Replace with:**
```go
	grantedMutations, err := e.runGrantedMovement(faction, factionState, paths, collectors.Phase, observer, phaseLog)
	if err != nil {
		if factionErrors.IsRecoverable(err) {
			phaseLog.Warn("recoverable failure", "err", err)
			observer.OnError(faction, err)
			return nil // continue past this sub-engine call
		}
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return err
	}

	observer.OnMovementResolved(faction, append(decisionMutations, grantedMutations...))
```

The error branch matches the phase's existing boilerplate. The early `return nil`
branches above it are unchanged. A recoverable decision failure still skips the
grants, as it skips the checkpoint today.

##### Task 11 — tests

- **`hooks/dispatch/rules_test.go`** — `TestGrantedMovementAbilities_WalksFactionThenAssets`:
  a stub granter at faction scope and two at asset scope (asset IDs `b`, `a`). Result
  order is faction, `a`, `b`. A granter registered on a rival's asset is not asked.
  Nil registry returns nil.
- **`hooks/registry_test.go`** — `AssetMovementGrantersForAsset` returns only that
  asset's bucket.
- **`orchestrator_test.go`** — package `engine`. Build an `Engine` with a stub granter
  registered on alpha's asset that returns one `MovementAbility` (other-scope,
  `MaxHex 3`, `CoinCost 1`, flat). Use a world router that puts every world in range,
  and a `minimalPhaseCollector` whose `selectRelocateOrders` moves one candidate.
  Cases:
  - `TestRunGrantedMovement_AppliesEachGrant` — the asset is at its destination
    after the call, and Coin dropped by 1.
  - `TestRunGrantedMovement_DeclineSkips` — the collector returns no orders: no error,
    no mutations.
  - `TestRunGrantedMovement_UnaffordableSkipsWithoutPrompt` — Coin 0: the collector is
    not called, the observer gets one error, and the call returns no error.
  - `TestRunGrantedMovement_CargoLoadedThisPhaseIsNotOffered` — an asset whose
    location has no `WorldID` (in transit as cargo) is not among the candidates.
  Re-ground the `Engine` fields, the test observer and the history path against the
  existing `orchestrator_test.go` / `orchestrator_apply_test.go` fixtures before
  writing.

##### Task 12 — `docs/architecture/engine/hooks.md`

Reconciliation item: the page says the movement phase calls `GrantedMovementAbilities`.
That becomes true in this commit. Re-read the resolver paragraph (around line 122) and
make it state:
- `GrantedMovementAbilities` asks the faction's own hooks, then each of its assets', in
  the same way as `ResolveWorldTechLevel`;
- the movement phase runs each grant once, through the shared `relocate` package, after
  the drift decisions apply.

Update the `AssetMovementGranter` signature wherever the page quotes it. Re-read
`docs/architecture/engine/world-movement.md` and `turn-pipeline.md`. Where they list
the movement phase's steps, add "granted moves" after the decision batch.

##### Commit message

```
feat: wire granted movement abilities into the movement phase

- a movement grant is faction-level and carries its source asset
- the resolver walks the faction's hooks, then each owned asset's
- the phase collector gains the relocate prompt, shared with the action collector
- the movement phase runs each grant once, after the decisions apply
- relocate checks the cost of one move before it prompts
- correct the hooks page on the movement resolver
```

#### Commit 4 — `feat: grant_movement effect`

##### Task 1 — `internal/faction/engine/effect/effects/grant_movement.go` (new)

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantMovementRegistrar reads owner = "own" as "the grant serves the declaring
// asset's faction", and scope = "faction" as "the faction's own assets move". The
// range and the asset filters come from the relocate block, measured from the
// declaring asset.
var grantMovementRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerMovementOptionsQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerGrantMovement,
}

func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialGrantMovement)+":"+ctx.Def.ID,
		&movementGranter{
			owner:    ctx.Faction,
			assetID:  ctx.Asset.ID,
			relocate: ctx.Entry.GrantMovement.Relocate,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner    *domain.Faction
	assetID  string
	relocate domain.RelocateParams
}

func (granter *movementGranter) GrantMovementAbilities(faction *domain.Faction) []hooks.MovementAbility {
	if faction.ID != granter.owner.ID {
		return nil
	}
	asset := granter.owner.Assets[granter.assetID]
	if !active(asset) {
		return nil
	}
	return []hooks.MovementAbility{{Source: asset, Relocate: granter.relocate}}
}
```

##### Task 2 — `internal/faction/engine/effect/effects/registrars.go`

Add `domain.SpecialGrantMovement: grantMovementRegistrar,` after
`SpecialRestrictAttack`.

##### Task 3 — asset data (both `rulebooks/swn/assets/wealth_assets.toml` and `campaigns/test-camp/rulebook/assets/wealth_assets.toml`)

After `[assets.TransitWeb.attack]`:

```toml
[[assets.TransitWeb.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"
[assets.TransitWeb.effects.filters]
owner = "own"
scope = "faction"
[assets.TransitWeb.effects.grant_movement.relocate]
scope          = "other"
range          = "hexes"
max_hex        = 3
coin_cost      = 1
coin_cost_mode = "flat"
categories     = ["Cunning", "Wealth"]
exclude_types  = ["Starship"]
```

The description is not edited (Out of Scope).

##### Task 4 — `internal/faction/engine/effect/effects/grant_movement_test.go` (new)

`package effects_test`. Register through `effects.Registrars()[domain.SpecialGrantMovement]`
and read back with `registry.AssetMovementGrantersForAsset("alpha", "web")`.

- **Registers at asset scope.** One granter at `AssetScope(alpha, web)`; none at
  `FactionScope(alpha)`.
- **Grants its relocate block from its own asset.** Asked for alpha: one ability,
  `Source` is the live asset, `Relocate` equals the entry's params.
- **Rival faction gets nothing.** Asked for beta: nil.
- **Inactive asset grants nothing.** Over `inactiveStates`.
- **Removed asset grants nothing.** Delete it from `alpha.Assets` after registration:
  nil.

##### Task 5 — `internal/faction/engine/testharness/harness.go`

The stub map puts no worlds in range (`AllWorlds` returns nil), so a relocate scenario
cannot run. Add a `Worlds []string` field to `StubSpatialMap`:

```go
// Worlds, when set, are the worlds AllWorlds returns. Every stub location shares the
// zero hex, so each one is at distance 0 from every other and in range of any relocate.
Worlds []string
```

`AllWorlds` returns a `stubLocation` per entry. `DistancesFrom` returns
`map[spatial.RegionHex]int{{}: 0}` when `Worlds` is set, and `nil, nil` otherwise.
Expose the map on the `Harness` if it is not already (`SpatialMap` field). Add
`DefTransitWeb = "SWN-W7-003"` to the constants.

##### Task 6 — `internal/faction/engine/testharness/scenarios/effects_test.go`

`TestEffects_TransitWebMovesForOneCoin`: `h.SpatialMap.Worlds = ["Krylos", "Tartarus"]`.
Alpha owns on Tartarus:
- an active Transit Web;
- a Franchise (Wealth Facility);
- a Party Machine (Cunning Logistics Facility);
- a Security Personnel (Force Military Unit).

`SelectRelocateOrdersFn` asserts the candidates are exactly the Franchise and the Party
Machine, then moves both to Krylos. Run one faction turn with `skipAction`. Assert:
- both assets are on Krylos;
- alpha's Coin dropped by exactly 1;
- one `asset_moved` per asset in history.

Re-ground on how the harness handles world names and `AddFaction`'s homeworld before
writing.

##### Task 7 — `docs/architecture/engine/ability-catalog.md`

Reconciliation item: the catalog files free `relocate-other` as a Cat 3
`MutationReactor` (line 157). It is a Cat 4 `AssetMovementGranter`, run by the movement
phase through the shared `relocate` package. Mark Transit Web shipped (✅), as Party
Machine's row is. Re-read the near-miss note at lines 227–228. Where it says both
surfaces need a sibling shape, state that both now run through `engine/relocate`.

##### Commit message

```
feat: grant_movement effect

- grant_movement registrar as an asset-scoped movement granter
- grant lapses when its asset is inactive or gone
- author Transit Web
- harness stub map can put worlds in range
- refile free relocate-other as cat 4 in the catalog
```

---

### Phase 3 — Purchase route

#### Commit 5 — `feat: on_purchase effect route and apply_stealth`

##### Task 1 — `internal/faction/engine/effect/effect.go`

Add after `FireTurnStart`:

```go
// FirePurchase runs every on_purchase entry of a just-bought asset through its
// registrar's phase route. A direct definition read, not a hook: the asset was
// bought mid-turn, so it has no hooks until the next cycle. No active check: a
// bought asset is never ready, and its purchase effect is part of what was bought.
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	def := rulebook.Assets[asset.DefinitionID]
	if def == nil {
		return nil
	}
	var mutations []domain.Mutation
	for _, entry := range def.Effects {
		if entry.Trigger != domain.TriggerOnPurchase {
			continue
		}
		// deliberate: no nil check on Fire, because RegisterEffectsList rejects a phase-route entry without one.
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
		})...)
	}
	return mutations
}
```

##### Task 2 — `internal/faction/engine/orchestrator.go`

Add beside the action-phase helpers:

```go
// firePurchaseEffects runs the on_purchase effects of every asset the action
// bought, in batch order. The asset is the copy the AssetAdded carries, because
// it is not in state until the batch applies.
func (e *Engine) firePurchaseEffects(faction *domain.Faction, actionMutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	var mutations []domain.Mutation
	for _, mutation := range actionMutations {
		added, ok := mutation.(domain.AssetAdded)
		// deliberate: whitelist the purchase causes, not exclude refit, so a new AssetAdded emitter fires nothing until it is judged a purchase.
		if !ok || (added.Cause != "buy" && added.Cause != "ability") {
			continue
		}
		mutations = append(mutations, e.Effect.FirePurchase(faction, &added.Asset, factionState, e.Rulebook, e.Rand)...)
	}
	return mutations
}
```

In `runActionPhase`, **find:**
```go
	goalMutations := e.Goal.UpdateProgress(faction.ID, actionMutations, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))
```
**Replace with:**
```go
	actionMutations = append(actionMutations, e.firePurchaseEffects(faction, actionMutations, factionState)...)
	goalMutations := e.Goal.UpdateProgress(faction.ID, actionMutations, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))
```

Re-ground on the name of the effects engine field (`e.Effect`) and the roller
(`e.Rand`) against `runBookkeepingPhase`'s `FireTurnStart` call.

##### Task 3 — `internal/faction/engine/effect/effects/apply_stealth.go` (new)

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

var applyStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Fire: fireApplyStealth,
}

func fireApplyStealth(ctx effect.FireContext) []domain.Mutation {
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialApplyStealth),
		CausedByFactionID: ctx.Faction.ID,
	}}
}
```

Add `domain.SpecialApplyStealth: applyStealthRegistrar,` to `registrars.go`.

Update the `AssetStealthApplied` doc comment in `domain/mutation.go`: it is emitted by
Buy Asset for a Stealth purchase, *and* by `apply_stealth` when a stealthed-on-purchase
asset is bought.

##### Task 4 — `internal/faction/engine/effect/effects/clear_stealth.go`

Tripwire Cells learn the purchase case (Decision 13). Replace `OnMutations` so each
match yields the arriving asset, then one shared gate:

```go
func (reactor *clearStealthReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	owner := factionState.Factions[reactor.ownerFactionID]
	if owner == nil || !active(owner.Assets[reactor.assetID]) {
		return nil
	}
	worldID := owner.Assets[reactor.assetID].Location.WorldID
	if worldID == "" {
		return nil
	}
	var extra []domain.Mutation
	for _, mutation := range mutations {
		factionID, assetID, ok := stealthArrival(mutation, mutations, factionState, worldID)
		if !ok || factionID == owner.ID {
			continue
		}
		target := factionState.Factions[factionID]
		if target == nil {
			continue
		}
		// deliberate: compare true, not the state's flag, because a bought asset is stealthed by this batch and is not in state yet.
		if reactor.stealthy != nil && !*reactor.stealthy {
			continue
		}
		if reactor.contest != nil && !contestPasses(*reactor.contest, owner, target, reactor.roller) {
			continue
		}
		extra = append(extra, domain.AssetStealthCleared{
			FactionID:         factionID,
			AssetID:           assetID,
			Cause:             string(domain.SpecialClearStealth),
			CausedByFactionID: owner.ID,
		})
	}
	return extra
}

// stealthArrival reports the asset a mutation brings onto worldID stealthed. A
// landing is an AssetMoved into the world from elsewhere by an asset already
// stealthy (Decision Record 10). A purchase is an AssetStealthApplied whose asset
// the same batch adds on the world; a stealth applied to an asset already there is
// not an arrival.
func stealthArrival(mutation domain.Mutation, batch []domain.Mutation, factionState *state.FactionState, worldID string) (string, string, bool) {
	switch m := mutation.(type) {
	case domain.AssetMoved:
		if m.ToLocation.WorldID != worldID || m.FromLocation.WorldID == worldID {
			return "", "", false
		}
		faction := factionState.Factions[m.FactionID]
		if faction == nil || faction.Assets[m.AssetID] == nil || !faction.Assets[m.AssetID].Stealthy {
			return "", "", false
		}
		return m.FactionID, m.AssetID, true
	case domain.AssetStealthApplied:
		for _, other := range batch {
			added, ok := other.(domain.AssetAdded)
			if ok && added.FactionID == m.FactionID && added.Asset.ID == m.AssetID && added.Asset.Location.WorldID == worldID {
				return m.FactionID, m.AssetID, true
			}
		}
	}
	return "", "", false
}
```

Behaviour on the landing path is unchanged. Every asset that reaches the filter
check is stealthy, so the old `Stealthy != *reactor.stealthy` comparison equals
`!*reactor.stealthy`. Update the registrar's doc comment: drop "Arrival by purchase is
Effort 5's…" and say that `stealth_arrival` covers a landing and a stealthed purchase.

##### Task 5 — asset data (both `rulebooks/swn/assets/force_assets.toml` and `campaigns/test-camp/rulebook/assets/force_assets.toml`)

After `[assets.PsychicAssassins.attack]`:

```toml
[[assets.PsychicAssassins.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"
[assets.PsychicAssassins.effects.filters]
owner = "own"
scope = "self"
```

##### Task 6 — tests

- **`effect/effects/apply_stealth_test.go`** (new, `package effects_test`):
  - fires `AssetStealthApplied` for the asset with cause `apply_stealth`;
  - emits nothing for an asset already stealthy;
  - fires with the asset not `Ready` (no active check).
- **`effect/effect_test.go`** — `TestFirePurchase_FiresOnPurchaseEntriesOnly`: a def
  with one `on_purchase` and one `turn_start` entry. Only the purchase entry's
  mutations return. An unknown definition returns nil.
- **`orchestrator_test.go`** — `TestFirePurchaseEffects_CausesWhitelisted`: an action
  batch with `AssetAdded` causes `buy`, `ability` and `refit`, all of a Psychic
  Assassins def. Two `AssetStealthApplied` result, for the `buy` and `ability` assets
  only.
- **`effect/effects/clear_stealth_test.go`** — add:
  - a rival's `AssetAdded` on the Cells' world plus `AssetStealthApplied` in one batch:
    cleared after a won contest;
  - the same on another world: nothing;
  - `AssetStealthApplied` with no `AssetAdded` in the batch (Stealth on an existing
    asset): nothing;
  - the owner's own purchase: nothing.
  The existing landing tests stay green.
- **`testharness/scenarios/effects_test.go`** —
  `TestEffects_PsychicAssassinsBoughtStealthed`. Alpha (Force 5+, enough Coin) buys
  Psychic Assassins on its homeworld through `SelectBuyOrderFn`. Assert:
  - the new asset is stealthy after the turn;
  - history holds `asset_stealth_applied` with cause `apply_stealth`.

  A second case puts beta's active Tripwire Cells on the same world, with a
  `FixedRoller` set to win their contest. Assert the asset ends unstealthed and history
  holds `asset_stealth_cleared` with cause `clear_stealth`. Add
  `DefPsychicAssassins = "SWN-F5-003"` and `DefTripwireCells = "SWN-C4-003"` to the
  harness constants.

##### Task 7 — `docs/architecture/`

- `engine/ability-catalog.md` — mark Psychic Assassins shipped through the purchase
  phase route, beside Party Machine's turn-start row (line 163).
- Re-read `engine/effect-mutation.md` and `engine/turn-pipeline.md`. Wherever the
  turn-start phase route is described, add the purchase route: fired from the action
  phase's scan of `AssetAdded` (causes `buy`, `ability`), before goal progress and the
  reactors.

##### Commit message

```
feat: on_purchase effect route and apply_stealth

- FirePurchase reads a bought asset's on_purchase entries directly
- the action phase fires it for buy and enabler purchases, not refits
- apply_stealth emits AssetStealthApplied after the AssetAdded
- Tripwire Cells test a rival asset bought stealthed on their world
- author Psychic Assassins
```
