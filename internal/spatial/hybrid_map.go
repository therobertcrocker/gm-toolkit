package spatial

type HexCoord struct{ Q, R int }

type BoundaryConnection struct {
	From     HexCoord
	ToRegion string
	To       HexCoord
}

type Region struct {
	ID         string
	Name       string
	Hexes      map[HexCoord]bool
	Boundaries []BoundaryConnection
}

type Fragment struct {
	FragmentID     string
	FragmentName   string
	FragTechLevel  int
	FragPopulation int
	Region         string
	Hex            HexCoord
}

func (fragment *Fragment) ID() string      { return fragment.FragmentID }
func (fragment *Fragment) Name() string    { return fragment.FragmentName }
func (fragment *Fragment) TechLevel() int  { return fragment.FragTechLevel }
func (fragment *Fragment) Population() int { return fragment.FragPopulation }

type HybridMap struct {
	regions   map[string]*Region
	fragments map[string]*Fragment
}
