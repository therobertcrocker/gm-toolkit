package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (a *Adapter) Run() tea.Cmd {
	return func() tea.Msg {
		err := a.engine.RunCycle(a.factionState, a.paths, a.Collectors(), a.Observer())
		return EngineDoneMsg{Err: err}
	}
}

// ObserverPump reads one event from eventCh and returns it as a tea.Msg.
// The root Update re-issues this command after each receipt to keep the pump alive.
func (a *Adapter) ObserverPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.eventCh
		if !ok {
			return nil
		}
		return msg
	}
}

// Stop closes eventCh so any pending ObserverPump unblocks and returns nil,
// signaling Update to stop re-issuing the pump command.
func (a *Adapter) Stop() {
	close(a.eventCh)
}
