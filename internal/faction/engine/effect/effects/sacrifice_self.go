package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// sacrificeSelfRegistrar is False Front: when damage would destroy another own
// asset on its world, the owner may sacrifice the False Front to nullify the
// killing blow, and the struck asset keeps its HP from before that hit (Effort 2
// Decision 14). owner = "own" reads as "the struck asset is the declaring asset's
// faction's", scope = "world" as "on the declaring asset's world". An asset_type
// filter guards only the listed types (Glamoured Doubles guard Operatives); with
// none it guards every type (False Front, Unpaid Interns).
var sacrificeSelfRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerSacrificeSelf,
}

func registerSacrificeSelf(ctx effect.RegisterContext) {
	interceptor := &sacrificeSelfInterceptor{holder: refOf(ctx.Holder), assetTypes: ctx.Entry.Filters.AssetTypes}
	ctx.Registry.RegisterLethalInterceptor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialSacrificeSelf), interceptor)
}

type sacrificeSelfInterceptor struct {
	holder     holderRef
	assetTypes []domain.AssetType
}

func (interceptor *sacrificeSelfInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialSacrificeSelf
	// deliberate: the removal's cause is the sacrifice, not "attack", because the owner gave the False Front up and no one scores a kill (Effort 2 Decision 14).
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptSacrifice,
		Mutations: append(hpTo(ctx, ctx.HPAfter+ctx.Damage, cause, self.OwnerID), domain.AssetRemoved{
			FactionID:         self.OwnerID,
			AssetID:           self.ID,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
