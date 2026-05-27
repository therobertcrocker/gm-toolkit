package styles

import "github.com/charmbracelet/lipgloss"

var (
	ModeBarActive    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Background(lipgloss.Color("0")).Padding(0, 2)
	ModeBarInactive  = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 2)
	ModeBarDisabled  = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 2)
	ModeBarSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString(" │ ")

	Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Padding(2, 4).Italic(true)

	ConfirmExit = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Padding(1, 2)

	HelpStub = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(1, 2)

	SaveError = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)
