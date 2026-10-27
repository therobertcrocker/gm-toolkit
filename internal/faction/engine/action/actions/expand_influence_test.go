package actions

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

func makeExpandRulebook() *rulebook.Rulebook {
	return &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{
			"force-unit": {
				ID:       "force-unit",
				Name:     "Infantry",
				Category: domain.StatForce,
				HP:       8,
				Attack: &domain.AttackProfile{
					AttackerStat: domain.StatForce,
					DefenderStat: domain.StatCunning,
					Damage:       &domain.DiceRoll{NumDice: 1, Sides: 6},
				},
			},
			"special-unit": {
				ID:       "special-unit",
				Name:     "Treachery",
				Category: domain.StatCunning,
				HP:       2,
				Attack:   &domain.AttackProfile{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			},
		},
	}
}

func TestExpandInfluence_Validate(t *testing.T) {
	t.Run("no coin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
		faction := &domain.Faction{ID: "f1", Coin: 0, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewExpandInfluence(mocks.NewMockCollector(ctrl), nil, nil, indexFromState(factionState), nil).Validate(faction, factionState, nil) {
			t.Error("expected false when coin < 1")
		}
	})

	t.Run("no expansion options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
		// Base on Krylos at max HP and at faction.MaxHP — no new-base, no heal, no grow option.
		base := &domain.Base{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 5, MaxHP: 5, IsHomeworld: false}
		faction := &domain.Faction{ID: "f1", Coin: 3, MaxHP: 5, Assets: map[string]*domain.Asset{"a1": asset}, Bases: []*domain.Base{base}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if NewExpandInfluence(mocks.NewMockCollector(ctrl), nil, nil, indexFromState(factionState), nil).Validate(faction, factionState, nil) {
			t.Error("expected false when no expansion options available")
		}
	})

	t.Run("world available for new base", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
		faction := &domain.Faction{ID: "f1", Coin: 3, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": asset}}
		factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}
		if !NewExpandInfluence(mocks.NewMockCollector(ctrl), nil, nil, indexFromState(factionState), nil).Validate(faction, factionState, nil) {
			t.Error("expected true when faction has an asset on a world with no base")
		}
	})
}

// TestExpandInfluence_NewBase_Uncontested: no rivals on target world.
// Emits CoinDelta(-HPAmount) + BaseAdded(Ready=false).
func TestExpandInfluence_NewBase_Uncontested(t *testing.T) {
	ctrl := gomock.NewController(t)
	asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 3}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{8}}, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok || coin.Delta != -3 || coin.Cause != "expand" {
		t.Errorf("mutations[0] = %v, want CoinDelta{-3, expand}", mutations[0])
	}
	added, ok := mutations[1].(domain.BaseAdded)
	if !ok || added.Base.Location.WorldID != "Krylos" || added.Base.CurrentHP != 3 {
		t.Errorf("mutations[1] = %v, want BaseAdded{Krylos, HP=3}", mutations[1])
	}
	if added.Base.Ready {
		t.Error("new base must have Ready=false")
	}
}

// TestExpandInfluence_Reinforce_Heal: restore HP on a damaged non-homeworld base.
// Emits BaseHealed(+amount) + CoinDelta(-amount).
func TestExpandInfluence_Reinforce_Heal(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := &domain.Base{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 2, MaxHP: 5, IsHomeworld: false}
	faction := &domain.Faction{ID: "f1", Coin: 5, MaxHP: 10, Bases: []*domain.Base{base}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{
		Mode: action.ExpandModeReinforce, BaseID: "b1",
		SubMode: action.ReinforceHeal, HPAmount: 2,
	}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{1}}, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	healed, ok := mutations[0].(domain.BaseHealed)
	if !ok || healed.BaseID != "b1" || healed.Delta != 2 {
		t.Errorf("mutations[0] = %v, want BaseHealed{b1, +2}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -2 {
		t.Errorf("mutations[1] = %v, want CoinDelta{-2}", mutations[1])
	}
}

// TestExpandInfluence_Reinforce_Max: increase a non-homeworld base's MaxHP.
// Emits BaseExpanded(+amount) + CoinDelta(-amount).
func TestExpandInfluence_Reinforce_Max(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := &domain.Base{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 5, MaxHP: 5, IsHomeworld: false}
	faction := &domain.Faction{ID: "f1", Coin: 5, MaxHP: 10, Bases: []*domain.Base{base}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	order := action.ExpandInfluenceOrder{
		Mode: action.ExpandModeReinforce, BaseID: "b1",
		SubMode: action.ReinforceMax, HPAmount: 3,
	}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{1}}, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, nil); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, nil); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	expanded, ok := mutations[0].(domain.BaseExpanded)
	if !ok || expanded.BaseID != "b1" || expanded.Delta != 3 {
		t.Errorf("mutations[0] = %v, want BaseExpanded{b1, +3}", mutations[0])
	}
	coin, ok := mutations[1].(domain.CoinDelta)
	if !ok || coin.Delta != -3 {
		t.Errorf("mutations[1] = %v, want CoinDelta{-3}", mutations[1])
	}
}

