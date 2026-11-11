# Goals

> **Code:** `internal/faction/engine/goal/` (`goal.go`, `compile.go`, `vocabulary.go`, `events.go`, `measures.go`, `xp.go`, `evaluate.go`, `offers.go`), `internal/faction/rulebook/goals.go`, `internal/faction/domain/goal.go`

## Purpose

A faction pursues one active goal at a time. The goal engine reads each goal's
behavior from the rulebook, as six named parts, and decides after every mutation
batch whether that batch moved, reset, completed or failed each faction's goal. It
also offers the goals a faction may pick at the start of its turn and builds the
pick. Like every sub-engine it emits mutations and applies nothing. No Go code
names a goal ID.

## Shape

### Goals as data

A goal in `rulebooks/_core/goals.toml` holds its text (`id`, `name`,
`description`) and up to six parts, each an inline table:

| Part | Job | Keys |
|---|---|---|
| `start` | What the pick needs and when it may offer the goal | `inputs` (`target_faction`), `needs_actions`, `gate` |
| `progress` | How progress moves: an amount per matching event, or a count read from state | `event` + `amount` + `filters`, or `measure` + `filters` |
| `reset` | An event that sets progress to 0 | `event` + `filters` |
| `threshold` | The progress that completes the goal | `ratings` (summed, times `multiplier`), or `constant` |
| `complete` | The first matching event completes the goal | `event` + `filters` |
| `xp` | The XP paid on completion | `formula` + its parameters |

A goal has exactly one of `threshold` or `complete`. `threshold` and `reset` need
`progress`.

```toml
[goals.PeaceableKingdom]
id = "CORE-G-007"
progress = { event = "action_taken", amount = "one", filters = { faction = "own", action_not = ["Attack"] } }
reset = { event = "action_taken", filters = { faction = "own", action = ["Attack"] } }
threshold = { constant = 4 }
xp = { formula = "constant", amount = 1 }
```

### Two checks: the loader, then the engine

`loadGoals` (`rulebook/goals.go`) reads the parts strictly, refuses unknown keys,
and converts stat names (`"Force"`) to `FactionStat`. Every other name (an event,
a filter value, an amount, a measure, a formula, a start input) stays a string in
`domain.Goal`, because `rulebook` cannot import the goal engine.

`goal.New(rulebook, actionNames, log)` compiles every goal. `compile` binds each
part to the spec that runs it and refuses an unknown name, a filter the event does
not accept, a bad value, a wrong combination of parts, or an `action` filter that
names no registered action. `NewWithRulebook` builds the goal engine after the
actions register and returns its error, so a bad goal stops the engine from
starting. `needs_actions` is not checked there: Seize Planet is unregistered on
purpose until `F.040`.

### The vocabulary

The events, measures and formulas are tables in the goal package.

| Event | Matches | Filters | Amount |
|---|---|---|---|
| `kill` | `AssetRemoved` | `owner`, `caused_by`, `category`, `rated_above` | `one` |
| `damage` | `AssetHPDelta`, `BaseHPDelta` with a negative delta, capped at the HP left | `owner`, `caused_by` | `damage` |
| `base_planted` | `BaseAdded` | `caused_by`, `new_world` | `one` |
| `influence_bought` | `InfluenceDelta` with a positive delta | `caused_by` | `delta` |
| `government_taken` | `WorldGovernmentChanged` | `to` | `one` |
| `faction_destroyed` | The `FactionHPDelta` that takes a faction to 0 HP or less | `target`, `caused_by` | `one` |
| `action_taken` | `ActionTaken` | `faction`, `action`, `action_not` | `one` |

- **Filters** — `owner` and `caused_by` take `own` or `rival`, read against the
  goal's holder; any cause counts when the holder caused it. `category` and
  `rated_above` read the removed asset's definition from the state before the
  batch. `target = "start_input"` names the faction the pick recorded.
- **Measure** — `stealthed_assets` counts the holder's stealthed assets on rival
  ground: a world a living rival governs, or the independent homeworld of a
  living rival. With `exclude_held_at_start`, the assets recorded at the pick
  never count.
