package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func registerPopularMovement(t *testing.T, owner *domain.Faction, movement *domain.Asset) hooks.PermissionGranter {
	t.Helper()
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialGrantPermission].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: owner, Asset: movement, Def: &domain.AssetDefinition{ID: "SWN-C7-001"}}, Registry: registry,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerPermissionQuery,
			Effect:  domain.SpecialGrantPermission,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld},
		},
	})
	granters := registry.PermissionGrantersForAsset(owner.ID, movement.ID)
	if len(granters) != 1 {
		t.Fatalf("granters = %d, want 1", len(granters))
	}
	return granters[0].Hook
}

func TestGrantPermission_GrantsItsOwnerOnItsWorld(t *testing.T) {
	movement := &domain.Asset{ID: "pm", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 16}
	alpha := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{"pm": movement}}
	bravo := &domain.Faction{ID: "bravo"}
	granter := registerPopularMovement(t, alpha, movement)
	def := &domain.AssetDefinition{ID: "gunships"}

	cases := []struct {
		name  string
		query hooks.PermissionQuery
		want  bool
	}{
		{"its owner on its world", hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}, true},
		{"its owner elsewhere", hooks.PermissionQuery{Requester: alpha, WorldID: "Hadrian", Def: def}, false},
		{"a rival on its world", hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := granter.GrantsPermission(tc.query); got != tc.want {
				t.Errorf("GrantsPermission = %v, want %v", got, tc.want)
			}
		})
	}

	movement.Ready = false
	if granter.GrantsPermission(hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}) {
		t.Error("an inactive Popular Movement granted, want no grant")
	}
}
