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
	ErrTriggerLoop        = errors.New("trigger loop")
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
		combatant.Initiative = die + combatant.Creature.Perception + fight.modifierTotal(combatant, []string{"perception", "initiative"}, nil)
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
	for _, combatant := range slices.Clone(fight.order) {
		fight.fire(combatant, TriggerEvent{Kind: creature.TriggerInitiative})
	}
	if fight.fault != nil {
		return NoOutcome, fight.fault
	}
	for fight.round = 1; fight.round <= fight.maxRounds; fight.round++ {
		fight.observer.Observe(RoundStarted{Round: fight.round})
		for index := 0; index < len(fight.order); {
			actor := fight.order[index]
			if err := fight.takeTurn(actor); err != nil {
				return NoOutcome, err
			}
			if fight.fault != nil {
				return NoOutcome, fight.fault
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
	fight.current = actor
	fight.startTurn(actor)
	err := fight.act(actor)
	if err == nil {
		fight.endTurn(actor)
	}
	fight.current = nil
	return err
}

func (fight *Fight) act(actor *Combatant) error {
	if actor.Dead {
		return nil
	}
	fight.observer.Observe(TurnStarted{Round: fight.round, Combatant: actor.ID})
	actor.ReactionUsed = false
	fight.tickLimits(actor)
	healedWhileDown := !actor.Standing()
	// deliberate: a downed creature's fast healing comes before its recovery check, not after, because the start-of-turn steps run in any order and healing first is never worse (Decision Record — Execution 22).
	if healedWhileDown {
		fight.fastHeal(actor)
	}
	if actor.Dying > 0 {
		fight.recoveryCheck(actor)
		if !actor.Dead {
			fight.regainActions(actor)
		}
		return nil
	}
	if !actor.Standing() {
		fight.regainActions(actor)
		return nil
	}
	actor.AttacksThisTurn = 0
	actor.lastPenalty = 0
	fight.regainActions(actor)
	// deliberate: turn-start triggers see the HP the turn starts with, not the HP after fast healing, because Robert ruled Last Stand first (Decision 4).
	fight.fire(actor, TriggerEvent{Kind: creature.TriggerTurnStart})
	if !healedWhileDown {
		fight.fastHeal(actor)
	}
	for actor.ActionsLeft > 0 && actor.Standing() {
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
	if choice.Kind == ChooseStand {
		if actor.Condition("prone") == 0 || actor.OpenActions() < 1 {
			return fmt.Errorf("%w: %s cannot Stand", ErrBadChoice, actor.ID)
		}
		actor.ActionsLeft--
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: "stand", Actions: 1})
		if fight.provoke(actor, "stand", TriggerEvent{Traits: []string{"move"}}) {
			fight.stand(actor)
		}
		return nil
	}
	if choice.Kind == ChooseSustain {
		effect, found := actor.sustained[choice.ID]
		if !found || effect.done || actor.OpenActions() < 1 {
			return fmt.Errorf("%w: %s cannot Sustain %q", ErrBadChoice, actor.ID, choice.ID)
		}
		ability := actor.Creature.Abilities[choice.ID]
		actor.ActionsLeft--
		effect.done = true
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: targetID(effect.target)})
		fight.observer.Observe(Sustained{Combatant: actor.ID, Ability: choice.ID})
		ability.Steps = ability.Sustain
		fight.useAbility(abilityUse{actor: actor, ability: ability, target: effect.target, actions: 1, fixed: true})
		return nil
	}
	switch choice.Kind {
	case ChooseStrike:
		strike, found := actor.Creature.Strikes[choice.ID]
		if !found {
			return fmt.Errorf("%w: %s has no strike %q", ErrBadChoice, actor.ID, choice.ID)
		}
		if choice.Target == nil || !choice.Target.Standing() {
			return fmt.Errorf("%w: %s chose %q with no standing target", ErrBadChoice, actor.ID, choice.ID)
		}
		actor.ActionsLeft--
		actor.StrikeOnlyActions = max(actor.StrikeOnlyActions-1, 0)
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: choice.Target.ID})
		fight.strike(actor, choice.Target, strike, strikeOptions{})
		return nil
	case ChooseAbility:
		ability, found := actor.Creature.Abilities[choice.ID]
		if !found || ability.Ignore || ability.Trigger != nil {
			return fmt.Errorf("%w: %s has no usable ability %q", ErrBadChoice, actor.ID, choice.ID)
		}
		if !fight.checkTarget(actor, ability, choice.Target) {
			return fmt.Errorf("%w: %s chose %q with no fitting target", ErrBadChoice, actor.ID, choice.ID)
		}
		cost := ability.Actions
		if choice.Actions != 0 {
			cost = choice.Actions
		}
		low, high := ability.CostRange()
		if cost < low || cost > high {
			return fmt.Errorf("%w: %s chose %s for %d actions; it takes %d to %d", ErrBadChoice, actor.ID, choice.ID, cost, low, high)
		}
		if cost > actor.OpenActions() {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d open", ErrBadChoice, actor.ID, choice.ID, cost, actor.OpenActions())
		}
		if reason := actor.Blocked(ability); reason != "" {
			return fmt.Errorf("%w: %s chose %s: %s", ErrBadChoice, actor.ID, choice.ID, reason)
		}
		actor.ActionsLeft -= cost
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: cost, Target: targetID(choice.Target)})
		fight.spend(actor, ability)
		if fight.provoke(actor, ability.ID, TriggerEvent{Ability: ability.ID, Traits: ability.Traits}) {
			fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
			if len(ability.Sustain) > 0 && actor.Standing() {
				actor.sustained[ability.ID] = &sustainedEffect{target: choice.Target, done: true}
			}
			if slices.Contains(ability.Tags, creature.TagLoyalty) {
				fight.loyalty(actor, ability.ID)
			}
		}
		if ability.TurnEnds {
			actor.ActionsLeft = 0
			actor.StrikeOnlyActions = 0
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
