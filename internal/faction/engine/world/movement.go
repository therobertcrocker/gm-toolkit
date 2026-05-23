package world

import (
	"fmt"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type MovementDecisionKind int

const (
	MovementDecisionIssue MovementDecisionKind = iota
	MovementDecisionRevise
	MovementDecisionCancel
)

type MovementDecision struct {
	AssetID       string
	Kind          MovementDecisionKind
	Destination   *domain.Location // nil for Cancel
	CargoAssetIDs []string
}

func (engine *WorldEngine) TickMovementOrders(faction *domain.Faction, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation
	log.Info("ticking movement orders")

	for _, asset := range faction.Assets {
		if asset.CurrentOrder == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return nil, fmt.Errorf("TickMovementOrders: unknown asset definition %q", asset.DefinitionID)
		}
		newStepIdx := asset.CurrentOrder.StepIdx + def.Speed
		log.Debug("ticking movement order for asset", "asset_id", asset.ID, "from_hex", asset.Location.RegionHex, "to_hex", asset.CurrentOrder.Destination.RegionHex)
		if newStepIdx >= len(asset.CurrentOrder.Path)-1 {
			mutations = append(mutations, domain.MovementOrderCompleted{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				FinalLocation:     asset.CurrentOrder.Destination,
				Cause:             domain.CauseMovementTick,
				CausedByFactionID: faction.ID,
			})
		} else {
			step := asset.CurrentOrder.Path[newStepIdx]
			mutations = append(mutations, domain.MovementOrderProgressed{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				NewStepIdx:        newStepIdx,
				RegionHex:         step,
				Cause:             domain.CauseMovementTick,
				CausedByFactionID: faction.ID,
			})
		}
	}

	log.Info("movement orders ticked", "mutations", len(mutations))
	return mutations, nil
}

func (engine *WorldEngine) BuildMovementMutations(decisions []MovementDecision, faction *domain.Faction, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation

	for _, decision := range decisions {
		asset, ok := faction.Assets[decision.AssetID]
		if !ok {
			return nil, fmt.Errorf("movement decision for unknown asset ID %q", decision.AssetID)
		}
		switch decision.Kind {
		case MovementDecisionIssue:
			def, ok := rulebook.Assets[asset.DefinitionID]
			if !ok {
				return nil, fmt.Errorf("BuildMovementMutations: unknown asset definition %q", asset.DefinitionID)
			}
			crossingCost := rulebook.DriftCost(def.DriftRating)

			destLoc, ok := engine.spatialMap.Location(decision.Destination.WorldID)
			if !ok {
				return nil, fmt.Errorf("movement: unknown destination world %q", decision.Destination.WorldID)
			}
			destHex, ok := destLoc.(spatial.RegionLocation)
			if !ok {
				return nil, fmt.Errorf("movement: destination %q is not a hex location", decision.Destination.WorldID)
			}
			from := asset.Location.RegionHex
			to := destHex.RegionHex()

			path, _, err := engine.spatialMap.Path(from, to, crossingCost)
			if err != nil {
				return nil, fmt.Errorf("movement: %w", err)
			}
			order := domain.MovementOrder{
				AssetID:       asset.ID,
				Origin:        asset.Location,
				Destination:   *decision.Destination,
				Path:          path,
				StepIdx:       0,
				DriftRating:   def.DriftRating,
				CargoAssetIDs: decision.CargoAssetIDs,
			}
			mutations = append(mutations, domain.MovementOrderIssued{
				FactionID:         faction.ID,
				AssetID:           asset.ID,
				Order:             order,
				Cause:             domain.CauseMovementIssue,
				CausedByFactionID: faction.ID,
			})

		case MovementDecisionRevise:
			if asset.CurrentOrder == nil {
				return nil, fmt.Errorf("movement revision: asset %q has no current order", asset.ID)
			}
			def, ok := rulebook.Assets[asset.DefinitionID]
			if !ok {
				return nil, fmt.Errorf("BuildMovementMutations: unknown asset definition %q", asset.DefinitionID)
			}
			crossingCost := rulebook.DriftCost(def.DriftRating)
			from := asset.Location.RegionHex
			to := decision.Destination.RegionHex
			newPath, _, err := engine.spatialMap.Path(from, to, crossingCost)
			if err != nil {
				return nil, fmt.Errorf("movement revision: %w", err)
			}

			newOrder := domain.MovementOrder{
				AssetID:       asset.ID,
				Origin:        asset.Location,
				Destination:   *decision.Destination,
				Path:          newPath,
				DriftRating:   def.DriftRating,
				CargoAssetIDs: asset.CurrentOrder.CargoAssetIDs,
			}
			mutations = append(mutations,
				domain.CoinDelta{FactionID: faction.ID, Delta: -1, Cause: domain.CauseMovementRevision, CausedByFactionID: faction.ID},
				domain.MovementOrderRevised{
					FactionID:         faction.ID,
					AssetID:           asset.ID,
					NewOrder:          newOrder,
					Cause:             domain.CauseMovementRevision,
					CausedByFactionID: faction.ID,
				},
			)

		case MovementDecisionCancel:
			if asset.CurrentOrder == nil {
				return nil, fmt.Errorf("movement cancellation: asset %q has no current order", asset.ID)
			}
			mutations = append(mutations, domain.MovementOrderCancelled{
				FactionID:         faction.ID,
				AssetID:           asset.ID,
				StrandedAt:        asset.Location,
				Cause:             domain.CauseMovementCancellation,
				CausedByFactionID: faction.ID,
			})
		}
	}
	return mutations, nil
}
