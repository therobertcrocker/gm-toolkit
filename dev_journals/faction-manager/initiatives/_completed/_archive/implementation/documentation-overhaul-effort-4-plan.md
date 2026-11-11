# Documentation Overhaul — Effort 4 Plan (Contributor Guides)

> Per-effort plan under `documentation-overhaul-plan.md` (the overview plan), per
> `session-modes.md` section 8. This effort builds `docs/contributing/`: a shared
> front door plus two per-side guides, organized around per-subsystem **extension
> recipes** that cross-link the Effort 3 architecture pages for shape and rationale
> rather than re-narrating them.
>
> The recipe inventory, registration sites, and test surfaces below are anchored to
> the Effort 3 architecture pages — themselves source-verified against commit
> `1ff3100`. Execution sessions re-verify each recipe's file paths and registration
> sites against current source before writing, but do not re-derive the inventory.

## Context / Goal

Effort 4 produces the "how do I extend this?" layer the doc set has been missing.
Efforts 1–3 settled *what the system is* (the architecture wiki) and *how we work*
(the process docs). The contributor guides answer the next question a developer
asks: given a subsystem, what are the concrete steps to add one more of its kind —
one more action, one more tag handler, one more overlay.

**Shape reversal (overview plan Decision 6).** The overview plan scoped the
contributor guides as two task-oriented, *not*-wiki-fied docs, with Effort 4
classified small (straight-to-execution). Robert reversed the shape toward a
**hybrid**: the guides are organized around per-subsystem **extension recipes**
("Add an action," "Implement a tag handler," "Wire a prompt"), each recipe a short
procedural unit that links its architecture page for the *why* and *shape* and
spends its own words only on *steps*. That reversal is why Effort 4 earns a Plan
session at all. It is logged in the overview plan's Decision Record (Effort 4) and
ratified below.

The reversal does **not** discard Decision 6's task-orientation — it sharpens it.
The guides stay task-oriented; the recipe is the unit of task. What changes is that
the guides now have internal structure keyed to the same subsystem axis as the arch
wiki, so a contributor moves *arch page → recipe* on one axis instead of hunting a
flat prose doc.

## Decisions Ratified in Planning

1. **Three deliverable docs, not two: a shared front door plus two side guides.**
   Setup and build are identical for both sides (one Go module, one binary); only
   test tooling and conventions differ. `docs/contributing/overview.md` is the
   front door — setup, build, run, and the *shared extension pattern* — and routes
   to `engine.md` and `interface.md`, which carry only side-specific test tooling,
   conventions, and recipes. This mirrors the arch wiki's spine-plus-pages shape and
   keeps the one-Go-module setup in exactly one place. (Expands Decision 6's
   two-doc framing to three; logged as a reconciliation in the overview Decision
   Record.)

2. **The recipe is the unit of the guide. Fixed shape:** a one-line intent + an
   arch-page cross-link, then numbered **steps** (each naming the file(s) to touch
   and what to register/wire), then **Test** (the harness + what to assert), then
   **See also** (the sibling recipe across the engine/interface boundary, where the
   task crosses it). Skeleton in Shared Context. The cross-link carries the shape
   and the "why"; the recipe never repeats them.

