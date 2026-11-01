package effects

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestHolderRef_Live(t *testing.T) {
	asset := &domain.Asset{ID: "a-1", Ready: true, Maintained: true, CurrentHP: 1}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-020"}, Assets: map[string]*domain.Asset{"a-1": asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	assetRef := refOf(effect.Holder{Faction: alpha, Asset: asset, Def: &domain.AssetDefinition{ID: "SWN-C7-002"}})
	tagRef := refOf(effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-020"}})

	if got, live := assetRef.live(factionState); !live || got != asset {
		t.Errorf("active asset: live = %v, asset = %v", live, got)
	}
	if got, live := tagRef.live(factionState); !live || got != nil {
		t.Errorf("held tag: live = %v, asset = %v; want live and no asset", live, got)
	}
	if _, live := tagRef.live(nil); live {
		t.Error("nil state: live, want not")
	}

	asset.Ready = false
	if _, live := assetRef.live(factionState); live {
		t.Error("unready asset: live, want not")
	}
	alpha.Tags = nil
	if _, live := tagRef.liveIn(alpha); live {
		t.Error("dropped tag: live, want not")
	}
	delete(factionState.Factions, "alpha")
	if _, live := assetRef.live(factionState); live {
		t.Error("removed faction: live, want not")
	}
}

func TestRefuseAssetWideRollScope(t *testing.T) {
	cases := map[string]struct {
		effect  domain.SpecialEffectType
		scope   domain.EffectScope
		wantErr bool
	}{
		"an asset's bonus_die at world scope":     {effect: domain.SpecialBonusDie, scope: domain.ScopeWorld},
		"an asset's bonus_die at faction scope":   {effect: domain.SpecialBonusDie, scope: domain.ScopeFaction, wantErr: true},
		"an asset's bonus_die at homeworld scope": {effect: domain.SpecialBonusDie, scope: domain.ScopeHomeworld, wantErr: true},
		"an asset's reroll at self scope":         {effect: domain.SpecialReroll, scope: domain.ScopeSelf},
		"an asset's reroll at faction scope":      {effect: domain.SpecialReroll, scope: domain.ScopeFaction, wantErr: true},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			definition := &domain.AssetDefinition{ID: "roller", Effects: []domain.EffectDefinition{{
				Trigger: domain.TriggerOwnRoll,
				Effect:  testCase.effect,
				Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: testCase.scope},
			}}}
			err := effect.New(Registrars(), slog.New(slog.DiscardHandler)).RegisterEffectsList(definition)
			if (err != nil) != testCase.wantErr {
				t.Errorf("RegisterEffectsList error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}

	t.Run("a tag's dice at faction and homeworld scope", func(t *testing.T) {
		tag := &domain.Tag{ID: "T-1", Effects: []domain.EffectDefinition{
			{Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialBonusDie, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld}},
			{Trigger: domain.TriggerOwnRoll, Effect: domain.SpecialReroll, Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction}},
		}}
		if err := effect.New(Registrars(), slog.New(slog.DiscardHandler)).CheckTagEffects(tag); err != nil {
			t.Errorf("CheckTagEffects: %v, want nil", err)
		}
	})
}
