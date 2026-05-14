package goals

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func indexWithRivalAsset(locationID, ownerID string) *world.Index {
	return &world.Index{
		AssetsByLocation: map[string][]*domain.Asset{
			locationID: {{ID: "a1", OwnerID: ownerID, Location: locationID}},
		},
		BasesByLocation: make(map[string][]*domain.Base),
	}
}

func indexWithRivalBase(locationID, ownerID string) *world.Index {
	return &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation: map[string][]*domain.Base{
			locationID: {{ID: "b1", OwnerID: ownerID, Location: locationID}},
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

func TestRivalHasPlanetaryGovernmentOnWorld_RivalHoldsBase(t *testing.T) {
	rivalBase := &domain.Base{ID: "b1", OwnerID: "f2", Location: "krylos"}
	rivalFaction := &domain.Faction{
		ID:   "f2",
		Tags: []*domain.Tag{{ID: "T-011"}},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{"f1": {ID: "f1"}, "f2": rivalFaction},
	}
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  map[string][]*domain.Base{"krylos": {rivalBase}},
	}

	if !rivalHasPlanetaryGovernmentOnWorld("f1", "krylos", factionState, idx) {
		t.Error("expected true when rival holds PG base on world")
	}
}

func TestRivalHasPlanetaryGovernmentOnWorld_RivalNoTag(t *testing.T) {
	rivalBase := &domain.Base{ID: "b1", OwnerID: "f2", Location: "krylos"}
	rivalFaction := &domain.Faction{ID: "f2"} // no tags
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{"f1": {ID: "f1"}, "f2": rivalFaction},
	}
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  map[string][]*domain.Base{"krylos": {rivalBase}},
	}

	if rivalHasPlanetaryGovernmentOnWorld("f1", "krylos", factionState, idx) {
		t.Error("expected false when rival has base but no PG tag")
	}
}

func TestRivalHasPlanetaryGovernmentOnWorld_OwnBaseIgnored(t *testing.T) {
	ownBase := &domain.Base{ID: "b1", OwnerID: "f1", Location: "krylos"}
	actingFaction := &domain.Faction{
		ID:   "f1",
		Tags: []*domain.Tag{{ID: "T-011"}},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{"f1": actingFaction},
	}
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  map[string][]*domain.Base{"krylos": {ownBase}},
	}

	if rivalHasPlanetaryGovernmentOnWorld("f1", "krylos", factionState, idx) {
		t.Error("expected false when only own base is present")
	}
}
