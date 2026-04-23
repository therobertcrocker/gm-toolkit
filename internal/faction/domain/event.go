package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// MutationRecord is the serializable form of a Mutation written to history.
// Payload holds the structured mutation fields as raw JSON, keyed by the
// concrete type's json tags.
type MutationRecord struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewMutationRecord marshals a Mutation into its history record form.
func NewMutationRecord(mutation Mutation) (MutationRecord, error) {
	data, err := json.Marshal(mutation)
	if err != nil {
		return MutationRecord{}, fmt.Errorf("marshaling mutation: %w", err)
	}
	return MutationRecord{Type: mutation.Type(), Payload: json.RawMessage(data)}, nil
}

// EventRecord captures everything that happened during a single faction's turn.
// One record is appended to history.jsonl per faction per cycle.
type EventRecord struct {
	Cycle     int              `json:"cycle"`
	FactionID string           `json:"faction_id"`
	Timestamp time.Time        `json:"timestamp"`
	Mutations []MutationRecord `json:"mutations"`
	Detail    map[string]any   `json:"detail,omitempty"`
}
