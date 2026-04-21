package domain

// TurnState tracks an in-progress faction turn. Persisted inside FactionState so
// a paused turn survives a restart. A nil value means no turn is in progress.
//
// Mid-turn saves (pause/resume) write directly to faction_state.toml. The
// end-of-turn atomic commit (mutation list + history append) is a separate concern
// owned by the future Mutation and History engines — AdvanceTurn returning true is
// the seam where those engines plug in.
type TurnState struct {
	InProgress         bool     `toml:"in_progress"`
	TurnNumber         int      `toml:"turn_number"`
	FactionOrder       []string `toml:"faction_order"`
	CurrentIndex       int      `toml:"current_index"`
	BookkeepingApplied bool     `toml:"bookkeeping_applied"`
}
