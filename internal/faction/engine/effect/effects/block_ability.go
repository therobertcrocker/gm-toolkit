package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// blockAbilityRegistrar is Saboteurs: the asset it attacks, hit or miss, cannot
// use its ability until the start of the attacking faction's next turn.
var blockAbilityRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttack: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Matchup: matchupBlockAbility,
}

func matchupBlockAbility(ctx effect.MatchupContext) effect.MatchupOutcome {
	return effect.MatchupOutcome{Mutations: []domain.Mutation{domain.AssetConditionApplied{
		FactionID: ctx.Opponent.ID,
		AssetID:   ctx.OpponentAsset.ID,
		Condition: domain.Condition{
			Kind:            domain.ConditionNoAbility,
			SourceAssetID:   ctx.Asset.ID,
			ExpiresAtTurnOf: ctx.Holder.Faction.ID,
		},
		Cause:             string(domain.SpecialBlockAbility),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}}
}
