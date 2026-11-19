package creature

import (
	"fmt"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const (
	StepStrike      = "strike"
	StepSave        = "save"
	StepSpellAttack = "spell_attack"
	StepDamage      = "damage"
	StepHeal        = "heal"
	StepTempHP      = "temp_hp"
	StepWeaken      = "weaken"
)

type DefenseChange struct {
	Weakness   int      `toml:"weakness,omitzero"`
	Resistance int      `toml:"resistance,omitzero"`
	Types      []string `toml:"types"`
}

var (
	SaveNames   = []string{"fortitude", "reflex", "will"}
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken}
	rolledSteps = []string{StepStrike, StepSave, StepSpellAttack}
	areaShapes  = []string{"burst", "cone", "emanation", "line"}
	dcSources   = []string{"class", "spell"}
)

type Area struct {
	Shape string `toml:"shape"`
	Size  int    `toml:"size"`
}

type Outcomes struct {
	CriticalSuccess []Step `toml:"critical_success,omitempty"`
	Success         []Step `toml:"success,omitempty"`
	Failure         []Step `toml:"failure,omitempty"`
	CriticalFailure []Step `toml:"critical_failure,omitempty"`
}

type Step struct {
	Strike      string          `toml:"strike,omitempty"`
	Save        string          `toml:"save,omitempty"`
	SpellAttack bool            `toml:"spell_attack,omitempty"`
	Damage      []DamagePart    `toml:"damage,omitempty"`
	Target      string          `toml:"target,omitempty"`
	DC          int             `toml:"dc,omitzero"`
	DCFrom      string          `toml:"dc_from,omitempty"`
	Basic       bool            `toml:"basic,omitempty"`
	Shift       int             `toml:"shift,omitzero"`
	Condition   string          `toml:"condition,omitempty"`
	Value       int             `toml:"value,omitzero"`
	Duration    *Duration       `toml:"duration,omitempty"`
	Unless      string          `toml:"unless,omitempty"`
	Lethal      bool            `toml:"lethal,omitempty"`
	Heal        string          `toml:"heal,omitempty"`
	TempHP      string          `toml:"temp_hp,omitempty"`
	HealRoll    dice.Expression `toml:"-"`
	TempHPRoll  dice.Expression `toml:"-"`
	Weaken      *DefenseChange  `toml:"weaken,omitempty"`
	Outcomes
}

func (outcomes Outcomes) For(degree int, attack bool) []Step {
	if attack && degree == 3 && len(outcomes.CriticalSuccess) == 0 {
		return outcomes.Success
	}
	return outcomes.lists()[degree]
}

func (outcomes Outcomes) lists() [][]Step {
	return [][]Step{outcomes.CriticalFailure, outcomes.Failure, outcomes.Success, outcomes.CriticalSuccess}
}

func (step Step) Kind() string {
	if heads := step.heads(); len(heads) == 1 {
		return heads[0]
	}
	return ""
}

func (step Step) heads() []string {
	var heads []string
	if step.Strike != "" {
		heads = append(heads, StepStrike)
	}
	if step.Save != "" {
		heads = append(heads, StepSave)
	}
	if step.SpellAttack {
		heads = append(heads, StepSpellAttack)
	}
	if len(step.Damage) > 0 && step.Save == "" && !step.SpellAttack {
		heads = append(heads, StepDamage)
	}
	if step.Heal != "" {
		heads = append(heads, StepHeal)
	}
	if step.TempHP != "" {
		heads = append(heads, StepTempHP)
	}
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	if step.Weaken != nil {
		heads = append(heads, StepWeaken)
	}
	return heads
}

