package rulebook

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// Rulebook holds all static game data loaded from the data directory.
// It is the authoritative source of rules for the faction engine.
type Rulebook struct {
	Assets     map[string]*domain.AssetDefinition
	Tags       map[string]*domain.Tag
	Goals      map[string]*domain.Goal
	DriftCosts []int
}

// DriftCost returns the hex-equivalent crossing cost for the given drift rating.
// Out-of-range ratings fall back to DriftCosts[0] (worst cost).
func (rulebook *Rulebook) DriftCost(driftRating int) int {
	if driftRating < 1 || driftRating > len(rulebook.DriftCosts) {
		return rulebook.DriftCosts[0]
	}
	return rulebook.DriftCosts[driftRating-1]
}

// Load reads all static data files from dataDir and returns a populated Rulebook.
func Load(dataDir string) (*Rulebook, error) {
	assets, err := loadAssets(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading assets: %w", err)
	}

	tags, err := loadTags(filepath.Join(dataDir, "tags.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading tags: %w", err)
	}

	goals, err := loadGoals(filepath.Join(dataDir, "goals.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading goals: %w", err)
	}

	driftCosts, err := loadDriftCosts(filepath.Join(dataDir, "drift_costs.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading drift costs: %w", err)
	}

	return &Rulebook{
		Assets:     assets,
		Tags:       tags,
		Goals:      goals,
		DriftCosts: driftCosts,
	}, nil
}

// ---------------------------------------------------------------------------
// Assets
// ---------------------------------------------------------------------------

type assetFile struct {
	Assets map[string]assetRecord `toml:"assets"`
}

type assetRecord struct {
	ID          string           `toml:"id"`
	Name        string           `toml:"name"`
	Category    string           `toml:"category"`
	MinRating   int              `toml:"min_rating"`
	HP          int              `toml:"hp"`
	Cost        int              `toml:"cost"`
	Maintenance int              `toml:"maintenance"`
	TechLevel   int              `toml:"tech_level"`
	Speed       int              `toml:"speed"`
	DriftRating int              `toml:"drift_rating"`
	Type        string           `toml:"type"`
	Flags       []string         `toml:"flags"`
	Counter     string           `toml:"counter"`
	Description string           `toml:"description"`
	Attack      *attackRecord    `toml:"attack"`
	Ability     *abilityRecord   `toml:"ability"`
	Transport   *transportRecord `toml:"transport"`
	Effects     []effectRecord   `toml:"effects"`
}

type transportRecord struct {
	MaxHex            int      `toml:"max_hex"`
	CoinCost          int      `toml:"coin_cost"`
	CargoTypes        []string `toml:"cargo_types"`
	MaxCargo          int      `toml:"max_cargo"`
	ExcludeCategories []string `toml:"exclude_categories"`
}

type attackRecord struct {
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
	Damage       string `toml:"damage"`
}

type abilityRecord struct {
	abilityStepRecord
	Steps []abilityStepRecord `toml:"steps"`
}

type abilityStepRecord struct {
	Effect       string          `toml:"effect"`
	Target       string          `toml:"target"`
	AttackerStat string          `toml:"attacker_stat"`
	DefenderStat string          `toml:"defender_stat"`
	Die          string          `toml:"die"`
	Outcomes     []outcomeRecord `toml:"outcomes"`
	Relocate     *relocateRecord `toml:"relocate"`
	Purchase     *purchaseRecord `toml:"purchase"`
}

type relocateRecord struct {
	Scope        string   `toml:"scope"`
	Range        string   `toml:"range"`
	MaxHex       int      `toml:"max_hex"`
	CoinCost     int      `toml:"coin_cost"`
	CoinCostMode string   `toml:"coin_cost_mode"`
	AssetType    string   `toml:"asset_type"`
	Categories   []string `toml:"categories"`
	ExcludeTypes []string `toml:"exclude_types"`
	MaxAssets    int      `toml:"max_assets"`
}

type purchaseRecord struct {
	Category     string `toml:"category"`
	MaxTechLevel int    `toml:"max_tech_level"`
	CostPercent  int    `toml:"cost_percent"`
}

type outcomeRecord struct {
	From        int  `toml:"from"`
	To          int  `toml:"to"`
	Coin        int  `toml:"coin"`
	DestroySelf bool `toml:"destroy_self"`
}

