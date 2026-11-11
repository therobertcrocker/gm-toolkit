# S-flag Rule Extensions (A.001.6) — Effort 4 Plan: Purchase and government

## In Plain Words

- **What:** the campaign's saved state now records which faction governs each world.
  Every purchase of a restricted asset asks that government for permission; the GM
  answers for a world no faction governs. A refusal ends the action with nothing
  bought. Popular Movement makes its world always say yes to its owner. Capital Fleet
  needs permission to be bought. Stealth stops being a hard-coded special case and
  becomes data: a quality bought for an asset the faction already holds.
- **Why:** today "permission" means "owns a Base on the world". Several factions can
  pass that test at once, and nothing ever grants or refuses anything, which disagrees
  with the book. Planetary Seizure never records which world it took, and its XP is
  always 1.
- **How it fits:** the fourth effort to run in A.001.6 (see the
  [overview plan](s-flag-rule-extensions-plan.md)). It depends on no other effort.
  Effort 5 extends the same permission request from purchases to movement.

## Context / Goal

Build the effort the overview plan scopes as "Purchase and government", on branch
`feature/s-flag-rule-extensions`.

- Discovery: [s-flag-rule-extensions-discovery.md](../../discovery/completed/s-flag-rule-extensions-discovery.md)
  — Purchase That Targets Another Asset, Permission Model (government and purchase
  rows); D-2 to D-6, D-19 to D-23, D-26, D-35, D-41, D-42.
- Overview plan: [s-flag-rule-extensions-plan.md](s-flag-rule-extensions-plan.md) —
  Decisions 8, 10, 13, 18, 20 and 22 bind here. This plan reshapes D-26 (Decision 1)
  and overview Decision 10's Inside Enemy Territory reading (Decision 6); the overview
  is reconciled to match.
- Rules source: `docs/rules/swn-assets.md` (Stealth, Capital Fleet, Popular
  Movement) and the book's permission rule (page 216, quoted in the discovery).

## Decisions Ratified in Planning

Re-grounding against the code found five gaps (F1–F5). Each decision names the gap
or the overview open question it closes.

1. **Capital Fleet's raise-only permission is a definition field,
   `purchase_permission = true`.** Robert's call. F1: hooks register on asset
   instances at cycle start (`effect.go:56`), and the fleet being raised has no
   instance, so D-26's `permission_query → require_permission` hook has nothing to
   register on. `AssetDefinition.NeedsPurchasePermission()` reads the `P` flag or the
   field. This mirrors D-25's `ignores_permission` typed field. Reshapes D-26; the
   Cat 4 permission family has one member, `PermissionGranter`
   (`permission_query → grant_permission`, Popular Movement). `B.009` still closes.
   Rejected: a fourth registrar route, a direct read of the bought definition (new
   machinery for one asset).
2. **A quality fires through `EffectsEngine.FireQuality`, raised by Buy Asset.** F2:
   `firePurchaseEffects` (`orchestrator.go:651`) fires `on_purchase` entries only for
   an `AssetAdded`, and a quality adds none. Buy Asset gains the effects engine (as
   Attack did, overview Decision 21) and fires the quality's one entry on the chosen
   target: the "fourth direct definition read" the overview names. `Registrar` gains
   `Targets`, so the registrar, not Buy Asset, defines what a quality's filters select
   (D-20). `RegisterEffectsList` refuses a quality whose effect has no `Targets`.
   Rejected: Buy Asset emits `AssetStealthApplied` on a `grant_stealth` check (the
   action keys on an effect name).
3. **A quality is never an asset.** F4: today Refit offers Stealth as a replacement
   for any Cunning asset, and the wizard offers it as a starting asset; both would
   make a 0-HP "Stealth" asset. Refit, the wizard and the purchase enabler skip
   `def.Quality` (the enabler since DR-5). A quality bought through the enabler stays
   Out of Scope.
4. **The government record is read through `FactionState.Ruler(worldID)`, and
   `IsIndependent(worldID)` names its empty answer.** One nil-safe lookup: a nil state
   or a missing entry is an independent world, one no tracked faction governs. The
   fiction still has a government there, and the GM answers for it (D-4). Storage
   stays derived (no entry, no sentinel); the code and the digest say "independent"
   by name. T-011 gets a name, `domain.TagPlanetaryGovernment`, and `Faction.HasTag`
   replaces the goal helper's own loop. Follows overview Decisions 8 and 10.
   *(Revised 2026-10-22, Robert's call: "independent" replaces "NPC", which read as
   "a different tracked faction".)*
5. **Planetary Seizure's completion takes the government and reads its XP from the
   previous ruler.** Robert's call on the XP. F3: Seize Planet never sets
   `TargetFactionID`, so `calcPlanetarySeizureXP("")` always returns 1; the book uses
   the ruling faction's ratings. The completion emits `WorldGovernmentChanged`,
   `TagAdded` only when the seizer lacks T-011, and `TagRemoved` for a previous ruler
   that governs no other world and holds the tag (overview Decision 10). A seizer
   that already rules the world changes no government; Seize Planet does not offer
   such a world (DR-3).
6. **Inside Enemy Territory counts a world that is a rival's ground.** Robert's call.
   A world counts when a rival tracked faction governs it, or when it is independent
   and a rival's homeworld. A world the acting faction governs never counts, a
   seized rival homeworld included ("it's ours now"). An independent world that is no
   rival's homeworld never counts. The faction's own homeworld counts only when a
   rival governs it. An asset with no world (in transit) does not count. Homeworld and
   government stay separate facts: a homeworld is where a faction's leadership sits,
   not a world it governs, so it never enters the record. Reshapes overview Decision
   10's "a rival rules this world". *(Revised 2026-10-22: the first version counted
   every world the acting faction does not rule, independent worlds included, a
   ruling made while "NPC" read as a tracked faction.)*
7. **The wizard asks for the governed world on a T-011 pick, and faction CRUD keeps
   the record consistent.** A hidden form group appears when T-011 is picked. Its
   options are the worlds no other faction governs, the homeworld first.
   `state.CreateFaction` takes the world and writes the entry in the same
   rollback-safe save, and refuses a world another faction governs (DR-6). F5: `DeleteFaction` would leave entries naming a faction that
   no longer exists, so it clears them, and restores them if the save fails.
   Resolves the overview OQ "the wizard's world prompt".
8. **The permission request is `hooks.PermissionCollector.RequestPermission(PermissionRequest) (bool, error)`,
   driven by `dispatch.RequestPermission`.** `hooks`, the lowest package with a
   collector, as `PaymentCollector` (Effort 3 Decision 5). The request names the
   requester, the ruler (nil for an independent world, so the GM answers, D-4), the
   world and the definition. `dispatch.RequestPermission` asks only for a definition
   that needs it, grants a purchase on the requester's homeworld (Decision 15), grants
   a ruler's own request (D-6), grants on a grant hook (Commit 5),
   and otherwise asks. It returns the refusal record, or nil. `action.Collector`
   embeds the interface now; `PhaseCollector` embeds it in Effort 5, when arrivals
   ask (overview Decision 18: a change rides with its first user). The prompt
   ignores Esc (Effort 3 Decision Record 7: a cancel drops earlier assets' results).
   Resolves the overview OQ "the D-41 request method".
9. **The request runs after the choice; a refusal is the action's only mutation.**
   Buy Asset and Refit ask in `Inputs`, after the order; Buy Asset asks before a
   quality's target, and a refusal asks for none (DR-4). The purchase enabler asks
   after its pick, and a refusal ends that asset's step with the record; earlier
   assets in the action keep their results (Effort 3 Decision Record 1). No Coin is
   charged; the action is used (D-35, overview Decision 22).
10. **`PermissionRefused{FactionID, WorldID, DefinitionID, DecidedByFactionID, Cause, CausedByFactionID}`.**
    `DecidedByFactionID` is the ruler, empty when the GM decided. `Cause` is the
    purchase path: `buy`, `refit` or `ability`. Record only: its apply arm changes
    nothing and records no miss. Resolves the overview OQ "`PermissionRefused`
    fields".
11. **No refit for an asset in transit.** Robert's call: an asset between worlds
    cannot be acted on. `validRefitReplacements` returns nothing for an asset with no
    world, which also removes "which government to ask" for it. The general rule (no
    action touches an in-transit asset) is filed as `F.028` for an audit.
12. **The Buy Asset menu offers `P` assets on every world.** D-35: permission is
    asked after the choice. `CanPurchase` loses its `P` case; `HasPlanetaryGovernment`,
    `PurchaseTerms.HoldsGovernment` and the enabler's `holdsGovernment` retire
    (discovery, Replaces / Retires).
13. **Digest and TUI wording (overview Decision 13).** `WorldGovernmentChanged` is a
    goal event on the new ruler's beat ("{faction} now governs {world}.");
    `TagRemoved` is a goal event on the loser's beat ("{faction} lost the tag:
    {tag}."); `PermissionRefused` is a refusal line on the requester's beat ("{decider}
    refused {faction} permission to raise {asset} on {world}.", "the independent
    government" for the GM). The TUI shows the government change and the refusal, and
    folds `TagRemoved` as it folds `TagAdded`.
14. **Five commits in four phases, one design surface each.** Phase 1 is the
    quality; Phase 2 the government record, engine (Commit 2) then the wizard
    (Commit 3); Phase 3 the request at three purchase paths; Phase 4 the rule
    queries. Commit 5 adds the registry to `RequestPermission` and to Refit's
    constructor, because the grant hook is their first user (overview Decision 18).
    Phases 1 to 3 pass the 800-line budget: the bulk is literal test code, and no
    green seam splits Commits 1, 2 or 4 further (a new mutation ships its digest and
    TUI lines in its commit, overview Decision 13; a new collector method needs every
    implementation in the same build).
15. **A faction never asks permission to purchase on its own homeworld.** Robert's
    call, reading the book's homeworld clause ("Factions can always purchase assets
    and take actions on their homeworld"). It holds whoever governs the homeworld, an
    independent government or a rival. `dispatch.RequestPermission` grants it before
    any other check but the `P` test, so Buy Asset, Refit and the purchase enabler
    share it. Purchases only: an arrival on the homeworld still asks (Effort 5), and a
    fence-sign at the clause marks that. Reverses the discovery's effect that a faction
    with no record entry for its homeworld is asked, through the GM, for a `P` asset
    there (User-Facing Impact). *(Added 2026-10-22.)*

## Decision Record — Execution

**DR-1 — `actions.md`'s world-engine Key Decision names the cost chain (Commit 4).**
The entry "Effect handlers receive the world engine by parameter" gave "the
purchase-enabler's government check" as a user of the world engine. Commit 4 retires
that check (Decision 12). The enabler still reads the world engine, through
`world.AssetsOnWorld` in its cost chain, so the example now says "the
purchase-enabler's cost chain". This is a doc edit the plan did not list, made so the
page agrees with the source.

**DR-2 — The Capital Fleet buy test moves the homeworld off the purchase world
(Commit 5).** Task 14's `TestBuyAsset_PurchasePermissionAsksForARaise` bought on
`Tartarus`, which was also the faction's homeworld. Decision 15 skips the ask on the
homeworld, so the expected `RequestPermission` call would never come. The test now
sets the homeworld to `Anchorage` and buys on `Tartarus`. The code is unchanged; the
plan's test contradicted its own Decision 15.

**DR-3 — Seize Planet does not offer a world the faction governs (interim review).**
Robert's call. Decision 5 reads the XP from the previous ruler, so a seizure of a
world the seizer already governs paid XP from the seizer's own ratings (3 XP at
6/6/6). Both copies of the target list, `seizePlanetTargetWorlds` (Validate) and the
adapter's `deriveContestedWorlds`, now skip a world the faction governs. This lifts
the Out of Scope item "Seize Planet's target list offering a world the faction already
rules". `takeGovernment` keeps its own-world guard.

**DR-4 — Buy Asset asks permission before the quality target (interim review).**
Commit 4 asked for the quality's target, then asked permission, so a restricted
quality could prompt for a target and then be refused. `Inputs` now asks permission
first, and a refusal asks for no target. No effect on the SWN data today: Stealth
needs no permission. Decision 9 still holds: the ask comes after the order.

**DR-5 — The purchase enabler skips a quality (interim review).** Decision 3 kept a
quality out of Refit and the wizard and relied on "no Force quality exists" for the
enabler. A GM rulebook can add one, and the enabler would then raise it as a 0-HP
asset whose `on_purchase` fires on itself. `eligiblePurchaseDefs` now skips
`def.Quality`, as Refit and the wizard do. Not in `domain.CanPurchase`: Buy Asset
offers qualities through it.

**DR-6 — `CreateFaction` refuses a world another faction governs (interim review).**
Decision 7 relied on the wizard's picker, which lists only the worlds no faction
governs. `CreateFaction` now returns `ErrWorldAlreadyGoverned` for a governed world,
so no caller can leave the old ruler with T-011 and no record. A save rollback is then
a plain delete of the entry.

**DR-7 — Capital Fleet carries the book's `A` flag (interim review).** The book gives
A, S (`swn-assets.md`); `force_assets.toml` had only `S`, a gap older than this
effort that the Commit 5 loader test's text exposed. With `A`, Capital Fleet appears in
Use Asset Ability through the GM-confirm fallback (no `[ability]` block), as the other
A-flagged assets without a built ability do.

**DR-8 — Review cleanups.** `FactionState`'s methods take the receiver
`factionState`, not `s`. `FactionState.WorldsRuledBy` (ID order) replaces
`state.worldsRuledBy` and the seizure's `governsNoOtherWorld`. This adds a second
accessor beside Decision 4's `Ruler`.

---

## Shared Context

- **Every commit ends green:** `go build ./... && go test ./internal/faction/...`.
- **Mock regeneration.** After a change to `action.Collector`, run
  `go generate ./internal/faction/engine/action/...`.
- **A new `AskKind`** goes before `AskKindCount` and gets an `askEntries` row;
  `TestAskEntriesCoverEveryKind` catches a missing row.
- **Test patterns.** Action tests: gomock `MockCollector`, `makeBuyRulebook`,
  `makeRefitRulebook`. Effects tests: `effects.Registrars()` through
  `effect.New(...)`. Mutation tests: `newMutationTestState`. Digest tests:
  `minState`, `eventRec`, `mutRec`, `buildOne`, `findBeat`.
- **Causes.** Seizure's government mutations use `Cause: "goal_completed"`, as its
  `TagAdded` does. Stealth's mutation uses its effect name, `grant_stealth`, as
  `apply_stealth` does.
- **The SWN rulebook only.** Asset and tag TOML changes go to `rulebooks/swn/`. The
  test campaign's copy is regenerated after the arc (overview Decision 8).

## Out of Scope

- Movement permission, the arrival hold, Lobbyists, Transport Lockdown, the bypasses,
  and `PhaseCollector` embedding `PermissionCollector` (Effort 5).
- A quality bought through the purchase enabler (no Force quality exists; the enabler
  skips one, DR-5).
- A wizard T-011 pick when every world is governed (the select has no options).
- The general in-transit rule (`F.028`, filed in Commit 4).
- The `"G.012"` half of `B.001` (overview Decision 20).

---

## Work Breakdown

**Overview**

- **Phase 1 — Stealth as a quality**
  - Commit 1 — `refactor: Stealth as a purchase quality` (12 tasks)
- **Phase 2 — The government record**
  - Commit 2 — `feat: world government record` (16 tasks)
  - Commit 3 — `feat: the wizard records a world's government` (5 tasks)
- **Phase 3 — The permission request**
  - Commit 4 — `feat: permission request at purchase` (22 tasks)
- **Phase 4 — Permission rule queries**
  - Commit 5 — `feat: permission rule queries` (15 tasks)

### Phase 1 — Stealth as a quality

**In plain words:** Stealth, which hides an asset from other factions, stops being a
special case in the code. The rulebook data now describes it as an upgrade that a
faction buys for an asset it already owns.

#### Commit 1 — `refactor: Stealth as a purchase quality`

##### Task 1 — `internal/faction/domain/asset.go`

In `AssetDefinition`, add the quality mark (D-19).

**Find:**
```go
	Transport   *TransportProfile
	Effects     []EffectDefinition
}
```
**Replace with:**
```go
	Transport   *TransportProfile
	Effects     []EffectDefinition
	// Quality marks a definition that adds no asset: buying it runs its one
	// on_purchase entry on an asset the buyer already holds (Stealth).
	Quality bool
}
```

##### Task 2 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialShelterAsset   SpecialEffectType = "shelter_asset"
)
```
**Replace with:**
```go
	SpecialShelterAsset   SpecialEffectType = "shelter_asset"
	SpecialGrantStealth   SpecialEffectType = "grant_stealth"
)
```

##### Task 3 — `internal/faction/rulebook/effects.go`

Add the effect to the loader's accepted list.

**Find:**
```go
		domain.SpecialRestoreAsset, domain.SpecialDiscountRepair, domain.SpecialSeizeTarget,
	}
```
**Replace with:**
```go
		domain.SpecialRestoreAsset, domain.SpecialDiscountRepair, domain.SpecialSeizeTarget,
		domain.SpecialGrantStealth,
	}
```

##### Task 4 — `internal/faction/rulebook/rulebook.go`

Three edits. In `assetRecord`:

**Find:**
```go
	Transport   *transportRecord `toml:"transport"`
	Effects     []effectRecord   `toml:"effects"`
}
```
**Replace with:**
```go
	Transport   *transportRecord `toml:"transport"`
	Effects     []effectRecord   `toml:"effects"`
	Quality     bool             `toml:"quality"`
}
```

In `convertAsset`, check the quality after the no-damage check:

**Find:**
```go
		return nil, fmt.Errorf("attack has no damage and no own_attack_hit effect")
	}

	return &domain.AssetDefinition{
```
**Replace with:**
```go
		return nil, fmt.Errorf("attack has no damage and no own_attack_hit effect")
	}
	if err := checkQuality(r.Quality, effects); err != nil {
		return nil, err
	}

	return &domain.AssetDefinition{
```

**Find:**
```go
		Transport:   transport,
		Effects:     effects,
	}, nil
}
```
**Replace with:**
```go
		Transport:   transport,
		Effects:     effects,
		Quality:     r.Quality,
	}, nil
}

