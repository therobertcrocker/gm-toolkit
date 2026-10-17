# Handoff — Prep the Project & Process Review

> **Purpose of this doc:** Robert wants a review of the whole project: how it is
> structured, organized and set up (an architectural review), and whether the process
> earns its cost (is the Discovery → Plan → Execution machinery worth it?). The review
> itself runs later, in a **Fable** session. This doc is the plan for the **prep
> session** (model: **Opus**), which gathers the evidence that review reads first.
> Delete this doc once the review doc exists.

<br/>
<br/>

# Situation

- Arc `A.001`'s serial spine is done (`A.001.1`–`A.001.3` merged; `v1.13.0`). This is a
  natural pause point, and `A.001.3` is a fresh, well-documented case study.
- The `A.001.3` pre-merge session (2026-10-17) surfaced process friction. One outcome is
  already tracked: backlog `D.001` (split the pre-merge review from execution).
- A prior review exists: `docs/initiatives/implementation/completed/CODEBASE_REVIEW.md`
  (Claude, 2026-05-22, on `feature/movement-redesign`). It is the baseline.
- **Bias risk:** Claude co-designed most of this process. The review must rest on
  measured evidence and on Robert's own experience, not on Claude's opinion of its own
  process. That is why the evidence is gathered first, in a separate session.
- **Slim scope:** the prep is kept small on purpose. A heavy evidence pass would be the
  same cost the review sets out to test. Robert's interview comes first and sets which
  facts to gather (see Decision Record).

<br/>
<br/>

# Session Plan

Do the steps in this order. Check in with Robert at the end of each step.

1. **Tracker row.** Ask Robert whether `D.002` absorbs `D.001`. Then add `D.002` —
   "Project & process review" — to the **Docs** table in `planned-work.md` Backlog.
   Trigger: `next process session (Fable)`. Record his answer in the row's Detail.
2. **Interview.** Ask the questions in Evidence Pack section 1 (AskUserQuestion where a
   question has natural options, prose otherwise). Record the answers in his own words.
3. **Claim list.** From the answers, list each claim the record can test (for example,
   "I never reopen discovery docs" or "plans change a lot during execution"). Confirm
   the list with Robert before gathering.
4. **Evidence.** Write `docs/initiatives/discovery/project-review-evidence.md`: sections
   1–6 below.
5. **Pre-commit summary**, then commit on Robert's go (`docs:` prefix). Push only if
   asked.

**Out of scope:** findings, judgements, recommendations, process edits, code changes,
scaffolding the review doc. The prep session reports facts; the Fable session
interprets them.

<br/>
<br/>

# Rules for the Evidence Pack

- **Facts only.** No adjectives that judge ("heavy", "worth it", "healthy").
- **Reproducible.** Every number states the command or file that produced it.
- **Verified.** A claim about the code is checked against the source, not carried from
  a doc (memory: "Arch docs reflect source").
- **Say "not measurable" rather than guess.** Session counts and time spent per session
  are not in git; say so and use what proxies exist (commit dates).
- **No extra sections.** Gather only what sections 1–6 ask for. If a fact is
  interesting but no section asks for it, leave it out.
- Use the LSP tools for code questions where they fit (CLAUDE.md item 5).

<br/>
<br/>

# Evidence Pack — Sections

## 1. Robert's Input

Ask, and record verbatim:

1. Which process steps do you skip or shortcut in practice?
2. Which artifact have you gone back to and been glad it existed? Which have you never
   reopened?
3. Where did a session feel slow or ceremonial?
4. What broke, or nearly broke, that the process caught? What slipped through?
5. If you dropped one process element tomorrow, which one?
6. What does "the project works" mean to you right now — what do you use it for?

<br/>

## 2. Claim Checks

One entry per claim from Session Plan step 3: the claim (quoted), the evidence for or
against it, and the command or file that produced the evidence. When the record cannot
test a claim, write "not measurable" and say why.

