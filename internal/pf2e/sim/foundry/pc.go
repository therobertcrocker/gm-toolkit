package foundry

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var (
	specializationSlugs = map[string]string{"weapon-specialization": "normal", "greater-weapon-specialization": "greater"}
	abilityActionTypes  = []string{"action", "reaction", "free"}
)

type characterSystem struct {
	Details levelDetails   `json:"details"`
	Build   characterBuild `json:"build"`
}

type characterBuild struct {
	Attributes buildAttributes `json:"attributes"`
}

type buildAttributes struct {
	Boosts map[string][]string `json:"boosts"`
}

type boostChoice struct {
	Value    []string `json:"value"`
	Selected string   `json:"selected"`
}

type classSystem struct {
	HP           int            `json:"hp"`
	KeyAbility   boostChoice    `json:"keyAbility"`
	Perception   int            `json:"perception"`
	SavingThrows map[string]int `json:"savingThrows"`
	Attacks      classAttacks   `json:"attacks"`
	Defenses     map[string]int `json:"defenses"`
	Spellcasting int            `json:"spellcasting"`
}

type classAttacks struct {
	Simple   int `json:"simple"`
	Martial  int `json:"martial"`
	Advanced int `json:"advanced"`
	Unarmed  int `json:"unarmed"`
}

type ancestrySystem struct {
	HP                      int                    `json:"hp"`
	Speed                   int                    `json:"speed"`
	Boosts                  map[string]boostChoice `json:"boosts"`
	Flaws                   map[string]boostChoice `json:"flaws"`
	AlternateAncestryBoosts []string               `json:"alternateAncestryBoosts"`
	Traits                  itemTraits             `json:"traits"`
}

type backgroundSystem struct {
	Boosts map[string]boostChoice `json:"boosts"`
}

type featSystem struct {
	Subfeatures subfeatures `json:"subfeatures"`
}

type subfeatures struct {
	Proficiencies map[string]rankEntry `json:"proficiencies"`
}

type rankEntry struct {
	Rank int `json:"rank"`
}

type armorSystem struct {
	Category     string     `json:"category"`
	Group        string     `json:"group"`
	ACBonus      int        `json:"acBonus"`
	DexCap       int        `json:"dexCap"`
	Strength     int        `json:"strength"`
	SpeedPenalty int        `json:"speedPenalty"`
	Runes        armorRunes `json:"runes"`
}

type armorRunes struct {
	Potency   int      `json:"potency"`
	Resilient int      `json:"resilient"`
	Property  []string `json:"property"`
}

type weaponSystem struct {
	Category     string       `json:"category"`
	Group        string       `json:"group"`
	BaseItem     string       `json:"baseItem"`
	Damage       weaponDamage `json:"damage"`
	Runes        weaponRunes  `json:"runes"`
	Range        *int         `json:"range"`
	Bonus        valueOf[int] `json:"bonus"`
	SplashDamage valueOf[int] `json:"splashDamage"`
}

type weaponDamage struct {
	Dice       int               `json:"dice"`
	Die        string            `json:"die"`
	DamageType string            `json:"damageType"`
	Persistent *persistentDamage `json:"persistent"`
}

type persistentDamage struct {
	Number int    `json:"number"`
	Type   string `json:"type"`
}

type weaponRunes struct {
	Potency  int      `json:"potency"`
	Striking int      `json:"striking"`
	Property []string `json:"property"`
}

type spellSystem struct {
	Location valueOf[string] `json:"location"`
}

type pcImporter struct {
	source                 *actor
	report                 *Report
	level                  int
	build                  *creature.Build
	traits                 []string
	classSlug              string
	ranks                  map[string]int
	resist                 map[string]creature.Adjustment
	armor                  *armorSystem
	armorName              string
	dexterityCaps          []int
	speedPenaltySuppressed bool
	weaponIDs              map[string]bool
	modifierIDs            map[string]bool
	entries                []*item
	spellsByEntry          map[string]int
}

