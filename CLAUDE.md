# GM Toolkit — Claude Instructions

## Collaboration

1. Design is always collaborative. Implementation first pass is usually Claude's.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. Robert is designing, Claude is advising.
4. Always follow Robert's decisions exactly. When he specifies order, structure, or behavior, implement it precisely — do not substitute own judgement.
5. Before writing any code, ask. Before making design suggestions, wait for a prompt.
6. Before merging a branch, do  the following:
   1. A code-review as if you were a senior engineer reviewing a junior's PR. Be critical, but constructive. Don't just point out issues — suggest specific improvements.
   2. Update the dev journal (`docs/dev-journal-factions.md`) and decisions log (`docs/decisions-log.md`) with any relevant notes, decisions, or reflections from the implementation process. This is crucial for maintaining a clear record of the project's evolution and rationale behind decisions.
7. After every branch merge, assess whether the work warrants a patch/minor/major bump and tag accordingly (current version: **v0.10.1**).
8.  Use conventional commits: `type: description` (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Concise, imperative mood.

## Always Consult docs/ First

Before any feature discussion or implementation — and whenever a design question arises mid-implementation — read the relevant doc:

- `docs/dev-journal-factions.md` — overall design, decisions log, progress
- `docs/decisions-log.md` — ratified decisions
- `docs/discovery/turn-engine-discovery.md` — turn engine sub-systems and interfaces
- `docs/discovery/action-resolution-discovery.md` — per-action resolution logic
- `docs/discovery/goal-engine-discovery.md` — Goal Engine and multi-turn processes
- `docs/tracking/turn-engine-journal.md` — feature status and design decisions
- `docs/swn-faction-mechanics.md` — rules reference
- `docs/style-guide.md` — doc and journal style; consult before creating any new doc

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
