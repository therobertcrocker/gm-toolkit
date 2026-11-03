package ability

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/mutation"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// StepContext carries the ambient inputs and the pipeline-resolved target into a
// terminal effect handler. The pipeline owns target-selection and the resolution
// gate; a handler that is called has, by halt-on-failure, already passed its gate.
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset           // nil for a tag holder
	Def          *domain.AssetDefinition // nil for a tag holder
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
	RollResult   int // resolved by the roll stage when Step.Roll != nil; 0 otherwise
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
	Collector    action.Collector   // for handlers with a secondary prompt (relocate)
	WorldEngine  *world.WorldEngine // for movement effects (relocate); nil for non-spatial
	Registry     *hooks.Registry    // for purchase_enabler's asset-cost chain
}

// EffectHandler emits the mutations for one resolved step.
type EffectHandler func(StepContext) ([]domain.Mutation, error)

// NewView copies every faction, so Dispatch can apply each step's mutations and a
// later step, or a later asset in the same action, reads what changed. Rolls spend
// budgets on the real factions, not on the view (Effort 3 Decision 3).
func NewView(factionState *state.FactionState) *state.FactionState {
	view := &state.FactionState{Factions: map[string]*domain.Faction{}}
	if factionState == nil {
		return view
	}
	*view = *factionState
	view.Factions = make(map[string]*domain.Faction, len(factionState.Factions))
	for id, faction := range factionState.Factions {
		view.Factions[id] = faction.Clone()
	}
	return view
}

var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
	domain.EffectRelocate:         relocateAssets,
	domain.EffectPurchaseEnabler:  purchaseEnabler,
	domain.EffectDisableAsset:     disableAsset,
	domain.EffectBlockAttack:      blockAttack,
	domain.EffectLevyCoin:         levyCoin,
	domain.EffectRaiseTechLevel:   raiseTechLevel,
}

// Dispatch runs a holder's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations. An
// asset holder's definition is read from the rulebook; a tag has none. Assets with
// no ability block, or whose effect has no registered handler, fall through to
// confirmApplied (the not-yet-built effects). The loader lets a tag carry only
// built effects.
func Dispatch(
	faction *domain.Faction,
	holder domain.AbilityHolder,
	collector action.Collector,
	roller dice.Roller,
	registry *hooks.Registry,
	factionState *state.FactionState,
	view *state.FactionState,
	rulebook *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
	asset := holder.Asset
	def, err := holderDefinition(holder, rulebook)
	if err != nil {
		return nil, err
	}
	abilityDefinition := holderAbility(holder, def)
	if abilityDefinition == nil || len(abilityDefinition.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	// Pre-flight: if any step's effect is not yet built, the whole ability is a
	// not-yet-built no-op. Bail before executing any step, so a built early step
	// never half-applies (moves an asset, spends Coin) only to be discarded when a
	// later unbuilt step short-circuits to confirmApplied.
	for _, step := range abilityDefinition.Steps {
		if _, ok := handlers[step.Effect]; !ok {
			return confirmApplied(asset, def, collector)
		}
	}
	workingFaction := view.Factions[faction.ID]
	if workingFaction == nil {
		return nil, fmt.Errorf("ability: acting faction %s not in the view", faction.ID)
	}
	var workingAsset *domain.Asset
	if asset != nil {
		workingAsset = workingFaction.Assets[asset.ID]
		if workingAsset == nil {
			// deliberate: fizzle, not act from real state, because an asset missing from the view was removed by an earlier asset in this action.
			return nil, nil
		}
	}
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)

	var mutations []domain.Mutation
	for _, step := range abilityDefinition.Steps {
		handler := handlers[step.Effect]
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector, worldEngine, registry)
		// deliberate: fizzle this asset, not fail the action, because an earlier asset in the same action may have used up the targets and its result must stand (Effort 3 Decision Record 1).
		if errors.Is(err, errNoTargets) {
			break
		}
		if err != nil {
			return nil, err
		}
		// deliberate: the real faction, not workingFaction, because a roll hook's budget spent on the view would be lost (Effort 3 plan, Decision 12).
		passed, err := passesGate(step, faction, workingAsset, target, registry, collector, roller, factionState, rulebook)
		if err != nil {
			return nil, err
		}
		if !passed {
			// deliberate: halt the chain on a failed gate, not collect-and-continue,
			// because a missed contest fizzles the rest of the ability. Multi-step
			// abilities must therefore author ungated steps (relocate, roll-table)
			// before gated ones (Effort 1 Plan, Decision 1).
			break
		}
		stepMutations, err := handler(StepContext{
			Faction:      workingFaction,
			Asset:        workingAsset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: view,
			Rulebook:     rulebook,
			Collector:    collector,
			WorldEngine:  worldEngine,
			Registry:     registry,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
		// deliberate: drop the projection's apply error, not return it, because Output() applies the same mutations against real state, where a genuine miss still surfaces.
		_ = mutationEngine.Apply(view, stepMutations, discardLog)
		if asset == nil {
			continue
		}
		if updated, ok := workingFaction.Assets[asset.ID]; ok {
			workingAsset = updated
		}
	}
	return mutations, nil
}

// holderAbility returns the ability a holder uses: its tag's, or its asset
// definition's.
// holderDefinition returns an asset holder's definition from the rulebook, or nil
// for a tag, which has none.
func holderDefinition(holder domain.AbilityHolder, rulebook *rulebook.Rulebook) (*domain.AssetDefinition, error) {
	if holder.Asset == nil {
		return nil, nil
	}
	if rulebook != nil {
		if def, ok := rulebook.Assets[holder.Asset.DefinitionID]; ok {
			return def, nil
		}
	}
	return nil, fmt.Errorf("ability: definition not found: %s", holder.Asset.DefinitionID)
}

func holderAbility(holder domain.AbilityHolder, def *domain.AssetDefinition) *domain.AbilityDefinition {
	if holder.Tag != nil {
		return holder.Tag.Ability
	}
	return def.Ability
}

func confirmApplied(asset *domain.Asset, def *domain.AssetDefinition, collector action.Collector) ([]domain.Mutation, error) {
	if _, err := collector.ConfirmAbilityApplied(asset, def); err != nil {
		return nil, err
	}
	return nil, nil
}
