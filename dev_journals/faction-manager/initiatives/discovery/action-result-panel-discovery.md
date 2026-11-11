# Action Result + Narrative Panel — Discovery

> Per-initiative Discovery within the **tui-turn arc**. Arc-level decisions (framework,
> engine boundary, overlay/collector conventions) are not re-litigated here — this doc
> covers only the result/narrative-panel decisions.

## Problem

Action results currently surface only as transient lines in the rolling event-stream
viewport (center column). They scroll past as the next faction's events arrive — there
is no held surface a GM can read after the fact. Two coupled gaps:

- **Gap A — placement (TUI-only).** Captured mutations from `OnActionResolved` scroll
  away in the event stream. Nothing holds the result of the action just resolved.
- **Gap B — fidelity (engine → observer).** Combat rolls and narrative (attacker vs.
  defender, ties, per-matchup damage, counters) never cross the `TurnObserver`
  boundary. `OnActionResolved` (`engine/observer.go:21`) carries only
  `(faction, selected, muts)`. The rolls are `RollResult` values
  (`hooks/types.go:26`) returned by `dispatch.RollWithHooks`; attack.go keeps only
  `.Sum` (lines 132, 146) and discards the rest. They are not persisted anywhere —
  `EventRecord` (`domain/event.go:28`) is mutation-derived only.

The gaps are coupled: you cannot place (A) rolls you are not emitting (B).

### Requirements (from Robert)

1. **Solve attack now,** structured for reuse in an eventual post-tui-turn UX pass —
   not a full cross-cutting build today.
2. **Pause/visibility serves GM decision-making,** not table narration. The held
   result pauses; the running result also informs the per-matchup choices in flight.
3. **Deliver both A and B** — the held result *and* the rolls/narrative.
4. **Lives in the right panel,** the way movement reshapes it today (`movementDetail`).

## Design Summary

The work splits into **two surfaces fed by one data structure**, because the engine's
shape forces the split: the `TurnObserver` is terminal and fire-and-forget, while the
`action.Collector` is the only channel that round-trips *during* `Resolve`.

- **Held result (terminal).** Attack accumulates a **matchup ledger** as it resolves.
  The final ledger is returned as a structured `Result` and carried to the observer
  via `OnActionResolved`, exactly mirroring how `OnBookkeepingApplied` carries a
  `BookkeepingResult` alongside mutations (`turn/bookkeeping.go:14`). The TUI holds the
  last result in a `Model` field and renders it as a right-panel card.
- **Live decision context (mid-resolution).** The *running* ledger is threaded into the
  existing `SelectDefender` / `ConfirmRedirectToBase` ask payloads, so the GM sees
  prior matchups while choosing the next defender or whether to redirect. No new
  observer event — the data rides channels that already round-trip.

The held result **pauses**: `AwaitCheckpoint(CheckpointActionResult)` already fires
right after `OnActionResolved` (`orchestrator.go:469`) but is swallowed by
`phase_collector.go:34` outside per-faction cadence. This initiative un-gates it so the
post-action pause reaches the TUI as the moment the result card is read.

This is **increment 1**, attack-scoped. The general mechanism (a streaming per-roll
event across *all* rolling paths, and folding the specialized detail cards into one
framework) is **increment 2**, deferred to the post-tui-turn UX pass.

### Decisions (ratified in this Discovery)

| # | Decision |
|---|----------|
| D1 | **Data path.** Widen `Output()` to `(Result, []Mutation, error)`; thread the result through `ActionEngine.Run` and into `OnActionResolved(faction, selected, result, muts)`. Mirrors the `BookkeepingResult` precedent. Chosen over an optional capability interface for a uniform action contract (Maintainability over one-time churn). |
| D2 | **Two surfaces, one ledger.** The matchup ledger feeds both the live ask payloads (running) and the terminal result card (final). |
| D3 | **Pause-at-outcome.** Un-gate `CheckpointActionResult` so the held result is read at a genuine pause point. |
| D4 | **Race-safety.** The ledger captures resolved display strings at build time inside `Resolve` (where the engine owns the data), not live pointers read later by the TUI — consistent with the `snapshotOf` / `resolveMovables` discipline. |
| D5 | **Generalization deferred.** Streaming per-roll event and unified detail framework are increment 2 (out of scope). |

## Domain Model Changes

### `action.Result` (new — `engine/action/action.go`)

A marker interface returned by actions, type-asserted by the TUI to the concrete
result — the same shape as `domain.Mutation` / `action.Action` (interface in the
engine, concrete types asserted at the edge).

```go
type Result interface { isActionResult() }
```

Actions with nothing to report return `nil`.

### Widened action contract

