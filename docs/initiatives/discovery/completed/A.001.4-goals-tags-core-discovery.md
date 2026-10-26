# Goals/Tags Neutralization + `_core` Home — Discovery

> **Initiative:** `A.001.4`, in the [data-driven-rulebook arc](../../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md).
> Binding arc decisions: ACD-6 (one shared home, deduped at scaffold) and ACD-7
> (neutralization is data-only), in the [arc-discovery](../../arcs/data-driven-rulebook/data-driven-rulebook-arc-discovery.md).
> This doc does not re-open them. Its agenda is the one question the arc-plan routed
> here (the flavor-vs-mechanic line, and the compose mechanics), plus internal slicing.
>
> **Status:** written; waits for Robert's sign-off. D-1 to D-13 are decided. The scope answer (D-7) amended the arc-plan as AP-8: ACD-6 and ACD-7 narrow
> to goals, and A.001.7 and A.001.8 join the arc.

## In Plain Words

- **What:** move the goals file out of the SWN rulebook folder into a shared folder,
  `rulebooks/_core/`. Campaign creation copies the shared folder and the chosen
  rulebook into the campaign. Then remove the SWN setting words from the goal text, so
  the same file reads correctly for SWN and for H.A.W.K. Tags stay in each rulebook.
  The tag and goal loader also becomes strict: unknown keys and duplicate IDs fail.
- **Why:** goals are the same rules in every game this tool runs. Today they live
  inside the SWN folder and use SWN words (for example "planet"). A second rulebook
  would have to copy them. Tags are the setting's identity, so each rulebook keeps its
  own.
- **How it fits:** this initiative has no dependency on the ability and effect
  registries. It can run at any time in the arc. A playable H.A.W.K campaign needs it,
  because H.A.W.K gets its goals from the shared folder. A.001.7 and A.001.8 (tags and
  goals as data) build on its strict loader.

## Problem

<br/>

**One source folder per campaign.** `campaigns.CopyRulebook`
(`internal/campaigns/scaffold.go:50`) walks one source folder and copies every
`.toml` into `<campaign>/rulebook/`. Its two CLI callers (`campaign create --rules`,
`campaign add-rules --rules`) take one path. `rulebooks/swn/` holds the goals and
tags, so a second rulebook has no shared source to copy them from.

**SWN setting words in shared rules.** `goals.toml` uses engine words in SWN terms
("planet", "planetary government"). A shared rule must read correctly in both games.
`tags.toml` names SWN organizations, SWN technology and SWN space terms, but after D-7
it is not shared. See the inventory below.

**Loose tag and goal loaders.** The loaders do not check for unknown keys or duplicate
IDs; a second record with the same `id` silently replaces the first (`E.002`, Tag and
goal data).

## Architecture Context

Pages read:
[architecture-overview](../../../architecture/architecture-overview.md),
[persistence](../../../architecture/engine/persistence.md),
[goals](../../../architecture/engine/goals.md),
[effect-mutation](../../../architecture/engine/effect-mutation.md).

Key Decisions that constrain this design:

- **GMs own all data; the binary embeds none** (persistence). No compiled-in
  defaults and no hard-coded data path. The compose step must not find `_core` from
  a path written into the binary.
- **`Paths` is tag-derived** (persistence). ACD-6 adds no `Paths` field; the campaign
  `rulebook/` folder keeps its current layout.
- **Rules validate at load** (persistence). `rulebook.Load` reads `tags.toml` and
  `goals.toml` with `toml.DecodeFile`. A missing file is a load error, so a campaign
  that misses the `_core` files fails loudly at startup.
- **Data-only goals and tags are silent skips** (goals, effect-mutation). Handlers
  key on the bare ID (`G-001`, `T-013`). A change to a name, description or effect
  string does not change what fires.
- **`goals.md` "Data-only goals are intentionally no-ops"** — the arc routed a wording
  fix here: the skip is an incremental seam, not a deliberate end state.

## Audit Findings

<br/>

### Compose surfaces

