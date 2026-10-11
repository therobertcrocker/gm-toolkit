package rulebook

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

var dicePattern = regexp.MustCompile(`^(\d+)d(\d+)([+-]\d+)?$`)

// parseDice converts a dice notation string (e.g. "2d8+2") into a DiceRoll.
// Returns nil without error for empty strings (no roll).
func parseDice(s string) (*domain.DiceRoll, error) {
	if s == "" {
		return nil, nil
	}
	matches := dicePattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid dice notation: %q", s)
	}
	numDice, _ := strconv.Atoi(matches[1])
	sides, _ := strconv.Atoi(matches[2])
	if numDice > 0 && sides == 0 {
		return nil, fmt.Errorf("invalid dice notation: %q (sides must be > 0)", s)
	}
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}
	return &domain.DiceRoll{NumDice: numDice, Sides: sides, Modifier: modifier}, nil
}

var flatAmountPattern = regexp.MustCompile(`^\d+$`)

// parseAmount converts an effect amount into a DiceRoll. A bare integer ("1") is a
// fixed amount — a zero-dice roll whose modifier is the value; anything else is dice
// notation. Returns nil without error for an empty string.
func parseAmount(s string) (*domain.DiceRoll, error) {
	if flatAmountPattern.MatchString(s) {
		value, _ := strconv.Atoi(s)
		return &domain.DiceRoll{Modifier: value}, nil
	}
	return parseDice(s)
}