- **XP formulas** — `constant` (`amount`, plus `contested_bonus` when a living
  rival has an asset or Base on the completing event's world), `half_progress`
  (half the progress, rounded up), `target_average` (1 plus the target's average
  rating, rounded up), `ruler_average` (half the resisting ruler's average
  rating, rounded up, minimum 1; 1 when no ruler resisted).
- **Gates** — `needs_actions` is open when every named action is registered.
  `rating_above_last_completion` is open when the named rating is above its value
  in the faction's `CompletionRatings` record for the goal, or when there is no
  record.

### One beat: every batch

`UpdateProgress(mutations, factionState, log)` runs inside the orchestrator's
`applyAndRecord`, for every batch in every phase, after the phase's reactors and
before the batch applies. It reads every faction's goal in faction ID order. It
skips a faction with no goal, a destroyed faction, and a goal the engine cannot
run (`Runs`: a goal ID the rulebook does not know, or a target goal with no
target, from an older save).

For one holder and one batch, `evaluate`:

1. Returns nothing when the batch abandons the held goal.
2. Checks `reset`. A match sets progress to 0 and ends the batch for that goal.
3. Adds the batch's progress: the sum of the event's match amounts, or the gap
   between the measure's count and the progress held. A measure counts in a copy
   of the state with the batch applied.
4. Checks completion: the `complete` event, or the threshold in a batch that moved
   progress. On completion it emits `GoalCompleted` and `XPAwarded`, with XP from
   the state before the batch.
5. When the goal did not complete, checks the built-in lost-target rule: a goal
   with a `target_faction` input fails with `GoalFailed` (cause `target_lost`)
   when its target is gone, already destroyed, or destroyed in this batch. The
   failure replaces the batch's progress.

The orchestrator runs the reactors over the goal mutations alone, then applies the
whole batch at once.

### The turn-start pick

`Offers(faction, factionState)` lists the goals whose gates are open, in goal ID
order, each with its target candidates when it takes a `target_faction` input:
every other faction above 0 HP. A goal with an input and no candidate is left
out. `Pick(faction, offers, choice)` checks the choice against the offers and
builds `GoalPicked`; a goal whose progress uses `exclude_held_at_start` records
the faction's stealthed asset IDs there. The [orchestrator](orchestrator.md) runs
the pick after stat-raise; `GoalPicked` is the one mutation that sets
`ActiveGoal`.

## Goals Catalogue

The eleven goals in `rulebooks/_core/goals.toml`. The rulings behind the numbers
are in [`swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md), Goal
Rulings.

| Goal | Moves on | Completes when | XP |
|---|---|---|---|
| Military Conquest (`CORE-G-001`) | `kill` of a rival Force asset | Kills reach Force | `half_progress` |
| Commercial Expansion (`CORE-G-002`) | `kill` of a rival Wealth asset | Kills reach Wealth | `half_progress` |
| Intelligence Coup (`CORE-G-003`) | `kill` of a rival Cunning asset | Kills reach Cunning | `half_progress` |
| World Seizure (`CORE-G-004`) | — | `government_taken` to the holder; needs Seize Planet (`F.040`) | `ruler_average` |
| Expand Influence (`CORE-G-005`) | — | `base_planted` on a world with no holder Base | 1, +1 contested |
| Blood the Enemy (`CORE-G-006`) | `damage` to rival assets and Bases | Damage reaches Force + Cunning + Wealth | 2 |
| Peaceable Kingdom (`CORE-G-007`) | Own `action_taken` that is not Attack; an Attack resets | 4 | 1 |
| Destroy the Foe (`CORE-G-008`) | — | The holder's blow destroys the picked target | `target_average` |
| Inside Enemy Territory (`CORE-G-009`) | `stealthed_assets` on rival ground | The count reaches Cunning | 2 |
| Invincible Valor (`CORE-G-010`) | — | `kill` of a rival Force asset rated above the holder's Force | 2 |
| Wealth of Worlds (`CORE-G-011`) | `influence_bought` | Coin spent reaches 4 × Wealth; offered again only above the last completion's Wealth | 2 |

## Key Decisions

- **Behavior comes from data, checked at construction.** A goal states its
  behavior in six parts, and the engine refuses to start on a name it cannot run.
  This closes the old "data-only goals are a silent skip" seam: no goal in the
  data does nothing. Rejected: one `shape` field, and Go kernels for the process
  goals. (A.001.8 discovery, `E.002` G-1; reverses frozen log 111.)
- **The loader checks shape; the goal engine checks names.** `rulebook` cannot
  import the goal engine, so goal-only names stay strings in `domain.Goal` and
  `goal.New` compiles them. (A.001.8 Effort 1 Decision 6.)
- **Goals have their own events and filters.** A goal event is a matcher over a
  finished batch; an effect trigger picks a hook route. They share only the
  layout, the value words (`own`, `rival`, stat names) and strict loading, so a
  change on one side is never a question for the other. Goals' `kill` counts any
  cause; effects' `asset_destroyed` counts combat only. (A.001.8 discovery D-18.)
- **No Go names a goal ID.** Behavior is keyed on parts and roles; goal IDs carry
  the `CORE-` prefix and appear only in data. (A.001.8 discovery D-26.)
- **One beat: every batch, for every faction.** Goals read every mutation batch in
  every phase, on any faction's turn, so a kill made while defending counts and a
  goal can complete on a rival's turn. Replaces "two advancement paths, by beat"
  (`CheckLock` before the action, `UpdateProgress` after it). Rejected: the
  holder's own turn only. (A.001.8 discovery D-13; Effort 2 Decision 1.)
- **Goals run after the reactors and before the batch applies.** Progress reads
  about-to-be-removed assets in live state, and XP is worked out before the loss,
  so the apply must follow. A reactor's capture or removal counts as progress; a
  reactor pass over the goal output alone follows. (Frozen log 110–123, 119;
  A.001.6 overview Decision 17.)
- **A measure reads a projected copy of the state.** A count read from state
  would be one batch late, so the measure clones the state and applies the batch
  to the copy. Accepted: a full clone per holder per batch, while states are small.
  (A.001.8 Effort 1 Decision 2; Effort 2 Decision 10.)
- **A reset wins its batch.** When `reset` matches, progress goes to 0 and the
  batch's progress events do not count. (A.001.8 Effort 1 Decision 8.)
- **A lost target fails the goal, after completion is checked.** A goal with a
  `target_faction` input that can no longer complete fails with `GoalFailed`, by
  whatever hand the target fell; a blow that completes the goal is not a loss. The
  failure replaces the batch's progress. Rejected: a per-goal `fail` part, and
  `GoalAbandoned` with a cause. (A.001.8 discovery D-21; Effort 2 Decisions 8 and
  9, Decision Record 6.)
- **A destroyed faction counts nothing and is no rival.** The engine keeps a
  faction at 0 HP in state, so its goal is skipped and its leftovers neither
  contest a world nor make rival ground. (A.001.8 Effort 2 Decision Records 4 and
  10.)
- **The goal engine owns the offers, the gates and the start record.** The gate
  names and the start-record rule stay with the vocabulary. The pick's choice is
  checked against the offers; anything else is a collector bug, not recoverable.
  (A.001.8 Effort 3 Decisions 5 to 8.)
- **A held goal the engine cannot run is replaced at the pick.** An older save can
  hold an unknown goal ID or a target goal with no target. Progress skips it, and
  the turn-start pick treats it as no goal, so it never blocks the faction.
  (A.001.8 Effort 1 Decision Record 16.)
- **Goals emit, never mutate.** Progress, completion, XP and failure are mutations
  the orchestrator applies; `GoalPicked` sets `ActiveGoal` and `GoalCompleted`
  writes the completion record when they apply.

## Dependencies

**Depends on** `domain` (the `Goal` parts, `ActiveGoal`, the goal mutations
`GoalProgressed`, `GoalCompleted`, `GoalFailed`, `GoalPicked` and `XPAwarded`),
[persistence & static data](persistence.md) for the rulebook and `FactionState`,
and the [mutation](effect-mutation.md) engine, which the stealth measure uses to
project a batch onto a copy of the state.

**Depended on by** the [orchestrator](orchestrator.md), which builds the engine in
`NewWithRulebook`, calls `UpdateProgress` in `applyAndRecord` for every batch, and
calls `Offers` and `Pick` in the goal-pick phase. `Abandon Goal`, an
[action](actions.md), ends a goal from outside the engine. Exchange Consulate's
`xp_roll` [effect](effect-mutation.md) reacts to `GoalCompleted`.
