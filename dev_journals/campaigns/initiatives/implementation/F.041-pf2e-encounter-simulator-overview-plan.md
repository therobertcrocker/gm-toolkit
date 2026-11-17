# PF2e Encounter Simulator (F.041) — Implementation Plan (Overview)

## In Plain Words

- **What:** the overview plan for the PF2e encounter simulator: a Go tool that plays a Pathfinder 2e fight thousands of times and reports how it tends to go. It fixes the order of the six efforts, the package and command names, the campaign folder, and what each effort hands to the next. Each effort gets its own Plan session, so this file holds no code spec.
- **Why:** the discovery designed the tool and its six efforts. This session re-read the code and the Fall of Luna data, settled the names the discovery left to Plan, and found one clash in the effort order: the description conversion came before the rules it needs. The fix is to run rules coverage before conversion.
- **How it fits:** the second mode of a multi-effort feature, on branch `feature/pf2e-encounter-simulator`. The first product work under `campaigns`. Version 1 of a larger PF2e tool; the ruleset as data and an encounter designer come later, in their own initiatives.

## Context / Goal

Build the six efforts the [discovery](../discovery/F.041-pf2e-encounter-simulator-discovery.md) designed (D-21), in the order Decision 1 sets: engine core, Foundry import, rules coverage, description conversion, tactics and position, then scenario and runner.

- Discovery decisions D-1 to D-26 stand. Decision 1 amends D-21's order and the effort numbers the discovery uses: the discovery's E3 (conversion) is this plan's Effort 4, and its E4 (rules coverage) is this plan's Effort 3. Elsewhere, "E3" and "E4" mean this plan's numbers.
- The discovery's one Open Question (where the Claude conversion runs) is resolved in its timing by Decision 2; the choice itself is made hands-on in Effort 4.
- Acceptance is the discovery's four gates (D-20): stat parity, every rule tested, deterministic runs, replay review. Each gate lands in a named effort (Effort Summary).

## Decisions Ratified in Planning

1. **Rules coverage runs before description conversion: Effort 3 is rules coverage, Effort 4 is conversion.** Robert's call; amends D-21's order. In the discovery order, conversion ran against Effort 1's core schema, so most boss abilities (Destiny Spores needs slowed, Warp Mind needs confused) would come back "unsupported". Each rules primitive added afterwards would bump the schema version and send every approved ability back for approval, because the approval cache is keyed on text and schema version. With rules first, conversion runs once against the full schema. Rules coverage never needed conversion: gate 2 tests each rule with hand-written ability files. A mechanic the schema cannot express, found during conversion, is added in Effort 4 before its approval pass, so no approval is redone. Cost: the two conversion options are tried one effort later than D-21 intended. Rejected: the discovery order with one re-conversion pass at the end of rules coverage (a second approval pass, and Effort 3's end becomes "converted or unsupported"); the full schema designed in Effort 1 (schema designed before the engine code that runs it, so rules coverage would likely reshape it and bump the version anyway).

2. **The conversion trial opens Effort 4.** Robert's call; resolves the timing half of the discovery's Open Question. A thin version of each option (the Claude API in the importer with a `review` command, and a Claude Code skill) runs on the real output of the Effort 2 importer, so the trial converts exactly the sentences the converter will see, against Effort 3's full schema. Robert picks one; the rest of Effort 4 builds it in full and removes the other. Rejected: the trial at the start of Effort 2 (hand-fed text, before tags are converted).

3. **Every effort gets its own Plan session; this file carries no effort spec. The efforts run in order, `E1 → E2 → E3 → E4 → E5 → E6`.** Robert's call. Each effort is a design surface large enough for its own plan, and Efforts 2 to 6 write literal code against packages that earlier efforts create. Each effort plan is written just before its effort runs, against code that exists. Same shape as the A.001.8 overview, Decision 1. Rejected: Effort 1 specified here (an overview mixed with one effort's code).

4. **All v1 code lives under `internal/pf2e/sim/`.** Robert's call. `internal/pf2e/` is the PF2e root, a sibling of `internal/faction/` and `internal/spatial/`; `internal/pf2e/rules/` is left for the later ruleset-as-data work. The creature files, the importer and the conversion sit under `sim/` too, since v1 has no other PF2e consumer. Each effort plan names its sub-packages. Rejected: `internal/campaigns/pf2e/` (puts an engine under the campaign registry package).

