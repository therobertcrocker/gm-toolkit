package engine

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// TestHawkSeed_MatchesRulebookAndMap catches a seed that names a definition, tag
// or world the H.A.W.K rulebook or map no longer has. Efforts 4 to 6 replace the
// SWN stand-ins, and without this check a stale ID shows up only in a live turn.
func TestHawkSeed_MatchesRulebookAndMap(t *testing.T) {
	camp := &campaigns.Campaign{Root: t.TempDir()}
	if _, err := campaigns.ComposeRulebook(camp, "../../../rulebooks/hawk", false); err != nil {
		t.Fatalf("compose hawk rulebook: %v", err)
	}
	rb, err := rulebook.Load(camp.Paths().FactionDataDir)
	if err != nil {
		t.Fatalf("load hawk rulebook: %v", err)
	}
	regionMap, err := spatial.LoadRegionMap("../../../campaigns/hawk/spatial")
	if err != nil {
		t.Fatalf("load hawk map: %v", err)
	}
	seed, err := state.Load("../../../campaigns/hawk/state/state.seed.toml")
	if err != nil {
		t.Fatalf("load hawk seed: %v", err)
	}
	if len(seed.Factions) == 0 {
		t.Fatal("hawk seed has no factions")
	}

	checkLocation := func(what string, location domain.Location) {
		t.Helper()
		world, ok := regionMap.Location(location.WorldID)
		if !ok {
			t.Errorf("%s: world %q is not on the map", what, location.WorldID)
			return
		}
		if worldHex := world.(spatial.RegionLocation).RegionHex(); worldHex != location.RegionHex {
			t.Errorf("%s: region hex %v, but world %q sits at %v", what, location.RegionHex, location.WorldID, worldHex)
		}
	}

	for factionID, faction := range seed.Factions {
		checkLocation(factionID+" homeworld", faction.Homeworld)
		for _, tagID := range faction.Tags {
			if _, ok := rb.Tags[tagID]; !ok {
				t.Errorf("%s: tag %q is not in the rulebook", factionID, tagID)
			}
		}
		for _, base := range faction.Bases {
			checkLocation(base.ID, base.Location)
		}
		for assetID, asset := range faction.Assets {
			if _, ok := rb.Assets[asset.DefinitionID]; !ok {
				t.Errorf("%s: definition %q is not in the rulebook", assetID, asset.DefinitionID)
			}
			checkLocation(assetID, asset.Location)
		}
	}
	for worldID, factionID := range seed.Governments {
		if _, ok := regionMap.Location(worldID); !ok {
			t.Errorf("government of %q: world is not on the map", worldID)
		}
		if _, ok := seed.Factions[factionID]; !ok {
			t.Errorf("government of %q: faction %q is not in the seed", worldID, factionID)
		}
	}
}
