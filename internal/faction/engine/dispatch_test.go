package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/eventhooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// funcReactor adapts a function to eventhooks.MutationReactor.
type funcReactor func([]domain.Mutation, *state.FactionState, *rulebook.Rulebook) []domain.Mutation

func (f funcReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rb *rulebook.Rulebook) []domain.Mutation {
	return f(mutations, factionState, rb)
}

// oneShotReactor implements MutationReactor and fires exactly once.
type oneShotReactor struct {
	called bool
	result []domain.Mutation
}

func (r *oneShotReactor) OnMutations(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if r.called {
		return nil
	}
	r.called = true
	return r.result
}

// nopObserver satisfies TurnObserver and records OnError calls.
type nopObserver struct {
	errors []error
}

func (n *nopObserver) OnFactionTurnStarted(_ *domain.Faction)                                              {}
func (n *nopObserver) OnFactionSkipped(_ *domain.Faction)                                                  {}
func (n *nopObserver) OnGoalLockApplied(_ *domain.Faction, _ goal.GoalLock, _ []domain.Mutation)          {}
func (n *nopObserver) OnBookkeepingApplied(_ *domain.Faction, _ turn.BookkeepingResult, _ []domain.Mutation) {}
func (n *nopObserver) OnActionSelected(_ *domain.Faction, _ action.Action)                                 {}
func (n *nopObserver) OnActionResolved(_ *domain.Faction, _ action.Action, _ []domain.Mutation)            {}
func (n *nopObserver) OnFactionTurnCompleted(_ *domain.Faction)                                             {}
func (n *nopObserver) OnCycleCompleted(_ int, _ *state.FactionState)                                       {}
func (n *nopObserver) OnError(_ *domain.Faction, err error) {
	n.errors = append(n.errors, err)
}

func TestDispatchMutationReactors_EmptyRegistry(t *testing.T) {
	registry := eventhooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	input := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1}}
	obs := &nopObserver{}

	result := dispatchMutationReactors(registry, faction, input, nil, nil, obs)

	if len(result) != 1 {
		t.Errorf("empty registry: expected 1 mutation (passthrough), got %d", len(result))
	}
	if len(obs.errors) != 0 {
		t.Errorf("empty registry: expected no errors, got %d", len(obs.errors))
	}
}

func TestDispatchMutationReactors_OneReactorAddsOne(t *testing.T) {
	registry := eventhooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	added := domain.CoinDelta{FactionID: "alpha", Delta: 99}
	reactor := &oneShotReactor{result: []domain.Mutation{added}}
	registry.RegisterMutationReactor(eventhooks.FactionScope("alpha"), "test", reactor)
	obs := &nopObserver{}

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "beta", Delta: 0}}
	result := dispatchMutationReactors(registry, faction, initial, nil, nil, obs)

	if len(result) != 2 {
		t.Fatalf("one reactor: expected 2 mutations, got %d: %v", len(result), result)
	}
	got, ok := result[1].(domain.CoinDelta)
	if !ok || got.Delta != 99 {
		t.Errorf("one reactor: expected CoinDelta{Delta:99} at [1], got %v", result[1])
	}
}

func TestDispatchMutationReactors_Recurses(t *testing.T) {
	registry := eventhooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	obs := &nopObserver{}

	callCount := 0
	reactor := funcReactor(func(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
		callCount++
		if callCount <= 2 {
			return []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: callCount}}
		}
		return nil
	})
	registry.RegisterMutationReactor(eventhooks.FactionScope("alpha"), "recurse", reactor)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 0}}
	result := dispatchMutationReactors(registry, faction, initial, nil, nil, obs)

	if callCount < 2 {
		t.Errorf("recursion: reactor called %d time(s), want ≥2", callCount)
	}
	if len(result) != 3 {
		t.Errorf("recursion: expected 3 mutations (initial + 2 reactor), got %d", len(result))
	}
}

func TestDispatchMutationReactors_DepthCap(t *testing.T) {
	registry := eventhooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	obs := &nopObserver{}

	callCount := 0
	reactor := funcReactor(func(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
		callCount++
		return []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: callCount}}
	})
	registry.RegisterMutationReactor(eventhooks.FactionScope("alpha"), "always-adds", reactor)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 0}}
	result := dispatchMutationReactors(registry, faction, initial, nil, nil, obs)

	if callCount != maxHookDepth {
		t.Errorf("depth cap: reactor called %d time(s), want %d (maxHookDepth)", callCount, maxHookDepth)
	}
	if len(obs.errors) != 1 {
		t.Errorf("depth cap: expected 1 cap error, got %d", len(obs.errors))
	}
	if len(result) != 1+maxHookDepth {
		t.Errorf("depth cap: mutation count got %d, want %d", len(result), 1+maxHookDepth)
	}
}

func TestDispatchMutationReactors_RegistrationOrder(t *testing.T) {
	registry := eventhooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	obs := &nopObserver{}

	reactorA := &oneShotReactor{result: []domain.Mutation{domain.CoinDelta{FactionID: "A", Delta: 1}}}
	reactorB := &oneShotReactor{result: []domain.Mutation{domain.CoinDelta{FactionID: "B", Delta: 2}}}
	registry.RegisterMutationReactor(eventhooks.FactionScope("alpha"), "A", reactorA)
	registry.RegisterMutationReactor(eventhooks.FactionScope("alpha"), "B", reactorB)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "initial", Delta: 0}}
	result := dispatchMutationReactors(registry, faction, initial, nil, nil, obs)

	if len(result) != 3 {
		t.Fatalf("order: expected 3 mutations (initial+A+B), got %d", len(result))
	}
	coinA, okA := result[1].(domain.CoinDelta)
	coinB, okB := result[2].(domain.CoinDelta)
	if !okA || !okB {
		t.Fatalf("order: expected CoinDelta at [1] and [2], got %T and %T", result[1], result[2])
	}
	if coinA.FactionID != "A" || coinB.FactionID != "B" {
		t.Errorf("order: got [%s, %s], want [A, B]", coinA.FactionID, coinB.FactionID)
	}
}
