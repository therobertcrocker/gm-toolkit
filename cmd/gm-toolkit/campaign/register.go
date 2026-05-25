package campaign

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func newRegisterCmd() *cobra.Command {
	var activate bool

	cmd := &cobra.Command{
		Use:   "register <path>",
		Short: "Register an existing campaign directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("register: resolve path: %w", err)
			}
			m, err := campaigns.LoadManifest(absPath)
			if err != nil {
				return fmt.Errorf("register: %w", err)
			}
			reg, err := campaigns.LoadRegistry()
			if err != nil {
				return fmt.Errorf("register: %w", err)
			}
			if err := reg.Register(campaigns.RegistryEntry{ID: m.Campaign.ID, Path: absPath}); err != nil {
				return fmt.Errorf("register: %w", err)
			}
			if reg.Active == "" || activate {
				if err := reg.SetActive(m.Campaign.ID); err != nil {
					return fmt.Errorf("register: %w", err)
				}
			}
			if err := campaigns.SaveRegistry(reg); err != nil {
				return fmt.Errorf("register: %w", err)
			}
			fmt.Printf("Registered campaign %q (%s) at %s\n", m.Campaign.ID, m.Campaign.Name, absPath)
			if reg.Active == m.Campaign.ID {
				fmt.Printf("active campaign: %s\n", m.Campaign.ID)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&activate, "activate", false, "set as active even if another campaign is already active")
	return cmd
}
