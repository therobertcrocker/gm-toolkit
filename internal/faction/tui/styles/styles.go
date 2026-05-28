package styles

import "github.com/charmbracelet/lipgloss"

var (
	AppTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	AppCampaign = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	AppDivider  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	ModeBarActive    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89B4FA")).Background(lipgloss.Color("#1E1E2E")).Padding(0, 2)
	ModeBarInactive  = lipgloss.NewStyle().Foreground(lipgloss.Color("#BAC2DE")).Padding(0, 2)
	ModeBarDisabled  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Padding(0, 2)
	ModeBarSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).SetString(" │ ")

	Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")).Padding(2, 4).Italic(true)

	ConfirmExit = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8")).Padding(1, 2)

	HelpStub = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Padding(1, 2)

	SaveError = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Bold(true)
)
