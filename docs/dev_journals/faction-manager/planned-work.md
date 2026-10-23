# Planned Work

Pre-discovery initiative tracker. Deferred items move to Current Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins.

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

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Current Initiatives when fully scoped for discovery.

**Trigger values:**

- `--` — no gate; pick up at any time
- `ready` — scoped; can be promoted now
- `after <ID>` — waits for another row to ship
- `when <condition>` — waits for an event in the code or the project
- `needs triage` — the cause is not known yet; diagnose before scoping

**Layers.** Refactors and Bugfixes are grouped by the layer where the fix lands (the lowest layer that must change). An empty layer shows `_None._`.

1. **Data** — campaign files on disk, loaders (`rulebook`, `spatial`, `state`, `campaigns`), domain types (`faction/domain`)
2. **Engine** — `faction/engine/…` and the seam it defines (collectors, `TurnObserver`)
3. **Driver** — `tui/adapter`, `engine/testharness`, the `--dryrun` driver, the Cobra entry (`cmd/gm-toolkit`)
4. **Presentation** — prompt labels, feed lines, the right-hand card, `narrative/digest`
5. **Rendering** — views, overlays, styles

### Arcs

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `A.002` | Edit Mode | -- | Freeform state manipulation outside turn rules; campaign setup, world-building, corrections; no mechanical validation. Spun out of the TUI rebuild (arc-plan Decision 4). Was `F.019`. |
| `A.003` | AI Decision-Making | -- | Goal-oriented and in-character modes using the source-agnostic action interface; traditional game AI (not LLM). Was `F.020`. |

### Features

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `F.010` | Remaining Tags | -- | Implement remaining tag functionality |
| `F.013` | Goal State Predicates | when `A.003` starts | Add `Satisfied(state) bool` alongside per-goal `progressX`; both shapes coexist |
| `F.015` | TUI Manage: Map View | -- | Add a map view to the manage screen, showing the faction's homeworld and the planets it controls. |
| `F.016` | TUI Manage: Right Panel | -- | Add a right-hand panel to the manage screen, showing additional faction details at a glance. |
| `F.017` | Mid-cycle pause/cancel | -- | A running cycle currently locks the mode bar (Tab is a no-op) and can't be paused or aborted; add safe pause/resume + cancel so the GM can step out and back without wedging the engine. |
| `F.018` | UX Pass: Confirm-on-Choice | -- | Binding cadence principle for the TUI UX pass: every consequential choice gets a confirm/pause. First instance ships in the action-result-panel work. |
| `F.021` | Narrative Recap Surface | -- | `digest` builder + wire `Renderer` are built and unit-tested but unwired; no command or view renders a cycle recap. Add a frontend (CLI and/or TUI) that runs LoadCycle → Build → Render. |
| `F.023` | Region-scoped movement limits | ready | Generalize relocate's `range = "region"` beyond the instant primitive: let speed/drift movement abilities cap reach by region membership, not just hex distance. Relocate proved the enumerator pattern (`WorldsInRegion`); the drift subsystem (`MovementOrder`/`DriftCost`/movement phase) is the harder lift deferred here. |
| `F.024` | Collector form back-navigation | -- | Cancelling a multi-step ability form (e.g. relocate destinations) spends the faction's whole action via `ErrTurnCanceled`, and confirming a completed-but-empty form surfaces a false `ErrNoSelection` failure; add a discoverable cancel/back that unwinds to the asset/ability picker (then `SelectAction`) instead of burning the turn. Needs a re-prompt loop the linear action phase lacks today. Deferred from C11 (a-flag-ability-registry pre-merge review). |
| `F.025` | Per-faction roll confirmation | after `A.003` | `RollWithHooks` asks the roll's one collector to confirm every elective directive, so a forced reroll (Book of Secrets) is confirmed by the roller's driver; route each directive to its `FactionID`'s driver once factions get separate drivers. |
| `F.026` | Show why a target is refused | after `F.018` | An attack-eligibility hook (`restrict_attack`) drops a defender from the target list or cancels a counter with no reason shown; surface the refusing asset so the GM sees why a target is missing. Found in `A.001.3` Effort 4 Commit 4 review. |
| `F.027` | TUI adapter tests | when next interface initiative | `tui/adapter` (1,176 lines) is the only implementation of the engine seam and has no tests; cover collector and observer paths with the engine test harness; `D.002` N7 |
| `F.028` | In-transit assets can't be acted on | -- | An asset between worlds (no `WorldID`, cargo included) should be untouchable by actions; Refit already refuses it (`A.001.6` Effort 4 Decision 11); audit repair, sale, ability targets and the rest, then block them. |
| `F.029` | Digest line for a paid lift | -- | A Marketers pay-to-lift (`turn.OfferConditionLifts`) reads only as Coin spent in bookkeeping; give the lift its own narrative line. Deferred from `A.001.6` Effort 3. |
| `F.030` | Show who a Lockdown holds out | -- | No view shows the GM which factions a Transport Lockdown holds out of a world (`locked_out` sits on the Lockdown asset); surface it. Deferred from `A.001.6` Effort 5. |
| `F.031` | Loyalty roll for captured cargo | -- | House rule: cargo aboard a transport that changes sides rides on under its old owner (`A.001.6` Effort 1 Decision 8); a loyalty roll could decide whether it defects. Deferred from `A.001.6` Effort 1. |
| `F.032` | Reroll prompt shows the roll | -- | `NewConfirmReroll` gets only the `RerollDirective`, so the GM cannot see the dice; pass the `RollResult` and `RollContext` through `ConfirmReroll` and show dice + total (mark the die to reroll), roller + phase + stat, "Asked of", and on a defense roll the attack total to beat. Found in the sandbox combat test. |
| `F.033` | Show goal start and phase changes | -- | `GoalInitiated` and `GoalPhaseAdvanced` are folded on purpose in the execution feed (`format.go:179`) and the digest (`build.go:582-589`), so the GM never sees Seize Planet start or reach phase 2; surface both. Found in the sandbox combat test. |
| `F.034` | Attacker labels name the world | -- | `attackerLabel` (`overlay/attack.go:71`) shows name, stat and damage only, so two Treacheries on different worlds read the same; add the world name. This does not fix `B.029`: same-world rows still match. Found in the sandbox combat test. |
| `F.035` | Drop `(surcharged)` noise | -- | `overlay/purchaseenabler.go:19` tags every row `(surcharged)`, which says nothing when all rows carry it; say it once in the header and show the base cost next to the surcharged one. Found in manual test one, screenshot 06. |
| `F.036` | Focus shown by more than color | -- | `styles.FormTheme()` (`styles/styles.go:47`) keeps huh's Charm confirm buttons, where only the fill color marks the focused choice; add a text cue (`> Keep`, brackets or underline) for monochrome and colorblind use. Found in manual test one, screenshots 09, 10 and 13. |

