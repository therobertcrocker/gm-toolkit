package effects_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

func stealthQuality() *domain.AssetDefinition {
	stealthy := false
	return &domain.AssetDefinition{ID: "SWN-C3-002", Name: "Stealth", Quality: true, Effects: []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialGrantStealth,
		Filters: domain.EffectFilters{
			Owner:      domain.OwnerOwn,
			Scope:      domain.ScopeWorld,
			AssetTypes: []domain.AssetType{domain.TypeSpecialForces},
			Stealthy:   &stealthy,
		},
	}}}
}

func TestGrantStealth_TargetsUnstealthedSpecialForcesOnTheWorld(t *testing.T) {
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"sf":  {ID: "sf", Type: domain.TypeSpecialForces},
		"mil": {ID: "mil", Type: domain.TypeMilitaryUnit},
	}}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{
		"a-2": {ID: "a-2", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}},
		"a-1": {ID: "a-1", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}},
		"a-3": {ID: "a-3", DefinitionID: "sf", Location: domain.Location{WorldID: "Krylos"}, Stealthy: true},
		"a-4": {ID: "a-4", DefinitionID: "mil", Location: domain.Location{WorldID: "Krylos"}},
		"a-5": {ID: "a-5", DefinitionID: "sf", Location: domain.Location{WorldID: "Tartarus"}},
	}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	var ids []string
	for _, target := range engine.QualityTargets(stealthQuality(), faction, "Krylos", rb) {
		ids = append(ids, target.ID)
	}
	if !reflect.DeepEqual(ids, []string{"a-1", "a-2"}) {
		t.Errorf("targets = %v, want [a-1 a-2] (unstealthed Special Forces on Krylos, in ID order)", ids)
	}
}

func TestGrantStealth_StealthsTheTarget(t *testing.T) {
	faction := &domain.Faction{ID: "alpha"}
	target := &domain.Asset{ID: "a-1", OwnerID: "alpha"}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	got := engine.FireQuality(faction, stealthQuality(), target, nil)
	want := []domain.Mutation{domain.AssetStealthApplied{FactionID: "alpha", AssetID: "a-1", Cause: "grant_stealth", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}
