# Effect-keyed A-dispatch + SWN re-prefix (A.001.1) — Implementation Plan

## Context / Goal

The `data-driven-rulebook` arc's spine entry (A.001.1), gating A.001.2. Two ordered
efforts:

- **E1 — effect-keyed dispatch (R.018).** Re-key `ability.Dispatch` off the ability's
  effect (`def.Ability.Effect`) instead of the asset definition ID (`def.ID`). Pure
  re-key — no registry authored, no behavior change.
- **E2 — SWN re-prefix.** Rename SWN *asset* IDs (`C1-001` → `SWN-C1-001`) across
  `rulebooks/swn/assets/`, then chase every stranded reference. Sequenced after E1 so
  the rename never touches dispatch code (E1 deletes the ID-keyed map).

This initiative runs Discovery folded into Plan per arc-plan **AP-5**: its one routed
open question is a binary fact-check, answered below, not a design exploration.

- Arc plan: [`data-driven-rulebook-arc-plan.md`](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md) — "A.001.1" section, AP-1, AP-5.
- Arc discovery: [`data-driven-rulebook-arc-discovery.md`](../arcs/data-driven-rulebook/data-driven-rulebook-arc-discovery.md) — "Per-initiative seed", SWN re-prefix block.

## Decisions Ratified in Planning

