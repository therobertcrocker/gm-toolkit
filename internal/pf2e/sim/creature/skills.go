package creature

import (
	"maps"
	"slices"
)

const (
	SkillCheckSelector  = "skill-check"
	ProficiencyModifier = "proficiency"
)

var (
	skillAttributes = map[string]string{
		"acrobatics": "dex", "arcana": "int", "athletics": "str", "crafting": "int",
		"deception": "cha", "diplomacy": "cha", "intimidation": "cha", "medicine": "wis",
		"nature": "wis", "occultism": "int", "performance": "cha", "religion": "wis",
		"society": "int", "stealth": "dex", "survival": "wis", "thievery": "dex",
	}
	SkillNames = slices.Sorted(maps.Keys(skillAttributes))
)

func IsSkill(name string) bool {
	_, found := skillAttributes[name]
	return found
}

func IsSkillSelector(selector string) bool {
	return IsSkill(selector) || selector == SkillCheckSelector
}

func SkillAttribute(name string) string {
	return skillAttributes[name]
}

func (build *Build) skills(level int, attributes map[string]int, armor Armor) map[string]int {
	skills := make(map[string]int, len(SkillNames))
	for _, name := range SkillNames {
		attribute := skillAttributes[name]
		rank := max(slices.Index(RankWords, build.Skills[name]), 0)
		// deliberate: the higher of the rank's bonus and a proficiency modifier counts, not both, because bonuses of one type do not stack and the rank's own bonus is a proficiency bonus.
		proficiency := max(proficiencyBonus(level, rank), build.proficiencyModifier(level, name, SkillCheckSelector))
		modifier := proficiency + attributes[attribute] + build.total(level, 0, nil, name, SkillCheckSelector)
		// deliberate: the armor check penalty lands on Strength and Dexterity skills only, not on all skills, because that is the armor rule.
		if attribute == "str" || attribute == "dex" {
			modifier += armor.CheckPenalty
		}
		skills[name] = modifier
	}
	return skills
}

func (build *Build) proficiencyModifier(level int, selectors ...string) int {
	highest := 0
	for _, modifier := range build.Modifiers {
		if modifier.Type == ProficiencyModifier && slices.Contains(selectors, modifier.Selector) {
			highest = max(highest, modifier.Value+modifier.PerLevel*level)
		}
	}
	return highest
}

func (creature *Creature) checkSkills(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.Skills)) {
		if !IsSkill(name) {
			report("skill.%s: not a skill", name)
		}
	}
}
