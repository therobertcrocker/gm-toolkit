package world

import (
	"fmt"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
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
	RunHot        bool
	Destination   *domain.Location // nil for Cancel
	CargoAssetIDs []string
}

// TickMovementOrders spends each open order's banked points plus the asset's
// speed, as registry's speed floors resolve it (Mercenary Group): a hex step costs 1, a Drift jump its rated cost, and what is left is banked.
func (engine *WorldEngine) TickMovementOrders(faction *domain.Faction, rulebook *rulebook.Rulebook, registry *hooks.Registry, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation
	log.Info("ticking movement orders")

	speedOf := dispatch.AssetSpeedResolver(registry, faction)
	for _, asset := range domain.SortedAssets(faction) {
		if asset.CurrentOrder == nil {
			continue
		}
		def, ok := rulebook.Assets[asset.DefinitionID]
		if !ok {
			return nil, fmt.Errorf("TickMovementOrders: unknown asset definition %q", asset.DefinitionID)
		}
		order := asset.CurrentOrder
		points := order.Banked + speedOf(def, order.RunHot)
		newStepIdx := order.StepIdx
		for newStepIdx < len(order.Path)-1 {
			cost := stepCost(order.Path[newStepIdx], order.Path[newStepIdx+1], rulebook, order.DriftRating)
			if cost > points {
				break
			}
			points -= cost
			newStepIdx++
		}
		log.Debug("ticking movement order for asset", "asset_id", asset.ID, "from_hex", asset.Location.RegionHex, "to_hex", order.Destination.RegionHex, "banked", points)
		if newStepIdx >= len(order.Path)-1 {
			mutations = append(mutations, domain.MovementOrderCompleted{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				FinalLocation:     asset.CurrentOrder.Destination,
				Cause:             domain.CauseMovementTick,
				CausedByFactionID: faction.ID,
			})
		} else {
			step := order.Path[newStepIdx]
			mutations = append(mutations, domain.MovementOrderProgressed{
				FactionID:         asset.OwnerID,
				AssetID:           asset.ID,
				NewStepIdx:        newStepIdx,
				RegionHex:         step,
				Banked:            points,
				Cause:             domain.CauseMovementTick,
				CausedByFactionID: faction.ID,
			})
		}
	}

	log.Info("movement orders ticked", "mutations", len(mutations))
	return mutations, nil
}

// stepCost is the movement points one step of a path takes: 1 for a neighbouring
// hex, the Drift cost of driftRating for a lane jump.
func stepCost(from, to spatial.RegionHex, rulebook *rulebook.Rulebook, driftRating int) int {
	if spatial.Adjacent(from.Coord, to.Coord) {
		return 1
	}
	return rulebook.DriftCost(driftRating)
}

// BuildMovementMutations issues, revises and cancels movement orders. A crossing
// is priced from the drift rating registry's drift-rating hooks resolve
// (Drift-Runners).
func (engine *WorldEngine) BuildMovementMutations(decisions []MovementDecision, faction *domain.Faction, rulebook *rulebook.Rulebook, registry *hooks.Registry, log *slog.Logger) ([]domain.Mutation, error) {
	var mutations []domain.Mutation

	driftRatingOf := dispatch.AssetDriftRatingResolver(registry, faction)
	speedOf := dispatch.AssetSpeedResolver(registry, faction)
	// runsHot is the decision's choice, or forced when the asset moves only hot.
	runsHot := func(decision MovementDecision, def *domain.AssetDefinition) bool {
		return decision.RunHot || speedOf(def, false) == 0
	}
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
			driftRating := driftRatingOf(def)
			crossingCost := rulebook.DriftCost(driftRating)

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
				DriftRating:   driftRating,
				CargoAssetIDs: decision.CargoAssetIDs,
				RunHot:        runsHot(decision, def),
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
			driftRating := driftRatingOf(def)
			crossingCost := rulebook.DriftCost(driftRating)
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
				DriftRating:   driftRating,
				CargoAssetIDs: asset.CurrentOrder.CargoAssetIDs,
				RunHot:        runsHot(decision, def),
				// deliberate: banked points carry into the new order, not reset, because they are turns the asset already spent waiting (Effort 4 Decision Record 3).
				Banked: asset.CurrentOrder.Banked,
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
