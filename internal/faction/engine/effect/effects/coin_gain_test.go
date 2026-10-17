package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
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
				Faction: faction,
				Entry:   entry,
				Roller:  fixedRoller{value: 3},
			})
			want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: tc.want, Cause: "coin_gain", CausedByFactionID: "alpha"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Fire = %+v, want %+v", got, want)
			}
		})
	}
}
