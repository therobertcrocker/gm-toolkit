package testharness

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ScriptedCollector is an InputCollector whose responses are pre-set by tests.
// Each method delegates to its *Fn override if non-nil; otherwise it returns a
// safe default (zero value, or — for SelectAction — the first available
// action, or nil if the slice is empty).
//
// AwaitCheckpoint always returns nil so headless tests do not block.
type ScriptedCollector struct {
	SelectAssetFn                func([]*domain.Asset, *loader.Rulebook) (*domain.Asset, error)
	SelectRepairOrdersFn         func(*domain.Faction, []*domain.Asset, *loader.Rulebook) ([]engine.RepairOrder, error)
	SelectBuyOrderFn             func([]string, []*domain.AssetDefinition) (engine.BuyOrder, error)
	SelectRefitOrderFn           func([]engine.RefitOption, *loader.Rulebook) (engine.RefitOrder, error)
	SelectAttackersFn            func([]*domain.Asset, *loader.Rulebook) ([]*domain.Asset, error)
	SelectDefenderFn             func(*domain.Asset, []*domain.Asset, *loader.Rulebook) (*domain.Asset, error)
	ConfirmRedirectToBaseFn      func(*domain.Faction, *domain.Base, int) (bool, error)
	SelectExpandInfluenceOrderFn func(*domain.Faction, *state.FactionState) (engine.ExpandInfluenceOrder, error)
	ConfirmRivalFreeAttackFn     func(*domain.Faction, int, int) (bool, error)
	SelectBaseAttackersFn        func(*domain.Faction, []*domain.Asset, *loader.Rulebook) ([]*domain.Asset, error)
	SelectAbilityAssetsFn        func(*domain.Faction, []*domain.Asset, *loader.Rulebook) ([]*domain.Asset, error)
	SelectMoveDestinationFn      func(*domain.Asset, []string) (string, error)
	SelectFactionTestTargetFn    func(*domain.Asset, domain.AbilityEffectType, []*domain.Faction) (*domain.Faction, error)
	ConfirmAbilityAppliedFn      func(*domain.Asset, *domain.AssetDefinition) (bool, error)
	SelectBribeTargetFn          func(*domain.Faction, *state.FactionState) (*domain.Base, int, error)
	SelectSeizeTargetFn          func(*domain.Faction, *state.FactionState) (string, error)
	SelectActionFn               func(*domain.Faction, []engine.Action) (engine.Action, error)
}

func (c *ScriptedCollector) SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error) {
	if c.SelectAssetFn != nil {
		return c.SelectAssetFn(assets, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *loader.Rulebook) ([]engine.RepairOrder, error) {
	if c.SelectRepairOrdersFn != nil {
		return c.SelectRepairOrdersFn(faction, damaged, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectBuyOrder(worlds []string, purchasable []*domain.AssetDefinition) (engine.BuyOrder, error) {
	if c.SelectBuyOrderFn != nil {
		return c.SelectBuyOrderFn(worlds, purchasable)
	}
	return engine.BuyOrder{}, nil
}

func (c *ScriptedCollector) SelectRefitOrder(options []engine.RefitOption, rulebook *loader.Rulebook) (engine.RefitOrder, error) {
	if c.SelectRefitOrderFn != nil {
		return c.SelectRefitOrderFn(options, rulebook)
	}
	return engine.RefitOrder{}, nil
}

func (c *ScriptedCollector) SelectAttackers(eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error) {
	if c.SelectAttackersFn != nil {
		return c.SelectAttackersFn(eligible, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error) {
	if c.SelectDefenderFn != nil {
		return c.SelectDefenderFn(attacker, eligible, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	if c.ConfirmRedirectToBaseFn != nil {
		return c.ConfirmRedirectToBaseFn(defenderFaction, base, damage)
	}
	return false, nil
}

func (c *ScriptedCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState) (engine.ExpandInfluenceOrder, error) {
	if c.SelectExpandInfluenceOrderFn != nil {
		return c.SelectExpandInfluenceOrderFn(faction, factionState)
	}
	return engine.ExpandInfluenceOrder{}, nil
}

func (c *ScriptedCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	if c.ConfirmRivalFreeAttackFn != nil {
		return c.ConfirmRivalFreeAttackFn(rival, rivalRoll, factionRoll)
	}
	return false, nil
}

func (c *ScriptedCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error) {
	if c.SelectBaseAttackersFn != nil {
		return c.SelectBaseAttackersFn(rival, eligible, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *loader.Rulebook) ([]*domain.Asset, error) {
	if c.SelectAbilityAssetsFn != nil {
		return c.SelectAbilityAssetsFn(faction, candidates, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error) {
	if c.SelectMoveDestinationFn != nil {
		return c.SelectMoveDestinationFn(asset, worlds)
	}
	return "", nil
}

func (c *ScriptedCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	if c.SelectFactionTestTargetFn != nil {
		return c.SelectFactionTestTargetFn(asset, effect, candidates)
	}
	return nil, nil
}

func (c *ScriptedCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	if c.ConfirmAbilityAppliedFn != nil {
		return c.ConfirmAbilityAppliedFn(asset, def)
	}
	return true, nil
}

func (c *ScriptedCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	if c.SelectBribeTargetFn != nil {
		return c.SelectBribeTargetFn(faction, factionState)
	}
	return nil, 0, nil
}

func (c *ScriptedCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	if c.SelectSeizeTargetFn != nil {
		return c.SelectSeizeTargetFn(faction, factionState)
	}
	return "", nil
}

func (c *ScriptedCollector) SelectAction(faction *domain.Faction, available []engine.Action) (engine.Action, error) {
	if c.SelectActionFn != nil {
		return c.SelectActionFn(faction, available)
	}
	if len(available) == 0 {
		return nil, nil
	}
	return available[0], nil
}

func (c *ScriptedCollector) AwaitCheckpoint(_ string) error { return nil }
