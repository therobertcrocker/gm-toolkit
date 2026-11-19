# Encounter Lab — Arc Plan

> **Status:** written and signed off 2026-11-19. This doc cuts the arc into initiatives, sets their order and boundaries, and gives every open question an owner. It holds no design for any one initiative.
>
> **Inputs:** the [arc concept](A.004-encounter-lab-arc-concept.md), the [arc-discovery](A.004-encounter-lab-arc-discovery.md) (ACD-1 to ACD-10, and its reading of D-1 to D-26), the [`F.041` overview plan](../../implementation/F.041-pf2e-encounter-simulator-overview-plan.md) and the [rules scope](../../implementation/F.041-pf2e-encounter-simulator-rules-scope.md).

## In Plain Words

- **What** — the arc is four initiatives. Sim core (`F.041`) plays the real Fall of Luna fight and prints raw facts. Tactics and flowcharts (`A.004.2`) adds position, compare and the page the GM runs the fight from. Measurement (`A.004.3`) adds measures that were checked against evidence. Description conversion (`A.004.4`) turns an ability's text into data.
- **Why** — the arc-discovery set the rules that all four must follow, and left open who builds what and in what order. Without that cut, `F.041` still plans three efforts that no longer belong to it, and the other three have no boundary to start a Discovery from.
- **How it fits** — the third and last arc-level doc of `A.004`. After it, each initiative runs its own Discovery, Plan and Execution. `F.041` is three efforts in; it continues with the Effort 4 Plan session.

## Decisions Ratified in Arc-Plan

All are Robert's calls, made in this session.

| # | Decision | Rationale |
|---|----------|-----------|
| AP-1 | Four initiatives, in the order sim core, tactics and flowcharts, measurement; description conversion depends on sim core alone and runs last | Tactics brings position, so measures are checked on a fight without the "all engaged" approximation. Rejected: measurement second (each measure needs a second check when position lands); conversion second (no fight waits on it, ACD-8); a fifth initiative for compare |
| AP-2 | Sim core keeps eight efforts. Efforts 6, 7 and 8 keep their themes and shrink: abilities and `check`, rule lists, scenario and `run` | The hand-written abilities need the rules of Efforts 4 and 5, the rule lists name those abilities, and `run` needs both. Rejected: two efforts (one effort with two design surfaces); scenario and `run` first (a toy fight until rule lists land, against ACD-5) |
| AP-3 | Effort 6 writes by hand every kept item of the rules scope that the import does not already cover, for the two bosses and the three PCs | The rules scope keeps about 60 PC items beside about 30 boss items, and D-8 refuses a run while one lacks mechanics. Extends the arc-discovery, which named the boss abilities only. Rejected: each rules effort keeps its own test files as hand layers; PCs wait for conversion (the first `run` could not play the real party) |
| AP-4 | `F.041` gets one short Discovery before Effort 6, for the scenario file, the rule-list format and the fight record | The `F.041` discovery only sketched the three, and they refer to each other. The fight record is the arc's shared contract (ACD-4). Rejected: the effort plans carry the design; a Discovery for the fight record alone |
| AP-5 | Tactics and flowcharts builds compare, variants, presets and the scenario dials (ACD-9, D-16) | ACD-7 already gives tactics its dials and the comparing of candidate tactics; comparing two encounters uses the same parts, so compare is built once. Rejected: sim core; measurement; its own initiative |
| AP-6 | Gate 4 (the replay review) lands in tactics and flowcharts. When it passes, `cmd/scripts/hawk_finale_sim.py` is archived into `dev_journals/campaigns/initiatives/seeds/`, not deleted | Gate 4 cannot pass before position exists. `seeds/second_fight/ythrel_sim.py` imports the script, and Robert wants it kept for reference. Amends D-20 ("the script retires") and the old Effort 8 ("the script deleted"). Rejected: delete it; archive it when sim core ends |
| AP-7 | Sim core writes `docs/architecture/pf2e/sim.md`: Effort 7 starts it, Effort 8 extends it. It stays one page for now | D-23: each effort writes its page as its last step. A later initiative extends it, or splits off its own page in its own plan if the page outgrows one. Rejected: pages named now, before the later Discoveries have run |
| AP-8 | The new initiatives are `A.004.2`, `A.004.3` and `A.004.4`, in run order. `F.041` keeps its ID and is the arc's first initiative | The `A.001` pattern; the ID shows the arc. No history rewrite (ACD-1), so `F.041` is the one exception |
| AP-9 | `scenario create` leaves the arc for the Backlog, as `F.060` | A scenario is a hand-written text file that `check` validates; a wizard waits until writing one by hand hurts |
| AP-10 | The `F.041` overview plan is rewritten in this session, as its own commit after this doc | It still describes the old eight efforts and D-17 scoring. Rejected: the `F.041.E4` Plan session (a session on triggers would also carry the rewrite) |

