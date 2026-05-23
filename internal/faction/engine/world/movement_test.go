package world_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// regionWorldEngine creates a WorldEngine backed by a minimal RegionMap with one
// region "void" containing hexes (0,0)–(4,0) along the Q axis.
func regionWorldEngine(t *testing.T) *world.WorldEngine {
	t.Helper()
	dir := t.TempDir()
	regionsToml := "[[region]]\nid = \"void\"\nname = \"The Void\"\nhexes = [[0,0],[1,0],[2,0],[3,0],[4,0]]\n"
	for _, entry := range []struct{ name, body string }{
		{"regions.toml", regionsToml},
		{"worlds.toml", ""},
	} {
		if err := os.WriteFile(filepath.Join(dir, entry.name), []byte(entry.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	we, err := world.New(dir, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("regionWorldEngine: %v", err)
	}
	return we
}

func reviseRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"test-def": {ID: "test-def", Speed: 2, DriftRating: 1},
		},
		DriftCosts: []int{5},
	}
}

func TestBuildMovementMutations_RevisionCost(t *testing.T) {
	we := regionWorldEngine(t)

	currentHex := spatial.HexCoord{Q: 1, R: 0}
	destHex := spatial.HexCoord{Q: 3, R: 0}

	faction := &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"asset-1": {
				ID:           "asset-1",
				DefinitionID: "test-def",
				OwnerID:      "alpha",
				Location:     domain.Location{RegionHex: spatial.RegionHex{Coord: currentHex, RegionID: "void"}},
				CurrentOrder: &domain.MovementOrder{
					AssetID: "asset-1",
					Path:    makePath(5),
					StepIdx: 1,
				},
			},
		},
	}

	decisions := []world.MovementDecision{
		{
			AssetID:     "asset-1",
			Kind:        world.MovementDecisionRevise,
			Destination: &domain.Location{RegionHex: spatial.RegionHex{Coord: destHex, RegionID: "void"}},
		},
	}

	mutations, err := we.BuildMovementMutations(decisions, faction, reviseRulebook(), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("expected 2 mutations, got %d", len(mutations))
	}

	coinDelta, ok := mutations[0].(domain.CoinDelta)
	if !ok {
		t.Fatalf("expected CoinDelta as first mutation, got %T", mutations[0])
	}
	if coinDelta.Delta != -1 {
		t.Errorf("expected CoinDelta.Delta=-1, got %d", coinDelta.Delta)
	}

	_, ok = mutations[1].(domain.MovementOrderRevised)
	if !ok {
		t.Fatalf("expected MovementOrderRevised as second mutation, got %T", mutations[1])
	}
}
