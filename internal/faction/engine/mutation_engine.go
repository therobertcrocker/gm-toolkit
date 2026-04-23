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
			for _, faction := range factionState.Factions {
				if faction.ID == v.FactionID {
					faction.Coin += v.Delta
					break
				}
			}
		case domain.AssetRemoved:
			for _, faction := range factionState.Factions {
				if faction.ID == v.FactionID {
					surviving := make([]*domain.Asset, 0, len(faction.Assets))
					for _, asset := range faction.Assets {
						if asset.ID != v.AssetID {
							surviving = append(surviving, asset)
						}
					}
					faction.Assets = surviving
					break
				}
			}
		case domain.AssetMaintainedFlag:
			for _, faction := range factionState.Factions {
				if faction.ID == v.FactionID {
					for _, asset := range faction.Assets {
						if asset.ID == v.AssetID {
							asset.Maintained = v.Maintained
							break
						}
					}
					break
				}
			}
		}
	}
}
