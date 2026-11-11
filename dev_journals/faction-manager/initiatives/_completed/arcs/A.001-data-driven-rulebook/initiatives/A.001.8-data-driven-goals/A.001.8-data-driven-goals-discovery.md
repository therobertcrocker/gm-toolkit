# Data-Driven Goals — Discovery

## In Plain Words

- **What:** goals stop getting their behavior from twelve Go handlers, one per goal ID. Each of the eleven rules goals in `rulebooks/_core/goals.toml` states its behavior as data, built from six named parts: what the pick needs, how progress moves, what resets it, the threshold, what completes it, and the XP it pays. Goals read every batch of changes on any turn, so a kill made while defending counts. A faction with no goal picks one at the start of its turn, in the engine and in the TUI.
- **Why:** a goal with no Go handler does nothing today, so a second rulebook gets new goal behavior only by writing Go. And the goal flow is broken in play: after a faction finishes a goal, nothing lets it pick the next one, Destroy the Foe never gets a target, the wizard offers goals that cannot work, and Peaceable Kingdom misses a missed attack.
- **How it fits:** the last goal-side initiative of the `A.001` arc (arc-plan AP-8), the twin of `A.001.7` (data-driven tags), following `E.002`'s G-1 and I-1. Seize Planet and Change Homeworld were goals in disguise; they become multi-turn actions in a separate row, `F.040`, and are off until it lands. `A.001.8` gates nothing: `_core` goals serve SWN and H.A.W.K before and after it. It retires the goal half of `R.004` and leaves `F.013` (goal state predicates) a place to plug in.

## Problem

Goal behavior lives in Go, keyed on bare goal IDs:

- **Per-ID handlers.** `goal.New` (`engine/goal/goal.go`) registers twelve `Handler` types from `engine/goal/goals/`, each returning a bare ID from `GoalID()` (`"G-001"` … `"G-012"`). A goal in `goals.toml` with no handler is a silent skip (goals page, Key Decisions). So H.A.W.K, or a GM, can add goal text but not goal behavior.
- **IDs hard-coded outside the goal engine.** `seize_planet.go` checks and writes `"G-004"`; `change_homeworld.go` writes `"G-012"` (`B.001`). Tag data names goals by bare ID: Exchange Consulate names `G-007`, Imperialists names `G-004` (`rulebooks/swn/tags.toml`).
- **XP is not read from data.** `difficulty` holds a formula name (`half_assets_destroyed`, `moderate`, `none`, …) that the engine never reads; each handler computes XP in Go (`B.005`).

The goal flow is also broken in play (`E.002` audit, read from code, not a run):

- **No new goal after completion (`B.039`).** `GoalCompleted` and `GoalAbandoned` set `ActiveGoal` to nil, and no turn step picks the next goal. The old TUI's goal select went with `798cd45`. The rules let a faction pick a new goal at the start of its next turn.
- **Destroy the Foe has no target (`B.040`).** Only `destroy_the_foe.go` reads `ActiveGoal.TargetFactionID`; no production code writes it.
- **The wizard offers goals that need start data (`B.041`).** `goalsToOptions` lists every goal. `G-004` from the wizard never advances past phase 0; `G-012` from the wizard counts `TurnsRemaining` below 0 and skips the faction's turn forever.
- **Peaceable Kingdom misses a missed attack (`B.013`).** It detects an Attack from damage or stealth-loss mutations; a miss by an unstealthed attacker emits neither. `UpdateProgress` gets the mutations, not the action name.

## Architecture Context

Pages read: [architecture-overview](../../architecture/architecture-overview.md), [goals](../../architecture/engine/goals.md), [orchestrator](../../architecture/engine/orchestrator.md), [effect-mutation](../../architecture/engine/effect-mutation.md) (Key Decisions), [persistence](../../architecture/engine/persistence.md) (Key Decisions on prefixed IDs and the re-seed). Left to Plan: [actions](../../architecture/engine/actions.md) (unregistering Seize Planet and Change Homeworld) and the interface [state-machine](../../architecture/interface/state-machine.md) and [overlays](../../architecture/interface/overlays.md) pages (the goal-select overlay).

Binding Key Decisions from the goals page:

- **Two advancement paths, by beat.** Time and condition progress runs in `CheckLock` (goal-lock phase, after bookkeeping and movement, before the action). Event progress runs in `UpdateProgress`, after the action.
- **`UpdateProgress` runs before mutations apply, and after `MutationReactors`.** Progress reads about-to-be-destroyed assets in live state; a reactor's capture or removal counts as progress; no reactor reads goal output.
- **`locks` is a separate leaf package**, to break the `goal` / `goals` import cycle.
- **Data-only goals are a silent skip, as an incremental seam.** The page names this initiative as the end of that seam.
- **Goals emit, never mutate.** Completion, XP, ticks, phase advances and the government change are all mutations the orchestrator applies.

