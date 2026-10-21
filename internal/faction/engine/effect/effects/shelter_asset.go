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

// shelterAssetRegistrar is Boltholes: an own asset of a listed type on its world
// that damage would destroy is set to 0 HP and sheltered instead, until a repair
// brings it to full HP. If damage destroys the Boltholes first, its sheltered
// assets are destroyed with it. owner = "own" reads as "the struck asset is the
// declaring asset's faction's", scope = "world" as "on the declaring asset's world".
var shelterAssetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerLethalDamage: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeWorld},
			Filters: []domain.EffectFilterName{domain.FilterAssetType},
		},
	},
	Register: registerShelterAsset,
}

func registerShelterAsset(ctx effect.RegisterContext) {
	scope := hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID)
	source := string(domain.SpecialShelterAsset) + ":" + ctx.Def.ID
	ctx.Registry.RegisterLethalInterceptor(scope, source, &shelterAssetInterceptor{
		factionID:  ctx.Faction.ID,
		assetID:    ctx.Asset.ID,
		assetTypes: ctx.Entry.Filters.AssetTypes,
	})
	ctx.Registry.RegisterMutationReactor(scope, source, &shelterLossReactor{assetID: ctx.Asset.ID})
}

type shelterAssetInterceptor struct {
	factionID  string
	assetID    string
	assetTypes []domain.AssetType
}

func (interceptor *shelterAssetInterceptor) OfferInterception(ctx hooks.LethalContext) *hooks.Interception {
	self := liveAsset(ctx.FactionState, interceptor.factionID, interceptor.assetID)
	if !guardsTarget(self, ctx, interceptor.assetTypes) {
		return nil
	}
	cause := domain.SpecialShelterAsset
	return &hooks.Interception{
		Source: self,
		Ask:    hooks.InterceptAutomatic,
		Mutations: append(hpTo(ctx, 0, cause, self.OwnerID), domain.AssetConditionApplied{
			FactionID:         ctx.Target.OwnerID,
			AssetID:           ctx.Target.ID,
			Condition:         domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: self.ID},
			Cause:             string(cause),
			CausedByFactionID: self.OwnerID,
		}),
	}
}

// shelterLossCauses are the removal causes that mean damage destroyed the asset.
var shelterLossCauses = []string{"attack", string(domain.SpecialCounterstrike), string(domain.SpecialSelfDamage)}

// shelterLossReactor destroys the assets its Boltholes shelters when damage
// destroys the Boltholes. A sale or a refit does not (Effort 2 Decision 13).
type shelterLossReactor struct {
	assetID string
}

func (reactor *shelterLossReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if !ok || removed.AssetID != reactor.assetID || !slices.Contains(shelterLossCauses, removed.Cause) {
			continue
		}
		// deliberate: no active check, because a disabled or unready Boltholes that is destroyed still takes its sheltered assets with it.
		sheltered := domain.Condition{Kind: domain.ConditionSheltered, SourceAssetID: reactor.assetID}
		var extra []domain.Mutation
		// deliberate: every faction, not only the Boltholes' owner, because a capture moves the Boltholes but not the assets it shelters.
		for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
			for _, asset := range domain.SortedAssets(factionState.Factions[factionID]) {
				if !slices.ContainsFunc(asset.Conditions, sheltered.Is) && !appliedInBatch(mutations, asset.ID, sheltered) {
					continue
				}
				if domain.RemovesOrMoves(mutations, asset.ID) {
					continue
				}
				// deliberate: copy the Boltholes removal's cause, not one of its own, because a combat kill of the Boltholes is also a combat kill of each asset it held (Effort 2 Decision 13).
				extra = append(extra, domain.AssetRemoved{FactionID: factionID, AssetID: asset.ID, Cause: removed.Cause, CausedByFactionID: removed.CausedByFactionID})
			}
		}
		return extra
	}
	return nil
}

// appliedInBatch reports whether mutations apply condition to assetID. Reactors
// run before the batch applies, so an asset this Boltholes sheltered earlier in
// the same action does not yet carry the condition in state.
func appliedInBatch(mutations []domain.Mutation, assetID string, condition domain.Condition) bool {
	return slices.ContainsFunc(mutations, func(mutation domain.Mutation) bool {
		applied, ok := mutation.(domain.AssetConditionApplied)
		return ok && applied.AssetID == assetID && condition.Is(applied.Condition)
	})
}
