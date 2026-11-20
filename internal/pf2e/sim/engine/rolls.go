package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type rerollEffect struct {
	keep    string
	rolls   []string
	uses    int
	ability string
	clock   timing
}

type grantedBonus struct {
	modifier creature.CheckModifier
	uses     int
	clock    timing
}

func (fight *Fight) d20(roller *Combatant, kind string) int {
	die := fight.roller.Die(20)
	worse := roller.takeReroll(creature.RerollWorse, kind)
	better := roller.takeReroll(creature.RerollBetter, kind)
	switch {
	case worse == nil && better == nil:
		return die
	case worse != nil && better != nil:
		// deliberate: fortune and misfortune on one roll cancel, not stack, because the rules say the roll is then made once.
		fight.observer.Observe(RerollsCancelled{Combatant: roller.ID, Roll: kind})
		return die
	}
	second := fight.roller.Die(20)
	applied, kept := better, max(die, second)
	if worse != nil {
		applied, kept = worse, min(die, second)
	}
	fight.observer.Observe(Rerolled{Combatant: roller.ID, Roll: kind, Keep: applied.keep, First: die, Second: second, Kept: kept, Ability: applied.ability})
	return kept
}

func (combatant *Combatant) takeReroll(keep, kind string) *rerollEffect {
	for index, effect := range combatant.rerolls {
		if effect.keep != keep || !slices.Contains(effect.rolls, kind) {
			continue
		}
		if effect.uses > 0 {
			effect.uses--
			if effect.uses == 0 {
				combatant.rerolls = slices.Delete(combatant.rerolls, index, index+1)
			}
		}
		return effect
	}
	return nil
}

func (combatant *Combatant) spendBonuses(selector string) {
	combatant.bonuses = slices.DeleteFunc(combatant.bonuses, func(granted *grantedBonus) bool {
		if granted.uses == 0 || granted.modifier.Selector != selector {
			return false
		}
		granted.uses--
		return granted.uses == 0
	})
}

func (fight *Fight) addReroll(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.rerolls = append(target.rerolls, &rerollEffect{
		keep:    step.Reroll,
		rolls:   step.Rolls,
		uses:    step.Uses,
		ability: source.ability,
		clock:   fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(RerollGained{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability, Keep: step.Reroll})
}

func (fight *Fight) addBonus(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.bonuses = append(target.bonuses, &grantedBonus{
		modifier: *step.Bonus,
		uses:     step.Uses,
		clock:    fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(BonusGained{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability, Selector: step.Bonus.Selector, Value: step.Bonus.Value})
}

func (fight *Fight) expireEffects(bearer, actor *Combatant, boundary string) {
	bearer.rerolls = slices.DeleteFunc(bearer.rerolls, func(effect *rerollEffect) bool {
		return effect.clock.expires(actor, bearer, boundary)
	})
	bearer.bonuses = slices.DeleteFunc(bearer.bonuses, func(granted *grantedBonus) bool {
		return granted.clock.expires(actor, bearer, boundary)
	})
	bearer.recoveryCaps = slices.DeleteFunc(bearer.recoveryCaps, func(eased *recoveryCap) bool {
		return eased.clock.expires(actor, bearer, boundary)
	})
	bearer.grants = slices.DeleteFunc(bearer.grants, func(grant *grantedTrait) bool {
		return grant.clock.expires(actor, bearer, boundary)
	})
}
