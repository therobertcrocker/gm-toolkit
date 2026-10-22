# S-flag Rule Extensions (A.001.6) — Effort 3 Plan: A-side abilities

## In Plain Words

- **What:** three assets whose ability a faction spends its action on. Marketers make a
  rival pay half an asset's price, or lose the use of that asset until it pays.
  Seditionists stop one enemy asset on their world from attacking. Monopoly makes a
  rival pay 1 Coin, or lose an asset of its choice there if it cannot pay. The effort
  also fixes an older fault: when one action uses several assets, each asset now sees
  what the earlier ones did.
- **Why:** today an ability can target only a whole faction, not one asset, and
  nothing can ask a faction "pay this, or else". All three assets need one or both.
  The book lets one action use "one or more assets", so a second Monopoly must see the
  Coin the first one took.
- **How it fits:** the second effort to run in A.001.6 (see the
  [overview plan](s-flag-rule-extensions-plan.md)). It extends Effort 1's conditions
  with "disabled" and "cannot attack". Effort 2 (Medical Center) reuses the pay question
  it adds.

## Context / Goal

Build the effort the overview plan scopes as "A-side abilities", on branch
`feature/s-flag-rule-extensions`.

- Discovery: [s-flag-rule-extensions-discovery.md](../discovery/s-flag-rule-extensions-discovery.md)
  — State Held on an Asset, Prompt to a Faction That Is Not Acting; D-12, D-14, D-38.
- Overview plan: [s-flag-rule-extensions-plan.md](s-flag-rule-extensions-plan.md) —
  Decisions 4, 6, 11, 13 and 19 bind here. This plan reshapes Decision 6 (Decision 1
  below); the overview is reconciled to match.
- Effort 1 plan: [completed/s-flag-rule-extensions-effort-1-plan.md](completed/s-flag-rule-extensions-effort-1-plan.md)
  — `Condition`, `AssetConditionApplied` / `Cleared`, `turn.ExpireConditions`.
- Rules source: `docs/rules/swn-assets.md`; the Use Asset Ability text is the book's
  "Use the special abilities of one or more assets".

## Decisions Ratified in Planning

Re-grounding against the code found four gaps (F1–F4); a side agent found a fifth
(F5). Each decision names the gap or the overview open question it closes.

1. **The separate defender test (D-38) lands here, not in Effort 2.** F1: once
   `disabled` is in `Active()`, `filterOpposingDefenders` (`eligibility.go:58`) also
   drops a disabled asset as a target. A new `Asset.Targetable()` holds today's test
   (ready, maintained, alive) and the defender filter reads it. Effort 2 adds "not
   `sheltered`" to `Targetable()`. Cause: the run order puts Effort 3 first (overview
   Decision 1). Reshapes overview Decision 6.
2. **A disabled defender rolls its defense, but deals no counter and fires no
   `own_defense_win` entries.** Robert's call. "Disabled and useless" in the book. The
   Attack action gates both on `defender.Active()`. Its own roll hooks are already quiet,
   because `effects.active()` reads `Active()`.
3. **Use Asset Ability resolves every selected asset against one view of every
   faction.** F5: each asset read the state from before the action, so two Monopolies
   could push a 1-Coin rival to -1. The book's "one or more assets" makes it real.
   Robert's call: fix it here. `UseAssetAbility.Resolve` builds `ability.NewView` once. `Dispatch`
   applies each step's mutations to the view, so a later step and a later asset read
   what changed (Coin, removed assets, new conditions). Rolls still spend budgets on
   the real factions: `passesGate` finds both factions in real state by ID. That also
   removes the cause of `B.011`. Rejected: a per-action Coin ledger (fixes Coin only,
   not a removed asset or a new condition).
4. **An ability asset with no target now is not offered.** Robert's call. F2: an empty
   candidate list made `resolveTarget` return `ErrNoSelection`, which drops the whole
   action. `ability.HasTargets` checks the first step only, because a later step's
   candidates depend on what earlier steps change (Seductress moves first). The D-22
   rule for Stealth, applied to abilities. A step that finds no target later in the
   action fizzles its asset, not the action (Decision Record 1).
5. **The pay prompt is `hooks.PaymentCollector.ConfirmPayment(PaymentOffer) (bool,
   error)`.** Robert's call on the package. F3: `hooks` is the lowest package with a
   collector, so Effort 2's Cat 6 dispatch can call it; `action` could not be called
   from `hooks/dispatch`. Embedded in `action.Collector` and `PhaseCollector`.
   `PaymentOffer{Payer, Asset, Amount, Reason}`; this effort adds the reasons
   `avoid_disable` and `lift_disable`, Effort 2 adds its own. Resolves the overview OQ
   "the pay-prompt method's name and signature".
6. **Asset-target candidates are rival, unstealthed assets on the acting asset's world,
   in faction-ID then asset-ID order, narrowed per effect.** `assetTargetRules` maps an
   effect to a test; Commit 5 adds `factionTargetRules` the same way. An acting asset in
   transit (no world) has no candidates. Rejected: candidate filters in TOML (a new
   part of the step model for two effects).
7. **Marketers' price is the definition's listed `Cost / 2`.** The book's "half the
   asset's purchase cost, rounded down". Not the `ResolveAssetCost` chain, whose
   surcharges apply to a buyer, not to this price. A price of 0 (a 1-Coin asset) does
   nothing and asks nothing. An asset already disabled is not a candidate (overview
   lean). The payment is `Cause: "ability"`, caused by the Marketers' faction, so a
   Bank can absorb it (a loss "imposed by another faction"). Resolves the overview OQ
   "Marketers aimed at an asset already disabled".
8. **The pay-to-lift prompt runs first in the turn-start gate, in its own record.**
   `turn.OfferConditionLifts` asks for each disabled asset of the faction, in ID order.
   A price it cannot pay is not offered. The lifts apply before `ExpireConditions` and
   `FireTurnStart`, so a lifted asset is active and its own `turn_start` entries fire
   this turn. The price is paid from Coin before income (D-12: turn start). Self-caused
   (`Cause: "condition_lifted"`), and no reactors run in bookkeeping.
9. **Seditionists: `block_attack`, cost from `die = "1d4"`.** The pipeline rolls the die
   (`RollResult`). If the faction cannot pay the roll, nothing attaches and nothing is
   charged (Robert's call; resolves the overview OQ). Attaching clears this source's
   `no_attack` from every other asset. An asset that already holds this source's
   attachment is not a candidate.
10. **`no_attack` is read through `domain.AttackBlocked`.** The block holds while the
    source exists, shares the target's world (not empty), and belongs to another
    faction. The source does not need to be active. Robert's calls (two): a capture
    into the Seditionists' own faction lapses the block when read (resolves the
    overview OQ); a disabled or unready source still holds it (the book names only two
    ends). Consequence: a lapsed attachment works again when both conditions hold
    again. The record stays on the target after the Seditionists die, inert. Readers:
    Attack eligibility and Expand Influence's base attackers. Counters are not blocked
    (D-14).
11. **Monopoly: `levy_coin`, targeting a faction.** Candidates are other factions with
    an unstealthed asset on Monopoly's world. The target pays the owner 1 Coin. If it
    cannot pay, it picks an asset to lose from all its assets on the world, stealthed
    ones too (Robert's call), through a new `SelectAssetToLose(faction, …)` that names
    the faction asked. The removal is `Cause: "levy_coin"`. Monopoly keeps its book
    attack and its `A` flag; its description becomes the book's text. Resolves the
    overview OQ "Monopoly's target list; the effect name", and the ability catalog's
    "Monopoly flag-vs-text" reconciliation.
12. **Effect names `disable_asset`, `block_attack`, `levy_coin`.** Verb-object, like
    Effort 1's `block_ability`.
13. **An ability's condition joins an ability strike in the digest, not a combat event.**
    F4: `asset_condition_applied` always opened a `CrossAttack`. `Cause: "ability"` now
    opens a `CrossAbilityStrike`, and that template renders the condition clauses. This
    effort adds no mutation type, so overview Decision 13 needs only wording.
14. **Five commits in three phases.** Commit 1 is the view fix; Commit 2 the asset
    target and the pay prompt, tested with fixture definitions; Commits 3–5 one asset
    each.

## Decision Record — Execution

1. **An asset with no target left fizzles; the action stands.** Commit 2. A side agent
   found that `HasTargets` runs once, at `Inputs`, but with one view per action
   (Decision 3) an earlier asset can use up a later asset's targets (Marketers #1
   disables the rival's only asset on the world; a Monopoly levy removes a faction's
   last one). The empty list went to the collector, Esc gave `ErrNoSelection`, and
   `ActionEngine.Run` dropped every mutation of the action, the earlier assets'
   included. Robert's call: that asset fizzles like a failed gate. `resolveTarget`
   returns the unexported `errNoTargets` for an empty faction or asset list, before
   any prompt; `Dispatch` halts that asset's chain on it, so its earlier steps and
   the earlier assets keep their results. Decision 4 and Commit 2 Tasks 2 and 8 are
   updated to match. Rejected: keep the plan (the action is lost after a rival may
   have answered a pay prompt); defer to Commit 3.
