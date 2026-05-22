package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// eligibleAttackers returns faction assets that may attack this turn:
// Ready (not on purchase/refit cooldown), alive, and maintained.
func eligibleAttackers(faction *domain.Faction) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range faction.Assets {
		if asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
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

// eligibleDefendersAtHex returns targetable assets at a given hex coordinate:
// owned by a different faction, non-stealthy, Ready, alive, and maintained.
// Used when the attacker is mid-flight (empty WorldID).
func eligibleDefendersAtHex(attackerFactionID string, hex spatial.RegionHex, index *world.Index) []*domain.Asset {
	return filterOpposingDefenders(attackerFactionID, index.AssetsByHex[hex])
}

// filterOpposingDefenders applies the defender eligibility filter to a candidate
// slice: excludes the attacker's own assets and assets that are stealthy, not
// Ready, destroyed, or unmaintained.
func filterOpposingDefenders(attackerFactionID string, candidates []*domain.Asset) []*domain.Asset {
	var result []*domain.Asset
	for _, asset := range candidates {
		if asset.OwnerID != attackerFactionID && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
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
