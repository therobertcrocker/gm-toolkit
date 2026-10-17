package engine

import (
	"errors"
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/relocate"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/turn"
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
	selectRelocateOrders func(*domain.Asset, []*domain.Asset, []string, int) ([]action.RelocateOrder, error)
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
func (c *minimalPhaseCollector) SelectRelocateOrders(acting *domain.Asset, candidates []*domain.Asset, candidateWorlds []string, maxAssets int) ([]action.RelocateOrder, error) {
	if c.selectRelocateOrders != nil {
		return c.selectRelocateOrders(acting, candidates, candidateWorlds, maxAssets)
	}
	return nil, nil
}

// stubWorldMap satisfies world.HexRouter with no-op implementations.
type stubWorldMap struct{}

func (s *stubWorldMap) Location(id string) (spatial.Location, bool)         { return nil, false }
func (s *stubWorldMap) Distance(_, _ spatial.RegionHex, _ int) (int, error) { return 1, nil }
func (s *stubWorldMap) DistancesFrom(_ spatial.RegionHex, _ int) (map[spatial.RegionHex]int, error) {
	return nil, nil
}
func (s *stubWorldMap) Path(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return nil, 0, nil
}
func (s *stubWorldMap) AllWorlds() []spatial.Location { return nil }

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

// errorObserver satisfies TurnObserver and records only the errors it is sent.
type errorObserver struct{ errs []error }

func (o *errorObserver) OnError(_ *domain.Faction, err error) {
	o.errs = append(o.errs, err)
}

func (o *errorObserver) OnFactionTurnStarted(_ *domain.Faction) {}

func (o *errorObserver) OnFactionSkipped(_ *domain.Faction) {}

func (o *errorObserver) OnGoalLockApplied(_ *domain.Faction, _ locks.GoalLock, _ []domain.Mutation) {}

func (o *errorObserver) OnBookkeepingApplied(_ *domain.Faction, _ turn.BookkeepingResult, _ []domain.Mutation) {
}

func (o *errorObserver) OnActionSelected(_ *domain.Faction, _ action.Action) {}

func (o *errorObserver) OnActionResolved(_ *domain.Faction, _ action.Action, _ []domain.Mutation) {}

func (o *errorObserver) OnFactionTurnCompleted(_ *domain.Faction) {}

func (o *errorObserver) OnCycleCompleted(_ int, _ *state.FactionState) {}

func (o *errorObserver) OnStatRaiseApplied(_ *domain.Faction, _ *domain.FactionStat, _ []domain.Mutation) {
}

func (o *errorObserver) OnStatRaiseSkipped(_ *domain.Faction) {}

func (o *errorObserver) OnMovementTicked(_ *domain.Faction, _ []domain.Mutation) {}

func (o *errorObserver) OnMovementResolved(_ *domain.Faction, _ []domain.Mutation) {}

func (o *errorObserver) OnIndexSkipped(_ []string) {}

// sameHexRouter puts every world on the zero hex, so each is at distance 0 from
// every other and in range of any relocate.
type sameHexRouter struct{ worlds []string }

func (r *sameHexRouter) Location(id string) (spatial.Location, bool) {
	if !slices.Contains(r.worlds, id) {
		return nil, false
	}
	return sameHexLoc(id), true
}
func (r *sameHexRouter) Distance(_, _ spatial.RegionHex, _ int) (int, error) { return 0, nil }
func (r *sameHexRouter) DistancesFrom(_ spatial.RegionHex, _ int) (map[spatial.RegionHex]int, error) {
	return map[spatial.RegionHex]int{{}: 0}, nil
}
func (r *sameHexRouter) Path(_, _ spatial.RegionHex, _ int) ([]spatial.RegionHex, int, error) {
	return nil, 0, nil
}
func (r *sameHexRouter) AllWorlds() []spatial.Location {
	locations := make([]spatial.Location, 0, len(r.worlds))
	for _, id := range r.worlds {
		locations = append(locations, sameHexLoc(id))
	}
	return locations
}

type sameHexLoc string

func (l sameHexLoc) ID() string                   { return string(l) }
func (l sameHexLoc) Name() string                 { return string(l) }
func (l sameHexLoc) TechLevel() int               { return 4 }
func (l sameHexLoc) Population() int              { return 0 }
func (l sameHexLoc) Coords() (int, int)           { return 0, 0 }
func (l sameHexLoc) RegionID() string             { return "" }
func (l sameHexLoc) RegionHex() spatial.RegionHex { return spatial.RegionHex{} }

// stubMovementGranter offers the same grants whatever faction asks.
type stubMovementGranter struct{ abilities []hooks.MovementAbility }

func (stub *stubMovementGranter) GrantMovementAbilities(_ *domain.Faction) []hooks.MovementAbility {
	return stub.abilities
}

// grantFixture builds alpha with a granting web and a facility on Krylos. The
// web offers one flat-cost, other-scope grant of 1 Coin.
func grantFixture(t *testing.T, coin int) (*Engine, *state.FactionState, *domain.Faction, *campaigns.Paths) {
	t.Helper()
	discard := slog.New(slog.DiscardHandler)
	web := &domain.Asset{ID: "web", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	facility := &domain.Asset{ID: "wf", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	alpha := &domain.Faction{ID: "alpha", Coin: coin, Assets: map[string]*domain.Asset{"web": web, "wf": facility}}
	registry := hooks.NewRegistry()
	registry.RegisterAssetMovementGranter(hooks.AssetScope("alpha", "web"), "test", &stubMovementGranter{abilities: []hooks.MovementAbility{{
		Source: web,
		Relocate: domain.RelocateParams{
			Scope: domain.RelocateOther, Range: domain.RangeHexes, MaxHex: 3,
			CoinCost: 1, CoinCostMode: domain.RelocateCostFlat,
		},
	}}})
	eng := &Engine{
		Rulebook: &rulebook.Rulebook{},
		Hooks:    registry,
		World:    world.NewWithMap(&sameHexRouter{worlds: []string{"Krylos", "Tartarus"}}, discard),
		Mutation: mutation.New(discard),
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	paths := &campaigns.Paths{HistoryPath: t.TempDir() + "/history.jsonl"}
	return eng, factionState, alpha, paths
}

func TestRunGrantedMovement_AppliesEachGrant(t *testing.T) {
	eng, factionState, alpha, paths := grantFixture(t, 5)
	collector := &minimalPhaseCollector{
		selectRelocateOrders: func(_ *domain.Asset, candidates []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
			return []action.RelocateOrder{{Asset: candidates[0], DestinationWorld: "Tartarus"}}, nil
		},
	}
	observer := &errorObserver{}

	resolved, err := eng.runGrantedMovement(alpha, factionState, paths, collector, observer, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("runGrantedMovement: %v", err)
	}
	if got := alpha.Assets["wf"].Location.WorldID; got != "Tartarus" {
		t.Errorf("wf on %q, want Tartarus", got)
	}
	if alpha.Coin != 4 {
		t.Errorf("Coin = %d, want 4", alpha.Coin)
	}
	if len(resolved) != 2 {
		t.Errorf("len(resolved) = %d, want 2 (AssetMoved + CoinDelta)", len(resolved))
	}
}

func TestRunGrantedMovement_DeclineSkips(t *testing.T) {
	cases := []struct {
		name  string
		reply error
	}{
		{"no orders", nil},
		{"esc", action.ErrTurnCanceled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, paths := grantFixture(t, 5)
			collector := &minimalPhaseCollector{
				selectRelocateOrders: func(_ *domain.Asset, _ []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
					return nil, tc.reply
				},
			}

			resolved, err := eng.runGrantedMovement(alpha, factionState, paths, collector, &errorObserver{}, slog.New(slog.DiscardHandler))
			if err != nil {
				t.Fatalf("runGrantedMovement: %v, want a skipped grant", err)
			}
			if len(resolved) != 0 || alpha.Coin != 5 {
				t.Errorf("resolved = %v, Coin = %d; want nothing applied", resolved, alpha.Coin)
			}
		})
	}
}

func TestRunGrantedMovement_UnaffordableSkipsWithoutPrompt(t *testing.T) {
	eng, factionState, alpha, paths := grantFixture(t, 0)
	prompted := false
	collector := &minimalPhaseCollector{
		selectRelocateOrders: func(_ *domain.Asset, _ []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
			prompted = true
			return nil, nil
		},
	}
	observer := &errorObserver{}

	resolved, err := eng.runGrantedMovement(alpha, factionState, paths, collector, observer, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("runGrantedMovement: %v, want a skipped grant", err)
	}
	if prompted {
		t.Error("collector prompted for a grant the faction cannot pay for")
	}
	if len(observer.errs) != 1 || !errors.Is(observer.errs[0], relocate.ErrInsufficientCoin) {
		t.Errorf("observer errors = %v, want one ErrInsufficientCoin", observer.errs)
	}
	if len(resolved) != 0 {
		t.Errorf("resolved = %v, want none", resolved)
	}
}

func TestRunGrantedMovement_CargoLoadedThisPhaseIsNotOffered(t *testing.T) {
	eng, factionState, alpha, paths := grantFixture(t, 5)
	alpha.Assets["cargo"] = &domain.Asset{ID: "cargo", OwnerID: "alpha"} // in transit: no WorldID
	var offered []string
	collector := &minimalPhaseCollector{
		selectRelocateOrders: func(_ *domain.Asset, candidates []*domain.Asset, _ []string, _ int) ([]action.RelocateOrder, error) {
			for _, candidate := range candidates {
				offered = append(offered, candidate.ID)
			}
			return nil, nil
		},
	}

	if _, err := eng.runGrantedMovement(alpha, factionState, paths, collector, &errorObserver{}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("runGrantedMovement: %v", err)
	}
	if !slices.Equal(offered, []string{"wf"}) {
		t.Errorf("offered = %v, want [wf] (cargo in transit is not a candidate)", offered)
	}
}

func TestFirePurchaseEffects_CausesWhitelisted(t *testing.T) {
	assassins := &domain.AssetDefinition{ID: "SWN-F5-003", Effects: []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialApplyStealth,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
	}}}
	eng := &Engine{
		Rulebook: &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{assassins.ID: assassins}},
		Effect:   effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)),
	}
	alpha := &domain.Faction{ID: "alpha"}
	added := func(assetID, cause string) domain.AssetAdded {
		return domain.AssetAdded{FactionID: "alpha", Asset: domain.Asset{ID: assetID, DefinitionID: assassins.ID}, Cause: cause, CausedByFactionID: "alpha"}
	}
	actionMutations := []domain.Mutation{
		added("bought", "buy"),
		added("enabled", "ability"),
		added("refitted", "refit"),
	}

	var stealthed []string
	for _, mutation := range eng.firePurchaseEffects(alpha, actionMutations, nil) {
		stealthed = append(stealthed, mutation.(domain.AssetStealthApplied).AssetID)
	}
	if want := []string{"bought", "enabled"}; !slices.Equal(stealthed, want) {
		t.Errorf("stealthed %v, want %v (a refit is not a purchase)", stealthed, want)
	}
}