Arc decisions that bind (arc-discovery ACD-2, ACD-4, ACD-6; arc-plan AP-8, AP-9): behavior comes from data on every surface; one primitive, many harnesses; goals stay in `_core`, composed into each campaign; goal IDs take a rulebook prefix here, not earlier; this initiative renames the goal references in tag data.

Historical decisions from the frozen [decisions log](../../dev_journals/faction-manager/archive/decisions-log.md) (`feature/goal-engine`, 110–123; movement-redesign cutover, 237):

| # | Decision | Bearing here |
|---|---|---|
| 111 | Goal behavior lives in Go handlers; TOML holds metadata only, because goal logic branches on mutations, stats and world state | This initiative reverses it |
| 112 | Change Homeworld is a goal, not an action, because it spans turns and holds lock state | Still holds for the multi-turn half |
| 113 | The Seize Planet action starts the Planetary Seizure process; the goal engine runs the later turns | Intent is "the action creates the goal state". The code from `92f38f8` instead needs `G-004` already picked at phase 0 |
| 114 | Abandon Goal is an action that costs the turn's income | Unchanged |
| 119, 123 | XP for variable goals is computed at completion, before destructive mutations apply | Binds the `xp` part |
| 121 | Inside Enemy Territory counts `AssetStealthApplied` events so that assets stealthed before adoption do not count | Context for the `G-009` rules call |
| 237 | Reverses 112's start: Change Homeworld starts from an action, "mirroring" Seize Planet | Not a true mirror. Change Homeworld needs no active goal; Seize Planet needs `G-004` picked first |

**Key Decisions this initiative changes** (to graduate at its close):

- **Goals page, "Two advancement paths, by beat."** With the locks and countdowns gone to `F.040` (D-1) and `tick` replaced by `action_taken` (D-19), goals have one beat: after each batch, on every batch (D-13). `CheckLock` keeps no goal work.
- **Goals page, "Data-only goals are a silent skip."** The seam closes: every goal's behavior is its data, checked at engine construction.
- **Goals page, catalogue and worked example.** Planetary Seizure's process leaves the goals; `G-012` leaves the catalogue (D-5).
- **Orchestrator page, phase order.** The turn gains the goal pick after stat-raise (D-2), and goal progress runs on every phase's batch (D-13).

## Design Summary

