# Tag Engine — Discovery & Planning

The Tag Engine is the single authority for the mechanical effects of faction tags. Any subsystem that needs to know what a tag does — rolls, purchases, events, actions, goals — consults the Tag Engine rather than open-coding tag logic. This prevents tag mechanics from being scattered across the codebase and gives GMs a single extension point (`tags.toml`) for adding or customizing tags.

<br/>

## Responsibilities

- Interpret tag effect data from static TOML and expose it as structured queries
- Dispatch to bespoke code handlers for tags whose effects cannot be expressed as static data
- Subscribe to domain events (asset destroyed, goal completed, etc.) and emit mutations for tags that react
- Advertise tag-granted abilities and special actions so the relevant subsystems can surface them
- Enforce tag lifecycle rules (tags lost on state changes, tags capped per-faction or per-world)

<br/>

## Tag Effect Categories

A walk of the 20 tags in `tags.toml` revealed seven categories of mechanical effect. A single tag may carry effects in multiple categories — e.g. Deep Rooted is both a roll modifier and a lifecycle-constrained tag.

### A. Roll modifiers
Alter the dice rolled on a contested check or test. Conditions vary: attribute used, role (attacker/defender), asset properties, location, action kind. May be once-per-turn limited or always-on. Also covers reroll rules and tie-outcome overrides.

Examples: Machiavellian, Theocratic, Warlike, Deep Rooted, Fanatical, Perimeter Agency (b).

### B. Purchase modifiers
Affect Buy or Refit flow: unlock assets, restrict assets, modify cost, override gate checks, or apply post-purchase effects (e.g. asset starts Stealthed).

Examples: Eugenics Cult (a), Preceptor Archive (a), Secretive, Planetary Government (a), Technical Expertise (b).

### C. World-property modifiers
Change how a faction perceives a world's properties (e.g. tech level floor). Rarely observed directly — feeds into Category B checks.

Examples: Colonists (a, b), Technical Expertise (a).

### D. Ability and action grants
Add abilities to assets (surfaced via Use Asset Ability) or unlock new top-level faction actions.

Examples: Mercenary Group, Preceptor Archive (b).

### E. Event hooks
React to domain events (asset destroyed, asset moved onto a Base, goal completed). The Tag Engine subscribes and emits mutations in response.

Examples: Scavengers, Pirates, Exchange Consulate (a).

### F. Rival interference
Allow a faction to modify or disrupt another faction's rolls.

Examples: Psychic Academy (b).

### G. Lifecycle constraints
Govern when a tag is held or lost. Orthogonal to the tag's mechanical effects — any effect in A–F may also carry a lifecycle rule.

Examples: Deep Rooted (lost on homeworld change), Planetary Government (held once per planet controlled).

<br/>

## Surface Kinds

The seven categories collapse into three ways the Tag Engine interacts with the rest of the system.

### Queries
Pure functions: given a context, return structured data. Used when another subsystem is about to do something (roll, buy, use ability) and needs to know what tags contribute.

- Per-roll policy — bonus dice, keep-highest, reroll rules, tie overrides (A, F)
- Per-purchase policy — gates, cost modifiers, post-purchase effects (B)
- Treated world properties (C) — consumed by per-purchase policy
- Advertised abilities and actions (D)

### Event subscribers
Reactive. The Tag Engine listens for domain events and emits mutations for tags that respond (E).

### Lifecycle
Orthogonal to queries and subscribers. Tracks tag-held conditions and handles tag gain/loss when triggering state changes occur (G).

<br/>

## Data Model

Two-tier approach to tag effects in TOML:

- **Structured effects** — each tag lists one or more effect records with a `type` discriminator indicating its category (e.g. `roll_modifier`, `purchase_modifier`, `world_modifier`, `event_hook`, `ability_grant`, `special_action`). Each type has a sub-schema defining its conditions and outcomes.
- **Bespoke handlers** — for effects that cannot be expressed as static data (e.g. Psychic Academy's rival reroll, Mercenary Group's hex-movement ability), the tag references a handler by name. The handler is registered in code with the Tag Engine.

Lifecycle rules attach to a tag as a top-level field, independent of any individual effect.

GMs can add new tags by editing TOML when their effects fit an existing type. Bespoke effects require a new code handler — intentionally friction-inducing so the schema is not stretched beyond what it can cleanly express.

<br/>

## Phased Rollout

The Tag Engine's full surface area is broad and most of it has no live consumer yet. Scope is broken into phases aligned with the features that consume each surface.

### Phase 1 — Roll modifiers
Unblocks Attack. Covers category A, the primary effect of 13 of the 20 tags. Schema is focused: conditions, limiters, always-on rules, tie overrides.

### Later phases (order TBD)
- **Event hooks** — available once Attack produces destroyed-asset events
- **Purchase modifiers + world-property modifiers** — paired, since they compose; covers retroactive wiring of Buy and Refit for Secretive, Preceptor Archive, and Planetary Government's `P`-flag
- **Rival interference** — schema extension of Phase 1 once Psychic Academy warrants it
- **Ability and action grants** — blocked on Use Asset Ability and a command-tree decision for top-level tag-granted actions
- **Goal triggers** — blocked on Goal Engine and dependent systems (e.g. XP)

<br/>

## Deferred Tag Effects

Tag effects whose consuming systems do not yet exist. These are discovery-stage gaps: the tag cannot be designed against a surface that is not defined.

- **Colonists (c)** — "can't build Spaceships if population < 100k" — requires world population data
- **Exchange Consulate (a)** — Peaceable Kingdom bonus XP — requires XP tracking and Goal Engine
- **Psychic Academy (a)** — "psionic mentor training" — not mechanically defined
- **Pirates** — movement-into-Base charge — requires movement mechanics beyond Mercenary Group's hex move
- **Preceptor Archive (b)** — Teach Planetary Population — depends on command-tree decision for tag-granted actions

<br/>

> Related: [turn-engine-discovery.md](turn-engine-discovery.md) | [action-resolution-discovery.md](action-resolution-discovery.md) | [goal-engine-discovery.md](goal-engine-discovery.md)
