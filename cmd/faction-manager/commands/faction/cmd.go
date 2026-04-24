package faction

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/paths"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func NewCmd(rb *loader.Rulebook) *cobra.Command {
	var campaignID string

	cmd := &cobra.Command{
		Use:   "faction",
		Short: "Manage factions",
	}

	cmd.PersistentFlags().StringVar(&campaignID, "campaign", "", "campaign ID (required)")
	cmd.MarkPersistentFlagRequired("campaign")

	cmd.AddCommand(newCreateCmd(rb, &campaignID))
	cmd.AddCommand(newListCmd(&campaignID))
	cmd.AddCommand(newDeleteCmd(&campaignID))

	return cmd
}

func newCreateCmd(rb *loader.Rulebook, campaignID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new faction",
		RunE: func(cmd *cobra.Command, args []string) error {
			faction, err := runCreateFactionWizard(rb)
			if err != nil {
				return err
			}

			p := paths.New(*campaignID)
			s, err := state.Load(p.State)
			if err != nil {
				return fmt.Errorf("loading state: %w", err)
			}

			s.CampaignID = *campaignID
			s.Factions[faction.ID] = faction

			if err := state.Save(p.State, s); err != nil {
				return fmt.Errorf("saving state: %w", err)
			}

			fmt.Printf("\nFaction %q created successfully.\n", faction.Name)
			return nil
		},
	}
}

func newListCmd(campaignID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all factions",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := paths.New(*campaignID)
			s, err := state.Load(p.State)
			if err != nil {
				return fmt.Errorf("loading state: %w", err)
			}
			if len(s.Factions) == 0 {
				fmt.Println("No factions found.")
				return nil
			}
			for _, f := range s.Factions {
				goalName := "(none)"
				if f.Goal != nil {
					goalName = f.Goal.Name
				}
				fmt.Printf("%-30s [ %s ]  HP: %d/%d  Goal: %s\n",
					f.Name, f.Scale, f.CurrentHP, f.MaxHP, goalName)
			}
			return nil
		},
	}
}

func newDeleteCmd(campaignID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a faction",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteFactionWizard(*campaignID)
		},
	}
}
