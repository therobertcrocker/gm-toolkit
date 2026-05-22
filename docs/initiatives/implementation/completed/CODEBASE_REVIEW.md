# gm-toolkit — Codebase Review

**Reviewer:** Claude (Opus 4.7), commissioned 2026-05-22
**Branch reviewed:** `feature/movement-redesign` (5 commits ahead of `main`, includes EffectsEngine + transport reactor wip)
**Build/test status:** `go build ./...` clean, `go test ./...` all packages pass
**LOC:** ~17,000 lines of Go (engine + spatial + narrative + tests)

This review is intentionally critical. The codebase has real strengths; the absence of compliments is not absence of approval — it's compression. Where I praise something, it's because I want to call it out as load-bearing.

---

## 1 — Executive Summary

You have a **mechanically sound, well-architected game engine** with no delivery vehicle. The engine — domain types, mutation pipeline, hook system, orchestrator, action implementations, goal/tag handlers, spatial primitive, narrative renderer — is the work of someone taking the craft seriously. Test coverage is real (1,800 LOC of action tests, 1,600 LOC of scenario integration tests). The composition root wires cleanly. Mutations flow through a single writer. The hook dispatch is honest about its categories and depth.

But **no human can actually use this tool today.** The root `main.go` is a 14-line stub that prints "Usage: gm-toolkit <tool>" and exits. The `cmd/faction-manager/` directory described extensively in both `README.md` and `docs/architecture-overview.md` does not exist on disk. None of the user-facing commands (`faction create`, `narrate`, etc.) — described in those docs as built or in progress — have actual entry points. The Cobra command tree shown in the README has no source files. The "TUI" listed as "Not started" in the dev journal is accurate; the CLI listed as "Complete" is not.

The gap between **engine readiness** and **product existence** is the central fact of this codebase. Everything else flows from it.

---

## 2 — Methodology

