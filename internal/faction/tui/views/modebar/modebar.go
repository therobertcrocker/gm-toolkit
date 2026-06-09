package modebar

import (
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type Mode int

const (
	ModeManage Mode = iota
	ModeTurn
	ModeSpatial
)

func (m Mode) String() string {
	switch m {
	case ModeManage:
		return "Manage"
	case ModeTurn:
		return "Turn"
	case ModeSpatial:
		return "Spatial"
	default:
		return "?"
	}
}

type slot struct {
	mode     Mode
	disabled bool
}

type Model struct {
	slots  []slot
	active int
}

func New() Model {
	return Model{
		slots: []slot{
			{mode: ModeManage, disabled: false},
			{mode: ModeTurn, disabled: false},
			{mode: ModeSpatial, disabled: true},
		},
		active: 0,
	}
}

func (m Model) Active() Mode { return m.slots[m.active].mode }

// Next moves the active highlight forward, skipping disabled slots. Wraps.
func (m Model) Next() Model {
	n := len(m.slots)
	for i := 1; i <= n; i++ {
		candidate := (m.active + i) % n
		if !m.slots[candidate].disabled {
			m.active = candidate
			return m
		}
	}
	return m
}

// Prev mirrors Next in reverse.
func (m Model) Prev() Model {
	n := len(m.slots)
	for i := 1; i <= n; i++ {
		candidate := (m.active - i + n) % n
		if !m.slots[candidate].disabled {
			m.active = candidate
			return m
		}
	}
	return m
}

// SetActive jumps to a specific Mode if it exists and is enabled; otherwise no-op.
func (m Model) SetActive(target Mode) Model {
	for i, s := range m.slots {
		if s.mode == target && !s.disabled {
			m.active = i
			return m
		}
	}
	return m
}

func (m Model) View() string {
	var parts []string
	for i, s := range m.slots {
		label := s.mode.String()
		switch {
		case s.disabled:
			parts = append(parts, styles.ModeBarDisabled.Render(label))
		case i == m.active:
			parts = append(parts, styles.ModeBarActive.Render(label))
		default:
			parts = append(parts, styles.ModeBarInactive.Render(label))
		}
	}
	return strings.Join(parts, styles.ModeBarSeparator.String())
}
