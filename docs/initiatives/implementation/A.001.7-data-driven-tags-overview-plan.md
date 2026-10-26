# Data-Driven Tags (A.001.7) — Implementation Plan (Overview)

## In Plain Words

- **What:** the overview plan for making all 20 SWN tags work from rulebook data, as
  assets do. It records the decisions that all four efforts share, sets their order,
  and says which efforts get their own Plan session (all four). It holds no code spec.
- **Why:** today a tag works only when Go code knows its bare ID, and 16 of the 20
  tags do nothing in play. The discovery designed the fix. This plan session re-read
  the code, answered the discovery's three open questions, and fixed the order of
  the work, so that each effort plan starts from one agreed base.
- **How it fits:** the second mode of a multi-effort feature (`session-modes.md`
  section 8), on branch `feature/data-driven-tags`. Initiative `A.001.7` of arc
  `A.001`. It gates the H.A.W.K tag pass in `A.001.5`, and `A.001.8` (goals) reuses
  the faction holder that Effort 1 builds.

## Context / Goal

Build the four efforts the [discovery](../discovery/A.001.7-data-driven-tags-discovery.md)
designed: the faction holder and tag IDs (Effort 1), bonus dice and simple grants
(Effort 2), new effects with world and setting rules (Effort 3), and the A-side tag
holder (Effort 4).

- Arc: [data-driven-rulebook](../arcs/A.001-data-driven-rulebook/A.001-data-driven-rulebook-arc-plan.md),
  initiative A.001.7, as amended by AP-9.
- Input: [`E.002` exploration](../explorations/E.002-data-driven-tags-goals-exploration.md),
  options T-1 (tags) and I-1 (IDs).
- Discovery decisions D-1 to D-20 stand. The decisions below answer the discovery's
  Open Questions and fix the effort order. None of them changes a D-decision, except
  Decision 2, which narrows D-7 for a time (see there).
- Rules source: `docs/rules/swn-faction-mechanics.md` and the book, not the `effect`
  text in `tags.toml`.

## Decisions Ratified in Planning

1. **Every effort gets its own Plan session; this file carries no effort spec. DAG
   `1 → 2 → {3, 4}`; run order `1 → 2 → 3 → 4`.** Robert's call. Efforts 2 to 4 write
   literal code against Effort 1's faction holder, which does not exist yet. Effort 1
   is too large to spec beside the shared decisions: the tag-ID change reaches about
   eight readers, plus the holder, four handlers becoming data, the government flag
   and the re-seed. Each effort plan is written just before its effort executes,
   against code that exists. Effort 2 is a dependency of Efforts 3 and 4, because it
   opens the opponent's faction bucket to the roll dispatcher and adds the `homeworld`
   scope. Psychic Academy (Effort 3), Colonists' homeworld TL floor (Effort 3) and the
   Eugenics Cult defense die (Effort 4) need one or both. Efforts 3 and 4 do not
   depend on each other, but both change `dispatch.ResolveWorldTechLevel` and
   `domain.CanPurchase`, so they run in sequence. Rejected: the Effort 1 spec in this
   file (too large); Effort 4 before Effort 3 (no gain). Same shape as the A.001.6
   overview, Decision 2.

2. **The three `R.023` contests roll through `RollWithHooks` with a nil collector:
   no one is asked, free extra dice are taken, and everything else waits. Letting
   them ask the GM is a follow-on row after the roll card.** Robert's call. Resolves
   Open Question 1.
   - Tripwire Cells and Panopticon Matrix run where there is no collector: a Cat 3
     reactor, the `on_purchase` phase route and the `turn_start` phase route.
     Lobbyists runs inside `contestGrant`, which holds a `PermissionCollector`, but
     not the roll collector.
   - With a nil collector, `RollWithHooks` asks nothing:
     - **Cat 1:** it takes every offer with no limit (an empty `BudgetKey`) and skips
       every offer with one. An extra die where the highest is kept is never worse
       for the roller, so a free die is what any player picks. A once-per-turn die
       (Machiavellian, Warlike) spent on a roll the player never chose could cost
       them a later attack, so it waits. The offer filters run first, as today, so
       Blackmail still removes tag dice.
     - **Cat 2:** an elective reroll is declined. A forced reroll still applies, so
       Fanatical's 1s reroll in all three contests.
   - The precedent is `dispatch.InterceptLethal`, where a nil collector asks nothing
     and only automatic offers run (hooks Key Decision on Cat 6). This decision adds
     the free-die rule for Cat 1.
   - **D-7 is narrowed until the follow-on lands.** When `R.025` (the roll card) adds
     Psychic Academy's entry, it will not reach these three contests. Book of Secrets'
     elective reroll and once-per-turn tag dice do not reach them either.
   - Found in review of this decision: the book calls Tripwire Cells' roll "a Cunning
     vs. Cunning attack". Whether Machiavellian's once-per-turn die applies there is
     for the follow-on row to ask; until then, its limit keeps it out.
   - **The follow-on is its own row, after `R.025`:** these three contests show the
     roll card. It decides who sees the card, and when, for a roll that fires after a
     batch, sometimes on another faction's turn. Lobbyists is the cheapest of the
     three, because its call site already prompts. The row is added to
     `planned-work.md` at pre-merge.
   - **This effort lays the structure for the follow-on.** Each contest passes its
     collector to `RollWithHooks` as one argument, `nil` for now, marked with a
     fence-sign that names the follow-on row. The follow-on swaps that argument for
     a real collector. Its real work is to get a collector into the reactor and
     phase routes, not to change the roll path.
   - Rejected: giving reactors and phase effects a collector now. It changes every
     reactor's interface (the reason the `Roller` parameter on `OnMutations` was
     rejected), it brings back the `FireContext` collector that A.001.6 overview
     Decision 4 dropped, and `R.025` would replace those prompts soon after.

