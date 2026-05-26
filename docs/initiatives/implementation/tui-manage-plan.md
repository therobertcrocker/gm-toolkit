# tui-manage — Implementation Plan

## Context / Goal

This plan turns [`tui-manage-discovery.md`](../discovery/tui-manage-discovery.md) into an executable sequence of commits. Manage is Initiative 2 of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md). After this initiative ships, the TUI replaces hand-edited TOML for the entire faction authoring loop — list, detail, create (full SWN creation wizard), edit, delete.

Discovery ratified the structural decisions — router-with-sub-model-per-view, state-package CRUD ownership, completion-message pattern, context-strip placement, help-overlay shape. This plan ratifies the four open questions Discovery deferred, settles two implementation-time calls Discovery left to Plan (refresh shape, rollback semantics), and breaks the work into commits sized one-per-execution-session.

**Related artifacts:**
- Discovery: [`tui-manage-discovery.md`](../discovery/tui-manage-discovery.md)
- Arc-Plan: [`tui-rebuild-arc-plan.md`](../arcs/tui-rebuild/tui-rebuild-arc-plan.md)
- Arc-Discovery: [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md)
- Foundation Plan (exemplar): [`completed/tui-foundation-plan.md`](./completed/tui-foundation-plan.md)
- SWN creation checklist: [`docs/rules/swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md) § Faction Creation Checklist

## Decisions Ratified in Planning

1. **Create wizard walks the full SWN Faction Creation Checklist.** Resolves Discovery's "Faction CRUD form scope" open question — Robert chose the broader scope over identity-only. The wizard is a multi-step `huh.Form` covering: Scale, Attribute assignment (which stat is primary/secondary/tertiary), HP display (derived `huh.Note`), Tags (1–2 from rulebook), Starting Goal (from rulebook), Homeworld (world picker from spatial), Starting Assets (filtered by rating + tech-level, quota per scale), Starting Coin (GM discretion, numeric input). The auto-derived steps (HP, homeworld Base of Influence at max HP) appear as confirmation notes, not interactive inputs.
2. **Edit reuses the create wizard with pre-population.** Same multi-step form, different initial draft buffer. Create starts from `domain.Faction{}` plus Scale defaults; edit starts from a deep copy of the current faction. Constructor flag determines completion message — `CreatedMsg` vs `EditSavedMsg`. One wizard package, two entry points.
3. **Discard-confirm is internal state on edit/create models.** Resolves Discovery's "structural home" choice. Each form gains `confirmingDiscard bool`; View overlays a y/n prompt when set; Update intercepts y/n locally and emits `CancelMsg` on y. No new package, no `viewDiscardConfirm` router branch. The prompt fires unconditionally per Discovery (no "only-when-dirty" gate).
4. **Help dispatch: modes implement a `Help()` passthrough.** Resolves Discovery's "root vs. passthrough" choice. Manage's `Help() help.KeyMap` returns its active sub-model's `Help()`. Root holds a `Helper` interface; `?` handler calls `subs[m.bar.Active()].(Helper).Help()` and composes with globals. Root stays agnostic of any mode's internal view-stack shape; Turn slots in the same way.
5. **Empty-list grammar: centered hint with key prompt.** Resolves Discovery's "empty-state UX" choice. List area renders `"No factions yet."` in normal style on one line, dimmed `"Press n to create one"` below. Turn (Initiative 3) inherits: `"No factions in roster."` + `"Press n in Manage to add one."`.
6. **Spatial data loaded alongside rulebook in `factionRun`.** The wizard's homeworld picker needs the world list; the wizard's asset step needs each world's `TechLevel()`. `cmd/gm-toolkit/faction.go` adds `spatial.LoadRegionMap(paths.SpatialDataDir)` after the rulebook load; `tui.Run` signature expands to take a `*spatial.RegionMap`. Manage receives it via the root Model. Foundation's existing data flow (rulebook → tui.Run → root → adapter) is the template — spatial threads the same path. If `LoadRegionMap` returns an error (no spatial seeded), `factionRun` surfaces it with a `hint:` line per the existing rulebook error pattern; the seed command is TBD per the F-012 backlog item, so the hint text is left to execution-time re-grounding.
7. **State CRUD lives in a new file: `internal/faction/state/crud.go`.** Keeps `faction_state.go`'s `Load`/`Save` contract clean; CRUD is its own concern. Functions and sentinels declared there (signatures in Shared Context). Save errors do not roll back the in-memory mutation — per Discovery's MVP framing. Inline `SaveErrorMsg` surfaces the error; user retries. Revisit only if execution surfaces brittleness.
8. **Context strip refresh: `manage.Update` reassigns the strip after a successful CRUD message.** After `CreatedMsg` / `EditSavedMsg` / `DeletedMsg` is handled (state CRUD succeeds), `manage.Update` calls `m.strip = contextstrip.New(m.factionState)`. The strip is a stateless snapshot constructor — cheaper than a stateful `Refresh()` method, and the recompute cost (read three fields from FactionState) is trivial.
9. **`bubbles/list` powers the list view; `bubbles/help` powers the help overlay.** Add these in the commits that first use them — list in Commit 2, help in Commit 7. No global "dependency bump" commit; tracked changes land with the consumer.
10. **List view's item shape: Name · Scale · current/max HP · Coin.** Single-column `list.Item`. Selection state and j/k/up/down scrolling inherited from `bubbles/list`. Faction `ID` shown dimmed in the detail view, not the list (factions are GM-facing; names are the working identifier).
11. **Seven commits across four phases.** Phase 1 (state CRUD): 1 commit. Phase 2 (router + read-only views): 2 commits. Phase 3 (wizard + edit): 2 commits. Phase 4 (presentation polish): 2 commits. Arc-Plan estimated 5–8; this lands within range. The Phase 3 create-wizard commit (Commit 4) is the heaviest — flagged for execution-time re-grounding and possible mid-phase split if it grows past one session.

## Open Questions — To Ratify at Implementation Time

1. **Spatial seed-command hint text** — `factionRun`'s error message for a missing spatial dir mirrors the rulebook pattern (`hint: gm-toolkit campaign ...`). The seed command for spatial data has not yet shipped (F-012 backlog). Re-grounding step at Commit 4 picks the placeholder text — likely `hint: (spatial seed command TBD — see F-012)` until F-012 lands.
2. **Wizard asset-step quota enforcement** — Per SWN: Minor = 1 primary + 1 any; Major = 2 primary + 2 any; Hegemon = 4 primary + 4 any. `huh.MultiSelect` validators can enforce the constraint, but the cleanest UX is two grouped multi-selects (one filtered to primary attribute, one filtered to all). Commit 4 picks at execution time based on what `huh.Group` chaining looks like in practice — both shapes are valid; choice is presentational.
3. **Faction ID generation strategy** — SWN doesn't specify IDs; existing factions in `state.toml` have IDs assigned by hand. Options: user-typed (huh.Input), auto-derived from Name (`slug(name)`), or auto-incremented (`F-001`, `F-002`). Plan defers to first use in Commit 4 — recommendation: derive from Name with a uniqueness check against `fs.Factions`, prompt the user only on collision.

## Shared Context

This section holds the structural definitions that multiple commits reference. Tasks under Work Breakdown point back here rather than re-stating struct shapes.

### Final Package Layout

```
gm-toolkit/
├── cmd/gm-toolkit/
│   └── faction.go                       # +spatial.LoadRegionMap, +pass to tui.Run
├── internal/
│   ├── faction/
│   │   ├── state/
│   │   │   ├── faction_state.go         # untouched
│   │   │   └── crud.go                  # NEW — Create/Update/Delete/Get + sentinels
│   │   └── tui/
│   │       ├── tui.go                   # Run signature +spatial param
│   │       ├── model.go                 # +spatial field; +manage.New(adapter, spatial)
│   │       ├── view.go                  # help-overlay branch replaces stub
│   │       ├── update.go                # ? handler routes through Helper
│   │       ├── helper.go                # NEW — Helper interface
│   │       └── views/manage/
│   │           ├── manage.go            # router: Model, view enum, dispatch, Help() passthrough
│   │           ├── messages.go          # NEW — completion message types
│   │           ├── list/
│   │           │   └── list.go          # bubbles/list + empty-state hint
│   │           ├── detail/
│   │           │   └── detail.go        # display all faction fields; e/d/Esc
│   │           ├── deleteconfirm/
│   │           │   └── deleteconfirm.go # y/n confirm
│   │           ├── wizard/
│   │           │   └── wizard.go        # huh multi-step form; create + edit entry points
│   │           └── contextstrip/
│   │               └── contextstrip.go  # campaign / cycle / faction count
│   └── spatial/                          # untouched (loader already exists)
```

The `views/manage/` directory replaces the placeholder `manage.go` shell shipped by Foundation. Foundation's `tui.Placeholder` style remains in use for sub-view stubs during incremental commits and for the (still-empty) Turn shell.

### Dependency Additions

Added incrementally with the consumer. No standalone "deps" commit.

- **Commit 2** — `github.com/charmbracelet/bubbles/list` (list view).
- **Commit 4** — no new direct deps; `huh` is already in `go.mod` from Foundation.
- **Commit 7** — `github.com/charmbracelet/bubbles/help` and `github.com/charmbracelet/bubbles/key` (help overlay + KeyMap).

Both `bubbles/*` packages are siblings of `bubbletea` and already in the module graph as transitive deps; `go get` to make them direct deps.

### State CRUD Function Signatures (Final Shape)

All in `internal/faction/state/crud.go`:

```go
package state

import (
    "errors"
    "fmt"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

var (
    ErrFactionAlreadyExists = errors.New("state: faction already exists")
    ErrFactionNotFound      = errors.New("state: faction not found")
    ErrInvalidFactionID     = errors.New("state: invalid faction id (empty)")
)

// CreateFaction validates uniqueness, applies the mutation to fs.Factions,
// then persists to path. Returns ErrFactionAlreadyExists on duplicate ID,
// ErrInvalidFactionID on empty ID. On Save failure, the in-memory mutation
// is NOT rolled back (caller surfaces the error and retries).
func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error

// UpdateFaction validates that faction.ID exists in fs.Factions, replaces
// the entry, then persists. Returns ErrFactionNotFound if absent.
func UpdateFaction(path string, fs *FactionState, faction *domain.Faction) error

// DeleteFaction validates that id exists in fs.Factions, removes the entry,
// then persists. Returns ErrFactionNotFound if absent.
func DeleteFaction(path string, fs *FactionState, id string) error

// GetFaction returns the live *domain.Faction for id (no copy). Callers that
// intend to mutate must deep-copy first.
func GetFaction(fs *FactionState, id string) (*domain.Faction, error)
```

Error wrapping follows the campaign-manager pattern (`fmt.Errorf("state: <op>: %w", err)`); sentinels are checked with `errors.Is`.

### Completion Message Types (Final Shape)

All in `internal/faction/tui/views/manage/messages.go`:

```go
package manage

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

type CreatedMsg struct{ Faction *domain.Faction }
type EditSavedMsg struct{ Faction *domain.Faction }
type DeletedMsg struct{ ID string }
type CancelMsg struct{} // discard from edit/create, no from deleteconfirm
type SaveErrorMsg struct{ Err error }
```

Sub-models emit these via `tea.Cmd`s; `manage.Update` handles each centrally — calls the appropriate `state.*Faction` function, then sets `m.view` to the back-target, then reassigns the strip. On error, forwards `SaveErrorMsg` to the active sub-model for inline display.

### `manage.Model` Shape

```go
package manage

import (
    "github.com/charmbracelet/bubbles/help"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/contextstrip"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/deleteconfirm"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/detail"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/list"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/wizard"
    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type view int

const (
    viewList view = iota
    viewDetail
    viewCreate
    viewEdit
    viewDeleteConfirm
)

type Model struct {
    factionState *state.FactionState
    paths        *campaigns.Paths
    rulebook     *rulebook.Rulebook
    spatialMap   *spatial.RegionMap

    view          view
    list          list.Model
    detail        detail.Model
    create        wizard.Model
    edit          wizard.Model
    deleteConfirm deleteconfirm.Model
    strip         contextstrip.Model

    saveErr error // surfaced inline in the active view's render
}

func New(
    factionState *state.FactionState,
    paths *campaigns.Paths,
    rulebook *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
) Model

func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
func (m Model) Help() help.KeyMap // passthrough to active sub-model
```

Back-edge map (static, used by `manage.Update` when it handles `CancelMsg` / completion):

```go
var backTarget = map[view]view{
    viewDetail:        viewList,
    viewCreate:        viewList,
    viewEdit:          viewDetail,
    viewDeleteConfirm: viewDetail,
}
```

`viewList` is the top of the stack — no back-edge. `Tab` (handled at root) leaves the mode.

### Helper Interface and Help Dispatch

New file `internal/faction/tui/helper.go`:

```go
package tui

import "github.com/charmbracelet/bubbles/help"

// Helper is implemented by sub-models that contribute keybindings to the
// help overlay. The root combines the active sub's Help() with global
// bindings (Tab, Shift-Tab, q, ?) for the ? overlay's render.
type Helper interface {
    Help() help.KeyMap
}
```

Each sub-view's `Help()` returns a `help.KeyMap` (a struct that satisfies `help.KeyMap`'s `ShortHelp() []key.Binding` / `FullHelp() [][]key.Binding`). Manage's `Help()` returns its active sub-model's `Help()`:

```go
func (m Model) Help() help.KeyMap {
    switch m.view {
    case viewList:    return m.list.Help()
    case viewDetail:  return m.detail.Help()
    case viewCreate:  return m.create.Help()
    case viewEdit:    return m.edit.Help()
    case viewDeleteConfirm: return m.deleteConfirm.Help()
    }
    return emptyKeyMap{}
}
```

Root's `?` handler (in `update.go`) checks whether the active sub satisfies `Helper`; if so, calls `Help()` and renders via `help.New().View(km)` composed with the globals' keymap. Foundation's `showHelpStub` toggle stays; the fixed string is replaced by the composed render.

### Faction Creation Wizard — Step Inventory

The `wizard.Model` holds a draft `domain.Faction` and walks `huh.NewForm(...)` groups. Step order matches the SWN checklist:

| # | Step | huh widget | Source data | Notes |
|---|------|-----------|-------------|-------|
| 1 | Scale | `huh.NewSelect[domain.FactionScale]` | constants in `domain` | On change, seeds Force/Cunning/Wealth defaults via `RatingsFromScale` (assignment step refines which is which). |
| 2 | Attribute assignment | three `huh.NewSelect[domain.FactionStat]` | derived | User picks which stat is primary, secondary, tertiary. Validator: all three distinct. |
| 3 | HP display | `huh.NewNote` | derived | Shows computed MaxHP (`CalcMaxHP`); read-only confirmation. |
| 4 | Tags | `huh.NewMultiSelect[*domain.Tag]` | `rulebook.Tags` | Cap 2 via validator. |
| 5 | Goal | `huh.NewSelect[*domain.Goal]` | `rulebook.Goals` | Sets `ActiveGoal{GoalID: goal.ID}` with zero progress. |
| 6 | Homeworld | `huh.NewSelect[string]` (world IDs) | spatial worlds | Auto-creates Base of Influence at max HP on the picked world. |
| 7 | Starting assets | `huh.NewMultiSelect[*domain.AssetDefinition]` (×2 groups) | `rulebook.AssetDefinitions` filtered by rating + homeworld tech_level | Quota by scale (see Open Question 2). Selected definitions instantiate as `*domain.Asset` with generated IDs. |
| 8 | Starting Coin | `huh.NewInput` (int validator) | — | Default 0. |

Wizard constructor signature:

```go
func New(
    rulebook *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
    initial *domain.Faction, // nil for create; pre-populated for edit
) Model
```

Internal `confirmingDiscard bool` overlays the y/n prompt on Esc; on y, emits `CancelMsg`; on n, returns to the form. The `huh.Form` is driven via standard `tea.Model` plumbing (huh forms satisfy `tea.Model`).

Completion: when `wizard.Model.form.State == huh.StateCompleted`, emit `CreatedMsg{Faction: &draft}` or `EditSavedMsg{Faction: &draft}` based on the constructor's `initial == nil` test (stored as a `boolean isCreate` field).

### Update Discipline (Inherited)

Same comment block Foundation installed at the top of `internal/faction/tui/update.go` applies. No changes to update discipline rules in this initiative; the MAY/MAY NOT list is sufficient. Each sub-view enforces it locally — sub-views never call state CRUD directly; they emit completion messages and let `manage.Update` route.

### Out-of-Initiative Verification

Per `docs/process/session-modes.md` § 9 "Re-grounding before action," every execution session opens with a quick re-read of the relevant current source. For this initiative, the most-likely drift surfaces are:

- `domain.Faction` field set (Tags/Bases shape) if a domain-shaping commit lands during execution.
- `state.FactionState.Factions` map type if state is refactored.
- `spatial.RegionMap` API if F-012 (Spatial Map CLI) lands mid-arc — though it's `--` triggered in backlog, this is unlikely.
- `bubbles/list`, `bubbles/help` API surface if minor-version bumps land — Charm releases are frequent.

## Out of Scope

Inherited from Discovery (`tui-manage-discovery.md` § Out of Scope), restated for the executor's convenience:

- **Anything Turn-related.** Initiative 3.
- **Tabbed faction-detail** (Stats / Assets / Goals / History split). Sub-view growth deferred. Detail is one scrollable view in this initiative.
- **`Adapter.Stop()` cancellation design.** Routed to Initiative 3.
- **Save-error rollback.** Discovery flagged as MVP-resolved; this plan ratifies (Decision 7).
- **Filter / sort / search on the list.** No per-field filter, no fuzzy search.
- **Multi-select / bulk operations.** One faction at a time.
- **Concurrency-with-Turn.** Turn's problem.
- **Validation beyond invariants.** `huh` form validators handle UX-level constraints; state CRUD enforces only uniqueness / existence.

Out of Scope from Plan (not in Discovery's list):

- **Spatial seed command** (F-012). Wizard expects spatial data to exist; missing-data error surfaces with a `hint:` line per the rulebook pattern. Seed command itself is a separate initiative.
- **`bubbles/help` per-mode color theming.** Default Charm theme until a theming initiative ships.
- **Wizard step-skipping or step-back navigation.** `huh.Form` supports `Group` navigation; the wizard uses default Group order without custom skip logic. Esc opens discard-confirm; there is no "back one step" key.
- **Detail view scroll polish.** A long faction's detail may overflow; basic scrolling via `bubbles/viewport` is OK if needed, but pixel-perfect layout is out of scope.

---

## Work Breakdown

### Phase 1 — State Package CRUD

Adds the write-through CRUD surface the TUI consumes. Standalone, no TUI dependency. After this phase the state package exposes Create/Update/Delete/Get; nothing in the TUI uses them yet.

#### Commit 1 — `feat(state): faction CRUD with write-through persistence`

##### Task 1 — `internal/faction/state/crud.go` (new file)

Implement the four functions and three sentinels per Shared Context § "State CRUD Function Signatures (Final Shape)". Concrete implementation sketch:

```go
package state

import (
    "errors"
    "fmt"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

var (
    ErrFactionAlreadyExists = errors.New("state: faction already exists")
    ErrFactionNotFound      = errors.New("state: faction not found")
    ErrInvalidFactionID     = errors.New("state: invalid faction id (empty)")
)

func CreateFaction(path string, fs *FactionState, faction *domain.Faction) error {
    if faction.ID == "" {
        return ErrInvalidFactionID
    }
    if _, exists := fs.Factions[faction.ID]; exists {
        return fmt.Errorf("state: create %q: %w", faction.ID, ErrFactionAlreadyExists)
    }
    fs.Factions[faction.ID] = faction
    if err := Save(path, fs); err != nil {
        return fmt.Errorf("state: create %q: save: %w", faction.ID, err)
    }
    return nil
}

func UpdateFaction(path string, fs *FactionState, faction *domain.Faction) error {
    if faction.ID == "" {
        return ErrInvalidFactionID
    }
    if _, exists := fs.Factions[faction.ID]; !exists {
        return fmt.Errorf("state: update %q: %w", faction.ID, ErrFactionNotFound)
    }
    fs.Factions[faction.ID] = faction
    if err := Save(path, fs); err != nil {
        return fmt.Errorf("state: update %q: save: %w", faction.ID, err)
    }
    return nil
}

func DeleteFaction(path string, fs *FactionState, id string) error {
    if id == "" {
        return ErrInvalidFactionID
    }
    if _, exists := fs.Factions[id]; !exists {
        return fmt.Errorf("state: delete %q: %w", id, ErrFactionNotFound)
    }
    delete(fs.Factions, id)
    if err := Save(path, fs); err != nil {
        return fmt.Errorf("state: delete %q: save: %w", id, err)
    }
    return nil
}

func GetFaction(fs *FactionState, id string) (*domain.Faction, error) {
    if id == "" {
        return nil, ErrInvalidFactionID
    }
    faction, ok := fs.Factions[id]
    if !ok {
        return nil, fmt.Errorf("state: get %q: %w", id, ErrFactionNotFound)
    }
    return faction, nil
}
```

##### Task 2 — `internal/faction/state/crud_test.go` (new file)

Table-driven tests covering: create + read-back; create duplicate → `ErrFactionAlreadyExists`; update missing → `ErrFactionNotFound`; delete missing → `ErrFactionNotFound`; create empty ID → `ErrInvalidFactionID`; round-trip create → Save → Load → Get returns equivalent faction. Use `t.TempDir()` for the path. Mirrors the testing style in `faction_state.go`'s existing tests (verify shape at execution time).

##### Commit message

```
feat(state): faction CRUD with write-through persistence

- add CreateFaction / UpdateFaction / DeleteFaction / GetFaction in
  internal/faction/state/crud.go
- declare ErrFactionAlreadyExists / ErrFactionNotFound / ErrInvalidFactionID
  sentinels; check with errors.Is
- each mutation validates invariants, applies in-memory, then persists via
  Save in one operation; no rollback on Save failure
```

---

### Phase 2 — Manage Router and Read-Only Views

Stands up the Manage router, the list view, the detail view, and the delete-confirm sub-view. After this phase a GM can browse and delete factions but cannot create or edit. Mutating sub-views (create wizard, edit) follow in Phase 3.

#### Commit 2 — `feat(tui/manage): router shell, list view, empty-state hint`

Replaces Foundation's single-banner `manage.go` shell with the router + list view. The other sub-view packages are scaffolded as placeholder Models (so the router compiles and the switch dispatches), with their full implementations following in Commits 3–5. The router emits no completion handling yet — that wiring lands incrementally as each sub-view ships.

File-creation order (each step compiles before the next):

##### Task 1 — `cmd/gm-toolkit/faction.go`

Expand `factionRun` to load spatial alongside rulebook and thread both through `tui.Run`. Spatial loading is scheduled here (rather than Commit 4) because Commit 4 grows large; the infra add belongs with the router commit that first exposes `*spatial.RegionMap` to the root Model.

```go
// after rb load, before state.Load:
spatialMap, err := spatial.LoadRegionMap(paths.SpatialDataDir)
if err != nil {
    return fmt.Errorf("faction: load spatial from %s: %w\n  hint: (spatial seed command TBD — see F-012)", paths.SpatialDataDir, err)
}
// ...
return tui.Run(eng, factionState, &paths, rb, spatialMap, log)
```

Add the `spatial` import. Verify `spatial.LoadRegionMap` signature at re-grounding (`internal/spatial/region_map.go:85`).

##### Task 2 — `internal/faction/tui/tui.go`

Expand `Run` signature to accept `*rulebook.Rulebook` and `*spatial.RegionMap`. Pass both to the root `Model` constructor.

```go
func Run(
    eng *engine.Engine,
    factionState *state.FactionState,
    paths *campaigns.Paths,
    rb *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
    log *slog.Logger,
) error {
    adp := adapter.New(eng, factionState, log)
    root := NewModel(adp, factionState, paths, rb, spatialMap)
    prog := tea.NewProgram(root, tea.WithAltScreen())
    _, err := prog.Run()
    return err
}
```

##### Task 3 — `internal/faction/tui/model.go`

Expand `NewModel` to accept rulebook + spatial. Construct `manage.New(factionState, paths, rulebook, spatialMap)` instead of `manage.New()`. Hold rulebook + spatial in the root struct only as long as needed to construct subs (then discard if nothing else uses them — or keep as fields if Turn will reuse).

The root `subs` map keying remains `modebar.Mode` → `tea.Model`. Manage's new shape satisfies `tea.Model`.

```go
type Model struct {
    bar          modebar.Model
    subs         map[modebar.Mode]tea.Model
    adapter      *adapter.Adapter
    confirmExit  bool
    priorMode    modebar.Mode
    showHelpStub bool
}

func NewModel(
    adp *adapter.Adapter,
    factionState *state.FactionState,
    paths *campaigns.Paths,
    rb *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
) Model {
    return Model{
        bar:     modebar.New(),
        adapter: adp,
        subs: map[modebar.Mode]tea.Model{
            modebar.ModeManage: manage.New(factionState, paths, rb, spatialMap),
            modebar.ModeTurn:   turn.New(),
        },
    }
}
```

##### Task 4 — `internal/faction/tui/views/manage/messages.go` (new file)

Per Shared Context § "Completion Message Types (Final Shape)". Pure declarations; no logic.

##### Task 5 — `internal/faction/tui/views/manage/manage.go` (rewrite)

Replace the Foundation stub with the router. Initial shape:

```go
package manage

import (
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/campaigns"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/contextstrip"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/deleteconfirm"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/detail"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/list"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage/wizard"
    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type view int

const (
    viewList view = iota
    viewDetail
    viewCreate
    viewEdit
    viewDeleteConfirm
)

var backTarget = map[view]view{
    viewDetail:        viewList,
    viewCreate:        viewList,
    viewEdit:          viewDetail,
    viewDeleteConfirm: viewDetail,
}

type Model struct {
    factionState *state.FactionState
    paths        *campaigns.Paths
    rulebook     *rulebook.Rulebook
    spatialMap   *spatial.RegionMap

    view          view
    list          list.Model
    detail        detail.Model
    create        wizard.Model
    edit          wizard.Model
    deleteConfirm deleteconfirm.Model
    strip         contextstrip.Model

    saveErr error
}

func New(
    factionState *state.FactionState,
    paths *campaigns.Paths,
    rb *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
) Model {
    return Model{
        factionState: factionState,
        paths:        paths,
        rulebook:     rb,
        spatialMap:   spatialMap,
        view:         viewList,
        list:         list.New(factionState),
        detail:       detail.Model{},        // populated on transition
        create:       wizard.Model{},        // populated on transition
        edit:         wizard.Model{},        // populated on transition
        deleteConfirm: deleteconfirm.Model{}, // populated on transition
        strip:        contextstrip.New(factionState),
    }
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Completion-message dispatch (centralized persist + transition).
    switch msg := msg.(type) {
    case CreatedMsg:
        if err := state.CreateFaction(m.paths.StatePath, m.factionState, msg.Faction); err != nil {
            m.saveErr = err
            return m, nil
        }
        m.strip = contextstrip.New(m.factionState)
        m.view = backTarget[viewCreate]
        return m, nil

    case EditSavedMsg:
        if err := state.UpdateFaction(m.paths.StatePath, m.factionState, msg.Faction); err != nil {
            m.saveErr = err
            return m, nil
        }
        m.strip = contextstrip.New(m.factionState)
        m.view = backTarget[viewEdit]
        return m, nil

    case DeletedMsg:
        if err := state.DeleteFaction(m.paths.StatePath, m.factionState, msg.ID); err != nil {
            m.saveErr = err
            return m, nil
        }
        m.strip = contextstrip.New(m.factionState)
        m.view = viewList // delete returns to list, not detail (detail is gone)
        return m, nil

    case CancelMsg:
        m.view = backTarget[m.view]
        return m, nil
    }

    // Forward to active sub-model; also handle forward-transition keys at this level.
    return m.routeForward(msg)
}

func (m Model) routeForward(msg tea.Msg) (Model, tea.Cmd) {
    // Forward keys that drive transitions (Enter, n, e, d, Esc) are handled by
    // each sub-view; the sub-view emits the appropriate transition message
    // (e.g., list emits a viewDetailRequestedMsg{ID} on Enter) and manage
    // constructs the next sub-model.
    //
    // Implementation note: rather than per-key intercepting at manage, sub-views
    // return transition messages of their own; manage handles them here in the
    // outer switch. The transition messages are also declared in messages.go.
    // ... (see Task 6 for the transition-message types and forwarding)
    return m, nil
}
```

Forward-transition messages (declared in `messages.go` alongside completion messages):

```go
type RequestDetailMsg struct{ FactionID string } // list → detail
type RequestCreateMsg struct{}                    // list → create
type RequestEditMsg struct{ Faction *domain.Faction }    // detail → edit
type RequestDeleteMsg struct{ Faction *domain.Faction }  // detail → deleteConfirm
```

Each handled in `manage.Update`:

```go
case RequestDetailMsg:
    faction, _ := state.GetFaction(m.factionState, msg.FactionID)
    m.detail = detail.New(faction)
    m.view = viewDetail
    return m, nil
case RequestCreateMsg:
    m.create = wizard.New(m.rulebook, m.spatialMap, nil)
    m.view = viewCreate
    return m, m.create.Init()
case RequestEditMsg:
    m.edit = wizard.New(m.rulebook, m.spatialMap, msg.Faction)
    m.view = viewEdit
    return m, m.edit.Init()
case RequestDeleteMsg:
    m.deleteConfirm = deleteconfirm.New(msg.Faction)
    m.view = viewDeleteConfirm
    return m, nil
```

Default-forward to active sub:

```go
var cmd tea.Cmd
switch m.view {
case viewList:
    m.list, cmd = m.list.Update(msg)
case viewDetail:
    m.detail, cmd = m.detail.Update(msg)
// ... (one case per view)
}
return m, cmd
```

`View()`:

```go
func (m Model) View() string {
    var content string
    switch m.view {
    case viewList:
        content = lipgloss.JoinHorizontal(
            lipgloss.Top,
            m.list.View(),
            m.strip.View(),
        )
    case viewDetail:
        content = m.detail.View()
    case viewCreate:
        content = m.create.View()
    case viewEdit:
        content = m.edit.View()
    case viewDeleteConfirm:
        content = m.deleteConfirm.View()
    }
    if m.saveErr != nil {
        content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("save error: " + m.saveErr.Error())
    }
    return content
}

func (m Model) Help() help.KeyMap {
    // populated in Commit 7
    return nil
}
```

The `Help()` method exists with a `nil` return until Commit 7 wires the per-view bindings. Root's `?` handler tolerates `nil` (renders the existing stub).

##### Task 6 — Sub-view package stubs

Create the following files so `manage.go` compiles. Each is a minimal `tea.Model` with an empty Update and a single Placeholder view; the real implementations land in Commits 3–5.

- `internal/faction/tui/views/manage/list/list.go` — IMPLEMENTED (full list view, this commit; see Task 7).
- `internal/faction/tui/views/manage/detail/detail.go` — stub Model + New + Init/Update/View returning `tui.Placeholder.Render("Detail — Commit 3")`.
- `internal/faction/tui/views/manage/deleteconfirm/deleteconfirm.go` — stub Model + `New(*domain.Faction)` + stubs returning placeholder.
- `internal/faction/tui/views/manage/wizard/wizard.go` — stub Model + `New(*rulebook.Rulebook, *spatial.RegionMap, *domain.Faction)` + stubs returning placeholder.
- `internal/faction/tui/views/manage/contextstrip/contextstrip.go` — stub Model + `New(*state.FactionState)` + stubs returning placeholder (`tui.Placeholder.Render("Strip — Commit 6")`).

Each stub's `Update` returns `(m, nil)` and ignores the input. This keeps the router code path live without committing to the real shape until the dedicated commit.

##### Task 7 — `internal/faction/tui/views/manage/list/list.go` (full implementation)

The list view, this commit. Built on `bubbles/list`. Items are factions keyed by ID, rendered as `Name · Scale · HP/Max · Coin`. Empty-state hint when the underlying `factionState.Factions` map is empty.

```go
package list

import (
    "fmt"
    "sort"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
)

type item struct {
    id    string
    name  string
    scale string
    hp    string
    coin  string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return fmt.Sprintf("%s · HP %s · Coin %s", i.scale, i.hp, i.coin) }
func (i item) FilterValue() string { return i.name }

type Model struct {
    list   list.Model
    keys   keyMap
    state  *state.FactionState // reference for re-checking empty-state
}

type keyMap struct {
    New    key.Binding
    Select key.Binding
}

func New(fs *state.FactionState) Model {
    items := buildItems(fs)
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Factions"
    l.SetShowHelp(false)       // root help overlay handles bindings
    l.SetShowStatusBar(false)
    return Model{
        list:  l,
        state: fs,
        keys: keyMap{
            New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
            Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
        },
    }
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.New):
            return m, func() tea.Msg { return manage.RequestCreateMsg{} }
        case key.Matches(msg, m.keys.Select):
            if it, ok := m.list.SelectedItem().(item); ok {
                return m, func() tea.Msg { return manage.RequestDetailMsg{FactionID: it.id} }
            }
        }
    case tea.WindowSizeMsg:
        // list sizing — Plan defers to execution-time tuning; reserve right ~30 cols for context strip.
        m.list.SetSize(msg.Width-32, msg.Height-4)
    }
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if len(m.state.Factions) == 0 {
        return emptyState()
    }
    return m.list.View()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }
func (h helpKeys) ShortHelp() []key.Binding { return []key.Binding{h.keys.New, h.keys.Select} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.New, h.keys.Select}} }

// buildItems sorts factions by ID for stable ordering and projects to list.Item.
func buildItems(fs *state.FactionState) []list.Item {
    ids := make([]string, 0, len(fs.Factions))
    for id := range fs.Factions {
        ids = append(ids, id)
    }
    sort.Strings(ids)

    items := make([]list.Item, 0, len(ids))
    for _, id := range ids {
        f := fs.Factions[id]
        items = append(items, item{
            id:    id,
            name:  f.Name,
            scale: string(f.Scale),
            hp:    fmt.Sprintf("%d/%d", f.CurrentHP, f.MaxHP),
            coin:  fmt.Sprintf("%d", f.Coin),
        })
    }
    return items
}

func emptyState() string {
    return lipgloss.JoinVertical(
        lipgloss.Center,
        lipgloss.NewStyle().Render("No factions yet."),
        lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Press n to create one"),
    )
}
```

Note the import cycle risk: `views/manage/list` imports `views/manage` for the transition message types. Since `manage` only forward-declares the messages (no imports of `list`), there is no cycle. Verify with `go build` at end of commit.

If a cycle does appear, lift the transition messages to a sibling subpackage `views/manage/messages` and import from both. Pick at execution time.

##### Task 8 — `internal/faction/tui/styles/styles.go` (small additions)

Add styles needed by Manage. Empty-state styles are inlined in list (see emptyState above); add only what is shared.

- `ListFrame lipgloss.Style` — optional border for the list area (defer to execution-time UX call).
- `StripFrame lipgloss.Style` — optional border for the context strip (Commit 6 uses).
- `SaveError lipgloss.Style` — used by manage.View when saveErr is set.

Add only `SaveError` in this commit; the frame styles land with their consumers.

```go
SaveError = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
```

##### Commit message

```
feat(tui/manage): router shell, list view, empty-state hint

- replace Foundation's manage placeholder with the router shell:
  view enum, transition + completion message dispatch, sub-model
  fields, back-edge map
- list view (bubbles/list) with item shape Name·Scale·HP·Coin and
  centered empty-state hint
- scaffold stub Models for detail / deleteconfirm / wizard / contextstrip
  packages (filled in Commits 3-6)
- load spatial alongside rulebook in factionRun; expand tui.Run
  signature to accept *rulebook.Rulebook and *spatial.RegionMap
```

---

#### Commit 3 — `feat(tui/manage): detail view and delete-confirm`

Fills in the two read-only sub-views. Detail shows the full faction; `d` opens delete-confirm; delete-confirm y/n emits `DeletedMsg` or `CancelMsg`.

##### Task 1 — `internal/faction/tui/views/manage/detail/detail.go`

Replaces the stub. Renders a faction's full state as a single scrollable block. `e` emits `RequestEditMsg`; `d` emits `RequestDeleteMsg`; Esc emits `CancelMsg` (which `manage.Update` routes back to list).

```go
package detail

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
)

type Model struct {
    faction *domain.Faction
    keys    keyMap
}

type keyMap struct {
    Edit   key.Binding
    Delete key.Binding
    Back   key.Binding
}

func New(faction *domain.Faction) Model {
    return Model{
        faction: faction,
        keys: keyMap{
            Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
            Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
            Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
        },
    }
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if km, ok := msg.(tea.KeyMsg); ok {
        switch {
        case key.Matches(km, m.keys.Edit):
            return m, func() tea.Msg { return manage.RequestEditMsg{Faction: m.faction} }
        case key.Matches(km, m.keys.Delete):
            return m, func() tea.Msg { return manage.RequestDeleteMsg{Faction: m.faction} }
        case key.Matches(km, m.keys.Back):
            return m, func() tea.Msg { return manage.CancelMsg{} }
        }
    }
    return m, nil
}

func (m Model) View() string {
    if m.faction == nil {
        return "(no faction)"
    }
    var b strings.Builder
    f := m.faction
    fmt.Fprintf(&b, "%s  (%s)\n", f.Name, f.ID)
    fmt.Fprintf(&b, "Scale: %s\n", f.Scale)
    fmt.Fprintf(&b, "Force: %d  Cunning: %d  Wealth: %d\n", f.Force, f.Cunning, f.Wealth)
    fmt.Fprintf(&b, "HP: %d/%d   Coin: %d   XP: %d\n", f.CurrentHP, f.MaxHP, f.Coin, f.XP)
    fmt.Fprintf(&b, "Homeworld: %s\n", f.Homeworld.ID())
    fmt.Fprintf(&b, "\nTags:\n")
    for _, tag := range f.Tags {
        fmt.Fprintf(&b, "  - %s\n", tag.Name)
    }
    if f.ActiveGoal != nil {
        fmt.Fprintf(&b, "\nActive Goal: %s (progress %d)\n", f.ActiveGoal.GoalID, f.ActiveGoal.Progress)
    }
    fmt.Fprintf(&b, "\nAssets:\n")
    for _, asset := range f.Assets {
        fmt.Fprintf(&b, "  - %s\n", asset.ID)
    }
    fmt.Fprintf(&b, "\nBases:\n")
    for _, base := range f.Bases {
        fmt.Fprintf(&b, "  - %s\n", base.WorldID)
    }
    return b.String()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }
func (h helpKeys) ShortHelp() []key.Binding {
    return []key.Binding{h.keys.Edit, h.keys.Delete, h.keys.Back}
}
func (h helpKeys) FullHelp() [][]key.Binding {
    return [][]key.Binding{{h.keys.Edit, h.keys.Delete, h.keys.Back}}
}
```

Verify `domain.Faction`'s field set at re-grounding — the Survey reported `Tags []*Tag`, `ActiveGoal *ActiveGoal`, `Assets map[string]*Asset`, `Bases []*Base`, `Homeworld Location` (interface). The render shape above assumes those exist; adjust on drift. `Location.ID()` interface method per `internal/spatial/spatial.go:10`.

##### Task 2 — `internal/faction/tui/views/manage/deleteconfirm/deleteconfirm.go`

Replaces the stub. Single y/n prompt.

```go
package deleteconfirm

import (
    "fmt"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
)

type Model struct {
    faction *domain.Faction
    keys    keyMap
}

type keyMap struct {
    Confirm key.Binding
    Cancel  key.Binding
}

func New(faction *domain.Faction) Model {
    return Model{
        faction: faction,
        keys: keyMap{
            Confirm: key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
            Cancel:  key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "no")),
        },
    }
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if km, ok := msg.(tea.KeyMsg); ok {
        switch {
        case key.Matches(km, m.keys.Confirm):
            return m, func() tea.Msg { return manage.DeletedMsg{ID: m.faction.ID} }
        case key.Matches(km, m.keys.Cancel):
            return m, func() tea.Msg { return manage.CancelMsg{} }
        }
    }
    return m, nil
}

func (m Model) View() string {
    prompt := fmt.Sprintf("Delete faction %q (%s)? [y/N]", m.faction.Name, m.faction.ID)
    return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Padding(1, 2).Render(prompt)
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }
func (h helpKeys) ShortHelp() []key.Binding  { return []key.Binding{h.keys.Confirm, h.keys.Cancel} }
func (h helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{h.keys.Confirm, h.keys.Cancel}} }
```

##### Commit message

```
feat(tui/manage): detail view and delete-confirm

- detail view renders Name/ID/scale/stats/HP/Coin/XP/homeworld/tags/
  goal/assets/bases; e/d/Esc keys emit RequestEditMsg /
  RequestDeleteMsg / CancelMsg
- deleteconfirm: y/n prompt; y emits DeletedMsg, n/Esc emits CancelMsg
- both sub-views expose Help() with their bindings (consumed in Commit 7)
```

---

### Phase 3 — Mutating Sub-Views (Create Wizard, Edit)

After this phase the GM can author factions end-to-end. The wizard walks the SWN creation checklist; edit reuses the wizard with pre-populated values.

#### Commit 4 — `feat(tui/manage): faction creation wizard`

**Flagged as the heaviest commit in the plan.** Implements the multi-step `huh.Form` covering all SWN creation steps. If the execution session surfaces complexity that warrants splitting (e.g., asset-step quota logic balloons), pause and re-plan: extract the asset step into a Commit 4b. Default is single commit.

##### Task 1 — `internal/faction/tui/views/manage/wizard/wizard.go`

Replace the stub with the full wizard. Structure:

```go
package wizard

import (
    "fmt"
    "strconv"

    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
    isCreate   bool                // true: emit CreatedMsg; false: emit EditSavedMsg
    draft      *domain.Faction
    form       *huh.Form
    rulebook   *rulebook.Rulebook
    spatialMap *spatial.RegionMap

    confirmingDiscard bool
    discardKeys       discardKeyMap
}

type discardKeyMap struct {
    Confirm key.Binding
    Cancel  key.Binding
}

// New builds the wizard. initial == nil means "create" mode; non-nil means
// "edit" mode with the form pre-populated from a deep copy of initial.
func New(rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, initial *domain.Faction) Model {
    isCreate := initial == nil
    var draft *domain.Faction
    if isCreate {
        draft = newDraft()
    } else {
        draft = deepCopy(initial)
    }
    return Model{
        isCreate:   isCreate,
        draft:      draft,
        form:       buildForm(draft, rb, spatialMap),
        rulebook:   rb,
        spatialMap: spatialMap,
        discardKeys: discardKeyMap{
            Confirm: key.NewBinding(key.WithKeys("y")),
            Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
        },
    }
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if m.confirmingDiscard {
        if km, ok := msg.(tea.KeyMsg); ok {
            switch {
            case key.Matches(km, m.discardKeys.Confirm):
                return m, func() tea.Msg { return manage.CancelMsg{} }
            case key.Matches(km, m.discardKeys.Cancel):
                m.confirmingDiscard = false
                return m, nil
            }
        }
        return m, nil
    }

    if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
        m.confirmingDiscard = true
        return m, nil
    }

    formAny, cmd := m.form.Update(msg)
    m.form = formAny.(*huh.Form)
    if m.form.State == huh.StateCompleted {
        if m.isCreate {
            return m, func() tea.Msg { return manage.CreatedMsg{Faction: m.draft} }
        }
        return m, func() tea.Msg { return manage.EditSavedMsg{Faction: m.draft} }
    }
    return m, cmd
}

func (m Model) View() string {
    if m.confirmingDiscard {
        return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Padding(1, 2).
            Render("Discard changes? [y/N]")
    }
    return m.form.View()
}

func (m Model) Help() help.KeyMap { return helpKeys{} }

type helpKeys struct{}
func (helpKeys) ShortHelp() []key.Binding {
    return []key.Binding{
        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "advance")),
        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "discard")),
    }
}
func (helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{helpKeys{}.ShortHelp()} }
```

`buildForm` constructs the multi-group `huh.Form`. Each `huh.NewGroup(...)` covers one or more SWN-checklist steps. Pseudocode per Shared Context § "Faction Creation Wizard — Step Inventory":

```go
func buildForm(draft *domain.Faction, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
    // Step 1 — Scale
    scaleGroup := huh.NewGroup(
        huh.NewSelect[domain.FactionScale]().
            Title("Scale").
            Options(
                huh.NewOption("Minor", domain.ScaleMinor),
                huh.NewOption("Major", domain.ScaleMajor),
                huh.NewOption("Hegemon", domain.ScaleHegemon),
            ).
            Value(&draft.Scale),
    )

    // Step 2 — Attribute assignment (which stat is primary/secondary/tertiary)
    // Implementation detail: collect three FactionStat selections, then post-process
    // to assign Force/Cunning/Wealth based on draft.Scale's primary/secondary/tertiary
    // values (RatingsFromScale). Validator ensures all three are distinct.
    var primary, secondary, tertiary domain.FactionStat
    statsGroup := huh.NewGroup(
        huh.NewSelect[domain.FactionStat]().Title("Primary attribute").
            Options(statOptions...).Value(&primary),
        huh.NewSelect[domain.FactionStat]().Title("Secondary attribute").
            Options(statOptions...).Value(&secondary),
        huh.NewSelect[domain.FactionStat]().Title("Tertiary attribute").
            Options(statOptions...).Value(&tertiary),
    ).WithHide(func() bool { return draft.Scale == "" })

    // Step 3 — HP display (note)
    hpGroup := huh.NewGroup(
        huh.NewNote().Title("Max HP").Description(func() string {
            // post-process the previous step to write Force/Cunning/Wealth into draft,
            // then call domain.CalcMaxHP(draft) and render.
            applyAttributes(draft, primary, secondary, tertiary)
            return fmt.Sprintf("%d HP (auto-derived from attributes)", domain.CalcMaxHP(draft))
        }()),
    )

    // Step 4 — Tags
    tagOptions := tagsToOptions(rb.Tags)
    var selectedTags []*domain.Tag
    tagsGroup := huh.NewGroup(
        huh.NewMultiSelect[*domain.Tag]().
            Title("Tags (max 2)").
            Options(tagOptions...).
            Value(&selectedTags).
            Validate(func(t []*domain.Tag) error {
                if len(t) > 2 { return fmt.Errorf("at most 2 tags") }
                return nil
            }),
    )

    // Step 5 — Starting Goal
    goalOptions := goalsToOptions(rb.Goals)
    var selectedGoal *domain.Goal
    goalGroup := huh.NewGroup(
        huh.NewSelect[*domain.Goal]().
            Title("Starting goal").
            Options(goalOptions...).
            Value(&selectedGoal),
    )

    // Step 6 — Homeworld
    worldOptions := worldsToOptions(spatialMap)
    var homeworldID string
    homeworldGroup := huh.NewGroup(
        huh.NewSelect[string]().
            Title("Homeworld").
            Options(worldOptions...).
            Value(&homeworldID),
    )

    // Step 7 — Starting assets (filtered by rating + homeworld tech)
    // See Open Question 2 for two-group vs single-group shape.
    assetGroup := buildAssetGroup(draft, rb, spatialMap, &homeworldID)

    // Step 8 — Starting Coin
    var coinStr string
    coinGroup := huh.NewGroup(
        huh.NewInput().
            Title("Starting Coin").
            Value(&coinStr).
            Validate(func(s string) error {
                if _, err := strconv.Atoi(s); err != nil { return fmt.Errorf("must be an integer") }
                return nil
            }),
    )

    form := huh.NewForm(
        scaleGroup, statsGroup, hpGroup, tagsGroup, goalGroup, homeworldGroup, assetGroup, coinGroup,
    )

    // form.OnComplete (or post-completion in Update) writes selectedTags, selectedGoal,
    // homeworldID, coinStr back into draft and constructs the Base of Influence at
    // max HP on the homeworld.
    return form
}
```

Key implementation details to nail at execution time:

1. **The post-processing path** — `huh.Form` collects values into bound vars; the wizard's `Update` (after `form.State == huh.StateCompleted`) projects them into `draft` fields. The current sketch above intermixes projection inside `buildForm`'s closures (cleaner for live HP display) but may need to move to the completion path if `huh` doesn't support intra-form recomputation cleanly. Re-ground at execution time against `huh`'s API.
2. **Asset step shape** — Open Question 2. Two `huh.MultiSelect` groups (one filtered to primary attribute, one to all) is the cleanest UX; collapse to one group with a single quota validator if `huh` makes the two-group flow awkward.
3. **Auto-Base on homeworld** — On completion, append a `*domain.Base` to `draft.Bases` for the homeworld at `draft.MaxHP` HP. Verify `domain.Base` struct shape at re-grounding.
4. **Faction ID generation** — Open Question 3. Recommended: derive `slug(draft.Name)` after Step 8; check uniqueness against `fs.Factions`; if collision, prompt for an explicit ID. Implementation may add a hidden Step 9 (ID input) shown only on collision.
5. **Edit mode pre-population** — The `deepCopy(initial)` populates `draft`; the `huh.Form` constructors use `Value(&draft.Scale)` etc., so the form opens with current values selected. The Scale-derived attribute defaults must be detected ("don't reseed if draft.Scale is already set") to avoid clobbering edit-mode values.

##### Task 2 — `domain.Faction` constructor helper (if needed)

`newDraft()` returns an empty `domain.Faction` with sane zero values:

```go
// in wizard.go or domain/faction.go — picker's call at execution time.
func newDraft() *domain.Faction {
    return &domain.Faction{
        Tags:   []*domain.Tag{},
        Assets: map[string]*domain.Asset{},
        Bases:  []*domain.Base{},
    }
}
```

If `domain` gains a `NewFaction()` constructor as part of this commit, place it there and import; otherwise inline in wizard.

`deepCopy(faction)` — straightforward: marshal/unmarshal via TOML round-trip, or hand-implement field-by-field. TOML round-trip is shortest; hand-implementation is faster. Execution-time call.

##### Task 3 — Helper functions in `wizard.go`

- `tagsToOptions([]*domain.Tag) []huh.Option[*domain.Tag]`
- `goalsToOptions([]*domain.Goal) []huh.Option[*domain.Goal]`
- `worldsToOptions(*spatial.RegionMap) []huh.Option[string]` — iterates spatial map, returns `{Title: world.Name() + " (TL " + world.TechLevel() + ")", Value: world.ID()}` options.
- `buildAssetGroup(draft, rb, spatialMap, *homeworldID) *huh.Group` — filters `rb.AssetDefinitions` by attribute rating and homeworld tech_level; constructs the multi-select(s) per Open Question 2.
- `applyAttributes(draft, primary, secondary, tertiary)` — looks up scale's primary/secondary/tertiary values via `RatingsFromScale`, assigns to the named `FactionStat`s on draft. Sets `draft.MaxHP = CalcMaxHP(draft)` and `draft.CurrentHP = draft.MaxHP`.

Verify all helper data sources at re-grounding — `rulebook.Rulebook` field names (`Tags`, `Goals`, `AssetDefinitions`), `spatial.RegionMap`'s world iteration surface.

##### Commit message

```
feat(tui/manage): faction creation wizard

- multi-step huh.Form covering all SWN creation steps:
  Scale, attribute assignment, HP (note), tags (cap 2), starting
  goal, homeworld, starting assets (rating + tech filter, quota
  per scale), starting Coin
- wizard.New(rb, spatial, initial) — initial=nil for create,
  pre-populated for edit; emits CreatedMsg or EditSavedMsg on
  completion based on entry point
- Esc opens internal discard-confirm overlay; y emits CancelMsg
- on completion: auto-creates Base of Influence at max HP on
  homeworld; derives MaxHP from attributes via CalcMaxHP
```

---

#### Commit 5 — `feat(tui/manage): edit form via wizard pre-population`

A small commit. The wizard already accepts a non-nil `initial` faction (Commit 4); this commit wires the `RequestEditMsg` path through manage's router and addresses edge cases surfaced by edit mode.

##### Task 1 — `internal/faction/tui/views/manage/manage.go`

Confirm the `RequestEditMsg` case (already scaffolded in Commit 2) calls `wizard.New(m.rulebook, m.spatialMap, msg.Faction)` correctly — Commit 2's stub may need adjusting now that the wizard constructor is real.

##### Task 2 — `internal/faction/tui/views/manage/wizard/wizard.go`

Edit-mode polish identified during Commit 4 implementation. Likely items:

- **Don't reseed stat defaults from Scale if `initial != nil` and stats are already set.** Edit mode opens with current stat values; Scale change in the form should ask whether to reseed or preserve. MVP: silently reseed (matches create behavior). Document in a `// TODO` if a richer UX is wanted.
- **Edit-mode discard-confirm** — same overlay; emits `CancelMsg` which manage routes back to detail (per `backTarget[viewEdit] = viewDetail`).
- **ID immutability in edit mode** — the wizard does not prompt for ID; `draft.ID = initial.ID` is preserved by `deepCopy`. Verify in execution.

##### Task 3 — `internal/faction/tui/views/manage/wizard/wizard_edit_test.go` (optional)

If execution-time testing is warranted: a unit test that constructs the wizard with a populated `*domain.Faction`, drives the form to completion without changing any values, and asserts the emitted `EditSavedMsg.Faction` is deep-equal to the input. Defer to execution-time call; bubbletea form tests are non-trivial.

##### Commit message

```
feat(tui/manage): edit form via wizard pre-population

- RequestEditMsg path constructs wizard.New with deep-copied current
  faction; on completion emits EditSavedMsg routed through state.UpdateFaction
- edit mode preserves ID; discard-confirm routes back to detail per
  backTarget[viewEdit]
- wizard reuse keeps form definition in one place; isCreate flag
  distinguishes completion message at exit
```

---

### Phase 4 — Presentation Polish

After this phase the context strip shows campaign metadata on the list view and the help overlay renders per-view bindings.

#### Commit 6 — `feat(tui/manage): context strip composed on list view`

##### Task 1 — `internal/faction/tui/views/manage/contextstrip/contextstrip.go`

Replace the stub with the real strip. Three-line snapshot: Campaign ID, Cycle (current → next), Faction count.

```go
package contextstrip

import (
    "fmt"

    "github.com/charmbracelet/lipgloss"

    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type Model struct {
    campaignID   string
    cycleCurrent int
    cycleNext    int
    factionCount int
}

func New(fs *state.FactionState) Model {
    return Model{
        campaignID:   fs.CampaignID,
        cycleCurrent: fs.CycleNumber,
        cycleNext:    fs.CycleNumber + 1,
        factionCount: len(fs.Factions),
    }
}

func (m Model) View() string {
    label := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
    value := lipgloss.NewStyle().Bold(true)

    lines := []string{
        label.Render("Campaign  ") + value.Render(m.campaignID),
        label.Render("Cycle     ") + value.Render(fmt.Sprintf("%d → %d", m.cycleCurrent, m.cycleNext)),
        label.Render("Factions  ") + value.Render(fmt.Sprintf("%d", m.factionCount)),
    }
    return lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("8")).
        Padding(1, 2).
        Width(28).
        Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
```

The strip is stateless; `manage.Update` reassigns `m.strip = contextstrip.New(m.factionState)` after each successful CRUD (already wired in Commit 2).

##### Task 2 — `internal/faction/tui/views/manage/manage.go`

Verify the composition path. `manage.View()`'s `viewList` branch joins `m.list.View()` and `m.strip.View()` horizontally (already sketched in Commit 2). Reserve ~30 cols for the strip in the list's `SetSize` call (Commit 2 sketched `msg.Width - 32`).

##### Commit message

```
feat(tui/manage): context strip composed on list view

- contextstrip component renders campaign / cycle / faction count
  as a bordered 3-line snapshot
- composed to the right of the list view by manage.View when view == viewList
- stateless: manage.Update reassigns strip after each successful CRUD
```

---

#### Commit 7 — `feat(tui): help overlay with per-view bindings`

Replaces Foundation's `showHelpStub` fixed string with a composed `help.View` reading per-view bindings via the `Helper` interface.

##### Task 1 — `internal/faction/tui/helper.go` (new file)

Per Shared Context § "Helper Interface and Help Dispatch":

```go
package tui

import "github.com/charmbracelet/bubbles/help"

type Helper interface {
    Help() help.KeyMap
}
```

##### Task 2 — `internal/faction/tui/model.go`

Add a `help.Model` field for rendering:

```go
import "github.com/charmbracelet/bubbles/help"

type Model struct {
    // ... existing fields
    help help.Model
}

// in NewModel:
help: help.New(),
```

##### Task 3 — `internal/faction/tui/update.go`

The existing `?` handler toggles `showHelpStub`. Keep the toggle; on render the toggle now drives a real help-overlay composition (see Task 4).

No changes to update.go for the toggle; the View change is the substantive piece.

Add global keymap declaration alongside the existing bindings:

```go
type globalKeys struct {
    Tab     key.Binding
    ShiftTab key.Binding
    Help    key.Binding
    Quit    key.Binding
}

var globals = globalKeys{
    Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next mode")),
    ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev mode")),
    Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
    Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}
```

##### Task 4 — `internal/faction/tui/view.go`

Replace the help stub branch:

```go
if m.showHelpStub {
    var subBindings help.KeyMap = emptyKeyMap{}
    if helper, ok := m.subs[m.bar.Active()].(Helper); ok {
        if km := helper.Help(); km != nil {
            subBindings = km
        }
    }
    composed := combineKeyMaps(globalsKeyMap{}, subBindings)
    return m.help.View(composed)
}

type emptyKeyMap struct{}
func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }

type globalsKeyMap struct{}
func (globalsKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{globals.Tab, globals.ShiftTab, globals.Help, globals.Quit}
}
func (globalsKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{{globals.Tab, globals.ShiftTab}, {globals.Help, globals.Quit}}
}

// combineKeyMaps merges two help.KeyMaps into one for the overlay.
func combineKeyMaps(a, b help.KeyMap) help.KeyMap { /* return a struct that concatenates */ }
```

The composition shape (single-row vs split-row, ordering) is a presentation call — pick at execution time. Charm's bubbles/help renders FullHelp as a multi-row layout; ShortHelp as one row.

##### Task 5 — `internal/faction/tui/views/manage/manage.go`

The `Help()` passthrough already exists in skeleton from Commit 2; this commit removes the `return nil` and fills:

```go
func (m Model) Help() help.KeyMap {
    switch m.view {
    case viewList:    return m.list.Help()
    case viewDetail:  return m.detail.Help()
    case viewCreate:  return m.create.Help()
    case viewEdit:    return m.edit.Help()
    case viewDeleteConfirm: return m.deleteConfirm.Help()
    }
    return nil
}
```

Each sub-view's `Help()` was implemented in its respective commit (2 / 3 / 4 / 5), so no per-sub-view changes needed here — only the manage passthrough and root composition.

##### Task 6 — `internal/faction/tui/views/turn/turn.go`

Add a placeholder `Help()` returning `nil` (or an empty keymap) so Turn satisfies `Helper` once Initiative 3 wires its bindings. No behavior change in this initiative.

##### Commit message

```
feat(tui): help overlay with per-view bindings

