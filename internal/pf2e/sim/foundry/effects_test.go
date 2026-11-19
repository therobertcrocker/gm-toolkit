package foundry

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestImportDegreeShifts(t *testing.T) {
	layer, report := importTestPC(t, "kas")
	want := map[string]creature.DegreeShift{
		"juggernaut":               {Selector: "fortitude", From: "success", To: "critical-success"},
		"bravery":                  {Selector: "will", Against: []string{"fear"}, From: "success", To: "critical-success"},
		"gutsy-halfling":           {Selector: "saving-throw", Against: []string{"emotion"}, From: "success", To: "critical-success"},
		"irrepressible-halfling":   {Selector: "saving-throw", Against: []string{"emotion"}, From: "success", To: "critical-success"},
		"irrepressible-halfling-2": {Selector: "saving-throw", Against: []string{"emotion"}, From: "critical-failure", To: "failure"},
	}
	if !reflect.DeepEqual(layer.DegreeShifts, want) {
		t.Errorf("DegreeShifts = %+v, want %+v", layer.DegreeShifts, want)
	}
	inextinguishable := Skip{Item: "Inextinguishable", What: "AdjustDegreeOfSuccess saving-throw without an adjustment", Reason: ReasonUnsupported}
	if !slices.Contains(report.Skipped, inextinguishable) {
		t.Errorf("report lacks %+v", inextinguishable)
	}
}

func TestImportDyingLimits(t *testing.T) {
	kas, report := importTestPC(t, "kas")
	if kas.DyingMax != 6 || kas.RecoveryDC != 6 {
		t.Errorf("kas dying max %d, recovery DC %d; want 6 and 6", kas.DyingMax, kas.RecoveryDC)
	}
	stoutness := Skip{Item: "Mountain's Stoutness", What: "ActiveEffectLike system.attributes.dying.recoveryDC", Reason: ReasonConditional}
	if !slices.Contains(report.Skipped, stoutness) {
		t.Errorf("report lacks %+v", stoutness)
	}
	for _, name := range []string{"tavish", "arabesque"} {
		layer, _ := importTestPC(t, name)
		if layer.DyingMax != 0 || layer.RecoveryDC != 0 {
			t.Errorf("%s dying max %d, recovery DC %d; want both unset", name, layer.DyingMax, layer.RecoveryDC)
		}
	}
}

func TestImportInitiative(t *testing.T) {
	layer, _ := importTestPC(t, "kas")
	want := map[string]creature.CheckModifier{
		"battlefield-surveyor": {Selector: "initiative", Type: "circumstance", Value: 2},
	}
	if !reflect.DeepEqual(layer.Modifiers, want) {
		t.Errorf("Modifiers = %+v, want %+v", layer.Modifiers, want)
	}
}

func TestImportCriticalSpecialization(t *testing.T) {
	kas, _ := importTestPC(t, "kas")
	if kas.Build.CriticalSpecialization != "master" {
		t.Errorf("kas critical specialization = %q, want master", kas.Build.CriticalSpecialization)
	}
	arabesque, report := importTestPC(t, "arabesque")
	if arabesque.Build.CriticalSpecialization != "" {
		t.Errorf("arabesque critical specialization = %q, want none", arabesque.Build.CriticalSpecialization)
	}
	expertise := Skip{Item: "Weapon Expertise", What: "CriticalSpecialization", Reason: ReasonConditional}
	if !slices.Contains(report.Skipped, expertise) {
		t.Errorf("report lacks %+v", expertise)
	}
}

func TestImportBombDamage(t *testing.T) {
	layer, report := importTestPC(t, "tavish")
	fire := layer.Build.Weapons["alchemists-fire-greater"]
	if want := (creature.DamagePart{Dice: "3", Type: "fire"}); fire.Persistent == nil || *fire.Persistent != want || fire.Splash != 3 {
		t.Errorf("alchemists-fire-greater persistent %+v, splash %d; want %+v and 3", fire.Persistent, fire.Splash, want)
	}
	want := creature.Modifier{Selector: "splash", Attribute: "int", WeaponTraits: []string{"splash"}}
	if got := layer.Build.Modifiers["expanded-splash"]; !reflect.DeepEqual(got, want) {
		t.Errorf("expanded-splash = %+v, want %+v", got, want)
	}
	for _, skip := range report.Skipped {
		if skip.Item == "Alchemist's Fire (Greater)" && (strings.HasPrefix(skip.What, "persistent") || strings.HasPrefix(skip.What, "splash")) {
			t.Errorf("report skips %+v, want persistent damage and splash imported", skip)
		}
	}
	calculated := Skip{Item: "Calculated Splash", What: "AdjustModifier alchemical-bomb-damage", Reason: ReasonConditional}
	if !slices.Contains(report.Skipped, calculated) {
		t.Errorf("report lacks %+v", calculated)
	}

	directory := filepath.Join(t.TempDir(), "tavish")
	if err := writeLayer(directory, "actors/tavish.json", layer); err != nil {
		t.Fatalf("writeLayer: %v", err)
	}
	loaded, err := creature.Load(directory)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	damage := loaded.Strikes["alchemists-fire-greater"].Damage
	if len(damage) == 0 {
		t.Fatal("alchemists-fire-greater has no damage after Load")
	}
	last := damage[len(damage)-1]
	if last.Dice != "7" || last.Type != "fire" || last.Category != "splash" {
		t.Errorf("last damage part = %+v, want 7 fire splash (3 plus Intelligence +4)", last)
	}
}

func TestImportColdIron(t *testing.T) {
	layer, _ := importTestPC(t, "tavish")
	if traits := layer.Build.Weapons["nail-bomb-greater"].Traits; !slices.Contains(traits, "cold-iron") {
		t.Errorf("nail-bomb-greater traits = %v, want cold-iron", traits)
	}
}
