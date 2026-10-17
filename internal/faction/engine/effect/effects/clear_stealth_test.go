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

type seqRoller struct {
	values []int
	next   int
}

func (roller *seqRoller) Roll(int) int {
	value := roller.values[roller.next]
	roller.next++
	return value
}

var (
	krylos   = domain.Location{WorldID: "Krylos"}
	tartarus = domain.Location{WorldID: "Tartarus"}
	inFlight = domain.Location{}
)

func cellsAsset(location domain.Location) *domain.Asset {
	cells := activeAsset("cells", "SWN-C4-003")
	cells.Location = location
	return cells
}

// tripwireState puts alpha's Cells (Cunning 3) on Krylos; beta (Cunning 2) owns
// a stealthed spy and an open courier; alpha owns a stealthed agent of its own.
func tripwireState(cellsAt domain.Location) *state.FactionState {
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{
			"cells": cellsAsset(cellsAt),
			"agent": {ID: "agent", OwnerID: "alpha", Location: tartarus, Stealthy: true},
		}},
		"beta": {ID: "beta", Cunning: 2, Assets: map[string]*domain.Asset{
			"spy":     {ID: "spy", OwnerID: "beta", Location: tartarus, Stealthy: true},
			"courier": {ID: "courier", OwnerID: "beta", Location: tartarus},
		}},
	}}
}

func tripwireReactor(t *testing.T, factionState *state.FactionState, roller domain.Roller) hooks.MutationReactor {
	t.Helper()
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	stealthy := true
	effects.Registrars()[domain.SpecialClearStealth].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C4-003"}, Registry: registry, Roller: roller,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerStealthArrival,
			Effect:  domain.SpecialClearStealth,
			Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld, Stealthy: &stealthy},
			Gates: domain.EffectGates{Contest: &domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
				Roller:  domain.ContestRollerSelf,
			}},
		},
	})
	return registry.MutationReactorsForAsset("alpha", "cells")[0].Hook
}

func moved(factionID, assetID string, from, to domain.Location) domain.AssetMoved {
	return domain.AssetMoved{FactionID: factionID, AssetID: assetID, FromLocation: from, ToLocation: to, Cause: "move", CausedByFactionID: factionID}
}

var spyUnmasked = domain.AssetStealthCleared{FactionID: "beta", AssetID: "spy", Cause: "clear_stealth", CausedByFactionID: "alpha"}

func TestClearStealth(t *testing.T) {
	cases := []struct {
		name     string
		cellsAt  domain.Location
		rolls    []int
		mutation domain.Mutation
		want     []domain.Mutation
	}{
		{
			name:     "a stealthed rival landing loses its stealth when the Cells win",
			cellsAt:  krylos,
			rolls:    []int{6, 2}, // 6+3=9 vs 2+2=4
			mutation: moved("beta", "spy", tartarus, krylos),
			want:     []domain.Mutation{spyUnmasked},
		},
		{
			name:     "a landing from flight counts",
			cellsAt:  krylos,
			rolls:    []int{6, 2},
			mutation: moved("beta", "spy", inFlight, krylos),
			want:     []domain.Mutation{spyUnmasked},
		},
		{
			name:     "the asset keeps its stealth when the Cells lose",
			cellsAt:  krylos,
			rolls:    []int{1, 10}, // 1+3=4 vs 10+2=12
			mutation: moved("beta", "spy", tartarus, krylos),
		},
		{
			name:     "a tie goes to the defender",
			cellsAt:  krylos,
			rolls:    []int{5, 6}, // 5+3=8 vs 6+2=8
			mutation: moved("beta", "spy", tartarus, krylos),
		},
		{
			name:     "an open asset landing is ignored",
			cellsAt:  krylos,
			mutation: moved("beta", "courier", tartarus, krylos),
		},
		{
			name:     "the owner's own stealthed asset landing is ignored",
			cellsAt:  krylos,
			mutation: moved("alpha", "agent", tartarus, krylos),
		},
		{
			name:     "a landing on another world is ignored",
			cellsAt:  krylos,
			mutation: moved("beta", "spy", krylos, tartarus),
		},
		{
			name:     "a move into flight is not a landing",
			cellsAt:  krylos,
			mutation: moved("beta", "spy", tartarus, inFlight),
		},
		{
			name:     "a move within the Cells' world is not a landing",
			cellsAt:  krylos,
			mutation: moved("beta", "spy", krylos, krylos),
		},
		{
			name:     "Cells in flight watch no world",
			cellsAt:  inFlight,
			mutation: moved("beta", "spy", tartarus, inFlight),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := tripwireState(tc.cellsAt)
			roller := &seqRoller{values: tc.rolls}
			got := tripwireReactor(t, factionState, roller).OnMutations([]domain.Mutation{tc.mutation}, factionState, nil)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
			if roller.next != len(tc.rolls) {
				t.Errorf("rolled %d dice, want %d", roller.next, len(tc.rolls))
			}
		})
	}
}

