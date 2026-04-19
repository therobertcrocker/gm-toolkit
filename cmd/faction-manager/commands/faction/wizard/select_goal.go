package wizard

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func SelectGoal(goals map[string]*domain.Goal) (*domain.Goal, error) {
	defs := make([]*domain.Goal, 0, len(goals))
	for _, g := range goals {
		defs = append(defs, g)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })

	opts := make([]huh.Option[string], len(defs))
	for i, g := range defs {
		opts[i] = huh.NewOption(
			fmt.Sprintf("%s (%s) — %s", g.Name, FormatDifficulty(g.Difficulty), g.Description),
			g.ID,
		)
	}

	var selectedID string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Starting Goal").
				Options(opts...).
				Value(&selectedID),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	return goals[selectedID], nil
}

func FormatDifficulty(d string) string {
	switch d {
	case "half_assets_destroyed":
		return "half assets destroyed"
	case "half_avg_ruling_faction":
		return "half avg of ruling faction's stats"
	case "low_plus_contested":
		return "low (+1 if contested)"
	case "one_plus_avg_target":
		return "1 + avg of target faction's stats"
	default:
		return d
	}
}
