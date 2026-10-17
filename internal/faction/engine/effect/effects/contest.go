package effects

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// contestPasses rolls an effect's contest gate and reports whether the effect
// applies. With roller "self" the declaring faction attacks and must win; with
// "target" the affected faction tests to keep what it has, and the effect
// applies when that test fails. Ties go to the defender, as in the A-side gate.
func contestPasses(contest domain.EffectContest, declaring, target *domain.Faction, roller domain.Roller) bool {
	// deliberate: rolls the roller directly, not through RollWithHooks, because a reactor or phase effect has no collector to prompt (Effort 3 plan, Decision 12).
	attacker, defender := declaring, target
	if contest.Roller == domain.ContestRollerTarget {
		attacker, defender = target, declaring
	}
	attack := roller.Roll(10) + attacker.StatScore(contest.AttackerStat)
	defense := roller.Roll(10) + defender.StatScore(contest.DefenderStat)
	attackerWins := attack > defense
	if contest.Roller == domain.ContestRollerTarget {
		return !attackerWins
	}
	return attackerWins
}
