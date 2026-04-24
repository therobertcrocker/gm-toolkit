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

// FactionHPDelta adjusts a faction's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds MaxHP.
type FactionHPDelta struct {
	FactionID string `json:"faction_id"`
	Delta     int    `json:"delta"`
}

func (mutation FactionHPDelta) Type() string { return "faction_hp_delta" }

// AssetHPDelta adjusts an asset's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds the asset's definition max HP.
type AssetHPDelta struct {
	FactionID string `json:"faction_id"`
	AssetID   string `json:"asset_id"`
	Delta     int    `json:"delta"`
}

func (mutation AssetHPDelta) Type() string { return "asset_hp_delta" }

// AssetAdded adds a newly purchased or created asset to a faction's roster.
// The asset is flagged Ready: false (inactive until the start of the next turn).
type AssetAdded struct {
	FactionID string `json:"faction_id"`
	Asset     Asset  `json:"asset"`
}

func (mutation AssetAdded) Type() string { return "asset_added" }
