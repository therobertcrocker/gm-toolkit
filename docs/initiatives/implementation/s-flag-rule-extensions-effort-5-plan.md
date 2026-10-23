# S-flag Rule Extensions (A.001.6) — Effort 5 Plan: Movement permission

## In Plain Words

- **What:** a restricted (`P`) asset that lands on a world, at the end of a drift order
  or by a relocate, asks that world's government for permission. A refused drift
  order waits one hex short of the world and asks again next turn; a refused relocate
  leaves that one asset where it is. Deep Strike Landers, Blockade Runners and
  stealthed assets skip the question. Lobbyists can contest any permission granted to
  a rival on their world and take it back. Transport Lockdown makes a rival pay 1d4
  Coin and wait a turn for every arrival on its world, and bars its relocates there.
- **Why:** the book's permission rule covers moving onto a world, not only buying
  there, and four assets bend that rule. None of them works until arrivals ask.
- **How it fits:** the last effort of A.001.6 (see the
  [overview plan](s-flag-rule-extensions-plan.md)). It extends Effort 4's permission
  request from purchases to landings, and uses Effort 1's conditions and no-damage
  attack.

## Context / Goal

Build the effort the overview plan scopes as "Movement permission", on branch
`feature/s-flag-rule-extensions`.

- Discovery: [s-flag-rule-extensions-discovery.md](../discovery/s-flag-rule-extensions-discovery.md)
  — Permission Model (movement rows and "What bends it"); D-24, D-25, D-27 to D-34,
  D-36, D-41.
- Overview plan: [s-flag-rule-extensions-plan.md](s-flag-rule-extensions-plan.md) —
  Decisions 12, 13, 15, 16, 18, 22 and 24 bind here. This plan reshapes Decision 12
  (Decisions 8 and 9 below) and confirms Decision 24 (Decision 13).
- Effort 4 plan: [completed/s-flag-rule-extensions-effort-4-plan.md](completed/s-flag-rule-extensions-effort-4-plan.md)
  — Decisions 8, 9, 10 and 15 (the request, its place in each path, the refusal
  record, the homeworld clause).
- Rules source: `docs/rules/swn-assets.md` (Lobbyists, Transport Lockdown, Deep Strike
  Landers, Blockade Runners, Stealth) and the book's permission rule (page 216,
  quoted in the discovery).

## Decisions Ratified in Planning

Re-grounding against the code found four gaps (F1–F4). Robert made the rules calls
named below on 2026-10-22.

1. **The arrival check is an orchestrator step between the tick and its reactors.**
   `Engine.checkArrivals` (new `engine/arrival.go`) reads each `MovementOrderCompleted`
   the tick emits and asks for every asset aboard: the carrier, then its cargo in
   manifest order, each for its current owner. The first refusal ends the asking and
   replaces the completion with `PermissionRefused` and `MovementOrderHeld`; the
   transport reactor's `Held` arm then moves the cargo to the edge with its carrier
   (D-24, D-27, D-28). F1: five of the six SWN `P` assets have speed 0, so most
   arrival asks are for cargo, not for the carrier. Rejected: a check inside
   `world.TickMovementOrders` (the world engine would import `dispatch` and a
   collector).
2. **`TickMovementOrders` walks assets in ID order.** It walks `faction.Assets`, a
   map, today. The arrival prompts and Lockdown's 1d4 rolls follow the tick's order,
   so map order would make both non-deterministic (overview Shared Context,
   Determinism).
3. **`MovementOrderHeld{FactionID, AssetID, WorldID, NewStepIdx, RegionHex, Reason, Cause, CausedByFactionID}`.**
   `WorldID` is the destination, for the digest. `Reason` is a `domain.HoldReason`:
   `permission_refused` (Commit 1), `lockdown_paid` and `lockdown_unpaid` (Commit 5).
   The hold step is `max(len(Path)-2, 0)`; its apply arm sets the step and hex as
   `Progressed` does (overview Decisions 16 and 22).
