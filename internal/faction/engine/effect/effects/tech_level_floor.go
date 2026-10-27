package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// techLevelFloorRegistrar reads owner = "own" as "the owner is buying", scope =
// "world" as "the declaring asset stands on the purchase world", and scope =
// "faction" as "any world". category narrows which definitions the floor serves.
var techLevelFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTechLevelQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld, domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCategory},
		},
	},
	Register: registerTechLevelFloor,
}

func registerTechLevelFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterWorldTechLevelModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialTechLevelFloor),
		&techLevelFloorModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			scope:      ctx.Entry.Filters.Scope,
			categories: ctx.Entry.Filters.Categories,
			level:      ctx.Entry.TechLevelFloor.Level,
		},
	)
}

// techLevelFloorModifier holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type techLevelFloorModifier struct {
	owner      *domain.Faction
	holder     holderRef
	scope      domain.EffectScope
	categories []domain.FactionStat
	level      int
}

func (modifier *techLevelFloorModifier) ModifyWorldTechLevel(faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int {
	asset, live := modifier.holder.liveIn(modifier.owner)
	if !live || faction.ID != modifier.owner.ID {
		return baseTL
	}
	if modifier.scope == domain.ScopeWorld && asset.Location.WorldID != world {
		return baseTL
	}
	if len(modifier.categories) > 0 && !slices.Contains(modifier.categories, def.Category) {
		return baseTL
	}
	return max(baseTL, modifier.level)
}
