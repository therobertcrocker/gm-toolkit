package paths

import "path/filepath"

// Paths holds the resolved file paths for a campaign's persisted data.
type Paths struct {
	State      string
	History    string
	Narratives string
}

// New returns the canonical Paths for a given campaign ID.
// All paths are relative to the current working directory.
func New(campaignID string) Paths {
	base := filepath.Join(".", "campaigns", campaignID)
	return Paths{
		State:      filepath.Join(base, "faction_state.toml"),
		History:    filepath.Join(base, "history.jsonl"),
		Narratives: filepath.Join(base, "narratives"),
	}
}
