# S-flag Rule Extensions (A.001.6) — Discovery

## In Plain Words

- **What:** 14 SWN assets have a special rule that does nothing in play today. This
  initiative builds what they need: a step that can stop an asset's destruction, asset
  states such as "disabled" or "sheltered", a way to change who owns an asset, attacks
  that do no damage, purchases that change another asset (Stealth), and a model of who
  governs each world, so that buying or moving restricted assets asks that government's
  permission.
- **Why:** until they work, the SWN rulebook is incomplete, and later rulebooks
  (H.A.W.K) need the same engine pieces. The asset rules are now checked against the book
  itself (`docs/rules/swn-assets.md`), which corrected several of our asset entries.
- **How it fits:** the last open S-side initiative of arc `A.001`. It depends on A.001.3
  (done) and gates nothing. The permission work grew past the arc-plan's estimate: it
  adds campaign state, and changes Planetary Seizure and Buy Asset. Five proposed
  efforts; government and movement permission are the largest.

<br/>

## Problem

A.001.3 shipped the S-side model (*trigger → filters → gates → effect*) and every effect
that is a hook or a rule query. The 14 assets left can each register through that model,
but the thing their effect acts on does not exist: game state, a prompt, or a subsystem.
The A.001.3 discovery bounded this set ("Successor initiative — 14 assets") and did not
design it.

| Mechanism | Assets |
|---|---|
| Change of asset owner | Hostile Takeover `SWN-W7-002`, Treachery `SWN-C7-003` |
| State held on an asset | Saboteurs `SWN-C2-002`, Marketers `SWN-W5-001`, Seditionists `SWN-C4-004`, Boltholes `SWN-C5-003` |
| Prompt to a faction that is not acting | False Front `SWN-C1-003`, Medical Center `SWN-W4-003`, Monopoly `SWN-W4-002` |
| Purchase that targets another asset | Stealth `SWN-C3-002` |
| Permission model for movement | Lobbyists `SWN-C2-001`, Popular Movement `SWN-C7-001`, Transport Lockdown `SWN-C6-001`; bypass clauses of Deep Strike Landers `SWN-F7-001` and Blockade Runners `SWN-W5-003` |
| Damage redirected into combat | Cracked Comms `SWN-C5-002` |

<br/>

## Architecture Context

Pages read: `engine/effect-mutation.md`, `engine/hooks.md`, `engine/ability-catalog.md`,
`engine/world-movement.md`, `engine/actions.md`, `engine/turn-pipeline.md`, and (in the
doc review) `engine/goals.md`, `engine/persistence.md`, `engine/orchestrator.md`.

Key Decisions that bind this design:

- **Reactors append; they never cancel** (effect-mutation). `OnMutations` cannot
  suppress a mutation. Bank absorbs a loss with a compensating gain.
- **Mutations are the only state-change currency; `Apply` panics on an unknown type**
  (effect-mutation). Every new state needs a mutation and an `Apply` arm.
- **The S side has two routes: hooks at cycle start, and a direct definition read at
  phase events** (effect-mutation). `on_purchase` fires from a scan for `AssetAdded`
  with cause `buy` or `ability`.
- **An S effect is data keyed on its effect name; the registrar picks the hook
  category** (effect-mutation). The data never names a hook category.
- **Every S effect registers at asset scope and checks `Asset.Active()` at fire time**
  (hooks).
- **Cat 4 attack eligibility refuses on the first no; the query names the striker**
  (hooks). `restrict_attack` already answers it.
- **The collector seam keeps elective choices out of the engine** (hooks).
- **One relocate orchestration serves both harnesses** (world-movement). Drift orders,
  transport cargo, and relocate are the three ways an asset changes world.
- **Buy Asset and the purchase enabler share one purchase path** through
  `domain.CanPurchase` (actions).
- **Hook budgets reset in each faction's own bookkeeping** (turn-pipeline, hooks).
- **Goals emit, never mutate** (goals). Planetary Seizure's change of government is a
  mutation, as its tag already is.
- **`FactionState` is the entire mutable game; load degrades gracefully** (persistence).
  A government that changes in play is state, and a missing record loads as "every
  world NPC".
- **Two input contracts, split by audience** (orchestrator). `PhaseCollector` serves the
  orchestrator's phases; `action.Collector` serves the actions. A prompt asked in both
  places lives in a narrow shared interface, as `hooks.Collector` does.
- Arc decisions not re-opened: the trigger → filters → gates → effect model, and ACD-5's
  transport hybrid.

<br/>

## Current-State Survey (verified this session)

Each hand-off claim checked against source. One does not hold as stated, and one holds
only with a lever the hand-off did not name.

