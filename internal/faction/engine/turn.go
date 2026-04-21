package engine

import (
	"errors"
	"math/rand/v2"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var (
	ErrTurnInProgress = errors.New("a turn is already in progress")
	ErrNoTurnActive   = errors.New("no turn is currently in progress")
	ErrNoFactions     = errors.New("no factions in campaign")
)

// TurnEngine manages turn lifecycle: ordering, bookkeeping, step control,
// and pause/resume. It is the seam where the future Mutation and History
// engines will plug in at turn completion.
type TurnEngine struct{}

func newTurnEngine() *TurnEngine {
	return &TurnEngine{}
}

// InProgress reports whether a turn is currently active.
func (t *TurnEngine) InProgress(s *state.FactionState) bool {
	return s.CurrentTurn != nil && s.CurrentTurn.InProgress
}

// Start initializes a new turn by rolling faction order. Returns ErrTurnInProgress
// if a turn is already active.
func (t *TurnEngine) Start(s *state.FactionState) error {
	if t.InProgress(s) {
		return ErrTurnInProgress
	}
	if len(s.Factions) == 0 {
		return ErrNoFactions
	}

	s.TurnNumber++
	s.CurrentTurn = &domain.TurnState{
		InProgress:   true,
		TurnNumber:   s.TurnNumber,
		FactionOrder: buildFactionOrder(s.Factions),
		CurrentIndex: 0,
		Phase:        domain.PhaseBookkeeping,
	}
	return nil
}

// CurrentFaction returns the faction whose turn it currently is.
func (t *TurnEngine) CurrentFaction(s *state.FactionState) (*domain.Faction, error) {
	if !t.InProgress(s) {
		return nil, ErrNoTurnActive
	}
	id := s.CurrentTurn.FactionOrder[s.CurrentTurn.CurrentIndex]
	for _, f := range s.Factions {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, errors.New("faction not found: " + id)
}

// ApplyBookkeeping calculates and applies income and maintenance for the current
// faction. Safe to call on resume — skips silently if already applied this step.
func (t *TurnEngine) ApplyBookkeeping(s *state.FactionState) error {
	if !t.InProgress(s) {
		return ErrNoTurnActive
	}
	if s.CurrentTurn.Phase != domain.PhaseBookkeeping {
		return nil
	}

	f, err := t.CurrentFaction(s)
	if err != nil {
		return err
	}

	f.Coin += calcIncome(f)
	applyMaintenance(f)
	s.CurrentTurn.Phase = domain.PhaseAction
	return nil
}

// Advance marks the current faction's turn complete and moves to the next.
// Returns true when all factions have acted. The caller is responsible for
// triggering the Mutation and History engines before clearing turn state.
func (t *TurnEngine) Advance(s *state.FactionState) (bool, error) {
	if !t.InProgress(s) {
		return false, ErrNoTurnActive
	}

	s.CurrentTurn.CurrentIndex++
	s.CurrentTurn.Phase = domain.PhaseBookkeeping

	if s.CurrentTurn.CurrentIndex >= len(s.CurrentTurn.FactionOrder) {
		s.CurrentTurn = nil
		return true, nil
	}
	return false, nil
}

// Abandon clears the in-progress turn. Bookkeeping changes already staged in
// TurnState are discarded along with it — faction Coin and asset state revert
// to whatever was last persisted to disk.
func (t *TurnEngine) Abandon(s *state.FactionState) {
	s.CurrentTurn = nil
}

// calcIncome returns Coin earned this turn: floor(Wealth/2) + floor((Force+Cunning)/4).
func calcIncome(f *domain.Faction) int {
	return f.Wealth/2 + (f.Force+f.Cunning)/4
}

// applyMaintenance deducts per-asset maintenance costs. Assets that cannot be
// paid are marked unmaintained; assets already unmaintained are destroyed.
//
// Note: structured maintenance costs per asset definition are not yet modelled —
// maintenanceCost returns 0 for all assets until that data is added to AssetDefinition.
func applyMaintenance(f *domain.Faction) {
	surviving := make([]*domain.Asset, 0, len(f.Assets))
	for _, a := range f.Assets {
		cost := maintenanceCost(a)
		if cost == 0 {
			a.Maintained = true
			surviving = append(surviving, a)
			continue
		}
		if f.Coin >= cost {
			f.Coin -= cost
			a.Maintained = true
			surviving = append(surviving, a)
		} else {
			if !a.Maintained {
				// second consecutive missed payment — asset is lost
				continue
			}
			a.Maintained = false
			surviving = append(surviving, a)
		}
	}
	f.Assets = surviving
}

// maintenanceCost returns the per-turn Coin cost for an asset. Returns 0 until
// maintenance cost data is added to AssetDefinition and resolved via Rulebook.
func maintenanceCost(_ *domain.Asset) int {
	return 0
}

// buildFactionOrder rolls a starting index and returns faction IDs as a rotation.
// Die size equals the number of factions, satisfying the "no smaller than" rule.
func buildFactionOrder(factions []*domain.Faction) []string {
	n := len(factions)
	start := rand.IntN(n)
	order := make([]string, n)
	for i := range n {
		order[i] = factions[(start+i)%n].ID
	}
	return order
}
