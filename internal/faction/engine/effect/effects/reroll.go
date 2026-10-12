package effects

import (
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// rerollRegistrar reads own_roll as the owner rerolling its own dice, and
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll.
// owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
		domain.TriggerEnemyRollAtMe: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerReroll,
}

func registerReroll(ctx effect.RegisterContext) {
	source := ctx.Def.Name
	if ctx.Entry.Trigger == domain.TriggerEnemyRollAtMe {
		source += " (forced)"
	}
	ctx.Registry.RegisterRollResultHook(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialReroll)+":"+ctx.Def.ID,
		&rerollHook{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			defID:          ctx.Def.ID,
			source:         source,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
			oncePer:        ctx.Entry.Gates.OncePer,
			params:         *ctx.Entry.Reroll,
		},
	)
}

type rerollHook struct {
	ownerFactionID string
	assetID        string
	defID          string
	source         string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
	oncePer        domain.OncePerKey
	params         domain.RerollParams
}

func (hook *rerollHook) OnRollResult(ctx hooks.RollContext, result hooks.RollResult, factionState *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	asset := liveAsset(factionState, hook.ownerFactionID, hook.assetID)
	if !active(asset) || !rollMatches(ctx, hook.trigger, hook.ownerFactionID, asset, hook.filters) {
		return hooks.RerollDirective{}
	}
	forced := hook.trigger == domain.TriggerEnemyRollAtMe
	directive := hooks.RerollDirective{
		Source:   hook.source,
		Indices:  pickDice(result.Dice, hook.params.Dice, forced),
		Elective: hook.params.Elective,
	}
	if hook.oncePer != "" {
		directive.BudgetKey = budgetKey(hook.oncePer, domain.SpecialReroll, hook.trigger, hook.defID, hook.assetID, otherSideID(ctx, forced))
	}
	return directive
}

// pickDice returns the indices of the n dice a reroll targets: the owner's own
// lowest dice, or a forced rival's highest.
func pickDice(dice []int, n int, highest bool) []int {
	indices := make([]int, len(dice))
	for i := range dice {
		indices[i] = i
	}
	sort.SliceStable(indices, func(a, b int) bool {
		if highest {
			return dice[indices[a]] > dice[indices[b]]
		}
		return dice[indices[a]] < dice[indices[b]]
	})
	return indices[:min(n, len(indices))]
}

// otherSideID names the faction across the roll from the owner, for a
// target_faction budget: the opponent of the owner's own roll, or the forced roller.
func otherSideID(ctx hooks.RollContext, forced bool) string {
	if forced {
		return ctx.Actor.ID
	}
	if ctx.Opponent == nil {
		return ""
	}
	return ctx.Opponent.ID
}
