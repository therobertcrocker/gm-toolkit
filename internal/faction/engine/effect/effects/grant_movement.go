package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantMovementRegistrar reads owner = "own" as "the grant serves the declaring
// asset's faction", and scope = "faction" as "the faction's own assets move". The
// range and the asset filters come from the relocate block, measured from the
// declaring asset.
var grantMovementRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerMovementOptionsQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerGrantMovement,
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset; Effort 2's self-move mode lifts this (Effort 1 Decision 2).
	Validate: refuseTagHolder,
}

func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantMovement),
		&movementGranter{
			owner:    ctx.Holder.Faction,
			holder:   refOf(ctx.Holder),
			relocate: ctx.Entry.GrantMovement.Relocate,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner    *domain.Faction
	holder   holderRef
	relocate domain.RelocateParams
}

func (granter *movementGranter) GrantMovementAbilities(faction *domain.Faction) []hooks.MovementAbility {
	if faction.ID != granter.owner.ID {
		return nil
	}
	asset, live := granter.holder.liveIn(granter.owner)
	if !live {
		return nil
	}
	return []hooks.MovementAbility{{Source: asset, Relocate: granter.relocate}}
}
