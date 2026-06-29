# A-flag ability registry (A.001.2) — Discovery

> Initiative of arc `A.001` (data-driven-rulebook). Cross-cutting decisions
> ACD-1…ACD-9 are ratified in the
> [arc-discovery](../arcs/data-driven-rulebook/data-driven-rulebook-arc-discovery.md)
> and bind here without re-litigation. Boundary set by the
> [arc-plan](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md), A.001.2.
> Mechanical decomposition input: the
> [ability & effect catalog](../../architecture/engine/ability-catalog.md) (ACD-8).

## Problem

A.001.1 re-keyed A-side dispatch onto the ability's effect, but the machinery to
*compose* a special from data does not exist yet. Concretely:

- **No composition pipeline.** `ability/dispatch.go` routes on
  `domain.AbilityEffectType`, but there is exactly one real handler
  (`reveal_stealth`); everything else falls through to `confirmApplied`.
- **The stat-contest is not a primitive.** It is hard-wired inside the
  `reveal_stealth` handler (`reveal_stealth.go:31-32` — `roller.Roll(10) + statScore`
  for both sides, inline). Any other contest-based effect would have to re-implement it.
- **The roll-table is dead.** `domain.AbilityDefinition` carries `Die *DiceRoll` and
  `Outcomes []AbilityOutcome` (`asset.go:58-59`), but `abilityRecord`
  (`rulebook.go:101-105`) has no `die`/`outcomes` parser path — they cannot be filled
  from TOML. The catalog's `roll-table-coin` effects (Harvesters, Postech, Venture
  Capital, Pretech Manufactory) need them.
- **Speculative dead enum entries.** `EffectCoinDrain`/`EffectCoinSteal` exist as
  `AbilityEffectType` constants with parser cases (`rulebook.go:240-243`), but no asset
  declares either, and the catalog routes both coin effects to the **S-side**.

## Architecture Context

Pages read for re-grounding (their `Key Decisions` are binding context):

- [actions](../../architecture/engine/actions.md) — the A-side action pipeline that
  hosts `Use Asset Ability`.
- [effect-mutation](../../architecture/engine/effect-mutation.md) — handlers return
  `[]domain.Mutation`; the rule layer registers, never dispatches.
- [hooks](../../architecture/engine/hooks.md) — the five canonical categories; the
  `AssetMovementGranter` (Cat 4) seam relevant to the relocate harness.
- [world-movement](../../architecture/engine/world-movement.md) — the drift-order
  movement path the new instant-relocate primitive must stay distinct from.
- [ability & effect catalog](../../architecture/engine/ability-catalog.md) — the A-flag
  primitive table (ACD-8, binding).

## Design Summary

The A-side composition pipeline (ACD-3): *target-selection →
(stat-contest | die-roll + outcome-table) → effect → mutations*, keyed on the effect.
An ability is a **uniform sequence of steps** (D2) — each step an effect-unit with its
own optional contest/roll; a single-effect ability is a one-step list. A.001.2 builds
the per-step pipeline, extracts the inline stat-contest into a generic stage, gives
`Die`/`Outcomes` a parser path, and grows `abilityRecord` into the step-carrying block.
It also owns the shared `relocate` world-engine primitive (ACD-4) and its action-spent
harness.

**Spine scope, post-Monopoly decision:** contest / roll / relocate. No coin handlers
land in A.001.2 (see Decisions).

**Scope is the full A-flag table, authored from prose (D4).** Only Informers is modeled
today; every other A-flag special is description-text only. A.001.2 authors the
`[ability]` block for each and builds a handler per effect family. From the catalog's
A-flag table (Monopoly excluded → S-side per D1), the families are:

| Effect family | Assets (SWN-) | World reach? |
|---|---|---|
| stat-contest → reveal-stealth | Informers `C1-002` ✅, Seductress `C2-004` (+ self-move) | no |
| roll-table → coin | Harvesters `W1-002`, Postech Industry `W3-001`, Venture Capital `W6-001`, Pretech Manufactory `W7-001` | no |
| roll-table → purchase-discount | Commodities Broker `W6-003` | no |
| stat-contest → coin-or-disable | Marketers `W5-001` | no |
| relocate (self / other) | Seductress `C2-004`, Covert Shipping `C3-003`, Covert Transit Net `C6-002` | **yes** (world engine) |
| purchase-enabler | Pretech Logistics `F5-002` | no |
| attach / restrict-action | Seditionists `C4-004` | no |

## Decisions (settled this session)

- **D1 — Monopoly is an S-side reactor; it ships in A.001.3, not here.** Flagged `A`
  but its mechanic fires off a combat outcome ("when it successfully attacks…"), not a
  spent action. Per ACD-1 (mechanic picks the engine), it is a Cat 3 `MutationReactor`,
  matching the catalog's filing. **Consequence:** A.001.2 builds no coin handlers; the
  coin effects are A.001.3's. The speculative `EffectCoinDrain`/`EffectCoinSteal` enum
  entries are candidates for removal here (see Open Question Q-coin).

