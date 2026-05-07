package tags

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const ScavengersTagID = "T-014"

// ScavengersReactor implements hooks.MutationReactor for the Scavengers tag.
// Grants +1 Coin to the owning faction per asset destroyed in combat, own or rival.
type ScavengersReactor struct {
	FactionID string
	credited  map[string]bool // guards against re-crediting the same asset across recursion rounds
}

func (reactor *ScavengersReactor) OnMutations(mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if reactor.credited == nil {
		reactor.credited = make(map[string]bool)
	}
	var extra []domain.Mutation
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if ok && removed.Cause == "attack" && !reactor.credited[removed.AssetID] {
			reactor.credited[removed.AssetID] = true
			extra = append(extra, domain.CoinDelta{
				FactionID:         reactor.FactionID,
				Delta:             1,
				Cause:             "scavengers",
				CausedByFactionID: reactor.FactionID,
			})
		}
	}
	return extra
}

func RegisterScavengers(eng *engine.Engine, faction *domain.Faction) {
	eng.Hooks.RegisterMutationReactor(
		hooks.FactionScope(faction.ID),
		"scavengers",
		&ScavengersReactor{FactionID: faction.ID},
	)
}
