package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// holderRef is what a hook keeps of its holder: IDs, read again at fire time,
// because a held pointer keeps a removed asset in effect (Decision Record 19).
// assetID is empty for a tag; tagID is empty for an asset. sourceID is the
// definition or tag ID that budget keys name.
type holderRef struct {
	factionID string
	assetID   string
	tagID     string
	sourceID  string
}

func refOf(holder effect.Holder) holderRef {
	ref := holderRef{factionID: holder.Faction.ID, sourceID: holder.ID()}
	if holder.Asset != nil {
		ref.assetID = holder.Asset.ID
	}
	if holder.Tag != nil {
		ref.tagID = holder.Tag.ID
	}
	return ref
}

// live returns the declaring asset from current state, and whether the holder
// still applies: its asset is active, or its faction still holds the tag. A tag
// has no asset. A nil state (a dispatcher test) is no holder.
func (ref holderRef) live(factionState *state.FactionState) (*domain.Asset, bool) {
	if factionState == nil {
		return nil, false
	}
	return ref.liveIn(factionState.Factions[ref.factionID])
}

// liveIn is live for a Cat 4 hook, which gets no FactionState: it reads the owner
// it kept from registration, whose maps stay live all cycle.
func (ref holderRef) liveIn(owner *domain.Faction) (*domain.Asset, bool) {
	if owner == nil {
		return nil, false
	}
	if ref.tagID != "" {
		return nil, owner.HasTag(ref.tagID)
	}
	asset := owner.Assets[ref.assetID]
	return asset, active(asset)
}

// refuseTagHolder is the Validate of a registrar whose faction scope still reads
// the declaring asset, so a tag cannot hold it yet (Effort 1 Decision 2).
func refuseTagHolder(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder == effect.HolderTag {
		return fmt.Errorf("a tag cannot hold this effect yet: it reads the declaring asset")
	}
	return nil
}
