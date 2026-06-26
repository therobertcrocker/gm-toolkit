package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// relocateData.dests is index-aligned with the candidates slice; each entry is the
// chosen destination world ID for that asset ("" = leave in place).
type relocateData struct{ dests []string }

// RelocateOrders prompts a destination for each relocatable asset: one Select per
// candidate over the in-range worlds plus a "do not move" option. The composite
// doubles as the asset multi-select (an asset left on "do not move" is omitted).
// Answer: []action.RelocateOrder for rows with a chosen destination. Esc cancels.
type RelocateOrders struct {
	data       *relocateData
	candidates []*domain.Asset
	form       *huh.Form
}

func NewRelocateOrders(p adapter.SelectRelocateOrdersPayload, rb *rulebook.Rulebook) RelocateOrders {
	data := &relocateData{dests: make([]string, len(p.Candidates))}

	worldOpts := make([]huh.Option[string], 0, len(p.CandidateWorlds)+1)
	worldOpts = append(worldOpts, huh.NewOption("— do not move —", ""))
	for _, worldID := range p.CandidateWorlds {
		label := p.WorldNames[worldID]
		if label == "" {
			label = worldID
		}
		worldOpts = append(worldOpts, huh.NewOption(label, worldID))
	}

	groups := make([]*huh.Group, 0, len(p.Candidates))
	for i, asset := range p.Candidates {
		idx := i
		name := asset.DefinitionID
		if def, ok := rb.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s — relocate to?", name)).
				Options(worldOpts...).
				Value(&data.dests[idx]),
		))
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return RelocateOrders{data: data, candidates: p.Candidates, form: form}
}

func (o RelocateOrders) Init() tea.Cmd { return o.form.Init() }

func (o RelocateOrders) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		var orders []action.RelocateOrder
		for i, dest := range o.data.dests {
			if dest != "" {
				orders = append(orders, action.RelocateOrder{Asset: o.candidates[i], DestinationWorld: dest})
			}
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: orders} }
	}
	return o, cmd
}

func (o RelocateOrders) View() string      { return o.form.View() }
func (o RelocateOrders) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = RelocateOrders{}
