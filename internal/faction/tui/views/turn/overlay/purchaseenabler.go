package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// NewPurchaseEnabler prompts the GM to pick one asset from the enabler-eligible
// candidates. Labels show the surcharged cost (half again, rounded up) so the premium
// is visible at selection. Answer: *domain.AssetDefinition. Esc cancels.
func NewPurchaseEnabler(payload adapter.SelectPurchaseEnablerAssetPayload) SelectOverlay[*domain.AssetDefinition] {
	opts := make([]huh.Option[*domain.AssetDefinition], 0, len(payload.Candidates))
	for _, candidate := range payload.Candidates {
		opts = append(opts, huh.NewOption(
			fmt.Sprintf("%s · TL %d · cost %d Coin (surcharged)", candidate.Def.Name, candidate.Def.TechLevel, candidate.Cost),
			candidate.Def,
		))
	}
	return NewSelect[*domain.AssetDefinition]("Purchase which asset? (Pretech Logistics)", "", opts)
}
