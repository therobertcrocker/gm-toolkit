package engine

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// stubAction implements action.Action with a fixed name and no-op methods.
type stubAction struct{ name string }

func (s *stubAction) Name() string { return s.name }
func (s *stubAction) Validate(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool {
	return true
}
func (s *stubAction) Inputs(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	return nil
}
func (s *stubAction) Resolve(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error {
	return nil
}
func (s *stubAction) Output() ([]domain.Mutation, error) { return nil, nil }

// minimalPhaseCollector satisfies PhaseCollector with configurable fns and
// safe no-op defaults. Used in unit tests that exercise orchestrator helpers
// without running the full turn pipeline.
type minimalPhaseCollector struct {
	movementDecisionsFn  func(*domain.Faction, []*domain.Asset) ([]world.MovementDecision, error)
	selectTransportCargo func(*domain.Asset, []*domain.Asset, *domain.TransportProfile) ([]*domain.Asset, error)
}

func (c *minimalPhaseCollector) AwaitCheckpoint(_ string) error { return nil }
func (c *minimalPhaseCollector) SelectAction(_ *domain.Faction, _ []action.Action) (action.Action, error) {
	return nil, nil
}
func (c *minimalPhaseCollector) SelectStatRaise(_ *domain.Faction, _ []domain.FactionStat) (*domain.FactionStat, error) {
	return nil, nil
}
func (c *minimalPhaseCollector) SelectMovementDecisions(faction *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
	if c.movementDecisionsFn != nil {
		return c.movementDecisionsFn(faction, eligible)
	}
	return nil, nil
}
func (c *minimalPhaseCollector) SelectTransportCargo(transport *domain.Asset, eligibleCargo []*domain.Asset, profile *domain.TransportProfile) ([]*domain.Asset, error) {
	if c.selectTransportCargo != nil {
		return c.selectTransportCargo(transport, eligibleCargo, profile)
	}
	return nil, nil
}

// stubWorldMap satisfies world.HexRouter with no-op implementations.
type stubWorldMap struct{}

func (s *stubWorldMap) Location(id string) (spatial.Location, bool)         { return nil, false }
func (s *stubWorldMap) Distance(_, _ spatial.RegionHex, _ int) (int, error) { return 1, nil }
func (s *stubWorldMap) Path(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return nil, 0, nil
}

func transportRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"transport-def": {
				ID:    "transport-def",
				Speed: 2,
				Transport: &domain.TransportProfile{
					MaxHex:     4,
					CoinCost:   1,
					CargoTypes: []domain.AssetType{domain.TypeSpecialForces},
					MaxCargo:   1,
				},
			},
			"sf-def": {ID: "sf-def", Type: domain.TypeSpecialForces},
			"mu-def": {ID: "mu-def", Type: domain.TypeMilitaryUnit},
		},
		DriftCosts: []int{5},
	}
}

