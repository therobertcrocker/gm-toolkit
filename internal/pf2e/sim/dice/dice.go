package dice

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

var ErrBadExpression = errors.New("bad dice expression")

type Roller interface {
	Die(sides int) int
}

type Seeded struct {
	source *rand.Rand
}

func NewSeeded(seed uint64) *Seeded {
	return &Seeded{source: rand.New(rand.NewPCG(seed, seed))}
}

func (roller *Seeded) Die(sides int) int {
	return roller.source.IntN(sides) + 1
}

type Expression struct {
	Count    int
	Sides    int
	Modifier int
}

func Parse(text string) (Expression, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Expression{}, fmt.Errorf("%w: empty", ErrBadExpression)
	}
	countText, rest, hasDice := strings.Cut(trimmed, "d")
	if !hasDice {
		modifier, err := strconv.Atoi(trimmed)
		if err != nil {
			return Expression{}, fmt.Errorf("%w: %q", ErrBadExpression, text)
		}
		return Expression{Modifier: modifier}, nil
	}
	sidesText, modifierText := rest, ""
	if index := strings.IndexAny(rest, "+-"); index >= 0 {
		sidesText, modifierText = rest[:index], rest[index:]
	}
	count, err := strconv.Atoi(countText)
	if err != nil || count < 1 {
		return Expression{}, fmt.Errorf("%w: %q: dice count must be a whole number of at least 1", ErrBadExpression, text)
	}
	sides, err := strconv.Atoi(sidesText)
	if err != nil || sides < 2 {
		return Expression{}, fmt.Errorf("%w: %q: die size must be a whole number of at least 2", ErrBadExpression, text)
	}
	modifier := 0
	if modifierText != "" {
		modifier, err = strconv.Atoi(modifierText)
		if err != nil {
			return Expression{}, fmt.Errorf("%w: %q: bad modifier", ErrBadExpression, text)
		}
	}
	return Expression{Count: count, Sides: sides, Modifier: modifier}, nil
}

func (expression Expression) Roll(roller Roller) int {
	total := expression.Modifier
	for range expression.Count {
		total += roller.Die(expression.Sides)
	}
	return total
}

func (expression Expression) String() string {
	if expression.Count == 0 {
		return strconv.Itoa(expression.Modifier)
	}
	text := fmt.Sprintf("%dd%d", expression.Count, expression.Sides)
	switch {
	case expression.Modifier > 0:
		text += fmt.Sprintf("+%d", expression.Modifier)
	case expression.Modifier < 0:
		text += strconv.Itoa(expression.Modifier)
	}
	return text
}
