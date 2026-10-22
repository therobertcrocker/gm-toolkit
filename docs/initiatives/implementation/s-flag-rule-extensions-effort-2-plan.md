# S-flag Rule Extensions (A.001.6) — Effort 2 Plan: Cat 6 interception

## In Plain Words

- **What:** a new step that runs just before damage destroys an asset, and can put
  something else in place of the destruction. Boltholes hide the asset at 0 HP until
  it is repaired. False Front is destroyed in its place. Medical Center brings it back
  at 1 HP for a price, and also makes repairs on its world cheaper. Hostile Takeover
  takes the asset over at 1 HP.
- **Why:** the engine records a destroyed asset as soon as the damage lands, and
  other rules read that record (Scavengers' Coin, kill goals, the narrative). These
  four assets say the asset was *not* destroyed, so the change must happen before the
  record is written, not after.
- **How it fits:** the third effort to run in A.001.6 (see the
  [overview plan](s-flag-rule-extensions-plan.md)). It uses Effort 1's conditions and
  change of owner, and Effort 3's pay question. Effort 4 runs next.

## Context / Goal

Build the effort the overview plan scopes as "Cat 6 interception", on branch
`feature/s-flag-rule-extensions`.

- Discovery: [s-flag-rule-extensions-discovery.md](../discovery/s-flag-rule-extensions-discovery.md)
  — Lethal Damage Interception (Cat 6), Change of Asset Owner, State Held on an Asset;
  D-1, D-7, D-8, D-10, D-37, D-38.
- Overview plan: [s-flag-rule-extensions-plan.md](s-flag-rule-extensions-plan.md) —
  Decisions 5, 6, 7, 13, 17, 18 and 19 bind here. This plan answers the three Effort 2
  open questions (Decisions 1–3, 9 and 12 below).
- Effort 1 plan: [completed/s-flag-rule-extensions-effort-1-plan.md](completed/s-flag-rule-extensions-effort-1-plan.md)
  — `AssetOwnerChanged`, the two-pass apply, `capture_target`; Decision 15 and
  Decision Records 11 and 14.
- Effort 3 plan: [completed/s-flag-rule-extensions-effort-3-plan.md](completed/s-flag-rule-extensions-effort-3-plan.md)
  — `Targetable()`, `hooks.PaymentCollector`, `HalfPrice()`; Decisions 2 and 3,
  Decision Records 2, 6 and 7.
- Rules source: `docs/rules/swn-assets.md` (Boltholes line 59, False Front 73,
  Hostile Takeover 247, Medical Center 257).

## Decisions Ratified in Planning

Every rules reading below is Robert's call, asked this session with both readings and
their effect at the table.

1. **A Cat 6 hook offers; the dispatcher decides.** Robert's call.
   `LethalInterceptor.OfferInterception(LethalContext) *Interception` returns an offer
   or nil. The dispatcher `dispatch.InterceptLethal(registry, ctx, collector)` orders
   the offers, asks the owner of an elective one, checks a price, and spends the
   budget on a yes. Hooks never see a collector: the Cat 1 pattern. Resolves the
   overview OQ "the Cat 6 interface" (signature). Rejected: the hook takes the
   collector (each elective hook repeats the prompt, price and budget code).
2. **The outcome is its mutations; a second `AssetHPDelta` sets the HP.** Robert's
   call. The hit's `AssetHPDelta` (cause `attack`) stays, so `hits()`, the digest's
   damage and the hit reactors read it as before. The save adds a second delta, caused
   by the save's effect, that moves the HP to 0 or 1 (or back, for False Front): the
   Bank pattern, where history shows the loss and the absorption. Resolves the
   overview OQ (outcome type). Rejected: outcome fields beside the mutations (two
   descriptions of one fact); a new `AssetHPSet` mutation (one more apply arm, and
   every `batchHP` reader must learn it).
3. **The Attack tracker folds the mutations.** `sequenceTracker{hp, gone}` replaces
   `assetHPTracker`. An `AssetHPDelta` adds to `hp`; an `AssetRemoved` or an
   `AssetOwnerChanged` marks the asset `gone`. `live` is not gone and HP above 0. The
   four outcomes read: sheltered at 0 HP is not live; restored at 1 HP is live;
   captured is gone (it is now the attacker's own asset, or a counter-captured
   attacker is now the defender's); a sacrificed False Front is gone and its target
   is back at its HP from before the hit. Resolves the overview OQ (tracker).
4. **The walk is D-7, made concrete.** The dispatcher asks every asset the target's
   owner holds, in ID order (each hook checks its own reach), then the striker's own
   bucket. It takes the target side's automatic offers, then its elective ones, then
   the striker's. The first that takes the hit ends the walk; a declined prompt moves
   to the next offer. An interceptor that the batch already removes or moves to
   another faction offers nothing.
5. **Two elective prompts, both in the framework commit.** Robert's call (the Effort 3
   Commit 2 pattern). `hooks.InterceptCollector` embeds `PaymentCollector` and adds
   `ConfirmSacrifice(SacrificeOffer)` (False Front). Medical Center uses
   `ConfirmPayment` with a new `PaymentRestore` reason (overview Decision 19).
   `action.Collector` embeds `InterceptCollector`; `PhaseCollector` does not, because
   no phase destroys an asset by damage. Both overlays ignore Esc (Effort 3 Decision
   Record 7).
6. **Inside a reactor, only automatic interceptors run (D-10).** The `self_damage`
   reactor calls `InterceptLethalAutomatic`. Hostile Takeover never fires there: a
   self-hit has no striker.
7. **Only the two damage emitters dispatch.** `applyAssetDamage` (attack, counter and
   Cracked Comms' self-attack) and the `self_damage` reactor. The other `AssetRemoved`
   emitters are not damage and stay outside Cat 6: refit, sell, the `coin_delta`
   self-removal, `levy_coin`, `capture_target` and the maintenance loss in
   bookkeeping.
8. **An intercepted kill emits no `AssetRemoved`, by design.** Its readers: Scavengers
   pays nothing; the kill goals (`countAssetKillsByCategory`, Invincible Valor) count
   nothing; the digest shows the save, not a loss. Planetary Seizure reads a capture
   through overview Decision 17 and a shelter through Decision 11.
9. **An intercepted defender still counters on a tie.** Robert's call. The book's tie
   is simultaneous ("both Attack and Counterattack succeed"), so a killed defender
   counters today, and Effort 1 Decision 15 lets Treachery's capture land though the
   counter kills it. The counter gate keeps reading state (`defender.Active()`), not
   the tracker; a fence-sign gives the reason. A captured defender's counter lands
   under its old owner. Resolves the overview OQ "the tie case".
10. **A disabled or unready interceptor does nothing.** Robert's call. Every Cat 6 hook
    checks `effects.active()`, as every S hook does (Effort 3 Decision Record 2).
11. **`sheltered` is untouchable and unusable everywhere.** Robert's calls (three).
    `Targetable()` reads it, so `Active()` does too. Ability targets skip it: Marketers
    and Seditionists candidates, the Monopoly target test and its loss list. It stays
    put: drift eligibility, transport cargo and relocate candidates skip it. Planetary
    Seizure does not count it as a rival present, and a shelter in the batch clears it
    as a kill does.
12. **Repair Asset clears `sheltered` when a repair reaches full HP.** Robert's call.
    One writer, in the repair's own batch, which works even when the Boltholes are
    disabled. Resolves the overview OQ "when `sheltered` clears". Rejected: a Boltholes
    reactor (a disabled Boltholes keeps the asset sheltered at full HP); the
    bookkeeping step (untouchable through every rival turn until the owner's turn).
13. **A Boltholes destroyed by damage takes its sheltered assets with it.** Robert's
    calls (two). Only the damage causes (`attack`, `counterstrike`, `self_damage`); a
    sold or refit Boltholes leaves its assets sheltered until repaired. Each removal
    copies the Boltholes removal's cause and faction, so a combat kill of the Boltholes
    is also a combat kill of each asset it held (Scavengers, kill goals). The reactor
    scans every faction, because a capture moves the Boltholes but not the assets it
    shelters, and it does not check `active()`.
14. **False Front undoes the blow; the sacrifice is not a kill.** Robert's calls (two).
    The saved asset keeps its HP from before the hit. The False Front is removed with
    cause `sacrifice_self`, so Scavengers and the kill goals ignore it.
15. **Medical Center: the saved asset's half price, once per Medical Center.** Robert's
    calls (three). The price is `HalfPrice()` of the saved asset's definition
    (Effort 3 Decision Record 6). `once_per = "asset"`: two Medical Centers can each
    save one asset between the owner's turns, so the dispatcher counts the Coin the
    batch already spends (state Coin changes only at apply). The repair discount is
    the owner's only, 1 Coin off per Special Forces or Military Unit asset repaired on
    that world, never below 0, and a second Medical Center adds nothing (`no_stack`).
    The discount is a new Cat 4 member, `RepairCostModifier`, with a `RepairCostQuery`
    that carries the `no_stack` marks (the `AssetCostQuery` pattern). Repair Asset
    gains the registry (overview Decision 18).
16. **Hostile Takeover captures with any damage it deals; Treachery yields to it.**
    Robert's call: attack and counter. Its interceptor checks only that it is the
    striker and the target is a rival's. `capture_target` (Treachery) now also
    fizzles when the batch already moves its target (the first capture wins, Effort 1
    Decision Record 11) or leaves it at 0 HP (sheltered after a later kill).
17. **Names.** Robert's call. Effects `shelter_asset`, `sacrifice_self`,
    `restore_asset`, `seize_target`, `discount_repair`; triggers `lethal_damage` and
    `repair_cost_query`; condition `sheltered`. Each registrar documents its owner and
    scope reading.
18. **No new mutation type; the digest keys on the save's cause.** Each save emits one
    mutation only it emits: `AssetConditionApplied` (`shelter_asset`), `AssetRemoved`
    (`sacrifice_self`), `CoinDelta` (`restore_asset`), `AssetOwnerChanged`
    (`seize_target`). That mutation adds the effect to `CrossEvent.Interceptions`, and
    the renderer adds a clause that names no side, because the saved asset can be
    either side's. Medical Center's Coin goes on the payer's beat. The TUI labels the
    second `AssetHPDelta` with the save. Overview Decision 13 needs only wording.
19. **Five commits in three phases.** Phase 1 is the framework, tested with fixture
    interceptors. Phases 2 and 3 are two assets each. Over budget, not split: Phase 1
    is about 1,130 lines against 800, because 430 of them are `attack.go` in full,
    which is safer to apply whole than as ten hunks; Phases 2 and 3 are about 830.

## Decision Record — Execution

1. **The capture case of `TestAttack_InterceptedDefenderLeavesTheFight` also sets
   1 HP.** Commit 1. With only the `AssetOwnerChanged` the plan names, `d1` stays at
   −3 HP after the hit, so the HP test stops `a2` and the case passes without the
   tracker's `gone` mark. The fixture now adds an `AssetHPDelta` of +4 (1 HP, as
   Hostile Takeover leaves it), so only `gone` takes `d1` out of the fight. The
   fixture guard `g1` is unready, so it is never a defender candidate.
2. **A pre-existing gofmt miss is fixed in Commit 1.** `purchase_enabler_test.go`
   (from the Effort 1 commit) had two misaligned struct fields, so `gofmt -l` was
   not empty before this effort began. Whitespace only.
3. **The `seize_target` unit test has no counter case.** Commit 5. `LethalContext`
   carries no side, so a counter offer is the same call as an attack offer.
   `TestAttack_HostileTakeoverCounterSeizesAttacker` covers the counter.
4. **The Hostile Takeover scenario checks income, not the registry.** Commit 5. The
   Party Machine's `coin_gain` fires at turn start and registers no hook, so "the
   registry holds it under alpha" cannot fail. The test is now
   `TestHostileTakeoverSeizesAndPaysTheNewOwner`: the next cycle's `coin_gain` pays
   alpha, not beta. The registry move for a captured asset with a hook is already
   tested by the Effort 1 capture scenario.

---

## Shared Context

- **Test patterns.** Actions: `fixedRoller{values}` and the gomock `MockCollector`
  (`attack_test.go`, `runAttack`). Effects: `fixedRoller{value}`, `activeAsset`
  (`effects` tests). Scenarios: `testharness.NewHarness`, `addActiveAsset`,
  `runOrderedCycle`, `countByTypeAndCause`, `testharness.CheckStep` with
  `ScriptedCollector`. The harness gives every asset the zero `RegionHex`; put the
  fight's assets on one world and keep other factions off it (Effort 3 Decision Record
  4). A `hooks/dispatch` test cannot import the action mocks; it uses a hand-written
  fake collector.
- **Fixture interceptor** (tests in Commit 1): a type in each test file that needs it,
  `offerInterceptor{offer *hooks.Interception}` whose `OfferInterception` returns
  `offer`, registered with `registry.RegisterLethalInterceptor(hooks.AssetScope(...), "fixture", ...)`.
- **Mock regeneration.** After a change to `action.Collector`, run
  `go generate ./internal/faction/engine/action/...`.
- **Every commit ends green:** `gofmt -l ./internal` prints nothing, and
  `go build ./... && go test ./internal/faction/...` passes. Find anchors in later
  commits assume gofmt ran on the earlier ones.
- **A new `AskKind`** goes before `AskKindCount` and gets an `askEntries` row;
  `TestAskEntriesCoverEveryKind` catches a missing row.
- **Causes.** A save's mutations carry its effect name as `Cause` and the
  interceptor's owner as `CausedByFactionID`. The repair clear is `Cause: "repair"`.

## Out of Scope

- **The repair overlay's price.** `overlay/repairorders.go` shows `n(n+1)/2` per asset,
  without Medical Center's discount, and `RepairAsset.Validate` still needs 1 Coin
  when a discounted repair would be free. The engine charges the right amount. A
  candidate backlog item.
- **Medical Center for a loss that is not damage** (Monopoly, maintenance, a sale):
  Cat 6 is damage only (D-1, Decision 7).
- **The digest's empty `{defender_asset}`** on a hit that destroys nothing (Effort 1
  Out of Scope). A sheltered or saved defender renders the same way.
- **One `DefenderAsset` per cross event.** A Boltholes kill that takes sheltered assets
  with it names the last removal in the cross event, as several kills in one Attack
  do today.

---

## Work Breakdown

**Overview**

- **Phase 1 — The Cat 6 framework**
  - Commit 1 — `feat: cat 6 lethal interception` (12 tasks)
