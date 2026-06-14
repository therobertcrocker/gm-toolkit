# Turn Pipeline

> **Code:** `internal/faction/engine/turn/{turn,bookkeeping,order,history,event_record}.go`

## Purpose

The turn engine owns the *lifecycle* of a faction cycle: opening a turn, rolling
the rotation order, tracking the pause/resume cursor, computing bookkeeping
(income and maintenance), and appending history. It is deliberately the thinnest
of the sub-engines — it computes and *emits* mutations and display results but
never applies them to state. If you are asking "whose turn is it, what does a
faction earn and owe this turn, and how is the cycle recorded," this is the page;
the [orchestrator](orchestrator.md) page covers how these pieces are sequenced.

## Shape

`TurnEngine` (`turn.go`) holds only a roller, the rulebook, and a logger. All
turn *state* lives on the `FactionState` it is handed — specifically the
`domain.TurnState` pointer (nil when no turn is in progress), which is persisted
so a paused turn survives a restart.

### The lifecycle and the resume cursor

Four methods bracket a cycle:

- **`Start`** increments `CycleNumber`, readies every asset, and builds a fresh
  `TurnState` with a rolled `FactionOrder`. It refuses (`ErrTurnInProgress`) if a
  turn is already live.
- **`CurrentFaction`** resolves the faction at the cursor (`FactionOrder[CurrentIndex]`).
- **`Advance`** marks the current faction done: it increments `CurrentIndex`,
  resets `BookkeepingApplied`, and — when the index runs past the order — nils
  `CurrentTurn` and returns `true` to signal the cycle closed.
- **`Abandon`** drops the in-progress turn outright; mutations already applied
  stay applied.

```go
type TurnState struct {
	InProgress         bool
	CycleNumber        int
	FactionOrder       []string
	CurrentIndex       int
	BookkeepingApplied bool
}
```

`CurrentIndex` is the across-faction cursor; `BookkeepingApplied` is the
within-faction cursor. Together they let a turn pause mid-pipeline (at a
checkpoint) and resume exactly where it stopped — the orchestrator never has to
remember anything the cursor doesn't already carry.

### Rotation order

`buildFactionOrder` (`order.go`) sorts the faction IDs, rolls a start index in
`[0, n)`, and rotates the sorted slice from there. Sorting before randomizing is
the whole point: the same faction set under the same seed produces the same
rotation, which is what makes turn tests and history replays deterministic.

### Bookkeeping

`ApplyBookkeeping` is the substance of the engine. It returns
`(BookkeepingResult, []domain.Mutation, error)` — the result is display-level
(income totals, lists of assets lost or left unmaintained), the mutations are the
actual state-change requests for the orchestrator to apply. It is **idempotent on
resume**: if `BookkeepingApplied` is already set it returns a zero result and nil
mutations, so a turn resumed after the bookkeeping checkpoint never double-charges.

The math, in order:

1. **Income** — `Wealth/2 + (Force+Cunning)/4`, emitted as a single positive
   `CoinDelta`.
2. **Hook-budget reset** — `faction.HookBudgets = nil`, so each registered hook's
   "once per turn" allowance starts fresh this cycle.
3. **Maintenance** — `applyMaintenance` walks the faction's assets (in
   `SortedAssets` order, for determinism), charging each asset's per-turn cost
   against a running Coin total seeded at `Coin + income`.

Maintenance carries two SWN mechanics. First, the **category surcharge**: a
faction supports assets up to its rating in each stat (Force/Cunning/Wealth) for
free; every asset beyond that rating costs +1, distributed one-per-asset until
the overage is exhausted. Second, the **grace turn**: when a faction can't afford
an asset's upkeep, an already-unmaintained asset is removed (`AssetRemoved`), but
a currently-maintained one is only flagged unmaintained (`AssetMaintainedFlag`
false) — it survives one lean turn before loss.

Per-asset base cost comes from `rulebook.Assets[...].Maintenance`, resolved
through `dispatch.ResolveMaintenanceCost` so registered `MaintenanceCostModifier`
hooks can adjust it.

### History

`RecordHistory` (`history.go`) builds one `domain.EventRecord` — cycle number,
faction ID, timestamp, and the turn's mutations serialized as `MutationRecord`s
(type tag + raw-JSON payload) — and appends it as a single line to the campaign's
`history.jsonl`. One record per faction per cycle. This is an append to the
history file only; it is **not** a state save (see Key Decisions).

## Key Decisions

- **The engine emits, the orchestrator applies.** `TurnEngine` computes
  bookkeeping into `[]domain.Mutation` and returns it; it never writes faction
  state. Mutation application is owned by the [orchestrator](orchestrator.md) via
  the [effect & mutation](effect-mutation.md) apply layer. This keeps mutations
  the only state-change currency and keeps the engine trivially testable against
  expected mutation lists.
- **Bookkeeping is idempotent on resume.** The `BookkeepingApplied` flag guards
  `ApplyBookkeeping` so a turn paused at the bookkeeping checkpoint and resumed
  never grants income or charges upkeep twice. The cursor, not the caller, owns
  resume safety. (Frozen log 40–45.)
- **Rotation is sorted-then-rotated for determinism.** IDs are sorted before the
  roller picks a start index, so a given faction set and seed always yield the
  same order — the precondition for reproducible turn tests and history replay.
- **A grace turn before asset loss.** An unaffordable asset is removed only if it
  was *already* unmaintained; a maintained asset is merely flagged unmaintained
  first. Faction fortunes degrade gradually rather than collapsing in one bad
  turn — the SWN upkeep rule, encoded in the maintenance branch.
- **Hook budgets reset each turn.** Clearing `HookBudgets` at the top of
  bookkeeping is what makes "once per turn" hooks fire once per turn; budgets
  re-accumulate from scratch rather than persisting across cycles.
- **History append is decoupled from state save.** `RecordHistory` appends to
  `history.jsonl`; durable faction state is persisted separately, by the
  orchestrator, at phase gates. History grows per turn; on-disk state lands only
  on phase boundaries (cross-link [persistence & static data](persistence.md)).
  (Frozen log 195–196.)

## Dependencies

**Depends on** `domain` (the `TurnState` cursor, `Faction`, the `CoinDelta` /
`AssetRemoved` / `AssetMaintainedFlag` mutation types, `EventRecord`),
[persistence & static data](persistence.md) for `FactionState` and the rulebook
(asset categories and base maintenance), and [hooks](hooks.md) /
`hooks/dispatch` for maintenance-cost resolution.

**Depended on by** the [orchestrator](orchestrator.md), which drives `Start` /
`CurrentFaction` / `Advance` across the cycle and calls `ApplyBookkeeping` and
`RecordHistory` within the pipeline, and the [narrative digest](narrative-digest.md),
which reads back the `history.jsonl` this engine writes.
