package dispatch

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type stubPermissionCollector struct {
	grant   bool
	contest bool
	asked   []hooks.PermissionRequest
	offered []hooks.PermissionContestOffer
}

func (stub *stubPermissionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	stub.asked = append(stub.asked, request)
	return stub.grant, nil
}

func (stub *stubPermissionCollector) ConfirmPermissionContest(offer hooks.PermissionContestOffer) (bool, error) {
	stub.offered = append(stub.offered, offer)
	return stub.contest, nil
}

func permissionState() (*state.FactionState, *domain.Faction, *domain.Faction) {
	requester := &domain.Faction{ID: "f1", Name: "Alpha", Homeworld: domain.Location{WorldID: "Anchorage"}}
	ruler := &domain.Faction{ID: "f2", Name: "Bravo"}
	factionState := &state.FactionState{
		Factions:    map[string]*domain.Faction{"f1": requester, "f2": ruler},
		Governments: map[string]string{"Tartarus": "f1", "Krylos": "f2", "Anchorage": "f2"},
	}
	return factionState, requester, ruler
}

func TestRequestPermission(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	plainDef := &domain.AssetDefinition{ID: "militia"}
	factionState, requester, ruler := permissionState()

	cases := []struct {
		name        string
		def         *domain.AssetDefinition
		world       string
		grant       bool
		wantAsked   []hooks.PermissionRequest
		wantRefusal *domain.PermissionRefused
	}{
		{name: "no permission needed", def: plainDef, world: "Krylos"},
		{name: "the requester rules the world", def: permissionDef, world: "Tartarus"},
		{name: "the requester's homeworld, a rival its ruler", def: permissionDef, world: "Anchorage"},
		{
			name:      "a rival ruler grants",
			def:       permissionDef,
			world:     "Krylos",
			grant:     true,
			wantAsked: []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef, Cause: "buy"}},
		},
		{
			name:        "a rival ruler refuses",
			def:         permissionDef,
			world:       "Krylos",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef, Cause: "buy"}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1"},
		},
		{
			name:        "the GM refuses for an independent world",
			def:         permissionDef,
			world:       "Hadrian",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, WorldID: "Hadrian", Def: permissionDef, Cause: "buy"}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Hadrian", DefinitionID: "gunships", Cause: "buy", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			collector := &stubPermissionCollector{grant: tc.grant}

			refusal, err := RequestPermission(nil, collector, factionState, requester, tc.def, tc.world, "buy")

			if err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			if !reflect.DeepEqual(collector.asked, tc.wantAsked) {
				t.Errorf("asked = %+v, want %+v", collector.asked, tc.wantAsked)
			}
			if !reflect.DeepEqual(refusal, tc.wantRefusal) {
				t.Errorf("refusal = %+v, want %+v", refusal, tc.wantRefusal)
			}
		})
	}
}

type stubGranter struct{ grant bool }

func (stub *stubGranter) GrantsPermission(hooks.PermissionQuery) bool { return stub.grant }

func TestRequestPermission_AGrantHookIsNotAsked(t *testing.T) {
	factionState, requester, _ := permissionState()
	requester.Assets = map[string]*domain.Asset{"pm": {ID: "pm"}}
	registry := hooks.NewRegistry()
	registry.RegisterPermissionGranter(hooks.AssetScope("f1", "pm"), "pm", &stubGranter{grant: true})
	collector := &stubPermissionCollector{}
	def := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}

	refusal, err := RequestPermission(registry, collector, factionState, requester, def, "Krylos", "buy")

	if err != nil || refusal != nil {
		t.Errorf("RequestPermission = (%+v, %v), want a grant", refusal, err)
	}
	if len(collector.asked) != 0 {
		t.Errorf("asked = %+v, want no prompt when a grant hook serves the request", collector.asked)
	}
}

func TestPermissionGranted_AsksOnlyTheRequestersHooks(t *testing.T) {
	_, requester, ruler := permissionState()
	ruler.Assets = map[string]*domain.Asset{"pm": {ID: "pm"}}
	registry := hooks.NewRegistry()
	registry.RegisterPermissionGranter(hooks.AssetScope("f2", "pm"), "pm", &stubGranter{grant: true})

	if PermissionGranted(registry, hooks.PermissionQuery{Requester: requester, WorldID: "Krylos"}) {
		t.Error("a rival's grant hook granted the requester, want a grant to serve its owner only")
	}
	if PermissionGranted(nil, hooks.PermissionQuery{Requester: requester}) {
		t.Error("a nil registry granted, want false")
	}
}

