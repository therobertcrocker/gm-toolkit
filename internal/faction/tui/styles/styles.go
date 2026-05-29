package styles

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	// App chrome
	AppTitle    = AccentAlt.Bold(true)
	AppCampaign = Dim
	AppDivider  = Rule

	// Mode bar
	ModeBarActive    = AccentAlt.Bold(true).Background(Surface).Padding(0, 2)
	ModeBarInactive  = Subtle.Padding(0, 2)
	ModeBarDisabled  = Dim.Padding(0, 2)
	ModeBarSeparator = Rule.SetString(" │ ")

	// Section heading (underline rule composed at call-site from Rule)
	SectionHeading = Accent.Bold(true)

	// Prompts
	DangerPrompt = Danger.Bold(true).Padding(1, 2)  // quit, delete-confirm
	WarnPrompt   = Warning.Bold(true).Padding(1, 2) // discard-new-faction

	// Misc
	Placeholder = Warning.Padding(2, 4).Italic(true)
	HelpStub    = Dim.Padding(1, 2)
	SaveError   = Danger.Bold(true)
)

// HealthColor maps an HP ratio (0..1) to a palette color for the HP bar.
func HealthColor(ratio float64) lipgloss.Color {
	switch {
	case ratio > 0.5:
		return Teal
	case ratio > 0.25:
		return Yellow
	default:
		return Red
	}
}

// FormTheme adapts huh's Charm theme to the Bubblegum palette.
// Replaces wizard.wizardTheme.
func FormTheme() *huh.Theme {
	theme := huh.ThemeCharm()
	theme.Focused.SelectedPrefix = theme.Focused.SelectedPrefix.SetString("[✓] ")
	theme.Focused.UnselectedPrefix = theme.Focused.UnselectedPrefix.SetString("[ ] ")
	theme.Blurred.SelectedPrefix = theme.Blurred.SelectedPrefix.SetString("[✓] ")
	theme.Blurred.UnselectedPrefix = theme.Blurred.UnselectedPrefix.SetString("[ ] ")
	return theme
}
