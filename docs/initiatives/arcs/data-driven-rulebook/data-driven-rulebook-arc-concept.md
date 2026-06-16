# Data-Driven Rulebook — Arc Concept

## Vision

The engine runs a real, second campaign — H.A.W.K — end to end, and in doing so stops being a single-rulebook tool. Every `A`- and `S`-flag special resolves from rulebook data through a primitive registry rather than per-ID dispatch; goals and tags are genuinely campaign-agnostic, composed from one shared source home so every rulebook inherits the same mechanics; and H.A.W.K ships as a full, playable rulebook backed by a living lore canon. When the arc completes, the tool finally runs the game it was built to serve, and the engine has demonstrably held a *non-reference* rulebook — the direct precondition for Worlds Without Number and any rulebook after it.

## Problem

The engine has only ever run against SWN reference data: one rulebook's goals, tags, and assets, with behavior hardcoded by ID and validated against the source material it was modeled on. There is no real campaign content — H.A.W.K exists only as intent, so the tool cannot yet run the game it exists to serve. Two structural gaps keep it single-rulebook. First, `A`/`S`-flag behavior is reached by per-ID dispatch, which doesn't obviously survive past the rulebook it was modeled on; the standing question — *does hardcoded-by-ID hold past one rulebook?* — stays speculative. Second, the goals/tags data carries latent SWN-specific flavor, so what *should* be reusable across rulebooks isn't: it can't be shared without dragging SWN fiction along.

## Why an Arc

The work is several independently shippable initiatives bound by one architectural spine, not a single multi-effort initiative. The binding, cross-cutting decisions — the data-driven registry shape across *both* dispatch surfaces (`A`-flag → abilities engine, `S`-flag → effects engine), the surface contract that makes that split visible in TOML, and the shared `_core` home mechanism for goals/tags — constrain every initiative below and have to be ratified once, up front, before any of them slices cleanly. The [ability/effect catalog](../../../architecture/engine/ability-catalog.md) (already done) is the binding mechanical input those decisions build on.

Shipping order is itself a design question: the effect-keyed A-dispatch increment must land before the A-flag registry can be authored against it, and the lore canon must be seeded before asset authoring can draw on it. A single initiative wouldn't surface those ordering constraints as first-class; an arc does. The registry, the content authoring, the agnostic-core promotion, and the durable docs each ship on their own and are each meaningful alone — but only cohere because they share the spine. That's an arc.

## Candidate Initiatives

Rough and non-binding — Arc-Plan does the real slicing (and decides which of these are independent initiatives versus efforts within one):

- **R.018 — effect-keyed A-dispatch** — re-key ability dispatch on `def.Ability.Effect` instead of `def.ID`; the prerequisite increment for the A-flag registry.
- **A-flag ability registry** — data-driven active abilities (target-select / stat-contest / die-roll → effects).
- **S-flag effect registry** — data-driven passive/reactive effects mapped onto the engine's hook categories.
- **Goals/tags neutralization + shared `_core` home** — data-only flavor neutralization plus the scaffold-compose change that gives every ruleset the same goals/tags.
- **H.A.W.K asset authoring** — the rulebook content itself, fed by a precursor lore session and a per-pass lore-gate that grows the canon.

The fourth deliverable, the durable docs, spans the arc: the catalog is done; the **H.A.W.K lore canon** (`docs/campaigns/hawk/`) is seeded by the lore session and grown through authoring.

## Trigger

Triggered now, at promotion. A pre-plan ability-catalog session surveyed all `A`/`S`-flag specials and settled the **`R.004` call** — build the full data-driven registry now, across both dispatch surfaces, rather than bespoke-only or deferred to a second rulebook. That decision is what tipped F.022 from a single feature into this arc.