<br/>

## 3. Per-Initiative Ledger

One row per completed initiative. Find them from the merge commits on `main`, tags, and
the `completed/` folders. Columns:

- start date (first branch commit) → merge date; tag
- discovery + plan doc lines (all effort plans included)
- code diff at merge: production lines and test lines added/removed
- Decision Record entry count; entries marked *Reversed* or *Amended*
- `fix:` commits that landed after the last `feat:` of the initiative

Add a column only when a claim in section 2 needs it.

<br/>

## 4. Rework and Catch Signals

- Review findings where recorded (interim and pre-merge reviews; the `A.001.3` plan's
  Decision Records 24–28 list the last one): found / fixed / deferred / rejected.
- Bugs a review or test caught before merge vs bugs found after merge (`B.` rows whose
  Detail names an initiative; `fix:` commits on `main` after a merge).

<br/>

## 5. Doc Drift

Instances where a doc said something the code or tracker no longer did. Seeds from
`A.001.3` (verify each, then search for more in Decision Records — grep "stale",
"became false", "reconciled"):

- `planned-work.md` A.001.3 row status still read "`A.001.2` (relocate)" at merge.
- `effect-mutation.md` "Asset features" stale since Effort 1 (Effort 5 plan, DR 5).
- Two `world-movement.md` Key Decisions became false (Effort 5 plan, DR 4).
- `hooks.md` Cat 4 list omitted attack eligibility.

<br/>

## 6. Baseline Comparison

For each recommendation in `CODEBASE_REVIEW.md`: what it said, and the evidence of what
happened since (commit, file, or "no change found"). Status words limited to: *done*,
*partly*, *no change found*.

<br/>

## Dropped from the Original Handoff

- **Inventory** (Go lines per package, doc lines per folder, process-surface line
  counts) — no review decision depends on it
- **Tracker flow** (row counts per section at each tag) — no review decision depends
  on it
- **Ledger: commits by conventional type** — the `fix:`-after-last-`feat:` column
  carries the rework signal

<br/>
<br/>

# Read First

`CLAUDE.md` → `docs/process/session-modes.md` → `docs/dev_journals/faction-manager/planned-work.md`
→ `docs/architecture/architecture-overview.md` → `CODEBASE_REVIEW.md` (above). Read the
plan docs only where a ledger column or claim check needs them (they are long;
`A.001.3` alone is four files).

<br/>
<br/>

# Decision Record

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Slim the prep: drop inventory and tracker flow; trim the ledger | A full seven-section pass is the cost the review sets out to test; dropped sections feed no review decision |
| 2 | Interview first, then gather evidence | Robert's answers show which facts matter; gathering first collects numbers no question asks for |
| 3 | This doc is the plan for the prep session | One small session; a separate plan doc would duplicate it |
| 4 | `D.002` does not absorb `D.001` | Robert's call; `D.001` stays its own row |
| 5 | Interview answer 6 is the review's goal, not a claim; no facts gathered for it | Robert: it is "the actual piece of judgement that should come out of this" — the Fable session's call |
| 6 | Ledger keeps every row but does no extra digging; gaps say "not measurable" | Robert's call mid-gather; start dates for squash-landed initiatives would need per-doc history archaeology for no review decision |
| 7 | Ledger maps docs to initiatives by filename stem, not by git ancestry | The ancestry script (`git log --follow` per doc × merge-base per release) ran 12+ minutes without output; the stem rule is reproducible and states its own exceptions |
| 8 | Robert's big-picture concern goes in as a section 1 addendum plus claim C11, no new digging | Raised after the interview; it is first-hand experience, and three facts already gathered bear on it |

<br/>
<br/>

# After This Session

The **Fable** session runs the review: it reads the evidence pack, writes the review
doc (findings + recommended next steps, Discovery-style), and makes no process changes
itself. Changes it recommends become their own tracked items, like `D.001`.
