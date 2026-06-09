package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
)

func (m Model) View() string {
	var sb strings.Builder

	ruleWidth := m.width
	if ruleWidth <= 0 {
		ruleWidth = 80
	}
	rule := styles.AppDivider.Render(strings.Repeat("─", ruleWidth))

	var infoStr string
	if m.factionState != nil {
		infoStr = fmt.Sprintf("  %s  ·  Cycle %d→%d  ·  %d Factions  ",
			m.factionState.CampaignID,
			m.factionState.CycleNumber,
			m.factionState.CycleNumber+1,
			len(m.factionState.Factions))
	}
	left := styles.AppTitle.Render("  F A C T I O N   M A N A G E R")
	right := styles.AppCampaign.Render(infoStr)
	gap := ruleWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	titleLine := left + strings.Repeat(" ", gap) + right

	sb.WriteString(rule + "\n")
	sb.WriteString(titleLine + "\n")
	sb.WriteString(rule + "\n")
	sb.WriteString(lipgloss.PlaceHorizontal(ruleWidth, lipgloss.Center, m.bar.View()))
	sb.WriteString("\n\n")

	var content string
	if m.confirmExit {
		content = styles.DangerPrompt.Render("Quit? Press Enter to confirm, Esc to cancel.")
	} else {
		switch m.bar.Active() {
		case modebar.ModeSpatial:
			content = styles.Placeholder.Render("Spatial — reserved for F-012 (Spatial Map CLI)")
		default:
			if sub, ok := m.subs[m.bar.Active()]; ok {
				content = sub.View()
			} else {
				content = styles.Placeholder.Render(fmt.Sprintf("Mode %v has no registered sub-model.", m.bar.Active()))
			}
		}
	}
	sb.WriteString(lipgloss.NewStyle().Height(m.contentHeight()).Render(content))

	sb.WriteString(m.footerView())

	return sb.String()
}

// footerView renders the always-present bottom bar: a status line from the
// active sub-model when it has one, otherwise the help bar. View and
// contentHeight both call this so the reserved footer height always matches
// what's drawn.
func (m Model) footerView() string {
	ruleWidth := m.width
	if ruleWidth <= 0 {
		ruleWidth = 80
	}
	rule := styles.AppDivider.Render(strings.Repeat("─", ruleWidth))

	var inner string
	if text, severity, ok := activeStatusLine(m); ok {
		inner = statusStyle(severity).Render(text)
	} else {
		var subBindings help.KeyMap = emptyKeyMap{}
		if helper, ok := m.subs[m.bar.Active()].(Helper); ok {
			if km := helper.Help(); km != nil {
				subBindings = km
			}
		}
		inner = m.help.View(combineKeyMaps(globalsKeyMap{}, subBindings))
	}
	return "\n" + rule + "\n  " + inner + "\n" + rule
}

// activeStatusLine returns the active sub-model's status line when it implements
// chrome.StatusLiner and the text is non-empty.
func activeStatusLine(m Model) (text string, severity chrome.Severity, ok bool) {
	statusLiner, isLiner := m.subs[m.bar.Active()].(chrome.StatusLiner)
	if !isLiner {
		return "", 0, false
	}
	text, severity = statusLiner.StatusLine()
	return text, severity, text != ""
}

func statusStyle(severity chrome.Severity) lipgloss.Style {
	switch severity {
	case chrome.Fatal:
		return styles.Danger
	case chrome.Recoverable:
		return styles.Warning
	default:
		return styles.Body
	}
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
