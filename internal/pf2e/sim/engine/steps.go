package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type effectSource struct {
	actor       *Combatant
	ability     string
	strike      string
	traits      []string
	lethal      bool
	melee       bool
	actions     int
	triggeredBy string
	tags        []string
}

type abilityUse struct {
	actor       *Combatant
	ability     creature.Ability
	target      *Combatant
	actions     int
	fixed       bool
	triggeredBy string
}

func (fight *Fight) useAbility(use abilityUse) {
	actor, ability, target := use.actor, use.ability, use.target
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions, triggeredBy: use.triggeredBy, tags: ability.Tags}
	friendly := ability.Targets != ""
	if ability.Targets == creature.TargetsSelf {
		target = actor
	}
	if ability.Area != nil {
		targets := fight.areaTargets(actor, ability)
		for _, step := range ability.Steps {
			if !actor.Standing() {
				return
			}
			fight.runStep(source, step, targets)
		}
		return
	}
	if !friendly && len(ability.Steps) > 0 && ability.Steps[0].Kind() == creature.StepSave && ability.Steps[0].Target != creature.TargetSelf && !fight.passesConcealment(actor, target) {
		return
	}
	for _, step := range ability.Steps {
		// deliberate: the ability stops when its user falls, not at its last step, because a reaction can drop the user partway through (Decision 28).
		if !actor.Standing() {
			return
		}
		if step.Target != creature.TargetSelf && !target.Standing() && !(friendly && !target.Dead) {
			// deliberate: a triggered or friendly ability does not move to a new target, not even when its target fell, because it is for one creature; a dying ally still counts.
			if use.fixed || friendly {
				continue
			}
			next, found := fight.decider.NextTarget(fight, actor)
			if !found {
				return
			}
			target = next
		}
		stepTarget := target
		if step.Target == creature.TargetOther {
			stepTarget = fight.otherFoe(actor, target)
		}
		fight.runStep(source, step, []*Combatant{stepTarget})
	}
}

// deliberate: "other" is the first standing foe that is not the ability's target, not a decider pick, because Walk the Timelines is a stand-in (Decision 13); with one foe standing it is that foe.
func (fight *Fight) otherFoe(actor, target *Combatant) *Combatant {
	for _, foe := range fight.StandingFoes(actor) {
		if foe != target {
			return foe
		}
	}
	return target
}

func (fight *Fight) runSteps(source effectSource, steps []creature.Step, target *Combatant) {
	for _, step := range steps {
		fight.runStep(source, step, []*Combatant{target})
	}
}

func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	switch step.Kind() {
	case creature.StepGainActions:
		source.actor.ActionsLeft += step.GainActions
		fight.observer.Observe(ActionsGained{Combatant: source.actor.ID, Ability: source.ability, Amount: step.GainActions})
		return
	case creature.StepRecharge:
		by := source.ability
		if by == "" {
			by = source.strike
		}
		fight.recharge(source.actor, step.Recharge, by)
		return
	}
	source.lethal = step.Lethal
	if step.Target == creature.TargetSelf {
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
			if step.Reduce > 0 {
				fight.reduceCondition(target, step.Condition, step.Reduce)
				continue
			}
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
				fight.strike(source.actor, target, strike, strikeOptions{
					ability:     source.ability,
					traits:      source.traits,
					extraDice:   step.ExtraDice,
					extraDamage: step.ExtraDamage,
					counts:      step.Counts,
					penalty:     step.Penalty,
					disrupts:    step.Disrupts,
					lists:       step.Outcomes,
					triggeredBy: source.triggeredBy,
				})
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
		rolled := fight.rollDamage(scaleDamage(step.Damage, source.actions), false, 0)
		for _, target := range targets {
			if !target.Dead {
				fight.applyDamage(source, target, rolled, false)
			}
		}
	case creature.StepReroll:
		for _, target := range targets {
			fight.addReroll(source, target, step)
		}
	case creature.StepBonus:
		for _, target := range targets {
			fight.addBonus(source, target, step)
		}
	case creature.StepRecovery:
		for _, target := range targets {
			fight.easeRecovery(source, target, step)
		}
	case creature.StepGrant:
		for _, target := range targets {
			fight.grantTrait(source, target, step)
		}
	case creature.StepCheck:
		for _, target := range targets {
			if target.Standing() {
				fight.runSteps(source, step.For(int(fight.skillCheck(source, step, target)), false), target)
			}
		}
	}
}

func (fight *Fight) saveStep(source effectSource, step creature.Step, targets []*Combatant) {
	dc := fight.stepDC(source.actor, step)
	var rolled damageRoll
	if step.Basic {
		rolled = fight.rollDamage(scaleDamage(step.Damage, source.actions), false, 0)
	}
	for _, target := range targets {
		if target.Dead {
			continue
		}
		degree := fight.savingThrow(target, step.Save, dc, step.Shift, source)
		if step.Basic && degree < CriticalSuccess {
			taken := rolled.basic(degree)
			if degree == Success && step.SuccessDamage == creature.SuccessFull {
				taken = rolled
			}
			fight.applyDamage(source, target, taken, degree == CriticalFailure)
		}
		fight.runSteps(source, step.For(int(degree), false), target)
		if degree >= Success && source.actor.Side != target.Side && slices.Contains(source.tags, creature.TagTestsLoyalty) {
			fight.loyalty(target, source.ability)
		}
		if source.actor.Side != target.Side {
			fight.fire(target, TriggerEvent{
				Kind:     creature.TriggerSaved,
				By:       source.actor,
				Ability:  source.ability,
				Traits:   source.traits,
				Degree:   degree,
				Damaging: dealsDamage(step),
			})
		}
	}
}

func dealsDamage(step creature.Step) bool {
	if len(step.Damage) > 0 {
		return true
	}
	for _, list := range [][]creature.Step{step.CriticalSuccess, step.Success, step.Failure, step.CriticalFailure} {
		if slices.ContainsFunc(list, dealsDamage) {
			return true
		}
	}
	return false
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
	penalty := 0
	if attacker == fight.current {
		penalty = attackPenalty(attacker.AttacksThisTurn, false)
		attacker.AttacksThisTurn++
		attacker.lastPenalty = penalty
	}
	if !fight.passesConcealment(attacker, target) {
		return
	}
	attack := TriggerEvent{Kind: creature.TriggerTargeted, By: attacker, Ability: source.ability, Traits: source.traits}
	fight.fire(target, attack)
	if !attacker.Standing() {
		return
	}
	die := fight.d20(attacker, creature.RollAttack)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, source.traits)
	target.spendBonuses("ac")
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
		fight.applyDamage(source, target, fight.rollDamage(scaleDamage(step.Damage, source.actions), critical, 0), critical)
	}
	if !target.Dead {
		fight.runSteps(source, step.For(int(degree), true), target)
	}
	attack.Kind, attack.Degree = creature.TriggerAttacked, degree
	fight.fire(target, attack)
}

func scaleDamage(parts []creature.DamagePart, actions int) []creature.DamagePart {
	if actions < 2 || !slices.ContainsFunc(parts, func(part creature.DamagePart) bool { return part.PerAction }) {
		return parts
	}
	scaled := slices.Clone(parts)
	for index := range scaled {
		if scaled[index].PerAction {
			scaled[index].Roll.Count *= actions
			scaled[index].Roll.Modifier *= actions
		}
	}
	return scaled
}
