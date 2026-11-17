package foundry

import (
	"errors"
	"testing"
)

func TestRollOptionsHold(t *testing.T) {
	options := rollOptions{
		flags: map[string]bool{"feat:toughness": true, "armor:category:heavy": true, "armor:group:plate": true},
	}
	for _, test := range []struct {
		name      string
		predicate []any
		want      bool
	}{
		{"empty", nil, true},
		{"all flags", []any{"armor:category:heavy", "armor:group:plate"}, true},
		{"missing flag", []any{"armor:category:heavy", "armor:group:composite"}, false},
		{"not", []any{map[string]any{"not": "feat:expanded-splash"}}, true},
		{"or", []any{map[string]any{"or": []any{"target:hidden", "feat:toughness"}}}, true},
		{"and", []any{map[string]any{"and": []any{"target:hidden", "feat:toughness"}}}, false},
		{"nor", []any{map[string]any{"nor": []any{"target:hidden", "target:undetected"}}}, true},
	} {
		got, err := options.holds(test.predicate)
		if err != nil || got != test.want {
			t.Errorf("%s: holds = %v, %v; want %v", test.name, got, err, test.want)
		}
	}
	for _, test := range []struct {
		name      string
		predicate []any
	}{
		{"xor", []any{map[string]any{"xor": []any{}}}},
		{"comparison", []any{map[string]any{"gte": []any{"self:level", 10.0}}}},
		{"not comparison", []any{map[string]any{"not": map[string]any{"gte": []any{"self:level", 10.0}}}}},
	} {
		if held, err := options.holds(test.predicate); held || !errors.Is(err, errUnsupportedPredicate) {
			t.Errorf("%s: holds = %v, %v; want false, errUnsupportedPredicate", test.name, held, err)
		}
	}
}

func TestResolveTemplates(t *testing.T) {
	legend := &item{ID: "abc123", System: itemCommon{Rules: []map[string]any{
		{"key": "ChoiceSet", "flag": "weaponLegend", "selection": "hammer"},
	}}}
	for _, test := range []struct{ text, want string }{
		{"item:group:{item|flags.pf2e.rulesSelections.weaponLegend}", "item:group:hammer"},
		{"{item|_id}-attack", "abc123-attack"},
		{"no template", "no template"},
	} {
		got, err := resolveTemplates(test.text, legend)
		if err != nil || got != test.want {
			t.Errorf("resolveTemplates(%q) = %q, %v; want %q", test.text, got, err, test.want)
		}
	}
	if _, err := resolveTemplates("{actor|flags.pf2e.x}", legend); !errors.Is(err, errUnsupportedTemplate) {
		t.Errorf("actor template err = %v, want errUnsupportedTemplate", err)
	}
}
