# Encounter Lab — Arc Discovery

> **Status:** written 2026-11-19 for sign-off. The decisions here bind every initiative in the arc. This doc does not slice the arc into initiatives; that is Arc-Plan.
>
> **Inputs:** the [arc concept](A.004-encounter-lab-arc-concept.md), the [`F.041` discovery](../../discovery/F.041-pf2e-encounter-simulator-discovery.md) (D-1 to D-26), the [`F.041` overview plan](../../implementation/F.041-pf2e-encounter-simulator-overview-plan.md) (Decisions 1 to 12), the [rules scope](../../implementation/F.041-pf2e-encounter-simulator-rules-scope.md) and the seed files in `seeds/second_fight/`.

## In Plain Words

- **What** — the rules that four projects must all follow so they fit together: a simulator that plays a PF2e fight by the real rules (sim core), measures that say whether the fight is good (measurement), the way each creature chooses its turn plus the page the GM runs the fight from (tactics and flowcharts), and a tool that turns an ability's text into data the simulator can run (description conversion).
- **Why** — the simulator was planned as one feature, `F.041`. It turned out to be four different kinds of work, and its design doc made decisions for all four at once. Some still hold for everyone, some now belong to one project, and seven no longer hold. Each later project needs to know which is which before its own design starts.
- **How it fits** — the second arc-level doc of `A.004`, after the concept. Arc-Plan comes next and cuts the work into initiatives and an order. `F.041` stays, narrowed, as the first initiative; three of its eight efforts are built.

## Problem

Robert builds a boss from published monsters and wants to know how the fight will go before the table sees it. The work is a loop: run the fight many times, read what the tool says, change the encounter, run it again. When the fight measures well, a flowchart page for the table is built from the tactics that were tested.

`F.041` was designed to do all of that as one feature. Four things now stand in the way:

- **One design doc holds decisions for four projects.** D-1 to D-26 cover data formats, rules, tactics, scenarios, scoring and the CLI. The arc splits those across initiatives, and nothing says which decisions still bind whom.
- **The planned scoring has no evidence behind it.** D-17 copies the finale script's threat and triumph scheme. Robert's later patch (`seeds/second_fight/hawk_finale_sim-guardrails.patch`) stops printing that verdict and adds plain measures and a win-chance curve. `audit_measures.py` and `winprob_check.py` compare measures with win chance, but they print rates, record no verdict, and have fallen behind the patch. So the approach exists and no measure has passed a check.
- **The second fight shows what the numbers must catch.** Robert's critique of the Ythrel sim: it likely ignores that the party never makes a save and struggles to hit, and it leaned too long on one ability to carry the fight. The seed files bear this out in part. Party save success is reported for Ythrel only (`ythrel_sim.py`). Attack results are counted per fight and never reported. Damage dealt to the PCs is not recorded by source at all, so "one ability carries the fight" could not show up.
- **Sim core cannot play a real fight on its own.** After Effort 5 the engine has the rules, but who is fighting, how each creature chooses its turn and what the bosses' own abilities do were all planned for Efforts 6 to 8, which leave sim core. A basic `run` built on what remains would print numbers about two stat blocks trading Strikes.

The two flowchart pages made before this arc (The Fall of Luna, The Soft Collapse) are examples of the output, not a spec for it.

## Architecture Context

Pages read for this doc: `docs/architecture/pf2e/overview.md`, `engine.md` (the decider, the events, Key Decisions) and `data-files.md` (Key Decisions). Their Key Decisions bind every initiative in the arc:

- **The engine is headless** — decisions come in through a `Decider`, and events go out through one `Observe(Event)`. Events are plain values (`engine.md`).
- **One seeded roller per fight** — the same seed and files give the same events (`engine.md`).
- **Rules are modelled as written, and nothing is faked** (`engine.md`).
- **`FirstFitting` is a test behaviour, not tactics** — rule lists replace it behind the same seam (`engine.md`).
- **The simulator has its own data model; Foundry is an import source** (`overview.md`).
- **A creature is an import layer plus hand layers, merged on load** (`data-files.md`).
- **All code lives under `internal/pf2e/sim/`, and the commands are `gm-toolkit pf2e sim <command>`** (`overview.md`; overview plan Decisions 4 and 5).

