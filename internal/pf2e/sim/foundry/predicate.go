package foundry

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	errUnsupportedPredicate = errors.New("unsupported predicate")
	errUnsupportedTemplate  = errors.New("unsupported template")
	templatePattern         = regexp.MustCompile(`\{(\w+)\|([^}]+)\}`)
)

type rollOptions struct {
	flags map[string]bool
}

func (options rollOptions) holds(predicate []any) (bool, error) {
	for _, term := range predicate {
		held, err := options.term(term)
		if err != nil || !held {
			return false, err
		}
	}
	return true, nil
}

func (options rollOptions) term(term any) (bool, error) {
	switch typed := term.(type) {
	case string:
		return options.flags[typed], nil
	case map[string]any:
		if len(typed) == 1 {
			for operator, operand := range typed {
				return options.compound(operator, operand)
			}
		}
	}
	return false, fmt.Errorf("%w: %v", errUnsupportedPredicate, term)
}

func (options rollOptions) compound(operator string, operand any) (bool, error) {
	switch operator {
	case "not":
		held, err := options.term(operand)
		return err == nil && !held, err
	case "and", "or", "nor":
		terms, isList := operand.([]any)
		if !isList {
			break
		}
		anyHeld := false
		for _, term := range terms {
			held, err := options.term(term)
			if err != nil {
				return false, err
			}
			if operator == "and" && !held {
				return false, nil
			}
			anyHeld = anyHeld || held
		}
		switch operator {
		case "and":
			return true, nil
		case "or":
			return anyHeld, nil
		default:
			return !anyHeld, nil
		}
	}
	return false, fmt.Errorf("%w: %s %v", errUnsupportedPredicate, operator, operand)
}

func resolveTemplates(text string, current *item) (string, error) {
	var failure error
	resolved := templatePattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := templatePattern.FindStringSubmatch(match)
		value, err := templateValue(parts[1], parts[2], current)
		if err != nil {
			failure = err
			return match
		}
		return value
	})
	return resolved, failure
}

func templateValue(scope, path string, current *item) (string, error) {
	if scope == "item" && path == "_id" {
		return current.ID, nil
	}
	if flag, found := strings.CutPrefix(path, "flags.pf2e.rulesSelections."); scope == "item" && found {
		for _, rule := range current.System.Rules {
			if rule["key"] == "ChoiceSet" && rule["flag"] == flag {
				if selection, isText := rule["selection"].(string); isText {
					return selection, nil
				}
			}
		}
	}
	return "", fmt.Errorf("%w: {%s|%s}", errUnsupportedTemplate, scope, path)
}
