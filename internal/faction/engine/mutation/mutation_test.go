package mutation

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func newMutationTestState() *state.FactionState {
	return &state.FactionState{
		Factions: map[string]*domain.Faction{
			"f1": {
				ID:   "f1",
				Name: "Faction One",
				Coin: 10,
				Assets: map[string]*domain.Asset{
					"a1": {ID: "a1", DefinitionID: "infantry", Location: domain.Location{WorldID: "Anchorage"}, Maintained: true},
					"a2": {ID: "a2", DefinitionID: "spy_net", Location: domain.Location{WorldID: "Anchorage"}, Maintained: false},
				},
			},
		},
	}
}

func TestMutationEngine_CoinDelta(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: 5},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Coin; got != 15 {
		t.Errorf("Coin = %d, want 15", got)
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "f1", Delta: -3},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Coin; got != 12 {
		t.Errorf("Coin = %d, want 12", got)
	}
}

func TestMutationEngine_AssetRemoved(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetRemoved{FactionID: "f1", AssetID: "a1"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	assets := s.Factions["f1"].Assets
	if len(assets) != 1 {
		t.Fatalf("len(Assets) = %d, want 1", len(assets))
	}
	if assets["a2"].ID != "a2" {
		t.Errorf("remaining asset ID = %q, want %q", assets["a2"].ID, "a2")
	}
}

func TestMutationEngine_PurchaseDiscount(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.PurchaseDiscountGranted{FactionID: "f1", Amount: 5},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].PendingPurchaseDiscount; got != 5 {
		t.Errorf("PendingPurchaseDiscount = %d, want 5 after grant", got)
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.PurchaseDiscountConsumed{FactionID: "f1"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].PendingPurchaseDiscount; got != 0 {
		t.Errorf("PendingPurchaseDiscount = %d, want 0 after consume", got)
	}
}

func TestMutationEngine_AssetMaintainedFlag(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a1", Maintained: false},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if s.Factions["f1"].Assets["a1"].Maintained {
		t.Error("expected a1 to be unmaintained")
	}

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetMaintainedFlag{FactionID: "f1", AssetID: "a2", Maintained: true},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if !s.Factions["f1"].Assets["a2"].Maintained {
		t.Error("expected a2 to be maintained")
	}
}

func TestMutationEngine_UnknownFactionErrors(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	err := me.Apply(s, []domain.Mutation{
		domain.CoinDelta{FactionID: "missing", Delta: 100},
	}, slog.New(slog.DiscardHandler))
	if err == nil {
		t.Fatal("Apply() expected error for unknown faction, got nil")
	}
	if len(err.Misses) != 1 {
		t.Fatalf("len(Misses) = %d, want 1", len(err.Misses))
	}
	if got := s.Factions["f1"].Coin; got != 10 {
		t.Errorf("Coin = %d, want 10 (unrelated faction must be unmodified)", got)
	}
}

func TestApply_GoalFailedClearsTheGoal(t *testing.T) {
	t.Run("the holder's goal is cleared and nothing else moves", func(t *testing.T) {
		factionState := newMutationTestState()
		holder := factionState.Factions["f1"]
		holder.XP = 3
		holder.ActiveGoal = &domain.ActiveGoal{GoalID: "CORE-G-008", TargetFactionID: "f2"}
		mutationEngine := New(slog.New(slog.DiscardHandler))

		if err := mutationEngine.Apply(factionState, []domain.Mutation{
			domain.GoalFailed{FactionID: "f1", GoalID: "CORE-G-008", Cause: domain.GoalFailedTargetLost},
		}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply() unexpected error: %v", err)
		}
		if holder.ActiveGoal != nil {
			t.Errorf("ActiveGoal = %+v, want nil", holder.ActiveGoal)
		}
		if holder.XP != 3 {
			t.Errorf("XP = %d, want 3", holder.XP)
		}
		if holder.Coin != 10 {
			t.Errorf("Coin = %d, want 10", holder.Coin)
		}
	})

	t.Run("an unknown faction is a miss", func(t *testing.T) {
		factionState := newMutationTestState()
		mutationEngine := New(slog.New(slog.DiscardHandler))

		err := mutationEngine.Apply(factionState, []domain.Mutation{
			domain.GoalFailed{FactionID: "missing", GoalID: "CORE-G-008", Cause: domain.GoalFailedTargetLost},
		}, slog.New(slog.DiscardHandler))
		if err == nil {
			t.Fatal("Apply() expected error for unknown faction, got nil")
		}
		if len(err.Misses) != 1 {
			t.Fatalf("len(Misses) = %d, want 1", len(err.Misses))
		}
	})
}

