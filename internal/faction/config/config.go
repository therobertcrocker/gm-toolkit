// Package config holds runtime configuration for engine consumers — the paths
// the engine needs to load data, persist state, and append history.
package config

// Config carries all runtime paths injected by the caller.
type Config struct {
	FactionDataDir string
	SpatialDataDir string
	StatePath      string
	HistoryPath    string
	NarrativesPath string
}
