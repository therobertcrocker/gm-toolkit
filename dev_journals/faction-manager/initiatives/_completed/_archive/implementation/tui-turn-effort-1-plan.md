# TUI Turn — Effort 1 Plan: Turn-flow scaffold

`F-005.3`, Effort 1 of 3. Per-effort detail under the top-level [`tui-turn-plan.md`](./tui-turn-plan.md); design authority is [`tui-turn-discovery.md`](../discovery/tui-turn-discovery.md). This file resolves the seven Effort-1 open questions the overview routed here (OQ 1–7) and breaks Effort 1 into seven commits, each one execution session.

- Overview: [`tui-turn-plan.md`](./tui-turn-plan.md) — branch/merge strategy, the ratified `Overlay` seam, model discipline.
- Discovery: [`tui-turn-discovery.md`](../discovery/tui-turn-discovery.md) — *Setup view*, *Execution view*, *Event stream rendering*, *Contextual bottom bar*, *Modal overlay infrastructure*, *Logging wiring*, Decisions 8–12.

## Context / Goal

Effort 1 is the **structural** Effort: it lands a fully-wired Turn workflow that drives the **real** engine end-to-end (`Adapter.Run()` → `RunCycle` on a goroutine), with every collector prompt answered by a generic placeholder overlay that emits the legal "no decision" default on a keypress. No real modals, but every seam in place — router, three-pane layout, the symmetric **ask-pump**, the dispatch switch over all five `AskKind`s, the event-stream renderer over all 14 observer events plus the `Mutation → string` formatter, and the two app-wide pieces (campaign logging, contextual bottom bar). Efforts 2 and 3 then swap placeholders for real overlays without touching this plumbing.

**Outcome:** launch Turn → press Start → keypress through each prompt → watch a real cycle stream events to completion, every decision auto-defaulted, then return to setup.

## Decisions Ratified in Planning

1. **Adapter ownership: per-cycle, built by the turn router; sole-writer channel close (resolves OQ 1 ownership + OQ 4).** The engine is threaded into `turn.New`; the router builds a **fresh** `adapter.Adapter` at each Start and discards it when the cycle ends. `Run()`'s goroutine closes both `eventCh` and `askCh` via `defer` after `RunCycle` returns — the observer and the asker are the sole writers, so a sole-writer close eliminates the close-during-emit panic by construction. `Adapter.Stop()` and `tui.Run`'s `defer adp.Stop()` are **removed**; the root no longer builds or holds an adapter. Each cycle gets fresh channels, so re-running a second cycle is clean. (Chosen over reuse-with-guarded-Stop: the reused design can't sole-writer-close, since the channels must survive for a re-run, forcing a reader-side close guard instead of designing the race out.) **Reader-side note:** sole-writer close removes the *send*-side race, but the router still tears the cycle down on `EngineDoneMsg` (nils the adapter) while a buffered event may still be in flight on a separate goroutine — so the router's pump re-issue arms are nil-guarded (Commit 5 Task 4). That hazard is orthogonal to the channel-close discipline: it lives on the reference the router uses to re-arm the pump, which the channel close can't see.
2. **Cycle-started message rides `eventCh` as `EvtCycleStarted` (resolves OQ 2).** No new channel. `Run()` emits one `ObserverEventMsg{Kind: EvtCycleStarted, Payload: CycleStartedPayload{...}}` at the top of the closure — after `Turn.Start` rolls the order, before `RunCycle` proceeds. The buffered `eventCh` send is the happens-before; the existing `ObserverPump` delivers it as the first event. The payload carries the cycle number and a **copy** of the order with display names (built pre-loop, single-goroutine, no race).
3. **The ask-pump mirrors `ObserverPump` exactly (resolves OQ 3).** `Adapter.AskPump() tea.Cmd` reads one `CollectorAskMsg` off `askCh` and returns it as a `tea.Msg` (the surfaced `CollectorAskMsg` itself), `nil` on close. The router re-issues it after each receipt, exactly as it re-issues `ObserverPump`, and stops re-issuing on `EngineDoneMsg`.
4. **`SelectAction` gains a nil-reply branch (part of OQ 7).** A boxed untyped `nil` does not satisfy `raw.(action.Action)`, so the placeholder's "skip faction" default would currently error. The adapter's `phaseCollector.SelectAction` is reshaped to return `(nil, nil)` when the reply is nil — the legal skip path (`orchestrator.go:418`). Pointer- and slice-typed defaults (`*FactionStat`, `[]MovementDecision`, `[]*Asset`) survive boxing as typed nils/empties and need no adapter change.
5. **Effort 1 keeps `stubActionCollector` untouched.** `SelectAction` auto-returns `nil` (skip), so `action.Collector` is never reached; its stubs stay loud-failing. The action surface is Effort 3's problem.
6. **The placeholder waits for a keypress (paces the cycle).** Per Discovery's *Modal overlay infrastructure → Effort 1 placeholder*: each prompt renders (which ask, which faction) and emits its default only on a keypress. The final `cycle_summary` checkpoint placeholder is therefore the deliberate "done reading" pause; on its ack `RunCycle` returns and the router transitions back to setup. No special end-of-cycle state is needed for the happy path.
7. **`Severity` + `StatusLiner` live in a leaf `chrome` package, not package `tui` (resolves OQ 6's home, surfaced in planning).** Defining them in `tui` would force `execution → tui → turn → execution` once the execution sub returns a severity (`tui` imports `turn`, `turn` imports `execution`). A leaf package both `tui` and the turn sub-tree import breaks the cycle — the same trick the existing `Helper` interface gets for free by returning the external `help.KeyMap`. The root maps `Severity` to a style (`Recoverable → styles.Warning`, `Fatal → styles.Danger`); the bottom bar is always present, and `?` toggles compact↔full help rather than the bar's existence. (Chosen over returning a pre-styled string — the typed severity keeps styling decisions at the root, where the bar is composed.)
8. **`turn/msgs` leaf package for cross-sub messages; setup owns its own `bubbles/list.Model` (resolves two Commit-5 sketch drifts, surfaced in planning).** (a) `StartCycleMsg` can't live in package `turn`: `setup` constructs it while the router imports `setup`, so a `turn`-owned type is a `setup → turn` cycle — the same constraint `views/manage/msgs` already solves, so `turn` follows the convention with `views/turn/msgs`. (b) The plan's "reuse `views/manage/list`" is unworkable: that list is wired to manage's `n`/`enter` messages, advertises New/Select, and its empty-state points at `n`. Setup builds its own display-only `list.Model` (Start on `enter`, empty-state pointing back to Manage via `tab`), duplicating the small item projection; a shared read-only list is a deferred refactor (two consumers don't yet justify the extraction).
9. **Commit 5 ships a minimal `execution.Model` placeholder; Commit 6 rewrites it in full (sequencing, surfaced in planning).** The router (`turn.go`) owns an `execution execution.Model` field, so the package must exist for Commit 5 to compile and route. Rather than split the router's wiring across two commits (editing `turn.go` twice), Commit 5 creates a minimal `execution.go` whose `New(width, height int)` constructor and method set form the stable contract the router binds to; Commit 6 rewrites the file (full contents) with the three-pane rendering, preserving that contract. `turn.go` is therefore final after Commit 5; Commits 6–7 touch only the `execution` and `overlay` packages.
10. **The `overlay` interface package is created in Commit 6, not Commit 7 (resolves the Commit-6 struct-import sequencing, surfaced in planning).** `execution.Model` holds an `overlay.Overlay` field. For the struct to be defined once — in Commit 6 Task 3's full rewrite — the `overlay/overlay.go` interface must already exist at Commit 6, so Commit 6 Task 2 creates it (`Overlay` + `OverlayDoneMsg`). Commit 7 then adds only the concrete `placeholder.go` and the dispatch that mounts it; it never edits the `execution.Model` struct. (Chosen over splitting the struct across Commits 6–7: a single struct definition beats a two-commit field-by-field assembly. The cost is one interface file landing a commit before its first implementer — inert until then.)
11. **The mounted overlay replaces the panes; it is not composited over them (resolves the Commit-7 `View` rendering, surfaced in planning).** While `overlay != nil`, `execution.View` returns the prompt centered in the content area (`lipgloss.Place`) — the three panes are hidden behind it, matching every existing modal in the app (root quit `view.go:47`, manage delete-confirm `manage.go:171`); the codebase has no over-background compositing helper. The discovery's "over dimmed panes" look is deferred to Effort 2, where real overlays benefit from pane context — and since it is a `View`-only change (not a dispatch-flow change), deferring it does not churn the seam. (Chosen over building a string-splice compositing helper now: YAGNI, and it matches the universal app pattern.)

## Open Questions — Resolved Here

All seven routed to Effort 1 are resolved in this plan (no deferrals to Execution):

| OQ | Where resolved |
|----|----------------|
| 1 — struct shapes | `turn.Model` Commit 5 Task 4; `setup.Model` Commit 5 Task 2; `execution.Model` Commit 6 Task 3 |
| 2 — cycle-started msg | Decision 2; Commit 4 Tasks 1–2 |
| 3 — AskPump shape | Decision 3; Commit 4 Task 2 |
| 4 — Stop safety | Decision 1; Commit 4 Task 2 |
| 5 — Mutation formatter | Commit 6 Task 1 |
| 6 — StatusLiner signature | Decision 7; Commit 3 Task 1 |
| 7 — placeholder defaults | Decision 4; Commit 7 Task 2 (`defaultFor`) |

## Shared Context

### Package layout (OQ 1)

New shared layout package (Commit 1) and the `turn` sub-tree (Commits 5–7):

```
internal/faction/tui/layout/        layout.go  — Region, panel, regionWidths, compose (promoted from manage)
internal/faction/tui/chrome/        chrome.go  — Severity + StatusLiner (app-wide bottom-bar contract; leaf pkg)
internal/faction/tui/views/turn/
  turn.go        router Model
  msgs/msgs.go   StartCycleMsg (leaf pkg; setup emits, router handles — avoids setup→turn cycle)
  setup/setup.go     setup.Model — read-only roster + Start
  execution/execution.go   execution.Model — three-pane layout + overlay ownership + StatusLiner
  execution/format.go      event header phrasing + Mutation → string formatter
  overlay/overlay.go       Overlay interface + OverlayDoneMsg + Severity-free; placeholder.go
```

All three `turn` struct shapes are defined in full in their owning tasks: `turn.Model` Commit 5 Task 4, `setup.Model` Commit 5 Task 2, `execution.Model` Commit 6 Task 3 (the minimal Commit 5 Task 3 placeholder shares its `New`/method-set contract).

### Model discipline

We are in Plan (Opus). Execution sessions are Sonnet **default**, with **Opus suggested** for Commit 4 (the channel round-trip seam) and Commit 7 (the dispatch switch / ask round-trip — the first code to exercise the real `askCh`). Prompt to switch at the Plan→Execution transition.

## Out of Scope (Effort 1)

- Real modals for any kind — Efforts 2 (non-action) and 3 (action). Effort 1 is all placeholder.
- `Esc`-cancel routing / `ErrTurnCanceled` — Effort 3 (overview OQ 14). Effort 1 needs only happy-path answers.
- Phase-reactive detail pane — Effort 2. Effort 1 ships the static identity + stats card.
- Per-faction cadence — dropped (Discovery Decision 12); the adapter flag stays at zero.
- Touching `stubActionCollector` (Decision 5) and the action `AskKind`s (Effort 3).

---

## Work Breakdown

Seven commits, one execution session each. Order respects dependencies: 1 (layout) and 4 (adapter) before 6; 3 (StatusLiner) before 6; 6 before 7. App-wide commits (1–3) and the adapter commit (4) are independent of the turn sub-tree and can land first.

### Phase 1 — App-wide foundations

#### Commit 1 — `refactor(tui): promote Region layout to shared tui/layout package`

##### Task 1 — `internal/faction/tui/layout/layout.go` (new)

Verbatim move of the current `views/manage/layout.go`, with package renamed to `layout` and the three callers' identifiers exported (`panel` → `Panel` with `Content`/`Style` fields, `regionWidths` → `RegionWidths`, `compose` → `Compose`; `Region`/`Left`/`Center`/`Right` already exported; `regionOrder` stays private). Full contents:

```go
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Region identifies one of the columns in the working area. To move a component
// to a different column, assign its View to a different Region key — no other
// change is needed; widths and separators follow.
type Region int

const (
	Left Region = iota
	Center
	Right
)

// regionOrder is the left-to-right render order used by Compose.
var regionOrder = []Region{Left, Center, Right}

// Panel is a region's content plus the base style that governs its alignment
// and color. Compose applies the region's width and height on top.
type Panel struct {
	Content string
	Style   lipgloss.Style
}

// RegionWidths splits the working area into a wide center flanked by equal-width
// left and right columns, accounting for the two single-column separators
// between them. Remainder goes to the center.
func RegionWidths(termWidth int) map[Region]int {
	totalW := max(termWidth-2, 4) // two 1-col separators
	left := totalW / 4
	right := totalW / 4
	return map[Region]int{
		Left:   left,
		Center: totalW - left - right,
		Right:  right,
	}
}

// Compose renders each region's panel into its width at the given height and
// joins them left-to-right with vertical separators between.
func Compose(panels map[Region]Panel, widths map[Region]int, height int) string {
	sep := styles.Rule.Render(
		strings.Repeat("│\n", height-1) + "│",
	)
	blocks := make([]string, 0, len(regionOrder)*2-1)
	for i, region := range regionOrder {
		if i > 0 {
			blocks = append(blocks, sep)
		}
		p := panels[region]
		blocks = append(blocks, p.Style.Width(widths[region]).Height(height).Render(p.Content))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}
```

##### Task 2 — `internal/faction/tui/views/manage/manage.go`

Repoint the four call sites at the new package and add its import. The `styles` import stays (still used by `Dim`/`SaveError`).

Add the `layout` import (alphabetical, before `styles`):

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/layout"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
```

The `WindowSizeMsg` forward (one occurrence, in `Update`'s `tea.WindowSizeMsg` case):

**Find:**
```go
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: regionWidths(msg.Width)[Left], Height: msg.Height})
```
**Replace with:**
```go
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: layout.RegionWidths(msg.Width)[layout.Left], Height: msg.Height})
```

The post-CRUD list rebuild (two identical occurrences, in the `CreatedMsg` and `DeletedMsg` cases — apply to both):

**Find:**
```go
			m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: regionWidths(m.termWidth)[Left], Height: m.termHeight})