2. **Every `Active()` reader now skips a disabled asset; Robert accepts each one.**
   Commit 3. An audit of the callers found: attacker selection, Expand Influence's
   base attackers, ability candidates, `FireTurnStart`, the new counter gate, and
   `effects.active()` (13 hook effects). Robert's calls (three): a disabled Planetary
   Defenses or Integral Protocols loses its `restrict_attack` defense limit (as an
   unready one does today); a Refit of a disabled asset makes a new asset with no
   conditions, so the refit clears `disabled`; a disabled asset can still be
   repaired and moved (neither reads `Active()` or `Ready`). No code change.
   Rejected: `Targetable()` in `restrict_attack`; carry conditions through a refit;
   gate repair and move.
3. **Test placement follows the files that exist.** Commit 3, Task 11. The rulebook
   tests have no step-validation table, so the `target = "faction"` check is its own
   test (`TestConvertAbilityStep_DisableAssetNeedsAnAssetTarget`, through
   `loadAssetsFromTOML`). There is no `eligibility_test.go`, so the defender-filter
   test is in `attack_test.go`. `TestAsset_Active` became
   `TestAsset_ActiveAndTargetable` (one table, two columns). The hand-off's
   multi-asset check is in `use_asset_ability_test.go`. Task 11 is updated to match.
4. **The Seditionists scenario checks the Attack candidates, not whether Attack is
   offered.** Commit 4, Task 7. The harness gives every asset the zero `RegionHex`, so
   `eligibleDefendersAtHex` puts beta's Security Personnel on Hadrian on one hex with
   the Tartarus assets, and beta's turn offers Attack with the Saboteurs blocked. Beta
   now takes Attack each cycle, the test records the `SelectAttackers` candidates, and
   beta selects none. The test checks that the Saboteurs are absent in cycle 1 and
   present in cycle 2. Beta's second asset is Informers, which have no attack. Task 7
   is updated to match. Rejected: give each asset a `RegionHex` (a harness change for
   one test); remove beta's Security Personnel (hides the reason).
5. **Monopoly's loss pick is required.** Commit 5. A side agent found that the plan's
   nil pick returned `ErrNoSelection`, which drops every mutation of the action, the
   earlier assets' included (the case Decision Record 1 ruled out). A nil pick is
   never a real "no choice": a levy target always has an unstealthed asset on the
   world. Robert's call: the book gives the target no way out of the loss, so the pick
   is required. `levyCoin` returns a plain error for a nil pick, not a sentinel. The
   TUI uses a new `overlay.SelectAssetToLose`, which swallows Esc and shows
   `formHelp` (the `StatRaise` pattern). Tasks 3, 5 and 9 are updated to match.
   Rejected: fizzle this Monopoly (Esc lets the target escape the levy); cancel the
   whole action (the plan; undoes earlier assets after a rival may have paid).

---

## Shared Context

- **Test patterns.** Ability tests: `seqRoller{values}` and gomock `MockCollector`
  (`ability/dispatch_test.go`); action tests: `makeAbilityRulebook` / `makeAbilityState`
  (`use_asset_ability_test.go`). Scenarios: `testharness.NewHarness`, `addActiveAsset`,
  `runOrderedCycle` (`scenarios/capture_test.go`) with `ScriptedCollector`.
- **Mock regeneration.** After a change to `action.Collector`, run
  `go generate ./internal/faction/engine/action/...`.
- **Every commit ends green:** `go build ./... && go test ./internal/faction/...`.
- **A new `AskKind`** goes before `AskKindCount` and gets an `askEntries` row;
  `TestAskEntriesCoverEveryKind` catches a missing row.
- **Causes.** Ability mutations use `Cause: "ability"` (the digest's ability-strike
  path), except the Monopoly removal (`"levy_coin"`) and the lift (`"condition_lifted"`).

## Out of Scope

- `sheltered` and "not `sheltered`" in `Targetable()` (Effort 2).
- The other uses of the pay prompt: Medical Center (Effort 2).
- Clearing an inert `no_attack` record when its Seditionists die (Decision 10); a
  candidate cleanup if saves grow noisy.
- A digest line for a paid lift: the lift reads only as Coin spent in bookkeeping.
- Position changes across assets in one action: relocate reads the world index, not
  the view (Decision 3 covers faction state only).

---

## Work Breakdown

### Phase 1 — One view per action, asset targets and the pay prompt

#### Commit 1 — `fix: abilities in one action see earlier assets' results`

##### Task 1 — `internal/faction/engine/action/actions/ability/dispatch.go`

Add below `EffectHandler`:
```go
// NewView copies every faction, so Dispatch can apply each step's mutations and a
// later step, or a later asset in the same action, reads what changed. Rolls spend
// budgets on the real factions, not on the view (Effort 3 Decision 3).
func NewView(factionState *state.FactionState) *state.FactionState {
	view := &state.FactionState{Factions: map[string]*domain.Faction{}}
	if factionState == nil {
		return view
	}
	*view = *factionState
	view.Factions = make(map[string]*domain.Faction, len(factionState.Factions))
	for id, faction := range factionState.Factions {
		view.Factions[id] = faction.Clone()
	}
	return view
}
```
Replace `Dispatch` from its signature to the end of the function:
```go
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	registry *hooks.Registry,
	factionState *state.FactionState,
	view *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	// Pre-flight: if any step's effect is not yet built, the whole ability is a
	// not-yet-built no-op. Bail before executing any step, so a built early step
	// never half-applies (moves an asset, spends Coin) only to be discarded when a
	// later unbuilt step short-circuits to confirmApplied.
	for _, step := range def.Ability.Steps {
		if _, ok := handlers[step.Effect]; !ok {
			return confirmApplied(asset, def, collector)
		}
	}
	workingFaction := view.Factions[faction.ID]
	if workingFaction == nil {
		return nil, fmt.Errorf("ability: acting faction %s not in the view", faction.ID)
	}
	workingAsset := workingFaction.Assets[asset.ID]
	if workingAsset == nil {
		workingAsset = asset.Clone()
	}
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)

	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		handler := handlers[step.Effect]
		target, err := resolveTarget(step, workingFaction, workingAsset, view, collector)
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
		if updated, ok := workingFaction.Assets[asset.ID]; ok {
			workingAsset = updated
		}
	}
	return mutations, nil
}
```
Add `"fmt"` to the imports.

##### Task 2 — `internal/faction/engine/action/actions/ability/stage.go`

In `passesGate`, **find:**
```go
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
```
**Replace with:**
```go
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	// deliberate: the real target faction, found by ID, not target.Faction, because the target stage reads the action's view and a defense hook's budget spent on a copy would be lost (Effort 3 Decision 3).
	defender := factionState.Factions[target.Faction.ID]
	if defender == nil {
		return false, fmt.Errorf("contest: target faction %s not in state", target.Faction.ID)
	}
```
In both `RollWithHooks` calls below it, replace every `target.Faction` with `defender`
(the attack roll's `Opponent`; the defense roll's `Actor`, its `statScore` and its
budget-holder argument).

##### Task 3 — `internal/faction/engine/action/actions/use_asset_ability.go`