| Claim | Verdict | Source |
|---|---|---|
| `Asset.OwnerID` exists; owner change needs one mutation plus an `Apply` arm | Holds | `domain/asset.go:201`. No mutation changes it today |
| No asset-held state of this kind exists | Holds, with a lever | `Asset` carries `Ready`, `Maintained`, `Stealthy`, `CurrentOrder` only. `Asset.Active()` (`asset.go:213`) is the one usability test: attackers, defenders (`eligibility.go:58`), abilities (`use_asset_ability.go:43`) and S effects all read it |
| Cat 3 hooks receive no collector | Holds | `hooks/interfaces.go:32-34` |
| The purchase path cannot apply a quality to a different asset | **Does not hold** | Buy Asset already does it for Stealth, hard-coded on `SWN-C3-002` (`buy_asset.go:63-77`, `:98-110`): no `AssetAdded`, an `AssetStealthApplied` on a chosen Special Forces asset. This is `B.001` |
| `CanPurchase` is the only permission check; no movement path consults it | Holds | `domain/asset.go:297`; callers `buy_asset.go:179`, `purchase_enabler.go:95`. `prepareMovementDecisions`, `relocate.Run` and cargo selection check nothing |
| No mutation lets the attacker take its own attack | Holds | `attack.go:195-206` — the counter branch reads `defenderDef.Counter` only |

Findings beyond the hand-off:

- **"Planetary government" means "owns a Base on the world".** `HasPlanetaryGovernment`
  (`asset.go:273`) checks Base ownership. The engine has no Planetary Government tag and
  no notion of a world's ruling faction. The rules give permission to the government
  (`swn-faction-mechanics.md:237-241`).
- **Cat 3 runs after the action has emitted its whole list.** A lethal hit emits
  `AssetHPDelta` and `AssetRemoved` together inside the action (`attack.go:218-241`).
  Lethal damage has two emitters only: `applyAssetDamage` (attack and counter) and the
  `self_damage` reactor (`self_damage.go:70-77`). Scavengers, goal progress and the
  digest read `AssetRemoved`.
- **A prompt to a non-acting faction already exists.** `ConfirmRedirectToBase(defenderFaction, …)`
  (`attack.go:167`) asks the defender through the acting faction's collector. One
  shared driver serves every faction today (`F.025`).
- **Three assets have no attack profile their rules text needs.** Treachery and Transport
  Lockdown say "on a successful Cunning vs. Cunning attack" but carry no `[attack]` block.
  Cracked Comms acts "if it succeeds in defending" and carries `counter = ""`. Hit
  detection (`effects/hits.go`) keys on an `AssetHPDelta` with cause `attack`.
- **The `on_purchase` route cannot see a Stealth purchase**, because Stealth adds no
  asset.
- **Medical Center has a second clause:** Repair Asset on its world costs 1 less for
  Special Forces and Military Units. Repair cost has no rule query (`repair_asset.go:77`).

### Rules source check

Checked against the SWN book (`reference/Faction_Rules_SWN.pdf`, pages 218–223,
transcribed to `docs/rules/swn-assets.md`). The TOML descriptions are paraphrases, and
some are wrong. For this initiative, the book is authoritative.

| Asset | Book says | TOML today | Consequence |
|---|---|---|---|
| Monopoly `SWN-W4-002` | "As an action, owners of a monopoly may force one other faction with unstealthed assets on that world to pay them one FacCred. If the target faction can't pay, they lose one asset of their choice on the world." Note `S` | Triggers "when a Monopoly successfully attacks"; flag `A` | An action, not an on-hit effect. Coin goes to the owner. The target chooses only when it cannot pay. The book's own Note (`S`) contradicts its text |
| Treachery `SWN-C7-003` | Attack "Cunning vs. Cunning, Special" | No `[attack]` | Needs a no-damage attack |
| Transport Lockdown `SWN-C6-001` | Attack "Cunning vs. Cunning, special" | No `[attack]` | Needs a no-damage attack |
| Lobbyists `SWN-C2-001` | Attack "Cunning vs. Cunning, special" | No `[attack]` | Same shape. Informers and Seductress (A-side, shipped) also carry it |
| Cracked Comms `SWN-C5-002` | Counter "Special" | `counter = ""` | Needs a no-damage counter |
| Medical Center `SWN-W4-003` | "Once between turns, if a Special Forces or Military Unit asset on the world is destroyed…" | No "once between turns" | A budget gate the TOML text drops |
| Capital Fleet `SWN-F8-001` | Note `A, S`, no `P`. "Planetary government permission is required to raise a capital fleet, but not to move one into a system" | `["S"]` | Resolves `B.009`: not a missing `P`. `P` means "raise *or* transport in", and Capital Fleet needs only the first |
| Deep Strike Landers, Blockade Runners | Bypass text confirmed. Note `A` | `["S"]` | As the hand-off states |
| Space Marines `SWN-F7-003` | "…whether or not the planetary government permits it." Note `A` | `[]` | A third bypass clause. Inert under the book rule below, because Space Marines are not `P` |
| Stealth `SWN-C3-002` | "If the unit normally requires the permission of a planetary government to be moved onto a planet, that permission may be foregone" | Same | Movement only, not purchase |

The book's permission rule (page 216): *"Some assets require the permission of a
planetary government in order to purchase them or move them onto the planet… Whether or
not this permission will be extended will usually depend on the faction's relationship
with a planet's government, though bribes of 1d4 FacCreds can occasionally suffice…
Planetary governments cannot prevent the landing of assets without the
permission-needed quality."* Permission gates `P` assets only, and the government
grants it by judgment, not by a fixed test.

