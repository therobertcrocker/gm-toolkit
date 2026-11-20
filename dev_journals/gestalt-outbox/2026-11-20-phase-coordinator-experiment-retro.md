# Phase Coordinator Experiment — Retro

> **What this is:** Claude's own view of one experiment, written by the session that ran it, as raw material for a possible gestalt skill
> **Run:** 2026-11-19 to 2026-11-20, gm-toolkit, branch `feature/pf2e-encounter-simulator`, effort `F.041.E4` (six phases, 13 planned commits)
> **Not yet known:** the result of `/gestalt:effort-close F.041.E4`. It is the real score for this experiment; see section 5

## In Plain Words

Normally each phase of a plan is one Execution session that Robert starts and reads himself. In this experiment one long-lived session, the coordinator, ran on Fable and did no phase work. For each of the six phases it wrote a brief, started one Opus sub-agent to execute that phase and commit it, checked the result on its own, and brought Robert a review with the open calls. Robert decided, the same agent made one fix commit, Robert signed off, and the next phase started.

The aim was that easy-to-catch problems get caught phase by phase, so the effort close has less to clean up.

All six phases landed: 13 planned commits, 6 fix commits (one per phase) and 1 chore commit, every agent reply on Opus, the Decision Record at 72 entries. In every phase the code the agent wrote matched the plan. Every real finding was a place where the plan itself was wrong, most often a game rule the plan had written from memory.

## 1 — Did we do what we set out to do?

Robert's guidelines, and how each held:

| Guideline | Result |
|---|---|
| All sub-agents on Opus | Held. Every spawn set `model: opus`, and the coordinator confirmed it after each run by counting the `"model"` values in the agent's transcript file |
| One agent handles one phase, then reports | Held. Six execution agents. A seventh, also Opus, did one research task (quoting rules text) |
| Coordinator and Robert review; Robert signs off | Held for all six phases |
| Each next agent is set up better than the last | Held, with evidence: the briefs grew from about 8,000 characters (Phase 1) to 11,000, 14,000, 14,000, 19,000 and 14,000, and each addition traces to a problem in the phase before |
| Less cleanup at the effort close | Not measured yet |

What was added that was not planned:

- **The agent commits its own phase.** Robert's call at the start, so two agents never share an uncommitted tree.
- **The same agent makes the fix.** It is resumed with its context intact, so the fix brief can be short (about 4,000 to 9,000 characters) and the agent already knows the code.
- **A rules text check.** From Phase 2 on, every brief listed the game rules the phase would build and told the agent to fetch the published text, compare it with the planned code, and report each difference with a quote and a link. The agent never decided a difference; Robert did.
- **A two-stage phase.** Phase 6 needed Robert to run a command on his real campaign data between two commits. The agent built the first commit, stopped without committing, and was resumed after Robert ran it.

## 2 — Is there a better way we didn't see?

**The coordinator's review could run before the agent commits.** As run, the agent committed, then the review found problems, then a fix commit followed: six fix commits for six phases. A review before the commit would fold most fixes into the planned commits. It would cost a second stop per phase and lose the clean record of what the review changed. Robert has said he wants review fixes visible as their own commits, so the chosen shape still holds. It is the first thing to revisit if the fix commits become noise.

**The rules text check belongs in the Plan session, not in Execution.** Nearly every finding was the plan's own error. The check found them one phase at a time, each costing a review round and a fix commit. Run once while the plan is written, it would have fixed the plan before any code existed. This is the largest saving on offer, and it does not need a coordinator at all.

**The coordinator does not need to be the stronger model for everything it did.** Its work was of three kinds: reading ahead in the plan against the current source to write seam notes, checking the agent's result, and framing calls for Robert. The first and third gained from holding all six phases in one head. The second (model count, gate, diff read) is mechanical and could be a script or a hook.

**What still holds:** one phase per agent, the same agent for the fix, and Robert making every rules call. None of these showed a cost worth trading.

## 3 — Does it fit the process?

Where it matches the gestalt rules:

- **One phase is one session.** Each agent was a fresh context with one phase of scope, which is the Execution rule in another form.
- **Planned commits.** Agents committed with the plan's messages, the `E4` scope, `git add -A` and `git commit -F`.
- **Decision Record, append and reconcile.** Agents logged deviations and fixed plan lines that the source proved wrong.
- **Questions.** Findings went to Robert as text first, then as a question block with the recommended option first.
- **Fable for structure, Opus for the work.** The coordinator's job (set up many later sessions) fits the model rule's intent, though the rule does not name it.

Where it diverges or strains:

