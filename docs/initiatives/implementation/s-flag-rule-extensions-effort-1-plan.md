# S-flag Rule Extensions (A.001.6) — Effort 1 Plan: Combat foundations

## In Plain Words

- **What:** teaches combat three new things. An asset can carry a lasting mark (a
  "condition") that stops it doing something. An asset can change sides and keep its
  identity, its orders and anything riding on it. And an attack can "hit" without doing
  damage, so the hit triggers a special rule instead. It also lets an asset's rule fire
  at each attack, hit or miss. Three assets use these: Saboteurs (a target cannot use its
  ability until the Saboteurs' owner's next turn), Treachery (a hit steals the target)
  and Cracked Comms (a failed attack on it can turn the attacker on itself).
- **Why:** every later combat asset in this initiative builds on these pieces, so they
  come first.
- **How it fits:** the first of five efforts in A.001.6 (see the
  [overview plan](s-flag-rule-extensions-plan.md)). Efforts 2 and 3 use its conditions
  and its change of owner; Effort 5 uses its no-damage attack.

## Context / Goal

Build the effort the overview plan scopes as "Combat foundations", on branch
`feature/s-flag-rule-extensions`.

- Discovery: [s-flag-rule-extensions-discovery.md](../discovery/s-flag-rule-extensions-discovery.md)
  — State Held on an Asset, Change of Asset Owner, Special Attacks and the Matchup
  Route, Damage Redirected into Combat.
- Overview plan: [s-flag-rule-extensions-plan.md](s-flag-rule-extensions-plan.md) —
  Decisions 3, 4, 5, 6, 9, 13, 17, 21 and 23 bind here. This plan reshapes 14 and 17
  and the D-40 part of 5 (Decision 8 below); the overview is reconciled to match.
- Rules source: `docs/rules/swn-assets.md`.

## Decisions Ratified in Planning

Re-grounding against the code found eight gaps (F1–F8). Each decision names the gap
or open question it closes.

1. **`Condition{Kind, SourceAssetID, ExpiresAtTurnOf}`.** The condition step clears
   every condition whose `ExpiresAtTurnOf` is the faction whose turn starts. Saboteurs
   sets it to the attacking faction at apply time. Other kinds leave it empty and end
   through their own reader. F2: the end cannot be found through the source asset,
   which a counter can kill or a capture can move before the end comes. Rejected: a
   source faction with the end derived from the kind (a reader takes it for the current
   owner); a stored end struct for every kind (stores ends only one reader acts on).
   Resolves overview OQ "condition shape".
2. **`AssetConditionApplied` / `AssetConditionCleared`, keyed on Kind + SourceAssetID.**
   Each carries the whole `Condition`. Apply replaces an entry with the same key, so the
   list never holds duplicates. Two Saboteurs on one target hold two entries. Rejected:
   clear by kind only (two sources would clear each other).
3. **`NextAssetID` scans every faction for the prefix.** F5: after a capture the old
   owner's next purchase would re-issue the captured asset's ID, and `hits()` matches
   the striker by ID alone (`hits.go:36`). Rejected: matching hits on (owner, ID) — a
   condition's `SourceAssetID` stays ambiguous; a campaign-wide counter (new state).
