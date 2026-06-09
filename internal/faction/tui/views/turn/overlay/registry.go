package overlay

// ImplementedActions is the set of action Names whose collector prompts the TUI
// can drive. The SelectAction overlay disables any available action not in this
// set. Seeded with the zero-prompt and shared-only actions (no per-action commit
// needed); each per-action commit adds one entry.
var ImplementedActions = map[string]bool{
	"Sell Asset":        true, // shared SelectAsset only (built in foundation)
	"Repair Faction":    true, // zero-prompt
	"Abandon Goal":      true, // zero-prompt
	"Buy Asset":         true,
	"Refit Asset":       true,
	"Repair Asset":      true,
	"Bribe":             true,
	"Seize Planet":      true,
	"Change Homeworld":  true,
	"Attack":            true,
	"Expand Influence":  true,
	"Use Asset Ability": true,
}
