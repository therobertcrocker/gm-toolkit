package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// --- helpers ---

func newTestState(factionIDs ...string) *state.FactionState {
	s := &state.FactionState{CampaignID: "test"}
	for _, id := range factionIDs {
		s.Factions = append(s.Factions, &domain.Faction{
			ID:     id,
			Force:  4,
			Cunning: 3,
			Wealth: 6,
			Coin:   0,
		})
	}
	return s
}

func newTurn() *TurnEngine {
	return newTurnEngine(newMutationEngine())
}

// --- InProgress ---

func TestInProgress_NoTurn(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	if te.InProgress(s) {
		t.Error("InProgress() = true, want false when no turn active")
	}
}

func TestInProgress_ActiveTurn(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	_ = te.Start(s)
	if !te.InProgress(s) {
		t.Error("InProgress() = false, want true after Start()")
	}
}

// --- Start ---

func TestStart_InitializesTurnState(t *testing.T) {
	te := newTurn()
	s := newTestState("a", "b", "c")

	if err := te.Start(s); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if s.CycleNumber != 1 {
		t.Errorf("CycleNumber: got %d, want 1", s.CycleNumber)
	}
	if s.CurrentTurn == nil {
		t.Fatal("CurrentTurn: got nil, want non-nil")
	}
	if !s.CurrentTurn.InProgress {
		t.Error("CurrentTurn.InProgress: got false, want true")
	}
	if s.CurrentTurn.CycleNumber != 1 {
		t.Errorf("CurrentTurn.CycleNumber: got %d, want 1", s.CurrentTurn.CycleNumber)
	}
	if s.CurrentTurn.CurrentIndex != 0 {
		t.Errorf("CurrentTurn.CurrentIndex: got %d, want 0", s.CurrentTurn.CurrentIndex)
	}
	if s.CurrentTurn.Phase != domain.PhaseBookkeeping {
		t.Errorf("CurrentTurn.Phase: got %v, want PhaseBookkeeping", s.CurrentTurn.Phase)
	}
	if len(s.CurrentTurn.FactionOrder) != 3 {
		t.Errorf("FactionOrder length: got %d, want 3", len(s.CurrentTurn.FactionOrder))
	}
}

func TestStart_ErrorIfAlreadyInProgress(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	_ = te.Start(s)
	if err := te.Start(s); err != ErrTurnInProgress {
		t.Errorf("Start() error = %v, want ErrTurnInProgress", err)
	}
}

func TestStart_ErrorIfNoFactions(t *testing.T) {
	te := newTurn()
	s := newTestState()
	if err := te.Start(s); err != ErrNoFactions {
		t.Errorf("Start() error = %v, want ErrNoFactions", err)
	}
}

func TestStart_IncrementsCycleNumber(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	s.CycleNumber = 4

	_ = te.Start(s)
	if s.CycleNumber != 5 {
		t.Errorf("CycleNumber: got %d, want 5", s.CycleNumber)
	}
}

// --- buildFactionOrder ---

func TestBuildFactionOrder_ContainsAllIDs(t *testing.T) {
	factions := []*domain.Faction{
		{ID: "alpha"},
		{ID: "beta"},
		{ID: "gamma"},
		{ID: "delta"},
	}

	seen := map[string]int{}
	for range 100 {
		order := buildFactionOrder(factions)
		if len(order) != len(factions) {
			t.Fatalf("order length: got %d, want %d", len(order), len(factions))
		}
		for _, id := range order {
			seen[id]++
		}
	}

	for _, f := range factions {
		if seen[f.ID] == 0 {
			t.Errorf("faction %q never appeared in order", f.ID)
		}
	}
}

func TestBuildFactionOrder_IsRotation(t *testing.T) {
	factions := []*domain.Faction{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}
	ids := []string{"a", "b", "c"}

	for range 50 {
		order := buildFactionOrder(factions)
		startIdx := -1
		for i, id := range ids {
			if id == order[0] {
				startIdx = i
				break
			}
		}
		for i, got := range order {
			want := ids[(startIdx+i)%len(ids)]
			if got != want {
				t.Errorf("order[%d] = %q, want %q (start=%d)", i, got, want, startIdx)
			}
		}
	}
}

// --- CurrentFaction ---

func TestCurrentFaction_ReturnsCorrectFaction(t *testing.T) {
	te := newTurn()
	s := newTestState("alpha", "beta", "gamma")
	_ = te.Start(s)

	f, err := te.CurrentFaction(s)
	if err != nil {
		t.Fatalf("CurrentFaction() error: %v", err)
	}
	want := s.CurrentTurn.FactionOrder[0]
	if f.ID != want {
		t.Errorf("CurrentFaction().ID = %q, want %q", f.ID, want)
	}
}

