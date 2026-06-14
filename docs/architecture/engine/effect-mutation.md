# Tags, Effects & Mutation

> **Code:** `internal/faction/engine/{tag,effect,mutation}/`, `internal/faction/domain/mutation.go`

## Purpose

This page covers the two ends of the engine's data-driven rule machinery: how
*rulebook data becomes runtime behavior* — the tag and effect engines, which
turn faction tags and asset special features into registered [hooks](hooks.md) at
the start of each cycle — and how *every state change becomes state* — the
mutation vocabulary and the single `Apply` switch that writes it. Mutations are
the only currency of state change in the engine; every other subsystem emits
them, and this layer is where they land.

## Shape

The machinery has two layers that bracket a turn. At cycle start the **reactive
rule layer** (tag + effect) reads data and registers hooks. Throughout the turn
sub-engines emit mutations; at each phase gate the **apply layer** (mutation)
writes them to state.

### The reactive rule layer

Most of what makes one faction play differently from another is *reactive*: a
bonus die in the right kind of fight, a Coin gained when an asset dies, a cheaper
purchase. SWN expresses these as tag effects and asset special features, and the
engine routes all of them through the [hooks](hooks.md) framework — five
categories of hook the orchestrator consults at fixed points in the pipeline. The
tag and effect engines are the two **registrars** that translate game data into
those hooks. They never dispatch anything; like `RegisterDefaultActions`, they
populate a registry the orchestrator drives.

Both follow the same shape — a `map[string]Handler` keyed by a data ID and an
`ApplyAll` that, once per cycle, walks live entities and lets the matching handler
register its hooks. A tag or asset with no registered handler is **silently
skipped**: the fence-signed data-only contract shared with [goals](goals.md), and
the mechanism that lets the *data* lead the *implementation*. Both engines are
deliberately framework-complete but handler-partial — the event-hooks effort
shipped all five categories and a proof-of-life handler each, leaving the full
sets to follow (see Key Decisions).

#### Tags → faction-scoped hooks

`TagEngine` registers a fixed set of Go handlers in `New` and applies them by
walking each `faction.Tags` entry — so *which factions hold which tags* is
faction-state data, the behavior is code. Four of the rulebook's twenty tags are
implemented today, and between them they exercise all five hook
categories — concrete proof the registration path is general, not bolted on for
one shape:

| Tag | Hook category | Effect |
|-----|---------------|--------|
| Warlike | `RollModifier` (Cat 1) | +1d10 on a Force attack, keep highest (once/turn) |
| Fanatical | `RollResultHook` (Cat 2) + `TieResolver` (Cat 5) | rerolls any die showing 1; always loses ties |
| Scavengers | `MutationReactor` (Cat 3) | +1 Coin whenever an asset is destroyed in combat |
| Preceptor Archive | `AssetCostModifier` (Cat 4) | −1 Coin on TL4+ asset purchases |

The other sixteen tags are valid rulebook data — many with real mechanics
(Plutocratic's Wealth-attack die, Pirates' movement surcharge, Deep Rooted's
homeworld defense) — that stay inert until a handler is written. A tag may
register hooks in any combination of categories; Fanatical already does two.

#### Asset features → asset-scoped hooks

`EffectsEngine` is the home for asset **`S`-flag** features: the passive and
reactive specialness an asset carries just by existing. (The `A` flag — *active*
abilities a faction spends its action on — is resolved by the
[actions](actions.md) ability dispatch instead; `P` is a buy-time government-
permission check. The effect engine owns only the passive `S` side.) Its handlers
are registered the same *data-driven* way by the orchestrator at construction:
`NewWithRulebook` registers one handler per asset definition that declares the
relevant profile.

Only the **transport** family is wired today — `HeavyDropAssets`,
`BeachheadLanders`, `ExtendedTheater`, `DeepStrikeLanders`, every asset whose
definition carries a `Transport` profile. `TransportHandler.Apply` registers a
faction-scoped `TransportReactor` for each such asset; the reactor (a
`MutationReactor`) watches the movement-order mutations for its asset and, when
that transport moves with cargo, charges the per-issuance Coin cost and emits an
`AssetMoved` for each cargo asset so cargo location stays in lockstep with the
transport. That last point is deliberate — it keeps cargo recoverable if the
transport is destroyed mid-flight.

Every other `S`-flag feature is defined in asset data and still waiting for a
handler: Zealots take self-damage on a successful attack (a `MutationReactor`),
Blockade Fleet steals Coin on a hit (a once-per-turn `MutationReactor`), Integral
Protocols add a defense die against Cunning (a `RollModifier`), Psychic Assassins
enter play already stealthy (a `RuleModifier`), Planetary Defenses may only defend
against Starship-type assets (a `TieResolver`/rule alteration). The shapes are all
known and have homes; the engine just hasn't grown the handlers yet.

### The apply layer: the mutation vocabulary

A mutation is any value satisfying `domain.Mutation`:

```go
type Mutation interface {
	Type() string // stable discriminator, used for history serialization
}
```

Each concrete mutation is a small struct carrying a `FactionID`, the fields it
changes, a `Cause` string, and (for most) a `CausedByFactionID` — the provenance
pair that lets history and the [narrative digest](narrative-digest.md) attribute
every change to the action and faction that caused it. The vocabulary is the
finite list of everything that can change in the game.

#### `MutationEngine.Apply`

