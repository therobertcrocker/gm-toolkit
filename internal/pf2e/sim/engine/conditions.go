package engine

import (
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var allChecksAndDCs = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "spell-attack", "spell-dc", "class-dc"}

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
	"clumsy":     {selectors: []string{"ac", "reflex", "ranged-attack"}, kind: "status"},
	"concealed":  {},
	"dazzled":    {},
	"doomed":     {},
	"drained":    {selectors: []string{"fortitude"}, kind: "status"},
	"encumbered": {implies: map[string]int{"clumsy": 1}},
	"enfeebled":  {selectors: []string{"melee-attack", "melee-damage"}, kind: "status"},
	"frightened": {selectors: allChecksAndDCs, kind: "status", decays: true},
	"off-guard":  {selectors: []string{"ac"}, kind: "circumstance", fixed: 2},
	"prone":      {selectors: []string{"attack", "spell-attack"}, kind: "circumstance", fixed: 2, implies: map[string]int{"off-guard": 0}},
	"sickened":   {selectors: allChecksAndDCs, kind: "status"},
	"stupefied":  {selectors: []string{"will", "perception", "spell-attack", "spell-dc"}, kind: "status"},
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
	// deliberate: an end-of-turn boundary applied during the owner's own turn skips that turn's end, because "until the end of your next turn" means the next one (Decision 9).
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
	if creature.IsValuedCondition(step.Condition) {
		value += target.Creature.ConditionAdjustments[step.Condition].Gain
		if value < 1 {
			return
		}
	}
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	drainedBefore := target.Condition("drained")
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
	if step.Condition == "drained" {
		fight.drain(source, target, drainedBefore)
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
	}
}

func (fight *Fight) endTurn(actor *Combatant) {
	fight.takePersistentDamage(actor)
	fight.decayConditions(actor)
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "end")
		fight.expireDefenseChanges(bearer, actor, "end")
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

func immuneToEffect(target *Combatant, traits []string) bool {
	return slices.ContainsFunc(target.Creature.Defense.Immune, func(key string) bool {
		return slices.Contains(traits, strings.TrimSuffix(key, "-effects"))
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
