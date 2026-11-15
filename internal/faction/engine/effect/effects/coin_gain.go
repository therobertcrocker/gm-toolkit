package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start: scope = "self" for an asset, "faction" for a tag (Aether Refiners),
// and "world" for an asset that pays the amount once per own asset of the
// asset_type filter on its place, up to max; no_stack lets one such holder per
// place pay (Company Coin). asset_destroyed reacts to each asset or Base destroyed
// in a fight: owner reads as whose asset it was ("own" the owner's, "rival"
// another faction's), and caused_by as who destroyed it. scope = "faction" is any
// world. once_per caps the payouts (War Profiteers). own_attack_hit and
// own_counter_hit pay once per hit the declaring asset lands (Debtors' Fleet).
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction, domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
			Gates:   []domain.EffectGateName{domain.GateNoStack},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
		domain.TriggerOwnAttackHit:  coinOnHitAccepts,
		domain.TriggerOwnCounterHit: coinOnHitAccepts,
	},
	Register: registerCoinGain,
	Fire:     fireCoinGain,
}

var coinOnHitAccepts = effect.Accepts{
	Owners: []domain.EffectOwner{domain.OwnerOwn},
	Scopes: []domain.EffectScope{domain.ScopeSelf},
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	amount := dice.Amount(ctx.Entry.CoinGain.Amount, ctx.Roller)
	if ctx.Entry.Filters.Scope == domain.ScopeWorld {
		amount *= countedAssets(ctx)
	}
	if amount <= 0 {
		return nil
	}
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Holder.Faction.ID,
		Delta:             amount,
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}

// countedAssets counts the holder's faction's active assets on the holder's place
// whose type the asset_type filter names, up to the entry's max. A holder in
// transit counts nothing. Under no_stack, a holder outranked by a lower-ID active
// asset of its own definition on the same place counts nothing.
func countedAssets(ctx effect.FireContext) int {
	holder := ctx.Holder.Asset
	worldID := holder.Location.WorldID
	if worldID == "" {
		return 0
	}
	count := 0
	for _, asset := range domain.SortedAssets(ctx.Holder.Faction) {
		if asset.Location.WorldID != worldID || !asset.Active() {
			continue
		}
		// deliberate: the same definition outranks, not any counting coin_gain, because no_stack means one Company Coin per place, and a different counting asset is a different payout.
		if ctx.Entry.Gates.NoStack && asset.DefinitionID == holder.DefinitionID && asset.ID < holder.ID {
			return 0
		}
		def := ctx.Rulebook.Assets[asset.DefinitionID]
		if def != nil && slices.Contains(ctx.Entry.Filters.AssetTypes, def.Type) {
			count++
		}
	}
	if limit := ctx.Entry.CoinGain.Max; limit > 0 {
		count = min(count, limit)
	}
	return count
}

func registerCoinGain(ctx effect.RegisterContext) {
	if ctx.Entry.Trigger == domain.TriggerOwnAttackHit || ctx.Entry.Trigger == domain.TriggerOwnCounterHit {
		ctx.Registry.RegisterMutationReactor(
			ctx.Holder.Scope(),
			ctx.Holder.HookID(domain.SpecialCoinGain),
			&coinOnHitReactor{
				holder:  refOf(ctx.Holder),
				counter: ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
				amount:  ctx.Entry.CoinGain.Amount,
				roller:  ctx.Roller,
			},
		)
		return
	}
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialCoinGain),
		&coinGainReactor{
			holder:   refOf(ctx.Holder),
			owner:    ctx.Entry.Filters.Owner,
			causedBy: ctx.Entry.Filters.CausedBy,
			oncePer:  ctx.Entry.Gates.OncePer,
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
	)
}

// combatLossCauses are the removal causes that mean damage in a fight destroyed an
// asset or a Base: an attack or counterattack, a Cracked Comms self-attack, an
// asset's own on-hit damage, or the free attack on a new Base. A sale, a refit or
// a missed maintenance is not one.
var combatLossCauses = []string{"attack", string(domain.SpecialCounterstrike), string(domain.SpecialSelfDamage), "expand"}

// combatLoss reports whose asset or Base a mutation destroyed in a fight, and who
// destroyed it. A Base of Influence counts as an asset (Scavengers, A.001.7 close).
func combatLoss(mutation domain.Mutation) (factionID, causedByFactionID string, ok bool) {
	switch loss := mutation.(type) {
	case domain.AssetRemoved:
		return loss.FactionID, loss.CausedByFactionID, slices.Contains(combatLossCauses, loss.Cause)
	case domain.BaseDestroyed:
		return loss.FactionID, loss.CausedByFactionID, slices.Contains(combatLossCauses, loss.Cause)
	}
	return "", "", false
}

type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	oncePer  domain.OncePerKey
	amount   domain.DiceRoll
	roller   dice.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match, while its once_per budget lasts.
func (reactor *coinGainReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	ownerID := reactor.holder.factionID
	var extra []domain.Mutation
	for _, mutation := range mutations {
		lostBy, causedBy, ok := combatLoss(mutation)
		if !ok {
			continue
		}
		if !onSide(reactor.owner, lostBy, ownerID) {
			continue
		}
		if reactor.causedBy != "" && !onSide(reactor.causedBy, causedBy, ownerID) {
			continue
		}
		if reactor.oncePer != "" && !spendBudget(factionState.Factions[ownerID], budgetKey(reactor.oncePer, domain.SpecialCoinGain, reactor.holder.sourceID, reactor.holder.assetID, lostBy)) {
			continue
		}
		extra = append(extra, domain.CoinDelta{
			FactionID:         ownerID,
			Delta:             dice.Amount(reactor.amount, reactor.roller),
			Cause:             string(domain.SpecialCoinGain),
			CausedByFactionID: ownerID,
		})
	}
	return extra
}

// coinOnHitReactor pays the owner the amount for each hit its asset lands in the
// batch: attack hits, or counterattack hits when counter is set. The Coin comes
// from no one (Effort 6 Decision 14).
type coinOnHitReactor struct {
	holder  holderRef
	counter bool
	amount  domain.DiceRoll
	roller  dice.Roller
}

func (reactor *coinOnHitReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	var extra []domain.Mutation
	for range hitTargets(mutations, reactor.holder.assetID, reactor.counter) {
		extra = append(extra, domain.CoinDelta{
			FactionID:         reactor.holder.factionID,
			Delta:             dice.Amount(reactor.amount, reactor.roller),
			Cause:             string(domain.SpecialCoinGain),
			CausedByFactionID: reactor.holder.factionID,
		})
	}
	return extra
}

// onSide reports whether factionID stands on side relative to ownerID: "own" is
// the owner, "rival" another faction, "any" either.
func onSide(side domain.EffectOwner, factionID, ownerID string) bool {
	switch side {
	case domain.OwnerOwn:
		return factionID == ownerID
	case domain.OwnerRival:
		return factionID != "" && factionID != ownerID
	}
	return true
}
