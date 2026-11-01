package dispatch

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// stubAssetCostModifier reduces cost by a fixed delta.
type stubAssetCostModifier struct{ delta int }

func (stub *stubAssetCostModifier) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	query.Cost -= stub.delta
}

// stubMaintenanceCostModifier reduces cost by a fixed delta.
type stubMaintenanceCostModifier struct{ delta int }

func (stub *stubMaintenanceCostModifier) ModifyMaintenanceCost(_ *domain.Faction, _ *domain.Asset, cost int) int {
	return cost - stub.delta
}

// stubTieResolver always returns a fixed outcome.
type stubTieResolver struct{ outcome hooks.TieOutcome }

func (stub *stubTieResolver) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return stub.outcome
}

func TestResolveAssetCost_NoModifiers(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, nil).Cost
	if got != 5 {
		t.Errorf("ResolveAssetCost = %d, want 5 (no modifiers)", got)
	}
}

func TestResolveAssetCost_OneModifier(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "test", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, nil).Cost
	if got != 4 {
		t.Errorf("ResolveAssetCost = %d, want 4 (base 5 - 1)", got)
	}
}

func TestResolveAssetCost_TwoModifiersChain(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 10}
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "first", &stubAssetCostModifier{delta: 2})
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "second", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 10, nil).Cost
	// first reduces 10→8, second sees 8 and reduces 8→7
	if got != 7 {
		t.Errorf("ResolveAssetCost = %d, want 7 (10-2-1 chained)", got)
	}
}

func TestResolveAssetCost_RivalAssetOnWorld(t *testing.T) {
	registry := hooks.NewRegistry()
	buyer := &domain.Faction{ID: "buyer"}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	rivalAsset := &domain.Asset{ID: "rival-a1", OwnerID: "rival"}
	registry.RegisterAssetCostModifier(hooks.AssetScope("rival", "rival-a1"), "surcharge", &stubAssetCostModifier{delta: -1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, []*domain.Asset{rivalAsset}).Cost
	if got != 6 {
		t.Errorf("ResolveAssetCost = %d, want 6 (rival asset on world adds 1)", got)
	}
}

func TestResolveAssetCost_BuyersOwnAssetAskedOnce(t *testing.T) {
	registry := hooks.NewRegistry()
	ownAsset := &domain.Asset{ID: "own-a1", OwnerID: "buyer"}
	buyer := &domain.Faction{ID: "buyer", Assets: map[string]*domain.Asset{ownAsset.ID: ownAsset}}
	def := &domain.AssetDefinition{ID: "cheap", Cost: 5}
	registry.RegisterAssetCostModifier(hooks.AssetScope("buyer", "own-a1"), "discount", &stubAssetCostModifier{delta: 1})
	got := ResolveAssetCost(registry, buyer, def, "Tartarus", 5, []*domain.Asset{ownAsset}).Cost
	if got != 4 {
		t.Errorf("ResolveAssetCost = %d, want 4 (the buyer's own asset asked once, not again as a world asset)", got)
	}
}

func TestDiscountedCost(t *testing.T) {
	cases := []struct {
		name     string
		quote    hooks.AssetCostQuery
		baseCost int
		discount int
		want     int
	}{
		{"discount above half base floors at half base", hooks.AssetCostQuery{Cost: 4}, 4, 3, 2},
		{"small discount subtracts", hooks.AssetCostQuery{Cost: 6}, 6, 1, 5},
		{"floor never above a hook-reduced price", hooks.AssetCostQuery{Cost: 1}, 4, 3, 1},
		{"surcharge survives a floored discount", hooks.AssetCostQuery{Cost: 5, Surcharges: []hooks.Surcharge{{FactionID: "rival", Amount: 1}}}, 4, 3, 3},
		{"surcharge survives a small discount", hooks.AssetCostQuery{Cost: 7, Surcharges: []hooks.Surcharge{{FactionID: "rival", Amount: 1}}}, 6, 1, 6},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := DiscountedCost(testCase.quote, testCase.baseCost, testCase.discount); got != testCase.want {
				t.Errorf("DiscountedCost = %d, want %d", got, testCase.want)
			}
		})
	}
}

