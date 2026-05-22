package domain

// TurnState tracks an in-progress faction turn. Persisted inside FactionState so
// a paused turn survives a restart. A nil value means no turn is in progress.
type TurnState struct {
	InProgress           bool     `toml:"in_progress"`
	CycleNumber          int      `toml:"cycle_number"`
	FactionOrder         []string `toml:"faction_order"`
	CurrentIndex         int      `toml:"current_index"`
	BookkeepingApplied   bool     `toml:"bookkeeping_applied"`
}
