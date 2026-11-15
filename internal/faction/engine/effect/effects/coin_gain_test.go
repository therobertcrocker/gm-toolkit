package effects_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func TestCoinGain_PaysTheOwnerTheAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount domain.DiceRoll
		want   int
	}{
		{"flat amount", domain.DiceRoll{Modifier: 1}, 1},
		{"rolled amount", domain.DiceRoll{NumDice: 1, Sides: 4}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faction := &domain.Faction{ID: "alpha"}
			entry := domain.EffectDefinition{
				Trigger:  domain.TriggerTurnStart,
				Effect:   domain.SpecialCoinGain,
				Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
				CoinGain: &domain.CoinGainParams{Amount: tc.amount},
			}
			got := effects.Registrars()[domain.SpecialCoinGain].Fire(effect.FireContext{
				Holder: effect.Holder{Faction: faction},
				Entry:  entry,
				Roller: dicetest.Constant{Value: 3},
			})
			want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: tc.want, Cause: "coin_gain", CausedByFactionID: "alpha"}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Fire = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCoinGain_ScavengersPaysOwnKillsAndLossesOnly(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-016"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	holder := effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-016", Name: "Scavengers"}}
	for _, filters := range []domain.EffectFilters{
		{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		{Owner: domain.OwnerRival, Scope: domain.ScopeFaction, CausedBy: domain.OwnerOwn},
	} {
		effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
			Holder: holder, Registry: registry, Roller: dicetest.Constant{Value: 1},
			Entry: domain.EffectDefinition{Trigger: domain.TriggerAssetDestroyed, Effect: domain.SpecialCoinGain,
				Filters: filters, CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}}},
		})
	}
	paid := func(removed domain.Mutation) int {
		emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], []domain.Mutation{removed}, factionState, nil)
		if err != nil {
			t.Fatalf("MutationReactors: %v", err)
		}
		total := 0
		for _, mutation := range emitted {
			if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" {
				total += gain.Delta
			}
		}
		return total
	}

	cases := []struct {
		name    string
		removed domain.Mutation
		want    int
	}{
		{"own loss in an attack", domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}, 1},
		{"own kill in an attack", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "alpha"}, 1},
		{"own kill by Cracked Comms", domain.AssetRemoved{FactionID: "beta", Cause: "counterstrike", CausedByFactionID: "alpha"}, 1},
		{"own loss to its own on-hit damage", domain.AssetRemoved{FactionID: "alpha", Cause: "self_damage", CausedByFactionID: "alpha"}, 1},
		{"a fight between two others (B.042)", domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"}, 0},
		{"a sale", domain.AssetRemoved{FactionID: "alpha", Cause: "sell", CausedByFactionID: "alpha"}, 0},
		{"own Base lost in an attack", domain.BaseDestroyed{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}, 1},
		{"own Base lost to an expansion's free attack", domain.BaseDestroyed{FactionID: "alpha", Cause: "expand", CausedByFactionID: "beta"}, 1},
		{"a rival Base destroyed", domain.BaseDestroyed{FactionID: "beta", Cause: "attack", CausedByFactionID: "alpha"}, 1},
		{"a Base lost between two others", domain.BaseDestroyed{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := paid(tc.removed); got != tc.want {
				t.Errorf("paid %d, want %d", got, tc.want)
			}
		})
	}

	alpha.Tags = nil
	if got := paid(domain.AssetRemoved{FactionID: "alpha", Cause: "attack", CausedByFactionID: "beta"}); got != 0 {
		t.Errorf("dropped tag: paid %d, want 0", got)
	}
}