// stubWorldTechLevelModifier raises the tech level by a fixed delta and records
// the definition it was asked about.
type stubWorldTechLevelModifier struct {
	delta   int
	seenDef *domain.AssetDefinition
}

func (stub *stubWorldTechLevelModifier) ModifyWorldTechLevel(_ *domain.Faction, def *domain.AssetDefinition, _ string, baseTL int) int {
	stub.seenDef = def
	return baseTL + stub.delta
}

func TestResolveWorldTechLevel_OwnAssetOffWorld(t *testing.T) {
	registry := hooks.NewRegistry()
	asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Elsewhere"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
	def := &domain.AssetDefinition{ID: "lab-good"}
	registry.RegisterWorldTechLevelModifier(hooks.AssetScope("f1", "a1"), "floor", &stubWorldTechLevelModifier{delta: 1})
	got := ResolveWorldTechLevel(registry, faction, def, "Tartarus", 3)
	if got != 4 {
		t.Errorf("ResolveWorldTechLevel = %d, want 4 (own asset off world is asked; the hook decides reach)", got)
	}
}

func TestResolveWorldTechLevel_RivalAssetNotAsked(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "f1"}
	def := &domain.AssetDefinition{ID: "lab-good"}
	registry.RegisterWorldTechLevelModifier(hooks.AssetScope("rival", "rival-a1"), "floor", &stubWorldTechLevelModifier{delta: 1})
	got := ResolveWorldTechLevel(registry, faction, def, "Tartarus", 3)
	if got != 3 {
		t.Errorf("ResolveWorldTechLevel = %d, want 3 (rival's asset is never asked)", got)
	}
}

func TestResolveWorldTechLevel_PassesDefinition(t *testing.T) {
	registry := hooks.NewRegistry()
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
	def := &domain.AssetDefinition{ID: "lab-good"}
	stub := &stubWorldTechLevelModifier{}
	registry.RegisterWorldTechLevelModifier(hooks.AssetScope("f1", "a1"), "floor", stub)
	ResolveWorldTechLevel(registry, faction, def, "Tartarus", 3)
	if stub.seenDef != def {
		t.Errorf("hook saw def %v, want %v", stub.seenDef, def)
	}
}

// stubAssetSpeedModifier raises speed to a fixed floor.
type stubAssetSpeedModifier struct{ floor int }

func (stub stubAssetSpeedModifier) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	return max(speed, stub.floor)
}

func TestAssetSpeedResolver(t *testing.T) {
	definition := &domain.AssetDefinition{ID: "facility"}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": {ID: "a1", OwnerID: "f1"}}}
	if got := AssetSpeedResolver(nil, faction)(definition); got != 0 {
		t.Errorf("nil registry: speed = %d, want the definition's 0", got)
	}

	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.AssetScope("f1", "a1"), "own asset", stubAssetSpeedModifier{floor: 1})
	registry.RegisterAssetSpeedModifier(hooks.AssetScope("rival", "r1"), "rival", stubAssetSpeedModifier{floor: 9})
	if got := AssetSpeedResolver(registry, faction)(definition); got != 1 {
		t.Errorf("speed = %d, want 1: the own asset's floor is asked, the rival's never", got)
	}

	registry.RegisterAssetSpeedModifier(hooks.FactionScope("f1"), "faction", stubAssetSpeedModifier{floor: 2})
	if got := AssetSpeedResolver(registry, faction)(definition); got != 2 {
		t.Errorf("speed = %d, want 2 from the faction's floor", got)
	}
}

func TestResolveMaintenanceCost_NoModifiers(t *testing.T) {
	registry := hooks.NewRegistry()
	owner := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}
	got := ResolveMaintenanceCost(registry, owner, asset, 0)
	if got != 0 {
		t.Errorf("ResolveMaintenanceCost = %d, want 0", got)
	}
}

