package overlay

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

var (
	keySpace = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
	keyUp    = tea.KeyMsg{Type: tea.KeyUp}
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
)

// pressRollCard sends each key to the overlay and returns it with the last command.
func pressRollCard(card Overlay, keys ...tea.KeyMsg) (Overlay, tea.Cmd) {
	var cmd tea.Cmd
	for _, pressed := range keys {
		card, cmd = card.Update(pressed)
	}
	return card, cmd
}

func rollCardAnswer(t *testing.T, cmd tea.Cmd) any {
	t.Helper()
	if cmd == nil {
		t.Fatal("no command, want OverlayDoneMsg")
	}
	done, ok := cmd().(OverlayDoneMsg)
	if !ok {
		t.Fatalf("command returned %T, want OverlayDoneMsg", cmd())
	}
	return done.Answer
}

func numberedOffers(count int) []string {
	offers := make([]string, 0, count)
	for index := range count {
		offers = append(offers, fmt.Sprintf("offer-%02d", index))
	}
	return offers
}

func TestPreRollCard_TogglesOneRowOfTwoWithTheSameLabel(t *testing.T) {
	card := NewPreRollCard(adapter.PreRollCardPayload{Offers: []string{"Add 1d10", "Add 1d10"}})

	_, cmd := pressRollCard(card, keyDown, keySpace, keyEnter)

	if answer := rollCardAnswer(t, cmd); !reflect.DeepEqual(answer, []int{1}) {
		t.Errorf("answer = %#v, want []int{1}", answer)
	}
}

func TestPreRollCard_SpaceTwiceTakesNothing(t *testing.T) {
	card := NewPreRollCard(adapter.PreRollCardPayload{Offers: []string{"Add 1d10"}})

	_, cmd := pressRollCard(card, keySpace, keySpace, keyEnter)

	if answer := rollCardAnswer(t, cmd); !reflect.DeepEqual(answer, []int{}) {
		t.Errorf("answer = %#v, want an empty []int", answer)
	}
}

func TestPreRollCard_NoOffersAnswersAnEmptyNonNilSlice(t *testing.T) {
	_, cmd := pressRollCard(NewPreRollCard(adapter.PreRollCardPayload{}), keyEnter)

	answer, ok := rollCardAnswer(t, cmd).([]int)
	if !ok || answer == nil || len(answer) != 0 {
		t.Errorf("answer = %#v, want an empty, non-nil []int", answer)
	}
}

func TestPreRollCard_StruckRowShowsAndCannotBeToggled(t *testing.T) {
	card := NewPreRollCard(adapter.PreRollCardPayload{
		Offers: []string{"Add 1d10"},
		Struck: []string{"Warlike: used this turn"},
	})

	updated, cmd := pressRollCard(card, keyDown, keySpace)
	if cmd != nil {
		t.Errorf("space on a struck row returned a command")
	}
	view := updated.View()
	if !strings.Contains(view, "> --- Warlike: used this turn") {
		t.Errorf("the cursor did not reach the struck row:\n%s", view)
	}

	_, cmd = pressRollCard(updated, keyEnter)
	if answer := rollCardAnswer(t, cmd); !reflect.DeepEqual(answer, []int{}) {
		t.Errorf("answer = %#v, want an empty []int", answer)
	}
}

func TestRollCards_EscDoesNothing(t *testing.T) {
	cards := map[string]Overlay{
		"before": NewPreRollCard(adapter.PreRollCardPayload{Offers: []string{"Add 1d10"}}),
		"after":  NewPostRollCard(adapter.PostRollCardPayload{Rerolls: []string{"Reroll the 7"}}),
	}
	for name, card := range cards {
		t.Run(name, func(t *testing.T) {
			if _, cmd := card.Update(keyEsc); cmd != nil {
				t.Errorf("Esc returned a command (%T), want none", cmd())
			}
		})
	}
}

func TestPostRollCard_OpensOnKeep(t *testing.T) {
	card := NewPostRollCard(adapter.PostRollCardPayload{Rerolls: []string{"Reroll the 7"}})

	_, cmd := pressRollCard(card, keyEnter)

	if answer := rollCardAnswer(t, cmd); answer != adapter.KeepRoll {
		t.Errorf("answer = %#v, want adapter.KeepRoll", answer)
	}
}

func TestPostRollCard_UpThenEnterTakesTheLastReroll(t *testing.T) {
	card := NewPostRollCard(adapter.PostRollCardPayload{Rerolls: []string{"Reroll the 7", "Reroll the 2"}})

	_, cmd := pressRollCard(card, keyUp, keyEnter)

	if answer := rollCardAnswer(t, cmd); answer != 1 {
		t.Errorf("answer = %#v, want 1", answer)
	}
}

func TestPreRollCard_ScrollsWithTheCursor(t *testing.T) {
	keys := make([]tea.KeyMsg, 11)
	for index := range keys {
		keys[index] = keyDown
	}

	card, _ := pressRollCard(NewPreRollCard(adapter.PreRollCardPayload{Offers: numberedOffers(12)}), keys...)

	view := card.View()
	if !strings.Contains(view, "offer-11") {
		t.Errorf("view does not hold the cursor's row:\n%s", view)
	}
	if strings.Contains(view, "offer-00") {
		t.Errorf("view still holds row 0:\n%s", view)
	}
}

func TestRollCards_ShowTheirKeys(t *testing.T) {
	cases := []struct {
		name string
		card Overlay
		want string
	}{
		{"before with offers", NewPreRollCard(adapter.PreRollCardPayload{Offers: []string{"Add 1d10"}}), "space toggles"},
		{"before with none", NewPreRollCard(adapter.PreRollCardPayload{}), "Press enter to roll"},
		{"after", NewPostRollCard(adapter.PostRollCardPayload{Rerolls: []string{"Reroll the 7"}}), "enter chooses"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if view := tc.card.View(); !strings.Contains(view, tc.want) {
				t.Errorf("view does not hold %q:\n%s", tc.want, view)
			}
		})
	}
}

func TestPreRollCard_FitsTheWizardBand(t *testing.T) {
	view := NewPreRollCard(adapter.PreRollCardPayload{Offers: numberedOffers(12)}).View()

	if height := strings.Count(view, "\n") + 1; height > 14 {
		t.Errorf("view is %d lines, want at most 14", height)
	}
}
