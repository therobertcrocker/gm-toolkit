package contextstrip

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type Model struct {
	factionState *state.FactionState
}

func New(factionState *state.FactionState) Model { return Model{factionState: factionState} }
func (m Model) View() string {
	return styles.Placeholder.Render("Strip — Commit 5")
}
