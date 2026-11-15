# PF2e Encounter Simulator — Discovery

> **Status:** signed off 2026-11-15.

## In Plain Words

- **What:** a tool that plays a Pathfinder 2e fight thousands of times and reports how it tends to go: win and loss odds, deaths, fight length, how tense it felt, and what did the damage. It starts as a rebuild of the one-off H.A.W.K. finale simulator, `cmd/scripts/hawk_finale_sim.py`.
- **Why:** the finale script grew piece by piece in one long session. Its numbers are typed in by hand, its bosses and players are written as code, and it cannot run any other fight. The bosses have changed since, and the script has not kept up. A structured, more accurate simulator makes the balance calls drawn from it more accurate too.
- **How it fits:** this is the first version of a larger PF2e tool. Later work adds the full PF2e ruleset as data and an encounter designer that uses it. F.041 builds the simulator so those can grow on top of it, but does not build them. The Fall of Luna is the first scenario and the test that proves v1 works.

## Problem

The finale script answers one question about one fight. It was built to check whether two custom bosses were balanced, and it grew a turn at a time, so its structure follows the order things were asked, not the shape of the game:

- **Stat blocks are typed in.** `PARTY` holds hand-copied HP, AC, saves and attack bonuses for Kas, Tavish and Arabesque at levels 14 and 15. The level-15 row is an estimate. Boss numbers are knobs (`luna_hp`, `avatar_hp`).
- **Behaviour is code.** Each boss turn (`luna_turn`, `avatar_turn`) and each PC turn (`kas_turn`, `tav_turn`, `ara_turn`) is a Python function. The GM flowchart is written as if-chains inside them, and every player habit is a knob that only one function reads.
- **Rules are partial.** Positions are probabilities, not places. Arabesque casts one spell. Warpwave's random effects, Arabesque's metamagic and defensive spells are left out. Many conditions are reduced to a simple stand-in.
- **The data has moved on.** The script and the flowchart both use 270 HP for Luna and 300 for the Avatar of Lilith. The current Foundry sheets have 300 and 390. The flowchart's odds (43% great finales) came from the script, so they rest on the old numbers.

The parts worth keeping are the ones that make it a balance tool and not only a dice roller:

- tuning knobs and named presets, and running every preset side by side (`--compare`)
- one fight replayed turn by turn from a seed (`--log --seed`)
- outcome buckets (Cruise, Comfortable, Tense, Pyrrhic, Loss) and the drama score (Threat and Triumph)
- per-PC death rates, what knocks PCs down, each source's share of the damage, rounds per phase
- what-if switches such as turning off all crits for one side

## Architecture Context

No architecture page covers PF2e. Every page describes the SWN faction engine. Pages read for this doc:

- `docs/architecture/architecture-overview.md`
- `docs/architecture/engine/ability-catalog.md`
- `docs/architecture/engine/dice.md`

No faction code is reused, but three of its standing decisions apply as principles:

- **Headless engine** (`architecture-overview.md`, design principles). The faction engine takes decisions through a collector and reports through an observer, and cannot tell who is on the other side. A simulator that picks moves from data and a later interactive mode or encounter designer would then drive the same engine.
- **Rules content lives in data the GM owns** (`architecture-overview.md`, design principles; `engine/ability-catalog.md`, Key Decisions). Abilities are data run by handlers keyed on the effect, not Go handlers keyed on one ability's ID.
- **Every roll takes a named route** (`engine/dice.md`, Key Decisions). One roller, seeded, so a run replays exactly.

PF2e gets its own side of the wiki (D-22), each page written by the effort that builds its subsystem (D-23):

```
docs/architecture/pf2e/
├── overview.md      spine: import → files → load → engine → runner; the seams
├── data-files.md    creature layers, file formats, import, conversion, approval
├── engine.md        rules, roller, events, the position and decision seams
└── sim.md           tactics, range bands, scenario, scoring, CLI
```

| Page | Written by |
|------|------------|
| `overview.md` | E1, extended by every effort |
| `data-files.md` | E1 (formats, layers), extended by E2 (import) and E3 (conversion) |
| `engine.md` | E1, extended by E4 (rules) |
| `sim.md` | E5 (tactics, position), extended by E6 (scenario, scoring, CLI) |

