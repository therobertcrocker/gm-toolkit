package engine

import "math/rand/v2"

// RandRoller is the production Roller implementation. Roll returns a value in
// the inclusive range [1, sides] using the package-level random source.
type RandRoller struct{}

func NewRandRoller() *RandRoller { return &RandRoller{} }

func (roller *RandRoller) Roll(sides int) int {
	return rand.IntN(sides) + 1
}
