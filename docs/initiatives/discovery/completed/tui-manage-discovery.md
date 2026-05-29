# TUI Manage — Discovery

`F-005.2`, Initiative 2 of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md). Companion to [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md) and [`tui-foundation-discovery.md`](./tui-foundation-discovery.md) — Manage inherits the engine boundary, root Model dispatch, mode-bar router, and adapter sub-package from those docs and does not re-litigate them. This doc ratifies the **initiative-specific** decisions Manage needs before Plan: the internal view-stack representation, the state-package CRUD surface Manage consumes, the context strip's structural placement and content, and the help-overlay wiring.

## Problem

Manage is the first feature workflow filled into a mode shell. Foundation shipped the root Model dispatch, the mode-bar router, and the adapter, but `views/manage/manage.go` is a single-banner placeholder. After Manage ships, the TUI replaces hand-edited TOML for the entire faction authoring loop: list, detail, create, edit, delete.

Two structural concerns concentrate here:

1. **CRUD is the only TUI workflow that does not cross the engine boundary.** Arc-discovery framed the TUI as a pure consumer of the engine; Manage is the exception. The mistake to avoid is letting the TUI become the *new* canonical store — V1's failure mode applied at the state layer rather than the engine layer. The defense: Manage holds a pointer to `*state.FactionState` only as a hand-off; **CRUD logic lives in the `state` package**, not in the TUI. Manage calls `state.Create/Update/Delete`; the TUI never writes the map or calls `state.Save` itself.

> Question: Do we want the TUI to hold a reference to `*state.FactionState` at all, or do we want to pass the relevant data into each sub-model and have the state package's CRUD functions take care of loading/saving the TOML? The former is more efficient and straightforward; the latter is a stronger encapsulation of the state layer. .

2. **Manage's view-stack pattern is also Turn's template.** Initiative 3 will follow the same sub-model-per-view shape for `setup → execution` and the modal overlay. Decisions made here about navigation, focus, and Esc semantics propagate. The empty-state grammar Manage picks is explicitly inherited by Turn per arc-plan.

The view stack itself is small (`list`, `detail`, `edit`, `create`, `deleteConfirm`) and back-edges are statically known — no general stack machinery is warranted. The form scope (identity-only vs. inline-goals/assets) is the one CRUD shape question Plan must resolve; everything else is structural and ratified here.

## Design Summary

Manage is built as a **router sub-model** at `internal/faction/tui/views/manage/`, one level below the root Model's mode dispatch. `manage.Model` holds a `view` enum (`viewList`, `viewDetail`, `viewEdit`, `viewCreate`, `viewDeleteConfirm`) and a `tea.Model` per view; Update and View route to the active sub-model. Each sub-view is its own package under `views/manage/` (e.g., `views/manage/list/`, `views/manage/detail/`, etc.) with the standard `Init` / `Update` / `View` quartet.

**State surface.** `manage.Model` is constructed with `manage.New(factionState *state.FactionState, paths *campaigns.Paths)`. It holds these references as a hand-off into the `state` package's CRUD calls — **not as a working buffer.** All CRUD goes through new functions on the `state` package: `state.CreateFaction(path, fs, faction)`, `state.UpdateFaction(path, fs, faction)`, `state.DeleteFaction(path, fs, id)`, `state.GetFaction(fs, id)`. Each mutation is **write-through** — the in-memory map and the on-disk TOML are updated in one call. Manage never writes `fs.Factions` directly and never calls `state.Save` itself.

**View stack flow.** Forward transitions are key-driven (`Enter` on a list row → detail, `e` from detail → edit, `n` from list → create, `d` from detail → delete-confirm). Each forward transition constructs the next sub-model with the data it needs (`detail.New(currentFaction)`, `edit.New(currentFaction)`, etc.). Back-edges are statically known per view: `detail → list`, `edit → detail`, `create → list`, `deleteConfirm → detail`. Sub-models emit **completion messages** (e.g., `EditSavedMsg{Faction}`, `CreatedMsg{Faction}`, `DeletedMsg{ID}`, `CancelMsg`) via `tea.Cmd`; `manage.Update` handles these centrally — calls the appropriate state CRUD function, then sets `m.view` to the back-target. Sub-models don't perform persistence or transitions themselves.

