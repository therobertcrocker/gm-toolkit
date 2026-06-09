package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// repairData.steps is index-aligned with the targets slice; each entry is the
// chosen heal-step count for that asset.
type repairData struct{ steps []int }

// RepairOrders is the numeric Repair composite: one heal-step Select per damaged
// asset over 0..MaxSteps (the adapter-derived useful ceiling). Step n heals
// min(n*HealHP, Missing) and costs n(n+1)/2 Coin (escalating; the engine
// re-validates the total against Coin — repair_asset.go:92-97). Answer:
// []action.RepairOrder for rows with a non-zero step count. Esc cancels.
type RepairOrders struct {
	data    *repairData
	targets []adapter.RepairTarget
	form    *huh.Form
}

func NewRepairOrders(targets []adapter.RepairTarget, rb *rulebook.Rulebook) RepairOrders {
	data := &repairData{steps: make([]int, len(targets))}

	groups := make([]*huh.Group, 0, len(targets))
	for i, target := range targets {
		idx := i
		name := target.Asset.DefinitionID
		if def, ok := rb.Assets[target.Asset.DefinitionID]; ok {
			name = def.Name
		}

		opts := make([]huh.Option[int], 0, target.MaxSteps+1)
		opts = append(opts, huh.NewOption("no repair", 0))
		for step := 1; step <= target.MaxSteps; step++ {
			healed := min(step*target.HealHP, target.Missing)
			cost := step * (step + 1) / 2
			opts = append(opts, huh.NewOption(fmt.Sprintf("heal %d HP (%d step(s), ≈%d Coin)", healed, step, cost), step))
		}

		groups = append(groups, huh.NewGroup(
			huh.NewSelect[int]().
				Title(fmt.Sprintf("%s · HP %d/%d — repair?", name, target.Asset.CurrentHP, target.Asset.CurrentHP+target.Missing)).
				Options(opts...).
				Value(&data.steps[idx]),
		))
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return RepairOrders{data: data, targets: targets, form: form}
}

func (o RepairOrders) Init() tea.Cmd { return o.form.Init() }

func (o RepairOrders) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		var orders []action.RepairOrder
		for i, step := range o.data.steps {
			if step > 0 {
				orders = append(orders, action.RepairOrder{Asset: o.targets[i].Asset, HealCount: step})
			}
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: orders} }
	}
	return o, cmd
}

func (o RepairOrders) View() string      { return o.form.View() }
func (o RepairOrders) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = RepairOrders{}
