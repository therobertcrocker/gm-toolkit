package domain

// Goal is one rulebook goal: its text and the parts that state its behavior.
// Stat names are converted at load. Every other name in a part is a plain string
// that the goal engine checks when it is built (A.001.8 Effort 1 Decision 6).
type Goal struct {
	ID          string
	Name        string
	Description string
	Start       *GoalStart
	Progress    *GoalProgress
	Reset       *GoalEvent
	Threshold   *GoalThreshold
	Complete    *GoalEvent
	XP          *GoalXP
}

// GoalStart is what the pick needs before it may offer the goal.
type GoalStart struct {
	Inputs       []string
	NeedsActions []string
	Gate         *GoalGate
}

type GoalGate struct {
	RatingAboveLastCompletion FactionStat
}

// GoalEvent names an event and the filters a mutation must pass to match it.
type GoalEvent struct {
	Event   string
	Filters GoalFilters
}

// GoalProgress moves progress one of two ways: Event with Amount adds for each
// match, and Measure sets progress to a count read from state.
type GoalProgress struct {
	Event   string
	Amount  string
	Measure string
	Filters GoalFilters
}

// GoalThreshold is the progress that completes the goal: Constant, or the sum of
// the holder's Ratings times Multiplier (1 when it is 0).
type GoalThreshold struct {
	Ratings    []FactionStat
	Multiplier int
	Constant   int
}

type GoalXP struct {
	Formula        string
	Amount         int
	ContestedBonus int
}

const (
	GoalFilterOwner              = "owner"
	GoalFilterCausedBy           = "caused_by"
	GoalFilterFaction            = "faction"
	GoalFilterCategory           = "category"
	GoalFilterRatedAbove         = "rated_above"
	GoalFilterNewWorld           = "new_world"
	GoalFilterTo                 = "to"
	GoalFilterTarget             = "target"
	GoalFilterAction             = "action"
	GoalFilterActionNot          = "action_not"
	GoalFilterGround             = "ground"
	GoalFilterExcludeHeldAtStart = "exclude_held_at_start"
)

// GoalFilters narrows which mutations match a goal event or what a measure
// counts. A zero field is an absent filter; only "= true" filters for the bools.
type GoalFilters struct {
	Owner              string
	CausedBy           string
	Faction            string
	Category           []FactionStat
	RatedAbove         FactionStat
	NewWorld           bool
	To                 string
	Target             string
	Action             []string
	ActionNot          []string
	Ground             string
	ExcludeHeldAtStart bool
}

func (filters GoalFilters) Used() []string {
	var present []string
	if filters.Owner != "" {
		present = append(present, GoalFilterOwner)
	}
	if filters.CausedBy != "" {
		present = append(present, GoalFilterCausedBy)
	}
	if filters.Faction != "" {
		present = append(present, GoalFilterFaction)
	}
	if len(filters.Category) > 0 {
		present = append(present, GoalFilterCategory)
	}
	if filters.RatedAbove != "" {
		present = append(present, GoalFilterRatedAbove)
	}
	if filters.NewWorld {
		present = append(present, GoalFilterNewWorld)
	}
	if filters.To != "" {
		present = append(present, GoalFilterTo)
	}
	if filters.Target != "" {
		present = append(present, GoalFilterTarget)
	}
	if len(filters.Action) > 0 {
		present = append(present, GoalFilterAction)
	}
	if len(filters.ActionNot) > 0 {
		present = append(present, GoalFilterActionNot)
	}
	if filters.Ground != "" {
		present = append(present, GoalFilterGround)
	}
	if filters.ExcludeHeldAtStart {
		present = append(present, GoalFilterExcludeHeldAtStart)
	}
	return present
}
