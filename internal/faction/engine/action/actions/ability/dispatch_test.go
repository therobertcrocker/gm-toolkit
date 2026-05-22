package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"go.uber.org/mock/gomock"
)

type seqRoller struct {
	values []int
	pos    int
}

func (r *seqRoller) Roll(_ int) int {
	v := r.values[r.pos%len(r.values)]
	r.pos++
	return v
}

func makeDispatchFixture() (*domain.Faction, *domain.Asset, *domain.AssetDefinition, *state.FactionState) {
	stealthyTarget := &domain.Asset{
		ID:           "t1",
		DefinitionID: "C1-002",
		OwnerID:      "f2",
		Location:     domain.Location{WorldID: "Anchorage"},
		Stealthy:     true,
	}
	actingAsset := &domain.Asset{
		ID:           "a1",
		DefinitionID: "C1-002",
		OwnerID:      "f1",
		Location:     domain.Location{WorldID: "Anchorage"},
	}
	actingFaction := &domain.Faction{ID: "f1", Cunning: 0}
	targetFaction := &domain.Faction{
		ID:      "f2",
		Cunning: 0,
		Assets:  map[string]*domain.Asset{"t1": stealthyTarget},
	}
	def := &domain.AssetDefinition{
		ID:    "C1-002",
		Name:  "Informers",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			AttackerStat: domain.StatCunning,
			DefenderStat: domain.StatCunning,
			Effect:       domain.EffectRevealStealth,
		},
	}
	factionState := &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": actingFaction,
			"f2": targetFaction,
		},
	}
	return actingFaction, actingAsset, def, factionState
}

// TestDispatch_Informers_AttackerWins: attack roll beats defense, stealthy target asset cleared.
func TestDispatch_Informers_AttackerWins(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)

	// attack=8, defense=3 — attacker wins
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, factionState, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1; got %v", len(mutations), mutations)
	}
	cleared, ok := mutations[0].(domain.AssetStealthCleared)
	if !ok {
		t.Fatalf("mutations[0] type = %T, want AssetStealthCleared", mutations[0])
	}
	if cleared.AssetID != "t1" || cleared.FactionID != "f2" {
		t.Errorf("AssetStealthCleared = %+v, want {t1, f2}", cleared)
	}
}

// TestDispatch_Informers_DefenderWins: defense roll beats attack, no mutations.
func TestDispatch_Informers_DefenderWins(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)

	// attack=3, defense=8 — defender wins
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, factionState, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

// TestDispatch_Stub_ConfirmApplied: W1-002 routes to confirmApplied, which calls ConfirmAbilityApplied.
func TestDispatch_Stub_ConfirmApplied(t *testing.T) {
	faction := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", DefinitionID: "W1-002", OwnerID: "f1"}
	def := &domain.AssetDefinition{
		ID:    "W1-002",
		Name:  "Harvesters",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Effect: domain.EffectCoinDrain,
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().ConfirmAbilityApplied(asset, def).Return(false, nil)

	mutations, err := Dispatch(faction, asset, def, collector, nil, factionState, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for stub handler, got %v", mutations)
	}
}
