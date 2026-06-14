# Dev Journal: Faction Tracker

Entry point for the faction tracker's design documentation. Append-only north-star: links to other docs (Quick Reference) and durable Design Principles. For active work and open questions, see [Planned Work](planned-work.md).

<br/>

## Quick Reference

| Document | Purpose |
|----------|---------|
| [Architecture](../../architecture/architecture-overview.md) | Front door to the architecture wiki: engine and interface spines plus per-subsystem pages |
| [Planned Work](planned-work.md) | Pre-discovery initiative tracker; all known deferred work |
| [Decisions Log](archive/decisions-log.md) | Frozen historical archive; 294 decisions across all branches |
| [Contributor Guide](../contributor-guide.md) | Setup, workflow, and conventions for development |
| [Implementation Plans](../implementation/) | Per-feature implementation plans, written before execution |
| [Discovery Docs](../discovery/) | Pre-implementation design explorations |

<br/>

## Design Principles

- **CLI-first** — core logic is decoupled from any UI layer
- **Source-agnostic actions** — actions can be driven by human input or AI equally
- **Separation of concerns** — review and turn execution are distinct flows
- **Extensible static data** — GMs customize assets and flavor without touching code
- **Robust history** — every state change is recorded; nothing is lost
- **Built to grow** — a TUI or web frontend can be layered on later without touching core logic
