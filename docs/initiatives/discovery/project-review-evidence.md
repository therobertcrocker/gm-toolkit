# Evidence Pack — Project & Process Review

> **Purpose:** facts for the `D.002` review (Fable). Facts only, no judgements. Every
> number names the command or file that produced it. Plan and rules:
> [project-review-prep-handoff.md](project-review-prep-handoff.md).
> Gathered 2026-10-17, at `v1.13.0` (`main` = `cb11c74`).

**Limits that apply to every section:**

- Sessions, session length and time spent are not in git. Commit times are the only
  proxy.
- Eight releases landed on `main` as one squashed commit (`v1.3.0`–`v1.8.0`, `v1.10.0`,
  `v1.11.0`). Their branch history is gone, so start date, commit types and
  `fix:`-after-`feat:` are not measurable for them.
- A.001.3 squashed each effort into one commit (memory: git commit cadence). Its branch
  commit times are squash times, not work times. The pre-squash history of Effort 5
  survives on `backup/effort-5-presquash`.
- Keyword counts (Decision Records, process rules) match words, not meaning. Each
  states its pattern so it can be re-run or read by hand.

<br/>
<br/>

# 1. Robert's Input

Answers recorded verbatim.

**1. Which process steps do you skip or shortcut in practice?**

> The only one I've been skipping so far (that I am aware of) is seperate sessions per
> commit - I have been trying out how doing multiple commit works in one session, half
> as a test of the current Opus 5.5 capabilities, and also because I'm worried things
> get lost between sessions.

**2. Which artifact have you gone back to and been glad it existed? Which have you never reopened?**

> Planned Work, contributing/ and architecture/ I have all referenced or gone back to -
> I don't ever look back at the completed docs, but that's kind of by design? I also
> don't know how much the arc based artifacts actually pay off

**3. Where did a session feel slow or ceremonial?**

> When we try to switch models - I basically just stay on Opus at this point. Also, I
> have turned auto mode on basically the whole time since I would really only ever
> approve changes

**4. What broke, or nearly broke, that the process caught? What slipped through?**

> I think the re-grounding has found a lot of potential misses - and I think what
> consistently slips through is actual process following, a step will be skipped or done
> in a wrong order sometimes

**5. If you dropped one process element tomorrow, which one?**

> I would say the single session per commit

**6. What does "the project works" mean to you right now — what do you use it for?**

> The project works (at least right now) is less talking about the tool that we are
> building, and more the idea that what we are doing is producing a tool that is
> usable, easy-to-maintain, and would be respected if reviewed by a real person.

**Addendum — raised by Robert after the six questions:**

> One thing I am legitimattely worried about with this project (specifically during
> execution and planning sessions) is that it has become difficult for a session to
> truly understand what it is doing in the bigger picture - like we will get hyper
> focused on the specific details (and the technical / project language used) that the
> "concept" is lost. One way I've been handling this is to have the session give me a
> plain language summary of what we are doing, why, and how it relates to the project /
> plan as a whole - so I can confirm that we are not just lost in the weeds of our own
> design.

<br/>
<br/>

# 2. Claim Checks

Claims drawn from section 1, confirmed with Robert. Answer 6 is not a claim: it is the
review's goal (usable, easy to maintain, respected by a real reviewer). The Fable
session judges against it; this pack gathers no facts for it.

## C1 — Recent sessions make more than one commit (Q1, Q5)

- **A.001.3 Effort 5, pre-squash** (`git log --reverse --date=format:'%Y-%m-%d %H:%M'
  $(git merge-base feature/s-flag-effect-registry backup/effort-5-presquash)..backup/effort-5-presquash`):
  the plan doc and four code commits on 2026-10-14 between 17:14 and 18:57 (103
  minutes), then one more commit on 2026-10-17 05:29.
- **A.001.3 Effort 4, planned:** plan Decision Record 12
  (`s-flag-effect-registry-plan.md`) sets "Commits 1–3 … in one session and Commits
  4–5 … in another".
- **A.001.2** (`git log v1.11.1..feature/a-flag-ability-registry`): seven commits on
  2026-06-26 between 16:11 and 22:40.
- **A.001.1** (`git log 9d4c294^1..9d4c294^2`): five commits on 2026-06-17 between 15:25
  and 17:42.
- Whether commits close in time came from one session: not measurable.

