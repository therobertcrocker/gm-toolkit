# A-flag Ability Registry (A.001.2) — Effort 4 Plan: purchase-enabler

## Context / Goal

Build the `purchase_enabler` effect (Pretech Logistics) on top of Effort 1's pipeline:
a self-targeted ability that spends the action to buy one asset of a given category,
up to a tech-level cap, on the acting asset's world, at a cost surcharge. It is the
first consumer of the `Purchase *PurchaseParams` sub-struct that Effort 1 (Decision 6)
reserved, mirroring how `relocate` first consumed `Relocate *RelocateParams` in Effort 3.

**This effort is `purchase_enabler` only.** Re-grounding the overview's other two Effort-4
assets against the built engine showed both depend on *persistent-restriction enforcement*
the arc assigns to the S-side hook engine (out of scope for A.001.2). They reroute to
A.001.3+ (see Decisions 1–2). Effort 4 collapses from the overview's three commits to one.

- Overview plan: [`a-flag-ability-registry-plan.md`](./a-flag-ability-registry-plan.md)
  (Effort 4 row; the overview's three-commit Work Breakdown is superseded by this file,
  and its Decision Record logs the reroute + the per-effort-plan deviation from Decision 8).
- Effort 1 plan: [`-effort-1-plan.md`](./a-flag-ability-registry-effort-1-plan.md)
  (the `StepContext`/`EffectHandler` contract, the reserved `Purchase *PurchaseParams`
  sub-struct this effort first consumes).
- Effort 3 plan: [`-effort-3-plan.md`](./a-flag-ability-registry-effort-3-plan.md)
  (the direct-emit-in-handler precedent, the threaded `Collector`/`WorldEngine` on
  `StepContext`, the parser sub-table + collector-fan-out pattern this effort mirrors).
- Architecture: [ability-catalog](../../architecture/engine/ability-catalog.md)
  (the `purchase-enabler` row — corrected from "discounted" to "surcharged", Decision 3).

## Decisions Ratified in Planning

1. **Effort 4 is `purchase_enabler` only; Marketers (`coin_or_disable`) and Seditionists
   (`attach`/restrict-action) reroute to S-side (A.001.3)** (Robert, this session). The
   A-side abilities engine emits mutations *now*, in one action; a persistent restriction
   ("the asset is disabled until paid" / "the affected asset cannot attack") has no "now"
   — it only exists as something checked at a *future* enemy attack, which the catalog
   assigns to the S-side hook engine (Cat 3/5), explicitly out of scope for A.001.2. No
   attack/disable-restriction hook exists today (Saboteurs/Lawyers/Planetary Defenses are
   all unbuilt). Building a `Disabled` flag or attachment marker that nothing enforces
   would be a flag-nothing-reads (violates Correctness + YAGNI). Both assets defer whole
   to the S-side effort that builds restriction enforcement. Reverses the overview's
   three-commit Effort-4 Work Breakdown.

2. **Marketers' contest + pay-half branch is genuinely A-side, but does not ship alone.**
   The Cunning-vs-Wealth contest and the forced pay-half `CoinDelta` are A-side-resolvable
   (and would be the first consumer of the unwired `TargetAsset` stage). But the "or be
   disabled" alternative is inseparable from the rule — a half-built Marketers that can
   only force payment, never disable, misrepresents the asset. It reroutes whole with its
   disable branch, not as an A-side fragment. Records the rationale so a future session
   doesn't "rescue" the pay-branch in isolation.

3. **Pretech Logistics is a surcharge, not a discount** (re-grounding correction). SWN:
   *"at half again as many Coin as usual, rounded up."* "Half again" = 1.5× base cost — a
   premium paid for buying a Force asset above the world's tech level (up to TL 5), not a
   discount. The overview Work Breakdown and the ability-catalog's `purchase-enabler` row
   both read "discounted"; both are wrong. The surcharge is data (`cost_percent = 150`),
   honoring the catalog's data-driven thesis (cf. Effort 3 making `asset_type` data). The
   catalog row is corrected at pre-merge (graduation step).