5. **The CLI is `gm-toolkit pf2e sim <command>`, with all eight v1 commands under `sim`, and flags at each level.** Robert's call. `pf2e` carries a persistent `--campaign` that every sub-command inherits, as `faction` does; `sim` and its commands carry their own flags (`--seed`, run count, preset, variant). The commands follow the package line: if `rules/` later takes over the creature files, `import` and `show` move up then. Commands land with their effort:

   | Command | Effort |
   |---------|--------|
   | `show <creature>` | 1 |
   | `import <fight>` | 2 |
   | `check <fight>` | 4 |
   | `review <fight>` (if the API option wins) | 4 |
   | `run`, `compare`, `replay`, `scenario create` | 6 |

   Rejected: `fight` or `encounter` as the group (fits the later designer and ruleset less well); the data commands one level up under `pf2e` (guesses now what `rules/` will own).

6. **Files every fight reuses live in `campaigns/<id>/pf2e/`; each fight lives in `campaigns/<id>/encounters/<fight>/`.** Robert's call; settles D-19's folder names. `pf2e/` holds `party/<pc>/` and `scoring.toml`, as the discovery's State Storage sketch shows; `encounters/fall-of-luna/` already exists with its `source/` and `reference/`. File names inside the folders (the creature layers, the approval cache, the scenario file) are set by the effort plan that first writes them. Rejected: `pf2e/sim/` (deeper, and the party is not sim-only data); a bare `party/` (loses the system name in a campaign that also runs SWN factions).

7. **The wiki page owners follow the new order.** D-23's table becomes: `overview.md` written by Effort 1 and extended by every effort; `data-files.md` written by Effort 1, extended by Effort 2 (import) and Effort 4 (conversion); `engine.md` written by Effort 1, extended by Effort 3 (rules); `sim.md` written by Effort 5, extended by Effort 6. D-24's front-door restructure stays in Effort 1.

## Decision Record — Execution

## Effort Summary

| Done | # | Effort | Scope | Gate | File |
|------|---|--------|-------|------|------|
| [x] | 1 | Engine core | Stat block, seeded roller, checks, Strikes, damage, dying, turn loop, events; creature and ability formats, layer merge, `show`; wiki front door and three pages | 3 — deterministic | [`F.041.E1-pf2e-encounter-simulator-plan.md`](./F.041.E1-pf2e-encounter-simulator-plan.md) |
| [x] | 2 | Foundry import | NPC import (numbers and Strikes), PC import as choices, load-time PC stats, import report, `import` | 1 — stat parity | [`F.041.E2-pf2e-encounter-simulator-plan.md`](./F.041.E2-pf2e-encounter-simulator-plan.md) |
| [ ] | 3 | Rules coverage | The rest of the discovery's Rules coverage table, each rule in the engine and the schema, each tested | 2 — every rule tested | [`F.041.E3-pf2e-encounter-simulator-plan.md`](./F.041.E3-pf2e-encounter-simulator-plan.md) |
| [ ] | 4 | Description conversion | Thin trial of both options, then the chosen one in full: drafts, approval, cache, three kinds of answer; D-8 in `check` | — | [`F.041.E4-pf2e-encounter-simulator-plan.md`](./F.041.E4-pf2e-encounter-simulator-plan.md) |
| [ ] | 5 | Tactics and position | Rule-list tactics, the condition vocabulary, dials, range bands and spread, PC and boss tactics files | — | [`F.041.E5-pf2e-encounter-simulator-plan.md`](./F.041.E5-pf2e-encounter-simulator-plan.md) |
| [ ] | 6 | Scenario and runner | Scenario file, phases and carry-over, variants, presets, scoring, `run`, `compare`, `replay`, `scenario create`; the script deleted | 4 — replay review | [`F.041.E6-pf2e-encounter-simulator-plan.md`](./F.041.E6-pf2e-encounter-simulator-plan.md) |

## Shared Context

### Branch and git cadence

