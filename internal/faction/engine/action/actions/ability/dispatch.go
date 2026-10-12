package ability

import (
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
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
	domain.EffectPurchaseEnabler:  purchaseEnabler,
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
	registry *hooks.Registry,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
	worldEngine *world.WorldEngine,
) ([]domain.Mutation, error) {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	// Pre-flight: if any step's effect is not yet built, the whole ability is a
	// not-yet-built no-op. Bail before executing any step, so a built early step
	// never half-applies (moves an asset, spends Coin) only to be discarded when a
	// later unbuilt step short-circuits to confirmApplied.
	for _, step := range def.Ability.Steps {
		if _, ok := handlers[step.Effect]; !ok {
			return confirmApplied(asset, def, collector)
		}
	}
	var mutations []domain.Mutation

	// Cross-step working view: after each step, re-project the acting faction's own
	// post-step state (Location, Coin, …) onto a clone via the same mutation.Apply the
	// real Output() runs — so a later step reads what an earlier step changed, without a
	// parallel projection that could drift from the engine. The clone, not shared state,
	// is what each step sees.
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)
	workingFaction := faction.Clone()
	workingState := &state.FactionState{Factions: map[string]*domain.Faction{faction.ID: workingFaction}}
	workingAsset := workingFaction.Assets[asset.ID]
	if workingAsset == nil {
		workingAsset = asset.Clone()
	}

	for _, step := range def.Ability.Steps {
		handler := handlers[step.Effect]
		target, err := resolveTarget(step, workingFaction, workingAsset, factionState, collector)
		if err != nil {
			return nil, err
		}
		// deliberate: the real faction, not workingFaction, because a roll hook's budget spent on the clone would be lost (Effort 3 plan, Decision 12).
		passed, err := passesGate(step, faction, workingAsset, target, registry, collector, roller, factionState, rb)
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
			FactionState: factionState,
			Rulebook:     rb,
			Collector:    collector,
			WorldEngine:  worldEngine,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
		// deliberate: drop the projection's apply error — workingState holds only the
		// acting faction, so any other-faction mutation (e.g. a reveal clearing a target
		// faction's stealth) is an expected miss; a genuine miss still surfaces when
		// Output() applies these same mutations against full state.
		_ = mutationEngine.Apply(workingState, stepMutations, discardLog)
		if updated, ok := workingFaction.Assets[asset.ID]; ok {
			workingAsset = updated
		}
	}
	return mutations, nil
}

func confirmApplied(asset *domain.Asset, def *domain.AssetDefinition, collector action.Collector) ([]domain.Mutation, error) {
	if _, err := collector.ConfirmAbilityApplied(asset, def); err != nil {
		return nil, err
	}
	return nil, nil
}
