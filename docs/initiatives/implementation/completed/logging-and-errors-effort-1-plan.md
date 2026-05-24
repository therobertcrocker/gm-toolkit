# Logging + Errors — Effort 1: Logging

`F-003`, effort 1 of 2. Implements the logging capability described in the discovery doc's [Logging](../discovery/logging-and-errors-discovery.md#logging) section.

## Context / Goal

- Top-level plan: [`logging-and-errors-plan.md`](./logging-and-errors-plan.md)
- Discovery doc: [`logging-and-errors-discovery.md`](../discovery/logging-and-errors-discovery.md)

Stand up `internal/logging` with the custom `slog.Handler`, plumb a logger field through every sub-engine and the orchestrator, cascade attrs at each layer, emit banners at hook points, and seed Info/Debug lines at phase boundaries. After this effort lands, the engine emits structured logs to `logs/<timestamp>.log` (with `latest.log` symlink) anytime it runs.

## Decisions Ratified in Planning

1. **Q6 — Attr vocabulary follows the bare/`_id` split.** Cascade attrs use bare nouns (`faction`, `world`, `asset`); out-of-context lines (the `MutationApplyError` consumer's per-miss emissions) use `_id` / `_type` suffixes. Full table in the top-level plan's Shared Context.
2. **Q7 — All five phases get banners.** Banner emissions happen in `RunFactionTurn` (turn banner) and at the top of each `run*Phase` method (phase banner): `runGoalLockPhase`, `runBookkeepingPhase`, `runStatRaisePhase`, `runMovementPhase`, `runActionPhase`. Setup / finish (`setupFactionTurn`, `finishFactionTurn`) emit Info-level entry/exit lines, no banner. The run header is emitted once at the top of `RunCycle`.
3. **Q8 — `internal/logging` public API.** Four functions, single package (banner helpers in `banner.go`, same `logging` namespace):

   ```go
   func New(logsDir string, debug bool) (*slog.Logger, error)
   func RunHeader(log *slog.Logger, version, campaign string, factions int)
   func TurnStart(log *slog.Logger, turn int, faction string)
   func PhaseStart(log *slog.Logger, phase string)
   ```

   All three banner helpers emit a single slog line with a `_banner` attribute carrying the banner kind (`header`, `turn`, `phase`). The custom handler dispatches on `_banner` and renders the appropriate visual block instead of an ordinary log line.

4. **Q9 — Bottom-up plumbing.** (Ratified in top-level plan, restated here.) Sub-engine constructors take `*slog.Logger` first (one commit covers all 7 — they have no cross-deps); then `engine.New` / `engine.NewWithRulebook` accept and thread it. Orchestrator `.With(...)` cascades happen in the next commit.

5. **Q10 — Cleanup failure mode: open-new-first, Warn on cleanup failure, never abort.** `logging.New` opens the new log file first (so file-open failures surface as a real `error` return that aborts startup). Cleanup runs after, against the old files. If cleanup fails (permissions, missing dir, etc.), emit a `Warn`-level line into the new log file and continue. The count cap is briefly violated (N+1 files) until the next run cleans up — acceptable.

## Open Questions — To Ratify at Implementation Time

- **`latest.log` symlink behavior on filesystems without symlink support.** On WSL2 / Linux it's a non-issue; if the toolkit ever runs on a non-symlink-capable target, the symlink step should fail gracefully (Warn-and-continue, same posture as cleanup). Decide at first observation; not blocking.
- **Exact pre-cleanup count.** Working assumption: keep 50. Adjustable as a single named constant. No need to settle now.

## Work Breakdown

5 commits, each = one execution session. Branch: `feature/logging-and-errors`.

### Commit 1 — `feat(logging): scaffold internal/logging with file handler and cleanup`

**Recommended model:** Opus. Custom handler rendering involves enough detail (column alignment, `_banner` dispatch, three banner shapes) that the heavier model pays off.

#### Task 1 — `internal/logging/logging.go`

`New(logsDir string, debug bool) (*slog.Logger, error)`:

1. `os.MkdirAll(logsDir, 0o755)`.
2. Build filename: `logs/<YYYY-MM-DD_HHMMSS>.log` using local time.
3. `os.OpenFile(...)` with `O_CREATE|O_WRONLY|O_APPEND`. Return error if it fails (real startup failure).
4. Update `logs/latest.log` symlink: `os.Remove(symlink); os.Symlink(filename, symlink)`. If either step errors, ignore for now (revisit per Open Question).
5. Construct the custom handler with the open file and level (Debug if `debug`, else Info; `AddSource` if `debug`).
6. Build `log := slog.New(handler)`.
7. **Run cleanup** against `logs/*.log` (excluding the symlink and the just-opened file). Keep the last 50, lex-sort, delete the rest. On error, `log.Warn("log cleanup failed", "err", err)` and continue.
8. Return `log, nil`.

Constants:
- `keepCount = 50`
- `latestSymlink = "latest.log"`

#### Task 2 — `internal/logging/handler.go`

Custom handler implementing `slog.Handler`. Target ~50 LOC plus rendering helpers.

Required methods on the handler interface:
- `Enabled(ctx, level) bool` — filter by min level.
- `Handle(ctx, record) error` — the rendering entry point.
- `WithAttrs(attrs) Handler` — return a copy carrying bound attrs (the cascade mechanism).
- `WithGroup(name) Handler` — return self (we don't use groups; document this).

`Handle` dispatch:
1. Inspect `record.Attrs` (and the bound attrs from `WithAttrs`) for `_banner`.
2. If `_banner` present:
   - `"header"` → render the multi-line file prologue using the carried attrs (`version`, `campaign`, `factions`, plus the timestamp from the record).
   - `"turn"` → render the box (`┌─...─┐` / `│ Turn N · Faction │` / `└─...─┘`) using `turn` and `faction` attrs.
   - `"phase"` → render `  --- <Phase> Phase ---` using the `phase` attr (titlecased).
3. Else: render an aligned ordinary line — `HH:MM:SS  LEVEL  source  message  attrs`. Suppress already-banner-displayed attrs (`turn`, `faction`, `phase`) from the trailing attr list — they're inherent context, redundant on every line.

**`_banner`-suppression list** (attrs *not* rendered as trailing key=value because they appear in a banner above): `turn`, `faction`, `phase`. The handler still receives them via the cascade; it just doesn't print them per-line.

#### Task 3 — `internal/logging/handler_test.go`

Table-driven golden tests covering each rendering shape:
- Ordinary Info line with cascade attrs (turn/faction/phase suppressed, event-specific attrs rendered).
- Debug line with `AddSource` enabled (file:line appears).
- Warn / Error lines (level column renders correctly).
- Each banner shape (`header`, `turn`, `phase`) — full multi-line output asserted against golden strings.

Use a `bytes.Buffer` as the writer. No file I/O in handler tests.

#### Task 4 — `internal/logging/logging_test.go`

Minimal end-to-end check:
- Construct `New` against a temp dir.
- Assert the file exists, the symlink resolves to it, and an emitted log line lands in the file.
- One test for cleanup: seed 60 fake `<timestamp>.log` files in a temp dir, call `New`, assert exactly 50 remain (49 old + the new one). No assertion on which 50 — `lex-sort` covers the predictability.

---

### Commit 2 — `feat(logging): add banner helpers`

**Recommended model:** Sonnet. Mechanical helpers; the design landed in Commit 1's handler.

#### Task 1 — `internal/logging/banner.go`

Three helpers, each one-liner bodies:

```go
func RunHeader(log *slog.Logger, version, campaign string, factions int) {
    log.Info("", "_banner", "header", "version", version, "campaign", campaign, "factions", factions)
}

func TurnStart(log *slog.Logger, turn int, faction string) {
    log.Info("", "_banner", "turn", "turn", turn, "faction", faction)
}

func PhaseStart(log *slog.Logger, phase string) {
    log.Info("", "_banner", "phase", "phase", phase)
}
```

Empty message strings are intentional — the handler renders the banner shape, not a message line. The `_banner` attribute is the dispatch key; the trailing attrs are the rendering payload.

#### Task 2 — `internal/logging/banner_test.go`

Golden-output tests for each helper, using a buffer-backed instance of the custom handler (reusing the Commit 1 handler's rendering). Asserts the multi-line block landed verbatim.

---

### Commit 3 — `feat(logging): plumb logger field into sub-engines and Engine`

**Recommended model:** Sonnet. Mechanical: a `*slog.Logger` parameter and a struct field per sub-engine, plus the testharness change.

No logger *usage* yet — this commit only adds the field and threads it through constructors. The end state compiles, all existing tests pass with `slog.New(slog.DiscardHandler)`, and no emission lines exist outside `internal/logging`'s own tests.

#### Task 1 — Sub-engine constructors (7 sub-engines)

Files and current signatures (verified at re-grounding time — confirm during the execution session):

- `internal/faction/engine/turn/turn.go` — `turn.New(rand, rulebook)` → `turn.New(rand, rulebook, log)`
- `internal/faction/engine/tag/tag.go` — `tag.New()` → `tag.New(log)`
- `internal/faction/engine/effect/effect.go` — `effect.New()` → `effect.New(log)`
- `internal/faction/engine/mutation/mutation.go` — `mutation.New()` → `mutation.New(log)`
- `internal/faction/engine/action/action.go` — `action.New()` → `action.New(log)`
- `internal/faction/engine/goal/goal.go` — `goal.New()` → `goal.New(log)`
- `internal/faction/engine/world/world.go` — `world.New(spatialDataDir)` → `world.New(spatialDataDir, log)`

Each sub-engine: add `log *slog.Logger` as the final constructor parameter, store it as `log` on the engine struct.

#### Task 2 — `internal/faction/engine/core.go`

- `func New(cfg *config.Config) (*Engine, error)` → `func New(cfg *config.Config, log *slog.Logger) (*Engine, error)`
- `func NewWithRulebook(rulebook *rulebook.Rulebook) *Engine` → `func NewWithRulebook(rulebook *rulebook.Rulebook, log *slog.Logger) *Engine`
- Add `log *slog.Logger` field to `Engine`.
- Thread `log` into each sub-engine constructor at composition time.

#### Task 3 — `internal/faction/engine/testharness/harness.go`

Sole production caller of `engine.NewWithRulebook`. Construct a discard logger inline:

```go
log := slog.New(slog.DiscardHandler)
eng := engine.NewWithRulebook(rb, log)
```

No nil-guards anywhere in the engine — the discard handler is cheap.

#### Task 4 — Test files for each sub-engine

Wherever sub-engine constructors are called in `*_test.go`, pass a discard logger. Compiler will surface every site; just thread the param. No assertion changes.

---

### Commit 4 — `feat(logging): cascade attrs and emit banners through orchestrator`

**Recommended model:** Opus. The cascade design is the substantive piece of the effort — bind points, parameter passing, and the corresponding banner emissions all need to compose cleanly.

#### Task 1 — `internal/faction/engine/orchestrator.go`: `RunCycle`

At the top of `RunCycle`:
- Bind the base logger with `version` if not already (decision: do it here, since `RunCycle` is the engine-level entry point that has visibility into version).
- Call `logging.RunHeader(e.log, version, campaign, factionCount)`. Source for these: version from build info or a constant; campaign from `cfg.FactionDataDir` or a derived label; faction count from the state set passed in.

Open at first-use: exact source of `version` (build-info ldflags vs a `version` package constant). Resolve at execution time.

#### Task 2 — `RunFactionTurn`

```go
turnLog := e.log.With("turn", turnNum, "faction", faction.ID)
logging.TurnStart(turnLog, turnNum, faction.ID)
// pass turnLog down to setupFactionTurn, run*Phase methods, finishFactionTurn
```

All phase methods take an additional `*slog.Logger` parameter (the turn-bound logger). Pass-through, not a struct field — each turn binds fresh attrs.

#### Task 3 — Each `run*Phase` method

Five methods: `runGoalLockPhase`, `runBookkeepingPhase`, `runStatRaisePhase`, `runMovementPhase`, `runActionPhase`. Pattern at the top of each:

```go
phaseLog := turnLog.With("phase", "<phase_name>")
logging.PhaseStart(phaseLog, "<phase_name>")
// use phaseLog for the rest of the method
```

Phase names: `goal_lock`, `bookkeeping`, `stat_raise`, `movement`, `action` (snake_case for multi-word, matches the canonical attr vocabulary).

#### Task 4 — Sub-engine delegation sites

Where the phase calls into a sub-engine (e.g., `e.Movement.TickMovementOrders(...)`, `e.Mutation.Apply(...)`), bind `engineLog := phaseLog.With("engine", "<engine_name>")` and pass it. Sub-engines that need to log use the bound logger.

This implies sub-engine methods grow a `log *slog.Logger` parameter for delegation calls. Alternative: store the *base* logger on the sub-engine field, and let the orchestrator pass a turn-bound logger per call. The latter is the pattern.

**Decision for this commit (ratified here):** Sub-engine *struct fields* hold the base logger (used for sub-engine-internal lines emitted without orchestrator context — startup, registration, etc.). Sub-engine *methods* called from the orchestrator take a `log *slog.Logger` parameter — the orchestrator passes the engine-bound logger per call. Two-channel design: per-call logger for cascaded context; field logger for context-free internal lines.

#### Task 5 — `setupFactionTurn` and `finishFactionTurn`

Take the `turnLog` parameter; emit Info-level entry/exit lines, no banner. Example:

```go
turnLog.Info("turn setup", "result", "ok")
// ...
turnLog.Info("turn finished", "outcome", outcome)
```

---

### Commit 5 — `feat(logging): seed Info/Debug emissions at phase and sub-engine boundaries`

**Recommended model:** Sonnet. Once the cascade is wired, adding the first emission lines is mechanical and limited in scope.

Goal: a small, deliberate first cut of Info / Debug lines so the log file has substantive content from day one. Not exhaustive. Subsequent execution sessions add more as needed.

#### Task 1 — Phase-boundary lines

Within each `run*Phase`:

- Top of method, after `PhaseStart` banner: `phaseLog.Info("phase begin")` (with whatever attrs are natural for that phase — e.g., movement gets `orders=N`, action gets `available=N`).
- Bottom of method, on success: `phaseLog.Info("phase end", "outcome", ...)`.

#### Task 2 — Sub-engine entry lines

For each sub-engine method called from the orchestrator (one or two primary entry points each), emit:

- `Info` on entry (the fact of, not the args).
- `Info` on exit with the operation outcome (e.g., `"mutations applied"`, `mutations=N`).

#### Task 3 — Debug seeds (selective)

A handful of Debug lines that pay off during typical debugging:

- `mutation.Apply` — per-mutation log at Debug level (`type`, `entity_id`).
- `world.TickMovementOrders` — per-asset transit at Debug level (`asset`, `from_hex`, `to_hex`).
- Roller — per-roll result at Debug (`die`, `result`).

Pick maybe 3-5 sites; not every code path. The seed cut is enough to demonstrate the pattern; expansion is opportunistic in later work.

#### Task 4 — Spot-check the rendered output

Manual: run the testharness with a real handler (not discard), inspect the log file, confirm:
- Run header renders correctly with version/campaign/factions.
- Turn and phase banners appear in the right order.
- Cascade attrs (`turn`, `faction`, `phase`) are suppressed from individual line rendering (per Commit 1's `_banner`-suppression list).
- Debug lines appear only with `--debug` flag set.

Not an automated test — a visual confirmation before declaring the effort complete.
