package dice

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ResolveTie asks the global hooks, then the attacker's side, then the defender's:
// on each side its faction's hooks, then each of its assets' in ID order. The first
// answer that is not TieStandard wins, because with asset buckets in the walk a
// resolver that does not match must not end the search (Effort 1 Decision 4).
// TieStandard when none answers.
func ResolveTie(registry *hooks.Registry, ctx hooks.RollContext, factionState *state.FactionState) hooks.TieOutcome {
	if registry == nil {
		return hooks.TieStandard
	}
	scopes := []hooks.Scope{hooks.GlobalScope()}
	for _, side := range []*domain.Faction{ctx.Actor, ctx.Opponent} {
		if side == nil {
			continue
		}
		scopes = append(scopes, hooks.FactionScope(side.ID))
		for _, asset := range domain.SortedAssets(side) {
			scopes = append(scopes, hooks.AssetScope(side.ID, asset.ID))
		}
	}
	for _, scope := range scopes {
		for _, registered := range registry.TieResolversAt(scope) {
			if outcome := registered.Hook.ResolveTie(ctx, factionState); outcome != hooks.TieStandard {
				return outcome
			}
		}
	}
	return hooks.TieStandard
}
