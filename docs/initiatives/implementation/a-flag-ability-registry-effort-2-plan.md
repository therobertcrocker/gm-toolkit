# A-flag Ability Registry (A.001.2) — Effort 2 Plan: Roll-table effects

## Context / Goal

Build the two roll-table effects the [overview](./a-flag-ability-registry-plan.md)
scoped to Effort 2 (`coin_delta`, `purchase_discount`) on top of Effort 1's built
pipeline. Effort 1 parses `Roll` (`Die` + `Outcomes`) and treats any non-`Contest`
step as auto-pass, but **never rolls the die** (`stage.go:40-43`, comment: "Roll-table
outcome resolution lands in Effort 2"). This effort adds the roll stage, threads the
result into `StepContext`, and registers the two effect handlers that map a roll
through `Outcomes` to mutations. Authors the five wealth assets that resolve through
them.

- Overview plan: [`a-flag-ability-registry-plan.md`](./a-flag-ability-registry-plan.md)
  (Effort 2 row; the overview's Commit 1/Commit 2 outline is superseded by this file's
  Work Breakdown, re-grounded against Effort 1's built API).
- Effort 1 plan: [`-effort-1-plan.md`](./a-flag-ability-registry-effort-1-plan.md)
  (the `StepContext`/`EffectHandler` contract, halt-on-failure loop, `RollTable`/
  `AbilityOutcome` shapes this effort binds to).
- Architecture: [ability-catalog](../../architecture/engine/ability-catalog.md)
  (roll-table-coin and roll-table→purchase-discount clusters).

## Decisions Ratified in Planning

1. **The roll is a pipeline stage, not handler work.** Mirroring the Contest gate,
   `Dispatch` rolls `step.Roll.Die` once per step and threads the result into
   `StepContext.RollResult`; handlers map a result, they never roll. Keeps `Roller`
   out of the handler contract (Effort 1 Decision 3: "Effort 2 adds the resolved roll
   result") and keeps a step's randomness resolved in one place.

2. **An unmatched roll band is zero payout (no mutation), expressed as the absence of
   a row.** Harvesters ("3+ → 1 Coin") is a single `from=3 to=6 coin=1` row; rolls
   1–2 match no outcome and emit nothing. The data stays sparse — the "0 Coin" band is
   the lack of a row, not a row of zeros. `outcomeFor` returns `(_, false)` on no match.

3. **Insolvency rule + unconditional `DestroySelf` flag** (Robert, this session).
   The `coin_delta` handler destroys the acting asset when a *negative* payout the
   faction cannot afford comes up (Postech roll=1: `coin=-1`, no flag — destroyed only
   if `faction.Coin < 1`). `DestroySelf=true` means *unconditional* destroy regardless
   of Coin (Venture Capital roll=1: `coin=0`, `destroy_self=true`). One implicit handler
   rule covers the conditional case; the flag covers the unconditional one. Resolves the
   overview's "value-dependent payout" sub-detail for the destroy case.

4. **Pending purchase discount: a Faction field set by mutation, consumed at the next
   Buy Asset** (Robert, this session). `purchase_discount` emits
   `PurchaseDiscountGranted` → the apply engine writes `Faction.PendingPurchaseDiscount`;
   `BuyAsset.Resolve` reads it, subtracts from the resolved cost (floored at half the
   *base* cost, rounded down), and emits `PurchaseDiscountConsumed` to zero it. The
   grant **sets** (overwrites), it does not accumulate — SWN scopes the discount to "the
   next purchase," so a second roll before buying replaces the first. The hook-based
   `AssetCostModifier` seam was rejected: that registry is rebuilt from static S-profiles
   at cycle start and can't carry a one-shot, action-granted, cross-turn discount.

5. **The discount bookkeeping mutations are not surfaced in the digest.**
   `PurchaseDiscountGranted`/`Consumed` are internal state-tracking; the player-facing
   signal is the reduced `CoinDelta` the purchase already emits. `NewMutationRecord`
   (`domain/event.go:18`) marshals them generically, so they record to history without
   a serialization case; the digest's type-string switch ignores unknown types, so no
   digest case is added.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

---

## Open Questions — To Ratify at Implementation Time

- **`coin_delta` Cause string.** Tasks below use `Cause: "ability"` for both the
  `CoinDelta` and `AssetRemoved` payouts, matching `revealStealth`'s `AssetStealthCleared`
  cause. If the digest/narrative wants to distinguish roll-table income from a contest
  reveal, refine at first use — `"ability"` is the safe default and no digest case reads
  it yet.
- **Effort-1 carried question (multi-step built + unbuilt effects) does not fire here.**
  All five Effort-2 assets are single-step roll-tables with registered handlers; none is
  a description-only or mixed multi-step ability, so the no-handler fallthrough's
  mutation-discard hazard never triggers. Remains deferred to the first authored
  description-only special.

## Shared Context

**Model.** This Plan session runs on Opus. For execution: both commits are
handler-authoring on an established pipeline — Sonnet is appropriate. Commit 2 touches
Buy Asset cost logic and a new Faction field, so it carries slightly more design weight;
flag if it feels heavier than mechanical. Surface the recommendation at the Plan →
Execution transition.

**Re-ground before coding.** Effort 1's built API was verified this session and matches
its plan (`StepContext`/`EffectHandler` in `dispatch.go`, `resolveTarget`/`passesGate`
in `stage.go`, `RollTable`/`AbilityOutcome{Coin signed, DestroySelf, Discount}` in
`domain/asset.go`). Re-confirm at execution.

**Test patterns.** Follow `ability/dispatch_test.go`: `seqRoller{values}` for
deterministic rolls, gomock `MockCollector` for prompts. Roll-table steps use
`Target: TargetNone` (self-payout) and no `Contest`, so `resolveTarget` never calls the
collector and `passesGate` auto-passes — handler tests can pass a `MockCollector` with
no expectations, or drive `Dispatch` end-to-end.

## Out of Scope

- `relocate` (Effort 3); `coin_or_disable`, `purchase_enabler`, `attach` (Effort 4).
- The S-side coin reactors and Monopoly (A.001.3+).
- Any change to the Contest gate, target stages, or the flat-block/`[[steps]]` parser
  beyond adding the two new effect strings.

---

## Work Breakdown

Single phase, two commits, executed in order.

### Phase 1 — Roll-table effects

#### Commit 1 — `feat: coin_delta roll-table ability`

##### Task 1 — `internal/faction/domain/asset.go`

Add the `coin_delta` effect constant.

**Find:**
```go
const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
)
```
**Replace with:**
```go
const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
	EffectCoinDelta     AbilityEffectType = "coin_delta"
)
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Add the parser case for `coin_delta`.

**Find:**
```go
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
```
**Replace with:**
```go
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	case "coin_delta":
		return domain.EffectCoinDelta, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
```

##### Task 3 — `internal/faction/engine/action/actions/ability/dispatch.go`

Add `RollResult` to `StepContext` and roll the die in the loop after the gate passes.

**Find:**
```go
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
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
}
```

**Find:**
```go
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Step:         step,
			Target:       target,
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
		})
