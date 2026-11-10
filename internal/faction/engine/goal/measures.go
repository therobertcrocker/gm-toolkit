package goal

import (
	"log/slog"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// countStealthedAssets counts the holder's stealthed assets in the state the
// batch leaves (Decision 2, D-10). An asset in transit has no world, so it is on
// no one's ground.
func countStealthedAssets(evaluation evalContext, filters domain.GoalFilters) int {
	projected := project(evaluation.factionState, evaluation.batch)
	holder := projected.Factions[evaluation.faction.ID]
	if holder == nil {
		return 0
	}
	rivalHomeworlds := rivalHomeworldsOf(holder.ID, projected)
	count := 0
	for _, asset := range holder.Assets {
		if !asset.Stealthy {
			continue
		}
		if filters.Ground == groundRival && !isRivalGround(holder.ID, asset.Location.WorldID, rivalHomeworlds, projected) {
			continue
		}
		if filters.ExcludeHeldAtStart && slices.Contains(evaluation.faction.ActiveGoal.StealthedAtStart, asset.ID) {
			continue
		}
		count++
	}
	return count
}

// project returns a copy of factionState with batch applied. The goal engine
// runs before the batch applies, and a measure counts what the batch leaves.
func project(factionState *state.FactionState, batch []domain.Mutation) *state.FactionState {
	projected := factionState.Clone()
	discard := slog.New(slog.DiscardHandler)
	// deliberate: a miss on the copy is dropped, not reported, because the orchestrator's own apply of this batch reports it.
	_ = mutation.New(discard).Apply(projected, batch, discard)
	return projected
}

func rivalHomeworldsOf(factionID string, factionState *state.FactionState) map[string]bool {
	homeworlds := make(map[string]bool, len(factionState.Factions))
	for otherID, other := range factionState.Factions {
		if otherID != factionID {
			homeworlds[other.Homeworld.WorldID] = true
		}
	}
	return homeworlds
}

// isRivalGround reports whether worldID is a rival's ground for factionID: a
// rival governs it, or it is independent and a rival's homeworld. A world the
// faction governs never is, a seized rival homeworld included.
func isRivalGround(factionID, worldID string, rivalHomeworlds map[string]bool, factionState *state.FactionState) bool {
	if worldID == "" {
		return false
	}
	if !factionState.IsIndependent(worldID) {
		return factionState.Ruler(worldID) != factionID
	}
	return rivalHomeworlds[worldID]
}