```
**Replace with:**
```go
			m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: layout.RegionWidths(m.termWidth)[layout.Left], Height: m.termHeight})
```

The `View` compose block (in the `viewList` case):

**Find:**
```go
		panels := map[Region]panel{
			Left:   {content: m.list.View(), style: lipgloss.NewStyle()},
			Center: {content: "—", style: placeholder},
			Right:  {content: "—", style: placeholder},
		}
		content = compose(panels, regionWidths(m.termWidth), contentH)
```
**Replace with:**
```go
		panels := map[layout.Region]layout.Panel{
			layout.Left:   {Content: m.list.View(), Style: lipgloss.NewStyle()},
			layout.Center: {Content: "—", Style: placeholder},
			layout.Right:  {Content: "—", Style: placeholder},
		}
		content = layout.Compose(panels, layout.RegionWidths(m.termWidth), contentH)
```

Then delete the now-empty source file:

```
rm internal/faction/tui/views/manage/layout.go
```

##### Task 3 — verify
`go build ./...` and `go test ./internal/faction/tui/...`. No behavior change; Manage renders identically.

##### Commit message
```
refactor(tui): promote Region layout to shared tui/layout package

- move Region/Left/Center/Right, regionWidths, panel, compose from
  views/manage/layout.go into new internal/faction/tui/layout package
- export as Region/RegionWidths/Panel{Content,Style}/Compose; keep
  regionOrder private
- repoint manage.go imports and call sites; delete views/manage/layout.go
- pure extraction, no behavior change; both Manage and Turn import it
```

#### Commit 2 — `feat: app-wide campaign logging via --debug flag`

`logging.New(logsDir string, debug bool) (*slog.Logger, error)` already exists (`internal/logging/logging.go`) and writes a timestamped file + `latest.log`. Spatial loading is already in `factionRun` (shipped by the manage initiative); this commit only adds the `--debug` flag and swaps `factionRun`'s stderr logger for a campaign-file logger. `RunDryRun` keeps the stderr logger.

##### Task 1 — `cmd/gm-toolkit/faction.go`

Add the `logging` import (alphabetical, after `.../faction/tui`, before `.../spatial`):

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui"
	"github.com/therobertcrocker/gm-toolkit/internal/logging"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```

Declare the `--debug` flag variable:

**Find:**
```go
	var dryRun bool
	var campaignOverride string
```
**Replace with:**
```go
	var dryRun bool
	var debug bool
	var campaignOverride string
```

Pass `debug` (not the stderr `log`) into `factionRun` — the stderr handler still drives the `dryRun` branch:

**Find:**
```go
			return factionRun(log, campaignOverride)
```
**Replace with:**
```go
			return factionRun(debug, campaignOverride)
```

Register the flag:

**Find:**
```go
	cmd.Flags().BoolVar(&dryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
	cmd.Flags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
```
**Replace with:**
```go
	cmd.Flags().BoolVar(&dryRun, "dryrun", false, "run a one-faction smoke cycle and exit")
	cmd.Flags().BoolVar(&debug, "debug", false, "write debug-level logs to the campaign log file")
	cmd.Flags().StringVar(&campaignOverride, "campaign", "", "campaign id to use instead of the active one")
```