| Site | What it does today | Effect of removing goals/tags from `rulebooks/swn/` |
|---|---|---|
| `cmd/gm-toolkit/campaign/create.go:99` | `CopyRulebook(camp, rules, false)` with one `--rules` path | New campaign misses goals/tags unless `_core` is also copied |
| `cmd/gm-toolkit/campaign/add_rules.go:34` | `CopyRulebook(camp, source, replace)` with one `--rules` path | Same |
| `internal/faction/engine/testharness/harness.go:28` | `CopyRulebook` from `rulebooks/swn` into a temp campaign | Harness `rulebook.Load` fails |
| `internal/faction/rulebook/rulebook_test.go:12` | `Load("../../../rulebooks/swn")` directly (four calls); asserts tag `T-001` name `"Colonists"` and goal `G-001` | Fails; also pins tag text |
| `internal/faction/engine/core_test.go:29` | `rulebook.Load("../../../rulebooks/swn")` directly | Fails |
| `internal/faction/engine/testharness/scenarios/spot_check_test.go:29` | `Load` on `testDataDir` (`rulebooks/swn`, from `scenarios/helpers.go`) | Fails |
| `README.md:9` | Shows `--rules ./rulebooks/swn/` | Wrong instructions |

Three test sites call `rulebook.Load` on the repo source folder, not on a campaign.
ACD-6's "`rulebook.Load` never learns `_core` exists" holds for campaigns, but these
tests lose their data source. They need a composed folder.

### Existing campaigns

- `campaigns/sandbox/` and `campaigns/sandbox-claude/` hold their own copies of
  `goals.toml` and `tags.toml`. ACD-6 keeps the on-disk result the same, so they need
  no migration and keep running.
- **Both copies match the repo exactly** (checked with `diff`). They use only the keys
  `id`, `name`, `description`, `effect` and `difficulty`, and hold no duplicate IDs, so
  the strict loader (D-8) loads them without errors.
- **Saves hold goals by ID only.** `active_goal` stores `goal_id = "G-004"` and its
  start data, not goal text, so a goal text change needs no state migration.
- **`state.toml` embeds full tag copies.** `domain.Tag` carries `name`,
  `description` and `effect`, and `tag_added` writes the whole struct into the
  faction. Example: `campaigns/sandbox-claude/state/state.toml:25-29` holds the full
  T-011 text. The Manage detail view (`tui/views/manage/detail/detail.go:107`) shows
  the faction's copy. A re-seed of `rulebook/` does not change tags that factions
  already hold.

### Which IDs have Go behind them

| ID | Reached by |
|---|---|
| `G-001`…`G-012` | All twelve goal handlers (`engine/goal/goals/`); `G-004` and `G-012` also by their start actions |
| `T-005` Fanatical, `T-013` Preceptor Archive, `T-016` Scavengers, `T-020` Warlike | `TagEngine` handlers (`engine/tag/tags/`) |
| `T-011` Planetary Government | `domain.TagPlanetaryGovernment` (government record coupling) |
| The other 15 tags | Nothing — data-only. Their `effect` text is the only statement of the rule; the GM applies it |

Go identifiers carry tag names: `PreceptorArchiveHandler`,
`PreceptorArchiveCostModifier`, `PreceptorArchiveTagID`, and a comment in
`action/actions/ability/purchase_enabler.go:94`.

### Flavor inventory — tags

> Out of scope after D-7: tag text does not change. Kept as input for A.001.7 and for
> A.001.5's tag pass.

Classes: **SWN proper** (a named SWN organization, technology or term), **genre**
(space-setting words), **engine** (words that name a thing the engine models: world,
hex, tech level, `Starship` asset type, Base of Influence, Stealthed, Coin, action and
goal names), **neutral**.

| ID | Name | SWN-specific words | Class | Go |
|---|---|---|---|---|
| T-001 | Colonists | "planet"; "Spaceship-type assets" (the engine type is `Starship`) | genre, engine | — |
| T-002 | Deep Rooted | — | neutral | — |
| T-003 | Eugenics Cult | "maltech", "gengineered"; effect grants a "Gengineered Slaves" asset with inline stats that is not in any asset file | SWN proper, rule-bearing | — |
| T-004 | Exchange Consulate | name and description name the Exchange; effect names the Peaceable Kingdom goal | SWN proper | — |
| T-005 | Fanatical | — | neutral | handler |
| T-006 | Imperialists | "the sector" | genre | — |
| T-007 | Machiavellian | — | neutral | — |
| T-008 | Mercenary Group | — | neutral | — |
| T-009 | Perimeter Agency | name and description name the Perimeter | SWN proper | — |
| T-010 | Pirates | "scourge of the spacelanes" | genre | — |
| T-011 | Planetary Government | "planet" | engine | domain |
| T-012 | Plutocratic | — | neutral | — |
| T-013 | Preceptor Archive | name and description name the Preceptors; effect has the special action "Teach Planetary Population" that the engine does not have | SWN proper, rule-bearing | handler |
| T-014 | Psychic Academy | psychics; effect "psionic mentor training" | SWN proper (psionics), rule-bearing | — |
| T-015 | Savage | — | neutral | — |
| T-016 | Scavengers | — | neutral | handler |
| T-017 | Secretive | — | neutral | — |
| T-018 | Technical Expertise | "Starship-type assets", "planets" | engine | — |
| T-019 | Theocratic | — | neutral | — |
| T-020 | Warlike | — | neutral | handler |

