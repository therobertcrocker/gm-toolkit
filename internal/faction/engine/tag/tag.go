package tag

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Handler interface {
	TagID() string
	Apply(faction *domain.Faction, hookRegistry *hooks.Registry)
}

type TagEngine struct {
	handlers map[string]Handler
	log      *slog.Logger
}

func New(log *slog.Logger) *TagEngine {
	e := &TagEngine{handlers: make(map[string]Handler), log: log}
	e.Register(tags.ScavengersHandler{})
	e.Register(tags.WarlikeHandler{})
	e.Register(tags.FanaticalHandler{})
	e.Register(tags.PreceptorArchiveHandler{})
	return e
}

func (e *TagEngine) Register(handler Handler) {
	e.handlers[handler.TagID()] = handler
}

// ApplyAll walks every faction-tag in factionState and invokes the registered
// handler for that tag ID. Tags with no registered handler are silently
// skipped — they are data-only entries from the rulebook.
func (e *TagEngine) ApplyAll(factionState *state.FactionState, hookRegistry *hooks.Registry, log *slog.Logger) {

	for _, faction := range factionState.Factions {
		for _, tag := range faction.Tags {
			handler, ok := e.handlers[tag.ID]
			if !ok {
				// Data-only tag: present in tags TOML, no Go handler registered. Intentional.
				continue
			}
			handler.Apply(faction, hookRegistry)
		}
	}
}
