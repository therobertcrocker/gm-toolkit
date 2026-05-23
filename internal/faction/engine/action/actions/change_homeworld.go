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

type ChangeHomeworld struct {
	factionID      string
	collector      action.Collector
	world          *world.WorldEngine
	rulebook       *rulebook.Rulebook
	targetWorld    *domain.Location
	turnsRemaining int
}

func NewChangeHomeworld(collector action.Collector, worldEngine *world.WorldEngine, rulebook *rulebook.Rulebook) *ChangeHomeworld {
	return &ChangeHomeworld{collector: collector, world: worldEngine, rulebook: rulebook}
}

func (c *ChangeHomeworld) Name() string { return "Change Homeworld" }

func (c *ChangeHomeworld) Validate(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
	if faction.ActiveGoal != nil {
		return false // can't initiate while another goal is active
	}
	return len(changeHomeworldTargets(faction, c.world, c.rulebook)) > 0
}

func (c *ChangeHomeworld) Inputs(faction *domain.Faction, factionState *state.FactionState, _ *rulebook.Rulebook) error {
	targetWorldID, err := c.collector.SelectChangeHomeworldTarget(faction, factionState)
	if err != nil {
		return fmt.Errorf("change homeworld: %w", err)
	}
	target, ok := c.world.Location(targetWorldID)
	if !ok {
		return fmt.Errorf("change homeworld: unknown world %q", targetWorldID)
	}
	c.targetWorld = &domain.Location{WorldID: targetWorldID, RegionHex: target.RegionHex()}
	c.factionID = faction.ID
	return nil
}

func (c *ChangeHomeworld) Resolve(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	if c.targetWorld == nil {
		return fmt.Errorf("change homeworld: %w", action.ErrNoSelection)
	}
	crossingCost := c.rulebook.DriftCost(3) // faction-level default drift rating
	dist, err := c.world.Distance(faction.Homeworld.RegionHex, c.targetWorld.RegionHex, crossingCost)
	if err != nil {
		return fmt.Errorf("change homeworld: computing distance: %w", err)
	}
	c.turnsRemaining = 1 + dist
	return nil
}

func (c *ChangeHomeworld) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.GoalInitiated{
			FactionID:         c.factionID,
			GoalID:            "G-012",
			ProcessPhase:      0,
			TargetWorld:       *c.targetWorld,
			Cause:             "change homeworld",
			CausedByFactionID: c.factionID,
		},
		domain.GoalPhaseAdvanced{
			FactionID:      c.factionID,
			GoalID:         "G-012",
			ProcessPhase:   0,
			TurnsRemaining: c.turnsRemaining,
			Cause:          "change homeworld",
		},
	}, nil
}

func changeHomeworldTargets(faction *domain.Faction, worldEngine *world.WorldEngine, rulebook *rulebook.Rulebook) []string {
	var targets []string
	crossingCost := rulebook.DriftCost(3)
	for _, base := range faction.Bases {
		if base.Location.WorldID == faction.Homeworld.WorldID {
			continue // current homeworld is excluded
		}
		target, ok := worldEngine.Location(base.Location.WorldID)
		if !ok {
			continue
		}
		if _, err := worldEngine.Distance(faction.Homeworld.RegionHex, target.RegionHex(), crossingCost); err != nil {
			continue
		}
		targets = append(targets, base.Location.WorldID)
	}
	sort.Strings(targets)
	return targets
}