- **Phase 2 — Boltholes and False Front**
  - Commit 2 — `feat: Boltholes` (15 tasks)
  - Commit 3 — `feat: False Front` (8 tasks)
- **Phase 3 — Medical Center and Hostile Takeover**
  - Commit 4 — `feat: Medical Center` (14 tasks)
  - Commit 5 — `feat: Hostile Takeover` (9 tasks)

### Phase 1 — The Cat 6 framework

Add the hook family, its dispatcher and both prompts, and route both damage emitters
through it. No asset uses it yet; fixture interceptors test it.

#### Commit 1 — `feat: cat 6 lethal interception`

##### Task 1 — `internal/faction/engine/hooks/lethal.go` (new) — full contents:

```go
package hooks

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// LethalContext is one hit that would destroy Target, before its AssetRemoved is
// emitted (Cat 6). HPAfter is Target's HP once the hit lands, at most 0, and Damage
// is the hit's damage. Striker is the asset that dealt it, nil for self-damage.
// Batch is the mutations emitted so far in the action or the reactor round, the hit
// included; state has not applied them yet.
type LethalContext struct {
	Target       *domain.Asset
	TargetDef    *domain.AssetDefinition
	HPAfter      int
	Damage       int
	Striker      *domain.Asset
	FactionState *state.FactionState
	Batch        []domain.Mutation
}

// InterceptAsk says what an interception asks its owner before it takes the hit.
type InterceptAsk int

const (
	// InterceptAutomatic takes the hit with no prompt (Boltholes, Hostile Takeover).
	InterceptAutomatic InterceptAsk = iota
	// InterceptSacrifice asks the owner to give up Source (False Front).
	InterceptSacrifice
	// InterceptRestore asks the owner to pay Price (Medical Center).
	InterceptRestore
)

// Interception is one interceptor's offer to take a lethal hit. Mutations replace
// the target's AssetRemoved and carry the whole outcome, the HP it sets included.
// BudgetKey is spent on Source's owner when the offer is taken; empty is unlimited.
type Interception struct {
	Source    *domain.Asset
	Ask       InterceptAsk
	Price     int
	BudgetKey string
	Mutations []domain.Mutation
}

// LethalInterceptor offers to take a hit that would destroy an asset (Cat 6), or
// returns nil when it does not apply. It never prompts: the dispatcher orders the
// offers, asks the owner of an elective one, and spends the budget.
type LethalInterceptor interface {
	OfferInterception(ctx LethalContext) *Interception
}
```

##### Task 2 — `internal/faction/engine/hooks/collector.go`

**Find:**
```go
	// PaymentLiftDisable: pay to lift a disabled condition at the owner's turn start.
	PaymentLiftDisable PaymentReason = "lift_disable"
)
```
**Replace with:**
```go
	// PaymentLiftDisable: pay to lift a disabled condition at the owner's turn start.
	PaymentLiftDisable PaymentReason = "lift_disable"
	// PaymentRestore: pay to restore an asset at 1 HP instead of losing it (Medical Center).
	PaymentRestore PaymentReason = "restore"
)
```

Append to the file:
```go

// SacrificeOffer asks Owner whether to give up Sacrifice so that Saves survives a
// killing blow (False Front).
type SacrificeOffer struct {
	Owner     *domain.Faction
	Sacrifice *domain.Asset
	Saves     *domain.Asset
}

// InterceptCollector asks the owner of an elective Cat 6 interceptor: give up one
// asset, or pay to keep another. action.Collector embeds it; no phase destroys an
// asset by damage, so PhaseCollector does not.
type InterceptCollector interface {
	PaymentCollector
	ConfirmSacrifice(offer SacrificeOffer) (bool, error)
}
```

##### Task 3 — `internal/faction/engine/hooks/registry.go`

**Find:**
```go
type RegisteredTieResolver struct {
	Scope  Scope
	Source string
	Hook   TieResolver
}
```
**Replace with:**
```go
type RegisteredTieResolver struct {
	Scope  Scope
	Source string
	Hook   TieResolver
}

type RegisteredLethalInterceptor struct {
	Scope  Scope
	Source string
	Hook   LethalInterceptor
}
```

**Find:**
```go
	tieResolvers             map[Scope][]RegisteredTieResolver
}
```
**Replace with:**
```go
	tieResolvers             map[Scope][]RegisteredTieResolver
	lethalInterceptors       map[Scope][]RegisteredLethalInterceptor
}
```

**Find:**
```go
	registry.tieResolvers = make(map[Scope][]RegisteredTieResolver)
}
```
**Replace with:**
```go
	registry.tieResolvers = make(map[Scope][]RegisteredTieResolver)
	registry.lethalInterceptors = make(map[Scope][]RegisteredLethalInterceptor)
}
```

**Find:**
```go
	delete(registry.tieResolvers, scope)
}
```
**Replace with:**
```go
	delete(registry.tieResolvers, scope)
	delete(registry.lethalInterceptors, scope)
}
```

**Find:**
```go
// RegisterTieResolver registers a Cat 5 hook under the given scope.
func (registry *Registry) RegisterTieResolver(scope Scope, source string, hook TieResolver) {
	registry.tieResolvers[scope] = append(registry.tieResolvers[scope], RegisteredTieResolver{Scope: scope, Source: source, Hook: hook})
}
```
**Replace with:**
```go
// RegisterTieResolver registers a Cat 5 hook under the given scope.
func (registry *Registry) RegisterTieResolver(scope Scope, source string, hook TieResolver) {
	registry.tieResolvers[scope] = append(registry.tieResolvers[scope], RegisteredTieResolver{Scope: scope, Source: source, Hook: hook})
}

// RegisterLethalInterceptor registers a Cat 6 hook under the given scope.
func (registry *Registry) RegisterLethalInterceptor(scope Scope, source string, hook LethalInterceptor) {
	registry.lethalInterceptors[scope] = append(registry.lethalInterceptors[scope], RegisteredLethalInterceptor{Scope: scope, Source: source, Hook: hook})
}
```

Append to the file:
```go

// LethalInterceptorsForAsset returns the Cat 6 hooks registered against one asset
// instance. Every interceptor is an asset's, so there is no global or faction lookup.
func (registry *Registry) LethalInterceptorsForAsset(factionID, assetInstanceID string) []RegisteredLethalInterceptor {
	return registry.lethalInterceptors[AssetScope(factionID, assetInstanceID)]
}
```

##### Task 4 — `internal/faction/engine/hooks/doc.go`

**Find:**
```go
// Package hooks defines the five reactive-mechanic hook families and the
// scoped registry that binds them.
//
// # The five categories
//
//	Cat 1  RollModifier / RollOfferFilter  pre-roll dice augmentation and offer removal  interfaces.go
//	Cat 2  RollResultHook                  post-roll inspection & rerolls                interfaces.go
//	Cat 3  MutationReactor                 post-action mutation reactions                interfaces.go
//	Cat 4  RuleModifier family             query-time rule alterations                   rule_modifier.go
//	Cat 5  TieResolver                     tie-point rule overrides                      interfaces.go
```
**Replace with:**
```go
// Package hooks defines the six reactive-mechanic hook families and the
// scoped registry that binds them.
//
// # The six categories
//
//	Cat 1  RollModifier / RollOfferFilter  pre-roll dice augmentation and offer removal  interfaces.go
//	Cat 2  RollResultHook                  post-roll inspection & rerolls                interfaces.go
//	Cat 3  MutationReactor                 post-action mutation reactions                interfaces.go
//	Cat 4  RuleModifier family             query-time rule alterations                   rule_modifier.go
//	Cat 5  TieResolver                     tie-point rule overrides                      interfaces.go
//	Cat 6  LethalInterceptor               lethal-damage interception                    lethal.go
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/lethal.go` (new) — full contents:

```go
package dispatch

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// InterceptLethal asks the Cat 6 interceptors whether one takes ctx's lethal hit,
// in D-7's order: the target owner's automatic interceptors, then its elective
// ones, then the striker's, each group in asset ID order. The first that takes the
// hit ends the walk, and its mutations replace the target's AssetRemoved. It
// returns nil when none takes it. A nil collector asks nothing, so only automatic
// interceptors can take the hit (D-10).
func InterceptLethal(registry *hooks.Registry, ctx hooks.LethalContext, collector hooks.InterceptCollector) ([]domain.Mutation, error) {
	if registry == nil {
		return nil, nil
	}
	targetSide, strikerSide := lethalOffers(registry, ctx)
	for _, side := range [][]*hooks.Interception{targetSide, strikerSide} {
		for _, elective := range []bool{false, true} {
			for _, offer := range side {
				if (offer.Ask != hooks.InterceptAutomatic) != elective {
					continue
				}
				taken, err := takeInterception(offer, ctx, collector)
				if err != nil {
					return nil, err
				}
				if taken {
					return offer.Mutations, nil
				}
			}
		}
	}
	return nil, nil
}

// InterceptLethalAutomatic runs only the automatic interceptors, for a lethal hit
// inside a Cat 3 reactor, which has no collector to ask (D-10).
func InterceptLethalAutomatic(registry *hooks.Registry, ctx hooks.LethalContext) []domain.Mutation {
	// deliberate: drop the error, not return it, because with no collector nothing is asked and nothing can fail.
	mutations, _ := InterceptLethal(registry, ctx, nil)
	return mutations
}

// lethalOffers collects the offers for ctx: from each asset the target's owner
// holds, in ID order, then from the striker. Each interceptor decides whether its
// reach covers the target.
func lethalOffers(registry *hooks.Registry, ctx hooks.LethalContext) (targetSide, strikerSide []*hooks.Interception) {
	if owner := ctx.FactionState.Factions[ctx.Target.OwnerID]; owner != nil {
		for _, asset := range domain.SortedAssets(owner) {
			targetSide = append(targetSide, offersFrom(registry.LethalInterceptorsForAsset(owner.ID, asset.ID), ctx)...)
		}
	}
	if ctx.Striker != nil {
		strikerSide = offersFrom(registry.LethalInterceptorsForAsset(ctx.Striker.OwnerID, ctx.Striker.ID), ctx)
	}
	return targetSide, strikerSide
}

// offersFrom asks each registered interceptor for its offer. An interceptor that
// the batch already removes or moves to another faction offers nothing: a False
// Front sacrificed earlier in the action cannot be sacrificed again.
func offersFrom(registered []hooks.RegisteredLethalInterceptor, ctx hooks.LethalContext) []*hooks.Interception {
	var offers []*hooks.Interception
	for _, interceptor := range registered {
		offer := interceptor.Hook.OfferInterception(ctx)
		if offer != nil && !leftInBatch(ctx.Batch, offer.Source.ID) {
			offers = append(offers, offer)
		}
	}
	return offers
}

// leftInBatch reports whether batch removes assetID or moves it to another faction.
func leftInBatch(batch []domain.Mutation, assetID string) bool {
	for _, mutation := range batch {
		switch m := mutation.(type) {
		case domain.AssetRemoved:
			if m.AssetID == assetID {
				return true
			}
		case domain.AssetOwnerChanged:
			if m.AssetID == assetID {
				return true
			}
		}
	}
	return false
}

// takeInterception reports whether offer takes the hit. An automatic offer always
// does. An elective one does when its owner says yes, and a restore only when the
// owner can pay. A taken offer spends its budget on the owner.
func takeInterception(offer *hooks.Interception, ctx hooks.LethalContext, collector hooks.InterceptCollector) (bool, error) {
	owner := ctx.FactionState.Factions[offer.Source.OwnerID]
	if owner == nil || owner.BudgetSpent(offer.BudgetKey) {
		return false, nil
	}
	switch offer.Ask {
	case hooks.InterceptSacrifice:
		if collector == nil {
			return false, nil
		}
		confirmed, err := collector.ConfirmSacrifice(hooks.SacrificeOffer{Owner: owner, Sacrifice: offer.Source, Saves: ctx.Target})
		if err != nil || !confirmed {
			return false, err
		}
	case hooks.InterceptRestore:
		// deliberate: count the Coin the batch already spends, not only state Coin, because two Medical Centers can each restore in one Attack and state Coin changes only at apply (Effort 2 Decision 15).
		if collector == nil || owner.Coin+batchCoin(ctx.Batch, owner.ID) < offer.Price {
			return false, nil
		}
		confirmed, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: owner, Asset: ctx.Target, Amount: offer.Price, Reason: hooks.PaymentRestore})
		if err != nil || !confirmed {
			return false, err
		}
	}
	owner.SpendBudget(offer.BudgetKey)
	return true, nil
}

// batchCoin is the net Coin change batch makes to factionID before it applies.
func batchCoin(batch []domain.Mutation, factionID string) int {
	total := 0
	for _, mutation := range batch {
		if coin, ok := mutation.(domain.CoinDelta); ok && coin.FactionID == factionID {
			total += coin.Delta
		}
	}
	return total
}
```

##### Task 6 — `internal/faction/engine/action/collector.go`

In `Collector`, **find** `	hooks.PaymentCollector` **replace with**
`	hooks.InterceptCollector`. Run `go generate ./internal/faction/engine/action/...`.

##### Task 7 — `internal/faction/engine/action/actions/attack.go` (rewritten) — full contents:

