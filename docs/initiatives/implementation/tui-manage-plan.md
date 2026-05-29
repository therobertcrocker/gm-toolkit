# tui-manage — Implementation Plan

## Context / Goal

This plan turns [`tui-manage-discovery.md`](../discovery/tui-manage-discovery.md) into an executable sequence of commits. Manage is Initiative 2 of the [TUI Rebuild arc](../arcs/tui-rebuild/tui-rebuild-arc-plan.md). After this initiative ships, the TUI covers list, detail, create (full SWN creation wizard), and delete; editing existing factions remains hand-edited TOML until a follow-up initiative ships (see Decision 2 for rationale).

Discovery ratified the structural decisions — router-with-sub-model-per-view, state-package CRUD ownership, completion-message pattern, context-strip placement, help-overlay shape. This plan ratifies the four open questions Discovery deferred, settles two implementation-time calls Discovery left to Plan (refresh shape, rollback semantics), splits edit into its own follow-up initiative (Decision 2), establishes the wizard → `domain.NewFaction` → state separation (Decision 12), and breaks the work into commits sized one-per-execution-session.

**Related artifacts:**
- Discovery: [`tui-manage-discovery.md`](../discovery/tui-manage-discovery.md)
- Arc-Plan: [`tui-rebuild-arc-plan.md`](../arcs/tui-rebuild/tui-rebuild-arc-plan.md)
- Arc-Discovery: [`tui-rebuild-arc-discovery.md`](../arcs/tui-rebuild/tui-rebuild-arc-discovery.md)
- Foundation Plan (exemplar): [`completed/tui-foundation-plan.md`](./completed/tui-foundation-plan.md)
- SWN creation checklist: [`docs/rules/swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md) § Faction Creation Checklist

## Decisions Ratified in Planning

1. **Create wizard walks the full SWN Faction Creation Checklist.** Resolves Discovery's "Faction CRUD form scope" open question — Robert chose the broader scope over identity-only. The wizard is a multi-step `huh.Form` covering: Scale, Attribute assignment (which stat is primary/secondary/tertiary), HP display (derived `huh.Note`), Tags (1–2 from rulebook), Starting Goal (from rulebook), Homeworld (world picker from spatial), Starting Assets (filtered by rating + tech-level, quota per scale), Starting Coin (GM discretion, numeric input). The auto-derived steps (HP, homeworld Base of Influence at max HP) appear as confirmation notes, not interactive inputs.
2. **Edit deferred to a follow-up initiative.** This initiative ships list + detail + create + delete only. The original plan reused the create wizard for edit; that decision surfaced a structural conflict — the wizard's "starting assets" and "starting Coin" steps don't map onto a faction with accrued runtime state (Coin spent, assets bought/destroyed). Rather than narrow edit's scope mid-plan, edit is split into its own initiative. **UX gap acknowledged:** GMs modify existing factions by hand-editing TOML until the edit initiative ships.
3. **Discard-confirm is internal state on the create model.** Resolves Discovery's "structural home" choice. The wizard gains `confirmingDiscard bool`; View overlays a y/n prompt when set; Update intercepts y/n locally and emits `CancelMsg` on y. No new package, no `viewDiscardConfirm` router branch. The prompt fires unconditionally per Discovery (no "only-when-dirty" gate).
4. **Help dispatch: modes implement a `Help()` passthrough.** Resolves Discovery's "root vs. passthrough" choice. Manage's `Help() help.KeyMap` returns its active sub-model's `Help()`. Root holds a `Helper` interface; `?` handler calls `subs[m.bar.Active()].(Helper).Help()` and composes with globals. Root stays agnostic of any mode's internal view-stack shape; Turn slots in the same way.
5. **Empty-list grammar: centered hint with key prompt.** Resolves Discovery's "empty-state UX" choice. List area renders `"No factions yet."` in normal style on one line, dimmed `"Press n to create one"` below. Turn (Initiative 3) inherits: `"No factions in roster."` + `"Press n in Manage to add one."`.
6. **Spatial data loaded alongside rulebook in `factionRun`.** The wizard's homeworld picker needs the world list; the wizard's asset step needs each world's `TechLevel()`. `cmd/gm-toolkit/faction.go` adds `spatial.LoadRegionMap(paths.SpatialDataDir)` after the rulebook load; `tui.Run` signature expands to take a `*spatial.RegionMap`. Manage receives it via the root Model. Foundation's existing data flow (rulebook → tui.Run → root → adapter) is the template — spatial threads the same path. If `LoadRegionMap` returns an error (no spatial seeded), `factionRun` surfaces it with a `hint:` line per the existing rulebook error pattern; the seed command is TBD per the F-012 backlog item, so the hint text is left to execution-time re-grounding.
7. **State CRUD lives in a new file: `internal/faction/state/crud.go`.** Keeps `faction_state.go`'s `Load`/`Save` contract clean; CRUD is its own concern. CRUD surface is `CreateFaction` + `DeleteFaction` only — `UpdateFaction` has no consumer until the edit initiative ships (the engine mutates Faction objects in place and calls `Save` directly); `GetFaction` is unused (callers access `fs.Factions[id]` directly with map indexing, valid since IDs flow from the list). Sentinels: `ErrFactionAlreadyExists`, `ErrFactionNotFound`, `ErrInvalidFactionID`. Save errors do not roll back the in-memory mutation — per Discovery's MVP framing. Inline `SaveErrorMsg` surfaces the error; user retries. Revisit only if execution surfaces brittleness.
8. **Context strip refresh: `manage.Update` reassigns the strip after a successful CRUD message.** After `CreatedMsg` / `DeletedMsg` is handled (state CRUD succeeds), `manage.Update` calls `m.strip = contextstrip.New(m.factionState)`. The strip is a stateless snapshot constructor — cheaper than a stateful `Refresh()` method, and the recompute cost (read three fields from FactionState) is trivial.
9. **`bubbles/list` powers the list view; `bubbles/help` powers the help overlay.** Add these in the commits that first use them — list in Commit 2, help in Commit 6. No global "dependency bump" commit; tracked changes land with the consumer.
10. **List view's item shape: Name · Scale · current/max HP · Coin.** Single-column `list.Item`. Selection state and j/k/up/down scrolling inherited from `bubbles/list`. Faction `ID` shown dimmed in the detail view, not the list (factions are GM-facing; names are the working identifier).
11. **Six commits across four phases.** Phase 1 (state CRUD): 1 commit. Phase 2 (router + read-only views): 2 commits. Phase 3 (wizard): 1 commit. Phase 4 (presentation polish): 2 commits. Arc-Plan estimated 5–8; this lands within range. The Phase 3 create-wizard commit (Commit 4) is the heaviest — flagged for execution-time re-grounding and possible mid-phase split if it grows past one session.
12. **Wizard collects inputs; `domain.NewFaction(...)` constructs; state persists.** Derivation logic (HP via `CalcMaxHP`, attribute ratings via `RatingsFromScale`, asset quota via `AssetCountsFromScale`, Base of Influence auto-creation at MaxHP on homeworld) lives in `domain`, not `wizard`. The wizard holds bound primitives for each form step (scale, picked stats, tags, goal, homeworld, asset picks, coin); at form completion, calls `domain.NewFaction(...)` to synthesize the `*domain.Faction`. `state.CreateFaction` is a thin persistence boundary that validates uniqueness, applies the map mutation, and saves. This keeps three roles separate: TUI = input collection, domain = rules and construction, state = persistence.

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
│   │   │   └── crud.go                  # NEW — Create/Delete + sentinels
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
│   │           │   └── detail.go        # display all faction fields; d/Esc
│   │           ├── deleteconfirm/
│   │           │   └── deleteconfirm.go # y/n confirm
│   │           ├── wizard/
│   │           │   └── wizard.go        # huh multi-step form; create-only (edit deferred)
│   │           └── contextstrip/
│   │               └── contextstrip.go  # campaign / cycle / faction count
│   └── spatial/                          # untouched (loader already exists)
```

The `views/manage/` directory replaces the placeholder `manage.go` shell shipped by Foundation. Foundation's `tui.Placeholder` style remains in use for sub-view stubs during incremental commits and for the (still-empty) Turn shell.

### Dependency Additions

Added incrementally with the consumer. No standalone "deps" commit.

- **Commit 2** — `github.com/charmbracelet/bubbles/list` (list view).
- **Commit 4** — no new direct deps; `huh` is already in `go.mod` from Foundation.
- **Commit 6** — `github.com/charmbracelet/bubbles/help` and `github.com/charmbracelet/bubbles/key` (help overlay + KeyMap).

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

// DeleteFaction validates that id exists in fs.Factions, removes the entry,
// then persists. Returns ErrFactionNotFound if absent.
func DeleteFaction(path string, fs *FactionState, id string) error
```

Error wrapping follows the campaign-manager pattern (`fmt.Errorf("state: <op>: %w", err)`); sentinels are checked with `errors.Is`.

`UpdateFaction` is intentionally absent — no consumer exists in this initiative. The engine mutates `*domain.Faction` objects in place and calls `state.Save` at turn boundaries; the wizard is create-only; the edit initiative will add `UpdateFaction` when it ships. `GetFaction` is also absent — read access uses direct map indexing (`fs.Factions[id]`), since IDs in the TUI flow from the list (valid by construction).

### `domain.NewFaction` Constructor

Lives in `internal/faction/domain/faction.go`. Composes existing derivation primitives (`RatingsFromScale`, `CalcMaxHP`) and creates the Base of Influence on the homeworld at MaxHP. Wizard calls this at form completion; the returned `*domain.Faction` is what flows into `manage.CreatedMsg` and eventually `state.CreateFaction`.

```go
// NewFaction synthesizes a Faction from creation-time inputs. Applies the
// SWN derivations (attribute ratings from scale, MaxHP from attributes,
// CurrentHP = MaxHP, BoI at MaxHP on homeworld). XP starts at 0.
func NewFaction(
    id, name string,
    scale FactionScale,
    primaryStat, secondaryStat, tertiaryStat FactionStat,
    tags []*Tag,
    goal *Goal,
    homeworld Location,
    assets []*Asset,
    coin int,
) *Faction
```

Implementation outline (mechanical — primitives all exist):

1. Build the `Faction` struct with `ID`, `Name`, `Scale`, `Tags`, `Coin`, `Homeworld`, `Assets` map (keyed by `Asset.ID`).
2. Call `RatingsFromScale(scale)` → assign the three values to `Force`/`Cunning`/`Wealth` based on the picked `primaryStat`/`secondaryStat`/`tertiaryStat`.
3. Set `MaxHP = CalcMaxHP(faction)`; `CurrentHP = MaxHP`.
4. If `goal != nil`, set `ActiveGoal = &ActiveGoal{GoalID: goal.ID}` (zero progress).
5. Append a `*Base` to `faction.Bases` for the homeworld with `CurrentHP = MaxHP`, `Influence = 0` (or scale-derived default at execution time), `IsHomeworld = true`.
6. Return the faction.

`Asset` instantiation (with `NextAssetID`-generated IDs) happens upstream in the wizard from picked `*AssetDefinition` values, since `NextAssetID` requires the faction-in-progress and the wizard owns asset selection UX. The constructor receives fully-formed `*Asset` slices.

Faction `ID` generation lives outside `NewFaction` — see Open Question 3.

### Completion Message Types (Final Shape)

All in `internal/faction/tui/views/manage/messages.go`:

```go
package manage

import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"

type CreatedMsg struct{ Faction *domain.Faction }
type DeletedMsg struct{ ID string }
type CancelMsg struct{} // discard from create, no from deleteconfirm
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
    case viewDeleteConfirm: return m.deleteConfirm.Help()
    }
    return emptyKeyMap{}
}
```

Root's `?` handler (in `update.go`) checks whether the active sub satisfies `Helper`; if so, calls `Help()` and renders via `help.New().View(km)` composed with the globals' keymap. Foundation's `showHelpStub` toggle stays; the fixed string is replaced by the composed render.

### Faction Creation Wizard — Step Inventory

The `wizard.Model` holds bound primitives for each step (one field per selection) and walks `huh.NewForm(...)` groups. At completion, the wizard calls `domain.NewFaction(...)` (see Shared Context § `domain.NewFaction` Constructor) to synthesize the `*domain.Faction`. No "draft" object is maintained progressively — the form binds to primitives, the constructor synthesizes at the end. Step order matches the SWN checklist:

| # | Step | huh widget | Source data | Notes |
|---|------|-----------|-------------|-------|
| 1 | Scale | `huh.NewSelect[domain.FactionScale]` | constants in `domain` | Picked value flows into `NewFaction` for `RatingsFromScale` lookup. |
| 2 | Attribute assignment | three `huh.NewSelect[domain.FactionStat]` | derived | User picks which stat is primary, secondary, tertiary. Validator: all three distinct. Picks flow into `NewFaction` as `primaryStat`/`secondaryStat`/`tertiaryStat`. |
| 3 | HP display | `huh.NewNote` | derived | Shows computed MaxHP — wizard precomputes via `RatingsFromScale` + `CalcMaxHP` on a throwaway `*Faction` for display only; the authoritative value comes from `NewFaction` at completion. |
| 4 | Tags | `huh.NewMultiSelect[*domain.Tag]` | `rulebook.Tags` | Cap 2 via validator. |
| 5 | Goal | `huh.NewSelect[*domain.Goal]` | `rulebook.Goals` | Picked goal flows into `NewFaction`; the constructor sets `ActiveGoal{GoalID: goal.ID}`. |
| 6 | Homeworld | `huh.NewSelect[string]` (world IDs) | spatial worlds | `NewFaction` auto-creates Base of Influence at MaxHP on the picked world. |
| 7 | Starting assets | `huh.NewMultiSelect[*domain.AssetDefinition]` (×2 groups) | `rulebook.AssetDefinitions` filtered by rating + homeworld tech_level | Quota by scale (see Open Question 2). Wizard instantiates picked definitions as `*domain.Asset` (using `NextAssetID` against the in-flight asset slice) before handing to `NewFaction`. |
| 8 | Starting Coin | `huh.NewInput` (int validator) | — | Default 0. Parsed `int` flows into `NewFaction`. |

Wizard constructor signature:

```go
func New(
    rulebook *rulebook.Rulebook,
    spatialMap *spatial.RegionMap,
) Model
```

Internal `confirmingDiscard bool` overlays the y/n prompt on Esc; on y, emits `CancelMsg`; on n, returns to the form. The `huh.Form` is driven via standard `tea.Model` plumbing (huh forms satisfy `tea.Model`).

Completion: when `wizard.Model.form.State == huh.StateCompleted`, the wizard:

1. Generates a faction ID (Open Question 3 — likely `slug(name)` with collision handling).
2. Builds the `Location` value (`{WorldID: homeworldID, RegionHex: ...}`) from the spatial map.
3. Instantiates picked `*AssetDefinition` values as `*Asset` (HP from definition, location = homeworld, etc.).
4. Calls `domain.NewFaction(id, name, scale, primaryStat, secondaryStat, tertiaryStat, tags, goal, homeworld, assets, coin)`.
5. Emits `CreatedMsg{Faction: faction}`.

ID generation needs read access to `fs.Factions` for uniqueness check. Two implementation paths (pick at execution time): (a) `wizard.New` takes `*state.FactionState` as a fourth param; or (b) the wizard emits an inputs-bundle message and `manage.Update` computes the ID before calling `state.CreateFaction`. Option (a) is simpler and keeps construction in one place.

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

- **Edit faction.** Deferred to a follow-up initiative — see Decision 2. GMs hand-edit TOML to modify an existing faction until that initiative ships. The wizard-reuse path surfaced a structural conflict (starting-assets and starting-Coin steps don't apply to a faction with accrued runtime state); the cleaner shape is a dedicated edit flow scoped to identity-only fields, designed in its own discovery session.
- **Spatial seed command** (F-012). Wizard expects spatial data to exist; missing-data error surfaces with a `hint:` line per the rulebook pattern. Seed command itself is a separate initiative.
- **`bubbles/help` per-mode color theming.** Default Charm theme until a theming initiative ships.
- **Wizard step-skipping or step-back navigation.** `huh.Form` supports `Group` navigation; the wizard uses default Group order without custom skip logic. Esc opens discard-confirm; there is no "back one step" key.
- **Detail view scroll polish.** A long faction's detail may overflow; basic scrolling via `bubbles/viewport` is OK if needed, but pixel-perfect layout is out of scope.

---

## Work Breakdown

### Phase 1 — State Package CRUD

Adds the write-through CRUD surface the TUI consumes. Standalone, no TUI dependency. After this phase the state package exposes Create + Delete; nothing in the TUI uses them yet.

#### Commit 1 — `feat(state): faction CRUD with write-through persistence`

##### Task 1 — `internal/faction/state/crud.go` (new file)

Implement the two functions and three sentinels per Shared Context § "State CRUD Function Signatures (Final Shape)". Concrete implementation sketch:

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
```