Wider drift found by a mechanical diff of all 72 assets is out of scope here. It is a
candidate backlog audit: transport assets are `A` in the book and `S` in TOML (ACD-5
territory), flag gaps on Mercenaries, Surveyors, Tripwire Cells, Strike Fleet, Space
Marines and Pretech Manufactory, and wrong numbers on Pretech Researchers (cost, TL) and
Surveyors (TL).

<br/>

## Design Summary

The six mechanisms collapse to five pieces of engine, plus one data fix. Every effect
still registers through A.001.3's *trigger → filters → gates → effect* model.

- **Cat 6 lethal interception** — a new hook family at the two places damage removes an
  asset. Boltholes, False Front, Medical Center and Hostile Takeover replace the removal
  instead of reacting to it. Reactors still never cancel.
- **Asset conditions** — one list on the affected asset (`no_ability`, `disabled`,
  `no_attack`, `sheltered`) for Saboteurs, Marketers, Seditionists and Boltholes.
  `Active()` (use) reads `disabled` and `sheltered`; defender filtering reads only
  `sheltered`.
- **The matchup route and special attacks** — the Attack action reads `own_attack` and
  `own_defense_win` entries directly (Saboteurs, Cracked Comms). A no-damage attack hits
  with a zero-damage `AssetHPDelta` (Treachery, Transport Lockdown).
  `AssetOwnerChanged` serves Hostile Takeover and Treachery.
- **A modeled government and one permission request** — a per-world ruling-faction
  record. Purchase, refit and every arrival ask the ruler (the GM for an NPC world).
  Popular Movement, Lobbyists, Transport Lockdown, Capital Fleet and the bypass clauses
  bend that one request.
- **Purchase qualities** — Stealth leaves its hard-coded ID and becomes a `quality`
  definition with a `grant_stealth` entry.

The "prompt to a faction that is not acting" mechanism needs no piece of its own: every
such prompt runs inside an action, addressed to a faction. The data fix: Monopoly is an
action (book text), and Treachery, Transport Lockdown and Cracked Comms gain their
book attack and counter lines.

### Decisions settled this session

- **D-1 — An effect that changes a lethal outcome intercepts before removal; it does
  not react after.** A new hook point at the two lethal-damage emitters
  (`applyAssetDamage`, the `self_damage` reactor) runs before `AssetRemoved` is emitted.
  It serves False Front, Boltholes, Medical Center and Hostile Takeover. Reasons: the
  history never records a death the rules say did not happen, so Scavengers, goal
  progress and the digest stay correct. It runs inside the action, where a collector
  exists. "Reactors never cancel" stays intact, because this is not a reactor.
  Rejected: compensating reactors that re-add the asset, which make every
  `AssetRemoved` reader see a false death. Also rejected: splitting by rules wording
  (Medical Center says the asset "is destroyed"), which needs two mechanisms for one
  outcome.
- **D-2 — Permission comes from a modeled ruling faction per world.** Each world has a
  government: a faction, or none (an NPC government). A request for permission goes to
  that government. Reasons: the book makes permission the government's judgment, and
  Lobbyists react to a rival "gaining" permission, which needs someone to grant it.
  Rejected: the Base-owner shortcut (several factions pass at once, nothing is ever
  granted, and it disagrees with the book); asking the GM with no government model;
  the shortcut with a GM override (two sources of truth for one rule).
  Found while settling it: the Planetary Government tag `T-011` that Planetary Seizure
  adds (`planetary_seizure.go`) carries no world (`domain.Tag`, `faction.go:25`), and
  nothing removes it from the previous ruler.