## C2 — Things get lost between sessions (Q1)

- Hand-off prompts are not committed. What a session received at start: not measurable.
- Memory `feedback_handoff_session_scope.md` (modified 2026-10-12) records a
  correction. Robert: "Remember that the hand-off told you we are doing commits 1-3 in
  one session."
- `a-flag-ability-registry-plan.md` links `a-flag-ability-registry-pre-merge-review.md`.
  That file has no git history (`git log --all --
  '*a-flag-ability-registry-pre-merge-review*'` is empty). The A.001.2 review's finding
  list survives only as C5, C10 and C11 (plan and `915b861` body).
- The `planned-work.md` A.001.3 status cell stayed stale for 110 days across sessions
  (section 5, item 1).

## C3 — The one-commit-per-session rule is already loosened (Q5)

- The rule as written: `CLAUDE.md` Session Boundaries, "One session per Commit" (first
  in `f09e4cc`, 2026-05-22: `git log --reverse -S'One session per Commit' --
  CLAUDE.md`); `session-modes.md:97` "One session per commit"; `session-modes.md:126`
  "Execution (N sessions, one per commit)".
- The exception: memory `feedback_handoff_session_scope.md` (modified 2026-10-12) — a
  hand-off or Decision Record that sets a multi-commit session overrides the rule.
- Committed docs that set a multi-commit session: one found, A.001.3 plan Decision Record
  12 (`grep -n -iE 'one session' s-flag-*.md a-flag-*.md`).
- `CLAUDE.md` and `session-modes.md` have no text on the exception.

## C4 — Completed docs are not reopened (Q2)

- Completed docs: 81 (`ls docs/initiatives/{discovery,implementation}/completed`).
- Active docs (`docs/architecture`, `docs/process`, `docs/contributing`,
  `planned-work.md`, `dev-journal-factions.md`, `CLAUDE.md`) that name a completed doc
  by filename: 1 doc named (`documentation-overhaul-plan.md`, in 2 files).
- Files outside `completed/` that contain a `completed/<name>.md` path: 2 (the A.001 arc
  plan and the prep handoff).
- Completed docs that name a completed doc of a different initiative: 28 links. (31 by a
  filename-stem rule; 3 are the same initiative on inspection.) None of the A.001.x docs
  (`effect-keyed-*`, `a-flag-*`, `s-flag-*`) names an earlier initiative's completed doc.
- Completed docs edited after their move into `completed/`: 1
  (`goal-engine-implementation.md`).
- Robert's reading and Claude's reading during sessions: not measurable.

## C5 — Arc artifacts may not pay off (Q2)

- Size (`wc -l docs/initiatives/arcs/*/*.md`): A.001 arc 776 lines (concept 31,
  discovery 366, plan 379); tui-rebuild arc 383 lines (discovery 231, plan 152).
- Edits (`git log -- <file>`): A.001 arc plan 4 commits, last 2026-10-17; A.001 concept
  and discovery 1 commit each; tui-rebuild docs 1 commit each.
- Arc plan decisions: AP-1 to AP-7. AP-6 (from A.001.2 planning) and AP-7 (from A.001.3
  Discovery) came after sign-off. Commit `ce2d790` says "slice into five initiatives";
  the plan body now reads "The arc in six initiatives".
- Mentions of the arc docs (`arc-plan`, `arc-discovery`, `arc-concept`, `AP-N`) in A.001.x
  docs: `effect-keyed-a-dispatch-plan.md` 12; `a-flag-…-discovery.md` 4,
  `a-flag-…-plan.md` 11, each A.001.2 effort plan 0; `s-flag-…-discovery.md` 12,
  `s-flag-…-plan.md` 3, each A.001.3 effort plan 0.
- Arc reconciliation items: arc discovery surfaced 4; arc plan routes 6 to initiatives
  (5 to A.001.3, 1 to A.001.4).
- Whether the arc docs changed an outcome: not measurable.

## C6 — Planned Work, `contributing/` and `architecture/` get used (Q2)

Commits measure edits, not reads. Inbound links count active files only (`grep -rlF
<path> docs CLAUDE.md`, excluding the target folder and `completed/`).