func TestEligibleCargoForTransport(t *testing.T) {
	loc := domain.Location{
		WorldID:   "world-a",
		RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 0, R: 0}},
	}
	otherLoc := domain.Location{
		WorldID:   "world-a",
		RegionHex: spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: 2, R: 0}},
	}
	rb := transportRulebook()
	profile := rb.Assets["transport-def"].Transport

	transport := &domain.Asset{
		ID:           "transport-1",
		DefinitionID: "transport-def",
		OwnerID:      "alpha",
		Location:     loc,
	}

	cases := []struct {
		name   string
		asset  *domain.Asset
		wantIn bool
	}{
		{
			name:   "co-located allowed type",
			asset:  &domain.Asset{ID: "sf-1", DefinitionID: "sf-def", OwnerID: "alpha", Location: loc},
			wantIn: true,
		},
		{
			name:   "disallowed asset type",
			asset:  &domain.Asset{ID: "mu-1", DefinitionID: "mu-def", OwnerID: "alpha", Location: loc},
			wantIn: false,
		},
		{
			name:   "co-located but different world",
			asset:  &domain.Asset{ID: "sf-2", DefinitionID: "sf-def", OwnerID: "alpha", Location: domain.Location{WorldID: "world-b", RegionHex: loc.RegionHex}},
			wantIn: false,
		},
		{
			name:   "same world but different hex",
			asset:  &domain.Asset{ID: "sf-3", DefinitionID: "sf-def", OwnerID: "alpha", Location: otherLoc},
			wantIn: false,
		},
		{
			name:   "already in-flight cargo",
			asset:  &domain.Asset{ID: "sf-4", DefinitionID: "sf-def", OwnerID: "alpha", Location: loc, CurrentOrder: &domain.MovementOrder{AssetID: "sf-4"}},
			wantIn: false,
		},
		{
			name:   "transport itself excluded",
			asset:  transport,
			wantIn: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{
				ID: "alpha",
				Assets: map[string]*domain.Asset{
					transport.ID: transport,
					tc.asset.ID:  tc.asset,
				},
			}
			eligible := eligibleCargoForTransport(faction, transport, profile, rb)
			found := false
			for _, a := range eligible {
				if a.ID == tc.asset.ID {
					found = true
					break
				}
			}
			if found != tc.wantIn {
				t.Errorf("asset %q: in eligible set = %v, want %v", tc.asset.ID, found, tc.wantIn)
			}
		})
	}
}

func TestPrepareMovementDecisions_MaxCargoViolation(t *testing.T) {
	rb := transportRulebook()

	transportLoc := domain.Location{WorldID: "world-a"}
	faction := &domain.Faction{
		ID: "alpha",
		Assets: map[string]*domain.Asset{
			"transport-1": {
				ID:           "transport-1",
				DefinitionID: "transport-def",
				OwnerID:      "alpha",
				Location:     transportLoc,
			},
		},
	}

	tooManyCargo := []*domain.Asset{
		{ID: "sf-1", DefinitionID: "sf-def"},
		{ID: "sf-2", DefinitionID: "sf-def"},
	}

	collector := &minimalPhaseCollector{
		movementDecisionsFn: func(_ *domain.Faction, eligible []*domain.Asset) ([]world.MovementDecision, error) {
			return []world.MovementDecision{{
				AssetID:     "transport-1",
				Kind:        world.MovementDecisionIssue,
				Destination: &domain.Location{WorldID: "world-c"},
			}}, nil
		},
		selectTransportCargo: func(_ *domain.Asset, _ []*domain.Asset, _ *domain.TransportProfile) ([]*domain.Asset, error) {
			return tooManyCargo, nil
		},
	}

	we := world.NewWithMap(&stubWorldMap{}, slog.New(slog.DiscardHandler))
	_, err := prepareMovementDecisions(faction, collector, we, rb, slog.New(slog.DiscardHandler))
	if err == nil {
		t.Fatal("expected error when collector returns more cargo than MaxCargo allows, got nil")
	}
}

func TestFilterAllowedActions(t *testing.T) {
	available := []action.Action{
		&stubAction{"Attack"},
		&stubAction{"Sell Asset"},
		&stubAction{"Expand Influence"},
	}

	tests := []struct {
		name    string
		allowed []string
		want    []string
	}{
		{"nil allowed returns nothing", nil, nil},
		{"empty allowed returns nothing", []string{}, nil},
		{"no-match returns nothing", []string{"Buy Asset"}, nil},
		{"single match", []string{"Attack"}, []string{"Attack"}},
		{"multiple matches", []string{"Attack", "Expand Influence"}, []string{"Attack", "Expand Influence"}},
		{"preserves available order not allowed order", []string{"Expand Influence", "Sell Asset"}, []string{"Sell Asset", "Expand Influence"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterAllowedActions(available, tc.allowed)
			var gotNames []string
			for _, a := range got {
				gotNames = append(gotNames, a.Name())
			}
			if !reflect.DeepEqual(gotNames, tc.want) {
				t.Errorf("got %v, want %v", gotNames, tc.want)
			}
		})
	}
}
