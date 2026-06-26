# A-flag Ability Registry (A.001.2) — Implementation Plan (Overview)

## Context / Goal

Build the A-side composition pipeline the [discovery](../discovery/a-flag-ability-registry-discovery.md)
designed: *target-selection → (stat-contest | die-roll + outcome-table) → effect →
mutations*, keyed on the ability's effect (ACD-3). A.001.1 re-keyed dispatch onto
the effect but left exactly one real handler (`reveal_stealth`) and no machinery to
*compose* a special from data. A.001.2 grows `AbilityDefinition` into a step-carrying
block, extracts the inline stat-contest into a generic stage, gives `Die`/`Outcomes`
a parser path, owns the shared `relocate` world-engine primitive (ACD-4), and authors
the full A-flag table from prose (D4).

- Arc: [data-driven-rulebook](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md),
  initiative A.001.2. Cross-cutting ACD-1…ACD-9 bind here without re-litigation.
- Mechanical input: the [ability & effect catalog](../../architecture/engine/ability-catalog.md) (ACD-8).
- This is a **multi-effort** initiative (section 8). This file is the overview plan;
  Efforts 1 and 3 get their own per-effort Plan session (see Effort Summary).

## Decisions Ratified in Planning

1. **Four-effort slicing, DAG `1 → {2, 3, 4}`.** Effort 1 (pipeline + parser
   foundation) gates everything; Efforts 2 (roll-table), 3 (relocate), 4
   (contest/enabler/attach) each depend only on 1 and are mutually independent.
   Resolves the discovery's "effort slicing" open question.

2. **Value-dependent roll payouts are enumerated, not evaluated.** Pretech
   Manufactory's `ceil(roll/2)` and Commodities Broker's `discount = roll` are
   expressed as explicit range→payout rows in `Outcomes` (Manufactory collapses to
   4 rows; Commodities is 8 one-per-face rows). **No `formula` field, no expression
   evaluator** — the engine does only the range-lookup the fixed-payout assets
   already need, and payouts stay visible/editable in data (ACD-5, YAGNI). Resolves
   the D7 sub-detail.

3. **`relocate` is a new world-engine method**, `WorldEngine.BuildRelocateMutations`
   — validates hex-distance via the `HexRouter` seam and emits `AssetMoved` (+
   optional `CoinDelta`), instant and distinct from the drift-order path. The
   relocate ability-step calls it; the world engine is threaded into the ability
   dispatch path. Keeps geometry behind the world seam (world-movement Key
   Decisions; ACD-4 "one primitive, many harnesses"). Resolves the "relocate move
   API + mutations" open question.

4. **`relocate` reuses the existing `AssetMoved` mutation — no new mutation type.**
   Re-grounding found `AssetMoved` (`domain/mutation.go:157`) already minted,
   applied (`mutation/mutation.go:181`), digest-aware (`narrative/digest/build.go`),
   and emitted today by the transport cargo-follow reactor — its doc comment already
   reads *"Emitted by movement ability steps."* Effort 3 supplies an emitter, not a
   vocabulary entry.

5. **Drop the legacy `type` step discriminator.** Seductress's TOML step carries
   `type = "faction_test"` (`cunning_assets.toml:130`), a closed step-kind marker
   that contradicts D2 (a step *is* an effect-unit; no closed vocabulary). The
   parser ignores/rejects it; the field is removed from authored TOML.

6. **The domain shape (D6/D7) is the contract Effort 1 builds.** `AbilityDefinition`
   carries `Steps []AbilityStep`; a step holds optional `Contest` **xor** `Roll`, an
   `Effect`, a `Target`, and typed per-effect param sub-structs (`Relocate`,
   `Attach`, `Purchase`) set iff their effect applies. `AbilityOutcome` generalizes
   from `{From,To,Coin}` to add signed `Coin`, `DestroySelf`, `Discount`. The literal
   Go lands in Effort 1's first task; this overview ratifies the shape as the shared
   contract (see Shared Context).