| Doc | First commit | Commits touching | Since `v1.11.0` | Linked from |
|-----|--------------|------------------|-----------------|-------------|
| `planned-work.md` | 2026-05-09 | 33 | 15 | 13 files |
| `docs/architecture/` | 2026-06-14 | 12 | 11 | 16 files |
| `docs/contributing/` | 2026-06-14 | 2 | 1 | 2 files |
| `session-modes.md` | 2026-05-14 | 7 | 1 | — |
| `dev-journal-factions.md` | 2026-05-09 | 5 | 0 | — |

Non-merge commits in `v1.11.0..HEAD`: 34.

## C7 — Work runs on Opus whatever the model table says (Q3)

`Co-Authored-By: Claude <model>` trailers by commit month (`git log --all`, one count
per commit):

| Month | Opus | Sonnet | No Claude trailer |
|-------|------|--------|-------------------|
| 2026-04 | 3 (4.7) | 61 (4.6) | 21 |
| 2026-05 | 7 (4.7 ×6, 4.8 ×1) | 51 (4.6) | 50 |
| 2026-06 | 22 (4.8) | 3 (4.6) | 3 |
| 2026-10 | 17 (5.5) | 0 | 0 |

- The model table entered `CLAUDE.md` in `c2efe98` (2026-05-11). It sets Sonnet for
  "End-of-phase mechanical docs & pre-merge checklist". Memory
  `feedback_model_selection.md` records Opus for execution as Robert's choice on
  2026-10-11.
- A trailer names the model at commit time. Switches inside a session and squashed
  commits' original models: not measurable.
- Auto-mode use: not measurable.

## C8 — Confirmation prompts are friction (Q3)

Pattern (`grep -ciE`): `\b(ask|asks|asked|confirm|confirmation|check-in|check
in|go-ahead|approval|approve)\b|/model|prompt (robert )?to switch|until (told|robert)|on
robert.s go`.

- `CLAUDE.md`: 6 matching lines of 84 (items 2 and 6, the commit item, docs-conflict
  rule, Implementation mode, Model Selection).
- `session-modes.md`: 4 matching lines of 437.
- Memory: 18 of 42 files contain a match.
- Friction felt: not measurable.

## C9 — Re-grounding catches misses (Q4)

- Decision Record entries that contain "re-ground"/"reground": 13 of 139 entries across
  the 13 plans with an execution Decision Record (documentation-overhaul effort 3: 8;
  A.001.2 overview: 2; A.001.2 effort 1: 1; A.001.3 overview: 1; documentation-overhaul
  overview: 1).
- Entries that contain "stale", "became false", "no longer true" or "drift": 10.
- `documentation-overhaul-effort-4-plan.md`: its re-verify-against-source step "caught
  four stale" references in one place and "three stale references" in another.
- A.001.2 overview Decision Record: re-grounding "surfaced two design decisions" before
  Effort 2 and "surfaced that two of the three assets cannot ship" before Effort 4
  (became AP-6).
- Misses that re-grounding did not catch: section 4, "found after merge".

## C10 — Process steps get skipped or done out of order (Q4)

Landing diffs since `session-modes.md` existed (2026-05-14), checked per release
(`git diff --name-status -M <first-parent> <landing>`; `git ls-tree -r <tag>`):

| Release | Tagged | `planned-work.md` in diff | Own docs left outside `completed/` at its tag | `docs/architecture/` files in diff |
|---------|--------|---------------------------|-----------------------------------------------|------------------------------------|
| `v1.3.1` sub-engine | yes | yes | 3 (alignment discovery + plan, self-bootstrap plan) | 0 |
| `v1.4.0` asset movement | yes | yes | 0 | 0 |
| `v1.5.0` logging & errors | yes | yes | 4 | 0 |
| `v1.6.0` tui-foundation | yes | yes | 2 | 0 |
| `v1.7.0` campaign manager | yes | yes | 2 | 0 |
| `v1.8.0` spatial map CLI | yes | yes | 2 | 0 |
| `v1.9.0` tui-manage | yes | yes | 2 | 0 |
| `v1.10.0` tui-turn | yes | yes | 3 | 0 |
| `v1.11.0` documentation overhaul | yes | yes | 1 (the review doc) | 19 |
| `v1.11.1` A.001.1 | yes | yes | 0 | 2 |
| `v1.12.0` A.001.2 | yes | yes | 0 | 6 |
| `v1.13.0` A.001.3 | yes | yes | 0 | 6 |

