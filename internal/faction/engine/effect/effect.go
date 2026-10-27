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
		if err := e.validateEntry(entry, HolderAsset); err != nil {
			return fmt.Errorf("asset %q effects[%d]: %w", def.ID, i, err)
		}
	}
	if def.Quality && (len(def.Effects) != 1 || e.registrars[def.Effects[0].Effect].Targets == nil) {
		return fmt.Errorf("asset %q: a quality needs one effect with a Targets route", def.ID)
	}
	return e.Register(&effectsListHandler{def: def, registrars: e.registrars})
}

// CheckTagEffects checks a tag's [[effects]] entries for a faction holder. A tag
// registers no handler: ApplyAll and the phase routes read its entries from the
// rulebook, whose every tag NewWithRulebook has checked.
func (e *EffectsEngine) CheckTagEffects(tag *domain.Tag) error {
	for i, entry := range tag.Effects {
		if err := e.validateEntry(entry, HolderTag); err != nil {
			return fmt.Errorf("tag %q effects[%d]: %w", tag.ID, i, err)
		}
	}
	return nil
}

// ApplyAll walks every faction by ID, then its assets by ID, then the tags it
// holds, so hook registration order is the same every run. An asset invokes the
// handler for its definition; a tag registers its hook-route entries at its
// faction's scope. A missing definition, handler or tag registers nothing.
func (e *EffectsEngine) ApplyAll(factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller, log *slog.Logger) {
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		faction := factionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(faction) {
			e.ApplyAsset(faction, asset, rulebook, hookRegistry, roller)
		}
		for _, tag := range heldTags(faction, rulebook) {
			for _, entry := range tag.Effects {
				if !IsHookTrigger(entry.Trigger) {
					continue
				}
				// deliberate: no nil check on Register, because NewWithRulebook ran CheckTagEffects on every tag, which rejects a hook-route entry without one.
				e.registrars[entry.Effect].Register(RegisterContext{
					Holder:   Holder{Faction: faction, Tag: tag},
					Entry:    entry,
					Registry: hookRegistry,
					Roller:   roller,
				})
			}
		}
	}
}

// heldTags returns the rulebook tags faction holds, in the order it holds them. A
// held tag the rulebook does not know is skipped, as an asset with no definition
// is (Effort 1 Decision 10).
func heldTags(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Tag {
	var tags []*domain.Tag
	for _, tagID := range faction.Tags {
		if tag := rulebook.Tags[tagID]; tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
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
// order, then on the tags it holds, through each registrar's phase route. A direct
// definition read, not a hook: turn start is a phase event the orchestrator raises
// once per turn.
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
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerTurnStart, asset, factionState, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerTurnStart, nil, factionState, roller)...)
	}
	return mutations
}

// fireEntries runs holder's entries for trigger through each registrar's phase
// route. subject is the asset the event is about: the bought asset at purchase, the
// holder's own asset at turn start, nil for a tag at turn start.
func (e *EffectsEngine) fireEntries(holder Holder, entries []domain.EffectDefinition, trigger domain.EffectTrigger, subject *domain.Asset, factionState *state.FactionState, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	for _, entry := range entries {
		if entry.Trigger != trigger {
			continue
		}
		// deliberate: no nil check on Fire, because the rulebook is fixed after NewWithRulebook, which ran RegisterEffectsList on every definition and CheckTagEffects on every tag; both reject a phase-route entry without one.
		mutations = append(mutations, e.registrars[entry.Effect].Fire(FireContext{
			Holder:       holder,
			Asset:        subject,
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
		})...)
	}
	return mutations
}

// FirePurchase runs every on_purchase entry of a just-bought asset, then of each tag
// the buyer holds, through its registrar's phase route. A direct definition read,
// not a hook: the asset was bought mid-turn, so it has no hooks until the next
// cycle. No active check: a bought asset is never ready, and its purchase effect is
// part of what was bought.
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	if def := rulebook.Assets[asset.DefinitionID]; def != nil {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerOnPurchase, asset, factionState, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerOnPurchase, asset, factionState, roller)...)
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
// FirePurchase never sees it. The holder has no asset; FireContext.Asset is the target.
func (e *EffectsEngine) FireQuality(faction *domain.Faction, def *domain.AssetDefinition, target *domain.Asset, factionState *state.FactionState) []domain.Mutation {
	entry := def.Effects[0]
	// deliberate: no roller, because no quality effect rolls; give FireQuality one when one does.
	return e.registrars[entry.Effect].Fire(FireContext{
		Holder:       Holder{Faction: faction, Def: def},
		Asset:        target,
		Entry:        entry,
		FactionState: factionState,
	})
}

// FireMatchup runs the entries for trigger of the fighting asset's definition
// (ctx.Holder on the way in), then of each tag its faction holds, through each
// registrar's matchup route. A direct definition read, raised by the Attack
// action, which owns the matchup.
func (e *EffectsEngine) FireMatchup(trigger domain.EffectTrigger, ctx MatchupContext, rulebook *rulebook.Rulebook) MatchupOutcome {
	var outcome MatchupOutcome
	fire := func(holder Holder, entries []domain.EffectDefinition) {
		for _, entry := range entries {
			if entry.Trigger != trigger {
				continue
			}
			ctx.Holder, ctx.Entry = holder, entry
			// deliberate: no nil check on Matchup, because RegisterEffectsList and CheckTagEffects reject a matchup-route entry without one.
			result := e.registrars[entry.Effect].Matchup(ctx)
			outcome.Mutations = append(outcome.Mutations, result.Mutations...)
			outcome.Counterstrike = outcome.Counterstrike || result.Counterstrike
		}
	}
	fire(ctx.Holder, ctx.Holder.Def.Effects)
	faction := ctx.Holder.Faction
	for _, tag := range heldTags(faction, rulebook) {
		fire(Holder{Faction: faction, Tag: tag}, tag.Effects)
	}
	return outcome
}
