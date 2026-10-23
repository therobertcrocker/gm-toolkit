package overlay

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewRequestPermission asks a world's government whether a faction may raise an
// asset there, or land one. Its header names who decides: the ruler, or the GM for
// a world no faction governs. Esc does nothing: a cancel would drop the action, and
// with it every earlier asset of a Use Asset Ability. Answer: bool.
func NewRequestPermission(request hooks.PermissionRequest, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of the GM (no faction governs %s)", worldName)
	if request.Ruler != nil {
		header = fmt.Sprintf("Asked of %s, the government of %s", request.Ruler.Name, worldName)
	}
	return NoCancelConfirm{NewConfirm(permissionTitle(request, worldName), header, "Grant", "Refuse")}
}

func permissionTitle(request hooks.PermissionRequest, worldName string) string {
	return fmt.Sprintf("Permit %s to %s %s on %s?", request.Requester.Name, permissionVerb(request.Cause), request.Def.Name, worldName)
}

// permissionVerb words a permission cause: "raise" for a purchase, "land" for an
// arrival or a relocate.
func permissionVerb(cause string) string {
	if domain.IsLanding(cause) {
		return "land"
	}
	return "raise"
}

// NewConfirmPermissionContest asks a Lobbyists' owner whether to contest a
// permission a rival was just granted. Esc does nothing, as with the request it
// follows. Answer: bool.
func NewConfirmPermissionContest(offer hooks.PermissionContestOffer, worldName string) NoCancelConfirm {
	header := fmt.Sprintf("Asked of %s, for its asset %s", offer.Owner.Name, offer.Source.ID)
	return NoCancelConfirm{NewConfirm(contestTitle(offer, worldName), header, "Contest", "Let it stand")}
}

func contestTitle(offer hooks.PermissionContestOffer, worldName string) string {
	return fmt.Sprintf("Contest %s's permission to %s %s on %s?", offer.Requester.Name, permissionVerb(offer.Cause), offer.Def.Name, worldName)
}
