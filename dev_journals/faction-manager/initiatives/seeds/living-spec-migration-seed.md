# Initiative Seed — Behaviour Spec Migration

> **Type:** Initiative seed, not a plan or a directive
> **Status of the work it describes:** Concept (pre-plan). Nothing here is approved, scheduled, or promoted.
> **Source:** gestalt `E.001` (OpenSpec comparison), sessions 1 and 2, 2026-11-16. Doc:
> `~/.claude/plugins/marketplaces/gestalt/dev_journals/initiatives/explorations/E.001-openspec-comparison-exploration.md`
> (present once the gestalt marketplace has updated past `09673ad`).
> **Written:** 2026-11-16, against gm-toolkit `main` @ `8b1f896`

<br/>

## What an initiative seed is

A seed sits *before* the initiative lifecycle. It records a piece of work that has been thought through at the concept level but has not entered the backlog. A seed carries no commitment: it may become a backlog item, change shape, or never be planted. Whether it enters the lifecycle is Robert's call.

<br/>

## Read this first

This file exists so that sessions in this repo know a restructure of the docs is coming and don't build against it. It is **awareness, not authorization.**

- **Do not start the migration.** No spec files, no runbooks, no page slimming.
- **Do not promote it** in any `planned-work.md`. Promotion is Robert's call. You may suggest; you may not move it.
- **Re-ground before relying on any repo fact below.** The doc facts are a snapshot at `8b1f896`. Pages, counts and line numbers move.
- **YAGNI still applies.** The only legitimate effect of this seed on present work is *avoiding a choice that would make the migration painful*, and flagging it when you see one.

*Practical effect today:*

- **`D.010`** (refresh the architecture page bodies for arc `A.001`) is the work this seed would absorb. Before starting `D.010` on its own, ask Robert whether to plant this seed instead.
- **A close that updates an architecture page** should not grow the Shape section with new behaviour step lists or counts. Those are the parts that drift (below). A behaviour ruling goes in Key Decisions, as today.

<br/>
<br/>

# What's coming

The docs split by **the question a reader asks**, so "what does the system do" gets a home of its own, kept current by the change process.

```
question                 home                          kept current by
what does it do?     ->  behaviour spec, per area  ->  delta in discovery, merged at close
why is it this way?  ->  Key Decisions (as today)  ->  graduation at close (as today)
how is code shaped?  ->  arch page Shape, slimmed  ->  seams and contracts only, no counts
how do I extend it?  ->  runbook, per area         ->  checked when a change touches the area
the game's own rules ->  docs/rules/ (the book)    ->  rarely changes
```

- **The behaviour spec** is one file per behaviour area, in plain game terms with no function names. Prose by default; requirement and scenario blocks only where they help. `docs/architecture/pf2e/engine.md` is already close to this style.
- **It is carved, not added.** The spec is seeded by moving content: the engine's readings out of `docs/rules/swn-faction-mechanics.md`, and the catalogues, Shape step lists and behaviour rulings out of the architecture pages. Nothing is copied.
- **Each rule cites the Key Decision behind it,** so decision records gain a target in behaviour.
- **The close merges behaviour changes against the merged code,** so "docs reflect the source" still holds: a discovery's behaviour delta is intent until the close, and the merged spec records what landed.
- **`docs/contributing/` becomes runbooks,** one per layer or behaviour area, for how to extend it.

The idea comes from OpenSpec (`Fission-AI/OpenSpec`), whose centre is a living spec kept current by merging each finished change into it. gestalt stays the process; only the idea is taken.

<br/>
<br/>

# Why

From a full read of `docs/` at `8b1f896`, recorded in `E.001` ("gm-toolkit's doc layers").

- **Behaviour has no home of its own.** It sits in five places: the rules doc, the page catalogues, Key Decisions, Shape step lists and the PF2e pages. None is organised by the question a session asks. "What does Lobbyists do?" spans `hooks.md`, `world-movement.md`, `actions.md` and the rules doc.
- **What close appends stays current; what nobody owns rots.** Key Decisions are graduated at every close and match the code. Shape bodies and the contributing recipes are touched by no step. Nine cases were checked against the code:

| Doc | Says | Code |
|---|---|---|
| `effect-mutation.md`, Shape | `TagEngine` with four Go tag handlers; "five categories" | `engine/tag` is gone; six categories |
| `effect-mutation.md`, catalogue | `tag_added` embeds the full `Tag` | `TagAdded` carries `TagID` |
| `actions.md`, Purpose | Four actions "and eight others" | 10 registered |
| `overlays.md` | "all twelve actions are now wired" | 10 in `ImplementedActions` |
| `hooks.md`, Key Decisions | "Eight small interfaces" for Cat 4 | 12 |
| `world-movement.md`, Shape | An order advances by `def.Speed` hexes | Movement points with `Banked` |
| `persistence.md`, Shape | Drift costs in `drift_costs.toml`; out of range takes the worst cost | `settings.toml`; the rating is clamped to the table |
| `contributing/engine.md` | "Implement a tag handler" in `engine/tag/` | The package is gone |
| `contributing/interface.md` | A `newOverlay` switch whose default is a placeholder | The `askEntries` table; a missing row panics |

- **The old decisions log failed on the same point.** It was archived because "nobody retrieves on that axis" (branch and time). Key Decisions fixed that for why, filed per subsystem. The spec does the same for what, filed per behaviour area.
- **Doc volume is a hard limit.** `D.002` found doc volume to be "the mechanism behind the big-picture loss". Only a net move of content keeps this inside that limit.

