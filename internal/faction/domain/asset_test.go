package domain_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestNewPurchasedAsset(t *testing.T) {
	faction := &domain.Faction{
		ID:     "f1",
		Assets: map[string]*domain.Asset{},
	}
	def := &domain.AssetDefinition{ID: "SWN-F1-001", HP: 6, Cost: 4}

	asset := domain.NewPurchasedAsset(nil, faction, def, "world-a")

	if asset.DefinitionID != def.ID {
		t.Errorf("DefinitionID = %q, want %q", asset.DefinitionID, def.ID)
	}
	if asset.OwnerID != faction.ID {
		t.Errorf("OwnerID = %q, want %q", asset.OwnerID, faction.ID)
	}
	if asset.Location.WorldID != "world-a" {
		t.Errorf("Location.WorldID = %q, want %q", asset.Location.WorldID, "world-a")
	}
	if asset.CurrentHP != def.HP {
		t.Errorf("CurrentHP = %d, want %d", asset.CurrentHP, def.HP)
	}
	if asset.Ready {
		t.Error("Ready = true, want false (activates next turn)")
	}
	if !asset.Maintained {
		t.Error("Maintained = false, want true")
	}
}

func TestAsset_ActiveAndTargetable(t *testing.T) {
	disabled := []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}
	sheltered := []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "b1"}}
	cases := []struct {
		name           string
		asset          domain.Asset
		wantActive     bool
		wantTargetable bool
	}{
		{"ready, maintained, alive", domain.Asset{Ready: true, Maintained: true, CurrentHP: 1}, true, true},
		{"not ready", domain.Asset{Ready: false, Maintained: true, CurrentHP: 1}, false, false},
		{"unmaintained", domain.Asset{Ready: true, Maintained: false, CurrentHP: 1}, false, false},
		{"destroyed", domain.Asset{Ready: true, Maintained: true, CurrentHP: 0}, false, false},
		{"disabled", domain.Asset{Ready: true, Maintained: true, CurrentHP: 1, Conditions: disabled}, false, true},
		{"sheltered", domain.Asset{Ready: true, Maintained: true, CurrentHP: 2, Conditions: sheltered}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.asset.Active(); got != tc.wantActive {
				t.Errorf("Active() = %v, want %v", got, tc.wantActive)
			}
			if got := tc.asset.Targetable(); got != tc.wantTargetable {
				t.Errorf("Targetable() = %v, want %v", got, tc.wantTargetable)
			}
		})
	}
}

