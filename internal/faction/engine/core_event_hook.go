package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// EventHook reacts to mutations the engine is about to apply and may return
// additional mutations to apply alongside them. The natural consumer is the
// Tag Engine. The dispatcher is intentionally not implemented in this
// refactor — see the dispatch site comment in orchestrator.go (RunFactionTurn).
//
// When the dispatcher lands it will iterate registered hooks in registration
// order, append returned mutations to the working list, and recurse with a
// depth bound of 5; on cap trip the engine logs and stops.
type EventHook interface {
	OnMutations(
		mutations []domain.Mutation,
		factionState *state.FactionState,
		rulebook *loader.Rulebook,
	) []domain.Mutation
}
