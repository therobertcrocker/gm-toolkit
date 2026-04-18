package loader

import (
	"testing"
)

const testDataDir = "../data"

func TestLoad(t *testing.T) {
	rb, err := Load(testDataDir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	t.Run("assets populated", func(t *testing.T) {
		if len(rb.Assets) == 0 {
			t.Fatal("expected assets to be loaded, got none")
		}
	})

	t.Run("asset fields correct", func(t *testing.T) {
		asset, ok := rb.Assets["F1-001"]
		if !ok {
			t.Fatal("expected asset F1-001 (Security Personnel) to be present")
		}
		if asset.Name != "Security Personnel" {
			t.Errorf("name: got %q, want %q", asset.Name, "Security Personnel")
		}
		if asset.HP != 3 {
			t.Errorf("hp: got %d, want 3", asset.HP)
		}
		if asset.MinRating != 1 {
			t.Errorf("min_rating: got %d, want 1", asset.MinRating)
		}
	})

	t.Run("attack profile parsed", func(t *testing.T) {
		asset, ok := rb.Assets["F1-001"]
		if !ok {
			t.Fatal("asset F1-001 not found")
		}
		if asset.Attack == nil {
			t.Fatal("expected attack profile, got nil")
		}
		d := asset.Attack.Damage
		if d.NumDice != 1 || d.Sides != 3 || d.Modifier != 1 {
			t.Errorf("damage: got %+v, want 1d3+1", d)
		}
	})

	t.Run("counter parsed", func(t *testing.T) {
		asset, ok := rb.Assets["F1-001"]
		if !ok {
			t.Fatal("asset F1-001 not found")
		}
		if asset.Counter == nil {
			t.Fatal("expected counter, got nil")
		}
		if asset.Counter.NumDice != 1 || asset.Counter.Sides != 4 {
			t.Errorf("counter: got %+v, want 1d4", asset.Counter)
		}
	})

	t.Run("nil attack and counter for non-combat asset", func(t *testing.T) {
		asset, ok := rb.Assets["F2-001"]
		if !ok {
			t.Fatal("expected asset F2-001 (Heavy Drop Assets) to be present")
		}
		if asset.Attack != nil {
			t.Errorf("expected nil attack, got %+v", asset.Attack)
		}
		if asset.Counter != nil {
			t.Errorf("expected nil counter, got %+v", asset.Counter)
		}
	})

	t.Run("tags populated", func(t *testing.T) {
		if len(rb.Tags) == 0 {
			t.Fatal("expected tags to be loaded, got none")
		}
		tag, ok := rb.Tags["T-001"]
		if !ok {
			t.Fatal("expected tag T-001 (Colonists) to be present")
		}
		if tag.Name != "Colonists" {
			t.Errorf("tag name: got %q, want %q", tag.Name, "Colonists")
		}
	})

	t.Run("goals populated", func(t *testing.T) {
		if len(rb.Goals) == 0 {
			t.Fatal("expected goals to be loaded, got none")
		}
		goal, ok := rb.Goals["G-001"]
		if !ok {
			t.Fatal("expected goal G-001 (Military Conquest) to be present")
		}
		if goal.Name != "Military Conquest" {
			t.Errorf("goal name: got %q, want %q", goal.Name, "Military Conquest")
		}
		if goal.Difficulty != "half_assets_destroyed" {
			t.Errorf("difficulty: got %q, want %q", goal.Difficulty, "half_assets_destroyed")
		}
	})
}
