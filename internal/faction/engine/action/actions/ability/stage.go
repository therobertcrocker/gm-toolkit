package ability

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// errNoTargets: the step has no candidate left, because an earlier step or asset in
// the same action changed the view. Dispatch halts the asset's chain on it.
var errNoTargets = errors.New("ability: no target left")

func resolveTarget(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	factionState *state.FactionState,
	collector action.Collector,
) (domain.AbilityTarget, error) {
	candidates, picks := stepCandidates(step, faction.ID, asset, factionState)
	if picks && len(candidates.Factions)+len(candidates.Assets) == 0 {
		return domain.AbilityTarget{}, errNoTargets
	}
	switch step.Target {
	case domain.TargetNone:
		return domain.AbilityTarget{Kind: domain.TargetNone}, nil
	case domain.TargetSelf:
		return domain.AbilityTarget{Kind: domain.TargetSelf, Asset: asset, Faction: faction}, nil
	case domain.TargetFaction:
		target, err := collector.SelectAbilityTarget(asset, step, candidates)
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.Faction == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		return target, nil
	case domain.TargetAsset:
		target, err := collector.SelectAbilityTarget(asset, step, candidates)
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.Asset == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		owner := factionState.Factions[target.Asset.OwnerID]
		if owner == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: target owner %s not in state", step.Effect, target.Asset.OwnerID)
		}
		return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: target.Asset, Faction: owner}, nil
	default:
		// deliberate: world targets are not unbuilt-by-oversight — no asset targets a world yet. See domain.TargetWorld.
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
	// deliberate: the real target faction, found by ID, not target.Faction, because the target stage reads the action's view and a defense hook's budget spent on a copy would be lost (Effort 3 Decision 3).
	defender := factionState.Factions[target.Faction.ID]
	if defender == nil {
		return false, fmt.Errorf("contest: target faction %s not in state", target.Faction.ID)
	}
	worldID := asset.Location.WorldID
	// deliberate: a literal, not UseAssetAbility.Name(), because actions imports ability (import cycle).
	const actionName = "Use Asset Ability"
	attack := dispatch.RollWithHooks(
		hooks.RollContext{
			Phase:         hooks.PhaseFactionTest,
			Actor:         faction,
			Opponent:      defender,
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
			Actor:         defender,
			Opponent:      faction,
			Attribute:     string(step.Contest.DefenderStat),
			Asset:         target.Asset,
			OpponentAsset: asset,
			World:         worldID,
			Action:        actionName,
		},
		domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: statScore(defender, step.Contest.DefenderStat)},
		registry, collector, roller, defender, factionState, rulebook,
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

// assetTargetRules narrows a step's asset candidates by its effect. An effect with
// no rule takes every base candidate.
var assetTargetRules = map[domain.AbilityEffectType]func(candidate, acting *domain.Asset) bool{
	domain.EffectDisableAsset: func(candidate, _ *domain.Asset) bool {
		return !candidate.HasCondition(domain.ConditionDisabled)
	},
	domain.EffectBlockAttack: func(candidate, acting *domain.Asset) bool {
		attached := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: acting.ID}
		return !slices.ContainsFunc(candidate.Conditions, attached.Is)
	},
}

// assetCandidates returns other factions' unstealthed, unsheltered assets on the
// acting asset's world, in faction-ID then asset-ID order, that pass the effect's
// rule. An asset in transit has no world and no candidates.
func assetCandidates(effect domain.AbilityEffectType, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) []*domain.Asset {
	world := acting.Location.WorldID
	if world == "" {
		return nil
	}
	rule := assetTargetRules[effect]
	var candidates []*domain.Asset
	for _, id := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if id == actingFactionID {
			continue
		}
		for _, candidate := range domain.SortedAssets(factionState.Factions[id]) {
			// deliberate: skip a sheltered asset, not offer it, because Boltholes make it untouchable to every ability, not only to Attack (Effort 2 Decision 11).
			if candidate.Location.WorldID != world || candidate.Stealthy || candidate.HasCondition(domain.ConditionSheltered) {
				continue
			}
			if rule == nil || rule(candidate, acting) {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

// HasTargets reports whether the first step of def's ability can pick a target
// now. A step that picks no target always can.
func HasTargets(faction *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, factionState *state.FactionState) bool {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return true
	}
	// deliberate: only the first step, not every step, because a later step's candidates depend on what earlier steps change (Seductress moves before it reveals).
	candidates, picks := stepCandidates(def.Ability.Steps[0], faction.ID, asset, factionState)
	return !picks || len(candidates.Factions)+len(candidates.Assets) > 0
}

// stepCandidates returns the candidates for a step that picks a faction or an
// asset. picks is false for a step that picks neither.
func stepCandidates(step domain.AbilityStep, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) (candidates domain.AbilityTargetCandidates, picks bool) {
	switch step.Target {
	case domain.TargetFaction:
		return domain.AbilityTargetCandidates{Factions: factionCandidates(step.Effect, actingFactionID, acting, factionState)}, true
	case domain.TargetAsset:
		return domain.AbilityTargetCandidates{Assets: assetCandidates(step.Effect, actingFactionID, acting, factionState)}, true
	}
	return domain.AbilityTargetCandidates{}, false
}

// factionTargetRules narrows a step's faction candidates by its effect. An effect
// with no rule takes every other faction (Informers).
var factionTargetRules = map[domain.AbilityEffectType]func(candidate *domain.Faction, acting *domain.Asset) bool{
	domain.EffectLevyCoin: func(candidate *domain.Faction, acting *domain.Asset) bool {
		world := acting.Location.WorldID
		return world != "" && slices.ContainsFunc(domain.SortedAssets(candidate), func(asset *domain.Asset) bool {
			return asset.Location.WorldID == world && !asset.Stealthy && !asset.HasCondition(domain.ConditionSheltered)
		})
	},
}

// factionCandidates returns every other faction, name-sorted, that passes the
// effect's rule.
func factionCandidates(effect domain.AbilityEffectType, actingFactionID string, acting *domain.Asset, factionState *state.FactionState) []*domain.Faction {
	rule := factionTargetRules[effect]
	var candidates []*domain.Faction
	for _, candidate := range otherFactions(factionState, actingFactionID) {
		if rule == nil || rule(candidate, acting) {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}
