package overlay

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
)

type namedAction struct {
	action.Action
	name string
}

func (a *namedAction) Name() string { return a.name }

// firstFrame is what the overlay draws after its first message; huh builds a
// group's view only on Update.
func firstFrame(o Overlay) string {
	updated, _ := o.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return updated.View()
}

// A list whose starting value is its "do nothing" option must list that option
// first, so the list opens with every option in view, not scrolled down to it.
func TestListsOpenWithEveryOptionInView(t *testing.T) {
	faction := &domain.Faction{Force: 3, Cunning: 2, Wealth: 1}
	cases := []struct {
		name      string
		view      string
		wantFirst string
		wantLast  string
	}{
		{
			"select action",
			firstFrame(NewSelectAction([]action.Action{&namedAction{name: "Attack"}, &namedAction{name: "Expand Influence"}})),
			"Skip faction", "Expand Influence",
		},
		{
			"stat raise",
			firstFrame(NewStatRaise(faction, []domain.FactionStat{domain.StatForce, domain.StatWealth})),
			"Decline raise", "Wealth (currently 1)",
		},
		{
			"goal pick",
			firstFrame(NewGoalPick(goalPickOffers())),
			"Decline goal", "Destroy the Foe",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.view, tc.wantFirst) {
				t.Errorf("first frame lacks the first option %q:\n%s", tc.wantFirst, tc.view)
			}
			if !strings.Contains(tc.view, tc.wantLast) {
				t.Errorf("first frame lacks the last option %q:\n%s", tc.wantLast, tc.view)
			}
		})
	}
}
