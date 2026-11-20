# Initiative Seed — Repo Layout by Tool

> **Type:** Initiative seed, not a plan or a directive
> **Status of the work it describes:** Concept (pre-plan). The rules below are agreed; the work is not scheduled or promoted.
> **Source:** a design discussion between Robert and Claude on 2026-11-20, held while `F.041.E4` ran
> **Tracker row:** `R.037` in `dev_journals/campaigns/planned-work.md`
> **Written:** 2026-11-20, against gm-toolkit `feature/pf2e-encounter-simulator` @ `85b6d58`

<br/>

## What an initiative seed is

A seed sits *before* the initiative lifecycle. It records a piece of work that has been thought through at the concept level. A seed carries no commitment: it may change shape or never be planted. Whether it enters the lifecycle is Robert's call.

<br/>

## Read this first

This file exists so that sessions in this repo know a restructure of the folders is coming and do not build against it. It is **awareness, not authorization.**

- **Do not start the moves.** No folder renames, no package moves, no doc moves.
- **Do not promote `R.037`.** Promotion is Robert's call.
- **Re-ground before relying on any repo fact below.** The folder and package facts are a snapshot at `85b6d58`.
- **The rules are settled; the layout is a sketch.** Robert agreed each rule in "The rules" one at a time. The trees under "The picture" show where the rules lead; Discovery checks them against the source.

*Practical effect today:*

- **New data starts in its tool's space** (rule 1). A new folder at the root of a campaign needs a reason.
- **No tool reads another tool's files** (rule 2). If a second tool needs some data, stop and raise it; do not open the other tool's folder.
- **A new architecture page for the faction tool** should not add to the general names `engine/` and `interface/` without a thought for the move below.

<br/>
<br/>

# The problem

gm-toolkit started as one tool and is now a home for several. The folders still show the first shape:

- **A campaign folder does not say what belongs to what.** In `campaigns/hawk/`, the faction tool's `rulebook/`, `state/` and `narratives/` sit loose at the root beside `spatial/` (shared), `pf2e/` (a tool), `encounters/` (PF2e data outside `pf2e/`) and `canon/` (the seed of the codex's folder).
- **Names say "the engine" where there are now two.** `docs/architecture/engine/` and `interface/` are the faction tool's pages; the PF2e engine page sits beside them at `pf2e/engine.md`.
- **One tool's data sits at the repo root.** `rulebooks/` holds the faction tool's rulebook sources.

<br/>

# The rules

The foundation is **tool by default, with strict rules for when data becomes shared.** Robert expects the shared part to grow as tools are added, with only some data staying specific to each tool.

1. **Tool by default.** Data starts in its tool's space.
2. **Two readers means one home apart from both.** A tool never reads another tool's data.
3. **Shared data has a universal model and one owner.** The owner is one package, and it is the only code that touches the files. Any tool may ask the owner to write.
4. **Tool-specific models stay in tool space.**
5. **No duplication.** If two tools do the same thing, it is shared. The test is "if one changes, must the other?", because two things can look alike and be different things.
6. **Folders mirror packages.** A tool's folder takes the name of its top package under `internal/`, in data, commands and docs alike. So `faction`, not `factions` or `faction-manager`; and `pf2e`, not `sim`.
7. **Shared things live under a `shared/` wrapper, in data and in code.** A shared domain has the shape `spatial` has today: an owner package, an optional `source/` for hand-written input, and built files beside it.
8. **Data that is the same for every campaign lives in `library/<tool>/`.**

Notes on the rules:

