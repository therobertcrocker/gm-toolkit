# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `A.001` | Data-Driven Rulebook (H.A.W.K) | `arc` | In-Progress | Engine goes truly multi-rulebook: data-driven A/S registry both surfaces, agnostic goals/tags + shared `_core`, H.A.W.K authored; see arc concept |
| `A.001.6` | S-flag rule extensions | `feature` | In-Progress | The 14 SWN S-flag assets that need new state, a new prompt, or a new subsystem; registers through A.001.3's model; wide Discovery; see arc-plan |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `D.003` | Written session rules match practice | `docs` | before `A.001.6` | Hand-off sets execution scope (one commit default); model table → one line; drop prompt-to-switch; `CLAUDE.md` item 6 check-ins for design/scope only; prune redundant memories; `D.002` N1 |
| `D.004` | Plain concept header + prose budget | `docs` | before `A.001.6` | Required plain what/why/how-it-fits section (<15 lines) atop every plan and discovery doc; one plan-template line sets a per-effort size target; `D.002` N2 |
| `A.001.4` | Goals/tags neutralization + `_core` home | `feature` | ready | Shared `rulebooks/_core/` composed at scaffold + data-only flavor neutralization; see arc-plan |
| `A.001.5` | H.A.W.K asset authoring | `feature` | ready | Lore-seed session (E1, startable early) then Force→Cunning→Wealth passes; owns lore canon; see arc-plan |


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
| `F.023` | Region-scoped movement limits | ready | Generalize relocate's `range = "region"` beyond the instant primitive: let speed/drift movement abilities cap reach by region membership, not just hex distance. Relocate proved the enumerator pattern (`WorldsInRegion`); the drift subsystem (`MovementOrder`/`DriftCost`/movement phase) is the harder lift deferred here. |
| `F.024` | Collector form back-navigation | `post-tui-turn` | Cancelling a multi-step ability form (e.g. relocate destinations) spends the faction's whole action via `ErrTurnCanceled`, and confirming a completed-but-empty form surfaces a false `ErrNoSelection` failure; add a discoverable cancel/back that unwinds to the asset/ability picker (then `SelectAction`) instead of burning the turn. Needs a re-prompt loop the linear action phase lacks today. Deferred from C11 (a-flag-ability-registry pre-merge review). |
| `F.025` | Per-faction roll confirmation | `F.020` | `RollWithHooks` asks the roll's one collector to confirm every elective directive, so a forced reroll (Book of Secrets) is confirmed by the roller's driver; route each directive to its `FactionID`'s driver once factions get separate drivers. |
| `F.026` | Show why a target is refused | `F.018` | An attack-eligibility hook (`restrict_attack`) drops a defender from the target list or cancels a counter with no reason shown; surface the refusing asset so the GM sees why a target is missing. Found in `A.001.3` Effort 4 Commit 4 review. |
| `F.027` | TUI adapter tests | next interface initiative | `tui/adapter` (1,176 lines) is the only implementation of the engine seam and has no tests; cover collector and observer paths with the engine test harness; `D.002` N7 |

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
| `R.019` | Migrate `[transport]` onto the effects list | `A.001.6` | Transport keeps its own typed profile and registration branch (`core.go:58-61`) beside the `[[effects]]` model; fold it in once the S-flag model is fully built, so one parse and dispatch path serves every S effect. |
| `R.021` | Remove dead world-scoped defender filter | -- | `liveDefendersOnWorld` → `eligibleDefendersOnWorld` has no caller outside `spatial_index_test.go` and skips the attack-eligibility hooks; delete it (move its tests to `eligibleDefendersAtHex`) so no future caller bypasses `restrict_attack`. Found in `A.001.3` Effort 4 Commit 4 review. |
| `R.022` | Orchestrator error-handling consolidation | -- | Collapse the 26 recoverable-or-fatal blocks in `orchestrator.go`; only collector calls can return the three recoverable sentinels; target <500 lines, no behaviour change; `D.002` N3 |

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
| `B.009` | Capital Fleet missing `P` flag | -- | `SWN-F8-001` reads "permission is required to raise a Capital Fleet" but carries `flags = ["S"]`; Mercenaries `SWN-W3-003` states the same rule with `["S", "P"]`. Data error found in `A.001.3` discovery. |
| `B.010` | Reroll uses base die sides | first non-d10 Cat 1 die | `roll.go:86` rerolls via `roller.Roll(baseRoll.Sides)`, so a Cat 1 die of other sides rerolls wrong. Inert today (all d10); Book of Secrets is the first real reroll user. |
| `B.011` | TargetSelf contest gets the clone | first `TargetSelf` contest | `ability/dispatch.go:87` passes the `workingFaction` clone to `resolveTarget`, which returns it as `target.Faction` for `TargetSelf`; a self-contest's target-side roll would spend hook budgets on the clone. Latent: no asset authors it. Found in `A.001.3` Effort 3, Session 2. |
| `B.012` | Old budget keys in saves | -- | `A.001.3` added a trigger segment to effect budget keys (`e047af7`); a mid-cycle save from before holds old keys, so one extra steal/absorb is possible that turn. Self-heals at the owner's next bookkeeping. |