3. **The arch-page boundary is a hard rule: recipes link, never duplicate.** A
   recipe states *what to do*; the linked architecture page states *what the
   subsystem is and why it's shaped that way*. No recipe re-narrates a `Key
   Decisions` entry — it links to it. The same discipline governs the guides'
   **Conventions** sections: each convention is a one-line "rule you'll trip over"
   plus a link to the spine's `Cross-subsystem conventions` or the relevant page's
   `Key Decisions`, not a re-statement. This guards the just-built wiki against a
   second, drifting copy of itself in the contributor guides.

4. **The recipe template stays in this plan; it is not promoted to
   `docs/process/templates/`.** Unlike `arch-page.md` (one file *per subsystem*, so a
   copy-template pays off), recipes are *sections within* two guides — there are no
   new per-recipe files to scaffold. The template lives in Shared Context and the
   guides conform to it; promoting it would be a template nobody copies. (YAGNI;
   contrast Effort 3 Decision 4, where per-file pages justified a template file.)

5. **Engine recipe set (6):** add an action, implement a tag handler, implement an
   asset-effect (`S`-flag) handler, implement a goal handler, add a mutation type,
   add a hook at a new dispatch site. The first four instantiate one shared pattern —
   *a handler registry keyed by a data ID, plus a data-only silent-skip* — which the
   front door teaches once. The tag/effect/goal recipes **register a reactor into an
   already-dispatched hook category**; the hook recipe **adds the dispatch site
   itself** (a rarer, framework-level task carrying the recursion-bound and
   scope-lookup caveats). Splitting them on that frequency boundary keeps the common
   case short.

6. **Interface recipe set (3):** wire a prompt (overlay + adapter ask, end-to-end on
   the interface side), enable a mode slot, add a manage sub-view. The interface side
   has no single extension spine — overlay, adapter ask, view composition, and mode
   slot are genuinely distinct shapes — so each recipe is self-contained against its
   own page.

7. **The flagship "Add an action" is sliced per-side and cross-linked, not
   end-to-end.** The engine recipe covers the engine half (the `Action` contract,
   factory registration, the `Collector` method *signature*) and ends at the
   collector boundary with a hand-off to the interface guide's "Wire a prompt"
   recipe, which owns the adapter ask, the overlay, and the `ImplementedActions`
   gate. Each side's guide stays self-contained and reaches across the boundary only
   by link — mirroring the per-side arch-wiki split and the import boundary the code
   already enforces.

8. **Session slicing: two execution sessions, on Opus.** Session 1 writes
   `overview.md` + `engine.md` (the shared front door and the larger recipe set);
   Session 2 writes `interface.md`. Each session mines one side's arch pages and
   source, mirroring Effort 3's per-cluster clustering. One `wip:` commit per
   session. **Model override: Opus, not the table's Sonnet default for docs
   execution** — Robert's call: he does not trust Sonnet for documentation work.
   (Effort 3's pages were authored on Opus for the same reason; this brings Effort 4
   into line.)

## Open Questions — To Ratify at Execution Time

None routed forward. The four branch points this Plan session opened (recipe
coverage, the action seam, the setup home, session slicing) are all ratified above.
Recipe *prose* discovers-at-write per the docs-flow license; the inventory,
registration sites, and cross-links are fixed here.

## Effort 4 Session Summary

| # | Session | Deliverables | Depends on |
|---|---------|--------------|------------|
| 1 | Front door + engine guide | `contributing/overview.md` (setup/build/run + shared extension pattern); `contributing/engine.md` (6 engine recipes + engine test tooling + engine conventions) | Effort 2 (process settled), Effort 3 (arch pages exist) |
| 2 | Interface guide + closeout | `contributing/interface.md` (3 interface recipes + interface test tooling + interface conventions); flip the CLAUDE.md routing-table `(lands in D-002 Effort 4)` interim note to `contributing/overview.md` | Session 1 (the front door both guides link) |

Both sessions run on **Opus** (Robert does not trust Sonnet for documentation
work — Decision 8). One `wip:` commit per session.

```
Session 1 (overview + engine) ──► Session 2 (interface + closeout)
```

The dependency is light — `interface.md` links `overview.md`, so the front door
must exist first. The recipe *content* of the two side guides is independent.

## Per-Session Breakdown

### Session 1 — Front door + engine guide

**Scope:** The shared setup surface, the shared extension-pattern preamble both
guides lean on, and the six engine recipes.

**Deliverables:**

- `docs/contributing/overview.md` — front door, per the outline in Shared Context:
  what the two side guides are and when to read each; setup (`$GM_TOOLKIT_HOME`, the
  campaign data directory, Go prerequisites, getting rulebook TOML in place); build
  & run (build `./cmd/gm-toolkit`, open the TUI via `gm-toolkit faction`, scaffold a
  campaign with `gm-toolkit campaign create`); the **shared
  extension pattern** (the registry-keyed-by-data-ID + data-only-skip convention the
  engine recipes instantiate, taught once here); and an index linking `engine.md`,
  `interface.md`, and the architecture wiki front door
  (`../architecture/architecture-overview.md`).
- `docs/contributing/engine.md` — per the engine-guide outline: a one-paragraph
  orientation (headless engine; you extend it by registering handlers and emitting
  mutations) linking `engine/overview.md`; a **Conventions that bite** section
  (emit-never-mutate, data-only TOML silently skips, explicit deps to dodge the
  engine import cycle, recoverable sentinels live in the action package — each a
  one-liner linking the spine or the relevant page's Key Decisions); engine **test
  tooling** (the `testharness` golden-cycle harness, where fixtures live); and the
  **six engine recipes** (source maps below).

**Dependencies:** Effort 2 (the workflow the guides describe) and Effort 3 (the arch
pages every recipe links). Gates Session 2.

### Session 2 — Interface guide + closeout

**Scope:** The three interface recipes and the Effort-4 interim-note flip.

**Deliverables:**

- `docs/contributing/interface.md` — per the interface-guide outline: a
  one-paragraph orientation (the TUI is a projection; you extend it by adding
  prompts/views and never call the engine directly) linking `interface/overview.md`;
  a **Conventions that bite** section (the Update Discipline MAY/MAY-NOT, views hold
  display snapshots not engine pointers, message sub-packages break import cycles,
  lipgloss only inside `tui/` — each a one-liner linking the interface spine);
  interface **test tooling** (the smoke/dryrun harness; `Update`-driven component
  unit tests — no `teatest`); and the **three interface recipes** (source maps below).
- **Closeout:** flip the CLAUDE.md routing-table row "Setup, workflow, or dev
  conventions → `docs/contributing/` (lands in D-002 Effort 4)" to point at
  `docs/contributing/overview.md`, the interim note removed. (The branch-level
  interim sweep still runs at pre-merge after Effort 5; this flip is in scope here
  because the target now exists.)

**Dependencies:** Session 1 (the front door both side guides link).

## Per-Recipe Source Maps

Format per recipe: **Arch** (the page(s) the recipe links for shape/why) · **Touch**
(the files/registration sites the steps name — anchored to the arch pages, verified
in this Plan session, re-verified by execution against current source) · **Test**
(the harness and what to assert) · **Seam** (the sibling recipe across the boundary,
where the task crosses it). All engine paths are under `internal/faction/`.

### Engine recipes (Session 1)

**Add an action**
- Arch: [`engine/actions.md`](../../architecture/engine/actions.md)
- Touch: implement the four-method `Action` contract in
  `engine/action/actions/<name>.go`; register the factory in
  `RegisterDefaultActions` (`engine/action/actions/register.go`) with explicit deps
  (no `engine` import); add the `Collector` method *signature* in
  `engine/action/collector.go` if the action needs new GM input; add a recoverable
  sentinel in the action package if it can cancel; add an eligibility predicate in
  `actions/eligibility.go` if combat-shaped.
- Test: `testharness` golden cycle exercising the action; unit test on
  `Validate`/`Resolve`.
- Seam: the GM-facing prompt for the new `Collector` method, the overlay, and the
  `ImplementedActions` entry are the interface guide's
  **[Wire a prompt]** recipe.

**Implement a tag handler**
- Arch: [`engine/effect-mutation.md`](../../architecture/engine/effect-mutation.md)
  (tag registrar) + [`engine/hooks.md`](../../architecture/engine/hooks.md) (the hook
  category the handler registers into)
- Touch: write the Go handler in `engine/tag/tags/<name>.go` (implements
  `tag.Handler`: `TagID()` + `Apply(faction, hookRegistry)`); register it in
  `TagEngine.New` (`engine/tag/tag.go`) via `e.Register(...)`; the tag ID must
  already exist in `tags.toml` (data-only entries are inert until a handler exists);
  pick the hook categories the effect needs (Warlike/Fanatical/Scavengers/Preceptor
  are the four worked precedents).
- Test: `testharness` with a faction holding the tag; assert the registered hook
  fires at its category's dispatch beat.
- Seam: none — fully engine-side.

**Implement an asset-effect (`S`-flag) handler**
- Arch: [`engine/effect-mutation.md`](../../architecture/engine/effect-mutation.md)
  (effect registrar) + [`engine/hooks.md`](../../architecture/engine/hooks.md)
- Touch: write the handler in `engine/effect/effects/<feature>.go` (implements
  `effect.Handler`: `AssetDefinitionID()` + `Apply(faction, asset, hookRegistry)`;
  `transport.go` is the worked model); wire it into the registration loop in the
  engine composition root `NewWithRulebook` (`engine/core.go`), keyed on the asset
  definition's profile (the `def.Transport != nil` arm is the precedent —
  `effect.New` itself starts empty, unlike `tag.New`); the asset definition carries
  the `S`-flag/profile in rulebook TOML. (Zealots/Blockade Fleet/Integral Protocols/
  etc. are the data-awaiting candidates named on the arch page.)
- Test: `testharness`; assert the reactor emits its mutation on the trigger.
- Seam: none — fully engine-side.

**Implement a goal handler**
- Arch: [`engine/goals.md`](../../architecture/engine/goals.md)
- Touch: implement the `Handler` contract (`GoalID`/`CheckLock`/`UpdateProgress`) in
  `engine/goal/goals/<goal>.go`; register it in `GoalEngine.New`; choose the
  advancement beat(s) — `CheckLock` (time/condition-driven) vs `UpdateProgress`
  (event-driven, must run before `MutationEngine.Apply`); the goal ID must exist in
  `goals.toml`.
- Test: `testharness` multi-turn run; assert progress, lock behavior, and
  completion mutations.
- Seam: none — fully engine-side.

**Add a mutation type**
- Arch: [`engine/effect-mutation.md`](../../architecture/engine/effect-mutation.md)
  (the Mutation catalogue + the apply switch)
- Touch: define the struct and its `Type()` discriminator in `domain/mutation.go`;
  add an apply arm to the `MutationEngine.Apply` type switch (exhaustive-by-panic, so
  the missing arm fails loudly the first time it's emitted); emit it from the
  sub-engine that owns the change; add a row to the Mutation catalogue on the arch
  page.
- Test: apply-layer unit test; history serialization round-trip (the `Type()` string
  is the history key).
- Seam: none — fully engine-side.

**Add a hook at a new dispatch site**
- Arch: [`engine/hooks.md`](../../architecture/engine/hooks.md)
- Touch: if introducing a new category, define the typed hook interface in
  `engine/hooks/` (Cat 4-style typed family, for compile-time safety) — otherwise
  reuse an existing category; add the dispatch call at the new orchestrator or action
  site; respect the framework invariants (only Cat 3 recurses, bounded at depth 5;
  the global → faction → asset scope lookup order; reactors recurse on newly emitted
  mutations only).
- Test: `testharness` asserting the hook is consulted at the new beat.
- Seam: tag/effect handlers *register reactors into* categories this recipe
  *dispatches* — cross-link the two tag/effect recipes above.

### Interface recipes (Session 2)

**Wire a prompt** (overlay + adapter ask, end-to-end)
- Arch: [`interface/overlays.md`](../../architecture/interface/overlays.md) +
  [`interface/event-stream.md`](../../architecture/interface/event-stream.md)
- Touch: add an `AskKind` enum value (append-only, `tui/adapter/`); implement the
  `Collector` method on the adapter (`tui/adapter/action_collector.go` or
  `phase_collector.go`) using the `ask` helper to park the engine and carry an
  adapter-built display payload (views never hold engine pointers); implement the
  overlay — a generic archetype (`archetypes.go`: `Select`/`MultiSelect`/`Confirm`)
  or a bespoke `huh` form in `tui/views/turn/overlay/<name>.go`; add the `newOverlay`
  switch arm (append-only, placeholder default) in the execution view
  (`tui/views/turn/execution/execution.go` — `newOverlay` is a `Model` method there,
  not in the overlay package); choose the Esc keymap (`formhelp.go`:
  `formHelp`/`declineFormHelp`/`cancelFormHelp`); add the action name to
  `ImplementedActions` (`registry.go`) if this completes an action's drivability.
- Test: smoke/dryrun harness (`tui/adapter/smoke.go`, `tui/dryrun.go`); overlay unit
  test driving `Update` and asserting the emitted `OverlayDoneMsg.Answer` (no
  `teatest` — see Decision Record Session 2).
- Seam: the engine-side `Action` and `Collector` method are the engine guide's
  **[Add an action]** recipe.

**Enable a mode slot**
- Arch: [`interface/state-machine.md`](../../architecture/interface/state-machine.md)
  (+ [`interface/regions.md`](../../architecture/interface/regions.md) for the layout
  budget)
- Touch: add the `Mode` constant + a seeded `slot{mode, disabled}` in
  `tui/views/modebar/modebar.go` (`New` builds the slice; Spatial is enabled by
  flipping `disabled`); add the sub-model to the `subs map[modebar.Mode]tea.Model`
  that `NewModel` builds (`tui/model.go`); implement the capability predicates
  (`inputCapturer`/`modeLocker`) the mode needs. Sizing is automatic — `resizeSubs`
  forwards the content budget to every sub; the mode only touches
  `tui/layout/layout.go` if it wants the three-column area, via `layout.Compose`
  (`Region` is an intra-view column key, not a per-mode handle). The disabled Spatial
  slot is the worked precedent (registered absence: rendered, skipped on cycle,
  placeholder view).
- Test: root-model unit test driving a Tab `tea.KeyMsg`; assert key routing and the
  mode's lifecycle (and Tab-refusal if it reports `ModeLocked`).
- Seam: none — fully interface-side.

**Add a manage sub-view**
- Arch: [`interface/views.md`](../../architecture/interface/views.md)
- Touch: add the sub-view under `tui/views/manage/<name>/`; wire it into the manage
  router and its `backTarget` map (`tui/views/manage/manage.go`); add a message
  sub-package (`msgs/`) if it signals up; route state changes through the atomic-CRUD
  wizard pattern if it mutates.
- Test: manage-model unit test driving navigation and asserting the `backTarget`
  return path (drive `Update`; no `teatest`).
- Seam: none — fully interface-side.

## Shared Context

### Recipe template (the guides conform; not promoted to `templates/`)

```markdown
### <verb> a <thing>