4. **The handler buys by direct emission, not by reusing the Buy Asset action** (Robert,
   this session). `purchaseEnabler` gathers candidate definitions, prompts a dedicated
   collector pick, computes the surcharged cost from `def.Cost`, checks affordability, and
   emits `AssetAdded` (Ready:false) + a surcharged `CoinDelta` — mirroring how `relocate`
   emits via the world engine rather than invoking a sibling action. Reuses
   `domain.NewPurchasedAsset` for the new asset (Decision 8) and the package-local
   `statScore` for the rating gate; does **not** couple ability dispatch to the `BuyAsset`
   action's collector flow.

5. **The purchase ignores cost-modifier hooks; the surcharge is computed off raw
   `def.Cost`.** `dispatch.ResolveAssetCost` applies the Cat-4 cost-modifier registry
   (Local Investments surcharge, tech-uplift, Bank), but the hooks `Registry` is **not**
   threaded into `StepContext` and threading it is more wiring than this one-asset effect
   warrants (YAGNI). "Half again as much as usual" reads as a flat multiple of the base
   listed cost; `surchargedCost(def.Cost, CostPercent)` uses the raw base. If a future
   asset needs hook-modified pricing in an ability, thread the registry then.

6. **"Only one asset per turn" is enforced by the action economy, no extra bookkeeping.**
   Using Pretech Logistics' ability spends the faction's turn action and the handler buys
   exactly one asset per invocation, so the SWN "only one asset can be purchased per turn"
   clause falls out of the action model — no per-turn counter or flag is added.

7. **The candidate filter applies the P-flag / Planetary-Government gate** (Robert, this
   session — ratifying the first Open Question). The data has five P-flagged Force defs at
   TL ≤ 5 (Militia Unit, Elite Skirmishers, Postech Infantry, Pretech Infantry, Gravtank
   Formation), so the gate is live, not hypothetical. Pretech Logistics waives the
   *tech-level* requirement, never the *permission* requirement; omitting the gate would let
   a faction buy a P-flagged asset on a world where it holds no Planetary Government —
   something normal Buy Asset (`buy_asset.go:210-214`) forbids. `eligiblePurchaseDefs`
   reuses the graduated `domain.HasPlanetaryGovernment` (see Decision 8) rather than
   re-inlining the check.

