package creature_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const minimalHand = `schema = 1
name = "Minimal"
kind = "pc"
level = 1
perception = 3

[defense]
hp = 10
ac = 15

[strike.fist]
name = "Fist"
bonus = 5
damage = [{ dice = "1d4", type = "bludgeoning" }]
`

func writeLayers(t *testing.T, layers map[string]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "subject")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	for name, content := range layers {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	return dir
}

func TestLoadMergesImportAndHand(t *testing.T) {
	loaded, err := creature.Load("testdata/layered")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != "layered" {
		t.Errorf("ID = %q, want layered", loaded.ID)
	}
	if want := []string{"import.toml", "hand.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
	claw := loaded.Strikes["claw"]
	if claw.Bonus != 14 {
		t.Errorf("claw bonus = %d, want 14 from the hand layer", claw.Bonus)
	}
	if claw.Name != "Claw" || claw.ID != "claw" {
		t.Errorf("claw name, id = %q, %q, want Claw, claw kept from the import layer", claw.Name, claw.ID)
	}
	if want := (dice.Expression{Count: 2, Sides: 6, Modifier: 4}); claw.Damage[0].Roll != want {
		t.Errorf("claw damage = %+v, want %+v", claw.Damage[0].Roll, want)
	}
	if !claw.Agile() {
		t.Error("claw is not agile")
	}
	if _, found := loaded.Strikes["bite"]; !found {
		t.Error("bite from the import layer is missing")
	}
	if loaded.Defense.HP != 60 {
		t.Errorf("HP = %d, want 60", loaded.Defense.HP)
	}
	if got := loaded.Defense.Resist["fire"].Value; got != 5 {
		t.Errorf("fire resistance = %d, want 5", got)
	}
	if rend := loaded.Abilities["rend"]; rend.ID != "rend" || rend.Actions != 2 || len(rend.Steps) != 2 {
		t.Errorf("rend = %+v, want a 2-action ability with 2 steps", rend)
	}
	if howl := loaded.Abilities["howl"]; !howl.Ignore || howl.Name != "Howl" {
		t.Errorf("howl = %+v, want the import layer's Howl marked ignore", howl)
	}
	if want := []string{"howl", "rend"}; !slices.Equal(loaded.AbilityIDs(), want) {
		t.Errorf("AbilityIDs = %v, want %v", loaded.AbilityIDs(), want)
	}
}

func TestLoadAppliesVariantsLast(t *testing.T) {
	loaded, err := creature.Load("testdata/layered", "elite")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 80 || loaded.Defense.AC != 22 {
		t.Errorf("HP, AC = %d, %d, want 80, 22", loaded.Defense.HP, loaded.Defense.AC)
	}
	if want := []string{"elite"}; !slices.Equal(loaded.Traits, want) {
		t.Errorf("Traits = %v, want %v (a list is replaced, not merged)", loaded.Traits, want)
	}
	if want := []string{"import.toml", "hand.toml", "elite.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
}

func TestLoadHandOnly(t *testing.T) {
	dir := writeLayers(t, map[string]string{"hand.toml": minimalHand})
	loaded, err := creature.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != "subject" || !slices.Equal(loaded.Layers, []string{"hand.toml"}) {
		t.Errorf("ID, Layers = %q, %v, want subject, [hand.toml]", loaded.ID, loaded.Layers)
	}
}

func TestLoadErrors(t *testing.T) {
	const flurry = "\n[ability.flurry]\nname = \"Flurry\"\nactions = 2\nsteps = [{ strike = \"kick\" }]\n"
	const slowFlurry = "\n[ability.flurry]\nname = \"Flurry\"\nactions = 4\nsteps = [{ strike = \"fist\" }]\n"
	cases := []struct {
		name     string
		layers   map[string]string
		variants []string
		wantErr  error
		wantText string
	}{
		{name: "no layers", layers: map[string]string{}, wantErr: creature.ErrNoLayers},
		{name: "unknown key", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "level = 1", "level = 1\nlevle = 2", 1)}, wantText: "unknown keys: levle"},
		{name: "wrong schema", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "schema = 1", "schema = 2", 1)}, wantText: "schema 2, want 1"},
		{name: "missing variant", layers: map[string]string{"hand.toml": minimalHand}, variants: []string{"elite"}, wantText: "elite.toml"},
		{name: "capital key", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "hp = 10", "HP = 10", 1)}, wantText: "keys must be lower case: defense.HP"},
		{name: "variant outside the folder", layers: map[string]string{"hand.toml": minimalHand}, variants: []string{"../subject/hand"}, wantErr: creature.ErrBadVariant},
		{name: "variant names a base layer", layers: map[string]string{"hand.toml": minimalHand}, variants: []string{"hand"}, wantErr: creature.ErrBadVariant},
		{name: "empty variant", layers: map[string]string{"hand.toml": minimalHand}, variants: []string{""}, wantErr: creature.ErrBadVariant},
		{name: "no armor class", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "ac = 15\n", "", 1)}, wantErr: creature.ErrInvalid, wantText: "defense.ac"},
		{name: "negative resistance", layers: map[string]string{"hand.toml": minimalHand + "\n[defense.resist.fire]\nvalue = -5\n"}, wantErr: creature.ErrInvalid, wantText: "defense.resist.fire: value must be at least 1"},
		{name: "weakness with no value", layers: map[string]string{"hand.toml": minimalHand + "\n[defense.weakness.fire]\nexcept = [\"holy\"]\n"}, wantErr: creature.ErrInvalid, wantText: "defense.weakness.fire: value must be at least 1"},
		{name: "bad kind", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, `kind = "pc"`, `kind = "boss"`, 1)}, wantErr: creature.ErrInvalid, wantText: `kind "boss"`},
		{name: "bad dice", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, `"1d4"`, `"1d"`, 1)}, wantErr: creature.ErrInvalid, wantText: "strike.fist.damage[0]"},
		{name: "bad damage type", layers: map[string]string{"hand.toml": strings.Replace(minimalHand, "bludgeoning", "blunt", 1)}, wantErr: creature.ErrInvalid, wantText: `unknown damage type "blunt"`},
		{name: "unknown step strike", layers: map[string]string{"hand.toml": minimalHand + flurry}, wantErr: creature.ErrInvalid, wantText: `unknown strike "kick"`},
		{name: "too many actions", layers: map[string]string{"hand.toml": minimalHand + slowFlurry}, wantErr: creature.ErrInvalid, wantText: "actions 4"},
		{name: "bad resist key", layers: map[string]string{"hand.toml": minimalHand + "\n[defense.resist.blunt]\nvalue = 5\n"}, wantErr: creature.ErrInvalid, wantText: "defense.resist.blunt"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := writeLayers(t, testCase.layers)
			_, err := creature.Load(dir, testCase.variants...)
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				t.Errorf("err = %v, want %v", err, testCase.wantErr)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("err = %v, want it to mention %q", err, testCase.wantText)
			}
		})
	}
}
