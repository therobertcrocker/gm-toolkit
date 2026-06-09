# TUI Turn — Code-Review Fixes Plan

Post-Effort-1 review remediation. Effort 1 of `F-005.3` landed via [`tui-turn-effort-1-plan.md`](./tui-turn-effort-1-plan.md); a senior-engineer review of `e3a4538..HEAD` surfaced seven items. This plan resolves all seven across three commits.

## Context / Goal

Effort 1 shipped the fully-wired turn workflow (router, three-pane execution view, ask-pump round-trip, placeholder overlays). The review found two correctness defects in the cycle lifecycle, one UX dead-end, and four code-quality nits. None blocks the seam Efforts 2–3 build on; this plan hardens the lifecycle and clears the nits before merge.

- Effort 1 plan: [`tui-turn-effort-1-plan.md`](./tui-turn-effort-1-plan.md) — the seam this builds on (Decision 1 channel-close, Decision 3 pumps, Decision 11 overlay).
- The seven items, by severity:

| # | Severity | Item | Commit |
|---|----------|------|--------|
| 1 | High     | Mid-cycle `Tab` orphans the event/ask pumps → engine wedges on the unbuffered `askCh` send | 1 |
| 2 | Medium   | Tail events (e.g. "Cycle N completed") can be silently dropped — `EngineDoneMsg` races the buffered events | 2 |
| 3 | Medium   | Fatal-error path strands the user in a dead execution view with no return-to-setup affordance | 2 |
| 7 | Low      | `EvtError` sets `latestErr` but not `fatal`; a recoverable error lingers for the rest of the cycle | 2 |
| 4 | Low      | `setup.buildItems` uses single-letter `f` — violates the full-word code-style rule | 3 |
| 5 | Low      | `execution.Model.cycleNumber` is write-only dead state | 3 |
| 6 | Low      | `SelectStatRaise` relies on an undocumented typed-nil contract with no guard or test | 3 |

## Decisions Ratified in Planning

1. **Item 1 — lock mode-switch only, keep `q`/`?` live (Option A).** The root suppresses globals via the all-or-nothing `inputCapturer`; rather than capture the whole keyboard during a cycle (which would also disable quit and help), add a narrow `modeLocker` predicate gating only `Tab`/`Shift-Tab`. While a cycle runs, mode-switching is a no-op; quit and help still work. (Chosen over full `CapturesInput`-during-execution: that locks `q`/`?` too. Robust mid-cycle pause/cancel — the deeper feature that would make a mid-cycle mode-switch *safe* rather than merely blocked — is deferred future work; this is the minimal correctness fix.)
2. **Item 2 — teardown is driven by the event stream closing, not by `EngineDoneMsg`.** `ObserverPump` emits a `StreamClosedMsg` sentinel when `eventCh` is drained and closed (replacing the swallowed `nil`). The router stashes the result on `EngineDoneMsg` but only resets the view on `StreamClosedMsg` — which the pump emits strictly *after* every buffered event. This revises Effort-1 Decision 1's reader-side note: the `adapter == nil` re-issue guards are **removed**, because draining-to-close orders the events ahead of teardown by construction, designing the race out instead of patching the panic. (`close(eventCh)` runs as `Run` returns `EngineDoneMsg`, so the close is visible to the pump only once all ≤64 buffered events are delivered; `EngineDoneMsg` is one enqueue and reliably precedes the pump's final `StreamClosedMsg`, so the stashed error is always present at teardown.)
3. **Item 3 — fatal recovery is a manual keypress, not an auto-return.** On `EngineDoneMsg{Err != nil}` the execution view stays mounted (the error is shown in the stream + status line); any keypress emits `ReturnToSetupMsg`, which the router handles identically to a clean completion. Manual (not automatic) so the GM can read the error before the roster replaces it. The happy path keeps its automatic return (now via `StreamClosedMsg`).

## Out of Scope

- **Robust mid-cycle pause / cancel (`F-017`).** Decision 1 blocks `Tab` mid-cycle; it does not add a clean way to *pause or abort* a running cycle. That (and `Esc`-cancel / `ErrTurnCanceled`, already Effort-3 work) stays deferred — tracked as backlog `F-017`.
- Already-logged backlog: per-frame allocations in `Compose`/`RegionWidths` (`R-011`), engine-construction shape (`R-012`), event-stream flash (`B-006`).
- The `setup`/`manage` shared read-only list refactor (Effort-1 Decision 8 deferral) and the over-dimmed-panes overlay composite (Effort 2).

