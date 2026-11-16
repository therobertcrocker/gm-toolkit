package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

type Side int

const (
	SideParty Side = iota
	SideEnemies
)

func (side Side) String() string {
	if side == SideParty {
		return "party"
	}
	return "enemies"
}

type Combatant struct {
	ID              string
	Side            Side
	Creature        *creature.Creature
	HP              int
	Dying           int
	Wounded         int
	Dead            bool
	Initiative      int
	ActionsLeft     int
	AttacksThisTurn int
	ReactionUsed    bool
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	return &Combatant{ID: source.ID, Side: side, Creature: source, HP: source.Defense.HP}
}

func (combatant *Combatant) Standing() bool {
	return !combatant.Dead && combatant.HP > 0
}
