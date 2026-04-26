package phases

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/tui/style"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

type BookkeepingDoneMsg struct{}

type BookkeepingModel struct {
	result   engine.BookkeepingResult
	rulebook *loader.Rulebook
}

func NewBookkeepingModel(result engine.BookkeepingResult, rulebook *loader.Rulebook) BookkeepingModel {
	return BookkeepingModel{result: result, rulebook: rulebook}
}

func (m BookkeepingModel) Init() tea.Cmd { return nil }

func (m BookkeepingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, func() tea.Msg { return BookkeepingDoneMsg{} }
	}
	return m, nil
}

func (m BookkeepingModel) View() string {
	var sb strings.Builder

	sb.WriteString(style.SectionTitle.Render("Bookkeeping"))
	sb.WriteString("\n\n")

	fmt.Fprintf(&sb, "Income: %s  %s\n",
		style.HP.Render(fmt.Sprintf("+%d Coin", m.result.IncomeGained)),
		style.Muted.Render(fmt.Sprintf("(Wealth %d, Stats %d)", m.result.WealthIncome, m.result.StatIncome)),
	)

	for _, ref := range m.result.AssetsLost {
		fmt.Fprintf(&sb, "\n%s  %s on %s — lost (unpaid 2 turns)",
			style.LowHP.Render("✗"),
			assetDisplayName(ref.DefinitionID, m.rulebook),
			ref.Location,
		)
	}

	for _, ref := range m.result.AssetsUnmaintained {
		fmt.Fprintf(&sb, "\n%s  %s on %s — first missed payment",
			style.LowHP.Render("!"),
			assetDisplayName(ref.DefinitionID, m.rulebook),
			ref.Location,
		)
	}

	sb.WriteString("\n\n")
	sb.WriteString(style.Muted.Render("Press any key to continue"))

	return sb.String()
}

func assetDisplayName(definitionID string, rulebook *loader.Rulebook) string {
	if def, ok := rulebook.Assets[definitionID]; ok {
		return def.Name
	}
	return definitionID
}
