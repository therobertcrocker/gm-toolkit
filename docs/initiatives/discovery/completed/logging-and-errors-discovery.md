# Logging + Errors — Discovery

`F-003`. Discovery covers two threads — **logging** and **error handling** — held in a single doc because they share design questions about how problems carry context (chained `error` wrapping ≈ structured log attrs; the surface choice for one constrains the shape of the other).

This discovery is run across two sessions:
- **Session 1** — logging thread.
- **Session 2** — error audit + conventions.

## Guiding Principles

Carried forward from `planned-work.md`:

- **Consistency** — one pattern for errors, one pattern for logging, everywhere.
- **Ease of use** — should not require ceremony; the right behavior should be the default behavior.
- **Structured over vague** — errors carry typed context; log output is human readable but also machine-parseable; avoid unstructured strings where possible.
- **Surface over silence** — when something goes wrong, it should be visible in a log, not just an error return that may or may not be checked.

## Problem

No structured logging exists. `main.go` has a single `fmt.Println` for usage output; nothing else reaches a log surface. Error handling across 25+ files is ad hoc: some packages define sentinel errors (`spatial`, `engine/core`), others inline `fmt.Errorf` strings with no consistency. `TurnObserver.OnError` accepts a raw `error` — there is no type information, no severity, and no structured context attached. When something goes wrong, there is no trail and no way to distinguish a recoverable data-load failure from a logic bug.

The TUI Rebuild (`F-005`) is the next major initiative. The TUI needs a log surface decoupled from display output. Designing logging now without TUI integration in mind risks a retrofit when F-005 lands.

## Design Summary

**Logging (Session 1 — complete).**

`log/slog` (stdlib) is the backend. Output is **file-only**: per-run timestamped files in `logs/` (e.g. `logs/2026-05-23_143045.log`), plus a `logs/latest.log` symlink. Count-based cleanup at startup keeps the last 50 runs. A custom ~50-LOC `slog.Handler` renders human-readable lines with aligned columns and short `HH:MM:SS` timestamps; full-date metadata lives in a per-file header. Banner content lives in `pkg/logging/banner.go` and is emitted from the orchestrator at well-defined hook points using a `_banner` attribute pattern; the handler renders banners as visual separators.

Levels are slog defaults — `Debug` / `Info` / `Warn` / `Error`, no custom levels. **Info is the engine's view of its own behavior** (sub-engine activity, phase transitions, validation outcomes), distinct from the narrative / history log which already covers faction-level events. `--debug` lowers to Debug and enables `AddSource` for file:line. Fatal events are folded into Error + `os.Exit(1)`.

Loggers are plumbed as **sub-engine struct fields** (and free-function params where needed) — no `context.Context` carrier, no global default. The orchestrator cascades `.With(...)` calls down the layer stack: `version → turn → faction → phase → engine → event-specific attrs`. Every log line carries its full context automatically; call sites add only locally-novel attrs.

The public API of `pkg/logging` exposes `*slog.Logger`, not the handler. This leaves a clean path for `F-005` (TUI Rebuild) to wrap the file handler in a multi-handler fan-out and add a TUI sink without touching engine code. No fan-out is pre-built.

**Errors (Session 2 — complete).**

The codebase has three error shapes today — sentinels (8 declared), one typed error (`MutationApplyError`), and 107 `fmt.Errorf` sites — with no convention enforced between them and no `errors.Is`/`errors.As` discipline in production code. Session 2 ratifies a convention layer, a recoverability mechanism, and a logging interplay rule.

**Conventions.** Sentinels are declared for stable, package-recognized failure modes (legibility criterion, not "callers must branch"). Typed errors are declared only when a named consumer reads at least one field — stricter, to avoid `MutationApplyError`'s "structured payload nobody unwraps" trap. `fmt.Errorf` standardizes: `%w` mandatory when wrapping; the `%w`-of-sentinel pattern (currently only in `spatial`) becomes the default for returning a known kind with situation-specific detail; bare-string `fmt.Errorf`s are banned and promoted to sentinels (`"world engine not found"` is the canonical case). Prefix discipline is light — operation/function-name prefix preferred, no required format, package author's call on whether sentinel messages carry a package-ish scope.

**Recoverable vs Fatal.** The orchestrator branches on error identity at each sub-engine call: `isRecoverable(err)` checks an explicit, named list of recoverable sentinels and typed cases; everything else is fatal-by-default. Recoverable → `logger.Warn` and continue; fatal → `logger.Error` and propagate up. The initial recoverable list is small (action-precondition failures and collector "no selection" returns, sentinels to be declared during Plan). `MutationApplyError` is **fatal** — its structured payload (`Misses`) drives per-miss `Error`-level log emissions for diagnostic visibility but does not gate recovery.

**`TurnObserver.OnError`.** Kept as-is for F-003 — the redesign is deferred to F-005 (TUI Rebuild). The predicted future shape (recorded so it doesn't get re-debated) is **Shape B**: replace `OnError(faction, err)` with semantic gameplay events keyed to recoverable category (`OnActionPrecluded(faction, action, reason)` or similar). Recoverable errors are gameplay outcomes from the player's perspective, not errors; fatal errors leave the observer entirely and live on the logger fan-out plus the orchestrator's return path. Designing the exact signature requires F-005 to have shape.

**Logging / Error Interplay.** Every returned error is logged **exactly once**, at the orchestrator's Recoverable/Fatal branch. Sub-engines return errors silently — no pre-log at the return site. Wraps (`fmt.Errorf("prefix: %w", err)`) add display context, not emission. Sub-engines may log freely about their own activity (Session 1's domain) but not about their returned errors. Typed errors with structured payloads (`MutationApplyError`) expand to multiple log lines at the orchestrator as part of the single logging step, not as duplicated emission. The framing that makes the rule stick: wraps *label* errors for eventual rendering; logging *announces* them — and announcement is the orchestrator's job.

