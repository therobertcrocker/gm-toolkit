package steps

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func Movement(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	_ domain.Roller,
	factionState *state.FactionState,
	_ *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	destination, err := collector.SelectMoveDestination(asset, worldsFromState(factionState))
	if err != nil {
		return nil, err
	}
	var mutations []domain.Mutation
	if step.CoinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{FactionID: faction.ID, Delta: -step.CoinCost, Cause: "ability", CausedByFactionID: faction.ID})
	}
	mutations = append(mutations, domain.AssetMoved{
		FactionID:         faction.ID,
		AssetID:           asset.ID,
		FromLocation:      asset.Location,
		ToLocation:        destination,
		Cause:             "ability",
		CausedByFactionID: faction.ID,
	})
	return mutations, nil
}

func worldsFromState(factionState *state.FactionState) []string {
	seen := map[string]bool{}
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			if asset.Location != "" {
				seen[asset.Location] = true
			}
		}
		for _, base := range faction.Bases {
			if base.Location != "" {
				seen[base.Location] = true
			}
		}
	}
	worlds := make([]string, 0, len(seen))
	for world := range seen {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	worlds = append(worlds, "Astral Sea")
	return worlds
}