- Branch `feature/pf2e-encounter-simulator`, cut when Discovery started. Every effort session runs on it, Plan sessions included.
- Every commit for effort n carries the scope `E<n>`, its plan doc included (`docs(E1): ...`). This overview's commit is unscoped.
- At each effort boundary, `/gestalt:effort-close F.041.E<n>` runs the interim review, one `fix(E<n>):` commit and the squash. Merge to `main` with `--no-ff` at `/gestalt:initiative-close F.041`.
- Each effort writes or extends its wiki page as its last step (D-23, Decision 7).

### Before Effort 2's Plan session

- **The gate 1 fixture** is [`F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`](./F.041.E2-pf2e-encounter-simulator-gate1-fixture.md): the level-14 numbers derived by hand from the exports, checked by Robert against the Foundry sheets (E2 plan, Decision 1).

### Findings for the effort plans

Verified against the source in this session (2026-11-15). Each effort plan re-reads the code before it relies on them.

- **CLI root.** `NewRootCmd` (`cmd/gm-toolkit/root.go`) adds `campaign`, `spatial` and `faction`. Its `Short` and `Long` still say "Stars Without Number"; the effort that adds `pf2e` rewords them. `faction` shows the `--campaign` override pattern (`cmd/gm-toolkit/faction.go`, `campaigns.ResolveActive`). (Effort 1)
- **Campaign paths.** `campaigns.Paths` (`internal/campaigns/campaigns.go:47`) lists only faction and spatial paths, through `path:` and `kind:` struct tags read by `Paths()` (`:59`). No PF2e path exists yet. (Efforts 1 and 2)
- **Dice.** The faction roller (`internal/faction/engine/dice/`) is bound to SWN hooks and the faction domain, so nothing lifts out; the PF2e roller is new (discovery, Architecture Context). (Effort 1)
- **Exports.** All five actor exports are pf2e system 8.6.1, Foundry 14.365. PCs are level 14 (Kas 99 items, Tavish 170, Arabesque 138); Abyssal Luna is level 16 with 19 items, the Avatar of Lilith level 17 with 22. `source/items/` holds two more exports, `effect-faithful-strike.json` and `swear-fidelity.json`, which the importer must place. (Effort 2)
- **Tags before rules.** Tags convert in Go with no approval, but every tag on the two bosses names a rule the engine lacks until Effort 3, so the tag converter itself moved to Effort 3 (E2 plan, Decision 4). Effort 2 lists each ability in the import report; Effort 3 builds the converter with each primitive and re-runs the import at its end. A re-import costs nothing. (Efforts 2 and 3)
- **Reference material.** `reference/gm-flowchart.zip`, `build-order.zip` and `quick-numbers.zip` are zipped HTML pages (React bundles with images), not text. Effort 5 unzips the flowchart to write the boss tactics; gate 4 reads it again. (Efforts 5 and 6)
- **The script.** `cmd/scripts/hawk_finale_sim.py`, 816 lines: `PARTY` at `:46`, `DEFAULTS` (every knob) at `:60`, `RECOMMENDED` and `PRESETS` at `:107`, `luna_turn` at `:453`, `avatar_turn` at `:539`, `transition` at `:658`, `score` at `:698`, `compare` at `:760`. The knobs and presets are the source for Effort 6's presets and dials, and the boss turns cross-check Effort 5's tactics against the flowchart. (Efforts 5 and 6)
- **Default scoring.** The binary embeds no default (D-19). `rulebooks/` holds shared sources (`swn`, `_core`, `hawk`), and `campaign add-rules` (`cmd/gm-toolkit/campaign/add_rules.go`, `campaigns.ComposeRulebook`) copies them into the campaign's `rulebook/`, the faction folder. How a campaign gets its `pf2e/scoring.toml` is Effort 6's to settle; `rulebooks/pf2e/` is a likely source. (Effort 6)
- **Wizard pattern.** `campaign create` (`cmd/gm-toolkit/campaign/create.go`) is the `huh`-when-bare, flags-otherwise model for `scenario create` (D-15). (Effort 6)
- **Claude SDK.** `go.mod` has no `anthropic-sdk-go`. The thin API trial adds it; if the skill option wins, Effort 4 removes it again. (Effort 4)

