package eventhooks

// ScopeKind identifies what a Scope targets.
type ScopeKind int

const (
	ScopeFaction ScopeKind = iota
	ScopeAsset
	ScopeGlobal
)

// Scope identifies the registration target for a hook. Comparable, so safe as
// a map key. Global-scoped hooks fire for every faction/asset query.
type Scope struct {
	Kind            ScopeKind
	FactionID       string
	AssetInstanceID string
}

// FactionScope returns a Scope targeting a specific faction.
func FactionScope(factionID string) Scope {
	return Scope{Kind: ScopeFaction, FactionID: factionID}
}

// AssetScope returns a Scope targeting a specific asset instance within a faction.
func AssetScope(factionID, assetInstanceID string) Scope {
	return Scope{Kind: ScopeAsset, FactionID: factionID, AssetInstanceID: assetInstanceID}
}

// GlobalScope returns a Scope that fires for any faction or asset query.
func GlobalScope() Scope {
	return Scope{Kind: ScopeGlobal}
}
