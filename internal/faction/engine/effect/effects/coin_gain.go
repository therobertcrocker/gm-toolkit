package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Fire: fireCoinGain,
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Faction.ID,
		Delta:             ctx.Entry.CoinGain.Amount.Roll(ctx.Roller),
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Faction.ID,
	}}
}
