package action

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Collector abstracts all user-input methods that action implementations may call.
type Collector interface {
	SelectAsset(assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
	SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *rulebook.Rulebook) ([]RepairOrder, error)
	SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (BuyOrder, error)
	SelectRefitOrder(options []RefitOption, rulebook *rulebook.Rulebook) (RefitOrder, error)
	SelectAttackers(eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
	SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error)
	ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error)
	SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState) (ExpandInfluenceOrder, error)
	ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
	SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
	SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error)
	ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
	SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
	SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
}

// RepairOrder describes a single asset repair instruction.
type RepairOrder struct {
	Asset     *domain.Asset
	HealCount int
}

// BuyOrder describes a Buy Asset instruction.
type BuyOrder struct {
	World      string
	Definition *domain.AssetDefinition
}

// RefitOption bundles a refittable asset with valid replacement definitions.
type RefitOption struct {
	Asset        *domain.Asset
	Replacements []*domain.AssetDefinition
}

// RefitOrder describes a Refit Asset instruction.
type RefitOrder struct {
	OldAsset      *domain.Asset
	NewDefinition *domain.AssetDefinition
}

// ExpandMode distinguishes between placing a new Base and reinforcing an existing one.
type ExpandMode string

const (
	ExpandModeNew       ExpandMode = "new"
	ExpandModeReinforce ExpandMode = "reinforce"
)

// ReinforceMode distinguishes between healing a damaged Base and increasing its max HP.
type ReinforceMode string

const (
	ReinforceHeal ReinforceMode = "heal"
	ReinforceMax  ReinforceMode = "max"
)

// ExpandInfluenceOrder describes an Expand Influence instruction.
type ExpandInfluenceOrder struct {
	Mode     ExpandMode
	World    string
	BaseID   string
	SubMode  ReinforceMode
	HPAmount int
}
