package engine

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func arrivalHex(q int) spatial.RegionHex {
	return spatial.RegionHex{RegionID: "void", Coord: spatial.HexCoord{Q: q}}
}

func arrivalRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"ship": {ID: "ship", Name: "Ship", Speed: 2, Transport: &domain.TransportProfile{
			MaxCargo: 1, CargoTypes: []domain.AssetType{domain.TypeMilitaryUnit},
		}},
		"troops":  {ID: "troops", Name: "Troops", Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
		"militia": {ID: "militia", Name: "Militia", Type: domain.TypeMilitaryUnit},
		"mercs":   {ID: "mercs", Name: "Mercs", Speed: 2, Type: domain.TypeMilitaryUnit, Flags: []domain.AssetFlag{domain.FlagPermission}},
		"landers": {ID: "landers", Name: "Landers", Speed: 3, Transport: &domain.TransportProfile{
			MaxCargo: 1, CargoTypes: []domain.AssetType{domain.TypeMilitaryUnit}, IgnoresLockdown: true,
		}},
		"lockdown": {ID: "lockdown", Name: "Lockdown", Effects: []domain.EffectDefinition{{
			Effect: domain.SpecialLockOut, LockOut: &domain.LockOutParams{Fee: domain.DiceRoll{NumDice: 1, Sides: 4}},
		}}},
	}}
}

// arrivalFixture builds alpha with a carrier "t1" (definition carrierDef) whose
// order completes onto Krylos along pathLen hexes, carrying "c1" (definition
// cargoDef) when cargoDef is set. Bravo governs Krylos.
func arrivalFixture(carrierDef, cargoDef string, pathLen int) (*Engine, *state.FactionState, *domain.Faction, domain.MovementOrderCompleted) {
	path := make([]spatial.RegionHex, pathLen)
	for i := range path {
		path[i] = arrivalHex(i)
	}
	destination := domain.Location{WorldID: "Krylos", RegionHex: path[pathLen-1]}
	carrier := &domain.Asset{ID: "t1", DefinitionID: carrierDef, OwnerID: "alpha", Location: domain.Location{RegionHex: path[pathLen-2]},
		CurrentOrder: &domain.MovementOrder{AssetID: "t1", Destination: destination, Path: path, StepIdx: pathLen - 2}}
	alpha := &domain.Faction{ID: "alpha", Name: "Alpha", Coin: 5, Homeworld: domain.Location{WorldID: "Anchorage"}, Assets: map[string]*domain.Asset{"t1": carrier}}
	if cargoDef != "" {
		alpha.Assets["c1"] = &domain.Asset{ID: "c1", DefinitionID: cargoDef, OwnerID: "alpha", Location: carrier.Location}
		carrier.CurrentOrder.CargoAssetIDs = []string{"c1"}
	}
	bravo := &domain.Faction{ID: "bravo", Name: "Bravo"}
	factionState := &state.FactionState{
		Factions:    map[string]*domain.Faction{"alpha": alpha, "bravo": bravo},
		Governments: map[string]string{"Krylos": "bravo"},
	}
	discard := slog.New(slog.DiscardHandler)
	eng := &Engine{Rulebook: arrivalRulebook(), Hooks: hooks.NewRegistry(), Mutation: mutation.New(discard)}
	completed := domain.MovementOrderCompleted{FactionID: "alpha", AssetID: "t1", FinalLocation: destination, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"}
	return eng, factionState, alpha, completed
}

func TestCheckArrivals(t *testing.T) {
	cases := []struct {
		name       string
		carrierDef string
		cargoDef   string
		stealthy   bool
		rulesSelf  bool
		grant      bool
		wantAsked  []string
		wantHeld   bool
	}{
		{name: "nothing aboard needs permission", carrierDef: "ship", cargoDef: "militia"},
		{name: "a P cargo is granted", carrierDef: "ship", cargoDef: "troops", grant: true, wantAsked: []string{"troops"}},
		{name: "a P cargo is refused", carrierDef: "ship", cargoDef: "troops", wantAsked: []string{"troops"}, wantHeld: true},
		{name: "a P carrier is refused", carrierDef: "mercs", wantAsked: []string{"mercs"}, wantHeld: true},
		{name: "a stealthed P cargo is not asked", carrierDef: "ship", cargoDef: "troops", stealthy: true},
		{name: "the faction rules the world", carrierDef: "ship", cargoDef: "troops", rulesSelf: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture(tc.carrierDef, tc.cargoDef, 3)
			if cargo := alpha.Assets["c1"]; cargo != nil {
				cargo.Stealthy = tc.stealthy
			}
			if tc.rulesSelf {
				factionState.Governments["Krylos"] = "alpha"
			}
			var asked []string
			collector := &minimalPhaseCollector{requestPermission: func(request hooks.PermissionRequest) (bool, error) {
				asked = append(asked, request.Def.ID)
				if request.Cause != domain.PermissionCauseArrival || request.Ruler == nil || request.Ruler.ID != "bravo" {
					t.Errorf("request = %+v, want an arrival asked of bravo", request)
				}
				return tc.grant, nil
			}}

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if !reflect.DeepEqual(asked, tc.wantAsked) {
				t.Errorf("asked = %v, want %v", asked, tc.wantAsked)
			}
			if !tc.wantHeld {
				if !reflect.DeepEqual(got, []domain.Mutation{completed}) {
					t.Errorf("mutations = %+v, want the completion unchanged", got)
				}
				return
			}
			refused := tc.wantAsked[0]
			want := []domain.Mutation{
				domain.PermissionRefused{FactionID: "alpha", WorldID: "Krylos", DefinitionID: refused, DecidedByFactionID: "bravo", Cause: domain.PermissionCauseArrival, CausedByFactionID: "alpha"},
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldRefused, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_FirstRefusalEndsTheAsking(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "troops", 3)
	asks := 0
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) {
		asks++
		return false, nil
	}}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	if asks != 1 {
		t.Errorf("asked %d times, want 1 (one refusal holds the whole order, D-28)", asks)
	}
	if len(got) != 2 {
		t.Errorf("mutations = %+v, want the refusal and the hold", got)
	}
}

func TestCheckArrivals_TwoHexPathHoldsAtTheStart(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "", 2)
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) { return false, nil }}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	held, ok := got[len(got)-1].(domain.MovementOrderHeld)
	if !ok || held.NewStepIdx != 0 || held.RegionHex != arrivalHex(0) {
		t.Errorf("last mutation = %+v, want a hold at step 0 on the path's first hex", got[len(got)-1])
	}
}