The wiki describes the code as of Effort 3. Checked against source in this session: `cmd/gm-toolkit/pf2e/` holds only `show` and `import`, and `internal/pf2e/sim/engine/` holds one decider file. No `sim.md` page exists yet (D-22).

## Design Summary

The arc keeps the engine, the data model and the import as `F.041` designed them, and changes three things.

- **The decisions are sorted.** Each of D-1 to D-26 is marked as binding the arc, staying inside sim core, or changed (The `F.041` decisions).
- **Sim core's first `run` plays the real fight.** Sim core takes a small, hand-made piece of each thing a fight needs: a scenario file, rule lists, the bosses' abilities and the check that nothing is missing (ACD-5). It grows, and its numbers mean something from the first day.
- **One record of what a fight produced.** The engine's events are the only source. Sim core owns one per-fight record built from them, and the other initiatives read it (ACD-4).

```
              sim core (F.041, narrowed)
  import -> files -> load -> engine -> run -> fight record
                                                  |
               +----------------+-----------------+
               |                |                 |
          measurement    tactics, flowcharts   (replay)

  description conversion writes ability files that sim core loads
```

## The arc decisions

Tagged **carried in** (settled before this session, not re-opened) or **ratified here**.

| # | Decision | Rationale | Tag |
|---|----------|-----------|-----|
| ACD-1 | `A.004` is an arc, and `F.041`, narrowed, is its first initiative on the same branch with no history rewrite | The work is four kinds, each needing its own Discovery; reverses D-2 | Carried in |
| ACD-2 | It is a simulation, not a projection: rules coverage stays as complete as makes sense, and every remaining approximation is named in the output | Efforts 3 to 5 are not trimmed and no approximate model replaces a rule; D-18 with overview Decision 9 | Carried in |
| ACD-3 | A measure counts only after it is checked against evidence; none is invented | The threat and triumph scheme was invented and then dropped; the method of checking is the Measurement Discovery's | Carried in |
| ACD-4 | The engine's events are the one source of what happened. Sim core owns one per-fight record built from them; the other initiatives only read it, and ask sim core for anything missing | Measures, flowcharts and tactics comparison all read the same facts, so one count of each fact. The field list is set in sim core's planning. Rejected: the full field list fixed here (decided before measurement has evidence); no shared record (the same thing counted three ways) | Ratified here |
| ACD-5 | Sim core's basic `run` plays the real Fall of Luna fight. Sim core takes the scenario file (party, enemies, phases with carry-over, run count, seed), the rule-list format with a small condition set, hand-written rule lists and boss abilities, and `check` with D-8 | Numbers about a toy fight cannot be trusted or checked. Rejected: a fixed default chooser (damage share, rounds and knockdowns mean little); one phase only (Fall of Luna has two); Strikes only; waiting for conversion | Ratified here |
| ACD-6 | A fight is one scenario file that several initiatives extend | Sim core writes the first sections; variants, presets and compare are added later to the same file; amends D-13 | Ratified here |
| ACD-7 | Tactics are data in one rule-list format for bosses and PCs. Sim core owns the format and the first conditions; the tactics initiative owns position, dials, comparing candidate tactics and the flowchart | The flowchart comes from the tactics the fight was tested with; D-12 stands | Ratified here |
| ACD-8 | A boss is a standard stat block with added or changed abilities. Until conversion exists, an added ability is written by hand in the ability format; the hand-written Fall of Luna abilities are the answer key for the conversion trial | The import and the conversion earn their place; no fight waits on conversion | Ratified here |
| ACD-9 | The work is a loop: run, read, change, run again. Running two versions of a fight side by side is part of the arc | Insight comes from comparing an encounter before and after a change; who builds compare is Arc-Plan's call | Ratified here |
| ACD-10 | Every initiative runs its own full cycle of Discovery, Plan and Execution. 2026-11-23 is a hope, not a constraint on order or scope | The arc docs make a per-initiative Discovery lighter, never skipped | Carried in |

