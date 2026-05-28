package tui

import "github.com/charmbracelet/bubbles/help"

// Helper is implemented by sub-models that contribute keybindings to the
// help overlay. Root combines the active sub's Help() with global bindings.
type Helper interface {
	Help() help.KeyMap
}