> Shape & why: [<subsystem>](../architecture/<side>/<page>.md)

<One sentence: what extending this gets you, or when you reach for it.>

1. **`<file path>`** — <what to add / register / wire>.
2. **`<file path>`** — <…>.

**Test:** <harness + what to assert>.

**See also:** <sibling recipe across the engine/interface boundary, if the task
crosses it>.
```

### Front-door outline (`docs/contributing/overview.md`)

1. **What this is** — the two side guides and when to read each; how to read a
   recipe (the template anatomy).
2. **Setup** — `$GM_TOOLKIT_HOME`, the campaign data directory, Go prerequisites,
   placing rulebook TOML. (GMs own all rulebook data — no embedded defaults; link
   [`persistence`](../architecture/engine/persistence.md).)
3. **Build & run** — build `./cmd/gm-toolkit`, open the TUI via `gm-toolkit
   faction`, scaffold a campaign with `gm-toolkit campaign create`.
4. **The extension pattern** — the registry-keyed-by-data-ID + data-only-silent-skip
   convention the engine recipes share, taught once; links
   [`effect-mutation`](../architecture/engine/effect-mutation.md) and
   [`goals`](../architecture/engine/goals.md) as the worked instances.
5. **Index** — links to `engine.md`, `interface.md`, and
   [`architecture-overview.md`](../architecture/architecture-overview.md).

### Engine-guide outline (`docs/contributing/engine.md`)

1. **Orientation** — one paragraph; links [`engine/overview.md`](../architecture/engine/overview.md).
2. **Conventions that bite** — emit-never-mutate; data-only TOML silently skips (add
   the data first); explicit deps to dodge the engine import cycle; recoverable
   sentinels live in the action package. Each a one-liner + link, no re-narration.
3. **Test tooling** — the `testharness` golden-cycle harness; fixture location.
4. **Recipes** — the six, in the source-map order above.

### Interface-guide outline (`docs/contributing/interface.md`)

1. **Orientation** — one paragraph; links [`interface/overview.md`](../architecture/interface/overview.md).
2. **Conventions that bite** — the Update Discipline (MAY/MAY-NOT); views hold
   display snapshots, never engine pointers; message sub-packages break import
   cycles; lipgloss only inside `tui/`. Each a one-liner + link.
3. **Test tooling** — the smoke/dryrun harness; component unit tests that drive
   `Update` and assert the emitted message (no `teatest`).
4. **Recipes** — the three, in the source-map order above.

### Target tree

```
docs/contributing/
  overview.md     ← front door: setup, build, the shared extension pattern (Session 1)
  engine.md       ← 6 engine recipes + engine test tooling + conventions (Session 1)
  interface.md    ← 3 interface recipes + interface test tooling + conventions (Session 2)
