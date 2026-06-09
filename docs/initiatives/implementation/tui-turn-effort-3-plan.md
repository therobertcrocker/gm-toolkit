# TUI Turn — Effort 3 Plan (The action phase)

`F-005.3`, Effort 3 of 3 on `feature/tui-turn`. The largest Effort: make the action phase interactive. Swap Effort 1's placeholder on `AskSelectAction` for a real picker, replace the stubbed `action.Collector` with a real channel-backed one, build a reusable archetype toolkit, wire the `Esc`-cancel path, and ship **all 12** registered actions — each as a self-contained commit. Additive at the adapter (new `AskKind` values + payloads) and the TUI (new `newOverlay` arms); Efforts 1 and 2 are untouched.

- Overview: [`tui-turn-plan.md`](./tui-turn-plan.md) (routes OQ 11/12/13/14 here)
- Discovery: [`tui-turn-discovery.md`](../discovery/tui-turn-discovery.md) — Decision 7 + 14, *Effort 3 modals*, *Esc-cancel path*
- Effort 1 (landed): [`completed/tui-turn-effort-1-plan.md`](./completed/tui-turn-effort-1-plan.md)
- Effort 2 (landed): [`completed/tui-turn-effort-2-plan.md`](./completed/tui-turn-effort-2-plan.md)

## Context / Goal

Effort 1 ran the real engine end-to-end with placeholders; Effort 2 filled the four non-action phase modals. The action phase is the last gap — and the deepest, because it is **two collectors**, not one:

- `PhaseCollector.SelectAction(faction, available)` — pick-one-or-skip (already an `AskKind`, currently on the Effort-1 placeholder).
- `action.Collector` (16 methods) + embedded `hooks.Collector` (2 methods) — fires *inside* `Action.Run` during `runActionPhase`, conditional on which action the GM selected. The adapter ships these as `stubActionCollector` today (`action_collector.go`): the 16 error-returning methods return `ErrActionNotImplemented` (unclassified → **fatal**, aborts the cycle) and the 2 hook methods **panic**. The dry-run was shaped so none were reachable; Effort 3 deliberately reaches them.

The design is fully locked in Discovery (Decisions 7 & 14, the archetype table, the per-action contract). This plan resolves the four routed Plan-detail OQs (11–14), designs the concrete archetype toolkit, and lays out the foundation-plus-nine-actions commit sequence.

**Ground truth re-verified against current source (2026-06-05):**
- `action.Collector` / `hooks.Collector` signatures — `engine/action/collector.go`, `engine/hooks/collector.go`.
- The stub being replaced — `adapter/action_collector.go`; its wiring `adapter.Action()` / `Collectors()` — `adapter/adapter.go:41-44`.
- The seam: `Overlay` interface (`overlay/overlay.go`), `newOverlay` factory + `activeAsk` dispatch + `OverlayDoneMsg` round-trip (`execution/execution.go:135-150, 408-432`), the ask helper + reply round-trip (`phase_collector.go:17-25`), the shared ask channel + `AskKind`/payloads (`channels.go:13-28, 65-75`), `ErrTurnCanceled` (`channels.go:114-116`, declared, **unreferenced**).
- Error classification: `runActionPhase` (`orchestrator.go:388-482`) and `errors/recoverable.go` (`recoverableErrors` currently holds only `action.ErrNoSelection`).
- Every action's `Inputs`/`Resolve` flow — `engine/action/actions/*.go` (all 12 read this session).
- The `huh` overlay idiom to mirror — Effort-2 `overlay/statraise.go`, `overlay/movement.go`, `overlay/cargo.go`.

## Decisions Ratified in Planning

