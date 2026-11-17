package creature

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const SchemaVersion = 1

type Kind string

const (
	KindPC  Kind = "pc"
	KindNPC Kind = "npc"
)

type Creature struct {
	ID           string             `toml:"-"`
	Layers       []string           `toml:"-"`
	Attributes   map[string]int     `toml:"-"`
	Schema       int                `toml:"schema"`
	Name         string             `toml:"name,omitempty"`
	Kind         Kind               `toml:"kind,omitempty"`
	Level        int                `toml:"level,omitzero"`
	Traits       []string           `toml:"traits,omitempty"`
	Perception   int                `toml:"perception,omitzero"`
	ClassDC      int                `toml:"class_dc,omitzero"`
	Speed        Speed              `toml:"speed,omitempty"`
	Defense      Defense            `toml:"defense,omitempty"`
	Spellcasting *Spellcasting      `toml:"spellcasting,omitempty"`
	Build        *Build             `toml:"build,omitempty"`
	Strikes      map[string]Strike  `toml:"strike,omitempty"`
	Abilities    map[string]Ability `toml:"ability,omitempty"`
}

type Speed struct {
	Land int `toml:"land,omitzero"`
	Fly  int `toml:"fly,omitzero"`
}

type Defense struct {
	HP        int                   `toml:"hp,omitzero"`
	AC        int                   `toml:"ac,omitzero"`
	Fortitude int                   `toml:"fortitude,omitzero"`
	Reflex    int                   `toml:"reflex,omitzero"`
	Will      int                   `toml:"will,omitzero"`
	Immune    []string              `toml:"immune,omitempty"`
	Resist    map[string]Adjustment `toml:"resist,omitempty"`
	Weakness  map[string]Adjustment `toml:"weakness,omitempty"`
}

type Adjustment struct {
	Value  int      `toml:"value"`
	Except []string `toml:"except,omitempty"`
}

type Spellcasting struct {
	Attack  int `toml:"attack,omitzero"`
	DC      int `toml:"dc,omitzero"`
	TopRank int `toml:"top_rank,omitzero"`
}

type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name,omitempty"`
	Bonus  int          `toml:"bonus,omitzero"`
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
}

type DamagePart struct {
	Dice string          `toml:"dice"`
	Type string          `toml:"type"`
	Roll dice.Expression `toml:"-"`
}

type Ability struct {
	ID      string `toml:"-"`
	Name    string `toml:"name,omitempty"`
	Actions int    `toml:"actions,omitzero"`
	Ignore  bool   `toml:"ignore,omitempty"`
	Steps   []Step `toml:"steps,omitempty"`
}

type Step struct {
	Strike string `toml:"strike"`
}

func (strike Strike) Agile() bool {
	return slices.Contains(strike.Traits, "agile")
}

func (creature *Creature) StrikeIDs() []string {
	return slices.Sorted(maps.Keys(creature.Strikes))
}

func (creature *Creature) AbilityIDs() []string {
	return slices.Sorted(maps.Keys(creature.Abilities))
}

var (
	damageTypes   = []string{"acid", "bleed", "bludgeoning", "cold", "electricity", "fire", "force", "mental", "piercing", "poison", "slashing", "sonic", "spirit", "vitality", "void"}
	physicalTypes = []string{"bludgeoning", "piercing", "slashing"}
)

func IsDamageType(name string) bool {
	return slices.Contains(damageTypes, name)
}

func IsPhysical(damageType string) bool {
	return slices.Contains(physicalTypes, damageType)
}
