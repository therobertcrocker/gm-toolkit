package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// HistoryEngine appends faction turn event records to the campaign history file.
type HistoryEngine struct{}

func newHistoryEngine() *HistoryEngine {
	return &HistoryEngine{}
}

// Record marshals event to JSON and appends it as a single line to historyPath,
// creating the file if it does not exist.
func (history *HistoryEngine) Record(historyPath string, event domain.EventRecord) error {
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
