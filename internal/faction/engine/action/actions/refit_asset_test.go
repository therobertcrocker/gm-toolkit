package actions

import (
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
	"go.uber.org/mock/gomock"
)

func makeRefitRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"old-def": {ID: "old-def", Name: "Militia", Category: domain.StatForce, HP: 4, Cost: 2},
			"new-def": {ID: "new-def", Name: "Infantry", Category: domain.StatForce, HP: 6, Cost: 4},
		},
	}
}

func TestRefitAsset_Validate(t *testing.T) {
	rulebook := makeRefitRulebook()

	t.Run("no assets", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5}
		if NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction has no assets")
		}
	})

	t.Run("cannot afford upgrade delta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// delta = 4-2 = 2; Coin=1 < 2 → no valid replacement
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{WorldID: "Tartarus"}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 1, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when faction cannot afford the cost delta")
		}
	})

	t.Run("sheltered asset cannot be refit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{WorldID: "Tartarus"},
			Conditions: []domain.Condition{{Kind: domain.ConditionSheltered, SourceAssetID: "b1"}}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when the only asset is sheltered")
		}
	})

	t.Run("an asset in transit cannot be refit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{RegionHex: spatial.RegionHex{RegionID: "drift"}}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected false when the only asset has no world")
		}
	})

	t.Run("a quality is never a replacement", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// built from makeRefitRulebook: the test's local rulebook shadows the package name
		qualityRulebook := makeRefitRulebook()
		delete(qualityRulebook.Assets, "new-def")
		qualityRulebook.Assets["stealth"] = &domain.AssetDefinition{ID: "stealth", Category: domain.StatForce, Cost: 2, Quality: true}
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{WorldID: "Tartarus"}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, qualityRulebook) {
			t.Error("expected false when the only other definition is a quality")
		}
	})

	t.Run("valid refit exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", DefinitionID: "old-def", Location: domain.Location{WorldID: "Tartarus"}}
		faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
		if !NewRefitAsset(mocks.NewMockCollector(ctrl), nil, nil).Validate(faction, nil, rulebook) {
			t.Error("expected true when faction can afford the refit")
		}
	})
}

// TestRefitAsset_Output: swap old-def(cost=2) for new-def(cost=4).
// Emits AssetRemoved + AssetAdded(Ready=false, same Location) + CoinDelta(-2).
func TestRefitAsset_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRefitRulebook()
	newDef := rulebook.Assets["new-def"]
	oldAsset := &domain.Asset{ID: "f1-old-def-1", DefinitionID: "old-def", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"f1-old-def-1": oldAsset}}

	order := action.RefitOrder{OldAsset: oldAsset, NewDefinition: newDef}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRefitOrder(gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewRefitAsset(collector, nil, nil)
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

	if len(mutations) != 3 {
		t.Fatalf("len(mutations) = %d, want 3; got %v", len(mutations), mutations)
	}
	removed, ok := mutations[0].(domain.AssetRemoved)
	if !ok || removed.AssetID != "f1-old-def-1" || removed.Cause != "refit" {
		t.Errorf("mutations[0] = %v, want AssetRemoved{f1-old-def-1, refit}", mutations[0])
	}
	added, ok := mutations[1].(domain.AssetAdded)
	if !ok || added.Asset.DefinitionID != "new-def" {
		t.Errorf("mutations[1] = %v, want AssetAdded{new-def}", mutations[1])
	}
	if added.Asset.Ready {
		t.Error("refitted asset must have Ready=false until next turn start")
	}
	if added.Asset.Location != (domain.Location{WorldID: "Tartarus"}) {
		t.Errorf("refitted asset Location = %v, want Tartarus", added.Asset.Location)
	}
	coin, ok := mutations[2].(domain.CoinDelta)
	if !ok || coin.FactionID != "f1" || coin.Delta != -2 || coin.Cause != "refit" {
		t.Errorf("mutations[2] = %v, want CoinDelta{f1, -2, refit}", mutations[2])
	}
}

