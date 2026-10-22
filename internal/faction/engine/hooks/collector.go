package hooks

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

// Collector abstracts the two input methods that hook dispatchers require from
// the GM. SelectModifiers is called with the full Cat 1 offer list; the GM
// returns the subset to apply. ConfirmReroll is called for each elective Cat 2
// directive; returning false skips that reroll.
type Collector interface {
	SelectModifiers(offers []ModifierOffer) []ModifierOffer
	ConfirmReroll(directive RerollDirective) bool
}

// PaymentReason says what a payment buys, so a prompt can word it.
type PaymentReason string

const (
	// PaymentAvoidDisable: pay now, or the asset is disabled (Marketers).
	PaymentAvoidDisable PaymentReason = "avoid_disable"
	// PaymentLiftDisable: pay to lift a disabled condition at the owner's turn start.
	PaymentLiftDisable PaymentReason = "lift_disable"
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
