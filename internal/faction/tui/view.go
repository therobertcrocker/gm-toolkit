package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(m.bar.View())
	sb.WriteString("\n\n")

	if m.confirmExit {
		sb.WriteString(styles.ConfirmExit.Render("Quit? Press Enter to confirm, Esc to cancel."))
	} else {
		switch m.bar.Active() {
		case modebar.ModeSpatial:
			sb.WriteString(styles.Placeholder.Render("Spatial — reserved for F-012 (Spatial Map CLI)"))
		case modebar.ModeQuit:
			sb.WriteString(styles.ConfirmExit.Render("Quit slot active. Press Enter to confirm."))
		default:
			if sub, ok := m.subs[m.bar.Active()]; ok {
				sb.WriteString(sub.View())
			} else {
				sb.WriteString(styles.Placeholder.Render(fmt.Sprintf("Mode %v has no registered sub-model.", m.bar.Active())))
			}
		}
	}

	if m.showHelpStub {
		var subBindings help.KeyMap = emptyKeyMap{}
		if helper, ok := m.subs[m.bar.Active()].(Helper); ok {
			if km := helper.Help(); km != nil {
				subBindings = km
			}
		}
		composed := combineKeyMaps(globalsKeyMap{}, subBindings)
		sb.WriteString("\n\n")
		sb.WriteString(m.help.View(composed))
	}

	return sb.String()
}

type emptyKeyMap struct{}

func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }

type globalsKeyMap struct{}

func (globalsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{globals.Tab, globals.ShiftTab, globals.Help, globals.Quit}
}
func (globalsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{globals.Tab, globals.ShiftTab}, {globals.Help, globals.Quit}}
}

type combinedKeyMap struct {
	a help.KeyMap
	b help.KeyMap
}

func (c combinedKeyMap) ShortHelp() []key.Binding {
	return append(c.a.ShortHelp(), c.b.ShortHelp()...)
}
func (c combinedKeyMap) FullHelp() [][]key.Binding {
	return append(c.a.FullHelp(), c.b.FullHelp()...)
}

func combineKeyMaps(a, b help.KeyMap) help.KeyMap {
	return combinedKeyMap{a: a, b: b}
}