- **Composed parts (`E.002` G-1, narrowed).** Each goal composes six parts: `start`, `progress`, `reset`, `threshold`, `complete` and `xp`. Each name is a Go primitive with parameters, checked at engine construction, as an effect name is. `E.002`'s `phases` went with the process goals (D-1), and its `tick` became an `action_taken` event (D-19). The full vocabulary and all eleven goals are in Goal Vocabulary, below. Rejected in `E.002`: one `shape` field (G-2) and Go kernels for the process goals (G-3).
- **A goal vocabulary of its own (D-18).** Goal events and filters are named and owned by the goal package; they share only the sub-table layout (D-17), the value words and strict loading with `[[effects]]`.
- **One beat, every batch (D-13).** Every faction's goal reads every mutation batch, on any turn and in any phase, after the reactors and before the batch applies. Any kill or damage the faction causes counts (D-14). Each action batch opens with a record-only `ActionTaken` (D-19).
- **Seize Planet and Change Homeworld become multi-turn actions in `F.040` (D-1, D-7).** World Seizure becomes an ordinary goal that completes when the faction takes a world's government; `G-012` leaves the data (D-5). Until `F.040` lands both actions are unregistered, and the pick leaves out World Seizure, which names the action it needs (D-12).
- **A turn-start goal pick (D-2 to D-4, D-6).** After stat-raise, a faction with no goal may pick one through a `PhaseCollector` ask and a TUI overlay; the ask gathers the start inputs the goal declares (Destroy the Foe's target faction). The wizard sets no goal (D-25).
- **Rules readings settled (D-8 to D-11, D-15, D-16, D-20 to D-22).** Every XP formula rounds up; World Seizure pays from a resisting ruler only; Inside Enemy Territory counts assets held; Wealth of Worlds reopens when Wealth rises above its last completion; Destroy the Foe completes only on the faction's own blow and fails (`GoalFailed`) when a third party lands it.
- **Roles, not IDs.** No Go code names a goal ID.
- **Prefixed IDs (`E.002` I-1, D-26).** Goal IDs become `CORE-G-001` to `CORE-G-011`, and the goal references in tag data follow (AP-9), checked at load. Saves move by a sandbox re-seed, with no migration code.
- **Three efforts (D-27):** goals as data, then every batch, then the turn-start pick.

## Decisions

- **D-1 — Seize Planet and Change Homeworld become multi-turn actions in a separate row; `A.001.8` does not wait for it.** Robert's call. In the rules, both are actions, and Planetary Seizure is a goal that completes when the faction takes a world; the engine fused each action to a goal because only the goal engine held state across turns (frozen log D112, D113, D237). A new row gives actions their own multi-turn state and moves the process (lock, phases, countdown, the government change and `T-011`) out of the goals. `A.001.8` writes `G-004` in the new goal data form (D-5 drops `G-012`), and its behavior in play waits for that row, as Psychic Academy's entry waited on its own row. Rejected: the new row gating `A.001.8`'s Plan, the multi-turn work as an effort of `A.001.8`, and keeping the two process handlers as fenced Go.
  - **The goal vocabulary loses the process parts.** With the process in the action, no goal needs `phases` or a lock; `CheckLock`'s goal side keeps only the per-turn beat (`tick`). To confirm when the vocabulary is settled.
  - **In the gap, Seize Planet and Change Homeworld leave the action menu.** Robert's call. Neither action is registered until the multi-turn row registers it again, so the GM never spends a turn on an action that does nothing, and "is the action registered" is the signal D-12 reads. Rejected: keeping them in the menu as no-ops. *(Amended by D-12: first written as "registered, with `Validate` returning false".)*
  - **An in-flight `G-004` or `G-012` is dropped.** Robert's call. The faction holds no goal and gets the turn-start pick next turn; the gap loses the process state anyway. Rejected: moving an in-flight `G-004` to a plain `CORE-G-004`. *(Amended by D-26: first written as "the save migration clears it"; saves move by re-seed, with no migration code.)*
- **D-2 — The turn-start goal pick runs after stat-raise and before bookkeeping.** Robert's call. The turn reads `setup → stat-raise → goal pick → bookkeeping → …`. The rules put the pick "at the start of their next turn", and XP spend is also a turn-start step; with the raise first, a Wealth raise this turn can satisfy `G-011`'s gate for the same turn's pick. Rejected: the pick before stat-raise.
- **D-3 — The pick asks only the start inputs a goal's `start` part declares.** Robert's call. Today that is `G-008`'s target faction, chosen from the living rivals; a future goal adds an input in its data, not in the ask. This closes `B.040` by its shape. Rejected: a fixed set of fields in every ask, which hides what a goal needs.
- **D-4 — A faction may decline the pick.** Robert's call, from the rules text "may pick one". The collector returns no goal, as `SelectStatRaise` returns nil, the turn goes on, and the ask comes back next turn. Rejected: a required pick.
- **D-5 — `G-012` leaves `goals.toml`.** Robert's call. Change Homeworld is not in the rules' goal list; once it is a multi-turn action (D-1), a goal of the same name only fills the goal slot and pays 0 XP. The multi-turn action row owns Change Homeworld. Rejected: porting `G-012` as a pickable data goal, and porting it as an entry the pick leaves out.
- **D-6 — The TUI's goal-select overlay is in this initiative.** Robert's call. The engine step and the overlay that answers its ask land together, so `B.039` is fixed in play at merge, and the step is no longer lost with a frontend. Rejected: the engine step alone, with the overlay in a later row.
- **D-7 — The multi-turn action work is a standalone feature row, outside the arc.** Robert's call. It is filed in the Backlog as `F.040`. The arc can close without it; Seize Planet and Change Homeworld stay off until it lands. Rejected: an in-arc `A.001.9` with an arc-plan amendment.
- **D-8 — `G-004` pays from the previous ruler only if that ruler resisted.** Robert's rules call, on the `goals.toml` text "If no opposing faction resists, counts as difficulty low". The XP is half the previous ruler's average rating when the ruler had an unstealthed asset on the world when the seizure began; otherwise it is 1 (low), as it is for a world with no faction ruler. Under D-1 the seizure is `F.040`'s multi-turn action, so the action records who was on the world at the start and the government-change event carries it; `A.001.8`'s `xp` formula reads it from the event. Only unstealthed assets count, the same assets the seizure must destroy; stealthed assets are not resistance. Rejected: today's reading, where the ruler's ratings count whether or not it had assets there, and counting stealthed assets as resistance.
- **D-9 — Every XP formula rounds up.** Robert's rules call; the rules text gives no rounding. `G-004` pays the ceiling of (Force + Cunning + Wealth) / 6, minimum 1, so ratings 3/4/4 pay 2. `G-008` pays 1 + the ceiling of the target's average rating. `G-001`–`G-003` pay the ceiling of half the kills, so 3 kills pay 2. Rejected: today's floors (integer division), rounding up for `G-004` only, and rounding averages up while half-kills stay floored.
- **D-10 — `G-009` counts the assets the faction holds, not stealth events.** Robert's rules call, on "Have stealthed assets on worlds with other governments equal to Cunning". The goal completes when, at one time, the faction holds Cunning-many stealthed assets on rival ground. The goal records the IDs of the assets already stealthed when it starts, and those never count. This reverses frozen log D121, which counted `AssetStealthApplied` events, so an asset that lost stealth or left still counted and one stealthed elsewhere then moved in never did. Rejected: keeping the event count.
- **D-11 — `G-011`'s gate compares against Wealth at its last completion.** Robert's rules call, on "The faction's Wealth rating must increase before this goal can be selected again". When `G-011` completes, the faction records its Wealth; the pick offers `G-011` again only when Wealth is above that record. Abandoning sets no gate. Rejected: a record written on abandon too, and Wealth at the time of the last pick.
- **D-12 — A goal can name the actions it needs, and the pick offers it only when they are registered.** Robert's call, from a side-agent review. Under D-1 nothing can complete `G-004` in the gap: the `G-004` handler is the only producer of `WorldGovernmentChanged`, and it goes away. So `G-004`'s `start` part names `seize_planet`, and the turn-start pick and the faction wizard leave the goal out while that action is not registered. When `F.040` registers Seize Planet again, World Seizure comes back with no data change, which matters because each campaign holds its own copy of the rulebook. This is a second kind of pick gate, next to `G-011`'s Wealth gate (D-11). *(The wizard half was superseded by D-25: the wizard offers no goal.)* Rejected: a `selectable = false` flag that `F.040` removes, which leaves a stale flag in every existing campaign's copy.
- **D-13 — Every faction's goal reads every mutation batch, on any turn and in any phase.** Robert's rules call. Today `UpdateProgress` runs only for the acting faction, on its action batch, so a kill or damage dealt while defending on a rival's turn, or anything in a movement batch, never counts. The rules text ("destroy", "inflict") names no turn. Each goal filters the batch by what its faction caused. A goal can complete on a rival's turn; the faction picks its next goal at the start of its own next turn (D-2). Rejected: every batch of the faction's own turn only, and today's own action batch.
- **D-14 — Any kill or damage the faction causes counts.** Robert's rules call, for `G-001`–`G-003`, `G-006` and `G-010`. A mutation counts when its `CausedByFactionID` is the faction, whatever the cause: an attack, a counterstrike, an ability or an effect. Today only cause `attack` counts. Rejected: combat only (attack and counterstrike), and attack only.
- **D-15 — Destroy the Foe completes only when the faction itself brings the target to 0 HP.** Robert's rules call. The HP loss that takes the target faction to 0 must be caused by the faction. A third party finishing the target off ends the goal without completing it; the mutation that records that end is still to settle. Rejected: completing whoever caused the loss.
- **D-16 — A pass turn does not count for Peaceable Kingdom.** Robert's rules call. Only an own turn where the faction takes an action other than Attack advances `G-007`; a turn with no action at all neither advances nor resets it. This keeps today's behavior, which the code has as a side effect (`runActionPhase` returns before goal progress on a pass), as a rule. Rejected: counting a pass as a peaceful turn.
- **D-17 — A goal's parts are TOML sub-tables.** Robert's call. Each part is its own table (`[goals.MilitaryConquest.progress]`), with nested tables for its filters (`[goals.MilitaryConquest.progress.filters]`), the same layout as an `[[effects]]` block. Rejected: inline tables, which get hard to read once filters nest.
- **D-18 — Goals have their own event names and filter set, owned by the goal package.** Robert's call. A goal event is a matcher over a finished batch; an effect trigger picks a hook route and a registrar. They are different systems, and a shared name makes each change on one side a question for the other (effects' `asset_destroyed` already means combat losses only, while D-14 counts any cause). Goals share only the sub-table layout (D-17), the value words (`own`, `rival`, the stat names) and strict loading with `[[effects]]`; no goal matcher calls into `engine/effect`, and Scavengers' data does not change. Rejected: one shared vocabulary with a `cause` filter, and shared names with two meanings.
- **D-19 — Each action batch opens with a record-only `ActionTaken` mutation (`B.013`).** Robert's call. Nothing in a batch names the action today, so Peaceable Kingdom infers an Attack from damage or stealth loss and misses a clean miss. `ActionTaken{FactionID, Action}` changes no state; like `GoalInitiated` and `PermissionRefused`, it gets a no-op apply arm, a digest line and a TUI line (effect-mutation Key Decisions). Goals read it as an event, so they still read batches and nothing else (D-13), and history records the action, not only its effects. It also removes the `tick` part: `G-007` progresses on an own `action_taken` that is not an Attack and resets on one that is, and a pass emits none, which gives D-16 by its shape. Rejected: passing the action name to the goal engine as a new input, a special case for the acting faction's own batch.
- **D-20 — Blood the Enemy counts damage to a rival's homeworld base.** Robert's rules call. A hit on a homeworld base emits `FactionHPDelta`, not `BaseHPDelta` (frozen log D75), so today's `G-006` misses it; the rules say "enemy faction assets or bases" and name no exception. The `damage` event also matches a negative `FactionHPDelta` against a rival that the faction caused. Rejected: counting only `AssetHPDelta` and `BaseHPDelta`.
- **D-21 — A lost target ends a goal with a new `GoalFailed` mutation.** Robert's call, for D-15. A goal with a `target_faction` input ends when its target is destroyed and its own `complete` did not match: a built-in rule, not a part. `GoalFailed{FactionID, GoalID, Cause}` (cause `target_lost`) clears `ActiveGoal`, pays no XP and costs no income; it has its own apply arm, digest line and TUI line, so history and the digest tell a lost target apart from a GM's Abandon Goal. Rejected: `GoalAbandoned` with a cause, and a `fail` part only `G-008` would use.
- **D-22 — Expand Influence is contested when a rival asset or base is on the world.** Robert's rules call, keeping today's reading of "+1 if contested by a rival faction". Rejected: only a rival that opposed the attempt itself.
- **D-23 — `start` has no `kind` key.** Robert's call. Every goal is picked (D-1, D-5), so `kind = "select"` would say nothing; it comes back only if a later goal starts another way.
- **D-24 — `ActiveGoal` loses `TargetWorld`, `ProcessPhase` and `TurnsRemaining`.** Robert's call. They have no goal reader once the process moves to `F.040` (D-1), and `F.040` keeps its process state on the action, not the goal. Rejected: leaving them unused until `F.040`.
- **D-25 — The faction wizard sets no goal (`B.041`).** Robert's call. Every new faction starts with no goal and picks one at its first turn through the turn-start pick (D-2), so the pick is the one place a goal is chosen. This also removes a gap D-3 would open: the wizard builds factions one at a time, so the first faction has no rival to name as Destroy the Foe's target. The rules' setup step "Select starting goal" moves to the faction's first turn. `goalsToOptions` and the wizard's goal step go. Rejected: a wizard that offers only goals with no start inputs, and a wizard that offers every goal and asks its inputs.
- **D-26 — Saves move to prefixed goal IDs by a sandbox re-seed, with no migration code.** Robert's call, following the `A.001.7` precedent (persistence page, Key Decisions). Goal IDs become `CORE-G-001` to `CORE-G-011` in `_core/goals.toml`, and the two tag references (Exchange Consulate's `G-007`, Imperialists' `G-004`) follow, checked at load. `sandbox-claude` is re-seeded by Claude and `sandbox` by Robert. Rejected: migration code that renames IDs in state, which would serve only the two sandbox campaigns.
- **D-27 — Three efforts: goals as data, every batch, then the turn-start pick.** Robert's call. Each boundary is a separate design surface and a review point.
  - **E1 — Goals as data.** The vocabulary, the part loader and its checks, `CORE-G-` IDs, the eleven goals in data, `ActionTaken`, the per-ID handlers deleted, Seize Planet and Change Homeworld unregistered, `ActiveGoal` slimmed, the re-seed. Goals still read only the acting faction's action batch, so behavior changes only where the data does.
  - **E2 — Every batch.** Goals read every batch in every phase for every faction (D-13), and `GoalFailed` (D-21). An orchestrator change across every phase, reviewed apart from the data swap.
  - **E3 — Turn-start pick.** The pick step and its `PhaseCollector` ask, start inputs, the `G-009` start record, both gates and `G-011`'s completion record, the TUI overlay, and the wizard's goal step removed (D-25).
  - **Order: E1, E2, E3.** The pick lands on the final engine behavior. Between E1 and E3 the wizard still sets goals and `G-008` still has no target, as today, so `B.040` closes at E3. Rejected: two efforts (the data swap mixed with the orchestrator change), one effort, and the pick before every-batch reading.

