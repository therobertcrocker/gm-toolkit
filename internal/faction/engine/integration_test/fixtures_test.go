package integration_test

import (
	"fmt"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

// addBase appends a non-homeworld Base of Influence to a faction.
func addBase(faction *domain.Faction, world string, hp int) *domain.Base {
	base := &domain.Base{
		ID:          fmt.Sprintf("%s-base-%s-1", faction.ID, world),
		OwnerID:     faction.ID,
		Location:    world,
		CurrentHP:   hp,
		MaxHP:       hp,
		Ready:       true,
		IsHomeworld: false,
	}
	faction.Bases = append(faction.Bases, base)
	return base
}

// addAssetOnWorld appends a Security Personnel asset to a faction on a specific world.
func addAssetOnWorld(faction *domain.Faction, world string) *domain.Asset {
	asset := &domain.Asset{
		ID:           fmt.Sprintf("%s-%s-extra", faction.ID, world),
		DefinitionID: defSecurityPersonnel,
		OwnerID:      faction.ID,
		Location:     world,
		CurrentHP:    3,
		Ready:        true,
		Maintained:   true,
	}
	faction.Assets = append(faction.Assets, asset)
	return asset
}

// checkStep logs a labeled assertion result.
// With -v: ✓ lines appear on pass; ✗ lines surface as test errors with detail.
func checkStep(t *testing.T, description string, ok bool, detail string) {
	t.Helper()
	if ok {
		t.Logf("  ✓ %s", description)
	} else {
		t.Errorf("  ✗ %s: %s", description, detail)
	}
}
