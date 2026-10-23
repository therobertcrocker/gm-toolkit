package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// hpTo returns the AssetHPDelta that moves ctx's target from its HP after the hit
// to hp, or nil when it is there already. A Cat 6 save keeps the hit in history
// and records the save beside it, as a Bank records the loss it absorbs (Effort 2
// Decision 2).
func hpTo(ctx hooks.LethalContext, hp int, cause domain.SpecialEffectType, causedByFactionID string) []domain.Mutation {
	if ctx.HPAfter == hp {
		return nil
	}
	return []domain.Mutation{domain.AssetHPDelta{
		FactionID:         ctx.Target.OwnerID,
		AssetID:           ctx.Target.ID,
		Delta:             hp - ctx.HPAfter,
		Cause:             string(cause),
		CausedByFactionID: causedByFactionID,
	}}
}

// guardsTarget reports whether self, a defending interceptor, covers ctx's target:
// self is active, and the target is another asset of self's faction on self's
// world, of one of assetTypes (any type when empty).
func guardsTarget(self *domain.Asset, ctx hooks.LethalContext, assetTypes []domain.AssetType) bool {
	if !active(self) || ctx.Target.ID == self.ID || ctx.Target.OwnerID != self.OwnerID {
		return false
	}
	if self.Location.WorldID == "" || ctx.Target.Location.WorldID != self.Location.WorldID {
		return false
	}
	return len(assetTypes) == 0 || (ctx.TargetDef != nil && slices.Contains(assetTypes, ctx.TargetDef.Type))
}