// loadAssets globs all *_assets.toml files in <dataDir>/assets/ and merges them into one map.
func loadAssets(dataDir string) (map[string]*domain.AssetDefinition, error) {
	files, err := filepath.Glob(filepath.Join(dataDir, "assets", "*_assets.toml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no asset files found in %s", dataDir)
	}

	result := make(map[string]*domain.AssetDefinition)
	for _, path := range files {
		var f assetFile
		metadata, err := toml.DecodeFile(path, &f)
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", path, err)
		}
		// deliberate: unknown keys fail the load, not skip, because the decoder drops them silently and a misspelt filter or gate would do nothing.
		if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
			keys := make([]string, len(undecoded))
			for i, key := range undecoded {
				keys[i] = key.String()
			}
			return nil, fmt.Errorf("unknown keys in %s: %s", path, strings.Join(keys, ", "))
		}
		for _, record := range f.Assets {
			def, err := convertAsset(record)
			if err != nil {
				return nil, fmt.Errorf("asset %q in %s: %w", record.ID, path, err)
			}
			if _, exists := result[def.ID]; exists {
				return nil, fmt.Errorf("duplicate asset ID %q in %s", def.ID, path)
			}
			result[def.ID] = def
		}
	}
	return result, nil
}

func convertAsset(r assetRecord) (*domain.AssetDefinition, error) {
	category, err := toFactionStat(r.Category)
	if err != nil {
		return nil, err
	}
	assetType, err := toAssetType(r.Type)
	if err != nil {
		return nil, err
	}
	flags, err := toAssetFlags(r.Flags)
	if err != nil {
		return nil, err
	}
	counter, err := parseDice(r.Counter)
	if err != nil {
		return nil, fmt.Errorf("counter: %w", err)
	}

	var attack *domain.AttackProfile
	if r.Attack != nil {
		attackerStat, err := toFactionStat(r.Attack.AttackerStat)
		if err != nil {
			return nil, fmt.Errorf("attack attacker_stat: %w", err)
		}
		defenderStat, err := toFactionStat(r.Attack.DefenderStat)
		if err != nil {
			return nil, fmt.Errorf("attack defender_stat: %w", err)
		}
		damage, err := parseDice(r.Attack.Damage)
		if err != nil {
			return nil, fmt.Errorf("attack damage: %w", err)
		}
		attack = &domain.AttackProfile{
			AttackerStat: attackerStat,
			DefenderStat: defenderStat,
			Damage:       damage,
		}
	}

	ability, err := convertAbility(r.Ability)
	if err != nil {
		return nil, fmt.Errorf("ability: %w", err)
	}

	transport, err := convertTransport(r.Transport)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	effects, err := convertEffects(r.Effects)
	if err != nil {
		return nil, err
	}
	// deliberate: a no-damage attack needs a hit rule, not a silent load, because an omitted damage line would otherwise hit for nothing and do nothing.
	if attack != nil && attack.Damage == nil && !slices.ContainsFunc(effects, func(entry domain.EffectDefinition) bool {
		return entry.Trigger == domain.TriggerOwnAttackHit
	}) {
		return nil, fmt.Errorf("attack has no damage and no own_attack_hit effect")
	}

	return &domain.AssetDefinition{
		ID:          r.ID,
		Name:        r.Name,
		Category:    category,
		MinRating:   r.MinRating,
		HP:          r.HP,
		Cost:        r.Cost,
		Maintenance: r.Maintenance,
		TechLevel:   r.TechLevel,
		Speed:       r.Speed,
		DriftRating: r.DriftRating,
		Type:        assetType,
		Attack:      attack,
		Counter:     counter,
		Flags:       flags,
		Description: r.Description,
		Ability:     ability,
		Transport:   transport,
		Effects:     effects,
	}, nil
}

func convertAbility(r *abilityRecord) (*domain.AbilityDefinition, error) {
	if r == nil {
		return nil, nil
	}
	if len(r.Steps) > 0 {
		if r.Effect != "" || r.AttackerStat != "" || r.DefenderStat != "" || r.Die != "" {
			return nil, fmt.Errorf("use either a flat [ability] block or [[ability.steps]], not both")
		}
		steps := make([]domain.AbilityStep, 0, len(r.Steps))
		for i, sr := range r.Steps {
			step, err := convertAbilityStep(sr)
			if err != nil {
				return nil, fmt.Errorf("step %d: %w", i, err)
			}
			steps = append(steps, step)
		}
		return &domain.AbilityDefinition{Steps: steps}, nil
	}
	step, err := convertAbilityStep(r.abilityStepRecord)
	if err != nil {
		return nil, err
	}
	return &domain.AbilityDefinition{Steps: []domain.AbilityStep{step}}, nil
}