- `docs/architecture/` did not exist before `v1.11.0`. The Key Decisions promotion step
  entered `CLAUDE.md` in `857412e` (2026-06-14).
- `action-result-panel-discovery.md` has sat outside `completed/` since `v1.10.0`
  (2026-06-09). `planned-work.md` names it only inside the `F.018` row.
- A.001.3: the `planned-work.md` status cell was stale at the pre-merge review (section
  5, item 1). The A.001.2 review doc was not committed (C2).
- Memory index entries that state a process-step rule after a correction (read from
  `MEMORY.md`): "Do only the asked step", "Don't commit until told", "Hand-off session
  scope wins", "Pre-commit plain-language summary".

## C11 — Sessions lose the big-picture concept in detail and project language (addendum)

- Memory `communication_style.md`, "Plain language" (standing rule 2026-06-16): names a
  "tendency to drift verbose and pack complicated ideas into compressed shorthand" and
  that "Robert can lose the thread".
- Memory `feedback_pre_commit_summary.md` (2026-10-11): asks for a plain-language summary
  of "what changed and why" before each commit. Its text does not cover how the work
  relates to the plan or the project, the third part of Robert's check.
- Discovery + plan doc lines for the A.001 initiatives (section 3): A.001.1 321, A.001.2
  4598, A.001.3 9283.
- Whether a session understood the concept: not measurable.

<br/>
<br/>

# 3. Per-Initiative Ledger

One row per release or merge on `main`'s first-parent line. Version-only tags `v0.1.0`
to `v0.3.0` (no Go change) are left out.

**Sources:**

- Dates, commits, Go diff: for a merge `M`, branch = `M^1..M^2` and diff =
  `git diff --numstat M^1 M -- '*.go'`, split on `_test.go`. For a single commit `C`,
  diff = `C^..C`. "n/m" = not measurable (squashed).
- Doc lines: `wc -l` on the initiative's discovery and plan docs in `completed/`,
  matched by filename stem.
- Decision Record (DR): entries in each plan's `## Decision Record — Execution` section
  (entry starts: `N. **`, `- **`, `**DR-N`). "rev" = entries that contain
  "revers"/"supersed"; "amd" = entries that contain "amend". The section exists only from
  `v1.11.0`. Earlier decisions went to the frozen decisions log.
- Fix after feat: `fix:` commits after the last `feat:` on the branch.