- add Helper interface; modes implement Help() passthrough to active sub
- manage.Help() routes to current view's sub-model Help()
- root ? handler composes globals (Tab/Shift-Tab/q/?) with active sub's
  KeyMap and renders via bubbles/help
- replaces Foundation's showHelpStub fixed string; toggle behavior preserved
```

---

## Verification

End-of-initiative verification (before pre-merge checklist):

1. **Compile + tests pass.** `go build ./... && go test ./internal/faction/state/... && go vet ./...`.
2. **End-to-end manual sweep against a real campaign.** With an active campaign that has rulebook + spatial seeded:
   - Launch: `gm-toolkit faction`.
   - List shows empty-state hint when no factions exist; centered "No factions yet." + "Press n to create one".
   - Press `n` → wizard walks all 8 SWN steps to completion → returns to list, new faction visible.
   - Press Enter on a row → detail shows all fields → `e` opens edit (wizard pre-populated) → modify name → submit → returns to detail with updated name.
   - `d` from detail opens delete-confirm → y deletes (file shrinks); n returns to detail.
   - `?` opens help with view-specific bindings; toggling between modes (Tab/Shift-Tab) updates the help overlay's per-view content.
   - Esc from edit/create opens discard-confirm; y discards; n returns to form.
   - Context strip on list view shows Campaign / Cycle / Count and updates after each CRUD.
3. **State persistence.** Restart the binary after each CRUD; verify the change persists.
4. **Engine smoke not regressed.** `gm-toolkit faction --dryrun` still passes (Foundation's smoke path is untouched but tui.Run signature changed — verify the dryrun entrypoint compiles against the new signature; Foundation's dryrun.go may need a small adjustment to construct a no-op spatial map if `factionRun` is no longer the only caller of `tui.Run`).
5. **No engine-boundary violations.** Senior-engineer review per arc-plan: sub-views never call `state.*` directly; only `manage.Update` does. Update discipline comment block in `update.go` still accurate.

## Pre-Merge Checklist

Per `CLAUDE.md` Collaboration item 8:

1. Senior-engineer code review of the branch.
2. Update `docs/dev_journals/faction-manager/planned-work.md`:
   - Remove F-005.2 from Current Initiatives and Up Next.
   - Add F-005.3 (tui-turn) to Up Next (trigger now met).
   - Capture deferred items that surfaced during execution (e.g., save-error rollback if it proved brittle, wizard step-back if requested, detail scroll polish, etc.).
3. Move the discovery and plan docs to their `completed/` subdirectories.
4. Per CLAUDE.md item 9: assess version bump (likely minor — substantive new user-visible feature) and tag accordingly.

Per `docs/process/session-modes.md` § 8 model-selection: switch to Sonnet for the pre-merge checklist session.
