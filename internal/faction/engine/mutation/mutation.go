package mutation

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// MutationMiss records one missing-entity hit during Apply.
type MutationMiss struct {
	MutationType string
	FactionID    string
	EntityID     string // asset, base or tag ID; empty for faction-level misses
}

// MutationApplyError is returned by Apply when one or more referenced entities
// were not found. Fatal at the orchestrator: applyAndRecord emits one
// structured Error-level log line per Miss and propagates the error to abort
// the turn.
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

	// deliberate: owner changes apply after the rest of the batch, not in emitted order, because reactors read pre-apply state and address the asset's old owner (overview Decision 5).
	ordered := make([]domain.Mutation, 0, len(mutations))
	var ownerChanges []domain.Mutation
	for _, mutation := range mutations {
		if _, ok := mutation.(domain.AssetOwnerChanged); ok {
			ownerChanges = append(ownerChanges, mutation)
			continue
		}
		ordered = append(ordered, mutation)
	}
	for _, mutation := range append(ordered, ownerChanges...) {
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
		case domain.PurchaseDiscountGranted:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = v.Amount
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.PurchaseDiscountConsumed:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				faction.PendingPurchaseDiscount = 0
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
		case domain.TagRemoved:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				index := slices.IndexFunc(faction.Tags, func(tag *domain.Tag) bool { return tag.ID == v.Tag.ID })
				if index >= 0 {
					faction.Tags = slices.Delete(faction.Tags, index, index+1)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.Tag.ID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.WorldGovernmentChanged:
			_, known := factionState.Factions[v.ToFactionID]
			switch {
			case v.ToFactionID == "":
				delete(factionState.Governments, v.WorldID)
			case !known:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.ToFactionID})
			default:
				if factionState.Governments == nil {
					factionState.Governments = make(map[string]string)
				}
				factionState.Governments[v.WorldID] = v.ToFactionID
			}
		case domain.PermissionRefused:
			// record only: the refused action bought nothing
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
		case domain.AssetConditionApplied:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Conditions = append(slices.DeleteFunc(asset.Conditions, v.Condition.Is), v.Condition)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetConditionCleared:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok {
					asset.Conditions = slices.DeleteFunc(asset.Conditions, v.Condition.Is)
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
			} else {
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			}
		case domain.AssetOwnerChanged:
			from, fromOK := factionState.Factions[v.FactionID]
			to, toOK := factionState.Factions[v.ToFactionID]
			switch {
			case !fromOK:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID})
			case !toOK:
				errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.ToFactionID})
			default:
				if asset, ok := from.Assets[v.AssetID]; ok {
					delete(from.Assets, v.AssetID)
					asset.OwnerID = v.ToFactionID
					if to.Assets == nil {
						to.Assets = make(map[string]*domain.Asset)
					}
					to.Assets[v.AssetID] = asset
				} else {
					errs.Misses = append(errs.Misses, MutationMiss{MutationType: mutation.Type(), FactionID: v.FactionID, EntityID: v.AssetID})
				}
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
		case domain.MovementOrderHeld:
			if faction, ok := factionState.Factions[v.FactionID]; ok {
				if asset, ok := faction.Assets[v.AssetID]; ok && asset.CurrentOrder != nil {
					asset.CurrentOrder.StepIdx = v.NewStepIdx
					asset.Location = domain.Location{RegionHex: v.RegionHex}
					if v.Reason == domain.HoldLockdownPaid {
						asset.CurrentOrder.LockdownPaid = true
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
