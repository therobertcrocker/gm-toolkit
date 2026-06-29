package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func revealStealth(ctx StepContext) ([]domain.Mutation, error) {
	targetFaction := ctx.Target.Faction
	var mutations []domain.Mutation
	for _, targetAsset := range targetFaction.Assets {
		if targetAsset.Location.WorldID == ctx.Asset.Location.WorldID && targetAsset.Stealthy {
			mutations = append(mutations, domain.AssetStealthCleared{
				FactionID:         targetFaction.ID,
				AssetID:           targetAsset.ID,
				Cause:             "ability",
				CausedByFactionID: ctx.Faction.ID,
			})
		}
	}
	return mutations, nil
}
