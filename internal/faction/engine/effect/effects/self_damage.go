package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
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
	},
	Register: registerSelfDamage,
}

func registerSelfDamage(ctx effect.RegisterContext) {
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
	roller   domain.Roller
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
	effectiveHP, removed := batchHP(mutations, asset)
	if removed {
		return nil
	}
	damage := reactor.amount.Roll(reactor.roller)
	extra := []domain.Mutation{domain.AssetHPDelta{
		FactionID:         reactor.holder.factionID,
		AssetID:           reactor.holder.assetID,
		Delta:             -damage,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: reactor.holder.factionID,
	}}
	if effectiveHP-damage <= 0 {
		// deliberate: only automatic interceptors, not elective ones, because a reactor has no collector to ask (D-10).
		saved := dispatch.InterceptLethalAutomatic(reactor.registry, hooks.LethalContext{
			Target:       asset,
			TargetDef:    reactor.def,
			HPAfter:      effectiveHP - damage,
			Damage:       damage,
			FactionState: factionState,
			Batch:        append(slices.Clone(mutations), extra...),
		})
		if saved != nil {
			return append(extra, saved...)
		}
		extra = append(extra, domain.AssetRemoved{
			FactionID:         reactor.holder.factionID,
			AssetID:           reactor.holder.assetID,
			Cause:             string(domain.SpecialSelfDamage),
			CausedByFactionID: reactor.holder.factionID,
		})
	}
	return extra
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
