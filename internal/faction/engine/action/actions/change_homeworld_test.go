package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

// ---- test infrastructure -----------------------------------------------

var _ world.HexRouter = (*testHexRouter)(nil)

type testHexRouter struct {
	worlds     map[string]spatial.RegionHex
	distanceFn func(from, to spatial.RegionHex, crossingCost int) (int, error)
}

func (r *testHexRouter) Location(id string) (spatial.Location, bool) {
	hex, ok := r.worlds[id]
	if !ok {
		return nil, false
	}
	return &testWorldLoc{id: id, hex: hex}, true
}

func (r *testHexRouter) Distance(from, to spatial.RegionHex, crossingCost int) (int, error) {
	if r.distanceFn != nil {
		return r.distanceFn(from, to, crossingCost)
	}
	return 1, nil
}

func (r *testHexRouter) Path(from, to spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return []spatial.RegionHex{from, to}, 1, nil
}

var _ spatial.RegionLocation = (*testWorldLoc)(nil)

type testWorldLoc struct {
	id  string
	hex spatial.RegionHex
}

func (l *testWorldLoc) ID() string                   { return l.id }
func (l *testWorldLoc) Name() string                 { return l.id }
func (l *testWorldLoc) TechLevel() int               { return 5 }
func (l *testWorldLoc) Population() int              { return 0 }
func (l *testWorldLoc) Coords() (q, r int)           { return l.hex.Coord.Q, l.hex.Coord.R }
func (l *testWorldLoc) RegionID() string             { return l.hex.RegionID }
func (l *testWorldLoc) RegionHex() spatial.RegionHex { return l.hex }

// DriftCost(3) = DriftCosts[2] = 5
func testCHWRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{DriftCosts: []int{15, 10, 5}}
}

// ---- Task 6: cross-region distance test --------------------------------

// TestChangeHomeworld_CrossRegionDistance: homeworld in r1, target in r2.
// Distance = 7 (1 intra-r1 + 5 boundary + 1 intra-r2); TurnsRemaining = 1+7 = 8.
func TestChangeHomeworld_CrossRegionDistance(t *testing.T) {
	ctrl := gomock.NewController(t)

	aHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	cHex := spatial.RegionHex{RegionID: "r2", Coord: spatial.HexCoord{Q: 0, R: 0}}

	router := &testHexRouter{
		worlds: map[string]spatial.RegionHex{"a": aHex, "c": cHex},
		distanceFn: func(from, to spatial.RegionHex, _ int) (int, error) {
			if from == aHex && to == cHex {
				return 7, nil
			}
			return 1, nil
		},
	}
	eng := world.NewWithMap(router)
	rb := testCHWRulebook()

	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: domain.Location{WorldID: "a", RegionHex: aHex},
		Bases:     []*domain.Base{{ID: "b-c", OwnerID: "f1", Location: domain.Location{WorldID: "c"}}},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectChangeHomeworldTarget(gomock.Any(), gomock.Any()).Return("c", nil)

	act := NewChangeHomeworld(collector, eng, rb)
	if !act.Validate(faction, factionState, rb) {
		t.Fatal("Validate: expected true")
	}
	if err := act.Inputs(faction, factionState, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2", len(mutations))
	}
	initiated, ok := mutations[0].(domain.GoalInitiated)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want GoalInitiated", mutations[0])
	}
	if initiated.GoalID != "G-012" || initiated.TargetWorld.WorldID != "c" {
		t.Errorf("GoalInitiated = %+v, want GoalID=G-012 TargetWorld=c", initiated)
	}

	advanced, ok := mutations[1].(domain.GoalPhaseAdvanced)
	if !ok {
		t.Fatalf("mutations[1] type = %T, want GoalPhaseAdvanced", mutations[1])
	}
	const expectedDist = 7
	if advanced.TurnsRemaining != 1+expectedDist {
		t.Errorf("GoalPhaseAdvanced.TurnsRemaining = %d, want %d", advanced.TurnsRemaining, 1+expectedDist)
	}
}

// ---- Task 7: same-region distance test ---------------------------------

// TestChangeHomeworld_SameRegionDistance: homeworld and target both in r1.
// Distance = 3 (pure hex steps); TurnsRemaining = 1+3 = 4.
func TestChangeHomeworld_SameRegionDistance(t *testing.T) {
	ctrl := gomock.NewController(t)

	aHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	bHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 3, R: 0}}

	router := &testHexRouter{
		worlds: map[string]spatial.RegionHex{"a": aHex, "b": bHex},
		distanceFn: func(from, to spatial.RegionHex, _ int) (int, error) {
			if from == aHex && to == bHex {
				return 3, nil
			}
			return 1, nil
		},
	}
	eng := world.NewWithMap(router)
	rb := testCHWRulebook()

	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: domain.Location{WorldID: "a", RegionHex: aHex},
		Bases:     []*domain.Base{{ID: "b-b", OwnerID: "f1", Location: domain.Location{WorldID: "b"}}},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectChangeHomeworldTarget(gomock.Any(), gomock.Any()).Return("b", nil)

	act := NewChangeHomeworld(collector, eng, rb)
	if err := act.Inputs(faction, factionState, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	advanced, ok := mutations[1].(domain.GoalPhaseAdvanced)
	if !ok {
		t.Fatalf("mutations[1] type = %T, want GoalPhaseAdvanced", mutations[1])
	}
	const expectedDist = 3
	if advanced.TurnsRemaining != 1+expectedDist {
		t.Errorf("GoalPhaseAdvanced.TurnsRemaining = %d, want %d", advanced.TurnsRemaining, 1+expectedDist)
	}
}

