package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// grantPermissionRegistrar is Popular Movement: on the declaring asset's world,
// its owner's requests are granted at once. owner = "own" reads as "the requester
// is the owner", scope = "world" as "the request is for the declaring asset's world".
var grantPermissionRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerPermissionQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
		},
	},
	Register: registerGrantPermission,
}

func registerGrantPermission(ctx effect.RegisterContext) {
	ctx.Registry.RegisterPermissionGranter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialGrantPermission)+":"+ctx.Def.ID,
		&permissionGranter{owner: ctx.Faction, assetID: ctx.Asset.ID},
	)
}

// permissionGranter holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19).
type permissionGranter struct {
	owner   *domain.Faction
	assetID string
}

func (granter *permissionGranter) GrantsPermission(query hooks.PermissionQuery) bool {
	self := granter.owner.Assets[granter.assetID]
	if !active(self) || query.Requester.ID != granter.owner.ID {
		return false
	}
	return self.Location.WorldID != "" && self.Location.WorldID == query.WorldID
}