4. **An arrival asks through `dispatch.RequestArrivalPermission`, beside
   `RequestPermission`; both share `askGovernment`.** F4: `RequestPermission` reads
   `NeedsPurchasePermission` and skips the homeworld, which are purchase rules. The
   arrival test is `AssetDefinition.NeedsArrivalPermission(asset)`: the `P` flag, and
   the asset is not stealthed (D-25, Stealth's "that permission may be foregone").
   Capital Fleet's `purchase_permission` covers raising only. The homeworld still
   asks (Effort 4 Decision 15). `askGovernment` keeps D-6 (a ruler's own request) and
   the grant hook (Popular Movement's "purchase or movement requests").
5. **`PermissionRequest` gains `Cause`.** `domain.PermissionCauseArrival` and
   `PermissionCauseRelocate`, with `domain.IsLanding(cause)`, let the prompt say
   "land" for an arrival and "raise" for a purchase. The refusal digest line reads the
   same helper.
6. **`PhaseCollector` embeds `hooks.PermissionCollector`** (D-41; deferred to here by
   Effort 4 Decision 8). The adapter's ask moves to `Adapter.requestPermission`, which
   both collectors call, as `confirmPayment` is shared.
7. **The transport bypass is `TransportProfile.IgnoresPermission`
   (`ignores_permission`), and it covers the carrier and its cargo** (D-25). Deep
   Strike Landers move "any one non-Starship asset … even if the local planetary
   government objects"; Blockade Runners move "itself or any one Military Unit or
   Special Forces … even past planetary government opposition".
8. **A relocate asks for each order after the pick, and a refusal drops that order
   only.** Robert's call. F3: `SelectRelocateOrders` takes one world list for every
   candidate, so a per-asset permission cannot filter it. `relocate.Run` asks for each
   picked order whose asset needs permission. A refused order leaves a
   `PermissionRefused` (cause `relocate`) and does not move; the other orders move;
   under the flat cost mode the first order that moves carries the cost. If every
   order is refused, the relocation emits only the records, and the action (or the
   grant) is still used (D-35). The cost check before the pick is unchanged: the full
   pick's cost is the most the relocation can charge. Reshapes overview Decision 12's
   "Relocate's destination list drops a world whose key is spent" (Decision 9).
   Rejected: dropping the world from the list (blocks non-`P` assets too); a world
   list per asset (changes the collector and the relocate overlay for one SWN case,
   Transit Web moving Mercenaries).
9. **Only a Lobbyists withdrawal sets the per-turn mark.** Robert's call (overview
   Decision 12's open question). The key is `permission_withdrawn:<worldID>:<withdrawerID>`
   on the requester's `HookBudgets` (`dispatch.PermissionWithdrawnKey`; the
   withdrawer was added in execution, Decision Record 4), renamed from
   Decision 12's `permission_refused:` because a plain refusal no longer sets it.
   `askGovernment` reads it first: a later request for that world in the same turn is
   refused with no ask (`RefusalBarred`), in every path. A plain government refusal
   sets nothing; the government may be asked again. Inside Use Asset Ability the mark
   lands on the action's view, which is enough: the view lives until the action
   ends, and nothing after the action phase asks.
10. **Lobbyists is a Cat 4 member, `PermissionContester`.** Two methods:
    `OfferContest(query) *PermissionContest` (nil when the grant is not its to
    contest) and `WinsContest(query) bool`. The dispatcher asks the Lobbyists' owner
    through `hooks.PermissionCollector.ConfirmPermissionContest`, then calls
    `WinsContest`. The hook rolls bare dice with `contestPasses` and the roller it kept
    at registration, as every effect contest does (Effort 3 Decision 12). Data:
    trigger `permission_granted`, effect `withdraw_permission`, owner `rival`, scope
    `world`, and a `contest` gate (roller `self`, Cunning vs. Cunning) that the
    registrar requires. Resolves the overview open question "the Lobbyists contest's
    prompt to its owner inside the request path". Bare dice is Robert's call
    (2026-10-22): today it skips only Fanatical's reroll of 1s and Book of Secrets'
    reroll. Moving every effect contest (Tripwire Cells, Panopticon Matrix, Lobbyists)
    onto roll hooks together is filed as `R.023` (Commit 4 Task 17). Rejected: rolling
    through `RollWithHooks` now (needs `hooks.Collector` in `PhaseCollector` and a
    rulebook at all five call sites, and leaves Lobbyists the one hooked effect
    contest); the dispatcher rolling (it would duplicate `contestPasses`).
11. **Which grants a Lobbyists contests.** Every grant `askGovernment` makes: a ruler's
    own request (D-6), a grant hook, or a yes (D-29). Not a purchase on the buyer's
    homeworld, which is no grant (Robert's call, Effort 4 Decision 15), and not a
    request for an asset that needs no permission. Each active rival Lobbyists on the
    world is offered its contest, rival factions in ID order, then assets in ID order;
    the first win withdraws the grant and ends the walk (Robert's call). The
    requester's own Lobbyists are never offered, nor are the Lobbyists of a ruler
    that said yes (Decision Record 3).
12. **`PermissionRefused.Reason` names the kind of refusal:** the zero value,
    `RefusalDenied`, is a government's no; `withdrawn` is a Lobbyists' win; `barred` is
    a request after a withdrawal this turn. `omitempty`. A plain refusal keeps the
    zero value, so Effort 4's records, history and tests read unchanged. On a
    withdrawal or a barred request, `DecidedByFactionID` is the Lobbyists' owner
    (Decision Record 4).
13. **Transport Lockdown's lock-out is the condition `locked_out` on the Lockdown asset,
    with `Condition.TargetFactionID` and `Condition.WorldID`** (the world of the hit;
    added in execution, Decision Record 5). Overview Decision 24 confirmed.
    `Condition.Is` compares the target faction and the world too, so a lock-out
    against a second rival, or on a second world, does not replace the first.
    `domain.LockedOut(factions, worldID, factionID)` reads it: an active asset of
    another owner on the world that carries `locked_out` from itself against the
    faction, struck on that world. The record lives and dies with its source (D-31,
    D-36); a capture into the struck faction lapses it when read, as Seditionists'
    block does (Effort 3 Decision 10).
14. **Lockdown at a drift arrival: checked before permission, one 1d4 per order,
    charged automatically.** Robert's calls. The arrival check rolls 1d4 for an order
    whose faction is locked out of the destination. If the faction can pay, it emits a
    `CoinDelta` (cause `transport_lockdown`) and `MovementOrderHeld{Reason:
    lockdown_paid}`, whose apply arm sets `MovementOrder.LockdownPaid`. The next tick
    skips the lockdown and asks for permission at the real landing. A faction that
    cannot pay holds with `lockdown_unpaid` and pays at a later tick (overview Decision
    15). The check keeps a running total of the Coin spent in the tick, so two arrivals
    cannot both spend the same Coin. A revised order is a new order and pays again.
    Resolves the overview open question "Transport Lockdown's paid-and-waiting mark".
15. **Deep Strike Landers skip Lockdown; Blockade Runners and stealthed assets do not.**
    Robert's call. Lockdown is shipping pressure, not the government, so the
    permission bypasses do not cover it, except for Deep Strike Landers. A second
    profile field, `TransportProfile.IgnoresLockdown` (`ignores_lockdown`), on Deep
    Strike Landers only.
16. **D-33 holds: a relocate cannot land on a world the faction is locked out of.**
    Robert re-confirmed it. `relocate.Run` drops a locked world from `candidateWorlds`,
    for every asset of the locked faction (D-32).
17. **Lockdown strikes with the book's special attack.** `[attack]` Cunning vs.
    Cunning with no `damage` (overview Decision 3); `lock_out` is an `own_attack_hit`
    reactor (owner `own`, scope `self`), as Treachery's `capture_target` is. A hit on a
    rival Base counts: the book says "against a rival faction".
18. **Digest and TUI wording (overview Decision 13).** `MovementOrderHeld` is a hold
    line on the order owner's beat. The refusal line says "raise" or "land", and has
    its own wording for `withdrawn` and `barred`. `locked_out` is a condition clause on
    the cross event between the attacker and the struck faction; the condition sits on
    the attacker's own asset, so the digest keys it on `TargetFactionID`. The TUI shows
    the hold and its reason, and the refusal's reason.
19. **Five commits in four phases.** Phase 1 is the arrival check and the transport
    bypass; Phase 2 the relocate request; Phase 3 Lobbyists; Phase 4 Transport
    Lockdown. The bypass rides in Phase 1 because the arrival check is its only reader.
    Phases 1, 3 and 4 pass the 800-line budget: the bulk is literal test code, and no
    green seam splits them further (a new mutation ships its digest and TUI lines in
    its commit, overview Decision 13; a new collector method needs every
    implementation in the same build).

## Decision Record — Execution

1. **Commit 3: the two Seductress dispatch tests pass `relocRulebook()`, not `nil`.**
   A self-scope relocate did not read the rulebook before; `askLanding` now reads
   the moved asset's definition. Production always passes a rulebook, so the tests
   change and `askLanding` gets no nil guard. Not in the plan.
2. **Commit 3: the Decision 8 Key Decision bullet goes after the whole harness
   bullet**, not straight after "pure parts." as Task 6 shows: that bullet runs on
   (the import note, the sentinels, the provenance), and the plan's anchor would
   have split it.
3. **Commit 4: a ruler that said yes is not offered its own grant to contest.**
   Robert's call at review; amends Decision 11. `askGovernment` passes the ruler
   whose driver said yes to `contestGrant` as `grantedBy`, and the walk skips that
   faction. A grant that went around the ruler (a grant hook, Popular Movement) is
   still open to the ruler's Lobbyists; an independent world's GM yes skips no one.
   New test: `TestRequestPermission_TheRulerDoesNotContestItsOwnYes`.
4. **Commit 4: the withdrawal mark names the withdrawer, and a barred refusal
   carries it.** Robert's call at review; amends Decisions 9 and 12. The key is
   `permission_withdrawn:<worldID>:<withdrawerID>` (`HookBudgets` is a counter map,
   so the faction rides in the key). `withdrawnBy` checks each faction's key in ID
   order, and a `RefusalBarred` record sets `DecidedByFactionID` to the withdrawer.
   The barred digest line names it: "{decider}'s lobbying kept {faction} from asking
   again to …".
5. **Commit 5: a lock-out is pinned to the world of the hit.** Robert's calls in
   execution; amends Decision 13. Found by a side review: `AssetMoved` keeps an
   asset's conditions, and a Lockdown (a Cunning Tactic) can move as cargo or by
   Transit Web, so the plan's `LockedOut`, which read only the Lockdown's current
   world, let the lock follow the asset to a new world with no new attack. That reads
   against D-31 ("active on that world") and the book ("onto that planet").
   `Condition` gains `WorldID`; `lock_out` sets it to the Lockdown's world at the
   hit; `LockedOut` matches only a lock struck on the asked world. A Lockdown that
   leaves makes its lock sleep, and a return wakes it (checked when read, D-31).
   `Condition.Is` compares `WorldID`, so hits on two worlds keep two locks (Robert's
   call: "keep both locks"). The code in Tasks 1, 6, 11 and 12 shows the pre-pin
   form; the source is the truth. New `TestLockedOut` cases: a moved Lockdown frees
   both its new world and the world of the hit.

---

## Shared Context

- **Every commit ends green:** `go build ./... && go test ./internal/faction/...`.
- **Mock regeneration.** After a change to `action.Collector` or an interface it
  embeds, run `go generate ./internal/faction/engine/action/...`.
- **`gofmt -w` every touched Go file.** A task that adds a line to an aligned block
  (a const list, a struct, a map literal) gives the line, not the realigned block.
- **A new `AskKind`** goes before `AskKindCount` and gets an `askEntries` row;
  `TestAskEntriesCoverEveryKind` catches a missing row.
- **Test patterns.** Engine tests: `minimalPhaseCollector`, `errorObserver` and
  `stubWorldMap` (`orchestrator_test.go`). Dispatch tests: `permissionState`,
  `stubPermissionCollector`. Effects tests: `effects.Registrars()` through
  `effect.New(...)`, `seqRoller`. Relocate tests: `relocEngine`, `relocRulebook`,
  `mocks.NewMockCollector`. Mutation tests: `newMutationTestState`. Digest tests:
  `minState`, `eventRec`, `mutRec`, `buildOne`, `findBeat`.
- **Causes.** A permission request names its path: `buy`, `refit`, `ability`
  (Effort 4), `arrival` and `relocate` (here). Lockdown's charge uses
  `transport_lockdown`; its condition uses the effect name, `lock_out`.
- **The SWN rulebook only.** Asset TOML changes go to `rulebooks/swn/`. The test
  campaign's copy is regenerated after the arc (overview Decision 8).

## Out of Scope

- Roll hooks on the Lobbyists contest (Decision 10; filed as `R.023`).
- A world list per relocate candidate (Decision 8).
- A bribe as a modeled option (D-34) and per-faction drivers (`F.020`, `F.025`).
- A view that shows the GM which factions a Lockdown holds out.
- The digest's `unknown mutation` notes for the five existing `movement_order_*`
  types (found in planning: `build.go` has no case for them, and the renderer drops
  notes). A candidate backlog row; Robert decides at pre-merge.

---

## Work Breakdown

**Overview**

- **Phase 1 — Arrival permission**
  - Commit 1 — `feat: arrival permission and hold` (17 tasks)
  - Commit 2 — `feat: permission bypasses` (5 tasks)
- **Phase 2 — Relocate permission**
  - Commit 3 — `feat: relocate permission` (6 tasks)
- **Phase 3 — Lobbyists**
  - Commit 4 — `feat: Lobbyists` (17 tasks)
- **Phase 4 — Transport Lockdown**
  - Commit 5 — `feat: Transport Lockdown` (14 tasks)

### Phase 1 — Arrival permission

**In plain words:** a drift order that reaches its world now asks that world's
government for every restricted asset aboard; a refusal stops the order one hex short,
and it asks again next turn. Deep Strike Landers and Blockade Runners skip the question.

#### Commit 1 — `feat: arrival permission and hold`

##### Task 1 — `internal/faction/domain/mutation.go`

Add the landing causes above `PermissionRefused`, and widen its comment.

**Find:**
```go
// PermissionRefused records a refused permission request. It changes no state:
// the action it ends bought nothing (D-35, overview Decision 22).
// DecidedByFactionID is the world's ruler, empty when the GM decided for an
// independent world. Cause is the purchase path: buy, refit or ability.
```
**Replace with:**
```go
// The permission causes that ask to land an asset, not to raise one. The purchase
// causes are buy, refit and ability.
const (
	PermissionCauseArrival  = "arrival"
	PermissionCauseRelocate = "relocate"
)

// IsLanding reports whether a permission cause asks to land an asset on a world.
func IsLanding(cause string) bool {
	return cause == PermissionCauseArrival || cause == PermissionCauseRelocate
}

// PermissionRefused records a refused permission request. It changes no state:
// the purchase or landing it refuses did not happen (D-35, overview Decision 22).
// DecidedByFactionID is the world's ruler, empty when the GM decided for an
// independent world. Cause is the path: buy, refit or ability for a purchase,
// arrival or relocate for a landing.
```

Add the hold after `MovementOrderCompleted`.

**Find:**
```go
func (m MovementOrderCompleted) Type() string { return "movement_order_completed" }
```
**Replace with:**
```go
func (m MovementOrderCompleted) Type() string { return "movement_order_completed" }

// HoldReason says why an arriving order holds at the edge of its destination.
type HoldReason string

const (
	// HoldRefused: the destination's government refused an asset aboard (D-27).
	HoldRefused HoldReason = "permission_refused"
)

// MovementOrderHeld keeps an order that would complete at the last hex before its
// destination, WorldID. The order stays open and asks again at the next tick
// (D-27). Its apply arm sets the step and hex as MovementOrderProgressed does.
type MovementOrderHeld struct {
	FactionID         string            `json:"faction_id"`
	AssetID           string            `json:"asset_id"`
	WorldID           string            `json:"world_id"`
	NewStepIdx        int               `json:"new_step_idx"`
	RegionHex         spatial.RegionHex `json:"region_hex"`
	Reason            HoldReason        `json:"reason"`
	Cause             string            `json:"cause"`
	CausedByFactionID string            `json:"caused_by_faction_id"`
}

func (m MovementOrderHeld) Type() string { return "movement_order_held" }
```

##### Task 2 — `internal/faction/domain/asset.go`

**Find:**
```go
// HalfPrice is half the listed cost, rounded down, but at least 1 Coin: the price
```
**Replace with:**
```go
// NeedsArrivalPermission reports whether landing asset on a world asks the world's
// government. A P asset does (the book: "purchase them or move them onto the
// planet"), unless it is stealthed (Stealth: "that permission may be foregone",
// D-25). Capital Fleet's PurchasePermission covers raising only.
func (def *AssetDefinition) NeedsArrivalPermission(asset *Asset) bool {
	return def.HasFlag(FlagPermission) && !asset.Stealthy
}

// HalfPrice is half the listed cost, rounded down, but at least 1 Coin: the price
```

##### Task 3 — `internal/faction/engine/mutation/mutation.go`

Add the hold's apply arm before the revision's.

**Find:**
```go
		case domain.MovementOrderRevised:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					order := v.NewOrder
```
**Replace with:**
```go
		case domain.MovementOrderHeld:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{RegionHex: v.RegionHex}
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderRevised:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					order := v.NewOrder
```

##### Task 4 — `internal/faction/engine/hooks/collector.go`

**Find:**
```go
// PermissionRequest asks a world's government whether Requester may raise Def
// there. Ruler is nil for a world no faction governs; the GM answers for it (D-4).
type PermissionRequest struct {
	Requester *domain.Faction
	Ruler     *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
}

// PermissionCollector asks for a government's permission. action.Collector embeds
// it; PhaseCollector will when arrivals ask (A.001.6 Effort 5, D-41).
```
**Replace with:**
```go
// PermissionRequest asks a world's government whether Requester may raise or land
// Def there. Ruler is nil for a world no faction governs; the GM answers for it
// (D-4). Cause names the path, so a prompt can say "raise" or "land"
// (domain.IsLanding).
type PermissionRequest struct {
	Requester *domain.Faction
	Ruler     *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
	Cause     string
}

// PermissionCollector asks for a government's permission. action.Collector and
// the engine's PhaseCollector both embed it (D-41): purchases and relocates ask in
// an action, drift arrivals in the movement phase.
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/permission.go`

Rewritten — full contents:
```go
package dispatch

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RequestPermission asks worldID's government whether requester may raise def
// there, and returns the refusal record, or nil when the purchase may go ahead. A
// definition that needs no permission is not asked about, nor is a purchase on the
// requester's homeworld (Effort 4 Decision 15). cause names the purchase path on
// the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// RequestArrivalPermission asks worldID's government whether requester may land
// asset there, at the end of a drift order or by a relocate (cause), and returns
// the refusal record, or nil when the asset may land. An asset that needs no
// permission is not asked about (AssetDefinition.NeedsArrivalPermission).
func RequestArrivalPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsArrivalPermission(asset) {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// askGovernment grants a ruler's own request at once (D-6), and one a grant hook
// of the requester serves (Popular Movement). Otherwise it asks the ruler's driver,
// or the GM for an independent world (D-4).
func askGovernment(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	rulerID := factionState.Ruler(worldID)
	if rulerID == requester.ID {
		return nil, nil
	}
	if PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def}) {
		return nil, nil
	}
	request := hooks.PermissionRequest{Requester: requester, WorldID: worldID, Def: def, Cause: cause}
	if rulerID != "" {
		request.Ruler = factionState.Factions[rulerID]
	}
	granted, err := collector.RequestPermission(request)
	if err != nil {
		return nil, fmt.Errorf("permission request: %w", err)
	}
	if granted {
		return nil, nil
	}
	decidedBy := ""
	if request.Ruler != nil {
		decidedBy = request.Ruler.ID
	}
	return &domain.PermissionRefused{
		FactionID:          requester.ID,
		WorldID:            worldID,
		DefinitionID:       def.ID,
		DecidedByFactionID: decidedBy,
		Cause:              cause,
		CausedByFactionID:  requester.ID,
	}, nil
}
```

##### Task 6 — `internal/faction/engine/collector.go`

**Find:**
```go
// PhaseCollector abstracts the orchestrator-owned input prompts: the phase
// checkpoints, action choice, stat raise, movement decisions, transport cargo and
// the relocate orders of a free movement grant, and the pay-to-lift question at turn start.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type PhaseCollector interface {
	hooks.PaymentCollector
```
**Replace with:**
```go
// PhaseCollector abstracts the orchestrator-owned input prompts: the phase
// checkpoints, action choice, stat raise, movement decisions, transport cargo and
// the relocate orders of a free movement grant, the pay-to-lift question at turn
// start, and an arriving asset's permission request.
// The GM implementation uses interactive prompts; AI and test implementations
// use scripted or goal-driven logic.
type PhaseCollector interface {
	hooks.PaymentCollector
	hooks.PermissionCollector
```

##### Task 7 — `internal/faction/engine/world/movement.go`

In `TickMovementOrders`, walk the assets in ID order (Decision 2).

**Find:**
```go
	for _, asset := range faction.Assets {
		if asset.CurrentOrder == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return nil, fmt.Errorf("TickMovementOrders: unknown asset definition %q", asset.DefinitionID)
		}
```
**Replace with:**
```go
	for _, asset := range domain.SortedAssets(faction) {
		if asset.CurrentOrder == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return nil, fmt.Errorf("TickMovementOrders: unknown asset definition %q", asset.DefinitionID)
		}
```

##### Task 8 — `internal/faction/engine/arrival.go` (new)

Full contents:
```go
package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// checkArrivals asks, for each order the tick completes, whether everything aboard
// may land, and returns the tick's mutations with each refused completion replaced
// by the refusal and a hold at the edge (D-24, D-27). The government asked is the
// one ruling at arrival. It runs between the tick and its reactors, so the
// transport reactor moves the cargo to the edge with its carrier (D-28).
func (e *Engine) checkArrivals(faction *domain.Faction, factionState *state.FactionState, tickMutations []domain.Mutation, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	checked := make([]domain.Mutation, 0, len(tickMutations))
	for _, mutation := range tickMutations {
		completed, ok := mutation.(domain.MovementOrderCompleted)
		if !ok {
			checked = append(checked, mutation)
			continue
		}
		hold, err := e.arrivalHold(faction, factionState, completed, collector)
		if err != nil {
			return nil, err
		}
		if hold == nil {
			checked = append(checked, mutation)
			continue
		}
		checked = append(checked, hold...)
	}
	return checked, nil
}

// arrivalHold asks for each asset aboard completed's order, the carrier first, then
// its cargo in manifest order, each for its own owner. The first refusal holds the
// whole order and ends the asking (D-28). It returns the refusal and the hold that
// replace the completion, or nil when everything aboard may land.
func (e *Engine) arrivalHold(faction *domain.Faction, factionState *state.FactionState, completed domain.MovementOrderCompleted, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	carrier := faction.Assets[completed.AssetID]
	if carrier == nil || carrier.CurrentOrder == nil {
		return nil, nil
	}
	order := carrier.CurrentOrder
	for _, asset := range aboard(factionState, carrier) {
		def := e.Rulebook.Assets[asset.DefinitionID]
		requester := factionState.Factions[asset.OwnerID]
		if def == nil || requester == nil {
			continue
		}
		refusal, err := dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)
		if err != nil {
			return nil, err
		}
		if refusal != nil {
			return []domain.Mutation{*refusal, holdAtEdge(completed, order, domain.HoldRefused)}, nil
		}
	}
	return nil, nil
}

// aboard returns the carrier and the cargo its order carries, each found under its
// current owner: a capture can make a rider's owner differ from the carrier's.
func aboard(factionState *state.FactionState, carrier *domain.Asset) []*domain.Asset {
	assets := []*domain.Asset{carrier}
	for _, cargoID := range carrier.CurrentOrder.CargoAssetIDs {
		if owner := domain.AssetOwner(factionState.Factions, cargoID); owner != nil {
			assets = append(assets, owner.Assets[cargoID])
		}
	}
	return assets
}

// holdAtEdge replaces completed with a hold at the last hex before the destination
// (overview Decision 16). Path[0] is the asset's own hex, so on a two-hex path the
// hold is the in-orbit state right after issue.
func holdAtEdge(completed domain.MovementOrderCompleted, order *domain.MovementOrder, reason domain.HoldReason) domain.MovementOrderHeld {
	step := max(len(order.Path)-2, 0)
	return domain.MovementOrderHeld{
		FactionID:         completed.FactionID,
		AssetID:           completed.AssetID,
		WorldID:           order.Destination.WorldID,
		NewStepIdx:        step,
		RegionHex:         order.Path[step],
		Reason:            reason,
		Cause:             domain.CauseMovementTick,
		CausedByFactionID: completed.CausedByFactionID,
	}
}
```

##### Task 9 — `internal/faction/engine/orchestrator.go`

In `runMovementPhase`, check the arrivals before the tick's reactors.

**Find:**
```go
	if len(tickMutations) > 0 {
		tickMutations, err = dispatch.MutationReactors(e.Hooks, faction, tickMutations, factionState, e.Rulebook)
```
**Replace with:**
```go
	if len(tickMutations) > 0 {
		tickMutations, err = e.checkArrivals(faction, factionState, tickMutations, collectors.Phase)
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
		tickMutations, err = dispatch.MutationReactors(e.Hooks, faction, tickMutations, factionState, e.Rulebook)
```

##### Task 10 — `internal/faction/engine/effect/effects/transport.go`

Cargo follows a held carrier to the edge.

**Find:**
```go
		case domain.MovementOrderCompleted:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				m.FinalLocation,
				"transport_completed",
			)...)
```
**Replace with:**
```go
		case domain.MovementOrderCompleted:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				m.FinalLocation,
				"transport_completed",
			)...)

		case domain.MovementOrderHeld:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				domain.Location{RegionHex: m.RegionHex},
				"transport_held",
			)...)
```

##### Task 11 — `internal/faction/narrative/digest/types.go`, `build.go`, `headline.go`

`types.go` — the beat gains holds, and a refusal knows whether it was a landing.

**Find:**
```go
	Refusals     []RefusalEvent
	Notes        []string
}

// RefusalEvent is a purchase a government refused. DecidedBy is the ruler's name,
// empty when the GM decided for an independent world.
type RefusalEvent struct {
	AssetName string
	World     string
	DecidedBy string
}
```
**Replace with:**
```go
	Refusals     []RefusalEvent
	Holds        []HoldEvent
	Notes        []string
}

// RefusalEvent is a purchase or a landing a government refused. DecidedBy is the
// ruler's name, empty when the GM decided for an independent world. Landing is
// true for an arrival or a relocate.
type RefusalEvent struct {
	AssetName string
	World     string
	DecidedBy string
	Landing   bool
}

// HoldEvent is an arriving order held one hex short of World (D-27).
type HoldEvent struct {
	AssetName string
	World     string
	Reason    domain.HoldReason
}
```

`build.go` — the refusal reads its cause, and the hold gets its case.

**Find:**
```go
				refused.Refusals = append(refused.Refusals, RefusalEvent{
					AssetName: resolveDefinitionName(m.DefinitionID, rulebook),
					World:     m.WorldID,
					DecidedBy: decidedBy,
				})
```
**Replace with:**
```go
				refused.Refusals = append(refused.Refusals, RefusalEvent{
					AssetName: resolveDefinitionName(m.DefinitionID, rulebook),
					World:     m.WorldID,
					DecidedBy: decidedBy,
					Landing:   domain.IsLanding(m.Cause),
				})

			case "movement_order_held":
				var m domain.MovementOrderHeld
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				beat.Holds = append(beat.Holds, HoldEvent{
					AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
					World:     m.WorldID,
					Reason:    m.Reason,
				})
```

`headline.go` — a hold makes a faction active.

**Find:**
```go
		len(beat.GoalEvents) > 0 || len(beat.StealthOps) > 0 || len(beat.Refusals) > 0 {
```
**Replace with:**
```go
		len(beat.GoalEvents) > 0 || len(beat.StealthOps) > 0 || len(beat.Refusals) > 0 ||
		len(beat.Holds) > 0 {
```

##### Task 12 — `internal/faction/narrative/wire_renderer.go`, `wire_templates.go`

`wire_renderer.go` — render holds after refusals.

**Find:**
```go
	for _, refusal := range beat.Refusals {
		sb.WriteString(renderRefusal(refusal, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
```
**Replace with:**
```go
	for _, refusal := range beat.Refusals {
		sb.WriteString(renderRefusal(refusal, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, hold := range beat.Holds {
		sb.WriteString(renderHold(hold, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
```

**Find:**
```go
	tmpl := pick(rng, refusalTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", refusal.AssetName)
	result = strings.ReplaceAll(result, "{world}", refusal.World)
	return strings.ReplaceAll(result, "{decider}", decider)
}
```
**Replace with:**
```go
	verb := "raise"
	if refusal.Landing {
		verb = "land"
	}
	tmpl := pick(rng, refusalTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{verb}", verb)
	result = strings.ReplaceAll(result, "{asset}", refusal.AssetName)
	result = strings.ReplaceAll(result, "{world}", refusal.World)
	return strings.ReplaceAll(result, "{decider}", decider)
}

func renderHold(hold digest.HoldEvent, factionName string, rng *rand.Rand) string {
	pool, ok := holdTemplates[hold.Reason]
	if !ok {
		pool = holdTemplates[domain.HoldRefused]
	}
	result := strings.ReplaceAll(pick(rng, pool), "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", hold.AssetName)
	return strings.ReplaceAll(result, "{world}", hold.World)
}
```
**Find:**
```go
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"
)
```
**Replace with:**
```go
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"
)
```

`wire_templates.go`:

**Find:**
```go
var refusalTemplates = []string{
	"{decider} refused {faction} permission to raise {asset} on {world}.",
	"Permission denied: {decider} would not let {faction} raise {asset} on {world}.",
}
```
**Replace with:**
```go
var refusalTemplates = []string{
	"{decider} refused {faction} permission to {verb} {asset} on {world}.",
	"Permission denied: {decider} would not let {faction} {verb} {asset} on {world}.",
}

var holdTemplates = map[domain.HoldReason][]string{
	domain.HoldRefused: {
		"{faction}'s {asset} holds in orbit over {world}, waiting for leave to land.",
		"Turned back at the edge of {world}, {faction}'s {asset} waits to ask again.",
	},
}
```

##### Task 13 — `internal/faction/tui/views/turn/execution/format.go`

**Find:**
```go
	case domain.MovementOrderCompleted:
		return "move completed", true
```
**Replace with:**
```go
	case domain.MovementOrderCompleted:
		return "move completed", true
	case domain.MovementOrderHeld:
		return fmt.Sprintf("move held short of %s (%s)", mut.WorldID, mut.Reason), true
```

##### Task 14 — `internal/faction/tui/adapter/action_collector.go`, `phase_collector.go`, `smoke.go`, `internal/faction/tui/views/turn/overlay/permission.go`

`action_collector.go` — the ask moves to the adapter, so both collectors share it.

**Find:**
```go
// RequestPermission asks the world's ruler, or the GM for an independent world; the ask
// names the ruler, who is often not the acting faction.
func (a *actionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	worldName := request.WorldID
	if loc, ok := a.adapter.engine.World.Location(request.WorldID); ok {
		worldName = loc.Name()
	}
	raw, err := a.adapter.ask(AskRequestPermission, request.Ruler, RequestPermissionPayload{Request: request, WorldName: worldName})
	if err != nil {
		return false, err
	}
	granted, _ := raw.(bool)
	return granted, nil
}
```
**Replace with:**
```go
func (a *actionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	return a.adapter.requestPermission(request)
}

// requestPermission asks the world's ruler, or the GM for an independent world; the
// ask names the ruler, who is often not the acting faction. The action collector
// (purchases, relocates) and the phase collector (arrivals) share it.
func (a *Adapter) requestPermission(request hooks.PermissionRequest) (bool, error) {
	worldName := request.WorldID
	if loc, ok := a.engine.World.Location(request.WorldID); ok {
		worldName = loc.Name()
	}
	raw, err := a.ask(AskRequestPermission, request.Ruler, RequestPermissionPayload{Request: request, WorldName: worldName})
	if err != nil {
		return false, err
	}
	granted, _ := raw.(bool)
	return granted, nil
}
```

`phase_collector.go`:

**Find:**
```go
func (p *phaseCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return p.adapter.confirmPayment(offer)
}
```
**Replace with:**
```go
func (p *phaseCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return p.adapter.confirmPayment(offer)
}

func (p *phaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	return p.adapter.requestPermission(request)
}
```

`smoke.go`:

**Find:**
```go
func (s *SmokePhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil // the dry run declines every lift
}
```
**Replace with:**
```go
func (s *SmokePhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil // the dry run declines every lift
}

func (s *SmokePhaseCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, nil // the smoke faction has no assets, so nothing arrives to ask
}
```

`overlay/permission.go` — rewritten, full contents:
```go
package overlay

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewRequestPermission asks a world's government whether a faction may raise an
// asset there, or land one. Its header names who decides: the ruler, or the GM for
// a world no faction governs. Esc does nothing: a cancel would drop the action, and
// with it every earlier asset of a Use Asset Ability. Answer: bool.
func NewRequestPermission(request hooks.PermissionRequest, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of the GM (no faction governs %s)", worldName)
	if request.Ruler != nil {
		header = fmt.Sprintf("Asked of %s, the government of %s", request.Ruler.Name, worldName)
	}
	return NoCancelConfirm{NewConfirm(permissionTitle(request, worldName), header, "Grant", "Refuse")}
}

func permissionTitle(request hooks.PermissionRequest, worldName string) string {
	return fmt.Sprintf("Permit %s to %s %s on %s?", request.Requester.Name, permissionVerb(request.Cause), request.Def.Name, worldName)
}

// permissionVerb words a permission cause: "raise" for a purchase, "land" for an
// arrival or a relocate.
func permissionVerb(cause string) string {
	if domain.IsLanding(cause) {
		return "land"
	}
	return "raise"
}
```

##### Task 15 — engine tests

`internal/faction/domain/asset_test.go` — append:
```go
func TestAssetDefinition_NeedsArrivalPermission(t *testing.T) {
	restricted := &domain.AssetDefinition{ID: "troops", Flags: []domain.AssetFlag{domain.FlagPermission}}
	raiseOnly := &domain.AssetDefinition{ID: "fleet", PurchasePermission: true}
	cases := []struct {
		name  string
		def   *domain.AssetDefinition
		asset *domain.Asset
		want  bool
	}{
		{"a P asset", restricted, &domain.Asset{}, true},
		{"a stealthed P asset", restricted, &domain.Asset{Stealthy: true}, false},
		{"raise-only permission", raiseOnly, &domain.Asset{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.def.NeedsArrivalPermission(tc.asset); got != tc.want {
				t.Errorf("NeedsArrivalPermission = %v, want %v", got, tc.want)
			}
		})
	}
}
```

`internal/faction/engine/mutation/mutation_test.go` — append:
```go
func TestMutationEngine_MovementOrderHeld(t *testing.T) {
	s := newMutationTestState()
	hexes := []spatial.RegionHex{{RegionID: "r", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 1}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 2}}}
	s.Factions["f1"].Assets["a1"].CurrentOrder = &domain.MovementOrder{AssetID: "a1", Path: hexes, Destination: domain.Location{WorldID: "Krylos", RegionHex: hexes[2]}}
	me := New(slog.New(slog.DiscardHandler))

	held := domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: hexes[1], Reason: domain.HoldRefused}
	if err := me.Apply(s, []domain.Mutation{held}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	asset := s.Factions["f1"].Assets["a1"]
	if asset.CurrentOrder == nil || asset.CurrentOrder.StepIdx != 1 {
		t.Errorf("order = %+v, want it open at step 1", asset.CurrentOrder)
	}
	if asset.Location != (domain.Location{RegionHex: hexes[1]}) {
		t.Errorf("Location = %+v, want the edge hex with no world", asset.Location)
	}
}

func TestMutationEngine_MovementOrderHeldMisses(t *testing.T) {
	cases := []struct {
		name string
		held domain.MovementOrderHeld
	}{
		{"unknown faction", domain.MovementOrderHeld{FactionID: "missing", AssetID: "a1"}},
		{"unknown asset", domain.MovementOrderHeld{FactionID: "f1", AssetID: "missing"}},
		{"asset with no order", domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(slog.New(slog.DiscardHandler)).Apply(newMutationTestState(), []domain.Mutation{tc.held}, slog.New(slog.DiscardHandler))
			if err == nil || len(err.Misses) != 1 {
				t.Errorf("Apply = %v, want one miss", err)
			}
		})
	}
}
```
**Find** (the test file's imports):
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)
```

`internal/faction/engine/hooks/dispatch/permission_test.go` — every existing expected
request carries the cause the test passes. In `TestRequestPermission`, replace each of
the three expected requests:

**Find:**
```go
			wantAsked: []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
		},
```
**Replace with:**
```go
			wantAsked: []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef, Cause: "buy"}},
		},
