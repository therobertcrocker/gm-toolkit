package ability

import (
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action/actions/mocks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
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
	actingFaction := &domain.Faction{ID: "f1", Cunning: 0, Assets: map[string]*domain.Asset{"a1": actingAsset}}
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

// TestNewView_CopiesEveryFaction: a change to the view's rival faction or its asset
// leaves the real state as it was.
func TestNewView_CopiesEveryFaction(t *testing.T) {
	_, _, _, factionState := makeDispatchFixture()
	view := NewView(factionState)

	if len(view.Factions) != len(factionState.Factions) {
		t.Fatalf("len(view.Factions) = %d, want %d", len(view.Factions), len(factionState.Factions))
	}
	view.Factions["f2"].Coin = 9
	view.Factions["f2"].Assets["t1"].CurrentHP = 7

	if coin := factionState.Factions["f2"].Coin; coin != 0 {
		t.Errorf("real f2.Coin = %d, want 0", coin)
	}
	if hp := factionState.Factions["f2"].Assets["t1"].CurrentHP; hp != 0 {
		t.Errorf("real t1.CurrentHP = %d, want 0", hp)
	}
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
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, nil, factionState, NewView(factionState), nil, nil)
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
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, nil, factionState, NewView(factionState), nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations when defender wins, got %v", mutations)
	}
}

// factionTestDie offers one budgeted extra d10, keep highest, on the acting side
// of a contest only.
type factionTestDie struct{}

const factionTestDieKey = "test:faction-test-die"

func (factionTestDie) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	if ctx.Phase != hooks.PhaseFactionTest {
		return nil
	}
	return []hooks.ModifierOffer{{
		Source:    "test",
		BudgetKey: factionTestDieKey,
		Apply: func(rollState *hooks.RollState) {
			rollState.AddDie(10)
			rollState.SetKeepHighest(1)
		},
	}}
}

// TestDispatch_ContestRollsThroughHooks: the contest's attack roll takes a hook's
// extra die. Rolls [3, 9, 8]: attack keeps 9, defense 8, so the gate passes. Without
// the hook, [3, 8] fails (TestDispatch_RevealStealth_DefenderWins). The offer's budget
// is spent on the real acting faction, not the cross-step clone.
func TestDispatch_ContestRollsThroughHooks(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)
	collector.EXPECT().SelectModifiers(gomock.Any()).
		DoAndReturn(func(offers []hooks.ModifierOffer) []hooks.ModifierOffer { return offers }).
		Times(1)

	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), factionTestDieKey, factionTestDie{})

	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 9, 8}}, registry, factionState, NewView(factionState), nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (gate passed); got %v", len(mutations), mutations)
	}
	if _, ok := mutations[0].(domain.AssetStealthCleared); !ok {
		t.Errorf("mutations[0] type = %T, want AssetStealthCleared", mutations[0])
	}
	if spent := faction.HookBudgets[factionTestDieKey]; spent != 1 {
		t.Errorf("real faction HookBudgets[%q] = %d, want 1", factionTestDieKey, spent)
	}
}

// defenseDie offers one budgeted extra d10, keep highest, on the defending side of
// a contest only.
type defenseDie struct{}

const defenseDieKey = "test:defense-die"

func (defenseDie) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	if ctx.Phase != hooks.PhaseContested {
		return nil
	}
	return []hooks.ModifierOffer{{
		Source:    "test",
		BudgetKey: defenseDieKey,
		Apply: func(rollState *hooks.RollState) {
			rollState.AddDie(10)
			rollState.SetKeepHighest(1)
		},
	}}
}

// TestDispatch_DefenseBudgetSpentOnRealDefender: the collector picks the view's
// copy of f2, as it does in play, and the defense hook's budget is still spent on
// the real f2 (Effort 3 Decision 3).
func TestDispatch_DefenseBudgetSpentOnRealDefender(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	view := NewView(factionState)

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: view.Factions["f2"]}, nil)
	collector.EXPECT().SelectModifiers(gomock.Any()).
		DoAndReturn(func(offers []hooks.ModifierOffer) []hooks.ModifierOffer { return offers }).
		Times(1)

	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), defenseDieKey, defenseDie{})

	if _, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3, 9}}, registry, factionState, view, nil, nil); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if spent := factionState.Factions["f2"].HookBudgets[defenseDieKey]; spent != 1 {
		t.Errorf("real f2 HookBudgets[%q] = %d, want 1", defenseDieKey, spent)
	}
	if spent := view.Factions["f2"].HookBudgets[defenseDieKey]; spent != 0 {
		t.Errorf("view f2 HookBudgets[%q] = %d, want 0", defenseDieKey, spent)
	}
}

