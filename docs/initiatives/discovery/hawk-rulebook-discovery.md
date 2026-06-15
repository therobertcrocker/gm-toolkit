# H.A.W.K Rulebook (F.022) — Arc Discovery

> **Status: elevated to an arc.** The ability-catalog pre-plan session settled `R.004` — build the data-driven ability/effect registry in full, across *both* dispatch surfaces — which makes this an arc. It delivers four things:
>
> 1. **All `A`- and `S`-flag specials, data-driven**, with a crisp surface split: `A` → abilities engine, `S` → effects engine. (`R.018` — effect-keyed A-dispatch — is the prerequisite increment.)
> 2. **A full H.A.W.K rulebook** — assets fully defined and functional.
> 3. **Goals + tags made campaign-agnostic** via data-only neutralization (bare-ID handlers unchanged), plus a **shared source home** so every ruleset composes the same goals/tags.
> 4. **Durable docs:** the [ability & effect catalog](../../architecture/engine/ability-catalog.md) (done) and a **H.A.W.K lore canon** — seeded by a precursor lore session, then grown through asset authoring.
>
> Catalog-session decisions are folded into the sections below; the arc-plan slices these into efforts.

## Problem

The engine has only ever run against SWN reference data: one rulebook's goals, tags, and assets, with behavior hardcoded by ID and validated against the source material it was modeled on. There is no real campaign content. Robert's actual campaign — H.A.W.K, a standalone rulebook — exists only as intent, so the tool cannot yet run the game it was built to serve.

Two problems motivate the arc. First, the standing question — *does hardcoded-by-ID behavior hold past a single rulebook?* — is answered by going **data-driven**: the arc replaces per-ID dispatch with a primitive registry across both surfaces. Second, the goals/tags data carries latent SWN-specific flavor, so what *should* be campaign-agnostic isn't yet: it can't be reused across rulebooks without dragging SWN fiction along.

