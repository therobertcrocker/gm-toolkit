package loader

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// AssetLoader loads asset definitions from a TOML file into StaticData.
type AssetLoader struct {
	data *StaticData
}

func (l *AssetLoader) Load(path string) error {
	var raw rawAssetFile
	if err := decodeTOML(path, &raw); err != nil {
		return err
	}

	l.data.Assets = make(map[string]*domain.AssetDefinition, len(raw.Assets))
	for name, r := range raw.Assets {
		def, err := convertAsset(name, r)
		if err != nil {
			return fmt.Errorf("invalid asset %q: %w", name, err)
		}
		l.data.Assets[name] = def
	}

	return nil
}

// --- raw TOML shapes ---

type rawAssetFile struct {
	Assets map[string]rawAsset `toml:"assets"`
}

type rawAsset struct {
	Category    string     `toml:"category"`
	MinRating   int        `toml:"min_rating"`
	HP          int        `toml:"hp"`
	Cost        int        `toml:"cost"`
	TechLevel   int        `toml:"tech_level"`
	Type        string     `toml:"type"`
	Attack      *rawAttack `toml:"attack"`
	Counter     string     `toml:"counter"`
	Flags       []string   `toml:"flags"`
	Description string     `toml:"description"`
}

type rawAttack struct {
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
	Damage       string `toml:"damage"`
}

// --- conversion ---

func convertAsset(name string, r rawAsset) (*domain.AssetDefinition, error) {
	def := &domain.AssetDefinition{
		Name:        name,
		Category:    domain.FactionStat(r.Category),
		MinRating:   r.MinRating,
		HP:          r.HP,
		Cost:        r.Cost,
		TechLevel:   r.TechLevel,
		Type:        domain.AssetType(r.Type),
		Description: r.Description,
	}

	for _, f := range r.Flags {
		def.Flags = append(def.Flags, domain.AssetFlag(f))
	}

	if r.Attack != nil {
		damage, err := parseDice(r.Attack.Damage)
		if err != nil {
			return nil, fmt.Errorf("invalid attack damage: %w", err)
		}
		def.Attack = &domain.AttackProfile{
			AttackerStat: domain.FactionStat(r.Attack.AttackerStat),
			DefenderStat: domain.FactionStat(r.Attack.DefenderStat),
			Damage:       damage,
		}
	}

	if r.Counter != "" {
		counter, err := parseDice(r.Counter)
		if err != nil {
			return nil, fmt.Errorf("invalid counter damage: %w", err)
		}
		def.Counter = &counter
	}

	return def, nil
}
