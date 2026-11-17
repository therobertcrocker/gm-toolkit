package foundry

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestReadActorKeepsDamageRollOrder(t *testing.T) {
	luna, err := readActor(filepath.Join("testdata", "actors", "abyssal-luna.json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	if luna.Name != "Abyssal Luna" || luna.Type != "npc" {
		t.Errorf("actor = %q %q, want Abyssal Luna npc", luna.Name, luna.Type)
	}
	index := slices.IndexFunc(luna.Items, func(current item) bool { return current.Name == "Slash" })
	if index < 0 {
		t.Fatal("no Slash item")
	}
	var system struct {
		DamageRolls damageRolls `json:"damageRolls"`
	}
	if err := luna.Items[index].decode(&system); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := damageRolls{{Damage: "3d6+14", DamageType: "slashing"}, {Damage: "1d8", DamageType: "mental"}}
	if !slices.Equal(system.DamageRolls, want) {
		t.Errorf("damage rolls = %+v, want %+v", system.DamageRolls, want)
	}
	if traits := luna.Items[index].System.Traits.Value; !slices.Contains(traits, "agile") {
		t.Errorf("Slash traits = %v, want agile among them", traits)
	}
}

func TestReadActorRejectsAnItemExport(t *testing.T) {
	_, err := readActor(filepath.Join("testdata", "items", "swear-fidelity.json"))
	if !errors.Is(err, ErrNotAnActor) {
		t.Errorf("err = %v, want ErrNotAnActor", err)
	}
}

func TestReadItem(t *testing.T) {
	loose, err := readItem(filepath.Join("testdata", "items", "swear-fidelity.json"))
	if err != nil {
		t.Fatalf("readItem: %v", err)
	}
	if loose.Name != "Swear Fidelity" || loose.Type != "action" {
		t.Errorf("item = %q %q, want Swear Fidelity action", loose.Name, loose.Type)
	}
}

func TestSlugify(t *testing.T) {
	for _, test := range []struct{ name, want string }{
		{"Protector's Fury", "protectors-fury"},
		{"Alchemist's Fire (Greater)", "alchemists-fire-greater"},
		{"Artisan's Mark™", "artisans-mark"},
		{"Bands of Force (Greater)", "bands-of-force-greater"},
		{"  ", ""},
	} {
		if got := slugify(test.name); got != test.want {
			t.Errorf("slugify(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestUniqueID(t *testing.T) {
	taken := map[string]bool{}
	for _, want := range []struct{ name, id string }{
		{"Slash", "slash"},
		{"Slash", "slash-2"},
		{"Slash", "slash-3"},
		{"™", "item"},
	} {
		if got := uniqueID(want.name, taken); got != want.id {
			t.Errorf("uniqueID(%q) = %q, want %q", want.name, got, want.id)
		}
	}
}
