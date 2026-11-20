package engine

import (
	"maps"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

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
	ID                string
	Side              Side
	Creature          *creature.Creature
	HP                int
	Dying             int
	Wounded           int
	Dead              bool
	Initiative        int
	ActionsLeft       int
	StrikeOnlyActions int
	AttacksThisTurn   int
	ReactionUsed      bool
	HeroPoints        int
	TempHP            int
	conditions        []*conditionInstance
	persistent        []*persistentInstance
	defenseChanges    []*defenseChange
	adaptiveType      string
	uses              map[string]int
	roundUses         map[string]int
	recharging        map[string]int
	pools             map[string]int
	flourished        bool
	lastPenalty       int
	extraReactions    map[string]int
	downed            bool
	rerolls           []*rerollEffect
	bonuses           []*grantedBonus
	recoveryCaps      []*recoveryCap
	sustained         map[string]*sustainedEffect
	grants            []*grantedTrait
	loyalRound        int
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	pools := map[string]int{}
	maps.Copy(pools, source.Pools)
	return &Combatant{
		ID:         source.ID,
		Side:       side,
		Creature:   source,
		HP:         source.Defense.HP,
		HeroPoints: source.HeroPoints,
		uses:       map[string]int{},
		roundUses:  map[string]int{},
		recharging: map[string]int{},
		pools:      pools,
		sustained:  map[string]*sustainedEffect{},
	}
}

func (combatant *Combatant) Standing() bool {
	return !combatant.Dead && combatant.HP > 0
}
