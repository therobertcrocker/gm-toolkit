package creature

import (
	"maps"
	"slices"
	"strings"
)

const StepCondition = "condition"

var (
	ConditionNames     = []string{"blinded", "clumsy", "concealed", "dazzled", "deafened", "doomed", "drained", "encumbered", "enfeebled", "frightened", "off-guard", "prone", "quickened", "sickened", "slowed", "stunned", "stupefied"}
	valuedConditions   = []string{"clumsy", "doomed", "drained", "enfeebled", "frightened", "sickened", "slowed", "stunned", "stupefied"}
	DurationBoundaries = []string{"source-turn-start", "source-turn-end", "target-turn-start", "target-turn-end"}
)

type Duration struct {
	Rounds int    `toml:"rounds,omitzero"`
	Until  string `toml:"until,omitempty"`
}

type ConditionAdjustment struct {
	Gain  int `toml:"gain,omitzero"`
	Decay int `toml:"decay,omitzero"`
}

func IsCondition(name string) bool {
	return slices.Contains(ConditionNames, name)
}

func IsValuedCondition(name string) bool {
	return slices.Contains(valuedConditions, name)
}

func checkCondition(path string, step *Step, report func(string, ...any)) {
	valued := IsValuedCondition(step.Condition)
	stunnedWithNoValue := step.Condition == "stunned" && step.Value == 0 && step.ValueDice == "" && step.Reduce == 0
	switch {
	case !IsCondition(step.Condition):
		report("%s: unknown condition %q", path, step.Condition)
	case valued && step.Value != 0 && step.ValueDice != "":
		report("%s: set value or value_dice, not both", path)
	case stunnedWithNoValue && step.Duration == nil:
		report("%s: stunned needs a value or a duration", path)
	case stunnedWithNoValue && step.Stack:
		report("%s: stack needs a value", path)
	case stunnedWithNoValue:
	case step.Reduce != 0 && (!valued || step.Value != 0 || step.ValueDice != "" || step.Stack || step.Duration != nil):
		report("%s: reduce needs a valued condition and no value, stack or duration", path)
	case valued && step.Reduce == 0 && step.Value < 1 && step.ValueDice == "":
		report("%s: %s needs a value of at least 1", path, step.Condition)
	case !valued && (step.Value != 0 || step.ValueDice != "" || step.Stack):
		report("%s: %s takes no value", path, step.Condition)
	}
	if step.ValueDice != "" {
		step.ValueRoll = parseAmount(path+".value_dice", step.ValueDice, report)
	}
}

func checkDuration(path string, duration *Duration, report func(string, ...any)) {
	if duration == nil {
		return
	}
	switch {
	case duration.Rounds < 0:
		report("%s.duration: rounds must not be negative", path)
	case duration.Rounds > 0 && duration.Until != "":
		report("%s.duration: set rounds or until, not both", path)
	case duration.Rounds == 0 && !slices.Contains(DurationBoundaries, duration.Until):
		report("%s.duration: until %q: want %s", path, duration.Until, strings.Join(DurationBoundaries, ", "))
	}
}

func (creature *Creature) checkConditionAdjustments(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.ConditionAdjustments)) {
		adjustment := creature.ConditionAdjustments[name]
		if !IsCondition(name) {
			report("condition_adjust.%s: not a condition", name)
		}
		if adjustment.Decay < 0 {
			report("condition_adjust.%s: decay must not be negative", name)
		}
		if adjustment.Gain == 0 && adjustment.Decay == 0 {
			report("condition_adjust.%s: sets neither gain nor decay", name)
		}
	}
}
