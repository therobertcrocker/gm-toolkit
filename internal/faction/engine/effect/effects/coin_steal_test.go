package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func coinStealEntry(amount domain.DiceRoll) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:   domain.TriggerOwnAttackHit,
		Effect:    domain.SpecialCoinSteal,
		Filters:   domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld},
		Gates:     domain.EffectGates{OncePer: domain.OncePerTargetFaction},
		CoinSteal: &domain.CoinStealParams{Amount: amount},
	}
}

// stealState holds thief "alpha" with two Franchise copies and two rival factions.
func stealState(betaCoin, gammaCoin int) *state.FactionState {
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Coin: 0, Assets: map[string]*domain.Asset{
			"fr-1": activeAsset("fr-1", "SWN-W1-001"),
			"fr-2": activeAsset("fr-2", "SWN-W1-001"),
		}},
		"beta":  {ID: "beta", Coin: betaCoin},
		"gamma": {ID: "gamma", Coin: gammaCoin},
	}}
}

// stealRulebook declares coin_steal on the Franchise with entry, and on the
// Blockade Fleet with a 1d4 steal.
func stealRulebook(entry domain.EffectDefinition) *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"SWN-W1-001": {ID: "SWN-W1-001", Effects: []domain.EffectDefinition{entry}},
		"SWN-F5-001": {ID: "SWN-F5-001", Effects: []domain.EffectDefinition{coinStealEntry(domain.DiceRoll{NumDice: 1, Sides: 4})}},
	}}
}

// stealReactors registers coin_steal for each of alpha's assets and returns the
// reactors in asset ID order, as dispatch would run them.
func stealReactors(t *testing.T, factionState *state.FactionState, stealBook *rulebook.Rulebook, roller domain.Roller) []hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	var reactors []hooks.MutationReactor
	for _, asset := range domain.SortedAssets(alpha) {
		def := stealBook.Assets[asset.DefinitionID]
		effects.Registrars()[domain.SpecialCoinSteal].Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: def.Effects[0], Registry: registry, Roller: roller,
		})
		for _, registered := range registry.MutationReactorsForAsset("alpha", asset.ID) {
			reactors = append(reactors, registered.Hook)
		}
	}
	return reactors
}

func runReactors(reactors []hooks.MutationReactor, mutations []domain.Mutation, factionState *state.FactionState) []domain.Mutation {
	return runReactorsWith(reactors, mutations, factionState, nil)
}

func runReactorsWith(reactors []hooks.MutationReactor, mutations []domain.Mutation, factionState *state.FactionState, reactorBook *rulebook.Rulebook) []domain.Mutation {
	var extra []domain.Mutation
	for _, reactor := range reactors {
		extra = append(extra, reactor.OnMutations(mutations, factionState, reactorBook)...)
	}
	return extra
}

func attackHit(targetFactionID, causedByAssetID string) domain.AssetHPDelta {
	return domain.AssetHPDelta{FactionID: targetFactionID, AssetID: targetFactionID + "-x", Delta: -2, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: causedByAssetID}
}

func steal(targetFactionID string, amount int) []domain.Mutation {
	return []domain.Mutation{
		domain.CoinDelta{FactionID: targetFactionID, Delta: -amount, Cause: "coin_steal", CausedByFactionID: "alpha"},
		domain.CoinDelta{FactionID: "alpha", Delta: amount, Cause: "coin_steal", CausedByFactionID: "alpha"},
	}
}