| Initiative | Tag | Start → land | Doc lines | Prod Go | Test Go | DR (rev, amd) | Fix after feat |
|------------|-----|--------------|-----------|---------|---------|---------------|----------------|
| cli-scaffold | — | 04-17 → 04-18 | — | +715/−173 | +105/−0 | — | 0 |
| faction-create-wizard | — | 04-19 → 04-19 | — | +484/−142 | 0 | — | 0 |
| internal code review | — | 04-19 → 04-19 | — | +69/−64 | +204/−0 | — | no feat |
| faction-list | — | 04-19 → 04-19 | — | +19/−3 | 0 | — | 0 |
| turn-engine cleanup | — | 04-21 → 04-21 | — | +26/−20 | +13/−13 | — | no feat |
| turn-command | `v0.4.0` | 04-22 → 04-22 | 164 | +434/−77 | +115/−22 | — | 0 |
| action-engine | `v0.5.0` | 04-23 → 04-23 | 248 | +339/−118 | 0 | — | 0 |
| history engine | `v0.6.0` | n/m → 04-23 | — | +180/−48 | +169/−0 | — | n/m |
| code review 2 | — | 04-23 → 04-23 | — | +84/−67 | 0 | — | no feat |
| simple-actions | `v0.7.0` | 04-24 → 04-24 | — | +713/−58 | +41/−41 | — | 0 |
| attack-action | `v0.8.0` | 04-25 → 04-25 | — | +475/−0 | +533/−0 | — | 0 |
| tui (first) | `v0.9.0` | 04-26 → 04-26 | 913 | +2042/−572 | 0 | — | 0 |
| tui-qol | `v0.10.0` | 04-26 → 04-26 | — | +318/−30 | 0 | — | 0 |
| fix commit | `v0.10.1` | n/m → 04-27 | — | +24/−5 | 0 | — | n/m |
| expand-influence | `v0.11.0` | n/m → 04-27 | 166 | +1123/−13 | +9/−0 | — | n/m |
| use-asset-ability | `v0.12.0` | 04-27 → 04-27 | 428 | +1060/−30 | +457/−0 | — | 1 |
| goal-engine | `v0.13.0` | 04-29 → 04-30 | 439 | +1719/−137 | +22/−9 | — | 1 |
| narrative-renderer | `v1.0.0` | 05-01 → 05-01 | 814 | +1340/−4 | +1107/−0 | — | 1 |
| orchestrator | `v1.1.0` | 05-03 → 05-03 | 652 | +1299/−5173 | +464/−46 | — | 0 |
| testing-suite | `v1.1.1` | 05-04 → 05-05 | 178 | +451/−145 | +2208/−399 | — | no feat |
| event-hooks | `v1.2.0` | 05-06 → 05-07 | 849 | +1577/−152 | +1328/−284 | — | 1 |
| spatial-model hardening | `v1.2.1` | n/m → 05-10 | 802 (both rows) | +61/−35 | +336/−31 | — | n/m |
| spatial-model effort 2 | `v1.3.0` | n/m → 05-11 | (above) | +410/−1040 | +377/−99 | — | n/m |
| sub-engine self-bootstrap | `v1.3.1` | 05-13 → 05-14 | 1808 | +1078/−824 | +63/−2 | — | no feat |
| asset movement | `v1.4.0` | n/m → 05-22 | 2577 (+636 folded ability-engine) | +1784/−902 | +3500/−1065 | — | n/m |
| logging & errors | `v1.5.0` | n/m → 05-23 | 1285 | +657/−88 | +811/−65 | — | n/m |
| tui-foundation | `v1.6.0` | n/m → 05-24 | 2024 | +1001/−14 | +56/−0 | — | n/m |
| campaign manager | `v1.7.0` | n/m → 05-25 | 1497 | +849/−96 | +415/−75 | — | n/m |
| spatial map CLI | `v1.8.0` | n/m → 05-26 | 1599 | +784/−0 | +587/−0 | — | n/m |
| tui-manage | `v1.9.0` | 05-27 → 05-29 | 1814 | +1585/−39 | +213/−0 | — | 1 |
| tui-turn | `v1.10.0` | n/m → 06-09 | 7212 | +3754/−450 | +892/−198 | — | n/m |
| documentation overhaul | `v1.11.0` | n/m → 06-14 | 1603 | +7/−7 | 0 | 60 (13, 2) | n/m |
| A.001.1 effect-keyed dispatch | `v1.11.1` | 06-17 → 06-18 | 321 | +23/−29 | +77/−77 | 5 (1, 0) | no feat |
| A.001.2 A-flag registry | `v1.12.0` | 06-18 → 06-29 | 4598 | +1419/−266 | +1392/−31 | 13 (4, 0) | 1 |
| A.001.3 S-flag registry | `v1.13.0` | 10-11 → 10-17 | 9283 | +3299/−425 | +5707/−272 | 61 (4, 5) | 1 |

**Row notes:**

- `v0.8.0` and `v0.11.0` tags sit on version-bump commits. The `v0.11.0` diff is
  `7c1bded..6aec616`.
- `v1.2.1`: spatial-model Effort 1 commits (`2f7796e`–`9e9bc5e`) landed straight on
  `main` before the tag and are not in this diff.
- `v1.3.1`: doc lines include `sub-engine-alignment-*` and
  `sub-engine-self-bootstrap-plan.md`.
- `v1.4.0`: 2577 lines = `asset-movement-*` and `feature-movement-redesign-*` (pre-merge
  plan and review). The folded `ability-engine-redesign-*` docs add 636.
- `v1.5.0`: two commits on `main` (`cf379fc`, `4ebebc4`). Diff = `f09e4cc..4ebebc4`.
- `v1.10.0`: doc lines include `tui-turn-code-review-fixes-plan.md` (629).
- A.001.3 start date is the Effort 1 squash time. DR detail per plan:
  overview 28, Effort 1 13, Effort 3 15, Effort 5 5.
- Arc docs are not in any row: A.001 776 lines, tui-rebuild 383.

<br/>
<br/>

# 4. Rework and Catch Signals

## Review findings

