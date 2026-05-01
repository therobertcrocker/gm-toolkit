package narrative

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func writeTempJSONL(t *testing.T, lines []string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "history-*.jsonl")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	for _, line := range lines {
		f.WriteString(line + "\n")
	}
	f.Close()
	return f.Name()
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshaling: %v", err)
	}
	return string(b)
}

func TestLoadCycle(t *testing.T) {
	cycle1 := mustMarshal(t, domain.EventRecord{Cycle: 1, FactionID: "faction-a"})
	cycle2a := mustMarshal(t, domain.EventRecord{Cycle: 2, FactionID: "faction-a"})
	cycle2b := mustMarshal(t, domain.EventRecord{Cycle: 2, FactionID: "faction-b"})

	tests := []struct {
		name        string
		lines       []string
		historyPath func(dir string) string
		cycle       int
		wantCount   int
		wantErr     string
	}{
		{
			name:        "file not found",
			historyPath: func(dir string) string { return filepath.Join(dir, "missing.jsonl") },
			cycle:       1,
			wantErr:     "missing.jsonl",
		},
		{
			name:      "two cycles — returns only matching",
			lines:     []string{cycle1, cycle2a, cycle2b},
			cycle:     2,
			wantCount: 2,
		},
		{
			name:      "one cycle matching",
			lines:     []string{cycle1},
			cycle:     1,
			wantCount: 1,
		},
		{
			name:    "one cycle non-matching",
			lines:   []string{cycle1},
			cycle:   99,
			wantErr: "no history records for cycle 99",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			if tc.historyPath != nil {
				path = tc.historyPath(t.TempDir())
			} else {
				path = writeTempJSONL(t, tc.lines)
			}

			records, err := LoadCycle(path, tc.cycle)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(records) != tc.wantCount {
				t.Fatalf("expected %d records, got %d", tc.wantCount, len(records))
			}
		})
	}
}
