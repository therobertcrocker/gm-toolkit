package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const bankDefID = "SWN-W4-004"

// bankState gives "alpha" the named Banks plus one unrelated asset.
func bankState(bankIDs ...string) *state.FactionState {
	assets := map[string]*domain.Asset{
		"aaa-other": {ID: "aaa-other", DefinitionID: "SWN-F1-001", OwnerID: "alpha"},
	}
	for _, id := range bankIDs {
		assets[id] = activeAsset(id, bankDefID)
	}
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Coin: 10, Assets: assets},
		"beta":  {ID: "beta"},
	}}
}

// bankReactors registers every Bank of alpha's in asset ID order, as dispatch runs them.
func bankReactors(t *testing.T, factionState *state.FactionState, oncePer domain.OncePerKey) []hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	def := &domain.AssetDefinition{ID: bankDefID}
	entry := domain.EffectDefinition{
		Trigger: domain.TriggerCoinLoss,
		Effect:  domain.SpecialIgnoreCoinLoss,
		Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction},
		Gates:   domain.EffectGates{OncePer: oncePer},
	}
	var reactors []hooks.MutationReactor
	for _, asset := range domain.SortedAssets(alpha) {
		if asset.DefinitionID != bankDefID {
			continue
		}
		effects.Registrars()[domain.SpecialIgnoreCoinLoss].Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: alpha, Asset: asset, Def: def}, Entry: entry, Registry: registry,
		})
		for _, registered := range registry.MutationReactorsForAsset("alpha", asset.ID) {
			reactors = append(reactors, registered.Hook)
		}
	}
	return reactors
}

func lossTo(factionID string, delta int, causedBy string) domain.CoinDelta {
	return domain.CoinDelta{FactionID: factionID, Delta: delta, Cause: "coin_steal", CausedByFactionID: causedBy}
}

func absorbed(amount int) domain.CoinDelta {
	return domain.CoinDelta{FactionID: "alpha", Delta: amount, Cause: "bank_absorbed", CausedByFactionID: "alpha"}
}

func TestIgnoreCoinLoss(t *testing.T) {
	cases := []struct {
		name      string
		banks     []string
		mutations []domain.Mutation
		want      []domain.Mutation
	}{
		{
			name:      "a loss another faction imposed is paid back",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("alpha", -3, "beta")},
			want:      []domain.Mutation{absorbed(3)},
		},
		{
			name:  "a purchase surcharge another faction imposed is paid back, the price is not",
			banks: []string{"bank-1"},
			mutations: []domain.Mutation{
				domain.CoinDelta{FactionID: "alpha", Delta: -4, Cause: "buy", CausedByFactionID: "alpha"},
				domain.CoinDelta{FactionID: "alpha", Delta: -1, Cause: "cost_surcharge", CausedByFactionID: "beta"},
			},
			want: []domain.Mutation{absorbed(1)},
		},
		{
			name:  "a Transport Lockdown fee is paid back, the hold is not",
			banks: []string{"bank-1"},
			mutations: []domain.Mutation{
				domain.CoinDelta{FactionID: "alpha", Delta: -3, Cause: "transport_lockdown", CausedByFactionID: "beta"},
				domain.MovementOrderHeld{FactionID: "alpha", AssetID: "t1", WorldID: "Krylos", Reason: domain.HoldLockdownPaid, Cause: domain.CauseMovementTick, CausedByFactionID: "alpha"},
			},
			want: []domain.Mutation{absorbed(3)},
		},
		{
			name:      "a loss the owner imposed on itself is not absorbed",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("alpha", -3, "alpha")},
		},
		{
			name:      "a loss with no imposing faction is not absorbed",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("alpha", -3, "")},
		},
		{
			name:      "a gain is not absorbed",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("alpha", 3, "beta")},
		},
		{
			name:      "another faction's loss is not absorbed",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("beta", -3, "alpha")},
		},
		{
			name:      "one Bank absorbs one loss of two",
			banks:     []string{"bank-1"},
			mutations: []domain.Mutation{lossTo("alpha", -1, "beta"), lossTo("alpha", -2, "beta")},
			want:      []domain.Mutation{absorbed(1)},
		},
		{
			name:      "two Banks absorb two losses, each once",
			banks:     []string{"bank-1", "bank-2"},
			mutations: []domain.Mutation{lossTo("alpha", -1, "beta"), lossTo("alpha", -2, "beta")},
			want:      []domain.Mutation{absorbed(1), absorbed(2)},
		},
		{
			name:      "two Banks absorb a single loss once",
			banks:     []string{"bank-1", "bank-2"},
			mutations: []domain.Mutation{lossTo("alpha", -3, "beta")},
			want:      []domain.Mutation{absorbed(3)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := bankState(tc.banks...)
			got := runReactors(bankReactors(t, factionState, domain.OncePerAsset), tc.mutations, factionState)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestIgnoreCoinLoss_InactiveBankAbsorbsNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := bankState("bank-1")
			reactors := bankReactors(t, factionState, domain.OncePerAsset)
			deactivate(factionState.Factions["alpha"].Assets["bank-1"])

			if got := runReactors(reactors, []domain.Mutation{lossTo("alpha", -3, "beta")}, factionState); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
		})
	}
}

func TestIgnoreCoinLoss_InactiveLowestBankHandsTheLeadOn(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := bankState("bank-1", "bank-2", "bank-3")
			reactors := bankReactors(t, factionState, domain.OncePerAsset)
			deactivate(factionState.Factions["alpha"].Assets["bank-1"])

			batch := []domain.Mutation{lossTo("alpha", -1, "beta"), lossTo("alpha", -2, "beta"), lossTo("alpha", -4, "beta")}
			got := runReactors(reactors, batch, factionState)
			if want := []domain.Mutation{absorbed(1), absorbed(2)}; !reflect.DeepEqual(got, want) {
				t.Errorf("mutations = %+v, want %+v: two active Banks absorb two losses", got, want)
			}
			budgets := factionState.Factions["alpha"].HookBudgets
			if _, spent := budgets["effect:ignore_coin_loss:coin_loss:"+bankDefID+":bank-1"]; spent {
				t.Error("the inactive Bank's budget was spent")
			}
			if budgets["effect:ignore_coin_loss:coin_loss:"+bankDefID+":bank-2"] != 1 {
				t.Errorf("bank-2's budget is not spent under the expected key; budgets = %v", budgets)
			}
		})
	}
}

func TestIgnoreCoinLoss_EachBankAbsorbsOncePerTurn(t *testing.T) {
	factionState := bankState("bank-1", "bank-2")
	reactors := bankReactors(t, factionState, domain.OncePerAsset)
	batch := []domain.Mutation{lossTo("alpha", -1, "beta")}

	for round, want := range [][]domain.Mutation{{absorbed(1)}, {absorbed(1)}, nil} {
		if got := runReactors(reactors, batch, factionState); !reflect.DeepEqual(got, want) {
			t.Errorf("batch %d: mutations = %+v, want %+v", round+1, got, want)
		}
	}
}

func TestIgnoreCoinLoss_OwnerFactionBudgetIsSharedByEveryBank(t *testing.T) {
	factionState := bankState("bank-1", "bank-2")
	reactors := bankReactors(t, factionState, domain.OncePerOwnerFaction)

	got := runReactors(reactors, []domain.Mutation{lossTo("alpha", -1, "beta"), lossTo("alpha", -2, "beta")}, factionState)
	if want := []domain.Mutation{absorbed(1)}; !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}
