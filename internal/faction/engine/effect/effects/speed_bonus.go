package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// speedBonusRegistrar reads owner and scope as speed_floor does: every asset the
// owner holds. asset_type narrows the bonus to those types, and hot to orders
// run hot (Aether Burners).
var speedBonusRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerSpeedQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAssetType, domain.FilterHot},
		},
	},
	Register: registerSpeedBonus,
}

func registerSpeedBonus(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetSpeedModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSpeedBonus),
		&speedBonusModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			amount: ctx.Entry.SpeedBonus.Amount,
			types:  ctx.Entry.Filters.AssetTypes,
			hot:    ctx.Entry.Filters.Hot,
		},
	)
}

// speedBonusModifier holds its owner, as speedFloorModifier does: a Cat 4 hook
// gets no FactionState.
type speedBonusModifier struct {
	owner  *domain.Faction
	holder holderRef
	amount int
	types  []domain.AssetType
	hot    bool
}

func (modifier *speedBonusModifier) ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int, hot bool) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	if modifier.hot && !hot {
		return speed
	}
	if len(modifier.types) > 0 && !slices.Contains(modifier.types, def.Type) {
		return speed
	}
	return speed + modifier.amount
}