```

##### Task 4 — `internal/faction/engine/action/actions/ability/stage.go`

Add the roll stage and the outcome-band lookup helper alongside the existing stages.

After `passesGate` (before `otherFactions`), **add:**
```go
// resolveRoll evaluates the step's roll table, or returns 0 for a non-roll step.
// The Contest gate and the roll table are mutually exclusive (parser-enforced), so
// a step that reaches a handler has at most one source of randomness.
func resolveRoll(step domain.AbilityStep, roller domain.Roller) int {
	if step.Roll == nil {
		return 0
	}
	return step.Roll.Die.Roll(roller)
}

// outcomeFor returns the first outcome whose [From,To] range contains roll. An
// unmatched roll yields (zero, false) — the absence of a row is a zero payout.
func outcomeFor(roll int, outcomes []domain.AbilityOutcome) (domain.AbilityOutcome, bool) {
	for _, outcome := range outcomes {
		if roll >= outcome.From && roll <= outcome.To {
			return outcome, true
		}
	}
	return domain.AbilityOutcome{}, false
}
```

##### Task 5 — `internal/faction/engine/action/actions/ability/coin_delta.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// coinDelta resolves a roll-table coin ability: map the rolled result through the
// step's Outcomes to a signed CoinDelta. A row flagged DestroySelf destroys the
// acting asset unconditionally; a negative payout the faction cannot afford also
// destroys it (the SWN insolvency rule). An unmatched roll pays nothing.
func coinDelta(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Roll == nil {
		return nil, fmt.Errorf("coin_delta: step has no roll table")
	}
	outcome, ok := outcomeFor(ctx.RollResult, ctx.Step.Roll.Outcomes)
	if !ok {
		return nil, nil
	}
	if outcome.DestroySelf || (outcome.Coin < 0 && ctx.Faction.Coin < -outcome.Coin) {
		return []domain.Mutation{domain.AssetRemoved{
			FactionID:         ctx.Faction.ID,
			AssetID:           ctx.Asset.ID,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		}}, nil
	}
	if outcome.Coin == 0 {
		return nil, nil
	}
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             outcome.Coin,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
```

##### Task 6 — `internal/faction/engine/action/actions/ability/dispatch.go`

Register the handler.

**Find:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth: revealStealth,
}
```
**Replace with:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth: revealStealth,
	domain.EffectCoinDelta:     coinDelta,
}
```

##### Task 7 — `rulebooks/swn/assets/wealth_assets.toml`

Add a flat `[ability]` block to each of the four assets (none has one today — they
fall through to `confirmApplied`). Insert each block immediately after the asset's
`description` line. Each is a single-step coin_delta with no target.

Harvesters — **Find:**
```toml
description = "Workers gathering natural resources — ore, biologicals, or other unprocessed goods. As an ability action, the owning faction rolls 1d6; on a 3 or higher, they gain 1 Coin."
```
**Replace with:**
```toml
description = "Workers gathering natural resources — ore, biologicals, or other unprocessed goods. As an ability action, the owning faction rolls 1d6; on a 3 or higher, they gain 1 Coin."

