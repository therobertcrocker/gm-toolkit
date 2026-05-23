package engine

import (
	"log/slog"
	"math/rand/v2"
)

// RandRoller is the production Roller implementation. Roll returns a value in
// the inclusive range [1, sides] using the package-level random source.
type RandRoller struct {
	log *slog.Logger
}

func NewRandRoller(log *slog.Logger) *RandRoller { return &RandRoller{log: log} }

func (roller *RandRoller) Roll(sides int) int {
	result := rand.IntN(sides) + 1
	roller.log.Debug("roll", "die", sides, "result", result)
	return result
}