## Shared Context

**Model discipline.** We are in Plan (Opus). Commits 1 and 3 are mechanical — Sonnet. Commit 2 reshapes the cycle-teardown ordering and removes the race guards — **suggest Opus**. Prompt to switch at the Plan→Execution transition.

---

## Work Breakdown

Three commits, one execution session each. Commit 1 is independent. Commit 2 depends on nothing but is the subtle one. Commit 3 is independent nits.

### Commit 1 — `fix(tui): lock mode-switch while the active sub is busy`

Resolves item 1. Adds a `modeLocker` predicate parallel to `inputCapturer` and has the turn router report locked while a cycle runs.

#### Task 1 — `internal/faction/tui/update.go`

Add the `modeLocker` interface immediately after `inputCapturer`.

**Find:**
```go
// inputCapturer is implemented by sub-models that own the full keyboard while
// active (e.g. a text-entry form). When the active sub captures input, the root
// forwards every key to it and suppresses the global bindings below.
type inputCapturer interface {
	CapturesInput() bool
}
```
**Replace with:**
```go
// inputCapturer is implemented by sub-models that own the full keyboard while
// active (e.g. a text-entry form). When the active sub captures input, the root
// forwards every key to it and suppresses the global bindings below.
type inputCapturer interface {
	CapturesInput() bool
}

// modeLocker is implemented by sub-models that forbid switching modes while
// busy (e.g. the turn view while a cycle is running). Tab/Shift-Tab are no-ops
// while the active sub reports locked, so a running cycle's event/ask pumps —
// which re-arm only from that sub's Update — are never orphaned by a switch to
// another mode. Unlike inputCapturer this gates only mode-switch; q and ?
// stay live.
type modeLocker interface {
	ModeLocked() bool
}
```

Gate the `tab`/`shift+tab` global bindings.

**Find:**
```go
		switch msg.String() {
		case "tab":
			m.bar = m.bar.Next()
			return m.afterModeChange()
		case "shift+tab":
			m.bar = m.bar.Prev()
			return m.afterModeChange()
```
**Replace with:**
```go
		switch msg.String() {
		case "tab":
			if m.modeSwitchLocked() {
				return m, nil
			}
			m.bar = m.bar.Next()
			return m.afterModeChange()
		case "shift+tab":
			if m.modeSwitchLocked() {
				return m, nil
			}
			m.bar = m.bar.Prev()
			return m.afterModeChange()
```

Add the helper after `afterModeChange`.

**Find:**
```go
func (m Model) afterModeChange() (Model, tea.Cmd) {
	return m, nil
}
```
**Replace with:**
```go
func (m Model) afterModeChange() (Model, tea.Cmd) {
	return m, nil
}

// modeSwitchLocked reports whether the active sub-model forbids switching modes
// right now (e.g. turn while a cycle is running).
func (m Model) modeSwitchLocked() bool {
	if locker, ok := m.subs[m.bar.Active()].(modeLocker); ok {
		return locker.ModeLocked()
	}
	return false
}
```

#### Task 2 — `internal/faction/tui/views/turn/turn.go`

Implement `ModeLocked`. Add it immediately after `CapturesInput`.

**Find:**
```go
func (m Model) CapturesInput() bool {
	return m.view == viewExecution && m.execution.CapturesInput()
}
```
**Replace with:**
```go
func (m Model) CapturesInput() bool {
	return m.view == viewExecution && m.execution.CapturesInput()
}

// ModeLocked blocks Tab/Shift-Tab while a cycle is running: the event/ask pumps
// re-arm only from this router's Update, so switching to another mode would
// orphan them and wedge the engine on the unbuffered askCh send. (Robust
// mid-cycle pause/cancel is deferred — for now a cycle runs to completion.)
func (m Model) ModeLocked() bool { return m.view == viewExecution }
```

#### Task 3 — verify
`go build ./...` && `go test ./internal/faction/tui/...`. Manually: Start a cycle, press `Tab`/`Shift-Tab` mid-stream → nothing happens (mode bar stays on Turn). Confirm `q` still opens the quit confirm and `?` still toggles compact↔full help while a cycle runs. After the cycle returns to setup, `Tab` switches modes again.

#### Commit message
```
fix(tui): lock mode-switch while the active sub is busy

- add modeLocker interface (parallel to inputCapturer); Tab/Shift-Tab are
  no-ops when the active sub reports ModeLocked, q and ? stay live
- turn.Model.ModeLocked() is true while a cycle runs (viewExecution), so a
  mode switch can no longer orphan the event/ask pumps and wedge the engine
  on the unbuffered askCh send
```