```go
package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// AttackAction targets rival assets with one or more of the faction's own
// assets. All attackers are committed up front; the matchup sequence runs to
// completion. Each attacker may attack once; a defender may defend multiple
// times. Only known (non-stealthy) rival assets may be targeted.
type AttackAction struct {
	collector action.Collector
	roller    domain.Roller
	registry  *hooks.Registry
	index     *world.Index
	effects   *effect.EffectsEngine
	attackers []*domain.Asset
	mutations []domain.Mutation
}

func NewAttack(collector action.Collector, roller domain.Roller, registry *hooks.Registry, index *world.Index, effects *effect.EffectsEngine) *AttackAction {
	return &AttackAction{collector: collector, roller: roller, registry: registry, index: index, effects: effects}
}

func (attack *AttackAction) Name() string { return "Attack" }

func (attack *AttackAction) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, attacker := range eligibleAttackers(faction, factionState) {
		if attackerHasTarget(attacker, rulebook, attack.registry, attack.index) {
			return true
		}
	}
	return false
}

func (attack *AttackAction) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	var candidates []*domain.Asset
	for _, attacker := range eligibleAttackers(faction, factionState) {
		if attackerHasTarget(attacker, rulebook, attack.registry, attack.index) {
			candidates = append(candidates, attacker)
		}
	}

	selected, err := attack.collector.SelectAttackers(candidates, rulebook)
	if err != nil {
		return fmt.Errorf("attack: %w", err)
	}
	attack.attackers = selected
	return nil
}

func (attack *AttackAction) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// The tracker reflects what this sequence did before its mutations apply, so
	// mid-sequence re-checks see each asset's HP and whether it left the fight.
	tracker := newSequenceTracker()
	baseHPTracker := make(map[string]int)
	// Stealth-cleared tracks assets already cleared this sequence; defenders can
	// face multiple matchups and we emit the mutation only once.
	stealthCleared := make(map[string]bool)

	for _, attacker := range attack.attackers {
		attackerDef, ok := rulebook.Assets[attacker.DefinitionID]
		if !ok || attackerDef.Attack == nil {
			continue
		}

		// Re-check: skip if destroyed or taken out of the fight earlier in this sequence.
		if !tracker.live(attacker) {
			continue
		}

		defenders := liveDefendersAtHex(attacker, attackerDef, tracker, rulebook, attack.registry, attack.index)
		if len(defenders) == 0 {
			continue
		}

		defender, err := attack.collector.SelectDefender(attacker, defenders, rulebook)
		if err != nil {
			return fmt.Errorf("attack: selecting defender: %w", err)
		}

		defenderDef, ok := rulebook.Assets[defender.DefinitionID]
		if !ok {
			return fmt.Errorf("attack: defender definition not found: %s", defender.DefinitionID)
		}

		defenderFaction := ownerFaction(factionState, defender)
		if defenderFaction == nil {
			return fmt.Errorf("attack: owner faction not found for defender %s", defender.ID)
		}

		// Stealth loss is emitted before damage so the EventRecord timeline is consistent.
		if attacker.Stealthy && !stealthCleared[attacker.ID] {
			attack.mutations = append(attack.mutations, domain.AssetStealthCleared{
				FactionID:         faction.ID,
				AssetID:           attacker.ID,
				Cause:             "attack",
				CausedByFactionID: faction.ID,
			})
			stealthCleared[attacker.ID] = true
		}
		if defender.Stealthy && !stealthCleared[defender.ID] {
			attack.mutations = append(attack.mutations, domain.AssetStealthCleared{
				FactionID:         defenderFaction.ID,
				AssetID:           defender.ID,
				Cause:             "attack",
				CausedByFactionID: faction.ID,
			})
			stealthCleared[defender.ID] = true
		}

		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState).Mutations...)

		attackResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:         hooks.PhaseAttack,
				Actor:         faction,
				Opponent:      defenderFaction,
				Attribute:     string(attackerDef.Attack.AttackerStat),
				Asset:         attacker,
				OpponentAsset: defender,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.AttackerStat)},
			attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
		)
		attackRoll := attackResult.Sum

		defResult := dispatch.RollWithHooks(
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         defenderFaction,
				Opponent:      faction,
				Attribute:     string(attackerDef.Attack.DefenderStat),
				Asset:         defender,
				OpponentAsset: attacker,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
			domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(defenderFaction, attackerDef.Attack.DefenderStat)},
			attack.registry, attack.collector, attack.roller, defenderFaction, factionState, rulebook,
		)
		defenseRoll := defResult.Sum

		tieCtx := hooks.RollContext{
			Phase:    hooks.PhaseAttack,
			Actor:    faction,
			Opponent: defenderFaction,
			Asset:    attacker,
			World:    attacker.Location.WorldID,
		}
		tieOutcome := dispatch.ResolveTie(attack.registry, tieCtx, factionState)
		attackerWins := attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins)
		defenderWins := defenseRoll > attackRoll || (defenseRoll == attackRoll && tieOutcome != hooks.TieAttackerWins)

		// Attack damage: attacker wins on tie (TieStandard/TieAttackerWins) or strictly greater.
		if attackerWins {
			if attackerDef.Attack.Damage == nil {
				// deliberate: a special hit skips the redirect prompt, because it has no damage to send to a Base (D-15).
				if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, 0, tracker, attacker, false, "attack"); err != nil {
					return err
				}
			} else {
				damage := attackerDef.Attack.Damage.Roll(attack.roller)
				base := factionBaseOnWorld(defenderFaction, attacker.Location.WorldID)

				// Redirect prompt offered only when the defender has a live Base on this world.
				if base != nil && base.CurrentHP+baseHPTracker[base.ID] > 0 {
					redirect, err := attack.collector.ConfirmRedirectToBase(defenderFaction, base, damage)
					if err != nil {
						return fmt.Errorf("attack: redirect prompt: %w", err)
					}
					if redirect {
						attack.mutations = append(attack.mutations,
							domain.BaseHPDelta{FactionID: defenderFaction.ID, BaseID: base.ID, Delta: -damage, Cause: "attack", CausedByFactionID: faction.ID, CausedByAssetID: attacker.ID},
							// SWN: damage to a Base is also dealt directly to faction HP.
							domain.FactionHPDelta{FactionID: defenderFaction.ID, Delta: -damage, Cause: "attack", CausedByFactionID: faction.ID},
						)
						baseHPTracker[base.ID] -= damage
						if base.CurrentHP+baseHPTracker[base.ID] <= 0 {
							attack.mutations = append(attack.mutations, domain.BaseDestroyed{
								FactionID:         defenderFaction.ID,
								BaseID:            base.ID,
								Cause:             "attack",
								CausedByFactionID: faction.ID,
							})
						}
					} else if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, damage, tracker, attacker, false, "attack"); err != nil {
						return err
					}
				} else if err := attack.applyAssetDamage(factionState, rulebook, defenderFaction.ID, faction.ID, defender, damage, tracker, attacker, false, "attack"); err != nil {
					return err
				}
			}
		}

		// deliberate: a disabled defender rolls its defense but deals no counter and fires no own_defense_win entry, because the book makes it useless (Effort 3 Decision 2).
		// deliberate: read state, not the tracker, because on a tie both sides hit at once: a defender this matchup killed or intercepted still counters (Effort 2 Decision 9).
		defenderActive := defender.Active()

		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenderWins && defenderActive {
			if defenderDef.Counter != nil && dispatch.ResolveAttackEligibility(attack.registry, hooks.AttackEligibilityQuery{
				Attacker:    defender,
				AttackerDef: defenderDef,
				Defender:    attacker,
				DefenderDef: attackerDef,
				Side:        hooks.SideCounter,
			}) {
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				if err := attack.applyAssetDamage(factionState, rulebook, faction.ID, defenderFaction.ID, attacker, counterDamage, tracker, defender, true, "attack"); err != nil {
					return err
				}
			}
		}

		// deliberate: own_defense_win needs an outright win, not a tie, because on a tie the attack also landed and the defense did not succeed (Decision 13).
		if defenderWins && !attackerWins && defenderActive {
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState)
			attack.mutations = append(attack.mutations, outcome.Mutations...)
			// deliberate: no prompt when the attacker has no damage and no counter, because its self-attack could do nothing (Decision Record 9).
			canHurtItself := attackerDef.Attack.Damage != nil || attackerDef.Counter != nil
			if outcome.Counterstrike && canHurtItself && tracker.live(attacker) {
				confirmed, err := attack.collector.ConfirmCounterstrike(defenderFaction, defender, attacker)
				if err != nil {
					return fmt.Errorf("attack: counterstrike prompt: %w", err)
				}
				if confirmed {
					if err := attack.counterstrike(faction, attacker, attackerDef, defenderFaction, defender, tracker, factionState, rulebook); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// counterstrike turns the attacker on itself (D-18): its attack roll against its
// own defense. A hit deals its attack damage; a miss deals its counter damage, if
// it has one. The cause is not "attack", so no own_attack_hit fires on it.
func (attack *AttackAction) counterstrike(faction *domain.Faction, attacker *domain.Asset, attackerDef *domain.AssetDefinition, defenderFaction *domain.Faction, source *domain.Asset, tracker *sequenceTracker, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	worldID := attacker.Location.WorldID
	// deliberate: Opponent nil, not faction, because Actor == Opponent would ask every asset hook on the world twice (overview Decision 9).
	attackRoll := dispatch.RollWithHooks(
		hooks.RollContext{Phase: hooks.PhaseAttack, Actor: faction, Attribute: string(attackerDef.Attack.AttackerStat), Asset: attacker, World: worldID, Action: attack.Name()},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.AttackerStat)},
		attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
	).Sum
	defenseRoll := dispatch.RollWithHooks(
		hooks.RollContext{Phase: hooks.PhaseDefense, Actor: faction, Attribute: string(attackerDef.Attack.DefenderStat), Asset: attacker, World: worldID, Action: attack.Name()},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, attackerDef.Attack.DefenderStat)},
		attack.registry, attack.collector, attack.roller, faction, factionState, rulebook,
	).Sum
	tieOutcome := dispatch.ResolveTie(attack.registry, hooks.RollContext{Phase: hooks.PhaseAttack, Actor: faction, Asset: attacker, World: worldID}, factionState)

	damage := attackerDef.Counter
	if attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins) {
		damage = attackerDef.Attack.Damage
	}
	if damage == nil {
		return nil
	}
	return attack.applyAssetDamage(factionState, rulebook, faction.ID, defenderFaction.ID, attacker, damage.Roll(attack.roller), tracker, source, false, string(domain.SpecialCounterstrike))
}

func (attack *AttackAction) Output() ([]domain.Mutation, error) {
	return attack.mutations, nil
}

// fireMatchup raises the matchup route for the declaring asset. A nil effects
// engine (action tests) raises nothing.
func (attack *AttackAction) fireMatchup(trigger domain.EffectTrigger, faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, opponent *domain.Faction, opponentAsset *domain.Asset, factionState *state.FactionState) effect.MatchupOutcome {
	if attack.effects == nil {
		return effect.MatchupOutcome{}
	}
	return attack.effects.FireMatchup(trigger, effect.MatchupContext{
		Faction:       faction,
		Asset:         asset,
		Def:           def,
		Opponent:      opponent,
		OpponentAsset: opponentAsset,
		FactionState:  factionState,
	})
}

// applyAssetDamage emits AssetHPDelta and, when the hit is lethal, the mutations
// of the Cat 6 interceptor that takes it, or else AssetRemoved. factionID is the
// owner of the damaged asset; causedByFactionID is the striker's. Every mutation
// goes through the tracker, so later re-checks see the outcome.
func (attack *AttackAction) applyAssetDamage(factionState *state.FactionState, rulebook *rulebook.Rulebook, factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker *sequenceTracker, causedByAsset *domain.Asset, counter bool, cause string) error {
	causedByAssetID := ""
	if causedByAsset != nil {
		causedByAssetID = causedByAsset.ID
	}
	attack.emit(tracker, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             cause,
		CausedByFactionID: causedByFactionID,
		CausedByAssetID:   causedByAssetID,
		Counter:           counter,
	})
	hpAfter := tracker.hpOf(asset)
	if hpAfter > 0 {
		return nil
	}
	intercepted, err := dispatch.InterceptLethal(attack.registry, hooks.LethalContext{
		Target:       asset,
		TargetDef:    rulebook.Assets[asset.DefinitionID],
		HPAfter:      hpAfter,
		Damage:       damage,
		Striker:      causedByAsset,
		FactionState: factionState,
		Batch:        attack.mutations,
	}, attack.collector)
	if err != nil {
		return fmt.Errorf("attack: lethal interception: %w", err)
	}
	if intercepted != nil {
		attack.emit(tracker, intercepted...)
		return nil
	}
	attack.emit(tracker, domain.AssetRemoved{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Cause:             cause,
		CausedByFactionID: causedByFactionID,
	})
	return nil
}

// emit appends mutations to the action's output and folds them into the tracker.
func (attack *AttackAction) emit(tracker *sequenceTracker, mutations ...domain.Mutation) {
	attack.mutations = append(attack.mutations, mutations...)
	tracker.record(mutations)
}

// sequenceTracker is what the Attack sequence has done to each asset before its
// batch applies: the HP change, and the assets that left the fight (removed, or
// moved to another faction).
type sequenceTracker struct {
	hp   map[string]int
	gone map[string]bool
}

func newSequenceTracker() *sequenceTracker {
	return &sequenceTracker{hp: make(map[string]int), gone: make(map[string]bool)}
}

// record folds mutations in: HP deltas add up, and a removal or an owner change
// takes the asset out of the rest of the sequence (Effort 2 Decision 3).
func (tracker *sequenceTracker) record(mutations []domain.Mutation) {
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetHPDelta:
			tracker.hp[m.AssetID] += m.Delta
		case domain.AssetRemoved:
			tracker.gone[m.AssetID] = true
		case domain.AssetOwnerChanged:
			tracker.gone[m.AssetID] = true
		}
	}
}

// hpOf is asset's HP once this sequence's mutations apply.
func (tracker *sequenceTracker) hpOf(asset *domain.Asset) int {
	return asset.CurrentHP + tracker.hp[asset.ID]
}

// live reports whether asset can still fight in this sequence.
func (tracker *sequenceTracker) live(asset *domain.Asset) bool {
	return !tracker.gone[asset.ID] && tracker.hpOf(asset) > 0
}

// attackerHasTarget returns true if the attacker has a valid Attack profile and
// at least one eligible defender on its world.
func attackerHasTarget(attacker *domain.Asset, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) bool {
	def, ok := rulebook.Assets[attacker.DefinitionID]
	if !ok || def.Attack == nil {
		return false
	}
	return len(eligibleDefendersAtHex(attacker, def, rulebook, registry, index)) > 0
}

// liveDefendersOnWorld filters eligibleDefendersOnWorld to the assets still live
// in the current Resolve pass.
func liveDefendersOnWorld(attackerFactionID, worldID string, tracker *sequenceTracker, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersOnWorld(attackerFactionID, worldID, index), tracker)
}

// liveDefendersAtHex filters eligibleDefendersAtHex to the assets still live in
// the current Resolve pass. Used when the attacker is mid-flight (empty WorldID).
func liveDefendersAtHex(attacker *domain.Asset, attackerDef *domain.AssetDefinition, tracker *sequenceTracker, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) []*domain.Asset {
	return filterLiveDefenders(eligibleDefendersAtHex(attacker, attackerDef, rulebook, registry, index), tracker)
}

func filterLiveDefenders(defenders []*domain.Asset, tracker *sequenceTracker) []*domain.Asset {
	live := make([]*domain.Asset, 0, len(defenders))
	for _, defender := range defenders {
		if tracker.live(defender) {
			live = append(live, defender)
		}
	}
	return live
}

// ownerFaction returns the faction that owns the given asset, or nil.
func ownerFaction(factionState *state.FactionState, asset *domain.Asset) *domain.Faction {
	return factionState.Factions[asset.OwnerID]
}
```

