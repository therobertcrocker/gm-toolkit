package campaign

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func newSetActiveCmd() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "set-active",
		Short: "Set the active campaign",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := campaigns.LoadRegistry()
			if err != nil {
				return fmt.Errorf("set-active: %w", err)
			}
			if len(reg.Campaigns) == 0 {
				return fmt.Errorf("set-active: no campaigns registered\n  hint: gm-toolkit campaign create <id> --path <dir>")
			}

			id := target
			if id == "" {
				type displayEntry struct{ id, label string }
				displays := make([]displayEntry, 0, len(reg.Campaigns))
				for _, entry := range reg.Campaigns {
					m, mErr := campaigns.LoadManifest(entry.Path)
					name := entry.ID
					if mErr == nil {
						name = m.Campaign.Name
					}
					displays = append(displays, displayEntry{entry.ID, fmt.Sprintf("%s  —  %s", entry.ID, name)})
				}
				sort.Slice(displays, func(i, j int) bool { return displays[i].label < displays[j].label })
				opts := make([]huh.Option[string], 0, len(displays))
				for _, d := range displays {
					opts = append(opts, huh.NewOption(d.label, d.id))
				}
				if reg.Active != "" {
					id = reg.Active
				}
				sel := huh.NewSelect[string]().Title("Select active campaign").Options(opts...).Value(&id)
				if err := huh.NewForm(huh.NewGroup(sel)).Run(); err != nil {
					return fmt.Errorf("set-active: %w", err)
				}
			}

			if err := reg.SetActive(id); err != nil {
				return fmt.Errorf("set-active: %w", err)
			}
			if err := campaigns.SaveRegistry(reg); err != nil {
				return fmt.Errorf("set-active: %w", err)
			}
			fmt.Printf("active campaign: %s\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "campaign", "", "campaign id (non-interactive)")
	return cmd
}