[assets.Harvesters.ability]
effect = "coin_delta"
die = "1d6"
[[assets.Harvesters.ability.outcomes]]
from = 3
to = 6
coin = 1
```

Postech Industry — **Find:**
```toml
description = "Advanced manufacturing facilities. As an ability action, roll 1d6: on a 1, lose 1 Coin (if this cannot be paid, the asset is destroyed); on a 2–4, gain 1 Coin; on a 5–6, gain 2 Coin."
```
**Replace with:**
```toml
description = "Advanced manufacturing facilities. As an ability action, roll 1d6: on a 1, lose 1 Coin (if this cannot be paid, the asset is destroyed); on a 2–4, gain 1 Coin; on a 5–6, gain 2 Coin."

[assets.PostechIndustry.ability]
effect = "coin_delta"
die = "1d6"
[[assets.PostechIndustry.ability.outcomes]]
from = 1
to = 1
coin = -1
[[assets.PostechIndustry.ability.outcomes]]
from = 2
to = 4
coin = 1
[[assets.PostechIndustry.ability.outcomes]]
from = 5
to = 6
coin = 2
```

Venture Capital — **Find:**
```toml
description = "Financial instruments that grow resources through aggressive investment. As an ability action, roll 1d8: on a 1, the asset is destroyed; on a 2–3, gain 1 Coin; on 4–7, gain 2 Coin; on an 8, gain 3 Coin."
```
**Replace with:**
```toml
description = "Financial instruments that grow resources through aggressive investment. As an ability action, roll 1d8: on a 1, the asset is destroyed; on a 2–3, gain 1 Coin; on 4–7, gain 2 Coin; on an 8, gain 3 Coin."