**Esc semantics.** Esc on `edit` or `create` *always* opens a discard-confirm prompt, even when the form has no changes. The prompt is a transient overlay on the active form (Plan picks whether it's a sub-view in manage's stack or internal state on edit/create). Esc on `detail` returns to `list` directly (no prompt). Esc on `deleteConfirm` cancels and returns to `detail`. Esc on `list` is inert — `list` is the top of Manage's stack; the user uses `Tab` to leave the mode.

**Context strip.** A separate component at `views/manage/contextstrip/`, composed by `manage.View` to the right of `list.View` when `view == viewList`. The strip is hidden during detail/edit/create/deleteConfirm — those views take over the full content area below the mode bar. MVP content: **Campaign ID**, **Cycle (current / next)**, **Faction count**. Save and load errors surface **inline in the active view**, not in the strip — the strip stays calm.

**Help overlay.** Each Manage sub-model exposes a `Help() help.KeyMap` method (standard `bubbles/help` pattern). The root TUI's existing `?` handler reads the active sub-model's bindings and assembles them with the globals (`Tab`, `Shift-Tab`, `?`, `q`) into the overlay. Foundation's `showHelpStub` is extended; the per-view content is sourced from each sub-model rather than centralized.

**Empty state.** Per arc-discovery's sketch, Manage's list shows a **non-blocking hint** when `len(Factions) == 0`. The structural choice is ratified here; copy, key prompt, and visual treatment are Plan's call (and Turn's setup view inherits the same grammar).

## Decisions Ratified in Discovery

1. **`manage.Model` is a router with sub-model-per-view.** Each view (`list`, `detail`, `edit`, `create`, `deleteConfirm`) is its own `tea.Model` package under `views/manage/`. `manage.Model` holds a `view` enum + the sub-model instances and dispatches Update/View to the active one. Mirrors the root mode-bar router one level down.
2. **Esc on `edit` and `create` always opens a discard-confirm prompt.** No "only when dirty" gate; the prompt fires unconditionally. Plan picks whether the prompt is a sub-view in Manage's stack or internal state on edit/create.
3. **CRUD lives in the `state` package, not the TUI.** New write-through functions: `state.CreateFaction(path, fs, faction) error`, `state.UpdateFaction(path, fs, faction) error`, `state.DeleteFaction(path, fs, id) error`, `state.GetFaction(fs, id) (*domain.Faction, error)`. Each mutating call updates the in-memory map and persists to disk in one operation. Manage never writes the map or calls `state.Save` directly.
4. **`manage.Model` holds `*state.FactionState` and `*campaigns.Paths`** as hand-off references into the state package's CRUD calls. They are passed into `manage.New(...)` from the root Model.
5. **Sub-models emit completion messages; `manage.Update` handles transitions and persistence.** Sub-models do not call state CRUD themselves and do not set `m.view`. The completion-message pattern keeps the persist-then-transition sequence in one place.
6. **Forward transitions construct fresh sub-models** with the data they need (`edit.New(faction)`, `delete.New(id)`, etc.). Back-edges are statically known per view; no general stack machinery.
7. **Context strip is its own component at `views/manage/contextstrip/`**, composed by `manage.View` to the right of `list.View` when `view == viewList`. Hidden in all other Manage views.
8. **Strip MVP content: Campaign ID, Cycle number (current / next), Faction count.** No last-error preview in the strip — errors surface inline in the active view.
9. **Help overlay reads per-view bindings from the active sub-model** via a `Help() help.KeyMap` method on each sub-model. Standard `bubbles/help` idiom. Root combines with globals.
10. **Empty list shows a non-blocking hint** — the user is not forced into create. The structural pattern is ratified; exact copy and visual treatment are Plan's.

## Open Questions

1. **Faction CRUD form scope.** What is editable in `create` vs. `edit`? Identity and stats only, with goals and assets handled in their own (future) sub-views — or inline at create time so a new faction can ship configured? The narrower scope ships sooner and matches the deferred-tabs scope of detail. The broader scope mirrors how factions actually exist on disk and avoids a partially-created state in the TOML. Plan picks.
2. **Empty-state UX copy and visual treatment.** Structural choice (non-blocking hint on the list) is ratified above. Plan picks the exact hint text, the keybinding prompt rendering, and the lipgloss styling — Turn's setup-view empty state inherits the grammar.

## Per-Area Design Details

### View stack and sub-model layout

Manage's sub-views live in their own packages so each can hold view-local state, its own `tea.Model` quartet, and its own `Help()` method without sharing a struct:

