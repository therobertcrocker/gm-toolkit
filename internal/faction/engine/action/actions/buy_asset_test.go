package actions

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

func makeBuyRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"cheap": {ID: "cheap", Name: "Militia", Category: domain.StatForce, HP: 4, Cost: 4, MinRating: 2},
		},
	}
}

func TestBuyAsset_Validate(t *testing.T) {
	rulebook := makeBuyRulebook()

	t.Run("eligible asset and sufficient coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}
		if !NewBuyAsset(mocks.NewMockCollector(ctrl), nil, nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction can afford an eligible asset")
		}
	})

	t.Run("not enough coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 2, Homeworld: domain.Location{WorldID: "Tartarus"}}
		if NewBuyAsset(mocks.NewMockCollector(ctrl), nil, nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction cannot afford any asset")
		}
	})

	t.Run("stat below MinRating", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 1, Coin: 10, Homeworld: domain.Location{WorldID: "Tartarus"}}
		if NewBuyAsset(mocks.NewMockCollector(ctrl), nil, nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction stat is below MinRating")
		}
	})
}

// TestBuyAsset_Output: purchase emits AssetAdded(Ready=false) then CoinDelta(-cost).
func TestBuyAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeBuyRulebook()
	def := rulebook.Assets["cheap"]
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(
		action.BuyOrder{World: "Tartarus", Definition: def}, nil,
	)

	act := NewBuyAsset(collector, nil, nil, nil)
	if err := act.Inputs(faction, nil, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, &state.FactionState{}, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	added, ok := mutations[0].(domain.AssetAdded)
	if !ok || added.FactionID != "f1" || added.Asset.DefinitionID != "cheap" {
		t.Errorf("mutations[0] = %v, want AssetAdded{f1, cheap}", mutations[0])
	}
	if added.Asset.Ready {
		t.Error("new asset must have Ready=false until next turn start")
	}
	if added.Asset.Location != (domain.Location{WorldID: "Tartarus"}) {
		t.Errorf("new asset Location = %v, want Tartarus", added.Asset.Location)
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -4 || coin.Cause != "buy" {
		t.Errorf("mutations[1] = %v, want CoinDelta{f1, -4, buy}", mutations[1])
	}
}

// TestBuyAsset_PendingDiscount: a faction's pending purchase discount reduces the
// emitted CoinDelta, is floored at half the base cost, and is consumed by a single
// PurchaseDiscountConsumed mutation. With no pending discount, none is emitted.
func TestBuyAsset_PendingDiscount(t *testing.T) {
	rulebook := makeBuyRulebook()
	def := rulebook.Assets["cheap"] // base Cost 4, so half-base floor is 2

	run := func(t *testing.T, discount int) []domain.Mutation {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{
			ID: "f1", Force: 3, Coin: 10,
			Homeworld:               domain.Location{WorldID: "Tartarus"},
			PendingPurchaseDiscount: discount,
		}
		collector := mocks.NewMockCollector(ctrl)
		collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(
			action.BuyOrder{World: "Tartarus", Definition: def}, nil,
		)
		act := NewBuyAsset(collector, nil, nil, nil)
		if err := act.Inputs(faction, nil, rulebook); err != nil {
			t.Fatalf("Inputs: %v", err)
		}
		if err := act.Resolve(faction, &state.FactionState{}, rulebook); err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		muts, err := act.Output()
		if err != nil {
			t.Fatalf("Output: %v", err)
		}
		return muts
	}

	inspect := func(muts []domain.Mutation) (delta int, consumed bool) {
		for _, m := range muts {
			switch v := m.(type) {
			case domain.CoinDelta:
				delta = v.Delta
			case domain.PurchaseDiscountConsumed:
				consumed = true
			}
		}
		return delta, consumed
	}

	t.Run("discount reduces cost and is consumed", func(t *testing.T) {
		delta, consumed := inspect(run(t, 1))
		if delta != -3 {
			t.Errorf("CoinDelta.Delta = %d, want -3 (base 4 - 1 discount)", delta)
		}
		if !consumed {
			t.Error("expected a PurchaseDiscountConsumed mutation")
		}
	})

	t.Run("discount floored at half base cost", func(t *testing.T) {
		delta, consumed := inspect(run(t, 3)) // 4 - 3 = 1, clamped up to floor 2
		if delta != -2 {
			t.Errorf("CoinDelta.Delta = %d, want -2 (floored at half base)", delta)
		}
		if !consumed {
			t.Error("expected a PurchaseDiscountConsumed mutation")
		}
	})

	t.Run("no pending discount emits no consume", func(t *testing.T) {
		delta, consumed := inspect(run(t, 0))
		if delta != -4 {
			t.Errorf("CoinDelta.Delta = %d, want -4 (no discount)", delta)
		}
		if consumed {
			t.Error("expected no PurchaseDiscountConsumed when no pending discount")
		}
	})
}

// stubCostReducer reduces asset purchase cost by 1 Coin.
type stubCostReducer struct{}

func (stub *stubCostReducer) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	query.Cost--
}

