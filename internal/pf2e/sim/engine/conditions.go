package engine

import (
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var effectImmunityTraits = []string{"auditory", "curse", "disease", "emotion", "fear", "mental", "olfactory", "poison", "polymorph", "possession", "sleep", "visual"}

var allChecksAndDCs = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "spell-attack", "spell-dc", "class-dc", "skill-check"}

type conditionRule struct {
	selectors []string
	kind      string
	fixed     int
	decays    bool
	implies   map[string]int
}

// deliberate: each condition is a row of shared parts, not a handler, because the ruleset as data (D-2) will lift this table into files (Decision 2).
var conditionRules = map[string]conditionRule{
	"blinded":    {implies: map[string]int{"off-guard": 0}},
	"clumsy":     {selectors: []string{"ac", "reflex", "ranged-attack", "dex-skill"}, kind: "status"},
	"concealed":  {},
	"dazzled":    {},
	"deafened":   {},
	"doomed":     {},
	"drained":    {selectors: []string{"fortitude"}, kind: "status"},
	"encumbered": {implies: map[string]int{"clumsy": 1}},
	"enfeebled":  {selectors: []string{"melee-attack", "melee-damage", "str-skill"}, kind: "status"},
	"frightened": {selectors: allChecksAndDCs, kind: "status", decays: true},
	"off-guard":  {selectors: []string{"ac"}, kind: "circumstance", fixed: 2},
	"prone":      {selectors: []string{"attack", "spell-attack"}, kind: "circumstance", fixed: 2, implies: map[string]int{"off-guard": 0}},
	"quickened":  {},
	"sickened":   {selectors: allChecksAndDCs, kind: "status"},
	"slowed":     {},
	"stunned":    {},
	"stupefied":  {selectors: []string{"will", "perception", "spell-attack", "spell-dc", "int-skill", "wis-skill", "cha-skill"}, kind: "status"},
}

type timing struct {
	source *Combatant
	rounds int
	until  string
	armed  bool
}

type conditionInstance struct {
	name  string
	value int
	clock timing
}

func (fight *Fight) newTiming(source, bearer *Combatant, duration *creature.Duration) timing {
	if duration == nil {
		return timing{source: source}
	}
	clock := timing{source: source, rounds: duration.Rounds, until: duration.Until}
	// deliberate: an end-of-turn boundary applied during the owner's own turn ends at the next turn's end, not this one's, because "until the end of your next turn" means the next one (Decision 9).
	clock.armed = fight.current != clock.owner(bearer)
	return clock
}

func (clock timing) owner(bearer *Combatant) *Combatant {
	if strings.HasPrefix(clock.until, "target-") {
		return bearer
	}
	return clock.source
}

func (clock *timing) expires(actor, bearer *Combatant, boundary string) bool {
	if clock.rounds > 0 {
		if boundary != "start" || clock.source != actor {
			return false
		}
		clock.rounds--
		return clock.rounds == 0
	}
	if clock.until == "" || clock.owner(bearer) != actor || !strings.HasSuffix(clock.until, "-"+boundary) {
		return false
	}
	if boundary == "end" && !clock.armed {
		clock.armed = true
		return false
	}
	return true
}

func (combatant *Combatant) Condition(name string) int {
	value := 0
	for _, instance := range combatant.conditions {
		if instance.name == name {
			value = max(value, instance.value, 1)
		}
		if implied, found := conditionRules[instance.name].implies[name]; found {
			value = max(value, implied, 1)
		}
	}
	return value
}

func (fight *Fight) conditionPenalties(combatant *Combatant, selectors []string, pool *creature.BonusPool) {
	for _, name := range creature.ConditionNames {
		rule := conditionRules[name]
		value := combatant.Condition(name)
		if value == 0 || !overlaps(rule.selectors, selectors) {
			continue
		}
		if rule.fixed > 0 {
			value = rule.fixed
		}
		pool.Add(rule.kind, -value)
	}
}

func overlaps(first, second []string) bool {
	return slices.ContainsFunc(first, func(entry string) bool { return slices.Contains(second, entry) })
}

func (fight *Fight) applyCondition(source, target *Combatant, step creature.Step) {
	if target.Dead || slices.Contains(target.Creature.Defense.Immune, step.Condition) {
		return
	}
	value := step.Value
	if step.ValueDice != "" {
		value = step.ValueRoll.Roll(fight.roller)
	}
	// deliberate: a stunned step with no value lands as a value-0 instance, not dropped as a value under 1, because stunned can have a duration instead of a value (Player Core pg. 446).
	durationOnly := step.Condition == "stunned" && step.Value == 0 && step.ValueDice == ""
	if creature.IsValuedCondition(step.Condition) && !durationOnly {
		value += target.Creature.ConditionAdjustments[step.Condition].Gain
		if value < 1 {
			return
		}
		if step.Stack {
			value += target.Condition(step.Condition)
		}
	}
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	drainedBefore, before := target.Condition("drained"), target.Condition(step.Condition)
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
	if target.Condition(step.Condition) > before {
		fight.gained(target, step.Condition)
	}
	if step.Condition == "drained" {
		fight.drain(source, target, drainedBefore)
	}
	if step.Condition == "doomed" && target.Dying > 0 && target.Dying >= target.deathValue() {
		fight.setDying(target, target.Dying)
	}
	if step.Condition == "stunned" && target == fight.current && target.Standing() {
		fight.loseActionsToStunned(target)
	}
}

