package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	keepCount     = 50
	latestSymlink = "latest.log"
)

func New(logsDir string, debug bool) (*slog.Logger, error) {
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create logs dir: %w", err)
	}

	filename := time.Now().Format("2006-01-02_150405") + ".log"
	logPath := filepath.Join(logsDir, filename)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	symlinkPath := filepath.Join(logsDir, latestSymlink)
	_ = os.Remove(symlinkPath)
	symlinkErr := os.Symlink(filename, symlinkPath)

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	log := slog.New(newHandler(f, level, debug))

	if symlinkErr != nil {
		log.Warn("latest.log symlink failed", "err", symlinkErr)
	}
	if err := cleanup(logsDir, filename); err != nil {
		log.Warn("log cleanup failed", "err", err)
	}

	return log, nil
}

func cleanup(logsDir, currentFile string) error {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return fmt.Errorf("read logs dir: %w", err)
	}

	var logs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == latestSymlink || name == currentFile {
			continue
		}
		if filepath.Ext(name) != ".log" {
			continue
		}
		logs = append(logs, name)
	}

	sort.Strings(logs)

	excess := len(logs) - (keepCount - 1)
	if excess <= 0 {
		return nil
	}

	var failed int
	var lastErr error
	for _, name := range logs[:excess] {
		if err := os.Remove(filepath.Join(logsDir, name)); err != nil {
			failed++
			lastErr = err
		}
	}
	if failed > 0 {
		return fmt.Errorf("cleanup: %d file(s) not removed (last error: %w)", failed, lastErr)
	}
	return nil
}
