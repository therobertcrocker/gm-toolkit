package actions

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Bribe struct {
	collector  engine.InputCollector
	factionID  string
	baseID     string
	coinAmount int
}

func NewBribe(collector engine.InputCollector) *Bribe {
	return &Bribe{collector: collector}
}

func (b *Bribe) Name() string { return "Bribe" }

func (b *Bribe) Validate(faction *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) bool {
	return len(faction.Bases) > 0 && faction.Coin >= 1
}

func (b *Bribe) Inputs(faction *domain.Faction, factionState *state.FactionState, _ *loader.Rulebook) error {
	base, amount, err := b.collector.SelectBribeTarget(faction, factionState)
	if err != nil {
		return fmt.Errorf("bribe: %w", err)
	}
	b.factionID = faction.ID
	b.baseID = base.ID
	b.coinAmount = amount
	return nil
}

func (b *Bribe) Resolve(_ *domain.Faction, _ *state.FactionState, _ *loader.Rulebook) error {
	return nil
}

func (b *Bribe) Output() ([]domain.Mutation, error) {
	return []domain.Mutation{
		domain.CoinDelta{FactionID: b.factionID, Delta: -b.coinAmount, Cause: "bribe", CausedByFactionID: b.factionID},
		domain.InfluenceDelta{FactionID: b.factionID, BaseID: b.baseID, Delta: b.coinAmount, Cause: "bribe", CausedByFactionID: b.factionID},
	}, nil
}
