package domain

import "github.com/therobertcrocker/gm-toolkit/internal/spatial"

const (
	CauseMovementTick         = "movement_tick"
	CauseMovementIssue        = "movement_issue"
	CauseMovementRevision     = "movement_revision"
	CauseMovementCancellation = "movement_cancellation"
)

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

func (m CoinDelta) Type() string { return "coin_delta" }

// AssetRemoved removes an asset from a faction's roster.
type AssetRemoved struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Cause             string `json:"cause"` // "attack", "sell", "refit", "bookkeeping"
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetRemoved) Type() string { return "asset_removed" }

// AssetMaintainedFlag updates the Maintained flag on a specific asset.
type AssetMaintainedFlag struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Maintained        bool   `json:"maintained"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetMaintainedFlag) Type() string { return "asset_maintained_flag" }

// FactionHPDelta adjusts a faction's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds MaxHP.
type FactionHPDelta struct {
	FactionID         string `json:"faction_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m FactionHPDelta) Type() string { return "faction_hp_delta" }

// AssetHPDelta adjusts an asset's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds the asset's definition max HP.
type AssetHPDelta struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetHPDelta) Type() string { return "asset_hp_delta" }

// AssetAdded adds a newly purchased or created asset to a faction's roster.
// The asset is flagged Ready: false (inactive until the start of the next turn).
type AssetAdded struct {
	FactionID         string `json:"faction_id"`
	Asset             Asset  `json:"asset"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetAdded) Type() string { return "asset_added" }

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

func (m AssetStealthCleared) Type() string { return "asset_stealth_cleared" }

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

func (m BaseHPDelta) Type() string { return "base_hp_delta" }

// BaseDestroyed removes a Base of Influence from a faction. Emitted inline
// after a BaseHPDelta that takes CurrentHP to 0 or below.
type BaseDestroyed struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m BaseDestroyed) Type() string { return "base_destroyed" }

// BaseAdded places a new Base of Influence on a world. The base is flagged
// Ready: false (inactive until the start of the next turn).
type BaseAdded struct {
	FactionID         string `json:"faction_id"`
	Base              Base   `json:"base"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m BaseAdded) Type() string { return "base_added" }

// BaseHealed restores CurrentHP on a Base of Influence without changing MaxHP.
// Unlike BaseHPDelta (attack damage), healing carries no faction HP side effect.
type BaseHealed struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m BaseHealed) Type() string { return "base_healed" }

// BaseExpanded increases both MaxHP and CurrentHP on a Base of Influence by
// Delta. Used when purchasing additional HP capacity via Expand Influence.
type BaseExpanded struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m BaseExpanded) Type() string { return "base_expanded" }

// AssetMoved updates an asset's Location. Emitted by movement ability steps.
type AssetMoved struct {
	FactionID         string   `json:"faction_id"`
	AssetID           string   `json:"asset_id"`
	FromLocation      Location `json:"from_location"`
	ToLocation        Location `json:"to_location"`
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m AssetMoved) Type() string { return "asset_moved" }

// AssetStealthApplied marks an asset as stealthy. Emitted by Buy Asset when
// a Stealth-type Cunning asset is purchased.
type AssetStealthApplied struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m AssetStealthApplied) Type() string { return "asset_stealth_applied" }

// GoalAbandoned records that a faction abandoned their active goal. Income for
// the turn is forfeited via a separate CoinDelta emitted alongside this.
type GoalAbandoned struct {
	FactionID         string `json:"faction_id"`
	GoalID            string `json:"goal_id"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m GoalAbandoned) Type() string { return "goal_abandoned" }

// GoalCompleted records successful goal resolution. XPAwarded is the amount
// granted; a separate XPAwarded mutation applies it to faction.XP.
type GoalCompleted struct {
	FactionID         string `json:"faction_id"`
	GoalID            string `json:"goal_id"`
	XPAwarded         int    `json:"xp_awarded"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m GoalCompleted) Type() string { return "goal_completed" }

// XPAwarded increments a faction's XP by Amount.
type XPAwarded struct {
	FactionID         string `json:"faction_id"`
	Amount            int    `json:"amount"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m XPAwarded) Type() string { return "xp_awarded" }

// HomeworldChanged updates a faction's homeworld on Change Homeworld completion.
type HomeworldChanged struct {
	FactionID         string   `json:"faction_id"`
	FromWorld         Location `json:"from_world"`
	ToWorld           Location `json:"to_world"`
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m HomeworldChanged) Type() string { return "homeworld_changed" }

// TagAdded grants a tag to a faction. The full Tag is embedded so Apply does
// not need a Rulebook lookup.
type TagAdded struct {
	FactionID         string `json:"faction_id"`
	Tag               Tag    `json:"tag"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m TagAdded) Type() string { return "tag_added" }

