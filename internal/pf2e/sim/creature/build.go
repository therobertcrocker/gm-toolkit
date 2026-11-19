package creature

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const (
	WeaponGroupPrefix = "weapon-group-"
	WeaponBasePrefix  = "weapon-base-"
)

var ErrComputedKey = errors.New("key is computed from the build")

var (
	AttributeNames = []string{"str", "dex", "con", "int", "wis", "cha"}
	RankWords      = []string{"untrained", "trained", "expert", "master", "legendary"}

	fixedProficiencies   = []string{"perception", "fortitude", "reflex", "will", "class-dc", "spellcasting", "unarmored", "light", "medium", "heavy", "simple", "martial", "advanced", "unarmed"}
	armorCategories      = []string{"unarmored", "light", "medium", "heavy"}
	weaponCategories     = []string{"simple", "martial", "advanced", "unarmed"}
	modifierSelectors    = []string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage", "splash"}
	weaponSelectors      = []string{"attack", "damage", "splash"}
	modifierTypes        = []string{"", "item", "status", "circumstance"}
	fixedBoostSources    = []string{"ancestry", "background", "class"}
	saveAttributes       = map[string]string{"fortitude": "con", "reflex": "dex", "will": "wis"}
	specializationDamage = map[string][]int{"normal": {0, 0, 2, 3, 4}, "greater": {0, 0, 4, 6, 8}}
	computedKeys         = [][]string{
		{"perception"}, {"class_dc"}, {"spellcasting"}, {"strike"}, {"speed", "land"},
		{"defense", "hp"}, {"defense", "ac"}, {"defense", "fortitude"}, {"defense", "reflex"}, {"defense", "will"},
	}
)

type Build struct {
	AncestryHP             int                 `toml:"ancestry_hp,omitzero"`
	ClassHP                int                 `toml:"class_hp,omitzero"`
	Key                    string              `toml:"key,omitempty"`
	Speed                  int                 `toml:"speed,omitzero"`
	Flaws                  []string            `toml:"flaws,omitempty"`
	WeaponSpecialization   string              `toml:"weapon_specialization,omitempty"`
	CriticalSpecialization string              `toml:"critical_specialization,omitempty"`
	Boosts                 map[string][]string `toml:"boosts,omitempty"`
	Proficiency            map[string]string   `toml:"proficiency,omitempty"`
	Armor                  *Armor              `toml:"armor,omitempty"`
	Weapons                map[string]Weapon   `toml:"weapon,omitempty"`
	Modifiers              map[string]Modifier `toml:"modifier,omitempty"`
	Spellcasting           *Casting            `toml:"spellcasting,omitempty"`
}

type Armor struct {
	Name         string `toml:"name,omitempty"`
	Category     string `toml:"category"`
	ACBonus      int    `toml:"ac_bonus,omitzero"`
	DexCap       *int   `toml:"dex_cap,omitempty"`
	Potency      int    `toml:"potency,omitzero"`
	Resilient    int    `toml:"resilient,omitzero"`
	SpeedPenalty int    `toml:"speed_penalty,omitzero"`
}

type Weapon struct {
	Name       string      `toml:"name"`
	Category   string      `toml:"category"`
	Group      string      `toml:"group,omitempty"`
	Base       string      `toml:"base,omitempty"`
	Dice       int         `toml:"dice"`
	Die        int         `toml:"die"`
	Type       string      `toml:"type"`
	Potency    int         `toml:"potency,omitzero"`
	Striking   int         `toml:"striking,omitzero"`
	ItemBonus  int         `toml:"item_bonus,omitzero"`
	Range      int         `toml:"range,omitzero"`
	Traits     []string    `toml:"traits,omitempty"`
	Persistent *DamagePart `toml:"persistent,omitempty"`
	Splash     int         `toml:"splash,omitzero"`
	Outcomes
}

