package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// coinStealRegistrar reads owner = "rival" as "the faction the hit landed on".
// scope = "world" needs no check: an attack only reaches assets that share the
// attacker's world (or hex, in flight).
var coinStealRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeWorld},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerCoinSteal,
}

func registerCoinSteal(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialCoinSteal)+":"+ctx.Def.ID,
		&coinStealReactor{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			roller:         ctx.Roller,
		},
	)
}

// coinStealReactor reads its entries from the rulebook, not from registration,
// because the lead thief runs the entries of every thief its faction owns.
type coinStealReactor struct {
	ownerFactionID string
	assetID        string
	roller         domain.Roller
}

// OnMutations steals for every hit an active thief of the owner landed, in batch
// order, and caps each steal at what the target has left after earlier steals.
func (reactor *coinStealReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	owner := factionState.Factions[reactor.ownerFactionID]
	if owner == nil {
		return nil
	}
	thieves := activeAssetsWithEffect(owner, rulebook, domain.SpecialCoinSteal)
	// deliberate: only the lowest-ID active thief reacts, for every thief of every definition, because sibling reactors cannot see each other's output in one dispatch round and would together take more Coin than the target has.
	if len(thieves) == 0 || thieves[0].ID != reactor.assetID {
		return nil
	}
	remaining := make(map[string]int)
	var extra []domain.Mutation
	for _, found := range hits(mutations, false) {
		thief := owner.Assets[found.strikerID]
		target := factionState.Factions[found.targetFactionID]
		if thief == nil || !active(thief) || target == nil || target.ID == owner.ID {
			continue
		}
		if _, seen := remaining[target.ID]; !seen {
			remaining[target.ID] = target.Coin
		}
		def := rulebook.Assets[thief.DefinitionID]
		for _, entry := range def.Effects {
			if entry.Effect != domain.SpecialCoinSteal {
				continue
			}
			// deliberate: an empty purse is checked before the budget, because a steal that takes nothing has not "happened" and must not use up the turn's one theft.
			if remaining[target.ID] <= 0 {
				break
			}
			if entry.Gates.OncePer != "" && !spendBudget(owner, budgetKey(entry.Gates.OncePer, domain.SpecialCoinSteal, entry.Trigger, def.ID, thief.ID, target.ID)) {
				continue
			}
			stolen := min(entry.CoinSteal.Amount.Roll(reactor.roller), remaining[target.ID])
			remaining[target.ID] -= stolen
			extra = append(extra,
				domain.CoinDelta{FactionID: target.ID, Delta: -stolen, Cause: string(domain.SpecialCoinSteal), CausedByFactionID: owner.ID},
				domain.CoinDelta{FactionID: owner.ID, Delta: stolen, Cause: string(domain.SpecialCoinSteal), CausedByFactionID: owner.ID},
			)
		}
	}
	return extra
}

// activeAssetsWithEffect returns faction's active assets, in ID order, whose
// definition declares effectName.
func activeAssetsWithEffect(faction *domain.Faction, rulebook *rulebook.Rulebook, effectName domain.SpecialEffectType) []*domain.Asset {
	var matching []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil || !active(asset) {
			continue
		}
		for _, entry := range def.Effects {
			if entry.Effect == effectName {
				matching = append(matching, asset)
				break
			}
		}
	}
	return matching
}
