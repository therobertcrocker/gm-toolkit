package overlay

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// buyOrderData is heap-allocated so huh's pointer bindings survive the
// value-copy of the overlay. world is bound by the first group; def by the
// second, whose options rebuild from world via OptionsFunc.
type buyOrderData struct {
	world string
	def   *domain.AssetDefinition
}

// BuyOrder is the two-step Buy composite: a world Select, then a definition
// Select whose options derive from the chosen world via OptionsFunc (re-evaluated
// when navigation enters the second group; eval.go:33 bindings hash). Definition
// labels carry the sticker cost vs Coin as a budgeting cue — the engine
// re-resolves the true cost via dispatch.ResolveAssetCost at Resolve
// (buy_asset.go:82), so the label is informational, not authoritative.
// Answer: action.BuyOrder. Esc cancels.
type BuyOrder struct {
	data *buyOrderData
	form *huh.Form
}

func NewBuyOrder(purchasableByWorld map[string][]*domain.AssetDefinition, worldNames map[string]string, coin int) BuyOrder {
	data := &buyOrderData{}

	worlds := make([]string, 0, len(purchasableByWorld))
	for world := range purchasableByWorld {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)

	worldOpts := make([]huh.Option[string], len(worlds))
	for i, world := range worlds {
		label := worldNames[world]
		if label == "" {
			label = world
		}
		worldOpts[i] = huh.NewOption(label, world)
	}

	worldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Buy on which world?").
			Options(worldOpts...).
			Value(&data.world),
	)

	defGroup := huh.NewGroup(
		huh.NewSelect[*domain.AssetDefinition]().
			Title("Purchase which asset?").
			Value(&data.def).
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				defs := purchasableByWorld[data.world]
				opts := make([]huh.Option[*domain.AssetDefinition], len(defs))
				for i, def := range defs {
					opts[i] = huh.NewOption(
						fmt.Sprintf("%s · TL %d · cost %d/%d Coin", def.Name, def.TechLevel, def.Cost, coin),
						def,
					)
				}
				return opts
			}, &data.world),
	)

	form := huh.NewForm(worldGroup, defGroup).WithTheme(styles.FormTheme())
	return BuyOrder{data: data, form: form}
}

func (o BuyOrder) Init() tea.Cmd { return o.form.Init() }

func (o BuyOrder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := action.BuyOrder{World: o.data.world, Definition: o.data.def}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o BuyOrder) View() string      { return o.form.View() }
func (o BuyOrder) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = BuyOrder{}