func TestRequestArrivalPermission(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	raiseOnlyDef := &domain.AssetDefinition{ID: "fleet", PurchasePermission: true}
	factionState, requester, ruler := permissionState()

	cases := []struct {
		name        string
		def         *domain.AssetDefinition
		asset       *domain.Asset
		world       string
		wantAsked   []hooks.PermissionRequest
		wantRefusal *domain.PermissionRefused
	}{
		{name: "raise-only permission", def: raiseOnlyDef, asset: &domain.Asset{}, world: "Krylos"},
		{name: "a stealthed P asset", def: permissionDef, asset: &domain.Asset{Stealthy: true}, world: "Krylos"},
		{name: "the requester rules the world", def: permissionDef, asset: &domain.Asset{}, world: "Tartarus"},
		{
			name:        "the homeworld still asks",
			def:         permissionDef,
			asset:       &domain.Asset{},
			world:       "Anchorage",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Anchorage", Def: permissionDef, Cause: "arrival"}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Anchorage", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "arrival", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			collector := &stubPermissionCollector{}

			refusal, err := RequestArrivalPermission(nil, collector, factionState, requester, tc.asset, tc.def, tc.world, domain.PermissionCauseArrival)

			if err != nil {
				t.Fatalf("RequestArrivalPermission: %v", err)
			}
			if !reflect.DeepEqual(collector.asked, tc.wantAsked) {
				t.Errorf("asked = %+v, want %+v", collector.asked, tc.wantAsked)
			}
			if !reflect.DeepEqual(refusal, tc.wantRefusal) {
				t.Errorf("refusal = %+v, want %+v", refusal, tc.wantRefusal)
			}
		})
	}
}

type stubContester struct {
	source *domain.Asset
	wins   bool
	rolled int
}

func (stub *stubContester) OfferContest(hooks.PermissionQuery) *hooks.PermissionContest {
	if stub.source == nil {
		return nil
	}
	return &hooks.PermissionContest{Source: stub.source}
}

func (stub *stubContester) WinsContest(hooks.PermissionQuery) bool {
	stub.rolled++
	return stub.wins
}

// lobbyistsOf registers a contester on a new asset of faction and returns it.
func lobbyistsOf(registry *hooks.Registry, faction *domain.Faction, assetID string, wins bool) *stubContester {
	if faction.Assets == nil {
		faction.Assets = map[string]*domain.Asset{}
	}
	asset := &domain.Asset{ID: assetID, OwnerID: faction.ID}
	faction.Assets[assetID] = asset
	contester := &stubContester{source: asset, wins: wins}
	registry.RegisterPermissionContester(hooks.AssetScope(faction.ID, assetID), assetID, contester)
	return contester
}

func TestRequestPermission_LobbyistsContestAGrant(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	cases := []struct {
		name        string
		contest     bool
		wins        bool
		wantRolled  int
		wantRefusal *domain.PermissionRefused
	}{
		{name: "the owner lets it stand"},
		{name: "the owner contests and loses", contest: true, wantRolled: 1},
		{
			name: "the owner contests and wins", contest: true, wins: true, wantRolled: 1,
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Tartarus", DefinitionID: "gunships", DecidedByFactionID: "f2", Reason: domain.RefusalWithdrawn, Cause: "buy", CausedByFactionID: "f1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, requester, ruler := permissionState()
			registry := hooks.NewRegistry()
			lobbyists := lobbyistsOf(registry, ruler, "lob", tc.wins)
			collector := &stubPermissionCollector{contest: tc.contest}

			// Tartarus is the requester's own world: the grant is automatic (D-6), and still contested (D-29).
			refusal, err := RequestPermission(registry, collector, factionState, requester, permissionDef, "Tartarus", "buy")

			if err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			wantOffer := []hooks.PermissionContestOffer{{Owner: ruler, Source: lobbyists.source, Requester: requester, WorldID: "Tartarus", Def: permissionDef, Cause: "buy"}}
			if !reflect.DeepEqual(collector.offered, wantOffer) {
				t.Errorf("offered = %+v, want %+v", collector.offered, wantOffer)
			}
			if lobbyists.rolled != tc.wantRolled {
				t.Errorf("rolled %d times, want %d", lobbyists.rolled, tc.wantRolled)
			}
			if !reflect.DeepEqual(refusal, tc.wantRefusal) {
				t.Errorf("refusal = %+v, want %+v", refusal, tc.wantRefusal)
			}
			if got := requester.BudgetSpent(PermissionWithdrawnKey("Tartarus", "f2")); got != (tc.wantRefusal != nil) {
				t.Errorf("withdrawal mark spent = %v, want %v", got, tc.wantRefusal != nil)
			}
		})
	}
}

