package domain

// Roller produces pseudo-random die rolls. Production uses a math/rand-backed
// implementation; tests inject a deterministic fake. Roll returns a value in
// the inclusive range [1, sides].
type Roller interface {
	Roll(sides int) int
}

// Roll evaluates the DiceRoll: sum of NumDice rolls of Sides, plus Modifier.
func (diceRoll DiceRoll) Roll(roller Roller) int {
	total := diceRoll.Modifier
	for range diceRoll.NumDice {
		total += roller.Roll(diceRoll.Sides)
	}
	return total
}
