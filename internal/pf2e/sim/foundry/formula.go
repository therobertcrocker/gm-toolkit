package foundry

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var errUnsupportedFormula = errors.New("unsupported formula")

func evaluateFormula(text string, variables map[string]float64) (int, error) {
	tokens, err := tokenize(text)
	if err != nil {
		return 0, err
	}
	parser := &formulaParser{tokens: tokens, variables: variables}
	value, err := parser.expression()
	if err != nil {
		return 0, fmt.Errorf("%w: %q: %v", errUnsupportedFormula, text, err)
	}
	if parser.position != len(tokens) {
		return 0, fmt.Errorf("%w: %q: unexpected %q", errUnsupportedFormula, text, parser.peek())
	}
	if value != math.Trunc(value) {
		return 0, fmt.Errorf("%w: %q gives %v, not a whole number", errUnsupportedFormula, text, value)
	}
	return int(value), nil
}

func tokenize(text string) ([]string, error) {
	var tokens []string
	characters := []rune(text)
	for index := 0; index < len(characters); {
		character := characters[index]
		start := index
		switch {
		case unicode.IsSpace(character):
			index++
			continue
		case strings.ContainsRune("+-*/(),", character):
			index++
		case unicode.IsDigit(character) || character == '.':
			for index < len(characters) && (unicode.IsDigit(characters[index]) || characters[index] == '.') {
				index++
			}
		case character == '@' || unicode.IsLetter(character):
			for index < len(characters) && isNameCharacter(characters[index]) {
				index++
			}
		default:
			return nil, fmt.Errorf("%w: %q: unexpected %q", errUnsupportedFormula, text, character)
		}
		tokens = append(tokens, string(characters[start:index]))
	}
	return tokens, nil
}

func isNameCharacter(character rune) bool {
	return character == '@' || character == '.' || character == '_' || unicode.IsLetter(character) || unicode.IsDigit(character)
}

type formulaParser struct {
	tokens    []string
	position  int
	variables map[string]float64
}

func (parser *formulaParser) peek() string {
	if parser.position < len(parser.tokens) {
		return parser.tokens[parser.position]
	}
	return ""
}

func (parser *formulaParser) next() string {
	token := parser.peek()
	parser.position++
	return token
}

func (parser *formulaParser) expression() (float64, error) {
	value, err := parser.term()
	for err == nil && (parser.peek() == "+" || parser.peek() == "-") {
		operator := parser.next()
		var right float64
		right, err = parser.term()
		if operator == "+" {
			value += right
		} else {
			value -= right
		}
	}
	return value, err
}

func (parser *formulaParser) term() (float64, error) {
	value, err := parser.factor()
	for err == nil && (parser.peek() == "*" || parser.peek() == "/") {
		operator := parser.next()
		var right float64
		right, err = parser.factor()
		switch {
		case err != nil:
		case operator == "*":
			value *= right
		case right == 0:
			err = errors.New("division by zero")
		default:
			value /= right
		}
	}
	return value, err
}

func (parser *formulaParser) factor() (float64, error) {
	token := parser.next()
	switch {
	case token == "":
		return 0, errors.New("unexpected end")
	case token == "-":
		value, err := parser.factor()
		return -value, err
	case token == "(":
		value, err := parser.expression()
		if err == nil && parser.next() != ")" {
			err = errors.New("missing )")
		}
		return value, err
	case strings.HasPrefix(token, "@"):
		value, found := parser.variables[token]
		if !found {
			return 0, fmt.Errorf("unknown variable %s", token)
		}
		return value, nil
	case unicode.IsDigit(rune(token[0])) || token[0] == '.':
		return strconv.ParseFloat(token, 64)
	default:
		return parser.call(token)
	}
}

func (parser *formulaParser) call(name string) (float64, error) {
	if parser.next() != "(" {
		return 0, fmt.Errorf("unexpected %s", name)
	}
	var arguments []float64
	for {
		value, err := parser.expression()
		if err != nil {
			return 0, err
		}
		arguments = append(arguments, value)
		separator := parser.next()
		if separator == ")" {
			break
		}
		if separator != "," {
			return 0, fmt.Errorf("unexpected %q in %s()", separator, name)
		}
	}
	switch {
	case name == "max":
		return slices.Max(arguments), nil
	case name == "min":
		return slices.Min(arguments), nil
	case name == "floor" && len(arguments) == 1:
		return math.Floor(arguments[0]), nil
	case name == "ceil" && len(arguments) == 1:
		return math.Ceil(arguments[0]), nil
	}
	return 0, fmt.Errorf("unknown function %s", name)
}