func (creature *Creature) checkStep(path string, step *Step, report func(string, ...any)) {
	heads := step.heads()
	if len(heads) != 1 {
		report("%s: want exactly one of %s; got %d", path, strings.Join(stepKinds, ", "), len(heads))
		return
	}
	kind := heads[0]
	switch kind {
	case StepStrike:
		if _, found := creature.Strikes[step.Strike]; !found {
			report("%s: unknown strike %q", path, step.Strike)
		}
	case StepSave:
		if !slices.Contains(SaveNames, step.Save) {
			report("%s: save %q: want %s", path, step.Save, strings.Join(SaveNames, ", "))
		}
		creature.checkDC(path, step, report)
		if len(step.Damage) > 0 && !step.Basic {
			report("%s: damage on a save needs basic = true; put other damage in the degree lists", path)
		}
		if step.Basic && len(step.Damage) == 0 {
			report("%s: basic needs damage", path)
		}
	case StepSpellAttack:
		if creature.Spellcasting == nil {
			report("%s: spell_attack needs [spellcasting]", path)
		}
	case StepCondition:
		checkCondition(path, step, report)
	case StepHeal:
		step.HealRoll = parseAmount(path+".heal", step.Heal, report)
	case StepTempHP:
		step.TempHPRoll = parseAmount(path+".temp_hp", step.TempHP, report)
	}
	if kind != StepCondition && step.Value != 0 {
		report("%s: value belongs on a condition", path)
	}
	if kind != StepCondition && kind != StepWeaken && step.Duration != nil {
		report("%s: duration belongs on a condition or weaken", path)
	}
	if kind == StepWeaken {
		change := step.Weaken
		if change.Weakness < 0 || change.Resistance < 0 || change.Weakness+change.Resistance == 0 {
			report("%s.weaken: want a weakness or a resistance above 0, and neither below", path)
		}
		if len(change.Types) == 0 || !slices.ContainsFunc(change.Types, IsDamageType) {
			report("%s.weaken: want damage types", path)
		}
	}
	checkDuration(path, step.Duration, report)
	if step.Unless != "" && !IsCondition(step.Unless) {
		report("%s: unless %q: not a condition", path, step.Unless)
	}
	if step.Lethal && len(step.Damage) == 0 {
		report("%s: lethal needs damage", path)
	}
	if kind != StepSave && (step.DC != 0 || step.DCFrom != "" || step.Basic || step.Shift != 0) {
		report("%s: dc, dc_from, basic and shift belong on a save", path)
	}
	if step.Shift < -3 || step.Shift > 3 {
		report("%s: shift %d: want -3 to 3", path, step.Shift)
	}
	if step.Target != "" && step.Target != "self" {
		report("%s: target %q: want self, or leave it unset", path, step.Target)
	}
	prepareDamage(path, step.Damage, report)
	creature.checkOutcomes(path, step.Outcomes, slices.Contains(rolledSteps, kind), report)
}

func (creature *Creature) checkOutcomes(path string, outcomes Outcomes, rolled bool, report func(string, ...any)) {
	for degree, list := range outcomes.lists() {
		word := strings.ReplaceAll(DegreeWords[degree], "-", "_")
		if len(list) > 0 && !rolled {
			report("%s: %s needs a strike, save or spell_attack", path, word)
		}
		for index := range list {
			creature.checkStep(fmt.Sprintf("%s.%s[%d]", path, word, index), &list[index], report)
		}
	}
}

func (creature *Creature) checkDC(path string, step *Step, report func(string, ...any)) {
	switch {
	case step.DC > 0 && step.DCFrom != "":
		report("%s: set dc or dc_from, not both", path)
	case step.DC > 0:
	case !slices.Contains(dcSources, step.DCFrom):
		report("%s: a save needs dc, or dc_from class or spell", path)
	case step.DCFrom == "class" && creature.ClassDC < 1:
		report("%s: dc_from class needs a class DC", path)
	case step.DCFrom == "spell" && creature.Spellcasting == nil:
		report("%s: dc_from spell needs [spellcasting]", path)
	}
}

func checkArea(path string, area *Area, report func(string, ...any)) {
	if area == nil {
		return
	}
	if !slices.Contains(areaShapes, area.Shape) {
		report("%s.area: shape %q: want %s", path, area.Shape, strings.Join(areaShapes, ", "))
	}
	if area.Size < 5 {
		report("%s.area: size must be at least 5", path)
	}
}

func prepareDamage(path string, parts []DamagePart, report func(string, ...any)) {
	for index := range parts {
		part := &parts[index]
		expression, err := dice.Parse(part.Dice)
		if err != nil {
			report("%s.damage[%d]: %v", path, index, err)
		}
		part.Roll = expression
		if !IsDamageType(part.Type) {
			report("%s.damage[%d]: unknown damage type %q", path, index, part.Type)
		}
		if !slices.Contains(damageCategories, part.Category) {
			report("%s.damage[%d]: category %q: want persistent, splash or none", path, index, part.Category)
		}
	}
}

func parseAmount(path, text string, report func(string, ...any)) dice.Expression {
	expression, err := dice.Parse(text)
	if err != nil {
		report("%s: %v", path, err)
	}
	return expression
}