// TestBuyAsset_AssetCostModifier: a registered AssetCostModifier reduces the
// purchase cost by 1; the emitted CoinDelta reflects the reduced cost.
func TestBuyAsset_AssetCostModifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	rb := makeBuyRulebook()
	def := rb.Assets["cheap"]
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}

	registry := hooks.NewRegistry()
	registry.RegisterAssetCostModifier(hooks.FactionScope("f1"), "discount", &stubCostReducer{})

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(
		action.BuyOrder{World: "Tartarus", Definition: def}, nil,
	)

	act := NewBuyAsset(collector, registry, nil, nil)
	if err := act.Inputs(faction, nil, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, &state.FactionState{}, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -3 {
		t.Errorf("CoinDelta.Delta = %v, want -3 (base 4 - 1 discount)", mutations[1])
	}
}

// stubSurcharge raises the cost by one, as cost_surcharge does.
type stubSurcharge struct{}

func (stub *stubSurcharge) ModifyAssetCost(_ *domain.Faction, _ *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	query.Cost++
	query.Surcharges = append(query.Surcharges, hooks.Surcharge{FactionID: "rival", Amount: 1})
}

// surchargedWorld places a rival asset carrying stubSurcharge on world "a".
func surchargedWorld() (*world.WorldEngine, *hooks.Registry, spatial.RegionHex) {
	worldHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	worldEngine := world.NewWithMap(&testHexRouter{worlds: map[string]spatial.RegionHex{"a": worldHex}}, slog.New(slog.DiscardHandler))
	rivalAsset := &domain.Asset{ID: "li", OwnerID: "rival", Location: domain.Location{WorldID: "a", RegionHex: worldHex}}
	worldEngine.Index = &world.Index{AssetsByLocation: map[string][]*domain.Asset{"a": {rivalAsset}}}
	registry := hooks.NewRegistry()
	registry.RegisterAssetCostModifier(hooks.AssetScope("rival", "li"), "surcharge", &stubSurcharge{})
	return worldEngine, registry, worldHex
}

// TestPurchasableDefinitions_PricesWithSurcharge: the menu prices a def at the cost
// Resolve charges, so a buyer holding only the base price is not offered it.
func TestPurchasableDefinitions_PricesWithSurcharge(t *testing.T) {
	worldEngine, registry, worldHex := surchargedWorld()
	rb := makeBuyRulebook()
	for coin, wantOffered := range map[int]bool{4: false, 5: true} {
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: coin, Homeworld: domain.Location{WorldID: "a", RegionHex: worldHex}}
		offered := len(purchasableDefinitions(faction, rb, "a", worldEngine, registry, nil)) > 0
		if offered != wantOffered {
			t.Errorf("Coin %d: offered = %v, want %v (base 4 + surcharge 1)", coin, offered, wantOffered)
		}
	}
}

// TestBuyAsset_DiscountKeepsSurcharge: base 4 + surcharge 1, discount 3. The
// discount floors at half base (2) and the surcharge rides on top: 3.
func TestBuyAsset_DiscountKeepsSurcharge(t *testing.T) {
	worldEngine, registry, worldHex := surchargedWorld()
	rb := makeBuyRulebook()
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 10, PendingPurchaseDiscount: 3, Homeworld: domain.Location{WorldID: "a", RegionHex: worldHex}}
	act := NewBuyAsset(nil, registry, worldEngine, nil)
	act.buyOrder = action.BuyOrder{World: "a", Definition: rb.Assets["cheap"]}
	if err := act.Resolve(faction, &state.FactionState{}, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if act.cost != 3 {
		t.Errorf("cost = %d, want 3 (floor 2 + surcharge 1)", act.cost)
	}
}

