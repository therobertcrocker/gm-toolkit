package overlay

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type bribeData struct {
	base   *domain.Base
	amount string // parsed to int on submit
}

// Bribe is the numeric Bribe composite: a base Select over rival-owned bases
// (derived adapter-side), then a Coin-amount Input validated to 1..Coin. The
// answer is carried as adapter.BribeReply. Esc cancels.
type Bribe struct {
	data *bribeData
	form *huh.Form
}

func NewBribe(bases []*domain.Base, ownerNames map[string]string, coin int) Bribe {
	data := &bribeData{}

	opts := make([]huh.Option[*domain.Base], len(bases))
	for i, base := range bases {
		owner := ownerNames[base.OwnerID]
		if owner == "" {
			owner = base.OwnerID
		}
		opts[i] = huh.NewOption(
			fmt.Sprintf("%s base @ %s · infl %d", owner, base.Location.WorldID, base.Influence),
			base,
		)
	}

	baseGroup := huh.NewGroup(
		huh.NewSelect[*domain.Base]().
			Title("Bribe which base?").
			Options(opts...).
			Value(&data.base),
	)

	amountGroup := huh.NewGroup(
		huh.NewInput().
			Title(fmt.Sprintf("Coin to spend (1–%d)", coin)).
			Value(&data.amount).
			Validate(amountValidator(1, coin)),
	)

	form := huh.NewForm(baseGroup, amountGroup).WithTheme(styles.FormTheme())
	return Bribe{data: data, form: form}
}

func (o Bribe) Init() tea.Cmd { return o.form.Init() }

func (o Bribe) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		amount, _ := strconv.Atoi(o.data.amount) // Validate guarantees a valid int
		answer := adapter.BribeReply{Base: o.data.base, Amount: amount}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o Bribe) View() string      { return o.form.View() }
func (o Bribe) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = Bribe{}
