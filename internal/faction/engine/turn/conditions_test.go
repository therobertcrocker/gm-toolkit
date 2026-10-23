package turn

import (
	"errors"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// paymentFake answers every offer with answer, or fails with err, and keeps the
// offers it saw.
type paymentFake struct {
	answer bool
	err    error
	offers []hooks.PaymentOffer
}

func (fake *paymentFake) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	fake.offers = append(fake.offers, offer)
	return fake.answer, fake.err
}

func TestExpireConditions(t *testing.T) {
	endsAtF1 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src", ExpiresAtTurnOf: "f1"}
	endsAtF3 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src2", ExpiresAtTurnOf: "f3"}
	endsElsewhere := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "src3"}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {ID: "f1", Assets: map[string]*domain.Asset{
				"f1-a": {ID: "f1-a", Conditions: []domain.Condition{endsElsewhere}},
			}},
			"f2": {ID: "f2", Assets: map[string]*domain.Asset{
				"f2-a": {ID: "f2-a", Conditions: []domain.Condition{endsAtF1, endsAtF3}},
			}},
			"f3": {ID: "f3", Assets: map[string]*domain.Asset{
				"f3-a": {ID: "f3-a", Conditions: []domain.Condition{endsAtF1}},
			}},
		},
	}

	mutations := ExpireConditions(factionState, "f1")

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2: %+v", len(mutations), mutations)
	}
	wantHolders := []string{"f2", "f3"}
	for i, mutation := range mutations {
		cleared, ok := mutation.(domain.AssetConditionCleared)
		if !ok {
			t.Fatalf("mutations[%d] is %T, want AssetConditionCleared", i, mutation)
		}
		if cleared.FactionID != wantHolders[i] || cleared.Condition != endsAtF1 {
			t.Errorf("mutations[%d] = %+v, want a clear of %+v on %s", i, cleared, endsAtF1, wantHolders[i])
		}
		if cleared.Cause != "condition_expired" || cleared.CausedByFactionID != "f1" {
			t.Errorf("mutations[%d] cause = %q by %q, want condition_expired by f1", i, cleared.Cause, cleared.CausedByFactionID)
		}
	}
}

func TestOfferConditionLifts(t *testing.T) {
	disabled := domain.Condition{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}
	makeFaction := func() *domain.Faction {
		return &domain.Faction{ID: "f1", Coin: 6, Assets: map[string]*domain.Asset{
			"x1": {ID: "x1", DefinitionID: "cost-10", Conditions: []domain.Condition{disabled}},
			"x2": {ID: "x2", DefinitionID: "cost-10", Conditions: []domain.Condition{disabled}},
			"x3": {ID: "x3", DefinitionID: "cost-10", Conditions: []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "s1"}}},
		}}
	}
	rulebook := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{"cost-10": {ID: "cost-10", Cost: 10}}}

	t.Run("pays the first lift; the second is past what is left", func(t *testing.T) {
		collector := &paymentFake{answer: true}
		mutations, err := OfferConditionLifts(makeFaction(), rulebook, collector)
		if err != nil {
			t.Fatalf("OfferConditionLifts: %v", err)
		}
		want := []domain.Mutation{
			domain.CoinDelta{FactionID: "f1", Delta: -5, Cause: "condition_lifted", CausedByFactionID: "f1"},
			domain.AssetConditionCleared{FactionID: "f1", AssetID: "x1", Condition: disabled, Cause: "condition_lifted", CausedByFactionID: "f1"},
		}
		if !reflect.DeepEqual(mutations, want) {
			t.Errorf("mutations = %+v, want %+v", mutations, want)
		}
		if len(collector.offers) != 1 || collector.offers[0].Asset.ID != "x1" || collector.offers[0].Amount != 5 || collector.offers[0].Reason != hooks.PaymentLiftDisable {
			t.Errorf("offers = %+v, want one lift_disable offer of 5 for x1", collector.offers)
		}
	})

	t.Run("declined offers emit nothing", func(t *testing.T) {
		collector := &paymentFake{answer: false}
		mutations, err := OfferConditionLifts(makeFaction(), rulebook, collector)
		if err != nil || mutations != nil {
			t.Errorf("mutations, err = %+v, %v; want nil, nil", mutations, err)
		}
		if len(collector.offers) != 2 {
			t.Errorf("offers = %d, want 2 (x1 and x2)", len(collector.offers))
		}
	})

	t.Run("a collector error is returned", func(t *testing.T) {
		promptErr := errors.New("prompt failed")
		_, err := OfferConditionLifts(makeFaction(), rulebook, &paymentFake{err: promptErr})
		if !errors.Is(err, promptErr) {
			t.Errorf("err = %v, want it to wrap %v", err, promptErr)
		}
	})
}