1. **Strict one-commit-per-action (Robert's call).** Foundation commit, then one commit each for Buy, Refit, Repair Asset, Bribe, Seize Planet, Change Homeworld, Attack, Expand Influence, Use Asset Ability — 10 commits total. Sell Asset (shared-only), Repair Faction, and Abandon Goal (zero-prompt) land **in** the foundation commit. Maximal isolation per Discovery's star-graph; each commit is one execution session. *(Resolves the commit-shape question.)*

2. **`Esc`-cancel is scoped to action overlays only (Robert's call).** The foundation wires `Esc → ErrTurnCanceled` into the archetype primitives (every action prompt). The four Effort-2 overlays (stat-raise, movement, cargo, checkpoint) **keep swallowing `Esc`** as they do now — no retrofit. Hook modals support `Esc` but it degrades to the legal no-op (their methods return no `error`). *(Resolves overview OQ 14's scope.)*

3. **Candidate derivation lives adapter-side via exported APIs (Robert's call).** The four prompts whose engine collector derives its own candidates — `SelectSeizeTarget`, `SelectChangeHomeworldTarget`, `SelectExpandInfluenceOrder` (reinforce targets), `SelectBribeTarget` — have that logic re-implemented in their adapter `actionCollector` method using already-exported primitives (`a.engine.World.Index`, `WorldEngine.Location`/`Distance`, `Rulebook.DriftCost`, `Base.EffectiveMaxHP`). Package `actions` is **not** modified; ~2 small predicates (Seize, Change Homeworld) are re-derived. The overlay stays a generic Select over the payload's candidates — no game logic in view code. *(Resolves a re-grounding finding not anticipated in Discovery.)*

4. **`ErrTurnCanceled` is Recoverable; cancel reuses the existing round-trip (resolves OQ 14).** The sentinel is **moved** from `package adapter` to `package action` (engine-side, so `errors/recoverable.go` can classify it without a TUI import — it already imports `action`; the sentinel is currently unreferenced, so the move is safe). It is added to `recoverableErrors`, so canceling an action prompt skips that faction's action and the cycle continues (this *is* "Esc-aborts-current-turn"). No new message type: an overlay emits `OverlayDoneMsg{Answer: action.ErrTurnCanceled}`, the execution model forwards it on `pendingReply` exactly as a normal answer, and the adapter's `ask` helper (`phase_collector.go:21`) already turns an `error` reply into a returned `error`.

5. **Not-yet-built action-unique methods return a recoverable `ErrActionUnavailable` (resolves OQ 13).** The real `actionCollector` replaces the stub; in the foundation commit its action-*unique* methods (everything except the shared `SelectAsset` and the two hooks) return a new `action.ErrActionUnavailable` sentinel, also added to `recoverableErrors`. So an unbuilt method, if ever reached, skips the faction's action and is surfaced via `OnError` — never `panic`, never unclassified-fatal. The UI disable-set (Decision 6) is the *primary* guard that keeps it unreachable; this is the backstop. Each per-action commit replaces its methods with real ask-backed implementations.

6. **The implemented-action registry is a package-level `map[string]bool` keyed by `Name()` (resolves OQ 12).** Lives in the overlay package (`overlay/registry.go`) since the `SelectAction` overlay is the only reader. Foundation seeds it with `{Sell Asset, Repair Faction, Abandon Goal}`; each per-action commit adds one entry. The `SelectAction` overlay renders unimplemented actions with a "(not yet available)" suffix and a `Validate` that rejects selecting one (`huh` has no native disabled-option, so reject-on-submit is the mechanism).

7. **Three generic archetype primitives + three concrete composites (resolves OQ 11's structural half).** `SelectOverlay[T]`, `MultiSelectOverlay[T]`, `ConfirmOverlay` are generic `huh`-backed `Overlay` types carrying an optional context header — these cover 12 of the ~19 prompts. The composites (`twostep`, `numeric`, `expand`) are **concrete paged-`huh.Group` overlays per prompt** sharing the multi-group idiom the movement form already proves, *not* a single generic type — their answer assembly differs per prompt (`BuyOrder` vs `RefitOrder` vs `[]RepairOrder` …), so a generic composite would buy nothing (`feedback_yagni`; matches Effort-2 Decision 1's concrete-per-modal for complex forms). Decision-14's "context-renderer" is realized as option labels carrying the engine detail (cost vs Coin, HP, owner) plus an optional header string above the form — the StatRaise idiom (`"Force (currently 3)"`) generalized.

8. **The action ask helper is hoisted to `*Adapter` (resolves a foundation plumbing detail).** `phase_collector.go`'s `ask` becomes `Adapter.ask`; `phaseCollector.ask` delegates to it (one-line change, call sites unchanged) and the new `actionCollector` uses it too. One shared, unbuffered `askCh` carries both collectors' asks (Discovery Decision 4 — no per-collector channels).

9. **`huh` composite mechanisms + the overlay→adapter value-type seam (ratified during the task-fidelity pass; Robert's calls).**
   - **Two-step composites (Buy, Refit) use `huh.OptionsFunc(fn, &parentValue)`** — present in v1.0.0 (`field_select.go:229`). The child Select rebuilds its options from the committed parent value when navigation enters its group (bindings-hash re-eval, `eval.go:33`; `form.go:570-597` activates the next group, whose `Update` then fires `shouldUpdate`). Chosen over the movement-style HideFunc-group-per-value because each two-step here is *one* parent → *one* child (so OptionsFunc is one group-pair, not N gated groups). Pointer-valued `Select[*domain.AssetDefinition]` is legal — pointers are `comparable`.
   - **`SelectRepairOrders` carries adapter-pre-derived per-asset caps**, not the faction. `HealCount` (`action.RepairOrder`) is the number of escalating-cost heal *steps* — each step restores up to the faction's attribute score in HP and costs `step+1` Coin (`repair_asset.go:72-78`) — **not** an HP total. The adapter derives `adapter.RepairTarget{Asset, HealHP, MaxSteps, Missing}` per damaged asset so the overlay holds no game math. *(Corrects the earlier draft's `0..(def.HP-CurrentHP)` HP-amount range, which mis-modeled the engine.)*
   - **Overlays may import `package adapter` for plain value types** — the `RepairTarget` payload row (Commit 4), the `BribeReply` answer carrier (Commit 5), and the `ReinforceTarget` per-base cap (Commit 9) — but never for the `CollectorAskMsg`/`Reply` plumbing. `adapter` imports neither `overlay` nor `execution`, so the edge is acyclic. The first such import (Commit 4) updates `overlay.go`'s "never import the adapter" comment to scope it to the channels/Reply, not value types.

## Open Questions — To Ratify at Implementation Time

- **Bribe candidate policy (Commit 5).** The engine applies *no* filter to bribe targets (`bribe.go` — the collector defines the set). **Locked to the plan default: all rival-owned bases across `factionState` (`OwnerID != faction.ID`)**, labeled with owner + world + influence (`deriveBribeTargets`, Commit 5 Task 3). Narrow at execution only if play-testing shows the list is unwieldy.
- **Per-prompt label/header copy (every commit).** Exact option-label fields and header wording (OQ 11's cosmetic half) are specified per prompt below but may be tuned at execution for fit within the 14-row wizard band; load-bearing content (the decision-relevant context) is fixed, phrasing is not.
- **`ConfirmAbilityApplied` is a one-sided ack (Commit 10).** `ability/dispatch.go:63` discards the returned bool (only the `error` is checked). Modeled as a `ConfirmOverlay`, but "No" is inert — it proceeds regardless. Built as an acknowledgment confirm; revisit if a real veto is ever wired engine-side.

## Shared Context

### Model discipline (per commit)

| Commit | Model | Why |
|--------|-------|-----|
| 1 — foundation (toolkit + real collector + Esc-cancel) | **Opus** | The generic archetypes, the composite pattern, the collector reshape, and the sentinel/classification changes are the design-heavy core (overview model table) |
| 2 — Buy Asset | **Opus** | First `twostep` composite (world→def) + the stealth-target follow-on `SelectAsset` reuse |
| 3 — Refit Asset | Sonnet | Second `twostep`; follows Commit 2's pattern |
| 4 — Repair Asset | **Opus** | First `numeric` composite (per-asset heal counts + escalating-cost preview) |
| 5 — Bribe | Sonnet | `numeric` (base + amount); adapter derives candidates |
| 6 — Seize Planet | Sonnet | Single `SelectOverlay`; adapter derives contested worlds |
| 7 — Change Homeworld | Sonnet | Single `SelectOverlay`; adapter derives reachable base worlds |
| 8 — Attack | **Opus** | Three prompts incl. the sequential `SelectDefender` loop fired mid-`Resolve` |
| 9 — Expand Influence | **Opus** | The `expand` branching wizard + `ConfirmRivalFreeAttack` + `SelectBaseAttackers` |
| 10 — Use Asset Ability | **Opus** | Three prompts across the `ability` dispatch surface |

Prompt Robert to `/model` at each commit's session start, and again at the shift into the branch-level pre-merge checklist after Commit 10.

### The per-action commit recipe (Commits 2–10)

Every per-action commit is the same five literal moves; each commit below fills the recipe with its specifics rather than re-stating the shape:

1. **AskKind(s)** — append new `const`(s) to the `AskKind` iota block (`channels.go:15-21`). Appending only — never reorder (the placeholder's `defaultFor`/`kindLabel` and `execution`'s `askPhaseLabel` switch on these).
2. **Payload type(s)** — one struct per new kind in `channels.go`'s payload block, carrying the **pre-derived candidates** the overlay renders (Decision 3).
3. **Adapter method(s)** — replace the foundation stub on `*actionCollector` with the real implementation: derive/forward candidates, `a.adapter.ask(kind, faction, payload)`, type-assert the reply into the engine's expected return.
4. **Overlay arm(s)** — add `case adapter.AskFoo:` to `newOverlay` (`execution.go:417`) constructing the archetype; for sub-prompts fired mid-`Resolve` (e.g. `SelectDefender`), the same arm handles each — `activeAsk`/banner already key off `msg.Kind`, but new kinds need a label in `askPhaseLabel` (`execution.go:255`).
5. **Registry entry** — add the action's `Name()` to `overlay.ImplementedActions` (Commit 1 Task 7), enabling it in the `SelectAction` picker.

### Candidate derivation split (Decision 3)

The 14 **forwarding** prompts (`SelectAsset`, `SelectAttackers`, `SelectDefender`, `SelectRepairOrders`, `SelectBuyOrder`, `SelectRefitOrder`, `SelectBaseAttackers`, `SelectAbilityAssets`, `SelectFactionTestTarget`, the two hooks, plus `SelectAction`) receive a ready candidate slice from the engine — the adapter method packs it straight into the payload. The 4 **deriving** prompts (`SelectSeizeTarget`, `SelectChangeHomeworldTarget`, `SelectExpandInfluenceOrder` reinforce side, `SelectBribeTarget`) get only `(faction, factionState)` — their adapter method re-derives candidates engine-side via exported APIs before packing. In both cases the overlay is pure UI over `payload.Candidates`.

### What Effort 3 does NOT touch

The `adapter` `Run`/pumps/observer; the `turn.Model` router and `execution` event-stream/rail/Effort-2 detail panes (action-phase pane enrichment is a small addition in Commit 1 Task 8, not a rewrite); the four Effort-2 overlays; the `phaseCollector` methods (only the shared `ask` is hoisted). The engine `actions` package is untouched (Decision 3). The engine's only change is the sentinel relocation + two `recoverableErrors` additions (Commit 1 Task 1).

## Out of Scope

- **Per-faction cadence, pause/resume, skip-without-abort, reorder, review-last-cycle** — arc deferrals, unchanged from Discovery.
- **Retrofitting `Esc`-cancel onto the Effort-2 overlays** — Decision 2; they keep swallowing `Esc`.
- **A real veto on `ConfirmAbilityApplied`** — the engine discards its bool; out of scope to rewire (OQ above).
- **Deferred consolidation (capture at pre-merge):** `worldOptions`/`worldNames`/`assetName` now exist in `overlay/movement.go`; action overlays reuse those package helpers where possible, but any new option-builder duplication is matched, not extracted mid-effort (`feedback_deferred_refactor`).

---

## Work Breakdown

### Commit Index

One execution session per commit. Grep `### Commit N` to jump. Model per commit is in [Model discipline](#model-discipline-per-commit) above; this index maps each commit to its new overlay file(s) and the collector prompt(s) it wires.

| # | Action | New overlay file(s) | Collector prompt(s) wired |
|---|--------|---------------------|---------------------------|
| 1 | Foundation (Sell / Repair Faction / Abandon Goal) | `archetypes.go`, `registry.go`, `selectaction.go`, `asset.go`, `hooks.go` | `SelectAction`, `SelectAsset`, `SelectModifiers`, `ConfirmReroll` |
| 2 | Buy Asset | `buyorder.go` | `SelectBuyOrder` (+ stealth follow-on reuses `SelectAsset`) |
| 3 | Refit Asset | `refitorder.go` | `SelectRefitOrder` |
| 4 | Repair Asset | `repairorders.go` | `SelectRepairOrders` |
| 5 | Bribe | `bribe.go` | `SelectBribeTarget` |
| 6 | Seize Planet | `worldselect.go` (shared) | `SelectSeizeTarget` |
| 7 | Change Homeworld | — (reuses `worldselect.go`) | `SelectChangeHomeworldTarget` |
| 8 | Attack | `attack.go` | `SelectAttackers`, `SelectDefender`, `ConfirmRedirectToBase` |
| 9 | Expand Influence | `expand.go` | `SelectExpandInfluenceOrder`, `ConfirmRivalFreeAttack`, `SelectBaseAttackers` |
| 10 | Use Asset Ability | `ability.go` | `SelectAbilityAssets`, `ConfirmAbilityApplied`, `SelectFactionTestTarget` |

> **Re-grounding (every execution session):** re-read the target file(s) before editing — Commit 1 reshapes the adapter collector, the overlay package, and the sentinel home; Commits 2–10 each extend `channels.go`, `action_collector.go`, `execution.go:newOverlay`, and `overlay/registry.go`. Verify the anchors against current source first. Each commit's Find-anchor assumes the prior commit's Replace output ("as left by Commit N-1") — execute in order.
>
> Every commit opens with a **Task 0 — ground the prompt surface**: re-read the action's source and confirm the collector methods this commit wires are *exactly* the set its `Inputs`/`Resolve` reaches — no missing prompt, no orphan. Do this before any edit; if `Resolve` reaches a collector method the commit doesn't wire, stop and reconcile the plan first.

### Commit 1 — `feat(tui/turn): action-flow foundation — archetype toolkit, real collector, Esc-cancel`

The design-heavy commit (**Opus**). Ships everything shared so the per-action commits form a star: the archetype toolkit, the real channel-backed `actionCollector` (shared `SelectAsset` + both hooks real; action-unique methods recoverable-stubbed), the real `SelectAction` picker + disable-set, `Esc`-cancel routing, and the three zero-prompt/shared-only actions.

##### Task 0 — ground the prompt surface (do this first)

Re-read `sell_asset.go`, `repair_faction.go`, `abandon_goal.go`, and `engine/hooks/collector.go`. Confirm: `sell_asset.go`'s `Inputs`/`Resolve` reaches **only `SelectAsset`**; `repair_faction.go` and `abandon_goal.go` make **no** collector calls (zero-prompt — they are enabled purely by the `ImplementedActions` seed); the two real hook methods match `hooks.Collector` exactly (`SelectModifiers`, `ConfirmReroll`). If any of the three foundation actions reaches a collector method beyond this, stop and reconcile before writing code.

##### Task 1 — sentinel relocation + recoverable classification

(a) `internal/faction/tui/adapter/channels.go` — remove the local sentinel. **Find:**
```go
// ErrTurnCanceled is the cancel sentinel returned when the user dismisses an
// ask prompt mid-turn. Routing logic for this lives in the action-selection initiative.
var ErrTurnCanceled = errors.New("turn canceled by user")
```
**Replace with:** (delete the block; also drop the now-unused `"errors"` import if nothing else in the file uses it — verify, `channels.go` imports `errors` only for this)

(b) `internal/faction/engine/action/action.go` — add both sentinels next to `ErrNoSelection`. **Find:**
```go
var (
	ErrNoSelection = errors.New("action: no selection")
)
```
**Replace with:**
```go
var (
	ErrNoSelection = errors.New("action: no selection")
	// ErrTurnCanceled is sent on an ask Reply when the GM dismisses an action
	// prompt (Esc). Classified Recoverable, so the faction's action is skipped
	// and the cycle continues. Lives here (not the TUI adapter) so the engine's
	// error classifier can reference it without a TUI import.
	ErrTurnCanceled = errors.New("action: turn canceled by user")
	// ErrActionUnavailable is returned by the adapter's actionCollector for an
	// action-unique prompt whose per-action commit has not landed. Recoverable +
	// surfaced; the SelectAction disable-set keeps it unreachable in practice.
	ErrActionUnavailable = errors.New("action: prompt not yet available")
)
```

(c) `internal/faction/errors/recoverable.go` — classify both as recoverable. **Find:**
```go
var recoverableErrors = []error{
	action.ErrNoSelection,
}
```
**Replace with:**
```go
var recoverableErrors = []error{
	action.ErrNoSelection,
	action.ErrTurnCanceled,
	action.ErrActionUnavailable,
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/archetypes.go` (new) — the generic primitives — full contents:

```go
package overlay

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// cancelOnEsc is the shared Esc handler for every action overlay. Returning a
// non-nil cmd means "canceled": the overlay emits OverlayDoneMsg carrying
// action.ErrTurnCanceled, which the execution model forwards on the reply
// channel; the adapter's ask helper turns it into a returned error the
// orchestrator classifies Recoverable (skip the faction's action, continue).
func cancelOnEsc(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return func() tea.Msg { return OverlayDoneMsg{Answer: action.ErrTurnCanceled} }
	}
	return nil
}

// header renders an optional context block above a form. Empty -> nothing.
func withHeader(header, formView string) string {
	if header == "" {
		return formView
	}
	return lipgloss.JoinVertical(lipgloss.Left, styles.Subtle.Render(header), "", formView)
}

// --- SelectOverlay[T]: single-select with optional context header ---

type selectData[T any] struct{ choice T }

// SelectOverlay is a generic single-select. The skip/decline sentinel (if any)
// is just an option whose value is the zero T (e.g. nil action.Action). The
// answer is the chosen T; Esc cancels.
type SelectOverlay[T any] struct {
	data   *selectData[T]
	form   *huh.Form
	header string
}

func NewSelect[T any](title, header string, opts []huh.Option[T]) SelectOverlay[T] {
	data := &selectData[T]{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[T]().Title(title).Options(opts...).Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return SelectOverlay[T]{data: data, form: form, header: header}
}

func (o SelectOverlay[T]) Init() tea.Cmd { return o.form.Init() }

func (o SelectOverlay[T]) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o SelectOverlay[T]) View() string         { return withHeader(o.header, o.form.View()) }
func (o SelectOverlay[T]) Help() help.KeyMap     { return cancelFormHelp{} }

// --- MultiSelectOverlay[T]: multi-select with optional cap ---

type multiData[T any] struct{ chosen []T }

type MultiSelectOverlay[T any] struct {
	data   *multiData[T]
	form   *huh.Form
	header string
}

// NewMultiSelect builds a multi-select. cap <= 0 means uncapped.
func NewMultiSelect[T any](title, header string, opts []huh.Option[T], cap int) MultiSelectOverlay[T] {
	data := &multiData[T]{}
	sel := huh.NewMultiSelect[T]().Title(title).Options(opts...).Value(&data.chosen)
	if cap > 0 {
		sel = sel.Validate(func(picked []T) error {
			if len(picked) > cap {
				return capError(cap)
			}
			return nil
		})
	}
	form := huh.NewForm(huh.NewGroup(sel)).WithTheme(styles.FormTheme())
	return MultiSelectOverlay[T]{data: data, form: form, header: header}
}

func (o MultiSelectOverlay[T]) Init() tea.Cmd { return o.form.Init() }

func (o MultiSelectOverlay[T]) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.chosen
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o MultiSelectOverlay[T]) View() string     { return withHeader(o.header, o.form.View()) }
func (o MultiSelectOverlay[T]) Help() help.KeyMap { return cancelFormHelp{} }

// --- ConfirmOverlay: yes/no ---

type confirmData struct{ choice bool }

type ConfirmOverlay struct {
	data   *confirmData
	form   *huh.Form
	header string
}

func NewConfirm(title, header, affirmative, negative string) ConfirmOverlay {
	data := &confirmData{}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(title).Affirmative(affirmative).Negative(negative).Value(&data.choice),
		),
	).WithTheme(styles.FormTheme())
	return ConfirmOverlay{data: data, form: form, header: header}
}

func (o ConfirmOverlay) Init() tea.Cmd { return o.form.Init() }

func (o ConfirmOverlay) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o ConfirmOverlay) View() string     { return withHeader(o.header, o.form.View()) }
func (o ConfirmOverlay) Help() help.KeyMap { return cancelFormHelp{} }

var (
	_ Overlay = SelectOverlay[int]{}
	_ Overlay = MultiSelectOverlay[int]{}
	_ Overlay = ConfirmOverlay{}
)
```

> `capError` and `cancelFormHelp` are defined in Task 3. The Effort-2 `formHelp` (no Esc) stays for the four non-action overlays; action overlays use `cancelFormHelp` (adds an Esc/cancel binding).

##### Task 3 — `internal/faction/tui/views/turn/overlay/formhelp.go` — add the cancel-aware help + the cap-error helper. **Find** (the end of the file, after the existing `formHelp` block) and append:

```go

// cancelFormHelp is the keymap for action overlays: form nav plus Esc-cancel.
type cancelFormHelp struct{}

func (cancelFormHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel action")),
	}
}
func (cancelFormHelp) FullHelp() [][]key.Binding { return [][]key.Binding{cancelFormHelp{}.ShortHelp()} }

func capError(n int) error { return fmt.Errorf("at most %d", n) }
```

> Add `"fmt"` to the imports if absent.

##### Task 4 — `internal/faction/tui/views/turn/overlay/registry.go` (new) — full contents:

```go
package overlay

// ImplementedActions is the set of action Names whose collector prompts the TUI
// can drive. The SelectAction overlay disables any available action not in this
// set. Seeded with the zero-prompt and shared-only actions (no per-action commit
// needed); each per-action commit adds one entry.
var ImplementedActions = map[string]bool{
	"Sell Asset":     true, // shared SelectAsset only (built in foundation)
	"Repair Faction": true, // zero-prompt
	"Abandon Goal":   true, // zero-prompt
}
```

##### Task 5 — `internal/faction/tui/views/turn/overlay/selectaction.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type selectActionData struct{ choice action.Action }

// SelectAction is the action-phase picker. It lists every available (engine
// Validate-passed) action; actions absent from ImplementedActions render with a
// "(not yet available)" suffix and are rejected by Validate (huh has no native
// disabled option). A "Skip faction" sentinel maps to a nil action.Action, the
// legal skip path (orchestrator.go:418). Esc also cancels -> skip the action.
type SelectAction struct {
	data *selectActionData
	form *huh.Form
}

func NewSelectAction(available []action.Action) SelectAction {
	data := &selectActionData{}
	opts := make([]huh.Option[action.Action], 0, len(available)+1)
	for _, a := range available {
		label := a.Name()
		if !ImplementedActions[a.Name()] {
			label += " (not yet available)"
		}
		opts = append(opts, huh.NewOption(label, a))
	}
	opts = append(opts, huh.NewOption("Skip faction (take no action)", action.Action(nil)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[action.Action]().
				Title("Select action").
				Options(opts...).
				Value(&data.choice).
				Validate(func(a action.Action) error {
					if a != nil && !ImplementedActions[a.Name()] {
						return fmt.Errorf("%s is not yet available", a.Name())
					}
					return nil
				}),
		),
	).WithTheme(styles.FormTheme())
	return SelectAction{data: data, form: form}
}

func (o SelectAction) Init() tea.Cmd { return o.form.Init() }

func (o SelectAction) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := o.data.choice // action.Action or nil
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o SelectAction) View() string     { return o.form.View() }
func (o SelectAction) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = SelectAction{}
```

> The `nil` answer flows through `OverlayDoneMsg{Answer: nil}` -> `pendingReply <- nil` -> `phaseCollector.SelectAction` sees `raw == nil` -> returns `(nil, nil)` -> skip (`orchestrator.go:418`). The execution `OverlayDoneMsg` handler already sends `msg.Answer` regardless of nilness.

##### Task 6 — `internal/faction/tui/views/turn/overlay/asset.go` (new) — the shared `SelectAsset` overlay — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAsset builds the shared single-asset picker (Sell, and Buy's stealth
// follow-on). Options are labeled name + HP + location. Answer: *domain.Asset.
func NewSelectAsset(title string, assets []*domain.Asset, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(assets))
	for _, a := range assets {
		opts = append(opts, huh.NewOption(assetLabel(a, rb), a))
	}
	return NewSelect[*domain.Asset](title, "", opts)
}

// assetLabel renders "<name> · HP x/y" for an owned asset. assetName lives in
// movement.go (Effort 2). HP max comes from the definition.
func assetLabel(a *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(a, rb)
	if def, ok := rb.Assets[a.DefinitionID]; ok {
		return fmt.Sprintf("%s · HP %d/%d", name, a.CurrentHP, def.HP)
	}
	return name
}

var _ help.KeyMap = cancelFormHelp{}
```

##### Task 7 — `internal/faction/tui/views/turn/overlay/hooks.go` (new) — the two hook modals — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

// NewSelectModifiers builds the pre-roll modifier multi-select (cross-cutting:
// fires during any rolling action). Answer: []hooks.ModifierOffer (the subset to
// apply). Esc -> the execution model receives ErrTurnCanceled, but SelectModifiers
// returns no error, so the adapter method maps cancel to "none applied" (no-op).
func NewSelectModifiers(offers []hooks.ModifierOffer) MultiSelectOverlay[hooks.ModifierOffer] {
	opts := make([]huh.Option[hooks.ModifierOffer], 0, len(offers))
	for _, o := range offers {
		label := o.Description
		if o.Source != "" {
			label = fmt.Sprintf("%s (%s)", o.Description, o.Source)
		}
		opts = append(opts, huh.NewOption(label, o))
	}
	return NewMultiSelect[hooks.ModifierOffer]("Apply pre-roll modifiers", "", opts, 0)
}

// NewConfirmReroll builds the elective-reroll confirm. Answer: bool. Esc maps to
// false (skip the reroll) in the adapter method (ConfirmReroll returns no error).
func NewConfirmReroll(directive hooks.RerollDirective) ConfirmOverlay {
	header := ""
	if directive.Source != "" {
		header = fmt.Sprintf("Source: %s", directive.Source)
	}
	return NewConfirm(fmt.Sprintf("Reroll %d di(c)e?", len(directive.Indices)), header, "Reroll", "Keep")
}
```

##### Task 8 — `internal/faction/tui/adapter/action_collector.go` — replace the stub with the real `actionCollector` — full file rewrite:

```go
package adapter

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// actionCollector is the real, channel-backed action.Collector. Each method
// either forwards engine-supplied candidates or derives them (Decision 3), packs
// a payload, and blocks on the shared ask round-trip. Action-unique methods not
// yet built by a per-action commit return the recoverable action.ErrActionUnavailable
// (never panic, never unclassified-fatal); the SelectAction disable-set keeps
// them unreachable in practice.
type actionCollector struct{ adapter *Adapter }

func newActionCollector(a *Adapter) *actionCollector { return &actionCollector{adapter: a} }

// --- hooks.Collector (no error return; Esc/cancel degrades to the legal no-op) ---

func (a *actionCollector) SelectModifiers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	raw, err := a.adapter.ask(AskSelectModifiers, nil, SelectModifiersPayload{Offers: offers})
	if err != nil {
		return nil // Esc/cancel -> no modifiers applied
	}
	chosen, _ := raw.([]hooks.ModifierOffer)
	return chosen
}

func (a *actionCollector) ConfirmReroll(directive hooks.RerollDirective) bool {
	raw, err := a.adapter.ask(AskConfirmReroll, nil, ConfirmRerollPayload{Directive: directive})
	if err != nil {
		return false // Esc/cancel -> skip the reroll
	}
	confirmed, _ := raw.(bool)
	return confirmed
}

// --- action.Collector: shared (real in foundation) ---

func (a *actionCollector) SelectAsset(assets []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAsset, nil, SelectAssetPayload{Assets: assets})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Asset)
	return chosen, nil
}

