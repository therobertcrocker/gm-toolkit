# S-flag Rule Extensions (A.001.6) — Implementation Plan (Overview)

## In Plain Words

- **What:** the overview plan for the 14 SWN assets whose special rule does nothing in
  play today. It fixes the decisions shared by all the work, cuts the work into five
  efforts, and sets their order. Each effort gets its own short Plan session, which
  writes the code-level spec.
- **Why:** the discovery settled 42 design decisions. A plan session re-read the code and
  found two decisions whose shape had to change and twelve questions to answer. This
  file records those answers, so that every effort plan starts from one agreed base.
  An approval review on 2026-10-20 amended Decisions 5, 8, 10 and 16 and added
  Decisions 21 to 24; the initiative is signed off under them.
- **How it fits:** the second mode of a multi-effort feature (`session-modes.md`
  section 8), on branch `feature/s-flag-rule-extensions`. The last open S-side initiative
  of arc `A.001`. It gates nothing.

## Context / Goal

Build the five pieces of engine the [discovery](../discovery/s-flag-rule-extensions-discovery.md)
designed, and author the 14 assets on them: lethal interception (Cat 6), asset
conditions, change of owner, the matchup route with special attacks, purchase qualities,
and a modeled world government with one permission request.

- Arc: [data-driven-rulebook](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md),
  initiative A.001.6. The trigger → filters → gates → effect model and ACD-5's transport
  hybrid bind here without re-litigation.
- Discovery decisions D-1 to D-42 stand, except where a decision below changes their
  shape. Those changes name the discovery decision they reshape.
- Rules source: `docs/rules/swn-assets.md` (the book), not the TOML descriptions.
- This is a **multi-effort** initiative. This file is the overview plan. All five efforts
  get their own Plan session (Planning Decision 2); no effort executes from this file.

## Decisions Ratified in Planning

1. **Five efforts, regrouped by surface. DAG `1 → {2, 3}`, `{1, 4} → 5`. Run order
   `1 → 3 → 2 → 4 → 5`.** Marketers and Seditionists leave Effort 1 and join Monopoly in
   Effort 3: all three are A-side abilities that target an asset or a faction and prompt
   it to pay, and they are the AP-6 catalog group. Stealth moves to the head of Effort 4,
   so one effort reworks Buy Asset. Effort 1 becomes pure combat. Effort 3 runs before
   Effort 2 because it is lighter, and it defines the shared pay prompt that Medical
   Center reuses (Decision 19). Reshapes the discovery's Proposed Effort Slicing.

2. **Every effort gets its own Plan session; this file carries no effort spec.** Efforts
   2, 3 and 5 spec literal code against Effort 1's condition model and owner-change
   mutation, which do not exist yet. Effort 1 (about six commits) and Effort 4 (a new
   Cat 4 family and a new collector interface) are too large for the 1,000-line target
   beside the shared decisions. Each effort plan is written just before its effort
   executes, against code that exists. Rejected: specifying Effort 1 here (the overview
   would pass its size target).

3. **A special attack is an `[attack]` block with no `damage`; `AttackProfile.Damage`
   becomes `*domain.DiceRoll`, nil for special.** Found in planning: `convertAsset`
   dereferences the result of `parseDice` (`rulebook.go:211-218`), and `parseDice("")`
   returns nil, so D-16 as written panics at load. Every reader of `.Attack.Damage`
   gains a nil branch. Rejected: a `Special bool` beside a value `Damage` (two fields for
   one fact). Reshapes D-16; resolves Finding 1.

4. **Condition upkeep is a dedicated step at the start of the bookkeeping phase, not the
   `turn_start` effects route.** Found in planning: `FireTurnStart`
   (`effect/effect.go:76`) walks only the current faction's *active* assets and reads
   each asset's *own* `turn_start` entries. A disabled asset fails `Active()`, and its
   definition has no Marketers entry. Saboteurs' condition sits on rival assets. The step
   sits beside `FireTurnStart`, gated on `BookkeepingApplied` the same way
   (`orchestrator.go:209`), and uses the phase collector. It clears every condition
   whose end is "the source owner's next turn start" and whose source asset the current
   faction owns, on any faction's asset (Saboteurs, Effort 1). It offers the current
   faction's disabled assets the pay-to-lift prompt (Marketers, Effort 3). D-12's
   intent stands: turn start, no action spent. D-39's "`FireContext` gains a collector"
   is dropped. Reshapes D-12 and D-39; resolves Finding 2 and Open Question 4.