### Refactors

#### 1. Data

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.005` | Adjacency vs. Warp Cost | -- | Split `Region.Boundary` into adjacency-edges (cost 1) and warp-edges (cost `crossingCost`). Deferred from `F.012`. |

#### 2. Engine

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.003` | Action Preconditions | when `A.003` starts | Lift `Action.Validate` into `Action.Preconditions []Precondition` |
| `R.004` | Data-Driven Handlers | when dispatch pain is high | Replace hardcoded goal/ability handler dispatch with TOML-defined handlers + typed primitive registry |
| `R.012` | Engine construction shape | -- | Review constructors so every user (TUI, CLI, tests) gets a fully-wired engine or fails fast at launch. |
| `R.013` | Candidate derivation boundary | when an action changes | Collector methods should always receive engine-derived candidates; retires the adapter's `derive*`/`statRating` ports of engine eligibility logic (revisits tui-turn Decision 3). Investigate drift: `deriveContestedWorlds` copies `seizePlanetTargetWorlds` and needed the same fix twice (`A.001.6` Effort 4 review, DR-3). |
| `R.019` | Migrate `[transport]` onto the effects list | ready | Transport keeps its own typed profile and registration branch (`core.go:58-61`) beside the `[[effects]]` model; fold it in once the S-flag model is fully built, so one parse and dispatch path serves every S effect. |
| `R.021` | Remove dead world-scoped defender filter | -- | `liveDefendersOnWorld` → `eligibleDefendersOnWorld` has no caller outside `spatial_index_test.go` and skips the attack-eligibility hooks; delete it (move its tests to `eligibleDefendersAtHex`) so no future caller bypasses `restrict_attack`. Found in `A.001.3` Effort 4 Commit 4 review. |
| `R.022` | Orchestrator error-handling consolidation | -- | Collapse the 26 recoverable-or-fatal blocks in `orchestrator.go`; only collector calls can return the three recoverable sentinels; target <500 lines, no behaviour change; `D.002` N3 |
| `R.023` | Effect contests through roll hooks | when effects reach a collector | Tripwire Cells, Panopticon Matrix and Lobbyists roll bare 1d10 + stat (`contestPasses`), so Fanatical's and Book of Secrets' rerolls skip them; move all three onto `RollWithHooks` together once reactors and phase effects can reach a collector. `A.001.6` Effort 5 Decision 10. |
| `R.024` | Clear inert `no_attack` records | when saves grow noisy | A Seditionists' `no_attack` stays on its target after the Seditionists die; `domain.AttackBlocked` skips it when read, so it only clutters saves. Clear it when its source leaves play. Deferred from `A.001.6` Effort 3 Decision 10. |

