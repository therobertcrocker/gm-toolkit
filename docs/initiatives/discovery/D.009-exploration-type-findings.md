# Exploration Type — Findings

> **Initiative:** `D.009` (define the Exploration `E.` type), pre-work.
> **Purpose:** a running log of what worked and what did not in the first explorations.
> Each exploration adds one section when it closes. `D.009` reads this log, then writes
> the `E.` flow in `session-modes.md` and the template.
> **Status:** collecting. One exploration logged (`E.002`).

## Conflicts to settle in `D.009`

- **Template location.** The `D.009` row says "a template under
  `docs/initiatives/explorations/`". `CLAUDE.md` says doc templates live in
  `docs/process/templates/`.
- **Close-out.** `CLAUDE.md` pre-merge step 9.4 moves discovery and plan docs to
  `completed/`. For `E.002`, Robert chose to keep the doc in place and to keep the
  planned-work row (Status `Complete`) and its write-up. The `E.` flow must say which
  rule applies.

<br/>

## `E.002` — Data-driven tags and goals

Doc: [`E.002-data-driven-tags-goals-exploration.md`](../explorations/E.002-data-driven-tags-goals-exploration.md).
Two sessions on Opus, as planned. Session 2 is observed directly. Session 1 is read
from its doc and its hand-off.

### Doc shape

The sections, in order: In Plain Words, Questions, Current State, Existing Patterns,
Options, Recommendation, Follow-on Rows.

- **Header blockquote** — `Initiative`, `Feeds` (the paused doc that waits on the
  answer), `Status`. The `Feeds` link told session 2 where the answer had to land.
- **In Plain Words was written last.** It was a placeholder until session 2, and it
  became a short answer (what, short answer, why, how it lands), not a problem
  statement. A template can say "write last".
- **Questions held the criteria.** The line "cost is not a deciding factor; the aim
  is one consistent way" was used in every option comparison. A template can ask for
  the deciding criteria in this section.
- **Current State and Existing Patterns cite `file:line`.** Session 2 still re-read
  the code before it wrote options (registrars, roll dispatch, loader, reactor
  dispatch). The citations made that fast. The re-read also found new facts (see
  Findings during options).
- **Options numbered per question** (`T-1`…`T-4`, `G-1`…`G-3`, `I-1`…`I-3`). The
  Recommendation and the follow-on rows cite them by number. A baseline option (keep
  today's shape) helped as a reference.
- **A shared line above the options.** Where all options of one question share a
  precondition (IDs: "in every option, no Go code names a tag or goal ID"), it is
  stated once above them.
- **A Terms block at the top of Options.** The session 1 hand-off flagged a naming
  trap ("event scope" against "declaring source"). Session 2 asked Robert first
  (`holder`), then defined both words above the options.
- **Recommendation sub-sections.** "How it lands" (new rows, row changes, order,
  amendments to the arc-level decisions) and "Rules calls to carry" (the open ones,
  for Robert). Both were needed.
- **Follow-on Rows table** — `ID`, `Item`, `Type`, `Detail`, the same columns as
  `planned-work.md`. It also held row changes and closures, not only new rows.
- **Depth.** Options went to `####` (question, then option). The tag-fit tables made
  Options the longest section.

### Flow

- **Session split** — audit (Current State, Existing Patterns), then options and a
  recommendation. The split held. The hand-off between them carried: key audit
  findings, the naming trap, unverified claims, and the reminder that rules readings
  are Robert's.
- **Unverified claims in the hand-off were useful.** Session 2 checked the XP column
  against `swn-faction-mechanics.md`. It did not run the app; rows found from code
  say "from reading the code, not a run".
- **Rules calls during the session.** One rules call (Scavengers) was asked and
  decided in the session through `AskUserQuestion`. The calls that stayed open went
  to "Rules calls to carry".
- **Rows are filed after review.** The doc first said "proposed; not yet filed".
  After Robert's review, a separate step filed the rows in `planned-work.md`, then
  the doc's note changed to "filed".
- **The paused doc was not edited.** The answer for the paused `A.001.4` discovery
  went into its planned-work row. The `A.001.4` session applies it. This kept the
  discovery-doc rule (no edit after the fact) intact.
- **Close-out and merge.** Status line: in progress → written, waits on review →
  complete (reviewed). Branch `exploration/<name>`, merged into `main` with `--no-ff`.
  No code review and no version tag, because the branch changed docs only.

### Findings during options

Writing the options found facts that the audit did not:

- A bug: `B.042` (Scavengers pays for other factions' fights). A side agent found
  the same bug at the same time.
- A rules gate the code does not have: `G-011`'s "Wealth must increase before this
  goal can be selected again".
- The root cause of `B.013`: `UpdateProgress` gets no action name.

A template needs a place for such facts in the doc (they went into Options and
Follow-on Rows), and the flow should say how to file them: as rows, with rules calls
asked when they come up.

### Open for `D.009`

- Is In Plain Words always written last, or is a first draft written after the audit?
- Should a Terms block be required, or added only when a naming trap shows?
- Does a "Findings during options" heading earn its own section?
