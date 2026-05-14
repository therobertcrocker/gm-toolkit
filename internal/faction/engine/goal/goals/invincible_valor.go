package goals

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type InvincibleValor struct{}

func (InvincibleValor) GoalID() string { return "G-010" }

func (InvincibleValor) CheckLock(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation) {
	return locks.GoalLock{Type: locks.LockNone}, nil
}

func (InvincibleValor) UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, _ *world.Index) []domain.Mutation {
	for _, mutation := range mutations {
		v, ok := mutation.(domain.AssetRemoved)
		if !ok || v.Cause != "attack" || v.CausedByFactionID != actingFaction.ID || v.FactionID == actingFaction.ID {
			continue
		}
		rivalFaction, ok := factionState.Factions[v.FactionID]
		if !ok {
			continue
		}
		asset := findAsset(rivalFaction, v.AssetID)
		if asset == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		if def.Category == domain.StatForce && def.MinRating > actingFaction.Force {
			return completeGoal(actingFaction, 2)
		}
	}
	return nil
}