[assets.VentureCapital.ability]
effect = "coin_delta"
die = "1d8"
[[assets.VentureCapital.ability.outcomes]]
from = 1
to = 1
destroy_self = true
[[assets.VentureCapital.ability.outcomes]]
from = 2
to = 3
coin = 1
[[assets.VentureCapital.ability.outcomes]]
from = 4
to = 7
coin = 2
[[assets.VentureCapital.ability.outcomes]]
from = 8
to = 8
coin = 3
```

Pretech Manufactory — **Find:**
```toml
description = "Rare, functioning Mandate-era manufacturing facilities. As an action, the owning faction may roll 1d8 and gain half that many Coin, rounded up."
```
**Replace with:**
```toml
description = "Rare, functioning Mandate-era manufacturing facilities. As an action, the owning faction may roll 1d8 and gain half that many Coin, rounded up."

[assets.PretechManufactory.ability]
effect = "coin_delta"
die = "1d8"
[[assets.PretechManufactory.ability.outcomes]]
from = 1
to = 2
coin = 1
[[assets.PretechManufactory.ability.outcomes]]
from = 3
to = 4
coin = 2
[[assets.PretechManufactory.ability.outcomes]]
from = 5
to = 6
coin = 3
[[assets.PretechManufactory.ability.outcomes]]
from = 7
to = 8
coin = 4
```

##### Task 8 — `internal/faction/engine/action/actions/ability/coin_delta_test.go` (new) — full contents:

Table-driven over the roll bands; one subtest proves the unconditional vs insolvency
destroy split. Hand-builds defs on the step shape (no parser), mirroring `dispatch_test.go`.

```go
package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func coinDeltaStep(die domain.DiceRoll, outcomes []domain.AbilityOutcome) domain.AbilityStep {
	return domain.AbilityStep{
		Effect: domain.EffectCoinDelta,
		Target: domain.TargetNone,
		Roll:   &domain.RollTable{Die: die, Outcomes: outcomes},
	}
}

func TestCoinDelta_Payouts(t *testing.T) {
	postech := []domain.AbilityOutcome{
		{From: 1, To: 1, Coin: -1},
		{From: 2, To: 4, Coin: 1},
		{From: 5, To: 6, Coin: 2},
	}
	harvesters := []domain.AbilityOutcome{{From: 3, To: 6, Coin: 1}}

	tests := []struct {
		name       string
		outcomes   []domain.AbilityOutcome
		roll       int
		factionCoin int
		wantCoin   int  // expected CoinDelta.Delta; 0 => no CoinDelta
		wantRemove bool // expected AssetRemoved
	}{
		{"harvesters miss", harvesters, 2, 0, 0, false},
		{"harvesters hit", harvesters, 3, 0, 1, false},
		{"postech gain", postech, 5, 0, 2, false},
		{"postech loss affordable", postech, 1, 3, -1, false},
		{"postech loss insolvent", postech, 1, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := StepContext{
				Faction:    &domain.Faction{ID: "f1", Coin: tt.factionCoin},
				Asset:      &domain.Asset{ID: "a1", OwnerID: "f1"},
				Step:       coinDeltaStep(domain.DiceRoll{NumDice: 1, Sides: 6}, tt.outcomes),
				RollResult: tt.roll,
			}
			muts, err := coinDelta(ctx)
			if err != nil {
				t.Fatalf("coinDelta: %v", err)
			}
			assertPayout(t, muts, tt.wantCoin, tt.wantRemove)
		})
	}
}

func TestCoinDelta_DestroySelfUnconditional(t *testing.T) {
	outcomes := []domain.AbilityOutcome{{From: 1, To: 1, DestroySelf: true}}
	ctx := StepContext{
		Faction:    &domain.Faction{ID: "f1", Coin: 99}, // affordable, still destroyed
		Asset:      &domain.Asset{ID: "a1", OwnerID: "f1"},
		Step:       coinDeltaStep(domain.DiceRoll{NumDice: 1, Sides: 8}, outcomes),
		RollResult: 1,
	}
	muts, err := coinDelta(ctx)
	if err != nil {
		t.Fatalf("coinDelta: %v", err)
	}
	assertPayout(t, muts, 0, true)
}