### Gaps between efforts

Known and accepted. Each closes where named.

- **Until Effort 2,** every creature is hand-written: Effort 1's Luna and dummy PCs.
- **Until Effort 3,** only Effort 1's core rules run. Imported abilities that need other rules are skipped in the import report.
- **Until Effort 4,** description sentences are not converted, and `check` does not exist, so nothing enforces D-8. No `run` exists yet either, so no odds rest on the gap.
- **Until Effort 5,** creatures have no tactics files; fights in Efforts 1 to 4 use the fixed behaviour Effort 1 sets up for tests.
- **Until Effort 6,** there is no `run`, `compare` or `replay`. The Python script stays the working balance tool, with its old numbers.

### Pre-merge additions

- **Graduate to Key Decisions:** each `docs/architecture/pf2e/` page is written by its effort (D-23), so this step checks that each page's Key Decisions carry the durable calls: own data model with Foundry as an import source (D-1, D-6), creature layers (D-7, D-10), conversion approval and "the sim never calls Claude" (D-9), the position and decision seams (D-11, D-12), scoring as data (D-17), and Decisions 1, 4 and 5 here.
- **Front door:** confirm `architecture-overview.md` reads by product (D-24) and routes to the `pf2e/` side.
- **Contributing guide:** check whether `docs/contributing/overview.md` needs a PF2e route (build, test, the `pf2e sim` commands).
- **`planned-work.md`:** remove `F.041`. Add Backlog rows for what surfaced, at least: the ruleset as data (D-2), the encounter designer (D-2), a grid position model behind the seam (D-11), complex hazards (D-14) and a starting-state scenario setting (D-26).
- **The script** is gone (gate 4, Effort 6); confirm no doc still points at it.

## Out of Scope

- **Complex hazards,** including the Gate to the Abyss (D-14).
- **The full PF2e ruleset as data** (D-2), the future `internal/pf2e/rules/`.
- **An encounter designer** (D-2, D-15), including any interactive builder that edits tactics or tunes numbers.
- **A grid or map position model** (D-11); range bands only, behind the seam.
- **Classes beyond the finale's three** (fighter, sorcerer, legacy alchemist); a later fight's class is found by the import report and `check`.

---

## Work Breakdown

### Effort 1 — Engine core

**Scope:** discovery Proposed Effort Slicing, E1.
- The stat block, a seeded roller, checks and the four degrees of success, Strikes and the multiple attack penalty, damage with resistance, weakness and immunity, HP and dying, initiative, the turn loop with three actions and a reaction, and event reporting through an observer (headless, discovery Architecture Context).
- The creature and ability file formats and the layer merge (D-7), so the first fight is data, not Go. The ability schema covers the core rules; Effort 3 extends it.
- `internal/pf2e/sim/` and its first sub-packages (Decision 4); the `pf2e` and `sim` command groups with `--campaign` on `pf2e`, and `show` (Decision 5); the PF2e campaign paths (Decision 6).
- Wiki: restructure the front door around the products (D-24); write `pf2e/overview.md`, `data-files.md` (formats, layers) and `engine.md` (Decision 7).

**Ends with:** a hand-written Luna fights dummy PCs, and the same seed and files give identical results (gate 3).

**Depends on:** nothing in this initiative.

**Own Plan session:** yes (Decision 3). Expected size: several phases. The formats and merge, the engine core, and the CLI and wiki are separate seams.

### Effort 2 — Foundry import

**Scope:** discovery Import as a separate step and PC import.
- NPC import: numbers and Strikes as stored. Each ability is listed in the import report, not converted: the tag converter moved to Effort 3 (E2 plan, Decision 4), and sentences are Effort 4's.
- PC import as build choices (D-10), and the load-time computation: core math and the rule elements the gate numbers need (`FlatModifier`, `Resistance`, rank upgrades, `MartialProficiency`). `AdjustDegreeOfSuccess` and `DamageDice` moved to Effort 3 (E2 plan, Decision 5). Mixed editions follow the sheet (D-5).
- The import ignores current HP, conditions and temporary effects (D-26).
- The import report: per creature, what converted and what was skipped, with the reason; per PC, the computed numbers.
- `import`; `show` prints computed PC stats. Wiki: extend `data-files.md` with import.

