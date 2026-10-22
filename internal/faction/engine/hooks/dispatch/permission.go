package dispatch

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RequestPermission asks worldID's government whether requester may raise def
// there, and returns the refusal record, or nil when the purchase may go ahead. A
// definition that needs no permission is not asked about, nor is a purchase on the
// requester's homeworld (Effort 4 Decision 15). A ruler's own request is granted at
// once (D-6), and so is one a grant hook of the requester serves (Popular Movement).
// Otherwise the ruler's driver is asked, or the GM for an independent world (D-4).
// cause names the purchase path on the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsPurchasePermission() {
		return nil, nil
	}
	// deliberate: the homeworld skips the ask for purchases, not arrivals, because the book's homeworld clause covers buying (Effort 4 Decision 15).
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	rulerID := factionState.Ruler(worldID)
	if rulerID == requester.ID {
		return nil, nil
	}
	if PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: worldID, Def: def}) {
		return nil, nil
	}
	request := hooks.PermissionRequest{Requester: requester, WorldID: worldID, Def: def}
	if rulerID != "" {
		request.Ruler = factionState.Factions[rulerID]
	}
	granted, err := collector.RequestPermission(request)
	if err != nil {
		return nil, fmt.Errorf("permission request: %w", err)
	}
	if granted {
		return nil, nil
	}
	decidedBy := ""
	if request.Ruler != nil {
		decidedBy = request.Ruler.ID
	}
	return &domain.PermissionRefused{
		FactionID:          requester.ID,
		WorldID:            worldID,
		DefinitionID:       def.ID,
		DecidedByFactionID: decidedBy,
		Cause:              cause,
		CausedByFactionID:  requester.ID,
	}, nil
}