- **D-3 — A world's government is a per-world record in campaign state.** World ID →
  ruling faction ID; an absent entry is an NPC government. Every permission check asks
  "who rules this world?", and the record answers in one lookup. `T-011` stays a
  faction-level label. Rejected: a world-bearing tag instance per ruled world (the
  lookup scans every faction's tags, and `Tag` stops being a rulebook copy); both
  representations kept in step (two copies of one fact).
- **D-4 — A permission request goes to the ruler's driver, and to the GM for an NPC
  government.** The request names the faction asked, the way `ConfirmRedirectToBase`
  names the defender. With one shared driver the GM answers every request today; the
  routing becomes real when `F.020`/`F.025` split drivers, with no change here.
  Rejected: always the GM, which drops the "who is asked" fact the model exists to hold.
- **D-5 — Governments are GM-authored campaign data, defaulting to NPC.** Existing
  campaigns load unchanged, with every world NPC-governed. Colonists' homeworld
  government is an entry the GM seeds, not engine logic. Rejected: deriving it from
  `T-011` (the tag never recorded a world, so the engine would guess); no authoring
  (a campaign could not start with an established government).
- **D-6 — A faction is never asked for permission on a world it rules.** The grant is
  automatic, but it is still a grant, so Lobbyists still get their contest. Rejected:
  prompting anyway, a question whose answer is almost always yes.
- **D-7 — The defender's protections resolve before the attacker's Hostile Takeover;
  within the defender, automatic before elective.** A protected asset never reaches
  "destroyed", so a Takeover has nothing to claim. A prompt is asked only when nothing
  automatic saved the asset. Rejected: Takeover first (Boltholes would be useless
  against Wealth attackers); the defender chooses (a prompt the rules never offer).
- **D-8 — Lethal interception is a new hook family, Cat 6.** It returns an outcome and
  may prompt, so it is neither a Cat 4 rule query nor a Cat 3 reaction. `hooks/doc.go`
  gains a sixth category. Rejected: a Cat 4 member, which would be the first Cat 4 hook
  to take a collector and emit mutations.
- **D-9 — A captured asset keeps its ID.** Movement orders, cargo lists, history and the
  digest keep pointing at it, and the new owner's `NextAssetID` cannot collide with it.
  The ID's faction prefix then no longer names the owner; `OwnerID` is the truth.
  Rejected: a new ID, which breaks every reference to the asset.
- **D-10 — Interception inside a Cat 3 reactor runs automatic interceptors only.** The
  `self_damage` reactor has no collector, so Boltholes saves a Zealot there, and False
  Front and Medical Center are skipped. A fence-sign marks the gap. Rejected: a
  collector on Cat 3 (changes every reactor for one edge); no interception in Cat 3
  (reads against the Boltholes text).
- **D-11 — A restriction is a condition on the affected asset.** `{kind, source asset,
  how it ends}`, with kinds `no_ability`, `disabled`, `no_attack` and `sheltered`. One
  mutation pair applies and clears a condition. Rejected: a typed field per need (three
  or four mutation pairs, and a new field for every new restriction).
- **D-12 — Marketers' price is offered at the start of the disabled asset's owner's
  turn.** Pay half the purchase cost to lift the condition, or leave it. This reuses the
  `turn_start` phase route and spends no action. Rejected: inside Repair Asset (spends
  the action, which the book does not ask for); any time (no such entry point exists).
- **D-13 — The Attack action fires the attacker's `own_attack` entries at each matchup,
  hit or miss.** This is a direct definition read, the same pattern as `FirePurchase`.
  It serves Saboteurs. Rejected: an "attack made" mutation that changes no state;
  hit-only (drops a clause the book states).
- **D-14 — Seditionists block attacks, not counterattacks.** The book names counters
  when it means them (Lawyers: "cannot attack or counterattack").
- **D-15 — A special hit emits a zero-damage `AssetHPDelta`.** `hits()` sees it
  unchanged, so Treachery and Transport Lockdown are ordinary `own_attack_hit`
  reactions: one trigger, one route. A special attack skips the base-redirect prompt.
  Rejected: firing hit entries from the matchup route, which gives `own_attack_hit` two
  routes when the registrar contract keys the route on the trigger.
- **D-16 — An `[attack]` block with no `damage` is a special attack. Only Treachery
  and Transport Lockdown get one.** Informers and Seductress (A-side abilities) and
  Lobbyists (whose test is the permission contest) do not attack through the Attack
  action. Giving them a block would offer an attack that does nothing.
- **D-17 — Cracked Comms fires from the matchup route on `own_defense_win`.** When the
  defender wins, the Attack action fires the defender's `own_defense_win` entries by
  direct read: the same route as D-13. Rejected: a second new hook family in the
  counter branch for one asset.
- **D-18 — Cracked Comms' self-attack is a full roll.** The attacker rolls its attack
  against its own defense. A hit deals its own attack damage to itself; a miss deals its
  own counter damage, if it has one. The defender chooses whether to use it ("can").
  Lethal results go through Cat 6.
- **D-19 — A definition that adds no asset says so explicitly.** A field such as
  `quality = true`. The loader requires a quality to carry one `on_purchase` effect
  aimed at another asset. Rejected: inferring it from the effect (the reader must know
  a hidden rule). Follows A.001.3's D-4.
- **D-20 — `filters` keep one block; the registrar defines what they select.** For
  `grant_stealth`, filters select the asset the effect acts on. The model already reads
  owner and scope per trigger, and each registrar documents its own reading. This closes
  A.001.3 Open Question 2. A `[target]` block stays available for a future effect that
  needs both meanings at once. Rejected: a `[target]` block now (a fifth part to the
  model for one asset).
- **D-21 — Stealth's effect is a new `grant_stealth`.** `apply_stealth` stays
  self-only. Each registrar's accepted set stays simple. Rejected: one effect with two
  readings.
- **D-22 — A quality with no eligible target is not offered for purchase.** Today Stealth
  can be bought with no target, and the Coin is lost.
- **D-23 — Refit asks for permission through the same request as Buy Asset and the
  purchase enabler.** The book requires "governmental permission" to refit into a `P`
  asset. Today refit checks nothing (`CanPurchase` has two callers). Rejected: purchase
  only (leaves a third purchase-like path drifting).
- **D-24 — A drift order asks for permission when it arrives, not when it is issued.**
  The government asked is the one ruling at arrival: the book's "vulnerable entry phase".
  The request runs between the tick and its apply, inside the movement phase, where the
  phase collector is available. Relocate is instant, so it asks at its one step.
  Rejected: asking at issue (a government that changes during a long transit is never
  asked). What a refusal does is settled below.
- **D-25 — The transport bypass is a `TransportProfile` field.** `ignores_permission`
  on Deep Strike Landers and Blockade Runners. Transport stays a typed profile, so
  `R.019` stays out of this initiative. Stealth's bypass is an engine rule: a stealthed
  `P` asset moving onto a world needs no permission. Rejected: S effects, which pull in
  the migration of eight working transport assets.
- **D-26 — Capital Fleet's raise-only permission is an S effect.** Flags stay as the
  book has them (no `P`). An entry `permission_query` → `require_permission`, scoped to
  purchase. Closes `B.009`. Rejected: a definition field beside the `P` flag (two ways
  to say "needs permission").
- **D-27 — A refused arrival holds at the edge and asks again next tick.** The asset
  stops at the last hex before the world, and the order stays open. The GM may revise or
  cancel it. Rejected: cancel (the GM must re-issue to try again); turn back (a reverse
  order built for a refusal).
- **D-28 — One refusal holds the whole order, cargo included.** Immobile cargo is never
  left in space without its carrier. Rejected: landing the transport and leaving the
  refused cargo behind.
- **D-29 — Lobbyists contest every grant, automatic ones included.** Lobbyists act after
  a grant, and Popular Movement's "always grants" decides the grant, so the two compose.
- **D-30 — "Cannot be re-attempted until next turn" needs no stored denial.** One action
  means one purchase per turn, and a held order is asked again only at the next tick.
  Several relocates in one turn are the only way to retry; a per-turn mark in
  `HookBudgets` covers them. Rejected: a stored (faction, world) denial record.
- **D-31 — Transport Lockdown lasts while its asset is active on that world.** Checked
  when read, like Seditionists' link. Rejected: one turn (a 20-Coin asset buying one
  turn of friction); permanent (no event in the book clears it).
