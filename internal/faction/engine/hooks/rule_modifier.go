package hooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// MovementAbility is a free relocation a granter offers its faction this turn.
// Source is the granting asset; range is measured from it.
type MovementAbility struct {
	Source   *domain.Asset
	Relocate domain.RelocateParams
}

// AssetCostQuery is the state an asset-cost chain carries from hook to hook.
// Surcharges are the parts of Cost that surcharge hooks added, in chain order:
// a purchase discount floors below them without removing them, and each is
// charged as its own Coin loss, caused by the faction that imposed it. Applied
// holds the effect names whose no_stack gate has already fired.
type AssetCostQuery struct {
	Cost       int
	Surcharges []Surcharge
	Applied    map[string]bool
}

type Surcharge struct {
	FactionID string
	Amount    int
}

func (query AssetCostQuery) SurchargeTotal() int {
	total := 0
	for _, surcharge := range query.Surcharges {
		total += surcharge.Amount
	}
	return total
}

// AssetCostModifier adjusts the Coin cost of an asset purchase (Cat 4).
type AssetCostModifier interface {
	ModifyAssetCost(buyer *domain.Faction, def *domain.AssetDefinition, world string, query *AssetCostQuery)
}

// MaintenanceCostModifier adjusts the per-turn maintenance cost of an asset
// (Cat 4).
type MaintenanceCostModifier interface {
	ModifyMaintenanceCost(owner *domain.Faction, asset *domain.Asset, baseCost int) int
}

// WorldTechLevelModifier adjusts the effective tech level of a world for a
// specific faction, used by TL-gated asset purchase logic (Cat 4).
type WorldTechLevelModifier interface {
	ModifyWorldTechLevel(faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int
}

// AttackSide names which part of a fight an eligibility hook answers for.
type AttackSide string

const (
	SideAttack  AttackSide = "attack"
	SideCounter AttackSide = "counter"
	SideDefense AttackSide = "defense"
)

// AttackEligibilityQuery asks whether Attacker may deal damage to Defender.
// Attacker is the asset that strikes: on a counterattack it is the original
// defender. The definitions ride along because a Cat 4 hook has no rulebook.
type AttackEligibilityQuery struct {
	Attacker    *domain.Asset
	AttackerDef *domain.AssetDefinition
	Defender    *domain.Asset
	DefenderDef *domain.AssetDefinition
	Side        AttackSide
}

// AttackEligibilityRule reports whether one asset may fight another (Cat 4).
type AttackEligibilityRule interface {
	MayFight(query AttackEligibilityQuery) bool
}

// AssetMovementGranter reports the free movement a tag or special effect offers
// faction this turn (Cat 4). A granter asked for a faction it does not serve
// returns nothing.
type AssetMovementGranter interface {
	GrantMovementAbilities(faction *domain.Faction) []MovementAbility
}

// RepairCostQuery is the state a repair-cost chain carries from hook to hook.
// Applied holds the effect names whose no_stack gate has already fired.
type RepairCostQuery struct {
	Cost    int
	Applied map[string]bool
}

// RepairCostModifier adjusts what Repair Asset charges to repair one asset
// (Cat 4). repairer is the faction that repairs; asset and def are the asset
// repaired.
type RepairCostModifier interface {
	ModifyRepairCost(repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, query *RepairCostQuery)
}
