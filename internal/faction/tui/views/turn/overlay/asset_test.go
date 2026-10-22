package overlay

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func TestSelectAssetToLoseIgnoresEsc(t *testing.T) {
	assets := []*domain.Asset{{ID: "b-1", DefinitionID: "machine"}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"machine": {ID: "machine", Name: "Machine"}}}

	updated, cmd := NewSelectAssetToLose("lose which asset?", assets, rb).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("Esc produced a command (%T), want none", cmd())
	}
	if _, ok := updated.(SelectAssetToLose); !ok {
		t.Errorf("Update returned %T, want SelectAssetToLose", updated)
	}
}
