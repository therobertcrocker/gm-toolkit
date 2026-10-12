package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// rollCollector takes every offer and reroll when accept is set, declines them
// otherwise, and counts how often it was asked.
type rollCollector struct {
	accept     bool
	selections int
	rerolls    int
}

func (collector *rollCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	collector.selections++
	if collector.accept {
		return offers
	}
	return nil
}

func (collector *rollCollector) ConfirmReroll(hooks.RerollDirective) bool {
	collector.rerolls++
	return collector.accept
}

func bonusDieEntry(filters domain.EffectFilters, count int) domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  filters,
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: count},
	}
}

var integralProtocolsFilters = domain.EffectFilters{
	Owner: domain.OwnerOwn, Scope: domain.ScopeSelf,
	Phases:    []domain.EffectRollPhase{domain.RollPhaseDefense},
	Attribute: domain.StatCunning,
}

// bonusDieState puts alpha's asset on Krylos; beta is the rival there.
func bonusDieState(asset *domain.Asset) *state.FactionState {
	asset.Location = krylos
	return &state.FactionState{Factions: map[string]*domain.Faction{
		"alpha": {ID: "alpha", Cunning: 4, Assets: map[string]*domain.Asset{asset.ID: asset}},
		"beta":  {ID: "beta", Cunning: 3},
	}}
}

// registerBonusDie registers entry for alpha's asset and returns the registry.
func registerBonusDie(factionState *state.FactionState, assetID string, def *domain.AssetDefinition, entry domain.EffectDefinition) *hooks.Registry {
	registry := hooks.NewRegistry()
	alpha := factionState.Factions["alpha"]
	effects.Registrars()[domain.SpecialBonusDie].Register(effect.RegisterContext{
		Faction: alpha, Asset: alpha.Assets[assetID], Def: def, Entry: entry, Registry: registry,
	})
	return registry
}

func bonusDieOffers(registry *hooks.Registry, assetID string, ctx hooks.RollContext, factionState *state.FactionState) []hooks.ModifierOffer {
	var offers []hooks.ModifierOffer
	for _, registered := range registry.RollModifiersAt(hooks.AssetScope("alpha", assetID)) {
		offers = append(offers, registered.Hook.OfferModifiers(ctx, factionState, nil)...)
	}
	return offers
}

// cunningDefense is alpha's asset ip defending with Cunning against beta.
func cunningDefense(factionState *state.FactionState, attribute string) hooks.RollContext {
	alpha := factionState.Factions["alpha"]
	return hooks.RollContext{
		Phase: hooks.PhaseDefense, Actor: alpha, Opponent: factionState.Factions["beta"],
		Attribute: attribute, Asset: alpha.Assets["ip"], World: "Krylos", Action: "Attack",
	}
}

var integralProtocolsDef = &domain.AssetDefinition{ID: "SWN-F7-002", Name: "Integral Protocols"}

func TestBonusDie_MatchingRollOffersDieWithKeepHighest(t *testing.T) {
	cases := []struct {
		count    int
		wantDice []int
	}{
		{1, []int{10}},
		{2, []int{10, 10}},
	}
	for _, tc := range cases {
		factionState := bonusDieState(activeAsset("ip", "SWN-F7-002"))
		registry := registerBonusDie(factionState, "ip", integralProtocolsDef, bonusDieEntry(integralProtocolsFilters, tc.count))

		offers := bonusDieOffers(registry, "ip", cunningDefense(factionState, "Cunning"), factionState)
		if len(offers) != 1 {
			t.Fatalf("count %d: offers = %d, want 1", tc.count, len(offers))
		}
		if offers[0].Source != "Integral Protocols" {
			t.Errorf("count %d: Source = %q, want the definition name", tc.count, offers[0].Source)
		}
		var rollState hooks.RollState
		offers[0].Apply(&rollState)
		if !reflect.DeepEqual(rollState.ExtraDice(), tc.wantDice) || rollState.KeepHighest() != 1 {
			t.Errorf("count %d: extra dice %v keep %d, want %v keep 1", tc.count, rollState.ExtraDice(), rollState.KeepHighest(), tc.wantDice)
		}
	}
}

func TestBonusDie_NonMatchingRollOffersNothing(t *testing.T) {
	factionState := bonusDieState(activeAsset("ip", "SWN-F7-002"))
	registry := registerBonusDie(factionState, "ip", integralProtocolsDef, bonusDieEntry(integralProtocolsFilters, 1))

	if offers := bonusDieOffers(registry, "ip", cunningDefense(factionState, "Force"), factionState); len(offers) != 0 {
		t.Errorf("Force defense offers = %d, want 0", len(offers))
	}
}

func TestBonusDie_InactiveAssetOffersNothing(t *testing.T) {
	for name, deactivate := range inactiveStates {
		t.Run(name, func(t *testing.T) {
			asset := activeAsset("ip", "SWN-F7-002")
			deactivate(asset)
			factionState := bonusDieState(asset)
			registry := registerBonusDie(factionState, "ip", integralProtocolsDef, bonusDieEntry(integralProtocolsFilters, 1))

			if offers := bonusDieOffers(registry, "ip", cunningDefense(factionState, "Cunning"), factionState); len(offers) != 0 {
				t.Errorf("offers = %d, want 0", len(offers))
			}
		})
	}
}

func TestBonusDie_EndToEndThroughDispatcher(t *testing.T) {
	panopticonFilters := domain.EffectFilters{
		Owner: domain.OwnerOwn, Scope: domain.ScopeWorld,
		Phases:    []domain.EffectRollPhase{domain.RollPhaseAttack, domain.RollPhaseDefense},
		Attribute: domain.StatCunning,
	}
	factionState := bonusDieState(activeAsset("pm", "SWN-C8-001"))
	registry := registerBonusDie(factionState, "pm", &domain.AssetDefinition{ID: "SWN-C8-001", Name: "Panopticon Matrix"}, bonusDieEntry(panopticonFilters, 1))
	alpha, beta := factionState.Factions["alpha"], factionState.Factions["beta"]

	t.Run("owner's Cunning attack gets the die", func(t *testing.T) {
		collector := &rollCollector{accept: true}
		ctx := hooks.RollContext{Phase: hooks.PhaseAttack, Actor: alpha, Opponent: beta, Attribute: "Cunning", World: "Krylos", Action: "Attack"}
		result := dispatch.RollWithHooks(ctx, domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: alpha.Cunning}, registry, collector, &seqRoller{values: []int{3, 9}}, alpha, factionState, nil)
		if result.Sum != 9+alpha.Cunning {
			t.Errorf("Sum = %d, want %d", result.Sum, 9+alpha.Cunning)
		}
	})

	t.Run("rival's defense against the owner gets no die", func(t *testing.T) {
		collector := &rollCollector{accept: true}
		ctx := hooks.RollContext{Phase: hooks.PhaseDefense, Actor: beta, Opponent: alpha, Attribute: "Cunning", World: "Krylos", Action: "Attack"}
		result := dispatch.RollWithHooks(ctx, domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: beta.Cunning}, registry, collector, &seqRoller{values: []int{3}}, beta, factionState, nil)
		if result.Sum != 3+beta.Cunning || collector.selections != 0 {
			t.Errorf("Sum = %d, selections = %d; want %d and 0", result.Sum, collector.selections, 3+beta.Cunning)
		}
	})
}