7. **Target-selection generalizes into new collector methods.** Today's
   `SelectFactionTestTarget` is reveal-specific. Effort 1 introduces a generic
   faction/asset/world target stage; Efforts 3 and 4 add the asset-target and
   world-target collector methods they need. Exact signatures are an Effort-1 /
   first-use detail (see Open Questions).

8. **Per-effort planning.** Efforts 1 and 3 are design-heavy (pipeline + domain
   reshape; world-engine wiring + new seam) and get their own lightweight Plan
   session. Efforts 2 and 4 are handler-authoring on an established pipeline; they
   are specified at commit granularity in this overview and execute straight from
   it, re-grounding against Effort 1's built API at execution time.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

- **Effort 2 gets its own Plan session (deviation from Planning Decision 8).** Decision 8
  slated Efforts 2 and 4 to execute straight from this overview. At Effort 2's planning
  Robert chose a per-effort plan doc instead — re-grounding against Effort 1's built API
  surfaced two design decisions worth ratifying in writing (the roll-stage / `RollResult`
  threading, the insolvency-vs-`DestroySelf` destroy rule, and the pending-discount carry
  mechanism). The detail now lives in
  [`-effort-2-plan.md`](./a-flag-ability-registry-effort-2-plan.md); this overview's
  Effort-2 Work Breakdown is superseded by it. Decision 8 still stands for Effort 4 unless
  similarly revisited.

---

## Open Questions — To Ratify at Implementation Time

- **Generic target-selection collector signatures** (asset-target, world-target) —
  resolved in Effort 1's Plan session; Efforts 3/4 consume them.
- **Multi-step failure semantics** — when an early step's contest/roll fails, does a
  later step still run? Worked case: Seductress relocates (step 1) then contests
  (step 2) — the relocate should stand even if the contest loses. Ratify the
  step-loop's collect-vs-halt rule in Effort 1.
- **`coin_or_disable` resolution** (Marketers `W5-001`) — "pay half an asset's cost
  or have it disabled" needs a target-asset choice and a disable representation.
  Resolve at Effort 4 first use; check whether "disabled" reuses an existing asset
  flag/mutation or needs a new one.
- **`attach` / `restrict-action` representation** (Seditionists `C4-004`) — attaching
  to an enemy asset and suppressing its attack has **no existing vocabulary**. Decide
  in Effort 4 whether it needs a new mutation, an asset field, or an effect-engine
  hook (it may reveal itself as S-side and reroute — flag early). This is the
  initiative's one genuine unknown.

## Effort Summary

DAG: Effort 1 is foundational; 2, 3, 4 depend on 1 and are mutually independent.
An optional interim effort review (section 8) is worth considering at the Effort 1
boundary, since 2/3/4 all build on its API.

| # | Effort | Scope | Plan |
|---|--------|-------|------|
| 1 | Pipeline + parser foundation | Step model, generic Contest/Roll/target stages, grow `abilityRecord` + flat-sugar, generalize `AbilityOutcome`, remove dead coin enum. Convert Informers + Seductress(reveal) as proof. No new effects. | own session → [`-effort-1-plan.md`](./a-flag-ability-registry-effort-1-plan.md) |
| 2 | Roll-table effects | `coin_delta` + `purchase_discount` handlers (Roll stage + enumerated `Outcomes`). Author Harvesters, Postech, Venture Capital, Manufactory, Commodities Broker. | own session → [`-effort-2-plan.md`](./a-flag-ability-registry-effort-2-plan.md) |
| 3 | Relocate primitive | `WorldEngine.BuildRelocateMutations`, thread world engine into dispatch, `relocate` effect + self/other target. Author Covert Shipping, Covert Transit Net, Seductress self-step (completes Seductress end-to-end). | own session → [`-effort-3-plan.md`](./a-flag-ability-registry-effort-3-plan.md) |
| 4 | Contest / enabler / attach | `coin_or_disable` (Marketers), `purchase_enabler` (Pretech Logistics), `attach`/restrict-action (Seditionists). | this overview (Work Breakdown) |

