# D.002 — Project & Process Review

> **Purpose:** judge the project against Robert's own goal, and judge whether the
> process earns its cost. Written in a Fable session on 2026-10-17 at `v1.13.0`.
> Inputs: the evidence pack [project-review-evidence.md](project-review-evidence.md)
> (facts only, gathered in a separate Opus session), the May baseline
> [`CODEBASE_REVIEW.md`](../implementation/completed/CODEBASE_REVIEW.md), and one
> bounded read of the code (listed in Architecture Context).
>
> This doc makes no process changes. Each recommended change is a candidate tracked
> item for `planned-work.md`; Robert assigns IDs and promotes them.

**How to read the findings.** Claude co-designed most of this process. To keep the
review honest, each finding names the pack item (`C1`–`C11`, sections 3–6) or Robert's
answer (`Q1`–`Q6`) it rests on. Where a finding rests on this session's own reading of
the code, it says so. Keyword counts (`C8`, `C9`) carry little weight. A gap the record
cannot measure is treated as a finding in itself.

<br/>

## Problem

Robert's goal, in his words (Q6): the work should produce "a tool that is usable,
easy-to-maintain, and would be respected if reviewed by a real person."

The May review judged the engine high quality but the tool unusable (no CLI or TUI), and
closed with one process warning: do not expand the process docs further. Five months
later the tool exists, Arc `A.001` has shipped three initiatives, and the process has
grown. Two things need a judgement:

1. **The tool.** Does it meet the three parts of Q6 today?
2. **The process.** Which parts of Discovery → Plan → Execution, the arc layer, the
   tracker and the decision records pay back what they cost? Which parts do not?

Robert added a third concern after the interview: sessions lose the big picture in the
detail and the project's own language (addendum, `C11`). He now asks each session for a
plain-language summary by hand.

<br/>

## Architecture Context

Pages read: `architecture-overview.md`. The review did not need a subsystem page, because
it judges shape and process, not one subsystem's design.

Code read for the tool verdict (this session, 2026-10-17):

- Package tree with production and test line counts per package
- `go vet ./...` and `go test ./...` (both pass), `.github/workflows/ci.yml`
- `internal/faction/engine/core.go`, `orchestrator.go`, `effect/effect.go`,
  `effect/registrar.go`, `internal/faction/errors/`
- `cmd/gm-toolkit/` and the TUI entry points
- `CODEBASE_REVIEW.md` sections 4, 5 and 10

<br/>

## Findings — The Tool

### T1. Usable: yes as a program, unknown as a tool

The May review's "viability as a tool today: zero" is resolved. A `gm-toolkit` binary
exists with `campaign`, `spatial` and `faction` commands, and `faction` opens the TUI
(pack section 6, rows 1 and 2). CI runs gofmt, vet, build and test on every push.

What the record does not show is use. Nothing in the pack says the tool has run a real
faction turn for Robert's campaign. No `B.` row comes from play. Robert's own Q6 answer
moves the goal from "a tool I use" to "a tool that would be respected." That move is
the finding. The May review's closing warning was that the code never meets a user. The
tool now exists, and the same question stands one level up: does it meet its one user?
This is Open Question 1.

### T2. Easy to maintain: yes, with one mark a reviewer would make first

**Strengths, verified this session:**

- Boundaries hold. The engine is headless and talks to its driver through two
  interfaces (collector in, observer out). The composition root is 80 lines and wires
  everything explicitly.
- Tests are real. Production Go 21,198 lines, test Go 19,362 lines. Engine packages
  carry more test than production code (`effects/` 1,376 vs 2,627; `actions/` 1,705 vs
  2,624; `hooks/dispatch` 419 vs 1,090).
- The effect registry validates every data entry at engine start and refuses a second
  handler for the same definition instead of overwriting it. Rulebook errors fail at
  load, not mid-turn.
- Zero `TODO` or `FIXME` markers. 33 `// deliberate:` fence-signs mark the places a
  future reader would be tempted to change.
- No `fix:` commit has landed straight on `main` since 2026-05-12 (section 4).