func TestResolveMaintenanceCost_OneModifier(t *testing.T) {
	registry := hooks.NewRegistry()
	owner := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}
	registry.RegisterMaintenanceCostModifier(hooks.FactionScope("f1"), "test", &stubMaintenanceCostModifier{delta: -2})
	got := ResolveMaintenanceCost(registry, owner, asset, 0)
	// delta=-2 means cost increases by 2
	if got != 2 {
		t.Errorf("ResolveMaintenanceCost = %d, want 2", got)
	}
}

func TestResolveTie_NoResolver(t *testing.T) {
	registry := hooks.NewRegistry()
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieStandard {
		t.Errorf("ResolveTie = %v, want TieStandard when no resolver registered", got)
	}
}

func TestResolveTie_DefenderWins(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "fanatical", &stubTieResolver{outcome: hooks.TieDefenderWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieDefenderWins {
		t.Errorf("ResolveTie = %v, want TieDefenderWins", got)
	}
}

func TestResolveTie_AssetResolverAnswers(t *testing.T) {
	registry := hooks.NewRegistry()
	attacker := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": {ID: "a1"}}}
	registry.RegisterTieResolver(hooks.AssetScope("f1", "a1"), "asset", &stubTieResolver{outcome: hooks.TieAttackerWins})
	if got := ResolveTie(registry, hooks.RollContext{Actor: attacker}, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want TieAttackerWins from the attacker's asset", got)
	}
}

func TestResolveTie_StandardDoesNotEndTheSearch(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "silent", &stubTieResolver{outcome: hooks.TieStandard})
	registry.RegisterTieResolver(hooks.FactionScope("f2"), "loses", &stubTieResolver{outcome: hooks.TieAttackerWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}, Opponent: &domain.Faction{ID: "f2"}}
	if got := ResolveTie(registry, ctx, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want the defender's TieAttackerWins past the attacker's TieStandard", got)
	}
}

// stubEligibilityRule records each side it is asked as and refuses the sides in refuse.
type stubEligibilityRule struct {
	refuse map[hooks.AttackSide]bool
	asked  []hooks.AttackSide
}

func (stub *stubEligibilityRule) MayFight(query hooks.AttackEligibilityQuery) bool {
	stub.asked = append(stub.asked, query.Side)
	return !stub.refuse[query.Side]
}

func eligibilityQuery(side hooks.AttackSide) hooks.AttackEligibilityQuery {
	return hooks.AttackEligibilityQuery{
		Attacker: &domain.Asset{ID: "a1", OwnerID: "f1"},
		Defender: &domain.Asset{ID: "d1", OwnerID: "f2"},
		Side:     side,
	}
}

func TestResolveAttackEligibility_NilRegistryAllows(t *testing.T) {
	if !ResolveAttackEligibility(nil, eligibilityQuery(hooks.SideAttack)) {
		t.Error("ResolveAttackEligibility(nil) = false, want true")
	}
}

func TestResolveAttackEligibility_NoRulesAllows(t *testing.T) {
	if !ResolveAttackEligibility(hooks.NewRegistry(), eligibilityQuery(hooks.SideAttack)) {
		t.Error("ResolveAttackEligibility = false, want true with no rules")
	}
}

func TestResolveAttackEligibility_AttackAsksStrikerThenTargetAsDefense(t *testing.T) {
	registry := hooks.NewRegistry()
	striker := &stubEligibilityRule{}
	target := &stubEligibilityRule{}
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f1", "a1"), "striker", striker)
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f2", "d1"), "target", target)

	if !ResolveAttackEligibility(registry, eligibilityQuery(hooks.SideAttack)) {
		t.Fatal("ResolveAttackEligibility = false, want true when no rule refuses")
	}
	if len(striker.asked) != 1 || striker.asked[0] != hooks.SideAttack {
		t.Errorf("striker asked as %v, want [attack]", striker.asked)
	}
	if len(target.asked) != 1 || target.asked[0] != hooks.SideDefense {
		t.Errorf("target asked as %v, want [defense]", target.asked)
	}
}

