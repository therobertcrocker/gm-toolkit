package scenarios

import (
	"errors"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defMarketers = "SWN-W5-001"

// TestMarketersDisableThenLift: alpha's Marketers win their contest against beta's
// Party Machine. Beta declines both offers in cycle 1 (avoid the condition, then lift
// it at its turn start), so the Party Machine pays no income. In cycle 2 beta pays the
// lift at its turn start, and the Party Machine's income fires in that turn
// (Effort 3 Decision 8).
func TestMarketersDisableThenLift(t *testing.T) {
	h, marketers, partyMachine := setupMarketersOnPartyMachine(t)
	var reasons []hooks.PaymentReason
	h.Collector.ConfirmPaymentFn = func(offer hooks.PaymentOffer) (bool, error) {
		reasons = append(reasons, offer.Reason)
		return len(reasons) > 2, nil
	}

	runOrderedCycle(t, h, "alpha", "beta")

	disabled := domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: marketers.ID}
	testharness.CheckStep(t, "the Party Machine is disabled", slices.Equal(partyMachine.Conditions, []domain.Condition{disabled}),
		"the Party Machine does not carry the condition")
	testharness.CheckStep(t, "beta declined to avoid it, then to lift it",
		slices.Equal(reasons, []hooks.PaymentReason{hooks.PaymentAvoidDisable, hooks.PaymentLiftDisable}),
		"the offers were not avoid_disable then lift_disable")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	testharness.CheckStep(t, "the Party Machine paid no income", countByTypeAndCause(records, "coin_delta", string(domain.SpecialCoinGain)) == 0,
		"the disabled Party Machine paid income")

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the condition is lifted", len(partyMachine.Conditions) == 0, "the Party Machine still carries the condition")
	records = testharness.ReadHistory(t, h.Paths.HistoryPath)
	testharness.CheckStep(t, "the lift is a coin_delta(cause=condition_lifted)", countByTypeAndCause(records, "coin_delta", "condition_lifted") == 1,
		"no condition_lifted payment in history")
	testharness.CheckStep(t, "the Party Machine's income fired in the lift turn", countByTypeAndCause(records, "coin_delta", string(domain.SpecialCoinGain)) == 1,
		"the lifted Party Machine paid no income this turn")
}

// TestMarketersCanceledLiftKeepsTheTurn: a canceled lift prompt at beta's turn start
// is reported, lifts nothing and charges nothing, and beta's turn goes on.
func TestMarketersCanceledLiftKeepsTheTurn(t *testing.T) {
	h, _, partyMachine := setupMarketersOnPartyMachine(t)
	h.Collector.ConfirmPaymentFn = func(offer hooks.PaymentOffer) (bool, error) {
		if offer.Reason == hooks.PaymentLiftDisable {
			return false, action.ErrTurnCanceled
		}
		return false, nil
	}

	runOrderedCycle(t, h, "alpha", "beta")

	testharness.CheckStep(t, "the Party Machine is still disabled", partyMachine.HasCondition(domain.ConditionDisabled),
		"the canceled lift cleared the condition")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	testharness.CheckStep(t, "nothing was paid for a lift", countByTypeAndCause(records, "coin_delta", "condition_lifted") == 0,
		"a condition_lifted payment is in history")
	var reported, completed bool
	for _, event := range h.Observer.Events {
		if event.Kind == "Error" && event.Faction.ID == "beta" {
			err, _ := event.Payload.(error)
			reported = errors.Is(err, action.ErrTurnCanceled)
		}
		if event.Kind == "TurnCompleted" && event.Faction.ID == "beta" {
			completed = true
		}
	}
	testharness.CheckStep(t, "the cancel was reported for beta", reported, "no ErrTurnCanceled reported for beta")
	testharness.CheckStep(t, "beta's turn completed", completed, "beta's turn did not complete")
}

// setupMarketersOnPartyMachine gives alpha active Marketers and beta (10 Coin) an
// active Party Machine, both on Hadrian. Alpha's first action uses the Marketers on
// the Party Machine and wins the contest. The caller sets ConfirmPaymentFn.
func setupMarketersOnPartyMachine(t *testing.T) (*testharness.Harness, *domain.Asset, *domain.Asset) {
	t.Helper()
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	beta.Coin = 10
	marketers := addActiveAsset(alpha, "marketers", defMarketers, "Hadrian")
	partyMachine := addActiveAsset(beta, "party", testharness.DefPartyMachine, "Hadrian")

	// Rolls: Cunning 10+3 vs Wealth 1+2, a win.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1}}
	alphaActed := false
	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" || alphaActed {
			return nil, nil
		}
		for _, candidate := range available {
			if candidate.Name() == "Use Asset Ability" {
				alphaActed = true
				return candidate, nil
			}
		}
		t.Fatal("Use Asset Ability not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAbilityAssetsFn = func(*domain.Faction, []*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{marketers}, nil
	}
	h.Collector.SelectAbilityTargetFn = func(_ *domain.Asset, _ domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
		for _, candidate := range candidates.Assets {
			if candidate.ID == partyMachine.ID {
				return domain.AbilityTarget{Kind: domain.TargetAsset, Asset: candidate}, nil
			}
		}
		t.Fatal("the Party Machine is not a Marketers candidate")
		return domain.AbilityTarget{}, nil
	}
	return h, marketers, partyMachine
}
