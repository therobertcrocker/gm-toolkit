# PF2e Encounter Simulator (F.041) — Implementation Plan (Overview)

## In Plain Words

- **What:** the overview plan for the PF2e encounter simulator: a Go tool that plays a Pathfinder 2e fight thousands of times and reports how it tends to go. It fixes the order of the eight efforts, the package and command names, the campaign folder, and what each effort hands to the next. Each effort gets its own Plan session, so this file holds no code spec. Since 2026-11-19 the tool is sim core, the first initiative of the `A.004` arc: it ends with a `run` that plays the real Fall of Luna fight and prints raw facts.
- **Why:** the discovery designed the tool and its six efforts. This session re-read the code and the Fall of Luna data, settled the names the discovery left to Plan, and found one clash in the effort order: the description conversion came before the rules it needs. The fix is to run rules coverage before conversion. The Effort 3 Plan session later found rules coverage too large for one effort and split it into three, with a scope list of what the sheets need (Decisions 8 to 12). The `A.004` arc plan then moved position, compare, scoring and description conversion to three other initiatives, and narrowed Efforts 6 to 8 to what a first real `run` needs (Decision 13).
- **How it fits:** the second mode of a multi-effort feature, on branch `feature/pf2e-encounter-simulator`. The first product work under `campaigns`, and the first initiative of the [`A.004` arc](../arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md); tactics and flowcharts (`A.004.2`), measurement (`A.004.3`) and description conversion (`A.004.4`) build on it. The ruleset as data and an encounter designer come later, in their own initiatives.

## Context / Goal

Build the efforts the [discovery](../discovery/F.041-pf2e-encounter-simulator-discovery.md) designed (D-21), in the order Decisions 1 and 8 set and with the scope Decision 13 narrows: engine core, Foundry import, rules coverage in three efforts (effect core, triggers and limits, spells and tables), abilities and `check`, rule lists, then scenario and `run`.

- Discovery decisions D-1 to D-26 are read as the [arc-discovery](../arcs/A.004-encounter-lab/A.004-encounter-lab-arc-discovery.md) sorts them (its table, The `F.041` decisions): D-2 is reversed, D-17 is withdrawn, and D-13, D-19, D-20 and D-22 are amended. Decision 9 amends D-18. Decisions 1 and 8 amend D-21's order and the effort numbers the discovery uses: the discovery's E3 (conversion) is this plan's Effort 6, its E4 (rules coverage) is this plan's Efforts 3 to 5, its E5 (tactics and position) is Effort 7, and its E6 (scenario and runner) is Effort 8. Elsewhere, effort numbers are this plan's.
- The discovery's one Open Question (where the Claude conversion runs) left this initiative with description conversion; `A.004.4` owns it (Decision 13).
- The rules scope: [`F.041-pf2e-encounter-simulator-rules-scope.md`](./F.041-pf2e-encounter-simulator-rules-scope.md) (Decision 10).
- Acceptance is gates 1 to 3 of the discovery's four (D-20): stat parity, every rule tested, deterministic runs. Each lands in a named effort (Effort Summary). Gate 4, the replay review, lands in `A.004.2` (arc plan AP-6).

## Decisions Ratified in Planning

1. **Rules coverage runs before description conversion: Effort 3 is rules coverage, Effort 4 is conversion.** Robert's call; amends D-21's order. In the discovery order, conversion ran against Effort 1's core schema, so most boss abilities (Destiny Spores needs slowed, Warp Mind needs confused) would come back "unsupported". Each rules primitive added afterwards would bump the schema version and send every approved ability back for approval, because the approval cache is keyed on text and schema version. With rules first, conversion runs once against the full schema. Rules coverage never needed conversion: gate 2 tests each rule with hand-written ability files. A mechanic the schema cannot express, found during conversion, is added in Effort 4 before its approval pass, so no approval is redone. Cost: the two conversion options are tried one effort later than D-21 intended. Rejected: the discovery order with one re-conversion pass at the end of rules coverage (a second approval pass, and Effort 3's end becomes "converted or unsupported"); the full schema designed in Effort 1 (schema designed before the engine code that runs it, so rules coverage would likely reshape it and bump the version anyway). Decision 8 later splits rules coverage into Efforts 3 to 5, so conversion is Effort 6; the order this decision sets holds. Decision 13 then moves conversion out of this initiative, to `A.004.4`; it still runs after the rules.

