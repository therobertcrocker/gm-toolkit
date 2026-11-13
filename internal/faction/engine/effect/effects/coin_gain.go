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
// turn start: scope = "self" for an asset, "faction" for a tag (Aether Refiners).
// asset_destroyed reacts to each asset or Base destroyed in a fight: owner reads as
// whose asset it was ("own" the owner's, "rival" another faction's), and caused_by
// as who destroyed it. scope = "faction" is any world. once_per caps the payouts
// (War Profiteers).
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerCoinGain,
	Fire:     fireCoinGain,
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Holder.Faction.ID,
		Delta:             dice.Amount(ctx.Entry.CoinGain.Amount, ctx.Roller),
		Cause:             string(domain.SpecialCoinGain),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}

func registerCoinGain(ctx effect.RegisterContext) {
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