// deliberate: a group's critical specialization is a rider step in a table, not code per group, because the effect is data; one row exists because only Kas's hammer earns it in the scope.
var criticalSpecializations = map[string]Step{
	"hammer": {Save: "fortitude", DCFrom: "class", Outcomes: Outcomes{
		Failure:         []Step{{Condition: "prone"}},
		CriticalFailure: []Step{{Condition: "prone"}},
	}},
}

type Modifier struct {
	Selector     string   `toml:"selector"`
	Type         string   `toml:"type,omitempty"`
	Value        int      `toml:"value,omitzero"`
	PerLevel     int      `toml:"per_level,omitzero"`
	Attribute    string   `toml:"attribute,omitempty"`
	WeaponTraits []string `toml:"weapon_traits,omitempty"`
}

type Casting struct {
	Attribute string `toml:"attribute"`
	TopRank   int    `toml:"top_rank,omitzero"`
}

func IsProficiencyKey(key string) bool {
	if slices.Contains(fixedProficiencies, key) {
		return true
	}
	for _, prefix := range []string{WeaponGroupPrefix, WeaponBasePrefix} {
		if rest, found := strings.CutPrefix(key, prefix); found && rest != "" {
			return true
		}
	}
	return false
}

func (build *Build) Attributes() (map[string]int, error) {
	sources, err := boostSources(build.Boosts)
	if err != nil {
		return nil, err
	}
	modifiers := map[string]int{}
	for _, name := range AttributeNames {
		modifiers[name] = 0
	}
	for _, flaw := range build.Flaws {
		if _, known := modifiers[flaw]; !known {
			return nil, fmt.Errorf("build.flaws: %q is not an attribute", flaw)
		}
		modifiers[flaw]--
	}
	partial := map[string]bool{}
	for _, source := range sources {
		for _, attribute := range build.Boosts[source] {
			if _, known := modifiers[attribute]; !known {
				return nil, fmt.Errorf("build.boosts.%s: %q is not an attribute", source, attribute)
			}
			switch {
			case modifiers[attribute] < 4:
				modifiers[attribute]++
			case partial[attribute]:
				modifiers[attribute]++
				partial[attribute] = false
			default:
				partial[attribute] = true
			}
		}
	}
	return modifiers, nil
}

func boostSources(boosts map[string][]string) ([]string, error) {
	type levelSource struct {
		level  int
		source string
	}
	var ordered []string
	for _, source := range fixedBoostSources {
		if _, found := boosts[source]; found {
			ordered = append(ordered, source)
		}
	}
	var levels []levelSource
	for _, source := range slices.Sorted(maps.Keys(boosts)) {
		if slices.Contains(fixedBoostSources, source) {
			continue
		}
		levelText, isLevel := strings.CutPrefix(source, "level-")
		level, err := strconv.Atoi(levelText)
		if !isLevel || err != nil || level < 1 {
			return nil, fmt.Errorf("build.boosts.%s: want ancestry, background, class or level-<n>", source)
		}
		levels = append(levels, levelSource{level: level, source: source})
	}
	slices.SortStableFunc(levels, func(first, second levelSource) int { return first.level - second.level })
	for _, entry := range levels {
		ordered = append(ordered, entry.source)
	}
	return ordered, nil
}