```

**Find:**
```go
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
```
**Replace with:**
```go
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef, Cause: "buy"}},
```

**Find:**
```go
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, WorldID: "Hadrian", Def: permissionDef}},
```
**Replace with:**
```go
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, WorldID: "Hadrian", Def: permissionDef, Cause: "buy"}},
```

Append:
```go
func TestRequestArrivalPermission(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	raiseOnlyDef := &domain.AssetDefinition{ID: "fleet", PurchasePermission: true}
	factionState, requester, ruler := permissionState()

	cases := []struct {
		name        string
		def         *domain.AssetDefinition
		asset       *domain.Asset
		world       string
		wantAsked   []hooks.PermissionRequest
		wantRefusal *domain.PermissionRefused
	}{
		{name: "raise-only permission", def: raiseOnlyDef, asset: &domain.Asset{}, world: "Krylos"},
		{name: "a stealthed P asset", def: permissionDef, asset: &domain.Asset{Stealthy: true}, world: "Krylos"},
		{name: "the requester rules the world", def: permissionDef, asset: &domain.Asset{}, world: "Tartarus"},
		{
			name:        "the homeworld still asks",
			def:         permissionDef,
			asset:       &domain.Asset{},
			world:       "Anchorage",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Anchorage", Def: permissionDef, Cause: "arrival"}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Anchorage", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "arrival", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			collector := &stubPermissionCollector{}

			refusal, err := RequestArrivalPermission(nil, collector, factionState, requester, tc.asset, tc.def, tc.world, domain.PermissionCauseArrival)

			if err != nil {
				t.Fatalf("RequestArrivalPermission: %v", err)
			}
			if !reflect.DeepEqual(collector.asked, tc.wantAsked) {
				t.Errorf("asked = %+v, want %+v", collector.asked, tc.wantAsked)
			}
			if !reflect.DeepEqual(refusal, tc.wantRefusal) {
				t.Errorf("refusal = %+v, want %+v", refusal, tc.wantRefusal)
			}
		})
	}
}
```

`internal/faction/engine/action/actions/buy_asset_test.go` — the two exact-match
expectations carry the cause.

**Find:**
```go
					RequestPermission(hooks.PermissionRequest{Requester: faction, Ruler: ruler, WorldID: tc.world, Def: def}).
```
**Replace with:**
```go
					RequestPermission(hooks.PermissionRequest{Requester: faction, Ruler: ruler, WorldID: tc.world, Def: def, Cause: "buy"}).
```

**Find:**
```go
	collector.EXPECT().RequestPermission(hooks.PermissionRequest{Requester: faction, WorldID: "Tartarus", Def: def}).Return(false, nil)
```
**Replace with:**
```go
	collector.EXPECT().RequestPermission(hooks.PermissionRequest{Requester: faction, WorldID: "Tartarus", Def: def, Cause: "buy"}).Return(false, nil)
```

`internal/faction/engine/orchestrator_test.go` — `minimalPhaseCollector` answers
permission requests; it grants unless a test scripts a refusal.

**Find:**
```go
	selectRelocateOrders func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)
}
```
**Replace with:**
```go
	selectRelocateOrders func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)
	requestPermission    func(hooks.PermissionRequest) (bool, error)
}
```

**Find:**
```go
func (c *minimalPhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil
}
```
**Replace with:**
```go
func (c *minimalPhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil
}
func (c *minimalPhaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	if c.requestPermission != nil {
		return c.requestPermission(request)
	}
	return true, nil
}
```

`internal/faction/engine/arrival_test.go` (new) — full contents:
```go
package engine

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func arrivalHex(q int) spatial.RegionHex {
	return spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: q}}
}

func arrivalRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"ship": {ID: "ship", Name: "Ship", Speed: 2, Transport: &domain.TransportProfile{
			MaxCargo: 1, CargoTypes: []domain.AssetType{domain.TypeMilitaryUnit},
		}},
		"troops": {ID: "troops", Name: "Troops", Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
		"militia": {ID: "militia", Name: "Militia", Type: domain.TypeMilitaryUnit},
		"mercs":   {ID: "mercs", Name: "Mercs", Speed: 2, Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
}

// arrivalFixture builds alpha with a carrier "t1" (definition carrierDef) whose
// order completes onto Krylos along pathLen hexes, carrying "c1" (definition
// cargoDef) when cargoDef is set. Bravo governs Krylos.
func arrivalFixture(carrierDef, cargoDef string, pathLen int) (*Engine, *state.FactionState, *domain.Faction, domain.MovementOrderCompleted) {
	path := make([]spatial.RegionHex, pathLen)
	for i := range path {
		path[i] = arrivalHex(i)
	}
	destination := domain.Location{WorldID: "Krylos", RegionHex: path[pathLen-1]}
	carrier := &domain.Asset{ID: "t1", DefinitionID: carrierDef, OwnerID: "alpha", Location: domain.Location{RegionHex: path[pathLen-2]},
		CurrentOrder: &domain.MovementOrder{AssetID: "t1", Destination: destination, Path: path, StepIdx: pathLen - 2}}
	alpha := &domain.Faction{ID: "alpha", Name: "Alpha", Coin: 5, Homeworld: domain.Location{WorldID: "Anchorage"}, Assets: map[string]*domain.Asset{"t1": carrier}}
	if cargoDef != "" {
		alpha.Assets["c1"] = &domain.Asset{ID: "c1", DefinitionID: cargoDef, OwnerID: "alpha", Location: carrier.Location}
		carrier.CurrentOrder.CargoAssetIDs = []string{"c1"}
	}
	bravo := &domain.Faction{ID: "bravo", Name: "Bravo"}
	factionState := &state.FactionState{
		Factions:    map[string]*domain.Faction{"alpha": alpha, "bravo": bravo},
		Governments: map[string]string{"Krylos": "bravo"},
	}
	discard := slog.New(slog.DiscardHandler)
	eng := &Engine{Rulebook: arrivalRulebook(), Hooks: hooks.NewRegistry(), Mutation: mutation.New(discard)}
	completed := domain.MovementOrderCompleted{FactionID: "alpha", AssetID: "t1", FinalLocation: destination, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"}
	return eng, factionState, alpha, completed
}

func TestCheckArrivals(t *testing.T) {
	cases := []struct {
		name       string
		carrierDef string
		cargoDef   string
		stealthy   bool
		rulesSelf  bool
		grant      bool
		wantAsked  []string
		wantHeld   bool
	}{
		{name: "nothing aboard needs permission", carrierDef: "ship", cargoDef: "militia"},
		{name: "a P cargo is granted", carrierDef: "ship", cargoDef: "troops", grant: true, wantAsked: []string{"troops"}},
		{name: "a P cargo is refused", carrierDef: "ship", cargoDef: "troops", wantAsked: []string{"troops"}, wantHeld: true},
		{name: "a P carrier is refused", carrierDef: "mercs", wantAsked: []string{"mercs"}, wantHeld: true},
		{name: "a stealthed P cargo is not asked", carrierDef: "ship", cargoDef: "troops", stealthy: true},
		{name: "the faction rules the world", carrierDef: "ship", cargoDef: "troops", rulesSelf: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture(tc.carrierDef, tc.cargoDef, 3)
			if cargo := alpha.Assets["c1"]; cargo != nil {
				cargo.Stealthy = tc.stealthy
			}
			if tc.rulesSelf {
				factionState.Governments["Krylos"] = "alpha"
			}
			var asked []string
			collector := &minimalPhaseCollector{requestPermission: func(request hooks.PermissionRequest) (bool, error) {
				asked = append(asked, request.Def.ID)
				if request.Cause != domain.PermissionCauseArrival || request.Ruler == nil || request.Ruler.ID != "bravo" {
					t.Errorf("request = %+v, want an arrival asked of bravo", request)
				}
				return tc.grant, nil
			}}

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if !reflect.DeepEqual(asked, tc.wantAsked) {
				t.Errorf("asked = %v, want %v", asked, tc.wantAsked)
			}
			if !tc.wantHeld {
				if !reflect.DeepEqual(got, []domain.Mutation{completed}) {
					t.Errorf("mutations = %+v, want the completion unchanged", got)
				}
				return
			}
			refused := tc.wantAsked[0]
			want := []domain.Mutation{
				domain.PermissionRefused{FactionID: "alpha", WorldID: "Krylos", DefinitionID: refused, DecidedByFactionID: "bravo", Cause: domain.PermissionCauseArrival, CausedByFactionID: "alpha"},
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldRefused, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_FirstRefusalEndsTheAsking(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "troops", 3)
	asks := 0
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) {
		asks++
		return false, nil
	}}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	if asks != 1 {
		t.Errorf("asked %d times, want 1 (one refusal holds the whole order, D-28)", asks)
	}
	if len(got) != 2 {
		t.Errorf("mutations = %+v, want the refusal and the hold", got)
	}
}

func TestCheckArrivals_TwoHexPathHoldsAtTheStart(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "", 2)
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) { return false, nil }}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	held, ok := got[len(got)-1].(domain.MovementOrderHeld)
	if !ok || held.NewStepIdx != 0 || held.RegionHex != arrivalHex(0) {
		t.Errorf("last mutation = %+v, want a hold at step 0 on the path's first hex", got[len(got)-1])
	}
}

func TestRunMovementPhase_RefusedArrivalHoldsWithItsCargo(t *testing.T) {
	eng, factionState, alpha, _ := arrivalFixture("ship", "troops", 3)
	discard := slog.New(slog.DiscardHandler)
	eng.World = world.NewWithMap(&stubWorldMap{}, discard)
	effects.NewTransportHandler(eng.Rulebook.Assets["ship"]).Apply(alpha, alpha.Assets["t1"], eng.Hooks, nil)
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) { return false, nil }}
	paths := &campaigns.Paths{HistoryPath: t.TempDir() + "/history.jsonl"}

	err := eng.runMovementPhase(alpha, factionState, paths, Collectors{Phase: collector}, &errorObserver{}, discard)

	if err != nil {
		t.Fatalf("runMovementPhase: %v", err)
	}
	carrier, cargo := alpha.Assets["t1"], alpha.Assets["c1"]
	if carrier.CurrentOrder == nil || carrier.CurrentOrder.StepIdx != 1 {
		t.Errorf("carrier order = %+v, want it open at step 1", carrier.CurrentOrder)
	}
	if carrier.Location != (domain.Location{RegionHex: arrivalHex(1)}) {
		t.Errorf("carrier at %+v, want the edge hex in orbit", carrier.Location)
	}
	if cargo.Location != (domain.Location{RegionHex: arrivalHex(1)}) {
		t.Errorf("cargo at %+v, want it held with its carrier (D-28)", cargo.Location)
	}
}
```

`internal/faction/engine/world/movement_test.go` — append:
```go
func TestTickMovementOrders_InAssetIDOrder(t *testing.T) {
	path := []spatial.RegionHex{{RegionID: "void", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "void", Coord: spatial.HexCoord{Q: 1}}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{}}
	for _, id := range []string{"c", "a", "d", "b"} {
		faction.Assets[id] = &domain.Asset{ID: id, OwnerID: "f1", DefinitionID: "test-def",
			CurrentOrder: &domain.MovementOrder{AssetID: id, Path: path}}
	}

	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, reviseRulebook(), slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("TickMovementOrders: %v", err)
	}
	var order []string
	for _, mutation := range mutations {
		order = append(order, mutation.(domain.MovementOrderCompleted).AssetID)
	}
	if want := []string{"a", "b", "c", "d"}; !slices.Equal(order, want) {
		t.Errorf("tick order = %v, want %v", order, want)
	}
}
```
Add `"slices"` to the test file's imports.

`internal/faction/engine/effect/effects/transport_test.go` — append:
```go
func TestTransportReactor_Held(t *testing.T) {
	path := testPath(5)
	transport := &domain.Asset{ID: "transport-1", OwnerID: "alpha",
		CurrentOrder: &domain.MovementOrder{AssetID: "transport-1", Path: path, CargoAssetIDs: []string{"cargo-1"}}}
	cargo := &domain.Asset{ID: "cargo-1", OwnerID: "alpha", Location: domain.Location{RegionHex: path[2]}}

	result := newReactor().OnMutations([]domain.Mutation{
		domain.MovementOrderHeld{FactionID: "alpha", AssetID: "transport-1", WorldID: "Krylos", NewStepIdx: 3, RegionHex: path[3], Reason: domain.HoldRefused},
	}, stateWithTransport(transport, cargo), nil)

	if len(result) != 1 {
		t.Fatalf("result = %+v, want one cargo move", result)
	}
	moved, ok := result[0].(domain.AssetMoved)
	if !ok || moved.AssetID != "cargo-1" || moved.ToLocation != (domain.Location{RegionHex: path[3]}) || moved.Cause != "transport_held" {
		t.Errorf("result[0] = %+v, want cargo-1 moved to the hold hex", result[0])
	}
}
```

##### Task 16 — digest, renderer and TUI tests

`internal/faction/narrative/digest/digest_test.go` — append:
```go
func TestMovementOrderHeldIsAHoldLine(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha"})
	factionState.Factions["f1"].Assets = map[string]*domain.Asset{"t1": {ID: "t1", DefinitionID: "ship"}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"ship":   {ID: "ship", Name: "Freighter"},
		"troops": {ID: "troops", Name: "Troops"},
	}}
	records := []domain.EventRecord{eventRec("f1",
		mutRec(t, "permission_refused", domain.PermissionRefused{
			FactionID: "f1", WorldID: "Krylos", DefinitionID: "troops", Cause: domain.PermissionCauseArrival, CausedByFactionID: "f1",
		}),
		mutRec(t, "movement_order_held", domain.MovementOrderHeld{
			FactionID: "f1", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, Reason: domain.HoldRefused, Cause: domain.CauseMovementTick, CausedByFactionID: "f1",
		}),
	)}

	beat := findBeat(buildOne(t, records, factionState, rb), "f1")

	if beat == nil {
		t.Fatal("f1 is quiet, want a hold to make it active")
	}
	if want := []RefusalEvent{{AssetName: "Troops", World: "Krylos", Landing: true}}; !reflect.DeepEqual(beat.Refusals, want) {
		t.Errorf("Refusals = %+v, want %+v", beat.Refusals, want)
	}
	if want := []HoldEvent{{AssetName: "Freighter", World: "Krylos", Reason: domain.HoldRefused}}; !reflect.DeepEqual(beat.Holds, want) {
		t.Errorf("Holds = %+v, want %+v", beat.Holds, want)
	}
	if len(beat.Notes) != 0 {
		t.Errorf("Notes = %v, want none", beat.Notes)
	}
}
```
`internal/faction/narrative/wire_renderer_test.go` — append:
```go
func TestRenderRefusalAndHold_Landing(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	refusal := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo", Landing: true}, "Alpha", rng)
	if !strings.Contains(refusal, "land Troops") {
		t.Errorf("refusal = %q, want the landing verb", refusal)
	}
	raise := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo"}, "Alpha", rng)
	if !strings.Contains(raise, "raise Troops") {
		t.Errorf("refusal = %q, want the purchase verb", raise)
	}
	hold := renderHold(digest.HoldEvent{AssetName: "Freighter", World: "Krylos", Reason: domain.HoldRefused}, "Alpha", rng)
	if !strings.Contains(hold, "Freighter") || !strings.Contains(hold, "Krylos") {
		t.Errorf("hold = %q, want the asset and the world", hold)
	}
}
```

`internal/faction/tui/views/turn/execution/format_test.go` — add a row after the
`MovementOrderCompleted` row:

**Find:**
```go
		{"MovementOrderCompleted", domain.MovementOrderCompleted{}, "move completed"},
```
**Replace with:**
```go
		{"MovementOrderCompleted", domain.MovementOrderCompleted{}, "move completed"},
		{"MovementOrderHeld", domain.MovementOrderHeld{WorldID: "Krylos", Reason: domain.HoldRefused}, "move held short of Krylos (permission_refused)"},
```

`internal/faction/tui/views/turn/overlay/asset_test.go` — append:
```go
func TestPermissionTitle(t *testing.T) {
	request := hooks.PermissionRequest{
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "troops", Name: "Troops"},
		Cause:     domain.PermissionCauseArrival,
	}
	if got, want := permissionTitle(request, "Krylos"), "Permit Alpha to land Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	request.Cause = "buy"
	if got, want := permissionTitle(request, "Krylos"), "Permit Alpha to raise Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}
```

##### Task 17 — docs

`docs/architecture/engine/world-movement.md` — the tick step gains the arrival check.

**Find:**
```markdown
1. **Tick first.** `TickMovementOrders` advances open orders; its mutations go
   through `MutationReactors` then `applyAndRecord`, and the driver is notified
   via `OnMovementTicked`.
```
**Replace with:**
```markdown
1. **Tick first.** `TickMovementOrders` advances open orders, in asset ID order.
   `checkArrivals` (`engine/arrival.go`) then reads each `MovementOrderCompleted`:
   it asks the destination's government for every `P` asset aboard (the carrier,
   then its cargo), and the first refusal replaces the completion with a
   `PermissionRefused` and a `MovementOrderHeld` one hex short of the world. The
   tick's mutations then go through `MutationReactors` (so held cargo stays with its
   carrier) and `applyAndRecord`, and the driver is notified via `OnMovementTicked`.
```

Add a Key Decision after "Tick before decisions.":

**Find:**
```markdown
- **The movement phase dispatches reactors per batch.**
```
**Replace with:**
```markdown
- **An arrival asks at the tick that would complete it, and a refusal holds at the
  edge.** The government asked is the one ruling when the order arrives, not when it
  was issued, and a held order asks again at its next tick (D-24, D-27). One refusal
  holds the whole order with its cargo, so an immobile rider is never left in space
  (D-28). The hold is its own record, `MovementOrderHeld`, so history never shows a
  move that did not happen. (A.001.6 overview Decisions 16 and 22; Effort 5 Decision 1.)
