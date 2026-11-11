# S-flag effect registry (A.001.3) — Discovery

> Initiative of arc `A.001` (data-driven-rulebook). Cross-cutting decisions
> ACD-1…ACD-9 are ratified in the
> [arc-discovery](../../arcs/data-driven-rulebook/data-driven-rulebook-arc-discovery.md)
> and bind here without re-litigation. Boundary set by the
> [arc-plan](../../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md), A.001.3.
> Mechanical decomposition input: the
> [ability & effect catalog](../../../architecture/engine/ability-catalog.md) (ACD-8).
>
> **This doc scopes two initiatives, not one.** The arc-plan treated A.001.3 as a
> single initiative. The asset inventory below shows two bodies of work of very
> different kinds, so this session splits them: A.001.3 (this doc's subject) builds the
> model and every effect that is a hook or a rule query; a successor builds the effects
> that need new state, a new prompt, or a new subsystem. See "The split" below.

## Problem

An asset's `S` flag means the asset has a special feature it applies without spending
the faction's action. 43 SWN assets carry the flag. One mechanic is implemented.

- **Transport is the only built S-flag mechanic.** `core.go:58-61` registers a
  `TransportHandler` for every definition where `def.Transport != nil`. It covers 8
  assets. Nothing else on the S side has a handler.
- **There is nowhere in the data to declare anything else.** `AssetDefinition` carries
  exactly one S-side profile field, `Transport *TransportProfile` (`asset.go:181`).
  A second mechanic needs a second field and a second branch in `core.go`. The pattern
  holds at one mechanic. It does not hold at nineteen.
- **30 declared abilities do nothing.** Their rules text sits in the asset
  descriptions. No data block holds their parameters, and no handler reads them. The
  engine reports nothing, because the data-only skip is a deliberate contract
  (`effect-mutation.md`, "Framework first, handlers incremental").
- **The flag is not a dispatch key, and will not become one.** `domain.FlagSpecial` has
  three references in the repository: its declaration (`asset.go:27`), the parser that
  produces it (`rulebook.go:566`), and one parser test. No engine code reads it. Compare
  `FlagAction`, read by `hasActionFlag` (`use_asset_ability.go:85`), and
  `FlagPermission`, read by `CanPurchase` (`asset.go:284`). This is ACD-1 already in
  force — the flag picks the harness, the mechanic picks the engine — and it is stated
  here so a reader does not go looking for the code that "handles the S flag".

## Architecture Context

Pages read for re-grounding (their `Key Decisions` are binding context):

- [hooks](../../../architecture/engine/hooks.md) — the five categories, the scope lookup,
  and the distributed-dispatch decision. This initiative extends the dispatch, so the
  page's claims were checked line by line against source (see Current-State Survey).
- [effect-mutation](../../../architecture/engine/effect-mutation.md) — the `EffectsEngine`
  registrar, the data-only skip, and the mutation vocabulary every effect emits.
- [actions](../../../architecture/engine/actions.md) — the combat roll path that Cat 1,
  Cat 2 and Cat 5 hooks are dispatched from.
- [world-movement](../../../architecture/engine/world-movement.md) — the movement phase
  and the relocate primitive built by A.001.2.
- [ability & effect catalog](../../../architecture/engine/ability-catalog.md) — the S-flag
  survey (ACD-8, binding as an inventory; its hook-category labels are corrected here).

The binding decisions these pull in: **the rule layer registers hooks, it never
dispatches** (effect-mutation) — the shape this initiative generalizes; **dispatch is
distributed** (hooks) — each category is consulted where its question is asked, so
adding a dispatch site is a local change; **framework first, handlers incremental**
(effect-mutation) — the reason 30 missing abilities are not errors.

## Design Summary

Every S-flag ability in the SWN set decomposes into the same four parts:

> **trigger** (when it fires, and whose turn) → **filters** (who and where it reaches)
> → **gates** (does it apply this time) → **effect** (what changes).

An asset declares a list of these in its TOML. The engine holds a registry keyed on
the **effect name**, mirroring the A-side's `map[domain.AbilityEffectType]EffectHandler`
(`ability/dispatch.go:33`). Each registrar reads its typed parameters and registers
whatever hooks the effect needs.

Three consequences follow.

**The hook category is an implementation detail.** A GM writes a trigger, not "Cat 3".
The registrar maps the trigger to the dispatch point that serves it. Where no
dispatch point serves a trigger, this initiative adds one.

**One declared effect may register more than one hook.** Transport is the proof: it
needs a grant at decision time (offer to carry cargo) and a reactor at resolution time
(keep cargo location in step). The data names one effect; the registrar decides how
many hooks it takes.

**The `S` flag stays a label.** Nothing in the engine reads it today and nothing will
after this work. Registration is driven by the presence of an effects list. This is
ACD-1 already in force: the flag picks the harness, the mechanic picks the engine.

## The vocabularies

Derived from all 33 abilities in scope (30 S-flag, plus Marketers, Seditionists and
Monopoly rerouted from A.001.2 by AP-6). The vocabularies cover both initiatives; the
split table marks which entries A.001.3 builds.

### Triggers

| Trigger | Meaning | Assets |
|---|---|---|
| `turn_start` | The owning faction's turn begins | Party Machine, Panopticon Matrix |
| `on_purchase` | The asset is bought | Psychic Assassins |
| `own_attack_hit` | My asset attacked and won | Zealots, Blockade Fleet, Franchise, Monopoly, Hostile Takeover, Treachery |
| `own_counter_hit` | My asset counterattacked and won | Zealots |
| `own_roll` | My asset is rolling, attack or defense | Integral Protocols, Panopticon Matrix, Surveyors, Book of Secrets |
| `enemy_roll_at_me` | An opponent rolls against my asset | Blackmail, Book of Secrets |
| `own_defense_win` | My asset defended successfully | Cracked Comms |
| `own_asset_dies` | My asset is about to be destroyed | False Front, Boltholes, Medical Center |
| `stealth_arrival` | A rival's stealthed asset lands on my world | Tripwire Cells |
| `coin_loss` | Coin is about to leave my treasury | Bank |
| `tech_level_query` | The engine prices tech level for a purchase | Laboratory, Pretech Researchers, R&D Department |
| `asset_cost_query` | The engine prices an asset | Local Investments |
| `movement_options_query` | The movement phase asks what a faction may do | Transit Web |
| `attack_eligibility_query` | May this asset fight that one? | Planetary Defenses, Lawyers, Integral Protocols |
| `permission_query` | May this faction buy or move here? | Popular Movement, Lobbyists, Transport Lockdown |
| `ability_eligibility_query` | May this asset use its ability? | Saboteurs, Marketers, Seditionists |

### Filters

Filters narrow which events fire the ability. For every asset in the set, the thing
filtered is also the thing the effect acts on, so one block serves both readings. The
first asset that breaks this needs a real answer (see Open Questions).

| Filter | Values | Assets |
|---|---|---|
| `owner` | `own`, `rival`, `any` | every entry |
| `scope` | `self`, `world`, `faction` | every entry |
| `attribute` | a `FactionStat` | Integral Protocols, Panopticon Matrix |
| `phase` | a `hooks.RollPhase` — `attack`, `defense`, `faction_test`, `contested` | Integral Protocols |
| `action` | an action name, matching `Action.Name()` | Surveyors |
| `asset_type` | a list of `AssetType` | Planetary Defenses, Lawyers, Boltholes, Medical Center |
| `category` | a list of `FactionStat` | Pretech Researchers, R&D Department |
| `stealthy` | bool, matching `Asset.Stealthy` | Panopticon Matrix |

`owner` and `scope` are stated on every entry. They have no defaults, so a reader never
has to know one.

### Gates

| Gate | Meaning | Assets |
|---|---|---|
| `once_per` | A per-turn budget. Values `owner_faction`, `target_faction`, `asset` | Blockade Fleet, Franchise, Book of Secrets, Bank |
| `no_stack` | Several copies do not multiply the effect within one query | Local Investments |
| `contest` | A stat test decides whether the effect applies. Carries `roller`, `attacker_stat`, `defender_stat` | Tripwire Cells, Panopticon Matrix, Lobbyists, Marketers, Transport Lockdown |

### Effects

| Effect | Assets | Initiative |
|---|---|---|
| `coin_gain` | Party Machine | A.001.3 |
| `coin_steal` | Blockade Fleet, Franchise | A.001.3 |
| `ignore_coin_loss` | Bank | A.001.3 |
| `self_damage` | Zealots | A.001.3 |
| `apply_stealth` | Psychic Assassins | A.001.3 |
| `clear_stealth` | Tripwire Cells, Panopticon Matrix | A.001.3 |
| `bonus_die` | Integral Protocols, Surveyors, Panopticon Matrix | A.001.3 |
| `remove_tag_dice` | Blackmail | A.001.3 |
| `reroll` | Book of Secrets | A.001.3 |
| `tech_level_floor` | Laboratory, Pretech Researchers, R&D Department | A.001.3 |
| `cost_surcharge` | Local Investments | A.001.3 |
| `grant_movement` | Transit Web | A.001.3 |
| `restrict_attack` | Planetary Defenses, Lawyers, Integral Protocols | A.001.3 |
| `coin_or_forfeit` | Monopoly | successor |
| `prevent_death` | False Front, Boltholes, Medical Center | successor |
| `change_owner` | Hostile Takeover, Treachery | successor |
| `restrict_ability` | Saboteurs, Marketers, Seditionists | successor |
| `grant_permission` / `revoke_permission` | Popular Movement, Lobbyists, Transport Lockdown | successor |
| `counterstrike` | Cracked Comms | successor |

## The data shape

Four named parts per entry. Each answers a different question, and none can stand in
for another.

| Block | Question | Required |
|---|---|---|
| the entry itself | When does it fire, and what does it do? | always |
| `filters` | Who and where does it reach? | always |
| `gates` | Does it apply this time? | when the ability has a gate |
| `<effect name>` | What are the effect's own numbers? | when the effect takes parameters |

The parameter block is named after the declared effect, matching the A-side pattern
(`[assets.CovertShipping.ability.relocate]`). A mismatch between the declared effect
and the parameter block name is therefore detectable at load.

```toml
# a per-turn budget, counted against the target faction
[[assets.Franchise.effects]]
trigger = "own_attack_hit"
effect  = "coin_steal"
[assets.Franchise.effects.filters]
owner = "rival"
scope = "world"
[assets.Franchise.effects.gates]
once_per = "target_faction"
[assets.Franchise.effects.coin_steal]
amount = 1

# an inverted contest: the rival rolls, and fails to keep stealth
[[assets.PanopticonMatrix.effects]]
trigger = "turn_start"
effect  = "clear_stealth"
[assets.PanopticonMatrix.effects.filters]
owner    = "rival"
scope    = "world"
stealthy = true
[assets.PanopticonMatrix.effects.gates.contest]
roller        = "target"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```

The full set of nineteen entries is specified in the Plan session, not here.

### Loud failure

The loader already rejects an unknown asset flag rather than skipping it
(`toAssetFlags`, `rulebook.go:557`). The effects list follows the same discipline:

1. An unknown `trigger`, `effect`, `filter` or `gate` name fails the load.
2. A parameter block whose name does not match the declared effect fails the load.
3. A missing `owner` or `scope` fails the load.
4. A gate or filter the effect's registrar does not accept fails the load.

Rule 4 requires each registrar to declare which gates and filters it accepts. Without
it, a GM can write `once_per` on an effect that ignores it and never learn the ability
does not do what they wrote.

This refines the data-only skip rather than breaking it. An asset with **no** effects
list stays silently inert, exactly as today. An asset with an effects list the engine
cannot honor now fails loudly. Silence is for absent data, not for wrong data.

## Domain Model Changes

- **`AssetDefinition.Effects []EffectDefinition`** — the new list. `Transport
  *TransportProfile` stays alongside it, unchanged (see Decision D-3).
- **`EffectDefinition`** — `Trigger`, `Effect`, `Filters`, `Gates`, plus typed
  parameter pointers per effect, mirroring `AbilityStep`'s
  `Relocate *RelocateParams` / `Purchase *PurchaseParams` shape (`asset.go:129`).
- **`EffectFilters`** — `Owner`, `Scope`, `Attribute`, `Phase`, `Action`, `AssetTypes`,
  `Categories`, `Stealthy`.
- **`EffectGates`** — `OncePer`, `NoStack`, `Contest *Contest`. `Contest` is the
  existing A-side type (`asset.go:113`) plus a `Roller` field naming which side tests.
- **`RelocateParams`** grows `Categories []FactionStat` and `ExcludeTypes []AssetType`.
  Transit Web reaches "non-Starship Cunning or Wealth assets", which the current single
  `AssetType` field cannot express. `TransportProfile` already carries the equivalent
  pair as `CargoTypes` and `ExcludeCategories`.
- **`hooks.MovementAbility`** grows from its `{Name, Description}` placeholder
  (`rule_modifier.go:7`) to carry an effect discriminator and typed parameters, so the
  movement phase can execute what a granter offers.

## What the dispatch must grow

The hook framework is complete in its categories and incomplete in its reach. Nothing
below is a defect. Transport is an attacker-side, own-turn mechanic, so no handler has
yet needed any of it. Each item was verified against source this session.

**1. Cat 3 reactors reach only the acting faction.** All three call sites pass the
faction whose turn it is (`orchestrator.go:315`, `:351`, `:447`), and
`MutationReactorsFor` returns global plus that faction's hooks (`registry.go:144`).
A defender's reactor never runs. Needed by: Zealots (its counterattack half), Tripwire
Cells, Bank, Panopticon Matrix.

**2. Cat 3 reactors never reach asset scope.** All three call sites pass `""` as the
asset ID, and `registry.go:148` skips the asset bucket when the ID is empty.
`TransportHandler` avoids this by registering at faction scope
(`effects/transport.go:25`). `hooks.md:74-75` states that the effects engine registers
asset-scoped hooks, which is not what the code does. `effect-mutation.md` carries the
same claim only in a heading (line 63); its body is correct and says faction-scoped
(line 76). Correct the page or the dispatch — this doc recommends the dispatch, because
per-asset effects are the normal case for the nineteen.

**3. A roll hook sees only the roller's own hooks.** `RollWithHooks` looks up hooks for
the faction it is given (`roll.go:34`, `:73`). `attack.go` calls it twice, once with
the attacker and once with `defenderFaction` (`attack.go:144`), so a defender's own
hooks do work. An **opponent's** hook never applies to your roll. Needed by: Blackmail,
Book of Secrets (its forced-reroll half).

**4. Cat 4 asset-cost lookup is buyer-scoped.** `ResolveAssetCost` reads modifiers for
the buyer (`rules.go:17`). Local Investments belongs to a rival and charges the buyer,
so the owner's hook is never found. This one needs a world-scoped lookup, not just a
wider audience.

**5. `GrantedMovementAbilities` has no caller.** Verified by LSP: the only reference to
`dispatch.GrantedMovementAbilities` is its own definition (`rules.go:54`). The movement
phase must call it. `hooks.md:89-92` says the movement phase already does; that is the
arc-plan's routed reconciliation item, and it is fixed when the call is wired.

**6. `on_purchase` is a second dispatch route, not a hook.** `EffectsEngine.ApplyAll`
runs once per cycle (`effect.go:34`), so an asset bought mid-turn has no registered
hook until the next cycle. Psychic Assassins must be stealthy the moment it is bought.
The purchase path therefore reads the definition's effects list directly. The S side
has two dispatch routes: hooks registered at cycle start, and a direct definition read
at purchase. This is accepted, not worked around (D-5).

**7. Roll hooks are wired into one action only.** `dispatch.RollWithHooks` has exactly
two call sites, both in the Attack action (`attack.go:120`, `:134`). Every other roll
calls the roller directly: the Expand Influence faction test and its rival roll
(`expand_influence.go:92`, `:102`), the free attack inside it (`:182-183`), and the
A-side stat contest. So Cat 1 and Cat 2 hooks fire during a normal attack and nowhere
else. Two supporting signs that this is unfinished rather than intended:
`hooks.RollPhase` defines `PhaseFactionTest` and `PhaseContested` (`types.go:11-12`)
that no call site ever produces, and `RollContext` carries no action name
(`types.go:16-23`), so a hook cannot tell which action it is in.

Needed by: Surveyors (its die applies to Expand Influence), Book of Secrets (any roll
on the world), Blackmail (any attempt to attack or defend against it, including the
free attack inside Expand Influence), Panopticon Matrix (all Cunning attacks and
defenses). The work is to route the remaining rolls through `RollWithHooks` and to add
the action name to `RollContext` so the `action` filter has a source.

**8. No dispatch point asks whether one asset may fight another.** Eligibility is
decided by three plain functions with no hook consulted: `attackerHasTarget`
(`attack.go:230`) and `eligibleDefendersAtHex` (`attack.go:235`) choose who may fight,
and the counterattack branch reads `defenderDef.Counter` directly (`attack.go:192-197`).
Planetary Defenses restricts who may attack it, Lawyers restricts what it may attack
**and** counterattack, and Integral Protocols restricts what it may defend against. All
three are rule queries with a filter, so they need a new Cat 4 query, not new state.

## Decisions settled this session

- **D-1 — An S-flag ability is trigger, filters, gates, effect.** Derived by reading
  all 33 abilities as one set rather than by extending the transport profile. The four
  parts are what the SWN rules text actually varies. Alternative rejected: one typed
  profile per mechanic, following transport. That needs 17 new fields on
  `AssetDefinition` and 17 branches in `core.go`.
- **D-2 — The registry keys on the effect name, and the registrar picks the hook
  category in Go.** Mirrors the A-side (`ability/dispatch.go:33`). The alternative —
  naming the hook category in the data — gives a GM a field they can set wrong and the
  engine must then validate.
- **D-3 — `[transport]` stays as it is.** *Reverses a decision taken earlier in this
  session.* The session first decided transport would migrate into the effects list, on
  the ground that a model which cannot express the one built mechanic is unproven.
  Robert's call: transport works, so leave it. The migration is a candidate refactor
  once the S-flag model is fully built, and is filed as a backlog item. This also means
  ACD-5's hybrid (typed profiles beside a generic block) stands exactly as the
  arc-discovery wrote it, and needs no refinement.
- **D-4 — Gates and filters are separate named blocks, and `owner`/`scope` are always
  explicit.** Robert's call, on the ground that the engine's job should be clear and an
  author should find it hard to write something wrong. Defaults were rejected for the
  same reason: a reader should never have to know one.
- **D-5 — The S side has two dispatch routes.** Hooks registered at cycle start, and a
  direct definition read for `on_purchase`. Accepted rather than worked around, because
  the alternative is registering hooks mid-turn on asset creation, which would make
  registration order depend on turn events.
- **D-6 — A.001.3 owns every effect that is a hook or a rule query. The successor owns
  every effect that needs new state, a new prompt, or a new subsystem.** The line moved
  twice this session before it settled. The first draft split on what the existing
  dispatch could reach, which moves whenever the dispatch changes. The second split on
  whether an effect needs new game state, which was still wrong: `restrict_attack`
  stores nothing and belongs with the other rule queries. The final line is a property
  of the effect, so it stays true as the engine grows. The split itself is driven by
  size — nothing in the successor is blocked by anything in A.001.3.

## The split

The line is what an effect *is*, not what the engine happens to reach today.

> **A.001.3 owns every effect that is a hook or a rule query.** The successor owns
> every effect that needs new state, a new prompt, or a new subsystem.

The split is driven by size, not by blockage. Nothing in the successor is blocked.
A.001.3 already carries 19 assets and eight dispatch items, one of which reaches
outside the S side entirely (item 7). Folding in the rest would make one initiative
nobody can hold in their head.

### A.001.3 — 19 assets

| Asset | Effect | Dispatch work it needs |
|---|---|---|
| Party Machine `SWN-C4-001` | `coin_gain` | none |
| Zealots `SWN-F3-001` | `self_damage` | item 1 (counterattack half) |
| Blockade Fleet `SWN-F5-001` | `coin_steal` | none |
| Franchise `SWN-W1-001` | `coin_steal` | none |
| Psychic Assassins `SWN-F5-003` | `apply_stealth` | item 6 |
| Tripwire Cells `SWN-C4-003` | `clear_stealth` | item 1 |
| Bank `SWN-W4-004` | `ignore_coin_loss` | item 1 |
| Integral Protocols `SWN-F7-002` | `bonus_die`, `restrict_attack` | item 8 |
| Surveyors `SWN-W2-004` | `bonus_die` | item 7 |
| Panopticon Matrix `SWN-C8-001` | `bonus_die`, `clear_stealth` | items 1, 7 |
| Blackmail `SWN-C2-003` | `remove_tag_dice` | items 3, 7 |
| Book of Secrets `SWN-C7-002` | `reroll` | items 3, 7 |
| Planetary Defenses `SWN-F6-002` | `restrict_attack` | item 8 |
| Lawyers `SWN-W2-002` | `restrict_attack` | item 8 |
| Laboratory `SWN-W3-002` | `tech_level_floor` | none |
| Pretech Researchers `SWN-W5-002` | `tech_level_floor` | none |
| R&D Department `SWN-W6-002` | `tech_level_floor` | none |
| Local Investments `SWN-W1-003` | `cost_surcharge` | item 4 |
| Transit Web `SWN-W7-003` | `grant_movement` | item 5 |

`restrict_attack` sits here, not in the successor, because it is a rule query with a
filter — the same shape as `tech_level_query` and `asset_cost_query`, which are already
in this initiative. It stores nothing. Integral Protocols therefore ships whole rather
than split across two initiatives.

Transit Web is the free relocate harness the arc-plan assigned to this initiative
(ACD-4). It invokes A.001.2's relocate primitive without spending the action. Where the
shared relocate code lives is a Plan decision; the requirement here is that there is
**one copy**, reached by both the action-spent harness and the free harness.

### Successor initiative — 14 assets

Grouped by the mechanism each group needs. The sizes are the successor's own Discovery
agenda, and they are not equal — a permission model is a subsystem, while a change of
owner is one mutation.

| Mechanism it needs | Size | Assets |
|---|---|---|
| A change of asset owner | small — one mutation plus an `Apply` arm; `Asset.OwnerID` already exists | Hostile Takeover `SWN-W7-002`, Treachery `SWN-C7-003` |
| State held on an asset — a restriction with a lifetime, and an unusable-until-repaired state | medium | Saboteurs `SWN-C2-002`, Marketers `SWN-W5-001`, Seditionists `SWN-C4-004`, Boltholes `SWN-C5-003` |
| A prompt to a faction that is not acting | medium — Cat 3 receives no collector (`interfaces.go:25-27`), so the reaction path needs one | False Front `SWN-C1-003`, Medical Center `SWN-W4-003`, Monopoly `SWN-W4-002` |
| A purchase that targets another asset | medium — buying it applies a quality to a different asset, which the purchase path cannot express | Stealth `SWN-C3-002` |
| A permission model for movement | large — nothing of the kind exists. `CanPurchase` is the engine's only permission check, and no movement path consults it | Lobbyists `SWN-C2-001`, Popular Movement `SWN-C7-001`, Transport Lockdown `SWN-C6-001` |
| Damage redirected into combat | medium — the attacker takes its own attack, which no mutation expresses | Cracked Comms `SWN-C5-002` |

Three of these assets arrive from A.001.2 by AP-6: Marketers, Seditionists and Monopoly.
Their trigger is A-flag, but the durable effect is S-side, which is why the arc-plan
rerouted them.

The permission model also picks up a clause this initiative does not touch: Deep Strike
Landers `SWN-F7-001` and Blockade Runners `SWN-W5-003` move cargo past an objecting
planetary government. Their transport profiles work; the bypass clause does not.

**This doc bounds the successor. It does not design it.** The successor runs its own
Discovery, whose agenda is the six mechanisms above.

## Current-State Survey (re-grounding, verified this session)

| Claim | Verified in source |
|---|---|
| 43 SWN assets carry the `S` flag | counted across the three asset files: Force 10, Cunning 18, Wealth 15 |
| 8 of them carry a `transport` profile | counted across the three asset files |
| Transport is the only S-flag mechanic with a handler | `core.go:58-61` — one `TransportHandler` per def with `def.Transport != nil` |
| The S flag is unread by the engine | LSP `findReferences` on `domain.FlagSpecial`: 3 hits — the constant, `rulebook.go:566`, one parser test |
| The A and P flags are read | `use_asset_ability.go:85` (`hasActionFlag`); `asset.go:284` via `CanPurchase`, reached from `buy_asset.go:174` and `purchase_enabler.go:95` |
| Transport is implemented in two places | `effects/transport.go:45` (the reactor) and `orchestrator.go:664-668` (the movement phase reads `def.Transport` directly and prompts for cargo) |
| Cat 3 dispatch is acting-faction only | `orchestrator.go:315`, `:351`, `:447`; `registry.go:144` |
| Cat 3 dispatch never uses asset scope | all three call sites pass `""`; `registry.go:148` |
| Cat 3 has no collector | `interfaces.go:25-27` — `OnMutations(mutations, factionState, rulebook)` |
| Cat 1 and Cat 2 do serve the defender | `attack.go:144` calls `RollWithHooks` a second time with `defenderFaction` |
| Roll hooks fire in the Attack action only | `dispatch.RollWithHooks` has two call sites, `attack.go:120` and `:134`. Expand Influence calls the roller directly (`expand_influence.go:92`, `:102`, `:182-183`) |
| Two `RollPhase` values are never produced | `PhaseFactionTest` and `PhaseContested` (`types.go:11-12`) appear at no call site |
| `RollContext` carries no action name | `types.go:16-23` — `Phase`, `Actor`, `Opponent`, `Attribute`, `Asset`, `World` |
| No hook is consulted on who may fight whom | `attackerHasTarget` (`attack.go:230`), `eligibleDefendersAtHex` (`:235`), and the counterattack branch (`:192-197`) all decide it in plain Go |
| `GrantedMovementAbilities` has no caller | LSP `findReferences`: one hit, its own definition at `rules.go:54` |
| Per-turn budgets already exist | `ModifierOffer.BudgetKey` and `faction.HookBudgets` (`roll.go:47`, `:78`), reset each turn by bookkeeping |
| Government permission is checked only at purchase | `CanPurchase` has two callers, both purchase paths. No movement path consults it |
| The permission gate on movement is absent | no equivalent of `HasPlanetaryGovernment` in the movement phase |

Two SWN assets carry the `S` flag and need no code: Capital Fleet `SWN-F8-001` and
Mercenaries `SWN-W3-003`. Their whole special is a maintenance cost, which the
`maintenance` field already carries. A later reader should not go looking for a missing
handler.

Two of the eight transport assets carry a second clause their profile does not express.
Deep Strike Landers `SWN-F7-001` and Blockade Runners `SWN-W5-003` both move cargo past
an objecting planetary government. That clause needs the permission model, so it belongs
to the successor initiative, not here. The eight transport profiles themselves are
untouched by this work.

Three assets carry both `A` and `S`: Informers `SWN-C1-002`, Seductress `SWN-C2-004`,
Covert Shipping `SWN-C3-003`. A.001.2 shipped the active mechanic for each. Their `S`
text restates the same ability, so no S-side handler is owed.

## Reconciliation items

The arc-plan routed three to this initiative. A fourth was found this session.

| Item | Resolution |
|---|---|
| Catalog hook-category labels drifted (Cat 4 = RuleModifier family, Cat 5 = TieResolver) | Renumber the catalog to match `hooks/doc.go` |
| Catalog files free `relocate-other` as Cat 3 | It is a Cat 4 `AssetMovementGranter`. Correct when Transit Web lands |
| `hooks.md:89-92` says the movement phase calls `GrantedMovementAbilities` | It has no caller. Correct the page when the call is wired |
| **New:** `hooks.md:74-75` says the effects engine registers asset-scoped hooks | It registers at faction scope (`effects/transport.go:25`), and Cat 3 dispatch would not find asset-scoped hooks anyway. Fixed by dispatch item 2. `effect-mutation.md` carries the same claim in its heading at line 63 while its body is correct at line 76 |
| **New:** the catalog's S-flag table routes two effects to `coin_steal` and `coin_drain` | Both `AbilityEffectType` values were removed by A.001.2. `domain.AbilityEffectType` now has five constants (`asset.go:52-58`), none of them coin transfer. The catalog's "latent vocabulary" table is stale on the same point |

## Open Questions

| # | Question | Resolved by |
|---|---|---|
| 1 | Where does the shared relocate code live, so the action harness and the free harness use one copy? Candidates: split the pure parts into `world` and keep each harness thin; a new shared package; export it from `ability` and import that from the orchestrator | Plan |
| 2 | `filters` currently serves both "which events fire this" and "which assets the effect acts on". Every asset in the SWN set makes those the same. What happens at the first asset that separates them? | Deferred until an asset needs it. Recorded so the answer is not invented silently |
| 3 | Panopticon Matrix says "at the beginning of every turn". Every faction's turn, or its owner's? The reading changes how often rivals test | Rules reading, before its commit |
| 4 | Book of Secrets says "force an enemy faction to reroll one die". Any roll that faction makes on that world, or only a roll against the owner? | Rules reading, before its commit |
| 5 | Does `once_per = "target_faction"` key its budget on the acting faction's `HookBudgets` map, or does the target faction hold it? The wording says the limit belongs to the target | Plan |
| 6 | How many efforts does A.001.3 split into, and where do the boundaries fall? The dispatch work and the seventeen effects are plainly separable | Plan |
| 7 | Capital Fleet `SWN-F8-001` has `flags = ["S"]`, but its description says government permission is required to raise it. Mercenaries `SWN-W3-003` states the same rule and carries `["S", "P"]`. Is the missing `P` a data error? | Robert, before the successor's permission work |

## Out of Scope

- **The successor initiative's six mechanisms.** A change of asset owner, state held on
  an asset, a prompt to a faction that is not acting, a purchase that targets another
  asset, a permission model for movement, and damage redirected into combat. Bounded
  above, designed in its own Discovery.
- **Migrating `[transport]` onto the effects list.** Deferred by D-3. Filed as a
  backlog refactor, triggered when the S-flag model is fully built.
- **`B.002` (semantic table keys).** Untriggered — effects key on their own names, not
  on goal or tag table keys.
- **Retiring the `S` flag.** It stays as a human-facing label on the asset card. This
  initiative does not make the engine read it, and does not remove it.
- **H.A.W.K assets.** A.001.5 authors them against the vocabularies settled here.