## Open Questions

| Question | Source | Status |
|---|---|---|
| `G-004` XP when no faction resists: the text says "counts as difficulty low"; Go gives 1 only when no faction rules the world | `E.002` rules call | Settled: D-8, D-9 |
| `G-009` count: the rules say "have Stealthed assets equal to Cunning"; Go counts stealth events, so an asset that loses stealth still counts | `E.002` rules call | Settled: D-10 |
| `G-011` gate: what "Wealth must increase before this goal can be selected again" compares against | `E.002` rules call | Settled: D-11 |
| The exact part vocabulary and its TOML shape | `E.002` G-1 sketch | Settled: Goal Vocabulary, below |
| Where the turn-start goal pick sits in the turn, and the shape of its collector ask | `B.039` | Settled: D-2, D-3, D-4 |
| Whether the TUI's goal-select overlay is in this initiative | `B.039` | Settled: D-6 |
| `G-012` as a goal once Change Homeworld is an action: it is not in the rules' goal list, and pays 0 XP | D-1 | Settled: D-5 |
| The two actions and in-flight saves during the gap | D-1 | Settled: D-1 |
| The mutation that ends `G-008` when a third party destroys the target | D-15 | Settled: D-21 |
| Goal filter vocabulary: reuse the `[[effects]]` trigger and filter names, or a goal-only set | Vocabulary | Settled: D-18 |
| Goal TOML shape: inline tables or sub-tables | Vocabulary | Settled: D-17 |
| The multi-turn action row: in-arc (`A.001.9`) or a standalone `F.` row | D-1 | Settled: D-7 |
| How `UpdateProgress` learns the action taken (`B.013`) | `E.002` | Settled: D-19 |
| What the faction wizard offers at creation (`B.041`) | `B.041` | Settled: D-25 |
| Save migration for the prefixed IDs | I-1 | Settled: D-26 |
| Effort slicing | Arc-plan session shape | Settled: D-27 |