// checkQuality holds a quality to one on_purchase entry (D-19), and grant_stealth to
// a quality: the effect acts on an asset the buyer already holds, so on an asset of
// its own it would fire on the asset just bought.
func checkQuality(quality bool, effects []domain.EffectDefinition) error {
	if quality && (len(effects) != 1 || effects[0].Trigger != domain.TriggerOnPurchase) {
		return fmt.Errorf("a quality needs exactly one on_purchase effect")
	}
	if !quality && slices.ContainsFunc(effects, func(entry domain.EffectDefinition) bool {
		return entry.Effect == domain.SpecialGrantStealth
	}) {
		return fmt.Errorf("grant_stealth needs quality = true")
	}
	return nil
}
```

##### Task 5 — `internal/faction/engine/effect/registrar.go`

Two edits. The registrar gains the quality's target route (D-20).

**Find:**
```go
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
**Replace with:**
```go
// means the effect has no such route. Validate is optional: it refuses a
// combination of fields that Accepts, which checks each field alone, cannot express.
// Targets lists the assets a quality's entry may act on; only an effect a quality
// can carry sets it.
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
	Matchup  func(MatchupContext) MatchupOutcome
	Validate func(domain.EffectDefinition) error
	Targets  func(TargetContext) []*domain.Asset
}
```

**Find:**
```go
// MatchupContext is one matchup of the Attack action, seen from the declaring
```
**Replace with:**
```go
// TargetContext is a quality's purchase before it is made: the buyer, the world it
// buys on, and the quality's one entry. Rulebook rides along because a filter such
// as asset_type reads the candidate's definition.
type TargetContext struct {
	Faction  *domain.Faction
	WorldID  string
	Entry    domain.EffectDefinition
	Rulebook *rulebook.Rulebook
}

// MatchupContext is one matchup of the Attack action, seen from the declaring
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"` to the
imports.

##### Task 6 — `internal/faction/engine/effect/effect.go`

Two edits. `RegisterEffectsList` refuses a quality with no target route:

**Find:**
```go
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
	}
	return e.Register(&effectsListHandler{def: def, registrars: e.registrars})
}
```
**Replace with:**
```go
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
	}
	if def.Quality && (len(def.Effects) != 1 || e.registrars[def.Effects[0].Effect].Targets == nil) {
		return fmt.Errorf("asset %q: a quality needs one effect with a Targets route", def.ID)
	}
	return e.Register(&effectsListHandler{def: def, registrars: e.registrars})
}
```

Then add the quality route after `FirePurchase`:

**Find:**
```go
// FireMatchup runs the declaring asset's entries for trigger through each
```
**Replace with:**
```go
// QualityTargets lists the assets faction holds on worldID that quality def may act
// on, as its registrar reads the entry's filters. A definition that is not a quality
// has none.
func (e *EffectsEngine) QualityTargets(def *domain.AssetDefinition, faction *domain.Faction, worldID string, rulebook *rulebook.Rulebook) []*domain.Asset {
	if !def.Quality || len(def.Effects) != 1 {
		return nil
	}
	entry := def.Effects[0]
	registrar, ok := e.registrars[entry.Effect]
	if !ok || registrar.Targets == nil {
		return nil
	}
	return registrar.Targets(TargetContext{Faction: faction, WorldID: worldID, Entry: entry, Rulebook: rulebook})
}

// FireQuality runs quality def's one entry on target through its phase route. A
// direct definition read raised by Buy Asset: a quality adds no asset, so
// FirePurchase never sees it. FireContext.Asset is the target, not a declaring asset.
func (e *EffectsEngine) FireQuality(faction *domain.Faction, def *domain.AssetDefinition, target *domain.Asset, factionState *state.FactionState) []domain.Mutation {
	entry := def.Effects[0]
	// deliberate: no roller, because no quality effect rolls; give FireQuality one when one does.
	return e.registrars[entry.Effect].Fire(FireContext{
		Faction:      faction,
		Asset:        target,
		Def:          def,
		Entry:        entry,
		FactionState: factionState,
	})
}

// FireMatchup runs the declaring asset's entries for trigger through each
```

##### Task 7 — `internal/faction/engine/effect/effects/grant_stealth.go` (new)

Full contents:
```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// grantStealthRegistrar is Stealth, a quality bought for another asset (D-19 to
// D-21). Its filters select the asset it acts on (D-20): owner = "own" reads as "the
// buyer holds it", scope = "world" as "it stands on the world of the purchase", and
// asset_type and stealthy narrow it.
var grantStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType, domain.FilterStealthy},
		},
	},
	Fire:    fireGrantStealth,
	Targets: grantStealthTargets,
}

// grantStealthTargets lists the buyer's assets on the world that the filters
// select, in ID order.
func grantStealthTargets(ctx effect.TargetContext) []*domain.Asset {
	if ctx.WorldID == "" {
		return nil
	}
	filters := ctx.Entry.Filters
	var targets []*domain.Asset
	for _, asset := range domain.SortedAssets(ctx.Faction) {
		if asset.Location.WorldID != ctx.WorldID {
			continue
		}
		if filters.Stealthy != nil && asset.Stealthy != *filters.Stealthy {
			continue
		}
		def := ctx.Rulebook.Assets[asset.DefinitionID]
		if def == nil || (len(filters.AssetTypes) > 0 && !slices.Contains(filters.AssetTypes, def.Type)) {
			continue
		}
		targets = append(targets, asset)
	}
	return targets
}

func fireGrantStealth(ctx effect.FireContext) []domain.Mutation {
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialGrantStealth),
		CausedByFactionID: ctx.Faction.ID,
	}}
}
```

In `effects/registrars.go`:

**Find:**
```go
		domain.SpecialSeizeTarget:    seizeTargetRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialSeizeTarget:    seizeTargetRegistrar,
		domain.SpecialGrantStealth:   grantStealthRegistrar,
	}
```

##### Task 8 — `internal/faction/engine/action/actions/buy_asset.go`

Rewritten: every `SWN-C3-002` check goes (D-19 to D-22, overview Decision 20). Full
contents:
```go
package actions

import (
	"fmt"
	"math"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// BuyAsset purchases a new asset and places it on a target world.
// The asset is flagged inactive (Ready: false) until the start of the next turn.
// A quality (Stealth) adds no asset: its effect acts on one the faction holds there.
type BuyAsset struct {
	collector        action.Collector
	registry         *hooks.Registry
	worldEngine      *world.WorldEngine
	effectsEngine    *effect.EffectsEngine
	factionID        string
	buyOrder         action.BuyOrder
	newAsset         domain.Asset
	qualityTarget    *domain.Asset
	qualityMutations []domain.Mutation
	quote            hooks.AssetCostQuery
	cost             int
	discountConsumed bool
}

func NewBuyAsset(collector action.Collector, registry *hooks.Registry, worldEngine *world.WorldEngine, effectsEngine *effect.EffectsEngine) *BuyAsset {
	return &BuyAsset{collector: collector, registry: registry, worldEngine: worldEngine, effectsEngine: effectsEngine}
}

func (ba *BuyAsset) Name() string { return "Buy Asset" }

func (ba *BuyAsset) Validate(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) bool {
	for _, worldID := range availableWorlds(faction) {
		if len(purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry, ba.effectsEngine)) > 0 {
			return true
		}
	}
	return false
}

func (ba *BuyAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
	purchasableByWorld := make(map[string][]*domain.AssetDefinition)
	for _, worldID := range availableWorlds(faction) {
		defs := purchasableDefinitions(faction, rulebook, worldID, ba.worldEngine, ba.registry, ba.effectsEngine)
		if len(defs) > 0 {
			purchasableByWorld[worldID] = defs
		}
	}

	order, err := ba.collector.SelectBuyOrder(purchasableByWorld)
	if err != nil {
		return fmt.Errorf("buy asset: %w", err)
	}
	ba.factionID = faction.ID
	ba.buyOrder = order

	if order.Definition.Quality {
		target, err := ba.selectQualityTarget(faction, rulebook)
		if err != nil {
			return fmt.Errorf("buy asset: %w", err)
		}
		ba.qualityTarget = target
	}
	return nil
}

func (ba *BuyAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	def := ba.buyOrder.Definition
	ba.quote = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost, world.AssetsOnWorld(ba.worldEngine, ba.buyOrder.World))
	ba.cost = ba.quote.Cost
	if faction.PendingPurchaseDiscount > 0 {
		ba.cost = dispatch.DiscountedCost(ba.quote, def.Cost, faction.PendingPurchaseDiscount)
		ba.discountConsumed = true
	}
	if faction.Coin < ba.cost {
		return fmt.Errorf("insufficient Coin: need %d, have %d", ba.cost, faction.Coin)
	}
	if def.Quality {
		ba.qualityMutations = ba.effectsEngine.FireQuality(faction, def, ba.qualityTarget, factionState)
		return nil
	}
	ba.newAsset = domain.NewPurchasedAsset(factionState.Factions, faction, def, ba.buyOrder.World)
	return nil
}

func (ba *BuyAsset) Output() ([]domain.Mutation, error) {
	mutations := dispatch.PurchaseCharges(ba.factionID, ba.cost, ba.quote, "buy")
	if ba.buyOrder.Definition.Quality {
		mutations = append(mutations, ba.qualityMutations...)
	} else {
		mutations = append([]domain.Mutation{
			domain.AssetAdded{FactionID: ba.factionID, Asset: ba.newAsset, Cause: "buy", CausedByFactionID: ba.factionID},
		}, mutations...)
	}
	if ba.discountConsumed {
		mutations = append(mutations, domain.PurchaseDiscountConsumed{
			FactionID:         ba.factionID,
			Cause:             "buy",
			CausedByFactionID: ba.factionID,
		})
	}
	return mutations, nil
}

// selectQualityTarget picks the asset a quality acts on: the one candidate without
// a prompt, else the faction's choice. The menu offers a quality only where it has
// a candidate (D-22).
func (ba *BuyAsset) selectQualityTarget(faction *domain.Faction, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
	def := ba.buyOrder.Definition
	targets := ba.effectsEngine.QualityTargets(def, faction, ba.buyOrder.World, rulebook)
	switch len(targets) {
	case 0:
		return nil, fmt.Errorf("%s has no target on %s", def.Name, ba.buyOrder.World)
	case 1:
		return targets[0], nil
	}
	target, err := ba.collector.SelectAsset(targets, rulebook)
	if err != nil {
		return nil, fmt.Errorf("select %s target: %w", def.Name, err)
	}
	if target == nil {
		return nil, fmt.Errorf("select %s target: %w", def.Name, action.ErrNoSelection)
	}
	return target, nil
}

// availableWorlds returns the faction's homeworld plus the unique set of
// worlds where the faction already has assets.
func availableWorlds(faction *domain.Faction) []string {
	seen := map[string]struct{}{faction.Homeworld.WorldID: {}}
	for _, asset := range faction.Assets {
		seen[asset.Location.WorldID] = struct{}{}
	}
	worlds := make([]string, 0, len(seen))
	for world := range seen {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	return worlds
}

// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum attribute rating, within
// the world's tech level, and holding Planetary Government for P-flagged defs. A
// quality also needs a target on the world (D-22). An unknown
// world imposes no tech ceiling (math.MaxInt) but cannot confer government, so P-flagged
// defs are rejected — matching the purchase_enabler path.
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	holdsGovernment := false
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
		if worldEngine.Index != nil {
			holdsGovernment = domain.HasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID])
		}
	}

	onWorld := world.AssetsOnWorld(worldEngine, fragmentID)
	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if def.Quality && (effectsEngine == nil || len(effectsEngine.QualityTargets(def, faction, fragmentID, rulebook)) == 0) {
			continue
		}
		techCap := worldTechLevel
		if worldEngine != nil {
			techCap = dispatch.ResolveWorldTechLevel(registry, faction, def, fragmentID, worldTechLevel)
		}
		cost := dispatch.ResolveAssetCost(registry, faction, def, fragmentID, def.Cost, onWorld).Cost
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap, HoldsGovernment: holdsGovernment}) {
			result = append(result, def)
		}
	}
	return result
}
```

##### Task 9 — `internal/faction/engine/action/actions/register.go`

Two edits.

**Find:**
```go
// The Attack action raises the matchup route itself, so it takes the effects engine.
```
**Replace with:**
```go
// The Attack action raises the matchup route itself, and Buy Asset fires a quality's
// effect itself, so both take the effects engine.
```

**Find:**
```go
	ae.Register(func(c action.Collector) action.Action { return NewBuyAsset(c, hookRegistry, worldEngine) })
```
**Replace with:**
```go
	ae.Register(func(c action.Collector) action.Action {
		return NewBuyAsset(c, hookRegistry, worldEngine, effectsEngine)
	})
```

##### Task 10 — `internal/faction/engine/action/actions/refit_asset.go` and `internal/faction/tui/views/manage/wizard/wizard.go`

A quality is never an asset (Decision 3). In `validRefitReplacements`:

**Find:**
```go
	for _, def := range rulebook.Assets {
		if def.ID == oldDef.ID {
			continue
		}
```
**Replace with:**
```go
	for _, def := range rulebook.Assets {
		if def.ID == oldDef.ID || def.Quality {
			continue
		}
```

In `wizard.go`, `buildAssetsForm`:

**Find:**
```go
	for _, def := range rb.Assets {
		allDefs = append(allDefs, def)
	}
```
**Replace with:**
```go
	for _, def := range rb.Assets {
		if def.Quality {
			continue
		}
		allDefs = append(allDefs, def)
	}
```

##### Task 11 — `rulebooks/swn/assets/cunning_assets.toml`

Stealth becomes a quality with its `grant_stealth` entry.

**Find:**
```toml
flags = ["S"]
counter = ""
description = "Not a standard asset but a special quality purchased for another Special Forces asset on the planet. A Stealthed asset cannot be detected or attacked. If the unit normally requires government permission to land, that permission may be foregone. Stealth is lost if the asset attacks or defends."
```
**Replace with:**
```toml
flags = ["S"]
quality = true
counter = ""
description = "Not a standard asset but a special quality purchased for another Special Forces asset on the planet. A Stealthed asset cannot be detected or attacked. If the unit normally requires government permission to land, that permission may be foregone. Stealth is lost if the asset attacks or defends."

[[assets.Stealth.effects]]
trigger = "on_purchase"
effect = "grant_stealth"
[assets.Stealth.effects.filters]
owner = "own"
scope = "world"
asset_type = ["Special Forces"]
stealthy = false
```

##### Task 12 — tests and docs

