# Data-Driven Rulebook — Arc Discovery

> **What this doc is.** The cross-cutting decisions that bind *every* initiative in
> the `data-driven-rulebook` arc (A.001). It does **not** slice the arc into
> initiatives — that is Arc-Plan. It ratifies the spine: how a special resolves from
> data, how the surface split shows up in the data, the shared `_core` home for
> goals/tags, and the catalog as binding input.
>
> **Inputs mined (not re-derived):** the [arc concept](data-driven-rulebook-arc-concept.md),
> the [arc-shaped discovery](../../discovery/hawk-rulebook-discovery.md), and the
> [ability & effect catalog](../../../architecture/engine/ability-catalog.md).
>
> **Status:** arc-discovery, written for sign-off. Decisions here are binding on the
> initiatives below.

## Problem

The engine has only ever run one rulebook: SWN reference data. Two things keep it
single-rulebook, and a third is just missing content.

- **Behavior is reached by ID, not by data.** An asset's special ("what it does
  beyond a plain attack") is resolved by looking up its asset ID in hand-written Go.
  That works for the one rulebook it was modeled on, but it does not obviously
  survive a second rulebook. The standing question — *does hardcoded-by-ID hold past
  one rulebook?* — stays unanswered until behavior comes from the data.
- **The shared rules carry SWN flavor.** Goals and tags are *meant* to be reusable
  across rulebooks, but the data still reads SWN-specific, so they can't be shared
  without dragging SWN fiction into a new game.
- **There is no real campaign.** H.A.W.K — the game this tool exists to run — exists
  only as intent. No assets, no canon.

This arc makes behavior come from data across both of the engine's dispatch
surfaces, promotes goals and tags to a genuinely shared core, and ships H.A.W.K as
the first non-reference rulebook — the precondition for any rulebook after it.

## Architecture Context

Pages read for re-grounding (their `Key Decisions` are binding context):
[architecture-overview](../../../architecture/architecture-overview.md),
[persistence](../../../architecture/engine/persistence.md),
[actions](../../../architecture/engine/actions.md),
[effect-mutation](../../../architecture/engine/effect-mutation.md),
[hooks](../../../architecture/engine/hooks.md), and the
[ability & effect catalog](../../../architecture/engine/ability-catalog.md).

Load-bearing claims verified directly against source this session (docs had drifted
in two places — noted under Reconciliation):

| Claim | Verified in source |
|---|---|
| A-side dispatch is keyed on asset **ID** | `dispatch.go` — `map[string]AbilityHandler` keyed on `def.ID`: one real handler (`C1-002`→`informers`) plus nine `confirmApplied` stubs; unknown IDs fall through to `confirmApplied` |
| S-side dispatch is **data-driven** | `core.go:58` — `for _, def := range rulebook.Assets { if def.Transport != nil { … } }`; one reactor registered per asset that declares a transport profile |
| The ability TOML record is **flat** | `rulebook.go` — `abilityRecord` has exactly three fields (`attacker_stat`, `defender_stat`, `effect`). The domain's `Die`/`Outcomes` fields have **no parser path at all** — they cannot be filled from TOML today |
| The rulebook copy step reads **one** source folder | `scaffold.go` — `CopyRulebook(c, srcDir, replace)` walks a single `srcDir`; `Scaffold` builds the campaign tree by reflecting over `Paths` struct tags |
| The canonical hook numbering | `hooks/doc.go` — Cat 1 `RollModifier`, Cat 2 `RollResultHook`, Cat 3 `MutationReactor`, **Cat 4 `RuleModifier` family** (`AssetCostModifier`, `MaintenanceCostModifier`, `WorldTechLevelModifier`, `AssetMovementGranter`), Cat 5 `TieResolver` |
| A movement-grant hook **exists but is unwired** | `hooks/rule_modifier.go`, `registry.go`, `dispatch/rules.go` define `AssetMovementGranter` and `GrantedMovementAbilities`, but nothing calls `GrantedMovementAbilities` — the movement phase does not dispatch it yet. `MovementAbility` is a `{Name, Description}` placeholder, not an executable move |

