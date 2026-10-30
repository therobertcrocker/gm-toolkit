# Ability Engine Redesign — Discovery

> Initiative: R-001 (`planned-work.md`)
> Related: F-002 Asset Movement Redesign (extracts movement; this redesign reverses one of its planning decisions — see *Cross-cutting impact* below)

## TODO — Discovery WIP

This discovery doc is in-progress. Outstanding items needed before sign-off:

- [x] Q1 — dispatch shape (resolved in-session: Option A asset-ID-keyed handlers)
- [x] Q2 — TOML schema fate (resolved in-session: retire `steps` sub-schema, keep slimmed `ability` parameter block)
- [x] Q3 — shared primitives location (resolved in-session: shared helpers as a sibling subpackage under `action/actions/`)
- [x] Q4 — Tripwire Cells flag anomaly (resolved in-session: drop `A`, keep `S` only)
- [x] Q5 — stub implementation scope (resolved in-session: structural-only; implementation of the 9 stubs added to backlog)
- [x] Q6 — confirmation of deletion scope (resolved in-session: retire list ratified; `domain.Ability` reshape and `ConfirmAbilityApplied` survival called out explicitly)

## Problem

The ability sub-engine at `internal/faction/engine/ability/` was built when the "Use Asset Ability" action was the only landing pad for any asset-driven mechanic. Since then, two things have changed the landscape:

1. The **hooks subsystem** (`internal/faction/engine/hooks/`) absorbed the S-flag and tag-driven mechanics — passive features, event-triggered effects, reactive once-per-turn behavior, and tag-granted modifiers — leaving the ability engine with only the A-flag (player-action-triggered) domain.
2. The **Asset Movement Redesign** (F-002) extracts the entire movement portion of the A-flag domain into a dedicated Movement Phase, with per-asset `Speed` replacing the `MaxHex` step parameter.

What's left of the A-flag domain after movement is extracted is small (10 entries) and structurally simple (dice rolls and opposed checks). The current engine — with its peer-sub-engine status, step abstraction, custom-handler escape hatch, and TOML-step schema — is over-engineered for the residue. Compounding this: **9 of the 10 residual A-flag abilities are not actually engine-handled today**; they fall through to `ConfirmAbilityApplied` for manual GM resolution. The engine is effectively a single-use machine sized to fit Informers.

The redesign question this initiative answers is: *given what's leaving and what remains, what's the right shape for the A-flag domain in code?*

## Audit Findings

### Current ability sub-engine inventory

- Package: `internal/faction/engine/ability/` (82 lines in `ability.go`).
- Two step handlers in `ability/steps/`: `Movement` (carved out by F-002) and `FactionCheck`.
- Step dispatch is a `switch` on `step.Type` (`ability.go:73-80`); only `movement` and `faction_test` are recognized.
- `customHandlers map[string]CustomAbilityHandler` escape hatch (`ability.go:22-24`) — zero callers register anything; confirmed via LSP `findReferences`.
- External coupling: 2 importers. `engine/core.go` (composition) and `action/actions/use_asset_ability.go` (consumer).

### Adjacent subsystem shapes

**Action sub-engine** (`internal/faction/engine/action/`):
- 11 actions registered via factory pattern (`actions/register.go:11-23`); each implements the `Action` interface — `Validate / Inputs / Resolve / Output` (`action/action.go:10-16`).
- No existing action has internal sub-dispatch. `Attack`'s loop is per-target iteration, not handler dispatch.
- `UseAssetAbility` already lives in this package (`actions/use_asset_ability.go`); it currently delegates to `ability.AbilityEngine.Run` and falls through to `ConfirmAbilityApplied` when `def.Ability == nil` (lines 71-76).

**Hooks subsystem** (`internal/faction/engine/hooks/`):
- Typed-category model: `RollModifier`, `RollResultHook`, `MutationReactor`, four cost/rule modifiers (`AssetCostModifier`, `MaintenanceCostModifier`, `WorldTechLevelModifier`, `AssetMovementGranter`), and `TieResolver`.
- Scope-based registration (`hooks.GlobalScope`, `FactionScope`, `AssetScope`).
- All current production handlers are tag-driven (`scavengers.go`, `warlike.go`, `fanatical.go`, `preceptor_archive.go`).
- No movement-event trigger exists; the natural seam for transport-on-movement is a `MutationReactor` observing `MovementOrderProgressed` / `MovementOrderCompleted` mutations.

### A-flag residue (post-movement-carve-out)

