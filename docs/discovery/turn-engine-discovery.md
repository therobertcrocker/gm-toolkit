# Turn Engine — Discovery & Planning

A breakdown of the features and sub-tasks required to build the faction turn engine.

---

## 1. Turn Scaffolding

The infrastructure for ordering, stepping through, pausing, and resuming a faction turn.

### Sub-tasks

1. **Faction ordering** — at turn start, roll a die no smaller than the number of factions to determine which faction acts first; sequence the rest in order from there
2. **Turn state** — a structure that tracks where we are mid-turn: which factions have gone, which is current, which are pending
3. **Persist mid-turn state** — save turn state to the campaign file so a paused turn survives a restart
4. **Resume detection** — on startup, detect if an in-progress turn exists and offer to resume or abandon it
5. **Step controller** — the logic that advances from one faction to the next and knows when the turn is complete

---

## 2. Per-Faction Bookkeeping

Automatic calculations that occur at the start of each faction's turn, before an action is taken.

### Sub-tasks

1. **Goal selection** — if the faction has no current goal, prompt the GM to select one before proceeding
2. **Income calculation** — apply the Coin formula (`floor(Wealth / 2) + floor((Force + Cunning) / 4)`) and add result to the faction's Coin balance
3. **Maintenance payment** — deduct maintenance costs for all assets that require it
4. **Unpaid asset tracking** — mark assets as unusable if maintenance cannot be paid
5. **Asset loss** — destroy assets that have been unpaid for two consecutive turns

---

## 3. Action Selection

Determines which of the 9 actions are available to the current faction and presents them to the GM. Unavailable actions are filtered out entirely.

### Sub-tasks

1. **Validate available actions** — evaluate each action against current faction state using the rules below
2. **Present filtered action list** — show only valid actions to the GM for selection

### Validation Rules

| Action | Available if... |
|--------|----------------|
| **Attack** | Faction has at least one asset on the same world as a rival's known (non-stealthed) asset |
| **Buy Asset** | Faction has Coin ≥ cheapest available asset; has a Base of Influence (or is on homeworld); at least one asset exists that meets attribute and tech level requirements |
| **Change Homeworld** | Faction has a Base of Influence on at least one world that isn't the current homeworld |
| **Expand Influence** | Faction has at least one asset on a world where it doesn't already have a Base of Influence, and has enough Coin to purchase at least 1 HP |
| **Refit Asset** | Faction has at least one asset; a valid refit target exists (same type, meets requirements) |
| **Repair Asset/Faction** | Faction has at least one damaged asset OR faction HP is below max; has at least 1 Coin |
| **Sell Asset** | Faction has at least one asset (Bases of Influence excluded) |
| **Seize Planet** | Faction has at least one unstealthed asset on a world with a rival government; no seize already in progress on a different world |
| **Use Asset Ability** | Faction has at least one asset with the `A` flag that is usable (not unmaintained) |

---

## 4. Action Resolution

Each action implements a common interface. The Action Engine calls these in order and provides shared utilities as dependencies — it never needs to know what a specific action does, only that it implements the contract.

### The Action Interface

| Method | Responsibility |
|--------|---------------|
| **Inputs** | Defines what the engine must collect from the GM (or AI agent) before resolution begins — selected assets, target world, etc. |
| **Validate** | Confirms preconditions are met; gates whether the action is available at all |
| **Resolve** | Executes the action-specific logic; calls shared utilities as needed |
| **Output** | Produces a mutation list and an event record; the engine applies both at turn end |

### Shared Utilities

- **Dice roller** — shared roll primitive
- **Faction test** — `1d10 + acting attribute` vs. `1d10 + target attribute`; ties to defender; not an attack
- **Tag Engine** — evaluates tag relevance and applies +1d10 keep highest modifier
- **Ability Engine** — resolves named asset abilities (called from Use Asset Ability)

### Design Note

