package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// active reports whether asset can use its special ability now: an asset is not
// usable until the turn after purchase, while unmaintained, or once destroyed.
// Reactors check at reaction time, not at registration, because Ready and
// Maintained both change mid-cycle.
func active(asset *domain.Asset) bool {
	return asset != nil && asset.Ready && asset.Maintained && asset.CurrentHP > 0
}

// liveAsset returns the declaring asset from current state, or nil when its
// faction or the asset is gone. A nil state (a dispatcher test) is no asset.
func liveAsset(factionState *state.FactionState, factionID, assetID string) *domain.Asset {
	if factionState == nil {
		return nil
	}
	faction := factionState.Factions[factionID]
	if faction == nil {
		return nil
	}
	return faction.Assets[assetID]
}
