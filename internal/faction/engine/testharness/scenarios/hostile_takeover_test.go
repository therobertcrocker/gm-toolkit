package scenarios

import (
	"encoding/json"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

const defHostileTakeover = "SWN-W7-002"

// coinGainsByFaction counts the coin_gain Coin changes in records, by faction.
func coinGainsByFaction(t *testing.T, records []domain.EventRecord) map[string]int {
	t.Helper()
	gains := map[string]int{}
	for _, record := range records {
		for _, mutation := range record.Mutations {
			if mutation.Type != "coin_delta" {
				continue
			}
			var coin domain.CoinDelta
			if err := json.Unmarshal(mutation.Payload, &coin); err != nil {
				t.Fatalf("parsing coin_delta: %v", err)
			}
			if coin.Cause == string(domain.SpecialCoinGain) {
				gains[coin.FactionID]++
			}
		}
	}
	return gains
}

// TestHostileTakeoverSeizesAndPaysTheNewOwner: alpha's Hostile Takeover deals a
// killing blow to beta's Party Machine on Hadrian. The Party Machine joins alpha
// at 1 HP with no removal, and the next cycle its turn-start income pays alpha,
// not beta (Effort 2 Decision 16).
func TestHostileTakeoverSeizesAndPaysTheNewOwner(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	beta := h.AddFaction("beta", "Hadrian", 4, 3, 2)
	takeover := addActiveAsset(alpha, "takeover", defHostileTakeover, "Hadrian")
	partyMachine := addActiveAsset(beta, "party", testharness.DefPartyMachine, "Hadrian")
	partyMachine.CurrentHP = 3

	// Rolls: attack 10+2 vs defense 1+2, a hit; damage 2d10 rolls 3 and 3, so 6.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3, 3}}
	h.Collector.SelectActionFn = takeAction(t, alpha.ID, "Attack")
	h.Collector.SelectAttackersFn = func([]*domain.Asset, *rulebook.Rulebook) ([]*domain.Asset, error) {
		return []*domain.Asset{takeover}, nil
	}
	h.Collector.SelectDefenderFn = func(*domain.Asset, []*domain.Asset, *rulebook.Rulebook) (*domain.Asset, error) {
		return partyMachine, nil
	}

	runOrderedCycle(t, h, alpha.ID, beta.ID)

	seized, ok := alpha.Assets[partyMachine.ID]
	testharness.CheckStep(t, "the Party Machine is alpha's at 1 HP",
		ok && seized.OwnerID == alpha.ID && seized.CurrentHP == 1 && beta.Assets[partyMachine.ID] == nil,
		"the Party Machine did not move to alpha at 1 HP")
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	removals := 0
	for _, record := range records {
		for _, mutation := range record.Mutations {
			if mutation.Type == "asset_removed" {
				removals++
			}
		}
	}
	testharness.CheckStep(t, "history has no asset_removed", removals == 0, "an asset was removed")
	firstCycle := len(records)

	h.Collector.SelectActionFn = skipAction
	runOrderedCycle(t, h, alpha.ID, beta.ID)

	gains := coinGainsByFaction(t, testharness.ReadHistory(t, h.Paths.HistoryPath)[firstCycle:])
	testharness.CheckStep(t, "the next cycle's income pays alpha", gains[alpha.ID] == 1, "alpha got no coin_gain")
	testharness.CheckStep(t, "and not beta", gains[beta.ID] == 0, "beta still got coin_gain")
}