2. **The conversion trial opens Effort 6** (Effort 4 before Decision 8; since Decision 13, the trial opens `A.004.4`). Robert's call; resolves the timing half of the discovery's Open Question. A thin version of each option (the Claude API in the importer with a `review` command, and a Claude Code skill) runs on the real output of the Effort 2 importer, so the trial converts exactly the sentences the converter will see, against the full schema of Efforts 3 to 5. Robert picks one; the rest of Effort 6 builds it in full and removes the other. Rejected: the trial at the start of Effort 2 (hand-fed text, before tags are converted).

3. **Every effort gets its own Plan session; this file carries no effort spec. The efforts run in order, `E1 → E2 → … → E8`.** Robert's call. Each effort is a design surface large enough for its own plan, and Efforts 2 to 8 write literal code against packages that earlier efforts create. Each effort plan is written just before its effort runs, against code that exists. Same shape as the A.001.8 overview, Decision 1. Rejected: Effort 1 specified here (an overview mixed with one effort's code).

4. **All v1 code lives under `internal/pf2e/sim/`.** Robert's call. `internal/pf2e/` is the PF2e root, a sibling of `internal/faction/` and `internal/spatial/`; `internal/pf2e/rules/` is left for the later ruleset-as-data work. The creature files, the importer and the conversion sit under `sim/` too, since v1 has no other PF2e consumer. Each effort plan names its sub-packages. Rejected: `internal/campaigns/pf2e/` (puts an engine under the campaign registry package).

5. **The CLI is `gm-toolkit pf2e sim <command>`, with all eight v1 commands under `sim`, and flags at each level.** Robert's call. `pf2e` carries a persistent `--campaign` that every sub-command inherits, as `faction` does; `sim` and its commands carry their own flags (`--seed`, run count, preset, variant). The commands follow the package line: if `rules/` later takes over the creature files, `import` and `show` move up then. Commands land with their effort:

   | Command | Effort |
   |---------|--------|
   | `show <creature>` | 1 |
   | `import <fight>` | 2 |
   | `check <fight>` | 6 |
   | `review <fight>` (if the API option wins) | 6 |
   | `run`, `compare`, `replay`, `scenario create` | 8 |

   Amended by Decision 13: `check` stays in Effort 6, and `run` and `replay` in Effort 8. `compare` moves to `A.004.2`, `review` to `A.004.4`, and `scenario create` to the Backlog (`F.060`).

   Rejected: `fight` or `encounter` as the group (fits the later designer and ruleset less well); the data commands one level up under `pf2e` (guesses now what `rules/` will own).

6. **Files every fight reuses live in `campaigns/<id>/pf2e/`; each fight lives in `campaigns/<id>/encounters/<fight>/`.** Robert's call; settles D-19's folder names. `pf2e/` holds `party/<pc>/` and `scoring.toml`, as the discovery's State Storage sketch shows (the scoring file went with D-17, Decision 13); `encounters/fall-of-luna/` already exists with its `source/` and `reference/`. File names inside the folders (the creature layers, the approval cache, the scenario file) are set by the effort plan that first writes them. Rejected: `pf2e/sim/` (deeper, and the party is not sim-only data); a bare `party/` (loses the system name in a campaign that also runs SWN factions).

7. **The wiki page owners follow the new order.** D-23's table becomes: `overview.md` written by Effort 1 and extended by every effort; `data-files.md` written by Effort 1, extended by Effort 2 (import), Efforts 3 to 5 (the ability schema) and Effort 6 (conversion); `engine.md` written by Effort 1, extended by Efforts 3 to 5 (rules); `sim.md` written by Effort 7, extended by Effort 8. D-24's front-door restructure stays in Effort 1. (Effort numbers as amended by Decision 8.) Amended by Decision 13: `data-files.md` is extended by Effort 6 with the hand-written abilities and `check`, and conversion's part is `A.004.4`'s; `sim.md` is still written by Effort 7 and extended by Effort 8, with the narrowed scope.

