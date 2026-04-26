package style

import "github.com/charmbracelet/lipgloss"

var (
	Header       = lipgloss.NewStyle().Bold(true)
	Muted        = lipgloss.NewStyle().Faint(true)
	HP           = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	LowHP        = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Coin         = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	PanelBorder  = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	SectionTitle = lipgloss.NewStyle().Bold(true).Underline(true)
)
