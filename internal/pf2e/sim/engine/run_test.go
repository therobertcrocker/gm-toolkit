package engine

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

func loadTestCreatures(t *testing.T, ids ...string) []*creature.Creature {
	t.Helper()
	creatures := make([]*creature.Creature, len(ids))
	for index, id := range ids {
		loaded, err := creature.Load("testdata/" + id)
		if err != nil {
			t.Fatalf("load %s: %v", id, err)
		}
		creatures[index] = loaded
	}
	return creatures
}

func runLunaFight(t *testing.T, seed uint64) (Outcome, []Event) {
	t.Helper()
	events := &recorder{}
	outcome, err := Run(Setup{
		Party:     loadTestCreatures(t, "dummy-fighter", "dummy-alchemist", "dummy-sorcerer"),
		Enemies:   loadTestCreatures(t, "abyssal-luna"),
		Roller:    dice.NewSeeded(seed),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 20,
	})
	if err != nil {
		t.Fatalf("Run(seed %d): %v", seed, err)
	}
	return outcome, events.events
}

func TestRunIsDeterministic(t *testing.T) {
	firstSeedEvents := []Event(nil)
	seedsDiffer := false
	for seed := uint64(1); seed <= 10; seed++ {
		outcome, events := runLunaFight(t, seed)
		againOutcome, againEvents := runLunaFight(t, seed)
		if outcome != againOutcome || !reflect.DeepEqual(events, againEvents) {
			t.Fatalf("seed %d gave two different fights", seed)
		}
		ended, isEnd := events[len(events)-1].(FightEnded)
		if !isEnd || ended.Outcome != outcome {
			t.Errorf("seed %d: last event = %+v, want FightEnded with %v", seed, events[len(events)-1], outcome)
		}
		if firstSeedEvents == nil {
			firstSeedEvents = events
		} else if !reflect.DeepEqual(events, firstSeedEvents) {
			seedsDiffer = true
		}
	}
	if !seedsDiffer {
		t.Error("seeds 1 to 10 all gave the same fight")
	}
}

func TestInitiativeTieGoesToTheNPC(t *testing.T) {
	pc := testCombatant("pc", SideParty, creature.KindPC, 30, 20).Creature
	npc := testCombatant("npc", SideEnemies, creature.KindNPC, 30, 20).Creature
	fight, err := newFight(Setup{
		Party:     []*creature.Creature{pc},
		Enemies:   []*creature.Creature{npc},
		Roller:    script(t, 10, 10),
		Observer:  &recorder{},
		Decider:   FirstFitting{},
		MaxRounds: 1,
	})
	if err != nil {
		t.Fatalf("newFight: %v", err)
	}
	if want := []string{"npc", "pc"}; !slices.Equal(orderIDs(fight), want) {
		t.Errorf("order = %v, want %v", orderIDs(fight), want)
	}
}

func TestFirstFittingTakesAbilityThenStrike(t *testing.T) {
	luna := loadTestCreatures(t, "abyssal-luna")[0]
	fighter := loadTestCreatures(t, "dummy-fighter")[0]
	actor := NewCombatant(SideEnemies, luna)
	foe := NewCombatant(SideParty, fighter)
	fight := &Fight{combatants: []*Combatant{foe, actor}}

	actor.ActionsLeft = 3
	choice, acting := FirstFitting{}.Decide(fight, actor)
	if !acting || choice.Kind != ChooseAbility || choice.ID != "abyssal-frenzy" || choice.Target != foe {
		t.Errorf("with 3 actions: %+v, %v, want abyssal-frenzy at the fighter", choice, acting)
	}
	actor.ActionsLeft = 1
	choice, acting = FirstFitting{}.Decide(fight, actor)
	if !acting || choice.Kind != ChooseStrike || choice.ID != "slash" {
		t.Errorf("with 1 action: %+v, %v, want the slash Strike", choice, acting)
	}
}

func TestRoundLimitIsAStalemate(t *testing.T) {
	wall := func(id string, kind creature.Kind) *creature.Creature {
		return testCombatant(id, SideParty, kind, 30, 100).Creature
	}
	events := &recorder{}
	outcome, err := Run(Setup{
		Party:     []*creature.Creature{wall("left", creature.KindPC)},
		Enemies:   []*creature.Creature{wall("right", creature.KindNPC)},
		Roller:    dice.NewSeeded(1),
		Observer:  events,
		Decider:   FirstFitting{},
		MaxRounds: 2,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if outcome != Stalemate {
		t.Errorf("outcome = %v, want stalemate", outcome)
	}
	if last := events.events[len(events.events)-1]; last != (FightEnded{Rounds: 2, Outcome: Stalemate}) {
		t.Errorf("last event = %+v, want FightEnded after 2 rounds", last)
	}
}

func TestRunRejectsBadSetups(t *testing.T) {
	fighter := loadTestCreatures(t, "dummy-fighter")[0]
	luna := loadTestCreatures(t, "abyssal-luna")[0]
	valid := Setup{
		Party:     []*creature.Creature{fighter},
		Enemies:   []*creature.Creature{luna},
		Roller:    dice.NewSeeded(1),
		Observer:  &recorder{},
		Decider:   FirstFitting{},
		MaxRounds: 5,
	}
	duplicate := valid
	duplicate.Enemies = []*creature.Creature{luna, luna}
	noRounds := valid
	noRounds.MaxRounds = 0
	noEnemies := valid
	noEnemies.Enemies = nil

	if _, err := Run(duplicate); !errors.Is(err, ErrDuplicateCombatant) {
		t.Errorf("duplicate ids: err = %v, want ErrDuplicateCombatant", err)
	}
	if outcome, err := Run(noRounds); !errors.Is(err, ErrBadSetup) || outcome != NoOutcome {
		t.Errorf("no rounds: outcome, err = %v, %v, want no outcome, ErrBadSetup", outcome, err)
	}
	if outcome, err := Run(noEnemies); !errors.Is(err, ErrBadSetup) || outcome != NoOutcome {
		t.Errorf("no enemies: outcome, err = %v, %v, want no outcome, ErrBadSetup", outcome, err)
	}
}
