package scenarios

import (
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defMedicalCenter = "SWN-W4-003"

// setupMedicalCenterFight puts beta's Medical Centers and two Security Personnel
// (Military Units, 3 HP, cost 2) on Hadrian with alpha's two attackers. Each
// attacker hits one unit for 4, a killing blow. The caller sets ConfirmPaymentFn
// and runs the cycle.
func setupMedicalCenterFight(t *testing.T, medicalCenters int, coin int) (h *testharness.Harness, beta *domain.Faction, first, second *domain.Asset) {
	t.Helper()
	h = testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta = h.AddFaction("beta", "Hadrian", 4, 3, 2)
	beta.Coin = coin
	firstAttacker := addActiveAsset(alpha, "attacker-1", testharness.DefSecurityPersonnel, "Hadrian")
	secondAttacker := addActiveAsset(alpha, "attacker-2", testharness.DefSecurityPersonnel, "Hadrian")
	for index := range medicalCenters {
		addActiveAsset(beta, fmt.Sprintf("mc-%d", index), defMedicalCenter, "Hadrian")
	}
	first = beta.Assets["beta-asset-1"]
	second = addActiveAsset(beta, "unit-2", testharness.DefSecurityPersonnel, "Hadrian")
	second.CurrentHP = 3

	// Rolls per attacker: attack 10+4 vs defense 1+4, a hit; damage 1d3+1 rolls 3, so 4.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3, 10, 1, 3}}
	h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Attack")
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{firstAttacker, secondAttacker}, nil
	}
	h.Collector.SelectDefenderFn = func(attacker *domain.Asset, _ []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		if attacker.ID == firstAttacker.ID {
			return first, nil
		}
		return second, nil
	}
	return h, beta, first, second
}

// TestMedicalCenterRestoresOncePerTurn: both of beta's units take a killing blow.
// The one Medical Center restores the first at 1 HP for its half price; its budget
// is spent, so the second is lost with no prompt (Effort 2 Decision 15).
func TestMedicalCenterRestoresOncePerTurn(t *testing.T) {
	h, beta, first, second := setupMedicalCenterFight(t, 1, 10)
	var offers []hooks.PaymentOffer
	h.Collector.ConfirmPaymentFn = func(offer hooks.PaymentOffer) (bool, error) {
		offers = append(offers, offer)
		return true, nil
	}

	runOrderedCycle(t, h, "alpha", beta.ID)

	testharness.CheckStep(t, "one restore offer at the unit's half price",
		len(offers) == 1 && offers[0].Reason == hooks.PaymentRestore && offers[0].Amount == 1,
		"want one PaymentRestore offer of 1 Coin")
	testharness.CheckStep(t, "the first unit is beta's at 1 HP", beta.Assets[first.ID] == first && first.CurrentHP == 1,
		"the first unit was removed or its HP is not 1")
	testharness.CheckStep(t, "the second unit is removed", beta.Assets[second.ID] == nil, "the second unit is still in play")
}

// TestTwoMedicalCentersCountTheBatchCoin: two Medical Centers can each restore once,
// but beta's 1 Coin pays for one restore only. The dispatcher counts the Coin the
// batch already spends, so the second unit is lost with no prompt.
func TestTwoMedicalCentersCountTheBatchCoin(t *testing.T) {
	h, beta, first, second := setupMedicalCenterFight(t, 2, 1)
	var offers []hooks.PaymentOffer
	h.Collector.ConfirmPaymentFn = func(offer hooks.PaymentOffer) (bool, error) {
		offers = append(offers, offer)
		return true, nil
	}

	runOrderedCycle(t, h, "alpha", beta.ID)

	testharness.CheckStep(t, "one restore offer", len(offers) == 1, "want one offer")
	testharness.CheckStep(t, "the first unit is beta's at 1 HP", beta.Assets[first.ID] == first && first.CurrentHP == 1,
		"the first unit was removed or its HP is not 1")
	testharness.CheckStep(t, "the second unit is removed", beta.Assets[second.ID] == nil, "the second unit is still in play")
}
