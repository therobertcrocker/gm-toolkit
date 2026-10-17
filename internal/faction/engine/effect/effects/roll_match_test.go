package effects

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func TestRollMatches(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha"}
	beta := &domain.Faction{ID: "beta"}
	gamma := &domain.Faction{ID: "gamma"}
	declaring := &domain.Asset{ID: "x", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	other := &domain.Asset{ID: "y", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}

	self := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf}
	world := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld}
	rivalSelf := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf}
	rivalWorld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld}
	cunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf, Attribute: domain.StatCunning}
	attackOrDefense := domain.EffectFilters{
		Owner: domain.OwnerOwn, Scope: domain.ScopeSelf,
		Phases: []domain.EffectRollPhase{domain.RollPhaseAttack, domain.RollPhaseDefense},
	}
	expansion := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, Action: "Expand Influence"}

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
		{"phase in list", domain.TriggerOwnRoll, attackOrDefense, hooks.RollContext{Actor: alpha, Asset: declaring, Phase: hooks.PhaseDefense}, true},
		{"phase not in list", domain.TriggerOwnRoll, attackOrDefense, hooks.RollContext{Actor: alpha, Asset: declaring, Phase: hooks.PhaseContested}, false},
		{"action match", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Expand Influence"}, true},
		{"action mismatch", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Attack"}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rollMatches(tc.ctx, tc.trigger, "alpha", declaring, tc.filters); got != tc.want {
				t.Errorf("rollMatches = %v, want %v", got, tc.want)
			}
		})
	}
}
