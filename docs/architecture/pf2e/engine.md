# Engine

> **Code:** `internal/pf2e/sim/engine/`, `internal/pf2e/sim/dice/`

## Purpose

The rules and the fight loop. `engine.Run` takes the two sides, a roller, an
observer and a decider, plays one fight to its end, and reports every step as
an event. The engine is headless: it asks a `Decider` what each creature does
and tells an `Observer` what happened, and cannot tell who is on either side of
those seams. Effort 1 builds the core rules; Effort 3 adds the rest.

## Shape

### The roller

`dice.Roller` is one method, `Die(sides int) int`. The real one is
`dice.Seeded`, made with `dice.NewSeeded(seed)` from a seeded PCG source, so one
seed always gives the same rolls. A fight takes one roller, and every roll in it
goes through that roller: initiative, attack rolls, damage and recovery checks.

`dice.Parse` reads a dice expression such as `3d6+14`, `1d8` or a plain number
into a `dice.Expression` (count, sides, modifier). `Expression.Roll` rolls it
through a roller, and `Expression.String` writes it back.

### Setup and Run

`engine.Setup` holds the party, the enemies (both as loaded
`creature.Creature` values), the roller, the observer, the decider and a round
limit. `engine.Run(setup)` returns the `Outcome` and an error; beside an error
the outcome is `NoOutcome`, the zero value, so a failed fight never counts as a
win. It refuses a
setup with an empty side, a round limit below 1, or no roller, observer or
decider (`ErrBadSetup`), and two creatures with the same ID
(`ErrDuplicateCombatant`). Each creature becomes a `Combatant` that holds its
side and its fight state: HP, dying, wounded, initiative, actions left and the
attacks made this turn.

### Initiative

Each combatant rolls d20 plus Perception, in setup order: the party first, then
the enemies. The order runs from the highest total down. On a tie an NPC goes
before a PC; within one kind the setup order holds.

### The turn

A round walks the initiative order. On its turn a combatant:

- **is dead** — does nothing, and no event fires
- **is dying** — makes a recovery check and nothing else
- **is at 0 HP and stable** — does nothing
- **is standing** — gets three actions. Its attack count and its reaction slot
  reset. It asks the decider for a choice and performs it, again and again,
  until it has no actions left, the decider has nothing to do, or the fight is
  over.

A Strike costs one action. An ability costs its action cost and runs its steps
in order, each step a Strike. If the target falls partway through, the engine
asks the decider for the next target; with none left, the ability stops. A
choice the creature cannot make (an unknown or ignored ability, too few actions
left, no standing target) fails the run with `ErrBadChoice`.

The reaction slot is reset each turn, but nothing spends it until Effort 3's
reactions.

### Checks and degrees of success

`DegreeOf(die, total, dc)` gives the four degrees: a critical success at the DC
plus 10 or more, a success at the DC or more, a critical failure at the DC
minus 10 or less, a failure otherwise. A natural 20 then moves the degree up
one step, and a natural 1 moves it down one step.

### Strikes and the multiple attack penalty

A Strike rolls d20 plus its bonus plus the multiple attack penalty against the
target's AC. The penalty counts the attacks made this turn: none for the first,
−5 for the second and −10 for every one after. An agile Strike takes −4 and −8
instead. A success deals damage; a critical success deals double damage.

### Damage

Each damage part is rolled, a negative result is floored at 0, and a critical
hit doubles it. Parts of one type are added together. Then, for each type:

1. **Immunity** — if the target is immune to the type, it takes 0 of it.
2. **Weakness** — the strongest weakness that covers the type is added.
3. **Resistance** — the strongest resistance that covers the type is
   subtracted, and the result is floored at 0.

An immunity, resistance or weakness covers a type when its key is that type,
`physical` (for bludgeoning, piercing and slashing) or `all`. Its `except` list names
damage types or Strike traits that it does not cover. The target loses the sum,
and HP stops at 0.

### Dying and death

When a standing creature drops to 0 HP:

- **an NPC dies**
- **a PC goes dying** — dying 1 plus its wounded value, plus 1 more on a
  critical hit. The PC also moves in the initiative order to directly before
  the creature that knocked it down.

Damage that gets through to a dying PC raises its dying value by 1, or by 2 on
a critical hit. A stable PC at 0 HP that takes damage goes dying again with the
knock-down value, but does not move in the initiative order. At dying 4 the PC
dies.

