package domain

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Location struct {
	WorldID   string            `toml:"world_id"`
	RegionHex spatial.RegionHex `toml:"region_hex"`
}

// MarshalTOML writes a location as one inline table, so a saved state.toml has no
// location headers.
func (location Location) MarshalTOML() ([]byte, error) {
	regionHex, err := location.RegionHex.MarshalTOML()
	if err != nil {
		return nil, err
	}
	// deliberate: Go %q quoting, not a TOML escaper, because IDs are slugs; TestSaveWritesLocationsInline pins quotes and backslashes
	return fmt.Appendf(nil, "{ world_id = %q, region_hex = %s }", location.WorldID, regionHex), nil
}

// IsInFlight reports whether loc represents an asset mid-transit between worlds.
// An in-flight asset has no settled WorldID; only its current RegionHex is meaningful.
func IsInFlight(loc Location) bool {
	return loc.WorldID == ""
}

type MovementOrder struct {
	AssetID       string              `toml:"asset_id" json:"asset_id"`
	Origin        Location            `toml:"origin" json:"origin"`
	Destination   Location            `toml:"destination" json:"destination"`
	Path          []spatial.RegionHex `toml:"path" json:"path"`
	StepIdx       int                 `toml:"step_idx" json:"step_idx"`
	DriftRating   int                 `toml:"drift_rating" json:"drift_rating"`
	CargoAssetIDs []string            `toml:"cargo_asset_ids,omitempty" json:"cargo_asset_ids,omitempty"`
	// Banked is the movement points carried toward the next step: a Drift jump
	// that costs more than a turn's speed is paid over several turns (A.001.5
	// Effort 4 Decision 19).
	Banked int `toml:"banked,omitempty" json:"banked,omitempty"`
	// RunHot marks an order run hot: hot-only speed effects apply, and hot_move
	// effects fire each turn it moves (Aether Burners).
	RunHot bool `toml:"run_hot,omitempty" json:"run_hot,omitempty"`
	// LockdownPaid marks an order that has paid its way through a Transport
	// Lockdown and waited its turn; it lands without paying again (Effort 5
	// Decision 14). A revised order is a new order and pays again.
	LockdownPaid bool `toml:"lockdown_paid,omitempty" json:"lockdown_paid,omitempty"`
	// Held marks an order held at its destination's edge. Its Banked points pay
	// only for landing there, so a revise drops them (A.001.5 Effort 4 Decision
	// Record 6).
	Held bool `toml:"held,omitempty" json:"held,omitempty"`
}
