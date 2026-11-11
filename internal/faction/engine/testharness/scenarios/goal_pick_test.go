package scenarios

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
)

func offeredGoalIDs(offers []domain.GoalOffer) []string {
	goalIDs := make([]string, len(offers))
	for i, offer := range offers {
		goalIDs[i] = offer.Goal.ID
	}
	return goalIDs
}

// eventIndex is the position of the first observer event of kind for the faction,
// or -1 when there is none.
func eventIndex(events []testharness.ObservedEvent, kind, factionID string) int {
	return slices.IndexFunc(events, func(event testharness.ObservedEvent) bool {
		return event.Kind == kind && event.Faction != nil && event.Faction.ID == factionID
	})
}

func TestRunCycle_GoalPick_PicksAGoalWithATarget(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	beta := h.AddFaction("beta", "Tartarus", 2, 2, 2)
	scriptAction(t, h, "")
	destroyTheFoe := goalID(t, h, "Destroy the Foe")

	var alphaOffers []domain.GoalOffer
	h.Collector.SelectGoalFn = func(faction *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error) {
		if faction.ID != "alpha" {
			return nil, nil
		}
		alphaOffers = offers
		return &domain.GoalChoice{GoalID: destroyTheFoe, TargetFactionID: "beta"}, nil
	}

	runOneCycle(t, h)

	offered := offeredGoalIDs(alphaOffers)
	testharness.CheckStep(t, "World Seizure is not offered", !slices.Contains(offered, goalID(t, h, "World Seizure")),
		fmt.Sprintf("offered = %v", offered))
	destroyIndex := slices.Index(offered, destroyTheFoe)
	if destroyIndex < 0 {
		t.Fatalf("Destroy the Foe not offered to alpha: %v", offered)
	}
	var targetIDs []string
	for _, target := range alphaOffers[destroyIndex].Targets {
		targetIDs = append(targetIDs, target.ID)
	}
	testharness.CheckStep(t, "Destroy the Foe targets only beta", slices.Equal(targetIDs, []string{"beta"}),
		fmt.Sprintf("targets = %v", targetIDs))

	if alpha.ActiveGoal == nil {
		t.Fatal("alpha has no active goal after picking")
	}
	testharness.CheckStep(t, "alpha holds Destroy the Foe", alpha.ActiveGoal.GoalID == destroyTheFoe,
		fmt.Sprintf("goal = %s", alpha.ActiveGoal.GoalID))
	testharness.CheckStep(t, "alpha's target is beta", alpha.ActiveGoal.TargetFactionID == "beta",
		fmt.Sprintf("target = %q", alpha.ActiveGoal.TargetFactionID))
	testharness.CheckStep(t, "beta declined and holds no goal", beta.ActiveGoal == nil,
		fmt.Sprintf("beta goal = %+v", beta.ActiveGoal))

	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	record, found := testharness.FindMutationType(records, "goal_picked")
	if !found {
		t.Fatal("no goal_picked mutation in history")
	}
	var picked domain.GoalPicked
	if err := json.Unmarshal(record.Payload, &picked); err != nil {
		t.Fatalf("unmarshalling goal_picked payload: %v", err)
	}
	testharness.CheckStep(t, "goal_picked is alpha's", picked.FactionID == "alpha",
		fmt.Sprintf("faction = %s", picked.FactionID))

	events := h.Observer.Events
	raiseSkipped := eventIndex(events, "StatRaiseSkipped", "alpha")
	goalPicked := eventIndex(events, "GoalPicked", "alpha")
	bookkeeping := eventIndex(events, "BookkeepingApplied", "alpha")
	testharness.CheckStep(t, "GoalPicked follows StatRaiseSkipped and precedes BookkeepingApplied",
		raiseSkipped >= 0 && goalPicked > raiseSkipped && bookkeeping > goalPicked,
		fmt.Sprintf("StatRaiseSkipped=%d GoalPicked=%d BookkeepingApplied=%d", raiseSkipped, goalPicked, bookkeeping))
}