Reshape `factionRun`'s signature — drop the `log` param (unused before `paths`; the early `LoadRegistry`/`ResolveActive` errors are returned, not logged), take `debug`:

**Find:**
```go
func factionRun(log *slog.Logger, campaignOverride string) error {
```
**Replace with:**
```go
func factionRun(debug bool, campaignOverride string) error {
```

Build the campaign logger right after `paths` resolves; `log` then flows unchanged into `engine.NewWithRulebook(rb, log)` and `tui.Run(..., log)` below (`:=` is valid — `err` already exists from `LoadRegistry`, `log` is new):

**Find:**
```go
	paths := camp.Paths()

	rb, err := rulebook.Load(paths.FactionDataDir)
```
**Replace with:**
```go
	paths := camp.Paths()

	log, err := logging.New(paths.LogsDir, debug)
	if err != nil {
		return fmt.Errorf("faction: init logging in %s: %w", paths.LogsDir, err)
	}

	rb, err := rulebook.Load(paths.FactionDataDir)
```

##### Task 2 — verify
`go build ./...`. Run `gm-toolkit faction` against a campaign → confirm `<campaign>/state/logs/<ts>.log` and the `latest.log` symlink are written. Run with `--debug` → confirm the file captures Debug-level lines. `dryRun` output still goes to stderr.

##### Commit message
```
feat: app-wide campaign logging via --debug flag

- add --debug flag to the faction command
- factionRun builds a campaign-file logger via logging.New(paths.LogsDir,
  debug) instead of taking the stderr logger; log flows into the engine
  and TUI
- RunDryRun keeps the stderr logger
```

#### Commit 3 — `feat(tui): contextual bottom bar + StatusLiner`

The `if m.showHelp` footer becomes an always-present bar so the content area never shifts when an error appears. `Severity` + `StatusLiner` live in a new **leaf** package `chrome` (not package `tui`) so the turn sub-tree can return a severity without an `execution → tui → turn → execution` import cycle. `?` is repurposed from "show/hide the bar" to "compact ↔ full help". `View` and `contentHeight` share one `footerView()` so the reserved height always matches what's drawn (resolves OQ 6).

##### Task 1 — `internal/faction/tui/chrome/chrome.go` (new)

Full contents:

```go
package chrome

// Severity ranks a status-line message for the root bottom bar. The root maps
// it to a style (Recoverable -> Warning, Fatal -> Danger).
type Severity int

const (
	Info Severity = iota
	Recoverable
	Fatal
)

// StatusLiner is implemented by sub-models that surface a status line in the
// root's bottom bar. Empty text means "no status — fall through to help".
type StatusLiner interface {
	StatusLine() (text string, severity Severity)
}
```

##### Task 2 — `internal/faction/tui/view.go`

Add the `chrome` import (alphabetical, before `styles`):

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
```

Replace the `showHelp`-gated footer block with a single `footerView()` call (the top-of-`View` `rule` variable is still used by the header, so it stays):

**Find:**
```go
	if m.showHelp {
		var subBindings help.KeyMap = emptyKeyMap{}
		if helper, ok := m.subs[m.bar.Active()].(Helper); ok {
			if km := helper.Help(); km != nil {
				subBindings = km
			}
		}
		composed := combineKeyMaps(globalsKeyMap{}, subBindings)
		sb.WriteString("\n" + rule + "\n")
		sb.WriteString("  " + m.help.View(composed) + "\n")
		sb.WriteString(rule)
	}
```
**Replace with:**
```go
	sb.WriteString(m.footerView())
```

Add `footerView` and its two helpers immediately after `View`'s closing brace:

**Find:**
```go
	return sb.String()
}

type emptyKeyMap struct{}
```
**Replace with:**
```go
	return sb.String()
}

// footerView renders the always-present bottom bar: a status line from the
// active sub-model when it has one, otherwise the help bar. View and
// contentHeight both call this so the reserved footer height always matches
// what's drawn.
func (m Model) footerView() string {
	ruleWidth := m.width
	if ruleWidth <= 0 {
		ruleWidth = 80
	}
	rule := styles.AppDivider.Render(strings.Repeat("─", ruleWidth))

	var inner string
	if text, severity, ok := activeStatusLine(m); ok {
		inner = statusStyle(severity).Render(text)
	} else {
		var subBindings help.KeyMap = emptyKeyMap{}
		if helper, ok := m.subs[m.bar.Active()].(Helper); ok {
			if km := helper.Help(); km != nil {
				subBindings = km
			}
		}
		inner = m.help.View(combineKeyMaps(globalsKeyMap{}, subBindings))
	}
	return "\n" + rule + "\n  " + inner + "\n" + rule
}

// activeStatusLine returns the active sub-model's status line when it implements
// chrome.StatusLiner and the text is non-empty.
func activeStatusLine(m Model) (text string, severity chrome.Severity, ok bool) {
	statusLiner, isLiner := m.subs[m.bar.Active()].(chrome.StatusLiner)
	if !isLiner {
		return "", 0, false
	}
	text, severity = statusLiner.StatusLine()
	return text, severity, text != ""
}

func statusStyle(severity chrome.Severity) lipgloss.Style {
	switch severity {
	case chrome.Fatal:
		return styles.Danger
	case chrome.Recoverable:
		return styles.Warning
	default:
		return styles.Body
	}
}

type emptyKeyMap struct{}
```

##### Task 3 — `internal/faction/tui/update.go`

Add the `lipgloss` import (`contentHeight` now measures the rendered footer):

**Find:**
```go
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
```
**Replace with:**
```go
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/modebar"
```

Repurpose `?` from toggling the bar's existence to toggling full help:

**Find:**
```go
		case "?":
			m.showHelp = !m.showHelp
			return m.resizeSubs()
```
**Replace with:**
```go
		case "?":
			m.help.ShowAll = !m.help.ShowAll
			return m.resizeSubs()
```

Always reserve the footer, measured from the rendered bar (drop the `footerHeight` const and the `showHelp` gate):

**Find:**
```go
// Layout budget owned by the root. The header is the title rule/line/rule plus
// the mode bar and a blank line; the footer is the help line bracketed by two
// rules. Sub-models are handed the remaining height and never see the chrome.
const (
	headerHeight = 5
	footerHeight = 3
)

func (m Model) contentHeight() int {
	reserved := headerHeight
	if m.showHelp {
		reserved += footerHeight
	}
	if h := m.height - reserved; h > 1 {
		return h
	}
	return 1
}
```
**Replace with:**
```go
// Layout budget owned by the root. The header is the title rule/line/rule plus
// the mode bar and a blank line. The footer (always present) is the status/help
// bar bracketed by two rules; its height varies with compact vs. full help, so
// it is measured from the rendered bar rather than a constant. Sub-models are
// handed the remaining height and never see the chrome.
const headerHeight = 5

func (m Model) contentHeight() int {
	reserved := headerHeight + lipgloss.Height(m.footerView())
	if h := m.height - reserved; h > 1 {
		return h
	}
	return 1
}
```

##### Task 4 — `internal/faction/tui/model.go`

Retire the `showHelp` field (the help model's `ShowAll`, initialized `false` at the line above, now carries the compact/full state):

**Find:**
```go
	confirmExit bool
	priorMode   modebar.Mode
	showHelp    bool
}
```
**Replace with:**
```go
	confirmExit bool
	priorMode   modebar.Mode
}
```

Drop its initializer:

**Find:**
```go
		bar:          modebar.New(),
		adapter:      adp,
		help:         h,
		showHelp:     true,
		factionState: factionState,
```
**Replace with:**
```go
		bar:          modebar.New(),
		adapter:      adp,
		help:         h,
		factionState: factionState,
```

##### Task 5 — verify
`go build ./...` && `go test ./internal/faction/tui/...`. Manually: Manage (no `StatusLiner`) shows the help bar; `?` toggles compact↔full and the content area shrinks/grows to match (never overlaps the panes); an error never shifts the layout because the footer row count is always reserved.

##### Commit message
```
feat(tui): contextual bottom bar + StatusLiner

- add chrome leaf package: Severity + StatusLiner contract (separate
  package so the turn sub-tree can return a severity without an import
  cycle through tui)
- view.go: always-present bottom bar via footerView() — status line from
  the active sub when present, else help; recoverable->Warning,
  fatal->Danger
- update.go: ? toggles compact<->full help (not bar existence); footer
  height measured from the rendered bar so content never shifts