- **D-32 — Transport Lockdown catches every arrival of the rival faction on that
  world:** drift order, cargo or relocate, `P` or not. One check, at the arrival point
  D-24 adds. Rejected: carried moves only (the engine would have to tell carried moves
  from self-propelled ones).
- **D-33 — Under Transport Lockdown, a drift arrival pays 1d4 Coin and holds one turn at
  the edge; a relocate cannot land there at all.** An instant move cannot wait. Rejected:
  turning the relocate into a pending order (new state for relocate); charging and
  landing at once (drops the wait).
- **D-34 — A bribe is the GM's judgment, not a modeled option.** The request's answer is
  grant or refuse. Revisit when AI factions decide permission (`F.020`).
- **D-35 — A purchase refused permission ends the action with no purchase.** No asset,
  no Coin charged, and the turn's action is used. The attempt was made and failed, which
  matches Lobbyists' "cannot be re-attempted until next turn". Rejected: picking again
  (a refusal would cost nothing).
- **D-36 — Transport Lockdown's lock-out list lives on the Lockdown asset.** A mutation
  adds the struck faction on each hit. The record lives and dies with its source (D-31),
  so no cleanup is needed. Rejected: a campaign-state record (needs explicit cleanup).
- **D-37 — Medical Center's repair discount is a new Cat 4 member,
  `RepairCostModifier`.** It chains like the other cost resolvers, and Repair Asset
  consults it for each asset repaired. Follows "Cat 4 is a typed family". Rejected:
  deferring the clause (the asset ships half done).

Settled in the doc review:

- **D-38 — Use and targeting are separate tests; a disabled asset can still be
  attacked.** `Asset.Active()` (use) reads `disabled` and `sheltered`. Defender
  filtering (`eligibility.go:58`, which reads `Active()` today) gets its own test: the
  current one plus "not `sheltered`", without `disabled`. The book makes Marketers'
  target "useless", not "untouchable"; only Boltholes says untouchable. Existing
  behavior for unready and unmaintained assets does not change. Rejected: one gate (a
  Marketers hit would shield its target).
- **D-39 — The `turn_start` phase route carries the phase collector.** D-12's prompt
  needs one, and `FireContext` has none (`effect/registrar.go:46`). The bookkeeping
  phase already receives `collectors` (`orchestrator.go:195`).
- **D-40 — A change of owner also clears the asset's cross-faction links.** It cancels
  the asset's own movement order and removes it from any transport manifest it rides
  in. Otherwise a later cargo `AssetMoved` addresses the asset under its old owner, which
  is a fatal `MutationMiss`. A captured transport loses its order, so the old owner's
  cargo stays where it is. Conditions sourced by the captured asset keep working,
  because they name the asset, not its owner.
- **D-41 — A permission request is one method in a narrow shared collector
  interface,** embedded in both `PhaseCollector` (movement-phase arrivals) and
  `action.Collector` (purchase, refit, relocate), as `hooks.Collector` is today. Follows
  "two input contracts, split by audience".
- **D-42 — A change of government is a mutation, `WorldGovernmentChanged`,** emitted by
  Planetary Seizure on completion. The record lives in `FactionState` (persistence:
  the entire mutable game), and a missing record loads as "every world NPC".

