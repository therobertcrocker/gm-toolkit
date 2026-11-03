package dice

import (
	"log/slog"
	"math/rand/v2"
)

// Roller produces pseudo-random die rolls. Production uses RandRoller; tests inject
// a deterministic fake. Roll returns a value in the inclusive range [1, sides].
type Roller interface {
	Roll(sides int) int
}

// RandRoller is the production Roller.
type RandRoller struct {
	log *slog.Logger
}

func NewRandRoller(log *slog.Logger) *RandRoller { return &RandRoller{log: log} }

func (roller *RandRoller) Roll(sides int) int {
	result := rand.IntN(sides) + 1
	roller.log.Debug("roll", "die", sides, "result", result)
	return result
}
