package relocate

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

// relocRouter places worlds on a Q-line; hex distance = |ΔQ| ignoring region, and
// each world carries a RegionID so the hex and region enumerators are exercised
// independently.
type relocRouter struct{ worlds map[string]spatial.RegionHex }

func (r *relocRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &relocLoc{id: id, hex: hex}, true
}

func (r *relocRouter) Distance(from, to spatial.RegionHex, _ int) (int, error) {
	dist := from.Coord.Q - to.Coord.Q
	if dist < 0 {
		dist = -dist
	}
	return dist, nil
}

func (r *relocRouter) DistancesFrom(from spatial.RegionHex, crossingCost int) (map[spatial.RegionHex]int, error) {
	distances := make(map[spatial.RegionHex]int, len(r.worlds))
	for _, hex := range r.worlds {
		dist, err := r.Distance(from, hex, crossingCost)
		if err != nil {
			return nil, err
		}
		distances[hex] = dist
	}
	return distances, nil
}

func (r *relocRouter) Path(from, to spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return []spatial.RegionHex{from, to}, 1, nil
}

func (r *relocRouter) AllWorlds() []spatial.Location {
	worlds := make([]spatial.Location, 0, len(r.worlds))
	for id, hex := range r.worlds {
		worlds = append(worlds, &relocLoc{id: id, hex: hex})
	}
	return worlds
}

var _ world.HexRouter = (*relocRouter)(nil)

type relocLoc struct {
	id  string
	hex spatial.RegionHex
}

func (l *relocLoc) ID() string                   { return l.id }
func (l *relocLoc) Name() string                 { return l.id }
func (l *relocLoc) TechLevel() int               { return 5 }
func (l *relocLoc) Population() int              { return 0 }
func (l *relocLoc) Coords() (int, int)           { return l.hex.Coord.Q, l.hex.Coord.R }
func (l *relocLoc) RegionID() string             { return l.hex.RegionID }
func (l *relocLoc) RegionHex() spatial.RegionHex { return l.hex }

func relocHex(region string, q int) spatial.RegionHex {
	return spatial.RegionHex{RegionID: region, Coord: spatial.HexCoord{Q: q, R: 0}}
}

// relocEngine seeds region r1 with origin/near/mid/far (Q 0/1/2/5) and r2 with
// offregion (Q 10), parked far out so it never falls in hex range.
func relocEngine() *world.WorldEngine {
	return world.NewWithMap(&relocRouter{worlds: map[string]spatial.RegionHex{
		"origin":    relocHex("r1", 0),
		"near":      relocHex("r1", 1),
		"mid":       relocHex("r1", 2),
		"far":       relocHex("r1", 5),
		"offregion": relocHex("r2", 10),
	}}, slog.New(slog.DiscardHandler))
}

// relocRulebook types and categorises each definition ID for matchesFilters.
func relocRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"SF": {ID: "SF", Category: domain.StatForce, Type: domain.TypeSpecialForces},
		"MU": {ID: "MU", Category: domain.StatForce, Type: domain.TypeMilitaryUnit},
		"WF": {ID: "WF", Category: domain.StatWealth, Type: domain.TypeFacility},
		"CS": {ID: "CS", Category: domain.StatCunning, Type: domain.TypeStarship},
		"WS": {ID: "WS", Category: domain.StatWealth, Type: domain.TypeStarship},
	}}
}

// transitWebParams is Transit Web's relocate block: Cunning or Wealth, no
// Starships, 1 Coin for any number of moves.
func transitWebParams() *domain.RelocateParams {
	return &domain.RelocateParams{
		Scope:        domain.RelocateOther,
		Range:        domain.RangeHexes,
		MaxHex:       3,
		CoinCost:     1,
		CoinCostMode: domain.RelocateCostFlat,
		Categories:   []domain.FactionStat{domain.StatCunning, domain.StatWealth},
		ExcludeTypes: []domain.AssetType{domain.TypeStarship},
	}
}

