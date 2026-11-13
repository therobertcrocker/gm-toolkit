package effects_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func upkeepEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:    domain.TriggerUpkeep,
		Effect:     domain.SpecialCoinUpkeep,
		Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		CoinUpkeep: &domain.CoinUpkeepParams{Amount: 1, OnUnpaid: domain.UnpaidLapse},
	}
}

func TestCoinUpkeep_DueWhileHeldLapsedOrNot(t *testing.T) {
	tag := &domain.Tag{ID: "HAWK-T-009", Name: "Outsourced"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{tag.ID}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCoinUpkeep].Register(effect.RegisterContext{Holder: effect.Holder{Faction: alpha, Tag: tag}, Entry: upkeepEntry(), Registry: registry})
	chargers := registry.UpkeepChargersFor("alpha")
	if len(chargers) != 1 {
		t.Fatalf("chargers = %d, want 1", len(chargers))
	}

	upkeep, due := chargers[0].Hook.UpkeepDue(alpha)
	if !due || upkeep.TagID != tag.ID || upkeep.Params.Amount != 1 {
		t.Errorf("held: upkeep = %+v, due = %v; want the tag's upkeep due", upkeep, due)
	}
	alpha.LapsedTags = []string{tag.ID}
	if _, due := chargers[0].Hook.UpkeepDue(alpha); !due {
		t.Error("lapsed: due = false, want the upkeep still due")
	}
	alpha.Tags = nil
	if _, due := chargers[0].Hook.UpkeepDue(alpha); due {
		t.Error("dropped: due = true, want no upkeep once the tag is gone")
	}
}

func TestCoinUpkeep_IsForTagsOnly(t *testing.T) {
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	asset := &domain.AssetDefinition{ID: "X-1", Effects: []domain.EffectDefinition{upkeepEntry()}}
	if err := engine.RegisterEffectsList(asset); err == nil || !strings.Contains(err.Error(), "only a tag pays upkeep") {
		t.Errorf("asset holder: err = %v, want an only-a-tag refusal", err)
	}
	if err := engine.CheckTagEffects(&domain.Tag{ID: "T-1", Effects: []domain.EffectDefinition{upkeepEntry()}}); err != nil {
		t.Errorf("tag holder: err = %v, want none", err)
	}
}
