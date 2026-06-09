package adapter

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// actionCollector is the real, channel-backed action.Collector. Each method
// either forwards engine-supplied candidates or derives them (Decision 3), packs
// a payload, and blocks on the shared ask round-trip. The recoverable
// action.ErrActionUnavailable sentinel remains a classified backstop, but with all
// 12 actions wired no method returns it; the SelectAction disable-set is the guard.
type actionCollector struct{ adapter *Adapter }

func newActionCollector(a *Adapter) *actionCollector { return &actionCollector{adapter: a} }

// currentFaction returns the faction whose turn is being processed, or nil.
// Safe from a collector method: the engine goroutine is parked on this ask's
// reply, so factionState is quiescent.
func (a *Adapter) currentFaction() *domain.Faction {
	turn := a.factionState.CurrentTurn
	if turn == nil || turn.CurrentIndex >= len(turn.FactionOrder) {
		return nil
	}
	return a.factionState.Factions[turn.FactionOrder[turn.CurrentIndex]]
}

// ownerNames maps each asset's owning faction ID to its display name, for
// labels on prompts that mix factions' assets (defender/base-attacker picks).
func (a *Adapter) ownerNames(assets []*domain.Asset) map[string]string {
	names := make(map[string]string)
	for _, asset := range assets {
		if _, seen := names[asset.OwnerID]; seen {
			continue
		}
		if faction := a.factionState.Factions[asset.OwnerID]; faction != nil {
			names[asset.OwnerID] = faction.Name
		}
	}
	return names
}

// --- hooks.Collector (no error return; Esc/cancel degrades to the legal no-op) ---

func (a *actionCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	raw, err := a.adapter.ask(AskSelectModifiers, nil, SelectModifiersPayload{Offers: offers})
	if err != nil {
		return nil // Esc/cancel -> no modifiers applied
	}
	chosen, _ := raw.([]*hooks.ModifierOffer)
	out := make([]hooks.ModifierOffer, len(chosen))
	for i, offer := range chosen {
		out[i] = *offer
	}
	return out
}

func (a *actionCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	raw, err := a.adapter.ask(AskConfirmReroll, nil, ConfirmRerollPayload{Directive: directive})
	if err != nil {
		return false // Esc/cancel -> skip the reroll
	}
	confirmed, _ := raw.(bool)
	return confirmed
}

// --- action.Collector: shared (real in foundation) ---

func (a *actionCollector) SelectAsset(assets []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAsset, nil, SelectAssetPayload{Assets: assets})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Asset)
	return chosen, nil
}

// --- action.Collector: action-unique (recoverable-stubbed until each action's commit) ---

func (a *actionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
		Asset:      asset,
		Effect:     effect,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Faction)
	return chosen, nil
}
func (a *actionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
	targets := make([]RepairTarget, 0, len(damaged))
	for _, asset := range damaged {
		def, ok := rb.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		healHP := statRating(faction, def.Category)
		if healHP <= 0 {
			continue
		}
		missing := def.HP - asset.CurrentHP
		targets = append(targets, RepairTarget{
			Asset:    asset,
			HealHP:   healHP,
			MaxSteps: (missing + healHP - 1) / healHP, // ceil(missing/healHP)
			Missing:  missing,
		})
	}
	raw, err := a.adapter.ask(AskSelectRepairOrders, faction, SelectRepairOrdersPayload{Targets: targets})
	if err != nil {
		return nil, err
	}
	orders, _ := raw.([]action.RepairOrder)
	return orders, nil
}

