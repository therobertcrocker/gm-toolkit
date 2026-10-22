package ability

import (
	"errors"
	"fmt"
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
	view *state.FactionState,
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
	workingFaction := view.Factions[faction.ID]
	if workingFaction == nil {
		return nil, fmt.Errorf("ability: acting faction %s not in the view", faction.ID)
	}
	workingAsset := workingFaction.Assets[asset.ID]
	if workingAsset == nil {
		workingAsset = asset.Clone()
	}
	discardLog := slog.New(slog.DiscardHandler)
	mutationEngine := mutation.New(discardLog)

	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		handler := handlers[step.Effect]
		target, err := resolveTarget(step, workingFaction, workingAsset, view, collector)
		// deliberate: fizzle this asset, not fail the action, because an earlier asset in the same action may have used up the targets and its result must stand (Effort 3 Decision Record 1).
		if errors.Is(err, errNoTargets) {
			break
		}
		if err != nil {
			return nil, err
		}
		// deliberate: the real faction, not workingFaction, because a roll hook's budget spent on the view would be lost (Effort 3 plan, Decision 12).
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
			FactionState: view,
			Rulebook:     rb,
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
