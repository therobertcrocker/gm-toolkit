package execution

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// stubAction satisfies action.Action with only Name() returning a real value.
type stubAction struct{ name string }

func (s stubAction) Name() string                                                                 { return s.name }
func (s stubAction) Validate(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) bool { return true }
func (s stubAction) Inputs(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error  { return nil }
func (s stubAction) Resolve(_ *domain.Faction, _ *state.FactionState, _ *rulebook.Rulebook) error { return nil }
func (s stubAction) Output() ([]domain.Mutation, error)                                           { return nil, nil }

var _ action.Action = stubAction{}

func TestFormatMutation_Surfaced(t *testing.T) {
	cases := []struct {
		name string
		mut  domain.Mutation
		want string
	}{
		{"FactionHPDelta positive", domain.FactionHPDelta{Delta: 3}, "HP +3"},
		{"FactionHPDelta negative", domain.FactionHPDelta{Delta: -2}, "HP -2"},
		{"AssetHPDelta", domain.AssetHPDelta{AssetID: "rifle-1", Delta: -1}, "asset rifle-1 HP -1"},
		{"BaseHPDelta", domain.BaseHPDelta{BaseID: "b1", Delta: 2}, "base b1 HP +2"},
		{"BaseHealed", domain.BaseHealed{BaseID: "b1", Delta: 4}, "base b1 healed +4"},
		{"BaseExpanded", domain.BaseExpanded{BaseID: "b1", Delta: 1}, "base b1 expanded +1"},
		{"CoinDelta positive", domain.CoinDelta{Delta: 10}, "Coin +10"},
		{"CoinDelta negative", domain.CoinDelta{Delta: -5}, "Coin -5"},
		{"AssetAdded", domain.AssetAdded{Asset: domain.Asset{DefinitionID: "def-rifle"}}, "+asset def-rifle"},
		{"AssetRemoved", domain.AssetRemoved{AssetID: "a2"}, "-asset a2"},
		{"BaseAdded", domain.BaseAdded{Base: domain.Base{ID: "base-1"}}, "+base base-1"},
		{"BaseDestroyed", domain.BaseDestroyed{BaseID: "base-1"}, "base base-1 destroyed"},
		{"StatRaised", domain.StatRaised{Stat: domain.StatForce, OldRating: 2, NewRating: 3}, "Force 2→3"},
		{"GoalCompleted", domain.GoalCompleted{XPAwarded: 5}, "goal completed (+5 XP)"},
		{"GoalAbandoned", domain.GoalAbandoned{}, "goal abandoned"},
		{"HomeworldChanged", domain.HomeworldChanged{ToWorld: domain.Location{WorldID: "Andama"}}, "homeworld → Andama"},
		{"MovementOrderIssued", domain.MovementOrderIssued{}, "move issued"},
		{"MovementOrderRevised", domain.MovementOrderRevised{}, "move revised"},
		{"MovementOrderCancelled", domain.MovementOrderCancelled{}, "move cancelled"},
		{"MovementOrderCompleted", domain.MovementOrderCompleted{}, "move completed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, include := formatMutation(tc.mut)
			if !include {
				t.Fatalf("formatMutation(%T): include=false, want true", tc.mut)
			}
			if got != tc.want {
				t.Errorf("formatMutation(%T): got %q, want %q", tc.mut, got, tc.want)
			}
		})
	}
}

func TestFormatMutation_Folded(t *testing.T) {
	folded := []domain.Mutation{
		domain.GoalTurnsTick{},
		domain.GoalProgressed{},
		domain.AssetMaintainedFlag{},
		domain.XPAwarded{},
		domain.MovementOrderProgressed{},
		domain.InfluenceDelta{},
		domain.AssetMoved{},
		domain.TagAdded{},
	}
	for _, mut := range folded {
		t.Run(fmt.Sprintf("%T", mut), func(t *testing.T) {
			got, include := formatMutation(mut)
			if include {
				t.Errorf("formatMutation(%T): include=true, want false; rendered %q", mut, got)
			}
			if got != "" {
				t.Errorf("formatMutation(%T): got %q, want empty string", mut, got)
			}
		})
	}
}

// contains checks that each of the given substrings appears in s.
func contains(t *testing.T, s string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			t.Errorf("output missing %q\nfull output: %q", sub, s)
		}
	}
}

func TestRenderEvent_CycleStarted(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtCycleStarted,
		Payload: adapter.CycleStartedPayload{CycleNumber: 3},
	})
	contains(t, out, "Cycle 3 started")
}

func TestRenderEvent_CycleCompleted(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtCycleCompleted,
		Payload: adapter.CycleCompletedPayload{CycleNumber: 3},
	})
	contains(t, out, "Cycle 3 completed")
}

