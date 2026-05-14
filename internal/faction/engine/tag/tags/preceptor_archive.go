package tags

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

const PreceptorArchiveTagID = "T-013"

// PreceptorArchiveCostModifier implements hooks.AssetCostModifier for the
// Preceptor Archive tag. Reduces the Coin cost of TL4+ asset purchases by 1.
type PreceptorArchiveCostModifier struct{}

func (modifier *PreceptorArchiveCostModifier) ModifyAssetCost(_ *domain.Faction, def *domain.AssetDefinition, _ string, baseCost int) int {
	var newCost int
	if def.TechLevel >= 4 {
		newCost = baseCost - 1
	} else {
		newCost = baseCost
	}

	return max(newCost, 0)

}

type PreceptorArchiveHandler struct{}

func (PreceptorArchiveHandler) TagID() string { return PreceptorArchiveTagID }

func (PreceptorArchiveHandler) Apply(faction *domain.Faction, hookRegistry *hooks.Registry) {
	hookRegistry.RegisterAssetCostModifier(
		hooks.FactionScope(faction.ID),
		"preceptor-archive",
		&PreceptorArchiveCostModifier{},
	)
}
