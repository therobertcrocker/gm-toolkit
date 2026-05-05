package engine

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// stubAction implements action.Action with a fixed name and no-op methods.
type stubAction struct{ name string }

func (s *stubAction) Name() string { return s.name }
func (s *stubAction) Validate(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
	return true
}
func (s *stubAction) Inputs(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	return nil
}
func (s *stubAction) Resolve(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	return nil
}
func (s *stubAction) Output() ([]domain.Mutation, error) { return nil, nil }

func TestFilterAllowedActions(t *testing.T) {
	available := []action.Action{
		&stubAction{"Attack"},
		&stubAction{"Sell Asset"},
		&stubAction{"Expand Influence"},
	}

	tests := []struct {
		name    string
		allowed []string
		want    []string
	}{
		{"nil allowed returns nothing", nil, nil},
		{"empty allowed returns nothing", []string{}, nil},
		{"no-match returns nothing", []string{"Buy Asset"}, nil},
		{"single match", []string{"Attack"}, []string{"Attack"}},
		{"multiple matches", []string{"Attack", "Expand Influence"}, []string{"Attack", "Expand Influence"}},
		{"preserves available order not allowed order", []string{"Expand Influence", "Sell Asset"}, []string{"Sell Asset", "Expand Influence"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterAllowedActions(available, tc.allowed)
			var gotNames []string
			for _, a := range got {
				gotNames = append(gotNames, a.Name())
			}
			if !reflect.DeepEqual(gotNames, tc.want) {
				t.Errorf("got %v, want %v", gotNames, tc.want)
			}
		})
	}
}