3. **The base Starship rule checks in `domain.CanPurchase`, once, for both purchase
   paths. `requires_tag` (D-13) checks there too.** Robert's call. Resolves Open
   Question 2. Buy Asset (`purchasableDefinitions`, `buy_asset.go`) and the purchase
   enabler (`eligiblePurchaseDefs`, `ability/purchase_enabler.go`) both call
   `CanPurchase`. Its doc comment says the shared checks live there "so the two paths
   can't drift". `PurchaseTerms` gains the world's population and the buyer's
   Starship floor; each caller fills them, for the buy world or for the acting
   asset's world. The Effort 3 plan names the fields and designs how the floor is
   resolved (setting base, tag percentages, the waiver flag). The Effort 4 plan adds
   the `requires_tag` check to the same function.

4. **The digest names a tag from the rulebook; `TagAdded` and `TagRemoved` carry the
   ID only.** Robert's call. Resolves Open Question 3. `digest.Build` already takes
   `*rulebook.Rulebook` (`narrative/digest/build.go:80`). A `resolveTagName(tagID,
   rulebook)` goes beside `resolveGoalName` in `resolve.go`, with the same ID
   fallback, so a tag that left the rulebook still renders. The Manage detail view
   (`tui/views/manage/detail/detail.go:101`) reads names from the rulebook the same
   way. The TUI turn formatter has no tag case today
   (`tui/views/turn/execution/format.go:181`), so it does not change.

5. **The faction holder's shape is the Effort 1 plan's to design, under five fixed
   constraints.** Found in planning: 22 registrar files read `ctx.Asset` or `ctx.Def`
   for the hook ID, the offer's `Source` name and the budget key. Only the registrars
   that tags use change. The constraints:
   - A registrar that a tag does not use stays unchanged. A tag entry that names it
     fails at engine construction, because `Accepts` lists the holders per trigger
     (D-14).
   - The context names the source by ID and display name in one place, so the hook
     ID, the offer `Source` and the budget key read the same way for both holders. A
     tag's budget key uses the tag ID in place of the definition ID.
   - The fire-time check for a faction holder is "the live faction still holds the
     tag", read from state, in place of `active(asset)`.
   - The registrar sets `FromTag` when the holder is a faction, so Blackmail's
     `remove_tag_dice` keeps working.
   - `domain.ScopeFaction` already exists as an effect scope. The Effort 1 plan checks
     whether a faction holder can use it before it adds a scope.

6. **Perimeter Agency's detect die applies to three rolls: the A-side reveal ability,
   Tripwire Cells and Panopticon Matrix.** Robert's call (rules reading). The text is
   "roll an extra die when making a test to detect Stealthed assets". The discovery
   counted only the A-side reveal ability (`ability/reveal_stealth.go`), which already
   rolls through `RollWithHooks` with a collector.
   - **Tripwire Cells:** the holder's Cells attack to strip a stealthed asset's
     stealth. The die goes on the holder's attack roll.
   - **Panopticon Matrix:** the stealthed rival tests to keep its stealth, and the
     holder defends. The contest as a whole is a detection test, so the die goes on
     the holder's defense roll.
   - The die has no once-per-turn limit, so Decision 2's free-die rule takes it in
     both contests with no one asked.
   - The Effort 3 plan designs how the entry recognises a detection roll. It also
     names the roll phase each contest roll carries. That phase decides which
     existing dice reach the contests: Panopticon Matrix's own free die (`own_roll`,
     phase `attack` or `defense`, Cunning) would reach its own contest under
     `attack`/`defense`, but not under `contested`. The Effort 3 plan asks Robert
     that reading. Reshapes the discovery's group C row for Perimeter Agency.

