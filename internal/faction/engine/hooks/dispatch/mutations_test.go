package dispatch

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type funcReactor func([]domain.Mutation, *state.FactionState, *rulebook.Rulebook) []domain.Mutation

func (f funcReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rb *rulebook.Rulebook) []domain.Mutation {
	return f(mutations, factionState, rb)
}

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

func TestMutationReactors_Empty(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	input := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1}}

	result, err := MutationReactors(registry, faction, input, nil, nil)

	if err != nil {
		t.Fatalf("empty registry: unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("empty registry: expected 1 mutation (passthrough), got %d", len(result))
	}
}

func TestMutationReactors_OneReactorAddsOne(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}
	added := domain.CoinDelta{FactionID: "alpha", Delta: 99}
	reactor := &oneShotReactor{result: []domain.Mutation{added}}
	registry.RegisterMutationReactor(hooks.FactionScope("alpha"), "test", reactor)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "beta", Delta: 0}}
	result, err := MutationReactors(registry, faction, initial, nil, nil)

	if err != nil {
		t.Fatalf("one reactor: unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("one reactor: expected 2 mutations, got %d: %v", len(result), result)
	}
	got, ok := result[1].(domain.CoinDelta)
	if !ok || got.Delta != 99 {
		t.Errorf("one reactor: expected CoinDelta{Delta:99} at [1], got %v", result[1])
	}
}

func TestMutationReactors_Recurses(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}

	callCount := 0
	reactor := funcReactor(func(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
		callCount++
		if callCount <= 2 {
			return []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: callCount}}
		}
		return nil
	})
	registry.RegisterMutationReactor(hooks.FactionScope("alpha"), "recurse", reactor)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 0}}
	result, err := MutationReactors(registry, faction, initial, nil, nil)

	if err != nil {
		t.Fatalf("recursion: unexpected error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("recursion: reactor called %d time(s), want ≥2", callCount)
	}
	if len(result) != 3 {
		t.Errorf("recursion: expected 3 mutations (initial + 2 reactor), got %d", len(result))
	}
}

func TestMutationReactors_DepthCap(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}

	callCount := 0
	reactor := funcReactor(func(_ []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
		callCount++
		return []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: callCount}}
	})
	registry.RegisterMutationReactor(hooks.FactionScope("alpha"), "always-adds", reactor)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 0}}
	result, err := MutationReactors(registry, faction, initial, nil, nil)

	if err == nil {
		t.Error("depth cap: expected error, got nil")
	}
	if callCount != maxHookDepth {
		t.Errorf("depth cap: reactor called %d time(s), want %d (maxHookDepth)", callCount, maxHookDepth)
	}
	if len(result) != 1+maxHookDepth {
		t.Errorf("depth cap: mutation count got %d, want %d", len(result), 1+maxHookDepth)
	}
}

func TestMutationReactors_RegistrationOrder(t *testing.T) {
	registry := hooks.NewRegistry()
	faction := &domain.Faction{ID: "alpha"}

	reactorA := &oneShotReactor{result: []domain.Mutation{domain.CoinDelta{FactionID: "A", Delta: 1}}}
	reactorB := &oneShotReactor{result: []domain.Mutation{domain.CoinDelta{FactionID: "B", Delta: 2}}}
	registry.RegisterMutationReactor(hooks.FactionScope("alpha"), "A", reactorA)
	registry.RegisterMutationReactor(hooks.FactionScope("alpha"), "B", reactorB)

	initial := []domain.Mutation{domain.CoinDelta{FactionID: "initial", Delta: 0}}
	result, err := MutationReactors(registry, faction, initial, nil, nil)

	if err != nil {
		t.Fatalf("order: unexpected error: %v", err)
	}
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
