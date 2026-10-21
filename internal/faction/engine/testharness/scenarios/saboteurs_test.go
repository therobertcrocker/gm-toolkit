package scenarios

import (
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const (
	defSaboteurs = "SWN-C2-002"
	defInformers = "SWN-C1-002"
)

// saboteursAttack scripts alpha's Saboteurs to attack beta's Informers, and
// records whether each of beta's turns offers Use Asset Ability.
func saboteursAttack(t *testing.T, h *testharness.Harness, saboteurs, informers *domain.Asset, betaOffered *[]bool) {
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID == "beta" {
			*betaOffered = append(*betaOffered, slices.ContainsFunc(available, func(candidate action.Action) bool {
				return candidate.Name() == "Use Asset Ability"
			}))
			return nil, nil
		}
		if len(*betaOffered) > 0 {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Attack" {
				return candidate, nil
			}
		}
		t.Fatal("Attack not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{saboteurs}, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, _ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return informers, nil
	}
}

// TestSaboteursBlockAbilityUntilNextTurn: alpha's Saboteurs attack beta's
// Informers and miss. Beta's turn offers no Use Asset Ability; at alpha's next
// turn start the condition clears, and beta's next turn offers it again.
func TestSaboteursBlockAbilityUntilNextTurn(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	saboteurs := addActiveAsset(alpha, "saboteurs", defSaboteurs, "Hadrian")
	informers := addActiveAsset(beta, "informers", defInformers, "Hadrian")

	// Rolls: attack 1+3 vs defense 10+3, a miss; Informers have no counter.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{1, 10}}
	var betaOffered []bool
	saboteursAttack(t, h, saboteurs, informers, &betaOffered)

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the Informers carry no_ability from the Saboteurs",
		slices.Equal(informers.Conditions, []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: saboteurs.ID, ExpiresAtTurnOf: "alpha"}}),
		"the Informers do not carry the condition")
	testharness.CheckStep(t, "beta's turn offers no Use Asset Ability", slices.Equal(betaOffered, []bool{false}),
		"Use Asset Ability was offered for the sabotaged Informers")

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the condition clears at alpha's next turn start", len(informers.Conditions) == 0,
		"the Informers still carry the condition")
	testharness.CheckStep(t, "beta's next turn offers Use Asset Ability", slices.Equal(betaOffered, []bool{false, true}),
		"Use Asset Ability was not offered once the condition cleared")
}

// TestSaboteursConditionPrecedesALethalHit: the Saboteurs' damage kills the
// Informers. The condition lands before the removal, so the batch applies with
// no miss (F3).
func TestSaboteursConditionPrecedesALethalHit(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	saboteurs := addActiveAsset(alpha, "saboteurs", defSaboteurs, "Hadrian")
	informers := addActiveAsset(beta, "informers", defInformers, "Hadrian")
	informers.CurrentHP = 1

	// Rolls: attack 10+3 vs defense 1+3, a hit; 2d4 damage 2+2 kills the Informers.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 2, 2}}
	var betaOffered []bool
	saboteursAttack(t, h, saboteurs, informers, &betaOffered)

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the Informers are destroyed", beta.Assets[informers.ID] == nil, "the Informers are still in play")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "asset_condition_applied", "block_ability")
	testharness.CheckStep(t, "asset_condition_applied(cause=block_ability) in history", found, "no condition recorded")
}
