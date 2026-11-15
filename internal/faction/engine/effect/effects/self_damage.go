package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

var selfDamageAccepts = effect.Accepts{
	Owners: []domain.EffectOwner{domain.OwnerOwn},
	Scopes: []domain.EffectScope{domain.ScopeSelf},
}

var selfDamageRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit:  selfDamageAccepts,
		domain.TriggerOwnCounterHit: selfDamageAccepts,
		domain.TriggerHotMove: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerSelfDamage,
}

func registerSelfDamage(ctx effect.RegisterContext) {
	if ctx.Entry.Trigger == domain.TriggerHotMove {
		ctx.Registry.RegisterMutationReactor(
			ctx.Holder.Scope(),
			ctx.Holder.HookID(domain.SpecialSelfDamage),
			&hotBurnReactor{
				holder:   refOf(ctx.Holder),
				amount:   ctx.Entry.SelfDamage.Amount,
				roller:   ctx.Roller,
				registry: ctx.Registry,
			},
		)
		return
	}
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSelfDamage),
		&selfDamageReactor{
			holder:   refOf(ctx.Holder),
			def:      ctx.Holder.Def,
			counter:  ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
			amount:   ctx.Entry.SelfDamage.Amount,
			roller:   ctx.Roller,
			registry: ctx.Registry,
		},
	)
}

// selfDamageReactor keeps its definition and the registry, because a lethal
// self-hit asks the Cat 6 interceptors (D-10).
type selfDamageReactor struct {
	holder   holderRef
	def      *domain.AssetDefinition
	counter  bool
	amount   domain.DiceRoll
	roller   dice.Roller
	registry *hooks.Registry
}

func (reactor *selfDamageReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if len(hitTargets(mutations, reactor.holder.assetID, reactor.counter)) == 0 {
		return nil
	}
	asset, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	return damageSelf(reactor.registry, factionState, mutations, reactor.holder.factionID, asset, reactor.def, reactor.amount, reactor.roller)
}

// hotBurnReactor is running hot's cost: each asset of its owner that a movement
// tick moves on an order run hot takes the entry's damage (Aether Burners).
type hotBurnReactor struct {
	holder   holderRef
	amount   domain.DiceRoll
	roller   dice.Roller
	registry *hooks.Registry
}

func (reactor *hotBurnReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	owner := factionState.Factions[reactor.holder.factionID]
	var extra []domain.Mutation
	for _, asset := range hotMovers(mutations, owner) {
		batch := append(slices.Clone(mutations), extra...)
		extra = append(extra, damageSelf(reactor.registry, factionState, batch, owner.ID, asset, rulebook.Assets[asset.DefinitionID], reactor.amount, reactor.roller)...)
	}
	return extra
}

// hotMovers returns owner's assets that a movement tick in mutations moved on an
// order run hot: a step taken, points banked toward a Drift jump, an arrival, or
// the first hold at a locked place.
func hotMovers(mutations []domain.Mutation, owner *domain.Faction) []*domain.Asset {
	var movers []*domain.Asset
	for _, mutation := range mutations {
		var factionID, assetID string
		switch movement := mutation.(type) {
		case domain.MovementOrderProgressed:
			factionID, assetID = movement.FactionID, movement.AssetID
		case domain.MovementOrderCompleted:
			factionID, assetID = movement.FactionID, movement.AssetID
		case domain.MovementOrderHeld:
			// deliberate: an order already held does not burn, not every held tick, because a ship waiting at the edge is parked; the tick that brings it there moved (A.001.5 Effort 4 Decision Record 10).
			if asset, ok := owner.Assets[movement.AssetID]; ok && asset.CurrentOrder != nil && asset.CurrentOrder.Held {
				continue
			}
			factionID, assetID = movement.FactionID, movement.AssetID
		}
		if factionID != owner.ID {
			continue
		}
		if asset, ok := owner.Assets[assetID]; ok && asset.CurrentOrder != nil && asset.CurrentOrder.RunHot {
			movers = append(movers, asset)
		}
	}
	return movers
}

// damageSelf is a hit an asset deals itself, rolled from amount: the HP loss and,
// when it is lethal, an automatic interceptor's save or the asset's removal.
// mutations is the batch the hit joins, read for damage already under way.
func damageSelf(registry *hooks.Registry, factionState *state.FactionState, mutations []domain.Mutation, factionID string, asset *domain.Asset, def *domain.AssetDefinition, amount domain.DiceRoll, roller dice.Roller) []domain.Mutation {
	effectiveHP, removed := batchHP(mutations, asset)
	if removed {
		return nil
	}
	damage := dice.Amount(amount, roller)
	extra := []domain.Mutation{domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: factionID,
	}}
	if effectiveHP-damage > 0 {
		return extra
	}
	// deliberate: only automatic interceptors, not elective ones, because a reactor has no collector to ask (D-10).
	saved := dispatch.InterceptLethalAutomatic(registry, hooks.LethalContext{
		Target:       asset,
		TargetDef:    def,
		HPAfter:      effectiveHP - damage,
		Damage:       damage,
		FactionState: factionState,
		Batch:        append(slices.Clone(mutations), extra...),
	})
	if saved != nil {
		return append(extra, saved...)
	}
	return append(extra, domain.AssetRemoved{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: factionID,
	})
}

// batchHP returns asset's HP once the damage in mutations lands, and whether
// mutations already remove it. A reactor sees state before the batch applies.
func batchHP(mutations []domain.Mutation, asset *domain.Asset) (int, bool) {
	hp := asset.CurrentHP
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetHPDelta:
			if m.AssetID == asset.ID {
				hp += m.Delta
			}
		case domain.AssetRemoved:
			if m.AssetID == asset.ID {
				return hp, true
			}
		}
	}
	return hp, false
}
