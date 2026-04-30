# GM Toolkit — Claude Instructions

## Collaboration

1. Design is always collaborative. Implementation first pass is usually Claude's.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. Robert is designing, Claude is advising.
4. Always follow Robert's decisions exactly. When he specifies order, structure, or behavior, implement it precisely — do not substitute own judgement.
5. Before writing any code, ask. Before making design suggestions, wait for a prompt.
6. Before making a commit, do the following:
   1. Update the decisions log (`docs/decisions-log.md`) with any new decisions or changes to existing decisions.
7. Before merging a branch, do  the following:
   1. Update the dev journal (`docs/dev-journal-factions.md`) with a summary of the work and any relevant notes (e.g. open questions, design decisions, next steps)
   2. Update the tracking journal (`docs/tracking/turn-engine-journal.md`) with the status of relevant features and any open questions that arose during implementation
   3. A code-review as if you were a senior engineer reviewing a junior's PR. Be critical, but constructive. Don't just point out issues — suggest specific improvements.
8. After every branch merge, Check current version with git describe --tags --abbrev=0 before tagging.
9.  Use conventional commits: `type: description` (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Concise, imperative mood.

## Always Consult docs/ First

Before any feature discussion or implementation — and whenever a design question arises mid-implementation — read the relevant doc:

- `docs/dev-journal-factions.md` — overall design, decisions log, progress
- `docs/decisions-log.md` — ratified decisions
- `docs/architecture-overview.md` — high-level architecture and design rationale
- `docs/tracking/turn-engine-journal.md` — feature status and design decisions
- `docs/swn-faction-mechanics.md` — rules reference
- `docs/style-guide.md` — doc and journal style; consult before creating any new doc

## Implementation Discipline

When an implementation doc specifies the work (signatures, triggers, thresholds, formulas), read only the files you'll directly use — the relevant domain types and any data files the spec references. Do not read existing action/engine files for pattern context. Start writing; let the compiler surface gaps.

## Development Setup

Work from `cmd/faction-manager/` as the working directory:

```bash
go build -o bin/faction-manager .
FACTION_DATA_DIR=/workspaces/gm-toolkit/internal/faction/data ./bin/faction-manager <command>
```

- `bin/` holds the compiled binary; `campaigns/` sits alongside it
- `FACTION_DATA_DIR` must point to `internal/faction/data`
- Test campaign: `campaigns/test/faction_state.toml` — run with `--campaign test`
- Tests: `go test ./...` from the repo root

## Code Style

- **Parameter names:** Always full words — `factionState` not `s`, `faction` not `f`, `rulebook` not `rb`.
- **Interactive forms:** Use `huh` for wizard-style prompts and interactive forms.
- **Output styling:** For TUI output, use `lipgloss` styles defined in `cmd/faction-manager/tui/styles.go`. For Cobra CLI output, use the 'huh' package and ANSI styling.
- **YAGNI:** Solve the concrete present problem. Don't pitch stronger guarantees (extensibility hooks, future-proofing) unless there is a real, present-day consequence of not doing so. Lead with the simplest fix that addresses the actual bug.
- **No comments** unless the WHY is non-obvious. No docstrings. No "added for X" comments.

## Terminology

- Currency is **Coin** — never "FacCreds" (that's the SWN source term; this is the GM's own system).