func TestCoinGain_OncePerPaysOneRemovalATurn(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"HAWK-T-007"}, HookBudgets: map[string]int{}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "HAWK-T-007", Name: "War Profiteers"}},
		Registry: registry,
		Roller:   dicetest.Constant{Value: 1},
		Entry: domain.EffectDefinition{
			Trigger:  domain.TriggerAssetDestroyed,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction},
			Gates:    domain.EffectGates{OncePer: domain.OncePerOwnerFaction},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
		},
	})
	removals := []domain.Mutation{
		domain.AssetRemoved{FactionID: "beta", Cause: "attack", CausedByFactionID: "gamma"},
		domain.AssetRemoved{FactionID: "gamma", Cause: "attack", CausedByFactionID: "beta"},
	}
	paid := func() int {
		emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], removals, factionState, nil)
		if err != nil {
			t.Fatalf("MutationReactors: %v", err)
		}
		total := 0
		for _, mutation := range emitted {
			if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" {
				total += gain.Delta
			}
		}
		return total
	}

	if got := paid(); got != 1 {
		t.Errorf("first batch: paid %d, want 1 for two removals under once_per", got)
	}
	if got := paid(); got != 0 {
		t.Errorf("second batch: paid %d, want 0 with the budget spent", got)
	}
}

// countingBook is a rulebook with a counting holder ("coin"), a Workforce and a
// Facility definition.
func countingBook() *rulebook.Rulebook {
	return &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"coin":     {ID: "coin", Type: "Stratagem"},
		"worker":   {ID: "worker", Type: "Workforce"},
		"facility": {ID: "facility", Type: "Facility"},
	}}
}

func assetOn(id, defID, worldID string) *domain.Asset {
	asset := activeAsset(id, defID)
	asset.Location = domain.Location{WorldID: worldID}
	return asset
}

func countingEntry(max int, noStack bool) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:  domain.TriggerTurnStart,
		Effect:   domain.SpecialCoinGain,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, AssetTypes: []domain.AssetType{"Workforce"}},
		Gates:    domain.EffectGates{NoStack: noStack},
		CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}, Max: max},
	}
}

// firedCoin fires entry for holder, an asset of faction, and returns the Coin paid,
// or -1 when Fire returns no mutation.
func firedCoin(t *testing.T, faction *domain.Faction, holder *domain.Asset, entry domain.EffectDefinition) int {
	t.Helper()
	book := countingBook()
	got := effects.Registrars()[domain.SpecialCoinGain].Fire(effect.FireContext{
		Holder:   effect.Holder{Faction: faction, Asset: holder, Def: book.Assets[holder.DefinitionID]},
		Asset:    holder,
		Entry:    entry,
		Roller:   dicetest.Constant{Value: 1},
		Rulebook: book,
	})
	if len(got) == 0 {
		return -1
	}
	if len(got) != 1 {
		t.Fatalf("Fire = %+v, want one mutation", got)
	}
	gain, ok := got[0].(domain.CoinDelta)
	if !ok || gain.FactionID != faction.ID || gain.Cause != "coin_gain" {
		t.Fatalf("Fire = %+v, want a coin_gain CoinDelta to %s", got, faction.ID)
	}
	return gain.Delta
}

func factionWith(assets ...*domain.Asset) *domain.Faction {
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{}}
	for _, asset := range assets {
		faction.Assets[asset.ID] = asset
	}
	return faction
}

func TestCoinGain_CountsWorkforceOnItsPlace(t *testing.T) {
	holder := assetOn("a-coin", "coin", "veil")
	faction := factionWith(
		holder,
		assetOn("a-w1", "worker", "veil"),
		assetOn("a-w2", "worker", "veil"),
		assetOn("a-w3", "worker", "haven"),
		assetOn("a-f1", "facility", "veil"),
	)
	if got := firedCoin(t, faction, holder, countingEntry(0, false)); got != 2 {
		t.Errorf("paid %d, want 2", got)
	}
}

func TestCoinGain_CountCapsAtMax(t *testing.T) {
	holder := assetOn("a-coin", "coin", "veil")
	assets := []*domain.Asset{holder}
	for _, id := range []string{"a-w1", "a-w2", "a-w3", "a-w4", "a-w5"} {
		assets = append(assets, assetOn(id, "worker", "veil"))
	}
	faction := factionWith(assets...)
	if got := firedCoin(t, faction, holder, countingEntry(3, false)); got != 3 {
		t.Errorf("max 3: paid %d, want 3", got)
	}
	if got := firedCoin(t, faction, holder, countingEntry(0, false)); got != 5 {
		t.Errorf("max 0: paid %d, want 5", got)
	}
}

