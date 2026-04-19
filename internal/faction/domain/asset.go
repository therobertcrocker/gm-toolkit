package domain

type FactionStat string

const (
	StatForce   FactionStat = "Force"
	StatCunning FactionStat = "Cunning"
	StatWealth  FactionStat = "Wealth"
)

type AssetType string

const (
	TypeMilitaryUnit      AssetType = "Military Unit"
	TypeSpecialForces     AssetType = "Special Forces"
	TypeFacility          AssetType = "Facility"
	TypeStarship          AssetType = "Starship"
	TypeTactic            AssetType = "Tactic"
	TypeLogisticsFacility AssetType = "Logistics Facility"
	TypeSpecial           AssetType = "Special"
)

type AssetFlag string

const (
	FlagPermission AssetFlag = "P"
	FlagAction     AssetFlag = "A"
	FlagSpecial    AssetFlag = "S"
)

type DiceRoll struct {
	NumDice  int
	Sides    int
	Modifier int
}

type AttackProfile struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
	Damage       DiceRoll
}

type AssetDefinition struct {
	ID          string
	Name        string
	Category    FactionStat
	MinRating   int
	HP          int
	Cost        int
	TechLevel   int
	Type        AssetType
	Attack      *AttackProfile
	Counter     *DiceRoll
	Flags       []AssetFlag
	Description string
}

type Asset struct {
	ID           string `toml:"id"`
	DefinitionID string `toml:"definition_id"`
	OwnerID      string `toml:"owner_id"`
	Location     string `toml:"location"`
	CurrentHP    int    `toml:"current_hp"`
	Stealthy     bool   `toml:"stealthy"`
	Ready        bool   `toml:"ready"`
	Maintained   bool   `toml:"maintained"`
}
