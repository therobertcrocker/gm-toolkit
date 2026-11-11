# feature/movement-redesign — Pre-Merge Resolutions Plan

## Context / Goal

Resolves the open coherence questions and engine-bug findings surfaced in `CODEBASE_REVIEW.md` §5 (Opus 4.7, 2026-05-22) for the `feature/movement-redesign` branch. All work in this plan lands as **one bundled pre-merge commit** alongside the prior session's quick-win cleanup (CI gate, HANDOFF.md delete, gofmt pass).

- Review: [`/CODEBASE_REVIEW.md`](../../../CODEBASE_REVIEW.md) §5
- Companion review (tracked separately): [`feature-movement-redesign-pre-merge-review.md`](./feature-movement-redesign-pre-merge-review.md) — C1–C4 / A1–A5 are out of scope here

<br/>

## Decisions Ratified in Planning

1. **§5.3 — `TurnPhase` is a one-bit cursor** — Delete `PhaseMovement` and `PhaseComplete` from `domain.TurnPhase`. The phase field has exactly one consumer (the bookkeeping idempotency guard); `Checkpoint*` constants are the real pause/resume seam. Full state-machine tracking deferred to F-004 when a real driver exists. YAGNI applies.
2. **§5.9 — Code is source of truth: StatRaise runs before Bookkeeping** — Income calculation uses post-raise stat values. Decision #183 (StatRaise *after* `CheckpointBookkeeping`) is superseded. `architecture-overview.md` Phase 2 / Phase 2B sections need to swap to match.
3. **§5.4 — `dispatch.MutationReactors` enforces all-or-nothing rollback at the type signature** — Returns `(nil, err)` on cap-trip. Docstring updated. Honors Decision #138's "fail loud — surfaces content bugs" intent by construction rather than by caller discipline.

<br/>

## Shared Context

### Re-verification requirement (engine bugs only)

The five engine-bug items in §5.1, §5.2, §5.5, §5.6, §5.7 were identified against a branch snapshot ~4 commits behind current HEAD. **Each must be re-verified against current source before fixing.** Per `feedback_refactor_reground` and `feedback_verify_before_suggesting`: read the named symbol/file, confirm the bug still exists, then fix. If a bug has already been fixed by intervening work, skip it with a one-line note.

### Bundled commit

Everything in this plan lands as **one commit** on `feature/movement-redesign`. Do not split into per-section commits — the working tree stays dirty across all steps and the bundle is committed at the end. This is an explicit exception to the "one commit per execution session" rule for this pre-merge pass.

### Decisions log entries

Per CLAUDE.md rule #7, decisions-log entries are written before the pre-merge commit. This plan needs:

- One new decision for each coherence resolution (§5.3, §5.4) added to the `feature/movement-redesign — engine (Effort 2)` section
- An amendment or superseding entry for Decision #183 covering the §5.9 ordering
- New decisions for any engine-bug fix that changes behaviour rather than just adding a defensive lookup — use judgement; trivial guard adds don't warrant an entry

<br/>

## Out of Scope

- C1–C4 / A1–A5 from `feature-movement-redesign-pre-merge-review.md` — tracked in that review
- §5.8 stale maintenance-cost comment — resolved in the prior session, already in working tree
- CI workflow, HANDOFF.md deletion, gofmt normalization — resolved in the prior session
- F-004 CLI Rebuild prerequisites (proper resume state-machine wiring, structured logging from F-003) — deferred to the driver initiative

<br/><br/>

# Work Breakdown

Order is the order to execute. Coherence resolutions first (well-specified, no verification cost), then engine bugs (each starts with verification). Steps 4–8 are flagged `(NEEDS CONFIRMATION)` — re-verify before changing code.

## Step 1 — §5.3: Delete dead `Phase` enum values

- **Files:** `internal/faction/domain/turn.go` — remove `PhaseMovement` and `PhaseComplete` constants
- **Verify:** LSP `findReferences` confirms both are still single-ref (definition only) at execution time
- **Tests:** none expected; if any test references either constant, remove the test or rewrite
- **Decisions log:** new entry citing the one-bit-cursor rationale and the F-004 deferral

<br/>

## Step 2 — §5.4: `dispatch.MutationReactors` returns `(nil, err)` on cap-trip

- **File:** `internal/faction/engine/hooks/dispatch/mutations.go` — change `reactDispatch` line 38 from `return combined, fmt.Errorf(...)` to `return nil, fmt.Errorf(...)`
- **Docstring:** update `MutationReactors` doc comment to state the all-or-nothing contract explicitly; reference Decision #138 in inline rationale
- **Tests:** add a `mutations_test.go` case asserting `(nil, non-nil err)` on cap-trip — pick a synthetic reactor that fires infinitely
- **Decisions log:** new entry; reference Decision #138 as the ratifying authority

