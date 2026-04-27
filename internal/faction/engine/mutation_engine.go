package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// MutationEngine applies mutation lists to campaign state atomically. It is
// the sole writer to campaign state during a turn. The Apply method will
// also trigger history recording when the History engine is built.
type MutationEngine struct{}

func newMutationEngine() *MutationEngine {
	return &MutationEngine{}
}

// Apply applies a list of Mutations to campaign state in order.
func (me *MutationEngine) Apply(factionState *state.FactionState, mutations []domain.Mutation) {
	for _, mutation := range mutations {
		switch v := mutation.(type) {
		case domain.CoinDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.Coin += v.Delta
			}
		case domain.AssetRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				surviving := make([]*domain.Asset, 0, len(faction.Assets))
				for _, asset := range faction.Assets {
					if asset.ID != v.AssetID {
						surviving = append(surviving, asset)
					}
				}
				faction.Assets = surviving
			}
		case domain.AssetMaintainedFlag:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, asset := range faction.Assets {
					if asset.ID == v.AssetID {
						asset.Maintained = v.Maintained
						break
					}
				}
			}
		case domain.FactionHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.CurrentHP += v.Delta
			}
		case domain.AssetHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, asset := range faction.Assets {
					if asset.ID == v.AssetID {
						asset.CurrentHP += v.Delta
						break
					}
				}
			}
		case domain.AssetAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				asset := v.Asset
				faction.Assets = append(faction.Assets, &asset)
			}
		case domain.AssetStealthCleared:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, asset := range faction.Assets {
					if asset.ID == v.AssetID {
						asset.Stealthy = false
						break
					}
				}
			}
		case domain.BaseHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.CurrentHP += v.Delta
						break
					}
				}
			}
		case domain.BaseDestroyed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				surviving := make([]*domain.Base, 0, len(faction.Bases))
				for _, base := range faction.Bases {
					if base.ID != v.BaseID {
						surviving = append(surviving, base)
					}
				}
				faction.Bases = surviving
			}
		case domain.BaseAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				base := v.Base
				faction.Bases = append(faction.Bases, &base)
			}
		case domain.BaseHealed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.CurrentHP += v.Delta
						if max := base.EffectiveMaxHP(faction); base.CurrentHP > max {
							base.CurrentHP = max
						}
						break
					}
				}
			}
		case domain.BaseExpanded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.MaxHP += v.Delta
						base.CurrentHP += v.Delta
						break
					}
				}
			}
		}
	}
}