`internal/faction/rulebook/effects_test.go` — append:
```go
func TestLoadEffects_QualityNeedsOneOnPurchaseEntry(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+"quality = true\n")
	if err == nil || !strings.Contains(err.Error(), "a quality needs exactly one on_purchase effect") {
		t.Fatalf("err = %v, want a quality shape error", err)
	}
}

func TestLoadEffects_GrantStealthNeedsAQuality(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+`
[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "grant_stealth"
[assets.Machine.effects.filters]
owner = "own"
scope = "world"
`)
	if err == nil || !strings.Contains(err.Error(), "grant_stealth needs quality = true") {
		t.Fatalf("err = %v, want a grant_stealth quality error", err)
	}
}
```

`internal/faction/rulebook/rulebook_test.go` — in `TestLoad`:

**Find:**
```go
	t.Run("goals populated", func(t *testing.T) {
```
**Replace with:**
```go
	t.Run("stealth is a quality with grant_stealth", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C3-002"]
		if !ok {
			t.Fatal("asset SWN-C3-002 not found")
		}
		if !asset.Quality {
			t.Error("Quality = false, want true")
		}
		stealthy := false
		want := []domain.EffectDefinition{{
			Trigger: domain.TriggerOnPurchase,
			Effect:  domain.SpecialGrantStealth,
			Filters: domain.EffectFilters{
				Owner:      domain.OwnerOwn,
				Scope:      domain.ScopeWorld,
				AssetTypes: []domain.AssetType{domain.TypeSpecialForces},
				Stealthy:   &stealthy,
			},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("goals populated", func(t *testing.T) {
```

`internal/faction/engine/effect/effect_test.go` — append:
```go
func TestRegisterEffectsList_QualityNeedsTargets(t *testing.T) {
	entry := domain.EffectDefinition{Trigger: domain.TriggerOnPurchase, Effect: domain.SpecialApplyStealth, Filters: ownSelf}
	def := &domain.AssetDefinition{ID: "Q-001", Quality: true, Effects: []domain.EffectDefinition{entry}}

	withoutTargets := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Fire:     noopFire,
		},
	}
	if err := effect.New(withoutTargets, discardLog).RegisterEffectsList(def); err == nil {
		t.Error("a quality whose effect has no Targets registered; want an error")
	}

	withTargets := map[domain.SpecialEffectType]effect.Registrar{
		domain.SpecialApplyStealth: {
			Triggers: map[domain.EffectTrigger]effect.Accepts{domain.TriggerOnPurchase: acceptsOwnSelf},
			Fire:     noopFire,
			Targets:  func(effect.TargetContext) []*domain.Asset { return nil },
		},
	}
	if err := effect.New(withTargets, discardLog).RegisterEffectsList(def); err != nil {
		t.Errorf("RegisterEffectsList: %v", err)
	}
}
```

`internal/faction/engine/effect/effects/grant_stealth_test.go` (new) — full contents:
```go
package effects_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func stealthQuality() *domain.AssetDefinition {
	stealthy := false
	return &domain.AssetDefinition{ID: "SWN-C3-002", Name: "Stealth", Quality: true, Effects: []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialGrantStealth,
		Filters: domain.EffectFilters{
			Owner:      domain.OwnerOwn,
			Scope:      domain.ScopeWorld,
			AssetTypes: []domain.AssetType{domain.TypeSpecialForces},
			Stealthy:   &stealthy,
		},
	}}}
}

func TestGrantStealth_TargetsUnstealthedSpecialForcesOnTheWorld(t *testing.T) {
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"sf":  {ID: "sf", Type: domain.TypeSpecialForces},
		"mil": {ID: "mil", Type: domain.TypeMilitaryUnit},
	}}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{
		"a-2": {ID: "a-2", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}},
		"a-1": {ID: "a-1", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}},
		"a-3": {ID: "a-3", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}, Stealthy: true},
		"a-4": {ID: "a-4", DefinitionID: "mil", Location: domain.Location{WorldID: "Krylos"}},
		"a-5": {ID: "a-5", DefinitionID: "sf", Location: domain.Location{WorldID: "Tartarus"}},
	}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	var ids []string
	for _, target := range engine.QualityTargets(stealthQuality(), faction, "Krylos", rb) {
		ids = append(ids, target.ID)
	}
	if !reflect.DeepEqual(ids, []string{"a-1", "a-2"}) {
		t.Errorf("targets = %v, want [a-1 a-2] (unstealthed Special Forces on Krylos, in ID order)", ids)
	}
}

func TestGrantStealth_StealthsTheTarget(t *testing.T) {
	faction := &domain.Faction{ID: "alpha"}
	target := &domain.Asset{ID: "a-1", OwnerID: "alpha"}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	got := engine.FireQuality(faction, stealthQuality(), target, nil)
	want := []domain.Mutation{domain.AssetStealthApplied{FactionID: "alpha", AssetID: "a-1", Cause: "grant_stealth", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}
```

`internal/faction/engine/action/actions/buy_asset_test.go` — every
`NewBuyAsset(x, y, z)` call gains a fourth argument `nil`, and both
`purchasableDefinitions(..., worldEngine, registry)` calls gain a trailing `nil`.
Add `effect` and `effects` to the imports
(`.../engine/effect`, `.../engine/effect/effects`) and append:
```go
func stealthBuyRulebook() *rulebook.Rulebook {
	stealthy := false
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"stealth": {ID: "stealth", Name: "Stealth", Category: domain.StatCunning, Cost: 2, MinRating: 1, Quality: true,
			Effects: []domain.EffectDefinition{{
				Trigger: domain.TriggerOnPurchase,
				Effect:  domain.SpecialGrantStealth,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld,
					AssetTypes: []domain.AssetType{domain.TypeSpecialForces}, Stealthy: &stealthy},
			}}},
		"sf": {ID: "sf", Name: "Saboteurs", Category: domain.StatCunning, Cost: 5, HP: 6, MinRating: 9, Type: domain.TypeSpecialForces},
	}}
}

func stealthBuyer(withTarget bool) *domain.Faction {
	faction := &domain.Faction{ID: "f1", Cunning: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}, Assets: map[string]*domain.Asset{}}
	if withTarget {
		faction.Assets["f1-sf-1"] = &domain.Asset{ID: "f1-sf-1", DefinitionID: "sf", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	}
	return faction
}

func TestBuyAsset_QualityOfferedOnlyWithATarget(t *testing.T) {
	effectsEngine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	rb := stealthBuyRulebook()

	if NewBuyAsset(nil, nil, nil, effectsEngine).Validate(stealthBuyer(false), nil, rb) {
		t.Error("Validate = true with no Special Forces to stealth, want false (D-22)")
	}
	if !NewBuyAsset(nil, nil, nil, effectsEngine).Validate(stealthBuyer(true), nil, rb) {
		t.Error("Validate = false with a Special Forces asset to stealth, want true")
	}
}

func TestBuyAsset_QualityStealthsItsTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	rb := stealthBuyRulebook()
	faction := stealthBuyer(true)
	effectsEngine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: "Tartarus", Definition: rb.Assets["stealth"]}, nil)

	act := NewBuyAsset(collector, nil, nil, effectsEngine)
	if err := act.Inputs(faction, nil, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "buy", CausedByFactionID: "f1"},
		domain.AssetStealthApplied{FactionID: "f1", AssetID: "f1-sf-1", Cause: "grant_stealth", CausedByFactionID: "f1"},
	}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v (no AssetAdded for a quality)", mutations, want)
	}
}
```

`internal/faction/engine/action/actions/refit_asset_test.go` — in
`TestRefitAsset_Validate`:

**Find:**
```go
	t.Run("valid refit exists", func(t *testing.T) {
```
**Replace with:**
```go
	t.Run("a quality is never a replacement", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// built from makeRefitRulebook: the test's local rulebook shadows the package name
		qualityRulebook := makeRefitRulebook()
		delete(qualityRulebook.Assets, "new-def")
		qualityRulebook.Assets["stealth"] = &domain.AssetDefinition{ID: "stealth", Category: domain.StatForce, Cost: 2, Quality: true}
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{WorldID: "Tartarus"}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, qualityRulebook) {
			t.Error("expected false when the only other definition is a quality")
		}
	})

	t.Run("valid refit exists", func(t *testing.T) {
```

`docs/architecture/engine/actions.md` — the Buy Asset row:

**Find:**
```markdown
| Buy Asset | Purchase an asset onto a world, inactive until next turn | `SelectBuyOrder` (`SelectAsset` for a stealth target) → `AssetAdded`, `CoinDelta` (`AssetStealthApplied`, `PurchaseDiscountConsumed`) |
```
**Replace with:**
```markdown
| Buy Asset | Purchase an asset onto a world, inactive until next turn; a quality (Stealth) adds no asset and acts on one the faction holds there | `SelectBuyOrder` (`SelectAsset` for a quality's target) → `AssetAdded`, `CoinDelta` (a quality's effect mutations, such as `AssetStealthApplied`, in place of `AssetAdded`; `PurchaseDiscountConsumed`) |
```

`docs/architecture/engine/effect-mutation.md`:

**Find:**
```markdown
the reactors. So Psychic Assassins enter play stealthed, and Tripwire Cells see
them arrive. A third route,
```
**Replace with:**
```markdown
the reactors. So Psychic Assassins enter play stealthed, and Tripwire Cells see
them arrive. A quality (`quality = true`, Stealth) adds no asset, so Buy Asset
fires its one `on_purchase` entry itself (`FireQuality`), on a target picked from
the registrar's `Targets`. A third route,
```

`docs/architecture/engine/ability-catalog.md`:

**Find:**
```markdown
| **grant-stealth** (quality bought for another asset) | Cat 4 `RuleModifier` | Stealth `SWN-C3-002` |
```
**Replace with:**
```markdown
| **grant-stealth** (quality bought for another asset) | purchase quality: Buy Asset fires its `on_purchase` entry on a chosen asset (`FireQuality`) → `grant_stealth` | Stealth `SWN-C3-002` ✅ |
```

##### Commit message

```
refactor: Stealth as a purchase quality

- quality = true: a definition that adds no asset; the loader holds it to one
  on_purchase entry, and grant_stealth to a quality
- Registrar.Targets and EffectsEngine.QualityTargets / FireQuality: Buy Asset
  fires the quality's entry on a target the registrar selects
- grant_stealth: Stealth onto data; every SWN-C3-002 check in Buy Asset goes
- a quality is offered only with a target (D-22), never as a refit or a
  starting asset
```

### Phase 2 — The government record

**In plain words:** The saved campaign now records which faction governs each world.
A faction that completes its goal to seize a world becomes its government, the GM
can set the government when they create or edit a world, and no purchase asks for
permission yet.

#### Commit 2 — `feat: world government record`

##### Task 1 — `internal/faction/state/faction_state.go`

The record (overview Decision 8) and its one lookup (Decision 4).

**Find:**
```go
type FactionState struct {
	CampaignID  string                     `toml:"campaign_id"`
	CycleNumber int                        `toml:"cycle_number"`
	Factions    map[string]*domain.Faction `toml:"factions"`
	CurrentTurn *domain.TurnState          `toml:"current_turn,omitempty"`
}
```
**Replace with:**
```go
type FactionState struct {
	CampaignID  string                     `toml:"campaign_id"`
	CycleNumber int                        `toml:"cycle_number"`
	Factions    map[string]*domain.Faction `toml:"factions"`
	// Governments maps a world ID to the faction that rules it. A world with no
	// entry is independent: no tracked faction governs it (A.001.6 overview Decision 8).
	Governments map[string]string `toml:"governments,omitempty"`
	CurrentTurn *domain.TurnState `toml:"current_turn,omitempty"`
}

// Ruler returns the ID of the faction that governs worldID, or "" for an
// independent world. A nil state has no governments.
func (s *FactionState) Ruler(worldID string) string {
	if s == nil {
		return ""
	}
	return s.Governments[worldID]
}

// IsIndependent reports whether no tracked faction governs worldID.
func (s *FactionState) IsIndependent(worldID string) bool {
	return s.Ruler(worldID) == ""
}
```

##### Task 2 — `internal/faction/domain/faction.go`

T-011 gets a name, and a faction can say whether it holds a tag (Decision 4).

**Find:**
```go
type Tag struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}
```
**Replace with:**
```go
type Tag struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}

// TagPlanetaryGovernment marks a faction that governs at least one world; the
// government record says which (A.001.6 overview Decision 10).
const TagPlanetaryGovernment = "T-011"
```

**Find:**
```go
// Clone deep-copies a faction so a caller can apply mutations to the copy without
```
**Replace with:**
```go
// HasTag reports whether the faction holds the tag with ID tagID.
func (faction *Faction) HasTag(tagID string) bool {
	return slices.ContainsFunc(faction.Tags, func(tag *Tag) bool { return tag.ID == tagID })
}

// Clone deep-copies a faction so a caller can apply mutations to the copy without
```

##### Task 3 — `internal/faction/domain/mutation.go`

**Find:**
```go
func (m TagAdded) Type() string { return "tag_added" }
```
**Replace with:**
```go
func (m TagAdded) Type() string { return "tag_added" }

// TagRemoved takes one tag from a faction. The full Tag rides along, as on
// TagAdded, so the digest can name it after the faction no longer holds it.
type TagRemoved struct {
	FactionID         string `json:"faction_id"`
	Tag               Tag    `json:"tag"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m TagRemoved) Type() string { return "tag_removed" }

