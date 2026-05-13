package tag

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Handler interface {
	TagID() string
	Apply(faction *domain.Faction, hookRegistry *hooks.Registry)
}

type TagEngine struct {
	handlers map[string]Handler
}

func New() *TagEngine {
	return &TagEngine{handlers: make(map[string]Handler)}
}

func (e *TagEngine) Register(handler Handler) {
	e.handlers[handler.TagID()] = handler
}

// ApplyAll walks every faction-tag in factionState and invokes the registered
// handler for that tag ID. Tags with no registered handler are silently
// skipped — they are data-only entries from the rulebook.
func (e *TagEngine) ApplyAll(factionState *state.FactionState, hookRegistry *hooks.Registry) {
	for _, faction := range factionState.Factions {
		for _, tag := range faction.Tags {
			handler, ok := e.handlers[tag.ID]
			if !ok {
				continue
			}
			handler.Apply(faction, hookRegistry)
		}
	}
}
