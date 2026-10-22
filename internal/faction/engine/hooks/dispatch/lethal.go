package dispatch

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// InterceptLethal asks the Cat 6 interceptors whether one takes ctx's lethal hit,
// in D-7's order: the target owner's automatic interceptors, then its elective
// ones, then the striker's, each group in asset ID order. The first that takes the
// hit ends the walk, and its mutations replace the target's AssetRemoved. It
// returns nil when none takes it. A nil collector asks nothing, so only automatic
// interceptors can take the hit (D-10).
func InterceptLethal(registry *hooks.Registry, ctx hooks.LethalContext, collector hooks.InterceptCollector) ([]domain.Mutation, error) {
	if registry == nil {
		return nil, nil
	}
	targetSide, strikerSide := lethalOffers(registry, ctx)
	for _, side := range [][]*hooks.Interception{targetSide, strikerSide} {
		for _, elective := range []bool{false, true} {
			for _, offer := range side {
				if (offer.Ask != hooks.InterceptAutomatic) != elective {
					continue
				}
				taken, err := takeInterception(offer, ctx, collector)
				if err != nil {
					return nil, err
				}
				if taken {
					return offer.Mutations, nil
				}
			}
		}
	}
	return nil, nil
}

// InterceptLethalAutomatic runs only the automatic interceptors, for a lethal hit
// inside a Cat 3 reactor, which has no collector to ask (D-10).
func InterceptLethalAutomatic(registry *hooks.Registry, ctx hooks.LethalContext) []domain.Mutation {
	// deliberate: drop the error, not return it, because with no collector nothing is asked and nothing can fail.
	mutations, _ := InterceptLethal(registry, ctx, nil)
	return mutations
}

// lethalOffers collects the offers for ctx: from each asset the target's owner
// holds, in ID order, then from the striker. Each interceptor decides whether its
// reach covers the target.
func lethalOffers(registry *hooks.Registry, ctx hooks.LethalContext) (targetSide, strikerSide []*hooks.Interception) {
	if owner := ctx.FactionState.Factions[ctx.Target.OwnerID]; owner != nil {
		for _, asset := range domain.SortedAssets(owner) {
			targetSide = append(targetSide, offersFrom(registry.LethalInterceptorsForAsset(owner.ID, asset.ID), ctx)...)
		}
	}
	if ctx.Striker != nil {
		strikerSide = offersFrom(registry.LethalInterceptorsForAsset(ctx.Striker.OwnerID, ctx.Striker.ID), ctx)
	}
	return targetSide, strikerSide
}

// offersFrom asks each registered interceptor for its offer. An interceptor that
// the batch already removes or moves to another faction offers nothing: a False
// Front sacrificed earlier in the action cannot be sacrificed again.
func offersFrom(registered []hooks.RegisteredLethalInterceptor, ctx hooks.LethalContext) []*hooks.Interception {
	var offers []*hooks.Interception
	for _, interceptor := range registered {
		offer := interceptor.Hook.OfferInterception(ctx)
		if offer != nil && !leftInBatch(ctx.Batch, offer.Source.ID) {
			offers = append(offers, offer)
		}
	}
	return offers
}

// leftInBatch reports whether batch removes assetID or moves it to another faction.
func leftInBatch(batch []domain.Mutation, assetID string) bool {
	for _, mutation := range batch {
		switch m := mutation.(type) {
		case domain.AssetRemoved:
			if m.AssetID == assetID {
				return true
			}
		case domain.AssetOwnerChanged:
			if m.AssetID == assetID {
				return true
			}
		}
	}
	return false
}

// takeInterception reports whether offer takes the hit. An automatic offer always
// does. An elective one does when its owner says yes, and a restore only when the
// owner can pay. A taken offer spends its budget on the owner.
func takeInterception(offer *hooks.Interception, ctx hooks.LethalContext, collector hooks.InterceptCollector) (bool, error) {
	owner := ctx.FactionState.Factions[offer.Source.OwnerID]
	if owner == nil || owner.BudgetSpent(offer.BudgetKey) {
		return false, nil
	}
	switch offer.Ask {
	case hooks.InterceptSacrifice:
		if collector == nil {
			return false, nil
		}
		confirmed, err := collector.ConfirmSacrifice(hooks.SacrificeOffer{Owner: owner, Sacrifice: offer.Source, Saves: ctx.Target})
		if err != nil || !confirmed {
			return false, err
		}
	case hooks.InterceptRestore:
		// deliberate: count the Coin the batch already spends, not only state Coin, because two Medical Centers can each restore in one Attack and state Coin changes only at apply (Effort 2 Decision 15).
		if collector == nil || owner.Coin+batchCoin(ctx.Batch, owner.ID) < offer.Price {
			return false, nil
		}
		confirmed, err := collector.ConfirmPayment(hooks.PaymentOffer{Payer: owner, Asset: ctx.Target, Amount: offer.Price, Reason: hooks.PaymentRestore})
		if err != nil || !confirmed {
			return false, err
		}
	}
	owner.SpendBudget(offer.BudgetKey)
	return true, nil
}

// batchCoin is the net Coin change batch makes to factionID before it applies.
func batchCoin(batch []domain.Mutation, factionID string) int {
	total := 0
	for _, mutation := range batch {
		if coin, ok := mutation.(domain.CoinDelta); ok && coin.FactionID == factionID {
			total += coin.Delta
		}
	}
	return total
}
