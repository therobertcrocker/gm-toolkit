package rulebook

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const effectTestAsset = `
[assets.Machine]
id = "T-001"
name = "Machine"
category = "Cunning"
min_rating = 1
hp = 1
cost = 1
tech_level = 0
type = "Facility"
flags = ["S"]
counter = ""
description = "Test asset."
`

func loadAssetsFromTOML(t *testing.T, body string) (map[string]*domain.AssetDefinition, error) {
	t.Helper()
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("creating assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "test_assets.toml"), []byte(body), 0o644); err != nil {
		t.Fatalf("writing assets file: %v", err)
	}
	return loadAssets(dir)
}

func TestLoadEffects_ParsesEntries(t *testing.T) {
	assets, err := loadAssetsFromTOML(t, effectTestAsset+`
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"

[[assets.Machine.effects]]
trigger = "own_attack_hit"
effect  = "coin_steal"
[assets.Machine.effects.filters]
owner      = "rival"
scope      = "world"
attribute  = "Cunning"
phase      = ["defense"]
action     = "Expand Influence"
asset_type = ["Starship"]
category   = ["Force", "Wealth"]
stealthy   = true
[assets.Machine.effects.gates]
once_per = "target_faction"
no_stack = true
[assets.Machine.effects.gates.contest]
roller        = "target"
attacker_stat = "Cunning"
defender_stat = "Wealth"
[assets.Machine.effects.coin_steal]
amount = "1d4"

[[assets.Machine.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"
[assets.Machine.effects.filters]
owner = "own"
scope = "faction"
[assets.Machine.effects.grant_movement.relocate]
scope          = "other"
range          = "hexes"
max_hex        = 3
coin_cost      = 1
coin_cost_mode = "flat"
categories     = ["Cunning", "Wealth"]
exclude_types  = ["Starship"]

[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
`)
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	stealthy := true
	want := []domain.EffectDefinition{
		{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		},
		{
			Trigger: domain.TriggerOwnAttackHit,
			Effect:  domain.SpecialCoinSteal,
			Filters: domain.EffectFilters{
				Owner:      domain.OwnerRival,
				Scope:      domain.ScopeWorld,
				Attribute:  domain.StatCunning,
				Phases:     []domain.EffectRollPhase{domain.RollPhaseDefense},
				Action:     "Expand Influence",
				AssetTypes: []domain.AssetType{domain.TypeStarship},
				Categories: []domain.FactionStat{domain.StatForce, domain.StatWealth},
				Stealthy:   &stealthy,
			},
			Gates: domain.EffectGates{
				OncePer: domain.OncePerTargetFaction,
				NoStack: true,
				Contest: &domain.EffectContest{
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
					Roller:  domain.ContestRollerTarget,
				},
			},
			CoinSteal: &domain.CoinStealParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}},
		},
		{
			Trigger: domain.TriggerMovementOptionsQuery,
			Effect:  domain.SpecialGrantMovement,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			GrantMovement: &domain.GrantMovementParams{Relocate: domain.RelocateParams{
				Scope:        domain.RelocateOther,
				Range:        domain.RangeHexes,
				MaxHex:       3,
				CoinCost:     1,
				CoinCostMode: domain.RelocateCostFlat,
				Categories:   []domain.FactionStat{domain.StatCunning, domain.StatWealth},
				ExcludeTypes: []domain.AssetType{domain.TypeStarship},
			}},
		},
		{
			Trigger: domain.TriggerOnPurchase,
			Effect:  domain.SpecialApplyStealth,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		},
	}
	if got := assets["T-001"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("effects:\n got  %+v\n want %+v", got, want)
	}
}

func TestLoadEffects_ParsesEveryParameterBlock(t *testing.T) {
	entry := func(trigger, effect, block string) string {
		return `
[[assets.Machine.effects]]
trigger = "` + trigger + `"
effect  = "` + effect + `"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
[assets.Machine.effects.` + effect + `]
` + block
	}
	assets, err := loadAssetsFromTOML(t, effectTestAsset+
		entry("own_attack_hit", "self_damage", `amount = "1d4"`)+
		entry("own_roll", "bonus_die", "sides = 10\ncount = 1")+
		entry("own_roll", "reroll", "dice = 1\nelective = true")+
		entry("tech_level_query", "tech_level_floor", "level = 4")+
		entry("asset_cost_query", "cost_surcharge", "amount = 1")+
		entry("attack_eligibility_query", "restrict_attack", `side = "defense"`+"\n"+`allow = "only_listed"`))
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	effects := assets["T-001"].Effects
	if len(effects) != 6 {
		t.Fatalf("effects: got %d, want 6", len(effects))
	}
	checks := []struct {
		name string
		got  any
		want any
	}{
		{"self_damage", effects[0].SelfDamage, &domain.SelfDamageParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}}},
		{"bonus_die", effects[1].BonusDie, &domain.BonusDieParams{Sides: 10, Count: 1}},
		{"reroll", effects[2].Reroll, &domain.RerollParams{Dice: 1, Elective: true}},
		{"tech_level_floor", effects[3].TechLevelFloor, &domain.TechLevelFloorParams{Level: 4}},
		{"cost_surcharge", effects[4].CostSurcharge, &domain.CostSurchargeParams{Amount: 1}},
		{"restrict_attack", effects[5].RestrictAttack, &domain.RestrictAttackParams{Side: domain.RestrictDefense, Allow: domain.AllowOnlyListed}},
	}
	for _, check := range checks {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("%s: got %+v, want %+v", check.name, check.got, check.want)
		}
	}
}

