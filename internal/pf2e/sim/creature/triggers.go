package creature

import (
	"maps"
	"slices"
	"strings"
)

const (
	TriggerTargeted   = "targeted"
	TriggerAttacked   = "attacked"
	TriggerSaved      = "saved"
	TriggerDamaged    = "damaged"
	TriggerTurnStart  = "turn-start"
	TriggerInitiative = "initiative"
	TriggerFoeActs    = "foe-acts"
	TriggerAllyGains  = "ally-gains"
	TriggerStoodUp    = "stood-up"
	TriggerLoyalty    = "loyalty"

	CostReaction = "reaction"
	CostFree     = "free"

	AttackStrike      = "strike"
	AttackMeleeStrike = "melee-strike"
)

var (
	triggerKinds   = []string{TriggerTargeted, TriggerAttacked, TriggerSaved, TriggerDamaged, TriggerTurnStart, TriggerInitiative, TriggerFoeActs, TriggerAllyGains, TriggerStoodUp, TriggerLoyalty}
	triggerCosts   = []string{"", CostReaction, CostFree}
	triggerAttacks = []string{"", AttackStrike, AttackMeleeStrike}
)

type Trigger struct {
	On         string   `toml:"on"`
	Cost       string   `toml:"cost,omitempty"`
	Attack     string   `toml:"attack,omitempty"`
	Degrees    []string `toml:"degrees,omitempty"`
	Damaging   bool     `toml:"damaging,omitempty"`
	Melee      bool     `toml:"melee,omitempty"`
	Traits     []string `toml:"traits,omitempty"`
	Ranged     bool     `toml:"ranged,omitempty"`
	Conditions []string `toml:"conditions,omitempty"`
	HPAtMost   int      `toml:"hp_at_most,omitzero"`
}

func checkTrigger(path string, ability Ability, report func(string, ...any)) {
	trigger := ability.Trigger
	if !slices.Contains(triggerKinds, trigger.On) {
		report("%s.trigger: on %q: want %s", path, trigger.On, strings.Join(triggerKinds, ", "))
		return
	}
	if !slices.Contains(triggerCosts, trigger.Cost) {
		report("%s.trigger: cost %q: want reaction, free, or leave it unset", path, trigger.Cost)
	}
	if ability.Actions != 0 || ability.MaxActions != 0 || ability.Free || ability.TurnEnds {
		report("%s: a triggered ability sets no actions, max_actions, free or turn_ends; its cost is trigger.cost", path)
	}
	onAttack := trigger.On == TriggerTargeted || trigger.On == TriggerAttacked
	if !slices.Contains(triggerAttacks, trigger.Attack) || (trigger.Attack != "" && !onAttack) {
		report("%s.trigger: attack %q: want strike or melee-strike, on targeted or attacked", path, trigger.Attack)
	}
	if len(trigger.Degrees) > 0 && trigger.On != TriggerAttacked && trigger.On != TriggerSaved {
		report("%s.trigger: degrees belong on attacked or saved", path)
	}
	for _, degree := range trigger.Degrees {
		if !slices.Contains(DegreeWords, degree) {
			report("%s.trigger: degree %q: want %s", path, degree, strings.Join(DegreeWords, ", "))
		}
	}
	if trigger.Damaging && trigger.On != TriggerSaved {
		report("%s.trigger: damaging belongs on saved", path)
	}
	if trigger.Melee && trigger.On != TriggerDamaged {
		report("%s.trigger: melee belongs on damaged", path)
	}
	names := len(trigger.Traits) > 0 || trigger.Ranged
	if names != (trigger.On == TriggerFoeActs) {
		report("%s.trigger: foe-acts needs traits or ranged, and no other kind takes them", path)
	}
	if (len(trigger.Conditions) > 0) != (trigger.On == TriggerAllyGains) {
		report("%s.trigger: ally-gains needs conditions, and no other kind takes them", path)
	}
	for _, name := range trigger.Conditions {
		if !IsCondition(name) {
			report("%s.trigger: %q is not a condition", path, name)
		}
	}
	if trigger.HPAtMost < 0 || (trigger.HPAtMost != 0 && trigger.On != TriggerTurnStart) {
		report("%s.trigger: hp_at_most must be positive, on turn-start", path)
	}
}

func (creature *Creature) checkExtraReactions(report func(string, ...any)) {
	for _, id := range slices.Sorted(maps.Keys(creature.ExtraReactions)) {
		ability, found := creature.Abilities[id]
		if !found || ability.Trigger == nil || ability.Trigger.Cost != CostReaction {
			report("extra_reaction.%s: want an ability whose trigger costs a reaction", id)
		}
		if creature.ExtraReactions[id] < 1 {
			report("extra_reaction.%s: must be at least 1", id)
		}
	}
}