- **D2 — An ability is a uniform sequence of steps; a step *is* an effect-unit.**
  There is no closed vocabulary of step-kinds. Every step has the same shape — the one
  today's single-effect block describes (an `effect`, with optional `contest`, optional
  `roll`, and its target/params). A multi-step ability is an ordered list of those; a
  single-effect ability is a list of one. The engine runs the steps in order and
  collects mutations; it does not special-case arity.
  - **Authoring:** the flat `[ability]` block is **sugar** that parses into a one-step
    list; `[[ability.steps]]` is the explicit general form. Both accepted.
  - **Domain:** `AbilityDefinition` carries `Steps []AbilityStep`; a step holds the
    optional `Contest`, optional `Roll` (`Die`+`Outcomes`), `Effect`, and target/params.
    The flat `attacker_stat`/`defender_stat` fold into a step's `Contest`; `Die`/
    `Outcomes` fold into a step's `Roll`. This resolves the dropped `[[ability.steps]]`
    (kept and generalized) and gives the dead roll-table fields a parser path.
  - **Consequence:** *every* ability runs the per-step pipeline, including Informers.
    The `reveal_stealth` handler stops owning the contest and target-selection (now
    generic per-step stages) and shrinks to "given the resolved target, emit the
    stealth-cleared mutations."
  - Worked: Informers = 1 step (`reveal_stealth` + contest); Seductress = 2 steps
    (`relocate_self` then `reveal_stealth` + contest).

- **D3 — One parameterized `relocate` primitive, not two.** `relocate_self` (Seductress:
  move the acting asset, 1 hex, no Coin) and `relocate-other` (Covert Shipping / Covert
  Transit Net: move a chosen friendly asset, 3 hex, 1 Coin) are the **same world-engine
  move** parameterized by target (self vs other), Coin, and range — ACD-4's "one
  primitive, many harnesses." Building it to cover both is barely more than
  `relocate-other` alone, and it keeps Seductress wired end-to-end so the step engine
  has a live consumer in A.001.2. **Resolves Q-relocate-kinds.**
  - **This is distinct from Speed-based self-movement.** An asset with `Speed > 0`
    moves itself in the **movement phase** (`eligibleMovableAssets`,
    world-movement.md:104) — that is *not* an ability and *not* relocate. The relocate
    primitive is the *instant* action/free move beyond the asset's own Speed. Seductress
    is a deliberate **special case**: it bundles an instant self-relocate into its
    *action* (step 1), rather than using its Speed-1 movement-phase move. Robert's call:
    keep Seductress multi-step.

- **D4 — Scope is the full A-flag table, authored from prose.** Confirmed: only
  Informers is modeled today; A.001.2 authors the `[ability]` block for every A-flag
  asset (table in Design Summary) and builds a handler per effect family. This is the
  "fill the dead fields" work made concrete — it is authoring, not just a parser path.

