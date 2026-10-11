package engine

import (
	"log/slog"
	"strings"
	"testing"

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
