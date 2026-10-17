package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// costSurchargeRegistrar reads owner = "rival" as "the buyer is not the owner",
// and scope = "world" as "the declaring asset stands on the purchase world". The
// surcharge is lost Coin, not a gain to the owner, so the hook only raises the
// cost and emits nothing.
var costSurchargeRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerAssetCostQuery: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateNoStack},
		},
	},
	Register: registerCostSurcharge,
}

func registerCostSurcharge(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetCostModifier(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialCostSurcharge)+":"+ctx.Def.ID,
		&costSurchargeModifier{
			owner:   ctx.Faction,
			assetID: ctx.Asset.ID,
			noStack: ctx.Entry.Gates.NoStack,
			amount:  ctx.Entry.CostSurcharge.Amount,
		},
	)
}

// costSurchargeModifier holds its owner, not its asset, because a Cat 4 hook gets
// no FactionState: the owner pointer is live all cycle, and an asset removed
// mid-cycle reads as nil from owner.Assets where a held asset pointer would not.
type costSurchargeModifier struct {
	owner   *domain.Faction
	assetID string
	noStack bool
	amount  int
}

func (modifier *costSurchargeModifier) ModifyAssetCost(buyer *domain.Faction, _ *domain.AssetDefinition, world string, query *hooks.AssetCostQuery) {
	asset := modifier.owner.Assets[modifier.assetID]
	if !active(asset) || buyer.ID == modifier.owner.ID || asset.Location.WorldID != world {
		return
	}

	key := string(domain.SpecialCostSurcharge)
	if modifier.noStack && query.Applied[key] {
		return
	}

	query.Cost += modifier.amount
	query.Surcharges = append(query.Surcharges, hooks.Surcharge{FactionID: modifier.owner.ID, Amount: modifier.amount})
	query.Applied[key] = true
}