- **The coordinator session is very long.** It ran out of context once and was compacted. After that it worked from a summary, and its own notes were less exact.
- **Two writers on one plan doc.** The coordinator added Decision Record entries while agents also edited the plan. The fix was a rule: the coordinator does not touch the plan while an agent runs.
- **Ad-hoc work in the same session.** While agents ran, Robert and the coordinator held a design discussion on the repo's folder layout. It worked, but anything the coordinator writes to disk would ride into the agent's `git add -A`, so writing waited for sign-off.
- **An unscoped commit inside the effort's range.** The coordinator's own commit for that discussion sits after the `E4` commits and before the close.

## 4 — What does it change?

**What made each next agent do better.** These are the parts of the brief, in the order they were added, each with the problem that caused it.

| Added in | Brief element | Caused by |
|---|---|---|
| Phase 1 | The plan and read list; "how to treat the plan" (copy the literal code, but verify each Find block against the source; a mismatch is a deviation to record); the coordinator's notes from its own walk of the phase against the source; project rules; a fixed, numbered report that opens with "needs a call" and ends with "doubts and things for the next phase" | The start, and the plan traps catalogue |
| Phase 2 | "If plan mode switches on by accident, wait and carry on" | Robert cycled modes by accident during Phase 1 and it reached the sub-agent |
| Phase 2 | The rules text check: a list of the rules this phase builds, where to fetch the text, compare, quote, never decide; and a report line per rule | Robert asked "where are we getting the rules from?" The answer was: from Claude's memory |
| Phase 2 | "What the earlier phases left you": the commits so far, and every place a review fix made the source differ from what the plan's later code expects | The Phase 1 fix changed code that later Find blocks expected in the old shape |
| Phase 2 | A report section where the agent answers the coordinator's notes one by one | A note is only worth writing if someone checks it |
| Phase 3 | Prove a test's teeth with Edit and Edit back; never `git stash` or `git checkout` | The Phase 2 agent used `git stash` to show a test could fail |
| Phase 3 | Longer coordinator notes, numbered, each naming a call into older code or a Find block that will not match | Review fixes kept moving the source away from the plan |
| Phase 5 | An import format check in the report: every field and mode of each outside data shape, with anything not handled reported | Plan trap 4, and a number the coordinator itself got wrong by reading one field |
| Phase 6 | Two stages; which data the agent may run against and which copies are frozen | The phase needed Robert to act between two commits, on his real data |
| Phase 6 | A report section "what an effort-close reviewer should look at first" | The last agent has read all the source for the docs; it found one real bug this way |

**The pattern of findings.** Across six phases, the agent's code matched the plan every time. The findings came from three sources only:

- **Fetched rules text against planned code.** The largest source. Examples: stunned gained on your own turn, flourish counted per round, fast healing on a downed creature, Haste on the caster.
- **The agent's own "needs a call" flags**, where the plan was silent or the source disagreed.
- **The agent's "look here first" list** at the end, which found a latent bug no test covered.

No finding came from the coordinator reading a diff and spotting a coding error the agent had missed. The coordinator's diff reads confirmed; they did not discover.

**A new plan trap.** Rules written into a plan from the model's memory. The plan stated game rules with confidence and no source, and several were wrong in the cases the plan had not thought about. Check: every rule a plan builds carries a quote and a link, fetched in the Plan session. This belongs in the plan traps catalogue as trap 8.

**The coordinator's own mistakes.** Each one cost Robert a correction.

- It called a rule a house rule and stated rules from memory. Robert challenged it; the fetch showed other real errors.
- It derived one of Robert's numbers differently, and told him his number was the slip. His number was right; the coordinator had read one field and missed another.
- It refreshed the importer's test copies of exported files along with the real ones, which broke four tests. The test copies are a frozen snapshot.
- It twice ended a turn with findings and no question block, and Robert had to ask for the questions.
- It reported 13 labels in a doc as stale and missed a legend line in the same doc that explained them. The fix left that legend orphaned.
- One brief told the agent to use one commit trailer and also to match the last commit's trailer; they differed. The agent picked one and said so.

The common cause of the first, second and fifth: the coordinator stated a conclusion about a source it had read only in part. The rule it gives agents ("fetch the text, quote it, do not decide") applies to the coordinator too.

## 5 — Rating

The Rework row cannot be scored until the effort close runs. That close is the test of the experiment's aim.

| Metric | Score | Evidence |
|---|---|---|
| Guidelines held | 5 | All agents on Opus and verified; one phase per agent; six sign-offs |
| Each agent better set up | 4 | Every brief addition traces to a real problem; but the same kind of finding (a rule from memory) came back in every phase, because the cause was in the plan |
| Findings caught early | 4 | One fix commit per phase, each small; the last agent's own list found a latent bug |
| Robert's load | 3 | He made every call quickly from a question block, but he also corrected the coordinator about six times |
| Rework at the close | not scored | Fill in after `/gestalt:effort-close F.041.E4` |
| **Total** | **16 / 20 (mean 4.0)** | Four metrics scored |

