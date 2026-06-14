package tags

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

const ScavengersTagID = "T-016"

// ScavengersReactor implements hooks.MutationReactor for the Scavengers tag.
// Grants +1 Coin to the owning faction per asset destroyed in combat, own or rival.
type ScavengersReactor struct {
	FactionID string
}

func (reactor *ScavengersReactor) OnMutations(mutations []domain.Mutation, _ *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	var extra []domain.Mutation
	for _, mutation := range mutations {
		removed, ok := mutation.(domain.AssetRemoved)
		if ok && removed.Cause == "attack" {
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

type ScavengersHandler struct{}

func (ScavengersHandler) TagID() string { return ScavengersTagID }

func (ScavengersHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
	hookRegistry.RegisterMutationReactor(
		hooks.FactionScope(faction.ID),
		"scavengers",
		&ScavengersReactor{FactionID: faction.ID},
	)
}
