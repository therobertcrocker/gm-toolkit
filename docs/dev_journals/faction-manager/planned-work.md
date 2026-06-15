# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `F.022` | H.A.W.K Rulebook | `feature` | In-Progress | Author H.A.W.K's TOML from SWN as reference; handle rows 1–2 drift, collect row-3 evidence for `R.004` |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| *(none)* | | | | |


---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Planned Initiatives when fully scoped for discovery.

### Features

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `F.010` | Remaining Tags | -- | Implement remaining tag functionality |
| `F.013` | Goal State Predicates | AI/planner | Add `Satisfied(state) bool` alongside per-goal `progressX`; both shapes coexist |
| `F.014` | [A, S]-flag Abilities | `R.001` | Nine A-flag abilities stubbed. Unknown Number of S-flag abilities. |
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
| `R.018` | Ability dispatch keyed by effect | `F.022` | Re-key `ability.Dispatch` on `def.Ability.Effect` not `def.ID`; rename `informers` → `revealStealth`; abilities become mechanic-keyed like tags/goals — R.004-increment-1 |

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

## F.022 — H.A.W.K Rulebook

**Problem** — The engine has only ever been exercised against SWN reference data: one rulebook's worth of goals, tags, and assets, with behavior hardcoded by ID and validated against the source material it was modeled on. There is no real campaign content. Robert's actual campaign — H.A.W.K, a standalone rulebook — exists only as intent, so the tool can't yet run the game it exists to serve, and the standing question (does hardcoded-by-ID behavior hold past a single rulebook?) stays speculative.

**Approach** — Author H.A.W.K's TOML (goals, tags, assets) in the existing rulebook format, using SWN as the reference scaffold and reflavoring as the picture of the assets sharpens — the reflavoring itself is expected to surface drift. Each element sorts into one of three rows as it is authored:

- **Row 1** — same mechanic, reused mechanism — pure data (`name`/`description`), no code — *in scope*
- **Row 2** — a genuinely new mechanic — one handler via the existing data-only-skip pattern (`tag.go:42`, `goal.go:55`) — *in scope*
- **Row 3** — drift pervasive enough to warrant data-driven handlers — **not built here**; collect the evidence (count and shape of misfits) that would trigger `R.004` as its own effort

Convention: **IDs track mechanics** — reuse an ID when reusing a mechanic, mint a new one for a new mechanic. Keeps `B.002` out of scope so long as shared mechanics retain their SWN IDs. *Open for Discovery:* H.A.W.K is a standalone rulebook — whether it mints its own ID namespace even for shared mechanics (which pulls `B.002` back in) is unresolved.

**Ability dispatch is the exception (`R.018`).** Unlike tags and goals, `ability.Dispatch` (`dispatch.go:47`) keys on asset ID, not mechanic — so a reused ability effect would otherwise need a per-asset map entry. `R.018` re-keys it on `def.Ability.Effect` (R.004-increment-1), making abilities mechanic-keyed like the rest; F.022's Discovery sequences it first.

**Flavor is co-authored, not derived.** The mechanical sorting above is only half the initiative. The reflavoring — the names, descriptions, and fiction of every goal, tag, and asset — is a first-class collaborative pass, not a substitution run off SWN: Robert is interviewed about the campaign's flavor and the rulebook is authored together. The flavor is also an **output**: the campaign's canon (setting, factions, the fiction behind each element) is saved as a standing reference doc — durable, not an archived planning artifact — so future efforts can draw on it. *Open for Discovery:* the interview cadence (a dedicated phase up front, threaded through implementation, or both), and where the canon doc lives and how it is structured.

**Rulebook convention to canonicalize (`Discovery`).** A *rulebook* is an on-disk directory, `rulebooks/<name>/` — F.022 creates the second one (`rulebooks/hawk/`). The layout is documented today only in `docs/contributing/overview.md` and frozen decision 267, not in the canonical static-data page (`persistence.md`), and stale `rulebooks/swn/factions/` references survive in completed docs. Discovery states it crisply in `persistence.md` before `rulebooks/hawk/` is created.

**Unlocks** —

- The tool can run the actual H.A.W.K campaign — the project's reason to exist
- An evidence-grounded call on `R.004` — the drift tally shows whether hardcoded-by-ID is comfortable or painful, rather than guessing
- First proof the engine holds a *non-reference* rulebook — directly relevant to Worlds Without Number, a near-identical faction system with its own drift, the next beneficiary of `R.004` should the evidence call for it

**Trigger** — Promoted directly from post-E2E direction-setting (2026-06-15); no prior Backlog entry. Subsumes the campaign-content motivation behind `F.010`/`F.014`; is the evidence source for `R.004`.

**Status** — In-Progress — next session is Discovery (`docs/initiatives/discovery/hawk-rulebook.md`, to be written).

<br />