## Decision Record — Execution

<!-- Empty at plan time. Each effort plan carries its own Decision Record; this one
holds only reversals of the decisions above. -->

---

## Effort Summary

| # | Effort | Scope | File |
|---|--------|-------|------|
| 1 | Holder foundation | Tag IDs with the `SWN-` prefix, held as IDs; the faction holder in the effects engine; the four Go handlers as data; `government = true`; `engine/tag` deleted; sandbox re-seed | [`-effort-1-plan.md`](./A.001.7-data-driven-tags-effort-1-plan.md) *(to be written)* |
| 2 | Bonus dice and simple grants | `bonus_die` filters (defense phase, asset TL, `homeworld` scope, goal target world); the opponent's faction bucket on rolls; Mercenary Group; Secretive | [`-effort-2-plan.md`](./A.001.7-data-driven-tags-effort-2-plan.md) *(to be written)* |
| 3 | New effects, world and setting rules | `settings.toml`; the base Starship rule and its waiver; Colonists and Technical Expertise; Pirates; Deep Rooted; Exchange Consulate; roll hooks on the detect and `R.023` contests; the Psychic Academy registrar | [`-effort-3-plan.md`](./A.001.7-data-driven-tags-effort-3-plan.md) *(to be written)* |
| 4 | A-side tag holder | The ability registry accepts a faction holder; Teach Planetary Population and its per-faction TL record; Gengineered Slaves, `requires_tag` and the Eugenics Cult die | [`-effort-4-plan.md`](./A.001.7-data-driven-tags-effort-4-plan.md) *(to be written)* |

## Shared Context

### Branch and git cadence

- Branch `feature/data-driven-tags`, cut from `main` at the first Effort 1 execution
  session.
- Commit per session. At each effort boundary: an interim `/code-review` of the
  effort's diff, then `fix:` commits for its findings, then a squash of the effort's
  session commits into one `effort N` commit. Later fixes stay separate `fix:`
  commits. Merge to `main` with `--no-ff`.
- Claude runs the app only with `--campaign sandbox-claude`. Robert's `sandbox` is
  re-seeded by Robert (D-4).

### Findings for the effort plans

Verified against the source in this session (2026-10-26). Each effort plan re-reads
the code before it relies on them.

- **Tag readers today.** `Faction.Tags` is `[]*Tag` (`domain/faction.go:65`), read by
  `HasTag`, `Clone`, the `TagAdded`/`TagRemoved` apply arms
  (`engine/mutation/mutation.go:249–264`), Planetary Seizure
  (`goal/goals/planetary_seizure.go:75–88`), the Manage detail view and the wizard
  (`wizard.go:237` for the picker, `:296` for `pickedGovernmentTag`).
  `domain.TagPlanetaryGovernment` has those two production readers. (Effort 1)
- **Tag loader.** `loadTags` (`rulebook/rulebook.go:584`) is strict through
  `refuseUnknownKeys`. Asset effects convert through `convertEffects`
  (`rulebook/effects.go:116`), which the tag loader reuses. (Effort 1)
- **Effects engine.** `RegisterEffectsList` checks one asset definition's entries;
  `ApplyAll` walks factions by ID, then assets by ID. Tags join both (D-14). The
  orchestrator primes `e.Tag.ApplyAll` then `e.Effect.ApplyAll`
  (`engine/orchestrator.go:45–46`); `core.go:44,59` builds `TagEngine`. (Effort 1)
- **Warlike's once-per-turn** rides on `BudgetKey: "tag:Warlike"` today. The
  `bonus_die` registrar accepts no gates, so the data form needs `once_per` on
  `bonus_die`. (Effort 1)
- **Ties.** `dispatch.ResolveTie` (`hooks/dispatch/rules.go:185`) already reads the
  actor's faction bucket, then the opponent's, and the first resolver wins. Fanatical's
  `lose_ties` at a faction holder needs no dispatcher change. (Effort 1)
- **Roll reach.** Only `rollScopes` (`hooks/dispatch/roll.go`) leaves the opponent's
  faction bucket out, under a fence-sign whose reason is the Go tag hooks. Effort 1
  deletes those hooks; Effort 2 changes `rollScopes` and removes the fence-sign.
  Between the two, a tag entry reaches only its holder's own rolls. (Effort 2)
