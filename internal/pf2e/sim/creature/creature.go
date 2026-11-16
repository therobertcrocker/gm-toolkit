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
	ID         string             `toml:"-"`
	Layers     []string           `toml:"-"`
	Schema     int                `toml:"schema"`
	Name       string             `toml:"name"`
	Kind       Kind               `toml:"kind"`
	Level      int                `toml:"level"`
	Traits     []string           `toml:"traits"`
	Perception int                `toml:"perception"`
	Speed      Speed              `toml:"speed"`
	Defense    Defense            `toml:"defense"`
	Strikes    map[string]Strike  `toml:"strike"`
	Abilities  map[string]Ability `toml:"ability"`
}

type Speed struct {
	Land int `toml:"land"`
	Fly  int `toml:"fly"`
}

type Defense struct {
	HP        int                   `toml:"hp"`
	AC        int                   `toml:"ac"`
	Fortitude int                   `toml:"fortitude"`
	Reflex    int                   `toml:"reflex"`
	Will      int                   `toml:"will"`
	Immune    []string              `toml:"immune"`
	Resist    map[string]Adjustment `toml:"resist"`
	Weakness  map[string]Adjustment `toml:"weakness"`
}

type Adjustment struct {
	Value  int      `toml:"value"`
	Except []string `toml:"except"`
}

type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name"`
	Bonus  int          `toml:"bonus"`
	Traits []string     `toml:"traits"`
	Damage []DamagePart `toml:"damage"`
}

type DamagePart struct {
	Dice string          `toml:"dice"`
	Type string          `toml:"type"`
	Roll dice.Expression `toml:"-"`
}

type Ability struct {
	ID      string `toml:"-"`
	Name    string `toml:"name"`
	Actions int    `toml:"actions"`
	Ignore  bool   `toml:"ignore"`
	Steps   []Step `toml:"steps"`
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