## Initiative List

| Order | ID | Initiative | Type | Depends on | Ships |
|-------|----|------------|------|------------|-------|
| 1 | `F.041` | Sim core | `feature` | — | `run` plays the real Fall of Luna fight and prints raw facts |
| 2 | `A.004.2` | Tactics and flowcharts | `feature` | `F.041` | Position, compare, and the flowchart from the tested tactics; gate 4 |
| 3 | `A.004.3` | Measurement | `feature` | `F.041`; `A.004.2` (soft) | Measures, each checked against evidence |
| 4 | `A.004.4` | Description conversion | `feature` | `F.041` | An ability's text converted to data, with drafts and approval |

```
  F.041 sim core ---+---> A.004.2 tactics and flowcharts
                    |              :
                    |              : (soft)
                    |              v
                    +---> A.004.3 measurement
                    |
                    +---> A.004.4 description conversion
```

- **Hard edges** — every initiative needs sim core's merged `run`, fight record and file formats.
- **The soft edge** — measurement can play fights on sim core's rule lists alone. It runs after tactics so that its checks are made on a fight with position (AP-1). If the order must change, a measure checked before position is checked again after it.
- **Conversion** — depends on nothing but sim core. It runs last by choice (AP-1), and can move earlier without a change to any boundary.
- **Branches** — `F.041` stays on `feature/pf2e-encounter-simulator` and merges to `main` at its close. Each later initiative cuts its own branch from `main` after that merge.

## Per-Initiative Boundary

### `F.041` — Sim core

The engine, the import, the rules, and a first `run` that plays the real fight (ACD-5). The overview plan holds the effort detail; this is the boundary.

| Effort | Name | In | Ends with |
|--------|------|-----|-----------|
| 1 to 3 | Engine core, Foundry import, effect core | Built | Gates 1 and 3; gate 2 in part |
| 4 | Triggers and limits | Unchanged from the overview plan | Every Effort 4 row of the rules scope is tested |
| 5 | Spells and tables | Unchanged | Gate 2 |
| — | Formats Discovery (AP-4) | The scenario file, the rule-list format and the fight record, designed together | A signed-off discovery doc |
| 6 | Abilities and `check` | Every kept item the import does not cover, written by hand in the hand layers of all five creatures (AP-3); `ignore` marks for the dropped ones; `check` with the D-8 refusal | `check` passes for Fall of Luna |
| 7 | Rule lists | The rule-list format, a small condition set, the decider that replaces `FirstFitting`, hand-written lists for the five creatures; starts `sim.md` | The five creatures play their lists in test fights |
| 8 | Scenario and `run` | The scenario file (party, enemies, phases with carry-over, run count, seed); `run`; the fight record (ACD-4); raw facts; the named approximations in the report; `replay` and its log; extends `sim.md` | `run` plays Fall of Luna and prints raw facts |

