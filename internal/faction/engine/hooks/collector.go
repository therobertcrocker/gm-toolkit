package hooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Collector abstracts the two asks of a hooked roll. SelectModifiers is called before
// the dice on every roll, offers or not, and returns the offers to apply. SelectReroll
// is called once the roll settles, rerolls on offer or not, and again after each
// reroll it takes; it returns the reroll to take, or nil to keep the dice.
type Collector interface {
	SelectModifiers(card PreRollCard) []ModifierOffer
	SelectReroll(card PostRollCard) *RerollDirective
}

// PaymentReason says what a payment buys, so a prompt can word it.
type PaymentReason string

const (
	// PaymentAvoidDisable: pay now, or the asset is disabled (Marketers).
	PaymentAvoidDisable PaymentReason = "avoid_disable"
	// PaymentLiftDisable: pay to lift a disabled condition at the owner's turn start.
	PaymentLiftDisable PaymentReason = "lift_disable"
	// PaymentRestore: pay to restore an asset at 1 HP instead of losing it (Medical Center).
	PaymentRestore PaymentReason = "restore"
)

// PaymentOffer asks Payer whether to pay Amount Coin for Reason on Asset.
type PaymentOffer struct {
	Payer  *domain.Faction
	Asset  *domain.Asset
	Amount int
	Reason PaymentReason
}

// PaymentCollector asks one faction whether to pay one amount. action.Collector and
// the engine's PhaseCollector both embed it, so an action, the bookkeeping step and
// a hook dispatcher all ask the same question (A.001.6 overview Decision 19).
type PaymentCollector interface {
	ConfirmPayment(offer PaymentOffer) (bool, error)
}

// SacrificeOffer asks Owner whether to give up Sacrifice so that Saves survives a
// killing blow (False Front).
type SacrificeOffer struct {
	Owner     *domain.Faction
	Sacrifice *domain.Asset
	Saves     *domain.Asset
}

// InterceptCollector asks the owner of an elective Cat 6 interceptor: give up one
// asset, or pay to keep another. action.Collector embeds it; no phase destroys an
// asset by damage, so PhaseCollector does not.
type InterceptCollector interface {
	PaymentCollector
	ConfirmSacrifice(offer SacrificeOffer) (bool, error)
}

// PermissionRequest asks a world's government whether Requester may raise or land
// Def there. Ruler is nil for a world no faction governs; the GM answers for it
// (D-4). Cause names the path, so a prompt can say "raise" or "land"
// (domain.IsLanding).
type PermissionRequest struct {
	Requester *domain.Faction
	Ruler     *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
	Cause     string
}

// PermissionContestOffer asks Owner whether its Source (Lobbyists) contests the
// permission Requester was just granted to raise or land Def on WorldID. Cause is
// the request's path, as on PermissionRequest.
type PermissionContestOffer struct {
	Owner     *domain.Faction
	Source    *domain.Asset
	Requester *domain.Faction
	WorldID   string
	Def       *domain.AssetDefinition
	Cause     string
}

// PermissionCollector asks for a government's permission, and asks a Lobbyists'
// owner whether to contest a grant. action.Collector and the engine's
// PhaseCollector both embed it (D-41): purchases and relocates ask in an action,
// drift arrivals in the movement phase.
type PermissionCollector interface {
	RequestPermission(request PermissionRequest) (bool, error)
	ConfirmPermissionContest(offer PermissionContestOffer) (bool, error)
}
