# Data Files

> **Code:** `internal/pf2e/sim/creature/`, `internal/pf2e/sim/foundry/`, `internal/campaigns/` (`Paths`)

## Purpose

How a creature's files become one stat block. Each creature is a folder of
TOML layers in the campaign. `creature.Load` decodes each layer strictly,
merges them in a fixed order, and checks the result into one
`creature.Creature` that the engine and `show` read. `creature.Locate` finds a
creature's folder by its name.

## Shape

### Where creatures live

Two campaign folders hold creatures. Both come from `campaigns.Paths`, so
`campaign create` makes them in every new campaign.

| Folder | `Paths` field | Holds |
|--------|---------------|-------|
| `pf2e/party/<pc>/` | `PartyDir` | The party, reused by every fight |
| `encounters/<fight>/<creature>/` | `EncountersDir` | One fight's creatures, next to its `source/` and `reference/` |

The folder name is the creature's ID. The engine names combatants by it, and
`show` looks it up.

### Layers and their order

A creature folder holds up to three kinds of layer, and `creature.Load` reads
them in this order:

1. `import.toml` — written by `gm-toolkit pf2e sim import` (see
   [Import](#import))
2. `hand.toml` — written by hand: fixes, homebrew and ability logic
3. `<variant>.toml` — one per variant asked for, in the order asked. On the
   command line each is a `--variant <name>` flag, repeated to stack.

At least one of `import.toml` and `hand.toml` must exist, or the load fails
with `ErrNoLayers`. A variant that was asked for must exist. A variant name is
a plain file name: it cannot be empty, `.` or `..`, hold a slash or a
backslash, or name `import` or `hand`; such a name fails with `ErrBadVariant`.
The loaded creature keeps the list of layers it read in `Layers`, and `show`
prints it.

### The merge rule

Each layer is decoded to a plain TOML tree, and the trees merge in order by one
rule:

- **A table merges key by key.** When both the earlier tree and the later layer
  hold a table under the same key, the merge goes inside it.
- **Every other value is replaced.** Numbers, strings, booleans and lists in a
  later layer replace the earlier value whole. A list is never merged item by
  item.

Strikes and abilities are keyed tables (`[strike.slash]`,
`[ability.abyssal-frenzy]`), so a hand layer can change one Strike's bonus
without restating the Strike, or add an ability beside the imported ones. A
Strike's `damage` and `traits` are lists, so a layer that sets either replaces
it whole. Merged tables keep no file order: code that walks Strikes or abilities
sorts them by ID (`StrikeIDs`, `AbilityIDs`).

### The file format

Abyssal Luna's two layers, cut to one Strike, written here as one file. Her
`import.toml` holds the numbers and Strikes, and her `hand.toml` holds only the
ability (`campaigns/hawk/encounters/fall-of-luna/abyssal-luna/`):

```toml
schema = 1

name       = "Abyssal Luna"
kind       = "npc"
level      = 16
perception = 29

[defense]
hp        = 300
ac        = 36
immune    = ["confused", "doomed", "paralyzed", "sleep"]

[spellcasting]
attack = 31
dc     = 39

[strike.slash]
name   = "Slash"
bonus  = 30
traits = ["agile", "magical", "reach-10", "unarmed"]
damage = [
  { dice = "3d6+14", type = "slashing" },
  { dice = "1d8", type = "mental" },
]

[ability.abyssal-frenzy]
name    = "Abyssal Frenzy"
actions = 2
steps   = [{ strike = "slash" }, { strike = "slash" }, { strike = "stab" }]
```

The top level holds `name`, `kind` (`pc` or `npc`), `level`, `traits`,
`perception`, `class_dc`, `hero_points`, `dying_max` (4 when unset) and
`recovery_dc` (10 when unset). `[speed]` holds `land` and `fly`. `[defense]`
holds `hp`, `ac`, the three saves, `immune`, the `resist` and `weakness`
tables, each entry a `value` with an optional `except` list, and `adaptive`, a
resistance `value` that moves to the last of its `types` the creature took
damage from. A `resist` or `weakness` key is a damage type, `physical`, `all`,
or a trait (`adamantine`, `cold-iron`, `holy`, `silver`, `unholy`) that applies
once per hit. `[spellcasting]` holds `attack`, `dc` and `top_rank`. An NPC sets
`class_dc` and `[spellcasting]` directly; a PC's come from its build (see
[PC builds](#pc-builds)).

A Strike has a name, an attack bonus, a `range` (unset for melee), traits and
one or more damage parts. A damage part is a dice expression, a damage type
and an optional `category`: `persistent` or `splash`. A Strike also carries
riders: the four degree lists `critical_success`, `success`, `failure` and
`critical_failure`, each a list of steps.

Three tables hold rules that depend on the incoming effect, for PCs and NPCs
alike:

| Table | Holds |
|-------|-------|
| `[modifier.<id>]` | A fight-time bonus or penalty: `selector`, `type`, `value`, optional `against` traits, optional `per_condition` |
| `[degree_shift.<id>]` | A save result that moves: `selector` (a save or `saving-throw`), optional `against` traits, `from` and `to` degrees |
| `[condition_adjust.<condition>]` | How a condition lands on this creature: `gain` added to each value, `decay` per turn |

An ability has a name, an action cost, `traits`, an optional `area` (`shape`
of `burst`, `cone`, `emanation` or `line`, and a `size` in feet) and a list of
steps. Each step names one kind by its key:

| Kind | Key | Other keys |
|------|-----|------------|
| Strike | `strike = "<id>"` | Degree lists |
| Save | `save = "fortitude"` | `dc` or `dc_from` (`class`, `spell`), `basic` with `damage`, `shift`, degree lists |
| Spell attack | `spell_attack = true` | `damage`, degree lists |
| Damage | `damage = [...]` | `lethal` |
| Heal | `heal = "<dice>"` | — |
| Temporary HP | `temp_hp = "<dice>"` | — |
| Condition | `condition = "<name>"` | `value` for a valued condition, `duration` |
| Weaken | `weaken = { weakness, resistance, types }` | `duration` |

Every step may also take `target = "self"` and `unless = "<condition>"`. A
`duration` is `{ rounds = <n> }` or `{ until = "<boundary>" }`, where the
boundary is `source-turn-start`, `source-turn-end`, `target-turn-start` or
`target-turn-end`. See [the engine](engine.md#abilities-and-steps) for what
each kind does.

An ability with an area, a basic save and a `failure` list (an example, not a
real creature's ability):

```toml
[ability.spore-cloud]
name    = "Spore Cloud"
actions = 2
traits  = ["occult", "fungus"]
area    = { shape = "cone", size = 30 }
steps   = [
  { save = "will", dc_from = "spell", basic = true, damage = [{ dice = "6d6", type = "mental" }],
    failure = [{ condition = "frightened", value = 1 }] },
]
```

Speed is loaded and shown, but no rule reads it yet.

### Strict per-layer decode

Every layer declares `schema = 1`. Each layer is first decoded on its own into
the typed `Creature`: a key the type does not know fails the load and names the
file and the key, and a schema other than `creature.SchemaVersion` fails it
too. Every key must be lower case: the typed decode matches field names without
case but the merge does not, so `HP` beside `hp` would survive the merge and
either could win. Only then is the layer read as a plain tree for the merge. A
typo is caught in the file that holds it, not in the merged result.

### What the load checks

A creature with a `[build]` table is first checked and computed from its build
(see [PC builds](#pc-builds)). Then `prepare` checks the creature and reports
every problem it finds at once, wrapped in `ErrInvalid`:

- the name is set, and the kind is `pc` or `npc`
- `defense.hp` and `defense.ac` are each at least 1
- `hero_points` and `recovery_dc` are not negative, and `dying_max` is unset or
  at least 2
- each `resist` and `weakness` key is a damage type, `physical`, `all` or a
  defense trait, and its value is at least 1
- `defense.adaptive` has a value of at least 1 and only damage types
- each `[modifier.*]` has a known selector and bonus type and a value other
  than 0; `against` sits only on `ac` or a save, and `per_condition` names a
  condition and sits only on an attack or damage selector
- each `[degree_shift.*]` has a save selector and two known degrees
- each `[condition_adjust.*]` names a condition, sets `gain` or `decay`, and
  has no negative `decay`
- each Strike has a name, a range that is not negative, and at least one damage
  part with no category; each part's dice parse, its type is a known damage
  type, and its category is `persistent`, `splash` or none
- each ability that is not ignored has a name, an action cost of 1 to 3, at
  least one step, and a known area shape of at least 5 feet
- each step, in an ability or a degree list, names exactly one kind; a Strike
  step names a Strike the creature has; a save names a save and one DC source,
  and `dc_from` finds a class DC or `[spellcasting]`; a spell attack needs
  `[spellcasting]`; `basic` and damage on a save come together; `dc`,
  `dc_from`, `basic` and `shift` sit only on a save, and `shift` is −3 to 3
- degree lists sit only on a Strike, a save or a spell attack
- a condition is known, with a value of at least 1 when valued and none
  otherwise; a duration sits only on a condition or a weaken step and sets
  `rounds` or a known `until`, not both; `unless` names a condition; `lethal`
  needs damage; `target` is `self` or unset; a weaken step names damage types
  and a weakness or resistance above 0

`prepare` also parses each damage part's dice and each heal and temporary HP
amount into a `dice.Expression`, and stamps each Strike and ability with its
ID.

### PC builds

A PC whose layers hold a `[build]` table stores its choices, not its numbers.
`Load` computes the numbers after the merge and before `prepare`, so `show`, the
import report and the engine all read the same result. A level-up is a layer of
new choices (`level = 15`, a `level-15` boost list), and every number follows.

| Key | Holds |
|-----|-------|
| `ancestry_hp` | Hit points from the ancestry |
| `class_hp` | Hit points per level from the class; at least 1 |
| `key` | The key attribute, for the class DC |
| `speed` | Base land speed |
| `flaws` | Attributes that take a flaw |
| `weapon_specialization` | `normal`, `greater`, or unset for none |
| `critical_specialization` | The lowest weapon rank that grants critical specialization (`trained` to `legendary`), or unset for none |
| `[build.boosts]` | One attribute list per source: `ancestry`, `background`, `class`, `level-<n>` |
| `[build.proficiency]` | A rank word per proficiency: `perception`, the saves, `class-dc`, `spellcasting`, armor and weapon categories, `weapon-group-<group>`, `weapon-base-<base>` |
| `[build.armor]` | The worn armor: `name`, `category`, `ac_bonus`, `dex_cap`, `potency`, `resilient`, `speed_penalty` |
| `[build.weapon.<id>]` | One weapon per Strike: name, category, group, base, dice, die, damage type, potency, striking, `item_bonus`, range, traits, `persistent` (a dice and a type), `splash`, and riders in the four degree lists |
| `[build.modifier.<id>]` | One bonus or penalty: `selector`, `type`, `value`, `per_level`, `weapon_traits`, and `attribute` on a `splash` modifier |
| `[build.spellcasting]` | The casting `attribute` and the `top_rank` |

**Attributes.** Each starts at +0. Flaws come off first. Then boosts apply in a
fixed order: `ancestry`, `background`, `class`, then each `level-<n>` list by
number. A boost raises an attribute by 1 up to +4; above that, a boost is a
partial, and two partials make +1. A boost source with any other name fails the
load.

**Ranks.** A rank is a word: `untrained`, `trained`, `expert`, `master` or
`legendary`. A proficiency the build does not name is untrained. The
proficiency bonus is the level plus twice the rank (trained 2, legendary 8), and
0 when untrained. A weapon uses the highest of its category's rank, its
`weapon-group-<group>` rank and its `weapon-base-<base>` rank. An advanced
weapon skips the group rank: Foundry gives a group's advanced weapons a lower
rank than its simple and martial ones, and the format holds only the second.

**The numbers.**

- **HP** — `ancestry_hp`, plus the level times `class_hp` plus Con, plus `hp`
  modifiers.
- **AC** — 10, plus Dex up to the armor's `dex_cap`, plus the armor category's
  proficiency, plus `ac` modifiers. The armor's `ac_bonus` plus its potency is
  an item bonus. With no `[build.armor]`, the PC is unarmored with no cap.
- **Saves** — the save's proficiency plus Con (Fortitude), Dex (Reflex) or Wis
  (Will), plus modifiers on that save and on `saving-throw`. The armor's
  `resilient` is an item bonus.
- **Perception** — its proficiency plus Wis, plus `perception` modifiers.
- **Land speed** — `speed` plus the armor's `speed_penalty`, plus `land-speed`
  modifiers.
- **Class DC** — 10 plus the `class-dc` proficiency plus the key attribute.
- **Spell attack and DC** — the `spellcasting` proficiency plus the casting
  attribute; the DC is 10 more. `top_rank` is copied.
- **A Strike's attack** — the weapon's proficiency plus an attribute, plus
  `attack` modifiers. The attribute is Dex for a weapon with a range, the higher
  of Str and Dex for a finesse weapon, and Str otherwise. The higher of potency
  and the weapon's own `item_bonus` is an item bonus.
- **A Strike's damage** — `dice` plus `striking` dice of size `die`. The
  modifier adds Str in melee and for a thrown weapon, half of a positive Str (or
  all of a negative one) for a propulsive weapon, and nothing at any other
  range. Specialization adds 2, 3 or 4 at expert, master or legendary, or 4, 6
  or 8 when `greater`. `damage` modifiers add last. A weapon's `persistent`
  becomes a persistent part. Its `splash`, plus `splash` modifiers and the
  attribute each one names (Expanded Splash adds Intelligence), becomes a
  splash part of the weapon's damage type.
- **A Strike's riders** — the weapon's degree lists. At or above
  `critical_specialization`, the weapon's group adds its critical
  specialization step from a Go table to the critical success list; the one
  row today is the hammer (a Fortitude save against the class DC, prone on a
  failure or a critical failure).

**Bonus stacking.** A modifier's `type` is `item`, `status`, `circumstance`, or
unset for untyped. Per type, only the highest bonus and the lowest penalty
count; untyped values all add. A modifier with `weapon_traits` applies only to
weapons that carry every trait it names. `per_level` is multiplied by the level
and added to `value`.

**Choices only.** If any layer has `[build]`, no layer may set a computed key:
`perception`, `class_dc`, `spellcasting`, any `[strike.*]`, `speed.land`, or
`defense.hp`, `ac`, `fortitude`, `reflex` or `will`. Each layer is checked as
written, before the merge, and the load fails with `ErrComputedKey`, naming each
file and key. A fix goes in as a build choice or a `[build.modifier.*]`.
Resistances, weaknesses, immunities, fly speed and abilities stay hand-settable.

Bad choices fail the load with `ErrInvalid`, every problem at once: a build on
an `npc`, a level below 1, an unknown attribute, rank word, proficiency, armor
or weapon category, damage type, selector or bonus type, `weapon_traits` on a
selector other than `attack`, `damage` or `splash`, an `attribute` on a
selector other than `splash`, a `persistent` whose dice or type the load
rejects, a `critical_specialization` that is not a rank above untrained, and
negative numbers where none belong.

### Ignored abilities

`ignore = true` on an ability marks it as skipped on purpose. `prepare` does not
check it, the engine's `FirstFitting` never picks it, and the engine refuses it
as a choice. It stays on the creature, and `show` lists it as
`<id> (ignored)`, so the choice to skip it stays visible.

### Finding a creature

`creature.Locate(partyDir, encountersDir, id)` looks for a folder named `id` in
the party folder and in every fight folder under `encounters/`. It returns the
one folder it finds. A name found in more than one place fails with
`ErrCreatureAmbiguous`, which lists every folder found; a name found nowhere
fails with `ErrCreatureNotFound`. An empty name, `.`, `..` or a name with a
path separator fails with `ErrBadCreatureID`. A campaign with no `encounters/`
folder is not an error.

### Import

`gm-toolkit pf2e sim import <fight>` runs `foundry.ImportFight` on the active
campaign. It reads every `encounters/<fight>/source/actors/*.json` Foundry actor
export. The export's file name, without `.json`, is the creature's folder name:
a PC goes to `pf2e/party/<pc>/` and an NPC to `encounters/<fight>/<npc>/`. The
file name must be a slug (lowercase letters, digits and dashes), and an NPC
cannot be named `source`, the fight's export folder; the import stops at the
first file that breaks this. A fight with no actor exports fails with
`ErrNoExports`. Loose item exports in
`source/items/*.json` are read and reported as not placed on a creature.

Each written `import.toml` starts with a header comment that names its export
and says the import rewrites it. The import writes the file whole every time;
changes belong in `hand.toml`.

**NPCs.** Numbers come across as Foundry stores them: level, traits (with the
rarity first when it is not common), Perception, speeds, HP, AC, saves and
immunities. Resistances and weaknesses keep their `except` lists, and
`all-damage` becomes `all`; one keyed on a defense trait (the Avatar's `holy`)
comes across as it is. Strikes keep the export's damage order, and a part
with a `persistent` or `splash` category keeps it. The first spellcasting
entry gives the spell attack and DC. Left out, each with its reason in the
report: every ability, every spell, a weakness or resistance keyed on anything
else, the doubling of a resistance (the resistance itself is written), an
immunity's exceptions, attack effects, damage with any other category, damage
whose dice or type the load would reject, and a Strike left with no
uncategorized damage.

**PCs.** The importer writes the build, never the sheet's totals. Where each
choice comes from:

- **Class** — `class_hp`, the key attribute and its `class` boost, and the
  starting ranks. The class DC starts at trained, as pf2e system 8.6.1 gives
  every class.
- **Ancestry** — `ancestry_hp`, `speed`, traits, boosts and flaws. Alternate
  ancestry boosts replace both the boosts and the flaws.
- **Background** — its boosts.
- **Level boosts** — the sheet's per-level lists, as `level-<n>`.
- **Feats** — rank raises from their proficiency subfeatures; a raise that names
  the class raises `class-dc`. The `weapon-specialization` and
  `greater-weapon-specialization` feats set `weapon_specialization` by slug.
- **Armor** — the one worn (and, if it takes investing, invested) armor, with
  its runes. Property runes are reported, to be typed into the hand layer.
- **Weapons** — every weapon with a quantity above 0, carried or stowed, whose
  damage dice and type the load accepts, with its persistent damage and
  splash. Property runes are reported, to be typed into the hand layer as
  riders (Kas's Greater Crushing and Greater Fearsome).
- **Spellcasting** — the entries that hold spells and name an attribute; the
  first sets the attribute, and the top rank is the highest slot with any uses.

Foundry keeps item bonuses and rank changes in rule elements. The importer
resolves the ones the computed numbers need into build choices:

| Rule element | Feeds |
|--------------|-------|
| `FlatModifier` on `hp`, `ac`, a save, `saving-throw`, `perception` or `land-speed` | A `[build.modifier.*]`; a value of exactly `@actor.level` becomes `per_level = 1`, any other formula is worked out at the export's level |
| `FlatModifier` on an attack or damage selector, scoped by weapon traits | A modifier with `weapon_traits` |
| `Resistance` | `defense.resist`, the highest value per damage type |
| `ActiveEffectLike` with mode `upgrade` on a rank | `[build.proficiency]` |
| `MartialProficiency` for a weapon group's martial weapons | A `weapon-group-<group>` rank; the rule for the group's advanced weapons is reported as not supported |
| `AdjustModifier` that suppresses `armor-speed-penalty` | `speed_penalty = 0` |
| `AdjustDegreeOfSuccess` on a save or `saving-throw` | A `[degree_shift.*]` per adjusted degree; a bare trait in the predicate (`fear`, `item:trait:emotion`) becomes `against` |
| `ActiveEffectLike` on `dying.max` (mode `upgrade`) | `dying_max`, the highest value |
| `ActiveEffectLike` on `dying.recoveryDC` (modes `subtract`, `downgrade`) | `recovery_dc`: 10 minus every subtraction, then capped by the lowest downgrade, as Foundry orders the modes |
| `FlatModifier` on `initiative` | A `[modifier.*]` on `initiative`, when its predicate needs nothing beyond `perception` |
| `FlatModifier` with `damageCategory` `splash` and an attribute formula | A `splash` build modifier with `attribute` and `weapon_traits` |
| `CriticalSpecialization` with a proficiency-rank predicate | `critical_specialization`, the lowest rank found |
| `AdjustStrike` on `materials` | The material trait added to that weapon (the Nail Bomb's `cold-iron`) |
| `DexterityModifierCap` | The lowest cap becomes the armor's `dex_cap` |
| `RollOption` that is a toggle | Nothing: reported as a toggle, imported off |

Rules count only on items whose effect is always on: rules on conditions,
effects and spells are not read, and a physical item's rules count only when it
is worn or held and, if it takes investing, invested. A rule's predicate sees
only the actor's class, feats, heritage and worn armor. A rule that needs a
target, an action, a condition or a toggle does not hold, and it is reported as
conditional. A numeric comparison (`gte`, `lt` and the like) is reported as not
supported, except the proficiency rank that `CriticalSpecialization` reads. Every toggle is imported off, even one the sheet stored on. Every
other rule element is reported with its reason, or counted when the sim never
uses its kind (`Note`, `GrantItem` and the like).

**The report.** For each creature, the command loads the written folder (the
import layer plus any hand layer) and prints its stat block as `show` does.
After it come the skipped items grouped by reason, then the counts of rule
kinds the sim never uses. A creature that fails to load is reported where its
stat block would be, and the command fails after every creature has printed.

## Key Decisions

- **A creature is an import layer the importer owns plus hand layers, merged on
  load.** Re-import stays one safe step; fixes, homebrew and level-up variants
  sit in hand layers. (`F.041` D-7.)
- **One merge rule over keyed tables.** Strikes and abilities are keyed tables,
  so tables merge key by key and every other value is replaced by the later
  layer. No list is merged by ID, and merged tables keep no file order.
  (`F.041.E1` Decision 4.)
- **Every layer declares its schema version and is decoded strictly on its
  own.** An unknown key names its file, and Effort 6's approval cache keys on
  the schema version. (`F.041.E1` Decision 4.)
- **An ignored ability stays on the creature.** The engine skips it, but `show`
  and `check` can still see that it was ignored on purpose. (`F.041.E1`
  Decision 4; D-8.)
- **A PC's import layer holds build choices, and load computes the
  numbers.** A level-up layer states what changed and the numbers follow;
  `show`, the import report and the sim share one computation. (`F.041`
  D-10; `F.041.E2` Decision 2.)
- **The importer resolves Foundry's rule elements; the format never holds a
  predicate or a formula.** Only `@actor.level` survives, as `per_level`.
  Foundry's rule-element language stays out of the format and the engine.
  (`F.041` D-1, D-4; `F.041.E2` Decision 2.)
- **A PC with a build takes choices only.** A layer that sets a computed
  number fails the load, so no stale override hides a level-up.
  (`F.041.E2` Decision 3.)
- **The import is a baseline: toggles off, sheet state dropped, conditional
  rules skipped.** Every run starts from the same character, and the report
  lists each thing left out. (`F.041` D-26; `F.041.E2` Decisions 5 and 6.)
- **Fight-time modifiers and degree shifts are creature tables for PCs and NPCs alike.** They depend on the incoming effect, so a build cannot compute them, and a monster needs them too. (`F.041.E3` Decision 3.)
- **Critical specialization is a rider from a table by weapon group.** The build holds the rank that grants it, and the load adds the group's rider to each weapon at or above it. (`F.041.E3` Decision 15.)

## Dependencies

**Depends on** [`dice`](engine.md#the-roller) for dice expressions and
`BurntSushi/toml` for decoding. `creature` does not import `internal/campaigns`:
the commands read the folders from `campaigns.Paths` and pass them to `Locate`.

**Used by** the [engine](engine.md), which plays loaded creatures, and
`gm-toolkit pf2e sim show`. The importer, `internal/pf2e/sim/foundry`, builds
`creature.Creature` values and writes them as `import.toml`; the `import`
command then loads each one. `creature` does not import `foundry`. `check`
(planned, Effort 6) will read loaded creatures.
