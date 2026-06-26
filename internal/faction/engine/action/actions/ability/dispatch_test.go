package ability

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
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
		DefinitionID: "SWN-C1-002",
		OwnerID:      "f2",
		Location:     domain.Location{WorldID: "Anchorage"},
		Stealthy:     true,
	}
	actingAsset := &domain.Asset{
		ID:           "a1",
		DefinitionID: "SWN-C1-002",
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
		ID:    "SWN-C1-002",
		Name:  "Informers",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{{
				Effect: domain.EffectRevealStealth,
				Target: domain.TargetFaction,
				Contest: &domain.Contest{
					AttackerStat: domain.StatCunning,
					DefenderStat: domain.StatCunning,
				},
			}},
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

// TestDispatch_RevealStealth_AttackerWins: attack roll beats defense, stealthy target asset cleared.
func TestDispatch_RevealStealth_AttackerWins(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	// attack=8, defense=3 — attacker wins
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, factionState, nil, nil)
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

// TestDispatch_RevealStealth_DefenderWins: defense roll beats attack, no mutations.
func TestDispatch_RevealStealth_DefenderWins(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	// attack=3, defense=8 — defender wins
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, factionState, nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

// makeSeductressFixture builds the two-step Seductress (relocate self ≤1 hex, then
// reveal faction) plus a world engine that places the Seductress at "origin" and a
// target faction's stealthed asset at the destination "near" (1 hex away).
func makeSeductressFixture() (*domain.Faction, *domain.Asset, *domain.AssetDefinition, *state.FactionState, *world.WorldEngine) {
	seductress := &domain.Asset{
		ID:           "a1",
		DefinitionID: "SWN-C2-004",
		OwnerID:      "f1",
		Location:     domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)},
	}
	stealthyTarget := &domain.Asset{
		ID:           "t1",
		DefinitionID: "SWN-C1-002",
		OwnerID:      "f2",
		Location:     domain.Location{WorldID: "near", RegionHex: relocHex("r1", 1)},
		Stealthy:     true,
	}
	actingFaction := &domain.Faction{ID: "f1", Cunning: 0, Assets: map[string]*domain.Asset{"a1": seductress}}
	targetFaction := &domain.Faction{ID: "f2", Cunning: 0, Assets: map[string]*domain.Asset{"t1": stealthyTarget}}
	def := &domain.AssetDefinition{
		ID:    "SWN-C2-004",
		Name:  "Seductress",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{
				{
					Effect:   domain.EffectRelocate,
					Target:   domain.TargetNone,
					Relocate: &domain.RelocateParams{Scope: domain.RelocateSelf, Range: domain.RangeHexes, MaxHex: 1},
				},
				{
					Effect:  domain.EffectRevealStealth,
					Target:  domain.TargetFaction,
					Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
				},
			},
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": actingFaction, "f2": targetFaction}}
	return actingFaction, seductress, def, factionState, relocEngine()
}

// TestDispatch_Seductress_RelocateThenReveal_Wins: the two-step ability moves the
// Seductress to "near", and the won contest reveals the target's stealthed asset at
// that destination world — proving applyAssetMovedView threaded the post-move
// location into the reveal step.
func TestDispatch_Seductress_RelocateThenReveal_Wins(t *testing.T) {
	faction, asset, def, factionState, engine := makeSeductressFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: asset, DestinationWorld: "near"}}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	// attack=8, defense=3 — attacker wins
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, factionState, nil, engine)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 2 {
		t.Fatalf("len(mutations) = %d, want 2 (AssetMoved + AssetStealthCleared); got %v", len(mutations), mutations)
	}
	moved, ok := mutations[0].(domain.AssetMoved)
	if !ok || moved.AssetID != "a1" || moved.ToLocation.WorldID != "near" {
		t.Fatalf("mutations[0] = %+v, want AssetMoved a1 -> near", mutations[0])
	}
	cleared, ok := mutations[1].(domain.AssetStealthCleared)
	if !ok || cleared.AssetID != "t1" {
		t.Errorf("mutations[1] = %+v, want AssetStealthCleared t1 (revealed at the destination world)", mutations[1])
	}
}

// TestDispatch_Seductress_RelocateThenReveal_Loses: the relocate (ungated, first)
// still applies; the lost contest halts the chain so only the AssetMoved survives —
// proving halt-on-failure preserves the earlier ungated step.
func TestDispatch_Seductress_RelocateThenReveal_Loses(t *testing.T) {
	faction, asset, def, factionState, engine := makeSeductressFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: asset, DestinationWorld: "near"}}, nil)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	// attack=3, defense=8 — defender wins, reveal halts
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, factionState, nil, engine)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (only AssetMoved); got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.AssetMoved); !ok {
		t.Errorf("mutations[0] type = %T, want AssetMoved", mutations[0])
	}
}

// TestDispatch_Stub_ConfirmApplied: an effect with no registered handler (CoinDrain) falls through to confirmApplied, which calls ConfirmAbilityApplied.
func TestDispatch_Stub_ConfirmApplied(t *testing.T) {
	faction := &domain.Faction{ID: "f1"}
	asset := &domain.Asset{ID: "a1", DefinitionID: "SWN-W1-002", OwnerID: "f1"}
	def := &domain.AssetDefinition{
		ID:    "SWN-W1-002",
		Name:  "Harvesters",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{{
				Effect: domain.AbilityEffectType("not_built"),
				Target: domain.TargetNone,
			}},
		},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().ConfirmAbilityApplied(asset, def).Return(false, nil)

	mutations, err := Dispatch(faction, asset, def, collector, nil, factionState, nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for stub handler, got %v", mutations)
	}
}
