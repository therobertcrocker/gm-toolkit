package effects

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// rollRecorder is a Cat 1 hook that records each roll it sees and offers nothing.
type rollRecorder struct {
	seen []hooks.RollContext
}

func (recorder *rollRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.seen = append(recorder.seen, ctx)
	return nil
}

func cunningContest(roller domain.ContestRoller, roll domain.ContestRoll) domain.EffectContest {
	return domain.EffectContest{
		Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning, Roll: roll},
		Roller:  roller,
	}
}

func TestContestPasses(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 3}
	target := &domain.Faction{ID: "beta", Cunning: 2}
	cases := []struct {
		name   string
		roller domain.ContestRoller
		rolls  []int
		want   bool
	}{
		{"self: the declaring faction wins", domain.ContestRollerSelf, []int{6, 2}, true},
		{"self: the declaring faction loses", domain.ContestRollerSelf, []int{1, 10}, false},
		{"self: a tie fails", domain.ContestRollerSelf, []int{5, 6}, false},
		{"target: the target fails its test", domain.ContestRollerTarget, []int{1, 10}, true},
		{"target: the target passes its test", domain.ContestRollerTarget, []int{10, 1}, false},
		{"target: a tie passes the effect", domain.ContestRollerTarget, []int{6, 5}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roll := contestRoll{
				contest:   cunningContest(tc.roller, domain.ContestRollTest),
				declaring: declaring,
				target:    target,
				roller:    &dicetest.Sequence{Values: tc.rolls},
			}
			if got := contestPasses(roll); got != tc.want {
				t.Errorf("contestPasses = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContestPasses_RollPhasesAndSides(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 3, Wealth: 1}
	target := &domain.Faction{ID: "beta", Cunning: 2, Wealth: 4}
	cells := &domain.Asset{ID: "cells"}
	spy := &domain.Asset{ID: "spy"}
	cases := []struct {
		name            string
		roll            domain.ContestRoll
		attack, defense hooks.RollPhase
	}{
		{"an attack", domain.ContestRollAttack, hooks.PhaseAttack, hooks.PhaseDefense},
		{"a test", domain.ContestRollTest, hooks.PhaseFactionTest, hooks.PhaseContested},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &rollRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)
			contest := domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth, Roll: tc.roll},
				Roller:  domain.ContestRollerSelf,
			}

			contestPasses(contestRoll{
				contest: contest, declaring: declaring, declaringAsset: cells, target: target, targetAsset: spy,
				world: "Krylos", registry: registry, roller: &dicetest.Sequence{Values: []int{5, 5}},
			})

			want := []hooks.RollContext{
				{Phase: tc.attack, Actor: declaring, Opponent: target, Attribute: "Cunning", OpponentAttribute: "Wealth", Asset: cells, OpponentAsset: spy, World: "Krylos"},
				{Phase: tc.defense, Actor: target, Opponent: declaring, Attribute: "Wealth", OpponentAttribute: "Cunning", Asset: spy, OpponentAsset: cells, World: "Krylos"},
			}
			if !reflect.DeepEqual(recorder.seen, want) {
				t.Errorf("rolls = %+v, want %+v", recorder.seen, want)
			}
		})
	}
}

func TestContestPasses_AttackTiesAskTheTieResolvers(t *testing.T) {
	cases := []struct {
		name      string
		fanatical string
		roll      domain.ContestRoll
		want      bool
	}{
		{"an attack: a Fanatical defender loses the tie", "beta", domain.ContestRollAttack, true},
		{"an attack: a Fanatical attacker loses the tie", "alpha", domain.ContestRollAttack, false},
		{"a test: a Fanatical defender still wins the tie", "beta", domain.ContestRollTest, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			declaring := &domain.Faction{ID: "alpha", Cunning: 1}
			target := &domain.Faction{ID: "beta", Cunning: 1}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": declaring, "beta": target}}
			holder := factionState.Factions[tc.fanatical]
			holder.Tags = []string{"F"}
			registry := hooks.NewRegistry()
			loseTiesRegistrar.Register(effect.RegisterContext{
				Holder: effect.Holder{Faction: holder, Tag: &domain.Tag{ID: "F", Name: "Fanatical"}},
				Entry: domain.EffectDefinition{
					Trigger: domain.TriggerAttackTie,
					Effect:  domain.SpecialLoseTies,
					Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
				},
				Registry: registry,
			})
			roll := contestRoll{
				contest:      cunningContest(domain.ContestRollerSelf, tc.roll),
				declaring:    declaring,
				target:       target,
				registry:     registry,
				factionState: factionState,
				roller:       &dicetest.Sequence{Values: []int{5, 5}},
			}

			if got := contestPasses(roll); got != tc.want {
				t.Errorf("contestPasses = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContestPasses_TakesOnlyFreeDice(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 1, Tags: []string{"T"}, HookBudgets: map[string]int{}}
	target := &domain.Faction{ID: "beta", Cunning: 1}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": declaring, "beta": target}}
	registry := hooks.NewRegistry()
	free := domain.EffectDefinition{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}
	limited := free
	limited.Gates = domain.EffectGates{OncePer: domain.OncePerOwnerFaction}
	for _, entry := range []domain.EffectDefinition{free, limited} {
		bonusDieRegistrar.Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: declaring, Tag: &domain.Tag{ID: "T", Name: "Tag"}}, Entry: entry, Registry: registry,
		})
	}
	roll := contestRoll{
		contest:      cunningContest(domain.ContestRollerSelf, domain.ContestRollTest),
		declaring:    declaring,
		target:       target,
		registry:     registry,
		factionState: factionState,
		// alpha rolls 2 and its free die 9, keeping 9; beta rolls 5.
		roller: &dicetest.Sequence{Values: []int{2, 9, 5}},
	}

	if !contestPasses(roll) {
		t.Error("contestPasses = false, want true: the free die's 9 beats beta's 5")
	}
	if len(declaring.HookBudgets) != 0 {
		t.Errorf("HookBudgets = %v, want none spent: the once-per-turn die waits", declaring.HookBudgets)
	}
}

func TestContestPasses_DetectionMarksTheDeclaringSide(t *testing.T) {
	cases := []struct {
		name   string
		roller domain.ContestRoller
		want   []bool
	}{
		{"self: the declaring side attacks", domain.ContestRollerSelf, []bool{true, false}},
		{"target: the declaring side defends", domain.ContestRollerTarget, []bool{false, true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &rollRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)

			contestPasses(contestRoll{
				contest:          cunningContest(tc.roller, domain.ContestRollTest),
				declaringDetects: true,
				declaring:        &domain.Faction{ID: "alpha"},
				target:           &domain.Faction{ID: "beta"},
				registry:         registry,
				roller:           &dicetest.Sequence{Values: []int{5, 5}},
			})

			var got []bool
			for _, ctx := range recorder.seen {
				got = append(got, ctx.Detection)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Detection per roll = %v, want %v", got, tc.want)
			}
		})
	}
}