func checkComputedKeys(names []string, layers []map[string]any) error {
	hasBuild := slices.ContainsFunc(layers, func(layer map[string]any) bool {
		_, found := layer["build"]
		return found
	})
	if !hasBuild {
		return nil
	}
	var problems []string
	for index, layer := range layers {
		for _, path := range computedKeys {
			if hasPath(layer, path) {
				problems = append(problems, fmt.Sprintf("%s sets %s", names[index], strings.Join(path, ".")))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %s; change the build instead", ErrComputedKey, strings.Join(problems, ", "))
	}
	return nil
}

func hasPath(tree map[string]any, path []string) bool {
	for index, key := range path {
		value, found := tree[key]
		if !found {
			return false
		}
		if index == len(path)-1 {
			return true
		}
		next, isTable := value.(map[string]any)
		if !isTable {
			return false
		}
		tree = next
	}
	return false
}

func (creature *Creature) checkBuild() []string {
	build := creature.Build
	var problems []string
	report := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if creature.Kind != KindPC {
		report("build: kind %q: only a pc has a build", creature.Kind)
	}
	if creature.Level < 1 {
		report("build: level must be at least 1")
	}
	if !slices.Contains(AttributeNames, build.Key) {
		report("build.key %q: not an attribute", build.Key)
	}
	if build.AncestryHP < 0 {
		report("build.ancestry_hp must not be negative")
	}
	if build.ClassHP < 1 {
		report("build.class_hp must be at least 1")
	}
	if build.Speed < 0 {
		report("build.speed must not be negative")
	}
	if _, known := specializationDamage[build.WeaponSpecialization]; build.WeaponSpecialization != "" && !known {
		report("build.weapon_specialization %q: want normal or greater", build.WeaponSpecialization)
	}
	if build.CriticalSpecialization != "" && slices.Index(RankWords, build.CriticalSpecialization) < 1 {
		report("build.critical_specialization %q: want a rank from trained to legendary", build.CriticalSpecialization)
	}
	for _, key := range slices.Sorted(maps.Keys(build.Proficiency)) {
		if !IsProficiencyKey(key) {
			report("build.proficiency.%s: unknown proficiency", key)
		}
		if rank := build.Proficiency[key]; !slices.Contains(RankWords, rank) {
			report("build.proficiency.%s: rank %q: want %s", key, rank, strings.Join(RankWords, ", "))
		}
	}
	if armor := build.Armor; armor != nil {
		if !slices.Contains(armorCategories, armor.Category) {
			report("build.armor.category %q: want %s", armor.Category, strings.Join(armorCategories, ", "))
		}
		if armor.ACBonus < 0 || armor.Potency < 0 || armor.Resilient < 0 || (armor.DexCap != nil && *armor.DexCap < 0) {
			report("build.armor: ac_bonus, dex_cap, potency and resilient must not be negative")
		}
		if armor.SpeedPenalty > 0 {
			report("build.armor.speed_penalty must not be positive")
		}
	}
	for _, id := range slices.Sorted(maps.Keys(build.Weapons)) {
		weapon := build.Weapons[id]
		if weapon.Name == "" {
			report("build.weapon.%s: name is empty", id)
		}
		if !slices.Contains(weaponCategories, weapon.Category) {
			report("build.weapon.%s: category %q: want %s", id, weapon.Category, strings.Join(weaponCategories, ", "))
		}
		if weapon.Dice < 1 || weapon.Die < 2 {
			report("build.weapon.%s: dice must be at least 1 and die at least 2", id)
		}
		if !IsDamageType(weapon.Type) {
			report("build.weapon.%s: unknown damage type %q", id, weapon.Type)
		}
		if weapon.Potency < 0 || weapon.Striking < 0 || weapon.ItemBonus < 0 || weapon.Range < 0 || weapon.Splash < 0 {
			report("build.weapon.%s: potency, striking, item_bonus, range and splash must not be negative", id)
		}
		if persistent := weapon.Persistent; persistent != nil {
			if _, err := dice.Parse(persistent.Dice); err != nil || !IsDamageType(persistent.Type) {
				report("build.weapon.%s.persistent: want parsable dice and a damage type", id)
			}
		}
	}
	for _, id := range slices.Sorted(maps.Keys(build.Modifiers)) {
		modifier := build.Modifiers[id]
		if !slices.Contains(modifierSelectors, modifier.Selector) {
			report("build.modifier.%s: selector %q: want %s", id, modifier.Selector, strings.Join(modifierSelectors, ", "))
		}
		if !slices.Contains(modifierTypes, modifier.Type) {
			report("build.modifier.%s: type %q: want item, status, circumstance or none", id, modifier.Type)
		}
		if len(modifier.WeaponTraits) > 0 && !slices.Contains(weaponSelectors, modifier.Selector) {
			report("build.modifier.%s: weapon_traits needs selector attack, damage or splash", id)
		}
		if modifier.Attribute != "" && (modifier.Selector != "splash" || !slices.Contains(AttributeNames, modifier.Attribute)) {
			report("build.modifier.%s: attribute needs selector splash and an attribute name", id)
		}
	}
	if casting := build.Spellcasting; casting != nil {
		if !slices.Contains(AttributeNames, casting.Attribute) {
			report("build.spellcasting.attribute %q: not an attribute", casting.Attribute)
		}
		if casting.TopRank < 0 || casting.TopRank > 10 {
			report("build.spellcasting.top_rank %d: want 0 to 10", casting.TopRank)
		}
	}
	return problems
}

func (creature *Creature) computeBuild() error {
	if problems := creature.checkBuild(); len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
	}
	build := creature.Build
	attributes, err := build.Attributes()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	level := creature.Level
	creature.Attributes = attributes

	armor := Armor{Category: "unarmored"}
	if build.Armor != nil {
		armor = *build.Armor
	}
	dexterity := attributes["dex"]
	if armor.DexCap != nil {
		dexterity = min(dexterity, *armor.DexCap)
	}
	saveTotal := func(save string) int {
		return build.proficiency(level, save) + attributes[saveAttributes[save]] + build.total(level, armor.Resilient, nil, save, "saving-throw")
	}

	creature.Defense.HP = build.AncestryHP + level*(build.ClassHP+attributes["con"]) + build.total(level, 0, nil, "hp")
	creature.Defense.AC = 10 + dexterity + build.proficiency(level, armor.Category) + build.total(level, armor.ACBonus+armor.Potency, nil, "ac")
	creature.Defense.Fortitude = saveTotal("fortitude")
	creature.Defense.Reflex = saveTotal("reflex")
	creature.Defense.Will = saveTotal("will")
	creature.Perception = build.proficiency(level, "perception") + attributes["wis"] + build.total(level, 0, nil, "perception")
	creature.Speed.Land = build.Speed + armor.SpeedPenalty + build.total(level, 0, nil, "land-speed")
	creature.ClassDC = 10 + build.proficiency(level, "class-dc") + attributes[build.Key]
	if casting := build.Spellcasting; casting != nil {
		attack := build.proficiency(level, "spellcasting") + attributes[casting.Attribute]
		creature.Spellcasting = &Spellcasting{Attack: attack, DC: 10 + attack, TopRank: casting.TopRank}
	}
	creature.Strikes = map[string]Strike{}
	for id, weapon := range build.Weapons {
		creature.Strikes[id] = build.strike(level, attributes, weapon)
	}
	return nil
}

func (build *Build) rank(key string) int {
	return max(slices.Index(RankWords, build.Proficiency[key]), 0)
}

func (build *Build) proficiency(level int, key string) int {
	return proficiencyBonus(level, build.rank(key))
}

func proficiencyBonus(level, rank int) int {
	if rank == 0 {
		return 0
	}
	return level + 2*rank
}

func (build *Build) total(level, itemBonus int, weaponTraits []string, selectors ...string) int {
	pool := NewBonusPool()
	pool.Add("item", itemBonus)
	for _, id := range slices.Sorted(maps.Keys(build.Modifiers)) {
		modifier := build.Modifiers[id]
		if !slices.Contains(selectors, modifier.Selector) || !containsAll(weaponTraits, modifier.WeaponTraits) {
			continue
		}
		pool.Add(modifier.Type, modifier.Value+modifier.PerLevel*level)
	}
	return pool.Total()
}

func (build *Build) strike(level int, attributes map[string]int, weapon Weapon) Strike {
	rank := max(build.rank(weapon.Category), build.rank(WeaponBasePrefix+weapon.Base))
	if weapon.Category != "advanced" {
		rank = max(rank, build.rank(WeaponGroupPrefix+weapon.Group))
	}
	attack := proficiencyBonus(level, rank) + attackAttribute(attributes, weapon) + build.total(level, max(weapon.Potency, weapon.ItemBonus), weapon.Traits, "attack")
	modifier := damageAttribute(attributes, weapon) + build.specialization(rank) + build.total(level, 0, weapon.Traits, "damage")
	expression := dice.Expression{Count: weapon.Dice + weapon.Striking, Sides: weapon.Die, Modifier: modifier}
	return Strike{
		Name:     weapon.Name,
		Bonus:    attack,
		Range:    weapon.Range,
		Traits:   slices.Clone(weapon.Traits),
		Damage:   build.weaponDamage(level, attributes, weapon, expression),
		Outcomes: build.riders(weapon, rank),
	}
}

func (build *Build) riders(weapon Weapon, rank int) Outcomes {
	outcomes := weapon.Outcomes
	minimum := slices.Index(RankWords, build.CriticalSpecialization)
	if specialization, found := criticalSpecializations[weapon.Group]; found && minimum > 0 && rank >= minimum {
		outcomes.CriticalSuccess = append(slices.Clone(outcomes.For(3, true)), specialization)
	}
	return outcomes
}

func (build *Build) weaponDamage(level int, attributes map[string]int, weapon Weapon, expression dice.Expression) []DamagePart {
	damage := []DamagePart{{Dice: expression.String(), Type: weapon.Type}}
	if persistent := weapon.Persistent; persistent != nil {
		damage = append(damage, DamagePart{Dice: persistent.Dice, Type: persistent.Type, Category: "persistent"})
	}
	if weapon.Splash > 0 {
		splash := weapon.Splash + build.total(level, 0, weapon.Traits, "splash") + build.attributeBonus(attributes, weapon.Traits, "splash")
		damage = append(damage, DamagePart{Dice: strconv.Itoa(splash), Type: weapon.Type, Category: "splash"})
	}
	return damage
}

func (build *Build) attributeBonus(attributes map[string]int, weaponTraits []string, selector string) int {
	bonus := 0
	for _, modifier := range build.Modifiers {
		if modifier.Selector == selector && modifier.Attribute != "" && containsAll(weaponTraits, modifier.WeaponTraits) {
			bonus += attributes[modifier.Attribute]
		}
	}
	return bonus
}

func (build *Build) specialization(rank int) int {
	table, found := specializationDamage[build.WeaponSpecialization]
	if !found {
		return 0
	}
	return table[rank]
}

func attackAttribute(attributes map[string]int, weapon Weapon) int {
	switch {
	case weapon.Range > 0:
		return attributes["dex"]
	case slices.Contains(weapon.Traits, "finesse"):
		return max(attributes["str"], attributes["dex"])
	default:
		return attributes["str"]
	}
}

func damageAttribute(attributes map[string]int, weapon Weapon) int {
	strength := attributes["str"]
	switch {
	case weapon.Range == 0 || isThrown(weapon.Traits):
		return strength
	case slices.Contains(weapon.Traits, "propulsive") && strength > 0:
		return strength / 2
	case slices.Contains(weapon.Traits, "propulsive"):
		return strength
	default:
		return 0
	}
}

func isThrown(traits []string) bool {
	return slices.ContainsFunc(traits, func(trait string) bool {
		return trait == "thrown" || strings.HasPrefix(trait, "thrown-")
	})
}

func containsAll(have, want []string) bool {
	for _, entry := range want {
		if !slices.Contains(have, entry) {
			return false
		}
	}
	return true
}

type BonusPool struct {
	untyped   int
	bonuses   map[string]int
	penalties map[string]int
}

func NewBonusPool() *BonusPool {
	return &BonusPool{bonuses: map[string]int{}, penalties: map[string]int{}}
}

func (pool *BonusPool) Add(kind string, value int) {
	switch {
	case kind == "":
		pool.untyped += value
	case value >= 0:
		pool.bonuses[kind] = max(pool.bonuses[kind], value)
	default:
		pool.penalties[kind] = min(pool.penalties[kind], value)
	}
}

func (pool *BonusPool) Total() int {
	total := pool.untyped
	for _, value := range pool.bonuses {
		total += value
	}
	for _, value := range pool.penalties {
		total += value
	}
	return total
}