The front door, `architecture-overview.md`, opens with "The codebase splits along one seam: a headless engine, and whatever drives it", which is true only of the faction manager. E1 restructures it around the products (`campaigns`, `faction-manager`, `codex`), with the faction engine and interface as the faction manager's two sides and PF2e under `campaigns` (D-24).

## Design Summary

The shape in brief; the sections below give the detail, and the Decisions table the reasons.

- **F.041 is version 1 of the larger PF2e tool, not a cleanup of the script.** The script's features carry over; its structure does not.
- **The Fall of Luna is the first scenario and the acceptance test.** Its data comes from the current Foundry sheets, not the script's or the flowchart's numbers.
- **The full ruleset as data and the encounter designer are out of scope,** but v1 is shaped so both can build on it.
- **The simulator has its own data model, and Foundry is an import source.** The engine reads one GM-owned format. A Foundry importer converts actor exports into it. Custom abilities, hazards and, later, the designer's output use the same format.
- **It is Go, inside gm-toolkit.** The Python script retires once the finale scenario runs in the new tool.
- **Each creature is layers.** An import layer the importer owns, plus hand layers for logic, fixes and variants, merged on load. PCs are stored as build choices and computed at load; NPCs as numbers.
- **Claude converts ability wording, and the GM approves it.** Structured data and inline tags convert in Go; the sim itself never calls Claude.
- **The fight is data.** Tactics are priority rule lists, position is range bands behind a seam, and a scenario file holds phases, variants, presets and scoring.
- **Nothing is faked.** v1 models the finale's rules as written, and the report names every remaining approximation.

### What the Foundry exports hold

The five actor exports (pf2e system 8.6.1, Foundry 14) hold two different kinds of data:

| Export | Holds | Missing |
|--------|-------|---------|
| PCs | Choices: ancestry, class, feats, attribute boosts, gear with runes; rule elements on items (`FlatModifier`, `AdjustDegreeOfSuccess`, `Resistance`, `DamageDice` and others) | Final numbers. Only current HP is stored; max HP, AC, saves and attack bonuses are computed by Foundry when the sheet opens |
| Bosses | Final numbers (Luna 300 HP, AC 36; Lilith 390 HP, AC 39), Strikes, spells, and the full mechanics of each custom ability in its description | Rule elements for custom abilities. Their mechanics are in the description, not in machine-run rules |

So the custom boss abilities need the simulator's own format whatever the import does.

#### Custom ability descriptions

A custom ability's description holds its mechanics in two forms:

- **Tags** — Foundry's inline markup, exact and machine-readable: `@Damage[15d6[mental]]`, `@Check[will|dc:39]`, `@Template[cone|distance:40]`, condition links such as `@UUID[…]{Slowed 2}`, and recharge rolls such as `[[/gmr 1d4 #Recharge Destiny Spores]]`.
- **Sentences** — stock PF2e wording that ties the tags together: which degree of success gets which condition, triggers, frequencies, "makes two slash Strikes and one stab Strike".

An audit of Abyssal Luna's six combat abilities found nothing absent. Destiny Spores is almost all tags, with one sentence for the degrees ("slowed 1 for 1 round on a failure, slowed 2 on a critical failure"). Walk the Timelines is all sentences: the text is complete, but the mechanic is unusual. Lilith's sheet records uses only for Warp Mind (3) and marks Dispel Magic "At Will"; her other six innate spells (Translocate, Disintegrate, Cursed Metamorphosis, Confusion, Unfathomable Song, Divine Wrath) carry no frequency on the sheet, so PF2e's rule applies: each is once per day, which is once per fight (D-25).

### Import as a separate step

```
Foundry JSON ──import──► simulator files ──► sim
                               ▲
                 hand edits, homebrew, designer output (later)
```

- The importer runs when a Foundry sheet changes, not on every sim run.
- It writes the simulator's own files: PC build choices (see PC import), boss numbers as stored, and custom abilities converted from their descriptions. Tags convert exactly in Go; sentences are converted by Claude (D-9).
- It ignores the sheet's current HP, conditions and temporary effects, so every run starts fresh at full HP (D-26).
- Hand layers cover what the import does not: wording the importer cannot read, data that is truly absent from the sheet, homebrew, fixes and variants.
- The import report lists what was converted, and what was skipped with the reason.
- Each creature is made of layers. The importer owns the import layer and rewrites it freely; nobody edits it by hand. Hand work goes in a hand layer beside it: custom ability logic, fixes, homebrew and level-up variants. The sim merges the layers when it loads, and a `show` command prints the merged result.
- A creature built from scratch has a hand layer and no import layer.

