package domain_test

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestFaction_HookBudgets_TOMLRoundTrip(t *testing.T) {
	t.Run("nil budgets omitted and round-trip to nil", func(t *testing.T) {
		original := &domain.Faction{ID: "f1", HookBudgets: nil}
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(original); err != nil {
			t.Fatalf("encode: %v", err)
		}
		encoded := buf.String()
		if bytes.Contains([]byte(encoded), []byte("hook_budgets")) {
			t.Errorf("expected hook_budgets to be omitted when nil, got:\n%s", encoded)
		}
		var decoded domain.Faction
		if _, err := toml.Decode(encoded, &decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.HookBudgets != nil {
			t.Errorf("HookBudgets = %v, want nil after round-trip of nil", decoded.HookBudgets)
		}
	})

	t.Run("populated budgets preserved after round-trip", func(t *testing.T) {
		original := &domain.Faction{
			ID: "f1",
			HookBudgets: map[string]int{
				"tag:Warlike":   1,
				"tag:Fanatical": 0,
			},
		}
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(original); err != nil {
			t.Fatalf("encode: %v", err)
		}
		var decoded domain.Faction
		if _, err := toml.Decode(buf.String(), &decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(decoded.HookBudgets) != len(original.HookBudgets) {
			t.Fatalf("HookBudgets len = %d, want %d", len(decoded.HookBudgets), len(original.HookBudgets))
		}
		for key, want := range original.HookBudgets {
			if got := decoded.HookBudgets[key]; got != want {
				t.Errorf("HookBudgets[%q] = %d, want %d", key, got, want)
			}
		}
	})
}
