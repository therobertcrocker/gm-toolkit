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

func NewRelocateOrders(payload adapter.SelectRelocateOrdersPayload, rb *rulebook.Rulebook) RelocateOrders {
	data := &relocateData{dests: make([]string, len(payload.Candidates))}

	worldOpts := make([]huh.Option[string], 0, len(payload.CandidateWorlds)+1)
	worldOpts = append(worldOpts, huh.NewOption("— do not move —", ""))
	for _, worldID := range payload.CandidateWorlds {
		label := payload.WorldNames[worldID]
		if label == "" {
			label = worldID
		}
		worldOpts = append(worldOpts, huh.NewOption(label, worldID))
	}

	groups := make([]*huh.Group, 0, len(payload.Candidates))
	for i, asset := range payload.Candidates {
		idx := i
		name := asset.DefinitionID
		if def, ok := rb.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		sel := huh.NewSelect[string]().
			Title(fmt.Sprintf("%s — relocate to?", name)).
			Options(worldOpts...).
			Value(&data.dests[idx])
		if payload.MaxAssets > 0 {
			sel = sel.Validate(capValidator(data, idx, payload.MaxAssets))
		}
		groups = append(groups, huh.NewGroup(sel))
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return RelocateOrders{data: data, candidates: payload.Candidates, form: form}
}

// capValidator rejects a destination pick when it would push the count of assets
// being relocated past maxAssets. dests[idx] is already bound to the incoming value
// by the time huh runs Validate, so the passed val is counted via the slice.
func capValidator(data *relocateData, idx, maxAssets int) func(string) error {
	return func(val string) error {
		if val == "" {
			return nil
		}
		count := 0
		for i, dest := range data.dests {
			if dest != "" || i == idx {
				count++
			}
		}
		if count > maxAssets {
			return fmt.Errorf("at most %d asset(s) may be relocated by this ability", maxAssets)
		}
		return nil
	}
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
