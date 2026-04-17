package domain

type Tag struct {
	Name        string
	Description string
	Effect      string
}

type Goal struct {
	Name        string
	Description string
	Difficulty  int
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
