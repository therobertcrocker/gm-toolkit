package foundry

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func newSkillImporter() *pcImporter {
	return &pcImporter{
		source:      &actor{Name: "Tester"},
		report:      &Report{},
		level:       14,
		build:       &creature.Build{Modifiers: map[string]creature.Modifier{}},
		ranks:       map[string]int{},
		modifierIDs: map[string]bool{},
	}
}

func TestImportPCSkills(t *testing.T) {
	for _, test := range []struct {
		pc   string
		want map[string]string
	}{
		{"kas", map[string]string{"intimidation": "master", "athletics": "master", "acrobatics": "master"}},
		{"tavish", map[string]string{"crafting": "master", "deception": "master", "stealth": "master"}},
		{"arabesque", map[string]string{"performance": "expert", "acrobatics": "expert"}},
	} {
		layer, _ := importTestPC(t, test.pc)
		for _, name := range slices.Sorted(maps.Keys(test.want)) {
			if got := layer.Build.Skills[name]; got != test.want[name] {
				t.Errorf("%s %s = %q, want %q", test.pc, name, got, test.want[name])
			}
		}
		for name, rank := range layer.Build.Skills {
			if !slices.Contains(creature.RankWords[1:], rank) {
				t.Errorf("%s %s = %q, want only trained or better skills written", test.pc, name, rank)
			}
		}
		if _, found := layer.Build.Proficiency["skill:intimidation"]; found {
			t.Errorf("%s proficiency holds a skill key", test.pc)
		}
	}
	layer, _ := importTestPC(t, "kas")
	if _, found := layer.Build.Skills["arcana"]; found {
		t.Error("kas arcana is written, want no entry for an untrained skill")
	}
	arabesque, _ := importTestPC(t, "arabesque")
	if got := arabesque.Build.Skills["arcana"]; got != "trained" {
		t.Errorf("arabesque arcana = %q, want trained from Bloodline: Imperial's rank upgrade", got)
	}
}

func TestImportPCSkillsReportsBadRanks(t *testing.T) {
	importer := newSkillImporter()
	importer.skills(map[string]skillRank{"stealth": {Rank: 2}, "piloting": {Rank: 1}, "arcana": {Rank: 5}, "society": {Rank: -1}})
	if importer.ranks[skillRankPrefix+"stealth"] != 2 {
		t.Errorf("stealth rank = %d, want 2", importer.ranks[skillRankPrefix+"stealth"])
	}
	for _, what := range []string{"skill piloting rank 1", "skill arcana rank 5", "skill society rank -1"} {
		if !slices.Contains(importer.report.Skipped, Skip{Item: "Tester", What: what, Reason: ReasonUnsupported}) {
			t.Errorf("report lacks %q as unsupported: %+v", what, importer.report.Skipped)
		}
	}
	if len(importer.ranks) != 1 {
		t.Errorf("ranks = %v, want stealth only", importer.ranks)
	}
}