func assertPayout(t *testing.T, muts []domain.Mutation, wantCoin int, wantRemove bool) {
	t.Helper()
	var gotCoin int
	var gotRemove bool
	for _, m := range muts {
		switch v := m.(type) {
		case domain.CoinDelta:
			gotCoin = v.Delta
		case domain.AssetRemoved:
			gotRemove = true
		default:
			t.Errorf("unexpected mutation %T", m)
		}
	}
	if gotCoin != wantCoin {
		t.Errorf("CoinDelta.Delta = %d, want %d", gotCoin, wantCoin)
	}
	if gotRemove != wantRemove {
		t.Errorf("AssetRemoved present = %v, want %v", gotRemove, wantRemove)
	}
}
```

##### Task 9 — `internal/faction/rulebook/rulebook_test.go`

Add one parse assertion guarding the TOML authoring (Venture, since it exercises both
a `destroy_self` row and signed coin). Place after the existing Seductress subtest.

```go
	t.Run("venture capital coin_delta roll table parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W6-001"]
		if !ok {
			t.Fatal("asset SWN-W6-001 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectCoinDelta {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectCoinDelta)
		}
		if step.Roll == nil || step.Roll.Die.Sides != 8 {
			t.Fatalf("roll: got %+v, want 1d8", step.Roll)
		}
		if len(step.Roll.Outcomes) != 4 {
			t.Fatalf("outcomes: got %d, want 4", len(step.Roll.Outcomes))
		}
		if !step.Roll.Outcomes[0].DestroySelf {
			t.Errorf("outcome[0]: want destroy_self on roll 1")
		}
	})
```

##### Commit message
```
feat: coin_delta roll-table ability

- roll stage: Dispatch rolls step.Roll.Die and threads RollResult into StepContext
- coin_delta handler maps RollResult through Outcomes to a signed CoinDelta;
  unaffordable negative payout or DestroySelf emits AssetRemoved for the acting asset
- outcomeFor band lookup; unmatched roll pays nothing
- EffectCoinDelta constant + parser case
- author Harvesters, Postech Industry, Venture Capital, Pretech Manufactory
```

#### Commit 2 — `feat: purchase_discount roll-table ability`

##### Task 1 — `internal/faction/domain/asset.go`

Add the `purchase_discount` effect constant.

**Find:**
```go
const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
	EffectCoinDelta     AbilityEffectType = "coin_delta"
)
```
**Replace with:**
```go
const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
)
```

##### Task 2 — `internal/faction/domain/mutation.go`

Add the grant/consume mutations. Place after `AssetRemoved` (lines 29–37).

**Find:**
```go
func (m AssetRemoved) Type() string { return "asset_removed" }
```
**Replace with:**
```go
func (m AssetRemoved) Type() string { return "asset_removed" }

// PurchaseDiscountGranted sets a faction's pending discount on its next asset
// purchase (Commodities Broker). It overwrites any existing pending discount —
// the discount is scoped to the next purchase, not accumulated.
type PurchaseDiscountGranted struct {
	FactionID         string `json:"faction_id"`
	Amount            int    `json:"amount"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m PurchaseDiscountGranted) Type() string { return "purchase_discount_granted" }

// PurchaseDiscountConsumed clears a faction's pending purchase discount, emitted by
// Buy Asset once the discount has been applied to a purchase.
type PurchaseDiscountConsumed struct {
	FactionID         string `json:"faction_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m PurchaseDiscountConsumed) Type() string { return "purchase_discount_consumed" }
```

##### Task 3 — `internal/faction/domain/faction.go`

Add the pending-discount field to `Faction`.

**Find:**
```go
	HookBudgets map[string]int    `toml:"hook_budgets,omitempty"`
}
```
**Replace with:**
```go
	HookBudgets map[string]int    `toml:"hook_budgets,omitempty"`
	PendingPurchaseDiscount int   `toml:"pending_purchase_discount,omitempty"`
}
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

Add apply cases for the two mutations. Place after the `AssetRemoved` case (lines 61–66).

**Find:**
```go
		case domain.AssetRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				delete(faction.Assets, v.AssetID)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```
