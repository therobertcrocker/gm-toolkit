package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// raiseTechLevel is Teach Planetary Population: pay the step's cost, and on a passed
// target roll the target world counts as the step's tech level for the acting
// faction's purchases (D-11). The cost is paid on a miss too.
func raiseTechLevel(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.RaiseTechLevel
	if params == nil || ctx.Step.TargetRoll == nil {
		return nil, fmt.Errorf("raise_tech_level: step has no target roll or parameters")
	}
	// deliberate: an error, not a quiet no-op, because raiseTechLevelWorlds offers no world to a faction that cannot pay, so reaching here means the filter and the handler disagree.
	if ctx.Faction.Coin < params.CoinCost {
		return nil, fmt.Errorf("raise_tech_level: %s has %d Coin, needs %d; the world filter should have offered no target", ctx.Faction.ID, ctx.Faction.Coin, params.CoinCost)
	}
	var mutations []domain.Mutation
	if params.CoinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{
			FactionID:         ctx.Faction.ID,
			Delta:             -params.CoinCost,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		})
	}
	if ctx.Rolled.Passed {
		mutations = append(mutations, domain.WorldTechLevelRaised{
			FactionID:         ctx.Faction.ID,
			WorldID:           ctx.Target.World,
			TechLevel:         params.TechLevel,
			Cause:             "ability",
			CausedByFactionID: ctx.Faction.ID,
		})
	}
	return mutations, nil
}