#### 3. Driver

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.015` | Collector reply rigor | -- | `action_collector` swallows bad reply types (`raw.(T)` zero-values); match `phase_collector`'s loud type errors, add the `SelectMovementDecisions` nil guard, drop the `phaseCollector.ask` delegation wrapper. |

#### 4. Presentation

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.016` | Per-ask detail cards | when a 4th card lands | Move per-ask knowledge out of `execution.detailView`'s `activeAsk` switch into an optional overlay interface (`DetailView() string`). |

#### 5. Rendering

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.006` | Shared `styles.Dim` constant | -- | `#6C7086` duplicated across detail/list/manage/layout; extract one shared constant. |
| `R.007` | Extract `rebuildList()` helper | -- | Created/Deleted branches in manage.go copy-paste the rebuild sequence; extract a shared helper. |
| `R.008` | `world.Location` + `campaigns.ValidateID` | -- | `synthesize`/`slugify` reinvent what helpers already do |
| `R.009` | Remove write-only `detail.Model.height` | -- | Field is assigned but never read; dead weight in the detail model. |
| `R.010` | Unify `WindowSizeMsg` routing in manage.go | -- | Per-view asymmetry in how size messages are forwarded; normalise to one pattern. |
| `R.011` | Cache per-frame allocations in `compose()` / root `View()` | -- | Low priority: repeated allocations on hot path; pre-allocate or cache where safe. |
| `R.017` | Evaluate `teatest` for UI tests | -- | TUI tests currently drive `Update` by hand and assert emitted messages; explore `charmbracelet/x/exp/teatest` as a golden-output harness for full model behavior — adopt across overlays/views or document why not. |

### Bugfixes