8. **Shared purchase logic graduates to `domain` instead of being mirrored** (Robert, this
   session — superseding the review's "small copy" framing in Decision 7's first draft).
   The handler buys by direct emission (Decision 4 stands — dispatch is *not* coupled to the
   `BuyAsset` action), but the **leaf** logic both purchase paths need moves down to
   `domain`, the only package both `actions` and `ability` import (the cycle runs
   `actions → ability` via `use_asset_ability.go:72`, so sideways sharing is impossible).
   Two helpers graduate: `domain.NewPurchasedAsset(faction, def, worldID)` (the standard
   `{NextAssetID, CurrentHP, Ready:false, Maintained:true}` asset) and
   `domain.HasPlanetaryGovernment(factionID, basesOnWorld)`. `buy_asset.go` is refactored
   onto both — which also retires its `nextAssetSuffix`, a pre-existing duplicate of
   `domain.NextAssetID`. The eligibility *predicate* is deliberately **not** graduated: its
   TL semantics (fixed cap vs world TL) and pricing (surcharge vs `ResolveAssetCost` +
   discount) legitimately diverge between the two callers (Robert: "asset-build + base-check
   only"). This adds a refactor commit ahead of the feature commit (see Work Breakdown).

## Decision Record — Execution

<!-- Execution-time decisions and reversals land here, logged before the commit they
ride. Append AND reconcile. -->

- **`refit_asset.go` also consumed `nextAssetSuffix`; updated to `domain.NextAssetID` directly
  (not `domain.NewPurchasedAsset`).** Task B in the plan only mentioned `buy_asset.go`, but
  `nextAssetSuffix` was package-local in `actions` and used by both files. The fix uses
  `domain.NextAssetID(faction, newDef.ID)` directly rather than `domain.NewPurchasedAsset`
  because refit copies the full `Location` (including `RegionHex`) from the old asset — the
  new asset stays at the exact hex the old one occupied, not just the world. `NewPurchasedAsset`
  takes only `worldID` and zero-initialises `RegionHex`, which would be wrong here.

- **Handler tests use `mocks.MockCollector` (gomock), not `testharness.ScriptedCollector`.**
  The plan suggested `ScriptedCollector`, but all existing tests in the `ability` package use
  the mock pattern. `testharness` imports `world`, which creates a test-time import cycle when
  imported from `ability` tests (the `go test` tool detects it at build time even though the
  compiled binary doesn't carry the cycle). `MockCollector` from the `actions/mocks` package
  has no cycle and is the established convention in this package.

---

## Open Questions — To Ratify at Implementation Time

- ~~**P-flag (Permission) requirement on enabler purchases.**~~ *Resolved (Decisions 7 & 8):*
  the data has P-flagged Force defs at TL ≤ 5, so `eligiblePurchaseDefs` includes the
  Planetary-Government gate, reusing the graduated `domain.HasPlanetaryGovernment` (behind a
  nil-guarding `holdsGovernment` wrapper) shared with `buy_asset.go`.
- **Effort-1 carried question (multi-step built + unbuilt effects) does not fire here.**
  Pretech Logistics is a single-step ability with a registered handler; no description-only
  or mixed multi-step path is exercised. Remains deferred to the first authored
  description-only special.

## Shared Context

**Model.** This Plan session runs on Opus. For execution: the handler + parser + authoring
are mechanical (Sonnet). The commit's heaviest surface is the new `SelectPurchaseEnablerAsset`
collector method across all five implementations (mock regen, scripted, smoke, TUI adapter)
plus a TUI selection overlay — comparable to Effort 3 Commit 3's `SelectRelocateOrders`
fan-out. Sonnet is appropriate; flag if the overlay feels heavier than mechanical. Surface
the recommendation at the Plan → Execution transition.

**Re-ground before coding (verified this session).** Effort 1–3's built API matches their
plans: `StepContext{..., Collector, WorldEngine}` + the `handlers` map in `dispatch.go:14-35`;
`AbilityStep{Effect, Target, Contest, Roll, Relocate}` in `domain/asset.go:114-120` (this
effort adds `Purchase`); `convertAbilityStep(effectStr, targetStr, attackerStr, defenderStr,
dieStr string, outcomes []outcomeRecord, relocate *relocateRecord)` in `rulebook.go:264`
(this effort adds a `purchase *purchaseRecord` param, threaded from both call sites at
`rulebook.go:249,257`); `toFactionStat` (`rulebook.go:504`) parses the category;
`domain.NextAssetID` (`asset.go:188`); `statScore` (`stage.go:90`, package-local). The
`TargetAsset`/`TargetWorld` cases of `resolveTarget` stay unwired — `purchase_enabler`
uses `Target = none` and reads the acting asset's world directly. Re-confirm at execution.

**Test patterns.** Follow `ability/dispatch_test.go`: a gomock `MockCollector` (or the
testharness `ScriptedCollector`) for the def pick. The handler reads `ctx.Rulebook.Assets`,
so tests build a `&rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{...}}` with
a small spread of Force/non-Force defs across tech levels and costs. `Target = none` means
`resolveTarget` never calls the collector for target selection; only `SelectPurchaseEnablerAsset`
is expected.

## Out of Scope

- **Marketers `coin_or_disable` and Seditionists `attach`/restrict-action** — rerouted to
  S-side (A.001.3), pending the restriction/disable-enforcement hook engine (Decisions 1–2).
- Wiring the `TargetAsset`/`TargetWorld` cases of `resolveTarget` (Marketers would have been
  `TargetAsset`'s first consumer; it reroutes, so the case stays unwired).
- Threading the hooks `Registry` into `StepContext` for hook-modified ability pricing
  (Decision 5).
- All other S-side work (coin reactors, Monopoly, Transit Web, `GrantedMovementAbilities`),
  goals/tags, `_core`, and H.A.W.K authoring (later initiatives).

---

## Work Breakdown

Single phase, two commits: a refactor that graduates the shared purchase leaves to `domain`
(Decision 8), then the feature built on top.

### Phase 1 — purchase-enabler

#### Commit 1 — `refactor: graduate purchased-asset construction + planetary-government check to domain`

##### Task A — `internal/faction/domain/asset.go` (graduate the shared helpers)

After `NextAssetID`, **add** the asset-construction helper both purchase paths use:
```go
// NewPurchasedAsset builds a freshly-bought asset for a faction: a fresh ID, full HP, not
// yet Ready (activates next turn), and Maintained. Shared by the Buy Asset action and the
// purchase_enabler ability so the two purchase paths can't drift on the asset shape.
func NewPurchasedAsset(faction *Faction, def *AssetDefinition, worldID string) Asset {
	return Asset{
		ID:           NextAssetID(faction, def.ID),
		DefinitionID: def.ID,
		OwnerID:      faction.ID,
		Location:     Location{WorldID: worldID},
		CurrentHP:    def.HP,
		Ready:        false,
		Maintained:   true,
	}
}

// HasPlanetaryGovernment reports whether the faction owns a base on the world (the SWN
// P-flag permission gate). basesOnWorld is the slice for one world (a nil slice is "no
// bases" → false), so callers pass worldEngine.Index.BasesByLocation[worldID].
func HasPlanetaryGovernment(factionID string, basesOnWorld []*Base) bool {
	for _, base := range basesOnWorld {
		if base.OwnerID == factionID {
			return true
		}
	}
	return false
}
```
Re-ground at execution: confirm `domain.Base.OwnerID` and that `domain` does not already
import `world` (these helpers must stay world-engine-free — they take the `[]*Base` slice,
not the engine).

##### Task B — `internal/faction/engine/action/actions/buy_asset.go` (refactor onto the helpers)

- In `Resolve`, replace the `ba.newAsset = domain.Asset{...}` literal (lines 95-103) with
  `ba.newAsset = domain.NewPurchasedAsset(faction, def, ba.buyOrder.World)`.
- In `purchasableDefinitions`, replace the `factionHasPlanetaryGovernment(faction.ID, ...)`
  call with `domain.HasPlanetaryGovernment(faction.ID, ...)`.
- **Delete** the now-unused local funcs `nextAssetSuffix` and `factionHasPlanetaryGovernment`.
- Prune imports that those funcs alone used: `strconv` and `strings` (re-ground — confirm no
  other use in the file before removing). `fmt`/`sort` stay.
- Existing `buy_asset` tests must stay green unchanged — the refactor is behavior-preserving
  (`NextAssetID` and `nextAssetSuffix` produce identical IDs; the government check is the
  same loop).

##### Task C — `internal/faction/domain/asset_test.go` (guard the graduated helpers)

Add focused tests: `NewPurchasedAsset` sets `Ready:false`/`Maintained:true`/`CurrentHP=def.HP`
and a monotonic ID; `HasPlanetaryGovernment` returns true for an owned base, false for a
rival-only slice and for nil. Match the existing table-test style in the file.

#### Commit 2 — `feat: purchase_enabler ability`

##### Task 1 — `internal/faction/domain/asset.go`

Add the effect constant, the `PurchaseParams` sub-struct, and the `Purchase` field on
`AbilityStep`.

**Find:**
```go
const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
	EffectRelocate         AbilityEffectType = "relocate"
)
```
**Replace with:**
```go
const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
	EffectRelocate         AbilityEffectType = "relocate"
	EffectPurchaseEnabler  AbilityEffectType = "purchase_enabler"
)
```

After the `RelocateParams` struct (ends `MaxAssets int`), **add:**
```go
// PurchaseParams configures a purchase_enabler step (Pretech Logistics). Category is the
// asset category the ability may buy (Force); MaxTechLevel caps the required tech level of
// a buyable def, ignoring the world's own tech level (the ability is precisely what buys
// above it); CostPercent is the surcharge as a percent of the base cost, rounded up
// (150 = "half again as much").
type PurchaseParams struct {
	Category     FactionStat
	MaxTechLevel int
	CostPercent  int
}
```

**Find:**
```go
type AbilityStep struct {
	Effect   AbilityEffectType
	Target   TargetKind
	Contest  *Contest
	Roll     *RollTable
	Relocate *RelocateParams
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
	Purchase *PurchaseParams
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Parse the effect string and a `[purchase]` sub-table on the flat block and steps.

In `toAbilityEffect`'s switch, **add** the case (after `relocate`):
```go
	case "purchase_enabler":
		return domain.EffectPurchaseEnabler, nil
```

Add the record type after `relocateRecord`:
```go
type purchaseRecord struct {
	Category     string `toml:"category"`
	MaxTechLevel int    `toml:"max_tech_level"`
	CostPercent  int    `toml:"cost_percent"`
}
```

Add `Purchase *purchaseRecord` (`toml:"purchase"`) to **both** `abilityRecord` (after
`Relocate`) and `abilityStepRecord` (after `Relocate`).

Thread the record through `convertAbilityStep`. **Find:**
```go
func convertAbilityStep(effectStr, targetStr, attackerStr, defenderStr, dieStr string, outcomes []outcomeRecord, relocate *relocateRecord) (domain.AbilityStep, error) {
```
**Replace with:**
```go
func convertAbilityStep(effectStr, targetStr, attackerStr, defenderStr, dieStr string, outcomes []outcomeRecord, relocate *relocateRecord, purchase *purchaseRecord) (domain.AbilityStep, error) {
```

Update both call sites in `convertAbility` (`rulebook.go:249,257`) to pass the purchase
record: the steps call gains `sr.Purchase`, the flat call gains `r.Purchase` (each as the
new trailing argument).

In `convertAbilityStep`, after the `relocate` block (the `step.Relocate = params` close),
**add:**
```go
	if purchase != nil {
		category, err := toFactionStat(purchase.Category)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("purchase category: %w", err)
		}
		step.Purchase = &domain.PurchaseParams{
			Category:     category,
			MaxTechLevel: purchase.MaxTechLevel,
			CostPercent:  purchase.CostPercent,
		}
	}
```
Re-ground at execution: confirm the exact close of the relocate block and that no
contest-xor-roll-xor-relocate guard rejects a `purchase`-only step (the flat Pretech
Logistics block carries `effect` + `[purchase]`, no contest/roll/relocate).

##### Task 3 — `internal/faction/engine/action/collector.go`

Add the collector method. In the `Collector` interface, after `SelectRelocateOrders`,
**add:**
```go
	SelectPurchaseEnablerAsset(world string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error)
```

##### Task 4 — `internal/faction/engine/action/actions/ability/purchase_enabler.go` (new) — full contents:

```go
package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// purchaseEnabler resolves Pretech Logistics: spend the action to buy one asset of the
// configured category (Force), with required tech level at or below the cap, on the
// acting asset's world, at a surcharge (half again, rounded up). The world's own tech
// level is ignored — the ability is precisely what lets the faction buy above it. Emits
// AssetAdded (Ready:false) + the surcharged CoinDelta.
func purchaseEnabler(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.Purchase
	if params == nil {
		return nil, fmt.Errorf("purchase_enabler: step has no purchase params")
	}
	worldID := ctx.Asset.Location.WorldID
	candidates := eligiblePurchaseDefs(ctx, params)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("purchase_enabler: no affordable %s assets at tech level %d or below", params.Category, params.MaxTechLevel)
	}
	def, err := ctx.Collector.SelectPurchaseEnablerAsset(worldID, candidates)
	if err != nil {
		return nil, fmt.Errorf("purchase_enabler: %w", err)
	}
	if def == nil {
		return nil, fmt.Errorf("purchase_enabler: %w", action.ErrNoSelection)
	}
	cost := surchargedCost(def.Cost, params.CostPercent)
	if ctx.Faction.Coin < cost {
		return nil, fmt.Errorf("purchase_enabler: insufficient Coin: need %d, have %d", cost, ctx.Faction.Coin)
	}
	newAsset := domain.NewPurchasedAsset(ctx.Faction, def, worldID)
	return []domain.Mutation{
		domain.AssetAdded{FactionID: ctx.Faction.ID, Asset: newAsset, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
		domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: -cost, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
	}, nil
}

// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category, required tech level at or below the cap, the faction meeting the
// min rating, holding Planetary Government when the def is P-flagged (Decision 7), and
// affordable at the surcharged price. Name-sorted for a deterministic candidate order
// (mirrors otherFactions / the buy-asset candidate convention).
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	worldID := ctx.Asset.Location.WorldID
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category || def.TechLevel > params.MaxTechLevel {
			continue
		}
		if statScore(ctx.Faction, def.Category) < def.MinRating {
			continue
		}
		if def.HasFlag(domain.FlagPermission) && !holdsGovernment(ctx, worldID) {
			continue
		}
		if ctx.Faction.Coin < surchargedCost(def.Cost, params.CostPercent) {
			continue
		}
		eligible = append(eligible, def)
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].Name < eligible[j].Name })
	return eligible
}

