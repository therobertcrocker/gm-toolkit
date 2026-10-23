# Planned Work

Pre-discovery initiative tracker. Deferred items move to Planned Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `A.001` | Data-Driven Rulebook (H.A.W.K) | `arc` | In-Progress | Engine goes truly multi-rulebook: data-driven A/S registry both surfaces, agnostic goals/tags + shared `_core`, H.A.W.K authored; see arc concept |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
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
| `F.028` | In-transit assets can't be acted on | -- | An asset between worlds (no `WorldID`, cargo included) should be untouchable by actions; Refit already refuses it (`A.001.6` Effort 4 Decision 11); audit repair, sale, ability targets and the rest, then block them. |
| `F.029` | Digest line for a paid lift | -- | A Marketers pay-to-lift (`turn.OfferConditionLifts`) reads only as Coin spent in bookkeeping; give the lift its own narrative line. Deferred from `A.001.6` Effort 3. |
| `F.030` | Show who a Lockdown holds out | -- | No view shows the GM which factions a Transport Lockdown holds out of a world (`locked_out` sits on the Lockdown asset); surface it. Deferred from `A.001.6` Effort 5. |
| `F.031` | Loyalty roll for captured cargo | -- | House rule: cargo aboard a transport that changes sides rides on under its old owner (`A.001.6` Effort 1 Decision 8); a loyalty roll could decide whether it defects. Deferred from `A.001.6` Effort 1. |

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
| `R.013` | Candidate derivation boundary | next action change | Collector methods should always receive engine-derived candidates; retires the adapter's `derive*`/`statRating` ports of engine eligibility logic (revisits tui-turn Decision 3). Investigate drift: `deriveContestedWorlds` copies `seizePlanetTargetWorlds` and needed the same fix twice (`A.001.6` Effort 4 review, DR-3). |
| `R.015` | Collector reply rigor | -- | `action_collector` swallows bad reply types (`raw.(T)` zero-values); match `phase_collector`'s loud type errors, add the `SelectMovementDecisions` nil guard, drop the `phaseCollector.ask` delegation wrapper. |
| `R.016` | Per-ask detail cards | 4th specialized card | Move per-ask knowledge out of `execution.detailView`'s `activeAsk` switch into an optional overlay interface (`DetailView() string`). |
| `R.017` | Evaluate `teatest` for UI tests | -- | TUI tests currently drive `Update` by hand and assert emitted messages; explore `charmbracelet/x/exp/teatest` as a golden-output harness for full model behavior — adopt across overlays/views or document why not. |
| `R.019` | Migrate `[transport]` onto the effects list | ready | Transport keeps its own typed profile and registration branch (`core.go:58-61`) beside the `[[effects]]` model; fold it in once the S-flag model is fully built, so one parse and dispatch path serves every S effect. |
| `R.021` | Remove dead world-scoped defender filter | -- | `liveDefendersOnWorld` → `eligibleDefendersOnWorld` has no caller outside `spatial_index_test.go` and skips the attack-eligibility hooks; delete it (move its tests to `eligibleDefendersAtHex`) so no future caller bypasses `restrict_attack`. Found in `A.001.3` Effort 4 Commit 4 review. |
| `R.022` | Orchestrator error-handling consolidation | -- | Collapse the 26 recoverable-or-fatal blocks in `orchestrator.go`; only collector calls can return the three recoverable sentinels; target <500 lines, no behaviour change; `D.002` N3 |
| `R.023` | Effect contests through roll hooks | When reactors and phase effects can reach a collector | Tripwire Cells, Panopticon Matrix and Lobbyists roll bare 1d10 + stat (`contestPasses`), so Fanatical's and Book of Secrets' rerolls skip them; move all three onto `RollWithHooks` together. `A.001.6` Effort 5 Decision 10. |
| `R.024` | Clear inert `no_attack` records | if saves grow noisy | A Seditionists' `no_attack` stays on its target after the Seditionists die; `domain.AttackBlocked` skips it when read, so it only clutters saves. Clear it when its source leaves play. Deferred from `A.001.6` Effort 3 Decision 10. |