8. **Rules coverage is three efforts: Effort 3 effect core, Effort 4 triggers and limits, Effort 5 spells and tables. Conversion becomes Effort 6, tactics and position Effort 7, scenario and runner Effort 8.** Robert's call, in the Effort 3 Plan session. An audit of the five exports found about 50 rule primitives (13 reaction triggers, about 18 conditions, slots and heightening, persistent and splash damage, recharge and frequency); as literal code that is 12 to 15 phases, past what one Plan session can write and far past the 400-line front-matter budget. The cut follows seams: effects a step applies, then when things fire and how often, then spells and their tables. Each effort fixes the Effort 2 import findings for the rules it builds. Rejected: one effort planned across several Plan sessions (no review until the end, front matter far over budget); two efforts (the first still 8 to 10 phases).

9. **A stand-in effect may replace a rule that is hard to build on its own; this amends D-18.** Robert's call. A stand-in is built only from shared building blocks, lives in the creature's hand layer with an `approximation` note, and the run report prints it under the named approximations, as the script's CAVEATS block does. So it is simpler, never hidden. Gate 2 tests each stand-in as specified. Each stand-in's wording is Robert's call. The stand-ins and the other named approximations are listed in the rules scope, section 2. Rejected: D-18 as written for every item (custom code for each hard item); dropping every hard item (the fight loses Walk the Timelines and Sin Vulnerability).

10. **The rules efforts model what the rules scope keeps: about 110 items from the sheets, filtered three ways.** Robert's call. Step 1 kept what is useful in this fight; step 2 picked as written, stand-in or drop for each item that is hard on its own; step 3 picked keep or drop for each survivor that needs code of its own and is not in the script. The list is [`F.041-pf2e-encounter-simulator-rules-scope.md`](./F.041-pf2e-encounter-simulator-rules-scope.md), committed beside the plans like the gate 1 fixture. It replaces the discovery's Rules coverage table as gate 2's checklist. Every dropped ability is marked `ignore` by Effort 6, so D-8 holds; Effort 6 does that by hand (Decision 13). Rejected: every active on the sheets (Wall of Stone, summons and spell stealing need a map or their own systems); only the actives the script uses (drops most of Arabesque's spell list and Kas's fighter feats).

11. **Rule data the exports lack is typed into the creature's hand layer.** Robert's call. The Warpwave table, Wails of the Damned and Kas's two property runes live in Foundry's compendium, not in the exports. Each goes in the hand layer of the creature that uses it, by the effort that builds its rule (rules scope, section 4). Rejected: a shared campaign file the layers refer to (a new format and a new load step for four entries).

12. **The tag converter moves to Effort 6.** Robert's call; amends the Effort 2 plan's Decision 4. Every boss ability with tags also has sentences (Destiny Spores needs its "slowed on a failure" sentence), so in a rules effort nothing would use the converter's output. Effort 6 builds it where tags and sentences become one ability, and its output shape follows that use. Rejected: building it piece by piece with each rule (output unused until Effort 6, and likely reshaped there). Since Decision 13 the converter goes to `A.004.4` with the rest of conversion.

13. **The `A.004` arc narrows this initiative to sim core: Efforts 6 to 8 keep their themes and shrink, and one short Discovery runs before Effort 6.** Robert's calls, in the Arc-Discovery and Arc-Plan sessions (2026-11-19); the [arc plan](../arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md) holds the full rationale (AP-1 to AP-10). Amends Decisions 1, 2, 5, 6, 7, 10 and 12, and reads D-17 as withdrawn and D-20 as amended.

    - **Effort 6 is abilities and `check`.** Every kept item of the rules scope that the import does not cover is written by hand in the hand layers of all five creatures, bosses and PCs (AP-3, ACD-8). The files are the answer key for the conversion trial in `A.004.4`.
    - **Effort 7 is rule lists.** The format, a small condition set, the decider that replaces `FirstFitting`, and hand-written lists for the five creatures. Rule lists are plain first-match-wins; chance, dials and position are `A.004.2`'s (ACD-7).
    - **Effort 8 is scenario and `run`.** The scenario file, `run`, the fight record that the other initiatives read (ACD-4), raw facts and `replay`. No scoring: D-17 is withdrawn, and `A.004.3` decides what replaces it.
    - **A formats Discovery runs before Effort 6** (AP-4). It designs the scenario file, the rule-list format and the fight record together, and answers the two questions the arc-discovery sent to sim core: the small condition set and the fight record's field list.
    - **Left this initiative:** position, chance, dials, variants, presets, `compare`, the flowchart, gate 4 and the script's archive (`A.004.2`); measures and scoring (`A.004.3`); the tag converter, the trial, drafts, approval, the cache and `review` (`A.004.4`); `scenario create` (Backlog, `F.060`).
    - **The script is archived, not deleted,** by `A.004.2` when gate 4 passes (AP-6).

    Rejected: the borrowed work in two efforts; scenario and `run` before the rule lists (a toy fight, against ACD-5); the PCs' abilities left for conversion; the formats designed inside the effort plans.

