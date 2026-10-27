package effect

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// Registrar is one S-side effect's spec: what the effect accepts in the data, and
// the code that runs it. Triggers keys the accepted sets on the trigger, because
// owner and scope read differently per trigger: an independent set per field would
// accept a pairing the effect never supports. An entry runs through Matchup when its
// trigger fires inside the Attack action (IsMatchupTrigger), through Fire when it is
// a phase event (IsPhaseTrigger), and through Register otherwise; a nil function
// means the effect has no such route. Validate is optional: it refuses a
// combination of fields that Accepts, which checks each field alone, cannot express.
// Targets lists the assets a quality's entry may act on; only an effect a quality
// can carry sets it.
type Registrar struct {
	Triggers map[domain.EffectTrigger]Accepts
	Register func(RegisterContext)
	Fire     func(FireContext) []domain.Mutation
	Matchup  func(MatchupContext) MatchupOutcome
	Validate func(domain.EffectDefinition, HolderKind) error
	Targets  func(TargetContext) []*domain.Asset
}

// HolderKind is what declared an entry, as engine construction checks it.
type HolderKind int

const (
	HolderAsset HolderKind = iota
	HolderTag
)

type Accepts struct {
	Owners  []domain.EffectOwner
	Scopes  []domain.EffectScope
	Filters []domain.EffectFilterName
	Gates   []domain.EffectGateName
}

// Holder is who declared an [[effects]] entry: an asset through its definition,
// or a faction through a tag (Tag set, Asset and Def nil). A registrar reads its
// holder only through this value, so one registrar serves both. A quality's
// purchase has a Def and no Asset.
type Holder struct {
	Faction *domain.Faction
	Asset   *domain.Asset
	Def     *domain.AssetDefinition
	Tag     *domain.Tag
}

// ID names the source in a hook ID and a budget key: the definition ID, or the
// tag ID.
func (holder Holder) ID() string {
	if holder.Tag != nil {
		return holder.Tag.ID
	}
	return holder.Def.ID
}

// Name is the source's display name, for an offer's Source.
func (holder Holder) Name() string {
	if holder.Tag != nil {
		return holder.Tag.Name
	}
	return holder.Def.Name
}

// Scope is the registry bucket the holder's hooks go in: its asset's, or its
// faction's for a tag.
func (holder Holder) Scope() hooks.Scope {
	if holder.Tag != nil {
		return hooks.FactionScope(holder.Faction.ID)
	}
	return hooks.AssetScope(holder.Faction.ID, holder.Asset.ID)
}

// HookID names a registered hook for diagnostics.
func (holder Holder) HookID(effectName domain.SpecialEffectType) string {
	return string(effectName) + ":" + holder.ID()
}

// RegisterContext carries Roller because a Cat 3 reactor's OnMutations takes none;
// a reactor that rolls dice keeps this one.
type RegisterContext struct {
	Holder   Holder
	Entry    domain.EffectDefinition
	Registry *hooks.Registry
	Roller   domain.Roller
}

// FireContext carries FactionState because a phase effect may reach other
// factions' assets (Panopticon Matrix tests every rival on its world). Asset is
// the subject: the bought asset, a quality's target, or at turn start the
// holder's own asset.
type FireContext struct {
	Holder       Holder
	Asset        *domain.Asset
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
}

// TargetContext is a quality's purchase before it is made: the buyer, the world it
// buys on, and the quality's one entry. Rulebook rides along because a filter such
// as asset_type reads the candidate's definition.
type TargetContext struct {
	Faction  *domain.Faction
	WorldID  string
	Entry    domain.EffectDefinition
	Rulebook *rulebook.Rulebook
}

// MatchupContext is one matchup of the Attack action, seen from one side. Asset is
// that side's fighting asset: its attacker for own_attack, its defender for
// own_defense_win. Holder declared the entry.
type MatchupContext struct {
	Holder        Holder
	Asset         *domain.Asset
	Entry         domain.EffectDefinition
	Opponent      *domain.Faction
	OpponentAsset *domain.Asset
	FactionState  *state.FactionState
}

// MatchupOutcome carries the route's mutations and the decisions the Attack
// action carries out itself, because they need its rolls, prompts and HP tracker.
type MatchupOutcome struct {
	Mutations     []domain.Mutation
	Counterstrike bool
}

// IsMatchupTrigger reports whether trigger fires from the Attack action at each
// matchup: a direct definition read the action raises (overview Decision 21).
func IsMatchupTrigger(trigger domain.EffectTrigger) bool {
	return trigger == domain.TriggerOwnAttack || trigger == domain.TriggerOwnDefenseWin
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

func (handler *effectsListHandler) Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry, roller domain.Roller) {
	for _, entry := range handler.def.Effects {
		if IsPhaseTrigger(entry.Trigger) || IsMatchupTrigger(entry.Trigger) {
			continue
		}
		// deliberate: no nil check on Register, because RegisterEffectsList rejects a hook-route entry without one.
		handler.registrars[entry.Effect].Register(RegisterContext{
			Holder:   Holder{Faction: faction, Asset: asset, Def: handler.def},
			Entry:    entry,
			Registry: hookRegistry,
			Roller:   roller,
		})
	}
}

func (e *EffectsEngine) validateEntry(entry domain.EffectDefinition, holder HolderKind) error {
	registrar, ok := e.registrars[entry.Effect]
	if !ok {
		return fmt.Errorf("effect %q has no registrar", entry.Effect)
	}
	accepts, ok := registrar.Triggers[entry.Trigger]
	if !ok {
		return fmt.Errorf("effect %q does not accept trigger %q", entry.Effect, entry.Trigger)
	}
	switch {
	case IsMatchupTrigger(entry.Trigger):
		if registrar.Matchup == nil {
			return fmt.Errorf("effect %q has no matchup route for trigger %q", entry.Effect, entry.Trigger)
		}
	case IsPhaseTrigger(entry.Trigger):
		if registrar.Fire == nil {
			return fmt.Errorf("effect %q has no phase route for trigger %q", entry.Effect, entry.Trigger)
		}
	default:
		if registrar.Register == nil {
			return fmt.Errorf("effect %q has no hook route for trigger %q", entry.Effect, entry.Trigger)
		}
	}
	if !slices.Contains(accepts.Owners, entry.Filters.Owner) {
		return fmt.Errorf("effect %q does not accept owner %q for trigger %q", entry.Effect, entry.Filters.Owner, entry.Trigger)
	}
	if !slices.Contains(accepts.Scopes, entry.Filters.Scope) {
		return fmt.Errorf("effect %q does not accept scope %q for trigger %q", entry.Effect, entry.Filters.Scope, entry.Trigger)
	}
	// deliberate: scope decides the holder, not a holders list, because self and world measure from the declaring asset and a tag has none (Effort 1 Decision 2).
	if holder == HolderTag && (entry.Filters.Scope == domain.ScopeSelf || entry.Filters.Scope == domain.ScopeWorld) {
		return fmt.Errorf("effect %q: a tag has no asset, so scope %q cannot apply", entry.Effect, entry.Filters.Scope)
	}
	if holder == HolderTag && entry.Gates.OncePer == domain.OncePerAsset {
		return fmt.Errorf("effect %q: a tag has no asset, so once_per %q cannot apply", entry.Effect, domain.OncePerAsset)
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
	if registrar.Validate != nil {
		if err := registrar.Validate(entry, holder); err != nil {
			return fmt.Errorf("effect %q: %w", entry.Effect, err)
		}
	}
	return nil
}