func (fight *Fight) changeConditions(bearer *Combatant, change func()) {
	before := make(map[string]int, len(creature.ConditionNames))
	for _, name := range creature.ConditionNames {
		before[name] = bearer.Condition(name)
	}
	change()
	for _, name := range creature.ConditionNames {
		if after := bearer.Condition(name); after != before[name] {
			fight.observer.Observe(ConditionChanged{Combatant: bearer.ID, Condition: name, From: before[name], To: after})
		}
	}
}

func (fight *Fight) startTurn(actor *Combatant) {
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "start")
		fight.expireDefenseChanges(bearer, actor, "start")
		fight.expireEffects(bearer, actor, "start")
	}
}

func (fight *Fight) endTurn(actor *Combatant) {
	fight.takePersistentDamage(actor)
	fight.decayConditions(actor)
	fight.endSustains(actor)
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "end")
		fight.expireDefenseChanges(bearer, actor, "end")
		fight.expireEffects(bearer, actor, "end")
	}
}

type defenseChange struct {
	change creature.DefenseChange
	clock  timing
}

func (fight *Fight) weaken(source, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	change := *step.Weaken
	target.defenseChanges = append(target.defenseChanges, &defenseChange{change: change, clock: fight.newTiming(source, target, step.Duration)})
	fight.observer.Observe(DefenseChanged{Combatant: target.ID, Source: source.ID, Weakness: change.Weakness, Resistance: change.Resistance, Types: change.Types})
}

func (fight *Fight) expireDefenseChanges(bearer, actor *Combatant, boundary string) {
	var kept []*defenseChange
	for _, applied := range bearer.defenseChanges {
		if applied.clock.expires(actor, bearer, boundary) {
			fight.observer.Observe(DefenseRestored{Combatant: bearer.ID, Types: applied.change.Types})
			continue
		}
		kept = append(kept, applied)
	}
	bearer.defenseChanges = kept
}

func (fight *Fight) expire(bearer, actor *Combatant, boundary string) {
	if len(bearer.conditions) == 0 {
		return
	}
	fight.changeConditions(bearer, func() {
		var kept []*conditionInstance
		for _, instance := range bearer.conditions {
			if !instance.clock.expires(actor, bearer, boundary) {
				kept = append(kept, instance)
			}
		}
		bearer.conditions = kept
	})
}

func (fight *Fight) decayConditions(actor *Combatant) {
	if len(actor.conditions) == 0 {
		return
	}
	fight.changeConditions(actor, func() {
		var kept []*conditionInstance
		for _, instance := range actor.conditions {
			if conditionRules[instance.name].decays {
				decay := actor.Creature.ConditionAdjustments[instance.name].Decay
				if decay == 0 {
					decay = 1
				}
				instance.value -= decay
				if instance.value < 1 {
					continue
				}
			}
			kept = append(kept, instance)
		}
		actor.conditions = kept
	})
}

func (fight *Fight) reduceCondition(bearer *Combatant, name string, amount int) {
	fight.changeConditions(bearer, func() {
		var kept []*conditionInstance
		for _, instance := range bearer.conditions {
			// deliberate: a value-0 instance is left alone, not dropped, because a stunned with a duration ends only by its clock.
			if instance.name == name && instance.value > 0 {
				instance.value -= amount
				if instance.value < 1 {
					continue
				}
			}
			kept = append(kept, instance)
		}
		bearer.conditions = kept
	})
}

func immuneToEffect(target *Combatant, traits []string) bool {
	return slices.ContainsFunc(target.Creature.Defense.Immune, func(key string) bool {
		trait, isEffectKey := strings.CutSuffix(key, "-effects")
		if !isEffectKey && !slices.Contains(effectImmunityTraits, key) {
			return false
		}
		return slices.Contains(traits, trait)
	})
}

func (fight *Fight) affected(source effectSource, targets []*Combatant) []*Combatant {
	var affected []*Combatant
	for _, target := range targets {
		if immuneToEffect(target, source.traits) {
			fight.observer.Observe(ImmuneToEffect{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability})
			continue
		}
		affected = append(affected, target)
	}
	return affected
}

func (fight *Fight) passesConcealment(attacker, target *Combatant) bool {
	dc := 0
	switch {
	case attacker.Condition("blinded") > 0:
		dc = 11
	case target.Condition("concealed") > 0 || attacker.Condition("dazzled") > 0:
		dc = 5
	default:
		return true
	}
	die := fight.roller.Die(20)
	fight.observer.Observe(FlatChecked{Combatant: attacker.ID, Reason: "concealment", Die: die, DC: dc})
	return die >= dc
}
