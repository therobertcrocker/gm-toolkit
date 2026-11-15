package adapter

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestRelocateVia(t *testing.T) {
	perAsset := domain.RelocateParams{Scope: domain.RelocateOther, CoinCost: 1, CoinCostMode: domain.RelocateCostPerAsset}
	flat := domain.RelocateParams{Scope: domain.RelocateOther, CoinCost: 1, CoinCostMode: domain.RelocateCostFlat}
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"authority": {ID: "authority", Name: "Aether-Gate Transit Authority", Effects: []domain.EffectDefinition{{
			Effect:        domain.SpecialGrantMovement,
			GrantMovement: &domain.GrantMovementParams{Relocate: perAsset, Public: true},
		}}},
		"web": {ID: "web", Name: "Transit Web", Effects: []domain.EffectDefinition{{
			Effect:        domain.SpecialGrantMovement,
			GrantMovement: &domain.GrantMovementParams{Relocate: flat},
		}}},
		"convoy": {ID: "convoy", Name: "Merchant Convoy", Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{{Relocate: &perAsset}}}},
		"plain":  {ID: "plain", Name: "Plain"},
	}}
	bridge := &Adapter{
		engine: &engine.Engine{Rulebook: book},
		factionState: &state.FactionState{Factions: map[string]*domain.Faction{
			"fey":   {ID: "fey", Name: "Fey Royals"},
			"rider": {ID: "rider", Name: "Thymian Armada"},
		}},
	}
	authority := &domain.Asset{ID: "a1", DefinitionID: "authority", OwnerID: "fey"}
	riderAsset := &domain.Asset{ID: "r1", DefinitionID: "plain", OwnerID: "rider"}
	feyAsset := &domain.Asset{ID: "f1", DefinitionID: "plain", OwnerID: "fey"}

	cases := []struct {
		name       string
		acting     *domain.Asset
		candidates []*domain.Asset
		want       string
	}{
		{"a self-jump names nothing", authority, []*domain.Asset{authority}, ""},
		{"no candidates names nothing", authority, nil, ""},
		{"a rival rides a public grant for its fare", authority, []*domain.Asset{riderAsset}, "via Aether-Gate Transit Authority (Fey Royals) · 1 Coin per asset"},
		{"the owner rides its public grant free", authority, []*domain.Asset{feyAsset}, "via Aether-Gate Transit Authority (Fey Royals) · free"},
		{"a private flat grant", &domain.Asset{ID: "w1", DefinitionID: "web", OwnerID: "fey"}, []*domain.Asset{feyAsset}, "via Transit Web (Fey Royals) · 1 Coin"},
		{"an ability step", &domain.Asset{ID: "c1", DefinitionID: "convoy", OwnerID: "fey"}, []*domain.Asset{feyAsset}, "via Merchant Convoy (Fey Royals) · 1 Coin per asset"},
		{"a definition with no relocate names nothing", &domain.Asset{ID: "p1", DefinitionID: "plain", OwnerID: "fey"}, []*domain.Asset{feyAsset}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bridge.relocateVia(tc.acting, tc.candidates); got != tc.want {
				t.Errorf("relocateVia = %q, want %q", got, tc.want)
			}
		})
	}
}