// holdsGovernment guards the (possibly nil) world engine before delegating to
// domain.HasPlanetaryGovernment (Decision 8). Pretech Logistics waives the tech-level
// requirement but not the P-flag permission requirement, so a P-flagged def is only
// eligible when the faction owns a base on the world.
func holdsGovernment(ctx StepContext, worldID string) bool {
	if ctx.WorldEngine == nil || ctx.WorldEngine.Index == nil {
		return false
	}
	return domain.HasPlanetaryGovernment(ctx.Faction.ID, ctx.WorldEngine.Index.BasesByLocation[worldID])
}

// surchargedCost applies the enabler's cost surcharge: base * percent / 100, rounded up.
// 150 ("half again") on a 5-cost asset yields 8.
func surchargedCost(base, percent int) int {
	return (base*percent + 99) / 100
}
```
Re-ground at execution: confirm `action.ErrNoSelection` exists (used by `relocate`), that
`statScore` is package-local in `stage.go`, and that Commit 1 has landed
`domain.NewPurchasedAsset` and `domain.HasPlanetaryGovernment` (this handler depends on
both).

##### Task 5 — `internal/faction/engine/action/actions/ability/dispatch.go`

Register the handler.

**Find:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
	domain.EffectRelocate:         relocate,
}
```
**Replace with:**
```go
var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
	domain.EffectRelocate:         relocate,
	domain.EffectPurchaseEnabler:  purchaseEnabler,
}
```

