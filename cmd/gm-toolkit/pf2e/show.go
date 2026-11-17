package pf2e

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const (
	ansiBold  = "\033[1m"
	ansiReset = "\033[0m"
)

func newShowCmd(campaignOverride *string) *cobra.Command {
	var variants []string
	cmd := &cobra.Command{
		Use:   "show <creature>",
		Short: "Print a creature's merged stat block",
		Long:  "Merges the creature's layers (import.toml, hand.toml, then each --variant in order) and prints the result. Looks in pf2e/party/ and then in each encounters/<fight>/ folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			campaign, err := resolveCampaign(*campaignOverride)
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			paths := campaign.Paths()
			directory, err := creature.Locate(paths.PartyDir, paths.EncountersDir, args[0])
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			loaded, err := creature.Load(directory, variants...)
			if err != nil {
				return fmt.Errorf("pf2e sim show: %w", err)
			}
			out := cmd.OutOrStdout()
			printCreature(out, loaded, isTerminal(out))
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&variants, "variant", nil, "variant layer to apply on top, by name; repeat to stack, applied in order")
	return cmd
}

func isTerminal(out io.Writer) bool {
	file, isFile := out.(*os.File)
	return isFile && os.Getenv("NO_COLOR") == "" && isatty.IsTerminal(file.Fd())
}

func printCreature(out io.Writer, shown *creature.Creature, styled bool) {
	bold, reset := "", ""
	if styled {
		bold, reset = ansiBold, ansiReset
	}
	fmt.Fprintf(out, "%s%s%s  %s %d", bold, shown.Name, reset, shown.Kind, shown.Level)
	if len(shown.Traits) > 0 {
		fmt.Fprintf(out, "  [%s]", strings.Join(shown.Traits, ", "))
	}
	fmt.Fprintln(out)

	defense := shown.Defense
	fmt.Fprintf(out, "Perception %+d\n", shown.Perception)
	if len(shown.Attributes) > 0 {
		texts := make([]string, len(creature.AttributeNames))
		for index, name := range creature.AttributeNames {
			texts[index] = fmt.Sprintf("%s %+d", strings.ToUpper(name[:1])+name[1:], shown.Attributes[name])
		}
		fmt.Fprintln(out, strings.Join(texts, "  "))
	}
	fmt.Fprintf(out, "HP %d  AC %d  Fort %+d  Ref %+d  Will %+d\n", defense.HP, defense.AC, defense.Fortitude, defense.Reflex, defense.Will)
	if len(defense.Immune) > 0 {
		fmt.Fprintf(out, "Immune %s\n", strings.Join(defense.Immune, ", "))
	}
	printAdjustments(out, "Resist", defense.Resist)
	printAdjustments(out, "Weakness", defense.Weakness)
	fmt.Fprintf(out, "Speed %s\n", speedText(shown.Speed))
	if shown.ClassDC > 0 {
		fmt.Fprintf(out, "Class DC %d\n", shown.ClassDC)
	}
	if casting := shown.Spellcasting; casting != nil {
		fmt.Fprintf(out, "Spell attack %+d  DC %d", casting.Attack, casting.DC)
		if casting.TopRank > 0 {
			fmt.Fprintf(out, "  top rank %d", casting.TopRank)
		}
		fmt.Fprintln(out)
	}

	if len(shown.Strikes) > 0 {
		fmt.Fprintf(out, "%sStrikes%s\n", bold, reset)
		for _, id := range shown.StrikeIDs() {
			strike := shown.Strikes[id]
			fmt.Fprintf(out, "  %s %+d", strike.Name, strike.Bonus)
			if len(strike.Traits) > 0 {
				fmt.Fprintf(out, " (%s)", strings.Join(strike.Traits, ", "))
			}
			fmt.Fprintf(out, "  %s\n", damageText(strike.Damage))
		}
	}

	if len(shown.Abilities) > 0 {
		fmt.Fprintf(out, "%sAbilities%s\n", bold, reset)
		for _, id := range shown.AbilityIDs() {
			ability := shown.Abilities[id]
			if ability.Ignore {
				fmt.Fprintf(out, "  %s (ignored)\n", id)
				continue
			}
			steps := make([]string, len(ability.Steps))
			for index, step := range ability.Steps {
				steps[index] = "strike " + step.Strike
			}
			fmt.Fprintf(out, "  %s [%d]  %s\n", ability.Name, ability.Actions, strings.Join(steps, ", "))
		}
	}

	fmt.Fprintf(out, "Layers %s\n", strings.Join(shown.Layers, " + "))
}

func printAdjustments(out io.Writer, label string, adjustments map[string]creature.Adjustment) {
	if len(adjustments) == 0 {
		return
	}
	entries := make([]string, 0, len(adjustments))
	for _, key := range slices.Sorted(maps.Keys(adjustments)) {
		adjustment := adjustments[key]
		entry := fmt.Sprintf("%s %d", key, adjustment.Value)
		if len(adjustment.Except) > 0 {
			entry += fmt.Sprintf(" (except %s)", strings.Join(adjustment.Except, ", "))
		}
		entries = append(entries, entry)
	}
	fmt.Fprintf(out, "%s %s\n", label, strings.Join(entries, ", "))
}

func speedText(speed creature.Speed) string {
	text := fmt.Sprintf("%d ft", speed.Land)
	if speed.Fly > 0 {
		text += fmt.Sprintf(", fly %d ft", speed.Fly)
	}
	return text
}

func damageText(parts []creature.DamagePart) string {
	texts := make([]string, len(parts))
	for index, part := range parts {
		texts[index] = part.Roll.String() + " " + part.Type
	}
	return strings.Join(texts, " plus ")
}
