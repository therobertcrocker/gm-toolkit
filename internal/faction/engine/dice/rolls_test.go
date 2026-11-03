package dice

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type scriptedRoller struct {
	values []int
	next   int
}

func (roller *scriptedRoller) Roll(int) int {
	value := roller.values[roller.next]
	roller.next++
	return value
}

func TestAmount_FlatRollsNoDice(t *testing.T) {
	roller := &scriptedRoller{}
	if got := Amount(domain.DiceRoll{Modifier: 1}, roller); got != 1 {
		t.Errorf("Amount = %d, want 1", got)
	}
	if roller.next != 0 {
		t.Errorf("rolled %d dice, want 0", roller.next)
	}
}

func TestAmount_SumsDiceAndModifier(t *testing.T) {
	roller := &scriptedRoller{values: []int{3, 4}}
	roll := domain.DiceRoll{NumDice: 2, Sides: 6, Modifier: 1}
	if got := Amount(roll, roller); got != 8 {
		t.Errorf("Amount = %d, want 8", got)
	}
}
