package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// UseAssetAbility activates the special ability of one or more A-flagged assets.
// Assets are committed up front in resolution order; each is resolved in sequence.
type UseAssetAbility struct {
	collector      action.Collector
	roller         domain.Roller
	registry       *hooks.Registry
	world          *world.WorldEngine
	selectedAssets []*domain.Asset
	mutations      []domain.Mutation
}

func NewUseAssetAbility(collector action.Collector, roller domain.Roller, registry *hooks.Registry, worldEngine *world.WorldEngine) *UseAssetAbility {
	return &UseAssetAbility{
		collector: collector,
		roller:    roller,
		registry:  registry,
		world:     worldEngine,
	}
}

func (u *UseAssetAbility) Name() string { return "Use Asset Ability" }

func (u *UseAssetAbility) Validate(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	return len(abilityCandidates(faction, factionState, rulebook)) > 0
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := u.collector.SelectAbilityAssets(faction, abilityCandidates(faction, factionState, rulebook), rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selectedAssets = selected
	return nil
}

func (u *UseAssetAbility) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// deliberate: one view for the whole action, not one per asset, because the book lets one action use several assets and each must see what the earlier ones did (Effort 3 Decision 3).
	view := ability.NewView(factionState)
	for _, asset := range u.selectedAssets {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return fmt.Errorf("use asset ability: definition not found: %s", asset.DefinitionID)
		}
		mutations, err := ability.Dispatch(faction, asset, def, u.collector, u.roller, u.registry, factionState, view, rulebook, u.world)
		if err != nil {
			return fmt.Errorf("use asset ability: %w", err)
		}
		u.mutations = append(u.mutations, mutations...)
	}
	return nil
}

func (u *UseAssetAbility) Output() ([]domain.Mutation, error) {
	return u.mutations, nil
}

func hasActionFlag(def *domain.AssetDefinition) bool {
	for _, flag := range def.Flags {
		if flag == domain.FlagAction {
			return true
		}
	}
	return false
}

// abilityCandidates returns the faction's assets that may use their ability now,
// in ID order. An asset whose ability has no target now is not offered (Effort 3
// Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) []*domain.Asset {
	var candidates []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, asset, def, factionState) {
			candidates = append(candidates, asset)
		}
	}
	return candidates
}