func TestCheckArrivals_ABypassingCarrierAsksNothing(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("ship", "troops", 3)
	eng.Rulebook.Assets["ship"].Transport.IgnoresPermission = true
	collector := &minimalPhaseCollector{requestPermission: func(request hooks.PermissionRequest) (bool, error) {
		t.Errorf("asked %+v, want no ask past a bypassing carrier", request)
		return false, nil
	}}

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	if !reflect.DeepEqual(got, []domain.Mutation{completed}) {
		t.Errorf("mutations = %+v, want the completion unchanged", got)
	}
}

func TestRunMovementPhase_RefusedArrivalHoldsWithItsCargo(t *testing.T) {
	eng, factionState, alpha, _ := arrivalFixture("ship", "troops", 3)
	discard := slog.New(slog.DiscardHandler)
	eng.World = world.NewWithMap(&stubWorldMap{}, discard)
	effects.NewTransportHandler(eng.Rulebook.Assets["ship"]).Apply(alpha, alpha.Assets["t1"], eng.Hooks, nil)
	collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) { return false, nil }}
	paths := &campaigns.Paths{HistoryPath: t.TempDir() + "/history.jsonl"}

	err := eng.runMovementPhase(alpha, factionState, paths, Collectors{Phase: collector}, &errorObserver{}, discard)

	if err != nil {
		t.Fatalf("runMovementPhase: %v", err)
	}
	carrier, cargo := alpha.Assets["t1"], alpha.Assets["c1"]
	if carrier.CurrentOrder == nil || carrier.CurrentOrder.StepIdx != 1 {
		t.Errorf("carrier order = %+v, want it open at step 1", carrier.CurrentOrder)
	}
	if carrier.Location != (domain.Location{RegionHex: arrivalHex(1)}) {
		t.Errorf("carrier at %+v, want the edge hex in orbit", carrier.Location)
	}
	if cargo.Location != (domain.Location{RegionHex: arrivalHex(1)}) {
		t.Errorf("cargo at %+v, want it held with its carrier (D-28)", cargo.Location)
	}
}

