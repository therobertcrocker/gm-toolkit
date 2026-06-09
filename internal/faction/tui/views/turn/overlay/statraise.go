package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// statRaiseData is heap-allocated and pointer-bound to the huh field so the
// value-copy of the overlay (bubbletea semantics) never orphans the binding.
type statRaiseData struct{ choice *domain.FactionStat }

// StatRaise is a single-select over the eligible stats plus an explicit Decline
// sentinel (mapping to a nil *FactionStat). Decline is in-modal, never Esc.
type StatRaise struct {
	data *statRaiseData
	form *huh.Form
}

func NewStatRaise(faction *domain.Faction, eligible []domain.FactionStat) StatRaise {
	data := &statRaiseData{}
	ratingOf := map[domain.FactionStat]int{
		domain.StatForce:   faction.Force,
		domain.StatCunning: faction.Cunning,
		domain.StatWealth:  faction.Wealth,
	}
	opts := make([]huh.Option[*domain.FactionStat], 0, len(eligible)+1)
	for i := range eligible {
		stat := &eligible[i]
		opts = append(opts, huh.NewOption(
			fmt.Sprintf("%s (currently %d)", eligible[i], ratingOf[eligible[i]]),
			stat,
		))
	}
	opts = append(opts, huh.NewOption("Decline raise", (*domain.FactionStat)(nil)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.FactionStat]().
				Title("Raise a stat").
				Options(opts...).
				Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return StatRaise{data: data, form: form}
}

func (o StatRaise) Init() tea.Cmd { return o.form.Init() }

func (o StatRaise) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil // cancel routing is Effort 3; swallow Esc so the form can't abort
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o StatRaise) View() string { return o.form.View() }

func (o StatRaise) Help() help.KeyMap { return formHelp{} }

var _ Overlay = StatRaise{}
