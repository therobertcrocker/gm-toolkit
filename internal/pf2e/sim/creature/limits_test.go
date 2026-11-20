package creature_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const limitedAbilities = `
[pool]
potions = 3

[ability.drink]
name    = "Drink"
actions = 2
limit   = { per_fight = 1, recharge = "1d4", pool = "potions" }
steps   = [{ strike = "fist" }]

[ability.flurry]
name        = "Flurry"
actions     = 1
max_actions = 3
steps       = [{ strike = "fist" }]

[ability.spark]
name  = "Spark"
free  = true
limit = { per_round = 1 }
steps = [{ strike = "fist" }]
`

func TestLoadReadsLimits(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + limitedAbilities}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Pools["potions"] != 3 {
		t.Errorf("pool potions = %d, want 3", loaded.Pools["potions"])
	}

	drink := loaded.Abilities["drink"]
	if drink.Limit == nil {
		t.Fatal("drink limit = nil, want a limit")
	}
	if limit := drink.Limit; limit.PerFight != 1 || limit.Recharge != "1d4" || limit.Pool != "potions" {
		t.Errorf("drink limit = %+v, want per_fight 1, recharge 1d4, pool potions", limit)
	}
	if want := (dice.Expression{Count: 1, Sides: 4}); drink.Limit.RechargeRoll != want {
		t.Errorf("drink RechargeRoll = %+v, want %+v", drink.Limit.RechargeRoll, want)
	}
	for _, testCase := range []struct {
		id        string
		low, high int
	}{
		{id: "drink", low: 2, high: 2},
		{id: "flurry", low: 1, high: 3},
		{id: "spark", low: 0, high: 0},
	} {
		low, high := loaded.Abilities[testCase.id].CostRange()
		if low != testCase.low || high != testCase.high {
			t.Errorf("%s CostRange() = %d, %d, want %d, %d", testCase.id, low, high, testCase.low, testCase.high)
		}
	}
}

func loadAbility(t *testing.T, keys, extra string) error {
	t.Helper()
	ability := fmt.Sprintf("\n[ability.bad]\nname = \"Bad\"\n%s\nsteps = [{ strike = \"fist\" }]\n\n%s", keys, extra)
	_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + ability}))
	return err
}

func TestLoadRejectsBadCosts(t *testing.T) {
	cases := []struct {
		name     string
		keys     string
		wantText string
	}{
		{name: "no actions and not free", keys: "", wantText: "ability.bad: actions 0: want 1 to 3"},
		{name: "free with actions", keys: "free = true\nactions = 1\nlimit = { per_fight = 1 }", wantText: "a free ability takes no actions"},
		{name: "free with no limit", keys: "free = true", wantText: "a free ability needs a limit"},
		{name: "max_actions equal to actions", keys: "actions = 1\nmax_actions = 1", wantText: "max_actions 1: want more than actions and at most 3"},
		{name: "max_actions above 3", keys: "actions = 1\nmax_actions = 4", wantText: "max_actions 4: want more than actions and at most 3"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := loadAbility(t, testCase.keys, "")
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadRejectsBadLimits(t *testing.T) {
	cases := []struct {
		name     string
		limit    string
		extra    string
		wantText string
	}{
		{name: "empty limit", limit: "{}", wantText: "ability.bad.limit: sets nothing"},
		{name: "negative per_fight", limit: "{ per_fight = -1 }", wantText: "per_fight and per_round must not be negative"},
		{name: "recharge that does not parse", limit: `{ recharge = "soon" }`, wantText: "ability.bad.limit.recharge:"},
		{name: "unknown pool", limit: `{ pool = "missing" }`, wantText: `ability.bad.limit: unknown pool "missing"`},
		{name: "negative pool", limit: `{ pool = "potions" }`, extra: "[pool]\npotions = -1\n", wantText: "pool.potions: must not be negative"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := loadAbility(t, "actions = 1\nlimit = "+testCase.limit, testCase.extra)
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadRejectsBadTargetsAndSustain(t *testing.T) {
	cases := []struct {
		name     string
		keys     string
		wantText string
	}{
		{name: "unknown targets", keys: "actions = 1\ntargets = \"party\"", wantText: `ability.bad: targets "party": want self, ally, others, or leave it unset`},
		{name: "others without an area", keys: "actions = 1\ntargets = \"others\"", wantText: "ability.bad: targets others needs an area"},
		{name: "self with an area", keys: "actions = 1\ntargets = \"self\"\narea = { shape = \"burst\", size = 10 }", wantText: "ability.bad: targets self takes no area"},
		{name: "sustain on a triggered ability", keys: "trigger = { on = \"damaged\", cost = \"reaction\" }\nsustain = [{ strike = \"fist\" }]", wantText: "ability.bad: a triggered ability cannot be sustained"},
		{name: "a bad step inside sustain", keys: "actions = 1\nsustain = [{ strike = \"kick\" }]", wantText: `ability.bad.sustain[0]: unknown strike "kick"`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := loadAbility(t, testCase.keys, "")
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadRejectsBadTags(t *testing.T) {
	cases := []struct {
		name     string
		ability  string
		wantText string
	}{
		{name: "an unknown tag", ability: "actions = 1\ntags = [\"betrayal\"]\nsteps = [{ strike = \"fist\" }]", wantText: `ability.bad: tag "betrayal": want loyalty, tests-loyalty`},
		{name: "an untagged ability with no steps", ability: "actions = 1", wantText: "ability.bad: no steps"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + "\n[ability.bad]\nname = \"Bad\"\n" + testCase.ability + "\n"}))
			if !errors.Is(err, creature.ErrInvalid) {
				t.Fatalf("Load error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("Load error = %v, want text %q", err, testCase.wantText)
			}
		})
	}
}

func TestLoadReadsATaggedAbilityWithNoSteps(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + "\n[ability.oath]\nname = \"Oath\"\nactions = 1\ntags = [\"loyalty\"]\ntargets = \"self\"\n"}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if oath := loaded.Abilities["oath"]; !slices.Equal(oath.Tags, []string{creature.TagLoyalty}) || len(oath.Steps) != 0 {
		t.Errorf("oath = %+v, want tag loyalty and no steps", oath)
	}
}

func TestLoadReadsTargetsAndSustain(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"hand.toml": minimalHand + "\n[ability.song]\nname = \"Song\"\nactions = 2\ntargets = \"ally\"\nsustain = [{ heal = \"1d6\" }]\nsteps = [{ heal = \"2d6\" }]\n"}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	song := loaded.Abilities["song"]
	if song.Targets != creature.TargetsAlly || len(song.Sustain) != 1 || song.Sustain[0].HealRoll.Count != 1 || song.Sustain[0].HealRoll.Sides != 6 {
		t.Errorf("song = %+v, want targets ally and a parsed 1d6 sustain heal", song)
	}
}

func TestIgnoredAbilitySkipsLimitChecks(t *testing.T) {
	if err := loadAbility(t, "ignore = true\nfree = true\nactions = 2\nlimit = { pool = \"missing\" }", ""); err != nil {
		t.Errorf("Load error = %v, want an ignored ability with a bad cost and limit to load", err)
	}
}