func TestImportTrainedSkills(t *testing.T) {
	tavish, tavishReport := importTestPC(t, "tavish")
	if got := tavish.Build.Skills["thievery"]; got != "trained" {
		t.Errorf("tavish thievery = %q, want trained from Saboteur", got)
	}
	if got := tavish.Build.Skills["crafting"]; got != "master" {
		t.Errorf("tavish crafting = %q, want master: the Alchemist's trained crafting does not lower it", got)
	}
	kas, kasReport := importTestPC(t, "kas")
	if got := kas.Build.Skills["intimidation"]; got != "master" {
		t.Errorf("kas intimidation = %q, want master: the Warrior's trained intimidation does not lower it", got)
	}
	for _, want := range []struct {
		report *Report
		skip   Skip
	}{
		{tavishReport, Skip{Item: "Saboteur", What: "trained lore Engineering", Reason: ReasonNotCombat}},
		{kasReport, Skip{Item: "Warrior", What: "trained lore Warfare Lore", Reason: ReasonNotCombat}},
	} {
		if !slices.Contains(want.report.Skipped, want.skip) {
			t.Errorf("report lacks %+v", want.skip)
		}
	}
	if slices.ContainsFunc(tavishReport.Skipped, func(skip Skip) bool { return skip.Item == "Alchemist" || skip.What == "trained skill " }) {
		t.Errorf("report skips the Alchemist's trained skills or an empty custom: %+v", tavishReport.Skipped)
	}

	importer := newSkillImporter()
	importer.raise(skillRankPrefix+"stealth", 3)
	importer.trainedSkills("Odd Background", trainedSkills{
		Value:      []string{"stealth", "piloting", "medicine"},
		Additional: 3,
		Custom:     "Basket Weaving",
	})
	if want := map[string]int{skillRankPrefix + "stealth": 3, skillRankPrefix + "medicine": 1}; !maps.Equal(importer.ranks, want) {
		t.Errorf("ranks = %v, want %v", importer.ranks, want)
	}
	wantSkipped := []Skip{
		{Item: "Odd Background", What: "trained skill piloting", Reason: ReasonUnsupported},
		{Item: "Odd Background", What: "trained skill Basket Weaving", Reason: ReasonUnsupported},
	}
	if !slices.Equal(importer.report.Skipped, wantSkipped) {
		t.Errorf("Skipped = %+v, want %+v", importer.report.Skipped, wantSkipped)
	}
}

func TestImportSkillModifiers(t *testing.T) {
	layer, report := importTestPC(t, "kas")
	if got, want := layer.Build.Modifiers["demon-mask"], (creature.Modifier{Selector: "intimidation", Type: "item", Value: 2}); !reflect.DeepEqual(got, want) {
		t.Errorf("demon-mask = %+v, want %+v", got, want)
	}
	for _, want := range []Skip{
		{Item: "Intimidating Prowess", What: "FlatModifier intimidation", Reason: ReasonConditional},
		{Item: "Feather Steps", What: "FlatModifier athletics", Reason: ReasonConditional},
	} {
		if !slices.Contains(report.Skipped, want) {
			t.Errorf("report lacks %+v", want)
		}
	}
	if !slices.ContainsFunc(report.Skipped, func(skip Skip) bool {
		return skip.Item == "Intimidating Prowess" && skip.Reason == ReasonUnsupported && strings.HasPrefix(skip.What, "AdjustModifier intimidation")
	}) {
		t.Errorf("report lacks Intimidating Prowess's AdjustModifier as unsupported: %+v", report.Skipped)
	}
}

func TestImportProficiencyModifier(t *testing.T) {
	const improvisation = "match(when(lte(@actor.level, 6), floor(@actor.level / 2)), when(gte(@actor.level, 7), @actor.level))"
	importer := newSkillImporter()
	current := &item{Name: "Untrained Improvisation"}
	options := rollOptions{flags: map[string]bool{}}
	importer.flatModifier(current, map[string]any{"key": "FlatModifier", "selector": "skill-check", "type": "proficiency", "value": improvisation}, options)
	importer.flatModifier(&item{Name: "Sharp Eyes"}, map[string]any{"key": "FlatModifier", "selector": "perception", "type": "proficiency", "value": 3.0}, options)
	importer.flatModifier(&item{Name: "Lore Study"}, map[string]any{"key": "FlatModifier", "selector": "arcana", "type": "ability", "value": 3.0}, options)
	want := map[string]creature.Modifier{"untrained-improvisation": {Selector: "skill-check", Type: "proficiency", Value: 14}}
	if !reflect.DeepEqual(importer.build.Modifiers, want) {
		t.Errorf("Modifiers = %+v, want %+v", importer.build.Modifiers, want)
	}
	wantSkipped := []Skip{
		{Item: "Sharp Eyes", What: "FlatModifier perception (proficiency)", Reason: ReasonNotCombat},
		{Item: "Lore Study", What: "FlatModifier arcana (ability)", Reason: ReasonNotCombat},
	}
	if !slices.Equal(importer.report.Skipped, wantSkipped) {
		t.Errorf("Skipped = %+v, want %+v", importer.report.Skipped, wantSkipped)
	}
}