---

## Goal Vocabulary

The parts, events, filters and formulas below state all eleven goals. Names follow D-18: goal-only event names, the `[[effects]]` value words (`own`, `rival`, stat names, action display names such as `"Attack"`), sub-tables (D-17). Every name is checked at engine construction, as an effect name is.

### Parts

| Part | Job | Keys |
|---|---|---|
| `start` | What the pick needs and when it may offer the goal. Optional; every goal is picked (D-1 leaves no action-started goal) | `inputs` (`target_faction`), `needs_actions`, `gate` |
| `progress` | How progress moves: add an amount per matching event, or set it to a count read from state after each batch | `event` + `amount` + `filters`, or `measure` + `filters` |
| `reset` | An event that sets progress to 0 | `event` + `filters` |
| `threshold` | The progress that completes the goal, from the holder's own ratings | `ratings` (summed), `multiplier`, or `constant` |
| `complete` | What completes the goal: progress at the threshold (the default when `threshold` is present), or the first matching event | `event` + `filters` |
| `xp` | The XP paid on completion, computed before the batch applies (frozen log D119); every formula rounds up (D-9) | `formula` + its parameters |

### Events

Each event is a matcher over a batch's mutations (D-13: every batch, any turn, any phase).

| Event | Matches | Filters | Amounts |
|---|---|---|---|
| `kill` | `AssetRemoved` | `owner`, `caused_by`, `category`, `rated_above` | `one` |
| `damage` | `AssetHPDelta`, `BaseHPDelta` and `FactionHPDelta` (a homeworld base hit, D-20) with a negative delta | `owner`, `caused_by` | `damage` |
| `base_planted` | `BaseAdded` | `caused_by`, `new_world` | `one` |
| `influence_bought` | `InfluenceDelta` with a positive delta | `caused_by` | `delta` |
| `government_taken` | `WorldGovernmentChanged` | `to` | `one` |
| `faction_destroyed` | `FactionHPDelta` that takes a faction to 0 HP or less, against its HP before the batch plus the batch's earlier deltas | `target`, `caused_by` | `one` |
| `action_taken` | `ActionTaken` (D-19) | `faction`, `action`, `action_not` | `one` |

