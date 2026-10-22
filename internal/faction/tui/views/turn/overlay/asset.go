package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAsset builds the shared single-asset picker (Sell, and Buy's stealth
// follow-on). Options are labeled name + HP + location. Answer: *domain.Asset.
func NewSelectAsset(title string, assets []*domain.Asset, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(assets))
	for _, a := range assets {
		opts = append(opts, huh.NewOption(assetLabel(a, rb), a))
	}
	return NewSelect[*domain.Asset](title, "", opts)
}

// SelectAssetToLose is Monopoly's loss pick. Esc does nothing: the book gives the
// target no way out of the loss (Effort 3 Decision Record 5).
type SelectAssetToLose struct {
	SelectOverlay[*domain.Asset]
}

func NewSelectAssetToLose(title string, assets []*domain.Asset, rb *rulebook.Rulebook) SelectAssetToLose {
	return SelectAssetToLose{NewSelectAsset(title, assets, rb)}
}

func (o SelectAssetToLose) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil
	}
	updated, cmd := o.SelectOverlay.Update(msg)
	o.SelectOverlay = updated.(SelectOverlay[*domain.Asset])
	return o, cmd
}

func (o SelectAssetToLose) Help() help.KeyMap { return formHelp{} }

// assetLabel renders "<name> · HP x/y" for an owned asset. assetName lives in
// movement.go (Effort 2). HP max comes from the definition.
func assetLabel(a *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(a, rb)
	if def, ok := rb.Assets[a.DefinitionID]; ok {
		return fmt.Sprintf("%s · HP %d/%d", name, a.CurrentHP, def.HP)
	}
	return name
}

var _ help.KeyMap = cancelFormHelp{}