## Decision Record — Execution

## Effort Summary

| Done | # | Effort | Scope | Gate | File |
|------|---|--------|-------|------|------|
| [x] | 1 | Engine core | Stat block, seeded roller, checks, Strikes, damage, dying, turn loop, events; creature and ability formats, layer merge, `show`; wiki front door and three pages | 3 — deterministic | [`F.041.E1-pf2e-encounter-simulator-plan.md`](./F.041.E1-pf2e-encounter-simulator-plan.md) |
| [x] | 2 | Foundry import | NPC import (numbers and Strikes), PC import as choices, load-time PC stats, import report, `import` | 1 — stat parity | [`F.041.E2-pf2e-encounter-simulator-plan.md`](./F.041.E2-pf2e-encounter-simulator-plan.md) |
| [x] | 3 | Effect core | Ability schema v2 (save, damage, heal, condition, spell attack and area steps); typed bonuses and trait modifiers in the fight; conditions and durations; persistent and splash damage; adaptive and trait-keyed defenses; the Down row | 2 — part | [`F.041.E3-pf2e-encounter-simulator-plan.md`](./F.041.E3-pf2e-encounter-simulator-plan.md) |
| [x] | 4 | Triggers and limits | Reactions and triggers, fortune and misfortune, recharge, frequency and uses, action economy, the loyalty tag and Sin Vulnerability | 2 — part | [`F.041.E4-pf2e-encounter-simulator-plan.md`](./F.041.E4-pf2e-encounter-simulator-plan.md) |
| [ ] | 5 | Spells and tables | Spell import (entries, slots, heightening, innate uses), counteract, incapacitation, temporary immunity, random tables; the import re-run | 2 — every rule tested | [`F.041.E5-pf2e-encounter-simulator-plan.md`](./F.041.E5-pf2e-encounter-simulator-plan.md) |
| [ ] | — | Formats Discovery | The scenario file, the rule-list format and the fight record, designed together (Decision 13) | — | A discovery doc, named by its session |
| [ ] | 6 | Abilities and `check` | Every kept item the import does not cover, written by hand for all five creatures; `ignore` marks; `check` with the D-8 refusal | — | [`F.041.E6-pf2e-encounter-simulator-plan.md`](./F.041.E6-pf2e-encounter-simulator-plan.md) |
| [ ] | 7 | Rule lists | The rule-list format, a small condition set, the rule-list decider, hand-written lists for the five creatures; `sim.md` | — | [`F.041.E7-pf2e-encounter-simulator-plan.md`](./F.041.E7-pf2e-encounter-simulator-plan.md) |
| [ ] | 8 | Scenario and `run` | Scenario file, phases and carry-over, run count and seed, `run`, the fight record, raw facts, `replay` | — | [`F.041.E8-pf2e-encounter-simulator-plan.md`](./F.041.E8-pf2e-encounter-simulator-plan.md) |

## Shared Context

### Branch and git cadence

- Branch `feature/pf2e-encounter-simulator`, cut when Discovery started. Every effort session runs on it, Plan sessions included.
- Every commit for effort n carries the scope `E<n>`, its plan doc included (`docs(E1): ...`). This overview's commit is unscoped.
- At each effort boundary, `/gestalt:effort-close F.041.E<n>` runs the interim review, one `fix(E<n>):` commit and the squash. Merge to `main` with `--no-ff` at `/gestalt:initiative-close F.041`. The arc's docs ride on this branch too (ACD-1); each later initiative of `A.004` cuts its own branch after this merge.
- Each effort writes or extends its wiki page as its last step (D-23, Decision 7).

### Before Effort 2's Plan session

- **The gate 1 fixture** is [`F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`](./F.041.E2-pf2e-encounter-simulator-gate1-fixture.md): the level-14 numbers derived by hand from the exports, checked by Robert against the Foundry sheets (E2 plan, Decision 1).

### Findings for the effort plans

Verified against the source in the overview Plan session (2026-11-15). Each effort plan re-reads the code before it relies on them. The effort tags follow Decision 13; a finding tagged with an `A.004` ID is for that initiative's Discovery.

