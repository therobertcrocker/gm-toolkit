package overlay

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewRequestPermission asks a world's government whether a faction may raise an
// asset there. Its header names who decides: the ruler, or the GM for a world no
// faction governs. Esc does nothing: a cancel would drop the action, and with it
// every earlier asset of a Use Asset Ability. Answer: bool.
func NewRequestPermission(request hooks.PermissionRequest, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of the GM (no faction governs %s)", worldName)
	if request.Ruler != nil {
		header = fmt.Sprintf("Asked of %s, the government of %s", request.Ruler.Name, worldName)
	}
	title := fmt.Sprintf("Permit %s to raise %s on %s?", request.Requester.Name, request.Def.Name, worldName)
	return NoCancelConfirm{NewConfirm(title, header, "Grant", "Refuse")}
}