##### Task 8 — `internal/faction/engine/action/actions/spatial_index_test.go`

In `TestLiveDefenders_ExcludesTrackerDead`, **find:**
```go
	tracker := map[string]int{"a3": -4}
```
**Replace with:**
```go
	tracker := newSequenceTracker()
	tracker.hp["a3"] = -4
```

##### Task 9 — `internal/faction/engine/effect/effects/self_damage.go`

Add the imports `slices` and `github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch`.

**Find:**
```go
		&selfDamageReactor{
			factionID: ctx.Faction.ID,
			assetID:   ctx.Asset.ID,
			counter:   ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
			amount:    ctx.Entry.SelfDamage.Amount,
			roller:    ctx.Roller,
		},
```
**Replace with:**
```go
		&selfDamageReactor{
			factionID: ctx.Faction.ID,
			assetID:   ctx.Asset.ID,
			def:       ctx.Def,
			counter:   ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
			amount:    ctx.Entry.SelfDamage.Amount,
			roller:    ctx.Roller,
			registry:  ctx.Registry,
		},
```

**Find:**
```go
type selfDamageReactor struct {
	factionID string
	assetID   string
	counter   bool
	amount    domain.DiceRoll
	roller    domain.Roller
}
```
**Replace with:**
```go
// selfDamageReactor keeps its definition and the registry, because a lethal
// self-hit asks the Cat 6 interceptors (D-10).
type selfDamageReactor struct {
	factionID string
	assetID   string
	def       *domain.AssetDefinition
	counter   bool
	amount    domain.DiceRoll
	roller    domain.Roller
	registry  *hooks.Registry
}
```

**Find:**
```go
	if effectiveHP-damage <= 0 {
		extra = append(extra, domain.AssetRemoved{
```
**Replace with:**
```go
	if effectiveHP-damage <= 0 {
		// deliberate: only automatic interceptors, not elective ones, because a reactor has no collector to ask (D-10).
		saved := dispatch.InterceptLethalAutomatic(reactor.registry, hooks.LethalContext{
			Target:       asset,
			TargetDef:    reactor.def,
			HPAfter:      effectiveHP - damage,
			Damage:       damage,
			FactionState: factionState,
			Batch:        append(slices.Clone(mutations), extra...),
		})
		if saved != nil {
			return append(extra, saved...)
		}
		extra = append(extra, domain.AssetRemoved{
```

##### Task 10 — prompt surfaces

`tui/adapter/channels.go` — after `AskSelectAssetToLose` add `AskConfirmSacrifice`;
after `type ConfirmPaymentPayload struct{ Offer hooks.PaymentOffer }` add:
```go
type ConfirmSacrificePayload struct{ Offer hooks.SacrificeOffer }
```

`tui/adapter/action_collector.go` — **find:**
```go
func (a *actionCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return a.adapter.confirmPayment(offer)
}
```
**Replace with:**
```go
func (a *actionCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return a.adapter.confirmPayment(offer)
}

// ConfirmSacrifice asks the owner of the False Front, who is often not the acting
// faction.
func (a *actionCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmSacrifice, offer.Owner, ConfirmSacrificePayload{Offer: offer})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
```

`tui/views/turn/overlay/ability.go` — **find:**
```go
	if offer.Reason == hooks.PaymentLiftDisable {
		return ConfirmPayment{NewConfirm(fmt.Sprintf("Pay %d Coin to bring %s back into use?", offer.Amount, name), header, "Pay", "Leave it disabled")}
	}
	return ConfirmPayment{NewConfirm(fmt.Sprintf("Pay %d Coin, or %s is disabled?", offer.Amount, name), header, "Pay", "Take the disable")}
```
**Replace with:**
```go
	switch offer.Reason {
	case hooks.PaymentLiftDisable:
		return ConfirmPayment{NewConfirm(fmt.Sprintf("Pay %d Coin to bring %s back into use?", offer.Amount, name), header, "Pay", "Leave it disabled")}
	case hooks.PaymentRestore:
		return ConfirmPayment{NewConfirm(fmt.Sprintf("Pay %d Coin to restore %s at 1 HP?", offer.Amount, name), header, "Pay", "Let it be destroyed")}
	}
	return ConfirmPayment{NewConfirm(fmt.Sprintf("Pay %d Coin, or %s is disabled?", offer.Amount, name), header, "Pay", "Take the disable")}
```

`tui/views/turn/overlay/attack.go` — add the imports `github.com/charmbracelet/bubbles/help`,
`tea "github.com/charmbracelet/bubbletea"` and
`github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks`. After
`NewConfirmCounterstrike` add:
```go
// ConfirmSacrifice asks the owner whether to give up one asset to save another from
// a killing blow (False Front). Esc does nothing: the decline is an explicit option,
// and a cancel would drop the attacker's whole action (Effort 3 Decision Record 7).
type ConfirmSacrifice struct {
	ConfirmOverlay
}

// NewConfirmSacrifice's header names the faction asked, which is often not the
// acting one. Answer: bool.
func NewConfirmSacrifice(offer hooks.SacrificeOffer, rb *rulebook.Rulebook) ConfirmSacrifice {
	title := fmt.Sprintf("Sacrifice %s to save %s?", assetName(offer.Sacrifice, rb), assetName(offer.Saves, rb))
	header := fmt.Sprintf("Asked of %s", offer.Owner.Name)
	return ConfirmSacrifice{NewConfirm(title, header, "Sacrifice", "Let it fall")}
}

func (o ConfirmSacrifice) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil
	}
	updated, cmd := o.ConfirmOverlay.Update(msg)
	o.ConfirmOverlay = updated.(ConfirmOverlay)
	return o, cmd
}

func (o ConfirmSacrifice) Help() help.KeyMap { return formHelp{} }
```

`tui/views/turn/execution/execution.go` — **find:**
```go
		return overlay.NewSelectAssetToLose(fmt.Sprintf("%s cannot pay: lose which asset?", p.Faction.Name), p.Assets, m.rulebook)
	}},
```
**Replace with:**
```go
		return overlay.NewSelectAssetToLose(fmt.Sprintf("%s cannot pay: lose which asset?", p.Faction.Name), p.Assets, m.rulebook)
	}},
	adapter.AskConfirmSacrifice: {"Sacrifice?", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmSacrifice(msg.Payload.(adapter.ConfirmSacrificePayload).Offer, m.rulebook)
	}},
```

`tui/adapter/smoke.go` — after `SmokeActionCollector.ConfirmPayment` add:
```go
func (SmokeActionCollector) ConfirmSacrifice(_ hooks.SacrificeOffer) (bool, error) {
	return false, action.ErrActionUnavailable
}
```

`engine/testharness/collector.go` — after the `ConfirmPaymentFn` field add
`ConfirmSacrificeFn            func(hooks.SacrificeOffer) (bool, error)`, and after
`ScriptedCollector.ConfirmPayment` add:
```go
func (c *ScriptedCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	if c.ConfirmSacrificeFn != nil {
		return c.ConfirmSacrificeFn(offer)
	}
	return false, nil
}
```

##### Task 11 — docs

`docs/architecture/engine/hooks.md`:
- **Find:**
  ```
  a turn without the core engine knowing they exist. The package defines **five
  families of reactive mechanic** — pre-roll dice augmentation, post-roll rerolls,
  post-mutation reactions, query-time rule alterations, and tie-breaks — plus a
  ```
  **Replace with:**
  ```
  a turn without the core engine knowing they exist. The package defines **six
  families of reactive mechanic** — pre-roll dice augmentation, post-roll rerolls,
  post-mutation reactions, query-time rule alterations, tie-breaks, and lethal-damage
  interception — plus a
  ```
- **find** `### The five categories` **replace with** `### The six categories`.
- After the table row that starts `| 5 | \`TieResolver\`` add:
  ```
  | 6 | `LethalInterceptor` | when a hit would destroy an asset — may take the hit in place of the removal | `InterceptLethal` (the Attack action's `applyAssetDamage`, the `self_damage` reactor) |
  ```
- Before the bullet that starts `- **\`ResolveTie\`** is consulted at a combat tie,` add:
  ```
  - **`InterceptLethal`** is Cat 6. It runs when a hit would destroy an asset,
    before `AssetRemoved` is emitted: in the Attack action's `applyAssetDamage`
    (attack, counter and Cracked Comms' self-attack) and in the `self_damage`
    reactor. It asks the bucket of each asset the target's owner holds, in ID
    order, then the striker's own bucket. Each `LethalInterceptor` returns an
    `Interception` offer or nil, and the dispatcher decides: the target side's
    automatic offers, then its elective ones, then the striker's. **The first that
    takes the hit wins**, and its mutations replace the removal. An elective offer
    asks its owner through the `InterceptCollector`; a restore is offered only when
    the owner can pay, counting the Coin the batch already spends; a yes spends the
    offer's budget. An interceptor that the batch already removes or moves offers
    nothing. A reactor has no collector, so `self_damage` calls
    `InterceptLethalAutomatic`, which skips elective offers.
  ```
- After the paragraph that ends `` `action.Collector` and `PhaseCollector` both embed it. `` add:
  ```

  `InterceptCollector` embeds `PaymentCollector` and adds
  `ConfirmSacrifice(SacrificeOffer) (bool, error)`: the Cat 6 dispatcher asks the
  owner of an elective interceptor whether to give up one asset to save another, or
  to pay to keep it. Only `action.Collector` embeds it, because no phase destroys
  an asset by damage, and a reactor asks nothing.
  ```
- **find** `The other four families are` **replace with** `The other five families are`.
- In the diagram, after the line
  `    Elig["ResolveAttackEligibility<br>Cat 4 — first no refuses"]:::process` add
  `    Lethal["InterceptLethal<br>Cat 6 — first taker wins"]:::process`; after
  `  Registry --> Elig` add `  Registry --> Lethal`; after
  `  Roll <-->|"offers / selection"| Collector` add
  `  Lethal <-->|"sacrifice / pay"| Collector`.
- **find** `Cat 4 attack eligibility in combat)` **replace with**
  `Cat 4 attack eligibility and Cat 6 interception in combat)`.

`docs/architecture/engine/actions.md` — in the Attack row, **find**
`` `SelectAttackers`, `SelectDefender`, `ConfirmRedirectToBase` → `AssetHPDelta`, `AssetRemoved`, `BaseHPDelta`, `FactionHPDelta`, `BaseDestroyed`, `AssetStealthCleared` | ``
**replace with**
`` `SelectAttackers`, `SelectDefender`, `ConfirmRedirectToBase` (`ConfirmSacrifice`, `ConfirmPayment` for a Cat 6 interceptor) → `AssetHPDelta`, `AssetRemoved`, `BaseHPDelta`, `FactionHPDelta`, `BaseDestroyed`, `AssetStealthCleared` (a Cat 6 interception's mutations in place of `AssetRemoved`) | ``.

`docs/architecture/engine/ability-catalog.md`:
- **find** `**five hook` **replace with** `**six hook`.
- **find** `| **death-redirect** (intercept a killing blow) | Cat 3 \`MutationReactor\` |`
  **replace with** `| **death-redirect** (intercept a killing blow) | Cat 6 \`LethalInterceptor\` |`.
- **find**
  `| **acquire-on-kill** (take the asset instead of destroying) | Cat 3 \`MutationReactor\` | Hostile Takeover \`SWN-W7-002\`, Treachery \`SWN-C7-003\` |`
  **replace with**
  `| **acquire-on-kill** (take the asset instead of destroying) | Cat 6 \`LethalInterceptor\` (Hostile Takeover); Cat 3 \`MutationReactor\` → \`capture_target\` (Treachery) | Hostile Takeover \`SWN-W7-002\`, Treachery \`SWN-C7-003\` ✅ |`.

##### Task 12 — tests

- `hooks/registry_test.go`: `TestLethalInterceptor_ForAssetReturnsOnlyThatBucket`
  (two assets, each one interceptor; each lookup returns only its own).
  `TestClearAsset_RemovesOnlyThatAssetsBucket` also registers an interceptor on the
  cleared asset and checks its bucket is empty.
- `hooks/dispatch/lethal_test.go` (new). A fake `InterceptCollector` records each
  `SacrificeOffer` and `PaymentOffer` and returns set answers and an error. State:
  `f1` (`Coin: 3`) holds target `t1`, guards `g1` and `g2`; `f2` holds striker `s1`.
  `TestInterceptLethal`, one table:
  - nil registry → nil, nil.
  - no interceptor → nil, nil, no prompt.
  - `g1` sacrifice (answer yes), `g2` automatic, `s1` automatic → `g2`'s mutations,
    no prompt (automatic first, although `g1` has the lower ID).
  - `g1` sacrifice (yes), `s1` automatic → `g1`'s mutations; the offer names `f1`,
    `g1`, `t1`.
  - `g1` sacrifice (no), `s1` automatic → `s1`'s mutations.
  - nil collector, `g1` sacrifice, `s1` automatic → `s1`'s mutations
    (`InterceptLethalAutomatic`).
  - `g1` restore, `Price: 2`, budget `"k"`, answer yes →
    `PaymentOffer{f1, t1, 2, PaymentRestore}`, `g1`'s mutations, `f1` spent `"k"`.
    The same call again → no prompt, nil.
  - `g1` restore, answer no → nil, `"k"` not spent.
  - `g1` restore, `Price: 2`, batch holds `CoinDelta{f1, -2}` → no prompt, nil.
  - `g2` automatic, batch holds `AssetRemoved{g2}` → nil; holds
    `AssetOwnerChanged{g2}` → nil.
  - sacrifice prompt returns an error → the error, nil mutations.
- `actions/attack_test.go`, with the fixture interceptor on `f2`'s guard `g1`
  (added to `f2.Assets`; `d1` at 3 HP, damage 6):
  - `TestAttack_InterceptedKillEmitsNoRemoval` — automatic, mutations
    `[AssetHPDelta{f2, d1, +3, "fixture"}]` → output ends with the hit `-6` and the
    fixture delta; no `AssetRemoved`.
  - `TestAttack_InterceptedDefenderLeavesTheFight` — two attackers `a1`, `a2`, one
    defender `d1`. The fixture returns `AssetOwnerChanged{f2, d1 → f1}` → `a2` finds
    no defender (`SelectDefender` called once). A second case returns only a delta to
    0 HP → the same.
  - `TestAttack_DeclinedSacrificeRemoves` — fixture `InterceptSacrifice`;
    `ConfirmSacrifice` returns false → `AssetRemoved{d1}`; the offer has `Owner` `f2`
    and `Saves` `d1`.
  - `TestAttack_InterceptionErrorAbortsResolve` — `ConfirmSacrifice` returns
    `action.ErrTurnCanceled` → `Resolve` returns an error that `errors.Is` it.
  - `TestAttack_InterceptedDefenderCountersOnTie` — tie roll, `d1` (with a counter)
    killed and intercepted by an automatic fixture → the counter `AssetHPDelta` on
    `a1` is in the output (Decision 9).
- `actions/spatial_index_test.go`: `TestLiveDefenders_ExcludesTrackerDead` gains an
  asset `a4` marked `tracker.gone` → excluded.
- `effects/self_damage_test.go`: `zealotReactors` takes an extra
  `interceptor hooks.LethalInterceptor` (nil for none), registered under
  `AssetScope("alpha", "bolt")`, and `zealotState` gives alpha an asset `bolt`.
  `TestSelfDamage_AutomaticInterceptorSaves` — Zealots at 2 HP, roll 3, an automatic
  fixture → `[zealotDamage(-3), fixture mutations]`, no `zealotRemoved`. An elective
  fixture → not used: `[zealotDamage(-3), zealotRemoved]`.
- `overlay/asset_test.go`: `TestConfirmPaymentIgnoresEsc` adds `PaymentRestore` to its
  reasons. `TestConfirmSacrificeIgnoresEsc` mirrors it for `NewConfirmSacrifice`.
- `execution_test.go`: `TestAskEntriesCoverEveryKind` passes with the new kind.

##### Commit message

```
feat: cat 6 lethal interception

- hooks.LethalInterceptor offers; dispatch.InterceptLethal decides (D-7 order)
- the first taker's mutations replace AssetRemoved; a second AssetHPDelta sets HP
- Attack's sequenceTracker folds every mutation: HP, removed, changed sides
- self_damage runs automatic interceptors only (D-10)
- hooks.InterceptCollector: ConfirmSacrifice and the PaymentRestore wording
- adapter, overlay (Esc ignored), smoke and scripted collectors
- hooks.md, doc.go: six categories
```

### Phase 2 — Boltholes and False Front

The two defender interceptors with no price: `sheltered` and everything that reads
it, then the sacrifice.

#### Commit 2 — `feat: Boltholes`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
	ConditionNoAttack  ConditionKind = "no_attack"
)
```
**Replace with:**
```go
	ConditionNoAttack  ConditionKind = "no_attack"
	ConditionSheltered ConditionKind = "sheltered"
)
```

**Find:**
```go
// Active reports whether the asset can act or use its special ability now: an
// asset is not usable until the turn after purchase, while unmaintained, once
// destroyed, or while disabled.
func (asset *Asset) Active() bool {
	return asset.Targetable() && !asset.HasCondition(ConditionDisabled)
}

