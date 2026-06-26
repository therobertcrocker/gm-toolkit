package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// StepContext carries the ambient inputs and the pipeline-resolved target into a
// terminal effect handler. The pipeline owns target-selection and the resolution
// gate; a handler that is called has, by halt-on-failure, already passed its gate.
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
	RollResult   int // resolved by the roll stage when Step.Roll != nil; 0 otherwise
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
	Collector    action.Collector   // for handlers with a secondary prompt (relocate)
	WorldEngine  *world.WorldEngine // for movement effects (relocate); nil for non-spatial
}

// EffectHandler emits the mutations for one resolved step.
type EffectHandler func(StepContext) ([]domain.Mutation, error)

var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth:    revealStealth,
	domain.EffectCoinDelta:        coinDelta,
	domain.EffectPurchaseDiscount: purchaseDiscount,
	domain.EffectRelocate:         relocate,
}

// Dispatch runs an asset's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations.
// Assets with no ability block, or whose effect has no registered handler, fall
// through to confirmApplied (the not-yet-built effects).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	var mutations []domain.Mutation
	workingAsset := asset
	for _, step := range def.Ability.Steps {
		handler, ok := handlers[step.Effect]
		if !ok {
			return confirmApplied(asset, def, collector)
		}
		target, err := resolveTarget(step, faction, workingAsset, factionState, collector)
		if err != nil {
			return nil, err
		}
		passed, err := passesGate(step, faction, target, roller)
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
			Faction:      faction,
			Asset:        workingAsset,
			Def:          def,
			Step:         step,
			Target:       target,
			RollResult:   resolveRoll(step, roller),
			FactionState: factionState,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
		workingAsset = applyAssetMovedView(workingAsset, stepMutations)
	}
	return mutations, nil
}

// applyAssetMovedView returns a location-updated copy of asset if muts contains an
// AssetMoved targeting it, so later steps in the same ability read the post-move
// location without writing shared state (the real AssetMoved still applies later).
func applyAssetMovedView(asset *domain.Asset, muts []domain.Mutation) *domain.Asset {
	for _, mutation := range muts {
		if moved, ok := mutation.(domain.AssetMoved); ok && moved.AssetID == asset.ID {
			clone := *asset
			clone.Location = moved.ToLocation
			return &clone
		}
	}
	return asset
}

func confirmApplied(asset *domain.Asset, def *domain.AssetDefinition, collector action.Collector) ([]domain.Mutation, error) {
	if _, err := collector.ConfirmAbilityApplied(asset, def); err != nil {
		return nil, err
	}
	return nil, nil
}
