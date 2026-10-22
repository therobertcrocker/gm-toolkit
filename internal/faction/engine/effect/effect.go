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
	if def.Quality && (len(def.Effects) != 1 || e.registrars[def.Effects[0].Effect].Targets == nil) {
		return fmt.Errorf("asset %q: a quality needs one effect with a Targets route", def.ID)
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
			e.ApplyAsset(faction, asset, rulebook, hookRegistry, roller)
		}
	}
}

// ApplyAsset registers one asset's hooks under faction. A missing definition or
// handler registers nothing.
func (e *EffectsEngine) ApplyAsset(faction *domain.Faction, asset *domain.Asset, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller) {
	def := rulebook.Assets[asset.DefinitionID]
	if def == nil {
		return
	}
	handler, ok := e.handlers[def.ID]
	if !ok {
		return
	}
	handler.Apply(faction, asset, hookRegistry, roller)
}

// FireTurnStart runs every turn_start entry on faction's active assets, in asset ID
// order, through its registrar's phase route. A direct definition read, not a hook:
// turn start is a phase event the orchestrator raises once per turn.
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	for _, asset := range domain.SortedAssets(faction) {
		if !asset.Active() {
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
				Faction:      faction,
				Asset:        asset,
				Def:          def,
				Entry:        entry,
				Roller:       roller,
				FactionState: factionState,
			})...)
		}
	}
	return mutations
}

// FirePurchase runs every on_purchase entry of a just-bought asset through its
// registrar's phase route. A direct definition read, not a hook: the asset was
// bought mid-turn, so it has no hooks until the next cycle. No active check: a
// bought asset is never ready, and its purchase effect is part of what was bought.
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	def := rulebook.Assets[asset.DefinitionID]
	if def == nil {
		return nil
	}
	var mutations []domain.Mutation
	for _, entry := range def.Effects {
		if entry.Trigger != domain.TriggerOnPurchase {
			continue
		}
		// deliberate: no nil check on Fire, because RegisterEffectsList rejects a phase-route entry without one.
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
		})...)
	}
	return mutations
}

// QualityTargets lists the assets faction holds on worldID that quality def may act
// on, as its registrar reads the entry's filters. A definition that is not a quality
// has none.
func (e *EffectsEngine) QualityTargets(def *domain.AssetDefinition, faction *domain.Faction, worldID string, rulebook *rulebook.Rulebook) []*domain.Asset {
	if !def.Quality || len(def.Effects) != 1 {
		return nil
	}
	entry := def.Effects[0]
	registrar, ok := e.registrars[entry.Effect]
	if !ok || registrar.Targets == nil {
		return nil
	}
	return registrar.Targets(TargetContext{Faction: faction, WorldID: worldID, Entry: entry, Rulebook: rulebook})
}

// FireQuality runs quality def's one entry on target through its phase route. A
// direct definition read raised by Buy Asset: a quality adds no asset, so
// FirePurchase never sees it. FireContext.Asset is the target, not a declaring asset.
func (e *EffectsEngine) FireQuality(faction *domain.Faction, def *domain.AssetDefinition, target *domain.Asset, factionState *state.FactionState) []domain.Mutation {
	entry := def.Effects[0]
	// deliberate: no roller, because no quality effect rolls; give FireQuality one when one does.
	return e.registrars[entry.Effect].Fire(FireContext{
		Faction:      faction,
		Asset:        target,
		Def:          def,
		Entry:        entry,
		FactionState: factionState,
	})
}

// FireMatchup runs the declaring asset's entries for trigger through each
// registrar's matchup route. A direct definition read, raised by the Attack
// action, which owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext) MatchupOutcome {
	var outcome MatchupOutcome
	for _, entry := range ctx.Def.Effects {
		if entry.Trigger != trigger {
			continue
		}
		ctx.Entry = entry
		// deliberate: no nil check on Matchup, because RegisterEffectsList rejects a matchup-route entry without one.
		result := e.registrars[entry.Effect].Matchup(ctx)
		outcome.Mutations = append(outcome.Mutations, result.Mutations...)
		outcome.Counterstrike = outcome.Counterstrike || result.Counterstrike
	}
	return outcome
}
