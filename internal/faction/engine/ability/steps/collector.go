package steps

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Collector is the subset of input collection the step handlers need.
// Structurally equivalent to ability.Collector; defined here to avoid an import cycle.
type Collector interface {
	SelectMoveDestination(asset *domain.Asset, worlds []string) (string, error)
	SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
}
