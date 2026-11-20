package pf2e

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
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
	if adaptive := defense.Adaptive; adaptive != nil {
		fmt.Fprintf(out, "Adaptive resistance %d (%s)\n", adaptive.Value, strings.Join(adaptive.Types, ", "))
	}
	if shown.FastHealing > 0 {
		fmt.Fprintf(out, "Fast healing %d\n", shown.FastHealing)
	}
	fmt.Fprintf(out, "Speed %s\n", speedText(shown.Speed))
	if shown.ClassDC > 0 {
		fmt.Fprintf(out, "Class DC %d\n", shown.ClassDC)
	}
	if len(shown.Skills) > 0 {
		entries := make([]string, 0, len(shown.Skills))
		for _, name := range slices.Sorted(maps.Keys(shown.Skills)) {
			entries = append(entries, fmt.Sprintf("%s %+d", name, shown.Skills[name]))
		}
		fmt.Fprintf(out, "Skills %s\n", strings.Join(entries, ", "))
	}
	if shown.HeroPoints > 0 || shown.DyingMax > 0 || shown.RecoveryDC > 0 {
		fmt.Fprintf(out, "Hero points %d  dying max %d  recovery DC %d\n", shown.HeroPoints, cmp.Or(shown.DyingMax, 4), cmp.Or(shown.RecoveryDC, 10))
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
			if strike.Range > 0 {
				fmt.Fprintf(out, " range %d ft", strike.Range)
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
				steps[index] = stepText(step)
			}
			fmt.Fprintf(out, "  %s [%s]%s  %s\n", ability.Name, costText(ability), limitText(ability.Limit), strings.Join(steps, ", "))
		}
	}

	if len(shown.Pools) > 0 {
		entries := make([]string, 0, len(shown.Pools))
		for _, name := range slices.Sorted(maps.Keys(shown.Pools)) {
			entries = append(entries, fmt.Sprintf("%s %d", name, shown.Pools[name]))
		}
		fmt.Fprintf(out, "Pools %s\n", strings.Join(entries, ", "))
	}
	printFightRules(out, shown, bold, reset)
	fmt.Fprintf(out, "Layers %s\n", strings.Join(shown.Layers, " + "))
}

func printFightRules(out io.Writer, shown *creature.Creature, bold, reset string) {
	if len(shown.Modifiers) > 0 {
		fmt.Fprintf(out, "%sModifiers%s\n", bold, reset)
		for _, id := range slices.Sorted(maps.Keys(shown.Modifiers)) {
			modifier := shown.Modifiers[id]
			line := fmt.Sprintf("  %s  %+d", id, modifier.Value)
			if modifier.Type != "" {
				line += " " + modifier.Type
			}
			line += " " + modifier.Selector
			if len(modifier.Against) > 0 {
				line += " against " + strings.Join(modifier.Against, ", ")
			}
			if modifier.PerCondition != "" {
				line += " per " + modifier.PerCondition
			}
			fmt.Fprintln(out, line)
		}
	}
	if len(shown.DegreeShifts) > 0 {
		fmt.Fprintf(out, "%sDegree shifts%s\n", bold, reset)
		for _, id := range shown.DegreeShiftIDs() {
			shift := shown.DegreeShifts[id]
			line := fmt.Sprintf("  %s  %s %s -> %s", id, shift.Selector, shift.From, shift.To)
			if len(shift.Against) > 0 {
				line += " against " + strings.Join(shift.Against, ", ")
			}
			fmt.Fprintln(out, line)
		}
	}
	if len(shown.ConditionAdjustments) > 0 {
		fmt.Fprintf(out, "%sCondition adjustments%s\n", bold, reset)
		for _, name := range slices.Sorted(maps.Keys(shown.ConditionAdjustments)) {
			adjustment := shown.ConditionAdjustments[name]
			fmt.Fprintf(out, "  %s  gain %+d  decay %d\n", name, adjustment.Gain, adjustment.Decay)
		}
	}
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
		if part.Category != "" {
			texts[index] += " " + part.Category
		}
	}
	return strings.Join(texts, " plus ")
}

func stepText(step creature.Step) string {
	if step.Kind() == creature.StepStrike {
		return "strike " + step.Strike
	}
	return step.Kind()
}

func costText(ability creature.Ability) string {
	low, high := ability.CostRange()
	switch {
	case ability.Trigger != nil:
		return triggerText(*ability.Trigger)
	case ability.Free:
		return "free"
	case high > low:
		return fmt.Sprintf("%d-%d", low, high)
	default:
		return strconv.Itoa(low)
	}
}

func triggerText(trigger creature.Trigger) string {
	cost := trigger.Cost
	if cost == "" {
		cost = "passive"
	}
	return cost + " on " + trigger.On
}

func limitText(limit *creature.Limit) string {
	if limit == nil {
		return ""
	}
	var parts []string
	if limit.PerFight > 0 {
		parts = append(parts, fmt.Sprintf("%d per fight", limit.PerFight))
	}
	if limit.PerRound > 0 {
		parts = append(parts, fmt.Sprintf("%d per round", limit.PerRound))
	}
	if limit.Recharge != "" {
		parts = append(parts, "recharge "+limit.Recharge)
	}
	if limit.Pool != "" {
		parts = append(parts, "pool "+limit.Pool)
	}
	return " (" + strings.Join(parts, ", ") + ")"
}