4. **`capture_target` fizzles when its input batch removes the target.** F4: Treachery
   hits X, a later matchup kills X, and the owner change would apply after the removal
   and miss. The reactor uses `batchHP` as `self_damage` does. A removal emitted by a
   sibling reactor in the same round (Zealots' `self_damage`) is not visible and still
   misses; a fence-sign marks it.
5. **Planetary Seizure phase 1 re-checks rival presence in `CheckLock`, and Decision 17
   stays in `UpdateProgress`.** F1: goal progress reads the action's batch before the
   reactors (`orchestrator.go:478,481`), so Treachery's capture is invisible to it; a
   rival that leaves on its own turn was already invisible. The lock re-check advances
   the goal at the seizer's next turn. One `rivalPresentOn` helper serves both sites.
   Rejected: running goal progress after the reactors (changes the input of all 12 goal
   handlers).
6. **A `Matchup` route on `Registrar`; the effect decides and Attack runs the
   self-attack.** `FireMatchup(trigger, MatchupContext) MatchupOutcome`. `block_ability`
   returns mutations; `counterstrike` returns `Counterstrike: true`, and Attack asks the
   defender (`ConfirmCounterstrike`), rolls and applies through `applyAssetDamage`. No
   collector, registry or HP tracker crosses into `effect`. `own_attack` fires once the
   defender is chosen, before the rolls, so the condition precedes the matchup's damage
   (F3). Self-damage carries `Cause: "counterstrike"`, so `hits()` ignores it and no
   `own_attack_hit` fires on it (F6). Rejected: callbacks in `FireContext` (control flow
   behind a closure); the route emits and Attack folds (moves the lethal check out of
   `applyAssetDamage`, which Effort 2's Cat 6 needs). Resolves overview OQs "matchup
   route context" and "Cracked Comms' self-damage".
7. **Effect names `block_ability`, `capture_target`, `counterstrike`.** Verb-object, as
   `restrict_attack`; `block_ability` pairs with Effort 3's `block_attack`. Resolves the
   overview OQ "effect names".
8. **A capture changes exactly one asset, and every order carries over.** The captured
   asset keeps its own order. A captured transport keeps its order and its cargo; the
   cargo stays with its old owner as riders, who can attack the transport from its hex
   (the "revolt"). A captured rider stays aboard its old owner's transport (the planted
   traitor). No order is cancelled or edited. "Only the target switches sides" is the
   book's text; "riders stay aboard" is a house reading. Robert's call. Reverses overview
   Decision 14, drops D-40's cancel and manifest removal, and closes the overview OQ
   "D-40's manifest removal" with no mutation.
9. **`primeHooks` runs again after an action batch that holds an `AssetOwnerChanged`.**
   Hooks register once per cycle (`orchestrator.go:44-46`), and `allReactors` finds a
   reactor through its owner's assets in state. Without this, a captured transport has
   no reactor for the rest of the cycle, and a completed order strands its cargo in
   space. Budgets live on factions, so nothing resets. Reverses the discovery's
   "captured effects stay quiet until the next cycle".
10. **`buildCargoMoved` addresses each cargo `AssetMoved` to the cargo's current
    owner,** found by ID (`domain.AssetOwner`; IDs are unique after Decision 3). A
    mixed-owner manifest then never misses. Rejected: an owner per manifest entry
    (persisted format change plus a mutation).
11. **Expand Influence offers only damaging attackers as base attackers.** A special
    attack does nothing to a Base, and a zero-damage `BaseHPDelta` would fire Treachery
    with no asset to take (F8). The same filter drops assets with no attack at all,
    which `SelectBaseAttackers` offers today and `resolve` then skips. Resolves the
    overview OQ "special-attack asset in Expand Influence".
12. **R.014 runs first as a wip refactor.** `askEntries` replaces the two parallel
    switches; a kind with no row panics, and a test walks every kind. The placeholder
    overlay goes with the old default arm. Resolves the overview OQ on R.014.
13. **Cracked Comms fires only when the defense succeeds outright.** Robert's call. On a
    standard tie both sides hit (`attack.go:161,195`; `swn-faction-mechanics.md:150`,
    "both Attack and Counterattack succeed"), so a tie could read as a successful
    defense. It is ruled not to be: the attack landed, so the attacker was not turned.
    Fire on `defenderWins && !attackerWins`. A tie that a resolver gives to the defender
    (`TieDefenderWins`) is an outright defender win and does fire. Rejected: firing on
    `defenderWins` (every standard tie).
14. **The condition step is `turn.ExpireConditions`, joined to the turn-start batch.**
    It runs inside the `BookkeepingApplied` gate before `FireTurnStart`, and its
    mutations reach `OnBookkeepingApplied` with the rest of `startMutations`.
15. **Treachery's capture lands on a tie even when the counter kills Treachery.** Robert's
    call. Under the tie rule both sides succeed, so the target switches sides and the
    owner gains the Coin. The reactor skips its own `AssetRemoved` when the batch
    already removes Treachery, so history records the loss once (the apply arm's
    `delete` would drop the second one silently). Rejected: the capture fizzles when
    Treachery dies (overrides "both succeed").
16. **One plan, over the total budget.** About 1,710 lines against a total budget of
    1,500; prose (about 585) and every commit (largest 346) are inside their budgets.
    Not split, because the overrun is literal code and the seven commits share one
    decision set. Robert's call.

## Decision Record — Execution

1. **Commit 0, Task 4: range over `AskKindCount`.** The test loops
   `for kind := range adapter.AskKindCount`, not a counted loop. Same walk; gopls's
   `rangeint` check flagged the counted form (Go 1.26). Anchor drift: the old
   `newOverlay` doc comment started at line 457, not 455; the text matched.
2. **Commit 2, Task 3: scenario path.** The scenarios live in
   `engine/testharness/scenarios`, not `testharness/scenarios`. The test goes in a new
   `conditions_test.go` there; the unit test is `TestExpireConditions`.
3. **Commit 3, Task 4: tests that pass a nil `FactionState`.** `Resolve` and
   `purchaseEnabler` now read `factionState.Factions`, so seven test call sites that
   passed `nil` (or left `StepContext.FactionState` unset) panicked. They now pass an
   empty `&state.FactionState{}`: `buy_asset_test.go` (five `Resolve` calls),
   `refit_asset_test.go` (one) and `purchase_enabler_test.go` (four contexts; the
   insufficient-Coin case returns before the call and stays as is). No production
   caller passes `nil`. The planned `TestNextAssetID_ScansEveryFaction` went in
   `domain/asset_test.go` as written; the plan's "two calls" there are to
   `NewPurchasedAsset`, not `NextAssetID`.
4. **Commit 4, Task 3: `rulebook_test.go:50` is a read, not a literal.**
   `d := asset.Attack.Damage` still compiles once `Damage` is a pointer (field access
   dereferences it). It gains a nil guard, so a regression fails the test and does not
   panic. The other two literals changed as planned.
5. **Commit 4, Task 8: test placement.** The scenario lives in
   `engine/testharness/scenarios/capture_test.go` (same path drift as Decision Record
   2). It pins the turn order to `[alpha, beta]` after `Turn.Start`, because the order
   is rolled and beta moving first would tick the transport off Treachery's hex. It
   also checks the hook registry between the two cycles: cycle 2 re-registers every
   hook, which would hide Decision 9, so only that check shows the mid-cycle re-prime.
   The "capture_target with no block fails to load" case went into
   `TestLoadEffects_LoudFailure` in `effects_test.go`, where its sibling cases live;
   the Treachery load check is a `TestLoad` subtest in `rulebook_test.go`.
6. **Commit 4, Task 2: a no-damage attack needs an `own_attack_hit` entry.** Robert's
   call, added in execution. Once `Damage` is a pointer, an attack line with the
   `damage` key left out (or empty) loads as a special attack; before, the loader
   crashed on it. With no hit rule, such an asset hits for nothing and does nothing.
   `convertAsset` now fails the load when an attack has no damage and no
   `own_attack_hit` effect, with a fence-sign. A misspelt key needed no change: the
   loader already rejects undecoded keys (`rulebook.go`, fence-signed). If Effort 5's
   no-damage asset fires on another trigger, the check widens to that trigger.
   Rejected: filing it for later (a GM's omission loads silently today).
7. **Commit 5, Task 9: `registrars_test.go` checks three routes.**
   `TestRegistrars_DeclareARouteForEveryTrigger` held the same two-route rule as
   `validateEntry`, so it failed on `block_ability` ("hook trigger needs Register").
   It gets the same three-way `switch` as Task 3. The plan's "the real SWN rulebook
   still validates" needed no new test: every scenario builds the engine through
   `NewWithRulebook` from the real rulebook, which runs `RegisterEffectsList`.
8. **Commit 5, Task 9: scenario placement.** Same path drift as Decision Record 2.
   The tests go in a new `engine/testharness/scenarios/saboteurs_test.go`. The lethal
   variant is its own test, `TestSaboteursConditionPrecedesALethalHit`; a miss in
   its batch would fail `RunCycle`. Both use `runOrderedCycle`. The first test also
   checks that beta's turn after the clear offers Use Asset Ability again. In Task 8,
   the catalogue's `asset_owner_changed` line says the cargo keeps its old owner (the
   apply arm moves only the one asset).
9. **Commit 6, Task 4: no counterstrike prompt for an attacker that cannot hurt
   itself.** Robert's call, added in execution. When the attacker has no attack damage
   and no counter (Treachery is the only one today), its self-attack can do nothing,
   so Cracked Comms does not offer the prompt and no dice are rolled. A guard
   `canHurtItself` joins the prompt condition, with a fence-sign. Rejected: prompt
   anyway (the GM gets a prompt with no effect). `counterstrike` keeps its nil-damage
   return for an attacker that has only one of the two.
10. **Commit 6, Task 8: goals need no change, and two more tests.** The `"counterstrike"`
    cause does not hide credit from a goal. Goal progress runs only for the acting
    faction (the attacker). Each goal reader that checks `Cause == "attack"` also
    checks `CausedByFactionID == acting faction`, and the self-damage is caused by the
    defender. Added: a `TestAttack_Counterstrike` case for Decision Record 9, and
    `TestRenderCrossEvent_Counterstruck` in `wire_renderer_test.go` for the new clause.
    The attack tests are one table, `TestAttack_Counterstrike`, which also checks how
    many rolls each case uses.

---

## Shared Context

- **Test patterns.** Actions: `fixedRoller{values}` and gomock `MockCollector`
  (`attack_test.go`). Effects: `fixedRoller{value}`, `activeAsset` (`effects` tests).
  Scenarios: `testharness.NewHarness` with the real SWN rulebook and `ScriptedCollector`.
- **Mock regeneration.** After a change to `action.Collector`, run
  `go generate ./internal/faction/engine/action/...`.
- **Every commit ends green:** `go build ./... && go test ./internal/faction/...`.
- **Digest wording.** Conditions ride the attack's cross event as a clause per kind;
  owner changes land as a loss and an acquisition with `Cause: "captured"`; a cross
  attack with no damage, no Base hit and no kill uses its own template. Cleared
  conditions are dropped (expiry is bookkeeping).

## Out of Scope

- `disabled`, `no_attack`, `sheltered` (Efforts 2 and 3, overview Decision 6).
- Cat 6 interception for `counterstrike`'s lethal damage — `applyAssetDamage` is the
  seam Effort 2 dispatches from.
- The digest's cross events name an asset only when it is destroyed (`build.go:183,196`),
  so `{defender_asset}` renders empty on a non-lethal hit. Pre-existing; candidate
  backlog bug, not fixed here.
- A loyalty roll for captured cargo (a house rule Robert may file).

---

## Work Breakdown

### Phase 1 — Prompt table and conditions

#### Commit 0 — `refactor: AskKind dispatch table`

##### Task 1 — `internal/faction/tui/adapter/channels.go`

**Find:**
```go
	AskSelectPurchaseEnablerAsset
)
```
**Replace with:**
```go
	AskSelectPurchaseEnablerAsset
	// AskKindCount counts the kinds above; keep it last.
	AskKindCount
)
```

##### Task 2 — `internal/faction/tui/views/turn/execution/execution.go`

At the banner (line 250), **find** `phase = askPhaseLabel(m.activeAsk)` and **replace
with** `phase = askEntryFor(m.activeAsk).label`.

**Delete** the whole `askPhaseLabel` function (lines 255–310) and the whole `newOverlay`
method with its doc comment (lines 455–533). **Add** in their place:

```go
// askEntry is one AskKind's banner label and overlay factory.
type askEntry struct {
	label   string
	overlay func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay
}

var askEntries = map[adapter.AskKind]askEntry{
	adapter.AskAwaitCheckpoint: {"Cycle Checkpoint", func(m Model, _ adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewCheckpoint(m.cycleNumber)
	}},
	adapter.AskSelectAction: {"Action", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAction(msg.Payload.(adapter.SelectActionPayload).Available)
	}},
	adapter.AskSelectStatRaise: {"Stat Raise", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewStatRaise(msg.Faction, msg.Payload.(adapter.SelectStatRaisePayload).Eligible)
	}},
	adapter.AskSelectMovementDecisions: {"Movement", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
	}},
	adapter.AskSelectTransportCargo: {"Load Cargo", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectTransportCargoPayload)
		return overlay.NewTransportCargo(p.Transport, p.EligibleCargo, p.Profile, m.rulebook)
	}},
	adapter.AskSelectAsset: {"Select Asset", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAsset("Select asset", msg.Payload.(adapter.SelectAssetPayload).Assets, m.rulebook)
	}},
	adapter.AskSelectModifiers: {"Pre-roll Modifiers", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectModifiers(msg.Payload.(adapter.SelectModifiersPayload).Offers)
	}},
	adapter.AskConfirmReroll: {"Reroll?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmReroll(msg.Payload.(adapter.ConfirmRerollPayload).Directive)
	}},
	adapter.AskSelectBuyOrder: {"Buy Asset", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectBuyOrderPayload)
		return overlay.NewBuyOrder(p.PurchasableByWorld, p.WorldNames, p.Coin)
	}},
	adapter.AskSelectRefitOrder: {"Refit Asset", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewRefitOrder(msg.Payload.(adapter.SelectRefitOrderPayload).Options, m.rulebook)
	}},
	adapter.AskSelectRepairOrders: {"Repair Asset", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewRepairOrders(msg.Payload.(adapter.SelectRepairOrdersPayload).Targets, m.rulebook)
	}},
	adapter.AskSelectBribeTarget: {"Bribe", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectBribeTargetPayload)
		return overlay.NewBribe(p.Bases, p.OwnerNames, p.Coin)
	}},
	adapter.AskSelectSeizeTarget: {"Seize Planet", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
	}},
	adapter.AskSelectChangeHomeworldTarget: {"Change Homeworld", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
	}},
	adapter.AskSelectAttackers: {"Select Attackers", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAttackers(msg.Payload.(adapter.SelectAttackersPayload).Eligible, m.rulebook)
	}},
	adapter.AskSelectDefender: {"Select Defender", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectDefenderPayload)
		return overlay.NewSelectDefender(p.Attacker, p.Eligible, p.OwnerNames, m.rulebook)
	}},
	adapter.AskConfirmRedirectToBase: {"Redirect to Base?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.ConfirmRedirectToBasePayload)
		return overlay.NewConfirmRedirectToBase(p.DefenderFaction, p.Base, p.Damage)
	}},
	adapter.AskSelectExpandInfluenceOrder: {"Expand Influence", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectExpandInfluenceOrderPayload)
		return overlay.NewExpand(p.NewBaseWorlds, p.WorldNames, p.ReinforceBases, p.Coin)
	}},
	adapter.AskConfirmRivalFreeAttack: {"Rival Free Attack?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.ConfirmRivalFreeAttackPayload)
		return overlay.NewConfirmRivalFreeAttack(p.Rival, p.RivalRoll, p.FactionRoll)
	}},
	adapter.AskSelectBaseAttackers: {"Base Attackers", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectBaseAttackersPayload)
		return overlay.NewSelectBaseAttackers(p.Rival, p.Eligible, m.rulebook)
	}},
	adapter.AskSelectAbilityAssets: {"Ability Assets", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewSelectAbilityAssets(msg.Payload.(adapter.SelectAbilityAssetsPayload).Candidates, m.rulebook)
	}},
	adapter.AskConfirmAbilityApplied: {"Apply Ability?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.ConfirmAbilityAppliedPayload)
		return overlay.NewConfirmAbilityApplied(p.Asset, p.Def)
	}},
	adapter.AskSelectFactionTestTarget: {"Ability Target", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.SelectFactionTestTargetPayload)
		return overlay.NewSelectFactionTestTarget(p.Asset, p.Effect, p.Candidates)
	}},
	adapter.AskSelectRelocateOrders: {"Relocate", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewRelocateOrders(msg.Payload.(adapter.SelectRelocateOrdersPayload), m.rulebook)
	}},
	adapter.AskSelectPurchaseEnablerAsset: {"Purchase (Enabler)", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewPurchaseEnabler(msg.Payload.(adapter.SelectPurchaseEnablerAssetPayload))
	}},
}

// askEntryFor fails loudly on a kind with no askEntries row: a new AskKind must
// add one, and TestAskEntriesCoverEveryKind catches the miss before runtime.
func askEntryFor(kind adapter.AskKind) askEntry {
	entry, ok := askEntries[kind]
	if !ok {
		panic(fmt.Sprintf("execution: no askEntries row for AskKind %d", kind))
	}
	return entry
}

// newOverlay builds the overlay for an ask. A method so the factory can read the
// cycle number, rulebook and spatial map that the ask does not carry.
func (m Model) newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay {
	return askEntryFor(msg.Kind).overlay(m, msg)
}
```

##### Task 3 — delete `internal/faction/tui/views/turn/overlay/placeholder.go` and `placeholder_test.go`

`NewPlaceholder` was reached only from the old default arm.

##### Task 4 — `internal/faction/tui/views/turn/execution/execution_test.go`

Append:
```go
func TestAskEntriesCoverEveryKind(t *testing.T) {
	for kind := range adapter.AskKindCount {
		if _, ok := askEntries[kind]; !ok {
			t.Errorf("askEntries has no row for AskKind %d", kind)
		}
	}
}
```

##### Commit message

```
refactor: AskKind dispatch table (R.014)

- askEntries maps each AskKind to its banner label and overlay factory
- a missing row panics; a test walks every kind up to AskKindCount
- drop the placeholder overlay the old default arm used
```

#### Commit 1 — `feat: asset conditions`

##### Task 1 — `internal/faction/domain/asset.go`

**Find:**
```go
type Asset struct {
```
**Replace with:**
```go
type ConditionKind string

const ConditionNoAbility ConditionKind = "no_ability"

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

type Asset struct {
```

**Find:**
```go
	CurrentOrder *MovementOrder `toml:"current_order,omitempty"`
}
```
**Replace with:**
```go
	CurrentOrder *MovementOrder `toml:"current_order,omitempty"`
	Conditions   []Condition    `toml:"conditions,omitempty"`
}

func (asset *Asset) HasCondition(kind ConditionKind) bool {
	return slices.ContainsFunc(asset.Conditions, func(condition Condition) bool { return condition.Kind == kind })
}
```

In `Clone`, **find** `	clone := *asset` and **replace with**:
```go
	clone := *asset
	clone.Conditions = slices.Clone(asset.Conditions)
```

##### Task 2 — `internal/faction/domain/mutation.go`

Append:
```go
// AssetConditionApplied puts a condition on an asset, replacing one with the same
// kind and source.
type AssetConditionApplied struct {
	FactionID         string    `json:"faction_id"`
	AssetID           string    `json:"asset_id"`
	Condition         Condition `json:"condition"`
	Cause             string    `json:"cause"`
	CausedByFactionID string    `json:"caused_by_faction_id"`
}

func (m AssetConditionApplied) Type() string { return "asset_condition_applied" }

// AssetConditionCleared removes the condition with Condition's kind and source.
type AssetConditionCleared struct {
	FactionID         string    `json:"faction_id"`
	AssetID           string    `json:"asset_id"`
	Condition         Condition `json:"condition"`
	Cause             string    `json:"cause"`
	CausedByFactionID string    `json:"caused_by_faction_id"`
}

func (m AssetConditionCleared) Type() string { return "asset_condition_cleared" }
```

##### Task 3 — `internal/faction/engine/mutation/mutation.go`

Add `"slices"` to the imports. **Find:**
```go
		case domain.MovementOrderIssued:
```
**Replace with:**
```go
		case domain.AssetConditionApplied:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Conditions = append(slices.DeleteFunc(asset.Conditions, v.Condition.Is), v.Condition)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetConditionCleared:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Conditions = slices.DeleteFunc(asset.Conditions, v.Condition.Is)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderIssued:
```

##### Task 4 — `internal/faction/engine/action/actions/use_asset_ability.go`

`Validate` and `Inputs` each build the candidate list; both must read `no_ability`, or
the action is offered when every ability asset is sabotaged. Replace both bodies:

```go
func (u *UseAssetAbility) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	return len(abilityCandidates(faction, rulebook)) > 0
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := u.collector.SelectAbilityAssets(faction, abilityCandidates(faction, rulebook), rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selectedAssets = selected
	return nil
}
```
and add below `hasActionFlag`:
```go
// abilityCandidates returns the faction's assets that may use their ability now,
// in ID order.
func abilityCandidates(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Asset {
	var candidates []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) {
			candidates = append(candidates, asset)
		}
	}
	return candidates
}
```

##### Task 5 — `internal/faction/narrative/digest/types.go`, `build.go`, `../wire_renderer.go`, `../wire_templates.go`

`types.go` — in `CrossEvent`, after `StealthBroken bool`, add
`Conditions []domain.ConditionKind` and add the `domain` import.

`build.go` — **find** `			case "goal_initiated":` and insert before it:
```go
			case "asset_condition_applied":
				var m domain.AssetConditionApplied
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.FactionID != record.FactionID {
					getOrCreateBeat(beats, m.FactionID, factionState)
					seenFactionIDs[m.FactionID] = struct{}{}
					ce := getOrCreateCross(crossMap, record.FactionID, m.FactionID, CrossAttack, factionState)
					ce.Conditions = append(ce.Conditions, m.Condition.Kind)
				}

			case "asset_condition_cleared":
				// dropped — expiry is bookkeeping
```

`wire_templates.go` — append:
```go
var conditionClauses = map[domain.ConditionKind]string{
	domain.ConditionNoAbility: "{defender} cannot use asset abilities there until {attacker}'s next turn.",
}

var crossNoDamageTemplates = []string{
	"{attacker} moved against {defender}, but dealt no damage.",
	"{attacker} struck at {defender} — no damage recorded.",
}
```
(add the `domain` import).

`wire_renderer.go` — replace the `case digest.CrossAttack:` arm of `renderCrossEvent`:
```go
	case digest.CrossAttack:
		templates := crossAttackTemplates
		if ce.DamageToDefender == 0 && ce.BaseHit == nil && !ce.DefenderAssetDestroyed {
			templates = crossNoDamageTemplates
		}
		result := strings.ReplaceAll(pick(rng, templates), "{attacker}", ce.Attacker.Name)
		result = strings.ReplaceAll(result, "{defender}", ce.Defender.Name)
		result = strings.ReplaceAll(result, "{attacker_asset}", ce.AttackerAsset.AssetName)
		result = strings.ReplaceAll(result, "{defender_asset}", ce.DefenderAsset.AssetName)
		result = strings.ReplaceAll(result, "{damage}", fmt.Sprintf("%d", ce.DamageToDefender))
		for _, kind := range ce.Conditions {
			if clause, ok := conditionClauses[kind]; ok {
				clause = strings.ReplaceAll(clause, "{attacker}", ce.Attacker.Name)
				result += " " + strings.ReplaceAll(clause, "{defender}", ce.Defender.Name)
			}
		}
		return result
```

##### Task 6 — `internal/faction/tui/views/turn/execution/format.go`

**Find** `	case domain.AssetRemoved:` and insert before it:
```go
	case domain.AssetConditionApplied:
		return fmt.Sprintf("asset %s +%s", mut.AssetID, mut.Condition.Kind), true
	case domain.AssetConditionCleared:
		return fmt.Sprintf("asset %s -%s", mut.AssetID, mut.Condition.Kind), true
```

##### Task 7 — tests

- `mutation/mutation_test.go`: `TestMutationEngine_AssetCondition` — apply
  `{no_ability, "sab-1", "f2"}` to `a1`; apply it again (still one entry); apply
  `{no_ability, "sab-2", "f2"}` (two entries); clear `sab-1` (one left, `sab-2`); an
  applied condition on asset `"ghost"` returns one miss with `EntityID: "ghost"`.
- `domain/asset_test.go`: `TestAsset_CloneCopiesConditions` — mutate the clone's
  `Conditions[0].Kind`; the original is unchanged. `TestAsset_HasCondition`.
- `actions/use_asset_ability_test.go`: in `TestUseAssetAbility_Validate` add a case
  where `a1` carries `no_ability` → `Validate` is false.
- `digest/digest_test.go`: `TestConditionAppliedJoinsCrossAttack` — a record for `f1`
  with an `asset_condition_applied` on `f2`'s asset → one cross event `f1`→`f2` with
  `Conditions == [no_ability]`; a cleared condition adds nothing.
- `narrative/wire_renderer_test.go`: a cross attack with `DamageToDefender: 0` and
  `Conditions: [no_ability]` renders a no-damage template plus the clause.
- `execution/format_test.go`: add the two rows to `TestFormatMutation_Surfaced`
  (`"asset x +no_ability"`, `"asset x -no_ability"`).

##### Commit message

```
feat: asset conditions

- Asset.Conditions with Condition{Kind, SourceAssetID, ExpiresAtTurnOf}, cloned
- AssetConditionApplied/Cleared keyed on kind + source; apply replaces same key
- no_ability read by Use Asset Ability's candidates (Validate and Inputs)
- digest clause on the cross attack; TUI lines; no-damage cross template
```

#### Commit 2 — `feat: condition step at turn start`

##### Task 1 — `internal/faction/engine/turn/conditions.go` (new) — full contents:

```go
package turn

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ExpireConditions clears every condition, on any faction's asset, that ends at
// the start of factionID's turn. Factions in ID order, then assets in ID order.
func ExpireConditions(factionState *state.FactionState, factionID string) []domain.Mutation {
	var mutations []domain.Mutation
	for _, holderID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		for _, asset := range domain.SortedAssets(factionState.Factions[holderID]) {
			for _, condition := range asset.Conditions {
				if condition.ExpiresAtTurnOf != factionID {
					continue
				}
				mutations = append(mutations, domain.AssetConditionCleared{
					FactionID:         holderID,
					AssetID:           asset.ID,
					Condition:         condition,
					Cause:             "condition_expired",
					CausedByFactionID: factionID,
				})
			}
		}
	}
	return mutations
}
```

##### Task 2 — `internal/faction/engine/orchestrator.go`

In `runBookkeepingPhase`, **find:**
```go
		startMutations = e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)
```
**Replace with:**
```go
		startMutations = append(turn.ExpireConditions(factionState, faction.ID), e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)...)
```

##### Task 3 — tests

- `turn/conditions_test.go` (new): three factions; `f2`'s asset carries
  `{no_ability, src, ExpiresAtTurnOf: "f1"}` and `{no_ability, src2, "f3"}`; `f3`'s
  asset carries one for `"f1"`. `ExpireConditions(state, "f1")` returns two clears,
  `f2` first; a condition with empty `ExpiresAtTurnOf` is never cleared.
- `engine/testharness/scenarios/conditions_test.go` (new):
  `TestConditionExpiresAtSourceTurnStart` — alpha's asset holds
  a `no_ability` expiring at beta's turn; run a cycle; after it the condition is gone,
  and history has `asset_condition_cleared` with cause `condition_expired`.

##### Commit message

```
feat: clear expiring conditions at turn start

- turn.ExpireConditions clears conditions whose ExpiresAtTurnOf is the new turn's faction
- joined to the turn-start batch inside the BookkeepingApplied gate
```

### Phase 2 — Change of owner and special attacks

#### Commit 3 — `feat: change of asset owner`

##### Task 1 — `internal/faction/domain/mutation.go`

Append:
```go
// AssetOwnerChanged moves an asset from FactionID to ToFactionID. The asset keeps
// its ID, its order and its conditions.
type AssetOwnerChanged struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	ToFactionID       string `json:"to_faction_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetOwnerChanged) Type() string { return "asset_owner_changed" }
```

##### Task 2 — `internal/faction/engine/mutation/mutation.go`

In `Apply`, **find:**
```go
	for _, mutation := range mutations {
		log.Debug("mutation", "type", mutation.Type())
```
**Replace with:**
```go
	// deliberate: owner changes apply after the rest of the batch, not in emitted order, because reactors read pre-apply state and address the asset's old owner (overview Decision 5).
	ordered := make([]domain.Mutation, 0, len(mutations))
	var ownerChanges []domain.Mutation
	for _, mutation := range mutations {
		if _, ok := mutation.(domain.AssetOwnerChanged); ok {
			ownerChanges = append(ownerChanges, mutation)
			continue
		}
		ordered = append(ordered, mutation)
	}
	for _, mutation := range append(ordered, ownerChanges...) {
		log.Debug("mutation", "type", mutation.Type())
```
and add the arm before `case domain.MovementOrderIssued:`:
```go
		case domain.AssetOwnerChanged:
			from, fromOK := factionState.Factions[v.FactionID]
			to, toOK := factionState.Factions[v.ToFactionID]
			switch {
			case !fromOK:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			case !toOK:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.ToFactionID})
			default:
				if asset, ok := from.Assets[v.AssetID]; ok {
					delete(from.Assets, v.AssetID)
					asset.OwnerID = v.ToFactionID
					if to.Assets == nil {
						to.Assets = make(map[string]*domain.Asset)
					}
					to.Assets[v.AssetID] = asset
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			}
```

##### Task 3 — `internal/faction/domain/asset.go` and `faction.go`

Replace `NextAssetID` and `NewPurchasedAsset`:
```go
func NextAssetID(factions map[string]*Faction, faction *Faction, definitionID string) string {
	prefix := fmt.Sprintf("%s-%s-", faction.ID, definitionID)
	highest := highestSuffix(faction, prefix)
	// deliberate: scan every faction, not only the owner, because a captured asset keeps its ID in its new owner's map (D-9).
	for _, holder := range factions {
		highest = max(highest, highestSuffix(holder, prefix))
	}
	return fmt.Sprintf("%s%d", prefix, highest+1)
}

func highestSuffix(faction *Faction, prefix string) int {
	highest := 0
	for id := range faction.Assets {
		if !strings.HasPrefix(id, prefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(id, prefix))
		if err == nil && n > highest {
			highest = n
		}
	}
	return highest
}

// NewPurchasedAsset builds a freshly-bought asset for a faction: a fresh ID, full HP, not
// yet Ready (activates next turn), and Maintained. Shared by the Buy Asset action and the
// purchase_enabler ability so the two purchase paths can't drift on the asset shape.
func NewPurchasedAsset(factions map[string]*Faction, faction *Faction, def *AssetDefinition, worldID string) Asset {
	return Asset{
		ID:           NextAssetID(factions, faction, def.ID),
		DefinitionID: def.ID,
		OwnerID:      faction.ID,
		Location:     Location{WorldID: worldID},
		CurrentHP:    def.HP,
		Ready:        false,
		Maintained:   true,
	}
}
```
`faction.go` — append below `SortedAssets`:
```go
// AssetOwner returns the faction that holds assetID, or nil. Asset IDs are unique
// across factions (NextAssetID), so at most one matches.
func AssetOwner(factions map[string]*Faction, assetID string) *Faction {
	for _, faction := range factions {
		if _, ok := faction.Assets[assetID]; ok {
			return faction
		}
	}
	return nil
}
```

##### Task 4 — callers of `NextAssetID` / `NewPurchasedAsset`

- `actions/buy_asset.go` `Resolve`: rename the `_ *state.FactionState` parameter to
  `factionState`; `domain.NewPurchasedAsset(factionState.Factions, faction, def, ba.buyOrder.World)`.
- `actions/refit_asset.go` `Resolve`: same rename;
  `ID: domain.NextAssetID(factionState.Factions, faction, newDef.ID),`.
- `actions/ability/purchase_enabler.go`:
  `domain.NewPurchasedAsset(ctx.FactionState.Factions, ctx.Faction, def, worldID)`.
  `ctx.Faction` is the working clone; `NextAssetID` scans it as well as the map.
- `tui/views/manage/wizard/wizard.go` `instantiateAssets`:
  `domain.NextAssetID(map[string]*domain.Faction{tempFaction.ID: tempFaction}, tempFaction, def.ID)`.
  A new faction's ID cannot prefix any existing asset.
- `domain/asset_test.go`: pass `nil` as the map in the two existing
  `NewPurchasedAsset` calls.
- Tests that call `Resolve` or `purchaseEnabler` with no state pass
  `&state.FactionState{}` (Decision Record 3).

##### Task 5 — `internal/faction/engine/effect/effects/transport.go`

Replace `buildCargoMoved`:
```go
// buildCargoMoved addresses the move to the cargo's current owner, which a
// capture can make differ from the transport's (Decision 8).
func buildCargoMoved(factionState *state.FactionState, transportOwnerID, cargoID string, toLoc domain.Location, cause string) domain.AssetMoved {
	ownerID := transportOwnerID
	var from domain.Location
	if owner := domain.AssetOwner(factionState.Factions, cargoID); owner != nil {
		ownerID = owner.ID
		from = owner.Assets[cargoID].Location
	}
	return domain.AssetMoved{
		FactionID:         ownerID,
		AssetID:           cargoID,
		FromLocation:      from,
		ToLocation:        toLoc,
		Cause:             cause,
		CausedByFactionID: transportOwnerID,
	}
}
```

##### Task 6 — `internal/faction/engine/goal/goals/planetary_seizure.go`

Replace the phase-1 branch of `CheckLock`:
```go
	if goal.ProcessPhase == 1 {
		// deliberate: re-check presence here, not only in UpdateProgress, because a rival can leave, or be taken by a reactor such as capture_target, where UpdateProgress cannot see it, and the Attack-only lock would then hold forever.
		if !rivalPresentOn(goal.TargetWorld.WorldID, faction.ID, factionState, nil) {
			return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{phaseOneCleared(faction.ID, goal.GoalID)}
		}
		return locks.GoalLock{Type: locks.LockRestrictActions, AllowedActions: []string{"Attack"}}, nil
	}
```
Replace `UpdateProgress` from `removedIDs := …` to the end:
```go
	cleared := make(map[string]bool)
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetRemoved:
			cleared[m.AssetID] = true
		case domain.AssetOwnerChanged:
			if m.ToFactionID == actingFaction.ID {
				cleared[m.AssetID] = true
			}
		}
	}
	if rivalPresentOn(goal.TargetWorld.WorldID, actingFaction.ID, factionState, cleared) {
		return nil
	}
	return []domain.Mutation{phaseOneCleared(actingFaction.ID, goal.GoalID)}
}

