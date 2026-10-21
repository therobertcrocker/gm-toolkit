package domain_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestNewPurchasedAsset(t *testing.T) {
	faction := &domain.Faction{
		ID:     "f1",
		Assets: map[string]*domain.Asset{},
	}
	def := &domain.AssetDefinition{ID: "SWN-F1-001", HP: 6, Cost: 4}

	asset := domain.NewPurchasedAsset(nil, faction, def, "world-a")

	if asset.DefinitionID != def.ID {
		t.Errorf("DefinitionID = %q, want %q", asset.DefinitionID, def.ID)
	}
	if asset.OwnerID != faction.ID {
		t.Errorf("OwnerID = %q, want %q", asset.OwnerID, faction.ID)
	}
	if asset.Location.WorldID != "world-a" {
		t.Errorf("Location.WorldID = %q, want %q", asset.Location.WorldID, "world-a")
	}
	if asset.CurrentHP != def.HP {
		t.Errorf("CurrentHP = %d, want %d", asset.CurrentHP, def.HP)
	}
	if asset.Ready {
		t.Error("Ready = true, want false (activates next turn)")
	}
	if !asset.Maintained {
		t.Error("Maintained = false, want true")
	}
}

func TestAsset_Active(t *testing.T) {
	cases := []struct {
		name  string
		asset domain.Asset
		want  bool
	}{
		{"ready, maintained, alive", domain.Asset{Ready: true, Maintained: true, CurrentHP: 1}, true},
		{"not ready", domain.Asset{Ready: false, Maintained: true, CurrentHP: 1}, false},
		{"unmaintained", domain.Asset{Ready: true, Maintained: false, CurrentHP: 1}, false},
		{"destroyed", domain.Asset{Ready: true, Maintained: true, CurrentHP: 0}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.asset.Active(); got != tc.want {
				t.Errorf("Active() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewPurchasedAsset_MonotonicID(t *testing.T) {
	def := &domain.AssetDefinition{ID: "SWN-F1-001", HP: 4}
	faction := &domain.Faction{
		ID: "f1",
		Assets: map[string]*domain.Asset{
			"f1-SWN-F1-001-1": {ID: "f1-SWN-F1-001-1"},
			"f1-SWN-F1-001-2": {ID: "f1-SWN-F1-001-2"},
		},
	}
	asset := domain.NewPurchasedAsset(nil, faction, def, "w1")
	if asset.ID != "f1-SWN-F1-001-3" {
		t.Errorf("ID = %q, want %q", asset.ID, "f1-SWN-F1-001-3")
	}
}

func TestNextAssetID_ScansEveryFaction(t *testing.T) {
	owner := &domain.Faction{
		ID:     "f1",
		Assets: map[string]*domain.Asset{"f1-SWN-F1-001-1": {ID: "f1-SWN-F1-001-1"}},
	}
	captor := &domain.Faction{
		ID:     "f2",
		Assets: map[string]*domain.Asset{"f1-SWN-F1-001-3": {ID: "f1-SWN-F1-001-3"}},
	}
	factions := map[string]*domain.Faction{"f1": owner, "f2": captor}

	if got := domain.NextAssetID(factions, owner, "SWN-F1-001"); got != "f1-SWN-F1-001-4" {
		t.Errorf("NextAssetID = %q, want %q (a captured asset still holds its suffix)", got, "f1-SWN-F1-001-4")
	}
}

func TestHasPlanetaryGovernment(t *testing.T) {
	owned := &domain.Base{OwnerID: "f1"}
	rival := &domain.Base{OwnerID: "f2"}

	if !domain.HasPlanetaryGovernment("f1", []*domain.Base{owned}) {
		t.Error("want true when faction owns a base")
	}
	if domain.HasPlanetaryGovernment("f1", []*domain.Base{rival}) {
		t.Error("want false when only rival owns a base")
	}
	if domain.HasPlanetaryGovernment("f1", nil) {
		t.Error("want false for nil slice")
	}
	if !domain.HasPlanetaryGovernment("f1", []*domain.Base{rival, owned}) {
		t.Error("want true when faction base is present among rivals")
	}
}

func TestAsset_CloneCopiesConditions(t *testing.T) {
	original := &domain.Asset{ID: "a1", Conditions: []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1"}}}

	clone := original.Clone()
	clone.Conditions[0].Kind = "changed"

	if got := original.Conditions[0].Kind; got != domain.ConditionNoAbility {
		t.Errorf("original Conditions[0].Kind = %q, want %q (clone must not share the slice)", got, domain.ConditionNoAbility)
	}
}

func TestAsset_HasCondition(t *testing.T) {
	asset := &domain.Asset{ID: "a1"}
	if asset.HasCondition(domain.ConditionNoAbility) {
		t.Error("HasCondition(no_ability) = true on an asset with no conditions")
	}

	asset.Conditions = []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1"}}
	if !asset.HasCondition(domain.ConditionNoAbility) {
		t.Error("HasCondition(no_ability) = false on an asset that carries it")
	}
	if asset.HasCondition("other") {
		t.Error("HasCondition(other) = true; only no_ability is present")
	}
}
