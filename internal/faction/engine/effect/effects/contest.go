package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// contestRoll is one effect contest, seen from the effect: the declaring faction
// and its asset, the affected faction and its asset (nil when no asset stands for
// it: Lobbyists' requester, or an asset bought in this batch), the world, and what
// the hooked rolls read. collector is nil until the roll-card follow-on row gives
// each contest one (overview Decision 2). declaringDetects marks the declaring
// side's roll as a detection roll (Effort 3 Decision 5).
type contestRoll struct {
	contest          domain.EffectContest
	declaringDetects bool
	declaring        *domain.Faction
	declaringAsset   *domain.Asset
	target           *domain.Faction
	targetAsset      *domain.Asset
	world            string
	collector        hooks.Collector
	registry         *hooks.Registry
	factionState     *state.FactionState
	rulebook         *rulebook.Rulebook
	roller           domain.Roller
}

// contestPasses rolls an effect's contest gate through the roll hooks and reports
// whether the effect applies. With roller "self" the declaring faction attacks
// and must win; with "target" the affected faction tests to keep what it has, and
// the effect applies when that test fails. Ties go to the defender, as in the
// A-side gate. roll "attack" rolls an attack and its defense; roll "test" a
// faction test and its contested roll, as the A-side gate does (Effort 3
// Decision 4).
func contestPasses(roll contestRoll) bool {
	attacker, attackerAsset, defender, defenderAsset := roll.declaring, roll.declaringAsset, roll.target, roll.targetAsset
	if roll.contest.Roller == domain.ContestRollerTarget {
		attacker, attackerAsset, defender, defenderAsset = roll.target, roll.targetAsset, roll.declaring, roll.declaringAsset
	}
	attackPhase, defensePhase := hooks.PhaseFactionTest, hooks.PhaseContested
	if roll.contest.Roll == domain.ContestRollAttack {
		attackPhase, defensePhase = hooks.PhaseAttack, hooks.PhaseDefense
	}
	attackerStat, defenderStat := roll.contest.AttackerStat, roll.contest.DefenderStat
	attack := roll.side(attackPhase, attacker, defender, attackerAsset, defenderAsset, attackerStat, defenderStat)
	defense := roll.side(defensePhase, defender, attacker, defenderAsset, attackerAsset, defenderStat, attackerStat)
	attackerWins := attack > defense
	if roll.contest.Roller == domain.ContestRollerTarget {
		return !attackerWins
	}
	return attackerWins
}

// side rolls one side of the contest: 1d10 plus actor's stat, with the stat the
// other side rolls as OpponentAttribute (Effort 2 Decision 3).
func (roll contestRoll) side(phase hooks.RollPhase, actor, opponent *domain.Faction, asset, opponentAsset *domain.Asset, stat, opponentStat domain.FactionStat) int {
	ctx := hooks.RollContext{
		Phase:             phase,
		Actor:             actor,
		Opponent:          opponent,
		Attribute:         string(stat),
		OpponentAttribute: string(opponentStat),
		Asset:             asset,
		OpponentAsset:     opponentAsset,
		World:             roll.world,
		Detection:         roll.declaringDetects && actor == roll.declaring,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: actor.StatScore(stat)}
	return dispatch.RollWithHooks(ctx, base, roll.registry, roll.collector, roll.roller, actor, roll.factionState, roll.rulebook).Sum
}
