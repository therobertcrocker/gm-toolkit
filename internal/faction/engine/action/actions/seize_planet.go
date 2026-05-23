package actions

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type SeizePlanet struct {
	factionID    string
	collector    action.Collector
	index        *world.Index
	targetWorld  *domain.Location
	processPhase int
}

func NewSeizePlanet(collector action.Collector, index *world.Index) *SeizePlanet {
	return &SeizePlanet{collector: collector, index: index}
}

func (s *SeizePlanet) Name() string { return "Seize Planet" }

func (s *SeizePlanet) Validate(faction *domain.Faction, factionState *state.FactionState, _ *rulebook.Rulebook) bool {
	if faction.ActiveGoal == nil || faction.ActiveGoal.GoalID != "G-004" || faction.ActiveGoal.ProcessPhase != 0 {
		return false
	}
	return len(seizePlanetTargetWorlds(faction, s.index)) > 0
}

func (s *SeizePlanet) Inputs(faction *domain.Faction, factionState *state.FactionState, _ *rulebook.Rulebook) error {
	world, err := s.collector.SelectSeizeTarget(faction, factionState)
	if err != nil {
		return fmt.Errorf("seize planet: %w", err)
	}
	s.targetWorld = &domain.Location{WorldID: world}
	s.factionID = faction.ID
	return nil
}

func (s *SeizePlanet) Resolve(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	if s.targetWorld == nil {
		return fmt.Errorf("seize planet: %w", action.ErrNoSelection)
	}
	s.processPhase = 1
	return nil
}

func (s *SeizePlanet) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.GoalInitiated{
			FactionID:         s.factionID,
			GoalID:            "G-004",
			TargetWorld:       *s.targetWorld,
			ProcessPhase:      s.processPhase,
			Cause:             "seize planet",
			CausedByFactionID: s.factionID,
		},
	}, nil
}

// seizePlanetTargetWorlds returns worlds where the faction has at least one
// unstealthed asset and a rival also has at least one unstealthed asset.
func seizePlanetTargetWorlds(faction *domain.Faction, index *world.Index) []string {
	factionFragments := map[string]struct{}{}
	for _, asset := range faction.Assets {
		if !asset.Stealthy {
			factionFragments[asset.Location.WorldID] = struct{}{}
		}
	}

	contested := map[string]struct{}{}
	for locationID := range factionFragments {
		for _, asset := range index.AssetsByLocation[locationID] {
			if asset.OwnerID != faction.ID && !asset.Stealthy {
				contested[locationID] = struct{}{}
				break
			}
		}
	}

	result := make([]string, 0, len(contested))
	for locationID := range contested {
		result = append(result, locationID)
	}
	sort.Strings(result)
	return result
}