// --- action.Collector: action-unique (recoverable-stubbed until each action's commit) ---

func (a *actionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	return action.BuyOrder{}, action.ErrActionUnavailable
}
func (a *actionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
	return action.RefitOrder{}, action.ErrActionUnavailable
}
func (a *actionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (a *actionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	return action.ExpandInfluenceOrder{}, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (a *actionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (a *actionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	return nil, 0, action.ErrActionUnavailable
}
func (a *actionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", action.ErrActionUnavailable
}
func (a *actionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", action.ErrActionUnavailable
}

var _ action.Collector = (*actionCollector)(nil)
```

> The old `stubActionCollector`, `NewStubActionCollector`, and `ErrActionNotImplemented` are deleted (this is a full rewrite of the file). `adapter.go` is updated in Task 9 to construct `newActionCollector`.

##### Task 9 — `internal/faction/tui/adapter/adapter.go` — wire the real collector. **Find:**
```go
func (a *Adapter) Action() action.Collector { return &stubActionCollector{} }
```
**Replace with:**
```go
func (a *Adapter) Action() action.Collector { return newActionCollector(a) }
```

##### Task 10 — `internal/faction/tui/adapter/phase_collector.go` — hoist `ask` to `*Adapter`. **Find:**
```go
// ask sends a CollectorAskMsg to the UI thread and blocks until the reply arrives.
func (p *phaseCollector) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	reply := make(chan any, 1)
	p.adapter.askCh <- CollectorAskMsg{Kind: kind, Faction: faction, Payload: payload, Reply: reply}
	received := <-reply
	if err, ok := received.(error); ok {
		return nil, err
	}
	return received, nil
}
```
**Replace with:**
```go
// ask sends a CollectorAskMsg to the UI thread and blocks until the reply
// arrives. Hoisted onto *Adapter so both the phase and action collectors share
// the one unbuffered askCh (Discovery Decision 4).
func (a *Adapter) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	reply := make(chan any, 1)
	a.askCh <- CollectorAskMsg{Kind: kind, Faction: faction, Payload: payload, Reply: reply}
	received := <-reply
	if err, ok := received.(error); ok {
		return nil, err
	}
	return received, nil
}

func (p *phaseCollector) ask(kind AskKind, faction *domain.Faction, payload any) (any, error) {
	return p.adapter.ask(kind, faction, payload)
}
```

##### Task 11 — `internal/faction/tui/adapter/channels.go` — add the foundation ask kinds + payloads.

(a) Extend the `AskKind` block. **Find:**
```go
const (
	AskAwaitCheckpoint AskKind = iota
	AskSelectAction
	AskSelectStatRaise
	AskSelectMovementDecisions
	AskSelectTransportCargo
)
```
**Replace with:**
```go
const (
	AskAwaitCheckpoint AskKind = iota
	AskSelectAction
	AskSelectStatRaise
	AskSelectMovementDecisions
	AskSelectTransportCargo
	// Action-surface kinds (Effort 3). Appended only — never reorder.
	AskSelectAsset
	AskSelectModifiers
	AskConfirmReroll
)
```

(b) Add the payloads after the existing payload block. **Find:**
```go
type SelectTransportCargoPayload struct {
	Transport     *domain.Asset
	EligibleCargo []*domain.Asset
	Profile       *domain.TransportProfile
}
```
**Replace with:**
```go
type SelectTransportCargoPayload struct {
	Transport     *domain.Asset
	EligibleCargo []*domain.Asset
	Profile       *domain.TransportProfile
}

// Action-surface payloads (Effort 3).
type SelectAssetPayload struct{ Assets []*domain.Asset }
type SelectModifiersPayload struct{ Offers []hooks.ModifierOffer }
type ConfirmRerollPayload struct{ Directive hooks.RerollDirective }
```

> Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"` to `channels.go` imports.

##### Task 12 — `internal/faction/tui/views/turn/execution/execution.go` — swap the SelectAction arm, add the three foundation arms, and label the new kinds.

(a) Replace the placeholder arm + add the shared arms in `newOverlay`. **Find:**
```go
	case adapter.AskSelectAction:
		return overlay.NewPlaceholder(msg.Kind, name)
	case adapter.AskSelectMovementDecisions:
```
**Replace with:**
```go
	case adapter.AskSelectAction:
		return overlay.NewSelectAction(msg.Payload.(adapter.SelectActionPayload).Available)
	case adapter.AskSelectAsset:
		return overlay.NewSelectAsset("Select asset", msg.Payload.(adapter.SelectAssetPayload).Assets, m.rulebook)
	case adapter.AskSelectModifiers:
		return overlay.NewSelectModifiers(msg.Payload.(adapter.SelectModifiersPayload).Offers)
	case adapter.AskConfirmReroll:
		return overlay.NewConfirmReroll(msg.Payload.(adapter.ConfirmRerollPayload).Directive)
	case adapter.AskSelectMovementDecisions:
```

(b) Label the new kinds in `askPhaseLabel`. **Find:**
```go
	case adapter.AskAwaitCheckpoint:
		return "Cycle Checkpoint"
	default:
		return "resolving…"
	}
```
**Replace with:**
```go
	case adapter.AskAwaitCheckpoint:
		return "Cycle Checkpoint"
	case adapter.AskSelectAsset:
		return "Select Asset"
	case adapter.AskSelectModifiers:
		return "Pre-roll Modifiers"
	case adapter.AskConfirmReroll:
		return "Reroll?"
	default:
		return "Action"
	}
```

> The `default: "Action"` makes every not-yet-labeled action sub-kind (Commits 2–10) read "Action" in the banner until each commit adds its own label — a sensible fallback, not a placeholder gap.

(c) Action-phase detail card. **Find** (the action gap in `detailView`, after the movement case):
```go
		case adapter.AskSelectMovementDecisions, adapter.AskSelectTransportCargo:
			return m.movementDetail()
		}
	}
	return m.baseDetail()
```
**Replace with:**
```go
		case adapter.AskSelectMovementDecisions, adapter.AskSelectTransportCargo:
			return m.movementDetail()
		case adapter.AskSelectAction:
			return m.actionDetail()
		}
	}
	return m.baseDetail()
```
and add the helper after `movementDetail`:
```go
// actionDetail augments the base card with asset/base counts during the action
// pick — a light "what can this faction bring to bear" cue. The per-prompt
// overlays carry the fine-grained decision context (attacker stats, costs).
func (m Model) actionDetail() string {
	return m.baseDetail() // Effort 3: counts deferred — base card is sufficient for the pick
}
```

> Kept intentionally minimal (Decision: action-pane enrichment is not a rewrite). `actionDetail` is a seam a later pass can fill; the base card already shows stats/HP/Coin.

##### Task 13 — register the foundation actions in the implemented set: already done in Task 4 (`Sell Asset`, `Repair Faction`, `Abandon Goal` seeded). No engine change — `register.go` already registers all 12; the disable-set governs reachability.

##### Commit message
```
feat(tui/turn): action-flow foundation — toolkit, real collector, Esc-cancel

- archetypes.go: generic SelectOverlay[T]/MultiSelectOverlay[T]/ConfirmOverlay
  with Esc-cancel (OverlayDoneMsg carrying action.ErrTurnCanceled)
- real channel-backed actionCollector replaces the stub: SelectAsset + both
  hook modals real; action-unique methods return recoverable ErrActionUnavailable
- SelectAction picker + ImplementedActions disable-set (Sell/RepairFaction/
  AbandonGoal enabled); hoist ask helper to *Adapter for the shared askCh
- move ErrTurnCanceled to package action, add it + ErrActionUnavailable to
  recoverableErrors; new AskSelectAsset/AskSelectModifiers/AskConfirmReroll kinds
```

---

### Commit 2 — `feat(tui/turn): Buy Asset action`

The first two-step composite (world → definition) via `OptionsFunc` (Decision 9), plus reuse of the foundation `SelectAsset` for the C3-002 stealth follow-on (`buy_asset.go:62-75`). `SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition)` carries the full definitions, so the overlay needs no `rulebook` — pure forwarding plus an adapter-supplied Coin (for cost-vs-budget labels) and world names.

##### Task 0 — ground the prompt surface (do this first)

Re-read `buy_asset.go`. Confirm its `Inputs`/`Resolve` reaches **exactly `SelectBuyOrder`** and — on the C3-002 stealth follow-on (`buy_asset.go:62-75`) — the foundation `SelectAsset`; nothing else. If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add the ask kind and payload.

(a) **Find** (the foundation kinds, as left by Commit 1):
```go
	AskSelectModifiers
	AskConfirmReroll
)
```
**Replace with:**
```go
	AskSelectModifiers
	AskConfirmReroll
	AskSelectBuyOrder
)
```

(b) **Find** (the last Effort-3 payload, as left by Commit 1):
```go
type ConfirmRerollPayload struct{ Directive hooks.RerollDirective }
```
**Replace with:**
```go
type ConfirmRerollPayload struct{ Directive hooks.RerollDirective }
type SelectBuyOrderPayload struct {
	PurchasableByWorld map[string][]*domain.AssetDefinition
	WorldNames         map[string]string // worldID -> display name (adapter-built)
	Coin               int               // for cost-vs-budget labels
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/buyorder.go` (new) — full contents:

```go
package overlay

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// buyOrderData is heap-allocated so huh's pointer bindings survive the
// value-copy of the overlay. world is bound by the first group; def by the
// second, whose options rebuild from world via OptionsFunc.
type buyOrderData struct {
	world string
	def   *domain.AssetDefinition
}

// BuyOrder is the two-step Buy composite: a world Select, then a definition
// Select whose options derive from the chosen world via OptionsFunc (re-evaluated
// when navigation enters the second group; eval.go:33 bindings hash). Definition
// labels carry the sticker cost vs Coin as a budgeting cue — the engine
// re-resolves the true cost via dispatch.ResolveAssetCost at Resolve
// (buy_asset.go:82), so the label is informational, not authoritative.
// Answer: action.BuyOrder. Esc cancels.
type BuyOrder struct {
	data *buyOrderData
	form *huh.Form
}

func NewBuyOrder(purchasableByWorld map[string][]*domain.AssetDefinition, worldNames map[string]string, coin int) BuyOrder {
	data := &buyOrderData{}

	worlds := make([]string, 0, len(purchasableByWorld))
	for world := range purchasableByWorld {
		worlds = append(worlds, world)
	}
	sort.Strings(worlds)

	worldOpts := make([]huh.Option[string], len(worlds))
	for i, world := range worlds {
		label := worldNames[world]
		if label == "" {
			label = world
		}
		worldOpts[i] = huh.NewOption(label, world)
	}

	worldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Buy on which world?").
			Options(worldOpts...).
			Value(&data.world),
	)

	defGroup := huh.NewGroup(
		huh.NewSelect[*domain.AssetDefinition]().
			Title("Purchase which asset?").
			Value(&data.def).
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				defs := purchasableByWorld[data.world]
				opts := make([]huh.Option[*domain.AssetDefinition], len(defs))
				for i, def := range defs {
					opts[i] = huh.NewOption(
						fmt.Sprintf("%s · TL %d · cost %d/%d Coin", def.Name, def.TechLevel, def.Cost, coin),
						def,
					)
				}
				return opts
			}, &data.world),
	)

	form := huh.NewForm(worldGroup, defGroup).WithTheme(styles.FormTheme())
	return BuyOrder{data: data, form: form}
}

func (o BuyOrder) Init() tea.Cmd { return o.form.Init() }

func (o BuyOrder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := action.BuyOrder{World: o.data.world, Definition: o.data.def}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o BuyOrder) View() string      { return o.form.View() }
func (o BuyOrder) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = BuyOrder{}
```

##### Task 3 — `action_collector.go`: replace the foundation `SelectBuyOrder` stub. **Find:**
```go
func (a *actionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	return action.BuyOrder{}, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectBuyOrder(purchasablePerWorld map[string][]*domain.AssetDefinition) (action.BuyOrder, error) {
	coin := 0
	if faction := a.adapter.currentFaction(); faction != nil {
		coin = faction.Coin
	}
	names := make(map[string]string, len(purchasablePerWorld))
	for worldID := range purchasablePerWorld {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectBuyOrder, nil, SelectBuyOrderPayload{
		PurchasableByWorld: purchasablePerWorld,
		WorldNames:         names,
		Coin:               coin,
	})
	if err != nil {
		return action.BuyOrder{}, err
	}
	order, _ := raw.(action.BuyOrder)
	return order, nil
}
```

> `SelectBuyOrder`'s engine signature carries no faction, but the cost-vs-Coin label needs it. Add `Adapter.currentFaction()` once in `action_collector.go` (reused by every Coin-budget prompt — cite it from Refit/Repair/Expand rather than redefining). It reads the acting faction off `factionState.CurrentTurn` — verified `*domain.TurnState` with `FactionOrder []string` / `CurrentIndex int` (`state/faction_state.go:16`, `engine/turn/turn.go:61`); race-free because the engine goroutine is the only writer and it is parked on this ask's reply. **Append after `newActionCollector`:**
> ```go
> // currentFaction returns the faction whose turn is being processed, or nil.
> // Safe from a collector method: the engine goroutine is parked on this ask's
> // reply, so factionState is quiescent.
> func (a *Adapter) currentFaction() *domain.Faction {
> 	turn := a.factionState.CurrentTurn
> 	if turn == nil || turn.CurrentIndex >= len(turn.FactionOrder) {
> 		return nil
> 	}
> 	return a.factionState.Factions[turn.FactionOrder[turn.CurrentIndex]]
> }
> ```

##### Task 4 — `execution.go`: add the arm and label.

(a) **Find** (the `AskSelectMovementDecisions` arm, the first non-foundation arm in `newOverlay`):
```go
	case adapter.AskSelectMovementDecisions:
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
```
**Replace with:**
```go
	case adapter.AskSelectBuyOrder:
		p := msg.Payload.(adapter.SelectBuyOrderPayload)
		return overlay.NewBuyOrder(p.PurchasableByWorld, p.WorldNames, p.Coin)
	case adapter.AskSelectMovementDecisions:
		return overlay.NewMovement(msg.Payload.(adapter.SelectMovementDecisionsPayload).Eligible, m.rulebook, m.spatialMap)
```

(b) **Find** (in `askPhaseLabel`, the action sub-kind fallback added by Commit 1):
```go
	case adapter.AskConfirmReroll:
		return "Reroll?"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskConfirmReroll:
		return "Reroll?"
	case adapter.AskSelectBuyOrder:
		return "Buy Asset"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go`: enable the action. **Find:**
```go
	"Abandon Goal":   true, // zero-prompt
}
```
**Replace with:**
```go
	"Abandon Goal":   true, // zero-prompt
	"Buy Asset":      true,
}
```

##### Commit message
```
feat(tui/turn): Buy Asset action

- twostep BuyOrder overlay (world -> definition via OptionsFunc, cost vs Coin
  labels)
- real SelectBuyOrder adapter method (+ Adapter.currentFaction helper); the
  C3-002 stealth-target follow-on reuses the shared SelectAsset
- enable Buy Asset in the SelectAction picker
```

---

### Commit 3 — `feat(tui/turn): Refit Asset action`

Second two-step (refittable asset → replacement definition), following Commit 2's `OptionsFunc` pattern. `SelectRefitOrder(options []action.RefitOption, rulebook)` already carries `RefitOption{Asset, Replacements}` (pre-filtered to *affordable* same-category replacements, `refit_asset.go:97-120`), so the overlay needs no Coin budget — the cost delta is the only decision context.

##### Task 0 — ground the prompt surface (do this first)

Re-read `refit_asset.go`. Confirm its `Inputs`/`Resolve` reaches **exactly `SelectRefitOrder`** and nothing else. If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add the ask kind and payload.

(a) **Find:**
```go
	AskSelectBuyOrder
)
```
**Replace with:**
```go
	AskSelectBuyOrder
	AskSelectRefitOrder
)
```

(b) **Find:**
```go
type SelectBuyOrderPayload struct {
	PurchasableByWorld map[string][]*domain.AssetDefinition
	WorldNames         map[string]string // worldID -> display name (adapter-built)
	Coin               int               // for cost-vs-budget labels
}
```
**Replace with:**
```go
type SelectBuyOrderPayload struct {
	PurchasableByWorld map[string][]*domain.AssetDefinition
	WorldNames         map[string]string // worldID -> display name (adapter-built)
	Coin               int               // for cost-vs-budget labels
}
type SelectRefitOrderPayload struct{ Options []action.RefitOption }
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/refitorder.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type refitOrderData struct {
	assetID string
	newDef  *domain.AssetDefinition
}

// RefitOrder is the two-step Refit composite: an asset Select, then a
// replacement-definition Select whose options derive from the chosen asset via
// OptionsFunc. Replacement labels show the cost delta the engine charges
// (max(0, newDef.Cost-oldDef.Cost), refit_asset.go:65). Answer: action.RefitOrder.
// Esc cancels.
type RefitOrder struct {
	data      *refitOrderData
	assetByID map[string]*domain.Asset
	form      *huh.Form
}

func NewRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) RefitOrder {
	data := &refitOrderData{}
	assetByID := make(map[string]*domain.Asset, len(options))
	replByAsset := make(map[string][]*domain.AssetDefinition, len(options))

	assetOpts := make([]huh.Option[string], 0, len(options))
	for _, opt := range options {
		assetByID[opt.Asset.ID] = opt.Asset
		replByAsset[opt.Asset.ID] = opt.Replacements
		assetOpts = append(assetOpts, huh.NewOption(assetLabel(opt.Asset, rb), opt.Asset.ID))
	}

	assetGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Refit which asset?").
			Options(assetOpts...).
			Value(&data.assetID),
	)

	defGroup := huh.NewGroup(
		huh.NewSelect[*domain.AssetDefinition]().
			Title("Replace with?").
			Value(&data.newDef).
			OptionsFunc(func() []huh.Option[*domain.AssetDefinition] {
				oldCost := 0
				if asset := assetByID[data.assetID]; asset != nil {
					if oldDef, ok := rb.Assets[asset.DefinitionID]; ok {
						oldCost = oldDef.Cost
					}
				}
				repls := replByAsset[data.assetID]
				opts := make([]huh.Option[*domain.AssetDefinition], len(repls))
				for i, def := range repls {
					opts[i] = huh.NewOption(
						fmt.Sprintf("%s · cost +%d Coin", def.Name, max(0, def.Cost-oldCost)),
						def,
					)
				}
				return opts
			}, &data.assetID),
	)

	form := huh.NewForm(assetGroup, defGroup).WithTheme(styles.FormTheme())
	return RefitOrder{data: data, assetByID: assetByID, form: form}
}

func (o RefitOrder) Init() tea.Cmd { return o.form.Init() }

func (o RefitOrder) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		answer := action.RefitOrder{OldAsset: o.assetByID[o.data.assetID], NewDefinition: o.data.newDef}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o RefitOrder) View() string      { return o.form.View() }
func (o RefitOrder) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = RefitOrder{}
```

> `assetLabel` is the shared helper from `asset.go` (Commit 1 Task 6): `"<name> · HP x/y"`.

##### Task 3 — `action_collector.go`: replace the `SelectRefitOrder` stub. **Find:**
```go
func (a *actionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
	return action.RefitOrder{}, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectRefitOrder(options []action.RefitOption, rb *rulebook.Rulebook) (action.RefitOrder, error) {
	raw, err := a.adapter.ask(AskSelectRefitOrder, nil, SelectRefitOrderPayload{Options: options})
	if err != nil {
		return action.RefitOrder{}, err
	}
	order, _ := raw.(action.RefitOrder)
	return order, nil
}
```

##### Task 4 — `execution.go`: add the arm and label.

(a) **Find:**
```go
	case adapter.AskSelectBuyOrder:
		p := msg.Payload.(adapter.SelectBuyOrderPayload)
		return overlay.NewBuyOrder(p.PurchasableByWorld, p.WorldNames, p.Coin)
```
**Replace with:**
```go
	case adapter.AskSelectBuyOrder:
		p := msg.Payload.(adapter.SelectBuyOrderPayload)
		return overlay.NewBuyOrder(p.PurchasableByWorld, p.WorldNames, p.Coin)
	case adapter.AskSelectRefitOrder:
		return overlay.NewRefitOrder(msg.Payload.(adapter.SelectRefitOrderPayload).Options, m.rulebook)
```

(b) **Find:**
```go
	case adapter.AskSelectBuyOrder:
		return "Buy Asset"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectBuyOrder:
		return "Buy Asset"
	case adapter.AskSelectRefitOrder:
		return "Refit Asset"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go`. **Find:**
```go
	"Buy Asset":      true,
}
```
**Replace with:**
```go
	"Buy Asset":      true,
	"Refit Asset":    true,
}
```

##### Commit message
```
feat(tui/turn): Refit Asset action

- twostep RefitOrder overlay (asset -> replacement via OptionsFunc, +cost-delta
  labels)
- real SelectRefitOrder adapter method; enable Refit Asset in the picker
```

---

### Commit 4 — `feat(tui/turn): Repair Asset action`

First numeric composite (**Opus**). `SelectRepairOrders(faction, damaged, rulebook)` → `[]action.RepairOrder{Asset, HealCount}`, where `HealCount` is the number of escalating-cost heal *steps* (each restores up to the faction's attribute score in HP; step `n` costs `n` Coin — `repair_asset.go:72-78`). The adapter pre-derives per-asset `MaxSteps`/`HealHP`/`Missing` (Decision 9) so the overlay renders `0..MaxSteps` with a cost preview and holds no game math. This is also the first overlay→adapter value-type import (Decision 9): `repairorders.go` reads `adapter.RepairTarget`.

##### Task 0 — ground the prompt surface (do this first)

Re-read `repair_asset.go` (esp. the `HealCount`/escalating-cost math at `repair_asset.go:72-78`). Confirm its `Inputs`/`Resolve` reaches **exactly `SelectRepairOrders`** and nothing else, and that the adapter-derived `RepairTarget{HealHP, MaxSteps, Missing}` still matches the engine's step model. If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `internal/faction/tui/views/turn/overlay/overlay.go`: scope the import rule to the channels/Reply (Decision 9). **Find:**
```go
// Overlay is a modal prompt mounted over the execution view. Overlays never
// import the adapter channels and never see Reply: they render a prompt and
// emit OverlayDoneMsg{Answer} via a tea.Cmd; the execution sub-model owns the
// channel send.
```
**Replace with:**
```go
// Overlay is a modal prompt mounted over the execution view. Overlays never
// touch the adapter's ask/Reply plumbing (CollectorAskMsg, the Reply channel):
// they render a prompt and emit OverlayDoneMsg{Answer} via a tea.Cmd; the
// execution sub-model owns the channel send. Some Effort-3 overlays import plain
// adapter value types (RepairTarget, BribeReply); that edge is acyclic — adapter
// imports neither overlay nor execution.
```

##### Task 2 — `channels.go`: add the ask kind, the `RepairTarget` row, and the payload.

(a) **Find:**
```go
	AskSelectRefitOrder
)
```
**Replace with:**
```go
	AskSelectRefitOrder
	AskSelectRepairOrders
)
```

(b) **Find:**
```go
type SelectRefitOrderPayload struct{ Options []action.RefitOption }
```
**Replace with:**
```go
type SelectRefitOrderPayload struct{ Options []action.RefitOption }

// RepairTarget is a per-asset, adapter-pre-derived repair cap. HealCount in the
// returned action.RepairOrder is a number of escalating-cost heal steps, each
// restoring up to HealHP (the faction's attribute score for the asset category);
// MaxSteps = ceil(Missing/HealHP) is the useful ceiling (repair_asset.go:72-78).
type RepairTarget struct {
	Asset    *domain.Asset
	HealHP   int
	MaxSteps int
	Missing  int
}
type SelectRepairOrdersPayload struct{ Targets []RepairTarget }
```

##### Task 3 — `internal/faction/tui/views/turn/overlay/repairorders.go` (new) — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// repairData.steps is index-aligned with the targets slice; each entry is the
// chosen heal-step count for that asset.
type repairData struct{ steps []int }

// RepairOrders is the numeric Repair composite: one heal-step Select per damaged
// asset over 0..MaxSteps (the adapter-derived useful ceiling). Step n heals
// min(n*HealHP, Missing) and costs n(n+1)/2 Coin (escalating; the engine
// re-validates the total against Coin — repair_asset.go:92-97). Answer:
// []action.RepairOrder for rows with a non-zero step count. Esc cancels.
type RepairOrders struct {
	data    *repairData
	targets []adapter.RepairTarget
	form    *huh.Form
}

func NewRepairOrders(targets []adapter.RepairTarget, rb *rulebook.Rulebook) RepairOrders {
	data := &repairData{steps: make([]int, len(targets))}

	groups := make([]*huh.Group, 0, len(targets))
	for i, target := range targets {
		idx := i
		name := target.Asset.DefinitionID
		if def, ok := rb.Assets[target.Asset.DefinitionID]; ok {
			name = def.Name
		}

		opts := make([]huh.Option[int], 0, target.MaxSteps+1)
		opts = append(opts, huh.NewOption("no repair", 0))
		for step := 1; step <= target.MaxSteps; step++ {
			healed := min(step*target.HealHP, target.Missing)
			cost := step * (step + 1) / 2
			opts = append(opts, huh.NewOption(fmt.Sprintf("heal %d HP (%d step(s), ≈%d Coin)", healed, step, cost), step))
		}

		groups = append(groups, huh.NewGroup(
			huh.NewSelect[int]().
				Title(fmt.Sprintf("%s · HP %d/%d — repair?", name, target.Asset.CurrentHP, target.Asset.CurrentHP+target.Missing)).
				Options(opts...).
				Value(&data.steps[idx]),
		))
	}

	form := huh.NewForm(groups...).WithTheme(styles.FormTheme())
	return RepairOrders{data: data, targets: targets, form: form}
}

func (o RepairOrders) Init() tea.Cmd { return o.form.Init() }

func (o RepairOrders) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		var orders []action.RepairOrder
		for i, step := range o.data.steps {
			if step > 0 {
				orders = append(orders, action.RepairOrder{Asset: o.targets[i].Asset, HealCount: step})
			}
		}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: orders} }
	}
	return o, cmd
}

