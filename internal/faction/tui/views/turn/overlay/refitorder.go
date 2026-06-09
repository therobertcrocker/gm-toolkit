package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type refitOrderData struct {
	assetID string
	newDef  *domain.AssetDefinition
}

// RefitOrder is the two-step Refit composite: an asset Select, then a
// replacement-definition Select whose options derive from the chosen asset via
// OptionsFunc. Replacement labels show the cost delta the engine charges
// (max(0, newDef.Cost-oldDef.Cost), refit_asset.go:65). Answer: action.RefitOrder.
// Esc cancels.
type RefitOrder struct {
	data      *refitOrderData
	assetByID map[string]*domain.Asset
	form      *huh.Form
}

func NewRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) RefitOrder {
	data := &refitOrderData{}
	assetByID := make(map[string]*domain.Asset, len(options))
	replByAsset := make(map[string][]*domain.AssetDefinition, len(options))

	assetOpts := make([]huh.Option[string], 0, len(options))
	for _, opt := range options {
		assetByID[opt.Asset.ID] = opt.Asset
		replByAsset[opt.Asset.ID] = opt.Replacements
		assetOpts = append(assetOpts, huh.NewOption(assetLabel(opt.Asset, rb), opt.Asset.ID))
	}

	assetGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Refit which asset?").
			Options(assetOpts...).
			Value(&data.assetID),
	)

	defGroup := huh.NewGroup(
		huh.NewSelect[*domain.AssetDefinition]().
			Title("Replace with?").
			Value(&data.newDef).
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				oldCost := 0
				if asset := assetByID[data.assetID]; asset != nil {
					if oldDef, ok := rb.Assets[asset.DefinitionID]; ok {
						oldCost = oldDef.Cost
					}
				}
				repls := replByAsset[data.assetID]
				opts := make([]huh.Option[*domain.AssetDefinition], len(repls))
				for i, def := range repls {
					opts[i] = huh.NewOption(
						fmt.Sprintf("%s · cost +%d Coin", def.Name, max(0, def.Cost-oldCost)),
						def,
					)
				}
				return opts
			}, &data.assetID),
	)

	form := huh.NewForm(assetGroup, defGroup).WithTheme(styles.FormTheme())
	return RefitOrder{data: data, assetByID: assetByID, form: form}
}

func (o RefitOrder) Init() tea.Cmd { return o.form.Init() }

func (o RefitOrder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := action.RefitOrder{OldAsset: o.assetByID[o.data.assetID], NewDefinition: o.data.newDef}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o RefitOrder) View() string      { return o.form.View() }
func (o RefitOrder) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = RefitOrder{}