func TestApply_GoalPickedSetsTheGoal(t *testing.T) {
	t.Run("the picker holds the goal", func(t *testing.T) {
		factionState := newMutationTestState()
		picker := factionState.Factions["f1"]
		mutationEngine := New(slog.New(slog.DiscardHandler))
		picked := domain.GoalPicked{FactionID: "f1", GoalID: "CORE-G-009", StealthedAtStart: []string{"a2"}}

		if err := mutationEngine.Apply(factionState, []domain.Mutation{picked}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply() unexpected error: %v", err)
		}
		want := &domain.ActiveGoal{GoalID: "CORE-G-009", Progress: 0, StealthedAtStart: []string{"a2"}}
		if !reflect.DeepEqual(picker.ActiveGoal, want) {
			t.Fatalf("ActiveGoal = %+v, want %+v", picker.ActiveGoal, want)
		}
		picked.StealthedAtStart[0] = "a1"
		if picker.ActiveGoal.StealthedAtStart[0] != "a2" {
			t.Errorf("StealthedAtStart = %v, want it apart from the mutation's slice", picker.ActiveGoal.StealthedAtStart)
		}
	})

	t.Run("the target is kept", func(t *testing.T) {
		factionState := newMutationTestState()
		mutationEngine := New(slog.New(slog.DiscardHandler))

		if err := mutationEngine.Apply(factionState, []domain.Mutation{
			domain.GoalPicked{FactionID: "f1", GoalID: "CORE-G-008", TargetFactionID: "f2"},
		}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply() unexpected error: %v", err)
		}
		if got := factionState.Factions["f1"].ActiveGoal.TargetFactionID; got != "f2" {
			t.Errorf("TargetFactionID = %q, want f2", got)
		}
	})

	t.Run("an unknown faction is a miss", func(t *testing.T) {
		factionState := newMutationTestState()
		mutationEngine := New(slog.New(slog.DiscardHandler))

		err := mutationEngine.Apply(factionState, []domain.Mutation{
			domain.GoalPicked{FactionID: "missing", GoalID: "CORE-G-009"},
		}, slog.New(slog.DiscardHandler))
		if err == nil {
			t.Fatal("Apply() expected error for unknown faction, got nil")
		}
		if len(err.Misses) != 1 {
			t.Fatalf("len(Misses) = %d, want 1", len(err.Misses))
		}
	})
}