// Targetable reports whether a rival may pick the asset as a defender. It is
// Active without the conditions that stop use only: a disabled asset is useless,
// not untouchable (D-38).
func (asset *Asset) Targetable() bool {
	return asset.Ready && asset.Maintained && asset.CurrentHP > 0
}
```
**Replace with:**
```go
// Active reports whether the asset can act or use its special ability now: an
// asset is not usable until the turn after purchase, while unmaintained, once
// destroyed, while sheltered, or while disabled.
func (asset *Asset) Active() bool {
	return asset.Targetable() && !asset.HasCondition(ConditionDisabled)
}

// Targetable reports whether a rival may pick the asset as a defender. It is
// Active without the conditions that stop use only: a disabled asset is useless,
// not untouchable (D-38). A sheltered asset is both.
func (asset *Asset) Targetable() bool {
	return asset.Ready && asset.Maintained && asset.CurrentHP > 0 && !asset.HasCondition(ConditionSheltered)
}
```

##### Task 2 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
```
**Replace with:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
	SpecialShelterAsset   SpecialEffectType = "shelter_asset"
```

**Find:**
```go
	TriggerOwnDefenseWin          EffectTrigger = "own_defense_win"
```
**Replace with:**
```go
	TriggerOwnDefenseWin          EffectTrigger = "own_defense_win"
	TriggerLethalDamage           EffectTrigger = "lethal_damage"
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find** `domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin,` **replace with**
`domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin, domain.TriggerLethalDamage,`.

**Find:**
```go
		domain.SpecialCounterstrike,
	}
```
**Replace with:**
```go
		domain.SpecialCounterstrike, domain.SpecialShelterAsset,
	}
```

**Find:**
```go
// convertEffectParams fills the parameter pointer for definition.Effect. An effect
// that takes parameters must carry its block; the four that take none
// (ignore_coin_loss, apply_stealth, clear_stealth, remove_tag_dice) fall through.
```
**Replace with:**
```go
// convertEffectParams fills the parameter pointer for definition.Effect. An effect
// that takes parameters must carry its block; an effect that takes none falls
// through.
```

##### Task 4 — `internal/faction/engine/effect/effects/lethal.go` (new) — full contents:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// hpTo returns the AssetHPDelta that moves ctx's target from its HP after the hit
// to hp, or nil when it is there already. A Cat 6 save keeps the hit in history
// and records the save beside it, as a Bank records the loss it absorbs (Effort 2
// Decision 2).
func hpTo(ctx hooks.LethalContext, hp int, cause domain.SpecialEffectType, causedByFactionID string) []domain.Mutation {
	if ctx.HPAfter == hp {
		return nil
	}
	return []domain.Mutation{domain.AssetHPDelta{
		FactionID:         ctx.Target.OwnerID,
		AssetID:           ctx.Target.ID,
		Delta:             hp - ctx.HPAfter,
		Cause:             string(cause),
		CausedByFactionID: causedByFactionID,
	}}
}

// guardsTarget reports whether self, a defending interceptor, covers ctx's target:
// self is active, and the target is another asset of self's faction on self's
// world, of one of assetTypes (any type when empty).
func guardsTarget(self *domain.Asset, ctx hooks.LethalContext, assetTypes []domain.AssetType) bool {
	if !active(self) || ctx.Target.ID == self.ID || ctx.Target.OwnerID != self.OwnerID {
		return false
	}
	if self.Location.WorldID == "" || ctx.Target.Location.WorldID != self.Location.WorldID {
		return false
	}
	return len(assetTypes) == 0 || (ctx.TargetDef != nil && slices.Contains(assetTypes, ctx.TargetDef.Type))
}
```

##### Task 5 — `internal/faction/engine/effect/effects/shelter_asset.go` (new) — full contents:

```go
package effects

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// shelterAssetRegistrar is Boltholes: an own asset of a listed type on its world
// that damage would destroy is set to 0 HP and sheltered instead, until a repair
// brings it to full HP. If damage destroys the Boltholes first, its sheltered
// assets are destroyed with it. owner = "own" reads as "the struck asset is the
// declaring asset's faction's", scope = "world" as "on the declaring asset's world".
var shelterAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerShelterAsset,
}

func registerShelterAsset(ctx effect.RegisterContext) {
	scope := hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID)
	source := string(domain.SpecialShelterAsset) + ":" + ctx.Def.ID
	ctx.Registry.RegisterLethalInterceptor(scope, source, &shelterAssetInterceptor{
		factionID:  ctx.Faction.ID,
		assetID:    ctx.Asset.ID,
		assetTypes: ctx.Entry.Filters.AssetTypes,
	})
	ctx.Registry.RegisterMutationReactor(scope, source, &shelterLossReactor{assetID: ctx.Asset.ID})
}

type shelterAssetInterceptor struct {
	factionID  string
	assetID    string
	assetTypes []domain.AssetType
}

func (interceptor *shelterAssetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialShelterAsset
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptAutomatic,
		Mutations: append(hpTo(ctx, 0, cause, self.OwnerID), domain.AssetConditionApplied{
			FactionID:         ctx.Target.OwnerID,
			AssetID:           ctx.Target.ID,
			Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: self.ID},
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}

// shelterLossCauses are the removal causes that mean damage destroyed the asset.
var shelterLossCauses = []string{"attack", string(domain.SpecialCounterstrike), string(domain.SpecialSelfDamage)}

// shelterLossReactor destroys the assets its Boltholes shelters when damage
// destroys the Boltholes. A sale or a refit does not (Effort 2 Decision 13).
type shelterLossReactor struct {
	assetID string
}

func (reactor *shelterLossReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || removed.AssetID != reactor.assetID || !slices.Contains(shelterLossCauses, removed.Cause) {
			continue
		}
		// deliberate: no active check, because a disabled or unready Boltholes that is destroyed still takes its sheltered assets with it.
		sheltered := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: reactor.assetID}
		var extra []domain.Mutation
		// deliberate: every faction, not only the Boltholes' owner, because a capture moves the Boltholes but not the assets it shelters.
		for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
			for _, asset := range domain.SortedAssets(factionState.Factions[factionID]) {
				if !slices.ContainsFunc(asset.Conditions, sheltered.Is) {
					continue
				}
				if _, gone := batchHP(mutations, asset); gone {
					continue
				}
				// deliberate: copy the Boltholes removal's cause, not one of its own, because a combat kill of the Boltholes is also a combat kill of each asset it held (Effort 2 Decision 13).
				extra = append(extra, domain.AssetRemoved{FactionID: factionID, AssetID: asset.ID, Cause: removed.Cause, CausedByFactionID: removed.CausedByFactionID})
			}
		}
		return extra
	}
	return nil
}
```

##### Task 6 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialCounterstrike:  counterstrikeRegistrar,
```
**Replace with:**
```go
		domain.SpecialCounterstrike:  counterstrikeRegistrar,
		domain.SpecialShelterAsset:   shelterAssetRegistrar,
```

##### Task 7 — `internal/faction/engine/effect/effects/capture_target.go`

**Find:**
```go
		if _, removed := batchHP(mutations, target); removed {
			return nil
		}
```
**Replace with:**
```go
		// deliberate: fizzle when the batch leaves the target at 0 HP, not capture, because a later kill that Boltholes took shelters it, and a sheltered asset is untouchable (Effort 2 Decision 16).
		if hp, removed := batchHP(mutations, target); removed || hp <= 0 {
			return nil
		}
```

##### Task 8 — `internal/faction/engine/action/actions/repair_asset.go`

**Find:**
```go
				Cause:             "repair",
				CausedByFactionID: ra.factionID,
			})
		}
	}
```
**Replace with:**
```go
				Cause:             "repair",
				CausedByFactionID: ra.factionID,
			})
			if order.Asset.CurrentHP+totalHeal >= def.HP {
				mutations = append(mutations, shelterCleared(ra.factionID, order.Asset)...)
			}
		}
	}
```

Append to the file:
```go

// shelterCleared clears every sheltered condition on asset: a repair to full HP
// ends the shelter (Effort 2 Decision 12).
func shelterCleared(factionID string, asset *domain.Asset) []domain.Mutation {
	var cleared []domain.Mutation
	for _, condition := range asset.Conditions {
		if condition.Kind == domain.ConditionSheltered {
			cleared = append(cleared, domain.AssetConditionCleared{
				FactionID:         factionID,
				AssetID:           asset.ID,
				Condition:         condition,
				Cause:             "repair",
				CausedByFactionID: factionID,
			})
		}
	}
	return cleared
}
```

##### Task 9 — ability targets: `ability/stage.go`, `ability/levy_coin.go`

`stage.go` — **find:**
```go
// assetCandidates returns other factions' unstealthed assets on the acting asset's
// world, in faction-ID then asset-ID order, that pass the effect's rule. An asset
// in transit has no world and no candidates.
```
**Replace with:**
```go
// assetCandidates returns other factions' unstealthed, unsheltered assets on the
// acting asset's world, in faction-ID then asset-ID order, that pass the effect's
// rule. An asset in transit has no world and no candidates.
```
**Find:**
```go
			if candidate.Location.WorldID != world || candidate.Stealthy {
				continue
			}
```
**Replace with:**
```go
			// deliberate: skip a sheltered asset, not offer it, because Boltholes make it untouchable to every ability, not only to Attack (Effort 2 Decision 11).
			if candidate.Location.WorldID != world || candidate.Stealthy || candidate.HasCondition(domain.ConditionSheltered) {
				continue
			}
```
**Find** `			return asset.Location.WorldID == world && !asset.Stealthy`
**replace with** `			return asset.Location.WorldID == world && !asset.Stealthy && !asset.HasCondition(domain.ConditionSheltered)`.

`levy_coin.go` — **find:**
```go
// levyCoin is Monopoly: the target faction pays the owner 1 Coin. A faction that
// cannot pay loses one asset of its choice on the world instead, from all its
// assets there, stealthed ones too (Effort 3 Decision 11).
```
**Replace with:**
```go
// levyCoin is Monopoly: the target faction pays the owner 1 Coin. A faction that
// cannot pay loses one asset of its choice on the world instead, from all its
// assets there, stealthed ones too (Effort 3 Decision 11), but not a sheltered one
// (Effort 2 Decision 11).
```
**Find:**
```go
		if asset.Location.WorldID == ctx.Asset.Location.WorldID {
```
**Replace with:**
```go
		if asset.Location.WorldID == ctx.Asset.Location.WorldID && !asset.HasCondition(domain.ConditionSheltered) {
```

##### Task 10 — movement: `engine/orchestrator.go`, `engine/relocate/relocate.go`

`orchestrator.go`, in `eligibleMovableAssets` — **find:**
```go
		if def != nil && def.Speed > 0 {
			eligible = append(eligible, asset)
		}
```
**Replace with:**
```go
		// deliberate: a sheltered asset stays put, not moves, because Boltholes make it unusable until repaired (Effort 2 Decision 11).
		if def != nil && def.Speed > 0 && !asset.HasCondition(domain.ConditionSheltered) {
			eligible = append(eligible, asset)
		}
```
In `eligibleCargoForTransport` — **find**
`		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 {`
**replace with**
`		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 || asset.HasCondition(domain.ConditionSheltered) {`.

`relocate.go`, in `eligibleAssets` — **find:**
```go
		if asset.ID == request.Acting.ID || asset.CurrentOrder != nil {
```
**Replace with:**
```go
		// deliberate: skip a sheltered asset, because it stays put until repaired (Effort 2 Decision 11).
		if asset.ID == request.Acting.ID || asset.CurrentOrder != nil || asset.HasCondition(domain.ConditionSheltered) {
```

##### Task 11 — `internal/faction/engine/goal/goals/planetary_seizure.go`

