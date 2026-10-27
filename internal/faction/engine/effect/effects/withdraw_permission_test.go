package effects_test

import (
	"log/slog"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

func lobbyistsEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger: domain.TriggerPermissionGranted,
		Effect:  domain.SpecialWithdrawPermission,
		Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld},
		Gates: domain.EffectGates{Contest: &domain.EffectContest{
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
		}},
	}
}

func registerLobbyists(t *testing.T, owner *domain.Faction, lobbyists *domain.Asset, roller domain.Roller) hooks.PermissionContester {
	t.Helper()
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialWithdrawPermission].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: owner, Asset: lobbyists, Def: &domain.AssetDefinition{ID: "SWN-C2-001"}}, Registry: registry, Roller: roller,
		Entry: lobbyistsEntry(),
	})
	contesters := registry.PermissionContestersForAsset(owner.ID, lobbyists.ID)
	if len(contesters) != 1 {
		t.Fatalf("contesters = %d, want 1", len(contesters))
	}
	return contesters[0].Hook
}

func lobbyistsFixture() (*domain.Faction, *domain.Faction, *domain.Asset) {
	lobbyists := &domain.Asset{ID: "lob", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}, Ready: true, Maintained: true, CurrentHP: 4}
	alpha := &domain.Faction{ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{"lob": lobbyists}}
	bravo := &domain.Faction{ID: "bravo", Cunning: 2}
	return alpha, bravo, lobbyists
}

func TestWithdrawPermission_OffersOnARivalsGrantOnItsWorld(t *testing.T) {
	alpha, bravo, lobbyists := lobbyistsFixture()
	contester := registerLobbyists(t, alpha, lobbyists, nil)
	def := &domain.AssetDefinition{ID: "gunships"}

	cases := []struct {
		name  string
		query hooks.PermissionQuery
		want  bool
	}{
		{"a rival on its world", hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}, true},
		{"its owner on its world", hooks.PermissionQuery{Requester: alpha, WorldID: "Krylos", Def: def}, false},
		{"a rival elsewhere", hooks.PermissionQuery{Requester: bravo, WorldID: "Hadrian", Def: def}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			offer := contester.OfferContest(tc.query)
			if (offer != nil) != tc.want {
				t.Fatalf("OfferContest = %+v, want an offer: %v", offer, tc.want)
			}
			if offer != nil && offer.Source != lobbyists {
				t.Errorf("Source = %+v, want the Lobbyists", offer.Source)
			}
		})
	}

	lobbyists.Ready = false
	if contester.OfferContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos", Def: def}) != nil {
		t.Error("an inactive Lobbyists offered a contest, want none")
	}
}

func TestWithdrawPermission_WinsContestRollsCunning(t *testing.T) {
	cases := []struct {
		name  string
		rolls []int
		want  bool
	}{
		{"the owner rolls higher", []int{5, 5}, true},
		{"a tie goes to the requester", []int{3, 4}, false},
		{"the requester rolls higher", []int{1, 9}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alpha, bravo, lobbyists := lobbyistsFixture()
			contester := registerLobbyists(t, alpha, lobbyists, &seqRoller{values: tc.rolls})

			if got := contester.WinsContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos"}); got != tc.want {
				t.Errorf("WinsContest = %v, want %v (Cunning 3 vs 2)", got, tc.want)
			}
		})
	}
}

func TestWithdrawPermission_NeedsAContestGate(t *testing.T) {
	entry := lobbyistsEntry()
	entry.Gates = domain.EffectGates{}
	def := &domain.AssetDefinition{ID: "SWN-C2-001", Effects: []domain.EffectDefinition{entry}}

	if err := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler)).RegisterEffectsList(def); err == nil {
		t.Error("RegisterEffectsList accepted withdraw_permission with no contest gate")
	}
}
