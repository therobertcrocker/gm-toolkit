package dice

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// stubTieResolver always returns a fixed outcome.
type stubTieResolver struct{ outcome hooks.TieOutcome }

func (stub *stubTieResolver) ResolveTie(_ hooks.RollContext, _ *state.FactionState) hooks.TieOutcome {
	return stub.outcome
}

func TestResolveTie_NoResolver(t *testing.T) {
	registry := hooks.NewRegistry()
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieStandard {
		t.Errorf("ResolveTie = %v, want TieStandard when no resolver registered", got)
	}
}

func TestResolveTie_DefenderWins(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.GlobalScope(), "fanatical", &stubTieResolver{outcome: hooks.TieDefenderWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}}
	got := ResolveTie(registry, ctx, nil)
	if got != hooks.TieDefenderWins {
		t.Errorf("ResolveTie = %v, want TieDefenderWins", got)
	}
}

func TestResolveTie_AssetResolverAnswers(t *testing.T) {
	registry := hooks.NewRegistry()
	attacker := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": {ID: "a1"}}}
	registry.RegisterTieResolver(hooks.AssetScope("f1", "a1"), "asset", &stubTieResolver{outcome: hooks.TieAttackerWins})
	if got := ResolveTie(registry, hooks.RollContext{Actor: attacker}, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want TieAttackerWins from the attacker's asset", got)
	}
}

func TestResolveTie_StandardDoesNotEndTheSearch(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterTieResolver(hooks.FactionScope("f1"), "silent", &stubTieResolver{outcome: hooks.TieStandard})
	registry.RegisterTieResolver(hooks.FactionScope("f2"), "loses", &stubTieResolver{outcome: hooks.TieAttackerWins})
	ctx := hooks.RollContext{Actor: &domain.Faction{ID: "f1"}, Opponent: &domain.Faction{ID: "f2"}}
	if got := ResolveTie(registry, ctx, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie = %v, want the defender's TieAttackerWins past the attacker's TieStandard", got)
	}
}
