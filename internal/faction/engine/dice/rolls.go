package dice

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Amount rolls dice whose total is the number. No hook reaches it.
func Amount(roll domain.DiceRoll, roller Roller) int {
	total := roll.Modifier
	for range roll.NumDice {
		total += roller.Roll(roll.Sides)
	}
	return total
}
