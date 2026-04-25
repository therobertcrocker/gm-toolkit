package faction

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/commands/faction/wizard"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

func runCreateFactionWizard(rb *loader.Rulebook) (*domain.Faction, error) {
	var (
		name      string
		homeworld string
		scale     domain.FactionScale
		primary   domain.FactionStat
		secondary domain.FactionStat
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
			huh.NewSelect[domain.FactionScale]().
				Title("Scale").
				Description("Determines starting attribute ratings").
				Options(
					huh.NewOption("Minor  (primary 4 / secondary 3 / tertiary 1)", domain.ScaleMinor),
					huh.NewOption("Major  (primary 6 / secondary 5 / tertiary 3)", domain.ScaleMajor),
					huh.NewOption("Hegemon  (primary 8 / secondary 7 / tertiary 5)", domain.ScaleHegemon),
				).
				Value(&scale),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 2: Primary attribute
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[domain.FactionStat]().
				Title("Primary Attribute").
				Description("This stat receives the highest rating").
				Options(
					huh.NewOption("Force", domain.StatForce),
					huh.NewOption("Cunning", domain.StatCunning),
					huh.NewOption("Wealth", domain.StatWealth),
				).
				Value(&primary),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Step 3: Secondary attribute (filtered)
	secondaryOptions := make([]huh.Option[domain.FactionStat], 0, 2)
	for _, stat := range []domain.FactionStat{domain.StatForce, domain.StatCunning, domain.StatWealth} {
		if stat != primary {
			secondaryOptions = append(secondaryOptions, huh.NewOption(string(stat), stat))
		}
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[domain.FactionStat]().
				Title("Secondary Attribute").
				Description("This stat receives the middle rating").
				Options(secondaryOptions...).
				Value(&secondary),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Derive tertiary
	var tertiary domain.FactionStat
	for _, stat := range []domain.FactionStat{domain.StatForce, domain.StatCunning, domain.StatWealth} {
		if stat != primary && stat != secondary {
			tertiary = stat
			break
		}
	}

	// Calculate ratings
	primaryRating, secondaryRating, tertiaryRating := domain.RatingsFromScale(scale)
	ratings := map[domain.FactionStat]int{
		primary:   primaryRating,
		secondary: secondaryRating,
		tertiary:  tertiaryRating,
	}

	faction := &domain.Faction{
		ID:        slugify(name),
		Name:      name,
		Scale:     scale,
		Homeworld: homeworld,
		Force:     ratings[domain.StatForce],
		Cunning:   ratings[domain.StatCunning],
		Wealth:    ratings[domain.StatWealth],
	}
	faction.MaxHP = domain.CalcMaxHP(faction)
	faction.CurrentHP = faction.MaxHP
	faction.Bases = []*domain.Base{
		{
			ID:          fmt.Sprintf("%s-homeworld", faction.ID),
			OwnerID:     faction.ID,
			Location:    homeworld,
			CurrentHP:   faction.MaxHP,
			MaxHP:       faction.MaxHP,
			Ready:       true,
			IsHomeworld: true,
		},
	}

	// Step 4: Starting asset selection
	otherStats := []domain.FactionStat{}
	for _, stat := range []domain.FactionStat{domain.StatForce, domain.StatCunning, domain.StatWealth} {
		if stat != primary {
			otherStats = append(otherStats, stat)
		}
	}

	assets, err := wizard.SelectStartingAssets(faction.ID, homeworld, primary, otherStats, ratings, scale, rb.Assets)
	if err != nil {
		return nil, err
	}
	faction.Assets = assets

	fmt.Printf("\nA Base of Influence has been placed on %s at maximum HP (%d).\n", homeworld, faction.MaxHP)

	// Step 5: Tag selection
	tags, err := wizard.SelectTags(rb.Tags)
	if err != nil {
		return nil, err
	}
	faction.Tags = tags

	// Step 6: Goal selection
	goal, err := wizard.SelectGoal(rb.Goals)
	if err != nil {
		return nil, err
	}
	faction.Goal = goal

	// Step 7: Confirmation
	assetNames := make([]string, len(assets))
	for i, asset := range assets {
		def := rb.Assets[asset.DefinitionID]
		assetNames[i] = fmt.Sprintf("  • %s (%s)", def.Name, def.Category)
	}
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = fmt.Sprintf("  • %s", t.Name)
	}
	summary := fmt.Sprintf(
		"Name: %s\nHomeworld: %s\nForce: %d  Cunning: %d  Wealth: %d\nMax HP: %d\n\nTags:\n%s\n\nGoal: %s\n\nStarting Assets:\n%s",
		faction.Name, faction.Homeworld,
		faction.Force, faction.Cunning, faction.Wealth,
		faction.MaxHP,
		strings.Join(tagNames, "\n"),
		faction.Goal.Name,
		strings.Join(assetNames, "\n"),
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

func slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}
