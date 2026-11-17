package foundry

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var (
	statSelectors = map[string]string{
		"hp": "hp", "ac": "ac", "fortitude": "fortitude", "reflex": "reflex", "will": "will",
		"saving-throw": "saving-throw", "perception": "perception", "land-speed": "land-speed",
	}
	weaponStatSelectors = map[string]string{
		"attack": "attack", "strike-attack-roll": "attack", "weapon-damage": "damage", "strike-damage": "damage",
	}
	laterSelectors = []string{"initiative", "unarmed-attack", "unarmed-damage", "alchemical-bomb-damage", "melee-strike-damage"}
	bonusTypes     = map[string]string{"": "", "untyped": "", "item": "item", "status": "status", "circumstance": "circumstance"}
	laterRuleKeys  = []string{"AdjustDegreeOfSuccess", "DamageDice", "CriticalSpecialization", "AdjustStrike", "Strike", "Weakness", "Immunity"}
	unusedRuleKeys = []string{"Note", "GrantItem", "ChoiceSet", "Sense", "TokenLight", "TokenEffectIcon", "ItemAlteration", "CraftingAbility"}
	physicalTypes  = []string{"weapon", "armor", "equipment", "consumable", "shield", "backpack", "treasure", "ammo", "book"}
	unreadTypes    = []string{"condition", "effect", "spell"}
)

func (importer *pcImporter) applyRules() {
	options := importer.actorOptions()
	for index := range importer.source.Items {
		current := &importer.source.Items[index]
		if len(current.System.Rules) == 0 || slices.Contains(unreadTypes, current.Type) {
			continue
		}
		if _, handled := specializationSlugs[current.System.Slug]; handled {
			continue
		}
		if !rulesActive(current) {
			importer.report.skip(current.Name, fmt.Sprintf("%d rule elements", len(current.System.Rules)), ReasonInactive)
			continue
		}
		for _, rule := range current.System.Rules {
			importer.applyRule(current, rule, options)
		}
	}
}

func rulesActive(current *item) bool {
	if !slices.Contains(physicalTypes, current.Type) {
		return true
	}
	equipped := current.System.Equipped
	carried := equipped.CarryType == "worn" || equipped.CarryType == "held"
	return carried && (equipped.Invested == nil || *equipped.Invested)
}

func (importer *pcImporter) actorOptions() rollOptions {
	options := rollOptions{flags: map[string]bool{}}
	for _, current := range importer.source.Items {
		switch current.Type {
		case "class":
			options.flags["class:"+current.System.Slug] = true
		case "feat":
			options.flags["feat:"+current.System.Slug] = true
		case "heritage":
			options.flags["self:heritage:"+current.System.Slug] = true
		}
	}
	if armor := importer.armor; armor != nil {
		options.flags["armor:category:"+armor.Category] = true
		options.flags["armor:group:"+armor.Group] = true
	}
	return options
}

func (importer *pcImporter) applyRule(current *item, rule map[string]any, options rollOptions) {
	key, _ := rule["key"].(string)
	switch {
	case key == "FlatModifier":
		importer.flatModifier(current, rule, options)
	case key == "Resistance":
		importer.resistance(current, rule, options)
	case key == "ActiveEffectLike":
		importer.activeEffectLike(current, rule, options)
	case key == "MartialProficiency":
		importer.martialProficiency(current, rule)
	case key == "AdjustModifier":
		importer.adjustModifier(current, rule)
	case key == "DexterityModifierCap":
		importer.dexterityCap(current, rule)
	case key == "RollOption":
		importer.rollOption(current, rule)
	case slices.Contains(laterRuleKeys, key):
		importer.report.skip(current.Name, key, ReasonEffort3)
	case slices.Contains(unusedRuleKeys, key):
		importer.report.unused(key)
	default:
		importer.report.skip(current.Name, key, ReasonUnsupported)
	}
}

