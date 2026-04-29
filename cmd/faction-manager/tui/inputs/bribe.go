package inputs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type BribeOrderSelectedMsg struct {
	Base   *domain.Base
	Amount int
}

type bribeStep int

const (
	bribeStepBase   bribeStep = iota
	bribeStepAmount
)

type BribeModel struct {
	faction      *domain.Faction
	step         bribeStep
	baseCursor   int
	selectedBase *domain.Base
	coinAmount   int
	maxCoin      int
}

func NewBribeModel(faction *domain.Faction) BribeModel {
	return BribeModel{
		faction:    faction,
		coinAmount: 1,
		maxCoin:    faction.Coin,
	}
}

func (m BribeModel) Init() tea.Cmd { return nil }

func (m BribeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.step {
	case bribeStepBase:
		return m.updateBase(key)
	case bribeStepAmount:
		return m.updateAmount(key)
	}
	return m, nil
}

func (m BribeModel) updateBase(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.baseCursor > 0 {
			m.baseCursor--
		}
	case "down", "j":
		if m.baseCursor < len(m.faction.Bases)-1 {
			m.baseCursor++
		}
	case "enter":
		if len(m.faction.Bases) == 0 {
			break
		}
		m.selectedBase = m.faction.Bases[m.baseCursor]
		m.step = bribeStepAmount
	}
	return m, nil
}

func (m BribeModel) updateAmount(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "+", "=", "up", "k":
		if m.coinAmount < m.maxCoin {
			m.coinAmount++
		}
	case "-", "down", "j":
		if m.coinAmount > 1 {
			m.coinAmount--
		}
	case "enter":
		return m, func() tea.Msg { return BribeOrderSelectedMsg{Base: m.selectedBase, Amount: m.coinAmount} }
	}
	return m, nil
}

func (m BribeModel) View() string {
	switch m.step {
	case bribeStepBase:
		return m.viewBase()
	case bribeStepAmount:
		return m.viewAmount()
	}
	return ""
}

func (m BribeModel) viewBase() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render("Bribe — Select Base"))
	sb.WriteString("\n\n")
	for i, base := range m.faction.Bases {
		cursor := "  "
		if i == m.baseCursor {
			cursor = "> "
		}
		fmt.Fprintf(&sb, "%s%-24s  Influence: %d\n", cursor, base.Location, base.Influence)
	}
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("↑/↓ navigate  Enter select"))
	return sb.String()
}

func (m BribeModel) viewAmount() string {
	var sb strings.Builder
	sb.WriteString(style.SectionTitle.Render(fmt.Sprintf("Bribe — %s", m.selectedBase.Location)))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "Amount:  %s  (max %d)\n\n",
		style.Coin.Render(fmt.Sprintf("%d Coin", m.coinAmount)), m.maxCoin)
	sb.WriteString(style.Muted.Render("+/- adjust  Enter confirm"))
	return sb.String()
}