func (o RepairOrders) View() string      { return o.form.View() }
func (o RepairOrders) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = RepairOrders{}
```

##### Task 4 — `action_collector.go`: replace the `SelectRepairOrders` stub and add the `statRating` helper. **Find:**
```go
func (a *actionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
	return nil, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectRepairOrders(faction *domain.Faction, damaged []*domain.Asset, rb *rulebook.Rulebook) ([]action.RepairOrder, error) {
	targets := make([]RepairTarget, 0, len(damaged))
	for _, asset := range damaged {
		def, ok := rb.Assets[asset.DefinitionID]
		if !ok {
			continue
		}
		healHP := statRating(faction, def.Category)
		if healHP <= 0 {
			continue
		}
		missing := def.HP - asset.CurrentHP
		targets = append(targets, RepairTarget{
			Asset:    asset,
			HealHP:   healHP,
			MaxSteps: (missing + healHP - 1) / healHP, // ceil(missing/healHP)
			Missing:  missing,
		})
	}
	raw, err := a.adapter.ask(AskSelectRepairOrders, faction, SelectRepairOrdersPayload{Targets: targets})
	if err != nil {
		return nil, err
	}
	orders, _ := raw.([]action.RepairOrder)
	return orders, nil
}

// statRating returns the faction's attribute rating for an asset category,
// mirroring actions.statScore (the engine's repair/heal math). Kept adapter-side
// so the engine actions package is untouched (Decision 3).
func statRating(faction *domain.Faction, stat domain.FactionStat) int {
	switch stat {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}
```

##### Task 5 — `execution.go`: add the arm and label.

(a) **Find:**
```go
	case adapter.AskSelectRefitOrder:
		return overlay.NewRefitOrder(msg.Payload.(adapter.SelectRefitOrderPayload).Options, m.rulebook)
```
**Replace with:**
```go
	case adapter.AskSelectRefitOrder:
		return overlay.NewRefitOrder(msg.Payload.(adapter.SelectRefitOrderPayload).Options, m.rulebook)
	case adapter.AskSelectRepairOrders:
		return overlay.NewRepairOrders(msg.Payload.(adapter.SelectRepairOrdersPayload).Targets, m.rulebook)
```

(b) **Find:**
```go
	case adapter.AskSelectRefitOrder:
		return "Refit Asset"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectRefitOrder:
		return "Refit Asset"
	case adapter.AskSelectRepairOrders:
		return "Repair Asset"
	default:
		return "Action"
	}
```

##### Task 6 — `overlay/registry.go`. **Find:**
```go
	"Refit Asset":    true,
}
```
**Replace with:**
```go
	"Refit Asset":    true,
	"Repair Asset":   true,
}
```

##### Commit message
```
feat(tui/turn): Repair Asset action

