# TUI Turn — Effort 2 Plan (Phases except action)

`F-005.3`, Effort 2 of 3 on `feature/tui-turn`. Swap Effort 1's generic placeholder for real `huh`-backed overlays on the four **non-action** phase kinds, and make the right faction-detail pane phase-reactive. Purely additive on Effort 1's seam — the `Overlay` interface, the `OverlayDoneMsg` round-trip, the ask-pump, and the adapter contract are all untouched.

- Overview: [`tui-turn-plan.md`](./tui-turn-plan.md) (routes OQ 8/9/10 here)
- Discovery: [`tui-turn-discovery.md`](../discovery/tui-turn-discovery.md) — Decision 13 + *Effort 2 modals*
- Effort 1 (landed): [`completed/tui-turn-effort-1-plan.md`](./completed/tui-turn-effort-1-plan.md)

## Context / Goal

Effort 1 proved the whole cycle against the real engine with one generic placeholder backing every `AskKind`. Effort 2 fills in real modals for the four non-action kinds:

| Kind | Overlay | Answer type |
|------|---------|-------------|
| `AskAwaitCheckpoint` | `Checkpoint` — bare cycle-complete ack | `true` (discarded) |
| `AskSelectStatRaise` | `StatRaise` — single-select + Decline | `*domain.FactionStat` |
| `AskSelectMovementDecisions` | `Movement` — sequential per-asset order form | `[]world.MovementDecision` |
| `AskSelectTransportCargo` | `TransportCargo` — capped multi-select | `[]*domain.Asset` |

`AskSelectAction` stays on the placeholder (auto-skip → `nil`); it becomes real in Effort 3. The work is filling the four arms of `newOverlay` (`execution.go:240`) and rendering the detail pane by phase — no change to the dispatch flow, the pumps, or the adapter.

**Ground truth re-verified against current source (2026-06-03):** the `Overlay` interface (`overlay/overlay.go`), the `newOverlay` factory + ask/overlay dispatch (`execution.go:104-115, 236-259`), the ask payloads (`channels.go:65-75`), the `huh` wizard idiom (`wizard.go`), and the movement engine's Issue/Revise hex paths (`movement.go:82-91, 124`).

## Decisions Ratified in Planning

1. **Overlays stay concrete-per-modal; no archetype toolkit in Effort 2.** Each of the four overlays is its own file following the wizard idiom (`huh.Form` + heap-allocated pointer-bound data + `StateCompleted` → emit). The generic six-archetype toolkit is **Effort 3's** job (Discovery Decision 14); building it now would front-load Effort 3's design. Minor boilerplate duplication across the four files is accepted (`feedback_deferred_refactor` — match the existing concrete pattern, generalize later). *(Resolves no OQ; sets the build style.)*

2. **`rulebook` + `spatial.RegionMap` are threaded into Turn, mirroring Manage.** The movement modal needs world options + hex resolution (OQ 9) and asset names; the cargo modal needs asset names; the movement detail pane needs both. The root (`model.go:48-49`) already holds `rb` and `spatialMap` and hands them to `manage.New` — Effort 2 threads the same two values into `turn.New → execution.New → the overlay constructors`. No adapter-payload changes; the engine pointer never crosses into the view. *(Enables OQ 8/9 and the movement pane.)*

