package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type Fight struct {
	roller     dice.Roller
	observer   Observer
	decider    Decider
	maxRounds  int
	combatants []*Combatant
	order      []*Combatant
	round      int
	current    *Combatant
}

func (fight *Fight) StandingFoes(actor *Combatant) []*Combatant {
	var foes []*Combatant
	for _, combatant := range fight.combatants {
		if combatant.Side != actor.Side && combatant.Standing() {
			foes = append(foes, combatant)
		}
	}
	return foes
}

func (fight *Fight) moveBefore(moved, anchor *Combatant) {
	if moved == anchor {
		return
	}
	from := slices.Index(fight.order, moved)
	fight.order = slices.Delete(fight.order, from, from+1)
	to := slices.Index(fight.order, anchor)
	fight.order = slices.Insert(fight.order, to, moved)
	fight.observer.Observe(InitiativeMoved{Combatant: moved.ID, Before: anchor.ID})
}
