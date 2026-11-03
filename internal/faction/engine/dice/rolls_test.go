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
