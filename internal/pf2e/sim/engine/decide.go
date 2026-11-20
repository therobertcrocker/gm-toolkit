package engine

type ChoiceKind int

const (
	ChooseStrike ChoiceKind = iota
	ChooseAbility
	ChooseStand
	ChooseSustain
)

type Choice struct {
	Kind    ChoiceKind
	ID      string
	Target  *Combatant
	Actions int
}

type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
	React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool)
}

// deliberate: abilities and Strikes are tried in ID order, not file order, because merged keyed tables keep no order (E1 Decision 6).
type FirstFitting struct{}

func (behaviour FirstFitting) Decide(fight *Fight, actor *Combatant) (Choice, bool) {
	target, found := behaviour.NextTarget(fight, actor)
	if !found {
		return Choice{}, false
	}
	if actor.Condition("prone") > 0 && actor.OpenActions() > 0 {
		return Choice{Kind: ChooseStand}, true
	}
	if waiting := actor.Sustaining(); len(waiting) > 0 && actor.OpenActions() > 0 {
		return Choice{Kind: ChooseSustain, ID: waiting[0]}, true
	}
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Trigger == nil && ability.Targets == "" && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
			return Choice{Kind: ChooseAbility, ID: id, Target: target}, true
		}
	}
	strikeIDs := actor.Creature.StrikeIDs()
	if len(strikeIDs) == 0 {
		return Choice{}, false
	}
	return Choice{Kind: ChooseStrike, ID: strikeIDs[0], Target: target}, true
}

func (FirstFitting) NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool) {
	foes := fight.StandingFoes(actor)
	if len(foes) == 0 {
		return nil, false
	}
	return foes[0], true
}

// deliberate: FirstFitting always reacts with the first ability offered, not by weighing the trigger, because it is a test behaviour; Effort 7's rule lists decide (Decision 21).
func (FirstFitting) React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool) {
	return offered[0], true
}
