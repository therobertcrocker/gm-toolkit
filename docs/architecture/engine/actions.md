# Actions

> **Code:** `internal/faction/engine/action/{action,collector}.go`, `internal/faction/engine/action/actions/`

## Purpose

The action engine resolves the one discretionary thing a faction does each turn:
its chosen action — Attack, Buy Asset, Expand Influence, Use Asset Ability, and
eight others. Every action implements a four-method contract (`Validate` →
`Inputs` → `Resolve` → `Output`), and the engine's job is to offer the GM the
actions a faction is *eligible* for and then run the selected one to a list of
mutations. Like the rest of the engine it emits mutations and never applies them.

## Shape

### The `Action` contract

`action.go` defines the interface every action satisfies:

```go
type Action interface {
	Name() string
	Validate(faction, factionState, rulebook) bool
	Inputs(faction, factionState, rulebook) error
	Resolve(faction, factionState, rulebook) error
	Output() ([]domain.Mutation, error)
}
```

The four working methods run in a fixed sequence with distinct jobs:

- **`Validate`** — cheap, pure eligibility check (does the faction have a base, a
  ready asset, the Coin?). Used to build the availability list; must not prompt.
- **`Inputs`** — gather GM decisions through the `Collector` (which asset, which
  target, how much Coin).
- **`Resolve`** — compute the outcome (rolls, comparisons, cascades).
- **`Output`** — return the resulting `[]domain.Mutation`.

The split is deliberately loose about *where* the work sits. `Bribe` does
everything in `Inputs` and emits a fixed `CoinDelta` + `InfluenceDelta` pair from
`Output`, with an empty `Resolve`. `UseAssetAbility` selects assets in `Inputs`
and does the real work in `Resolve`. Same contract, different center of gravity.

### Factory registration

Actions are not registered as singletons but as **factories** —
`ActionFactory func(Collector) Action`. `RegisterDefaultActions`
(`actions/register.go`) wires the twelve standard SWN actions into the engine,
each as a closure that builds a fresh action bound to the turn's collector.
`ActionEngine` holds them as an ordered `[]ActionFactory`.

`RegisterDefaultActions` takes its dependencies — the roller resolver, the hook
registry, the world engine, the rulebook — as **explicit parameters** rather than
reaching into the parent engine, so the `actions` package never imports `engine`.
That is what lets `engine` own registration without an import cycle.

### Two passes per turn

`AvailableActions` instantiates *every* factory with the collector and keeps the
ones whose `Validate` passes — so each turn gets a fresh, freshly-wired set of
candidate actions. The orchestrator presents that list, the GM picks one, and
`Run` calls `Inputs` → `Resolve` → `Output` on the selected action, returning its
mutations. The orchestrator then folds in goal-progress mutations before
dispatching reactor hooks (see [orchestrator](orchestrator.md)).

### The collector seam

`collector.go` defines `action.Collector`, the prompt surface the actions call
into. It embeds `hooks.Collector` and adds sixteen action-specific methods —
`SelectAttackers`, `SelectBuyOrder`, `SelectBribeTarget`,
`SelectExpandInfluenceOrder`, and so on — plus the small order/option structs
those methods exchange (`BuyOrder`, `RefitOrder`, `ExpandInfluenceOrder`, …).
This is the richer of the engine's [two input contracts](orchestrator.md): the
orchestrator threads it down here but drives the GM through its own
`PhaseCollector`.

### Eligibility helpers and the ability fold

`actions/eligibility.go` holds the shared targeting predicates the combat-shaped
actions reuse — `eligibleAttackers` (ready, alive, maintained),
`eligibleDefendersOnWorld` / `eligibleDefendersAtHex` (opposing, non-stealthy,
reachable). The hex variant exists because an attacker can be mid-flight with no
world ID, indexed only by hex.

Asset *abilities* are not a top-level sub-engine — they are folded into the
action layer as the `actions/ability/` subpackage. The `Use Asset Ability` action
gathers the A-flagged assets, then loops `ability.Dispatch` over them; each
dispatch resolves one asset's ability into mutations that fold into the action's
output.

## Actions catalogue

