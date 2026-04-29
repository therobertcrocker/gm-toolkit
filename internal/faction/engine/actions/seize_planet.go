package actions

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type SeizePlanet struct {
	collector   engine.InputCollector
	targetWorld string
}

func NewSeizePlanet(collector engine.InputCollector) *SeizePlanet {
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
	world, err := s.collector.SelectSiezeTarget(faction, factionState)
	if err != nil {
		return fmt.Errorf("seize planet: %w", err)
	}
	s.targetWorld = world
	return nil
}

func (s *SeizePlanet) Resolve(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	faction.ActiveGoal.TargetWorld = s.targetWorld
	faction.ActiveGoal.ProcessPhase = 1
	return nil
}

func (s *SeizePlanet) Output() ([]domain.Mutation, error) {
	return nil, nil
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
