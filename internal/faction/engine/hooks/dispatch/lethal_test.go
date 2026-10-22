package dispatch

import (
	"errors"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// offerInterceptor is a fixture Cat 6 hook: it always makes the same offer.
type offerInterceptor struct{ offer *hooks.Interception }

func (interceptor *offerInterceptor) OfferInterception(_ hooks.LethalContext) *hooks.Interception {
	return interceptor.offer
}

// fakeInterceptCollector records each prompt and gives the same answer to all.
type fakeInterceptCollector struct {
	answer     bool
	err        error
	sacrifices []hooks.SacrificeOffer
	payments   []hooks.PaymentOffer
}

func (collector *fakeInterceptCollector) ConfirmSacrifice(offer hooks.SacrificeOffer) (bool, error) {
	collector.sacrifices = append(collector.sacrifices, offer)
	return collector.answer, collector.err
}

func (collector *fakeInterceptCollector) ConfirmPayment(offer hooks.PaymentOffer) (bool, error) {
	collector.payments = append(collector.payments, offer)
	return collector.answer, collector.err
}

// guardOffer is the offer one fixture interceptor makes, by asset ID.
type guardOffer struct {
	ask       hooks.InterceptAsk
	price     int
	budgetKey string
}

// takenBy is the marker mutation an interceptor on assetID offers.
func takenBy(assetID string) []domain.Mutation {
	return []domain.Mutation{domain.AssetHPDelta{FactionID: "f1", AssetID: "t1", Delta: 1, Cause: assetID}}
}

// lethalFixture is f1 (3 Coin) holding target t1 and guards g1 and g2, and f2
// holding striker s1. Each guard offer is registered on its asset.
func lethalFixture(guards map[string]guardOffer) (*hooks.Registry, hooks.LethalContext) {
	assets := map[string]*domain.Asset{
		"t1": {ID: "t1", OwnerID: "f1"},
		"g1": {ID: "g1", OwnerID: "f1"},
		"g2": {ID: "g2", OwnerID: "f1"},
		"s1": {ID: "s1", OwnerID: "f2"},
	}
	f1 := &domain.Faction{ID: "f1", Coin: 3, Assets: map[string]*domain.Asset{"t1": assets["t1"], "g1": assets["g1"], "g2": assets["g2"]}}
	f2 := &domain.Faction{ID: "f2", Assets: map[string]*domain.Asset{"s1": assets["s1"]}}

	registry := hooks.NewRegistry()
	for assetID, guard := range guards {
		asset := assets[assetID]
		registry.RegisterLethalInterceptor(hooks.AssetScope(asset.OwnerID, asset.ID), "fixture", &offerInterceptor{offer: &hooks.Interception{
			Source:    asset,
			Ask:       guard.ask,
			Price:     guard.price,
			BudgetKey: guard.budgetKey,
			Mutations: takenBy(assetID),
		}})
	}
	return registry, hooks.LethalContext{
		Target:       assets["t1"],
		Striker:      assets["s1"],
		FactionState: &state.FactionState{Factions: map[string]*domain.Faction{"f1": f1, "f2": f2}},
	}
}

func TestInterceptLethal(t *testing.T) {
	promptErr := errors.New("prompt failed")
	testCases := []struct {
		name           string
		nilRegistry    bool
		guards         map[string]guardOffer
		collector      *fakeInterceptCollector // nil: InterceptLethalAutomatic
		batch          []domain.Mutation
		want           []domain.Mutation
		wantErr        error
		wantSacrifices []hooks.SacrificeOffer
		wantPayments   []hooks.PaymentOffer
		wantSpent      bool
		again          bool // a second call is taken by nothing and asks nothing
	}{
		{
			name:        "nil registry",
			nilRegistry: true,
			collector:   &fakeInterceptCollector{answer: true},
		},
		{
			name:      "no interceptor",
			collector: &fakeInterceptCollector{answer: true},
		},
		{
			name: "automatic before elective, although the elective has the lower ID",
			guards: map[string]guardOffer{
				"g1": {ask: hooks.InterceptSacrifice},
				"g2": {ask: hooks.InterceptAutomatic},
				"s1": {ask: hooks.InterceptAutomatic},
			},
			collector: &fakeInterceptCollector{answer: true},
			want:      takenBy("g2"),
		},
		{
			name: "target side's elective before the striker's automatic",
			guards: map[string]guardOffer{
				"g1": {ask: hooks.InterceptSacrifice},
				"s1": {ask: hooks.InterceptAutomatic},
			},
			collector:      &fakeInterceptCollector{answer: true},
			want:           takenBy("g1"),
			wantSacrifices: []hooks.SacrificeOffer{{Sacrifice: &domain.Asset{ID: "g1", OwnerID: "f1"}, Saves: &domain.Asset{ID: "t1", OwnerID: "f1"}}},
		},
		{
			name: "declined sacrifice moves to the striker",
			guards: map[string]guardOffer{
				"g1": {ask: hooks.InterceptSacrifice},
				"s1": {ask: hooks.InterceptAutomatic},
			},
			collector:      &fakeInterceptCollector{answer: false},
			want:           takenBy("s1"),
			wantSacrifices: []hooks.SacrificeOffer{{Sacrifice: &domain.Asset{ID: "g1", OwnerID: "f1"}, Saves: &domain.Asset{ID: "t1", OwnerID: "f1"}}},
		},
		{
			name: "no collector skips the elective offer",
			guards: map[string]guardOffer{
				"g1": {ask: hooks.InterceptSacrifice},
				"s1": {ask: hooks.InterceptAutomatic},
			},
			want: takenBy("s1"),
		},
		{
			name:         "paid restore spends its budget once",
			guards:       map[string]guardOffer{"g1": {ask: hooks.InterceptRestore, price: 2, budgetKey: "k"}},
			collector:    &fakeInterceptCollector{answer: true},
			want:         takenBy("g1"),
			wantPayments: []hooks.PaymentOffer{{Asset: &domain.Asset{ID: "t1", OwnerID: "f1"}, Amount: 2, Reason: hooks.PaymentRestore}},
			wantSpent:    true,
			again:        true,
		},
		{
			name:         "declined restore spends nothing",
			guards:       map[string]guardOffer{"g1": {ask: hooks.InterceptRestore, price: 2, budgetKey: "k"}},
			collector:    &fakeInterceptCollector{answer: false},
			wantPayments: []hooks.PaymentOffer{{Asset: &domain.Asset{ID: "t1", OwnerID: "f1"}, Amount: 2, Reason: hooks.PaymentRestore}},
		},
		{
			name:      "restore the batch's Coin cannot cover is not offered",
			guards:    map[string]guardOffer{"g1": {ask: hooks.InterceptRestore, price: 2, budgetKey: "k"}},
			collector: &fakeInterceptCollector{answer: true},
			batch:     []domain.Mutation{domain.CoinDelta{FactionID: "f1", Delta: -2}},
		},
		{
			name:      "interceptor the batch removes offers nothing",
			guards:    map[string]guardOffer{"g2": {ask: hooks.InterceptAutomatic}},
			collector: &fakeInterceptCollector{answer: true},
			batch:     []domain.Mutation{domain.AssetRemoved{FactionID: "f1", AssetID: "g2"}},
		},
		{
			name:      "interceptor the batch moves offers nothing",
			guards:    map[string]guardOffer{"g2": {ask: hooks.InterceptAutomatic}},
			collector: &fakeInterceptCollector{answer: true},
			batch:     []domain.Mutation{domain.AssetOwnerChanged{FactionID: "f1", AssetID: "g2", ToFactionID: "f2"}},
		},
		{
			name:           "prompt error aborts the walk",
			guards:         map[string]guardOffer{"g1": {ask: hooks.InterceptSacrifice}, "s1": {ask: hooks.InterceptAutomatic}},
			collector:      &fakeInterceptCollector{err: promptErr},
			wantErr:        promptErr,
			wantSacrifices: []hooks.SacrificeOffer{{Sacrifice: &domain.Asset{ID: "g1", OwnerID: "f1"}, Saves: &domain.Asset{ID: "t1", OwnerID: "f1"}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry, ctx := lethalFixture(testCase.guards)
			if testCase.nilRegistry {
				registry = nil
			}
			ctx.Batch = testCase.batch
			owner := ctx.FactionState.Factions["f1"]

			intercept := func() ([]domain.Mutation, error) {
				if testCase.collector == nil {
					return InterceptLethalAutomatic(registry, ctx), nil
				}
				return InterceptLethal(registry, ctx, testCase.collector)
			}

			got, err := intercept()
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("mutations = %v, want %v", got, testCase.want)
			}
			if owner.BudgetSpent("k") != testCase.wantSpent {
				t.Errorf("budget k spent = %v, want %v", owner.BudgetSpent("k"), testCase.wantSpent)
			}
			if testCase.collector == nil {
				return
			}
			checkPrompts(t, testCase.collector, owner, testCase.wantSacrifices, testCase.wantPayments)

			if testCase.again {
				prompts := len(testCase.collector.sacrifices) + len(testCase.collector.payments)
				got, err := intercept()
				if err != nil || got != nil {
					t.Errorf("second call = %v, %v, want nil, nil", got, err)
				}
				if len(testCase.collector.sacrifices)+len(testCase.collector.payments) != prompts {
					t.Error("second call prompted, want no prompt once the budget is spent")
				}
			}
		})
	}
}

// checkPrompts compares the recorded prompts with want. The owner and payer are
// always the guards' faction, so want leaves them out.
func checkPrompts(t *testing.T, collector *fakeInterceptCollector, owner *domain.Faction, wantSacrifices []hooks.SacrificeOffer, wantPayments []hooks.PaymentOffer) {
	t.Helper()
	for index := range wantSacrifices {
		wantSacrifices[index].Owner = owner
	}
	for index := range wantPayments {
		wantPayments[index].Payer = owner
	}
	if !reflect.DeepEqual(collector.sacrifices, wantSacrifices) {
		t.Errorf("sacrifice prompts = %+v, want %+v", collector.sacrifices, wantSacrifices)
	}
	if !reflect.DeepEqual(collector.payments, wantPayments) {
		t.Errorf("payment prompts = %+v, want %+v", collector.payments, wantPayments)
	}
}
