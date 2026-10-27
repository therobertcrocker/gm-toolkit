package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinGainRegistrar pays the owner the amount. turn_start fires at the owner's
// turn start. asset_destroyed reacts to each asset destroyed in a fight: owner
// reads as whose asset it was ("own" the owner's, "rival" another faction's), and
// caused_by as who destroyed it. scope = "faction" is any world.
var coinGainRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerTurnStart: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
		domain.TriggerAssetDestroyed: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn, domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterCausedBy},
		},
	},
	Register: registerCoinGain,
	Fire:     fireCoinGain,
}

func fireCoinGain(ctx effect.FireContext) []domain.Mutation {
	return []domain.Mutation{domain.CoinDelta{
		FactionID:         ctx.Holder.Faction.ID,
		Delta:             ctx.Entry.CoinGain.Amount.Roll(ctx.Roller),
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
			amount:   ctx.Entry.CoinGain.Amount,
			roller:   ctx.Roller,
		},
	)
}

// combatLossCauses are the removal causes that mean an asset was destroyed in a
// fight: an attack or counterattack, or a Cracked Comms self-attack.
var combatLossCauses = []string{"attack", string(domain.SpecialCounterstrike)}

type coinGainReactor struct {
	holder   holderRef
	owner    domain.EffectOwner
	causedBy domain.EffectOwner
	amount   domain.DiceRoll
	roller   domain.Roller
}

// OnMutations pays once for each combat removal in the batch that the entry's
// owner and caused_by filters match.
func (reactor *coinGainReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	ownerID := reactor.holder.factionID
	var extra []domain.Mutation
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || !slices.Contains(combatLossCauses, removed.Cause) {
			continue
		}
		if !onSide(reactor.owner, removed.FactionID, ownerID) {
			continue
		}
		if reactor.causedBy != "" && !onSide(reactor.causedBy, removed.CausedByFactionID, ownerID) {
			continue
		}
		extra = append(extra, domain.CoinDelta{
			FactionID:         ownerID,
			Delta:             reactor.amount.Roll(reactor.roller),
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