// rivalPresentOn reports whether a faction other than seizerID holds an
// unstealthed asset on world that cleared does not name.
func rivalPresentOn(world, seizerID string, factionState *state.FactionState, cleared map[string]bool) bool {
	for factionID, faction := range factionState.Factions {
		if factionID == seizerID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location.WorldID == world && !asset.Stealthy && !cleared[asset.ID] {
				return true
			}
		}
	}
	return false
}

func phaseOneCleared(factionID, goalID string) domain.GoalPhaseAdvanced {
	return domain.GoalPhaseAdvanced{
		FactionID:      factionID,
		GoalID:         goalID,
		ProcessPhase:   2,
		TurnsRemaining: 3,
		Cause:          "planetary_seizure_phase_advance",
	}
}
```

##### Task 7 — `internal/faction/engine/orchestrator.go`

In `RunCycle`, **find:**
```go
	// deliberate: reset before registering, because the TUI runs every cycle on one engine and each ApplyAll would add another copy of every hook.
	e.Hooks.Reset()
	e.Tag.ApplyAll(factionState, e.Hooks, e.log.With("engine", "tag"))
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.Rand, e.log.With("engine", "effect"))
```
**Replace with** `	e.primeHooks(factionState)` and add:
```go
// primeHooks registers every tag and effect hook from current state.
func (e *Engine) primeHooks(factionState *state.FactionState) {
	// deliberate: reset before registering, because the TUI runs every cycle on one engine and each ApplyAll would add another copy of every hook.
	e.Hooks.Reset()
	e.Tag.ApplyAll(factionState, e.Hooks, e.log.With("engine", "tag"))
	e.Effect.ApplyAll(factionState, e.Rulebook, e.Hooks, e.Rand, e.log.With("engine", "effect"))
}