- **CLI root.** `NewRootCmd` (`cmd/gm-toolkit/root.go`) adds `campaign`, `spatial` and `faction`. Its `Short` and `Long` still say "Stars Without Number"; the effort that adds `pf2e` rewords them. `faction` shows the `--campaign` override pattern (`cmd/gm-toolkit/faction.go`, `campaigns.ResolveActive`). (Effort 1)
- **Campaign paths.** `campaigns.Paths` (`internal/campaigns/campaigns.go:47`) lists only faction and spatial paths, through `path:` and `kind:` struct tags read by `Paths()` (`:59`). No PF2e path exists yet. (Efforts 1 and 2)
- **Dice.** The faction roller (`internal/faction/engine/dice/`) is bound to SWN hooks and the faction domain, so nothing lifts out; the PF2e roller is new (discovery, Architecture Context). (Effort 1)
- **Exports.** All five actor exports are pf2e system 8.6.1, Foundry 14.365. PCs are level 14 (Kas 99 items, Tavish 170, Arabesque 138); Abyssal Luna is level 16 with 19 items, the Avatar of Lilith level 17 with 22. `source/items/` holds two more exports, `effect-faithful-strike.json` and `swear-fidelity.json`, which the importer must place. (Effort 2)
- **Tags before rules.** Tags convert in Go with no approval, but every tag on the two bosses names a rule the engine lacks until Efforts 3 to 5, and every tagged ability also has sentences. The tag converter moved to Effort 3 (E2 plan, Decision 4), then to Effort 6 (Decision 12). Effort 2 lists each ability in the import report; Effort 5 re-runs the import once the rules exist. A re-import costs nothing. The converter itself is now `A.004.4`'s (Decision 13). (Efforts 2 and 5, `A.004.4`)
- **Reference material.** `reference/gm-flowchart.zip`, `build-order.zip` and `quick-numbers.zip` are zipped HTML pages (React bundles with images), not text. Effort 7 unzips the flowchart to write the boss rule lists; gate 4 reads it again. (Effort 7, `A.004.2`)
- **The script.** `cmd/scripts/hawk_finale_sim.py`, 816 lines: `PARTY` at `:46`, `DEFAULTS` (every knob) at `:60`, `RECOMMENDED` and `PRESETS` at `:107`, `luna_turn` at `:453`, `avatar_turn` at `:539`, `transition` at `:658`, `score` at `:698`, `compare` at `:760`. The knobs and presets are the source for the presets and dials of `A.004.2`, and the boss turns cross-check Effort 7's rule lists against the flowchart. (Effort 7, `A.004.2`)
- **Default scoring.** The binary embeds no default (D-19). `rulebooks/` holds shared sources (`swn`, `_core`, `hawk`), and `campaign add-rules` (`cmd/gm-toolkit/campaign/add_rules.go`, `campaigns.ComposeRulebook`) copies them into the campaign's `rulebook/`, the faction folder. D-17 is withdrawn (Decision 13), so no scoring file is planned; whether measures are data a campaign owns is the `A.004.3` Discovery's question. (`A.004.3`)
- **Wizard pattern.** `campaign create` (`cmd/gm-toolkit/campaign/create.go`) is the `huh`-when-bare, flags-otherwise model for `scenario create` (D-15). (Backlog, `F.060`)
- **Strike range.** Found at the Effort 2 close (2026-11-17): `creature.Weapon` has `range`, but `creature.Strike` does not, so a PC's computed Strike drops it. An NPC's Foundry Strike has a `range` field (null on all five Fall of Luna Strikes) that the importer does not read; reach comes across as a `reach-<n>` trait. Position needs range on the Strike. (`A.004.2`)
- **Claude SDK.** `go.mod` has no `anthropic-sdk-go`. The thin API trial adds it; if the skill option wins, the trial's initiative removes it again. (`A.004.4`)
- **Sheet surprises.** Found in the Effort 3 Plan session (2026-11-17), for Robert to check in Foundry: Tavish's key attribute is stored as Str; Spellsap Grenade's data says 1d6 but its text 3d4 plus splash (dropped anyway, rules scope section 5); Protector's Fury has a maul's stats and a hooked hammer's description. The Avatar is immune to death effects, so Vampiric Exsanguination does nothing to her, though the script deals it. (Efforts 3 and 6)
- **Demoralize moved from Effort 3 to Effort 4.** It was listed under Effort 3 and built in Effort 4, with the skill computation (E4 Decision 16): a check needs skill modifiers, and those are computed from the build. (Effort 4)
- **A bare bomb Strike spends nothing.** Found in Effort 4 (E4 Decision Record 62): the quick-bombs pool limits only the `quick-bomb` abilities, so a bare Strike with any imported bomb, the brewed Nail Bomb included, has no limit. The import report's "typed as a pool" line promises pools no hand layer types. Design question: each bomb becomes an ability on a pool, and a bare bomb Strike stops being a choice. (Effort 6)
- **The E4 abilities exist twice.** Found in Effort 4 (E4 Decision Record 71): in `internal/pf2e/sim/engine/testdata/scope/` and in the real hand layers under `campaigns/hawk`, and the two can drift. Effort 6 names one as the source. (Effort 6)
- **An ordinary turn logs no action count.** Found in Effort 4 (E4 Decision Record 70): `ActionsRegained` fires only when a creature is quickened or loses actions, so the fight record cannot read a plain turn's action count from events alone. (Effort 8)

