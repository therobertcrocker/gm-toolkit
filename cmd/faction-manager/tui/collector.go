package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type TUICollector struct {
	selectedAsset          *domain.Asset
	buyOrder               engine.BuyOrder
	refitOrder             engine.RefitOrder
	repairOrders           []engine.RepairOrder
	attackers              []*domain.Asset
	defenders              map[string]*domain.Asset // attacker ID → defender
	expandInfluenceOrder   engine.ExpandInfluenceOrder
	baseAttackers          []*domain.Asset
	eventCh                chan tea.Msg // used by ConfirmRedirectToBase and ConfirmRivalFreeAttack
}

// ExpandInfluenceRivalMsg is sent from the Expand Influence resolution goroutine
// to the TUI event loop when a rival ties or beats the contested roll.
type ExpandInfluenceRivalMsg struct {
	Rival       *domain.Faction
	RivalRoll   int
	FactionRoll int
	ResponseCh  chan bool
}

// ExpandInfluenceBaseAttackersMsg is sent when a rival has confirmed their free
// attack and the TUI must collect which of their assets attack the new base.
type ExpandInfluenceBaseAttackersMsg struct {
	Rival      *domain.Faction
	Eligible   []*domain.Asset
	ResponseCh chan []*domain.Asset
}

func (c *TUICollector) SelectAsset(_ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	return c.selectedAsset, nil
}

func (c *TUICollector) SelectRepairOrders(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]engine.RepairOrder, error) {
	return c.repairOrders, nil
}

func (c *TUICollector) SelectBuyOrder(_ []string, _ []*domain.AssetDefinition) (engine.BuyOrder, error) {
	return c.buyOrder, nil
}

func (c *TUICollector) SelectRefitOrder(_ []engine.RefitOption, _ *loader.Rulebook) (engine.RefitOrder, error) {
	return c.refitOrder, nil
}

func (c *TUICollector) SelectAttackers(_ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	return c.attackers, nil
}

func (c *TUICollector) SelectDefender(attacker *domain.Asset, _ []*domain.Asset, _ *loader.Rulebook) (*domain.Asset, error) {
	return c.defenders[attacker.ID], nil
}

// ConfirmRedirectToBase sends a prompt to the TUI event loop and blocks until
// the user answers. Called from the attack resolution goroutine.
func (c *TUICollector) ConfirmRedirectToBase(defFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	responseCh := make(chan bool, 1)
	c.eventCh <- AttackRedirectMsg{
		DefFaction: defFaction,
		Base:       base,
		Damage:     damage,
		ResponseCh: responseCh,
	}
	return <-responseCh, nil
}

func (c *TUICollector) SelectExpandInfluenceOrder(_ *domain.Faction, _ *state.FactionState) (engine.ExpandInfluenceOrder, error) {
	return c.expandInfluenceOrder, nil
}

// ConfirmRivalFreeAttack sends a prompt to the TUI event loop and blocks until
// the GM answers. Called from the Expand Influence resolution goroutine.
func (c *TUICollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	responseCh := make(chan bool, 1)
	c.eventCh <- ExpandInfluenceRivalMsg{
		Rival:       rival,
		RivalRoll:   rivalRoll,
		FactionRoll: factionRoll,
		ResponseCh:  responseCh,
	}
	return <-responseCh, nil
}

// SelectBaseAttackers sends a prompt to the TUI event loop and blocks until the
// GM selects which rival assets attack the new Base. Called from the Expand
// Influence resolution goroutine after a rival confirms their free attack.
func (c *TUICollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	responseCh := make(chan []*domain.Asset, 1)
	c.eventCh <- ExpandInfluenceBaseAttackersMsg{
		Rival:      rival,
		Eligible:   eligible,
		ResponseCh: responseCh,
	}
	return <-responseCh, nil
}