func TestAsset_Hidden(t *testing.T) {
	cases := []struct {
		name  string
		asset domain.Asset
		want  bool
	}{
		{"in the open", domain.Asset{}, false},
		{"stealthed", domain.Asset{Stealthy: true}, true},
		{"sheltered", domain.Asset{Conditions: []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "b1"}}}, true},
		{"disabled", domain.Asset{Conditions: []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.asset.Hidden(); got != tc.want {
				t.Errorf("Hidden() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRemovesOrMoves(t *testing.T) {
	cases := []struct {
		name      string
		mutations []domain.Mutation
		want      bool
	}{
		{"removed", []domain.Mutation{domain.AssetRemoved{AssetID: "a1"}}, true},
		{"moved to another faction", []domain.Mutation{domain.AssetOwnerChanged{AssetID: "a1", ToFactionID: "f2"}}, true},
		{"only damaged", []domain.Mutation{domain.AssetHPDelta{AssetID: "a1", Delta: -9}}, false},
		{"another asset removed", []domain.Mutation{domain.AssetRemoved{AssetID: "a2"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.RemovesOrMoves(tc.mutations, "a1"); got != tc.want {
				t.Errorf("RemovesOrMoves = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewPurchasedAsset_MonotonicID(t *testing.T) {
	def := &domain.AssetDefinition{ID: "SWN-F1-001", HP: 4}
	faction := &domain.Faction{
		ID: "f1",
		Assets: map[string]*domain.Asset{
			"f1-SWN-F1-001-1": {ID: "f1-SWN-F1-001-1"},
			"f1-SWN-F1-001-2": {ID: "f1-SWN-F1-001-2"},
		},
	}
	asset := domain.NewPurchasedAsset(nil, faction, def, "w1")
	if asset.ID != "f1-SWN-F1-001-3" {
		t.Errorf("ID = %q, want %q", asset.ID, "f1-SWN-F1-001-3")
	}
}

func TestNextAssetID_ScansEveryFaction(t *testing.T) {
	owner := &domain.Faction{
		ID:     "f1",
		Assets: map[string]*domain.Asset{"f1-SWN-F1-001-1": {ID: "f1-SWN-F1-001-1"}},
	}
	captor := &domain.Faction{
		ID:     "f2",
		Assets: map[string]*domain.Asset{"f1-SWN-F1-001-3": {ID: "f1-SWN-F1-001-3"}},
	}
	factions := map[string]*domain.Faction{"f1": owner, "f2": captor}

	if got := domain.NextAssetID(factions, owner, "SWN-F1-001"); got != "f1-SWN-F1-001-4" {
		t.Errorf("NextAssetID = %q, want %q (a captured asset still holds its suffix)", got, "f1-SWN-F1-001-4")
	}
}

func TestAsset_CloneCopiesConditions(t *testing.T) {
	original := &domain.Asset{ID: "a1", Conditions: []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1"}}}

	clone := original.Clone()
	clone.Conditions[0].Kind = "changed"

	if got := original.Conditions[0].Kind; got != domain.ConditionNoAbility {
		t.Errorf("original Conditions[0].Kind = %q, want %q (clone must not share the slice)", got, domain.ConditionNoAbility)
	}
}

func TestAssetDefinition_HalfPrice(t *testing.T) {
	cases := []struct {
		cost int
		want int
	}{
		{cost: 10, want: 5},
		{cost: 7, want: 3},
		{cost: 1, want: 1},
	}
	for _, tc := range cases {
		def := &domain.AssetDefinition{Cost: tc.cost}
		if got := def.HalfPrice(); got != tc.want {
			t.Errorf("HalfPrice() with Cost %d = %d, want %d", tc.cost, got, tc.want)
		}
	}
}

func TestAsset_HasCondition(t *testing.T) {
	asset := &domain.Asset{ID: "a1"}
	if asset.HasCondition(domain.ConditionNoAbility) {
		t.Error("HasCondition(no_ability) = true on an asset with no conditions")
	}

	asset.Conditions = []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1"}}
	if !asset.HasCondition(domain.ConditionNoAbility) {
		t.Error("HasCondition(no_ability) = false on an asset that carries it")
	}
	if asset.HasCondition("other") {
		t.Error("HasCondition(other) = true; only no_ability is present")
	}
}

func TestAssetDefinition_NeedsArrivalPermission(t *testing.T) {
	restricted := &domain.AssetDefinition{ID: "troops", Flags: []domain.AssetFlag{domain.FlagPermission}}
	raiseOnly := &domain.AssetDefinition{ID: "fleet", PurchasePermission: true}
	cases := []struct {
		name  string
		def   *domain.AssetDefinition
		asset *domain.Asset
		want  bool
	}{
		{"a P asset", restricted, &domain.Asset{}, true},
		{"a stealthed P asset", restricted, &domain.Asset{Stealthy: true}, false},
		{"raise-only permission", raiseOnly, &domain.Asset{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.def.NeedsArrivalPermission(tc.asset); got != tc.want {
				t.Errorf("NeedsArrivalPermission = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCondition_IsComparesTheTargetFaction(t *testing.T) {
	bravo := domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "bravo"}
	charlie := domain.Condition{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "charlie"}
	if bravo.Is(charlie) {
		t.Error("lock-outs against two factions are one condition, want two")
	}
	if !bravo.Is(bravo) {
		t.Error("a lock-out is not itself")
	}
	elsewhere := bravo
	elsewhere.WorldID = "Hadrian"
	if bravo.Is(elsewhere) {
		t.Error("lock-outs on two worlds are one condition, want two")
	}
}

func TestLockedOutBy(t *testing.T) {
	lockdown := func(owner string) *domain.Asset {
		return &domain.Asset{ID: "lock", OwnerID: owner, Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 10,
			Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock", TargetFactionID: "bravo", WorldID: "Krylos"}}}
	}
	movedTo := func(worldID string) func(*domain.Asset) {
		return func(asset *domain.Asset) { asset.Location.WorldID = worldID }
	}
	cases := []struct {
		name    string
		holder  string
		mutate  func(*domain.Asset)
		world   string
		faction string
		want    string
	}{
		{name: "the struck faction on the Lockdown's world", holder: "alpha", world: "Krylos", faction: "bravo", want: "alpha"},
		{name: "another world", holder: "alpha", world: "Hadrian", faction: "bravo"},
		{name: "another faction", holder: "alpha", world: "Krylos", faction: "charlie"},
		{name: "an inactive Lockdown", holder: "alpha", mutate: func(asset *domain.Asset) { asset.Maintained = false }, world: "Krylos", faction: "bravo"},
		{name: "captured into the struck faction", holder: "bravo", world: "Krylos", faction: "bravo"},
		{name: "the Lockdown moved: its new world is free", holder: "alpha", mutate: movedTo("Hadrian"), world: "Hadrian", faction: "bravo"},
		{name: "the Lockdown moved: the world of the hit is free", holder: "alpha", mutate: movedTo("Hadrian"), world: "Krylos", faction: "bravo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			asset := lockdown(tc.holder)
			if tc.mutate != nil {
				tc.mutate(asset)
			}
			factions := map[string]*domain.Faction{tc.holder: {ID: tc.holder, Assets: map[string]*domain.Asset{"lock": asset}}}
			if got := domain.LockedOutBy(factions, tc.world, tc.faction); got != tc.want {
				t.Errorf("LockedOutBy = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLockedOutBy_TwoHoldersNameTheFirstByID(t *testing.T) {
	lockdown := func(owner string) *domain.Asset {
		return &domain.Asset{ID: "lock-" + owner, OwnerID: owner, Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 10,
			Conditions: []domain.Condition{{Kind: domain.ConditionLockedOut, SourceAssetID: "lock-" + owner, TargetFactionID: "bravo", WorldID: "Krylos"}}}
	}
	factions := map[string]*domain.Faction{}
	for _, owner := range []string{"delta", "alpha", "charlie"} {
		factions[owner] = &domain.Faction{ID: owner, Assets: map[string]*domain.Asset{"lock-" + owner: lockdown(owner)}}
	}
	for range 20 {
		if got := domain.LockedOutBy(factions, "Krylos", "bravo"); got != "alpha" {
			t.Fatalf("LockedOutBy = %q, want alpha, the first holder by ID", got)
		}
	}
}
