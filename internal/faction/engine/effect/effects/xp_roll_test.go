package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// exchangeConsulateReactor registers Exchange Consulate's XP roll for alpha with
// the given dice and returns its reactor.
func exchangeConsulateReactor(alpha *domain.Faction, dice ...int) hooks.MutationReactor {
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialXPRoll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-004", Name: "Exchange Consulate"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerGoalCompleted,
			Effect:  domain.SpecialXPRoll,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Goal: "G-007"},
			XPRoll:  &domain.XPRollParams{Die: domain.DiceRoll{NumDice: 1, Sides: 6}, Target: 4, Amount: 1},
		},
		Registry: registry,
		Roller:   &dicetest.Sequence{Values: dice},
	})
	return registry.MutationReactorsForFaction("alpha")[0].Hook
}

func TestXPRoll_RollsOnTheNamedGoal(t *testing.T) {
	bonus := []domain.Mutation{domain.XPAwarded{FactionID: "alpha", Amount: 1, Cause: "xp_roll", CausedByFactionID: "alpha"}}
	cases := []struct {
		name      string
		completed domain.GoalCompleted
		dice      []int
		want      []domain.Mutation
	}{
		{"Peaceable Kingdom, 4", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-007"}, []int{4}, bonus},
		{"Peaceable Kingdom, 3", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-007"}, []int{3}, nil},
		{"another goal rolls nothing", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-012"}, nil, nil},
		{"a rival's completion rolls nothing", domain.GoalCompleted{FactionID: "beta", GoalID: "G-007"}, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-004"}}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
			reactor := exchangeConsulateReactor(alpha, tc.dice...)

			if got := reactor.OnMutations([]domain.Mutation{tc.completed}, factionState, nil); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
