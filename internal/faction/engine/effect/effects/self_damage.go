package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialSelfDamage)+":"+ctx.Def.ID,
		&selfDamageReactor{
			factionID: ctx.Faction.ID,
			assetID:   ctx.Asset.ID,
			counter:   ctx.Entry.Trigger == domain.TriggerOwnCounterHit,
			amount:    ctx.Entry.SelfDamage.Amount,
			roller:    ctx.Roller,
		},
	)
}

type selfDamageReactor struct {
	factionID string
	assetID   string
	counter   bool
	amount    domain.DiceRoll
	roller    domain.Roller
}

func (reactor *selfDamageReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if len(hitTargets(mutations, reactor.assetID, reactor.counter)) == 0 {
		return nil
	}
	faction := factionState.Factions[reactor.factionID]
	if faction == nil {
		return nil
	}
	asset := faction.Assets[reactor.assetID]
	if !active(asset) {
		return nil
	}
	effectiveHP, removed := batchHP(mutations, asset)
	if removed {
		return nil
	}
	damage := reactor.amount.Roll(reactor.roller)
	extra := []domain.Mutation{domain.AssetHPDelta{
		FactionID:         reactor.factionID,
		AssetID:           reactor.assetID,
		Delta:             -damage,
		Cause:             string(domain.SpecialSelfDamage),
		CausedByFactionID: reactor.factionID,
	}}
	if effectiveHP-damage <= 0 {
		extra = append(extra, domain.AssetRemoved{
			FactionID:         reactor.factionID,
			AssetID:           reactor.assetID,
			Cause:             string(domain.SpecialSelfDamage),
			CausedByFactionID: reactor.factionID,
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
