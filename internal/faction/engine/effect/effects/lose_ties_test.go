package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestLoseTies_OwnerLosesOnEitherSide(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-005"}}
	beta := &domain.Faction{ID: "beta"}
	gamma := &domain.Faction{ID: "gamma"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta, "gamma": gamma}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialLoseTies].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-005", Name: "Fanatical"}},
		Entry: domain.EffectDefinition{Trigger: domain.TriggerAttackTie, Effect: domain.SpecialLoseTies,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		Registry: registry,
	})

	cases := []struct {
		name string
		ctx  hooks.RollContext
		want hooks.TieOutcome
	}{
		{"owner attacks", hooks.RollContext{Actor: alpha, Opponent: beta}, hooks.TieDefenderWins},
		{"owner defends", hooks.RollContext{Actor: beta, Opponent: alpha}, hooks.TieAttackerWins},
		{"owner turns an asset on itself", hooks.RollContext{Actor: alpha}, hooks.TieDefenderWins},
		{"others fight", hooks.RollContext{Actor: beta, Opponent: gamma}, hooks.TieStandard},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dispatch.ResolveTie(registry, tc.ctx, factionState); got != tc.want {
				t.Errorf("ResolveTie = %v, want %v", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := dispatch.ResolveTie(registry, hooks.RollContext{Actor: alpha, Opponent: beta}, factionState); got != hooks.TieStandard {
		t.Errorf("dropped tag: ResolveTie = %v, want TieStandard", got)
	}
}
