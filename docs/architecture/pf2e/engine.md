# Engine

> **Code:** `internal/pf2e/sim/engine/`, `internal/pf2e/sim/dice/`

## Purpose

The rules and the fight loop. `engine.Run` takes the two sides, a roller, an
observer and a decider, plays one fight to its end, and reports every step as
an event. The engine is headless: it asks a `Decider` what each creature does
and tells an `Observer` what happened, and cannot tell who is on either side of
those seams. Effort 1 built the core and Effort 3 the effect core: checks with
typed bonuses, saves, ability steps, conditions, damage kinds and the Down row.
Effort 4 built triggers and limits: the action count, limits and pools,
reactions and other triggered abilities, fortune and misfortune, abilities
aimed at allies, acts of loyalty, skill checks and Sustain. Effort 5 adds
spells and tables.

## Shape

### The roller

`dice.Roller` is one method, `Die(sides int) int`. The real one is
`dice.Seeded`, made with `dice.NewSeeded(seed)` from a seeded PCG source, so one
seed always gives the same rolls. A fight takes one roller, and every roll in it
goes through that roller: initiative, attack rolls, saves, skill checks,
damage, flat checks, recovery checks and recharge rolls.

`dice.Parse` reads a dice expression such as `3d6+14`, `1d8` or a plain number
into a `dice.Expression` (count, sides, modifier). `Expression.Roll` rolls it
through a roller, and `Expression.String` writes it back.

### Setup and Run

