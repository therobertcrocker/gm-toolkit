package domain

import "fmt"

// Mutation represents a discrete state change. All mutations carry a stable
// type identifier for history serialization and a human-readable description
// for the narrative renderer.
type Mutation interface {
	Type() string
	Describe() string
}

// CoinDelta adjusts a faction's Coin balance by Delta (positive or negative).
type CoinDelta struct {
	FactionID string
	Delta     int
}

func (m CoinDelta) Type() string { return "coin_delta" }
func (m CoinDelta) Describe() string {
	if m.Delta >= 0 {
		return fmt.Sprintf("faction %s gained %d Coin", m.FactionID, m.Delta)
	}
	return fmt.Sprintf("faction %s lost %d Coin", m.FactionID, -m.Delta)
}

// AssetRemoved removes an asset from a faction's roster.
type AssetRemoved struct {
	FactionID string
	AssetID   string
}

func (m AssetRemoved) Type() string    { return "asset_removed" }
func (m AssetRemoved) Describe() string {
	return fmt.Sprintf("faction %s lost asset %s", m.FactionID, m.AssetID)
}

// AssetMaintainedFlag updates the Maintained flag on a specific asset.
type AssetMaintainedFlag struct {
	FactionID  string
	AssetID    string
	Maintained bool
}

func (m AssetMaintainedFlag) Type() string { return "asset_maintained_flag" }
func (m AssetMaintainedFlag) Describe() string {
	if m.Maintained {
		return fmt.Sprintf("asset %s (faction %s) is now maintained", m.AssetID, m.FactionID)
	}
	return fmt.Sprintf("asset %s (faction %s) is now unmaintained", m.AssetID, m.FactionID)
}