**Replace with:**
```go
		case domain.AssetRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				delete(faction.Assets, v.AssetID)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.PurchaseDiscountGranted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = v.Amount
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.PurchaseDiscountConsumed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = 0
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```

##### Task 5 — `internal/faction/rulebook/rulebook.go`

Add the parser case.

**Find:**
```go
	case "coin_delta":
		return domain.EffectCoinDelta, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
```
**Replace with:**
```go
	case "coin_delta":
		return domain.EffectCoinDelta, nil
	case "purchase_discount":
		return domain.EffectPurchaseDiscount, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
```

##### Task 6 — `internal/faction/engine/action/actions/ability/purchase_discount.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// purchaseDiscount resolves a roll-table discount ability (Commodities Broker): map
// the rolled result through the step's Outcomes to a pending discount on the
// faction's next purchase. The half-price floor is applied at purchase time
// (BuyAsset), not here. An unmatched or zero-discount roll grants nothing.
func purchaseDiscount(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Step.Roll == nil {
		return nil, fmt.Errorf("purchase_discount: step has no roll table")
	}
	outcome, ok := outcomeFor(ctx.RollResult, ctx.Step.Roll.Outcomes)
	if !ok || outcome.Discount == 0 {
		return nil, nil
	}
	return []domain.Mutation{domain.PurchaseDiscountGranted{
		FactionID:         ctx.Faction.ID,
		Amount:            outcome.Discount,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
```

##### Task 7 — `internal/faction/engine/action/actions/ability/dispatch.go`

Register the handler.

**Find:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth: revealStealth,
	domain.EffectCoinDelta:     coinDelta,
}
```
**Replace with:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
}
```

##### Task 8 — `internal/faction/engine/action/actions/buy_asset.go`

Apply the pending discount at cost resolution and emit the consume mutation.

Add a field to the `BuyAsset` struct — **Find:**
```go
	cost          int
	stealthTarget string // asset ID to stealth when buying SWN-C3-002; empty if no eligible target
}
```
**Replace with:**
```go
	cost             int
	discountConsumed bool
	stealthTarget    string // asset ID to stealth when buying SWN-C3-002; empty if no eligible target
}
```

In `Resolve`, apply the discount after the cost is resolved — **Find:**
```go
	def := ba.buyOrder.Definition
	ba.cost = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost)
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
```
**Replace with:**
```go
	def := ba.buyOrder.Definition
	ba.cost = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost)
	if faction.PendingPurchaseDiscount > 0 {
		floor := def.Cost / 2 // min half normal (base) price, rounded down
		ba.cost -= faction.PendingPurchaseDiscount
		if ba.cost < floor {
			ba.cost = floor
		}
		ba.discountConsumed = true
	}
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
```

In `Output`, emit the consume mutation when a discount was applied — **Find:**
```go
	if ba.stealthTarget != "" {
		mutations = append(mutations, domain.AssetStealthApplied{
			FactionID:         ba.factionID,
			AssetID:           ba.stealthTarget,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	return mutations, nil
```
**Replace with:**
```go
	if ba.stealthTarget != "" {
		mutations = append(mutations, domain.AssetStealthApplied{
			FactionID:         ba.factionID,
			AssetID:           ba.stealthTarget,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	if ba.discountConsumed {
		mutations = append(mutations, domain.PurchaseDiscountConsumed{
			FactionID:         ba.factionID,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	return mutations, nil
```

##### Task 9 — `rulebooks/swn/assets/wealth_assets.toml`

Add a flat `[ability]` block to Commodities Broker (8 one-per-face discount rows).

**Find:**
```toml
description = "Specialists who reduce large-scale investment costs through precise market timing. As an ability action, roll 1d8; that many Coin are subtracted from the cost of the faction's next asset purchase, down to a minimum of half normal price rounded down."
```
**Replace with:**
```toml
description = "Specialists who reduce large-scale investment costs through precise market timing. As an ability action, roll 1d8; that many Coin are subtracted from the cost of the faction's next asset purchase, down to a minimum of half normal price rounded down."

[assets.CommoditiesBroker.ability]
effect = "purchase_discount"
die = "1d8"
[[assets.CommoditiesBroker.ability.outcomes]]
from = 1
to = 1
discount = 1
[[assets.CommoditiesBroker.ability.outcomes]]
from = 2
to = 2
discount = 2
[[assets.CommoditiesBroker.ability.outcomes]]
from = 3
to = 3
discount = 3
[[assets.CommoditiesBroker.ability.outcomes]]
from = 4
to = 4
discount = 4
[[assets.CommoditiesBroker.ability.outcomes]]
from = 5
to = 5
discount = 5
[[assets.CommoditiesBroker.ability.outcomes]]
from = 6
to = 6
discount = 6
[[assets.CommoditiesBroker.ability.outcomes]]
from = 7
to = 7
discount = 7
[[assets.CommoditiesBroker.ability.outcomes]]
from = 8
to = 8
discount = 8
```

##### Task 10 — `internal/faction/engine/action/actions/ability/purchase_discount_test.go` (new) — full contents:

```go
package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestPurchaseDiscount_GrantsRolledAmount(t *testing.T) {
	outcomes := make([]domain.AbilityOutcome, 8)
	for i := range outcomes {
		outcomes[i] = domain.AbilityOutcome{From: i + 1, To: i + 1, Discount: i + 1}
	}
	ctx := StepContext{
		Faction: &domain.Faction{ID: "f1"},
		Asset:   &domain.Asset{ID: "a1", OwnerID: "f1"},
		Step: domain.AbilityStep{
			Effect: domain.EffectPurchaseDiscount,
			Target: domain.TargetNone,
			Roll:   &domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 8}, Outcomes: outcomes},
		},
		RollResult: 5,
	}
	muts, err := purchaseDiscount(ctx)
	if err != nil {
		t.Fatalf("purchaseDiscount: %v", err)
	}
	if len(muts) != 1 {
		t.Fatalf("len(muts) = %d, want 1", len(muts))
	}
	granted, ok := muts[0].(domain.PurchaseDiscountGranted)
	if !ok {
		t.Fatalf("muts[0] type = %T, want PurchaseDiscountGranted", muts[0])
	}
	if granted.Amount != 5 || granted.FactionID != "f1" {
		t.Errorf("granted = %+v, want {f1, 5}", granted)
	}
}
```

##### Task 11 — `internal/faction/engine/action/actions/buy_asset_test.go`

Add a subtest proving the pending discount reduces the cost, is floored at half base,
and is consumed. Re-ground against the file's existing fixtures at execution — match
its faction/world/rulebook setup helpers rather than the sketch below if they differ.

Add a test in the style of the file's existing cases:
```go
func TestBuyAsset_PendingDiscount(t *testing.T) {
	// Build the file's standard buy fixture; set faction.PendingPurchaseDiscount,
	// drive Inputs/Resolve/Output for a def whose resolved cost > def.Cost/2, and assert:
	//   - the emitted buy CoinDelta.Delta == -(resolvedCost - discount), not below -(def.Cost/2)
	//   - a PurchaseDiscountConsumed mutation is emitted
	//   - a second purchase with discount==0 emits no PurchaseDiscountConsumed
	// Floor case: a discount larger than (resolvedCost - def.Cost/2) clamps cost to def.Cost/2
	// and still emits PurchaseDiscountConsumed.
}
```

##### Task 12 — `internal/faction/engine/mutation/mutation_test.go`

Add a test that `PurchaseDiscountGranted` sets `Faction.PendingPurchaseDiscount` and
`PurchaseDiscountConsumed` clears it. Match the file's existing `Apply` test pattern.

##### Commit message
```
feat: purchase_discount roll-table ability

- EffectPurchaseDiscount + parser case; purchase_discount handler grants a pending
  discount sized by the roll-table Discount outcome
- PurchaseDiscountGranted/Consumed mutations + apply cases; Faction.PendingPurchaseDiscount
- Buy Asset subtracts the pending discount (floored at half base cost) and consumes it
- author Commodities Broker
```