##### Task 6 — non-TUI collector implementations

The `Collector` interface gained a method; every implementation must satisfy it.
- `internal/faction/engine/action/actions/mocks/mock_collector.go` — regenerate
  (`go generate ./internal/faction/engine/action/...`).
- `internal/faction/engine/testharness/collector.go` — add a scripted hook:
  ```go
	SelectPurchaseEnablerAssetFn func(string, []*domain.AssetDefinition) (*domain.AssetDefinition, error)
  ```
  and the method:
  ```go
  func (c *ScriptedCollector) SelectPurchaseEnablerAsset(world string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
  	if c.SelectPurchaseEnablerAssetFn != nil {
  		return c.SelectPurchaseEnablerAssetFn(world, candidates)
  	}
  	return nil, nil
  }
  ```
- `internal/faction/tui/adapter/smoke.go` — stub:
  ```go
  func (SmokeActionCollector) SelectPurchaseEnablerAsset(_ string, _ []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
  	return nil, action.ErrActionUnavailable
  }
  ```
  Re-ground: match the receiver name / return style of the sibling smoke stubs (e.g.
  `SelectRelocateOrders`).

##### Task 7 — TUI purchase-enabler overlay (`internal/faction/tui/adapter/action_collector.go` + a new overlay)

Implement `SelectPurchaseEnablerAsset` on `actionCollector` via a new ask channel +
payload, and add the overlay view. Model the def picker on the Buy Asset overlay
(`buyasset.go` or equivalent) — a single `Select` over the candidate definitions —
and **show the surcharged cost** (`ceil(def.Cost * 150%)`) so the GM sees the premium at
selection, not the base price. The adapter method shape:
```go
func (a *actionCollector) SelectPurchaseEnablerAsset(world string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
	raw, err := a.adapter.ask(AskSelectPurchaseEnablerAsset, nil, SelectPurchaseEnablerAssetPayload{
		World:      world,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}
	def, _ := raw.(*domain.AssetDefinition)
	return def, nil
}
```
Add `AskSelectPurchaseEnablerAsset` to the ask-message enum, `SelectPurchaseEnablerAssetPayload`,
and the overlay registration. This is UI work — Claude implements it (Robert does not take
UI slices); the contract is fixed (world + candidate defs in, one def out). The payload may
carry an adapter-built display map (cost/name) matching sibling world/buy payloads (cf.
Effort 3 DR-5). Re-ground against the live overlay/registry patterns at execution.

