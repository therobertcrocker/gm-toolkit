# H.A.W.K Asset Authoring — Overview Plan

## In Plain Words

- **What** — The plan for six efforts that give H.A.W.K its own canon, tags, campaign and assets. Effort 1 is written here in full: scaffold the H.A.W.K campaign, then, with Robert's setting docs as background, interview him through two fill-in forms to seed its written canon. Efforts 2 to 6 each get their own Plan session, because each one designs from the canon Effort 1 writes.
- **Why** — The engine runs every asset, tag and goal from rulebook data, but only SWN has ever used it. H.A.W.K is the campaign the tool exists for, and today it has no rulebook, no campaign and no written canon.
- **How it fits** — The last initiative of arc `A.001`. The order is lore → tags → campaign setup → Force → Cunning → Wealth. From Effort 3 onward, each effort ends with Robert playing a live turn in the H.A.W.K campaign. After Effort 6, H.A.W.K is playable and the arc close opens.

## Context / Goal

Turn the discovery's six efforts into an executable sequence: specify Effort 1 to the task, and scope Efforts 2 to 6 for their own Plan sessions.

- Discovery doc: [`A.001.5-hawk-asset-authoring-discovery.md`](../discovery/A.001.5-hawk-asset-authoring-discovery.md)
- Arc plan: [`A.001-data-driven-rulebook-arc-plan.md`](../arcs/A.001-data-driven-rulebook/A.001-data-driven-rulebook-arc-plan.md), section "A.001.5 — H.A.W.K asset authoring", AP-10 and AP-11

## Decisions Ratified in Planning

