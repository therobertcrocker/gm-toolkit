package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var rollPhaseNames = map[hooks.RollPhase]domain.EffectRollPhase{
	hooks.PhaseAttack:      domain.RollPhaseAttack,
	hooks.PhaseDefense:     domain.RollPhaseDefense,
	hooks.PhaseFactionTest: domain.RollPhaseFactionTest,
	hooks.PhaseContested:   domain.RollPhaseContested,
}

// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world; scope = "faction" to any such roll, so asset is nil for a
// tag. The dispatcher reaches both sides, so this check, not the registration
// scope, is what keeps an entry to its own side.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters) bool {
	if ctx.Actor == nil {
		return false
	}
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		subject = ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		subject = ctx.OpponentAsset
	default:
		return false
	}
	switch filters.Scope {
	case domain.ScopeSelf:
		if subject == nil || subject.ID != asset.ID {
			return false
		}
	case domain.ScopeWorld:
		if ctx.World == "" || ctx.World != asset.Location.WorldID {
			return false
		}
	case domain.ScopeFaction:
	default:
		return false
	}
	if filters.Attribute != "" && ctx.Attribute != string(filters.Attribute) {
		return false
	}
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, rollPhaseNames[ctx.Phase]) {
		return false
	}
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	return true
}
