package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesFileAndSymlink(t *testing.T) {
	dir := t.TempDir()

	log, err := New(dir, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	log.Info("hello world", "k", "v")

	logFile := findCurrentLogFile(t, dir)

	target, err := os.Readlink(filepath.Join(dir, latestSymlink))
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != logFile {
		t.Errorf("symlink target = %q, want %q", target, logFile)
	}

	content, err := os.ReadFile(filepath.Join(dir, logFile))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(content), "hello world") {
		t.Errorf("expected emitted line in log file, got: %q", string(content))
	}
}

func TestNewCleanupKeepsCap(t *testing.T) {
	dir := t.TempDir()

	for i := range 60 {
		name := filepath.Join(dir, fmt.Sprintf("2024-01-01_%06d.log", i))
		if err := os.WriteFile(name, []byte{}, 0o644); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	if _, err := New(dir, false); err != nil {
		t.Fatalf("New: %v", err)
	}

	count := countLogFiles(t, dir)
	if count != keepCount {
		t.Errorf("got %d .log files, want %d", count, keepCount)
	}
}

func TestNewCleanupUnderCapNoOp(t *testing.T) {
	dir := t.TempDir()

	for i := range 10 {
		name := filepath.Join(dir, fmt.Sprintf("2024-01-01_%06d.log", i))
		if err := os.WriteFile(name, []byte{}, 0o644); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	if _, err := New(dir, false); err != nil {
		t.Fatalf("New: %v", err)
	}

	count := countLogFiles(t, dir)
	want := 11
	if count != want {
		t.Errorf("got %d .log files, want %d (10 seeded + 1 new)", count, want)
	}
}

func TestNewCreatesLogsDir(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "logs")

	if _, err := New(dir, false); err != nil {
		t.Fatalf("New: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory at %s", dir)
	}
}

func findCurrentLogFile(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if e.Name() == latestSymlink {
			continue
		}
		if filepath.Ext(e.Name()) == ".log" {
			return e.Name()
		}
	}
	t.Fatal("no log file created")
	return ""
}

func countLogFiles(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.Name() == latestSymlink {
			continue
		}
		if filepath.Ext(e.Name()) == ".log" {
			count++
		}
	}
	return count
}