##### Task 2 — `internal/faction/state/crud_test.go` (new file)

Table-driven tests covering: create + read-back via direct map access; create duplicate → `ErrFactionAlreadyExists`; delete missing → `ErrFactionNotFound`; create empty ID → `ErrInvalidFactionID`; delete empty ID → `ErrInvalidFactionID`; round-trip create → Save → Load returns equivalent faction. Use `t.TempDir()` for the path. Mirrors the testing style in `faction_state.go`'s existing tests (verify shape at execution time).

##### Commit message

```
feat(state): faction CRUD with write-through persistence

- add CreateFaction / DeleteFaction in internal/faction/state/crud.go
- declare ErrFactionAlreadyExists / ErrFactionNotFound / ErrInvalidFactionID
  sentinels; check with errors.Is
- each mutation validates invariants, applies in-memory, then persists via
  Save in one operation; no rollback on Save failure
```

---

### Phase 2 — Manage Router and Read-Only Views

Stands up the Manage router, the list view, the detail view, and the delete-confirm sub-view. After this phase a GM can browse and delete factions but cannot create one. The create wizard follows in Phase 3.

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
    viewDeleteConfirm
)

var backTarget = map[view]view{
    viewDetail:        viewList,
    viewCreate:        viewList,
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
type RequestDeleteMsg struct{ Faction *domain.Faction }  // detail → deleteConfirm
```

Each handled in `manage.Update`:

```go
case RequestDetailMsg:
    faction := m.factionState.Factions[msg.FactionID]
    m.detail = detail.New(faction)
    m.view = viewDetail
    return m, nil
case RequestCreateMsg:
    m.create = wizard.New(m.rulebook, m.spatialMap)
    m.view = viewCreate
    return m, m.create.Init()
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
    case viewDeleteConfirm:
        content = m.deleteConfirm.View()
    }
    if m.saveErr != nil {
        content += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("save error: " + m.saveErr.Error())
    }
    return content
}