func TestLoadEffects_LoudFailure(t *testing.T) {
	const filters = `
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
`
	cases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name: "unknown trigger",
			body: `
[[assets.Machine.effects]]
trigger = "dawn"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `unknown effect trigger: "dawn"`,
		},
		{
			name: "unknown effect",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_fountain"` + filters,
			wantErr: `unknown special effect: "coin_fountain"`,
		},
		{
			name: "unknown filter name",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `bogus_filter = "x"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "bogus_filter",
		},
		{
			name: "unknown gate name",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.gates]
bogus_gate = true
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "bogus_gate",
		},
		{
			name: "unknown parameter block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_fountain]
amount = "1"
`,
			wantErr: "coin_fountain",
		},
		{
			name:    "unknown key elsewhere in the asset",
			body:    "bogus_field = 1\n",
			wantErr: "bogus_field",
		},
		{
			name: "parameter block does not match the effect",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_steal]
amount = "1"
`,
			wantErr: `parameter block [coin_steal] does not match effect "coin_gain"`,
		},
		{
			name: "parameter block on an effect that takes none",
			body: `
[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `parameter block [coin_gain] does not match effect "apply_stealth"`,
		},
		{
			name: "missing parameter block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters,
			wantErr: `effect "coin_gain" requires a [coin_gain] block`,
		},
		{
			name: "capture_target missing its block",
			body: `
[[assets.Machine.effects]]
trigger = "own_attack_hit"
effect  = "capture_target"` + filters,
			wantErr: `effect "capture_target" requires a [capture_target] block`,
		},
		{
			name: "discount_repair missing its block",
			body: `
[[assets.Machine.effects]]
trigger = "repair_cost_query"
effect  = "discount_repair"` + filters,
			wantErr: `effect "discount_repair" requires a [discount_repair] block`,
		},
		{
			name: "discount_repair amount 0",
			body: `
[[assets.Machine.effects]]
trigger = "repair_cost_query"
effect  = "discount_repair"` + filters + `
[assets.Machine.effects.discount_repair]
amount = 0
`,
			wantErr: "discount_repair amount must be > 0, got 0",
		},
		{
			name: "no-damage attack with no own_attack_hit effect",
			body: `
[assets.Machine.attack]
attacker_stat = "Cunning"
defender_stat = "Cunning"
`,
			wantErr: "attack has no damage and no own_attack_hit effect",
		},
		{
			name: "missing amount",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
`,
			wantErr: "coin_gain amount is required",
		},
		{
			name: "missing filters block",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "missing [filters] block",
		},
		{
			name: "missing owner",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "filters.owner is required",
		},
		{
			name: "missing scope",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "own"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: "filters.scope is required",
		},
		{
			name: "unknown owner value",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.Machine.effects.filters]
owner = "mine"
scope = "self"
[assets.Machine.effects.coin_gain]
amount = "1"
`,
			wantErr: `unknown effect owner: "mine"`,
		},
		{
			name: "unknown phase value",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.Machine.effects.filters]
owner = "own"
scope = "self"
phase = ["defence"]
[assets.Machine.effects.bonus_die]
sides = 10
count = 1
`,
			wantErr: "filters.phase[0]",
		},
		{
			name: "contest without a roller",
			body: `
[[assets.Machine.effects]]
trigger = "stealth_arrival"
effect  = "clear_stealth"` + filters + `
[assets.Machine.effects.gates.contest]
attacker_stat = "Cunning"
defender_stat = "Cunning"
`,
			wantErr: `unknown contest roller: ""`,
		},
		{
			name: "grant_movement without a relocate block",
			body: `
