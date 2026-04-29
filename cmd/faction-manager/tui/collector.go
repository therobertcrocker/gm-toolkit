package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type TUICollector struct {
	selectedAsset        *domain.Asset
	buyOrder             engine.BuyOrder
	refitOrder           engine.RefitOrder
	repairOrders         []engine.RepairOrder
	attackers            []*domain.Asset
	defenders            map[string]*domain.Asset // attacker ID → defender
	expandInfluenceOrder engine.ExpandInfluenceOrder
	baseAttackers        []*domain.Asset
	abilityAssets        []*domain.Asset
	bribeBase            *domain.Base
	bribeAmount          int
	seizeWorld           string
	eventCh              chan tea.Msg
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

// AbilityMoveMsg is sent by TUICollector.SelectMoveDestination to ask the GM
// where to relocate the asset. Called from the ability resolution goroutine.
type AbilityMoveMsg struct {
	Asset      *domain.Asset
	Worlds     []string
	ResponseCh chan string
}

// AbilityFactionTestMsg is sent by TUICollector.SelectFactionTestTarget to ask
// the GM which faction to target with a faction test ability step.
type AbilityFactionTestMsg struct {
	Asset      *domain.Asset
	Effect     domain.AbilityEffectType
	Candidates []*domain.Faction
	ResponseCh chan *domain.Faction
}

// AbilityConfirmMsg is sent by TUICollector.ConfirmAbilityApplied to ask the
// GM to confirm a GM-adjudicated ability was applied.
type AbilityConfirmMsg struct {
	Asset      *domain.Asset
	Def        *domain.AssetDefinition
	ResponseCh chan bool
}

func (c *TUICollector) SelectAbilityAssets(_ *domain.Faction, _ []*domain.Asset, _ *loader.Rulebook) ([]*domain.Asset, error) {
	return c.abilityAssets, nil
}

func (c *TUICollector) SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error) {
	responseCh := make(chan string, 1)
	c.eventCh <- AbilityMoveMsg{Asset: asset, Worlds: worlds, ResponseCh: responseCh}
	return <-responseCh, nil
}

func (c *TUICollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	responseCh := make(chan *domain.Faction, 1)
	c.eventCh <- AbilityFactionTestMsg{Asset: asset, Effect: effect, Candidates: candidates, ResponseCh: responseCh}
	return <-responseCh, nil
}

func (c *TUICollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	responseCh := make(chan bool, 1)
	c.eventCh <- AbilityConfirmMsg{Asset: asset, Def: def, ResponseCh: responseCh}
	return <-responseCh, nil
}

func (c *TUICollector) SelectBribeTarget(_ *domain.Faction, _ *state.FactionState) (*domain.Base, int, error) {
	return c.bribeBase, c.bribeAmount, nil
}

func (c *TUICollector) SelectSiezeTarget(_ *domain.Faction, _ *state.FactionState) (string, error) {
	return c.seizeWorld, nil
}
