package goal

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/goals"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/goal/locks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Handler resolves goal logic for a single goal ID.
type Handler interface {
	GoalID() string
	CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) (locks.GoalLock, []domain.Mutation)
	UpdateProgress(actingFaction *domain.Faction, mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook, index *world.Index) []domain.Mutation
}

// GoalEngine evaluates and advances faction goal state.
type GoalEngine struct {
	handlers map[string]Handler
	log      *slog.Logger
}

func New(log *slog.Logger) *GoalEngine {
	engine := &GoalEngine{handlers: make(map[string]Handler), log: log}
	engine.Register(goals.MilitaryConquest{})
	engine.Register(goals.CommercialExpansion{})
	engine.Register(goals.IntelligenceCoup{})
	engine.Register(goals.PlanetarySeizure{})
	engine.Register(goals.ExpandInfluence{})
	engine.Register(goals.BloodTheEnemy{})
	engine.Register(goals.PeaceableKingdom{})
	engine.Register(goals.DestroyTheFoe{})
	engine.Register(goals.InsideEnemyTerritory{})
	engine.Register(goals.InvincibleValor{})
	engine.Register(goals.WealthOfWorlds{})
	engine.Register(goals.ChangeHomeworld{})
	return engine
}

func (ge *GoalEngine) Register(handler Handler) {
	ge.handlers[handler.GoalID()] = handler
}

// CheckLock evaluates the faction's active goal and returns the appropriate lock
// state. Must be called before bookkeeping and action selection each turn.
func (ge *GoalEngine) CheckLock(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, log *slog.Logger) (locks.GoalLock, []domain.Mutation) {
	log.Info("checking goal lock")
	if faction.ActiveGoal == nil {
		return locks.GoalLock{Type: locks.LockNone}, nil
	}
	handler, ok := ge.handlers[faction.ActiveGoal.GoalID]
	if !ok {
		// Data-only goal: present in goals.toml, no Go handler registered. Intentional.
		return locks.GoalLock{Type: locks.LockNone}, nil
	}
	lock, mutations := handler.CheckLock(faction, factionState, rulebook)
	log.Info("goal lock", "type", lock.Type, "mutations", len(mutations))
	return lock, mutations
}

// UpdateProgress inspects the acting faction's mutation list for goal-relevant
// events and returns supplemental mutations if the goal advances or completes.
// Must be called before MutationEngine.Apply so destructive mutations have not
// yet fired.
func (ge *GoalEngine) UpdateProgress(
	actingFactionID string,
	mutations []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
	index *world.Index,
	log *slog.Logger,
) []domain.Mutation {
	actingFaction, ok := factionState.Factions[actingFactionID]
	if !ok || actingFaction.ActiveGoal == nil {
		return nil
	}
	handler, ok := ge.handlers[actingFaction.ActiveGoal.GoalID]
	if !ok {
		// Data-only goal: present in goals.toml, no Go handler registered. Intentional.
		return nil
	}
	return handler.UpdateProgress(actingFaction, mutations, factionState, rulebook, index)
}
