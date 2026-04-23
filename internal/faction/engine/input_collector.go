package engine

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/loader"
)

// InputCollector abstracts input collection for action resolution. The GM
// implementation uses interactive prompts; an AI implementation uses
// goal-driven selection logic. Actions call only the methods they need.
type InputCollector interface {
	SelectAsset(assets []*domain.Asset, rulebook *loader.Rulebook) (*domain.Asset, error)
}
