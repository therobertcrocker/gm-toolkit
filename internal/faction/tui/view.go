package tui

import (
	"fmt"
	"strings"

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
		sb.WriteString("\n\n")
		sb.WriteString(styles.HelpStub.Render("Help: Tab / Shift-Tab cycle modes  ·  q quit  ·  ? toggle help"))
	}

	return sb.String()
}
