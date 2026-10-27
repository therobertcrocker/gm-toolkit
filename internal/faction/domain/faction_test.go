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

	faction := domain.NewFaction(
		"test-faction", "Test Faction",
		domain.ScaleMinor,
		domain.StatForce, domain.StatCunning, domain.StatWealth,
		[]string{"SWN-T-020"},
		goal,
		homeworld,
		nil,
		10,
	)

	if !faction.HasTag("SWN-T-020") || len(faction.Tags) != 1 {
		t.Errorf("Tags = %v, want [SWN-T-020]", faction.Tags)
	}

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

func TestFaction_SpendBudget(t *testing.T) {
	t.Run("a spent key reads as spent and leaves other keys unspent", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1"}
		faction.SpendBudget("k")
		if !faction.BudgetSpent("k") {
			t.Error("BudgetSpent(k) = false after spend, want true")
		}
		if faction.BudgetSpent("other") {
			t.Error("BudgetSpent(other) = true, want false")
		}
	})

	t.Run("an empty key is unlimited and creates no map", func(t *testing.T) {
		faction := &domain.Faction{ID: "f1"}
		faction.SpendBudget("")
		if faction.BudgetSpent("") {
			t.Error("BudgetSpent(\"\") = true, want false")
		}
		if faction.HookBudgets != nil {
			t.Errorf("HookBudgets = %v, want nil", faction.HookBudgets)
		}
	})
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

func TestAttackBlocked(t *testing.T) {
	// f1's Seditionists s1 and f2's b2 share Anchorage; b2 holds s1's no_attack.
	makeFactions := func() (map[string]*domain.Faction, *domain.Asset) {
		source := &domain.Asset{ID: "s1", OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}, Ready: true, Maintained: true, CurrentHP: 8}
		target := &domain.Asset{
			ID:         "b2",
			OwnerID:    "f2",
			Location:   domain.Location{WorldID: "Anchorage"},
			Conditions: []domain.Condition{{Kind: domain.ConditionNoAttack, SourceAssetID: "s1"}},
		}
		return map[string]*domain.Faction{
			"f1": {ID: "f1", Assets: map[string]*domain.Asset{"s1": source}},
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b2": target}},
		}, target
	}
	cases := []struct {
		name  string
		setup func(factions map[string]*domain.Faction, target *domain.Asset)
		want  bool
	}{
		{
			name:  "the source shares the world and has another owner",
			setup: func(map[string]*domain.Faction, *domain.Asset) {},
			want:  true,
		},
		{
			name:  "the source is gone",
			setup: func(factions map[string]*domain.Faction, _ *domain.Asset) { delete(factions["f1"].Assets, "s1") },
		},
		{
			name: "the source is on another world",
			setup: func(factions map[string]*domain.Faction, _ *domain.Asset) {
				factions["f1"].Assets["s1"].Location.WorldID = "Hadrian"
			},
		},
		{
			name: "both are in transit (no world)",
			setup: func(factions map[string]*domain.Faction, target *domain.Asset) {
				factions["f1"].Assets["s1"].Location.WorldID = ""
				target.Location.WorldID = ""
			},
		},
		{
			name: "the source is captured into the target's faction",
			setup: func(factions map[string]*domain.Faction, _ *domain.Asset) {
				source := factions["f1"].Assets["s1"]
				delete(factions["f1"].Assets, "s1")
				source.OwnerID = "f2"
				factions["f2"].Assets["s1"] = source
			},
		},
		{
			name: "a disabled, unready source still holds the block",
			setup: func(factions map[string]*domain.Faction, _ *domain.Asset) {
				source := factions["f1"].Assets["s1"]
				source.Ready = false
				source.Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m9"}}
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factions, target := makeFactions()
			tc.setup(factions, target)
			if got := domain.AttackBlocked(factions, target); got != tc.want {
				t.Errorf("AttackBlocked = %v, want %v", got, tc.want)
			}
		})
	}
}
