# Planned Work — Campaigns

Pre-discovery initiative tracker for the campaigns product. Deferred items move to Current Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins. IDs share one counter with the other products (see [`project-planning.md`](../project-planning.md)).

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|
| `F.041` | PF2e encounter simulator | `feature` | `In-Progress` | Grow `cmd/scripts/hawk_finale_sim.py` into a reusable, more useful PF2e fight simulator; first scenario is The Fall of Luna; see [write-up](#f041--pf2e-encounter-simulator) |


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
_None._
  

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

### Refactors

| ID | Item | Trigger | Detail |
|----|------|---------|--------|
| `R.036` | One campaign resolve-and-hint helper | -- | Four commands load the registry, call `campaigns.ResolveActive` and map its errors to hints, each in its own copy: `faction.go` (`formatActiveResolveError`), `pf2e/pf2e.go` (`resolveCampaign`), `campaign/add_rules.go` and `spatial/build.go`; the last two hint only `ErrNoActiveCampaign`, not `ErrCampaignNotRegistered` or `ErrCampaignRootMissing`. Share one helper that takes the command's prefix. Filed at the `F.041.E1` close (Decision Record 3). |

---
<br />
<br />

# Initiative Write-Ups
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.

<br />

## F.041 — PF2e encounter simulator

**Problem** — `cmd/scripts/hawk_finale_sim.py` simulates one fight, the H.A.W.K finale. It was written for that fight alone, so it can't be reused for the next encounter, and it answers fewer questions than a GM needs when tuning a fight.

**Approach** — expand the script into a tool that is reusable across encounters and more useful as a balancing aid. Discovery starts fresh; the earlier draft discovery doc is not binding and is mined only for what fits. Inputs:

- `campaigns/hawk/encounters/fall-of-luna/source/` — PF2e data for the PCs and the two bosses of the first scenario (Foundry actor exports), plus the battlemap
- `campaigns/hawk/encounters/fall-of-luna/reference/` — infographics: a flowchart for running the fight, and guides for building hazards and creature data
- Official PF2e rules on Archives of Nethys: [Building Creatures](https://2e.aonprd.com/Rules.aspx?ID=2874), [Adjusting Creatures](https://2e.aonprd.com/Rules.aspx?ID=3262), [Encounter Design](https://2e.aonprd.com/Rules.aspx?ID=2715); the site is a full rules reference

**Unlocks** — testing a PF2e encounter's balance before the session instead of at the table, starting with The Fall of Luna; later encounters reuse the tool.

**Trigger** — `ready`

**Status** — `In-Progress`
