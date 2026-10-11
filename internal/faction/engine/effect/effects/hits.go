package effects

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// hitTargets returns the faction struck by each hit assetID dealt in mutations,
// in mutation order. counter picks counterattack hits over attack hits. A hit the
// defender redirected to its Base is still an attack hit; a counterattack never
// strikes a Base.
func hitTargets(mutations []domain.Mutation, assetID string, counter bool) []string {
	var targets []string
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetHPDelta:
			if m.Cause == "attack" && m.CausedByAssetID == assetID && m.Counter == counter {
				targets = append(targets, m.FactionID)
			}
		case domain.BaseHPDelta:
			if !counter && m.Cause == "attack" && m.CausedByAssetID == assetID {
				targets = append(targets, m.FactionID)
			}
		}
	}
	return targets
}
