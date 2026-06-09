package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Checkpoint is the end-of-cycle wrap. In whole-cycle mode the only checkpoint
// that reaches the TUI is cycle_summary (phase_collector.go:28), so this always
// renders the cycle-complete ack. Answer is discarded by AwaitCheckpoint.
type Checkpoint struct {
	cycle int
	ack   key.Binding
}

func NewCheckpoint(cycle int) Checkpoint {
	return Checkpoint{
		cycle: cycle,
		ack:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "return to setup")),
	}
}

func (c Checkpoint) Init() tea.Cmd { return nil }

func (c Checkpoint) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, c.ack) {
		return c, func() tea.Msg { return OverlayDoneMsg{Answer: true} }
	}
	return c, nil
}

func (c Checkpoint) View() string {
	return styles.WarnPrompt.Render(fmt.Sprintf("Cycle %d complete — press Enter to return to setup", c.cycle))
}

func (c Checkpoint) Help() help.KeyMap { return checkpointHelp{c.ack} }

type checkpointHelp struct{ ack key.Binding }

func (h checkpointHelp) ShortHelp() []key.Binding  { return []key.Binding{h.ack} }
func (h checkpointHelp) FullHelp() [][]key.Binding { return [][]key.Binding{{h.ack}} }

var _ Overlay = Checkpoint{}
