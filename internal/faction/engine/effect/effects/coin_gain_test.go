package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type fixedRoller struct{ value int }

func (roller fixedRoller) Roll(int) int { return roller.value }

func TestCoinGain_PaysTheOwnerTheAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount domain.DiceRoll
		want   int
	}{
		{"flat amount", domain.DiceRoll{Modifier: 1}, 1},
		{"rolled amount", domain.DiceRoll{NumDice: 1, Sides: 4}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "alpha"}
			entry := domain.EffectDefinition{
				Trigger:  domain.TriggerTurnStart,
				Effect:   domain.SpecialCoinGain,
				Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
				CoinGain: &domain.CoinGainParams{Amount: tc.amount},
			}
			got := effects.Registrars()[domain.SpecialCoinGain].Fire(effect.FireContext{
				Holder: effect.Holder{Faction: faction},
				Entry:  entry,
				Roller: fixedRoller{value: 3},
			})
			want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: tc.want, Cause: "coin_gain", CausedByFactionID: "alpha"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Fire = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCoinGain_ScavengersPaysOwnKillsAndLossesOnly(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-016"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	holder := effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-016", Name: "Scavengers"}}
	for _, filters := range []domain.EffectFilters{
		{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		{Owner: domain.OwnerRival, Scope: domain.ScopeFaction, CausedBy: domain.OwnerOwn},
	} {
		effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
			Holder: holder, Registry: registry, Roller: fixedRoller{value: 1},
			Entry: domain.EffectDefinition{Trigger: domain.TriggerAssetDestroyed, Effect: domain.SpecialCoinGain,
				Filters: filters, CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}}},
		})
	}
	paid := func(removed domain.AssetRemoved) int {
		emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], []domain.Mutation{removed}, factionState, nil)
		if err != nil {
			t.Fatalf("MutationReactors: %v", err)
		}
		total := 0
		for _, mutation := range emitted {
			if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" {
				total += gain.Delta
			}
		}
		return total
	}

	cases := []struct {
		name    string
		removed domain.AssetRemoved
		want    int
	}{
		{"own loss in an attack", domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}, 1},
		{"own kill in an attack", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "alpha"}, 1},
		{"own kill by Cracked Comms", domain.AssetRemoved{FactionID: "beta", Cause: "counterstrike", CausedByFactionID: "alpha"}, 1},
		{"a fight between two others (B.042)", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"}, 0},
		{"a sale", domain.AssetRemoved{FactionID: "alpha", Cause: "sell", CausedByFactionID: "alpha"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := paid(tc.removed); got != tc.want {
				t.Errorf("paid %d, want %d", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := paid(domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}); got != 0 {
		t.Errorf("dropped tag: paid %d, want 0", got)
	}
}
