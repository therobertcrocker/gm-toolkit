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

// AssetStealthCleared flips an asset's Stealthy flag to false. Per SWN,
// stealth is lost when an asset attacks or defends; Attack resolution emits
// this before any HP mutation in the same matchup so the timeline in the
// EventRecord is consistent.
type AssetStealthCleared struct {
	FactionID string `json:"faction_id"`
	AssetID   string `json:"asset_id"`
}

func (mutation AssetStealthCleared) Type() string { return "asset_stealth_cleared" }

// BaseHPDelta adjusts a Base of Influence's CurrentHP by Delta. Per SWN,
// damage to a Base is also dealt to faction HP; callers must emit an
// accompanying FactionHPDelta with the same delta value.
type BaseHPDelta struct {
	FactionID string `json:"faction_id"`
	BaseID    string `json:"base_id"`
	Delta     int    `json:"delta"`
}

func (mutation BaseHPDelta) Type() string { return "base_hp_delta" }

// BaseDestroyed removes a Base of Influence from a faction. Emitted inline
// after a BaseHPDelta that takes CurrentHP to 0 or below.
type BaseDestroyed struct {
	FactionID string `json:"faction_id"`
	BaseID    string `json:"base_id"`
}

func (mutation BaseDestroyed) Type() string { return "base_destroyed" }

// BaseAdded places a new Base of Influence on a world. The base is flagged
// Ready: false (inactive until the start of the next turn).
type BaseAdded struct {
	FactionID string `json:"faction_id"`
	Base      Base   `json:"base"`
}

func (mutation BaseAdded) Type() string { return "base_added" }

// BaseHealed restores CurrentHP on a Base of Influence without changing MaxHP.
// Unlike BaseHPDelta (attack damage), healing carries no faction HP side effect.
type BaseHealed struct {
	FactionID string `json:"faction_id"`
	BaseID    string `json:"base_id"`
	Delta     int    `json:"delta"`
}

func (mutation BaseHealed) Type() string { return "base_healed" }

// BaseExpanded increases both MaxHP and CurrentHP on a Base of Influence by
// Delta. Used when purchasing additional HP capacity via Expand Influence.
type BaseExpanded struct {
	FactionID string `json:"faction_id"`
	BaseID    string `json:"base_id"`
	Delta     int    `json:"delta"`
}

func (mutation BaseExpanded) Type() string { return "base_expanded" }

// AssetMoved updates an asset's Location. Emitted by movement ability steps.
type AssetMoved struct {
	FactionID    string `json:"faction_id"`
	AssetID      string `json:"asset_id"`
	FromLocation string `json:"from_location"`
	ToLocation   string `json:"to_location"`
}

func (mutation AssetMoved) Type() string { return "asset_moved" }
