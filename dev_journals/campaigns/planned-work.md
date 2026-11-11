# Planned Work — Campaigns

Pre-discovery initiative tracker for the campaigns product. Deferred items move to Current Initiatives when scoped -- at that point a write-up is added below and a discovery doc is written before implementation begins. IDs share one counter with the other products (see [`project-planning.md`](../project-planning.md)).

<br/>

## Current Initiatives

Full write-ups below. Each item has been scoped enough to warrant a dedicated discovery phase.

| ID      | Item | Type | Status | Detail |
|---------|------|------|--------|--------|

_None._


---
<br/>

## Up Next (Ready or Near Ready)
The queue of deferred items that are ready to become initiatives. These are scoped and waiting for their trigger conditions to be met.

| ID      | Item | Type | Trigger | Detail |
|---------|------|------|---------|--------|
| `F.041` | PF2e encounter simulator | `feature` | ready | Plays a PF2e fight thousands of times from Foundry exports and reports outcome odds, fight length and ability use; new `pf2e` rules engine plus a `sim` runner, six efforts; first scenario is the H.A.W.K finale (`campaigns/hawk/encounters/fall-of-luna/`); draft discovery doc on `feature/pf2e-encounter-simulator`, not yet signed off; no wiki page covers PF2e, so it adds `architecture/pf2e/overview.md` and `pf2e/sim.md` |


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

_None._

---
<br />
<br />

# Initiative Write-Ups
This section contains detailed write-ups for each planned initiative, including problem statements, proposed approaches, tradeoffs, and triggers. This will be the source of truth when it comes time to start discovery work on any of these items.
