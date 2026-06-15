# Handoff — Start the H.A.W.K Arc (build the Arc-Discovery)

> **Purpose of this doc:** F.022 (H.A.W.K Rulebook) grew past a single initiative
> during a pre-plan ability-catalog session and is now an **arc**. Rather than
> keep patching a per-initiative discovery doc into arc shape, we're restarting
> through the proper arc flow (`session-modes.md` section 9). This handoff hands
> the next session everything it needs to draft the **Arc-Concept** and run the
> **Arc-Discovery** session. Delete this doc once the arc-discovery exists.

## Situation

- A pre-plan **ability-catalog session** ran and produced a durable design doc.
  Its key output: the **`R.004` call** — build the data-driven ability/effect
  registry in full, across both dispatch surfaces — which is what tipped this
  from one initiative into an arc.
- The existing per-initiative discovery (`hawk-rulebook-discovery.md`) was then
  edited repeatedly to reflect arc scope. It's accurate but living in the wrong
  place and at the wrong altitude (it mixes cross-cutting arc decisions with
  per-initiative detail). **Treat it as input, not as the arc-discovery.**

## Inputs the next session inherits

**Artifacts (read these first):**
- `docs/initiatives/discovery/hawk-rulebook-discovery.md` — current discovery, arc-shaped. Source for decisions + candidate efforts. Do **not** rename it into the arc-discovery; mine it.
- `docs/architecture/engine/ability-catalog.md` — **done, durable.** The A/S ability decomposition into primitives and the `R.004` design. This is the binding mechanical design input; arc-discovery should not re-derive it.
- `docs/architecture/engine/overview.md` — page index already carries the catalog row (status: Design).

**Decisions already ratified (do not re-litigate in arc-discovery — carry them in):**
1. **`R.004` — build the full data-driven registry now, both surfaces.** Not bespoke-only, not deferred. `R.018` (effect-keyed A-dispatch) is the prerequisite increment.
2. **Surface split is load-bearing:** `A`-flag → abilities engine, `S`-flag → effects engine. The flag picks the engine; the effect picks the primitive (same effect can appear on both surfaces — e.g. relocate-other).
3. **Goals/tags neutralization is data-only** — bare-ID handlers unchanged; only the data loses SWN flavor.
4. **Shared goals/tags home — source-library dedupe:** a shared `rulebooks/_core/` composed at scaffold so every ruleset gets the same goals/tags. A `campaigns`/scaffold change, not a loader change.
5. **Lore canon is a living doc:** seeded by a **precursor lore session**, then grown through asset authoring. It is deliverable #4, at `docs/campaigns/hawk/`.
6. **Catalog location/axis:** durable, in `docs/architecture/engine/`, organized by primitive.

**The four arc deliverables (from Robert):**
1. All A- and S-flag specials delivered data-driven, with the surface split explicit.
2. A full H.A.W.K rulebook — assets fully defined and functional.
3. Goals + tags made campaign-agnostic (data-only) + the shared source home.
4. Durable docs: the ability/effect catalog (done) + the H.A.W.K lore canon.

## The arc process to follow (session-modes.md section 9)

Run these as **separate sessions** — one mode per session (section 10).

**Step 0 — Promotion + Arc-Concept** (lightweight, at promotion; Robert's call to promote)
- Robert promotes the arc to Current Initiatives (`A.` prefix) — manual.
- Draft the **Arc-Concept** from `templates/arc-concept.md` → `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-concept.md`. Required sections: Vision, Problem, Why an Arc, Candidate Initiatives, Trigger. A page is plenty; it makes **no** design decisions and slices **no** initiatives.

**Session A — Arc-Discovery** (model: **Fable**)
- Artifact: `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-discovery.md` (structurally like a discovery doc, but its decisions bind every initiative in the arc).
- **Goal:** ratify the cross-cutting architectural decisions — the registry shape (both engines), the surface contract, the `_core` shared-home mechanism, the data shape that makes the surface visible in TOML, the catalog as binding input. **Do not slice into initiatives** (that's Arc-Plan).
- Re-ground first (section 10): read `architecture-overview.md`, then `persistence.md`, `actions.md`, `effect-mutation.md`, and the new `ability-catalog.md`; verify load-bearing claims against current source (`dispatch.go`, `core.go`, `effect/`, `rulebook.go`).
- Ends when the arc-discovery is written and Robert signs off. **Do not start Arc-Plan in the same session.**

**Session B — Arc-Plan** (model: **Fable**)
- Artifact: `docs/initiatives/arcs/<arc-name>/<arc-name>-arc-plan.md`.
- Slice the arc into ordered initiatives (a DAG), set per-initiative boundaries, route each Arc-Discovery open question to a specific initiative, list cross-arc/engine prereqs and out-of-arc deferrals. Add each initiative to `planned-work.md` referencing the arc-plan.

After that, each initiative runs the standard (lighter) Discovery → Plan → Execution lifecycle.

## Guardrails

- **Model:** Fable for Arc-Discovery and Arc-Plan (section 10 model table). Prompt Robert to `/model` before each.
- **One mode per session.** Concept, Arc-Discovery, Arc-Plan are three separate sessions. No code, no execution — these produce docs only.
- **Arc artifacts live in `docs/initiatives/arcs/<arc-name>/`** and stay there permanently (nothing moves to `completed/`). Per-initiative discoveries/plans use the standard locations.
- **Altitude discipline:** the arc-discovery holds only what binds *every* initiative. Per-initiative detail in the current discovery (the asset-authoring pass loop, lore-gate mechanics, SWN re-prefix migration steps) is **seed for per-initiative discoveries**, not arc-discovery content.

## Recommended arc name & candidate initiatives (seed for the Concept)

**Suggested arc name:** `data-driven-rulebook` — the cross-cutting theme is making the engine truly multi-rulebook by going data-driven, with H.A.W.K as the first real campaign that proves it. (Alternative: `hawk-rulebook`, if you'd rather name it after the marquee deliverable. Robert's call at promotion.)

**Rough candidate initiatives** (Arc-Plan does the real slicing — these are non-binding):
- **R.018** — effect-keyed A-dispatch (replaces the `map[def.ID]`). Prerequisite increment.
- **A-flag ability registry** — data-driven active abilities (target-select / stat-contest / die-roll → effects).
- **S-flag effect registry** — data-driven passive/reactive effects mapped onto the five hook categories.
- **Goals/tags neutralization + shared `_core` home** — data-only neutralization + the scaffold-compose change.
- **H.A.W.K asset authoring** — the rulebook content, with the precursor lore session + per-pass lore-gate feeding the canon.

A real Arc-Plan question: which of these are independent shippable initiatives vs. efforts within one. Robert chose "arc," so default to initiatives, but flag any that are really efforts.

## Open questions to carry / route (from the current discovery)

- **Initiative slicing & ordering** → Arc-Plan.
- **SWN re-prefix migration reach** → route to whichever initiative owns the SWN data (likely R.018 or asset-side); whether any scaffolded campaign needs migrating.
- **Monopoly flag-vs-text mismatch** → the A-registry initiative.
- **Precursor lore session** still to run → likely attaches to the H.A.W.K asset-authoring initiative.
- **Custom session types** (lore + ability-catalog) and the `goals.md` "intentional no-op" wording → deferred, process-level (Fable).

## Housekeeping

Current working tree has **uncommitted** changes from the catalog session:
`ability-catalog.md` (new), the `overview.md` index row, the discovery rewrite, and this handoff. Recommend committing them as a checkpoint (`docs:`) before the clean restart so the arc sessions begin from a committed baseline — Robert's call.
