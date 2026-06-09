package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewSelectModifiers builds the pre-roll modifier multi-select (cross-cutting:
// fires during any rolling action). Answer: []*hooks.ModifierOffer (the subset to
// apply); the adapter derefs back to the engine's []hooks.ModifierOffer. Pointer
// element type because ModifierOffer carries a func field and so is not comparable
// — huh.Option requires comparable. Esc -> the execution model receives
// ErrTurnCanceled, but SelectModifiers returns no error, so the adapter method
// maps cancel to "none applied" (no-op).
func NewSelectModifiers(offers []hooks.ModifierOffer) MultiSelectOverlay[*hooks.ModifierOffer] {
	opts := make([]huh.Option[*hooks.ModifierOffer], 0, len(offers))
	for i := range offers {
		offer := &offers[i]
		label := offer.Description
		if offer.Source != "" {
			label = fmt.Sprintf("%s (%s)", offer.Description, offer.Source)
		}
		opts = append(opts, huh.NewOption(label, offer))
	}
	return NewMultiSelect[*hooks.ModifierOffer]("Apply pre-roll modifiers", "", opts, 0)
}

// NewConfirmReroll builds the elective-reroll confirm. Answer: bool. Esc maps to
// false (skip the reroll) in the adapter method (ConfirmReroll returns no error).
func NewConfirmReroll(directive hooks.RerollDirective) ConfirmOverlay {
	header := ""
	if directive.Source != "" {
		header = fmt.Sprintf("Source: %s", directive.Source)
	}
	return NewConfirm(fmt.Sprintf("Reroll %d di(c)e?", len(directive.Indices)), header, "Reroll", "Keep")
}
