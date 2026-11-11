package goal

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func offeredGoalIDs(offers []domain.GoalOffer) []string {
	goalIDs := make([]string, len(offers))
	for i, offer := range offers {
		goalIDs[i] = offer.Goal.ID
	}
	return goalIDs
}

func findOffer(offers []domain.GoalOffer, goalID string) *domain.GoalOffer {
	for i := range offers {
		if offers[i].Goal.ID == goalID {
			return &offers[i]
		}
	}
	return nil
}

func targetIDs(offer *domain.GoalOffer) []string {
	var factionIDs []string
	for _, target := range offer.Targets {
		factionIDs = append(factionIDs, target.ID)
	}
	return factionIDs
}

func TestOffers_ShippedGoals(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine := newEngine(t, fixture.rulebook)
	worldSeizure := shippedGoal(t, fixture.rulebook, "World Seizure").goal.ID
	destroyTheFoe := shippedGoal(t, fixture.rulebook, "Destroy the Foe").goal.ID

	offers := goalEngine.Offers(fixture.holder, fixture.factionState)

	goalIDs := offeredGoalIDs(offers)
	if len(goalIDs) != 10 {
		t.Fatalf("offers = %v, want 10", goalIDs)
	}
	if !slices.IsSorted(goalIDs) {
		t.Errorf("offers = %v, want goal ID order", goalIDs)
	}
	if slices.Contains(goalIDs, worldSeizure) {
		t.Errorf("World Seizure is on offer without Seize Planet registered")
	}
	for _, offer := range offers {
		want := []string(nil)
		if offer.Goal.ID == destroyTheFoe {
			want = []string{"f2", "f3"}
		}
		if got := targetIDs(&offer); !reflect.DeepEqual(got, want) {
			t.Errorf("%s targets = %v, want %v", offer.Goal.Name, got, want)
		}
	}
}

func TestOffers_NeedsActions(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine, err := New(fixture.rulebook, append(slices.Clone(testActionNames), "Seize Planet"), discardLog())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	worldSeizure := shippedGoal(t, fixture.rulebook, "World Seizure").goal.ID

	offers := goalEngine.Offers(fixture.holder, fixture.factionState)

	if findOffer(offers, worldSeizure) == nil {
		t.Errorf("offers = %v, want World Seizure", offeredGoalIDs(offers))
	}
}

func TestOffers_TargetsAreLivingRivals(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine := newEngine(t, fixture.rulebook)
	destroyTheFoe := shippedGoal(t, fixture.rulebook, "Destroy the Foe").goal.ID

	fixture.third.CurrentHP = 0
	offer := findOffer(goalEngine.Offers(fixture.holder, fixture.factionState), destroyTheFoe)
	if offer == nil {
		t.Fatal("Destroy the Foe is not on offer with one rival standing")
	}
	if got := targetIDs(offer); !reflect.DeepEqual(got, []string{"f2"}) {
		t.Errorf("targets = %v, want [f2]", got)
	}

	fixture.rival.CurrentHP = 0
	offers := goalEngine.Offers(fixture.holder, fixture.factionState)
	if findOffer(offers, destroyTheFoe) != nil {
		t.Error("Destroy the Foe is on offer with no rival standing")
	}
	if len(offers) != 9 {
		t.Errorf("offers = %v, want the other 9", offeredGoalIDs(offers))
	}
}

func TestOffers_WealthGate(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine := newEngine(t, fixture.rulebook)
	wealthOfWorlds := shippedGoal(t, fixture.rulebook, "Wealth of Worlds").goal.ID
	militaryConquest := shippedGoal(t, fixture.rulebook, "Military Conquest").goal.ID

	tests := []struct {
		name    string
		record  map[string]map[domain.FactionStat]int
		offered bool
	}{
		{"no record", nil, true},
		{"recorded at the same Wealth", map[string]map[domain.FactionStat]int{wealthOfWorlds: {domain.StatWealth: 2}}, false},
		{"recorded at a lower Wealth", map[string]map[domain.FactionStat]int{wealthOfWorlds: {domain.StatWealth: 1}}, true},
		{"a record for another goal only", map[string]map[domain.FactionStat]int{militaryConquest: {domain.StatWealth: 2}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture.holder.CompletionRatings = tt.record
			offers := goalEngine.Offers(fixture.holder, fixture.factionState)
			if got := findOffer(offers, wealthOfWorlds) != nil; got != tt.offered {
				t.Errorf("Wealth of Worlds offered = %v, want %v", got, tt.offered)
			}
		})
	}
}

