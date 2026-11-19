# PF2e Simulator — Overview

> The PF2e side's spine. Cross-subsystem narrative plus the page index.
> Anything about a single subsystem lives on that subsystem's page; only
> narrative that spans subsystems lives here.

> **Code:** `internal/pf2e/sim/dice/`, `internal/pf2e/sim/creature/`, `internal/pf2e/sim/foundry/`, `internal/pf2e/sim/engine/`, `cmd/gm-toolkit/pf2e/`

## Purpose

The PF2e side is a simulator: it plays one Pathfinder 2e fight many times and
reports how it tends to go, so the GM can tune an encounter before the table
sees it. Version 1 is `F.041`, built for the Fall of Luna finale. The ruleset
as data and an encounter designer are later work, in their own initiatives.

<br/>

## Shape

A fight moves through five stages. Four exist today; the runner arrives with
Effort 8 of `F.041`.

```
import → files → load → engine → runner (planned)
```

- **import** — `gm-toolkit pf2e sim import <fight>` reads the fight's Foundry
  exports and writes each creature's `import.toml` layer, then prints a report
  of what it left out. See [data files](data-files.md#import).
- **files** — each creature is a folder of TOML layers in the campaign, which
  the GM owns. The import writes one layer, and the GM writes the rest. See
  [data files](data-files.md).
- **load** — `creature.Load` merges a creature's layers in order and checks
  the result into one `creature.Creature`. See [data files](data-files.md).
- **engine** — `engine.Run` plays one fight from an `engine.Setup` and reports
  every step as an event. It holds the core rules and the effect core: saves,
  conditions, damage kinds and the Down row. See [engine](engine.md).
- **runner (planned, Effort 8)** — plays a scenario many times and scores the
  results.

<!-- diagram: the five-stage pipeline, with the planned runner stage dashed -->

**Packages.**

| Package | Holds |
|---------|-------|
| `internal/pf2e/sim/dice` | The seeded roller and dice expressions such as `3d6+14` |
| `internal/pf2e/sim/creature` | The creature format, the layer merge, PC builds computed on load, the checks on load and the folder lookup |
| `internal/pf2e/sim/foundry` | Reads Foundry exports and writes import layers, with a report |
| `internal/pf2e/sim/engine` | The rules, the fight loop, the events and the decision seam |
| `cmd/gm-toolkit/pf2e` | The `pf2e` and `sim` command groups, `show` and `import` |

The dependencies run one way: `engine` reads `creature` and `dice`, `foundry`
reads `creature`, `creature` reads `dice`, and the commands read `creature`,
`foundry` and `internal/campaigns`. No command runs a fight yet; today only the
engine's tests call `engine.Run`.

**The two seams.** The engine is headless. It talks to whatever drives it
through two interfaces:

- **`Decider` (decisions in)** — each step of a turn, the engine asks the
  decider what the creature does and at whom. Today the one decider is
  `FirstFitting`, a fixed behaviour for tests. Effort 7's tactics replace it
  behind the same seam.
- **`Observer` (events out)** — the engine reports every step through one
  `Observe(Event)` method. Events are plain values, so a test, a scorer or a
  replay can keep them in a list.

**Position (planned, Effort 7).** Every creature counts as engaged with every
foe, so an area catches every standing foe and splash lands on the primary
target only. Reach traits such as `reach-10` are loaded and shown, but nothing
checks them yet.

**Commands.** `gm-toolkit pf2e sim show <creature>` finds a creature in the
active campaign, merges its layers (plus any `--variant` layers) and prints the
stat block. `gm-toolkit pf2e sim import <fight>` writes the import layers for
every creature in a fight's Foundry exports and prints each one's stat block
and what it left out. `--campaign <id>` is a persistent flag on `pf2e`, so
every command under it takes it.

<br/>

## Page index

| Page | Scope | Status |
|------|-------|--------|
| [data files](data-files.md) | Creature folders, layers, the merge rule, the file format, PC builds, lookup, the import | Written |
| [engine](engine.md) | The roller, checks and modifiers, saves, ability steps, conditions, damage and dying, the fight loop, the decider and the events | Written |
| sim (planned) | Tactics, position, scenarios and the runner | Efforts 7 and 8 |

<br/>

## Key Decisions

- **The simulator has its own data model; Foundry is an import source.** The
  engine reads only its own files, so imported, hand-written and designer-made
  creatures are the same kind of file. (`F.041` D-1, D-6.)
- **All v1 code lives under `internal/pf2e/sim/`.** `internal/pf2e/` is the
  PF2e root; `internal/pf2e/rules/` is left for the ruleset as data. (`F.041`
  overview Decision 4.)
- **The commands are `gm-toolkit pf2e sim <command>`, with `--campaign`
  persistent on `pf2e`.** (`F.041` overview Decision 5.)
