package domain_test

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func TestNewFaction(t *testing.T) {
	homeworld := domain.Location{
		WorldID:   "world-a",
		RegionHex: spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 1, R: 2}},
	}
	goal := &domain.Goal{ID: "goal-expand", Name: "Expand Influence"}
	tag := &domain.Tag{ID: "tag-warlike", Name: "Warlike"}

	faction := domain.NewFaction(
		"test-faction", "Test Faction",
		domain.ScaleMinor,
		domain.StatForce, domain.StatCunning, domain.StatWealth,
		[]*domain.Tag{tag},
		goal,
		homeworld,
		nil,
		10,
	)

	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(domain.ScaleMinor)

	if faction.Force != primaryRating {
		t.Errorf("Force = %d, want %d (primary)", faction.Force, primaryRating)
	}
	if faction.Cunning != secondaryRating {
		t.Errorf("Cunning = %d, want %d (secondary)", faction.Cunning, secondaryRating)
	}
	if faction.Wealth != tertiaryRating {
		t.Errorf("Wealth = %d, want %d (tertiary)", faction.Wealth, tertiaryRating)
	}

	expectedMaxHP := domain.CalcMaxHP(faction)
	if faction.MaxHP != expectedMaxHP {
		t.Errorf("MaxHP = %d, want %d", faction.MaxHP, expectedMaxHP)
	}
	if faction.CurrentHP != faction.MaxHP {
		t.Errorf("CurrentHP = %d, want MaxHP (%d)", faction.CurrentHP, faction.MaxHP)
	}

	if faction.ActiveGoal == nil || faction.ActiveGoal.GoalID != goal.ID {
		t.Errorf("ActiveGoal = %v, want GoalID=%q", faction.ActiveGoal, goal.ID)
	}

	homeworldBases := 0
	for _, base := range faction.Bases {
		if base.IsHomeworld {
			homeworldBases++
			if base.Location.WorldID != homeworld.WorldID {
				t.Errorf("homeworld base WorldID = %q, want %q", base.Location.WorldID, homeworld.WorldID)
			}
			if base.CurrentHP != faction.MaxHP {
				t.Errorf("homeworld base CurrentHP = %d, want %d", base.CurrentHP, faction.MaxHP)
			}
		}
	}
	if homeworldBases != 1 {
		t.Errorf("homeworld base count = %d, want 1", homeworldBases)
	}
}

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