func (importer *pcImporter) flatModifier(current *item, rule map[string]any, options rollOptions) {
	for _, field := range []string{"damageCategory", "damageType", "critical"} {
		if _, found := rule[field]; found {
			importer.report.skip(current.Name, "FlatModifier with "+field, ReasonEffort3)
			return
		}
	}
	selectors := textList(rule["selector"])
	if len(selectors) == 0 {
		importer.report.skip(current.Name, "FlatModifier without a selector", ReasonUnsupported)
		return
	}
	for _, selector := range selectors {
		importer.flatModifierFor(current, rule, selector, options)
	}
}

func (importer *pcImporter) flatModifierFor(current *item, rule map[string]any, selector string, options rollOptions) {
	what := "FlatModifier " + selector
	stat, isStat := statSelectors[selector]
	weaponStat, isWeaponStat := weaponStatSelectors[selector]
	switch {
	case slices.Contains(laterSelectors, selector):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	case !isStat && !isWeaponStat:
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	kind, _ := rule["type"].(string)
	bonusType, isBonusType := bonusTypes[kind]
	if !isBonusType {
		importer.report.skip(current.Name, what+" ("+kind+")", ReasonNotCombat)
		return
	}
	modifier := creature.Modifier{Selector: stat, Type: bonusType}
	predicate, _ := rule["predicate"].([]any)
	if isWeaponStat {
		traits, applies := importer.weaponTraits(predicate)
		if !applies {
			importer.report.skip(current.Name, what, ReasonConditional)
			return
		}
		modifier.Selector = weaponStat
		modifier.WeaponTraits = traits
	} else {
		held, err := options.holds(predicate)
		if err != nil {
			importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
			return
		}
		if !held {
			importer.report.skip(current.Name, what, ReasonConditional)
			return
		}
	}
	if text, isText := rule["value"].(string); isText && strings.TrimSpace(text) == "@actor.level" {
		modifier.PerLevel = 1
	} else {
		value, err := importer.number(rule["value"])
		if err != nil {
			importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
			return
		}
		modifier.Value = value
	}
	importer.build.Modifiers[uniqueID(current.Name, importer.modifierIDs)] = modifier
}

func (importer *pcImporter) weaponTraits(predicate []any) ([]string, bool) {
	var traits []string
	for _, term := range predicate {
		text, isText := term.(string)
		if !isText {
			return nil, false
		}
		trait := strings.TrimPrefix(text, "item:trait:")
		carried := false
		for _, weapon := range importer.build.Weapons {
			carried = carried || slices.Contains(weapon.Traits, trait)
		}
		if !carried {
			return nil, false
		}
		traits = append(traits, trait)
	}
	return traits, true
}

func (importer *pcImporter) resistance(current *item, rule map[string]any, options rollOptions) {
	kindText, _ := rule["type"].(string)
	kind, err := resolveTemplates(kindText, current)
	what := "Resistance " + kind
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	held, err := options.holds(predicate)
	switch {
	case err != nil:
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	case !creature.IsDamageType(kind):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	}
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	if value > importer.resist[kind].Value {
		importer.resist[kind] = creature.Adjustment{Value: value}
	}
}

func (importer *pcImporter) activeEffectLike(current *item, rule map[string]any, options rollOptions) {
	path, _ := rule["path"].(string)
	what := "ActiveEffectLike " + path
	key, isRank := rankKey(path)
	switch {
	case strings.HasPrefix(path, "system.attributes.dying."):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	case !isRank:
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	if mode, _ := rule["mode"].(string); mode != "upgrade" {
		importer.report.skip(current.Name, what+" mode "+mode, ReasonUnsupported)
		return
	}
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
	rank, err := importer.number(rule["value"])
	if err != nil || rank < 0 || rank >= len(creature.RankWords) {
		importer.report.skip(current.Name, fmt.Sprintf("%s value %v", what, rule["value"]), ReasonUnsupported)
		return
	}
	importer.raise(key, rank)
}

func rankKey(path string) (string, bool) {
	inner, isSystem := strings.CutPrefix(path, "system.")
	inner, isRank := strings.CutSuffix(inner, ".rank")
	if !isSystem || !isRank {
		return "", false
	}
	switch {
	case inner == "perception":
		return "perception", true
	case inner == "proficiencies.spellcasting":
		return "spellcasting", true
	case strings.HasPrefix(inner, "saves."):
		key := strings.TrimPrefix(inner, "saves.")
		return key, creature.IsProficiencyKey(key)
	}
	for _, prefix := range []string{"proficiencies.attacks.", "proficiencies.defenses."} {
		if key, found := strings.CutPrefix(inner, prefix); found {
			return key, creature.IsProficiencyKey(key)
		}
	}
	return "", false
}

func (importer *pcImporter) martialProficiency(current *item, rule map[string]any) {
	definition, _ := rule["definition"].([]any)
	var group string
	var categories []string
	collect := func(text string) {
		if rest, found := strings.CutPrefix(text, "item:group:"); found {
			group = rest
		}
		if rest, found := strings.CutPrefix(text, "item:category:"); found {
			categories = append(categories, rest)
		}
	}
	for _, term := range definition {
		switch typed := term.(type) {
		case string:
			resolved, err := resolveTemplates(typed, current)
			if err != nil {
				importer.report.skip(current.Name, "MartialProficiency: "+err.Error(), ReasonUnsupported)
				return
			}
			collect(resolved)
		case map[string]any:
			alternatives, _ := typed["or"].([]any)
			for _, alternative := range alternatives {
				if text, isText := alternative.(string); isText {
					collect(text)
				}
			}
		}
	}
	what := fmt.Sprintf("MartialProficiency group %q, categories %s", group, strings.Join(categories, ", "))
	rank, err := importer.number(rule["value"])
	if group == "" || !slices.Contains(categories, "martial") || err != nil || rank < 0 || rank >= len(creature.RankWords) {
		importer.report.skip(current.Name, what, ReasonUnsupported)
		return
	}
	importer.raise(creature.WeaponGroupPrefix+group, rank)
}

func (importer *pcImporter) adjustModifier(current *item, rule map[string]any) {
	slug, _ := rule["slug"].(string)
	if suppress, _ := rule["suppress"].(bool); suppress && slug == "armor-speed-penalty" {
		importer.speedPenaltySuppressed = true
		return
	}
	selectors := append(textList(rule["selector"]), textList(rule["selectors"])...)
	importer.report.skip(current.Name, "AdjustModifier "+strings.Join(selectors, ", "), ReasonUnsupported)
}

func (importer *pcImporter) dexterityCap(current *item, rule map[string]any) {
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, "DexterityModifierCap: "+err.Error(), ReasonUnsupported)
		return
	}
	importer.dexterityCaps = append(importer.dexterityCaps, value)
}

