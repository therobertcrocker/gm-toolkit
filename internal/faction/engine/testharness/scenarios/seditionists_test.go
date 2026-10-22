package scenarios

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defSeditionists = "SWN-C4-004"

// TestSeditionistsBlockAndMove: alpha's Seditionists attach to beta's Saboteurs on
// Tartarus, so beta's Attack does not offer them. In cycle 2 they attach to beta's
// Informers instead; the Saboteurs' block is cleared and beta's Attack offers them
// again. Beta takes Attack each cycle only to read its attacker candidates, and
// selects none.
func TestSeditionistsBlockAndMove(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	alpha.Coin = 10
	seditionists := addActiveAsset(alpha, "seditionists", defSeditionists, "Tartarus")
	saboteurs := addActiveAsset(beta, "saboteurs", defSaboteurs, "Tartarus")
	informers := addActiveAsset(beta, "informers", defInformers, "Tartarus")

	// Rolls: the 1d4 cost, 2 in cycle 1 and 3 in cycle 2.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{2, 3}}
	target := saboteurs
	var offeredAttackers []string
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		name := "Use Asset Ability"
		if faction.ID == "beta" {
			name = "Attack"
		}
		for _, candidate := range available {
			if candidate.Name() == name {
				return candidate, nil
			}
		}
		t.Fatalf("%s not available for %s", name, faction.ID)
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func(candidates []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		offeredAttackers = offeredAttackers[:0]
		for _, candidate := range candidates {
			offeredAttackers = append(offeredAttackers, candidate.ID)
		}
		return nil, nil
	}
	h.Collector.SelectAbilityAssetsFn = func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{seditionists}, nil
	}
	h.Collector.SelectAbilityTargetFn = func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
		for _, candidate := range candidates.Assets {
			if candidate.ID == target.ID {
				return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidate}, nil
			}
		}
		t.Fatalf("%s is not a Seditionists candidate", target.ID)
		return domain.AbilityTarget{}, nil
	}

	runOrderedCycle(t, h, "alpha", "beta")

	attached := domain.Condition{Kind: domain.ConditionNoAttack, SourceAssetID: seditionists.ID}
	testharness.CheckStep(t, "the Saboteurs carry the block", slices.Equal(saboteurs.Conditions, []domain.Condition{attached}),
		"the Saboteurs do not carry the no_attack condition")
	testharness.CheckStep(t, "beta's Attack does not offer the Saboteurs", !slices.Contains(offeredAttackers, saboteurs.ID),
		"the blocked Saboteurs were offered as an attacker")

	target = informers
	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the Saboteurs' block is cleared", len(saboteurs.Conditions) == 0, "the Saboteurs still carry the condition")
	testharness.CheckStep(t, "the Informers carry the block", slices.Equal(informers.Conditions, []domain.Condition{attached}),
		"the Informers do not carry the no_attack condition")
	testharness.CheckStep(t, "beta's Attack offers the Saboteurs again", slices.Contains(offeredAttackers, saboteurs.ID),
		"the cleared Saboteurs were not offered as an attacker")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	testharness.CheckStep(t, "alpha paid the roll on each use", countByTypeAndCause(records, "coin_delta", "ability") == 2,
		"history does not hold two coin_delta(cause=ability)")
}
