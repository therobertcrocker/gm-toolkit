package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
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
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialLockOut), &lockOutReactor{holder: refOf(ctx.Holder)})
}

type lockOutReactor struct {
	holder holderRef
}

func (reactor *lockOutReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	self, live := reactor.holder.live(factionState)
	if !live {
		return nil
	}
	// deliberate: no lock from a Lockdown the batch removes or captures, because its condition would apply after its removal and miss.
	if domain.RemovesOrMoves(mutations, reactor.holder.assetID) {
		return nil
	}
	var struck []string
	var locks []domain.Mutation
	for _, targetID := range hitTargets(mutations, reactor.holder.assetID, false) {
		if slices.Contains(struck, targetID) {
			continue
		}
		struck = append(struck, targetID)
		locks = append(locks, domain.AssetConditionApplied{
			FactionID: reactor.holder.factionID,
			AssetID:   reactor.holder.assetID,
			Condition: domain.Condition{
				Kind:            domain.ConditionLockedOut,
				SourceAssetID:   reactor.holder.assetID,
				TargetFactionID: targetID,
				// deliberate: the hit's world, not wherever the Lockdown goes later, because the book locks "that planet" (Effort 5 Decision Record 5).
				WorldID: self.Location.WorldID,
			},
			Cause:             string(domain.SpecialLockOut),
			CausedByFactionID: reactor.holder.factionID,
		})
	}
	return locks
}
