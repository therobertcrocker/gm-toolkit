package domain

// Base represents a Base of Influence — a faction's foothold on a world.
// Every faction has exactly one Base on its homeworld (IsHomeworld=true) at
// the faction's current MaxHP. Non-homeworld Bases are placed by the Expand
// Influence action.
//
// Per SWN, damage to a Base is also dealt directly to faction HP; attack
// resolution emits a FactionHPDelta alongside every BaseHPDelta.
type Base struct {
	ID          string `toml:"id"`
	OwnerID     string `toml:"owner_id"`
	Location    string `toml:"location"`
	CurrentHP   int    `toml:"current_hp"`
	MaxHP       int    `toml:"max_hp"`
	Influence   int    `toml:"influence"`
	Ready       bool   `toml:"ready"`
	IsHomeworld bool   `toml:"is_homeworld"`
}

// EffectiveMaxHP returns the Base's usable max HP. Homeworld Bases track the
// owning faction's current MaxHP (growing with attribute raises); non-homeworld
// Bases use the stored MaxHP captured when HP was purchased.
func (base *Base) EffectiveMaxHP(faction *Faction) int {
	if base.IsHomeworld {
		return faction.MaxHP
	}
	return base.MaxHP
}
