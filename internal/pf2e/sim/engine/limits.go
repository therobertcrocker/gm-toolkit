package engine

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func (combatant *Combatant) UsesLeft(ability creature.Ability) int {
	if ability.Limit == nil || ability.Limit.PerFight == 0 {
		return -1
	}
	return ability.Limit.PerFight - combatant.uses[ability.ID]
}

func (combatant *Combatant) RechargeLeft(abilityID string) int {
	return combatant.recharging[abilityID]
}

func (combatant *Combatant) Pool(name string) int {
	return combatant.pools[name]
}

func (combatant *Combatant) Blocked(ability creature.Ability) string {
	if limit := ability.Limit; limit != nil {
		switch {
		case limit.PerFight > 0 && combatant.uses[ability.ID] >= limit.PerFight:
			return "no uses left"
		case limit.PerRound > 0 && combatant.roundUses[ability.ID] >= limit.PerRound:
			return "used this round"
		case combatant.recharging[ability.ID] > 0:
			return "recharging"
		case limit.Pool != "" && combatant.pools[limit.Pool] < 1:
			return "pool " + limit.Pool + " is empty"
		}
	}
	switch {
	case ability.HasTrait("flourish") && combatant.flourished:
		return "flourish already used this round"
	case ability.HasTrait("press") && combatant.AttacksThisTurn == 0:
		return "press needs an earlier attack this turn"
	}
	return ""
}

func (fight *Fight) spend(actor *Combatant, ability creature.Ability) {
	if ability.HasTrait("flourish") {
		actor.flourished = true
	}
	limit := ability.Limit
	if limit == nil {
		return
	}
	if limit.PerFight > 0 {
		actor.uses[ability.ID]++
		fight.observer.Observe(UseSpent{Combatant: actor.ID, Ability: ability.ID, Left: limit.PerFight - actor.uses[ability.ID]})
	}
	if limit.PerRound > 0 {
		actor.roundUses[ability.ID]++
	}
	if limit.Pool != "" {
		actor.pools[limit.Pool]--
		fight.observer.Observe(PoolSpent{Combatant: actor.ID, Pool: limit.Pool, Ability: ability.ID, Left: actor.pools[limit.Pool]})
	}
	if limit.Recharge != "" {
		rounds := max(limit.RechargeRoll.Roll(fight.roller), 1)
		actor.recharging[ability.ID] = rounds
		fight.observer.Observe(RechargeRolled{Combatant: actor.ID, Ability: ability.ID, Rounds: rounds})
	}
}

// deliberate: a recharge counts down at the start of the user's turn, not at the end, because a roll of 1 must mean ready next turn (Decision 3).
func (fight *Fight) tickLimits(actor *Combatant) {
	// deliberate: flourish and per-round uses come back at the start of the user's turn, not at the top of the initiative order, because a round for one creature runs to the start of its next turn, as a 1-round duration does (Decision Record 10).
	clear(actor.roundUses)
	actor.flourished = false
	for _, id := range slices.Sorted(maps.Keys(actor.recharging)) {
		actor.recharging[id]--
		if actor.recharging[id] < 1 {
			delete(actor.recharging, id)
			fight.observer.Observe(Recharged{Combatant: actor.ID, Ability: id})
		}
	}
}

func (fight *Fight) recharge(actor *Combatant, abilityID, by string) {
	if actor.recharging[abilityID] < 1 {
		return
	}
	delete(actor.recharging, abilityID)
	fight.observer.Observe(Recharged{Combatant: actor.ID, Ability: abilityID, By: by})
}