- numeric RepairOrders overlay (per-asset heal-STEP selects, escalating-cost
  preview); adapter pre-derives per-asset caps (HealHP/MaxSteps/Missing)
- real SelectRepairOrders adapter method (+ statRating helper); scope overlay.go's
  adapter-import note to the channels/Reply; enable Repair Asset in the picker
```

---

### Commit 5 — `feat(tui/turn): Bribe action`

Numeric (base + amount). `SelectBribeTarget(faction, factionState) (*domain.Base, int, error)` — a **deriving** prompt (Decision 3): the adapter enumerates candidate bases (locked default: all rival-owned bases across `factionState` — see Open Questions). The answer is carried as `adapter.BribeReply` so the adapter asserts one value and never imports `overlay` (Decision 9).

##### Task 0 — ground the prompt surface (do this first)

Re-read `bribe.go`. Confirm its `Inputs`/`Resolve` reaches **exactly `SelectBribeTarget`** and nothing else, and that the engine applies **no** target filter (the adapter's `deriveBribeTargets` owns the candidate policy — Open Questions). If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add the ask kind, the `BribeReply` carrier, and the payload.

(a) **Find:**
```go
	AskSelectRepairOrders
)
```
**Replace with:**
```go
	AskSelectRepairOrders
	AskSelectBribeTarget
)
```

(b) **Find:**
```go
type SelectRepairOrdersPayload struct{ Targets []RepairTarget }
```
**Replace with:**
```go
type SelectRepairOrdersPayload struct{ Targets []RepairTarget }
type SelectBribeTargetPayload struct {
	Bases      []*domain.Base
	OwnerNames map[string]string // ownerID -> faction name
	Coin       int
}

// BribeReply pairs the two return values of SelectBribeTarget into a single
// OverlayDoneMsg.Answer the adapter type-asserts (Decision 9: the overlay emits
// this so the adapter need not import overlay).
type BribeReply struct {
	Base   *domain.Base
	Amount int
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/bribe.go` (new) — full contents:

```go
package overlay

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

type bribeData struct {
	base   *domain.Base
	amount string // parsed to int on submit
}

// Bribe is the numeric Bribe composite: a base Select over rival-owned bases
// (derived adapter-side), then a Coin-amount Input validated to 1..Coin. The
// answer is carried as adapter.BribeReply. Esc cancels.
type Bribe struct {
	data *bribeData
	form *huh.Form
}

func NewBribe(bases []*domain.Base, ownerNames map[string]string, coin int) Bribe {
	data := &bribeData{}

	opts := make([]huh.Option[*domain.Base], len(bases))
	for i, base := range bases {
		owner := ownerNames[base.OwnerID]
		if owner == "" {
			owner = base.OwnerID
		}
		opts[i] = huh.NewOption(
			fmt.Sprintf("%s base @ %s · infl %d", owner, base.Location.WorldID, base.Influence),
			base,
		)
	}

	baseGroup := huh.NewGroup(
		huh.NewSelect[*domain.Base]().
			Title("Bribe which base?").
			Options(opts...).
			Value(&data.base),
	)

	amountGroup := huh.NewGroup(
		huh.NewInput().
			Title(fmt.Sprintf("Coin to spend (1–%d)", coin)).
			Value(&data.amount).
			Validate(func(s string) error {
				n, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("enter a number")
				}
				if n < 1 || n > coin {
					return fmt.Errorf("1–%d", coin)
				}
				return nil
			}),
	)

	form := huh.NewForm(baseGroup, amountGroup).WithTheme(styles.FormTheme())
	return Bribe{data: data, form: form}
}

func (o Bribe) Init() tea.Cmd { return o.form.Init() }

func (o Bribe) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		amount, _ := strconv.Atoi(o.data.amount) // Validate guarantees a valid int
		answer := adapter.BribeReply{Base: o.data.base, Amount: amount}
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: answer} }
	}
	return o, cmd
}

func (o Bribe) View() string      { return o.form.View() }
func (o Bribe) Help() help.KeyMap { return cancelFormHelp{} }

var _ Overlay = Bribe{}
```

##### Task 3 — `action_collector.go`: replace the `SelectBribeTarget` stub and add `deriveBribeTargets`. **Find:**
```go
func (a *actionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	return nil, 0, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectBribeTarget(faction *domain.Faction, factionState *state.FactionState) (*domain.Base, int, error) {
	bases, ownerNames := deriveBribeTargets(factionState, faction.ID)
	raw, err := a.adapter.ask(AskSelectBribeTarget, faction, SelectBribeTargetPayload{
		Bases:      bases,
		OwnerNames: ownerNames,
		Coin:       faction.Coin,
	})
	if err != nil {
		return nil, 0, err
	}
	reply, _ := raw.(BribeReply)
	return reply.Base, reply.Amount, nil
}

// deriveBribeTargets returns every base owned by a faction other than the actor
// (the locked candidate policy), plus an ownerID->name map for labels. The
// engine applies no filter of its own (bribe.go), so the adapter defines the set.
func deriveBribeTargets(factionState *state.FactionState, factionID string) ([]*domain.Base, map[string]string) {
	var bases []*domain.Base
	names := make(map[string]string)
	for id, faction := range factionState.Factions {
		if id == factionID {
			continue
		}
		names[id] = faction.Name
		bases = append(bases, faction.Bases...)
	}
	sort.Slice(bases, func(i, j int) bool { return bases[i].ID < bases[j].ID })
	return bases, names
}
```

> Add `"sort"` to `action_collector.go`'s imports (first use).

##### Task 4 — `execution.go`: add the arm and label.

(a) **Find:**
```go
	case adapter.AskSelectRepairOrders:
		return overlay.NewRepairOrders(msg.Payload.(adapter.SelectRepairOrdersPayload).Targets, m.rulebook)
```
**Replace with:**
```go
	case adapter.AskSelectRepairOrders:
		return overlay.NewRepairOrders(msg.Payload.(adapter.SelectRepairOrdersPayload).Targets, m.rulebook)
	case adapter.AskSelectBribeTarget:
		p := msg.Payload.(adapter.SelectBribeTargetPayload)
		return overlay.NewBribe(p.Bases, p.OwnerNames, p.Coin)
```

(b) **Find:**
```go
	case adapter.AskSelectRepairOrders:
		return "Repair Asset"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectRepairOrders:
		return "Repair Asset"
	case adapter.AskSelectBribeTarget:
		return "Bribe"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go`. **Find:**
```go
	"Repair Asset":   true,
}
```
**Replace with:**
```go
	"Repair Asset":   true,
	"Bribe":          true,
}
```

##### Commit message
```
feat(tui/turn): Bribe action

- numeric Bribe overlay (rival-base select + capped Coin-amount input), answer
  carried as adapter.BribeReply
- real SelectBribeTarget adapter method (deriveBribeTargets: all rival-owned
  bases from factionState); enable Bribe in the picker
```

---

### Commit 6 — `feat(tui/turn): Seize Planet action`

Single `SelectOverlay[string]` over contested worlds. **Deriving** prompt: the adapter re-implements `seizePlanetTargetWorlds` using the exported `a.engine.World.Index.AssetsByLocation` (Decision 3).

##### Task 0 — ground the prompt surface (do this first)

Re-read `seize_planet.go`. Confirm its `Inputs`/`Resolve` reaches **exactly `SelectSeizeTarget`** and nothing else, and that the adapter's re-derived `seizePlanetTargetWorlds` predicate still matches the engine's contested-world filter. If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add the ask kind and payload.

(a) **Find** (the iota tail as left by Commit 5):
```go
	AskSelectBribeTarget
)
```
**Replace with:**
```go
	AskSelectBribeTarget
	AskSelectSeizeTarget
)
```

(b) **Find** (the `BribeReply` block as left by Commit 5):
```go
// BribeReply pairs the two return values of SelectBribeTarget into a single
// OverlayDoneMsg.Answer the adapter type-asserts (Decision 9: the overlay emits
// this so the adapter need not import overlay).
type BribeReply struct {
	Base   *domain.Base
	Amount int
}
```
**Replace with:**
```go
// BribeReply pairs the two return values of SelectBribeTarget into a single
// OverlayDoneMsg.Answer the adapter type-asserts (Decision 9: the overlay emits
// this so the adapter need not import overlay).
type BribeReply struct {
	Base   *domain.Base
	Amount int
}
type SelectSeizeTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/worldselect.go` (new) — the shared world-Select helper (reused by Commit 7) — full contents:

```go
package overlay

import "github.com/charmbracelet/huh"

// NewWorldSelect builds a single-world SelectOverlay from parallel world IDs and
// an adapter-supplied worldID->name map. Label = name (falling back to the raw
// ID), value = the world ID. Shared by Seize Planet and Change Homeworld.
func NewWorldSelect(title, header string, worlds []string, names map[string]string) SelectOverlay[string] {
	opts := make([]huh.Option[string], len(worlds))
	for i, worldID := range worlds {
		label := names[worldID]
		if label == "" {
			label = worldID
		}
		opts[i] = huh.NewOption(label, worldID)
	}
	return NewSelect[string](title, header, opts)
}
```

##### Task 3 — `action_collector.go`: replace the `SelectSeizeTarget` stub and add `deriveContestedWorlds`. **Find:**
```go
func (a *actionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectSeizeTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	worlds := deriveContestedWorlds(faction, a.adapter.engine.World.Index)
	names := make(map[string]string, len(worlds))
	for _, worldID := range worlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectSeizeTarget, faction, SelectSeizeTargetPayload{
		Worlds:     worlds,
		WorldNames: names,
	})
	if err != nil {
		return "", err
	}
	selected, _ := raw.(string)
	return selected, nil
}

// deriveContestedWorlds ports seizePlanetTargetWorlds (seize_planet.go): worlds
// where the faction has at least one unstealthed asset and some rival also has an
// unstealthed asset there. Reads the engine's world Index — quiescent because the
// engine goroutine is parked on this ask's reply, and the same source the engine's
// own Validate consulted to enable the action.
func deriveContestedWorlds(faction *domain.Faction, index *world.Index) []string {
	factionWorlds := map[string]struct{}{}
	for _, asset := range faction.Assets {
		if !asset.Stealthy {
			factionWorlds[asset.Location.WorldID] = struct{}{}
		}
	}
	contested := make([]string, 0, len(factionWorlds))
	for worldID := range factionWorlds {
		for _, asset := range index.AssetsByLocation[worldID] {
			if asset.OwnerID != faction.ID && !asset.Stealthy {
				contested = append(contested, worldID)
				break
			}
		}
	}
	sort.Strings(contested)
	return contested
}
```

> Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"` to `action_collector.go`'s imports (first use). `sort` is already imported (Commit 5's `deriveBribeTargets`). The inline `WorldNames` build mirrors Commit 2's `SelectBuyOrder` rather than extracting a helper — duplication is matched, not refactored mid-effort (pre-merge consolidation item).

##### Task 4 — `execution.go`: add the arm and label.

(a) **Find** (the Bribe arm as left by Commit 5):
```go
	case adapter.AskSelectBribeTarget:
		p := msg.Payload.(adapter.SelectBribeTargetPayload)
		return overlay.NewBribe(p.Bases, p.OwnerNames, p.Coin)
```
**Replace with:**
```go
	case adapter.AskSelectBribeTarget:
		p := msg.Payload.(adapter.SelectBribeTargetPayload)
		return overlay.NewBribe(p.Bases, p.OwnerNames, p.Coin)
	case adapter.AskSelectSeizeTarget:
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
```