func changesOwner(mutations []domain.Mutation) bool {
	return slices.ContainsFunc(mutations, func(mutation domain.Mutation) bool {
		_, ok := mutation.(domain.AssetOwnerChanged)
		return ok
	})
}
```
In `runActionPhase`, **find:**
```go
	observer.OnActionResolved(faction, selectedAction, combined)
```
**Replace with:**
```go
	if changesOwner(combined) {
		// deliberate: re-register now, not at the next cycle, because a captured asset's hooks are keyed under its old owner and nothing would find them (Decision 9).
		e.primeHooks(factionState)
	}
	observer.OnActionResolved(faction, selectedAction, combined)
```
(`slices` is added to the imports if absent.)

##### Task 8 — digest and TUI

`digest/build.go` — before `case "goal_initiated":` add:
```go
			case "asset_owner_changed":
				var m domain.AssetOwnerChanged
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				name := resolveAssetName(m.ToFactionID, m.AssetID, factionState, rulebook)
				loser := getOrCreateBeat(beats, m.FactionID, factionState)
				winner := getOrCreateBeat(beats, m.ToFactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				seenFactionIDs[m.ToFactionID] = struct{}{}
				loser.Losses = append(loser.Losses, AssetMove{AssetID: m.AssetID, AssetName: name, Cause: "captured"})
				winner.Acquisitions = append(winner.Acquisitions, AssetMove{AssetID: m.AssetID, AssetName: name, Cause: "captured"})
```
(`resolveAssetName` reads the end-of-cycle state, where the new owner holds it.)

`wire_renderer.go` — in `renderAcquisition`, **find:**
```go
	if move.From != "" {
		result += fmt.Sprintf(" (upgraded from %s)", move.From)
	}
	return result
```
**Replace with:**
```go
	if move.From != "" {
		result += fmt.Sprintf(" (upgraded from %s)", move.From)
	}
	if move.Cause == "captured" {
		result += " (captured)"
	}
	return result
```
and replace the body of `renderLoss`:
```go
	tmpl := pick(rng, lossTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", move.AssetName)
	if move.Cause == "captured" {
		result += " (captured)"
	}
	return result
```

`format.go` — before `case domain.AssetRemoved:` add:
```go
	case domain.AssetOwnerChanged:
		return fmt.Sprintf("asset %s captured by %s", mut.AssetID, mut.ToFactionID), true
```

##### Task 9 — tests

- `mutation/mutation_test.go`:
  `TestMutationEngine_AssetOwnerChanged` — `a1` moves `f1`→`f2`, keeps its ID, order
  and conditions, `OwnerID == "f2"`; unknown `ToFactionID` is a faction miss.
  `TestMutationEngine_OwnerChangeAppliesLast` (Decision 5) — batch
  `[AssetOwnerChanged{a1 f1→f2}, AssetHPDelta{f1 a1 -2}]` applies with no miss and `a1`
  ends in `f2` at HP −2.
- `domain/asset_test.go`: `TestNextAssetID_ScansEveryFaction` — `f2` holds
  `f1-SWN-F1-001-3`; `f1` holds `-1`; the next ID is `-4`.
- `effects/transport_test.go`: `TestTransportReactor_RiderOfAnotherFaction` — beta's
  transport carries `alpha-x` in its manifest; a `Progressed` emits `AssetMoved` with
  `FactionID: "alpha"`, `CausedByFactionID: "beta"`.
- `goals/lock_test.go`: `TestPlanetarySeizure_Phase1` gains a rival `f2` asset on
  Krylos (else it now advances). Add `TestPlanetarySeizure_Phase1_NoRivalAdvances`
  (empty world → `LockNone` + `GoalPhaseAdvanced{2,3}`) and a stealthed-rival case
  (also advances).
- `goals/progress_test.go`: `AssetOwnerChanged` of the last rival asset to the seizer
  advances; to a third faction it does not.
- The end-to-end capture scenario needs an emitter, so it lands with Treachery
  (Commit 4 Task 8).

##### Commit message

```
feat: change of asset owner

- AssetOwnerChanged moves an asset between factions and keeps ID, order, conditions
- Apply runs owner changes after the rest of the batch (overview Decision 5)
- NextAssetID scans every faction; AssetOwner finds a holder by ID
- cargo moves address the rider's current owner
- Planetary Seizure: lock re-check in phase 1; captures to the seizer count as cleared
- hooks re-register after a batch that changes an owner
- digest loss/acquisition "captured"; TUI line
```

#### Commit 4 — `feat: special attacks and Treachery`

##### Task 1 — `internal/faction/domain/asset.go`, `effect.go`

`asset.go` — in `AttackProfile`, replace `Damage       DiceRoll` with:
```go
	// Damage is nil for a special attack, which hits for no damage and leaves
	// its effect to the asset's own_attack_hit entries.
	Damage *DiceRoll
```
`effect.go` — add `SpecialCaptureTarget SpecialEffectType = "capture_target"` to the
const block; add
```go
type CaptureTargetParams struct {
	Coin DiceRoll
}
```
and `CaptureTarget *CaptureTargetParams` as the last field of `EffectDefinition`.

##### Task 2 — `internal/faction/rulebook/rulebook.go`, `effects.go`

`rulebook.go` — in `convertAsset`, replace `Damage:       *damage,` with
`Damage:       damage,`. After `effects` is built, fail the load when
`attack != nil && attack.Damage == nil` and no entry has
`Trigger == domain.TriggerOwnAttackHit`:
`attack has no damage and no own_attack_hit effect` (Decision Record 6; a
`LoudFailure` case in `effects_test.go` covers it).

`effects.go`:
- `effectRecord`: add `CaptureTarget  *amountRecord         \`toml:"capture_target"\``.
- `specialEffects`: append `domain.SpecialCaptureTarget`.
- `presentParamBlocks`: add
  `if record.CaptureTarget != nil { present = append(present, domain.SpecialCaptureTarget) }`.
- `convertEffectParams`: add
```go
	case domain.SpecialCaptureTarget:
		coin, err := convertAmount(definition.Effect, record.CaptureTarget)
		if err != nil {
			return err
		}
		definition.CaptureTarget = &domain.CaptureTargetParams{Coin: coin}
```

##### Task 3 — `Damage` readers

- `actions/attack.go` — replace the hit branch's first two lines
  (`damage := attackerDef.Attack.Damage.Roll(attack.roller)` and
  `base := factionBaseOnWorld(…)`) and wrap the existing redirect logic:
```go
		if attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins) {
			if attackerDef.Attack.Damage == nil {
				// deliberate: a special hit skips the redirect prompt, because it has no damage to send to a Base (D-15).
				attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, 0, assetHPTracker, attacker, false)
			} else {
				damage := attackerDef.Attack.Damage.Roll(attack.roller)
				base := factionBaseOnWorld(defenderFaction, attacker.Location.WorldID)
				// … existing redirect block unchanged …
			}
		}
```
- `actions/expand_influence.go` — in `baseAttack.resolve`, replace
  `eligible := rivalAssetsOnWorld(rival, target.Location.WorldID)` with
  `eligible := damagingAttackers(rivalAssetsOnWorld(rival, target.Location.WorldID), rulebook)`
  and add:
```go
// damagingAttackers keeps the assets whose attack deals damage. A special attack
// does nothing to a Base, and its hit would fire own_attack_hit with no asset hit.
func damagingAttackers(assets []*domain.Asset, rulebook *rulebook.Rulebook) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range assets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && def.Attack != nil && def.Attack.Damage != nil {
			result = append(result, asset)
		}
	}
	return result
}
```
- `tui/views/turn/overlay/attack.go` `attackerLabel` — replace the return:
```go
	damage := "special"
	if def.Attack.Damage != nil {
		damage = "dmg " + diceLabel(*def.Attack.Damage)
	}
	return fmt.Sprintf("%s · %s · %s", name, def.Attack.AttackerStat, damage)
```
- Test literals: `Damage: domain.DiceRoll{…}` → `Damage: &domain.DiceRoll{…}` in
  `attack_test.go` and `expand_influence_test.go`; `rulebook_test.go` reads `Damage`
  and gains a nil guard (Decision Record 4).

##### Task 4 — `internal/faction/engine/effect/effects/hits.go`

Add `targetAssetID string` to `hit` (empty for a Base hit), set it in the
`AssetHPDelta` arm: `hit{strikerID: m.CausedByAssetID, targetFactionID: m.FactionID, targetAssetID: m.AssetID}`.

##### Task 5 — `internal/faction/engine/effect/effects/capture_target.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// captureTargetRegistrar is Treachery: on a hit, the declaring asset is spent, its
// owner gains Coin, and the asset it hit changes sides at its current HP.
var captureTargetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerCaptureTarget,
}

func registerCaptureTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialCaptureTarget)+":"+ctx.Def.ID,
		&captureTargetReactor{
			factionID: ctx.Faction.ID,
			assetID:   ctx.Asset.ID,
			coin:      ctx.Entry.CaptureTarget.Coin,
			roller:    ctx.Roller,
		},
	)
}