```
abyssal-luna.import.toml   (importer rewrites freely)
+ abyssal-luna.toml        (hand layer: logic, fixes, variants)
= Luna as the sim sees her
```

File names here show the idea; the real layout is set at Plan.

### Converting description sentences

Claude converts the sentences in a custom ability's description into the simulator's ability format (D-9). Two ways to run it are both kept open, to be tried hands-on before one is chosen (see Open Questions). These parts are the same in both:

- **One ability schema.** Claude's output must match the schema the sim loads, and the loader validates it again. Output that does not validate is a failed draft, not a guess.
- **Three kinds of answer.** A draft holds mechanics; or says "unsupported", with the reason, when the mechanic does not fit the schema (the ability then falls under D-8); or says "no combat effect" for a passive such as a lore or crafting feat. An approved "no combat effect" satisfies D-8.
- **Drafts are staged, then approved.** A draft is pending until the GM approves it. A pending or rejected draft counts as "no mechanics", so D-8 stops the run. Only approved drafts enter the import layer.
- **Approval is cached by text.** Each approved conversion is stored under a hash of the description text and the schema version, and committed. Re-import asks again only for abilities whose text changed.
- **The sim never calls Claude.** Runs stay deterministic and free.

#### Option 1 — Claude API in the importer, with a review command

```
import ──► Go: structured data + tags ──────────────────► import layer
       └─► Claude API: sentences ──► staged drafts (pending)
                                          │
review ◄──────────────────────────────────┘
  description | drafted mechanics, side by side
  [a]pprove  [r]eject  [e]dit in $EDITOR  [n]ote + regenerate
                                          │
                             approved ──► import layer + cache
```

- The Go importer calls Claude through `anthropic-sdk-go`, with structured outputs set to the ability schema, once per ability whose text changed.
- A `review` command walks the pending drafts with `huh` prompts. "Note + regenerate" sends the GM's note back with the draft and asks again.
- Import needs an API key; review and the sim do not.
- Strengths: one tool writes the import layer; the conversion is code, so it can be tested; repeatable.

#### Option 2 — A Claude Code skill

- A skill reads the pending abilities' descriptions, drafts each conversion against the same schema, discusses it with the GM in the session, and writes the approved result to the import layer and cache.
- Strengths: free-form back-and-forth, including asking why a sentence was read a certain way; no API key in the tool.
- Costs: import is two tools, not one; the conversion logic lives in a prompt, not in tested code; results depend on the session.

### PC import

The importer writes each PC's build choices into the import layer (D-10). When a scenario loads, the tool merges the PC's layers and computes the numbers from those choices, once, before any fight. The computation covers:

- **Core math** — proficiency, attribute modifiers from the recorded boosts, max HP, AC, saves, Perception, attack and damage, class and spell DCs, armor and weapon runes, item bonuses.
- **Common rule elements** — a supported subset of the `rules` arrays on items, starting with `FlatModifier`, `AdjustDegreeOfSuccess`, `Resistance` and `DamageDice`.
- **Import report** — lists, per PC, the computed numbers and every item or rule element skipped, with the reason. Nothing drops out silently.
- **Hand layers** — corrections, missing items and level-up variants on top of the import layer (D-7). A level-up variant is the new choices (boosts, feats, rune upgrades); the load-time computation turns them into numbers.

The GM checks the computed numbers against the Foundry sheet once per character. v1 covers the three classes the finale needs: fighter, sorcerer and legacy alchemist.

