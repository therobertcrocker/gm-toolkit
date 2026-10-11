package effect

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Handler interface {
	AssetDefinitionID() string
	Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry, roller domain.Roller)
}

type EffectsEngine struct {
	handlers   map[string]Handler
	registrars map[domain.SpecialEffectType]Registrar
	log        *slog.Logger
}

func New(registrars map[domain.SpecialEffectType]Registrar, log *slog.Logger) *EffectsEngine {
	return &EffectsEngine{handlers: make(map[string]Handler), registrars: registrars, log: log}
}

// Register binds a handler to its asset definition. A second handler for the same
// definition is an error, not an overwrite: a definition carrying both a transport
// profile and an effects list would otherwise lose one silently.
func (e *EffectsEngine) Register(handler Handler) error {
	id := handler.AssetDefinitionID()
	if _, exists := e.handlers[id]; exists {
		return fmt.Errorf("effect handler for asset %q already registered", id)
	}
	e.handlers[id] = handler
	return nil
}

// Checking here, not in a separate pass, means no list can be registered unchecked.
func (e *EffectsEngine) RegisterEffectsList(def *domain.AssetDefinition) error {
	for i, entry := range def.Effects {
		if err := e.validateEntry(entry); err != nil {
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
	}
	return e.Register(&effectsListHandler{def: def, registrars: e.registrars})
}

// ApplyAll walks every asset across all factions — factions by ID, then assets by
// ID, so hook registration order is the same every run — looks up the handler for
// that asset's definition ID, and invokes it. Assets whose definitions are missing
// or whose def has no registered handler are silently skipped.
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller, log *slog.Logger) {
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(faction) {
			def := rulebook.Assets[asset.DefinitionID]
			if def == nil {
				continue
			}
			handler, ok := e.handlers[def.ID]
			if !ok {
				continue
			}
			handler.Apply(faction, asset, hookRegistry, roller)
		}
	}
}

// FireTurnStart runs every turn_start entry on faction's active assets, in asset ID
// order, through its registrar's phase route. A direct definition read, not a hook:
// turn start is a phase event the orchestrator raises once per turn.
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		if !asset.Ready || !asset.Maintained {
			continue
		}
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil {
			continue
		}
		for _, entry := range def.Effects {
			if entry.Trigger != domain.TriggerTurnStart {
				continue
			}
			// deliberate: no nil check on Fire, because the rulebook is fixed after NewWithRulebook, so every definition here passed RegisterEffectsList, which rejects a phase-route entry without one.
			mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
				Faction: faction,
				Asset:   asset,
				Def:     def,
				Entry:   entry,
				Roller:  roller,
			})...)
		}
	}
	return mutations
}
