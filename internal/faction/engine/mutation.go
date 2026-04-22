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
func (m *MutationEngine) Apply(s *state.FactionState, mutations []domain.Mutation) {
	for _, mut := range mutations {
		switch v := mut.(type) {
		case domain.CoinDelta:
			for _, f := range s.Factions {
				if f.ID == v.FactionID {
					f.Coin += v.Delta
					break
				}
			}
		case domain.AssetRemoved:
			for _, f := range s.Factions {
				if f.ID == v.FactionID {
					surviving := make([]*domain.Asset, 0, len(f.Assets))
					for _, a := range f.Assets {
						if a.ID != v.AssetID {
							surviving = append(surviving, a)
						}
					}
					f.Assets = surviving
					break
				}
			}
		case domain.AssetMaintainedFlag:
			for _, f := range s.Factions {
				if f.ID == v.FactionID {
					for _, a := range f.Assets {
						if a.ID == v.AssetID {
							a.Maintained = v.Maintained
							break
						}
					}
					break
				}
			}
		}
	}
}
