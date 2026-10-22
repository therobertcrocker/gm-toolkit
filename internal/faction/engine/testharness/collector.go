package testharness

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var _ action.Collector = (*ScriptedCollector)(nil)
var _ hooks.Collector = (*ScriptedCollector)(nil)

// ScriptedCollector satisfies PhaseCollector, action.Collector, hooks.Collector,
// and ability.Collector. Each method delegates to its *Fn override if non-nil;
// otherwise it returns a safe default (zero value, or — for SelectAction — the
// first available action, or nil if the slice is empty).
//
// AwaitCheckpoint always returns nil so headless tests do not block.
type ScriptedCollector struct {
	SelectAssetFn                 func([]*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error)
	SelectAssetToLoseFn           func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error)
	SelectRepairOrdersFn          func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]action.RepairOrder, error)
	SelectBuyOrderFn              func(map[string][]*domain.AssetDefinition) (action.BuyOrder, error)
	SelectRefitOrderFn            func([]action.RefitOption, *rulebook.Rulebook) (action.RefitOrder, error)
	SelectAttackersFn             func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error)
	SelectDefenderFn              func(*domain.Asset, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error)
	ConfirmRedirectToBaseFn       func(*domain.Faction, *domain.Base, int) (bool, error)
	ConfirmCounterstrikeFn        func(*domain.Faction, *domain.Asset, *domain.Asset) (bool, error)
	ConfirmPaymentFn              func(hooks.PaymentOffer) (bool, error)
	ConfirmSacrificeFn            func(hooks.SacrificeOffer) (bool, error)
	RequestPermissionFn           func(hooks.PermissionRequest) (bool, error)
	SelectExpandInfluenceOrderFn  func(*domain.Faction, *state.FactionState, []string) (action.ExpandInfluenceOrder, error)
	ConfirmRivalFreeAttackFn      func(*domain.Faction, int, int) (bool, error)
	SelectBaseAttackersFn         func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error)
	SelectAbilityAssetsFn         func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error)
	SelectMoveDestinationFn       func(*domain.Asset, []string) (string, error)
	SelectAbilityTargetFn         func(*domain.Asset, domain.AbilityStep, domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
	SelectRelocateOrdersFn        func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)
	ConfirmAbilityAppliedFn       func(*domain.Asset, *domain.AssetDefinition) (bool, error)
	SelectBribeTargetFn           func(*domain.Faction, *state.FactionState) (*domain.Base, int, error)
	SelectSeizeTargetFn           func(*domain.Faction, *state.FactionState) (string, error)
	SelectActionFn                func(*domain.Faction, []action.Action) (action.Action, error)
	SelectModifiersFn             func([]hooks.ModifierOffer) []hooks.ModifierOffer
	ConfirmRerollFn               func(hooks.RerollDirective) bool
	SelectStatRaiseFn             func(*domain.Faction, []domain.FactionStat) (*domain.FactionStat, error)
	SelectMovementDecisionsFn     func(*domain.Faction, []*domain.Asset) ([]world.MovementDecision, error)
	SelectTransportCargoFn        func(*domain.Asset, []*domain.Asset, *domain.TransportProfile) ([]*domain.Asset, error)
	SelectChangeHomeworldTargetFn func(*domain.Faction, *state.FactionState) (string, error)
	SelectPurchaseEnablerAssetFn  func(string, []action.PurchaseEnablerOption) (*domain.AssetDefinition, error)
}

