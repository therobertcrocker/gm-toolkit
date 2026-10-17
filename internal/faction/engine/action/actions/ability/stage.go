package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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
		// deliberate: asset/world targets await A.001.3 (Marketers/Seditionists),
		// not unbuilt-by-oversight — see domain.TargetAsset/TargetWorld.
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
	}
}

// passesGate rolls the step's contest through the roll hooks. faction is the real
// acting faction, so a budget a hook spends persists past this step.
func passesGate(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	target domain.AbilityTarget,
	registry *hooks.Registry,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) (bool, error) {
	if step.Contest == nil {
		// Ungated (self-move, roll-table) — always proceeds. Roll-table outcome
		// resolution lands in Effort 2.
		return true, nil
	}
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	worldID := asset.Location.WorldID
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	attack := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:         hooks.PhaseFactionTest,
			Actor:         faction,
			Opponent:      target.Faction,
			Attribute:     string(step.Contest.AttackerStat),
			Asset:         asset,
			OpponentAsset: target.Asset,
			World:         worldID,
			Action:        actionName,
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(faction, step.Contest.AttackerStat)},
		registry, collector, roller, faction, factionState, rulebook,
	)
	defense := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:         hooks.PhaseContested,
			Actor:         target.Faction,
			Opponent:      faction,
			Attribute:     string(step.Contest.DefenderStat),
			Asset:         target.Asset,
			OpponentAsset: asset,
			World:         worldID,
			Action:        actionName,
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(target.Faction, step.Contest.DefenderStat)},
		registry, collector, roller, target.Faction, factionState, rulebook,
	)
	return attack.Sum > defense.Sum, nil
}

// resolveRoll evaluates the step's roll table, or returns 0 for a non-roll step.
// The Contest gate and the roll table are mutually exclusive (parser-enforced), so
// a step that reaches a handler has at most one source of randomness.
func resolveRoll(step domain.AbilityStep, roller domain.Roller) int {
	if step.Roll == nil {
		return 0
	}
	return step.Roll.Die.Roll(roller)
}

// outcomeFor returns the first outcome whose [From,To] range contains roll. An
// unmatched roll yields (zero, false) — the absence of a row is a zero payout.
func outcomeFor(roll int, outcomes []domain.AbilityOutcome) (domain.AbilityOutcome, bool) {
	for _, outcome := range outcomes {
		if roll >= outcome.From && roll <= outcome.To {
			return outcome, true
		}
	}
	return domain.AbilityOutcome{}, false
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
	return faction.StatScore(stat)
}
