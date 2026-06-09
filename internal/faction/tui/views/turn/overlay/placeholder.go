package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Placeholder is Effort 1's stand-in for every real overlay. It names the
// prompt (kind + faction, when the ask carries one) and, on any keypress,
// emits OverlayDoneMsg carrying the kind's legal "no decision" default.
// Efforts 2-3 replace it kind-by-kind via newOverlay's factory switch.
type Placeholder struct {
	kind    adapter.AskKind
	faction string
	ack     key.Binding
}

func NewPlaceholder(kind adapter.AskKind, faction string) Placeholder {
	return Placeholder{
		kind:    kind,
		faction: faction,
		ack:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("any key", "continue")),
	}
}

func (p Placeholder) Init() tea.Cmd { return nil }

// Update acks on any key. The keypress requirement is deliberate: it paces the
// cycle so each prompt — including the final cycle_summary checkpoint — is a
// "done reading" beat before the engine proceeds.
func (p Placeholder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		answer := defaultFor(p.kind)
		return p, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return p, nil
}

func (p Placeholder) View() string {
	label := kindLabel(p.kind)
	if p.faction != "" {
		label += " · " + p.faction
	}
	return styles.WarnPrompt.Render(label + " · press any key")
}

func (p Placeholder) Help() help.KeyMap { return placeholderHelp{p.ack} }

type placeholderHelp struct{ ack key.Binding }

func (h placeholderHelp) ShortHelp() []key.Binding  { return []key.Binding{h.ack} }
func (h placeholderHelp) FullHelp() [][]key.Binding { return [][]key.Binding{{h.ack}} }

// defaultFor returns the legal "no decision" reply for each ask kind. Every
// value must be the exact type the phase collector asserts on
// (phase_collector.go:35-81), or the assertion errors. The available defaults:
//   - AskAwaitCheckpoint: a non-error ack (discarded by AwaitCheckpoint)
//   - AskSelectAction: skip this faction (untyped nil → orchestrator.go:418)
//   - AskSelectStatRaise: decline the raise (a typed nil *domain.FactionStat)
//   - AskSelectMovementDecisions: no movement (empty []world.MovementDecision)
//   - AskSelectTransportCargo: no cargo (empty []*domain.Asset)
func defaultFor(kind adapter.AskKind) any {
	switch kind {
	case adapter.AskAwaitCheckpoint:
		return true
	case adapter.AskSelectAction:
		return nil
	case adapter.AskSelectStatRaise:
		return (*domain.FactionStat)(nil)
	case adapter.AskSelectMovementDecisions:
		return []world.MovementDecision{}
	case adapter.AskSelectTransportCargo:
		return []*domain.Asset{}
	default:
		return nil
	}
}

func kindLabel(kind adapter.AskKind) string {
	switch kind {
	case adapter.AskAwaitCheckpoint:
		return "Checkpoint"
	case adapter.AskSelectAction:
		return "Select Action"
	case adapter.AskSelectStatRaise:
		return "Select Stat Raise"
	case adapter.AskSelectMovementDecisions:
		return "Select Movement"
	case adapter.AskSelectTransportCargo:
		return "Select Transport Cargo"
	default:
		return "Prompt"
	}
}

var _ Overlay = Placeholder{}