func (c *ScriptedCollector) SelectAsset(assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
	if c.SelectAssetFn != nil {
		return c.SelectAssetFn(assets, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectAssetToLose(faction *domain.Faction, assets []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
	if c.SelectAssetToLoseFn != nil {
		return c.SelectAssetToLoseFn(faction, assets, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rulebook *rulebook.Rulebook) ([]action.RepairOrder, error) {
	if c.SelectRepairOrdersFn != nil {
		return c.SelectRepairOrdersFn(faction, damaged, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	if c.SelectBuyOrderFn != nil {
		return c.SelectBuyOrderFn(purchasablePerWorld)
	}
	return action.BuyOrder{}, nil
}

func (c *ScriptedCollector) SelectRefitOrder(options []action.RefitOption, rulebook *rulebook.Rulebook) (action.RefitOrder, error) {
	if c.SelectRefitOrderFn != nil {
		return c.SelectRefitOrderFn(options, rulebook)
	}
	return action.RefitOrder{}, nil
}

func (c *ScriptedCollector) SelectAttackers(eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error) {
	if c.SelectAttackersFn != nil {
		return c.SelectAttackersFn(eligible, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rulebook *rulebook.Rulebook) (*domain.Asset, error) {
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

func (c *ScriptedCollector) ConfirmCounterstrike(defenderFaction *domain.Faction, source *domain.Asset, attacker *domain.Asset) (bool, error) {
	if c.ConfirmCounterstrikeFn != nil {
		return c.ConfirmCounterstrikeFn(defenderFaction, source, attacker)
	}
	return false, nil
}

func (c *ScriptedCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	if c.ConfirmPaymentFn != nil {
		return c.ConfirmPaymentFn(offer)
	}
	return false, nil
}

func (c *ScriptedCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	if c.ConfirmSacrificeFn != nil {
		return c.ConfirmSacrificeFn(offer)
	}
	return false, nil
}

// RequestPermission grants by default, so a scripted buy of a P asset goes through
// unless a test scripts a refusal.
func (c *ScriptedCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	if c.RequestPermissionFn != nil {
		return c.RequestPermissionFn(request)
	}
	return true, nil
}

func (c *ScriptedCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	if c.SelectExpandInfluenceOrderFn != nil {
		return c.SelectExpandInfluenceOrderFn(faction, factionState, eligibleNewBaseWorlds)
	}
	return action.ExpandInfluenceOrder{}, nil
}

func (c *ScriptedCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	if c.ConfirmRivalFreeAttackFn != nil {
		return c.ConfirmRivalFreeAttackFn(rival, rivalRoll, factionRoll)
	}
	return false, nil
}

func (c *ScriptedCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error) {
	if c.SelectBaseAttackersFn != nil {
		return c.SelectBaseAttackersFn(rival, eligible, rulebook)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rulebook *rulebook.Rulebook) ([]*domain.Asset, error) {
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

func (c *ScriptedCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	if c.SelectAbilityTargetFn != nil {
		return c.SelectAbilityTargetFn(asset, step, candidates)
	}
	return domain.AbilityTarget{}, nil
}

func (c *ScriptedCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	if c.SelectRelocateOrdersFn != nil {
		return c.SelectRelocateOrdersFn(acting, candidates, candidateWorlds, maxAssets)
	}
	return nil, nil
}

func (c *ScriptedCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	if c.ConfirmAbilityAppliedFn != nil {
		return c.ConfirmAbilityAppliedFn(asset, def)
	}
	return true, nil
}

func (c *ScriptedCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
	if c.SelectTransportCargoFn != nil {
		return c.SelectTransportCargoFn(transport, eligibleCargo, profile)
	}
	return nil, nil
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

func (c *ScriptedCollector) SelectAction(faction *domain.Faction, available []action.Action) (action.Action, error) {
	if c.SelectActionFn != nil {
		return c.SelectActionFn(faction, available)
	}
	if len(available) == 0 {
		return nil, nil
	}
	return available[0], nil
}

func (c *ScriptedCollector) AwaitCheckpoint(_ string) error { return nil }

// SelectModifiers takes all offered modifiers by default; override with SelectModifiersFn.
func (c *ScriptedCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	if c.SelectModifiersFn != nil {
		return c.SelectModifiersFn(offers)
	}
	return offers
}

// ConfirmReroll confirms all rerolls by default; override with ConfirmRerollFn.
func (c *ScriptedCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	if c.ConfirmRerollFn != nil {
		return c.ConfirmRerollFn(directive)
	}
	return true
}

func (c *ScriptedCollector) SelectStatRaise(faction *domain.Faction, eligible []domain.FactionStat) (*domain.FactionStat, error) {
	if c.SelectStatRaiseFn != nil {
		return c.SelectStatRaiseFn(faction, eligible)
	}
	return nil, nil // default: skip raise
}

func (c *ScriptedCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
	if c.SelectMovementDecisionsFn != nil {
		return c.SelectMovementDecisionsFn(faction, eligible)
	}
	return nil, nil
}

func (c *ScriptedCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	if c.SelectChangeHomeworldTargetFn != nil {
		return c.SelectChangeHomeworldTargetFn(faction, factionState)
	}
	return "", nil
}

func (c *ScriptedCollector) SelectPurchaseEnablerAsset(world string, candidates []action.PurchaseEnablerOption) (*domain.AssetDefinition, error) {
	if c.SelectPurchaseEnablerAssetFn != nil {
		return c.SelectPurchaseEnablerAssetFn(world, candidates)
	}
	return nil, nil
}
