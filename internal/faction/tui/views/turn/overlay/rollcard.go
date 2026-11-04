package overlay

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// rollCardRows is how many rows a card lists at once. Three fact rows, a blank, these
// rows, a blank and the prompt fill the 14-row wizard band.
const rollCardRows = 8

// visibleRows returns the rows to draw, so the cursor stays on screen.
func visibleRows(cursor, count int) (first, last int) {
	if cursor >= rollCardRows {
		first = cursor - rollCardRows + 1
	}
	return first, min(first+rollCardRows, count)
}

func moveCursor(pressed string, cursor, count int) int {
	switch pressed {
	case "up", "k":
		return max(cursor-1, 0)
	case "down", "j":
		return min(cursor+1, max(count-1, 0))
	}
	return cursor
}

func cursorMark(selected bool) string {
	if selected {
		return "> "
	}
	return "  "
}

func rollFacts(headline, detail, dice string) []string {
	return []string{styles.Strong.Render(headline), styles.Subtle.Render(detail), styles.AccentAlt.Render(dice)}
}

// PreRollCard is the before card. It only draws what the adapter worded, and its
// answer is the indices of the offers taken.
type PreRollCard struct {
	card   adapter.PreRollCardPayload
	cursor int
	chosen map[int]bool
}

func NewPreRollCard(card adapter.PreRollCardPayload) PreRollCard {
	return PreRollCard{card: card, chosen: make(map[int]bool)}
}

func (o PreRollCard) Init() tea.Cmd { return nil }

func (o PreRollCard) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyMsg)
	if !ok {
		return o, nil
	}
	// deliberate: no esc case, not today's "none", because a stray key must not throw away a bonus die (D-7).
	switch keyPress.String() {
	case " ":
		if o.cursor < len(o.card.Offers) {
			o.chosen[o.cursor] = !o.chosen[o.cursor]
		}
	case "enter":
		answer := make([]int, 0, len(o.chosen))
		for index := range o.card.Offers {
			if o.chosen[index] {
				answer = append(answer, index)
			}
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	default:
		o.cursor = moveCursor(keyPress.String(), o.cursor, len(o.card.Offers)+len(o.card.Struck))
	}
	return o, nil
}

func (o PreRollCard) View() string {
	lines := rollFacts(o.card.Headline, o.card.Detail, o.card.Dice)
	lines = append(lines, "")
	count := len(o.card.Offers) + len(o.card.Struck)
	first, last := visibleRows(o.cursor, count)
	for row := first; row < last; row++ {
		mark := cursorMark(row == o.cursor)
		if row >= len(o.card.Offers) {
			lines = append(lines, styles.Dim.Render(mark+"--- "+o.card.Struck[row-len(o.card.Offers)]))
			continue
		}
		box := "[ ] "
		if o.chosen[row] {
			box = "[x] "
		}
		lines = append(lines, mark+box+o.card.Offers[row])
	}
	if count > 0 {
		lines = append(lines, "")
	}
	// deliberate: the keys are on the card, not only in Help, because nothing draws an overlay's Help and a hand-drawn card has no huh hint line.
	prompt := "Press enter to roll"
	if len(o.card.Offers) > 0 {
		prompt = "↑/↓ move, space toggles, enter rolls"
	}
	lines = append(lines, styles.Subtle.Render(prompt))
	return strings.Join(lines, "\n")
}

func (o PreRollCard) Help() help.KeyMap { return preRollHelp{} }

// PostRollCard is the after card. Its answer is an index into the rerolls, or
// adapter.KeepRoll. The cursor opens on Keep, so a stray enter spends nothing, and Esc
// does nothing (D-7).
type PostRollCard struct {
	card   adapter.PostRollCardPayload
	cursor int
}

func NewPostRollCard(card adapter.PostRollCardPayload) PostRollCard {
	return PostRollCard{card: card, cursor: len(card.Rerolls)}
}

func (o PostRollCard) Init() tea.Cmd { return nil }

func (o PostRollCard) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyMsg)
	if !ok {
		return o, nil
	}
	if keyPress.String() == "enter" {
		answer := adapter.KeepRoll
		if o.cursor < len(o.card.Rerolls) {
			answer = o.cursor
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	o.cursor = moveCursor(keyPress.String(), o.cursor, len(o.card.Rerolls)+1)
	return o, nil
}

func (o PostRollCard) View() string {
	lines := rollFacts(o.card.Headline, o.card.Detail, o.card.Dice)
	for _, note := range o.card.Applied {
		lines = append(lines, styles.Subtle.Render(note))
	}
	lines = append(lines, "")
	first, last := visibleRows(o.cursor, len(o.card.Rerolls)+1)
	for row := first; row < last; row++ {
		text := "Keep"
		if row < len(o.card.Rerolls) {
			text = o.card.Rerolls[row]
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursorMark(row == o.cursor), text))
	}
	lines = append(lines, "", styles.Subtle.Render("↑/↓ move, enter chooses"))
	return strings.Join(lines, "\n")
}

func (o PostRollCard) Help() help.KeyMap { return postRollHelp{} }

type preRollHelp struct{}

func (preRollHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "roll")),
	}
}
func (preRollHelp) FullHelp() [][]key.Binding { return [][]key.Binding{preRollHelp{}.ShortHelp()} }

type postRollHelp struct{}

func (postRollHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
	}
}
func (postRollHelp) FullHelp() [][]key.Binding { return [][]key.Binding{postRollHelp{}.ShortHelp()} }

var (
	_ Overlay = PreRollCard{}
	_ Overlay = PostRollCard{}
)
