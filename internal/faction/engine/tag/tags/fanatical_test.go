package tags

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func TestFanaticalRollResultHook_FlagsOnesForReroll(t *testing.T) {
	hook := &FanaticalRollResultHook{}
	result := hooks.RollResult{Dice: []int{5, 1, 3, 1}}
	directive := hook.OnRollResult(hooks.RollContext{}, result, nil, nil)

	want := []int{1, 3}
	if !reflect.DeepEqual(directive.Indices, want) {
		t.Errorf("Indices: got %v, want %v", directive.Indices, want)
	}
	if directive.Elective {
		t.Error("directive should not be elective")
	}
}

func TestFanaticalRollResultHook_NoOnesNoReroll(t *testing.T) {
	hook := &FanaticalRollResultHook{}
	result := hooks.RollResult{Dice: []int{5, 3, 7}}
	directive := hook.OnRollResult(hooks.RollContext{}, result, nil, nil)
	if len(directive.Indices) != 0 {
		t.Errorf("expected empty Indices when no 1s, got %v", directive.Indices)
	}
}

func TestFanaticalTieResolver_AttackerLosesTie(t *testing.T) {
	resolver := &FanaticalTieResolver{FactionID: "alpha"}
	ctx := hooks.RollContext{
		Actor:    &domain.Faction{ID: "alpha"},
		Opponent: &domain.Faction{ID: "beta"},
	}
	if got := resolver.ResolveTie(ctx, nil); got != hooks.TieDefenderWins {
		t.Errorf("ResolveTie: got %v, want TieDefenderWins when Fanatical is attacker", got)
	}
}

func TestFanaticalTieResolver_DefenderLosesTie(t *testing.T) {
	resolver := &FanaticalTieResolver{FactionID: "alpha"}
	ctx := hooks.RollContext{
		Actor:    &domain.Faction{ID: "beta"},
		Opponent: &domain.Faction{ID: "alpha"},
	}
	if got := resolver.ResolveTie(ctx, nil); got != hooks.TieAttackerWins {
		t.Errorf("ResolveTie: got %v, want TieAttackerWins when Fanatical is defender", got)
	}
}
