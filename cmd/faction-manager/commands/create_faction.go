package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type factionScale string

const (
	scaleMajor   factionScale = "major"
	scaleHegemon factionScale = "hegemon"
)

func runCreateFactionWizard() (*domain.Faction, error) {
	var (
		name      string
		homeworld string
		scale     string
		primary   string
		secondary string
		confirmed bool
	)

	// Step 1: Basic info and scale
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Faction Name").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}).
				Value(&name),
			huh.NewInput().
				Title("Homeworld").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("homeworld cannot be empty")
					}
					return nil
				}).
				Value(&homeworld),
			huh.NewSelect[string]().
				Title("Scale").
				Description("Determines starting attribute ratings").
				Options(
					huh.NewOption("Minor  (primary 4 / secondary 3 / tertiary 1)", "minor"),
					huh.NewOption("Major  (primary 6 / secondary 5 / tertiary 3)", "major"),
					huh.NewOption("Hegemon  (primary 8 / secondary 7 / tertiary 5)", "hegemon"),
				).
				Value(&scale),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 2: Primary attribute
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Primary Attribute").
				Description("This stat receives the highest rating").
				Options(
					huh.NewOption("Force", "Force"),
					huh.NewOption("Cunning", "Cunning"),
					huh.NewOption("Wealth", "Wealth"),
				).
				Value(&primary),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 3: Secondary attribute (filtered)
	secondaryOptions := make([]huh.Option[string], 0, 2)
	for _, stat := range []string{"Force", "Cunning", "Wealth"} {
		if stat != primary {
			secondaryOptions = append(secondaryOptions, huh.NewOption(stat, stat))
		}
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Secondary Attribute").
				Description("This stat receives the middle rating").
				Options(secondaryOptions...).
				Value(&secondary),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Derive tertiary
	tertiary := ""
	for _, stat := range []string{"Force", "Cunning", "Wealth"} {
		if stat != primary && stat != secondary {
			tertiary = stat
			break
		}
	}

	// Calculate ratings
	primaryRating, secondaryRating, tertiaryRating := ratingsFromScale(factionScale(scale))
	ratings := map[string]int{
		primary:   primaryRating,
		secondary: secondaryRating,
		tertiary:  tertiaryRating,
	}

	// Select starting assets

	faction := &domain.Faction{
		ID:        slugify(name),
		Name:      name,
		Scale:     domain.ScaleFromString(scale),
		Homeworld: homeworld,
		Force:     ratings["Force"],
		Cunning:   ratings["Cunning"],
		Wealth:    ratings["Wealth"],
	}
	faction.MaxHP = calcMaxHP(faction)
	faction.CurrentHP = faction.MaxHP

	// Step 4: Confirmation
	summary := fmt.Sprintf(
		"Name: %s\nHomeworld: %s\nForce: %d  Cunning: %d  Wealth: %d\nMax HP: %d",
		faction.Name, faction.Homeworld,
		faction.Force, faction.Cunning, faction.Wealth,
		faction.MaxHP,
	)

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm Faction").
				Description(summary).
				Value(&confirmed),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	if !confirmed {
		return nil, fmt.Errorf("faction creation cancelled")
	}

	return faction, nil
}

func ratingsFromScale(scale factionScale) (primary, secondary, tertiary int) {
	switch scale {
	case scaleMajor:
		return 6, 5, 3
	case scaleHegemon:
		return 8, 7, 5
	default: // minor
		return 4, 3, 1
	}
}

func calcMaxHP(f *domain.Faction) int {
	return 4 + hpValueForRating(f.Force) + hpValueForRating(f.Cunning) + hpValueForRating(f.Wealth)
}

// hpValueForRating returns the HP contribution for a given rating per SWN rules.
func hpValueForRating(rating int) int {
	values := map[int]int{1: 1, 2: 2, 3: 4, 4: 6, 5: 9, 6: 12, 7: 16, 8: 20}
	if v, ok := values[rating]; ok {
		return v
	}
	return 0
}

// slugify creates a simple ID from a faction name.
func slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}