**The mark.** `orchestrator.go` is 823 lines. It holds 26 copies of the same eight-line
block: test `IsRecoverable`, warn and continue, or log and return. `IsRecoverable`
matches only three sentinels, all from action selection (`internal/faction/errors`).
So the copies wrapped around `state.Save`, `applyAndRecord`, `Turn.Advance`,
`TickMovementOrders` and `ApplyBookkeeping` can never take the recoverable branch. The
code reads as if each call has its own error policy. In fact there is one policy: a
GM cancel from a collector is recoverable, everything else is fatal. The May review
called the orchestrator "properly factored" (4.3). Movement, granted movement and
purchase effects have been added since. A senior reviewer would open this file first
and mark this first. Recommendation N3.

**Smaller marks:**

- The interface side is untested. `tui/adapter` is 1,176 lines with no tests, and it
  is the only implementation of the engine's two interfaces. `wizard` 511 and the TUI
  root 481 have none either. Engine-side coverage carries the project; the seam does
  not. Recommendation N7.
- The package comment in `core.go` lists `dispatch/` as a direct child of `engine`; it
  lives under `hooks/`. The comment sits below the imports, so Go does not treat it as
  the package doc. Small, and the kind of drift the process says it catches.
- Two findings survived more than one review. `liveDefendersOnWorld` was flagged unused
  in May (A4) and again in October (`R.021`); still open. `RunCycle` re-registered
  hooks on every cycle from `v1.4.0` until the A.001.3 pre-merge review: 148 days and
  two pre-merge reviews in between (section 4). Diff-scoped reviews do not see code
  outside the diff. Recommendation N5.

### T3. Respected by a real reviewer: yes for the code, with a question about the paper

A reviewer would respect the engine. The boundaries, typed mutations, data-driven
effect registry, test depth and clean CI are above what a solo project usually carries.
The orchestrator would cost one paragraph of the review.

The same reviewer would stop at the docs tree. Initiative docs (discovery, plans, arcs)
total about 48,000 lines against 21,000 lines of production Go. A.001.3 shipped 9,283
doc lines for 3,299 production lines (ledger). Across the arc, doc lines per initiative
went 321 → 4,598 → 9,283. The reviewer would not call the docs bad. They would ask
where the hours go, and whether the volume is a cause of the drift it keeps finding
(section 5). That question is the bridge to the process findings.

<br/>

## Findings — The Process

The process has two kinds of element. **Catching** elements find mistakes: re-grounding,
pre-merge review, Decision Records that graduate to architecture pages. **Structuring**
elements shape the work: session boundaries, the model table, the arc layer, templates,
tracker stages. The evidence treats them very differently.

### P1. The catching elements earn their cost

- **Pre-merge review is the highest-yield step.** Every recorded review found work: 6
  (event hooks), 13 (documentation overhaul), at least 11 (A.001.2), 11 (A.001.3). Fix
  rates where tallied: 6/6, 7/13, 7/11 (section 4).
- **Re-grounding has a catch record.** Robert's Q4 says it "has found a lot of potential
  misses." The pack agrees: 13 Decision Record entries name re-grounding, 10 name stale
  or drifted docs, and the documentation overhaul's re-verify step alone caught seven
  stale references (`C9`). All four A.001.3 drift seeds were found by the process, not
  by a user (section 5).
- **Architecture pages are read.** 16 active files link into `docs/architecture/`; the
  tree has 11 commits since `v1.11.0`, and the Key Decisions step changed it in each
  of the last three releases (`C6`, `C10`). The May review's "no CLI, stale docs"
  complaint is gone (section 6, done on 6 of 7 recommendations).
- **`planned-work.md` is the working surface.** Robert names it first among the docs
  he goes back to (Q2). It is the most-edited and most-linked doc in the tree: 33
  commits, 15 since `v1.11.0`, 13 inbound links (`C6`). The tracker is pulling its
  weight; the one stale cell (section 5) is a compliance miss, not a design one.
- **Defect escape is low.** Four fix commits ever landed straight on `main`, none since
  May. Two bugs were found by later initiatives in five months (section 4).

