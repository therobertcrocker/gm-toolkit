package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// populationFloorRegistrar reads owner = "own" as "the owner is buying", and scope
// = "faction" as "on any world" (Colonists, Technical Expertise). asset_type narrows
// which definitions it serves. The entry's percent is of the setting's floor for
// the type; the lowest floor wins (Effort 4 Decision 4).
var populationFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPopulationFloorQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerPopulationFloor,
}

func registerPopulationFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPopulationFloorModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialPopulationFloor),
		&populationFloorModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			assetTypes: ctx.Entry.Filters.AssetTypes,
			percent:    ctx.Entry.PopulationFloor.Percent,
		},
	)
}

// populationFloorModifier holds its owner, not its holder's asset: a Cat 4 hook
// gets no FactionState (Decision Record 19).
type populationFloorModifier struct {
	owner      *domain.Faction
	holder     holderRef
	assetTypes []domain.AssetType
	percent    int
}

func (modifier *populationFloorModifier) ModifyPopulationFloor(faction *domain.Faction, def *domain.AssetDefinition, baseFloor, floor int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return floor
	}
	if len(modifier.assetTypes) > 0 && !slices.Contains(modifier.assetTypes, def.Type) {
		return floor
	}
	return min(floor, baseFloor*modifier.percent/100)
}