func TestResolveAttackEligibility_StrikerRefuses(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f1", "a1"), "striker", &stubEligibilityRule{refuse: map[hooks.AttackSide]bool{hooks.SideAttack: true}})
	if ResolveAttackEligibility(registry, eligibilityQuery(hooks.SideAttack)) {
		t.Error("ResolveAttackEligibility = true, want false when the striker refuses")
	}
}

func TestResolveAttackEligibility_TargetRefusesDefense(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f2", "d1"), "target", &stubEligibilityRule{refuse: map[hooks.AttackSide]bool{hooks.SideDefense: true}})
	if ResolveAttackEligibility(registry, eligibilityQuery(hooks.SideAttack)) {
		t.Error("ResolveAttackEligibility = true, want false when the target refuses to defend")
	}
}

func TestResolveAttackEligibility_CounterDoesNotAskTarget(t *testing.T) {
	registry := hooks.NewRegistry()
	striker := &stubEligibilityRule{}
	target := &stubEligibilityRule{refuse: map[hooks.AttackSide]bool{hooks.SideDefense: true}}
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f1", "a1"), "striker", striker)
	registry.RegisterAttackEligibilityRule(hooks.AssetScope("f2", "d1"), "target", target)

	if !ResolveAttackEligibility(registry, eligibilityQuery(hooks.SideCounter)) {
		t.Error("ResolveAttackEligibility = false, want true: a counter is not a defense")
	}
	if len(striker.asked) != 1 || striker.asked[0] != hooks.SideCounter {
		t.Errorf("striker asked as %v, want [counter]", striker.asked)
	}
	if len(target.asked) != 0 {
		t.Errorf("target asked as %v, want not asked on a counter", target.asked)
	}
}

func TestResolveAttackEligibility_GlobalAndFactionRulesAskedOnce(t *testing.T) {
	registry := hooks.NewRegistry()
	global := &stubEligibilityRule{}
	strikerFaction := &stubEligibilityRule{}
	targetFaction := &stubEligibilityRule{}
	registry.RegisterAttackEligibilityRule(hooks.GlobalScope(), "global", global)
	registry.RegisterAttackEligibilityRule(hooks.FactionScope("f1"), "striker-faction", strikerFaction)
	registry.RegisterAttackEligibilityRule(hooks.FactionScope("f2"), "target-faction", targetFaction)

	ResolveAttackEligibility(registry, eligibilityQuery(hooks.SideAttack))

	if len(global.asked) != 1 || global.asked[0] != hooks.SideAttack {
		t.Errorf("global rule asked as %v, want [attack] once", global.asked)
	}
	if len(strikerFaction.asked) != 1 {
		t.Errorf("striker's faction rule asked %d times, want 1", len(strikerFaction.asked))
	}
	if len(targetFaction.asked) != 0 {
		t.Errorf("target's faction rule asked as %v, want not asked: the target answers through its asset only", targetFaction.asked)
	}
}

// stubMovementGranter records each ask and offers one grant whose Source names it.
type stubMovementGranter struct {
	name  string
	asked *[]string
}

func (stub *stubMovementGranter) GrantMovementAbilities(_ *domain.Faction) []hooks.MovementAbility {
	*stub.asked = append(*stub.asked, stub.name)
	return []hooks.MovementAbility{{Source: &domain.Asset{ID: stub.name}}}
}

