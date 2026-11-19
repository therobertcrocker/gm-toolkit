package foundry

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestImportPCRuleElementsKas(t *testing.T) {
	layer, report := importTestPC(t, "kas")
	build := layer.Build
	checkRanks(t, build, map[string]string{"will": "expert", "perception": "master", "weapon-group-hammer": "legendary"})
	if build.Armor.SpeedPenalty != 0 {
		t.Errorf("armor speed penalty = %d, want 0 (Unburdened Iron)", build.Armor.SpeedPenalty)
	}
	wantModifiers := map[string]creature.Modifier{
		"toughness":                   {Selector: "hp", PerLevel: 1},
		"mountains-stoutness":         {Selector: "hp", PerLevel: 1},
		"fleet":                       {Selector: "land-speed", Value: 5},
		"amulet-of-kinships-strength": {Selector: "will", Type: "item", Value: 2},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	if want := map[string]creature.Adjustment{"slashing": {Value: 4}}; !reflect.DeepEqual(layer.Defense.Resist, want) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, want)
	}
	for _, want := range []struct{ item, reason string }{
		{"Keen Eyes", ReasonConditional},
		{"Vicious Swing", ReasonToggle},
		{"Demon Mask", ReasonNotCombat},
		{"Weapon Legend (Hammer)", ReasonUnsupported},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
	for _, imported := range []string{"Bravery", "Toughness"} {
		if slices.ContainsFunc(report.Skipped, func(skip Skip) bool { return skip.Item == imported }) {
			t.Errorf("report skips %s, want its rules imported", imported)
		}
	}
	if report.Unused["Note"] == 0 {
		t.Errorf("Unused = %v, want Note counted", report.Unused)
	}
}

func TestImportPCRuleElementsTavish(t *testing.T) {
	layer, report := importTestPC(t, "tavish")
	build := layer.Build
	checkRanks(t, build, map[string]string{"will": "expert", "simple": "expert", "weapon-base-alchemical-bomb": "expert", "class-dc": "expert"})
	wantModifiers := map[string]creature.Modifier{
		"alchemist-goggles":         {Selector: "attack", Type: "item", Value: 1, WeaponTraits: []string{"alchemical", "bomb"}},
		"boots-of-bounding-greater": {Selector: "land-speed", Type: "item", Value: 10},
		"expanded-splash":           {Selector: "splash", Attribute: "int", WeaponTraits: []string{"splash"}},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	wantResist := map[string]creature.Adjustment{"poison": {Value: 7}, "electricity": {Value: 5}}
	if !reflect.DeepEqual(layer.Defense.Resist, wantResist) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, wantResist)
	}
	if !slices.Contains(report.Skipped, Skip{Item: "Sticky Bomb", What: "toggle sticky-bomb (stored on)", Reason: ReasonToggle}) {
		t.Error("report lacks Sticky Bomb as a toggle stored on")
	}
	if !hasSkip(report, "Torch", ReasonInactive) {
		t.Error("report lacks the stowed Torch as inactive")
	}
}

func TestImportPCRuleElementsArabesque(t *testing.T) {
	layer, report := importTestPC(t, "arabesque")
	build := layer.Build
	wantModifiers := map[string]creature.Modifier{
		"bands-of-force-greater":   {Selector: "ac", Type: "item", Value: 2},
		"bands-of-force-greater-2": {Selector: "saving-throw", Type: "item", Value: 2},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	if armor := build.Armor; armor == nil || armor.Category != "unarmored" || armor.DexCap == nil || *armor.DexCap != 5 {
		t.Errorf("Armor = %+v, want unarmored with dex cap 5", armor)
	}
	if layer.Defense.Resist != nil {
		t.Errorf("Resist = %+v, want none", layer.Defense.Resist)
	}
	for _, want := range []struct{ item, reason string }{
		{"Bands of Force", ReasonInactive},
		{"Diverting Vortex", ReasonConditional},
		{"Diverting Vortex", ReasonToggle},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}
