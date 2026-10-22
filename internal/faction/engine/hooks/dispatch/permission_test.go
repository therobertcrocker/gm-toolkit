package dispatch

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type stubPermissionCollector struct {
	grant bool
	asked []hooks.PermissionRequest
}

func (stub *stubPermissionCollector) RequestPermission(request hooks.PermissionRequest) (bool, error) {
	stub.asked = append(stub.asked, request)
	return stub.grant, nil
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
			wantAsked: []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
		},
		{
			name:        "a rival ruler refuses",
			def:         permissionDef,
			world:       "Krylos",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, Ruler: ruler, WorldID: "Krylos", Def: permissionDef}},
			wantRefusal: &domain.PermissionRefused{FactionID: "f1", WorldID: "Krylos", DefinitionID: "gunships", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1"},
		},
		{
			name:        "the GM refuses for an independent world",
			def:         permissionDef,
			world:       "Hadrian",
			wantAsked:   []hooks.PermissionRequest{{Requester: requester, WorldID: "Hadrian", Def: permissionDef}},
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
