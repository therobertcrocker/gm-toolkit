package domain

// Mutation represents a discrete state change produced by an action or
// bookkeeping phase. Type returns a stable string discriminator used for
// history serialization.
type Mutation interface {
	Type() string
}

// CoinDelta adjusts a faction's Coin balance by Delta (positive or negative).
type CoinDelta struct {
	FactionID string `json:"faction_id"`
	Delta     int    `json:"delta"`
}

func (mutation CoinDelta) Type() string { return "coin_delta" }

// AssetRemoved removes an asset from a faction's roster.
type AssetRemoved struct {
	FactionID string `json:"faction_id"`
	AssetID   string `json:"asset_id"`
}

func (mutation AssetRemoved) Type() string { return "asset_removed" }

// AssetMaintainedFlag updates the Maintained flag on a specific asset.
type AssetMaintainedFlag struct {
	FactionID  string `json:"faction_id"`
	AssetID    string `json:"asset_id"`
	Maintained bool   `json:"maintained"`
}

func (mutation AssetMaintainedFlag) Type() string { return "asset_maintained_flag" }