(b) **Find** (the Bribe label as left by Commit 5):
```go
	case adapter.AskSelectBribeTarget:
		return "Bribe"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectBribeTarget:
		return "Bribe"
	case adapter.AskSelectSeizeTarget:
		return "Seize Planet"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go`. **Find:**
```go
	"Bribe":          true,
}
```
**Replace with:**
```go
	"Bribe":          true,
	"Seize Planet":   true,
}
```

##### Commit message
```
feat(tui/turn): Seize Planet action

- SelectOverlay over contested worlds; adapter derives candidates from the
  world index (ports seizePlanetTargetWorlds via exported APIs)
- real SelectSeizeTarget adapter method; enable Seize Planet in the picker
```

---

### Commit 7 — `feat(tui/turn): Change Homeworld action`

Single `SelectOverlay[string]` over reachable base worlds. **Deriving** prompt: the adapter ports `changeHomeworldTargets` using exported `a.engine.World.Location`/`Distance` + `a.engine.Rulebook.DriftCost` (`change_homeworld.go:83-101`).

##### Task 0 — ground the prompt surface (do this first)

Re-read `change_homeworld.go` (esp. the target derivation at `change_homeworld.go:83-101`). Confirm its `Inputs`/`Resolve` reaches **exactly `SelectChangeHomeworldTarget`** and nothing else, and that the adapter's ported `changeHomeworldTargets` still matches the engine's reachability + drift-cost filter. If a roll on this path raises hook offers, `SelectModifiers`/`ConfirmReroll` are foundation-built — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add the ask kind and payload.

(a) **Find** (the iota tail as left by Commit 6):
```go
	AskSelectSeizeTarget
)
```
**Replace with:**
```go
	AskSelectSeizeTarget
	AskSelectChangeHomeworldTarget
)
```

(b) **Find** (the `SelectSeizeTargetPayload` as left by Commit 6):
```go
type SelectSeizeTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
```
**Replace with:**
```go
type SelectSeizeTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string // worldID -> display name (adapter-built)
}
type SelectChangeHomeworldTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string
}
```

##### Task 2 — `action_collector.go`: replace the `SelectChangeHomeworldTarget` stub and add `deriveHomeworldTargets`. **Find:**
```go
func (a *actionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	return "", action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectChangeHomeworldTarget(faction *domain.Faction, factionState *state.FactionState) (string, error) {
	worlds := deriveHomeworldTargets(faction, a.adapter.engine.World, a.adapter.engine.Rulebook)
	names := make(map[string]string, len(worlds))
	for _, worldID := range worlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectChangeHomeworldTarget, faction, SelectChangeHomeworldTargetPayload{
		Worlds:     worlds,
		WorldNames: names,
	})
	if err != nil {
		return "", err
	}
	selected, _ := raw.(string)
	return selected, nil
}

// deriveHomeworldTargets ports changeHomeworldTargets (change_homeworld.go):
// non-homeworld base worlds reachable from the current homeworld at the default
// drift rating (DriftCost(3)). Excludes the current homeworld and any world the
// router can't path to.
func deriveHomeworldTargets(faction *domain.Faction, worldEngine *world.WorldEngine, rb *rulebook.Rulebook) []string {
	crossingCost := rb.DriftCost(3)
	var targets []string
	for _, base := range faction.Bases {
		if base.Location.WorldID == faction.Homeworld.WorldID {
			continue
		}
		target, ok := worldEngine.Location(base.Location.WorldID)
		if !ok {
			continue
		}
		if _, err := worldEngine.Distance(faction.Homeworld.RegionHex, target.RegionHex(), crossingCost); err != nil {
			continue
		}
		targets = append(targets, base.Location.WorldID)
	}
	sort.Strings(targets)
	return targets
}
```

> `world` and `rulebook` are already imported (Commit 6 added `world`; `rulebook` is in the Commit-1 rewrite). The inline `WorldNames` build mirrors Seize/Buy.

##### Task 3 — `execution.go`: add the arm and label.

(a) **Find** (the Seize arm as left by Commit 6):
```go
	case adapter.AskSelectSeizeTarget:
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
```
**Replace with:**
```go
	case adapter.AskSelectSeizeTarget:
		p := msg.Payload.(adapter.SelectSeizeTargetPayload)
		return overlay.NewWorldSelect("Seize which world?", "", p.Worlds, p.WorldNames)
	case adapter.AskSelectChangeHomeworldTarget:
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
```

(b) **Find** (the Seize label as left by Commit 6):
```go
	case adapter.AskSelectSeizeTarget:
		return "Seize Planet"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectSeizeTarget:
		return "Seize Planet"
	case adapter.AskSelectChangeHomeworldTarget:
		return "Change Homeworld"
	default:
		return "Action"
	}
```

##### Task 4 — `overlay/registry.go`: add the key (re-aligns the block — `"Change Homeworld"` is now the widest key). **Find** (the whole map as left by Commit 6):
```go
var ImplementedActions = map[string]bool{
	"Sell Asset":     true, // shared SelectAsset only (built in foundation)
	"Repair Faction": true, // zero-prompt
	"Abandon Goal":   true, // zero-prompt
	"Buy Asset":      true,
	"Refit Asset":    true,
	"Repair Asset":   true,
	"Bribe":          true,
	"Seize Planet":   true,
}
```
**Replace with:**
```go
var ImplementedActions = map[string]bool{
	"Sell Asset":       true, // shared SelectAsset only (built in foundation)
	"Repair Faction":   true, // zero-prompt
	"Abandon Goal":     true, // zero-prompt
	"Buy Asset":        true,
	"Refit Asset":      true,
	"Repair Asset":     true,
	"Bribe":            true,
	"Seize Planet":     true,
	"Change Homeworld": true,
}
```

##### Commit message
```
feat(tui/turn): Change Homeworld action

- SelectOverlay over reachable base worlds; adapter derives candidates via
  WorldEngine.Distance + Rulebook.DriftCost (ports changeHomeworldTargets)
- real SelectChangeHomeworldTarget adapter method; enable in the picker
```

---

### Commit 8 — `feat(tui/turn): Attack action`

Three prompts (**Opus**), one fired **mid-`Resolve`** in a loop: `SelectAttackers` (commit up front), then per surviving attacker `SelectDefender`, and conditionally `ConfirmRedirectToBase` (`attack.go:52,85,164`). All forwarding. The sequential nature is already handled by the seam — each `ask` blocks the engine goroutine and the TUI mounts one overlay at a time; no special plumbing.

##### Task 0 — ground the prompt surface (do this first)

Re-read `attack.go` (the three call sites at `attack.go:52,85,164`). Confirm its `Inputs`/`Resolve` reaches **exactly `SelectAttackers`, `SelectDefender`, `ConfirmRedirectToBase`** — and verify the `SelectDefender` loop still fires once per surviving attacker, with `ConfirmRedirectToBase` only on the redirect branch. This action **rolls**, so `SelectModifiers`/`ConfirmReroll` (foundation-built) will fire mid-combat — confirm they're reachable on this path, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add three ask kinds and payloads.

(a) **Find** (the iota tail as left by Commit 7):
```go
	AskSelectChangeHomeworldTarget
)
```
**Replace with:**
```go
	AskSelectChangeHomeworldTarget
	AskSelectAttackers
	AskSelectDefender
	AskConfirmRedirectToBase
)
```

(b) **Find** (the `SelectChangeHomeworldTargetPayload` as left by Commit 7):
```go
type SelectChangeHomeworldTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string
}
```
**Replace with:**
```go
type SelectChangeHomeworldTargetPayload struct {
	Worlds     []string
	WorldNames map[string]string
}
type SelectAttackersPayload struct{ Eligible []*domain.Asset }
type SelectDefenderPayload struct {
	Attacker   *domain.Asset
	Eligible   []*domain.Asset
	OwnerNames map[string]string // ownerID -> faction name (defenders may be rivals')
}
type ConfirmRedirectToBasePayload struct {
	DefenderFaction *domain.Faction
	Base            *domain.Base
	Damage          int
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/attack.go` (new) — three constructors over the generic archetypes — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAttackers builds the multi-select of committable attackers. Each label
// carries the attacker's attack stat and damage so the GM can weigh the
// commitment. Answer: []*domain.Asset (uncapped). Esc cancels.
func NewSelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(attackerLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Commit which attackers?", "", opts, 0)
}

// NewSelectDefender builds the per-attacker defender pick. The header names the
// attacker and its attack matchup; defender labels carry HP plus the owning
// faction (defenders belong to rivals). Answer: *domain.Asset. Esc cancels the
// whole attack action (the engine wraps the returned error → recoverable).
func NewSelectDefender(attacker *domain.Asset, eligible []*domain.Asset, ownerNames map[string]string, rb *rulebook.Rulebook) SelectOverlay[*domain.Asset] {
	header := fmt.Sprintf("%s attacks", assetName(attacker, rb))
	if def, ok := rb.Assets[attacker.DefinitionID]; ok && def.Attack != nil {
		header = fmt.Sprintf("%s attacks (%s vs %s)", assetName(attacker, rb), def.Attack.AttackerStat, def.Attack.DefenderStat)
	}
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, defender := range eligible {
		owner := ownerNames[defender.OwnerID]
		if owner == "" {
			owner = defender.OwnerID
		}
		opts = append(opts, huh.NewOption(fmt.Sprintf("%s · %s", assetLabel(defender, rb), owner), defender))
	}
	return NewSelect[*domain.Asset]("Target which defender?", header, opts)
}

// NewConfirmRedirectToBase asks whether a winning hit lands on the defender's
// Base (which also damages faction HP) instead of the asset. Answer: bool. Esc
// cancels the attack action.
func NewConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) ConfirmOverlay {
	header := fmt.Sprintf("Base HP %d/%d — redirecting also deals %d to %s's faction HP", base.CurrentHP, base.MaxHP, damage, defenderFaction.Name)
	title := fmt.Sprintf("Redirect %d damage to %s's base @ %s?", damage, defenderFaction.Name, base.Location.WorldID)
	return NewConfirm(title, header, "Redirect to base", "Hit the asset")
}

// attackerLabel renders "<name> · <stat> · dmg NdM(±K)" for a committable attacker.
func attackerLabel(asset *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(asset, rb)
	def, ok := rb.Assets[asset.DefinitionID]
	if !ok || def.Attack == nil {
		return name
	}
	return fmt.Sprintf("%s · %s · dmg %s", name, def.Attack.AttackerStat, diceLabel(def.Attack.Damage))
}

func diceLabel(d domain.DiceRoll) string {
	if d.Modifier != 0 {
		return fmt.Sprintf("%dd%d%+d", d.NumDice, d.Sides, d.Modifier)
	}
	return fmt.Sprintf("%dd%d", d.NumDice, d.Sides)
}
```

> No `var _ Overlay` assertions here — the constructors return the generic archetypes (`MultiSelectOverlay`/`SelectOverlay`/`ConfirmOverlay`), which already assert in `archetypes.go`. `assetName` (movement.go) and `assetLabel` (asset.go) are reused.

##### Task 3 — `action_collector.go`: replace the three Attack stubs and add the `ownerNames` helper. **Find:**
```go
func (a *actionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectAttackers(eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAttackers, nil, SelectAttackersPayload{Eligible: eligible})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) SelectDefender(attacker *domain.Asset, eligible []*domain.Asset, rb *rulebook.Rulebook) (*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectDefender, nil, SelectDefenderPayload{
		Attacker:   attacker,
		Eligible:   eligible,
		OwnerNames: a.adapter.ownerNames(eligible),
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) ConfirmRedirectToBase(defenderFaction *domain.Faction, base *domain.Base, damage int) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmRedirectToBase, nil, ConfirmRedirectToBasePayload{
		DefenderFaction: defenderFaction,
		Base:            base,
		Damage:          damage,
	})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
```

> Append the `ownerNames` helper after `currentFaction` (in `action_collector.go`); reused by Commit 9's `SelectBaseAttackers`:
> ```go
> // ownerNames maps each asset's owning faction ID to its display name, for
> // labels on prompts that mix factions' assets (defender/base-attacker picks).
> func (a *Adapter) ownerNames(assets []*domain.Asset) map[string]string {
> 	names := make(map[string]string)
> 	for _, asset := range assets {
> 		if _, seen := names[asset.OwnerID]; seen {
> 			continue
> 		}
> 		if faction := a.factionState.Factions[asset.OwnerID]; faction != nil {
> 			names[asset.OwnerID] = faction.Name
> 		}
> 	}
> 	return names
> }
> ```

##### Task 4 — `execution.go`: add the three arms and labels.

(a) **Find** (the Change Homeworld arm as left by Commit 7):
```go
	case adapter.AskSelectChangeHomeworldTarget:
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
```
**Replace with:**
```go
	case adapter.AskSelectChangeHomeworldTarget:
		p := msg.Payload.(adapter.SelectChangeHomeworldTargetPayload)
		return overlay.NewWorldSelect("New homeworld?", "", p.Worlds, p.WorldNames)
	case adapter.AskSelectAttackers:
		return overlay.NewSelectAttackers(msg.Payload.(adapter.SelectAttackersPayload).Eligible, m.rulebook)
	case adapter.AskSelectDefender:
		p := msg.Payload.(adapter.SelectDefenderPayload)
		return overlay.NewSelectDefender(p.Attacker, p.Eligible, p.OwnerNames, m.rulebook)
	case adapter.AskConfirmRedirectToBase:
		p := msg.Payload.(adapter.ConfirmRedirectToBasePayload)
		return overlay.NewConfirmRedirectToBase(p.DefenderFaction, p.Base, p.Damage)
```

(b) **Find** (the Change Homeworld label as left by Commit 7):
```go
	case adapter.AskSelectChangeHomeworldTarget:
		return "Change Homeworld"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectChangeHomeworldTarget:
		return "Change Homeworld"
	case adapter.AskSelectAttackers:
		return "Select Attackers"
	case adapter.AskSelectDefender:
		return "Select Defender"
	case adapter.AskConfirmRedirectToBase:
		return "Redirect to Base?"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go` (`"Attack"` is shorter than the widest key, so no re-alignment). **Find:**
```go
	"Change Homeworld": true,
}
```
**Replace with:**
```go
	"Change Homeworld": true,
	"Attack":           true,
}
```

##### Commit message
```
feat(tui/turn): Attack action

- MultiSelect attackers, Select defender (per matchup), Confirm base-redirect —
  combat-context labels (stats, damage, HP, owner)
- real SelectAttackers/SelectDefender/ConfirmRedirectToBase adapter methods;
  the mid-Resolve defender loop rides the existing one-overlay-at-a-time seam
