package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

var applyStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Fire: fireApplyStealth,
}

func fireApplyStealth(ctx effect.FireContext) []domain.Mutation {
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialApplyStealth),
		CausedByFactionID: ctx.Faction.ID,
	}}
}
