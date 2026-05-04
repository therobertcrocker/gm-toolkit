package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type AbandonGoal struct {
	factionID string
	goalID    string
	income    int
}

func NewAbandonGoal() *AbandonGoal { return &AbandonGoal{} }

func (a *AbandonGoal) Name() string { return "Abandon Goal" }

func (a *AbandonGoal) Validate(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) bool {
	return faction.ActiveGoal != nil
}

func (a *AbandonGoal) Inputs(_ *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	return nil
}

func (a *AbandonGoal) Resolve(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	a.factionID = faction.ID
	a.goalID = faction.ActiveGoal.GoalID
	a.income = faction.Wealth/2 + (faction.Force+faction.Cunning)/4
	return nil
}

func (a *AbandonGoal) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.CoinDelta{
			FactionID:         a.factionID,
			Delta:             -a.income,
			Cause:             "abandon_goal",
			CausedByFactionID: a.factionID,
		},
		domain.GoalAbandoned{
			FactionID:         a.factionID,
			GoalID:            a.goalID,
			Cause:             "player_choice",
			CausedByFactionID: a.factionID,
		},
	}, nil
}
