package ability

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// shelter is Leased Demiplane: the target, one of the faction's own assets on the
// holder's world, goes inside. A sheltered asset is hidden, untouchable and cannot
// act (domain.Asset.Targetable). Its owner may bring it out free at its turn start
// (turn.OfferShelterReleases), and it is destroyed if damage destroys the holder
// (shelter_asset's asset_destroyed route).
func shelter(ctx StepContext) ([]domain.Mutation, error) {
	if ctx.Asset == nil {
		return nil, fmt.Errorf("shelter: needs an asset holder")
	}
	return []domain.Mutation{domain.AssetConditionApplied{
		FactionID:         ctx.Faction.ID,
		AssetID:           ctx.Target.Asset.ID,
		Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: ctx.Asset.ID},
		Cause:             "ability",
		CausedByFactionID: ctx.Faction.ID,
	}}, nil
}
