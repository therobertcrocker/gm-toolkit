package foundry

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type npcSystem struct {
	Details    levelDetails  `json:"details"`
	Attributes npcAttributes `json:"attributes"`
	Perception npcPerception `json:"perception"`
	Saves      npcSaves      `json:"saves"`
	Traits     itemTraits    `json:"traits"`
}

type levelDetails struct {
	Level valueOf[int] `json:"level"`
}

type npcAttributes struct {
	HP          npcHP           `json:"hp"`
	AC          valueOf[int]    `json:"ac"`
	AllSaves    valueOf[string] `json:"allSaves"`
	Speed       npcSpeed        `json:"speed"`
	Immunities  []npcAdjustment `json:"immunities"`
	Resistances []npcAdjustment `json:"resistances"`
	Weaknesses  []npcAdjustment `json:"weaknesses"`
}

type npcHP struct {
	Max int `json:"max"`
}

type npcSpeed struct {
	Value       *int            `json:"value"`
	OtherSpeeds []npcOtherSpeed `json:"otherSpeeds"`
}

type npcOtherSpeed struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type npcAdjustment struct {
	Type       string   `json:"type"`
	Value      int      `json:"value"`
	Exceptions []string `json:"exceptions"`
	DoubleVs   []string `json:"doubleVs"`
}

type npcPerception struct {
	Mod int `json:"mod"`
}

type npcSaves struct {
	Fortitude valueOf[int] `json:"fortitude"`
	Reflex    valueOf[int] `json:"reflex"`
	Will      valueOf[int] `json:"will"`
}

type meleeSystem struct {
	Bonus         valueOf[int]      `json:"bonus"`
	DamageRolls   damageRolls       `json:"damageRolls"`
	Traits        itemTraits        `json:"traits"`
	AttackEffects valueOf[[]string] `json:"attackEffects"`
}

type spellcastingEntrySystem struct {
	Ability valueOf[string]      `json:"ability"`
	SpellDC spellDC              `json:"spelldc"`
	Slots   map[string]slotEntry `json:"slots"`
}

type spellDC struct {
	Value int `json:"value"`
	DC    int `json:"dc"`
}

type slotEntry struct {
	Max int `json:"max"`
}