func TestRelocate_OtherSingle_Hex(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sfInRange := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	wrongType := &domain.Asset{ID: "mu-near", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sfOutOfRange := &domain.Asset{ID: "sf-far", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far", RegionHex: relocHex("r1", 5)}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{
		"cs": shipping, "sf-near": sfInRange, "mu-near": wrongType, "sf-far": sfOutOfRange,
	}}
	params := &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sfInRange, DestinationWorld: "mid"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: shipping, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (AssetMoved + CoinDelta)", len(mutations))
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok || moved.AssetID != "sf-near" || moved.ToLocation.WorldID != "mid" {
		t.Fatalf("mutations[0] = %+v, want AssetMoved sf-near -> mid", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -1 {
		t.Errorf("mutations[1] = %+v, want CoinDelta -1", mutations[1])
	}
}

func TestRelocate_OtherSingle_InsufficientCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sfInRange := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 0, Assets: map[string]*domain.Asset{"cs": shipping, "sf-near": sfInRange}}
	params := &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}

	collector := mocks.NewMockCollector(ctrl) // no EXPECT: an unaffordable move is never prompted

	mutations, err := Run(Request{
		Faction: faction, Acting: shipping, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if !errors.Is(err, ErrInsufficientCoin) {
		t.Fatalf("err = %v, want ErrInsufficientCoin", err)
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil on affordability failure", mutations)
	}
}

func TestRelocate_ExceedsMaxAssets(t *testing.T) {
	ctrl := gomock.NewController(t)
	shipping := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sf1 := &domain.Asset{ID: "sf-1", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sf2 := &domain.Asset{ID: "sf-2", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "mid", RegionHex: relocHex("r1", 2)}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"cs": shipping, "sf-1": sf1, "sf-2": sf2}}
	params := &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1, AssetType: domain.TypeSpecialForces, MaxAssets: 1,
	}

	// Collector over-selects (two orders) against a MaxAssets=1 cap; the guard in Run rejects it.
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sf1, DestinationWorld: "mid"}, {Asset: sf2, DestinationWorld: "near"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: shipping, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if err == nil {
		t.Fatal("Run: expected MaxAssets-exceeded error, got nil")
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil when the cap is exceeded", mutations)
	}
}

