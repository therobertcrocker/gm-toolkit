package world_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
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

	mutations, err := we.BuildMovementMutations(decisions, faction, reviseRulebook(), nil, slog.New(slog.DiscardHandler))
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

func TestBuildMovementMutations_RevisedOrderStoresTheResolvedDriftRating(t *testing.T) {
	we := regionWorldEngine(t)
	rules := &rulebook.Rulebook{
		Assets:     map[string]*domain.AssetDefinition{"unrated-def": {ID: "unrated-def", Speed: 2}},
		DriftCosts: []int{5},
	}
	faction := &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"asset-1": {
				ID:           "asset-1",
				DefinitionID: "unrated-def",
				OwnerID:      "alpha",
				Location:     domain.Location{RegionHex: spatial.RegionHex{Coord: spatial.HexCoord{Q: 1}, RegionID: "void"}},
				CurrentOrder: &domain.MovementOrder{AssetID: "asset-1", Path: makePath(5), StepIdx: 1},
			},
		},
	}
	decisions := []world.MovementDecision{{
		AssetID:     "asset-1",
		Kind:        world.MovementDecisionRevise,
		Destination: &domain.Location{RegionHex: spatial.RegionHex{Coord: spatial.HexCoord{Q: 3}, RegionID: "void"}},
	}}

	mutations, err := we.BuildMovementMutations(decisions, faction, rules, nil, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	revised, ok := mutations[len(mutations)-1].(domain.MovementOrderRevised)
	if !ok {
		t.Fatalf("expected MovementOrderRevised last, got %T", mutations[len(mutations)-1])
	}
	if revised.NewOrder.DriftRating != 1 {
		t.Errorf("NewOrder.DriftRating = %d, want 1, the rating the path was priced at", revised.NewOrder.DriftRating)
	}
}

func TestTickMovementOrders_InAssetIDOrder(t *testing.T) {
	path := []spatial.RegionHex{{RegionID: "void", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "void", Coord: spatial.HexCoord{Q: 1}}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{}}
	for _, id := range []string{"c", "a", "d", "b"} {
		faction.Assets[id] = &domain.Asset{ID: id, OwnerID: "f1", DefinitionID: "test-def",
			CurrentOrder: &domain.MovementOrder{AssetID: id, Path: path}}
	}

	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, reviseRulebook(), nil, slog.New(slog.DiscardHandler))

	if err != nil {
		t.Fatalf("TickMovementOrders: %v", err)
	}
	var order []string
	for _, mutation := range mutations {
		order = append(order, mutation.(domain.MovementOrderCompleted).AssetID)
	}
	if want := []string{"a", "b", "c", "d"}; !slices.Equal(order, want) {
		t.Errorf("tick order = %v, want %v", order, want)
	}
}

func driftRulebook(speed int) *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets:     map[string]*domain.AssetDefinition{"ship": {ID: "ship", Speed: speed}},
		DriftCosts: []int{6, 5, 4, 3, 2},
	}
}

func laneOrder(driftRating, banked int) *domain.Asset {
	path := []spatial.RegionHex{
		{RegionID: "a", Coord: spatial.HexCoord{Q: 0}},
		{RegionID: "b", Coord: spatial.HexCoord{Q: 5}},
		{RegionID: "b", Coord: spatial.HexCoord{Q: 6}},
	}
	return &domain.Asset{ID: "s1", OwnerID: "f1", DefinitionID: "ship",
		CurrentOrder: &domain.MovementOrder{AssetID: "s1", Path: path, DriftRating: driftRating, Banked: banked}}
}

func tickOne(t *testing.T, asset *domain.Asset, rb *rulebook.Rulebook) domain.Mutation {
	t.Helper()
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{asset.ID: asset}}
	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, rb, nil, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("TickMovementOrders: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want one", mutations)
	}
	return mutations[0]
}

func TestTickMovementOrders_DriftJumpCostsItsRatedPoints(t *testing.T) {
	first := tickOne(t, laneOrder(1, 0), driftRulebook(3))
	if got, ok := first.(domain.MovementOrderProgressed); !ok || got.NewStepIdx != 0 || got.Banked != 3 {
		t.Errorf("turn 1 = %+v, want no step and 3 points banked toward the 6-point jump", first)
	}

	second := tickOne(t, laneOrder(1, 3), driftRulebook(3))
	if got, ok := second.(domain.MovementOrderProgressed); !ok || got.NewStepIdx != 1 || got.Banked != 0 {
		t.Errorf("turn 2 = %+v, want the jump made with the banked points and nothing left", second)
	}

	rated := tickOne(t, laneOrder(5, 0), driftRulebook(3))
	if _, ok := rated.(domain.MovementOrderCompleted); !ok {
		t.Errorf("rating 5 = %+v, want the 2-point jump and the last step in one turn", rated)
	}
}

// hotFloor stands in for Aether Burners' hot speed_floor: an asset moving hot has speed 5 at least.
type hotFloor struct{}

func (hotFloor) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int, hot bool) int {
	if hot {
		return max(speed, 5)
	}
	return speed
}

func runHotRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"ship":   {ID: "ship", Speed: 2},
			"fast":   {ID: "fast", Speed: 10},
			"troops": {ID: "troops"},
		},
		DriftCosts: []int{5},
	}
}