| Review | Date | Record | Found | Fixed | Deferred | Rejected |
|--------|------|--------|-------|-------|----------|----------|
| Internal code review | 04-19 | `12da6ab` (13 files, Go +273/−64) | not recorded | — | — | — |
| Code review 2 | 04-23 | `c5699c6` merge | not recorded | — | — | — |
| Narrative renderer | 05-01 | `ee2e4fb` (6 files, Go +21/−20) | not recorded | — | — | — |
| Event hooks pre-merge | 05-07 | `a2e4bef` body | 6 | 6 | 0 | 0 |
| Spatial pre-merge | 05-10 | `25b6562` (8 files, Go +397/−66) | not recorded | — | — | — |
| Movement pre-merge | 05-22 | `feature-movement-redesign-pre-merge-review.md` | 9 (C1–C4, A1–A5) | not tallied | — | — |
| tui-manage pre-merge | 05-29 | `cd5ac48` (9 files, Go +70/−49) | not recorded | — | — | — |
| tui-turn code review | 06 | `tui-turn-code-review-fixes-plan.md` | not tallied | — | — | — |
| Documentation overhaul | 06 | `documentation-overhaul-review.md`; plan DR 5 | 13 (H1–H2, M1–M3, L1–L8) | 7 named in DR 5 | — | 1 left as-is (L8) |
| A.001.2 pre-merge | 06-28 | review doc not committed; plan; `915b861` (33 files, Go +596/−229) | at least 11 (C1–C11) | 2 named (C5, C10) | 1 (C11 → `F.024`) | not recorded |
| A.001.3 Effort 4, Commit 4 | 10 | `planned-work.md` rows | 2 | 0 | 2 (`F.026`, `R.021`) | 0 |
| A.001.3 Effort 4 end | 10 | plan DR 23 | 4 | 4 | 0 | 0 |
| A.001.3 pre-merge | 10-17 | plan DR 24–28 and the rejected list; `3a80182` (33 files, Go +330/−116) | 11 | 7 (DR 24–27, plus 3 cleanups in DR 28) | 0 | 4 |

## Found before merge vs after merge

- **Fix commits made straight on `main`** (`git log --first-parent --no-merges main`,
  subject matches `fix` or `bug`): 4, dated 2026-04-20 to 2026-05-12 (`f67cd0d` README
  formatting, `7c1bded`, `7152b6f`, `872c967`). None after 2026-05-12. Fixes inside
  squashed releases: not measurable.
- **Found after merge, by a later initiative:**
  - `RunCycle` registered hooks again on every cycle. Introduced in `f09e4cc` (`v1.4.0`,
    2026-05-22), found in the A.001.3 pre-merge review (2026-10-17, DR 24): 148 days.
  - `liveDefendersOnWorld` unused: finding A4 of the movement pre-merge review
    (2026-05-22). Filed again as `R.021` from the A.001.3 Effort 4 review. Still open.
  - Stale Scavengers tag ID and turn-order test: fixed in A.001.1 (`35e8c5e`,
    2026-06-17).
  - `B.009` (Capital Fleet data error): found in A.001.3 Discovery. Open.
  - `B.008` (stale `ApplyBookkeeping` comment): found by documentation-overhaul
    re-grounding (June). Open.
- **Found before merge, inside the same initiative:** `B.013` (found A.001.3 Effort 3,
  closed by DR 22 in Effort 4); `B.011` (found A.001.3 Effort 3, open, latent); `B.012`
  (made by A.001.3, filed at merge).

<br/>
<br/>

# 5. Doc Drift

## Seeds from A.001.3 — verified

1. **`planned-work.md` A.001.3 status.** The Status cell read "`A.001.2` (relocate)" at
   `v1.12.0` (2026-06-29) and at `3a80182^` (2026-10-17); `3a80182` replaced it. Stale
   for 110 days. (`git show <rev>:docs/dev_journals/faction-manager/planned-work.md |
   grep A.001.3`)
2. **`effect-mutation.md` "Asset features".** At the Effort 1 squash (`b4ada2a`) it read
   "Only the **transport** family is wired today" and "Every other `S`-flag feature … still
   waiting for a handler". Corrected in Effort 5 (Effort 5 plan DR 5).
3. **`world-movement.md` Key Decisions.** At `v1.12.0`, line 192 "The movement phase
   dispatches reactors twice" and line 205 "share the world engine but not the movement
   phase". Neither phrase is in the page now (Effort 5 plan DR 4).