func TestPick(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine := newEngine(t, fixture.rulebook)
	destroyTheFoe := shippedGoal(t, fixture.rulebook, "Destroy the Foe").goal.ID
	militaryConquest := shippedGoal(t, fixture.rulebook, "Military Conquest").goal.ID
	worldSeizure := shippedGoal(t, fixture.rulebook, "World Seizure").goal.ID
	offers := goalEngine.Offers(fixture.holder, fixture.factionState)

	tests := []struct {
		name    string
		choice  domain.GoalChoice
		want    domain.GoalPicked
		wantErr string
	}{
		{
			name:   "Destroy the Foe with a target on offer",
			choice: domain.GoalChoice{GoalID: destroyTheFoe, TargetFactionID: "f2"},
			want:   domain.GoalPicked{FactionID: "f1", GoalID: destroyTheFoe, TargetFactionID: "f2"},
		},
		{
			name:    "Destroy the Foe with no target",
			choice:  domain.GoalChoice{GoalID: destroyTheFoe},
			wantErr: "needs a target faction",
		},
		{
			name:    "Destroy the Foe with a target not on offer",
			choice:  domain.GoalChoice{GoalID: destroyTheFoe, TargetFactionID: "f9"},
			wantErr: "needs a target faction",
		},
		{
			name:    "Military Conquest with a target",
			choice:  domain.GoalChoice{GoalID: militaryConquest, TargetFactionID: "f2"},
			wantErr: "takes no target",
		},
		{
			name:    "World Seizure, not on offer",
			choice:  domain.GoalChoice{GoalID: worldSeizure},
			wantErr: "not on offer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goalEngine.Pick(fixture.holder, offers, tt.choice)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Pick() error = %v, want one that names %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Pick() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pick() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPick_RecordsStealthedAtStart(t *testing.T) {
	fixture := newGoalFixture(t, "Military Conquest", 0)
	goalEngine := newEngine(t, fixture.rulebook)
	insideEnemyTerritory := shippedGoal(t, fixture.rulebook, "Inside Enemy Territory").goal.ID
	militaryConquest := shippedGoal(t, fixture.rulebook, "Military Conquest").goal.ID
	fixture.give(fixture.holder, "a2", "cunning-def", "Krylos").Stealthy = true
	fixture.give(fixture.holder, "a1", "cunning-def", "Tartarus").Stealthy = true
	fixture.give(fixture.holder, "a3", "cunning-def", "Krylos")
	offers := goalEngine.Offers(fixture.holder, fixture.factionState)

	picked, err := goalEngine.Pick(fixture.holder, offers, domain.GoalChoice{GoalID: insideEnemyTerritory})
	if err != nil {
		t.Fatalf("Pick() unexpected error: %v", err)
	}
	if want := []string{"a1", "a2"}; !reflect.DeepEqual(picked.StealthedAtStart, want) {
		t.Errorf("Inside Enemy Territory StealthedAtStart = %v, want %v", picked.StealthedAtStart, want)
	}

	picked, err = goalEngine.Pick(fixture.holder, offers, domain.GoalChoice{GoalID: militaryConquest})
	if err != nil {
		t.Fatalf("Pick() unexpected error: %v", err)
	}
	if picked.StealthedAtStart != nil {
		t.Errorf("Military Conquest StealthedAtStart = %v, want nil", picked.StealthedAtStart)
	}
}
