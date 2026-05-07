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
}

func (reactor *ScavengersReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	var extra []domain.Mutation
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if ok && removed.Cause == "attack" {
			// Grant +1 Coin to the owning faction for each asset removed in combat.
			extra = append(extra,
				domain.CoinDelta{
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
