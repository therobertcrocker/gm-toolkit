package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// loseTiesRegistrar is Fanatical's tie rule: the owner loses every tie in an
// attack, as attacker or defender. owner = "own" reads as "the owner is on either
// side of the tie", scope = "faction" as "anywhere".
var loseTiesRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAttackTie: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerLoseTies,
}

func registerLoseTies(ctx effect.RegisterContext) {
	ctx.Registry.RegisterTieResolver(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialLoseTies), &loseTiesResolver{holder: refOf(ctx.Holder)})
}

type loseTiesResolver struct {
	holder holderRef
}

func (resolver *loseTiesResolver) ResolveTie(ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if _, live := resolver.holder.live(factionState); !live {
		return hooks.TieStandard
	}
	switch {
	case ctx.Actor != nil && ctx.Actor.ID == resolver.holder.factionID:
		return hooks.TieDefenderWins
	case ctx.Opponent != nil && ctx.Opponent.ID == resolver.holder.factionID:
		return hooks.TieAttackerWins
	}
	return hooks.TieStandard
}