func TestCurrentFaction_ErrorIfNoTurn(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	if _, err := te.CurrentFaction(s); err != ErrNoTurnActive {
		t.Errorf("CurrentFaction() error = %v, want ErrNoTurnActive", err)
	}
}

// --- ApplyBookkeeping ---

func TestApplyBookkeeping_Income(t *testing.T) {
	tests := []struct {
		name        string
		force, cunning, wealth int
		wantIncome  int
	}{
		{"minor", 4, 3, 6, 4},   // 6/2 + (4+3)/4 = 3+1 = 4
		{"major", 6, 5, 8, 6},   // 8/2 + (6+5)/4 = 4+2 = 6
		{"low", 1, 1, 1, 0},     // 0 + 0 = 0
		{"hegemon", 8, 7, 8, 7}, // 8/2 + (8+7)/4 = 4+3 = 7
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTurn()
			s := &state.FactionState{
				CampaignID: "test",
				Factions: []*domain.Faction{
					{ID: "f", Force: tt.force, Cunning: tt.cunning, Wealth: tt.wealth, Coin: 0},
				},
			}
			_ = te.Start(s)
			if _, err := te.ApplyBookkeeping(s); err != nil {
				t.Fatalf("ApplyBookkeeping() error: %v", err)
			}
			if s.Factions[0].Coin != tt.wantIncome {
				t.Errorf("Coin after income: got %d, want %d", s.Factions[0].Coin, tt.wantIncome)
			}
		})
	}
}

func TestApplyBookkeeping_NoDoubleApply(t *testing.T) {
	te := newTurn()
	s := newTestState("a") // Force=4, Cunning=3, Wealth=6 → income=4
	_ = te.Start(s)
	_, _ = te.ApplyBookkeeping(s)
	_, _ = te.ApplyBookkeeping(s) // second call should be a no-op

	if s.Factions[0].Coin != 4 {
		t.Errorf("Coin after double apply: got %d, want 4", s.Factions[0].Coin)
	}
}

func TestApplyBookkeeping_SetsFlag(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	_ = te.Start(s)
	_, _ = te.ApplyBookkeeping(s)

	if s.CurrentTurn.Phase != domain.PhaseAction {
		t.Errorf("Phase: got %v, want PhaseAction after ApplyBookkeeping()", s.CurrentTurn.Phase)
	}
}

func TestApplyBookkeeping_ErrorIfNoTurn(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	if _, err := te.ApplyBookkeeping(s); err != ErrNoTurnActive {
		t.Errorf("ApplyBookkeeping() error = %v, want ErrNoTurnActive", err)
	}
}

// --- Advance ---

func TestAdvance_MovesToNextFaction(t *testing.T) {
	te := newTurn()
	s := newTestState("a", "b", "c")
	_ = te.Start(s)

	done, err := te.Advance(s)
	if err != nil {
		t.Fatalf("Advance() error: %v", err)
	}
	if done {
		t.Error("Advance() = true after first advance with 3 factions, want false")
	}
	if s.CurrentTurn.CurrentIndex != 1 {
		t.Errorf("CurrentIndex: got %d, want 1", s.CurrentTurn.CurrentIndex)
	}
	if s.CurrentTurn.Phase != domain.PhaseBookkeeping {
		t.Errorf("Phase: got %v, want PhaseBookkeeping after Advance()", s.CurrentTurn.Phase)
	}
}

func TestAdvance_ReturnsTrueOnCompletion(t *testing.T) {
	te := newTurn()
	s := newTestState("a", "b")
	_ = te.Start(s)

	_, _ = te.Advance(s)
	done, err := te.Advance(s)
	if err != nil {
		t.Fatalf("Advance() error: %v", err)
	}
	if !done {
		t.Error("Advance() = false after last faction, want true")
	}
	if s.CurrentTurn != nil {
		t.Error("CurrentTurn: got non-nil after turn complete, want nil")
	}
}

func TestAdvance_ErrorIfNoTurn(t *testing.T) {
	te := newTurn()
	s := newTestState("a")
	if _, err := te.Advance(s); err != ErrNoTurnActive {
		t.Errorf("Advance() error = %v, want ErrNoTurnActive", err)
	}
}

// --- Abandon ---

func TestAbandon_ClearsTurnState(t *testing.T) {
	te := newTurn()
	s := newTestState("a", "b")
	_ = te.Start(s)
	te.Abandon(s)

	if s.CurrentTurn != nil {
		t.Error("CurrentTurn: got non-nil after Abandon(), want nil")
	}
	if te.InProgress(s) {
		t.Error("InProgress() = true after Abandon(), want false")
	}
}
