package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner; rival_roll on any rival's roll, whether or not the owner is in it.
// scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world; scope = "homeworld" to any such roll on the owner's
// homeworld; scope = "faction" to any such roll, so asset is nil for a tag. The
// dispatcher reaches both sides, so this check, not the registration scope, is
// what keeps an entry to its own side. attribute reads the stat this roll uses;
// opponent_attribute the stat the other side rolls against it. tech_level reads
// the rolling asset's definition, opponent_tech_level the other side's; neither
// matches without a rulebook. detection matches only a roll made by the side
// testing to detect stealth.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters, rulebook *rulebook.Rulebook) bool {
	if ctx.Actor == nil {
		return false
	}
	var owner *domain.Faction
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Actor, ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Opponent, ctx.OpponentAsset
	case domain.TriggerRivalRoll:
		// deliberate: return here, not read scope or filters, because the owner takes no part in the roll and the registrar accepts only scope faction and no filters (Effort 3 Decision 7).
		return ctx.Actor.ID != ownerFactionID
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
	case domain.ScopeHomeworld:
		if ctx.World == "" || ctx.World != owner.Homeworld.WorldID {
			return false
		}
	case domain.ScopeFaction:
	default:
		return false
	}
	if filters.Attribute != "" && ctx.Attribute != filters.Attribute {
		return false
	}
	if filters.OpponentAttribute != "" && ctx.OpponentAttribute != filters.OpponentAttribute {
		return false
	}
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, ctx.Phase) {
		return false
	}
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	if filters.Detection && !ctx.Detection {
		return false
	}
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
	if filters.AssetDefinition != "" && (ctx.Asset == nil || ctx.Asset.DefinitionID != filters.AssetDefinition) {
		return false
	}
	return true
}

// techLevelMatches reports whether asset's definition requires exactly want. A nil
// want is no filter; a missing asset, rulebook or definition matches nothing.
func techLevelMatches(asset *domain.Asset, want *int, rulebook *rulebook.Rulebook) bool {
	if want == nil {
		return true
	}
	if asset == nil || rulebook == nil {
		return false
	}
	definition := rulebook.Assets[asset.DefinitionID]
	return definition != nil && definition.TechLevel == *want
}
