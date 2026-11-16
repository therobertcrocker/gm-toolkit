package dice_test

import (
	"errors"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type fixedRoller int

func (value fixedRoller) Die(int) int { return int(value) }

func TestParse(t *testing.T) {
	valid := []struct {
		text string
		want dice.Expression
	}{
		{"3d6+14", dice.Expression{Count: 3, Sides: 6, Modifier: 14}},
		{"1d8", dice.Expression{Count: 1, Sides: 8}},
		{"2d4-1", dice.Expression{Count: 2, Sides: 4, Modifier: -1}},
		{"5", dice.Expression{Modifier: 5}},
		{" 1d20 ", dice.Expression{Count: 1, Sides: 20}},
	}
	for _, testCase := range valid {
		got, err := dice.Parse(testCase.text)
		if err != nil {
			t.Errorf("Parse(%q): %v", testCase.text, err)
			continue
		}
		if got != testCase.want {
			t.Errorf("Parse(%q) = %+v, want %+v", testCase.text, got, testCase.want)
		}
	}
	for _, text := range []string{"", "d6", "0d6", "1d1", "1d", "1d6+", "two", "1d6+x"} {
		if _, err := dice.Parse(text); !errors.Is(err, dice.ErrBadExpression) {
			t.Errorf("Parse(%q) err = %v, want ErrBadExpression", text, err)
		}
	}
}

func TestExpressionStringRoundTrips(t *testing.T) {
	for _, text := range []string{"3d6+14", "1d8", "2d4-1", "5"} {
		expression, err := dice.Parse(text)
		if err != nil {
			t.Fatalf("Parse(%q): %v", text, err)
		}
		if got := expression.String(); got != text {
			t.Errorf("String() = %q, want %q", got, text)
		}
	}
}

func TestRollAddsDiceAndModifier(t *testing.T) {
	expression := dice.Expression{Count: 3, Sides: 6, Modifier: 2}
	if got := expression.Roll(fixedRoller(4)); got != 14 {
		t.Errorf("Roll = %d, want 14", got)
	}
}

func TestSeededRepeatsForTheSameSeed(t *testing.T) {
	first, second, other := dice.NewSeeded(42), dice.NewSeeded(42), dice.NewSeeded(43)
	differs := false
	for range 200 {
		value := first.Die(20)
		if value < 1 || value > 20 {
			t.Fatalf("Die(20) = %d, out of range", value)
		}
		if again := second.Die(20); again != value {
			t.Fatalf("same seed gave %d then %d", value, again)
		}
		if other.Die(20) != value {
			differs = true
		}
	}
	if !differs {
		t.Error("seeds 42 and 43 gave the same 200 rolls")
	}
}
