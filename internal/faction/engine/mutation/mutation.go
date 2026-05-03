package mutation

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type MutationEngine struct{}

func New() *MutationEngine {
	return &MutationEngine{}
}

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
		case domain.AssetMoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, asset := range faction.Assets {
					if asset.ID == v.AssetID {
						asset.Location = v.ToLocation
						break
					}
				}
			}
		case domain.AssetStealthApplied:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, asset := range faction.Assets {
					if asset.ID == v.AssetID {
						asset.Stealthy = true
						break
					}
				}
			}
		case domain.GoalAbandoned:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			}
		case domain.GoalCompleted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			}
		case domain.XPAwarded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.XP += v.Amount
			}
		case domain.HomeworldChanged:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.Homeworld = v.ToWorld
			}
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				tag := v.Tag
				faction.Tags = append(faction.Tags, &tag)
			}
		case domain.InfluenceDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.Influence += v.Delta
						break
					}
				}
			}
		case domain.GoalInitiated:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = &domain.ActiveGoal{
					GoalID:       v.GoalID,
					ProcessPhase: v.ProcessPhase,
					TargetWorld:  v.TargetWorld,
				}
			}
		case domain.GoalProgressed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.Progress += v.Delta
				}
			}
		case domain.GoalTurnsTick:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.TurnsRemaining--
				}
			}
		case domain.GoalPhaseAdvanced:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.ProcessPhase = v.ProcessPhase
					faction.ActiveGoal.TurnsRemaining = v.TurnsRemaining
				}
			}
		default:
			panic(fmt.Sprintf("unhandled mutation type: %s", mutation.Type()))
		}
	}
}