5. **Owner changes apply last in their batch.** `MutationEngine.Apply` applies a batch
   in two passes: every mutation but `AssetOwnerChanged` in order, then the owner
   changes in order. History keeps the emitted order. *(Amended in the approval review:
   the rule first sat in `applyAndRecord`; a mutation-type rule belongs in the apply
   layer, which already owns the type switch, not in the orchestrator's generic choke
   point. A fence-sign marks the two-pass.)* Reactors read pre-apply state, so they name the
   old owner, and their output applies after the action's. Example: a captured asset
   that counter-hit earlier in the action, with a `self_damage` entry, emits an
   `AssetHPDelta` under its old owner, which is a fatal `MutationMiss` if the owner
   change applied first. Found in planning: `AssetRemoved`'s apply arm calls `delete`
   and records no miss (`mutation.go:61-66`), so the same fault there is silent.
   *(The Effort 1 plan dropped D-40's order cancel and manifest removal: a capture
   changes one asset and every order carries over, Effort 1 Decision 8.)* Rejected: each emitter re-addresses (spread over every present and future
   reactor); `Apply` finds an asset by ID wherever it lives (hides the wrong-owner bugs
   the miss contract exists to catch). Resolves Finding 3 and Open Question 9.

6. **Each condition kind lands with the code that reads it.** Effort 1 builds the model
   and `no_ability`; Effort 3 adds `disabled` (read by `Active()`), D-38's defender test
   (`Asset.Targetable()`) and `no_attack`; Effort 2 adds `sheltered` and puts "not
   `sheltered`" in `Targetable()`. No kind ships before its reader. *(Reshaped in the
   Effort 3 plan, Decision 1: Effort 3 runs first, and `disabled` in `Active()` would
   drop a disabled asset as a target without the separate test.)*

7. **The Cat 6 interface is designed in the Effort 2 plan.** The signature, the outcome
   type, and how the Attack action's HP tracker (`attack.go:62`) learns the outcome. One
   constraint is fixed now: an interceptor runs before `AssetRemoved` is emitted and
   returns the mutations that replace it (D-1). The plan must cover each outcome's read
   in the tracker: sheltered at 0 HP (already reads as not live), restored at 1 HP
   (live), captured at 1 HP (live, but now the attacker's asset; `filterOpposingDefenders`
   reads `OwnerID` from state, `eligibility.go:58`), and sacrificed (False Front dead,
   target live). Resolves Open Question 1 by routing.

8. **The government record is `FactionState.Governments map[string]string`** (TOML
   `governments`, `omitempty`): world ID → ruling faction ID, with an absent key meaning
   an NPC government. A missing table loads as "every world NPC" (D-5) with no load
   change, because a nil map reads as empty. *(Amended in the approval review.)* The
   authoring path is the faction wizard: picking the Planetary Government tag asks
   which world, and the wizard writes the entry (Decision 10). Hand-editing `state.toml`
   stays possible but is not the designed path. No load-time derivation from existing
   tags: a campaign saved before this record loads all-NPC, and Robert regenerates the
   test campaign after the arc. Rejected: a CLI command now (the wizard already owns
   faction authoring); deriving the record at load (the tag never recorded a world, and
   the one campaign that holds it is being regenerated). Resolves Open Question 2.

9. **Cracked Comms' self-attack rolls with `Opponent: nil` and `OpponentAsset: nil`.**
   With `Actor == Opponent`, `rollScopes` (`hooks/dispatch/roll.go:114`) appends the
   faction's asset buckets twice, so every asset hook on that world offers twice. With
   no opponent, each of the faction's hooks is asked once per roll, and its own bonuses
   apply to both its attack and its defense. `ResolveTie` already handles a nil
   `Opponent`. Rejected: deduplicating scopes in `rollScopes` (changes the shared
   dispatcher for one asset). Resolves Open Question 3.