type captureTargetReactor struct {
	factionID string
	assetID   string
	coin      domain.DiceRoll
	roller    domain.Roller
}

func (reactor *captureTargetReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	self := liveAsset(factionState, reactor.factionID, reactor.assetID)
	if !active(self) {
		return nil
	}
	for _, found := range hits(mutations, false) {
		if found.strikerID != reactor.assetID || found.targetAssetID == "" {
			continue
		}
		target := liveAsset(factionState, found.targetFactionID, found.targetAssetID)
		if target == nil {
			return nil
		}
		// deliberate: fizzle when the batch removes the target, not capture, because the owner change applies after the removal and would miss. A removal from a sibling reactor this round (Zealots) is not visible here and still misses.
		if _, removed := batchHP(mutations, target); removed {
			return nil
		}
		cause := string(domain.SpecialCaptureTarget)
		var extra []domain.Mutation
		// deliberate: skip the self-removal when the batch already removes this asset, not emit it twice, because on a tie the defender's counter can kill Treachery in the matchup it captures in (Decision 15).
		if _, removed := batchHP(mutations, self); !removed {
			extra = append(extra, domain.AssetRemoved{FactionID: reactor.factionID, AssetID: reactor.assetID, Cause: cause, CausedByFactionID: reactor.factionID})
		}
		return append(extra,
			domain.CoinDelta{FactionID: reactor.factionID, Delta: reactor.coin.Roll(reactor.roller), Cause: cause, CausedByFactionID: reactor.factionID},
			domain.AssetOwnerChanged{FactionID: found.targetFactionID, AssetID: target.ID, ToFactionID: reactor.factionID, Cause: cause, CausedByFactionID: reactor.factionID},
		)
	}
	return nil
}
```
Register it in `registrars.go`: `domain.SpecialCaptureTarget: captureTargetRegistrar,`.

##### Task 6 — `rulebooks/swn/assets/cunning_assets.toml`

After Treachery's `description` line add:
```toml