func TestGrantedMovementAbilities_WalksFactionThenAssets(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
		"a": {ID: "a", OwnerID: "f1"},
		"b": {ID: "b", OwnerID: "f1"},
	}}
	var asked []string
	registry.RegisterAssetMovementGranter(hooks.AssetScope("f1", "b"), "b", &stubMovementGranter{name: "b", asked: &asked})
	registry.RegisterAssetMovementGranter(hooks.AssetScope("f1", "a"), "a", &stubMovementGranter{name: "a", asked: &asked})
	registry.RegisterAssetMovementGranter(hooks.FactionScope("f1"), "faction", &stubMovementGranter{name: "faction", asked: &asked})
	registry.RegisterAssetMovementGranter(hooks.AssetScope("rival", "r1"), "rival", &stubMovementGranter{name: "rival", asked: &asked})

	abilities := GrantedMovementAbilities(registry, faction)

	var sources []string
	for _, ability := range abilities {
		sources = append(sources, ability.Source.ID)
	}
	want := []string{"faction", "a", "b"}
	if !slices.Equal(sources, want) {
		t.Errorf("grant sources = %v, want %v (faction, then own assets in ID order)", sources, want)
	}
	if slices.Contains(asked, "rival") {
		t.Errorf("asked = %v: a rival's asset granter must not be asked", asked)
	}
}

func TestGrantedMovementAbilities_NilRegistry(t *testing.T) {
	if got := GrantedMovementAbilities(nil, &domain.Faction{ID: "f1"}); got != nil {
		t.Errorf("GrantedMovementAbilities(nil) = %v, want nil", got)
	}
}

// stubRepairCostModifier takes a fixed amount off the repair cost.
type stubRepairCostModifier struct{ delta int }

func (stub *stubRepairCostModifier) ModifyRepairCost(_ *domain.Faction, _ *domain.Asset, _ *domain.AssetDefinition, query *hooks.RepairCostQuery) {
	query.Cost -= stub.delta
}

func TestResolveRepairCost(t *testing.T) {
	repairer := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"own": {ID: "own", OwnerID: "f1"}}}
	repaired := repairer.Assets["own"]
	def := &domain.AssetDefinition{ID: "unit"}

	t.Run("nil registry returns the base cost", func(t *testing.T) {
		if got := ResolveRepairCost(nil, repairer, repaired, def, 3); got != 3 {
			t.Errorf("cost = %d, want 3", got)
		}
	})
	t.Run("a modifier on the repairer's own asset lowers the cost", func(t *testing.T) {
		registry := hooks.NewRegistry()
		registry.RegisterRepairCostModifier(hooks.AssetScope("f1", "own"), "own", &stubRepairCostModifier{delta: 1})
		if got := ResolveRepairCost(registry, repairer, repaired, def, 3); got != 2 {
			t.Errorf("cost = %d, want 2", got)
		}
	})
	t.Run("a modifier on a rival's asset is not asked", func(t *testing.T) {
		registry := hooks.NewRegistry()
		registry.RegisterRepairCostModifier(hooks.AssetScope("rival", "r1"), "rival", &stubRepairCostModifier{delta: 1})
		if got := ResolveRepairCost(registry, repairer, repaired, def, 3); got != 3 {
			t.Errorf("cost = %d, want 3", got)
		}
	})
	t.Run("the cost never drops below 0", func(t *testing.T) {
		registry := hooks.NewRegistry()
		registry.RegisterRepairCostModifier(hooks.AssetScope("f1", "own"), "own", &stubRepairCostModifier{delta: 5})
		if got := ResolveRepairCost(registry, repairer, repaired, def, 1); got != 0 {
			t.Errorf("cost = %d, want 0", got)
		}
	})
}

// stubPopulationFloor answers percent of the setting's floor, never raising it.
type stubPopulationFloor struct{ percent int }

func (stub *stubPopulationFloor) ModifyPopulationFloor(_ *domain.Faction, _ *domain.AssetDefinition, baseFloor, floor int) int {
	return min(floor, baseFloor*stub.percent/100)
}

