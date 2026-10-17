package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
		},
	},
	Register: registerBonusDie,
}

func registerBonusDie(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollModifier(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialBonusDie)+":"+ctx.Def.ID,
		&bonusDieModifier{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			source:         ctx.Def.Name,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
			params:         *ctx.Entry.BonusDie,
		},
	)
}

type bonusDieModifier struct {
	ownerFactionID string
	assetID        string
	source         string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
	params         domain.BonusDieParams
}

func (modifier *bonusDieModifier) OfferModifiers(ctx hooks.RollContext, factionState *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	asset := liveAsset(factionState, modifier.ownerFactionID, modifier.assetID)
	if !active(asset) || !rollMatches(ctx, modifier.trigger, modifier.ownerFactionID, asset, modifier.filters) {
		return nil
	}
	params := modifier.params
	// deliberate: one offer per asset copy, not one per definition, because each copy grants its own die (Effort 3 plan, Decision Record 12).
	return []hooks.ModifierOffer{{
		Source:      modifier.source,
		Description: fmt.Sprintf("roll %d extra d%d, keep highest", params.Count, params.Sides),
		Apply: func(rollState *hooks.RollState) {
			for range params.Count {
				rollState.AddDie(params.Sides)
			}
			// deliberate: keep 1, not 1 + count, because every faction roll is one d10 and an extra die is pick-the-best ("an additional d10, keep the highest").
			rollState.SetKeepHighest(1)
		},
	}}
}