- **Raw facts** — hit rates per side, save rates, damage share by source (damage to the PCs included), knockdowns and rounds.
- **Pushed to `A.004.2`** — position (range bands, reach, speed, spread), chance on a rule, dials and habits, variants, presets, `compare`, the flowchart, gate 4, the script's archive.
- **Pushed to `A.004.3`** — measures, tiers, verdicts, and whatever replaces D-17 scoring.
- **Pushed to `A.004.4`** — the tag converter (overview Decision 12), the conversion trial, drafts, approval, the approval cache, `review`, and the approval states in `check`.
- **Size** — Efforts 1 to 3 closed on 2026-11-16, 11-17 and 11-19: one effort in one to two calendar days, plan and close included. Five efforts and one short Discovery remain, so about six to nine calendar days. Effort 6 is the largest of the narrowed three (AP-3). Sim core alone runs past 2026-11-23.

### `A.004.2` — Tactics and flowcharts

- **In** — position behind the engine seam (D-11): range bands, reach, speed, area capture and the spread dial. Range on the Strike (overview plan, Findings, Strike range). More conditions for the rule lists, chance on a rule, dials and habits. The sin discovery curve as a scenario setting. Variants, presets, the scenario dials and `compare` (AP-5). Comparing candidate tactics. The flowchart for the table. Gate 4 and the script's archive (AP-6).
- **Extends, does not own** — the rule-list format and the scenario file are sim core's (ACD-6, ACD-7). A change to either is made with sim core's formats as the base, and the fight record is only read (ACD-4).
- **Pushed out** — measures (`A.004.3`). Compare shows raw facts until measures exist. `scenario create` (AP-9).

### `A.004.3` — Measurement

- **In** — what a measure is, how one is checked against evidence (ACD-3), the first set of checked measures, and how `run` and `compare` show them. The reading of the seed files in `seeds/second_fight/` and of any table logs.
- **Asks of sim core** — anything the fight record lacks, as a request under ACD-4. The first known request is copying a fight at a round end and playing it on, for the win-chance curve.
- **Pushed out** — tactics and position (`A.004.2`). No measure ships unchecked, so a measure with no evidence yet stays out.

### `A.004.4` — Description conversion

- **In** — the tag converter, the trial of both options (the Claude API in the importer with `review`, and a Claude Code skill), then the chosen one in full: drafts, approval, the approval cache, the three kinds of answer, and the approval states in `check`. The hand-written Fall of Luna abilities of Effort 6 are the answer key for the trial (ACD-8).
- **Fixed by the arc** — Claude converts sentences and the sim never calls Claude (D-9). A converted creature is the same kind of file as an imported or hand-written one (D-6, D-7).
- **Pushed out** — nothing in the arc waits on it.

## Open Question Routing

The seven questions the arc-discovery sent to Arc-Plan, answered:

| Question | Answer |
|---|---|
| How the arc is cut into initiatives, and the order of tactics and conversion | Four initiatives; tactics second, measurement third, conversion last (AP-1) |
| Where the basic `run` and the borrowed pieces land among the `F.041` efforts | Efforts 6, 7 and 8, narrowed, after one short Discovery (AP-2, AP-3, AP-4) |
| How the overview plan's Efforts 6 to 8 are cut down, and where each remaining part goes | Per-Initiative Boundary, `F.041`, the three "Pushed to" lines |
| Where gate 4 and the deletion of `cmd/scripts/hawk_finale_sim.py` land | Both in `A.004.2`; the script is archived, not deleted (AP-6) |
| Which initiative builds compare and gives the scenario file its dials | `A.004.2` (AP-5) |
| Who writes `sim.md`, and whether it stays one page | Sim core, Efforts 7 and 8; one page for now (AP-7) |
| The narrowed `F.041` tracker row and write-up, and rows for the other initiatives | Tracker Rows, below |

The seven it routed to an initiative, each owner confirmed:

