# Logging + Errors — Implementation Plan

`F-003`. Top-level plan covering the logging and errors initiative. Two efforts ship in sequence on a single feature branch:

- **Effort 1 — Logging.** Stand up `internal/logging` (custom `slog.Handler`, file destination, cascading attrs), plumb a logger field into every sub-engine, emit banners and seed Info/Debug emissions. [`logging-and-errors-effort-1-plan.md`](./logging-and-errors-effort-1-plan.md)
- **Effort 2 — Errors.** Declare promoted sentinels, ratify wrap conventions, wire the orchestrator's Recoverable/Fatal branch (calls `logger.Warn` / `logger.Error`), add the `errors.As` consumer for `MutationApplyError`. [`logging-and-errors-effort-2-plan.md`](./logging-and-errors-effort-2-plan.md)

## Context / Goal

- Discovery doc: [`docs/initiatives/discovery/logging-and-errors-discovery.md`](../discovery/logging-and-errors-discovery.md)
- The two efforts share design questions (errors' Recoverable/Fatal branch calls into logging's `Warn` / `Error`; the `MutationApplyError` consumer emits structured log lines per miss). The dependency direction is one-way: logging is a prerequisite for errors, so logging ships first.

## Decisions Ratified in Planning

1. **Two-effort split, logging first.** Logging is a prerequisite for the errors effort's Recoverable/Fatal branch and `MutationApplyError` consumer. Errors plan-time decisions assume the logger is plumbed.
2. **Single feature branch for both efforts: `feature/logging-and-errors`.** wip: commits per layer; squash to one conventional commit per effort on merge. Pre-merge checklist runs once at branch level (per `feedback_branch_initiative_layering`).
3. **Production CLI entry point deferred.** F-003 ships the `internal/logging` capability and its plumbing through the engine. The production caller (`main.go` is currently a 14-line stub; no production caller of `engine.New` exists) stays a future-work item — either folded into F-005 (TUI Rebuild) or carved out as its own initiative when that decision is made. For F-003, the testharness is the only consumer of `engine.New`, and uses `slog.New(slog.DiscardHandler{})` for tests that don't care about output.
4. **Plumbing order: bottom-up.** Sub-engine constructors take a `*slog.Logger` parameter first (one commit covering all 7 sub-engines — they have no cross-deps and the change is mechanical); then `engine.New` / `engine.NewWithRulebook` accept and thread it; orchestrator `.With(...)` cascades happen in the next commit.
5. **Decisions log scope.** Discovery and Plan decisions live in their respective artifacts. The decisions log (`docs/dev_journals/faction-manager/decisions-log.md`) records only revisions or new decisions made during execution (per `feedback_decisions_log_scope`).

## Open Questions — To Ratify at Implementation Time

