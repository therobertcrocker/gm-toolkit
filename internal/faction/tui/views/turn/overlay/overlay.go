package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

// Overlay is a modal prompt mounted over the execution view. Overlays never
// touch the adapter's ask/Reply plumbing (CollectorAskMsg, the Reply channel):
// they render a prompt and emit OverlayDoneMsg{Answer} via a tea.Cmd; the
// execution sub-model owns the channel send. Some Effort-3 overlays import plain
// adapter value types (RepairTarget, BribeReply); that edge is acyclic — adapter
// imports neither overlay nor execution.
type Overlay interface {
	Init() tea.Cmd
	Update(tea.Msg) (Overlay, tea.Cmd)
	View() string
	Help() help.KeyMap
}

// OverlayDoneMsg carries the overlay's answer back to the execution sub-model,
// which forwards it on the active ask's Reply channel. Cancel is deferred to
// Effort 3 (overview OQ 14); Effort 1 needs only the happy-path Answer.
type OverlayDoneMsg struct{ Answer any }
