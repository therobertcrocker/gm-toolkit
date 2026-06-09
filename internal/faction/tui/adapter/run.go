package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (a *Adapter) Run() tea.Cmd {
	return func() tea.Msg {
		defer close(a.eventCh)
		defer close(a.askCh)
		if !a.engine.Turn.InProgress(a.factionState) {
			if err := a.engine.Turn.Start(a.factionState); err != nil {
				return EngineDoneMsg{Err: err}
			}
		}
		a.emitCycleStarted()
		err := a.engine.RunCycle(a.factionState, a.paths, a.Collectors(), a.Observer())
		return EngineDoneMsg{Err: err}
	}
}

// emitCycleStarted sends the synthetic EvtCycleStarted as the cycle's first
// event. It runs on the Run goroutine before RunCycle, single-threaded against
// CurrentTurn (non-nil here: either InProgress was already true or Start just
// set it), so reading the order races nothing. The order is copied with display
// names so the execution view never holds engine pointers.
func (a *Adapter) emitCycleStarted() {
	current := a.factionState.CurrentTurn
	order := make([]RailEntry, 0, len(current.FactionOrder))
	for _, id := range current.FactionOrder {
		name := id
		if faction, ok := a.factionState.Factions[id]; ok {
			name = faction.Name
		}
		order = append(order, RailEntry{ID: id, Name: name})
	}
	a.eventCh <- ObserverEventMsg{
		Kind: EvtCycleStarted,
		Payload: CycleStartedPayload{
			CycleNumber: current.CycleNumber,
			Order:       order,
		},
	}
}

// ObserverPump reads one event from eventCh and returns it as a tea.Msg. The
// router re-issues this command after each receipt to keep the pump alive; on a
// drained, closed eventCh it returns StreamClosedMsg, the router's teardown
// signal.
func (a *Adapter) ObserverPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.eventCh
		if !ok {
			return StreamClosedMsg{}
		}
		return msg
	}
}

// AskPump reads one CollectorAskMsg from askCh and returns it as a tea.Msg,
// mirroring ObserverPump. The router re-issues it after each receipt and stops
// on EngineDoneMsg. Returns nil when Run closes askCh.
func (a *Adapter) AskPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.askCh
		if !ok {
			return nil
		}
		return msg
	}
}