// lockKrylos gives bravo an active Transport Lockdown on Krylos that holds alpha out.
func lockKrylos(factionState *state.FactionState) {
	factionState.Factions["bravo"].Assets = map[string]*domain.Asset{"lock": {
		ID: "lock", DefinitionID: "lockdown", OwnerID: "bravo", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 10,
		Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "alpha", WorldID: "Krylos"}},
	}}
}

func TestCheckArrivals_Lockdown(t *testing.T) {
	cases := []struct {
		name      string
		coin      int
		paid      bool
		landers   bool
		wantAsked bool
		want      []domain.Mutation
	}{
		{
			name: "pays and waits", coin: 5,
			want: []domain.Mutation{
				domain.CoinDelta{FactionID: "alpha", Delta: -3, Cause: "lock_out", CausedByFactionID: "bravo"},
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldLockdownPaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			},
		},
		{
			name: "cannot pay", coin: 2,
			want: []domain.Mutation{
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldLockdownUnpaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			},
		},
		{name: "already paid: asks the government", coin: 5, paid: true, wantAsked: true},
		{name: "Deep Strike Landers get through", coin: 5, landers: true, wantAsked: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture("ship", "troops", 3)
			eng.Rand = dicetest.Constant{Value: 3}
			eng.Rulebook.Assets["ship"].Transport.IgnoresLockdown = tc.landers
			lockKrylos(factionState)
			alpha.Coin = tc.coin
			alpha.Assets["t1"].CurrentOrder.LockdownPaid = tc.paid
			asked := false
			collector := &minimalPhaseCollector{requestPermission: func(hooks.PermissionRequest) (bool, error) {
				asked = true
				return true, nil
			}}

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, collector)

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if asked != tc.wantAsked {
				t.Errorf("asked the government = %v, want %v (the lockdown comes first)", asked, tc.wantAsked)
			}
			want := tc.want
			if want == nil {
				want = []domain.Mutation{completed}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_LockdownFeeComesFromTheEntry(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("ship", "", 3)
	eng.Rand = dicetest.Constant{Value: 3}
	eng.Rulebook.Assets["lockdown"].Effects[0].LockOut.Fee = domain.DiceRoll{NumDice: 1, Sides: 6, Modifier: 1}
	lockKrylos(factionState)

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	fee, ok := got[0].(domain.CoinDelta)
	if !ok || fee.Delta != -4 {
		t.Errorf("first mutation = %+v, want a CoinDelta of -4 from 1d6+1 on a 3", got[0])
	}
}

func TestCheckArrivals_LockdownWithNoFeeIsAnError(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("ship", "", 3)
	eng.Rand = &dicetest.Sequence{}
	eng.Rulebook.Assets["lockdown"].Effects = nil
	lockKrylos(factionState)

	if _, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, &minimalPhaseCollector{}); err == nil {
		t.Fatal("checkArrivals: err = nil, want an error naming the missing fee")
	}
}

func TestCheckArrivals_AFactionThatCannotPayStillRolls(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("ship", "", 3)
	roller := &dicetest.Sequence{Values: []int{3}}
	eng.Rand = roller
	lockKrylos(factionState)
	alpha.Coin = 2

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	if roller.Rolled() != 1 {
		t.Errorf("rolled %d dice, want 1", roller.Rolled())
	}
	held, ok := got[0].(domain.MovementOrderHeld)
	if !ok || held.Reason != domain.HoldLockdownUnpaid {
		t.Errorf("first mutation = %+v, want an unpaid hold", got[0])
	}
}

func TestCheckArrivals_TwoLockdownsCannotSpendTheSameCoin(t *testing.T) {
	eng, factionState, alpha, completed := arrivalFixture("mercs", "", 3)
	eng.Rand = dicetest.Constant{Value: 3}
	lockKrylos(factionState)
	alpha.Coin = 4
	second := alpha.Assets["t1"].Clone()
	second.ID = "t2"
	second.CurrentOrder.AssetID = "t2"
	alpha.Assets["t2"] = second
	secondCompleted := completed
	secondCompleted.AssetID = "t2"

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed, secondCompleted}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	var reasons []domain.HoldReason
	for _, mutation := range got {
		if held, ok := mutation.(domain.MovementOrderHeld); ok {
			reasons = append(reasons, held.Reason)
		}
	}
	if want := []domain.HoldReason{domain.HoldLockdownPaid, domain.HoldLockdownUnpaid}; !reflect.DeepEqual(reasons, want) {
		t.Errorf("hold reasons = %v, want %v (4 Coin pays one 3-Coin fee)", reasons, want)
	}
}

// stubPirates charges 1 Coin on every world but its owner's own assets.
type stubPirates struct{ ownerID string }

func (stub *stubPirates) ArrivalToll(mover *domain.Faction, _ string) int {
	if mover.ID == stub.ownerID {
		return 0
	}
	return 1
}

func TestCheckArrivals_Tolls(t *testing.T) {
	toll := func(payerID string) []domain.Mutation {
		return []domain.Mutation{
			domain.CoinDelta{FactionID: payerID, Delta: -1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
			domain.CoinDelta{FactionID: "bravo", Delta: 1, Cause: "arrival_toll", CausedByFactionID: "bravo"},
		}
	}
	cases := []struct {
		name          string
		carrierDef    string
		coin          int
		stealthyCargo bool
		capturedCargo bool
		wantTolls     []string
		wantHeld      bool
	}{
		{name: "carrier and cargo each pay", carrierDef: "ship", coin: 5, wantTolls: []string{"alpha", "alpha"}},
		{name: "a stealthed rider pays nothing", carrierDef: "ship", coin: 5, stealthyCargo: true, wantTolls: []string{"alpha"}},
		{name: "a captured rider's owner pays for it", carrierDef: "ship", coin: 5, capturedCargo: true, wantTolls: []string{"alpha", "charlie"}},
		{name: "a lockdown bypass covers the whole order", carrierDef: "landers", coin: 5},
		{name: "an owner who cannot pay holds the order", carrierDef: "ship", coin: 1, wantHeld: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng, factionState, alpha, completed := arrivalFixture(tc.carrierDef, "militia", 3)
			delete(factionState.Governments, "Krylos")
			alpha.Coin = tc.coin
			alpha.Assets["c1"].Stealthy = tc.stealthyCargo
			if tc.capturedCargo {
				rider := alpha.Assets["c1"]
				delete(alpha.Assets, "c1")
				rider.OwnerID = "charlie"
				factionState.Factions["charlie"] = &domain.Faction{ID: "charlie", Coin: 5, Assets: map[string]*domain.Asset{"c1": rider}}
			}
			eng.Hooks.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "pirates", &stubPirates{ownerID: "bravo"})

			got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{completed}, &minimalPhaseCollector{})

			if err != nil {
				t.Fatalf("checkArrivals: %v", err)
			}
			if tc.wantHeld {
				want := []domain.Mutation{domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: arrivalHex(1), Reason: domain.HoldTollUnpaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"}}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("mutations = %+v, want %+v", got, want)
				}
				return
			}
			want := []domain.Mutation{completed}
			for _, payerID := range tc.wantTolls {
				want = append(want, toll(payerID)...)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCheckArrivals_TwoTollsCannotSpendTheSameCoin(t *testing.T) {
	eng, factionState, alpha, first := arrivalFixture("ship", "", 3)
	delete(factionState.Governments, "Krylos")
	alpha.Coin = 1
	second := alpha.Assets["t1"].Clone()
	second.ID = "t2"
	second.CurrentOrder.AssetID = "t2"
	second.DefinitionID = "militia"
	alpha.Assets["t2"] = second
	secondCompleted := first
	secondCompleted.AssetID = "t2"
	eng.Hooks.RegisterArrivalTollRule(hooks.FactionScope("bravo"), "pirates", &stubPirates{ownerID: "bravo"})

	got, err := eng.checkArrivals(alpha, factionState, []domain.Mutation{first, secondCompleted}, &minimalPhaseCollector{})

	if err != nil {
		t.Fatalf("checkArrivals: %v", err)
	}
	held, ok := got[len(got)-1].(domain.MovementOrderHeld)
	if !ok || held.AssetID != "t2" || held.Reason != domain.HoldTollUnpaid {
		t.Errorf("last mutation = %+v, want t2 held unpaid: the first toll spent the one Coin", got[len(got)-1])
	}
}
