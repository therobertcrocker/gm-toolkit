package eventhooks

// Collector abstracts the two input methods that hook dispatchers require from
// the GM. SelectModifiers is called with the full Cat 1 offer list; the GM
// returns the subset to apply. ConfirmReroll is called for each elective Cat 2
// directive; returning false skips that reroll.
type Collector interface {
	SelectModifiers(offers []ModifierOffer) []ModifierOffer
	ConfirmReroll(directive RerollDirective) bool
}
