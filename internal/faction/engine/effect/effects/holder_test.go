package effects

import (
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