Replace `Resolve`:
```go
func (u *UseAssetAbility) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// deliberate: one view for the whole action, not one per asset, because the book lets one action use several assets and each must see what the earlier ones did (Effort 3 Decision 3).
	view := ability.NewView(factionState)
	for _, asset := range u.selectedAssets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return fmt.Errorf("use asset ability: definition not found: %s", asset.DefinitionID)
		}
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, u.registry, factionState, view, rulebook, u.world)
		if err != nil {
			return fmt.Errorf("use asset ability: %w", err)
		}
		u.mutations = append(u.mutations, mutations...)
	}
	return nil
}
```

##### Task 4 — `internal/faction/engine/action/actions/ability/dispatch_test.go`

At every `Dispatch(` call, insert `NewView(factionState)` after the `factionState`
argument. A test that passes a state without the acting faction now gets
the "not in the view" error; add the faction to its state.

##### Task 5 — `docs/architecture/engine/actions.md`

In "Eligibility helpers and the ability fold", **find:**
```
gate** so a missed contest fizzles the rest. Across steps the dispatcher threads a
**working clone** of the acting faction, re-projecting each step's own mutations
onto the clone via the same `mutation.Apply` the real `Output()` runs — so a later
step reads what an earlier one changed without a parallel projection that could
drift.
```
**Replace with:**
```
gate** so a missed contest fizzles the rest. The action builds one **view** of every
faction (`ability.NewView`) and every `Dispatch` re-projects each step's mutations
onto it via the same `mutation.Apply` the real `Output()` runs — so a later step, or a
later asset in the same action, reads what an earlier one changed without a parallel
projection that could drift. Rolls still spend hook budgets on the real factions.
```
In Key Decisions, **find:**
```
  failed gate, and a working clone of the acting faction carries each step's
  mutations forward so a later step reads what an earlier one changed. Abilities
```
**Replace with:**
```
  failed gate, and one view of every faction, built once per action, carries each
  step's mutations forward, so a later step or a later asset reads what an earlier
  one changed (A.001.6 Effort 3, Decision 3). Abilities
```

##### Task 6 — tests

- `use_asset_ability_test.go`: `TestUseAssetAbility_LaterAssetSeesEarlierCoin` — a
  definition `"postech"` (A-flagged, one step `coin_delta`, `Roll{Die: 1d6, Outcomes:
  [{From: 1, To: 1, Coin: -1}]}`); `f1` has `Coin: 1` and two active `postech` assets
  `p1`, `p2`; `SelectAbilityAssets` returns both; roller always 1. Output is exactly
  `[CoinDelta{f1, -1}, AssetRemoved{f1, p2, Cause: "ability"}]` (before the fix: two
  `CoinDelta`s).
- `ability/dispatch_test.go`: `TestNewView_CopiesEveryFaction` — change the view's
  `f2.Coin` and an asset's `CurrentHP`; the real state is unchanged.

##### Commit message

```
fix: abilities in one action see earlier assets' results

- ability.NewView copies every faction once per Use Asset Ability
- Dispatch applies each step to the view; later steps and assets read it
- passesGate rolls with the real factions, found by ID, so budgets persist
- actions.md: the view replaces the acting-faction working clone
```

#### Commit 2 — `feat: asset targets and the pay prompt`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
	// deliberate: TargetAsset/TargetWorld are dormant scalar-target vocab for A.001.3
	// (Marketers/Seditionists), not dead — don't drop without that effort.
	TargetAsset TargetKind = "asset"
	TargetWorld TargetKind = "world"
```
**Replace with:**
```go
	TargetAsset   TargetKind = "asset"
	// deliberate: TargetWorld is dormant vocab, not dead — no asset targets a world yet.
	TargetWorld TargetKind = "world"
```
**Find:**
```go
	Factions []*Faction
	// deliberate: Assets/Worlds are dormant for A.001.3 scalar targets, not dead —
	// don't drop without that effort. See TargetAsset/TargetWorld.
	Assets []*Asset
	Worlds []string
```
**Replace with:**
```go
	Factions []*Faction
	Assets   []*Asset
	// deliberate: Worlds is dormant, not dead. See TargetWorld.
	Worlds []string
```

##### Task 2 — `internal/faction/engine/action/actions/ability/stage.go`

Add `"maps"` and `"slices"` to the imports. In `resolveTarget`, **find:**
```go
	default:
		// deliberate: asset/world targets await A.001.3 (Marketers/Seditionists),
		// not unbuilt-by-oversight — see domain.TargetAsset/TargetWorld.
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
```
**Replace with:**
```go
	case domain.TargetAsset:
		candidates := assetCandidates(step.Effect, faction.ID, asset, factionState)
		if len(candidates) == 0 {
			return domain.AbilityTarget{}, errNoTargets
		}
		target, err := collector.SelectAbilityTarget(asset, step, domain.AbilityTargetCandidates{Assets: candidates})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.Asset == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		owner := factionState.Factions[target.Asset.OwnerID]
		if owner == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: target owner %s not in state", step.Effect, target.Asset.OwnerID)
		}
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: target.Asset, Faction: owner}, nil
	default:
		// deliberate: world targets are not unbuilt-by-oversight — no asset targets a world yet. See domain.TargetWorld.
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
```
In the `TargetFaction` case, return `errNoTargets` the same way when `otherFactions`
is empty. Add `"errors"` to the imports and, above `resolveTarget`:
```go
// errNoTargets: the step has no candidate left, because an earlier step or asset in
// the same action changed the view. Dispatch halts the asset's chain on it.
var errNoTargets = errors.New("ability: no target left")
```
In `dispatch.go`, add `"errors"` to the imports; after the `resolveTarget` call, before
`if err != nil {`:
```go
		// deliberate: fizzle this asset, not fail the action, because an earlier asset in the same action may have used up the targets and its result must stand (Effort 3 Decision Record 1).
		if errors.Is(err, errNoTargets) {
			break
		}
```
Append to `stage.go`:
```go
// assetTargetRules narrows a step's asset candidates by its effect. An effect with
// no rule takes every base candidate.
var assetTargetRules = map[domain.AbilityEffectType]func(candidate, acting *domain.Asset) bool{}

// assetCandidates returns other factions' unstealthed assets on the acting asset's
// world, in faction-ID then asset-ID order, that pass the effect's rule. An asset
// in transit has no world and no candidates.
func assetCandidates(effect domain.AbilityEffectType, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) []*domain.Asset {
	world := acting.Location.WorldID
	if world == "" {
		return nil
	}
	rule := assetTargetRules[effect]
	var candidates []*domain.Asset
	for _, id := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if id == actingFactionID {
			continue
		}
		for _, candidate := range domain.SortedAssets(factionState.Factions[id]) {
			if candidate.Location.WorldID != world || candidate.Stealthy {
				continue
			}
			if rule == nil || rule(candidate, acting) {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

// HasTargets reports whether the first step of def's ability can pick a target
// now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, factionState *state.FactionState) bool {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	step := def.Ability.Steps[0]
	switch step.Target {
	case domain.TargetAsset:
		return len(assetCandidates(step.Effect, faction.ID, asset, factionState)) > 0
	case domain.TargetFaction:
		return len(otherFactions(factionState, faction.ID)) > 0
	}
	return true
}
```

##### Task 3 — `internal/faction/engine/action/actions/use_asset_ability.go`

Replace `Validate`, `Inputs` and `abilityCandidates`:
```go
func (u *UseAssetAbility) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	return len(abilityCandidates(faction, factionState, rulebook)) > 0
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := u.collector.SelectAbilityAssets(faction, abilityCandidates(faction, factionState, rulebook), rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selectedAssets = selected
	return nil
}
```
```go
// abilityCandidates returns the faction's assets that may use their ability now,
// in ID order. An asset whose ability has no target now is not offered (Effort 3
// Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []*domain.Asset {
	var candidates []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, asset, def, factionState) {
			candidates = append(candidates, asset)
		}
	}
	return candidates
}
```
`TestUseAssetAbility_Validate`, "usable A-flagged asset": it passes a nil state and
reaches `HasTargets`. Pass
`&state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": {ID: "f2"}}}`.
The other subtests stop before `HasTargets` and keep `nil`.