```
internal/faction/tui/views/manage/
├── manage.go              -- router: Model, view enum, Update/View dispatch
├── messages.go            -- completion message types (EditSavedMsg, CreatedMsg, etc.)
├── list/                  -- faction list (table + selection state + n/Enter keys)
├── detail/                -- single-faction detail view (e/d/Esc keys)
├── edit/                  -- edit form + draft buffer (huh-driven)
├── create/                -- create form + draft buffer (huh-driven)
├── deleteconfirm/         -- y/n confirm for delete
└── contextstrip/          -- strip component composed by manage.View on list
```

`manage.Model` fields (sketch):

```go
type view int

const (
    viewList view = iota
    viewDetail
    viewEdit
    viewCreate
    viewDeleteConfirm
)

type Model struct {
    factionState *state.FactionState
    paths        *campaigns.Paths

    view          view
    list          list.Model
    detail        detail.Model
    edit          edit.Model
    create        create.Model
    deleteConfirm deleteconfirm.Model
    strip         contextstrip.Model
}
```

The discard-confirm overlay's structural home (sub-view in Manage's stack vs. internal state on edit/create) is a Plan call. Discovery only commits to the always-prompt behavior.

### State-package CRUD additions

The `state` package gains a new file (or extends `faction_state.go`) with write-through CRUD on `*FactionState`:

```go
// internal/faction/state/crud.go (or extension of faction_state.go)

func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error
func UpdateFaction(path string, fs *FactionState, faction *domain.Faction) error
func DeleteFaction(path string, fs *FactionState, id string) error
func GetFaction(fs *FactionState, id string) (*domain.Faction, error)
```

Each mutating function:
1. Validates invariants (e.g., `Create` rejects a duplicate ID; `Update` and `Delete` reject an unknown ID).
2. Applies the mutation to `fs.Factions`.
3. Calls the existing persistence path (`Save(path, fs)` internally or its equivalent).
4. Returns any error from validation or persistence.

The exact contract — whether disk-write failure rolls back the in-memory change, error types/sentinels, validation rules — is Plan's. Discovery ratifies only the surface shape.

