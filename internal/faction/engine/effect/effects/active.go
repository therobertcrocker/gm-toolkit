package effects

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// active reports whether asset can use its special ability now: an asset is not
// usable until the turn after purchase, while unmaintained, or once destroyed.
// Reactors check at reaction time, not at registration, because Ready and
// Maintained both change mid-cycle.
func active(asset *domain.Asset) bool {
	return asset != nil && asset.Ready && asset.Maintained && asset.CurrentHP > 0
}