##### Task 4 — `internal/faction/engine/hooks/collector.go`

Add `import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"` and append:
```go
// PaymentReason says what a payment buys, so a prompt can word it.
type PaymentReason string

const (
	// PaymentAvoidDisable: pay now, or the asset is disabled (Marketers).
	PaymentAvoidDisable PaymentReason = "avoid_disable"
	// PaymentLiftDisable: pay to lift a disabled condition at the owner's turn start.
	PaymentLiftDisable PaymentReason = "lift_disable"
)

// PaymentOffer asks Payer whether to pay Amount Coin for Reason on Asset.
type PaymentOffer struct {
	Payer  *domain.Faction
	Asset  *domain.Asset
	Amount int
	Reason PaymentReason
}

// PaymentCollector asks one faction whether to pay one amount. action.Collector and
// the engine's PhaseCollector both embed it, so an action, the bookkeeping step and
// a hook dispatcher all ask the same question (A.001.6 overview Decision 19).
type PaymentCollector interface {
	ConfirmPayment(offer PaymentOffer) (bool, error)
}
```

##### Task 5 — `internal/faction/engine/action/collector.go` and `internal/faction/engine/collector.go`

`action/collector.go` — in `Collector`, after `hooks.Collector` add `hooks.PaymentCollector`.
Run `go generate ./internal/faction/engine/action/...`.

`engine/collector.go` — add the `hooks` import; in `PhaseCollector`, before
`AwaitCheckpoint`, add `hooks.PaymentCollector`. Append to the doc comment's first
sentence: `, and the pay-to-lift question at turn start`.

##### Task 6 — prompt surfaces

`tui/adapter/channels.go` — after `AskConfirmCounterstrike` add
`AskSelectAbilityAssetTarget` and `AskConfirmPayment`; after
`ConfirmCounterstrikePayload` add:
```go
type SelectAbilityAssetTargetPayload struct {
	Asset      *domain.Asset
	Effect     domain.AbilityEffectType
	Candidates []*domain.Asset
	OwnerNames map[string]string // ownerID -> faction name (candidates are rivals')
}
type ConfirmPaymentPayload struct{ Offer hooks.PaymentOffer }
```

`tui/adapter/action_collector.go` — in `SelectAbilityTarget`, before `default:` add:
```go
	case domain.TargetAsset:
		raw, err := a.adapter.ask(AskSelectAbilityAssetTarget, nil, SelectAbilityAssetTargetPayload{
			Asset:      asset,
			Effect:     step.Effect,
			Candidates: candidates.Assets,
			OwnerNames: a.adapter.ownerNames(candidates.Assets),
		})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		chosen, _ := raw.(*domain.Asset)
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: chosen}, nil
```
and after `ConfirmCounterstrike` add:
```go
func (a *actionCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return a.adapter.confirmPayment(offer)
}

// confirmPayment asks the paying faction. The action collector (Marketers) and the
// phase collector (the turn-start lift) share it.
func (a *Adapter) confirmPayment(offer hooks.PaymentOffer) (bool, error) {
	raw, err := a.ask(AskConfirmPayment, offer.Payer, ConfirmPaymentPayload{Offer: offer})
	if err != nil {
		return false, err
	}
	paid, _ := raw.(bool)
	return paid, nil
}
```
`tui/adapter/phase_collector.go` — add:
```go
func (p *phaseCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	return p.adapter.confirmPayment(offer)
}
```
(add the `hooks` import).

`tui/views/turn/overlay/ability.go` — add the `hooks` import and append:
```go
// NewSelectAbilityAssetTarget picks the rival asset an ability acts on (Marketers,
// Seditionists). Labels carry the owner. Answer: *domain.Asset. Esc cancels.
func NewSelectAbilityAssetTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Asset, ownerNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(candidates))
	for _, candidate := range candidates {
		owner := ownerNames[candidate.OwnerID]
		if owner == "" {
			owner = candidate.OwnerID
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s · %s", assetLabel(candidate, rb), owner), candidate))
	}
	return NewSelect[*domain.Asset]("Target which asset?", fmt.Sprintf("%s · %s", assetName(asset, rb), effect), opts)
}

// NewConfirmPayment asks the paying faction whether to pay. The header names the
// faction asked, because it is often not the acting one. Answer: bool.
func NewConfirmPayment(offer hooks.PaymentOffer, rb *rulebook.Rulebook) ConfirmOverlay {
	name := assetName(offer.Asset, rb)
	header := fmt.Sprintf("Asked of %s", offer.Payer.Name)
	if offer.Reason == hooks.PaymentLiftDisable {
		return NewConfirm(fmt.Sprintf("Pay %d Coin to bring %s back into use?", offer.Amount, name), header, "Pay", "Leave it disabled")
	}
	return NewConfirm(fmt.Sprintf("Pay %d Coin, or %s is disabled?", offer.Amount, name), header, "Pay", "Take the disable")
}
```
`tui/views/turn/execution/execution.go` — after the `AskConfirmCounterstrike` row add:
```go
	adapter.AskSelectAbilityAssetTarget: {"Ability Target", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectAbilityAssetTargetPayload)
		return overlay.NewSelectAbilityAssetTarget(p.Asset, p.Effect, p.Candidates, p.OwnerNames, m.rulebook)
	}},
	adapter.AskConfirmPayment: {"Pay?", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmPayment(msg.Payload.(adapter.ConfirmPaymentPayload).Offer, m.rulebook)
	}},
```
`tui/adapter/smoke.go`:
```go
func (s *SmokePhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil // the dry run declines every lift
}
```
```go
func (SmokeActionCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
`engine/testharness/collector.go` — field
`ConfirmPaymentFn func(hooks.PaymentOffer) (bool, error)` and:
```go
func (c *ScriptedCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	if c.ConfirmPaymentFn != nil {
		return c.ConfirmPaymentFn(offer)
	}
	return false, nil
}
```
`engine/orchestrator_test.go` — `minimalPhaseCollector`:
```go
func (c *minimalPhaseCollector) ConfirmPayment(_ hooks.PaymentOffer) (bool, error) {
	return false, nil
}
```
(add the `hooks` import).

##### Task 7 — `docs/architecture/engine/hooks.md` and `orchestrator.md`

`hooks.md` — after the paragraph that ends "`SelectModifiers` is not called when no offer
survives the budget filter.", add:
```
Beside it sits `PaymentCollector` (`ConfirmPayment(PaymentOffer) (bool, error)`):
one faction is asked whether to pay one amount. It lives here, the lowest package
with a collector, so a hook dispatcher can ask it as well as an action or the
bookkeeping step. `action.Collector` and `PhaseCollector` both embed it.
```
`orchestrator.md` — **find** `` `PhaseCollector` is the orchestrator's own prompt surface: `AwaitCheckpoint`, ``
**replace with** `` `PhaseCollector` is the orchestrator's own prompt surface: `ConfirmPayment` (embedded `hooks.PaymentCollector`; the turn-start lift), `AwaitCheckpoint`, ``.

##### Task 8 — tests

- `ability/stage_test.go` (new): `TestAssetCandidates` — `f1` acts from Anchorage;
  `f2` holds `b2` (Anchorage), `b1` (Anchorage, stealthed), `b3` (Hadrian); `f3` holds
  `c1` (Anchorage); `f1` holds `a2` (Anchorage). Result `[b2, c1]`. An acting asset with
  empty `WorldID` → nil. A rule in `assetTargetRules` (set and reset inside the test)
  that rejects `c1` → `[b2]`.
- `TestResolveTarget_Asset` — the collector returns `b2`; the target has `Kind: asset`,
  `Asset: b2`, `Faction` = `f2` from the state. The collector returns a nil asset →
  `ErrNoSelection`. No candidate → `errNoTargets`, and the collector is not called.
- `dispatch_test.go`: `TestDispatch_NoTargetsLeft_Fizzles` — step 1 `coin_delta` (+1),
  step 2 targets an asset with no rival asset on the world → no error, no prompt,
  exactly step 1's `CoinDelta`.
- `TestHasTargets` — a `TargetAsset` first step with no candidate → false; with one →
  true; `TargetFaction` with one rival → true; a relocate-first ability → true.
- `use_asset_ability_test.go`: `TestUseAssetAbility_Validate` adds a case — an asset
  whose fixture ability targets an asset, with no rival asset on its world → false.
- `execution_test.go`: `TestAskEntriesCoverEveryKind` passes with the two kinds.

##### Commit message

```
feat: asset targets and the pay prompt