## The `F.041` decisions

Each decision is read as amended by the overview plan. "Binds the arc" means a second initiative must obey it.

| # | Decision, in short | Status | Note |
|---|--------------------|--------|------|
| D-1 | Own data model; a Foundry importer feeds it | Binds the arc | Conversion writes it and tactics name its abilities |
| D-2 | One feature, not an arc | Reversed | By ACD-1 |
| D-3 | Go, in gm-toolkit | Binds the arc | |
| D-4 | PC numbers computed from build choices | Sim core only | |
| D-5 | Mixed editions follow the sheet | Sim core only | |
| D-6 | The sim reads only its own files; import is a separate step | Binds the arc | An imported, hand-written or converted creature is the same kind of file |
| D-7 | Import layer plus hand layers, merged on load | Binds the arc | Conversion fills one; the GM's changes go in the other |
| D-8 | No run while an ability that can affect the fight lacks mechanics | Binds the arc | `check` and the refusal move into sim core (ACD-5) |
| D-9 | Claude converts sentences; the sim never calls Claude | Binds the arc | Keeps every run free and repeatable |
| D-10 | A PC's import layer holds build choices; load computes the numbers | Sim core only | |
| D-11 | Range bands behind one engine seam | Binds the arc | Until they land, every creature counts as engaged, printed as a named approximation |
| D-12 | Behaviour is an ordered rule list; first match wins | Binds the arc | Format ownership per ACD-7 |
| D-13 | One scenario file holds the whole fight | Amended | By ACD-6: one file, several owners |
| D-14 | Complex hazards are out of scope | Sim core only | |
| D-15 | A CLI runs fights; content is text files checked by `check` | Binds the arc | |
| D-16 | "Gate sealed" is a scenario dial | Sim core only | Which initiative gives the scenario file its dials is routed |
| D-17 | Scoring is a default scoring file | Withdrawn | Built on threat and triumph; the Measurement Discovery decides what replaces it (ACD-3) |
| D-18 | Rules as written; every approximation named | Binds the arc | With overview Decision 9; ACD-2 |
| D-19 | PCs and default scoring at campaign level; bosses in the fight's folder | Amended | The folders hold (overview Decision 6); the scoring file goes with D-17 |
| D-20 | Four acceptance gates; the script retires when they pass | Amended | Gates 1 to 3 stay with sim core; gate 4 and the script's deletion are routed |
| D-21 | Six efforts | Superseded | By overview Decision 8 and then by the arc plan |
| D-22 | A `pf2e/` wiki side with three subsystem pages | Amended | `overview.md`, `data-files.md` and `engine.md` exist; who writes `sim.md`, and whether it is one page, is routed |
| D-23 | Each effort writes its wiki page as its last step | Binds the arc | |
| D-24 | Wiki front door restructured by product | Sim core only | Done in Effort 1 |
| D-25 | Lilith's unmarked innate spells are once per fight | Sim core only | |
| D-26 | The importer ignores sheet state; every run starts at full HP | Sim core only | |

Overview plan Decisions 4 (package root), 5 (command group) and 6 (campaign folders) bind the arc. Decisions 1 to 3 and 7 to 12 order and scope the efforts; Arc-Plan re-cuts them.

## What sim core now holds

The effect of ACD-5 on `F.041`, for Arc-Plan to turn into efforts. Nothing here is sliced.

