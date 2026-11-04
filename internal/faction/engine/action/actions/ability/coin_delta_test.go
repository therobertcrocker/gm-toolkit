package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
)

func coinDeltaStep(die domain.DiceRoll, outcomes []domain.AbilityOutcome) domain.AbilityStep {
	return domain.AbilityStep{
		Effect: domain.EffectCoinDelta,
		Target: domain.TargetNone,
		Table:  &domain.RollTable{Die: die, Outcomes: outcomes},
	}
}

func TestCoinDelta_Payouts(t *testing.T) {
	postech := []domain.AbilityOutcome{
		{From: 1, To: 1, Coin: -1},
		{From: 2, To: 4, Coin: 1},
		{From: 5, To: 6, Coin: 2},
	}
	harvesters := []domain.AbilityOutcome{{From: 3, To: 6, Coin: 1}}

	tests := []struct {
		name        string
		outcomes    []domain.AbilityOutcome
		row         *domain.AbilityOutcome // the row the die picked; nil for no match
		factionCoin int
		wantCoin    int  // expected CoinDelta.Delta; 0 => no CoinDelta
		wantRemove  bool // expected AssetRemoved
	}{
		{"harvesters miss", harvesters, nil, 0, 0, false},
		{"harvesters hit", harvesters, &harvesters[0], 0, 1, false},
		{"postech gain", postech, &postech[2], 0, 2, false},
		{"postech loss affordable", postech, &postech[0], 3, -1, false},
		{"postech loss insolvent", postech, &postech[0], 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := StepContext{
				Faction: &domain.Faction{ID: "f1", Coin: tt.factionCoin},
				Asset:   &domain.Asset{ID: "a1", OwnerID: "f1"},
				Step:    coinDeltaStep(domain.DiceRoll{NumDice: 1, Sides: 6}, tt.outcomes),
				Rolled:  dice.StepResult{Row: tt.row},
			}
			muts, err := coinDelta(ctx)
			if err != nil {
				t.Fatalf("coinDelta: %v", err)
			}
			assertPayout(t, muts, tt.wantCoin, tt.wantRemove)
		})
	}
}

func TestCoinDelta_DestroySelfUnconditional(t *testing.T) {
	outcomes := []domain.AbilityOutcome{{From: 1, To: 1, DestroySelf: true}}
	ctx := StepContext{
		Faction: &domain.Faction{ID: "f1", Coin: 99}, // affordable, still destroyed
		Asset:   &domain.Asset{ID: "a1", OwnerID: "f1"},
		Step:    coinDeltaStep(domain.DiceRoll{NumDice: 1, Sides: 8}, outcomes),
		Rolled:  dice.StepResult{Row: &outcomes[0]},
	}
	muts, err := coinDelta(ctx)
	if err != nil {
		t.Fatalf("coinDelta: %v", err)
	}
	assertPayout(t, muts, 0, true)
}

func assertPayout(t *testing.T, muts []domain.Mutation, wantCoin int, wantRemove bool) {
	t.Helper()
	var gotCoin int
	var gotRemove bool
	for _, m := range muts {
		switch v := m.(type) {
		case domain.CoinDelta:
			gotCoin = v.Delta
		case domain.AssetRemoved:
			gotRemove = true
		default:
			t.Errorf("unexpected mutation %T", m)
		}
	}
	if gotCoin != wantCoin {
		t.Errorf("CoinDelta.Delta = %d, want %d", gotCoin, wantCoin)
	}
	if gotRemove != wantRemove {
		t.Errorf("AssetRemoved present = %v, want %v", gotRemove, wantRemove)
	}
}
