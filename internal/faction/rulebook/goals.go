package rulebook

import (
	"fmt"
	"maps"
	"slices"

	"github.com/BurntSushi/toml"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

type goalFile struct {
	Goals map[string]goalRecord `toml:"goals"`
}

type goalRecord struct {
	ID          string               `toml:"id"`
	Name        string               `toml:"name"`
	Description string               `toml:"description"`
	Start       *goalStartRecord     `toml:"start"`
	Progress    *goalProgressRecord  `toml:"progress"`
	Reset       *goalEventRecord     `toml:"reset"`
	Threshold   *goalThresholdRecord `toml:"threshold"`
	Complete    *goalEventRecord     `toml:"complete"`
	XP          *goalXPRecord        `toml:"xp"`
}

type goalStartRecord struct {
	Inputs       []string        `toml:"inputs"`
	NeedsActions []string        `toml:"needs_actions"`
	Gate         *goalGateRecord `toml:"gate"`
}

type goalGateRecord struct {
	RatingAboveLastCompletion string `toml:"rating_above_last_completion"`
}

type goalEventRecord struct {
	Event   string             `toml:"event"`
	Filters *goalFiltersRecord `toml:"filters"`
}

type goalProgressRecord struct {
	Event   string             `toml:"event"`
	Amount  string             `toml:"amount"`
	Measure string             `toml:"measure"`
	Filters *goalFiltersRecord `toml:"filters"`
}

type goalThresholdRecord struct {
	Ratings    []string `toml:"ratings"`
	Multiplier int      `toml:"multiplier"`
	Constant   int      `toml:"constant"`
}

type goalXPRecord struct {
	Formula        string `toml:"formula"`
	Amount         int    `toml:"amount"`
	ContestedBonus int    `toml:"contested_bonus"`
}

type goalFiltersRecord struct {
	Owner              string   `toml:"owner"`
	CausedBy           string   `toml:"caused_by"`
	Faction            string   `toml:"faction"`
	Category           []string `toml:"category"`
	RatedAbove         string   `toml:"rated_above"`
	NewWorld           bool     `toml:"new_world"`
	To                 string   `toml:"to"`
	Target             string   `toml:"target"`
	Action             []string `toml:"action"`
	ActionNot          []string `toml:"action_not"`
	Ground             string   `toml:"ground"`
	ExcludeHeldAtStart bool     `toml:"exclude_held_at_start"`
}

// loadGoals reads the goals strictly and converts their stat names. It does not
// check a goal-only name (an event, a formula): rulebook must not import the goal
// engine, so engine construction checks those (A.001.8 Effort 1 Decision 6).
func loadGoals(path string) (map[string]*domain.Goal, error) {
	var file goalFile
	metadata, err := toml.DecodeFile(path, &file)
	if err != nil {
		return nil, err
	}
	if err := refuseUnknownKeys(path, metadata); err != nil {
		return nil, err
	}
	result := make(map[string]*domain.Goal, len(file.Goals))
	for _, key := range slices.Sorted(maps.Keys(file.Goals)) {
		record := file.Goals[key]
		if _, exists := result[record.ID]; exists {
			return nil, fmt.Errorf("duplicate goal ID %q in %s", record.ID, path)
		}
		goal, err := convertGoal(record)
		if err != nil {
			return nil, fmt.Errorf("goal %q in %s: %w", record.ID, path, err)
		}
		result[record.ID] = goal
	}
	return result, nil
}

func convertGoal(record goalRecord) (*domain.Goal, error) {
	goal := &domain.Goal{ID: record.ID, Name: record.Name, Description: record.Description}
	if start := record.Start; start != nil {
		goal.Start = &domain.GoalStart{Inputs: start.Inputs, NeedsActions: start.NeedsActions}
		if start.Gate != nil {
			rating, err := toFactionStat(start.Gate.RatingAboveLastCompletion)
			if err != nil {
				return nil, fmt.Errorf("start.gate rating_above_last_completion: %w", err)
			}
			goal.Start.Gate = &domain.GoalGate{RatingAboveLastCompletion: rating}
		}
	}
	if progress := record.Progress; progress != nil {
		filters, err := convertGoalFilters(progress.Filters)
		if err != nil {
			return nil, fmt.Errorf("progress: %w", err)
		}
		goal.Progress = &domain.GoalProgress{Event: progress.Event, Amount: progress.Amount, Measure: progress.Measure, Filters: filters}
	}
	reset, err := convertGoalEvent("reset", record.Reset)
	if err != nil {
		return nil, err
	}
	goal.Reset = reset
	if threshold := record.Threshold; threshold != nil {
		ratings, err := toFactionStats(threshold.Ratings)
		if err != nil {
			return nil, fmt.Errorf("threshold ratings%w", err)
		}
		goal.Threshold = &domain.GoalThreshold{Ratings: ratings, Multiplier: threshold.Multiplier, Constant: threshold.Constant}
	}
	complete, err := convertGoalEvent("complete", record.Complete)
	if err != nil {
		return nil, err
	}
	goal.Complete = complete
	if xp := record.XP; xp != nil {
		goal.XP = &domain.GoalXP{Formula: xp.Formula, Amount: xp.Amount, ContestedBonus: xp.ContestedBonus}
	}
	return goal, nil
}

func convertGoalEvent(part string, record *goalEventRecord) (*domain.GoalEvent, error) {
	if record == nil {
		return nil, nil
	}
	filters, err := convertGoalFilters(record.Filters)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", part, err)
	}
	return &domain.GoalEvent{Event: record.Event, Filters: filters}, nil
}

func convertGoalFilters(record *goalFiltersRecord) (domain.GoalFilters, error) {
	if record == nil {
		return domain.GoalFilters{}, nil
	}
	filters := domain.GoalFilters{
		Owner:              record.Owner,
		CausedBy:           record.CausedBy,
		Faction:            record.Faction,
		NewWorld:           record.NewWorld,
		To:                 record.To,
		Target:             record.Target,
		Action:             record.Action,
		ActionNot:          record.ActionNot,
		Ground:             record.Ground,
		ExcludeHeldAtStart: record.ExcludeHeldAtStart,
	}
	category, err := toFactionStats(record.Category)
	if err != nil {
		return domain.GoalFilters{}, fmt.Errorf("filters.category%w", err)
	}
	filters.Category = category
	if record.RatedAbove != "" {
		ratedAbove, err := toFactionStat(record.RatedAbove)
		if err != nil {
			return domain.GoalFilters{}, fmt.Errorf("filters.rated_above: %w", err)
		}
		filters.RatedAbove = ratedAbove
	}
	return filters, nil
}

// toFactionStats converts a list of stat names. Its error starts with the index,
// so the caller's message reads "filters.category[1]: unknown faction stat".
func toFactionStats(names []string) ([]domain.FactionStat, error) {
	if len(names) == 0 {
		return nil, nil
	}
	stats := make([]domain.FactionStat, 0, len(names))
	for index, name := range names {
		stat, err := toFactionStat(name)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", index, err)
		}
		stats = append(stats, stat)
	}
	return stats, nil
}