- **The movement phase dispatches reactors per batch.**
```

`docs/architecture/engine/effect-mutation.md` — the movement catalogue line:

**Find:**
```markdown
- `movement_order_issued` / `_progressed` / `_revised` / `_cancelled` / `_completed` — the in-flight order lifecycle (see [world & movement](world-movement.md))
```
**Replace with:**
```markdown
- `movement_order_issued` / `_progressed` / `_revised` / `_cancelled` / `_completed` — the in-flight order lifecycle (see [world & movement](world-movement.md))
- `movement_order_held` — an order that would complete stops one hex short of its world and stays open: a refused arrival. Its apply arm sets the step and hex as `_progressed` does
```

And the refusal line:

**Find:**
```markdown
- `permission_refused` — record only: a government refused a purchase, which bought nothing
```
**Replace with:**
```markdown
- `permission_refused` — record only: a government refused a purchase (which bought nothing) or a landing (the asset holds or stays)
```

`docs/architecture/engine/orchestrator.md` — the phase collector's prompt list:

**Find:**
```markdown
`PhaseCollector` is the orchestrator's own prompt surface: `ConfirmPayment` (embedded `hooks.PaymentCollector`; the turn-start lift), `AwaitCheckpoint`,
```
**Replace with:**
```markdown
`PhaseCollector` is the orchestrator's own prompt surface: `ConfirmPayment` (embedded `hooks.PaymentCollector`; the turn-start lift), `RequestPermission` (embedded `hooks.PermissionCollector`; an arriving asset's permission request), `AwaitCheckpoint`,
```

`docs/rules/swn-faction-mechanics.md` — the Government Permission paragraph:

**Find:**
```markdown
The campaign's government record names the ruler of each world; that ruler grants or denies this permission there, and the GM decides for a world no faction governs. The **Planetary Government** tag marks a faction that governs at least one world.
```
**Replace with:**
```markdown
The campaign's government record names the ruler of each world; that ruler grants or denies this permission there, and the GM decides for a world no faction governs. The **Planetary Government** tag marks a faction that governs at least one world.

A `P` asset asks when it lands on a world: at the end of a drift order (as cargo or under its own order) or by a relocate. A refused drift order holds one hex short of the world and asks again at its next tick; the whole order holds, cargo included. A stealthed asset lands without asking. A faction never asks on a world it governs, but it does ask to land on its own homeworld when another government holds it.
```

##### Commit message

```
feat: arrival permission and hold

- A drift order that completes asks the destination's government for each
  P asset aboard (carrier, then cargo); a refusal replaces the completion
  with PermissionRefused and MovementOrderHeld one hex short (D-24, D-27, D-28)
- dispatch.RequestArrivalPermission beside RequestPermission, sharing
  askGovernment; a stealthed asset lands without asking (D-25)
- PermissionRequest.Cause words the prompt "raise" or "land"
- PhaseCollector embeds hooks.PermissionCollector (D-41)
- TickMovementOrders walks assets in ID order
- The transport reactor's Held arm keeps cargo with its carrier
- Digest hold line, landing refusal wording, TUI hold line
```

#### Commit 2 — `feat: permission bypasses`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
type TransportProfile struct {
	MaxHex            int
	CoinCost          int
	CargoTypes        []AssetType
	MaxCargo          int
	ExcludeCategories []FactionStat
}
```
**Replace with:**
```go
// TransportProfile is a transport's typed data. IgnoresPermission lets its order
// land the carrier and its cargo without asking the government (D-25).
type TransportProfile struct {
	MaxHex            int
	CoinCost          int
	CargoTypes        []AssetType
	MaxCargo          int
	ExcludeCategories []FactionStat
	IgnoresPermission bool
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
type transportRecord struct {
	MaxHex            int      `toml:"max_hex"`
	CoinCost          int      `toml:"coin_cost"`
	CargoTypes        []string `toml:"cargo_types"`
	MaxCargo          int      `toml:"max_cargo"`
	ExcludeCategories []string `toml:"exclude_categories"`
}
```
**Replace with:**
```go
type transportRecord struct {
	MaxHex            int      `toml:"max_hex"`
	CoinCost          int      `toml:"coin_cost"`
	CargoTypes        []string `toml:"cargo_types"`
	MaxCargo          int      `toml:"max_cargo"`
	ExcludeCategories []string `toml:"exclude_categories"`
	IgnoresPermission bool     `toml:"ignores_permission"`
}
```

**Find:**
```go
	return &domain.TransportProfile{
		MaxHex:            r.MaxHex,
		CoinCost:          r.CoinCost,
		CargoTypes:        cargoTypes,
		MaxCargo:          r.MaxCargo,
		ExcludeCategories: excludeCategories,
	}, nil
```
**Replace with:**
```go
	return &domain.TransportProfile{
		MaxHex:            r.MaxHex,
		CoinCost:          r.CoinCost,
		CargoTypes:        cargoTypes,
		MaxCargo:          r.MaxCargo,
		ExcludeCategories: excludeCategories,
		IgnoresPermission: r.IgnoresPermission,
	}, nil
```

##### Task 3 — `internal/faction/engine/arrival.go`

In `arrivalHold`, a bypassing carrier asks for nothing.

**Find:**
```go
	order := carrier.CurrentOrder
	for _, asset := range aboard(factionState, carrier) {
```
**Replace with:**
```go
	order := carrier.CurrentOrder
	// deliberate: the carrier's bypass covers its cargo, because the book lets these transports land "any one … asset" past the government's objection (D-25).
	if carrierDef := e.Rulebook.Assets[carrier.DefinitionID]; carrierDef != nil && carrierDef.Transport != nil && carrierDef.Transport.IgnoresPermission {
		return nil, nil
	}
	for _, asset := range aboard(factionState, carrier) {
```

##### Task 4 — `rulebooks/swn/assets/force_assets.toml`, `wealth_assets.toml`

**Find** (`force_assets.toml`):
```toml
[assets.DeepStrikeLanders.transport]
max_hex = 3
coin_cost = 2
cargo_types = ["Facility", "Military Unit", "Special Forces", "Tactic", "Logistics Facility"]
max_cargo = 1
```
**Replace with:**
```toml
[assets.DeepStrikeLanders.transport]
max_hex = 3
coin_cost = 2
cargo_types = ["Facility", "Military Unit", "Special Forces", "Tactic", "Logistics Facility"]
max_cargo = 1
ignores_permission = true
```

**Find** (`wealth_assets.toml`):
```toml
[assets.BlockadeRunners.transport]
max_hex = 3
coin_cost = 2
cargo_types = ["Military Unit", "Special Forces"]
max_cargo = 1
```
**Replace with:**
```toml
[assets.BlockadeRunners.transport]
max_hex = 3
coin_cost = 2
cargo_types = ["Military Unit", "Special Forces"]
max_cargo = 1
ignores_permission = true
```

##### Task 5 — tests

`internal/faction/engine/arrival_test.go` — append:
```go
func TestCheckArrivals_ABypassingCarrierAsksNothing(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("ship", "troops", 3)
	eng.Rulebook.Assets["ship"].Transport.IgnoresPermission = true
	collector := &minimalPhaseCollector{requestPermission: func(request hooks.PermissionRequest) (bool, error) {
		t.Errorf("asked %+v, want no ask past a bypassing carrier", request)
		return false, nil
	}}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	if !reflect.DeepEqual(got, []domain.Mutation{completed}) {
		t.Errorf("mutations = %+v, want the completion unchanged", got)
	}
}
```

`internal/faction/rulebook/rulebook_test.go` — `testDataDir` is the real
`rulebooks/swn`. Append:
```go
func TestLoad_TransportBypasses(t *testing.T) {
	rb, err := Load(testDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	for _, id := range []string{"SWN-F7-001", "SWN-W5-003"} {
		def := rb.Assets[id]
		if def == nil || def.Transport == nil || !def.Transport.IgnoresPermission {
			t.Errorf("%s transport = %+v, want ignores_permission", id, def)
		}
	}
	if rb.Assets["SWN-W2-001"].Transport.IgnoresPermission {
		t.Error("Freighter Contract ignores permission, want only the two bypass transports")
	}
}
```
##### Commit message

```
feat: permission bypasses

- TransportProfile.IgnoresPermission (ignores_permission): the carrier and
  its cargo land without asking (D-25)
- Deep Strike Landers and Blockade Runners carry it
```

### Phase 2 — Relocate permission

**In plain words:** a relocate (an asset's ability, or a free move such as Transit
Web's) now asks the destination's government for each restricted asset it moves; a
refused asset stays where it is, and the others still move.

#### Commit 3 — `feat: relocate permission`

##### Task 1 — `internal/faction/engine/relocate/relocate.go`

Imports gain `hooks`, `dispatch` and `state`.

**Find:**
```go
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
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Collector is the prompts a relocation asks: the orders, and the destination
// government's permission for a P asset. The action collector and the phase
// collector both satisfy it, so the action-spent harness and the free harness
// share the prompts as well as the logic.
type Collector interface {
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error)
	hooks.PermissionCollector
}
```

**Find:**
```go
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
```
**Replace with:**
```go
// Request is one relocation. Acting is the asset range is measured from: the
// ability's own asset, or the asset that granted a free move. Registry and
// FactionState serve the permission a P asset asks to land (D-24).
type Request struct {
	Faction      *domain.Faction
	Acting       *domain.Asset
	Params       *domain.RelocateParams
	World        *world.WorldEngine
	Rulebook     *rulebook.Rulebook
	Registry     *hooks.Registry
	FactionState *state.FactionState
	Collector    Collector
}

// Run resolves a relocation: gather the movable candidates (the acting asset for
// self-scope; range- and type-filtered friendly assets for other-scope), prompt
// for the moved assets and their destinations, ask the destination's government
// for each P asset, and emit the relocate mutations via the world engine. Range is
// measured from the acting asset. A refused order stays where it is and leaves its
// refusal record; the other orders still move (Effort 5 Decision 8).
```

**Find:**
```go
	var mutations []domain.Mutation
	for i, order := range orders {
		orderCost := params.CoinCost
		if params.CoinCostMode == domain.RelocateCostFlat && i > 0 {
			orderCost = 0
		}
		loc, ok := request.World.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, candidateWorlds, orderCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
	}
	return mutations, nil
}
```
**Replace with:**
```go
	var mutations []domain.Mutation
	moved := 0
	for _, order := range orders {
		loc, ok := request.World.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		refusal, err := askLanding(request, order)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		if refusal != nil {
			mutations = append(mutations, *refusal)
			continue
		}
		orderCost := params.CoinCost
		// deliberate: the first order that moves carries a flat cost, not the first picked, because a refused order moves nothing and charges nothing.
		if params.CoinCostMode == domain.RelocateCostFlat && moved > 0 {
			orderCost = 0
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, candidateWorlds, orderCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
		moved++
	}
	return mutations, nil
}

// askLanding asks the destination's government whether order's asset may land
// there: a relocate asks at its one step (D-24). It returns the refusal, or nil.
func askLanding(request Request, order action.RelocateOrder) (*domain.PermissionRefused, error) {
	def := request.Rulebook.Assets[order.Asset.DefinitionID]
	if def == nil {
		return nil, nil
	}
	return dispatch.RequestArrivalPermission(request.Registry, request.Collector, request.FactionState, request.Faction, order.Asset, def, order.DestinationWorld, domain.PermissionCauseRelocate)
}
```

##### Task 2 — `internal/faction/engine/action/actions/ability/relocate.go`

**Find:**
```go
	return relocate.Run(relocate.Request{
		Faction:   ctx.Faction,
		Acting:    ctx.Asset,
		Params:    ctx.Step.Relocate,
		World:     ctx.WorldEngine,
		Rulebook:  ctx.Rulebook,
		Collector: ctx.Collector,
	})
```
**Replace with:**
```go
	return relocate.Run(relocate.Request{
		Faction:      ctx.Faction,
		Acting:       ctx.Asset,
		Params:       ctx.Step.Relocate,
		World:        ctx.WorldEngine,
		Rulebook:     ctx.Rulebook,
		Registry:     ctx.Registry,
		FactionState: ctx.FactionState,
		Collector:    ctx.Collector,
	})
```

##### Task 3 — `internal/faction/engine/orchestrator.go`

In `runGrantedMovement`:

**Find:**
```go
		mutations, err := relocate.Run(relocate.Request{
			Faction:   faction,
			Acting:    grant.Source,
			Params:    &params,
			World:     e.World,
			Rulebook:  e.Rulebook,
			Collector: collector,
		})
```
**Replace with:**
```go
		mutations, err := relocate.Run(relocate.Request{
			Faction:      faction,
			Acting:       grant.Source,
			Params:       &params,
			World:        e.World,
			Rulebook:     e.Rulebook,
			Registry:     e.Hooks,
			FactionState: factionState,
			Collector:    collector,
		})
```

##### Task 4 — `internal/faction/engine/relocate/relocate_test.go`

Imports gain `hooks` and `state`:

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
```

`relocRulebook` gains a `P` Military Unit that Transit Web may move:

**Find:**
```go
		"WS": {ID: "WS", Category: domain.StatWealth, Type: domain.TypeStarship},
	}}
}
```
**Replace with:**
```go
		"WS": {ID: "WS", Category: domain.StatWealth, Type: domain.TypeStarship},
		"MP": {ID: "MP", Name: "Mercenaries", Category: domain.StatWealth, Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
}
```

Append:
```go
// landingFixture is Transit Web on origin with mercenaries (P) and a facility on
// near; f2 governs mid.
func landingFixture() (*domain.Faction, *domain.Faction, *domain.Asset, *domain.Asset, *domain.Asset, *state.FactionState) {
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	mercs := &domain.Asset{ID: "mp-1", DefinitionID: "MP", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	works := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"web": web, "mp-1": mercs, "wf-1": works}}
	ruler := &domain.Faction{ID: "f2"}
	factionState := &state.FactionState{
		Factions:    map[string]*domain.Faction{"f1": faction, "f2": ruler},
		Governments: map[string]string{"mid": "f2"},
	}
	return faction, ruler, web, mercs, works, factionState
}

func TestRun_RefusedOrderStaysAndLeavesARecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	faction, ruler, web, mercs, works, factionState := landingFixture()
	rb := relocRulebook()
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: mercs, DestinationWorld: "mid"}, {Asset: works, DestinationWorld: "mid"}}, nil)
	collector.EXPECT().RequestPermission(hooks.PermissionRequest{Requester: faction, Ruler: ruler, WorldID: "mid", Def: rb.Assets["MP"], Cause: "relocate"}).
		Return(false, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: rb, FactionState: factionState,
	})

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 3 {
		t.Fatalf("mutations = %+v, want the refusal, then the facility's move and the flat cost", mutations)
	}
	wantRefusal := domain.PermissionRefused{FactionID: "f1", WorldID: "mid", DefinitionID: "MP", DecidedByFactionID: "f2", Cause: "relocate", CausedByFactionID: "f1"}
	if refusal, ok := mutations[0].(domain.PermissionRefused); !ok || refusal != wantRefusal {
		t.Errorf("mutations[0] = %+v, want %+v", mutations[0], wantRefusal)
	}
	if moved, ok := mutations[1].(domain.AssetMoved); !ok || moved.AssetID != "wf-1" {
		t.Errorf("mutations[1] = %+v, want wf-1 moved", mutations[1])
	}
	if coin, ok := mutations[2].(domain.CoinDelta); !ok || coin.Delta != -1 {
		t.Errorf("mutations[2] = %+v, want the flat cost on the first order that moves", mutations[2])
	}
}

func TestRun_EveryOrderRefusedChargesNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	faction, _, web, mercs, _, factionState := landingFixture()
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: mercs, DestinationWorld: "mid"}}, nil)
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(), FactionState: factionState,
	})

	if err != nil {
		t.Fatalf("Run: %v, want the refusal as the relocation's only result", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want only the refusal", mutations)
	}
	if _, ok := mutations[0].(domain.PermissionRefused); !ok {
		t.Errorf("mutations[0] = %T, want PermissionRefused", mutations[0])
	}
}

func TestRun_StealthedPAssetIsNotAsked(t *testing.T) {
	ctrl := gomock.NewController(t)
	faction, _, web, mercs, _, factionState := landingFixture()
	mercs.Stealthy = true
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: mercs, DestinationWorld: "mid"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(), FactionState: factionState,
	})

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if moved, ok := mutations[0].(domain.AssetMoved); !ok || moved.AssetID != "mp-1" {
		t.Errorf("mutations[0] = %+v, want the stealthed mercenaries moved without an ask (D-25)", mutations[0])
	}
}
```

##### Task 5 — `internal/faction/engine/orchestrator_test.go`

Append:
```go
func TestRunGrantedMovement_RefusedLandingStays(t *testing.T) {
	eng, factionState, alpha, paths := grantFixture(t, 5)
	eng.Rulebook = &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"troops": {ID: "troops", Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
	alpha.Assets["wf"].DefinitionID = "troops"
	collector := &minimalPhaseCollector{
		selectRelocateOrders: func(_ *domain.Asset, candidates []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
			return []action.RelocateOrder{{Asset: candidates[0], DestinationWorld: "Tartarus"}}, nil
		},
		requestPermission: func(request hooks.PermissionRequest) (bool, error) {
			if request.Cause != domain.PermissionCauseRelocate {
				t.Errorf("Cause = %q, want relocate", request.Cause)
			}
			return false, nil
		},
	}

	resolved, err := eng.runGrantedMovement(alpha, factionState, paths, collector, &errorObserver{}, slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("runGrantedMovement: %v", err)
	}
	if got := alpha.Assets["wf"].Location.WorldID; got != "Krylos" {
		t.Errorf("wf on %q, want it still on Krylos", got)
	}
	if alpha.Coin != 5 {
		t.Errorf("Coin = %d, want 5 (a refused landing charges nothing)", alpha.Coin)
	}
	if len(resolved) != 1 {
		t.Fatalf("resolved = %+v, want only the refusal", resolved)
	}
	if _, ok := resolved[0].(domain.PermissionRefused); !ok {
		t.Errorf("resolved[0] = %T, want PermissionRefused", resolved[0])
	}
}
```

##### Task 6 — docs

`docs/architecture/engine/world-movement.md` — the relocate paragraph:

**Find:**
```markdown
prompts through a one-method `OrderCollector`, checks Coin, and calls
`BuildRelocateMutations` per order. Under the `flat` cost mode only the first order
carries the cost.
```
**Replace with:**
```markdown
prompts through `relocate.Collector`, checks Coin, asks the destination's
government for each `P` asset it moves (`dispatch.RequestArrivalPermission`, cause
`relocate`), and calls `BuildRelocateMutations` for each order that may land. A
refused order stays where it is and leaves a `PermissionRefused`; the others move.
Under the `flat` cost mode only the first order that moves carries the cost.
```

The harness Key Decision:

**Find:**
```markdown
- **One relocate orchestration serves both harnesses, behind a one-method
  collector.** `engine/relocate` holds the single copy of enumerate → filter →
  prompt → build. The action-spent harness (`ability`) and the free harness (the
  movement phase) both call `relocate.Run`, and both of their collectors satisfy
  `relocate.OrderCollector`, so the prompt is shared too, not just the pure parts.
```
**Replace with:**
```markdown
- **One relocate orchestration serves both harnesses, behind one collector.**
  `engine/relocate` holds the single copy of enumerate → filter → prompt → ask →
  build. The action-spent harness (`ability`) and the free harness (the movement
  phase) both call `relocate.Run`, and both of their collectors satisfy
  `relocate.Collector` (the orders and the permission request), so the prompts are
  shared too, not just the pure parts.
- **A relocate asks for each order after the pick, and a refusal drops that order
  only.** The prompt offers one world list for every candidate, so a per-asset
  permission cannot filter it; asking after the pick keeps the collector unchanged.
  The other orders still move, and a refused order charges nothing. (A.001.6
  Effort 5 Decision 8.)
```

`docs/architecture/engine/orchestrator.md`:

**Find:**
```markdown
shares the signature of the action collector's, so both satisfy
`relocate.OrderCollector`).
```
**Replace with:**
```markdown
shares the signature of the action collector's, so both satisfy
`relocate.Collector`).
```

`docs/rules/swn-faction-mechanics.md` — append to the paragraph Commit 1 added:

**Find:**
```markdown
A faction never asks on a world it governs, but it does ask to land on its own homeworld when another government holds it.
```
**Replace with:**
```markdown
A faction never asks on a world it governs, but it does ask to land on its own homeworld when another government holds it. A refused relocate leaves that asset where it is; the other assets of the same relocate still move.
```

##### Commit message

```
feat: relocate permission

