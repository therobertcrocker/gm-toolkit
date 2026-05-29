package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Text weight / brightness
	Strong = lipgloss.NewStyle().Bold(true).Foreground(Text) // names, values, items
	Body   = lipgloss.NewStyle().Foreground(Text)            // default body
	Subtle = lipgloss.NewStyle().Foreground(Label)           // descriptions, scale, inactive
	Dim    = lipgloss.NewStyle().Foreground(Muted)           // ids, "none", bullets, hints

	// Accents
	Accent    = lipgloss.NewStyle().Foreground(Mauve) // section headers, active mode
	AccentAlt = lipgloss.NewStyle().Foreground(Sky)   // primary-stat highlight

	// State
	Success = lipgloss.NewStyle().Foreground(Teal)
	Warning = lipgloss.NewStyle().Foreground(Yellow)
	Danger  = lipgloss.NewStyle().Foreground(Red)

	// Structure
	Rule = lipgloss.NewStyle().Foreground(Border) // dividers, separators, section underlines
)
