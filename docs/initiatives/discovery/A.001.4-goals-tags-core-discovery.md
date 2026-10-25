# Goals/Tags Neutralization + `_core` Home — Discovery

> **Initiative:** `A.001.4`, in the [data-driven-rulebook arc](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md).
> Binding arc decisions: ACD-6 (one shared home, deduped at scaffold) and ACD-7
> (neutralization is data-only), in the [arc-discovery](../arcs/data-driven-rulebook/data-driven-rulebook-arc-discovery.md).
> This doc does not re-open them. Its agenda is the one question the arc-plan routed
> here (the flavor-vs-mechanic line, and the compose mechanics), plus internal slicing.
>
> **Status:** paused, waiting on `E.002` (data-driven tags and goals). Findings and
> D-1 to D-6 are written. The scope question below ("Core, per-rulebook, or
> hybrid?") may change ACD-6 and ACD-7 and need an arc-plan amendment.

## In Plain Words

- **What:** move the goals and tags files out of the SWN rulebook folder into a
  shared folder, `rulebooks/_core/`. Campaign creation copies the shared folder and
  the chosen rulebook into the campaign. Then remove the SWN setting words from the
  goal and tag text, so the same files read correctly for SWN and for H.A.W.K.
- **Why:** goals and tags are the same rules in every game this tool runs. Today they
  live inside the SWN folder and use SWN names (for example "Preceptor Archive"). A
  second rulebook would have to copy them, or carry SWN fiction into a different game.
- **How it fits:** this initiative has no dependency on the ability and effect
  registries. It can run at any time in the arc. A playable H.A.W.K campaign needs it,
  because H.A.W.K gets its goals and tags from the shared folder.

## Problem

<br/>

**One source folder per campaign.** `campaigns.CopyRulebook`
(`internal/campaigns/scaffold.go:50`) walks one source folder and copies every
`.toml` into `<campaign>/rulebook/`. Its two CLI callers (`campaign create --rules`,
`campaign add-rules --rules`) take one path. `rulebooks/swn/` holds the goals and
tags, so a second rulebook has no shared source to copy them from.

**SWN setting words in shared rules.** `tags.toml` names SWN organizations, SWN
technology and SWN space terms. A neutral rule must read correctly in both games.
`goals.toml` is almost neutral already. See the inventory below.

## Architecture Context

Pages read:
[architecture-overview](../../architecture/architecture-overview.md),
[persistence](../../architecture/engine/persistence.md),
[goals](../../architecture/engine/goals.md),
[effect-mutation](../../architecture/engine/effect-mutation.md).

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

### Arc-plan text that does not match the repo

- The arc-plan says "strip goals/tags from `rulebooks/swn/` **and `rulebooks/hawk/`**".
  `rulebooks/hawk/` does not exist yet; A.001.5 creates it. Only SWN has files to
  remove.
- The arc-discovery Out-of-scope list says "Goals copy verbatim", while ACD-7 names
  "goals/tags neutralization". The inventory shows the goals need no change under any
  line that keeps engine vocabulary, so the two may agree in practice. To confirm.

## Design Summary

_In progress. Decisions so far:_

- **D-1 — A rulebook manifest names the folders it includes.** A rulebook folder
  may hold `rulebook.toml` with `includes = ["../_core"]`. Each path resolves
  relative to the rulebook folder, so the path lives in GM-owned data, not in the
  binary. `--rules` stays one path: `campaign create my-game --rules ./rulebooks/swn`.
  The compose result in `<campaign>/rulebook/` is the same as today, so
  `rulebook.Load` does not change.
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
  | `description` | Fiction | Free rewrite to read setting-free |
  | `name` | The GM's handle; the UI label | May change; handlers key on the ID. Go follow-through is open (see Open Questions) |
  | `effect` | The rule itself (the only statement of it for 15 of 20 tags) | **Nouns only.** Swap setting words; the sentence keeps its rule exactly. A clause whose rule depends on setting content goes to a per-tag rules call |

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

### Setting context — H.A.W.K

H.A.W.K is magic-fantasy interstellar: magic as technology (in the style of
Eberron), travel between worlds by magic (in the style of Spelljammer), and hard,
rule-bound magic (in the style of Brandon Sanderson). Shared text must read
correctly in that setting and in SWN's hard sci-fi setting at the same time.

## Open Questions

- **Core, per-rulebook, or hybrid? (scope; waits on `E.002`).** Robert's eventual
  goal is that tags and goals become data-driven, like assets. That raises three
  shapes:
  - **No core** — each rulebook defines its own flavorful tags and goals. Works
    only once behavior comes from data: today, behavior keys on bare IDs, so a
    rulebook gets Warlike's behavior only by reusing `T-020`.
  - **Hybrid (Claude's recommendation)** — `_core` holds the goals and the
    tags that are already neutral; each rulebook adds its own setting tags. SWN keeps
    Exchange Consulate, Perimeter Agency, Preceptor Archive, Psychic Academy and
    Eugenics Cult unchanged, so the T-003/T-013/T-014 rules calls leave scope.
    `T-011` must be core (G-004 grants it; the government record couples to it).
    Merge by record, using the asset pattern: `rulebook.Load` globs
    `*_tags.toml` / `*_goals.toml` and refuses a duplicate ID (`_core/core_tags.toml`
    + `swn/swn_tags.toml`). A small loader change that ACD-6 ruled out, but the
    pattern already exists. D-1's file-level refusal still holds because the file
    names differ.
  - **Full core** — the current ACD-6/ACD-7 design: every rulebook composes the
    same neutralized set.

  Reasoning so far: goals are the faction game's objective structure and
  near-neutral, so they probably stay core. Tags are setting identity, with 5 small
  hook bundles that may map onto `A.001.3`'s `[[effects]]` at faction scope (not
  verified). **Cost is not a criterion:** Robert wants goals data-driven too, for one
  consistent way to handle data files and a project-wide move away from bespoke
  handlers. Core-vs-rulebook is decided on fit, not on what data-driving costs. Open risk under
  hybrid or no-core: a rulebook tag that reuses a bare ID (`T-013`) silently gets
  that handler; rulebook-prefixed IDs (`SWN-T-013`) or data-driving would close it.
  Hybrid or no-core amends ACD-6 (core becomes a shared base that rulebooks add to)
  and narrows ACD-7 to what moves into core.
- **Go follow-through for renamed names.** Go identifiers, file names and comments
  carry tag and goal names (`PreceptorArchiveHandler`, `PreceptorArchiveTagID`, the
  goal handler types). Robert wants tags and goals to become data-driven, which would
  make this concern smaller. An exploration of how much that work adds to this
  initiative runs after this discovery; the answer here waits on it.