func TestCoinSteal(t *testing.T) {
	flatOne := domain.DiceRoll{Modifier: 1}
	oneDFour := domain.DiceRoll{NumDice: 1, Sides: 4}

	cases := []struct {
		name      string
		amount    domain.DiceRoll
		betaCoin  int
		gammaCoin int
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{
			name:      "an attack hit steals the amount",
			amount:    oneDFour,
			betaCoin:  5,
			mutations: []domain.Mutation{attackHit("beta", "fr-1")},
			want:      steal("beta", 3),
		},
		{
			name:     "a hit redirected to a Base still steals",
			amount:   flatOne,
			betaCoin: 5,
			mutations: []domain.Mutation{
				domain.BaseHPDelta{FactionID: "beta", BaseID: "beta-base", Delta: -2, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "fr-1"},
			},
			want: steal("beta", 1),
		},
		{
			name:     "a free attack on a new Base during Expand Influence steals",
			amount:   flatOne,
			betaCoin: 5,
			mutations: []domain.Mutation{
				domain.BaseHPDelta{FactionID: "beta", BaseID: "beta-base", Delta: -2, Cause: "expand", CausedByFactionID: "alpha", CausedByAssetID: "fr-1"},
			},
			want: steal("beta", 1),
		},
		{
			name:     "a counterattack hit does not steal",
			amount:   flatOne,
			betaCoin: 5,
			mutations: []domain.Mutation{
				domain.AssetHPDelta{FactionID: "beta", AssetID: "beta-x", Delta: -1, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: "fr-1", Counter: true},
			},
		},
		{
			name:      "a hit by another asset does not steal",
			amount:    flatOne,
			betaCoin:  5,
			mutations: []domain.Mutation{attackHit("beta", "someone-else")},
		},
		{
			name:      "a hit on the owner's own faction does not steal",
			amount:    flatOne,
			mutations: []domain.Mutation{attackHit("alpha", "fr-1")},
		},
		{
			name:      "the steal is capped at the target's Coin",
			amount:    oneDFour,
			betaCoin:  2,
			mutations: []domain.Mutation{attackHit("beta", "fr-1")},
			want:      steal("beta", 2),
		},
		{
			name:      "two copies hitting one target steal once",
			amount:    flatOne,
			betaCoin:  5,
			mutations: []domain.Mutation{attackHit("beta", "fr-1"), attackHit("beta", "fr-2")},
			want:      steal("beta", 1),
		},
		{
			name:      "two copies hitting two targets steal from each",
			amount:    flatOne,
			betaCoin:  5,
			gammaCoin: 5,
			mutations: []domain.Mutation{attackHit("beta", "fr-1"), attackHit("gamma", "fr-2")},
			want:      append(steal("beta", 1), steal("gamma", 1)...),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := stealState(tc.betaCoin, tc.gammaCoin)
			stealBook := stealRulebook(coinStealEntry(tc.amount))
			reactors := stealReactors(t, factionState, stealBook, fixedRoller{value: 3})
			got := runReactorsWith(reactors, tc.mutations, factionState, stealBook)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestCoinSteal_TwoDefinitionsDoNotTakeMoreThanTheTargetHas(t *testing.T) {
	factionState := stealState(1, 0)
	factionState.Factions["alpha"].Assets["bf-1"] = activeAsset("bf-1", "SWN-F5-001")
	stealBook := stealRulebook(coinStealEntry(domain.DiceRoll{Modifier: 1}))
	reactors := stealReactors(t, factionState, stealBook, fixedRoller{value: 3})

	got := runReactorsWith(reactors, []domain.Mutation{attackHit("beta", "fr-1"), attackHit("beta", "bf-1")}, factionState, stealBook)
	if !reflect.DeepEqual(got, steal("beta", 1)) {
		t.Errorf("mutations = %+v, want one steal of the target's last Coin", got)
	}
	fleetKey := "effect:coin_steal:own_attack_hit:SWN-F5-001:beta"
	if factionState.Factions["alpha"].BudgetSpent(fleetKey) {
		t.Errorf("the Blockade Fleet's budget is spent, but it took nothing from an empty purse")
	}
}

func TestCoinSteal_InactiveAssetStealsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := stealState(5, 0)
			stealBook := stealRulebook(coinStealEntry(domain.DiceRoll{Modifier: 1}))
			reactors := stealReactors(t, factionState, stealBook, fixedRoller{value: 3})
			deactivate(factionState.Factions["alpha"].Assets["fr-1"])

			if got := runReactorsWith(reactors, []domain.Mutation{attackHit("beta", "fr-1")}, factionState, stealBook); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
		})
	}
}

func TestCoinSteal_EmptyPurseLeavesTheBudgetUnspent(t *testing.T) {
	factionState := stealState(0, 0)
	stealBook := stealRulebook(coinStealEntry(domain.DiceRoll{Modifier: 1}))
	reactors := stealReactors(t, factionState, stealBook, fixedRoller{value: 3})

	if got := runReactorsWith(reactors, []domain.Mutation{attackHit("beta", "fr-1")}, factionState, stealBook); len(got) != 0 {
		t.Errorf("mutations = %+v, want none from an empty purse", got)
	}
	if budgets := factionState.Factions["alpha"].HookBudgets; len(budgets) != 0 {
		t.Errorf("alpha.HookBudgets = %v, want empty", budgets)
	}
}

func TestCoinSteal_SpentBudgetStopsALaterAction(t *testing.T) {
	factionState := stealState(5, 0)
	stealBook := stealRulebook(coinStealEntry(domain.DiceRoll{Modifier: 1}))
	reactors := stealReactors(t, factionState, stealBook, fixedRoller{value: 3})

	first := runReactorsWith(reactors, []domain.Mutation{attackHit("beta", "fr-1")}, factionState, stealBook)
	if !reflect.DeepEqual(first, steal("beta", 1)) {
		t.Fatalf("first batch = %+v, want one steal", first)
	}
	if second := runReactorsWith(reactors, []domain.Mutation{attackHit("beta", "fr-2")}, factionState, stealBook); len(second) != 0 {
		t.Errorf("second batch = %+v, want none: the target-faction budget is spent", second)
	}
	if len(factionState.Factions["beta"].HookBudgets) != 0 {
		t.Errorf("beta.HookBudgets = %v, want empty: the budget lives on the owner", factionState.Factions["beta"].HookBudgets)
	}
}
