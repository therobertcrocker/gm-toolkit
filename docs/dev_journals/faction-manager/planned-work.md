# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

> **Scheduling note (2026.05-23):** The TUI rebuild (`F.005`) is the first of a planned arc to bring the TUI to feature completeness. No new engine features will be promoted from Backlog to Up Next until that arc is complete. Bugfixes and refactors are unaffected.
>
> **Addendum (2026.05-25):** `F.012` (spatial-map-cli) promoted as content-authoring tooling, exempt from the engine-feature freeze. It is a prerequisite for `F.005.2` (Manage needs an authored spatial map to render).
>
> **Addendum (2026.05-26):** `F.012` shipped (`feature/spatial-map-cli` merged). Prerequisite for `F.005.2` is met.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| *(none)* | | | | |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|


---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Planned Initiatives when fully scoped for discovery.

### Features

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `F.004` | CLI Rebuild | -- | -- |
| `F.010` | Tag-Granted Assets | -- | -- |
| `F.011` | Tag Reminders | -- | -- |
| `F.013` | Goal State Predicates | AI/planner | Add `Satisfied(state) bool` alongside per-goal `progressX`; both shapes coexist |
| `F.014` | A-flag Abilities | `R.001` | Nine A-flag abilities stubbed |
| `F.015` | TUI Manage: Map View | -- | Add a map view to the manage screen, showing the faction's homeworld and the planets it controls. |
| `F.016` | TUI Manage: Right Panel | -- | Add a right-hand panel to the manage screen, showing additional faction details at a glance. |
| `F.017` | Mid-cycle pause/cancel | -- | A running cycle currently locks the mode bar (Tab is a no-op) and can't be paused or aborted; add safe pause/resume + cancel so the GM can step out and back without wedging the engine. |
| `F.018` | UX Pass: Confirm-on-Choice | post-tui-turn | Binding cadence principle for the TUI UX pass: every consequential choice gets a confirm/pause. First instance ships in the action-result-panel work. |
| `F.019` | Edit Mode | -- | Freeform state manipulation outside turn rules; campaign setup, world-building, corrections; no mechanical validation. |
| `F.020` | AI Decision-Making | -- | Goal-oriented and in-character modes using the source-agnostic action interface; traditional game AI (not LLM). |
| `F.021` | Narrative Recap Surface | -- | `digest` builder + wire `Renderer` are built and unit-tested but unwired; no command or view renders a cycle recap. Add a frontend (CLI and/or TUI) that runs LoadCycle → Build → Render. |

### Refactors

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.003` | Action Preconditions | AI/planner | Lift `Action.Validate` into `Action.Preconditions []Precondition` |
| `R.004` | Data-Driven Handlers | High Pain | Replace hardcoded goal/ability handler dispatch with TOML-defined handlers + typed primitive registry |
| `R.005` | Adjacency vs. Warp Cost | Deferred from `F.012` | Split `Region.Boundary` into adjacency-edges (cost 1) and warp-edges (cost `crossingCost`) |
| `R.006` | Shared `styles.Dim` constant | -- | `#6C7086` duplicated across detail/list/manage/layout; extract one shared constant. |
| `R.007` | Extract `rebuildList()` helper | -- | Created/Deleted branches in manage.go copy-paste the rebuild sequence; extract a shared helper. |
| `R.008` | `world.Location` + `campaigns.ValidateID` | -- | `synthesize`/`slugify` reinvent what helpers already do |
| `R.009` | Remove write-only `detail.Model.height` | -- | Field is assigned but never read; dead weight in the detail model. |
| `R.010` | Unify `WindowSizeMsg` routing in manage.go | -- | Per-view asymmetry in how size messages are forwarded; normalise to one pattern. |
| `R.011` | Cache per-frame allocations in `compose()` / root `View()` | -- | Low priority: repeated allocations on hot path; pre-allocate or cache where safe. |
| `R.012` | Engine construction shape | -- | Review constructors so every user (TUI, CLI, tests) gets a fully-wired engine or fails fast at launch. |
| `R.013` | Candidate derivation boundary | next action change | Collector methods should always receive engine-derived candidates; retires the adapter's `derive*`/`statRating` ports of engine eligibility logic (revisits tui-turn Decision 3). |
| `R.014` | AskKind dispatch table | next new action | Collapse the parallel `newOverlay` + `askPhaseLabel` switches into one `map[AskKind]{label, factory}` that fails loudly on a missing entry. |
| `R.015` | Collector reply rigor | -- | `action_collector` swallows bad reply types (`raw.(T)` zero-values); match `phase_collector`'s loud type errors, add the `SelectMovementDecisions` nil guard, drop the `phaseCollector.ask` delegation wrapper. |
| `R.016` | Per-ask detail cards | 4th specialized card | Move per-ask knowledge out of `execution.detailView`'s `activeAsk` switch into an optional overlay interface (`DetailView() string`). |
| `R.017` | Evaluate `teatest` for UI tests | -- | TUI tests currently drive `Update` by hand and assert emitted messages; explore `charmbracelet/x/exp/teatest` as a golden-output harness for full model behavior — adopt across overlays/views or document why not. |

### Bugfixes

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.001` | Hardcoded IDs | -- | `stealth_applicator` (asset) and `"G.012"` (ChangeHomeworld goal) are hardcoded; silently break if TOML IDs change |
| `B.002` | Goal/Tag Display IDs | -- | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only. |
| `B.003` | `SeizePlanet` History | -- | Currently invisible in history. `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| `B.004` | `Ability` Narration | -- | Add fallback text for asset ability narration |
| `B.005` | Goal XP | -- | Retype `Difficulty` from `string` to int or tagged sum so engine can dispatch XP |
| `B.006` | Turn event stream flash | -- | Center-pane event stream visibly redraws/flickers between collector prompts during a cycle; smooth the transition so it doesn't flash on each phase. |
| `B.007` | Self-warp validation gap | -- | `newRegionMap` accepts a hand-edited warp with `from_region == to_region`; harmless dead edge in pathfinding, but the validator should reject it. |
| `B.008` | Stale `ApplyBookkeeping` comment | -- | Doc comment claims it "advances the turn phase to `PhaseAction`", but `TurnState` has no phase field; phase order lives in the orchestrator. One-line comment fix. |
---
<br />

## Open Questions

Unresolved design questions that feed into a backlog item. When the item ships, the question is answered.

| # | Question | Relevant Feature | Backlog Item |
|---|----------|-----------------|--------------|
| 1 | Starting Coin for new factions — no explicit SWN rule; what is the right default (Wealth rating, fixed amount, GM prompt)? | Edit Mode | `F.019` |
| 2 | How does Edit Mode integrate into the Cobra command tree — sub-commands under a top-level `edit` command, or per-entity sub-commands (e.g. `faction edit`)? | Edit Mode | `F.019` |
| 3 | When AI decision-making is added, how does the GM designate which factions are AI-driven vs. manually controlled? | AI Decision-Making | `F.020` |
---
<br />
<br />

# Initiative Write-Ups
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

