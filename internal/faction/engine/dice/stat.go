package dice

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const statDieSides = 10

// Engine is what a hooked roll needs from the game. A nil Collector asks no one
// (hooks.md, Key Decisions).
type Engine struct {
	Registry     *hooks.Registry
	Collector    hooks.Collector
	Roller       Roller
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
}

// Side is one party to a roll. Asset is nil when no asset stands there (a faction
// test, a Base defending). Detects marks a side testing to detect stealth.
type Side struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Stat    domain.FactionStat
	Detects bool
}

// Stat is one stat roll: 1d10 plus the side's stat, through the hooks. A zero
// opponent is a roll with no one against it.
func (engine Engine) Stat(phase domain.RollPhase, side, opponent Side, world, action string) hooks.RollResult {
	ctx := hooks.RollContext{
		Phase:             phase,
		Actor:             side.Faction,
		Opponent:          opponent.Faction,
		Attribute:         side.Stat,
		OpponentAttribute: opponent.Stat,
		Asset:             side.Asset,
		OpponentAsset:     opponent.Asset,
		World:             world,
		Action:            action,
		Detection:         side.Detects,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: statDieSides, Modifier: side.Faction.StatScore(side.Stat)}
	return rollWithHooks(ctx, base, engine.Registry, engine.Collector, engine.Roller, side.Faction, engine.FactionState, engine.Rulebook)
}

// Outcome is a rolled contest. Both sides' rolls stay on it for the roll card.
type Outcome struct {
	Attack  hooks.RollResult
	Defense hooks.RollResult
	tie     hooks.TieOutcome
}

// AttackLands is the Attack action's reading: a tie lands, unless a resolver gives
// it to the defender. The counterstrike and the free attack read it too.
func (outcome Outcome) AttackLands() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie != hooks.TieDefenderWins
	}
	return outcome.Attack.Sum > outcome.Defense.Sum
}

// CounterLands is the defender's side of the same reading: a tie lands, unless a
// resolver gives it to the attacker.
func (outcome Outcome) CounterLands() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie != hooks.TieAttackerWins
	}
	return outcome.Defense.Sum > outcome.Attack.Sum
}

// GatePasses is a gate's reading: a tie goes to the defender, unless a resolver
// gives an attack's tie to the attacker.
func (outcome Outcome) GatePasses() bool {
	if outcome.Attack.Sum == outcome.Defense.Sum {
		return outcome.tie == hooks.TieAttackerWins
	}
	return outcome.Attack.Sum > outcome.Defense.Sum
}

// Contest rolls two stat rolls, attacker first, and asks the tie resolvers only on
// the tie of an attack. A test's tie has no resolver.
func (engine Engine) Contest(kind domain.ContestRoll, attacker, defender Side, world, action string) Outcome {
	attackPhase, defensePhase := domain.RollPhaseFactionTest, domain.RollPhaseContested
	if kind == domain.ContestRollAttack {
		attackPhase, defensePhase = domain.RollPhaseAttack, domain.RollPhaseDefense
	}
	attackerOpponent, defenderOpponent := defender, attacker
	// deliberate: Opponent nil, not the same faction, because Actor == Opponent would ask every asset hook on the world twice (overview Decision 9).
	if attacker.Faction.ID == defender.Faction.ID {
		attackerOpponent, defenderOpponent = Side{Stat: defender.Stat}, Side{Stat: attacker.Stat}
	}
	outcome := Outcome{
		Attack:  engine.Stat(attackPhase, attacker, attackerOpponent, world, action),
		Defense: engine.Stat(defensePhase, defender, defenderOpponent, world, action),
	}
	if outcome.Attack.Sum == outcome.Defense.Sum && kind == domain.ContestRollAttack {
		tieContext := hooks.RollContext{Phase: attackPhase, Actor: attacker.Faction, Opponent: attackerOpponent.Faction, Asset: attacker.Asset, World: world, Action: action}
		outcome.tie = ResolveTie(engine.Registry, tieContext, engine.FactionState)
	}
	return outcome
}

// Gate rolls a contest that decides if an ability or effect goes ahead. With roller
// "self" the declaring side attacks and must win. With "target" the target tests to
// keep what it has, and the gate passes when that test fails. Each side's stat comes
// from the contest, by the role it rolls in.
func (engine Engine) Gate(contest domain.Contest, roller domain.ContestRoller, declaring, target Side, world, action string) bool {
	attacker, defender := declaring, target
	if roller == domain.ContestRollerTarget {
		attacker, defender = target, declaring
	}
	attacker.Stat, defender.Stat = contest.AttackerStat, contest.DefenderStat
	passes := engine.Contest(contest.Roll, attacker, defender, world, action).GatePasses()
	if roller == domain.ContestRollerTarget {
		return !passes
	}
	return passes
}