- relocate.Run asks the destination's government for each P asset after
  the pick; a refused order stays and leaves a PermissionRefused, the others
  move, and a flat cost falls on the first order that moves (Decision 8)
- relocate.Collector replaces OrderCollector: the orders and the permission
  request; Request carries the registry and the faction state
- Both harnesses (Use Asset Ability, free movement grants) pass them
```

### Phase 3 — Lobbyists

**In plain words:** when a rival is granted a government's permission on a world where
a faction holds Lobbyists, that faction may test Cunning against the rival; a win takes
the permission back, and the rival cannot ask again for that world until its next turn.

#### Commit 4 — `feat: Lobbyists`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialGrantPermission SpecialEffectType = "grant_permission"
)
```
**Replace with:**
```go
	SpecialGrantPermission SpecialEffectType = "grant_permission"
	SpecialWithdrawPermission SpecialEffectType = "withdraw_permission"
)
```

**Find:**
```go
	TriggerPermissionQuery        EffectTrigger = "permission_query"
)
```
**Replace with:**
```go
	TriggerPermissionQuery        EffectTrigger = "permission_query"
	TriggerPermissionGranted      EffectTrigger = "permission_granted"
)
```

##### Task 2 — `internal/faction/domain/mutation.go`

**Find:**
```go
// PermissionRefused records a refused permission request. It changes no state:
// the purchase or landing it refuses did not happen (D-35, overview Decision 22).
// DecidedByFactionID is the world's ruler, empty when the GM decided for an
// independent world. Cause is the path: buy, refit or ability for a purchase,
// arrival or relocate for a landing.
type PermissionRefused struct {
	FactionID          string `json:"faction_id"`
	WorldID            string `json:"world_id"`
	DefinitionID       string `json:"definition_id"`
	DecidedByFactionID string `json:"decided_by_faction_id"`
	Cause              string `json:"cause"`
	CausedByFactionID  string `json:"caused_by_faction_id"`
}
```
**Replace with:**
```go
// RefusalReason says how a permission request was refused.
type RefusalReason string

const (
	// RefusalDenied (the zero value): the government said no.
	RefusalDenied RefusalReason = ""
	// RefusalWithdrawn: a rival's Lobbyists won its contest after the grant.
	RefusalWithdrawn RefusalReason = "withdrawn"
	// RefusalBarred: a withdrawal on this world earlier in the turn bars the request (D-30).
	RefusalBarred RefusalReason = "barred"
)

// PermissionRefused records a refused permission request. It changes no state:
// the purchase or landing it refuses did not happen (D-35, overview Decision 22).
// DecidedByFactionID is the world's ruler (empty when the GM decided for an
// independent world), or the Lobbyists' owner on a withdrawal. Cause is the path:
// buy, refit or ability for a purchase, arrival or relocate for a landing.
type PermissionRefused struct {
	FactionID          string        `json:"faction_id"`
	WorldID            string        `json:"world_id"`
	DefinitionID       string        `json:"definition_id"`
	DecidedByFactionID string        `json:"decided_by_faction_id"`
	Reason             RefusalReason `json:"reason,omitempty"`
	Cause              string        `json:"cause"`
	CausedByFactionID  string        `json:"caused_by_faction_id"`
}
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find:**
```go
		domain.TriggerPermissionQuery,
	}
```
**Replace with:**
```go
		domain.TriggerPermissionQuery, domain.TriggerPermissionGranted,
	}
```

**Find:**
```go
		domain.SpecialGrantStealth, domain.SpecialGrantPermission,
	}
```
**Replace with:**
```go
		domain.SpecialGrantStealth, domain.SpecialGrantPermission, domain.SpecialWithdrawPermission,
	}
```

##### Task 4 — `internal/faction/engine/hooks/rule_modifier.go`, `registry.go`, `doc.go`

`rule_modifier.go` — append:
```go
// PermissionContest is one asset's offer to contest a permission granted to a
// rival (Lobbyists).
type PermissionContest struct {
	Source *domain.Asset
}

// PermissionContester contests a permission granted to a rival (Cat 4).
// OfferContest returns nil when the grant is not its to contest. It never
// prompts: the dispatcher asks the owner, and on a yes calls WinsContest, which
// rolls the hook's own test against the requester.
type PermissionContester interface {
	OfferContest(query PermissionQuery) *PermissionContest
	WinsContest(query PermissionQuery) bool
}
```

`registry.go` — the field, the reset and the clear:

**Find:**
```go
	permissionGranters       map[Scope][]RegisteredPermissionGranter
}
```
**Replace with:**
```go
	permissionGranters       map[Scope][]RegisteredPermissionGranter
	permissionContesters     map[Scope][]RegisteredPermissionContester
}
```

**Find:**
```go
	registry.permissionGranters = make(map[Scope][]RegisteredPermissionGranter)
}
```
**Replace with:**
```go
	registry.permissionGranters = make(map[Scope][]RegisteredPermissionGranter)
	registry.permissionContesters = make(map[Scope][]RegisteredPermissionContester)
}
```

**Find:**
```go
	delete(registry.permissionGranters, scope)
}
```
**Replace with:**
```go
	delete(registry.permissionGranters, scope)
	delete(registry.permissionContesters, scope)
}
```

Append after `PermissionGrantersForAsset`:

**Find:**
```go
func (registry *Registry) PermissionGrantersForAsset(factionID, assetInstanceID string) []RegisteredPermissionGranter {
	return registry.permissionGranters[AssetScope(factionID, assetInstanceID)]
}
```
**Replace with:**
```go
func (registry *Registry) PermissionGrantersForAsset(factionID, assetInstanceID string) []RegisteredPermissionGranter {
	return registry.permissionGranters[AssetScope(factionID, assetInstanceID)]
}

// RegisteredPermissionContester is a Cat 4 permission-contester hook with its scope and source.
type RegisteredPermissionContester struct {
	Scope  Scope
	Source string
	Hook   PermissionContester
}

// RegisterPermissionContester registers a Cat 4 permission-contester hook under the given scope.
func (registry *Registry) RegisterPermissionContester(scope Scope, source string, hook PermissionContester) {
	registry.permissionContesters[scope] = append(registry.permissionContesters[scope], RegisteredPermissionContester{Scope: scope, Source: source, Hook: hook})
}

// PermissionContestersForAsset returns the Cat 4 permission-contester hooks
// registered against one asset instance. Every contester is an asset's
// (Lobbyists), so there is no global or faction lookup.
func (registry *Registry) PermissionContestersForAsset(factionID, assetInstanceID string) []RegisteredPermissionContester {
	return registry.permissionContesters[AssetScope(factionID, assetInstanceID)]
}
```

`doc.go`:

**Find:**
```go
// RepairCostModifier, PermissionGranter) rather
```
**Replace with:**
```go
// RepairCostModifier, PermissionGranter, PermissionContester) rather
```

##### Task 5 — `internal/faction/engine/hooks/collector.go`

**Find:**
```go
// PermissionCollector asks for a government's permission. action.Collector and
// the engine's PhaseCollector both embed it (D-41): purchases and relocates ask in
// an action, drift arrivals in the movement phase.
type PermissionCollector interface {
	RequestPermission(request PermissionRequest) (bool, error)
}
```
**Replace with:**
```go
// PermissionContestOffer asks Owner whether its Source (Lobbyists) contests the
// permission Requester was just granted to raise or land Def on WorldID. Cause is
// the request's path, as on PermissionRequest.
type PermissionContestOffer struct {
	Owner     *domain.Faction
	Source    *domain.Asset
	Requester *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
	Cause     string
}

// PermissionCollector asks for a government's permission, and asks a Lobbyists'
// owner whether to contest a grant. action.Collector and the engine's
// PhaseCollector both embed it (D-41): purchases and relocates ask in an action,
// drift arrivals in the movement phase.
type PermissionCollector interface {
	RequestPermission(request PermissionRequest) (bool, error)
	ConfirmPermissionContest(offer PermissionContestOffer) (bool, error)
}
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/permission.go`

Rewritten — full contents:
```go
package dispatch

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// PermissionWithdrawnKey is the HookBudgets mark a Lobbyists withdrawal spends on
// the requester. Until its next bookkeeping clears it, every request it makes for
// worldID is refused with no ask (D-30, Effort 5 Decision 9).
func PermissionWithdrawnKey(worldID string) string {
	return "permission_withdrawn:" + worldID
}

// RequestPermission asks worldID's government whether requester may raise def
// there, and returns the refusal record, or nil when the purchase may go ahead. A
// definition that needs no permission is not asked about, nor is a purchase on the
// requester's homeworld (Effort 4 Decision 15). cause names the purchase path on
// the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// RequestArrivalPermission asks worldID's government whether requester may land
// asset there, at the end of a drift order or by a relocate (cause), and returns
// the refusal record, or nil when the asset may land. An asset that needs no
// permission is not asked about (AssetDefinition.NeedsArrivalPermission).
func RequestArrivalPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsArrivalPermission(asset) {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// askGovernment refuses a request a withdrawal has barred this turn. It grants a
// ruler's own request at once (D-6), and one a grant hook of the requester serves
// (Popular Movement); otherwise it asks the ruler's driver, or the GM for an
// independent world (D-4). Every grant is then open to the rival Lobbyists on the
// world (D-29).
func askGovernment(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if budgetHolder(factionState, requester).BudgetSpent(PermissionWithdrawnKey(worldID)) {
		return refusal(requester, def, worldID, cause, "", domain.RefusalBarred), nil
	}
	rulerID := factionState.Ruler(worldID)
	granted := rulerID == requester.ID || PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def})
	if !granted {
		request := hooks.PermissionRequest{Requester: requester, WorldID: worldID, Def: def, Cause: cause}
		if rulerID != "" {
			request.Ruler = factionState.Factions[rulerID]
		}
		var err error
		if granted, err = collector.RequestPermission(request); err != nil {
			return nil, fmt.Errorf("permission request: %w", err)
		}
		if !granted {
			decidedBy := ""
			if request.Ruler != nil {
				decidedBy = request.Ruler.ID
			}
			return refusal(requester, def, worldID, cause, decidedBy, domain.RefusalDenied), nil
		}
	}
	return contestGrant(registry, collector, factionState, requester, def, worldID, cause)
}

// contestGrant offers a grant to each rival's contesters on worldID (Lobbyists):
// rival factions in ID order, then their assets in ID order (Effort 5 Decision 11).
// Each offer asks its owner whether to contest; on a yes the hook rolls, and the
// first win withdraws the grant and bars the requester from that world until its
// next turn. It returns the withdrawal, or nil when the grant stands.
func contestGrant(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if registry == nil || factionState == nil {
		return nil, nil
	}
	query := hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def}
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if factionID == requester.ID {
			continue
		}
		rival := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(rival) {
			for _, registered := range registry.PermissionContestersForAsset(rival.ID, asset.ID) {
				offer := registered.Hook.OfferContest(query)
				if offer == nil {
					continue
				}
				contest, err := collector.ConfirmPermissionContest(hooks.PermissionContestOffer{
					Owner: rival, Source: offer.Source, Requester: requester, WorldID: worldID, Def: def, Cause: cause,
				})
				if err != nil {
					return nil, fmt.Errorf("permission contest: %w", err)
				}
				if !contest || !registered.Hook.WinsContest(query) {
					continue
				}
				// deliberate: mark the requester the state holds, not the real faction, because in Use Asset Ability that is the action's view, and nothing after the action phase asks.
				budgetHolder(factionState, requester).SpendBudget(PermissionWithdrawnKey(worldID))
				return refusal(requester, def, worldID, cause, rival.ID, domain.RefusalWithdrawn), nil
			}
		}
	}
	return nil, nil
}

// budgetHolder is the faction whose HookBudgets carry requester's mark: its entry
// in factionState, or requester itself when the state does not hold it.
func budgetHolder(factionState *state.FactionState, requester *domain.Faction) *domain.Faction {
	if factionState != nil {
		if holder := factionState.Factions[requester.ID]; holder != nil {
			return holder
		}
	}
	return requester
}

func refusal(requester *domain.Faction, def *domain.AssetDefinition, worldID, cause, decidedBy string, reason domain.RefusalReason) *domain.PermissionRefused {
	return &domain.PermissionRefused{
		FactionID:          requester.ID,
		WorldID:            worldID,
		DefinitionID:       def.ID,
		DecidedByFactionID: decidedBy,
		Reason:             reason,
		Cause:              cause,
		CausedByFactionID:  requester.ID,
	}
}
```

##### Task 7 — `internal/faction/engine/effect/effects/withdraw_permission.go` (new), `registrars.go`

`withdraw_permission.go` — full contents:
```go
package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// withdrawPermissionRegistrar is Lobbyists: when a rival is granted a
// government's permission on the declaring asset's world, its owner may test the
// rival, and a win withdraws the permission. owner = "rival" reads as "the
// requester is a rival", scope = "world" as "the request is for the declaring
// asset's world". The contest gate carries the test (Effort 5 Decision 10).
var withdrawPermissionRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPermissionGranted: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerWithdrawPermission,
	Validate: validateWithdrawPermission,
}

func validateWithdrawPermission(entry domain.EffectDefinition) error {
	if entry.Gates.Contest == nil || entry.Gates.Contest.Roller != domain.ContestRollerSelf {
		return fmt.Errorf("needs a contest gate with roller %q", domain.ContestRollerSelf)
	}
	return nil
}

func registerWithdrawPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionContester(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialWithdrawPermission)+":"+ctx.Def.ID,
		&permissionContester{owner: ctx.Faction, assetID: ctx.Asset.ID, contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller its
// registration carried, as a reactor does.
type permissionContester struct {
	owner   *domain.Faction
	assetID string
	contest domain.EffectContest
	roller  domain.Roller
}

func (contester *permissionContester) OfferContest(query hooks.PermissionQuery) *hooks.PermissionContest {
	self := contester.owner.Assets[contester.assetID]
	if !active(self) || query.Requester.ID == contester.owner.ID {
		return nil
	}
	if self.Location.WorldID == "" || self.Location.WorldID != query.WorldID {
		return nil
	}
	return &hooks.PermissionContest{Source: self}
}

// WinsContest rolls the bare contest, as every effect contest does (Effort 3
// Decision 12): the request path carries no rulebook for roll hooks.
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery) bool {
	return contestPasses(contester.contest, contester.owner, query.Requester, contester.roller)
}
```

`registrars.go`:

**Find:**
```go
		domain.SpecialGrantPermission: grantPermissionRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialGrantPermission: grantPermissionRegistrar,
		domain.SpecialWithdrawPermission: withdrawPermissionRegistrar,
	}
```

##### Task 8 — `rulebooks/swn/assets/cunning_assets.toml`

**Find:**
```toml
description = "When a rival faction gains governmental permission to buy or transport an asset, the Lobbyists can make an immediate Cunning vs. Cunning test; if successful, the permission is withdrawn and cannot be re-attempted until next turn."

[assets.Saboteurs]
```
**Replace with:**
```toml
description = "When a rival faction gains governmental permission to buy or transport an asset, the Lobbyists can make an immediate Cunning vs. Cunning test; if successful, the permission is withdrawn and cannot be re-attempted until next turn."

[[assets.Lobbyists.effects]]
trigger = "permission_granted"
effect  = "withdraw_permission"
[assets.Lobbyists.effects.filters]
owner = "rival"
scope = "world"
[assets.Lobbyists.effects.gates.contest]
roller        = "self"
attacker_stat = "Cunning"
defender_stat = "Cunning"

[assets.Saboteurs]
```

##### Task 9 — `internal/faction/tui/adapter/channels.go`, `action_collector.go`, `phase_collector.go`, `smoke.go`

`channels.go`:

**Find:**
```go
	AskRequestPermission
	// AskKindCount counts the kinds above; keep it last.
```
**Replace with:**
```go
	AskRequestPermission
	AskConfirmPermissionContest
	// AskKindCount counts the kinds above; keep it last.
```

**Find:**
```go
type RequestPermissionPayload struct {
	Request   hooks.PermissionRequest
	WorldName string
}
```
**Replace with:**
```go
type RequestPermissionPayload struct {
	Request   hooks.PermissionRequest
	WorldName string
}
type ConfirmPermissionContestPayload struct {
	Offer     hooks.PermissionContestOffer
	WorldName string
}
```

`action_collector.go` — after `Adapter.requestPermission` (Commit 1 Task 14):

**Find:**
```go
	raw, err := a.ask(AskRequestPermission, request.Ruler, RequestPermissionPayload{Request: request, WorldName: worldName})
	if err != nil {
		return false, err
	}
	granted, _ := raw.(bool)
	return granted, nil
}
```
**Replace with:**
```go
	raw, err := a.ask(AskRequestPermission, request.Ruler, RequestPermissionPayload{Request: request, WorldName: worldName})
	if err != nil {
		return false, err
	}
	granted, _ := raw.(bool)
	return granted, nil
}

func (a *actionCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	return a.adapter.confirmPermissionContest(offer)
}

// confirmPermissionContest asks a Lobbyists' owner, who is never the acting
// faction, whether to contest a rival's grant. Both collectors share it.
func (a *Adapter) confirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	worldName := offer.WorldID
	if loc, ok := a.engine.World.Location(offer.WorldID); ok {
		worldName = loc.Name()
	}
	raw, err := a.ask(AskConfirmPermissionContest, offer.Owner, ConfirmPermissionContestPayload{Offer: offer, WorldName: worldName})
	if err != nil {
		return false, err
	}
	contest, _ := raw.(bool)
	return contest, nil
}
```

`phase_collector.go`:

**Find:**
```go
func (p *phaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	return p.adapter.requestPermission(request)
}
```
**Replace with:**
```go
func (p *phaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	return p.adapter.requestPermission(request)
}

func (p *phaseCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	return p.adapter.confirmPermissionContest(offer)
}
```

`smoke.go`:

**Find:**
```go
func (s *SmokePhaseCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, nil // the smoke faction has no assets, so nothing arrives to ask
}
```
**Replace with:**
```go
func (s *SmokePhaseCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, nil // the smoke faction has no assets, so nothing arrives to ask
}

func (s *SmokePhaseCollector) ConfirmPermissionContest(_ hooks.PermissionContestOffer) (bool, error) {
	return false, nil
}
```

**Find:**
```go
func (SmokeActionCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (SmokeActionCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (SmokeActionCollector) ConfirmPermissionContest(_ hooks.PermissionContestOffer) (bool, error) {
	return false, action.ErrActionUnavailable
}
```

##### Task 10 — `internal/faction/tui/views/turn/overlay/permission.go`, `execution/execution.go`

`overlay/permission.go` — append:
```go
// NewConfirmPermissionContest asks a Lobbyists' owner whether to contest a
// permission a rival was just granted. Esc does nothing, as with the request it
// follows. Answer: bool.
func NewConfirmPermissionContest(offer hooks.PermissionContestOffer, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of %s, for its asset %s", offer.Owner.Name, offer.Source.ID)
	return NoCancelConfirm{NewConfirm(contestTitle(offer, worldName), header, "Contest", "Let it stand")}
}

func contestTitle(offer hooks.PermissionContestOffer, worldName string) string {
	return fmt.Sprintf("Contest %s's permission to %s %s on %s?", offer.Requester.Name, permissionVerb(offer.Cause), offer.Def.Name, worldName)
}
```

`execution.go` — the `askEntries` row:

**Find:**
```go
	adapter.AskRequestPermission: {"Permission?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.RequestPermissionPayload)
		return overlay.NewRequestPermission(p.Request, p.WorldName)
	}},
```
**Replace with:**
```go
	adapter.AskRequestPermission: {"Permission?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.RequestPermissionPayload)
		return overlay.NewRequestPermission(p.Request, p.WorldName)
	}},
	adapter.AskConfirmPermissionContest: {"Contest?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.ConfirmPermissionContestPayload)
		return overlay.NewConfirmPermissionContest(p.Offer, p.WorldName)
	}},
