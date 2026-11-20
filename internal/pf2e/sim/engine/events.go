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

type ActionsRegained struct {
	Combatant string
	Actions   int
	Quickened bool
	Lost      int
}

type ActionsLost struct {
	Combatant string
	Condition string
	Lost      int
	Left      int
}

type StoodUp struct {
	Combatant string
}

type ActionUsed struct {
	Combatant string
	Action    string
	Actions   int
	Target    string
}

type UseSpent struct {
	Combatant string
	Ability   string
	Left      int
}

type PoolSpent struct {
	Combatant string
	Pool      string
	Ability   string
	Left      int
}

type RechargeRolled struct {
	Combatant string
	Ability   string
	Rounds    int
}

type Recharged struct {
	Combatant string
	Ability   string
	By        string
}

type ActionsGained struct {
	Combatant string
	Ability   string
	Amount    int
}

type Triggered struct {
	Combatant string
	Ability   string
	On        string
	Cost      string
	By        string
	ByAbility string
	ByStrike  string
}

type LoyaltyAct struct {
	Combatant string
	Reason    string
}

type TraitGranted struct {
	Combatant string
	Trait     string
	Against   string
	Ability   string
}

type TraitSpent struct {
	Combatant string
	Trait     string
	Against   string
}

type ActionDisrupted struct {
	Combatant string
	Action    string
	By        string
	Ability   string
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

type SkillChecked struct {
	Combatant string
	Target    string
	Ability   string
	Skill     string
	Die       int
	Bonus     int
	Total     int
	DC        int
	Degree    Degree
}

type RerollGained struct {
	Combatant string
	Source    string
	Ability   string
	Keep      string
}

type Rerolled struct {
	Combatant string
	Roll      string
	Keep      string
	First     int
	Second    int
	Kept      int
	Ability   string
}

type RerollsCancelled struct {
	Combatant string
	Roll      string
}

type BonusGained struct {
	Combatant string
	Source    string
	Ability   string
	Selector  string
	Value     int
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
	Source      string
	Target      string
	Strike      string
	Ability     string
	Critical    bool
	Parts       []DamageAmount
	Total       int
	Absorbed    int
	HPAfter     int
	Trait       int
	TriggeredBy string
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
	Source      string
	Target      string
	Ability     string
	Amount      int
	HPAfter     int
	TriggeredBy string
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

type RecoveryEased struct {
	Combatant string
	Ability   string
	DC        int
}

type Sustained struct {
	Combatant string
	Ability   string
}

type SustainEnded struct {
	Combatant string
	Ability   string
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
func (ActionsRegained) isEvent()   {}
func (ActionsLost) isEvent()       {}
func (StoodUp) isEvent()           {}
func (ActionUsed) isEvent()        {}
func (UseSpent) isEvent()          {}
func (PoolSpent) isEvent()         {}
func (RechargeRolled) isEvent()    {}
func (Recharged) isEvent()         {}
func (ActionsGained) isEvent()     {}
func (Triggered) isEvent()         {}
func (ActionDisrupted) isEvent()   {}
func (StrikeRolled) isEvent()      {}
func (SaveRolled) isEvent()        {}
func (SkillChecked) isEvent()      {}
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
func (RerollGained) isEvent()      {}
func (Rerolled) isEvent()          {}
func (RerollsCancelled) isEvent()  {}
func (BonusGained) isEvent()       {}
func (RecoveryEased) isEvent()     {}
func (Sustained) isEvent()         {}
func (SustainEnded) isEvent()      {}
func (LoyaltyAct) isEvent()        {}
func (TraitGranted) isEvent()      {}
func (TraitSpent) isEvent()        {}
