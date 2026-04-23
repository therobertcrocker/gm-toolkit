package turn

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func NewCmd(e *engine.Engine) *cobra.Command {
	var campaignID string

	cmd := &cobra.Command{
		Use:   "turn",
		Short: "Execute a faction turn",
		RunE: func(cmd *cobra.Command, args []string) error {
			statePath := filepath.Join(".", "campaigns", campaignID, "faction_state.toml")

			factionState, err := state.Load(statePath)
			if err != nil {
				return fmt.Errorf("loading state: %w", err)
			}
			if len(factionState.Factions) == 0 {
				return fmt.Errorf("no factions found in campaign %q", campaignID)
			}

			return runTurnWizard(e, factionState, statePath)
		},
	}

	cmd.Flags().StringVar(&campaignID, "campaign", "", "campaign ID (required)")
	cmd.MarkFlagRequired("campaign")

	return cmd
}
