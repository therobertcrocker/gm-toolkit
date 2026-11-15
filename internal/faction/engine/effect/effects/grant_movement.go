package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantMovementRegistrar reads owner = "own" as "the grant is the declaring
// asset's faction's", and scope = "faction" as "the riding faction's own assets
// move". The range and the asset filters come from the relocate block, measured
// from the declaring asset. A public grant serves every faction (Effort 6
// Decision 9).
var grantMovementRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerMovementOptionsQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerGrantMovement,
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset and a tag has none; Mercenary Group is a speed floor, not a grant (Effort 2 Decision 7).
	Validate: refuseTagHolder,
}

func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantMovement),
		&movementGranter{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			grant:  *ctx.Entry.GrantMovement,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner  *domain.Faction
	holder holderRef
	grant  domain.GrantMovementParams
}

// GrantMovementAbilities offers the grant to its owner, and to any other faction
// when it is public and its asset is not hidden. A rider other than the owner
// pays the fare to the owner.
func (granter *movementGranter) GrantMovementAbilities(faction *domain.Faction) []hooks.MovementAbility {
	rider := faction.ID != granter.owner.ID
	if rider && !granter.grant.Public {
		return nil
	}
	asset, live := granter.holder.liveIn(granter.owner)
	if !live || (rider && asset.Hidden()) {
		return nil
	}
	relocate := granter.grant.Relocate
	relocate.CoinCost = granter.grant.Fare(granter.owner.ID, faction.ID)
	ability := hooks.MovementAbility{Source: asset, Relocate: relocate}
	if rider {
		ability.Payee = granter.owner.ID
	}
	return []hooks.MovementAbility{ability}
}