<br/>
<br/>

# How it would run

**Migrate first, then record.** The shape depends on this repo's content, and gestalt's docs flow says discover-at-write work can't be pre-specified. So the migration runs here, and gestalt `F.005` writes what worked into the plugin afterwards: templates, the close merge step, the discovery delta section, the paths in `code-and-docs.md`, and `/gestalt:adopt`.

- **Flow:** a multi-effort docs initiative (gestalt `docs-chore-flow.md`): Concept → Plan → one effort per behaviour area. No Discovery.
- **Effort 1 is a pilot.** One area end to end: its spec file, its slimmed architecture page, its runbook, and its rulings moved out of `docs/rules/`. Robert reviews it, and that fixes the shape for the rest. The proposed pilot is goals: the freshest area, with a catalogue, step lists and the Goal Rulings to move.
- **A local close rule from the first moved area on.** Once an area is moved, `CLAUDE.md` carries a rule that each initiative's close merges its behaviour changes into that area's spec. A project's `CLAUDE.md` wins over the gestalt rules, so this is allowed until gestalt `F.005` moves the rule into the plugin. Areas not yet moved run as today.
- **The pilot names the spec folder,** and that name becomes gestalt's layout path. Adding a folder overrides no gestalt path (`docs/contributing/` and `docs/rules/` already sit outside the layout), but choose it knowing every gestalt project will use it.

<br/>
<br/>

# Where it touches this repo

- **`docs/architecture/`** — 24 pages, about 5,200 lines. Shape slims; catalogues and behaviour step lists move to the spec; Key Decisions and Dependencies stay.
- **`docs/rules/`** — `swn-assets.md` stays as it is (the book's asset data). `swn-faction-mechanics.md` keeps the book's rules; its engine readings (Goal Rulings, the permission and lockdown paragraphs under Government Permission) move to the spec.
- **`docs/contributing/`** — becomes runbooks per layer or behaviour area. Setup and build stay somewhere a newcomer finds first.
- **`CLAUDE.md`** — the local close rule, and the Extra docs table: "Rules behavior" points at the spec, not the rules doc.
- **`D.010`** — absorbed; the moved pages replace the stale bodies.
- **Out of scope:** `campaigns/sandbox/TESTING.md`. It is the sandbox campaign's guide to running game scenarios, not a tool doc.

<br/>
<br/>

# Decisions already made (Robert, gestalt `E.001`) — do not re-open

- Stay on gestalt; take OpenSpec's living spec as an idea, not its tool.
- Split the docs by question: spec for what, Key Decisions for why, slimmed pages for code shape.
- Carve the spec out of what exists; never add it on top.
- Spec rules in prose by default, with structure where it helps.
- `docs/contributing/` becomes runbooks per layer or behaviour area.
- gm-toolkit shapes the template; later projects inherit it.
- Migrate first, then record the process in gestalt.

<br/>
<br/>

# Open questions (for the Concept and Plan — not to be answered now)

- The list of behaviour areas and where their edges fall. A first guess for the faction manager: combat, purchase and permission, movement, goals, abilities and effects, the turn and its bookkeeping.
- The pilot area: goals, or another.
- The spec folder's name.
- How much of each Shape section survives, and what a "seam or contract only" page looks like.
- What a runbook holds, and how it differs from today's recipes.
- What behaviour the interface pages hold (Esc meanings, prompts, the goal-pick form), and whether the TUI gets spec areas of its own.
- Whether the PF2e pages are in scope; they belong to the `campaigns` product and are already close to the spec style.
- How a spec rule points at its tests. `E.001` did not read the tests.
- Where requirement and scenario blocks earn their place over prose.

<br/>
<br/>

# If this seed is planted (Robert to decide)

A reasonable shape for the entry, under Docs in `dev_journals/faction-manager/planned-work.md`, with the next free `D.` ID (`D.012` at the time of writing). Remove `D.010` in the same edit.

```markdown
| `D.0xx` | Behaviour spec migration | -- | Split the docs by question (gestalt `E.001`, M-6): a behaviour spec per area for what the system does, in plain game terms with no function names (prose by default, structure where it helps); Key Decisions stay on the architecture pages for why; Shape slims to seams and contracts, with no counts or step lists; `docs/contributing/` becomes runbooks per layer or behaviour area; `docs/rules/` keeps the book, and the engine's readings move into the spec. Seed by moving, not copying (`D.002`'s doc-volume limit). Multi-effort docs initiative: Concept → Plan → one effort per area; Effort 1 pilots one area end to end and fixes the shape, including the spec folder's name. Once an area is moved, a local close rule in `CLAUDE.md` merges each initiative's behaviour changes into its spec, until gestalt `F.005` takes the rule into the plugin. Absorbs `D.010`. Seed: [`behaviour-spec-migration-seed.md`](initiatives/seeds/behaviour-spec-migration-seed.md) |
```

Then tell gestalt the ID, so `F.005`'s trigger names it.

<br/>
<br/>

# Mini-glossary additions

- **Behaviour spec** — one file per behaviour area that says what the system does today, in plain game terms. Kept current by merging each initiative's behaviour changes at close.
- **Behaviour delta** — a discovery doc's list of the behaviour it adds, changes and removes. Intent until the close; merged into the spec once the code lands.
- **Runbook** — how to extend one layer or behaviour area; the successor to the `docs/contributing/` recipes.
- **Carve, not add** — the spec is made by moving content out of existing docs, never by copying it.
