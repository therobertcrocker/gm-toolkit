package domain

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
