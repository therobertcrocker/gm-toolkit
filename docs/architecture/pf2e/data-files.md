# Data Files

> **Code:** `internal/pf2e/sim/creature/`, `internal/campaigns/` (`Paths`)

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

1. `import.toml` — written by the importer (planned, Effort 2)
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

Luna's hand layer, cut to one Strike and one ability
(`campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml`):

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

The top level holds `name`, `kind` (`pc` or `npc`), `level`, `traits` and
`perception`. `[speed]` holds `land` and `fly`. `[defense]` holds `hp`, `ac`,
the three saves, `immune`, and the `resist` and `weakness` tables, each entry a
`value` with an optional `except` list. A Strike has a name, an attack bonus,
traits and one or more damage parts, each a dice expression and a damage type.
An ability has a name, an action cost and a list of steps; in schema 1 the only
step is a Strike, named by its ID.

Level, saves and speed are loaded and shown, but no rule reads them yet.

### Strict per-layer decode

Every layer declares `schema = 1`. Each layer is first decoded on its own into
the typed `Creature`: a key the type does not know fails the load and names the
file and the key, and a schema other than `creature.SchemaVersion` fails it
too. Every key must be lower case: the typed decode matches field names without
case but the merge does not, so `HP` beside `hp` would survive the merge and
either could win. Only then is the layer read as a plain tree for the merge. A
typo is caught in the file that holds it, not in the merged result.

### What the load checks

After the merge, `prepare` checks the creature and reports every problem it
finds at once, wrapped in `ErrInvalid`:

- the name is set, and the kind is `pc` or `npc`
- `defense.hp` and `defense.ac` are each at least 1
- each `resist` and `weakness` key is a damage type, `physical` or `all`, and
  its value is at least 1
- each Strike has a name and at least one damage part; each part's dice parse,
  and its type is a known damage type
- each ability that is not ignored has a name, an action cost of 1 to 3, at
  least one step, and steps that each name a Strike the creature has

`prepare` also parses each damage part's dice into a `dice.Expression` and
stamps each Strike and ability with its ID.

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

## Key Decisions

- **A creature is an import layer the importer owns plus hand layers, merged on
  load.** Re-import stays one safe step; fixes, homebrew and level-up variants
  sit in hand layers. (`F.041` D-7.)
- **One merge rule over keyed tables.** Strikes and abilities are keyed tables,
  so tables merge key by key and every other value is replaced by the later
  layer. No list is merged by ID, and merged tables keep no file order.
  (`F.041.E1` Decision 4.)
- **Every layer declares its schema version and is decoded strictly on its
  own.** An unknown key names its file, and Effort 4's approval cache keys on
  the schema version. (`F.041.E1` Decision 4.)
- **An ignored ability stays on the creature.** The engine skips it, but `show`
  and `check` can still see that it was ignored on purpose. (`F.041.E1`
  Decision 4; D-8.)

## Dependencies

**Depends on** [`dice`](engine.md#the-roller) for dice expressions and
`BurntSushi/toml` for decoding. `creature` does not import `internal/campaigns`:
the commands read the folders from `campaigns.Paths` and pass them to `Locate`.

**Used by** the [engine](engine.md), which plays loaded creatures, and
`gm-toolkit pf2e sim show`. The importer (planned, Effort 2) writes
`import.toml`, and `check` (planned, Effort 4) reads loaded creatures.