### Gaps between efforts

Known and accepted. Each closes where named.

- **Until Effort 2,** every creature is hand-written: Effort 1's Luna and dummy PCs.
- **Until Effort 5,** only part of the rules run. Imported abilities that need other rules are skipped in the import report, and each rules effort tests its rules with hand-written ability files.
- **Until Effort 6,** the kept abilities have no mechanics outside test files, and `check` does not exist, so nothing enforces D-8. No `run` exists yet either, so no odds rest on the gap.
- **Until Effort 7,** creatures have no rule lists; fights in Efforts 1 to 6 use the fixed behaviour Effort 1 sets up for tests.
- **Until Effort 8,** there is no `run` or `replay`. The Python script stays the working balance tool, with its old numbers.
- **Past this initiative,** every creature counts as engaged with every foe until `A.004.2` builds position; `run` prints that as a named approximation (D-11). A kept item whose last rule is position (Sudden Charge, splash on other creatures) gets its treatment in the Effort 6 Plan session, as a question to Robert per item. There is no `compare`, no measure and no conversion until `A.004.2`, `A.004.3` and `A.004.4`; the script stays in `cmd/scripts/` until gate 4.

### Pre-merge additions

- **Graduate to Key Decisions:** each `docs/architecture/pf2e/` page is written by its effort (D-23), so this step checks that each page's Key Decisions carry the durable calls: own data model with Foundry as an import source (D-1, D-6), creature layers (D-7, D-10), "the sim never calls Claude" (D-9), the position and decision seams (D-11, D-12), the one fight record that other initiatives read (ACD-4), one scenario file that several initiatives extend (ACD-6), stand-ins named in the report (Decision 9), and Decisions 4 and 5 here.
- **Front door:** confirm `architecture-overview.md` reads by product (D-24) and routes to the `pf2e/` side.
- **Contributing guide:** check whether `docs/contributing/overview.md` needs a PF2e route (build, test, the `pf2e sim` commands).
- **`planned-work.md`:** remove `F.041`; the `A.004` row stays until the arc closes. Add Backlog rows for what surfaced, at least: the ruleset as data (D-2), the encounter designer (D-2), a grid position model behind the seam (D-11), complex hazards (D-14) and a starting-state scenario setting (D-26).
- **The script** stays in `cmd/scripts/` at this merge; `A.004.2` archives it after gate 4 (Decision 13). Confirm `docs/architecture/pf2e/overview.md` no longer describes the runner as planned work.

## Out of Scope

- **Complex hazards,** including the Gate to the Abyss (D-14).
- **The full PF2e ruleset as data** (D-2), the future `internal/pf2e/rules/`.
- **An encounter designer** (D-2, D-15), including any interactive builder that edits tactics or tunes numbers.
- **A grid or map position model** (D-11); range bands only, behind the seam.
- **The rest of the `A.004` arc** (Decision 13): position, chance, dials, variants, presets, `compare`, the flowchart and gate 4 (`A.004.2`); measures and scoring (`A.004.3`); description conversion (`A.004.4`); `scenario create` (`F.060`).
- **Classes beyond the finale's three** (fighter, sorcerer, legacy alchemist); a later fight's class is found by the import report and `check`.

---

## Work Breakdown

### Effort 1 — Engine core

