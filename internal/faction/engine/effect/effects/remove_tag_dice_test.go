package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/tag/tags"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var blackmailEntry = domain.EffectDefinition{
	Trigger: domain.TriggerEnemyRollAtMe,
	Effect:  domain.SpecialRemoveTagDice,
	Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeSelf},
}

// blackmailState puts alpha's Blackmail "bm" and another asset "other" on Krylos;
// beta is the rival there.
func blackmailState(blackmail *domain.Asset) *state.FactionState {
	blackmail.Location = krylos
	other := activeAsset("other", "SWN-F1-001")
	other.Location = krylos
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Cunning: 2, Assets: map[string]*domain.Asset{"bm": blackmail, "other": other}},
		"beta":  {ID: "beta", Force: 3, Cunning: 2},
	}}
}

func registerBlackmail(factionState *state.FactionState, registry *hooks.Registry) hooks.RollOfferFilter {
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialRemoveTagDice].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["bm"], Def: &domain.AssetDefinition{ID: "SWN-C2-003", Name: "Blackmail"},
		Entry: blackmailEntry, Registry: registry,
	})
	return registry.RollOfferFiltersAt(hooks.AssetScope("alpha", "bm"))[0].Hook
}

var (
	tagOffer   = hooks.ModifierOffer{Source: "tag:Warlike", FromTag: true}
	assetOffer = hooks.ModifierOffer{Source: "Panopticon Matrix"}
)

func offerSources(offers []hooks.ModifierOffer) []string {
	var sources []string
	for _, offer := range offers {
		sources = append(sources, offer.Source)
	}
	return sources
}

func TestRemoveTagDice_FilterOffers(t *testing.T) {
	factionState := blackmailState(activeAsset("bm", "SWN-C2-003"))
	filter := registerBlackmail(factionState, hooks.NewRegistry())
	alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]
	blackmail, other := alpha.Assets["bm"], alpha.Assets["other"]

	cases := []struct {
		name string
		ctx  hooks.RollContext
		want []string
	}{
		{
			name: "rival attacking Blackmail loses tag dice",
			ctx:  hooks.RollContext{Phase: hooks.PhaseAttack, Actor: beta, Opponent: alpha, OpponentAsset: blackmail, World: "Krylos"},
			want: []string{"Panopticon Matrix"},
		},
		{
			name: "rival defending against Blackmail loses tag dice",
			ctx:  hooks.RollContext{Phase: hooks.PhaseDefense, Actor: beta, Opponent: alpha, OpponentAsset: blackmail, World: "Krylos"},
			want: []string{"Panopticon Matrix"},
		},
		{
			name: "rival rolling against another asset keeps tag dice",
			ctx:  hooks.RollContext{Phase: hooks.PhaseAttack, Actor: beta, Opponent: alpha, OpponentAsset: other, World: "Krylos"},
			want: []string{"tag:Warlike", "Panopticon Matrix"},
		},
		{
			name: "owner's own roll with Blackmail keeps tag dice",
			ctx:  hooks.RollContext{Phase: hooks.PhaseAttack, Actor: alpha, Opponent: beta, Asset: blackmail, World: "Krylos"},
			want: []string{"tag:Warlike", "Panopticon Matrix"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kept := filter.FilterOffers(tc.ctx, []hooks.ModifierOffer{tagOffer, assetOffer}, factionState)
			if got := offerSources(kept); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("kept %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRemoveTagDice_InactiveBlackmailStripsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			blackmail := activeAsset("bm", "SWN-C2-003")
			deactivate(blackmail)
			factionState := blackmailState(blackmail)
			filter := registerBlackmail(factionState, hooks.NewRegistry())
			alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]

			ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: beta, Opponent: alpha, OpponentAsset: blackmail, World: "Krylos"}
			kept := filter.FilterOffers(ctx, []hooks.ModifierOffer{tagOffer}, factionState)
			if len(kept) != 1 {
				t.Errorf("kept %d offers, want 1", len(kept))
			}
		})
	}
}

func TestRemoveTagDice_WarlikeAttackOnBlackmailRollsOneDie(t *testing.T) {
	factionState := blackmailState(activeAsset("bm", "SWN-C2-003"))
	registry := hooks.NewRegistry()
	registerBlackmail(factionState, registry)
	alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]
	tags.WarlikeHandler{}.Apply(beta, registry)
	collector := &rollCollector{accept: true}

	ctx := hooks.RollContext{
		Phase: hooks.PhaseAttack, Actor: beta, Opponent: alpha, Attribute: string(domain.StatForce),
		OpponentAsset: alpha.Assets["bm"], World: "Krylos", Action: "Attack",
	}
	result := dispatch.RollWithHooks(ctx, domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: beta.Force}, registry, collector, &seqRoller{values: []int{4}}, beta, factionState, nil)

	if len(result.Dice) != 1 || collector.selections != 0 {
		t.Errorf("dice %v, selections %d; want one die and no prompt", result.Dice, collector.selections)
	}
	if beta.HookBudgets["tag:Warlike"] != 0 {
		t.Errorf("Warlike budget spent %d, want 0", beta.HookBudgets["tag:Warlike"])
	}
}
