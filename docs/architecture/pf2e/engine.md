# Engine

> **Code:** `internal/pf2e/sim/engine/`, `internal/pf2e/sim/dice/`

## Purpose

The rules and the fight loop. `engine.Run` takes the two sides, a roller, an
observer and a decider, plays one fight to its end, and reports every step as
an event. The engine is headless: it asks a `Decider` what each creature does
and tells an `Observer` what happened, and cannot tell who is on either side of
those seams. Effort 1 built the core and Effort 3 the effect core: checks with
typed bonuses, saves, ability steps, conditions, damage kinds and the Down row.
Effort 4 adds triggers and limits, and Effort 5 spells and tables.

## Shape

### The roller

`dice.Roller` is one method, `Die(sides int) int`. The real one is
`dice.Seeded`, made with `dice.NewSeeded(seed)` from a seeded PCG source, so one
seed always gives the same rolls. A fight takes one roller, and every roll in it
goes through that roller: initiative, attack rolls, saves, damage, flat checks
and recovery checks.

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
side and its fight state: HP, temporary HP, dying, wounded, hero points,
initiative, actions left, the attacks made this turn, and its conditions,
persistent damage and lowered defenses.

### Initiative

Each combatant rolls d20 plus Perception plus its `perception` and
`initiative` modifiers (see [Checks and modifiers](#checks-and-modifiers)), in
setup order: the party first, then the enemies. The order runs from the
highest total down. On a tie an NPC goes before a PC; within one kind the setup
order holds.

### The turn

A round walks the initiative order. Each turn opens with the start-of-turn
boundary (durations that end there) and closes with the end-of-turn boundary
(persistent damage, then frightened's decay, then durations that end there).
Between them, a combatant:

- **is dead** — does nothing, and no event fires
- **is dying** — makes a recovery check and nothing else
- **is at 0 HP and stable** — does nothing
- **is standing** — gets three actions. Its attack count and its reaction slot
  reset. It asks the decider for a choice and performs it, again and again,
  until it has no actions left, the decider has nothing to do, or the fight is
  over.

A Strike costs one action. An ability costs its action cost and runs its steps
(see [Abilities and steps](#abilities-and-steps)). A choice the creature cannot
make (an unknown or ignored ability, too few actions left, no standing target)
fails the run with `ErrBadChoice`.

The reaction slot is reset each turn, but nothing spends it until Effort 4's
reactions.

### Checks and degrees of success

`DegreeOf(die, total, dc)` gives the four degrees: a critical success at the DC
plus 10 or more, a success at the DC or more, a critical failure at the DC
minus 10 or less, a failure otherwise. A natural 20 then moves the degree up
one step, and a natural 1 moves it down one step.

### Checks and modifiers

Each check or DC builds one bonus pool. The number from the sheet or the build
(a Strike's bonus, a save, AC, the class DC) counts as one untyped base. Then
the creature's fight-time modifiers (`[modifier.*]`) that match the check's
selectors add their values, and then its conditions add their penalties. Per
type (`item`, `status`, `circumstance`) only the highest bonus and the lowest
penalty count; untyped values all add.

| Check | Selectors read |
|-------|----------------|
| Melee Strike attack | `attack`, `melee-attack` |
| Ranged Strike attack | `attack`, `ranged-attack` |
| Strike damage | `damage`, plus `melee-damage` in melee |
| Spell attack | `spell-attack` |
| AC | `ac` |
| A save | the save, `saving-throw` |
| Initiative | `perception`, `initiative` |
| Class DC, spell DC | `class-dc`, `spell-dc` |

A modifier with `against` traits applies only when the incoming effect carries
one of them (Luna's +2 against occult). A modifier with `per_condition` is
multiplied by the target's value of that condition, so it counts only on an
attack or damage roll against a target.

### Saves

A save step's DC is its `dc` plus the source's `class-dc` modifiers, or the
source's class DC (`dc_from = "class"`) or spell DC (`dc_from = "spell"`) plus
the matching DC modifiers. The saver rolls d20 plus the save plus its modifiers
on that save and `saving-throw`, against the effect's traits. The degree then
moves by the step's own `shift`, clamped to the four degrees, and then the
saver's first degree shift by ID that matches the save, the traits and the
current degree replaces it (Juggernaut: a Fortitude success becomes a critical
success). A `basic` save deals the step's damage, rolled once for every
target: none on a critical success, half on a success, full on a failure and
double on a critical failure.

### Abilities and steps

An ability is a list of steps. Each step is one of eight kinds:

| Kind | Does |
|------|------|
| `strike` | Makes one of the creature's Strikes |
| `save` | Forces a save; with `basic`, deals damage by degree |
| `spell_attack` | Rolls the spell attack against AC; its damage lands on a success, doubled on a critical success |
| `damage` | Deals damage with no roll |
| `heal` | Restores HP |
| `temp_hp` | Grants temporary HP |
| `condition` | Applies a condition |
| `weaken` | Raises weaknesses or lowers resistances to damage types |

A rolled step (`strike`, `save`, `spell_attack`) holds a list of steps for each
degree, and runs the list for the degree it rolled. On an attack roll, a
critical success with no `critical_success` list runs the `success` list; a
save runs only its own degree's list. A Strike carries the same four lists as
riders, so a weapon rune or a critical specialization is a step list on the
Strike.

- **An area** catches every standing foe, through one function,
  `areaTargets`, which Effort 7's position replaces.
- **A targeted ability** runs each step at its target. If the target falls
  partway through, the engine asks the decider for the next target; with none
  left, the ability stops. `target = "self"` points a step at the user.
- **Concealment** — a flat check comes before every attack roll and before a
  targeted ability whose first step is a save: DC 5 when the target is
  concealed or the attacker dazzled, DC 11 when the attacker is blinded. A
  failed check ends that attack, or the whole ability when it guards the
  opening save.
- **Immunity to an effect** — an ability step skips a target immune to one of
  the ability's traits, with `death-effects` matching the `death` trait.
  Only `-effects` keys and the trait-style immunities (such as `poison`,
  `disease`, `mental`, `sleep`) count; an immunity to a damage type like
  `fire` only zeroes that damage. Strike steps and Strike riders are not checked: a Strike's damage types
  already meet immunities.
- **`unless`** skips a target that has the named condition.
- **`lethal`** damage kills outright at 0 HP instead of knocking a PC down.

### Strikes and the multiple attack penalty

A Strike rolls d20 plus its bonus plus the multiple attack penalty against the
target's AC. The penalty counts the attacks made this turn, spell attacks
included: none for the first, −5 for the second and −10 for every one after.
An agile Strike takes −4 and −8 instead. A success deals damage; a critical
success deals double damage. A failure deals only the Strike's splash damage;
a critical failure deals nothing.

### Conditions

A Go table, `conditionRules`, says what each condition does. Each row names the
checks it lowers, the bonus type, a fixed penalty or one equal to the value,
whether it decays, and the conditions it brings with it.

| Condition | Lowers |
|-----------|--------|
| `blinded` | Nothing itself; brings off-guard, and attacks take a DC 11 flat check |
| `clumsy` | AC, Reflex, ranged attacks (status, by value) |
| `concealed` | Nothing; attacks against it take a DC 5 flat check |
| `dazzled` | Nothing; its attacks take a DC 5 flat check |
| `doomed` | Nothing; lowers the death value (see [Dying and death](#dying-and-death)) |
| `drained` | Fortitude (status, by value); maximum HP |
| `encumbered` | Brings clumsy 1 |
| `enfeebled` | Melee attacks and melee damage (status, by value) |
| `frightened` | Every check and DC (status, by value); decays by 1 at the end of the bearer's turn |
| `off-guard` | AC (circumstance, 2) |
| `prone` | Attacks and spell attacks (circumstance, 2); brings off-guard |
| `sickened` | Every check and DC (status, by value) |
| `stupefied` | Will, Perception, spell attacks and spell DCs (status, by value) |

A creature holds every instance it was given; a condition's value is the
highest instance, or the highest value brought by another condition. An
instance ends with its duration:

- **in rounds** — counted down at the start of the source's turn; "1 minute"
  is 10 rounds
- **`until`** a turn boundary — the start or end of the source's or the
  target's turn. An end boundary applied during that same creature's turn skips
  the current end, so "until the end of your next turn" means the next one.
- **none** — the instance lasts the fight, or until it decays to 0

`[condition_adjust.*]` changes how a condition lands on its bearer: `gain` adds
to each value it is given, and `decay` sets how much frightened drops each turn.
A creature immune to a condition by name never takes it.

Slowed, stunned, quickened, confused, immobilized, deafened and the Stand
action wait for Effort 4's action economy.

### Damage

A damage list is rolled once per step, then applied to each target. Each part is
rolled, a negative result is floored at 0, and on a critical hit the
uncategorized parts double. A Strike's damage modifiers add to its first
uncategorized part. Parts of one type are added together. If the whole roll
comes to less than 1, it deals 1; a basic save's halving comes after.

A part's `category` sets how it lands:

- **none** — ordinary damage
- **`persistent`** — not dealt now: the target gains that persistent damage
  (see [Persistent damage](#persistent-damage))
- **`splash`** — dealt with the rest on a hit, and alone on a failure. Splash
  lands on the primary target only, until Effort 7's position.

Then, for each type:

1. **Immunity** — if the target is immune to the type, it takes 0 of it.
2. **Weakness** — the strongest weakness that covers the type is added, unless
   the type rolled 0. A `weaken` step's raised weakness counts if it is higher.
3. **Resistance** — the strongest resistance that covers the type, lowered by
   the largest `weaken` on that type, is subtracted, and the result is floored
   at 0. An adaptive resistance (`defense.adaptive`) counts for the last of its
   types the creature took damage from.

An immunity, resistance or weakness covers a type when its key is that type,
`physical` (for bludgeoning, piercing and slashing) or `all`. Its `except` list
names damage types or effect traits that it does not cover. A weakness or
resistance keyed on a trait (`holy`, `cold-iron`, `silver`, `adamantine`,
`unholy`) applies once per hit when the effect carries the trait: the weakness
when any damage got through, the resistance up to the damage left.

Temporary HP absorb the total first, and the rest comes off HP, which stops at
0.

### Persistent damage

A creature keeps one persistent damage per type; a new one of the same type
replaces it only when its average is higher. A critical hit doubles it. At the
end of the bearer's turn each one deals its damage through the same steps as
any damage, then the bearer rolls a DC 15 flat check, and a success ends it. A
creature immune to the type never gains it.

### Healing and temporary HP

Healing restores HP up to the maximum, which drained lowers. A dying PC healed
above 0 HP loses its dying value and gains 1 wounded. Temporary HP do not stack:
a new amount replaces the old only when it is higher.

### Dying and death

When a standing creature drops to 0 HP:

- **an NPC dies**
- **a PC goes dying** — dying 1 plus its wounded value, plus 1 more on a
  critical hit. The PC also moves in the initiative order to directly before
  the creature that knocked it down, unless it dropped on its own turn.

Damage that gets through to a dying PC raises its dying value by 1, or by 2 on
a critical hit. A stable PC at 0 HP that takes damage goes dying again with the
knock-down value, but does not move in the initiative order.

The **death value** is the creature's `dying_max` (4 when unset) minus its
doomed value. When a PC's dying value would reach it, the PC dies, unless it
has a hero point: then Heroic Recovery spends every hero point and sets dying
to 0. Hero points come from the creature's `hero_points`.

A dying PC makes a recovery check at the start of its turn: a flat check (the
d20 alone) against its `recovery_dc` (10 when unset) plus its dying value, with
natural 20 and natural 1 shifting the degree like any check. A critical success
lowers dying by 2, a success by 1; a failure raises it by 1, a critical failure
by 2. A PC whose dying value reaches 0 is stable: it stays at 0 HP, and its
wounded value rises by 1.

**Drained** N takes level × N HP when it is gained, and lowers maximum HP by
the same. A drain that takes a standing creature to 0 HP knocks it down.

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
| `StrikeRolled` | An attack roll is made, with its die, bonus, penalty, total, AC and degree |
| `SpellAttackRolled` | A spell attack is made, with the same fields |
| `SaveRolled` | A creature saves, with its source, ability, die, bonus, total, DC and final degree |
| `FlatChecked` | A flat check is rolled, for concealment or to end persistent damage |
| `ImmuneToEffect` | An ability step skips a target immune to its traits |
| `DamageDealt` | Damage lands: each type rolled and taken, the trait adjustment, the total, the temporary HP absorbed and the HP left |
| `HPLost` | HP are lost without damage (drained) |
| `Healed` | A creature is healed, with the HP after |
| `Revived` | Healing lifts a dying PC: the dying value it had and its new wounded value |
| `TempHPGained` | A creature gains temporary HP |
| `ConditionChanged` | A condition's value changes on a creature, from and to |
| `PersistentChanged` | A creature gains or replaces a persistent damage |
| `PersistentEnded` | A persistent damage ends on a flat check |
| `DefenseChanged` | A `weaken` step raises weaknesses or lowers resistances |
| `DefenseRestored` | A `weaken` effect's duration ends |
| `ResistanceAdapted` | An adaptive resistance switches to a new damage type |
| `InitiativeMoved` | A knocked-down PC moves to just before the creature that knocked it down |
| `DyingChanged` | A PC's dying value changes |
| `RecoveryChecked` | A dying PC makes a recovery check |
| `Stabilized` | A PC's dying value reaches 0 on a recovery check, with its new wounded value |
| `HeroicRecovery` | A PC spends its hero points to avoid death |
| `Died` | An NPC drops to 0 HP, lethal damage drops a creature, or a PC reaches its death value |
| `FightEnded` | One side has no standing creature, or the round limit passes (a stalemate) |

### Not modelled yet

- **Effort 4 (planned)** — reactions and triggers, fortune and misfortune,
  recharge, frequency and uses, the action economy and the conditions that
  change a turn (slowed, stunned, quickened, confused, immobilized, deafened,
  the Stand action), Sustain and fast healing.
- **Effort 5 (planned)** — spells: slots, ranks and heightening, counteract,
  incapacitation, temporary immunity and random tables.
- **Effort 7 (planned)** — position. Every creature counts as engaged with
  every foe: an area catches every standing foe, splash lands on the primary
  target only, and reach, range and fall damage are not checked.
- Weapon traits other than agile (deadly, fatal, versatile and the rest) are
  not read. The import copies every trait onto the Strike, and `show` and
  the import report print them.
- Speed is loaded, but no rule reads it.

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
  Effort 7's rule lists replace it behind the same seam. (`F.041.E1`
  Decision 6.)
- **Abilities are recursive steps with a list of steps per degree of success.** A save, a spell attack or a Strike runs the list for its degree, so a rider, a condition and a nested save share one shape, and the converter fills it the way PF2e text is written. (`F.041.E3` Decision 6.)
- **Condition rules are a Go table of shared parts.** Each row names what the condition lowers, the bonus type and whether it decays; code exists only where a part cannot say the rule. The ruleset as data will lift the table into files. (`F.041.E3` Decision 2.)
- **An area catches every standing foe until the position seam arrives.** One function, `areaTargets`, answers it, and Effort 7 replaces it. (`F.041.E3` Decision 1.)
- **The sheet's number is one untyped base; fight-time bonuses stack among themselves.** A fight-time bonus of a type the build already applies is written as the difference. (`F.041.E3` Decision 10.)

## Dependencies

**Depends on** [`creature`](data-files.md) for the loaded stat blocks, and on
`dice` (on this page) for the roller and dice expressions.

**Used by** the gate 3 test, `TestRunIsDeterministic`, which plays Luna against
three dummy PCs on ten seeds and checks that each seed gives the same events
twice. No command runs a fight yet; the runner (planned, Effort 8) will, and
its scoring will read the events.
