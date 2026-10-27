package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// ignoreCoinLossRegistrar reads owner = "rival" as "another faction imposed the
// loss", and scope = "faction" as "a loss anywhere in the owner's faction".
var ignoreCoinLossRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerCoinLoss: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
	},
	Register: registerIgnoreCoinLoss,
	// deliberate: refuse a tag, not accept it, because only the lowest-ID active copy reacts and a tag has no copies (Effort 1 Decision 2).
	Validate: refuseTagHolder,
}

func registerIgnoreCoinLoss(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialIgnoreCoinLoss),
		&ignoreCoinLossReactor{
			holder:  refOf(ctx.Holder),
			trigger: ctx.Entry.Trigger,
			oncePer: ctx.Entry.Gates.OncePer,
		},
	)
}

type ignoreCoinLossReactor struct {
	holder  holderRef
	trigger domain.EffectTrigger
	oncePer domain.OncePerKey
}

// OnMutations absorbs each loss another faction imposed on the owner, one per
// copy whose budget is unspent, by paying the loss back.
func (reactor *ignoreCoinLossReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	owner := factionState.Factions[reactor.holder.factionID]
	if owner == nil {
		return nil
	}
	copies := activeAssetsOfDefinition(owner, reactor.holder.sourceID)
	// deliberate: only the lowest-ID active copy reacts, for all active copies, because sibling reactors cannot see each other's output in one dispatch round and would absorb the same loss twice.
	if len(copies) == 0 || copies[0].ID != reactor.holder.assetID {
		return nil
	}
	var extra []domain.Mutation
	next := 0
	for _, mutation := range mutations {
		loss, ok := mutation.(domain.CoinDelta)
		if !ok || loss.FactionID != owner.ID || loss.Delta >= 0 || loss.CausedByFactionID == "" || loss.CausedByFactionID == owner.ID {
			continue
		}
		if reactor.oncePer != "" {
			for next < len(copies) && !spendBudget(owner, budgetKey(reactor.oncePer, domain.SpecialIgnoreCoinLoss, reactor.trigger, reactor.holder.sourceID, copies[next].ID, loss.CausedByFactionID)) {
				next++
			}
			if next == len(copies) {
				break
			}
			next++
		}
		extra = append(extra, domain.CoinDelta{
			FactionID:         owner.ID,
			Delta:             -loss.Delta,
			Cause:             "bank_absorbed",
			CausedByFactionID: owner.ID,
		})
	}
	return extra
}

func activeAssetsOfDefinition(faction *domain.Faction, defID string) []*domain.Asset {
	var matching []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		if asset.DefinitionID == defID && active(asset) {
			matching = append(matching, asset)
		}
	}
	return matching
}
