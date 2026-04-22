package turn

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func NewCmd(te *engine.TurnEngine) *cobra.Command {
	var campaignID string

	cmd := &cobra.Command{
		Use:   "turn",
		Short: "Execute a faction turn",
		RunE: func(cmd *cobra.Command, args []string) error {
			statePath := filepath.Join(".", "campaigns", campaignID, "faction_state.toml")

			s, err := state.Load(statePath)
			if err != nil {
				return fmt.Errorf("loading state: %w", err)
			}
			if len(s.Factions) == 0 {
				return fmt.Errorf("no factions found in campaign %q", campaignID)
			}

			return runTurnWizard(te, s, statePath)
		},
	}

	cmd.Flags().StringVar(&campaignID, "campaign", "", "campaign ID (required)")
	cmd.MarkFlagRequired("campaign")

	return cmd
}