<br/>

## Lethal Damage Interception (Cat 6)

D-1, D-7, D-8, D-10. One dispatch point serves four assets. It runs where an asset would
be removed by damage: `applyAssetDamage` (attack and counter) and the `self_damage`
reactor.

| Asset | Owner | Kind | Outcome in place of removal |
|---|---|---|---|
| Boltholes | defender | automatic | Target set to 0 HP and held unusable until full HP (asset state) |
| False Front | defender | elective — sacrifice | False Front removed; the killing blow is nullified |
| Medical Center | defender | elective — pay half purchase cost | Target set to 1 HP. Once between turns (`once_per`). Its repair discount is a separate Cat 4 `RepairCostModifier` (D-37) |
| Hostile Takeover | attacker | automatic | Target set to 1 HP and changes owner |

Order: the defender's automatic interceptors, then the defender's elective ones, then
the attacker's. The first that takes the hit ends the walk. Inside a Cat 3 reactor, only
the automatic ones run.

The action keeps its own HP tracker (`attack.go:62`). An intercepted asset must read
correctly for the rest of the matchup sequence: a sheltered asset is no longer a live
defender, and a captured asset is no longer a rival.

<br/>

## Change of Asset Owner

D-9. A new mutation, `AssetOwnerChanged`, moves an asset from one faction's `Assets` map
to another's and sets `OwnerID`. The ID is kept.

- **Hostile Takeover** emits it from its Cat 6 interceptor, together with the HP change
  to 1.
- **Treachery** emits it from an ordinary on-hit reaction, which only appends: the
  Treachery asset is removed, the owner gains 5 Coin, and the target changes sides at
  its current HP. Treachery's "special" attack (no damage) is designed with the
  other special attacks.
- "Even if the faction does not otherwise have the attributes": nothing new. The
  over-limit maintenance surcharge already applies.
- The captured asset's own S effects stay registered under the old owner until the next
  cycle. They fail `active()`, because the old owner no longer holds the asset, so they
  stay quiet rather than misfire.
- **Cross-faction links (D-40).** The change of owner also cancels the asset's movement
  order and removes it from any manifest it rides in. A captured transport's old-owner
  cargo stays where it is.

<br/>

## State Held on an Asset

D-11 to D-14. `Asset` gains a list of conditions. Each names its kind and its source
asset, and the code that reads it decides what it blocks.

