package domain

import (
	"fmt"
	"slices"
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
	EffectPurchaseEnabler  AbilityEffectType = "purchase_enabler"
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

// PurchaseParams configures a purchase_enabler step (Pretech Logistics). Category is the
// asset category the ability may buy (Force); MaxTechLevel caps the required tech level of
// a buyable def, ignoring the world's own tech level (the ability is precisely what buys
// above it); CostPercent is the surcharge as a percent of the base cost, rounded up
// (150 = "half again as much").
type PurchaseParams struct {
	Category     FactionStat
	MaxTechLevel int
	CostPercent  int
}

type TargetKind string

const (
	TargetNone    TargetKind = "none"
	TargetSelf    TargetKind = "self"
	TargetFaction TargetKind = "faction"
	// deliberate: TargetAsset/TargetWorld are dormant scalar-target vocab for A.001.3
	// (Marketers/Seditionists), not dead — don't drop without that effort.
	TargetAsset TargetKind = "asset"
	TargetWorld TargetKind = "world"
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
	Purchase *PurchaseParams
}

type AbilityOutcome struct {
	From        int
	To          int
	Coin        int // signed: gain (positive) or loss (negative)
	DestroySelf bool
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
	// deliberate: Assets/Worlds are dormant for A.001.3 scalar targets, not dead —
	// don't drop without that effort. See TargetAsset/TargetWorld.
	Assets []*Asset
	Worlds []string
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

// Clone deep-copies an asset, including its CurrentOrder pointer and that order's
// slices, so mutations applied to the copy never reach the original.
func (asset *Asset) Clone() *Asset {
	clone := *asset
	if asset.CurrentOrder != nil {
		order := *asset.CurrentOrder
		order.Path = slices.Clone(asset.CurrentOrder.Path)
		order.CargoAssetIDs = slices.Clone(asset.CurrentOrder.CargoAssetIDs)
		clone.CurrentOrder = &order
	}
	return &clone
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

// NewPurchasedAsset builds a freshly-bought asset for a faction: a fresh ID, full HP, not
// yet Ready (activates next turn), and Maintained. Shared by the Buy Asset action and the
// purchase_enabler ability so the two purchase paths can't drift on the asset shape.
func NewPurchasedAsset(faction *Faction, def *AssetDefinition, worldID string) Asset {
	return Asset{
		ID:           NextAssetID(faction, def.ID),
		DefinitionID: def.ID,
		OwnerID:      faction.ID,
		Location:     Location{WorldID: worldID},
		CurrentHP:    def.HP,
		Ready:        false,
		Maintained:   true,
	}
}

// HasPlanetaryGovernment reports whether the faction owns a base on the world (the SWN
// P-flag permission gate). basesOnWorld is the slice for one world (a nil slice is "no
// bases" → false), so callers pass worldEngine.Index.BasesByLocation[worldID].
func HasPlanetaryGovernment(factionID string, basesOnWorld []*Base) bool {
	for _, base := range basesOnWorld {
		if base.OwnerID == factionID {
			return true
		}
	}
	return false
}

// PurchaseTerms are the per-path inputs to CanPurchase that differ between the Buy Asset
// action and the purchase_enabler ability: Cost is the price to test affordability against
// (base cost for a buy, surcharged for an enabler); TechCap is the highest required tech
// level a def may carry (the target world's tech level for a buy, the enabler's own cap for
// an enabler); HoldsGovernment is whether the faction owns a base on the target world (the
// P-flag gate). The shared checks themselves live in CanPurchase so the two paths can't drift.
type PurchaseTerms struct {
	Cost            int
	TechCap         int
	HoldsGovernment bool
}

// CanPurchase reports whether the faction may buy def under terms: it can afford Cost, meets
// the def's minimum category rating, the def's required tech level is within TechCap, and it
// holds Planetary Government when the def is P-flagged.
func CanPurchase(faction *Faction, def *AssetDefinition, terms PurchaseTerms) bool {
	switch {
	case faction.Coin < terms.Cost:
		return false
	case def.MinRating > faction.StatScore(def.Category):
		return false
	case def.TechLevel > terms.TechCap:
		return false
	case def.HasFlag(FlagPermission) && !terms.HoldsGovernment:
		return false
	}
	return true
}
