package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// lockOutRegistrar is Transport Lockdown: on a hit, the faction it struck pays 1d4
// Coin and waits a turn for each arrival on the Lockdown's world, and cannot
// relocate there, while the Lockdown stays active on that world (D-31 to D-33).
// The record is a locked_out condition the Lockdown carries on itself, naming the
// struck faction and the world of the hit (D-36, overview Decision 24). A hit on a
// Base counts: the book's attack is "against a rival faction".
var lockOutRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerLockOut,
}

func registerLockOut(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialLockOut)+":"+ctx.Def.ID,
		&lockOutReactor{factionID: ctx.Faction.ID, assetID: ctx.Asset.ID},
	)
}

type lockOutReactor struct {
	factionID string
	assetID   string
}

func (reactor *lockOutReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	self := liveAsset(factionState, reactor.factionID, reactor.assetID)
	if !active(self) {
		return nil
	}
	// deliberate: no lock from a Lockdown the batch removes or captures, because its condition would apply after its removal and miss.
	if domain.RemovesOrMoves(mutations, reactor.assetID) {
		return nil
	}
	var struck []string
	var locks []domain.Mutation
	for _, targetID := range hitTargets(mutations, reactor.assetID, false) {
		if slices.Contains(struck, targetID) {
			continue
		}
		struck = append(struck, targetID)
		locks = append(locks, domain.AssetConditionApplied{
			FactionID: reactor.factionID,
			AssetID:   reactor.assetID,
			Condition: domain.Condition{
				Kind:            domain.ConditionLockedOut,
				SourceAssetID:   reactor.assetID,
				TargetFactionID: targetID,
				// deliberate: the hit's world, not wherever the Lockdown goes later, because the book locks "that planet" (Effort 5 Decision Record 5).
				WorldID: self.Location.WorldID,
			},
			Cause:             string(domain.SpecialLockOut),
			CausedByFactionID: reactor.factionID,
		})
	}
	return locks
}