- resolveTarget handles TargetAsset: rival, unstealthed, same world, ID order
- assetTargetRules narrow candidates per effect; target faction is the owner
- an ability asset with no target now is not offered (HasTargets)
- an asset whose targets an earlier asset used up fizzles, not the action
- hooks.PaymentCollector.ConfirmPayment, embedded in both collectors
- adapter, overlays, smoke, scripted and test collectors for both prompts
```

### Phase 2 — Marketers

#### Commit 3 — `feat: Marketers`

##### Task 1 — `internal/faction/domain/asset.go`

**Find** `const ConditionNoAbility ConditionKind = "no_ability"` and **replace with:**
```go
const (
	ConditionNoAbility ConditionKind = "no_ability"
	ConditionDisabled  ConditionKind = "disabled"
)
```
In the `AbilityEffectType` block, after `EffectPurchaseEnabler` add
`EffectDisableAsset AbilityEffectType = "disable_asset"`.

Replace `Active`:
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

##### Task 2 — `internal/faction/engine/action/actions/eligibility.go`

Replace `filterOpposingDefenders` and its doc comment:
```go
// filterOpposingDefenders applies the defender eligibility filter to a candidate
// slice: excludes the attacker's own assets and assets that are stealthy or not
// targetable.
func filterOpposingDefenders(attackerFactionID string, candidates []*domain.Asset) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range candidates {
		// deliberate: Targetable, not Active, because a disabled asset can still be attacked (D-38).
		if asset.OwnerID != attackerFactionID && !asset.Stealthy && asset.Targetable() {
			result = append(result, asset)
		}
	}
	return result
}
```

##### Task 3 — `internal/faction/engine/action/actions/attack.go`

**Find:**
```go
		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenderWins {
```
**Replace with:**
```go
		// deliberate: a disabled defender rolls its defense but deals no counter and fires no own_defense_win entry, because the book makes it useless (Effort 3 Decision 2).
		defenderActive := defender.Active()

		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenderWins && defenderActive {
```
**Find** `		if defenderWins && !attackerWins {` and **replace with**
`		if defenderWins && !attackerWins && defenderActive {`.

##### Task 4 — `internal/faction/engine/action/actions/ability/disable_asset.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// disableAsset is Marketers: the target faction pays half the target asset's
// listed cost, rounded down, or the asset is disabled until that price is paid.
// A faction that cannot pay takes the condition with no prompt.
func disableAsset(ctx StepContext) ([]domain.Mutation, error) {
	targetFaction, target := ctx.Target.Faction, ctx.Target.Asset
	def, ok := ctx.Rulebook.Assets[target.DefinitionID]
	if !ok {
		return nil, fmt.Errorf("disable_asset: definition not found: %s", target.DefinitionID)
	}
	price := def.Cost / 2
	if price == 0 {
		// deliberate: a zero price does nothing, not a prompt, because paying 0 is always possible (Effort 3 Decision 7).
		return nil, nil
	}
	if targetFaction.Coin >= price {
		paid, err := ctx.Collector.ConfirmPayment(hooks.PaymentOffer{
			Payer:  targetFaction,
			Asset:  target,
			Amount: price,
			Reason: hooks.PaymentAvoidDisable,
		})
		if err != nil {
			return nil, fmt.Errorf("disable_asset: %w", err)
		}
		if paid {
			return []domain.Mutation{domain.CoinDelta{
				FactionID:         targetFaction.ID,
				Delta:             -price,
				Cause:             "ability",
				CausedByFactionID: ctx.Faction.ID,
			}}, nil
		}
	}
	return []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         targetFaction.ID,
		AssetID:           target.ID,
		Condition:         domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: ctx.Asset.ID},
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
```

##### Task 5 — `internal/faction/engine/action/actions/ability/dispatch.go`, `stage.go`

`dispatch.go` — in `handlers` add `domain.EffectDisableAsset: disableAsset,`.

`stage.go` — replace the empty `assetTargetRules` map:
```go
var assetTargetRules = map[domain.AbilityEffectType]func(candidate, acting *domain.Asset) bool{
	domain.EffectDisableAsset: func(candidate, _ *domain.Asset) bool {
		return !candidate.HasCondition(domain.ConditionDisabled)
	},
}
```

##### Task 6 — `internal/faction/rulebook/rulebook.go`

In `toAbilityEffect`, before `default:` add:
```go
	case "disable_asset":
		return domain.EffectDisableAsset, nil
```
In `convertAbilityStep`, after the `reveal_stealth` check add:
```go
	if step.Effect == domain.EffectDisableAsset && (step.Target != domain.TargetAsset || step.Contest == nil) {
		return domain.AbilityStep{}, fmt.Errorf("disable_asset requires target = %q and a contest", "asset")
	}
```

##### Task 7 — `internal/faction/engine/turn/conditions.go`

Add `"fmt"`, `hooks` and `rulebook` to the imports and append:
```go
// OfferConditionLifts asks faction, for each of its disabled assets in ID order,
// whether to pay half the asset's listed cost, rounded down, to lift the
// condition. A price it cannot pay, after earlier lifts, is not offered.
func OfferConditionLifts(faction *domain.Faction, rulebook *rulebook.Rulebook, collector hooks.PaymentCollector) ([]domain.Mutation, error) {
	coin := faction.Coin
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		for _, condition := range asset.Conditions {
			if condition.Kind != domain.ConditionDisabled {
				continue
			}
			def, ok := rulebook.Assets[asset.DefinitionID]
			if !ok {
				return nil, fmt.Errorf("offer condition lifts: definition not found: %s", asset.DefinitionID)
			}
			price := def.Cost / 2
			if price > coin {
				continue
			}
			paid, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: faction, Asset: asset, Amount: price, Reason: hooks.PaymentLiftDisable})
			if err != nil {
				return nil, fmt.Errorf("offer condition lifts: %w", err)
			}
			if !paid {
				continue
			}
			coin -= price
			mutations = append(mutations,
				domain.CoinDelta{FactionID: faction.ID, Delta: -price, Cause: "condition_lifted", CausedByFactionID: faction.ID},
				domain.AssetConditionCleared{FactionID: faction.ID, AssetID: asset.ID, Condition: condition, Cause: "condition_lifted", CausedByFactionID: faction.ID},
			)
		}
	}
	return mutations, nil
}
```

##### Task 8 — `internal/faction/engine/orchestrator.go`

In `runBookkeepingPhase`, **find:**
```go
	var startMutations []domain.Mutation
	// deliberate: fire here, gated on BookkeepingApplied, not in turn setup, because setup re-runs on every resume and would pay a turn-start effect twice.
	if e.Turn.InProgress(factionState) && !factionState.CurrentTurn.BookkeepingApplied {
		startMutations = append(turn.ExpireConditions(factionState, faction.ID), e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)...)
```
**Replace with:**
```go
	var liftMutations, startMutations []domain.Mutation
	// deliberate: fire here, gated on BookkeepingApplied, not in turn setup, because setup re-runs on every resume and would pay a turn-start effect twice.
	if e.Turn.InProgress(factionState) && !factionState.CurrentTurn.BookkeepingApplied {
		var err error
		// deliberate: lifts apply before the turn-start effects, not with them, because a lifted asset is active again and its own turn_start entries fire this turn (Effort 3 Decision 8).
		if liftMutations, err = e.runConditionLifts(faction, factionState, paths, collectors.Phase, observer, phaseLog); err != nil {
			return err
		}
		startMutations = append(turn.ExpireConditions(factionState, faction.ID), e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)...)
```
**Find** `	observer.OnBookkeepingApplied(faction, bookResult, append(startMutations, bookMutations...))`
and **replace with**
`	observer.OnBookkeepingApplied(faction, bookResult, slices.Concat(liftMutations, startMutations, bookMutations))`.

Add below `runBookkeepingPhase`:
```go
// runConditionLifts offers the faction's disabled assets their price and applies
// the lifts it pays for. A recoverable failure is reported and lifts nothing.
func (e *Engine) runConditionLifts(faction *domain.Faction, factionState *state.FactionState, paths *campaigns.Paths, collector PhaseCollector, observer TurnObserver, phaseLog *slog.Logger) ([]domain.Mutation, error) {
	mutations, err := turn.OfferConditionLifts(faction, e.Rulebook, collector)
	if err == nil && len(mutations) > 0 {
		err = e.applyAndRecord(factionState, faction, mutations, paths, phaseLog)
	}
	if err == nil {
		return mutations, nil
	}
	if !factionErrors.IsRecoverable(err) {
		phaseLog.Error("failure", "err", err)
		observer.OnError(faction, err)
		return nil, err
	}
	phaseLog.Warn("recoverable failure", "err", err)
	observer.OnError(faction, err)
	return nil, nil
}
```

##### Task 9 — `rulebooks/swn/assets/wealth_assets.toml`

Before `[assets.Marketers.attack]` add:
```toml
[assets.Marketers.ability]
attacker_stat = "Cunning"
defender_stat = "Wealth"
effect = "disable_asset"
target = "asset"

```

##### Task 10 — digest

`digest/build.go` — in `case "asset_condition_applied":`, **find:**
```go
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
					if !slices.Contains(ce.Conditions, m.Condition.Kind) {
```
**Replace with:**
```go
					kind := CrossAttack
					if m.Cause == "ability" {
						kind = CrossAbilityStrike
					}
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, kind, factionState)
					if !slices.Contains(ce.Conditions, m.Condition.Kind) {
```
`wire_templates.go` — in `conditionClauses` add
`domain.ConditionDisabled: "{defender} must pay to bring a disabled asset back into use.",`.

`wire_renderer.go` — in the `CrossAttack` arm, **find:**
```go
		for _, kind := range ce.Conditions {
			if clause, ok := conditionClauses[kind]; ok {
				clause = strings.ReplaceAll(clause, "{attacker}", ce.Attacker.Name)
				result += " " + strings.ReplaceAll(clause, "{defender}", ce.Defender.Name)
			}
		}
		if ce.Counterstruck {
```
**Replace with:**
```go
		result += conditionText(ce)
		if ce.Counterstruck {
```
Replace the `CrossAbilityStrike` arm:
```go
	case digest.CrossAbilityStrike:
		tmpl := pick(rng, crossAbilityTemplates)
		result := strings.ReplaceAll(tmpl, "{attacker}", ce.Attacker.Name)
		return strings.ReplaceAll(result, "{defender}", ce.Defender.Name) + conditionText(ce)
```
Append:
```go
// conditionText renders one clause per condition on a cross event, each with a
// leading space, or "" when it carries none.
func conditionText(ce digest.CrossEvent) string {
	var text string
	for _, kind := range ce.Conditions {
		if clause, ok := conditionClauses[kind]; ok {
			clause = strings.ReplaceAll(clause, "{attacker}", ce.Attacker.Name)
			text += " " + strings.ReplaceAll(clause, "{defender}", ce.Defender.Name)
		}
	}
	return text
}
```

##### Task 11 — tests

- `domain/asset_test.go`: `TestAsset_ActiveAndTargetable` (was `TestAsset_Active`)
  — a ready, maintained, alive asset with `disabled` is `Targetable()` and not
  `Active()`; without it, both.
- `actions/attack_test.go`: `TestFilterOpposingDefenders_DisabledStaysTargetable` —
  a disabled rival asset is in `filterOpposingDefenders`; an unready one is not.
- `actions/attack_test.go`: defender `d1` carries `disabled`; the defense wins
  outright → no counter `AssetHPDelta` on the attacker, and a `counterstrike` fixture
  defender with `disabled` gets no `ConfirmCounterstrike` call.
- `actions/use_asset_ability_test.go`: `TestUseAssetAbility_SecondMarketersFizzles`
  — two Marketers, one rival asset on the world. The first wins and the target
  declines → `disabled`. The second has no candidate and fizzles: no prompt, no
  error, and the first one's `AssetConditionApplied` stands (Decision Record 1).
- `ability/disable_asset_test.go` (new), via `Dispatch` with a Marketers fixture
  (cost-10 target, contest won):
  - target faction `Coin: 9`, pays → one `CoinDelta{f2, -5, "ability", f1}`.
  - declines → `AssetConditionApplied{f2, b2, {disabled, a1}}`.
  - `Coin: 4` → no `ConfirmPayment` call; the condition.
  - a cost-1 target → no mutation, no prompt.
  - contest lost → no prompt, no mutation.
  - `TestAssetCandidates` gains: an already-disabled rival asset is not a Marketers
    candidate.
- `rulebook/rulebook_test.go`: replace "marketers has no ability section" with
  "marketers disables an asset" — one step, `EffectDisableAsset`, `TargetAsset`,
  contest Cunning vs Wealth. `TestConvertAbilityStep_DisableAssetNeedsAnAssetTarget`:
  `disable_asset` with `target = "faction"` fails to load.
- `turn/conditions_test.go`: `TestOfferConditionLifts` — `f1` (`Coin: 6`) holds two
  disabled assets of cost 10 (`x1`, `x2`) and one `no_ability` asset; the collector pays
  every offer → `x1` lifted (`CoinDelta -5`, cleared), `x2` not offered (1 Coin left);
  a declined offer emits nothing; a collector error is returned.
- `engine/testharness/scenarios/marketers_test.go` (new):
  `TestMarketersDisableThenLift` — order `[alpha, beta]`; alpha's Marketers target
  beta's Party Machine (`turn_start` income) on one world and win the contest; alpha
  passes in cycle 2. `ConfirmPaymentFn` declines every offer in cycle 1 (the
  `avoid_disable` offer, then the lift at beta's turn start) and pays in cycle 2. After
  cycle 1: the condition is on and the Party Machine paid no income. After cycle 2: the
  condition is gone, history has a `CoinDelta` with cause `condition_lifted`, and the
  Party Machine's income fired in the same turn.
- `digest/digest_test.go`: an `asset_condition_applied` with `Cause: "ability"` opens a
  `CrossAbilityStrike` with `Conditions == [disabled]`.
- `narrative/wire_renderer_test.go`: an ability strike with `[disabled]` ends with the
  disabled clause.

##### Commit message

```
feat: Marketers

- disabled condition: Active reads it; Targetable keeps the defender test (D-38)
- a disabled defender deals no counter and fires no own_defense_win entry
- disable_asset: pay half the listed cost or take the condition
- turn-start lift prompt, applied before the turn-start effects
- digest: an ability's condition joins the ability strike
```

### Phase 3 — Seditionists and Monopoly

#### Commit 4 — `feat: Seditionists`

##### Task 1 — `internal/faction/domain/asset.go`, `faction.go`

`asset.go` — add `ConditionNoAttack  ConditionKind = "no_attack"` to the condition
block, and `EffectBlockAttack  AbilityEffectType = "block_attack"` after
`EffectDisableAsset`.

`faction.go` — append below `AssetOwner`:
```go
// AttackBlocked reports whether a no_attack condition on asset holds now. It is
// checked when read: the source must still exist, share the asset's world, and
// belong to another faction. The source need not be active.
func AttackBlocked(factions map[string]*Faction, asset *Asset) bool {
	for _, condition := range asset.Conditions {
		if condition.Kind != ConditionNoAttack {
			continue
		}
		owner := AssetOwner(factions, condition.SourceAssetID)
		if owner == nil || owner.ID == asset.OwnerID {
			continue
		}
		source := owner.Assets[condition.SourceAssetID]
		if source.Location.WorldID != "" && source.Location.WorldID == asset.Location.WorldID {
			return true
		}
	}
	return false
}
```

##### Task 2 — `internal/faction/engine/action/actions/eligibility.go`, `attack.go`, `expand_influence.go`

`eligibility.go` — add the `state` import; replace `eligibleAttackers`:
```go
// eligibleAttackers returns faction assets that may attack this turn: the active
// ones whose attack no Seditionists block.
func eligibleAttackers(faction *domain.Faction, factionState *state.FactionState) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Active() && !domain.AttackBlocked(factionState.Factions, asset) {
			result = append(result, asset)
		}
	}
	return result
}
```
`attack.go` — in `Validate` and `Inputs`, `eligibleAttackers(faction)` →
`eligibleAttackers(faction, factionState)`.

`expand_influence.go` — `rivalAssetsOnWorld(rival *domain.Faction, world string,
factionState *state.FactionState)`, its condition
`asset.Location.WorldID == world && !asset.Stealthy && asset.Active() && !domain.AttackBlocked(factionState.Factions, asset)`,
and the caller in `baseAttack.resolve` passes `factionState`.

##### Task 3 — `internal/faction/engine/action/actions/ability/block_attack.go` (new) — full contents:

```go
package ability

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// blockAttack is Seditionists: pay the rolled cost to attach to the target, which
// cannot attack while the attachment holds (domain.AttackBlocked). Attaching moves
// the attachment: this source's no_attack leaves every other asset. A faction that
// cannot pay the roll attaches nothing and pays nothing.
func blockAttack(ctx StepContext) ([]domain.Mutation, error) {
	cost := ctx.RollResult
	if ctx.Faction.Coin < cost {
		return nil, nil
	}
	condition := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: ctx.Asset.ID}
	mutations := []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             -cost,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}
	for _, holderID := range slices.Sorted(maps.Keys(ctx.FactionState.Factions)) {
		for _, held := range domain.SortedAssets(ctx.FactionState.Factions[holderID]) {
			if held.ID != ctx.Target.Asset.ID && slices.ContainsFunc(held.Conditions, condition.Is) {
				mutations = append(mutations, domain.AssetConditionCleared{
					FactionID:         holderID,
					AssetID:           held.ID,
					Condition:         condition,
					Cause:             "ability",
					CausedByFactionID: ctx.Faction.ID,
				})
			}
		}
	}
	return append(mutations, domain.AssetConditionApplied{
		FactionID:         ctx.Target.Faction.ID,
		AssetID:           ctx.Target.Asset.ID,
		Condition:         condition,
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}), nil
}
```

##### Task 4 — `dispatch.go`, `stage.go`, `rulebook.go`

`dispatch.go` — `handlers` gains `domain.EffectBlockAttack: blockAttack,`.

`stage.go` — `assetTargetRules` gains:
```go
	domain.EffectBlockAttack: func(candidate, acting *domain.Asset) bool {
		attached := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: acting.ID}
		return !slices.ContainsFunc(candidate.Conditions, attached.Is)
	},