func TestCoinGain_CountRollsOncePerAsset(t *testing.T) {
	holder := assetOn("a-coin", "coin", "veil")
	faction := factionWith(
		holder,
		assetOn("a-w1", "worker", "veil"),
		assetOn("a-w2", "worker", "veil"),
		assetOn("a-w3", "worker", "veil"),
	)
	entry := countingEntry(0, false)
	entry.CoinGain.Amount = domain.DiceRoll{NumDice: 1, Sides: 4}
	book := countingBook()
	roller := &dicetest.Sequence{Values: []int{1, 4, 2}}
	got := effects.Registrars()[domain.SpecialCoinGain].Fire(effect.FireContext{
		Holder:   effect.Holder{Faction: faction, Asset: holder, Def: book.Assets[holder.DefinitionID]},
		Asset:    holder,
		Entry:    entry,
		Roller:   roller,
		Rulebook: book,
	})
	if len(got) != 1 || got[0].(domain.CoinDelta).Delta != 7 {
		t.Errorf("Fire = %+v, want one CoinDelta of 7 (1 + 4 + 2)", got)
	}
	if roller.Rolled() != 3 {
		t.Errorf("rolled %d dice, want 3", roller.Rolled())
	}
}

func TestCoinGain_CountSkipsInactiveAndInTransit(t *testing.T) {
	holder := assetOn("a-coin", "coin", "veil")
	disabled := assetOn("a-w2", "worker", "veil")
	disabled.Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}
	faction := factionWith(holder, assetOn("a-w1", "worker", "veil"), disabled)
	if got := firedCoin(t, faction, holder, countingEntry(0, false)); got != 1 {
		t.Errorf("paid %d, want 1 with one worker disabled", got)
	}

	holder.Location.WorldID = ""
	if got := firedCoin(t, faction, holder, countingEntry(0, false)); got != -1 {
		t.Errorf("in transit: paid %d, want no mutation", got)
	}
}

func TestCoinGain_NoStackLetsTheLowestIDPay(t *testing.T) {
	low := assetOn("a-coin1", "coin", "veil")
	high := assetOn("a-coin2", "coin", "veil")
	faction := factionWith(low, high, assetOn("a-w1", "worker", "veil"))
	if got := firedCoin(t, faction, high, countingEntry(0, true)); got != -1 {
		t.Errorf("no_stack, higher ID: paid %d, want no mutation", got)
	}
	if got := firedCoin(t, faction, low, countingEntry(0, true)); got != 1 {
		t.Errorf("no_stack, lower ID: paid %d, want 1", got)
	}
	if got := firedCoin(t, faction, high, countingEntry(0, false)); got != 1 {
		t.Errorf("stacking, higher ID: paid %d, want 1", got)
	}
}

// hitPayouts registers a coin_gain of 1d4 (rolling 3) on trigger for alpha's
// "fleet", dispatches mutations, and returns the Coin paid to alpha, one entry per
// payout.
func hitPayouts(t *testing.T, trigger domain.EffectTrigger, fleet *domain.Asset, mutations []domain.Mutation) []int {
	t.Helper()
	alpha := factionWith(fleet)
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialCoinGain].Register(effect.RegisterContext{
		Holder:   effect.Holder{Faction: alpha, Asset: fleet, Def: &domain.AssetDefinition{ID: "fleet-def"}},
		Registry: registry,
		Roller:   dicetest.Constant{Value: 3},
		Entry: domain.EffectDefinition{
			Trigger:  trigger,
			Effect:   domain.SpecialCoinGain,
			Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeSelf},
			CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{NumDice: 1, Sides: 4}},
		},
	})
	emitted, err := dispatch.MutationReactors(registry, factionState.Factions["beta"], mutations, factionState, nil)
	if err != nil {
		t.Fatalf("MutationReactors: %v", err)
	}
	var paid []int
	for _, mutation := range emitted {
		if gain, ok := mutation.(domain.CoinDelta); ok && gain.FactionID == "alpha" && gain.Cause == "coin_gain" && gain.CausedByFactionID == "alpha" {
			paid = append(paid, gain.Delta)
		}
	}
	return paid
}

