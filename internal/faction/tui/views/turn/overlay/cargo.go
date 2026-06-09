package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type cargoData struct{ chosen []*domain.Asset }

// TransportCargo is a capped multi-select fired once per transport that issued a
// move. The orchestrator re-validates the cap (orchestrator.go:672); in-modal
// Validate is for UX. Answer: the chosen assets (nil/empty = no cargo).
type TransportCargo struct {
	data *cargoData
	form *huh.Form
}

func NewTransportCargo(transport *domain.Asset, eligible []*domain.Asset, profile *domain.TransportProfile, rb *rulebook.Rulebook) TransportCargo {
	data := &cargoData{}
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(assetName(asset, rb), asset))
	}
	maxCargo := profile.MaxCargo

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[*domain.Asset]().
				Title(fmt.Sprintf("Load cargo onto %s (up to %d)", assetName(transport, rb), maxCargo)).
				Options(opts...).
				Value(&data.chosen).
				Validate(func(picked []*domain.Asset) error {
					if len(picked) > maxCargo {
						return fmt.Errorf("at most %d cargo", maxCargo)
					}
					return nil
				}),
		),
	).WithTheme(styles.FormTheme())
	return TransportCargo{data: data, form: form}
}

func (o TransportCargo) Init() tea.Cmd { return o.form.Init() }

func (o TransportCargo) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	// Esc declines with the legal empty answer (no cargo loaded). Cancel-as-error
	// stays scoped to action overlays (Decision 2).
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: []*domain.Asset{}} }
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.chosen
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o TransportCargo) View() string { return o.form.View() }

func (o TransportCargo) Help() help.KeyMap { return declineFormHelp{} }

var _ Overlay = TransportCargo{}
