package overlay

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const (
	militaryConquest = "CORE-G-001"
	destroyTheFoe    = "CORE-G-008"
)

func goalPickOffers() []domain.GoalOffer {
	return []domain.GoalOffer{
		{Goal: &domain.Goal{ID: militaryConquest, Name: "Military Conquest"}},
		{
			Goal:    &domain.Goal{ID: destroyTheFoe, Name: "Destroy the Foe"},
			Targets: []*domain.Faction{{ID: "f2", Name: "Beta"}, {ID: "f3", Name: "Gamma"}},
		},
	}
}

func TestGoalPickIgnoresEsc(t *testing.T) {
	updated, cmd := NewGoalPick(goalPickOffers()).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(GoalPick); !ok {
		t.Errorf("Update returned %T, want GoalPick", updated)
	}
}

func TestGoalPickAnswer(t *testing.T) {
	offers := goalPickOffers()
	cases := []struct {
		name     string
		offer    *domain.GoalOffer
		targetID string
		want     *domain.GoalChoice
	}{
		{"decline", nil, "", nil},
		{"goal without a target", &offers[0], "", &domain.GoalChoice{GoalID: militaryConquest}},
		{"goal with a target", &offers[1], "f2", &domain.GoalChoice{GoalID: destroyTheFoe, TargetFactionID: "f2"}},
		{"leftover target is dropped", &offers[0], "f2", &domain.GoalChoice{GoalID: militaryConquest}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pick := NewGoalPick(offers)
			pick.data.offer = tc.offer
			pick.data.targetID = tc.targetID
			got := pick.answer()
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("answer() = %+v, want nil", *got)
			case tc.want != nil && got == nil:
				t.Errorf("answer() = nil, want %+v", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Errorf("answer() = %+v, want %+v", *got, *tc.want)
			}
		})
	}
}

func TestGoalPickTargetOptions(t *testing.T) {
	offers := goalPickOffers()
	options := targetOptions(&offers[1])
	if len(options) != 2 || options[0].Key != "Beta" || options[0].Value != "f2" || options[1].Key != "Gamma" || options[1].Value != "f3" {
		t.Errorf("targetOptions(Destroy the Foe) = %+v, want Beta/f2 and Gamma/f3", options)
	}
	if options := targetOptions(&offers[0]); len(options) != 0 {
		t.Errorf("targetOptions(Military Conquest) = %+v, want none", options)
	}
	if options := targetOptions(nil); len(options) != 0 {
		t.Errorf("targetOptions(nil) = %+v, want none", options)
	}
}
