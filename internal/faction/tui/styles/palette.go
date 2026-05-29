package styles

import "github.com/charmbracelet/lipgloss"

// Catppuccin Frappé palette. The ONLY place hex appears.
var (
	Bg      = lipgloss.Color("#303446") // Base
	Surface = lipgloss.Color("#414559") // Surface 0
	Border  = lipgloss.Color("#626880") // Surface 2
	Muted   = lipgloss.Color("#9196a8") // Overlay 1
	Label   = lipgloss.Color("#a1aacc") // Subtext 0
	Text    = lipgloss.Color("#e1e7fa") // Text

	Mauve  = lipgloss.Color("#ca9ee6") // Mauve  — primary accent
	Sky    = lipgloss.Color("#8caaee") // Sky    — secondary accent
	Teal   = lipgloss.Color("#94e2d5") // Teal  — success / HP healthy
	Yellow = lipgloss.Color("#e5c890") // Yellow — warning / HP mid
	Red    = lipgloss.Color("#e78284") // Red    — danger / HP critical
)
