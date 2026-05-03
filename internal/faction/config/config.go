// Package config holds runtime configuration for engine consumers — the paths
// the engine needs to persist state and append history.
package config

// Config carries runtime paths injected by the caller. Future runtime knobs
// (dry-run, log level, AI settings) join here as the orchestrator grows.
type Config struct {
	StatePath   string
	HistoryPath string
}
