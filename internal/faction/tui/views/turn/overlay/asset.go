package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
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
