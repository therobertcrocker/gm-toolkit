# GM Toolkit — Claude Instructions

## Collaboration

1. Design is always collaborative. Implementation first pass is usually Claude's.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. Robert is designing, Claude is advising.
4. Always follow Robert's decisions exactly. When he specifies order, structure, or behavior, implement it precisely — do not substitute own judgement.
5. Always use the LSP tools available to search through the existing codebase - don't use grep
6. Always create a to-do list for any series of tasks, and make regular check-ins with Robert to confirm the next steps. Do not execute a series of tasks without explicit confirmation of the plan.
7. Before making a commit, do the following:
   1. Update the decisions log (`docs/dev_journals/faction-manager/decisions-log.md`) with any new decisions or changes to existing decisions.
   2. All file changes must be staged and committed, even if they aren't directly related to the feature at hand. Lost work is unacceptable. If you are unsure whether a change should be committed, ask.
8. Before merging a branch, do the following (step-by-step, in order):
   1. A code-review as if you were a senior engineer reviewing a junior's PR. Be critical, but constructive. Don't just point out issues — suggest specific improvements.
   2. Update the planned work doc (`docs/dev_journals/faction-manager/planned-work.md`) with any new features or deferred decisions that arose during the work. If a feature was completed, remove it from the planned work doc.
9. After every branch merge, assess whether the work warrants a patch/minor/major bump and tag accordingly. Check current version with `git describe --tags --abbrev=0` before tagging.
10. Use conventional commits: `type: description` (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Concise, imperative mood.
11. Three Principles of Code:
    1. **Clarity:** Code should be easy to read and understand. Prioritize readability over cleverness or brevity.
    2. **Correctness:** Code should do what it's supposed to do, and handle edge cases gracefully. Don't just solve the happy path.
    3. **Maintainability:** Code should be organized and structured in a way that makes it easy to modify and extend in the future. Avoid unnecessary complexity.

## Documentation Discipline

Before any feature discussion or implementation, read the doc(s) relevant to the work at hand:

| If you're working on…                | Read first                                  |
|--------------------------------------|---------------------------------------------|
| Overall design or open questions     | `docs/dev_journals/faction-manager/dev-journal-factions.md`          |
| Feature Request or Deferred Decision | `docs/dev_journals/faction-manager/planned-work.md`                      |
| Before making a design choice        | `docs/dev_journals/faction-manager/decisions-log.md` — use the index (see below) |
| Architecture or system shape         | `docs/architecture-overview.md`             |
| Rules behavior                       | `docs/swn-faction-mechanics.md`             |
| Creating or editing any doc          | `docs/style-guide.md`                       |
| When implementing a feature...       | `docs/initiatives/implementation/<feature>.md`          |
| Session modes, flows, or templates   | `docs/process/session-modes.md`             |

**Searching the decisions log:** Do not read the full file. Use one of:
- **Index first:** Read the index at the top (~25 lines), find the relevant section by topic, then read only that section using `offset`/`limit`.
- **Keyword grep:** `grep -n "keyword" docs/dev_journals/faction-manager/decisions-log.md` to find the line, then read the surrounding section.

When docs conflict, flag the conflict before proceeding. When docs are silent on something the implementation must decide, ask before deciding.

## Session Modes

See `docs/process/session-modes.md` for the full description of Discovery / Plan / Execution modes, the initiative lifecycle, per-type flows (feature / refactor / bugfix / docs / chore), session discipline, and templates. Always read the session modes doc before starting a new initiative or session, and refer back to it as needed.

## Session Boundaries
| Session Type | Boundary Definition |
|--------------|---------------------|
| Discovery    | Once the discovery doc is complete and the next steps are clear, the session is over. |
| Plan         | Once the implementation plan is complete and the next steps are clear, the session is over. |
| Execution    | One session per Commit. Once the commit is made, the session is over. |

## Model Selection

Before each mode transition (including mid-session shifts, e.g. moving from execution into the pre-merge checklist), surface the recommended model and prompt Robert to switch if needed (via `/model`).

| Session type                              | Model  | Notes                                                                 |
|-------------------------------------------|--------|-----------------------------------------------------------------------|
| Discovery                                 | Opus   | Always                                                                |
| Plan                                      | Opus   | Always                                                                |
| Execution                                 | Sonnet | Default; suggest Opus for Refactoring Efforts or if the phase involves heavy design decisions or unusual complexity |
| End-of-phase docs & pre-merge checklist  | Sonnet | Always — documentation and checklist tasks don't need Opus            |

## Code Style

- **Parameter names:** Always full words — `factionState` not `s`, `faction` not `f`, `rulebook` not `rb`.
- **Interactive forms:** Use `huh` for wizard-style prompts and interactive forms.
- **Output styling:** For TUI output, use `lipgloss` styles defined in `cmd/faction-manager/tui/styles.go`. For Cobra CLI output, use the `huh` package and ANSI styling.
- **YAGNI:** Solve the concrete present problem. Don't pitch stronger guarantees (extensibility hooks, future-proofing) unless there is a real, present-day consequence of not doing so. Lead with the simplest fix that addresses the actual bug.
- **No comments** unless the WHY is non-obvious. No docstrings. No "added for X" comments.