- enable Attack in the picker
```

---

### Commit 9 — `feat(tui/turn): Expand Influence action`

The `expand` branching wizard (**Opus**) + two more prompts: `SelectExpandInfluenceOrder` (mode → new: world+HP / reinforce: base+submode+HP), `ConfirmRivalFreeAttack`, `SelectBaseAttackers`. `SelectExpandInfluenceOrder` is **deriving** for the reinforce side (Decision 3): the engine passes `eligibleNewBaseWorlds`; the adapter derives reinforce targets (damaged + growable non-homeworld bases) from `faction.Bases` + exported `Base.EffectiveMaxHP` (`expand_influence.go:255-273`).

##### Task 0 — ground the prompt surface (do this first)

Re-read `expand_influence.go` (the call sites + the reinforce-eligibility filters at `expand_influence.go:255-273`). Confirm its `Inputs`/`Resolve` reaches **exactly `SelectExpandInfluenceOrder`, `ConfirmRivalFreeAttack`, `SelectBaseAttackers`** and nothing else, and that the two independent reinforce filters (`damagedNonHomeworldBases`, `growableNonHomeworldBases`) still match the `ReinforceTarget` row below. The rival-free-attack branch **rolls**, so `SelectModifiers`/`ConfirmReroll` (foundation-built) may fire — confirm reachable, do **not** re-wire. If `Resolve` reaches a collector method outside this set, stop and reconcile the plan before writing code.

> **Reinforce modeling (ratified during the task-fidelity pass).** The engine's
> reinforce eligibility is *two independent filters* — `damagedNonHomeworldBases`
> (`CurrentHP < EffectiveMaxHP`, heal-eligible) and `growableNonHomeworldBases`
> (`MaxHP < faction.MaxHP`, max-eligible) — and a base can be in one, both, or
> neither. A flat `[]*domain.Base` would let the wizard offer an invalid submode.
> Following the Decision-9 `RepairTarget` precedent, the adapter pre-derives
> `adapter.ReinforceTarget{Base, CanHeal, HealCap, CanMax, MaxCap}`; the wizard's
> submode `Select` uses `OptionsFunc` to offer only the valid submodes for the
> chosen base, and the HP `Input`'s `Validate` reads the live cap. No game math in
> the view (Decision 3).

##### Task 1 — `channels.go`: add three ask kinds, the `ReinforceTarget` row, and three payloads.

(a) **Find** (the iota tail as left by Commit 8):
```go
	AskConfirmRedirectToBase
)
```
**Replace with:**
```go
	AskConfirmRedirectToBase
	AskSelectExpandInfluenceOrder
	AskConfirmRivalFreeAttack
	AskSelectBaseAttackers
)
```

(b) **Find** (the `ConfirmRedirectToBasePayload` as left by Commit 8):
```go
type ConfirmRedirectToBasePayload struct {
	DefenderFaction *domain.Faction
	Base            *domain.Base
	Damage          int
}
```
**Replace with:**
```go
type ConfirmRedirectToBasePayload struct {
	DefenderFaction *domain.Faction
	Base            *domain.Base
	Damage          int
}

// ReinforceTarget is a per-base, adapter-pre-derived reinforce cap. CanHeal/CanMax
// are the two independent engine filters (damaged vs growable); HealCap and MaxCap
// are the Coin/HP headroom for each submode (expand_influence.go:121-150).
type ReinforceTarget struct {
	Base    *domain.Base
	CanHeal bool
	HealCap int // EffectiveMaxHP - CurrentHP
	CanMax  bool
	MaxCap  int // faction.MaxHP - base.MaxHP
}
type SelectExpandInfluenceOrderPayload struct {
	NewBaseWorlds  []string          // from the engine (eligibleNewBaseWorlds)
	WorldNames     map[string]string // worldID -> display name
	ReinforceBases []ReinforceTarget // adapter-derived
	Coin           int
}
type ConfirmRivalFreeAttackPayload struct {
	Rival       *domain.Faction
	RivalRoll   int
	FactionRoll int
}
type SelectBaseAttackersPayload struct {
	Rival    *domain.Faction
	Eligible []*domain.Asset
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/expand.go` (new) — the branching wizard + the two follow-on constructors — full contents:

```go
package overlay

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/adapter"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/tui/styles"
)

// expandData is heap-allocated so huh's pointer bindings survive the overlay's
// value-copy. The new-base fields (world, newHP) and reinforce fields (base,
// subMode, reinfHP) are gated by mode via group HideFuncs.
type expandData struct {
	mode    action.ExpandMode
	world   string
	newHP   string
	base    *domain.Base
	subMode action.ReinforceMode
	reinfHP string
}

// Expand is the Expand Influence wizard: a mode Select, then one of two gated
// branches. New: world Select + HP Input (HP == Coin cost). Reinforce: base
// Select → submode Select (options derive from the base's caps via OptionsFunc)
// → HP Input (cap derives from base+submode via the Validate closure). Group
// HideFunc is the only branching primitive huh exposes (movement.go). Answer:
// action.ExpandInfluenceOrder. Esc cancels.
type Expand struct {
	data *expandData
	form *huh.Form
}

func NewExpand(newBaseWorlds []string, worldNames map[string]string, reinforce []adapter.ReinforceTarget, coin int) Expand {
	data := &expandData{}

	capsByBase := make(map[string]adapter.ReinforceTarget, len(reinforce))
	for _, target := range reinforce {
		capsByBase[target.Base.ID] = target
	}

	modeOpts := make([]huh.Option[action.ExpandMode], 0, 2)
	if len(newBaseWorlds) > 0 {
		modeOpts = append(modeOpts, huh.NewOption("Place a new base", action.ExpandModeNew))
	}
	if len(reinforce) > 0 {
		modeOpts = append(modeOpts, huh.NewOption("Reinforce a base", action.ExpandModeReinforce))
	}
	modeGroup := huh.NewGroup(
		huh.NewSelect[action.ExpandMode]().
			Title("Expand influence — how?").
			Options(modeOpts...).
			Value(&data.mode),
	)

	// --- New-base branch (gated mode == New) ---
	worldOpts := make([]huh.Option[string], len(newBaseWorlds))
	for i, worldID := range newBaseWorlds {
		label := worldNames[worldID]
		if label == "" {
			label = worldID
		}
		worldOpts[i] = huh.NewOption(label, worldID)
	}
	newWorldGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Place new base on which world?").
			Options(worldOpts...).
			Value(&data.world),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeNew })

	newHPGroup := huh.NewGroup(
		huh.NewInput().
			Title(fmt.Sprintf("New base HP (= Coin cost, 1–%d)", coin)).
			Value(&data.newHP).
			Validate(amountValidator(1, coin)),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeNew })

	// --- Reinforce branch (gated mode == Reinforce) ---
	baseOpts := make([]huh.Option[*domain.Base], len(reinforce))
	for i, target := range reinforce {
		baseOpts[i] = huh.NewOption(reinforceBaseLabel(target), target.Base)
	}
	reinfBaseGroup := huh.NewGroup(
		huh.NewSelect[*domain.Base]().
			Title("Reinforce which base?").
			Options(baseOpts...).
			Value(&data.base),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	reinfSubGroup := huh.NewGroup(
		huh.NewSelect[action.ReinforceMode]().
			Title("Heal damage or raise max HP?").
			Value(&data.subMode).
			OptionsFunc(func() []huh.Option[action.ReinforceMode] {
				var opts []huh.Option[action.ReinforceMode]
				if data.base == nil {
					return opts
				}
				target := capsByBase[data.base.ID]
				if target.CanHeal {
					opts = append(opts, huh.NewOption(fmt.Sprintf("Heal damage (up to %d HP)", target.HealCap), action.ReinforceHeal))
				}
				if target.CanMax {
					opts = append(opts, huh.NewOption(fmt.Sprintf("Raise max HP (up to %d)", target.MaxCap), action.ReinforceMax))
				}
				return opts
			}, &data.base),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	reinfHPGroup := huh.NewGroup(
		huh.NewInput().
			Title("HP amount (Coin cost)").
			Value(&data.reinfHP).
			Validate(func(s string) error {
				return amountValidator(1, reinforceLimit(capsByBase, data.base, data.subMode, coin))(s)
			}),
	).WithHideFunc(func() bool { return data.mode != action.ExpandModeReinforce })

	form := huh.NewForm(modeGroup, newWorldGroup, newHPGroup, reinfBaseGroup, reinfSubGroup, reinfHPGroup).WithTheme(styles.FormTheme())
	return Expand{data: data, form: form}
}

func (o Expand) Init() tea.Cmd { return o.form.Init() }

func (o Expand) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if cmd := cancelOnEsc(msg); cmd != nil {
		return o, cmd
	}
	fm, cmd := o.form.Update(msg)
	o.form = fm.(*huh.Form)
	if o.form.State == huh.StateCompleted {
		order := o.order()
		return o, func() tea.Msg { return OverlayDoneMsg{Answer: order} }
	}
	return o, cmd
}

func (o Expand) View() string      { return o.form.View() }
func (o Expand) Help() help.KeyMap { return cancelFormHelp{} }

func (o Expand) order() action.ExpandInfluenceOrder {
	switch o.data.mode {
	case action.ExpandModeNew:
		hp, _ := strconv.Atoi(o.data.newHP)
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeNew, World: o.data.world, HPAmount: hp}
	case action.ExpandModeReinforce:
		hp, _ := strconv.Atoi(o.data.reinfHP)
		baseID := ""
		if o.data.base != nil {
			baseID = o.data.base.ID
		}
		return action.ExpandInfluenceOrder{Mode: action.ExpandModeReinforce, BaseID: baseID, SubMode: o.data.subMode, HPAmount: hp}
	}
	return action.ExpandInfluenceOrder{}
}

func reinforceBaseLabel(target adapter.ReinforceTarget) string {
	return fmt.Sprintf("Base @ %s · HP %d/%d", target.Base.Location.WorldID, target.Base.CurrentHP, target.Base.MaxHP)
}

// reinforceLimit is the live Coin/HP cap for the chosen base+submode, floored by
// Coin. Defensive coin fallback when nothing is selected yet (Validate fires
// before the binding resolves on the first paint).
func reinforceLimit(capsByBase map[string]adapter.ReinforceTarget, base *domain.Base, sub action.ReinforceMode, coin int) int {
	if base == nil {
		return coin
	}
	target := capsByBase[base.ID]
	limit := coin
	switch sub {
	case action.ReinforceHeal:
		if target.HealCap < limit {
			limit = target.HealCap
		}
	case action.ReinforceMax:
		if target.MaxCap < limit {
			limit = target.MaxCap
		}
	}
	return limit
}

// amountValidator parses a positive-integer Coin amount in [lo, hi]. Local to
// expand.go (two inputs); Bribe's single input inlines its own (matched, not
// extracted — pre-merge consolidation item).
func amountValidator(lo, hi int) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("enter a number")
		}
		if n < lo || n > hi {
			return fmt.Errorf("%d–%d", lo, hi)
		}
		return nil
	}
}

// NewConfirmRivalFreeAttack asks (on the rival's behalf — single-GM tool) whether
// a tying/beating rival makes its free attack on the new base. Answer: bool.
func NewConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) ConfirmOverlay {
	title := fmt.Sprintf("Let %s make a free attack on the new base?", rival.Name)
	header := fmt.Sprintf("Contested roll — %s rolled %d vs your %d", rival.Name, rivalRoll, factionRoll)
	return NewConfirm(title, header, "Allow attack", "Decline")
}

