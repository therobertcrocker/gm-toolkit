package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func liveAsset(id, ownerID, location string) *domain.Asset {
	return &domain.Asset{
		ID: id, OwnerID: ownerID, Location: location,
		CurrentHP: 8, Ready: true, Maintained: true,
	}
}

func indexWithAssets(assets ...*domain.Asset) *world.Index {
	idx := &world.Index{
		AssetsByLocation: make(map[string][]*domain.Asset),
		BasesByLocation:  make(map[string][]*domain.Base),
	}
	for _, asset := range assets {
		idx.AssetsByLocation[asset.Location] = append(idx.AssetsByLocation[asset.Location], asset)
	}
	return idx
}

func TestEligibleDefenders_ReturnsRivalAssets(t *testing.T) {
	f1Asset := liveAsset("a1", "f1", "krylos")   // own — excluded
	f2Asset := liveAsset("a2", "f2", "krylos")   // rival — included
	f3Asset := liveAsset("a3", "f3", "tartarus") // wrong fragment — excluded

	idx := indexWithAssets(f1Asset, f2Asset, f3Asset)
	result := eligibleDefenders("f1", "krylos", idx)

	if len(result) != 1 || result[0].ID != "a2" {
		t.Errorf("eligibleDefenders = %v, want [a2]", result)
	}
}

func TestEligibleDefenders_ExcludesStealthy(t *testing.T) {
	stealthy := liveAsset("a2", "f2", "krylos")
	stealthy.Stealthy = true

	result := eligibleDefenders("f1", "krylos", indexWithAssets(stealthy))
	if len(result) != 0 {
		t.Errorf("expected stealthy asset excluded, got %v", result)
	}
}

func TestEligibleDefenders_ExcludesDead(t *testing.T) {
	dead := liveAsset("a2", "f2", "krylos")
	dead.CurrentHP = 0

	result := eligibleDefenders("f1", "krylos", indexWithAssets(dead))
	if len(result) != 0 {
		t.Errorf("expected dead asset excluded, got %v", result)
	}
}

func TestEligibleDefenders_ExcludesNotReady(t *testing.T) {
	notReady := liveAsset("a2", "f2", "krylos")
	notReady.Ready = false

	result := eligibleDefenders("f1", "krylos", indexWithAssets(notReady))
	if len(result) != 0 {
		t.Errorf("expected not-Ready asset excluded, got %v", result)
	}
}

func TestEligibleDefenders_ExcludesUnmaintained(t *testing.T) {
	unmaintained := liveAsset("a2", "f2", "krylos")
	unmaintained.Maintained = false

	result := eligibleDefenders("f1", "krylos", indexWithAssets(unmaintained))
	if len(result) != 0 {
		t.Errorf("expected unmaintained asset excluded, got %v", result)
	}
}

func TestLiveDefenders_ExcludesTrackerDead(t *testing.T) {
	live := liveAsset("a2", "f2", "krylos")
	halfDead := liveAsset("a3", "f2", "krylos")
	halfDead.CurrentHP = 4 // tracker drives it to exactly 0

	idx := indexWithAssets(live, halfDead)
	tracker := map[string]int{"a3": -4}

	result := liveDefenders("f1", "krylos", tracker, idx)
	if len(result) != 1 || result[0].ID != "a2" {
		t.Errorf("liveDefenders = %v, want [a2]", result)
	}
}

func TestRivalsOnWorld_ReturnsRivalFaction(t *testing.T) {
	f2Asset := liveAsset("a2", "f2", "krylos")
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {ID: "f1"},
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{"a2": f2Asset}},
		},
	}

	result := rivalsOnWorld(factionState, "f1", "krylos", indexWithAssets(f2Asset))
	if len(result) != 1 || result[0].ID != "f2" {
		t.Errorf("rivalsOnWorld = %v, want [f2]", result)
	}
}

func TestRivalsOnWorld_DeduplicatesMultiAssetFaction(t *testing.T) {
	a2 := liveAsset("a2", "f2", "krylos")
	a3 := liveAsset("a3", "f2", "krylos")
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {ID: "f1"},
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{"a2": a2, "a3": a3}},
		},
	}

	result := rivalsOnWorld(factionState, "f1", "krylos", indexWithAssets(a2, a3))
	if len(result) != 1 {
		t.Errorf("rivalsOnWorld = %v, want exactly one entry for f2", result)
	}
}

func TestSeizePlanetTargetWorlds_ReturnsContestedFragments(t *testing.T) {
	f1Krylos := liveAsset("a1", "f1", "krylos")
	f1Tartarus := liveAsset("a2", "f1", "tartarus") // no rival here — excluded
	f2Krylos := liveAsset("a3", "f2", "krylos")

	f1 := &domain.Faction{
		ID:     "f1",
		Assets: map[string]*domain.Asset{"a1": f1Krylos, "a2": f1Tartarus},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": f1,
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{"a3": f2Krylos}},
		},
	}

	result := seizePlanetTargetWorlds(f1, indexFromState(factionState))
	if len(result) != 1 || result[0] != "krylos" {
		t.Errorf("seizePlanetTargetWorlds = %v, want [krylos]", result)
	}
}
