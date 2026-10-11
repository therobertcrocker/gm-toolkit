package effects_test

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// activeAsset builds an alpha-owned asset that may use its special ability.
func activeAsset(id, defID string) *domain.Asset {
	return &domain.Asset{ID: id, DefinitionID: defID, OwnerID: "alpha", CurrentHP: 4, Ready: true, Maintained: true}
}

// inactiveStates are the three ways an asset loses the use of its special ability.
var inactiveStates = map[string]func(*domain.Asset){
	"not yet ready": func(asset *domain.Asset) { asset.Ready = false },
	"unmaintained":  func(asset *domain.Asset) { asset.Maintained = false },
	"destroyed":     func(asset *domain.Asset) { asset.CurrentHP = 0 },
}
