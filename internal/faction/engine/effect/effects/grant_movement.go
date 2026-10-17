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
}

func registerGrantMovement(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetMovementGranter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialGrantMovement)+":"+ctx.Def.ID,
		&movementGranter{
			owner:    ctx.Faction,
			assetID:  ctx.Asset.ID,
			relocate: ctx.Entry.GrantMovement.Relocate,
		},
	)
}

// movementGranter reads its asset through its owner at query time (Decision
// Record 19), so a grant from an asset removed or made inactive mid-cycle lapses.
type movementGranter struct {
	owner    *domain.Faction
	assetID  string
	relocate domain.RelocateParams
}

func (granter *movementGranter) GrantMovementAbilities(faction *domain.Faction) []hooks.MovementAbility {
	if faction.ID != granter.owner.ID {
		return nil
	}
	asset := granter.owner.Assets[granter.assetID]
	if !active(asset) {
		return nil
	}
	return []hooks.MovementAbility{{Source: asset, Relocate: granter.relocate}}
}
