package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// fixedRoller returns values from a pre-set sequence, cycling if exhausted.
type fixedRoller struct {
	values []int
	index  int
}

func (roller *fixedRoller) Roll(_ int) int {
	val := roller.values[roller.index%len(roller.values)]
	roller.index++
	return val
}

// fakeCollector pre-configures responses for ability input methods.
type fakeCollector struct {
	moveDestination    string
	factionTestTarget  *domain.Faction
}

func (collector *fakeCollector) SelectMoveDestination(_ *domain.Asset, _ []string) (string, error) {
	return collector.moveDestination, nil
}

func (collector *fakeCollector) SelectFactionTestTarget(_ *domain.Asset, _ domain.AbilityEffectType, _ []*domain.Faction) (*domain.Faction, error) {
	return collector.factionTestTarget, nil
}

func makeAbilityState(actingFactionID, assetLocation string) (*state.FactionState, *domain.Faction, *domain.Asset) {
	asset := &domain.Asset{
		ID:       "a1",
		OwnerID:  actingFactionID,
		Location: assetLocation,
	}
	acting := &domain.Faction{
		ID:     actingFactionID,
		Force:  0,
		Assets: []*domain.Asset{asset},
	}
	rival := &domain.Faction{
		ID:     "f2",
		Force:  0,
		Assets: []*domain.Asset{{ID: "r1", OwnerID: "f2", Location: assetLocation}},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			actingFactionID: acting,
			"f2":            rival,
		},
	}
	return factionState, acting, asset
}

func TestAbilityEngine_Movement(t *testing.T) {
	ae := New()
	factionState, acting, asset := makeAbilityState("f1", "Tartarus")

	def := &domain.AssetDefinition{
		ID: "scout-def",
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{
				{Type: domain.AbilityStepMovement, CoinCost: 0},
			},
		},
	}
	collector := &fakeCollector{moveDestination: "Krylos"}

	mutations, err := ae.Run(acting, asset, def, collector, &fixedRoller{values: []int{5}}, factionState, &rulebook.Rulebook{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
	if moved.AssetID != "a1" || moved.FromLocation != "Tartarus" || moved.ToLocation != "Krylos" {
		t.Errorf("AssetMoved = %+v, want {a1, Tartarus → Krylos}", moved)
	}
}

func TestAbilityEngine_Movement_WithCoinCost(t *testing.T) {
	ae := New()
	factionState, acting, asset := makeAbilityState("f1", "Tartarus")

	def := &domain.AssetDefinition{
		ID: "scout-def",
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{
				{Type: domain.AbilityStepMovement, CoinCost: 2},
			},
		},
	}
	collector := &fakeCollector{moveDestination: "Krylos"}

	mutations, err := ae.Run(acting, asset, def, collector, &fixedRoller{values: []int{5}}, factionState, &rulebook.Rulebook{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// CoinDelta + AssetMoved
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2; got %v", len(mutations), mutations)
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok || coin.Delta != -2 {
		t.Errorf("mutations[0] = %v, want CoinDelta{-2}", mutations[0])
	}
	if _, ok := mutations[1].(domain.AssetMoved); !ok {
		t.Errorf("mutations[1] type = %T, want AssetMoved", mutations[1])
	}
}

func TestAbilityEngine_FactionTest_AttackerWins(t *testing.T) {
	ae := New()
	factionState, acting, asset := makeAbilityState("f1", "Tartarus")
	rival := factionState.Factions["f2"]

	drainAmount := 3
	def := &domain.AssetDefinition{
		ID: "spy-def",
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{
				{
					Type:         domain.AbilityStepFactionTest,
					AttackerStat: domain.StatCunning,
					DefenderStat: domain.StatCunning,
					Effect:       domain.EffectCoinDrain,
					EffectDice:   &domain.DiceRoll{NumDice: 1, Sides: 6, Modifier: 0},
				},
			},
		},
	}
	collector := &fakeCollector{factionTestTarget: rival}
	// Rolls: attack=8, defense=3, effect_dice=drainAmount
	roller := &fixedRoller{values: []int{8, 3, drainAmount}}

	mutations, err := ae.Run(acting, asset, def, collector, roller, factionState, &rulebook.Rulebook{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	coin, ok := mutations[0].(domain.CoinDelta)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want CoinDelta", mutations[0])
	}
	if coin.FactionID != "f2" || coin.Delta != -drainAmount {
		t.Errorf("CoinDelta = {%s, %d}, want {f2, -%d}", coin.FactionID, coin.Delta, drainAmount)
	}
}

func TestAbilityEngine_FactionTest_DefenderWins(t *testing.T) {
	ae := New()
	factionState, acting, asset := makeAbilityState("f1", "Tartarus")
	rival := factionState.Factions["f2"]

	def := &domain.AssetDefinition{
		ID: "spy-def",
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{
				{
					Type:         domain.AbilityStepFactionTest,
					AttackerStat: domain.StatCunning,
					DefenderStat: domain.StatCunning,
					Effect:       domain.EffectCoinDrain,
					EffectDice:   &domain.DiceRoll{NumDice: 1, Sides: 6},
				},
			},
		},
	}
	collector := &fakeCollector{factionTestTarget: rival}
	// Rolls: attack=3, defense=8 — defender wins, no effect applied
	roller := &fixedRoller{values: []int{3, 8}}

	mutations, err := ae.Run(acting, asset, def, collector, roller, factionState, &rulebook.Rulebook{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

func TestAbilityEngine_NoAbility(t *testing.T) {
	ae := New()
	factionState, acting, asset := makeAbilityState("f1", "Tartarus")

	def := &domain.AssetDefinition{ID: "plain-def"}

	mutations, err := ae.Run(acting, asset, def, &fakeCollector{}, &fixedRoller{values: []int{1}}, factionState, &rulebook.Rulebook{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for definition with no Ability, got %v", mutations)
	}
}
