package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// Collector abstracts the two user-input methods the ability engine requires.
type Collector interface {
	SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
	SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
}