// TestExpandInfluence_NewBase_Contested: rival ties the expansion roll and launches a free attack.
// Rolls: [5]=factionRoll, [5]=rivalRoll (tie → attack confirmed), [8]=attackRoll, [3]=defenseRoll, [4]=damage.
// Expects CoinDelta + BaseAdded + BaseHPDelta(-4, caused by f2) + FactionHPDelta(-4, caused by f2).
func TestExpandInfluence_NewBase_Contested(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeExpandRulebook()

	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
	f2Asset := &domain.Asset{
		ID: "a2", DefinitionID: "force-unit", OwnerID: "f2",
		Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 8, Ready: true, Maintained: true,
	}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": f1Asset}}
	rival := &domain.Faction{ID: "f2", Force: 0, Cunning: 0, Assets: map[string]*domain.Asset{"a2": f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)
	collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	collector.EXPECT().SelectBaseAttackers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{f2Asset}, nil)

	// Roll sequence: factionRoll=5, rivalRoll=5 (tie → confirmed attack), attackRoll=8, defenseRoll=3, damage=4.
	roller := &fixedRoller{values: []int{5, 5, 8, 3, 4}}
	act := NewExpandInfluence(collector, roller, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	if len(mutations) != 4 {
		t.Fatalf("len(mutations) = %d, want 4; got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.CoinDelta); !ok {
		t.Errorf("mutations[0] type = %T, want CoinDelta", mutations[0])
	}
	added, ok := mutations[1].(domain.BaseAdded)
	if !ok || added.Base.Location.WorldID != "Krylos" {
		t.Errorf("mutations[1] = %v, want BaseAdded{Krylos}", mutations[1])
	}
	baseDelta, ok := mutations[2].(domain.BaseHPDelta)
	if !ok || baseDelta.Delta != -4 || baseDelta.CausedByFactionID != "f2" {
		t.Errorf("mutations[2] = %v, want BaseHPDelta{-4, caused by f2}", mutations[2])
	}
	factionDelta, ok := mutations[3].(domain.FactionHPDelta)
	if !ok || factionDelta.Delta != -4 || factionDelta.CausedByFactionID != "f2" {
		t.Errorf("mutations[3] = %v, want FactionHPDelta{-4, caused by f2}", mutations[3])
	}
}

// TestExpandInfluence_FreeAttackTieAsksResolvers: a tied free attack asks the tie
// resolvers of both sides. Fanatical on the rival (attacker) makes it miss; on the
// expanding faction (defender) it still lands.
// Rolls: [5]=factionRoll, [5]=rivalRoll, [6]=attackRoll, [6]=defenseRoll, [4]=damage.
func TestExpandInfluence_FreeAttackTieAsksResolvers(t *testing.T) {
	cases := []struct {
		name          string
		resolverScope hooks.Scope
		outcome       hooks.TieOutcome
		wantMutations int
	}{
		{name: "no resolver: the tie lands", wantMutations: 4},
		{name: "rival loses ties: the attack misses", resolverScope: hooks.FactionScope("f2"), outcome: hooks.TieDefenderWins, wantMutations: 2},
		{name: "expander loses ties: the attack lands", resolverScope: hooks.FactionScope("f1"), outcome: hooks.TieAttackerWins, wantMutations: 4},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rulebook := makeExpandRulebook()

			f2Asset := &domain.Asset{
				ID: "a2", DefinitionID: "force-unit", OwnerID: "f2",
				Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 8, Ready: true, Maintained: true,
			}
			faction := &domain.Faction{ID: "f1", Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{}}
			rival := &domain.Faction{ID: "f2", Assets: map[string]*domain.Asset{"a2": f2Asset}}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

			registry := hooks.NewRegistry()
			if testCase.outcome != hooks.TieStandard {
				registry.RegisterTieResolver(testCase.resolverScope, "fanatical", &stubTieResolverForAttack{outcome: testCase.outcome})
			}

			order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
			collector := mocks.NewMockCollector(ctrl)
			collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)
			collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			collector.EXPECT().SelectBaseAttackers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{f2Asset}, nil)

			roller := &fixedRoller{values: []int{5, 5, 6, 6, 4}}
			act := NewExpandInfluence(collector, roller, registry, indexFromState(factionState), nil)
			if err := act.Inputs(faction, factionState, rulebook); err != nil {
				t.Fatalf("Inputs: %v", err)
			}
			if err := act.Resolve(faction, factionState, rulebook); err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			mutations, err := act.Output()
			if err != nil {
				t.Fatalf("Output: %v", err)
			}
			if len(mutations) != testCase.wantMutations {
				t.Fatalf("len(mutations) = %d, want %d; got %v", len(mutations), testCase.wantMutations, mutations)
			}
		})
	}
}