F.022 builds the data-driven registry that resolves every `A`- and `S`-flag special from rulebook data (`R.004`, the arc's mechanical spine), authors H.A.W.K's assets (the substantive content work), neutralizes goals/tags into a shareable core with a shared source home, and establishes a lore canon (seeded up front, grown through authoring).

## Architecture Context

Pages read: [`persistence.md`](../../architecture/engine/persistence.md) (the binding page — rulebook data, ownership, loading), [`goals.md`](../../architecture/engine/goals.md), [`actions.md`](../../architecture/engine/actions.md). Source verified directly for the dispatch surfaces (`dispatch.go`, `effect/effect.go`, `core.go`, the goal/tag registries) — the dispatch story is split in a way the arch pages don't fully spell out.

Binding Key Decisions and how each constrains F.022:

- **GMs own all data; the binary embeds none.** Every asset/tag/goal is GM-authored TOML under a campaign root; there are no compiled-in defaults. This is the constraint the "shared home for goals/tags" idea has to satisfy — a runtime core the loader reads from a second, non-campaign location would violate it.
- **The campaign is the addressing root; rulebooks are copied in.** `rulebooks/swn/` is a *source template*; `CopyRulebook` copies its TOML into each campaign's own `rulebook/` dir (`Paths.FactionDataDir`), and `rulebook.Load(dataDir)` reads goals/tags/assets from that one per-campaign directory. There is no runtime composition seam. So "don't copy goals/tags" resolves cleanly only at the **source-library** level (dedupe the repo's template files, compose at scaffold time), not at load time.
- **`Paths` is tag-derived.** Any new shared directory is a single struct field reflected over by both `Paths()` and `Scaffold()` — cheap to add if a shared-core source path is wanted.
- **Asset files glob and merge; rules validate at load.** Every `assets/*_assets.toml` merges into one map (duplicate IDs rejected), so HAWK assets can be split into themed files. An unknown `ability effect` is a load-time error — `Ability.Effect` is already a validated enum, and it is the field `R.018` keys dispatch on (replacing `def.ID`).

### Dispatch surfaces (verified from source)

A "mechanic" reaches code through one of four surfaces, with different reuse and prefix properties — this is the spine of the registry the arc builds and the surface split it must preserve:

| Surface | Registration | Keyed on | HAWK reuse cost |
|---|---|---|---|
| **Goals** | hardcoded handler list (`goal.go:29`) | bare mechanic ID | reused mechanic fires if HAWK keeps the ID; new mechanic = one handler |
| **Tags** | hardcoded handler list (`tag.go:24`) | bare mechanic ID | same as goals |
| **A-flag abilities** | hardcoded `map[def.ID]` (`dispatch.go:21`, ten entries: one real `C1-002`→`informers` plus nine `confirmApplied` stubs), via Use Asset Ability action | `def.ID` | reuse needs a per-ID entry; **the only surface the asset prefix breaks** — the `R.018` target |
| **S-flag effects** | data-driven off the rulebook (`core.go:58`, `def.Transport != nil`), one reactor per matching def | `def.ID` (per-def instance, not behavior-bound) | reused mechanic is **zero-code and prefix-immune**; e.g. Smugglers `C1-001` `[transport]` |

Two consequences for F.022:

- **R.004 spans both surfaces.** The arc builds data-driven resolution for the A-flag abilities (the ID-bound surface `R.018` fixes) *and* the S-flag effects (where only `transport` is wired today; the rest are unimplemented). Goals/tags stay on bare-ID dispatch — outside R.004, but in scope for data neutralization + the shared home.
- **The silent-skip (handler-less ID → no-op) is an incremental-implementation seam**, not a preferred end state: an entry with no handler means the mechanic is *not yet coded*, not that it is deliberately code-free. (`goals.md` currently frames data-only as "intentional flavor"; flagged for reconciliation at pre-merge.)

## Design Summary

F.022 is an arc with four deliverables (see Status). Its mechanical spine is the **data-driven ability/effect registry** (`R.004`) across both surfaces; its substantive content work is **authoring H.A.W.K's assets**; and it **promotes goals + tags to a campaign-agnostic core** with a shared source home, and establishes a **lore canon** (seeded by a precursor session, grown through authoring). The split follows the data's nature — assets are campaign-specific content; goals and tags are shared mechanics that only *look* SWN-specific; the registry is engine machinery both rulebooks reuse.

**Assets — the meat.** HAWK's asset set starts from SWN's, reviewed and reflavored, then drifts as the campaign's picture sharpens. Asset IDs carry a **campaign prefix** (`C1-001` → `SWN-C1-001`; HAWK assets `HAWK-…`) — provenance, not collision-avoidance (one rulebook ever loads at a time). Today the prefix touches the A-flag ability map (the `C1-002` entry and any reuse of it); once `R.018` makes dispatch effect-keyed, the prefix touches no dispatch code at all. Authoring runs as **iterative by-category passes** (Force → Cunning → Wealth, one `*_assets.toml` per pass), each asset sorted into one of three rows:

- **Row 1** — same mechanic, reused → pure data, no code (the primitive is already implemented in the registry)
- **Row 2** — a primitive the registry doesn't have yet → build the primitive (data-driven), not a one-off handler
- **Row 3** — a *parametric near-miss*: "like an existing A-type or S-type ability, but with one thing different" (a knob, a threshold, a target type) — neither clean reuse nor a genuinely fresh mechanic, and the clearest signal that a parameterized, data-defined handler is the right shape.

**Mechanics: catalog-first, decision made.** With ~55 of ~57 A/S specials unimplemented (only transport and reveal-stealth are coded), implementing them bespoke just to tally drift would be throwaway work. So the catalog came first: every A/S special surveyed and decomposed into candidate primitives, organized by primitive, living durably at [`ability-catalog.md`](../../architecture/engine/ability-catalog.md). It settled **`R.004`**: *build the full data-driven registry now, across both surfaces* — not bespoke-only, not deferred to a second rulebook. That decision is what makes this an arc. The two coded mechanics anchor the target model: **transport** is already behavior-data-driven (the S-side target shape); **reveal-stealth** is dispatch-only bespoke (the A-side shape `R.018` replaces) — one example at each level.

With `R.004` settled, the three rows now describe authoring-time classification against the *registry*: Row 1 reuses an implemented primitive (pure data), Row 2 needs a primitive not yet built, Row 3 is a parametric near-miss the registry parameterizes. `R.018` (effect-keyed A-dispatch, "R.004 increment 1") is the prerequisite increment and is in scope.

**Goals + tags — the agnostic core.** Goals are already mechanically agnostic (copied verbatim, bare IDs). Tags need a **neutralization pass** — strip SWN-specific flavor so they're genuinely reusable — which is what *earns* them agnostic status rather than a cosmetic edit. Both keep **bare mechanic IDs** (no prefix), so their hardcoded handlers keep firing across rulebooks. This is naturally a distinct effort from asset authoring.

*Shared home — decided.* Every ruleset composes the **same** goals and tags into its campaign dir. The resolution is **source-library dedupe**: a shared `rulebooks/_core/` (goals + tags) that `Scaffold`/`CopyRulebook` composes into each campaign at creation — the campaign stays self-contained (honoring "GMs own all data"), only the repo stops duplicating. This is a `campaigns`/scaffold change, not a loader change (the per-campaign on-disk result is unchanged). A distinct arc effort.

**Shape — an arc.** This arc-discovery, plus a **precursor lore session** (banking H.A.W.K's setting, factions, and terminology into the canon before authoring begins), feeds an arc-plan that slices the execution efforts. Candidate efforts: **R.018** (effect-keyed A-dispatch), the **A-flag ability registry**, the **S-flag effect registry**, **HAWK asset authoring**, and **goals/tags neutralization + shared-core home**. Ordering is an arc-plan call (R.018 precedes the A-registry). Each authoring pass opens with a **lore-gate** — a checkpoint confirming the batch's creative direction against the canon — then draws on the canon and folds new asset-level fiction back into it (`docs/campaigns/hawk/`).

## Decisions (catalog session)

- **`R.004` — build the full data-driven registry now, across both surfaces.** Not bespoke-only, not deferred to a second rulebook. This is the decision that makes F.022 an arc. `R.018` (effect-keyed A-dispatch) is the prerequisite increment, in scope.
- **Behavior catalog — durable, by primitive.** Lives at [`docs/architecture/engine/ability-catalog.md`](../../architecture/engine/ability-catalog.md), organized by candidate primitive, covering both surfaces with the A/S split as a load-bearing distinction (flag picks the engine, effect picks the primitive).
- **Shared goals/tags home — source-library dedupe.** A shared `rulebooks/_core/` composed at scaffold so every ruleset gets the same goals/tags; a `campaigns`/scaffold change, not a loader change.
- **Goals/tags neutralization — data-only.** Bare-ID handlers unchanged; only the data loses SWN flavor. No mechanic added, changed, or removed.
- **Lore canon — seeded up front, grown through authoring.** A precursor lore session banks the foundational canon (setting, factions, terminology); each authoring pass then draws on it and feeds asset-level fiction back. A living doc, not a one-shot.

## Open Questions

- **Effort ordering and slicing** → *arc-plan.* The arc's efforts (R.018, A-registry, S-registry, asset authoring, goals/tags + shared home, lore canon) and their dependencies/ordering.
- **SWN re-prefix migration reach** → *arc-plan.* Whether any already-scaffolded campaign carries old asset IDs needing migration (likely none this early).
- **Monopoly flag-vs-text mismatch** → *during the A-registry effort.* Flagged `A` but written reactively; resolve when its primitive is built.
- **Custom session types in the process** → *deferred, Fable-level.* Whether the lore and ability-catalog sessions warrant a `session-modes.md` note; and the `goals.md` "intentional no-op" wording reconciliation flagged in Architecture Context.

---

## Per-Area Design Details

### Pre-Plan sessions

Two pre-plan sessions feed the arc-plan. The **ability-catalog session** ran first (output: [`ability-catalog.md`](../../architecture/engine/ability-catalog.md)) and settled `R.004`, elevating this to an arc. A **precursor lore session** (still to run) banks H.A.W.K's foundational canon before asset authoring begins. They're independent — the catalog is mechanical (SWN-derived), the lore is creative (HAWK). Flow: **Discovery → {ability-catalog ✓, lore session} → arc-plan → execution**.

### Lore session & canon

The **H.A.W.K lore canon** is a *living* durable reference doc at `docs/campaigns/hawk/` — campaign canon, not an archived planning artifact. A **precursor lore session** seeds it: Robert is interviewed about H.A.W.K's setting, factions, terminology, and the fiction behind its elements, establishing the foundation authoring draws on. Then it grows: each authoring pass's **lore-gate** confirms the batch against the canon, and the asset-level fiction that reflavoring produces is folded back in. Seeded up front so authoring isn't inventing canon cold; grown by authoring so it stays complete.

### Ability-catalog session (done)

Surveyed all `A`- and `S`-flag specials (from SWN flags, descriptions, and existing profiles like `[transport]`) and decomposed each into **candidate primitives**, organized by primitive, at [`ability-catalog.md`](../../architecture/engine/ability-catalog.md). Anchored by the two coded exemplars: transport (behavior-data-driven) and reveal-stealth (dispatch-only). Output: the **`R.004` call** — build the full data-driven registry now, both surfaces — which made this an arc. The surface split is load-bearing: `A` → abilities engine, `S` → effects engine; the same effect can appear on either surface (relocate-other is both), so the flag picks the engine and the effect picks the primitive.

### Asset-authoring pass loop

The unit of work is **one category file**, worked Force → Cunning → Wealth. Each pass:

1. **Lore-gate** — a checkpoint confirming this batch's creative direction against the canon before authoring (the threaded "we need to talk about the campaign first" gate).
2. **Reflavor** — name, description, and fiction for each asset, co-authored from the canon; new asset-level fiction is folded back into the canon doc.
3. **Mechanical sort** — each asset's ability classified Row 1 / 2 / 3 against the catalog; Row 2 work builds the missing primitive in the registry (per the `R.004` decision).
4. **Review** before commit.

Asset IDs gain the `HAWK-` prefix, initially mirroring SWN's tier suffix (`HAWK-F1-001`) and drifting as the asset set diverges. Output per pass: `rulebooks/hawk/assets/<category>_assets.toml`, any registry primitives the sort requires, and canon-doc updates.

### Tags → agnostic core, and the shared home

Goals are already mechanically agnostic and copy verbatim (bare IDs). Tags need a **neutralization pass** — strip SWN-specific flavor from the four coded tags and the data-only tags so they read as campaign-neutral mechanics, keeping bare mechanic IDs so their handlers keep firing across rulebooks (data-only; no mechanic touched). The shared home is **decided**: a shared `rulebooks/_core/` (goals + tags) that `Scaffold`/`CopyRulebook` composes into each campaign at creation, so every ruleset gets the same goals/tags while the live campaign stays self-contained. A `campaigns`/scaffold change, not a loader change.

## State Storage

**Source rulebook layout.** F.022 creates `rulebooks/hawk/` as the second source-template rulebook (the first being `rulebooks/swn/`). Per `persistence.md`, a source rulebook is what `CopyRulebook` copies into each campaign's own `rulebook/` dir; the live rulebook is per-campaign. The HAWK source holds:

- `assets/*_assets.toml` — **authored** (the meat), globbed and merged at load, `HAWK-`-prefixed IDs.
- `goals.toml`, `tags.toml` — the **agnostic core**: composed from `rulebooks/_core/` (the decided shared home) at scaffold time rather than stored per-rulebook.
- `drift_costs.toml`, `spatial/` — as SWN today (out of scope to change).

**Shared core (decided).** `rulebooks/_core/goals.toml` + `_core/tags.toml` become the single home for the agnostic mechanics, and `Scaffold`/`CopyRulebook` compose core + campaign-specific into the campaign's `rulebook/`. This is a `campaigns` package change (a new `Paths`/scaffold source), not a loader change — the per-campaign on-disk result is unchanged, so `rulebook.Load` is untouched.

**SWN re-prefix is a migration.** Prefixing existing SWN asset IDs (`C1-001` → `SWN-C1-001`) is a data + code change, not just HAWK authoring: rename IDs across `rulebooks/swn/assets/`, chase any hardcoded references (`B.001`'s `stealth_applicator`, the `G.012` goal ref) and test fixtures. Once `R.018` makes A-dispatch effect-keyed, the `dispatch.go` ID entries fall away, so the re-prefix no longer touches dispatch — sequence it after R.018. Whether any already-scaffolded campaign needs migrating is an open question (likely none this early).

**Canon doc** lives at `docs/campaigns/hawk/` — deliverable #4, a living documentation artifact seeded by the precursor lore session and grown through asset authoring (not game data); noted here only because it is durable and campaign-scoped.

## Out of Scope

- **`B.002` (semantic table keys).** The assets-only prefix leaves goals/tags on bare mechanic IDs, so their handlers keep firing and `B.002` is not triggered.
- **Goals authoring / tag mechanic changes.** Goals copy verbatim; the tag pass is flavor-neutralization only — no mechanic added, changed, or removed.
- **Other rulebooks (WWN).** F.022 is the first non-reference rulebook; WWN is the next beneficiary of any data-driven work but is not touched here.
- **`spatial/` and `drift_costs.toml`.** Carried as-is from the SWN shape.
