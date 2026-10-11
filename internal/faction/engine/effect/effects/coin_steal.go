package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinStealRegistrar reads owner = "rival" as "the faction the hit landed on".
// scope = "world" needs no check: an attack only reaches assets that share the
// attacker's world (or hex, in flight).
var coinStealRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerCoinSteal,
}

func registerCoinSteal(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialCoinSteal)+":"+ctx.Def.ID,
		&coinStealReactor{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			defID:          ctx.Def.ID,
			amount:         ctx.Entry.CoinSteal.Amount,
			oncePer:        ctx.Entry.Gates.OncePer,
			roller:         ctx.Roller,
		},
	)
}

type coinStealReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	amount         domain.DiceRoll
	oncePer        domain.OncePerKey
	roller         domain.Roller
}

func (reactor *coinStealReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	var extra []domain.Mutation
	for _, targetID := range hitTargets(mutations, reactor.assetID, false) {
		owner := factionState.Factions[reactor.ownerFactionID]
		target := factionState.Factions[targetID]
		if owner == nil || target == nil || targetID == reactor.ownerFactionID {
			continue
		}
		// deliberate: an empty purse is checked before the budget, because a steal that takes nothing has not "happened" and must not use up the turn's one theft.
		if target.Coin <= 0 {
			continue
		}
		if reactor.oncePer != "" && !spendBudget(owner, budgetKey(reactor.oncePer, domain.SpecialCoinSteal, reactor.defID, reactor.assetID, targetID)) {
			continue
		}
		stolen := min(reactor.amount.Roll(reactor.roller), target.Coin)
		extra = append(extra,
			domain.CoinDelta{FactionID: targetID, Delta: -stolen, Cause: string(domain.SpecialCoinSteal), CausedByFactionID: reactor.ownerFactionID},
			domain.CoinDelta{FactionID: reactor.ownerFactionID, Delta: stolen, Cause: string(domain.SpecialCoinSteal), CausedByFactionID: reactor.ownerFactionID},
		)
	}
	return extra
}