// TestExpandInfluence_SpecialAttackerNotOffered: the rival's only asset has a
// special attack, so no base attackers are offered (no SelectBaseAttackers EXPECT)
// and the new Base takes no damage. Rolls: factionRoll=5, rivalRoll=5 (tie).
func TestExpandInfluence_SpecialAttackerNotOffered(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeExpandRulebook()

	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
	f2Asset := &domain.Asset{
		ID: "a2", DefinitionID: "special-unit", OwnerID: "f2",
		Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 2, Ready: true, Maintained: true,
	}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": f1Asset}}
	rival := &domain.Faction{ID: "f2", Cunning: 0, Assets: map[string]*domain.Asset{"a2": f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)
	collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{5, 5}}, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (CoinDelta + BaseAdded, no attack); got %v", len(mutations), mutations)
	}
}

// TestExpandInfluence_BlockedAttackerNotOffered: the rival's only asset holds a
// live no_attack from f1's a1, so no base attackers are offered (no
// SelectBaseAttackers EXPECT). Rolls: factionRoll=5, rivalRoll=5 (tie).
func TestExpandInfluence_BlockedAttackerNotOffered(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeExpandRulebook()

	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
	f2Asset := &domain.Asset{
		ID: "a2", DefinitionID: "force-unit", OwnerID: "f2",
		Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 8, Ready: true, Maintained: true,
		Conditions: []domain.Condition{{Kind: domain.ConditionNoAttack, SourceAssetID: "a1"}},
	}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": f1Asset}}
	rival := &domain.Faction{ID: "f2", Force: 0, Cunning: 0, Assets: map[string]*domain.Asset{"a2": f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)
	collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	act := NewExpandInfluence(collector, &fixedRoller{values: []int{5, 5}}, nil, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (CoinDelta + BaseAdded, no attack); got %v", len(mutations), mutations)
	}
}

// contextRecorder is a RollModifier that records every RollContext it sees and
// offers nothing, so the roll and the collector are untouched.
type contextRecorder struct {
	contexts []hooks.RollContext
}

func (recorder *contextRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.contexts = append(recorder.contexts, ctx)
	return nil
}

// TestExpandInfluence_RollsCarryContext: every roll in a contested expansion goes
// through RollWithHooks with its phase, parties and world. Same roll sequence as
// TestExpandInfluence_NewBase_Contested; the hook adds no dice.
func TestExpandInfluence_RollsCarryContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	rulebook := makeExpandRulebook()

	f1Asset := &domain.Asset{ID: "a1", OwnerID: "f1", Location: domain.Location{WorldID: "Krylos"}}
	f2Asset := &domain.Asset{
		ID: "a2", DefinitionID: "force-unit", OwnerID: "f2",
		Location: domain.Location{WorldID: "Krylos"}, CurrentHP: 8, Ready: true, Maintained: true,
	}
	faction := &domain.Faction{ID: "f1", Cunning: 0, Coin: 5, MaxHP: 10, Assets: map[string]*domain.Asset{"a1": f1Asset}}
	rival := &domain.Faction{ID: "f2", Force: 0, Cunning: 0, Assets: map[string]*domain.Asset{"a2": f2Asset}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction, "f2": rival}}

	order := action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: "Krylos", HPAmount: 5}
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectExpandInfluenceOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(order, nil)
	collector.EXPECT().ConfirmRivalFreeAttack(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	collector.EXPECT().SelectBaseAttackers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Asset{f2Asset}, nil)

	recorder := &contextRecorder{}
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "test:recorder", recorder)

	roller := &fixedRoller{values: []int{5, 5, 8, 3, 4}}
	act := NewExpandInfluence(collector, roller, registry, indexFromState(factionState), nil)
	if err := act.Inputs(faction, factionState, rulebook); err != nil {
		t.Fatalf("Inputs: %v", err)
	}
	if err := act.Resolve(faction, factionState, rulebook); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	mutations, err := act.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if len(mutations) != 4 {
		t.Fatalf("len(mutations) = %d, want 4 (outcome unchanged by the hook); got %v", len(mutations), mutations)
	}

	want := []hooks.RollContext{
		{Phase: hooks.PhaseFactionTest, Actor: faction, Attribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseContested, Actor: rival, Opponent: faction, Attribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseAttack, Actor: rival, Opponent: faction, Attribute: string(domain.StatForce), Asset: f2Asset},
		{Phase: hooks.PhaseDefense, Actor: faction, Opponent: rival, Attribute: string(domain.StatCunning), OpponentAsset: f2Asset},
	}
	if len(recorder.contexts) != len(want) {
		t.Fatalf("recorded %d roll contexts, want %d: %+v", len(recorder.contexts), len(want), recorder.contexts)
	}
	for i, expected := range want {
		expected.World = order.World
		expected.Action = "Expand Influence"
		if got := recorder.contexts[i]; got != expected {
			t.Errorf("roll %d context = %+v, want %+v", i+1, got, expected)
		}
	}
}