// stealthedPurchase is the batch an on_purchase apply_stealth leaves: the asset is
// added, then stealthed. Neither has applied, so the asset is not in state.
func stealthedPurchase(factionID, assetID string, at domain.Location) []domain.Mutation {
	return []domain.Mutation{
		domain.AssetAdded{FactionID: factionID, Asset: domain.Asset{ID: assetID, OwnerID: factionID, Location: at}, Cause: "buy", CausedByFactionID: factionID},
		domain.AssetStealthApplied{FactionID: factionID, AssetID: assetID, Cause: "apply_stealth", CausedByFactionID: factionID},
	}
}

func TestClearStealth_Purchase(t *testing.T) {
	cases := []struct {
		name    string
		cellsAt domain.Location
		rolls   []int
		batch   []domain.Mutation
		want    []domain.Mutation
	}{
		{
			name:    "a rival bought stealthed on the Cells' world loses its stealth when the Cells win",
			cellsAt: krylos,
			rolls:   []int{6, 2}, // 6+3=9 vs 2+2=4
			batch:   stealthedPurchase("beta", "assassins", krylos),
			want:    []domain.Mutation{domain.AssetStealthCleared{FactionID: "beta", AssetID: "assassins", Cause: "clear_stealth", CausedByFactionID: "alpha"}},
		},
		{
			name:    "a purchase on another world is ignored",
			cellsAt: krylos,
			batch:   stealthedPurchase("beta", "assassins", tartarus),
		},
		{
			name:    "stealth applied to an asset already on the world is not an arrival",
			cellsAt: tartarus,
			batch:   []domain.Mutation{domain.AssetStealthApplied{FactionID: "beta", AssetID: "courier", Cause: "buy", CausedByFactionID: "beta"}},
		},
		{
			name:    "the owner's own purchase is ignored",
			cellsAt: krylos,
			batch:   stealthedPurchase("alpha", "assassins", krylos),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := tripwireState(tc.cellsAt)
			roller := &seqRoller{values: tc.rolls}
			got := tripwireReactor(t, factionState, roller).OnMutations(tc.batch, factionState, nil)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
			if roller.next != len(tc.rolls) {
				t.Errorf("rolled %d dice, want %d", roller.next, len(tc.rolls))
			}
		})
	}
}

func TestClearStealth_InactiveCellsWatchNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			factionState := tripwireState(krylos)
			roller := &seqRoller{}
			reactor := tripwireReactor(t, factionState, roller)
			deactivate(factionState.Factions["alpha"].Assets["cells"])

			if got := reactor.OnMutations([]domain.Mutation{moved("beta", "spy", tartarus, krylos)}, factionState, nil); len(got) != 0 {
				t.Errorf("mutations = %+v, want none", got)
			}
			if roller.next != 0 {
				t.Errorf("rolled %d dice, want 0", roller.next)
			}
		})
	}
}

func TestClearStealth_OpenAssetIgnoredWithoutStealthyFilter(t *testing.T) {
	unfiltered := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld}

	t.Run("stealth_arrival", func(t *testing.T) {
		factionState := tripwireState(krylos)
		alpha := factionState.Factions["alpha"]
		roller := &seqRoller{}
		registry := hooks.NewRegistry()
		effects.Registrars()[domain.SpecialClearStealth].Register(effect.RegisterContext{
			Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C4-003"}, Registry: registry, Roller: roller,
			Entry: domain.EffectDefinition{Trigger: domain.TriggerStealthArrival, Effect: domain.SpecialClearStealth, Filters: unfiltered},
		})
		reactor := registry.MutationReactorsForAsset("alpha", "cells")[0].Hook

		if got := reactor.OnMutations([]domain.Mutation{moved("beta", "courier", tartarus, krylos)}, factionState, nil); len(got) != 0 {
			t.Errorf("mutations = %+v, want none", got)
		}
	})

	t.Run("turn_start", func(t *testing.T) {
		factionState := tripwireState(tartarus)
		delete(factionState.Factions["beta"].Assets, "spy")
		alpha := factionState.Factions["alpha"]
		roller := &seqRoller{}
		got := effects.Registrars()[domain.SpecialClearStealth].Fire(effect.FireContext{
			Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C8-001"}, Roller: roller, FactionState: factionState,
			Entry: domain.EffectDefinition{Trigger: domain.TriggerTurnStart, Effect: domain.SpecialClearStealth, Filters: unfiltered, Gates: domain.EffectGates{Contest: matrixContest}},
		})
		if len(got) != 0 {
			t.Errorf("mutations = %+v, want none", got)
		}
		if roller.next != 0 {
			t.Errorf("rolled %d dice, want 0", roller.next)
		}
	})
}

