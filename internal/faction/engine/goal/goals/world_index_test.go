package goals

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
)

func indexWithRivalAsset(locationID, ownerID string) *world.Index {
	return &world.Index{
		AssetsByLocation: map[string][]*domain.Asset{
			locationID: {{ID: "a1", OwnerID: ownerID, Location: domain.Location{WorldID: locationID}}},
		},
		BasesByLocation: make(map[string][]*domain.Base),
	}
}

func indexWithRivalBase(locationID, ownerID string) *world.Index {
	return &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation: map[string][]*domain.Base{
			locationID: {{ID: "b1", OwnerID: ownerID, Location: domain.Location{WorldID: locationID}}},
		},
	}
}

func TestWorldHasRivalPresence_RivalAsset(t *testing.T) {
	idx := indexWithRivalAsset("krylos", "f2")
	if !worldHasRivalPresence("krylos", "f1", idx) {
		t.Error("expected true when rival asset is on world")
	}
}

func TestWorldHasRivalPresence_OwnAssetOnly(t *testing.T) {
	idx := indexWithRivalAsset("krylos", "f1")
	if worldHasRivalPresence("krylos", "f1", idx) {
		t.Error("expected false when only own assets are present")
	}
}

func TestWorldHasRivalPresence_RivalBaseOnly(t *testing.T) {
	idx := indexWithRivalBase("krylos", "f2")
	if !worldHasRivalPresence("krylos", "f1", idx) {
		t.Error("expected true when rival base is on world with no assets")
	}
}