- **D5 — Correct the "moves itself as an action" prose on Speed-based assets.** Many
  asset descriptions (Strike Fleet `SWN-F4-003`, Space Marines `SWN-F7-003`, Capital
  Fleet `SWN-F8-001`, Scavenger Fleet `SWN-W8-001`, Blockade Fleet's move clause, …)
  say "as an action, may move itself within N hexes" — but that movement is their
  `Speed`, resolved in the movement phase, not an action ability. The prose is corrected
  (drop the action framing) as part of the authoring pass. **Genuine special cases that
  bundle a move into an action keep their prose** (Seductress). This is data-only
  description cleanup, no mechanic change.

- **D6 — `AbilityStep` shape: resolution + target + typed effect-params.** A step carries
  `Effect` (terminal primitive + dispatch key), `Target` (the selection kind, generalizing
  today's `SelectFactionTestTarget`), an optional resolution stage (`Contest` **xor**
  `Roll`), and typed per-effect param sub-structs (`Relocate`, `Attach`, `Purchase`) set
  iff their effect applies — the same pattern as the existing `attack`/`transport`
  profiles. Validated against all seven A-flag families (see Domain Model Changes).
  **Resolves Q-step-fields.**

- **D7 — Generalize `AbilityOutcome` (Option A: data-driven).** The roll's outcome table
  stays in TOML and grows beyond coin-only to carry the result-kinds the A-flag table
  actually needs: signed `Coin` (gain/loss), `DestroySelf` (Postech `W3-001`, Venture
  Capital `W6-001` on a 1), and `Discount` (Commodities Broker `W6-003`). Per ACD-5 the
  payouts stay visible/editable in the data rather than living in handler code.
  YAGNI-bounded to these result-kinds — not a general mutation-emitter. **Open sub-detail
  (→ Plan):** formula-style payouts — Pretech Manufactory `W7-001` ("half the roll,
  rounded up") and Commodities Broker ("discount = the rolled value") — resolve as either
  enumerated ranges or a small `formula` field.

## Domain Model Changes

The `[ability]` block parses into a step list; the flat block is sugar for a one-step
list (D2). Proposed domain shape (field-by-field details land in Plan):

```go
type AbilityDefinition struct {
    Steps []AbilityStep   // always; flat [ability] block -> 1-step list
}

type AbilityStep struct {
    Effect   AbilityEffectType   // terminal primitive + dispatch key
    Target   TargetKind          // faction | asset | world | self | none
    Contest  *Contest            // optional resolution (xor Roll): attacker/defender stat
    Roll     *RollTable          // optional resolution (xor Contest): Die + Outcomes
    Relocate *RelocateParams     // set iff Effect == relocate
    Attach   *AttachParams       // set iff Effect == attach
    Purchase *PurchaseParams     // set iff Effect == purchase_enabler
}

type AbilityOutcome struct {        // D7 — generalized from {From,To,Coin}
    From, To    int
    Coin        int    // signed: gain or loss
    DestroySelf bool
    Discount    int
}
```

Validation — every A-flag family (Monopoly excluded) maps onto this shape:

| Family | Effect | Target | Resolution | Param struct |
|---|---|---|---|---|
| reveal-stealth | `reveal_stealth` | faction | `Contest` | — |
| roll-table-coin | `coin_delta` | self/none | `Roll` | — (in `Outcomes`) |
| purchase-discount | `purchase_discount` | self/none | `Roll` | — (in `Outcomes`) |
| coin-or-disable | `coin_or_disable` | asset | `Contest` | — |
| relocate | `relocate` | asset + world | — | `Relocate` |
| purchase-enabler | `purchase_enabler` | world | — | `Purchase` |
| attach/restrict | `attach` | asset | — (cost-roll) | `Attach` |

The existing flat `abilityRecord` (`attacker_stat`/`defender_stat`/`effect`) folds into a
one-step list: the stats become the step's `Contest`. `EffectCoinDrain`/`EffectCoinSteal`
are not in this table (S-side per D1) — removed per D8.

## Decisions (continued)

- **D8 — Remove `EffectCoinDrain`/`EffectCoinSteal` from the A-side enum.** They sit in
  `AbilityEffectType` with parser cases but no asset declares them, and D1 routes both
  coin effects to the S-side. Leaving them would let an A-ability declare `coin_drain`
  and silently no-op. Removed here while in this code; A.001.3 re-models them via the
  S-side `[effect]` discriminator. **Resolves Q-coin.**

## Open Questions

All Discovery-level questions are resolved or handed to Plan. Remaining items are
**Plan-session work**, not open design forks:

- **Effort slicing (→ Plan, session-modes section 8).** A.001.2 clearly splits into
  efforts. Candidate seams (Plan finalizes): **(1)** pipeline + parser machinery (step
  model, `Contest`/`Roll` stages, generic target-selection, grow `abilityRecord`),
  converting Informers + Seductress as proof — no new effects; **(2)** roll-table effects
  + the generalized `AbilityOutcome` (Harvesters, Postech, Venture Capital, Manufactory,
  Commodities Broker); **(3)** the relocate primitive (world engine) + action-spent
  harness (Covert Shipping, Covert Transit Net, Seductress's self-step); **(4)** remaining
  contest/enabler/attach (Marketers, Pretech Logistics, Seditionists).
- **Relocate move API + mutations (→ Plan/Execution).** The *home* is arc-settled (world
  engine, a new instant move distinct from the drift-order path); only the concrete
  API/mutation shape is left.
- **Formula-style roll payouts (→ Plan).** Pretech Manufactory / Commodities Broker —
  enumerated ranges or a small `formula` field (D7 sub-detail).

**Resolved this session:** Q-ability-shape and the `[[ability.steps]]` fate (→ D2);
Monopoly's home (→ D1); `relocate_self` vs `relocate_other` (→ D3); full-table authoring
scope (→ D4); Speed-vs-action self-move prose (→ D5); `AbilityStep` shape (→ D6);
outcome generalization (→ D7); coin-enum removal (→ D8).

---

## Current-State Survey (re-grounding, verified this session)

| Claim | Verified in source |
|---|---|
| A-side dispatch is effect-keyed, one real handler | `ability/dispatch.go` — `map[AbilityEffectType]AbilityHandler`, only `reveal_stealth`; rest → `confirmApplied` |
| Stat-contest is inline, not a primitive | `ability/reveal_stealth.go:31-32` |
| `Die`/`Outcomes` have no parser path | `domain/asset.go:58-59` vs `rulebook.go:101-105` (`abilityRecord` = 3 flat fields) |
| `coin_*` effects unwired, no asset declares them | `domain/asset.go:53-54`, `rulebook.go:240-243`; no hit in `rulebooks/` |
| Catalog routes both coin effects to S-side | `ability-catalog.md:134-135` (Cat 3 `MutationReactor`) |
| World movement is the drift-order path | `world/movement.go` — `TickMovementOrders` / `BuildMovementMutations` |
| `GrantedMovementAbilities` defined but unconsumed | `hooks/dispatch/rules.go:54` — no caller (A.001.3 wires it) |

## Out of Scope

- All S-side work: coin reactors, Monopoly, the free relocate harness, wiring
  `GrantedMovementAbilities` (A.001.3).
- Goals/tags, `_core` (A.001.4); H.A.W.K authoring (A.001.5).
