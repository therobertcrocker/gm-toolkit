package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type AssetType string

const (
	TypeMilitaryUnit      AssetType = "Military Unit"
	TypeSpecialForces     AssetType = "Special Forces"
	TypeFacility          AssetType = "Facility"
	TypeStarship          AssetType = "Starship"
	TypeTactic            AssetType = "Tactic"
	TypeLogisticsFacility AssetType = "Logistics Facility"
	TypeSpecial           AssetType = "Special"
)

type AssetFlag string

const (
	FlagPermission AssetFlag = "P"
	FlagAction     AssetFlag = "A"
	FlagSpecial    AssetFlag = "S"
)

type DiceRoll struct {
	NumDice  int
	Sides    int
	Modifier int
}

type AttackProfile struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
	Damage       DiceRoll
}

type AbilityStepType string

const (
	AbilityStepMovement    AbilityStepType = "movement"
	AbilityStepFactionTest AbilityStepType = "faction_test"
)

type AbilityEffectType string

const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
	EffectCoinDrain     AbilityEffectType = "coin_drain"
	EffectCoinSteal     AbilityEffectType = "coin_steal"
)

type AbilityStep struct {
	Type AbilityStepType

	// movement fields
	MaxHex   int
	CoinCost int

	// faction_test fields
	AttackerStat FactionStat
	DefenderStat FactionStat
	Effect       AbilityEffectType
	EffectDice   *DiceRoll
}

type AbilityDefinition struct {
	Steps []AbilityStep
}

type AssetDefinition struct {
	ID          string
	Name        string
	Category    FactionStat
	MinRating   int
	HP          int
	Cost        int
	Maintenance int
	TechLevel   int
	DriftRating int
	Type        AssetType
	Attack      *AttackProfile
	Counter     *DiceRoll
	Flags       []AssetFlag
	Description string
	Ability     *AbilityDefinition
}

type Asset struct {
	ID           string `toml:"id"`
	DefinitionID string `toml:"definition_id"`
	OwnerID      string `toml:"owner_id"`
	Location     string `toml:"location"`
	CurrentHP    int    `toml:"current_hp"`
	Stealthy     bool   `toml:"stealthy"`
	Ready        bool   `toml:"ready"`
	Maintained   bool   `toml:"maintained"`
}

func (def *AssetDefinition) HasFlag(flag AssetFlag) bool {
	for _, f := range def.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

func NextAssetID(faction *Faction, definitionID string) string {
	prefix := fmt.Sprintf("%s-%s-", faction.ID, definitionID)
	highest := 0
	for id := range faction.Assets {
		if !strings.HasPrefix(id, prefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(id, prefix))
		if err == nil && n > highest {
			highest = n
		}
	}

	return fmt.Sprintf("%s%d", prefix, highest+1)
}
