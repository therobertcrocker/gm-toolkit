package spatial

import "errors"

type SpatialMap interface {
	Location(id string) (Location, bool)
}

type Location interface {
	ID() string
	Name() string
	TechLevel() int
	Population() int
}

type RegionLocation interface {
	Location
	Coords() (q, r int)
	RegionID() string
	RegionHex() RegionHex
}

var (
	ErrUnknownWorld   = errors.New("spatial: unknown world")
	ErrNoPath         = errors.New("spatial: no path")
	ErrNotImplemented = errors.New("spatial: not implemented")
	ErrInvalidCost    = errors.New("spatial: invalid cost")
)

var (
	_ SpatialMap     = (*RegionMap)(nil)
	_ RegionLocation = (*World)(nil)
)
