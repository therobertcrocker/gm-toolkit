package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// eligibleAttackers returns faction assets that may attack this turn: the active
// ones whose attack no Seditionists block.
func eligibleAttackers(faction *domain.Faction, factionState *state.FactionState) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Active() && !domain.AttackBlocked(factionState.Factions, asset) {
			result = append(result, asset)
		}
	}
	return result
}

// eligibleDefendersOnWorld returns targetable assets at a given world ID:
// owned by a different faction, non-stealthy, Ready, alive, and maintained.
func eligibleDefendersOnWorld(attackerFactionID, worldID string, index *world.Index) []*domain.Asset {
	return filterOpposingDefenders(attackerFactionID, index.AssetsByLocation[worldID])
}

// eligibleDefendersAtHex returns the assets at the attacker's hex that it may
// target: owned by a different faction, non-stealthy, Ready, alive, maintained,
// and not refused by an attack-eligibility hook. Works when the attacker is
// mid-flight (empty WorldID).
func eligibleDefendersAtHex(attacker *domain.Asset, attackerDef *domain.AssetDefinition, rulebook *rulebook.Rulebook, registry *hooks.Registry, index *world.Index) []*domain.Asset {
	var result []*domain.Asset
	for _, defender := range filterOpposingDefenders(attacker.OwnerID, index.AssetsByHex[attacker.Location.RegionHex]) {
		defenderDef, ok := rulebook.Assets[defender.DefinitionID]
		// deliberate: an unknown definition stays eligible, not dropped, so Resolve still reports it as an error
		if ok && !dispatch.ResolveAttackEligibility(registry, hooks.AttackEligibilityQuery{
			Attacker:    attacker,
			AttackerDef: attackerDef,
			Defender:    defender,
			DefenderDef: defenderDef,
			Side:        hooks.SideAttack,
		}) {
			continue
		}
		result = append(result, defender)
	}
	return result
}

// filterOpposingDefenders applies the defender eligibility filter to a candidate
// slice: excludes the attacker's own assets and assets that are stealthy or not
// targetable.
func filterOpposingDefenders(attackerFactionID string, candidates []*domain.Asset) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range candidates {
		// deliberate: Targetable, not Active, because a disabled asset can still be attacked (D-38).
		if asset.OwnerID != attackerFactionID && !asset.Stealthy && asset.Targetable() {
			result = append(result, asset)
		}
	}
	return result
}

// factionBaseOnWorld returns the faction's Base of Influence on the given world,
// or nil if the faction has no Base there.
func factionBaseOnWorld(faction *domain.Faction, world string) *domain.Base {
	for _, base := range faction.Bases {
		if base.Location.WorldID == world {
			return base
		}
	}
	return nil
}
