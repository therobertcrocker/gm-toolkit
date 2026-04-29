package domain

// Mutation represents a discrete state change produced by an action or
// bookkeeping phase. Type returns a stable string discriminator used for
// history serialization.
type Mutation interface {
	Type() string
}

// CoinDelta adjusts a faction's Coin balance by Delta (positive or negative).
type CoinDelta struct {
	FactionID         string `json:"faction_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation CoinDelta) Type() string { return "coin_delta" }

// AssetRemoved removes an asset from a faction's roster.
type AssetRemoved struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Cause             string `json:"cause"`              // "attack", "sell", "refit", "bookkeeping"
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetRemoved) Type() string { return "asset_removed" }

// AssetMaintainedFlag updates the Maintained flag on a specific asset.
type AssetMaintainedFlag struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Maintained        bool   `json:"maintained"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetMaintainedFlag) Type() string { return "asset_maintained_flag" }

// FactionHPDelta adjusts a faction's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds MaxHP.
type FactionHPDelta struct {
	FactionID         string `json:"faction_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation FactionHPDelta) Type() string { return "faction_hp_delta" }

// AssetHPDelta adjusts an asset's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds the asset's definition max HP.
type AssetHPDelta struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetHPDelta) Type() string { return "asset_hp_delta" }

// AssetAdded adds a newly purchased or created asset to a faction's roster.
// The asset is flagged Ready: false (inactive until the start of the next turn).
type AssetAdded struct {
	FactionID         string `json:"faction_id"`
	Asset             Asset  `json:"asset"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetAdded) Type() string { return "asset_added" }

// AssetStealthCleared flips an asset's Stealthy flag to false. Per SWN,
// stealth is lost when an asset attacks or defends; Attack resolution emits
// this before any HP mutation in the same matchup so the timeline in the
// EventRecord is consistent.
type AssetStealthCleared struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetStealthCleared) Type() string { return "asset_stealth_cleared" }

// BaseHPDelta adjusts a Base of Influence's CurrentHP by Delta. Per SWN,
// damage to a Base is also dealt to faction HP; callers must emit an
// accompanying FactionHPDelta with the same delta value.
type BaseHPDelta struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation BaseHPDelta) Type() string { return "base_hp_delta" }

// BaseDestroyed removes a Base of Influence from a faction. Emitted inline
// after a BaseHPDelta that takes CurrentHP to 0 or below.
type BaseDestroyed struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation BaseDestroyed) Type() string { return "base_destroyed" }

// BaseAdded places a new Base of Influence on a world. The base is flagged
// Ready: false (inactive until the start of the next turn).
type BaseAdded struct {
	FactionID         string `json:"faction_id"`
	Base              Base   `json:"base"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation BaseAdded) Type() string { return "base_added" }

// BaseHealed restores CurrentHP on a Base of Influence without changing MaxHP.
// Unlike BaseHPDelta (attack damage), healing carries no faction HP side effect.
type BaseHealed struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation BaseHealed) Type() string { return "base_healed" }

// BaseExpanded increases both MaxHP and CurrentHP on a Base of Influence by
// Delta. Used when purchasing additional HP capacity via Expand Influence.
type BaseExpanded struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation BaseExpanded) Type() string { return "base_expanded" }

// AssetMoved updates an asset's Location. Emitted by movement ability steps.
type AssetMoved struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	FromLocation      string `json:"from_location"`
	ToLocation        string `json:"to_location"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetMoved) Type() string { return "asset_moved" }

// AssetStealthApplied marks an asset as stealthy. Emitted by Buy Asset when
// a Stealth-type Cunning asset is purchased.
type AssetStealthApplied struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation AssetStealthApplied) Type() string { return "asset_stealth_applied" }

// GoalAbandoned records that a faction abandoned their active goal. Income for
// the turn is forfeited via a separate CoinDelta emitted alongside this.
type GoalAbandoned struct {
	FactionID         string `json:"faction_id"`
	GoalID            string `json:"goal_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation GoalAbandoned) Type() string { return "goal_abandoned" }

// GoalCompleted records successful goal resolution. XPAwarded is the amount
// granted; a separate XPAwarded mutation applies it to faction.XP.
type GoalCompleted struct {
	FactionID         string `json:"faction_id"`
	GoalID            string `json:"goal_id"`
	XPAwarded         int    `json:"xp_awarded"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation GoalCompleted) Type() string { return "goal_completed" }

// XPAwarded increments a faction's XP by Amount.
type XPAwarded struct {
	FactionID         string `json:"faction_id"`
	Amount            int    `json:"amount"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation XPAwarded) Type() string { return "xp_awarded" }

// HomeworldChanged updates a faction's homeworld on Change Homeworld completion.
type HomeworldChanged struct {
	FactionID         string `json:"faction_id"`
	FromWorld         string `json:"from_world"`
	ToWorld           string `json:"to_world"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation HomeworldChanged) Type() string { return "homeworld_changed" }

// TagAdded grants a tag to a faction. The full Tag is embedded so Apply does
// not need a Rulebook lookup.
type TagAdded struct {
	FactionID         string `json:"faction_id"`
	Tag               Tag    `json:"tag"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation TagAdded) Type() string { return "tag_added" }

// InfluenceDelta adds Delta to a Base's Influence field. Emitted by the Bribe action.
type InfluenceDelta struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (mutation InfluenceDelta) Type() string { return "influence_delta" }