**Verdict:** the loop works and is worth a skill, but its biggest finding is that the checking it did belongs one stage earlier, in the Plan session.

**What surprised me:** the coordinator never caught a coding error. The agents copied the plan faithfully, so the quality of the run was the quality of the plan plus the quality of the questions put to Robert.

## Follow-ups

Proposals for gestalt. Robert makes any change to the plugin.

| # | Proposal | Detail |
|---|---|---|
| 1 | A source check in the Plan session | For every outside rule a plan builds (game rules, a file format, an API), the plan carries a quote and a link fetched in that session. Add it as plan trap 8 |
| 2 | A coordinator skill | One command that runs an effort's phases as sub-agents with the loop in this doc. The brief template in the appendix is its core |
| 3 | A fixed execution report | The numbered report format, with "needs a call", "doubts and things for the next phase" and "shapes later Find blocks will not match", is useful in a normal Execution session too |
| 4 | "Look here first" at the end of an effort | The last phase's agent writes its own list of the weakest spots; the effort close reads it |
| 5 | Mechanical checks as a script | Model count from the transcript, a fresh-cache gate, a clean tree and the commit list do not need a model |
| 6 | Rules for the coordinator | It does not edit a doc an agent is editing; it writes nothing to disk while an agent with `git add -A` runs; it quotes a source before it states a conclusion from it |
| 7 | A two-stage phase | A phase that needs the user to act in the middle: the agent stops without committing and is resumed after |
| 8 | Score the experiment | After the effort close, add the Rework score here: how many findings the close made, and how many of them a phase review could have caught |

## Appendix A — The loop

```
for each phase:
  coordinator reads the phase against the current source
  coordinator writes the brief
  one Opus agent executes the phase and commits the planned commits
  agent reports in the fixed format
  coordinator verifies on its own:
      model count, gate with a fresh test cache, clean tree,
      commit messages, the diff, edits to older tests
  coordinator gives the user the review as text, with a recommendation per call
  user answers (a question block, recommended option first)
  the same agent, resumed, makes one fix commit
  coordinator verifies the fix commit
  user signs off
```

## Appendix B — The brief template, as it stood at Phase 5

1. **The task in one line**, with the plan path and the phase.
2. **Read list.** The plan sections, the architecture pages, the rules files, the skills to load.
3. **What the earlier phases left.** The commits so far, and each place a review fix made the source differ from what the plan's later code blocks expect.
4. **How to treat the plan.** The plan's code is literal; copy it, but verify each Find block against the source first. A mismatch is a deviation: follow the source, record it. A rules reading is never the agent's to decide; it goes under "needs a call".
5. **Source check.** The list of outside rules this phase builds, where to fetch each text, and the instruction to compare, quote, link and report, not decide.
6. **Coordinator's seam notes.** What the coordinator found when it read this phase against the source: calls into older code and their side effects, fields a new rule reads and who fills them.
7. **Project rules.** The ones agents break: file edits with Edit and Write only; the project's code search rule; comment rules; no comments in tests; prove a test's teeth with Edit and Edit back, never `git stash` or `git checkout`; which data the agent must not run the binary against.
8. **Commits.** The plan's messages, the effort scope, `git add -A`, `git commit -F` from a file outside the repo, no push, no amend. The gate command in full.
9. **Report format**, fixed and numbered, in this order: needs a call; commits; the gate output's tail; source check, one line per rule; deviations from the plan, with Find blocks that did not match; existing tests changed, each with its reason; tests written, one line per planned case; import format check; answers to the coordinator's notes; Decision Record entries added; doubts and things for the next phase.
10. **If plan mode switches on by accident,** wait and carry on.

## Appendix C — Where the raw material is

The coordinator's transcript holds every brief and every fix brief in full:

`~/.claude/projects/-workspaces-gm-toolkit/93b19d79-33e4-4917-916d-3ab83038405f.jsonl`

To print the six phase briefs and the research brief:

```
jq -r 'select(.type=="assistant") | .message.content[]?
  | select(.type=="tool_use" and .name=="Agent")
  | "==== \(.input.description) ====\n\(.input.prompt)\n"' <transcript>
```

For the fix briefs, use `.name=="SendMessage"` with `.input.summary` and `.input.message`.

The agents' own decisions are in the plan's Decision Record: `dev_journals/campaigns/initiatives/implementation/F.041.E4-pf2e-encounter-simulator-plan.md`, entries 1 to 72.
