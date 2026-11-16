package engine

type ChoiceKind int

const (
	ChooseStrike ChoiceKind = iota
	ChooseAbility
)

type Choice struct {
	Kind   ChoiceKind
	ID     string
	Target *Combatant
}

type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
}

// deliberate: abilities and Strikes are tried in ID order, not file order, because merged keyed tables keep no order (E1 Decision 6).
type FirstFitting struct{}

func (behaviour FirstFitting) Decide(fight *Fight, actor *Combatant) (Choice, bool) {
	target, found := behaviour.NextTarget(fight, actor)
	if !found {
		return Choice{}, false
	}
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Actions <= actor.ActionsLeft {
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
