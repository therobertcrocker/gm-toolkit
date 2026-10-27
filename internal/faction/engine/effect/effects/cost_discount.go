package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// costDiscountRegistrar lowers the owner's purchase cost by amount, never below 0.
// owner = "own" reads as "the owner is buying", scope = "faction" as "on any
// world". min_tech_level narrows it to definitions that need at least that tech
// level (Preceptor Archive: TL4+).
var costDiscountRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAssetCostQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterMinTechLevel},
		},
	},
	Register: registerCostDiscount,
}

func registerCostDiscount(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetCostModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCostDiscount),
		&costDiscountModifier{
			owner:        ctx.Holder.Faction,
			holder:       refOf(ctx.Holder),
			minTechLevel: ctx.Entry.Filters.MinTechLevel,
			amount:       ctx.Entry.CostDiscount.Amount,
		},
	)
}

// costDiscountModifier holds its owner: a Cat 4 hook gets no FactionState
// (Decision Record 19).
type costDiscountModifier struct {
	owner        *domain.Faction
	holder       holderRef
	minTechLevel int
	amount       int
}

func (modifier *costDiscountModifier) ModifyAssetCost(buyer *domain.Faction, def *domain.AssetDefinition, _ string, query *hooks.AssetCostQuery) {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || buyer.ID != modifier.owner.ID || def.TechLevel < modifier.minTechLevel {
		return
	}
	query.Cost = max(query.Cost-modifier.amount, 0)
}