- retire the showHelp field; help.ShowAll carries compact/full state
```

### Phase 2 — Adapter contract

#### Commit 4 — `feat(adapter): start/resume guard, ask-pump, cycle-started, writer-side channel close`
*(suggest Opus)*

This commit reshapes the adapter so the engine goroutine owns the channel lifecycle (sole-writer close, Decision 1), emits the synthetic cycle-started event (Decision 2), exposes the ask-pump (Decision 3), and gives `SelectAction` its nil-reply skip branch (Decision 4). It also unwires the root from the adapter: the router builds one per cycle (Commit 5), so the root no longer holds it.

##### Task 1 — `internal/faction/tui/adapter/channels.go`

Append `EvtCycleStarted` to the `EventKind` const block. It goes **last** so the existing iota values don't renumber (other code switches on these).

**Find:**
```go
	EvtIndexSkipped
)
```
**Replace with:**
```go
	EvtIndexSkipped
	EvtCycleStarted
)
```

Add the payload types `EvtCycleStarted` rides on. `RailEntry` is the display-name copy the execution rail captures once; both are defined here and cited by later commits (the router's `execution.Model.order` is `[]adapter.RailEntry`, Commit 6 Task 3).

**Find:**
```go
type IndexSkippedPayload struct{ Skipped []string }
```
**Replace with:**
```go
type IndexSkippedPayload struct{ Skipped []string }

type RailEntry struct{ ID, Name string }
type CycleStartedPayload struct {
	CycleNumber int
	Order       []RailEntry
}
```

##### Task 2 — `internal/faction/tui/adapter/run.go` (rewritten)

Full rewrite. `Run` gains the start/resume guard (RunCycle requires `Turn.Start` already called or a turn still `InProgress` — `orchestrator.go:31-32`), emits `EvtCycleStarted` pre-loop, and `defer close`s both channels (sole-writer close, Decision 1). `emitCycleStarted` reads `CurrentTurn` single-threaded before `RunCycle`, so the order read races nothing; it copies display names so the view never holds engine pointers. `AskPump` mirrors `ObserverPump`. `Stop` is deleted (the defers now own close).

```go
package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (a *Adapter) Run() tea.Cmd {
	return func() tea.Msg {
		defer close(a.eventCh)
		defer close(a.askCh)
		if !a.engine.Turn.InProgress(a.factionState) {
			if err := a.engine.Turn.Start(a.factionState); err != nil {
				return EngineDoneMsg{Err: err}
			}
		}
		a.emitCycleStarted()
		err := a.engine.RunCycle(a.factionState, a.paths, a.Collectors(), a.Observer())
		return EngineDoneMsg{Err: err}
	}
}

// emitCycleStarted sends the synthetic EvtCycleStarted as the cycle's first
// event. It runs on the Run goroutine before RunCycle, single-threaded against
// CurrentTurn (non-nil here: either InProgress was already true or Start just
// set it), so reading the order races nothing. The order is copied with display
// names so the execution view never holds engine pointers.
func (a *Adapter) emitCycleStarted() {
	current := a.factionState.CurrentTurn
	order := make([]RailEntry, 0, len(current.FactionOrder))
	for _, id := range current.FactionOrder {
		name := id
		if faction, ok := a.factionState.Factions[id]; ok {
			name = faction.Name
		}
		order = append(order, RailEntry{ID: id, Name: name})
	}
	a.eventCh <- ObserverEventMsg{
		Kind: EvtCycleStarted,
		Payload: CycleStartedPayload{
			CycleNumber: current.CycleNumber,
			Order:       order,
		},
	}
}

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

// AskPump reads one CollectorAskMsg from askCh and returns it as a tea.Msg,
// mirroring ObserverPump. The router re-issues it after each receipt and stops
// on EngineDoneMsg. Returns nil when Run closes askCh.
func (a *Adapter) AskPump() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-a.askCh
		if !ok {
			return nil
		}
		return msg
	}
}
```

> The `doneCh` field on `Adapter` (`adapter.go`) is now fully dead — `Run` returns `EngineDoneMsg` directly and nothing reads the channel. Leaving it is out of scope to churn; remove it only if the execution session confirms zero references. No behavior depends on it either way.

##### Task 3 — `internal/faction/tui/adapter/phase_collector.go`

In `SelectAction`, add the nil-reply branch so the placeholder's untyped-nil skip (Decision 4) maps to the engine's legal skip path (`orchestrator.go:418` — `selectedAction == nil` → `OnFactionSkipped`). Without it, a boxed `nil` fails `raw.(action.Action)` and errors.

**Find:**
```go
	raw, err := p.ask(AskSelectAction, faction, SelectActionPayload{Available: available})
	if err != nil {
		return nil, err
	}
	chosen, ok := raw.(action.Action)
```
**Replace with:**
```go
	raw, err := p.ask(AskSelectAction, faction, SelectActionPayload{Available: available})
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil // legal skip (orchestrator.go:418)
	}
	chosen, ok := raw.(action.Action)
```

##### Task 4 — `internal/faction/tui/tui.go`

`Run` no longer builds the adapter or defers `Stop` (the router owns the per-cycle adapter, Commit 5). Drop the now-unused `adapter` import and pass `eng` + `log` straight into `NewModel`.

Drop the `adapter` import:

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```

Replace the adapter construction + program build:

**Find:**
```go
	adp := adapter.New(eng, factionState, paths, log)
	defer adp.Stop()

	program := tea.NewProgram(NewModel(adp, factionState, paths, rb, spatialMap), tea.WithAltScreen())
```
**Replace with:**
```go
	program := tea.NewProgram(NewModel(eng, factionState, paths, rb, spatialMap, log), tea.WithAltScreen())
```

##### Task 5 — `internal/faction/tui/model.go`

`NewModel` drops the `adp *adapter.Adapter` param and the `adapter` field; gains `eng *engine.Engine` + `log *slog.Logger` (both consumed by `turn.New` in Commit 5 Task 4). Anchors assume Commit 3 has already landed (the `showHelp` field/initializer are gone).

> **Sequencing:** this commit keeps `subs[ModeTurn] = turn.New()` (the no-arg stub). `eng` and `log` are unused params here — legal Go, and their *types* keep the new `engine`/`slog` imports used, so `go build` stays green. Commit 5 Task 4 rewrites the call to `turn.New(factionState, paths, eng, log)`.

Swap the imports — drop `adapter`, add `log/slog` and `engine`:

**Find:**
```go
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
```
**Replace with:**
```go
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
```

Remove the `adapter` field:

**Find:**
```go
	subs         map[modebar.Mode]tea.Model
	adapter      *adapter.Adapter
	help         help.Model
```
**Replace with:**
```go
	subs         map[modebar.Mode]tea.Model
	help         help.Model
```

Reshape `NewModel`'s signature and drop the `adapter:` initializer:

**Find:**
```go
func NewModel(
	adp *adapter.Adapter,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
) Model {
	h := help.New()
	h.ShowAll = false
	return Model{
		bar:          modebar.New(),
		adapter:      adp,
		help:         h,
```
**Replace with:**
```go
func NewModel(
	eng *engine.Engine,
	factionState *state.FactionState,
	paths *campaigns.Paths,
	rb *rulebook.Rulebook,
	spatialMap *spatial.RegionMap,
	log *slog.Logger,
) Model {
	h := help.New()
	h.ShowAll = false
	return Model{
		bar:          modebar.New(),
		help:         h,
```

##### Task 6 — verify
`go build ./...` and `go test ./internal/faction/tui/adapter/...`. Confirm zero remaining references to `Stop` (`grep -rn "\.Stop()" internal/faction/tui/`). The TUI won't run a cycle end-to-end yet (the router/dispatch land in Commits 5–7), but the adapter package compiles and its tests pass.

##### Commit message
```
feat(adapter): start/resume guard, ask-pump, cycle-started, writer-side channel close

- Run() starts or resumes the turn, emits EvtCycleStarted pre-loop, and
  defer-closes eventCh+askCh; the engine goroutine is the sole writer, so
  the close can never race an in-flight send
- add AskPump (mirror of ObserverPump) and the RailEntry/CycleStartedPayload
  types; delete Stop()
- SelectAction returns (nil, nil) on a nil reply — the engine's legal
  faction-skip path (orchestrator.go:418)
- root no longer builds or holds the adapter: NewModel takes eng+log, tui.Run
  drops adapter.New/defer Stop (the router builds one per cycle, Commit 5)
```

### Phase 3 — Turn sub-tree

#### Commit 5 — `feat(tui/turn): router + setup view`

Lands the turn router, the setup roster, and the `msgs` leaf package. The router fully owns the per-cycle adapter lifecycle (build on Start, pump-alive loop, discard on done) — that wiring is final after this commit; Commits 6–7 only touch `execution`. To let the router compile and route, this commit ships a **minimal** `execution.Model` placeholder (Decision 8); Commit 6 rewrites it in full.

##### Task 1 — `internal/faction/tui/views/turn/msgs/msgs.go` (new)

The leaf message package — mirrors `views/manage/msgs`. `StartCycleMsg` is emitted by `setup` and handled by the router; a separate package is required because `setup` constructs the value while the router imports `setup`, so the type can't live in package `turn` (that would be a `setup → turn` cycle). Full contents:

```go
package msgs

// StartCycleMsg requests the turn router to build a fresh adapter and begin a
// cycle. Emitted by the setup view on the Start key; handled by turn.Model.
type StartCycleMsg struct{}
```

