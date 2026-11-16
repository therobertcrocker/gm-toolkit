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

type DamageDealt struct {
	Source   string
	Target   string
	Strike   string
	Critical bool
	Parts    []DamageAmount
	Total    int
	HPAfter  int
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

type FightEnded struct {
	Rounds  int
	Outcome Outcome
}

func (InitiativeRolled) isEvent() {}
func (RoundStarted) isEvent()     {}
func (TurnStarted) isEvent()      {}
func (ActionUsed) isEvent()       {}
func (StrikeRolled) isEvent()     {}
func (DamageDealt) isEvent()      {}
func (InitiativeMoved) isEvent()  {}
func (DyingChanged) isEvent()     {}
func (RecoveryChecked) isEvent()  {}
func (Stabilized) isEvent()       {}
func (Died) isEvent()             {}
func (FightEnded) isEvent()       {}