---

# Logging

## Backend

**Decision: `log/slog` (Go stdlib).**

Rationale:
- Stdlib — no dependency churn, ships with the toolchain.
- Structured by design — `slog.Info("msg", "key", value)` produces machine-parseable output without ceremony.
- Handler-pluggable — the `slog.Handler` interface lets us swap output destinations (JSON, text, custom TUI router) without touching call sites. This is the load-bearing property for `F-005`.
- Context-aware — `slog.With(...)` binds attributes (faction, turn, phase) once and they travel through nested calls.

Alternatives considered and rejected:
- `zerolog` / `zap` — faster and more allocation-conscious; irrelevant at single-user-tool scale. Stdlib wins on consistency.
- A custom thin wrapper — slog already *is* that wrapper. Re-wrapping adds ceremony rather than removing it.
- `charmbracelet/log` — same vendor as our TUI stack, but it's a slog handler rather than a replacement; would lock us into their evolution for no clear gain.

## Log Surface

**Decisions:**

- **Destination: file only.** No stderr output. TUI mode (`F-005`) would be corrupted by stray stderr writes, and a clean terminal is preferable in CLI mode too. The log file is the single source of truth for debugging.
- **Location: `logs/` in the project root**, gitignored. Co-located with campaign state for visibility; not stashed in an XDG state directory the GM would forget about.
- **File-split: per-run, timestamped + `latest.log` symlink.** Each program run produces `logs/<timestamp>.log` (e.g. `logs/2026-05-23_143045.log`); `logs/latest.log` is a symlink to the most recent. Each file maps to one coherent narrative (start → exit), eliminating the "single growing file is messy" problem.
- **Cleanup: count-based, at startup.** Keep the last 50 runs; drop the rest. List `logs/*.log`, sort lexicographically (timestamp-prefixed names sort chronologically for free), delete the oldest beyond the cap. No date parsing, no separate command, no flag clutter. If 50 ever feels wrong, it's a one-constant change.
- **Format: custom `slog.Handler` (~50 LOC).** `TextHandler`'s `key=value` mush doesn't scan well at debugging speed; the custom handler emits aligned columns: short `HH:MM:SS` timestamp, level, source path, message, then attrs trailing. Full date lives in the file header at the top of each run.
- **File header.** First lines of each run file capture run metadata: version, campaign file path, faction count, full start timestamp. Spares grepping for context.
- **Banners — `_banner` attribute pattern.** Banner *content* lives in `pkg/logging/banner.go` (e.g. `TurnStart(turn, faction)`, `PhaseStart(phase)`). The orchestrator emits banners by calling those helpers at well-defined hook points — one line per call, no formatting in the orchestrator. The custom handler watches for the `_banner` attribute and renders a visual separator instead of an ordinary log line. Same `slog.Logger` does both ordinary lines and banners; no new logging surface.

Example output:

```
=== faction-manager v0.X.Y ===
campaign:   campaigns/whispering-hand/
factions:   4
started:    2026-05-23 14:30:45

  ┌──────────────────────────────────────────────────────────────────┐
  │  Turn 14  ·  Cabal of the Whispering Hand                        │
  └──────────────────────────────────────────────────────────────────┘

  --- Movement Phase ---
14:30:45  INFO   movement.start     phase begin            faction=cabal
14:30:45  DEBUG  movement.transit   asset moving           asset=A-003 from=hex-12 to=hex-15
14:30:46  INFO   movement.complete  phase end              moved=3 stayed=2
```

## Level Scheme

**Decisions:**

- **Four levels — slog defaults: `Debug`, `Info`, `Warn`, `Error`.** No custom `Trace` or `Fatal`. YAGNI.
- **No Fatal level.** When something is un-recoverable, emit an `Error` log line and `os.Exit(1)`. Exit is control flow, not a log concern.
- **Default level: `Info`.** `--debug` lowers to `Debug` *and* enables `slog.HandlerOptions.AddSource` for file:line. Single flag, two effects, intentional.

**Calibration — what each level captures.**

Logging is **the engine's view of its own behavior**, not the GM's view of faction-level events. The latter is already covered by the narrative / history log. The two outputs serve different audiences and answer different questions:

- **History log** answers: "what did the factions do?"
- **Log file** answers: "what did the engine do, and why?"

This split keeps both useful. The log file is not a duplicate of history; it's the troubleshooting surface.

| Level | What it captures | Examples |
|-------|------------------|----------|
| `Debug` | Thorough internal detail — everything you'd want when troubleshooting a misbehavior. Off by default. | Per-roll dice values; intermediate calc steps; mutation diffs; individual asset transit hops; cache hit/miss; per-attribute state dumps. |
| `Info` | Normal-verbosity engine internals — sub-engine activity, phase boundaries, decision points, validation outcomes. The default file content. | Sub-engine started / completed; phase entered / exited; validation passed; goal-progress evaluated; mutation applied (fact of, not diff); banner emissions. |
| `Warn` | Expected anomaly handled gracefully — a thing we designed a fallback for, fired. | Optional TOML field missing → default used; deprecated config key encountered; ability narration missing → fallback rendered; recoverable validation skip. |
| `Error` | Something tried and failed, even if execution continues. Always worth reading. | Action validation failure; mutation rejected by downstream subsystem; required data missing → operation aborted; unexpected internal state. |

**Warn-vs-Error rule (the load-bearing boundary):**

> **Warn** — the engine *expected this could happen* and handled it gracefully (anomaly we designed for).
> **Error** — the engine *tried something and it failed* (a thing went wrong, even if recovery exists).