10. **The government record and the `T-011` tag are coupled: the record holds the
    worlds, the tag marks "governs at least one world".** *(Rewritten in the approval
    review; the first version demoted the tag to a label nothing reads, which is the
    flag-nothing-reads pattern AP-6 rejected, and it conflicted with
    `swn-faction-mechanics.md` and the tag's own effect text, both of which say the tag
    grants permission.)* The book makes the tag per-planet ("can be held once per planet
    controlled"); the record is that per-planet instance list, and the tag is its
    coarse presence marker. Three consequences:
    - Planetary Seizure emits `WorldGovernmentChanged` on completion
      (`planetary_seizure.go:41-47`), and `TagAdded` for `T-011` only when the faction
      does not already hold it. When the previous ruler loses its last world, Seizure
      also emits a new `TagRemoved` for it. One tag per faction, however many worlds.
    - The faction wizard, on a `T-011` pick, asks which world and writes the record entry
      at creation (Decision 8). The wizard is the authoring surface; a GM who picks the
      tag never ends up with a tag that governs nothing.
    - Inside Enemy Territory reads the record ("a rival rules this world") in place of
      its own test (`helpers.go:80-102`: a rival Base here and `T-011` anywhere).
    The tag's effect text in `tags.toml` and the Government Permission paragraph in
    `swn-faction-mechanics.md` are retexted at Effort 4 to name the record as the
    per-planet instance. D-3's "two copies of one fact" objection does not apply: the
    record and the tag hold different facts (which worlds, versus whether any). Rejected:
    the tag as a label with no upkeep (a GM-pickable book tag that does nothing); a tag
    instance per world (`Tag` stops being a rulebook copy, D-3). Resolves Open Question 5.

11. **Marketers and Seditionists use the A-side pipeline with an asset target.**
    `resolveTarget` (`ability/stage.go`) handles `TargetAsset`: the candidates are
    rival, unstealthed assets on the acting asset's world, in ID order, offered through
    `SelectAbilityTarget` with `AbilityTargetCandidates.Assets` (both fence-signed as
    reserved for this work). `AbilityTarget.Faction` is set to the asset's owner, so
    `passesGate` runs unchanged. Two new `AbilityEffectType`s: `disable_asset`
    (Marketers; contest gate; the handler prompts the target faction to pay half the
    purchase cost, rounded down, at least 1 Coin (Effort 3 Decision Record 6), or take
    `disabled`; a faction that cannot pay takes the
    condition with no prompt) and `block_attack` (Seditionists; no gate; the handler
    rolls 1d4, charges it, clears this source's `no_attack` from its old target and
    attaches it to the new one). Rejected: one generic `apply_condition` effect with a
    kind parameter (the two assets differ in gate, cost and prompt). Resolves Open
    Question 6.

12. **D-30's per-turn relocate mark is the budget key `permission_refused:<worldID>`**
    on the refused faction's `HookBudgets`. That map clears in the refused faction's own
    bookkeeping, which is the book's "until next turn". Relocate's destination list drops
    a world whose key is spent. The gap it closes: `runGrantedMovement`
    (`orchestrator.go:791`) can run several grants in one movement phase. Resolves Open
    Question 7.

13. **Every new mutation ships its digest and TUI cases in the commit that adds it.**
    The digest notes an unknown type as `"unknown mutation: X"` in the beat
    (`narrative/digest/build.go:441`), and the TUI formatter folds it silently
    (`tui/views/turn/execution/format.go:139`). Wording is written in that commit's spec,
    because it depends on the fields that effort fixes. A zero-damage hit already flows
    into the digest's cross-attack event as `0` damage; it needs a wording branch, not a
    new case. Resolves Open Question 8 as a rule.

14. **A transport captured mid-flight keeps its order and its cargo; the cargo keeps
    its owner.** *(Reversed in the Effort 1 plan, Decision 8, by Robert's call. The first
    version cancelled the order and stranded the cargo.)* The old owner's cargo rides on
    as riders and can attack the transport from its hex; a captured rider stays aboard
    its old owner's transport. Cargo moves address each rider's current owner (Effort 1
    Decision 10), and hooks re-register after a capture (Effort 1 Decision 9). Resolves
    Open Question 10.

15. **A rival that cannot pay Transport Lockdown's 1d4 holds at the edge and is asked
    again next tick, as a refusal does (D-27).** One hold behavior for both causes. The
    order's paid-and-waiting mark is designed in the Effort 5 plan. Resolves Open
    Question 11.

16. **A refused arrival needs no special case for a one-hop path.** The refusal replaces
    the tick's `MovementOrderCompleted` with `MovementOrderHeld{NewStepIdx: len(Path)-2,
    RegionHex: Path[len(Path)-2], Reason}` (Decision 22), whose apply arm sets the step
    and hex exactly as `Progressed` does. `Path[0]` is the asset's own hex, and issue
    already sets `Location = {RegionHex: Path[0]}` with no `WorldID`. So on a two-hex
    path the hold is the in-orbit state right after issue, and the next tick completes
    and asks again (`world/movement.go:41`). The transport reactor gains a `Held` arm
    beside its `Progressed` arm so cargo follows (D-28). *(Amended in the approval
    review: the first version rewrote the completion into a `Progressed`, which put a
    move that did not happen in history.)* Resolves Open Question 12.

17. **Planetary Seizure's phase-1 check treats a capture by the seizer as cleared.**
    `UpdateProgress` (`planetary_seizure.go:55-75`) skips a rival asset only when the
    batch removes it. Before apply, a captured asset is still in the rival's map, so it
    counts. If it was the last rival asset, the next turn's Attack fails `Validate`, the
    phase lock allows only Attack, and the faction skips every turn. The fix (Effort 1,
    with `AssetOwnerChanged`) also counts `AssetOwnerChanged` to the acting faction as
    cleared. Hostile Takeover (Effort 2) emits the same mutation, so one fix covers both.
    *(Extended in the Effort 1 plan, Decision 5: goal progress reads the action's batch
    before the reactors, so Treachery's reactor-emitted capture never reaches it. Phase 1
    also re-checks rival presence in `CheckLock`.)* Resolves Finding 7.

18. **Constructor and pipeline changes ride with their first user.** Repair Asset gains
    the hook registry in Effort 2 (`RepairCostModifier`, D-37). Refit Asset gains it in
    Effort 4 (Popular Movement's grant is a Cat 4 query; the request rides the collector,
    D-41). `TargetAsset` lands in Effort 3 (Decision 11). Today both constructors take
    only the collector (`actions/register.go:27,29`). Resolves Findings 8 and 9.

19. **The pay prompt is one method in a narrow shared collector interface, defined in
    Effort 3.** Marketers' "pay or be disabled" (inside the action), Marketers'
    pay-to-lift (the bookkeeping step, Decision 4) and Medical Center's restore (inside
    the action, Effort 2) all ask one faction whether to pay one amount. The interface is
    embedded in both `PhaseCollector` and `action.Collector`, the D-41 pattern. The
    Effort 3 plan names the method; Effort 2 reuses it. All three use one price,
    `AssetDefinition.HalfPrice()`: half the listed cost, rounded down, at least 1 Coin
    (Effort 3 Decision Record 6).

20. **`B.001` is narrowed, not closed.** It lists two hard-coded IDs. This initiative
    removes `stealth_applicator`'s (`SWN-C3-002` in `buy_asset.go`). The `"G.012"` check
    in Change Homeworld stays out of scope. At pre-merge, the planned-work row is
    rewritten to the `G.012` half. Corrects the discovery's Related Items. Resolves
    Finding 10.

Added in the approval review (2026-10-20):

21. **The matchup route reshapes the "orchestrator owns each call site" Key Decision,
    and says so.** `effect-mutation.md` binds that phase routes are raised by the
    orchestrator (`FireTurnStart`, `FirePurchase`). D-13 and D-17 have the Attack action
    raise the matchup route itself, because Cracked Comms needs an elective prompt and a
    full roll inside the matchup, which no post-action scan can give. The Attack
    constructor therefore gains the effects engine as an explicit dependency through
    `RegisterDefaultActions`, the same way it gets the hook registry. The Key Decision is
    reworded at Effort 1 ("the orchestrator or the action that owns the event raises it")
    and promoted at pre-merge. Rejected: an attack-made record mutation served by a Cat 3
    reactor (covers Saboteurs, not Cracked Comms, so it would leave two routes).

22. **A refusal and a hold each leave a record-only mutation.** `PermissionRefused`
    (Effort 4: faction, world, definition, the deciding party) and `MovementOrderHeld`
    (Effort 5, Decision 16) change no state beyond the hold's step and hex, and each
    gets an apply arm, a digest line and a TUI line (Decision 13). Without them a refused
    purchase that used the turn's action leaves nothing in history, and a held arrival
    read as a move. Precedent: `GoalInitiated` records an event with no mechanical
    effect. Amends D-35 ("ends the action with no purchase" now also emits the record).

23. **Each effort plan states its interface-side work, and the Effort 1 plan decides
    whether `R.014` runs first.** The discovery's prompt table adds about six ask kinds.
    Each one touches the adapter enum (`tui/adapter/channels.go`), an adapter method, the
    two parallel switches in the execution view that `R.014` names, and an overlay file,
    and `tui/adapter` has no tests (`F.027`). An effort plan that adds a prompt lists
    those files in its commit; the Effort 1 plan judges whether collapsing the switches
    into a table first (`R.014`, as a wip refactor on this branch) is cheaper than adding
    six cases to each. *(Effort 1 plan, Decision 12: `R.014` runs first, as Commit 0.)*

24. **Transport Lockdown's lock-out is a condition kind, not a second asset field.**
    Lean for the Effort 5 plan: the condition shape gains an optional target-faction
    field, and `locked_out{source: the Lockdown, faction: the struck rival}` sits on the
    Lockdown asset like any other condition. One asset-state shape instead of two, and
    the record still lives and dies with its source (D-31, D-36). The Effort 5 plan
    confirms once the Effort 1 condition shape exists. Amends D-36.

## Decision Record — Execution

*(Empty at plan time.)*

---

## Open Questions — To Ratify at Implementation Time

Each is answered by the named effort plan.

| Effort | Question |
|---|---|
| 1 | *(Resolved in the [Effort 1 plan](completed/s-flag-rule-extensions-effort-1-plan.md), Decisions 1–15: condition shape, no manifest mutation (orders carry over), the `Matchup` route, Cracked Comms through the Attack action, effect names, special attackers excluded from Expand Influence, `R.014` first)* |
| 3 | *(Resolved in the [Effort 3 plan](s-flag-rule-extensions-effort-3-plan.md), Decisions 5, 7, 9, 10 and 11: `hooks.PaymentCollector.ConfirmPayment`, a disabled asset is not a Marketers candidate, Seditionists' unpaid roll attaches nothing, a capture into the source's faction lapses the block when read, Monopoly's `levy_coin` and its target list)* |
| 2 | Open Question 1 (Decision 7) |
| 2 | When `sheltered` clears: a reactor on a repair that reaches full HP, or the bookkeeping step |
| 2 | The tie case: today the counter branch runs even when the attack killed the defender (`attack.go:195`), so a sheltered or captured defender would still counter. Decide whether an intercepted defender counters |
| 4 | The Cat 4 permission family's shape (`permission_query`, `require_permission`, `grant_permission`) and the D-41 request method |
| 4 | `PermissionRefused` fields and the wizard's world prompt for a `T-011` pick (Decisions 10 and 22) |
| 5 | Decision 12's mark: Lobbyists' withdrawal only (lean) or every refusal |
| 5 | Transport Lockdown's paid-and-waiting mark on the order (Decision 15) |
| 5 | The Lobbyists contest's prompt to its owner inside the request path |
| 5 | The condition shape's target-faction field for `locked_out` (Decision 24) |
| 5 | D-33 re-confirmed: refusing a relocate outright under Lockdown is the harshest of three readings of "spending 1d4 and waiting one turn" |
| each | Digest and TUI wording for its new mutations (Decision 13) |

## Effort Summary

```
1 Combat foundations ──┬──> 3 A-side abilities ──> 2 Cat 6 interception
                       └──────────────────────────────────────┐