##### Task 8 — `rulebooks/swn/assets/force_assets.toml`

Add a flat `[ability]` block to Pretech Logistics, immediately after its `description`
line (`force_assets.toml:298`).

**Find:**
```toml
description = "Assets representing caches, smugglers, or internal research and salvage programs. As an action, allows the owner to buy one Force asset on that world that requires up to tech level 5 to purchase, at half again as many Coin as usual, rounded up. Only one asset can be purchased per turn."
```
**Replace with:**
```toml
description = "Assets representing caches, smugglers, or internal research and salvage programs. As an action, allows the owner to buy one Force asset on that world that requires up to tech level 5 to purchase, at half again as many Coin as usual, rounded up. Only one asset can be purchased per turn."

[assets.PretechLogistics.ability]
effect = "purchase_enabler"
[assets.PretechLogistics.ability.purchase]
category = "Force"
max_tech_level = 5
cost_percent = 150
```

##### Task 9 — `internal/faction/engine/action/actions/ability/purchase_enabler_test.go` (new) — full contents:

Hand-build a `StepContext` with a small `rulebook.Rulebook` and a `ScriptedCollector`
returning the chosen def. Proves candidate filtering, the surcharge math (round-up),
affordability, and the emitted mutation pair.

```go
package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func enablerStep() domain.AbilityStep {
	return domain.AbilityStep{
		Effect:   domain.EffectPurchaseEnabler,
		Target:   domain.TargetNone,
		Purchase: &domain.PurchaseParams{Category: domain.StatForce, MaxTechLevel: 5, CostPercent: 150},
	}
}

func enablerRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"F-cheap":   {ID: "F-cheap", Name: "Aaa Cheap Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 3},
		"F-pricey":  {ID: "F-pricey", Name: "Bbb Pricey Force", Category: domain.StatForce, MinRating: 1, Cost: 5, HP: 8, TechLevel: 5},
		"F-toohigh": {ID: "F-toohigh", Name: "Ccc High TL Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 6},
		"F-rating":  {ID: "F-rating", Name: "Ddd High Rating Force", Category: domain.StatForce, MinRating: 9, Cost: 4, HP: 6, TechLevel: 2},
		"W-wealth":  {ID: "W-wealth", Name: "Eee Wealth", Category: domain.StatWealth, MinRating: 1, Cost: 4, HP: 6, TechLevel: 2},
		"F-perm":    {ID: "F-perm", Name: "Fff Permission Force", Category: domain.StatForce, MinRating: 1, Cost: 4, HP: 6, TechLevel: 4, Flags: []domain.AssetFlag{domain.FlagPermission}},
	}}
}

func TestPurchaseEnabler_BuysSurcharged(t *testing.T) {
	var offered []*domain.AssetDefinition
	collector := &testharness.ScriptedCollector{
		SelectPurchaseEnablerAssetFn: func(world string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
			offered = candidates
			for _, def := range candidates {
				if def.ID == "F-pricey" {
					return def, nil
				}
			}
			return nil, nil
		},
	}
	ctx := StepContext{
		Faction:   &domain.Faction{ID: "f1", Force: 5, Coin: 20},
		Asset:     &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:      enablerStep(),
		Rulebook:  enablerRulebook(),
		Collector: collector,
	}
	muts, err := purchaseEnabler(ctx)
	if err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}

	// candidate filter: wrong category, too-high TL, too-high rating, and the P-flagged
	// def (no WorldEngine, so no Planetary Government) all excluded.
	if len(offered) != 2 {
		t.Fatalf("offered %d candidates, want 2 (F-cheap, F-pricey); got %v", len(offered), offered)
	}

	var added *domain.AssetAdded
	var coin *domain.CoinDelta
	for _, m := range muts {
		switch v := m.(type) {
		case domain.AssetAdded:
			added = &v
		case domain.CoinDelta:
			coin = &v
		default:
			t.Errorf("unexpected mutation %T", m)
		}
	}
	if added == nil || coin == nil {
		t.Fatalf("want AssetAdded + CoinDelta, got %v", muts)
	}
	if added.Asset.DefinitionID != "F-pricey" || added.Asset.Location.WorldID != "w1" || added.Asset.Ready {
		t.Errorf("new asset = %+v, want F-pricey on w1, Ready=false", added.Asset)
	}
	if coin.Delta != -8 { // ceil(5 * 1.5) = 8
		t.Errorf("CoinDelta.Delta = %d, want -8", coin.Delta)
	}
}

func TestPurchaseEnabler_InsufficientCoin(t *testing.T) {
	collector := &testharness.ScriptedCollector{
		SelectPurchaseEnablerAssetFn: func(_ string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
			return candidates[0], nil
		},
	}
	ctx := StepContext{
		Faction:   &domain.Faction{ID: "f1", Force: 5, Coin: 5}, // F-pricey surcharge is 8
		Asset:     &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:      enablerStep(),
		Rulebook:  enablerRulebook(),
		Collector: collector,
	}
	// With Coin 5, only F-cheap (surcharge 6 > 5) and F-pricey (8 > 5) — both unaffordable.
	if _, err := purchaseEnabler(ctx); err == nil {
		t.Fatal("want error on no affordable candidates / insufficient Coin")
	}
}

func TestPurchaseEnabler_PermissionGate(t *testing.T) {
	var offered []*domain.AssetDefinition
	collector := &testharness.ScriptedCollector{
		SelectPurchaseEnablerAssetFn: func(_ string, candidates []*domain.AssetDefinition) (*domain.AssetDefinition, error) {
			offered = candidates
			return candidates[0], nil
		},
	}
	// A base owned by the acting faction on w1 = Planetary Government, unlocking the
	// P-flagged F-perm. Index/BasesByLocation are exported, so no constructor is needed.
	ctx := StepContext{
		Faction:     &domain.Faction{ID: "f1", Force: 5, Coin: 20},
		Asset:       &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:        enablerStep(),
		Rulebook:    enablerRulebook(),
		Collector:   collector,
		WorldEngine: &world.WorldEngine{Index: &world.Index{BasesByLocation: map[string][]*domain.Base{"w1": {{OwnerID: "f1"}}}}},
	}
	if _, err := purchaseEnabler(ctx); err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}
	// F-cheap, F-pricey, and now F-perm (government held) — 3 candidates.
	if len(offered) != 3 {
		t.Fatalf("with Planetary Government, offered %d candidates, want 3 (incl F-perm); got %v", len(offered), offered)
	}

	// A base owned by another faction does not unlock F-perm.
	ctx.WorldEngine = &world.WorldEngine{Index: &world.Index{BasesByLocation: map[string][]*domain.Base{"w1": {{OwnerID: "rival"}}}}}
	if _, err := purchaseEnabler(ctx); err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}
	if len(offered) != 2 {
		t.Fatalf("without Planetary Government, offered %d candidates, want 2; got %v", len(offered), offered)
	}
}
```
Re-ground at execution: confirm `ScriptedCollector`'s field/method naming convention and
the `domain.StatForce` constant name; adjust the affordability subtest if the candidate
filter already excludes everything (then it errors at "no candidates" rather than the
post-selection Coin check — either path asserts an error). For the permission-gate subtest,
confirm the exported field names used in the literal: `world.WorldEngine.Index`,
`world.Index.BasesByLocation` (`world.go:23-31`), `domain.Base.OwnerID`, and
`AssetDefinition.Flags []domain.AssetFlag` / `domain.FlagPermission` (`asset.go:24,179`).

