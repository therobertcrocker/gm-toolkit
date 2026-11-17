package foundry

import (
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func importTestPC(t *testing.T, name string) (*creature.Creature, *Report) {
	t.Helper()
	source, err := readActor(filepath.Join("testdata", "actors", name+".json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	layer, report, err := importPC(source)
	if err != nil {
		t.Fatalf("importPC: %v", err)
	}
	return layer, report
}

func checkRanks(t *testing.T, build *creature.Build, want map[string]string) {
	t.Helper()
	for _, key := range slices.Sorted(maps.Keys(want)) {
		if build.Proficiency[key] != want[key] {
			t.Errorf("proficiency %s = %q, want %q", key, build.Proficiency[key], want[key])
		}
	}
}

func TestImportPCBuildChoices(t *testing.T) {
	layer, _ := importTestPC(t, "kas")
	if layer.Name != "Kas Fringebramble" || layer.Kind != creature.KindPC || layer.Level != 14 {
		t.Errorf("header = %q %s %d, want Kas Fringebramble pc 14", layer.Name, layer.Kind, layer.Level)
	}
	if want := []string{"halfling", "humanoid"}; !slices.Equal(layer.Traits, want) {
		t.Errorf("Traits = %v, want %v", layer.Traits, want)
	}
	build := layer.Build
	if build.AncestryHP != 6 || build.ClassHP != 10 || build.Key != "str" || build.Speed != 25 || build.WeaponSpecialization != "normal" {
		t.Errorf("build header = %+v", build)
	}
	if build.Flaws != nil {
		t.Errorf("Flaws = %v, want none: alternate ancestry boosts replace the flaw", build.Flaws)
	}
	wantBoosts := map[string][]string{
		"ancestry":   {"str", "cha"},
		"background": {"con", "str"},
		"class":      {"str"},
		"level-1":    {"cha", "con", "str", "wis"},
		"level-5":    {"str", "con", "cha", "wis"},
		"level-10":   {"str", "con", "cha", "wis"},
	}
	if !reflect.DeepEqual(build.Boosts, wantBoosts) {
		t.Errorf("Boosts = %v, want %v", build.Boosts, wantBoosts)
	}
	checkRanks(t, build, map[string]string{
		"fortitude": "master", "reflex": "expert", "class-dc": "expert", "heavy": "expert",
		"simple": "master", "martial": "master", "unarmed": "master",
	})
	armor := build.Armor
	if armor == nil || armor.Name != "Full Plate" || armor.Category != "heavy" || armor.ACBonus != 6 ||
		armor.DexCap == nil || *armor.DexCap != 0 || armor.Potency != 2 || armor.Resilient != 2 {
		t.Errorf("Armor = %+v, want Full Plate heavy +6, dex cap 0, potency 2, resilient 2", armor)
	}
	wantFury := creature.Weapon{
		Name: "Protector's Fury", Category: "martial", Group: "hammer", Base: "maul",
		Dice: 1, Die: 12, Type: "bludgeoning", Potency: 2, Striking: 2,
		Traits: []string{"gnome", "trip", "versatile-p"},
	}
	if got := build.Weapons["protectors-fury"]; !reflect.DeepEqual(got, wantFury) {
		t.Errorf("protectors-fury = %+v, want %+v", got, wantFury)
	}
}

func TestImportPCSkipsWeaponsTheLoaderRejects(t *testing.T) {
	importer := &pcImporter{
		report:    &Report{},
		build:     &creature.Build{Weapons: map[string]creature.Weapon{}},
		weaponIDs: map[string]bool{},
	}
	for _, test := range []struct{ name, damage string }{
		{"No Dice", `{"dice": 0, "die": "d6", "damageType": "slashing"}`},
		{"One-Sided", `{"dice": 1, "die": "d1", "damageType": "slashing"}`},
		{"Untyped", `{"dice": 1, "die": "d6", "damageType": "untyped"}`},
	} {
		current := &item{
			Name:   test.name,
			Type:   "weapon",
			Raw:    []byte(`{"category": "simple", "damage": ` + test.damage + `}`),
			System: itemCommon{Quantity: 1},
		}
		if err := importer.weapon(current); err != nil {
			t.Fatalf("%s: weapon: %v", test.name, err)
		}
		if !hasSkip(importer.report, test.name, ReasonUnsupported) {
			t.Errorf("report lacks %s with reason %q", test.name, ReasonUnsupported)
		}
	}
	if len(importer.build.Weapons) != 0 {
		t.Errorf("weapons = %v, want none", slices.Sorted(maps.Keys(importer.build.Weapons)))
	}
}

func TestImportPCWeaponsAndQuantities(t *testing.T) {
	layer, report := importTestPC(t, "tavish")
	build := layer.Build
	if build.Key != "int" || !slices.Equal(build.Flaws, []string{"cha"}) {
		t.Errorf("key %q flaws %v, want int and [cha]", build.Key, build.Flaws)
	}
	if want := []string{"con", "wis", "cha"}; !slices.Equal(build.Boosts["ancestry"], want) {
		t.Errorf("ancestry boosts = %v, want %v", build.Boosts["ancestry"], want)
	}
	if len(build.Weapons) != 12 {
		t.Errorf("weapons = %v, want 12 (two bombs have quantity 0)", slices.Sorted(maps.Keys(build.Weapons)))
	}
	wantFire := creature.Weapon{
		Name: "Alchemist's Fire (Greater)", Category: "martial", Group: "bomb", Base: "alchemical-bomb",
		Dice: 3, Die: 8, Type: "fire", ItemBonus: 2, Range: 20,
		Traits: []string{"alchemical", "bomb", "consumable", "fire", "infused", "splash"},
	}
	if got := build.Weapons["alchemists-fire-greater"]; !reflect.DeepEqual(got, wantFire) {
		t.Errorf("alchemists-fire-greater = %+v, want %+v", got, wantFire)
	}
	if _, found := build.Weapons["frost-vial-moderate"]; found {
		t.Error("frost-vial-moderate imported, want it skipped at quantity 0")
	}
	armor := build.Armor
	if armor == nil || armor.Category != "light" || armor.ACBonus != 2 || *armor.DexCap != 3 || armor.Potency != 2 || armor.Resilient != 1 {
		t.Errorf("Armor = %+v, want light +2, dex cap 3, potency 2, resilient 1", armor)
	}
	if build.Spellcasting != nil {
		t.Errorf("Spellcasting = %+v, want none: the innate entry's spells sit nowhere", build.Spellcasting)
	}
	for _, want := range []struct{ item, reason string }{
		{"Frost Vial (Moderate)", ReasonQuantity},
		{"Alchemist's Fire (Greater)", ReasonEffort3},
		{"Arcane Innate Spells", ReasonEmpty},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}

func TestImportPCSpellcasting(t *testing.T) {
	layer, report := importTestPC(t, "arabesque")
	build := layer.Build
	if want := (creature.Casting{Attribute: "cha", TopRank: 7}); build.Spellcasting == nil || *build.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", build.Spellcasting, want)
	}
	if !slices.Equal(build.Flaws, []string{"con"}) {
		t.Errorf("Flaws = %v, want [con]", build.Flaws)
	}
	checkRanks(t, build, map[string]string{
		"spellcasting": "expert", "unarmored": "expert", "fortitude": "expert", "reflex": "expert", "will": "expert", "simple": "expert",
	})
	if blade := build.Weapons["blink-blade"]; blade.Striking != 1 || blade.Potency != 2 || blade.Die != 4 {
		t.Errorf("blink-blade = %+v, want a +2 striking d4", blade)
	}
	if !hasSkip(report, "Desolation Locket", ReasonUnsupported) {
		t.Error("report lacks the Desolation Locket entry")
	}
}

func TestImportPCIgnoresSheetState(t *testing.T) {
	_, kasReport := importTestPC(t, "kas")
	_, arabesqueReport := importTestPC(t, "arabesque")
	for _, want := range []struct {
		report       *Report
		item, reason string
	}{
		{kasReport, "Slowed", ReasonState},
		{kasReport, "Spell Effect: Haste", ReasonState},
		{kasReport, "Bramblecrest (Destroyed)", ReasonEffort3},
		{kasReport, "Arcane Charges Spells", ReasonEmpty},
		{arabesqueReport, "Frightened", ReasonState},
		{arabesqueReport, "Persistent Damage", ReasonState},
	} {
		if !hasSkip(want.report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}
