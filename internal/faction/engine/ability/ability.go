package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// StepHandler resolves one ability step for an asset.
type StepHandler func(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *loader.Rulebook,
) ([]domain.Mutation, error)

// CustomAbilityHandler is a full override for a specific asset definition ID.
type CustomAbilityHandler func(
	faction *domain.Faction,
	asset *domain.Asset,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *loader.Rulebook,
) ([]domain.Mutation, error)

type AbilityEngine struct {
	stepHandlers   map[domain.AbilityStepType]StepHandler
	customHandlers map[string]CustomAbilityHandler
}

func New() *AbilityEngine {
	ae := &AbilityEngine{
		stepHandlers:   make(map[domain.AbilityStepType]StepHandler),
		customHandlers: make(map[string]CustomAbilityHandler),
	}
	ae.stepHandlers[domain.AbilityStepMovement] = movementStepHandler
	ae.stepHandlers[domain.AbilityStepFactionTest] = factionTestStepHandler
	return ae
}

// RegisterCustomHandler registers a bespoke handler for a specific asset definition ID.
func (ae *AbilityEngine) RegisterCustomHandler(defID string, handler CustomAbilityHandler) {
	ae.customHandlers[defID] = handler
}

// Run resolves an asset's ability. Returns nil, nil when the definition has no Ability.
func (ae *AbilityEngine) Run(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *loader.Rulebook,
) ([]domain.Mutation, error) {
	if handler, ok := ae.customHandlers[def.ID]; ok {
		return handler(faction, asset, collector, roller, factionState, rulebook)
	}
	if def.Ability == nil {
		return nil, nil
	}
	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		handler, ok := ae.stepHandlers[step.Type]
		if !ok {
			return nil, fmt.Errorf("no handler for ability step type %q", step.Type)
		}
		stepMutations, err := handler(faction, asset, step, collector, roller, factionState, rulebook)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
	}
	return mutations, nil
}

func movementStepHandler(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	_ domain.Roller,
	factionState *state.FactionState,
	_ *loader.Rulebook,
) ([]domain.Mutation, error) {
	destination, err := collector.SelectMoveDestination(asset, worldsFromState(factionState))
	if err != nil {
		return nil, err
	}
	var mutations []domain.Mutation
	if step.CoinCost > 0 {
		mutations = append(mutations, domain.CoinDelta{FactionID: faction.ID, Delta: -step.CoinCost, Cause: "ability", CausedByFactionID: faction.ID})
	}
	mutations = append(mutations, domain.AssetMoved{
		FactionID:         faction.ID,
		AssetID:           asset.ID,
		FromLocation:      asset.Location,
		ToLocation:        destination,
		Cause:             "ability",
		CausedByFactionID: faction.ID,
	})
	return mutations, nil
}

func factionTestStepHandler(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	_ *loader.Rulebook,
) ([]domain.Mutation, error) {
	candidates := factionTestCandidates(factionState, faction.ID, asset.Location, step.Effect)
	targetFaction, err := collector.SelectFactionTestTarget(asset, step.Effect, candidates)
	if err != nil {
		return nil, err
	}
	if targetFaction == nil {
		return nil, fmt.Errorf("faction test: no target faction selected")
	}

	attackRoll := roller.Roll(10) + abilityStatScore(faction, step.AttackerStat)
	defenseRoll := roller.Roll(10) + abilityStatScore(targetFaction, step.DefenderStat)

	if attackRoll <= defenseRoll {
		return nil, nil
	}

	return applyAbilityEffect(faction, asset, step, targetFaction, roller)
}

func factionTestCandidates(factionState *state.FactionState, actingFactionID, world string, effect domain.AbilityEffectType) []*domain.Faction {
	var candidates []*domain.Faction
	for id, faction := range factionState.Factions {
		if id == actingFactionID {
			continue
		}
		if effect == domain.EffectRevealStealth {
			candidates = append(candidates, faction)
			continue
		}
		for _, asset := range faction.Assets {
			if asset.Location == world {
				candidates = append(candidates, faction)
				break
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

func applyAbilityEffect(
	actingFaction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	targetFaction *domain.Faction,
	roller domain.Roller,
) ([]domain.Mutation, error) {
	switch step.Effect {
	case domain.EffectRevealStealth:
		var mutations []domain.Mutation
		for _, targetAsset := range targetFaction.Assets {
			if targetAsset.Location == asset.Location && targetAsset.Stealthy {
				mutations = append(mutations, domain.AssetStealthCleared{
					FactionID:         targetFaction.ID,
					AssetID:           targetAsset.ID,
					Cause:             "ability",
					CausedByFactionID: actingFaction.ID,
				})
			}
		}
		return mutations, nil
	case domain.EffectCoinDrain:
		if step.EffectDice == nil {
			return nil, fmt.Errorf("effect %q requires effect_dice", step.Effect)
		}
		amount := step.EffectDice.Roll(roller)
		return []domain.Mutation{
			domain.CoinDelta{FactionID: targetFaction.ID, Delta: -amount, Cause: "ability", CausedByFactionID: actingFaction.ID},
		}, nil
	case domain.EffectCoinSteal:
		if step.EffectDice == nil {
			return nil, fmt.Errorf("effect %q requires effect_dice", step.Effect)
		}
		amount := step.EffectDice.Roll(roller)
		return []domain.Mutation{
			domain.CoinDelta{FactionID: targetFaction.ID, Delta: -amount, Cause: "ability", CausedByFactionID: actingFaction.ID},
			domain.CoinDelta{FactionID: actingFaction.ID, Delta: amount, Cause: "ability", CausedByFactionID: actingFaction.ID},
		}, nil
	default:
		return nil, fmt.Errorf("unknown ability effect %q", step.Effect)
	}
}

func worldsFromState(factionState *state.FactionState) []string {
	seen := map[string]bool{}
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			if asset.Location != "" {
				seen[asset.Location] = true
			}
		}
		for _, base := range faction.Bases {
			if base.Location != "" {
				seen[base.Location] = true
			}
		}
	}
	worlds := make([]string, 0, len(seen))
	for world := range seen {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)
	worlds = append(worlds, "Astral Sea")
	return worlds
}

func abilityStatScore(faction *domain.Faction, stat domain.FactionStat) int {
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
