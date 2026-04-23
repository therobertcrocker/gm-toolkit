package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func TestHistoryEngine_Record(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	he := newHistoryEngine()

	coinRecord, err := domain.NewMutationRecord(domain.CoinDelta{FactionID: "f1", Delta: 4})
	if err != nil {
		t.Fatalf("NewMutationRecord: %v", err)
	}
	assetRecord, err := domain.NewMutationRecord(domain.AssetRemoved{FactionID: "f1", AssetID: "a1"})
	if err != nil {
		t.Fatalf("NewMutationRecord: %v", err)
	}

	event := domain.EventRecord{
		Cycle:     3,
		FactionID: "f1",
		Timestamp: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC),
		Mutations: []domain.MutationRecord{coinRecord, assetRecord},
	}

	if err := he.Record(historyPath, event); err != nil {
		t.Fatalf("Record: %v", err)
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got domain.EventRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Cycle != 3 {
		t.Errorf("Cycle = %d, want 3", got.Cycle)
	}
	if got.FactionID != "f1" {
		t.Errorf("FactionID = %q, want %q", got.FactionID, "f1")
	}
	if len(got.Mutations) != 2 {
		t.Fatalf("len(Mutations) = %d, want 2", len(got.Mutations))
	}
	if got.Mutations[0].Type != "coin_delta" {
		t.Errorf("Mutations[0].Type = %q, want %q", got.Mutations[0].Type, "coin_delta")
	}
	if got.Mutations[1].Type != "asset_removed" {
		t.Errorf("Mutations[1].Type = %q, want %q", got.Mutations[1].Type, "asset_removed")
	}
}

func TestHistoryEngine_RecordAppends(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	he := newHistoryEngine()

	event1 := domain.EventRecord{Cycle: 1, FactionID: "f1", Timestamp: time.Now().UTC()}
	event2 := domain.EventRecord{Cycle: 1, FactionID: "f2", Timestamp: time.Now().UTC()}

	if err := he.Record(historyPath, event1); err != nil {
		t.Fatalf("Record 1: %v", err)
	}
	if err := he.Record(historyPath, event2); err != nil {
		t.Fatalf("Record 2: %v", err)
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines in history file, got %d", len(lines))
	}

	var got1, got2 domain.EventRecord
	if err := json.Unmarshal([]byte(lines[0]), &got1); err != nil {
		t.Fatalf("Unmarshal line 1: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &got2); err != nil {
		t.Fatalf("Unmarshal line 2: %v", err)
	}
	if got1.FactionID != "f1" {
		t.Errorf("line 1 FactionID = %q, want f1", got1.FactionID)
	}
	if got2.FactionID != "f2" {
		t.Errorf("line 2 FactionID = %q, want f2", got2.FactionID)
	}
}