func TestImportCustomModifiers(t *testing.T) {
	layer, _ := importTestPC(t, "tavish")
	if got, want := layer.Build.Modifiers["shadow-property"], (creature.Modifier{Selector: "stealth", Type: "item", Value: 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("shadow-property = %+v, want %+v", got, want)
	}

	var source map[string][]customModifier
	raw := `{
		"ac": [],
		"initiative": [{"label": "Lucky", "modifier": 2, "type": "circumstance", "enabled": true, "ignored": false, "predicate": []}],
		"skill-check": [
			{"label": "Focus", "modifier": 1, "type": "status", "enabled": true, "ignored": false, "predicate": [], "adjustments": [], "traits": [], "ability": null, "damageType": null, "damageCategory": null, "critical": null, "kind": "bonus", "force": false},
			{"label": "Off", "modifier": 1, "type": "item", "enabled": false, "ignored": false, "predicate": []},
			{"label": "Stacked Away", "modifier": 1, "type": "item", "enabled": true, "ignored": true, "predicate": []}
		],
		"athletics": [
			{"label": "Grip", "modifier": 2, "type": "item", "enabled": true, "ignored": false, "predicate": ["action:grapple"]},
			{"label": "Bent", "modifier": 2, "type": "item", "enabled": true, "ignored": false, "predicate": [], "adjustments": [{"slug": "x"}]},
			{"label": "Brawn", "modifier": 2, "type": "ability", "enabled": true, "ignored": false, "predicate": [], "ability": "str"},
			{"label": "Clumsy Gloves", "modifier": -1, "type": "item", "enabled": true, "ignored": false, "predicate": [], "kind": "penalty"}
		]
	}`
	if err := json.Unmarshal([]byte(raw), &source); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	importer := newSkillImporter()
	importer.customModifiers(source)
	want := map[string]creature.Modifier{
		"focus":         {Selector: "skill-check", Type: "status", Value: 1},
		"clumsy-gloves": {Selector: "athletics", Type: "item", Value: -1},
	}
	if !reflect.DeepEqual(importer.build.Modifiers, want) {
		t.Errorf("Modifiers = %+v, want %+v", importer.build.Modifiers, want)
	}
	wantSkipped := []Skip{
		{Item: "Grip", What: "custom modifier athletics", Reason: ReasonConditional},
		{Item: "Bent", What: "custom modifier athletics (item)", Reason: ReasonUnsupported},
		{Item: "Brawn", What: "custom modifier athletics (ability)", Reason: ReasonUnsupported},
		{Item: "Lucky", What: "custom modifier initiative", Reason: ReasonUnsupported},
		{Item: "Off", What: "custom modifier skill-check", Reason: ReasonToggle},
		{Item: "Stacked Away", What: "custom modifier skill-check", Reason: ReasonToggle},
	}
	if !slices.Equal(importer.report.Skipped, wantSkipped) {
		t.Errorf("Skipped = %+v, want %+v", importer.report.Skipped, wantSkipped)
	}
}

func TestImportCheckPenalty(t *testing.T) {
	kas, _ := importTestPC(t, "kas")
	if got := kas.Build.Armor.CheckPenalty; got != 0 {
		t.Errorf("kas check penalty = %d, want 0: Strength +5 meets Full Plate's 4", got)
	}
	tavish, _ := importTestPC(t, "tavish")
	if got := tavish.Build.Armor.CheckPenalty; got != -1 {
		t.Errorf("tavish check penalty = %d, want -1: Strength +0 is below Studded Leather's 1", got)
	}

	importer := &pcImporter{build: tavish.Build, armor: &armorSystem{Strength: 0, CheckPenalty: -1}}
	if got := importer.checkPenalty(); got != 0 {
		t.Errorf("check penalty at a threshold equal to Strength = %d, want 0", got)
	}
	importer.armor = &armorSystem{Strength: 1, CheckPenalty: -1}
	if got := importer.checkPenalty(); got != -1 {
		t.Errorf("check penalty at a threshold above Strength = %d, want -1", got)
	}
}
