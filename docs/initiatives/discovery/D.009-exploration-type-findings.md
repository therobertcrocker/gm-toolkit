# Exploration Type — Findings

> **Initiative:** `D.009` (define the Exploration `E.` type), pre-work.
> **Purpose:** a running log of what worked and what did not in the first explorations.
> Each exploration adds one section when it closes. `D.009` reads this log, then writes
> the `E.` flow in `session-modes.md` and the template.
> **Status:** collecting. Three explorations logged (`E.002`, `E.004`, `R.029`).

## Conflicts to settle in `D.009`

- **Template location.** The `D.009` row says "a template under
  `docs/initiatives/explorations/`". `CLAUDE.md` says doc templates live in
  `docs/process/templates/`.
- **Close-out.** `CLAUDE.md` pre-merge step 9.4 moves discovery and plan docs to
  `completed/`. For `E.002`, Robert chose to keep the doc in place and to keep the
  planned-work row (Status `Complete`) and its write-up. `E.004` did the same. The `E.`
  flow must say which rule applies.
- **Branch.** The git rules say initiative work never commits on `main`. `E.002` ran on
  `exploration/<name>`. For `E.004`, Robert said no branch is needed. The `E.` flow must
  say if an exploration is exempt, since it changes docs only.

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

<br/>

## `E.004` — Unifying dice and RollHooks

Doc: [`E.004-unifying-dice-rollhooks-exploration.md`](../explorations/E.004-unifying-dice-rollhooks-exploration.md).
One session on Opus, by Robert's choice, so that it can be compared with `E.002`'s two.

### Doc shape

Same sections as `E.002`, in the same order. What we did, then whether it worked:

- **Questions held Robert's prompt in his words,** and then his sharpened version after
  review. It worked: the deciding criterion came straight from the prompt.
- **In Plain Words had a first draft after the audit,** then two rewrites as the answer
  changed. It worked, and it answers an open question from `E.002`: a first draft is
  useful, as long as it is rewritten last.
- **A Terms block at the top of Options** again. Robert changed it in review, which added
  table and target rolls. It carried real design weight, so it earned its place.
- **Options numbered per question** (`C-`, `P-`, `O-`) again, with a table comparing the
  `C-` options. Two of the options that won came from Robert's review (`C-5`, `P-4`), and
  each was added beside the others, not in place of them. This kept the reasoning visible.
- **A "what moves, what stays" table in an option** (`P-4`). Each item went in front of
  Robert with a reason. It worked for a call made of many small items.

### Flow

- **One session held the audit, options and review.** Context carried over, and the
  session needed no hand-off. Against that, no second session re-read the code with fresh
  eyes. A claim I had made in chat was wrong (each caller keeps its own tie rule), and
  only writing the options caught it. In `E.002`, the hand-off's list of unverified claims
  did that job.
- **One open question before writing,** on what prompted the row. The row was vague, and
  the answer changed the audit: from counting duplicates to asking what decides a roll's
  route. Worth keeping, with a vague row.
- **The review changed the answer.** Robert read Current State and proposed the roll
  types, which became the winning option. So a review checkpoint after the audit, before
  the recommendation hardens, is worth building in.
- **Rows are filed after review,** as in `E.002`. It worked again.
- **Long silent stretches.** The audit ran many reads between messages, and the session
  twice had to send an update. A flow could ask for a short update between audit steps.
- **A spot check after the answer found more.** Robert asked what had been left out, after
  the rows were filed. It found seven items, and all seven went in: the largest was the
  ability step's data shape. It also caught a second wrong claim: the doc said the fake
  rollers behave too differently to merge, and nobody had read them all. A spot check, with
  every item put to Robert, may be the fresh re-read that one session lacks.
- **The exploration led to a plan, not a discovery.** `R.028` skips Discovery and goes
  straight to Plan. The audit is done and the target shape is chosen, so none of the refactor
  flow's reasons for a discovery apply. `E.002` was different: its answer fed a discovery that
  was paused and waiting on it. The plan still re-reads the code, because the refactor flow
  treats an earlier doc as a hypothesis.
- **Close-out.** No branch, no merge, no code review and no version tag.

### Findings during options