### P2. Half the structuring elements are already dead in practice

- **One session per commit.** Robert skips it (Q1), would drop it first (Q5), and the
  branch record shows five to seven commits inside two-hour windows for A.001.1 and
  A.001.2 (`C1`). The exception lives only in a memory file; `CLAUDE.md` and
  `session-modes.md` still state the rule (`C3`). The written process and the practised
  process disagree.
- **The model table.** Every commit since June carries an Opus trailer; zero Sonnet
  (`C7`). Robert: "I basically just stay on Opus" (Q3). The table and the "prompt to
  switch at each mode transition" rule produce a prompt Robert declines every time.
- **Check-in cadence.** Robert runs auto mode "basically the whole time since I would
  really only ever approve changes" (Q3). `CLAUDE.md` item 6 asks for regular check-ins
  and plan confirmation before any series of tasks; 18 of 42 memory files mention
  asking or confirming (`C8`, weighted lightly). A confirmation that is always granted
  is ceremony. The design calls still need a question; the "may I proceed" ones do not.
- **Completed docs.** 81 files; one is named by any active doc. No A.001.x doc cites an
  earlier initiative's completed doc (`C4`). Robert: "I don't ever look back at the
  completed docs, but that's kind of by design?" It is by design. The finding is that
  the pre-merge step that moves them is pure filing, and was skipped in 9 of 12
  releases before the numbered checklist (`C10`).
- **Arc artifacts, partly.** The arc plan is alive: 4 commits, last on 2026-10-17, and
  12 citations from the A.001.1 plan. The arc concept (31 lines, one commit) and arc
  discovery (one commit) are not cited by any effort plan (`C5`). Robert: "I also
  don't know how much the arc based artifacts actually pay off." The plan pays off.
  The concept does not earn a separate file.
- **`docs/contributing/`.** Two commits ever, one since `v1.11.0`, two inbound links
  (`C6`). Four files, 881 lines, for a project with one contributor.

### P3. Process compliance, not process design, is what slips

Robert, Q4: "what consistently slips through is actual process following, a step will
be skipped or done in a wrong order sometimes." The pack bears this out, and shows what
fixed it:

- Own docs left outside `completed/` at the tag: 9 of 12 releases before the numbered
  checklist entered `CLAUDE.md` (2026-06-14); 0 of 3 since (`C10`).
- Four memory entries exist only to restate a process step after a correction ("Do only
  the asked step", "Don't commit until told", "Hand-off session scope wins", "Pre-commit
  plain-language summary").
- The A.001.2 pre-merge review doc was never committed. Its finding list survives only
  as three fragments in other files (`C2`). The `planned-work.md` status cell for
  A.001.3 was wrong for 110 days (section 5).

Reading: the steps that are a numbered list in `CLAUDE.md` get done. The steps that
live in prose, memory, or a hand-off prompt get missed. Hand-offs are not committed at
all, so what a session was told at start is not recoverable (`C2`). `D.001` already
targets one of these gaps (review findings written into the plan).

### P4. Doc volume is the mechanism behind the big-picture loss

This is the one finding that joins the tool and process halves.

Robert's addendum: sessions "get hyper focused on the specific details (and the
technical / project language used)" until "the 'concept' is lost." He fixes this by
asking for a plain-language summary of "what we are doing, why, and how it relates to
the project / plan as a whole."

The pack puts a number next to that worry. A.001.3's session inputs were 9,283 lines of
discovery and plan, across four plan files, plus the arc plan (`C11`, ledger). A session
that must read that much reads narrowly, and reads in the project's own vocabulary,
because that is what the files are written in. The standing memory about plain language
dates from June (`C11`); the pre-commit summary memory from October covers "what changed
and why" but not "how it fits," the third part of Robert's check.

The May review's seventh recommendation was "Don't expand the process docs further."
Since then: `session-modes.md` 321 → 437 lines, templates 2 → 5, arcs and efforts added,
and per-initiative doc volume up 29× across the arc (section 6, ledger). The process did
the opposite of the one process recommendation it was given, and the cost shows up as
the exact problem Robert raised.

