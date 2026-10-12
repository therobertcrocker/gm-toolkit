package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// removeTagDiceRegistrar reads owner = "rival" as "a rival is rolling against the
// owner", and scope = "self" as "against the declaring asset itself" — attacking
// it, or defending against it.
var removeTagDiceRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerEnemyRollAtMe: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerRemoveTagDice,
}

func registerRemoveTagDice(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollOfferFilter(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialRemoveTagDice)+":"+ctx.Def.ID,
		&tagDiceFilter{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			trigger:        ctx.Entry.Trigger,
			filters:        ctx.Entry.Filters,
		},
	)
}

type tagDiceFilter struct {
	ownerFactionID string
	assetID        string
	trigger        domain.EffectTrigger
	filters        domain.EffectFilters
}

func (filter *tagDiceFilter) FilterOffers(ctx hooks.RollContext, offers []hooks.ModifierOffer, factionState *state.FactionState) []hooks.ModifierOffer {
	asset := liveAsset(factionState, filter.ownerFactionID, filter.assetID)
	if !active(asset) || !rollMatches(ctx, filter.trigger, filter.ownerFactionID, asset, filter.filters) {
		return offers
	}
	var kept []hooks.ModifierOffer
	for _, offer := range offers {
		if !offer.FromTag {
			kept = append(kept, offer)
		}
	}
	return kept
}
