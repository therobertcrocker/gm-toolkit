package loader

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
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}
	return &domain.DiceRoll{NumDice: numDice, Sides: sides, Modifier: modifier}, nil
}