func (m Model) Help() help.KeyMap {
    // populated in Commit 6
    return nil
}
```

The `Help()` method exists with a `nil` return until Commit 6 wires the per-view bindings. Root's `?` handler tolerates `nil` (renders the existing stub).

##### Task 6 — Sub-view package stubs

Create the following files so `manage.go` compiles. Each is a minimal `tea.Model` with an empty Update and a single Placeholder view; the real implementations land in Commits 3–5.

- `internal/faction/tui/views/manage/list/list.go` — IMPLEMENTED (full list view, this commit; see Task 7).
- `internal/faction/tui/views/manage/detail/detail.go` — stub Model + New + Init/Update/View returning `tui.Placeholder.Render("Detail — Commit 3")`.
- `internal/faction/tui/views/manage/deleteconfirm/deleteconfirm.go` — stub Model + `New(*domain.Faction)` + stubs returning placeholder.
- `internal/faction/tui/views/manage/wizard/wizard.go` — stub Model + `New(*rulebook.Rulebook, *spatial.RegionMap)` + stubs returning placeholder.
- `internal/faction/tui/views/manage/contextstrip/contextstrip.go` — stub Model + `New(*state.FactionState)` + stubs returning placeholder (`tui.Placeholder.Render("Strip — Commit 5")`).

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
  packages (filled in Commits 3-5)
- load spatial alongside rulebook in factionRun; expand tui.Run
  signature to accept *rulebook.Rulebook and *spatial.RegionMap
```

