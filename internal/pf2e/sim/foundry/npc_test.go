package foundry

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func hasSkip(report *Report, itemName, reason string) bool {
	return slices.ContainsFunc(report.Skipped, func(skip Skip) bool {
		return skip.Item == itemName && skip.Reason == reason
	})
}

func importTestNPC(t *testing.T, name string) (*creature.Creature, *Report) {
	t.Helper()
	source, err := readActor(filepath.Join("testdata", "actors", name+".json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	layer, report, err := importNPC(source)
	if err != nil {
		t.Fatalf("importNPC: %v", err)
	}
	return layer, report
}

func TestImportNPCLuna(t *testing.T) {
	layer, report := importTestNPC(t, "abyssal-luna")
	if layer.Name != "Abyssal Luna" || layer.Kind != creature.KindNPC || layer.Level != 16 || layer.Perception != 29 {
		t.Errorf("header = %q %s %d perception %d, want Abyssal Luna npc 16 perception 29", layer.Name, layer.Kind, layer.Level, layer.Perception)
	}
	if want := []string{"unique", "fey", "occult"}; !slices.Equal(layer.Traits, want) {
		t.Errorf("Traits = %v, want %v", layer.Traits, want)
	}
	if want := (creature.Speed{Land: 60, Fly: 180}); layer.Speed != want {
		t.Errorf("Speed = %+v, want %+v", layer.Speed, want)
	}
	defense := layer.Defense
	if defense.HP != 300 || defense.AC != 36 || defense.Fortitude != 27 || defense.Reflex != 28 || defense.Will != 29 {
		t.Errorf("defense = %+v, want HP 300 AC 36 saves 27/28/29", defense)
	}
	if want := []string{"confused", "doomed", "paralyzed", "sleep"}; !slices.Equal(defense.Immune, want) {
		t.Errorf("Immune = %v, want %v", defense.Immune, want)
	}
	slash := layer.Strikes["slash"]
	wantSlash := creature.Strike{
		Name:   "Slash",
		Bonus:  30,
		Traits: []string{"agile", "magical", "reach-10", "unarmed"},
		Damage: []creature.DamagePart{{Dice: "3d6+14", Type: "slashing"}, {Dice: "1d8", Type: "mental"}},
	}
	if !reflect.DeepEqual(slash, wantSlash) {
		t.Errorf("slash = %+v, want %+v", slash, wantSlash)
	}
	if ids := slices.Sorted(maps.Keys(layer.Strikes)); !slices.Equal(ids, []string{"slash", "stab"}) {
		t.Errorf("strikes = %v, want slash and stab", ids)
	}
	if want := (creature.Spellcasting{Attack: 31, DC: 39}); layer.Spellcasting == nil || *layer.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", layer.Spellcasting, want)
	}
	if len(layer.Abilities) != 0 {
		t.Errorf("Abilities = %v, want none (Decision 4)", layer.Abilities)
	}
	for _, want := range []struct{ item, reason string }{
		{"Destiny Spores", ReasonAbility},
		{"Challenge Fate", ReasonAbility},
		{"Slash", ReasonLaterRule},
		{"Ill Omen (At Will)", ReasonLaterRule},
		{"Abyssal Luna", ReasonConditional},
		{"Fortune-Telling Lore", ReasonNotCombat},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}

func TestImportNPCLilith(t *testing.T) {
	layer, report := importTestNPC(t, "avatar-of-lilith")
	if want := (creature.Speed{Land: 0, Fly: 60}); layer.Speed != want {
		t.Errorf("Speed = %+v, want %+v (no land speed on the sheet)", layer.Speed, want)
	}
	wantResist := map[string]creature.Adjustment{"all": {Value: 12, Except: []string{"holy", "cold-iron"}}}
	if !reflect.DeepEqual(layer.Defense.Resist, wantResist) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, wantResist)
	}
	if weakness := layer.Defense.Weakness; len(weakness) != 1 || weakness["holy"].Value != 15 {
		t.Errorf("Weakness = %+v, want holy 15: a trait-keyed weakness", weakness)
	}
	if hasSkip(report, "Avatar of Lilith", ReasonLaterRule) {
		t.Error("report skips an Avatar of Lilith defense, want the holy weakness imported")
	}
	if ids := slices.Sorted(maps.Keys(layer.Strikes)); !slices.Equal(ids, []string{"hand-of-chaos", "slash", "stinger"}) {
		t.Errorf("strikes = %v, want hand-of-chaos, slash, stinger", ids)
	}
	if want := (creature.Spellcasting{Attack: 32, DC: 38}); layer.Spellcasting == nil || *layer.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", layer.Spellcasting, want)
	}
	if !hasSkip(report, "Last Stand", ReasonAbility) {
		t.Error("report lacks Last Stand as an ability")
	}
}

func TestImportNPCSkipsDamageTheLoaderRejects(t *testing.T) {
	layer := &creature.Creature{}
	report := &Report{}
	taken := map[string]bool{}
	for _, test := range []struct{ name, rolls string }{
		{"Mixed", `{"a": {"damage": "2d8+10", "damageType": "slashing"}, "b": {"damage": "1d6", "damageType": "chaotic"}}`},
		{"Bad Dice", `{"a": {"damage": "2d", "damageType": "slashing"}}`},
		{"Persistent Only", `{"a": {"damage": "2d6", "damageType": "fire", "category": "persistent"}}`},
	} {
		current := &item{Name: test.name, Type: "melee", Raw: []byte(`{"damageRolls": ` + test.rolls + `}`)}
		if err := importNPCStrike(current, taken, layer, report); err != nil {
			t.Fatalf("%s: importNPCStrike: %v", test.name, err)
		}
	}
	if ids := slices.Sorted(maps.Keys(layer.Strikes)); !slices.Equal(ids, []string{"mixed"}) {
		t.Errorf("strikes = %v, want only mixed", ids)
	}
	if want := []creature.DamagePart{{Dice: "2d8+10", Type: "slashing"}}; !reflect.DeepEqual(layer.Strikes["mixed"].Damage, want) {
		t.Errorf("mixed damage = %+v, want %+v", layer.Strikes["mixed"].Damage, want)
	}
	for _, name := range []string{"Mixed", "Bad Dice", "Persistent Only"} {
		if !hasSkip(report, name, ReasonUnsupported) {
			t.Errorf("report lacks %s with reason %q", name, ReasonUnsupported)
		}
	}
}

func TestImportedNPCLoadsUnderItsHandLayer(t *testing.T) {
	layer, _ := importTestNPC(t, "abyssal-luna")
	directory := filepath.Join(t.TempDir(), "abyssal-luna")
	if err := writeLayer(directory, "actors/abyssal-luna.json", layer); err != nil {
		t.Fatalf("writeLayer: %v", err)
	}
	hand := "schema = 1\n\n[ability.abyssal-frenzy]\nname = \"Abyssal Frenzy\"\nactions = 2\nsteps = [{ strike = \"slash\" }, { strike = \"slash\" }, { strike = \"stab\" }]\n"
	if err := os.WriteFile(filepath.Join(directory, "hand.toml"), []byte(hand), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	loaded, err := creature.Load(directory)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 300 || loaded.Strikes["slash"].Bonus != 30 {
		t.Errorf("loaded HP %d, slash %+d; want 300 and +30", loaded.Defense.HP, loaded.Strikes["slash"].Bonus)
	}
	if steps := loaded.Abilities["abyssal-frenzy"].Steps; len(steps) != 3 {
		t.Errorf("abyssal-frenzy steps = %v, want 3", steps)
	}
	if want := []string{"import.toml", "hand.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
}