```

##### Task 11 — `internal/faction/engine/testharness/collector.go`, mocks, engine and dispatch test collectors

`testharness/collector.go`:

**Find:**
```go
	RequestPermissionFn           func(hooks.PermissionRequest) (bool, error)
```
**Replace with:**
```go
	RequestPermissionFn           func(hooks.PermissionRequest) (bool, error)
	ConfirmPermissionContestFn    func(hooks.PermissionContestOffer) (bool, error)
```

**Find:**
```go
	if c.RequestPermissionFn != nil {
		return c.RequestPermissionFn(request)
	}
	return true, nil
}
```
**Replace with:**
```go
	if c.RequestPermissionFn != nil {
		return c.RequestPermissionFn(request)
	}
	return true, nil
}

// ConfirmPermissionContest lets every grant stand by default, so a scripted grant is
// never withdrawn unless a test scripts a contest.
func (c *ScriptedCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	if c.ConfirmPermissionContestFn != nil {
		return c.ConfirmPermissionContestFn(offer)
	}
	return false, nil
}
```

Mocks — run `go generate ./internal/faction/engine/action/...`.

`internal/faction/engine/orchestrator_test.go`:

**Find:**
```go
func (c *minimalPhaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	if c.requestPermission != nil {
		return c.requestPermission(request)
	}
	return true, nil
}
```
**Replace with:**
```go
func (c *minimalPhaseCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	if c.requestPermission != nil {
		return c.requestPermission(request)
	}
	return true, nil
}
func (c *minimalPhaseCollector) ConfirmPermissionContest(_ hooks.PermissionContestOffer) (bool, error) {
	return false, nil
}
```

`internal/faction/engine/hooks/dispatch/permission_test.go`:

**Find:**
```go
type stubPermissionCollector struct {
	grant bool
	asked []hooks.PermissionRequest
}

func (stub *stubPermissionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	stub.asked = append(stub.asked, request)
	return stub.grant, nil
}
```
**Replace with:**
```go
type stubPermissionCollector struct {
	grant   bool
	contest bool
	asked   []hooks.PermissionRequest
	offered []hooks.PermissionContestOffer
}

func (stub *stubPermissionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	stub.asked = append(stub.asked, request)
	return stub.grant, nil
}

func (stub *stubPermissionCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	stub.offered = append(stub.offered, offer)
	return stub.contest, nil
}
```

##### Task 12 — `internal/faction/narrative/digest/types.go`, `build.go`

`types.go`:

**Find:**
```go
// RefusalEvent is a purchase or a landing a government refused. DecidedBy is the
// ruler's name, empty when the GM decided for an independent world. Landing is
// true for an arrival or a relocate.
type RefusalEvent struct {
	AssetName string
	World     string
	DecidedBy string
	Landing   bool
}
```
**Replace with:**
```go
// RefusalEvent is a purchase or a landing that was refused. DecidedBy is the
// ruler's name (empty when the GM decided for an independent world), or the
// Lobbyists' owner's on a withdrawal. Landing is true for an arrival or a relocate.
type RefusalEvent struct {
	AssetName string
	World     string
	DecidedBy string
	Landing   bool
	Reason    domain.RefusalReason
}
```

`build.go`:

**Find:**
```go
					DecidedBy: decidedBy,
					Landing:   domain.IsLanding(m.Cause),
				})
```
**Replace with:**
```go
					DecidedBy: decidedBy,
					Landing:   domain.IsLanding(m.Cause),
					Reason:    m.Reason,
				})
```

##### Task 13 — `internal/faction/narrative/wire_renderer.go`, `wire_templates.go`, `tui/views/turn/execution/format.go`

`wire_renderer.go`:

**Find:**
```go
	tmpl := pick(rng, refusalTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{verb}", verb)
```
**Replace with:**
```go
	pool := refusalTemplates
	switch refusal.Reason {
	case domain.RefusalWithdrawn:
		pool = withdrawalTemplates
	case domain.RefusalBarred:
		pool = barredTemplates
	}
	tmpl := pick(rng, pool)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{verb}", verb)
```

`wire_templates.go`:

**Find:**
```go
var holdTemplates = map[domain.HoldReason][]string{
```
**Replace with:**
```go
var withdrawalTemplates = []string{
	"{decider}'s lobbyists had {faction}'s permission to {verb} {asset} on {world} withdrawn.",
	"Lobbying by {decider} cost {faction} its leave to {verb} {asset} on {world}.",
}

var barredTemplates = []string{
	"{faction} could not ask again to {verb} {asset} on {world} this turn.",
}

var holdTemplates = map[domain.HoldReason][]string{
```

`format.go`:

**Find:**
```go
	case domain.PermissionRefused:
		return fmt.Sprintf("permission refused: %s on %s", mut.DefinitionID, mut.WorldID), true
```
**Replace with:**
```go
	case domain.PermissionRefused:
		if mut.Reason != domain.RefusalDenied {
			return fmt.Sprintf("permission %s: %s on %s", mut.Reason, mut.DefinitionID, mut.WorldID), true
		}
		return fmt.Sprintf("permission refused: %s on %s", mut.DefinitionID, mut.WorldID), true
```

##### Task 14 — dispatch and effects tests

`internal/faction/engine/hooks/dispatch/permission_test.go` — append:
```go
type stubContester struct {
	source *domain.Asset
	wins   bool
	rolled int
}

func (stub *stubContester) OfferContest(hooks.PermissionQuery) *hooks.PermissionContest {
	if stub.source == nil {
		return nil
	}
	return &hooks.PermissionContest{Source: stub.source}
}

func (stub *stubContester) WinsContest(hooks.PermissionQuery) bool {
	stub.rolled++
	return stub.wins
}

// lobbyistsOf registers a contester on a new asset of faction and returns it.
func lobbyistsOf(registry *hooks.Registry, faction *domain.Faction, assetID string, wins bool) *stubContester {
	if faction.Assets == nil {
		faction.Assets = map[string]*domain.Asset{}
	}
	asset := &domain.Asset{ID: assetID, OwnerID: faction.ID}
	faction.Assets[assetID] = asset
	contester := &stubContester{source: asset, wins: wins}
	registry.RegisterPermissionContester(hooks.AssetScope(faction.ID, assetID), assetID, contester)
	return contester
}

func TestRequestPermission_LobbyistsContestAGrant(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	cases := []struct {
		name        string
		contest     bool
		wins        bool
		wantRolled  int
		wantRefusal *domain.PermissionRefused
	}{
		{name: "the owner lets it stand"},
		{name: "the owner contests and loses", contest: true, wantRolled: 1},
		{
			name: "the owner contests and wins", contest: true, wins: true, wantRolled: 1,
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Tartarus", DefinitionID: "gunships", DecidedByFactionID: "f2", Reason: domain.RefusalWithdrawn, Cause: "buy", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, requester, ruler := permissionState()
			registry := hooks.NewRegistry()
			lobbyists := lobbyistsOf(registry, ruler, "lob", tc.wins)
			collector := &stubPermissionCollector{contest: tc.contest}

			// Tartarus is the requester's own world: the grant is automatic (D-6), and still contested (D-29).
			refusal, err := RequestPermission(registry, collector, factionState, requester, permissionDef, "Tartarus", "buy")

			if err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			wantOffer := []hooks.PermissionContestOffer{{Owner: ruler, Source: lobbyists.source, Requester: requester, WorldID: "Tartarus", Def: permissionDef, Cause: "buy"}}
			if !reflect.DeepEqual(collector.offered, wantOffer) {
				t.Errorf("offered = %+v, want %+v", collector.offered, wantOffer)
			}
			if lobbyists.rolled != tc.wantRolled {
				t.Errorf("rolled %d times, want %d", lobbyists.rolled, tc.wantRolled)
			}
			if !reflect.DeepEqual(refusal, tc.wantRefusal) {
				t.Errorf("refusal = %+v, want %+v", refusal, tc.wantRefusal)
			}
			if got := requester.BudgetSpent(PermissionWithdrawnKey("Tartarus")); got != (tc.wantRefusal != nil) {
				t.Errorf("withdrawal mark spent = %v, want %v", got, tc.wantRefusal != nil)
			}
		})
	}
}

func TestRequestPermission_ContestWalkStopsAtTheFirstWin(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	factionState, requester, ruler := permissionState()
	registry := hooks.NewRegistry()
	own := lobbyistsOf(registry, requester, "own", true)
	second := lobbyistsOf(registry, ruler, "lob-b", true)
	first := lobbyistsOf(registry, ruler, "lob-a", false)
	collector := &stubPermissionCollector{contest: true}

	refusal, err := RequestPermission(registry, collector, factionState, requester, permissionDef, "Tartarus", "buy")

	if err != nil {
		t.Fatalf("RequestPermission: %v", err)
	}
	if refusal == nil || refusal.Reason != domain.RefusalWithdrawn {
		t.Fatalf("refusal = %+v, want a withdrawal", refusal)
	}
	var sources []string
	for _, offer := range collector.offered {
		sources = append(sources, offer.Source.ID)
	}
	if want := []string{"lob-a", "lob-b"}; !reflect.DeepEqual(sources, want) {
		t.Errorf("offered %v, want %v (asset ID order, the requester's own never)", sources, want)
	}
	if own.rolled != 0 || first.rolled != 1 || second.rolled != 1 {
		t.Errorf("rolled own=%d first=%d second=%d, want 0, 1, 1", own.rolled, first.rolled, second.rolled)
	}
}

func TestRequestPermission_NoContestWithoutAGrant(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	plainDef := &domain.AssetDefinition{ID: "militia"}
	cases := []struct {
		name  string
		def   *domain.AssetDefinition
		world string
	}{
		{"a purchase on the homeworld", permissionDef, "Anchorage"},
		{"no permission needed", plainDef, "Krylos"},
		{"the government refuses", permissionDef, "Krylos"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, requester, ruler := permissionState()
			registry := hooks.NewRegistry()
			lobbyistsOf(registry, ruler, "lob", true)
			collector := &stubPermissionCollector{contest: true}

			if _, err := RequestPermission(registry, collector, factionState, requester, tc.def, tc.world, "buy"); err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			if len(collector.offered) != 0 {
				t.Errorf("offered = %+v, want no contest", collector.offered)
			}
		})
	}
}

func TestRequestPermission_AWithdrawalBarsTheWorldThisTurn(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	factionState, requester, _ := permissionState()
	requester.SpendBudget(PermissionWithdrawnKey("Krylos"))
	collector := &stubPermissionCollector{grant: true}

	refusal, err := RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Krylos", domain.PermissionCauseArrival)

	if err != nil {
		t.Fatalf("RequestArrivalPermission: %v", err)
	}
	want := &domain.PermissionRefused{FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", Reason: domain.RefusalBarred, Cause: "arrival", CausedByFactionID: "f1"}
	if !reflect.DeepEqual(refusal, want) {
		t.Errorf("refusal = %+v, want %+v", refusal, want)
	}
	if len(collector.asked) != 0 {
		t.Errorf("asked = %+v, want no ask while barred", collector.asked)
	}
	if refusal, _ := RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Hadrian", domain.PermissionCauseArrival); refusal != nil || len(collector.asked) != 1 {
		t.Errorf("Hadrian: refusal = %+v, asked %d; want another world asked as usual", refusal, len(collector.asked))
	}
}
```

`internal/faction/engine/effect/effects/withdraw_permission_test.go` (new) — full
contents:
```go
package effects_test

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func lobbyistsEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger: domain.TriggerPermissionGranted,
		Effect:  domain.SpecialWithdrawPermission,
		Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld},
		Gates: domain.EffectGates{Contest: &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
		}},
	}
}

func registerLobbyists(t *testing.T, owner *domain.Faction, lobbyists *domain.Asset, roller domain.Roller) hooks.PermissionContester {
	t.Helper()
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialWithdrawPermission].Register(effect.RegisterContext{
		Faction: owner, Asset: lobbyists, Def: &domain.AssetDefinition{ID: "SWN-C2-001"}, Registry: registry, Roller: roller,
		Entry: lobbyistsEntry(),
	})
	contesters := registry.PermissionContestersForAsset(owner.ID, lobbyists.ID)
	if len(contesters) != 1 {
		t.Fatalf("contesters = %d, want 1", len(contesters))
	}
	return contesters[0].Hook
}

func lobbyistsFixture() (*domain.Faction, *domain.Faction, *domain.Asset) {
	lobbyists := &domain.Asset{ID: "lob", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 4}
	alpha := &domain.Faction{ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{"lob": lobbyists}}
	bravo := &domain.Faction{ID: "bravo", Cunning: 2}
	return alpha, bravo, lobbyists
}

func TestWithdrawPermission_OffersOnARivalsGrantOnItsWorld(t *testing.T) {
	alpha, bravo, lobbyists := lobbyistsFixture()
	contester := registerLobbyists(t, alpha, lobbyists, nil)
	def := &domain.AssetDefinition{ID: "gunships"}

	cases := []struct {
		name  string
		query hooks.PermissionQuery
		want  bool
	}{
		{"a rival on its world", hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}, true},
		{"its owner on its world", hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}, false},
		{"a rival elsewhere", hooks.PermissionQuery{Requester: bravo, WorldID: "Hadrian", Def: def}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			offer := contester.OfferContest(tc.query)
			if (offer != nil) != tc.want {
				t.Fatalf("OfferContest = %+v, want an offer: %v", offer, tc.want)
			}
			if offer != nil && offer.Source != lobbyists {
				t.Errorf("Source = %+v, want the Lobbyists", offer.Source)
			}
		})
	}

	lobbyists.Ready = false
	if contester.OfferContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}) != nil {
		t.Error("an inactive Lobbyists offered a contest, want none")
	}
}

func TestWithdrawPermission_WinsContestRollsCunning(t *testing.T) {
	cases := []struct {
		name  string
		rolls []int
		want  bool
	}{
		{"the owner rolls higher", []int{5, 5}, true},
		{"a tie goes to the requester", []int{3, 4}, false},
		{"the requester rolls higher", []int{1, 9}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alpha, bravo, lobbyists := lobbyistsFixture()
			contester := registerLobbyists(t, alpha, lobbyists, &seqRoller{values: tc.rolls})

			if got := contester.WinsContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos"}); got != tc.want {
				t.Errorf("WinsContest = %v, want %v (Cunning 3 vs 2)", got, tc.want)
			}
		})
	}
}

func TestWithdrawPermission_NeedsAContestGate(t *testing.T) {
	entry := lobbyistsEntry()
	entry.Gates = domain.EffectGates{}
	def := &domain.AssetDefinition{ID: "SWN-C2-001", Effects: []domain.EffectDefinition{entry}}

	if err := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)).RegisterEffectsList(def); err == nil {
		t.Error("RegisterEffectsList accepted withdraw_permission with no contest gate")
	}
}
```

`internal/faction/rulebook/rulebook_test.go` — append:
```go
func TestLoad_LobbyistsWithdrawPermission(t *testing.T) {
	rb, err := Load(testDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	want := []domain.EffectDefinition{{
		Trigger: domain.TriggerPermissionGranted,
		Effect:  domain.SpecialWithdrawPermission,
		Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld},
		Gates: domain.EffectGates{Contest: &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
		}},
	}}
	if got := rb.Assets["SWN-C2-001"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Lobbyists effects = %+v, want %+v", got, want)
	}
}
```

##### Task 15 — digest, renderer and TUI tests

`internal/faction/narrative/digest/digest_test.go` — append:
```go
func TestPermissionWithdrawnCarriesItsReason(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha", "f2": "Bravo"})
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"gunships": {ID: "gunships", Name: "Gunships"}}}
	records := []domain.EventRecord{eventRec("f1", mutRec(t, "permission_refused", domain.PermissionRefused{
		FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Reason: domain.RefusalWithdrawn, Cause: "buy", CausedByFactionID: "f1",
	}))}

	beat := findBeat(buildOne(t, records, factionState, rb), "f1")

	want := []RefusalEvent{{AssetName: "Gunships", World: "Krylos", DecidedBy: "Bravo", Reason: domain.RefusalWithdrawn}}
	if beat == nil || !reflect.DeepEqual(beat.Refusals, want) {
		t.Errorf("beat = %+v, want Refusals %+v", beat, want)
	}
}
```

`internal/faction/narrative/wire_renderer_test.go` — append:
```go
func TestRenderRefusal_WithdrawnAndBarred(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	withdrawn := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", DecidedBy: "Bravo", Landing: true, Reason: domain.RefusalWithdrawn}, "Alpha", rng)
	if !strings.Contains(withdrawn, "Bravo") || !strings.Contains(withdrawn, "land Troops") {
		t.Errorf("withdrawn line = %q, want the Lobbyists' owner and the landing", withdrawn)
	}
	barred := renderRefusal(digest.RefusalEvent{AssetName: "Troops", World: "Krylos", Reason: domain.RefusalBarred}, "Alpha", rng)
	if !strings.Contains(barred, "could not ask again") {
		t.Errorf("barred line = %q, want the barred wording", barred)
	}
}
```

`internal/faction/tui/views/turn/execution/format_test.go`:

**Find:**
```go
		{"PermissionRefused", domain.PermissionRefused{DefinitionID: "gunships", WorldID: "Krylos"}, "permission refused: gunships on Krylos"},
```
**Replace with:**
```go
		{"PermissionRefused", domain.PermissionRefused{DefinitionID: "gunships", WorldID: "Krylos"}, "permission refused: gunships on Krylos"},
		{"PermissionRefused withdrawn", domain.PermissionRefused{DefinitionID: "gunships", WorldID: "Krylos", Reason: domain.RefusalWithdrawn}, "permission withdrawn: gunships on Krylos"},
```

`internal/faction/tui/views/turn/overlay/asset_test.go` — append:
```go
func TestContestTitle(t *testing.T) {
	offer := hooks.PermissionContestOffer{
		Owner:     &domain.Faction{ID: "f2", Name: "Bravo"},
		Source:    &domain.Asset{ID: "lob"},
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "troops", Name: "Troops"},
		Cause:     domain.PermissionCauseRelocate,
	}
	if got, want := contestTitle(offer, "Krylos"), "Contest Alpha's permission to land Troops on Krylos?"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}
```

##### Task 16 — docs

`docs/architecture/engine/hooks.md` — the Cat 4 family list:

**Find:**
```markdown
`AssetMovementGranter`, `RepairCostModifier`, `PermissionGranter` — each with its own signature
```
**Replace with:**
```markdown
`AssetMovementGranter`, `RepairCostModifier`, `PermissionGranter`, `PermissionContester` — each with its own signature
```

The resolver list:

**Find:**
```markdown
  it is the definition field `purchase_permission` (A.001.6 Effort 4 Decision 1).
- **`InterceptLethal`** is Cat 6.
```
**Replace with:**
```markdown
  it is the definition field `purchase_permission` (A.001.6 Effort 4 Decision 1).
