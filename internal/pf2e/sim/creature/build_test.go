package creature_test

import (
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const fighterBuild = `schema = 1
name  = "Kas"
kind  = "pc"
level = 14

[build]
ancestry_hp = 6
class_hp    = 10
key         = "str"
speed       = 25
weapon_specialization = "normal"

[build.boosts]
ancestry   = ["str", "cha"]
background = ["con", "str"]
class      = ["str"]
level-1    = ["cha", "con", "str", "wis"]
level-5    = ["str", "con", "cha", "wis"]
level-10   = ["str", "con", "cha", "wis"]

[build.proficiency]
perception = "master"
fortitude  = "master"
reflex     = "expert"
will       = "expert"
class-dc   = "expert"
heavy      = "expert"
simple     = "master"
martial    = "master"
unarmed    = "master"
weapon-group-hammer = "legendary"

[build.armor]
name     = "Full Plate"
category = "heavy"
ac_bonus = 6
dex_cap  = 0
potency  = 2
resilient = 2

[build.weapon.protectors-fury]
name     = "Protector's Fury"
category = "martial"
group    = "hammer"
base     = "maul"
dice     = 1
die      = 12
type     = "bludgeoning"
potency  = 2
striking = 2
traits   = ["trip", "versatile-p"]

[build.modifier.toughness]
selector  = "hp"
per_level = 1

[build.modifier.mountains-stoutness]
selector  = "hp"
per_level = 1

[build.modifier.fleet]
selector = "land-speed"
value    = 5

[build.modifier.amulet-of-kinships-strength]
selector = "will"
type     = "item"
value    = 2

[defense.resist.slashing]
value = 4
`

const scoutBuild = `schema = 1
name  = "Scout"
kind  = "pc"
level = 5

[build]
ancestry_hp = 8
class_hp    = 8
key         = "dex"
speed       = 25
weapon_specialization = "normal"

[build.boosts]
ancestry   = ["dex"]
background = ["dex", "str"]
class      = ["dex"]
level-1    = ["dex", "str", "con", "wis"]
level-5    = ["dex", "str", "con", "wis"]

[build.proficiency]
unarmored = "trained"
simple    = "expert"
martial   = "trained"
weapon-base-alchemical-bomb = "expert"

[build.weapon.needle]
name     = "Needle"
category = "simple"
dice     = 1
die      = 6
type     = "piercing"
traits   = ["finesse"]

[build.weapon.longbow]
name     = "Longbow"
category = "martial"
dice     = 1
die      = 8
type     = "piercing"
range    = 100

[build.weapon.sling]
name     = "Sling"
category = "simple"
dice     = 1
die      = 6
type     = "bludgeoning"
range    = 50
traits   = ["propulsive"]

[build.weapon.javelin]
name     = "Javelin"
category = "simple"
dice     = 1
die      = 6
type     = "piercing"
range    = 30
traits   = ["thrown"]

[build.weapon.fire-bomb]
name       = "Fire Bomb"
category   = "martial"
base       = "alchemical-bomb"
dice       = 2
die        = 8
type       = "fire"
item_bonus = 2
range      = 20
traits     = ["alchemical", "bomb"]

[build.weapon.weak-bomb]
name     = "Weak Bomb"
category = "martial"
base     = "alchemical-bomb"
dice     = 2
die      = 8
type     = "fire"
range    = 20
traits   = ["alchemical", "bomb"]

[build.modifier.goggles]
selector      = "attack"
type          = "item"
value         = 1
weapon_traits = ["bomb"]

[build.modifier.bracers]
selector = "ac"
type     = "item"
value    = 1

[build.modifier.bands]
selector = "ac"
type     = "item"
value    = 2

[build.modifier.cover]
selector = "ac"
type     = "circumstance"
value    = 1

[build.modifier.frightened]
selector = "ac"
type     = "status"
value    = -1

[build.modifier.talisman]
selector = "ac"
value    = 1
`

func TestBuildComputesFighterStats(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantAttributes := map[string]int{"str": 5, "dex": 0, "con": 4, "int": 0, "wis": 3, "cha": 4}
	if !maps.Equal(loaded.Attributes, wantAttributes) {
		t.Errorf("Attributes = %v, want %v", loaded.Attributes, wantAttributes)
	}
	for _, check := range []struct {
		name      string
		got, want int
	}{
		{"hp", loaded.Defense.HP, 230},
		{"ac", loaded.Defense.AC, 36},
		{"fortitude", loaded.Defense.Fortitude, 26},
		{"reflex", loaded.Defense.Reflex, 20},
		{"will", loaded.Defense.Will, 23},
		{"perception", loaded.Perception, 23},
		{"land speed", loaded.Speed.Land, 30},
		{"class dc", loaded.ClassDC, 33},
		{"slashing resistance", loaded.Defense.Resist["slashing"].Value, 4},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	fury, found := loaded.Strikes["protectors-fury"]
	if !found {
		t.Fatalf("no protectors-fury Strike; have %v", loaded.StrikeIDs())
	}
	if fury.ID != "protectors-fury" || fury.Bonus != 29 {
		t.Errorf("Strike = %q %+d, want protectors-fury +29", fury.ID, fury.Bonus)
	}
	if got := fury.Damage[0].Roll.String(); got != "3d12+9" || fury.Damage[0].Type != "bludgeoning" {
		t.Errorf("damage = %s %s, want 3d12+9 bludgeoning", got, fury.Damage[0].Type)
	}
	if loaded.Spellcasting != nil {
		t.Errorf("Spellcasting = %+v, want none without a spellcasting choice", *loaded.Spellcasting)
	}
}

func TestBuildLevelUpLayerFollowsThrough(t *testing.T) {
	directory := writeLayers(t, map[string]string{
		"import.toml":   fighterBuild,
		"level-15.toml": "schema = 1\nlevel = 15\n",
	})
	loaded, err := creature.Load(directory, "level-15")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 246 {
		t.Errorf("hp = %d, want 246 (Toughness and Mountain's Stoutness grow per level)", loaded.Defense.HP)
	}
	if bonus := loaded.Strikes["protectors-fury"].Bonus; bonus != 30 {
		t.Errorf("attack = %+d, want +30", bonus)
	}
}

func TestBuildGroupRankSkipsAdvancedWeapons(t *testing.T) {
	directory := writeLayers(t, map[string]string{
		"import.toml": fighterBuild,
		"advanced-hammer.toml": `schema = 1

[build.proficiency]
advanced = "master"

[build.weapon.advanced-hammer]
name     = "Advanced Hammer"
category = "advanced"
group    = "hammer"
dice     = 1
die      = 12
type     = "bludgeoning"
potency  = 2
striking = 2
`,
	})
	loaded, err := creature.Load(directory, "advanced-hammer")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	hammer := loaded.Strikes["advanced-hammer"]
	if hammer.Bonus != 27 || hammer.Damage[0].Roll.String() != "3d12+8" {
		t.Errorf("advanced-hammer = %+d %s, want +27 3d12+8 (master in advanced, not the legendary hammer group)", hammer.Bonus, hammer.Damage[0].Roll)
	}
	if fury := loaded.Strikes["protectors-fury"]; fury.Bonus != 29 {
		t.Errorf("protectors-fury = %+d, want +29 (the martial hammer keeps the group rank)", fury.Bonus)
	}
}

func TestBuildWeaponAttributesAndBonusStacking(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": scoutBuild}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.AC != 24 {
		t.Errorf("ac = %d, want 24 (item 2 of 1 and 2, circumstance +1, status -1, untyped +1)", loaded.Defense.AC)
	}
	if loaded.Defense.HP != 58 {
		t.Errorf("hp = %d, want 58", loaded.Defense.HP)
	}
	for _, want := range []struct {
		id     string
		bonus  int
		damage string
	}{
		{"needle", 13, "1d6+5"},
		{"longbow", 11, "1d8"},
		{"sling", 13, "1d6+3"},
		{"javelin", 13, "1d6+5"},
		{"fire-bomb", 15, "2d8+2"},
		{"weak-bomb", 14, "2d8+2"},
	} {
		strike := loaded.Strikes[want.id]
		if strike.Bonus != want.bonus || strike.Damage[0].Roll.String() != want.damage {
			t.Errorf("%s = %+d %s, want %+d %s", want.id, strike.Bonus, strike.Damage[0].Roll, want.bonus, want.damage)
		}
	}
}

func TestBuildAttributesTrackPartialBoosts(t *testing.T) {
	build := &creature.Build{
		Flaws: []string{"con"},
		Boosts: map[string][]string{
			"ancestry":   {"dex", "int", "cha"},
			"background": {"int", "cha"},
			"class":      {"cha"},
			"level-1":    {"int", "cha", "dex", "con"},
			"level-5":    {"cha", "int", "dex", "wis"},
			"level-10":   {"con", "dex", "int", "cha"},
		},
	}
	got, err := build.Attributes()
	if err != nil {
		t.Fatalf("Attributes: %v", err)
	}
	want := map[string]int{"str": 0, "dex": 4, "con": 1, "int": 4, "wis": 1, "cha": 5}
	if !maps.Equal(got, want) {
		t.Errorf("Attributes = %v, want %v", got, want)
	}

	build.Boosts["level-five"] = []string{"str"}
	if _, err := build.Attributes(); err == nil || !strings.Contains(err.Error(), "build.boosts.level-five") {
		t.Errorf("err = %v, want one naming build.boosts.level-five", err)
	}
}

func TestBuildRejectsComputedKeys(t *testing.T) {
	_, err := creature.Load(writeLayers(t, map[string]string{
		"import.toml": fighterBuild,
		"hand.toml":   "schema = 1\n\n[defense]\nac = 40\n",
	}))
	if !errors.Is(err, creature.ErrComputedKey) {
		t.Fatalf("err = %v, want ErrComputedKey", err)
	}
	if !strings.Contains(err.Error(), "hand.toml sets defense.ac") {
		t.Errorf("err = %v, want it to name hand.toml and defense.ac", err)
	}
}

func TestBuildRejectsBadChoices(t *testing.T) {
	hand := "schema = 1\n\n[build.proficiency]\nperception = \"grandmaster\"\n\n[build.modifier.lucky]\nselector = \"initiative\"\nvalue = 1\n"
	_, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild, "hand.toml": hand}))
	if !errors.Is(err, creature.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	for _, want := range []string{`build.proficiency.perception: rank "grandmaster"`, `build.modifier.lucky: selector "initiative"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to contain %q", err, want)
		}
	}
}
