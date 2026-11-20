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
	StepGainActions = "gain_actions"
	StepRecharge    = "recharge"
	StepReroll      = "reroll"
	StepBonus       = "bonus"
	StepRecovery    = "recovery_dc"
	StepGrant       = "grant"
	StepCheck       = "check"

	TagLoyalty      = "loyalty"
	TagTestsLoyalty = "tests-loyalty"

	TargetsSelf   = "self"
	TargetsAlly   = "ally"
	TargetsOthers = "others"

	RerollWorse  = "worse"
	RerollBetter = "better"
	RollAttack   = "attack"
	RollSave     = "save"
	RollSkill    = "skill"

	PenaltyShared = "shared"
	TargetSelf    = "self"
	TargetOther   = "other"
	SuccessFull   = "full"
)

type DefenseChange struct {
	Weakness   int      `toml:"weakness,omitzero"`
	Resistance int      `toml:"resistance,omitzero"`
	Types      []string `toml:"types"`
}

var (
	SaveNames      = []string{"fortitude", "reflex", "will"}
	stepKinds      = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken, StepGainActions, StepRecharge, StepReroll, StepBonus, StepRecovery, StepGrant, StepCheck}
	timedSteps     = []string{StepCondition, StepWeaken, StepReroll, StepBonus, StepRecovery, StepGrant}
	usedSteps      = []string{StepReroll, StepBonus, StepGrant}
	rollKinds      = []string{RollAttack, RollSave, RollSkill}
	abilityTargets = []string{"", TargetsSelf, TargetsAlly, TargetsOthers}
	abilityTags    = []string{TagLoyalty, TagTestsLoyalty}
	stepTargets    = []string{"", TargetSelf, TargetOther}
	rolledSteps    = []string{StepStrike, StepSave, StepSpellAttack, StepCheck}
	checkDefenses  = []string{"fortitude", "reflex", "will", "perception"}
	areaShapes     = []string{"burst", "cone", "emanation", "line"}
	dcSources      = []string{"class", "spell"}
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
	Strike        string          `toml:"strike,omitempty"`
	Save          string          `toml:"save,omitempty"`
	SpellAttack   bool            `toml:"spell_attack,omitempty"`
	Damage        []DamagePart    `toml:"damage,omitempty"`
	Target        string          `toml:"target,omitempty"`
	DC            int             `toml:"dc,omitzero"`
	DCFrom        string          `toml:"dc_from,omitempty"`
	Basic         bool            `toml:"basic,omitempty"`
	Shift         int             `toml:"shift,omitzero"`
	Condition     string          `toml:"condition,omitempty"`
	Value         int             `toml:"value,omitzero"`
	ValueDice     string          `toml:"value_dice,omitempty"`
	ValueRoll     dice.Expression `toml:"-"`
	Stack         bool            `toml:"stack,omitempty"`
	Duration      *Duration       `toml:"duration,omitempty"`
	Unless        string          `toml:"unless,omitempty"`
	Lethal        bool            `toml:"lethal,omitempty"`
	Heal          string          `toml:"heal,omitempty"`
	TempHP        string          `toml:"temp_hp,omitempty"`
	HealRoll      dice.Expression `toml:"-"`
	TempHPRoll    dice.Expression `toml:"-"`
	Weaken        *DefenseChange  `toml:"weaken,omitempty"`
	GainActions   int             `toml:"gain_actions,omitzero"`
	Recharge      string          `toml:"recharge,omitempty"`
	ExtraDice     int             `toml:"extra_dice,omitzero"`
	ExtraDamage   []DamagePart    `toml:"extra_damage,omitempty"`
	Counts        int             `toml:"counts,omitzero"`
	Penalty       string          `toml:"penalty,omitempty"`
	Disrupts      []string        `toml:"disrupts,omitempty"`
	Reduce        int             `toml:"reduce,omitzero"`
	Reroll        string          `toml:"reroll,omitempty"`
	Rolls         []string        `toml:"rolls,omitempty"`
	Bonus         *CheckModifier  `toml:"bonus,omitempty"`
	Uses          int             `toml:"uses,omitzero"`
	RecoveryDC    int             `toml:"recovery_dc,omitzero"`
	Grant         string          `toml:"grant,omitempty"`
	Check         string          `toml:"check,omitempty"`
	Against       string          `toml:"against,omitempty"`
	SuccessDamage string          `toml:"success_damage,omitempty"`
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
	if step.GainActions != 0 {
		heads = append(heads, StepGainActions)
	}
	if step.Recharge != "" {
		heads = append(heads, StepRecharge)
	}
	if step.Reroll != "" {
		heads = append(heads, StepReroll)
	}
	if step.Bonus != nil {
		heads = append(heads, StepBonus)
	}
	if step.RecoveryDC != 0 {
		heads = append(heads, StepRecovery)
	}
	if step.Grant != "" {
		heads = append(heads, StepGrant)
	}
	if step.Check != "" {
		heads = append(heads, StepCheck)
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
	case StepGainActions:
		if step.GainActions < 1 || step.GainActions > 3 {
			report("%s: gain_actions %d: want 1 to 3", path, step.GainActions)
		}
	case StepRecharge:
		if target, found := creature.Abilities[step.Recharge]; !found || target.Limit == nil || target.Limit.Recharge == "" {
			report("%s: recharge %q: want an ability with limit.recharge", path, step.Recharge)
		}
	case StepReroll:
		if step.Reroll != RerollWorse && step.Reroll != RerollBetter {
			report("%s: reroll %q: want worse or better", path, step.Reroll)
		}
		if len(step.Rolls) == 0 || slices.ContainsFunc(step.Rolls, func(kind string) bool { return !slices.Contains(rollKinds, kind) }) {
			report("%s: rolls: want one or more of %s", path, strings.Join(rollKinds, ", "))
		}
	case StepBonus:
		bonus := step.Bonus
		if !slices.Contains(CheckSelectors, bonus.Selector) || !slices.Contains(modifierTypes, bonus.Type) || bonus.Value == 0 || bonus.PerCondition != "" {
			report("%s.bonus: want a known selector and type, a value other than 0, and no per_condition", path)
		}
		if step.Uses > 0 && bonus.Selector != "ac" {
			report("%s: uses on a bonus needs selector ac", path)
		}
	case StepRecovery:
		if step.RecoveryDC < 1 || step.Duration == nil {
			report("%s: recovery_dc must be at least 1 and needs a duration", path)
		}
	case StepGrant:
		if !IsDefenseTrait(step.Grant) || step.Uses < 1 {
			report("%s: grant %q: want one of %s, with uses of at least 1", path, step.Grant, strings.Join(DefenseTraits, ", "))
		}
	case StepCheck:
		if _, found := creature.Skills[step.Check]; !IsSkill(step.Check) || !found {
			report("%s: check %q: want a skill the creature has a number for", path, step.Check)
		}
		if (step.Against == "") == (step.DC == 0) || (step.Against != "" && !slices.Contains(checkDefenses, step.Against)) {
			report("%s: a check needs one of dc and against (%s)", path, strings.Join(checkDefenses, ", "))
		}
	}
	if kind != StepCondition && (step.Value != 0 || step.ValueDice != "" || step.Stack) {
		report("%s: value belongs on a condition", path)
	}
	if !slices.Contains(timedSteps, kind) && step.Duration != nil {
		report("%s: duration belongs on one of %s", path, strings.Join(timedSteps, ", "))
	}
	if len(step.Rolls) > 0 && kind != StepReroll {
		report("%s: rolls belongs on a reroll", path)
	}
	switch used := slices.Contains(usedSteps, kind); {
	case step.Uses < 0 || (step.Uses != 0 && !used):
		report("%s: uses belongs on one of %s and must not be negative", path, strings.Join(usedSteps, ", "))
	case used && step.Uses == 0 && step.Duration == nil:
		report("%s: %s needs uses or a duration", path, kind)
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
	if kind != StepSave && (step.DCFrom != "" || step.Basic || step.Shift != 0) {
		report("%s: dc_from, basic and shift belong on a save", path)
	}
	if kind != StepSave && kind != StepCheck && step.DC != 0 {
		report("%s: dc belongs on a save or a check", path)
	}
	if kind != StepCheck && step.Against != "" {
		report("%s: against belongs on a check", path)
	}
	if step.Shift < -3 || step.Shift > 3 {
		report("%s: shift %d: want -3 to 3", path, step.Shift)
	}
	if !slices.Contains(stepTargets, step.Target) {
		report("%s: target %q: want self, other, or leave it unset", path, step.Target)
	}
	if kind != StepStrike && (step.ExtraDice != 0 || len(step.ExtraDamage) > 0 || step.Counts != 0 || step.Penalty != "" || len(step.Disrupts) > 0) {
		report("%s: extra_dice, extra_damage, counts, penalty and disrupts belong on a strike", path)
	}
	if step.Reduce != 0 && (kind != StepCondition || step.Reduce < 0) {
		report("%s: reduce belongs on a condition and must be positive", path)
	}
	if step.ExtraDice < 0 || step.Counts < 0 || step.Counts > 3 {
		report("%s: extra_dice must not be negative, and counts is 0 to 3", path)
	}
	if step.Penalty != "" && step.Penalty != PenaltyShared {
		report("%s: penalty %q: want shared, or leave it unset", path, step.Penalty)
	}
	if step.SuccessDamage != "" && (step.SuccessDamage != SuccessFull || !step.Basic) {
		report("%s: success_damage: want full, on a basic save", path)
	}
	prepareDamage(path, step.Damage, report)
	prepareDamage(path+".extra", step.ExtraDamage, report)
	creature.checkOutcomes(path, step.Outcomes, slices.Contains(rolledSteps, kind), report)
}

func (creature *Creature) checkOutcomes(path string, outcomes Outcomes, rolled bool, report func(string, ...any)) {
	for degree, list := range outcomes.lists() {
		word := strings.ReplaceAll(DegreeWords[degree], "-", "_")
		if len(list) > 0 && !rolled {
			report("%s: %s needs a strike, save, spell_attack or check", path, word)
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

func (step Step) clone() Step {
	step.Damage = slices.Clone(step.Damage)
	step.ExtraDamage = slices.Clone(step.ExtraDamage)
	step.CriticalSuccess = cloneSteps(step.CriticalSuccess)
	step.Success = cloneSteps(step.Success)
	step.Failure = cloneSteps(step.Failure)
	step.CriticalFailure = cloneSteps(step.CriticalFailure)
	return step
}

func cloneSteps(steps []Step) []Step {
	if steps == nil {
		return nil
	}
	cloned := make([]Step, len(steps))
	for index, step := range steps {
		cloned[index] = step.clone()
	}
	return cloned
}
