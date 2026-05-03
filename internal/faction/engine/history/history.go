package history

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type HistoryEngine struct{}

func New() *HistoryEngine { return &HistoryEngine{} }

// Record builds an EventRecord from factionState, faction, and mutations, then
// marshals it as a single JSON line appended to historyPath.
func (h *HistoryEngine) Record(historyPath string, factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) error {
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
