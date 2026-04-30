package actions

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// RepairFaction heals the faction itself. No Coin cost; heal amount is
// round((highest stat + lowest stat) / 2), capped at missing HP.
type RepairFaction struct {
	factionID  string
	healAmount int
}

func NewRepairFaction() *RepairFaction {
	return &RepairFaction{}
}

func (rf *RepairFaction) Name() string { return "Repair Faction" }

func (rf *RepairFaction) Validate(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) bool {
	return faction.CurrentHP < faction.MaxHP
}

func (rf *RepairFaction) Inputs(_ *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	return nil
}

func (rf *RepairFaction) Resolve(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	stats := []int{faction.Force, faction.Cunning, faction.Wealth}
	highest, lowest := stats[0], stats[0]
	for _, stat := range stats[1:] {
		if stat > highest {
			highest = stat
		}
		if stat < lowest {
			lowest = stat
		}
	}
	rawHeal := (highest + lowest + 1) / 2
	missing := faction.MaxHP - faction.CurrentHP
	if rawHeal > missing {
		rawHeal = missing
	}
	rf.factionID = faction.ID
	rf.healAmount = rawHeal
	return nil
}

func (rf *RepairFaction) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.FactionHPDelta{FactionID: rf.factionID, Delta: rf.healAmount, Cause: "repair", CausedByFactionID: rf.factionID},
	}, nil
}