func TestRefitAsset_PermissionRefused(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeRefitRulebook()
	newDef := rulebook.Assets["new-def"]
	newDef.Flags = []domain.AssetFlag{domain.FlagPermission}
	oldAsset := &domain.Asset{ID: "f1-old-def-1", DefinitionID: "old-def", OwnerID: "f1", Location: domain.Location{WorldID: "Tartarus"}}
	faction := &domain.Faction{ID: "f1", Force: 3, Coin: 5, Assets: map[string]*domain.Asset{"f1-old-def-1": oldAsset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRefitOrder(gomock.Any(), gomock.Any()).Return(action.RefitOrder{OldAsset: oldAsset, NewDefinition: newDef}, nil)
	collector.EXPECT().RequestPermission(gomock.Any()).Return(false, nil)

	act := NewRefitAsset(collector, nil, nil)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	want := []domain.Mutation{domain.PermissionRefused{FactionID: "f1", WorldID: "Tartarus", DefinitionID: "new-def", Cause: "refit", CausedByFactionID: "f1"}}
	if !reflect.DeepEqual(mutations, want) {
		t.Errorf("mutations = %+v, want only the refusal record", mutations)
	}
}

func TestValidRefitReplacements_RequiresTag(t *testing.T) {
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"guards": {ID: "guards", Category: domain.StatForce, MinRating: 1, Cost: 2},
		"slaves": {ID: "slaves", Category: domain.StatForce, MinRating: 1, Cost: 2, RequiresTag: "T-003"},
	}}
	asset := &domain.Asset{ID: "a1", DefinitionID: "guards", Location: domain.Location{WorldID: "Tartarus"}}
	without := &domain.Faction{ID: "f1", Force: 1, Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
	if got := validRefitReplacements(without, asset, rulebook, nil, nil); len(got) != 0 {
		t.Errorf("replacements without the tag = %v, want none", got)
	}
	with := &domain.Faction{ID: "f1", Force: 1, Coin: 5, Tags: []string{"T-003"}, Assets: map[string]*domain.Asset{"a1": asset}}
	if got := validRefitReplacements(with, asset, rulebook, nil, nil); len(got) != 1 || got[0].ID != "slaves" {
		t.Errorf("replacements with the tag = %v, want slaves", got)
	}
}

func TestValidRefitReplacements_ChecksTheWorldsPurchaseRules(t *testing.T) {
	worldHex := spatial.RegionHex{RegionID: "r1", Coord: spatial.HexCoord{Q: 0, R: 0}}
	worldEngine := world.NewWithMap(&testHexRouter{worlds: map[string]spatial.RegionHex{"a": worldHex}}, slog.New(slog.DiscardHandler))
	refitRulebook := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"freighter": {ID: "freighter", Category: domain.StatWealth, Cost: 1, TechLevel: 1, Type: domain.TypeFacility},
			"plant":     {ID: "plant", Category: domain.StatWealth, Cost: 1, TechLevel: 4, Type: domain.TypeFacility},
			"lab":       {ID: "lab", Category: domain.StatWealth, Cost: 1, TechLevel: 6, Type: domain.TypeFacility},
			"ship":      {ID: "ship", Category: domain.StatWealth, Cost: 1, TechLevel: 4, Type: domain.TypeStarship},
		},
		PurchasePopulation: map[domain.AssetType]int{domain.TypeStarship: 500000},
	}
	asset := &domain.Asset{ID: "a1", DefinitionID: "freighter", Location: domain.Location{WorldID: "a", RegionHex: worldHex}}
	faction := &domain.Faction{ID: "f1", Coin: 5, Assets: map[string]*domain.Asset{"a1": asset}}
	replacementIDs := func() []string {
		var ids []string
		for _, def := range validRefitReplacements(faction, asset, refitRulebook, worldEngine, hooks.NewRegistry()) {
			ids = append(ids, def.ID)
		}
		slices.Sort(ids)
		return ids
	}

	if got, want := replacementIDs(), []string{"plant"}; !slices.Equal(got, want) {
		t.Errorf("replacements = %v, want %v (lab is above the world's TL5, ship below its Starship floor)", got, want)
	}
	faction.PopulationWaiver = true
	if got, want := replacementIDs(), []string{"plant", "ship"}; !slices.Equal(got, want) {
		t.Errorf("replacements with the waiver = %v, want %v", got, want)
	}
}