// WorldGovernmentChanged sets the faction that governs a world. An empty
// ToFactionID makes the world independent; FromFactionID is the ruler it replaces,
// empty for an independent world.
type WorldGovernmentChanged struct {
	WorldID           string `json:"world_id"`
	FromFactionID     string `json:"from_faction_id"`
	ToFactionID       string `json:"to_faction_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m WorldGovernmentChanged) Type() string { return "world_government_changed" }
```

##### Task 4 — `internal/faction/engine/mutation/mutation.go`

Two edits.

**Find:**
```go
	EntityID     string // asset or base ID; empty for faction-level misses
```
**Replace with:**
```go
	EntityID     string // asset, base or tag ID; empty for faction-level misses
```

**Find:**
```go
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				tag := v.Tag
				faction.Tags = append(faction.Tags, &tag)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
```
**Replace with:**
```go
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				tag := v.Tag
				faction.Tags = append(faction.Tags, &tag)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.IndexFunc(faction.Tags, func(tag *domain.Tag) bool { return tag.ID == v.Tag.ID })
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.Tag.ID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.WorldGovernmentChanged:
			_, known := factionState.Factions[v.ToFactionID]
			switch {
			case v.ToFactionID == "":
				delete(factionState.Governments, v.WorldID)
			case !known:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.ToFactionID})
			default:
				if factionState.Governments == nil {
					factionState.Governments = make(map[string]string)
				}
				factionState.Governments[v.WorldID] = v.ToFactionID
			}
```

##### Task 5 — `internal/faction/engine/goal/goals/planetary_seizure.go`

The completion takes the government, and the XP reads the previous ruler (Decision 5).

**Find:**
```go
	newTurns := goal.TurnsRemaining - 1
	if newTurns == 0 {
		xp := calcPlanetarySeizureXP(goal.TargetFactionID, factionState)
		var pgTag domain.Tag
		if t, ok := rulebook.Tags["T-011"]; ok {
			pgTag = *t
		}
		return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{
			tick,
			domain.TagAdded{FactionID: faction.ID, Tag: pgTag, Cause: "goal_completed"},
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		}
	}
	return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{tick}
}
```
**Replace with:**
```go
	newTurns := goal.TurnsRemaining - 1
	if newTurns == 0 {
		worldID := goal.TargetWorld.WorldID
		previousRuler := factionState.Ruler(worldID)
		xp := calcPlanetarySeizureXP(previousRuler, factionState)
		mutations := []domain.Mutation{tick}
		mutations = append(mutations, takeGovernment(faction, worldID, previousRuler, factionState, rulebook)...)
		mutations = append(mutations,
			domain.GoalCompleted{FactionID: faction.ID, GoalID: goal.GoalID, XPAwarded: xp},
			domain.XPAwarded{FactionID: faction.ID, Amount: xp, Cause: "goal_completed"},
		)
		return locks.GoalLock{Type: locks.LockNone}, mutations
	}
	return locks.GoalLock{Type: locks.LockNone}, []domain.Mutation{tick}
}

// takeGovernment makes seizer the ruler of worldID (overview Decision 10): the
// record changes hands, the seizer gains T-011 unless it holds it, and a previous
// ruler that governs no other world loses the tag it holds. A seizer that already
// rules the world changes nothing.
func takeGovernment(seizer *domain.Faction, worldID, previousRuler string, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	if previousRuler == seizer.ID {
		return nil
	}
	mutations := []domain.Mutation{domain.WorldGovernmentChanged{
		WorldID:           worldID,
		FromFactionID:     previousRuler,
		ToFactionID:       seizer.ID,
		Cause:             "goal_completed",
		CausedByFactionID: seizer.ID,
	}}
	if !seizer.HasTag(domain.TagPlanetaryGovernment) {
		var tag domain.Tag
		if found, ok := rulebook.Tags[domain.TagPlanetaryGovernment]; ok {
			tag = *found
		}
		mutations = append(mutations, domain.TagAdded{FactionID: seizer.ID, Tag: tag, Cause: "goal_completed", CausedByFactionID: seizer.ID})
	}
	previous := factionState.Factions[previousRuler]
	if previous == nil || !governsNoOtherWorld(factionState, previousRuler, worldID) {
		return mutations
	}
	for _, held := range previous.Tags {
		if held.ID == domain.TagPlanetaryGovernment {
			return append(mutations, domain.TagRemoved{FactionID: previousRuler, Tag: *held, Cause: "goal_completed", CausedByFactionID: seizer.ID})
		}
	}
	return mutations
}

// governsNoOtherWorld reports whether factionID governs no world but worldID.
func governsNoOtherWorld(factionState *state.FactionState, factionID, worldID string) bool {
	for ruledWorld, ruler := range factionState.Governments {
		if ruler == factionID && ruledWorld != worldID {
			return false
		}
	}
	return true
}
```

##### Task 6 — `internal/faction/engine/goal/goals/inside_enemy_territory.go` and `helpers.go`

Inside Enemy Territory reads the record and the rivals' homeworlds (Decision 6). In
`inside_enemy_territory.go`:

**Find:**
```go
func (InsideEnemyTerritory) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, index *world.Index) []domain.Mutation {
```
**Replace with:**
```go
// UpdateProgress counts each asset stealthed this action on a rival's ground
// (Effort 4 Decision 6). An asset in transit has no world and does not count.
func (InsideEnemyTerritory) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
```

**Find:**
```go
		if !rivalHasPlanetaryGovernmentOnWorld(actingFaction.ID, asset.Location.WorldID, factionState, index) {
			continue
		}
```
**Replace with:**
```go
		if asset.Location.WorldID == "" || !isEnemyTerritory(actingFaction.ID, asset.Location.WorldID, factionState) {
			continue
		}
```

**Find:**
```go
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}
```
**Replace with:**
```go
	return append([]domain.Mutation{progressed}, completeGoal(actingFaction, 2)...)
}

// isEnemyTerritory reports whether worldID is a rival's ground: a rival governs it,
// or it is independent and a rival's homeworld. A world the acting faction governs
// never is, a seized rival homeworld included.
func isEnemyTerritory(actingFactionID, worldID string, factionState *state.FactionState) bool {
	if !factionState.IsIndependent(worldID) {
		return factionState.Ruler(worldID) != actingFactionID
	}
	for factionID, faction := range factionState.Factions {
		if factionID != actingFactionID && faction.Homeworld.WorldID == worldID {
			return true
		}
	}
	return false
}
```

In `helpers.go`, delete `rivalHasPlanetaryGovernmentOnWorld` and
`factionHasPlanetaryGovernmentTag`; `Faction.HasTag` (Task 2) replaces the second:

**Find:**
```go
func rivalHasPlanetaryGovernmentOnWorld(actingFactionID, locationID string, factionState *state.FactionState, index *world.Index) bool {
	for _, base := range index.BasesByLocation[locationID] {
		if base.OwnerID == actingFactionID {
			continue
		}
		rival, ok := factionState.Factions[base.OwnerID]
		if !ok {
			continue
		}
		if factionHasPlanetaryGovernmentTag(rival) {
			return true
		}
	}
	return false
}

func factionHasPlanetaryGovernmentTag(faction *domain.Faction) bool {
	for _, tag := range faction.Tags {
		if tag.ID == "T-011" {
			return true
		}
	}
	return false
}

```
**Replace with:** *(nothing)*

In `world_index_test.go`, delete `TestRivalHasPlanetaryGovernmentOnWorld_RivalHoldsBase`,
`TestRivalHasPlanetaryGovernmentOnWorld_RivalNoTag` and
`TestRivalHasPlanetaryGovernmentOnWorld_OwnBaseIgnored` (the helper is gone); drop
any import the file no longer uses.

##### Task 7 — `internal/faction/narrative/digest/types.go`, `build.go`

In `types.go`:

**Find:**
```go
	GoalHomeworldShift
	GoalTagGained
)
```
**Replace with:**
```go
	GoalHomeworldShift
	GoalTagGained
	GoalWorldGoverned
	GoalTagLost
)
```

In `build.go`:

**Find:**
```go
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:    GoalTagGained,
					TagName: m.Tag.Name,
				})
```
**Replace with:**
```go
				beat.GoalEvents = append(beat.GoalEvents, GoalEvent{
					Kind:    GoalTagGained,
					TagName: m.Tag.Name,
				})

			case "tag_removed":
				var m domain.TagRemoved
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				loser := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				loser.GoalEvents = append(loser.GoalEvents, GoalEvent{Kind: GoalTagLost, TagName: m.Tag.Name})

			case "world_government_changed":
				var m domain.WorldGovernmentChanged
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				if m.ToFactionID != "" {
					ruler := getOrCreateBeat(beats, m.ToFactionID, factionState)
					seenFactionIDs[m.ToFactionID] = struct{}{}
					ruler.GoalEvents = append(ruler.GoalEvents, GoalEvent{Kind: GoalWorldGoverned, ToWorld: m.WorldID})
				}
```

##### Task 8 — `internal/faction/narrative/wire_renderer.go`, `wire_templates.go`

In `wire_renderer.go`, `renderGoalEvent`:

**Find:**
```go
	case digest.GoalTagGained:
		tmpl := pick(rng, goalTagTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	}
```
**Replace with:**
```go
	case digest.GoalTagGained:
		tmpl := pick(rng, goalTagTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	case digest.GoalWorldGoverned:
		tmpl := pick(rng, goalGovernedTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{world}", ge.ToWorld)
	case digest.GoalTagLost:
		tmpl := pick(rng, goalTagLostTemplates)
		result := strings.ReplaceAll(tmpl, "{faction}", factionName)
		return strings.ReplaceAll(result, "{tag}", ge.TagName)
	}
```

In `wire_templates.go`:

**Find:**
```go
var goalTagTemplates = []string{
	"{faction} gained the tag: {tag}.",
	"New descriptor logged for {faction}: {tag}.",
}
```
**Replace with:**
```go
var goalTagTemplates = []string{
	"{faction} gained the tag: {tag}.",
	"New descriptor logged for {faction}: {tag}.",
}

var goalGovernedTemplates = []string{
	"{faction} now governs {world}.",
	"A change of rule: {world} answers to {faction}.",
}

var goalTagLostTemplates = []string{
	"{faction} lost the tag: {tag}.",
	"{faction} no longer holds {tag}.",
}
```

##### Task 9 — `internal/faction/tui/views/turn/execution/format.go`

Two edits.

**Find:**
```go
	case domain.AssetOwnerChanged:
		return fmt.Sprintf("asset %s captured by %s", mut.AssetID, mut.ToFactionID), true
```
**Replace with:**
```go
	case domain.AssetOwnerChanged:
		return fmt.Sprintf("asset %s captured by %s", mut.AssetID, mut.ToFactionID), true
	case domain.WorldGovernmentChanged:
		if mut.ToFactionID == "" {
			return fmt.Sprintf("%s → independent", mut.WorldID), true
		}
		return fmt.Sprintf("%s governed by %s", mut.WorldID, mut.ToFactionID), true
```

**Find:**
```go
		// XPSpent, XPAwarded, MovementOrderProgressed, InfluenceDelta, TagAdded.
```
**Replace with:**
```go
		// XPSpent, XPAwarded, MovementOrderProgressed, InfluenceDelta, TagAdded,
		// TagRemoved.
```

##### Task 10 — `rulebooks/swn/tags.toml`

The tag marks the holder; the record holds the worlds (overview Decision 10).

**Find:**
```toml
effect = "The faction's permission is required to buy or import assets marked as needing government permission. Can be held once per planet controlled."
```
**Replace with:**
```toml
effect = "Held while the faction governs at least one world; the campaign's government record lists each world it governs (one per planet controlled). On those worlds, the faction grants or refuses permission to buy or import assets marked as needing government permission."
```

##### Task 11 — `docs/rules/swn-faction-mechanics.md`

Two edits.

**Find:**
```markdown
- On success: gain the **Planetary Government** tag for that world
```
**Replace with:**
```markdown
- On success: become the world's ruler in the government record, and gain the **Planetary Government** tag if not already held; a previous ruler that governs no other world loses it
```

**Find:**
```markdown
The **Planetary Government** tag grants or denies this permission for that world.
```
**Replace with:**
```markdown
The campaign's government record names the ruler of each world; that ruler grants or denies this permission there, and the GM decides for a world no faction governs. The **Planetary Government** tag marks a faction that governs at least one world.
```

##### Task 12 — `docs/architecture/engine/goals.md`

Three edits.

**Find:**
```markdown
  emits the completion bundle — `GoalCompleted`, `XPAwarded`, and a
  `TagAdded` granting the `T-011` Planetary Government [tag](effect-mutation.md).
```
**Replace with:**
```markdown
  emits the completion bundle — `GoalCompleted`, `XPAwarded`, a
  `WorldGovernmentChanged` that makes the faction the world's ruler in the
  government record, a `TagAdded` for the `T-011` Planetary Government
  [tag](effect-mutation.md) when the faction does not hold it, and a `TagRemoved`
  for a previous ruler that governs no other world. The XP reads the previous
  ruler's ratings (half their average, at least 1); a world with no faction ruler
  gives 1.
```

**Find:**
```markdown
| Inside Enemy Territory (G-009) | Stealthing assets on rival-government worlds | Stealthed assets reach the faction's Cunning rating |
```
**Replace with:**
```markdown
| Inside Enemy Territory (G-009) | Stealthing assets on worlds a rival governs, or on a rival's independent homeworld | Stealthed assets reach the faction's Cunning rating |
```

**Find:**
```markdown
  completion, XP, ticks, phase advances, even the Planetary Government tag are all
```
**Replace with:**
```markdown
  completion, XP, ticks, phase advances, even a world's change of government are all
```

##### Task 13 — `docs/architecture/engine/persistence.md`

**Find:**
```markdown
`FactionState` is the entire mutable game: the `CampaignID`, the `CycleNumber`,
a `map[string]*Faction`, and the optional `CurrentTurn` cursor (the pause/resume
state from the [turn pipeline](turn-pipeline.md)). It is one TOML file.
```
**Replace with:**
```markdown
`FactionState` is the entire mutable game: the `CampaignID`, the `CycleNumber`,
a `map[string]*Faction`, the government record (`Governments`, world ID → ruling
faction ID; a world with no entry is independent, governed by no tracked faction,
so a file without the table loads as "every world independent"), and the optional
`CurrentTurn` cursor (the
pause/resume state from the [turn pipeline](turn-pipeline.md)). It is one TOML file.
```

##### Task 14 — `docs/architecture/engine/effect-mutation.md`

**Find:**
```markdown
- `tag_added` — grant a tag (the full `Tag` is embedded, so `Apply` needs no rulebook lookup)
```
**Replace with:**
```markdown
- `tag_added` / `tag_removed` — grant / take a tag (the full `Tag` is embedded, so `Apply` needs no rulebook lookup)
- `world_government_changed` — set or clear a world's ruler in the government record (Planetary Seizure)
```

##### Task 15 — engine tests

`internal/faction/engine/mutation/mutation_test.go` — append:
```go
func TestMutationEngine_WorldGovernmentChanged(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f2"] = &domain.Faction{ID: "f2"}
	me := New(slog.New(slog.DiscardHandler))
	apply := func(mutation domain.Mutation) *MutationApplyError {
		return me.Apply(s, []domain.Mutation{mutation}, slog.New(slog.DiscardHandler))
	}

	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", ToFactionID: "f1"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := s.Ruler("Anchorage"); got != "f1" {
		t.Errorf("Ruler after f1 takes it = %q, want f1", got)
	}
	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f1", ToFactionID: "f2"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := s.Ruler("Anchorage"); got != "f2" {
		t.Errorf("Ruler after f2 takes it = %q, want f2", got)
	}
	err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f2", ToFactionID: "missing"})
	if err == nil || len(err.Misses) != 1 {
		t.Fatalf("Apply to a missing faction = %v, want one miss", err)
	}
	if got := s.Ruler("Anchorage"); got != "f2" {
		t.Errorf("Ruler after a miss = %q, want f2 unchanged", got)
	}
	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f2"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, ok := s.Governments["Anchorage"]; ok {
		t.Error("an empty ToFactionID left an entry, want the world independent")
	}
}

func TestMutationEngine_TagRemoved(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f1"].Tags = []*domain.Tag{{ID: "T-001"}, {ID: domain.TagPlanetaryGovernment}}
	me := New(slog.New(slog.DiscardHandler))
	removal := domain.TagRemoved{FactionID: "f1", Tag: domain.Tag{ID: domain.TagPlanetaryGovernment}}

	if err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if tags := s.Factions["f1"].Tags; len(tags) != 1 || tags[0].ID != "T-001" {
		t.Errorf("Tags = %v, want [T-001]", tags)
	}
	err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].EntityID != domain.TagPlanetaryGovernment {
		t.Errorf("removing a tag not held = %v, want one miss naming %s", err, domain.TagPlanetaryGovernment)
	}
}
```

`internal/faction/engine/goal/goals/lock_test.go` — add `"reflect"` to the imports
and append:
```go
// makeOccupationState builds f1's last turn of occupation on Krylos, with rival f2
// at ratings 4/4/4 holding T-011.
func makeOccupationState() (*state.FactionState, *domain.Faction) {
	faction := &domain.Faction{
		ID: "f1",
		ActiveGoal: &domain.ActiveGoal{
			GoalID:         "G-004",
			TargetWorld:    domain.Location{WorldID: "Krylos"},
			ProcessPhase:   2,
			TurnsRemaining: 1,
		},
		Assets: map[string]*domain.Asset{"a1": {ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}},
	}
	rival := &domain.Faction{ID: "f2", Force: 4, Cunning: 4, Wealth: 4,
		Tags: []*domain.Tag{{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"}}}
	return &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}, faction
}

func TestPlanetarySeizure_CompletionTakesTheGovernment(t *testing.T) {
	tag := domain.Tag{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"}
	changed := func(from string) domain.WorldGovernmentChanged {
		return domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: from, ToFactionID: "f1", Cause: "goal_completed", CausedByFactionID: "f1"}
	}
	gained := domain.TagAdded{FactionID: "f1", Tag: tag, Cause: "goal_completed", CausedByFactionID: "f1"}
	completed := func(xp int) []domain.Mutation {
		return []domain.Mutation{
			domain.GoalCompleted{FactionID: "f1", GoalID: "G-004", XPAwarded: xp},
			domain.XPAwarded{FactionID: "f1", Amount: xp, Cause: "goal_completed"},
		}
	}
	cases := []struct {
		name        string
		governments map[string]string
		seizerTag   bool
		want        []domain.Mutation
	}{
		{
			name:        "a rival's last world",
			governments: map[string]string{"Krylos": "f2"},
			want: append([]domain.Mutation{changed("f2"), gained,
				domain.TagRemoved{FactionID: "f2", Tag: tag, Cause: "goal_completed", CausedByFactionID: "f1"}}, completed(2)...),
		},
		{
			name:        "the rival keeps another world",
			governments: map[string]string{"Krylos": "f2", "Hadrian": "f2"},
			want:        append([]domain.Mutation{changed("f2"), gained}, completed(2)...),
		},
		{
			name:        "an independent world, the seizer already a government",
			governments: map[string]string{"Tartarus": "f1"},
			seizerTag:   true,
			want:        append([]domain.Mutation{changed("")}, completed(1)...),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, faction := makeOccupationState()
			factionState.Governments = tc.governments
			if tc.seizerTag {
				faction.Tags = []*domain.Tag{&tag}
			}

			_, mutations := PlanetarySeizure{}.CheckLock(faction, factionState, makeLockRulebook())

			if len(mutations) == 0 {
				t.Fatal("no mutations, want a tick and the completion bundle")
			}
			if _, ok := mutations[0].(domain.GoalTurnsTick); !ok {
				t.Errorf("mutations[0] = %T, want GoalTurnsTick", mutations[0])
			}
			if !reflect.DeepEqual(mutations[1:], tc.want) {
				t.Errorf("completion = %+v, want %+v", mutations[1:], tc.want)
			}
		})
	}
}
```

`internal/faction/engine/goal/goals/inside_enemy_territory_test.go` (new) — full
contents:
```go
package goals

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestInsideEnemyTerritory_CountsRivalGround(t *testing.T) {
	assetWorlds := map[string]string{
		"governed":         "Tartarus",
		"rival-governed":   "Krylos",
		"independent":      "Hadrian",
		"rival-homeworld":  "Anchorage",
		"seized-homeworld": "Vesta",
		"lost-homeworld":   "Brennus",
		"in-transit":       "",
	}
	faction := &domain.Faction{
		ID:         "f1",
		Cunning:    5,
		Homeworld:  domain.Location{WorldID: "Brennus"},
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-009"},
		Assets:     map[string]*domain.Asset{},
	}
	var mutations []domain.Mutation
	for assetID, worldID := range assetWorlds {
		faction.Assets[assetID] = &domain.Asset{ID: assetID, Location: domain.Location{WorldID: worldID}}
		mutations = append(mutations, domain.AssetStealthApplied{FactionID: "f1", AssetID: assetID})
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": faction,
			"f2": {ID: "f2", Homeworld: domain.Location{WorldID: "Anchorage"}},
			"f3": {ID: "f3", Homeworld: domain.Location{WorldID: "Vesta"}},
		},
		Governments: map[string]string{"Tartarus": "f1", "Krylos": "f2", "Brennus": "f2", "Vesta": "f1"},
	}

	got := InsideEnemyTerritory{}.UpdateProgress(faction, mutations, factionState, nil, nil)

	want := []domain.Mutation{domain.GoalProgressed{FactionID: "f1", GoalID: "G-009", Delta: 3, Cause: "inside_enemy_territory"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v (Krylos, Anchorage and Brennus count)", got, want)
	}
}
```

##### Task 16 — state, digest and TUI tests

`internal/faction/state/faction_state_test.go` — append (add `os`, `path/filepath`,
`reflect` and `strings` to the imports where missing):
```go
func TestGovernmentsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	original := &FactionState{Factions: map[string]*domain.Faction{}, Governments: map[string]string{"Krylos": "f1"}}
	if err := Save(path, original); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(loaded.Governments, original.Governments) {
		t.Errorf("Governments = %v, want %v", loaded.Governments, original.Governments)
	}
}

func TestGovernmentsOmittedWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.toml")
	if err := Save(path, &FactionState{Factions: map[string]*domain.Faction{}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(body), "governments") {
		t.Errorf("state.toml = %q, want no governments table", body)
	}
}

func TestRuler(t *testing.T) {
	var nilState *FactionState
	if got := nilState.Ruler("Krylos"); got != "" {
		t.Errorf("nil state Ruler = %q, want none", got)
	}
	if got := (&FactionState{}).Ruler("Krylos"); got != "" {
		t.Errorf("no record Ruler = %q, want none (an independent world)", got)
	}
	if got := (&FactionState{Governments: map[string]string{"Krylos": "f1"}}).Ruler("Krylos"); got != "f1" {
		t.Errorf("Ruler = %q, want f1", got)
	}
}

func TestIsIndependent(t *testing.T) {
	var nilState *FactionState
	if !nilState.IsIndependent("Krylos") {
		t.Error("nil state IsIndependent = false, want true")
	}
	governed := &FactionState{Governments: map[string]string{"Krylos": "f1"}}
	if governed.IsIndependent("Krylos") {
		t.Error("a governed world IsIndependent = true, want false")
	}
	if !governed.IsIndependent("Hadrian") {
		t.Error("a world with no entry IsIndependent = false, want true")
	}
}
```

`internal/faction/narrative/digest/digest_test.go` — append:
```go
func TestGovernmentChangeAndTagLoss(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha", "f2": "Bravo"})
	records := []domain.EventRecord{
		eventRec("f1",
			mutRec(t, "world_government_changed", domain.WorldGovernmentChanged{
				WorldID: "Krylos", FromFactionID: "f2", ToFactionID: "f1", Cause: "goal_completed", CausedByFactionID: "f1",
			}),
			mutRec(t, "tag_removed", domain.TagRemoved{
				FactionID: "f2", Tag: domain.Tag{ID: domain.TagPlanetaryGovernment, Name: "Planetary Government"},
				Cause: "goal_completed", CausedByFactionID: "f1",
			}),
		),
	}
	cycleDigest := buildOne(t, records, factionState, nil)

	ruler := findBeat(cycleDigest, "f1")
	if ruler == nil || !slices.ContainsFunc(ruler.GoalEvents, func(event GoalEvent) bool {
		return event.Kind == GoalWorldGoverned && event.ToWorld == "Krylos"
	}) {
		t.Errorf("f1 beat = %+v, want a GoalWorldGoverned event for Krylos", ruler)
	}
	loser := findBeat(cycleDigest, "f2")
	if loser == nil || !slices.ContainsFunc(loser.GoalEvents, func(event GoalEvent) bool {
		return event.Kind == GoalTagLost && event.TagName == "Planetary Government"
	}) {
		t.Errorf("f2 beat = %+v, want a GoalTagLost event", loser)
	}
	for _, beat := range []*FactionBeat{ruler, loser} {
		if beat != nil && len(beat.Notes) != 0 {
			t.Errorf("%s Notes = %v, want none", beat.Faction.ID, beat.Notes)
		}
	}
}
```

`internal/faction/narrative/wire_renderer_test.go` — append (add `math/rand`,
`strings` and the `digest` import where missing):
```go
func TestRenderGoalEvent_GovernmentAndTagLoss(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	governed := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalWorldGoverned, ToWorld: "Krylos"}, "Alpha", rng)
	if !strings.Contains(governed, "Alpha") || !strings.Contains(governed, "Krylos") {
		t.Errorf("governed line = %q, want the faction and the world", governed)
	}
	lost := renderGoalEvent(digest.GoalEvent{Kind: digest.GoalTagLost, TagName: "Planetary Government"}, "Bravo", rng)
	if !strings.Contains(lost, "Bravo") || !strings.Contains(lost, "Planetary Government") {
		t.Errorf("tag-lost line = %q, want the faction and the tag", lost)
	}
}
```

`internal/faction/tui/views/turn/execution/format_test.go` — in
`TestFormatMutation_Surfaced`:

**Find:**
```go
		{"AssetMoved ability", domain.AssetMoved{AssetID: "sf-1", ToLocation: domain.Location{WorldID: "Andama"}, Cause: "ability"}, "asset sf-1 → Andama"},
	}
```
**Replace with:**
```go
		{"AssetMoved ability", domain.AssetMoved{AssetID: "sf-1", ToLocation: domain.Location{WorldID: "Andama"}, Cause: "ability"}, "asset sf-1 → Andama"},
		{"WorldGovernmentChanged", domain.WorldGovernmentChanged{WorldID: "Krylos", ToFactionID: "f1"}, "Krylos governed by f1"},
		{"WorldGovernmentChanged independent", domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: "f1"}, "Krylos → independent"},
	}
```

In `TestFormatMutation_Folded`:

**Find:**
```go
		domain.TagAdded{},
	}
```
**Replace with:**
```go
		domain.TagAdded{},
		domain.TagRemoved{},
	}
```

##### Commit message

```
feat: world government record

- FactionState.Governments (world → ruling faction; absent = independent),
  Ruler() and IsIndependent()
- WorldGovernmentChanged and TagRemoved, with apply, digest and TUI lines
- Planetary Seizure takes the government on completion; T-011 follows the
  record; XP reads the previous ruler's ratings
- Inside Enemy Territory counts a rival's ground: a rival-governed world, or a
  rival's independent homeworld
- T-011 effect text and the rules doc name the record
```

#### Commit 3 — `feat: the wizard records a world's government`

The GM's authoring surface for the record (Decision 7). Uses `FactionState.Ruler`,
defined in Commit 2 Task 1, and `domain.TagPlanetaryGovernment`, defined in Commit 2
Task 2.

##### Task 1 — `internal/faction/state/crud.go`

`CreateFaction` records the governed world in the same save, and `DeleteFaction`
clears the deleted faction's entries (F5). Three edits.

**Find:**
```go
// CreateFaction validates uniqueness, applies the mutation to fs.Factions,
// then persists to path. The mutation is atomic: if Save fails the in-memory
// insert is rolled back, so fs.Factions always matches what is on disk and the
// caller may safely retry.
func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error {
	if faction.ID == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[faction.ID]; exists {
		return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
	}
	fs.Factions[faction.ID] = faction
	if err := Save(path, fs); err != nil {
		delete(fs.Factions, faction.ID)
		return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
	}
	return nil
}
```
**Replace with:**
```go
// CreateFaction validates uniqueness, applies the mutation to fs.Factions,
// records governedWorldID as a world the faction rules (none when it is empty),
// then persists to path. The mutation is atomic: if Save fails the in-memory
// insert is rolled back, so fs.Factions always matches what is on disk and the
// caller may safely retry.
func CreateFaction(path string, fs *FactionState, faction *domain.Faction, governedWorldID string) error {
	if faction.ID == "" {
		return ErrInvalidFactionID
	}
	if _, exists := fs.Factions[faction.ID]; exists {
		return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
	}
	previousRuler, governed := fs.Governments[governedWorldID]
	fs.Factions[faction.ID] = faction
	if governedWorldID != "" {
		if fs.Governments == nil {
			fs.Governments = make(map[string]string)
		}
		fs.Governments[governedWorldID] = faction.ID
	}
	if err := Save(path, fs); err != nil {
		delete(fs.Factions, faction.ID)
		if governedWorldID != "" {
			restoreGovernment(fs, governedWorldID, previousRuler, governed)
		}
		return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
	}
	return nil
}

// restoreGovernment puts worldID's entry back as it was before a failed save.
func restoreGovernment(fs *FactionState, worldID, previousRuler string, governed bool) {
	if governed {
		fs.Governments[worldID] = previousRuler
		return
	}
	delete(fs.Governments, worldID)
}
```

**Find:**
```go
// DeleteFaction validates that id exists in fs.Factions, removes the entry,
// then persists.
```
**Replace with:**
```go
// DeleteFaction validates that id exists in fs.Factions, removes the entry and
// every world it governs from the record, then persists.
```

**Find:**
```go
	removed, exists := fs.Factions[id]
	if !exists {
		return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
	}
	delete(fs.Factions, id)
	if err := Save(path, fs); err != nil {
		fs.Factions[id] = removed
		return fmt.Errorf("state: delete %q: save: %w", id, err)
	}
	return nil
}
```
**Replace with:**
```go
	removed, exists := fs.Factions[id]
	if !exists {
		return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
	}
	ruled := worldsRuledBy(fs, id)
	delete(fs.Factions, id)
	for _, worldID := range ruled {
		delete(fs.Governments, worldID)
	}
	if err := Save(path, fs); err != nil {
		fs.Factions[id] = removed
		for _, worldID := range ruled {
			fs.Governments[worldID] = id
		}
		return fmt.Errorf("state: delete %q: save: %w", id, err)
	}
	return nil
}

func worldsRuledBy(fs *FactionState, factionID string) []string {
	var worlds []string
	for worldID, ruler := range fs.Governments {
		if ruler == factionID {
			worlds = append(worlds, worldID)
		}
	}
	return worlds
}
```

##### Task 2 — `internal/faction/tui/views/manage/msgs/msgs.go`, `manage.go`

In `msgs.go`:

**Find:**
```go
type CreatedMsg struct{ Faction *domain.Faction }
```
**Replace with:**
```go
// CreatedMsg carries a new faction and, for a Planetary Government faction, the
// world it governs (empty for none).
type CreatedMsg struct {
	Faction         *domain.Faction
	GovernedWorldID string
}
```

In `manage.go`:

**Find:**
```go
		if err := state.CreateFaction(m.paths.StatePath, m.factionState, msg.Faction); err != nil {
```
**Replace with:**
```go
		if err := state.CreateFaction(m.paths.StatePath, m.factionState, msg.Faction, msg.GovernedWorldID); err != nil {
```

##### Task 3 — `internal/faction/tui/views/manage/wizard/wizard.go`

The T-011 pick asks which world. Six edits. Add `"slices"` to the imports.

**Find:**
```go
	homeworldID     string
	primaryAssets   []*domain.AssetDefinition
```
**Replace with:**
```go
	homeworldID     string
	governedWorldID string
	primaryAssets   []*domain.AssetDefinition
```

**Find:**
```go
	m.form = buildForm(data, rb, spatialMap)
```
**Replace with:**
```go
	m.form = buildForm(data, rb, spatialMap, factionState)
```

**Find:**
```go
	if m.assetsForm.State == huh.StateCompleted {
		faction := m.synthesize()
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction} }
	}
```
**Replace with:**
```go
	if m.assetsForm.State == huh.StateCompleted {
		faction := m.synthesize()
		governedWorldID := ""
		if pickedGovernmentTag(m.data.selectedTags) {
			governedWorldID = m.data.governedWorldID
		}
		return m, func() tea.Msg { return msgs.CreatedMsg{Faction: faction, GovernedWorldID: governedWorldID} }
	}
```

**Find:**
```go
func buildForm(data *formData, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
```
**Replace with:**
```go
func buildForm(data *formData, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, factionState *state.FactionState) *huh.Form {
```

**Find:**
```go
	coinGroup := huh.NewGroup(
```
**Replace with:**
```go
	governmentGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Governed world (Planetary Government)").
			OptionsFunc(func() []huh.Option[string] {
				return governableWorldOptions(spatialMap, factionState, data.homeworldID)
			}, &data.homeworldID).
			Value(&data.governedWorldID),
	).WithHideFunc(func() bool { return !pickedGovernmentTag(data.selectedTags) })

	coinGroup := huh.NewGroup(
```

**Find:**
```go
		goalGroup,
		homeworldGroup,
		coinGroup,
	).WithTheme(styles.FormTheme())
}
```
**Replace with:**
```go
		goalGroup,
		homeworldGroup,
		governmentGroup,
		coinGroup,
	).WithTheme(styles.FormTheme())
}

func pickedGovernmentTag(tags []*domain.Tag) bool {
	return slices.ContainsFunc(tags, func(tag *domain.Tag) bool { return tag.ID == domain.TagPlanetaryGovernment })
}

// governableWorldOptions lists the worlds no other faction governs, the homeworld
// first: a government most often rules its own home.
func governableWorldOptions(spatialMap *spatial.RegionMap, factionState *state.FactionState, homeworldID string) []huh.Option[string] {
	var home, others []huh.Option[string]
	for _, option := range worldsToOptions(spatialMap) {
		if factionState.Ruler(option.Value) != "" {
			continue
		}
		if option.Value == homeworldID {
			home = append(home, option)
			continue
		}
		others = append(others, option)
	}
	return append(home, others...)
}
```

##### Task 4 — `docs/rules/swn-faction-mechanics.md`

**Find:**
```markdown
   - If the faction controls one or more worlds, add **Planetary Government** tags
```
**Replace with:**
```markdown
   - If the faction controls a world, add the **Planetary Government** tag; the faction wizard then asks which world it governs
```

##### Task 5 — `internal/faction/state/crud_test.go`

The three existing `CreateFaction(...)` calls gain a trailing `""`. Append:
```go
func TestCreateFactionRecordsGovernment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faction_state.toml")
	fs := newTestState()

	if err := CreateFaction(path, fs, minimalFaction("rulers"), "Krylos"); err != nil {
		t.Fatalf("CreateFaction() error: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := loaded.Ruler("Krylos"); got != "rulers" {
		t.Errorf("Ruler(Krylos) after load = %q, want rulers", got)
	}
}

func TestCreateFactionSaveFailureRollsBackGovernment(t *testing.T) {
	fs := newTestState()

	if err := CreateFaction(t.TempDir(), fs, minimalFaction("doomed"), "Krylos"); err == nil {
		t.Fatal("CreateFaction() expected save error, got nil")
	}
	if got := fs.Ruler("Krylos"); got != "" {
		t.Errorf("Ruler(Krylos) after failed create = %q, want none", got)
	}
}

func TestDeleteFactionClearsItsGovernments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "faction_state.toml")
	fs := newTestState()
	fs.Factions["rulers"] = minimalFaction("rulers")
	fs.Governments = map[string]string{"Krylos": "rulers", "Hadrian": "others"}

	if err := DeleteFaction(path, fs, "rulers"); err != nil {
		t.Fatalf("DeleteFaction() error: %v", err)
	}
	if got := fs.Ruler("Krylos"); got != "" {
		t.Errorf("Ruler(Krylos) = %q, want none", got)
	}
	if got := fs.Ruler("Hadrian"); got != "others" {
		t.Errorf("Ruler(Hadrian) = %q, want others untouched", got)
	}
}

func TestDeleteFactionSaveFailureRestoresGovernments(t *testing.T) {
	fs := newTestState()
	fs.Factions["rulers"] = minimalFaction("rulers")
	fs.Governments = map[string]string{"Krylos": "rulers"}

	if err := DeleteFaction(t.TempDir(), fs, "rulers"); err == nil {
		t.Fatal("DeleteFaction() expected save error, got nil")
	}
	if got := fs.Ruler("Krylos"); got != "rulers" {
		t.Errorf("Ruler(Krylos) after failed delete = %q, want rulers", got)
	}
}
```

##### Commit message

```
feat: the wizard records a world's government

- a T-011 pick asks which world the faction governs (worlds no other faction
  governs, the homeworld first)
- CreateFaction writes the entry in the same rollback-safe save; DeleteFaction
  clears a deleted faction's entries
```

### Phase 3 — The permission request

**In plain words:** When a faction buys an asset that needs permission, the game asks
the faction that governs the world, or asks the GM if no faction governs it. A
refusal ends the purchase with nothing bought, and the turn report shows the refusal.

#### Commit 4 — `feat: permission request at purchase`

##### Task 1 — `internal/faction/engine/hooks/collector.go`

The request (Decision 8). Append:
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
type PermissionCollector interface {
	RequestPermission(request PermissionRequest) (bool, error)
}
```

##### Task 2 — `internal/faction/engine/action/collector.go`

**Find:**
```go
	hooks.Collector
	hooks.InterceptCollector
```
**Replace with:**
```go
	hooks.Collector
	hooks.InterceptCollector
	hooks.PermissionCollector
```

Then run `go generate ./internal/faction/engine/action/...`.

##### Task 3 — `internal/faction/domain/asset.go`

The permission test moves from the menu to the request (Decision 12). Two edits.

**Find:**
```go
// HalfPrice is half the listed cost, rounded down, but at least 1 Coin: the price
```
**Replace with:**
```go
// NeedsPurchasePermission reports whether buying def asks the world's government.
func (def *AssetDefinition) NeedsPurchasePermission() bool {
	return def.HasFlag(FlagPermission)
}

// HalfPrice is half the listed cost, rounded down, but at least 1 Coin: the price
```

**Find:**
```go
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

// PurchaseTerms are the per-path inputs to CanPurchase that differ between the Buy Asset
// action and the purchase_enabler ability: Cost is the price to test affordability against
// (base cost for a buy, surcharged for an enabler); TechCap is the highest required tech
// level a def may carry (the target world's tech level for a buy, the enabler's own cap for
// an enabler); HoldsGovernment is whether the faction owns a base on the target world (the
// P-flag gate). The shared checks themselves live in CanPurchase so the two paths can't drift.
type PurchaseTerms struct {
	Cost            int
	TechCap         int
	HoldsGovernment bool
}

// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, the def's required tech level is within TechCap, and it
// holds Planetary Government when the def is P-flagged.
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
	case def.MinRating > faction.StatScore(def.Category):
		return false
	case def.TechLevel > terms.TechCap:
		return false
	case def.HasFlag(FlagPermission) && !terms.HoldsGovernment:
		return false
	}
	return true
}
```
**Replace with:**
```go
// PurchaseTerms are the per-path inputs to CanPurchase that differ between the Buy Asset
// action and the purchase_enabler ability: Cost is the price to test affordability against
// (base cost for a buy, surcharged for an enabler); TechCap is the highest required tech
// level a def may carry (the target world's tech level for a buy, the enabler's own cap for
// an enabler). The shared checks themselves live in CanPurchase so the two paths can't drift.
type PurchaseTerms struct {
	Cost    int
	TechCap int
}

// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, and the def's required tech level is within TechCap.
// Government permission is asked after the choice, not here (D-35).
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
	case def.MinRating > faction.StatScore(def.Category):
		return false
	case def.TechLevel > terms.TechCap:
		return false
	}
	return true
}
```

In `domain/asset_test.go`, delete `TestHasPlanetaryGovernment`.

##### Task 4 — `internal/faction/domain/mutation.go`

Uses `WorldGovernmentChanged`, defined in Commit 2 Task 3. The refusal record
(Decision 10).

**Find:**
```go
func (m WorldGovernmentChanged) Type() string { return "world_government_changed" }
```
**Replace with:**
```go
func (m WorldGovernmentChanged) Type() string { return "world_government_changed" }

// PermissionRefused records a refused permission request. It changes no state:
// the action it ends bought nothing (D-35, overview Decision 22).
// DecidedByFactionID is the world's ruler, empty when the GM decided for an
// independent world. Cause is the purchase path: buy, refit or ability.
type PermissionRefused struct {
	FactionID          string `json:"faction_id"`
	WorldID            string `json:"world_id"`
	DefinitionID       string `json:"definition_id"`
	DecidedByFactionID string `json:"decided_by_faction_id"`
	Cause              string `json:"cause"`
	CausedByFactionID  string `json:"caused_by_faction_id"`
}

func (m PermissionRefused) Type() string { return "permission_refused" }
```

##### Task 5 — `internal/faction/engine/mutation/mutation.go`

**Find:**
```go
		case domain.InfluenceDelta:
```
**Replace with:**
```go
		case domain.PermissionRefused:
			// record only: the refused action bought nothing
		case domain.InfluenceDelta:
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/permission.go` (new)

Full contents:
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
// requester's homeworld (Effort 4 Decision 15). A ruler's own request is granted at
// once (D-6). Otherwise the ruler's driver is asked, or the GM for an independent
// world (D-4). cause names the purchase path on the record.
func RequestPermission(collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	rulerID := factionState.Ruler(worldID)
	if rulerID == requester.ID {
		return nil, nil
	}
	request := hooks.PermissionRequest{Requester: requester, WorldID: worldID, Def: def}
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

##### Task 7 — `internal/faction/engine/action/actions/buy_asset.go`

The request after the choice (Decision 9); the menu offers `P` assets everywhere
(Decision 12). Edits against the file as Commit 1 Task 8 wrote it. Add a field:

**Find:**
```go
	qualityMutations []domain.Mutation
	quote            hooks.AssetCostQuery
```
**Replace with:**
```go
	qualityMutations []domain.Mutation
	refusal          *domain.PermissionRefused
	quote            hooks.AssetCostQuery