### Flavor inventory — goals

All twelve are near-neutral. The only setting words are engine vocabulary: "planet",
"planetary government" (G-004, G-009), "hex distance" (G-012), "Base of Influence"
(G-005, G-012). The `difficulty` field is loaded and never read (only
`rulebook.go:619` touches it).

A goal record has no `effect` field (`id`, `name`, `description`, `difficulty`). The
goal `description` is the only statement of the goal's rule: thresholds ("equal to
your faction's Force rating"), streaks ("four consecutive turns") and gates ("must
increase before this goal can be selected again"). D-3 gains a goal row for this.

### Arc-plan text that does not match the repo

- The arc-plan said "strip goals/tags from `rulebooks/swn/` **and `rulebooks/hawk/`**".
  `rulebooks/hawk/` does not exist yet; A.001.5 creates it. Only SWN has files to
  remove. *(Fixed in the arc-plan by AP-8.)*
- The arc-discovery Out-of-scope list says "Goals copy verbatim", while ACD-7 names
  "goals/tags neutralization". *(Resolved by D-2: goals are neutralized. AP-8 records
  it in the arc-plan's Out-of-arc deferrals.)*

## Design Summary

`goals.toml` moves from `rulebooks/swn/` to a new shared folder, `rulebooks/_core/`.
`rulebooks/swn/rulebook.toml` names `../_core` as an include, and one compose step
copies both folders into a campaign's `rulebook/` folder. It checks every file before
it writes any (D-1). `tags.toml` stays in `rulebooks/swn/` (D-7). The goal text loses
its SWN words. Only nouns change, and each sentence keeps its rule exactly (D-2 to
D-5). The tag and goal loader becomes strict (D-8). The four test sites compose into a
temp folder through the real compose step (D-11), and both sandboxes are re-seeded
(D-12). The `goals.md` "intentional no-op" wording is corrected to describe an
incremental seam (arc reconciliation item). Bare IDs and every Go handler stay as
they are; prefixes and data-driven behavior belong to A.001.7 and A.001.8 (D-9, D-10).

### Compose and coverage

- **D-1 — A rulebook manifest names the folders it includes.** A rulebook folder
  may hold `rulebook.toml` with `includes = ["../_core"]`. Each path resolves
  relative to the rulebook folder, so the path lives in GM-owned data, not in the
  binary. `--rules` stays one path: `campaign create my-game --rules ./rulebooks/swn`.
  The compose result in `<campaign>/rulebook/` is the same as today, so
  `rulebook.Load` does not learn about `_core` or manifests. (D-8 makes it strict;
  that is a separate change.)
  - **One compose step, checked before any write.** Walk the rulebook folder and
    each included folder, then run every check, then copy. A refused compose writes
    nothing.
  - **A duplicate across folders is always refused.** If two composed folders supply
    the same relative path, the compose returns `ErrRulebookConflict` in every mode.
    `--replace` governs only a clash with a file already in the campaign. A GM who
    wants different text edits the campaign's copy.
  - **The manifest stays out of the campaign.** `rulebook.toml` is not copied.
  - **One level of includes.** An included folder's own manifest is not read.
  - **No manifest: a plain folder.** A `--rules` folder with no `rulebook.toml`
    copies as it does today. If it lacks goals or tags, `rulebook.Load` fails
    loudly at startup.
  - **Reversal.** An earlier draft of D-1 chose a repeatable `--rules`
    (`--rules _core --rules swn`), one `CopyRulebook` call per source. Rejected,
    for two reasons. Every rulebook needs `_core`, so each GM, the README and each
    test site would repeat it. Also, a second pass checks clashes only against files
    already written, so with `--replace` a stray rulebook `goals.toml` silently
    overwrites the core file, and a refused second pass leaves the campaign
    half-copied.
  - **Rejected:** a sibling-folder convention (an implicit layout rule); a separate
    `--core` flag (two flags for one job); manifest plus a repeatable `--rules` (two
    ways to do one thing); a rulebook file overriding an included file (a layering
    feature with no present user); a required manifest (breaks plain folders).
- **D-2 — Goals are in scope.** `goals.toml` goes through the same flavor line as
  `tags.toml`; it does not move to `_core` verbatim. This overrides the
  arc-discovery Out-of-scope line "Goals copy verbatim"; ACD-7's "goals/tags
  neutralization" stands.

### The flavor line

The line is a frame of two axes: the **field** sets how far text may change, and the
**word class** sets what happens to a word.

- **D-3 — Field latitude.**

  | Field | Job | Rule |
  |---|---|---|
  | Tag `description` | Fiction | Free rewrite to read setting-free |
  | `name` | The GM's handle; the UI label | May change; handlers key on the ID. Go names do not follow (D-10) |
  | Tag `effect` | The rule itself (the only statement of it for 15 of 20 tags) | **Nouns only.** Swap setting words; the sentence keeps its rule exactly. A clause whose rule depends on setting content goes to a per-tag rules call |
  | Goal `description` | The rule itself (goals have no `effect` field) | **Nouns only**, as tag `effect` |

  After D-7 only the goal rows apply in this initiative. The tag rows stay as the
  frame for A.001.5's tag pass.

- **D-4 — Word classes.**

  | Class | Examples | Treatment |
  |---|---|---|
  | Engine vocabulary | `Starship` asset type, tech level, world, hex, Base of Influence, Stealthed, action names (Seize Planet) | **Neutral paraphrase** (D-5) |
  | Genre words the engine does not model | spacelanes, sector, Spaceship-type, advanced technology | Drop, or swap for a setting-free word |
  | SWN proper nouns | Exchange Consulate, Perimeter Agency, Preceptor Archive, maltech, gengineered | **Case by case** (D-6) |
  | Setting content that carries a rule | T-003 Gengineered Slaves asset, T-013 Teach Planetary Population, T-014 psionic mentor training, T-001/T-018 population thresholds | Rules call per item (Robert) |

- **D-5 — Engine vocabulary is paraphrased, not quoted.** Shared text describes an
  engine term in setting-free words (for example "a ship-type asset" for `Starship`)
  instead of the exact UI label. Rejected: exact UI words, which keep SWN-sounding
  terms (`Starship`, "Seize Planet") in shared text. Cost accepted: the GM maps the
  paraphrase to the UI label.
- **D-6 — SWN proper nouns are replaced case by case.** No single rule (role names,
  invented names); each tag's replacement is decided when its text is drafted.
  Rejected: role names as a fixed rule, invented neutral names as a fixed rule.
  *After D-7 nothing in scope uses it:* the goals contain no SWN proper nouns.

### Setting context — H.A.W.K

H.A.W.K is magic-fantasy interstellar: magic as technology (in the style of
Eberron), travel between worlds by magic (in the style of Spelljammer), and hard,
rule-bound magic (in the style of Brandon Sanderson). Shared text must read
correctly in that setting and in SWN's hard sci-fi setting at the same time.

### Scope (resumed after `E.002`)

- **D-7 — Goals go to `_core`; tags stay in each rulebook.** `E.002`'s answer. Goals are
  the faction game's objective structure and near-neutral, so they are shared. Tags are
  setting identity; once behavior comes from data (A.001.7), an H.A.W.K tag states its
  behavior in its own entry and shares no record with SWN. Each folder supplies a whole
  file, so no merge-by-record loader is needed. `T-011` need not be core (A.001.7 makes
  the government role a data marker). The `T-003`, `T-013`, `T-014` and `T-001`/`T-018`
  rules calls leave scope. Amends ACD-6 and ACD-7 through AP-8.
  - **Rejected:** full core (the setting tags lose their identity, and each needs a
    rules call); the hybrid of core goals plus the neutral tags (Claude's earlier
    recommendation; it needs a record-level merge loader and `T-011` in core, and with
    data-driven tags no tag needs to be shared).
- **D-8 — The tag and goal loader becomes strict.** `rulebook.Load` fails on unknown
  keys and refuses a duplicate ID in `tags.toml` and `goals.toml`, as `loadAssets`
  does. This gives D-1 a load-time duplicate-ID check beside its file-level refusal.
  It is the one loader change, so ACD-6's "`rulebook.Load` is untouched" no longer
  holds for it (AP-8).
- **D-9 — Rulebook ID prefixes (`E.002` I-1) go to A.001.7 and A.001.8, not here.** Go
  handlers look up the bare IDs, so prefixes in data alone would silently stop the
  handlers. Prefixes here would change 16 handler keys, `domain.TagPlanetaryGovernment`
  and `B.001`'s literal, and need a save migration, which breaks ACD-7's data-only
  rule. A.001.7 and A.001.8 delete the ID-keyed handlers, so Go changes only once.
  - **Rejected:** prefixes here (`E.002` as filed); the loader and the prefixes both
    later (the loader is small and D-1 needs its duplicate check now).
- **D-10 — Go names do not follow renamed goal names.** If a goal name changes (for
  example "Planetary Seizure" under D-5), Go types and files (`PlanetarySeizure`,
  `planetary_seizure.go`) keep the old name until A.001.8 deletes them. This initiative
  changes data only.
  - **Rejected:** rename the Go types here (work that A.001.8 throws away); keep every
    goal name as it is (the `name` field would keep SWN words that D-5 removes).

### Delivery

- **D-11 — Test sites compose into a temp folder through the real compose step.**
  `rulebook_test.go`, `core_test.go`, `spot_check_test.go` (through
  `scenarios/helpers.go`) and the harness (`harness.go:28`) compose `rulebooks/swn`
  into `t.TempDir()`, then call `rulebook.Load`. The tests read data the same way a
  campaign does, and D-1's checks get test coverage. There is no import cycle:
  `internal/campaigns` imports no package from this repo.
  - **Rejected:** a composed copy checked in under `testdata/` (a second copy of the
    goals that can drift from `_core`).
- **D-12 — Both sandboxes are re-seeded.** Execution runs
  `campaign add-rules --rules ./rulebooks/swn --replace` on `sandbox` and on
  `sandbox-claude`, as the end-to-end check of the compose. Robert's call; it
  includes his `sandbox`. No state migration: saves hold `goal_id` only.
  - **Rejected:** `sandbox-claude` only; no re-seed.
- **D-13 — The neutral goal text is drafted in Plan, as literal strings.** The plan
  holds the final TOML text for each goal that changes, and Robert reviews each string
  at plan sign-off. The words to change are listed in the goals flavor inventory.
  - **Rejected:** drafting the strings in this Discovery.

## Open Questions

- **Slicing — for Plan.** Robert left phases and commits to the Plan session. Two
  facts constrain the order: the move of `goals.toml` breaks the four test sites
  until D-11 lands, and the re-seed (D-12) needs the compose step and the new goal
  text.

## User-Facing Impact

- **CLI:** no change to the command. `campaign create --rules ./rulebooks/swn` and
  `campaign add-rules --rules …` take one path, as today. The wizard field is free
  text (`create.go:124`), so `_core` is never offered as a choice.
- **New refusals:** a compose refuses a file that two folders both supply
  (`ErrRulebookConflict`, D-1). A rulebook with an unknown key or a duplicate tag or
  goal ID fails at load (D-8). The repo data and both sandboxes pass both checks
  today.
- **Goal text:** the GM sees the neutral goal names and descriptions in the TUI after
  a re-seed.
- **README:** the bootstrap command stays correct. One sentence says that
  `rulebooks/swn/` pulls the shared goals from `rulebooks/_core/`.

## Out of Scope

- **Tag text and tag behavior.** `tags.toml` does not change. Tags become data in
  A.001.7; H.A.W.K tags come from A.001.5's tag pass (AP-8).
- **Goal behavior.** No handler, threshold, XP or lock changes. The goal vocabulary,
  the turn-start goal pick, `B.005` and the goal rules calls from `E.002` (`G-004`
  XP, `G-009` count, `G-011` gate) belong to A.001.8.
- **Rulebook ID prefixes** (D-9) and **Go renames** (D-10): A.001.7 and A.001.8.
- **`rulebooks/hawk/`** and its manifest: A.001.5.
- **A merge-by-record loader** (not needed under D-7), **nested includes**, and **a
  rulebook file that overrides an included file** (both rejected in D-1).