| Piece | Planned in | What sim core takes | What stays for a later initiative |
|-------|------------|---------------------|-----------------------------------|
| Engine, import, rules | Efforts 1 to 5 | All of it | — |
| Who is fighting | Effort 8 | The scenario file: party, enemies, phases with carry-over, run count, seed | Variants, presets, compare, `scenario create` |
| How a creature chooses | Effort 7 | The rule-list format, a small condition set, hand-written lists for the five Fall of Luna creatures | Position, dials, chance, comparing candidates, the flowchart |
| What the bosses do | Effort 6 | The kept boss abilities of the rules scope, written by hand in the hand layers | Conversion of description text, drafts and approval |
| Nothing missing | Effort 6 | `check` and the D-8 refusal | Conversion's approval states in `check` |
| What happened | Effort 8 | `run`, the fight record (ACD-4), raw facts, a replay log | Measures, tiers, verdicts |

The raw facts the concept names for the basic `run`: hit rates, save rates, damage share by source, knockdowns and rounds. The seed files show the gap they must close: hit rates per side, and damage to the PCs by source, are the two the scripts never reported.

## Open Questions — routed

| Question | Goes to |
|---|---|
| How the arc is cut into initiatives, and the order of tactics and conversion | Arc-Plan |
| Where the basic `run` and the pieces in What sim core now holds land among the `F.041` efforts | Arc-Plan |
| How the overview plan's Efforts 6 to 8 are cut down, and where each remaining part goes | Arc-Plan |
| Where gate 4 (the replay review) and the deletion of `cmd/scripts/hawk_finale_sim.py` land (D-20) | Arc-Plan |
| Which initiative builds compare (ACD-9) and gives the scenario file its dials (D-16) | Arc-Plan |
| Who writes `sim.md`, and whether it stays one page (D-22) | Arc-Plan |
| The narrowed `F.041` tracker row and write-up, and rows for the other initiatives | Arc-Plan |
| Which conditions make up sim core's "small condition set" | Sim core (the effort plan that builds rule lists) |
| The field list of the fight record | Sim core (the effort plan that builds `run`) |
| What a measure is checked against (table logs, win-chance replays, or both) and how many played fights are enough | Measurement Discovery |
| Whether measures are data a campaign owns, as D-17 and D-19 had it for scoring | Measurement Discovery |
| Whether the engine must copy a fight at a round end and play it on. The win-chance curve in the guardrails patch needs it, and `engine.Run` plays only from a `Setup` | Measurement Discovery, as a request to sim core under ACD-4 |
| What the tool emits for a flowchart: a finished page, or data a page is built from | Tactics and flowcharts Discovery |
| Where the Claude conversion runs: in the importer with a review command, or as a skill (the `F.041` open question) | Description conversion Discovery |

The 2026-11-23 table session is the first test of the tools, not evidence read beforehand. If Robert adds its logs to `seeds/second_fight/`, they are input for the Measurement Discovery.

## Out of Scope

- **A boss builder** — wanted eventually, far off.
- **A player playbook page** — the Kas page beside the Fall of Luna flowchart was a musing, not a planned output.
- **A starting state for the party** (arriving hurt) — D-26 stands; no fight in the arc needs it.
- **The full PF2e ruleset as data, and an encounter designer** — as in `F.041` (D-2's rationale still holds for these two).
- **A grid or map position model** (D-11), and **complex hazards** (D-14).

## Reconciliation Items

Doc drift found in this session. None blocks sign-off.

- **The arc concept** said the second fight's script was a copy and that the seed scripts check each measure. Corrected in `01c9503`.
- **The `F.041` discovery doc** is not edited (discovery docs are not updated after sign-off). This doc's table is the current reading of D-1 to D-26.
- **The `F.041` overview plan** still describes eight efforts, D-17 scoring and "no `run` exists yet, so no odds rest on the gap" under Gaps between efforts. Arc-Plan rewrites it.
- **`docs/architecture/pf2e/overview.md`** says the runner "arrives with Effort 8" and that the sim page comes from "Efforts 7 and 8". The effort that builds `run` corrects it.
- **`ythrel_sim.py`** still reports the great, flat and drama figures that the guardrails patch stopped printing for the finale. It is a seed, not maintained.