### Commit 2 — `fix(tui/turn): drain event stream before teardown; recover from fatal cycle`
*(suggest Opus)*

Resolves items 2, 3, 7. Flips teardown to the stream-closed signal so tail events render, adds a manual return from the fatal state, and bounds a recoverable error's lifetime.

#### Task 1 — `internal/faction/tui/adapter/channels.go`

Add the `StreamClosedMsg` sentinel after `CollectorAskMsg`.

**Find:**
```go
type CollectorAskMsg struct {
	Kind    AskKind
	Faction *domain.Faction
	Payload any
	Reply   chan<- any
}
```
**Replace with:**
```go
type CollectorAskMsg struct {
	Kind    AskKind
	Faction *domain.Faction
	Payload any
	Reply   chan<- any
}

// StreamClosedMsg is emitted by ObserverPump once eventCh is drained and closed
// — i.e. RunCycle has returned and every buffered event has been delivered. The
// router uses this (not EngineDoneMsg) as the teardown signal, so the final
// cycle-completed events always render before the view resets.
type StreamClosedMsg struct{}
```

#### Task 2 — `internal/faction/tui/adapter/run.go`

`ObserverPump` returns the sentinel instead of swallowing the closed channel as `nil`. `AskPump` is unchanged (asks are done at cycle end; its close has no teardown role).

**Find:**
```go
// ObserverPump reads one event from eventCh and returns it as a tea.Msg.
// The router re-issues this command after each receipt to keep the pump alive.
func (a *Adapter) ObserverPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.eventCh
		if !ok {
			return nil
		}
		return msg
	}
}
```
**Replace with:**
```go
// ObserverPump reads one event from eventCh and returns it as a tea.Msg. The
// router re-issues this command after each receipt to keep the pump alive; on a
// drained, closed eventCh it returns StreamClosedMsg, the router's teardown
// signal.
func (a *Adapter) ObserverPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.eventCh
		if !ok {
			return StreamClosedMsg{}
		}
		return msg
	}
}
```

#### Task 3 — `internal/faction/tui/views/turn/turn.go`

Three edits: stash the result field, rework the lifecycle cases (and drop the race guards), and extract the shared teardown.

Add the `doneErr` field.

**Find:**
```go
	adapter   *adapter.Adapter // nil until Start; built per cycle, discarded on return to setup
	view      view
	setup     setup.Model
	execution execution.Model

	termWidth, termHeight int
}
```
**Replace with:**
```go
	adapter   *adapter.Adapter // nil until Start; built per cycle, discarded on return to setup
	view      view
	setup     setup.Model
	execution execution.Model
	doneErr   error // stashed on EngineDoneMsg; consumed at StreamClosedMsg teardown

	termWidth, termHeight int
}
```

Rework the three lifecycle cases. `EngineDoneMsg` now only stashes; `StreamClosedMsg` tears down on success; `ReturnToSetupMsg` handles the manual fatal return (item 3). The `adapter == nil` guards are removed — `StreamClosedMsg` is strictly the last pump output, so no `ObserverEventMsg`/`CollectorAskMsg` can arrive after teardown.

**Find:**
```go
	case adapter.ObserverEventMsg:
		m.execution, _ = m.execution.Update(msg)
		if m.adapter == nil {
			return m, nil // a buffered event raced EngineDoneMsg; cycle already torn down
		}
		return m, m.adapter.ObserverPump() // re-issue to keep the pump alive

	case adapter.CollectorAskMsg:
		m.execution, _ = m.execution.Update(msg)
		if m.adapter == nil {
			return m, nil
		}
		return m, m.adapter.AskPump() // re-issue to keep the pump alive

	case adapter.EngineDoneMsg:
		if msg.Err != nil {
			m.execution, _ = m.execution.Update(msg) // surface the fatal error in place
			return m, nil
		}
		m.view = viewSetup
		m.adapter = nil
		m.setup = setup.New(m.factionState)
		if m.termWidth > 0 {
			m.setup, _ = m.setup.Update(tea.WindowSizeMsg{Width: m.termWidth, Height: m.termHeight})
		}
		return m, nil
	}
```
**Replace with:**
```go
	case adapter.ObserverEventMsg:
		m.execution, _ = m.execution.Update(msg)
		return m, m.adapter.ObserverPump() // re-issue to keep the pump alive

	case adapter.CollectorAskMsg:
		m.execution, _ = m.execution.Update(msg)
		return m, m.adapter.AskPump() // re-issue to keep the pump alive

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
```