3. **OQ 8 — `SelectMovementDecisions` is a sequential per-asset form (Robert's call).** `huh` v1.0.0 exposes `WithHideFunc` **only on `Group`**, not on individual fields (`Field` has a read-only `Skip()`, no settable per-field hide). The Discovery sketch ("one group per asset with a conditionally-hidden destination *field*") is therefore not buildable. The ratified structure: **two groups per eligible asset** — a kind `Select` (`{Hold, Issue, Revise, Cancel}`), then a destination-world `Select` carrying `.WithHideFunc(...)` that hides the whole group unless the row's kind is Issue/Revise. `huh` pages one group at a time and skips hidden groups on navigation, so the GM walks asset-by-asset and only sees a destination page when one is needed.

4. **OQ 9 — the movement modal sets both `WorldID` and `RegionHex` on every Issue/Revise `Destination`.** Verified asymmetry: Issue (`movement.go:82-91`) derives the hex from `WorldID` via the spatial map and ignores the passed `RegionHex`; Revise (`movement.go:124`) trusts the passed `RegionHex` directly. The modal resolves the chosen world's hex exactly as `wizard.synthesize` does (`wizard.go:136-142`: `Location(worldID)` → `RegionLocation` → `RegionHex()`) and populates both fields, satisfying both engine paths. No engine change.

5. **OQ 10 — the phase signal is derived from the observer event stream (Robert's call).** The execution model holds a `phase` field advanced by phase-boundary events. Because no "phase started" event exists, phase is set to the *next* interactive phase on the *prior* phase's completion event, so the pane is already correct when that phase's modal opens (e.g. `EvtFactionTurnStarted → phaseStatRaise`; `EvtStatRaiseApplied/Skipped/BookkeepingApplied → phaseMovement`). The pane's **stat-raise** enrichment renders from the existing race-safe `factionSnapshot` scalars; its **movement** enrichment renders from a `movables` summary stashed from the `SelectMovementDecisions` ask payload (the only race-free source — the engine is parked on `<-reply` when the ask is delivered; iterating the live `faction.Assets` map at event time would race the engine goroutine).

6. **Effort-2 overlays swallow `Esc`.** Cancel routing (`ErrTurnCanceled`) lands in Effort 3 (overview OQ 14). Until then, every `huh`-backed overlay intercepts and drops `Esc` so the form cannot enter `StateAborted` (which the `StateCompleted`-only check would leave mounted and wedged). This matches the placeholder's "you must answer to proceed" pacing.

7. **Commit breakdown is three commits (Robert's call).** (1) checkpoint + stat-raise; (2) movement + transport-cargo + dep threading; (3) phase-reactive detail pane. Each is one execution session; Commit 2 carries the OQ-8/9 design weight.

## Open Questions — To Ratify at Implementation Time

- **`huh` group-hide on back-navigation (Commit 2).** Forward flow is confirmed (the destination group's `HideFunc` reads the committed kind value when navigation reaches it). The back-then-change-kind edge (pick Issue, see destination, go back, switch to Hold) relies on `huh` re-evaluating group visibility on re-navigation. Validate this interactively while building the form; if it misbehaves, the fallback is to re-assert the kind→destination consistency in the submit assembly (already done — `decisions()` only emits a destination for Issue/Revise rows, so a stale hidden-group value is ignored regardless).
- **Stat-raise default selection (Commit 1).** Options are ordered eligible-stats-first, Decline last, so the default highlighted option is the first eligible stat. Executor may flip to Decline-first if "no change is the safe default" reads better; cosmetic, not load-bearing.

## Shared Context

### Model discipline (per commit)

| Commit | Model | Why |
|--------|-------|-----|
| 1 — checkpoint + stat-raise | Sonnet | Two small overlays on the established idiom |
| 2 — movement + cargo + threading | **Opus** | OQ-8 conditional form + OQ-9 hex resolution carry the design weight (overview model table) |
| 3 — phase-reactive detail pane | Sonnet | Event-driven phase field + render branches |

Prompt Robert to `/model` at each commit's session start, and again at the shift into the branch-level pre-merge checklist after Effort 3.

### The overlay-construction context seam

Effort 1's `newOverlay` is a package function taking only the `CollectorAskMsg`. The four real overlays need context the ask doesn't carry (cycle number, rulebook, spatial map), so **Commit 1 converts `newOverlay` into a method** `func (m Model) newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay`, reading `m.cycleNumber` / `m.rulebook` / `m.spatialMap`. The dispatch flow in `Update` is unchanged except the call site becomes `m.newOverlay(msg)`. This is the only structural change to the seam, and it is additive — the switch arms and the `CollectorAskMsg`/`OverlayDoneMsg` round-trip stay identical.

### What Effort 2 does NOT touch

The `adapter` package (channels, payloads, pumps, `Run`), the `turn.Model` router dispatch, the `OverlayDoneMsg`/`pendingReply` round-trip, and the `AskSelectAction` arm (Effort 3). The seam holds; this is filling-in.

## Out of Scope

- **`Esc`-cancel UX / `ErrTurnCanceled` routing** — Effort 3 (Decision 6 defers it; overlays swallow `Esc` until then).
- **The real `SelectAction` modal + the archetype toolkit** — Effort 3 (Decision 1).
- **Action-phase detail enrichment** — Effort 3; Effort 2's pane falls through to the base card for `phaseNone`.
- **Cargo-to-order attachment** — orchestrator-side. The movement form leaves `MovementDecision.CargoAssetIDs` empty; the engine collects cargo via the separate `SelectTransportCargo` ask (Discovery: "Cargo is not part of this form").
- **Deferred consolidation (capture at pre-merge, not now):** `worldsToOptions` + world→hex resolution now exist in both `wizard` and the movement overlay; a shared `tui` spatial-options helper is a future refactor. Per `feedback_deferred_refactor`, Effort 2 matches the existing duplicated pattern rather than extracting mid-feature.

---

## Work Breakdown

> **Re-grounding (every execution session):** re-read the target file(s) before editing — Effort-2 commits land in sequence and Commit 1 reshapes `newOverlay`, which Commits 2–3 then extend. Verify the anchors below against current source first.

### Commit 1 — `feat(tui/turn): real checkpoint and stat-raise overlays`

Swap the placeholder for real overlays on `AskAwaitCheckpoint` and `AskSelectStatRaise`. Introduces the shared form-help keymap and the `newOverlay`-as-method seam. No new deps (stat-raise reads the faction from the ask; checkpoint reads the cycle number off the model).

##### Task 1 — `internal/faction/tui/views/turn/overlay/formhelp.go` (new) — full contents:

```go
package overlay

import "github.com/charmbracelet/bubbles/key"

// formHelp is the keymap shared by every huh-backed overlay. Esc is intentionally
// absent: cancel routing lands in Effort 3, and Effort-2 overlays swallow Esc.
type formHelp struct{}

func (formHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}

func (formHelp) FullHelp() [][]key.Binding { return [][]key.Binding{formHelp{}.ShortHelp()} }
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/checkpoint.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// Checkpoint is the end-of-cycle wrap. In whole-cycle mode the only checkpoint
// that reaches the TUI is cycle_summary (phase_collector.go:28), so this always
// renders the cycle-complete ack. Answer is discarded by AwaitCheckpoint.
type Checkpoint struct {
	cycle int
	ack   key.Binding
}

func NewCheckpoint(cycle int) Checkpoint {
	return Checkpoint{
		cycle: cycle,
		ack:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "return to setup")),
	}
}

func (c Checkpoint) Init() tea.Cmd { return nil }

func (c Checkpoint) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, c.ack) {
		return c, func() tea.Msg { return OverlayDoneMsg{Answer: true} }
	}
	return c, nil
}

func (c Checkpoint) View() string {
	return styles.WarnPrompt.Render(fmt.Sprintf("Cycle %d complete — press Enter to return to setup", c.cycle))
}

func (c Checkpoint) Help() help.KeyMap { return checkpointHelp{c.ack} }

type checkpointHelp struct{ ack key.Binding }

func (h checkpointHelp) ShortHelp() []key.Binding  { return []key.Binding{h.ack} }
func (h checkpointHelp) FullHelp() [][]key.Binding { return [][]key.Binding{{h.ack}} }

var _ Overlay = Checkpoint{}
```

##### Task 3 — `internal/faction/tui/views/turn/overlay/statraise.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// statRaiseData is heap-allocated and pointer-bound to the huh field so the
// value-copy of the overlay (bubbletea semantics) never orphans the binding.
type statRaiseData struct{ choice *domain.FactionStat }

// StatRaise is a single-select over the eligible stats plus an explicit Decline
// sentinel (mapping to a nil *FactionStat). Decline is in-modal, never Esc.
type StatRaise struct {
	data *statRaiseData
	form *huh.Form
}

func NewStatRaise(faction *domain.Faction, eligible []domain.FactionStat) StatRaise {
	data := &statRaiseData{}
	ratingOf := map[domain.FactionStat]int{
		domain.StatForce:   faction.Force,
		domain.StatCunning: faction.Cunning,
		domain.StatWealth:  faction.Wealth,
	}
	opts := make([]huh.Option[*domain.FactionStat], 0, len(eligible)+1)
	for i := range eligible {
		stat := &eligible[i]
		opts = append(opts, huh.NewOption(
			fmt.Sprintf("%s (currently %d)", eligible[i], ratingOf[eligible[i]]),
			stat,
		))
	}
	opts = append(opts, huh.NewOption("Decline raise", (*domain.FactionStat)(nil)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*domain.FactionStat]().
				Title("Raise a stat").
				Options(opts...).
				Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return StatRaise{data: data, form: form}
}

func (o StatRaise) Init() tea.Cmd { return o.form.Init() }

func (o StatRaise) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil // cancel routing is Effort 3; swallow Esc so the form can't abort
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o StatRaise) View() string { return o.form.View() }

func (o StatRaise) Help() help.KeyMap { return formHelp{} }

var _ Overlay = StatRaise{}
```

##### Task 4 — `internal/faction/tui/views/turn/execution/execution.go` — add the cycle-number field, capture it, and convert `newOverlay` to a method swapping the two arms.

(a) Add `cycleNumber` to the model. **Find:**
```go
	detail factionSnapshot

	overlay      overlay.Overlay
	pendingReply chan<- any
```
**Replace with:**
```go
	detail      factionSnapshot
	cycleNumber int

	overlay      overlay.Overlay
	pendingReply chan<- any
```

(b) Capture the cycle number off `EvtCycleStarted`. **Find:**
```go
	case adapter.EvtCycleStarted:
		m.order = msg.Payload.(adapter.CycleStartedPayload).Order
```
**Replace with:**
```go
	case adapter.EvtCycleStarted:
		payload := msg.Payload.(adapter.CycleStartedPayload)
		m.order = payload.Order
		m.cycleNumber = payload.CycleNumber
```

(c) Update the dispatch call site. **Find:**
```go
		m.pendingReply = msg.Reply
		m.overlay = newOverlay(msg)
		return m, m.overlay.Init()
```
**Replace with:**
```go
		m.pendingReply = msg.Reply
		m.overlay = m.newOverlay(msg)
		return m, m.overlay.Init()
```

(d) Replace the factory. **Find:** the entire `func newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay { ... }` block (`execution.go:236-259`). **Replace with:**
```go
// newOverlay builds the overlay for an ask. One arm per AskKind so Efforts 2-3
// swap real overlays in kind-by-kind; the AskSelectAction arm stays on the
// placeholder until Effort 3. A method so it can supply the cycle number,
// rulebook, and spatial map the real overlays need but the ask does not carry.
func (m Model) newOverlay(msg adapter.CollectorAskMsg) overlay.Overlay {
	name := ""
	if msg.Faction != nil {
		name = msg.Faction.Name
	}
	switch msg.Kind {
	case adapter.AskAwaitCheckpoint:
		return overlay.NewCheckpoint(m.cycleNumber)
	case adapter.AskSelectStatRaise:
		return overlay.NewStatRaise(msg.Faction, msg.Payload.(adapter.SelectStatRaisePayload).Eligible)
	case adapter.AskSelectAction:
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

##### Commit message
```
feat(tui/turn): real checkpoint and stat-raise overlays

- add Checkpoint overlay (cycle-complete ack) and StatRaise overlay
  (single-select + Decline), swapping their newOverlay arms
- introduce shared formHelp keymap for huh-backed overlays
- convert newOverlay to a method so it can supply cycle/rulebook/spatial
  context; capture cycleNumber off EvtCycleStarted
```

---

### Commit 2 — `feat(tui/turn): movement and transport-cargo overlays`

The design-heavy commit (**Opus**). Threads `rulebook` + `spatialMap` into Turn, builds the sequential movement form (OQ 8) with both-field destination resolution (OQ 9), and the capped cargo multi-select. Swaps the last two non-action arms.

##### Task 1 — `internal/faction/tui/views/turn/overlay/movement.go` (new) — full contents:

```go
package overlay

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type orderKind int

const (
	orderNone orderKind = iota
	orderIssue
	orderRevise
	orderCancel
)

// assetRow is one eligible asset's chosen order. Bound by pointer into the
// heap-allocated movementData.rows slice (allocated once at full length so the
// addresses stay stable for huh's bindings).
type assetRow struct {
	assetID string
	kind    orderKind
	worldID string
}

type movementData struct{ rows []assetRow }

// Movement is the sequential per-asset order form: two groups per asset (a kind
// select, then a destination select gated by group HideFunc on the row's kind).
// huh v1.0.0 supports HideFunc only at group granularity (OQ 8), so the
// destination is its own gated group, not a hidden field.
type Movement struct {
	data       *movementData
	form       *huh.Form
	spatialMap *spatial.RegionMap
}

func NewMovement(eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Movement {
	data := &movementData{rows: make([]assetRow, len(eligible))}
	worldOpts := worldOptions(spatialMap)
	names := worldNames(spatialMap)

	groups := make([]*huh.Group, 0, len(eligible)*2)
	for i, asset := range eligible {
		idx := i
		data.rows[idx].assetID = asset.ID

		loc := names[asset.Location.WorldID]
		if loc == "" {
			loc = asset.Location.WorldID
		}

		kindOpts := []huh.Option[orderKind]{
			huh.NewOption("Hold position", orderNone),
			huh.NewOption("Issue move", orderIssue),
		}
		if asset.CurrentOrder != nil { // Revise/Cancel only for an asset with a live order
			kindOpts = append(kindOpts,
				huh.NewOption("Revise move", orderRevise),
				huh.NewOption("Cancel move", orderCancel),
			)
		}

		kindGroup := huh.NewGroup(
			huh.NewSelect[orderKind]().
				Title(fmt.Sprintf("Asset %d of %d: %s (at %s)", idx+1, len(eligible), assetName(asset, rb), loc)).
				Options(kindOpts...).
				Value(&data.rows[idx].kind),
		)

		destGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s → destination", assetName(asset, rb))).
				Options(worldOpts...).
				Value(&data.rows[idx].worldID),
		).WithHideFunc(func() bool {
			k := data.rows[idx].kind
			return k != orderIssue && k != orderRevise
		})

		groups = append(groups, kindGroup, destGroup)
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return Movement{data: data, form: form, spatialMap: spatialMap}
}

func (o Movement) Init() tea.Cmd { return o.form.Init() }

func (o Movement) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil // cancel routing is Effort 3
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		decisions := o.decisions()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: decisions} }
	}
	return o, cmd
}

func (o Movement) View() string { return o.form.View() }

func (o Movement) Help() help.KeyMap { return formHelp{} }

// decisions assembles the engine slice. Issue/Revise set BOTH WorldID and
// RegionHex on Destination: Issue derives the hex from WorldID (movement.go:91)
// and Revise trusts RegionHex (movement.go:124), so both are populated (OQ 9).
func (o Movement) decisions() []world.MovementDecision {
	var out []world.MovementDecision
	for _, r := range o.data.rows {
		switch r.kind {
		case orderNone:
			continue
		case orderCancel:
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: world.MovementDecisionCancel})
		case orderIssue, orderRevise:
			dest := &domain.Location{WorldID: r.worldID}
			if loc, ok := o.spatialMap.Location(r.worldID); ok {
				if rl, ok := loc.(spatial.RegionLocation); ok {
					dest.RegionHex = rl.RegionHex()
				}
			}
			kind := world.MovementDecisionIssue
			if r.kind == orderRevise {
				kind = world.MovementDecisionRevise
			}
			out = append(out, world.MovementDecision{AssetID: r.assetID, Kind: kind, Destination: dest})
		}
	}
	return out
}

// --- shared overlay helpers (also used by cargo.go) ---

func assetName(asset *domain.Asset, rb *rulebook.Rulebook) string {
	if def, ok := rb.Assets[asset.DefinitionID]; ok {
		return def.Name
	}
	return asset.DefinitionID
}

func worldOptions(spatialMap *spatial.RegionMap) []huh.Option[string] {
	worlds := spatialMap.AllWorlds()
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].Name() < worlds[j].Name() })
	opts := make([]huh.Option[string], len(worlds))
	for i, w := range worlds {
		opts[i] = huh.NewOption(fmt.Sprintf("%s (TL %d)", w.Name(), w.TechLevel()), w.ID())
	}
	return opts
}

func worldNames(spatialMap *spatial.RegionMap) map[string]string {
	out := make(map[string]string)
	for _, w := range spatialMap.AllWorlds() {
		out[w.ID()] = w.Name()
	}
	return out
}

var _ Overlay = Movement{}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/cargo.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type cargoData struct{ chosen []*domain.Asset }

// TransportCargo is a capped multi-select fired once per transport that issued a
// move. The orchestrator re-validates the cap (orchestrator.go:672); in-modal
// Validate is for UX. Answer: the chosen assets (nil/empty = no cargo).
type TransportCargo struct {
	data *cargoData
	form *huh.Form
}

func NewTransportCargo(transport *domain.Asset, eligible []*domain.Asset, profile *domain.TransportProfile, rb *rulebook.Rulebook) TransportCargo {
	data := &cargoData{}
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(assetName(asset, rb), asset))
	}
	maxCargo := profile.MaxCargo

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[*domain.Asset]().
				Title(fmt.Sprintf("Load cargo onto %s (up to %d)", assetName(transport, rb), maxCargo)).
				Options(opts...).
				Value(&data.chosen).
				Validate(func(picked []*domain.Asset) error {
					if len(picked) > maxCargo {
						return fmt.Errorf("at most %d cargo", maxCargo)
					}
					return nil
				}),
		),
	).WithTheme(styles.FormTheme())
	return TransportCargo{data: data, form: form}
}

func (o TransportCargo) Init() tea.Cmd { return o.form.Init() }

func (o TransportCargo) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return o, nil // cancel routing is Effort 3
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.chosen
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o TransportCargo) View() string { return o.form.View() }

func (o TransportCargo) Help() help.KeyMap { return formHelp{} }

var _ Overlay = TransportCargo{}
```

##### Task 3 — `internal/faction/tui/views/turn/execution/execution.go` — hold the two new deps and swap the movement/cargo arms.

(a) Add the imports (`rulebook`, `spatial`). **Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
```
and, in the same import block, after the `views/turn/overlay` line add:
```go
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```

(b) Add the fields. **Find:**
```go
	keys          keyMap
	width, height int
}
```
**Replace with:**
```go
	rulebook   *rulebook.Rulebook
	spatialMap *spatial.RegionMap

	keys          keyMap
	width, height int
}
```

(c) Extend the constructor. **Find:**
```go
func New(width, height int) Model {
	centerW := layout.RegionWidths(width)[layout.Center]
	return Model{
		stream: viewport.New(centerW, max(height, 1)),
		follow: true,
```
**Replace with:**
```go
func New(width, height int, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) Model {
	centerW := layout.RegionWidths(width)[layout.Center]
	return Model{
		rulebook:   rb,
		spatialMap: spatialMap,
		stream:     viewport.New(centerW, max(height, 1)),
		follow:     true,
```

(d) Swap the two arms in `newOverlay`. **Find:**
```go
	case adapter.AskSelectMovementDecisions:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectTransportCargo:
		return overlay.NewPlaceholder(msg.Kind, name)
```
**Replace with:**
```go
	case adapter.AskSelectMovementDecisions:
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
	case adapter.AskSelectTransportCargo:
		p := msg.Payload.(adapter.SelectTransportCargoPayload)
		return overlay.NewTransportCargo(p.Transport, p.EligibleCargo, p.Profile, m.rulebook)
```

##### Task 4 — `internal/faction/tui/views/turn/turn.go` — thread the deps through the router.

(a) Imports. **Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
```
and add, alongside the other internal imports:
```go
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
```

(b) Fields. **Find:**
```go
	factionState *state.FactionState
	paths        *campaigns.Paths
	eng          *engine.Engine
	log          *slog.Logger
```
**Replace with:**
```go
	factionState *state.FactionState
	paths        *campaigns.Paths
	eng          *engine.Engine
	rulebook     *rulebook.Rulebook
	spatialMap   *spatial.RegionMap
	log          *slog.Logger
```

(c) Constructor. **Find:**
```go
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
```
**Replace with:**
```go
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
```

(d) Pass them into `execution.New`. **Find:**
```go
		m.execution = execution.New(m.termWidth, m.termHeight)
```
**Replace with:**
```go
		m.execution = execution.New(m.termWidth, m.termHeight, m.rulebook, m.spatialMap)
```

##### Task 5 — `internal/faction/tui/model.go` — pass `rb`/`spatialMap` to `turn.New`. **Find:**
```go
			modebar.ModeTurn:   turn.New(factionState, paths, eng, log),
```
**Replace with:**
```go
			modebar.ModeTurn:   turn.New(factionState, paths, eng, rb, spatialMap, log),
```

##### Commit message
```
feat(tui/turn): movement and transport-cargo overlays

- add Movement overlay: sequential per-asset order form, destination as a
  HideFunc-gated group (huh has no per-field hide); Issue/Revise set both
  WorldID and resolved RegionHex
- add TransportCargo overlay: capped multi-select per transport
- thread rulebook + spatialMap through turn.New -> execution.New into the
  overlay constructors (mirrors manage.New); swap the two arms
```

---

### Commit 3 — `feat(tui/turn): phase-reactive faction detail pane`

The right pane reflects the active phase, derived from the observer event stream (OQ 10). Stat-raise emphasizes the three ratings (from the existing snapshot); movement lists movable assets (from a summary stashed off the movement ask — the race-free source).

##### Task 1 — `internal/faction/tui/views/turn/execution/execution.go` — phase field, transitions, stash, and the phase-reactive `detailView`.

(a) Add the phase type and the model fields. **Find:**
```go
	detail      factionSnapshot
	cycleNumber int
```
**Replace with:**
```go
	detail      factionSnapshot
	cycleNumber int
	phase       turnPhase
	movables    []movableLine
```
and add, directly above the `factionSnapshot` type declaration:
```go
// turnPhase is the active phase, derived from phase-boundary events. There is no
// "phase started" event, so it is set to the NEXT interactive phase on the prior
// phase's completion event — correct by the time that phase's modal opens.
// (phaseNone covers goal-lock, bookkeeping, action, and turn wrap-up — base card.)
type turnPhase int

const (
	phaseNone turnPhase = iota
	phaseStatRaise
	phaseMovement
)

// movableLine is a resolved, race-safe summary of one eligible asset for the
// movement detail pane. Built from the SelectMovementDecisions ask payload while
// the engine is parked on the reply; never read from the live faction map.
type movableLine struct{ name, location, order string }
```

(b) Stash the movables when the movement ask arrives. **Find:**
```go
	case adapter.CollectorAskMsg:
		m.pendingReply = msg.Reply
		m.overlay = m.newOverlay(msg)
		return m, m.overlay.Init()
```
**Replace with:**
```go
	case adapter.CollectorAskMsg:
		if msg.Kind == adapter.AskSelectMovementDecisions {
			m.movables = resolveMovables(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
		}
		m.pendingReply = msg.Reply
		m.overlay = m.newOverlay(msg)
		return m, m.overlay.Init()
```

(c) Advance the phase on the boundary events. **Find:**
```go
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
		if !m.fatal {
			m.latestErr = nil // a recoverable error scopes to the faction it occurred under
		}
	case adapter.EvtFactionSkipped, adapter.EvtFactionTurnCompleted, adapter.EvtStatRaiseSkipped:
		m.detail = snapshotOf(msg.Payload.(*domain.Faction))
```
**Replace with:**
```go
	case adapter.EvtFactionTurnStarted:
		faction := msg.Payload.(*domain.Faction)
		m.currentID = faction.ID
		m.detail = snapshotOf(faction)
		m.phase = phaseStatRaise // next interactive phase; correct when its modal opens
		m.movables = nil
		if !m.fatal {
			m.latestErr = nil // a recoverable error scopes to the faction it occurred under
		}
	case adapter.EvtStatRaiseSkipped:
		m.detail = snapshotOf(msg.Payload.(*domain.Faction))
		m.phase = phaseMovement
	case adapter.EvtStatRaiseApplied, adapter.EvtBookkeepingApplied:
		m.phase = phaseMovement
	case adapter.EvtMovementTicked, adapter.EvtMovementResolved:
		m.phase = phaseNone // action enrichment is Effort 3
	case adapter.EvtFactionSkipped, adapter.EvtFactionTurnCompleted:
		m.detail = snapshotOf(msg.Payload.(*domain.Faction))
		m.phase = phaseNone
```

(d) Make `detailView` phase-reactive and add the helpers. **Find:** the entire current `func (m Model) detailView() string { ... }` block (`execution.go:201-215`). **Replace with:**
```go
func (m Model) detailView() string {
	if m.detail.name == "" {
		return styles.Dim.Render("—")
	}
	switch m.phase {
	case phaseStatRaise:
		return m.statRaiseDetail()
	case phaseMovement:
		return m.movementDetail()
	default:
		return m.baseDetail()
	}
}

func (m Model) baseDetail() string {
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

// statRaiseDetail emphasizes the three ratings (the raise targets) over the base
// card. Renders from the race-safe snapshot scalars.
func (m Model) statRaiseDetail() string {
	d := m.detail
	var b strings.Builder
	b.WriteString(styles.Strong.Render(d.name) + "\n")
	b.WriteString(styles.Subtle.Render(d.scale) + "\n\n")
	b.WriteString(styles.Subtle.Render("Raise a stat") + "\n")
	b.WriteString(styles.Subtle.Render("Force ") + styles.AccentAlt.Render(strconv.Itoa(d.force)) + "\n")
	b.WriteString(styles.Subtle.Render("Cunning ") + styles.AccentAlt.Render(strconv.Itoa(d.cunning)) + "\n")
	b.WriteString(styles.Subtle.Render("Wealth ") + styles.AccentAlt.Render(strconv.Itoa(d.wealth)))
	return b.String()
}

// movementDetail lists the faction's movable assets with location and current
// order, from the stashed ask summary. Falls back to the base card if empty.
func (m Model) movementDetail() string {
	if len(m.movables) == 0 {
		return m.baseDetail()
	}
	var b strings.Builder
	b.WriteString(styles.Strong.Render(m.detail.name) + "\n")
	b.WriteString(styles.Subtle.Render("Movable assets") + "\n\n")
	for _, line := range m.movables {
		b.WriteString(styles.Strong.Render(line.name) + "\n")
		b.WriteString(styles.Subtle.Render("  at "+line.location) + "\n")
		b.WriteString(styles.Subtle.Render("  "+line.order) + "\n")
	}
	return b.String()
}

// resolveMovables builds the race-safe movement summary from the ask payload.
// Safe to read the live asset pointers here: the engine is parked on the ask's
// reply when this runs (phase_collector.go ask round-trip).
func resolveMovables(eligible []*domain.Asset, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) []movableLine {
	names := make(map[string]string)
	for _, w := range spatialMap.AllWorlds() {
		names[w.ID()] = w.Name()
	}
	lines := make([]movableLine, 0, len(eligible))
	for _, asset := range eligible {
		name := asset.DefinitionID
		if def, ok := rb.Assets[asset.DefinitionID]; ok {
			name = def.Name
		}
		location := names[asset.Location.WorldID]
		if location == "" {
			location = asset.Location.WorldID
		}
		order := "no order"
		if asset.CurrentOrder != nil {
			dest := asset.CurrentOrder.Destination.WorldID
			if n, ok := names[dest]; ok {
				dest = n
			}
			order = "→ " + dest
		}
		lines = append(lines, movableLine{name: name, location: location, order: order})
	}
	return lines
}
```

> Note: `resolveMovables` rebuilds the world-name map independently of `overlay.worldNames` — the deferred-consolidation duplication flagged in *Out of Scope*. Match the pattern; do not extract a shared helper mid-effort.

##### Commit message
```
feat(tui/turn): phase-reactive faction detail pane

- derive active phase from phase-boundary events (next-phase-on-completion)
- stat-raise: emphasize the three ratings from the snapshot
- movement: list movable assets (name/location/order) from a summary
  stashed off the movement ask payload (race-free; engine parked)
```

## Verification

No automated test harness is prescribed by Discovery for the overlays (they are `huh`-form views); verify by running a real cycle:

- The campaign used for manual verification must contain at least one faction with a **stat-raise-eligible** state and at least one asset with **`Speed > 0`** (else `SelectStatRaise` / `SelectMovementDecisions` never fire and the modals can't be exercised). Confirm or seed this in the test campaign before the Commit 2/3 sessions.
- `execution_test.go` / `format_test.go` already exist; extend only if a pure helper (e.g. `resolveMovables`, `Movement.decisions`) warrants a unit test. `decisions()` is a good candidate — it encodes the OQ-9 both-fields invariant.
- Per-transport cargo: exercise by issuing an Issue move on a transport asset; `SelectTransportCargo` then fires as a follow-on overlay after the movement form closes.

## Reference Exemplars

- `internal/faction/tui/views/manage/wizard/wizard.go` — the `huh.Form` idiom (heap `formData`, pointer bindings, `StateCompleted` transition, `worldsToOptions`, the `synthesize` world→hex resolution).
- `internal/faction/tui/views/turn/overlay/placeholder.go` — the Effort-1 overlay being swapped; its `defaultFor` documents the exact answer type each kind asserts.
- `internal/faction/engine/world/movement.go:66-163` — `BuildMovementMutations`; the Issue (`:82-91`) vs Revise (`:124`) hex asymmetry behind OQ 9.
- `internal/faction/tui/adapter/channels.go:65-75` + `phase_collector.go` — the ask payloads and the parked-engine reply round-trip the overlays answer.