After F-002 extracts every movement and self-transport ability, the residual A-flag domain is 10 entries across two mechanic-shapes:

**Economy (6)** — "as an action, roll dice and adjust Coin":

| Asset | TOML ID | Mechanic |
|---|---|---|
| Harvesters | `W1-002` | 1d6; +1 Coin on 3+ |
| Postech Industry | `W3-001` | 1d6; varies (1 lose, 2–4 +1, 5–6 +2) |
| Pretech Manufactory | `W7-001` | 1d8; gain half (rounded up) |
| Pretech Logistics | `F5-002` | Buy a Force asset on TL≤5 |
| Venture Capital | `W6-001` | 1d8; varies (1 destroy, 2–3 +1, 4–7 +2, 8 +3) |
| Commodities Broker | `W6-003` | 1d8; reduce next purchase cost |

**Special action (4)** — "as an action, do a non-standard one-off effect":

| Asset | TOML ID | Mechanic |
|---|---|---|
| Informers | `C1-002` | Cunning vs Cunning → reveal stealthed assets on world |
| Seditionists | `C4-004` | 1d4 Coin cost → attach to enemy asset; target cannot attack while attached |
| Marketers | `W5-001` | Cunning vs Wealth → target faction pays half asset cost or asset is disabled |
| Monopoly | `W4-002` | Force one rival faction with unstealthed assets on world to pay 1 Coin or lose an asset |

**State of implementation today**: 9 of these 10 have *no TOML ability block*; they fall through to `ConfirmAbilityApplied` (manual GM resolution). Only **Informers** is engine-handled (a `faction_test` step with `reveal_stealth` effect at `cunning_assets.toml:41-45`).

### Rulebook flag ambiguity

The SWN rulebook tags each asset with a `Note` column flagging `A` (action), `S` (special feature/cost), `P` (planetary permission), or combinations. The flag column is *internally inconsistent* with the prose for several entries — most notably, Pretech Manufactory is tagged `S` in the table but the prose explicitly says "As an action, the owning faction can roll 1d8…"; Monopoly has the same problem.

The principle we adopted in this discovery to resolve the ambiguity:

> **The prose is authoritative; the table flag is a secondary signal.**
> Rule: *"as an action" → A*, *except when the action moves other assets → that aspect becomes an S-flag hook firing on movement*.

This rule is mechanical and consistently applicable, where the rulebook's own table column is not.

### TOML flag corrections required

Applying the rule to the current TOML data reveals three mismatches that must be corrected as part of R-001:

| Asset | TOML today | Per rule | Change |
|---|---|---|---|
| Pretech Manufactory | `["S"]` | A | flip to `["A"]` |
| Monopoly | `["S"]` | A | flip to `["A"]` |
| Tripwire Cells | `["A", "S"]` | S only | drop A — prose has no "as an action"; purely event-triggered (fires when a stealthed asset lands/is purchased) |

