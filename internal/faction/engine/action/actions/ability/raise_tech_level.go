package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// raiseTechLevel is Teach Planetary Population: pay the step's cost, and on a
// matching outcome row the target world counts as the step's tech level for the
// acting faction's purchases (D-11). The cost is paid on a miss too. A faction
// that cannot pay does nothing.
func raiseTechLevel(ctx StepContext) ([]domain.Mutation, error) {
	params := ctx.Step.RaiseTechLevel
	if params == nil || ctx.Step.Roll == nil {
		return nil, fmt.Errorf("raise_tech_level: step has no roll table or parameters")
	}
	if ctx.Faction.Coin < params.CoinCost {
		return nil, nil
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
	if _, ok := outcomeFor(ctx.RollResult, ctx.Step.Roll.Outcomes); ok {
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