The binding decisions these pull in: **GMs own all data, the binary embeds none**
(persistence) — so any shared core must end up as ordinary TOML in the campaign, not
a runtime read from a second location; **the campaign is the addressing root,
rulebooks are copied in** (persistence) — so "share goals/tags" resolves at the
source-library level, not at load time; **the rule layer registers hooks, it never
dispatches** (effect-mutation) — the data-driven pattern the S-side already follows;
**dispatch is distributed** (hooks) — each hook category is consulted at the one
pipeline point where its question is asked.

## The cross-cutting decisions

These bind every initiative. Tagged **[carried-in]** (ratified before this session,
not re-litigated) or **[ratified here]** (settled or refined this session).

**ACD-1 — The flag picks the harness; the mechanic picks the engine. [ratified here]**
The SWN `A`/`S`/`P` flag is a *human-facing* tag: it says whether the GM spends the
faction's action, written for a person to read at the table. It is **not** a reliable
programmatic dispatch key, and for some assets it points at the wrong home. We
re-derive the programmatic home from what the special actually *does*. The flag still
matters — it chooses the **harness** (action-spent vs. free/passive) — but the
**mechanic** chooses which engine and which hook category hosts it. *(Refines the
carried-in "A→abilities engine, S→effects engine" — that holds as a default, not a
law.)*

**ACD-2 — Build the full data-driven registry now, both surfaces (R.004). [carried-in]**
Behavior comes from rulebook data on both dispatch surfaces, not bespoke-only and not
deferred to a second rulebook. **R.018** — re-keying A-side dispatch on the ability's
effect/primitive instead of its asset ID — is the prerequisite increment.

**ACD-3 — The registry shape, both engines. [ratified here]**
- **A-side (action-spent abilities):** a composition pipeline —
  *target-selection → (stat-contest | die-roll + outcome-table) → effect → mutations.*
  Dispatch is keyed on the effect/primitive, not the asset ID.
- **S-side (passive/reactive features):** a profile in the data registers the right
  hook from the [five canonical categories](../../../architecture/engine/hooks.md),
  exactly the way `transport` registers a Cat 3 reactor today. The mechanic decides
  the category (ACD-1).

**ACD-4 — One effect that appears on both surfaces is one primitive, many harnesses. [ratified here]**
When the same mechanic shows up action-spent on one asset and free/passive on
another, the *logic is written once* and reached through different harnesses. The
worked example is **`relocate-other`** (see below): one move primitive in the world
engine, reached by an action-spent harness (abilities engine) and a free harness (a
Cat 4 movement-granter consulted in the movement phase). Surface = harness, not a
re-implementation.

**ACD-5 — The surface is visible and self-describing in the data. [ratified here]**
A reader of an asset's TOML can see which engine owns its special without reading
code, and the block names its own primitive so dispatch is a data lookup.
- **A-side:** an `[ability]` block (effect, optional `[contest]`, optional
  `[roll]`/`outcomes`).
- **S-side: hybrid.** Typed profiles for genuinely parametric families (keep
  `[transport]`; add typed blocks where the parameters are rich), and a generic
  `[effect]` block (a `type` discriminator plus a few fields) for the simple
  flag-effects. The shape matches each effect's actual complexity.
- The flat `abilityRecord` grows to carry the structured blocks; the dead
  `Die`/`Outcomes` and the dropped `[[ability.steps]]` get resolved here (filled in
  as the real shape or removed), and `coin_steal`/`coin_drain` get their handlers.

**ACD-6 — Goals and tags share one source home, deduped at scaffold. [carried-in, refined]**
Every rulebook composes the *same* goals and tags. The home is a repo source folder,
`rulebooks/_core/` (goals + tags), that the scaffold step composes into each
campaign at creation. **Refinement from source:** the copy step reads one folder
today, so this is a real change to the `campaigns`/scaffold layer (compose from two
sources), **not** a loader change and **not** a new `Paths` field — the per-campaign
on-disk result is identical to today, so `rulebook.Load` is untouched. The campaign
stays self-contained (honoring "GMs own all data"); only the repo stops duplicating.