##### Task 10 — `internal/faction/rulebook/rulebook_test.go`

Add a parse assertion guarding the authoring. After the existing relocate/Seductress
subtests, **add:**
```go
	t.Run("pretech logistics purchase_enabler parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-F5-002"]
		if !ok {
			t.Fatal("asset SWN-F5-002 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectPurchaseEnabler {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectPurchaseEnabler)
		}
		if step.Purchase == nil {
			t.Fatal("purchase params not parsed")
		}
		if step.Purchase.Category != domain.StatForce || step.Purchase.MaxTechLevel != 5 || step.Purchase.CostPercent != 150 {
			t.Errorf("purchase = %+v, want {Force, 5, 150}", step.Purchase)
		}
	})
```

##### Commit messages

Commit 1:
```
refactor: graduate purchased-asset construction + planetary-government check to domain

- domain.NewPurchasedAsset (NextAssetID, full HP, Ready:false, Maintained:true) and
  domain.HasPlanetaryGovernment (P-flag base-ownership gate), the leaves both purchase
  paths share — domain is the only package both actions and ability import
- refactor buy_asset onto both; retire nextAssetSuffix (a NextAssetID duplicate) and the
  local factionHasPlanetaryGovernment; behavior-preserving
```

Commit 2:
```
feat: purchase_enabler ability

- EffectPurchaseEnabler + PurchaseParams (category, max_tech_level, cost_percent) on
  AbilityStep; parser case + [purchase] sub-table on flat block and steps
- purchase_enabler handler: gather category/TL/rating/permission/affordability-filtered
  candidates on the acting asset's world, prompt SelectPurchaseEnablerAsset, emit
  AssetAdded (Ready:false) + surcharged CoinDelta (half again, rounded up); buys off raw
  base cost via domain.NewPurchasedAsset, P-flag gate via domain.HasPlanetaryGovernment
- SelectPurchaseEnablerAsset across all collectors (TUI overlay + scripted/smoke/mock)
- author Pretech Logistics
```
