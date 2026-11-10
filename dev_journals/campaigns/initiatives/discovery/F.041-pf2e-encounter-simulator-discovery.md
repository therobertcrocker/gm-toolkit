# PF2e Encounter Simulator — Discovery

> **Status:** draft for sign-off. No code is written until this doc is approved.

## In Plain Words

- **What:** a reusable tool inside gm-toolkit that plays a Pathfinder 2e fight thousands of times, using real character and monster data exported from Foundry, and reports how the fight tends to go: how often the party wins, how often someone dies, how many rounds it lasts, and which abilities actually get used.
- **Why:** balancing a high-level boss fight by hand is guesswork. The party's numbers are high and unusual (skill bonuses of +19 to +24), bosses are custom, and small changes (10 HP, one resistance, one GM habit) move the odds a lot. A simulator lets the GM test a change before the session instead of discovering it at the table.
- **How it fits:** it is the first PF2e system in gm-toolkit. It adds a new `internal/pf2e` rules engine (deterministic, with a `build` and `content` split) and a `sim` package on top of it, as one multi-effort feature. The first encounter it is built and tested against is the H.A.W.K. finale, The Fall of Luna: Abyssal Luna, then the Avatar of Lilith, with a complex hazard called the Gate to the Abyss.

## Problem

There is no way today to check a PF2e encounter's balance except reading the numbers and guessing. The encounter-building rules (XP budgets, creature-building tables) give a starting point, but they assume a typical party, and the H.A.W.K. party is not typical. The rules also say nothing about how a specific boss's abilities interact with a specific party's defences. Examples from the finale:

- Lilith's Protean Anatomy grants resistance 25 to acid, electricity or sonic after she takes that damage type. Several of Tavish's bombs deal those types, so his damage output against her depends on the order in which he throws them.
- Luna's Untethered to Fate lets her ignore fortune and misfortune effects. Arabesque's True Target does nothing against her.
- The Sin of Betrayal mechanic (acts of loyalty deal damage and grant holy Strikes) only matters if the party discovers it, and the discovery timing changes the fight's length.

A one-off script can answer one question about one fight, but it has to be rewritten for every new encounter. What is missing is a tool where the fight is described in data (who is in it, how each side behaves, what the hazards do, what counts as a good outcome) and the rules engine stays the same from fight to fight.

## Architecture Context

No architecture page covers PF2e yet. Every existing page describes the SWN faction engine. Pages read for this doc:

- `docs/architecture/architecture-overview.md`
- `docs/architecture/engine/overview.md`
- `docs/architecture/engine/dice.md`
- `docs/architecture/engine/ability-catalog.md`

Three of their decisions carry over to this design as principles, even though no code is shared.

- **The engine is headless** (`architecture-overview.md`, `engine/overview.md`). The faction engine takes decisions from a collector and reports events to an observer, and never knows who is on the other side. The simulator follows the same split. The rules engine resolves actions, the tactics layer supplies decisions, and a report layer consumes events. A later TUI or interactive "play it out step by step" mode can then be added without touching the rules.
- **Rules live in data** (`engine/ability-catalog.md`). The faction engine is moving away from bespoke Go handlers keyed by ability ID: abilities are described in rulebook data and run through handlers keyed by effect type, and the remaining per-ID handler map is the thing to retire. The simulator follows the same direction for creature abilities and for tactics. How to treat the few abilities that do not fit the primitives is Open Question 1.
- **Every roll goes through one roller** (`engine/dice.md`). The dice page says each type of roll has one named route. The simulator has one `Roller` interface with a seeded implementation, so any simulated fight can be replayed exactly by its seed.

The faction `dice` package itself is not reused. Its rolls are SWN-specific (`1d10 + stat`, contests, ties). PF2e needs a d20 check with four degrees of success, so `pf2e` gets its own small dice layer that only takes the roller-interface pattern from the faction package.

Two architecture pages should exist when this ships: `docs/architecture/pf2e/overview.md` covering the rules engine and the build pipeline, and `docs/architecture/pf2e/sim.md` covering the runner, tactics and reports. They make `pf2e` a third side of the wiki next to engine and interface, so `architecture-overview.md` changes too. The `F.041` row in `planned-work.md` records this.

## Design Summary

The simulator is four layers, each with one job.

