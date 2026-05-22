package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// TransportHandler binds one transport-capable asset definition to the
// EffectsEngine. At ApplyAll time, it registers a TransportReactor for each
// instance of that definition owned by a faction.
type TransportHandler struct {
	def *domain.AssetDefinition
}

func NewTransportHandler(def *domain.AssetDefinition) *TransportHandler {
	return &TransportHandler{def: def}
}

func (handler *TransportHandler) AssetDefinitionID() string { return handler.def.ID }

func (handler *TransportHandler) Apply(faction *domain.Faction, asset *domain.Asset, hookRegistry *hooks.Registry) {
	hookRegistry.RegisterMutationReactor(
		hooks.FactionScope(faction.ID),
		"transport:"+handler.def.ID,
		&TransportReactor{
			FactionID: faction.ID,
			AssetID:   asset.ID,
			Profile:   handler.def.Transport,
		},
	)
}

// TransportReactor reacts to MovementOrder mutations targeting its bound
// transport asset. It charges the per-issuance Coin cost and keeps every
// cargo asset's Location in sync with the transport so cargo state is
// recoverable if the transport is later destroyed mid-flight.
type TransportReactor struct {
	FactionID string
	AssetID   string
	Profile   *domain.TransportProfile
}

func (reactor *TransportReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	var extra []domain.Mutation
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.MovementOrderIssued:
			if m.AssetID != reactor.AssetID || len(m.Order.CargoAssetIDs) == 0 {
				continue
			}
			extra = append(extra, domain.CoinDelta{
				FactionID:         reactor.FactionID,
				Delta:             -reactor.Profile.CoinCost,
				Cause:             "transport_cost",
				CausedByFactionID: reactor.FactionID,
			})
			for _, cargoID := range m.Order.CargoAssetIDs {
				extra = append(extra, buildCargoMoved(
					factionState,
					reactor.FactionID,
					cargoID,
					domain.Location{RegionHex: m.Order.Path[0]},
					"transport_issued",
				))
			}

		case domain.MovementOrderProgressed:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				domain.Location{RegionHex: m.RegionHex},
				"transport_progressed",
			)...)

		case domain.MovementOrderCompleted:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				m.FinalLocation,
				"transport_completed",
			)...)

		case domain.MovementOrderCancelled:
			extra = append(extra, reactor.cargoFollows(
				factionState,
				m.AssetID,
				m.StrandedAt,
				"transport_cancelled",
			)...)

		case domain.MovementOrderRevised:
			// No-op. Cargo manifest is carried forward by BuildMovementMutations;
			// subsequent Progressed mutations drive cargo along the new path.
		}
	}
	return extra
}

// cargoFollows resolves the transport from factionState, reads its cargo
// manifest off CurrentOrder, and builds one AssetMoved per cargo to toLoc.
// Returns nil for asset-ID mismatch, missing faction/asset, or a transport
// with no in-flight order — all "not my problem" cases that the reactor
// silently skips.
func (reactor *TransportReactor) cargoFollows(factionState *state.FactionState, mutationAssetID string, toLoc domain.Location, cause string) []domain.Mutation {
	if mutationAssetID != reactor.AssetID {
		return nil
	}
	faction, ok := factionState.Factions[reactor.FactionID]
	if !ok {
		return nil
	}
	transport, ok := faction.Assets[reactor.AssetID]
	if !ok || transport.CurrentOrder == nil {
		return nil
	}
	var moves []domain.Mutation
	for _, cargoID := range transport.CurrentOrder.CargoAssetIDs {
		moves = append(moves, buildCargoMoved(factionState, reactor.FactionID, cargoID, toLoc, cause))
	}
	return moves
}

func buildCargoMoved(factionState *state.FactionState, factionID, cargoID string, toLoc domain.Location, cause string) domain.AssetMoved {
	var from domain.Location
	if faction, ok := factionState.Factions[factionID]; ok {
		if cargo, ok := faction.Assets[cargoID]; ok {
			from = cargo.Location
		}
	}
	return domain.AssetMoved{
		FactionID:         factionID,
		AssetID:           cargoID,
		FromLocation:      from,
		ToLocation:        toLoc,
		Cause:             cause,
		CausedByFactionID: factionID,
	}
}
