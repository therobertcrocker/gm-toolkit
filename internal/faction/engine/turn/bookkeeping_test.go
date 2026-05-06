package turn

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestApplyBookkeeping_ClearsHookBudgets(t *testing.T) {
	te := newTurn()
	s := &state.FactionState{
		CampaignID: "test",
		Factions: map[string]*domain.Faction{
			"f1": {
				ID:      "f1",
				Force:   4,
				Cunning: 3,
				Wealth:  6,
				Coin:    10,
				HookBudgets: map[string]int{
					"tag:Warlike":   1,
					"tag:Fanatical": 2,
				},
			},
		},
	}
	if err := te.Start(s); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if _, _, err := te.ApplyBookkeeping(s, nil); err != nil {
		t.Fatalf("ApplyBookkeeping: %v", err)
	}

	faction := s.Factions["f1"]
	if faction.HookBudgets != nil {
		t.Errorf("HookBudgets = %v after ApplyBookkeeping, want nil", faction.HookBudgets)
	}
}