func importNPC(source *actor) (*creature.Creature, *Report, error) {
	var system npcSystem
	if err := source.decode(&system); err != nil {
		return nil, nil, err
	}
	report := &Report{Kind: creature.KindNPC}
	attributes := system.Attributes
	layer := &creature.Creature{
		Schema:     creature.SchemaVersion,
		Name:       source.Name,
		Kind:       creature.KindNPC,
		Level:      system.Details.Level.Value,
		Traits:     npcTraits(system.Traits),
		Perception: system.Perception.Mod,
		Speed:      npcSpeedOf(attributes.Speed),
		Defense: creature.Defense{
			HP:        attributes.HP.Max,
			AC:        attributes.AC.Value,
			Fortitude: system.Saves.Fortitude.Value,
			Reflex:    system.Saves.Reflex.Value,
			Will:      system.Saves.Will.Value,
		},
	}
	if text := attributes.AllSaves.Value; text != "" {
		report.skip(source.Name, "saves: "+text, ReasonConditional)
	}
	for _, immunity := range attributes.Immunities {
		layer.Defense.Immune = append(layer.Defense.Immune, immunity.Type)
		if len(immunity.Exceptions) > 0 {
			report.skip(source.Name, fmt.Sprintf("immunity %s except %s", immunity.Type, strings.Join(immunity.Exceptions, ", ")), ReasonLaterRule)
		}
	}
	layer.Defense.Resist = npcAdjustments(attributes.Resistances, "resistance", source.Name, report)
	layer.Defense.Weakness = npcAdjustments(attributes.Weaknesses, "weakness", source.Name, report)

	strikeIDs := map[string]bool{}
	for index := range source.Items {
		current := &source.Items[index]
		var err error
		switch current.Type {
		case "melee":
			err = importNPCStrike(current, strikeIDs, layer, report)
		case "spellcastingEntry":
			err = importNPCSpellcasting(current, layer, report)
		case "spell":
			report.skip(current.Name, "spell", ReasonLaterRule)
		case "action":
			report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
		case "lore":
			report.skip(current.Name, "lore", ReasonNotCombat)
		default:
			report.skip(current.Name, current.Type, ReasonUnsupported)
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return layer, report, nil
}

func npcTraits(source itemTraits) []string {
	var traits []string
	if source.Rarity != "" && source.Rarity != "common" {
		traits = append(traits, source.Rarity)
	}
	return append(traits, source.Value...)
}

func npcSpeedOf(source npcSpeed) creature.Speed {
	var speed creature.Speed
	if source.Value != nil {
		speed.Land = *source.Value
	}
	for _, other := range source.OtherSpeeds {
		if other.Type == "fly" {
			speed.Fly = other.Value
		}
	}
	return speed
}

func npcAdjustments(entries []npcAdjustment, label, owner string, report *Report) map[string]creature.Adjustment {
	adjustments := map[string]creature.Adjustment{}
	for _, entry := range entries {
		key := entry.Type
		if key == "all-damage" {
			key = "all"
		}
		if key != "all" && key != "physical" && !creature.IsDamageType(key) && !creature.IsDefenseTrait(key) {
			report.skip(owner, fmt.Sprintf("%s %s %d", label, entry.Type, entry.Value), ReasonLaterRule)
			continue
		}
		if len(entry.DoubleVs) > 0 {
			report.skip(owner, fmt.Sprintf("%s %s doubled against %s", label, key, strings.Join(entry.DoubleVs, ", ")), ReasonLaterRule)
		}
		adjustments[key] = creature.Adjustment{Value: entry.Value, Except: entry.Exceptions}
	}
	if len(adjustments) == 0 {
		return nil
	}
	return adjustments
}

func strikeRange(traits []string) int {
	for _, trait := range traits {
		for _, prefix := range []string{"range-increment-", "range-"} {
			if feet, found := strings.CutPrefix(trait, prefix); found {
				if value, err := strconv.Atoi(feet); err == nil {
					return value
				}
			}
		}
	}
	return 0
}

func importNPCStrike(current *item, taken map[string]bool, layer *creature.Creature, report *Report) error {
	var system meleeSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	strike := creature.Strike{Name: current.Name, Bonus: system.Bonus.Value, Traits: system.Traits.Value, Range: strikeRange(system.Traits.Value)}
	for _, roll := range system.DamageRolls {
		what := fmt.Sprintf("%s %s damage", roll.Damage, roll.DamageType)
		if roll.Category != "" && roll.Category != "persistent" && roll.Category != "splash" {
			report.skip(current.Name, what+" "+roll.Category, ReasonLaterRule)
			continue
		}
		if _, err := dice.Parse(roll.Damage); err != nil || !creature.IsDamageType(roll.DamageType) {
			report.skip(current.Name, what, ReasonUnsupported)
			continue
		}
		strike.Damage = append(strike.Damage, creature.DamagePart{Dice: roll.Damage, Type: roll.DamageType, Category: roll.Category})
	}
	for _, effect := range system.AttackEffects.Value {
		report.skip(current.Name, "attack effect "+effect, ReasonLaterRule)
	}
	if !slices.ContainsFunc(strike.Damage, func(part creature.DamagePart) bool { return part.Category == "" }) {
		report.skip(current.Name, "strike with no damage the sim can roll", ReasonUnsupported)
		return nil
	}
	if layer.Strikes == nil {
		layer.Strikes = map[string]creature.Strike{}
	}
	layer.Strikes[uniqueID(current.Name, taken)] = strike
	return nil
}

func importNPCSpellcasting(current *item, layer *creature.Creature, report *Report) error {
	var system spellcastingEntrySystem
	if err := current.decode(&system); err != nil {
		return err
	}
	if layer.Spellcasting != nil {
		report.skip(current.Name, "second spellcasting entry", ReasonUnsupported)
		return nil
	}
	layer.Spellcasting = &creature.Spellcasting{Attack: system.SpellDC.Value, DC: system.SpellDC.DC}
	return nil
}
