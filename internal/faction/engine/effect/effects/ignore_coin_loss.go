package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
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
}

func registerIgnoreCoinLoss(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialIgnoreCoinLoss)+":"+ctx.Def.ID,
		&ignoreCoinLossReactor{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			defID:          ctx.Def.ID,
			oncePer:        ctx.Entry.Gates.OncePer,
		},
	)
}

type ignoreCoinLossReactor struct {
	ownerFactionID string
	assetID        string
	defID          string
	oncePer        domain.OncePerKey
}

// OnMutations absorbs each loss another faction imposed on the owner, one per
// copy whose budget is unspent, by paying the loss back.
func (reactor *ignoreCoinLossReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	owner := factionState.Factions[reactor.ownerFactionID]
	if owner == nil {
		return nil
	}
	copies := assetsOfDefinition(owner, reactor.defID)
	// deliberate: only the lowest-ID copy reacts, for all copies, because sibling reactors cannot see each other's output in one dispatch round and would absorb the same loss twice.
	if len(copies) == 0 || copies[0].ID != reactor.assetID {
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
			for next < len(copies) && !spendBudget(owner, budgetKey(reactor.oncePer, domain.SpecialIgnoreCoinLoss, reactor.defID, copies[next].ID, loss.CausedByFactionID)) {
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

func assetsOfDefinition(faction *domain.Faction, defID string) []*domain.Asset {
	var matching []*domain.Asset
	for _, asset := range domain.SortedAssets(faction) {
		if asset.DefinitionID == defID {
			matching = append(matching, asset)
		}
	}
	return matching
}