Add the shared teardown helper after `routeForward`.

**Find:**
```go
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
```
**Replace with:**
```go
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
```

#### Task 4 — `internal/faction/tui/views/turn/msgs/msgs.go`

Add `ReturnToSetupMsg`.

**Find:**
```go
// StartCycleMsg requests the turn router to build a fresh adapter and begin a
// cycle. Emitted by the setup view on the Start key; handled by turn.Model.
type StartCycleMsg struct{}
```
**Replace with:**
```go
// StartCycleMsg requests the turn router to build a fresh adapter and begin a
// cycle. Emitted by the setup view on the Start key; handled by turn.Model.
type StartCycleMsg struct{}

// ReturnToSetupMsg requests the router to discard a finished cycle and return
// to the setup roster. Emitted by the execution view on a keypress after a
// fatal cycle error — the one terminal state with no automatic return (the
// happy path returns via StreamClosedMsg). Handled by turn.Model.
type ReturnToSetupMsg struct{}
```

#### Task 5 — `internal/faction/tui/views/turn/execution/execution.go`

Two edits: emit `ReturnToSetupMsg` on a keypress while fatal (item 3), and clear a recoverable error at the next faction turn (item 7). Both need the `msgs` import.

Add the `msgs` import (alphabetical, before `overlay`).

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
)
```

In `Update`'s `tea.KeyMsg` case, a key in the fatal state returns to setup before any scroll/overlay handling.

**Find:**
```go
	case tea.KeyMsg:
		if m.overlay != nil {
			var cmd tea.Cmd
			m.overlay, cmd = m.overlay.Update(msg)
			return m, cmd
		}
```
**Replace with:**
```go
	case tea.KeyMsg:
		if m.fatal {
			return m, func() tea.Msg { return msgs.ReturnToSetupMsg{} }
		}
		if m.overlay != nil {
			var cmd tea.Cmd
			m.overlay, cmd = m.overlay.Update(msg)
			return m, cmd
		}
```

In `applyEvent`, clear a stale recoverable error when the next faction's turn begins (item 7). A fatal error persists.

**Find:**
```go
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
```
**Replace with:**
```go
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
		if !m.fatal {
			m.latestErr = nil // a recoverable error scopes to the faction it occurred under
		}
```

#### Task 6 — verify
`go build ./...` && `go test ./internal/faction/tui/views/turn/...`. Then a live cycle, re-run several times:
- The stream ends on "Cycle N completed" every time (item 2) — never mid-stream — then auto-returns to setup with post-cycle HP/Coin.
- Force a fatal error (e.g. run before the world engine is wired) → the execution view holds the error; any key returns to setup (item 3); a fresh Start works.
- A recoverable `EvtError` shows in the status line, then clears when the next faction's turn starts (item 7).

#### Commit message
```
fix(tui/turn): drain event stream before teardown; recover from fatal cycle

- ObserverPump emits StreamClosedMsg on a drained+closed eventCh; the router
  tears down on that signal, not on EngineDoneMsg, so the final
  cycle-completed events always render before the view resets
- EngineDoneMsg now only stashes the result; remove the adapter==nil
  re-issue guards (drain-to-close ordering designs out the race they patched)
- fatal cycle error: execution emits ReturnToSetupMsg on any key, so the GM
  is no longer stranded in a dead execution view; happy path keeps its
  automatic return
- recoverable EvtError clears at the next faction turn instead of lingering
  for the rest of the cycle
```

### Commit 3 — `refactor(tui/turn): review nits — naming, dead field, stat-raise guard`

Resolves items 4, 5, 6. Independent, mechanical.

#### Task 1 — `internal/faction/tui/views/turn/setup/setup.go`

Item 4: rename the loop variable `f` → `faction` (full-word code-style rule).

**Find:**
```go
	items := make([]list.Item, len(factionIDs))
	for i, id := range factionIDs {
		f := factionState.Factions[id]
		items[i] = item{
			name:  f.Name,
			scale: string(f.Scale),
			hp:    fmt.Sprintf("%d/%d", f.CurrentHP, f.MaxHP),
			coin:  fmt.Sprintf("%d", f.Coin),
		}
	}