func importPC(source *actor) (*creature.Creature, *Report, error) {
	var system characterSystem
	if err := source.decode(&system); err != nil {
		return nil, nil, err
	}
	importer := &pcImporter{
		source: source,
		report: &Report{Kind: creature.KindPC},
		level:  system.Details.Level.Value,
		build: &creature.Build{
			Boosts:      map[string][]string{},
			Proficiency: map[string]string{},
			Weapons:     map[string]creature.Weapon{},
			Modifiers:   map[string]creature.Modifier{},
		},
		ranks:         map[string]int{"class-dc": 1},
		resist:        map[string]creature.Adjustment{},
		weaponIDs:     map[string]bool{},
		modifierIDs:   map[string]bool{},
		spellsByEntry: map[string]int{},
	}
	for level, boosts := range system.Build.Attributes.Boosts {
		importer.build.Boosts["level-"+level] = boosts
	}
	for index := range source.Items {
		if current := &source.Items[index]; current.Type == "class" {
			if err := importer.class(current); err != nil {
				return nil, nil, err
			}
		}
	}
	for index := range source.Items {
		if err := importer.item(&source.Items[index]); err != nil {
			return nil, nil, err
		}
	}
	importer.applyRules()
	if err := importer.spellcasting(); err != nil {
		return nil, nil, err
	}
	return importer.finish(), importer.report, nil
}

func (importer *pcImporter) item(current *item) error {
	switch current.Type {
	case "ancestry":
		return importer.ancestry(current)
	case "background":
		return importer.background(current)
	case "feat":
		return importer.feat(current)
	case "armor":
		return importer.armorItem(current)
	case "weapon":
		return importer.weapon(current)
	case "spellcastingEntry":
		importer.entries = append(importer.entries, current)
	case "spell":
		var system spellSystem
		if err := current.decode(&system); err != nil {
			return err
		}
		importer.spellsByEntry[system.Location.Value]++
	case "action":
		importer.report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
	case "condition", "effect":
		importer.report.skip(current.Name, current.Type, ReasonState)
	case "shield":
		importer.report.skip(current.Name, "shield", ReasonEffort3)
	}
	return nil
}

func (importer *pcImporter) raise(key string, rank int) {
	importer.ranks[key] = max(importer.ranks[key], rank)
}

func (importer *pcImporter) class(current *item) error {
	var system classSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.classSlug = current.System.Slug
	importer.build.ClassHP = system.HP
	switch {
	case system.KeyAbility.Selected != "":
		importer.build.Key = system.KeyAbility.Selected
	case len(system.KeyAbility.Value) == 1:
		importer.build.Key = system.KeyAbility.Value[0]
	default:
		importer.report.skip(current.Name, "key attribute not chosen", ReasonUnsupported)
	}
	if importer.build.Key != "" {
		importer.build.Boosts["class"] = []string{importer.build.Key}
	}
	importer.raise("perception", system.Perception)
	for _, save := range slices.Sorted(maps.Keys(system.SavingThrows)) {
		importer.raise(save, system.SavingThrows[save])
	}
	importer.raise("simple", system.Attacks.Simple)
	importer.raise("martial", system.Attacks.Martial)
	importer.raise("advanced", system.Attacks.Advanced)
	importer.raise("unarmed", system.Attacks.Unarmed)
	for _, category := range slices.Sorted(maps.Keys(system.Defenses)) {
		importer.raise(category, system.Defenses[category])
	}
	importer.raise("spellcasting", system.Spellcasting)
	return nil
}

func (importer *pcImporter) ancestry(current *item) error {
	var system ancestrySystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.build.AncestryHP = system.HP
	importer.build.Speed = system.Speed
	importer.traits = system.Traits.Value
	if system.AlternateAncestryBoosts != nil {
		importer.build.Boosts["ancestry"] = system.AlternateAncestryBoosts
		return nil
	}
	importer.build.Boosts["ancestry"] = importer.choices(current.Name, "boost", system.Boosts)
	importer.build.Flaws = importer.choices(current.Name, "flaw", system.Flaws)
	return nil
}

