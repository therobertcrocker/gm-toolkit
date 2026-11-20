package foundry

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const skillRankPrefix = "skill:"

type skillRank struct {
	Rank int `json:"rank"`
}

type trainedSkills struct {
	Value      []string `json:"value"`
	Lore       []string `json:"lore"`
	Additional int      `json:"additional"`
	Custom     string   `json:"custom"`
}

type customModifier struct {
	Label          string `json:"label"`
	Modifier       any    `json:"modifier"`
	Type           string `json:"type"`
	Enabled        bool   `json:"enabled"`
	Ignored        bool   `json:"ignored"`
	Force          bool   `json:"force"`
	Predicate      []any  `json:"predicate"`
	Adjustments    []any  `json:"adjustments"`
	Traits         []any  `json:"traits"`
	Ability        any    `json:"ability"`
	DamageType     any    `json:"damageType"`
	DamageCategory any    `json:"damageCategory"`
	Critical       any    `json:"critical"`
}

func (importer *pcImporter) skills(source map[string]skillRank) {
	for _, name := range slices.Sorted(maps.Keys(source)) {
		rank := source[name].Rank
		if !creature.IsSkill(name) || rank < 0 || rank >= len(creature.RankWords) {
			importer.report.skip(importer.source.Name, fmt.Sprintf("skill %s rank %d", name, rank), ReasonUnsupported)
			continue
		}
		importer.raise(skillRankPrefix+name, rank)
	}
}

func (importer *pcImporter) trainedSkills(owner string, trained trainedSkills) {
	for _, name := range trained.Value {
		if !creature.IsSkill(name) {
			importer.report.skip(owner, "trained skill "+name, ReasonUnsupported)
			continue
		}
		importer.raise(skillRankPrefix+name, 1)
	}
	for _, lore := range trained.Lore {
		importer.report.skip(owner, "trained lore "+lore, ReasonNotCombat)
	}
	if custom := strings.TrimSpace(trained.Custom); custom != "" {
		importer.report.skip(owner, "trained skill "+custom, ReasonUnsupported)
	}
}

func (importer *pcImporter) customModifiers(source map[string][]customModifier) {
	for _, selector := range slices.Sorted(maps.Keys(source)) {
		for _, entry := range source[selector] {
			importer.customModifier(selector, entry)
		}
	}
}

func (importer *pcImporter) customModifier(selector string, entry customModifier) {
	owner := entry.Label
	if owner == "" {
		owner = importer.source.Name
	}
	what := "custom modifier " + selector
	bonusType, isBonusType := bonusTypes[entry.Type]
	if entry.Type == creature.ProficiencyModifier {
		bonusType, isBonusType = entry.Type, true
	}
	switch {
	case !creature.IsSkillSelector(selector):
		importer.report.skip(owner, what, ReasonUnsupported)
		return
	case !entry.Enabled || entry.Ignored:
		importer.report.skip(owner, what, ReasonToggle)
		return
	case len(entry.Predicate) > 0:
		importer.report.skip(owner, what, ReasonConditional)
		return
	case !isBonusType || entry.Force || len(entry.Adjustments) > 0 || len(entry.Traits) > 0 ||
		entry.Ability != nil || entry.DamageType != nil || entry.DamageCategory != nil || entry.Critical != nil:
		importer.report.skip(owner, fmt.Sprintf("%s (%s)", what, entry.Type), ReasonUnsupported)
		return
	}
	value, err := importer.number(entry.Modifier)
	if err != nil {
		importer.report.skip(owner, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	importer.build.Modifiers[uniqueID(owner, importer.modifierIDs)] = creature.Modifier{Selector: selector, Type: bonusType, Value: value}
}

// deliberate: the check penalty is dropped when the Strength threshold is met, not reduced, because that is the armor rule; the speed penalty is the one that is only reduced.
func (importer *pcImporter) checkPenalty() int {
	system := importer.armor
	attributes, err := importer.build.Attributes()
	if system.CheckPenalty >= 0 || (err == nil && attributes["str"] >= system.Strength) {
		return 0
	}
	return system.CheckPenalty
}