func TestRequestPermission_ContestWalkStopsAtTheFirstWin(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	factionState, requester, ruler := permissionState()
	registry := hooks.NewRegistry()
	own := lobbyistsOf(registry, requester, "own", true)
	second := lobbyistsOf(registry, ruler, "lob-b", true)
	first := lobbyistsOf(registry, ruler, "lob-a", false)
	collector := &stubPermissionCollector{contest: true}

	refusal, err := RequestPermission(registry, collector, factionState, requester, permissionDef, "Tartarus", "buy")

	if err != nil {
		t.Fatalf("RequestPermission: %v", err)
	}
	if refusal == nil || refusal.Reason != domain.RefusalWithdrawn {
		t.Fatalf("refusal = %+v, want a withdrawal", refusal)
	}
	var sources []string
	for _, offer := range collector.offered {
		sources = append(sources, offer.Source.ID)
	}
	if want := []string{"lob-a", "lob-b"}; !reflect.DeepEqual(sources, want) {
		t.Errorf("offered %v, want %v (asset ID order, the requester's own never)", sources, want)
	}
	if own.rolled != 0 || first.rolled != 1 || second.rolled != 1 {
		t.Errorf("rolled own=%d first=%d second=%d, want 0, 1, 1", own.rolled, first.rolled, second.rolled)
	}
}

func TestRequestPermission_NoContestWithoutAGrant(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	plainDef := &domain.AssetDefinition{ID: "militia"}
	cases := []struct {
		name  string
		def   *domain.AssetDefinition
		world string
	}{
		{"a purchase on the homeworld", permissionDef, "Anchorage"},
		{"no permission needed", plainDef, "Krylos"},
		{"the government refuses", permissionDef, "Krylos"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, requester, ruler := permissionState()
			registry := hooks.NewRegistry()
			lobbyistsOf(registry, ruler, "lob", true)
			collector := &stubPermissionCollector{contest: true}

			if _, err := RequestPermission(registry, collector, factionState, requester, tc.def, tc.world, "buy"); err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			if len(collector.offered) != 0 {
				t.Errorf("offered = %+v, want no contest", collector.offered)
			}
		})
	}
}

func TestRequestPermission_TheRulerDoesNotContestItsOwnYes(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	cases := []struct {
		name        string
		grantHook   bool
		wantOffered int
	}{
		{name: "the ruler says yes", wantOffered: 0},
		{name: "a grant hook goes around the ruler", grantHook: true, wantOffered: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState, requester, ruler := permissionState()
			registry := hooks.NewRegistry()
			lobbyistsOf(registry, ruler, "lob", true)
			if tc.grantHook {
				requester.Assets = map[string]*domain.Asset{"pm": {ID: "pm"}}
				registry.RegisterPermissionGranter(hooks.AssetScope("f1", "pm"), "pm", &stubGranter{grant: true})
			}
			collector := &stubPermissionCollector{grant: true, contest: true}

			// Krylos is the ruler's (f2) world.
			if _, err := RequestPermission(registry, collector, factionState, requester, permissionDef, "Krylos", "buy"); err != nil {
				t.Fatalf("RequestPermission: %v", err)
			}
			if len(collector.offered) != tc.wantOffered {
				t.Errorf("offered %d contests, want %d", len(collector.offered), tc.wantOffered)
			}
		})
	}
}

func TestRequestPermission_AWithdrawalBarsTheWorldThisTurn(t *testing.T) {
	permissionDef := &domain.AssetDefinition{ID: "gunships", Flags: []domain.AssetFlag{domain.FlagPermission}}
	factionState, requester, _ := permissionState()
	requester.SpendBudget(PermissionWithdrawnKey("Krylos", "f2"))
	collector := &stubPermissionCollector{grant: true}

	refusal, err := RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Krylos", domain.PermissionCauseArrival)

	if err != nil {
		t.Fatalf("RequestArrivalPermission: %v", err)
	}
	want := &domain.PermissionRefused{FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Reason: domain.RefusalBarred, Cause: "arrival", CausedByFactionID: "f1"}
	if !reflect.DeepEqual(refusal, want) {
		t.Errorf("refusal = %+v, want %+v", refusal, want)
	}
	if len(collector.asked) != 0 {
		t.Errorf("asked = %+v, want no ask while barred", collector.asked)
	}
	if refusal, _ := RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Hadrian", domain.PermissionCauseArrival); refusal != nil || len(collector.asked) != 1 {
		t.Errorf("Hadrian: refusal = %+v, asked %d; want another world asked as usual", refusal, len(collector.asked))
	}
}
