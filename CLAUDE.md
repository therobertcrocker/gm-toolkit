# GM Toolkit — Claude Instructions

## Collaboration

1. Design is always collaborative. Implementation first pass is usually Claude's.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. Robert is designing, Claude is advising.
4. Always follow Robert's decisions exactly. When he specifies order, structure, or behavior, implement it precisely — do not substitute own judgement.
5. Before writing any code not specified in an implementation doc, ask. Before making design suggestions, wait for a prompt.
6. Before making a commit, do the following:
   1. Update the decisions log (`docs/dev_journal/decisions-log.md`) with any new decisions or changes to existing decisions.
   2. All file changes must be staged and committed, even if they aren't directly related to the feature at hand. Lost work is unacceptable. If you are unsure whether a change should be committed, ask.
7. Before merging a branch, do the following (step-by-step, in order):
   1. A code-review as if you were a senior engineer reviewing a junior's PR. Be critical, but constructive. Don't just point out issues — suggest specific improvements.
   2. Update the dev journal (`docs/dev_journal/dev-journal-factions.md`) with a summary of the work and any relevant notes (e.g. open questions, design decisions, next steps)
8. After every branch merge, assess whether the work warrants a patch/minor/major bump and tag accordingly. Check current version with `git describe --tags --abbrev=0` before tagging.
9.  Use conventional commits: `type: description` (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Concise, imperative mood.

## Documentation Discipline

Before any feature discussion or implementation, read the doc(s) relevant to the work at hand:

| If you're working on…                | Read first                                  |
|--------------------------------------|---------------------------------------------|
| Turn engine, actions, triggers       | `docs/dev_journal/completed/turn-engine-journal.md` |
| Overall design or open questions     | `docs/dev_journal/dev-journal-factions.md`          |
| A ratified decision                  | `docs/dev_journal/decisions-log.md`                 |
| Architecture or system shape         | `docs/architecture-overview.md`             |
| Rules behavior                       | `docs/swn-faction-mechanics.md`             |
| Creating or editing any doc          | `docs/style-guide.md`                       |
| When implementing a feature...       | `docs/implementation/<feature>.md`          |                   

When docs conflict, flag the conflict before proceeding. When docs are silent on something the implementation must decide, ask before deciding.

## Session Modes

Work flows through three modes. Each is its own Claude session.

1. **Discovery** — produces a discovery doc
2. **Plan** — produces an implementation plan, including its phase breakdown
3. **Execution** — implements one phase from the plan. **Each phase is its own session** — a plan with three phases takes three execution sessions, not one.

A session ends at its deliverable. Do not start the next mode's work in the same session, even if its inputs are ready.

**Most-violated boundary: Plan → Execution.** When the current session's output is an implementation plan, stop after the plan is written, approved, or handed off. Do not write code, edit source files, or stage edits. If the user pushes to keep going, suggest opening a new session.

**Plan mode is a strong signal.** When plan mode is active and the stated deliverable is a file (a discovery doc or implementation plan), that file is the entire session output. ExitPlanMode is not a handoff into implementation — it is the end of the session. Do not queue file edits, Bash commands, or follow-up work to run after exit. End the turn after the file is written.

## Code Style

- **Parameter names:** Always full words — `factionState` not `s`, `faction` not `f`, `rulebook` not `rb`.
- **Interactive forms:** Use `huh` for wizard-style prompts and interactive forms.
- **Output styling:** For TUI output, use `lipgloss` styles defined in `cmd/faction-manager/tui/styles.go`. For Cobra CLI output, use the `huh` package and ANSI styling.
- **YAGNI:** Solve the concrete present problem. Don't pitch stronger guarantees (extensibility hooks, future-proofing) unless there is a real, present-day consequence of not doing so. Lead with the simplest fix that addresses the actual bug.
- **No comments** unless the WHY is non-obvious. No docstrings. No "added for X" comments.