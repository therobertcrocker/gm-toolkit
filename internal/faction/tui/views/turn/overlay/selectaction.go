package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type selectActionData struct{ choice action.Action }

// SelectAction is the action-phase picker. It lists every available (engine
// Validate-passed) action; actions absent from ImplementedActions render with a
// "(not yet available)" suffix and are rejected by Validate (huh has no native
// disabled option). A "Skip faction" sentinel maps to a nil action.Action, the
// legal skip path (orchestrator.go:418). Esc also cancels -> skip the action.
type SelectAction struct {
	data *selectActionData
	form *huh.Form
}

func NewSelectAction(available []action.Action) SelectAction {
	data := &selectActionData{}
	opts := make([]huh.Option[action.Action], 0, len(available)+1)
	for _, a := range available {
		label := a.Name()
		if !ImplementedActions[a.Name()] {
			label += " (not yet available)"
		}
		opts = append(opts, huh.NewOption(label, a))
	}
	opts = append(opts, huh.NewOption("Skip faction (take no action)", action.Action(nil)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[action.Action]().
				Title("Select action").
				Options(opts...).
				Value(&data.choice).
				Validate(func(a action.Action) error {
					if a != nil && !ImplementedActions[a.Name()] {
						return fmt.Errorf("%s is not yet available", a.Name())
					}
					return nil
				}),
		),
	).WithTheme(styles.FormTheme())
	return SelectAction{data: data, form: form}
}

func (o SelectAction) Init() tea.Cmd { return o.form.Init() }

func (o SelectAction) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice // action.Action or nil
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o SelectAction) View() string     { return o.form.View() }
func (o SelectAction) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = SelectAction{}