<br/>

## Step 3 — §5.9: Align docs to code order

- **Files:**
  - `docs/dev_journals/faction-manager/decisions-log.md` — amend Decision #183 in place, or add a superseding entry in the movement-redesign section that links back to #183
  - `docs/architecture-overview.md` — swap the order of Phase 2 (Bookkeeping) and Phase 2B (Stat Raise) so the doc states StatRaise → Bookkeeping; rename "2B" if the new ordering makes the suffix awkward
- **Code:** none — the code is already correct
- **Decisions log:** the §5.9 amendment is itself the decisions-log change

<br/>

## Step 4 — §5.1 (NEEDS CONFIRMATION): MutationEngine silent no-op on missing entities

- **Verify:** read `internal/faction/engine/mutation/` Apply implementation; confirm every case still uses `if faction, ok := ...; ok { ... }` with no log or error path. Check git log for any recent commit that may have added a fail-loud signal
- **Fix (if confirmed):** smallest viable change — emit a structured error on missing-entity. Two design options, **surface to Robert** if confirmed:
  - **(a)** `*MutationApplyError` accumulator: collect all misses, return at end of `Apply` — orchestrator decides whether to fail the turn
  - **(b)** Fail-fast: first miss returns an error immediately
- **Decisions log:** new entry — codifies the failure semantic chosen

<br/>

## Step 5 — §5.2 (NEEDS CONFIRMATION): `ScavengersReactor.credited` map grows unbounded

- **Verify:** read `internal/faction/engine/tag/tags/scavengers.go`; confirm `credited` is still a struct field lazily initialised and never cleared. **Critically:** check whether Decision #235's trigger-batch semantics now make the recursion-guard redundant in the first place
- **Fix options (if confirmed):**
  - **(a)** Clear `credited` at start of each `OnMutations` call (preserve the recursion guard, scope it to one dispatch)
  - **(b)** Remove the map entirely if trigger-batch already prevents the re-fire it was guarding against — the cleaner fix if §235 made it dead code
- **Decisions log:** entry if option (b); option (a) is small enough to skip the log

<br/>

## Step 6 — §5.5 (NEEDS CONFIRMATION): `TickMovementOrders` nil-deref on missing definition

- **Verify:** read `internal/faction/engine/world/movement.go` `TickMovementOrders` (line 33–34 per review); confirm `def := rulebook.Assets[asset.DefinitionID]` still lacks a nil check
- **Fix (if confirmed):** mirror Attack's defensive pattern — `def, ok := rulebook.Assets[asset.DefinitionID]; if !ok { return error or skip }`. Alternative: hoist to a load-time hard error (rulebook validation rejects any state file referencing an unknown DefinitionID) — **surface to Robert** if leaning this way, it's a broader change
- **Decisions log:** trivial defensive add doesn't need an entry; load-time hardening does

<br/>

## Step 7 — §5.6 (NEEDS CONFIRMATION): `MovementOrderProgressed` field redundancy

- **Verify:** read `internal/faction/domain/mutation.go`; confirm `MovementOrderProgressed` still carries both `NewHexCoords HexCoord` and `Region string` instead of one `RegionHex`
- **Fix (if confirmed):** collapse to a single `RegionHex` field. Update writers (`engine/world/movement.go`, `engine/effect/effects/transport.go`) and readers (`mutation.Apply`, `narrative/digest/build.go`). Also re-check JSON tags here, since C1 in the companion review noted the new mutations are missing tags entirely — coordinate the two changes if both land in this commit
- **Decisions log:** entry — mutation schema change

<br/>

## Step 8 — §5.7 (NEEDS CONFIRMATION): `Location.WorldID == ""` in-flight sentinel needs a helper

- **Verify:** confirm Decision #226 is still the in-flight encoding (`WorldID == ""`, `RegionHex` set). LSP-search every call site that reads `WorldID == ""` or constructs a `Location` with empty `WorldID` — the review names `transport.go`, `mutation.Apply`, `RebuildIndex`, `digest/build.go`
- **Fix (if confirmed):** add `domain.IsInFlight(loc Location) bool`; route every call site through it. Optionally introduce `domain.WorldIDInFlight = ""` as a named constant if direct equality reads cleaner anywhere
- **Decisions log:** entry — codifies an invariant that was implicit

<br/>

## Sequencing notes

- **Steps 1–3 (coherence) are independent** of each other and of the engine-bug steps. Land them first — they're well-specified and have no verification cost
- **Step 7 (§5.6, mutation schema)** has the most blast radius — writers, readers, and possibly history serialization (C1). Save for last so any earlier test churn settles before refactoring the wire format
- **Re-verify before fixing:** if any §5.x bug is already resolved by intervening commits, skip with a one-line note in the commit message