### Docs

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `D.001` | Split pre-merge review from execution | next process session (Fable) | Review session writes triaged findings to the plan (design calls made there); execution session does fixes, then promotion, planned-work and the docs move; skip the split when no finding needs a design call; closing step writes the hand-off as the plan's last Decision Record entry (`D.002` N4) |
| `D.005` | Whole-tree review at arc boundaries | -- | One line in the arc flow: a whole-codebase review runs when an arc's serial spine completes; diff-scoped reviews missed `RunCycle` for 148 days; `D.002` N5 |
| `D.006` | Fold arc concept into arc plan; stop moving discovery docs | -- | Arc concept becomes the arc plan's first section; discovery docs stay put after merge (plans still move); templates 5 → 4; `D.002` N6 |
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

## A.001 — Data-Driven Rulebook (H.A.W.K)

**Arc** (rerouted from `F.022` feature — the ability-catalog pre-plan session settled `R.004` and tipped this from one feature into an arc). The full write-up lives in the arc artifacts: see the [arc concept](../../initiatives/arcs/data-driven-rulebook/data-driven-rulebook-arc-concept.md) for vision, problem, and why-an-arc; Arc-Discovery ratifies the cross-cutting decisions and Arc-Plan slices the constituent initiatives, which get their own rows and write-ups here as promoted. All arc artifacts live in `docs/initiatives/arcs/data-driven-rulebook/`.

**Provenance** — promoted directly from post-E2E direction-setting (2026-06-15); no prior Backlog entry. Subsumes the campaign-content motivation behind `F.010`/`F.014`; is the evidence source for `R.004` (which the catalog session has since settled — build the full data-driven registry, both surfaces).

**Status** — In-Progress — Arc-Discovery and Arc-Plan complete; sliced into `A.001.1`–`A.001.5` (see the [arc-plan](../../initiatives/arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md)). The former `R.018` refactor row is subsumed by `A.001.1`. **`A.001.1` complete** (effect-keyed dispatch + SWN re-prefix, on `refactor/effect-keyed-a-dispatch`). **`A.001.2` complete** (A-flag composition pipeline + instant-relocate primitive + action-spent harness, on `feature/a-flag-ability-registry`). **`A.001.3` complete** (data-driven S-flag `[[effects]]` model + 19 SWN assets + free relocate harness, on `feature/s-flag-effect-registry`); the serial spine is done. `A.001.4` and `A.001.5` are ready; `A.001.6` (S-flag rule extensions) is unblocked.

<br />

## A.001.6 — S-flag Rule Extensions

**Feature**, constituent of arc `A.001`. Added to the arc by the `A.001.3` Discovery (AP-7), which split the S side in two. Full boundary in the [arc-plan](../../initiatives/arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md), section "A.001.6 — S-flag rule extensions".

**Problem** — `A.001.3` built the S-side model (*trigger → filters → gates → effect*) and shipped every effect that is a hook or a rule query: 19 assets. 14 SWN S-flag assets remain inert. Each one can register through the model, but its effect acts on something the engine does not have: game state, a prompt, or a subsystem.

**Approach** — Six mechanisms, unequal in size. Each is a separate engine design:

- **Change of asset owner** (small): Hostile Takeover, Treachery.
- **State held on an asset** — a restriction with a lifetime, or unusable-until-repaired (medium): Saboteurs, Marketers, Seditionists, Boltholes.
- **A prompt to a faction that is not acting** — Cat 3 receives no collector (medium): False Front, Medical Center, Monopoly.
- **A purchase that targets another asset** (medium): Stealth.
- **A permission model for movement** (large): Lobbyists, Popular Movement, Transport Lockdown. It also takes the transport bypass clauses `A.001.3` left alone (Deep Strike Landers, Blockade Runners).
- **Damage redirected into combat** (medium): Cracked Comms.

Marketers, Seditionists and Monopoly were rerouted from `A.001.2` (AP-6) and passed on by `A.001.3`. The arc-plan expects a full, wide Discovery and a multi-effort plan.

**Related items** — `R.019` (migrate `[transport]` onto the effects list) triggers on this initiative. `B.009` (Capital Fleet missing `P` flag) belongs to the permission work. The `A.001.3` plan defers one model question to the first asset that needs it: `filters` answers both "which events fire this" and "which assets the effect acts on".

**Unlocks** — The complete SWN S side: every S-flag asset runs as data. A larger primitive set for the `A.001.5` mechanical sort, if this initiative runs first. Gates nothing.

**Trigger** — `A.001.3` complete (met).

**Status** — In-Progress — promoted 2026-10-17. Discovery not started.