func (importer *pcImporter) background(current *item) error {
	var system backgroundSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.build.Boosts["background"] = importer.choices(current.Name, "boost", system.Boosts)
	return nil
}

func (importer *pcImporter) choices(owner, label string, entries map[string]boostChoice) []string {
	var chosen []string
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		entry := entries[key]
		switch {
		case entry.Selected != "":
			chosen = append(chosen, entry.Selected)
		case len(entry.Value) == 1:
			chosen = append(chosen, entry.Value[0])
		default:
			importer.report.skip(owner, fmt.Sprintf("%s %s not chosen", label, key), ReasonUnsupported)
		}
	}
	return chosen
}

func (importer *pcImporter) feat(current *item) error {
	if level, found := specializationSlugs[current.System.Slug]; found && importer.build.WeaponSpecialization != "greater" {
		importer.build.WeaponSpecialization = level
	}
	if actionType := current.System.ActionType.Value; slices.Contains(abilityActionTypes, actionType) {
		importer.report.skip(current.Name, actionType, ReasonAbility)
	}
	var system featSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	for _, key := range slices.Sorted(maps.Keys(system.Subfeatures.Proficiencies)) {
		target := key
		if key == importer.classSlug {
			target = "class-dc"
		}
		if !creature.IsProficiencyKey(target) {
			importer.report.skip(current.Name, "proficiency "+key, ReasonUnsupported)
			continue
		}
		importer.raise(target, system.Subfeatures.Proficiencies[key].Rank)
	}
	return nil
}

func (importer *pcImporter) armorItem(current *item) error {
	equipped := current.System.Equipped
	if equipped.CarryType != "worn" || (equipped.Invested != nil && !*equipped.Invested) {
		importer.report.skip(current.Name, "armor", ReasonInactive)
		return nil
	}
	if importer.armor != nil {
		importer.report.skip(current.Name, "second worn armor", ReasonUnsupported)
		return nil
	}
	var system armorSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.armor = &system
	importer.armorName = current.Name
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonEffort3)
	}
	return nil
}

func (importer *pcImporter) weapon(current *item) error {
	if current.System.Quantity < 1 {
		importer.report.skip(current.Name, "weapon", ReasonQuantity)
		return nil
	}
	var system weaponSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	die, err := strconv.Atoi(strings.TrimPrefix(system.Damage.Die, "d"))
	if err != nil || die < 2 || system.Damage.Dice < 1 {
		importer.report.skip(current.Name, fmt.Sprintf("damage %d%s", system.Damage.Dice, system.Damage.Die), ReasonUnsupported)
		return nil
	}
	if !creature.IsDamageType(system.Damage.DamageType) {
		importer.report.skip(current.Name, "damage type "+system.Damage.DamageType, ReasonUnsupported)
		return nil
	}
	weapon := creature.Weapon{
		Name:      current.Name,
		Category:  system.Category,
		Group:     system.Group,
		Base:      system.BaseItem,
		Dice:      system.Damage.Dice,
		Die:       die,
		Type:      system.Damage.DamageType,
		Potency:   system.Runes.Potency,
		Striking:  system.Runes.Striking,
		ItemBonus: system.Bonus.Value,
		Traits:    current.System.Traits.Value,
	}
	if system.Range != nil {
		weapon.Range = *system.Range
	}
	if persistent := system.Damage.Persistent; persistent != nil {
		importer.report.skip(current.Name, fmt.Sprintf("persistent %d %s", persistent.Number, persistent.Type), ReasonEffort3)
	}
	if system.SplashDamage.Value > 0 {
		importer.report.skip(current.Name, fmt.Sprintf("splash %d", system.SplashDamage.Value), ReasonEffort3)
	}
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonEffort3)
	}
	if slices.Contains(weapon.Traits, "consumable") {
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, not tracked", current.System.Quantity), ReasonEffort3)
	}
	importer.build.Weapons[uniqueID(current.Name, importer.weaponIDs)] = weapon
	return nil
}

