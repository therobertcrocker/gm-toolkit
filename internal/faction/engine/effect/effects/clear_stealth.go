package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// clearStealthRegistrar reads owner = "rival" as "the arriving asset is a
// rival's", and scope = "world" as "it lands on the declaring asset's world".
// Arrival by purchase is Effort 5's, with the route that makes a purchase stealthy.
var clearStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerStealthArrival: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerClearStealth,
}

func registerClearStealth(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialClearStealth)+":"+ctx.Def.ID,
		&clearStealthReactor{
			ownerFactionID: ctx.Faction.ID,
			assetID:        ctx.Asset.ID,
			stealthy:       ctx.Entry.Filters.Stealthy,
			contest:        ctx.Entry.Gates.Contest,
			roller:         ctx.Roller,
		},
	)
}

type clearStealthReactor struct {
	ownerFactionID string
	assetID        string
	stealthy       *bool
	contest        *domain.EffectContest
	roller         domain.Roller
}

func (reactor *clearStealthReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	owner := factionState.Factions[reactor.ownerFactionID]
	if owner == nil || !active(owner.Assets[reactor.assetID]) {
		return nil
	}
	worldID := owner.Assets[reactor.assetID].Location.WorldID
	if worldID == "" {
		return nil
	}
	var extra []domain.Mutation
	for _, mutation := range mutations {
		moved, ok := mutation.(domain.AssetMoved)
		if !ok || moved.ToLocation.WorldID != worldID || moved.FromLocation.WorldID == worldID || moved.FactionID == owner.ID {
			continue
		}
		target := factionState.Factions[moved.FactionID]
		if target == nil || target.Assets[moved.AssetID] == nil {
			continue
		}
		if reactor.stealthy != nil && target.Assets[moved.AssetID].Stealthy != *reactor.stealthy {
			continue
		}
		if reactor.contest != nil && !contestPasses(*reactor.contest, owner, target, reactor.roller) {
			continue
		}
		extra = append(extra, domain.AssetStealthCleared{
			FactionID:         target.ID,
			AssetID:           moved.AssetID,
			Cause:             string(domain.SpecialClearStealth),
			CausedByFactionID: owner.ID,
		})
	}
	return extra
}
