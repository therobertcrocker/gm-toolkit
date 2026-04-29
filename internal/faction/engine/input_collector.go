package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// InputCollector abstracts input collection for action resolution. The GM
// implementation uses interactive prompts; an AI implementation uses
// goal-driven selection logic. Actions call only the methods they need.
type InputCollector interface {
	SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
	SelectRepairOrders(faction *domain.Faction, damagedAssets []*domain.Asset, rulebook *loader.Rulebook) ([]RepairOrder, error)
	SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (BuyOrder, error)
	SelectRefitOrder(options []RefitOption, rulebook *loader.Rulebook) (RefitOrder, error)
	// Attack
	SelectAttackers(eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
	SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
	ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error)
	// Expand Influence
	SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState) (ExpandInfluenceOrder, error)
	ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error)
	SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
	// Use Asset Ability
	SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error)
	SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
	SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
	ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error)
	// Bribe
	SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error)
	// Seize Planet
	SelectSiezeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error)
}

// RepairOrder describes a single asset repair instruction: which asset and how
// many heals to apply. Cost escalates per heal on the same asset (1 Coin for
// the first, +1 for each subsequent).
type RepairOrder struct {
	Asset     *domain.Asset
	HealCount int
}

// BuyOrder describes a Buy Asset instruction: the target world and the
// definition of the asset to purchase.
type BuyOrder struct {
	World      string
	Definition *domain.AssetDefinition
}

// RefitOption bundles a refittable asset with the list of valid replacement
// definitions for that specific asset.
type RefitOption struct {
	Asset        *domain.Asset
	Replacements []*domain.AssetDefinition
}

// RefitOrder describes a Refit Asset instruction: which asset to refit and
// which definition to replace it with.
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
	BaseID   string        // reinforce only
	SubMode  ReinforceMode // reinforce only
	HPAmount int
}
