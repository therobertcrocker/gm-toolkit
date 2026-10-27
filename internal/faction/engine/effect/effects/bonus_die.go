package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world;
// "faction" any owner roll anywhere. A tag's die is FromTag, so Blackmail strips it.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
			Gates:   []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerBonusDie,
}

func registerBonusDie(ctx effect.RegisterContext) {
	ctx.Registry.RegisterRollModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialBonusDie),
		&bonusDieModifier{
			holder:  refOf(ctx.Holder),
			source:  ctx.Holder.Name(),
			trigger: ctx.Entry.Trigger,
			filters: ctx.Entry.Filters,
			oncePer: ctx.Entry.Gates.OncePer,
			params:  *ctx.Entry.BonusDie,
		},
	)
}

type bonusDieModifier struct {
	holder  holderRef
	source  string
	trigger domain.EffectTrigger
	filters domain.EffectFilters
	oncePer domain.OncePerKey
	params  domain.BonusDieParams
}

func (modifier *bonusDieModifier) OfferModifiers(ctx hooks.RollContext, factionState *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	asset, live := modifier.holder.live(factionState)
	if !live || !rollMatches(ctx, modifier.trigger, modifier.holder.factionID, asset, modifier.filters) {
		return nil
	}
	params := modifier.params
	// deliberate: one offer per asset copy, not one per definition, because each copy grants its own die (Effort 3 plan, Decision Record 12).
	offer := hooks.ModifierOffer{
		Source:      modifier.source,
		Description: fmt.Sprintf("roll %d extra d%d, keep highest", params.Count, params.Sides),
		FromTag:     modifier.holder.tagID != "",
		Apply: func(rollState *hooks.RollState) {
			for range params.Count {
				rollState.AddDie(params.Sides)
			}
			// deliberate: keep 1, not 1 + count, because every faction roll is one d10 and an extra die is pick-the-best ("an additional d10, keep the highest").
			rollState.SetKeepHighest(1)
		},
	}
	if modifier.oncePer != "" {
		offer.BudgetKey = budgetKey(modifier.oncePer, domain.SpecialBonusDie, modifier.trigger, modifier.holder.sourceID, modifier.holder.assetID, otherSideID(ctx, false))
	}
	return []hooks.ModifierOffer{offer}
}
