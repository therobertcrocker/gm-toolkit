# Data-Driven Tags — Discovery

> **Initiative:** `A.001.7`, a feature in the `A.001` arc (arc-plan AP-8).
> **Input:** the [`E.002` exploration](../explorations/E.002-data-driven-tags-goals-exploration.md)
> (options T-1 for tags and I-1 for IDs, both recommended and taken into the arc).
> **Status:** written; waits for Robert's sign-off. D-1 to D-20 are decided. The
> arc plan carries D-1 as amendment AP-9.

## In Plain Words

- **What:** tags get their behavior from rulebook data, as assets do. A tag record
  carries the same `[[effects]]` list as an asset definition. The four Go tag handlers
  go away, and all 20 SWN tags get their rules in data. Some tags need new engine
  parts: new effects, a per-setting rules file, and a tag that grants an action.
- **Why:** today a tag works only when Go code knows its bare ID (`T-020`). A second
  rulebook (H.A.W.K) can get a tag behavior only by reusing an SWN ID. Factions also
  store a full copy of each tag, so a rulebook change never reaches a tag that a
  faction already holds. 16 of the 20 tags do nothing in the engine.
- **How it fits:** the asset side became data in `A.001.1`–`A.001.6`. This initiative
  does the same for tags, with the same registrars. It gates the H.A.W.K tag pass in
  `A.001.5`, and it proves the faction holder before `A.001.8` (goals) reuses the
  effect filters. It runs as four efforts.

<br/>

## Problem

- **Behavior keys on bare IDs.** `TagEngine` (`engine/tag/tag.go`) registers four
  handlers in `New` and finds them by `tag.ID` in `ApplyAll`. An unknown ID is skipped
  with no error.

  | ID | Tag | File | Hook |
  |---|---|---|---|
  | `T-005` | Fanatical | `tag/tags/fanatical.go` | `RollResultHook` (every die showing 1, not elective) + `TieResolver` (loses ties) |
  | `T-013` | Preceptor Archive | `tag/tags/preceptor_archive.go` | `AssetCostModifier` (−1 Coin on TL4+, floor 0) |
  | `T-016` | Scavengers | `tag/tags/scavengers.go` | `MutationReactor` (+1 Coin per combat `AssetRemoved`, any owner: `B.042`) |
  | `T-020` | Warlike | `tag/tags/warlike.go` | `RollModifier` (Force attack +1d10 keep highest, `BudgetKey` `tag:Warlike`, `FromTag`) |