**Find:**
```go
		case domain.AssetOwnerChanged:
			if m.ToFactionID == actingFaction.ID {
				cleared[m.AssetID] = true
			}
		}
```
**Replace with:**
```go
		case domain.AssetOwnerChanged:
			if m.ToFactionID == actingFaction.ID {
				cleared[m.AssetID] = true
			}
		case domain.AssetConditionApplied:
			if m.Condition.Kind == domain.ConditionSheltered {
				cleared[m.AssetID] = true
			}
		}
```
**Find:**
```go
// rivalPresentOn reports whether a faction other than seizerID holds an
// unstealthed asset on world that cleared does not name.
```
**Replace with:**
```go
// rivalPresentOn reports whether a faction other than seizerID holds an
// unstealthed, unsheltered asset on world that cleared does not name. A sheltered
// asset hides and cannot be attacked, so the Attack-only lock would hold forever
// (Effort 2 Decision 11).
```
**Find** `			if asset.Location.WorldID == world && !asset.Stealthy && !cleared[asset.ID] {`
**replace with**
`			if asset.Location.WorldID == world && !asset.Stealthy && !asset.HasCondition(domain.ConditionSheltered) && !cleared[asset.ID] {`.

##### Task 12 — `rulebooks/swn/assets/cunning_assets.toml`

After the Boltholes `description = ...` line (the one that ends
`the asset is destroyed with them."`) add:
```toml

[[assets.Boltholes.effects]]
trigger = "lethal_damage"
effect  = "shelter_asset"
[assets.Boltholes.effects.filters]
owner      = "own"
scope      = "world"
asset_type = ["Special Forces", "Military Unit"]
```

##### Task 13 — digest and TUI wording

`narrative/digest/types.go` — **find** `	Counterstruck          bool` (in `CrossEvent`)
**replace with:**
```go
	Counterstruck          bool
	Interceptions          []domain.SpecialEffectType
```

`narrative/digest/build.go` — after `findCrossTarget` add:
```go
// noteInterception adds a Cat 6 save to the cross event of the record's fight.
// ownerID is the faction whose asset the save touched. When it is the acting
// faction's own asset (a counter kill), the fight is with targetID.
func noteInterception(crossMap map[crossKey]*CrossEvent, record domain.EventRecord, ownerID, targetID string, effect domain.SpecialEffectType, factionState *state.FactionState) {
	rival := ownerID
	if ownerID == record.FactionID {
		rival = targetID
	}
	if rival == "" {
		return
	}
	ce := getOrCreateCross(crossMap, record.FactionID, rival, CrossAttack, factionState)
	if !slices.Contains(ce.Interceptions, effect) {
		ce.Interceptions = append(ce.Interceptions, effect)
	}
}
```
In the `asset_condition_applied` case — **find:**
```go
				if m.FactionID != record.FactionID {
					getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					kind := CrossAttack
```
**Replace with:**
```go
				if m.Cause == string(domain.SpecialShelterAsset) {
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialShelterAsset, factionState)
				} else if m.FactionID != record.FactionID {
					getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					kind := CrossAttack
```

`narrative/wire_templates.go` — after the `conditionClauses` map add:
```go
// interceptionClauses name each Cat 6 save without naming a side, because the
// saved asset can be the attacker's (a counter kill) or the defender's.
var interceptionClauses = map[domain.SpecialEffectType]string{
	domain.SpecialShelterAsset: "A stricken asset went to ground in its boltholes.",
}
```

`narrative/wire_renderer.go` — **find:**
```go
		result += conditionText(ce)
		if ce.Counterstruck {
```
**Replace with:**
```go
		result += conditionText(ce)
		result += interceptionText(ce)
		if ce.Counterstruck {
```
After `conditionText` add:
```go
// interceptionText renders one clause per Cat 6 save on a cross event, each with a
// leading space, or "" when it carries none.
func interceptionText(ce digest.CrossEvent) string {
	var text string
	for _, effect := range ce.Interceptions {
		if clause, ok := interceptionClauses[effect]; ok {
			text += " " + clause
		}
	}
	return text
}
```

`tui/views/turn/execution/format.go` — before `func formatMutation` add:
```go
// interceptionLabels name the Cat 6 save behind a mutation that is not damage.
var interceptionLabels = map[string]string{
	string(domain.SpecialShelterAsset): "sheltered",
}
```
**Find:**
```go
	case domain.AssetHPDelta:
		switch {
		case mut.Delta == 0:
```
**Replace with:**
```go
	case domain.AssetHPDelta:
		if label, ok := interceptionLabels[mut.Cause]; ok {
			return fmt.Sprintf("asset %s HP %+d (%s)", mut.AssetID, mut.Delta, label), true
		}
		switch {
		case mut.Delta == 0:
```

##### Task 14 — docs

`docs/architecture/engine/effect-mutation.md` (Mutation catalogue):
- **find** `` - `asset_hp_delta` — adjust an asset's HP `` **replace with**
  `` - `asset_hp_delta` — adjust an asset's HP. A Cat 6 save keeps the hit and sets the HP it saves at with a second delta, caused by the save's effect ``.
- **find** `keyed on kind and source asset (Saboteurs' \`no_ability\`)` **replace with**
  `keyed on kind and source asset (Saboteurs' \`no_ability\`, Marketers' \`disabled\`, Seditionists' \`no_attack\`, Boltholes' \`sheltered\`)`.

`docs/architecture/engine/actions.md` — in the Repair Asset row, **find**
`` `SelectRepairOrders` → `AssetHPDelta`, `CoinDelta` | `` **replace with**
`` `SelectRepairOrders` → `AssetHPDelta`, `CoinDelta` (`AssetConditionCleared` when a sheltered asset reaches full HP) | ``.

`docs/architecture/engine/ability-catalog.md` — **find** `Boltholes \`SWN-C5-003\``
**replace with** `Boltholes \`SWN-C5-003\` ✅ (\`shelter_asset\`)`.

##### Task 15 — tests

- `domain/asset_test.go`: `TestAsset_ActiveAndTargetable` adds a ready, maintained
  asset at 2 HP with `sheltered` → neither `Targetable()` nor `Active()`.
- `effects/shelter_asset_test.go` (new). Alpha holds Boltholes `bolt` and a Military
  Unit `mu` on Anchorage; the interceptor comes from
  `registry.LethalInterceptorsForAsset("alpha", "bolt")` after
  `effects.Registrars()[domain.SpecialShelterAsset].Register` with
  `asset_type = [Special Forces, Military Unit]`:
  - `mu`, `HPAfter: -5` → automatic offer, `Source` `bolt`, mutations
    `[AssetHPDelta{alpha, mu, +5, shelter_asset, alpha}, AssetConditionApplied{alpha, mu, {sheltered, bolt}, shelter_asset, alpha}]`.
  - `HPAfter: 0` → only the condition.
  - no offer: a rival's asset; `mu` on another world; a Facility; the Boltholes
    itself; an unready or disabled Boltholes; a Boltholes in transit (empty world).
  - loss reactor: `AssetRemoved{alpha, bolt, attack, beta}` with `mu` (alpha) and
    `g1` (gamma) sheltered by `bolt` → `[AssetRemoved{alpha, mu, attack, beta},
    AssetRemoved{gamma, g1, attack, beta}]`; the cause `sell` → nothing; `mu` already
    removed in the batch → only `g1`; a disabled Boltholes still drags.
- `effects/capture_target_test.go`: the target's batch HP is 0 (the hit and a
  `shelter_asset` delta) → the capture fizzles.
- `actions/repair_asset_test.go`: `TestRepairAsset_ClearsShelterAtFullHP` — a
  sheltered asset at 0 of 4 HP, Force 4, one heal →
  `[AssetHPDelta{+4}, AssetConditionCleared{{sheltered, bolt}, "repair"}, CoinDelta{-1}]`;
  at 0 of 8 HP → no clear.
- `ability/stage_test.go`: `TestAssetCandidates` gains a sheltered rival asset that is
  not a candidate. `ability/levy_coin_test.go`: a sheltered asset is not in the loss
  list; a faction whose only unstealthed asset on the world is sheltered is not a
  levy candidate.
- `engine/orchestrator_test.go`: `TestEligibleMovableAssets_SkipsSheltered` and
  `TestEligibleCargoForTransport_SkipsSheltered`. `relocate/relocate_test.go`: a
  sheltered asset is not eligible.
- `goals/lock_test.go`: Planetary Seizure phase 1 with the last rival asset sheltered
  → `CheckLock` advances to phase 2. `goals/progress_test.go`: a batch with
  `AssetConditionApplied{sheltered}` on the last rival asset → `phaseOneCleared`.
- `rulebook/rulebook_test.go`: a `TestLoad` subtest — Boltholes has one effect,
  `shelter_asset` on `lethal_damage`, asset types `[Special Forces, Military Unit]`.
- `engine/testharness/scenarios/boltholes_test.go` (new):
  `TestBoltholesShelterRepairAndLoss` — order `[alpha, beta]`; beta's Boltholes and a
  Military Unit on one world; alpha's attacker kills the Military Unit in cycle 1. After
  cycle 1: the unit is in beta at 0 HP with `sheltered`, and history has no
  `asset_removed`. Beta takes Repair Asset in cycle 2 and heals to full → the
  condition is gone. A second test, `TestBoltholesLossTakesShelteredAsset`: after the
  shelter, alpha kills the Boltholes → the unit is removed with cause `attack`.
- `digest/digest_test.go`: an `asset_condition_applied` with cause `shelter_asset`
  → `Interceptions == [shelter_asset]` and no `Conditions`.
  `narrative/wire_renderer_test.go`: a cross attack with that interception ends with
  its clause. `execution/format_test.go`: a `shelter_asset` `AssetHPDelta` of `+5` →
  `asset mu HP +5 (sheltered)`.

##### Commit message

```
feat: Boltholes

- shelter_asset: an own SF or MU on the world goes to 0 HP and sheltered
- sheltered: not Targetable, not Active, not an ability target, not moved
- Repair Asset clears it at full HP; a Boltholes destroyed by damage drags its assets
- Planetary Seizure does not count a sheltered rival; Treachery fizzles on one
- digest interception clause and TUI label
```

#### Commit 3 — `feat: False Front`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
```
**Replace with:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
	SpecialSacrificeSelf  SpecialEffectType = "sacrifice_self"
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find** `domain.SpecialShelterAsset,` **replace with**
`domain.SpecialShelterAsset, domain.SpecialSacrificeSelf,`.

##### Task 3 — `internal/faction/engine/effect/effects/sacrifice_self.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// sacrificeSelfRegistrar is False Front: when damage would destroy another own
// asset on its world, the owner may sacrifice the False Front to nullify the
// killing blow, and the struck asset keeps its HP from before that hit (Effort 2
// Decision 14). owner = "own" reads as "the struck asset is the declaring asset's
// faction's", scope = "world" as "on the declaring asset's world".
var sacrificeSelfRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
		},
	},
	Register: registerSacrificeSelf,
}

func registerSacrificeSelf(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialSacrificeSelf)+":"+ctx.Def.ID,
		&sacrificeSelfInterceptor{factionID: ctx.Faction.ID, assetID: ctx.Asset.ID},
	)
}

type sacrificeSelfInterceptor struct {
	factionID string
	assetID   string
}

