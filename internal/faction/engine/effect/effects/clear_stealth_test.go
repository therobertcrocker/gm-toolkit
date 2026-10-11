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