Writing the options, and answering Robert's review, found facts that the audit did not:

- Attack, counterstrike and free attack share one tie test, so three named rules cover
  every contest.
- One call (`ability/stage.go:161`) hides three types of roll. Two ability dice sit in a
  `RollTable` that has no rows, and two one-row tables are really target rolls.
- Moving `rulebook/dice.go` into the engine closes an import cycle through `hooks`.

As in `E.002`, these went into Current State and Options. This is a second data point
that options writing finds facts, so a template can expect it.

### Open for `D.009`

- Is one session or two the default? One session worked here, on a smaller question
  than `E.002`'s. A rule could turn on the question's size.
- If one session is allowed, what replaces the fresh re-read that a second session gives?
  Is a spot check before close a required step?
- What comes after an exploration? It has now led to a paused discovery (`E.002`) and
  straight to a plan (`E.004`). The `E.` flow should say how to choose, and whether a
  discovery that follows one needs less.
- Do explorations run on a branch? See Conflicts.

<br/>

## `R.029` — Migrate data files to TOML 1.1

Doc: [`R.029-toml-1-1-migration-exploration.md`](../explorations/R.029-toml-1-1-migration-exploration.md).
One session on Opus. The first exploration that is a refactor's own Discovery: it has
no `E.` ID.

### Doc shape

Same sections as `E.002` and `E.004`, in the same order. What we did, then whether it
worked:

- **Filed under the initiative's ID, in `explorations/`.** Robert chose the exploration
  shape for `R.029`'s Discovery, then chose to file it with the explorations "so the
  provenance makes sense". No `E.` row was made. It worked: the plan and the `A.001.8`
  row both point at one doc.
- **The deciding criterion was asked before the audit,** as in `E.004`. "Best for a GM
  who edits by hand" decided every option, and it turned the audit toward error
  messages, not only line counts.
- **A question was added mid-session.** Robert's review of the audit asked about the
  machine-written files, then about leaving TOML for state. Both went in as question 6,
  marked as added after the audit, with options `S-0` to `S-3` and Robert's call
  recorded in place. It kept the scope change visible.
- **A Terms block** at the top of Options again (entry, part, inline table, dotted key).
- **Options numbered per question** (`S-`, `H-`, `G-`), plus "Lean" tables for calls made
  of many small items (house style, Reach), as `E.004`'s "what moves, what stays" table.
  It worked: each line became one question.
- **In Plain Words was drafted after the options,** not after the audit. The answer did
  not change after that, so one rewrite was enough.

### Flow

- **The audit ran the real code.** A throwaway program in the gitignored `scratch/`
  folder loaded the rulebook in 1.0 and in 1.1 and compared the results, then fed it
  typos and syntax slips. Reading the code alone would not have settled question 1.
- **The review checkpoint after the audit** (`E.004`'s suggestion) paid off again.
  Robert's two questions there grew the scope by one unit (`S-1`) and closed a larger
  idea (another state format) before it reached the options.
- **Calls went out as questions after the draft,** in three batches (four, four, one).
  All nine matched the lean. The one change was a row Robert chose not to file.
- **A number check before review caught a wrong claim.** The doc said one asset carries
  six headers; a count found nine. A cheap check of every number before the review is
  worth keeping.
- **Long silent stretches again.** The harness asked for an update twice during the
  audit, even with updates between steps.
- **It led straight to a plan,** as `E.004` did for `R.028`.
- **No branch of its own.** It ran on `feature/data-driven-goals`, the branch of the
  initiative it amends, at Robert's call.

### Findings during options

- TOML 1.0 and 1.1 mix in one file, so `A.001.8` needs only the style settled, not
  `R.029` executed. This removed an order between the two initiatives.
- A measured line (the longest gate, 119 columns) decided the column limit.
- "`[[effects]]`" names a concept in twelve live files, not only a syntax.

### Open for `D.009`

- When an exploration serves one initiative, does it take that initiative's ID (as
  here) or an `E.` ID (as `E.004` for `R.028`)?
- Is a refactor's audit-first Discovery an exploration under another name? If yes, the
  refactor flow can point at the `E.` template.
- Does `R.029` count toward the row's "after 3 explorations" trigger?