Routine fallbacks at `Error` make `Error` lose its "look here" signal; genuine failures at `Warn` hide problems. This boundary is where the file's debugging value lives.

## Structured Attributes

**Decisions:**

- **Plumbing: sub-engine field + free-function param.** Sub-engines (and the orchestrator) carry `logger *slog.Logger` as a struct field, set at construction. Free helper functions that need to log take `*slog.Logger` as a parameter. No `context.Context` carrier (this engine doesn't deal with cancellation or deadlines — adding `ctx` purely to carry a logger is a misuse). No global default — bound attrs leaking to unrelated call sites defeats the point of structured logging.
- **Cascading binding model.** Each layer adds the attrs it knows about by calling `.With(...)` to produce a context-bound child logger and passing that down:

  | Layer | Binds | Example attrs added |
  |-------|-------|---------------------|
  | Base (startup) | `version` | `version=0.X.Y` |
  | Orchestrator (per turn) | `turn`, `faction` | `turn=14 faction=cabal` |
  | Orchestrator (per phase) | `phase` | `phase=movement` |
  | Sub-engine (at delegation) | `engine` | `engine=movement` |
  | Specific event | event-specific attrs | `asset=A-003 from=hex-12 to=hex-15` |

  A call site adds only the attrs locally novel; the context is inherited from the bound logger handed in. This is the mechanism that makes "ease of use — right behavior is the default" tractable for logging: the call site that doesn't know it's in turn 14 doesn't have to know.

- **Attr key naming: lowercase, single-word where possible.** `faction`, `turn`, `phase`, `engine`, `asset`. Compound keys use snake_case (e.g. `from_hex`, `to_hex`). Avoid `factionID` / `faction_id` redundancy when context makes it unambiguous.

**Note on handler interaction.** The custom handler can see the cascade as discrete attrs and choose to suppress already-banner-displayed attrs (`turn`, `faction`, `phase`) from individual line rendering, keeping the file scannable without losing structured data. Same log file: human-readable lines, machine-parseable attrs underneath.

## TUI Seam

**Decisions:**

- **No globals.** The base `*slog.Logger` is constructed once at startup (`cmd/faction-manager/main.go`) and passed into the orchestrator. No `slog.SetDefault`, no package-level loggers.
- **Public API exposes `*slog.Logger`, not `slog.Handler`.** The handler is an internal detail of `pkg/logging`. Callers depend only on `*slog.Logger`.
- **Defer fan-out construction to F-005.** Today the base logger has one handler (the custom file handler). When `F-005` lands and the TUI needs to receive log events, it'll wrap the file handler in a fan-out and add a second sink. Estimated cost when that day comes: ~5 minutes at logger-construction time, ~20 LOC for the fan-out handler. No reason to pre-build it.

**The seam is implicit in slog's design.** A `slog.Logger` wraps a `slog.Handler`; a `slog.Handler` can dispatch to *N* sub-handlers. F-005 plugs into this layer without touching the engine. The two decisions above (no globals, expose Logger not Handler) are what keep that path open.

**Anti-pattern to avoid: wrapping slog in a project `Logger` interface.** slog *is* the abstraction. Wrapping it adds ceremony, makes call sites less idiomatic, and pretends we might switch backends. The handler-swap seam is enough.

---

# Errors

## Audit Findings

Factual survey of the current error-handling landscape. Patterns and divergences only — preferences live in [Conventions](#conventions).

### Sentinel errors

**8 sentinels across 3 packages**, all declared as `var ErrFoo = errors.New("...")` in package-level `var (...)` blocks.

| Package | File | Sentinels |
|---------|------|-----------|
| `spatial` | `internal/spatial/spatial.go:24-27` | `ErrUnknownWorld`, `ErrNoPath`, `ErrNotImplemented`, `ErrInvalidCost` |
| `engine/turn` | `internal/faction/engine/turn/turn.go:12-14` | `ErrTurnInProgress`, `ErrNoTurnActive`, `ErrNoFactions` |
| `engine` (root) | `internal/faction/engine/core.go:21` | `ErrSpatialDataDirRequired` |

Naming is consistent (`Err` prefix, descriptive suffix). Message strings vary in shape — `spatial` prefixes with the package name (`"spatial: unknown world"`); `turn` and `engine` use plain phrases (`"a turn is already in progress"`).

### `errors.Is` / `errors.As` discipline

| Check | Production sites | Test sites |
|-------|------------------|------------|
| `errors.Is` | 1 (against stdlib `os.ErrNotExist` in `state/faction_state.go:23`) | 2 (against `spatial` sentinels in `region_map_test.go`) |
| `errors.As` | 0 | 0 |

**No production code does `errors.Is` against any project-defined sentinel.** The 7 project sentinels are declared, returned, and propagated, but their identity is only inspected by spatial's own tests. `errors.As` is used nowhere.

### Ad-hoc one-off errors

One outlier dodges both patterns:

- `engine/turn/turn.go:63` — `errors.New("faction not found: " + id)`. String concatenation into `errors.New`, no sentinel, no `fmt.Errorf`. The only place in the codebase using this shape.

### `fmt.Errorf` survey

**107 call sites across 22 files** (non-test). Density (top 5):

| File | Count |
|------|-------|
| `rulebook/rulebook.go` | 22 |
| `engine/orchestrator.go` | 13 |
| `spatial/region_map.go` | 11 |
| `engine/world/movement.go` | 10 |
| `engine/action/actions/expand_influence.go` | 10 |

**Wrapping discipline:**

| Pattern | Count | Share |
|---------|-------|-------|
| Uses `%w` (preserves chain) | 59 | ~55% |
| No `%w` (forms a new string, breaks chain) | 48 | ~45% |
| Bare string, no format verbs (effectively `errors.New`) | 6 | ~6% |

The 48 non-`%w` sites are predominantly **leaf origin** errors that have no upstream to wrap — `fmt.Errorf("unknown asset definition %q", id)`, `fmt.Errorf("expand influence: insufficient Coin: need %d, have %d", cost, coin)`. The format verbs there are attaching dynamic context, not formatting another error.

**Bare-string sites** (no format verbs, no `%w`, no dynamic context) — these are `errors.New` equivalents written as `fmt.Errorf`:

- `engine/orchestrator.go:211` and `:290` — both `fmt.Errorf("world engine not found")`. **Literal duplication.** Neither calls `observer.OnError` before returning.
- `engine/action/actions/change_homeworld.go:52` — `"change homeworld: no target selected"`
- `engine/action/actions/seize_planet.go:47` — `"seize planet: no target world selected"`
- `engine/action/actions/sell_asset.go:38` — `"sell asset: no asset selected"`
- `engine/action/actions/ability/informers.go:28` — `"informers: no target faction selected"`

**Prefix-as-context pattern** is widely used at action boundaries: `fmt.Errorf("expand influence: %w", err)`, `fmt.Errorf("attack: %w", err)`. Identifies the call-stack origin at wrap time. No standard for choice of prefix string (action name vs. package name vs. operation name).

**Outlier — sentinel + dynamic context in one wrap:** `spatial/region_map.go:246, 258, 265, 281` use `fmt.Errorf("%w: crossingCost=%d", ErrInvalidCost, ...)` — sentinel identity preserved for `errors.Is`, plus structured context for human reading. This pattern appears **only in `region_map.go`**.

### Typed errors

**Exactly one typed error in the codebase.**

- `engine/mutation/mutation.go:21-36` — `MutationApplyError struct { Misses []MutationMiss }`. Each `MutationMiss` has `MutationType`, `FactionID`, `EntityID`. The `Error()` method renders a multi-line summary.
- Returned by `(*MutationEngine).Apply` as **concrete pointer type** (`*MutationApplyError`), not the `error` interface. Returns `nil` on success at `mutation.go:362`, `&errs` on miss at `:364`.
- Caller `orchestrator.applyAndRecord` (`orchestrator.go:359-361`) does `if err := e.Mutation.Apply(...); err != nil { return fmt.Errorf("applying mutations: %w", err) }`. Pointer-nil check is correct because `err` is typed as `*MutationApplyError`, not `error` — the typed-nil-into-interface gotcha is dodged.
- The wrap with `%w` preserves the typed payload in the chain. **No caller anywhere uses `errors.As` to recover it.** The `Misses` slice is constructed on every miss but never read by any consumer.

### `TurnObserver.OnError`

**Definition** — `engine/observer.go:24`:

```go
OnError(faction *domain.Faction, err error)
```

Raw `error`. No severity, no typed envelope, no structured context.

**Implementors — 1** (LSP `goToImplementation`): `testharness/RecordingObserver.OnError` at `testharness/observer.go:80`. Body is a one-liner: `r.Events = append(r.Events, ObservedEvent{Kind: "Error", Faction: faction, Payload: err})`. Nothing acts on the error; it's recorded into a slice for test assertions.

**Call sites — 26**, all in `engine/orchestrator.go`. The dominant pattern across all 26:

```go
if err := someSubEngine(...); err != nil {
    observer.OnError(faction, err)
    return err   // or: return fmt.Errorf("context: %w", err)
}
```

Three structural observations:

1. **Double-duty error surfacing.** The orchestrator both calls `OnError` (side-channel) *and* returns the error up the stack. Every error path does both. With Session 1's logger plumbed through the orchestrator, the side-channel collapses to a logger-call equivalent — but the testharness consumer captures the event into a slice for assertions, which a logger sink doesn't naturally do.

2. **Observer / caller divergence at wrap sites.** A handful of sites pass *raw* `err` to `OnError` but a *wrapped* error to the caller — e.g. `orchestrator.go:79-81`:
   ```go
   if err := state.Save(...); err != nil {
       observer.OnError(faction, err)                         // raw
       return false, fmt.Errorf("saving state after bookkeeping: %w", err)  // wrapped
   }
   ```
   The observer sees the inner cause; the orchestrator's caller sees the outer prefix. Two views of the same failure, both meaningful, neither aware of the other.

3. **Inconsistent surfacing.** The two `"world engine not found"` sites (`orchestrator.go:211, 290`) return errors **without** calling `OnError`. Every other error path in `orchestrator.go` calls `OnError`. These two skip it.

`nil` is a legal `faction` argument — passed at `setupFactionTurn` sites (`:107, :116`) where the error occurs before a faction is selected.

### Summary of divergences

- Sentinels exist but their identity is unused outside `spatial` tests. The "use `errors.Is` to check kind" half of the sentinel contract isn't being honored.
- ~45% of `fmt.Errorf` sites break the wrap chain. Most are leaf-origin, but some are mid-stack and would benefit from `%w`.
- Six `errors.New`-equivalent `fmt.Errorf` calls, including one literal duplicate across two orchestrator sites.
- The one typed error in the codebase is unrecoverable in practice — its structured payload is built but never read.
- `OnError` surfacing is inconsistent at two orchestrator sites; everywhere else the pattern is mechanical and duplicated 24 times.
- The sentinel-plus-dynamic-context wrap pattern (`%w` of sentinel + structured trailing context) exists in exactly one file and is the most expressive shape in the codebase, but isn't used elsewhere.

## Conventions

### Sentinels — when to declare

**Decision: name any stable, package-recognized failure mode as a sentinel.**

The criterion is *legibility*, not *branching*. A `var (...)` block declaring a package's `ErrFoo`s documents the failure vocabulary that package surfaces — readable in one glance, available as a future target for `errors.Is`, and a single source of truth when the same error is returned from multiple sites.

The trigger to declare a sentinel:

- **Strong** — the error is returned from two or more sites (DRY); identity check is on the table even if not used yet.
- **Sufficient** — the error names a known failure mode of the package, even if returned from one site (documentation).
- **Required** — the bare-string `fmt.Errorf` form would duplicate a string across multiple sites (the `"world engine not found"` case).

The stdlib follows the same shape: `io.EOF`, `os.ErrNotExist`, `sql.ErrNoRows`, `context.Canceled` exist so the package's failure surface is named, not because every caller branches.

**At gm-toolkit's scale** (8 sentinels today, expected to reach ~12-15) the "named vocabulary" tax is negligible; the cost-benefit flips at perhaps 50+ sentinels, where unused identities become a maintenance load.

### Decisions on the current inventory

**Keep all 8 existing sentinels.** They all earn their place under the legibility criterion:

| Sentinel | Why it earns it |
|----------|-----------------|
| `spatial.ErrUnknownWorld`, `ErrNoPath`, `ErrInvalidCost` | DRY (multiple return sites) + checked by tests |
| `spatial.ErrNotImplemented` | Documentation — signals known-incomplete state |
| `turn.ErrNoTurnActive` | DRY (returned from `turn.go:57` and `:70`) |
| `turn.ErrTurnInProgress`, `turn.ErrNoFactions`, `engine.ErrSpatialDataDirRequired` | Documentation — well-known refusal modes |

**Promote `"world engine not found"` to a sentinel.** Currently duplicated as a bare-string `fmt.Errorf` at `orchestrator.go:211` and `:290`. Working name `ErrWorldEngineUnavailable` (subject to final naming at implementation time). This is the strongest promotion candidate the audit surfaced.

**Defer the "no target selected" action-error cluster** (`change_homeworld.go:52`, `seize_planet.go:47`, `sell_asset.go:38`, `informers.go:28`). Whether these should be sentinels depends on whether the orchestrator needs to distinguish "user/collector aborted a required selection" from other action failures — a Recoverable-vs-Fatal question, not a naming question. Revisit in that section.

### Typed errors — when to declare

**Decision: declare a typed error only when a consumer is named at design time and will read at least one field.**

This bar is intentionally stricter than the sentinel rule. Sentinels cost one `var` declaration; typed errors carry an `Error()` method, fields, and an implicit contract with the consumer. "Future maintainers might want to" is not a consumer. A typed error nobody unwraps is a `fmt.Errorf` with extra steps.

The consumer test:

- ✅ Named at design time (specific caller / package / observer)
- ✅ Reads at least one field (via `errors.As`, or by being the declared return type)
- ❌ "Decorative" structured payloads where the consumer just stringifies the error

#### `MutationApplyError` — decision

**Keep, with a named consumer.** Today the type is decorative — built and stringified. The fix is to wire up a real consumer:

- **Consumer:** orchestrator's `applyAndRecord` (`orchestrator.go:359-361`).
- **Pattern:** on mutation Apply failure, `errors.As(err, &mutationErr)`, then iterate `mutationErr.Misses` and log one structured line per miss using Session 1's slog logger. Attrs map directly: `mutation_type`, `faction_id`, `entity_id`. The level (Warn vs Error) firms up in [Recoverable vs Fatal](#recoverable-vs-fatal).
- After logging, wrap and return up the stack as today (`fmt.Errorf("applying mutations: %w", err)`).

The structured payload becomes the bridge between mutation-engine reality and the logger's attr-cascade model. The aspirational comment at `mutation.go:18-20` ("the orchestrator decides whether to treat this as a hard turn failure or a warning") finally gets a place where the decision happens.

#### Return-type signature for typed errors

**Decision: leaf functions returning a single typed error type may return the concrete pointer type** (e.g. `*MutationApplyError`); orchestrating callers wrap with `%w` into `error`.

`MutationEngine.Apply` keeps `*MutationApplyError` as its return type. Normalizing to `error` would expose the typed-nil-into-interface gotcha — `var errs *MutationApplyError; ... return errs` would produce a non-nil interface wrapping a nil pointer, breaking `err != nil` checks at the call site. The concrete-pointer return dodges that and is the correct idiom for "leaf returns exactly one error type."

Callers wrap with `fmt.Errorf("...: %w", err)` to promote to `error` for upstream propagation — at that point the value is non-nil and the gotcha doesn't bite.

### Wrap pattern — `fmt.Errorf` shapes

Three shapes earn explicit support; one is banned.

| Situation | Pattern | Notes |
|-----------|---------|-------|
| Wrapping an upstream `err` | `fmt.Errorf("prefix: %w", err)` | **`%w` is mandatory.** Never use `%v` for an error — silently drops chain participation. |
| Wrapping with a sentinel kind *and* situation-specific detail | `fmt.Errorf("%w: detail %d", ErrFoo, arg)` | The `spatial/region_map.go` pattern. Becomes the default when returning a known kind with dynamic context — preserves `errors.Is` checkability and human readability in one wrap. |
| Wrapping an upstream `err` *and* tagging with a sentinel kind | `fmt.Errorf("%w: prefix: %w", ErrFoo, err)` | Go 1.20+ allows multiple `%w` per call. Not used today; convention allows it for cases where both the kind and the cause matter. |
| Leaf origin (no upstream `err`, dynamic context) | `fmt.Errorf("prefix: bad input %q", arg)` | Legitimate non-`%w` use — format verbs carry context, not chain. The bulk of current non-`%w` sites are this shape. |
| Leaf origin (no upstream `err`, no dynamic context) | **Banned.** Use a sentinel instead. | The audit's six bare-string sites are missing sentinels by definition. |

#### Spot-check finding

48 non-`%w` `fmt.Errorf` sites were audited. Result: **zero mid-stack chain-drops.** All 42 dynamic-context sites are originating new errors (lookup, validation, parse, or structural failure). The remaining 6 are the bare-string cluster covered above. The convention is already mostly followed; Plan does not need to budget for retroactive wrap cleanup.

#### Rules

1. **`%w`, not `%v`, when wrapping an error.** Drift here is the single most common form of "errors are working in name only" — chains that look fine in source but break the moment a caller tries `errors.Is`/`errors.As`.
2. **The `%w`-of-sentinel pattern is the default for returning a known kind with situation-specific detail.** Use it freely; it's currently underused and is the most expressive shape in the codebase.
3. **No bare-string `fmt.Errorf`s.** A static string with no `%w` and no format verbs is a sentinel that hasn't been declared yet.
4. **Multiple `%w` per call is allowed** (Go 1.20+) when both the kind and the cause carry information.

### Prefix discipline

**Decision: a light rule, not a strict format.** The prefix's job is to orient a reader to *what was happening when this failed*.

Three guidelines, in priority order:

1. **Prefer operation / function name over action / package name.** `"selecting movement decisions: %w"` carries more local context than `"orchestrator: %w"`. The operation name is the most useful orientation.
2. **Don't double-prefix.** If a higher wrap adds `"expand influence: "`, the inner error should not repeat it. Avoids drift like `"expand influence: expand influence: insufficient Coin"`.
3. **Sentinel messages may optionally include a package-ish scope** (`"spatial: unknown world"`). Optional for standalone readability when the sentinel is logged or printed without surrounding wrap context. If used, be consistent within the package.

Not in scope:
- No required prefix format (e.g. `"action: "` vs `"[action] "`).
- No requirement that every wrap carry a prefix; sometimes the upstream is informative enough alone.

The audit shows variation but no hygiene problem — every individual prefix today is informative. A strict format rule would force churn for marginal readability gain.

## `TurnObserver.OnError`

**Decision: keep `OnError(faction *domain.Faction, err error)` as-is for F-003. Plan to redesign it as a semantic gameplay event when `F-005` (TUI Rebuild) lands.**

### Why not delete it now

The initial reading favored deletion: 26 call sites in `orchestrator.go`, one implementor (`testharness.RecordingObserver`), zero test assertions on the `"Error"` event kind. With Session 1's logger plumbed in and Session 2's Recoverable/Fatal mechanism deciding severity at the orchestrator, the observer hook looked redundant.

That reading underweighted what the observer interface *is*. The other callbacks — `OnFactionTurnStarted`, `OnGoalLockApplied`, `OnActionResolved` — aren't debug surfaces; they're **gameplay-event hooks** intended to drive the future TUI's turn-by-turn presentation. Errors split cleanly along the Recoverable/Fatal boundary:

| Error type | Nature | Right home |
|------------|--------|------------|
| **Recoverable** (action precluded, no selection, precondition failed) | Gameplay event — "this faction tried X, it didn't happen, here's why" | Observer — TUI surface |
| **Fatal** (state.Save failed, mutation refs missing, world engine unavailable) | System event — engine integrity broken | Logger fan-out (Session 1 seam) + return-path to CLI/TUI driver for the "engine died" surface |

The recoverable cluster *is* gameplay-shaped. Deleting `OnError` would either lose that signal or force the TUI to recover it by string-matching the logger output — both bad outcomes.

### Status quo for F-003

`OnError` stays exactly as it is today, fired on both recoverable and fatal branches (the orchestrator calls it before logging, then either continues or returns per the Recoverable/Fatal mechanism). `RecordingObserver` continues to record `"Error"` events to its slice; tests that don't currently assert on them remain free not to. No call-site changes beyond what Recoverable/Fatal already mandates.

This is **deferral, not endorsement.** The current shape is preserved because designing the right replacement requires the TUI to exist — F-005 will know what it needs from error events in a way today's design cannot.

### Predicted future shape — Shape B

When `F-005` lands and the TUI's error-handling needs are concrete, the expected redesign is to **replace `OnError` with a semantic gameplay event** keyed to the recoverable category:

```go
// Illustrative — actual signature settled at F-005 design time
OnActionPrecluded(faction *domain.Faction, action action.Action, reason error)
```

Why this is the predicted direction:

- **Inversion of framing.** Recoverable errors aren't *errors* from the player's perspective — they're *outcomes*. A faction tried something; it didn't happen; here's why. That's a gameplay event, not an error surface.
- **Matches observer interface flavor.** Every other `On*` method on `TurnObserver` describes a positive event ("X happened"). `OnError` is the only negation-shaped hook.
- **Fatal errors leave the observer entirely.** Engine-integrity failures aren't gameplay; they belong on the logger fan-out (Session 1) and the return path. The observer interface stays cleanly gameplay-only.
- **TUI rendering aligns with semantics.** A toast saying "Cabal can't expand influence — insufficient Coin" is a different rendering than an "engine error" overlay. The semantic event makes that split natural; a generic `OnError(err)` blurs it.

Shapes considered and rejected for F-005's eventual redesign (recorded so they don't get re-debated):

- **Shape A — narrow `OnError` to recoverable only, same signature.** Light change, but loses the gameplay-event flavor and forces the TUI to inspect `err` to know what kind of "didn't happen."
- **Shape C — structured `ErrorEvent` envelope.** Generic and over-engineered before F-005 has shape; the structure would inevitably miss the right fields.

The status-quo decision parks all of this until F-005 is in flight. The orchestrator's call-site changes from Recoverable/Fatal still apply (logger.Warn / logger.Error + branch); only the observer surface defers.

## Recoverable vs Fatal

**Decision: sentinel-driven recoverability with an explicit list, plus `errors.As` for typed cases.**

Today every error in `orchestrator.go` propagates up the stack — there is no "recover and continue" path. This section designs the mechanism that distinguishes the two outcomes, mapping directly onto Session 1's `Warn` (anomaly we designed for) and `Error` (something tried and failed) boundary.

### Mechanism

The orchestrator branches on error identity at each sub-engine call:

```go
if err := someSubEngine(...); err != nil {
    if isRecoverable(err) {
        logger.Warn("recoverable phase failure", "phase", phase, "err", err)
        // surface to observer per OnError decision (see next section)
        return nil  // or: continue this phase without aborting the turn
    }
    logger.Error("phase failed", "phase", phase, "err", err)
    return err  // propagate — abort the turn
}

func isRecoverable(err error) bool {
    for _, sentinel := range recoverableErrors {
        if errors.Is(err, sentinel) { return true }
    }
    return false
}
```

Recoverability is a property of error *identity*, not of wrap-site choice. The `recoverableErrors` list is a small, named, explicit constant. Adding to it is a deliberate decision.

Alternatives rejected:

- **Severity carried on the error** (a `SeverityError` interface/wrap). Disperses the recoverability decision across every wrap site instead of centralizing it. Adds machinery for a problem that doesn't exist at this scale.
- **Orchestrator decides by context** (which sub-engine, which phase). Brittle — recoverability lives in the wrong place; the orchestrator has to know context-specific rules for every caller.
- **Implicit by sentinel, no named list.** Same shape as the chosen approach but without the discipline of a finite, audit-friendly list. Every new error category becomes an implicit branch needing manual auditing.

### Recoverable list — initial contents

| Category | Sentinel / Type | Why recoverable |
|----------|----------------|-----------------|
| User / collector aborted a required selection | New sentinel — likely `action.ErrNoSelection` (covering the four "no target/asset/world/faction selected" cases in the audit) | UI flow, not engine failure. Skip the action, continue the turn. |
| Action precondition failed (insufficient Coin, missing target, unknown sub-mode, etc.) | New sentinel — likely `action.ErrPreconditionFailed` | The action just can't run. Skip and continue. |

The exact shape (one sentinel for the cluster vs. per-action sentinels) is a Plan-time call; the audit suggests the semantics are unified enough that one sentinel for "no selection" and one for "precondition failed" will cover the existing 4+ sites cleanly.

Everything else is **fatal** by default:

- `state.Save` failures
- `world.TickMovementOrders` failures
- `applyAndRecord` failures
- `Phase.AwaitCheckpoint` failures
- `turn` engine sentinels (`ErrTurnInProgress`, `ErrNoTurnActive`, `ErrNoFactions`) when surfaced to the orchestrator
- `spatial` errors propagated to the orchestrator (`ErrNoPath`, `ErrUnknownWorld`, `ErrInvalidCost`)
- `engine.ErrWorldEngineUnavailable` (the promoted bare-string)

### `MutationApplyError` — decision

**Fatal.** Missing entity references during mutation Apply indicate engine-or-data integrity is broken; continuing past it makes the underlying bug harder to root-cause.

The diagnostic value of the typed payload is preserved by the consumer wired up in [Typed errors](#typed-errors--when-to-declare): the orchestrator catches the error with `errors.As`, logs **one structured `Error`-level line per miss** (attrs: `mutation_type`, `faction_id`, `entity_id`), then propagates up to abort the turn. The structured payload is for *diagnostic visibility*, not for resumption.

This resolves the aspirational comment at `mutation.go:18-20` — the orchestrator's "decision" is "log and abort."

## Logging / Error Interplay

**Decision: every returned error is logged exactly once, at the orchestrator's Recoverable/Fatal branch.** Sub-engines do not log their own returned errors. The chain of `%w` wraps adds context for *display* of the error, not for additional emission.

### The invariant

> One returned error → one log line at the orchestrator boundary. Sub-engines log their own activity (any level) but not their returned errors. Wraps add context for display, not for emission.

### Rules

1. **Sub-engines return errors silently.** A sub-engine that encounters an error and returns it does **not** log before returning. The orchestrator is the unique logging point for returned errors. Matches `feedback_error_return_over_observer`.
2. **Wrap sites don't log.** `fmt.Errorf("prefix: %w", err)` adds context to the chain; it does not emit a log line. If every wrap-and-return logged, a single error would produce N log lines as it bubbles up.
3. **The orchestrator is the unique error-logging point.** Single emission at the Recoverable/Fatal branch. Orchestrator-internal helpers (`applyAndRecord`, `setupFactionTurn`, etc.) propagate errors up to the phase function; the outermost branch handles logging.
4. **Typed errors with structured payloads expand to multiple log lines at the orchestrator.** The `MutationApplyError` consumer emits one structured `Error`-level line per `MutationMiss`. This is part of the single logging step, not duplicated emission.
5. **Sub-engines may log freely about their own internal activity.** `Debug` lines about phase progress, `Info` lines about decisions made, `Warn` lines about internally-recovered anomalies (cache refresh, fallback used) — Session 1's domain. **Independent of the error-return discipline.**
6. **The orchestrator's caller renders, does not re-log.** Once the orchestrator returns up to `main.go` or the eventual TUI driver, the error has already been logged at its branch point. The caller displays / surfaces it but does not emit another log line.

### Edge cases

- **Sub-engine recovers internally.** Engine catches a stale cache, refreshes, succeeds — no error returned. Sub-engine emits a `Warn` line per Session 1's "anomaly we designed for." Not a return-path event; lives entirely in Session 1's discipline.
- **Sub-engine logs debug context before returning an error.** Legitimate if the debug line carries info *not* in the error chain (pre-failure state dump). Use `Debug` level so it doesn't pollute Info output. Distinct from logging the error itself.
- **Tests asserting on errors.** Tests assert against returned errors (or `errors.Is`/`errors.As` on them); they do not need to inspect log output. This decouples test stability from logging format choices.

### Why centralize at the orchestrator

Logging at every wrap site is the most common discipline failure in error-handling code — every layer logs as it wraps, and a single root cause becomes N log lines saying roughly the same thing with slightly different prefixes. Centralizing on the orchestrator boundary, combined with Session 1's cascading attr model (`turn → faction → phase → engine` already bound on the logger when the branch fires), gives a single emission with the full chain and the full context — all the visibility a debugger needs without the smear.

The framing that makes the rule stick: **wraps add display context, not emission.** `fmt.Errorf("attack: %w", err)` isn't announcing that an attack failed — it's *labeling* the error for when someone eventually renders it. Announcement is the logger's job; labeling is the wrap's. Conflating them is the source of double-logging.

---

## Open Questions

Carried into Plan. Each is a concrete decision Plan needs to ratify or defer; none block Discovery sign-off.

### Plan-scoped — Errors

1. **Naming of promoted sentinels.** Working names used in Discovery: `engine.ErrWorldEngineUnavailable` (collapses the two bare-string `"world engine not found"` duplicates), `action.ErrNoSelection` (collapses the four "no target/asset/world/faction selected" leaves), `action.ErrPreconditionFailed` (action precondition failures: insufficient Coin, unknown mode, missing target). Final names settled at implementation time.

2. **Granularity of the "no selection" sentinel.** One shared `action.ErrNoSelection` across all four action sites, or per-action sentinels? The four sites are semantically identical (collector returned no selection); a single sentinel keeps the recoverable list short and matches how `isRecoverable` is checked. Per-action sentinels offer finer rendering granularity if the eventual TUI wants per-action user messaging — but that's an F-005 concern and can be added later by declaring per-action sentinels that wrap the shared one with `%w`.

3. **Exact location of the `MutationApplyError` consumer.** The typed-error consumer (the `errors.As` + per-miss `Error`-level log emission) needs a concrete home. Most natural fit is `orchestrator.applyAndRecord` (`orchestrator.go:359-361`), since that's the immediate caller of `MutationEngine.Apply`. Plan should ratify the exact placement and confirm the structured-attr keys (`mutation_type`, `faction_id`, `entity_id`).

4. **Initial `recoverableErrors` list contents and location.** Likely lives in `engine/` (orchestrator-adjacent), declared as `var recoverableErrors = []error{action.ErrNoSelection, action.ErrPreconditionFailed}` or similar. Confirm placement, exported vs unexported, and whether typed-error recovery (none today, but the mechanism exists) belongs in the same helper or a separate one.

5. **Order of Plan-phase implementation.** Sentinels and convention adoption are leaf changes; the recoverable mechanism touches `orchestrator.go` broadly; the typed-error consumer touches `applyAndRecord`. Natural phase order: (1) declare new sentinels and demote the six bare-strings, (2) wire `isRecoverable` + the orchestrator branch, (3) wire `errors.As` consumer for `MutationApplyError`. Plan ratifies or reorders.

### Plan-scoped — Logging

6. **Canonical attr vocabulary.** The cascade model is decided and example attrs are listed, but the complete vocabulary across the engine isn't enumerated. Plan needs to pin the canonical keys (asset ID — `asset` or `asset_id`?, faction — `faction` or `faction_id`?, action name, mutation type, hex coords, etc.) before call sites start adding attrs, or the same concept will get keyed differently across the codebase and structured logs lose value.

7. **Concrete banner hook points.** The orchestrator emits banners "at well-defined hook points" (decided), but the exact mapping — which orchestrator method emits which banner, before or after which other emissions — needs to be ratified. Implied set: program-start (version banner), per-turn (turn + faction banner), per-phase. Plan locks the call sites.

8. **`pkg/logging` public API surface.** Decision exposes `*slog.Logger`, not `slog.Handler`. Plan ratifies the remaining surface: constructor signature (likely `logging.New(logsDir string, debug bool) (*slog.Logger, error)`), banner helper API (`banner.TurnStart`, `banner.PhaseStart`, etc.), and the package split between `pkg/logging` and `pkg/logging/banner`.

9. **Logger plumbing migration — order and test impact.** No sub-engine has a logger field today. Adding one ripples into every sub-engine constructor, `engine.New` / `NewWithRulebook`, all sub-engine tests, and the testharness. Plan picks an order (top-down or bottom-up) and decides the no-op test logger (`slog.New(slog.DiscardHandler)` vs nil-guarded call sites) for tests that don't care about output.

10. **Cleanup failure mode.** "Count-based cleanup at startup" is decided, but the failure mode is silent. Working assumption: cleanup runs *after* logger construction so the logger is available; cleanup errors are `Warn`-logged and never block startup. Plan ratifies (or chooses an alternative — e.g., silent swallow if `Warn` feels too noisy for a non-issue).

### Deferred to F-005

6. **Final `OnError` redesign.** Discovery predicts Shape B (semantic gameplay event); F-005 picks the actual signature(s) when the TUI's error-rendering needs are concrete.

7. **TUI error rendering surfaces.** Recoverable errors (gameplay outcomes) vs fatal errors (system failures) need distinct rendering paths. Recoverable goes through the observer surface (per Shape B); fatal goes through the logger fan-out and the engine's return path to the TUI driver. Exact rendering shape is F-005's concern.

## Out of Scope

- TUI implementation itself (`F-005`).
- Full conversion of existing error-return sites to new conventions — that's Plan/Execution work, not Discovery.
- Log aggregation / external sinks (Datadog, Loki, etc.) — single-user local tool; out of scope.
