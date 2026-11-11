package engine

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func TestNewWithRulebook_ValidatesEffects(t *testing.T) {
	def := &domain.AssetDefinition{ID: "T-001", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialCoinGain,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
	}}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), `asset "T-001" effects[0]`) {
		t.Fatalf("got %v, want a validation error naming the entry", err)
	}
}

func TestNewWithRulebook_ValidatesTagEffects(t *testing.T) {
	tag := &domain.Tag{ID: "SWN-T-020", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{}, Tags: map[string]*domain.Tag{tag.ID: tag}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), `tag "SWN-T-020" effects[0]`) || !strings.Contains(err.Error(), "a tag has no asset") {
		t.Fatalf("got %v, want a tag validation error naming the entry", err)
	}
}

func TestNewWithRulebook_RefusesBadGoalData(t *testing.T) {
	goal := &domain.Goal{
		ID:        "X-1",
		Name:      "Bad Goal",
		Progress:  &domain.GoalProgress{Event: "kill", Amount: "one"},
		Threshold: &domain.GoalThreshold{Constant: 2},
		XP:        &domain.GoalXP{Formula: "double"},
	}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{}, Goals: map[string]*domain.Goal{goal.ID: goal}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), "engine: goal") || !strings.Contains(err.Error(), "unknown formula") {
		t.Fatalf("got %v, want a goal error that names the unknown formula", err)
	}
}

// TestNewWithRulebook_ShippedRulebook catches an authored effects entry that
// loads but fails registrar validation, which otherwise surfaces only at runtime.
func TestNewWithRulebook_ShippedRulebook(t *testing.T) {
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/swn", false); err != nil {
		t.Fatalf("compose shipped rulebook: %v", err)
	}
	rb, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("load shipped rulebook: %v", err)
	}
	if _, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("construct engine from shipped rulebook: %v", err)
	}
}

func TestNewWithRulebook_RejectsTransportAndEffectsOnOneDefinition(t *testing.T) {
	def := &domain.AssetDefinition{
		ID:        "T-001",
		Transport: &domain.TransportProfile{MaxHex: 1},
		Effects: []domain.EffectDefinition{{
			Trigger:  domain.TriggerTurnStart,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		}},
	}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{def.ID: def}}
	_, err := NewWithRulebook(rb, nil, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("got %v, want an already-registered error", err)
	}
}