`engine.Setup` holds the party, the enemies (both as loaded
`creature.Creature` values), the roller, the observer, the decider and a round
limit. `engine.Run(setup)` returns the `Outcome` and an error; beside an error
the outcome is `NoOutcome`, the zero value, so a failed fight never counts as a
win. It refuses a setup with an empty side, a round limit below 1, or no
roller, observer or decider (`ErrBadSetup`), and two creatures with the same ID
(`ErrDuplicateCombatant`). Each creature becomes a `Combatant` that holds its
side and its fight state: HP, temporary HP, dying, wounded, hero points,
initiative, actions left and Strike-only actions, the attacks made this turn,
its reaction and extra reactions, and its conditions, persistent damage and
lowered defenses. It also holds the uses spent, the pools (copied from the
creature's `[pool]`), the recharges in progress, the reroll effects, granted
bonuses and traits, eased recovery DCs and the sustained effects.

A fight can also stop on a fault: a bad answer from the decider's `React`
(`ErrBadChoice`) or a trigger chain nested too deep (`ErrTriggerLoop`). The
first fault is kept, and `Run` returns it with `NoOutcome` after the current
turn.

### Initiative

Each combatant rolls d20 plus Perception plus its `perception` and
`initiative` modifiers (see [Checks and modifiers](#checks-and-modifiers)), in
setup order: the party first, then the enemies. The order runs from the
highest total down. On a tie an NPC goes before a PC; within one kind the setup
order holds. Then, before round 1, each combatant in initiative order fires its
`initiative` triggers (Battle Cry).

### The turn

A round walks the initiative order. Each turn opens with the start-of-turn
boundary (durations that end there) and closes with the end-of-turn boundary
(persistent damage, then frightened's decay, then sustained effects that were
not sustained, then durations that end there). Between them, a combatant that
is not dead:

1. Logs `TurnStarted` and gets its reaction back.
2. Ticks its limits: per-round uses and flourish come back, and each recharge
   counts down by 1 (see [Limits](#limits)).
3. If it is at 0 HP, its fast healing comes first (see below). Then a dying
   creature makes a recovery check and, unless the check kills it, regains its
   actions without acting; a stable creature at 0 HP regains its actions
   without acting.
4. A standing creature resets its attack count, regains its actions, fires its
   `turn-start` triggers (Last Stand), then gets its fast healing, then asks
   the decider for a choice and performs it, again and again, until it has no
   actions left, the decider has nothing to do, or the fight is over.

**Regaining actions.** A creature gets 3 actions, or 4 when quickened. Stunned
N takes N of them and lowers stunned by that many; slowed N takes N. The two
overlap: the actions lost are the higher of the two, never their sum. A stunned
with a duration and no value takes every action, quickened included, and ends
only by its clock. A creature that loses actions loses the quickened one first;
otherwise the quickened action is a Strike-only action, which only a Strike
can spend. `ActionsRegained` logs the count when quickened or when actions were
lost. A downed creature regains actions too, so stunned and slowed count down
while it is down. Its extra reactions come back here, so they arrive at its
first turn.

**Stunned on its own turn.** A creature that gains stunned during its own turn
loses its remaining actions at once, up to the value, the Strike-only one first,
and stunned drops by that many (`ActionsLost`). A stunned with a duration takes
every action left. The ability in progress completes.

**Fast healing.** A creature with `fast_healing` heals that amount at the start
of its turn when it is not dead and below its maximum HP. A standing creature
heals after its `turn-start` triggers. A creature at 0 HP heals before its
recovery check; healed above 0, it is revived (dying cleared, wounded raised)
and takes the standing branch in the same turn, still prone. It heals once a
turn either way.

**Choices.** The decider answers with one of four kinds:

- **A Strike** costs 1 action, and spends the Strike-only action first.
- **An ability** costs its action cost. A cost range (`actions` to
  `max_actions`) takes the choice's `Actions`, or the lowest cost when unset. A
  `free` ability costs nothing. The engine checks, in order, that the ability
  exists and is not ignored or triggered, that the target fits, that the cost
  is in range, that the open actions cover it, and that no limit blocks it. Then
  it spends the actions and the limits, lets foes answer it (see
  [Provoking and disrupting](#provoking-and-disrupting)), and runs the steps.
  `turn_ends` sets the actions left, the Strike-only one included, to 0
  afterwards.
- **Stand** costs 1 action and needs prone. It provokes like a move action, and
  a Stand that is not disrupted removes prone and logs `StoodUp`.
- **Sustain** costs 1 action (see [Sustain](#sustain)).

A choice the creature cannot make fails the run with `ErrBadChoice`.

### Checks and degrees of success

`DegreeOf(die, total, dc)` gives the four degrees: a critical success at the DC
plus 10 or more, a success at the DC or more, a critical failure at the DC
minus 10 or less, a failure otherwise. A natural 20 then moves the degree up
one step, and a natural 1 moves it down one step.

### Checks and modifiers

Each check or DC builds one bonus pool. The number from the sheet or the build
(a Strike's bonus, a save, AC, a skill, the class DC) counts as one untyped
base. Then the creature's fight-time modifiers (`[modifier.*]`) that match the
check's selectors add their values, then bonuses granted by `bonus` steps, and
then its conditions add their penalties. Per type (`item`, `status`,
`circumstance`) only the highest bonus and the lowest penalty count; untyped
values all add.

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
| Skill check | the skill, `skill-check`, and `<attribute>-skill` for condition penalties |

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
double on a critical failure. `success_damage = "full"` keeps full damage on a
success (Wail).

After each save against a foe's effect, the saver fires its `saved` triggers
with the degree, and with `damaging` set when the step or any of its degree
lists deals damage. A save of success or better against an ability tagged
`tests-loyalty` is an act of loyalty by the saver (see
[Acts of loyalty](#acts-of-loyalty)).

### Skill checks

A `check` step rolls d20 plus the creature's skill number (`creature.Skills`,
computed from the build for a PC, typed for an NPC) plus its modifiers on the
skill and `skill-check`, with the conditions that lower that skill's attribute.
The DC is the step's `dc`, or 10 plus the target's save or Perception and its
modifiers on it against the ability's traits (`against = "will"`). The step
runs the list for its degree; a critical degree with no list of its own runs
nothing. `SkillChecked` logs the roll. Demoralize, Battle Cry and Too Angry to
Die are checks against Will.

### Abilities and steps

An ability is a list of steps. Each step is one of fifteen kinds:

| Kind | Does |
|------|------|
| `strike` | Makes one of the creature's Strikes, with optional tweaks (below) |
| `save` | Forces a save; with `basic`, deals damage by degree |
| `spell_attack` | Rolls the spell attack against AC; its damage lands on a success, doubled on a critical success |
| `check` | Rolls a skill check against a DC or a target's defense |
| `damage` | Deals damage with no roll |
| `heal` | Restores HP |
| `temp_hp` | Grants temporary HP |
| `condition` | Applies a condition, or with `reduce` lowers one |
| `weaken` | Raises weaknesses or lowers resistances to damage types |
| `gain_actions` | Adds actions to the user's turn, with no limit on their use |
| `recharge` | Ends a recharge on one of the user's abilities now |
| `reroll` | Gives the target a fortune or misfortune effect |
| `bonus` | Gives the target a typed bonus for a number of uses or a duration |
| `recovery_dc` | Caps the target's recovery check DC for a duration |
| `grant` | Gives the target a trait for its next attacks against the user |

A rolled step (`strike`, `save`, `spell_attack`, `check`) holds a list of
steps for each degree, and runs the list for the degree it rolled. On an attack
roll, a critical success with no `critical_success` list runs the `success`
list; a save or a check runs only its own degree's list. A Strike carries the
same four lists as riders, so a weapon rune or a critical specialization is a
step list on the Strike. A Strike step's own lists run after the Strike's
riders.

A Strike step's tweaks: `extra_dice` adds weapon dice to the first rolled part
with no category, `extra_damage` adds parts, `counts = 2` makes the Strike
count as two attacks, `penalty = "shared"` reuses the penalty of the previous
attack this turn (a Strike or a spell attack) without counting,
`target = "other"` aims at the first
standing foe that is not the ability's target, and `disrupts` names the action
traits a critical hit disrupts. A damage part with `per_action` is multiplied
by the actions spent on a cost-range ability.

- **An area** catches the side it is for, through one function,
  `areaTargets`: every standing foe, or with `targets = "ally"` every ally
  that is not dead (the user included), or with `targets = "others"` every
  ally but the user. `A.004.2`'s position replaces it.
- **A targeted ability** runs each step at its target. If a foe target falls
  partway through, the engine asks the decider for the next target; with none
  left, the ability stops. `target = "self"` points a step at the user.
- **A friendly ability** (`targets` set) aims at the user (`self`) or a chosen
  ally that is not dead, a dying one included (`ally`). It never moves to a new
  target and skips the opening concealment check.
- **An ability stops when its user falls.** Each step checks that the user
  still stands, because a reaction can drop it partway through.
- **Concealment** — a flat check comes before every attack roll and before a
  targeted ability whose first step is a save: DC 5 when the target is
  concealed or the attacker dazzled, DC 11 when the attacker is blinded. A
  failed check ends that attack, or the whole ability when it guards the
  opening save.
- **Immunity to an effect** — an ability step skips a target immune to one of
  the ability's traits, with `death-effects` matching the `death` trait. Only
  `-effects` keys and the trait-style immunities (such as `poison`, `disease`,
  `mental`, `sleep`, `visual`) count; an immunity to a damage type like `fire`
  only zeroes that damage. Strike steps and steps in a Strike's degree lists
  are not checked.
- **`unless`** skips a target that has the named condition.
- **`lethal`** damage kills outright at 0 HP instead of knocking a PC down.

### Strikes and the multiple attack penalty

A Strike rolls d20 plus its bonus plus the multiple attack penalty against the
target's AC. The penalty applies only on the attacker's own turn and counts the
attacks made this turn, spell attacks included: none for the first, −5 for the
second and −10 for every one after. An agile Strike takes −4 and −8 instead. A
reaction Strike on another creature's turn takes no penalty and does not count.
A `penalty = "shared"` Strike takes the penalty of the last attack this turn
and does not count; at the start of a turn that penalty is 0.
A success deals damage; a critical success deals double damage. A failure deals
only the Strike's splash damage; a critical failure deals nothing.

The order of one Strike: a ranged Strike first provokes (a disrupted one never
rolls), then the penalty is counted, then granted traits against the target are
added and spent, then the concealment check, then the target's `targeted`
triggers, then the roll and the damage, then the riders and the step's lists,
then the target's `attacked` triggers with the degree.

### Limits

An ability's `limit` table holds up to four limits, and `Combatant.Blocked`
names the first one that stops it:

| Limit | Blocks when | Spent when | Comes back |
|-------|-------------|------------|------------|
| `per_fight` | the uses are gone | each use (`UseSpent`) | never |
| `per_round` | used that many times | each use | at the start of the user's next turn |
| `recharge` | it is recharging | each use rolls the dice, at least 1 (`RechargeRolled`) | counted down at the start of the user's turn; at 0 it is ready (`Recharged`), so a roll of 1 means ready next turn |
| `pool` | the named pool is at 0 | each use takes 1 (`PoolSpent`) | never |

A pool is a named count on the creature, set in `[pool]`, that several
abilities share (drinking and feeding one stock of potions). Two traits limit
too: a `flourish` ability is blocked once any flourish was used, until the
start of the user's next turn, on the same clock as `per_round`; a `press`
ability is blocked until the user has attacked this turn. Limits are spent when
the actions are spent, before foes can disrupt the ability, so a disrupted use
is still spent. A triggered ability is checked and spent the same way.

The decider can ask `UsesLeft` (−1 with no `per_fight`), `RechargeLeft` and
`Pool` on a combatant, and `Blocked`.

### Triggers and reactions

An ability with a `trigger` table fires on an event instead of being chosen.
Each kind has its own qualifiers:

| Kind | Fires when | Qualifiers |
|------|------------|------------|
| `targeted` | A foe's attack roll is about to be made against the owner, after the concealment check | `attack` (`strike` or `melee-strike`) |
| `attacked` | A foe's attack roll against the owner has resolved | `attack`, `degrees` (the attacker's degree) |
| `saved` | The owner saved against a foe's effect | `degrees` (the owner's degree), `damaging` |
| `damaged` | Damage from a foe got through to the owner (persistent damage does not count) | `melee` |
| `turn-start` | The owner's turn starts, after it regains its actions | `hp_at_most` |
| `initiative` | Initiative is rolled, before round 1 | — |
| `foe-acts` | A foe uses an ability it chose (not a triggered one or a Sustain), Stands, or makes a ranged Strike | `traits` (any overlap), `ranged` |
| `ally-gains` | An ally's value of a listed condition rises | `conditions` |
| `stood-up` | The owner Stands after being knocked out | — |
| `loyalty` | A foe performs an act of loyalty | — |

`cost` is `reaction`, `free` or unset:

- **Unset** — a passive effect. It fires whenever it matches and its limits
  allow, with no decision.
- **`reaction` or `free`** — a decision. The engine gathers every matching,
  unblocked ability the owner can pay for and asks the decider's `React` once
  per event, with the event and the IDs offered; it answers with one or none. A
  reaction spends the ability's own extra reaction first, then the shared one.
  A stunned owner is offered nothing.

The owner's reaction comes back when its turn starts, and reactions are allowed
before its first turn (Robert's table rule). `[extra_reaction]` gives one
ability extra reactions each turn (Tactical Reflexes), from the owner's first
turn on. A triggered ability aims at the creature that set it off and never
moves to a new target; with no such creature (`turn-start`, `initiative`,
`stood-up`) it asks the decider for a target. `Triggered` logs who, which ability, the kind
and what set it off.

Each firing nests inside the event that caused it. A chain nested more than 16
deep (`maxTriggerDepth`) stops the fight with `ErrTriggerLoop`, naming the
owner and the kind.

### Provoking and disrupting

A chosen ability, a Stand (with the `move` trait) and a ranged Strike each
`provoke`: every standing foe fires its `foe-acts` triggers, since every
creature counts as engaged with every foe. A Strike step with `disrupts` that
critically hits the creature whose action is pending, and whose list overlaps
that action's traits, disrupts it (`ActionDisrupted`): the actions and the
limits stay spent and the action does nothing. Attack of Opportunity disrupts
`manipulate`.

### Fortune and misfortune

A `reroll` step gives its target an effect that keeps the `better` (fortune)
or `worse` (misfortune) of two d20s on the roll kinds it names (`attack`,
`save`, `skill`), for `uses` rolls or a duration. Every attack roll, save and
skill check goes through `d20`; flat checks, recovery checks and initiative do
not. Of two effects of one kind, the first gained applies. Fortune and
misfortune on one roll cancel: the roll is made once (`RerollsCancelled`), and
each still loses a use. Challenge Fate is a reaction that gives the attacker a
one-use misfortune on its attack roll.

### Allies, healing and revives

A `heal` step restores HP up to the maximum, which drained lowers. A PC healed
above 0 HP loses its dying value and gains 1 wounded (`Revived`). A creature
that revives an ally does an act of loyalty (`revive`). Temporary HP do not
stack: a new amount replaces the old only when it is higher. A `bonus` step
with `uses` counts only on `ac` and is spent by each attack roll against the
bearer (Farabellus Flip). A `recovery_dc` step caps the target's recovery
check DC for its duration (Bewitching Bloom).

### Acts of loyalty

Three things are acts of loyalty: using an ability tagged `loyalty` (after it
runs, and only if it was not disrupted), reviving an ally, and passing a save
against a foe's ability tagged `tests-loyalty`. Each creature does at most one
per initiative round, and only while a standing foe has a `loyalty` trigger;
`LoyaltyAct` names the reason. The foe's trigger (Sin Vulnerability) aims at
the acting creature. A `grant` step gives the target a trait (Faithful Strike's
`holy`) for its next `uses` Strikes against the step's user, spent by the
Strike that carries it, hit or miss (`TraitGranted`, `TraitSpent`).

### Sustain

An ability with a `sustain` list is recorded when it is used and its user still
stands. Each later turn the user can Sustain it for 1 action, once per turn,
which runs the `sustain` steps at the same target. The effect lasts to the end
of the user's next turn: at the end of a turn in which it was not sustained, it
ends (`SustainEnded`). A user who drops cannot Sustain, so its effect ends at
the end of its next turn.

### Conditions

A Go table, `conditionRules`, says what each condition does. Each row names the
checks it lowers, the bonus type, a fixed penalty or one equal to the value,
whether it decays, and the conditions it brings with it.

| Condition | Lowers |
|-----------|--------|
| `blinded` | Nothing itself; brings off-guard, and attacks take a DC 11 flat check |
| `clumsy` | AC, Reflex, ranged attacks, Dexterity skill checks (status, by value) |
| `concealed` | Nothing; attacks against it take a DC 5 flat check |
| `dazzled` | Nothing; its attacks take a DC 5 flat check |
| `deafened` | Nothing; a name only |
| `doomed` | Nothing; lowers the death value (see [Dying and death](#dying-and-death)) |
| `drained` | Fortitude (status, by value); maximum HP |
| `encumbered` | Brings clumsy 1 |
| `enfeebled` | Melee attacks, melee damage, Strength skill checks (status, by value) |
| `frightened` | Every check and DC, skill checks included (status, by value); decays by 1 at the end of the bearer's turn |
| `off-guard` | AC (circumstance, 2) |
| `prone` | Attacks and spell attacks (circumstance, 2); brings off-guard; Stand removes it |
| `quickened` | Nothing; one more action each turn, for a Strike only |
| `sickened` | Every check and DC, skill checks included (status, by value) |
| `slowed` | Nothing; N fewer actions each turn |
| `stunned` | Nothing; N fewer actions, or all of them for a duration (see [The turn](#the-turn)) |
| `stupefied` | Will, Perception, spell attacks, spell DCs, Intelligence, Wisdom and Charisma skill checks (status, by value) |

A creature holds every instance it was given; a condition's value is the
highest instance, or the highest value brought by another condition. A
`condition` step with `reduce` lowers every instance with a value by that much;
a stunned with a duration has none and is left alone. When a creature's value
of a condition rises, its allies fire their `ally-gains` triggers. An instance
ends with its duration:

- **in rounds** — counted down at the start of the source's turn; "1 minute"
  is 10 rounds
- **`until`** a turn boundary — the start or end of the source's or the
  target's turn. An end boundary applied during that same creature's turn skips
  the current end, so "until the end of your next turn" means the next one.
- **none** — the instance lasts the fight, or until it decays to 0

`[condition_adjust.*]` changes how a condition lands on its bearer: `gain` adds
to each value it is given, and `decay` sets how much frightened drops each turn.
A creature immune to a condition by name never takes it.

Confused arrives with Effort 5, and immobilized with position (`A.004.2`).

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
  lands on the primary target only, until `A.004.2`'s position.

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
0. `DamageDealt` and `Healed` carry `TriggeredBy`, the creature whose action
set off the triggered ability that dealt or healed it, so damage from Sin
Vulnerability is credited to the PC who caused it.

### Persistent damage

A creature keeps one persistent damage per type; a new one of the same type
replaces it only when its average is higher. A critical hit doubles it. At the
end of the bearer's turn each one deals its damage through the same steps as
any damage, then the bearer rolls a DC 15 flat check, and a success ends it. A
creature immune to the type never gains it.

### Dying and death

When a standing creature drops to 0 HP:

- **an NPC dies**
- **a PC is knocked out** — it gains prone (logged as off-guard, then prone),
  then goes dying: dying 1 plus its wounded value, plus 1 more on a critical
  hit. The PC also moves in the initiative order to directly before the
  creature that knocked it down, unless it dropped on its own turn. It stays
  marked as knocked out until it Stands, which fires its `stood-up` triggers
  (Too Angry to Die), also after a fast-healing revive.

Damage that gets through to a dying PC raises its dying value by 1, or by 2 on
a critical hit. A stable PC at 0 HP that takes damage goes dying again with the
knock-down value, but does not move in the initiative order.

The **death value** is the creature's `dying_max` (4 when unset) minus its
doomed value. When a PC's dying value would reach it, the PC dies, unless it
has a hero point: then Heroic Recovery spends every hero point and sets dying
to 0. Hero points come from the creature's `hero_points`.

A dying PC makes a recovery check at the start of its turn: a flat check (the
d20 alone) against its `recovery_dc` (10 when unset) plus its dying value,
capped by any eased recovery DC, with natural 20 and natural 1 shifting the
degree like any check. A critical success lowers dying by 2, a success by 1; a
failure raises it by 1, a critical failure by 2. A PC whose dying value reaches
0 is stable: it stays at 0 HP, and its wounded value rises by 1.

**Drained** N takes level × N HP when it is gained, and lowers maximum HP by
the same. A drain that takes a standing creature to 0 HP knocks it down.

### The end of a fight

After each action and each turn the engine checks the sides. The party wins
when no enemy is standing; the enemies win when no party member is standing (a
dying PC is not standing). If the round limit passes with both sides standing,
the fight is a stalemate.

### The decider

```go
type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
	React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool)
}
```

A `Choice` has a kind (`ChooseStrike`, `ChooseAbility`, `ChooseStand`,
`ChooseSustain`), an ID, a target and `Actions`, the cost picked for a
cost-range ability (0 means the lowest). `React` answers a trigger with one of
the offered ability IDs, or none; an ID that was not offered is a fault.

`FirstFitting` is the only decider today, a fixed behaviour for tests. It
Stands when prone, then Sustains the first waiting effect, then takes the first
ability by ID that is not ignored, not triggered, has no `targets`, fits the
open actions and is not blocked; if none fits, it takes the creature's first
Strike by ID. Its target is the first standing foe in setup order, and
`NextTarget` gives the same answer. Its `React` always takes the first ability
offered. It never aims an ability at an ally.

### Events

The `Observer` has one method, `Observe(Event)`. Each event kind is its own
struct, and events name combatants by ID. Events are added, never changed: each
new rule has its own event that names who, which ability and what set it off.

| Event | Fires when |
|-------|------------|
| `InitiativeRolled` | A combatant rolls initiative, before the first round |
| `RoundStarted` | A round begins |
| `TurnStarted` | A combatant that is not dead starts its turn |
| `ActionsRegained` | A combatant regains its actions while quickened or losing some: the count, quickened, and how many were lost |
| `ActionsLost` | Stunned gained on the bearer's own turn takes actions at once |
| `ActionsGained` | A `gain_actions` step adds actions |
| `ActionUsed` | A combatant spends actions on a Strike, an ability, a Stand or a Sustain |
| `StoodUp` | A Stand removes prone |
| `UseSpent` | A `per_fight` use is spent, with the uses left |
| `PoolSpent` | A pool count is spent, with the count left |
| `RechargeRolled` | A recharge is rolled on use, with its rounds |
| `Recharged` | A recharge ends: counted down (no `By`), or by a `recharge` step (`By` names the ability, or the Strike when no ability set it off) |
| `Triggered` | A triggered ability fires: the owner, the ability, the kind, the cost, and the creature, ability and Strike that set it off |
| `ActionDisrupted` | A critical Strike disrupts a pending action |
| `StrikeRolled` | An attack roll is made, with its die, bonus, penalty, total, AC and degree |
| `SpellAttackRolled` | A spell attack is made, with the same fields |
| `SaveRolled` | A creature saves, with its source, ability, die, bonus, total, DC and final degree |
| `SkillChecked` | A skill check is made, with its target, ability, skill, die, bonus, total, DC and degree |
| `RerollGained` | A `reroll` step gives a fortune or misfortune effect |
| `Rerolled` | A d20 is rolled twice: both dice, the one kept and the ability |
| `RerollsCancelled` | Fortune and misfortune cancel on one roll |
| `BonusGained` | A `bonus` step gives a bonus |
| `RecoveryEased` | A `recovery_dc` step caps a recovery DC |
| `FlatChecked` | A flat check is rolled, for concealment or to end persistent damage |
| `ImmuneToEffect` | An ability step skips a target immune to its traits |
| `DamageDealt` | Damage lands: each type rolled and taken, the trait adjustment, the total, the temporary HP absorbed, the HP left and `TriggeredBy` |
| `HPLost` | HP are lost without damage (drained) |
| `Healed` | A creature is healed, with the HP after and `TriggeredBy` |
| `Revived` | Healing lifts a dying PC: the dying value it had and its new wounded value |
| `TempHPGained` | A creature gains temporary HP |
| `ConditionChanged` | A condition's value changes on a creature, from and to |
| `PersistentChanged` | A creature gains or replaces a persistent damage |
| `PersistentEnded` | A persistent damage ends on a flat check |
| `DefenseChanged` | A `weaken` step raises weaknesses or lowers resistances |
| `DefenseRestored` | A `weaken` effect's duration ends |
| `ResistanceAdapted` | An adaptive resistance switches to a new damage type |
| `LoyaltyAct` | A creature does an act of loyalty, with the reason |
| `TraitGranted` | A `grant` step gives a trait against a creature |
| `TraitSpent` | An attack carries and spends a granted trait |
| `Sustained` | A creature Sustains an effect |
| `SustainEnded` | A sustained effect ends |
| `InitiativeMoved` | A knocked-down PC moves to just before the creature that knocked it down |
| `DyingChanged` | A PC's dying value changes |
| `RecoveryChecked` | A dying PC makes a recovery check |
| `Stabilized` | A PC's dying value reaches 0 on a recovery check, with its new wounded value |
| `HeroicRecovery` | A PC spends its hero points to avoid death |
| `Died` | An NPC drops to 0 HP, lethal damage drops a creature, or a PC reaches its death value |
| `FightEnded` | One side has no standing creature, or the round limit passes (a stalemate) |

### Not modelled yet

- **Effort 5 (planned)** — spells: slots, ranks and heightening, counteract,
  incapacitation, temporary immunity, random tables and confused. Faithful
  Strike's "damaging spell" waits for spells.
- **Position (`A.004.2`)** — every creature counts as engaged with every foe:
  an area catches its whole side, splash lands on the primary target only,
  every foe may answer a provoking action, and reach, range, movement and fall
  damage are not checked. The stand-ins until then: Walk the Timelines is two
  Strikes and no move, Farabellus Flip is the AC bonus without the Step, Lash
  Out is the damage without the shove, Ring of the Ram has no push, Healing
  Wave catches the whole party, Return Force waits, and the only `move` action
  is Stand.
- **Not modelled by choice** — hands (no draw, re-grip or dropped weapon; a
  potion is 1 action, Robert's table rule) and the penalties of unconscious (no
  creature can target a downed PC).
- **Known gaps** — the plan's Decision Record lists each one; the larger are:
  a press reaction outside the user's turn is not blocked; one pick is offered
  per trigger, not one free action and one reaction; turn-start triggers run
  after actions are regained; a critical degree with no list of its own runs
  nothing on a save or a check; a triggered ability never provokes, so a
  manipulate reaction cannot be disrupted; passive triggered abilities run in
  ID order before the owner is asked to react.
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
- **Abilities are recursive steps with a list of steps per degree of success.** A save, a spell attack, a check or a Strike runs the list for its degree, so a rider, a condition and a nested save share one shape, and the converter fills it the way PF2e text is written. (`F.041.E3` Decision 6.)
- **Condition rules are a Go table of shared parts.** Each row names what the condition lowers, the bonus type and whether it decays; code exists only where a part cannot say the rule. The ruleset as data will lift the table into files. (`F.041.E3` Decision 2.)
- **The sheet's number is one untyped base; fight-time bonuses stack among themselves.** A fight-time bonus of a type the build already applies is written as the difference. (`F.041.E3` Decision 10.)
- **A triggered ability is an ordinary ability with a `trigger` table.** It has the same steps, limits and events as a chosen one; one of ten kinds sets when it fires, and an ability with two triggers is two abilities. (`F.041.E4` Decisions 20 and 35.)
- **Reactions and free triggered actions are decisions behind `React`.** The engine offers what fits; the decider picks one or none, so Effort 7's rule lists answer it behind the same seam. A passive trigger fires with no decision. (`F.041.E4` Decision 21.)
- **The multiple attack penalty is counted per own turn.** A reaction Strike on another creature's turn takes none and does not count, as the rules say. (`F.041.E4` Decision 24.)
- **Events are added, never changed, and each names its cause.** Each new rule has its own event naming who, which ability and what set it off; `TriggeredBy` lets damage be credited to the creature that caused it. (`F.041.E4` Decision 26.)
- **Every "all engaged" rule sits behind one seam.** `areaTargets` answers who an area catches, `provoke` answers who may answer an action, and a `foe-acts` trigger's `traits` name what provokes; `A.004.2`'s position replaces these and nothing else. (`F.041.E3` Decision 1; `F.041.E4` Decisions 12 and 13.)

## Dependencies

**Depends on** [`creature`](data-files.md) for the loaded stat blocks, and on
`dice` (on this page) for the roller and dice expressions.

**Used by** the gate 3 test, `TestRunIsDeterministic`, which plays Luna against
three dummy PCs on ten seeds and checks that each seed gives the same events
twice, and the scope tests (`TestScopeRows`), which run each Effort 4 rules-
scope row from `testdata/scope/`. No command runs a fight yet; the runner
(planned, Effort 8) will.