```

## Decision Record — Execution

*(Append and reconcile: reversals of planned decisions are logged here and the plan
body fixed. Architecturally-durable entries graduate to arch-page `Key Decisions` at
pre-merge — for this effort that promotion is usually a no-op, since the deliverables
link the arch pages rather than restate them. Empty at plan time.)*

**Session 1 (2026-06-12) — source re-verification corrections.** The
re-verify-against-current-source step (per the plan header) caught four stale
references in the source maps / outlines, fixed in the plan body and written
correctly in the deliverables:

1. **Binary is `gm-toolkit`, not `faction-manager`.** The CLI root is
   `gm-toolkit` (Cobra); the faction TUI is the `gm-toolkit faction` subcommand.
   Campaign creation is `gm-toolkit campaign create <id> --path <dir> --rules
   <rulebook-dir>`. The outline's `faction-manager` name was carried from an older
   working title.
2. **Tag handlers live in `engine/tag/tags/<name>.go`**, not directly in
   `engine/tag/`; registration is `e.Register(tags.<Name>Handler{})` in
   `tag.New` (`engine/tag/tag.go`).
3. **Effect handlers wire in the engine composition root, not on the effect
   engine.** `NewWithRulebook` is `engine/core.go` (the `Engine` constructor), which
   builds an *empty* `effect.New(log)` and then registers handlers in a rulebook
   loop (`if def.Transport != nil { e.Effect.Register(...) }`). The arch page's
   "`NewWithRulebook` registers one handler per asset definition" reads as if the
   method were on `EffectsEngine`; the recipe points at `core.go`'s loop. (Not a
   doc error to fix on the arch page — its prose is accurate about *what* happens,
   just ambiguous about *where*; the recipe disambiguates.)
4. **The `testharness` is a full-cycle integration harness, not golden-file-based.**
   `NewHarness(t)` scaffolds a temp campaign, copies `rulebooks/swn/`, builds a live
   engine with `StubSpatialMap` + `ScriptedCollector` + `RecordingObserver`, and
   tests assert against recorded `history.jsonl` via `ReadHistory` / `AssertKinds` /
   `FindMutationByCause`. The guide describes it as such; "golden cycle" in the arch
   overview is loose phrasing, left as-is.

All other source-map paths (`RegisterDefaultActions`, `action/collector.go`,
`actions/eligibility.go`, `goal/goals/` + `GoalEngine.New`, `domain/mutation.go` +
`MutationEngine.Apply`, the `hooks/` typed families) verified clean.

**Session 2 (2026-06-12) — source re-verification corrections.** The
re-verify-against-current-source step caught three stale references in the
interface source maps / outlines, fixed in the plan body and written correctly in
`interface.md`:

1. **`newOverlay` is a method on the execution view, not the overlay package.** The
   `AskKind → overlay` dispatch switch is `Model.newOverlay` in
   `tui/views/turn/execution/execution.go`, not anything under `tui/views/turn/overlay/`.
   The "Wire a prompt" recipe points the switch-arm step at `execution.go`; the
   overlay *file* itself stays under `overlay/<name>.go`.
2. **No `Register` on modebar, and "assign it a Region key" misframes layout.** The
   `Mode` enum and the seeded `slot{mode, disabled}` slice live in
   `tui/views/modebar/modebar.go`; a mode is added/enabled by editing the slice
   `New()` builds (Spatial is enabled by flipping `disabled`). The sub-model registers
   in the `subs map[modebar.Mode]tea.Model` that `NewModel` builds (`tui/model.go`).
   `layout.Region` (`Left`/`Center`/`Right`, `tui/layout/layout.go`) is an
   **intra-view column key, not a per-mode handle** — a newly-registered sub is sized
   automatically by `resizeSubs` and only touches `layout` if it wants the
   three-column working area. The "Enable a mode slot" recipe's step 4 was rewritten
   accordingly. (The plan's source map said "assign it a Region key"; corrected to
   sizing-is-automatic + optional `layout.Compose`.)
3. **`teatest` is not a dependency.** It appears in neither `go.mod` nor `go.sum`;
   the TUI suite drives models by hand — construct the component, feed `tea.Msg`s
   through `Update`, run the returned `tea.Cmd`, assert the emitted message
   (`OverlayDoneMsg.Answer`; `views/turn/overlay/movement_test.go` is the pattern).
   Every interface recipe's **Test** line and the guide's test-tooling section
   describe that pattern; the plan's "teatest" phrasing is dropped.

## Out of Scope

- **Diagrams** — Effort 5; recipes may leave a `<!-- diagram: ... -->` breadcrumb
  where a flow diagram would help, nothing more.
- **Re-narrating the arch wiki** — recipes and conventions link the architecture
  pages for shape and "why"; restating a `Key Decisions` entry in a guide is a
  Decision 3 violation.
- **Recipes for subsystems with no clean extension story** — `domain` types, the
  Cobra `cmd/` layer, persistence/CRUD internals earn recipes only when a recurring
  extension task appears.
- **A `CONTRIBUTING.md` at repo root** — the guides live under `docs/contributing/`
  per the CLAUDE.md routing table; a root pointer file, if wanted, is a separate
  chore.
- **The branch-level interim-note sweep** — the full sweep runs once at pre-merge
  after Effort 5; Effort 4 flips only the one routing-table row whose target it
  creates.
- **Promoting the recipe template to `docs/process/templates/`** — Decision 4; the
  template governs sections within two guides, not per-recipe files.
