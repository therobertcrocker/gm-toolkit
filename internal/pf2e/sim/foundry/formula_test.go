package foundry

import (
	"errors"
	"testing"
)

func TestEvaluateFormula(t *testing.T) {
	variables := map[string]float64{"@actor.level": 14, "@armor.system.runes.potency": 2}
	for _, test := range []struct {
		text string
		want int
	}{
		{"@actor.level", 14},
		{"2 + @armor.system.runes.potency", 4},
		{"max(1,floor(@actor.level/2))", 7},
		{"ceil(@actor.level/2) + 1", 8},
		{"floor((@actor.level+3)/2)", 8},
		{"(max(floor((@actor.level)/2),1))", 7},
		{"-3 * 2", -6},
	} {
		got, err := evaluateFormula(test.text, variables)
		if err != nil || got != test.want {
			t.Errorf("evaluateFormula(%q) = %d, %v; want %d", test.text, got, err, test.want)
		}
	}
	for _, text := range []string{
		"@actor.abilities.int.mod",
		"match(when(btwn(@actor.level, 1, 8), 1))",
		"7/2",
		"2 +",
		"1/0",
		"3 $ 4",
	} {
		if _, err := evaluateFormula(text, variables); !errors.Is(err, errUnsupportedFormula) {
			t.Errorf("evaluateFormula(%q) err = %v, want errUnsupportedFormula", text, err)
		}
	}
}