// TestDispatch_RemovedActingAssetFizzles: an acting asset that an earlier asset in
// the action removed from the view does nothing.
func TestDispatch_RemovedActingAssetFizzles(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	view := NewView(factionState)
	delete(view.Factions["f1"].Assets, asset.ID)

	collector := mocks.NewMockCollector(gomock.NewController(t))
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, nil, factionState, view, nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("mutations = %v, want none", mutations)
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
// that destination world — proving the cross-step projection threaded the post-move
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
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{8, 3}}, nil, factionState, NewView(factionState), relocRulebook(), engine)
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
	mutations, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, nil, factionState, NewView(factionState), relocRulebook(), engine)
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

// TestDispatch_CrossStep_CoinProjection: a two-step self-relocate where the faction
// can afford exactly one move. The cross-step projection applies step 1's CoinDelta to
// the working faction, so step 2 reads Coin 0 and fails its affordability check before
// it prompts — proving the projection threads Coin, not just Location. Without it, step
// 2 would read the stale pre-action balance and overdraw the faction.
func TestDispatch_CrossStep_CoinProjection(t *testing.T) {
	acting := &domain.Asset{
		ID:           "a1",
		DefinitionID: "SF",
		OwnerID:      "f1",
		Location:     domain.Location{WorldID: "origin", RegionHex: relocHex("r1", 0)},
	}
	faction := &domain.Faction{ID: "f1", Coin: 1, Assets: map[string]*domain.Asset{"a1": acting}}
	relocateStep := domain.AbilityStep{
		Effect:   domain.EffectRelocate,
		Target:   domain.TargetNone,
		Relocate: &domain.RelocateParams{Scope: domain.RelocateSelf, Range: domain.RangeHexes, MaxHex: 1, CoinCost: 1},
	}
	def := &domain.AssetDefinition{
		ID:      "SF",
		Name:    "Courier",
		Flags:   []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{relocateStep, relocateStep}},
	}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"f1": faction}}

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectRelocateOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]action.RelocateOrder{{Asset: acting, DestinationWorld: "near"}}, nil).
		Times(1)

	_, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{0}}, nil, factionState, NewView(factionState), relocRulebook(), relocEngine())
	if err == nil {
		t.Fatal("Dispatch: expected insufficient-Coin error on step 2, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient Coin") {
		t.Errorf("Dispatch error = %q, want it to mention insufficient Coin", err.Error())
	}
}

// TestDispatch_NoTargetsLeft_Fizzles: step 2 targets an asset, but no rival asset
// shares the acting world. The asset fizzles like a failed gate: no prompt, no
// error, and step 1's Coin stands.
func TestDispatch_NoTargetsLeft_Fizzles(t *testing.T) {
	acting := &domain.Asset{ID: "a1", DefinitionID: "fixture", OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": acting}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{
		"f1": faction,
		"f2": {ID: "f2", Assets: map[string]*domain.Asset{"b1": {ID: "b1", OwnerID: "f2", Location: domain.Location{WorldID: "Hadrian"}}}},
	}}
	def := &domain.AssetDefinition{
		ID:    "fixture",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{Steps: []domain.AbilityStep{
			{
				Effect: domain.EffectCoinDelta,
				Target: domain.TargetNone,
				Roll: &domain.RollTable{
					Die:      domain.DiceRoll{NumDice: 1, Sides: 6},
					Outcomes: []domain.AbilityOutcome{{From: 1, To: 6, Coin: 1}},
				},
			},
			{Effect: domain.EffectRevealStealth, Target: domain.TargetAsset},
		}},
	}
	collector := mocks.NewMockCollector(gomock.NewController(t))

	mutations, err := Dispatch(faction, acting, def, collector, &seqRoller{values: []int{1}}, nil, factionState, NewView(factionState), nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("len(mutations) = %d, want 1 (step 1's CoinDelta); got %v", len(mutations), mutations)
	}
	if delta, ok := mutations[0].(domain.CoinDelta); !ok || delta.Delta != 1 {
		t.Errorf("mutations[0] = %+v, want CoinDelta +1", mutations[0])
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

	mutations, err := Dispatch(faction, asset, def, collector, nil, nil, factionState, NewView(factionState), nil, nil)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(mutations) != 0 {
		t.Errorf("expected no mutations for stub handler, got %v", mutations)
	}
}
