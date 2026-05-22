package effect

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Handler interface {
	AssetDefinitionID() string
	Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry)
}

type EffectsEngine struct {
	handlers map[string]Handler
}

func New() *EffectsEngine {
	return &EffectsEngine{handlers: make(map[string]Handler)}
}

func (e *EffectsEngine) Register(handler Handler) {
	e.handlers[handler.AssetDefinitionID()] = handler
}

// ApplyAll walks every asset across all factions, looks up the handler for
// that asset's definition ID, and invokes it. Assets whose definitions are
// missing or whose def has no registered handler are silently skipped — the
// registration set is data-driven by the rulebook (e.g. def.Transport != nil).
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry) {
	for _, faction := range factionState.Factions {
		for _, asset := range faction.Assets {
			def := rulebook.Assets[asset.DefinitionID]
			if def == nil {
				continue
			}
			handler, ok := e.handlers[def.ID]
			if !ok {
				continue
			}
			handler.Apply(faction, asset, hookRegistry)
		}
	}
}