- **Rule 3 answers "one writer or many".** One owner package, which any tool can call, gives a single place where the model is enforced.
- **Rule 7 is not a new design.** `internal/spatial` is already the only loader of `campaigns/<id>/spatial/`, `internal/spatial/builder` is its only writer, and the faction packages read through it. The rule names that shape and makes it the pattern.
- **Share identity first.** When a second tool needs something a first tool owns, the first thing to move is usually the ID and name, not the whole model (Claude's suggestion; not yet tested on a real case).
- **Product names are a different axis.** `dev_journals/<product>/` stays as it is; a product can hold several tools.

<br/>

# The picture

A sketch of where the rules lead. Discovery confirms each move against the source.

```
campaigns/hawk/
  campaign.toml
  shared/
    spatial/
  faction/      rulebook/  state/  narratives/
  pf2e/         party/  encounters/
  canon/        becomes codex/ when the codex has a package

library/
  faction/      _core/  swn/  hawk/      (rulebooks/ today)

internal/
  campaigns/    the frame: registry, manifest, paths
  shared/
    spatial/
  faction/
  pf2e/

docs/architecture/
  architecture-overview.md
  faction/      engine/  interface/
  pf2e/
```

Decisions inside the picture:

- **`canon/` waits.** It becomes `codex/` when the codex gets its package, not before; a rename now would mirror a package that does not exist.
- **`shared/` starts with one tenant.** Only `spatial` has two readers in sight. Nothing else moves in until its second reader is real.
- **Inside an encounter, `source/` and `reference/` keep their names.** `source/` means hand-written or exported input, the same as in `spatial`.
- **The command stays `gm-toolkit spatial`.** Commands are what a user types; only the package moves under `shared/`.
- **`library/pf2e/` is expected** (a bestiary, or the rules reference that is coming from another project) but is not created until it has content.

<br/>

# What the read of the repo found

Places that break the rules at `85b6d58`:

| # | Place | Problem | Where the rules lead |
|---|-------|---------|----------------------|
| 1 | `campaigns/<id>/rulebook/`, `state/`, `narratives/` | The faction tool's data, loose at the campaign root | `campaigns/<id>/faction/` |
| 2 | `campaigns/<id>/encounters/` | PF2e data outside `pf2e/` | `campaigns/<id>/pf2e/encounters/` |
| 3 | `campaigns/<id>/spatial/`, `internal/spatial` | Shared, but not marked as shared | `shared/spatial` in both |
| 4 | `rulebooks/` at the repo root | One tool's library with a general name | `library/faction/` |
| 5 | `internal/campaigns.ComposeRulebook` | The frame does faction work | Move under `internal/faction` |
| 6 | `cmd/gm-toolkit/faction.go` | A loose file where the other tools have folders | `cmd/gm-toolkit/faction/` |
| 7 | `docs/architecture/engine/`, `interface/` | The faction tool's pages under general names | `docs/architecture/faction/` |
| 8 | `cmd/scripts/hawk_finale_sim.py` | The Python sim from before the Go simulator | `A.004.2` archives it at gate 4; not this work |

<br/>

# The `docs/` folder

Robert asked that the whole `docs/` folder get the same treatment, not only the pages with misleading names.

| Part | What to do |
|------|------------|
| `architecture/engine/`, `architecture/interface/` | Faction pages; move under `architecture/faction/` |
| `architecture/architecture-overview.md` | The front door; check that it describes a repo of several tools, not the faction tool alone |
| `rules/swn-*.md` | Faction rules docs; group by tool. The PF2e rules scope doc and the incoming rules reference need a home on the same pattern |
| `contributing/engine.md`, `interface.md`, `spatial-maps.md` | Faction and shared guides under general names; group by tool, with `shared` for spatial |
| `_archive/` | Review what is there (`process`, `spatial-engine`) and whether it still earns its place |
| Links and hooks | Every moved page breaks links, the "Extra docs" table in `CLAUDE.md` and the SessionStart hook's page list; fix them all in the same effort |

Two limits:

- **The gestalt rules fix `docs/architecture/` and its front door** as paths a project cannot override. The restructure happens inside that folder.
- **Another seed touches the same pages.** `dev_journals/faction-manager/initiatives/seeds/living-spec-migration-seed.md` plans to split the docs by the question a reader asks. Discovery decides the order of the two, so the pages move once.

<br/>

# Shape of the work

One refactor initiative, Discovery first, with three efforts in this order:

1. **Campaign data.** The moves inside each campaign folder and `rulebooks/` to `library/`, with the path code in `internal/campaigns` and a migration for existing campaigns. `hawk` and `sandbox` hold Robert's real files, so this effort goes first, while the picture is fresh, and has its own gate.
2. **Go packages.** `internal/spatial` to `internal/shared/spatial` (about 15 importing packages), `ComposeRulebook` to faction, `cmd/gm-toolkit/faction/`. Mechanical renames.
3. **Docs.** The table above.

Discovery is a structure-setting session, so it may run on Fable.

<br/>

# Open questions for Discovery

- **Migration.** How an existing campaign moves to the new layout: a `gm-toolkit` command, a one-time script, or by hand; and what the binary does when it meets the old layout.
- **`hawk-claude` and `sandbox-claude`.** They are git-ignored copies; the migration must work on them too.
- **Logs.** `campaigns/<id>/state/logs/` are written by `internal/logging` for the faction engine. Check whether logs are faction data or belong to the frame.
- **`campaigns/<id>/pf2e/party/`.** It sits in the PF2e tool's space today. If a non-PF2e tool ever needs the party, rule 2 moves the shared part out; nothing needs it yet.
- **`R.036`** (one campaign resolve-and-hint helper) touches the same frame code. Decide whether it rides with effort 2.
- **Order against the living-spec seed** (see "The `docs/` folder").
