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
func runTurnWizard(e *engine.Engine, factionState *state.FactionState, statePath string) error {
	if err := resumeOrStart(e, factionState, statePath); err != nil {
		return err
	}
	if !e.Turn.InProgress(factionState) {
		return nil // GM abandoned
	}

	for {
		faction, err := e.Turn.CurrentFaction(factionState)
		if err != nil {
			return err
		}

		printFactionHeader(faction)

		// TODO: goal selection if faction.Goal == nil (bookkeeping sub-task 1)

		result, err := e.Turn.ApplyBookkeeping(factionState)
		if err != nil {
			return err
		}
		printBookkeepingResult(result)

		if err := runActionPhase(e, faction, factionState); err != nil {
			return err
		}

		done, err := e.Turn.Advance(factionState)
		if err != nil {
			return err
		}
		if err := state.Save(statePath, factionState); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		if done {
			break
		}
	}

	printCycleSummary(factionState)
	return nil
}

// runActionPhase presents available actions to the GM, collects a selection,
// and applies the resulting mutations to faction state.
func runActionPhase(e *engine.Engine, faction *domain.Faction, factionState *state.FactionState) error {
	available := e.Action.AvailableActions(faction, factionState, e.Rulebook)
	if len(available) == 0 {
		fmt.Println("  No actions available.")
		return nil
	}

	options := make([]huh.Option[engine.Action], 0, len(available)+1)
	for _, action := range available {
		options = append(options, huh.NewOption(action.Name(), action))
	}
	options = append(options, huh.NewOption[engine.Action]("No Action", nil))

	var selected engine.Action
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[engine.Action]().
				Title("Select an action").
				Options(options...).
				Value(&selected),
		),
	).Run(); err != nil {
		return fmt.Errorf("action selection cancelled: %w", err)
	}

	if selected == nil {
		return nil
	}

	mutations, err := e.Action.Run(selected, faction, factionState, e.Rulebook)
	if err != nil {
		return err
	}
	e.Mutation.Apply(factionState, mutations)
	return nil
}

// resumeOrStart handles resume detection at Cycle entry. If a Cycle is already
// in progress the GM is prompted to resume or abandon. Otherwise a new Cycle is started.
func resumeOrStart(e *engine.Engine, factionState *state.FactionState, statePath string) error {
	if e.Turn.InProgress(factionState) {
		var resume bool
		current := factionName(factionState, factionState.CurrentTurn.FactionOrder[factionState.CurrentTurn.CurrentIndex])
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Cycle %d is already in progress", factionState.CurrentTurn.CycleNumber)).
					Description(fmt.Sprintf("Currently on: %s\nResume?", current)).
					Value(&resume),
			),
		).Run(); err != nil {
			return fmt.Errorf("prompt cancelled: %w", err)
		}

		if !resume {
			e.Turn.Abandon(factionState)
			if err := state.Save(statePath, factionState); err != nil {
				return fmt.Errorf("saving state: %w", err)
			}
			fmt.Println("Cycle abandoned.")
			return nil
		}
		fmt.Printf("Resuming Cycle %d — currently on: %s\n\n", factionState.CurrentTurn.CycleNumber, current)
		return nil
	}

	if err := e.Turn.Start(factionState); err != nil {
		return fmt.Errorf("starting cycle: %w", err)
	}
	if err := state.Save(statePath, factionState); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	names := make([]string, len(factionState.CurrentTurn.FactionOrder))
	for i, id := range factionState.CurrentTurn.FactionOrder {
		names[i] = fmt.Sprintf("  %d. %s", i+1, factionName(factionState, id))
	}
	fmt.Printf("%s=== Cycle %d Begins ===%s\nFaction order:\n%s\n\n",
		bold, factionState.CurrentTurn.CycleNumber, reset,
		strings.Join(names, "\n"),
	)
	return nil
}

// factionName returns the display name for a faction ID, falling back to the ID
// if not found.
func factionName(factionState *state.FactionState, id string) string {
	for _, faction := range factionState.Factions {
		if faction.ID == id {
			return faction.Name
		}
	}
	return id
}

func printFactionHeader(faction *domain.Faction) {
	goalName := "(none)"
	if faction.Goal != nil {
		goalName = faction.Goal.Name
	}
	fmt.Printf("%s─── %s %s%s\n", bold, faction.Name, strings.Repeat("─", max(0, 40-len(faction.Name))), reset)
	fmt.Printf("  Scale: %s  HP: %d/%d  Coin: %s%d%s  Goal: %s%s%s\n",
		faction.Scale,
		faction.CurrentHP, faction.MaxHP,
		bold, faction.Coin, reset,
		dim, goalName, reset,
	)
	fmt.Println()
}

func printBookkeepingResult(result engine.BookkeepingResult) {
	fmt.Printf("  Income: %s+%d Coin%s %s(Wealth: %d, Stats: %d)%s\n",
		green, result.IncomeGained, reset,
		dim, result.WealthIncome, result.StatIncome, reset,
	)
	for _, asset := range result.AssetsUnmaintained {
		fmt.Printf("  %s! %s on %s is now unmaintained%s\n", red, asset.DefinitionID, asset.Location, reset)
	}
	for _, asset := range result.AssetsLost {
		fmt.Printf("  %sx %s on %s has been lost%s\n", red, asset.DefinitionID, asset.Location, reset)
	}
	fmt.Println()
}

func printCycleSummary(factionState *state.FactionState) {
	fmt.Printf("%s=== Cycle %d Complete ===%s\n\n", bold, factionState.CycleNumber, reset)
	for _, faction := range factionState.Factions {
		goalName := "(none)"
		if faction.Goal != nil {
			goalName = faction.Goal.Name
		}
		fmt.Printf("%-30s  HP: %d/%d  Coin: %d  Goal: %s\n",
			faction.Name, faction.CurrentHP, faction.MaxHP, faction.Coin, goalName,
		)
	}
	fmt.Println()
}