func TestApply_GoalCompletedRecordsTheRatings(t *testing.T) {
	factionState := newMutationTestState()
	holder := factionState.Factions["f1"]
	holder.Force, holder.Cunning, holder.Wealth = 3, 2, 4
	holder.ActiveGoal = &domain.ActiveGoal{GoalID: "CORE-G-011"}
	mutationEngine := New(slog.New(slog.DiscardHandler))

	if err := mutationEngine.Apply(factionState, []domain.Mutation{
		domain.GoalCompleted{FactionID: "f1", GoalID: "CORE-G-011"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if holder.ActiveGoal != nil {
		t.Errorf("ActiveGoal = %+v, want nil", holder.ActiveGoal)
	}
	wantWealthOfWorlds := map[domain.FactionStat]int{domain.StatForce: 3, domain.StatCunning: 2, domain.StatWealth: 4}
	if got := holder.CompletionRatings["CORE-G-011"]; !reflect.DeepEqual(got, wantWealthOfWorlds) {
		t.Fatalf("CompletionRatings[CORE-G-011] = %v, want %v", got, wantWealthOfWorlds)
	}

	holder.Wealth = 5
	if err := mutationEngine.Apply(factionState, []domain.Mutation{
		domain.GoalPicked{FactionID: "f1", GoalID: "CORE-G-001"},
		domain.GoalCompleted{FactionID: "f1", GoalID: "CORE-G-001"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if len(holder.CompletionRatings) != 2 {
		t.Errorf("CompletionRatings has %d entries, want 2", len(holder.CompletionRatings))
	}
	if got := holder.CompletionRatings["CORE-G-011"][domain.StatWealth]; got != 4 {
		t.Errorf("CORE-G-011 Wealth = %d, want 4", got)
	}
}

func TestMutationEngine_AssetStealthCleared(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f1"].Assets["a1"].Stealthy = true
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.AssetStealthCleared{FactionID: "f1", AssetID: "a1"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if s.Factions["f1"].Assets["a1"].Stealthy {
		t.Error("expected a1 Stealthy to be false after AssetStealthCleared")
	}
}

func newMutationTestStateWithBase() *state.FactionState {
	s := newMutationTestState()
	s.Factions["f1"].Bases = []*domain.Base{
		{ID: "b1", OwnerID: "f1", Location: domain.Location{WorldID: "Anchorage"}, CurrentHP: 10, MaxHP: 10},
	}
	return s
}

func TestMutationEngine_BaseHPDelta(t *testing.T) {
	s := newMutationTestStateWithBase()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.BaseHPDelta{FactionID: "f1", BaseID: "b1", Delta: -4},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := s.Factions["f1"].Bases[0].CurrentHP; got != 6 {
		t.Errorf("Base CurrentHP = %d, want 6", got)
	}
}

func TestMutationEngine_BaseDestroyed(t *testing.T) {
	s := newMutationTestStateWithBase()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{
		domain.BaseDestroyed{FactionID: "f1", BaseID: "b1"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if got := len(s.Factions["f1"].Bases); got != 0 {
		t.Errorf("len(Bases) = %d, want 0 after BaseDestroyed", got)
	}
}

func TestMutationEngine_AssetCondition(t *testing.T) {
	factionState := newMutationTestState()
	engine := New(slog.New(slog.DiscardHandler))
	fromSab1 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1", ExpiresAtTurnOf: "f2"}
	fromSab2 := domain.Condition{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-2", ExpiresAtTurnOf: "f2"}
	apply := func(mutation domain.Mutation) {
		t.Helper()
		if err := engine.Apply(factionState, []domain.Mutation{mutation}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply(%s) unexpected error: %v", mutation.Type(), err)
		}
	}
	conditions := func() []domain.Condition { return factionState.Factions["f1"].Assets["a1"].Conditions }

	apply(domain.AssetConditionApplied{FactionID: "f1", AssetID: "a1", Condition: fromSab1})
	apply(domain.AssetConditionApplied{FactionID: "f1", AssetID: "a1", Condition: fromSab1})
	if got := len(conditions()); got != 1 {
		t.Fatalf("after re-applying one source: len(Conditions) = %d, want 1", got)
	}

	apply(domain.AssetConditionApplied{FactionID: "f1", AssetID: "a1", Condition: fromSab2})
	if got := len(conditions()); got != 2 {
		t.Fatalf("after a second source: len(Conditions) = %d, want 2", got)
	}

	apply(domain.AssetConditionCleared{FactionID: "f1", AssetID: "a1", Condition: fromSab1})
	if got := conditions(); len(got) != 1 || got[0].SourceAssetID != "sab-2" {
		t.Fatalf("after clearing sab-1: Conditions = %+v, want only sab-2", got)
	}

	err := engine.Apply(factionState, []domain.Mutation{
		domain.AssetConditionApplied{FactionID: "f1", AssetID: "ghost", Condition: fromSab1},
	}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].EntityID != "ghost" {
		t.Fatalf("Apply on missing asset: err = %v, want one miss with EntityID ghost", err)
	}
}

func newOwnerChangeTestState() *state.FactionState {
	factionState := newMutationTestState()
	asset := factionState.Factions["f1"].Assets["a1"]
	asset.OwnerID = "f1"
	asset.CurrentOrder = &domain.MovementOrder{AssetID: "a1", StepIdx: 2}
	asset.Conditions = []domain.Condition{{Kind: domain.ConditionNoAbility, SourceAssetID: "sab-1"}}
	factionState.Factions["f2"] = &domain.Faction{ID: "f2", Name: "Faction Two"}
	return factionState
}

func TestMutationEngine_AssetOwnerChanged(t *testing.T) {
	factionState := newOwnerChangeTestState()
	engine := New(slog.New(slog.DiscardHandler))

	if err := engine.Apply(factionState, []domain.Mutation{
		domain.AssetOwnerChanged{FactionID: "f1", AssetID: "a1", ToFactionID: "f2"},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if _, ok := factionState.Factions["f1"].Assets["a1"]; ok {
		t.Error("a1 still in f1 after the owner change")
	}
	asset, ok := factionState.Factions["f2"].Assets["a1"]
	if !ok {
		t.Fatal("a1 not in f2 after the owner change")
	}
	if asset.ID != "a1" || asset.OwnerID != "f2" {
		t.Errorf("asset = {ID: %q, OwnerID: %q}, want {a1, f2}", asset.ID, asset.OwnerID)
	}
	if asset.CurrentOrder == nil || asset.CurrentOrder.StepIdx != 2 {
		t.Errorf("CurrentOrder = %+v, want the order carried over at StepIdx 2", asset.CurrentOrder)
	}
	if !asset.HasCondition(domain.ConditionNoAbility) {
		t.Error("condition lost in the owner change")
	}

	err := engine.Apply(factionState, []domain.Mutation{
		domain.AssetOwnerChanged{FactionID: "f2", AssetID: "a1", ToFactionID: "ghost"},
	}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].FactionID != "ghost" || err.Misses[0].EntityID != "" {
		t.Fatalf("Apply to unknown faction: err = %v, want one faction miss on ghost", err)
	}
}

func TestMutationEngine_OwnerChangeAppliesLast(t *testing.T) {
	factionState := newOwnerChangeTestState()
	engine := New(slog.New(slog.DiscardHandler))

	if err := engine.Apply(factionState, []domain.Mutation{
		domain.AssetOwnerChanged{FactionID: "f1", AssetID: "a1", ToFactionID: "f2"},
		domain.AssetHPDelta{FactionID: "f1", AssetID: "a1", Delta: -2},
	}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply() unexpected error: %v (the HP delta addressed to the old owner must land first)", err)
	}
	asset, ok := factionState.Factions["f2"].Assets["a1"]
	if !ok {
		t.Fatal("a1 not in f2 after the batch")
	}
	if asset.CurrentHP != -2 {
		t.Errorf("CurrentHP = %d, want -2", asset.CurrentHP)
	}
}

func TestMutationEngine_WorldGovernmentChanged(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f2"] = &domain.Faction{ID: "f2"}
	me := New(slog.New(slog.DiscardHandler))
	apply := func(mutation domain.Mutation) *MutationApplyError {
		return me.Apply(s, []domain.Mutation{mutation}, slog.New(slog.DiscardHandler))
	}

	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", ToFactionID: "f1"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := s.Ruler("Anchorage"); got != "f1" {
		t.Errorf("Ruler after f1 takes it = %q, want f1", got)
	}
	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f1", ToFactionID: "f2"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := s.Ruler("Anchorage"); got != "f2" {
		t.Errorf("Ruler after f2 takes it = %q, want f2", got)
	}
	err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f2", ToFactionID: "missing"})
	if err == nil || len(err.Misses) != 1 {
		t.Fatalf("Apply to a missing faction = %v, want one miss", err)
	}
	if got := s.Ruler("Anchorage"); got != "f2" {
		t.Errorf("Ruler after a miss = %q, want f2 unchanged", got)
	}
	if err := apply(domain.WorldGovernmentChanged{WorldID: "Anchorage", FromFactionID: "f2"}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, ok := s.Governments["Anchorage"]; ok {
		t.Error("an empty ToFactionID left an entry, want the world independent")
	}
}

func TestMutationEngine_TagRemoved(t *testing.T) {
	s := newMutationTestState()
	s.Factions["f1"].Tags = []string{"SWN-T-001", "SWN-T-011"}
	me := New(slog.New(slog.DiscardHandler))
	removal := domain.TagRemoved{FactionID: "f1", TagID: "SWN-T-011"}

	if err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if tags := s.Factions["f1"].Tags; len(tags) != 1 || tags[0] != "SWN-T-001" {
		t.Errorf("Tags = %v, want [SWN-T-001]", tags)
	}
	err := me.Apply(s, []domain.Mutation{removal}, slog.New(slog.DiscardHandler))
	if err == nil || len(err.Misses) != 1 || err.Misses[0].EntityID != "SWN-T-011" {
		t.Errorf("removing a tag not held = %v, want one miss naming SWN-T-011", err)
	}
}

func TestMutationEngine_TagAdded(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	if err := me.Apply(s, []domain.Mutation{domain.TagAdded{FactionID: "f1", TagID: "SWN-T-020"}}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !s.Factions["f1"].HasTag("SWN-T-020") {
		t.Errorf("Tags = %v, want SWN-T-020 held", s.Factions["f1"].Tags)
	}
}

func TestMutationEngine_PermissionRefusedIsRecordOnly(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))

	refusal := domain.PermissionRefused{FactionID: "missing", WorldID: "Anchorage", DefinitionID: "gunships", Cause: "buy"}
	if err := me.Apply(s, []domain.Mutation{refusal}, slog.New(slog.DiscardHandler)); err != nil {
		t.Errorf("Apply = %v, want no miss for a record-only mutation", err)
	}
	if got := s.Factions["f1"].Coin; got != 10 {
		t.Errorf("Coin = %d, want 10 unchanged", got)
	}
}

func TestApply_ActionTakenChangesNoState(t *testing.T) {
	cases := []struct {
		name      string
		factionID string
	}{
		{name: "held faction", factionID: "f1"},
		{name: "unknown faction", factionID: "missing"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			s := newMutationTestState()
			before := s.Factions["f1"].Clone()
			me := New(slog.New(slog.DiscardHandler))

			actionTaken := domain.ActionTaken{FactionID: testCase.factionID, Action: "Attack"}
			if err := me.Apply(s, []domain.Mutation{actionTaken}, slog.New(slog.DiscardHandler)); err != nil {
				t.Errorf("Apply = %v, want no miss for a record-only mutation", err)
			}
			if !reflect.DeepEqual(s.Factions["f1"], before) {
				t.Errorf("faction after Apply = %+v, want %+v unchanged", s.Factions["f1"], before)
			}
		})
	}
}

func TestMutationEngine_MovementOrderHeld(t *testing.T) {
	s := newMutationTestState()
	hexes := []spatial.RegionHex{{RegionID: "r", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 1}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 2}}}
	s.Factions["f1"].Assets["a1"].CurrentOrder = &domain.MovementOrder{AssetID: "a1", Path: hexes, Destination: domain.Location{WorldID: "Krylos", RegionHex: hexes[2]}}
	me := New(slog.New(slog.DiscardHandler))

	held := domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1", WorldID: "Krylos", NewStepIdx: 1, RegionHex: hexes[1], Reason: domain.HoldRefused}
	if err := me.Apply(s, []domain.Mutation{held}, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	asset := s.Factions["f1"].Assets["a1"]
	if asset.CurrentOrder == nil || asset.CurrentOrder.StepIdx != 1 {
		t.Errorf("order = %+v, want it open at step 1", asset.CurrentOrder)
	}
	if asset.Location != (domain.Location{RegionHex: hexes[1]}) {
		t.Errorf("Location = %+v, want the edge hex with no world", asset.Location)
	}
}

func TestMutationEngine_MovementOrderHeldMisses(t *testing.T) {
	cases := []struct {
		name string
		held domain.MovementOrderHeld
	}{
		{"unknown faction", domain.MovementOrderHeld{FactionID: "missing", AssetID: "a1"}},
		{"unknown asset", domain.MovementOrderHeld{FactionID: "f1", AssetID: "missing"}},
		{"asset with no order", domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(slog.New(slog.DiscardHandler)).Apply(newMutationTestState(), []domain.Mutation{tc.held}, slog.New(slog.DiscardHandler))
			if err == nil || len(err.Misses) != 1 {
				t.Errorf("Apply = %v, want one miss", err)
			}
		})
	}
}

func TestMutationEngine_MovementOrderHeldMarksAPaidLockdown(t *testing.T) {
	for _, tc := range []struct {
		reason domain.HoldReason
		want   bool
	}{
		{domain.HoldLockdownPaid, true},
		{domain.HoldLockdownUnpaid, false},
		{domain.HoldRefused, false},
	} {
		s := newMutationTestState()
		hexes := []spatial.RegionHex{{RegionID: "r", Coord: spatial.HexCoord{Q: 0}}, {RegionID: "r", Coord: spatial.HexCoord{Q: 1}}}
		s.Factions["f1"].Assets["a1"].CurrentOrder = &domain.MovementOrder{AssetID: "a1", Path: hexes}
		held := domain.MovementOrderHeld{FactionID: "f1", AssetID: "a1", NewStepIdx: 0, RegionHex: hexes[0], Reason: tc.reason}

		if err := New(slog.New(slog.DiscardHandler)).Apply(s, []domain.Mutation{held}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("%s: Apply: %v", tc.reason, err)
		}
		if got := s.Factions["f1"].Assets["a1"].CurrentOrder.LockdownPaid; got != tc.want {
			t.Errorf("%s: LockdownPaid = %v, want %v", tc.reason, got, tc.want)
		}
	}
}

func TestMutationEngine_WorldTechLevelRaised(t *testing.T) {
	s := newMutationTestState()
	me := New(slog.New(slog.DiscardHandler))
	apply := func(level int) {
		t.Helper()
		if err := me.Apply(s, []domain.Mutation{
			domain.WorldTechLevelRaised{FactionID: "f1", WorldID: "Backwater", TechLevel: level},
		}, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatalf("Apply() unexpected error: %v", err)
		}
	}
	apply(4)
	if got := s.Factions["f1"].WorldTechLevels["Backwater"]; got != 4 {
		t.Errorf("WorldTechLevels[Backwater] = %d, want 4", got)
	}
	apply(3)
	if got := s.Factions["f1"].WorldTechLevels["Backwater"]; got != 4 {
		t.Errorf("WorldTechLevels[Backwater] after a lower value = %d, want 4 kept", got)
	}
}
