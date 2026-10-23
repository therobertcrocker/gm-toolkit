package dispatch

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// PermissionWithdrawnKey is the HookBudgets mark a Lobbyists withdrawal spends on
// the requester. Until its next bookkeeping clears it, every request it makes for
// worldID is refused with no ask (D-30, Effort 5 Decision 9). The key names the
// withdrawer, so the refusal can say whose lobbying bars the request.
func PermissionWithdrawnKey(worldID, withdrawerID string) string {
	return "permission_withdrawn:" + worldID + ":" + withdrawerID
}

// RequestPermission asks worldID's government whether requester may raise def
// there, and returns the refusal record, or nil when the purchase may go ahead. A
// definition that needs no permission is not asked about, nor is a purchase on the
// requester's homeworld (Effort 4 Decision 15). cause names the purchase path on
// the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// RequestArrivalPermission asks worldID's government whether requester may land
// asset there, at the end of a drift order or by a relocate (cause), and returns
// the refusal record, or nil when the asset may land. An asset that needs no
// permission is not asked about (AssetDefinition.NeedsArrivalPermission).
func RequestArrivalPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsArrivalPermission(asset) {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
}

// askGovernment refuses a request a withdrawal has barred this turn. It grants a
// ruler's own request at once (D-6), and one a grant hook of the requester serves
// (Popular Movement); otherwise it asks the ruler's driver, or the GM for an
// independent world (D-4). Every grant is then open to the rival Lobbyists on the
// world (D-29), except the ruler's when the ruler said yes.
func askGovernment(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if withdrawerID := withdrawnBy(factionState, requester, worldID); withdrawerID != "" {
		return refusal(requester, def, worldID, cause, withdrawerID, domain.RefusalBarred), nil
	}
	rulerID := factionState.Ruler(worldID)
	grantedBy := ""
	granted := rulerID == requester.ID || PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def})
	if !granted {
		request := hooks.PermissionRequest{Requester: requester, WorldID: worldID, Def: def, Cause: cause}
		if rulerID != "" {
			request.Ruler = factionState.Factions[rulerID]
		}
		var err error
		if granted, err = collector.RequestPermission(request); err != nil {
			return nil, fmt.Errorf("permission request: %w", err)
		}
		if !granted {
			decidedBy := ""
			if request.Ruler != nil {
				decidedBy = request.Ruler.ID
			}
			return refusal(requester, def, worldID, cause, decidedBy, domain.RefusalDenied), nil
		}
		grantedBy = rulerID
	}
	return contestGrant(registry, collector, factionState, requester, grantedBy, def, worldID, cause)
}

// withdrawnBy returns the faction whose Lobbyists withdrew requester's permission
// on worldID this turn, the first in faction ID order, or "" when none has.
func withdrawnBy(factionState *state.FactionState, requester *domain.Faction, worldID string) string {
	if factionState == nil {
		return ""
	}
	holder := budgetHolder(factionState, requester)
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if holder.BudgetSpent(PermissionWithdrawnKey(worldID, factionID)) {
			return factionID
		}
	}
	return ""
}

// contestGrant offers a grant to each rival's contesters on worldID (Lobbyists):
// rival factions in ID order, then their assets in ID order (Effort 5 Decision 11).
// grantedBy, the ruler who said yes, is not offered its own grant to contest. Each
// offer asks its owner whether to contest; on a yes the hook rolls, and the first
// win withdraws the grant and bars the requester from that world until its next
// turn. It returns the withdrawal, or nil when the grant stands.
func contestGrant(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, grantedBy string, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if registry == nil || factionState == nil {
		return nil, nil
	}
	query := hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def}
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if factionID == requester.ID || factionID == grantedBy {
			continue
		}
		rival := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(rival) {
			for _, registered := range registry.PermissionContestersForAsset(rival.ID, asset.ID) {
				offer := registered.Hook.OfferContest(query)
				if offer == nil {
					continue
				}
				contest, err := collector.ConfirmPermissionContest(hooks.PermissionContestOffer{
					Owner: rival, Source: offer.Source, Requester: requester, WorldID: worldID, Def: def, Cause: cause,
				})
				if err != nil {
					return nil, fmt.Errorf("permission contest: %w", err)
				}
				if !contest || !registered.Hook.WinsContest(query) {
					continue
				}
				// deliberate: mark the requester the state holds, not the real faction, because in Use Asset Ability that is the action's view, and nothing after the action phase asks.
				budgetHolder(factionState, requester).SpendBudget(PermissionWithdrawnKey(worldID, rival.ID))
				return refusal(requester, def, worldID, cause, rival.ID, domain.RefusalWithdrawn), nil
			}
		}
	}
	return nil, nil
}

// budgetHolder is the faction whose HookBudgets carry requester's mark: its entry
// in factionState, or requester itself when the state does not hold it.
func budgetHolder(factionState *state.FactionState, requester *domain.Faction) *domain.Faction {
	if factionState != nil {
		if holder := factionState.Factions[requester.ID]; holder != nil {
			return holder
		}
	}
	return requester
}

func refusal(requester *domain.Faction, def *domain.AssetDefinition, worldID, cause, decidedBy string, reason domain.RefusalReason) *domain.PermissionRefused {
	return &domain.PermissionRefused{
		FactionID:          requester.ID,
		WorldID:            worldID,
		DefinitionID:       def.ID,
		DecidedByFactionID: decidedBy,
		Reason:             reason,
		Cause:              cause,
		CausedByFactionID:  requester.ID,
	}
}