// NewSelectBaseAttackers picks which of the rival's assets attack the new base.
// All eligible belong to the one rival, so labels need no owner. Answer:
// []*domain.Asset (uncapped).
func NewSelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(eligible))
	for _, asset := range eligible {
		opts = append(opts, huh.NewOption(attackerLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Commit which attackers?", fmt.Sprintf("%s's assets attacking the new base", rival.Name), opts, 0)
}

var _ Overlay = Expand{}
```

> `attackerLabel` is the shared helper from `attack.go` (Commit 8). The two follow-on constructors return generic archetypes (no own `var _ Overlay`).

##### Task 3 — `action_collector.go`: replace the three Expand stubs and add `deriveReinforceTargets`. **Find:**
```go
func (a *actionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	return action.ExpandInfluenceOrder{}, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	return false, action.ErrActionUnavailable
}
func (a *actionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectExpandInfluenceOrder(faction *domain.Faction, factionState *state.FactionState, eligibleNewBaseWorlds []string) (action.ExpandInfluenceOrder, error) {
	names := make(map[string]string, len(eligibleNewBaseWorlds))
	for _, worldID := range eligibleNewBaseWorlds {
		if loc, ok := a.adapter.engine.World.Location(worldID); ok {
			names[worldID] = loc.Name()
		}
	}
	raw, err := a.adapter.ask(AskSelectExpandInfluenceOrder, faction, SelectExpandInfluenceOrderPayload{
		NewBaseWorlds:  eligibleNewBaseWorlds,
		WorldNames:     names,
		ReinforceBases: deriveReinforceTargets(faction),
		Coin:           faction.Coin,
	})
	if err != nil {
		return action.ExpandInfluenceOrder{}, err
	}
	order, _ := raw.(action.ExpandInfluenceOrder)
	return order, nil
}
func (a *actionCollector) ConfirmRivalFreeAttack(rival *domain.Faction, rivalRoll, factionRoll int) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmRivalFreeAttack, nil, ConfirmRivalFreeAttackPayload{
		Rival:       rival,
		RivalRoll:   rivalRoll,
		FactionRoll: factionRoll,
	})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
func (a *actionCollector) SelectBaseAttackers(rival *domain.Faction, eligible []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectBaseAttackers, nil, SelectBaseAttackersPayload{
		Rival:    rival,
		Eligible: eligible,
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}

// deriveReinforceTargets unions damagedNonHomeworldBases (heal-eligible) and
// growableNonHomeworldBases (max-eligible) from expand_influence.go into per-base
// caps. The two filters are independent: a base may be heal-eligible, max-eligible,
// both, or neither (excluded).
func deriveReinforceTargets(faction *domain.Faction) []ReinforceTarget {
	var targets []ReinforceTarget
	for _, base := range faction.Bases {
		if base.IsHomeworld {
			continue
		}
		healCap := base.EffectiveMaxHP(faction) - base.CurrentHP
		maxCap := faction.MaxHP - base.MaxHP
		canHeal := healCap > 0
		canMax := maxCap > 0
		if !canHeal && !canMax {
			continue
		}
		targets = append(targets, ReinforceTarget{
			Base:    base,
			CanHeal: canHeal,
			HealCap: healCap,
			CanMax:  canMax,
			MaxCap:  maxCap,
		})
	}
	return targets
}
```

##### Task 4 — `execution.go`: add the three arms and labels.

(a) **Find** (the redirect arm as left by Commit 8):
```go
	case adapter.AskConfirmRedirectToBase:
		p := msg.Payload.(adapter.ConfirmRedirectToBasePayload)
		return overlay.NewConfirmRedirectToBase(p.DefenderFaction, p.Base, p.Damage)
```
**Replace with:**
```go
	case adapter.AskConfirmRedirectToBase:
		p := msg.Payload.(adapter.ConfirmRedirectToBasePayload)
		return overlay.NewConfirmRedirectToBase(p.DefenderFaction, p.Base, p.Damage)
	case adapter.AskSelectExpandInfluenceOrder:
		p := msg.Payload.(adapter.SelectExpandInfluenceOrderPayload)
		return overlay.NewExpand(p.NewBaseWorlds, p.WorldNames, p.ReinforceBases, p.Coin)
	case adapter.AskConfirmRivalFreeAttack:
		p := msg.Payload.(adapter.ConfirmRivalFreeAttackPayload)
		return overlay.NewConfirmRivalFreeAttack(p.Rival, p.RivalRoll, p.FactionRoll)
	case adapter.AskSelectBaseAttackers:
		p := msg.Payload.(adapter.SelectBaseAttackersPayload)
		return overlay.NewSelectBaseAttackers(p.Rival, p.Eligible, m.rulebook)
```

(b) **Find** (the redirect label as left by Commit 8):
```go
	case adapter.AskConfirmRedirectToBase:
		return "Redirect to Base?"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskConfirmRedirectToBase:
		return "Redirect to Base?"
	case adapter.AskSelectExpandInfluenceOrder:
		return "Expand Influence"
	case adapter.AskConfirmRivalFreeAttack:
		return "Rival Free Attack?"
	case adapter.AskSelectBaseAttackers:
		return "Base Attackers"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go` (`"Expand Influence"` is 16 chars, same width as `"Change Homeworld"` — no re-alignment). **Find:**
```go
	"Attack":           true,
}
```
**Replace with:**
```go
	"Attack":           true,
	"Expand Influence": true,
}
```

##### Commit message
```
feat(tui/turn): Expand Influence action

- expand wizard (new-base: world+HP / reinforce: base+submode+HP) with branch
  gating; Confirm rival-free-attack (shows both rolls); MultiSelect base attackers
- real SelectExpandInfluenceOrder (derives reinforce bases) / ConfirmRivalFreeAttack
  / SelectBaseAttackers adapter methods; enable Expand Influence in the picker
```

---

### Commit 10 — `feat(tui/turn): Use Asset Ability action`

Three prompts (**Opus**) across the `ability` dispatch surface: `SelectAbilityAssets` (commit ability-capable assets), then per asset `ConfirmAbilityApplied` (ack — bool discarded, OQ) and, for the informers ability, `SelectFactionTestTarget` (`use_asset_ability.go:55`, `ability/dispatch.go:63`, `ability/informers.go:23`). All forwarding.

##### Task 0 — ground the prompt surface (do this first)

Re-read `use_asset_ability.go` **and the `ability/` subpackage** (`ability/dispatch.go`, `ability/informers.go`) — the collector reaches across the dispatch surface, so a prompt can hide in an ability handler, not just the action file. Confirm the reachable set is **exactly `SelectAbilityAssets`, `ConfirmAbilityApplied`, `SelectFactionTestTarget`**, that `ConfirmAbilityApplied`'s bool is still discarded (`ability/dispatch.go:63` — the one-sided-ack OQ), and that `SelectFactionTestTarget` is reached only on the informers path. Ability tests **roll**, so `SelectModifiers`/`ConfirmReroll` (foundation-built) may fire — confirm reachable, do **not** re-wire. If any dispatch path reaches a collector method outside this set, stop and reconcile the plan before writing code.

##### Task 1 — `channels.go`: add three ask kinds and payloads.

(a) **Find** (the iota tail as left by Commit 9):
```go
	AskSelectBaseAttackers
)
```
**Replace with:**
```go
	AskSelectBaseAttackers
	AskSelectAbilityAssets
	AskConfirmAbilityApplied
	AskSelectFactionTestTarget
)
```

(b) **Find** (the `SelectBaseAttackersPayload` as left by Commit 9):
```go
type SelectBaseAttackersPayload struct {
	Rival    *domain.Faction
	Eligible []*domain.Asset
}
```
**Replace with:**
```go
type SelectBaseAttackersPayload struct {
	Rival    *domain.Faction
	Eligible []*domain.Asset
}
type SelectAbilityAssetsPayload struct{ Candidates []*domain.Asset }
type ConfirmAbilityAppliedPayload struct {
	Asset *domain.Asset
	Def   *domain.AssetDefinition
}
type SelectFactionTestTargetPayload struct {
	Asset      *domain.Asset
	Effect     domain.AbilityEffectType
	Candidates []*domain.Faction
}
```

##### Task 2 — `internal/faction/tui/views/turn/overlay/ability.go` (new) — three constructors over the generic archetypes — full contents:

```go
package overlay

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)

// NewSelectAbilityAssets builds the multi-select of A-flagged assets whose
// abilities the faction activates (committed up front, resolved in order).
// Labels carry the ability effect. Answer: []*domain.Asset (uncapped). Esc cancels.
func NewSelectAbilityAssets(candidates []*domain.Asset, rb *rulebook.Rulebook) MultiSelectOverlay[*domain.Asset] {
	opts := make([]huh.Option[*domain.Asset], 0, len(candidates))
	for _, asset := range candidates {
		opts = append(opts, huh.NewOption(abilityLabel(asset, rb), asset))
	}
	return NewMultiSelect[*domain.Asset]("Activate which abilities?", "", opts, 0)
}

// NewConfirmAbilityApplied is the per-asset acknowledgment. The engine discards
// the returned bool (ability/dispatch.go:63), so "Skip" is inert — the modal only
// gates and paces resolution (OQ). Answer: bool. Esc cancels the action.
func NewConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) ConfirmOverlay {
	return NewConfirm(fmt.Sprintf("Apply %s ability?", def.Name), def.Description, "Apply", "Skip")
}

// NewSelectFactionTestTarget picks the opposed-test target for an ability (e.g.
// Informers' reveal-stealth). The header names the effect. Answer: *domain.Faction.
// Esc cancels.
func NewSelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) SelectOverlay[*domain.Faction] {
	opts := make([]huh.Option[*domain.Faction], 0, len(candidates))
	for _, faction := range candidates {
		opts = append(opts, huh.NewOption(faction.Name, faction))
	}
	return NewSelect[*domain.Faction]("Target which faction?", fmt.Sprintf("Ability effect: %s", effect), opts)
}

// abilityLabel renders "<name> · <effect>" (or just the name when the definition
// carries no ability effect — e.g. the structural-stub abilities).
func abilityLabel(asset *domain.Asset, rb *rulebook.Rulebook) string {
	name := assetName(asset, rb)
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && def.Ability.Effect != "" {
		return fmt.Sprintf("%s · %s", name, def.Ability.Effect)
	}
	return name
}
```

> The constructors return generic archetypes (no own `var _ Overlay`). `assetName` is the movement.go helper.

##### Task 3 — `action_collector.go`: replace the three Ability stubs (note `SelectFactionTestTarget` sits apart from the other two in the Commit-1 stub block).

(a) **Find** (the lone `SelectFactionTestTarget` stub — first action-unique method):
```go
func (a *actionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	return nil, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
		Asset:      asset,
		Effect:     effect,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Faction)
	return chosen, nil
}
```

(b) **Find** (the contiguous `SelectAbilityAssets` + `ConfirmAbilityApplied` stubs):
```go
func (a *actionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	return nil, action.ErrActionUnavailable
}
func (a *actionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	return false, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (a *actionCollector) SelectAbilityAssets(faction *domain.Faction, candidates []*domain.Asset, rb *rulebook.Rulebook) ([]*domain.Asset, error) {
	raw, err := a.adapter.ask(AskSelectAbilityAssets, faction, SelectAbilityAssetsPayload{Candidates: candidates})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.([]*domain.Asset)
	return chosen, nil
}
func (a *actionCollector) ConfirmAbilityApplied(asset *domain.Asset, def *domain.AssetDefinition) (bool, error) {
	raw, err := a.adapter.ask(AskConfirmAbilityApplied, nil, ConfirmAbilityAppliedPayload{Asset: asset, Def: def})
	if err != nil {
		return false, err
	}
	confirmed, _ := raw.(bool)
	return confirmed, nil
}
```

##### Task 4 — `execution.go`: add the three arms and labels.

(a) **Find** (the base-attackers arm as left by Commit 9):
```go
	case adapter.AskSelectBaseAttackers:
		p := msg.Payload.(adapter.SelectBaseAttackersPayload)
		return overlay.NewSelectBaseAttackers(p.Rival, p.Eligible, m.rulebook)
```
**Replace with:**
```go
	case adapter.AskSelectBaseAttackers:
		p := msg.Payload.(adapter.SelectBaseAttackersPayload)
		return overlay.NewSelectBaseAttackers(p.Rival, p.Eligible, m.rulebook)
	case adapter.AskSelectAbilityAssets:
		return overlay.NewSelectAbilityAssets(msg.Payload.(adapter.SelectAbilityAssetsPayload).Candidates, m.rulebook)
	case adapter.AskConfirmAbilityApplied:
		p := msg.Payload.(adapter.ConfirmAbilityAppliedPayload)
		return overlay.NewConfirmAbilityApplied(p.Asset, p.Def)
	case adapter.AskSelectFactionTestTarget:
		p := msg.Payload.(adapter.SelectFactionTestTargetPayload)
		return overlay.NewSelectFactionTestTarget(p.Asset, p.Effect, p.Candidates)
```

(b) **Find** (the base-attackers label as left by Commit 9):
```go
	case adapter.AskSelectBaseAttackers:
		return "Base Attackers"
	default:
		return "Action"
	}
```
**Replace with:**
```go
	case adapter.AskSelectBaseAttackers:
		return "Base Attackers"
	case adapter.AskSelectAbilityAssets:
		return "Ability Assets"
	case adapter.AskConfirmAbilityApplied:
		return "Apply Ability?"
	case adapter.AskSelectFactionTestTarget:
		return "Ability Target"
	default:
		return "Action"
	}
```

##### Task 5 — `overlay/registry.go` (`"Use Asset Ability"` is now the widest key — full-block re-align). **Find** (the whole map as left by Commit 9):
```go
var ImplementedActions = map[string]bool{
	"Sell Asset":       true, // shared SelectAsset only (built in foundation)
	"Repair Faction":   true, // zero-prompt
	"Abandon Goal":     true, // zero-prompt
	"Buy Asset":        true,
	"Refit Asset":      true,
	"Repair Asset":     true,
	"Bribe":            true,
	"Seize Planet":     true,
	"Change Homeworld": true,
	"Attack":           true,
	"Expand Influence": true,
}
```
**Replace with:**
```go
var ImplementedActions = map[string]bool{
	"Sell Asset":        true, // shared SelectAsset only (built in foundation)
	"Repair Faction":    true, // zero-prompt
	"Abandon Goal":      true, // zero-prompt
	"Buy Asset":         true,
	"Refit Asset":       true,
	"Repair Asset":      true,
	"Bribe":             true,
	"Seize Planet":      true,
	"Change Homeworld":  true,
	"Attack":            true,
	"Expand Influence":  true,
	"Use Asset Ability": true,
}
```

##### Commit message
```
feat(tui/turn): Use Asset Ability action

- MultiSelect ability assets, Confirm ability-applied (ack), Select faction
  test target — ability-context labels
- real SelectAbilityAssets/ConfirmAbilityApplied/SelectFactionTestTarget adapter
  methods; enable Use Asset Ability in the picker — all 12 actions now live
```

## Verification

The overlays are `huh`/`bubbletea` views (no prescribed automated harness — same as Effort 2); verify by running real cycles plus targeted unit tests on the pure helpers:

- **Per-commit smoke:** launch Turn against a campaign seeded so the action under test **validates** (e.g. a damaged asset for Repair, ≥2 factions co-located for Attack/Seize, a non-homeworld base for Change Homeworld/Expand-reinforce, Coin for Buy/Bribe). Confirm the action appears **enabled** in the picker, the modal opens in the wizard band, the answer round-trips, and the event stream shows the resolved mutations.
- **Disable-set:** before a per-action commit, its action shows "(not yet available)" and selecting it re-prompts (Validate reject); the recoverable stub is unreachable. After the commit, it's selectable.
- **Esc-cancel:** `Esc` on any action overlay surfaces a recoverable error in the bottom bar (not fatal), logs it in the event stream, and the cycle continues to the next faction. `Esc` on a hook modal is a silent no-op (none applied / skip reroll).
- **Unit candidates for `_test.go`:** the adapter derivation helpers (`deriveContestedWorlds`, `deriveHomeworldTargets`, the reinforce-base filters) and any composite's answer-assembly (`BuyOrder`/`RefitOrder`/`[]RepairOrder`/`ExpandInfluenceOrder`) — pure functions worth pinning. Extend `execution_test.go` / add `overlay/*_test.go` where an assembler encodes a non-trivial invariant (cost caps, the Issue/Revise-style both-fields cases).
- **Full-action regression:** one end-to-end cycle exercising several actions across factions, confirming state + history persist (the engine owns persistence; the TUI adds nothing).

## Reference Exemplars

- `internal/faction/tui/views/turn/overlay/{statraise,movement,cargo}.go` — the Effort-2 `huh` idiom (heap data, pointer binding, `StateCompleted` emit, multi-group `HideFunc` paging) the archetypes generalize.
- `internal/faction/tui/views/turn/execution/execution.go:135-150,408-432` — the `CollectorAskMsg`/`OverlayDoneMsg` round-trip and `newOverlay` factory each commit extends.
- `internal/faction/tui/adapter/{channels,phase_collector,action_collector,adapter}.go` — the ask seam, the stub being replaced, the wiring point.
- `internal/faction/engine/action/{collector.go,action.go}` + `actions/*.go` — the collector contract, the per-action `Inputs`/`Resolve` flows, and the candidate-derivation each adapter method mirrors.
- `internal/faction/engine/orchestrator.go:388-482` + `errors/recoverable.go` — the recoverable/fatal classification the `Esc`-cancel and `ErrActionUnavailable` paths depend on.
- `internal/faction/engine/hooks/{collector.go,types.go,dispatch/roll.go}` — the two hook prompts and where they fire mid-roll.
