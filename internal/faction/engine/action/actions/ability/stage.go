package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func resolveTarget(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	factionState *state.FactionState,
	collector action.Collector,
) (domain.AbilityTarget, error) {
	switch step.Target {
	case domain.TargetNone:
		return domain.AbilityTarget{Kind: domain.TargetNone}, nil
	case domain.TargetSelf:
		return domain.AbilityTarget{Kind: domain.TargetSelf, Asset: asset, Faction: faction}, nil
	case domain.TargetFaction:
		candidates := otherFactions(factionState, faction.ID)
		target, err := collector.SelectAbilityTarget(asset, step, domain.AbilityTargetCandidates{Factions: candidates})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.Faction == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		return target, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
	}
}

func passesGate(step domain.AbilityStep, faction *domain.Faction, target domain.AbilityTarget, roller domain.Roller) (bool, error) {
	if step.Contest == nil {
		// Ungated (self-move, roll-table) — always proceeds. Roll-table outcome
		// resolution lands in Effort 2.
		return true, nil
	}
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	attack := roller.Roll(10) + statScore(faction, step.Contest.AttackerStat)
	defense := roller.Roll(10) + statScore(target.Faction, step.Contest.DefenderStat)
	return attack > defense, nil
}

// otherFactions returns every faction but the acting one, name-sorted. For a
// faction-target ability every opponent is a valid target.
func otherFactions(factionState *state.FactionState, actingFactionID string) []*domain.Faction {
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