func convertAbilityStep(record abilityStepRecord) (domain.AbilityStep, error) {
	var step domain.AbilityStep
	if record.Effect != "" {
		effect, err := toAbilityEffect(record.Effect)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		step.Effect = effect
	}
	target, err := toTargetKind(record.Target)
	if err != nil {
		return domain.AbilityStep{}, err
	}
	step.Target = target
	if record.AttackerStat != "" || record.DefenderStat != "" {
		attacker, err := toFactionStat(record.AttackerStat)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("attacker_stat: %w", err)
		}
		defender, err := toFactionStat(record.DefenderStat)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("defender_stat: %w", err)
		}
		step.Contest = &domain.Contest{AttackerStat: attacker, DefenderStat: defender}
	}
	if record.Die != "" {
		die, err := parseDice(record.Die)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("die: %w", err)
		}
		roll := &domain.RollTable{Die: *die}
		roll.Outcomes = make([]domain.AbilityOutcome, 0, len(record.Outcomes))
		for _, o := range record.Outcomes {
			roll.Outcomes = append(roll.Outcomes, domain.AbilityOutcome{
				From:        o.From,
				To:          o.To,
				Coin:        o.Coin,
				DestroySelf: o.DestroySelf,
			})
		}
		step.Roll = roll
	}
	if relocate := record.Relocate; relocate != nil {
		params, err := convertRelocate(relocate)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		step.Relocate = params
	}
	if purchase := record.Purchase; purchase != nil {
		category, err := toFactionStat(purchase.Category)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("purchase category: %w", err)
		}
		if purchase.CostPercent <= 0 {
			return domain.AbilityStep{}, fmt.Errorf("purchase cost_percent must be > 0, got %d", purchase.CostPercent)
		}
		step.Purchase = &domain.PurchaseParams{
			Category:     category,
			MaxTechLevel: purchase.MaxTechLevel,
			CostPercent:  purchase.CostPercent,
		}
	}
	if step.Contest != nil && step.Roll != nil {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and roll; they are mutually exclusive")
	}
	if step.Effect == domain.EffectRevealStealth && step.Target != domain.TargetFaction {
		return domain.AbilityStep{}, fmt.Errorf("reveal_stealth requires target = %q", "faction")
	}
	if step.Effect == domain.EffectDisableAsset && (step.Target != domain.TargetAsset || step.Contest == nil) {
		return domain.AbilityStep{}, fmt.Errorf("disable_asset requires target = %q and a contest", "asset")
	}
	if step.Effect == domain.EffectBlockAttack && (step.Target != domain.TargetAsset || step.Roll == nil) {
		return domain.AbilityStep{}, fmt.Errorf("block_attack requires target = %q and a die", "asset")
	}
	if step.Effect == domain.EffectLevyCoin && (step.Target != domain.TargetFaction || step.Contest != nil || step.Roll != nil) {
		return domain.AbilityStep{}, fmt.Errorf("levy_coin requires target = %q and no contest or die", "faction")
	}
	return step, nil
}

func toRelocateScope(s string) (domain.RelocateScope, error) {
	switch s {
	case "self":
		return domain.RelocateSelf, nil
	case "other":
		return domain.RelocateOther, nil
	default:
		return "", fmt.Errorf("unknown relocate scope: %q", s)
	}
}

func toRelocateRange(s string) (domain.RelocateRange, error) {
	switch s {
	case "", "hexes":
		return domain.RangeHexes, nil
	case "region":
		return domain.RangeRegion, nil
	default:
		return "", fmt.Errorf("unknown relocate range: %q", s)
	}
}

func toRelocateCostMode(s string) (domain.RelocateCostMode, error) {
	switch s {
	case "", "per_asset":
		return domain.RelocateCostPerAsset, nil
	case "flat":
		return domain.RelocateCostFlat, nil
	default:
		return "", fmt.Errorf("unknown relocate coin_cost_mode: %q", s)
	}
}

func convertRelocate(record *relocateRecord) (*domain.RelocateParams, error) {
	scope, err := toRelocateScope(record.Scope)
	if err != nil {
		return nil, err
	}
	rng, err := toRelocateRange(record.Range)
	if err != nil {
		return nil, err
	}
	costMode, err := toRelocateCostMode(record.CoinCostMode)
	if err != nil {
		return nil, err
	}
	params := &domain.RelocateParams{
		Scope:        scope,
		Range:        rng,
		MaxHex:       record.MaxHex,
		CoinCost:     record.CoinCost,
		CoinCostMode: costMode,
		MaxAssets:    record.MaxAssets,
	}
	if record.AssetType != "" {
		assetType, err := toAssetType(record.AssetType)
		if err != nil {
			return nil, fmt.Errorf("relocate asset_type: %w", err)
		}
		params.AssetType = assetType
	}
	for i, value := range record.Categories {
		category, err := toFactionStat(value)
		if err != nil {
			return nil, fmt.Errorf("relocate categories[%d]: %w", i, err)
		}
		params.Categories = append(params.Categories, category)
	}
	for i, value := range record.ExcludeTypes {
		assetType, err := toAssetType(value)
		if err != nil {
			return nil, fmt.Errorf("relocate exclude_types[%d]: %w", i, err)
		}
		params.ExcludeTypes = append(params.ExcludeTypes, assetType)
	}
	return params, nil
}

