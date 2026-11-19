package engine

type Observer interface {
	Observe(event Event)
}

type Event interface {
	isEvent()
}

type Outcome int

// deliberate: the zero value is NoOutcome, not PartyWins, because Run returns it beside an error and a tally must not count a failed fight as a win.
const (
	NoOutcome Outcome = iota
	PartyWins
	EnemiesWin
	Stalemate
)

func (outcome Outcome) String() string {
	switch outcome {
	case PartyWins:
		return "party wins"
	case EnemiesWin:
		return "enemies win"
	case Stalemate:
		return "stalemate"
	default:
		return "no outcome"
	}
}

type InitiativeRolled struct {
	Combatant string
	Die       int
	Total     int
}

type RoundStarted struct {
	Round int
}

type TurnStarted struct {
	Round     int
	Combatant string
}

type ActionUsed struct {
	Combatant string
	Action    string
	Actions   int
	Target    string
}

type StrikeRolled struct {
	Attacker string
	Target   string
	Strike   string
	Die      int
	Bonus    int
	Penalty  int
	Total    int
	AC       int
	Degree   Degree
}

type DamageAmount struct {
	Type   string
	Rolled int
	Taken  int
}

type SaveRolled struct {
	Combatant string
	Source    string
	Ability   string
	Save      string
	Die       int
	Bonus     int
	Total     int
	DC        int
	Degree    Degree
}

type SpellAttackRolled struct {
	Attacker string
	Target   string
	Ability  string
	Die      int
	Bonus    int
	Penalty  int
	Total    int
	AC       int
	Degree   Degree
}

type DamageDealt struct {
	Source   string
	Target   string
	Strike   string
	Ability  string
	Critical bool
	Parts    []DamageAmount
	Total    int
	Absorbed int
	HPAfter  int
	Trait    int
}

type ResistanceAdapted struct {
	Combatant string
	Type      string
	Value     int
}

type DefenseChanged struct {
	Combatant  string
	Source     string
	Weakness   int
	Resistance int
	Types      []string
}

type DefenseRestored struct {
	Combatant string
	Types     []string
}

type Healed struct {
	Source  string
	Target  string
	Ability string
	Amount  int
	HPAfter int
}

type Revived struct {
	Combatant string
	DyingFrom int
	Wounded   int
}

type TempHPGained struct {
	Combatant string
	Amount    int
}

type PersistentChanged struct {
	Combatant string
	Type      string
	Dice      string
}

type PersistentEnded struct {
	Combatant string
	Type      string
}

type InitiativeMoved struct {
	Combatant string
	Before    string
}

type DyingChanged struct {
	Combatant string
	From      int
	To        int
}

type RecoveryChecked struct {
	Combatant string
	Die       int
	DC        int
	Degree    Degree
}

type Stabilized struct {
	Combatant string
	Wounded   int
}

type Died struct {
	Combatant string
}

type HeroicRecovery struct {
	Combatant string
}

type HPLost struct {
	Combatant string
	Reason    string
	Amount    int
	HPAfter   int
}

type ConditionChanged struct {
	Combatant string
	Condition string
	From      int
	To        int
}

type ImmuneToEffect struct {
	Combatant string
	Source    string
	Ability   string
}

type FlatChecked struct {
	Combatant string
	Reason    string
	Die       int
	DC        int
}

type FightEnded struct {
	Rounds  int
	Outcome Outcome
}

func (InitiativeRolled) isEvent()  {}
func (RoundStarted) isEvent()      {}
func (TurnStarted) isEvent()       {}
func (ActionUsed) isEvent()        {}
func (StrikeRolled) isEvent()      {}
func (SaveRolled) isEvent()        {}
func (SpellAttackRolled) isEvent() {}
func (DamageDealt) isEvent()       {}
func (InitiativeMoved) isEvent()   {}
func (DyingChanged) isEvent()      {}
func (RecoveryChecked) isEvent()   {}
func (Stabilized) isEvent()        {}
func (Died) isEvent()              {}
func (ConditionChanged) isEvent()  {}
func (HeroicRecovery) isEvent()    {}
func (HPLost) isEvent()            {}
func (Healed) isEvent()            {}
func (Revived) isEvent()           {}
func (TempHPGained) isEvent()      {}
func (PersistentChanged) isEvent() {}
func (PersistentEnded) isEvent()   {}
func (ResistanceAdapted) isEvent() {}
func (DefenseChanged) isEvent()    {}
func (DefenseRestored) isEvent()   {}
func (ImmuneToEffect) isEvent()    {}
func (FlatChecked) isEvent()       {}
func (FightEnded) isEvent()        {}
