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
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

// inputCapturer is implemented by sub-models that own the full keyboard while
// active (e.g. a text-entry form). When the active sub captures input, the root
// forwards every key to it and suppresses the global bindings below.
type inputCapturer interface {
	CapturesInput() bool
}

type globalKeys struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var globals = globalKeys{
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next mode")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev mode")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.resizeSubs()

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

		active := m.bar.Active()
		if sub, ok := m.subs[active]; ok {
			if capturer, isCapturer := sub.(inputCapturer); isCapturer && capturer.CapturesInput() {
				updated, cmd := sub.Update(msg)
				m.subs[active] = updated
				return m, cmd
			}
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
			m.showHelp = !m.showHelp
			return m.resizeSubs()
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

// Layout budget owned by the root. The header is the title rule/line/rule plus
// the mode bar and a blank line; the footer is the help line bracketed by two
// rules. Sub-models are handed the remaining height and never see the chrome.
const (
	headerHeight = 5
	footerHeight = 3
)

func (m Model) contentHeight() int {
	reserved := headerHeight
	if m.showHelp {
		reserved += footerHeight
	}
	if h := m.height - reserved; h > 1 {
		return h
	}
	return 1
}

// resizeSubs forwards the current content-area size to every sub-model and
// resizes the help renderer. Called on a terminal resize and whenever the help
// bar is toggled (which changes the content budget).
func (m Model) resizeSubs() (Model, tea.Cmd) {
	if w := m.width - 2; w > 1 {
		m.help.Width = w
	}
	sized := tea.WindowSizeMsg{Width: m.width, Height: m.contentHeight()}
	var cmds []tea.Cmd
	for mode, sub := range m.subs {
		updated, cmd := sub.Update(sized)
		m.subs[mode] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}