func fleetHit(strikerID string, counter bool) domain.AssetHPDelta {
	return domain.AssetHPDelta{FactionID: "beta", AssetID: "beta-x", Delta: -2, Cause: "attack", CausedByFactionID: "alpha", CausedByAssetID: strikerID, Counter: counter}
}

func TestCoinGain_PaysOnEachHit(t *testing.T) {
	batch := []domain.Mutation{fleetHit("fleet", false), fleetHit("fleet", false), fleetHit("other", false)}
	if got := hitPayouts(t, domain.TriggerOwnAttackHit, activeAsset("fleet", "fleet-def"), batch); !reflect.DeepEqual(got, []int{3, 3}) {
		t.Errorf("attack hits: paid %v, want [3 3]", got)
	}
	if got := hitPayouts(t, domain.TriggerOwnAttackHit, activeAsset("fleet", "fleet-def"), []domain.Mutation{fleetHit("fleet", true)}); len(got) != 0 {
		t.Errorf("counter hit on own_attack_hit: paid %v, want nothing", got)
	}
}

func TestCoinGain_PaysOnEachCounterHit(t *testing.T) {
	if got := hitPayouts(t, domain.TriggerOwnCounterHit, activeAsset("fleet", "fleet-def"), []domain.Mutation{fleetHit("fleet", true)}); !reflect.DeepEqual(got, []int{3}) {
		t.Errorf("counter hit: paid %v, want [3]", got)
	}
	if got := hitPayouts(t, domain.TriggerOwnCounterHit, activeAsset("fleet", "fleet-def"), []domain.Mutation{fleetHit("fleet", false)}); len(got) != 0 {
		t.Errorf("attack hit on own_counter_hit: paid %v, want nothing", got)
	}
}

func TestCoinGain_HitPaysNothingWhenInactive(t *testing.T) {
	fleet := activeAsset("fleet", "fleet-def")
	fleet.Conditions = []domain.Condition{{Kind: domain.ConditionDisabled, SourceAssetID: "m1"}}
	if got := hitPayouts(t, domain.TriggerOwnAttackHit, fleet, []domain.Mutation{fleetHit("fleet", false)}); len(got) != 0 {
		t.Errorf("disabled: paid %v, want nothing", got)
	}
}

func TestCoinGain_TagPaysAtTurnStartUnlessLapsed(t *testing.T) {
	refiners := &domain.Tag{ID: "HAWK-T-006", Name: "Aether Refiners", Effects: []domain.EffectDefinition{{
		Trigger:  domain.TriggerTurnStart,
		Effect:   domain.SpecialCoinGain,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		CoinGain: &domain.CoinGainParams{Amount: domain.DiceRoll{Modifier: 1}},
	}}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	if err := engine.CheckTagEffects(refiners); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{refiners.ID}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha}}
	book := &rulebook.Rulebook{Tags: map[string]*domain.Tag{refiners.ID: refiners}}

	got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1})
	want := []domain.Mutation{domain.CoinDelta{FactionID: "alpha", Delta: 1, Cause: "coin_gain", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FireTurnStart = %+v, want %+v", got, want)
	}

	alpha.LapsedTags = []string{refiners.ID}
	if got := engine.FireTurnStart(alpha, factionState, book, hooks.NewRegistry(), dicetest.Constant{Value: 1}); len(got) != 0 {
		t.Errorf("lapsed: FireTurnStart = %+v, want nothing", got)
	}
}
