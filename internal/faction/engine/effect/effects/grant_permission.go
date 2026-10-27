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
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialGrantPermission),
		&permissionGranter{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder)},
	)
}

// permissionGranter holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19).
type permissionGranter struct {
	owner  *domain.Faction
	holder holderRef
}

func (granter *permissionGranter) GrantsPermission(query hooks.PermissionQuery) bool {
	self, live := granter.holder.liveIn(granter.owner)
	if !live || query.Requester.ID != granter.owner.ID {
		return false
	}
	return self.Location.WorldID != "" && self.Location.WorldID == query.WorldID
}