These corrections move Pretech Manufactory and Monopoly *into* the A-flag residue (they're already counted in the table above) and move Tripwire Cells *out* of the A-flag domain entirely (it joins the existing hooks-handled S-flag set).

## Design Summary

The shape that emerges from the audit is a structural simplification, not a redesign of an existing engine. Specifically:

1. **Retire `internal/faction/engine/ability/` as a peer sub-engine.** Its 82 lines and one surviving step type (`faction_test`) do not warrant a separate package.
2. **Fold A-flag dispatch into `Use Asset Ability` inside the action sub-engine.** `actions/use_asset_ability.go` already exists as an action peer; the redesign moves the dispatch logic into that file (or a sibling package under `action/actions/`), with one handler per residual ability.
3. **Retire the `step` abstraction.** Movement is the only step type that ever benefited from the abstraction, and it is leaving. The surviving `faction_test` step can be inlined into the Informers handler (or shared as a helper function — see Q3 below). No new step types are needed.
4. **Retire the `customHandlers` / `CustomAbilityHandler` escape hatch.** Unused in the codebase; deletion has zero callers.
5. **Transport behavior migrates to the hooks subsystem as an S-flag handler firing on movement events.** This is the mechanism for "this asset can carry cargo when it moves itself" — the cargo capability is a property of *how the asset moves*, not a separate action the player invokes.

The exact dispatch shape (asset-ID-keyed handler table vs. shape-keyed mechanic types) and several other shape questions are still open — see *Open Questions* below.

### Cross-cutting impact: reverses Movement Redesign Decision #6

The Movement Redesign plan (`docs/initiatives/implementation/asset-movement-redesign-plan.md:23`) declared in *Decisions Ratified in Planning*:

> **Decision #6:** A distinct `AbilityStepTransport` step type is introduced. `AbilityStepMovement` is deleted entirely once the TOML cutover is complete.

Under R-001's framing, transport never becomes an ability step. It is expressed as an S-flag hook handler that fires on movement, not as a player-chosen action. Movement Redesign Phase 4 (currently paused pending R-001) is re-planned as a hook-handler addition rather than a step-handler addition. Decision #7 (Smugglers and Blockade Runner as "transport-only abilities" with retained Coin cost) is also reframed: their cargo behavior is the S-flag hook; the Coin cost is the hook handler's payload.

The Movement Redesign plan needs a follow-up edit to capture this reversal once R-001's discovery and plan finalize.

## Open Questions

### Q1 — Dispatch shape for the 10 residual abilities ~~OPEN~~ **RESOLVED**

Resolved in-session: **Option A** — asset-ID-keyed handler table. Each residual ability is a small Go handler function; shared patterns (dice → Coin outcome table) are extracted as helpers per Q3. The shape-keyed alternative was rejected because (a) 4 of 10 entries are idiosyncratic one-offs with no shared mechanic, and (b) the action sub-engine has no precedent for typed shape dispatch — Option A matches the existing house style. Migrating to Option B remains reversible if F-007's campaign-tuning direction later justifies it.

---

#### Original options (for reference)

The action sub-engine has no precedent for internal sub-dispatch. Two options on the table:

**Option A — Asset-ID-keyed handler table.** A map keyed by asset definition ID, each value a Go handler function. `Use Asset Ability` looks up the handler by `def.ID` and calls it. Each ability is its own ~20-line function. Common patterns (dice → Coin, opposed roll) are extracted as helper functions called from multiple handlers — pragmatic reuse without forced abstraction.

**Option B — Shape-keyed (typed by mechanic).** A small set of `Shape` types (`DiceForCoin`, `OpposedTest`, `AttachAndDisable`) each with one shared handler. Asset TOML names the shape + parameters; code dispatches on shape type. Data-driven; could lean into the F-007 campaign-scoped TOML tuning direction.

**Sketch — Option A:**

```go
// internal/faction/engine/action/actions/use_ability/
type AbilityHandler func(
    faction *domain.Faction,
    asset *domain.Asset,
    collector action.Collector,
    roller domain.Roller,
    factionState *state.FactionState,
    rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error)

var abilityHandlers = map[string]AbilityHandler{
    "W1-002": harvesters,
    "C1-002": informers,
    // ... 8 more, one per residual ability
}

func harvesters(faction *domain.Faction, asset *domain.Asset, /* ... */) ([]domain.Mutation, error) {
    roll, err := roller.Roll("1d6")
    if err != nil { return nil, err }
    if roll >= 3 {
        return []domain.Mutation{coinGain(faction.ID, 1)}, nil
    }
    return nil, nil
}
```

**Sketch — Option B:**

```go
// internal/faction/engine/action/actions/use_ability/shapes/
type AbilityShape interface {
    Resolve(faction *domain.Faction, asset *domain.Asset, /* ... */) ([]domain.Mutation, error)
}

type DiceForCoin struct {
    Die      string         // "1d6", "1d8"
    Outcomes []DiceOutcome  // [{Range: "3-6", Coin: 1}, ...]
}

type OpposedTest struct {
    AttackerStat string  // "Cunning", "Wealth"
    DefenderStat string
    Effect       string  // "reveal_stealth", "force_payment", ...
}
```

```toml
# Per-asset TOML carries shape + parameters under Option B:
[assets.Harvesters.ability]
shape = "dice_for_coin"
die = "1d6"
outcomes = [
    { range = "1-2", coin = 0 },
    { range = "3-6", coin = 1 },
]
```

**Working recommendation:** Option A. The four special-action entries (Informers, Seditionists, Marketers, Monopoly) are each idiosyncratic enough that shape-based abstraction wouldn't cleanly hold. Shared helper functions give the reuse benefit without an abstraction layer. Matches the existing house style of action handlers. Migrating to Option B later remains possible if F-007 lands.

**Counterpoint:** if campaign-tuning of ability parameters is on a near horizon, Option B starts paying off sooner.

### Q2 — TOML schema fate ~~OPEN~~ **RESOLVED**

Resolved in-session: **retire `[[assets.X.ability.steps]]`, keep a slimmed `[assets.X.ability]` block carrying parameters.** Dispatch is by asset ID (Q1 Option A), so the schema no longer needs to declare mechanic shape — but the static-data principle ([[project_data_location]]) means dice expressions, thresholds, and effect parameters should still live in TOML rather than be hard-coded into Go handlers. Each handler reads its parameters from `def.Ability.*` fields. Concrete shape of the slimmed block (field names, optional/required) is plan-mode work.

---

#### Original options (for reference)

The current TOML schema includes an `[assets.X.ability]` block with `[[assets.X.ability.steps]]` entries (`type`, parameters per step type). Of the 10 residual entries, only Informers populates this block today:

```toml
# cunning_assets.toml:41-45 — the only residual-A-flag asset with an ability block today
[[assets.Informers.ability.steps]]
type = "faction_test"
attacker_stat = "Cunning"
defender_stat = "Cunning"
effect = "reveal_stealth"
```

The 9 other residual entries have no `ability` block at all; they fall through to `ConfirmAbilityApplied` (manual GM resolution) in the engine path.

**Option A:** Retire the schema entirely. Asset-ID alone selects the handler (matches Q1 Option A). The 9 stub TOML entries lose their (currently absent) ability block; Informers' block is removed and the equivalent logic moves to a Go handler.

**Option B:** Keep a simplified schema. Per-ability TOML declares the mechanic key + parameters (matches Q1 Option B's shape-keyed model). GM can tune dice, thresholds, and effect parameters in TOML.

Depends on Q1.

### Q3 — Shared primitives location ~~OPEN~~ **RESOLVED**

Resolved in-session: **Option 2 — extract to a shared helpers subpackage sibling under `action/actions/`** (e.g. `action/actions/use_ability/dice/`). Called by the 4 economy abilities sharing the die+outcome-table sub-shape. The 2 economy outliers (Commodities Broker, Pretech Logistics) and 4 special-action handlers remain bespoke. Placement under `action/actions/` (rather than a broader utility location) follows YAGNI — refactor outward only if hooks or other systems later need the same primitive. Attack's roll machinery is correctly *not* reused: its opposed-defender contract doesn't fit unopposed economy rolls.

---

#### Original options (for reference)

The 6 economy abilities split into three sub-shapes (not a uniform "roll for Coin"):

| Asset | Sub-shape |
|---|---|
| Harvesters, Postech Industry, Pretech Manufactory, Venture Capital | die + outcome table → Coin delta |
| Commodities Broker | roll → state-set (deferred purchase discount) |
| Pretech Logistics | no roll — gated extra purchase action |

The 4 special-action abilities (Informers, Seditionists, Marketers, Monopoly) are each genuinely idiosyncratic — no shared mechanic shape across the set.

Three places primitives could live:

1. **Reuse `Attack` action's roll machinery directly.** Pros: code reuse. Cons: couples ability handlers to Attack's internal contract; Attack's roll path is structured around a specific defender model that doesn't fit unopposed economy rolls.
2. **Extract to a shared helpers package** (e.g. `internal/faction/engine/action/actions/use_ability/dice/` or a more general roll utility). Cleanly separates roll primitives from any one action handler.
3. **Each ability owns its own dice handling.** Each handler calls `roller.Roll(...)` independently. Simplest but most repetition.

**Working recommendation:** Option 2. With 4-of-10 abilities sharing the "die + outcome table" shape, a single `rollOutcomeTable(die, outcomes)` helper meaningfully reduces duplication. The 2 economy outliers (Commodities Broker, Pretech Logistics) and 4 special-action handlers stay bespoke. This recommendation holds regardless of Q1's outcome — under Q1 Option A the helper is called by 4 ability functions; under Q1 Option B the helper *is* the `DiceForCoin` shape's `Resolve`.

### Q4 — Tripwire Cells flag anomaly ~~OPEN~~ **RESOLVED**

Resolved during discovery: drop the `A` flag from Tripwire Cells (`["A", "S"]` → `["S"]`). Tripwire Cells is purely event-triggered (fires on stealthed-asset landing/purchase); its mechanic belongs entirely in the hooks subsystem alongside other S-flag handlers. Listed in *TOML flag corrections required* above.

### Q5 — Stub implementation scope ~~OPEN~~ **RESOLVED**

Resolved in-session: **Option A — structural only.** R-001 ports Informers into the new structure; the other 9 residual A-flag abilities remain stubbed (each is a one-line entry in the handler map that explicitly falls through to `ConfirmAbilityApplied` for manual GM resolution). Implementation of the 9 stubs is added to the backlog (`planned-work.md`) as a follow-up initiative. Reasoning: several of the 9 require non-trivial new domain primitives (e.g. Seditionists' attached-asset mechanic, Marketers' deferred-effect, Monopoly's cross-faction targeting) — those belong in feature initiatives, not in the refactor.

---

#### Original options (for reference)

9 of 10 residual A-flag abilities are stubs today (fall through to `ConfirmAbilityApplied`). Does R-001 include actually implementing them, or just defining the structural shape?

**Option A — Structural only.** R-001 defines the new dispatch shape, migrates Informers' implementation into the new structure, leaves the other 9 stubbed (still routed to `ConfirmAbilityApplied`). Each stub becomes a one-line handler call. Implementation of the 9 is a follow-up initiative.

**Option B — Structural + implementation.** R-001 also implements all 9 stubs. Larger scope; the redesign ships a fully working A-flag domain.

Working recommendation pending Robert's call. Option A keeps R-001 focused on the refactor question; Option B couples a long-standing implementation gap to the structural fix.

### Q6 — Confirmation of deletion scope ~~OPEN~~ **RESOLVED**

Resolved in-session: the retire list below is ratified. Two boundary cases were called out explicitly to prevent confusion in plan mode (see *Reshapes (not deletes)* and *Survives untouched* below the retire list).

**Retired:**

- `internal/faction/engine/ability/` package — deleted entirely
- `ability.AbilityEngine` struct and its `Run` method — deleted
- `ability.CustomAbilityHandler` type and `RegisterCustomHandler` method — deleted (zero callers; safe deletion)
- `ability/steps/` sub-package — deleted (movement step is F-002's territory; `faction_test` logic moves into the new use-ability handler structure per Q3 helpers)
- The `step` abstraction: `domain.AbilityStep` type and the `AbilityStepMovement` / `AbilityStepFactionTest` constants — deleted
- The `[[assets.X.ability.steps]]` TOML sub-schema — deleted (parent `[assets.X.ability]` block survives in slimmed parameter-only form per Q2 Option C)
- `engine.Engine.Ability` field on the core engine struct — deleted
- The `e.Ability` argument threading through `actions/register.go:19-21` — deleted

**Reshapes (not deletes):**

- `domain.Ability` struct — *reshaped*, not deleted. Loses `Steps`; gains parameter fields per Q2 (die expression, thresholds, opposed-stat names, effect key, etc.). Concrete field shape is plan-mode work. Importers of `def.Ability` (`UseAssetAbility` and any reader of `def.Ability.*`) must adjust to the new shape.

**Survives untouched:**

- `ConfirmAbilityApplied` — remains the fall-through path for the 9 currently-stubbed A-flag abilities (per Q5 Option A). Each stub is a one-line entry in the new handler map that routes to `ConfirmAbilityApplied` until the corresponding F-014 work picks it up.

## Replaces / Retires

See *Q6 — Confirmation of deletion scope* above for the full ratified list (retires, reshapes, survives).

## Out of Scope

- **Movement / transport implementation details** — owned by F-002. R-001 contributes the design reversal (transport-as-S-flag-hook) but does not implement the hook handler; that lands in F-002's Phase 4.
- **Tag-granted cross-cutting abilities** (Mercenary Group, Eugenics Cult, Pirates surcharge, etc.) — the hooks subsystem already owns these; R-001 does not touch the tag dispatch path.
- **F-007 campaign-scoped TOML tuning** — informs Q1/Q2 as context but is not in R-001's scope.
- **Implementing the 9 stubbed abilities** — only in scope if Q5 lands on Option B.

## Reference Exemplars

- **F-002 Asset Movement Redesign** — `docs/initiatives/implementation/asset-movement-redesign-plan.md` and its per-effort plans. R-001 reverses Decision #6 from this plan.
- **Sub-engine alignment discovery** — `docs/initiatives/discovery/completed/sub-engine-alignment-discovery.md`. Catalog of sub-engine shapes; relevant to "should the residue stay a sub-engine" (answer: no).
- **Engine-collector reshape discovery** — `docs/initiatives/discovery/completed/engine-collector-reshape-discovery.md`. Recent collector restructure; informs how `UseAssetAbility`'s collector embedding will reshape after the fold-in.