```go
// Action interface
Output() (Result, []domain.Mutation, error)

// ActionEngine.Run
func (ae *ActionEngine) Run(...) (Result, []domain.Mutation, error)

// orchestrator action phase
result, actionMutations, err := e.Action.Run(...)
// ...
observer.OnActionResolved(faction, selectedAction, result, combined)
```

Every existing action gains a `nil,` result return (mechanical).

### `OnActionResolved` signature

```go
OnActionResolved(faction *domain.Faction, selectedAction action.Action, result action.Result, mutations []domain.Mutation)
```

Adapter `observer.go` forwards `result` into a widened `ActionResolvedPayload`.

### `AttackResult` / `Matchup` (new — `engine/action/actions`, proposed; refine in Plan)

```go
type AttackResult struct {
    Matchups []Matchup
}
func (AttackResult) isActionResult() {}

type Matchup struct {
    Attacker        string          // resolved display label
    Defender        string          // resolved display label
    DefenderFaction string
    Attack          hooks.RollResult
    Defense         hooks.RollResult
    Tie             hooks.TieOutcome
    Damage          int
    Redirected      bool            // damage sent to defender's base
    Outcome         MatchupOutcome  // damaged / destroyed / base-hit / countered / no-effect
}
```

### Ask-payload enrichment

`SelectDefenderPayload` and `ConfirmRedirectToBasePayload` (`adapter/channels.go`) each
gain the running `[]Matchup` so the in-flight overlay/panel can show prior matchups.

## Per-Area Design Details

### Engine — attack Resolve

`Resolve` already accumulates `attack.mutations`; it additionally accumulates
`attack.matchups`. Each matchup record is built where the live data is in hand (rolls,
damage, redirect outcome) with display labels resolved at that point (D4). The running
slice is passed into the `SelectDefender` / `ConfirmRedirectToBase` collector calls;
the final slice is returned from `Output()` as an `AttackResult`.

### Engine — checkpoint un-gating

`phase_collector.go:34` currently returns early for every checkpoint except
`cycle_summary` outside per-faction cadence. `CheckpointActionResult` must reach the
TUI so the result card is read at a pause. (Exact un-gating shape is an open question
below.)

### TUI — held result card

A new `Model` field holds the last `AttackResult` (resolved into render-ready strings).
`applyEvent` populates it on `EvtActionResolved`; `detailView` renders it in the
no-overlay branch (replacing `baseDetail`) until the next faction's turn starts, where
`EvtFactionTurnStarted` already clears per-faction state (`execution.go:190`).

### TUI — live decision context

`newOverlay`'s `SelectDefender` / `ConfirmRedirectToBase` arms receive the running
ledger from the enriched payloads and render prior matchups alongside the choice.

## User-Facing Impact

During an attack the GM sees, in the right panel and in-flight overlays, the rolls and
outcomes of prior matchups as they pick each defender and decide redirects. After the
action resolves, the panel holds a matchup-by-matchup result card (attacker vs.
defender rolls, ties, damage, redirects, destruction/counters) at a pause the GM
acknowledges before the next faction proceeds.

## Out of Scope

- **Increment 2 — general result/narrative mechanism.** A streaming per-roll observer
  event (`OnRoll`/`OnMatchup`) covering *all* rolling paths (Expand rival free-attack,
  ability faction-tests), and folding `movementDetail` / `statRaiseDetail` into one
  unified detail framework. Deferred to the post-tui-turn UX pass.
- **"Anytime there's a choice, confirm" cadence principle.** A binding principle for
  the eventual TUI UX pass: every consequential choice gets a confirm/pause. Broader
  than attack (touches every action's cadence). Recorded as a Backlog entry in
  `planned-work.md` so it survives this doc's archival; not built here.

## Open Questions

For Plan to resolve:

1. **Checkpoint un-gating shape.** Un-gate only `CheckpointActionResult` in
   `phase_collector.go`, or restructure the cadence gate more generally? Narrowest
   change preferred unless it fights the existing structure.
2. **Live-ledger placement.** During `SelectDefender` / `ConfirmRedirectToBase`, render
   the prior-matchup ledger inside the overlay (center band) or in the right detail
   panel? Affects where the rendering code lives.

## Reference Exemplars

- **`BookkeepingResult` / `OnBookkeepingApplied`** (`turn/bookkeeping.go:14`,
  `engine/observer.go:19`) — the precedent for a structured phase result carried to the
  observer alongside mutations. D1 mirrors it.
- **`movementDetail` / `statRaiseDetail`** (`execution.go:355`, `369`) — right-panel
  specialized cards keyed off the mounted overlay; the held result card is a sibling.
- **`snapshotOf` / `resolveMovables`** (`execution.go:205`, `394`) — the race-safe
  capture discipline D4 follows.
- **`hooks.RollResult`** (`hooks/types.go:26`) — the per-die roll breakdown the ledger
  preserves instead of discarding.
