# GM Toolkit — Claude Instructions

## Collaboration

1. Design is always collaborative. Implementation first pass is usually Claude's.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. Robert is designing, Claude is advising.
4. Always follow Robert's decisions exactly. When he specifies order, structure, or behavior, implement it precisely — do not substitute own judgement.
5. Always use the LSP tools available to search through the existing codebase - don't use grep
6. Always create a to-do list for any series of tasks, and make regular check-ins with Robert to confirm the next steps. Do not execute a series of tasks without explicit confirmation of the plan.
7. Before making a commit, do the following:
   1. Log any execution-time decisions (or reversals of planned decisions) in the initiative plan's **Decision Record** section. Append and reconcile: when execution overrides a planned decision, record the reversal with rationale and fix the plan body so it doesn't lie. (The old decisions log is a frozen archive; see `session-modes.md` section 10, Decision records.)
   2. All file changes must be staged and committed, even if they aren't directly related to the feature at hand. Lost work is unacceptable. If you are unsure whether a change should be committed, ask.
8. Before merging a branch, do the following (step-by-step, in order):
   1. A code-review as if you were a senior engineer reviewing a junior's PR. Use /code-review skill. Be critical, but constructive. Don't just point out issues — suggest specific improvements.
   2. Promote durable decisions: graduate architecturally-significant entries from the plan's Decision Record into the relevant architecture page's `Key Decisions` section (`docs/architecture/`). Micro-decisions stay in the plan as provenance.
   3. Update the planned work doc (`docs/dev_journals/faction-manager/planned-work.md`) with any new features or deferred items that arose during the work. If a feature was completed, remove it from the planned work doc.
   4. Move the initiative's discovery and plan docs to their `completed/` subdirectories (arc artifacts stay in `docs/initiatives/arcs/<arc-name>/`).
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
| Before making a design choice        | The relevant architecture page's `Key Decisions` (`docs/architecture/`) plus the initiative plan's Decision Record |
| Architecture or system shape         | `docs/architecture/` wiki (start at `docs/architecture/architecture-overview.md`) |
| Setup, workflow, or dev conventions  | `docs/contributing/overview.md` (front door; routes to engine/interface guides) |
| Historical decision archaeology      | `docs/dev_journals/faction-manager/archive/decisions-log.md` — frozen archive, never appended |
| Rules behavior                       | `docs/rules/swn-faction-mechanics.md`       |
| Creating or editing any doc          | `docs/process/style-guide.md`                       |
| Creating or editing a diagram        | `docs/process/diagram-kit.md`                       |
| When implementing a feature...       | `docs/initiatives/implementation/<feature>.md`          |
| Session modes, flows, or templates   | `docs/process/session-modes.md`             |

**Searching the decisions log (historical only — log is frozen):** Do not read the full file. Use one of:
- **Index first:** Read the index at the top (~25 lines), find the relevant section by topic, then read only that section using `offset`/`limit`.
- **Keyword grep:** `grep -n "keyword" docs/dev_journals/faction-manager/archive/decisions-log.md` to find the line, then read the surrounding section.

When docs conflict, flag the conflict before proceeding. When docs are silent on something the implementation must decide, ask before deciding.

## Session Modes

See `docs/process/session-modes.md` for the full description of Discovery / Plan / Execution modes, the initiative lifecycle, per-type flows (arc / feature / refactor / bugfix / docs / chore), the effort mechanism for large initiatives, session discipline, and templates. Always read the session modes doc before starting a new initiative or session, and refer back to it as needed. Templates for docs are found in docs/process/templates/

**Discovery:** Start with a scaffolded doc in `docs/initiatives/discovery/<initiative-name>.md`. Use the template and fill in the sections as you go. The goal is to explore the problem space, gather information, and identify potential solutions. The output is a clear definition of the problem, a set of possible approaches, and a recommended next step (usually a Plan session).

**Implementation:** Read the discovery doc, then relevant code. Ask any open questions, and then determine the Phases, Commits, Tasks. Write the implementation doc in one go at `docs/initiatives/implementation/<initiative-name>.md` using the template. The goal is to create a clear, actionable plan for implementing the feature, refactor, or bugfix. The output is a detailed implementation plan that can be executed in the next phase.

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
| Arc-Discovery / Arc-Plan / process redesign | Fable  | Structure-setting sessions whose output constrains many future sessions (taxonomy, cross-cutting architecture, the process itself); most initiatives never need it |
| Discovery                                 | Opus   | Always                                                                |
| Plan                                      | Opus   | Always                                                                |
| Execution (code)                          | Sonnet | Default for code commits; suggest Opus for Refactoring Efforts or if the phase involves heavy design decisions or unusual complexity |
| Generative documentation                  | Opus   | Authoring real docs (architecture pages, contributor guides) — Robert does not trust Sonnet for documentation work. The split is mechanical-vs-generative, not code-vs-docs |
| Code review (interim or pre-merge)        | Opus   | Always — critical design review benefits from Opus reasoning          |
| End-of-phase mechanical docs & pre-merge checklist | Sonnet | *Mechanical* upkeep only — flipping a note, planned-work/journal table edits; the code-review step within the checklist uses Opus (above) |

## Code Style

- **Parameter names:** Always full words — `factionState` not `s`, `faction` not `f`, `rulebook` not `rb`.
- **YAGNI:** Solve the concrete present problem. Don't pitch stronger guarantees (extensibility hooks, future-proofing) unless there is a real, present-day consequence of not doing so. Lead with the simplest fix that addresses the actual bug.
- **No comments** unless the WHY is non-obvious. No docstrings. No "added for X" comments.
- **Fence-signs are the licensed exception:** `// deliberate: X, not Y, because <reason>` — a one-line marker at a spot a future reader would be tempted to "fix." This is exactly the non-obvious-WHY case. One line, placed at the fence; binding during re-grounding unless explicitly revisited with Robert.