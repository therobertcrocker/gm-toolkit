package rulebook

import (
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
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

	t.Run("nil ability for non-action asset", func(t *testing.T) {
		asset, ok := rb.Assets["F1-001"]
		if !ok {
			t.Fatal("asset F1-001 not found")
		}
		if asset.Ability != nil {
			t.Errorf("expected nil ability, got %+v", asset.Ability)
		}
	})

	t.Run("movement ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["F2-001"]
		if !ok {
			t.Fatal("asset F2-001 not found")
		}
		if asset.Ability == nil {
			t.Fatal("expected ability, got nil")
		}
		if len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %d, want 1", len(asset.Ability.Steps))
		}
		step := asset.Ability.Steps[0]
		if step.Type != domain.AbilityStepMovement {
			t.Errorf("step type: got %q, want %q", step.Type, domain.AbilityStepMovement)
		}
		if step.MaxHex != 1 {
			t.Errorf("max_hex: got %d, want 1", step.MaxHex)
		}
		if step.CoinCost != 1 {
			t.Errorf("coin_cost: got %d, want 1", step.CoinCost)
		}
	})

	t.Run("faction_test ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["C1-002"]
		if !ok {
			t.Fatal("asset C1-002 not found")
		}
		if asset.Ability == nil {
			t.Fatal("expected ability, got nil")
		}
		if len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %d, want 1", len(asset.Ability.Steps))
		}
		step := asset.Ability.Steps[0]
		if step.Type != domain.AbilityStepFactionTest {
			t.Errorf("step type: got %q, want %q", step.Type, domain.AbilityStepFactionTest)
		}
		if step.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectRevealStealth)
		}
	})

	t.Run("combo ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["C2-004"]
		if !ok {
			t.Fatal("asset C2-004 not found")
		}
		if asset.Ability == nil {
			t.Fatal("expected ability, got nil")
		}
		if len(asset.Ability.Steps) != 2 {
			t.Fatalf("steps: got %d, want 2", len(asset.Ability.Steps))
		}
		if asset.Ability.Steps[0].Type != domain.AbilityStepMovement {
			t.Errorf("step 0 type: got %q, want movement", asset.Ability.Steps[0].Type)
		}
		if asset.Ability.Steps[1].Type != domain.AbilityStepFactionTest {
			t.Errorf("step 1 type: got %q, want faction_test", asset.Ability.Steps[1].Type)
		}
	})

	t.Run("effect_dice parsed for coin_drain", func(t *testing.T) {
		asset, ok := rb.Assets["W5-001"]
		if !ok {
			t.Fatal("asset W5-001 not found")
		}
		if asset.Ability == nil {
			t.Fatal("expected ability, got nil")
		}
		testStep := asset.Ability.Steps[1]
		if testStep.Effect != domain.EffectCoinDrain {
			t.Errorf("effect: got %q, want coin_drain", testStep.Effect)
		}
		if testStep.EffectDice == nil {
			t.Fatal("expected effect_dice, got nil")
		}
		if testStep.EffectDice.NumDice != 1 || testStep.EffectDice.Sides != 4 {
			t.Errorf("effect_dice: got %+v, want 1d4", testStep.EffectDice)
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

func TestLoadDuplicateAssetID(t *testing.T) {
	_, err := Load("testdata/duplicate_ids")
	if err == nil {
		t.Fatal("expected error for duplicate asset ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate asset ID") {
		t.Errorf("error message %q does not mention duplicate asset ID", err.Error())
	}
}