[assets.Treachery.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"

[[assets.Treachery.effects]]
trigger = "own_attack_hit"
effect  = "capture_target"
[assets.Treachery.effects.filters]
owner = "own"
scope = "self"
[assets.Treachery.effects.capture_target]
amount = "5"
```

##### Task 7 — digest

`build.go` — in the `asset_removed` switch, change `case "sell":` to
`case "sell", "capture_target":` (a spent Treachery reads as a loss).

##### Task 8 — tests

- `rulebook/rulebook_test.go`: Treachery loads with `Attack.Damage == nil` and one
  `capture_target` entry with `Coin.Modifier == 5`. `rulebook/effects_test.go`: a
  `capture_target` with no block fails to load (Decision Record 5).
- `actions/attack_test.go`: `TestAttack_SpecialHit` — a `Damage: nil` attacker wins;
  output is one `AssetHPDelta{Delta: 0, CausedByAssetID: "a1"}`; with a defender Base,
  `ConfirmRedirectToBase` is not called (no `EXPECT`).
- `actions/expand_influence_test.go`: a rival whose only asset has a special attack is
  offered no base attackers.
- `effects/capture_target_test.go` (new): a hit by `treachery` on beta's `x` returns
  `AssetRemoved{treachery}`, `CoinDelta{+5}`, `AssetOwnerChanged{beta→alpha x}`; a
  counter hit, another striker's hit, a Base hit and an inactive Treachery return nil;
  a batch that also removes `x` returns nil; a batch that also removes `treachery`
  (a tie's counter) returns only the `CoinDelta` and the `AssetOwnerChanged`.
- `engine/testharness/scenarios/capture_test.go`: `TestTreacheryCapturesAndKeepsRiders`
  — alpha's Treachery (`SWN-C7-003`) hits beta's transport (Heavy Drop Assets) in
  flight with a beta rider, turn order pinned to `[alpha, beta]`; after the cycle the
  transport is alpha's with its order intact, the rider is still beta's and still in
  the manifest, and the transport's reactor is registered under alpha; a further tick
  (next cycle) moves both with no miss. Covers Decisions 8, 9 and 10 end to end
  (Decision Record 5).

##### Commit message

```
feat: special attacks and Treachery

- AttackProfile.Damage is a pointer; nil is a special attack
- a no-damage attack with no own_attack_hit effect fails to load
- a special hit emits AssetHPDelta{0} and skips the redirect prompt
- Expand Influence offers only damaging base attackers
- capture_target: spend the asset, +Coin, AssetOwnerChanged on the target
- Treachery gains its book attack line and capture_target entry
```

### Phase 3 — The matchup route

#### Commit 5 — `feat: matchup route and Saboteurs`

##### Task 1 — `internal/faction/domain/effect.go`

Add `TriggerOwnAttack EffectTrigger = "own_attack"` and
`TriggerOwnDefenseWin EffectTrigger = "own_defense_win"` to the trigger block, and
`SpecialBlockAbility SpecialEffectType = "block_ability"` to the effect block.

##### Task 2 — `internal/faction/rulebook/effects.go`

Append `domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin` to `effectTriggers` and
`domain.SpecialBlockAbility` to `specialEffects`.

##### Task 3 — `internal/faction/engine/effect/registrar.go`

Replace the `Registrar` doc comment and struct:
```go
// Registrar is one S-side effect's spec: what the effect accepts in the data, and
// the code that runs it. Triggers keys the accepted sets on the trigger, because
// owner and scope read differently per trigger: an independent set per field would
// accept a pairing the effect never supports. An entry runs through Matchup when its
// trigger fires inside the Attack action (IsMatchupTrigger), through Fire when it is
// a phase event (IsPhaseTrigger), and through Register otherwise; a nil function
// means the effect has no such route. Validate is optional: it refuses a
// combination of fields that Accepts, which checks each field alone, cannot express.
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
	Matchup  func(MatchupContext) MatchupOutcome
	Validate func(domain.EffectDefinition) error
}
```
After `FireContext` add:
```go
// MatchupContext is one matchup of the Attack action, seen from the declaring
// asset: its attacker for own_attack, its defender for own_defense_win.
type MatchupContext struct {
	Faction       *domain.Faction
	Asset         *domain.Asset
	Def           *domain.AssetDefinition
	Entry         domain.EffectDefinition
	Opponent      *domain.Faction
	OpponentAsset *domain.Asset
	FactionState  *state.FactionState
}

