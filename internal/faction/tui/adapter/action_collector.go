package adapter

import (
	"errors"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ErrActionNotImplemented is returned by every error-returning method on the
// stub action.Collector. The dry-run state is shaped so none are reachable;
// if one is called, the error is loud and immediate.
var ErrActionNotImplemented = errors.New("action.Collector: not implemented in stub")

// stubActionCollector satisfies action.Collector with loud-failure stubs.
type stubActionCollector struct{}

func NewStubActionCollector() action.Collector { return &stubActionCollector{} }

// --- hooks.Collector (embedded; non-error-returning, so panic on call) ---

func (s *stubActionCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	panic("hooks.Collector.SelectModifiers: not implemented in foundation")
}

func (s *stubActionCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	panic("hooks.Collector.ConfirmReroll: not implemented in foundation")
}

// --- action.Collector direct methods ---

func (s *stubActionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectAsset(assets []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	return action.BuyOrder{}, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
	return action.RefitOrder{}, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	return false, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	return action.ExpandInfluenceOrder{}, ErrActionNotImplemented
}
func (s *stubActionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	return false, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, ErrActionNotImplemented
}
func (s *stubActionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	return false, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	return nil, 0, ErrActionNotImplemented
}
func (s *stubActionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", ErrActionNotImplemented
}
func (s *stubActionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", ErrActionNotImplemented
}

var _ action.Collector = (*stubActionCollector)(nil)