- **The contest walk** runs inside `RequestPermission` and
  `RequestArrivalPermission` after every grant: a ruler's own, a grant hook's, or
  a yes. It asks the `PermissionContester`s registered on each rival's assets,
  rival factions in ID order, then assets in ID order (Lobbyists). Each offer is
  asked of its owner through the `PermissionCollector`, and on a yes the hook rolls
  its own contest. **The first win withdraws the grant** and spends
  `permission_withdrawn:<world>` on the requester, which refuses its later requests
  for that world until its next bookkeeping. A homeworld purchase is no grant, so it
  is never contested. (A.001.6 Effort 5 Decisions 9–11.)
- **`InterceptLethal`** is Cat 6.
```

`docs/architecture/engine/ability-catalog.md` — the two Lobbyists rows:

**Find:**
```markdown
| **auto-permission / revoke-permission** | Cat 4 `PermissionGranter` → `grant_permission` (grant); revoke in Effort 5 | Popular Movement `SWN-C7-001` ✅ (grant), Lobbyists `SWN-C2-001` (revoke, reactive contest) |
| **reactive contest** (test on an enemy event) | Cat 3 + stat-contest | Lobbyists `SWN-C2-001`, Tripwire Cells `SWN-C4-003` ✅ |
```
**Replace with:**
```markdown
| **auto-permission / revoke-permission** | Cat 4 `PermissionGranter` → `grant_permission` (grant); Cat 4 `PermissionContester` → `withdraw_permission` (revoke) | Popular Movement `SWN-C7-001` ✅ (grant), Lobbyists `SWN-C2-001` ✅ (revoke, reactive contest) |
| **reactive contest** (test on an enemy event) | Cat 3 + stat-contest; Cat 4 `PermissionContester` + stat-contest | Lobbyists `SWN-C2-001` ✅, Tripwire Cells `SWN-C4-003` ✅ |
```

`docs/architecture/engine/effect-mutation.md`:

**Find:**
```markdown
- `permission_refused` — record only: a government refused a purchase (which bought nothing) or a landing (the asset holds or stays)
```
**Replace with:**
```markdown
- `permission_refused` — record only: a purchase (which bought nothing) or a landing (the asset holds or stays) was refused. Its reason says whether the government said no (the zero value), a rival's Lobbyists withdrew the grant, or a withdrawal earlier in the turn barred the request
```

`docs/rules/swn-faction-mechanics.md`:

**Find:**
```markdown
A refused relocate leaves that asset where it is; the other assets of the same relocate still move.
```
**Replace with:**
```markdown
A refused relocate leaves that asset where it is; the other assets of the same relocate still move.

Every grant — a ruler's own, Popular Movement's, or a government's yes — is open to a rival's **Lobbyists** on that world. Each Lobbyists' owner may test Cunning against the requester, in faction then asset order, and the first win withdraws the permission. The requester cannot ask again for that world until its next turn. A purchase on a faction's homeworld is no grant, so it cannot be contested.
```

##### Task 17 — `docs/dev_journals/faction-manager/planned-work.md`

File the effect-contest row in the Refactors table (Decision 10).

**Find:**
```markdown
| `R.022` | Orchestrator error-handling consolidation | -- | Collapse the 26 recoverable-or-fatal blocks in `orchestrator.go`; only collector calls can return the three recoverable sentinels; target <500 lines, no behaviour change; `D.002` N3 |
```
**Replace with:**
```markdown
| `R.022` | Orchestrator error-handling consolidation | -- | Collapse the 26 recoverable-or-fatal blocks in `orchestrator.go`; only collector calls can return the three recoverable sentinels; target <500 lines, no behaviour change; `D.002` N3 |
| `R.023` | Effect contests through roll hooks | When reactors and phase effects can reach a collector | Tripwire Cells, Panopticon Matrix and Lobbyists roll bare 1d10 + stat (`contestPasses`), so Fanatical's and Book of Secrets' rerolls skip them; move all three onto `RollWithHooks` together. `A.001.6` Effort 5 Decision 10. |
```

##### Commit message

```
feat: Lobbyists

- Cat 4 PermissionContester (OfferContest, WinsContest); Lobbyists'
  permission_granted -> withdraw_permission with a Cunning contest gate
- Every grant (D-6, a grant hook, a yes) is open to each active rival
  Lobbyists on the world, in ID order, until one wins (Decisions 10, 11)
- A withdrawal spends permission_withdrawn:<world> on the requester; a later
  request there this turn is refused with no ask (Decision 9)
- PermissionRefused.Reason: denied (zero value), withdrawn, barred
- ConfirmPermissionContest on hooks.PermissionCollector; the TUI's contest
  prompt; digest and TUI wording for withdrawn and barred refusals
- The contest rolls bare dice, as the other effect contests do; R.023 filed
  to move all three onto roll hooks together
```

### Phase 4 — Transport Lockdown

**In plain words:** a successful Transport Lockdown attack shuts the struck faction out
of the Lockdown's world: each of its drift orders arriving there pays 1d4 Coin and waits
a turn in orbit, and none of its relocates may land there, for as long as the Lockdown
stays active on that world. Deep Strike Landers get through.

#### Commit 5 — `feat: Transport Lockdown`

##### Task 1 — `internal/faction/domain/asset.go`

The Lockdown bypass on the transport profile:

**Find:**
```go
// TransportProfile is a transport's typed data. IgnoresPermission lets its order
// land the carrier and its cargo without asking the government (D-25).
type TransportProfile struct {
	MaxHex            int
	CoinCost          int
	CargoTypes        []AssetType
	MaxCargo          int
	ExcludeCategories []FactionStat
	IgnoresPermission bool
}
```
**Replace with:**
```go
// TransportProfile is a transport's typed data. IgnoresPermission lets its order
// land the carrier and its cargo without asking the government (D-25);
// IgnoresLockdown lets it land past a Transport Lockdown (Effort 5 Decision 15).
type TransportProfile struct {
	MaxHex            int
	CoinCost          int
	CargoTypes        []AssetType
	MaxCargo          int
	ExcludeCategories []FactionStat
	IgnoresPermission bool
	IgnoresLockdown   bool
}
```

The condition kind, its target faction, and its reader:

**Find:**
```go
	ConditionSheltered ConditionKind = "sheltered"
)

// Condition is a lasting restriction one asset puts on another. The code that
// reads a kind decides what it blocks. ExpiresAtTurnOf names the faction whose
// turn start clears it; empty means the kind ends some other way.
type Condition struct {
	Kind            ConditionKind `toml:"kind" json:"kind"`
	SourceAssetID   string        `toml:"source_asset_id" json:"source_asset_id"`
	ExpiresAtTurnOf string        `toml:"expires_at_turn_of,omitempty" json:"expires_at_turn_of,omitempty"`
}

// Is reports whether condition and other are one condition: the same kind from
// the same source.
func (condition Condition) Is(other Condition) bool {
	return condition.Kind == other.Kind && condition.SourceAssetID == other.SourceAssetID
}
```
**Replace with:**
```go
	ConditionSheltered ConditionKind = "sheltered"
	ConditionLockedOut ConditionKind = "locked_out"
)

// Condition is a lasting restriction one asset puts on another, or on itself
// against a faction: Transport Lockdown's locked_out names TargetFactionID
// (overview Decision 24). The code that reads a kind decides what it blocks.
// ExpiresAtTurnOf names the faction whose turn start clears it; empty means the
// kind ends some other way.
type Condition struct {
	Kind            ConditionKind `toml:"kind" json:"kind"`
	SourceAssetID   string        `toml:"source_asset_id" json:"source_asset_id"`
	ExpiresAtTurnOf string        `toml:"expires_at_turn_of,omitempty" json:"expires_at_turn_of,omitempty"`
	TargetFactionID string        `toml:"target_faction_id,omitempty" json:"target_faction_id,omitempty"`
}

// Is reports whether condition and other are one condition: the same kind from
// the same source, against the same faction.
func (condition Condition) Is(other Condition) bool {
	return condition.Kind == other.Kind && condition.SourceAssetID == other.SourceAssetID && condition.TargetFactionID == other.TargetFactionID
}

// LockedOut reports whether a Transport Lockdown holds factionID out of worldID:
// an active asset of another faction on the world carries locked_out, from itself,
// against factionID. The lock lives and dies with that asset (D-31, D-36); a
// capture into factionID lapses it.
func LockedOut(factions map[string]*Faction, worldID, factionID string) bool {
	if worldID == "" {
		return false
	}
	for _, holder := range factions {
		if holder.ID == factionID {
			continue
		}
		for _, asset := range holder.Assets {
			if asset.Location.WorldID != worldID || !asset.Active() {
				continue
			}
			if slices.ContainsFunc(asset.Conditions, func(condition Condition) bool {
				return condition.Kind == ConditionLockedOut && condition.SourceAssetID == asset.ID && condition.TargetFactionID == factionID
			}) {
				return true
			}
		}
	}
	return false
}
```

##### Task 2 — `internal/faction/domain/location.go`

**Find:**
```go
	CargoAssetIDs []string            `toml:"cargo_asset_ids,omitempty" json:"cargo_asset_ids,omitempty"`
}
```
**Replace with:**
```go
	CargoAssetIDs []string            `toml:"cargo_asset_ids,omitempty" json:"cargo_asset_ids,omitempty"`
	// LockdownPaid marks an order that has paid its way through a Transport
	// Lockdown and waited its turn; it lands without paying again (Effort 5
	// Decision 14). A revised order is a new order and pays again.
	LockdownPaid bool `toml:"lockdown_paid,omitempty" json:"lockdown_paid,omitempty"`
}
```

##### Task 3 — `internal/faction/domain/mutation.go`, `effect.go`

`mutation.go`:

**Find:**
```go
	// HoldRefused: the destination's government refused an asset aboard (D-27).
	HoldRefused HoldReason = "permission_refused"
)
```
**Replace with:**
```go
	// HoldRefused: the destination's government refused an asset aboard (D-27).
	HoldRefused HoldReason = "permission_refused"
	// HoldLockdownPaid: the order paid a Transport Lockdown's 1d4 and waits a turn (D-33).
	HoldLockdownPaid HoldReason = "lockdown_paid"
	// HoldLockdownUnpaid: the faction cannot pay the 1d4 yet (overview Decision 15).
	HoldLockdownUnpaid HoldReason = "lockdown_unpaid"
)
```

`effect.go`:

**Find:**
```go
	SpecialWithdrawPermission SpecialEffectType = "withdraw_permission"
)
```
**Replace with:**
```go
	SpecialWithdrawPermission SpecialEffectType = "withdraw_permission"
	SpecialLockOut SpecialEffectType = "lock_out"
)
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

The hold's apply arm marks a paid lockdown.

**Find:**
```go
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{RegionHex: v.RegionHex}
				} else {
```
**Replace with:**
```go
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{RegionHex: v.RegionHex}
					if v.Reason == domain.HoldLockdownPaid {
						asset.CurrentOrder.LockdownPaid = true
					}
				} else {
```

##### Task 5 — `internal/faction/rulebook/rulebook.go`, `effects.go`

`rulebook.go`:

**Find:**
```go
	IgnoresPermission bool     `toml:"ignores_permission"`
}
```
**Replace with:**
```go
	IgnoresPermission bool     `toml:"ignores_permission"`
	IgnoresLockdown   bool     `toml:"ignores_lockdown"`
}
```

**Find:**
```go
		IgnoresPermission: r.IgnoresPermission,
	}, nil
```
**Replace with:**
```go
		IgnoresPermission: r.IgnoresPermission,
		IgnoresLockdown:   r.IgnoresLockdown,
	}, nil
```

`effects.go`:

**Find:**
```go
		domain.SpecialGrantStealth, domain.SpecialGrantPermission, domain.SpecialWithdrawPermission,
	}
```
**Replace with:**
```go
		domain.SpecialGrantStealth, domain.SpecialGrantPermission, domain.SpecialWithdrawPermission,
		domain.SpecialLockOut,
	}
```

##### Task 6 — `internal/faction/engine/effect/effects/lock_out.go` (new), `registrars.go`

`lock_out.go` — full contents:
```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// lockOutRegistrar is Transport Lockdown: on a hit, the faction it struck pays 1d4
// Coin and waits a turn for each arrival on the Lockdown's world, and cannot
// relocate there, while the Lockdown stays active on that world (D-31 to D-33).
// The record is a locked_out condition the Lockdown carries on itself, naming the
// struck faction (D-36, overview Decision 24). A hit on a Base counts: the book's
// attack is "against a rival faction".
var lockOutRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerLockOut,
}

func registerLockOut(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialLockOut)+":"+ctx.Def.ID,
		&lockOutReactor{factionID: ctx.Faction.ID, assetID: ctx.Asset.ID},
	)
}

type lockOutReactor struct {
	factionID string
	assetID   string
}

func (reactor *lockOutReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if !active(liveAsset(factionState, reactor.factionID, reactor.assetID)) {
		return nil
	}
	// deliberate: no lock from a Lockdown the batch removes or captures, because its condition would apply after its removal and miss.
	if domain.RemovesOrMoves(mutations, reactor.assetID) {
		return nil
	}
	var struck []string
	var locks []domain.Mutation
	for _, targetID := range hitTargets(mutations, reactor.assetID, false) {
		if slices.Contains(struck, targetID) {
			continue
		}
		struck = append(struck, targetID)
		locks = append(locks, domain.AssetConditionApplied{
			FactionID: reactor.factionID,
			AssetID:   reactor.assetID,
			Condition: domain.Condition{
				Kind:            domain.ConditionLockedOut,
				SourceAssetID:   reactor.assetID,
				TargetFactionID: targetID,
			},
			Cause:             string(domain.SpecialLockOut),
			CausedByFactionID: reactor.factionID,
		})
	}
	return locks
}
```

`registrars.go`:

**Find:**
```go
		domain.SpecialWithdrawPermission: withdrawPermissionRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialWithdrawPermission: withdrawPermissionRegistrar,
		domain.SpecialLockOut:            lockOutRegistrar,
	}
```

##### Task 7 — `internal/faction/engine/arrival.go`

Rewritten — full contents:
```go
package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// checkArrivals decides, for each order the tick completes, whether it may land,
// and returns the tick's mutations with each held completion replaced by the hold
// (D-24, D-27). The government asked is the one ruling at arrival. It keeps a
// running total of the Coin the tick's lockdowns spend, so two arrivals cannot
// spend the same Coin. It runs between the tick and its reactors, so the transport
// reactor moves the cargo to the edge with its carrier (D-28).
func (e *Engine) checkArrivals(faction *domain.Faction, factionState *state.FactionState, tickMutations []domain.Mutation, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	checked := make([]domain.Mutation, 0, len(tickMutations))
	spent := 0
	for _, mutation := range tickMutations {
		completed, ok := mutation.(domain.MovementOrderCompleted)
		if !ok {
			checked = append(checked, mutation)
			continue
		}
		hold, cost, err := e.arrivalHold(faction, factionState, completed, collector, faction.Coin-spent)
		if err != nil {
			return nil, err
		}
		spent += cost
		if hold == nil {
			checked = append(checked, mutation)
			continue
		}
		checked = append(checked, hold...)
	}
	return checked, nil
}

// arrivalHold decides one arriving order. A Transport Lockdown on the destination
// comes first (Effort 5 Decision 14). Then the government is asked for each asset
// aboard, the carrier first, then its cargo in manifest order, each for its own
// owner; the first refusal holds the whole order (D-28). coin is what the faction
// can still spend this tick. It returns the mutations that replace the completion
// and the Coin they spend, or nil when everything aboard may land.
func (e *Engine) arrivalHold(faction *domain.Faction, factionState *state.FactionState, completed domain.MovementOrderCompleted, collector hooks.PermissionCollector, coin int) ([]domain.Mutation, int, error) {
	carrier := faction.Assets[completed.AssetID]
	if carrier == nil || carrier.CurrentOrder == nil {
		return nil, 0, nil
	}
	order := carrier.CurrentOrder
	var profile *domain.TransportProfile
	if carrierDef := e.Rulebook.Assets[carrier.DefinitionID]; carrierDef != nil {
		profile = carrierDef.Transport
	}
	ignoresLockdown := profile != nil && profile.IgnoresLockdown
	if !order.LockdownPaid && !ignoresLockdown && domain.LockedOut(factionState.Factions, order.Destination.WorldID, faction.ID) {
		hold, cost := lockdownHold(completed, order, e.Rand.Roll(4), coin)
		return hold, cost, nil
	}
	// deliberate: the carrier's bypass covers its cargo, because the book lets these transports land "any one … asset" past the government's objection (D-25).
	if profile != nil && profile.IgnoresPermission {
		return nil, 0, nil
	}
	for _, asset := range aboard(factionState, carrier) {
		def := e.Rulebook.Assets[asset.DefinitionID]
		requester := factionState.Factions[asset.OwnerID]
		if def == nil || requester == nil {
			continue
		}
		refusal, err := dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)
		if err != nil {
			return nil, 0, err
		}
		if refusal != nil {
			return []domain.Mutation{*refusal, holdAtEdge(completed, order, domain.HoldRefused)}, 0, nil
		}
	}
	return nil, 0, nil
}

// lockdownHold charges the order's 1d4 fee and holds it a turn at the edge; a
// faction that cannot pay holds unpaid and tries again at its next tick (overview
// Decision 15). It returns the mutations and the Coin they spend.
func lockdownHold(completed domain.MovementOrderCompleted, order *domain.MovementOrder, fee, coin int) ([]domain.Mutation, int) {
	if coin < fee {
		return []domain.Mutation{holdAtEdge(completed, order, domain.HoldLockdownUnpaid)}, 0
	}
	return []domain.Mutation{
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: "transport_lockdown", CausedByFactionID: completed.CausedByFactionID},
		holdAtEdge(completed, order, domain.HoldLockdownPaid),
	}, fee
}

// aboard returns the carrier and the cargo its order carries, each found under its
// current owner: a capture can make a rider's owner differ from the carrier's.
func aboard(factionState *state.FactionState, carrier *domain.Asset) []*domain.Asset {
	assets := []*domain.Asset{carrier}
	for _, cargoID := range carrier.CurrentOrder.CargoAssetIDs {
		if owner := domain.AssetOwner(factionState.Factions, cargoID); owner != nil {
			assets = append(assets, owner.Assets[cargoID])
		}
	}
	return assets
}

// holdAtEdge replaces completed with a hold at the last hex before the destination
// (overview Decision 16). Path[0] is the asset's own hex, so on a two-hex path the
// hold is the in-orbit state right after issue.
func holdAtEdge(completed domain.MovementOrderCompleted, order *domain.MovementOrder, reason domain.HoldReason) domain.MovementOrderHeld {
	step := max(len(order.Path)-2, 0)
	return domain.MovementOrderHeld{
		FactionID:         completed.FactionID,
		AssetID:           completed.AssetID,
		WorldID:           order.Destination.WorldID,
		NewStepIdx:        step,
		RegionHex:         order.Path[step],
		Reason:            reason,
		Cause:             domain.CauseMovementTick,
		CausedByFactionID: completed.CausedByFactionID,
	}
}
```

##### Task 8 — `internal/faction/engine/relocate/relocate.go`

The locked worlds leave the destination list (D-33, Decision 16).

**Find:**
```go
	orders, err := request.Collector.SelectRelocateOrders(request.Acting, candidates, candidateWorlds, maxAssets)
```
**Replace with:**
```go
	destinations := lockedOutRemoved(request, candidateWorlds)
	if len(destinations) == 0 {
		return nil, fmt.Errorf("relocate: %w", ErrNoCandidates)
	}
	orders, err := request.Collector.SelectRelocateOrders(request.Acting, candidates, destinations, maxAssets)
```

