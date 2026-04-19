package domain

type factionScale string

const (
	ScaleMinor   factionScale = "minor"
	ScaleMajor   factionScale = "major"
	ScaleHegemon factionScale = "hegemon"
)

type Tag struct {
	ID          string
	Name        string
	Description string
	Effect      string
}

type Goal struct {
	ID          string
	Name        string
	Description string
	Difficulty  string // named label: "low", "moderate", or a formula e.g. "half_assets_destroyed"
}

type Faction struct {
	ID        string
	Name      string
	Scale     factionScale
	Force     int
	Cunning   int
	Wealth    int
	CurrentHP int
	MaxHP     int
	Coin      int
	XP        int
	Homeworld string
	Tags      []*Tag
	Goal      *Goal
	Assets    []*Asset
}

func ScaleFromString(s string) factionScale {
	switch s {
	case "minor":
		return ScaleMinor
	case "major":
		return ScaleMajor
	case "hegemon":
		return ScaleHegemon
	default:
		return "BAD_SCALE"
	}
}

func RatingsFromScale(s string) (primary, secondary, tertiary int) {
	switch s {
	case "major":
		return 6, 5, 3
	case "hegemon":
		return 8, 7, 5
	default: // minor
		return 4, 3, 1
	}
}

func AssetCountsFromScale(s string) (primary, other int) {
	switch s {
	case "major":
		return 2, 2
	case "hegemon":
		return 4, 4
	default: // minor
		return 1, 1
	}
}

func CalcMaxHP(f *Faction) int {
	return 4 + HPValueForRating(f.Force) + HPValueForRating(f.Cunning) + HPValueForRating(f.Wealth)
}

func HPValueForRating(rating int) int {
	values := map[int]int{1: 1, 2: 2, 3: 4, 4: 6, 5: 9, 6: 12, 7: 16, 8: 20}
	if v, ok := values[rating]; ok {
		return v
	}
	return 0
}
