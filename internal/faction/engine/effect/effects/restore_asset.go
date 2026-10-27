package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// restoreAssetRegistrar is Medical Center's save: when damage would destroy an own
// asset of a listed type on its world, the owner may pay the saved asset's half
// price to restore it at 1 HP. once_per = "asset" makes it once between the owner's
// turns for each Medical Center (Effort 2 Decision 15). owner = "own" reads as "the
// struck asset is the declaring asset's faction's", scope = "world" as "on the
// declaring asset's world".
var restoreAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerRestoreAsset,
}

func registerRestoreAsset(ctx effect.RegisterContext) {
	ctx.Registry.RegisterLethalInterceptor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialRestoreAsset),
		&restoreAssetInterceptor{
			holder:     refOf(ctx.Holder),
			assetTypes: ctx.Entry.Filters.AssetTypes,
			oncePer:    ctx.Entry.Gates.OncePer,
		},
	)
}

type restoreAssetInterceptor struct {
	holder     holderRef
	assetTypes []domain.AssetType
	oncePer    domain.OncePerKey
}

func (interceptor *restoreAssetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self, live := interceptor.holder.live(ctx.FactionState)
	if !live || !guardsTarget(self, ctx, interceptor.assetTypes) || ctx.TargetDef == nil {
		return nil
	}
	cause := domain.SpecialRestoreAsset
	// deliberate: the saved asset's half price, not the Medical Center's, because "its purchase cost" names the destroyed asset (Effort 2 Decision 15).
	price := ctx.TargetDef.HalfPrice()
	budget := ""
	if interceptor.oncePer != "" {
		budget = budgetKey(interceptor.oncePer, cause, domain.TriggerLethalDamage, interceptor.holder.sourceID, self.ID, ctx.Target.OwnerID)
	}
	return &hooks.Interception{
		Source:    self,
		Ask:       hooks.InterceptRestore,
		Price:     price,
		BudgetKey: budget,
		Mutations: append(hpTo(ctx, 1, cause, self.OwnerID), domain.CoinDelta{
			FactionID:         self.OwnerID,
			Delta:             -price,
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}
