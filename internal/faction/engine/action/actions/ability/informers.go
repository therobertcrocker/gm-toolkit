package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func informers(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	_ *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	candidates := informersCandidates(factionState, faction.ID, asset.Location.WorldID)
	targetFaction, err := collector.SelectFactionTestTarget(asset, def.Ability.Effect, candidates)
	if err != nil {
		return nil, err
	}
	if targetFaction == nil {
		return nil, fmt.Errorf("informers: %w", action.ErrNoSelection)
	}

	attackRoll := roller.Roll(10) + statScore(faction, def.Ability.AttackerStat)
	defenseRoll := roller.Roll(10) + statScore(targetFaction, def.Ability.DefenderStat)

	if attackRoll <= defenseRoll {
		return nil, nil
	}

	var mutations []domain.Mutation
	for _, targetAsset := range targetFaction.Assets {
		if targetAsset.Location.WorldID == asset.Location.WorldID && targetAsset.Stealthy {
			mutations = append(mutations, domain.AssetStealthCleared{
				FactionID:         targetFaction.ID,
				AssetID:           targetAsset.ID,
				Cause:             "ability",
				CausedByFactionID: faction.ID,
			})
		}
	}
	return mutations, nil
}

// informersCandidates returns all factions other than the acting faction.
// For reveal_stealth, every opponent is a valid target regardless of world presence.
func informersCandidates(factionState *state.FactionState, actingFactionID, _ string) []*domain.Faction {
	var candidates []*domain.Faction
	for id, faction := range factionState.Factions {
		if id == actingFactionID {
			continue
		}
		candidates = append(candidates, faction)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

func statScore(faction *domain.Faction, stat domain.FactionStat) int {
	switch stat {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}
