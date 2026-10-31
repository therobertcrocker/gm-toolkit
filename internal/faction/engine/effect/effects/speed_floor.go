package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// speedFloorRegistrar reads owner = "own" as "the moving asset is the owner's",
// and scope = "faction" as "every asset the owner holds" (Mercenary Group). The
// entry carries the floor; an asset already faster keeps its speed.
var speedFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerSpeedQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerSpeedFloor,
}

func registerSpeedFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetSpeedModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSpeedFloor),
		&speedFloorModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			speed:  ctx.Entry.SpeedFloor.Speed,
		},
	)
}

// speedFloorModifier holds its owner, not its holder's asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type speedFloorModifier struct {
	owner  *domain.Faction
	holder holderRef
	speed  int
}

func (modifier *speedFloorModifier) ModifyAssetSpeed(faction *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	return max(speed, modifier.speed)
}
