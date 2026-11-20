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

func TestEvaluateMatchWhen(t *testing.T) {
	const improvisation = "match(when(lte(@actor.level, 6), floor(@actor.level / 2)), when(gte(@actor.level, 7), @actor.level))"
	for _, test := range []struct {
		level float64
		text  string
		want  int
	}{
		{14, improvisation, 14},
		{7, improvisation, 7},
		{6, improvisation, 3},
		{5, improvisation, 2},
		{1, improvisation, 0},
		{3, "match(when(gte(@actor.level, 3), 10), when(gte(@actor.level, 1), 20))", 10},
		{3, "when(lte(@actor.level, 3), 4) + 1", 5},
	} {
		got, err := evaluateFormula(test.text, map[string]float64{"@actor.level": test.level})
		if err != nil || got != test.want {
			t.Errorf("level %v: evaluateFormula(%q) = %d, %v; want %d", test.level, test.text, got, err, test.want)
		}
	}
	for _, text := range []string{
		"match(when(gte(@actor.level, 20), 1))",
		"when(gte(@actor.level, 20), 1)",
		"match(when(gte(@actor.level, 20), 1) + 1, when(gte(@actor.level, 1), 2))",
		"lte(@actor.level)",
		"match(when(btwn(@actor.level, 1, 20), 1), when(gte(@actor.level, 1), 2))",
	} {
		if _, err := evaluateFormula(text, map[string]float64{"@actor.level": 14}); !errors.Is(err, errUnsupportedFormula) {
			t.Errorf("evaluateFormula(%q) err = %v, want errUnsupportedFormula", text, err)
		}
	}
}