1. **Combatant builder (`internal/pf2e/build`).** Reads Foundry actor exports (PCs, NPCs and hazards) and produces a *combatant*: a flat stat block with every number already computed. NPC and hazard exports already contain their numbers, so the builder reads them directly. PC exports do not. Foundry saves only the choices (ancestry, class, feats, gear, runes) and computes AC, saves and attack bonuses when the sheet opens. The builder recomputes them from the PF2e rules. Anything it cannot work out is listed in a build report and covered by a small hand-written *overlay* file instead of being guessed.

2. **Content (`internal/pf2e/content`).** Abilities, spells, conditions and hazard routines described as compositions of a small set of primitive effects: make a Strike, force a save, deal damage, apply a condition, heal, grant a bonus, recharge, and so on. Most of the finale's abilities fit the primitives. The few that do not (Walk the Timelines, Sin Vulnerability, Last Stand, Protean Anatomy) are Open Question 1.

3. **Rules engine (`internal/pf2e/engine`).** Runs one fight: initiative, turns of three actions plus a reaction, the multiple attack penalty, degrees of success, conditions and their durations, resistances and weaknesses, dying and wounded, recharges and frequencies, triggered reactions, hazard routines, round clocks and phase changes. It is deterministic for a given seed.

4. **Simulator (`internal/pf2e/sim`).** Loads a *scenario* (the fight's description), runs it thousands of times, sorts each run into outcome buckets the scenario defines, and reports the results. Each combatant's behaviour comes from a *tactics file*: an ordered list of conditions and actions, where the first matching entry decides the turn. This is the same shape as the turn trees in the Luna GM flowchart.

Position is modelled with **range bands**, not a grid. Each combatant is *engaged* with specific enemies, *near* (within one Stride), or *far*. Area effects hit a share of the party based on the bands and a per-scenario "bunching" setting. This is enough for every positional question the finale's tactics ask ("2+ PCs in a 40-ft cone", "PCs bunched in a 20-ft burst", "Arabesque exposed", "fly out of Kas's reach") without authoring a map for every fight. A grid can come later if a fight needs it.

The finale is the first scenario, and it is the test that proves v1 works. Level 14 and level 15 are two **variants** of one scenario. A variant applies overlay files on top of the base data (the level-15 overlays add the level-up choices and gear upgrades), so the comparison is one command, not two separate builds.

## Scope

### In v1

- Importing Foundry actor exports for PCs, NPCs and hazards (pf2e system 8.x, Foundry 14 format, as in the five exports already provided).
- Computing PC statistics for the three H.A.W.K. classes (fighter, sorcerer and legacy alchemist), with a build report listing anything not covered.
- Overlay files for corrections, missing items (for example Tavish's clan pistol) and level-up variants.
- The rules engine features the finale needs (listed under Rules Coverage below).
- Range-band positioning.
- TOML tactics files with a shared condition vocabulary and named GM dials.
- Complex hazards with parts, disable checks, a routine and round-based escalation, enough to run the Gate to the Abyss.
- Multi-phase scenarios with carry-over (Luna, the Transformation round, then Lilith).
- A Monte Carlo runner with a fixed seed option, outcome buckets defined in the scenario, and a plain-text CLI report plus a JSON results file.

## Out of Scope

- A grid or map-based movement.
- Full PF2e coverage. Each rule is added when a scenario needs it, and the build report says what is missing.
- Automatic level-up. A level-up variant is an overlay that the GM writes, because level-up choices are the player's.
- A TUI or a rendered report. The JSON results file is shaped so that either can be added later. The existing Dice Lab Foundry module could also read it.
- Player behaviour beyond the tactics file. The simulator plays the PCs the way their tactics files say, not the way the actual players would. This limit is permanent and the report states it.

## Domain Model

These are the main types. Field lists are indicative, not final.

### Combatant

A combatant is the fully computed stat block the engine plays with. It is the same type for PCs, NPCs and hazard parts.

```go
type Combatant struct {
    ID, Name     string
    Side         Side             // party or opposition
    Level        int
    Traits       []string
    HP           int
    AC           int
    Saves        map[Save]int     // fortitude, reflex, will
    Perception   int
    Skills       map[string]int
    Speeds       map[Movement]int // land, fly
    Immunities   []string
    Resistances  []Resistance     // value, damage type, exceptions
    Weaknesses   []Weakness
    Strikes      []Strike
    Abilities    []AbilityRef     // actions, reactions, passives, spells
    Resources    Resources        // spell slots, focus, consumables, uses per fight
}
```

### Ability and Effect

An ability is metadata (action cost, traits, frequency, recharge, trigger) plus a list of effect steps. Each step is one primitive. How the exceptions are expressed is Open Question 1.

| Primitive | Covers, for example |
|-----------|---------------------|
| `strike` | Slash, Stab, Hand of Chaos, Kas's hammer, bombs |
| `save` | Destiny Spores, Wail, Divine Wrath, Fireball |
| `spell_attack` | Disintegrate |
| `damage` | Lash Out, Sin Vulnerability's damage |
| `condition` | slowed, doomed, frightened, stunned, sickened, confused |
| `heal` | Healing Wave, potions, the Transformation checkpoint |
| `buff` | Haste, True Target, Faithful Strike |
| `counteract` | Dispel Magic |
| `multi` | Abyssal Frenzy, Demonic Frenzy (several Strikes in one activity) |

### Scenario

A scenario describes one fight: participants, variants, phases, hazards, dials and outcome buckets. A sketch for the finale:

```toml
[scenario]
id    = "fall-of-luna"
party = ["kas", "tavish", "arabesque"]
runs  = 5000

[[variant]]
id = "level-14"

[[variant]]
id       = "level-15"
overlays = ["party/level-15.toml"]

[[phase]]
id        = "luna"
enemies   = ["abyssal-luna"]
hazards   = ["gate-to-the-abyss"]
ends_when = "defeated:abyssal-luna"

[[phase]]
id         = "transformation"
free_round = true
checkpoint = { when = "party_struggling", heal_all = 75 }

[[phase]]
id      = "lilith"
enemies = ["avatar-of-lilith"]
carry   = ["doomed", "spent_resources", "hp"]

[dials]
dispel_mercy        = true
wail_spares_wounded = true
luna_picks_on_comfortable = true
```

### Tactics

A tactics file is an ordered list of rules. On its turn a combatant takes the first rule whose conditions all hold, and performs its action list. A sketch of Luna's flowchart as data:

```toml
[[rule]]
when = ["ready:destiny-spores", "in_area:cone-40 >= 2"]
do   = ["destiny-spores", "fly:away-from:kas"]

[[rule]]
when = ["dial:luna_picks_on_comfortable", "not party_struggling", "exposed:arabesque|tavish"]
do   = ["fly:engage", "abyssal-frenzy"]

[[rule]]
when = ["pinned_by:kas", "uses_left:walk-the-timelines"]
do   = ["walk-the-timelines", "fly:away-from:kas"]

[[rule]]
do   = ["ill-omen:kas", "strike:slash"]
```

The condition vocabulary is shared, so `party_struggling` means the same thing in every file. The flowchart's definition ("anyone is down, or the party's total HP is below half") becomes its default. PC tactics files use the same format and are written once per character, then reused in every scenario.

### Overlay

An overlay is a TOML patch applied to a combatant after the builder runs. It can set or adjust a number, add or remove an item, ability or resistance, and attach a tactics file. Overlays are how the build report's gaps get filled and how level-up variants are expressed.

## Per-Area Design Details

### Importing PCs from Foundry

The five provided exports confirm that Foundry saves PC choices and not results. The builder computes:

- **Proficiency bonuses:** level plus 2, 4, 6 or 8 for trained, expert, master and legendary, and 0 for untrained. Ranks come from the class item, the skill ranks, and class features that raise them (Weapon Legend, Expert Spellcaster, Juggernaut and similar).
- **Attribute modifiers:** from the ancestry, background, class and level boosts recorded under `build.attributes.boosts`.
- **AC, saves, attacks and damage:** from proficiency, attributes, armor and its runes (potency, resilient), weapon runes (potency, striking, property runes), and the item bonuses invested items give.
- **Foundry rule elements:** the builder reads a supported subset of the `rules` arrays on items: `FlatModifier`, `Resistance`, `Weakness`, `AdjustDegreeOfSuccess` (Juggernaut, Bravery, Inextinguishable) and `DamageDice`. Every unsupported rule element appears in the build report by item name, so nothing silently drops out.

Spells come with their mechanical data inside the export (damage formula, save, area, rank), so the builder reads them from there and does not need a separate spell database for v1.

The build report is printed by a `build` command and lists, for each PC: computed statistics, items and abilities that were understood, and anything skipped with the reason. The GM checks the computed numbers against the Foundry sheet once, then relies on them.

### Importing NPCs and hazards

NPC exports store final numbers (Luna's AC 36, Lilith's saves and resistances), so the builder reads them directly. Abilities whose description is text only (most custom ones, such as Challenge Fate or Sin Vulnerability) are matched to a content entry by name. An unmatched ability is listed in the build report.

### Rules coverage for v1

These are the rules the finale needs:

- Initiative from Perception, three actions and one reaction per turn, and the multiple attack penalty including agile.
- Four degrees of success, including natural 20 and natural 1 adjustments and abilities that shift degrees (Juggernaut, Bravery, Inextinguishable).
- Fortune and misfortune, including rolling twice, how the two cancel, and Untethered to Fate.
- Conditions: doomed, dying, wounded, frightened, sickened, slowed, stunned, confused, stupefied, blinded, quickened (haste), and persistent damage.
- Resistances, weaknesses and immunities, including exceptions (holy, cold iron) and Protean Anatomy's shifting resistance.
- Recharges (1d4 rounds), frequency limits (once per hour, once per fight, uses per day), and spell slots.
- Reactions and triggers: Attack of Opportunity, Challenge Fate, Vengeful Spite, Lash Out, Farabellus Flip, and Sin Vulnerability's loyalty triggers.
- Fast healing, healing, and temporary HP.
- Complex hazards: a routine of up to three actions, parts with their own HP or required successes, disable checks, and round-based escalation.

### Range bands

Each combatant has a band relative to each enemy: engaged, near or far. Moving one band takes a Stride or Fly action. Reach and speed decide who can close or escape in one action, so a flier with speed 180 can always reach far. Area effects compute how many party members they catch from the bands plus the scenario's bunching setting, which is the chance that a near ally is inside the area. The tactics condition `in_area:cone-40 >= 2` asks this model.

### Outcome buckets

The scenario defines its own buckets, so "good outcome" is a GM decision and not the tool's. The finale needs the four from the flowchart, and their exact definitions are Open Question 4. The proposed defaults are:

- **Party wiped:** every PC is dead, or the fight ends with no PC able to act.
- **Win with a death:** both bosses are defeated and at least one PC died.
- **Great finale:** both bosses are defeated, no PC died, and at least one PC was brought to dying or below a quarter of their HP at some point. The fight felt dangerous.
- **Flat fight:** both bosses are defeated and no PC ever fell below half HP.

A run that fits none of these is reported as "other" so that gaps in the definitions are visible.

### Reports

Each run produces an event log. The runner sums these logs into:

- outcome bucket percentages per variant, side by side
- rounds per phase (median and spread)
- HP over time for each PC and boss, averaged
- how often each ability and each tactics rule fired
- how often each dial changed the outcome, compared by turning it off for a batch

The CLI prints a summary table. The JSON file holds everything, so a later rendered report or the Dice Lab module can read it. Any single run can be replayed by its seed as a readable turn-by-turn log, which is how a surprising result gets checked.

### The Gate to the Abyss

The Gate is campaign design work, not code. It follows the existing Boss Fight + Complex Hazard build order and is designed alongside Efforts 1 to 4. Its stat block then becomes a hazard file the simulator loads. Lilith's Last Stand already references it ("none if the Gate to the Abyss was sealed"), so sealing the Gate is one of its disable outcomes. The simulator will show how often parties seal it in time.

## Proposed Effort Slicing

Each effort ships something that runs.

1. **Engine core.** Combatant type, seeded roller, d20 checks and degrees of success, Strikes, the multiple attack penalty, HP, initiative and the turn loop. NPC import. Test: Luna against fixed dummy PCs.
2. **PC builder.** Foundry PC import, stat computation, supported rule elements, the build report and overlays. Test: computed numbers match the three Foundry sheets.
3. **Content and conditions.** Primitive effects, the condition set, reactions and triggers, and named handlers for both bosses. Phases and carry-over.
4. **Tactics.** The TOML tactics format, the condition vocabulary, dials, range bands, and PC tactics files for Kas, Tavish and Arabesque.
5. **Hazards and clocks.** The complex hazard model and round-based escalation. The Gate is loaded once it is designed.
6. **Runner and report.** Monte Carlo runs, variants, outcome buckets, the CLI summary, the JSON output and seed replay.

## Known Data Decisions

These were settled in the session that produced this doc:

- The Foundry sheets are the source of truth for both bosses: Luna has 300 HP and Lilith 390, with every ability on their sheets. The GM flowchart's numbers (270 and 300) are out of date.
- Lilith's innate spells without a recorded frequency (Unfathomable Song, Divine Wrath, Confusion, Cursed Metamorphosis, Disintegrate, Translocate) can each be cast once per fight. Warp Mind keeps its three uses, and Dispel Magic is at will.
- Leftover effects on the exported sheets (Kas's slowed 1 and haste, Arabesque's frightened 1 and persistent damage) are ignored. Every run starts fresh.
- Tavish is simulated with the legacy alchemist features his sheet uses.
- Kas's shield is destroyed, so Shield Block is not modelled.

## Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D-1 | One multi-effort feature, `F.041`, not an arc | The `pf2e` engine has one consumer, the simulator; the efforts end in one deliverable. Reroute to an arc if a second PF2e tool appears |
| D-2 | Encounter files live under the campaign: `campaigns/<campaign>/encounters/<fight>/` | One campaign keeps its faction layer and its PF2e fights together; a `campaigns/hawk/` faction campaign moves in beside them later |
| D-3 | The Foundry exports and the battlemap are committed under `source/` | The players will not read the repo, and tracking the files keeps them from being lost |

## User-Facing Impact

- A new `pf2e` command group in `cmd/gm-toolkit`, beside `faction`. Exact command names are set at Plan.
- **Build** — prints the build report for each combatant: computed statistics, what was understood, and what was skipped with the reason.
- **Run** — runs a scenario's variants, prints the summary table and writes the JSON results file.
- **Replay** — plays one run again by its seed as a readable turn-by-turn log.

## State Storage

- **Per fight** — `campaigns/<campaign>/encounters/<fight>/`: the Foundry exports and battlemap under `source/`, the scenario file, the boss tactics files and the fight's overlays.
- **Per character** — PC tactics files and level-up overlays are reused in every scenario, so they sit at campaign level, not inside one fight. The exact layout is set at Plan.
- **Results** — the JSON results file. Where it is written, and whether it is tracked, is Open Question 7.
- No change to the faction manager's saves or the campaign registry.

## Replaces / Retires

- `cmd/scripts/hawk_finale_sim.py`, the one-off Python simulator of the finale, if Open Question 6 says so.

## Open Questions

1. **Abilities that do not fit the primitives.** Walk the Timelines, Sin Vulnerability, Last Stand and Protean Anatomy do not compose from the primitive effects. The project direction is data over per-ID Go handlers (`engine/ability-catalog.md`). Do these grow new primitives, or effect-keyed handlers driven by data, or do some stay bespoke for v1?
2. **Level-15 choices.** The level-15 variant needs each PC's level-up picks: attribute boosts, the general and skill picks, the class features gained at 15, and any feats or gear. Do you supply these, or do the players?
3. **Tavish's clan pistol and clan dagger.** Their granting feats are on his sheet but the weapons are not. Is the pistol plain, or does it have runes? Does it matter in this fight?
4. **Outcome buckets.** Are the proposed definitions of great finale, flat fight, win with a death and party wiped right, especially "great finale"? The existing script uses five other buckets (Cruise, Comfortable, Tense, Pyrrhic, Loss) plus a Threat and Triumph drama score; does any of that carry over?
5. **How well the PCs play.** Should PC tactics be close to optimal, or match how your players actually play, for example finding the Sin of Betrayal late? Should hero points be spent, and by what rule?
6. **The existing script.** `cmd/scripts/hawk_finale_sim.py` is in the repo. Should this tool replace it, and should the script be deleted once the finale scenario runs?
7. **Results files.** Where does the JSON results file go, and is it committed or gitignored like the faction manager's play-time state?