// TestBuyAsset_SurchargeIsItsOwnCharge: the surcharge is charged apart from the
// price and caused by the rival that imposed it, so a Bank can absorb it.
func TestBuyAsset_SurchargeIsItsOwnCharge(t *testing.T) {
	worldEngine, registry, worldHex := surchargedWorld()
	rb := makeBuyRulebook()
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 10, Homeworld: domain.Location{WorldID: "a", RegionHex: worldHex}}
	act := NewBuyAsset(nil, registry, worldEngine, nil)
	act.factionID = faction.ID
	act.buyOrder = action.BuyOrder{World: "a", Definition: rb.Assets["cheap"]}
	if err := act.Resolve(faction, &state.FactionState{}, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	var charges []domain.CoinDelta
	for _, mutation := range mutations {
		if coin, ok := mutation.(domain.CoinDelta); ok {
			charges = append(charges, coin)
		}
	}
	want := []domain.CoinDelta{
		{FactionID: "f1", Delta: -4, Cause: "buy", CausedByFactionID: "f1"},
		{FactionID: "f1", Delta: -1, Cause: "cost_surcharge", CausedByFactionID: "rival"},
	}
	if !reflect.DeepEqual(charges, want) {
		t.Errorf("charges = %+v, want %+v", charges, want)
	}
}

// stubWealthFloor lifts the tech level to 6 for Wealth definitions only.
type stubWealthFloor struct{}

func (stub *stubWealthFloor) ModifyWorldTechLevel(_ *domain.Faction, def *domain.AssetDefinition, _ string, baseTL int) int {
	if def.Category == domain.StatWealth {
		return max(baseTL, 6)
	}
	return baseTL
}

// TestPurchasableDefinitions_TechCapResolvesPerDefinition: on a TL5 world, a floor
// on the buyer's own asset that serves Wealth only offers the TL6 Wealth def and
// not the TL6 Force def — one cap per world could not split them.
func TestPurchasableDefinitions_TechCapResolvesPerDefinition(t *testing.T) {
	worldHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	worldEngine := world.NewWithMap(&testHexRouter{worlds: map[string]spatial.RegionHex{"a": worldHex}}, slog.New(slog.DiscardHandler))
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"wealth-tl6": {ID: "wealth-tl6", Category: domain.StatWealth, Cost: 1, TechLevel: 6},
		"force-tl6":  {ID: "force-tl6", Category: domain.StatForce, Cost: 1, TechLevel: 6},
	}}
	lab := &domain.Asset{ID: "lab", OwnerID: "f1", Location: domain.Location{WorldID: "a", RegionHex: worldHex}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Homeworld: domain.Location{WorldID: "a", RegionHex: worldHex}, Assets: map[string]*domain.Asset{"lab": lab}}
	registry := hooks.NewRegistry()
	registry.RegisterWorldTechLevelModifier(hooks.AssetScope("f1", "lab"), "wealth-floor", &stubWealthFloor{})

	defs := purchasableDefinitions(faction, rb, "a", worldEngine, registry, nil)
	if len(defs) != 1 || defs[0].ID != "wealth-tl6" {
		var ids []string
		for _, def := range defs {
			ids = append(ids, def.ID)
		}
		t.Errorf("purchasable = %v, want [wealth-tl6]", ids)
	}
}

func stealthBuyRulebook() *rulebook.Rulebook {
	stealthy := false
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"stealth": {ID: "stealth", Name: "Stealth", Category: domain.StatCunning, Cost: 2, MinRating: 1, Quality: true,
			Effects: []domain.EffectDefinition{{
				Trigger: domain.TriggerOnPurchase,
				Effect:  domain.SpecialGrantStealth,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld,
					AssetTypes: []domain.AssetType{domain.TypeSpecialForces}, Stealthy: &stealthy},
			}}},
		"sf": {ID: "sf", Name: "Saboteurs", Category: domain.StatCunning, Cost: 5, HP: 6, MinRating: 9, Type: domain.TypeSpecialForces},
	}}
}

func stealthBuyer(withTarget bool) *domain.Faction {
	faction := &domain.Faction{ID: "f1", Cunning: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}, Assets: map[string]*domain.Asset{}}
	if withTarget {
		faction.Assets["f1-sf-1"] = &domain.Asset{ID: "f1-sf-1", DefinitionID: "sf", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	}
	return faction
}

func TestBuyAsset_QualityOfferedOnlyWithATarget(t *testing.T) {
	effectsEngine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	rb := stealthBuyRulebook()

	if NewBuyAsset(nil, nil, nil, effectsEngine).Validate(stealthBuyer(false), nil, rb) {
		t.Error("Validate = true with no Special Forces to stealth, want false (D-22)")
	}
	if !NewBuyAsset(nil, nil, nil, effectsEngine).Validate(stealthBuyer(true), nil, rb) {
		t.Error("Validate = false with a Special Forces asset to stealth, want true")
	}
}