func TestRunCycle_GoalPick_HeldGoalIsNotAsked(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	scriptAction(t, h, "")
	militaryConquest := goalID(t, h, "Military Conquest")
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: militaryConquest}
	h.Collector.SelectGoalFn = func(faction *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
		t.Errorf("SelectGoal called for %s, which holds a goal", faction.ID)
		return nil, nil
	}

	runOneCycle(t, h)

	if alpha.ActiveGoal == nil {
		t.Fatal("alpha lost its goal")
	}
	testharness.CheckStep(t, "alpha still holds Military Conquest", alpha.ActiveGoal.GoalID == militaryConquest,
		fmt.Sprintf("goal = %s", alpha.ActiveGoal.GoalID))
}

func TestRunCycle_GoalPick_DestroyedFactionIsNotAsked(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	h.AddFaction("beta", "Tartarus", 2, 2, 2)
	scriptAction(t, h, "")
	alpha.CurrentHP = 0
	h.Collector.SelectGoalFn = func(faction *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
		if faction.ID == "alpha" {
			t.Error("SelectGoal called for alpha, which is destroyed")
		}
		return nil, nil
	}

	runOneCycle(t, h)

	testharness.CheckStep(t, "alpha holds no goal", alpha.ActiveGoal == nil,
		fmt.Sprintf("goal = %+v", alpha.ActiveGoal))
}

func TestGoalPick_ResumedTurnDoesNotAskAgain(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	h.Collector.SelectActionFn = skipAction
	h.Collector.SelectGoalFn = func(faction *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
		t.Errorf("SelectGoal called for %s on a turn whose bookkeeping already ran", faction.ID)
		return nil, nil
	}

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	h.FactionState.CurrentTurn.BookkeepingApplied = true
	if _, err := h.Engine.RunFactionTurn(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunFactionTurn: %v", err)
	}

	testharness.CheckStep(t, "alpha holds no goal", alpha.ActiveGoal == nil,
		fmt.Sprintf("goal = %+v", alpha.ActiveGoal))
}

func TestRunCycle_GoalPick_DeclineAsksAgainNextTurn(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
	scriptAction(t, h, "")
	asks := 0
	h.Collector.SelectGoalFn = func(_ *domain.Faction, _ []domain.GoalOffer) (*domain.GoalChoice, error) {
		asks++
		return nil, nil
	}

	runOneCycle(t, h)
	runOneCycle(t, h)

	testharness.CheckStep(t, "asked once per turn", asks == 2, fmt.Sprintf("asks = %d", asks))
	testharness.CheckStep(t, "alpha holds no goal", alpha.ActiveGoal == nil,
		fmt.Sprintf("goal = %+v", alpha.ActiveGoal))
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationType(records, "goal_picked")
	testharness.CheckStep(t, "no goal_picked in history", !found, "found a goal_picked mutation")
}

func TestRunCycle_GoalPick_RaiseOpensTheWealthGate(t *testing.T) {
	cases := []struct {
		name        string
		raise       *domain.FactionStat
		wantOffered bool
	}{
		{"raise Wealth", new(domain.StatWealth), true},
		{"decline the raise", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			alpha := h.AddFaction("alpha", "Tartarus", 3, 2, 1)
			alpha.XP = 2 // exactly covers Wealth 1→2: HPValueForRating(2) = 2
			scriptAction(t, h, "")
			wealthOfWorlds := goalID(t, h, "Wealth of Worlds")
			alpha.CompletionRatings = map[string]map[domain.FactionStat]int{
				wealthOfWorlds: {domain.StatForce: 3, domain.StatCunning: 2, domain.StatWealth: 1},
			}
			h.Collector.SelectStatRaiseFn = func(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
				return tc.raise, nil
			}
			var offered []string
			h.Collector.SelectGoalFn = func(_ *domain.Faction, offers []domain.GoalOffer) (*domain.GoalChoice, error) {
				offered = offeredGoalIDs(offers)
				return nil, nil
			}

			runOneCycle(t, h)

			testharness.CheckStep(t, fmt.Sprintf("Wealth of Worlds offered = %v", tc.wantOffered),
				slices.Contains(offered, wealthOfWorlds) == tc.wantOffered,
				fmt.Sprintf("Wealth=%d offered=%v", alpha.Wealth, offered))
		})
	}
}