### Filters

| Filter | Values | Reads |
|---|---|---|
| `owner` | `own`, `rival` | Whose asset or base it was, relative to the goal's faction |
| `caused_by` | `own`, `rival` | The mutation's `CausedByFactionID`; any cause counts (D-14) |
| `faction` | `own` | Whose action it was |
| `category` | stat names | The removed asset's definition category, read from live state before the batch applies |
| `rated_above` | a stat name | The removed asset's `MinRating` is above the goal faction's rating in that stat |
| `new_world` | `true` | The faction held no base on that world before the batch |
| `to` | `own` | The new ruler |
| `target` | `start_input` | The faction named by `start.inputs` |
| `action`, `action_not` | action display names | The action taken is, or is not, in the list |

### Measures

| Measure | Counts | Filters |
|---|---|---|
| `stealthed_assets` | The faction's stealthed assets now in play | `ground = "rival"` (a world a rival governs, or an independent rival homeworld; today's `isEnemyTerritory`), `exclude_held_at_start = true` (the pick records the IDs of the assets stealthed when the goal starts, D-10) |

### XP formulas

| Formula | Pays | Used by |
|---|---|---|
| `constant` | `amount`, plus `contested_bonus` when a rival asset or base is on the event's world | `G-005`–`G-007`, `G-009`–`G-011` |
| `half_progress` | ⌈progress / 2⌉ | `G-001`–`G-003` |
| `target_average` | 1 + ⌈the target faction's average rating⌉ | `G-008` |
| `ruler_average` | ⌈(the previous ruler's Force + Cunning + Wealth) / 6⌉, minimum 1, when the ruler resisted; else 1 (D-8, D-9). Reads the ruler and the resist flag from the `government_taken` event | `G-004` |

### Gates

| Gate | Opens when | Used by |
|---|---|---|
| `needs_actions` | Every named action is registered (D-12) | `G-004` |
| `rating_above_last_completion` | The named rating is above its value when the faction last completed this goal (D-11) | `G-011` |

### The eleven goals

```toml
[goals.MilitaryConquest]                  # G-002, G-003: category wealth / cunning
id          = "CORE-G-001"
name        = "Military Conquest"
description = "Destroy a number of Force assets of rival factions equal to your faction's Force rating."
[goals.MilitaryConquest.progress]
event  = "kill"
amount = "one"
[goals.MilitaryConquest.progress.filters]
owner     = "rival"
caused_by = "own"
category  = ["force"]
[goals.MilitaryConquest.threshold]
ratings = ["force"]
[goals.MilitaryConquest.xp]
formula = "half_progress"

[goals.WorldSeizure]
id = "CORE-G-004"
[goals.WorldSeizure.start]
needs_actions = ["Seize Planet"]
[goals.WorldSeizure.complete]
event = "government_taken"
[goals.WorldSeizure.complete.filters]
to = "own"
[goals.WorldSeizure.xp]
formula = "ruler_average"

[goals.ExpandInfluence]
id = "CORE-G-005"
[goals.ExpandInfluence.complete]
event = "base_planted"
[goals.ExpandInfluence.complete.filters]
caused_by = "own"
new_world = true
[goals.ExpandInfluence.xp]
formula         = "constant"
amount          = 1
contested_bonus = 1

[goals.BloodTheEnemy]
id = "CORE-G-006"
[goals.BloodTheEnemy.progress]
event  = "damage"
amount = "damage"
[goals.BloodTheEnemy.progress.filters]
owner     = "rival"
caused_by = "own"
[goals.BloodTheEnemy.threshold]
ratings = ["force", "cunning", "wealth"]
[goals.BloodTheEnemy.xp]
formula = "constant"
amount  = 2

[goals.PeaceableKingdom]
id = "CORE-G-007"
[goals.PeaceableKingdom.progress]
event  = "action_taken"
amount = "one"
[goals.PeaceableKingdom.progress.filters]
faction    = "own"
action_not = ["Attack"]
[goals.PeaceableKingdom.reset]
event = "action_taken"
[goals.PeaceableKingdom.reset.filters]
faction = "own"
action  = ["Attack"]
[goals.PeaceableKingdom.threshold]
constant = 4
[goals.PeaceableKingdom.xp]
formula = "constant"
amount  = 1

[goals.DestroyTheFoe]
id = "CORE-G-008"
[goals.DestroyTheFoe.start]
inputs = ["target_faction"]
[goals.DestroyTheFoe.complete]
event = "faction_destroyed"
[goals.DestroyTheFoe.complete.filters]
target    = "start_input"
caused_by = "own"
[goals.DestroyTheFoe.xp]
formula = "target_average"

[goals.InsideEnemyTerritory]
id = "CORE-G-009"
[goals.InsideEnemyTerritory.progress]
measure = "stealthed_assets"
[goals.InsideEnemyTerritory.progress.filters]
ground                = "rival"
exclude_held_at_start = true
[goals.InsideEnemyTerritory.threshold]
ratings = ["cunning"]
[goals.InsideEnemyTerritory.xp]
formula = "constant"
amount  = 2

[goals.InvincibleValor]
id = "CORE-G-010"
[goals.InvincibleValor.complete]
event = "kill"
[goals.InvincibleValor.complete.filters]
owner       = "rival"
caused_by   = "own"
category    = ["force"]
rated_above = "force"
[goals.InvincibleValor.xp]
formula = "constant"
amount  = 2

[goals.WealthOfWorlds]
id = "CORE-G-011"
[goals.WealthOfWorlds.start.gate]
rating_above_last_completion = "wealth"
[goals.WealthOfWorlds.progress]
event  = "influence_bought"
amount = "delta"
[goals.WealthOfWorlds.progress.filters]
caused_by = "own"
[goals.WealthOfWorlds.threshold]
ratings    = ["wealth"]
multiplier = 4
[goals.WealthOfWorlds.xp]
formula = "constant"
amount  = 2
```

`name` and `description` are shown once; every goal keeps them. `difficulty` goes: the `xp` part replaces it (`B.005`).

### State the parts need

- **`ActiveGoal`** keeps `GoalID`, `Progress` and `TargetFactionID`, and gains the asset IDs held stealthed at start (`G-009`). `TargetWorld`, `ProcessPhase` and `TurnsRemaining` have no goal reader once the process moves to `F.040` (D-1).
- **`Faction`** gains a record of its ratings at each goal's last completion, written when `GoalCompleted` applies (`G-011`).
- **`WorldGovernmentChanged`** gains the resist flag that `ruler_average` reads; `F.040` fills it (D-8).

### Built-in rules

- **A lost target fails the goal (D-21).** A goal with a `target_faction` input emits `GoalFailed` (cause `target_lost`) when its target is destroyed and its own `complete` did not match.

### Review points

All five settled: `G-006` and the homeworld (D-20), `G-008`'s end (D-21), `G-005`'s "contested" (D-22), no `start.kind` (D-23), `ActiveGoal`'s process fields (D-24).

## User-Facing Impact

- **A goal pick at turn start.** A faction with no goal sees a goal-select overlay after its stat raise, with the goals it may take and any start input (a target faction for Destroy the Foe); it may decline.
- **No goal in the wizard.** A new faction starts with no goal and picks at its first turn.
- **Seize Planet and Change Homeworld leave the action menu**, and World Seizure leaves the goal pick, until `F.040` lands.
- **Goals advance from more events.** Kills and damage on rivals' turns, from any cause, and hits on a homeworld base now count; Inside Enemy Territory tracks assets held, not stealth events.
- **More XP in the odd cases.** Every formula rounds up.
- **New lines in history, the digest and the TUI** for `ActionTaken` and `GoalFailed`.

## Replaces / Retires

- The twelve per-ID handlers in `engine/goal/goals/`, the `Handler` registry and its silent skip (frozen log D111 reversed).
- The goal side of `CheckLock` and the goal use of `locks`; the locks move with the process to `F.040`.
- `G-012` in `goals.toml`, `difficulty` on every goal record (`B.005`), and `ActiveGoal`'s `TargetWorld`, `ProcessPhase` and `TurnsRemaining`.
- The `"G-004"` and `"G-012"` literals in `seize_planet.go` and `change_homeworld.go` (`B.001`), and the bare goal IDs in tag data.
- The wizard's goal step and `goalsToOptions` (`B.041`).
- Closes `B.001`, `B.005`, `B.013`, `B.039`, `B.040` and `B.041`, and the goal half of `R.004`.

## Out of Scope

- **Multi-turn actions** — `F.040`: Seize Planet and Change Homeworld as actions, the seizure's completion, and the resist record D-8 reads.
  - **Imperialists goes with it.** Its `bonus_die` entry filters on holding World Seizure (`goal = "G-004"`), but the rule is "attacks made as part of a Seize Planet action". It is inert in the gap, because no seizure can happen. `A.001.8` only renames the reference to `CORE-G-004` (D-26), so the data still loads. `F.040` re-keys it to a seizure-in-progress filter, which also stops it firing in phase 0 and phase 2. Found by a side-agent review.
- **Goal state predicates** — `F.013`: a `complete` predicate kind plugs in beside `threshold` and `event`.
- **Goal start and phase lines** — `F.033`: its `GoalInitiated` and `GoalPhaseAdvanced` belong to the process, which moves to `F.040`.
- **Scavengers' combat-only reading.** Effects' `asset_destroyed` keeps matching attack and counterstrike losses only; D-18 keeps the goal side separate.
- **The rules doc.** `docs/rules/swn-faction-mechanics.md` takes the readings in D-8 to D-11, D-14 to D-16 and D-20 to D-22 when the work lands; Plan places that edit.