func (interceptor *sacrificeSelfInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, nil) {
		return nil
	}
	cause := domain.SpecialSacrificeSelf
	// deliberate: the removal's cause is the sacrifice, not "attack", because the owner gave the False Front up and no one scores a kill (Effort 2 Decision 14).
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptSacrifice,
		Mutations: append(hpTo(ctx, ctx.HPAfter+ctx.Damage, cause, self.OwnerID), domain.AssetRemoved{
			FactionID:         self.OwnerID,
			AssetID:           self.ID,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
```

##### Task 4 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialShelterAsset:   shelterAssetRegistrar,
```
**Replace with:**
```go
		domain.SpecialShelterAsset:   shelterAssetRegistrar,
		domain.SpecialSacrificeSelf:  sacrificeSelfRegistrar,
```

##### Task 5 — `rulebooks/swn/assets/cunning_assets.toml`

After the False Front `description = ...` line (the one that ends
`to nullify the killing blow."`) add:
```toml

[[assets.FalseFront.effects]]
trigger = "lethal_damage"
effect  = "sacrifice_self"
[assets.FalseFront.effects.filters]
owner = "own"
scope = "world"
```

##### Task 6 — digest and TUI wording

`narrative/digest/build.go`, in the `asset_removed` switch — **find:**
```go
				case "refit":
					// already captured in refitRemovedAssetID; encoded as From on the paired AssetAdded
```
**Replace with:**
```go
				case string(domain.SpecialSacrificeSelf):
					ownerBeat := getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					ownerBeat.Losses = append(ownerBeat.Losses, AssetMove{
						AssetID:   m.AssetID,
						AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
						Cause:     m.Cause,
					})
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialSacrificeSelf, factionState)
				case "refit":
					// already captured in refitRemovedAssetID; encoded as From on the paired AssetAdded
```

`narrative/wire_templates.go` — **find**
`var interceptionClauses = map[domain.SpecialEffectType]string{` **replace with:**
```go
var interceptionClauses = map[domain.SpecialEffectType]string{
	domain.SpecialSacrificeSelf: "A false front fell in place of the asset it covered.",
```

`tui/views/turn/execution/format.go` — **find**
`var interceptionLabels = map[string]string{` **replace with:**
```go
var interceptionLabels = map[string]string{
	string(domain.SpecialSacrificeSelf): "false front",
```
**Find:**
```go
	case domain.AssetRemoved:
		return fmt.Sprintf("-asset %s", mut.AssetID), true
```
**Replace with:**
```go
	case domain.AssetRemoved:
		if label, ok := interceptionLabels[mut.Cause]; ok {
			return fmt.Sprintf("-asset %s (%s)", mut.AssetID, label), true
		}
		return fmt.Sprintf("-asset %s", mut.AssetID), true
```

##### Task 7 — `docs/architecture/engine/ability-catalog.md`

**Find** `False Front \`SWN-C1-003\`,` **replace with**
`False Front \`SWN-C1-003\` ✅ (\`sacrifice_self\`),`.

##### Task 8 — tests

- `effects/sacrifice_self_test.go` (new). Alpha holds False Front `ff` and `x` on
  Anchorage; the interceptor from the registry:
  - `x`, `HPAfter: -3`, `Damage: 6` → `InterceptSacrifice`, `Source` `ff`, mutations
    `[AssetHPDelta{alpha, x, +6, sacrifice_self, alpha}, AssetRemoved{alpha, ff, sacrifice_self, alpha}]`
    (x back at 3 HP).
  - no offer: the False Front itself; a rival's asset; `x` on another world; an
    unready False Front.
- `actions/attack_test.go`: `TestAttack_FalseFrontSpentOnce` — the real registrar,
  `ConfirmSacrifice` answers yes; two attackers each make a lethal hit on a different
  beta asset on the world → the first is saved, the second prompt is not asked (the
  batch removes the False Front), and the second asset is removed.
- `tags/scavengers_test.go`: an `AssetRemoved` with cause `sacrifice_self` pays
  nothing.
- `rulebook/rulebook_test.go`: a `TestLoad` subtest — False Front has one effect,
  `sacrifice_self` on `lethal_damage`.
- `digest/digest_test.go`: an `asset_removed` with cause `sacrifice_self` → a loss on
  the owner's beat and `Interceptions == [sacrifice_self]`.
  `execution/format_test.go`: `-asset ff (false front)`.

##### Commit message

```
feat: False Front

- sacrifice_self: the owner may give up the False Front to undo a killing blow
- the saved asset keeps its HP from before the hit; the sacrifice is not a kill
- digest clause and loss; TUI label
```

### Phase 3 — Medical Center and Hostile Takeover

The priced elective save with its repair discount, then the attacker's capture.

#### Commit 4 — `feat: Medical Center`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
```
**Replace with:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
	SpecialRestoreAsset   SpecialEffectType = "restore_asset"
	SpecialDiscountRepair SpecialEffectType = "discount_repair"
```

**Find:**
```go
	TriggerOwnDefenseWin          EffectTrigger = "own_defense_win"
```
**Replace with:**
```go
	TriggerOwnDefenseWin          EffectTrigger = "own_defense_win"
	TriggerRepairCostQuery        EffectTrigger = "repair_cost_query"
```

**Find:**
```go
type CaptureTargetParams struct {
	Coin DiceRoll
}
```
**Replace with:**
```go
type CaptureTargetParams struct {
	Coin DiceRoll
}

type DiscountRepairParams struct {
	Amount int
}
```

**Find:**
```go
	CaptureTarget  *CaptureTargetParams
}
```
**Replace with:**
```go
	CaptureTarget  *CaptureTargetParams
	DiscountRepair *DiscountRepairParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

In `effectRecord`, **find** `	CaptureTarget  *amountRecord         \`toml:"capture_target"\``
**replace with:**
```go
	CaptureTarget  *amountRecord         `toml:"capture_target"`
	DiscountRepair *discountRepairRecord `toml:"discount_repair"`
```
After `type grantMovementRecord struct { ... }` add:
```go
type discountRepairRecord struct {
	Amount int `toml:"amount"`
}
```
**Find** `domain.TriggerLethalDamage,` **replace with**
`domain.TriggerLethalDamage, domain.TriggerRepairCostQuery,`.
**Find** `domain.SpecialSacrificeSelf,` **replace with:**
```go
domain.SpecialSacrificeSelf,
		domain.SpecialRestoreAsset, domain.SpecialDiscountRepair,
```
In `presentParamBlocks`, **find:**
```go
	if record.CaptureTarget != nil {
		present = append(present, domain.SpecialCaptureTarget)
	}
```
**Replace with:**
```go
	if record.CaptureTarget != nil {
		present = append(present, domain.SpecialCaptureTarget)
	}
	if record.DiscountRepair != nil {
		present = append(present, domain.SpecialDiscountRepair)
	}
```
In `convertEffectParams`, **find:**
```go
		definition.CaptureTarget = &domain.CaptureTargetParams{Coin: coin}
	}
	return nil
}
```
**Replace with:**
```go
		definition.CaptureTarget = &domain.CaptureTargetParams{Coin: coin}
	case domain.SpecialDiscountRepair:
		if record.DiscountRepair == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.DiscountRepair.Amount <= 0 {
			return fmt.Errorf("discount_repair amount must be > 0, got %d", record.DiscountRepair.Amount)
		}
		definition.DiscountRepair = &domain.DiscountRepairParams{Amount: record.DiscountRepair.Amount}
	}
	return nil
}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`

Append:
```go

// RepairCostQuery is the state a repair-cost chain carries from hook to hook.
// Applied holds the effect names whose no_stack gate has already fired.
type RepairCostQuery struct {
	Cost    int
	Applied map[string]bool
}

// RepairCostModifier adjusts what Repair Asset charges to repair one asset
// (Cat 4). repairer is the faction that repairs; asset and def are the asset
// repaired.
type RepairCostModifier interface {
	ModifyRepairCost(repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, query *RepairCostQuery)
}
```

##### Task 4 — `internal/faction/engine/hooks/registry.go`

**Find:**
```go
type RegisteredLethalInterceptor struct {
```
**Replace with:**
```go
type RegisteredRepairCostModifier struct {
	Scope  Scope
	Source string
	Hook   RepairCostModifier
}

type RegisteredLethalInterceptor struct {
```
**Find** `	lethalInterceptors       map[Scope][]RegisteredLethalInterceptor` **replace with:**
```go
	lethalInterceptors       map[Scope][]RegisteredLethalInterceptor
	repairCostModifiers      map[Scope][]RegisteredRepairCostModifier
```
**Find** `	registry.lethalInterceptors = make(map[Scope][]RegisteredLethalInterceptor)`
**replace with:**
```go
	registry.lethalInterceptors = make(map[Scope][]RegisteredLethalInterceptor)
	registry.repairCostModifiers = make(map[Scope][]RegisteredRepairCostModifier)
```
**Find** `	delete(registry.lethalInterceptors, scope)` **replace with:**
```go
	delete(registry.lethalInterceptors, scope)
	delete(registry.repairCostModifiers, scope)
```
Append:
```go

// RegisterRepairCostModifier registers a Cat 4 repair-cost hook under the given scope.
func (registry *Registry) RegisterRepairCostModifier(scope Scope, source string, hook RepairCostModifier) {
	registry.repairCostModifiers[scope] = append(registry.repairCostModifiers[scope], RegisteredRepairCostModifier{Scope: scope, Source: source, Hook: hook})
}

// RepairCostModifiersFor returns all Cat 4 repair-cost hooks matching the given faction and asset.
func (registry *Registry) RepairCostModifiersFor(factionID, assetInstanceID string) []RegisteredRepairCostModifier {
	var result []RegisteredRepairCostModifier
	result = append(result, registry.repairCostModifiers[GlobalScope()]...)
	result = append(result, registry.repairCostModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.repairCostModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// RepairCostModifiersForAsset returns the Cat 4 repair-cost hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) RepairCostModifiersForAsset(factionID, assetInstanceID string) []RegisteredRepairCostModifier {
	return registry.repairCostModifiers[AssetScope(factionID, assetInstanceID)]
}
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules.go`

After `GrantedMovementAbilities` add:
```go
// ResolveRepairCost runs the repairer's own RepairCostModifiers, then those
// registered against each asset it holds, in asset ID order, chaining the query.
// A repair discount serves its owner only, so a rival's assets are never asked.
// The cost never drops below 0.
func ResolveRepairCost(registry *hooks.Registry, repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, baseCost int) int {
	if registry == nil {
		return baseCost
	}
	query := &hooks.RepairCostQuery{Cost: baseCost, Applied: map[string]bool{}}
	for _, registered := range registry.RepairCostModifiersFor(repairer.ID, "") {
		registered.Hook.ModifyRepairCost(repairer, asset, def, query)
	}
	for _, held := range domain.SortedAssets(repairer) {
		for _, registered := range registry.RepairCostModifiersForAsset(repairer.ID, held.ID) {
			registered.Hook.ModifyRepairCost(repairer, asset, def, query)
		}
	}
	return max(0, query.Cost)
}
```

##### Task 6 — `internal/faction/engine/hooks/doc.go`

**Find:**
```go
// WorldTechLevelModifier, AttackEligibilityRule, AssetMovementGranter) rather
```
**Replace with:**
```go
// WorldTechLevelModifier, AttackEligibilityRule, AssetMovementGranter,
// RepairCostModifier) rather
```

##### Task 7 — `internal/faction/engine/effect/effects/restore_asset.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// restoreAssetRegistrar is Medical Center's save: when damage would destroy an own
// asset of a listed type on its world, the owner may pay the saved asset's half
// price to restore it at 1 HP. once_per = "asset" makes it once between the owner's
// turns for each Medical Center (Effort 2 Decision 15). owner = "own" reads as "the
// struck asset is the declaring asset's faction's", scope = "world" as "on the
// declaring asset's world".
var restoreAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerRestoreAsset,
}

func registerRestoreAsset(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialRestoreAsset)+":"+ctx.Def.ID,
		&restoreAssetInterceptor{
			factionID:  ctx.Faction.ID,
			assetID:    ctx.Asset.ID,
			defID:      ctx.Def.ID,
			assetTypes: ctx.Entry.Filters.AssetTypes,
			oncePer:    ctx.Entry.Gates.OncePer,
		},
	)
}

type restoreAssetInterceptor struct {
	factionID  string
	assetID    string
	defID      string
	assetTypes []domain.AssetType
	oncePer    domain.OncePerKey
}

