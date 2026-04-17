package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// parseDice parses dice notation (e.g. "2d6", "1d4+1", "3d10-2") into a DiceRoll.
func parseDice(s string) (domain.DiceRoll, error) {
	s = strings.TrimSpace(s)

	var modifier int
	if idx := strings.IndexAny(s, "+-"); idx != -1 {
		mod, err := strconv.Atoi(s[idx:])
		if err != nil {
			return domain.DiceRoll{}, fmt.Errorf("invalid modifier in %q", s)
		}
		modifier = mod
		s = s[:idx]
	}

	parts := strings.SplitN(s, "d", 2)
	if len(parts) != 2 {
		return domain.DiceRoll{}, fmt.Errorf("invalid dice notation %q", s)
	}

	numDice, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.DiceRoll{}, fmt.Errorf("invalid number of dice in %q", s)
	}

	sides, err := strconv.Atoi(parts[1])
	if err != nil {
		return domain.DiceRoll{}, fmt.Errorf("invalid number of sides in %q", s)
	}

	return domain.DiceRoll{NumDice: numDice, Sides: sides, Modifier: modifier}, nil
}