4 Purchase + government ──────────────────────────────────────┴──> 5 Movement permission

Run order: 1 → 3 → 2 → 4 → 5
```

`3 → 2` is a run-order edge, not a code dependency: Effort 2 reuses the pay prompt that
Effort 3 defines (Decision 19). An interim effort review (session-modes section 8) fits
best at the Effort 1 boundary, since Efforts 2, 3 and 5 build on its model.

| # | Effort | Scope | Assets | Plan |
|---|--------|-------|--------|------|
| 1 | Combat foundations | `R.014`, asset conditions (`no_ability`), the condition step, `AssetOwnerChanged` with Decisions 5 and 17 (orders carry over), the matchup route, special attacks | Saboteurs, Treachery, Cracked Comms | [`-effort-1-plan.md`](completed/s-flag-rule-extensions-effort-1-plan.md) |
| 2 | Cat 6 interception | The Cat 6 family at both lethal emitters, `sheltered` in `Active()` and `Targetable()`, `RepairCostModifier` | Boltholes, False Front, Medical Center, Hostile Takeover | own session → `-effort-2-plan.md` |
| 3 | A-side abilities | One ability view per action, `TargetAsset`, the shared pay prompt, `disabled` with D-38's `Targetable()`, `no_attack` | Marketers, Seditionists, Monopoly | [`-effort-3-plan.md`](s-flag-rule-extensions-effort-3-plan.md) |
| 4 | Purchase and government | Stealth as a quality, the government record with its wizard prompt and `T-011` coupling, Planetary Seizure and Inside Enemy Territory, the D-41 request at purchase, enabler and refit, `PermissionRefused`, the Cat 4 permission family | Stealth, Capital Fleet, Popular Movement | own session → `-effort-4-plan.md` |
| 5 | Movement permission | The arrival check and hold (`MovementOrderHeld`), the relocate request, the bypasses, the Lockdown `locked_out` condition | Lobbyists, Transport Lockdown, Deep Strike Landers, Blockade Runners | own session → `-effort-5-plan.md` |

Effort 4 depends on nothing; it runs fourth only so that the work runs in one line.

## Shared Context

**Where each new piece first lands.** Later efforts cite the effort that defines a piece;
they do not redefine it.

| Piece | Kind | Lands in |
|---|---|---|
| `Asset` condition list, apply and clear mutations | domain, mutation | 1 |
| Condition step at bookkeeping start | orchestrator | 1 (Saboteurs' expiry); 3 adds pay-to-lift |
| `AssetOwnerChanged`, the two-pass apply (Decision 5), the all-faction `NextAssetID`, hook re-registration after a capture | mutation, domain, orchestrator | 1 |
| The `Matchup` route on `Registrar` (`MatchupContext`, `MatchupOutcome`) | effects | 1 |
| `AttackProfile.Damage *DiceRoll`, zero-damage hit, no redirect prompt | domain, loader, action | 1 |
| Matchup route (`own_attack`, `own_defense_win`); Attack's effects-engine dependency (Decision 21) | effects phase route, action | 1 |
| Cat 6 family | hooks | 2 |
| `sheltered` in `Active()` and `Targetable()` | domain, actions | 2 |
| `RepairCostModifier` | hooks Cat 4 | 2 |
| One ability view per action (`ability.NewView`) | ability pipeline | 3 |
| `TargetAsset` in `resolveTarget`; per-effect candidate rules | ability pipeline | 3 |
| Pay prompt interface (`hooks.PaymentCollector`) | collectors | 3 |
| `disabled` in `Active()`; the separate defender test `Targetable()` (D-38); `no_attack` in attacker eligibility (`domain.AttackBlocked`) | domain, actions | 3 |
| `quality = true`, `grant_stealth` | loader, effects | 4 |
| `FactionState.Governments`, `WorldGovernmentChanged`, `TagRemoved`, the wizard world prompt | state, mutation, TUI manage | 4 |
| Permission request interface (D-41); `PermissionRefused` | collectors, mutation | 4 |
| Cat 4 permission family | hooks | 4 |
| `TransportProfile.IgnoresPermission` | domain, loader | 5 |
| `MovementOrderHeld`; the transport reactor's `Held` arm | mutation, effects | 5 |
| `locked_out` condition kind with a target-faction field (Decision 24) | domain, mutation | 5 |

**Mutation surface.** The new vocabulary this initiative adds, so the apply switch,
history, digest and TUI grow together (Decision 13):

| Mutation | Changes | Effort |
|---|---|---|
| condition applied, condition cleared (names in the Effort 1 plan) | `Asset` condition list | 1 |
| `AssetOwnerChanged` | moves an asset between faction maps, sets `OwnerID` | 1 |
| `AssetHPDelta` with `Delta: 0` | no new type; a special hit's record | 1 |
| `WorldGovernmentChanged` | `FactionState.Governments[world]` | 4 |
| `TagRemoved` | drops one tag from a faction | 4 |
| `PermissionRefused` | nothing; record only | 4 |
| `MovementOrderHeld` | order step and hex, as `Progressed` | 5 |

**State storage.** New durable fields: `Asset.Conditions` (Effort 1; in `Clone`),
`FactionState.Governments` (Effort 4; `omitempty`), the order's paid-and-waiting mark
(Effort 5 plan). New definition fields: `AttackProfile.Damage` as a pointer (Effort 1),
`quality` (Effort 4), `TransportProfile.IgnoresPermission` (Effort 5). No new file.

**Prompts addressed to a faction that is not acting.** Each names the faction it asks,
as `ConfirmRedirectToBase(defenderFaction, …)` does (`attack.go:167`). One shared driver
answers them all today; `F.020`/`F.025` route them later with no change here.

**No new hard-coded definition IDs.** Every asset in this initiative is reached through
its data (`[[effects]]`, `[ability]`, `[attack]`, a typed profile field). `B.001` is the
cost of the one exception that exists.

**Determinism.** Every fan-out over factions or assets iterates in sorted ID order
(`domain.SortedAssets`, an explicit sort for factions). The condition step walks
factions in ID order, then assets in ID order.

**Test patterns.** As in A.001.3: `seqRoller{values}` for rolls, the gomock
`MockCollector` for prompts. Every gate and filter a registrar accepts gets a test that
proves it suppresses. Every new mutation gets an apply test, a miss test, and a digest
test. Decision 5 gets an `Apply` test with a batch that names the old owner after the
owner change. Each new prompt gets an adapter-side case named in its commit
(Decision 23).

**Model discipline.** Opus for every Plan session, for execution, and for the pre-merge
review.

## Out of Scope

- **H.A.W.K assets** (A.001.5).
- **`R.019`** — the transport bypass is a profile field (D-25), so the migration is not
  needed here.
- **The `"G.012"` half of `B.001`** (Decision 20).
- **A CLI for authoring governments.** The wizard's `T-011` pick is the authoring path
  (Decisions 8 and 10); a load-time derivation for old campaigns is also out.
- **Bribes as a modeled option** (D-34) and **per-faction drivers** (`F.020`, `F.025`).
- **Refit's tech-level check.** `refit_asset.go` defers both checks; this initiative adds
  only the permission request (D-23).
- **Asset-data drift outside the 14** (the discovery's mechanical diff). A candidate
  backlog audit; Robert decides at pre-merge whether to file it.
- **Space Marines' bypass clause** — they are not `P`, so a government cannot stop them.

## Reconciliation Items

Each is fixed by the effort that makes its claim true, or at pre-merge.

| Item | Resolved by |
|---|---|
| `ability-catalog.md` places Marketers, Seditionists and Monopoly (AP-6); its "Monopoly flag-vs-text" open reconciliation | Effort 3 |
| `hooks/doc.go` and `hooks.md` list five categories | Effort 2 — add Cat 6 |
| `hooks.md` lists the Cat 4 family | Efforts 2 and 4 — add `RepairCostModifier` and the permission family |
| `effect-mutation.md` names two phase routes (`turn_start`, `on_purchase`) | Effort 1 — add the matchup route |
| `effect-mutation.md` Key Decision "the orchestrator still owns each call site" | Effort 1 — reword for the action-raised route (Decision 21) |
| `effect-mutation.md` mutation catalogue | Each effort — its new mutations |
| `goals.md` Key Decision cites "even the Planetary Government tag" as a goal mutation | Effort 4 — Seizure emits `WorldGovernmentChanged` beside the tag (Decision 10) |
| `tags.toml` `T-011` effect text, and `swn-faction-mechanics.md` Government Permission ("the tag grants or denies") | Effort 4 — the record is the per-planet instance; the tag marks the holder (Decision 10) |
| `actions.md` Buy Asset row names the stealth target prompt; P-flag claims under "one purchase path" | Effort 4 |
| `persistence.md` says `FactionState` holds the factions and the turn cursor | Effort 4 — add the government record |
| `world-movement.md` movement phase has no arrival check | Effort 5 |
| `domain.HasPlanetaryGovernment`, `PurchaseTerms.HoldsGovernment` | Effort 4 — retired (discovery, Replaces / Retires) |
| `planned-work.md`: `B.001` narrowed, `B.009` closed, `R.014` closed (Effort 1 runs it), `B.011` closed if review confirms (Effort 3 Decision 3 removes its cause), the drift audit and the digest's empty asset names (Effort 1 Out of Scope) if Robert files them | Pre-merge |

---

## Work Breakdown

Every effort carries its Phase/Commit/Task detail in its own plan file. The outlines
below are **scoping guidance for those Plan sessions**, not executable tasks. Each effort
plan re-grounds against the code as it stands then, and may split or merge these
commits.

Each effort below opens with an **In Plain Words** block. Each per-effort plan file
also opens with its own In Plain Words section (template rule), scoped to that effort.
It starts from the block here. When the effort plan changes the what, why or fit, it
also updates the block here, so the two never disagree.

### Effort 1 — Combat foundations *(own Plan session)*

**In Plain Words**

- **What:** teaches combat three new things. An asset can carry a lasting mark (a
  "condition") that stops it doing something. An asset can change sides and keep its
  identity, its orders and anything riding on it. And an attack can "hit" without doing
  damage, so the hit triggers a special rule instead. It also lets an asset's rule fire
  at each attack, hit or miss. Three assets use these: Saboteurs (a target cannot use
  its ability until the Saboteurs' owner's next turn), Treachery (a hit steals the
  target) and Cracked Comms (a failed attack on it can turn the attacker on itself).
- **Why:** every later combat asset in this initiative builds on these pieces, so they
  come first.
- **How it fits:** the first effort. Efforts 2 and 3 use its conditions and its change
  of owner; Effort 5 uses its no-damage attack.

Discovery: State Held on an Asset, Change of Asset Owner, Special Attacks and the
Matchup Route, Damage Redirected into Combat.

The [Effort 1 plan](completed/s-flag-rule-extensions-effort-1-plan.md) carries the spec. Its
commits:

- `refactor: AskKind dispatch table` — `R.014` first (Decision 23).
- `feat: asset conditions` — `Condition{Kind, SourceAssetID, ExpiresAtTurnOf}` on
  `Asset`; `AssetConditionApplied` / `Cleared`; `no_ability` read by Use Asset Ability.
- `feat: condition step at turn start` — `turn.ExpireConditions` in the turn-start batch
  (Decision 4).
- `feat: change of asset owner` — `AssetOwnerChanged` (keeps the ID, D-9), the two-pass
  apply (Decision 5), the all-faction `NextAssetID`, orders carry over, riders keep their
  owner, hooks re-register after a capture, the Planetary Seizure fix (Decision 17,
  extended).
- `feat: special attacks and Treachery` — `Damage *DiceRoll` (Decision 3); a zero-damage
  hit with no redirect (D-15); `capture_target`.
- `feat: matchup route and Saboteurs` — the `Matchup` route raised by the Attack action
  (D-13, D-17, Decision 21); `block_ability`.
- `feat: Cracked Comms self-attack` — `counterstrike`; `ConfirmCounterstrike`; the
  self-roll with `Opponent: nil` (Decision 9).

### Effort 3 — A-side abilities *(own Plan session)*

**In Plain Words**

- **What:** three assets whose ability a faction spends its action on. Marketers make a
  rival pay half an asset's price or lose the use of that asset until it pays.
  Seditionists stop one enemy asset on their world from attacking. Monopoly makes a
  rival pay 1 Coin, or lose an asset of its choice if it cannot pay. It also fixes an
  older fault: when one action uses several assets, each now sees what the earlier
  ones did.
- **Why:** abilities today can target only a whole faction, not one asset, and nothing
  can ask a faction "pay this, or else". All three assets need one or both.
- **How it fits:** runs after Effort 1, whose condition model it extends with "disabled"
  and "cannot attack". The pay question it adds is reused by Medical Center in Effort 2.

Discovery: State Held on an Asset, Prompt to a Faction That Is Not Acting.

The [Effort 3 plan](s-flag-rule-extensions-effort-3-plan.md) carries the spec. Its
commits:

- `fix: abilities in one action see earlier assets' results` — one `ability.NewView`
  per Use Asset Ability; rolls spend budgets on the real factions.
