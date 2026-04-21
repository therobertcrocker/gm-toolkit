package domain

// TurnPhase tracks where a faction's turn is within its lifecycle.
type TurnPhase int

const (
	PhaseBookkeeping TurnPhase = iota // income and maintenance not yet applied
	PhaseAction                       // bookkeeping done; awaiting action selection and resolution
	PhaseComplete                     // action resolved; ready for mutation/history commit and advance
)

// TurnState tracks an in-progress faction turn. Persisted inside FactionState so
// a paused turn survives a restart. A nil value means no turn is in progress.
type TurnState struct {
	InProgress   bool      `toml:"in_progress"`
	TurnNumber   int       `toml:"turn_number"`
	FactionOrder []string  `toml:"faction_order"`
	CurrentIndex int       `toml:"current_index"`
	Phase        TurnPhase `toml:"phase"`
}