The twelve standard actions wired by `RegisterDefaultActions`. "Prompts" are the
`Collector` methods each calls; "emits" lists the mutation types it can produce
(parenthesised ones are conditional). A `—` prompt means the action is fully
determined once selected. The rules behind the numbers live in
[`swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md).

| Action | What it does | Prompts → emits |
|--------|--------------|-----------------|
| Sell Asset | Remove an asset for half its Cost back | `SelectAsset` → `AssetRemoved`, `CoinDelta` |
| Repair Faction | Heal the faction's own HP, free | — → `FactionHPDelta` |
| Repair Asset | Heal damaged assets; cost escalates per heal on the same asset | `SelectRepairOrders` → `AssetHPDelta`, `CoinDelta` |
| Buy Asset | Purchase an asset onto a world, inactive until next turn | `SelectBuyOrder` (`SelectAsset` for a stealth target) → `AssetAdded`, `CoinDelta` (`AssetStealthApplied`) |
| Refit Asset | Swap an asset for another of its category, paying the cost difference | `SelectRefitOrder` → `AssetRemoved`, `AssetAdded`, `CoinDelta` |
| Attack | Resolve asset combat against rivals at a world/hex; damage may redirect to a base | `SelectAttackers`, `SelectDefender`, `ConfirmRedirectToBase` → `AssetHPDelta`, `AssetRemoved`, `BaseHPDelta`, `FactionHPDelta`, `BaseDestroyed`, `AssetStealthCleared` |
| Expand Influence | Place or reinforce a Base of Influence; a new base draws rival free attacks | `SelectExpandInfluenceOrder`, `ConfirmRivalFreeAttack`, `SelectBaseAttackers` → `BaseAdded`, `BaseHealed`, `BaseExpanded`, `CoinDelta`, `BaseHPDelta`, `FactionHPDelta`, `BaseDestroyed` |
| Bribe | Spend Coin to raise a rival Base's influence | `SelectBribeTarget` → `CoinDelta`, `InfluenceDelta` |
| Use Asset Ability | Activate the special ability of one or more A-flagged assets | `SelectAbilityAssets` (+ ability-specific prompts) → varies by ability |
| Abandon Goal | Drop the active goal, forfeiting this turn's income | — → `CoinDelta`, `GoalAbandoned` |
| Seize Planet | Begin the Planetary Seizure goal (G-004) on a contested world | `SelectSeizeTarget` → `GoalInitiated` |
| Change Homeworld | Begin the Change Homeworld goal (G-012) toward an owned base world | `SelectChangeHomeworldTarget` → `GoalInitiated`, `GoalPhaseAdvanced` |

`Seize Planet` and `Change Homeworld` are *goal-initiating* actions — they don't
resolve an outcome so much as start a multi-turn [goal](goals.md); `Abandon Goal`
is the inverse. The combat rows (`Attack`, `Expand Influence`) emit the widest
mutation sets because their rolls run through the [hooks](hooks.md) dispatch and
can cascade to bases and faction HP.

## Key Decisions

- **Actions are factories, instantiated fresh each turn.** `AvailableActions`
  rebuilds every candidate from its factory and binds it to the current
  collector, so an action never carries state between turns and always talks to
  the live prompt surface. (Frozen log 46–53.)
- **`RegisterDefaultActions` takes explicit dependencies.** The roller resolver,
  hook registry, world engine, and rulebook are passed as parameters so the
  `actions` package does not import `engine`. The parent engine owns registration
  without an import cycle.
- **The roller is resolved at invocation time, not captured.** Factories call
  `resolveRoller()` when the action is built (turn time), not when registered, so
  a test can swap `Engine.Rand` for a deterministic roller after construction and
  have it take effect. This is a fence-signed trap in `register.go` — capturing
  the roller by value would silently defeat the test seam.
- **Recoverable sentinels live in the action package.** `ErrNoSelection`,
  `ErrTurnCanceled`, and `ErrActionUnavailable` are defined here, not in the TUI,
  so the [logging & errors](logging-errors.md) classifier can recognize a
  GM-cancelled action as *recoverable* (skip the action, continue the cycle)
  without importing any interface code.
- **Abilities are folded into actions, not a sibling engine.** Ability resolution
  lives in `actions/ability/` and is reached only through the `Use Asset Ability`
  action, keeping the engine's top-level sub-engine count to the seven the
  orchestrator wires. (Frozen log 96–109; the ability-engine redesign.)

## Dependencies

**Depends on** `domain` (the mutation types every action emits, `Asset` /
`Faction`), [persistence & static data](persistence.md) for `FactionState` and
the rulebook (asset definitions, costs, flags), [hooks](hooks.md) (the embedded
`hooks.Collector` and the registry passed to combat actions), and
[world & movement](world-movement.md) for the spatial `Index` the world-aware
actions target through.

**Depended on by** the [orchestrator](orchestrator.md), which calls
`AvailableActions` and `Run` in the action phase and folds the result with
goal-progress mutations. The concrete `Collector` is implemented on the
interface side by the [event stream](../interface/event-stream.md) adapter.
