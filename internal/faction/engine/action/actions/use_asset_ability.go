package actions

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/ability"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// UseAssetAbility activates the abilities of one or more holders: A-flagged assets
// and the faction's tags (Effort 5 Decision 1). Holders are committed up front in
// resolution order; each is resolved in sequence.
type UseAssetAbility struct {
	collector action.Collector
	roller    domain.Roller
	registry  *hooks.Registry
	world     *world.WorldEngine
	selected  []domain.AbilityHolder
	mutations []domain.Mutation
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
	return len(abilityCandidates(faction, factionState, rulebook, u.world, u.registry)) > 0
}

func (u *UseAssetAbility) Inputs(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	selected, err := u.collector.SelectAbilityHolders(faction, abilityCandidates(faction, factionState, rulebook, u.world, u.registry), rulebook)
	if err != nil {
		return fmt.Errorf("use asset ability: %w", err)
	}
	u.selected = selected
	return nil
}

func (u *UseAssetAbility) Resolve(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook) error {
	// deliberate: one view for the whole action, not one per asset, because the book lets one action use several assets and each must see what the earlier ones did (Effort 3 Decision 3).
	view := ability.NewView(factionState)
	for _, holder := range u.selected {
		var def *domain.AssetDefinition
		if holder.Asset != nil {
			found, ok := rulebook.Assets[holder.Asset.DefinitionID]
			if !ok {
				return fmt.Errorf("use asset ability: definition not found: %s", holder.Asset.DefinitionID)
			}
			def = found
		}
		mutations, err := ability.Dispatch(faction, holder, def, u.collector, u.roller, u.registry, factionState, view, rulebook, u.world)
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

// abilityCandidates returns the holders that may use their ability now: the
// faction's A-flag assets in ID order, then its tags that carry an ability, in ID
// order. A holder whose ability has no target now is not offered (Effort 3
// Decision 4).
func abilityCandidates(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, worldEngine *world.WorldEngine, registry *hooks.Registry) []domain.AbilityHolder {
	var candidates []domain.AbilityHolder
	for _, asset := range domain.SortedAssets(faction) {
		def, ok := rulebook.Assets[asset.DefinitionID]
		holder := domain.AbilityHolder{Asset: asset}
		if ok && hasActionFlag(def) && asset.Active() && !asset.HasCondition(domain.ConditionNoAbility) && ability.HasTargets(faction, holder, def.Ability, factionState, worldEngine, registry) {
			candidates = append(candidates, holder)
		}
	}
	for _, tagID := range slices.Sorted(slices.Values(faction.Tags)) {
		tag, ok := rulebook.Tags[tagID]
		if !ok || tag.Ability == nil {
			continue
		}
		holder := domain.AbilityHolder{Tag: tag}
		if ability.HasTargets(faction, holder, tag.Ability, factionState, worldEngine, registry) {
			candidates = append(candidates, holder)
		}
	}
	return candidates
}