**ACD-7 — Goals/tags neutralization is data-only. [carried-in]**
Stripping SWN flavor changes the *data*, not the mechanics: bare-ID handlers keep
firing, no mechanic is added, changed, or removed. Because the same files now compose
into both rulebooks (ACD-6), a neutralized tag has to read neutral for *both* games
at once.

**ACD-8 — The catalog is binding mechanical input. [carried-in]**
The [ability & effect catalog](../../../architecture/engine/ability-catalog.md),
organized by primitive, is the decomposition every registry initiative builds
against; the arc does not re-derive it. **Anchor all hook-category references on the
canonical `hooks/doc.go` numbering** (Cat 4 = RuleModifier family, Cat 5 =
TieResolver); the catalog's own category labels drifted and get corrected (see
Reconciliation).

**ACD-9 — The H.A.W.K lore canon is a living doc. [carried-in]**
Campaign canon at `docs/campaigns/hawk/`, seeded by a precursor lore session, grown
through asset authoring. Its mechanics are per-initiative.

## How a special resolves — the surface model

The arc's core move is an inversion. **Today code is the index:** a `map[def.ID]`
handler list on the A-side, a hardcoded `if def.Transport != nil` on the S-side.
**After the arc, data is the index:** the TOML block names its primitive, and
dispatch is a lookup.

The flag and the mechanic do different jobs (ACD-1):

- The **flag** answers *does the GM spend the faction's one action?* — action-spent
  (`A`) vs. free/passive (`S`). That picks the **harness**.
- The **mechanic** answers *what does it do?* — relocate an asset, drain Coin,
  reveal stealth. That picks the **engine and hook category**.

Two surfaces, two shapes:

- **A-side, action-spent:** `Use Asset Ability` runs the composition pipeline
  (ACD-3) for the chosen asset and folds the resulting mutations into the action.
- **S-side, passive/reactive:** at cycle start the effect engine reads the asset's
  profile and registers the matching hook. Which category depends on the mechanic:
  an *event-reactor* (fires on an attack or a death) is Cat 3; a *rule/permission
  grant* (start stealthed, may make a free move) is Cat 4; a *tie-break* is Cat 5.

### Worked example — `relocate-other`

Three assets carry the same mechanic, differing only in harness and eligibility:

| Asset | Flag | Cost-trigger | Eligible targets | Range |
|---|---|---|---|---|
| Covert Transit Net `C6-002` | `A` | "As an action" | any Special Forces | ≤ 3 hex |
| Covert Shipping `C3-003` | `A` | 1 Coin, action (per flag) | one Special Forces | ≤ 3 hex |
| Transit Web `W7-003` | `S` | 1 Coin, "does not require an action" | any non-Starship Cunning/Wealth | ≤ 3 hex |

All three are the **same move**: an *instant* inter-world relocation, flat Coin,
range-gated by hex distance — *not* a multi-turn drift order, and *not* an
event-reactor. The catalog filed Transit Web as a Cat 3 reactor by taking its `S`
flag at face value; re-deriving from the mechanic (ACD-1) routes it elsewhere.

Resolution under this arc:

- **The primitive** (the move itself — check range, charge the Coin, place the asset
  at the destination now) lives **once**, in the world engine, as a new kind of move
  distinct from the drift-order path.
- **The action-spent harness (A-side):** `Use Asset Ability` invokes the primitive;
  it costs the faction's action.
- **The free harness (S-side):** the effect engine registers an **`AssetMovementGranter`**
  (the Cat 4 RuleModifier-family member built for exactly this) from the asset's
  profile. The **movement phase** consults it and offers the GM the free relocate,
  which invokes the same primitive without spending the action.

What the GM sees: for Transit Web, a menu item *during the movement phase* —
"relocate an eligible asset within 3 hexes for 1 Coin" — that leaves the turn's
action intact. For Covert Transit Net, an entry in the *action menu* under "Use Asset
Ability" — same move, but it is the turn's action.

