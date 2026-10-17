package effects

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

type hit struct {
	strikerID       string
	targetFactionID string
}

// hits returns every hit in mutations, in mutation order. counter picks
// counterattack hits over attack hits. A hit the defender redirected to its Base
// is still an attack hit, and so is a rival's free attack on a new Base during
// Expand Influence; a counterattack never strikes a Base.
func hits(mutations []domain.Mutation, counter bool) []hit {
	var found []hit
	for _, mutation := range mutations {
		switch m := mutation.(type) {
		case domain.AssetHPDelta:
			if m.Cause == "attack" && m.CausedByAssetID != "" && m.Counter == counter {
				found = append(found, hit{strikerID: m.CausedByAssetID, targetFactionID: m.FactionID})
			}
		case domain.BaseHPDelta:
			if !counter && (m.Cause == "attack" || m.Cause == "expand") && m.CausedByAssetID != "" {
				found = append(found, hit{strikerID: m.CausedByAssetID, targetFactionID: m.FactionID})
			}
		}
	}
	return found
}

// hitTargets returns the faction struck by each hit assetID dealt in mutations,
// in mutation order.
func hitTargets(mutations []domain.Mutation, assetID string, counter bool) []string {
	var targets []string
	for _, found := range hits(mutations, counter) {
		if found.strikerID == assetID {
			targets = append(targets, found.targetFactionID)
		}
	}
	return targets
}
