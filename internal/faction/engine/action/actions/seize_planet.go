package actions

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type SeizePlanet struct {
	factionID    string
	collector    action.Collector
	targetWorld  string
	processPhase int
}

func NewSeizePlanet(collector action.Collector) *SeizePlanet {
	return &SeizePlanet{collector: collector}
}

func (s *SeizePlanet) Name() string { return "Seize Planet" }

func (s *SeizePlanet) Validate(faction *domain.Faction, factionState *state.FactionState, _ *loader.Rulebook) bool {
	if faction.ActiveGoal == nil || faction.ActiveGoal.GoalID != "G-004" || faction.ActiveGoal.ProcessPhase != 0 {
		return false
	}
	return len(seizePlanetTargetWorlds(faction, factionState)) > 0
}

func (s *SeizePlanet) Inputs(faction *domain.Faction, factionState *state.FactionState, _ *loader.Rulebook) error {
	world, err := s.collector.SelectSeizeTarget(faction, factionState)
	if err != nil {
		return fmt.Errorf("seize planet: %w", err)
	}
	s.targetWorld = world
	s.factionID = faction.ID
	return nil
}

func (s *SeizePlanet) Resolve(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	if s.targetWorld == "" {
		return fmt.Errorf("seize planet: no target world selected")
	}
	s.processPhase = 1
	return nil
}

func (s *SeizePlanet) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.GoalInitiated{
			FactionID:         s.factionID,
			GoalID:            "G-004",
			TargetWorld:       s.targetWorld,
			ProcessPhase:      s.processPhase,
			Cause:             "seize planet",
			CausedByFactionID: s.factionID,
		},
	}, nil
}

// seizePlanetTargetWorlds returns worlds where the faction has at least one
// unstealthed asset and a rival also has at least one unstealthed asset.
func seizePlanetTargetWorlds(faction *domain.Faction, factionState *state.FactionState) []string {
	factionWorlds := map[string]struct{}{}
	for _, asset := range faction.Assets {
		if !asset.Stealthy {
			factionWorlds[asset.Location] = struct{}{}
		}
	}

	contested := map[string]struct{}{}
	for _, other := range factionState.Factions {
		if other.ID == faction.ID {
			continue
		}
		for _, asset := range other.Assets {
			if !asset.Stealthy {
				if _, ok := factionWorlds[asset.Location]; ok {
					contested[asset.Location] = struct{}{}
				}
			}
		}
	}

	result := make([]string, 0, len(contested))
	for world := range contested {
		result = append(result, world)
	}
	sort.Strings(result)
	return result
}
