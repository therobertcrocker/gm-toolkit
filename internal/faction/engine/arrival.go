package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// checkArrivals decides, for each order the tick completes, whether it may land,
// and returns the tick's mutations with each held completion replaced by the hold
// (D-24, D-27). The government asked is the one ruling at arrival. It keeps a
// running total of the Coin the tick's lockdowns spend, so two arrivals cannot
// spend the same Coin. It runs between the tick and its reactors, so the transport
// reactor moves the cargo to the edge with its carrier (D-28).
func (e *Engine) checkArrivals(faction *domain.Faction, factionState *state.FactionState, tickMutations []domain.Mutation, collector hooks.PermissionCollector) ([]domain.Mutation, error) {
	checked := make([]domain.Mutation, 0, len(tickMutations))
	spent := 0
	for _, mutation := range tickMutations {
		completed, ok := mutation.(domain.MovementOrderCompleted)
		if !ok {
			checked = append(checked, mutation)
			continue
		}
		hold, cost, err := e.arrivalHold(faction, factionState, completed, collector, faction.Coin-spent)
		if err != nil {
			return nil, err
		}
		spent += cost
		if hold == nil {
			checked = append(checked, mutation)
			continue
		}
		checked = append(checked, hold...)
	}
	return checked, nil
}

// arrivalHold decides one arriving order. A Transport Lockdown on the destination
// comes first (Effort 5 Decision 14). Then the government is asked for each asset
// aboard, the carrier first, then its cargo in manifest order, each for its own
// owner; the first refusal holds the whole order (D-28). coin is what the faction
// can still spend this tick. It returns the mutations that replace the completion
// and the Coin they spend, or nil when everything aboard may land.
func (e *Engine) arrivalHold(faction *domain.Faction, factionState *state.FactionState, completed domain.MovementOrderCompleted, collector hooks.PermissionCollector, coin int) ([]domain.Mutation, int, error) {
	carrier := faction.Assets[completed.AssetID]
	if carrier == nil || carrier.CurrentOrder == nil {
		return nil, 0, nil
	}
	order := carrier.CurrentOrder
	var profile *domain.TransportProfile
	if carrierDef := e.Rulebook.Assets[carrier.DefinitionID]; carrierDef != nil {
		profile = carrierDef.Transport
	}
	ignoresLockdown := profile != nil && profile.IgnoresLockdown
	if !order.LockdownPaid && !ignoresLockdown {
		if lockerID := domain.LockedOutBy(factionState.Factions, order.Destination.WorldID, faction.ID); lockerID != "" {
			hold, cost := lockdownHold(completed, order, lockerID, e.Rand.Roll(4), coin)
			return hold, cost, nil
		}
	}
	// deliberate: the carrier's bypass covers its cargo, because the book lets these transports land "any one … asset" past the government's objection (D-25).
	if profile != nil && profile.IgnoresPermission {
		return nil, 0, nil
	}
	for _, asset := range aboard(factionState, carrier) {
		def := e.Rulebook.Assets[asset.DefinitionID]
		requester := factionState.Factions[asset.OwnerID]
		if def == nil || requester == nil {
			continue
		}
		refusal, err := dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)
		if err != nil {
			return nil, 0, err
		}
		if refusal != nil {
			return []domain.Mutation{*refusal, holdAtEdge(completed, order, domain.HoldRefused)}, 0, nil
		}
	}
	return nil, 0, nil
}

// lockdownHold charges the order's 1d4 fee and holds it a turn at the edge; a
// faction that cannot pay holds unpaid and tries again at its next tick (overview
// Decision 15). The fee names lockerID, the Lockdown's owner, as its cause. It
// returns the mutations and the Coin they spend.
func lockdownHold(completed domain.MovementOrderCompleted, order *domain.MovementOrder, lockerID string, fee, coin int) ([]domain.Mutation, int) {
	if coin < fee {
		return []domain.Mutation{holdAtEdge(completed, order, domain.HoldLockdownUnpaid)}, 0
	}
	return []domain.Mutation{
		domain.CoinDelta{FactionID: completed.FactionID, Delta: -fee, Cause: "transport_lockdown", CausedByFactionID: lockerID},
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
