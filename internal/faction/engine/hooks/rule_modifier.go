package hooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// MovementAbility is a free relocation a granter offers its faction this turn.
// Source is the granting asset; range is measured from it.
type MovementAbility struct {
	Source   *domain.Asset
	Relocate domain.RelocateParams
}

// AssetCostQuery is the state an asset-cost chain carries from hook to hook.
// Surcharge is the part of Cost that surcharge hooks added, so a purchase
// discount can floor below it without removing it. Applied holds the effect
// names whose no_stack gate has already fired.
type AssetCostQuery struct {
	Cost      int
	Surcharge int
	Applied   map[string]bool
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
