// Update Discipline (per tui-rebuild-arc-discovery.md "Update Discipline").
//
// MAY:
//   - Forward messages to the active sub-model (subs[mode]).
//   - Handle global key bindings (Tab, Shift-Tab, q, ?, Esc) at the root.
//   - Update mode-bar state and switch the active mode.
//   - Update root-overlay state (help overlay, confirm-exit overlay).
//   - Return tea.Cmds for I/O performed by the adapter (engine launch, observer pump).
//
// MAY NOT:
//   - Call engine methods directly (engine.RunCycle, engine.RunFactionTurn, etc.).
//   - Mutate *state.FactionState or any *domain.* value.
//   - Perform I/O outside a tea.Cmd (filesystem reads, network calls, logging beyond
//     the structured logger's non-blocking calls).
//   - Embed game logic — turn ordering, action selection, mutation application all
//     belong to the engine. Update is a projection of engine state and a forwarder
//     of user intent.

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmExit {
			switch msg.String() {
			case "enter":
				return m, tea.Quit
			case "esc":
				m.confirmExit = false
				m.bar = m.bar.SetActive(m.priorMode)
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "tab":
			m.bar = m.bar.Next()
			return m.afterModeChange()
		case "shift+tab":
			m.bar = m.bar.Prev()
			return m.afterModeChange()
		case "q":
			m.priorMode = m.bar.Active()
			m.bar = m.bar.SetActive(modebar.ModeQuit)
			m.confirmExit = true
			return m, nil
		case "?":
			m.showHelpStub = !m.showHelpStub
			return m, nil
		}
	}

	active := m.bar.Active()
	if sub, ok := m.subs[active]; ok {
		updated, cmd := sub.Update(msg)
		m.subs[active] = updated
		return m, cmd
	}
	return m, nil
}

func (m Model) afterModeChange() (Model, tea.Cmd) {
	return m, nil
}
