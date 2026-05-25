package adapter

import (
	"log/slog"
	"sync/atomic"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const eventChanBuffer = 64

type Adapter struct {
	engine       *engine.Engine
	factionState *state.FactionState
	paths        *campaigns.Paths

	askCh   chan CollectorAskMsg
	eventCh chan ObserverEventMsg
	doneCh  chan EngineDoneMsg

	perFactionCadence atomic.Bool

	log *slog.Logger
}

func New(eng *engine.Engine, factionState *state.FactionState, paths *campaigns.Paths, log *slog.Logger) *Adapter {
	return &Adapter{
		engine:       eng,
		factionState: factionState,
		paths:        paths,
		askCh:        make(chan CollectorAskMsg),
		eventCh:      make(chan ObserverEventMsg, eventChanBuffer),
		doneCh:       make(chan EngineDoneMsg, 1),
		log:          log,
	}
}

func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }

func (a *Adapter) Action() action.Collector { return &stubActionCollector{} }
func (a *Adapter) Collectors() engine.Collectors {
	return engine.Collectors{Phase: a.Phase(), Action: a.Action()}
}