A dying PC makes a recovery check at the start of its turn: a flat check (the
d20 alone) against DC 10 plus its dying value, with natural 20 and natural 1
shifting the degree like any check. A critical success lowers dying by 2, a
success by 1; a failure raises it by 1, a critical failure by 2. A PC whose
dying value reaches 0 is stable: it stays at 0 HP, and its wounded value rises
by 1.

### The end of a fight

After each action and each turn the engine checks the sides. The party wins when no enemy is
standing; the enemies win when no party member is standing (a dying PC is not
standing). If the round limit passes with both sides standing, the fight is a
stalemate.

### The decider

```go
type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
}
```

A `Choice` is a Strike or an ability, by ID, and one target. `FirstFitting` is
the only decider today, a fixed behaviour for tests. It takes the first
ability by ID that is not ignored and whose action cost fits the actions left;
if none fits, it takes the creature's first Strike by ID. Its target is the
first standing foe in setup order, and `NextTarget` gives the same answer.

### Events

The `Observer` has one method, `Observe(Event)`. Each event kind is its own
struct, and events name combatants by ID.

| Event | Fires when |
|-------|------------|
| `InitiativeRolled` | A combatant rolls initiative, before the first round |
| `RoundStarted` | A round begins |
| `TurnStarted` | A combatant that is not dead starts its turn |
| `ActionUsed` | A combatant spends actions on a Strike or an ability |
| `StrikeRolled` | An attack roll is made, with its die, penalty, total, AC and degree |
| `DamageDealt` | A hit deals damage: each type rolled and taken, the total and the HP left |
| `InitiativeMoved` | A knocked-down PC moves to just before the creature that knocked it down |
| `DyingChanged` | A PC's dying value changes |
| `RecoveryChecked` | A dying PC makes a recovery check |
| `Stabilized` | A PC's dying value reaches 0 on a recovery check, with its new wounded value |
| `Died` | An NPC drops to 0 HP, or a PC reaches dying 4 |
| `FightEnded` | One side has no standing creature, or the round limit passes (a stalemate) |

### Not modelled yet

- **Effort 3 (planned)** — bonus-type stacking, basic saves, doomed, hero
  points, Inextinguishable, Too Angry to Die, conditions, reactions and
  triggers, spells, persistent damage, and the minimum of 1 damage when a
  penalty would take a damage roll below 1.
- **Effort 5 (planned)** — position. Every creature counts as engaged with
  every foe; reach traits are loaded but not checked.
- Weapon traits other than agile (deadly, fatal, versatile and the rest) are
  not read. Effort 2's import report names the ones the PCs carry.
- Level, saves, speed, and Perception beyond initiative are loaded, but no rule
  reads them.

## Key Decisions

- **The engine is headless: decisions come in through a `Decider`, events go
  out through one `Observe(Event)`.** Events are plain values that name
  combatants by ID, so scoring, replay and the determinism test can keep and
  compare them. Unlike the faction engine's method-per-event observer.
  (`F.041.E1` Decision 5.)
- **One seeded roller per fight; the same seed and files give the same
  events.** Every roll goes through it, and no map's order reaches a roll or an
  event: Strikes and abilities are walked in sorted ID order, and a map is read
  unsorted only where order cannot change the result (picking the strongest
  resistance or weakness). (`F.041` gate 3.)
- **Rules are modelled as written, and nothing is faked.** NPCs die at 0 HP and
  PCs go dying; an initiative tie goes to the NPC; a PC knocked out moves to
  just before the creature that did it. (`F.041` D-18; `F.041.E1` Decisions 2
  and 9.)
- **`FirstFitting` is a test behaviour, not tactics.** It takes the first
  fitting ability by ID, else the first Strike, at the first standing foe.
  Effort 5's rule lists replace it behind the same seam. (`F.041.E1`
  Decision 6.)

## Dependencies

**Depends on** [`creature`](data-files.md) for the loaded stat blocks, and on
`dice` (on this page) for the roller and dice expressions.

**Used by** the gate 3 test, `TestRunIsDeterministic`, which plays Luna against
three dummy PCs on ten seeds and checks that each seed gives the same events
twice. No command runs a fight yet; the runner (planned, Effort 6) will, and
its scoring will read the events.
