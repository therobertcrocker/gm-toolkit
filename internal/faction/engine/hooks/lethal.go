package hooks

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// LethalContext is one hit that would destroy Target, before its AssetRemoved is
// emitted (Cat 6). HPAfter is Target's HP once the hit lands, at most 0, and Damage
// is the hit's damage. Striker is the asset that dealt it, nil for self-damage.
// Batch is the mutations emitted so far in the action or the reactor round, the hit
// included; state has not applied them yet.
type LethalContext struct {
	Target       *domain.Asset
	TargetDef    *domain.AssetDefinition
	HPAfter      int
	Damage       int
	Striker      *domain.Asset
	FactionState *state.FactionState
	Batch        []domain.Mutation
}

// InterceptAsk says what an interception asks its owner before it takes the hit.
type InterceptAsk int

const (
	// InterceptAutomatic takes the hit with no prompt (Boltholes, Hostile Takeover).
	InterceptAutomatic InterceptAsk = iota
	// InterceptSacrifice asks the owner to give up Source (False Front).
	InterceptSacrifice
	// InterceptRestore asks the owner to pay Price (Medical Center).
	InterceptRestore
)

// Interception is one interceptor's offer to take a lethal hit. Mutations replace
// the target's AssetRemoved and carry the whole outcome, the HP it sets included.
// BudgetKey is spent on Source's owner when the offer is taken; empty is unlimited.
type Interception struct {
	Source    *domain.Asset
	Ask       InterceptAsk
	Price     int
	BudgetKey string
	Mutations []domain.Mutation
}

// LethalInterceptor offers to take a hit that would destroy an asset (Cat 6), or
// returns nil when it does not apply. It never prompts: the dispatcher orders the
// offers, asks the owner of an elective one, and spends the budget.
type LethalInterceptor interface {
	OfferInterception(ctx LethalContext) *Interception
}
