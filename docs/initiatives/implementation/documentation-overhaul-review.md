# Documentation Overhaul (D-002) — Full-Corpus Quality Review

> **Branch:** `docs/documentation-overhaul` · **Date:** 2026-06-13 · **Reviewer:** Claude (Opus)
> **Status:** Findings for triage — no edits applied. Pre-merge quality gate (CLAUDE.md §8).
> **Scope:** the ~30-doc D-002 corpus — `docs/architecture/` (overview + engine + interface),
> `docs/contributing/`, and `docs/process/` (diagram-kit, session-modes, style-guide, templates).

## For the implementing session

Context a fresh session needs that the findings below don't carry on their own:

- **Branch state.** The CLAUDE.md §8 pre-merge checklist is done through step 6 and sits
  **uncommitted** in the working tree (diagram-kit reconciliation, the orchestrator.md goal-lock Key
  Decision, planned-work/completed-work closeout, three plan docs `git mv`'d to
  `implementation/completed/`). There is also an unrelated `internal/faction/engine/orchestrator.go`
  phase-reorder (goal-lock now runs late) — **code, not docs.** Don't commit or merge until the fixes
  are folded in.
- **Commit shape.** Doc fixes fold into **one `docs:` commit**; the `orchestrator.go` reorder lands as
  **its own conventional commit** (Robert's call). Merge strategy (squash vs. keep-orchestrator-visible)
  and a version bump from `v1.10.0` are still open — step 7, after the fixes.
- **M2 is a decision gate, not a mechanical edit.** Every other finding is a mechanical correction
  (a path, a word, a count) safe to apply directly. **M2 (the disabled Spatial mode's purpose) needs
  Robert's decision first** — don't guess which of the three conflicting descriptions is canonical;
  resolve it with him, then align overview.md + `tui/view.go:53` + state-machine.md (and L5's ID
  format) together.
- **Model.** Mechanical doc corrections are Sonnet-appropriate (the mechanical-vs-generative split,
  not code-vs-docs). M2's resolution is a design call for Robert, not the model.
- **Decision Record.** None of these are logged yet. When fixes land, log execution decisions in the
  plan's Decision Record (the D-002 plan now in `completed/`); H2 and M2 may be durable enough to
  graduate to the relevant architecture page's Key Decisions at pre-merge — Robert's call.

## Method

Read the full corpus as a connected set, walking from both front doors
(`architecture/architecture-overview.md`, `contributing/overview.md`) the way a real reader
would, so routing/navigation problems surface and not just per-page nits. Lens: three axes
(clarity, coherence, organization) × two readers (human, LLM). Every load-bearing
numeric/structural claim was spot-checked against actual source; every relative link and
in-page anchor was resolved mechanically.

**Headline:** the corpus is in excellent shape. The architecture pages are unusually faithful
to source — every number sampled (12 actions, 20 tags, 16 collector methods, 12 goals, the
income formula, `eventChanBuffer=64`, `headerHeight=5`, `wizardBandHeight=14`, `keepCount=50`,
the `Apply` panic-default) matches the code. Recipe↔arch wiring is bidirectional and all anchors
resolve. The `block-beta`→`block` reconciliation is clean with no remnants. Findings concentrate
in two failure modes: **dead links from a wrong rules-doc path**, and **prose arithmetic / a
stale ordering claim the goal-lock reorder left behind**.

---

## HIGH — broken links & a source-contradicting claim

### H1. Three dead cross-links to the rules doc

The rules doc lives at `docs/rules/swn-faction-mechanics.md`, but three pages link to
`../../swn-faction-mechanics.md`, which resolves to `docs/swn-faction-mechanics.md` (does not exist):

- `architecture/engine/actions.md:97`
- `architecture/engine/effect-mutation.md:134`
- `architecture/engine/goals.md:95`

**Fix:** change each to `../../rules/swn-faction-mechanics.md`.

**Root cause / propagation:** the arch-page template seeds the wrong path in prose —
`process/templates/arch-page.md:55` ("a pointer to `docs/swn-faction-mechanics.md`"). Fix the
template too or it recurs on the next data-backed page. **Related (out of scope):** `CLAUDE.md:38`
carries the same wrong path — worth fixing in the same pass so the project's own routing table
isn't broken.

### H2. `goals.md` still says goal-lock runs before bookkeeping — contradicts the reorder

`architecture/engine/goals.md:58-59`: "`CheckLock` — start of turn… Called in the goal-lock phase
*before* bookkeeping and action selection." Per the orchestrator.go reorder (now correctly reflected
in `orchestrator.md` and `engine/overview.md`), goal-lock runs **after** bookkeeping and movement,
immediately before action. The "before bookkeeping" half is now false and disagrees with source and
two other pages.

**Fix:** reword to "before action selection" (drop "bookkeeping"), and reconcile the same framing
where it recurs lower on the page (the line 91 catalogue intro and the Key Decisions "two
advancement paths" entry both lean on the "start of turn" framing — verify each reads correctly
against late goal-lock).

---

## MEDIUM — coherence / source-fidelity

### M1. `effect-mutation.md:48` miscounts the hook categories

"the four implemented tags… exercise **four** of the five hook categories." The table directly below
covers Cat 1 (Warlike), Cat 2 + Cat 5 (Fanatical), Cat 3 (Scavengers), Cat 4 (Preceptor) — **all
five**. The page even says "Fanatical already does two." The error undercuts the very point it's
making (the proof is *stronger* than claimed).

**Fix:** "exercise **all five** hook categories."

### M2. The disabled "Spatial" mode's purpose conflicts across pages and source

`interface/overview.md:30-33` says the slot's "eventual role isn't settled; the current direction is
a faction-utilities / edit surface… 'Edit Mode'." But the source placeholder (`tui/view.go:53`) and
`interface/state-machine.md:151` say "reserved for **F-012 (Spatial Map CLI)**" — and `F.012`
(spatial-map-cli) has **already shipped** as a standalone CLI tool (`planned-work.md:9`). So three
sources disagree, and one points the slot at an already-shipped, separately-delivered feature. This
is a genuine conflict to resolve, not a wording nit.

**Fix (needs Robert's call):** decide the slot's intended purpose, then align overview.md, the source
placeholder string, and state-machine.md. (Also fixes L5: the placeholder uses hyphenated `F-012`
where the convention + planned-work use dotted `F.012`.)

### M3. `state-machine.md:184` says the root "holds an `*engine.Engine`" — it doesn't

The documented `Model` struct on the same page (and `tui/model.go`) has no engine field; `eng` is a
`NewModel` parameter forwarded to `turn.New` and never stored. Minor arch-docs-reflect-source
violation on a page that's otherwise precise.

**Fix:** "It receives an `*engine.Engine` at construction and forwards it to the turn view (never
storing or calling it); it holds `*state.FactionState` for the views to project."

---

## LOW — counts, nits, cosmetics

| # | Location | Issue | Fix |
|---|----------|-------|-----|
| L1 | `engine/actions.md:9` | "Attack, Buy Asset, Expand Influence, Use Asset Ability, **and nine others**" = 13, but the page says "twelve" twice and source has 12 | "and **eight** others" |
| L2 | `engine/orchestrator.md:20` | "the rulebook plus the **eight** collaborators" — the struct shown has 9 non-`log` fields beyond `Rulebook` (Rand, Hooks, + 7 sub-engines) | "**nine** collaborators" |
| L3 | `process/diagram-kit.md:5` | "the **twelve** architecture diagrams read as one set" — actual mermaid count is 11 | "**eleven**" (or recount if a diagram is planned) |
| L4 | `process/session-modes.md:435` | "**Both** templates carry inline guidance" — section 11 lists **five** templates | "**All** templates" |
| L5 | `state-machine.md:151` + `tui/view.go:53` (source) | hyphenated `F-012` vs the dotted convention (`F.012` in planned-work + session-modes §2) | normalize to `F.012` (source string change; doc quotes it faithfully) — bundle with M2 |
| L6 | `engine/logging-errors.md:84-95` | mermaid frontmatter is indented while every other diagram's is flush-left — inconsistent with the kit, minor render risk in strict parsers | de-indent to match the kit's header style |
| L7 | `interface/event-stream.md:82` | sequence diagram shows `Eng--)UI: StreamClosedMsg`, but the prose (correctly) says it's generated UI-side by `ObserverPump` reading the closed channel | reattribute as a UI-side `Note`/self-message, or accept as sequence-diagram simplification |
| L8 | `templates/docs-plan.md:19`, `session-modes.md:256` | bare reference to `documentation-overhaul-plan.md`, which this branch `git mv`'d to `implementation/completed/` | leave (bare filename, not a link) or add `completed/` for precision |

---

## Verified clean (no action)

- **Terminology:** "Coin" used throughout; no "FacCreds" anywhere.
- **Cross-link integrity:** every relative link resolves except H1; all in-page anchors
  (`#key-decisions`, `#mutation-catalogue`, `#the-update-discipline`, recipe anchors) resolve;
  recipe↔recipe wiring (engine "Add an action" ↔ interface "Wire a prompt") is bidirectional.
- **Recipe file paths:** all 20+ cited paths exist; the lone "miss" (`engine/action/errors.go`) is a
  package reference (`engine/action/ (errors)`), and the sentinels are confirmed in `action.go`.
- **orchestrator.md:** correctly reflects the late goal-lock reorder; the new "Goal-lock sits
  adjacent to the phase it gates" Key Decision matches source.
- **Numbers:** 12 actions, 20 tags (4 implemented), 16 collector methods, 12 goals, income
  `Wealth/2 + (Force+Cunning)/4`, `Apply` panic-default, all magic numbers — all match source.
- **diagram-kit.md:** reads cleanly post-edit; no `block-beta` remnants in the corpus.
- **session-modes / templates:** all section cross-references (§2, §7, §8, §9, §10) resolve to the
  correct headings.

### Source-side aside (not a doc issue)

The `RunFactionTurn` *code* doc-comment at `orchestrator.go:55` still says "drives one faction's turn
from **goal-lock check** through state save" — stale after the reorder, since the turn now opens with
setup/stat-raise/bookkeeping. Code comment; Robert's call whether it rides the orchestrator.go commit.

---

## Suggested triage order

- **H1 + H2** — unambiguous; should land regardless.
- **M1 / M3 / L1–L4** — trivial one-word / one-line corrections.
- **M2 (+L5)** — the only finding needing a real decision: what is the third TUI mode actually *for*?
  Resolve that, then align the three sources.

Per CLAUDE.md §8 sequencing, the doc fixes fold into the same docs commit; the `orchestrator.go`
reorder lands as its own conventional commit. Still at pre-merge step 7 (commit → merge strategy →
version bump from v1.10.0) after this gate.
