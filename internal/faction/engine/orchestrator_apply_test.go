package engine

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestApplyAndRecord_LogsMutationMisses(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	eng := &Engine{Mutation: mutation.New(slog.New(slog.DiscardHandler))}
	factionState := &state.FactionState{Factions: make(map[string]*domain.Faction)}
	paths := &campaigns.Paths{HistoryPath: t.TempDir() + "/history.jsonl"}

	t.Run("single miss emits one Error log line", func(t *testing.T) {
		buf.Reset()
		mutations := []domain.Mutation{
			domain.CoinDelta{FactionID: "ghost", Delta: 1},
		}
		err := eng.applyAndRecord(factionState, &domain.Faction{ID: "ghost"}, mutations, paths, log)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var applyErr *mutation.MutationApplyError
		if !errors.As(err, &applyErr) {
			t.Fatalf("error does not wrap *MutationApplyError: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "mutation miss") {
			t.Errorf("log missing 'mutation miss' line; output:\n%s", output)
		}
		if !strings.Contains(output, "ghost") {
			t.Errorf("log missing faction_id 'ghost'; output:\n%s", output)
		}
	})

	t.Run("N misses emit N Error log lines", func(t *testing.T) {
		buf.Reset()
		mutations := []domain.Mutation{
			domain.CoinDelta{FactionID: "ghost-a", Delta: 1},
			domain.CoinDelta{FactionID: "ghost-b", Delta: 1},
		}
		err := eng.applyAndRecord(factionState, &domain.Faction{ID: "ghost-a"}, mutations, paths, log)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		count := strings.Count(buf.String(), "mutation miss")
		if count != 2 {
			t.Errorf("expected 2 'mutation miss' log lines, got %d; output:\n%s", count, buf.String())
		}
	})
}
