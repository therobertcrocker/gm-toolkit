package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// discountRepairRegistrar is Medical Center's repair clause: the owner's Repair
// Asset costs amount less for each asset of a listed type repaired on the declaring
// asset's world. owner = "own" reads as "the repairing faction is the owner", scope
// = "world" as "the repaired asset is on the declaring asset's world". no_stack
// keeps a second Medical Center there from adding to it (Effort 2 Decision 15).
var discountRepairRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerRepairCostQuery: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateNoStack},
		},
	},
	Register: registerDiscountRepair,
}

func registerDiscountRepair(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRepairCostModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialDiscountRepair),
		&discountRepairModifier{
			owner:      ctx.Holder.Faction,
			holder:     refOf(ctx.Holder),
			noStack:    ctx.Entry.Gates.NoStack,
			amount:     ctx.Entry.DiscountRepair.Amount,
			assetTypes: ctx.Entry.Filters.AssetTypes,
		},
	)
}

// discountRepairModifier holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type discountRepairModifier struct {
	owner      *domain.Faction
	holder     holderRef
	noStack    bool
	amount     int
	assetTypes []domain.AssetType
}

func (modifier *discountRepairModifier) ModifyRepairCost(repairer *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, query *hooks.RepairCostQuery) {
	self, live := modifier.holder.liveIn(modifier.owner)
	if !live || repairer.ID != modifier.owner.ID {
		return
	}
	if self.Location.WorldID == "" || asset.Location.WorldID != self.Location.WorldID {
		return
	}
	if len(modifier.assetTypes) > 0 && !slices.Contains(modifier.assetTypes, def.Type) {
		return
	}
	key := string(domain.SpecialDiscountRepair)
	if modifier.noStack && query.Applied[key] {
		return
	}
	query.Cost -= modifier.amount
	query.Applied[key] = true
}