```
**Replace with:**
```go
	items := make([]list.Item, len(factionIDs))
	for i, id := range factionIDs {
		faction := factionState.Factions[id]
		items[i] = item{
			name:  faction.Name,
			scale: string(faction.Scale),
			hp:    fmt.Sprintf("%d/%d", faction.CurrentHP, faction.MaxHP),
			coin:  fmt.Sprintf("%d", faction.Coin),
		}
	}
```

#### Task 2 — `internal/faction/tui/views/turn/execution/execution.go`

Item 5: drop the write-only `cycleNumber` field; the cycle number already renders via `format.go`'s `eventHeader`.

Remove the field.

**Find:**
```go
type Model struct {
	cycleNumber int
	order       []adapter.RailEntry
	currentID   string
```
**Replace with:**
```go
type Model struct {
	order     []adapter.RailEntry
	currentID string
```

Drop its assignment in `applyEvent` (the `Order` is the only consumed part).

**Find:**
```go
	case adapter.EvtCycleStarted:
		p := msg.Payload.(adapter.CycleStartedPayload)
		m.cycleNumber = p.CycleNumber
		m.order = p.Order
```
**Replace with:**
```go
	case adapter.EvtCycleStarted:
		m.order = msg.Payload.(adapter.CycleStartedPayload).Order
```

#### Task 3 — `internal/faction/tui/adapter/phase_collector.go`

Item 6: guard `SelectStatRaise` against a nil reply, mirroring `SelectAction` — so an untyped-nil "decline" is legal, not just the placeholder's typed `(*domain.FactionStat)(nil)`.

**Find:**
```go
	raw, err := p.ask(AskSelectStatRaise, faction, SelectStatRaisePayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.(*domain.FactionStat)
```
**Replace with:**
```go
	raw, err := p.ask(AskSelectStatRaise, faction, SelectStatRaisePayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil // decline the raise (accept an untyped nil, not only a typed nil)
	}
	chosen, ok := raw.(*domain.FactionStat)
```

#### Task 4 — `internal/faction/tui/views/turn/overlay/placeholder_test.go` (new)

Item 6 safety net: pin each `AskKind`'s `defaultFor` value to the exact dynamic type its phase collector asserts on (`phase_collector.go`). If a default's type drifts from the collector contract, this fails in-package rather than mid-cycle. (A full round-trip through the live collector is heavier — it would need an `Adapter` with channels and an engine — and is deferred; this pins the placeholder side, which is what Efforts 2–3 replace.) Full contents:

```go
package overlay

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
)

// defaultFor's reply for each kind must be the exact dynamic type the phase
// collector asserts on (phase_collector.go), or the cycle errors at the
// assertion. AwaitCheckpoint's reply is discarded (any non-error value is
// legal); SelectAction's skip is an untyped nil.
func TestDefaultForMatchesCollectorAssertions(t *testing.T) {
	if _, ok := defaultFor(adapter.AskAwaitCheckpoint).(bool); !ok {
		t.Errorf("AskAwaitCheckpoint: want bool ack, got %T", defaultFor(adapter.AskAwaitCheckpoint))
	}
	if v := defaultFor(adapter.AskSelectAction); v != nil {
		t.Errorf("AskSelectAction: want untyped nil skip, got %T", v)
	}
	if _, ok := defaultFor(adapter.AskSelectStatRaise).(*domain.FactionStat); !ok {
		t.Errorf("AskSelectStatRaise: want *domain.FactionStat, got %T", defaultFor(adapter.AskSelectStatRaise))
	}
	if _, ok := defaultFor(adapter.AskSelectMovementDecisions).([]world.MovementDecision); !ok {
		t.Errorf("AskSelectMovementDecisions: want []world.MovementDecision, got %T", defaultFor(adapter.AskSelectMovementDecisions))
	}
	if _, ok := defaultFor(adapter.AskSelectTransportCargo).([]*domain.Asset); !ok {
		t.Errorf("AskSelectTransportCargo: want []*domain.Asset, got %T", defaultFor(adapter.AskSelectTransportCargo))
	}
}
```

#### Task 5 — verify
`go build ./...` && `go test ./internal/faction/...`. The new test passes; Manage and the setup roster render identically.

#### Commit message
```
refactor(tui/turn): review nits — naming, dead field, stat-raise guard

- setup.buildItems: rename loop var f -> faction (full-word code style)
- execution.Model: drop the write-only cycleNumber field; the cycle number
  already renders via format.go
- phase_collector.SelectStatRaise: guard raw == nil (decline) so an untyped
  nil reply is legal, matching SelectAction
- pin every AskKind's default to its collector assertion with a defaultFor test
```
