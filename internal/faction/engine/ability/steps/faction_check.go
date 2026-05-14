package steps

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func FactionTestStepHandler(
	faction *domain.Faction,
	asset *domain.Asset,
	step domain.AbilityStep,
	collector Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	_ *rulebook.Rulebook,
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
