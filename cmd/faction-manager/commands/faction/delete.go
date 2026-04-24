package faction

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/paths"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func runDeleteFactionWizard(campaignID string) error {
	p := paths.New(campaignID)
	s, err := state.Load(p.State)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	if len(s.Factions) == 0 {
		fmt.Println("No factions to delete.")
		return nil
	}

	options := make([]huh.Option[string], 0, len(s.Factions))
	for _, f := range s.Factions {
		options = append(options, huh.NewOption(f.Name, f.ID))
	}

	var targetID string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Delete Faction").
				Description("Select a faction to delete").
				Options(options...).
				Value(&targetID),
		),
	).Run(); err != nil {
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	target := s.Factions[targetID]

	summary := fmt.Sprintf(
		"Name: %s\nHomeworld: %s\nForce: %d  Cunning: %d  Wealth: %d\nHP: %d / %d",
		target.Name, target.Homeworld,
		target.Force, target.Cunning, target.Wealth,
		target.CurrentHP, target.MaxHP,
	)

	var confirmed bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete %q?", target.Name)).
				Description(summary+"\n\nThis cannot be undone.").
				Value(&confirmed),
		),
	).Run(); err != nil {
		return fmt.Errorf("wizard cancelled: %w", err)
	}

	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	delete(s.Factions, targetID)

	if err := state.Save(p.State, s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	fmt.Printf("Faction %q deleted.\n", target.Name)
	return nil
}