This is not a case for fewer decisions. The Decision Records are where the catches in P1
are recorded. It is a case for a size budget on plan prose, and for a short plain
concept at the top of every plan and discovery doc, so the summary Robert asks for by
hand is written once and read every session.

### P5. Judgement

Against Q6:

| Part of the goal | Verdict | Basis |
|------------------|---------|-------|
| Usable | Exists; use in play not shown | T1 |
| Easy to maintain | Yes; the orchestrator is the one real mark | T2 |
| Respected by a real reviewer | Yes for the code; the doc-to-code ratio would draw the first question | T3 |

On the process: the catching half earns its cost and should stay. The structuring half
has grown past use. Two of its rules are already ignored, the completed-docs filing and
the arc concept produce artifacts no one reads, and the doc volume is working against
the big-picture goal. The process does not need another flow or template. It needs
fewer words, a size budget, and its written rules brought level with its practised ones.

<br/>

## Recommended Next Steps

Each row is a candidate `planned-work.md` item. IDs are Robert's to assign. None of these
is done in this session. Ordered by the gap it closes, not by effort.

| # | Item | Type | Closes | Shape of the change |
|---|------|------|--------|---------------------|
| N1 | Bring the written session rules level with practice | docs | P2, P3, Q1, Q3, Q5 | Execution session scope is set by the hand-off, with one commit as the default when none is given; delete the model table in favour of one line (Opus default, Fable for arc and process sessions); delete the "prompt to switch" rule; narrow `CLAUDE.md` item 6 so check-ins are for design calls and scope changes, not for permission to start the next planned task. Edit `CLAUDE.md` and `session-modes.md` only; remove the memories the edit makes redundant |
| N2 | Plain concept header and prose budget for plans and discovery docs | docs | P4, C11 | One required section at the top of every plan and discovery doc: what this does, why, how it fits the arc and the project, in plain words, under 15 lines. One line in the plan template setting a target size per effort plan. No new template, no new flow |
| N3 | Orchestrator error-handling consolidation | refactor | T2 | Collapse the 26 recoverable-or-fatal blocks; make "recoverable" a property of collector calls only, since no other call can return the three sentinels. Target under 500 lines with no behaviour change. Run with the existing orchestrator tests as the guard |
| N4 | Commit the hand-off | docs | P3, C2 | The closing step of a session writes its hand-off for the next session as the last entry of the plan's Decision Record (or a `Next session` line beneath it). Hand-offs then live in the repo, not in a prompt. Pairs with `D.001` |
| N5 | Whole-tree review at arc boundaries | docs | T2, P1 | Diff-scoped reviews missed a bug outside the diff for 148 days. Add one line to the arc-level flow: a whole-codebase review runs when an arc's serial spine completes. This session is the first instance. No per-initiative change |
| N6 | Fold the arc concept into the arc plan; drop the `completed/` move for discovery docs | docs | P2, C4, C5 | The arc concept becomes the first section of the arc plan. Discovery docs stay where they are after merge and the pre-merge checklist loses one filing step; plans still move because they carry provenance. Template count goes 5 → 4 |
| N7 | Tests for the TUI adapter | feature | T2 | The adapter is the only implementation of the engine seam and has none. Cover the collector and observer paths with the engine's existing test harness. Low priority; file in Backlog with a trigger of "next interface initiative" |

Not recommended: a new review template, a new flow type, or any addition to
`session-modes.md` beyond the one-line changes above. The May warning stands.

<br/>

## Open Questions

1. **Has the tool run a real faction turn for the live campaign?** The pack has no
   evidence either way (T1). If not, the next initiative after this review should be
   that cycle, before `A.001.4`, with every rough edge filed as a `B.` or `F.` row. If
   yes, say where the record of it is, so the next review can measure use.
2. **Does `D.001` absorb N4?** Both write session outputs into the plan's Decision
   Record. One item or two is Robert's call.
3. **What is the plan-size target in N2?** The ledger gives the range: A.001.1 shipped
   on 321 lines, A.001.3 on 9,283. The right budget is a Plan-session question, not a
   review question.
