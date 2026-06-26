package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

// relocate resolves a movement ability: gather the movable candidates (the acting
// asset for self-scope; range- and type-filtered friendly assets for other-scope),
// prompt the GM for the moved assets and their destinations, and emit the relocate
// mutations via the world engine. Range is measured from the acting asset.
func relocate(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.Relocate
	if params == nil {
		return nil, fmt.Errorf("relocate: step has no relocate params")
	}
	if ctx.WorldEngine == nil {
		return nil, fmt.Errorf("relocate: no world engine")
	}
	actingHex := ctx.Asset.Location.RegionHex
	// DR-1: the range mode picks the enumerator; the rest of the handler is mode-agnostic.
	var (
		candidateWorlds []string
		err             error
	)
	switch params.Range {
	case domain.RangeRegion:
		candidateWorlds, err = ctx.WorldEngine.WorldsInRegion(actingHex)
	default: // RangeHexes (and the zero value, for safety)
		candidateWorlds, err = ctx.WorldEngine.WorldsWithin(actingHex, params.MaxHex)
	}
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}

	var candidates []*domain.Asset
	maxAssets := params.MaxAssets
	switch params.Scope {
	case domain.RelocateSelf:
		candidates = []*domain.Asset{ctx.Asset}
		maxAssets = 1
	case domain.RelocateOther:
		candidates = eligibleRelocateAssets(ctx, params, candidateWorlds)
	default:
		return nil, fmt.Errorf("relocate: unknown scope %q", params.Scope)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("relocate: no movable assets in range")
	}

	orders, err := ctx.Collector.SelectRelocateOrders(ctx.Asset, candidates, candidateWorlds, maxAssets)
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("relocate: %w", action.ErrNoSelection)
	}
	if params.MaxAssets > 0 && len(orders) > params.MaxAssets {
		return nil, fmt.Errorf("relocate: %d assets selected, at most %d allowed", len(orders), params.MaxAssets)
	}
	if ctx.Faction.Coin < params.CoinCost*len(orders) {
		return nil, fmt.Errorf("relocate: insufficient Coin: need %d, have %d", params.CoinCost*len(orders), ctx.Faction.Coin)
	}

	var mutations []domain.Mutation
	for _, order := range orders {
		loc, ok := ctx.WorldEngine.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := ctx.WorldEngine.BuildRelocateMutations(ctx.Faction, order.Asset, dest, candidateWorlds, params.CoinCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
	}
	return mutations, nil
}

// eligibleRelocateAssets returns the acting faction's assets that may be relocated by
// an other-scope step: matching AssetType (when set), currently in a world within
// range, and not the acting asset itself. ID-sorted for deterministic prompts.
func eligibleRelocateAssets(ctx StepContext, params *domain.RelocateParams, candidateWorlds []string) []*domain.Asset {
	inRange := make(map[string]struct{}, len(candidateWorlds))
	for _, worldID := range candidateWorlds {
		inRange[worldID] = struct{}{}
	}
	var eligible []*domain.Asset
	for _, asset := range ctx.Faction.Assets {
		if asset.ID == ctx.Asset.ID {
			continue
		}
		if _, ok := inRange[asset.Location.WorldID]; !ok {
			continue
		}
		if params.AssetType != "" {
			def, ok := ctx.Rulebook.Assets[asset.DefinitionID]
			if !ok || def.Type != params.AssetType {
				continue
			}
		}
		eligible = append(eligible, asset)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ID < eligible[j].ID
	})
	return eligible
}