- **Effort 1 — Logging:** all four discovery-level open questions (Q6 attr vocabulary, Q7 banner hook points, Q8 `internal/logging` API, Q10 cleanup failure mode) ratified in the effort-1 plan. Q9 (plumbing order) ratified above (decision #4).
- **Effort 2 — Errors:** Q1 (sentinel naming), Q2 ("no selection" granularity), Q3 (`MutationApplyError` consumer location), Q4 (`recoverableErrors` list location), Q5 (phase order within errors effort) — to be ratified in the effort-2 plan session.

## Effort Summary

| # | Effort | Scope | File |
|---|--------|-------|------|
| 1 | Logging | `internal/logging` package, custom slog handler, sub-engine plumbing, attr cascade, banner emissions, seed Info/Debug | [`logging-and-errors-effort-1-plan.md`](./logging-and-errors-effort-1-plan.md) |
| 2 | Errors | Promoted sentinels, wrap conventions, orchestrator Recoverable/Fatal branch, `MutationApplyError` consumer | `logging-and-errors-effort-2-plan.md` *(drafted in a later plan session)* |

## Shared Context

### Branch and merge strategy

- One feature branch: `feature/logging-and-errors`.
- wip: commits during execution (one per layer per CLAUDE.md item 10 and `feedback_git_commits`).
- Pre-merge checklist runs once at branch level after both efforts complete (code review, dev journal update, planned-work update).
- Squash to two conventional commits at merge: `feat(logging): ...` and `feat(errors): ...`. (Or a single `feat: logging + structured errors` — Robert's call at merge time.)

### Model discipline

Per CLAUDE.md model selection table:

| Session | Model | Why |
|---------|-------|-----|
| This plan session (top-level + effort-1) | Opus | Plan mode |
| Effort-2 plan session | Opus | Plan mode |
| Effort-1 execution sessions | Sonnet (default); Opus for the handler + cascade commits | Custom `slog.Handler` and the cascade design are heavier; the rest is mechanical plumbing |
| Effort-2 execution sessions | Sonnet | Mechanical sentinel declarations and orchestrator branch wiring |
| Pre-merge checklist | Sonnet | Always |

Prompt to switch model at each transition (`/model`).

### Canonical attr vocabulary

Ratified rule: **bare nouns in cascade context; `_id` / `_type` suffix only when out-of-context**. Cascade attrs ride a `.With(...)`-bound logger, so their identity is unambiguous from the surrounding emission. The `MutationApplyError` consumer emits per-miss lines *outside* the `.With(faction)` context, so it uses `_id` / `_type` suffixes to disambiguate.

| Attr | Type | Used by | Notes |
|------|------|---------|-------|
| `version` | string | base logger | semver |
| `turn` | int | orchestrator per-turn `.With` | |
| `faction` | string | orchestrator per-turn `.With` | faction ID |
| `phase` | string | orchestrator per-phase `.With` | lowercase phase name (`goal_lock`, `bookkeeping`, `stat_raise`, `movement`, `action`) |
| `engine` | string | sub-engine `.With` at delegation | lowercase sub-engine name (`turn`, `tag`, `effect`, `mutation`, `action`, `goal`, `world`) |
| `asset` | string | event-specific | asset ID |
| `action` | string | event-specific | action name |
| `world` | string | event-specific | world ID |
| `from_hex`, `to_hex` | string | event-specific | hex coords for transits |
| `mutation_type` | string | `MutationApplyError` consumer | out-of-context |
| `faction_id` | string | `MutationApplyError` consumer | out-of-context: per-miss lines fire outside `.With(faction)` |
| `entity_id` | string | `MutationApplyError` consumer | out-of-context |

### Testharness logger

`internal/faction/engine/testharness/harness.go` constructs the engine. After plumbing lands, the harness needs a logger:

```go
log := slog.New(slog.DiscardHandler)
engine, err := engine.NewWithRulebook(rulebook, log)
```

Tests that *do* want to assert on log output use a `slog.NewTextHandler` wired to a `bytes.Buffer`. Test files for logging itself (handler unit tests, banner emission tests) use a custom test recorder if needed.

No nil-guards at sub-engine call sites — the discard handler is cheap and avoids `if log != nil` clutter throughout the engine.

### Naming conventions

- Package: `internal/logging`.
- Banner helpers live in `internal/logging/banner.go` (same package, not a sub-package — keeps the import path short and the public API single-namespaced).
- Sentinel names finalized in effort-2 plan; working names from discovery: `engine.ErrWorldEngineUnavailable`, `action.ErrNoSelection`, `action.ErrPreconditionFailed`.

### What we are *not* changing in F-003

Captured as Out of Scope below — duplicated here for emphasis because both efforts will be tempted to creep into these:

- `TurnObserver.OnError` signature stays as-is (deferred to F-005, predicted Shape B per discovery).
- No exhaustive backfill of every potential Debug/Info site across the engine. Effort-1 commit 5 seeds the pattern at phase boundaries and sub-engine entry/exit; opportunistic additions happen during F-005 and beyond as the TUI needs them.
- No retroactive `fmt.Errorf` cleanup for the 42 leaf-origin sites the discovery audited and cleared. Convention applies going forward; existing leaf-origin sites stay as they are.

## Out of Scope

- **Production CLI driver** — `main.go` wire-up to `engine.New` is deferred. Either folded into F-005 or carved out as its own initiative.
- **TUI implementation** — F-005.
- **Multi-handler fan-out** — the seam (no globals, expose `*slog.Logger` not `slog.Handler`) is preserved; the fan-out itself is built when F-005 needs a TUI sink.
- **`TurnObserver.OnError` redesign** — deferred to F-005 per discovery decision.
- **Exhaustive Debug/Info backfill** across all sub-engines — only the seed lines in effort-1 commit 5.
- **Retroactive cleanup of leaf-origin `fmt.Errorf` sites** — convention applies going forward; the 42 audited sites stay.
- **Log aggregation / external sinks** (Datadog, Loki, etc.) — single-user local tool, out of scope per discovery.