```
`rulebook.go` — `toAbilityEffect` gains `case "block_attack": return domain.EffectBlockAttack, nil`;
`convertAbilityStep` gains:
```go
	if step.Effect == domain.EffectBlockAttack && (step.Target != domain.TargetAsset || step.Roll == nil) {
		return domain.AbilityStep{}, fmt.Errorf("block_attack requires target = %q and a die", "asset")
	}
```

##### Task 5 — `rulebooks/swn/assets/cunning_assets.toml`

After the Seditionists' `description` line add:
```toml

[assets.Seditionists.ability]
effect = "block_attack"
target = "asset"
die = "1d4"
```

##### Task 6 — `internal/faction/narrative/wire_templates.go`

`conditionClauses` gains
`domain.ConditionNoAttack: "{attacker}'s agitators keep one of {defender}'s assets from attacking.",`.

##### Task 7 — tests

- `domain/faction_test.go`: `TestAttackBlocked` — table: source on the same world,
  other faction → true; source gone → false; source on another world → false; target
  in transit (empty world) → false; source captured into the target's faction → false;
  source disabled and unready → true.
- `ability/block_attack_test.go` (new), via `Dispatch` with a Seditionists fixture
  (`Roll{Die: 1d4}`), roller 3:
  - `Coin: 5` → `CoinDelta{f1, -3}`, then `AssetConditionApplied{f2, b2, {no_attack, a1}}`.
  - `Coin: 2` → no mutation.
  - this source already on `c1` (faction `f3`) → `CoinDelta`, `AssetConditionCleared{f3, c1}`,
    `AssetConditionApplied{f2, b2}`, in that order.
  - `TestAssetCandidates` gains: an asset that holds this source's `no_attack` is not a
    candidate; one that holds another source's is.
- `actions/attack_test.go`: an attacker with a live `no_attack` → `Validate` false; with
  its source in another world → true. A blocked defender still counters (D-14).
- `actions/expand_influence_test.go`: a rival asset with a live `no_attack` is not
  offered as a base attacker.
- `rulebook/rulebook_test.go`: Seditionists load with `EffectBlockAttack`, `TargetAsset`,
  die 1d4; `block_attack` with no die fails to load.
- `engine/testharness/scenarios/seditionists_test.go` (new):
  `TestSeditionistsBlockAndMove` — alpha's Seditionists attach to beta's Saboteurs;
  beta's Attack does not offer the Saboteurs as an attacker. Alpha attaches to beta's
  Informers next cycle; the Saboteurs are cleared and are offered again (Decision
  Record — Execution 4).

##### Commit message

```
feat: Seditionists