`Apply` is the one place state is written. It is a single type switch over every
mutation type — `delete(faction.Assets, id)` for `AssetRemoved`,
`faction.Coin += Delta` for `CoinDelta`, and so on — applied in order against the
`FactionState`. Two structural notes:

- **Assets are map-addressed; bases are slice-scanned.** `faction.Assets` is a
  `map[string]*Asset`, so asset mutations index directly by ID; bases remain a
  slice walked linearly. (The asset-map conversion was a deliberate refactor for
  O(1) entity addressing.)
- **Exhaustive by panic.** The switch ends in `default: panic("unhandled
  mutation type")`. A new `domain.Mutation` with no apply arm fails loudly the
  first time it is emitted, which is what keeps the vocabulary and the apply layer
  in lockstep.

When a mutation references an entity that no longer exists, `Apply` does not
abort: it records a `MutationMiss`, keeps applying the rest, and at the end
returns a single `MutationApplyError` listing every miss. The orchestrator treats
that as fatal — one structured Error log line per miss, then the turn aborts.
Fail-loud, with a complete diagnostic rather than a first-failure stop.

## Mutation catalogue

Every `domain.Mutation`, grouped by what it touches. The `type()` discriminator
is the string written to history; the rules behind the numbers live in
[`swn-faction-mechanics.md`](../../rules/swn-faction-mechanics.md).

**Economy & faction**
- `coin_delta` — adjust a faction's Coin
- `influence_delta` — adjust a Base's Influence
- `faction_hp_delta` — adjust faction HP
- `xp_awarded` / `xp_spent` — credit / debit XP
- `stat_raised` — raise one attribute and recompute MaxHP
- `homeworld_changed` — relocate the faction's homeworld
- `tag_added` — grant a tag (the full `Tag` is embedded, so `Apply` needs no rulebook lookup)

**Assets**
- `asset_added` / `asset_removed` — add to / drop from the roster
- `asset_hp_delta` — adjust an asset's HP
- `asset_maintained_flag` — set the maintained flag
- `asset_stealth_applied` / `asset_stealth_cleared` — toggle stealth
- `asset_moved` — set an asset's Location (cargo follows a transport)

**Bases**
- `base_added` / `base_destroyed` — place / remove a Base of Influence
- `base_hp_delta` — damage a base (callers pair it with a `faction_hp_delta`)
- `base_healed` — restore HP, capped at effective max
- `base_expanded` — raise both MaxHP and CurrentHP

**Goals**
- `goal_initiated` / `goal_abandoned` / `goal_completed` — set / clear the active goal
- `goal_progressed` — increment (or reset) progress
- `goal_turns_tick` — decrement `TurnsRemaining`
- `goal_phase_advanced` — set ProcessPhase and TurnsRemaining atomically

**Movement orders**
- `movement_order_issued` / `_progressed` / `_revised` / `_cancelled` / `_completed` — the in-flight order lifecycle (see [world & movement](world-movement.md))

## Key Decisions

- **Mutations are the only state-change currency.** No sub-engine writes
  `FactionState`; they all return `[]domain.Mutation`, and `Apply` is the single
  choke point that writes it. This is what makes every change recordable,
  serializable to history, and replayed deterministically.
- **The apply switch is exhaustive by panic.** A missing apply arm panics rather
  than silently dropping a mutation, keeping the vocabulary and the writer in
  lockstep as the game grows.
- **Apply-all, then fail loud.** `Apply` collects every missing-entity reference
  into a `MutationApplyError` instead of aborting on the first, so a failed turn
  reports all its problems at once. The orchestrator treats the error as fatal.
  (Frozen log 239–240.)
- **The rule layer registers hooks, it never dispatches.** Tag and effect engines
  translate data into [hooks](hooks.md) and stop there; the orchestrator owns
  every dispatch site. This is the same consumer-owned-registration pattern as the
  action factories, and it is what lets the hook framework grow without the
  orchestrator knowing which tags or effects exist.
- **Framework first, handlers incremental.** The event-hooks effort shipped all
  five hook categories plus registry and dispatch, but only a proof-of-life
  handler per category — four tags, one asset feature. The data-only skip makes
  the partial state safe: unimplemented tags and `S`-flag features are inert
  rulebook data, not errors, so handlers can land one at a time without touching
  the orchestrator. (Frozen log 158–188, 227–235.)
- **Passive and active asset specialness are split.** `S`-flag features register
  reactive hooks through the effect engine; `A`-flag abilities are resolved by the
  [actions](actions.md) ability dispatch. The effect engine deliberately owns only
  the passive side.
- **Cargo location stays in sync with its transport.** The `TransportReactor`
  emits an `AssetMoved` for each cargo asset on every movement beat, so if a
  transport is destroyed mid-flight its cargo's last-known location is already
  recorded and recoverable.

## Dependencies

**Depends on** `domain` (the mutation vocabulary, `Tag`, `Asset`, `Base`),
[hooks](hooks.md) (the registry the tag and effect engines register reactors
into), and [persistence & static data](persistence.md) for the `FactionState`
that `Apply` writes and the rulebook (tag and asset definitions) the engines read.

**Depended on by** essentially every other engine subsystem: the
[orchestrator](orchestrator.md) primes `ApplyAll` at cycle start and routes all
mutations through `Apply` at phase gates, and the [turn pipeline](turn-pipeline.md),
[actions](actions.md), [goals](goals.md), and [world & movement](world-movement.md)
engines all emit mutations from this vocabulary. The
[narrative digest](narrative-digest.md) reads the serialized mutation records back
out of history.