// MatchupOutcome carries the route's mutations and the decisions the Attack
// action carries out itself, because they need its rolls, prompts and HP tracker.
type MatchupOutcome struct {
	Mutations     []domain.Mutation
	Counterstrike bool
}

// IsMatchupTrigger reports whether trigger fires from the Attack action at each
// matchup: a direct definition read the action raises (overview Decision 21).
func IsMatchupTrigger(trigger domain.EffectTrigger) bool {
	return trigger == domain.TriggerOwnAttack || trigger == domain.TriggerOwnDefenseWin
}
```
In `effectsListHandler.Apply`, replace `if IsPhaseTrigger(entry.Trigger) {` with
`if IsPhaseTrigger(entry.Trigger) || IsMatchupTrigger(entry.Trigger) {`.

In `validateEntry`, replace the two route checks with:
```go
	switch {
	case IsMatchupTrigger(entry.Trigger):
		if registrar.Matchup == nil {
			return fmt.Errorf("effect %q has no matchup route for trigger %q", entry.Effect, entry.Trigger)
		}
	case IsPhaseTrigger(entry.Trigger):
		if registrar.Fire == nil {
			return fmt.Errorf("effect %q has no phase route for trigger %q", entry.Effect, entry.Trigger)
		}
	default:
		if registrar.Register == nil {
			return fmt.Errorf("effect %q has no hook route for trigger %q", entry.Effect, entry.Trigger)
		}
	}
```

##### Task 4 — `internal/faction/engine/effect/effect.go`

Append:
```go
// FireMatchup runs the declaring asset's entries for trigger through each
// registrar's matchup route. A direct definition read, raised by the Attack
// action, which owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext) MatchupOutcome {
	var outcome MatchupOutcome
	for _, entry := range ctx.Def.Effects {
		if entry.Trigger != trigger {
			continue
		}
		ctx.Entry = entry
		// deliberate: no nil check on Matchup, because RegisterEffectsList rejects a matchup-route entry without one.
		result := e.registrars[entry.Effect].Matchup(ctx)
		outcome.Mutations = append(outcome.Mutations, result.Mutations...)
		outcome.Counterstrike = outcome.Counterstrike || result.Counterstrike
	}
	return outcome
}
```

##### Task 5 — `internal/faction/engine/effect/effects/block_ability.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// blockAbilityRegistrar is Saboteurs: the asset it attacks, hit or miss, cannot
// use its ability until the start of the attacking faction's next turn.
var blockAbilityRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttack: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Matchup: matchupBlockAbility,
}

func matchupBlockAbility(ctx effect.MatchupContext) effect.MatchupOutcome {
	return effect.MatchupOutcome{Mutations: []domain.Mutation{domain.AssetConditionApplied{
		FactionID: ctx.Opponent.ID,
		AssetID:   ctx.OpponentAsset.ID,
		Condition: domain.Condition{
			Kind:            domain.ConditionNoAbility,
			SourceAssetID:   ctx.Asset.ID,
			ExpiresAtTurnOf: ctx.Faction.ID,
		},
		Cause:             string(domain.SpecialBlockAbility),
		CausedByFactionID: ctx.Faction.ID,
	}}}
}
```
Register it in `registrars.go`: `domain.SpecialBlockAbility: blockAbilityRegistrar,`.

##### Task 6 — `internal/faction/engine/action/actions/attack.go`, `register.go`, `engine/core.go`

`attack.go` — add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"`
to the imports; add `effects   *effect.EffectsEngine` to `AttackAction` after `index`;
replace `NewAttack`:
```go
func NewAttack(collector action.Collector, roller domain.Roller, registry *hooks.Registry, index *world.Index, effects *effect.EffectsEngine) *AttackAction {
	return &AttackAction{collector: collector, roller: roller, registry: registry, index: index, effects: effects}
}
```
In `Resolve`, **find:**
```go
		attackResult := dispatch.RollWithHooks(
```
**Replace with:**
```go
		attack.mutations = append(attack.mutations, attack.fireMatchup(domain.TriggerOwnAttack, faction, attacker, attackerDef, defenderFaction, defender, factionState).Mutations...)

		attackResult := dispatch.RollWithHooks(
```
and add:
```go
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
```
All `NewAttack(...)` calls in `attack_test.go` gain a trailing `nil`.

`register.go` — add parameter `effectsEngine *effect.EffectsEngine` after
`rulebook *rulebook.Rulebook` (import `effect`), and pass it:
`return NewAttack(c, resolveRoller(), hookRegistry, worldIndex(worldEngine), effectsEngine)`.
Add to the doc comment: "The Attack action raises the matchup route itself, so it
takes the effects engine."

`core.go` — `actions.RegisterDefaultActions(e.Action, func() domain.Roller { return e.Rand }, e.Hooks, e.World, e.Rulebook, e.Effect)`.

##### Task 7 — `rulebooks/swn/assets/cunning_assets.toml`

After `[assets.Saboteurs.attack]`'s `damage = "2d4"` add:
```toml

[[assets.Saboteurs.effects]]
trigger = "own_attack"
effect  = "block_ability"
[assets.Saboteurs.effects.filters]
owner = "own"
scope = "self"
```

##### Task 8 — `docs/architecture/engine/effect-mutation.md`

In Shape, after "…before goal progress and the reactors.", add: "A third route,
`own_attack` and `own_defense_win`, is raised by the Attack action at each matchup
(`FireMatchup`), because Cracked Comms needs a prompt and a full roll inside the
matchup." In Key Decisions, in "The S side has two routes…", replace "the orchestrator
still owns each call site, and the effects engine resolves a phase event that the
orchestrator raises" with "the orchestrator, or the action that owns the event, raises
each route, and the effects engine resolves it (A.001.6 overview Decision 21)". Add
`asset_condition_applied` / `asset_condition_cleared` and `asset_owner_changed` to the
catalogue's Assets group.

##### Task 9 — tests

- `effect/effect_test.go`: a matchup entry on a registrar with no `Matchup` fails
  `RegisterEffectsList`; `ApplyAll` registers no hook for an `own_attack` entry;
  `FireMatchup` runs only entries of the given trigger and ORs `Counterstrike`.
- `effects/block_ability_test.go` (new): the condition lands on the opponent asset
  with `SourceAssetID` = declaring asset and `ExpiresAtTurnOf` = declaring faction.
- `effects/registrars_test.go`: `TestRegistrars_DeclareARouteForEveryTrigger` checks
  the three routes with the same `switch` as Task 3 (Decision Record 7). The real SWN
  rulebook validates through every scenario's `NewWithRulebook`.
- `engine/testharness/scenarios/saboteurs_test.go` (new):
  `TestSaboteursBlockAbilityUntilNextTurn` — alpha's Saboteurs attack beta's Informers
  and miss; beta's turn offers no Use Asset Ability for them; at alpha's next turn
  start the condition clears. `TestSaboteursConditionPrecedesALethalHit` — the
  Saboteurs' damage kills the defender and the batch applies with no miss (F3 order).

##### Commit message

```
feat: matchup route and Saboteurs

- Registrar.Matchup with MatchupContext/MatchupOutcome; own_attack, own_defense_win
- EffectsEngine.FireMatchup; the Attack action raises own_attack before the rolls
- RegisterDefaultActions hands the effects engine to Attack (overview Decision 21)
- block_ability puts no_ability on the attacked asset until the attacker's next turn
- effect-mutation.md: third route, reworded call-site Key Decision, new mutations
```

#### Commit 6 — `feat: Cracked Comms self-attack`

##### Task 1 — domain and loader

`domain/effect.go`: `SpecialCounterstrike SpecialEffectType = "counterstrike"`.
`rulebook/effects.go`: append `domain.SpecialCounterstrike` to `specialEffects`.

##### Task 2 — `internal/faction/engine/effect/effects/counterstrike.go` (new) — full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// counterstrikeRegistrar is Cracked Comms: after a successful defense, its owner
// may turn the attacking asset on itself. The Attack action asks and rolls.
var counterstrikeRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnDefenseWin: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Matchup: func(effect.MatchupContext) effect.MatchupOutcome {
		return effect.MatchupOutcome{Counterstrike: true}
	},
}
```
Register it in `registrars.go`: `domain.SpecialCounterstrike: counterstrikeRegistrar,`.

##### Task 3 — `internal/faction/engine/action/collector.go`

After `ConfirmRedirectToBase` add:
```go
	ConfirmCounterstrike(defenderFaction *domain.Faction, source *domain.Asset, attacker *domain.Asset) (bool, error)