[[assets.Machine.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"` + filters + `
[assets.Machine.effects.grant_movement]
`,
			wantErr: "grant_movement requires a [grant_movement.relocate] block",
		},
		{
			name: "bonus_die with zero sides",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "bonus_die"` + filters + `
[assets.Machine.effects.bonus_die]
sides = 0
count = 1
`,
			wantErr: "bonus_die sides and count must be > 0, got 0 and 1",
		},
		{
			name: "reroll with zero dice",
			body: `
[[assets.Machine.effects]]
trigger = "own_roll"
effect  = "reroll"` + filters + `
[assets.Machine.effects.reroll]
dice = 0
`,
			wantErr: "reroll dice must be > 0, got 0",
		},
		{
			name: "tech_level_floor with zero level",
			body: `
[[assets.Machine.effects]]
trigger = "tech_level_query"
effect  = "tech_level_floor"` + filters + `
[assets.Machine.effects.tech_level_floor]
level = 0
`,
			wantErr: "tech_level_floor level must be > 0, got 0",
		},
		{
			name: "cost_surcharge with zero amount",
			body: `
[[assets.Machine.effects]]
trigger = "asset_cost_query"
effect  = "cost_surcharge"` + filters + `
[assets.Machine.effects.cost_surcharge]
amount = 0
`,
			wantErr: "cost_surcharge amount must be > 0, got 0",
		},
		{
			name: "coin_gain with a zero amount",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "0"
`,
			wantErr: `coin_gain amount must be > 0, got "0"`,
		},
		{
			name: "coin_gain with zero-sided dice",
			body: `
[[assets.Machine.effects]]
trigger = "turn_start"
effect  = "coin_gain"` + filters + `
[assets.Machine.effects.coin_gain]
amount = "1d0"
`,
			wantErr: "sides must be > 0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadAssetsFromTOML(t, effectTestAsset+tc.body)
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestConvertAbilityStep_RelocateCarriesEffortFiveFields(t *testing.T) {
	assets, err := loadAssetsFromTOML(t, effectTestAsset+`
[assets.Machine.ability]
effect = "relocate"
[assets.Machine.ability.relocate]
scope          = "other"
max_hex        = 3
coin_cost      = 1
coin_cost_mode = "flat"
categories     = ["Wealth"]
exclude_types  = ["Starship"]
`)
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	steps := assets["T-001"].Ability.Steps
	if len(steps) != 1 || steps[0].Relocate == nil {
		t.Fatalf("steps = %+v, want one relocate step", steps)
	}
	relocate := steps[0].Relocate
	if !reflect.DeepEqual(relocate.Categories, []domain.FactionStat{domain.StatWealth}) {
		t.Errorf("Categories = %v, want [Wealth]", relocate.Categories)
	}
	if !reflect.DeepEqual(relocate.ExcludeTypes, []domain.AssetType{domain.TypeStarship}) {
		t.Errorf("ExcludeTypes = %v, want [Starship]", relocate.ExcludeTypes)
	}
	if relocate.CoinCostMode != domain.RelocateCostFlat {
		t.Errorf("CoinCostMode = %q, want %q", relocate.CoinCostMode, domain.RelocateCostFlat)
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		input string
		want  *domain.DiceRoll
	}{
		{"1", &domain.DiceRoll{Modifier: 1}},
		{"1d4", &domain.DiceRoll{NumDice: 1, Sides: 4}},
		{"2d6+1", &domain.DiceRoll{NumDice: 2, Sides: 6, Modifier: 1}},
		{"", nil},
	}
	for _, tc := range cases {
		got, err := parseAmount(tc.input)
		if err != nil {
			t.Fatalf("parseAmount(%q): %v", tc.input, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseAmount(%q) = %+v, want %+v", tc.input, got, tc.want)
		}
	}
	if _, err := parseAmount("lots"); err == nil {
		t.Error(`parseAmount("lots"): expected an error`)
	}
}

func TestLoadEffects_QualityNeedsOneOnPurchaseEntry(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+"quality = true\n")
	if err == nil || !strings.Contains(err.Error(), "a quality needs exactly one on_purchase effect") {
		t.Fatalf("err = %v, want a quality shape error", err)
	}
}

func TestLoadEffects_GrantStealthNeedsAQuality(t *testing.T) {
	_, err := loadAssetsFromTOML(t, effectTestAsset+`
[[assets.Machine.effects]]
trigger = "on_purchase"
effect  = "grant_stealth"
[assets.Machine.effects.filters]
owner = "own"
scope = "world"
`)
	if err == nil || !strings.Contains(err.Error(), "grant_stealth needs quality = true") {
		t.Fatalf("err = %v, want a grant_stealth quality error", err)
	}
}
