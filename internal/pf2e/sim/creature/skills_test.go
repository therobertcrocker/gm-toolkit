package creature_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const npcHand = `schema = 1
name = "Brute"
kind = "npc"
level = 5
perception = 10

[defense]
hp = 60
ac = 20
`

func loadSkills(t *testing.T, hand string) map[string]int {
	t.Helper()
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild, "hand.toml": "schema = 1\n" + hand}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Skills) != len(creature.SkillNames) {
		t.Fatalf("skills = %v, want all %d", loaded.Skills, len(creature.SkillNames))
	}
	return loaded.Skills
}

func checkSkills(t *testing.T, got, want map[string]int) {
	t.Helper()
	for name, value := range want {
		if got[name] != value {
			t.Errorf("%s = %+d, want %+d", name, got[name], value)
		}
	}
}

func TestSkillsFromBuild(t *testing.T) {
	const ranks = `
[build.skill]
intimidation = "master"
athletics    = "trained"

[build.modifier.demon-mask]
selector = "intimidation"
type     = "item"
value    = 2

[build.modifier.lesser-mask]
selector = "intimidation"
type     = "item"
value    = 1
`
	checkSkills(t, loadSkills(t, ranks), map[string]int{
		"intimidation": 26, "athletics": 21, "arcana": 0, "medicine": 3, "deception": 4, "acrobatics": 0,
	})
	checkSkills(t, loadSkills(t, ranks+"\n[build.modifier.lucky]\nselector = \"skill-check\"\ntype = \"status\"\nvalue = 1\n"), map[string]int{
		"intimidation": 27, "athletics": 22, "arcana": 1, "medicine": 4, "deception": 5, "acrobatics": 1,
	})
}

func TestSkillsUseTheHigherProficiency(t *testing.T) {
	const hand = `
[build.skill]
intimidation = "master"
athletics    = "trained"

[build.modifier.untrained-improvisation]
selector = "skill-check"
type     = "proficiency"
value    = 14

[build.modifier.stealth-training]
selector = "stealth"
type     = "proficiency"
value    = 20
`
	checkSkills(t, loadSkills(t, hand), map[string]int{
		"arcana": 14, "medicine": 17, "deception": 18, "athletics": 21, "intimidation": 24, "stealth": 20,
	})
	checkSkills(t, loadSkills(t, "[build.modifier.low]\nselector = \"skill-check\"\ntype = \"proficiency\"\nvalue = 3\n\n[build.skill]\nathletics = \"trained\"\n"), map[string]int{
		"arcana": 3, "athletics": 21,
	})
}

func TestArmorCheckPenalty(t *testing.T) {
	skills := loadSkills(t, "[build.armor]\ncheck_penalty = -2\n")
	checkSkills(t, skills, map[string]int{
		"athletics": 3, "acrobatics": -2, "stealth": -2, "thievery": -2,
		"arcana": 0, "crafting": 0, "deception": 4, "diplomacy": 4, "intimidation": 4, "medicine": 3,
		"nature": 3, "occultism": 0, "performance": 4, "religion": 3, "society": 0, "survival": 3,
	})
}

func TestLoadRejectsBadSkills(t *testing.T) {
	cases := []struct {
		name     string
		layers   map[string]string
		wantText string
	}{
		{"unknown skill in the build", map[string]string{"import.toml": fighterBuild, "hand.toml": "schema = 1\n[build.skill]\npiloting = \"trained\"\n"}, "build.skill.piloting: not a skill"},
		{"bad rank word", map[string]string{"import.toml": fighterBuild, "hand.toml": "schema = 1\n[build.skill]\nstealth = \"grandmaster\"\n"}, `build.skill.stealth: rank "grandmaster"`},
		{"positive check penalty", map[string]string{"import.toml": fighterBuild, "hand.toml": "schema = 1\n[build.armor]\ncheck_penalty = 1\n"}, "build.armor: speed_penalty and check_penalty must not be positive"},
		{"proficiency type off a skill", map[string]string{"import.toml": fighterBuild, "hand.toml": "schema = 1\n[build.modifier.odd]\nselector = \"will\"\ntype = \"proficiency\"\nvalue = 14\n"}, "build.modifier.odd: type proficiency needs a skill or skill-check selector"},
		{"unknown skill on an npc", map[string]string{"hand.toml": npcHand + "\n[skill]\npiloting = 12\n"}, "skill.piloting: not a skill"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := creature.Load(writeLayers(t, testCase.layers))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestSkillIsAComputedKey(t *testing.T) {
	_, err := creature.Load(writeLayers(t, map[string]string{
		"import.toml": fighterBuild,
		"hand.toml":   "schema = 1\n\n[skill]\nathletics = 30\n",
	}))
	if !errors.Is(err, creature.ErrComputedKey) {
		t.Fatalf("err = %v, want ErrComputedKey", err)
	}
	if !strings.Contains(err.Error(), "hand.toml sets skill") {
		t.Errorf("err = %v, want it to name hand.toml and skill", err)
	}
}

func TestNPCSkillsAreTyped(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": npcHand + "\n[skill]\nathletics = 30\n"}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Skills) != 1 || loaded.Skills["athletics"] != 30 {
		t.Errorf("Skills = %v, want athletics 30 only", loaded.Skills)
	}
}