**Ends with:** the computed level-14 stats for Kas, Tavish and Arabesque match Robert's fixture exactly (gate 1).

**Depends on:** Effort 1 (formats, merge, `show`); Robert's fixture (Shared Context).

**Own Plan session:** yes.

### Effort 3 — Rules coverage

**Scope:** the rest of the discovery's Rules coverage table, each rule as written (D-18).
- Conditions, reactions and triggers, fortune and misfortune, sustain and quickened, persistent damage, splash, fast healing and temporary HP, Protean Anatomy, spells with slots, areas, durations and counteract, recharge, frequency and uses per fight (D-25), and the random tables (Warpwave, Unfathomable Song).
- Effort 3 builds the tag converter (moved from Effort 2) and the `AdjustDegreeOfSuccess` and `DamageDice` imports. Each rule extends the ability schema and, where a tag names it, the converter. The import re-runs at the end so the import layers pick up the new rules.
- Named approximations stay listed (discovery, Named approximations). A rules reading that changes when or how often something fires is put to Robert as a question.
- Wiki: extend `engine.md`.

**Ends with:** every row of the Rules coverage table has tests (gate 2).

**Depends on:** Effort 2 (real creatures to test against, the tag converter to extend).

**Own Plan session:** yes. Expected size: the largest effort; the plan groups rules by seam into phases.

### Effort 4 — Description conversion

**Scope:** discovery Converting description sentences, with the trial order of Decision 2.
- First phase: thin versions of both options on the real import output. Robert tries both and picks one.
- Then the chosen option in full: one ability schema validated on both sides, the three kinds of answer (mechanics, unsupported, no combat effect), staged drafts, approval, the approval cache keyed on text and schema version (committed beside the import layers), and `review` if the API option wins.
- A mechanic the schema cannot express is added to the schema before the approval pass (Decision 1).
- `check`: validates every file and enforces D-8 without running. Wiki: extend `data-files.md` with conversion.

**Ends with:** every ability of Abyssal Luna and the Avatar of Lilith, and every PC passive and reaction, converted and approved, or marked `ignore` in a hand layer; `check` passes for Fall of Luna.

**Depends on:** Effort 2 (import layers), Effort 3 (the full schema).

**Own Plan session:** yes. Its trial phase is planned in full; the phases after it depend on the pick and may get a short follow-up Plan session.

### Effort 5 — Tactics and position

**Scope:** discovery Position and Tactics.
- Range bands, reach and speed, area capture by bands and the spread dial, behind the position seam (D-11).
- Rule-list tactics behind the decision seam (D-12): the shared condition vocabulary, chance on a rule, dials and habits.
- Tactics files for Luna, Lilith (from the GM flowchart, Findings, Reference material) and the three PCs (the script's habits as files or settings). The sin discovery curve as a scenario setting tactics read (`knows:sin`).
- Wiki: write `sim.md` (tactics, position).

**Ends with:** Luna and Lilith play the flowchart in test fights.

**Depends on:** Effort 3 (`ready:`, `in_area:` and the other rules conditions query), Effort 4 (the converted abilities tactics name).

**Own Plan session:** yes.

### Effort 6 — Scenario and runner

**Scope:** discovery Scenario, Scoring and User-Facing Impact.
- The scenario file (D-13): party, phases with carry-over, between-phase steps, variants as layer stacks, presets, run count and seed; dials such as "gate sealed" (D-16).
- Scoring as data (D-17): the default scoring file, per-scenario adjustments and events, the raw report lines, the named approximations in the report. How a campaign gets its scoring file (Findings, Default scoring).
- `run`, `compare`, `replay`, `scenario create`.
- Wiki: extend `sim.md` (scenario, scoring, CLI).
- Delete `cmd/scripts/hawk_finale_sim.py` once gate 4 passes.

**Ends with:** Robert reads a set of `replay` logs and confirms both bosses play the flowchart as written (gate 4); all four gates pass and the script is deleted.

**Depends on:** Effort 5.

**Own Plan session:** yes.
