package effects

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// clearStealthRegistrar reads owner = "rival" as "the affected asset is a
// rival's", and scope = "world" as "it stands on (stealth_arrival: lands on) the
// declaring asset's world". stealth_arrival is a reactor; turn_start tests every
// such asset when the owner's turn begins. Arrival by purchase is Effort 5's, with
// the route that makes a purchase stealthy.
var clearStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerStealthArrival: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
		domain.TriggerTurnStart: {
			Owners:  []domain.EffectOwner{domain.OwnerRival},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterStealthy},
			Gates:   []domain.EffectGateName{domain.GateContest},
		},
	},
	Register: registerClearStealth,
	Fire:     fireClearStealth,
}

// fireClearStealth tests every rival asset on the declaring asset's world, in
// faction then asset ID order.
func fireClearStealth(ctx effect.FireContext) []domain.Mutation {
	worldID := ctx.Asset.Location.WorldID
	if worldID == "" {
		return nil
	}
	var cleared []domain.Mutation
	for _, factionID := range slices.Sorted(maps.Keys(ctx.FactionState.Factions)) {
		if factionID == ctx.Faction.ID {
			continue
		}
		target := ctx.FactionState.Factions[factionID]
		for _, asset := range domain.SortedAssets(target) {
			if asset.Location.WorldID != worldID || !asset.Stealthy {
				continue
			}
			if ctx.Entry.Filters.Stealthy != nil && asset.Stealthy != *ctx.Entry.Filters.Stealthy {
				continue
			}
			if ctx.Entry.Gates.Contest != nil && !contestPasses(*ctx.Entry.Gates.Contest, ctx.Faction, target, ctx.Roller) {
				continue
			}
			cleared = append(cleared, domain.AssetStealthCleared{
				FactionID:         target.ID,
				AssetID:           asset.ID,
				Cause:             string(domain.SpecialClearStealth),
				CausedByFactionID: ctx.Faction.ID,
			})
		}
	}
	return cleared
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
		if target == nil || target.Assets[moved.AssetID] == nil || !target.Assets[moved.AssetID].Stealthy {
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
