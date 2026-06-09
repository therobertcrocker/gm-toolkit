package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// cancelOnEsc is the shared Esc handler for every action overlay. Returning a
// non-nil cmd means "canceled": the overlay emits OverlayDoneMsg carrying
// action.ErrTurnCanceled, which the execution model forwards on the reply
// channel; the adapter's ask helper turns it into a returned error the
// orchestrator classifies Recoverable (skip the faction's action, continue).
func cancelOnEsc(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return func() tea.Msg { return OverlayDoneMsg{Answer: action.ErrTurnCanceled} }
	}
	return nil
}

// withHeader renders an optional context block above a form. Empty -> nothing.
func withHeader(header, formView string) string {
	if header == "" {
		return formView
	}
	return lipgloss.JoinVertical(lipgloss.Left, styles.Subtle.Render(header), "", formView)
}

// --- SelectOverlay[T]: single-select with optional context header ---

type selectData[T any] struct{ choice T }

// SelectOverlay is a generic single-select. The skip/decline sentinel (if any)
// is just an option whose value is the zero T (e.g. nil action.Action). The
// answer is the chosen T; Esc cancels.
type SelectOverlay[T comparable] struct {
	data   *selectData[T]
	form   *huh.Form
	header string
}

func NewSelect[T comparable](title, header string, opts []huh.Option[T]) SelectOverlay[T] {
	data := &selectData[T]{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[T]().Title(title).Options(opts...).Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return SelectOverlay[T]{data: data, form: form, header: header}
}

func (o SelectOverlay[T]) Init() tea.Cmd { return o.form.Init() }

func (o SelectOverlay[T]) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o SelectOverlay[T]) View() string     { return withHeader(o.header, o.form.View()) }
func (o SelectOverlay[T]) Help() help.KeyMap { return cancelFormHelp{} }

// --- MultiSelectOverlay[T]: multi-select with optional cap ---

type multiData[T any] struct{ chosen []T }

type MultiSelectOverlay[T comparable] struct {
	data   *multiData[T]
	form   *huh.Form
	header string
}

// NewMultiSelect builds a multi-select. cap <= 0 means uncapped.
func NewMultiSelect[T comparable](title, header string, opts []huh.Option[T], cap int) MultiSelectOverlay[T] {
	data := &multiData[T]{}
	sel := huh.NewMultiSelect[T]().Title(title).Options(opts...).Value(&data.chosen)
	if cap > 0 {
		sel = sel.Validate(func(picked []T) error {
			if len(picked) > cap {
				return capError(cap)
			}
			return nil
		})
	}
	form := huh.NewForm(huh.NewGroup(sel)).WithTheme(styles.FormTheme())
	return MultiSelectOverlay[T]{data: data, form: form, header: header}
}

func (o MultiSelectOverlay[T]) Init() tea.Cmd { return o.form.Init() }

func (o MultiSelectOverlay[T]) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.chosen
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o MultiSelectOverlay[T]) View() string     { return withHeader(o.header, o.form.View()) }
func (o MultiSelectOverlay[T]) Help() help.KeyMap { return cancelFormHelp{} }

// --- ConfirmOverlay: yes/no ---

type confirmData struct{ choice bool }

type ConfirmOverlay struct {
	data   *confirmData
	form   *huh.Form
	header string
}

func NewConfirm(title, header, affirmative, negative string) ConfirmOverlay {
	data := &confirmData{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(title).Affirmative(affirmative).Negative(negative).Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return ConfirmOverlay{data: data, form: form, header: header}
}

func (o ConfirmOverlay) Init() tea.Cmd { return o.form.Init() }

func (o ConfirmOverlay) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o ConfirmOverlay) View() string     { return withHeader(o.header, o.form.View()) }
func (o ConfirmOverlay) Help() help.KeyMap { return cancelFormHelp{} }

var (
	_ Overlay = SelectOverlay[int]{}
	_ Overlay = MultiSelectOverlay[int]{}
	_ Overlay = ConfirmOverlay{}
)
