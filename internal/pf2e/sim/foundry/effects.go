package foundry

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const (
	baseRecoveryDC  = 10
	foundryDyingMax = 4
)

var (
	degreeKeys       = map[string]string{"criticalFailure": "critical-failure", "failure": "failure", "success": "success", "criticalSuccess": "critical-success"}
	degreeSteps      = map[string]int{"one-degree-better": 1, "one-degree-worse": -1}
	shiftSelectors   = []string{"saving-throw", "fortitude", "reflex", "will"}
	attributeFormula = regexp.MustCompile(`^@actor\.abilities\.(str|dex|con|int|wis|cha)\.mod$`)
)

func (importer *pcImporter) degreeShift(current *item, rule map[string]any, options rollOptions) {
	selector, _ := rule["selector"].(string)
	what := "AdjustDegreeOfSuccess " + selector
	if !slices.Contains(shiftSelectors, selector) {
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	adjustment, _ := rule["adjustment"].(map[string]any)
	if len(adjustment) == 0 {
		importer.report.skip(current.Name, what+" without an adjustment", ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	against, rest := splitTraitTerms(predicate)
	held, err := options.holds(rest)
	switch {
	case err != nil:
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	case len(against) > 1:
		importer.report.skip(current.Name, what+" against "+strings.Join(against, " and "), ReasonUnsupported)
		return
	}
	for _, key := range slices.Sorted(maps.Keys(adjustment)) {
		text, _ := adjustment[key].(string)
		from, knownDegree := degreeKeys[key]
		step, knownStep := degreeSteps[text]
		to := slices.Index(creature.DegreeWords, from) + step
		if !knownDegree || !knownStep || to < 0 || to >= len(creature.DegreeWords) {
			importer.report.skip(current.Name, fmt.Sprintf("%s %s %s", what, key, text), ReasonUnsupported)
			continue
		}
		importer.degreeShifts[uniqueID(current.Name, importer.fightIDs)] = creature.DegreeShift{Selector: selector, Against: against, From: from, To: creature.DegreeWords[to]}
	}
}

// deliberate: a bare trait term ("fear", "item:trait:emotion") names a trait of the effect, not an actor option, because Foundry tests degree adjustments against the incoming roll's traits.
func splitTraitTerms(predicate []any) ([]string, []any) {
	var traits []string
	var rest []any
	for _, term := range predicate {
		text, isText := term.(string)
		trait := strings.TrimPrefix(text, "item:trait:")
		if isText && !strings.Contains(trait, ":") {
			traits = append(traits, trait)
			continue
		}
		rest = append(rest, term)
	}
	return traits, rest
}

func (importer *pcImporter) criticalSpecialization(current *item, rule map[string]any, options rollOptions) {
	predicate, _ := rule["predicate"].([]any)
	minimum := 0
	var rest []any
	for _, term := range predicate {
		if rank, isRank := proficiencyRankAtLeast(term); isRank {
			minimum = rank
			continue
		}
		rest = append(rest, term)
	}
	held, err := options.holds(rest)
	switch {
	case err != nil:
		importer.report.skip(current.Name, "CriticalSpecialization: "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, "CriticalSpecialization", ReasonConditional)
		return
	case minimum < 1 || minimum >= len(creature.RankWords):
		importer.report.skip(current.Name, "CriticalSpecialization without a rank", ReasonUnsupported)
		return
	}
	existing := importer.build.CriticalSpecialization
	if existing == "" || minimum < slices.Index(creature.RankWords, existing) {
		importer.build.CriticalSpecialization = creature.RankWords[minimum]
	}
}

func proficiencyRankAtLeast(term any) (int, bool) {
	comparison, _ := term.(map[string]any)
	operands, found := comparison["gte"].([]any)
	if !found || len(operands) != 2 || operands[0] != "item:proficiency:rank" {
		return 0, false
	}
	rank, isNumber := operands[1].(float64)
	return int(rank), isNumber
}

func (importer *pcImporter) adjustStrike(current *item, rule map[string]any) {
	property, _ := rule["property"].(string)
	value, _ := rule["value"].(string)
	what := "AdjustStrike " + property
	if property != "materials" {
		importer.report.skip(current.Name, what, ReasonLaterRule)
		return
	}
	mode, _ := rule["mode"].(string)
	definition := textList(rule["definition"])
	namesItself := len(definition) == 0 || slices.Equal(definition, []string{"item:id:{item|id}"})
	id, found := importer.weaponByItem[current.ID]
	if !found || mode != "add" || !namesItself || !creature.IsDefenseTrait(value) {
		importer.report.skip(current.Name, what+" "+value, ReasonUnsupported)
		return
	}
	weapon := importer.build.Weapons[id]
	weapon.Traits = append(slices.Clone(weapon.Traits), value)
	importer.build.Weapons[id] = weapon
}

func (importer *pcImporter) initiativeModifier(current *item, rule map[string]any) {
	predicate, _ := rule["predicate"].([]any)
	for _, term := range predicate {
		if term != "perception" {
			importer.report.skip(current.Name, "FlatModifier initiative", ReasonConditional)
			return
		}
	}
	kind, _ := rule["type"].(string)
	bonusType, known := bonusTypes[kind]
	value, err := importer.number(rule["value"])
	if !known || err != nil || value == 0 {
		importer.report.skip(current.Name, fmt.Sprintf("FlatModifier initiative %v", rule["value"]), ReasonUnsupported)
		return
	}
	importer.fightModifiers[uniqueID(current.Name, importer.fightIDs)] = creature.CheckModifier{Selector: "initiative", Type: bonusType, Value: value}
}

func (importer *pcImporter) splashModifier(current *item, rule map[string]any) {
	value, _ := rule["value"].(string)
	match := attributeFormula.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		importer.report.skip(current.Name, "FlatModifier splash "+value, ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	traits, applies := importer.weaponTraits(predicate)
	if !applies {
		importer.report.skip(current.Name, "FlatModifier splash", ReasonConditional)
		return
	}
	importer.build.Modifiers[uniqueID(current.Name, importer.modifierIDs)] = creature.Modifier{Selector: "splash", Attribute: match[1], WeaponTraits: traits}
}

func (importer *pcImporter) dyingRule(current *item, rule map[string]any, options rollOptions, path string) {
	what := "ActiveEffectLike " + path
	predicate, _ := rule["predicate"].([]any)
	held, err := options.holds(predicate)
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	if !held {
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	}
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	mode, _ := rule["mode"].(string)
	switch {
	case path == "system.attributes.dying.max" && mode == "upgrade":
		importer.dyingMax = max(cmp.Or(importer.dyingMax, foundryDyingMax), value)
	case path == "system.attributes.dying.recoveryDC" && mode == "subtract":
		importer.recoverySubtract += value
	case path == "system.attributes.dying.recoveryDC" && mode == "downgrade":
		if importer.recoveryFloor == 0 || value < importer.recoveryFloor {
			importer.recoveryFloor = value
		}
	default:
		importer.report.skip(current.Name, what+" mode "+mode, ReasonUnsupported)
	}
}

// deliberate: subtractions apply before downgrades, not in item order, because Foundry applies ActiveEffectLike modes by priority (subtract before downgrade).
func (importer *pcImporter) recoveryDC() int {
	dc := baseRecoveryDC - importer.recoverySubtract
	if importer.recoveryFloor > 0 {
		dc = min(dc, importer.recoveryFloor)
	}
	return dc
}
