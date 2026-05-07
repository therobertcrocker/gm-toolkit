package tags

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func TestWarlikeRollModifier_OffersOnForceAttack(t *testing.T) {
	modifier := &WarlikeRollModifier{}
	ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Attribute: string(domain.StatForce)}
	offers := modifier.OfferModifiers(ctx, nil, nil)
	if len(offers) != 1 {
		t.Fatalf("offers: got %d, want 1", len(offers))
	}
	if offers[0].BudgetKey != "tag:Warlike" {
		t.Errorf("BudgetKey: got %q, want %q", offers[0].BudgetKey, "tag:Warlike")
	}
}

func TestWarlikeRollModifier_Apply_AddsDieAndSetsKeepHighest(t *testing.T) {
	modifier := &WarlikeRollModifier{}
	ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Attribute: string(domain.StatForce)}
	offers := modifier.OfferModifiers(ctx, nil, nil)

	var rollState hooks.RollState
	offers[0].Apply(&rollState)

	if got := rollState.ExtraDice(); len(got) != 1 || got[0] != 10 {
		t.Errorf("ExtraDice: got %v, want [10]", got)
	}
	if got := rollState.KeepHighest(); got != 1 {
		t.Errorf("KeepHighest: got %d, want 1", got)
	}
}

func TestWarlikeRollModifier_NoOfferOnCunningAttack(t *testing.T) {
	modifier := &WarlikeRollModifier{}
	ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Attribute: string(domain.StatCunning)}
	if got := modifier.OfferModifiers(ctx, nil, nil); len(got) != 0 {
		t.Errorf("expected no offers for Cunning attack, got %d", len(got))
	}
}

func TestWarlikeRollModifier_NoOfferOnDefense(t *testing.T) {
	modifier := &WarlikeRollModifier{}
	ctx := hooks.RollContext{Phase: hooks.PhaseDefense, Attribute: string(domain.StatForce)}
	if got := modifier.OfferModifiers(ctx, nil, nil); len(got) != 0 {
		t.Errorf("expected no offers for defense roll, got %d", len(got))
	}
}

func TestWarlikeRollModifier_NoOfferOnFactionTest(t *testing.T) {
	modifier := &WarlikeRollModifier{}
	ctx := hooks.RollContext{Phase: hooks.PhaseFactionTest, Attribute: string(domain.StatForce)}
	if got := modifier.OfferModifiers(ctx, nil, nil); len(got) != 0 {
		t.Errorf("expected no offers for faction test, got %d", len(got))
	}
}