func (interceptor *restoreAssetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, interceptor.assetTypes) || ctx.TargetDef == nil {
		return nil
	}
	cause := domain.SpecialRestoreAsset
	// deliberate: the saved asset's half price, not the Medical Center's, because "its purchase cost" names the destroyed asset (Effort 2 Decision 15).
	price := ctx.TargetDef.HalfPrice()
	budget := ""
	if interceptor.oncePer != "" {
		budget = budgetKey(interceptor.oncePer, cause, domain.TriggerLethalDamage, interceptor.defID, self.ID, ctx.Target.OwnerID)
	}
	return &hooks.Interception{
		Source:    self,
		Ask:       hooks.InterceptRestore,
		Price:     price,
		BudgetKey: budget,
		Mutations: append(hpTo(ctx, 1, cause, self.OwnerID), domain.CoinDelta{
			FactionID:         self.OwnerID,
			Delta:             -price,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
```

##### Task 8 — `internal/faction/engine/effect/effects/discount_repair.go` (new) — full contents:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// discountRepairRegistrar is Medical Center's repair clause: the owner's Repair
// Asset costs amount less for each asset of a listed type repaired on the declaring
// asset's world. owner = "own" reads as "the repairing faction is the owner", scope
// = "world" as "the repaired asset is on the declaring asset's world". no_stack
// keeps a second Medical Center there from adding to it (Effort 2 Decision 15).
var discountRepairRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerRepairCostQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateNoStack},
		},
	},
	Register: registerDiscountRepair,
}

func registerDiscountRepair(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRepairCostModifier(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialDiscountRepair)+":"+ctx.Def.ID,
		&discountRepairModifier{
			owner:      ctx.Faction,
			assetID:    ctx.Asset.ID,
			noStack:    ctx.Entry.Gates.NoStack,
			amount:     ctx.Entry.DiscountRepair.Amount,
			assetTypes: ctx.Entry.Filters.AssetTypes,
		},
	)
}

// discountRepairModifier holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type discountRepairModifier struct {
	owner      *domain.Faction
	assetID    string
	noStack    bool
	amount     int
	assetTypes []domain.AssetType
}

func (modifier *discountRepairModifier) ModifyRepairCost(repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, query *hooks.RepairCostQuery) {
	self := modifier.owner.Assets[modifier.assetID]
	if !active(self) || repairer.ID != modifier.owner.ID {
		return
	}
	if self.Location.WorldID == "" || asset.Location.WorldID != self.Location.WorldID {
		return
	}
	if len(modifier.assetTypes) > 0 && !slices.Contains(modifier.assetTypes, def.Type) {
		return
	}
	key := string(domain.SpecialDiscountRepair)
	if modifier.noStack && query.Applied[key] {
		return
	}
	query.Cost -= modifier.amount
	query.Applied[key] = true
}
```

##### Task 9 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialSacrificeSelf:  sacrificeSelfRegistrar,
```
**Replace with:**
```go
		domain.SpecialSacrificeSelf:  sacrificeSelfRegistrar,
		domain.SpecialRestoreAsset:   restoreAssetRegistrar,
		domain.SpecialDiscountRepair: discountRepairRegistrar,
```

##### Task 10 — Repair Asset: `repair_asset.go`, `register.go`, `repair_asset_test.go`

`repair_asset.go` — add the imports `.../engine/hooks` and `.../engine/hooks/dispatch`.

**Find:**
```go
// RepairAsset restores HP to one or more of the faction's damaged assets.
// Cost escalates per heal on the same asset; resets to 1 Coin for each new asset.
type RepairAsset struct {
	collector    action.Collector
	factionID    string
	repairOrders []action.RepairOrder
	mutations    []domain.Mutation
}

func NewRepairAsset(collector action.Collector) *RepairAsset {
	return &RepairAsset{collector: collector}
}
```
**Replace with:**
```go
// RepairAsset restores HP to one or more of the faction's damaged assets.
// Cost escalates per heal on the same asset; resets to 1 Coin for each new asset.
// A RepairCostModifier can lower each asset's price (Medical Center).
type RepairAsset struct {
	collector    action.Collector
	registry     *hooks.Registry
	factionID    string
	repairOrders []action.RepairOrder
	mutations    []domain.Mutation
}

func NewRepairAsset(collector action.Collector, registry *hooks.Registry) *RepairAsset {
	return &RepairAsset{collector: collector, registry: registry}
}
```
**Find:**
```go
		for healIndex := range order.HealCount {
			healAmount := min(attributeScore, missing-totalHeal)
			if healAmount <= 0 {
				break
			}
			totalCost += healIndex + 1 // 1-indexed escalating cost
			totalHeal += healAmount
		}

		if totalHeal > 0 {
```
**Replace with:**
```go
		assetCost := 0
		for healIndex := range order.HealCount {
			healAmount := min(attributeScore, missing-totalHeal)
			if healAmount <= 0 {
				break
			}
			assetCost += healIndex + 1 // 1-indexed escalating cost
			totalHeal += healAmount
		}

		if totalHeal > 0 {
			totalCost += dispatch.ResolveRepairCost(ra.registry, faction, order.Asset, def, assetCost)
```

`register.go` — **find**
`	ae.Register(func(c action.Collector) action.Action { return NewRepairAsset(c) })`
**replace with**
`	ae.Register(func(c action.Collector) action.Action { return NewRepairAsset(c, hookRegistry) })`.

`repair_asset_test.go` — every `NewRepairAsset(mocks.NewMockCollector(ctrl))` becomes
`NewRepairAsset(mocks.NewMockCollector(ctrl), nil)`, and `NewRepairAsset(collector)`
becomes `NewRepairAsset(collector, nil)`.

##### Task 11 — `rulebooks/swn/assets/wealth_assets.toml`

**Find:**
```toml
description = "Fully-equipped medical facilities. If a Special Forces or Military Unit asset on this planet is destroyed, the faction may immediately pay half its purchase cost to restore it at 1 HP. Any Repair Asset action on this planet costs 1 fewer Coin for Special Forces and Military Units."
```
**Replace with:**
```toml
description = "Fully-equipped medical facilities. Once between turns, if a Special Forces or Military Unit asset on this planet is destroyed, the faction may immediately pay half its purchase cost to restore it at 1 HP. Any Repair Asset action on this planet costs 1 fewer Coin for Special Forces and Military Units."

[[assets.MedicalCenter.effects]]
trigger = "lethal_damage"
effect  = "restore_asset"
[assets.MedicalCenter.effects.filters]
owner      = "own"
scope      = "world"
asset_type = ["Special Forces", "Military Unit"]
[assets.MedicalCenter.effects.gates]
once_per = "asset"

[[assets.MedicalCenter.effects]]
trigger = "repair_cost_query"
effect  = "discount_repair"
[assets.MedicalCenter.effects.filters]
owner      = "own"
scope      = "world"
asset_type = ["Special Forces", "Military Unit"]
[assets.MedicalCenter.effects.gates]
no_stack = true
[assets.MedicalCenter.effects.discount_repair]
amount = 1
```

##### Task 12 — digest and TUI wording

`narrative/digest/build.go`, in the `coin_delta` case — **find:**
```go
				if m.FactionID != record.FactionID && m.Cause == "ability" {
```
**Replace with:**
```go
				if m.Cause == string(domain.SpecialRestoreAsset) {
					payerBeat := getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					payerBeat.CoinDelta += m.Delta
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialRestoreAsset, factionState)
				} else if m.FactionID != record.FactionID && m.Cause == "ability" {
```

`narrative/wire_templates.go` — **find**
`var interceptionClauses = map[domain.SpecialEffectType]string{` **replace with:**
```go
var interceptionClauses = map[domain.SpecialEffectType]string{
	domain.SpecialRestoreAsset: "Medics pulled a stricken asset back from the brink.",
```

`tui/views/turn/execution/format.go` — **find**
`var interceptionLabels = map[string]string{` **replace with:**
```go
var interceptionLabels = map[string]string{
	string(domain.SpecialRestoreAsset): "restored",
```

##### Task 13 — docs

`docs/architecture/engine/hooks.md`:
- **find** `` `AssetMovementGranter` — each with its own signature `` **replace with**
  `` `AssetMovementGranter`, `RepairCostModifier` — each with its own signature ``.
- Before the bullet that starts `- **\`InterceptLethal\`** is Cat 6.` add:
  ```
  - **`ResolveRepairCost`** is the fifth Cat 4 resolver, asked by Repair Asset once
    per asset repaired. It walks the repairer's global and faction hooks, then each
    of its own assets in ID order, chaining a `RepairCostQuery` (whose `Applied`
    holds the `no_stack` marks), and never returns less than 0. A discount serves
    its owner only.
  ```
- In Key Decisions, **find**
  `(asset cost, maintenance, world TL, attack eligibility, movement grants)`
  **replace with**
  `(asset cost, maintenance, world TL, attack eligibility, movement grants, repair cost)`,
  and **find** `Five small interfaces keep every call site honest.` **replace with**
  `Six small interfaces keep every call site honest.`

`docs/architecture/engine/actions.md` — in the Repair Asset row, **find**
`| Heal damaged assets; cost escalates per heal on the same asset |` **replace with**
`| Heal damaged assets; cost escalates per heal on the same asset, and a \`RepairCostModifier\` can lower each asset's price |`.

`docs/architecture/engine/ability-catalog.md` — **find**
`Medical Center \`SWN-W4-003\` |` **replace with**
`Medical Center \`SWN-W4-003\` ✅ (\`restore_asset\`) |`. After the row that starts
`| **purchase surcharge on rivals**` add:
```
| **repair discount** (cheaper Repair Asset on its world) | Cat 4 `RepairCostModifier` → `discount_repair` | Medical Center `SWN-W4-003` ✅ |
```

##### Task 14 — tests

- `effects/restore_asset_test.go` (new). Alpha (`Coin: 10`) holds Medical Center `mc`
  and a Military Unit `mu` (cost 6) on Anchorage, `once_per = "asset"`:
  - `mu`, `HPAfter: -2` → `InterceptRestore`, `Price: 3`, a budget key that names
    `mc`, mutations `[AssetHPDelta{alpha, mu, +3, restore_asset, alpha}, CoinDelta{alpha, -3, restore_asset, alpha}]`.
  - a cost-1 asset → `Price: 1`.
  - no offer: a Facility; a rival's asset; another world; an unready Medical Center.
- `effects/discount_repair_test.go` (new): the modifier takes 1 off for alpha's own
  Military Unit on its world; nothing for a rival repairer, another world, a Facility,
  or an inactive Medical Center; with `no_stack`, a second Medical Center adds nothing.
- `hooks/dispatch/rules_test.go`: `TestResolveRepairCost` — nil registry → base; a
  fixture modifier on the repairer's own asset lowers the cost; one on a rival's asset
  is not asked; the result never drops below 0.
- `hooks/registry_test.go`: `TestClearAsset_RemovesOnlyThatAssetsBucket` also
  checks the repair-cost bucket.
- `actions/repair_asset_test.go`: `TestRepairAsset_DiscountLowersEachAssetsPrice` —
  a registry with a fixture modifier that takes 1 off for `a1` only; orders for `a1`
  (one heal) and `a2` (one heal) → one `CoinDelta{-1}` (0 + 1).
- `engine/testharness/scenarios/medical_center_test.go` (new), order `[alpha, beta]`,
  beta's Medical Center and two Military Units on one world, alpha with two attackers:
  - `TestMedicalCenterRestoresOncePerTurn` — both units take a lethal hit;
    `ConfirmPaymentFn` answers yes and records each offer. One offer
    (`PaymentRestore`, the unit's half price); the first unit is in beta at 1 HP;
    the second is removed.
  - `TestTwoMedicalCentersCountTheBatchCoin` — two Medical Centers, beta's Coin pays
    one restore only → one offer; the second unit is removed with no prompt.
- `rulebook/rulebook_test.go`: a `TestLoad` subtest — Medical Center has
  `restore_asset` (`once_per` asset) and `discount_repair` (amount 1, `no_stack`).
  `effects_test.go`: `TestLoadEffects_LoudFailure` gains `discount_repair` with no
  block and with `amount = 0`.
- `digest/digest_test.go`: a `coin_delta` with cause `restore_asset` lands on the
  payer's beat and adds `restore_asset` to the fight's interceptions.

##### Commit message

```
feat: Medical Center

- restore_asset: pay the saved asset's half price to restore it at 1 HP
- once between turns per Medical Center; the price counts the batch's Coin
- RepairCostModifier (Cat 4) and ResolveRepairCost; Repair Asset gains the registry
- discount_repair: 1 Coin off per own SF or MU repaired on the world, no stack
- digest clause and the payer's Coin; TUI label
```

#### Commit 5 — `feat: Hostile Takeover`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
```
**Replace with:**
```go
	SpecialCounterstrike  SpecialEffectType = "counterstrike"
	SpecialSeizeTarget    SpecialEffectType = "seize_target"
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find** `domain.SpecialRestoreAsset, domain.SpecialDiscountRepair,` **replace with**
`domain.SpecialRestoreAsset, domain.SpecialDiscountRepair, domain.SpecialSeizeTarget,`.

##### Task 3 — `internal/faction/engine/effect/effects/seize_target.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// seizeTargetRegistrar is Hostile Takeover: when damage it deals, attack or counter,
// would destroy a rival's asset, the asset is reduced to 1 HP and joins the Hostile
// Takeover's faction instead (Effort 2 Decision 16). owner = "rival" reads as "the
// struck asset is another faction's", scope = "self" as "the declaring asset dealt
// the blow".
var seizeTargetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerSeizeTarget,
}

func registerSeizeTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialSeizeTarget)+":"+ctx.Def.ID,
		&seizeTargetInterceptor{factionID: ctx.Faction.ID, assetID: ctx.Asset.ID},
	)
}

type seizeTargetInterceptor struct {
	factionID string
	assetID   string
}

func (interceptor *seizeTargetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !active(self) || ctx.Striker == nil || ctx.Striker.ID != self.ID || ctx.Target.OwnerID == self.OwnerID {
		return nil
	}
	cause := domain.SpecialSeizeTarget
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptAutomatic,
		Mutations: append(hpTo(ctx, 1, cause, self.OwnerID), domain.AssetOwnerChanged{
			FactionID:         ctx.Target.OwnerID,
			AssetID:           ctx.Target.ID,
			ToFactionID:       self.OwnerID,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
```

##### Task 4 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialDiscountRepair: discountRepairRegistrar,
```
**Replace with:**
```go
		domain.SpecialDiscountRepair: discountRepairRegistrar,
		domain.SpecialSeizeTarget:    seizeTargetRegistrar,
```

##### Task 5 — `internal/faction/engine/effect/effects/capture_target.go`

**Find:**
```go
		if hp, removed := batchHP(mutations, target); removed || hp <= 0 {
			return nil
		}
```
**Replace with:**
```go
		if hp, removed := batchHP(mutations, target); removed || hp <= 0 {
			return nil
		}
		// deliberate: fizzle when the batch already moves the target, because Hostile Takeover's capture in the action comes first and a second owner change would miss (Effort 1 Decision Record 11).
		if ownerChanged(mutations, target.ID) {
			return nil
		}
```
After `capturesOnHit` add:
```go
// ownerChanged reports whether mutations move assetID to another faction.
func ownerChanged(mutations []domain.Mutation, assetID string) bool {
	return slices.ContainsFunc(mutations, func(mutation domain.Mutation) bool {
		changed, ok := mutation.(domain.AssetOwnerChanged)
		return ok && changed.AssetID == assetID
	})
}
```

##### Task 6 — `rulebooks/swn/assets/wealth_assets.toml`

**Find:**
```toml
[assets.HostileTakeover.attack]
attacker_stat = "Wealth"
defender_stat = "Wealth"
damage = "2d10"
```
**Replace with:**
```toml
[assets.HostileTakeover.attack]
attacker_stat = "Wealth"
defender_stat = "Wealth"
damage = "2d10"

[[assets.HostileTakeover.effects]]
trigger = "lethal_damage"
effect  = "seize_target"
[assets.HostileTakeover.effects.filters]
owner = "rival"
scope = "self"
```

##### Task 7 — digest and TUI wording

`narrative/digest/build.go` — **find:**
```go
				winner.Acquisitions = append(winner.Acquisitions, AssetMove{AssetID: m.AssetID, AssetName: name, Cause: "captured"})
```
**Replace with:**
```go
				winner.Acquisitions = append(winner.Acquisitions, AssetMove{AssetID: m.AssetID, AssetName: name, Cause: "captured"})
				if m.Cause == string(domain.SpecialSeizeTarget) {
					noteInterception(crossMap, record, m.FactionID, targetID, domain.SpecialSeizeTarget, factionState)
				}
```

`narrative/wire_templates.go` — **find**
`var interceptionClauses = map[domain.SpecialEffectType]string{` **replace with:**
```go
var interceptionClauses = map[domain.SpecialEffectType]string{
	domain.SpecialSeizeTarget: "A stricken asset changed hands in a hostile takeover.",
```

`tui/views/turn/execution/format.go` — **find**
`var interceptionLabels = map[string]string{` **replace with:**
```go
var interceptionLabels = map[string]string{
	string(domain.SpecialSeizeTarget): "seized",
```

##### Task 8 — `docs/architecture/engine/ability-catalog.md`

**Find** `Hostile Takeover \`SWN-W7-002\`, Treachery` **replace with**
`Hostile Takeover \`SWN-W7-002\` ✅ (\`seize_target\`), Treachery`.

##### Task 9 — tests

- `effects/seize_target_test.go` (new). Alpha holds Hostile Takeover `ht`; beta holds
  `x`:
  - striker `ht`, target `x`, `HPAfter: -4` → automatic offer, mutations
    `[AssetHPDelta{beta, x, +5, seize_target, alpha}, AssetOwnerChanged{beta, x → alpha, seize_target, alpha}]`.
    `LethalContext` carries no side, so the counter case is the attack test below
    (Decision Record 3).
  - no offer: an own target; another striker; a nil striker (self-damage); an unready
    or disabled Hostile Takeover.
- `effects/capture_target_test.go`: a batch that already holds
  `AssetOwnerChanged{x}` → the capture fizzles, and the Treachery is not spent.
- `actions/attack_test.go`, with the real registrars through an effects-backed
  registry: `TestAttack_BoltholesBeforeHostileTakeover` — beta's Boltholes on the
  world and an HT kill on beta's Military Unit → sheltered, not seized (D-7).
  `TestAttack_HostileTakeoverCounterSeizesAttacker` — beta's Hostile Takeover
  counters alpha's attacker to death → `AssetOwnerChanged{alpha, a1 → beta}`, no
  `AssetRemoved{a1}`.
- `engine/testharness/scenarios/hostile_takeover_test.go` (new):
  `TestHostileTakeoverSeizesAndPaysTheNewOwner` — order `[alpha, beta]`; alpha's
  Hostile Takeover kills beta's Party Machine (`turn_start` income) on one world.
  After the cycle: the Party Machine is in alpha at 1 HP; history has no
  `asset_removed`. In the next cycle its `coin_gain` pays alpha, not beta (Decision
  Record 4).
- `rulebook/rulebook_test.go`: a `TestLoad` subtest — Hostile Takeover has one effect,
  `seize_target` on `lethal_damage`, owner `rival`, scope `self`.
- `digest/digest_test.go`: an `asset_owner_changed` with cause `seize_target` → the
  loss and acquisition as before, and `Interceptions == [seize_target]`.
  `execution/format_test.go`: a `seize_target` delta → `(seized)`.

##### Commit message

```
feat: Hostile Takeover

- seize_target: a kill by its attack or counter takes the asset at 1 HP instead
- D-7: the defender's protections run first
- Treachery fizzles on a target the batch already moved
- digest clause; TUI label
```
