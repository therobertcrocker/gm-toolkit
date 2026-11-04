package dice

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
)

func TestAmount_FlatRollsNoDice(t *testing.T) {
	roller := &dicetest.Sequence{}
	if got := Amount(domain.DiceRoll{Modifier: 1}, roller); got != 1 {
		t.Errorf("Amount = %d, want 1", got)
	}
	if roller.Rolled() != 0 {
		t.Errorf("rolled %d dice, want 0", roller.Rolled())
	}
}

func TestAmount_SumsDiceAndModifier(t *testing.T) {
	roller := &dicetest.Sequence{Values: []int{3, 4}}
	roll := domain.DiceRoll{NumDice: 2, Sides: 6, Modifier: 1}
	if got := Amount(roll, roller); got != 8 {
		t.Errorf("Amount = %d, want 8", got)
	}
}

func TestTable_PicksTheRowHoldingTheTotal(t *testing.T) {
	low := domain.AbilityOutcome{From: 2, To: 3, Coin: 1}
	high := domain.AbilityOutcome{From: 4, To: 5, Coin: 2}
	table := domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 6}, Outcomes: []domain.AbilityOutcome{low, high}}
	cases := []struct {
		name    string
		value   int
		want    domain.AbilityOutcome
		matched bool
	}{
		{"below every row", 1, domain.AbilityOutcome{}, false},
		{"on a row's low edge", 2, low, true},
		{"on a row's high edge", 3, low, true},
		{"on the next row's low edge", 4, high, true},
		{"above every row", 6, domain.AbilityOutcome{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, matched := Table(table, dicetest.Constant{Value: tc.value})
			if matched != tc.matched || got != tc.want {
				t.Errorf("Table = %+v, %v, want %+v, %v", got, matched, tc.want, tc.matched)
			}
		})
	}
}

func TestTable_CountsTheModifier(t *testing.T) {
	row := domain.AbilityOutcome{From: 5, To: 5, Coin: 1}
	table := domain.RollTable{Die: domain.DiceRoll{NumDice: 1, Sides: 6, Modifier: 1}, Outcomes: []domain.AbilityOutcome{row}}
	if got, matched := Table(table, dicetest.Constant{Value: 4}); !matched || got != row {
		t.Errorf("Table = %+v, %v, want %+v, true", got, matched, row)
	}
}

func TestRollStep(t *testing.T) {
	die := domain.DiceRoll{NumDice: 1, Sides: 6}
	row := domain.AbilityOutcome{From: 3, To: 6, Coin: 1}
	table := domain.RollTable{Die: die, Outcomes: []domain.AbilityOutcome{row}}

	t.Run("an amount roll gives its total", func(t *testing.T) {
		got := RollStep(domain.AbilityStep{Amount: &die}, dicetest.Constant{Value: 4})
		if got.Amount != 4 || got.Row != nil {
			t.Errorf("RollStep = %+v, want Amount 4 and no row", got)
		}
	})
	t.Run("a table roll gives the matching row", func(t *testing.T) {
		got := RollStep(domain.AbilityStep{Table: &table}, dicetest.Constant{Value: 4})
		if got.Row == nil || *got.Row != row || got.Amount != 0 {
			t.Errorf("RollStep = %+v, want row %+v and no amount", got, row)
		}
	})
	t.Run("a table roll with no match gives no row", func(t *testing.T) {
		roller := &dicetest.Sequence{Values: []int{2}}
		got := RollStep(domain.AbilityStep{Table: &table}, roller)
		if got != (StepResult{}) {
			t.Errorf("RollStep = %+v, want the zero result", got)
		}
		if roller.Rolled() != 1 {
			t.Errorf("rolled %d dice, want 1", roller.Rolled())
		}
	})
	t.Run("a step with no die rolls nothing", func(t *testing.T) {
		roller := &dicetest.Sequence{}
		if got := RollStep(domain.AbilityStep{}, roller); got != (StepResult{}) {
			t.Errorf("RollStep = %+v, want the zero result", got)
		}
		if roller.Rolled() != 0 {
			t.Errorf("rolled %d dice, want 0", roller.Rolled())
		}
	})
}