func TestRelocate_OtherMulti_Region(t *testing.T) {
	ctrl := gomock.NewController(t)
	net := &domain.Asset{ID: "net", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sf1 := &domain.Asset{ID: "sf-1", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	sf2 := &domain.Asset{ID: "sf-2", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far", RegionHex: relocHex("r1", 5)}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"net": net, "sf-1": sf1, "sf-2": sf2}}
	params := &domain.RelocateParams{
		Scope: domain.RelocateOther, Range: domain.RangeRegion, AssetType: domain.TypeSpecialForces, MaxAssets: 0,
	}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: sf1, DestinationWorld: "mid"}, {Asset: sf2, DestinationWorld: "near"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: net, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (two AssetMoved, no CoinDelta)", len(mutations))
	}
	for i, m := range mutations {
		if _, ok := m.(domain.AssetMoved); !ok {
			t.Errorf("mutations[%d] type = %T, want AssetMoved", i, m)
		}
	}
}

// TestRun_NoCandidatesIsSentinel proves an empty candidate set wraps ErrNoCandidates
// and never reaches the prompt.
func TestRun_NoCandidatesIsSentinel(t *testing.T) {
	ctrl := gomock.NewController(t)
	acting := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	sfFar := &domain.Asset{ID: "sf-far", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far", RegionHex: relocHex("r1", 5)}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"cs": acting, "sf-far": sfFar}}
	params := &domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 1, AssetType: domain.TypeSpecialForces}

	collector := mocks.NewMockCollector(ctrl) // no EXPECT: any prompt fails the test

	mutations, err := Run(Request{
		Faction: faction, Acting: acting, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if !errors.Is(err, ErrNoCandidates) {
		t.Fatalf("err = %v, want ErrNoCandidates", err)
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil", mutations)
	}
}

// TestEligibleAssets_Filtering proves the three exclusions: wrong AssetType,
// a world outside the range-filtered candidate set (hex out-of-range or out-of-region
// both manifest as "not in candidateWorlds"), and the acting asset itself.
func TestEligibleAssets_Filtering(t *testing.T) {
	acting := &domain.Asset{ID: "cs", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "origin"}}
	sfNear := &domain.Asset{ID: "sf-near", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "near"}}
	muNear := &domain.Asset{ID: "mu-near", DefinitionID: "MU", OwnerID: "f1", Location: domain.Location{WorldID: "near"}}
	sfFar := &domain.Asset{ID: "sf-far", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "far"}}
	sfOff := &domain.Asset{ID: "sf-off", DefinitionID: "SF", OwnerID: "f1", Location: domain.Location{WorldID: "offregion"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{
		"cs": acting, "sf-near": sfNear, "mu-near": muNear, "sf-far": sfFar, "sf-off": sfOff,
	}}
	request := Request{
		Faction:  faction,
		Acting:   acting,
		Params:   &domain.RelocateParams{Scope: domain.RelocateOther, AssetType: domain.TypeSpecialForces},
		Rulebook: relocRulebook(),
	}

	eligible := eligibleAssets(request, []string{"mid", "near"})

	if len(eligible) != 1 {
		t.Fatalf("eligible = %v, want [sf-near] (wrong-type, out-of-range, off-region, and self all excluded)", assetIDs(eligible))
	}
	if eligible[0].ID != "sf-near" {
		t.Errorf("eligible[0].ID = %q, want sf-near", eligible[0].ID)
	}
}

func TestMatchesFilters(t *testing.T) {
	book := relocRulebook()
	cases := []struct {
		name   string
		def    *domain.AssetDefinition
		params *domain.RelocateParams
		want   bool
	}{
		{"no filter passes any def", book.Assets["SF"], &domain.RelocateParams{}, true},
		{"no filter passes an unknown def", nil, &domain.RelocateParams{}, true},
		{"type match", book.Assets["SF"], &domain.RelocateParams{AssetType: domain.TypeSpecialForces}, true},
		{"type mismatch", book.Assets["MU"], &domain.RelocateParams{AssetType: domain.TypeSpecialForces}, false},
		{"category in list", book.Assets["WF"], &domain.RelocateParams{Categories: []domain.FactionStat{domain.StatWealth}}, true},
		{"category out of list", book.Assets["SF"], &domain.RelocateParams{Categories: []domain.FactionStat{domain.StatWealth}}, false},
		{"excluded type", book.Assets["WS"], &domain.RelocateParams{ExcludeTypes: []domain.AssetType{domain.TypeStarship}}, false},
		{"transit web passes a wealth facility", book.Assets["WF"], transitWebParams(), true},
		{"transit web refuses a cunning starship", book.Assets["CS"], transitWebParams(), false},
		{"transit web refuses a wealth starship", book.Assets["WS"], transitWebParams(), false},
		{"transit web refuses a force asset", book.Assets["SF"], transitWebParams(), false},
		{"unknown def fails a type filter", nil, &domain.RelocateParams{AssetType: domain.TypeSpecialForces}, false},
		{"unknown def fails a category filter", nil, &domain.RelocateParams{Categories: []domain.FactionStat{domain.StatWealth}}, false},
		{"unknown def fails an excluded-type filter", nil, &domain.RelocateParams{ExcludeTypes: []domain.AssetType{domain.TypeStarship}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := matchesFilters(tc.def, tc.params); got != tc.want {
				t.Errorf("matchesFilters = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEligibleAssets_SkipsAssetUnderOrders(t *testing.T) {
	acting := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin"}}
	idle := &domain.Asset{ID: "wf-idle", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near"}}
	leaving := &domain.Asset{ID: "wf-leaving", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near"},
		CurrentOrder: &domain.MovementOrder{AssetID: "wf-leaving"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"web": acting, "wf-idle": idle, "wf-leaving": leaving}}
	request := Request{Faction: faction, Acting: acting, Params: transitWebParams(), Rulebook: relocRulebook()}

	eligible := eligibleAssets(request, []string{"origin", "near"})

	if len(eligible) != 1 || eligible[0].ID != "wf-idle" {
		t.Errorf("eligible = %v, want [wf-idle] (the asset under orders is skipped)", assetIDs(eligible))
	}
}

func TestRun_FlatCostChargesOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	wf1 := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	wf2 := &domain.Asset{ID: "wf-2", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "mid", RegionHex: relocHex("r1", 2)}}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"web": web, "wf-1": wf1, "wf-2": wf2}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: wf1, DestinationWorld: "mid"}, {Asset: wf2, DestinationWorld: "near"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3 (AssetMoved, CoinDelta, AssetMoved)", len(mutations))
	}
	if _, ok := mutations[0].(domain.AssetMoved); !ok {
		t.Errorf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
	if coin, ok := mutations[1].(domain.CoinDelta); !ok || coin.Delta != -1 {
		t.Errorf("mutations[1] = %+v, want CoinDelta -1", mutations[1])
	}
	if _, ok := mutations[2].(domain.AssetMoved); !ok {
		t.Errorf("mutations[2] type = %T, want AssetMoved (no second CoinDelta)", mutations[2])
	}
}

func TestRun_FlatCostInsufficient(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	wf1 := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 0, Assets: map[string]*domain.Asset{"web": web, "wf-1": wf1}}

	collector := mocks.NewMockCollector(ctrl) // no EXPECT: an unaffordable move is never prompted

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: transitWebParams(),
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if !errors.Is(err, ErrInsufficientCoin) {
		t.Fatalf("err = %v, want ErrInsufficientCoin", err)
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil on affordability failure", mutations)
	}
}

func TestRun_UnaffordableDoesNotPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	wf1 := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	faction := &domain.Faction{ID: "f1", Coin: 0, Assets: map[string]*domain.Asset{"web": web, "wf-1": wf1}}
	params := &domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1}

	collector := mocks.NewMockCollector(ctrl) // no EXPECT: any prompt fails the test

	_, err := Run(Request{
		Faction: faction, Acting: web, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if !errors.Is(err, ErrInsufficientCoin) {
		t.Fatalf("err = %v, want ErrInsufficientCoin", err)
	}
}

// TestRun_InsufficientAfterSelection covers the post-prompt check: one move is
// affordable, so the GM is asked, but the chosen per-asset total is not.
func TestRun_InsufficientAfterSelection(t *testing.T) {
	ctrl := gomock.NewController(t)
	web := &domain.Asset{ID: "web", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)}}
	wf1 := &domain.Asset{ID: "wf-1", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)}}
	wf2 := &domain.Asset{ID: "wf-2", DefinitionID: "WF", OwnerID: "f1", Location: domain.Location{WorldID: "mid", RegionHex: relocHex("r1", 2)}}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"web": web, "wf-1": wf1, "wf-2": wf2}}
	params := &domain.RelocateParams{Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3, CoinCost: 1}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: wf1, DestinationWorld: "mid"}, {Asset: wf2, DestinationWorld: "near"}}, nil)

	mutations, err := Run(Request{
		Faction: faction, Acting: web, Params: params,
		Collector: collector, World: relocEngine(), Rulebook: relocRulebook(),
	})
	if !errors.Is(err, ErrInsufficientCoin) {
		t.Fatalf("err = %v, want ErrInsufficientCoin", err)
	}
	if mutations != nil {
		t.Errorf("mutations = %v, want nil on affordability failure", mutations)
	}
}

func assetIDs(assets []*domain.Asset) []string {
	ids := make([]string, len(assets))
	for i, asset := range assets {
		ids[i] = asset.ID
	}
	return ids
}
