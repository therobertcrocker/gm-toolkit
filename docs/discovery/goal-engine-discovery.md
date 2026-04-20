# Goal Engine — Discovery & Planning

The Goal Engine manages multi-turn faction objectives. It tracks progress toward a faction's current goal, applies turn-by-turn constraints (including action locks), and resolves completion conditions. It is a first-class part of the turn engine — consulted at the start of each faction's turn to determine what actions are available or locked.

When an AI agent is driving faction decisions, the Goal Engine will be the primary influence on which actions the agent considers — playing toward the active goal or in-character based on faction personality and tags.

---

## Responsibilities

- Track active goal and progress state per faction
- Lock or constrain available actions based on goal state (e.g. faction is mid-move, mid-seize)
- Decrement multi-turn counters each turn
- Detect and resolve goal completion, awarding XP
- Influence (but not fully own) action selection for AI-driven factions

---

## Multi-Turn Processes

These were initially considered actions but are better modeled as goal-engine-managed stateful processes, since they span multiple turns and lock faction behavior.

---

### Change Homeworld

**Complexity:** Medium — no dice, but introduces a multi-turn action lock.

#### Inputs
- Faction performing the action
- World selected as new homeworld
- Number of turns for the move (manually entered for now; will be calculated automatically when map engine is implemented)

#### Resolution Steps
1. GM selects destination world (must have an existing Base of Influence)
2. GM enters hex distance; calculate move duration: `1 + hex distance` turns
3. Record move-in-progress on faction state: destination world and turns remaining
4. Lock faction from taking actions for the duration of the move
5. Each subsequent turn: decrement turns remaining; skip action selection for this faction
6. On final turn: update homeworld to destination, remove move lock

#### Outputs
- Faction homeworld updated on move completion
- Faction action-locked for move duration

#### Notes
- **Depends on Map Engine (planned)** — hex distance will eventually be calculated automatically; for now the GM enters it manually
- Faction is completely skipped during action selection while move is in progress
- Validation ensures destination world has an existing Base of Influence

---

### Seize Planet

**Complexity:** High — multi-phase, multi-turn process with combat and occupation stages.

Seize Planet is a campaign objective, not a single action. The faction's available actions are constrained by the Goal Engine until the objective completes.

#### Phases

**Phase 1 — Combat:** Faction must destroy all unstealthed opposing assets on the target world. Each turn the faction may only take Attack actions targeting that world. This phase spans as many turns as needed.

**Phase 2 — Occupation:** Once all resistance is eliminated, the faction must maintain at least one unstealthed asset on the world for three consecutive turns. No other constraint on actions during this phase.

**Phase 3 — Resolution:** On occupation completion, award the **Planetary Government** tag for that world.

#### Notes
- **Depends on Attack resolution** — combat phase delegates to Attack action logic
- Goal Engine tracks current phase, target world, and occupation turn counter
- If the faction loses all unstealthed assets on the world during occupation, the seize fails

---
