package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// captureTargetRegistrar is Treachery: on a hit, the declaring asset is spent, its
// owner gains Coin, and the asset it hit changes sides at its current HP.
var captureTargetRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnAttackHit: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf},
		},
	},
	Register: registerCaptureTarget,
}

func registerCaptureTarget(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(
		hooks.AssetScope(ctx.Faction.ID, ctx.Asset.ID),
		string(domain.SpecialCaptureTarget)+":"+ctx.Def.ID,
		&captureTargetReactor{
			factionID: ctx.Faction.ID,
			assetID:   ctx.Asset.ID,
			coin:      ctx.Entry.CaptureTarget.Coin,
			roller:    ctx.Roller,
		},
	)
}

type captureTargetReactor struct {
	factionID string
	assetID   string
	coin      domain.DiceRoll
	roller    domain.Roller
}

func (reactor *captureTargetReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	self := liveAsset(factionState, reactor.factionID, reactor.assetID)
	if !active(self) {
		return nil
	}
	for _, found := range hits(mutations, false) {
		if found.strikerID != reactor.assetID || found.targetAssetID == "" {
			continue
		}
		target := liveAsset(factionState, found.targetFactionID, found.targetAssetID)
		if target == nil {
			return nil
		}
		// deliberate: fizzle when the batch removes the target, not capture, because the owner change applies after the removal and would miss. A removal from a sibling reactor this round (Zealots) is not visible here and still misses.
		if _, removed := batchHP(mutations, target); removed {
			return nil
		}
		cause := string(domain.SpecialCaptureTarget)
		var extra []domain.Mutation
		// deliberate: skip the self-removal when the batch already removes this asset, not emit it twice, because on a tie the defender's counter can kill Treachery in the matchup it captures in (Decision 15).
		if _, removed := batchHP(mutations, self); !removed {
			extra = append(extra, domain.AssetRemoved{FactionID: reactor.factionID, AssetID: reactor.assetID, Cause: cause, CausedByFactionID: reactor.factionID})
		}
		return append(extra,
			domain.CoinDelta{FactionID: reactor.factionID, Delta: reactor.coin.Roll(reactor.roller), Cause: cause, CausedByFactionID: reactor.factionID},
			domain.AssetOwnerChanged{FactionID: found.targetFactionID, AssetID: target.ID, ToFactionID: reactor.factionID, Cause: cause, CausedByFactionID: reactor.factionID},
		)
	}
	return nil
}
