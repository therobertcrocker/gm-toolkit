package turn

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/execution"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/setup"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type view int

const (
	viewSetup view = iota
	viewExecution
)

type Model struct {
	factionState *state.FactionState
	paths        *campaigns.Paths
	eng          *engine.Engine
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap
	log          *slog.Logger

	adapter   *adapter.Adapter // nil until Start; built per cycle, discarded on return to setup
	view      view
	setup     setup.Model
	execution execution.Model
	doneErr   error // stashed on EngineDoneMsg; consumed at StreamClosedMsg teardown

	termWidth, termHeight int
}

func New(factionState *state.FactionState, paths *campaigns.Paths, eng *engine.Engine, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, log *slog.Logger) Model {
	return Model{
		factionState: factionState,
		paths:        paths,
		eng:          eng,
		rulebook:     rb,
		spatialMap:   spatialMap,
		log:          log,
		view:         viewSetup,
		setup:        setup.New(factionState),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m.routeForward(msg)

	case msgs.StartCycleMsg:
		m.adapter = adapter.New(m.eng, m.factionState, m.paths, m.log)
		m.execution = execution.New(m.termWidth, m.termHeight, m.rulebook, m.spatialMap)
		m.view = viewExecution
		return m, tea.Batch(m.adapter.Run(), m.adapter.ObserverPump(), m.adapter.AskPump())

	case adapter.ObserverEventMsg:
		m.execution, _ = m.execution.Update(msg)
		return m, m.adapter.ObserverPump() // re-issue to keep the pump alive

	case adapter.CollectorAskMsg:
		var execCmd tea.Cmd
		m.execution, execCmd = m.execution.Update(msg)
		return m, tea.Batch(execCmd, m.adapter.AskPump()) // execCmd carries overlay.Init()

	case adapter.EngineDoneMsg:
		// Stash the result and surface a fatal error in place, but do NOT tear
		// down yet — events buffered in eventCh may still be in flight. Teardown
		// waits for StreamClosedMsg, which the pump emits only after eventCh is
		// drained and closed.
		m.doneErr = msg.Err
		if msg.Err != nil {
			m.execution, _ = m.execution.Update(msg)
		}
		return m, nil

	case adapter.StreamClosedMsg:
		if m.doneErr != nil {
			return m, nil // fatal: stay in the execution view; a keypress returns to setup
		}
		return m.returnToSetup()

	case msgs.ReturnToSetupMsg:
		return m.returnToSetup()
	}

	return m.routeForward(msg)
}

func (m Model) routeForward(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case viewSetup:
		m.setup, cmd = m.setup.Update(msg)
	case viewExecution:
		m.execution, cmd = m.execution.Update(msg)
	}
	return m, cmd
}

// returnToSetup discards the finished cycle and rebuilds the roster with
// post-cycle HP/Coin. Shared by the clean-completion (StreamClosedMsg) and
// fatal-recovery (ReturnToSetupMsg) paths.
func (m Model) returnToSetup() (tea.Model, tea.Cmd) {
	m.view = viewSetup
	m.adapter = nil
	m.doneErr = nil
	m.setup = setup.New(m.factionState)
	if m.termWidth > 0 {
		m.setup, _ = m.setup.Update(tea.WindowSizeMsg{Width: m.termWidth, Height: m.termHeight})
	}
	return m, nil
}

func (m Model) View() string {
	if m.view == viewExecution {
		return m.execution.View()
	}
	return m.setup.View()
}

func (m Model) Help() help.KeyMap {
	if m.view == viewExecution {
		return m.execution.Help()
	}
	return m.setup.Help()
}

func (m Model) CapturesInput() bool {
	return m.view == viewExecution && m.execution.CapturesInput()
}

// ModeLocked blocks Tab/Shift-Tab while a cycle is running: the event/ask pumps
// re-arm only from this router's Update, so switching to another mode would
// orphan them and wedge the engine on the unbuffered askCh send. (Robust
// mid-cycle pause/cancel is deferred — for now a cycle runs to completion.)
func (m Model) ModeLocked() bool { return m.view == viewExecution }

func (m Model) StatusLine() (string, chrome.Severity) {
	if m.view == viewExecution {
		return m.execution.StatusLine()
	}
	return "", chrome.Info
}