What I read in full:
- Architecture overview, planned-work, session-modes, README, HANDOFF
- First and last sections of decisions-log (the index and the most recent 8 named branches, decisions #181–234)
- `main.go`, `engine/core.go`, `orchestrator.go`, `collector.go`, `observer.go` references
- All of `domain/` (faction, asset, mutation, turn, location, base, event, roll)
- `state/faction_state.go`, `config/config.go`, `rulebook/rulebook.go`
- All of `engine/mutation/`, `engine/world/`, `engine/effect/`
- `engine/turn/turn.go`, `bookkeeping.go`, `order.go`
- `engine/action/action.go`, `collector.go`, `actions/register.go`, `actions/attack.go`, `actions/use_asset_ability.go`
- `engine/hooks/registry.go`, `interfaces.go`, `dispatch/mutations.go`, `dispatch/roll.go`
- `engine/goal/goal.go`, `goals/planetary_seizure.go`; `engine/tag/tag.go`, `tags/scavengers.go`
- `effect/effects/transport.go`
- All of `internal/spatial/`
- `narrative/wire_renderer.go`, `narrative/digest/build.go`, `narrative/digest/types.go`
- `testharness/harness.go`, sample integration tests (`full_cycle_test.go`, `movement_lifecycle_test.go`)
- Recent git history (30 commits) and version tags

What I didn't read:
- Remaining 8 goal handlers (I read Planetary Seizure as a representative sample), 3 of 4 tag handlers
- 7 action implementations (read Attack and UseAssetAbility; skimmed the registration of the rest)
- Hook test suites, mutation tests, action test bodies (only inspected sizes)
- Discovery docs and implementation plans (only enumerated them)
- Older sections of the decisions log (#1–180)
- TOML data files (rulebook content)
- Style guide, SWN mechanics reference

Confidence levels are flagged inline below where the reads were partial.

---

## 3 — Documentation vs. Reality

### 3.1 The missing CLI is not a small inaccuracy

`README.md` describes:
- `faction create` (interactive wizard)
- `faction list`
- `faction delete`
- `turn` (interactive turn wizard)

`docs/architecture-overview.md` describes a `cmd/faction-manager/` tree with `main.go`, `config.go`, `commands/app.go`, `commands/faction/`, `commands/narrate/`, `commands/paths/`, etc. — and uses present tense throughout ("The current delivery mechanism is a pure CLI").

**None of this exists.** `find . -type d -name 'cmd*'` returns nothing. The only `main.go` is at the root and does nothing.

`docs/dev_journals/faction-manager/dev-journal-factions.md` Feature List marks Faction CRUD, Turn Engine, All 9 SWN Actions, Narrative Renderer as **Complete**. The *engine* support for these is complete. The *commands* are not.

Decision #234 in the decisions log acknowledges this directly: "Production wiring (when a CLI/TUI driver materializes) will need its own hook-registration site, but the test API need not predict its shape." That's the honest take. The README and architecture overview are not.

**This is the single most important coherency problem in the project.** A new contributor reading the docs would expect to clone, build, and run `faction-manager turn`. They cannot.

### 3.2 Architecture doc is stale on three structural points

The arch doc describes:
1. **`ability/` as a sub-package** — it's gone. R-001 retired it; ability is now under `action/actions/ability/`. Decisions #216–217 ratified the move.
2. **5 turn phases ending with "Apply, Record, and Save"** — there are now 6 phases. The orchestrator interleaves `setupFactionTurn → runGoalLockPhase → runStatRaisePhase → runBookkeepingPhase + Save → runMovementPhase → runActionPhase + Save → finishFactionTurn`. The Movement phase is new (per F-002). Stat raise also moved *before* bookkeeping in code, while the doc places it *after* (Phase 2B).
3. **No mention of `effect/` or `world/`** — both exist and are wired into `Engine` (lines 37, 41 of `core.go`). Both are referenced in recent decisions (#197–210, #232).

The Sub-Engine Shapes section is internally consistent and useful, but the "Sub-engines: mutation" / "Sub-engines: action, hooks" classifications need updating to include `world` (Shape 3) and `effect` (Shape 2 data-mirroring).

### 3.3 Planned-work.md is honest

This doc is well-maintained and reflects current reality:
- F-001 (Spatial) — accurately Blocked
- F-002 (Movement Redesign) — accurately Paused (Effort 2 partial)
- R-001 (Ability Redesign) — accurately In-Progress
- F-004 (CLI Rebuild) and F-005 (TUI Rebuild) are in the Backlog with the trigger "When tool is mechanically complete"

The fact that F-004 ("CLI Rebuild") is *in the backlog* and the README describes the CLI as *already built* is the contradiction. Planned-work is right; README is fiction.

### 3.4 HANDOFF.md is dead context

References Phase 5 of the narrative renderer initiative, long since merged. Should be deleted or replaced.

---

## 4 — Engine Quality: Strengths

### 4.1 Domain model is clean and disciplined

`domain/` is pure data. No I/O, no framework imports. Mutations are typed structs with stable `Type() string` discriminators, JSON tags for history serialization, and `Cause` + `CausedByFactionID` fields for attribution. The cause/causedBy attribution is *load-bearing*: the digest builder uses `Cause` to disambiguate composite events (repair vs. attack vs. expand), and the narrative renderer relies on the same fields. This is the kind of disciplined modeling that pays off downstream.

The `Mutation` interface is sealed by convention: the `MutationEngine.Apply` switch panics on an unknown type, which forces every new mutation to be wired everywhere it needs to be. That's the right tradeoff.

### 4.2 The composition root is small and explicit

`engine.New(cfg)` does the right things in the right order. The Effects engine is auto-registered from the rulebook (`for _, def := range rulebook.Assets { if def.Transport != nil { ... } }`) — a clean example of data-driven registration. Compare against `TagEngine.New()` which hard-codes its handlers; the distinction is principled (tag behavior is fixed; transport behavior is parameterized) and Decision #232 articulates it well.

### 4.3 The orchestrator is properly factored

Phase methods extracted (`runGoalLockPhase`, `runBookkeepingPhase`, etc.) — recent refactor work per the commit log (`b5e764d wip: extract phase methods`). Each phase has the same shape: compute, optionally dispatch reactors, applyAndRecord, fire observer, await checkpoint. The `applyAndRecord`/`state.Save` split (Decision #195–196) is correct: history must be exhaustive, state checkpoints only at meaningful resume boundaries.

### 4.4 Hook system is honest

Five categories, distinct dispatch protocols, scoped registry (global / faction / asset). The fact that each category has its own interface and its own dispatch function is the right call — a single "hook" abstraction would have hidden the actually-different semantics. Decision #158's rejected alternative (flat slice with scope field) shows the right reasoning.

Roll dispatch (`dispatch/roll.go`) handles modifier offers, budget gating, elective rerolls, keep-highest trimming — all in one place. Cat 3 reactor recursion has a depth cap (5) that's enforced and tested.

### 4.5 Attack action is the high-water mark

`actions/attack.go` is the most complex action and it handles all the SWN attack-resolution edge cases: multi-attacker batch, per-matchup live HP tracking, stealth-cleared-once-per-sequence, redirect-to-base prompt, tie resolution via hook, counterattack damage, inline `AssetRemoved` on lethal hit. The local HP trackers (`assetHPTracker`, `baseHPTracker`) so that mid-sequence checks see effective HP is the right pattern; without it, the second attacker in a batch could try to hit a defender that the first attacker already killed.

If I were a senior engineer reviewing a junior's PR for this action, my notes would be minor — naming nits, maybe a sub-function extraction. The mechanics are right.

### 4.6 Spatial package stands on its own

`internal/spatial/` has zero faction-domain imports. It's a hex-based region map with Dijkstra pathfinding, boundary connections, drift-rated crossing costs. The validation at load time is thorough (boundary From/To membership, region existence, world hex membership). The `RegionMap` could be lifted out for reuse if you ever decided this work was useful beyond the faction tracker. The narrowing of `spatial.SpatialMap` to `Location()` only (Decision #219) — and pushing routing into the consumer-side `world.HexRouter` (Decision #222) — is the cleanest example of consumer-driven interface design in the codebase.

### 4.7 Test coverage is real, not theatrical

The action tests are 1,800 LOC. The scenario integration tests are 1,600 LOC. `movement_lifecycle_test.go` walks the full issue→tick→complete arc across three turns with post-Apply state assertions and history-record verification. The harness's `StubSpatialMap` is genuinely useful — it gives action tests a populated world index without disk I/O. The decisions log (#230) explicitly states the test strategy: "state inspection plus persisted-history checks ... each catches a distinct bug class."

This is real testing discipline. A lot of solo projects do not get here.

### 4.8 The narrative digest is more substantial than it looks

`narrative/digest/build.go` is 500 lines of mutation-stream-to-narrative-beats translation, handling cause-keyed composite events (repair = coin_delta + faction_hp_delta or coin_delta + asset_hp_delta), cross-faction attribution, base-hit aggregation, refit pre-scanning to set the "upgraded from X" link, and a stable headline-selection layer. The wire renderer uses seeded templates for natural-prose variation. This is a real piece of software, not a `fmt.Println` afterthought.

---

## 5 — Engine Quality: Concerns

These are things I'd flag in code review. They range from minor to "would bite in production."

### 5.1 Mutation engine: silent no-op on missing entities

Every case in `MutationEngine.Apply` is `if faction, ok := ...; ok { ... }`. A mutation targeting a deleted or unknown faction or asset silently does nothing. No log, no error, no panic. This is permissive in a way that masks bugs — if some upstream code emits a stale `FactionID`, you'd never know until state diverges.

**Recommendation:** at minimum, log unknown targets. Better: emit a structured error path (a `*MutationApplyError` accumulator) that the orchestrator can decide what to do with. This won't be expensive to add and would catch a class of bugs that are otherwise invisible.

### 5.2 ScavengersReactor has an unbounded growth bug

`tags/scavengers.go`:
```go
type ScavengersReactor struct {
    FactionID string
    credited  map[string]bool // guards against re-crediting the same asset across recursion rounds
}
```

The `credited` map is initialized lazily on first call and never cleared. Reactors are registered once at startup (`Tag.ApplyAll`) and persist for the lifetime of the engine. Over many cycles, this map grows unbounded with every asset ID the faction ever destroyed.

The cycle-recursion concern the comment alludes to is real — but the right scope for that map is "within one MutationReactors dispatch call," not "for the lifetime of the engine." A long campaign with thousands of kills will leak memory; in practice the engine probably isn't long-running enough for it to matter, but it's wrong.

**Recommendation:** clear `credited` at the start of each `OnMutations` call, or pass the recursion-aware deduplication down from the dispatcher (since this concern is shared across reactors).

### 5.3 Bookkeeping advances Phase past Movement

`turn/bookkeeping.go:57`: `factionState.CurrentTurn.Phase = domain.PhaseAction`.

But the orchestrator runs Movement *between* Bookkeeping and Action. A turn paused mid-Movement (the `AwaitCheckpoint(CheckpointMovement)`) will resume with `Phase == PhaseAction`. The `ApplyBookkeeping` early-return on `Phase != PhaseBookkeeping` will then prevent re-running bookkeeping (good — idempotent), but `runMovementPhase` has no phase guard and will re-tick movement orders on resume.

I am not certain this is a bug (the resume code path may not exist yet, since there's no driver). But the `domain.PhaseMovement` enum value is defined and unused. Either the state machine should transition through it, or the enum value should be removed. Right now you have a half-wired state machine where one phase is named but not tracked.

### 5.4 The `dispatch.MutationReactors` recursion error has a documentation mismatch

Its docstring says "already-collected mutations are still returned" when the depth cap trips. Technically true — the returned slice is non-nil. But the only caller (`orchestrator.runMovementPhase`, `runActionPhase`) checks `err != nil` and returns the error *without* applying the partial mutations. So the actual behavior is "all-or-nothing rollback on cap trip," which is probably what you want — but the docstring implies callers might choose to apply what they have. Either fix the docstring or document why the rollback is the chosen semantic.

### 5.5 `TickMovementOrders` will nil-deref if a definition is missing

`world/movement.go:33–34`:
```go
def := rulebook.Assets[asset.DefinitionID]
newStepIdx := asset.CurrentOrder.StepIdx + def.Speed
```

If `def == nil` (a rulebook that's been edited to remove an asset definition while a campaign has live assets referencing it), this panics. The orchestrator's eligibility filter (`eligibleMovableAssets`) checks `def != nil && def.Speed > 0`, but that filter is for *issuing* new orders, not *ticking* existing ones. An asset with a CurrentOrder whose definition has been removed will crash bookkeeping.

The same pattern appears elsewhere in actions (`AttackAction` checks `def, ok := rulebook.Assets[attacker.DefinitionID]; !ok` defensively; movement does not).

**Recommendation:** mirror Attack's defensive lookup, or treat missing-definition as a hard error at engine load time.

### 5.6 Movement mutation field redundancy

`MovementOrderProgressed` carries both `NewHexCoords HexCoord` and `Region string`, but the Path itself contains `RegionHex` (which has both fields). The mutation's separate-field representation is read by `mutation.Apply` to reconstruct a `Location{RegionHex: {RegionID: v.Region, Coord: v.NewHexCoords}}` — and the same reconstruction happens in `transport.go`. Just storing a `RegionHex` field would collapse this duplication. This is a small thing but it's the kind of paper cut that compounds across many readers and writers of the same mutation type.

### 5.7 `Location.WorldID == ""` is a load-bearing sentinel without a constant

Per Decision #226, an asset mid-flight has `Location.WorldID == ""` (only the `RegionHex` is set). This is fine, but it's an implicit invariant scattered across the codebase: `transport.go` builds `Location{RegionHex: ...}` leaving WorldID empty; `mutation.Apply` for `MovementOrderIssued` sets only `RegionHex`; `RebuildIndex` filters out unknown WorldIDs which silently includes the in-flight case; `digest/build.go` reads `m.ToLocation.WorldID` directly without considering the empty case.

A named constant or helper (`IsInFlight(loc Location) bool`) would document this invariant and prevent the next contributor from quietly mis-handling it. The decision is sound; the encoding deserves more visibility in code.

### 5.8 The maintenance cost comment is stale

`turn/bookkeeping.go:111` says "Returns 0 as the base until maintenance cost data is added to AssetDefinition; registered MaintenanceCostModifiers may override." But `rulebook.Assets[asset.DefinitionID].Maintenance` is now passed as the base — data was added. The comment lies. F-008 ("Maintenance costs per asset") is still in the backlog with trigger "Cost data added to AssetDefinition in TOML" — that trigger has fired but the backlog entry didn't get updated.

### 5.9 The orchestrator's stat-raise placement vs. SWN rules

Decision #183 says "Stat raise phase fires after `CheckpointBookkeeping`, before action selection" — but the code in `RunFactionTurn` orders it: GoalLock → StatRaise → Bookkeeping. Stat raise is *before* bookkeeping in the actual orchestrator. Either Decision #183 is now wrong, or the code is. The architecture doc agrees with the decision (Phase 2B after Phase 2 Bookkeeping). Three sources, three different positions.

Practically: putting stat raise before bookkeeping means the income calculation in bookkeeping uses the *new* stat values, which could be intentional (you raised Wealth, you get more Wealth income this turn) or accidental (the player paid XP for stats that don't help them this turn). Either way, decide and align.

**Confidence on these concerns:** high for 5.1, 5.2, 5.5, 5.6, 5.8, 5.9 (I read the code directly). Medium for 5.3 (resume code path doesn't exist, so it's a forward-looking concern). Medium for 5.4 (caller behavior depends on intent).

---

## 6 — Architectural Coherency

### 6.1 Internal coherency: high

The architecture as described (when described accurately) is followed consistently in the code:
- Mutation pipeline is a single writer
- Actions are pure: Validate → Inputs → Resolve → Output(mutations)
- Sub-engines don't call each other; the orchestrator coordinates
- Collector/observer split: input via collector, output via observer, both injected
- Hooks are registered at startup, dispatched at well-defined seams
- `state.Save` and history writes are deliberately decoupled with different cadences

The Sub-Engine Shapes taxonomy (Stateless Dispatcher / Open Registry / State-Owning Subsystem) is real and reflected in the code. New engines fit one of the shapes; the alignment refactor is making existing engines fit more cleanly.

### 6.2 External coherency: broken

The tool's stated purpose is "a CLI toolkit for tabletop RPG game masters." The CLI does not exist. A GM cannot use this tool. The engine doesn't run end-to-end against any UI — only against the test harness. Every claim of user-facing capability in the README is currently false.

This isn't a small gap. It's the difference between "this is a tool" and "this is a library that will become a tool someday." The architecture explicitly preserves the path to a CLI/TUI (InputCollector + TurnObserver are designed to be implemented by any driver), but until *some* driver exists, the engine is an artifact, not a tool.

### 6.3 Boundary discipline holds

The `internal/` vs. (would-be) `cmd/` split is meaningful. The engine knows nothing about terminals. The action layer abstracts input through `Collector` interfaces. No framework imports leak into domain. This is the right boundary for a long-lived project, and it's been respected even through major refactors.

When the CLI is built, none of the engine should need to change. That's the payoff of the discipline.

---

## 7 — Process & Documentation Trajectory

### 7.1 Process is unusually thorough for a solo project

The project has:
- A 321-line `session-modes.md` defining Discovery / Plan / Execution as separate sessions, with refactor-specific deltas (re-grounding before action, descriptive-vs-authoritative artifacts), bugfix tiering, docs/chore flow, and model-selection per session type
- A 234-numbered-decision log grouped by 27 named branches, with rationale and rejected alternatives for each
- A `planned-work.md` with Backlog → Up Next → Planned Initiative lifecycle, manual promotion only
- Per-initiative discovery and implementation plan files (movement-redesign has 4 plan files, ability redesign has its discovery doc)
- Template files for discovery docs and implementation plans
- Pre-merge checklist with senior-engineer review, dev journal update, version bump assessment
- A CLAUDE.md routing table mapping work areas to required-reading docs

This is more process than most three-person teams maintain. It's working for you (the engine is well-built and recent refactors landed cleanly), so I'm not saying *retire it*. But there are costs.

### 7.2 The cost is observable in two ways

**Doc-bloat ratio.** The docs are ~1,000+ lines just for tracking artifacts (decisions log + dev journal + planned-work + session-modes), plus discovery and plan files per initiative. The engine is 17,000 lines. The ratio is high. Every refactor regenerates doc burden: when the architecture doc fell behind (sections 3.2), it didn't get updated in real time — instead the canonical truth migrated to the decisions log. The architecture doc now serves new-contributor onboarding poorly because it describes an old shape.

**Refactor-dominated cadence.** Of the last ~30 commits, the majority are `refactor:` or `wip: refactor:`. The decisions log's last 7 named branches are: refactor/faction-assets-map, refactor/persistence-write-cadence, feat/spatial-effort-2, docs/planned-work-refactor, refactor/sub-engine-alignment, refactor/ability-engine-redesign, feature/movement-redesign. The one feature is a redesign of an existing feature. **User-visible functionality has not advanced in many sessions.** What's advanced is internal shape.

This isn't necessarily wrong — many of the refactors (sub-engine alignment, ability retirement, movement model) are real improvements that unblock future work. But there's a question implicit in the velocity: at what point does internal shape stop unblocking and start substituting for the work it was supposed to unblock?

The Backlog item F-004 ("CLI Rebuild") is gated on "When tool is mechanically complete." That's an open-ended trigger. The engine has 234 ratified decisions and the trigger hasn't fired.

### 7.3 The process is genuinely well-designed

Some specific things that earn credit:
- The refactor flow's re-grounding requirement (artifacts decay; verify against current code first) directly addresses a real failure mode and is articulated clearly
- The decisions-log scope rule ("implementation-time decisions only, not discovery or plan") prevents the log from becoming a discussion archive
- The deferred-refactor rule ("new code matches existing imperfect pattern, not target") prevents pre-refactor cherry-picking that produces mixed state
- The model-selection table (Opus for design, Sonnet for execution) is a sensible cost discipline
- Decisions consistently include rejected alternatives with rationale — invaluable for future-you trying to remember why

These are the artifacts of someone who has studied software engineering processes. They're worth keeping.

---

## 8 — Viability Assessment

### 8.1 Viability as a tool *today*: zero

A GM cannot install and use gm-toolkit. There is no binary that does anything useful when run. The engine cannot be exercised except through `go test`. This is not a stretch interpretation — it is literal.

### 8.2 Viability as a tool *if a CLI/TUI is added*: high

The engine is ready. The interfaces (`InputCollector`, `TurnObserver`, `PhaseCollector`) are stable and documented. A first-pass CLI built on `huh` (per project convention) implementing PhaseCollector and a minimal observer would put a working tool in your hands inside ~1 week of focused work. The engine already runs end-to-end in the scenario tests — the harness's `ScriptedCollector` is the spec for what a CLI's prompt-driver needs to do.

The test harness is your insurance policy here: when you build the CLI, the engine has been exercised by ~3,400 LOC of tests, including movement lifecycle, goal locks, attack matchups, hook dispatch, and full-cycle integration. The CLI will only need to drive the existing surface, not rebuild it.

### 8.3 Viability as a *continued investment*: high if direction is right

The codebase is a pleasure to work in. The discipline of the existing patterns means new contributions have a clear shape to follow. The test infrastructure makes adding features safe. The decision log makes pattern choices defensible.

The risk is not the code. The risk is the gravity of internal-refactor work. F-006 (Programmatic tag handling — 18 more tags), F-008 (Maintenance per asset — trigger condition met), F-009 (Starting Coin — trivial), B-001 through B-005 (small bugs) are all sitting in the backlog while the engine gets reshaped. Each of those is a *user-visible* improvement (when there is a user). Right now, none of them can be experienced.

### 8.4 Viability as a *learning vehicle*: excellent

If part of what this project is *for* is for Robert to practice Go engineering at scale, design APIs that survive refactors, and develop a personal process for collaborating with an AI co-pilot — then it's succeeding. The artifacts of that learning (the decisions log, the session-modes doc, the refactor flow) have value beyond this codebase.

I don't think it's wrong to optimize for this. But it should be named.

---

## 9 — Specific Risks

### 9.1 Doc-vs-reality drift will get worse

Every refactor adds load to keeping the architecture overview current. The current arch doc is already 3+ structural changes behind. The longer this continues, the harder it gets to recover. A new contributor (or future-you after a long gap) will hit the doc, build a mental model, and then have to debug the difference against the code. **Recommendation:** either update the arch doc on every merge (add to the pre-merge checklist) or rewrite it once before the next major refactor and accept that it will go stale until then. The current half-update state is the worst of both options.

### 9.2 The half-wired PhaseMovement / paused state machine

If you do build a CLI/TUI and someone pauses mid-Movement and resumes, the behavior is currently undefined (probably re-ticks movement, which double-charges transport coin and double-advances orders). This is a forward-looking risk, but it's worth fixing *before* the CLI exists, because the bug won't surface until then and will be confusing to debug because the engine "works" in tests.

### 9.3 No structured logging or error context

Per F-003 in Up Next: "No `log` package calls anywhere." When errors do happen (from `state.Save`, from `MutationReactors` cap-trip), the orchestrator wraps them with one-line context and bubbles up. There's no structured event log of "what happened during this turn." For a tool that already prides itself on a complete history log, the *operational* observability gap is striking. This becomes a real issue the first time something weird happens in a live campaign and you have nothing but the JSONL history to forensic from.

### 9.4 The 1.0 tag is misleading

You're at v1.3.1. Conventional semantics: 1.0 = first production release. There's been no production release because the tool can't be run. The version tags are tracking engine milestones, which is fine internally, but if you ever publish this anywhere public, the version numbers will communicate something that isn't true.

### 9.5 Effects/Hooks recursion has interaction surface

The Cat 3 dispatcher recurses until quiescence (cap 5). Multiple reactors registered at the same scope will all fire on every dispatch round, seeing each other's emitted mutations. The current set (Scavengers, Transport) is small and the interactions are simple. But the system makes it easy to register more, and the next reactor that emits a `CoinDelta` will be seen by Scavengers (which filters on `AssetRemoved`/`cause==attack` so it's safe), and so on. There is no central document of "what reactors exist, what they emit, what they react to" — that picture is currently distributed across each handler file. This will get harder to reason about as more S-flag assets and tags become handlers.

### 9.6 No CI

I didn't find `.github/workflows/`, `Makefile`, or any other CI signal. Tests pass when run locally; nothing is enforcing that they keep passing across commits. This is small to add and high value before the codebase grows further or any other contributor touches it.

---

## 10 — Recommendations

**Priorities in order, opinionated:**

1. **Reconcile the README and architecture doc with reality** (one focused session). Either describe the current state of the engine accurately and say "CLI is not yet built," or commit to building the minimum CLI first. The current state is misleading whether read by you-in-six-months or any contributor.

2. **Build the minimum CLI before the next refactor** (one initiative, ~1 week). One `faction-manager` binary, two commands: `turn` (runs `RunCycle` with a `huh`-based PhaseCollector and a stdout-printing observer) and `narrate <cycle>` (already-implemented engine surface). This unblocks F-006/F-008/F-009/B-001/B-005 by making them experienceable. It also exposes whatever issues the engine has when actually driven end-to-end (and there will be some — there always are).

3. **Add CI** (15 minutes). A GitHub Actions workflow running `go build ./... && go test ./...` on push to any branch. Cheap insurance.

4. **Fix the named bugs in this review** (one session each):
   - 5.2 ScavengersReactor unbounded `credited` map
   - 5.5 TickMovementOrders nil-deref on missing definition
   - 5.3/5.9 reconcile PhaseMovement state machine + stat-raise ordering vs. decisions log
   - 5.1 add some form of failure signal to MutationEngine on missing entities

5. **Add an "Architecture vs. Reality" check to the pre-merge checklist.** When a branch changes engine shape, the architecture overview is part of the diff. Currently the decisions log captures the change but the architecture doc doesn't get updated.

6. **Delete HANDOFF.md** (10 seconds). It's stale by months.

7. **Don't expand the process docs further.** The session-modes doc is at saturation for a project this size. Adding more structure has diminishing returns. If something doesn't fit the existing flow, it's worth asking whether the work itself is mis-scoped before adding a new flow variant.

---

## 11 — Closing

This codebase is a paradox: an engine of unusually high quality wrapped in zero user-facing surface. The discipline that built the engine — clean boundaries, typed mutations, principled refactors, real test coverage — is real and worth preserving. The process scaffolding that supports it is more elaborate than the project requires but is producing the artifacts it was designed to produce.

The risk isn't that the code is bad; it's that the code never meets a user. Every additional refactor cycle without a delivery vehicle deepens the trough between "the engine is ready" and "the tool exists." The trough is filled by writing the CLI — by deciding that the engine is mechanically complete *enough* and giving it a way for a GM to drive it.

Once the CLI exists, this becomes a tool. Until then, it's a very well-built library waiting for a user.

---

*This review was generated against `feature/movement-redesign` at commit 17708fe. Some claims about specific code may shift as the branch evolves before merge.*