**The framework already has the seam, half-built.** `AssetMovementGranter`,
`GrantedMovementAbilities`, and `MovementAbility` exist in `hooks/`, but
`GrantedMovementAbilities` has no caller — the movement phase doesn't dispatch it
yet — and `MovementAbility` is a `{Name, Description}` placeholder. So the S-side work
is concrete: register the granter from the profile, **wire `GrantedMovementAbilities`
into the movement phase**, and grow `MovementAbility` to carry (or reference) the
relocate parameters.

## The data contract

ACD-5 made concrete. The block a GM writes names the engine and the primitive.

**A-side — `[ability]` (action-spent):**

```toml
[ability]
effect = "reveal_stealth"      # the primitive
[ability.contest]              # optional: stat-contest before the effect
attacker_stat = "cunning"
defender_stat = "cunning"
# or, instead of a contest:
[ability.roll]                 # optional: die-roll + outcome-table
die = "1d8"
[[ability.outcomes]]
from = 1
to = 4
coin = 1
```

**S-side — hybrid:**

```toml
# rich, parametric family -> typed profile (the transport pattern)
[transport]
max_hex = 3
coin_cost = 1
cargo_types = ["Force"]

# simple flag-effect -> generic block, type names the primitive
[effect]
type = "coin-steal-on-attack"
amount = 1
once_per_turn = true
```

The parser record (`abilityRecord`, the transport/effect records) grows to carry
these structured blocks. The exact field-by-field shape and the fate of each dead
field (`Die`/`Outcomes`/`steps`) are per-initiative work for the A- and S-registry
initiatives; the binding decision here is *the shape is structured, self-describing,
and visible in the data.*

## The shared `_core` home

ACD-6, scoped. Today `rulebooks/swn/` holds goals, tags, assets, drift, and spatial,
and `CopyRulebook` copies that one folder into each campaign. Under the arc:

- `rulebooks/_core/` holds `goals.toml` + `tags.toml` — the shared mechanics.
- `rulebooks/swn/` and `rulebooks/hawk/` keep only what is genuinely
  rulebook-specific (assets, drift costs, spatial), and **lose** their own
  goals/tags.
- The scaffold step composes **core + the chosen rulebook** into the campaign's
  `rulebook/` dir, so the on-disk result is unchanged and `rulebook.Load` never
  learns `_core` exists.

What is *settled* here: the home, the source-library-dedupe mechanism, "no loader
change / no `Paths` field." What is *deferred to the goals/tags + `_core` initiative*:
the flavor-vs-mechanic line for neutralization (how far a rename goes when the file
must read neutral for both games), and the precise compose mechanics (two
`CopyRulebook` calls vs. a compose step).

## The other deliverables — scoped, routed per-initiative

Three of the four arc deliverables are mostly *downstream* of the spine. The
arc-discovery fixes their binding decisions (above) and routes the rest:

- **H.A.W.K asset authoring** — the actual content. Runs as iterative by-category
  passes, each with a lore-gate, reflavor, and a mechanical sort (reuse a primitive /
  build a missing primitive / parameterize a near-miss) against the catalog. All
  per-initiative; **routed to the asset-authoring initiative.**
- **Goals/tags neutralization + `_core` home** — the data-only neutralization and
  the scaffold compose. Binding decisions in ACD-6/ACD-7; the flavor line and compose
  mechanics are **routed to that initiative.**
- **Lore canon** — living doc, seeded by a precursor lore session. ACD-9 fixes its
  nature; the session flow and lore-gate mechanics are **routed to the lore session /
  asset-authoring initiative.**

## Per-initiative seed — for Arc-Plan to route

> This is **not** arc-spine. It is concrete detail the per-initiative discoveries
> will consume, lifted here from the original discovery doc so that doc can retire.
> Arc-Plan assigns each block to the initiative that owns it; the spec work happens
> in that initiative's own discovery.

**SWN re-prefix migration** *(→ the SWN-data / re-prefix initiative; sequence after
R.018).* Prefixing existing SWN asset IDs (`C1-001` → `SWN-C1-001`) is a data + code
change, not just HAWK authoring:

