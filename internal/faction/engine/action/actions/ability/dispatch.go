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
	Rolled       dice.StepResult // what the step's die gave; zero for a step with no die
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
	if factionState == nil {
		return &state.FactionState{Factions: map[string]*domain.Faction{}}
	}
	return factionState.Clone()
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
	domain.EffectShelter:          shelter,
}

// Dispatch runs a holder's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations. An
// asset holder's definition is read from the rulebook; a tag has none. The loader
// gives every A asset an ability and every step a known effect.
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
	if abilityDefinition == nil {
		return nil, errors.New("ability: the holder has no ability")
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
		if workingAsset.HasCondition(domain.ConditionSheltered) {
			// deliberate: fizzle, not act from real state, because an earlier asset in this action sheltered this one, and a sheltered asset cannot act.
			return nil, nil
		}
	}
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)

	var mutations []domain.Mutation
	for _, step := range abilityDefinition.Steps {
		handler, ok := handlers[step.Effect]
		if !ok {
			return nil, fmt.Errorf("ability: no handler for effect %q", step.Effect)
		}
		target, err := resolveTarget(step, workingFaction, domain.AbilityHolder{Asset: workingAsset, Tag: holder.Tag}, view, collector, worldEngine, registry, rulebook)
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
			Rolled:       dice.RollStep(step, roller),
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

// holderAbility returns the ability a holder uses: its tag's, or its asset
// definition's.
func holderAbility(holder domain.AbilityHolder, def *domain.AssetDefinition) *domain.AbilityDefinition {
	if holder.Tag != nil {
		return holder.Tag.Ability
	}
	return def.Ability
}
