package overlay

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
)

// formHelp is the keymap for overlays where Esc does nothing: checkpoint and
// stat-raise have no empty answer (decline is an explicit form option), so Esc
// is intentionally absent (Decision 2).
type formHelp struct{}

func (formHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}

func (formHelp) FullHelp() [][]key.Binding { return [][]key.Binding{formHelp{}.ShortHelp()} }

// declineFormHelp is the keymap for phase overlays with a legal empty answer
// (movement, cargo): form nav plus Esc-skip.
type declineFormHelp struct{}

func (declineFormHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "skip")),
	}
}
func (declineFormHelp) FullHelp() [][]key.Binding { return [][]key.Binding{declineFormHelp{}.ShortHelp()} }

// cancelFormHelp is the keymap for action overlays: form nav plus Esc-cancel.
type cancelFormHelp struct{}

func (cancelFormHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel action")),
	}
}
func (cancelFormHelp) FullHelp() [][]key.Binding { return [][]key.Binding{cancelFormHelp{}.ShortHelp()} }

func capError(n int) error { return fmt.Errorf("at most %d", n) }

// amountValidator parses a positive-integer Coin amount in [lo, hi]. Shared by
// every overlay with a numeric Coin input (Expand Influence, Bribe).
func amountValidator(lo, hi int) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("enter a number")
		}
		if n < lo || n > hi {
			return fmt.Errorf("%d–%d", lo, hi)
		}
		return nil
	}
}