The rules reference is the Archives of Nethys [Player's Guide](https://2e.aonprd.com/PlayersGuide.aspx), which serves the Remaster by default and keeps the legacy Core Rulebook pages beside it. The H.A.W.K. sheets mix the two: all three classes, and Kas's and Tavish's ancestries, are Core Rulebook items; Arabesque's ancestry and some feats are Player Core; Tavish has four homebrew feats.

### Rules coverage

v1 models the rules the finale needs, each as written and each tested (D-18). Rules a later fight needs are named by the import report and `check`, never faked. Items marked † are skipped or approximated by the script today.

| Area | Rules |
|------|-------|
| Turn | Initiative, three actions and a reaction, multiple attack penalty (agile), sustain †, quickened (haste) |
| Checks | Four degrees of success, natural 20 and 1, bonus-type stacking, basic saves, fortune and misfortune † (Ill Omen, Challenge Fate, True Target, Untethered to Fate) |
| Damage | Damage types; resistance, weakness and immunity with exceptions (holy); critical hits; persistent damage †; splash; Protean Anatomy † (the script uses a flat resist 12); fast healing; temporary HP |
| Down | Dying, wounded, doomed, recovery checks, hero points, Inextinguishable, Too Angry to Die † |
| Conditions | Frightened, sickened, slowed, stunned, confused, stupefied, blinded, prone, clumsy, enfeebled, off-guard † |
| Reactions | Challenge Fate, Vengeful Spite, Lash Out, Farabellus Flip, Attack of Opportunity †, Sin Vulnerability's loyalty triggers |
| Spells | Slots, areas, durations, counteract † (Dispel Magic), and Arabesque's real spell list † (the script casts only vampiric exsanguination) |
| Limits | Recharge 1d4, frequency (once per hour or day becomes once per fight), uses per fight |
| Random tables | Warpwave † (skipped today), Unfathomable Song's effects |

When the party works out the Sin of Betrayal is player knowledge, not a rule. It stays as the script models it, a per-round discovery curve, as a scenario setting that tactics read (`knows:sin`).

A rules reading that changes when or how often something fires is Robert's call, raised as a question at Plan or Execution, not settled as a micro-decision.

#### Named approximations

Some modelling choices remain approximations by design. The run report lists each one, as the script's CAVEATS block does:

- Position as range bands, and party spacing as the spread dial (D-11)
- Behaviour as tactics rules, including any chance on a rule (D-12)
- The sin discovery curve
- Abilities marked `ignore`

### Position

Position uses range bands (D-11):

- **Bands** — each pair of creatures is *engaged*, *near* (within one Stride) or *far*. Reach and speed decide who can close or escape in one action; Luna's 180-ft fly speed always reaches *far*.
- **Areas** — an area effect catches a share of the party set by the bands and a per-scenario **spread** dial: how tightly the party stands. A sweep over the dial answers "what if they spread out?".
- **The seam** — the engine asks position questions (who is engaged with whom, how many targets an area catches, can this creature reach that one) through one interface. Range bands are its first implementation; a grid, with or without a map, can replace it later as its own initiative.

The finale's position questions all map onto bands: "2+ PCs in a 40-ft cone", "PCs bunched in a 20-ft burst", "Arabesque or Tavish exposed", "Kas has her pinned down", "fly out of Kas's reach", the stinger's 20-ft reach, and the Wail lingering for a round.

### Tactics

Each creature's behaviour is a tactics file: an ordered list of rules (D-12).

- **Rule** — conditions plus an action sequence. On its turn a creature takes the first rule whose conditions all hold.
- **Conditions** — a shared vocabulary, so a term means the same thing in every file: `struggling` (the flowchart's "anyone is down, or the party's total HP is below half"), `ready:destiny-spores`, `in_area:cone-40 >= 2`, `exposed:arabesque`, `dial:wail_mercy`.
- **Chance** — a rule can carry a chance to fire, for a GM who is not fully predictable.
- **Dials and habits** — GM dials ("dispel only when they struggle") are conditions on named settings. Player habits (the script's `kas_simple`, `late_heal`, `waste_p`) are a different PC tactics file or a setting.
- **The seam** — the engine asks one question, "what does this creature do now?", through one interface. The rule list is its first implementation; another, such as utility scoring for generic monsters, can be added later.

A sketch of Luna's flowchart as rules (the format is set at Plan):

```toml
[[rule]]
when = ["ready:destiny-spores", "in_area:cone-40 >= 2"]
do   = ["destiny-spores", "fly:away-from:kas"]

[[rule]]
when = ["dial:luna_picks_on_comfortable", "not struggling", "exposed:arabesque|tavish"]
do   = ["fly:engage", "abyssal-frenzy"]

[[rule]]
when = ["pinned_by:kas", "uses_left:walk-the-timelines"]
do   = ["walk-the-timelines", "fly:away-from:kas"]

[[rule]]
do   = ["ill-omen:kas", "strike:slash"]
```

### Scenario

A scenario file describes one fight (D-13). Each piece replaces a part of the script:

| Script today | Scenario piece |
|--------------|----------------|
| Luna, the Transformation, then the Avatar (`fight()`, `transition()`) | **Phases** — who is in each, when it ends, what carries over (HP, doomed, spent slots and items) |
| The Transformation's free round and 75 HP checkpoint | **Between-phase steps** — a free round, a conditional heal |
| `--level 14` / `--level 15` | **Variants** — a named stack of layers (D-7, D-10) |
| `RECOMMENDED`, `casual`, `defensive` | **Presets** — named bundles of dials and habits |
| `--compare` | **Compare** — a grid of presets by variants |
| `-n 5000`, `--seed` | Run count and seed |

A sketch (the format is set at Plan):

```toml
[scenario]
party  = ["kas", "tavish", "arabesque"]
spread = 0.5

[[variant]]
id = "level-14"

[[variant]]
id     = "level-15"
layers = ["level-15"]

[[phase]]
id        = "luna"
enemies   = ["abyssal-luna"]
ends_when = "defeated:abyssal-luna"

[[phase]]
id         = "transformation"
free_round = true
heal       = { all = 75, when = "struggling" }

[[phase]]
id      = "lilith"
enemies = ["avatar-of-lilith"]
carry   = ["hp", "doomed", "spent"]

[preset.recommended]
dials = { dispel = "struggling", wail_mercy = true }
```

Scenario dials also cover rules that depend on things v1 does not model. "Gate sealed" (default unsealed) decides Last Stand's 40 HP heal, since the Gate to the Abyss is not simulated (D-14, D-16).

### Scoring

Each fight is scored from the events the engine reports (D-17). A default scoring file carries the script's three layers, with every threshold editable:

```
Bucket   Cruise · Comfortable · Tense · Pyrrhic · Loss    (PC HP low point, deaths)
Drama    √(Threat × Triumph), 0–10
           Threat:  party HP low point, knockdowns, two PCs down at once
           Triumph: 60+ damage hits, rescues, a 35%+ round
Verdict  Great (win, no deaths, both ≥ 4) · Flat (either < 2)
```

A scenario uses the default, adjusts its thresholds, or adds its own events. The finale adds "the party finds the Sin of Betrayal and lands a holy hit" as a Triumph, as the script does. The flowchart's summary set (great finale, flat fight, win with a death, party wiped) is a view over the same results.

The raw report lines stay whatever the scoring says: deaths per PC, what knocks PCs down, each source's share of the damage, rounds per phase, and stalemates (a phase that hits the round limit).

## User-Facing Impact

A new command group in `cmd/gm-toolkit`, beside `campaign`, `spatial` and `faction` (D-15). Fights are named within the active campaign. Command names are set at Plan.

| Command | Does | Replaces in the script |
|---------|------|------------------------|
| `scenario create` | Scaffolds a fight: picks party and enemies from existing creatures, lays out phases and variants, writes the folder and a scenario file. A `huh` wizard when run bare, flags otherwise, like `campaign create` | — |
| `import <fight>` | Converts Foundry exports into import layers (D-6) | Hand-typed `PARTY` and boss knobs |
| `review <fight>` | Approves or rejects pending ability drafts (Option 1 under Converting description sentences) | — |
| `show <creature>` | Prints the merged, computed stat block (D-10) | — |
| `check <fight>` | Validates every file and lists abilities without mechanics (D-8), without running | — |
| `run <fight>` | Runs a preset and variant: run count, seed, summary report | `python3 hawk_finale_sim.py --preset …` |
| `compare <fight>` | Runs a grid of presets by variants side by side | `--compare` |
| `replay <fight> --seed <n>` | Plays one fight turn by turn | `--log --seed` |

Scenario content (tactics, phases, carry-over, dials) is written in text files; `check` catches mistakes before a run. An interactive builder that edits tactics or tunes numbers is the encounter designer's job.

Output is plain CLI text with ANSI colour, as in the other Cobra commands; the faction TUI is not involved.

## Proposed Effort Slicing

Each effort ends with something that runs (D-21):

```
E1 engine core ──► E2 Foundry import ──► E3 description conversion
      │                                          │
      └──────► E4 rules coverage ◄───────────────┘
                     │
               E5 tactics + position ──► E6 scenario, runner, scoring, CLI
```

| Effort | Builds | Ends with |
|--------|--------|-----------|
| E1 — Engine core | Stat block, seeded roller, checks and degrees, Strikes, multiple attack penalty, damage and resistances, HP and dying, initiative, turn loop, event reporting; creature and ability file formats, layer merge, `show` | A hand-written Luna fights dummy PCs, deterministically (gate 3) |
| E2 — Foundry import | NPC import (numbers and tags), PC import as choices (D-10), load-time PC stat computation, import report | Stat parity with the three sheets (gate 1) |
| E3 — Description conversion | Thin versions of both conversion options for a hands-on trial, then the chosen one in full: staged drafts, approval, cache, the three kinds of answer, D-8 in `check` | Luna's and Lilith's abilities converted and approved |
| E4 — Rules coverage | The rest of the Rules coverage table: conditions, reactions and triggers, fortune and misfortune, spells and counteract, persistent damage, Protean Anatomy, random tables | Every rule tested (gate 2) |
| E5 — Tactics and position | Rule-list tactics, the condition vocabulary, dials, range bands and spread, PC and boss tactics files | Luna and Lilith play the flowchart |
| E6 — Scenario and runner | Scenario file, phases and carry-over, variants, presets, scoring, `run`, `compare`, `replay`, `scenario create` | Replay review (gate 4); the script is deleted |

- **E1 includes the file formats**, so even the first fight is data, not Go.
- **E3 targets the schema E1 sets up.** E4 adds primitives, and each one extends the schema new drafts can use.
- **E5 follows E4** because tactics conditions such as `ready:` and `in_area:` query the rules underneath.

## Acceptance

v1 is done when four gates pass (D-20):

1. **Stat parity** — the computed level-14 stats for Kas, Tavish and Arabesque (HP, AC, saves, Perception, attack bonuses, class and spell DCs) match their Foundry sheets exactly. The exports do not hold these numbers, so Robert reads them off the sheets once as a test fixture. The script's `PARTY[14]` row, marked "from the sheets", is a starting point if still current.
2. **Every rule tested** — each row of the Rules coverage table has tests.
3. **Deterministic** — the same seed and files give identical results.
4. **Replay review** — Robert reads a set of `replay` logs and confirms that Luna and the Avatar of Lilith play the GM flowchart as written.

When all four pass, `cmd/scripts/hawk_finale_sim.py` is deleted. Its odds are not a target: the bosses have changed and the rules are now modelled as written, so the new numbers are expected to differ.

## State Storage

Files reused by every fight sit at campaign level; files for one fight sit in its folder (D-19). Folder names are set at Plan.

```
campaigns/hawk/
├── pf2e/                        reused by every fight
│   ├── party/kas/               import layer, hand layer, level-15 layer, tactics
│   ├── party/tavish/
│   ├── party/arabesque/
│   └── scoring.toml             the default scoring (D-17)
└── encounters/fall-of-luna/     one fight
    ├── source/                  Foundry exports and battlemap
    ├── reference/               GM flowchart, build order, quick numbers
    ├── scenario.toml
    ├── abyssal-luna/            import layer, hand layer, tactics
    └── avatar-of-lilith/
```

- **PCs at campaign level, bosses in the fight.** A boss who recurs can move up to campaign level later.
- **The campaign owns its scoring.** The binary embeds no default. A campaign without a scoring file gets one copied from another campaign, the way `campaign add-rules` copies rulebook files.
- **The approval cache** for converted descriptions (see Converting description sentences) is committed next to the import layers it serves.
- No change to the faction manager's state, saves or the campaign registry.

## Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D-1 | Own data model; a Foundry importer feeds it | Custom boss abilities have no machine-readable rules, so an own format is needed anyway; the engine stays free of Foundry's internal schema; the designer will write creatures, not only read them. Rejected: Foundry's schema as the model (ties the engine to Foundry's rule-element language and its version changes); hand-written stat blocks only (numbers drift from the sheets, as the script's did) |
| D-2 | Stays one feature, not an arc | The ruleset-as-data layer and the encounter designer are not scoped yet; an arc now would plan unscoped work. They get their own Backlog rows when closer |
| D-3 | Go, in gm-toolkit | Fits the project and its conventions; the later PF2e tools live here too |
| D-4 | PC numbers are computed from the build's choices by the core PF2e math and a supported set of rule elements; a report lists what is skipped, and hand layers fill the gaps. D-10 later placed the computation at load time, not in the importer | Level-up variants become choices the rules turn into numbers, not hand estimates; the later ruleset-as-data layer replaces pieces of this math instead of all of it. Rejected: a full rule-element engine (rebuilds much of Foundry's sheet); typed-in numbers (go stale, as the script's did) |
| D-5 | Mixed editions follow the sheet: each item works as written in its own source | The sim plays the party as it is run at the table; Tavish stays a legacy alchemist. The shared core math (proficiency, AC, saves, bonus stacking) is written once. Rejected: Remaster everywhere (changes the characters); legacy everywhere (against AoN, Foundry and the bosses' own data) |
| D-6 | The sim reads only its own stand-alone files; importing from Foundry is a separate step that writes them | Imported, hand-written and designer-made creatures are the same kind of file, and the sim cannot tell them apart. Custom abilities carry their own numbers and logic; the importer converts them from the Foundry description. Rejected: reading Foundry exports at run time, with ability numbers pulled from the tags on every run |
| D-7 | A creature is an import layer the importer owns plus a hand layer, merged on load | Re-import stays one safe step after any Foundry tweak; level-up variants and homebrew sit in hand layers. Rejected: the file is yours after import and re-import only reports differences (every tweak is a hand merge); import once and overwrite on demand (forcing discards hand-written ability logic) |
| D-8 | The sim refuses to run a scenario while any ability that can affect the fight lacks mechanics, unless a hand layer marks it `ignore`. "Can affect the fight" means every passive and reaction, and every active ability a tactics file uses. A conversion can answer "no combat effect", approved like any draft | A boss missing an ability skews the odds without any sign; the choice to skip one (Prophetic Wings, a 10-minute activity) is written down. Refined during discovery: applied to every item, the rule would need hundreds of `ignore` entries for the PCs' skill feats, lores and unused spells; actives no tactics file uses are listed as "unused" in the import report instead. Rejected: run and list the gaps in the report (easy to miss); every item, `ignore` per item (hundreds of entries); strict for NPCs only (a boss's actives are all in its tactics file anyway) |
| D-9 | Description sentences are converted by Claude, with no Go pattern parser beside it; tags and structured data stay in Go | A Go parser cannot read PF2e wording well, and one path for all sentences avoids two converters that disagree. Rejected: tags only (hand layers repeat the description and drift); patterns first with Claude for the rest (two paths) |
| D-10 | A PC's import layer holds the build choices; stats are computed once when a scenario loads, before any fight | A level-up layer states what changed and the numbers follow; the sim, `show` and the import report share one computation. NPCs stay as numbers, as Foundry stores them and as PF2e builds monsters. Rejected: computed numbers in the import layer (level-ups go back to typed estimates or a second import run) |
| D-11 | Position is modelled with range bands, behind one engine seam so another model can replace it | Answers every position question the finale's flowchart asks; party spacing becomes a visible dial instead of a guess hidden in movement AI; no map files. Rejected for v1: an open grid (movement and spacing AI for every creature); a grid with a hand-written map (B's cost plus a map format); probability knobs (the script's guesses) |
| D-12 | Behaviour is a priority rule list in data: ordered rules, each with conditions and an action sequence, first match wins | It is the GM flowchart as written, so the sim tests the flowchart; PCs use the same format; dials and habits are conditions or settings. Rejected for v1: utility scoring (unreadable, tuned by guesswork); behaviour trees (a heavier format for what is today a flat list) |
| D-13 | A scenario file holds the whole fight: party, phases with carry-over, between-phase steps, variants, presets, run count and seed | Every structural piece of the script's `fight()` and its flags becomes data, so a new fight needs no code |
| D-14 | Complex hazards are out of scope for v1 | The finale runs without the Gate to the Abyss as a hazard; the hazard model can come with the encounter that needs it |
| D-15 | A CLI runs fights and scaffolds scenarios; scenario content is written in text files and checked by a `check` command | Follows `campaign create`'s pattern (a `huh` wizard when run bare, flags otherwise) and the active-campaign registry. Rejected for v1: an interactive builder that edits tactics and tunes numbers (that is the encounter designer, with its own initiative) |
| D-16 | "Gate sealed" is a scenario dial, default unsealed; Last Stand stays available and a dial can turn it off | Last Stand's heal depends on the Gate, which v1 does not model (D-14); the dial keeps the sealed case testable, and the off switch keeps the script's `last_stand` comparison. Rejected: always heal 40 (sealed case untestable); a chance per run (a guess) |
| D-17 | Scoring is data: a default scoring file a scenario uses, adjusts or extends with its own events | Keeps everything the script measures and gives new fights a working score; fight-specific events such as "sin found + holy hit" stay possible. Rejected: scoring fixed in Go (drops fight-specific events; thresholds need code changes); raw measures only (drops the verdicts used for real decisions) |
| D-18 | v1 covers the rules the finale needs, each modelled as written and tested; nothing is faked, and every remaining approximation is named in the report | Accuracy is the goal, so the script's stand-ins (flat resist 12 for Protean Anatomy, one spell for Arabesque, skipped Warpwave) go; rules a later fight needs are found by `check`, not guessed now. Rejected: the finale plus the full conditions list and common basic actions (rules tested against no real fight) |
| D-19 | PCs and the default scoring live at campaign level; bosses and the scenario live inside their fight's folder | Kas is in every fight, Luna in one; a recurring boss can move up later. The campaign owns its scoring data and the binary embeds none, as for the faction rulebook |
| D-20 | v1 is accepted on four gates: stat parity with the Foundry sheets, every rule tested, deterministic runs, and Robert's replay review; the script retires when they pass | Each gate can fail. Rejected as a gate: matching the script's odds (the bosses changed and the rules are now real, so a gap is expected) |
| D-21 | Six efforts, sliced so each ends with something that runs: engine core, Foundry import, description conversion, rules coverage, tactics and position, scenario and runner | Each effort is one design surface with a review at its boundary; conversion comes before rules coverage so both conversion options can be tried early; tactics come after the rules they query |
| D-22 | A new `docs/architecture/pf2e/` side: a spine (`overview.md`) and three subsystem pages (`data-files.md`, `engine.md`, `sim.md`) | Mirrors the wiki's spine-plus-subsystem shape and its one-page-per-subsystem routing rule. Rejected: two pages, or one (each page would span several subsystems) |
| D-23 | Each effort writes or extends its subsystem's page as its last step | Each page is checked against code written in the same session, and each effort review sees its page. Rejected: all pages at initiative close (written from memory of six efforts) |
| D-24 | The architecture front door is restructured around the products as part of F.041, in E1 | Adding PF2e breaks the front door's "the codebase is one engine with two sides" framing; reorganising by product (campaigns, faction-manager, codex) fixes the framing for every later product. Rejected: a minimal PF2e section with the old framing reworded |
| D-25 | Lilith's six innate spells with no recorded frequency are each once per fight | PF2e's rule makes an unmarked creature innate spell once per day, and once per day becomes once per fight (D-18). Rejected: at will (a boss buff the sheet does not state); a limit set per spell |
| D-26 | The importer ignores the sheet's current HP, conditions and temporary effects; every run starts fresh at full HP | The exports were saved mid-play (Tavish at 90 HP; Haste, Slowed, Frightened and persistent damage on Kas and Arabesque), so imported state would tie results to when the export was saved. A starting-state scenario setting can come with the fight that needs it |

## Open Questions

- **Where the Claude conversion runs: importer with review, or skill.** Both options are described under Converting description sentences. Robert tries both hands-on before choosing, so Plan schedules a thin version of each early, ahead of the full importer. Deferred because the choice depends on how each feels in use, which only a working version shows.
## Out of Scope

- Complex hazards, including the Gate to the Abyss (D-14).
- The full PF2e ruleset as data (planned).
- An encounter designer (planned), including any interactive builder that edits tactics or tunes numbers.