// InfluenceDelta adds Delta to a Base's Influence field. Emitted by the Bribe action.
type InfluenceDelta struct {
	FactionID         string `json:"faction_id"`
	BaseID            string `json:"base_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m InfluenceDelta) Type() string { return "influence_delta" }

// GoalInitiated records that a faction initiated a goal. Used for bookkeeping and validation of goal-specific actions; has no direct mechanical effect.
type GoalInitiated struct {
	FactionID         string   `json:"faction_id"`
	GoalID            string   `json:"goal_id"`
	TargetWorld       Location `json:"target_world"`
	ProcessPhase      int      `json:"process_phase"`
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m GoalInitiated) Type() string { return "goal_initiated" }

// GoalProgressed records a progress increment (or reset) on the active goal.
type GoalProgressed struct {
	FactionID         string `json:"faction_id"`
	GoalID            string `json:"goal_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}

func (m GoalProgressed) Type() string { return "goal_progressed" }

// GoalTurnsTick decrements ActiveGoal.TurnsRemaining by 1.
type GoalTurnsTick struct {
	FactionID string `json:"faction_id"`
	GoalID    string `json:"goal_id"`
	Cause     string `json:"cause"`
}

func (m GoalTurnsTick) Type() string { return "goal_turns_tick" }

// GoalPhaseAdvanced records a phase transition in a multi-phase goal,
// setting both ProcessPhase and TurnsRemaining atomically.
type GoalPhaseAdvanced struct {
	FactionID      string `json:"faction_id"`
	GoalID         string `json:"goal_id"`
	ProcessPhase   int    `json:"process_phase"`
	TurnsRemaining int    `json:"turns_remaining"`
	Cause          string `json:"cause"`
}

func (m GoalPhaseAdvanced) Type() string { return "goal_phase_advanced" }

// XPSpent decrements a faction's XP by Amount. Emitted by the stat raise phase.
type XPSpent struct {
	FactionID string `json:"faction_id"`
	Amount    int    `json:"amount"`
	Cause     string `json:"cause"`
}

func (m XPSpent) Type() string { return "xp_spent" }

// StatRaised increments one of a faction's attribute ratings by one and
// recalculates MaxHP. Emitted alongside XPSpent by the stat raise phase.
type StatRaised struct {
	FactionID string      `json:"faction_id"`
	Stat      FactionStat `json:"stat"`
	OldRating int         `json:"old_rating"`
	NewRating int         `json:"new_rating"`
	Cause     string      `json:"cause"`
}

func (m StatRaised) Type() string { return "stat_raised" }

type MovementOrderIssued struct {
	FactionID         string        `json:"faction_id"`
	AssetID           string        `json:"asset_id"`
	Order             MovementOrder `json:"order"`
	Cause             string        `json:"cause"`
	CausedByFactionID string        `json:"caused_by_faction_id"`
}

func (m MovementOrderIssued) Type() string { return "movement_order_issued" }

type MovementOrderProgressed struct {
	FactionID         string           `json:"faction_id"`
	AssetID           string           `json:"asset_id"`
	NewStepIdx        int              `json:"new_step_idx"`
	RegionHex         spatial.RegionHex `json:"region_hex"`
	Cause             string           `json:"cause"`
	CausedByFactionID string           `json:"caused_by_faction_id"`
}

func (m MovementOrderProgressed) Type() string { return "movement_order_progressed" }

type MovementOrderRevised struct {
	FactionID         string        `json:"faction_id"`
	AssetID           string        `json:"asset_id"`
	NewOrder          MovementOrder `json:"new_order"`
	Cause             string        `json:"cause"`
	CausedByFactionID string        `json:"caused_by_faction_id"`
}

func (m MovementOrderRevised) Type() string { return "movement_order_revised" }

type MovementOrderCancelled struct {
	FactionID         string   `json:"faction_id"`
	AssetID           string   `json:"asset_id"`
	StrandedAt        Location `json:"stranded_at"` // current hex at cancellation; WorldID empty if mid-flight
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m MovementOrderCancelled) Type() string { return "movement_order_cancelled" }

type MovementOrderCompleted struct {
	FactionID         string   `json:"faction_id"`
	AssetID           string   `json:"asset_id"`
	FinalLocation     Location `json:"final_location"`
	Cause             string   `json:"cause"`
	CausedByFactionID string   `json:"caused_by_faction_id"`
}

func (m MovementOrderCompleted) Type() string { return "movement_order_completed" }
