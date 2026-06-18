package rulebook

import (
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const testDataDir = "../../../rulebooks/swn"

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
		asset, ok := rb.Assets["SWN-F1-001"]
		if !ok {
			t.Fatal("expected asset SWN-F1-001 (Security Personnel) to be present")
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
		asset, ok := rb.Assets["SWN-F1-001"]
		if !ok {
			t.Fatal("asset SWN-F1-001 not found")
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
		asset, ok := rb.Assets["SWN-F1-001"]
		if !ok {
			t.Fatal("asset SWN-F1-001 not found")
		}
		if asset.Counter == nil {
			t.Fatal("expected counter, got nil")
		}
		if asset.Counter.NumDice != 1 || asset.Counter.Sides != 4 {
			t.Errorf("counter: got %+v, want 1d4", asset.Counter)
		}
	})

	t.Run("nil attack and counter for non-combat asset", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-F2-001"]
		if !ok {
			t.Fatal("expected asset SWN-F2-001 (Heavy Drop Assets) to be present")
		}
		if asset.Attack != nil {
			t.Errorf("expected nil attack, got %+v", asset.Attack)
		}
		if asset.Counter != nil {
			t.Errorf("expected nil counter, got %+v", asset.Counter)
		}
	})

	t.Run("nil ability for non-action asset", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-F1-001"]
		if !ok {
			t.Fatal("asset SWN-F1-001 not found")
		}
		if asset.Ability != nil {
			t.Errorf("expected nil ability, got %+v", asset.Ability)
		}
	})

	t.Run("faction_test ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C1-002"]
		if !ok {
			t.Fatal("asset SWN-C1-002 not found")
		}
		if asset.Ability == nil {
			t.Fatal("expected ability, got nil")
		}
		if asset.Ability.AttackerStat != domain.StatCunning {
			t.Errorf("attacker_stat: got %q, want %q", asset.Ability.AttackerStat, domain.StatCunning)
		}
		if asset.Ability.DefenderStat != domain.StatCunning {
			t.Errorf("defender_stat: got %q, want %q", asset.Ability.DefenderStat, domain.StatCunning)
		}
		if asset.Ability.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", asset.Ability.Effect, domain.EffectRevealStealth)
		}
	})

	t.Run("marketers has no ability section", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W5-001"]
		if !ok {
			t.Fatal("asset SWN-W5-001 not found")
		}
		if asset.Ability != nil {
			t.Errorf("expected nil ability, got %+v", asset.Ability)
		}
	})

	t.Run("speed populated on movers", func(t *testing.T) {
		cases := []struct {
			id    string
			speed int
		}{
			{"SWN-W2-001", 2}, // FreighterContract
			{"SWN-W2-004", 2}, // Surveyors
			{"SWN-W3-003", 1}, // Mercenaries
			{"SWN-W4-001", 2}, // ShippingCombine
			{"SWN-W5-003", 3}, // BlockadeRunners
			{"SWN-W8-001", 3}, // ScavengerFleet
			{"SWN-C1-001", 2}, // Smugglers
			{"SWN-C2-004", 1}, // Seductress
			{"SWN-F2-001", 1}, // HeavyDropAssets
			{"SWN-F4-001", 1}, // BeachheadLanders
			{"SWN-F4-002", 2}, // ExtendedTheater
			{"SWN-F4-003", 1}, // StrikeFleet
			{"SWN-F5-001", 1}, // BlockadeFleet
			{"SWN-F7-001", 3}, // DeepStrikeLanders
			{"SWN-F7-003", 1}, // SpaceMarines
			{"SWN-F8-001", 3}, // CapitalFleet
		}
		for _, tc := range cases {
			def, ok := rb.Assets[tc.id]
			if !ok {
				t.Errorf("asset %s not found", tc.id)
				continue
			}
			if def.Speed != tc.speed {
				t.Errorf("%s (%s): speed got %d, want %d", def.Name, tc.id, def.Speed, tc.speed)
			}
		}
		nonMover, ok := rb.Assets["SWN-F1-001"]
		if ok && nonMover.Speed != 0 {
			t.Errorf("Security Personnel: expected speed 0, got %d", nonMover.Speed)
		}
	})

	t.Run("transport profiles populated", func(t *testing.T) {
		allTypes := []domain.AssetType{
			domain.TypeFacility, domain.TypeStarship, domain.TypeMilitaryUnit,
			domain.TypeSpecialForces, domain.TypeTactic, domain.TypeLogisticsFacility,
		}
		nonStarship := []domain.AssetType{
			domain.TypeFacility, domain.TypeMilitaryUnit, domain.TypeSpecialForces,
			domain.TypeTactic, domain.TypeLogisticsFacility,
		}
		type transportExpect struct {
			id           string
			maxHex       int
			coinCost     int
			cargoTypes   []domain.AssetType
			maxCargo     int
			excludeStats []domain.FactionStat
		}
		cases := []transportExpect{
			{"SWN-W2-001", 2, 1, allTypes, 1, []domain.FactionStat{domain.StatForce}},                           // FreighterContract
			{"SWN-W4-001", 2, 1, allTypes, 10, []domain.FactionStat{domain.StatForce}},                          // ShippingCombine
			{"SWN-W5-003", 3, 2, []domain.AssetType{domain.TypeMilitaryUnit, domain.TypeSpecialForces}, 1, nil}, // BlockadeRunners
			{"SWN-C1-001", 2, 1, []domain.AssetType{domain.TypeSpecialForces}, 1, nil},                          // Smugglers
			{"SWN-F2-001", 1, 1, nonStarship, 1, nil},                                                           // HeavyDropAssets
			{"SWN-F4-001", 1, 1, allTypes, 10, nil},                                                             // BeachheadLanders
			{"SWN-F4-002", 2, 1, nonStarship, 1, nil},                                                           // ExtendedTheater
			{"SWN-F7-001", 3, 2, nonStarship, 1, nil},                                                           // DeepStrikeLanders
		}
		for _, tc := range cases {
			def, ok := rb.Assets[tc.id]
			if !ok {
				t.Errorf("asset %s not found", tc.id)
				continue
			}
			if def.Transport == nil {
				t.Errorf("%s: expected Transport != nil", tc.id)
				continue
			}
			tp := def.Transport
			if tp.MaxHex != tc.maxHex {
				t.Errorf("%s: MaxHex got %d, want %d", tc.id, tp.MaxHex, tc.maxHex)
			}
			if tp.CoinCost != tc.coinCost {
				t.Errorf("%s: CoinCost got %d, want %d", tc.id, tp.CoinCost, tc.coinCost)
			}
			if tp.MaxCargo != tc.maxCargo {
				t.Errorf("%s: MaxCargo got %d, want %d", tc.id, tp.MaxCargo, tc.maxCargo)
			}
			gotTypes := slices.Clone(tp.CargoTypes)
			slices.Sort(gotTypes)
			wantTypes := slices.Clone(tc.cargoTypes)
			slices.Sort(wantTypes)
			if !slices.Equal(gotTypes, wantTypes) {
				t.Errorf("%s: CargoTypes got %v, want %v", tc.id, gotTypes, wantTypes)
			}
			gotExcl := slices.Clone(tp.ExcludeCategories)
			slices.Sort(gotExcl)
			wantExcl := slices.Clone(tc.excludeStats)
			slices.Sort(wantExcl)
			if !slices.Equal(gotExcl, wantExcl) {
				t.Errorf("%s: ExcludeCategories got %v, want %v", tc.id, gotExcl, wantExcl)
			}
			if !def.HasFlag(domain.FlagSpecial) {
				t.Errorf("%s: expected FlagSpecial in flags, got %v", tc.id, def.Flags)
			}
		}
	})

	t.Run("self-movers have speed and no transport", func(t *testing.T) {
		selfMovers := []string{
			"SWN-W2-004", // Surveyors
			"SWN-W3-003", // Mercenaries
			"SWN-W8-001", // ScavengerFleet
			"SWN-C2-004", // Seductress
			"SWN-F4-003", // StrikeFleet
			"SWN-F7-003", // SpaceMarines
			"SWN-F8-001", // CapitalFleet
		}
		for _, id := range selfMovers {
			def, ok := rb.Assets[id]
			if !ok {
				t.Errorf("asset %s not found", id)
				continue
			}
			if def.Speed <= 0 {
				t.Errorf("%s: expected Speed > 0, got %d", id, def.Speed)
			}
			if def.Transport != nil {
				t.Errorf("%s: expected nil Transport, got %+v", id, def.Transport)
			}
		}
	})

	t.Run("deferred transport assets have no transport and no ability", func(t *testing.T) {
		deferred := []string{"SWN-C3-003", "SWN-C6-002"} // CovertShipping, CovertTransitNet
		for _, id := range deferred {
			def, ok := rb.Assets[id]
			if !ok {
				t.Errorf("asset %s not found", id)
				continue
			}
			if def.Transport != nil {
				t.Errorf("%s: expected nil Transport, got %+v", id, def.Transport)
			}
			if def.Ability != nil {
				t.Errorf("%s: expected nil Ability, got %+v", id, def.Ability)
			}
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
