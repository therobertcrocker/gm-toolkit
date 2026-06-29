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

	asset := domain.NewPurchasedAsset(faction, def, "world-a")

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

func TestNewPurchasedAsset_MonotonicID(t *testing.T) {
	def := &domain.AssetDefinition{ID: "SWN-F1-001", HP: 4}
	faction := &domain.Faction{
		ID: "f1",
		Assets: map[string]*domain.Asset{
			"f1-SWN-F1-001-1": {ID: "f1-SWN-F1-001-1"},
			"f1-SWN-F1-001-2": {ID: "f1-SWN-F1-001-2"},
		},
	}
	asset := domain.NewPurchasedAsset(faction, def, "w1")
	if asset.ID != "f1-SWN-F1-001-3" {
		t.Errorf("ID = %q, want %q", asset.ID, "f1-SWN-F1-001-3")
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
