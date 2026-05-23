package mutation

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// MutationMiss records one missing-entity hit during Apply.
type MutationMiss struct {
	MutationType string
	FactionID    string
	EntityID     string // asset or base ID; empty for faction-level misses
}

// MutationApplyError is returned by Apply when one or more referenced entities
// were not found. The orchestrator decides whether to treat this as a hard turn
// failure or a warning.
type MutationApplyError struct {
	Misses []MutationMiss
}

func (e *MutationApplyError) Error() string {
	var b strings.Builder
	b.WriteString("MutationApplyError: the following entity references were not found:\n")
	for _, miss := range e.Misses {
		if miss.EntityID != "" {
			fmt.Fprintf(&b, "- Mutation type '%s' references missing entity ID '%s' in faction '%s'\n", miss.MutationType, miss.EntityID, miss.FactionID)
		} else {
			fmt.Fprintf(&b, "- Mutation type '%s' references missing faction ID '%s'\n", miss.MutationType, miss.FactionID)
		}
	}
	return b.String()
}

type MutationEngine struct {
	log *slog.Logger
}

func New(log *slog.Logger) *MutationEngine {
	return &MutationEngine{log: log}
}

func (me *MutationEngine) Apply(factionState *state.FactionState, mutations []domain.Mutation, log *slog.Logger) *MutationApplyError {
	var errs MutationApplyError
	log.Info("applying mutations", "count", len(mutations))

	for _, mutation := range mutations {
		log.Debug("mutation", "type", mutation.Type())
		switch v := mutation.(type) {
		case domain.CoinDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.Coin += v.Delta
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				delete(faction.Assets, v.AssetID)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetMaintainedFlag:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Maintained = v.Maintained
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.FactionHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.CurrentHP += v.Delta
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.CurrentHP += v.Delta
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				asset := v.Asset
				faction.Assets[asset.ID] = &asset
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetStealthCleared:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Stealthy = false
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.BaseHPDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				found := false
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.CurrentHP += v.Delta
						found = true
						break
					}
				}
				if !found {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.BaseID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
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
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.BaseAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				base := v.Base
				faction.Bases = append(faction.Bases, &base)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.BaseHealed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				found := false
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.CurrentHP += v.Delta
						if max := base.EffectiveMaxHP(faction); base.CurrentHP > max {
							base.CurrentHP = max
						}
						found = true
						break
					}
				}
				if !found {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.BaseID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.BaseExpanded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				found := false
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.MaxHP += v.Delta
						base.CurrentHP += v.Delta
						found = true
						break
					}
				}
				if !found {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.BaseID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetMoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Location = v.ToLocation
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetStealthApplied:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Stealthy = true
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalAbandoned:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalCompleted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = nil
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.XPAwarded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.XP += v.Amount
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.HomeworldChanged:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.Homeworld = v.ToWorld
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.TagAdded:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				tag := v.Tag
				faction.Tags = append(faction.Tags, &tag)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.InfluenceDelta:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				found := false
				for _, base := range faction.Bases {
					if base.ID == v.BaseID {
						base.Influence += v.Delta
						found = true
						break
					}
				}
				if !found {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.BaseID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalInitiated:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.ActiveGoal = &domain.ActiveGoal{
					GoalID:       v.GoalID,
					ProcessPhase: v.ProcessPhase,
					TargetWorld:  v.TargetWorld,
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalProgressed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.Progress += v.Delta
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalTurnsTick:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.TurnsRemaining--
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.GoalPhaseAdvanced:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if faction.ActiveGoal != nil && faction.ActiveGoal.GoalID == v.GoalID {
					faction.ActiveGoal.ProcessPhase = v.ProcessPhase
					faction.ActiveGoal.TurnsRemaining = v.TurnsRemaining
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.XPSpent:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.XP -= v.Amount
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.StatRaised:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				switch v.Stat {
				case domain.StatForce:
					faction.Force = v.NewRating
				case domain.StatCunning:
					faction.Cunning = v.NewRating
				case domain.StatWealth:
					faction.Wealth = v.NewRating
				}
				faction.MaxHP = domain.CalcMaxHP(faction)
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderIssued:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					order := v.Order
					asset.CurrentOrder = &order
					asset.Location = domain.Location{
						RegionHex: v.Order.Path[0],
					}
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderProgressed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{
						RegionHex: v.RegionHex,
					}
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderRevised:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					order := v.NewOrder
					asset.CurrentOrder = &order
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderCancelled:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.CurrentOrder = nil
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.MovementOrderCompleted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.CurrentOrder = nil
					asset.Location = v.FinalLocation
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		default:
			panic(fmt.Sprintf("unhandled mutation type: %s", mutation.Type()))
		}
	}

	if len(errs.Misses) == 0 {
		log.Info("mutations applied", "count", len(mutations))
		return nil
	}
	log.Info("mutations applied", "count", len(mutations), "misses", len(errs.Misses))
	return &errs
}