```

**Find:**
```go
func (ba *BuyAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
```
**Replace with:**
```go
func (ba *BuyAsset) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
```

**Find:**
```go
		ba.qualityTarget = target
	}
	return nil
}
```
**Replace with:**
```go
		ba.qualityTarget = target
	}
	refusal, err := dispatch.RequestPermission(ba.collector, factionState, faction, order.Definition, order.World, "buy")
	if err != nil {
		return fmt.Errorf("buy asset: %w", err)
	}
	ba.refusal = refusal
	return nil
}
```

**Find:**
```go
func (ba *BuyAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	def := ba.buyOrder.Definition
```
**Replace with:**
```go
func (ba *BuyAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	if ba.refusal != nil {
		return nil
	}
	def := ba.buyOrder.Definition
```

**Find:**
```go
func (ba *BuyAsset) Output() ([]domain.Mutation, error) {
	mutations := dispatch.PurchaseCharges(ba.factionID, ba.cost, ba.quote, "buy")
```
**Replace with:**
```go
func (ba *BuyAsset) Output() ([]domain.Mutation, error) {
	if ba.refusal != nil {
		return []domain.Mutation{*ba.refusal}, nil
	}
	mutations := dispatch.PurchaseCharges(ba.factionID, ba.cost, ba.quote, "buy")
```

**Find:**
```go
// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum attribute rating, within
// the world's tech level, and holding Planetary Government for P-flagged defs. A
// quality also needs a target on the world (D-22). An unknown
// world imposes no tech ceiling (math.MaxInt) but cannot confer government, so P-flagged
// defs are rejected — matching the purchase_enabler path.
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	holdsGovernment := false
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
		if worldEngine.Index != nil {
			holdsGovernment = domain.HasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID])
		}
	}
```
**Replace with:**
```go
// purchasableDefinitions returns the definitions the faction may buy on fragmentID via
// domain.CanPurchase: affordable at the cost Resolve will charge, meeting the minimum
// attribute rating, and within the world's tech level. A quality also needs a target on
// the world (D-22). A P-flagged def is offered everywhere; Inputs asks its permission
// after the choice (D-35). An unknown world imposes no tech ceiling (math.MaxInt).
func purchasableDefinitions(faction *domain.Faction, rulebook *rulebook.Rulebook, fragmentID string, worldEngine *world.WorldEngine, registry *hooks.Registry, effectsEngine *effect.EffectsEngine) []*domain.AssetDefinition {
	worldTechLevel := math.MaxInt
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
	}
```

**Find:**
```go
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap, HoldsGovernment: holdsGovernment}) {
```
**Replace with:**
```go
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: cost, TechCap: techCap}) {
```

##### Task 8 — `internal/faction/engine/action/actions/refit_asset.go`

The same request (D-23), and no refit in transit (Decision 11). Six edits.

**Find:**
```go
// Tech-level filtering and P-flag (government permission) checks are deferred.
type RefitAsset struct {
	collector  action.Collector
	factionID  string
	refitOrder action.RefitOrder
	newAsset   domain.Asset
	costDelta  int
}
```
**Replace with:**
```go
// A replacement that needs permission asks the world's government after the
// choice (D-23). Tech-level filtering is deferred.
type RefitAsset struct {
	collector  action.Collector
	factionID  string
	refitOrder action.RefitOrder
	refusal    *domain.PermissionRefused
	newAsset   domain.Asset
	costDelta  int
}
```

**Find:**
```go
func (ra *RefitAsset) Inputs(faction *domain.Faction, _ *state.FactionState, rulebook *rulebook.Rulebook) error {
```
**Replace with:**
```go
func (ra *RefitAsset) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
```

**Find:**
```go
	ra.factionID = faction.ID
	ra.refitOrder = order
	return nil
}
```
**Replace with:**
```go
	ra.factionID = faction.ID
	ra.refitOrder = order
	refusal, err := dispatch.RequestPermission(ra.collector, factionState, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")
	if err != nil {
		return fmt.Errorf("refit asset: %w", err)
	}
	ra.refusal = refusal
	return nil
}
```

**Find:**
```go
func (ra *RefitAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	oldDef, ok := rulebook.Assets[ra.refitOrder.OldAsset.DefinitionID]
```
**Replace with:**
```go
func (ra *RefitAsset) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	if ra.refusal != nil {
		return nil
	}
	oldDef, ok := rulebook.Assets[ra.refitOrder.OldAsset.DefinitionID]
```

**Find:**
```go
func (ra *RefitAsset) Output() ([]domain.Mutation, error) {
	mutations := []domain.Mutation{
```
**Replace with:**
```go
func (ra *RefitAsset) Output() ([]domain.Mutation, error) {
	if ra.refusal != nil {
		return []domain.Mutation{*ra.refusal}, nil
	}
	mutations := []domain.Mutation{
```

**Find:**
```go
	if oldAsset.HasCondition(domain.ConditionSheltered) {
		return nil
	}
```
**Replace with:**
```go
	if oldAsset.HasCondition(domain.ConditionSheltered) {
		return nil
	}
	// deliberate: no refit for an asset in transit, because an asset between worlds cannot be acted on (Effort 4 Decision 11; the general rule is F.028).
	if oldAsset.Location.WorldID == "" {
		return nil
	}
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"`
to the imports.

##### Task 9 — `internal/faction/engine/action/actions/ability/purchase_enabler.go`

The enabler asks after its pick; a refusal ends this asset's step with the record
(Decision 9). Three edits.

**Find:**
```go
	if def == nil {
		return nil, fmt.Errorf("purchase_enabler: %w", action.ErrNoSelection)
	}
```
**Replace with:**
```go
	if def == nil {
		return nil, fmt.Errorf("purchase_enabler: %w", action.ErrNoSelection)
	}
	refusal, err := dispatch.RequestPermission(ctx.Collector, ctx.FactionState, ctx.Faction, def, worldID, "ability")
	if err != nil {
		return nil, fmt.Errorf("purchase_enabler: %w", err)
	}
	if refusal != nil {
		return []domain.Mutation{*refusal}, nil
	}
```

**Find:**
```go
// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category, required tech level at or below the cap, the faction meeting the
// min rating, holding Planetary Government when the def is P-flagged (Decision 7), and
// affordable at the surcharged price. Name-sorted for a deterministic candidate order.
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	worldID := ctx.Asset.Location.WorldID
	government := holdsGovernment(ctx, worldID)
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category {
			continue
		}
		terms := domain.PurchaseTerms{
			Cost:            enablerQuote(ctx, def, params).Cost,
			TechCap:         params.MaxTechLevel,
			HoldsGovernment: government,
		}
```
**Replace with:**
```go
// eligiblePurchaseDefs returns the rulebook definitions an enabler purchase may buy:
// matching category, required tech level at or below the cap, the faction meeting the
// min rating, and affordable at the surcharged price. A P-flagged def is offered; its
// permission is asked after the pick (D-35). Name-sorted for a deterministic candidate
// order.
func eligiblePurchaseDefs(ctx StepContext, params *domain.PurchaseParams) []*domain.AssetDefinition {
	var eligible []*domain.AssetDefinition
	for _, def := range ctx.Rulebook.Assets {
		if def.Category != params.Category {
			continue
		}
		terms := domain.PurchaseTerms{
			Cost:    enablerQuote(ctx, def, params).Cost,
			TechCap: params.MaxTechLevel,
		}
```

**Find:**
```go
// holdsGovernment guards the (possibly nil) world engine before delegating to
// domain.HasPlanetaryGovernment. Pretech Logistics waives the tech-level requirement
// but not the P-flag permission requirement.
func holdsGovernment(ctx StepContext, worldID string) bool {
	if ctx.WorldEngine == nil || ctx.WorldEngine.Index == nil {
		return false
	}
	return domain.HasPlanetaryGovernment(ctx.Faction.ID, ctx.WorldEngine.Index.BasesByLocation[worldID])
}

```
**Replace with:** *(nothing)*

##### Task 10 — `internal/faction/tui/adapter/channels.go`, `action_collector.go`, `smoke.go`

In `channels.go`:

**Find:**
```go
	AskConfirmSacrifice
	// AskKindCount counts the kinds above; keep it last.
```
**Replace with:**
```go
	AskConfirmSacrifice
	AskRequestPermission
	// AskKindCount counts the kinds above; keep it last.
```

**Find:**
```go
type ConfirmSacrificePayload struct{ Offer hooks.SacrificeOffer }
```
**Replace with:**
```go
type ConfirmSacrificePayload struct{ Offer hooks.SacrificeOffer }
type RequestPermissionPayload struct {
	Request   hooks.PermissionRequest
	WorldName string
}
```

In `action_collector.go`, after `ConfirmSacrifice`:

**Find:**
```go
// confirmPayment asks the paying faction. The action collector (Marketers) and the
```
**Replace with:**
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

// confirmPayment asks the paying faction. The action collector (Marketers) and the
```

In `smoke.go`:

**Find:**
```go
func (SmokeActionCollector) ConfirmSacrifice(_ hooks.SacrificeOffer) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (SmokeActionCollector) ConfirmSacrifice(_ hooks.SacrificeOffer) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (SmokeActionCollector) RequestPermission(_ hooks.PermissionRequest) (bool, error) {
	return false, action.ErrActionUnavailable
}
```

##### Task 11 — `internal/faction/tui/views/turn/overlay/permission.go` (new)

Full contents:
```go
package overlay

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewRequestPermission asks a world's government whether a faction may raise an
// asset there. Its header names who decides: the ruler, or the GM for a world no
// faction governs. Esc does nothing: a cancel would drop the action, and with it
// every earlier asset of a Use Asset Ability. Answer: bool.
func NewRequestPermission(request hooks.PermissionRequest, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of the GM (no faction governs %s)", worldName)
	if request.Ruler != nil {
		header = fmt.Sprintf("Asked of %s, the government of %s", request.Ruler.Name, worldName)
	}
	title := fmt.Sprintf("Permit %s to raise %s on %s?", request.Requester.Name, request.Def.Name, worldName)
	return NoCancelConfirm{NewConfirm(title, header, "Grant", "Refuse")}
}
```

##### Task 12 — `internal/faction/tui/views/turn/execution/execution.go`

**Find:**
```go
	adapter.AskConfirmSacrifice: {"Sacrifice?", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmSacrifice(msg.Payload.(adapter.ConfirmSacrificePayload).Offer, m.rulebook)
	}},
```
**Replace with:**
```go
	adapter.AskConfirmSacrifice: {"Sacrifice?", func(m Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		return overlay.NewConfirmSacrifice(msg.Payload.(adapter.ConfirmSacrificePayload).Offer, m.rulebook)
	}},
	adapter.AskRequestPermission: {"Permission?", func(_ Model, msg adapter.CollectorAskMsg) overlay.Overlay {
		p := msg.Payload.(adapter.RequestPermissionPayload)
		return overlay.NewRequestPermission(p.Request, p.WorldName)
	}},
```

##### Task 13 — `internal/faction/engine/testharness/collector.go`

Add the field after `ConfirmSacrificeFn`:

**Find:**
```go
	ConfirmSacrificeFn            func(hooks.SacrificeOffer) (bool, error)
```
**Replace with:**
```go
	ConfirmSacrificeFn            func(hooks.SacrificeOffer) (bool, error)
	RequestPermissionFn           func(hooks.PermissionRequest) (bool, error)
```

**Find:**
```go
func (c *ScriptedCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	if c.ConfirmSacrificeFn != nil {
		return c.ConfirmSacrificeFn(offer)
	}
	return false, nil
}
```
**Replace with:**
```go
func (c *ScriptedCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	if c.ConfirmSacrificeFn != nil {
		return c.ConfirmSacrificeFn(offer)
	}
	return false, nil
}

// RequestPermission grants by default, so a scripted buy of a P asset goes through
// unless a test scripts a refusal.
func (c *ScriptedCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	if c.RequestPermissionFn != nil {
		return c.RequestPermissionFn(request)
	}
	return true, nil
}
```

##### Task 14 — `internal/faction/narrative/digest/types.go`, `resolve.go`, `headline.go`

The refusal line (Decision 13). In `types.go`:

**Find:**
```go
	GoalEvents   []GoalEvent
	StealthOps   []StealthEvent
	Notes        []string
}
```
**Replace with:**
```go
	GoalEvents   []GoalEvent
	StealthOps   []StealthEvent
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

In `resolve.go`, append:
```go
// resolveDefinitionName names a definition, or returns its ID when the rulebook
// lacks it.
func resolveDefinitionName(definitionID string, rulebook *rulebook.Rulebook) string {
	if rulebook != nil {
		if def, ok := rulebook.Assets[definitionID]; ok {
			return def.Name
		}
	}
	return definitionID
}
```

In `headline.go`, `isQuiet`:

**Find:**
```go
		len(beat.GoalEvents) > 0 || len(beat.StealthOps) > 0 {
```
**Replace with:**
```go
		len(beat.GoalEvents) > 0 || len(beat.StealthOps) > 0 || len(beat.Refusals) > 0 {
```

##### Task 15 — `internal/faction/narrative/digest/build.go`

**Find:**
```go
			case "goal_initiated":
				// dropped — bookkeeping only
```
**Replace with:**
```go
			case "permission_refused":
				var m domain.PermissionRefused
				if err := json.Unmarshal(mr.Payload, &m); err != nil {
					return CycleDigest{}, err
				}
				refused := getOrCreateBeat(beats, m.FactionID, factionState)
				seenFactionIDs[m.FactionID] = struct{}{}
				decidedBy := ""
				if m.DecidedByFactionID != "" {
					decidedBy = resolveFactionName(m.DecidedByFactionID, factionState)
				}
				refused.Refusals = append(refused.Refusals, RefusalEvent{
					AssetName: resolveDefinitionName(m.DefinitionID, rulebook),
					World:     m.WorldID,
					DecidedBy: decidedBy,
				})

			case "goal_initiated":
				// dropped — bookkeeping only
```

##### Task 16 — `internal/faction/narrative/wire_renderer.go`, `wire_templates.go`

In `wire_renderer.go`, `renderFactionSection`:

**Find:**
```go
	for _, stealth := range beat.StealthOps {
		sb.WriteString(renderStealthOp(stealth, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
```
**Replace with:**
```go
	for _, stealth := range beat.StealthOps {
		sb.WriteString(renderStealthOp(stealth, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
	for _, refusal := range beat.Refusals {
		sb.WriteString(renderRefusal(refusal, beat.Faction.Name, rng))
		sb.WriteString("\n\n")
	}
```

Append to `wire_renderer.go`:
```go
func renderRefusal(refusal digest.RefusalEvent, factionName string, rng *rand.Rand) string {
	decider := refusal.DecidedBy
	if decider == "" {
		decider = "The independent government"
	}
	tmpl := pick(rng, refusalTemplates)
	result := strings.ReplaceAll(tmpl, "{faction}", factionName)
	result = strings.ReplaceAll(result, "{asset}", refusal.AssetName)
	result = strings.ReplaceAll(result, "{world}", refusal.World)
	return strings.ReplaceAll(result, "{decider}", decider)
}
```

In `wire_templates.go`, after `goalTagLostTemplates` (Commit 2 Task 8):

**Find:**
```go
var goalTagLostTemplates = []string{
	"{faction} lost the tag: {tag}.",
	"{faction} no longer holds {tag}.",
}
```
**Replace with:**
```go
var goalTagLostTemplates = []string{
	"{faction} lost the tag: {tag}.",
	"{faction} no longer holds {tag}.",
}

var refusalTemplates = []string{
	"{decider} refused {faction} permission to raise {asset} on {world}.",
	"Permission denied: {decider} would not let {faction} raise {asset} on {world}.",
}
```

##### Task 17 — `internal/faction/tui/views/turn/execution/format.go`

After the `WorldGovernmentChanged` case (Commit 2 Task 9):

**Find:**
```go
		return fmt.Sprintf("%s governed by %s", mut.WorldID, mut.ToFactionID), true
```
**Replace with:**
```go
		return fmt.Sprintf("%s governed by %s", mut.WorldID, mut.ToFactionID), true
	case domain.PermissionRefused:
		return fmt.Sprintf("permission refused: %s on %s", mut.DefinitionID, mut.WorldID), true
```

##### Task 18 — `internal/faction/engine/hooks/dispatch/permission_test.go` (new)

Full contents:
```go
package dispatch

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type stubPermissionCollector struct {
	grant bool
	asked []hooks.PermissionRequest
}

func (stub *stubPermissionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	stub.asked = append(stub.asked, request)
	return stub.grant, nil
}

func permissionState() (*state.FactionState, *domain.Faction, *domain.Faction) {
	requester := &domain.Faction{ID: "f1", Name: "Alpha", Homeworld: domain.Location{WorldID: "Anchorage"}}
	ruler := &domain.Faction{ID: "f2", Name: "Bravo"}
	factionState := &state.FactionState{
		Factions:    map[string]*domain.Faction{"f1": requester, "f2": ruler},
		Governments: map[string]string{"Tartarus": "f1", "Krylos": "f2", "Anchorage": "f2"},
	}
	return factionState, requester, ruler
}

func TestRequestPermission(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	plainDef := &domain.AssetDefinition{ID: "militia"}
	factionState, requester, ruler := permissionState()

	cases := []struct {
		name        string
		def         *domain.AssetDefinition
		world       string
		grant       bool
		wantAsked   []hooks.PermissionRequest
		wantRefusal *domain.PermissionRefused
	}{
		{name: "no permission needed", def: plainDef, world: "Krylos"},
		{name: "the requester rules the world", def: permissionDef, world: "Tartarus"},
		{name: "the requester's homeworld, a rival its ruler", def: permissionDef, world: "Anchorage"},
		{
			name:      "a rival ruler grants",
			def:       permissionDef,
			world:     "Krylos",
			grant:     true,
			wantAsked: []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
		},
		{
			name:        "a rival ruler refuses",
			def:         permissionDef,
			world:       "Krylos",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1"},
		},
		{
			name:        "the GM refuses for an independent world",
			def:         permissionDef,
			world:       "Hadrian",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, WorldID: "Hadrian", Def: permissionDef}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Hadrian", DefinitionID: "gunships", Cause: "buy", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			collector := &stubPermissionCollector{grant: tc.grant}

			refusal, err := RequestPermission(collector, factionState, requester, tc.def, tc.world, "buy")

			if err != nil {
				t.Fatalf("RequestPermission: %v", err)
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

##### Task 19 — action tests

`internal/faction/engine/action/actions/buy_asset_test.go` — append:
```go
func TestBuyAsset_AsksPermissionAfterTheChoice(t *testing.T) {
	def := &domain.AssetDefinition{ID: "perm", Name: "Gunships", Category: domain.StatForce, HP: 4, Cost: 4, MinRating: 1,
		Flags: []domain.AssetFlag{domain.FlagPermission}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"perm": def}}

	cases := []struct {
		name    string
		world   string
		ask     bool
		grant   bool
		refused bool
	}{
		{name: "refused", world: "Krylos", ask: true, refused: true},
		{name: "granted", world: "Krylos", ask: true, grant: true},
		{name: "own world", world: "Anchorage"},
		{name: "homeworld a rival governs", world: "Tartarus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}
			ruler := &domain.Faction{ID: "f2", Name: "Bravo"}
			factionState := &state.FactionState{
				Factions:    map[string]*domain.Faction{"f1": faction, "f2": ruler},
				Governments: map[string]string{"Tartarus": "f2", "Krylos": "f2", "Anchorage": "f1"},
			}
			if !NewBuyAsset(nil, nil, nil, nil).Validate(faction, nil, rb) {
				t.Fatal("Validate = false, want a P asset offered on a world the faction does not rule (D-35)")
			}

			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: tc.world, Definition: def}, nil)
			if tc.ask {
				collector.EXPECT().
					RequestPermission(hooks.PermissionRequest{Requester: faction, Ruler: ruler, WorldID: tc.world, Def: def}).
					Return(tc.grant, nil)
			}

			act := NewBuyAsset(collector, nil, nil, nil)
			if err := act.Inputs(faction, factionState, rb); err != nil {
				t.Fatalf("Inputs: %v", err)
			}
			if err := act.Resolve(faction, factionState, rb); err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			mutations, err := act.Output()
			if err != nil {
				t.Fatalf("Output: %v", err)
			}

			if tc.refused {
				want := []domain.Mutation{domain.PermissionRefused{
					FactionID: "f1", WorldID: tc.world, DefinitionID: "perm", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1",
				}}
				if !reflect.DeepEqual(mutations, want) {
					t.Errorf("mutations = %+v, want only the refusal record (no asset, no charge)", mutations)
				}
				return
			}
			if len(mutations) != 2 {
				t.Fatalf("mutations = %+v, want AssetAdded and CoinDelta", mutations)
			}
			if _, ok := mutations[0].(domain.AssetAdded); !ok {
				t.Errorf("mutations[0] = %T, want AssetAdded", mutations[0])
			}
		})
	}
}
```

`internal/faction/engine/action/actions/refit_asset_test.go` — in
`TestRefitAsset_Validate`:

**Find:**
```go
	t.Run("a quality is never a replacement", func(t *testing.T) {
```
**Replace with:**
```go
	t.Run("an asset in transit cannot be refit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{RegionHex: spatial.RegionHex{RegionID: "drift"}}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl)).Validate(faction, nil, rulebook) {
			t.Error("expected false when the only asset has no world")
		}
	})

	t.Run("a quality is never a replacement", func(t *testing.T) {
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/spatial"` to its imports, and
append:
```go
func TestRefitAsset_PermissionRefused(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRefitRulebook()
	newDef := rulebook.Assets["new-def"]
	newDef.Flags = []domain.AssetFlag{domain.FlagPermission}
	oldAsset := &domain.Asset{ID: "f1-old-def-1", DefinitionID: "old-def", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"f1-old-def-1": oldAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRefitOrder(gomock.Any(), gomock.Any()).Return(action.RefitOrder{OldAsset: oldAsset, NewDefinition: newDef}, nil)
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	act := NewRefitAsset(collector)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	want := []domain.Mutation{domain.PermissionRefused{FactionID: "f1", WorldID: "Tartarus", DefinitionID: "new-def", Cause: "refit", CausedByFactionID: "f1"}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want only the refusal record", mutations)
	}
}
```
Add `"reflect"` to its imports.

`internal/faction/engine/action/actions/ability/purchase_enabler_test.go` — in
`TestPurchaseEnabler_BuysSurcharged`, the P-flagged def is now offered:

**Find:**
```go
	// candidate filter: wrong category, too-high TL, too-high rating, and the P-flagged
	// def (no WorldEngine, so no Planetary Government) all excluded.
	if len(offered) != 2 {
		t.Fatalf("offered %d candidates, want 2 (F-cheap, F-pricey); got %v", len(offered), offered)
	}
```
**Replace with:**
```go
	// candidate filter: wrong category, too-high TL and too-high rating excluded; the
	// P-flagged def is offered and asks after the pick (D-35).
	if len(offered) != 3 {
		t.Fatalf("offered %d candidates, want 3 (F-cheap, F-pricey, F-perm); got %v", len(offered), offered)
	}
```

Replace `TestPurchaseEnabler_PermissionGate` whole with:
```go
func TestPurchaseEnabler_AsksPermissionAfterThePick(t *testing.T) {
	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().
		SelectPurchaseEnablerAsset(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
			for _, candidate := range candidates {
				if candidate.Def.ID == "F-perm" {
					return candidate.Def, nil
				}
			}
			t.Fatal("F-perm not offered")
			return nil, nil
		})
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	ctx := StepContext{
		Faction:      &domain.Faction{ID: "f1", Force: 5, Coin: 20, Assets: map[string]*domain.Asset{}},
		Asset:        &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "w1"}},
		Step:         enablerStep(),
		FactionState: &state.FactionState{},
		Rulebook:     enablerRulebook(),
		Collector:    collector,
	}
	mutations, err := purchaseEnabler(ctx)
	if err != nil {
		t.Fatalf("purchaseEnabler: %v", err)
	}
	want := []domain.Mutation{domain.PermissionRefused{FactionID: "f1", WorldID: "w1", DefinitionID: "F-perm", Cause: "ability", CausedByFactionID: "f1"}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want only the refusal record (no asset, no charge)", mutations)
	}
}
```
Add `"reflect"` to its imports if missing.

`internal/faction/engine/mutation/mutation_test.go` — append:
```go
func TestMutationEngine_PermissionRefusedIsRecordOnly(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	refusal := domain.PermissionRefused{FactionID: "missing", WorldID: "Anchorage", DefinitionID: "gunships", Cause: "buy"}
	if err := me.Apply(s, []domain.Mutation{refusal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Errorf("Apply = %v, want no miss for a record-only mutation", err)
	}
	if got := s.Factions["f1"].Coin; got != 10 {
		t.Errorf("Coin = %d, want 10 unchanged", got)
	}
}
```

##### Task 20 — interface and digest tests

`internal/faction/tui/views/turn/overlay/asset_test.go` — append:
```go
func TestRequestPermissionIgnoresEsc(t *testing.T) {
	request := hooks.PermissionRequest{
		Requester: &domain.Faction{ID: "f1", Name: "Alpha"},
		Ruler:     &domain.Faction{ID: "f2", Name: "Bravo"},
		WorldID:   "krylos",
		Def:       &domain.AssetDefinition{ID: "gunships", Name: "Gunships"},
	}

	updated, cmd := NewRequestPermission(request, "Krylos").Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(NoCancelConfirm); !ok {
		t.Errorf("Update returned %T, want NoCancelConfirm", updated)
	}
}
```

`internal/faction/narrative/digest/digest_test.go` — append:
```go
func TestPermissionRefusedIsARefusalLine(t *testing.T) {
	factionState := minState(map[string]string{"f1": "Alpha", "f2": "Bravo"})
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"gunships": {ID: "gunships", Name: "Gunships"}}}
	records := []domain.EventRecord{
		eventRec("f1", mutRec(t, "permission_refused", domain.PermissionRefused{
			FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1",
		})),
		eventRec("f1", mutRec(t, "permission_refused", domain.PermissionRefused{
			FactionID: "f1", WorldID: "Hadrian", DefinitionID: "gunships", Cause: "buy", CausedByFactionID: "f1",
		})),
	}
	cycleDigest := buildOne(t, records, factionState, rb)

	beat := findBeat(cycleDigest, "f1")
	if beat == nil {
		t.Fatal("f1 is quiet, want a refusal to make it active")
	}
	want := []RefusalEvent{
		{AssetName: "Gunships", World: "Krylos", DecidedBy: "Bravo"},
		{AssetName: "Gunships", World: "Hadrian"},
	}
	if !reflect.DeepEqual(beat.Refusals, want) {
		t.Errorf("Refusals = %+v, want %+v", beat.Refusals, want)
	}
	if len(beat.Notes) != 0 {
		t.Errorf("Notes = %v, want none", beat.Notes)
	}
}
```
Add `"reflect"` to its imports if missing.

`internal/faction/tui/views/turn/execution/format_test.go` — in
`TestFormatMutation_Surfaced`, after the two `WorldGovernmentChanged` rows (Commit 2
Task 16):

**Find:**
```go
		{"WorldGovernmentChanged independent", domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: "f1"}, "Krylos → independent"},
```
**Replace with:**
```go
		{"WorldGovernmentChanged independent", domain.WorldGovernmentChanged{WorldID: "Krylos", FromFactionID: "f1"}, "Krylos → independent"},
		{"PermissionRefused", domain.PermissionRefused{DefinitionID: "gunships", WorldID: "Krylos"}, "permission refused: gunships on Krylos"},
```

##### Task 21 — architecture docs

`docs/architecture/engine/actions.md` — two rows and one Key Decision.

**Find:**
```markdown
| Buy Asset | Purchase an asset onto a world, inactive until next turn; a quality (Stealth) adds no asset and acts on one the faction holds there | `SelectBuyOrder` (`SelectAsset` for a quality's target) → `AssetAdded`, `CoinDelta` (a quality's effect mutations, such as `AssetStealthApplied`, in place of `AssetAdded`; `PurchaseDiscountConsumed`) |
| Refit Asset | Swap an asset for another of its category, paying the cost difference; a sheltered asset cannot be refit | `SelectRefitOrder` → `AssetRemoved`, `AssetAdded`, `CoinDelta` |
```
**Replace with:**
```markdown
| Buy Asset | Purchase an asset onto a world, inactive until next turn; a quality (Stealth) adds no asset and acts on one the faction holds there; a permission asset asks the world's government after the choice | `SelectBuyOrder` (`SelectAsset` for a quality's target, `RequestPermission`) → `AssetAdded`, `CoinDelta` (a quality's effect mutations, such as `AssetStealthApplied`, in place of `AssetAdded`; `PurchaseDiscountConsumed`; only `PermissionRefused` on a refusal) |
| Refit Asset | Swap an asset for another of its category, paying the cost difference; a sheltered asset or one in transit cannot be refit; a permission replacement asks the world's government | `SelectRefitOrder` (`RequestPermission`) → `AssetRemoved`, `AssetAdded`, `CoinDelta` (only `PermissionRefused` on a refusal) |
```

**Find:**
```markdown
  `domain.CanPurchase` / `PurchaseTerms`, and both consume a
  `PendingPurchaseDiscount` under the same half-base-price floor clamp — so the two
  purchase surfaces cannot drift on cost, permission, or discount rules. Every
```
**Replace with:**
```markdown
  `domain.CanPurchase` / `PurchaseTerms`, and both consume a
  `PendingPurchaseDiscount` under the same half-base-price floor clamp — so the two
  purchase surfaces cannot drift on cost or discount rules. Permission is not a
  menu filter: Buy Asset, the purchase-enabler and Refit each ask
  `dispatch.RequestPermission` after the choice, and a refusal ends the purchase
  with a `PermissionRefused` record and no charge (A.001.6 Effort 4, D-35). Every
```

`docs/architecture/engine/effect-mutation.md` — two edits.

**Find:**
```markdown
[actions](actions.md) ability dispatch instead; `P` is a buy-time government-
permission check. The effect engine owns only the passive `S` side.) Its handlers
```
**Replace with:**
```markdown
[actions](actions.md) ability dispatch instead; `P` asks the world's government at
purchase, through `dispatch.RequestPermission`. The effect engine owns only the
passive `S` side.) Its handlers
```

**Find:**
```markdown
- `world_government_changed` — set or clear a world's ruler in the government record (Planetary Seizure)
```
**Replace with:**
```markdown
- `world_government_changed` — set or clear a world's ruler in the government record (Planetary Seizure)
- `permission_refused` — record only: a government refused a purchase, which bought nothing
```

`docs/architecture/engine/hooks.md`:

**Find:**
```markdown
to pay to keep it. Only `action.Collector` embeds it, because no phase destroys
an asset by damage, and a reactor asks nothing.
```
**Replace with:**
```markdown
to pay to keep it. Only `action.Collector` embeds it, because no phase destroys
an asset by damage, and a reactor asks nothing.

`PermissionCollector` (`RequestPermission(PermissionRequest) (bool, error)`) asks
a world's government whether a faction may raise an asset there. The request
names the ruler, or none for an independent world, where the GM answers.
`dispatch.RequestPermission` decides whether to ask at all; a purchase on the
requester's own homeworld never asks. `action.Collector`
embeds it for the purchase paths; `PhaseCollector` will when arrivals ask
(A.001.6 Effort 5).
```

##### Task 22 — `docs/dev_journals/faction-manager/planned-work.md`

File the general in-transit rule (Decision 11). In the Features table:

**Find:**
```markdown
| `F.027` | TUI adapter tests | next interface initiative | `tui/adapter` (1,176 lines) is the only implementation of the engine seam and has no tests; cover collector and observer paths with the engine test harness; `D.002` N7 |
```
**Replace with:**
```markdown
| `F.027` | TUI adapter tests | next interface initiative | `tui/adapter` (1,176 lines) is the only implementation of the engine seam and has no tests; cover collector and observer paths with the engine test harness; `D.002` N7 |
| `F.028` | In-transit assets can't be acted on | -- | An asset between worlds (no `WorldID`, cargo included) should be untouchable by actions; Refit already refuses it (`A.001.6` Effort 4 Decision 11); audit repair, sale, ability targets and the rest, then block them. |
```

##### Commit message

```
feat: permission request at purchase

- hooks.PermissionCollector.RequestPermission, embedded in action.Collector;
  dispatch.RequestPermission asks the ruler, or the GM for an independent world;
  a ruler's own request and a purchase on the homeworld are granted at once
  (D-4, D-6, Decision 15)
- Buy Asset, the purchase enabler and Refit ask after the choice; a refusal
  is the action's only mutation, PermissionRefused, with no charge (D-35)
- P assets are offered everywhere; HasPlanetaryGovernment and
  PurchaseTerms.HoldsGovernment retire
- no refit for an asset in transit; F.028 files the general rule
- the prompt (Esc ignored), the TUI line and the digest refusal line
```

### Phase 4 — Permission rule queries

**In plain words:** Two assets now change the permission request. Popular Movement
makes its world always give permission to the faction that owns it, and Capital Fleet
now needs permission before a faction can buy it.

#### Commit 5 — `feat: permission rule queries`

##### Task 1 — `internal/faction/domain/asset.go`

Capital Fleet's raise-only permission (Decision 1). Two edits; the second extends
Commit 4 Task 3.

**Find:**
```go
	// Quality marks a definition that adds no asset: buying it runs its one
	// on_purchase entry on an asset the buyer already holds (Stealth).
	Quality bool
}
```
**Replace with:**
```go
	// Quality marks a definition that adds no asset: buying it runs its one
	// on_purchase entry on an asset the buyer already holds (Stealth).
	Quality bool
	// deliberate: a definition field, not a require_permission hook, because the asset being raised has no instance to register a hook on (Effort 4 Decision 1).
	// PurchasePermission asks the government to raise this asset, though moving it
	// in needs none (Capital Fleet).
	PurchasePermission bool
}
```

**Find:**
```go
func (def *AssetDefinition) NeedsPurchasePermission() bool {
	return def.HasFlag(FlagPermission)
}
```
**Replace with:**
```go
func (def *AssetDefinition) NeedsPurchasePermission() bool {
	return def.HasFlag(FlagPermission) || def.PurchasePermission
}
```

##### Task 2 — `internal/faction/domain/effect.go`

Two edits.

**Find:**
```go
	SpecialGrantStealth   SpecialEffectType = "grant_stealth"
)
```
**Replace with:**
```go
	SpecialGrantStealth     SpecialEffectType = "grant_stealth"
	SpecialGrantPermission  SpecialEffectType = "grant_permission"
)
```

**Find:**
```go
	TriggerLethalDamage           EffectTrigger = "lethal_damage"
)
```
**Replace with:**
```go
	TriggerLethalDamage           EffectTrigger = "lethal_damage"
	TriggerPermissionQuery        EffectTrigger = "permission_query"
)
```

##### Task 3 — `internal/faction/rulebook/rulebook.go`, `effects.go`

In `rulebook.go`, `assetRecord`:

**Find:**
```go
	Quality     bool             `toml:"quality"`
}
```
**Replace with** (gofmt realigns the block):
```go
	Quality     bool             `toml:"quality"`
	PurchasePermission bool `toml:"purchase_permission"`
}
```

In `convertAsset`:

**Find:**
```go
		Quality:     r.Quality,
	}, nil
```
**Replace with:**
```go
		Quality:     r.Quality,
		PurchasePermission: r.PurchasePermission,
	}, nil
```

In `effects.go`:

**Find:**
```go
		domain.TriggerAttackEligibilityQuery, domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin, domain.TriggerLethalDamage, domain.TriggerRepairCostQuery,
	}
```
**Replace with:**
```go
		domain.TriggerAttackEligibilityQuery, domain.TriggerOwnAttack, domain.TriggerOwnDefenseWin, domain.TriggerLethalDamage, domain.TriggerRepairCostQuery,
		domain.TriggerPermissionQuery,
	}
```

**Find:**
```go
		domain.SpecialGrantStealth,
	}
```
**Replace with:**
```go
		domain.SpecialGrantStealth, domain.SpecialGrantPermission,
	}
```

##### Task 4 — `internal/faction/engine/hooks/rule_modifier.go`, `types.go`, `doc.go`

In `rule_modifier.go`, append:
```go
// PermissionQuery asks whether Requester's request to raise Def on WorldID is
// granted without asking the government.
type PermissionQuery struct {
	Requester *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
}

// PermissionGranter reports whether a permission request is granted at once
// (Cat 4). A granter asked about a request it does not serve returns false.
type PermissionGranter interface {
	GrantsPermission(query PermissionQuery) bool
}
```

In `types.go`, append:
```go
type RegisteredPermissionGranter struct {
	Scope  Scope
	Source string
	Hook   PermissionGranter
}
```

In `doc.go`:

**Find:**
```go
// WorldTechLevelModifier, AttackEligibilityRule, AssetMovementGranter,
// RepairCostModifier) rather
```
**Replace with:**
```go
// WorldTechLevelModifier, AttackEligibilityRule, AssetMovementGranter,
// RepairCostModifier, PermissionGranter) rather
```

##### Task 5 — `internal/faction/engine/hooks/registry.go`

Five edits, following `RepairCostModifier`.

**Find:**
```go
	repairCostModifiers      map[Scope][]RegisteredRepairCostModifier
}
```
**Replace with:**
```go
	repairCostModifiers      map[Scope][]RegisteredRepairCostModifier
	permissionGranters       map[Scope][]RegisteredPermissionGranter
}
```

**Find:**
```go
	registry.repairCostModifiers = make(map[Scope][]RegisteredRepairCostModifier)
}
```
**Replace with:**
```go
	registry.repairCostModifiers = make(map[Scope][]RegisteredRepairCostModifier)
	registry.permissionGranters = make(map[Scope][]RegisteredPermissionGranter)
}
```

**Find:**
```go
	delete(registry.repairCostModifiers, scope)
}
```
**Replace with:**
```go
	delete(registry.repairCostModifiers, scope)
	delete(registry.permissionGranters, scope)
}
```

Append to the end of the file:
```go
// RegisterPermissionGranter registers a Cat 4 permission-granter hook under the given scope.
func (registry *Registry) RegisterPermissionGranter(scope Scope, source string, hook PermissionGranter) {
	registry.permissionGranters[scope] = append(registry.permissionGranters[scope], RegisteredPermissionGranter{Scope: scope, Source: source, Hook: hook})
}

// PermissionGrantersFor returns all Cat 4 permission-granter hooks matching the given faction and asset.
func (registry *Registry) PermissionGrantersFor(factionID, assetInstanceID string) []RegisteredPermissionGranter {
	var result []RegisteredPermissionGranter
	result = append(result, registry.permissionGranters[GlobalScope()]...)
	result = append(result, registry.permissionGranters[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.permissionGranters[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// PermissionGrantersForAsset returns the Cat 4 permission-granter hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) PermissionGrantersForAsset(factionID, assetInstanceID string) []RegisteredPermissionGranter {
	return registry.permissionGranters[AssetScope(factionID, assetInstanceID)]
}
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/rules.go`, `permission.go`

In `rules.go`, after `GrantedMovementAbilities`:

**Find:**
```go
// ResolveRepairCost runs the repairer's own RepairCostModifiers, then those
```
**Replace with:**
```go
// PermissionGranted reports whether one of query.Requester's own grant hooks grants
// the request: its global and faction granters, then those registered against each
// asset it holds, in asset ID order. A grant serves its owner only, so a rival's
// assets are never asked; the first yes ends the walk. Returns false when registry
// is nil.
func PermissionGranted(registry *hooks.Registry, query hooks.PermissionQuery) bool {
	if registry == nil {
		return false
	}
	for _, registered := range registry.PermissionGrantersFor(query.Requester.ID, "") {
		if registered.Hook.GrantsPermission(query) {
			return true
		}
	}
	for _, asset := range domain.SortedAssets(query.Requester) {
		for _, registered := range registry.PermissionGrantersForAsset(query.Requester.ID, asset.ID) {
			if registered.Hook.GrantsPermission(query) {
				return true
			}
		}
	}
	return false
}

// ResolveRepairCost runs the repairer's own RepairCostModifiers, then those
```

In `permission.go` (Commit 4 Task 6), the request consults the grant hooks:

**Find:**
```go
// requester's homeworld (Effort 4 Decision 15). A ruler's own request is granted at
// once (D-6). Otherwise the ruler's driver is asked, or the GM for an independent
// world (D-4). cause names the purchase path on the record.
func RequestPermission(collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	rulerID := factionState.Ruler(worldID)
	if rulerID == requester.ID {
		return nil, nil
	}
```
**Replace with:**
```go
// requester's homeworld (Effort 4 Decision 15). A ruler's own request is granted at
// once (D-6), and so is one a grant hook of the requester serves (Popular Movement).
// Otherwise the ruler's driver is asked, or the GM for an independent world (D-4).
// cause names the purchase path on the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	rulerID := factionState.Ruler(worldID)
	if rulerID == requester.ID {
		return nil, nil
	}
	if PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def}) {
		return nil, nil
	}
```

##### Task 7 — the three callers and Refit's registry (overview Decision 18)

`buy_asset.go`:

**Find:**
```go
	refusal, err := dispatch.RequestPermission(ba.collector, factionState, faction, order.Definition, order.World, "buy")
```
**Replace with:**
```go
	refusal, err := dispatch.RequestPermission(ba.registry, ba.collector, factionState, faction, order.Definition, order.World, "buy")
```

`ability/purchase_enabler.go`:

**Find:**
```go
	refusal, err := dispatch.RequestPermission(ctx.Collector, ctx.FactionState, ctx.Faction, def, worldID, "ability")
```
**Replace with:**
```go
	refusal, err := dispatch.RequestPermission(ctx.Registry, ctx.Collector, ctx.FactionState, ctx.Faction, def, worldID, "ability")
```

`refit_asset.go` — four edits:

**Find:**
```go
type RefitAsset struct {
	collector  action.Collector
	factionID  string
```
**Replace with:**
```go
type RefitAsset struct {
	collector  action.Collector
	registry   *hooks.Registry
	factionID  string
```

**Find:**
```go
func NewRefitAsset(collector action.Collector) *RefitAsset {
	return &RefitAsset{collector: collector}
}
```
**Replace with:**
```go
func NewRefitAsset(collector action.Collector, registry *hooks.Registry) *RefitAsset {
	return &RefitAsset{collector: collector, registry: registry}
}
```

**Find:**
```go
	refusal, err := dispatch.RequestPermission(ra.collector, factionState, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")
```
**Replace with:**
```go
	refusal, err := dispatch.RequestPermission(ra.registry, ra.collector, factionState, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")
```

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"` to its
imports.

`register.go`:

**Find:**
```go
	ae.Register(func(c action.Collector) action.Action { return NewRefitAsset(c) })
```
**Replace with:**
```go
	ae.Register(func(c action.Collector) action.Action { return NewRefitAsset(c, hookRegistry) })
```

In `refit_asset_test.go`, every `NewRefitAsset(x)` call gains a second argument
`nil`. In `dispatch/permission_test.go`, the one `RequestPermission(collector, ...)`
call gains a leading `nil`.

##### Task 8 — `internal/faction/engine/effect/effects/grant_permission.go` (new)

Full contents:
```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantPermissionRegistrar is Popular Movement: on the declaring asset's world,
// its owner's requests are granted at once. owner = "own" reads as "the requester
// is the owner", scope = "world" as "the request is for the declaring asset's world".
var grantPermissionRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPermissionQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
		},
	},
	Register: registerGrantPermission,
}

func registerGrantPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionGranter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialGrantPermission)+":"+ctx.Def.ID,
		&permissionGranter{owner: ctx.Faction, assetID: ctx.Asset.ID},
	)
}

// permissionGranter holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19).
type permissionGranter struct {
	owner   *domain.Faction
	assetID string
}

func (granter *permissionGranter) GrantsPermission(query hooks.PermissionQuery) bool {
	self := granter.owner.Assets[granter.assetID]
	if !active(self) || query.Requester.ID != granter.owner.ID {
		return false
	}
	return self.Location.WorldID != "" && self.Location.WorldID == query.WorldID
}
```

In `effects/registrars.go`:

**Find:**
```go
		domain.SpecialGrantStealth:   grantStealthRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialGrantStealth:    grantStealthRegistrar,
		domain.SpecialGrantPermission: grantPermissionRegistrar,
	}
```

##### Task 9 — `rulebooks/swn/assets/force_assets.toml`, `cunning_assets.toml`

Capital Fleet, in `force_assets.toml`:

**Find:**
```toml
type = "Starship"
flags = ["S"]
counter = "3d8"
description = "The pride of an empire, a collection of massive capital warships without peer in most sectors.
```
**Replace with:**
```toml
type = "Starship"
flags = ["S"]
purchase_permission = true
counter = "3d8"
description = "The pride of an empire, a collection of massive capital warships without peer in most sectors.
```
(The `description` line continues unchanged; match on its first sentence.)

Popular Movement, in `cunning_assets.toml`:

**Find:**
```toml
[assets.PopularMovement.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"
damage = "2d6"
```
**Replace with:**
```toml
[assets.PopularMovement.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"
damage = "2d6"

[[assets.PopularMovement.effects]]
trigger = "permission_query"
effect = "grant_permission"
[assets.PopularMovement.effects.filters]
owner = "own"
scope = "world"
```

##### Task 10 — `internal/faction/rulebook/rulebook_test.go`

In `TestLoad`:

**Find:**
```go
	t.Run("stealth is a quality with grant_stealth", func(t *testing.T) {
```
**Replace with:**
```go
	t.Run("capital fleet needs permission to be raised", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-F8-001"]
		if !ok {
			t.Fatal("asset SWN-F8-001 not found")
		}
		if asset.HasFlag(domain.FlagPermission) {
			t.Error("Capital Fleet carries P, want the book's flags (A, S) only (D-26)")
		}
		if !asset.NeedsPurchasePermission() {
			t.Error("NeedsPurchasePermission = false, want true")
		}
	})

	t.Run("popular movement grant_permission effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C7-001"]
		if !ok {
			t.Fatal("asset SWN-C7-001 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger: domain.TriggerPermissionQuery,
			Effect:  domain.SpecialGrantPermission,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("stealth is a quality with grant_stealth", func(t *testing.T) {
```

##### Task 11 — `internal/faction/engine/effect/effects/grant_permission_test.go` (new)

Full contents:
```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func registerPopularMovement(t *testing.T, owner *domain.Faction, movement *domain.Asset) hooks.PermissionGranter {
	t.Helper()
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialGrantPermission].Register(effect.RegisterContext{
		Faction: owner, Asset: movement, Def: &domain.AssetDefinition{ID: "SWN-C7-001"}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerPermissionQuery,
			Effect:  domain.SpecialGrantPermission,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld},
		},
	})
	granters := registry.PermissionGrantersForAsset(owner.ID, movement.ID)
	if len(granters) != 1 {
		t.Fatalf("granters = %d, want 1", len(granters))
	}
	return granters[0].Hook
}

func TestGrantPermission_GrantsItsOwnerOnItsWorld(t *testing.T) {
	movement := &domain.Asset{ID: "pm", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 16}
	alpha := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{"pm": movement}}
	bravo := &domain.Faction{ID: "bravo"}
	granter := registerPopularMovement(t, alpha, movement)
	def := &domain.AssetDefinition{ID: "gunships"}

	cases := []struct {
		name  string
		query hooks.PermissionQuery
		want  bool
	}{
		{"its owner on its world", hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}, true},
		{"its owner elsewhere", hooks.PermissionQuery{Requester: alpha, WorldID: "Hadrian", Def: def}, false},
		{"a rival on its world", hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := granter.GrantsPermission(tc.query); got != tc.want {
				t.Errorf("GrantsPermission = %v, want %v", got, tc.want)
			}
		})
	}

	movement.Ready = false
	if granter.GrantsPermission(hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}) {
		t.Error("an inactive Popular Movement granted, want no grant")
	}
}
```

##### Task 12 — `internal/faction/engine/hooks/dispatch/permission_test.go`

The grant hook ends the request before the prompt. Append:
```go
type stubGranter struct{ grant bool }

func (stub *stubGranter) GrantsPermission(hooks.PermissionQuery) bool { return stub.grant }

func TestRequestPermission_AGrantHookIsNotAsked(t *testing.T) {
	factionState, requester, _ := permissionState()
	requester.Assets = map[string]*domain.Asset{"pm": {ID: "pm"}}
	registry := hooks.NewRegistry()
	registry.RegisterPermissionGranter(hooks.AssetScope("f1", "pm"), "pm", &stubGranter{grant: true})
	collector := &stubPermissionCollector{}
	def := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}

	refusal, err := RequestPermission(registry, collector, factionState, requester, def, "Krylos", "buy")

	if err != nil || refusal != nil {
		t.Errorf("RequestPermission = (%+v, %v), want a grant", refusal, err)
	}
	if len(collector.asked) != 0 {
		t.Errorf("asked = %+v, want no prompt when a grant hook serves the request", collector.asked)
	}
}

func TestPermissionGranted_AsksOnlyTheRequestersHooks(t *testing.T) {
	_, requester, ruler := permissionState()
	ruler.Assets = map[string]*domain.Asset{"pm": {ID: "pm"}}
	registry := hooks.NewRegistry()
	registry.RegisterPermissionGranter(hooks.AssetScope("f2", "pm"), "pm", &stubGranter{grant: true})

	if PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: "Krylos"}) {
		t.Error("a rival's grant hook granted the requester, want a grant to serve its owner only")
	}
	if PermissionGranted(nil, hooks.PermissionQuery{Requester: requester}) {
		t.Error("a nil registry granted, want false")
	}
}
```

##### Task 13 — `internal/faction/engine/hooks/registry_test.go`

In `TestClearAsset_RemovesOnlyThatAssetsBucket`:

**Find:**
```go
	registry.RegisterRepairCostModifier(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubRepairCostModifier{})
	registry.RegisterMutationReactor(hooks.AssetScope("f1", "asset-Y"), "asset-Y", &stubMutationReactor{})
```
**Replace with:**
```go
	registry.RegisterRepairCostModifier(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubRepairCostModifier{})
	registry.RegisterPermissionGranter(hooks.AssetScope("f1", "asset-X"), "asset-X", &stubPermissionGranter{})
	registry.RegisterMutationReactor(hooks.AssetScope("f1", "asset-Y"), "asset-Y", &stubMutationReactor{})
```

**Find:**
```go
	if got := registry.RepairCostModifiersForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("repair-cost modifiers under f1/asset-X = %v, want none", got)
	}
```
**Replace with:**
```go
	if got := registry.RepairCostModifiersForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("repair-cost modifiers under f1/asset-X = %v, want none", got)
	}
	if got := registry.PermissionGrantersForAsset("f1", "asset-X"); len(got) != 0 {
		t.Errorf("permission granters under f1/asset-X = %v, want none", got)
	}
```

**Find:**
```go
type stubRepairCostModifier struct{}
```
**Replace with:**
```go
type stubPermissionGranter struct{}

func (stub *stubPermissionGranter) GrantsPermission(hooks.PermissionQuery) bool { return false }

type stubRepairCostModifier struct{}
```

##### Task 14 — action tests

`internal/faction/engine/action/actions/buy_asset_test.go` — append:
```go
func TestBuyAsset_PurchasePermissionAsksForARaise(t *testing.T) {
	ctrl := gomock.NewController(t)
	def := &domain.AssetDefinition{ID: "fleet", Name: "Capital Fleet", Category: domain.StatForce, HP: 30, Cost: 4, MinRating: 1, PurchasePermission: true}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"fleet": def}}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: "Tartarus", Definition: def}, nil)
	collector.EXPECT().RequestPermission(hooks.PermissionRequest{Requester: faction, WorldID: "Tartarus", Def: def}).Return(false, nil)

	act := NewBuyAsset(collector, nil, nil, nil)
	if err := act.Inputs(faction, factionState, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want only the refusal record", mutations)
	}
	if _, ok := mutations[0].(domain.PermissionRefused); !ok {
		t.Errorf("mutations[0] = %T, want PermissionRefused", mutations[0])
	}
}
```

##### Task 15 — docs

`docs/architecture/engine/hooks.md` — two edits.

**Find:**
```markdown
`MaintenanceCostModifier`, `WorldTechLevelModifier`, `AttackEligibilityRule`,
`AssetMovementGranter`, `RepairCostModifier` — each with its own signature
```
**Replace with:**
```markdown
`MaintenanceCostModifier`, `WorldTechLevelModifier`, `AttackEligibilityRule`,
`AssetMovementGranter`, `RepairCostModifier`, `PermissionGranter` — each with its own signature
```

**Find:**
```markdown
- **`InterceptLethal`** is Cat 6. It runs when a hit would destroy an asset,
```
**Replace with:**
```markdown
- **`PermissionGranted`** is the sixth Cat 4 resolver, asked by
  `RequestPermission` before it prompts. It walks the requester's global and
  faction granters, then each of its own assets in ID order, and the first yes
  grants (Popular Movement). A grant serves its owner only. Capital Fleet's
  raise-only permission is not a hook: the fleet being raised has no instance, so
  it is the definition field `purchase_permission` (A.001.6 Effort 4 Decision 1).
- **`InterceptLethal`** is Cat 6. It runs when a hit would destroy an asset,
```

`docs/architecture/engine/ability-catalog.md`:

**Find:**
```markdown
| **auto-permission / revoke-permission** | Cat 4 `RuleModifier` | Popular Movement `SWN-C7-001` (grant), Lobbyists `SWN-C2-001` (revoke, reactive contest) |
```
**Replace with:**
```markdown
| **auto-permission / revoke-permission** | Cat 4 `PermissionGranter` → `grant_permission` (grant); revoke in Effort 5 | Popular Movement `SWN-C7-001` ✅ (grant), Lobbyists `SWN-C2-001` (revoke, reactive contest) |
```

##### Commit message

```
feat: permission rule queries

- PermissionGranter, the Cat 4 permission family's one member, and its
  PermissionGranted resolver; RequestPermission grants on a hook before it asks
- Popular Movement: permission_query → grant_permission on its world
- Capital Fleet: purchase_permission = true, a raise-only definition field
  (reshapes D-26; closes B.009)
- Refit takes the hook registry (overview Decision 18)
```