// statRating returns the faction's attribute rating for an asset category,
// mirroring actions.statScore (the engine's repair/heal math). Kept adapter-side
// so the engine actions package is untouched (Decision 3).
func statRating(faction *domain.Faction, stat domain.FactionStat) int {
	switch stat {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}
func (a *actionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	coin := 0
	if faction := a.adapter.currentFaction(); faction != nil {
		coin = faction.Coin
	}
	names := make(map[string]string, len(purchasablePerWorld))
	for worldID := range purchasablePerWorld {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectBuyOrder, nil, SelectBuyOrderPayload{
		PurchasableByWorld: purchasablePerWorld,
		WorldNames:         names,
		Coin:               coin,
	})
	if err != nil {
		return action.BuyOrder{}, err
	}
	order, _ := raw.(action.BuyOrder)
	return order, nil
}
func (a *actionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
	raw, err := a.adapter.ask(AskSelectRefitOrder, nil, SelectRefitOrderPayload{Options: options})
	if err != nil {
		return action.RefitOrder{}, err
	}
	order, _ := raw.(action.RefitOrder)
	return order, nil
}
func (a *actionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAttackers, nil, SelectAttackersPayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectDefender, nil, SelectDefenderPayload{
		Attacker:   attacker,
		Eligible:   eligible,
		OwnerNames: a.adapter.ownerNames(eligible),
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmRedirectToBase, nil, ConfirmRedirectToBasePayload{
		DefenderFaction: defenderFaction,
		Base:            base,
		Damage:          damage,
	})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
func (a *actionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	names := make(map[string]string, len(eligibleNewBaseWorlds))
	for _, worldID := range eligibleNewBaseWorlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectExpandInfluenceOrder, faction, SelectExpandInfluenceOrderPayload{
		NewBaseWorlds:  eligibleNewBaseWorlds,
		WorldNames:     names,
		ReinforceBases: deriveReinforceTargets(faction),
		Coin:           faction.Coin,
	})
	if err != nil {
		return action.ExpandInfluenceOrder{}, err
	}
	order, _ := raw.(action.ExpandInfluenceOrder)
	return order, nil
}
func (a *actionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmRivalFreeAttack, nil, ConfirmRivalFreeAttackPayload{
		Rival:       rival,
		RivalRoll:   rivalRoll,
		FactionRoll: factionRoll,
	})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
func (a *actionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectBaseAttackers, nil, SelectBaseAttackersPayload{
		Rival:    rival,
		Eligible: eligible,
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}

// deriveReinforceTargets unions damagedNonHomeworldBases (heal-eligible) and
// growableNonHomeworldBases (max-eligible) from expand_influence.go into per-base
// caps. The two filters are independent: a base may be heal-eligible, max-eligible,
// both, or neither (excluded).
func deriveReinforceTargets(faction *domain.Faction) []ReinforceTarget {
	var targets []ReinforceTarget
	for _, base := range faction.Bases {
		if base.IsHomeworld {
			continue
		}
		healCap := base.EffectiveMaxHP(faction) - base.CurrentHP
		maxCap := faction.MaxHP - base.MaxHP
		canHeal := healCap > 0
		canMax := maxCap > 0
		if !canHeal && !canMax {
			continue
		}
		targets = append(targets, ReinforceTarget{
			Base:    base,
			CanHeal: canHeal,
			HealCap: healCap,
			CanMax:  canMax,
			MaxCap:  maxCap,
		})
	}
	return targets
}
func (a *actionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAbilityAssets, faction, SelectAbilityAssetsPayload{Candidates: candidates})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmAbilityApplied, nil, ConfirmAbilityAppliedPayload{Asset: asset, Def: def})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
func (a *actionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	bases, ownerNames := deriveBribeTargets(factionState, faction.ID)
	raw, err := a.adapter.ask(AskSelectBribeTarget, faction, SelectBribeTargetPayload{
		Bases:      bases,
		OwnerNames: ownerNames,
		Coin:       faction.Coin,
	})
	if err != nil {
		return nil, 0, err
	}
	reply, _ := raw.(BribeReply)
	return reply.Base, reply.Amount, nil
}

// deriveBribeTargets returns every base owned by a faction other than the actor
// (the locked candidate policy), plus an ownerID->name map for labels. The
// engine applies no filter of its own (bribe.go), so the adapter defines the set.
func deriveBribeTargets(factionState *state.FactionState, factionID string) ([]*domain.Base, map[string]string) {
	var bases []*domain.Base
	names := make(map[string]string)
	for id, faction := range factionState.Factions {
		if id == factionID {
			continue
		}
		names[id] = faction.Name
		bases = append(bases, faction.Bases...)
	}
	sort.Slice(bases, func(i, j int) bool { return bases[i].ID < bases[j].ID })
	return bases, names
}
func (a *actionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	worlds := deriveContestedWorlds(faction, a.adapter.engine.World.Index)
	names := make(map[string]string, len(worlds))
	for _, worldID := range worlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectSeizeTarget, faction, SelectSeizeTargetPayload{
		Worlds:     worlds,
		WorldNames: names,
	})
	if err != nil {
		return "", err
	}
	selected, _ := raw.(string)
	return selected, nil
}

// deriveContestedWorlds ports seizePlanetTargetWorlds (seize_planet.go): worlds
// where the faction has at least one unstealthed asset and some rival also has an
// unstealthed asset there. Reads the engine's world Index — quiescent because the
// engine goroutine is parked on this ask's reply, and the same source the engine's
// own Validate consulted to enable the action.
func deriveContestedWorlds(faction *domain.Faction, index *world.Index) []string {
	factionWorlds := map[string]struct{}{}
	for _, asset := range faction.Assets {
		if !asset.Stealthy {
			factionWorlds[asset.Location.WorldID] = struct{}{}
		}
	}
	contested := make([]string, 0, len(factionWorlds))
	for worldID := range factionWorlds {
		for _, asset := range index.AssetsByLocation[worldID] {
			if asset.OwnerID != faction.ID && !asset.Stealthy {
				contested = append(contested, worldID)
				break
			}
		}
	}
	sort.Strings(contested)
	return contested
}
func (a *actionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	worlds := deriveHomeworldTargets(faction, a.adapter.engine.World, a.adapter.engine.Rulebook)
	names := make(map[string]string, len(worlds))
	for _, worldID := range worlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectChangeHomeworldTarget, faction, SelectChangeHomeworldTargetPayload{
		Worlds:     worlds,
		WorldNames: names,
	})
	if err != nil {
		return "", err
	}
	selected, _ := raw.(string)
	return selected, nil
}

// deriveHomeworldTargets ports changeHomeworldTargets (change_homeworld.go):
// non-homeworld base worlds reachable from the current homeworld at the default
// drift rating (DriftCost(3)). Excludes the current homeworld and any world the
// router can't path to.
func deriveHomeworldTargets(faction *domain.Faction, worldEngine *world.WorldEngine, rb *rulebook.Rulebook) []string {
	crossingCost := rb.DriftCost(3)
	var targets []string
	for _, base := range faction.Bases {
		if base.Location.WorldID == faction.Homeworld.WorldID {
			continue
		}
		target, ok := worldEngine.Location(base.Location.WorldID)
		if !ok {
			continue
		}
		if _, err := worldEngine.Distance(faction.Homeworld.RegionHex, target.RegionHex(), crossingCost); err != nil {
			continue
		}
		targets = append(targets, base.Location.WorldID)
	}
	sort.Strings(targets)
	return targets
}

var _ action.Collector = (*actionCollector)(nil)
