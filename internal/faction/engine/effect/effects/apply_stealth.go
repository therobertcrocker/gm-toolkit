package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// applyStealthRegistrar reads owner = "own" as "the holder's faction bought the
// asset". scope = "self" is the declaring asset's own purchase; scope = "faction"
// is every asset the faction buys, which serves a tag (Secretive): an asset's
// on_purchase entries fire on its own purchase only, so an asset cannot hold it.
var applyStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction},
		},
	},
	Fire:     fireApplyStealth,
	Validate: validateApplyStealth,
}

func validateApplyStealth(entry domain.EffectDefinition, holder effect.HolderKind) error {
	if holder == effect.HolderAsset && entry.Filters.Scope == domain.ScopeFaction {
		return fmt.Errorf("an asset's on_purchase entry fires on its own purchase only, so scope %q would act as %q", domain.ScopeFaction, domain.ScopeSelf)
	}
	return nil
}

func fireApplyStealth(ctx effect.FireContext) []domain.Mutation {
	// deliberate: a second apply_stealth on one purchase emits again, not nothing, because both fire before either applies and the repeat changes no state (Effort 2 Decision 10).
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Holder.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialApplyStealth),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}
