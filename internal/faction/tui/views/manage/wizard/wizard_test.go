package wizard

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestHomeworldGovernedElsewhere(t *testing.T) {
	factionState := &state.FactionState{Governments: map[string]string{"ruled": "f2"}}
	colonists := &domain.Tag{ID: "T-001", Name: "Colonists", HomeworldGovernment: true}
	warlike := &domain.Tag{ID: "T-020", Name: "Warlike"}

	cases := []struct {
		name        string
		tags        []*domain.Tag
		homeworldID string
		wantErr     bool
	}{
		{"Colonists on a governed homeworld", []*domain.Tag{colonists}, "ruled", true},
		{"Colonists on a free homeworld", []*domain.Tag{colonists}, "free", false},
		{"Colonists before a homeworld is picked", []*domain.Tag{colonists}, "", false},
		{"another tag on a governed homeworld", []*domain.Tag{warlike}, "ruled", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := homeworldGovernedElsewhere(testCase.tags, testCase.homeworldID, factionState)
			if (err != nil) != testCase.wantErr {
				t.Errorf("err = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestGovernmentTitle(t *testing.T) {
	proprietor := &rulebook.Rulebook{GovernmentTagID: "HAWK-T-001", Tags: map[string]*domain.Tag{"HAWK-T-001": {ID: "HAWK-T-001", Name: "Proprietor"}}}
	if got := governmentTitle(proprietor); got != "Governed world (Proprietor)" {
		t.Errorf("title = %q, want the rulebook's government tag named", got)
	}
	if got := governmentTitle(&rulebook.Rulebook{}); got != "Governed world" {
		t.Errorf("title with no government tag = %q, want %q", got, "Governed world")
	}
}
