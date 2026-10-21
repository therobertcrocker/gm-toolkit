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
	// Damage is nil for a special attack, which hits for no damage and leaves
	// its effect to the asset's own_attack_hit entries.
	Damage *DiceRoll
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
	EffectDisableAsset     AbilityEffectType = "disable_asset"
	EffectBlockAttack      AbilityEffectType = "block_attack"
	EffectLevyCoin         AbilityEffectType = "levy_coin"
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

// RelocateCostMode picks how a relocate charges CoinCost: once per moved asset, or once
// for the whole relocation (Transit Web: "for 1 Coin, any number of assets").
type RelocateCostMode string

const (
	RelocateCostPerAsset RelocateCostMode = "per_asset"
	RelocateCostFlat     RelocateCostMode = "flat"
)

// RelocateParams configures a relocate step. Scope picks what moves (the acting asset vs
// other friendly assets); Range picks hex-radius vs region reach; MaxHex is the radius for
// RangeHexes (ignored for RangeRegion); CoinCost is charged per CoinCostMode; AssetType
// filters "other" candidates (empty = any), Categories narrows them by asset category
// (empty = any) and ExcludeTypes drops types; MaxAssets caps the count moved (0 = unbounded).
type RelocateParams struct {
	Scope        RelocateScope
	Range        RelocateRange
	MaxHex       int
	CoinCost     int
	CoinCostMode RelocateCostMode
	AssetType    AssetType
	Categories   []FactionStat
	ExcludeTypes []AssetType
	MaxAssets    int
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
	TargetAsset   TargetKind = "asset"
	// deliberate: TargetWorld is dormant vocab, not dead — no asset targets a world yet.
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
	Assets   []*Asset
	// deliberate: Worlds is dormant, not dead. See TargetWorld.
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
	Effects     []EffectDefinition
}

// HalfPrice is half the listed cost, rounded down, but at least 1 Coin: the price
// to keep or bring back an asset (Marketers, the turn-start lift). A sale does not
// use it; a sale is worth Cost / 2 with no floor.
func (def *AssetDefinition) HalfPrice() int {
	return max(1, def.Cost/2)
}

type ConditionKind string

const (
	ConditionNoAbility ConditionKind = "no_ability"
	ConditionDisabled  ConditionKind = "disabled"
	ConditionNoAttack  ConditionKind = "no_attack"
)

// Condition is a lasting restriction one asset puts on another. The code that
// reads a kind decides what it blocks. ExpiresAtTurnOf names the faction whose
// turn start clears it; empty means the kind ends some other way.
type Condition struct {
	Kind            ConditionKind `toml:"kind" json:"kind"`
	SourceAssetID   string        `toml:"source_asset_id" json:"source_asset_id"`
	ExpiresAtTurnOf string        `toml:"expires_at_turn_of,omitempty" json:"expires_at_turn_of,omitempty"`
}

// Is reports whether condition and other are one condition: the same kind from
// the same source.
func (condition Condition) Is(other Condition) bool {
	return condition.Kind == other.Kind && condition.SourceAssetID == other.SourceAssetID
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
	Conditions   []Condition    `toml:"conditions,omitempty"`
}

func (asset *Asset) HasCondition(kind ConditionKind) bool {
	return slices.ContainsFunc(asset.Conditions, func(condition Condition) bool { return condition.Kind == kind })
}

// Active reports whether the asset can act or use its special ability now: an
// asset is not usable until the turn after purchase, while unmaintained, once
// destroyed, or while disabled.
func (asset *Asset) Active() bool {
	return asset.Targetable() && !asset.HasCondition(ConditionDisabled)
}

// Targetable reports whether a rival may pick the asset as a defender. It is
// Active without the conditions that stop use only: a disabled asset is useless,
// not untouchable (D-38).
func (asset *Asset) Targetable() bool {
	return asset.Ready && asset.Maintained && asset.CurrentHP > 0
}

// Clone deep-copies an asset, including its CurrentOrder pointer and that order's
// slices, so mutations applied to the copy never reach the original.
func (asset *Asset) Clone() *Asset {
	clone := *asset
	clone.Conditions = slices.Clone(asset.Conditions)
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

func NextAssetID(factions map[string]*Faction, faction *Faction, definitionID string) string {
	prefix := fmt.Sprintf("%s-%s-", faction.ID, definitionID)
	// deliberate: scan the owner on its own, not only through factions, because a map without the owner would reuse a live ID.
	highest := highestSuffix(faction, prefix)
	// deliberate: scan every faction, not only the owner, because a captured asset keeps its ID in its new owner's map (D-9).
	for _, holder := range factions {
		highest = max(highest, highestSuffix(holder, prefix))
	}
	return fmt.Sprintf("%s%d", prefix, highest+1)
}

func highestSuffix(faction *Faction, prefix string) int {
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
	return highest
}

// NewPurchasedAsset builds a freshly-bought asset for a faction: a fresh ID, full HP, not
// yet Ready (activates next turn), and Maintained. Shared by the Buy Asset action and the
// purchase_enabler ability so the two purchase paths can't drift on the asset shape.
func NewPurchasedAsset(factions map[string]*Faction, faction *Faction, def *AssetDefinition, worldID string) Asset {
	return Asset{
		ID:           NextAssetID(factions, faction, def.ID),
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