1. **Effort 3 gets its own Plan session, after Effort 2** — Partly reverses D-9, which had E3 run straight from this plan. Every file E3 writes holds fiction from E1: `layout.txt`, the places and their populations, the factions with their stats and homeworlds, and the starting assets. Written now, its tasks could only be prose. Its Plan session drafts the map, the seed state and the settings as literal content, and Robert's sign-off is the review. It comes after E2 because factions take H.A.W.K tags at setup (D-3). Rejected: E3 as a procedure in this plan, with the data drafted live during execution.
2. **Effort 1 runs as two phases, one lore form each** — Phase 1: the scaffold, the canon skeleton, the campaign premise, the setting and the glossary (Form 1). Phase 2: places, factions and people (Form 2). The setting has to exist before the factions and places can sit in it, so Form 2 goes to Robert only once Form 1 is in the canon. Rejected: one form for everything, in one phase or two.
3. **A canon heading carries its ID: `## Iron Concord — \`iron-concord\``** — Factions and places use the engine IDs they get in E3 (the seed state's faction ID, the world `id` in `data.toml`), so the canon, the map and the seed join on one key. Tags and assets use their `HAWK-` IDs (D-7). Everything else (the map, regions, terms, people) uses a kebab-case slug of its name. This settles the form of D-5's "stable ID". Rejected: the ID on its own body line; a separate canon ID scheme that needs a mapping later.
4. **Robert runs each live turn** — D-1 says Robert opens the campaign and runs a turn, and Claude drives the app only on `sandbox-claude`. Claude prepares the turn (build, compose, map build, seed reset). Robert plays it as the effort's last step and reports what he saw. Rejected: Claude playing `campaigns/hawk` in tmux; a gitignored `hawk-claude` copy.
5. **Every campaign command carries `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns`** — The registry lives at `$GM_TOOLKIT_HOME/campaigns-registry.toml` (`internal/campaigns/registry.go`), and Claude's shell does not set the variable. Writing the prefix every time makes each command work the same in any shell.
6. **Efforts 2 to 6 each get a plan file, `A.001.5.E<n>-hawk-asset-authoring-plan.md`. A pass with no build rows skips Phase 1** — D-9 gives each pass a Phase 1 for build rows only (primitive code and tests) and a Phase 2 for the data. When a pass's sort table has no build row, the effort plan has a single data phase. E3 is not a pass, and its plan sets its own phases.
7. **The lore interview is a Markdown form Robert fills in his editor. Robert's setting docs come first, as context only. Forms and docs live in gitignored `scratch/` and are deleted once the canon holds the facts** — Robert's calls: long answers are easier to write in a file than in chat, and the docs help Claude follow the answers without standing in for them. Each form's full text is in this plan, so the questions are fixed before the session starts. Claude reads the docs, then the filled form, asks follow-ups in chat about blanks and contradictions, then writes the canon. Canon text comes only from Robert's form answers and chat replies, never straight from the docs. The canon is the only kept copy of the fiction (D-5); `scratch/` is already in `.gitignore`, so `git add -A` stays safe. Rejected: an Artifact web form (the answers go through claude.ai and must be read back); committing the forms or the docs (a second copy of the fiction that drifts); pre-filling form answers from the docs.
8. **The campaign premise gets its own canon file, `campaign.md`** — Robert's call. It covers the PF2e campaign (the party, the story so far) and what the faction game does for it, which is about the campaign, not the world. Amends the discovery's Canon Layout by one file; `setting.md` stays about the world. Rejected: a `## Campaign` section at the top of `setting.md`.
9. **The forms ask in H.A.W.K's own terms and show the engine's fixed shape openly** — H.A.W.K is magic as technology, a fantasy and sci-fi mix (Eberron, Mistborn, Stormlight Archive, Spelljammer), not SWN's space setting. The engine still fixes the map's shape (hexes, regions, places it calls worlds, links between regions) and each faction's (one homeworld, three stats). Form 2 states that shape in a short box, then asks what plays each part in H.A.W.K. Answers stay descriptive ("a city", "best at guile"); E3 turns them into tech levels, populations and stats. Each open question carries two or three prompts, so no single answer has to be a whole document. From a review of the first form drafts, which assumed a space setting and SWN numbers.

## Decision Record — Execution

## Effort Summary

| Done | # | Effort | Scope | File |
|------|---|--------|-------|------|
| [ ] | 1 | Lore seed | Scaffold `campaigns/hawk/`, then the docs and two forms seed the canon at `campaigns/hawk/canon/` | This plan, Phases 1–2 |
| [ ] | 2 | Tag pass | `rulebooks/hawk/` begins: `rulebook.toml`, H.A.W.K `tags.toml`, SWN stand-ins for the asset files and `settings.toml` | `A.001.5.E2-hawk-asset-authoring-plan.md` (planned) |
| [ ] | 3 | Campaign setup | The map, the starting factions as a seed state, the real `settings.toml`, the first live turn | `A.001.5.E3-hawk-asset-authoring-plan.md` (planned) |
| [ ] | 4 | Force pass | `assets/force_assets.toml` replaces its stand-in; a live turn | `A.001.5.E4-hawk-asset-authoring-plan.md` (planned) |
| [ ] | 5 | Cunning pass | `assets/cunning_assets.toml` replaces its stand-in; a live turn | `A.001.5.E5-hawk-asset-authoring-plan.md` (planned) |
| [ ] | 6 | Wealth pass | `assets/wealth_assets.toml` replaces the last stand-in; "H.A.W.K playable" | `A.001.5.E6-hawk-asset-authoring-plan.md` (planned) |

The order is a straight line; each effort depends on the one before it:

```
E1 lore → E2 tags → E3 campaign setup → E4 Force → E5 Cunning → E6 Wealth
```

## Shared Context

### Branch and git cadence

- Branch `feature/hawk-asset-authoring`, cut when Discovery started. Every effort session runs on it, Plan sessions included.
- Every commit for effort n carries the scope `E<n>`, its plan doc included (`docs(E2): ...`).
- At each effort boundary, `/gestalt:effort-close A.001.5.E<n>` runs the interim review, one `fix(E<n>):` commit and the squash. Merge to `main` with `--no-ff` at the initiative close.
- Campaign data is committed like code. Play state (`state.toml`, `history.jsonl`, `logs/`) is gitignored; `state.seed.toml` is tracked (D-10).

### Campaign commands

Run from the repo root, with the binary built first:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
```

Every `gm-toolkit` command that touches a campaign takes the prefix `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns` (Decision 5). `campaigns/campaigns-registry.toml` is tracked and stores absolute paths. The active campaign stays `sandbox`; H.A.W.K commands name it with `--campaign hawk`.

### Stand-ins by effort

Which composed file is H.A.W.K's own and which is still an SWN copy after each effort (discovery, Campaign and Rulebook Mechanics). Goals always come from `_core` through the `rulebook.toml` include.

| After | `tags.toml` | `settings.toml` | `force_assets.toml` | `cunning_assets.toml` | `wealth_assets.toml` |
|-------|-------------|-----------------|---------------------|-----------------------|----------------------|
| E2 | H.A.W.K | SWN copy | SWN copy | SWN copy | SWN copy |
| E3 | H.A.W.K | H.A.W.K | SWN copy | SWN copy | SWN copy |
| E4 | H.A.W.K | H.A.W.K | H.A.W.K | SWN copy | SWN copy |
| E5 | H.A.W.K | H.A.W.K | H.A.W.K | H.A.W.K | SWN copy |
| E6 | H.A.W.K | H.A.W.K | H.A.W.K | H.A.W.K | H.A.W.K |

A tag–asset link is written by the pass that creates its second half (D-11).

### Live turn (Efforts 3 to 6)

The last step of each effort from E3 on. Claude runs steps 1 to 4; Robert runs step 5 (Decision 4).

1. `go build -o gm-toolkit ./cmd/gm-toolkit`
2. `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign add-rules --rules rulebooks/hawk --campaign hawk --replace`
3. Only when `campaigns/hawk/spatial/source/` changed: `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit spatial build --campaign hawk --replace` (E3's first build leaves out `--replace`)
4. `./cmd/scripts/reset-campaign.sh hawk`
5. Robert: `GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit faction --campaign hawk`, plays one turn, and reports back

An asset pass first swaps its category's starting assets in `campaigns/hawk/state/state.seed.toml` from the SWN stand-in IDs to the new `HAWK-` IDs, then resets (D-10). An asset instance ID is `<faction>-<definition>-<n>`, so the swap changes the instance ID as well as `definition_id`.

### Pass shape (Efforts 2, 4, 5, 6)

Each pass's Plan session runs lore-gate → reflavor → mechanical sort, and Robert's sign-off is the review (D-9). Before sorting, it reads today's primitives from the registrar tables (`effect/registrar.go`, `effects/registrars.go`, ability dispatch), not from the ability catalog page (D-8). The sort table uses one shape in every effort plan (D-8):

| HAWK ID | Name | SWN ancestor | Surface | Call | Primitive | Notes |
|---------|------|--------------|---------|------|-----------|-------|

`Call` is `reuse`, `parameterize` or `build`. The fiction for each row is drafted in the effort plan and written to `canon/tags.md` or `canon/assets.md` in the data phase. Phases follow Decision 6.

## Out of Scope

The discovery's Out of Scope list holds, unchanged. In addition:

- **A headless load check** — `faction --dryrun` ignores the campaign (`cmd/gm-toolkit/faction.go`). The live turn is the load check; building a per-campaign dry run is not this initiative's work.
- **Making `hawk` the active campaign** — `sandbox` stays active; H.A.W.K commands pass `--campaign hawk`.
- **A H.A.W.K test guide like `campaigns/sandbox/TESTING.md`** — the live turn is one turn of normal play, not a scripted scenario.

---

## Work Breakdown

### Effort 1 — Lore seed

**Scope:** discovery D-6 and the Canon Layout, plus `campaign.md` (Decision 8). Scaffold `campaigns/hawk/` with `campaign create` and write the canon skeleton. Robert hands over his setting docs as context, then fills two lore forms (Decisions 7 and 9), and Claude writes his answers into the canon. Form 2 gathers the facts E3 needs for the map and the starting factions. Tag and asset fiction waits for the passes.

**Depends on:** nothing in this initiative.

**Own Plan session:** no. Phases 1 and 2 below.

**Overview**

- [ ] **Phase 1 — Scaffold, campaign and setting**
  - Commit 1 — `feat(E1): scaffold the H.A.W.K campaign and canon` (3 tasks)
  - Commit 2 — `docs(E1): seed the H.A.W.K campaign, setting and glossary` (7 tasks)
- [ ] **Phase 2 — Places, factions and people**
  - Commit 3 — `docs(E1): seed the H.A.W.K places, factions and people` (7 tasks)

The canon tasks can't hold the fiction as literal text, because the fiction is what the session produces. What is literal is the form Robert fills in and the canon entry form each answer goes into. Each interview commit runs:

```
read the docs → write the form → Robert fills it → Claude reads it
  → follow-ups in chat → transcribe into the canon → delete the form → commit
```

Claude transcribes Robert's words, tidied into plain prose, and never fills a gap with fiction of its own or with text lifted from the docs. A blank the follow-ups don't settle is written as `TBD` and listed in the commit's wrap-up. A blank ID line means Claude proposes a kebab-case ID and Robert confirms it in chat (Decision 3).

### Phase 1 — Scaffold, campaign and setting

**In plain words:** Create the H.A.W.K campaign folder and the empty canon files, read Robert's setting docs, then have him fill in a form on the campaign, the setting and its words. When it lands, the campaign is registered and `campaign.md`, `setting.md` and `glossary.md` hold real canon.

#### Commit 1 — `feat(E1): scaffold the H.A.W.K campaign and canon`

##### Task 1 — scaffold with `campaign create`

From the repo root:

```sh
go build -o gm-toolkit ./cmd/gm-toolkit
GM_TOOLKIT_HOME=/workspaces/gm-toolkit/campaigns ./gm-toolkit campaign create hawk --path campaigns --name "H.A.W.K"
```

Expected output:

```
Created campaign "hawk" at /workspaces/gm-toolkit/campaigns/hawk
rulebook/ is empty — seed it with 'gm-toolkit campaign add-rules --rules <path>' or copy TOML manually before running the toolkit
```

No `active campaign:` line, because `sandbox` is already active and `--activate` is left out. Results:

- `campaigns/hawk/campaign.toml`:

  ```toml
  [campaign]
    id = "hawk"
    name = "H.A.W.K"
  ```

- `campaigns/campaigns-registry.toml` gains, after the `sandbox-claude` entry:

  ```toml
  [[campaigns]]
    id = "hawk"
    path = "/workspaces/gm-toolkit/campaigns/hawk"
  ```

- Empty folders from `Paths()`: `rulebook/assets/`, `spatial/`, `state/logs/`, `narratives/`. Git does not track them; later efforts fill the ones they need.

##### Task 2 — `campaigns/hawk/canon/README.md`

`campaigns/hawk/canon/README.md` (new) — full contents:

```markdown
# H.A.W.K Canon

The written canon of the H.A.W.K campaign and its setting: the fiction behind its factions, places, tags and assets. The campaign's rulebook and map are built from it. It is plain Markdown, so the Codex can take it over later.

## Files

| File | Holds |
|------|-------|
| [campaign.md](campaign.md) | The campaign, the story so far, and what the faction game does in it |
| [setting.md](setting.md) | Pitch, conflicts, history, tone, magic as technology, travel and scale |
| [glossary.md](glossary.md) | H.A.W.K terms, one entry per term |
| [places.md](places.md) | The map, its regions and its places |
| [factions.md](factions.md) | One entry per faction |
| [people.md](people.md) | Named characters |
| [tags.md](tags.md) | Tag fiction, one entry per `HAWK-T-` ID |
| [assets.md](assets.md) | Asset fiction, one entry per `HAWK-` asset ID |

## Entries

`campaign.md` and `setting.md` are prose under topic headings. Every other file holds one `##` entry per entity. The heading is the name, an em dash, and the entity's ID in backticks:

    ## Iron Concord — `iron-concord`

- **Factions** — the faction ID in `state/state.seed.toml`
- **Places** — the world `id` in `spatial/source/data.toml`
- **Tags and assets** — the `HAWK-` ID in the rulebook (`HAWK-T-001`, `HAWK-F1-001`)
- **Everything else** (the map, regions, terms, people) — a kebab-case slug of the name

An ID never changes once committed, even if the name does. The ID of a cut entity is never reused.

## Canon and rulebook

The canon entry holds an element's full fiction. A rulebook `description` summarizes the entry and is never its only copy. When a rulebook entry gains new fiction, its canon entry gains it first.
```

##### Task 3 — the eight topic files

Each file is new, with exactly these contents:

`campaigns/hawk/canon/campaign.md`:

```markdown
# H.A.W.K Campaign

The campaign, the story so far, and what the faction game does in it.
```

`campaigns/hawk/canon/setting.md`:

```markdown
# H.A.W.K Setting

Pitch, conflicts, history, tone, magic as technology, travel and scale.
```

`campaigns/hawk/canon/glossary.md`:

```markdown
# H.A.W.K Glossary

H.A.W.K terms, one entry per term.
```

`campaigns/hawk/canon/places.md`:

```markdown
# H.A.W.K Places

The map, its regions and its places.
```

`campaigns/hawk/canon/factions.md`:

```markdown
# H.A.W.K Factions

One entry per faction.
```

`campaigns/hawk/canon/people.md`:

```markdown
# H.A.W.K People

Named characters, one entry per person.
```

`campaigns/hawk/canon/tags.md`:

```markdown
# H.A.W.K Tags

The fiction behind each faction tag, one entry per `HAWK-T-` ID.
```

`campaigns/hawk/canon/assets.md`:

```markdown
# H.A.W.K Assets

The fiction behind each asset, one entry per `HAWK-` asset ID.
```

##### Commit message

```
feat(E1): scaffold the H.A.W.K campaign and canon

- campaign create hawk: campaigns/hawk/ with its manifest, registered
  in campaigns-registry.toml; sandbox stays active
- canon/README.md: the file index, the entry ID rule, and how the
  canon relates to rulebook descriptions
- the eight canon topic files, title and scope only
```

#### Commit 2 — `docs(E1): seed the H.A.W.K campaign, setting and glossary`

##### Task 1 — read Robert's setting docs

Ask Robert to put his setting docs in `scratch/A.001.5.E1-sources/` (gitignored), in any format, and to say when they are there. Read them. A PDF, or a doc too large to read whole, goes to a side agent that writes `scratch/A.001.5.E1-sources/_brief.md`: the main facts about the campaign and setting, each with the doc it came from. Phase 2 reads the brief again instead of the full docs.

The docs are context only (Decision 7). They help Claude follow Robert's answers and ask sharper follow-ups. No canon text comes from them: a fact found only in the docs reaches the canon only when Robert writes it in a form or says it in chat.

##### Task 2 — `scratch/A.001.5.E1-lore-form-1.md`

`scratch/A.001.5.E1-lore-form-1.md` (new, gitignored) — full contents:

```markdown
# H.A.W.K Lore Form 1 — Campaign, Setting and Glossary

Write your answers under each question, as long or as short as you like. The prompts under a question are there to help; answer the ones that matter and skip the rest. Leave a question blank to skip it; Claude asks about the blanks before writing the canon. Save the file and tell Claude when you're done.

## The campaign

### The campaign

What is the PF2e campaign about? Who is the party, and what drives them?



### The story so far

Where does the story stand now? What were the big turning points that got it there?



### The faction game

- What is the faction game for in this campaign: background that moves between sessions, a source of plot hooks, something the party takes part in, or something else?
- When in the story does it start?
- How long is one faction turn in story time?



## The setting

### Pitch

The setting in two or three sentences. If H.A.W.K is an acronym, what does it stand for?



### Main conflicts

What are the big tensions in the setting? Who stands on each side?



### History

What past events still shape things now?



### Tone

What does play feel like? What is out of bounds?



### Magic as technology

- What does magic power: industry, war, travel, daily life?
- Who controls it, and what limits it?
- How much does that vary from place to place?



### Travel

How do people, goods and armed forces get from place to place? What makes a route slow, fast or dangerous?



### Scale

How big is the area the faction game covers? Roughly how many places and how many factions matter in it?



## Glossary

H.A.W.K's own words, one line per term: the term, a dash, what it means.

### The faction game's concepts

H.A.W.K's word for each, if it has one. Leave the line empty if not.

- Faction —
- Asset —
- Tag —
- Base of influence —
- Homeworld —
- Coin —
- Force —
- Cunning —
- Wealth —
- Faction turn —

### Other terms

Any word the setting uses that a newcomer would trip on.

-
```

Then tell Robert the path and wait until he says the form is filled. Read it, list the blank questions and anything that contradicts itself or the docs, and settle those in chat before Task 3.

##### Task 3 — transcribe: `campaigns/hawk/canon/campaign.md`

Below the file's scope line, one `##` section per question under "The campaign", in this order: `## The Campaign`, `## The Story So Far`, `## The Faction Game`. Each holds Robert's answer as plain prose, plus what the follow-ups settled. A section with no answer holds `TBD`.

```markdown
## The Campaign

<Robert's answer>

## The Story So Far

<Robert's answer>

## The Faction Game

<Robert's answer>
```

The Faction Game section keeps the start point and the turn length as plain facts, because later efforts read them.

##### Task 4 — transcribe: `campaigns/hawk/canon/setting.md`

Below the file's scope line, one `##` section per question under "The setting", in this order: `## Pitch`, `## Main Conflicts`, `## History`, `## Tone`, `## Magic as Technology`, `## Travel`, `## Scale`. Same form and `TBD` rule as Task 3.

Magic as Technology feeds E3's tech levels, and Travel feeds its regions, links and drift costs. Keep every concrete fact in them, such as limits, named routes and named sources of power.

##### Task 5 — transcribe: `campaigns/hawk/canon/glossary.md`

One entry per term from the form's Glossary, sorted A to Z. A faction-game concept with an empty line gets no entry.

```markdown
## <Term> — `<kebab-case-slug>`

<One-sentence definition.>

<Optional: more fiction, and the faction-game concept it maps to, if any.>
```

##### Task 6 — delete the form

Delete `scratch/A.001.5.E1-lore-form-1.md` once Robert has read the canon from Tasks 3 to 5 and is happy with it. Keep `scratch/A.001.5.E1-sources/` for Phase 2. Both are gitignored, so nothing here touches the commit.

##### Task 7 — this plan

In the Work Breakdown Overview, tick Phase 1: `- [ ] **Phase 1 — Scaffold, campaign and setting**` becomes `- [x] **Phase 1 — Scaffold, campaign and setting**`.

##### Commit message

```
docs(E1): seed the H.A.W.K campaign, setting and glossary

- campaign.md: the campaign, the story so far, and what the faction
  game does in it, from lore form 1
- setting.md: pitch, conflicts, history, tone, magic as technology,
  travel and scale
- glossary.md: H.A.W.K terms, including its words for the faction
  game's concepts
```

### Phase 2 — Places, factions and people

**In plain words:** Have Robert fill in a second form on H.A.W.K's map and places, its factions and any named characters, collecting everything the campaign setup needs to build the map and the starting factions. When it lands, the canon is seeded and Effort 1 is done.

#### Commit 3 — `docs(E1): seed the H.A.W.K places, factions and people`

##### Task 1 — re-read the context

Read `campaigns/hawk/canon/campaign.md` and `setting.md`, then `scratch/A.001.5.E1-sources/_brief.md` if Phase 1 wrote one, or else the docs in `scratch/A.001.5.E1-sources/`. The docs stay context only (Decision 7).

##### Task 2 — `scratch/A.001.5.E1-lore-form-2.md`

Places come before factions in the form, because a faction's seat of power has to name a place that's already described.

`scratch/A.001.5.E1-lore-form-2.md` (new, gitignored) — full contents:

```markdown
# H.A.W.K Lore Form 2 — Places, Factions and People

Write your answers under each question, as long or as short as you like. Copy a block once per region, place, faction or person. The prompts are there to help; answer the ones that matter and skip the rest. Leave a line blank to skip it; Claude asks about the blanks before writing the canon. A blank ID means Claude proposes one for you to confirm. Save the file and tell Claude when you're done.

## What the engine needs

The faction game's map and factions have a fixed shape. Answer in H.A.W.K's own terms; Effort 3 fits your answers to this shape and turns them into numbers.

- **The map** — one grid of hexes with one name. The hexes group into named regions.
- **Places** — each sits on one hex inside a region. The engine calls them worlds, but a place can be whatever H.A.W.K needs: a planet, a city, a sky-island. Each gets a tech level from 0 to 5 and a population.
- **Travel** — crossing the map hex by hex is slow. Fast links join the edges of regions across open space.
- **Factions** — each has exactly one seat of power, a place on the map, and may control other places. Each gets a rating in Force, Cunning and Wealth.

## The map

### Name and shape

What is the area called? What is its shape, and what lies beyond its edges?



### Layout

- Where do the regions sit relative to each other?
- Which regions are joined by fast links, and which are cut off or slow to reach?
- Where are the empty stretches?

Plain sentences are fine; Claude turns them into the map in Effort 3.



## Regions

One block per region.

### Region: <name>

What sets it apart from its neighbours?



## Places

The places that matter when the faction game opens. One block per place.

### Place: <name>

- **ID** — (kebab-case; becomes the place's ID on the map)
- **Region** —
- **Magic and technology** — (how advanced it is, compared with other places)
- **Size** — (outpost, town, city, nation, …)

What is it known for? Who lives there? What is contested about it?



## Factions

The factions in play when the faction game opens. One block per faction.

### Faction: <name>

- **ID** — (kebab-case; becomes the faction's ID in the campaign)
- **Size and strength** — (compared with the other factions)
- **Best and worst at** — (force, guile or money)
- **Seat of power** — (one place)
- **Controls** — (other places it holds outright, or none)
- **What sets them apart** — (a trait, a method, something only they have)

Who are they, what do they want, how do they act, and who do they clash with?



## People

People who matter to the factions' moves: leaders, agents, figures from the history the factions still invoke. One block per person; delete the section if none.

### Person: <name>

- **Faction** —
- **Place** —

Who are they, and what do they want?


```

Then tell Robert the path and wait until he says the form is filled. Read it and settle in chat: blank lines, IDs to propose, contradictions with Phase 1's canon, and the cross-checks in Task 4.

##### Task 3 — transcribe: `campaigns/hawk/canon/places.md`

The map entry comes first, then each region followed by its places:

```markdown
## <Map name> — `<slug>`

<The map's fiction.>

**Layout** — <where the regions sit, the fast links, the empty stretches, in plain sentences>

## <Region name> — `<slug>`

<The region's fiction.>

## <Place name> — `<world-id>`

**Region** — `<region-slug>`
**Magic and technology** — <how advanced, compared with other places>
**Size** — <outpost, town, city, nation, …>

<The place's fiction.>
```

A place ID is kebab-case and becomes the world's `id` in `data.toml` (Decision 3), so Robert confirms every one. The form's Layout answer goes in the map entry's `**Layout**` line, kept whole; E3 builds `layout.txt` from it.

##### Task 4 — transcribe: `campaigns/hawk/canon/factions.md`

One entry per faction block in the form:

```markdown
## <Faction name> — `<faction-id>`

**Size and strength** — <compared with the other factions>
**Best and worst at** — <force, guile or money>
**Seat of power** — `<world-id>`
**Controls** — <world IDs, or none>
**What sets them apart** — <a trait, a method, something only they have>

<The faction's fiction.>
```

These fields are what E3 reads. It sets the scale and the three stats from `Size and strength` and `Best and worst at`, places the homeworld base at `Seat of power`, and works out the government record from `Controls`. E2 reads `What sets them apart` when it drafts the tags. The faction ID is kebab-case and becomes the faction's key in the seed state (Decision 3), so Robert confirms every one. Cross-checks, settled in chat before writing: every `Seat of power` and `Controls` names a place entry from Task 3, and a place two factions both claim to control is raised with Robert.

##### Task 5 — transcribe: `campaigns/hawk/canon/people.md`

One entry per person block in the form. With no blocks, the file keeps its scope line only.

```markdown
## <Name> — `<slug>`

**Faction** — `<faction-id>`, or none
**Place** — `<world-id>` or `<region-slug>`, if they are tied to one

<The person's fiction.>
```

##### Task 6 — delete the form and the docs

Delete `scratch/A.001.5.E1-lore-form-2.md` and the `scratch/A.001.5.E1-sources/` folder once Robert has read the canon from Tasks 3 to 5 and is happy with it. Both are gitignored, so the deletion is not part of the commit.

##### Task 7 — this plan

- In the Work Breakdown Overview, tick Phase 2: `- [ ] **Phase 2 — Places, factions and people**` becomes `- [x] **Phase 2 — Places, factions and people**`.
- In the Effort Summary, tick Effort 1: the row `| [ ] | 1 | Lore seed |` becomes `| [x] | 1 | Lore seed |`.

##### Commit message

```
docs(E1): seed the H.A.W.K places, factions and people

- places.md: the map and its layout, its regions, and its places
  with their magic and technology and their size, from lore form 2
- factions.md: the starting factions with size and strength,
  best and worst at, seat of power, controlled places, and what
  sets them apart
- people.md: named characters, if the form turned any up
```

### Effort 2 — Tag pass

**Scope:** discovery D-11 and the Effort Breakdown.
- `rulebooks/hawk/rulebook.toml` with `includes = ["../_core"]`.
- `rulebooks/hawk/tags.toml` in `A.001.7`'s data form, with `HAWK-T-<nnn>` IDs (D-7). A tag effect that needs an `asset_definition` filter on a H.A.W.K asset is left out until that asset's pass (D-11).
- SWN copies of the three asset files and `settings.toml` as stand-ins. In the Force copy, `requires_tag` is removed from Gengineered Slaves (D-11).
- The tag fiction in `canon/tags.md`.
- Composed into `campaigns/hawk`. The campaign has no map or seed yet, so the effort plan decides how E2 checks that the rulebook loads; the first live turn is E3's.

**Depends on:** Effort 1.

**Own Plan session:** yes. A pass (Pass shape above).

### Effort 3 — Campaign setup

**Scope:** discovery D-10 and the Effort Breakdown.
- `campaigns/hawk/spatial/source/layout.txt` and `data.toml`, from `canon/places.md`, built with `spatial build`. The descriptive answers become numbers here: each place's tech level and population (Decision 9).
- `campaigns/hawk/state/state.seed.toml`, from `canon/factions.md`: factions, scale and stats, homeworlds, governments, H.A.W.K tags, and starting assets that use the SWN stand-in IDs for now.
- The real `rulebooks/hawk/settings.toml`: drift costs from the canon's travel, and a population floor set against the map's populations.
- The first live turn.

**Depends on:** Effort 2. Factions take H.A.W.K tags.

**Own Plan session:** yes (Decision 1).

### Effort 4 — Force pass

**Scope:** `rulebooks/hawk/assets/force_assets.toml` replaces its stand-in, with any new primitives, the canon entries, any tag–asset links it completes (D-11), the starting Force assets swapped in the seed, and a live turn.

**Depends on:** Effort 3.

**Own Plan session:** yes. A pass.

### Effort 5 — Cunning pass

**Scope:** as Effort 4, for `cunning_assets.toml`.

**Depends on:** Effort 4.

**Own Plan session:** yes. A pass.

### Effort 6 — Wealth pass

**Scope:** as Effort 4, for `wealth_assets.toml`, the last stand-in. Its live turn reaches "H.A.W.K playable" (D-1).

**Depends on:** Effort 5.

**Own Plan session:** yes. A pass.