**Scope:** discovery Proposed Effort Slicing, E1.
- The stat block, a seeded roller, checks and the four degrees of success, Strikes and the multiple attack penalty, damage with resistance, weakness and immunity, HP and dying, initiative, the turn loop with three actions and a reaction, and event reporting through an observer (headless, discovery Architecture Context).
- The creature and ability file formats and the layer merge (D-7), so the first fight is data, not Go. The ability schema covers the core rules; Efforts 3 to 5 extend it.
- `internal/pf2e/sim/` and its first sub-packages (Decision 4); the `pf2e` and `sim` command groups with `--campaign` on `pf2e`, and `show` (Decision 5); the PF2e campaign paths (Decision 6).
- Wiki: restructure the front door around the products (D-24); write `pf2e/overview.md`, `data-files.md` (formats, layers) and `engine.md` (Decision 7).

**Ends with:** a hand-written Luna fights dummy PCs, and the same seed and files give identical results (gate 3).

**Depends on:** nothing in this initiative.

**Own Plan session:** yes (Decision 3). Expected size: several phases. The formats and merge, the engine core, and the CLI and wiki are separate seams.

### Effort 2 — Foundry import

**Scope:** discovery Import as a separate step and PC import.
- NPC import: numbers and Strikes as stored. Each ability is listed in the import report, not converted: tags and sentences are Effort 6's (E2 plan, Decision 4; Decision 12).
- PC import as build choices (D-10), and the load-time computation: core math and the rule elements the gate numbers need (`FlatModifier`, `Resistance`, rank upgrades, `MartialProficiency`). `AdjustDegreeOfSuccess` and `DamageDice` moved to the rules efforts (E2 plan, Decision 5). Mixed editions follow the sheet (D-5).
- The import ignores current HP, conditions and temporary effects (D-26).
- The import report: per creature, what converted and what was skipped, with the reason; per PC, the computed numbers.
- `import`; `show` prints computed PC stats. Wiki: extend `data-files.md` with import.

**Ends with:** the computed level-14 stats for Kas, Tavish and Arabesque match Robert's fixture exactly (gate 1).

**Depends on:** Effort 1 (formats, merge, `show`); Robert's fixture (Shared Context).

**Own Plan session:** yes.

### Rules coverage — Efforts 3 to 5, shared

- **Scope:** the items the rules scope keeps (Decision 10), each as written (D-18) or as its stand-in (Decision 9). The scope's "Last rule in" column names the effort whose building block completes each item. The items' own mechanics are written by hand in Effort 6 (Decision 13); a rules effort tests each rule with hand-written ability files.
- **Import fixes:** each rules effort fixes the Effort 2 findings (E2 plan, Decision Record 2 and 5) for the rules it builds, and reports the rest with the reason.
- **Rules calls:** a reading that changes when or how often something fires is put to Robert as a question (rules scope, section 6).
- **Wiki:** each extends `engine.md` and the ability schema in `data-files.md` (Decision 7).
- **Gate 2** closes at Effort 5's end: every kept item in the rules scope has its rule tested.

### Effort 3 — Effect core

