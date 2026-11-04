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

// Table rolls dice that pick a row: the first whose range holds the total. No row
// means no result; the absence of a row is a zero payout.
func Table(table domain.RollTable, roller Roller) (domain.AbilityOutcome, bool) {
	total := Amount(table.Die, roller)
	for _, outcome := range table.Outcomes {
		if total >= outcome.From && total <= outcome.To {
			return outcome, true
		}
	}
	return domain.AbilityOutcome{}, false
}

// StepResult is what an ability step's die gave, by the step's type: Amount for an
// amount roll, Row for a table roll with a matching row.
type StepResult struct {
	Amount int
	Row    *domain.AbilityOutcome
}

// RollStep rolls the step's die. A step has at most one (parser-enforced), and one
// with none rolls nothing.
func RollStep(step domain.AbilityStep, roller Roller) StepResult {
	switch {
	case step.Amount != nil:
		return StepResult{Amount: Amount(*step.Amount, roller)}
	case step.Table != nil:
		if row, matched := Table(*step.Table, roller); matched {
			return StepResult{Row: &row}
		}
	}
	return StepResult{}
}
