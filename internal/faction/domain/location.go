package domain

import "github.com/therobertcrocker/gm-toolkit/internal/spatial"

type Location struct {
	WorldID   string            `toml:"world_id"`
	RegionHex spatial.RegionHex `toml:"region_hex"`
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
}