4. **`hooks.md` Cat 4 list.** At `v1.12.0` the list and diagram (line 165) read "cost ·
   TL · movement". `ResolveAttackEligibility` is now named twice. Resolved at the Effort 4
   close-out (A.001.3 plan Reconciliation Items, DR 21).

## Further instances

- **A.001 Reconciliation Items:** the arc discovery listed 4 doc-vs-code mismatches; the
  arc plan routed 6; the A.001.3 plan table holds 6 (catalog Cat labels, catalog
  `relocate-other` category, two `hooks.md` claims, catalog latent vocabulary, `hooks.md`
  Cat 4 list).
- **Documentation overhaul:** review H2 — `goals.md` said goal-lock runs before
  bookkeeping, against the reorder. Effort 3 re-grounding found stale source comments
  (registry filename, `OverlayDoneMsg`, `TurnState` phase → `B.008`). Effort 4
  re-verify found 4 + 3 stale references.
- **Plan links:** `a-flag-ability-registry-plan.md` links a review doc that is not in
  the repo (C2).
- **Baseline (2026-05-22):** README and architecture overview described a CLI that did
  not exist (`CODEBASE_REVIEW.md` sections 3.1, 3.2). Section 6 covers what changed.

<br/>
<br/>

# 6. Baseline Comparison

`CODEBASE_REVIEW.md` section 10, against the source on 2026-10-17. Status words: *done*,
*partly*, *no change found*.

| # | Recommendation | Evidence since | Status |
|---|----------------|----------------|--------|
| 1 | Reconcile README and architecture doc with reality | README (last changed `9e3426f`, 2026-05-25) describes `gm-toolkit campaign create` and `gm-toolkit faction`; `cmd/gm-toolkit/faction.go:85` calls `tui.Run`. Architecture wiki rewritten in `857412e` (2026-06-14). Phase order in the `orchestrator.md` diagram matches `RunFactionTurn` (`orchestrator.go:66`): stat-raise → bookkeeping → save → movement → goal-lock → action | done |
| 2 | Build the minimum CLI before the next refactor | `cmd/gm-toolkit/` added and root `main.go` removed in `1d1fcf2` (2026-05-24); commands `faction`, `campaign`, `spatial`; TUI in `internal/faction/tui/` (`v1.6.0`, `v1.9.0`, `v1.10.0`) | done |
| 3 | Add CI | `.github/workflows/ci.yml` added in `f09e4cc` (2026-05-22): `go vet`, `go build`, `go test` | done |
| 4a | 5.2 Scavengers unbounded `credited` map | `ScavengersReactor` holds only `FactionID` (`tags/scavengers.go:14`); map removed in `f09e4cc` | done |
| 4b | 5.5 `TickMovementOrders` nil-deref | returns an error on an unknown definition (`world/movement.go:35`); `f09e4cc` | done |
| 4c | 5.3/5.9 phase state machine and stat-raise order | `TurnState` has no phase field (`domain/turn.go:5`); no `PhaseMovement` symbol (LSP workspace search); field removed in `f09e4cc`. Stat-raise order: code and `orchestrator.md` agree (row 1). `B.008` tracks a comment that still names the phase | done |
| 4d | 5.1 failure signal for mutations on missing entities | `MutationApplyError` with `Misses` (`mutation/mutation.go:23`), added in `f09e4cc`; LSP references in `mutation.go` and `orchestrator_apply_test.go` | done |
| 5 | Architecture-vs-reality check in the pre-merge checklist | Pre-merge step 2 promotes Decision Record entries into arch-page Key Decisions (`CLAUDE.md` item 9.2, `session-modes.md:146`; `857412e`). No step that checks pages against code. A.001 tracks page drift in Reconciliation Items tables | partly |
| 6 | Delete `HANDOFF.md` | No `HANDOFF.md` on disk. `git log --all -- HANDOFF.md docs/HANDOFF.md` is empty (never tracked) | done |
| 7 | Don't expand the process docs further | `session-modes.md` 321 lines (`v1.4.0`) → 437 (`v1.11.0`, unchanged since); templates 2 → 5 (`arc-concept`, `arch-page`, `docs-plan` added); `CLAUDE.md` 72 → 84 lines; arcs and the effort mechanism added | not covered by the status words: a "don't" recommendation. Facts only |
