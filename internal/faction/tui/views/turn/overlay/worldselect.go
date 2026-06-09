package overlay

import "github.com/charmbracelet/huh"

// NewWorldSelect builds a single-world SelectOverlay from parallel world IDs and
// an adapter-supplied worldID->name map. Label = name (falling back to the raw
// ID), value = the world ID. Shared by Seize Planet and Change Homeworld.
func NewWorldSelect(title, header string, worlds []string, names map[string]string) SelectOverlay[string] {
	opts := make([]huh.Option[string], len(worlds))
	for i, worldID := range worlds {
		label := names[worldID]
		if label == "" {
			label = worldID
		}
		opts[i] = huh.NewOption(label, worldID)
	}
	return NewSelect[string](title, header, opts)
}