`GetFaction` is non-mutating and returns the live pointer; sub-models hold copies if they need to mutate (the `edit.Model`'s draft buffer is a deep copy of the current faction).

### Completion messages and the transition seam

Sub-models communicate results to `manage.Update` via typed messages emitted from `tea.Cmd`s:

```go
// internal/faction/tui/views/manage/messages.go

type EditSavedMsg struct{ Faction *domain.Faction }
type CreatedMsg struct{ Faction *domain.Faction }
type DeletedMsg struct{ ID string }
type CancelMsg struct{}                    // discard / Esc-from-edit-or-create / no-from-deleteConfirm
type SaveErrorMsg struct{ Err error }      // surfaced inline in the active view
```

`manage.Update` handles each:
- `EditSavedMsg` / `CreatedMsg` → call state CRUD; on success set `m.view` to the back-target; on failure forward `SaveErrorMsg` to the active sub-model for inline display.
- `DeletedMsg` → call `state.DeleteFaction`; on success transition to `list`; on failure forward `SaveErrorMsg` to `deleteConfirm`.
- `CancelMsg` → transition to the back-target with no state change.

Plan picks the exact constructor signatures, the back-target lookup (likely a static map keyed by `view`), and whether `SaveErrorMsg` carries view-local context (e.g., the field that failed validation).

### Context strip composition

`manage.View()` renders the strip only when `view == viewList`:

```go
func (m Model) View() string {
    switch m.view {
    case viewList:
        return lipgloss.JoinHorizontal(
            lipgloss.Top,
            m.list.View(),
            m.strip.View(),
        )
    case viewDetail:
        return m.detail.View()
    // ... edit, create, deleteConfirm take over similarly
    }
}
```

The strip's `Update` is driven by manage forwarding state-changing messages (e.g., after a successful CRUD operation, `manage.Update` calls `m.strip = m.strip.Refresh(m.factionState)` or equivalent — Plan picks the refresh shape).

Strip layout (MVP) is three lines:
- Line 1: Campaign ID.
- Line 2: Cycle (`current → next`).
- Line 3: Faction count.

Exact lipgloss styling, vertical alignment, and width budget are Plan-level.

### Help overlay wiring

Each sub-model exposes:

```go
func (m list.Model) Help() help.KeyMap
func (m detail.Model) Help() help.KeyMap
// etc.
```

`bubbles/help` is the standard Charm package; the keymap is a struct of `key.Binding` values. The root `view.go`'s help-rendering branch (currently `m.showHelpStub` returning a fixed string) is extended to:

1. Read globals from the root Model.
2. Read the active mode's active sub-model's `Help() help.KeyMap`.
3. Compose them into the overlay output via `help.View`.

The dispatch from root to active mode's active sub-model is one additional hop the root Model takes when `?` is pressed; Plan picks whether the root asks the mode sub-model for its active sub's Help() or whether modes implement a passthrough.

## State Storage

No new on-disk schema. `*state.FactionState` and its TOML serialization are unchanged. The state package's exported surface gains CRUD functions (`CreateFaction`, `UpdateFaction`, `DeleteFaction`, `GetFaction`); `Load` and `Save` keep their current shape.

## User-Facing Impact

After F-005.2 lands, the GM can:

- Open `gm-toolkit faction` and see a list of all factions in the active campaign, with a context strip showing campaign / cycle / count.
- Press `Enter` on a row to view that faction's full state.
- Press `n` from the list to create a new faction via a `huh` form.
- Press `e` from a faction's detail to edit that faction via a `huh` form.
- Press `d` from a faction's detail and confirm to delete it.
- Press `Esc` from an edit/create form and confirm to discard changes.
- Press `?` from any view to see view-local keybindings plus globals.
- Press `Tab` from the list to switch into the Turn mode shell (still empty until Initiative 3).

What still does **not** work after this initiative: anything Turn-related (setup, execution, collector prompts), tabbed faction-detail, direct state edits outside the engine pipeline.

## Out of Scope

Manage explicitly does **not** ship:

- **Anything Turn-related.** Initiative 3.
- **Tabbed faction-detail** (Stats / Assets / Goals / History). Sub-view growth deferred per arc-discovery's Extensibility section.
- **`Adapter.Stop()` cancellation design.** Routed to Initiative 3 — the panic risk (`close(eventCh)` while engine goroutine is still emitting) materializes only during an active engine cycle. Manage never calls `adapter.Run()`, so its quit path is safe with the current implementation.
- **Save-error rollback semantics.** Whether a failed `state.Save` reverts the in-memory mutation is Plan-level; for MVP, "return the error and let the user retry" is sufficient if the user-facing surface (inline error in the active view) is clear.
- **Filter / sort / search on the list.** A long faction list scrolls; no per-field filter, no fuzzy search. If demand surfaces, a deferred item in `planned-work.md`.
- **Multi-select / bulk operations.** One faction at a time.
- **Concurrency-with-Turn.** Whether Tab is suppressed during a running cycle, whether Manage saves are blocked mid-cycle — Turn's problem.
- **Validation beyond invariants.** Stat ranges, name length limits, ID format constraints — handled by `huh` form validators if needed, not by state CRUD. State CRUD enforces only what data integrity demands (unique IDs on create, existence on update/delete).

## Reference Exemplars

- **Arc-Discovery — Manage section** (`tui-rebuild-arc-discovery.md`, "View Structure → Manage") — the detail-on-Enter pattern, context strip placement, draft-buffer-on-save framing.
- **Arc-Discovery — Update discipline** (same doc, "Architecture → Update discipline") — Update may emit `tea.Cmd`s for side-effects but may not call state CRUD inline. Completion-message pattern lives on top of this.
- **Arc-Discovery — Extensibility Shape — New sub-views within a mode** (same doc) — Manage establishes the sub-model-per-view pattern; tabbed faction-detail is the next-anticipated growth that slots into this pattern.
- **Arc-Plan — Initiative 2 entry** (`tui-rebuild-arc-plan.md`, "Initiative 2 — `tui-manage`") — the In Scope / Out of Scope boundary this Discovery refines.
- **Foundation-Discovery — Root Model dispatch** (`tui-foundation-discovery.md`, "Per-Area Design Details → Mode-bar router") — Manage's router pattern is the same one level down.
- **`internal/faction/state/faction_state.go`** — the existing `Load`/`Save` API that CRUD extensions sit alongside.
- **`internal/campaigns/`** — `*Paths` carries `StatePath`; Manage receives `*Paths` from the root Model and hands `paths.StatePath` into state CRUD calls.
- **Charm `bubbles/help`** — the standard help-overlay pattern; Manage's per-view `Help()` returns a `help.KeyMap`.
