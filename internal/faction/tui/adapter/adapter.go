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
		log:          log,
	}
}

func (a *Adapter) SetPerFactionCadence(on bool) { a.perFactionCadence.Store(on) }

// emit sends a notification without blocking: a full buffer drops it and logs.
func (a *Adapter) emit(kind EventKind, payload any) {
	select {
	case a.eventCh <- ObserverEventMsg{Kind: kind, Payload: payload}:
	default:
		a.log.Warn("observer event dropped — eventCh full", "kind", kind)
	}
}

func (a *Adapter) Action() action.Collector { return newActionCollector(a) }
func (a *Adapter) Collectors() engine.Collectors {
	return engine.Collectors{Phase: a.Phase(), Action: a.Action()}
}
