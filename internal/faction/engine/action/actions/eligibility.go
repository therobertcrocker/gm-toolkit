package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
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

// eligibleDefenders returns all assets on world owned by factions other than
// attackerFactionID that may be targeted: non-stealthy, Ready (SWN: inactive
// assets cannot defend), alive, and maintained.
func eligibleDefenders(factionState *state.FactionState, attackerFactionID, world string) []*domain.Asset {
	var result []*domain.Asset
	for factionID, faction := range factionState.Factions {
		if factionID == attackerFactionID {
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == world && !asset.Stealthy && asset.Ready && asset.CurrentHP > 0 && asset.Maintained {
				result = append(result, asset)
			}
		}
	}
	return result
}

// factionBaseOnWorld returns the faction's Base of Influence on the given world,
// or nil if the faction has no Base there.
func factionBaseOnWorld(faction *domain.Faction, world string) *domain.Base {
	for _, base := range faction.Bases {
		if base.Location == world {
			return base
		}
	}
	return nil
}
