package spatial

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
