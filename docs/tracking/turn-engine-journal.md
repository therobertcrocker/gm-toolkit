# Turn Engine — Tracking Journal

A high-level tracker for turn engine feature status and open questions.

> Discovery docs: [turn-engine-discovery.md](../discovery/turn-engine-discovery.md) | [action-resolution-discovery.md](../discovery/action-resolution-discovery.md) | [goal-engine-discovery.md](../discovery/goal-engine-discovery.md)
> 
> Decisions log: [decisions-log.md](../decisions-log.md)


<br/>

## Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Turn Scaffolding | Complete | |
| 2 | Per-Faction Bookkeeping | In progress | Income and maintenance skeleton done; maintenance costs deferred until AssetDefinition carries structured cost data |
| 3 | Action Selection | Complete | All nine SWN actions wired; Change Homeworld and Seize Planet gated by Goal Engine (not started) |
| 4 | Action Resolution | Complete | All nine SWN actions implemented; Change Homeworld and Seize Planet pending Goal Engine multi-turn lock |
| 5 | State Mutation | Complete | |
| 6 | Event Recording | Complete | Per-faction JSONL records; one EventRecord per faction per cycle; HistoryEngine wired into turn wizard |
| 7 | Goal Engine | Complete | |

<br/>
<br/>

# Open Questions

Design questions that are unresolved and will need answers before the relevant feature can be built.

| # | Question | Relevant Feature |
|---|----------|-----------------|
| 1 | ~~How does the Goal Engine communicate lock state back to the turn flow — method call, return value, or flag on `TurnState`?~~ Resolved: `GoalEngine.CheckLock` returns a `GoalLock` value (Option B); the TUI calls it before action selection and routes accordingly | Goal Engine |