- Rename the IDs across `rulebooks/swn/assets/`.
- Chase the hardcoded references: `B.001`'s `stealth_applicator` and the `G.012`
  goal ref.
- Fix the test fixtures that pin old IDs.
- Sequence **after R.018** — once A-dispatch is keyed on the effect, the `dispatch.go`
  ID entries fall away, so the re-prefix no longer touches dispatch code.
- Open: whether any already-scaffolded campaign carries old IDs needing migration
  (likely none this early).

**Asset-authoring pass loop** *(→ the asset-authoring initiative).* The unit of work
is one category file, worked Force → Cunning → Wealth. Each pass:

1. **Lore-gate** — confirm the batch's creative direction against the canon before
   authoring.
2. **Reflavor** — name, description, and fiction per asset, co-authored from the
   canon; new fiction folds back into the canon doc.
3. **Mechanical sort** — classify each asset's ability against the catalog: reuse an
   implemented primitive (pure data) / build a missing primitive / parameterize a
   near-miss.
4. **Review** before commit.

Asset IDs gain the `HAWK-` prefix, initially mirroring SWN's tier suffix
(`HAWK-F1-001`) and drifting as the set diverges. Output per pass:
`rulebooks/hawk/assets/<category>_assets.toml`, any registry primitives the sort
required, and canon-doc updates.

**Lore session & gate** *(→ the lore session / asset-authoring initiative).* A
precursor lore session seeds the canon: Robert is interviewed about H.A.W.K's setting,
factions, terminology, and the fiction behind its elements. Then it grows — each
authoring pass's lore-gate confirms the batch against the canon, and the asset-level
fiction that reflavoring produces folds back in. Seeded up front so authoring isn't
inventing canon cold; grown by authoring so it stays complete. The canon lives at
`docs/campaigns/hawk/`.

## Open Questions — routed

| Question | Goes to |
|---|---|
| How to slice the arc into initiatives, and in what order | Arc-Plan |
| Whether any already-scaffolded campaign needs SWN-ID re-prefix migration | Arc-Plan / the SWN-data initiative |
| Monopoly's flag-vs-text mismatch (flagged `A`, reads reactive) — re-derive its home from the mechanic, the way relocate-other was | A-registry initiative |
| The exact `[ability]` sub-block nesting and the fate of each dead field | A- and S-registry initiatives |
| Neutralization's flavor-vs-mechanic line, and the `_core` compose mechanics | Goals/tags + `_core` initiative |
| Precursor lore session — when it runs, how it seeds the canon | Lore session / asset-authoring initiative |
| Custom session types (lore, ability-catalog) + the `goals.md` "intentional no-op" wording | Deferred, process-level |

## Out of scope (arc-level)

- **`B.002` (semantic table keys).** Goals/tags keep bare mechanic IDs, so their
  handlers keep firing and `B.002` is not triggered.
- **Goals authoring / tag mechanic changes.** Goals copy verbatim; the tag pass is
  flavor-neutralization only.
- **Other rulebooks (WWN).** H.A.W.K is the first non-reference rulebook; WWN is the
  next beneficiary, not touched here.
- **`spatial/` and `drift_costs.toml`.** Carried as-is from the SWN shape.

## Reconciliation items surfaced this session

Durable doc/code drift found during re-grounding, to fix as the relevant initiatives
land (not blocking this doc):

- **Catalog hook-category labels drifted.** It labels RuleModifier-family effects
  "Cat 5" and TieResolver also "Cat 5"; canonical `hooks/doc.go` is Cat 4 =
  RuleModifier family, Cat 5 = TieResolver. Renumber the catalog to match.
- **Catalog mis-files `relocate-other` (free).** Listed as "Cat 3 `MutationReactor`";
  it is a **Cat 4 `AssetMovementGranter`** consulted in the movement phase.
- **`hooks.md` overstates the movement seam.** It says `GrantedMovementAbilities` is
  "called from the movement phase"; it is defined but unconsumed. Correct the page (or
  wire the call) when the S-side relocate lands.
- **`goals.md` "intentional flavor" wording.** The data-only skip is an incremental
  seam, not a deliberate end state; reconcile the wording (already flagged).
