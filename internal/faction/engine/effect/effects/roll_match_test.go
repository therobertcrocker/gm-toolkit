package effects

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func TestRollMatches(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Homeworld: domain.Location{WorldID: "Krylos"}}
	beta := &domain.Faction{ID: "beta"}
	gamma := &domain.Faction{ID: "gamma"}
	declaring := &domain.Asset{ID: "x", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	other := &domain.Asset{ID: "y", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	primitive := &domain.Asset{ID: "p", OwnerID: "alpha", DefinitionID: "tl0"}
	advanced := &domain.Asset{ID: "q", OwnerID: "beta", DefinitionID: "tl5"}
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"tl0": {ID: "tl0", TechLevel: 0},
		"tl5": {ID: "tl5", TechLevel: 5},
	}}
	zero, five := 0, 5

	self := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf}
	world := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld}
	rivalSelf := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf}
	rivalWorld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld}
	cunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf, Attribute: domain.StatCunning}
	attackOrDefense := domain.EffectFilters{
		Owner: domain.OwnerOwn, Scope: domain.ScopeSelf,
		Phases: []domain.RollPhase{domain.RollPhaseAttack, domain.RollPhaseDefense},
	}
	expansion := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, Action: "Expand Influence"}
	homeworld := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld}
	againstCunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentAttribute: domain.StatCunning}
	withPrimitive := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, TechLevel: &zero}
	againstAdvanced := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentTechLevel: &five}
	rivalHomeworld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeHomeworld}
	slavesRoll := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, AssetDefinition: "tl0"}

	cases := []struct {
		name    string
		trigger domain.EffectTrigger
		filters domain.EffectFilters
		ctx     hooks.RollContext
		want    bool
	}{
		{"own roll, self", domain.TriggerOwnRoll, self, hooks.RollContext{Actor: alpha, Asset: declaring}, true},
		{"own roll, other asset rolls", domain.TriggerOwnRoll, self, hooks.RollContext{Actor: alpha, Asset: other}, false},
		{"own roll, no rolling asset", domain.TriggerOwnRoll, self, hooks.RollContext{Actor: alpha}, false},
		{"own roll, rival rolling", domain.TriggerOwnRoll, world, hooks.RollContext{Actor: beta, World: "Krylos"}, false},
		{"own roll, world", domain.TriggerOwnRoll, world, hooks.RollContext{Actor: alpha, World: "Krylos"}, true},
		{"own roll, other world", domain.TriggerOwnRoll, world, hooks.RollContext{Actor: alpha, World: "Tartarus"}, false},
		{"own roll, in flight", domain.TriggerOwnRoll, world, hooks.RollContext{Actor: alpha, World: ""}, false},
		{"no actor", domain.TriggerOwnRoll, world, hooks.RollContext{World: "Krylos"}, false},
		{"enemy at me, self", domain.TriggerEnemyRollAtMe, rivalSelf, hooks.RollContext{Actor: beta, Opponent: alpha, OpponentAsset: declaring}, true},
		{"enemy at me, against other asset", domain.TriggerEnemyRollAtMe, rivalSelf, hooks.RollContext{Actor: beta, Opponent: alpha, OpponentAsset: other}, false},
		{"enemy at me, world", domain.TriggerEnemyRollAtMe, rivalWorld, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, true},
		{"enemy, not against owner", domain.TriggerEnemyRollAtMe, rivalWorld, hooks.RollContext{Actor: beta, Opponent: gamma, World: "Krylos"}, false},
		{"enemy, no opponent", domain.TriggerEnemyRollAtMe, rivalWorld, hooks.RollContext{Actor: beta, World: "Krylos"}, false},
		{"owner's own roll is not enemy", domain.TriggerEnemyRollAtMe, rivalWorld, hooks.RollContext{Actor: alpha, Opponent: beta, World: "Krylos"}, false},
		{"attribute match", domain.TriggerOwnRoll, cunning, hooks.RollContext{Actor: alpha, Asset: declaring, Attribute: "Cunning"}, true},
		{"attribute mismatch", domain.TriggerOwnRoll, cunning, hooks.RollContext{Actor: alpha, Asset: declaring, Attribute: "Force"}, false},
		{"phase in list", domain.TriggerOwnRoll, attackOrDefense, hooks.RollContext{Actor: alpha, Asset: declaring, Phase: domain.RollPhaseDefense}, true},
		{"phase not in list", domain.TriggerOwnRoll, attackOrDefense, hooks.RollContext{Actor: alpha, Asset: declaring, Phase: domain.RollPhaseContested}, false},
		{"action match", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Expand Influence"}, true},
		{"action mismatch", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Attack"}, false},
		{"own roll, homeworld", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: "Krylos"}, true},
		{"own roll, off homeworld", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: "Tartarus"}, false},
		{"own roll, homeworld scope in flight", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: ""}, false},
		{"enemy at me, the owner's homeworld", domain.TriggerEnemyRollAtMe, rivalHomeworld, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, true},
		{"opponent attribute match", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Wealth", OpponentAttribute: "Cunning"}, true},
		{"opponent attribute mismatch", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Cunning", OpponentAttribute: "Wealth"}, false},
		{"tech level match", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha, Asset: primitive}, true},
		{"tech level, asset with no definition", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
		{"tech level, no rolling asset", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha}, false},
		{"opponent tech level match", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: advanced}, true},
		{"opponent tech level mismatch", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: primitive}, false},
		{"detection, a detecting roll", domain.TriggerOwnRoll, domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Detection: true}, hooks.RollContext{Actor: alpha, Detection: true}, true},
		{"detection, any other roll", domain.TriggerOwnRoll, domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Detection: true}, hooks.RollContext{Actor: alpha}, false},
		{"rival roll, a bystander's roll", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: beta, Opponent: gamma, World: "Tartarus"}, true},
		{"rival roll, against the owner", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: beta, Opponent: alpha}, true},
		{"rival roll, the owner's own roll", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: alpha, Opponent: beta}, false},
		{"asset definition match", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha, Asset: primitive}, true},
		{"asset definition mismatch", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha, Asset: advanced}, false},
		{"asset definition, no rolling asset", domain.TriggerOwnRoll, slavesRoll, hooks.RollContext{Actor: alpha}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rollMatches(tc.ctx, tc.trigger, "alpha", declaring, tc.filters, book); got != tc.want {
				t.Errorf("rollMatches = %v, want %v", got, tc.want)
			}
		})
	}

	faction := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}
	if !rollMatches(hooks.RollContext{Actor: alpha, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction, nil) {
		t.Error("faction scope: the owner's roll on any world, with no asset, does not match")
	}
	if rollMatches(hooks.RollContext{Actor: beta, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction, nil) {
		t.Error("faction scope: a rival's roll matches")
	}
	if rollMatches(hooks.RollContext{Actor: alpha, Asset: primitive}, domain.TriggerOwnRoll, "alpha", nil, withPrimitive, nil) {
		t.Error("a tech-level filter with no rulebook matches")
	}
}