func TestRenderEvent_FactionEvents(t *testing.T) {
	faction := &domain.Faction{ID: "f1", Name: "Red Star"}
	cases := []struct {
		kind   adapter.EventKind
		phrase string
	}{
		{adapter.EvtFactionTurnStarted, "turn started"},
		{adapter.EvtFactionSkipped, "skipped"},
		{adapter.EvtFactionTurnCompleted, "turn completed"},
		{adapter.EvtStatRaiseSkipped, "stat raise skipped"},
	}
	for _, tc := range cases {
		t.Run(tc.phrase, func(t *testing.T) {
			out := renderEvent(adapter.ObserverEventMsg{Kind: tc.kind, Payload: faction})
			contains(t, out, "Red Star", tc.phrase)
		})
	}
}

func TestRenderEvent_GoalLockApplied_IncludesMutations(t *testing.T) {
	muts := []domain.Mutation{domain.CoinDelta{Delta: -2}}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtGoalLockApplied,
		Payload: adapter.GoalLockAppliedPayload{Mutations: muts},
	})
	contains(t, out, "goal-lock applied", "Coin -2")
}

func TestRenderEvent_BookkeepingApplied_IncludesMutations(t *testing.T) {
	muts := []domain.Mutation{domain.FactionHPDelta{Delta: -1}, domain.CoinDelta{Delta: 3}}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtBookkeepingApplied,
		Payload: adapter.BookkeepingAppliedPayload{Mutations: muts},
	})
	contains(t, out, "bookkeeping applied", "HP -1", "Coin +3")
}

func TestRenderEvent_StatRaiseApplied(t *testing.T) {
	muts := []domain.Mutation{domain.StatRaised{Stat: domain.StatWealth, OldRating: 1, NewRating: 2}}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtStatRaiseApplied,
		Payload: adapter.StatRaiseAppliedPayload{Mutations: muts},
	})
	contains(t, out, "stat raise applied", "Wealth 1→2")
}

func TestRenderEvent_MovementTicked(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtMovementTicked,
		Payload: adapter.MovementTickedPayload{},
	})
	contains(t, out, "movement ticked")
}

func TestRenderEvent_MovementResolved_IncludesMutations(t *testing.T) {
	muts := []domain.Mutation{domain.MovementOrderCompleted{}}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtMovementResolved,
		Payload: adapter.MovementResolvedPayload{Mutations: muts},
	})
	contains(t, out, "movement resolved", "move completed")
}

func TestRenderEvent_ActionSelected_Nil(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtActionSelected,
		Payload: adapter.ActionSelectedPayload{Selected: nil},
	})
	contains(t, out, "action selected: (none)")
}

func TestRenderEvent_ActionSelected_Named(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtActionSelected,
		Payload: adapter.ActionSelectedPayload{Selected: stubAction{name: "Expand Influence"}},
	})
	contains(t, out, "action selected: Expand Influence")
}

func TestRenderEvent_ActionResolved_Nil_IncludesMutations(t *testing.T) {
	muts := []domain.Mutation{domain.AssetAdded{Asset: domain.Asset{DefinitionID: "unit-marines"}}}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtActionResolved,
		Payload: adapter.ActionResolvedPayload{Selected: nil, Mutations: muts},
	})
	contains(t, out, "action resolved: (none)", "+asset unit-marines")
}

func TestRenderEvent_ActionResolved_Named(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtActionResolved,
		Payload: adapter.ActionResolvedPayload{Selected: stubAction{name: "Attack"}},
	})
	contains(t, out, "action resolved: Attack")
}

func TestRenderEvent_IndexSkipped(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtIndexSkipped,
		Payload: adapter.IndexSkippedPayload{Skipped: []string{"f1", "f2"}},
	})
	contains(t, out, "index skipped: f1, f2")
}

func TestRenderEvent_Error(t *testing.T) {
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtError,
		Payload: adapter.ErrorPayload{Err: errors.New("engine panic")},
	})
	contains(t, out, "error: engine panic")
}

func TestRenderEvent_FoldedMutationsOmitted(t *testing.T) {
	muts := []domain.Mutation{
		domain.GoalTurnsTick{},
		domain.CoinDelta{Delta: 5},
		domain.GoalProgressed{},
	}
	out := renderEvent(adapter.ObserverEventMsg{
		Kind:    adapter.EvtBookkeepingApplied,
		Payload: adapter.BookkeepingAppliedPayload{Mutations: muts},
	})
	contains(t, out, "bookkeeping applied", "Coin +5")
	// GoalTurnsTick and GoalProgressed fold — their types should not appear as text
	if strings.Contains(out, "GoalTurnsTick") || strings.Contains(out, "GoalProgressed") {
		t.Error("folded mutation type names leaked into rendered output")
	}
}
