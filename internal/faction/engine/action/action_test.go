package action

import (
	"errors"
	"log/slog"
	"maps"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	spentKey     = "effect:reroll:DEF"
	earlierSpend = "effect:bonus_die:DEF"
)

// spendingAction spends a budget on the acting faction and on a rival while it
// resolves, then fails where the test scripts it to.
type spendingAction struct {
	rival      *domain.Faction
	resolveErr error
	outputErr  error
}

func (*spendingAction) Name() string { return "Spend" }

func (*spendingAction) Validate(*domain.Faction, *state.FactionState, *rulebook.Rulebook) bool {
	return true
}

func (*spendingAction) Inputs(*domain.Faction, *state.FactionState, *rulebook.Rulebook) error {
	return nil
}

func (spending *spendingAction) Resolve(faction *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	faction.SpendBudget(spentKey)
	spending.rival.SpendBudget(spentKey)
	return spending.resolveErr
}

func (spending *spendingAction) Output() ([]domain.Mutation, error) {
	return nil, spending.outputErr
}

func TestRun_RestoresBudgetsOnAnyError(t *testing.T) {
	fatal, outputFailure := errors.New("broken"), errors.New("no output")
	cases := []struct {
		name       string
		resolveErr error
		outputErr  error
		wantErr    error
	}{
		{"a recoverable error", ErrTurnCanceled, nil, ErrTurnCanceled},
		{"a fatal error", fatal, nil, fatal},
		{"a failure in Output", nil, outputFailure, outputFailure},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actor := &domain.Faction{ID: "alpha", HookBudgets: map[string]int{earlierSpend: 1}}
			rival := &domain.Faction{ID: "beta"}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": actor, "beta": rival}}
			spending := &spendingAction{rival: rival, resolveErr: testCase.resolveErr, outputErr: testCase.outputErr}

			_, err := New(slog.New(slog.DiscardHandler)).Run(spending, actor, factionState, nil, slog.New(slog.DiscardHandler))

			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("err = %v, want %v", err, testCase.wantErr)
			}
			if !maps.Equal(actor.HookBudgets, map[string]int{earlierSpend: 1}) {
				t.Errorf("actor budgets = %v, want only the earlier spend", actor.HookBudgets)
			}
			if rival.HookBudgets != nil {
				t.Errorf("rival budgets = %v, want the nil map back", rival.HookBudgets)
			}
		})
	}
}

func TestRun_SuccessKeepsTheSpends(t *testing.T) {
	actor := &domain.Faction{ID: "alpha"}
	rival := &domain.Faction{ID: "beta"}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": actor, "beta": rival}}

	_, err := New(slog.New(slog.DiscardHandler)).Run(&spendingAction{rival: rival}, actor, factionState, nil, slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if actor.HookBudgets[spentKey] != 1 || rival.HookBudgets[spentKey] != 1 {
		t.Errorf("budgets = %v and %v, want the spend kept on both factions", actor.HookBudgets, rival.HookBudgets)
	}
}
