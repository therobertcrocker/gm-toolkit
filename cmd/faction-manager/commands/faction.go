package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var campaignID string

var factionCmd = &cobra.Command{
	Use:   "faction",
	Short: "Manage factions",
}

var factionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new faction",
	RunE: func(cmd *cobra.Command, args []string) error {
		faction, err := runCreateFactionWizard()
		if err != nil {
			return err
		}

		statePath := filepath.Join(".", "campaigns", campaignID, "faction_state.toml")

		s, err := state.Load(statePath)
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}

		s.CampaignID = campaignID
		s.Factions = append(s.Factions, faction)

		if err := state.Save(statePath, s); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		fmt.Printf("\nFaction %q created successfully.\n", faction.Name)
		return nil
	},
}

var factionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all factions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("List factions - not yet implemented")
		return nil
	},
}

var factionDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a faction",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeleteFactionWizard(campaignID)
	},
}

func init() {
	factionCmd.PersistentFlags().StringVar(&campaignID, "campaign", "", "campaign ID (required)")
	factionCmd.MarkPersistentFlagRequired("campaign")

	factionCmd.AddCommand(factionCreateCmd)
	factionCmd.AddCommand(factionListCmd)
	factionCmd.AddCommand(factionDeleteCmd)
}
