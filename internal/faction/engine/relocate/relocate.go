package relocate

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// OrderCollector is the one prompt a relocation needs. The action collector and
// the phase collector both satisfy it, so the action-spent harness and the free
// harness share the prompt as well as the logic.
type OrderCollector interface {
	SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error)
}

// ErrNoCandidates reports that nothing in range may move. The action-spent
// harness surfaces it; the free harness skips the grant.
var ErrNoCandidates = errors.New("no movable assets in range")

// ErrInsufficientCoin reports that the faction cannot pay for the relocation.
var ErrInsufficientCoin = errors.New("insufficient Coin")

// Request is one relocation. Acting is the asset range is measured from: the
// ability's own asset, or the asset that granted a free move.
type Request struct {
	Faction   *domain.Faction
	Acting    *domain.Asset
	Params    *domain.RelocateParams
	World     *world.WorldEngine
	Rulebook  *rulebook.Rulebook
	Collector OrderCollector
}

// Run resolves a relocation: gather the movable candidates (the acting asset for
// self-scope; range- and type-filtered friendly assets for other-scope), prompt
// for the moved assets and their destinations, and emit the relocate mutations
// via the world engine. Range is measured from the acting asset.
func Run(request Request) ([]domain.Mutation, error) {
	params := request.Params
	if request.World == nil {
		return nil, fmt.Errorf("relocate: no world engine")
	}
	actingHex := request.Acting.Location.RegionHex
	// other-scope may target the acting asset's own world (a friendly asset there is a
	// legal source, and the home world is a legal destination); self-scope must move off it.
	includeOrigin := params.Scope == domain.RelocateOther
	// DR-1: the range mode picks the enumerator; the rest of the handler is mode-agnostic.
	var (
		candidateWorlds []string
		err             error
	)
	switch params.Range {
	case domain.RangeRegion:
		candidateWorlds, err = request.World.WorldsInRegion(actingHex, includeOrigin)
	default: // RangeHexes (and the zero value, for safety)
		candidateWorlds, err = request.World.WorldsWithin(actingHex, params.MaxHex, includeOrigin)
	}
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}

	var candidates []*domain.Asset
	maxAssets := params.MaxAssets
	switch params.Scope {
	case domain.RelocateSelf:
		candidates = []*domain.Asset{request.Acting}
		maxAssets = 1
	case domain.RelocateOther:
		candidates = eligibleAssets(request, candidateWorlds)
	default:
		return nil, fmt.Errorf("relocate: unknown scope %q", params.Scope)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("relocate: %w", ErrNoCandidates)
	}
	if request.Faction.Coin < params.CoinCost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, params.CoinCost, request.Faction.Coin)
	}

	orders, err := request.Collector.SelectRelocateOrders(request.Acting, candidates, candidateWorlds, maxAssets)
	if err != nil {
		return nil, fmt.Errorf("relocate: %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("relocate: %w", action.ErrNoSelection)
	}
	if params.MaxAssets > 0 && len(orders) > params.MaxAssets {
		return nil, fmt.Errorf("relocate: %d assets selected, at most %d allowed", len(orders), params.MaxAssets)
	}
	if cost := totalCost(params, len(orders)); request.Faction.Coin < cost {
		return nil, fmt.Errorf("relocate: %w: need %d, have %d", ErrInsufficientCoin, cost, request.Faction.Coin)
	}

	var mutations []domain.Mutation
	for i, order := range orders {
		orderCost := params.CoinCost
		if params.CoinCostMode == domain.RelocateCostFlat && i > 0 {
			orderCost = 0
		}
		loc, ok := request.World.Location(order.DestinationWorld)
		if !ok {
			return nil, fmt.Errorf("relocate: unknown destination world %q", order.DestinationWorld)
		}
		dest := domain.Location{WorldID: order.DestinationWorld, RegionHex: loc.RegionHex()}
		// DR-2: the primitive validates membership in candidateWorlds (the range-filtered set).
		muts, err := request.World.BuildRelocateMutations(request.Faction, order.Asset, dest, candidateWorlds, orderCost)
		if err != nil {
			return nil, fmt.Errorf("relocate: %w", err)
		}
		mutations = append(mutations, muts...)
	}
	return mutations, nil
}

// eligibleAssets returns the faction's assets that an other-scope relocation may
// move: in a world within range, passing the definition filters, not the acting
// asset, and not under a movement order. ID-sorted for deterministic prompts.
func eligibleAssets(request Request, candidateWorlds []string) []*domain.Asset {
	inRange := make(map[string]struct{}, len(candidateWorlds))
	for _, worldID := range candidateWorlds {
		inRange[worldID] = struct{}{}
	}
	var eligible []*domain.Asset
	for _, asset := range request.Faction.Assets {
		// deliberate: skip an asset under orders, not just one in flight, because it keeps its origin WorldID until the next tick.
		if asset.ID == request.Acting.ID || asset.CurrentOrder != nil {
			continue
		}
		if _, ok := inRange[asset.Location.WorldID]; !ok {
			continue
		}
		if !matchesFilters(request.Rulebook.Assets[asset.DefinitionID], request.Params) {
			continue
		}
		eligible = append(eligible, asset)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ID < eligible[j].ID
	})
	return eligible
}

// matchesFilters reports whether def passes the relocation's type, category and
// excluded-type filters. An unknown definition passes only when no filter is set.
func matchesFilters(def *domain.AssetDefinition, params *domain.RelocateParams) bool {
	filtered := params.AssetType != "" || len(params.Categories) > 0 || len(params.ExcludeTypes) > 0
	if def == nil {
		return !filtered
	}
	if params.AssetType != "" && def.Type != params.AssetType {
		return false
	}
	if len(params.Categories) > 0 && !slices.Contains(params.Categories, def.Category) {
		return false
	}
	return !slices.Contains(params.ExcludeTypes, def.Type)
}

// totalCost is what a relocation of moved assets charges: CoinCost per asset, or
// CoinCost once under flat mode (Transit Web: "for 1 Coin, any number").
func totalCost(params *domain.RelocateParams, moved int) int {
	if params.CoinCostMode == domain.RelocateCostFlat {
		return params.CoinCost
	}
	return params.CoinCost * moved
}
