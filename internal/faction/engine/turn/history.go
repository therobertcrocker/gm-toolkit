package turn

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RecordHistory marshals an EventRecord for faction's turn mutations and appends
// it as a single JSON line to historyPath.
func RecordHistory(historyPath string, factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation, log *slog.Logger) error {
	event, err := buildEventRecord(factionState, faction, mutations)
	if err != nil {
		return fmt.Errorf("building event record: %w", err)
	}

	file, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening history file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	if _, err := fmt.Fprintf(file, "%s\n", data); err != nil {
		return fmt.Errorf("writing event: %w", err)
	}
	return nil
}
