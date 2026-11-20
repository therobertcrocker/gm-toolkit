# Planned Work — Campaigns

Pre-discovery initiative tracker for the campaigns product. Deferred items move to Current Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins. IDs share one counter with the other products (see [`project-planning.md`](../project-planning.md)).

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `A.004` | Encounter lab | `arc` | `In-Progress` | Tools that show the balance of a planned PF2e encounter and help build the flowchart to run it: sim core (`F.041`), tactics and flowcharts (`A.004.2`), measurement (`A.004.3`), description conversion (`A.004.4`). See the [arc plan](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md); [write-up](#a004--encounter-lab) |
| `F.041` | PF2e sim core | `feature` | `In-Progress` | First initiative of `A.004`: the engine, the Foundry import, the rules, and a `run` that plays the real Fall of Luna fight and prints raw facts. Efforts 1 to 3 built; see [write-up](#f041--pf2e-sim-core) |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `A.004.2` | Tactics and flowcharts | `feature` | `after F.041` | Position, more rule-list conditions, dials, variants, presets, `compare`, the table flowchart; gate 4 and the script's archive. Boundary in the [`A.004` arc plan](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md) |
  

---
<br />

## Backlog

Unscoped items waiting for their trigger. Move to Up Next when the trigger is close; move to Current Initiatives when fully scoped for discovery.

**Trigger values:**

- `--` — no gate; pick up at any time
- `ready` — scoped; can be promoted now
- `after <ID>` — waits for another row to ship
- `when <condition>` — waits for an event in the code or the project
- `needs triage` — the cause is not known yet; diagnose before scoping

### Arcs

Initiatives of an arc that wait for their turn. Each boundary is in the [`A.004` arc plan](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md).

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `A.004.3` | Measurement | `after A.004.2` | Measures a GM can act on, each checked against evidence before it counts; replaces D-17 scoring |
| `A.004.4` | Description conversion | `after F.041` | Turn an ability's text into data the sim runs: trial of the API and skill options, drafts, approval, cache |

### Features

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `F.060` | `scenario create` wizard | `when a hand-written scenario hurts` | A wizard that writes a fight's scenario file; pushed out of `A.004` by its arc plan (AP-9) |

### Refactors

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.036` | One campaign resolve-and-hint helper | -- | Four commands load the registry, call `campaigns.ResolveActive` and map its errors to hints, each in its own copy: `faction.go` (`formatActiveResolveError`), `pf2e/pf2e.go` (`resolveCampaign`), `campaign/add_rules.go` and `spatial/build.go`; the last two hint only `ErrNoActiveCampaign`, not `ErrCampaignNotRegistered` or `ErrCampaignRootMissing`. Share one helper that takes the command's prefix. Filed at the `F.041.E1` close (Decision Record 3). |
| `R.037` | Repo layout by tool | `after F.041.E4` | Folders mirror packages, tool by default: each campaign gets `faction/`, `pf2e/` and a `shared/` wrapper (first tenant `spatial`), `rulebooks/` becomes `library/faction/`, `internal/spatial` moves to `internal/shared/spatial`, and `docs/` is regrouped by tool. Eight agreed rules and three efforts (campaign data, Go packages, docs); Discovery first. Seed: [`repo-layout-seed.md`](./initiatives/seeds/repo-layout-seed.md) |

---
<br />
<br />

# Initiative Write-Ups
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

## A.004 — Encounter lab

Arc. The concept lives in [`A.004-encounter-lab-arc-concept.md`](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-concept.md). The initiatives, their order and their boundaries are in [`A.004-encounter-lab-arc-plan.md`](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md).

**Trigger** — `ready`

**Status** — `In-Progress`

<br />

## F.041 — PF2e sim core

**Problem** — `cmd/scripts/hawk_finale_sim.py` simulates one fight, the H.A.W.K finale. It can't be reused for the next encounter, and its numbers rest on approximations of the rules.

**Approach** — a Go simulator that plays a PF2e fight by the rules as written. Narrowed by the [`A.004` arc plan](initiatives/arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md) to sim core: the engine, the Foundry import, rules coverage (Efforts 1 to 5), then hand-written abilities with `check`, rule lists, and the scenario file with `run` (Efforts 6 to 8). Position, compare and the flowchart are `A.004.2`; measures are `A.004.3`; description conversion is `A.004.4`. Inputs:

- `campaigns/hawk/encounters/fall-of-luna/source/` — PF2e data for the PCs and the two bosses of the first scenario (Foundry actor exports), plus the battlemap
- `campaigns/hawk/encounters/fall-of-luna/reference/` — infographics: a flowchart for running the fight, and guides for building hazards and creature data
- Official PF2e rules on Archives of Nethys: [Building Creatures](https://2e.aonprd.com/Rules.aspx?ID=2874), [Adjusting Creatures](https://2e.aonprd.com/Rules.aspx?ID=3262), [Encounter Design](https://2e.aonprd.com/Rules.aspx?ID=2715); the site is a full rules reference

**Unlocks** — `run` plays The Fall of Luna and prints hit rates, save rates, damage share by source, knockdowns and rounds; the other three initiatives of `A.004` build on its fight record.

**Trigger** — `ready`

**Status** — `In-Progress`