## Shared Context

**The per-step pipeline.** Every ability — single- or multi-effect — runs the same
loop: for each `AbilityStep`, resolve its target (faction/asset/world/self/none),
run its optional resolution stage (`Contest` xor `Roll`), dispatch its terminal
`Effect` to the handler registered for it, and collect the returned mutations. A
single-effect ability is a one-step list (the flat `[ability]` block is sugar). The
loop does not special-case arity. `reveal_stealth` stops owning target-selection and
the contest (now generic stages) and shrinks to "given the resolved target, emit the
stealth-cleared mutations."

**Dispatch keying.** Handlers remain `map[domain.AbilityEffectType]AbilityHandler`
(`ability/dispatch.go:21`). A step whose effect has no registered handler still falls
through to `confirmApplied`; this lets the authoring pass (D4) land description-only
specials that don't yet have a handler without erroring.

**Domain shape (contract; literal Go is Effort 1 Task 1).** Direction ratified in
Planning Decision 6 — `AbilityDefinition.Steps []AbilityStep`; `AbilityStep{Effect,
Target, Contest*, Roll*, Relocate*, Attach*, Purchase*}`; `AbilityOutcome{From, To,
Coin (signed), DestroySelf, Discount}`. `EffectCoinDrain`/`EffectCoinSteal` and their
parser cases are removed (D8; S-side per D1).

