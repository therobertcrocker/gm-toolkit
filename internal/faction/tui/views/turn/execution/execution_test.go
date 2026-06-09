package execution

import (
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
)

func TestModel_RailAdvancesOnFactionTurnStarted(t *testing.T) {
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{
		Kind: adapter.EvtCycleStarted,
		Payload: adapter.CycleStartedPayload{
			CycleNumber: 1,
			Order: []adapter.RailEntry{
				{ID: "f1", Name: "Alpha"},
				{ID: "f2", Name: "Beta"},
			},
		},
	})
	m, _ = m.Update(adapter.ObserverEventMsg{
		Kind:    adapter.EvtFactionTurnStarted,
		Payload: &domain.Faction{ID: "f1", Name: "Alpha"},
	})
	view := m.View()
	if !strings.Contains(view, "▶") {
		t.Error("View() missing ▶ marker after EvtFactionTurnStarted")
	}
	if !strings.Contains(view, "Alpha") {
		t.Error("View() missing Alpha in rail")
	}
	if !strings.Contains(view, "Beta") {
		t.Error("View() missing Beta in rail")
	}
}

func TestModel_DetailCard_StatRaisePhase(t *testing.T) {
	faction := &domain.Faction{
		ID: "f1", Name: "Alliance", Scale: domain.ScaleMajor,
		Force: 3, Cunning: 2, Wealth: 1,
		CurrentHP: 14, MaxHP: 20, Coin: 5,
	}
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{Kind: adapter.EvtFactionTurnStarted, Payload: faction})
	// The specialized stat-raise card shows only while its modal is parked.
	m, _ = m.Update(adapter.CollectorAskMsg{
		Kind:    adapter.AskSelectStatRaise,
		Faction: faction,
		Payload: adapter.SelectStatRaisePayload{Eligible: []domain.FactionStat{domain.StatForce, domain.StatCunning, domain.StatWealth}},
	})
	view := m.View()
	contains(t, view, "Alliance", "Raise a stat", "Force")
}

// TestModel_DetailCard_NoFlashWithoutOverlay locks in the anti-flash gating:
// FactionTurnStarted sets phase=phaseStatRaise speculatively, but with no modal
// parked the detail card must stay on the steady base layout, not the stat-raise
// card that would flicker past during the event burst.
func TestModel_DetailCard_NoFlashWithoutOverlay(t *testing.T) {
	faction := &domain.Faction{
		ID: "f1", Name: "Alliance", Scale: domain.ScaleMajor,
		Force: 3, Cunning: 2, Wealth: 1,
		CurrentHP: 14, MaxHP: 20, Coin: 5,
	}
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{Kind: adapter.EvtFactionTurnStarted, Payload: faction})
	if got := m.detailView(); strings.Contains(got, "Raise a stat") {
		t.Errorf("detailView() showed stat-raise card with no overlay mounted:\n%s", got)
	}
	contains(t, m.detailView(), "Alliance", "14/20", "Coin")
}

func TestModel_DetailCard_BasePhase(t *testing.T) {
	faction := &domain.Faction{
		ID: "f1", Name: "Alliance", Scale: domain.ScaleMajor,
		Force: 3, Cunning: 2, Wealth: 1,
		CurrentHP: 14, MaxHP: 20, Coin: 5,
	}
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{Kind: adapter.EvtFactionTurnStarted, Payload: faction})
	m, _ = m.Update(adapter.ObserverEventMsg{Kind: adapter.EvtFactionTurnCompleted, Payload: faction})
	view := m.View()
	contains(t, view, "Alliance", "14/20", "Coin")
}

// TestModel_DetailCard_MovementSurvivesPriorTick locks in the activeAsk-driven
// card: the engine fires MovementTicked (advancing in-flight orders) *before* the
// movement ask, so a card keyed off turn events would have reset past movement by
// the time the modal opens. Keyed off the mounted modal, the asset list still shows.
func TestModel_DetailCard_MovementSurvivesPriorTick(t *testing.T) {
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{
		Kind:    adapter.EvtFactionTurnStarted,
		Payload: &domain.Faction{ID: "f1", Name: "Alliance"},
	})
	m, _ = m.Update(adapter.ObserverEventMsg{Kind: adapter.EvtMovementTicked, Payload: adapter.MovementTickedPayload{}})

	// Mount the movement modal. State is set directly (rather than via a
	// CollectorAskMsg) to keep resolveMovables — which needs a real spatial map —
	// out of this unit. overlay value is an arbitrary non-nil stand-in.
	m.overlay = overlay.NewCheckpoint(1)
	m.activeAsk = adapter.AskSelectMovementDecisions
	m.movables = []movableLine{{name: "Frigate", location: "Anvil", order: "→ Bastion"}}

	contains(t, m.detailView(), "Movable assets", "Frigate", "Anvil")
}

func TestModel_StreamReceivesEvents(t *testing.T) {
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.ObserverEventMsg{
		Kind:    adapter.EvtCycleStarted,
		Payload: adapter.CycleStartedPayload{CycleNumber: 1},
	})
	view := m.View()
	if !strings.Contains(view, "Cycle 1 started") {
		t.Error("View() missing cycle-started event in stream")
	}
}

func TestModel_StatusLine_NoError(t *testing.T) {
	m := New(80, 24, nil, nil)
	text, _ := m.StatusLine()
	if text != "" {
		t.Errorf("StatusLine() = %q, want empty before any error", text)
	}
}

func TestModel_StatusLine_FatalOnEngineDoneErr(t *testing.T) {
	m := New(80, 24, nil, nil)
	m, _ = m.Update(adapter.EngineDoneMsg{Err: errForTest("cycle failed")})
	text, sev := m.StatusLine()
	if !strings.Contains(text, "cycle failed") {
		t.Errorf("StatusLine() text = %q, want to contain %q", text, "cycle failed")
	}
	if !m.fatal {
		t.Error("fatal should be true after EngineDoneMsg with Err")
	}
	_ = sev
}

func TestModel_CapturesInputFalseWithoutOverlay(t *testing.T) {
	m := New(80, 24, nil, nil)
	if m.CapturesInput() {
		t.Error("CapturesInput() = true, want false when no overlay mounted")
	}
}

// errForTest is a minimal error for test fixtures.
type errForTest string

func (e errForTest) Error() string { return string(e) }