- `feat: asset targets and the pay prompt` — `TargetAsset` (Decision 11) with
  per-effect candidate rules; an ability with no target is not offered;
  `hooks.PaymentCollector` (Decision 19).
- `feat: Marketers` — `disable_asset`; `disabled` read by `Active()`, with
  `Targetable()` for D-38; no counter from a disabled defender; the pay-to-lift prompt
  at turn start (D-12).
- `feat: Seditionists` — `block_attack`; `no_attack` read through
  `domain.AttackBlocked` (source exists, same world, another owner).
- `feat: Monopoly` — `levy_coin`: a target faction with unstealthed assets on the
  world pays the owner 1 Coin, or loses one asset of its choice there; the TOML
  description from the book.

### Effort 2 — Cat 6 interception *(own Plan session)*

**In Plain Words**

- **What:** a new step that runs just before damage destroys an asset, and can replace
  the destruction with something else. Boltholes hide the asset at 0 HP until it is
  repaired. False Front is destroyed in its place. Medical Center brings it back at
  1 HP for a price. Hostile Takeover takes it over at 1 HP.
- **Why:** the engine records a destroyed asset as soon as the damage lands, and other
  rules react to that record (Scavengers' Coin, goal progress, the narrative). These
  four assets say the asset was *not* destroyed, so the change must happen before the
  record is written, not after.
- **How it fits:** runs after Efforts 1 and 3. It uses Effort 1's conditions and change
  of owner, and Effort 3's pay question. It is the most design-heavy effort, so its plan
  settles the step's exact shape.

Discovery: Lethal Damage Interception (Cat 6). Open Question 1 (Decision 7) is the
centre of this plan.

- `feat: cat 6 lethal interception` — the family, its registry buckets, dispatch at
  `applyAssetDamage` and in the `self_damage` reactor (automatic only, D-10), the D-7
  walk order, and the tracker outcome.
- `feat: Boltholes` — `sheltered`; `Active()` and `Targetable()` read it (the separate
  defender test lands in Effort 3, Decision 6); destroyed with its Boltholes; cleared
  at full HP.
- `feat: False Front` — elective sacrifice.
- `feat: Medical Center` — elective restore at 1 HP, once between turns; reuses the pay
  prompt; `RepairCostModifier` and Repair Asset's registry (Decision 18).
- `feat: Hostile Takeover` — the attacker's interceptor: 1 HP and `AssetOwnerChanged`.

### Effort 4 — Purchase and government *(own Plan session)*

**In Plain Words**

- **What:** records who governs each world, in the campaign's saved state, and makes
  every purchase of a restricted asset ask that government's permission (the GM answers
  for a world no faction governs). Popular Movement makes its world always say yes.
  Capital Fleet needs permission to buy. Stealth stops being a hard-coded special case
  and becomes data: a quality bought for another asset.