func TestResolvePopulationFloor(t *testing.T) {
	floors := map[domain.AssetType]int{domain.TypeStarship: 500000}
	starship := &domain.AssetDefinition{ID: "ship", Type: domain.TypeStarship}
	facility := &domain.AssetDefinition{ID: "plant", Type: domain.TypeFacility}
	asset := &domain.Asset{ID: "a1", OwnerID: "f1"}

	plain := &domain.Faction{ID: "f1"}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), plain, starship, floors); got != 500000 {
		t.Errorf("setting floor = %d, want 500000", got)
	}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), plain, facility, floors); got != 0 {
		t.Errorf("type with no floor = %d, want 0", got)
	}
	waived := &domain.Faction{ID: "f1", PopulationWaiver: true}
	if got := ResolvePopulationFloor(hooks.NewRegistry(), waived, starship, floors); got != 0 {
		t.Errorf("waived = %d, want 0", got)
	}

	registry := hooks.NewRegistry()
	registry.RegisterPopulationFloorModifier(hooks.FactionScope("f1"), "colonists", &stubPopulationFloor{percent: 20})
	registry.RegisterPopulationFloorModifier(hooks.AssetScope("f1", "a1"), "expertise", &stubPopulationFloor{percent: 2})
	registry.RegisterPopulationFloorModifier(hooks.AssetScope("rival", "r1"), "rival", &stubPopulationFloor{percent: 1})
	owner := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": asset}}
	if got := ResolvePopulationFloor(registry, owner, starship, floors); got != 10000 {
		t.Errorf("two floors = %d, want 10000 (the lowest wins; a rival's is never asked)", got)
	}
}

// stubToll charges a fixed amount on one world.
type stubToll struct {
	world  string
	amount int
}

func (stub *stubToll) ArrivalToll(_ *domain.Faction, world string) int {
	if world != stub.world {
		return 0
	}
	return stub.amount
}

func TestArrivalTolls_EveryRivalChargesInIDOrder(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterArrivalTollRule(hooks.FactionScope("zulu"), "pirates", &stubToll{world: "Krylos", amount: 1})
	registry.RegisterArrivalTollRule(hooks.AssetScope("bravo", "b1"), "toll-asset", &stubToll{world: "Krylos", amount: 2})
	registry.RegisterArrivalTollRule(hooks.FactionScope("alpha"), "own", &stubToll{world: "Krylos", amount: 5})
	registry.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "elsewhere", &stubToll{world: "Hadrian", amount: 3})
	alpha := &domain.Faction{ID: "alpha"}
	factions := map[string]*domain.Faction{
		"alpha": alpha,
		"bravo": {ID: "bravo", Assets: map[string]*domain.Asset{"b1": {ID: "b1", OwnerID: "bravo"}}},
		"zulu":  {ID: "zulu"},
	}

	got := ArrivalTolls(registry, factions, alpha, "Krylos")

	want := []hooks.ArrivalToll{{FactionID: "bravo", Amount: 2}, {FactionID: "zulu", Amount: 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tolls = %+v, want %+v (the mover's own rule never charges it)", got, want)
	}
}

func TestAssetArrivalTolls(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "pirates", &stubToll{world: "Krylos", amount: 1})
	factions := map[string]*domain.Faction{"alpha": {ID: "alpha"}, "bravo": {ID: "bravo"}}

	cases := []struct {
		name  string
		asset *domain.Asset
		want  []hooks.ArrivalToll
	}{
		{"landing asset owes its owner's tolls", &domain.Asset{ID: "a1", OwnerID: "alpha", Location: domain.Location{WorldID: "Hadrian"}}, []hooks.ArrivalToll{{FactionID: "bravo", Amount: 1}}},
		{"stealthed asset owes none", &domain.Asset{ID: "a1", OwnerID: "alpha", Stealthy: true, Location: domain.Location{WorldID: "Hadrian"}}, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := AssetArrivalTolls(registry, factions, testCase.asset, "Krylos"); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("tolls = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestTollCharges_PairsALossAndAGainCausedByTheHolder(t *testing.T) {
	got := TollCharges("alpha", []hooks.ArrivalToll{{FactionID: "bravo", Amount: 1}})
	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "alpha", Delta: -1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
		domain.CoinDelta{FactionID: "bravo", Delta: 1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("charges = %+v, want %+v", got, want)
	}
}

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
