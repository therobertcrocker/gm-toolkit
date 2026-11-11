package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// goalPickData is heap-allocated and pointer-bound to the huh fields so the
// value-copy of the overlay (bubbletea semantics) never orphans the bindings.
type goalPickData struct {
	offer    *domain.GoalOffer
	targetID string
}

// GoalPick is the turn-start goal pick: a select over the goals on offer, whose
// description follows the cursor, then a target select shown only for a goal
// that takes a target faction. Decline is an option (a nil offer), never Esc (D-4).
type GoalPick struct {
	data *goalPickData
	form *huh.Form
}

func NewGoalPick(offers []domain.GoalOffer) GoalPick {
	data := &goalPickData{}
	goalOptions := make([]huh.Option[*domain.GoalOffer], 0, len(offers)+1)
	// deliberate: decline first, not last, because huh scrolls the list down to the option matching the nil starting value.
	goalOptions = append(goalOptions, huh.NewOption("Decline goal", (*domain.GoalOffer)(nil)))
	for i := range offers {
		goalOptions = append(goalOptions, huh.NewOption(offers[i].Goal.Name, &offers[i]))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.GoalOffer]().
				Title("Pick a goal").
				DescriptionFunc(func() string { return goalPickDescription(data.offer) }, &data.offer).
				Options(goalOptions...).
				Value(&data.offer),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Target faction").
				OptionsFunc(func() []huh.Option[string] { return targetOptions(data.offer) }, &data.offer).
				Value(&data.targetID).
				Validate(func(targetID string) error {
					if targetID == "" {
						return fmt.Errorf("pick a target faction")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return data.offer == nil || len(data.offer.Targets) == 0 }),
	).WithTheme(styles.FormTheme())
	return GoalPick{data: data, form: form}
}

func goalPickDescription(offer *domain.GoalOffer) string {
	if offer == nil {
		return "No goal this turn; the pick comes back next turn."
	}
	return offer.Goal.Description
}

func targetOptions(offer *domain.GoalOffer) []huh.Option[string] {
	if offer == nil {
		return nil
	}
	options := make([]huh.Option[string], len(offer.Targets))
	for i, target := range offer.Targets {
		options[i] = huh.NewOption(target.Name, target.ID)
	}
	return options
}

func (o GoalPick) Init() tea.Cmd { return o.form.Init() }

func (o GoalPick) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		return o, nil // decline is an option, so Esc must not abort the form
	}
	updatedForm, cmd := o.form.Update(msg)
	o.form = updatedForm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.answer()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

// answer is nil for a decline. A target is sent only for a goal that takes one,
// so a target chosen for an earlier highlighted goal is never carried over.
func (o GoalPick) answer() *domain.GoalChoice {
	offer := o.data.offer
	if offer == nil {
		return nil
	}
	choice := &domain.GoalChoice{GoalID: offer.Goal.ID}
	if len(offer.Targets) > 0 {
		choice.TargetFactionID = o.data.targetID
	}
	return choice
}

func (o GoalPick) View() string { return o.form.View() }

func (o GoalPick) Help() help.KeyMap { return formHelp{} }

var _ Overlay = GoalPick{}
