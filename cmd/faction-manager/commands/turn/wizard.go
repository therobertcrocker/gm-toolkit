package turn

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/therobertcrocker/gm-toolkit/cmd/faction-manager/paths"
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
func runTurnWizard(e *engine.Engine, factionState *state.FactionState, p paths.Paths) error {
	if err := resumeOrStart(e, factionState, p); err != nil {
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

		skipped, err := promptSkipTurn(faction)
		if err != nil {
			return err
		}
		if skipped {
			done, err := e.Turn.Advance(factionState)
			if err != nil {
				return err
			}
			if err := state.Save(p.State, factionState); err != nil {
				return fmt.Errorf("saving state: %w", err)
			}
			if done {
				break
			}
			continue
		}

		// TODO: goal selection if faction.Goal == nil (bookkeeping sub-task 1)

		result, err := e.Turn.ApplyBookkeeping(factionState)
		if err != nil {
			return err
		}
		printBookkeepingResult(result)
		pressEnterToContinue()

		actionMutations, err := runActionPhase(e, faction, factionState)
		if err != nil {
			return err
		}

		done, err := e.Turn.Advance(factionState)
		if err != nil {
			return err
		}

		event, err := buildEventRecord(factionState, faction, append(result.RecordedMutations, actionMutations...))
		if err != nil {
			return fmt.Errorf("building event record: %w", err)
		}

		if err := state.Save(p.State, factionState); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		if err := e.History.Record(p.History, event); err != nil {
			return fmt.Errorf("recording history: %w", err)
		}

		if done {
			break
		}
	}

	printCycleSummary(factionState)
	return nil
}

// runActionPhase presents available actions to the GM, collects a selection,
// applies the resulting mutations, and returns them for history recording.
// Index-based selection is used so that huh compares plain integers rather
// than interface values, which avoids edge cases in equality checks.
func runActionPhase(e *engine.Engine, faction *domain.Faction, factionState *state.FactionState) ([]domain.Mutation, error) {
	available := e.Action.AvailableActions(faction, factionState, e.Rulebook)
	if len(available) == 0 {
		fmt.Println("  No actions available.")
		return nil, nil
	}

	// Available actions occupy indices 0..n-1; -1 is the No Action sentinel.
	options := make([]huh.Option[int], 0, len(available)+1)
	for i, action := range available {
		options = append(options, huh.NewOption(action.Name(), i))
	}
	options = append(options, huh.NewOption("No Action", -1))

	selectedIdx := 0
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select an action").
				Options(options...).
				Value(&selectedIdx),
		),
	).Run(); err != nil {
		return nil, fmt.Errorf("action selection cancelled: %w", err)
	}

	if selectedIdx == -1 {
		return nil, nil
	}

	mutations, err := e.Action.Run(available[selectedIdx], faction, factionState, e.Rulebook)
	if err != nil {
		return nil, err
	}
	e.Mutation.Apply(factionState, mutations)
	return mutations, nil
}

// buildEventRecord constructs an EventRecord for a faction's completed turn,
// converting all mutations to their serializable record form.
func buildEventRecord(factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) (domain.EventRecord, error) {
	records := make([]domain.MutationRecord, 0, len(mutations))
	for _, mutation := range mutations {
		record, err := domain.NewMutationRecord(mutation)
		if err != nil {
			return domain.EventRecord{}, fmt.Errorf("building mutation record: %w", err)
		}
		records = append(records, record)
	}
	return domain.EventRecord{
		Cycle:     factionState.CycleNumber,
		FactionID: faction.ID,
		Timestamp: time.Now().UTC(),
		Mutations: records,
	}, nil
}

// resumeOrStart handles resume detection at Cycle entry. If a Cycle is already
// in progress the GM is prompted to resume or abandon. Otherwise a new Cycle is started.
func resumeOrStart(e *engine.Engine, factionState *state.FactionState, p paths.Paths) error {
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
			if err := state.Save(p.State, factionState); err != nil {
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
	if err := state.Save(p.State, factionState); err != nil {
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

func promptSkipTurn(faction *domain.Faction) (bool, error) {
	var skip bool
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Skip %s's turn?", faction.Name)).
				Affirmative("Skip").
				Negative("Take turn").
				Value(&skip),
		),
	).Run(); err != nil {
		return false, fmt.Errorf("prompt cancelled: %w", err)
	}
	return skip, nil
}

func pressEnterToContinue() {
	fmt.Print("  Press Enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
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