func (importer *pcImporter) rollOption(current *item, rule map[string]any) {
	if toggleable, found := rule["toggleable"]; !found || toggleable == false {
		importer.report.unused("RollOption")
		return
	}
	option, _ := rule["option"].(string)
	state := "off"
	if stored, _ := rule["value"].(bool); stored {
		state = "stored on"
	}
	importer.report.skip(current.Name, fmt.Sprintf("toggle %s (%s)", option, state), ReasonToggle)
}

func (importer *pcImporter) number(value any) (int, error) {
	switch typed := value.(type) {
	case float64:
		if typed != math.Trunc(typed) {
			return 0, fmt.Errorf("%w: %v is not a whole number", errUnsupportedFormula, typed)
		}
		return int(typed), nil
	case string:
		return evaluateFormula(typed, importer.variables())
	}
	return 0, fmt.Errorf("%w: value %v", errUnsupportedFormula, value)
}

func (importer *pcImporter) variables() map[string]float64 {
	variables := map[string]float64{"@actor.level": float64(importer.level)}
	if importer.armor != nil {
		variables["@armor.system.runes.potency"] = float64(importer.armor.Runes.Potency)
	}
	return variables
}

func textList(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		var texts []string
		for _, entry := range typed {
			if text, isText := entry.(string); isText {
				texts = append(texts, text)
			}
		}
		return texts
	}
	return nil
}
