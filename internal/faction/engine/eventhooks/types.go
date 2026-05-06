package eventhooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// RollPhase identifies which dice roll in a faction action is being hooked.
type RollPhase int

const (
	PhaseAttack      RollPhase = iota
	PhaseDefense
	PhaseFactionTest
	PhaseContested
)

// RollContext carries the full context of a roll being processed by hooks.
type RollContext struct {
	Phase     RollPhase
	Actor     *domain.Faction
	Opponent  *domain.Faction
	Attribute string
	Asset     *domain.Asset
	World     string
}

// RollResult holds the per-die breakdown of a completed roll.
type RollResult struct {
	Dice     []int
	Modifier int
	Sum      int
}

// RollState is the mutable pre-roll view passed to ModifierOffer.Apply.
// Apply calls AddDie to expand the dice pool before any dice are rolled.
// (Populated with dispatch logic in Phase 3b.)
type RollState struct {
	extraDice []int
}

// AddDie queues an additional die with the given number of sides for the
// upcoming roll.
func (rollState *RollState) AddDie(sides int) {
	rollState.extraDice = append(rollState.extraDice, sides)
}

// ExtraDice returns the queued extra dice; read by RollWithHooks in Phase 3b.
func (rollState *RollState) ExtraDice() []int {
	return rollState.extraDice
}

// ModifierOffer is a pre-roll dice pool modification presented to the collector
// for selection. Unchosen offers are discarded.
type ModifierOffer struct {
	Source      string
	Description string
	BudgetKey   string           // namespaced, e.g. "tag:Warlike"; empty = unlimited
	Apply       func(*RollState) // expands the dice pool before rolling
}

// RerollDirective instructs the dispatcher to reroll specific dice after the
// initial roll. Elective directives require collector confirmation.
type RerollDirective struct {
	Source    string
	Indices   []int  // indices into RollResult.Dice to reroll
	BudgetKey string // namespaced; empty = unlimited
	Elective  bool   // if true, collector must confirm before applying
}

// TieOutcome is the result of a TieResolver call.
type TieOutcome int

const (
	TieStandard     TieOutcome = iota // fall back to default >= comparison
	TieAttackerWins                   // attacker wins ties unconditionally
	TieDefenderWins                   // defender wins ties unconditionally
)