**Find:**
```go
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, candidateWorlds, orderCost)
```
**Replace with:**
```go
		// DR-2: the primitive validates membership in destinations (the range-filtered set, less the locked worlds).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, destinations, orderCost)
```

Append after `askLanding`:
```go
// lockedOutRemoved returns the worlds a relocate may land on: candidateWorlds less
// those a Transport Lockdown holds the faction out of. An instant move cannot wait
// out the turn, so the lockdown refuses it, for every asset (D-32, D-33). The
// source worlds stay candidateWorlds: an asset on a locked world may still leave.
func lockedOutRemoved(request Request, candidateWorlds []string) []string {
	if request.FactionState == nil {
		return candidateWorlds
	}
	return slices.DeleteFunc(slices.Clone(candidateWorlds), func(worldID string) bool {
		return domain.LockedOut(request.FactionState.Factions, worldID, request.Faction.ID)
	})
}
```

##### Task 9 — `rulebooks/swn/assets/cunning_assets.toml`, `force_assets.toml`

`cunning_assets.toml` — the book's special attack and the lock:

**Find:**
```toml
description = "On a successful Cunning vs. Cunning attack against a rival faction, the rival faction cannot transport assets onto that planet without spending 1d4 Coins and waiting one turn."

[assets.CovertTransitNet]
```
**Replace with:**
```toml
description = "On a successful Cunning vs. Cunning attack against a rival faction, the rival faction cannot transport assets onto that planet without spending 1d4 Coins and waiting one turn."

[assets.TransportLockdown.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"

[[assets.TransportLockdown.effects]]
trigger = "own_attack_hit"
effect  = "lock_out"
[assets.TransportLockdown.effects.filters]
owner = "own"
scope = "self"

[assets.CovertTransitNet]
```

`force_assets.toml`:

**Find:**
```toml
max_cargo = 1
ignores_permission = true

[assets.IntegralProtocols]
```
**Replace with:**
```toml
max_cargo = 1
ignores_permission = true
ignores_lockdown = true

[assets.IntegralProtocols]
```

##### Task 10 — `internal/faction/narrative/digest/build.go`, `wire_templates.go`

`build.go` — a lock-out sits on the attacker's own asset, so the cross event keys on
the condition's target:

**Find:**
```go
				if m.Cause == string(domain.SpecialShelterAsset) {
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialShelterAsset, factionState)
				} else if m.FactionID != record.FactionID {
```
**Replace with:**
```go
				if m.Cause == string(domain.SpecialShelterAsset) {
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialShelterAsset, factionState)
				} else if m.Condition.Kind == domain.ConditionLockedOut {
					rival := m.Condition.TargetFactionID
					getOrCreateBeat(beats, rival, factionState)
					seenFactionIDs[rival] = struct{}{}
					ce := getOrCreateCross(crossMap, record.FactionID, rival, CrossAttack, factionState)
					if !slices.Contains(ce.Conditions, m.Condition.Kind) {
						ce.Conditions = append(ce.Conditions, m.Condition.Kind)
					}
				} else if m.FactionID != record.FactionID {
```

`wire_templates.go`:

**Find:**
```go
	domain.ConditionNoAttack:  "{attacker}'s agitators keep one of {defender}'s assets from attacking.",
}
```
**Replace with:**
```go
	domain.ConditionNoAttack:  "{attacker}'s agitators keep one of {defender}'s assets from attacking.",
	domain.ConditionLockedOut: "{attacker}'s shipping lockdown makes {defender} pay and wait to land there.",
}
```

**Find:**
```go
	domain.HoldRefused: {
		"{faction}'s {asset} holds in orbit over {world}, waiting for leave to land.",
		"Turned back at the edge of {world}, {faction}'s {asset} waits to ask again.",
	},
}
```
**Replace with:**
```go
	domain.HoldRefused: {
		"{faction}'s {asset} holds in orbit over {world}, waiting for leave to land.",
		"Turned back at the edge of {world}, {faction}'s {asset} waits to ask again.",
	},
	domain.HoldLockdownPaid: {
		"{faction}'s {asset} pays its way through the shipping lockdown at {world} and waits a turn in orbit.",
	},
	domain.HoldLockdownUnpaid: {
		"{faction}'s {asset} cannot pay its way through the shipping lockdown at {world} and waits in orbit.",
	},
}
```

##### Task 11 — domain, mutation and effects tests

`internal/faction/domain/asset_test.go` — append:
```go
func TestCondition_IsComparesTheTargetFaction(t *testing.T) {
	bravo := domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "bravo"}
	charlie := domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "charlie"}
	if bravo.Is(charlie) {
		t.Error("lock-outs against two factions are one condition, want two")
	}
	if !bravo.Is(bravo) {
		t.Error("a lock-out is not itself")
	}
}

func TestLockedOut(t *testing.T) {
	lockdown := func(owner string) *domain.Asset {
		return &domain.Asset{ID: "lock", OwnerID: owner, Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 10,
			Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "bravo"}}}
	}
	cases := []struct {
		name    string
		holder  string
		mutate  func(*domain.Asset)
		world   string
		faction string
		want    bool
	}{
		{name: "the struck faction on the Lockdown's world", holder: "alpha", world: "Krylos", faction: "bravo", want: true},
		{name: "another world", holder: "alpha", world: "Hadrian", faction: "bravo"},
		{name: "another faction", holder: "alpha", world: "Krylos", faction: "charlie"},
		{name: "an inactive Lockdown", holder: "alpha", mutate: func(asset *domain.Asset) { asset.Maintained = false }, world: "Krylos", faction: "bravo"},
		{name: "captured into the struck faction", holder: "bravo", world: "Krylos", faction: "bravo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			asset := lockdown(tc.holder)
			if tc.mutate != nil {
				tc.mutate(asset)
			}
			factions := map[string]*domain.Faction{tc.holder: {ID: tc.holder, Assets: map[string]*domain.Asset{"lock": asset}}}
			if got := domain.LockedOut(factions, tc.world, tc.faction); got != tc.want {
				t.Errorf("LockedOut = %v, want %v", got, tc.want)
			}
		})
	}
}
```

`internal/faction/engine/mutation/mutation_test.go` — append:
```go
func TestMutationEngine_MovementOrderHeldMarksAPaidLockdown(t *testing.T) {
	for _, tc := range []struct {
		reason domain.HoldReason
		want   bool
	}{
		{domain.HoldLockdownPaid, true},
		{domain.HoldLockdownUnpaid, false},
		{domain.HoldRefused, false},
	} {
		s := newMutationTestState()
		hexes := []spatial.RegionHex{{RegionID: "r", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 1}}}
		s.Factions["f1"].Assets["a1"].CurrentOrder = &domain.MovementOrder{AssetID: "a1", Path: hexes}
		held := domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1", NewStepIdx: 0, RegionHex: hexes[0], Reason: tc.reason}

		if err := New(slog.New(slog.DiscardHandler)).Apply(s, []domain.Mutation{held}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("%s: Apply: %v", tc.reason, err)
		}
		if got := s.Factions["f1"].Assets["a1"].CurrentOrder.LockdownPaid; got != tc.want {
			t.Errorf("%s: LockdownPaid = %v, want %v", tc.reason, got, tc.want)
		}
	}
}
```

`internal/faction/engine/effect/effects/lock_out_test.go` (new) — full contents:
```go
package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func lockdownState() *state.FactionState {
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Assets: map[string]*domain.Asset{"lock": activeAsset("lock", "SWN-C6-001")}},
		"beta":  {ID: "beta", Assets: map[string]*domain.Asset{"x": activeAsset("x", "SWN-F1-001")}},
	}}
}

func lockOutReactor(t *testing.T, factionState *state.FactionState) hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialLockOut].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["lock"], Def: &domain.AssetDefinition{ID: "SWN-C6-001"}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerOwnAttackHit,
			Effect:  domain.SpecialLockOut,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		},
	})
	reactors := registry.MutationReactorsForAsset("alpha", "lock")
	if len(reactors) != 1 {
		t.Fatalf("reactors = %d, want 1", len(reactors))
	}
	return reactors[0].Hook
}

var lockedOutBeta = domain.AssetConditionApplied{
	FactionID: "alpha", AssetID: "lock",
	Condition: domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "beta"},
	Cause:     "lock_out", CausedByFactionID: "alpha",
}

func TestLockOut(t *testing.T) {
	assetHit := domain.AssetHPDelta{FactionID: "beta", AssetID: "x", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "lock"}
	baseHit := domain.BaseHPDelta{FactionID: "beta", BaseID: "b", Delta: 0, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "lock"}
	counterHit := assetHit
	counterHit.Counter = true
	cases := []struct {
		name      string
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{"a hit on an asset", []domain.Mutation{assetHit}, []domain.Mutation{lockedOutBeta}},
		{"a hit on a Base", []domain.Mutation{baseHit}, []domain.Mutation{lockedOutBeta}},
		{"two hits on one faction lock once", []domain.Mutation{assetHit, baseHit}, []domain.Mutation{lockedOutBeta}},
		{"a counter is not an attack", []domain.Mutation{counterHit}, nil},
		{"a Lockdown the batch removes locks nothing", []domain.Mutation{assetHit, domain.AssetRemoved{FactionID: "alpha", AssetID: "lock", Cause: "attack"}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := lockdownState()

			got := lockOutReactor(t, factionState).OnMutations(tc.mutations, factionState, nil)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("OnMutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}
```

`internal/faction/rulebook/rulebook_test.go` — append:
```go
func TestLoad_TransportLockdown(t *testing.T) {
	rb, err := Load(testDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	lockdown := rb.Assets["SWN-C6-001"]
	wantAttack := &domain.AttackProfile{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning}
	if !reflect.DeepEqual(lockdown.Attack, wantAttack) {
		t.Errorf("Attack = %+v, want the book's special attack, Cunning vs. Cunning with no damage", lockdown.Attack)
	}
	wantEffects := []domain.EffectDefinition{{
		Trigger: domain.TriggerOwnAttackHit,
		Effect:  domain.SpecialLockOut,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
	}}
	if !reflect.DeepEqual(lockdown.Effects, wantEffects) {
		t.Errorf("Effects = %+v, want %+v", lockdown.Effects, wantEffects)
	}
	if !rb.Assets["SWN-F7-001"].Transport.IgnoresLockdown {
		t.Error("Deep Strike Landers do not ignore the lockdown, want them to (Effort 5 Decision 15)")
	}
	if rb.Assets["SWN-W5-003"].Transport.IgnoresLockdown {
		t.Error("Blockade Runners ignore the lockdown, want them caught")
	}
}
```

##### Task 12 — engine and relocate tests

`internal/faction/engine/arrival_test.go` — append:
```go
type fixedRoller struct{ value int }

func (roller fixedRoller) Roll(int) int { return roller.value }

// lockKrylos gives bravo an active Transport Lockdown on Krylos that holds alpha out.
func lockKrylos(factionState *state.FactionState) {
	factionState.Factions["bravo"].Assets = map[string]*domain.Asset{"lock": {
		ID: "lock", OwnerID: "bravo", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 10,
		Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "alpha"}},
	}}
}

func TestCheckArrivals_Lockdown(t *testing.T) {
	cases := []struct {
		name      string
		coin      int
		paid      bool
		landers   bool
		wantAsked bool
		want      []domain.Mutation
	}{
		{
			name: "pays and waits", coin: 5,
			want: []domain.Mutation{
				domain.CoinDelta{FactionID: "alpha", Delta: -3, Cause: "transport_lockdown", CausedByFactionID: "alpha"},
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldLockdownPaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			},
		},
		{
			name: "cannot pay", coin: 2,
			want: []domain.Mutation{
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldLockdownUnpaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			},
		},
		{name: "already paid: asks the government", coin: 5, paid: true, wantAsked: true},
		{name: "Deep Strike Landers get through", coin: 5, landers: true, wantAsked: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture("ship", "troops", 3)
			eng.Rand = fixedRoller{value: 3}
			eng.Rulebook.Assets["ship"].Transport.IgnoresLockdown = tc.landers
			lockKrylos(factionState)
			alpha.Coin = tc.coin
			alpha.Assets["t1"].CurrentOrder.LockdownPaid = tc.paid
			asked := false
			collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) {
				asked = true
				return true, nil
			}}

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if asked != tc.wantAsked {
				t.Errorf("asked the government = %v, want %v (the lockdown comes first)", asked, tc.wantAsked)
			}
			want := tc.want
			if want == nil {
				want = []domain.Mutation{completed}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_TwoLockdownsCannotSpendTheSameCoin(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "", 3)
	eng.Rand = fixedRoller{value: 3}
	lockKrylos(factionState)
	alpha.Coin = 4
	second := alpha.Assets["t1"].Clone()
	second.ID = "t2"
	second.CurrentOrder.AssetID = "t2"
	alpha.Assets["t2"] = second
	secondCompleted := completed
	secondCompleted.AssetID = "t2"

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed, secondCompleted}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	var reasons []domain.HoldReason
	for _, mutation := range got {
		if held, ok := mutation.(domain.MovementOrderHeld); ok {
			reasons = append(reasons, held.Reason)
		}
	}
	if want := []domain.HoldReason{domain.HoldLockdownPaid, domain.HoldLockdownUnpaid}; !reflect.DeepEqual(reasons, want) {
		t.Errorf("hold reasons = %v, want %v (4 Coin pays one 3-Coin fee)", reasons, want)
	}
}
```

`internal/faction/engine/relocate/relocate_test.go` — append:
```go
func TestRun_LockedWorldIsNotADestination(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	works := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "mid", RegionHex: relocHex("r1", 2)}}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"web": web, "wf-1": works}}
	lockdown := &domain.Asset{ID: "lock", OwnerID: "f2", Location: domain.Location{WorldID: "mid"}, Ready: true, Maintained: true, CurrentHP: 10,
		Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "f1"}}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"lock": lockdown}},
	}}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ *domain.Asset, candidates []*domain.Asset, worlds []string, _ int) ([]action.RelocateOrder, error) {
			if slices.Contains(worlds, "mid") {
				t.Errorf("destinations = %v, want the locked world left out (D-33)", worlds)
			}
			if ids := assetIDs(candidates); !slices.Contains(ids, "wf-1") {
				t.Errorf("candidates = %v, want wf-1 on the locked world still free to leave", ids)
			}
			return nil, nil
		})

	_, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(), FactionState: factionState,
	})

	if !errors.Is(err, action.ErrNoSelection) {
		t.Errorf("Run = %v, want ErrNoSelection from the empty pick", err)
	}
}
```
Add `"slices"` to the test file's imports.

##### Task 13 — digest and renderer tests

`internal/faction/narrative/digest/digest_test.go` — append:
```go
func TestLockedOutIsAConditionOnTheCrossEvent(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha", "f2": "Bravo"})
	records := []domain.EventRecord{eventRec("f1", mutRec(t, "asset_condition_applied", domain.AssetConditionApplied{
		FactionID: "f1", AssetID: "lock",
		Condition: domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "f2"},
		Cause:     "lock_out", CausedByFactionID: "f1",
	}))}

	cycleDigest := buildOne(t, records, factionState, minRulebook(nil))

	var found bool
	for _, ce := range cycleDigest.Cross {
		if ce.Attacker.ID == "f1" && ce.Defender.ID == "f2" && slices.Contains(ce.Conditions, domain.ConditionLockedOut) {
			found = true
		}
	}
	if !found {
		t.Errorf("Cross = %+v, want f1 → f2 carrying locked_out", cycleDigest.Cross)
	}
}
```

`internal/faction/narrative/wire_renderer_test.go` — append:
```go
func TestRenderHold_LockdownReasons(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, reason := range []domain.HoldReason{domain.HoldLockdownPaid, domain.HoldLockdownUnpaid} {
		line := renderHold(digest.HoldEvent{AssetName: "Freighter", World: "Krylos", Reason: reason}, "Alpha", rng)
		if !strings.Contains(line, "lockdown") || !strings.Contains(line, "Krylos") {
			t.Errorf("%s: hold line = %q, want the lockdown and the world", reason, line)
		}
	}
}
```

##### Task 14 — docs

`docs/architecture/engine/effect-mutation.md` — the condition line:

**Find:**
```markdown
- `asset_condition_applied` / `asset_condition_cleared` — add / drop a condition, keyed on kind and source asset (Saboteurs' `no_ability`, Marketers' `disabled`, Seditionists' `no_attack`, Boltholes' `sheltered`)
```
**Replace with:**
```markdown
- `asset_condition_applied` / `asset_condition_cleared` — add / drop a condition, keyed on kind, source asset and target faction (Saboteurs' `no_ability`, Marketers' `disabled`, Seditionists' `no_attack`, Boltholes' `sheltered`, and Transport Lockdown's `locked_out`, which sits on the Lockdown itself and names the struck faction)
```

`docs/architecture/engine/world-movement.md` — the tick step:

**Find:**
```markdown
   `checkArrivals` (`engine/arrival.go`) then reads each `MovementOrderCompleted`:
   it asks the destination's government for every `P` asset aboard (the carrier,
   then its cargo), and the first refusal replaces the completion with a
   `PermissionRefused` and a `MovementOrderHeld` one hex short of the world. The
```
**Replace with:**
```markdown
   `checkArrivals` (`engine/arrival.go`) then reads each `MovementOrderCompleted`.
   A Transport Lockdown on the destination comes first: the order pays 1d4 Coin and
   holds a turn (`MovementOrderHeld`, `lockdown_paid`, which sets the order's
   `LockdownPaid`), or holds unpaid until it can pay. Otherwise it asks the
   destination's government for every `P` asset aboard (the carrier, then its
   cargo), and the first refusal replaces the completion with a `PermissionRefused`
   and a `MovementOrderHeld` one hex short of the world. The
```

`docs/architecture/engine/ability-catalog.md`:

**Find:**
```markdown
| **restrict-action on target** | Cat 3 `MutationReactor` | Saboteurs `SWN-C2-002` (no abilities), Transport Lockdown `SWN-C6-001` (no transport in) |
```
**Replace with:**
```markdown
| **restrict-action on target** | Cat 3 `MutationReactor` | Saboteurs `SWN-C2-002` (no abilities), Transport Lockdown `SWN-C6-001` ✅ (`lock_out`: arrivals pay 1d4 and wait, relocates barred) |
```

`docs/rules/swn-faction-mechanics.md` — append after the Lobbyists paragraph (Commit 4
Task 16):

**Find:**
```markdown
A purchase on a faction's homeworld is no grant, so it cannot be contested.
```
**Replace with:**
```markdown
A purchase on a faction's homeworld is no grant, so it cannot be contested.

A successful **Transport Lockdown** attack locks the struck faction out of the Lockdown's world for as long as the Lockdown stays active there. Each of that faction's drift orders arriving on the world pays 1d4 Coin and waits one turn one hex short, then asks for permission as usual; an order that cannot pay waits until it can. The faction cannot relocate any asset onto the world. Deep Strike Landers land past a lockdown; Blockade Runners and stealthed assets do not.
```

##### Commit message

```
feat: Transport Lockdown

- The book's special attack (Cunning vs. Cunning, no damage) and lock_out:
  a hit puts locked_out on the Lockdown, naming the struck faction and the
  world of the hit (D-36, overview Decision 24); Condition gains
  TargetFactionID and WorldID (Decision Record 5)
- domain.LockedOut reads it while the Lockdown is active on the world of
  the hit (D-31)
- A drift arrival pays 1d4 and holds a turn (MovementOrder.LockdownPaid),
  or holds unpaid; the lockdown comes before the permission ask, one fee
  per order (Decision 14)
- A relocate cannot land on a locked world (D-33)
- TransportProfile.IgnoresLockdown on Deep Strike Landers (Decision 15)
- Digest: the lock-out clause and the lockdown hold lines
```
