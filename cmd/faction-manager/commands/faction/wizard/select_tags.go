package wizard

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

const planetaryGovernmentID = "T-011"

func SelectTags(tags map[string]*domain.Tag) ([]*domain.Tag, error) {
	defs := make([]*domain.Tag, 0, len(tags))
	for _, t := range tags {
		if t.ID != planetaryGovernmentID {
			defs = append(defs, t)
		}
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })

	opts := make([]huh.Option[string], len(defs))
	for i, t := range defs {
		opts[i] = huh.NewOption(fmt.Sprintf("%s — %s", t.Name, t.Description), t.ID)
	}

	var selectedID string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Tag").
				Description("Choose one tag that defines this faction's nature").
				Options(opts...).
				Value(&selectedID),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	selected := []*domain.Tag{tags[selectedID]}

	var controlsPlanet bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Planetary Government").
				Description("Does this faction control a planet?").
				Value(&controlsPlanet),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	if controlsPlanet {
		selected = append(selected, tags[planetaryGovernmentID])
	}

	return selected, nil
}
