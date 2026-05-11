package spatial

import "errors"

type SpatialMap interface {
	Location(id string) (Location, bool)
	Distance(fromID, toID string, crossingCost int) (int, error)
}

type Location interface {
	ID() string
	Name() string
	TechLevel() int
	Population() int
}

var (
	ErrUnknownWorld   = errors.New("spatial: unknown world")
	ErrNoPath         = errors.New("spatial: no path")
	ErrNotImplemented = errors.New("spatial: not implemented")
	ErrInvalidCost    = errors.New("spatial: invalid cost")
)

var (
	_ SpatialMap = (*HybridMap)(nil)
	_ SpatialMap = (*HexMap)(nil)
	_ SpatialMap = (*GraphMap)(nil)
	_ Location   = (*World)(nil)
)