// ---- Task 9: error path tests ------------------------------------------

// TestChangeHomeworld_ValidateFalseWhenNoEligibleTargets: faction has no base
// on any non-homeworld → eligible target set is empty → Validate returns false.
func TestChangeHomeworld_ValidateFalseWhenNoEligibleTargets(t *testing.T) {
	ctrl := gomock.NewController(t)

	aHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	router := &testHexRouter{worlds: map[string]spatial.RegionHex{"a": aHex}}
	eng := world.NewWithMap(router)
	rb := testCHWRulebook()

	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: domain.Location{WorldID: "a", RegionHex: aHex},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	act := NewChangeHomeworld(mocks.NewMockCollector(ctrl), eng, rb)
	if act.Validate(faction, factionState, rb) {
		t.Error("Validate: expected false when no eligible targets")
	}
}

// TestChangeHomeworld_NoPartialStateOnUnreachableTarget: faction has a base on
// reachable "b" (Validate = true), but collector injects unreachable "x". Resolve
// errors; no mutations are emitted; faction.ActiveGoal remains nil.
func TestChangeHomeworld_NoPartialStateOnUnreachableTarget(t *testing.T) {
	ctrl := gomock.NewController(t)

	aHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	bHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 1, R: 0}}
	xHex := spatial.RegionHex{RegionID: "r2", Coord: spatial.HexCoord{Q: 9, R: 9}}

	router := &testHexRouter{
		worlds: map[string]spatial.RegionHex{"a": aHex, "b": bHex, "x": xHex},
		distanceFn: func(_, to spatial.RegionHex, _ int) (int, error) {
			if to == xHex {
				return 0, spatial.ErrNoPath
			}
			return 1, nil
		},
	}
	eng := world.NewWithMap(router)
	rb := testCHWRulebook()

	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: domain.Location{WorldID: "a", RegionHex: aHex},
		Bases:     []*domain.Base{{ID: "b-b", OwnerID: "f1", Location: domain.Location{WorldID: "b"}}},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	// Validate passes because "b" is reachable.
	if !NewChangeHomeworld(mocks.NewMockCollector(ctrl), eng, rb).Validate(faction, factionState, rb) {
		t.Fatal("Validate: expected true (base on reachable world b)")
	}

	// Collector bypasses the eligibility filter and returns "x" (unreachable).
	injecting := mocks.NewMockCollector(ctrl)
	injecting.EXPECT().SelectChangeHomeworldTarget(gomock.Any(), gomock.Any()).Return("x", nil)

	act := NewChangeHomeworld(injecting, eng, rb)
	if err := act.Inputs(faction, factionState, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rb); err == nil {
		t.Error("Resolve: expected error for unreachable target, got nil")
	}
	// Output never called — no mutations applied — ActiveGoal stays nil.
	if faction.ActiveGoal != nil {
		t.Errorf("ActiveGoal: expected nil after Resolve failure, got %+v", faction.ActiveGoal)
	}
}

// ---- Task 10: base-of-influence precondition test ----------------------

// TestChangeHomeworldTargets_ExcludesWorldWithoutBase: worlds b and c are
// reachable and have bases; d is reachable but has no base. Expects [b c].
func TestChangeHomeworldTargets_ExcludesWorldWithoutBase(t *testing.T) {
	aHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	bHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 1, R: 0}}
	cHex := spatial.RegionHex{RegionID: "r2", Coord: spatial.HexCoord{Q: 0, R: 0}}
	dHex := spatial.RegionHex{RegionID: "r2", Coord: spatial.HexCoord{Q: 1, R: 0}}

	router := &testHexRouter{
		worlds: map[string]spatial.RegionHex{"a": aHex, "b": bHex, "c": cHex, "d": dHex},
	}
	eng := world.NewWithMap(router)
	rb := testCHWRulebook()

	faction := &domain.Faction{
		ID:        "f1",
		Homeworld: domain.Location{WorldID: "a", RegionHex: aHex},
		Bases: []*domain.Base{
			{ID: "b-a", OwnerID: "f1", Location: domain.Location{WorldID: "a"}}, // homeworld — excluded
			{ID: "b-b", OwnerID: "f1", Location: domain.Location{WorldID: "b"}},
			{ID: "b-c", OwnerID: "f1", Location: domain.Location{WorldID: "c"}},
			// no base on "d"
		},
	}

	targets := changeHomeworldTargets(faction, eng, rb)

	if len(targets) != 2 {
		t.Fatalf("changeHomeworldTargets = %v, want [b c]", targets)
	}
	if targets[0] != "b" || targets[1] != "c" {
		t.Errorf("changeHomeworldTargets = %v, want [b c]", targets)
	}
}
