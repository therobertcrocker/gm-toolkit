package rulebook

import (
	"reflect"
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
		if d == nil {
			t.Fatal("expected damage, got nil")
		}
		if d.NumDice != 1 || d.Sides != 3 || d.Modifier != 1 {
			t.Errorf("damage: got %+v, want 1d3+1", d)
		}
	})

	t.Run("special attack: Treachery", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C7-003"]
		if !ok {
			t.Fatal("asset SWN-C7-003 (Treachery) not found")
		}
		if asset.Attack == nil || asset.Attack.Damage != nil {
			t.Fatalf("attack: got %+v, want a profile with nil Damage", asset.Attack)
		}
		if len(asset.Effects) != 1 || asset.Effects[0].Effect != domain.SpecialCaptureTarget {
			t.Fatalf("effects: got %+v, want one capture_target entry", asset.Effects)
		}
		if capture := asset.Effects[0].CaptureTarget; capture == nil || capture.Coin.Modifier != 5 {
			t.Errorf("capture_target: got %+v, want Coin modifier 5", capture)
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
		if len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %d, want 1", len(asset.Ability.Steps))
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectRevealStealth)
		}
		if step.Contest == nil {
			t.Fatal("expected a contest on the reveal step, got nil")
		}
		if step.Contest.AttackerStat != domain.StatCunning {
			t.Errorf("attacker_stat: got %q, want %q", step.Contest.AttackerStat, domain.StatCunning)
		}
		if step.Contest.DefenderStat != domain.StatCunning {
			t.Errorf("defender_stat: got %q, want %q", step.Contest.DefenderStat, domain.StatCunning)
		}
		if step.Target != domain.TargetFaction {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetFaction)
		}
	})

	t.Run("seductress relocate-then-reveal steps parsed from [[steps]]", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C2-004"]
		if !ok {
			t.Fatal("asset SWN-C2-004 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 2 {
			t.Fatalf("steps: got %v, want 2", asset.Ability)
		}
		relocateStep := asset.Ability.Steps[0]
		if relocateStep.Effect != domain.EffectRelocate {
			t.Errorf("step 0 effect: got %q, want %q", relocateStep.Effect, domain.EffectRelocate)
		}
		if relocateStep.Relocate == nil || relocateStep.Relocate.Scope != domain.RelocateSelf || relocateStep.Relocate.MaxHex != 1 {
			t.Errorf("step 0 relocate: got %+v, want Scope=self MaxHex=1", relocateStep.Relocate)
		}
		if relocateStep.Relocate != nil && relocateStep.Relocate.Range != domain.RangeHexes {
			t.Errorf("step 0 range: got %q, want %q (default)", relocateStep.Relocate.Range, domain.RangeHexes)
		}
		revealStep := asset.Ability.Steps[1]
		if revealStep.Effect != domain.EffectRevealStealth {
			t.Errorf("step 1 effect: got %q, want %q", revealStep.Effect, domain.EffectRevealStealth)
		}
		if revealStep.Target != domain.TargetFaction {
			t.Errorf("step 1 target: got %q, want %q", revealStep.Target, domain.TargetFaction)
		}
		if revealStep.Contest == nil || revealStep.Contest.AttackerStat != domain.StatCunning {
			t.Errorf("step 1 contest: got %+v, want Cunning vs Cunning", revealStep.Contest)
		}
	})

	t.Run("covert shipping relocate ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C3-003"]
		if !ok {
			t.Fatal("asset SWN-C3-003 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectRelocate || step.Relocate == nil {
			t.Fatalf("step: got effect %q relocate %+v, want relocate effect with params", step.Effect, step.Relocate)
		}
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, CoinCostMode: domain.RelocateCostPerAsset, AssetType: domain.TypeSpecialForces, MaxAssets: 1}
		if !reflect.DeepEqual(*step.Relocate, want) {
			t.Errorf("relocate: got %+v, want %+v", *step.Relocate, want)
		}
	})

	t.Run("covert transit net region relocate ability parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C6-002"]
		if !ok {
			t.Fatal("asset SWN-C6-002 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectRelocate || step.Relocate == nil {
			t.Fatalf("step: got effect %q relocate %+v, want relocate effect with params", step.Effect, step.Relocate)
		}
		want := domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeRegion, CoinCostMode: domain.RelocateCostPerAsset, AssetType: domain.TypeSpecialForces, MaxAssets: 0}
		if !reflect.DeepEqual(*step.Relocate, want) {
			t.Errorf("relocate: got %+v, want %+v", *step.Relocate, want)
		}
	})

	t.Run("venture capital coin_delta roll table parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W6-001"]
		if !ok {
			t.Fatal("asset SWN-W6-001 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectCoinDelta {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectCoinDelta)
		}
		if step.Roll == nil || step.Roll.Die.Sides != 8 {
			t.Fatalf("roll: got %+v, want 1d8", step.Roll)
		}
		if len(step.Roll.Outcomes) != 4 {
			t.Fatalf("outcomes: got %d, want 4", len(step.Roll.Outcomes))
		}
		if !step.Roll.Outcomes[0].DestroySelf {
			t.Errorf("outcome[0]: want destroy_self on roll 1")
		}
	})

	t.Run("marketers disables an asset", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W5-001"]
		if !ok {
			t.Fatal("asset SWN-W5-001 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("ability: got %+v, want one step", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectDisableAsset {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectDisableAsset)
		}
		if step.Target != domain.TargetAsset {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetAsset)
		}
		if step.Contest == nil || step.Contest.AttackerStat != domain.StatCunning || step.Contest.DefenderStat != domain.StatWealth {
			t.Errorf("contest: got %+v, want Cunning vs Wealth", step.Contest)
		}
	})

	t.Run("seditionists block an asset's attack", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C4-004"]
		if !ok {
			t.Fatal("asset SWN-C4-004 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("ability: got %+v, want one step", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectBlockAttack {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectBlockAttack)
		}
		if step.Target != domain.TargetAsset {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetAsset)
		}
		if step.Roll == nil || step.Roll.Die.NumDice != 1 || step.Roll.Die.Sides != 4 {
			t.Errorf("roll: got %+v, want 1d4", step.Roll)
		}
	})

	t.Run("monopoly levies a faction and keeps its attack", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W4-002"]
		if !ok {
			t.Fatal("asset SWN-W4-002 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("ability: got %+v, want one step", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectLevyCoin {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectLevyCoin)
		}
		if step.Target != domain.TargetFaction {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetFaction)
		}
		if asset.Attack == nil {
			t.Error("attack: got nil, want the book attack kept")
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

	t.Run("deferred transport assets have no transport", func(t *testing.T) {
		// Their relocate abilities are asserted in the dedicated subtests above; here
		// only the absence of a TransportProfile matters (relocate is instant, not a
		// drift-order transport).
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

	t.Run("pretech logistics purchase_enabler parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-F5-002"]
		if !ok {
			t.Fatal("asset SWN-F5-002 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectPurchaseEnabler {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectPurchaseEnabler)
		}
		if step.Purchase == nil {
			t.Fatal("purchase params not parsed")
		}
		if step.Purchase.Category != domain.StatForce || step.Purchase.MaxTechLevel != 5 || step.Purchase.CostPercent != 150 {
			t.Errorf("purchase = %+v, want {Force, 5, 150}", step.Purchase)
		}
	})

	t.Run("party machine coin_gain effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C4-001"]
		if !ok {
			t.Fatal("asset SWN-C4-001 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("false front sacrifice_self effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C1-003"]
		if !ok {
			t.Fatal("asset SWN-C1-003 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialSacrificeSelf,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("boltholes shelter_asset effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C5-003"]
		if !ok {
			t.Fatal("asset SWN-C5-003 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialShelterAsset,
			Filters: domain.EffectFilters{
				Owner:      domain.OwnerOwn,
				Scope:      domain.ScopeWorld,
				AssetTypes: []domain.AssetType{domain.TypeSpecialForces, domain.TypeMilitaryUnit},
			},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("hostile takeover seize_target effect parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W7-002"]
		if !ok {
			t.Fatal("asset SWN-W7-002 not found")
		}
		want := []domain.EffectDefinition{{
			Trigger: domain.TriggerLethalDamage,
			Effect:  domain.SpecialSeizeTarget,
			Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf},
		}}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
		}
	})

	t.Run("medical center restore_asset and discount_repair effects parsed", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-W4-003"]
		if !ok {
			t.Fatal("asset SWN-W4-003 not found")
		}
		filters := domain.EffectFilters{
			Owner:      domain.OwnerOwn,
			Scope:      domain.ScopeWorld,
			AssetTypes: []domain.AssetType{domain.TypeSpecialForces, domain.TypeMilitaryUnit},
		}
		want := []domain.EffectDefinition{
			{
				Trigger: domain.TriggerLethalDamage,
				Effect:  domain.SpecialRestoreAsset,
				Filters: filters,
				Gates:   domain.EffectGates{OncePer: domain.OncePerAsset},
			},
			{
				Trigger:        domain.TriggerRepairCostQuery,
				Effect:         domain.SpecialDiscountRepair,
				Filters:        filters,
				Gates:          domain.EffectGates{NoStack: true},
				DiscountRepair: &domain.DiscountRepairParams{Amount: 1},
			},
		}
		if !reflect.DeepEqual(asset.Effects, want) {
			t.Errorf("effects: got %+v, want %+v", asset.Effects, want)
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

func TestConvertAbilityStep_DisableAssetNeedsAnAssetTarget(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+`
[assets.Machine.ability]
attacker_stat = "Cunning"
defender_stat = "Wealth"
effect = "disable_asset"
target = "faction"
`)
	if err == nil || !strings.Contains(err.Error(), "disable_asset requires") {
		t.Fatalf("err = %v, want a disable_asset target error", err)
	}
}

func TestConvertAbilityStep_BlockAttackNeedsADie(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+`
[assets.Machine.ability]
effect = "block_attack"
target = "asset"
`)
	if err == nil || !strings.Contains(err.Error(), "block_attack requires") {
		t.Fatalf("err = %v, want a block_attack die error", err)
	}
}

func TestConvertAbilityStep_LevyCoinRefusesAContest(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+`
[assets.Machine.ability]
attacker_stat = "Wealth"
defender_stat = "Wealth"
effect = "levy_coin"
target = "faction"
`)
	if err == nil || !strings.Contains(err.Error(), "levy_coin requires") {
		t.Fatalf("err = %v, want a levy_coin contest error", err)
	}
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
