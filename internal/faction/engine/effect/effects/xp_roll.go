package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// xpRollRegistrar is Exchange Consulate's bonus XP: when the owner completes a
// goal, it rolls the entry's die and gains the entry's XP on the target or more
// (D-18). owner = "own" reads as "the owner completed the goal"; goal names the
// completed goal, checked at load. No choice, so no prompt.
var xpRollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerGoalCompleted: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterGoal},
		},
	},
	Register: registerXPRoll,
}

func registerXPRoll(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialXPRoll), &xpRollReactor{
		holder: refOf(ctx.Holder),
		goal:   ctx.Entry.Filters.Goal,
		params: *ctx.Entry.XPRoll,
		roller: ctx.Roller,
	})
}

type xpRollReactor struct {
	holder holderRef
	goal   string
	params domain.XPRollParams
	roller domain.Roller
}

// OnMutations rolls once for each of the owner's GoalCompleted that names the
// entry's goal, or any goal when the entry names none.
func (reactor *xpRollReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	var awarded []domain.Mutation
	for _, mutation := range mutations {
		completed, ok := mutation.(domain.GoalCompleted)
		if !ok || completed.FactionID != reactor.holder.factionID || (reactor.goal != "" && completed.GoalID != reactor.goal) {
			continue
		}
		if dice.Amount(reactor.params.Die, reactor.roller) < reactor.params.Target {
			continue
		}
		awarded = append(awarded, domain.XPAwarded{
			FactionID:         reactor.holder.factionID,
			Amount:            reactor.params.Amount,
			Cause:             string(domain.SpecialXPRoll),
			CausedByFactionID: reactor.holder.factionID,
		})
	}
	return awarded
}