When an AI agent eventually drives faction decisions, **Inputs** is replaced by the agent's goal-driven selection logic. Validate, Resolve, and Output remain unchanged. The action itself has no knowledge of whether it is being driven by a human or an agent.

> Full breakdown of each action in [action-resolution-discovery.md](action-resolution-discovery.md).

---

## 5. State Mutation

Actions do not mutate state directly. Each action's Output method produces a **mutation list** — a structured description of what should change. The engine applies mutations atomically to campaign state and then persists to disk.

### Mutation Types

| Mutation | Triggered By |
|----------|-------------|
| Faction Coin balance updated | Income, maintenance, Buy Asset, Sell Asset, Refit Asset, Repair Asset, Expand Influence |
| Asset HP updated | Attack, Repair Asset |
| Asset added | Buy Asset, Refit Asset |
| Asset removed | Sell Asset, Refit Asset, Attack (on destruction) |
| Asset flags updated | Buy Asset, Refit Asset (inactive), Attack (stealth lost), bookkeeping (unmaintained) |
| Faction HP updated | Attack (Base damage), Repair Faction |
| Base of Influence HP updated or placed | Attack (damage redirected), Expand Influence |
| Faction action lock set or cleared | Goal Engine (Change Homeworld, Seize Planet) |
| Tag added or removed | Goal Engine (Planetary Government on Seize Planet completion) |
| XP awarded / attribute rating updated | Goal Engine (goal completion) |

### Design Notes

- Mutations are applied in order; the engine is the sole writer to campaign state
- After all mutations are applied, state is persisted to disk in a single write
- This makes actions fully testable — given inputs, assert the mutation list; no state required

---

## 6. Event Recording

Every turn produces two outputs committed together at turn end: a **mutation list** (applied to campaign state) and an **event record** (appended to history). Nothing is written until the turn completes — a paused mid-turn produces no partial writes, keeping state and history always in sync.

### Event Record Contents

The event record captures everything that happened during the turn at maximum granularity — far richer than the mutation list, which only tracks what changed in state.

| Category | What is recorded |
|----------|-----------------|
| **Turn metadata** | Turn number, faction order, timestamp |
| **Bookkeeping** | Income collected (with formula breakdown), maintenance paid per asset, assets marked unmaintained or lost, goal selected |
| **Action taken** | Action type, faction, inputs chosen (target world, selected assets, etc.) |
| **Roll detail** | Each die rolled, attribute bonus applied, tag applied and why, final totals |
| **Resolution detail** | Per-matchup outcomes, damage dealt, counterattack damage, stealth lost, assets destroyed |
| **State changes** | Coin deltas, HP changes, assets added/removed, flags changed, tags awarded, XP gained |

### Format

Append-only JSONL (`campaigns/<id>/history.jsonl`). One JSON object per turn. Never mutated after writing.

### Commit Order

1. Turn completes (all factions have acted)
2. Mutation list applied to campaign state
3. State persisted to `faction_state.toml`
4. Event record appended to `history.jsonl`

### Design Notes

- The event record is the source of truth for the narrative renderer and future AI agent memory
- Recording everything is intentional — more data is better; the renderer filters what to surface
- Pub/sub may be appropriate here in the future if multiple consumers need to react to turn events (narrative renderer, AI observer, etc.)

---

## 7. Goal Engine

The Goal Engine is a first-class part of the turn engine. It manages multi-turn faction objectives, tracks progress, applies action locks, and resolves completion conditions. It is consulted at the start of each faction's turn before action selection. Change Homeworld and Seize Planet are managed here rather than in the action engine.

> Full breakdown in [goal-engine-discovery.md](goal-engine-discovery.md).

---

## 8. Tag Engine

The Tag Engine evaluates tag relevance in the context of a roll or action and applies the +1d10 keep highest modifier when applicable. It is a shared utility called from within the action engine — not owned by any single action. Tags surface in Attack rolls, faction tests, and potentially other actions as the engine grows.

> Discovery doc to be written when Tag Engine scope is better understood.

---
