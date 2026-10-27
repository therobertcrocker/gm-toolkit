package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// seizeTargetRegistrar is Hostile Takeover: when damage it deals, attack or counter,
// would destroy a rival's asset, the asset is reduced to 1 HP and joins the Hostile
// Takeover's faction instead (Effort 2 Decision 16). owner = "rival" reads as "the
// struck asset is another faction's", scope = "self" as "the declaring asset dealt
// the blow".
var seizeTargetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerSeizeTarget,
}

func registerSeizeTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialSeizeTarget), &seizeTargetInterceptor{holder: refOf(ctx.Holder)})
}

type seizeTargetInterceptor struct {
	holder holderRef
}

func (interceptor *seizeTargetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || ctx.Striker == nil || ctx.Striker.ID != self.ID || ctx.Target.OwnerID == self.OwnerID {
		return nil
	}
	cause := domain.SpecialSeizeTarget
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptAutomatic,
		Mutations: append(hpTo(ctx, 1, cause, self.OwnerID), domain.AssetOwnerChanged{
			FactionID:         ctx.Target.OwnerID,
			AssetID:           ctx.Target.ID,
			ToFactionID:       self.OwnerID,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