**Scope:** what an ability's steps do to a creature.
- Ability schema v2: steps for a save (basic or a degree table), damage, healing, temporary HP, a condition with a duration, a spell attack and an area, beside the existing Strike step.
- Checks in the fight: typed bonuses and penalties with stacking, modifiers keyed on a trait of the source (+2 against occult, +3 against emotion), degree shifts on a trait (Juggernaut, Bravery).
- The condition set with values, durations and end-of-turn decay, and the hooks that change a gain or a decay (Bravery, Calm and Centered).
- Damage: persistent and splash, the minimum of 1, weaknesses and resistances keyed on a trait or a material, Protean Anatomy, applied weakness, Tenderizer Grenade, critical riders (property runes, the hammer's critical specialization).
- The Down row: doomed, drained, hero points, the dying maximum, the recovery DC. Disintegrate's degree shift and death at 0 HP.
- Import: `AdjustDegreeOfSuccess`, `DamageDice`, and the damage findings of Effort 2 (persistent, splash, `holy`, doubled resistances).

**Ends with:** every Effort 3 row of the rules scope is tested.

**Depends on:** Effort 2.

**Own Plan session:** yes.

### Effort 4 — Triggers and limits

**Scope:** when things fire and how often.
- Reactions and their trigger kinds, Tactical Reflexes, Attack of Opportunity's disrupt; passive triggers (start of turn, a critical hit, rolling initiative).
- Fortune and misfortune (Challenge Fate, Ill Omen).
- Recharge and its reset, frequency, uses per fight (D-25), the quick-bomb counter, consumable counts.
- Action economy: quickened, slowed and stunned, extra actions, sustain, variable action cost, +1-action add-ons, the multiple attack penalty tweaks (Vicious Swing, Furious Focus, Attack of Opportunity, Brutal Finish).
- The `loyalty` tag and Sin Vulnerability; the stand-ins whose last rule is here.

**Ends with:** every Effort 4 row of the rules scope is tested.

**Depends on:** Effort 3 (the steps a trigger runs).

**Own Plan session:** yes.

### Effort 5 — Spells and tables

**Scope:** spellcasting and what is left of the rules scope.
- Spell import: entries, slots, ranks, heightening, innate uses (D-25), focus, and each spell's structured data (range, area, defense, damage, duration).
- Counteract, incapacitation, temporary immunity, True Target, Quickened Casting.
- Random tables (Warpwave, Unfathomable Song), with the Warpwave table typed into the Avatar's hand layer (Decision 11).
- The stand-ins whose last rule is here.
- The import re-runs at the end so the import layers pick up the new rules.

**Ends with:** gate 2: every kept item in the rules scope has its rule tested.

**Depends on:** Effort 4.

**Own Plan session:** yes.

### Formats Discovery — before Effort 6

**Scope:** one short Discovery session (Decision 13, arc plan AP-4), not an effort.
- The scenario file: party, enemies, phases with carry-over, run count and seed. Later initiatives add sections to the same file (ACD-6).
- The rule-list format and its small condition set (D-12, ACD-7).
- The fight record and its field list: one per-fight record built from the engine's events (ACD-4).

**Ends with:** a signed-off discovery doc in `discovery/`, which the plans of Efforts 6 to 8 read.

**Depends on:** Effort 5 (the full rules, so the conditions and the record name real things).

### Effort 6 — Abilities and `check`

**Scope:** what each creature can do, as data, and the check that nothing is missing.
- Every kept item of the rules scope that the import does not already cover, written by hand in the ability format in the hand layers of Luna, Lilith, Kas, Tavish and Arabesque.
- Every other ability marked `ignore` in a hand layer (Decision 10).
- A kept item whose last rule is position: a named approximation or `ignore`, asked of Robert per item in the Plan session.
- A mechanic the schema cannot express is added to the schema here.
- `check`: validates every file and enforces D-8 without running. Wiki: extend `data-files.md`.

**Ends with:** every ability the rules scope keeps has mechanics, and every other ability is marked `ignore` in a hand layer; `check` passes for Fall of Luna.

**Depends on:** Effort 2 (import layers), Efforts 3 to 5 (the full schema), the formats Discovery.

**Own Plan session:** yes.

### Effort 7 — Rule lists

**Scope:** how each creature chooses its turn, as data (D-12).
- The rule-list format and the small condition set, as the formats Discovery designs them; plain first-match-wins, with no chance, dials or position.
- The decider that plays a rule list, behind the same seam as `FirstFitting`.
- Hand-written lists for Luna and Lilith (from the GM flowchart, Findings, Reference material) and for the three PCs (from the script's habits).
- Wiki: write `sim.md` (rule lists).

**Ends with:** the five creatures play their lists in test fights.

**Depends on:** Efforts 3 to 5 (the rules that the conditions query), Effort 6 (the abilities the lists name).

**Own Plan session:** yes.

### Effort 8 — Scenario and `run`

**Scope:** who is fighting, and what happened.
- The scenario file (D-13, ACD-6): party, enemies, phases with carry-over, between-phase steps, run count and seed.
- The fight record (ACD-4), built from the engine's events.
- `run`: plays the scenario many times and prints raw facts (hit rates per side, save rates, damage share by source, knockdowns, rounds) and the named approximations. It refuses to run when `check` fails (D-8).
- `replay`: one fight as a readable log.
- Wiki: extend `sim.md` (scenario, the fight record, the CLI) and correct the planned-runner lines in `overview.md`.

**Ends with:** `run` plays the real Fall of Luna fight and prints raw facts; gates 1 to 3 hold.

**Depends on:** Effort 7.

**Own Plan session:** yes.