1. **Routed open question resolved: no scaffolded campaign needs re-prefix migration.**
   `campaigns.CopyRulebook` copies the source rulebook TOMLs into `<campaign>/rulebook/`
   at scaffold time, and `rulebook.Load` globs that *own copy* (`Paths().FactionDataDir`).
   A campaign is self-contained — renaming the source `rulebooks/swn/` leaves existing
   campaigns internally consistent on their old IDs (copied rulebook + `state.toml`
   `definition_id` refs + embedded instance IDs like `iron-vanguard-C1-002-1` all agree).
   The only on-disk campaign is the committed `campaigns/test-camp/`, which no test
   references. **It stays on old IDs** — no migration, and it doubles as proof of
   self-containment. (Answers arc-discovery's "likely none this early".)

2. **Re-prefix touches assets only; goal IDs are untouched.** `G-012` is a *goal* ID
   (`change_homeworld.go` ×2, `goals/change_homeworld.go`), not an SWN asset, and goals
   move to rulebook-neutral `_core` in A.001.4 where an `SWN-` prefix would be wrong. The
   arc-discovery's "chase the G.012 goal ref" is read as imprecise bundling of B.001's
   two hardcoded IDs; for an *asset* re-prefix it is out of scope. The asset-ID regex
   `[CFW][0-9]-[0-9]{3}` naturally excludes `G-012`. B.001's goal-half stays open.

3. **`stealth_applicator` is already gone — nothing to chase.** Verified: zero
   references in any `.go`/`.toml` file (docs only). B.001's asset-half is moot; do not
   re-introduce a reference hunting for it.

4. **Chase-list expands beyond the arc artifacts: `buy_asset.go` hardcodes `C3-002`.**
   The arc artifacts named only `stealth_applicator` and `G.012`. The grounding sweep
   found `buy_asset.go:62,102` keying control flow on the literal `"C3-002"` (Covert
   Shipping). It is an SWN asset ID and **is re-prefixed** to `SWN-C3-002`. De-hardcoding
   it (B.001 / R.004) is *not* done here — rename only.

5. **E1 needs a `def.Ability == nil` guard.** Keying on `def.ID` was nil-safe; keying on
   `def.Ability.Effect` is not. Assets reaching `Dispatch` without an `[ability]` block
   must fall through to `confirmApplied`, not panic.

6. **Inline mock asset IDs in `dispatch_test.go`/`use_asset_ability_test.go` are
   re-prefixed for consistency.** They construct self-contained `AssetDefinition` mocks
   (no real-data load) and are behaviorally irrelevant after E1 keys on effect — but the
   initiative's scope is a *dataset-wide* SWN ID rename, and leaving bare `C1-002`
   literals beside `SWN-`-prefixed data reads as a miss. Low cost, scoped substitution.

7. **Known consequence, accepted (not fixed here): post-E2 engine + pre-E2 campaign.**
   With `buy_asset.go` rewritten to `SWN-C3-002`, a campaign scaffolded *before* E2 (old
   `C3-002` IDs) loses the Covert Shipping special-case — the exact B.001 fragility. No
   correctness impact: there are no real campaigns (Decision 1), and `test-camp` is not
   test-exercised. This reinforces B.001's priority; it does not block A.001.1.

## Decision Record — Execution

1. **E1 expanded: the handler is renamed to its effect (Commit 1).** The plan said
   "No other file changes" for Commit 1, but once dispatch keys on `def.Ability.Effect`,
   the handler named after one asset (`informers`, the Informers asset) is misnamed — it
   implements `reveal_stealth`, which any asset may route to. Reversal of "dispatch.go
   only": Commit 1 also renames `informers.go` → `reveal_stealth.go`, `informers` →
   `revealStealth`, `informersCandidates` → `revealStealthCandidates`, the error string
   `"informers:"` → `"reveal_stealth:"`, and the `TestDispatch_Informers_*` test names →
   `TestDispatch_RevealStealth_*`. Pure rename; no behavior change. (Robert directed,
   full-rename scope.)

2. **Plan's "both existing tests" miscount — harmless.** `dispatch_test.go` has *three*
   tests, not two; the third (`TestDispatch_Stub_ConfirmApplied`) uses `EffectCoinDrain`,
   absent from the new map, so it falls through to `confirmApplied` exactly as before. All
   three pass unchanged. No code impact; noted so the plan body doesn't mislead.

3. **Drive-by fix on this branch (separate `fix:` commit, out of A.001.1 scope).** Two
   pre-existing test failures on clean HEAD, unrelated to effect-keyed dispatch, fixed at
   Robert's request:
   - *Scavengers:* `tags_test.go` tagged the faction `T-014` (Psychic Academy); the
     rulebook and handler agree Scavengers is `T-016`. Stale test → `T-016`.
   - *Turn order:* `full_cycle_test.go` and `RunFactionTurn`'s doc comment expected
     goal-lock first; the engine runs it just before the action phase (income → movement
     → goal-lock → action). **Confirmed engine order is canonical** (goal-lock gates the
     action; a locked/skipped faction still collects income and resolves movement,
     consistent with SWN's "no actions during homeworld move"). Stale test + comment
     updated to match. Candidate to graduate to an architecture `Key Decisions` entry at
     pre-merge.

4. **E2 consistency re-prefix expanded to `faction_state_test.go` (Commit 2).** Plan
   Task 5 enumerated only `dispatch_test.go` and `use_asset_ability_test.go` for the
   Decision-6 consistency substitution, but the Verification grep expects *zero* bare
   `[CFW][0-9]-[0-9]{3}` matches in `internal/`. Grounding found a sixth file:
   `faction_state_test.go:36` carries `DefinitionID: "F1-001"` in an in-memory save/load
   round-trip mock (it never loads the real rulebook, so the ID is behaviorally inert —
   same footing as the Task-5 mocks). Re-prefixed to `SWN-F1-001` so the verification grep
   stays clean. (Robert directed; Decision 6's rationale extended one file.)

5. **Pre-merge: graduated effect-keyed dispatch to `actions.md` Key Decisions.** The
   R.018 re-key is the branch's load-bearing architectural change but lived only in the
   plan; `actions.md` (which owns the dispatch site) described the ability fold without
   stating the keying. Added a Key Decision — "A-dispatch keys on the effect, not the
   asset ID" — there. The turn-order candidate flagged in Decision 3 needed no graduation:
   `orchestrator.md` already documents goal-lock's adjacency to the action phase it gates.

## Out of Scope

- **The A-side composition pipeline and registry** (A.001.2) — E1 re-keys dispatch, it
  does not author a registry. The nine former stubs collapse into the existing
  unknown-effect fallthrough; they are *not* re-implemented.
- **De-hardcoding `C3-002` / `G-012`** — B.001 / R.004 territory. E2 renames the
  literals; it does not lift them into data.
- **Goal/tag IDs and `_core`** — A.001.4. `G-012` and any goal/tag ID stay as-is.
- **Migrating `campaigns/test-camp/`** — left on old IDs (Decision 1).
- **Contest/roll extraction** beyond what re-keying strictly needs (A.001.2).

---

## Work Breakdown

Two commits, each one execution session. Commit 2 (E2) must follow Commit 1 (E1): E1
deletes the ID-keyed dispatch map, so the rename never touches dispatch code.

### Phase 1 — Effect-keyed dispatch (E1 / R.018)

#### Commit 1 — `refactor: key ability dispatch on effect not definition ID`

##### Task 1 — `internal/faction/engine/action/actions/ability/dispatch.go`

Re-key the handler map on `domain.AbilityEffectType` and route `Dispatch` on
`def.Ability.Effect`, with a nil-`Ability` guard. The nine `confirmApplied` stubs are
deleted — unknown/absent effects already fall through to `confirmApplied`.

**Find:**
```go
var handlers = map[string]AbilityHandler{
	"C1-002": informers, // Informers

	// Structural stubs — real implementations land in F-014.
	"W1-002": confirmApplied, // Harvesters
	"W3-001": confirmApplied, // Postech Industry
	"W7-001": confirmApplied, // Pretech Manufactory
	"F5-002": confirmApplied, // Pretech Logistics
	"W6-001": confirmApplied, // Venture Capital
	"W6-003": confirmApplied, // Commodities Broker
	"C4-004": confirmApplied, // Seditionists
	"W5-001": confirmApplied, // Marketers
	"W4-002": confirmApplied, // Monopoly
}

// Dispatch routes an asset's ability to its registered handler.
// Falls through to confirmApplied for unregistered IDs (defensive; should not occur post-flag-corrections).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	handler, ok := handlers[def.ID]
	if !ok {
		return confirmApplied(faction, asset, def, collector, roller, factionState, rulebook)
	}
	return handler(faction, asset, def, collector, roller, factionState, rulebook)
}
```
**Replace with:**
```go
var handlers = map[domain.AbilityEffectType]AbilityHandler{
	domain.EffectRevealStealth: informers,
}

// Dispatch routes an asset's ability to the handler registered for its effect.
// Assets with no ability block, or an effect with no registered handler, fall
// through to confirmApplied (the not-yet-built effects).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	if def.Ability == nil {
		return confirmApplied(faction, asset, def, collector, roller, factionState, rulebook)
	}
	handler, ok := handlers[def.Ability.Effect]
	if !ok {
		return confirmApplied(faction, asset, def, collector, roller, factionState, rulebook)
	}
	return handler(faction, asset, def, collector, roller, factionState, rulebook)
}
```

No other file changes. `dispatch_test.go` already builds its fixtures with
`Effect` set (`EffectRevealStealth` → `informers`; `EffectCoinDrain` →
fallthrough → `confirmApplied`), so both existing tests pass unchanged — verify,
don't edit.

##### Commit message
```
refactor: key ability dispatch on effect not definition ID

- re-key ability.handlers map on domain.AbilityEffectType
- route Dispatch on def.Ability.Effect with a nil-Ability guard
- drop the nine confirmApplied ID stubs (absorbed by fallthrough)
```

### Phase 2 — SWN re-prefix (E2)

#### Commit 2 — `refactor: re-prefix SWN asset IDs with SWN-`

Sequenced after Commit 1. Each task is a substitution scoped so it touches only SWN
asset-ID literals (`[CFW][0-9]-[0-9]{3}`), never instance IDs (`...-C1-002-1`), goal
IDs (`G-012`), or TOML table names (`[assets.Smugglers]`).

##### Task 1 — `rulebooks/swn/assets/{cunning,force,wealth}_assets.toml`

Rename every `id =` field. Verified: the only `[CFW][0-9]-[0-9]{3}` matches in these
files are `id =` lines (descriptions are prose, no cross-asset refs), and
`loadAssets` keys the map on the `id` value — so this is the complete source change.

```sh
sed -i -E 's/^id = "([CFW][0-9]-[0-9]{3})"/id = "SWN-\1"/' \
  rulebooks/swn/assets/cunning_assets.toml \
  rulebooks/swn/assets/force_assets.toml \
  rulebooks/swn/assets/wealth_assets.toml
```

##### Task 2 — `internal/faction/engine/action/actions/buy_asset.go`

Re-prefix the two hardcoded Covert Shipping checks.

**Find:**
```go
	if order.Definition.ID == "C3-002" {
```
**Replace with:**
```go
	if order.Definition.ID == "SWN-C3-002" {
```

**Find:**
```go
	if ba.buyOrder.Definition.ID != "C3-002" {
```
**Replace with:**
```go
	if ba.buyOrder.Definition.ID != "SWN-C3-002" {
```

##### Task 3 — `internal/faction/engine/testharness/harness.go`

The harness loads the real `rulebooks/swn`; its ID constants must follow.

**Find:**
```go
const (
	DefSecurityPersonnel = "F1-001"
	DefHeavyDropAssets   = "F2-001"
)
```
**Replace with:**
```go
const (
	DefSecurityPersonnel = "SWN-F1-001"
	DefHeavyDropAssets   = "SWN-F2-001"
)
```

##### Task 4 — `internal/faction/rulebook/rulebook_test.go`

`rulebook_test.go` loads the real `../../../rulebooks/swn` and asserts on bare asset
IDs throughout (`rb.Assets["F1-001"]`, the transport/movement tables, etc.). Re-prefix
every asset-ID literal. The file has no instance IDs, so the scoped regex is safe:

```sh
sed -i -E 's/"([CFW][0-9]-[0-9]{3})"/"SWN-\1"/g' \
  internal/faction/rulebook/rulebook_test.go
```

##### Task 5 — `internal/faction/engine/action/actions/ability/dispatch_test.go`, `internal/faction/engine/action/actions/use_asset_ability_test.go`

Re-prefix the inline mock asset IDs for dataset consistency (Decision 6). Both files
use instance IDs that do *not* match the regex (`"t1"`, `"a1"`), so the scoped
substitution touches only the definition IDs (`"C1-002"`, `"W1-002"`):

```sh
sed -i -E 's/"([CFW][0-9]-[0-9]{3})"/"SWN-\1"/g' \
  internal/faction/engine/action/actions/ability/dispatch_test.go \
  internal/faction/engine/action/actions/use_asset_ability_test.go
```

##### Verification

After all tasks:
```sh
grep -rnE '"[CFW][0-9]-[0-9]{3}"' --include='*.go' internal/   # expect no matches (all now SWN-)
go build ./... && go test ./...
```
`campaigns/test-camp/` is intentionally left on old IDs (Decision 1) and is not loaded
by any test, so it does not affect the suite.

##### Commit message
```
refactor: re-prefix SWN asset IDs with SWN-

- rename id fields across rulebooks/swn/assets/*.toml (C1-001 -> SWN-C1-001)
- chase the hardcoded buy_asset.go Covert Shipping check (C3-002)
- re-prefix testharness ID constants and asset-ID test assertions
- test-camp left on old IDs (self-contained, no migration needed)
```