#### 1. Data

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.002` | Goal/Tag Display IDs | -- | Key `Rulebook.Goals`/`Rulebook.Tags` on semantic table key; `id` becomes display-only. |
| `B.005` | Goal XP | -- | Retype `Difficulty` from `string` to int or tagged sum so engine can dispatch XP |
| `B.007` | Self-warp validation gap | -- | `newRegionMap` accepts a hand-edited warp with `from_region == to_region`; harmless dead edge in pathfinding, but the validator should reject it. |
| `B.012` | Old budget keys in saves | -- | `A.001.3` added a trigger segment to effect budget keys (`e047af7`); a mid-cycle save from before holds old keys, so one extra steal/absorb is possible that turn. Self-heals at the owner's next bookkeeping. |
| `B.015` | Bought assets get no hex | -- | `NewPurchasedAsset` (`asset.go:381`) sets `WorldID` only, so bought assets on different worlds share the zero `RegionHex` and Attack (`eligibleDefendersAtHex`) lets them fight; harness `addActiveAsset`/`AddFaction` copy the gap. Confirmed in the app: sandbox buys save `RegionID = ""`, so they also miss defenders on their own world. Found in `A.001.6` Effort 3, Commit 4. |
| `B.021` | SWN asset data drift | -- | A diff of all 72 assets against the book: transports flagged `S` where the book says `A`; flag gaps on Mercenaries, Surveyors, Tripwire Cells, Strike Fleet, Space Marines, Pretech Manufactory; wrong cost/TL on Pretech Researchers, TL on Surveyors. Audit and fix. Found in `A.001.6` discovery. |

#### 2. Engine

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.001` | Hardcoded goal ID | -- | `"G-012"` (Change Homeworld goal) is hardcoded in `change_homeworld.go`; silently breaks if the TOML ID changes. The asset half (Stealth) closed in `A.001.6`. |
| `B.008` | Stale `ApplyBookkeeping` comment | -- | Doc comment claims it "advances the turn phase to `PhaseAction`", but `TurnState` has no phase field; phase order lives in the orchestrator. One-line comment fix. |
| `B.010` | Reroll uses base die sides | when a non-d10 Cat 1 die lands | `roll.go:86` rerolls via `roller.Roll(baseRoll.Sides)`, so a Cat 1 die of other sides rerolls wrong. Inert today (all d10); Book of Secrets is the first real reroll user. |
| `B.013` | Peaceable Kingdom misses a missed attack | -- | `peaceable_kingdom.go:20-36` detects an Attack from damage or attacker stealth-loss mutations; a miss by an unstealthed attacker emits neither, so the goal ("do not take an Attack action") does not reset. Found in `A.001.6` Effort 1, Commit 6. |
| `B.014` | Capture misses a sibling-removed target | when Treachery hits a self-damager | `capture_target.go` cannot see a sibling reactor's output in its round, so a target that removes itself (Zealots' `self_damage`) still gets `AssetOwnerChanged`; the owner change then misses and the turn aborts. Fence-signed; found in `A.001.6` Effort 1 interim review. |
| `B.016` | Hook budgets survive an aborted action | -- | `dispatch/lethal.go` and `dispatch/roll.go` spend once-per budgets on live state during Resolve; a prompt canceled later in the action drops its mutations but keeps the spend (a lost Medical Center restore or reroll); snapshot `HookBudgets` before `Run`, restore on error. Found in `A.001.6` Effort 2 interim review. |
| `B.017` | Repair price ignores Medical Center | -- | `overlay/repairorders.go` shows `n(n+1)/2` without the `RepairCostModifier` discount, and `RepairAsset.Validate` still needs 1 Coin when a discounted repair is free; the engine charges the right amount. Deferred from `A.001.6` Effort 2. |
| `B.022` | Enabler surcharge reads the turn-start index | when an enabler follows a removal | `purchase_enabler.go` prices through `world.AssetsOnWorld`, the turn-start index, not the action's view; a rival surcharge asset that an earlier asset in the same action removed (a Monopoly levy) still charges. Found at `A.001.6` pre-merge. |
| `B.023` | Transport range never enforced | -- | `TransportProfile.MaxHex` loads but nothing reads it: an order goes to any world. The movement-redesign Effort 2 plan (Commit 1 Task 6) planned the check; the orchestrator's cargo pick checks only `MaxCargo` (`orchestrator.go:779`). Rules call open: what `max_hex` means under multi-turn movement. Found in the sandbox build. |
| `B.024` | Book of Secrets rerolls twice a turn | -- | Its `own_roll` and `enemy_roll_at_me` entries each carry a `once_per` budget whose key includes the trigger, so one own and one forced reroll can both fire; rules call: one shared use per turn. A declined reroll spends no budget by design (`dispatch/roll.go:95-104`), so it re-offers on each roll at that world. Found in the sandbox combat test. |
| `B.030` | One asset loads onto two carriers | -- | `eligibleCargoForTransport` (`orchestrator.go:774`) skips only assets with a live order, and picks apply after the loop, so a later carrier is offered cargo an earlier one took; both carry it, and the second pulled it off Citadel after the first landed it. Found in the sandbox movement test. |
| `B.031` | Resume replays the movement phase | needs triage | Quit at the action phase, then resume: the turn ran its movement tick again (a second Lockdown roll, paid this time after a Coin edit), and `history.jsonl` kept both ticks while state kept one. Found in the sandbox movement test. |
| `B.032` | Unaffordable grant reads as an error | -- | `runGrantedMovement` skips a Transit Web grant the faction cannot pay for on purpose, but `orchestrator.go:862` sends it to `OnError`, so the feed shows `error: relocate: insufficient Coin`. Found in the sandbox movement test. |

