package effects_test

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func deepRootedEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger: domain.TriggerHomeworldChanged,
		Effect:  domain.SpecialRemoveTag,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
	}
}

func TestRemoveTag_DropsTheTagOnItsOwnersHomeworldChange(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-002"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialRemoveTag].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-002", Name: "Deep Rooted"}}, Entry: deepRootedEntry(), Registry: registry,
	})
	reactor := registry.MutationReactorsForFaction("alpha")[0].Hook
	moved := domain.HomeworldChanged{FactionID: "alpha", ToWorld: domain.Location{WorldID: "Krylos"}}

	got := reactor.OnMutations([]domain.Mutation{moved, moved}, factionState, nil)
	want := []domain.Mutation{domain.TagRemoved{FactionID: "alpha", TagID: "SWN-T-002", Cause: "remove_tag", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("own homeworld change: got %+v, want %+v (once)", got, want)
	}
	if got := reactor.OnMutations([]domain.Mutation{domain.HomeworldChanged{FactionID: "beta"}}, factionState, nil); got != nil {
		t.Errorf("rival's homeworld change: got %+v, want none", got)
	}
	alpha.Tags = nil
	if got := reactor.OnMutations([]domain.Mutation{moved}, factionState, nil); got != nil {
		t.Errorf("tag already gone: got %+v, want none", got)
	}
}

func TestRemoveTag_IsForTagsOnly(t *testing.T) {
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	asset := &domain.AssetDefinition{ID: "SWN-X", Effects: []domain.EffectDefinition{deepRootedEntry()}}
	if err := engine.RegisterEffectsList(asset); err == nil || !strings.Contains(err.Error(), "only a tag") {
		t.Errorf("asset holder: err = %v, want an only-a-tag refusal", err)
	}
}
