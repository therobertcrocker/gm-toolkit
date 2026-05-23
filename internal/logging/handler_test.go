package logging

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 5, 23, 14, 32, 17, 0, time.UTC)

func TestHandlerOrdinaryLine(t *testing.T) {
	tests := []struct {
		name    string
		level   slog.Level
		message string
		cascade []slog.Attr
		attrs   []slog.Attr
		want    string
	}{
		{
			name:    "info no attrs",
			level:   slog.LevelInfo,
			message: "engine started",
			want:    "14:32:17  INFO   engine started\n",
		},
		{
			name:    "warn level",
			level:   slog.LevelWarn,
			message: "cleanup partial",
			want:    "14:32:17  WARN   cleanup partial\n",
		},
		{
			name:    "error level",
			level:   slog.LevelError,
			message: "open failed",
			want:    "14:32:17  ERROR  open failed\n",
		},
		{
			name:    "debug level",
			level:   slog.LevelDebug,
			message: "tick",
			want:    "14:32:17  DEBUG  tick\n",
		},
		{
			name:    "event attrs",
			level:   slog.LevelInfo,
			message: "asset moved",
			attrs: []slog.Attr{
				slog.String("asset", "ALPHA"),
				slog.String("from_hex", "0101"),
				slog.String("to_hex", "0102"),
			},
			want: "14:32:17  INFO   asset moved  asset=ALPHA  from_hex=0101  to_hex=0102\n",
		},
		{
			name:    "cascade attrs suppressed",
			level:   slog.LevelInfo,
			message: "phase begin",
			cascade: []slog.Attr{
				slog.Int("turn", 3),
				slog.String("faction", "acme"),
				slog.String("phase", "movement"),
			},
			attrs: []slog.Attr{slog.Int("orders", 5)},
			want:  "14:32:17  INFO   phase begin  orders=5\n",
		},
		{
			name:    "value with spaces gets quoted",
			level:   slog.LevelInfo,
			message: "loaded campaign",
			attrs:   []slog.Attr{slog.String("name", "Cygnus Drift")},
			want:    "14:32:17  INFO   loaded campaign  name=\"Cygnus Drift\"\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			var sink slog.Handler = newHandler(&buf, slog.LevelDebug, false)
			if len(testCase.cascade) > 0 {
				sink = sink.WithAttrs(testCase.cascade)
			}
			record := slog.NewRecord(fixedTime, testCase.level, testCase.message, 0)
			record.AddAttrs(testCase.attrs...)
			if err := sink.Handle(context.Background(), record); err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if got := buf.String(); got != testCase.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, testCase.want)
			}
		})
	}
}

func TestHandlerAddSource(t *testing.T) {
	var buf bytes.Buffer
	sink := newHandler(&buf, slog.LevelDebug, true)

	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])

	record := slog.NewRecord(fixedTime, slog.LevelDebug, "ping", pcs[0])
	if err := sink.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "handler_test.go:") {
		t.Errorf("expected handler_test.go:<line> in output, got: %q", got)
	}
	if !strings.Contains(got, "DEBUG") {
		t.Errorf("expected DEBUG level in output, got: %q", got)
	}
	if !strings.Contains(got, "ping") {
		t.Errorf("expected message in output, got: %q", got)
	}
}

func TestHandlerBanners(t *testing.T) {
	tests := []struct {
		name  string
		attrs []slog.Attr
		want  string
	}{
		{
			name: "header",
			attrs: []slog.Attr{
				slog.String("_banner", "header"),
				slog.String("tool", "faction-manager"),
				slog.String("version", "0.6.4"),
				slog.String("campaign", "Cygnus Drift"),
				slog.Int("factions", 8),
			},
			want: "=== faction-manager 0.6.4 ===\n" +
				"Campaign: Cygnus Drift\n" +
				"Factions: 8\n" +
				"Started: 2026-05-23 14:32:17\n\n",
		},
		{
			name: "turn",
			attrs: []slog.Attr{
				slog.String("_banner", "turn"),
				slog.Int("turn", 1),
				slog.String("faction", "Acme"),
			},
			want: "┌─────────────────┐\n" +
				"│  Turn 1 · Acme  │\n" +
				"└─────────────────┘\n",
		},
		{
			name: "turn with long faction name",
			attrs: []slog.Attr{
				slog.String("_banner", "turn"),
				slog.Int("turn", 12),
				slog.String("faction", "Outer Rim Confederation"),
			},
			want: "┌─────────────────────────────────────┐\n" +
				"│  Turn 12 · Outer Rim Confederation  │\n" +
				"└─────────────────────────────────────┘\n",
		},
		{
			name: "phase",
			attrs: []slog.Attr{
				slog.String("_banner", "phase"),
				slog.String("phase", "goal_lock"),
			},
			want: "  --- Goal Lock Phase ---\n",
		},
		{
			name: "phase single word",
			attrs: []slog.Attr{
				slog.String("_banner", "phase"),
				slog.String("phase", "movement"),
			},
			want: "  --- Movement Phase ---\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			sink := newHandler(&buf, slog.LevelInfo, false)
			record := slog.NewRecord(fixedTime, slog.LevelInfo, "", 0)
			record.AddAttrs(testCase.attrs...)
			if err := sink.Handle(context.Background(), record); err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if got := buf.String(); got != testCase.want {
				t.Errorf("\ngot:\n%s\nwant:\n%s", got, testCase.want)
			}
		})
	}
}

func TestHandlerCascadeChains(t *testing.T) {
	var buf bytes.Buffer
	var sink slog.Handler = newHandler(&buf, slog.LevelInfo, false)
	sink = sink.WithAttrs([]slog.Attr{slog.Int("turn", 1)})
	sink = sink.WithAttrs([]slog.Attr{slog.String("phase", "movement")})

	record := slog.NewRecord(fixedTime, slog.LevelInfo, "tick", 0)
	record.AddAttrs(slog.Int("orders", 2))
	if err := sink.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	want := "14:32:17  INFO   tick  orders=2\n"
	if got := buf.String(); got != want {
		t.Errorf("\ngot:  %q\nwant: %q", got, want)
	}
}

func TestHandlerWithAttrsSiblingIsolation(t *testing.T) {
	var buf bytes.Buffer
	parent := newHandler(&buf, slog.LevelInfo, false)

	childA := parent.WithAttrs([]slog.Attr{slog.String("version", "1.0")})
	childB := parent.WithAttrs([]slog.Attr{slog.String("version", "2.0")})

	recordA := slog.NewRecord(fixedTime, slog.LevelInfo, "from-a", 0)
	recordB := slog.NewRecord(fixedTime, slog.LevelInfo, "from-b", 0)
	_ = childA.Handle(context.Background(), recordA)
	_ = childB.Handle(context.Background(), recordB)

	out := buf.String()
	if !strings.Contains(out, "from-a  version=1.0") {
		t.Errorf("expected from-a with version=1.0, got: %q", out)
	}
	if !strings.Contains(out, "from-b  version=2.0") {
		t.Errorf("expected from-b with version=2.0, got: %q", out)
	}
}