- **`T-011` is a Go constant.** `domain.TagPlanetaryGovernment` (`domain/faction.go:34`)
  has two production readers: `goal/goals/planetary_seizure.go:75–87` (grants the tag
  with a faction's first world, removes it with its last) and the Manage wizard
  (`wizard.go:296`, `pickedGovernmentTag`).
- **Factions hold full copies.** `Faction.Tags` is `[]*Tag` (`domain/faction.go:65`).
  `TagAdded` and `TagRemoved` carry the whole `Tag` (`domain/mutation.go:255–270`),
  and `Clone` deep-copies each tag. The production writers are `NewFaction` (wizard
  picks), the `TagAdded` apply arm (`mutation/mutation.go:257`) and Planetary Seizure.
  The readers that need the name or text are the digest (`TagAdded` / `TagRemoved`
  lines) and the Manage detail view.
- **16 tags are inert.** Their `effect` text is the only statement of the rule.
- **Tag IDs are bare.** `rulebooks/swn/tags.toml` uses `T-001`–`T-020`. Asset IDs
  carry the rulebook prefix since `A.001.1` (`SWN-C1-003`).
- **The engine has no Starship population rule.** The SWN book says Starship-type
  assets cannot be bought on a world with fewer than several hundred thousand
  people. `spatial.Location.Population()` exists, but no engine code reads it.
  Colonists and Technical Expertise change that rule, so it must exist first.

<br/>

## Architecture Context

Pages read: [effect-mutation](../../architecture/engine/effect-mutation.md),
[hooks](../../architecture/engine/hooks.md),
[persistence](../../architecture/engine/persistence.md). Binding Key Decisions:

- **An S effect is data keyed on its effect name; the registrar picks the hook
  category** (effect-mutation). A tag entry uses the same `EffectDefinition` and the
  same registrars. A tag never names a hook category.
- **A known effect with no registrar fails loud** (effect-mutation). Engine
  construction checks every entry. A tag entry passes the same check. The
  "data-only skip" stays only for a tag with no `[[effects]]`.
- **The rule layer registers hooks, it never dispatches** (effect-mutation).
- **The S side has two routes: hooks at cycle start, a direct definition read at
  phase events** (effect-mutation). A tag entry on a phase trigger (`on_purchase` for
  Secretive) needs the phase route to read the faction's tags too.
- **Every S effect registers at asset scope and checks that its asset is active when
  it fires** (hooks). **This changes (D-14).** A faction holder registers at
  `FactionScope` and checks at fire time that the faction still holds the tag.
- **A rule query that serves its owner walks the owner's own asset buckets** (hooks).
  `ResolveWorldTechLevel` and `GrantedMovementAbilities` walk global, faction, then
  own assets, so a faction-scope tag hook is already found.
- **A budget lives on its owner's `HookBudgets`, keyed by effect, trigger, definition
  and target** (hooks). A tag entry keys on the tag ID in place of the definition ID.
- **The roll dispatcher never reads the opponent's faction bucket** (hooks), with a
  fence-sign at `rollScopes` (`hooks/dispatch/roll.go`). Its reason is that Go tag
  hooks assume the roll is their own. **Revisited with Robert (D-7, D-15).** Data
  entries check the owner relation themselves (`rollMatches`), so the reason goes
  away. Defense-die tags and Psychic Academy need the opponent's faction bucket.
- **Elective rerolls keep per-roll timing** (hooks). Cat 2 runs right after each
  roll. This sets the timing of the roll card in D-7.
- **A rule about an asset that is not in play yet is a definition field, not a hook**
  (hooks). This sets D-13 (`requires_tag`) and the Starship floor (D-9).
- **Factions hold IDs, as assets hold `DefinitionID`** (`E.002` T-1). Persistence
  Key Decision on the government record: the record says which worlds, the `T-011`
  tag says whether any. The role marker keeps that split.
- **Strict tag loader** (persistence, `A.001.4` D-8). Unknown keys fail; a duplicate
  ID is refused. The `effects` key joins the tag record.
- **The compose manifest stays out of the campaign** (persistence, `A.001.4` D-1).
  `rulebook.toml` is a compose instruction, not campaign data. This sets D-10.

<br/>

## Audit Findings

### Fit of the 20 SWN tags

Grouped by what each tag needs. From `E.002` (Options, T-1), checked against the
current `tags.toml` text and the rules calls in this session.

| Group | Tags | Needs | Effort |
|---|---|---|---|
| **A. Go handler today** | Fanatical, Preceptor Archive (discount), Scavengers, Warlike | Fanatical: a selector for "every die showing 1" on `reroll`, and a new `lose_ties` effect (Cat 5). Preceptor Archive: a new `cost_discount` effect and a `min_tech_level` filter. Scavengers: an `asset_destroyed` trigger for `coin_gain` with an own-kill / own-loss filter (`B.042`). Warlike: a `once_per` gate on `bonus_die` | E1 |
| **B. Existing effect, new filter or scope** | Imperialists, Machiavellian, Plutocratic, Theocratic, Savage, Perimeter Agency (attack die), Deep Rooted (defense die), Exchange Consulate (defense die), Mercenary Group, Secretive | `bonus_die` filters: defense phase, tech level of the asset, `homeworld` scope, active goal target world. The roll dispatcher reads the opponent's faction bucket. `grant_movement` at a faction holder. `apply_stealth` at `on_purchase` from the buyer's tags | E2 |
| **C. New effect or trigger** | Pirates, Exchange Consulate (XP roll), Deep Rooted (loss), Perimeter Agency (detect die), Psychic Academy (rival reroll) | A movement toll paid to the holder; a `goal_completed` trigger with an XP roll; a tag that removes itself on a homeworld change; roll hooks on the stealth-detect contest; a reroll forced on a rival's roll | E3 |
| **D. World and setting rules** | Colonists, Technical Expertise | `settings.toml`, the base Starship population rule and its waiver, tech-level floors, a government-record entry at faction creation | E3 |
| **D. A-side and asset grants** | Preceptor Archive (Teach Planetary Population), Eugenics Cult | A tag that grants an action; a per-faction world TL record; an asset that only a tag holder can buy | E4 |
| **E. Role marker** | Planetary Government | No effects; `government = true` replaces `domain.TagPlanetaryGovernment` | E1 |
| **Text only** | Psychic Academy (mentor training) | No engine effect (D-7) | — |

### Roll sites

`dispatch.RollWithHooks` has three production callers: Attack (`attack.go`), Expand
Influence (`expand_influence.go`) and the A-side ability contest
(`ability/stage.go`). The bare contests in `R.023` (Tripwire Cells, Panopticon
Matrix, Lobbyists, through `contestPasses`) skip it, so no roll hook reaches them.

<br/>

## Decisions

- **D-1 — All 20 SWN tags are in scope.** Robert's call. This includes the tags that
  grant actions or assets. This overrode the arc plan's A.001.7 boundary ("Punts:
  tags that grant actions or assets"); the arc plan records the change as AP-9. A tag that grants an action needs an A-side
  harness, so the ability registry (`A.001.2`) gains a faction holder too.
- **D-2 — `A.001.7` absorbs `B.042`.** The data form of Scavengers must state an owner
  filter, so it ships with the `E.002` call: own kills and own losses only. The row
  closes when A.001.7 merges.
- **D-3 — `F.010` (Remaining Tags) closes when A.001.7 merges.** With D-1, no SWN tag
  is left over.
- **D-4 — Saves move by a sandbox re-seed, with no migration code.** As in `A.001.4`.
  The two sandbox campaigns are re-seeded or edited by hand. A `state.toml` of the old
  shape fails to load with a clear error.
- **D-5 — Imperialists: the die applies to the holder's Attacks on its Planetary
  Seizure target world.** While the holder's active goal is Planetary Seizure, each
  Attack on that goal's target world gets +1d10, keep highest. The text has no
  once-per-turn limit, so none. The entry names the goal by ID: a data-to-data
  reference, checked at load. `A.001.8` renames the reference when it prefixes goal
  IDs.
- **D-6 — Pirates: rivals pay, on three kinds of movement, and a mover with no Coin
  cannot move.** A rival asset that arrives on a world with the holder's Base of
  Influence pays 1 Coin to the holder. It counts the Move Asset action, each cargo
  asset aboard a transport, and relocate abilities (A-flag and granted). It does not
  count a purchase or a change of owner. A rival with no Coin cannot order the move,
  as with the other movement blockers today.
- **D-7 — Psychic Academy: any rival roll, chosen on a roll card; mentor training is
  text only.** Robert's reach call is the literal one: any roll a rival makes,
  anywhere, once per turn. Asking on each rival roll gives too many prompts. The
  answer is a **roll card** (as in Foundry VTT): before each roll, a card shows every
  Cat 1 modifier with a toggle. After the roll, a card shows the dice and every Cat 2
  reroll, including a rival's Psychic Academy. The holder picks the die there. The
  post-roll card keeps the per-roll timing Key Decision. The roll card is its own row
  (`R.025`), and it gates this effect: **A.001.7 builds the Psychic
  Academy registrar, the reach into the opponent's faction bucket, and tests. The
  `[[effects]]` entry stays out of `tags.toml` until the roll card lands.** The
  roll-card row adds the entry.
- **D-8 — Colonists: a government-record entry at creation, a TL4 homeworld, and 20%
  of the Starship floor.** Faction create writes a government-record entry for the
  homeworld and grants the `T-011` role tag, as for a Planetary Government pick. The
  wizard refuses Colonists when another faction governs the homeworld. The homeworld
  counts as at least TL4 (`tech_level_floor`, `homeworld` scope). The holder's
  Starship floor is 20% of the base, on any world (100,000 at the SWN default).
- **D-9 — A base Starship rule, with a faction waiver.** A Starship-type asset cannot
  be bought on a world with a population under the setting's Starship floor
  (`settings.toml`, SWN default 500,000). A tag lowers the floor for its holder with
  a percentage of the base: Colonists 20%, Technical Expertise 2% (10,000). A faction
  that holds both uses the lower floor. The book's exception ("ships provided or
  supported by another power") is a faction flag. The GM sets it with a Manage
  wizard toggle.
- **D-10 — Per-setting rule values live in `settings.toml`; drift costs fold in.** A
  rulebook folder may supply `settings.toml`, composed like any other data file.
  `rulebook.Load` reads it. `drift_costs.toml` goes away, and its values move into
  `settings.toml`. `rulebook.toml` stays compose-only, for the reason in `A.001.4` D-1:
  its `includes` paths resolve from the source folder, so they mean nothing in the
  campaign. Also, an included folder's manifest is never read, so `_core` could not
  supply a setting there.
- **D-11 — Teach Planetary Population targets a world with the holder's Base.**
  2 Coin, 1d12; on a 12 the world counts as TL4 for the holder from then on. The
  success is stored per faction and world in state, and `ResolveWorldTechLevel` reads
  it for that faction.
- **D-12 — Secretive: every purchased asset begins Stealthed, of any type.** This is
  wider than the Stealth quality, which accepts only Special Forces. It fires on the
  same purchase causes as `FirePurchase` (`buy` and `ability`, not `refit`).
- **D-13 — Gengineered Slaves is an SWN asset with `requires_tag`.** It goes in
  `force_assets.toml` (Force 1, 6 HP, 2 Coin, TL4, Attack Force vs Force 1d6, Counter
  1d4) with a purchase field `requires_tag = "SWN-T-003"`. Buy Asset does not offer it
  to a faction without the tag. The reference is checked at load. The Eugenics Cult
  die (once per turn, on an attack or defense by a Gengineered Slaves asset) is a tag
  entry with an asset-definition filter.
- **D-14 — The effects engine owns tag effects; `engine/tag` is deleted.** One engine
  checks and registers every `[[effects]]` entry, asset or tag. `ApplyAll` walks each
  faction's assets, then its tags. One check path, one home for the holder.
- **D-15 — Four efforts.** See Design Summary.
- **D-16 — `A.001.7` absorbs `R.023`, in E3.** Tripwire Cells, Panopticon Matrix and
  Lobbyists move onto `RollWithHooks`, so "any rival roll" (D-7) and Fanatical reach
  them. The row closes when A.001.7 merges. See Open Questions for the collector gap.
- **D-17 — Readings kept from today or fixed by the data, no call needed.** Fanatical
  rerolls each 1 once (today's dispatcher). Perimeter Agency's "requires tech level 5"
  is TL5, the highest asset TL in SWN data. Savage's "tech level 0" is TL0. Deep
  Rooted's defense die has no once-per-turn limit; the text has none.
- **D-18 — Exchange Consulate's XP roll is automatic.** When the holder completes
  Peaceable Kingdom, roll 1d6; on 4+ gain 1 XP. No choice, so no prompt. The goal is
  named by ID, checked at load, and renamed by `A.001.8`.
- **D-19 — Deep Rooted removes itself on a homeworld change.** On `HomeworldChanged`
  for the holder, the tag emits `TagRemoved` for itself.
- **D-20 — Factions hold tag IDs; tags carry the rulebook prefix.** `Faction.Tags`
  becomes a list of IDs (`tags = ["SWN-T-020"]`). `TagAdded` and `TagRemoved` carry
  the ID. Readers that need a name or text read the rulebook. SWN tags become
  `SWN-T-001`–`SWN-T-020` (`E.002` I-1).

<br/>

## Design Summary

T-1 from `E.002`, at full width. A tag record carries `[[tags.<Key>.effects]]` with
the asset `EffectDefinition`. The effects engine checks each tag entry at
construction against the same registrars. Each registrar states the holders it
accepts per trigger, beside owners, scopes, filters and gates. A faction holder
registers at `FactionScope`. At fire time the hook checks that the faction still
holds the tag, in place of `active(asset)`. The registrar sets `FromTag` when the
holder is a faction, so Blackmail's `remove_tag_dice` keeps working. A tag entry's
budget key uses the tag ID. Factions hold tag IDs, and `engine/tag` is deleted.

### Effort 1 — Holder foundation

- Tag IDs take the `SWN-` prefix; `Faction.Tags` holds IDs; `TagAdded` / `TagRemoved`
  carry the ID; the digest, the Manage detail and the wizard read names from the
  rulebook (D-20).
- `[[effects]]` on the tag record; the strict loader accepts the key.
- `Accepts` gains holders; `RegisterContext` names its holder; the fire-time holder
  check; `FromTag` from the holder; tag-ID budget keys (D-14).
- The four Go handlers become data (group A), with `B.042`'s owner filter (D-2).
- `government = true` on the tag record replaces `domain.TagPlanetaryGovernment`
  (group E).
- Sandbox re-seed (D-4).

### Effort 2 — Bonus dice and simple grants

- `bonus_die` filters for group B: defense phase, asset tech level, `homeworld`
  scope, active-goal target world (D-5).
- The roll dispatcher reads the opponent's faction bucket; the `rollScopes`
  fence-sign and its hooks Key Decision change.
- Mercenary Group through `grant_movement` at a faction holder.
- Secretive through `apply_stealth` at `on_purchase`; `FirePurchase` reads the
  buyer's tags (D-12).

### Effort 3 — New effects, world and setting rules

- `settings.toml` with the drift costs and the Starship floor (D-9, D-10).
- The base Starship rule at purchase, the faction waiver flag and its wizard toggle
  (D-9).
- Colonists (D-8) and Technical Expertise: TL floors, Starship floor percentages,
  the government-record entry at creation.
- Pirates' toll at arrival, and its refusal at order time (D-6).
- Deep Rooted's self-removal (D-19); Exchange Consulate's XP roll (D-18).
- Roll hooks on the stealth-detect contest (Perimeter Agency) and on the three `R.023`
  contests (D-16).
- The Psychic Academy registrar and tests, with no data entry (D-7).

### Effort 4 — A-side tag holder

- The ability registry (`A.001.2`) accepts a faction holder, so a tag can grant an
  action.
- Teach Planetary Population and its per-faction world TL record (D-11).
- Gengineered Slaves, `requires_tag`, and the Eugenics Cult die (D-13).

<br/>

## Replaces / Retires

- `engine/tag/` (`TagEngine`, the four handlers and their tests). The tests move to
  data-driven scenario tests.
- `domain.TagPlanetaryGovernment`.
- The full `Tag` copy in `Faction.Tags`, `TagAdded` and `TagRemoved`.
- `drift_costs.toml`.
- The `rollScopes` fence-sign and the hooks Key Decision "never the opponent's faction
  bucket".
- Rows: `B.042`, `F.010`, `R.023` close at merge. `B.002` is already removed.

<br/>

## Out of Scope

- **The roll card** (D-7). Its own row, `R.025`. It adds Psychic Academy's `[[effects]]` entry.
- **H.A.W.K tags.** `A.001.5`'s tag pass, which this initiative gates.
- **Goal IDs and the goal vocabulary.** `A.001.8`. Imperialists and Exchange Consulate
  name SWN goals by today's IDs; `A.001.8` renames the references.

<br/>

## Open Questions

- **`R.023`'s collector gap (D-16).** `R.023` waits on "when effects reach a
  collector": Tripwire Cells is a reactor and Panopticon Matrix a phase effect, and
  neither can ask the collector. On `RollWithHooks` with no collector, an elective
  offer cannot be asked. The Plan session picks between two answers. One gives
  reactors and phase effects a collector. The other runs those rolls with automatic
  offers only (Fanatical yes, Book of Secrets no) until the roll card lands.
- **Where the Starship rule checks.** Buy Asset and the purchase enabler both buy. The
  Plan session confirms one check covers both, as the purchase scan does for
  `FirePurchase`.
- **The digest needs the rulebook to name a tag (D-20).** Confirm the digest builder
  can reach it, or the mutation keeps the name for display only.
