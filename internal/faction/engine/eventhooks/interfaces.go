package eventhooks

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RollModifier offers pre-roll dice pool modifications (Cat 1). The collector
// selects which offers to apply; unchosen offers are discarded.
type RollModifier interface {
	OfferModifiers(ctx RollContext, factionState *state.FactionState, rulebook *rulebook.Rulebook) []ModifierOffer
}

// RollResultHook fires after dice are rolled and may instruct the dispatcher
// to reroll specific dice (Cat 2). Elective directives are confirmed by the
// collector; non-elective directives apply automatically.
type RollResultHook interface {
	OnRollResult(ctx RollContext, result RollResult, factionState *state.FactionState, rulebook *rulebook.Rulebook) RerollDirective
}

// MutationReactor responds to the completed mutation slice for a turn phase
// and may return additional mutations to append (Cat 3). The dispatcher
// recurses with a depth bound of 5; on cap trip it logs and stops.
// (Replaces the soon-renamed EventHook interface — rename lands in Phase 3a.)
type MutationReactor interface {
	OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation
}

// TieResolver decides the outcome when attack and defense rolls are equal
// (Cat 5). If multiple resolvers are registered, the first registration wins.
type TieResolver interface {
	ResolveTie(ctx RollContext, factionState *state.FactionState) TieOutcome
}