- **Why:** today "permission" means "owns a Base on the world". Several factions can
  pass that test at once, and nothing ever grants or refuses anything, which disagrees
  with the book. Planetary Seizure also never records which world was seized.
- **How it fits:** depends on no other effort. Effort 5 extends the same permission
  request from purchases to movement.

Discovery: Purchase That Targets Another Asset, Permission Model (government and
purchase rows).

- `refactor: Stealth as a purchase quality` — `quality = true`, `grant_stealth`, the
  fourth direct definition read in Buy Asset; every `SWN-C3-002` check goes (D-19 to
  D-22).
- `feat: world government record` — `FactionState.Governments` (Decision 8),
  `WorldGovernmentChanged` and `TagRemoved`, Planetary Seizure emitting both, Inside
  Enemy Territory on the record, the wizard's world prompt on a `T-011` pick, and the
  `tags.toml` and rules-doc retext (Decision 10).
- `feat: permission request at purchase` — the D-41 interface in both collectors and the
  TUI; the request at Buy Asset, the purchase enabler and Refit (D-23, D-35); the
  automatic grant on a world the faction rules (D-6); `PermissionRefused` on a refusal
  (Decision 22); retire `HasPlanetaryGovernment`.
- `feat: permission rule queries` — the Cat 4 permission family; Capital Fleet's
  `require_permission` (D-26, closes `B.009`); Popular Movement's `grant_permission`.

