package engine

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// checkArrivals decides, for each order the tick completes, whether it may land,
// and returns the tick's mutations with each completion replaced by what the
// arrival does (D-24, D-27). The government asked is the one ruling at arrival. It
// keeps a running total of the Coin each faction spends this tick on lockdowns and
// tolls, so two arrivals cannot spend the same Coin. It runs between the tick and
// its reactors, so the transport reactor moves the cargo to the edge with its
// carrier (D-28).
func (e *Engine) checkArrivals(faction *domain.Faction, factionState *state.FactionState, tickMutations []domain.Mutation, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	checked := make([]domain.Mutation, 0, len(tickMutations))
	spent := map[string]int{}
	for _, mutation := range tickMutations {
		completed, ok := mutation.(domain.MovementOrderCompleted)
		if !ok {
			checked = append(checked, mutation)
			continue
		}
		arrival, costs, err := e.resolveArrival(faction, factionState, completed, collector, spent)
		if err != nil {
			return nil, err
		}
		for payerID, cost := range costs {
			spent[payerID] += cost
		}
		checked = append(checked, arrival...)
	}
	return checked, nil
}

// resolveArrival decides one arriving order and returns the mutations that take
// the completion's place, with the Coin each faction spends on them. A Transport
// Lockdown on the destination comes first (Effort 5 Decision 14), then the tolls,
// then the government, asked for each asset aboard, the carrier first, then its
// cargo in manifest order, each for its own owner; the first refusal holds the
// whole order (D-28). The tolls come before the government so a faction that
// cannot pay is never asked (Effort 4 Decision 10). spent is the Coin each faction
// has spent this tick. An order that lands keeps its completion and pays its
// tolls after it.
func (e *Engine) resolveArrival(faction *domain.Faction, factionState *state.FactionState, completed domain.MovementOrderCompleted, collector hooks.PermissionCollector, spent map[string]int) ([]domain.Mutation, map[string]int, error) {
	carrier := faction.Assets[completed.AssetID]
	if carrier == nil || carrier.CurrentOrder == nil {
		return []domain.Mutation{completed}, nil, nil
	}
	order := carrier.CurrentOrder
	var profile *domain.TransportProfile
	if carrierDef := e.Rulebook.Assets[carrier.DefinitionID]; carrierDef != nil {
		profile = carrierDef.Transport
	}
	ignoresLockdown := profile != nil && profile.IgnoresLockdown
	if !order.LockdownPaid && !ignoresLockdown {
		if lockerID, lockdown := domain.LockedOutBy(factionState.Factions, order.Destination.WorldID, faction.ID); lockerID != "" {
			fee, err := e.lockdownFee(lockdown)
			if err != nil {
				return nil, nil, err
			}
			hold, cost := lockdownHold(completed, order, lockerID, fee, faction.Coin-spent[faction.ID])
			return hold, map[string]int{faction.ID: cost}, nil
		}
	}
	assets := aboard(factionState, carrier)
	var tolls []domain.Mutation
	owed := map[string]int{}
	// deliberate: the carrier's lockdown bypass covers every toll aboard, cargo included, because Robert reads the toll as shipping pressure, as a lockdown is (Effort 4 Decision 9).
	if !ignoresLockdown {
		tolls, owed = e.tollCharges(factionState, assets, order.Destination.WorldID)
	}
	for payerID, amount := range owed {
		if factionState.Factions[payerID].Coin-spent[payerID] < amount {
			return []domain.Mutation{holdAtEdge(completed, order, domain.HoldTollUnpaid)}, nil, nil
		}
	}
	// deliberate: the carrier's bypass covers its cargo, because the book lets these transports land "any one … asset" past the government's objection (D-25).
	if profile == nil || !profile.IgnoresPermission {
		for _, asset := range assets {
			def := e.Rulebook.Assets[asset.DefinitionID]
			requester := factionState.Factions[asset.OwnerID]
			if def == nil || requester == nil {
				continue
			}
			refusal, err := dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, e.Rulebook, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)
			if err != nil {
				return nil, nil, err
			}
			if refusal != nil {
				return []domain.Mutation{*refusal, holdAtEdge(completed, order, domain.HoldRefused)}, nil, nil
			}
		}
	}
	return append([]domain.Mutation{completed}, tolls...), owed, nil
}

// tollCharges returns the tolls the assets aboard owe on landing at worldID, each
// paid by the asset's current owner, and the Coin each owner owes in all.
func (e *Engine) tollCharges(factionState *state.FactionState, assets []*domain.Asset, worldID string) ([]domain.Mutation, map[string]int) {
	var charges []domain.Mutation
	owed := map[string]int{}
	for _, asset := range assets {
		tolls := dispatch.AssetArrivalTolls(e.Hooks, factionState.Factions, asset, worldID)
		if len(tolls) == 0 {
			continue
		}
		charges = append(charges, dispatch.TollCharges(asset.OwnerID, tolls)...)
		owed[asset.OwnerID] += dispatch.TollTotal(tolls)
	}
	return charges, owed
}

// lockdownFee rolls the fee on the lockdown's own lock_out entry, before the hold
// is decided: a faction that cannot pay has still rolled.
func (e *Engine) lockdownFee(lockdown *domain.Asset) (int, error) {
	if def := e.Rulebook.Assets[lockdown.DefinitionID]; def != nil {
		for _, entry := range def.Effects {
			if entry.LockOut != nil {
				return dice.Amount(entry.LockOut.Fee, e.Rand), nil
			}
		}
	}
	return 0, fmt.Errorf("arrival: lockdown %s has no lock_out fee", lockdown.ID)
}

// lockdownHold charges the order's fee and holds it a turn at the edge; a faction
// that cannot pay holds unpaid and tries again at its next tick (overview Decision
// 15). The fee names lockerID, the Lockdown's owner, as its cause. It returns the
// mutations and the Coin they spend.
func lockdownHold(completed domain.MovementOrderCompleted, order *domain.MovementOrder, lockerID string, fee, coin int) ([]domain.Mutation, int) {
	if coin < fee {
		return []domain.Mutation{holdAtEdge(completed, order, domain.HoldLockdownUnpaid)}, 0
	}
	return []domain.Mutation{
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: string(domain.SpecialLockOut), CausedByFactionID: lockerID},
		holdAtEdge(completed, order, domain.HoldLockdownPaid),
	}, fee
}

// aboard returns the carrier and the cargo its order carries, each found under its
// current owner: a capture can make a rider's owner differ from the carrier's.
func aboard(factionState *state.FactionState, carrier *domain.Asset) []*domain.Asset {
	assets := []*domain.Asset{carrier}
	for _, cargoID := range carrier.CurrentOrder.CargoAssetIDs {
		if owner := domain.AssetOwner(factionState.Factions, cargoID); owner != nil {
			assets = append(assets, owner.Assets[cargoID])
		}
	}
	return assets
}

// holdAtEdge replaces completed with a hold at the last hex before the destination
// (overview Decision 16). Path[0] is the asset's own hex, so on a two-hex path the
// hold is the in-orbit state right after issue.
func holdAtEdge(completed domain.MovementOrderCompleted, order *domain.MovementOrder, reason domain.HoldReason) domain.MovementOrderHeld {
	step := max(len(order.Path)-2, 0)
	return domain.MovementOrderHeld{
		FactionID:         completed.FactionID,
		AssetID:           completed.AssetID,
		WorldID:           order.Destination.WorldID,
		NewStepIdx:        step,
		RegionHex:         order.Path[step],
		Reason:            reason,
		Cause:             domain.CauseMovementTick,
		CausedByFactionID: completed.CausedByFactionID,
	}
}
