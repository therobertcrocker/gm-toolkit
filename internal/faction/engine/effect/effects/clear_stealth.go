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
// declaring asset's world". stealth_arrival is a reactor that covers a stealthed
// landing and a stealthed purchase; turn_start tests every such asset when the
// owner's turn begins.
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
		factionID, assetID, ok := stealthArrival(mutation, mutations, factionState, worldID)
		if !ok || factionID == owner.ID {
			continue
		}
		target := factionState.Factions[factionID]
		if target == nil {
			continue
		}
		// deliberate: compare true, not the state's flag, because a bought asset is stealthed by this batch and is not in state yet.
		if reactor.stealthy != nil && !*reactor.stealthy {
			continue
		}
		if reactor.contest != nil && !contestPasses(*reactor.contest, owner, target, reactor.roller) {
			continue
		}
		extra = append(extra, domain.AssetStealthCleared{
			FactionID:         factionID,
			AssetID:           assetID,
			Cause:             string(domain.SpecialClearStealth),
			CausedByFactionID: owner.ID,
		})
	}
	return extra
}

// stealthArrival reports the asset a mutation brings onto worldID stealthed. A
// landing is an AssetMoved into the world from elsewhere by an asset already
// stealthy (Decision Record 10). A purchase is an AssetStealthApplied whose asset
// the same batch adds on the world; a stealth applied to an asset already there is
// not an arrival.
func stealthArrival(mutation domain.Mutation, batch []domain.Mutation, factionState *state.FactionState, worldID string) (string, string, bool) {
	switch m := mutation.(type) {
	case domain.AssetMoved:
		if m.ToLocation.WorldID != worldID || m.FromLocation.WorldID == worldID {
			return "", "", false
		}
		faction := factionState.Factions[m.FactionID]
		if faction == nil || faction.Assets[m.AssetID] == nil || !faction.Assets[m.AssetID].Stealthy {
			return "", "", false
		}
		return m.FactionID, m.AssetID, true
	case domain.AssetStealthApplied:
		for _, other := range batch {
			added, ok := other.(domain.AssetAdded)
			if ok && added.FactionID == m.FactionID && added.Asset.ID == m.AssetID && added.Asset.Location.WorldID == worldID {
				return m.FactionID, m.AssetID, true
			}
		}
	}
	return "", "", false
}
