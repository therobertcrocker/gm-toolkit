package narrative

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// LoadCycle reads history.jsonl at historyPath and returns all EventRecords
// whose Cycle field matches cycleNumber. Returns an error if no records match.
func LoadCycle(historyPath string, cycleNumber int) ([]domain.EventRecord, error) {
	file, err := os.Open(historyPath)
	if err != nil {
		return nil, fmt.Errorf("opening history file %s: %w", historyPath, err)
	}
	defer file.Close()

	var records []domain.EventRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record domain.EventRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("parsing history record: %w", err)
		}
		if record.Cycle == cycleNumber {
			records = append(records, record)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading history file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no history records for cycle %d", cycleNumber)
	}
	return records, nil
}