### Effort 5 — Movement permission *(own Plan session)*

**In Plain Words**

- **What:** a restricted asset that arrives at a world, or jumps there by an ability,
  asks that world's government for permission. A refused arrival waits in orbit and asks
  again next turn. Deep Strike Landers, Blockade Runners and stealthed assets skip the
  question. Lobbyists can take back a permission given to a rival. Transport Lockdown
  makes a rival pay and wait one turn for every arrival on its world.
- **Why:** the book's permission rule covers moving onto a world, not only buying there,
  and five assets bend that rule. None of them works until arrivals ask.
- **How it fits:** the last effort. It needs Effort 4's government record and permission
  request, and Effort 1's no-damage attack (Transport Lockdown strikes with one).

Discovery: Permission Model (movement rows and "What bends it").

- `feat: arrival permission and hold` — the request between tick and apply (D-24); the
  hold at the edge as `MovementOrderHeld` (D-27, Decisions 16 and 22), the whole order
  with its cargo through the reactor's `Held` arm (D-28).
- `feat: relocate permission` — the request at the destination choice; the per-turn
  mark (Decision 12).
- `feat: permission bypasses` — `TransportProfile.IgnoresPermission` on Deep Strike
  Landers and Blockade Runners; the stealthed-`P` rule (D-25).
- `feat: Lobbyists` — the contest after any grant to a rival on its world (D-29, D-30).
- `feat: Transport Lockdown` — its `[attack]` line from the book; the `locked_out`
  condition on the Lockdown asset (D-36, Decision 24), read while it is active there
  (D-31); 1d4 and a one-turn hold on drift arrivals, refusal on relocates (D-32, D-33,
  Decision 15).