- no_attack condition, read by domain.AttackBlocked when attackers are chosen
- the block holds while the source exists, shares the world, and has another owner
- block_attack pays the rolled 1d4, moves the attachment, or does nothing
- Attack and Expand Influence's base attackers both read it; counters do not
```

#### Commit 5 — `feat: Monopoly`

##### Task 1 — `internal/faction/domain/asset.go`

After `EffectBlockAttack` add `EffectLevyCoin AbilityEffectType = "levy_coin"`.

##### Task 2 — `internal/faction/engine/action/actions/ability/stage.go`

In `resolveTarget`'s `TargetFaction` case, **find**
`		candidates := otherFactions(factionState, faction.ID)` and **replace with**
`		candidates := factionCandidates(step.Effect, faction.ID, asset, factionState)`.
In `HasTargets`, **find**
`		return len(otherFactions(factionState, faction.ID)) > 0` and **replace with**
`		return len(factionCandidates(step.Effect, faction.ID, asset, factionState)) > 0`.
Append:
```go
// factionTargetRules narrows a step's faction candidates by its effect. An effect
// with no rule takes every other faction (Informers).
var factionTargetRules = map[domain.AbilityEffectType]func(candidate *domain.Faction, acting *domain.Asset) bool{
	domain.EffectLevyCoin: func(candidate *domain.Faction, acting *domain.Asset) bool {
		world := acting.Location.WorldID
		return world != "" && slices.ContainsFunc(domain.SortedAssets(candidate), func(asset *domain.Asset) bool {
			return asset.Location.WorldID == world && !asset.Stealthy
		})
	},
}