##### Task 2 — `internal/faction/tui/views/turn/setup/setup.go` (new)

Read-only roster + Start. Builds its **own** `bubbles/list.Model` (the manage list can't be reused — it's wired to manage's `n`/`enter` messages). Display-only delegate, `Start = enter`, alphabetical by ID. Three-region `View` via the promoted `layout` package (Commit 1), `Left` = roster, `Center`/`Right` dimmed — matching `manage.View`'s list case. Empty-state points back to Manage via `tab` (the real mode-switch key, `update.go:43`), not a local `n`. The `item` projection (name/scale/HP/Coin) duplicates `manage/list`'s by design (decoupled; a shared read-only list is a deferred refactor). Full contents:

```go
package setup

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/layout"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
)

type item struct {
	name, scale, hp, coin string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return fmt.Sprintf("%s · HP %s · Coin %s", i.scale, i.hp, i.coin) }
func (i item) FilterValue() string { return i.name }

type keyMap struct {
	Start key.Binding
}

type Model struct {
	factionState *state.FactionState
	list         list.Model
	keys         keyMap
	termWidth    int
	termHeight   int
}

func New(factionState *state.FactionState) Model {
	l := list.New(buildItems(factionState), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Factions"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return Model{
		factionState: factionState,
		list:         l,
		keys: keyMap{
			Start: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "start cycle")),
		},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Start) && len(m.factionState.Factions) > 0 {
			return m, func() tea.Msg { return msgs.StartCycleMsg{} }
		}
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.list.SetSize(layout.RegionWidths(msg.Width)[layout.Left], msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	contentH := max(m.termHeight, 1)
	placeholder := styles.Dim.Align(lipgloss.Center).AlignVertical(lipgloss.Center)

	leftContent := m.list.View()
	if len(m.factionState.Factions) == 0 {
		leftContent = emptyState()
	}
	panels := map[layout.Region]layout.Panel{
		layout.Left:   {Content: leftContent, Style: lipgloss.NewStyle()},
		layout.Center: {Content: "—", Style: placeholder},
		layout.Right:  {Content: "—", Style: placeholder},
	}
	return layout.Compose(panels, layout.RegionWidths(m.termWidth), contentH)
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding  { return []key.Binding{h.keys.Start} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.Start}} }

func buildItems(factionState *state.FactionState) []list.Item {
	ids := make([]string, 0, len(factionState.Factions))
	for id := range factionState.Factions {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]list.Item, 0, len(ids))
	for _, id := range ids {
		faction := factionState.Factions[id]
		items = append(items, item{
			name:  faction.Name,
			scale: string(faction.Scale),
			hp:    fmt.Sprintf("%d/%d", faction.CurrentHP, faction.MaxHP),
			coin:  fmt.Sprintf("%d", faction.Coin),
		})
	}
	return items
}

func emptyState() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Render("No factions yet."),
		styles.Dim.Render("Create factions in Manage (tab), then return here to run a cycle."),
	)
}
```

##### Task 3 — `internal/faction/tui/views/turn/execution/execution.go` (new, minimal)

Minimal placeholder so the router compiles and routes (Decision 8). The constructor and method set defined here are the **stable contract** the router (Task 4) binds to; Commit 6 Task 3 rewrites this file in full keeping the same `New(width, height int) Model` signature and the same method set. Full contents:

```go
package execution

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Model renders a running cycle. This minimal version ships in Commit 5 so the
// router compiles and routes; Commit 6 rewrites it with the three-pane layout,
// event stream, and detail card.
type Model struct {
	width, height int
}

func New(width, height int) Model { return Model{width: width, height: height} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
	}
	return m, nil
}

func (m Model) View() string {
	return styles.Placeholder.Render("Running cycle… (execution view lands in Commit 6)")
}

func (m Model) Help() help.KeyMap { return nil }

func (m Model) CapturesInput() bool { return false }

func (m Model) StatusLine() (string, chrome.Severity) { return "", chrome.Info }
```

##### Task 4 — `internal/faction/tui/views/turn/turn.go` (rewritten)

The router. Replaces the stub. Owns the per-cycle adapter: on `StartCycleMsg` it builds a fresh `adapter.New(...)` and a fresh `execution.New(...)`, switches to `viewExecution`, and batches `Run` + both pumps; the pump-alive loop re-issues `ObserverPump`/`AskPump` on each receipt so `execution` stays a pure projection. On a closed channel the pumps return a nil `tea.Msg` (dropped by bubbletea), so the loop self-terminates — no explicit stop needed. On happy-path `EngineDoneMsg` it returns to setup, **nils the adapter**, and rebuilds the roster (post-cycle HP/Coin); on error it keeps the execution view and forwards the error there. The re-issue arms in the `ObserverEventMsg`/`CollectorAskMsg` cases are **nil-guarded**: `EngineDoneMsg` and a trailing buffered event race independently to the update loop (separate goroutines, no ordering guarantee — see Decision 1's reader-side note), so if `EngineDoneMsg` lands first and nils the adapter, a late `ObserverEventMsg` must not re-arm `ObserverPump()` on the nil adapter. Without the guard the re-issued pump closure dereferences `a.eventCh` on a nil `*Adapter` and panics. (The error path keeps the adapter, so only the happy path is exposed.) `Help`/`CapturesInput`/`StatusLine` delegate on `view`, satisfying the root's `chrome.StatusLiner` and `inputCapturer` assertions (`update.go:71`). This file is **final after Commit 5** — Commits 6–7 touch only `execution`. Full contents:

```go
package turn

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/campaigns"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/execution"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/msgs"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/setup"
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
	log          *slog.Logger

	adapter   *adapter.Adapter // nil until Start; built per cycle, discarded on return to setup
	view      view
	setup     setup.Model
	execution execution.Model

	termWidth, termHeight int
}

func New(factionState *state.FactionState, paths *campaigns.Paths, eng *engine.Engine, log *slog.Logger) Model {
	return Model{
		factionState: factionState,
		paths:        paths,
		eng:          eng,
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
		m.execution = execution.New(m.termWidth, m.termHeight)
		m.view = viewExecution
		return m, tea.Batch(m.adapter.Run(), m.adapter.ObserverPump(), m.adapter.AskPump())

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

func (m Model) StatusLine() (string, chrome.Severity) {
	if m.view == viewExecution {
		return m.execution.StatusLine()
	}
	return "", chrome.Info
}
```

##### Task 5 — `internal/faction/tui/model.go`

Finalize the `turn.New` call from Commit 4 Task 5 — now `eng` and `log` (NewModel params since Commit 4) are consumed.

**Find:**
```go
			modebar.ModeManage: manage.New(factionState, paths, rb, spatialMap),
			modebar.ModeTurn:   turn.New(),
```
**Replace with:**
```go
			modebar.ModeManage: manage.New(factionState, paths, rb, spatialMap),
			modebar.ModeTurn:   turn.New(factionState, paths, eng, log),
```

##### Task 6 — verify
`go build ./...` && `go test ./internal/faction/tui/...`. Manually: launch Turn → the alphabetical roster renders in the left pane; an empty campaign shows the empty-state hint. Press `enter` with factions present → the view switches to the execution placeholder without crashing. (The cycle stalls on the first collector ask — there's no reply path until Commit 7's dispatch — which is expected; quitting exits cleanly.)

##### Commit message
```
feat(tui/turn): router + setup view

- add views/turn/msgs leaf package (StartCycleMsg) — setup emits, router
  handles; separate package avoids a setup->turn import cycle
- setup view: read-only alphabetical roster on its own bubbles list, Start
  on enter, empty-state pointing back to Manage; three-region layout
- turn router: owns the per-cycle adapter (build on Start, pump-alive loop,
  discard on done), delegates Help/CapturesInput/StatusLine on view; returns
  to setup and rebuilds the roster on cycle completion
- ship a minimal execution placeholder so the router compiles and routes;
  Commit 6 rewrites it in full
- model.go: finalize turn.New(factionState, paths, eng, log)
```

#### Commit 6 — `feat(tui/turn): execution view — three-pane + event stream + mutation formatter`

Rewrites the minimal `execution.Model` from Commit 5 into the full three-pane live-cycle view, adds the event/mutation formatter, and creates the `overlay` interface package. The interface lands here (not Commit 7) so `execution.Model` can hold the `overlay.Overlay` field in a single struct definition (Decision 10); Commit 7 adds only the concrete placeholder and the dispatch that mounts it. `turn.go` is untouched (final after Commit 5) — the `New(width, height int)` constructor and method set from Commit 5 Task 3 are preserved, so the router keeps compiling.

##### Task 1 — `internal/faction/tui/views/turn/execution/format.go` (new)

`renderEvent` over all 14 observer events + the synthetic `EvtCycleStarted`, and `formatMutation` over every concrete `domain.Mutation`. Header phrasing names the faction for the four `*domain.Faction` events (`EvtFactionTurnStarted/Skipped/TurnCompleted/StatRaiseSkipped`, observer.go:24-27) and is phase-only for the rest — the mutation-carrying payloads (`GoalLockAppliedPayload`, `BookkeepingAppliedPayload`, etc., channels.go:72-96) do **not** include the faction, and the rail + detail card already show whose turn it is. `formatMutation` surfaces player-meaningful deltas and folds bookkeeping noise to `("", false)`.

> Two corrections from grounding against `domain/mutation.go`: (a) `AssetAdded` surfaces `Asset.DefinitionID`, not a name — `domain.Asset` (asset.go:91) has no `Name` field (it lives on the rulebook's `AssetDefinition`) and this formatter has no rulebook; (b) `AssetMoved` (mutation.go:157) is the 31st concrete mutation, absent from both lists in the original sketch — it folds as per-step positional noise (`MovementOrder*` carry the meaningful moves).

`internal/faction/tui/views/turn/execution/format.go` (new) — full contents:

```go
package execution

import (
	"fmt"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// renderEvent turns one observer event into a rendered stream block: a header
// line followed by indented, player-meaningful mutation lines. Faction-named
// events render their name; phase events (whose payloads carry no faction)
// render a phase phrase only.
func renderEvent(msg adapter.ObserverEventMsg) string {
	header, muts := eventHeader(msg)
	if len(muts) == 0 {
		return header
	}
	var b strings.Builder
	b.WriteString(header)
	for _, mut := range muts {
		if line, include := formatMutation(mut); include {
			b.WriteString("\n" + styles.Dim.Render("    "+line))
		}
	}
	return b.String()
}

// eventHeader returns the styled header line and the mutations to indent under
// it (nil for events that carry none).
func eventHeader(msg adapter.ObserverEventMsg) (string, []domain.Mutation) {
	switch msg.Kind {
	case adapter.EvtCycleStarted:
		p := msg.Payload.(adapter.CycleStartedPayload)
		return styles.Strong.Render(fmt.Sprintf("Cycle %d started", p.CycleNumber)), nil
	case adapter.EvtFactionTurnStarted:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "turn started"), nil
	case adapter.EvtFactionSkipped:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "skipped"), nil
	case adapter.EvtFactionTurnCompleted:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "turn completed"), nil
	case adapter.EvtStatRaiseSkipped:
		return factionHeader(msg.Payload.(*domain.Faction).Name, "stat raise skipped"), nil
	case adapter.EvtGoalLockApplied:
		return phaseHeader("goal-lock applied"), msg.Payload.(adapter.GoalLockAppliedPayload).Mutations
	case adapter.EvtBookkeepingApplied:
		return phaseHeader("bookkeeping applied"), msg.Payload.(adapter.BookkeepingAppliedPayload).Mutations
	case adapter.EvtStatRaiseApplied:
		return phaseHeader("stat raise applied"), msg.Payload.(adapter.StatRaiseAppliedPayload).Mutations
	case adapter.EvtMovementTicked:
		return phaseHeader("movement ticked"), msg.Payload.(adapter.MovementTickedPayload).Mutations
	case adapter.EvtMovementResolved:
		return phaseHeader("movement resolved"), msg.Payload.(adapter.MovementResolvedPayload).Mutations
	case adapter.EvtActionSelected:
		p := msg.Payload.(adapter.ActionSelectedPayload)
		name := "(none)"
		if p.Selected != nil {
			name = p.Selected.Name()
		}
		return phaseHeader("action selected: " + name), nil
	case adapter.EvtActionResolved:
		p := msg.Payload.(adapter.ActionResolvedPayload)
		name := "(none)"
		if p.Selected != nil {
			name = p.Selected.Name()
		}
		return phaseHeader("action resolved: " + name), p.Mutations
	case adapter.EvtCycleCompleted:
		p := msg.Payload.(adapter.CycleCompletedPayload)
		return styles.Strong.Render(fmt.Sprintf("Cycle %d completed", p.CycleNumber)), nil
	case adapter.EvtIndexSkipped:
		p := msg.Payload.(adapter.IndexSkippedPayload)
		return phaseHeader("index skipped: " + strings.Join(p.Skipped, ", ")), nil
	case adapter.EvtError:
		return styles.Danger.Render("error: " + msg.Payload.(adapter.ErrorPayload).Err.Error()), nil
	}
	return "", nil
}

func factionHeader(name, phrase string) string {
	return styles.Strong.Render(name) + styles.Subtle.Render("  "+phrase)
}

func phaseHeader(phrase string) string {
	return styles.Subtle.Render(phrase)
}

// formatMutation renders one mutation as a single indented line, returning
// include=false for bookkeeping noise the stream omits. AssetAdded surfaces the
// asset's DefinitionID — domain.Asset carries no display name (that lives on the
// rulebook's AssetDefinition, and this formatter has no rulebook).
func formatMutation(m domain.Mutation) (string, bool) {
	switch mut := m.(type) {
	case domain.FactionHPDelta:
		return fmt.Sprintf("HP %+d", mut.Delta), true
	case domain.AssetHPDelta:
		return fmt.Sprintf("asset %s HP %+d", mut.AssetID, mut.Delta), true
	case domain.BaseHPDelta:
		return fmt.Sprintf("base %s HP %+d", mut.BaseID, mut.Delta), true
	case domain.BaseHealed:
		return fmt.Sprintf("base %s healed %+d", mut.BaseID, mut.Delta), true
	case domain.BaseExpanded:
		return fmt.Sprintf("base %s expanded %+d", mut.BaseID, mut.Delta), true
	case domain.CoinDelta:
		return fmt.Sprintf("Coin %+d", mut.Delta), true
	case domain.AssetAdded:
		return fmt.Sprintf("+asset %s", mut.Asset.DefinitionID), true
	case domain.AssetRemoved:
		return fmt.Sprintf("-asset %s", mut.AssetID), true
	case domain.BaseAdded:
		return fmt.Sprintf("+base %s", mut.Base.ID), true
	case domain.BaseDestroyed:
		return fmt.Sprintf("base %s destroyed", mut.BaseID), true
	case domain.StatRaised:
		return fmt.Sprintf("%s %d→%d", mut.Stat, mut.OldRating, mut.NewRating), true
	case domain.GoalCompleted:
		return fmt.Sprintf("goal completed (+%d XP)", mut.XPAwarded), true
	case domain.GoalAbandoned:
		return "goal abandoned", true
	case domain.HomeworldChanged:
		return fmt.Sprintf("homeworld → %s", mut.ToWorld.WorldID), true
	case domain.MovementOrderIssued:
		return "move issued", true
	case domain.MovementOrderRevised:
		return "move revised", true
	case domain.MovementOrderCancelled:
		return "move cancelled", true
	case domain.MovementOrderCompleted:
		return "move completed", true
	default:
		// Folded as bookkeeping noise or redundant with a surfaced line:
		// GoalTurnsTick, GoalProgressed, GoalInitiated, GoalPhaseAdvanced,
		// AssetMaintainedFlag, AssetStealthApplied, AssetStealthCleared,
		// AssetMoved, XPSpent, XPAwarded, MovementOrderProgressed,
		// InfluenceDelta, TagAdded.
		return "", false
	}
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/overlay.go` (new)

The `Overlay` interface + `OverlayDoneMsg`, pinned in [`tui-turn-plan.md` → *The Overlay seam*]. Created here (Decision 10) so `execution.Model` (Task 3) can declare the `overlay.Overlay` field in one definition; Commit 7's placeholder implements this interface and the dispatch emits the `OverlayDoneMsg`. Full contents:

```go
package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

// Overlay is a modal prompt mounted over the execution view. Overlays never
// import the adapter channels and never see Reply: they render a prompt and
// emit OverlayDoneMsg{Answer} via a tea.Cmd; the execution sub-model owns the
// channel send.
type Overlay interface {
	Init() tea.Cmd
	Update(tea.Msg) (Overlay, tea.Cmd)
	View() string
	Help() help.KeyMap
}

// OverlayDoneMsg carries the overlay's answer back to the execution sub-model,
// which forwards it on the active ask's Reply channel. Cancel is deferred to
// Effort 3 (overview OQ 14); Effort 1 needs only the happy-path Answer.
type OverlayDoneMsg struct{ Answer any }
```

##### Task 3 — `internal/faction/tui/views/turn/execution/execution.go` (rewritten)

Full rewrite of the Commit 5 Task 3 minimal placeholder, preserving its `New(width, height int) Model` constructor and method set (`Init`/`Update`/`View`/`Help`/`CapturesInput`/`StatusLine`) so `turn.go` keeps compiling. Defines the full `Model` struct **including** the `overlay`/`pendingReply` fields — those stay nil in this commit (Commit 7's dispatch mounts them), and `CapturesInput` already reads `overlay != nil` (always false here) so Commit 7 need not touch it. The viewport is sized to the `Center` region width; follow-tail re-derives from `AtBottom()` after every scroll key, so any upward scroll (including PageUp) releases follow and returning to the bottom re-engages it. Full contents:

```go
package execution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/chrome"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/layout"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/turn/overlay"
)

type keyMap struct {
	Scroll key.Binding
	Follow key.Binding
}

// factionSnapshot is a copy of the acting faction's display fields, refreshed
// from each *domain.Faction event. The view never holds the engine pointer.
type factionSnapshot struct {
	name, scale            string
	force, cunning, wealth int
	curHP, maxHP, coin     int
}

// Model renders a running cycle as three panes: a left rail of the turn order,
// a center viewport streaming formatted events, and a right detail card for the
// acting faction. overlay/pendingReply are mounted by Commit 7's dispatch; here
// they stay nil (no reply path yet).
type Model struct {
	cycleNumber int
	order       []adapter.RailEntry
	currentID   string

	stream viewport.Model
	lines  []string
	follow bool

	detail factionSnapshot

	overlay      overlay.Overlay
	pendingReply chan<- any

	latestErr error
	fatal     bool

	keys          keyMap
	width, height int
}

func New(width, height int) Model {
	centerW := layout.RegionWidths(width)[layout.Center]
	return Model{
		stream: viewport.New(centerW, max(height, 1)),
		follow: true,
		keys: keyMap{
			Scroll: key.NewBinding(key.WithKeys("up", "down", "j", "k"), key.WithHelp("↑/↓", "scroll")),
			Follow: key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "follow")),
		},
		width:  width,
		height: height,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.stream.Width = layout.RegionWidths(msg.Width)[layout.Center]
		m.stream.Height = max(msg.Height, 1)
		m.stream.SetContent(strings.Join(m.lines, "\n"))
		if m.follow {
			m.stream.GotoBottom()
		}
		return m, nil

	case adapter.ObserverEventMsg:
		m.applyEvent(msg)
		m.lines = append(m.lines, renderEvent(msg))
		m.stream.SetContent(strings.Join(m.lines, "\n"))
		if m.follow {
			m.stream.GotoBottom()
		}
		return m, nil

	case adapter.EngineDoneMsg:
		if msg.Err != nil {
			m.latestErr = msg.Err
			m.fatal = true
		}
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Follow) {
			m.stream.GotoBottom()
			m.follow = true
			return m, nil
		}
		var cmd tea.Cmd
		m.stream, cmd = m.stream.Update(msg)
		m.follow = m.stream.AtBottom()
		return m, cmd
	}
	return m, nil
}

// applyEvent updates the rail/detail state captured from each event. Only the
// four *domain.Faction events carry a faction; mutation events touch neither
// the rail nor the card.
func (m *Model) applyEvent(msg adapter.ObserverEventMsg) {
	switch msg.Kind {
	case adapter.EvtCycleStarted:
		p := msg.Payload.(adapter.CycleStartedPayload)
		m.cycleNumber = p.CycleNumber
		m.order = p.Order
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
	case adapter.EvtFactionSkipped, adapter.EvtFactionTurnCompleted, adapter.EvtStatRaiseSkipped:
		m.detail = snapshotOf(msg.Payload.(*domain.Faction))
	case adapter.EvtError:
		m.latestErr = msg.Payload.(adapter.ErrorPayload).Err
	}
}

func snapshotOf(faction *domain.Faction) factionSnapshot {
	return factionSnapshot{
		name:    faction.Name,
		scale:   string(faction.Scale),
		force:   faction.Force,
		cunning: faction.Cunning,
		wealth:  faction.Wealth,
		curHP:   faction.CurrentHP,
		maxHP:   faction.MaxHP,
		coin:    faction.Coin,
	}
}

func (m Model) View() string {
	contentH := max(m.height, 1)
	panels := map[layout.Region]layout.Panel{
		layout.Left:   {Content: m.railView(), Style: lipgloss.NewStyle()},
		layout.Center: {Content: m.stream.View(), Style: lipgloss.NewStyle()},
		layout.Right:  {Content: m.detailView(), Style: lipgloss.NewStyle()},
	}
	return layout.Compose(panels, layout.RegionWidths(m.width), contentH)
}

func (m Model) railView() string {
	var b strings.Builder
	for _, entry := range m.order {
		if entry.ID == m.currentID {
			b.WriteString(styles.AccentAlt.Render("▶ "+entry.Name) + "\n")
		} else {
			b.WriteString(styles.Subtle.Render("  "+entry.Name) + "\n")
		}
	}
	return b.String()
}

func (m Model) detailView() string {
	if m.detail.name == "" {
		return styles.Dim.Render("—")
	}
	d := m.detail
	var b strings.Builder
	b.WriteString(styles.Strong.Render(d.name) + "\n")
	b.WriteString(styles.Subtle.Render(d.scale) + "\n\n")
	b.WriteString(styles.Subtle.Render("Force ") + styles.Strong.Render(strconv.Itoa(d.force)) + "\n")
	b.WriteString(styles.Subtle.Render("Cunning ") + styles.Strong.Render(strconv.Itoa(d.cunning)) + "\n")
	b.WriteString(styles.Subtle.Render("Wealth ") + styles.Strong.Render(strconv.Itoa(d.wealth)) + "\n\n")
	b.WriteString(styles.Subtle.Render("HP ") + styles.Strong.Render(fmt.Sprintf("%d/%d", d.curHP, d.maxHP)) + "\n")
	b.WriteString(styles.Subtle.Render("Coin ") + styles.Strong.Render(strconv.Itoa(d.coin)))
	return b.String()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }

func (h helpKeys) ShortHelp() []key.Binding  { return []key.Binding{h.keys.Scroll, h.keys.Follow} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.Scroll, h.keys.Follow}} }

func (m Model) CapturesInput() bool { return m.overlay != nil }

func (m Model) StatusLine() (string, chrome.Severity) {
	if m.latestErr == nil {
		return "", chrome.Info
	}
	if m.fatal {
		return m.latestErr.Error(), chrome.Fatal
	}
	return m.latestErr.Error(), chrome.Recoverable
}
```

##### Task 4 — verify

`go build ./...` && `go test ./internal/faction/tui/views/turn/...`. The formatter is unit-tested standalone: drive `renderEvent`/`formatMutation` against synthetic `adapter.ObserverEventMsg`s and each surfaced `domain.Mutation`, asserting header phrasing and the surface/fold split. A real cycle still can't complete (no reply path until Commit 7's dispatch); to eyeball the panes, feed the `Model` a scripted sequence of `ObserverEventMsg`s in a test and inspect `View()`. Confirm the rail advances on `EvtFactionTurnStarted`, the stream auto-follows, and the detail card reflects the acting faction.

> Note: Commits 6 and 7 are sequenced so the cycle can't run to completion until 7 lands. Commit 6 builds the rendering against synthetic events in tests; Commit 7 closes the loop.

##### Commit message
```
feat(tui/turn): execution view — three-pane + event stream + mutation formatter

- execution.Model: full three-pane view (turn-order rail / center viewport
  event stream / faction detail card), replacing Commit 5's placeholder;
  preserves the New(width,height)+method-set contract so turn.go is untouched
- format.go: renderEvent over all 14 observer events + EvtCycleStarted, and
  formatMutation surfacing player-meaningful deltas (HP/Coin/asset/base/stat/
  goal/homeworld/movement), folding bookkeeping noise
- follow-tail viewport: auto-scrolls on new events, releases on scroll-up,
  re-follows at the bottom or on G/end
- snapshot the acting faction's display fields per event (never hold the
  engine pointer); EvtError -> recoverable status line
- add the overlay interface package (Overlay + OverlayDoneMsg) so the Model
  holds the overlay field in one definition; Commit 7 adds placeholder + dispatch
```

#### Commit 7 — `feat(tui/turn): overlay seam + placeholder + dispatch switch`
*(suggest Opus — first real `askCh` round-trip)*

Closes the loop: the generic placeholder overlay and the dispatch that mounts it, drives its keypress, and forwards its answer on the ask's `Reply` channel. The `overlay` interface (`overlay/overlay.go`) and the `execution.Model.overlay`/`pendingReply` fields already exist from Commit 6 (Decision 10), and `CapturesInput` already reads `overlay != nil` — so this commit adds one new file and edits `execution.go` only; it never touches the struct or `turn.go`.

##### Task 1 — overlay interface (already created in Commit 6)

No work. `overlay/overlay.go` (`Overlay` + `OverlayDoneMsg`) lands in **Commit 6 Task 2** (Decision 10). This commit starts at the placeholder (Task 2) and the dispatch (Task 3).

##### Task 2 — `internal/faction/tui/views/turn/overlay/placeholder.go` (new)

The generic Effort-1 overlay. It implements `overlay.Overlay` (from Commit 6 Task 2), names the prompt, and on any keypress emits `OverlayDoneMsg` carrying the legal default for its `AskKind`. Each default is the exact type the phase collector asserts on (`phase_collector.go:35-81`): untyped `nil` → `SelectAction`'s skip (via the Commit 4 Task 3 nil branch → `orchestrator.go:418`); a typed nil `*domain.FactionStat` → decline raise; empty `[]world.MovementDecision{}` / `[]*domain.Asset{}` → no movement / no cargo; `true` → a non-error checkpoint ack (the value is discarded by `AwaitCheckpoint`). `AwaitCheckpoint` and `SelectTransportCargo` pass a **nil** `Faction` (`phase_collector.go:31,72`), so the placeholder renders kind-only when the name is empty. Full contents:

```go
package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Placeholder is Effort 1's stand-in for every real overlay. It names the
// prompt (kind + faction, when the ask carries one) and, on any keypress,
// emits OverlayDoneMsg carrying the kind's legal "no decision" default.
// Efforts 2-3 replace it kind-by-kind via newOverlay's factory switch.
type Placeholder struct {
	kind    adapter.AskKind
	faction string
	ack     key.Binding
}

func NewPlaceholder(kind adapter.AskKind, faction string) Placeholder {
	return Placeholder{
		kind:    kind,
		faction: faction,
		ack:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("any key", "continue")),
	}
}

func (p Placeholder) Init() tea.Cmd { return nil }

// Update acks on any key. The keypress requirement is deliberate: it paces the
// cycle so each prompt — including the final cycle_summary checkpoint — is a
// "done reading" beat before the engine proceeds.
func (p Placeholder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		answer := defaultFor(p.kind)
		return p, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return p, nil
}

func (p Placeholder) View() string {
	label := kindLabel(p.kind)
	if p.faction != "" {
		label += " · " + p.faction
	}
	return styles.WarnPrompt.Render(label + " · press any key")
}

func (p Placeholder) Help() help.KeyMap { return placeholderHelp{p.ack} }

type placeholderHelp struct{ ack key.Binding }

func (h placeholderHelp) ShortHelp() []key.Binding  { return []key.Binding{h.ack} }
func (h placeholderHelp) FullHelp() [][]key.Binding { return [][]key.Binding{{h.ack}} }

// defaultFor returns the legal "no decision" reply for each ask kind. Every
// value matches the type the phase collector asserts on.
func defaultFor(kind adapter.AskKind) any {
	switch kind {
	case adapter.AskAwaitCheckpoint:
		return true
	case adapter.AskSelectAction:
		return nil
	case adapter.AskSelectStatRaise:
		return (*domain.FactionStat)(nil)
	case adapter.AskSelectMovementDecisions:
		return []world.MovementDecision{}
	case adapter.AskSelectTransportCargo:
		return []*domain.Asset{}
	default:
		return nil
	}
}

func kindLabel(kind adapter.AskKind) string {
	switch kind {
	case adapter.AskAwaitCheckpoint:
		return "Checkpoint"
	case adapter.AskSelectAction:
		return "Select Action"
	case adapter.AskSelectStatRaise:
		return "Select Stat Raise"
	case adapter.AskSelectMovementDecisions:
		return "Select Movement"
	case adapter.AskSelectTransportCargo:
		return "Select Transport Cargo"
	default:
		return "Prompt"
	}
}

var _ Overlay = Placeholder{}
```

##### Task 3 — `internal/faction/tui/views/turn/execution/execution.go`

Three edits to the Commit 6 Task 3 file: mount/route/answer the overlay in `Update`, replace the panes with the centered prompt in `View` (Decision 11), and add the `newOverlay` factory. No import changes — `overlay`, `adapter`, and `lipgloss` are already imported by Commit 6.

In `Update`, add the dispatch cases and route keys to the overlay while one is mounted.

**Find:**
```go
	case adapter.EngineDoneMsg:
		if msg.Err != nil {
			m.latestErr = msg.Err
			m.fatal = true
		}
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Follow) {
			m.stream.GotoBottom()
			m.follow = true
			return m, nil
		}
		var cmd tea.Cmd
		m.stream, cmd = m.stream.Update(msg)
		m.follow = m.stream.AtBottom()
		return m, cmd
	}
	return m, nil
}
```
**Replace with:**
```go
	case adapter.EngineDoneMsg:
		if msg.Err != nil {
			m.latestErr = msg.Err
			m.fatal = true
		}
		return m, nil

	case adapter.CollectorAskMsg:
		m.pendingReply = msg.Reply
		m.overlay = newOverlay(msg)
		return m, m.overlay.Init()

	case overlay.OverlayDoneMsg:
		if m.pendingReply != nil {
			m.pendingReply <- msg.Answer // cap-1 buffered (phase_collector.go:18), non-blocking
			m.pendingReply = nil
		}
		m.overlay = nil
		return m, nil

	case tea.KeyMsg:
		if m.overlay != nil {
			var cmd tea.Cmd
			m.overlay, cmd = m.overlay.Update(msg)
			return m, cmd
		}
		if key.Matches(msg, m.keys.Follow) {
			m.stream.GotoBottom()
			m.follow = true
			return m, nil
		}
		var cmd tea.Cmd
		m.stream, cmd = m.stream.Update(msg)
		m.follow = m.stream.AtBottom()
		return m, cmd
	}
	return m, nil
}
```

In `View`, render the centered prompt instead of the panes while an overlay is up.

**Find:**
```go
func (m Model) View() string {
	contentH := max(m.height, 1)
	panels := map[layout.Region]layout.Panel{
```
**Replace with:**
```go
func (m Model) View() string {
	contentH := max(m.height, 1)
	if m.overlay != nil {
		return lipgloss.Place(m.width, contentH, lipgloss.Center, lipgloss.Center, m.overlay.View())
	}
	panels := map[layout.Region]layout.Panel{
```

Add the `newOverlay` factory after `StatusLine` (end of file).

**Find:**
```go
func (m Model) StatusLine() (string, chrome.Severity) {
	if m.latestErr == nil {
		return "", chrome.Info
	}
	if m.fatal {
		return m.latestErr.Error(), chrome.Fatal
	}
	return m.latestErr.Error(), chrome.Recoverable
}
```
**Replace with:**
```go
func (m Model) StatusLine() (string, chrome.Severity) {
	if m.latestErr == nil {
		return "", chrome.Info
	}
	if m.fatal {
		return m.latestErr.Error(), chrome.Fatal
	}
	return m.latestErr.Error(), chrome.Recoverable
}

// newOverlay builds the overlay for an ask. One arm per AskKind so Efforts 2-3
// swap real overlays in kind-by-kind; Effort 1 returns the generic placeholder
// for every kind. A nil Faction (AwaitCheckpoint, SelectTransportCargo) yields
// an empty name, which the placeholder renders kind-only.
func newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay {
	name := ""
	if msg.Faction != nil {
		name = msg.Faction.Name
	}
	switch msg.Kind {
	case adapter.AskAwaitCheckpoint:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectAction:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectStatRaise:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectMovementDecisions:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectTransportCargo:
		return overlay.NewPlaceholder(msg.Kind, name)
	default:
		return overlay.NewPlaceholder(msg.Kind, name)
	}
}
```

##### Task 4 — verify (end-to-end)

`go build ./... && go test ./...`. Then launch Turn → Start → keypress through each prompt → a real cycle streams events to completion with every decision defaulted → final `cycle_summary` keypress → `EngineDoneMsg` → return to setup with post-cycle HP/Coin on the roster. Confirm the `askCh` round-trip (block-send → overlay mount → keypress → reply) works and the engine goroutine never deadlocks or panics on channel close (sole-writer `defer close`, Decision 1). Confirm the return-to-setup transition survives the teardown race — the final `cycle_summary` ack flushes `OnFactionTurnCompleted`/`OnCycleCompleted` into the buffered `eventCh` just before `EngineDoneMsg`, so re-run the cycle a few times to exercise the ordering and confirm the nil-guarded re-issue arms (Commit 5 Task 4) never panic on a late `ObserverEventMsg`. Confirm the root yields `q`/`tab`/`?` to the prompt while it's up (`CapturesInput` → `overlay != nil`).

##### Commit message
```
feat(tui/turn): overlay seam + placeholder + dispatch switch

- overlay/placeholder.go: generic Effort-1 overlay — names the prompt
  (kind + faction) and acks on any key, emitting the legal default per
  AskKind (checkpoint ack / skip action / decline raise / no movement /
  empty cargo), each matching the phase collector's reply assertion
- execution dispatch: on CollectorAskMsg store Reply + mount the overlay via
  newOverlay's per-kind factory; route keys to the overlay while mounted; on
  OverlayDoneMsg send Answer on the cap-1 Reply channel and clear both
- View replaces the panes with the centered prompt while an overlay is up
  (matches the app's modal pattern; over-panes compositing deferred to E2)
- closes the askCh round-trip: Turn now runs a real cycle end-to-end with
  every decision auto-defaulted, then returns to setup
```