#### 3. Driver

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.028` | State campaign ID not checked | -- | `faction.go:76` fills `CampaignID` only when empty, so a `state.toml` from another campaign loads silently and the header names it; the sandbox scenario seeds hard-code `campaign_id = "sandbox"`. Found in the sandbox combat test. |
| `B.038` | Live faction pointers cross the event channel | needs triage | `adapter/observer.go:24-30` sends `*domain.Faction` on the buffered `eventCh`, and `execution.go:193,199` reads it in `snapshotOf()` on the TUI goroutine while the engine may still be writing; likely a data race. Confirm with `go test -race` once `F.027` adds adapter tests. Found while mapping the driver layer. |

#### 4. Presentation

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.004` | `Ability` Narration | -- | Add fallback text for asset ability narration |
| `B.018` | Digest drops the defender name | -- | Cross events name an asset only when it is destroyed (`digest/build.go`), so `{defender_asset}` renders empty on a non-lethal hit, a shelter or a save. Found in `A.001.6` Effort 1. |
| `B.019` | Movement orders read as unknown in the digest | -- | `digest/build.go` has a case for `movement_order_held` only; the five other `movement_order_*` types each become an `unknown mutation` note. Found in `A.001.6` Effort 5 planning. |
| `B.020` | T-011 pick with every world governed | -- | The faction wizard's world select for a T-011 pick lists only worlds no faction governs; when none is left, the select has no options. Found in `A.001.6` Effort 4. |
| `B.025` | Turn right panel shows stale Coin | -- | During a faction's turn the right panel shows Coin from before bookkeeping (Gilt Exchange 17 while the Bribe prompt allowed 24); refresh it after bookkeeping income. Found in the sandbox combat test. |
| `B.027` | Prompts show world IDs | -- | The redirect prompt (`overlay/attack.go:49`) and the Bribe base label (`overlay/bribe.go:39`) print `Location.WorldID` (`mireholt`), not the world's name; the feed's permission line (`execution/format.go:142`) prints the definition ID and world ID (`SWN-W3-003 on aurum`). Found in the sandbox combat and movement tests. |
| `B.029` | Same-label rows toggle together | -- | huh v1.0.0 toggles every option whose label matches the cursor row (`field_multiselect.go:414-421`), so three "Militia Unit" cargo rows (`overlay/cargo.go:29`) or two Treachery attackers (`overlay/attack.go:19`) cannot be picked one by one; make every `NewMultiSelect` label unique. Found in the sandbox movement test. |
| `B.033` | Destinations list the current world | -- | Move orders (`overlay/movement.go:50`, one shared `worldOptions` list) and Transit Web relocates (`overlay/relocateorders.go:36`) offer the world the asset is already on. Found in the sandbox movement test. |
| `B.034` | In-transit assets show a blank location | -- | An asset between worlds has no `WorldID`, so the movement form reads `(at )` (`overlay/movement.go:57-60`) and the right panel reads `at` (`execution/execution.go:475`); show the order's origin and destination. Found in the sandbox movement test. |
| `B.036` | Raw effect key in ability labels | -- | `overlay/ability.go:47` (and the header at `:63`) prints `def.Ability.Steps[0].Effect`, so the GM reads `Pretech Logistics · purchase_enabler`; map effect keys to readable phrases. Found in manual test one, screenshot 05. |
| `B.037` | Counter damage has no label in the feed | -- | `execution/format.go:106-116` labels `counterstrike` by `Cause` but never reads `AssetHPDelta.Counter`, so a normal counter prints as a bare `asset … HP -7`; the digest's version of this gap is `B.018`. Found in manual test one, screenshot 15. |

#### 5. Rendering

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `B.006` | Turn event stream flash | -- | Center-pane event stream visibly redraws/flickers between collector prompts during a cycle; smooth the transition so it doesn't flash on each phase. |
| `B.026` | Action select opens on Skip | -- | The select opens on "Skip faction" (last) with the rows above scrolled out of view; one Up shows them all. Likely cause: `selectaction.go:43` binds `Value(&data.choice)`, whose nil zero value equals the Skip sentinel. Found in the sandbox combat test; reproduced on every turn. |
| `B.035` | Tall overlays push the chrome off screen | -- | Likely cause: no overlay select sets a `.Height(`, so the 16-row Purchase list (`overlay/purchaseenabler.go`) grows past the content budget the root hands down, the title bar and mode tabs scroll off the top, and the feed spills outside the center column. Found in manual test one, screenshot 06. |

