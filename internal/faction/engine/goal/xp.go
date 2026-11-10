package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// xp is the XP a completion pays. progress is the progress after this batch;
// completing is the match of the complete event, zero for a threshold.
func (compiled *compiledGoal) xp(ctx evalContext, progress int, completing match) int {
	spec := compiled.goal.XP
	switch spec.Formula {
	case formulaHalfProgress:
		return ceilDiv(progress, 2)
	case formulaTargetAverage:
		target := ctx.factionState.Factions[ctx.faction.ActiveGoal.TargetFactionID]
		if target == nil {
			return 1
		}
		return 1 + ceilDiv(ratingSum(target), 3)
	case formulaRulerAverage:
		changed, _ := completing.mutation.(domain.WorldGovernmentChanged)
		ruler := ctx.factionState.Factions[changed.FromFactionID]
		if ruler == nil || !changed.Resisted {
			return 1
		}
		return max(1, ceilDiv(ratingSum(ruler), 6))
	default:
		amount := spec.Amount
		if spec.ContestedBonus > 0 && rivalPresentOn(completing.world, ctx.faction.ID, ctx.factionState) {
			amount += spec.ContestedBonus
		}
		return amount
	}
}

func ceilDiv(value, divisor int) int {
	return (value + divisor - 1) / divisor
}

func ratingSum(faction *domain.Faction) int {
	return faction.Force + faction.Cunning + faction.Wealth
}

// rivalPresentOn reports whether a faction other than factionID has an asset or
// a Base on worldID (D-22). It reads live state, not the world index, which is
// rebuilt only at turn setup (Decision 9).
func rivalPresentOn(worldID, factionID string, factionState *state.FactionState) bool {
	if worldID == "" {
		return false
	}
	for otherID, other := range factionState.Factions {
		if otherID == factionID {
			continue
		}
		if other.HasBaseOn(worldID) {
			return true
		}
		for _, asset := range other.Assets {
			if asset.Location.WorldID == worldID {
				return true
			}
		}
	}
	return false
}
