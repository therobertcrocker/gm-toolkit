package creature

import (
	"maps"
	"slices"
	"strings"
)

var (
	DegreeWords      = []string{"critical-failure", "failure", "success", "critical-success"}
	CheckSelectors   = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "melee-attack", "ranged-attack", "damage", "melee-damage", "spell-attack", "spell-dc", "class-dc"}
	saveSelectors    = []string{"fortitude", "reflex", "will", "saving-throw"}
	againstSelectors = []string{"ac", "fortitude", "reflex", "will", "saving-throw"}
)

type CheckModifier struct {
	Selector     string   `toml:"selector"`
	Type         string   `toml:"type,omitempty"`
	Value        int      `toml:"value"`
	Against      []string `toml:"against,omitempty"`
	PerCondition string   `toml:"per_condition,omitempty"`
}

var offenseSelectors = []string{"attack", "melee-attack", "ranged-attack", "damage", "melee-damage"}

type DegreeShift struct {
	Selector string   `toml:"selector"`
	Against  []string `toml:"against,omitempty"`
	From     string   `toml:"from"`
	To       string   `toml:"to"`
}

func (modifier CheckModifier) Applies(selectors, traits []string) bool {
	return slices.Contains(selectors, modifier.Selector) && matchesAgainst(modifier.Against, traits)
}

func (shift DegreeShift) Applies(selectors, traits []string) bool {
	return slices.Contains(selectors, shift.Selector) && matchesAgainst(shift.Against, traits)
}

func matchesAgainst(against, traits []string) bool {
	return len(against) == 0 || slices.ContainsFunc(against, func(trait string) bool { return slices.Contains(traits, trait) })
}

func (creature *Creature) DegreeShiftIDs() []string {
	return slices.Sorted(maps.Keys(creature.DegreeShifts))
}

func (creature *Creature) checkFightRules(report func(string, ...any)) {
	for _, id := range slices.Sorted(maps.Keys(creature.Modifiers)) {
		modifier := creature.Modifiers[id]
		if !slices.Contains(CheckSelectors, modifier.Selector) {
			report("modifier.%s: selector %q: want %s", id, modifier.Selector, strings.Join(CheckSelectors, ", "))
		}
		if !slices.Contains(modifierTypes, modifier.Type) {
			report("modifier.%s: type %q: want item, status, circumstance or none", id, modifier.Type)
		}
		if modifier.Value == 0 {
			report("modifier.%s: value must not be 0", id)
		}
		if len(modifier.Against) > 0 && !slices.Contains(againstSelectors, modifier.Selector) {
			report("modifier.%s: against needs selector ac or a save", id)
		}
		if modifier.PerCondition != "" && (!IsCondition(modifier.PerCondition) || !slices.Contains(offenseSelectors, modifier.Selector)) {
			report("modifier.%s: per_condition needs a condition and an attack or damage selector", id)
		}
	}
	for _, id := range creature.DegreeShiftIDs() {
		shift := creature.DegreeShifts[id]
		if !slices.Contains(saveSelectors, shift.Selector) {
			report("degree_shift.%s: selector %q: want %s", id, shift.Selector, strings.Join(saveSelectors, ", "))
		}
		if !slices.Contains(DegreeWords, shift.From) || !slices.Contains(DegreeWords, shift.To) {
			report("degree_shift.%s: from %q and to %q must each be one of %s", id, shift.From, shift.To, strings.Join(DegreeWords, ", "))
		}
	}
	creature.checkConditionAdjustments(report)
}
