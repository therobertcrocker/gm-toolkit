package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)

func TestBlockAbility_PutsNoAbilityOnTheAttackedAsset(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha"}
	beta := &domain.Faction{ID: "beta"}
	saboteurs := activeAsset("saboteurs", "SWN-C2-002")
	got := effects.Registrars()[domain.SpecialBlockAbility].Matchup(effect.MatchupContext{
		Holder:        effect.Holder{Faction: alpha, Asset: saboteurs},
		Asset:         saboteurs,
		Opponent:      beta,
		OpponentAsset: activeAsset("informers", "SWN-C1-002"),
	})

	want := effect.MatchupOutcome{Mutations: []domain.Mutation{domain.AssetConditionApplied{
		FactionID: "beta",
		AssetID:   "informers",
		Condition: domain.Condition{
			Kind:            domain.ConditionNoAbility,
			SourceAssetID:   "saboteurs",
			ExpiresAtTurnOf: "alpha",
		},
		Cause:             "block_ability",
		CausedByFactionID: "alpha",
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("outcome = %+v, want %+v", got, want)
	}
}