| Kind | Source | Blocks | Read by | Ends |
|---|---|---|---|---|
| `no_ability` | Saboteurs | Use Asset Ability | Use Asset Ability candidates | Start of the source owner's next turn |
| `disabled` | Marketers | all use; can still be attacked (D-38) | `Asset.Active()` | The owner pays half the purchase cost at its turn start (D-12, D-39) |
| `no_attack` | Seditionists | starting an attack (not counters, D-14) | attacker eligibility | The source attaches elsewhere, leaves the planet, or is gone (checked when read) |
| `sheltered` | Boltholes (from Cat 6) | all use; also untouchable | `Asset.Active()` and defender filtering (D-38) | Repaired to full HP. If the source is destroyed first, the asset is destroyed too (a reactor on the Boltholes' removal) |

- **Saboteurs** fire from the matchup route (D-13), hit or miss.
- **Marketers** (A-side ability): a contest, then a prompt to the target faction (pay
  now, or the asset is disabled). The prompt runs inside the action, addressed to the
  target, like `ConfirmRedirectToBase`.
- **Seditionists** (A-side ability): pay 1d4 Coin and choose an enemy asset on the
  planet. Attaching clears the condition from the previous target. The asset survives
  the target's death.
- **Boltholes**: setting HP to 0 already fails `Active()`. The condition keeps the asset
  unusable through partial repair until it reaches full HP.

<br/>

## Prompt to a Faction That Is Not Acting

No longer a mechanism of its own; earlier decisions dissolve it. The survey premise was
that these prompts arise in Cat 3, which has no collector. None of them does:

- **False Front, Medical Center** prompt from Cat 6, inside the action (D-1).
- **Monopoly** is an action (book text, Rules source check). It becomes an A-side
  ability: choose one other faction with unstealthed assets on the world; that faction
  pays the owner 1 Coin, or, if it cannot pay, loses one asset of its choice on the
  world. The choice is a prompt inside the action, addressed to the target. Its TOML
  description is corrected to the book's text; its `A` flag stays.
- **Marketers, Cracked Comms** prompt inside the action as well (D-12 is the one
  exception: a turn-start prompt to the acting faction itself).

Each prompt names the faction it asks, as `ConfirmRedirectToBase` does
(`attack.go:167`). One shared driver answers them all today; `F.025`/`F.020` route them
later with no change here.

<br/>

## Purchase That Targets Another Asset

D-19 to D-22. Closes `B.001`. Stealth's TOML gains `quality = true` and one entry:
`on_purchase` → `grant_stealth`, with filters `owner = "own"`, `scope = "world"`,
`asset_type = ["Special Forces"]`, `stealthy = false`. Read as the target.

Buy Asset drops every `SWN-C3-002` check:

- **Menu:** a quality is offered on a world only if its filters match at least one
  asset there (D-22).
- **Inputs:** for a quality, Buy Asset reads its entry and prompts for a target among
  the matching assets.
- **Output:** no `AssetAdded`; the Coin charges as usual, and `grant_stealth` emits
  `AssetStealthApplied` on the target.

This is a fourth direct definition read at an action event, beside `FirePurchase`,
which cannot serve here: it looks for `AssetAdded`, and the target prompt has to run in
`Inputs`.

<br/>

## Permission Model

D-2 to D-6, D-23 to D-34. One request, asked at every point where a `P` asset (or a
Capital Fleet purchase) enters a world.

**The government.** `FactionState` holds world → ruling faction (D-3, D-42), authored by
the GM and defaulting to NPC (D-5). Planetary Seizure emits `WorldGovernmentChanged` on
completion, which replaces the previous ruler (D-42). A request goes to the ruler's
driver, or to the GM for an NPC world (D-4), through one method in a narrow collector
interface that both `PhaseCollector` and `action.Collector` embed (D-41). A faction asking
on its own world is granted at once (D-6).

**Where a request is asked.**

| Path | When | On refusal |
|---|---|---|
| Buy Asset, purchase enabler, Refit (D-23) | After the asset is chosen | The action ends with no purchase and no charge (D-35) |
| Drift order, cargo included | At arrival, between tick and apply (D-24) | Hold at the edge; ask again next tick (D-27). One refusal holds the whole order (D-28) |
| Relocate (A-ability, Transit Web grant) | At the destination choice | Destination not allowed |

**What bends it.**

| Asset or rule | Effect |
|---|---|
| Popular Movement | On its world, its owner's requests are granted at once (`permission_query` → `grant_permission`) |
| Capital Fleet | Purchase needs permission although it has no `P` (`require_permission`, D-26) |
| Lobbyists | After any grant to a rival on its world (D-29), its owner may run a Cunning vs. Cunning contest; a win turns the grant into a refusal. No stored denial (D-30) |
| Transport Lockdown | On a hit, the rival faction's arrivals on that world (all of them, D-32) cost 1d4 Coin and hold one turn; relocates there are refused (D-33). Lasts while the Lockdown asset is active there (D-31). The lock-out list lives on the Lockdown asset (D-36) |
| Deep Strike Landers, Blockade Runners | `TransportProfile.ignores_permission` (D-25) |
| Stealth | A stealthed `P` asset moving onto a world needs no permission (D-25). Purchase still does |
| Bribe | The deciding party's judgment (D-34) |

Space Marines' "whether or not the planetary government permits it" needs nothing:
they are not `P`, and a government cannot stop a non-`P` asset.

<br/>

## Special Attacks and the Matchup Route

D-13, D-15, D-16. Two things are added to the Attack action:

- **A special attack** (`[attack]` with no `damage`) rolls as usual. On a hit it emits
  `AssetHPDelta{Delta: 0}` and skips the base-redirect prompt. The defender's counter
  works as usual. Treachery's and Transport Lockdown's effects are `own_attack_hit`
  reactions.
- **The matchup route** is a direct definition read at each matchup, beside the hook
  routes: the attacker's `own_attack` entries (hit or miss), and the defender's
  `own_defense_win` entries when the defender wins. This is the third phase route,
  after `turn_start` and `on_purchase`.

<br/>

## Damage Redirected into Combat

D-17, D-18. Cracked Comms `own_defense_win` → `counterstrike`. When Cracked Comms wins
a defense, its owner is asked whether to use it (inside the action, addressed to the
defender). If yes, the attacking asset rolls its attack against its own defense. A hit
deals its attack damage to itself; a miss deals its counter damage, if any. Lethal damage
goes through Cat 6. Cracked Comms has no counter of its own, so nothing else happens in
the counter branch.

<br/>

## User-Facing Impact

New prompts, each addressed to the faction that decides (one shared driver today):

| Prompt | Asked of | Where |
|---|---|---|
| Permission request | The ruler of the world, or the GM for NPC | Purchase, refit, relocate (action); arrival (movement phase) |
| Lobbyists' contest | Lobbyists' owner | After any grant to a rival on its world |
| False Front sacrifice | Defender | Cat 6, inside the action |
| Medical Center restore | Defender | Cat 6, inside the action |
| Marketers: pay or be disabled | Target faction | Inside Marketers' ability |
| Marketers: pay to lift | Owner of the disabled asset | Its turn start (D-39) |
| Cracked Comms self-attack | Defender | Matchup route, inside the action |
| Monopoly: asset to lose | Target faction, when it cannot pay | Inside Monopoly's ability |
| Seditionists' target, Stealth's target | Acting faction | Inside the ability; inside Buy Asset |

The GM also authors each world's starting government (D-5): in the faction wizard, a
Planetary Government tag pick asks which world (plan Decision 10). The Buy Asset menu now
offers `P` assets on worlds the faction does not rule; the request comes after the choice
(D-35). A world nobody governs is NPC, so a faction that has not been given its own
homeworld in the record is now asked, through the GM, when it buys a `P` asset there;
today the Base-owner shortcut passes it silently. The digest gains zero-damage hits, owner
changes, conditions, permission refusals and held arrivals (Open Question 8).

<br/>

## Replaces / Retires

- `domain.HasPlanetaryGovernment` and `PurchaseTerms.HoldsGovernment` — the Base-owner
  shortcut, replaced by the permission request (D-2)
- Every `SWN-C3-002` check in `buy_asset.go` — replaced by `quality` and `grant_stealth`
  (D-19 to D-22)
- `Active()` as the defender filter — defender filtering gets its own test (D-38)
- Monopoly's TOML description — replaced by the book's text
- Treachery's, Transport Lockdown's and Cracked Comms' missing attack and counter
  lines — added from the book (D-16; Cracked Comms' counter is its `own_defense_win`
  entry)

<br/>

## Proposed Effort Slicing

Agreed with Robert this session; Plan ratifies it and writes the overview plan.

| Effort | Contents | Depends on |
|---|---|---|
| 1 | **Combat foundations:** asset conditions, `AssetOwnerChanged`, the matchup route, special attacks. Saboteurs, Seditionists, Marketers, Treachery, Cracked Comms | — |
| 2 | **Cat 6 interception:** Boltholes, False Front, Medical Center (with `RepairCostModifier`), Hostile Takeover | 1 |
| 3 | **Small and independent:** Monopoly (A-side ability), Stealth onto data | — |
| 4 | **Government and purchase permission:** the world record, the Planetary Seizure update, the request at purchase, enabler and refit, Capital Fleet, Popular Movement | — |
| 5 | **Movement permission:** the arrival check and hold, relocate, the bypasses, Lobbyists, Transport Lockdown | 4, and 1 (special attack) |

Efforts 3 and 4 can run beside 1 and 2. Effort 5 is the largest and lands last; it is the
likeliest to need its own Plan session.

<br/>

## Related Items

| Item | In or out | Reason |
|---|---|---|
| `R.019` — `[transport]` onto `[[effects]]` | Out | The bypass is a `TransportProfile` field (D-25); nothing here needs the migration |
| `B.009` — Capital Fleet missing `P` | In, closed | Not a missing flag; a raise-only S effect (D-26) |
| `B.001` — `stealth_applicator` hard-coded | In, closed | Stealth becomes a `quality` with `grant_stealth` (D-19 to D-22) |
| `filters` with two meanings (A.001.3 Open Question 2) | Closed | The registrar defines what its filters select (D-20) |
| `F.025` / `F.020` — per-faction drivers | Not needed | Every prompt runs inside an action, addressed to a faction, on the shared driver |
| AP-6 catalog placement (Marketers, Seditionists, Monopoly) | In | Reconcile `ability-catalog.md` at pre-merge: all three are A-side abilities whose lasting effect is S-side state |

<br/>

## Open Questions

Genuine deferrals to Plan; the design above does not depend on their answers.

| # | Question | Resolved by |
|---|---|---|
| 1 | The Cat 6 interface: signature, outcome type, and how the action's HP tracker learns the outcome | Plan |
| 2 | The government record's shape in `FactionState` (D-42), and how the GM authors the starting record (by hand in `state.toml`, or through the CLI or TUI) | Plan |
| 3 | Cracked Comms' self-attack: which roll hooks apply when `Actor` and `Opponent` are one faction | Plan |
| 4 | Saboteurs' `no_ability` ends at the source owner's next turn start: cleared by that owner's turn-start route or by a sweep | Plan |
| 5 | Planetary Seizure and `T-011`: whether the tag is still added, and whether the previous ruler loses it (only if it rules no other world) | Plan |
| 6 | Marketers and Seditionists in the A-side step pipeline: new `AbilityEffectType`s and how the target prompt composes with the contest | Plan |
| 7 | The `HookBudgets` key for D-30's per-turn relocate denial | Plan |
| 8 | Digest wording for zero-damage hits, owner changes and permission refusals | Plan |
| 9 | Batch order around `AssetOwnerChanged`: a mutation later in the same batch that names the asset under its old owner is a fatal `MutationMiss`. Which emitters must re-address, and does the action's tracker cover them | Plan |
| 10 | A transport captured mid-flight: D-40 cancels its order, which strands the old owner's immobile cargo in space. Accept, or hand the cargo back to its owner at the transport's hex | Plan |
| 11 | Transport Lockdown's hold (D-33) needs a mark on the order (paid, waited). What happens when the rival cannot pay 1d4: hold until it can, or refuse | Plan |
| 12 | D-27's "last hex before the world" when that hex is the origin world itself (a one-hex path) | Plan |

<br/>

## Out of Scope

- The implementation plan and code changes
- H.A.W.K assets (A.001.5)
- Re-opening the trigger → filters → gates → effect model or ACD-5's transport hybrid
- `R.019` (see Related Items)
- **Asset-data drift outside the 14** — the mechanical diff against the book found
  transport assets flagged `S` where the book says `A`, flag gaps on six assets, and
  wrong cost or TL on Pretech Researchers and Surveyors. A candidate backlog audit; Robert
  decides whether to file it
