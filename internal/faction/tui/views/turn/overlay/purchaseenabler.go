package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// PurchaseEnabler prompts the GM to pick one asset from the enabler-eligible candidates.
// Labels show the surcharged cost (half again, rounded up) so the premium is visible
// at selection. Answer: *domain.AssetDefinition. Esc cancels.
type PurchaseEnabler struct {
	data *purchaseEnablerData
	form *huh.Form
}

type purchaseEnablerData struct {
	def *domain.AssetDefinition
}

func NewPurchaseEnabler(p adapter.SelectPurchaseEnablerAssetPayload) PurchaseEnabler {
	data := &purchaseEnablerData{}

	opts := make([]huh.Option[*domain.AssetDefinition], len(p.Candidates))
	for i, def := range p.Candidates {
		surcharged := (def.Cost*150 + 99) / 100
		opts[i] = huh.NewOption(
			fmt.Sprintf("%s · TL %d · cost %d Coin (surcharged)", def.Name, def.TechLevel, surcharged),
			def,
		)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.AssetDefinition]().
				Title("Purchase which asset? (Pretech Logistics)").
				Options(opts...).
				Value(&data.def),
		),
	).WithTheme(styles.FormTheme())

	return PurchaseEnabler{data: data, form: form}
}

func (o PurchaseEnabler) Init() tea.Cmd { return o.form.Init() }

func (o PurchaseEnabler) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: o.data.def} }
	}
	return o, cmd
}

func (o PurchaseEnabler) View() string      { return o.form.View() }
func (o PurchaseEnabler) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = PurchaseEnabler{}