### Bugfixes

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.001` | Hardcoded goal ID | -- | `"G-012"` (Change Homeworld goal) is hardcoded in `change_homeworld.go`; silently breaks if the TOML ID changes. The asset half (Stealth) closed in `A.001.6`. |
| `B.002` | Goal/Tag Display IDs | -- | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only. |
| `B.003` | `SeizePlanet` History | -- | Currently invisible in history. `Output()` returns no mutations; add `GoalPhaseAdvanced` mutation |
| `B.004` | `Ability` Narration | -- | Add fallback text for asset ability narration |
| `B.005` | Goal XP | -- | Retype `Difficulty` from `string` to int or tagged sum so engine can dispatch XP |
| `B.006` | Turn event stream flash | -- | Center-pane event stream visibly redraws/flickers between collector prompts during a cycle; smooth the transition so it doesn't flash on each phase. |
| `B.007` | Self-warp validation gap | -- | `newRegionMap` accepts a hand-edited warp with `from_region == to_region`; harmless dead edge in pathfinding, but the validator should reject it. |
| `B.008` | Stale `ApplyBookkeeping` comment | -- | Doc comment claims it "advances the turn phase to `PhaseAction`", but `TurnState` has no phase field; phase order lives in the orchestrator. One-line comment fix. |
| `B.010` | Reroll uses base die sides | first non-d10 Cat 1 die | `roll.go:86` rerolls via `roller.Roll(baseRoll.Sides)`, so a Cat 1 die of other sides rerolls wrong. Inert today (all d10); Book of Secrets is the first real reroll user. |
| `B.012` | Old budget keys in saves | -- | `A.001.3` added a trigger segment to effect budget keys (`e047af7`); a mid-cycle save from before holds old keys, so one extra steal/absorb is possible that turn. Self-heals at the owner's next bookkeeping. |
| `B.013` | Peaceable Kingdom misses a missed attack | -- | `peaceable_kingdom.go:20-36` detects an Attack from damage or attacker stealth-loss mutations; a miss by an unstealthed attacker emits neither, so the goal ("do not take an Attack action") does not reset. Found in `A.001.6` Effort 1, Commit 6. |
| `B.014` | Capture misses a sibling-removed target | first Treachery hit on a self-damage defender | `capture_target.go` cannot see a sibling reactor's output in its round, so a target that removes itself (Zealots' `self_damage`) still gets `AssetOwnerChanged`; the owner change then misses and the turn aborts. Fence-signed; found in `A.001.6` Effort 1 interim review. |
| `B.015` | Bought assets get no hex | -- | `NewPurchasedAsset` (`asset.go:305`) sets `WorldID` only, so bought assets on different worlds share the zero `RegionHex` and Attack (`eligibleDefendersAtHex`) lets them fight; harness `addActiveAsset`/`AddFaction` copy the gap. Unconfirmed in the app; found in `A.001.6` Effort 3, Commit 4. |
| `B.016` | Hook budgets survive an aborted action | -- | `dispatch/lethal.go` and `dispatch/roll.go` spend once-per budgets on live state during Resolve; a prompt canceled later in the action drops its mutations but keeps the spend (a lost Medical Center restore or reroll); snapshot `HookBudgets` before `Run`, restore on error. Found in `A.001.6` Effort 2 interim review. |
| `B.017` | Repair price ignores Medical Center | -- | `overlay/repairorders.go` shows `n(n+1)/2` without the `RepairCostModifier` discount, and `RepairAsset.Validate` still needs 1 Coin when a discounted repair is free; the engine charges the right amount. Deferred from `A.001.6` Effort 2. |
| `B.018` | Digest drops the defender name | -- | Cross events name an asset only when it is destroyed (`digest/build.go`), so `{defender_asset}` renders empty on a non-lethal hit, a shelter or a save. Found in `A.001.6` Effort 1. |
| `B.019` | Movement orders read as unknown in the digest | -- | `digest/build.go` has a case for `movement_order_held` only; the five other `movement_order_*` types each become an `unknown mutation` note. Found in `A.001.6` Effort 5 planning. |
| `B.020` | T-011 pick with every world governed | -- | The faction wizard's world select for a T-011 pick lists only worlds no faction governs; when none is left, the select has no options. Found in `A.001.6` Effort 4. |
| `B.021` | SWN asset data drift | -- | A diff of all 72 assets against the book: transports flagged `S` where the book says `A`; flag gaps on Mercenaries, Surveyors, Tripwire Cells, Strike Fleet, Space Marines, Pretech Manufactory; wrong cost/TL on Pretech Researchers, TL on Surveyors. Audit and fix. Found in `A.001.6` discovery. |
| `B.022` | Enabler surcharge reads the turn-start index | first enabler after an in-action removal | `purchase_enabler.go` prices through `world.AssetsOnWorld`, the turn-start index, not the action's view; a rival surcharge asset that an earlier asset in the same action removed (a Monopoly levy) still charges. Found at `A.001.6` pre-merge. |
| `B.023` | Transport range never enforced | -- | `TransportProfile.MaxHex` loads but nothing reads it: an order goes to any world. The movement-redesign Effort 2 plan (Commit 1 Task 6) planned the check; the orchestrator's cargo pick checks only `MaxCargo` (`orchestrator.go:779`). Rules call open: what `max_hex` means under multi-turn movement. Found in the sandbox build. |

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

**Status** — In-Progress — Arc-Discovery and Arc-Plan complete; sliced into `A.001.1`–`A.001.5` (see the [arc-plan](../../initiatives/arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md)). The former `R.018` refactor row is subsumed by `A.001.1`. **`A.001.1` complete** (effect-keyed dispatch + SWN re-prefix, on `refactor/effect-keyed-a-dispatch`). **`A.001.2` complete** (A-flag composition pipeline + instant-relocate primitive + action-spent harness, on `feature/a-flag-ability-registry`). **`A.001.3` complete** (data-driven S-flag `[[effects]]` model + 19 SWN assets + free relocate harness, on `feature/s-flag-effect-registry`); the serial spine is done. **`A.001.6` complete** (the 14 remaining SWN S-flag assets: asset conditions, change of owner, special attacks, Cat 6 lethal interception, purchase qualities, the world government record, and permission at purchase and arrival, on `feature/s-flag-rule-extensions`; [plan](../../initiatives/implementation/completed/s-flag-rule-extensions-plan.md)); every SWN S-flag asset now runs as data. `A.001.4` and `A.001.5` are ready.