- **Purchase checks.** Both purchase paths call `domain.CanPurchase`
  (`domain/asset.go:406`) with `PurchaseTerms{Cost, TechCap}` (Decision 3). (Efforts 3
  and 4)
- **Contests.** `contestPasses` (`effect/effects/contest.go`) is called from
  `clear_stealth.go:58` (phase route), `clear_stealth.go:117` (reactor) and
  `withdraw_permission.go:67` (Lobbyists). Its fence-sign names the missing
  collector; Decision 2 replaces it. (Effort 3)
- **Settings.** `rulebook.Load` reads `drift_costs.toml` through `loadDriftCosts`
  (`rulebook/rulebook.go:561`) into `Rulebook.DriftCosts`, read by `DriftCost`. D-10
  folds it into `settings.toml`. (Effort 3)

### Pre-merge additions

- **Graduate to Key Decisions:** the faction holder and the deletion of `engine/tag`
  (effect-mutation); the opponent's faction bucket on rolls, which replaces the
  "never the opponent's faction bucket" decision, and the nil-collector roll
  (hooks); tag IDs held by factions, `settings.toml` and the government role flag
  (persistence); tag names resolved from the rulebook (narrative-digest).
- **`planned-work.md`:** close `B.042`, `F.010` and `R.023` (D-2, D-3, D-16). Add the
  follow-on row from Decision 2. Remove A.001.7.
- **Arc plan:** A.001.5's tag pass is no longer gated.

## Out of Scope

- **The roll card** (`R.025`) and Psychic Academy's `[[effects]]` entry (D-7).
- **Contest rolls that ask the GM** (Decision 2). The follow-on row after `R.025`.
- **H.A.W.K tags.** `A.001.5`'s tag pass.
- **Goal IDs and the goal vocabulary.** `A.001.8` renames the goal references in the
  Imperialists and Exchange Consulate entries.

---

## Work Breakdown

### Effort 1 — Holder foundation

**Scope:** discovery Design Summary, Effort 1. Tag IDs become `SWN-T-001` to
`SWN-T-020`, and factions hold IDs (D-20); the tag record accepts `[[effects]]`;
`Accepts` names its holders; the faction holder registers at `FactionScope` and
checks at fire time that the faction still holds the tag (Decision 5); the four Go
handlers become data with `B.042`'s owner filter (D-2); `government = true` replaces
`domain.TagPlanetaryGovernment`; the digest and Manage detail read names from the
rulebook (Decision 4); `engine/tag` is deleted (D-14); sandbox re-seed (D-4).

**Depends on:** nothing in this initiative.

**Own Plan session:** yes (Decision 1). Expected size: several phases. The ID change
and the holder are separate seams.

### Effort 2 — Bonus dice and simple grants

**Scope:** discovery Design Summary, Effort 2. `bonus_die` filters for the defense
phase, the asset's tech level, the `homeworld` scope and the active goal's target
world (D-5); the roll dispatcher reads the opponent's faction bucket, and the
`rollScopes` fence-sign goes; Mercenary Group through `grant_movement` at a faction
holder; Secretive through `apply_stealth` at `on_purchase`, with `FirePurchase`
reading the buyer's tags (D-12).

**Depends on:** Effort 1.

**Own Plan session:** yes (Decision 1).

### Effort 3 — New effects, world and setting rules

**Scope:** discovery Design Summary, Effort 3. `settings.toml` with the drift costs
and the Starship floor (D-9, D-10); the base Starship rule in `CanPurchase`
(Decision 3), the waiver flag and its wizard toggle; Colonists (D-8) and Technical
Expertise; Pirates' toll and its refusal at order time (D-6); Deep Rooted's
self-removal (D-19); Exchange Consulate's XP roll (D-18); Perimeter Agency's detect
die on the reveal ability, Tripwire Cells and Panopticon Matrix (Decision 6); the
three `R.023` contests on `RollWithHooks` with a nil collector and the free-die rule
(Decision 2, D-16); the Psychic Academy registrar and tests, with no data entry
(D-7).

**Depends on:** Effort 2.

**Own Plan session:** yes (Decision 1).

### Effort 4 — A-side tag holder

**Scope:** discovery Design Summary, Effort 4. The ability registry (`A.001.2`)
accepts a faction holder, so a tag can grant an action; Teach Planetary Population
and its per-faction world TL record, read by `ResolveWorldTechLevel` (D-11);
Gengineered Slaves in `force_assets.toml` with `requires_tag`, checked in
`CanPurchase` (Decision 3, D-13); the Eugenics Cult die.

**Depends on:** Effort 2. Runs after Effort 3 (Decision 1).

**Own Plan session:** yes (Decision 1).
