package goal

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const (
	sideOwn   = "own"
	sideRival = "rival"

	targetStartInput   = "start_input"
	inputTargetFaction = "target_faction"
	groundRival        = "rival"

	amountOne    = "one"
	amountDamage = "damage"
	amountDelta  = "delta"

	eventGovernmentTaken = "government_taken"

	formulaConstant      = "constant"
	formulaHalfProgress  = "half_progress"
	formulaTargetAverage = "target_average"
	formulaRulerAverage  = "ruler_average"
)

// evalContext is one faction's goal read against one batch. factionState is the
// state before the batch applies.
type evalContext struct {
	faction      *domain.Faction
	batch        []domain.Mutation
	factionState *state.FactionState
	rulebook     *rulebook.Rulebook
}

// match is one mutation that passed an event's filters. world is empty for an
// event with no world; mutation is set only where an XP formula reads it.
type match struct {
	amount   int
	world    string
	mutation domain.Mutation
}

// eventSpec is one goal event: the filters and amounts the data may give it, and
// the matcher over a batch. hasWorld says its matches name a world.
type eventSpec struct {
	filters  []string
	amounts  []string
	hasWorld bool
	match    func(ctx evalContext, filters domain.GoalFilters) []match
}

type measureSpec struct {
	filters []string
	count   func(ctx evalContext, filters domain.GoalFilters) int
}

var events = map[string]eventSpec{
	"kill": {
		filters:  []string{domain.GoalFilterOwner, domain.GoalFilterCausedBy, domain.GoalFilterCategory, domain.GoalFilterRatedAbove},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchKill,
	},
	"damage": {
		filters: []string{domain.GoalFilterOwner, domain.GoalFilterCausedBy},
		amounts: []string{amountDamage},
		match:   matchDamage,
	},
	"base_planted": {
		filters:  []string{domain.GoalFilterCausedBy, domain.GoalFilterNewWorld},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchBasePlanted,
	},
	"influence_bought": {
		filters: []string{domain.GoalFilterCausedBy},
		amounts: []string{amountDelta},
		match:   matchInfluenceBought,
	},
	eventGovernmentTaken: {
		filters:  []string{domain.GoalFilterTo},
		amounts:  []string{amountOne},
		hasWorld: true,
		match:    matchGovernmentTaken,
	},
	"faction_destroyed": {
		filters: []string{domain.GoalFilterTarget, domain.GoalFilterCausedBy},
		amounts: []string{amountOne},
		match:   matchFactionDestroyed,
	},
	"action_taken": {
		filters: []string{domain.GoalFilterFaction, domain.GoalFilterAction, domain.GoalFilterActionNot},
		amounts: []string{amountOne},
		match:   matchActionTaken,
	},
}

var measures = map[string]measureSpec{
	"stealthed_assets": {
		filters: []string{domain.GoalFilterGround, domain.GoalFilterExcludeHeldAtStart},
		count:   countStealthedAssets,
	},
}

// sideMatches reads an own/rival filter against the goal's holder. An absent
// filter matches every faction.
func sideMatches(want, factionID, holderID string) bool {
	switch want {
	case sideOwn:
		return factionID == holderID
	case sideRival:
		return factionID != "" && factionID != holderID
	default:
		return true
	}
}