// factionCandidates returns every other faction, name-sorted, that passes the
// effect's rule.
func factionCandidates(effect domain.AbilityEffectType, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) []*domain.Faction {
	rule := factionTargetRules[effect]
	var candidates []*domain.Faction
	for _, candidate := range otherFactions(factionState, actingFactionID) {
		if rule == nil || rule(candidate, acting) {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}
```

##### Task 3 — `internal/faction/engine/action/actions/ability/levy_coin.go` (new) — full contents:

```go
package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const levyAmount = 1

// levyCoin is Monopoly: the target faction pays the owner 1 Coin. A faction that
// cannot pay loses one asset of its choice on the world instead, from all its
// assets there, stealthed ones too (Effort 3 Decision 11).
func levyCoin(ctx StepContext) ([]domain.Mutation, error) {
	target := ctx.Target.Faction
	if target.Coin >= levyAmount {
		return []domain.Mutation{
			domain.CoinDelta{FactionID: target.ID, Delta: -levyAmount, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
			domain.CoinDelta{FactionID: ctx.Faction.ID, Delta: levyAmount, Cause: "ability", CausedByFactionID: ctx.Faction.ID},
		}, nil
	}
	var onWorld []*domain.Asset
	for _, asset := range domain.SortedAssets(target) {
		if asset.Location.WorldID == ctx.Asset.Location.WorldID {
			onWorld = append(onWorld, asset)
		}
	}
	lost, err := ctx.Collector.SelectAssetToLose(target, onWorld, ctx.Rulebook)
	if err != nil {
		return nil, fmt.Errorf("levy_coin: %w", err)
	}
	if lost == nil {
		// deliberate: a plain error, not ErrNoSelection, because the book gives the target no way out of the loss (Effort 3 Decision Record 5).
		return nil, fmt.Errorf("levy_coin: %s must choose an asset to lose", target.ID)
	}
	return []domain.Mutation{domain.AssetRemoved{
		FactionID:         target.ID,
		AssetID:           lost.ID,
		Cause:             string(domain.EffectLevyCoin),
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
```
`dispatch.go` — `handlers` gains `domain.EffectLevyCoin: levyCoin,`.

##### Task 4 — `internal/faction/rulebook/rulebook.go`

`toAbilityEffect` gains `case "levy_coin": return domain.EffectLevyCoin, nil`;
`convertAbilityStep` gains:
```go
	if step.Effect == domain.EffectLevyCoin && (step.Target != domain.TargetFaction || step.Contest != nil || step.Roll != nil) {
		return domain.AbilityStep{}, fmt.Errorf("levy_coin requires target = %q and no contest or die", "faction")
	}
```

##### Task 5 — `internal/faction/engine/action/collector.go` and prompt surfaces

`collector.go` — after `SelectAsset` add:
```go
	SelectAssetToLose(faction *domain.Faction, assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
```
Run `go generate ./internal/faction/engine/action/...`.

`tui/adapter/channels.go` — after `AskConfirmPayment` add `AskSelectAssetToLose`, and:
```go
type SelectAssetToLosePayload struct {
	Faction *domain.Faction
	Assets  []*domain.Asset
}
```
`tui/adapter/action_collector.go` — after `SelectAsset`:
```go
func (a *actionCollector) SelectAssetToLose(faction *domain.Faction, assets []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAssetToLose, faction, SelectAssetToLosePayload{Faction: faction, Assets: assets})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Asset)
	return chosen, nil
}
```
`execution.go` `askEntries`:
```go
	adapter.AskSelectAssetToLose: {"Lose an Asset", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectAssetToLosePayload)
		return overlay.NewSelectAssetToLose(fmt.Sprintf("%s cannot pay: lose which asset?", p.Faction.Name), p.Assets, m.rulebook)
	}},
```
`tui/views/turn/overlay/asset.go` — after `NewSelectAsset`:
```go
// SelectAssetToLose is Monopoly's loss pick. Esc does nothing: the book gives the
// target no way out of the loss (Effort 3 Decision Record 5).
type SelectAssetToLose struct {
	SelectOverlay[*domain.Asset]
}

func NewSelectAssetToLose(title string, assets []*domain.Asset, rb *rulebook.Rulebook) SelectAssetToLose {
	return SelectAssetToLose{NewSelectAsset(title, assets, rb)}
}

func (o SelectAssetToLose) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil
	}
	updated, cmd := o.SelectOverlay.Update(msg)
	o.SelectOverlay = updated.(SelectOverlay[*domain.Asset])
	return o, cmd
}

func (o SelectAssetToLose) Help() help.KeyMap { return formHelp{} }
```
`tui/adapter/smoke.go`:
```go
func (SmokeActionCollector) SelectAssetToLose(_ *domain.Faction, _ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
```
`engine/testharness/collector.go` — field
`SelectAssetToLoseFn func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error)` and:
```go
func (c *ScriptedCollector) SelectAssetToLose(faction *domain.Faction, assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
	if c.SelectAssetToLoseFn != nil {
		return c.SelectAssetToLoseFn(faction, assets, rulebook)
	}
	return nil, nil
}
```

##### Task 6 — `rulebooks/swn/assets/wealth_assets.toml`

Replace the Monopoly `description` line:
```toml
description = "An open or tacit stranglehold on vital local businesses or resources. As an action, force one other faction with unstealthed assets on this world to pay the owner 1 Coin. If the target faction can't pay, it loses one asset of its choice on the world."
```
Before `[assets.Monopoly.attack]` add:
```toml
[assets.Monopoly.ability]
effect = "levy_coin"
target = "faction"

```

##### Task 7 — digest

`digest/build.go` — in the `asset_removed` cause switch, before `case "refit":` add:
```go
				case "levy_coin":
					lossBeat := getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					lossBeat.Losses = append(lossBeat.Losses, AssetMove{
						AssetID:   m.AssetID,
						AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
						Cause:     m.Cause,
					})
					getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAbilityStrike, factionState)
```
`wire_renderer.go` — in `renderLoss`, **find:**
```go
	if move.Cause == "captured" {
		result += " (captured)"
	}
	return result
}

func renderMovement(
```
**Replace with:**
```go
	switch move.Cause {
	case "captured":
		result += " (captured)"
	case "levy_coin":
		result += " (forfeit to a monopoly)"
	}
	return result
}

func renderMovement(
```

##### Task 8 — `docs/architecture/engine/ability-catalog.md`

After the "As-built (A.001.3 Effort 5)" note add:
```
> **As-built (A.001.6 Effort 3):** the three AP-6 assets ship as A-side abilities
> whose lasting effect is asset state. Marketers is `disable_asset` (a contest, then
> the target pays half the listed cost or takes `disabled`); Seditionists is
> `block_attack` (1d4 Coin; `no_attack` checked when read); Monopoly is `levy_coin`
> (an action, by the book's text: 1 Coin, or an asset of the target's choice). Asset
> targets are built (`TargetAsset`); the pay question is `hooks.PaymentCollector`.
```
In the A-flag effects table, mark Marketers, Monopoly and Seditionists ✅ and replace
the Monopoly row's notes with `an action by the book's text (levy_coin); the reactive TOML reading is retired`.
In the S-flag table, delete the `coin-drain-on-attack` row (its only asset is now an
ability). In "Open reconciliations", delete the "Monopoly's flag-vs-text mismatch"
bullet.

##### Task 9 — tests

- `ability/levy_coin_test.go` (new), via `Dispatch` with a Monopoly fixture on
  Anchorage:
  - `f2` `Coin: 3` → `[CoinDelta{f2, -1}, CoinDelta{f1, +1}]`, no prompt.
  - `f2` `Coin: 0`, assets `b1` (Anchorage, stealthed), `b2` (Anchorage), `b3` (Hadrian) →
    `SelectAssetToLose(f2, [b1, b2])`; it returns `b1` → `AssetRemoved{f2, b1, "levy_coin", f1}`.
  - the prompt returns nil → an error that is not `ErrNoSelection` (Decision Record —
    Execution 5).
  - `TestFactionCandidates`: a faction with only a stealthed asset on the world, or
    only assets elsewhere, is not a candidate; Informers' effect still takes every
    other faction.
- `use_asset_ability_test.go`: `TestUseAssetAbility_TwoMonopoliesSeeTheFirstLevy` —
  `f2` has `Coin: 1`; both Monopolies target `f2`. The first levy takes the Coin; the
  second calls `SelectAssetToLose` (Decision 3).
- `rulebook/rulebook_test.go`: Monopoly loads with `EffectLevyCoin`, `TargetFaction`,
  its attack kept; `levy_coin` with a contest fails to load.
- `digest/digest_test.go`: a `levy_coin` removal adds a loss to `f2`'s beat and an
  ability strike `f1`→`f2`.
- `narrative/wire_renderer_test.go`: a `levy_coin` loss ends with "(forfeit to a monopoly)".
- `execution_test.go`: `TestAskEntriesCoverEveryKind` passes with the new kind.
- `overlay/asset_test.go` (new): `TestSelectAssetToLoseIgnoresEsc` — Esc gives no
  command.

##### Commit message

```
feat: Monopoly

- levy_coin: a faction with unstealthed assets here pays the owner 1 Coin
- a faction that cannot pay loses an asset of its choice on the world
- SelectAssetToLose names the faction asked; adapter, smoke, scripted
- Monopoly is an action by the book's text; catalog reconciled
```
