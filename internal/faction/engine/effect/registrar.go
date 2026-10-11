package effect

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// Registrar is one S-side effect's spec: what the effect accepts in the data, and
// the code that runs it. Triggers keys the accepted sets on the trigger, because
// owner and scope read differently per trigger: an independent set per field would
// accept a pairing the effect never supports. An entry runs through Fire when its
// trigger is a phase event (IsPhaseTrigger) and through Register otherwise; a nil
// function means the effect has no such route.
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
}

type Accepts struct {
	Owners  []domain.EffectOwner
	Scopes  []domain.EffectScope
	Filters []domain.EffectFilterName
	Gates   []domain.EffectGateName
}

type RegisterContext struct {
	Faction  *domain.Faction
	Asset    *domain.Asset
	Def      *domain.AssetDefinition
	Entry    domain.EffectDefinition
	Registry *hooks.Registry
}

type FireContext struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Def     *domain.AssetDefinition
	Entry   domain.EffectDefinition
	Roller  domain.Roller
}

// IsPhaseTrigger reports whether trigger fires from a direct definition read at a
// phase event, not from a hook registered at cycle start. Turn start is a phase
// event, not a mutation; a purchased asset has no hooks until the next cycle.
func IsPhaseTrigger(trigger domain.EffectTrigger) bool {
	return trigger == domain.TriggerTurnStart || trigger == domain.TriggerOnPurchase
}

// effectsListHandler holds a definition's whole [[effects]] list, because the
// engine keys handlers by definition and one definition may declare several effects.
type effectsListHandler struct {
	def        *domain.AssetDefinition
	registrars map[domain.SpecialEffectType]Registrar
}

func (handler *effectsListHandler) AssetDefinitionID() string { return handler.def.ID }

func (handler *effectsListHandler) Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry) {
	for _, entry := range handler.def.Effects {
		if IsPhaseTrigger(entry.Trigger) {
			continue
		}
		// deliberate: no nil check on Register, because RegisterEffectsList rejects a hook-route entry without one.
		handler.registrars[entry.Effect].Register(RegisterContext{
			Faction:  faction,
			Asset:    asset,
			Def:      handler.def,
			Entry:    entry,
			Registry: hookRegistry,
		})
	}
}

func (e *EffectsEngine) validateEntry(entry domain.EffectDefinition) error {
	registrar, ok := e.registrars[entry.Effect]
	if !ok {
		return fmt.Errorf("effect %q has no registrar", entry.Effect)
	}
	accepts, ok := registrar.Triggers[entry.Trigger]
	if !ok {
		return fmt.Errorf("effect %q does not accept trigger %q", entry.Effect, entry.Trigger)
	}
	if IsPhaseTrigger(entry.Trigger) && registrar.Fire == nil {
		return fmt.Errorf("effect %q has no phase route for trigger %q", entry.Effect, entry.Trigger)
	}
	if !IsPhaseTrigger(entry.Trigger) && registrar.Register == nil {
		return fmt.Errorf("effect %q has no hook route for trigger %q", entry.Effect, entry.Trigger)
	}
	if !slices.Contains(accepts.Owners, entry.Filters.Owner) {
		return fmt.Errorf("effect %q does not accept owner %q for trigger %q", entry.Effect, entry.Filters.Owner, entry.Trigger)
	}
	if !slices.Contains(accepts.Scopes, entry.Filters.Scope) {
		return fmt.Errorf("effect %q does not accept scope %q for trigger %q", entry.Effect, entry.Filters.Scope, entry.Trigger)
	}
	for _, filter := range entry.Filters.Used() {
		if !slices.Contains(accepts.Filters, filter) {
			return fmt.Errorf("effect %q does not accept filter %q for trigger %q", entry.Effect, filter, entry.Trigger)
		}
	}
	for _, gate := range entry.Gates.Used() {
		if !slices.Contains(accepts.Gates, gate) {
			return fmt.Errorf("effect %q does not accept gate %q for trigger %q", entry.Effect, gate, entry.Trigger)
		}
	}
	return nil
}
