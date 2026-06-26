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

type TransportProfile struct {
	MaxHex            int
	CoinCost          int
	CargoTypes        []AssetType
	MaxCargo          int
	ExcludeCategories []FactionStat
}

type AbilityEffectType string

const (
	EffectRevealStealth    AbilityEffectType = "reveal_stealth"
	EffectCoinDelta        AbilityEffectType = "coin_delta"
	EffectPurchaseDiscount AbilityEffectType = "purchase_discount"
	EffectRelocate         AbilityEffectType = "relocate"
)

type RelocateScope string

const (
	RelocateSelf  RelocateScope = "self"
	RelocateOther RelocateScope = "other"
)

// RelocateRange discriminates how far a relocate may reach (DR-1). RangeHexes uses MaxHex
// (raw hex distance from the acting asset); RangeRegion ignores MaxHex and allows any world
// in the acting asset's region.
type RelocateRange string

const (
	RangeHexes  RelocateRange = "hexes"
	RangeRegion RelocateRange = "region"
)

// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per moved asset; AssetType
// filters "other" candidates (empty = any); MaxAssets caps the count moved (0 = unbounded).
type RelocateParams struct {
	Scope     RelocateScope
	Range     RelocateRange
	MaxHex    int
	CoinCost  int
	AssetType AssetType
	MaxAssets int
}

type TargetKind string

const (
	TargetNone    TargetKind = "none"
	TargetSelf    TargetKind = "self"
	TargetFaction TargetKind = "faction"
	TargetAsset   TargetKind = "asset"
	TargetWorld   TargetKind = "world"
)

type Contest struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
}

type RollTable struct {
	Die      DiceRoll
	Outcomes []AbilityOutcome
}

// AbilityDefinition is a uniform sequence of steps; a flat [ability] block parses
// to a one-step list.
type AbilityDefinition struct {
	Steps []AbilityStep
}

type AbilityStep struct {
	Effect   AbilityEffectType
	Target   TargetKind
	Contest  *Contest
	Roll     *RollTable
	Relocate *RelocateParams
}

type AbilityOutcome struct {
	From        int
	To          int
	Coin        int // signed: gain (positive) or loss (negative)
	DestroySelf bool
	Discount    int
}

// AbilityTarget is the step's resolved target, returned by the target stage and
// carried in the StepContext handed to an effect handler.
type AbilityTarget struct {
	Kind    TargetKind
	Faction *Faction
	Asset   *Asset
	World   string
}

// AbilityTargetCandidates bundles the selectable options for a target stage; the
// collector reads the slice matching the step's TargetKind.
type AbilityTargetCandidates struct {
	Factions []*Faction
	Assets   []*Asset
	Worlds   []string
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
	Speed       int
	DriftRating int
	Type        AssetType
	Attack      *AttackProfile
	Counter     *DiceRoll
	Flags       []AssetFlag
	Description string
	Ability     *AbilityDefinition
	Transport   *TransportProfile
}

type Asset struct {
	ID           string         `toml:"id"`
	DefinitionID string         `toml:"definition_id"`
	OwnerID      string         `toml:"owner_id"`
	Location     Location       `toml:"location"`
	CurrentHP    int            `toml:"current_hp"`
	Stealthy     bool           `toml:"stealthy"`
	Ready        bool           `toml:"ready"`
	Maintained   bool           `toml:"maintained"`
	CurrentOrder *MovementOrder `toml:"current_order,omitempty"`
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
