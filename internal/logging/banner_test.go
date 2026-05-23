package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRunHeader(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(newHandler(&buf, slog.LevelInfo, false))
	RunHeader(log, "faction-manager", "0.6.4", "Cygnus Drift", 8)

	got := buf.String()
	lines := strings.Split(got, "\n")
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d:\n%s", len(lines), got)
	}

	wantFixed := map[int]string{
		0: "=== faction-manager 0.6.4 ===",
		1: "Campaign: Cygnus Drift",
		2: "Factions: 8",
		4: "",
	}
	for idx, want := range wantFixed {
		if lines[idx] != want {
			t.Errorf("line %d: got %q, want %q", idx, lines[idx], want)
		}
	}
	if !strings.HasPrefix(lines[3], "Started: ") {
		t.Errorf("line 3: expected \"Started: ...\" prefix, got %q", lines[3])
	}
}

func TestTurnStart(t *testing.T) {
	tests := []struct {
		name    string
		turn    int
		faction string
		want    string
	}{
		{
			name:    "short faction",
			turn:    1,
			faction: "Acme",
			want: "┌─────────────────┐\n" +
				"│  Turn 1 · Acme  │\n" +
				"└─────────────────┘\n",
		},
		{
			name:    "long faction name",
			turn:    12,
			faction: "Outer Rim Confederation",
			want: "┌─────────────────────────────────────┐\n" +
				"│  Turn 12 · Outer Rim Confederation  │\n" +
				"└─────────────────────────────────────┘\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			TurnStart(slog.New(newHandler(&buf, slog.LevelInfo, false)), testCase.turn, testCase.faction)
			if got := buf.String(); got != testCase.want {
				t.Errorf("\ngot:\n%s\nwant:\n%s", got, testCase.want)
			}
		})
	}
}

func TestPhaseStart(t *testing.T) {
	tests := []struct {
		name  string
		phase string
		want  string
	}{
		{"single word", "movement", "  --- Movement Phase ---\n"},
		{"snake case two words", "goal_lock", "  --- Goal Lock Phase ---\n"},
		{"snake case two words alt", "stat_raise", "  --- Stat Raise Phase ---\n"},
		{"bookkeeping", "bookkeeping", "  --- Bookkeeping Phase ---\n"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			PhaseStart(slog.New(newHandler(&buf, slog.LevelInfo, false)), testCase.phase)
			if got := buf.String(); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}
