package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

var suppressedFromLine = map[string]bool{
	"_banner": true,
	"turn":    true,
	"faction": true,
	"phase":   true,
}

type handler struct {
	mu        *sync.Mutex
	writer    io.Writer
	level     slog.Level
	addSource bool
	attrs     []slog.Attr
}

func newHandler(writer io.Writer, level slog.Level, addSource bool) *handler {
	return &handler{
		mu:        &sync.Mutex{},
		writer:    writer,
		level:     level,
		addSource: addSource,
	}
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	next.attrs = append(next.attrs, h.attrs...)
	next.attrs = append(next.attrs, attrs...)
	return &next
}

// Groups are intentionally ignored: this handler renders flat key=value pairs.
func (h *handler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	all := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	all = append(all, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		all = append(all, a)
		return true
	})

	var buf strings.Builder
	if kind := bannerKind(all); kind != "" {
		renderBanner(&buf, kind, r, attrMap(all))
	} else {
		h.renderLine(&buf, r, all)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.writer, buf.String())
	return err
}

func bannerKind(attrs []slog.Attr) string {
	for _, a := range attrs {
		if a.Key == "_banner" {
			return a.Value.String()
		}
	}
	return ""
}

func attrMap(attrs []slog.Attr) map[string]slog.Value {
	m := make(map[string]slog.Value, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value
	}
	return m
}

func (h *handler) renderLine(buf *strings.Builder, r slog.Record, attrs []slog.Attr) {
	fmt.Fprintf(buf, "%s  %-5s  ", r.Time.Format("15:04:05"), levelString(r.Level))

	if h.addSource && r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		fmt.Fprintf(buf, "%s:%d  ", filepath.Base(frame.File), frame.Line)
	}

	buf.WriteString(r.Message)

	for _, a := range attrs {
		if suppressedFromLine[a.Key] {
			continue
		}
		fmt.Fprintf(buf, "  %s=%s", a.Key, formatValue(a.Value))
	}
	buf.WriteByte('\n')
}

func levelString(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	default:
		return l.String()
	}
}

func formatValue(v slog.Value) string {
	s := v.String()
	if strings.ContainsAny(s, " \t\"") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// renderBanner writes one of three banner shapes (header, turn, phase) into buf.
// Available attrs by kind:
//   - "header": tool (string), version (string), campaign (string), factions (int) + record.Time
//   - "turn":   turn (int), faction (string)
//   - "phase":  phase (string) — snake_case input, rendered in Title Case
//
// Each banner ends with a trailing newline; the header may span multiple lines.
func renderBanner(buf *strings.Builder, kind string, record slog.Record, attrs map[string]slog.Value) {
	switch kind {
	case "header":
		fmt.Fprintf(buf, "=== %s %s ===\n", attrs["tool"].String(), attrs["version"].String())
		fmt.Fprintf(buf, "Campaign: %s\n", attrs["campaign"].String())
		fmt.Fprintf(buf, "Factions: %d\n", attrs["factions"].Int64())
		fmt.Fprintf(buf, "Started: %s\n\n", record.Time.Format("2006-01-02 15:04:05"))

	case "turn":
		content := fmt.Sprintf("  Turn %d · %s  ", attrs["turn"].Int64(), attrs["faction"].String())
		rail := strings.Repeat("─", utf8.RuneCountInString(content))
		fmt.Fprintf(buf, "┌%s┐\n", rail)
		fmt.Fprintf(buf, "│%s│\n", content)
		fmt.Fprintf(buf, "└%s┘\n", rail)

	case "phase":
		phase := titleCase(attrs["phase"].String())
		fmt.Fprintf(buf, "  --- %s Phase ---\n", phase)
	}
}

func titleCase(snake string) string {
	parts := strings.Split(snake, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
