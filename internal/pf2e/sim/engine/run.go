package engine

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const actionsPerTurn = 3

var (
	ErrDuplicateCombatant = errors.New("duplicate combatant id")
	ErrBadSetup           = errors.New("bad fight setup")
	ErrBadChoice          = errors.New("bad choice")
)

type Setup struct {
	Party     []*creature.Creature
	Enemies   []*creature.Creature
	Roller    dice.Roller
	Observer  Observer
	Decider   Decider
	MaxRounds int
}

func Run(setup Setup) (Outcome, error) {
	fight, err := newFight(setup)
	if err != nil {
		return NoOutcome, err
	}
	return fight.run()
}

func newFight(setup Setup) (*Fight, error) {
	switch {
	case len(setup.Party) == 0 || len(setup.Enemies) == 0:
		return nil, fmt.Errorf("%w: a fight needs at least one party member and one enemy", ErrBadSetup)
	case setup.MaxRounds < 1:
		return nil, fmt.Errorf("%w: max rounds %d: want at least 1", ErrBadSetup, setup.MaxRounds)
	case setup.Roller == nil || setup.Observer == nil || setup.Decider == nil:
		return nil, fmt.Errorf("%w: roller, observer and decider are all required", ErrBadSetup)
	}
	fight := &Fight{
		roller:    setup.Roller,
		observer:  setup.Observer,
		decider:   setup.Decider,
		maxRounds: setup.MaxRounds,
	}
	seen := map[string]bool{}
	for _, group := range []struct {
		side      Side
		creatures []*creature.Creature
	}{
		{SideParty, setup.Party},
		{SideEnemies, setup.Enemies},
	} {
		for _, source := range group.creatures {
			if seen[source.ID] {
				return nil, fmt.Errorf("%w: %s", ErrDuplicateCombatant, source.ID)
			}
			seen[source.ID] = true
			fight.combatants = append(fight.combatants, NewCombatant(group.side, source))
		}
	}
	fight.rollInitiative()
	return fight, nil
}

func (fight *Fight) rollInitiative() {
	for _, combatant := range fight.combatants {
		die := fight.roller.Die(20)
		combatant.Initiative = die + combatant.Creature.Perception
		fight.observer.Observe(InitiativeRolled{Combatant: combatant.ID, Die: die, Total: combatant.Initiative})
	}
	fight.order = slices.Clone(fight.combatants)
	slices.SortStableFunc(fight.order, func(first, second *Combatant) int {
		return cmp.Or(
			cmp.Compare(second.Initiative, first.Initiative),
			cmp.Compare(tieRank(first), tieRank(second)),
		)
	})
}

func tieRank(combatant *Combatant) int {
	if combatant.Creature.Kind == creature.KindNPC {
		return 0
	}
	return 1
}

func (fight *Fight) run() (Outcome, error) {
	for fight.round = 1; fight.round <= fight.maxRounds; fight.round++ {
		fight.observer.Observe(RoundStarted{Round: fight.round})
		for index := 0; index < len(fight.order); {
			actor := fight.order[index]
			if err := fight.takeTurn(actor); err != nil {
				return NoOutcome, err
			}
			if outcome, over := fight.outcome(); over {
				fight.observer.Observe(FightEnded{Rounds: fight.round, Outcome: outcome})
				return outcome, nil
			}
			// deliberate: the next actor is found from the actor's current place, not the old index, because a knock-down moves the target before the attacker mid-round.
			index = slices.Index(fight.order, actor) + 1
		}
	}
	fight.observer.Observe(FightEnded{Rounds: fight.maxRounds, Outcome: Stalemate})
	return Stalemate, nil
}

func (fight *Fight) takeTurn(actor *Combatant) error {
	if actor.Dead {
		return nil
	}
	fight.observer.Observe(TurnStarted{Round: fight.round, Combatant: actor.ID})
	if actor.Dying > 0 {
		fight.recoveryCheck(actor)
		return nil
	}
	if !actor.Standing() {
		return nil
	}
	actor.ActionsLeft = actionsPerTurn
	actor.AttacksThisTurn = 0
	actor.ReactionUsed = false
	for actor.ActionsLeft > 0 {
		choice, acting := fight.decider.Decide(fight, actor)
		if !acting {
			return nil
		}
		if err := fight.perform(actor, choice); err != nil {
			return err
		}
		if _, over := fight.outcome(); over {
			return nil
		}
	}
	return nil
}

func (fight *Fight) perform(actor *Combatant, choice Choice) error {
	if choice.Target == nil || !choice.Target.Standing() {
		return fmt.Errorf("%w: %s chose %q with no standing target", ErrBadChoice, actor.ID, choice.ID)
	}
	switch choice.Kind {
	case ChooseStrike:
		strike, found := actor.Creature.Strikes[choice.ID]
		if !found {
			return fmt.Errorf("%w: %s has no strike %q", ErrBadChoice, actor.ID, choice.ID)
		}
		actor.ActionsLeft--
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: choice.Target.ID})
		fight.strike(actor, choice.Target, strike)
		return nil
	case ChooseAbility:
		ability, found := actor.Creature.Abilities[choice.ID]
		if !found || ability.Ignore {
			return fmt.Errorf("%w: %s has no usable ability %q", ErrBadChoice, actor.ID, choice.ID)
		}
		if ability.Actions > actor.ActionsLeft {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d left", ErrBadChoice, actor.ID, choice.ID, ability.Actions, actor.ActionsLeft)
		}
		actor.ActionsLeft -= ability.Actions
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: ability.Actions, Target: choice.Target.ID})
		target := choice.Target
		for _, step := range ability.Steps {
			if !target.Standing() {
				next, found := fight.decider.NextTarget(fight, actor)
				if !found {
					return nil
				}
				target = next
			}
			fight.strike(actor, target, actor.Creature.Strikes[step.Strike])
		}
		return nil
	default:
		return fmt.Errorf("%w: %s made a choice of unknown kind %d", ErrBadChoice, actor.ID, choice.Kind)
	}
}

func (fight *Fight) outcome() (Outcome, bool) {
	partyStanding, enemiesStanding := false, false
	for _, combatant := range fight.combatants {
		if !combatant.Standing() {
			continue
		}
		if combatant.Side == SideParty {
			partyStanding = true
		} else {
			enemiesStanding = true
		}
	}
	switch {
	case !enemiesStanding:
		return PartyWins, true
	case !partyStanding:
		return EnemiesWin, true
	default:
		return NoOutcome, false
	}
}
