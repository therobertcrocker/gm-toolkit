package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type effectSource struct {
	actor   *Combatant
	ability string
	strike  string
	traits  []string
	lethal  bool
}

func (fight *Fight) useAbility(actor *Combatant, ability creature.Ability, target *Combatant) {
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits}
	if ability.Area != nil {
		targets := fight.areaTargets(actor)
		for _, step := range ability.Steps {
			fight.runStep(source, step, targets)
		}
		return
	}
	if first := ability.Steps[0]; first.Kind() == creature.StepSave && first.Target != "self" && !fight.passesConcealment(actor, target) {
		return
	}
	for _, step := range ability.Steps {
		if step.Target != "self" && !target.Standing() {
			next, found := fight.decider.NextTarget(fight, actor)
			if !found {
				return
			}
			target = next
		}
		fight.runStep(source, step, []*Combatant{target})
	}
}

// deliberate: an area catches every standing foe, not a share of them, because every creature counts as engaged with every foe until Effort 7's position seam replaces this (Decision 1).
func (fight *Fight) areaTargets(actor *Combatant) []*Combatant {
	return fight.StandingFoes(actor)
}

func (fight *Fight) runSteps(source effectSource, steps []creature.Step, target *Combatant) {
	for _, step := range steps {
		fight.runStep(source, step, []*Combatant{target})
	}
}

func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	source.lethal = step.Lethal
	if step.Target == "self" {
		targets = []*Combatant{source.actor}
	}
	if step.Unless != "" {
		targets = slices.DeleteFunc(slices.Clone(targets), func(target *Combatant) bool { return target.Condition(step.Unless) > 0 })
	}
	if source.strike == "" && step.Kind() != creature.StepStrike {
		targets = fight.affected(source, targets)
	}
	switch step.Kind() {
	case creature.StepCondition:
		for _, target := range targets {
			fight.applyCondition(source.actor, target, step)
		}
	case creature.StepHeal:
		amount := step.HealRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.heal(source, target, amount)
		}
	case creature.StepTempHP:
		amount := step.TempHPRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.gainTempHP(target, amount)
		}
	case creature.StepWeaken:
		for _, target := range targets {
			fight.weaken(source.actor, target, step)
		}
	case creature.StepStrike:
		strike := source.actor.Creature.Strikes[step.Strike]
		for _, target := range targets {
			if target.Standing() {
				fight.strike(source.actor, target, strike, source.ability)
			}
		}
	case creature.StepSave:
		fight.saveStep(source, step, targets)
	case creature.StepSpellAttack:
		for _, target := range targets {
			if target.Standing() {
				fight.spellAttack(source, step, target)
			}
		}
	case creature.StepDamage:
		rolled := fight.rollDamage(step.Damage, false, 0)
		for _, target := range targets {
			if !target.Dead {
				fight.applyDamage(source, target, rolled, false)
			}
		}
	}
}

func (fight *Fight) saveStep(source effectSource, step creature.Step, targets []*Combatant) {
	dc := fight.stepDC(source.actor, step)
	var rolled damageRoll
	if step.Basic {
		rolled = fight.rollDamage(step.Damage, false, 0)
	}
	for _, target := range targets {
		if target.Dead {
			continue
		}
		degree := fight.savingThrow(target, step.Save, dc, step.Shift, source)
		if step.Basic && degree < CriticalSuccess {
			fight.applyDamage(source, target, rolled.basic(degree), degree == CriticalFailure)
		}
		fight.runSteps(source, step.For(int(degree), false), target)
	}
}

func (fight *Fight) stepDC(actor *Combatant, step creature.Step) int {
	switch step.DCFrom {
	case "class":
		return actor.Creature.ClassDC + fight.modifierTotal(actor, []string{"class-dc"}, nil)
	case "spell":
		return actor.Creature.Spellcasting.DC + fight.modifierTotal(actor, []string{"spell-dc"}, nil)
	default:
		return step.DC + fight.modifierTotal(actor, []string{"class-dc"}, nil)
	}
}

func (fight *Fight) spellAttack(source effectSource, step creature.Step, target *Combatant) {
	attacker := source.actor
	penalty := attackPenalty(attacker.AttacksThisTurn, false)
	attacker.AttacksThisTurn++
	if !fight.passesConcealment(attacker, target) {
		return
	}
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, source.traits)
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(SpellAttackRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Ability:  source.ability,
		Die:      die,
		Bonus:    bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	critical := degree == CriticalSuccess
	if degree >= Success && len(step.Damage) > 0 {
		fight.applyDamage(source, target, fight.rollDamage(step.Damage, critical, 0), critical)
	}
	if !target.Dead {
		fight.runSteps(source, step.For(int(degree), true), target)
	}
}