var matrixContest = &domain.EffectContest{
	Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
	Roller:  domain.ContestRollerTarget,
}

// fireMatrix runs Panopticon Matrix's turn_start entry from alpha's "cells" asset,
// which stands in for the Matrix: Fire reads the entry, not the asset's definition.
func fireMatrix(factionState *state.FactionState, contest *domain.EffectContest, roller domain.Roller) []domain.Mutation {
	alpha := factionState.Factions["alpha"]
	stealthy := true
	return effects.Registrars()[domain.SpecialClearStealth].Fire(effect.FireContext{
		Faction: alpha, Asset: alpha.Assets["cells"], Def: &domain.AssetDefinition{ID: "SWN-C8-001"}, Roller: roller, FactionState: factionState,
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerTurnStart,
			Effect:  domain.SpecialClearStealth,
			Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeWorld, Stealthy: &stealthy},
			Gates:   domain.EffectGates{Contest: contest},
		},
	})
}

func TestClearStealth_TurnStart(t *testing.T) {
	cases := []struct {
		name     string
		matrixAt domain.Location
		setup    func(*state.FactionState)
		contest  *domain.EffectContest
		rolls    []int
		want     []domain.Mutation
	}{
		{
			name:     "a stealthed rival that fails its test is unmasked",
			matrixAt: tartarus,
			contest:  matrixContest,
			rolls:    []int{2, 9}, // beta 2+2=4 vs alpha 9+3=12
			want:     []domain.Mutation{spyUnmasked},
		},
		{
			name:     "a stealthed rival that passes its test keeps stealth",
			matrixAt: tartarus,
			contest:  matrixContest,
			rolls:    []int{9, 2}, // beta 9+2=11 vs alpha 2+3=5
		},
		{
			name:     "a tie goes to the Matrix, so the rival's test fails",
			matrixAt: tartarus,
			contest:  matrixContest,
			rolls:    []int{3, 2}, // beta 3+2=5 vs alpha 2+3=5
			want:     []domain.Mutation{spyUnmasked},
		},
		{
			name:     "an open rival asset is not tested",
			matrixAt: tartarus,
			setup: func(factionState *state.FactionState) {
				delete(factionState.Factions["beta"].Assets, "spy")
				delete(factionState.Factions["alpha"].Assets, "agent")
			},
			contest: matrixContest,
		},
		{
			name:     "the owner's own stealthed asset is not tested",
			matrixAt: tartarus,
			setup: func(factionState *state.FactionState) {
				delete(factionState.Factions["beta"].Assets, "spy")
				delete(factionState.Factions["beta"].Assets, "courier")
			},
			contest: matrixContest,
		},
		{
			name:     "a rival on another world is not tested",
			matrixAt: krylos,
			contest:  matrixContest,
		},
		{
			name:     "a Matrix in flight tests nothing",
			matrixAt: inFlight,
			contest:  matrixContest,
		},
		{
			name:     "without a contest gate the rival is unmasked",
			matrixAt: tartarus,
			want:     []domain.Mutation{spyUnmasked},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factionState := tripwireState(tc.matrixAt)
			if tc.setup != nil {
				tc.setup(factionState)
			}
			roller := &seqRoller{values: tc.rolls}
			got := fireMatrix(factionState, tc.contest, roller)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mutations = %+v, want %+v", got, tc.want)
			}
			if roller.next != len(tc.rolls) {
				t.Errorf("rolled %d dice, want %d", roller.next, len(tc.rolls))
			}
		})
	}
}