---

#### Commit 3 — `feat(tui/manage): detail view and delete-confirm`

Fills in the two read-only sub-views. Detail shows the full faction; `d` opens delete-confirm; delete-confirm y/n emits `DeletedMsg` or `CancelMsg`.

##### Task 1 — `internal/faction/tui/views/manage/detail/detail.go`

Replaces the stub. Renders a faction's full state as a single scrollable block. `d` emits `RequestDeleteMsg`; Esc emits `CancelMsg` (which `manage.Update` routes back to list).

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
    Delete key.Binding
    Back   key.Binding
}

func New(faction *domain.Faction) Model {
    return Model{
        faction: faction,
        keys: keyMap{
            Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
            Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
        },
    }
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if km, ok := msg.(tea.KeyMsg); ok {
        switch {
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
    fmt.Fprintf(&b, "Homeworld: %s\n", f.Homeworld.WorldID)
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
        fmt.Fprintf(&b, "  - %s\n", base.Location.WorldID)
    }
    return b.String()
}

func (m Model) Help() help.KeyMap { return helpKeys{m.keys} }

type helpKeys struct{ keys keyMap }
func (h helpKeys) ShortHelp() []key.Binding {
    return []key.Binding{h.keys.Delete, h.keys.Back}
}
func (h helpKeys) FullHelp() [][]key.Binding {
    return [][]key.Binding{{h.keys.Delete, h.keys.Back}}
}
```

Verify `domain.Faction`'s field set at re-grounding — confirmed shape: `Tags []*Tag`, `ActiveGoal *ActiveGoal`, `Assets map[string]*Asset`, `Bases []*Base`, `Homeworld Location` (struct, not interface — fields `WorldID string` and `RegionHex spatial.RegionHex`). `Base.Location` is also a `Location` struct (per `internal/faction/domain/base.go`). Adjust on drift.

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
  goal/assets/bases; d/Esc keys emit RequestDeleteMsg / CancelMsg
- deleteconfirm: y/n prompt; y emits DeletedMsg, n/Esc emits CancelMsg
- both sub-views expose Help() with their bindings (consumed in Commit 6)
```

---

### Phase 3 — Create Wizard

After this phase the GM can author new factions via the SWN creation wizard. Edit is deferred to a follow-up initiative (see Decision 2).

#### Commit 4 — `feat(tui/manage): faction creation wizard`

**Flagged as the heaviest commit in the plan.** Implements `domain.NewFaction` (the Task 0 prerequisite) and the multi-step `huh.Form` covering all SWN creation steps. If the execution session surfaces complexity that warrants splitting (e.g., asset-step quota logic balloons), pause and re-plan: extract the asset step into a Commit 4b. Default is single commit.

##### Task 0 — `internal/faction/domain/faction.go` (add `NewFaction`)

Add the constructor per Shared Context § "`domain.NewFaction` Constructor". Mechanical — composes primitives that already exist (`RatingsFromScale`, `CalcMaxHP`). Adds:

- `NewFaction(...)` returning `*Faction` with attributes, MaxHP, CurrentHP, ActiveGoal, Assets map, Bases (including homeworld BoI) all populated.
- A small test: pick a Minor faction with primary=Force, secondary=Cunning, tertiary=Wealth; assert returned Force/Cunning/Wealth match `RatingsFromScale(ScaleMinor)`; assert `MaxHP == CalcMaxHP(...)`; assert exactly one Base with `IsHomeworld == true` at `MaxHP` on the picked homeworld.

This must land before Task 1 — the wizard's completion path imports `domain.NewFaction`.

##### Task 1 — `internal/faction/tui/views/manage/wizard/wizard.go`

Replace the stub with the full wizard. The wizard holds bound primitives for each step (no progressive draft); at completion, calls `domain.NewFaction(...)` to synthesize the faction. Structure:

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
    "github.com/therobertcrocker/gm-toolkit/internal/faction/state"
    "github.com/therobertcrocker/gm-toolkit/internal/faction/tui/views/manage"
    "github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

type Model struct {
    form       *huh.Form
    rulebook   *rulebook.Rulebook
    spatialMap *spatial.RegionMap
    factionState *state.FactionState // for ID-uniqueness check at completion

    // Bound form values — huh writes into these via Value(&...).
    name           string
    scale          domain.FactionScale
    primaryStat    domain.FactionStat
    secondaryStat  domain.FactionStat
    tertiaryStat   domain.FactionStat
    selectedTags   []*domain.Tag
    selectedGoal   *domain.Goal
    homeworldID    string
    selectedAssets []*domain.AssetDefinition
    coinStr        string

    confirmingDiscard bool
    discardKeys       discardKeyMap
}

type discardKeyMap struct {
    Confirm key.Binding
    Cancel  key.Binding
}

func New(rb *rulebook.Rulebook, spatialMap *spatial.RegionMap, fs *state.FactionState) Model {
    m := Model{
        rulebook:     rb,
        spatialMap:   spatialMap,
        factionState: fs,
        discardKeys: discardKeyMap{
            Confirm: key.NewBinding(key.WithKeys("y")),
            Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
        },
    }
    m.form = buildForm(&m, rb, spatialMap)
    return m
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
        faction := m.synthesize()
        return m, func() tea.Msg { return manage.CreatedMsg{Faction: faction} }
    }
    return m, cmd
}