func TestBuyAsset_QualityStealthsItsTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	rb := stealthBuyRulebook()
	faction := stealthBuyer(true)
	effectsEngine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: "Tartarus", Definition: rb.Assets["stealth"]}, nil)

	act := NewBuyAsset(collector, nil, nil, effectsEngine)
	if err := act.Inputs(faction, nil, rb); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}, rb); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	want := []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -2, Cause: "buy", CausedByFactionID: "f1"},
		domain.AssetStealthApplied{FactionID: "f1", AssetID: "f1-sf-1", Cause: "grant_stealth", CausedByFactionID: "f1"},
	}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want %+v (no AssetAdded for a quality)", mutations, want)
	}
}

func TestBuyAsset_RefusedQualityAsksForNoTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	rb := stealthBuyRulebook()
	rb.Assets["stealth"].PurchasePermission = true
	faction := stealthBuyer(false)
	faction.Homeworld = domain.Location{WorldID: "Anchorage"}
	for _, id := range []string{"f1-sf-1", "f1-sf-2"} {
		faction.Assets[id] = &domain.Asset{ID: id, DefinitionID: "sf", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
	effectsEngine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: "Tartarus", Definition: rb.Assets["stealth"]}, nil)
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	act := NewBuyAsset(collector, nil, nil, effectsEngine)
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
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want only the refusal record", mutations)
	}
	if _, ok := mutations[0].(domain.PermissionRefused); !ok {
		t.Errorf("mutations[0] = %T, want PermissionRefused", mutations[0])
	}
}

func TestBuyAsset_AsksPermissionAfterTheChoice(t *testing.T) {
	def := &domain.AssetDefinition{ID: "perm", Name: "Gunships", Category: domain.StatForce, HP: 4, Cost: 4, MinRating: 1,
		Flags: []domain.AssetFlag{domain.FlagPermission}}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"perm": def}}

	cases := []struct {
		name    string
		world   string
		ask     bool
		grant   bool
		refused bool
	}{
		{name: "refused", world: "Krylos", ask: true, refused: true},
		{name: "granted", world: "Krylos", ask: true, grant: true},
		{name: "own world", world: "Anchorage"},
		{name: "homeworld a rival governs", world: "Tartarus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Tartarus"}}
			ruler := &domain.Faction{ID: "f2", Name: "Bravo"}
			factionState := &state.FactionState{
				Factions:    map[string]*domain.Faction{"f1": faction, "f2": ruler},
				Governments: map[string]string{"Tartarus": "f2", "Krylos": "f2", "Anchorage": "f1"},
			}
			if !NewBuyAsset(nil, nil, nil, nil).Validate(faction, nil, rb) {
				t.Fatal("Validate = false, want a P asset offered on a world the faction does not rule (D-35)")
			}

			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: tc.world, Definition: def}, nil)
			if tc.ask {
				collector.EXPECT().
					RequestPermission(hooks.PermissionRequest{Requester: faction, Ruler: ruler, WorldID: tc.world, Def: def}).
					Return(tc.grant, nil)
			}

			act := NewBuyAsset(collector, nil, nil, nil)
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

			if tc.refused {
				want := []domain.Mutation{domain.PermissionRefused{
					FactionID: "f1", WorldID: tc.world, DefinitionID: "perm", DecidedByFactionID: "f2", Cause: "buy", CausedByFactionID: "f1",
				}}
				if !reflect.DeepEqual(mutations, want) {
					t.Errorf("mutations = %+v, want only the refusal record (no asset, no charge)", mutations)
				}
				return
			}
			if len(mutations) != 2 {
				t.Fatalf("mutations = %+v, want AssetAdded and CoinDelta", mutations)
			}
			if _, ok := mutations[0].(domain.AssetAdded); !ok {
				t.Errorf("mutations[0] = %T, want AssetAdded", mutations[0])
			}
		})
	}
}

func TestBuyAsset_PurchasePermissionAsksForARaise(t *testing.T) {
	ctrl := gomock.NewController(t)
	def := &domain.AssetDefinition{ID: "fleet", Name: "Capital Fleet", Category: domain.StatForce, HP: 30, Cost: 4, MinRating: 1, PurchasePermission: true}
	rb := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"fleet": def}}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Homeworld: domain.Location{WorldID: "Anchorage"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectBuyOrder(gomock.Any()).Return(action.BuyOrder{World: "Tartarus", Definition: def}, nil)
	collector.EXPECT().RequestPermission(hooks.PermissionRequest{Requester: faction, WorldID: "Tartarus", Def: def}).Return(false, nil)

	act := NewBuyAsset(collector, nil, nil, nil)
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
	if len(mutations) != 1 {
		t.Fatalf("mutations = %+v, want only the refusal record", mutations)
	}
	if _, ok := mutations[0].(domain.PermissionRefused); !ok {
		t.Errorf("mutations[0] = %T, want PermissionRefused", mutations[0])
	}
}