func toTargetKind(s string) (domain.TargetKind, error) {
	switch s {
	case "", "none":
		return domain.TargetNone, nil
	case "self":
		return domain.TargetSelf, nil
	case "faction":
		return domain.TargetFaction, nil
	case "asset":
		return domain.TargetAsset, nil
	case "world":
		return domain.TargetWorld, nil
	default:
		return "", fmt.Errorf("unknown target kind: %q", s)
	}
}

func toAbilityEffect(s string) (domain.AbilityEffectType, error) {
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	case "coin_delta":
		return domain.EffectCoinDelta, nil
	case "purchase_discount":
		return domain.EffectPurchaseDiscount, nil
	case "relocate":
		return domain.EffectRelocate, nil
	case "purchase_enabler":
		return domain.EffectPurchaseEnabler, nil
	case "disable_asset":
		return domain.EffectDisableAsset, nil
	case "block_attack":
		return domain.EffectBlockAttack, nil
	case "levy_coin":
		return domain.EffectLevyCoin, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
}

func convertTransport(r *transportRecord) (*domain.TransportProfile, error) {
	if r == nil {
		return nil, nil
	}
	cargoTypes := make([]domain.AssetType, len(r.CargoTypes))
	for i, s := range r.CargoTypes {
		t, err := toAssetType(s)
		if err != nil {
			return nil, fmt.Errorf("cargo_types[%d]: %w", i, err)
		}
		cargoTypes[i] = t
	}
	excludeCategories := make([]domain.FactionStat, len(r.ExcludeCategories))
	for i, s := range r.ExcludeCategories {
		stat, err := toFactionStat(s)
		if err != nil {
			return nil, fmt.Errorf("exclude_categories[%d]: %w", i, err)
		}
		excludeCategories[i] = stat
	}
	return &domain.TransportProfile{
		MaxHex:            r.MaxHex,
		CoinCost:          r.CoinCost,
		CargoTypes:        cargoTypes,
		MaxCargo:          r.MaxCargo,
		ExcludeCategories: excludeCategories,
	}, nil
}

// ---------------------------------------------------------------------------
// Drift Costs
// ---------------------------------------------------------------------------

type driftCostFile struct {
	DriftCosts []int `toml:"drift_costs"`
}

func loadDriftCosts(path string) ([]int, error) {
	var f driftCostFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	return f.DriftCosts, nil
}

// ---------------------------------------------------------------------------
// Tags
// ---------------------------------------------------------------------------

type tagFile struct {
	Tags map[string]tagRecord `toml:"tags"`
}

type tagRecord struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}

func loadTags(path string) (map[string]*domain.Tag, error) {
	var f tagFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Tag, len(f.Tags))
	for _, r := range f.Tags {
		result[r.ID] = &domain.Tag{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Effect:      r.Effect,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Goals
// ---------------------------------------------------------------------------

type goalFile struct {
	Goals map[string]goalRecord `toml:"goals"`
}

type goalRecord struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Difficulty  string `toml:"difficulty"`
}

func loadGoals(path string) (map[string]*domain.Goal, error) {
	var f goalFile
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(f.Goals))
	for _, r := range f.Goals {
		result[r.ID] = &domain.Goal{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Difficulty:  r.Difficulty,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func toFactionStat(s string) (domain.FactionStat, error) {
	switch s {
	case "Force":
		return domain.StatForce, nil
	case "Cunning":
		return domain.StatCunning, nil
	case "Wealth":
		return domain.StatWealth, nil
	default:
		return "", fmt.Errorf("unknown faction stat: %q", s)
	}
}

func toAssetType(s string) (domain.AssetType, error) {
	switch s {
	case "Military Unit":
		return domain.TypeMilitaryUnit, nil
	case "Special Forces":
		return domain.TypeSpecialForces, nil
	case "Facility":
		return domain.TypeFacility, nil
	case "Starship":
		return domain.TypeStarship, nil
	case "Tactic":
		return domain.TypeTactic, nil
	case "Logistics Facility":
		return domain.TypeLogisticsFacility, nil
	case "Special":
		return domain.TypeSpecial, nil
	default:
		return "", fmt.Errorf("unknown asset type: %q", s)
	}
}

func toAssetFlags(ss []string) ([]domain.AssetFlag, error) {
	flags := make([]domain.AssetFlag, 0, len(ss))
	for _, s := range ss {
		switch s {
		case "P":
			flags = append(flags, domain.FlagPermission)
		case "A":
			flags = append(flags, domain.FlagAction)
		case "S":
			flags = append(flags, domain.FlagSpecial)
		default:
			return nil, fmt.Errorf("unknown asset flag: %q", s)
		}
	}
	return flags, nil
}
