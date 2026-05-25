package rulebook

import (
	"fmt"
	"path/filepath"

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
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
	Effect       string `toml:"effect"`
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
		if _, err := toml.DecodeFile(path, &f); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", path, err)
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
			Damage:       *damage,
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
	}, nil
}

func convertAbility(r *abilityRecord) (*domain.AbilityDefinition, error) {
	if r == nil {
		return nil, nil
	}
	def := &domain.AbilityDefinition{}
	if r.AttackerStat != "" {
		stat, err := toFactionStat(r.AttackerStat)
		if err != nil {
			return nil, fmt.Errorf("attacker_stat: %w", err)
		}
		def.AttackerStat = stat
	}
	if r.DefenderStat != "" {
		stat, err := toFactionStat(r.DefenderStat)
		if err != nil {
			return nil, fmt.Errorf("defender_stat: %w", err)
		}
		def.DefenderStat = stat
	}
	if r.Effect != "" {
		effect, err := toAbilityEffect(r.Effect)
		if err != nil {
			return nil, err
		}
		def.Effect = effect
	}
	return def, nil
}

func toAbilityEffect(s string) (domain.AbilityEffectType, error) {
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	case "coin_drain":
		return domain.EffectCoinDrain, nil
	case "coin_steal":
		return domain.EffectCoinSteal, nil
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
