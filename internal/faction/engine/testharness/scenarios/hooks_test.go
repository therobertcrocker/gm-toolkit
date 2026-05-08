package scenarios

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- scenario 9: MutationReactor dispatch — CoinDelta on asset destroyed ---

// coinOnDestroyReactor emits one CoinDelta the first time it observes an
// AssetRemoved in the mutation slice. Stateful so it does not double-fire on
// recursive dispatch rounds.
type coinOnDestroyReactor struct {
	factionID string
	fired     bool
}

func (r *coinOnDestroyReactor) OnMutations(mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if r.fired {
		return nil
	}
	for _, m := range mutations {
		if _, ok := m.(domain.AssetRemoved); ok {
			r.fired = true
			return []domain.Mutation{domain.CoinDelta{FactionID: r.factionID, Delta: 1}}
		}
	}
	return nil
}

func TestMutationReactorDispatch_CoinOnAssetDestroyed(t *testing.T) {
	h := testharness.NewHarness(t, testDataDir)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	h.AddFaction("beta", "Tartarus", 2, 2, 2)

	// Roll sequence: attack hits and destroys beta's asset.
	// attack=10+force(4)=14, defense=1+force(2)=3 → hit; damage=3+1=4 ≥ HP(3) → destroyed.
	h.Engine.Rand = &testharness.FixedRoller{Values: []int{10, 1, 3}}

	stub := &coinOnDestroyReactor{factionID: "alpha"}
	h.Engine.Hooks.RegisterMutationReactor(hooks.FactionScope("alpha"), "stub-coin-on-destroy", stub)

	h.Collector.SelectActionFn = func(faction *domain.Faction, available []action.Action) (action.Action, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		for _, a := range available {
			if a.Name() == "Attack" {
				return a, nil
			}
		}
		t.Fatal("Attack not available for alpha")
		return nil, nil
	}
	h.Collector.SelectAttackersFn = func(eligible []*domain.Asset, _ *rulebook.Rulebook) ([]*domain.Asset, error) {
		return eligible, nil
	}
	h.Collector.SelectDefenderFn = func(_ *domain.Asset, eligible []*domain.Asset, _ *rulebook.Rulebook) (*domain.Asset, error) {
		return eligible[0], nil
	}

	initialCoin := alpha.Coin

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Cfg, h.Collector, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "beta asset destroyed", len(h.FactionState.Factions["beta"].Assets) == 0,
		"beta should have no assets after attack")
	testharness.CheckStep(t, "reactor fired on AssetRemoved", stub.fired,
		"reactor never saw AssetRemoved in mutation slice")

	// alpha income: wealth(2)/2=1 + (force(4)+cunning(3))/4=1 = 2; reactor: +1 → total +3
	wantCoin := initialCoin + 3
	testharness.CheckStep(t, "alpha Coin reflects reactor delta (+1 above bookkeeping income)", alpha.Coin == wantCoin,
		"alpha.Coin mismatch — reactor CoinDelta may not have been applied")

	records := testharness.ReadHistory(t, h.Cfg.HistoryPath)
	_, hasRemoved := testharness.FindMutationType(records, "asset_removed")
	testharness.CheckStep(t, "asset_removed in history", hasRemoved, "no asset_removed mutation found")
	// reactor's CoinDelta has no Cause; bookkeeping uses Cause="bookkeeping"
	_, hasReactorCoin := testharness.FindMutationByTypeAndCause(records, "coin_delta", "")
	testharness.CheckStep(t, "reactor coin_delta (cause='') in history", hasReactorCoin, "no reactor coin_delta in history")
}
