package execution

import (
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/dice/dicetest"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/testharness"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

func sandboxMap(t *testing.T) *spatial.RegionMap {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", "..")
	regionMap, err := spatial.LoadRegionMap(filepath.Join(repoRoot, "campaigns", "sandbox", "spatial"))
	if err != nil {
		t.Fatalf("LoadRegionMap: %v", err)
	}
	return regionMap
}

// answerFor answers each ask as a GM would: alpha attacks, and everyone else passes.
func answerFor(t *testing.T, ask adapter.CollectorAskMsg) any {
	t.Helper()
	switch ask.Kind {
	case adapter.AskAwaitCheckpoint:
		return true
	case adapter.AskSelectStatRaise, adapter.AskSelectGoal:
		return nil
	case adapter.AskSelectMovementDecisions:
		return []world.MovementDecision{}
	case adapter.AskSelectAction:
		if ask.Faction.ID != "alpha" {
			return nil
		}
		for _, candidate := range ask.Payload.(adapter.SelectActionPayload).Available {
			if candidate.Name() == "Attack" {
				return candidate
			}
		}
		t.Fatal("Attack not available for alpha")
	case adapter.AskSelectAttackers:
		return ask.Payload.(adapter.SelectAttackersPayload).Eligible
	case adapter.AskSelectDefender:
		return ask.Payload.(adapter.SelectDefenderPayload).Eligible[0]
	case adapter.AskConfirmRedirectToBase, adapter.AskConfirmCounterstrike:
		return false
	case adapter.AskPreRollCard:
		return []int{}
	case adapter.AskPostRollCard:
		return adapter.KeepRoll
	}
	t.Fatalf("no answer for AskKind %d", ask.Kind)
	return nil
}

// TestCycleThroughTheRealChannels drives one full cycle through the adapter and the
// real view. Its job is to run under -race, where snapshotOf reads each live faction
// pointer the way B.038 suspects. A race found here is logged on B.038, not fixed.
func TestCycleThroughTheRealChannels(t *testing.T) {
	harness := testharness.NewHarness(t)
	harness.AddFaction("alpha", "Tartarus", 4, 3, 2)
	harness.AddFaction("beta", "Tartarus", 2, 2, 2)
	harness.Engine.Rand = &dicetest.Cycle{Values: []int{7, 3}}
	bridge := adapter.New(harness.Engine, harness.FactionState, harness.Paths, slog.New(slog.DiscardHandler))
	model := New(120, 40, harness.Engine.Rulebook, sandboxMap(t))

	messages := make(chan tea.Msg)
	go func() { messages <- bridge.Run()() }()
	pump := func(next tea.Cmd) {
		for {
			msg := next()
			if msg == nil {
				return
			}
			messages <- msg
			if _, closed := msg.(adapter.StreamClosedMsg); closed {
				return
			}
		}
	}
	go pump(bridge.ObserverPump())
	go pump(bridge.AskPump())

	engineDone, streamClosed := false, false
	timeout := time.After(30 * time.Second)
	for !engineDone || !streamClosed {
		select {
		case msg := <-messages:
			switch typed := msg.(type) {
			case adapter.EngineDoneMsg:
				engineDone = true
				if typed.Err != nil {
					t.Fatalf("RunCycle: %v", typed.Err)
				}
			case adapter.StreamClosedMsg:
				streamClosed = true
			case adapter.CollectorAskMsg:
				model, _ = model.Update(typed)
				model, _ = model.Update(overlay.OverlayDoneMsg{Answer: answerFor(t, typed)})
			default:
				model, _ = model.Update(msg)
			}
		case <-timeout:
			t.Fatal("the cycle did not finish in 30 seconds")
		}
	}
	contains(t, strings.Join(model.lines, "\n"), " rolls ", "completed")
}
