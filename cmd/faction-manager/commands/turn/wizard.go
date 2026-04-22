package turn

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	bold  = "\033[1m"
	reset = "\033[0m"
	green = "\033[32m"
	red   = "\033[31m"
	dim   = "\033[2m"
)

// runTurnWizard drives the interactive Cycle wizard: resume detection,
// per-faction Turn loop, and Cycle completion summary.
func runTurnWizard(te *engine.TurnEngine, s *state.FactionState, statePath string) error {
	if err := resumeOrStart(te, s, statePath); err != nil {
		return err
	}
	if !te.InProgress(s) {
		return nil // GM abandoned
	}

	for {
		f, err := te.CurrentFaction(s)
		if err != nil {
			return err
		}

		printFactionHeader(f)

		// TODO: goal selection if f.Goal == nil (bookkeeping sub-task 1)

		result, err := te.ApplyBookkeeping(s)
		if err != nil {
			return err
		}
		printBookkeepingResult(result)

		printActionPlaceholder()

		done, err := te.Advance(s)
		if err != nil {
			return err
		}
		if err := state.Save(statePath, s); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		if done {
			break
		}
	}

	printCycleSummary(s)
	return nil
}

// resumeOrStart handles resume detection at Cycle entry. If a Cycle is already
// in progress the GM is prompted to resume or abandon. Otherwise a new Cycle is started.
func resumeOrStart(te *engine.TurnEngine, s *state.FactionState, statePath string) error {
	if te.InProgress(s) {
		var resume bool
		current := factionName(s, s.CurrentTurn.FactionOrder[s.CurrentTurn.CurrentIndex])
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Cycle %d is already in progress", s.CurrentTurn.CycleNumber)).
					Description(fmt.Sprintf("Currently on: %s\nResume?", current)).
					Value(&resume),
			),
		).Run(); err != nil {
			return fmt.Errorf("prompt cancelled: %w", err)
		}

		if !resume {
			te.Abandon(s)
			if err := state.Save(statePath, s); err != nil {
				return fmt.Errorf("saving state: %w", err)
			}
			fmt.Println("Cycle abandoned.")
			return nil
		}
		fmt.Printf("Resuming Cycle %d — currently on: %s\n\n", s.CurrentTurn.CycleNumber, current)
		return nil
	}

	if err := te.Start(s); err != nil {
		return fmt.Errorf("starting cycle: %w", err)
	}
	if err := state.Save(statePath, s); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	names := make([]string, len(s.CurrentTurn.FactionOrder))
	for i, id := range s.CurrentTurn.FactionOrder {
		names[i] = fmt.Sprintf("  %d. %s", i+1, factionName(s, id))
	}
	fmt.Printf("%s=== Cycle %d Begins ===%s\nFaction order:\n%s\n\n",
		bold, s.CurrentTurn.CycleNumber, reset,
		strings.Join(names, "\n"),
	)
	return nil
}

// factionName returns the display name for a faction ID, falling back to the ID
// if not found.
func factionName(s *state.FactionState, id string) string {
	for _, f := range s.Factions {
		if f.ID == id {
			return f.Name
		}
	}
	return id
}

func printFactionHeader(f *domain.Faction) {
	goalName := "(none)"
	if f.Goal != nil {
		goalName = f.Goal.Name
	}
	fmt.Printf("%s─── %s %s%s\n", bold, f.Name, strings.Repeat("─", max(0, 40-len(f.Name))), reset)
	fmt.Printf("  Scale: %s  HP: %d/%d  Coin: %s%d%s  Goal: %s%s%s\n",
		f.Scale,
		f.CurrentHP, f.MaxHP,
		bold, f.Coin, reset,
		dim, goalName, reset,
	)
	fmt.Println()
}

func printBookkeepingResult(r engine.BookkeepingResult) {
	fmt.Printf("  Income: %s+%d Coin%s %s(Wealth: %d, Stats: %d)%s\n",
		green, r.IncomeGained, reset,
		dim, r.WealthIncome, r.StatIncome, reset,
	)
	for _, a := range r.AssetsUnmaintained {
		fmt.Printf("  %s! %s on %s is now unmaintained%s\n", red, a.DefinitionID, a.Location, reset)
	}
	for _, a := range r.AssetsLost {
		fmt.Printf("  %sx %s on %s has been lost%s\n", red, a.DefinitionID, a.Location, reset)
	}
	fmt.Println()
}

func printCycleSummary(s *state.FactionState) {
	fmt.Printf("%s=== Cycle %d Complete ===%s\n\n", bold, s.CycleNumber, reset)
	for _, f := range s.Factions {
		goalName := "(none)"
		if f.Goal != nil {
			goalName = f.Goal.Name
		}
		fmt.Printf("%-30s  HP: %d/%d  Coin: %d  Goal: %s\n",
			f.Name, f.CurrentHP, f.MaxHP, f.Coin, goalName,
		)
	}
	fmt.Println()
}

func printActionPlaceholder() {
	fmt.Println("[ Action phase — not yet implemented ]")
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
	fmt.Println()
}