```
Run `go generate ./internal/faction/engine/action/...`.

##### Task 4 — `internal/faction/engine/action/actions/attack.go`

Change `applyAssetDamage` to take the cause: signature
`(factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker map[string]int, causedByAsset *domain.Asset, counter bool, cause string)`;
use `Cause: cause` in both mutations. Every existing call passes `"attack"`.

Replace the counter branch (from `// Counterattack damage:` to its closing brace) with:
```go
		attackerWins := attackRoll > defenseRoll || (attackRoll == defenseRoll && tieOutcome != hooks.TieDefenderWins)
		defenderWins := defenseRoll > attackRoll || (defenseRoll == attackRoll && tieOutcome != hooks.TieAttackerWins)

		// Counterattack damage: defender wins on tie (TieStandard/TieDefenderWins) or strictly greater.
		if defenderWins {
			if defenderDef.Counter != nil && dispatch.ResolveAttackEligibility(attack.registry, hooks.AttackEligibilityQuery{
				Attacker:    defender,
				AttackerDef: defenderDef,
				Defender:    attacker,
				DefenderDef: attackerDef,
				Side:        hooks.SideCounter,
			}) {
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, counterDamage, assetHPTracker, defender, true, "attack")
			}
		}

		// deliberate: own_defense_win needs an outright win, not a tie, because on a tie the attack also landed and the defense did not succeed (Decision 13).
		if defenderWins && !attackerWins {
			outcome := attack.fireMatchup(domain.TriggerOwnDefenseWin, defenderFaction, defender, defenderDef, faction, attacker, factionState)
			attack.mutations = append(attack.mutations, outcome.Mutations...)
			// deliberate: no prompt when the attacker has no damage and no counter, because its self-attack could do nothing (Decision Record 9).
			canHurtItself := attackerDef.Attack.Damage != nil || attackerDef.Counter != nil
			if outcome.Counterstrike && canHurtItself && attacker.CurrentHP+assetHPTracker[attacker.ID] > 0 {
				confirmed, err := attack.collector.ConfirmCounterstrike(defenderFaction, defender, attacker)
				if err != nil {
					return fmt.Errorf("attack: counterstrike prompt: %w", err)
				}
				if confirmed {
					attack.counterstrike(faction, attacker, attackerDef, defenderFaction, defender, assetHPTracker, factionState, rulebook)
				}
			}
		}
```
and use `if attackerWins {` for the hit branch's condition (the variables move above
it). Add:
```go
// counterstrike turns the attacker on itself (D-18): its attack roll against its
// own defense. A hit deals its attack damage; a miss deals its counter damage, if
// it has one. The cause is not "attack", so no own_attack_hit fires on it.
func (attack *AttackAction) counterstrike(faction *domain.Faction, attacker *domain.Asset, attackerDef *domain.AssetDefinition, defenderFaction *domain.Faction, source *domain.Asset, tracker map[string]int, factionState *state.FactionState, rulebook *rulebook.Rulebook) {
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
		return
	}
	attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, damage.Roll(attack.roller), tracker, source, false, string(domain.SpecialCounterstrike))
}
```

##### Task 5 — prompt surfaces

- `tui/adapter/channels.go`: add `AskConfirmCounterstrike` before `AskKindCount`; add
```go
type ConfirmCounterstrikePayload struct {
	DefenderFaction *domain.Faction
	Source          *domain.Asset
	Attacker        *domain.Asset
}
```
- `tui/adapter/action_collector.go`, after `ConfirmRedirectToBase`:
```go
func (a *actionCollector) ConfirmCounterstrike(defenderFaction *domain.Faction, source *domain.Asset, attacker *domain.Asset) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmCounterstrike, defenderFaction, ConfirmCounterstrikePayload{
		DefenderFaction: defenderFaction,
		Source:          source,
		Attacker:        attacker,
	})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
```
- `tui/views/turn/overlay/attack.go`:
```go
// NewConfirmCounterstrike asks the defender whether Cracked Comms turns the
// attacking asset on itself. Answer: bool.
func NewConfirmCounterstrike(defenderFaction *domain.Faction, source, attacker *domain.Asset, rb *rulebook.Rulebook) ConfirmOverlay {
	title := fmt.Sprintf("Turn %s's attack back on itself?", assetName(attacker, rb))
	header := fmt.Sprintf("%s defended with %s", defenderFaction.Name, assetName(source, rb))
	return NewConfirm(title, header, "Turn it", "Let it pass")
}
```
- `execution.go` `askEntries`:
```go
	adapter.AskConfirmCounterstrike: {"Counterstrike?", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.ConfirmCounterstrikePayload)
		return overlay.NewConfirmCounterstrike(p.DefenderFaction, p.Source, p.Attacker, m.rulebook)
	}},
```
- `tui/adapter/smoke.go`:
```go
func (SmokeActionCollector) ConfirmCounterstrike(_ *domain.Faction, _ *domain.Asset, _ *domain.Asset) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
- `engine/testharness/collector.go`: field
  `ConfirmCounterstrikeFn func(*domain.Faction, *domain.Asset, *domain.Asset) (bool, error)`
  and
```go
func (c *ScriptedCollector) ConfirmCounterstrike(defenderFaction *domain.Faction, source *domain.Asset, attacker *domain.Asset) (bool, error) {
	if c.ConfirmCounterstrikeFn != nil {
		return c.ConfirmCounterstrikeFn(defenderFaction, source, attacker)
	}
	return false, nil
}
```

##### Task 6 — `rulebooks/swn/assets/cunning_assets.toml`

After Cracked Comms' `description` line add:
```toml

[[assets.CrackedComms.effects]]
trigger = "own_defense_win"
effect  = "counterstrike"
[assets.CrackedComms.effects.filters]
owner = "own"
scope = "self"
```

##### Task 7 — digest and TUI

`build.go` — in `asset_hp_delta`, add:
```go
				case "counterstrike":
					ce := getOrCreateCross(crossMap, record.FactionID, m.CausedByFactionID, CrossAttack, factionState)
					ce.DamageToAttacker -= m.Delta
					ce.Counterstruck = true
```
and in the `asset_removed` switch, before `case "bookkeeping":`, add:
```go
				case "counterstrike":
					ce := getOrCreateCross(crossMap, record.FactionID, m.CausedByFactionID, CrossAttack, factionState)
					ce.AttackerAsset = AssetMove{
						AssetID:   m.AssetID,
						AssetName: resolveAssetName(m.FactionID, m.AssetID, factionState, rulebook),
						Cause:     m.Cause,
					}
					ce.AttackerAssetDestroyed = true
```
`types.go` — `CrossEvent` gains `Counterstruck bool`.
`wire_templates.go` — `const counterstruckClause = "{defender}'s countermeasures turned the attack back on {attacker}."`;
`wire_renderer.go` — in the `CrossAttack` arm, before `return result`:
```go
		if ce.Counterstruck {
			clause := strings.ReplaceAll(counterstruckClause, "{attacker}", ce.Attacker.Name)
			result += " " + strings.ReplaceAll(clause, "{defender}", ce.Defender.Name)
		}
```
`format.go` — replace the `AssetHPDelta` arm:
```go
	case domain.AssetHPDelta:
		switch {
		case mut.Delta == 0:
			return fmt.Sprintf("asset %s hit (no damage)", mut.AssetID), true
		case mut.Cause == string(domain.SpecialCounterstrike):
			return fmt.Sprintf("asset %s HP %+d (counterstrike)", mut.AssetID, mut.Delta), true
		}
		return fmt.Sprintf("asset %s HP %+d", mut.AssetID, mut.Delta), true
```

##### Task 8 — tests

- `actions/attack_test.go` (fixture def `"comms"` with an `own_defense_win` →
  `counterstrike` entry; a real `effect.New(effects.Registrars(), …)` with
  `RegisterEffectsList` for it, passed to `NewAttack`):
  - defender wins outright, confirm true; self rolls hit → `AssetHPDelta` on `a1` with
    `Cause: "counterstrike"`, `CausedByAssetID: "d1"`; lethal adds `AssetRemoved`.
  - self rolls miss, attacker has no counter → no damage.
  - confirm false → no self rolls consumed.
  - a standard tie → `ConfirmCounterstrike` not called (Decision 13).
  - a tie with a `TieDefenderWins` resolver → `ConfirmCounterstrike` is called.
  - an attacker with no damage and no counter → `ConfirmCounterstrike` not called
    (Decision Record 9).
- `narrative/wire_renderer_test.go`: a `Counterstruck` cross event ends with the
  counterstruck clause.
- `effects/self_damage_test.go`: a `counterstrike` `AssetHPDelta` caused by the Zealots'
  own asset does not fire `self_damage` (F6).
- `execution/format_test.go`: zero-delta and counterstrike rows.
- `digest/digest_test.go`: a counterstrike record sets `Counterstruck` and
  `DamageToAttacker`.
- `execution/execution_test.go`: `TestAskEntriesCoverEveryKind` passes with the new kind.

##### Commit message

```
feat: Cracked Comms self-attack

- counterstrike on own_defense_win; fires on an outright defense win only
- ConfirmCounterstrike asks the defender; the attacker rolls against itself
- self-damage carries cause "counterstrike", so own_attack_hit does not fire
- adapter, overlay, smoke and scripted collector for the new prompt
- digest counterstruck clause; TUI lines for zero-damage and counterstrike hits
```