// synthesize builds the final *domain.Faction from collected form values.
// Calls domain.NewFaction with all derivations applied; assigns a unique ID.
func (m Model) synthesize() *domain.Faction {
    coin, _ := strconv.Atoi(m.coinStr)
    id := generateID(m.name, m.factionState.Factions) // Open Question 3
    homeworld := domain.Location{
        WorldID:   m.homeworldID,
        RegionHex: m.spatialMap.RegionHexFor(m.homeworldID), // verify API at re-grounding
    }
    assets := instantiateAssets(m.selectedAssets, homeworld) // generate Asset IDs, set Location/HP/etc.
    return domain.NewFaction(
        id, m.name,
        m.scale,
        m.primaryStat, m.secondaryStat, m.tertiaryStat,
        m.selectedTags,
        m.selectedGoal,
        homeworld,
        assets,
        coin,
    )
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

`buildForm` constructs the multi-group `huh.Form`. Each `huh.NewGroup(...)` covers one or more SWN-checklist steps. The form binds directly to the wizard Model's fields; no progressive draft. Pseudocode per Shared Context § "Faction Creation Wizard — Step Inventory":

```go
func buildForm(m *Model, rb *rulebook.Rulebook, spatialMap *spatial.RegionMap) *huh.Form {
    // Step 0 — Name (added during execution; not in original step inventory)
    nameGroup := huh.NewGroup(
        huh.NewInput().Title("Faction name").Value(&m.name).
            Validate(func(s string) error {
                if s == "" { return fmt.Errorf("name required") }
                return nil
            }),
    )

    // Step 1 — Scale
    scaleGroup := huh.NewGroup(
        huh.NewSelect[domain.FactionScale]().
            Title("Scale").
            Options(
                huh.NewOption("Minor", domain.ScaleMinor),
                huh.NewOption("Major", domain.ScaleMajor),
                huh.NewOption("Hegemon", domain.ScaleHegemon),
            ).
            Value(&m.scale),
    )

    // Step 2 — Attribute assignment (which stat is primary/secondary/tertiary)
    // Validator ensures all three are distinct.
    statsGroup := huh.NewGroup(
        huh.NewSelect[domain.FactionStat]().Title("Primary attribute").
            Options(statOptions...).Value(&m.primaryStat),
        huh.NewSelect[domain.FactionStat]().Title("Secondary attribute").
            Options(statOptions...).Value(&m.secondaryStat),
        huh.NewSelect[domain.FactionStat]().Title("Tertiary attribute").
            Options(statOptions...).Value(&m.tertiaryStat),
    ).WithHide(func() bool { return m.scale == "" })

    // Step 3 — HP display (note). Precomputes for display only; authoritative
    // MaxHP comes from domain.NewFaction at completion.
    hpGroup := huh.NewGroup(
        huh.NewNote().Title("Max HP").DescriptionFunc(func() string {
            return fmt.Sprintf("%d HP (auto-derived from attributes)", previewMaxHP(m))
        }, nil),
    )

    // Step 4 — Tags
    tagOptions := tagsToOptions(rb.Tags)
    tagsGroup := huh.NewGroup(
        huh.NewMultiSelect[*domain.Tag]().
            Title("Tags (max 2)").
            Options(tagOptions...).
            Value(&m.selectedTags).
            Validate(func(t []*domain.Tag) error {
                if len(t) > 2 { return fmt.Errorf("at most 2 tags") }
                return nil
            }),
    )

    // Step 5 — Starting Goal
    goalOptions := goalsToOptions(rb.Goals)
    goalGroup := huh.NewGroup(
        huh.NewSelect[*domain.Goal]().
            Title("Starting goal").
            Options(goalOptions...).
            Value(&m.selectedGoal),
    )

    // Step 6 — Homeworld
    worldOptions := worldsToOptions(spatialMap)
    homeworldGroup := huh.NewGroup(
        huh.NewSelect[string]().
            Title("Homeworld").
            Options(worldOptions...).
            Value(&m.homeworldID),
    )

    // Step 7 — Starting assets (filtered by rating + homeworld tech)
    // See Open Question 2 for two-group vs single-group shape.
    assetGroup := buildAssetGroup(m, rb, spatialMap)

    // Step 8 — Starting Coin
    coinGroup := huh.NewGroup(
        huh.NewInput().
            Title("Starting Coin").
            Value(&m.coinStr).
            Validate(func(s string) error {
                if _, err := strconv.Atoi(s); err != nil { return fmt.Errorf("must be an integer") }
                return nil
            }),
    )

    return huh.NewForm(
        nameGroup, scaleGroup, statsGroup, hpGroup, tagsGroup, goalGroup, homeworldGroup, assetGroup, coinGroup,
    )
}
```

Key implementation details to nail at execution time:

1. **The HP-preview closure** — `previewMaxHP(m)` constructs a throwaway `*Faction` with `RatingsFromScale(m.scale)` applied to the picked stats, calls `domain.CalcMaxHP`, returns the int. Used only for display in Step 3; the authoritative MaxHP comes from `domain.NewFaction` at completion. If `huh` doesn't support dynamic descriptions in `Note`, drop Step 3 to a static message or compute on-completion only.
2. **Asset step shape** — Open Question 2. Two `huh.MultiSelect` groups (one filtered to primary attribute, one to all) is the cleanest UX; collapse to one group with a single quota validator if `huh` makes the two-group flow awkward.
3. **Faction ID generation** — Open Question 3. Recommended: `slug(m.name)` with collision check against `m.factionState.Factions` inside `synthesize()`. On collision, append `-2`, `-3`, etc.; if even that becomes a user-facing concern, a hidden Step 9 (ID input) appears only on collision. Pick at execution time.

##### Task 2 — Helper functions in `wizard.go`

- `tagsToOptions([]*domain.Tag) []huh.Option[*domain.Tag]`
- `goalsToOptions([]*domain.Goal) []huh.Option[*domain.Goal]`
- `worldsToOptions(*spatial.RegionMap) []huh.Option[string]` — iterates spatial map, returns `{Title: world.Name() + " (TL " + world.TechLevel() + ")", Value: world.ID()}` options.
- `buildAssetGroup(m, rb, spatialMap) *huh.Group` — filters `rb.AssetDefinitions` by attribute rating and homeworld tech_level; writes picks into `m.selectedAssets`; constructs the multi-select(s) per Open Question 2.
- `previewMaxHP(m *Model) int` — constructs a throwaway `*domain.Faction` with the current scale/stat picks projected via `RatingsFromScale`, returns `domain.CalcMaxHP(...)`.
- `instantiateAssets([]*domain.AssetDefinition, domain.Location) []*domain.Asset` — converts picked definitions to fully-formed `*Asset` values (HP from definition, location = homeworld, IDs from `NextAssetID`).
- `generateID(name string, existing map[string]*domain.Faction) string` — Open Question 3 implementation.

Verify all helper data sources at re-grounding — `rulebook.Rulebook` field names (`Tags`, `Goals`, `AssetDefinitions`), `spatial.RegionMap`'s world iteration surface, `NextAssetID` signature in `internal/faction/domain/asset.go`.

##### Task 3 — `manage.go` constructor call update

Commit 2 scaffolded `wizard.New(m.rulebook, m.spatialMap)` in the `RequestCreateMsg` handler with a stub signature. Update the call site to pass `m.factionState` as the third arg now that the wizard needs it for ID-uniqueness checks:

```go
case RequestCreateMsg:
    m.create = wizard.New(m.rulebook, m.spatialMap, m.factionState)
    m.view = viewCreate
    return m, m.create.Init()
```

##### Commit message

```
feat(tui/manage): faction creation wizard

- add domain.NewFaction constructor — applies RatingsFromScale,
  CalcMaxHP, BoI auto-create on homeworld; wizard calls at completion
- multi-step huh.Form covering all SWN creation steps:
  name, scale, attribute assignment, HP (note), tags (cap 2),
  starting goal, homeworld, starting assets (rating + tech filter,
  quota per scale), starting Coin
- wizard.New(rb, spatial, fs) holds bound primitives; on completion
  synthesizes *domain.Faction via domain.NewFaction and emits CreatedMsg
- Esc opens internal discard-confirm overlay; y emits CancelMsg
```

---

### Phase 4 — Presentation Polish

After this phase the context strip shows campaign metadata on the list view and the help overlay renders per-view bindings.

#### Commit 5 — `feat(tui/manage): context strip composed on list view`

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

#### Commit 6 — `feat(tui): help overlay with per-view bindings`

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
    case viewDeleteConfirm: return m.deleteConfirm.Help()
    }
    return nil
}
```

Each sub-view's `Help()` was implemented in its respective commit (2 / 3 / 4), so no per-sub-view changes needed here — only the manage passthrough and root composition.

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
   - Press `n` → wizard walks all SWN steps (name through starting Coin) to completion → returns to list, new faction visible.
   - Press Enter on a row → detail shows all fields (name, ID, scale, stats, HP/Coin/XP, homeworld, tags, goal, assets, bases).
   - `d` from detail opens delete-confirm → y deletes (file shrinks); n returns to detail.
   - `?` opens help with view-specific bindings; toggling between modes (Tab/Shift-Tab) updates the help overlay's per-view content.
   - Esc from create wizard opens discard-confirm; y discards; n returns to form.
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
   - Add a new entry for the deferred **edit-faction** initiative — identity-only scope, follow-up to this one, blocks closure of "TUI replaces hand-edited TOML" goal.
   - Capture other deferred items that surfaced during execution (e.g., save-error rollback if it proved brittle, wizard step-back if requested, detail scroll polish, etc.).
3. Move the discovery and plan docs to their `completed/` subdirectories.
4. Per CLAUDE.md item 9: assess version bump (likely minor — substantive new user-visible feature) and tag accordingly.

Per `docs/process/session-modes.md` § 8 model-selection: switch to Sonnet for the pre-merge checklist session.
