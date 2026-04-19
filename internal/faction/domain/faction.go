package domain

type FactionScale string

const (
	ScaleMinor   FactionScale = "minor"
	ScaleMajor   FactionScale = "major"
	ScaleHegemon FactionScale = "hegemon"
)

type Tag struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Effect      string `toml:"effect"`
}

type Goal struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Difficulty  string `toml:"difficulty"`
}

type Faction struct {
	ID        string       `toml:"id"`
	Name      string       `toml:"name"`
	Scale     FactionScale `toml:"scale"`
	Force     int          `toml:"force"`
	Cunning   int          `toml:"cunning"`
	Wealth    int          `toml:"wealth"`
	CurrentHP int          `toml:"current_hp"`
	MaxHP     int          `toml:"max_hp"`
	Coin      int          `toml:"coin"`
	XP        int          `toml:"xp"`
	Homeworld string       `toml:"homeworld"`
	Tags      []*Tag       `toml:"tags"`
	Goal      *Goal        `toml:"goal"`
	Assets    []*Asset     `toml:"assets"`
}

func RatingsFromScale(s FactionScale) (primary, secondary, tertiary int) {
	switch s {
	case ScaleMajor:
		return 6, 5, 3
	case ScaleHegemon:
		return 8, 7, 5
	default: // minor
		return 4, 3, 1
	}
}

func AssetCountsFromScale(s FactionScale) (primary, other int) {
	switch s {
	case ScaleMajor:
		return 2, 2
	case ScaleHegemon:
		return 4, 4
	default: // minor
		return 1, 1
	}
}

func CalcMaxHP(f *Faction) int {
	return 4 + HPValueForRating(f.Force) + HPValueForRating(f.Cunning) + HPValueForRating(f.Wealth)
}

func HPValueForRating(rating int) int {
	switch rating {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	case 4:
		return 6
	case 5:
		return 9
	case 6:
		return 12
	case 7:
		return 16
	case 8:
		return 20
	default:
		return 0
	}
}