### Docs

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `D.001` | Split pre-merge review from execution | when next process session | Process session (Fable). Review session writes triaged findings to the plan (design calls made there); execution session does fixes, then promotion, planned-work and the docs move; skip the split when no finding needs a design call; closing step writes the hand-off as the plan's last Decision Record entry (`D.002` N4) |
| `D.005` | Whole-tree review at arc boundaries | -- | One line in the arc flow: a whole-codebase review runs when an arc's serial spine completes; diff-scoped reviews missed `RunCycle` for 148 days; `D.002` N5 |
| `D.006` | Fold arc concept into arc plan; stop moving discovery docs | -- | Arc concept becomes the arc plan's first section; discovery docs stay put after merge (plans still move); templates 5 → 4; `D.002` N6 |
| `D.007` | Sandbox Cracked Comms wording | -- | `campaigns/sandbox/TESTING.md` says the prompt shows "if Veiled Hand defends outright"; the effect is `ScopeSelf`, so Cracked Comms itself must win the defense and the attacker must survive the counter. Found in the sandbox combat test. |
| `D.008` | Sandbox movement seed and wording | -- | `TESTING.md:268` says each carrier costs 2 Coin, but Heavy Drop costs 1; the seed leaves Ashfall at +2 income against −2 upkeep, so it hits 0 Coin and Heavy Drop can never pay without a state edit. Found in the sandbox movement test. |
| `D.009` | Define the Exploration (`E.`) type | when `E.001` nears promotion | Process session (Fable): add `E.` to `session-modes.md` (routing table, types × artifacts, a flow section) and a template under a new `docs/initiatives/explorations/`; settled so far: Discovery only, ends with findings + options + a recommendation, follow-ons are filed as their own rows, never plans or executes. Must land before `E.001` is promoted. |

### Explorations

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `E.001` | Display translation ownership | after `A.001` | Engine data reaches the GM through label, feed and card code spread across `tui/` with no single owner, and the sandbox runs placed most UX findings there; explore the current design (presentation, drivers, Manage mode, `narrative/digest`) and brainstorm options, from fixing in place to a frontend-neutral layer. Starting reference: [layer map](https://claude.ai/artifact/6UjHPwdY3mq8oQ24HRqvU1). |

### Captured Ideas

Prose notes attached to a Backlog row. Each one is an option to weigh when its row is worked, not a decision.

**`E.001` — driver core + transport split.** One option for the Exploration to weigh against the others, not a decision.

- **Observation:** the engine is synchronous (no goroutines, channels or `sync`). The only goroutine comes from Bubble Tea running the `tea.Cmd` from `Adapter.Run()`. Drivers live in three places: `tui/adapter` (async), `tui/adapter/smoke.go` (sync, `--dryrun`) and `engine/testharness` (sync).
- **Idea:** split the driver in two.
  - *Driver core*, frontend-neutral and synchronous: implements the collectors (33 asks) and `TurnObserver` (14 events) once. At each call, while state is still consistent, it runs a pure presentation step and passes finished views on. When an answer comes back as a key, it maps the key to the domain value.
  - *Transport*, one thin piece per frontend: roughly `Ask(prompt) → key` and `Tell(line)`. TUI: channels, Bubble Tea pump and huh widgets (today's `run.go`, 71 lines). CLI: plain terminal I/O, no goroutine. Harness: answer by script, record the lines.
- **Why it might help:** translation runs at the moment of the event, so names stay right after a capture and `B.038` goes away. A frontend handles a few prompt kinds, not 33 asks (19 overlays already reduce to `NewSelect`, `NewMultiSelect` and `NewConfirm`). Keyed options fix `B.029` by construction. The harness becomes a second consumer for golden tests.
- **Known hard parts:** about 7 multi-step forms (movement, relocate, repair, buy, refit, bribe, expand); the engine asks one question at a time; names of destroyed assets; data the seam does not carry (`F.032`). The adapter's `derive*` copies would move into the engine (`R.013`).

---
<br />

## Open Questions

Unresolved design questions that feed into a backlog item. When the item ships, the question is answered.

| # | Question | Relevant Feature | Backlog Item |
|---|----------|-----------------|--------------|
| 1 | Starting Coin for new factions — no explicit SWN rule; what is the right default (Wealth rating, fixed amount, GM prompt)? | Edit Mode | `A.002` |
| 2 | How does Edit Mode integrate into the Cobra command tree — sub-commands under a top-level `edit` command, or per-entity sub-commands (e.g. `faction edit`)? | Edit Mode | `A.002` |
| 3 | When AI decision-making is added, how does the GM designate which factions are AI-driven vs. manually controlled? | AI Decision-Making | `A.003` |
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

