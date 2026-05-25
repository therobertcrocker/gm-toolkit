package campaign

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
)

func newCreateCmd() *cobra.Command {
	var (
		name     string
		path     string
		rules    string
		activate bool
	)

	cmd := &cobra.Command{
		Use:   "create [id]",
		Short: "Create a new campaign",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := campaigns.LoadRegistry()
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}
			defaultParent := os.Getenv("GM_TOOLKIT_HOME")

			var id string
			if len(args) == 0 && path == "" && name == "" && rules == "" {
				id, name, path, rules, err = createWizard(defaultParent)
				if err != nil {
					return err
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("create: id is required in non-interactive mode\n  hint: gm-toolkit campaign create <id>\n  or:   gm-toolkit campaign create   # interactive wizard")
				}
				id = args[0]
				if err := campaigns.ValidateID(id); err != nil {
					return fmt.Errorf("create: %w", err)
				}
				if path == "" {
					path = defaultParent
				}
				if path == "" {
					return fmt.Errorf("create: --path is required when GM_TOOLKIT_HOME is not set\n  hint: gm-toolkit campaign create <id> --path <parent-dir>")
				}
				if name == "" {
					name = id
				}
			}

			parentAbs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("create: resolve path: %w", err)
			}
			absPath := filepath.Join(parentAbs, id)

			if _, exists := reg.Lookup(id); exists {
				regPath, _ := campaigns.RegistryPath()
				return fmt.Errorf("create: id %q is already registered; choose a different id or remove the entry from %s", id, regPath)
			}

			if _, statErr := os.Stat(absPath); statErr == nil {
				return fmt.Errorf("create: target %s already exists; choose a different id or remove the directory first", absPath)
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return fmt.Errorf("create: check target %s: %w", absPath, statErr)
			}

			camp, err := campaigns.Scaffold(absPath, id, name)
			if err != nil {
				return fmt.Errorf("create: %w", err)
			}

			if err := reg.Register(campaigns.RegistryEntry{ID: id, Path: absPath}); err != nil {
				return fmt.Errorf("create: %w", err)
			}
			if reg.Active == "" || activate {
				if err := reg.SetActive(id); err != nil {
					return fmt.Errorf("create: %w", err)
				}
			}
			if err := campaigns.SaveRegistry(reg); err != nil {
				return fmt.Errorf("create: %w\n  hint: campaign directory was created at %s; remove it manually before retrying", err, absPath)
			}

			fmt.Printf("Created campaign %q at %s\n", id, absPath)
			if reg.Active == id {
				fmt.Printf("active campaign: %s\n", id)
			}

			if rules != "" {
				copied, err := campaigns.CopyRulebook(camp, rules, false)
				if err != nil {
					return fmt.Errorf("create: seed rulebook: %w", err)
				}
				fmt.Printf("Copied %d rulebook file(s) into %s/rulebook/\n", copied, absPath)
			} else {
				fmt.Printf("rulebook/ is empty — seed it with 'gm-toolkit campaign add-rules --rules <path>' or copy TOML manually before running the toolkit\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name (defaults to id)")
	cmd.Flags().StringVar(&path, "path", "", "parent directory; campaign created at <path>/<id> (defaults to $GM_TOOLKIT_HOME)")
	cmd.Flags().StringVar(&rules, "rules", "", "directory of rulebook TOML to seed into rulebook/")
	cmd.Flags().BoolVar(&activate, "activate", false, "set as active even if another campaign is already active")
	return cmd
}

func createWizard(defaultParent string) (id, name, path, rules string, err error) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Campaign id (kebab-case)").Value(&id).Validate(campaigns.ValidateID),
			huh.NewInput().Title("Display name (default = id)").Value(&name),
			huh.NewInput().Title("Parent directory for campaign").Value(&path).Suggestions([]string{defaultParent}),
			huh.NewInput().Title("Rulebook source dir (optional)").Value(&rules),
		),
	)
	if err = form.Run(); err != nil {
		return
	}
	if name == "" {
		name = id
	}
	if path == "" {
		path = defaultParent
	}
	return
}