func hotRegistry() *hooks.Registry {
	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.FactionScope("alpha"), "burners", hotFloor{})
	return registry
}

func TestBuildMovementMutations_RunHot(t *testing.T) {
	order := func(id, def string) *domain.Asset {
		return &domain.Asset{ID: id, DefinitionID: def, OwnerID: "alpha",
			Location:     domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 1}}},
			CurrentOrder: &domain.MovementOrder{AssetID: id, Path: makePath(5), StepIdx: 1}}
	}
	dest := &domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 3}}}
	cases := []struct {
		name     string
		def      string
		chosen   bool
		registry *hooks.Registry
		want     bool
	}{
		{name: "cold when not chosen", def: "ship", registry: hotRegistry(), want: false},
		{name: "hot when chosen", def: "ship", chosen: true, registry: hotRegistry(), want: true},
		{name: "forced when it moves only hot", def: "troops", registry: hotRegistry(), want: true},
		{name: "cold when hot adds no speed", def: "fast", chosen: true, registry: hotRegistry(), want: false},
		{name: "cold at speed 0 for a faction that cannot run hot", def: "troops", registry: nil, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "alpha", Coin: 5, Assets: map[string]*domain.Asset{"a1": order("a1", tc.def)}}
			decision := world.MovementDecision{AssetID: "a1", Kind: world.MovementDecisionRevise, Destination: dest, RunHot: tc.chosen}

			mutations, err := regionWorldEngine(t).BuildMovementMutations([]world.MovementDecision{decision}, faction, runHotRulebook(), tc.registry, slog.New(slog.DiscardHandler))

			if err != nil {
				t.Fatalf("BuildMovementMutations: %v", err)
			}
			revised := mutations[len(mutations)-1].(domain.MovementOrderRevised)
			if revised.NewOrder.RunHot != tc.want {
				t.Errorf("RunHot = %v, want %v", revised.NewOrder.RunHot, tc.want)
			}
		})
	}
}

func TestChoosesHot(t *testing.T) {
	rb := runHotRulebook()
	faction := &domain.Faction{ID: "alpha"}
	cases := []struct {
		def      string
		registry *hooks.Registry
		want     bool
	}{
		{def: "ship", registry: hotRegistry(), want: true},
		{def: "fast", registry: hotRegistry(), want: false},
		{def: "troops", registry: hotRegistry(), want: false},
		{def: "ship", registry: nil, want: false},
	}
	for _, tc := range cases {
		speedOf := dispatch.AssetSpeedResolver(tc.registry, faction)
		if got := world.ChoosesHot(speedOf, rb.Assets[tc.def]); got != tc.want {
			t.Errorf("ChoosesHot(%s, hooks %v) = %v, want %v", tc.def, tc.registry != nil, got, tc.want)
		}
	}
}

func TestBuildMovementMutations_IssueOverAnOpenOrderIsRefused(t *testing.T) {
	ship := &domain.Asset{ID: "s1", DefinitionID: "ship", OwnerID: "alpha",
		Location:     domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 1}}},
		CurrentOrder: &domain.MovementOrder{AssetID: "s1", Path: makePath(5), StepIdx: 1, Banked: 3}}
	faction := &domain.Faction{ID: "alpha", Coin: 5, Assets: map[string]*domain.Asset{"s1": ship}}
	decision := world.MovementDecision{AssetID: "s1", Kind: world.MovementDecisionIssue, Destination: &domain.Location{WorldID: "beta"}}

	_, err := regionWorldEngine(t).BuildMovementMutations([]world.MovementDecision{decision}, faction, runHotRulebook(), nil, slog.New(slog.DiscardHandler))

	if err == nil {
		t.Error("BuildMovementMutations = nil, want an issue over an open order refused")
	}
}

func TestBuildMovementMutations_ReviseKeepsBankedPoints(t *testing.T) {
	cases := []struct {
		name       string
		held       bool
		wantBanked int
	}{
		{name: "a moving order carries its points", held: false, wantBanked: 3},
		{name: "a held order drops its points", held: true, wantBanked: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"ship": {ID: "ship", Speed: 3}}, DriftCosts: []int{6}}
			ship := &domain.Asset{ID: "s1", DefinitionID: "ship", OwnerID: "alpha",
				Location:     domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 1}}},
				CurrentOrder: &domain.MovementOrder{AssetID: "s1", Path: makePath(5), StepIdx: 1, Banked: 3, Held: tc.held}}
			faction := &domain.Faction{ID: "alpha", Coin: 5, Assets: map[string]*domain.Asset{"s1": ship}}
			dest := &domain.Location{RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 3}}}

			mutations, err := regionWorldEngine(t).BuildMovementMutations(
				[]world.MovementDecision{{AssetID: "s1", Kind: world.MovementDecisionRevise, Destination: dest}},
				faction, rb, nil, slog.New(slog.DiscardHandler))

			if err != nil {
				t.Fatalf("BuildMovementMutations: %v", err)
			}
			revised := mutations[len(mutations)-1].(domain.MovementOrderRevised)
			if revised.NewOrder.Banked != tc.wantBanked {
				t.Errorf("Banked = %d, want %d", revised.NewOrder.Banked, tc.wantBanked)
			}
			if revised.NewOrder.Held {
				t.Error("Held = true, want the new order unheld")
			}
		})
	}
}