| Question | Owner |
|---|---|
| Which conditions make up sim core's "small condition set" | `F.041`, the formats Discovery (was: the Effort 7 plan; moved by AP-4) |
| The field list of the fight record | `F.041`, the formats Discovery (was: the Effort 8 plan; moved by AP-4) |
| What a measure is checked against, and how many played fights are enough | `A.004.3` Discovery |
| Whether measures are data a campaign owns | `A.004.3` Discovery |
| Whether the engine must copy a fight at a round end and play it on | `A.004.3` Discovery, as a request to sim core under ACD-4 |
| What the tool emits for a flowchart: a finished page, or data a page is built from | `A.004.2` Discovery |
| Where the Claude conversion runs: in the importer with a review command, or as a skill | `A.004.4` Discovery, then its trial |

New in this session:

| Question | Owner |
|---|---|
| How `check` and `run` treat a kept item whose last rule is position (Sudden Charge, splash on other creatures) until `A.004.2` lands: a named approximation for each, or `ignore` | `F.041`, the Effort 6 Plan session, as a question to Robert per item |
| Whether the Ythrel fight is built in the tools as a second fight, and by which initiative | `A.004.3` Discovery, since its table logs are the first evidence |
| Where in `seeds/` the archived script sits, and how `ythrel_sim.py` finds it | `A.004.2` Plan |

## Cross-Arc Dependencies

Nothing outside the arc gates sim core. The later initiatives lean on these:

| Dependency | Gates | State |
|---|---|---|
| The Fall of Luna flowchart, `campaigns/hawk/encounters/fall-of-luna/reference/gm-flowchart.zip` | Gate 4 in `A.004.2` | Exists; zipped HTML, not text |
| Evidence for measures: the seed files in `seeds/second_fight/`, and the logs of the Ythrel table session of 2026-11-23 if Robert adds them | `A.004.3` | Seeds exist; the logs do not yet |
| The Claude API and `anthropic-sdk-go`, for the API option of the trial | `A.004.4` | Not in `go.mod` |
| Foundry exports of each later fight's creatures | Any fight after Fall of Luna | Robert exports them per fight |

`R.036` (one campaign resolve-and-hint helper) touches `cmd/gm-toolkit/pf2e/` and gates nothing.

## Out-of-Arc Deferrals

From the arc-discovery's Out of Scope, unchanged:

- **A boss builder** — wanted eventually, far off.
- **A player playbook page.**
- **A starting state for the party** (D-26).
- **The full PF2e ruleset as data, and an encounter designer** (D-2).
- **A grid or map position model** (D-11), and **complex hazards** (D-14).

Pushed out in this session:

- **`scenario create`** — the wizard for the scenario file; Backlog row `F.060` (AP-9).

The Backlog rows that the `F.041` overview plan adds at its close (ruleset as data, encounter designer, grid position, complex hazards, starting state) stay with that close.

## Tracker Rows

Written to `dev_journals/campaigns/planned-work.md` in this session.

- **`A.004`** — the row and write-up stay; the row's detail names the four IDs, and the write-up also links this doc.
- **`F.041`** — stays in Current Initiatives. The item becomes "PF2e sim core". The write-up is narrowed to the boundary above and marks the old Approach as superseded by the arc.
- **`A.004.2`** — Up Next, trigger `after F.041`.
- **`A.004.3`** — Backlog, trigger `after A.004.2`.
- **`A.004.4`** — Backlog, trigger `after F.041`.
- **`F.060`** — Backlog, Features: `scenario create`, trigger `when a hand-written scenario hurts`.

## What Comes Next

1. The `F.041` overview plan rewrite, in this session (AP-10).
2. The `F.041.E4` Plan session (triggers and limits), then Efforts 4 and 5.
3. The `F.041` formats Discovery (AP-4), then Efforts 6 to 8 and the `F.041` close.
4. `/gestalt:discovery A.004.2`.