**Authoring scope (D4/D5).** Effort 1 authors Informers (flat) + Seductress(reveal);
each later effort authors its family's assets. The D5 prose cleanup — correcting
"as an action, may move itself within N hexes" on Speed-based assets (Strike Fleet,
Space Marines, Capital Fleet, Scavenger Fleet, Blockade Fleet's move clause, …) — is
data-only description editing with no mechanic change; fold it into whichever effort
touches each file, or batch it. Genuine action-bundled moves (Seductress) keep their
prose.

**Test patterns.** Follow `ability/dispatch_test.go`: a `seqRoller{values}` for
deterministic rolls and a gomock `MockCollector` for prompts. Each new handler gets
attacker-wins / defender-wins (or roll-band) coverage; the step loop gets a
multi-step test once Effort 3 authors Seductress's two real steps.

**Re-grounding deltas (verified this session).** `AssetMoved` already exists and is
wired (Decision 4). The ability dispatch path has **no world-engine access** today —
`NewUseAssetAbility(c, resolveRoller())` and `ability.Dispatch`'s signature stop at
the rulebook; Effort 3 threads it, mirroring how `Attack`/`ExpandInfluence` receive
`worldIndex(worldEngine)` in `register.go`. The legacy `type` step field must be
dropped (Decision 5).

## Out of Scope

All S-side work, deferred to A.001.3+: coin reactors, Monopoly (S-side `MutationReactor`
per D1), the *free* relocate harness (Transit Web), wiring `GrantedMovementAbilities`.
Goals/tags and `_core` (A.001.4); H.A.W.K authoring (A.001.5). Speed-based
movement-phase moves are untouched — only their misleading prose is corrected (D5).

---

## Work Breakdown

Efforts 1 and 3 carry their Phase/Commit/Task detail in their own per-effort plan
files (written in their own Plan sessions). The commit outlines below for those two
are **scoping guidance for those sessions**, not executable tasks. Efforts 2 and 4
are specified at commit granularity to execute straight from this overview; their
literal-code tasks bind to Effort 1's final API, so each execution session
re-grounds against the built pipeline first (section 10).

### Effort 1 — Pipeline + parser foundation *(own Plan session)*

Scoping outline for the Effort-1 Plan session:

- `refactor: reshape ability domain into steps` — `AbilityDefinition.Steps`,
  `AbilityStep`, generalize `AbilityOutcome`, remove `EffectCoinDrain`/`CoinSteal`
  and their `toAbilityEffect` cases. (Defines the shared structures per KEY RULE 2.)
- `feat: parse ability steps with flat-block sugar` — grow `abilityRecord` into the
  step-carrying form; flat `[ability]` parses to a one-step list; `[[ability.steps]]`
  is the explicit form; parse `Contest`/`Roll`(`Die`+`Outcomes`); drop the legacy
  `type` discriminator.
- `feat: per-step ability pipeline with generic stages` — `Dispatch` loops steps;
  extract the inline stat-contest (`reveal_stealth.go:31-32`) into a generic
  `Contest` stage; generic target-selection stage; shrink `revealStealth` to the
  terminal effect; ratify the multi-step failure semantics (Open Questions).
- `feat: author Informers and Seductress reveal step` — Informers as flat sugar,
  Seductress's reveal as a one-step list (its relocate step lands in Effort 3);
  tests via `seqRoller` + `MockCollector`.

### Effort 2 — Roll-table effects *(own Plan session)*

Planned in its own session and re-grounded against Effort 1's built API — see
[`-effort-2-plan.md`](./a-flag-ability-registry-effort-2-plan.md) for the executable
Commit 1 (`coin_delta`) / Commit 2 (`purchase_discount`) task detail and the decisions
ratified there. The scoping outline that previously lived here is superseded by that
file (see Decision Record).

### Effort 3 — Relocate primitive *(own Plan session)*

Scoping outline for the Effort-3 Plan session:

- `feat: world-engine instant relocate primitive` — `WorldEngine.BuildRelocateMutations`
  (validate `Distance(from,to) ≤ maxHex` via `HexRouter`, emit `AssetMoved` +
  optional `CoinDelta`); thread the world engine into `UseAssetAbility` /
  `ability.Dispatch` (mirror `register.go` wiring).
- `feat: relocate ability effect (self / other)` — `relocate` handler + `RelocateParams`
  (target self vs other, `maxHex`, `coinCost`); asset-target selection for the
  "other" case; author **Covert Shipping** `C3-003`, **Covert Transit Net** `C6-002`,
  and **Seductress's self-relocate step** — completing Seductress as a live two-step
  ability (the step engine's first real multi-step consumer; add the multi-step test).

### Effort 4 — Contest / enabler / attach

Depends on Effort 1. Re-ground against the built contest/target stages before coding.

#### Commit 1 — `feat: coin_or_disable contest ability`

- New handler for the `coin_or_disable` effect: generic `Contest`; on success, the
  target faction pays half a chosen asset's cost or that asset is disabled.
- Author **Marketers** `W5-001`.
- Ratify at first use (Open Questions): asset-target choice surface and the "disabled"
  representation — reuse an existing flag/mutation if one fits, else introduce one.
- Tests: contest win → pay-or-disable branch; contest loss → no effect.

#### Commit 2 — `feat: purchase_enabler ability`

- New handler for the `purchase_enabler` effect + `PurchaseParams`: spend the action
  to buy one discounted Force asset on the world (world-target selection).
- Author **Pretech Logistics** `F5-002`.
- Tests: enabled purchase emits the expected `AssetAdded` + discounted `CoinDelta`.

#### Commit 3 — `feat: attach / restrict-action ability`

- New handler for the `attach` effect + `AttachParams`: pay `1d4` Coin to attach to
  an enemy asset, suppressing its attack.
- Author **Seditionists** `C4-004`.
- **Genuine unknown (Open Questions):** attach + restrict-action has no existing
  vocabulary. Decide the representation first (new mutation / asset field / effect
  hook) — and if it resolves as passive/reactive, reroute it to S-side (A.001.3) and
  drop this commit, noting the reroute in the Decision Record.
- Tests: attach emits the Coin cost + the restriction representation.