func (importer *pcImporter) spellcasting() error {
	slices.SortStableFunc(importer.entries, func(first, second *item) int { return strings.Compare(first.Name, second.Name) })
	var casting *creature.Casting
	for _, entry := range importer.entries {
		var system spellcastingEntrySystem
		if err := entry.decode(&system); err != nil {
			return err
		}
		ability := system.Ability.Value
		switch {
		case importer.spellsByEntry[entry.ID] == 0:
			importer.report.skip(entry.Name, "spellcasting entry with no spells", ReasonEmpty)
			continue
		case ability == "":
			importer.report.skip(entry.Name, "spellcasting entry on a class statistic", ReasonUnsupported)
			continue
		case casting == nil:
			casting = &creature.Casting{Attribute: ability}
		case casting.Attribute != ability:
			importer.report.skip(entry.Name, "spellcasting attribute "+ability, ReasonUnsupported)
			continue
		}
		casting.TopRank = max(casting.TopRank, topSlotRank(system.Slots))
	}
	importer.build.Spellcasting = casting
	spells := 0
	for _, count := range importer.spellsByEntry {
		spells += count
	}
	if spells > 0 {
		importer.report.skip(importer.source.Name, fmt.Sprintf("%d spells", spells), ReasonEffort3)
	}
	return nil
}

func topSlotRank(slots map[string]slotEntry) int {
	top := 0
	for key, slot := range slots {
		rank, err := strconv.Atoi(strings.TrimPrefix(key, "slot"))
		if err == nil && slot.Max > 0 {
			top = max(top, rank)
		}
	}
	return top
}

func (importer *pcImporter) finish() *creature.Creature {
	build := importer.build
	for _, key := range slices.Sorted(maps.Keys(importer.ranks)) {
		if rank := importer.ranks[key]; rank > 0 {
			build.Proficiency[key] = creature.RankWords[rank]
		}
	}
	build.Armor = importer.finishArmor()
	layer := &creature.Creature{
		Schema: creature.SchemaVersion,
		Name:   importer.source.Name,
		Kind:   creature.KindPC,
		Level:  importer.level,
		Traits: importer.traits,
		Build:  build,
	}
	if len(importer.resist) > 0 {
		layer.Defense.Resist = importer.resist
	}
	return layer
}

func (importer *pcImporter) finishArmor() *creature.Armor {
	var lowestCap *int
	if len(importer.dexterityCaps) > 0 {
		lowest := slices.Min(importer.dexterityCaps)
		lowestCap = &lowest
	}
	if importer.armor == nil {
		if lowestCap == nil {
			return nil
		}
		return &creature.Armor{Category: "unarmored", DexCap: lowestCap}
	}
	system := importer.armor
	armorCap := system.DexCap
	if lowestCap != nil {
		armorCap = min(armorCap, *lowestCap)
	}
	return &creature.Armor{
		Name:         importer.armorName,
		Category:     system.Category,
		ACBonus:      system.ACBonus,
		DexCap:       &armorCap,
		Potency:      system.Runes.Potency,
		Resilient:    system.Runes.Resilient,
		SpeedPenalty: importer.speedPenalty(),
	}
}

func (importer *pcImporter) speedPenalty() int {
	system := importer.armor
	if importer.speedPenaltySuppressed || system.SpeedPenalty >= 0 {
		return 0
	}
	attributes, err := importer.build.Attributes()
	if err != nil {
		importer.report.skip(importer.armorName, "speed penalty: "+err.Error(), ReasonUnsupported)
		return system.SpeedPenalty
	}
	if attributes["str"] >= system.Strength {
		return min(system.SpeedPenalty+5, 0)
	}
	return system.SpeedPenalty
}
