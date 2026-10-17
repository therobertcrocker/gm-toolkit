# S-flag Effect Registry (A.001.3) — Implementation Plan (Overview)

## Context / Goal

Build the data-driven S-side model the
[discovery](../discovery/s-flag-effect-registry-discovery.md) designed —
*trigger → filters → gates → effect*, keyed on the effect name — and ship every S-flag
effect that is a hook or a rule query. The engine registers effects from an asset's
`[[effects]]` list; each registrar reads its typed parameters and decides which hooks
its effect needs. The hook category is never named in the data.

The hook framework is complete in its categories and incomplete in its reach, because
transport is an attacker-side, own-turn mechanic that exercises none of it. This
initiative extends the dispatch and then authors 19 assets against it.

- Arc: [data-driven-rulebook](../arcs/data-driven-rulebook/data-driven-rulebook-arc-plan.md),
  initiative A.001.3 (boundary reset by AP-7). Cross-cutting ACD-1…ACD-9 bind here
  without re-litigation.
- Mechanical input: the [ability & effect catalog](../../architecture/engine/ability-catalog.md) (ACD-8).
- Consumes A.001.2's relocate primitive (ACD-4) for the free harness only.
- This is a **multi-effort** initiative (session-modes section 8). This file is the
  overview plan. Efforts 1, 3 and 5 get their own Plan session; Efforts 2 and 4 execute
  from this file.

## Decisions Ratified in Planning

1. **Five-effort slicing by dispatch surface. DAG `1 → {2, 3, 4, 5}`.** Effort 1 builds
   the model and gates everything. Efforts 2–5 each own one dispatch surface and the
   effects that ride it; they depend only on Effort 1 and are mutually independent.
   Resolves discovery Open Question 6.

2. **The shared relocate code is a new `relocate` package taking a narrow collector
   interface.** One copy of enumerate → filter → prompt → build serves both the
   action-spent harness (`ability/relocate.go`) and the free harness (the movement
   phase). A one-method `OrderCollector` interface lets the action collector and the
   phase collector both satisfy it, so the prompt is shared too, not just the pure
   parts. Rejected: exporting from `ability` and importing it in the orchestrator, which
   would make the engine depend on an action sub-package. Resolves Open Question 1.

3. **`once_per` budgets always live on the acting faction's `HookBudgets`, namespaced by
   target.** `faction.HookBudgets = nil` is cleared in that faction's own bookkeeping
   (`bookkeeping.go:46`), so a budget written to another faction's map resets at an
   unpredictable point in the cycle. Keying `effect:<name>:<defID>:<targetFactionID>` on
   the owner's map keeps every budget's reset cadence tied to its owner's turn, and
   matches the rules wording — "no matter how many Franchises attack" caps the
   attacker's repeats against one target. Resolves Open Question 5. *Key format
   amended by the Effort 3 plan, Decision 16:* the key carries the trigger,
   `effect:<name>:<trigger>:<defID>[:<target-or-asset>]`, so two entries of one
   definition (Book of Secrets) do not share a budget.

4. **`RelocateParams` grows `CoinCostMode`.** Transit Web reads "For 1 Coin, any number
   of assets may be moved", but `relocate.go:67` charges `CoinCost` per moved asset. The
   mode is `per_asset` (the zero value, so A-side assets are untouched) or `flat`.

5. **`turn_start` is a third dispatch route, and the discovery understates the work.**
   *Corrects the discovery's split table, which records "none" for Party Machine.*
   Verified this session: the bookkeeping phase makes no `dispatch.MutationReactors`
   call. The three Cat 3 call sites are the movement tick (`orchestrator.go:315`), the
   movement decisions (`:351`) and the action phase (`:447`). No dispatch point exists
   for a turn-start event at all. Party Machine and Panopticon Matrix both need one.
   A turn-start effect is a phase event, not a reaction to a mutation, so a Cat 3
   reactor is the wrong shape — the route is a direct definition read, the same shape
   D-5 already accepted for `on_purchase`. D-5 therefore reads: the S side has **hooks
   registered at cycle start, plus a direct definition read at phase events** (turn
   start, purchase). The read fires at the start of the bookkeeping phase, gated on
   `BookkeepingApplied`, not in the turn setup: setup re-runs on every resumed turn and
   would pay twice (Effort 1 plan, Decision 5).

6. **One composed effects handler per asset definition.** `effect.Handler` is keyed by
   asset definition ID (`effect.go:27`) and `Register` overwrites silently. Panopticon
   Matrix declares two effects and Zealots declares two, so a per-effect handler would
   lose all but the last. `EffectsListHandler` holds a def's whole `[[effects]]` list and
   registers each entry through its registrar. A def that carries both `Transport` and
   `Effects` must fail rather than let one handler overwrite the other; the guard is
   general — `EffectsEngine.Register` rejects any second handler for one definition
   (Effort 1 plan, Decision 8). No SWN def collides today; the guard stops it becoming a
   silent bug.

7. **Loud-failure rule 4 runs at engine construction, and `NewWithRulebook` grows an
   error return.** Rules 1–3 (unknown names, parameter block not matching the declared
   effect, missing `owner`/`scope`) are pure data and stay in the loader. Rule 4 — the
   registrar does not accept this gate or filter — needs both the data and the registrar
   table, and the registrars live in `engine/effect/effects`, which `rulebook` must not
   import. Engine construction is the first point that knows both. Four callers change
   (`cmd/gm-toolkit/faction.go:80`, `testharness/harness.go:98`,
   `scenarios/spot_check_test.go:38`, `tui/dryrun.go:29`). The error return lands in the
   same commit as the registrar contract, because the duplicate-handler guard needs it
   too (Effort 1 plan, Decision 9).

8. **Cat 3 gains a budget path.** `HookBudgets` is read and written only inside
   `RollWithHooks` (`roll.go:42-48`, `:78-91`). Franchise, Blockade Fleet, Bank and Book
   of Secrets all carry `once_per`, and the first three are Cat 3. Effort 2 adds a gate
   helper that reads and writes `HookBudgets` outside the roll path.

9. **Damage mutations record the causing asset.** `AssetHPDelta` carries
   `CausedByFactionID` but no causing asset ID (`mutation.go:85-91`), so a Franchise
   reactor cannot tell whether *it* landed the hit that a mutation batch describes.
   Effort 2 adds `CausedByAssetID string` and fills it in `applyAssetDamage`
   (`attack.go:209`). Without it, `own_attack_hit` cannot be filtered to the declaring
   asset.

10. **`ignore_coin_loss` emits a compensating gain, it does not suppress the loss.**
    `MutationReactor.OnMutations` returns mutations to append; it has no power to cancel
    one (`interfaces.go:25-27`). Bank emits `+CoinDelta` with cause `bank_absorbed`
    against the loss it matched. The history then shows both the loss and the
    absorption, which is the honest record of what happened at the table.

11. **`coin_steal` and `self_damage` take a dice string, not an integer.** Blockade Fleet
    steals `1d4`; Franchise steals a flat `1`; Zealots takes `1d4`. A fixed amount is the
    string `"1"`. `parseDice` rejects a bare integer, so amounts parse through
    `parseAmount`, which maps `"1"` to `DiceRoll{Modifier: 1}` and sends dice notation to
    `parseDice` (Effort 1 plan, Decision 15). The discovery's example (`amount = 1`) is
    superseded.

12. **Lawyers filters on `category`, not `asset_type`.** "Cannot attack or counterattack
    **Force** assets" — Force is a `FactionStat`, not an `AssetType`. The discovery's
    filter table files Lawyers under `asset_type`. Planetary Defenses is genuinely
    `asset_type` (Starship) and is unaffected.

13. **The domain shape is the contract Effort 1 builds.** Direction ratified below (see
    Shared Context); the literal Go lands in Effort 1's first task. Efforts 2–5 bind to
    it and re-ground against the built API at execution time.

14. **Integral Protocols ships across two efforts.** Its `bonus_die` lands in Effort 3
    and its `restrict_attack` in Effort 4. The discovery's "ships whole" constraint is
    about not splitting it across two *initiatives* (A.001.3 vs A.001.6); efforts share
    one branch, so the asset is whole at merge.

15. **A.001.2's Planning Decision 5 prose cleanup folds in.** Blockade Fleet and
    Surveyors both still carry "As an action, [it] may move itself to a world within N
    hexes" with no `A` flag and no `[ability]` block. Both have `speed` set, so this is
    the misleading Speed-based prose A.001.2 D5 identified and named Blockade Fleet for;
    the cleanup never ran. It is data-only description editing with no mechanic change.
    Fold it into whichever effort touches each file (Effort 2 for Blockade Fleet,
    Effort 3 for Surveyors). These are **not** missing-`A`-flag data errors.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

1. **Removing `MutationReactorsFor` moves three test files to the new accessors**
   (Effort 2, Commit 1). The plan listed only `mutations.go` as a caller. The four
   Cat 3 tests in `hooks/registry_test.go` and the transport registration test in
   `effect/effect_test.go` also called it. The registry tests now check each accessor
   alone, and a new test proves an asset-scope hook stays out of other assets and out of
   the faction lookup. The transport test now finds its reactor at asset scope and none
   at faction scope. The existing `mutations_test.go` cases passed a `nil` state, which
   `allReactors` dereferences; they now build one with a `stateOf` helper. Production
   code gets no nil guard, because every orchestrator call site has a state. Plan body
   reconciled: Effort 2 Commit 1 Task 6.

2. **Effort 2 Commit 1 carries its own arch-doc fixes** (Effort 2, Commit 1). The
   Reconciliation Items table assigns the `hooks.md` / `effect-mutation.md` scope claims
   to Effort 2 without naming a commit. This commit is the one that makes them true or
   false, so it fixes them: `hooks.md` notes that Cat 3 exposes its buckets separately
   and that `MutationReactors` runs every faction's and asset's reactors in sorted
   order; `effect-mutation.md` says the `TransportReactor` is asset-scoped. The
   `hooks.md:74-75` claim ("`EffectsEngine` registers asset-scoped hooks") becomes true
   and is unchanged. Plan body reconciled: Effort 2 Commit 1 Task 7.

3. **A Cat 3 reactor gets its roller at registration** (Effort 2, Commit 3).
   `MutationReactor.OnMutations` takes no `Roller`, but `coin_steal`, `self_damage` and
   the contest gate roll dice. `ApplyAll` takes the engine roller and
   `RegisterContext` carries it; each reactor keeps it. `ApplyAll` runs at the start of
   each cycle, so a test that swaps `Engine.Rand` before the run still controls the
   dice. Rejected: a `Roller` parameter on `OnMutations`, which changes every reactor
   for three effects. Plan body reconciled: Effort 2 Commit 3.

4. **Damage mutations mark a counterattack** (Effort 2, Commit 3). Attack damage and
   counterattack damage both emit `Cause: "attack"` with the dealing asset in
   `CausedByAssetID`, so a reactor cannot tell them apart. Blockade Fleet and Franchise
   would steal on a counterattack, and Zealots' two entries would both fire on every
   hit (2d4, not 1d4). `AssetHPDelta` grows `Counter bool` (`omitempty`).
   `BaseHPDelta` does not: a counterattack always strikes the attacking asset, and only
   attack damage can be redirected to a Base. Rejected: a `"counterattack"` cause, which changes every reader of
   `Cause == "attack"` (digest, Scavengers, goals); and passing the acting faction to
   `OnMutations`, which changes every reactor.

5. **A hit redirected to a Base counts as a hit** (Effort 2, Commit 3). When the
   defender redirects damage to its Base, the attack emits `BaseHPDelta` and no
   `AssetHPDelta`, so `own_attack_hit` would miss a successful attack. `BaseHPDelta`
   grows `CausedByAssetID`, and the attack-hit match reads both mutation types. This
   completes Commit 2, which covered `AssetHPDelta` only.

6. **A Bank absorbing a steal does not undo the thief's gain** (Effort 2, Commit 5).
   Robert's rules reading: Bank offsets its owner's loss only. The `coin_steal` gain to
   the thief stands, so net Coin rises by the absorbed amount.

7. **Tripwire Cells react to movement only; the purchase case moves to Effort 5**
   (Effort 2, Commit 6). No purchase creates a stealthed asset until Effort 5's
   `apply_stealth` (Psychic Assassins). If that route emits stealth as a separate
   `AssetStealthApplied`, a match on `AssetAdded.Asset.Stealthy` never fires. Effort 5
   wires the purchase case together with the producer, so the match and the mutation
   shape are designed together. Plan body reconciled: Effort 2 Commit 6; Effort 5
   outline.

8. **`self_damage` judges lethality against the batch, not the state** (Effort 2,
   Commit 4). A reactor sees state before the batch applies, but on a tie the batch
   already holds the defender's counter damage to the Zealots. The reactor adds the
   batch's `AssetHPDelta`s for the asset to its `CurrentHP` before it decides whether to
   emit `AssetRemoved` — the reactor-side twin of `applyAssetDamage`'s tracker. When the
   batch already removes the Zealots, the reactor emits nothing: an `AssetHPDelta` on a
   removed asset is a mutation miss at apply time.

9. **The lowest-ID Bank allocates losses for every Bank its faction owns** (Effort 2,
   Commit 5). Each Bank's reactor sees the same batch, and siblings cannot see each
   other's output in one dispatch round. With two Banks and two imposed losses, both
   would absorb the first loss and none the second. Every Bank still registers a
   reactor, but only the owner's lowest-ID Bank acts: it walks the losses in batch order
   and assigns each to the next Bank, in ID order, whose `once_per` budget is unspent.
   Budgets stay per asset (Planning Decision 3). Rejected: one faction-scope reactor,
   which needs a register-once guard in a registrar that runs once per asset.

10. **"Lands" means arrives from elsewhere, and the contest gate honours both rollers**
    (Effort 2, Commit 6). Tripwire Cells match an `AssetMoved` whose `ToLocation` is
    their world and whose `FromLocation` is not — a move into flight or within the
    world is not a landing; a landing from flight is. The contest helper
    (`effects/contest.go`) runs `roller = "target"` as well as `"self"`, with ties to
    the defender as in the A-side gate: a registrar that accepts a gate must honour
    every value of it, or wrong data runs silently. Panopticon Matrix (Effort 3) is the
    first `target` user and reuses the helper.

11. **Every S-side reactor checks that its asset is active when it reacts** (Effort 2,
    follow-up fix). The rules make a new asset unable to "use special abilities until
    the start of next turn" and an unmaintained asset "unusable". `active()` in
    `effects` is `Ready && Maintained && CurrentHP > 0`, matching attack eligibility.
    The check runs at reaction time, not at registration: `ApplyAll` runs at cycle
    start, but `Ready` flips at the owner's turn start and `Maintained` in each
    bookkeeping, so a Bank bought last cycle is registered before it is ready. Only Bank
    and Tripwire Cells had a live gap — attack hits already require an eligible
    attacker or defender — but all four reactors check, so no future passive effect can
    forget it. Bank's lead and allocation use active Banks only, so an inactive
    lowest-ID Bank hands the lead on instead of absorbing nothing for everyone. The
    inline copies outside the effects layer stay as they are (`R.020`). Effort 3–5
    reactors and hooks follow the same rule.

12. **Effort 4 runs as a design pass and two execution tracks** (Effort 4, start).
    Three commits waited on open decisions, and one decision changed Commit 1's scope.
    One design pass settled entries 13–16 first. Execution then runs Commits 1–3 (cost
    and tech level) in one session and Commits 4–5 (attack eligibility) in another. The
    two tracks share no code. Each commit stays its own unit. Rejected: the whole effort
    in one session, which reaches the open decisions late and changes built attack code
    at the end of a long context.

13. **`tech_level_floor` lifts the owner's purchases only** (Effort 4, start). Robert's
    rules reading: Laboratory and Pretech Researchers serve their owner, like R&D
    Department. All three entries keep `owner = "own"`. Resolves the Open Question on who
    benefits.

14. **`ResolveWorldTechLevel` walks the buyer's own assets, not the world's occupants**
    (Effort 4, start). The plan said owner-only needs no lookup change. That is false:
    every effect registrar registers at `AssetScope`, and
    `WorldTechLevelModifiersFor(faction.ID, "")` skips the asset bucket
    (`registry.go:191`). The resolver gains a tail that walks `buyer.Assets` in ID order
    and calls `WorldTechLevelModifiersForAsset`. The hook checks its own scope filter
    (`world`: its asset stands on the purchase world; `faction`: any world) and
    `active()`. One path serves Laboratory, Pretech Researchers and R&D Department, with
    no signature change and no index. Rejected: registration at `FactionScope`, which
    breaks the rule that every effect registers at asset scope. Plan body reconciled:
    Effort 4 Commit 1.

15. **`no_stack` rides on a query struct** (Effort 4, start). `ModifyAssetCost` passes
    only an `int` down the chain, so a mark has nowhere to go. The hook now takes
    `*hooks.AssetCostQuery` with `Cost int` and `Applied map[string]bool`. A `no_stack`
    registrar checks and sets its effect name in `Applied`. Preceptor Archive
    (`tags/preceptor_archive.go`) is the only implementer and moves to the new
    signature. Rejected: a lowest-ID lead like Bank (Decision 9), which needs world
    occupancy that the hook does not get. Plan body reconciled: Effort 4 Commit 2.

16. **Attack eligibility is one Cat 4 interface with a side** (Effort 4, start). One
    query carries the attacker, the defender and `side` (`attack`, `counter` or
    `defense`). One map, one `Register*`, one accessor and one `Resolve*`, consulted at
    all three sites. Lawyers covers `attack` and `counter` in one entry. Rejected: two
    typed interfaces, which double the registry plumbing and still need a counter flag.
    Resolves the Open Question on one interface or two. Plan body reconciled: Effort 4
    Commit 4.

17. **The tech-level hook sees the definition, and resolves per definition** (Effort 4,
    start). Pretech Researchers and R&D Department filter on `category`, but
    `ModifyWorldTechLevel` passes no definition, and `purchasableDefinitions` resolves
    one `techCap` per world before its def loop (`buy_asset.go:172`). Both
    `ModifyWorldTechLevel` and `ResolveWorldTechLevel` gain
    `def *domain.AssetDefinition`, and the resolve call moves into the def loop. No
    implementer exists, so nothing migrates. This matches `ModifyAssetCost`, which
    already takes the definition. Rejected: a hook that returns a floor per category,
    a shape that fits one effect only. Plan body reconciled: Effort 4 Commit 1.

18. **The tech-level chain uses `domain.SortedAssets`** (Effort 4, Commit 1). The plan
    collected `faction.Assets` with `maps.Values` and sorted it with a new helper.
    `domain.SortedAssets` (`faction.go:259`) already returns a faction's assets in ID
    order, and the Shared Context names it as the sort to use. `sortedByID` stays for the
    world index slice, and it uses `slices.SortFunc` with `strings.Compare`, the same
    form as `SortedAssets`, not `sort.Slice`. Re-ground for Task 4:
    `purchasableDefinitions` is the only reader of the world tech cap on the purchase
    path. `Resolve` does not check tech level again. The purchase-enabler cap
    (`purchase_enabler.go:77`) comes from the ability's own `MaxTechLevel`, so it does
    not get the per-definition call. Plan body reconciled: Effort 4 Commit 1 Tasks 1
    and 3.

19. **A Cat 4 hook reads its asset through its owner** (Effort 4, Commit 2). A Cat 4
    hook gets no `FactionState`, so it cannot call `liveAsset`. The hook keeps the
    owner's `*domain.Faction` from `RegisterContext` and reads `owner.Assets[assetID]`
    at query time. `ApplyAll` passes the live `factionState.Factions` pointer, and
    faction pointers do not change in a cycle. Mutation apply edits assets in place, so
    `Ready`, `Maintained`, HP and location stay current. `AssetRemoved` only deletes the
    map entry (`mutation.go:63`), so a removed asset reads as `nil` and `active()`
    fails. `tech_level_floor` (Commit 3) uses the same read. Rejected: holding the
    asset pointer, which keeps a removed Local Investments charging until the next
    cycle; and `FactionState` on every Cat 4 signature, which changes Commit 1's
    resolvers and every call site for one read. The signature change also moves three
    test stubs (`dispatch/rules_test.go`, `hooks/registry_test.go`,
    `actions/buy_asset_test.go`) beside Preceptor Archive. Plan body reconciled:
    Effort 4 Commit 2.

20. **A test builds the engine from the shipped rulebook** (Effort 4, Commit 3). No
    test ran the authored `[[effects]]` entries through `validateEntry`. The loader
    tests parse the TOML, and the construction tests use hand-built definitions. An
    entry that loads but names an owner, scope, filter or gate its registrar does not
    accept failed only at runtime. `TestNewWithRulebook_ShippedRulebook` (`core_test.go`)
    loads `rulebooks/swn` and calls `NewWithRulebook`. A one-off test in Commit 2 found
    the gap, and Robert added it to this commit, which authors three more entries. A
    mutation check (Laboratory `owner = "rival"`) proves that it fails and names the
    entry. In the same commit, `TestPurchasableDefinitions_TechCapResolvesPerDefinition`
    (`buy_asset_test.go`) covers Commit 1's per-definition loop. On a TL5 world, a
    Wealth-only floor on the buyer's own asset makes a TL6 Wealth def purchasable but
    not a TL6 Force def. One cap per world cannot give that result. Plan body
    reconciled: Effort 4 Commit 3.

21. **The eligibility query names the striker, and a counter asks only the striker**
    (Effort 4, Commit 4). Robert's call. `Attacker` is the asset that deals the
    damage, so on a counterattack the original defender is `Attacker` and Lawyers
    always finds its target in `Defender`. On `attack`, the striker's hooks answer
    and then the target's asset hooks answer as `defense`. On `counter`, the target
    is not asked: a counter is not a defense, and asking it would let Integral
    Protocols block counter damage to itself. The striker is asked through
    `AttackEligibilityRulesFor` (global, faction, asset) and the target through
    `AttackEligibilityRulesForAsset`, so a global rule is asked once. The query
    carries both definitions, because Planetary Defenses, Lawyers and Integral
    Protocols filter on definition fields and a Cat 4 hook has no rulebook. So
    `eligibleDefendersAtHex` takes the attacker asset, its definition, the rulebook
    and the registry. A defender with an unknown definition stays eligible, so
    `Resolve` still reports the error. `Resolve` picks defenders through the same
    function, so one check covers `Validate`, `Inputs` and the defender choice.
    `liveDefendersOnWorld` has no caller and does not get the check. Rejected:
    fixed roles, where each hook must find "the other asset" from its own ID. Plan
    body reconciled: Effort 4 Commit 4.

22. **`R.020` lands on this branch, before Commit 5** (Effort 4, after Commit 4).
    Robert's call. Its trigger, "next eligibility change", fired in Commit 4, and
    Commit 5 adds one more `active()` reader on the attack path. The rule moves to
    `domain.Asset.Active()`: `Ready && Maintained && CurrentHP > 0`. The method does
    not check for nil. `effects.active()` stays as the nil guard over it, because
    only a hook reads its asset through its owner and can see a removed asset as nil
    (Decision Record 19). `eligibleAttackers`, `filterOpposingDefenders`,
    `rivalAssetsOnWorld`, `UseAssetAbility` and `FireTurnStart` call the method. The
    last two gain the HP check, which also closes `B.013`. A lethal hit removes the
    asset in the same batch, so no live state changes; test fixtures that left
    `CurrentHP` at zero now set it. Own `refactor:` commit, so Commit 5 stays one
    unit. Rejected: a nil-safe `Active()` with `effects.active()` deleted, which
    makes every caller trust a nil receiver for a case only the effects layer has.

23. **Every purchase path prices through the cost chain, and a discount keeps the
    surcharge** (Effort 4, end-of-effort review). Robert's calls. Before
    `cost_surcharge`, every cost hook lowered the price, so code could check if a
    buyer can pay against `def.Cost`. The surcharge broke that in three places:
    - The Buy Asset menu (`purchasableDefinitions`) offered a def that `Resolve`
      then refused. It now checks against `ResolveAssetCost(...).Cost`.
    - `purchase_enabler` skipped the chain. It now runs the full chain on its 150%
      price, so Local Investments adds +1 and Preceptor Archive now also takes
      −1 on a TL4+ Pretech purchase. `StepContext` gains `Registry`.
    - A pending discount floored at `min(def.Cost/2, cost)` and could remove the
      surcharge. `AssetCostQuery` gains `Surcharge`, and `ResolveAssetCost` returns
      the query. `dispatch.DiscountedCost` floors the price before the surcharge at
      half base, then adds the surcharge back. Both purchase paths share it.

    In the same review, `Registrar` gains an optional `Validate func(EffectDefinition)
    error`, and `validateEntry` calls it last. `restrict_attack` uses it to refuse
    `attribute` with `side = attack_and_counter`. On that side the striker is the
    declaring asset, so the filter would test the asset's own profile. Rejected:
    taking the surcharge as the change made by the rival-asset loop, which counts
    any rival hook as a surcharge; and a pass that adds only the rival surcharge
    for Pretech, which needs a second resolver. Plan body reconciled: Effort 4
    Commit 2.

---

## Open Questions — To Ratify at Implementation Time

- ~~**Does `restrict_attack` need one Cat 4 interface or two?**~~ **Resolved**: one
  interface with a `side` parameter. Decision Record 16.
- ~~**Who benefits from a world-scoped `tech_level_floor`?**~~ **Resolved**: the owner
  only. Decision Record 13; the lookup that follows is Decision Record 14.
- ~~**Panopticon Matrix — "at the beginning of every turn".**~~ **Resolved** in the
  Effort 3 plan, Decision 1: the owner's turn. Discovery Open Question 3.
- ~~**Book of Secrets — "force an enemy faction to reroll one die".**~~ **Resolved** in
  the Effort 3 plan, Decision 2: only a roll made against the owner. Discovery Open
  Question 4.
- **Filters serving both readings** — `filters` currently answers both "which events fire
  this" and "which assets the effect acts on". Every asset in the SWN set makes those the
  same. Deferred until an asset separates them (discovery Open Question 2), recorded so
  the answer is not invented silently.

Discovery Open Question 7 (Capital Fleet and Mercenaries, and whether a missing `P` flag
is a data error) belongs to A.001.6's permission work and is not carried here.

The registrar's accepted-gate / accepted-filter declaration shape was resolved in
Effort 1's Plan session: a spec struct with accepted-name slices and two route
functions (Effort 1 plan, Decision 1).

## Effort Summary

DAG: Effort 1 is foundational. Efforts 2, 3, 4 and 5 depend only on Effort 1 and are
mutually independent. An interim effort review (session-modes section 8) is worth
considering at the Effort 1 boundary, since all four later efforts build on its API.

| # | Effort | Scope | Assets | Plan |
|---|--------|-------|--------|------|
| 1 | Foundation | Domain model, loader with rules 1–3, registrar contract, composed def-handler, rule-4 validation at engine construction, the `turn_start` route. Proof: `coin_gain`. | 1 | own session → `-effort-1-plan.md` |
| 2 | Cat 3 reach | Dispatch items 1 and 2, the causing-asset mutation field, the Cat 3 budget gate. `coin_steal`, `self_damage`, `ignore_coin_loss`, `clear_stealth`. | 5 | this file |
| 3 | Roll hooks | Dispatch items 3 and 7. `bonus_die`, `remove_tag_dice`, `reroll`, `clear_stealth` (Panopticon). | 5 | own session → `-effort-3-plan.md` |
| 4 | Rule queries | Dispatch items 4 and 8. `tech_level_floor`, `cost_surcharge`, `restrict_attack`. | 7 | this file |
| 5 | Movement and purchase | Dispatch items 5 and 6, the shared `relocate` package, `MovementAbility` reshape. `grant_movement`, `apply_stealth`. | 2 | own session → `-effort-5-plan.md` |

Efforts 1, 3 and 5 are design-heavy. Effort 1 sets the contract every later effort binds
to. Effort 3 is the only effort that reaches outside the S side — item 7 routes the
Expand Influence rolls through `RollWithHooks` and changes `RollContext`, so it can
destabilise built behaviour. Effort 5 builds a new package and reshapes a hook type.

Efforts 2 and 4 are dispatch plumbing plus handler authoring on an established pattern,
and are specified at commit granularity below. Their dispatch commits (2.1, 2.2, 4.1,
4.4) are independent of Effort 1's API and are written as literal tasks. Their effect
commits bind to Effort 1's registrar contract, so each of those execution sessions
re-grounds against the built API first (session-modes section 10).

## Shared Context

**The four-part model.** Every S-flag ability decomposes the same way: a **trigger**
(when it fires), **filters** (who and where it reaches), **gates** (does it apply this
time), and an **effect** (what changes). An asset declares a list of these. The engine
keys a registry on the effect name, mirroring the A-side's
`map[domain.AbilityEffectType]EffectHandler` (`ability/dispatch.go:33`). One declared
effect may register more than one hook — transport is the proof, needing a grant at
decision time and a reactor at resolution time.

**Domain shape (contract; literal Go is Effort 1's first task).** `AssetDefinition`
grows `Effects []EffectDefinition`; `Transport *TransportProfile` stays alongside it
unchanged (D-3). An `EffectDefinition` carries `Trigger`, `Effect`, `Filters`, `Gates`,
and typed per-effect parameter pointers set iff their effect applies — mirroring
`AbilityStep`'s `Relocate *RelocateParams` / `Purchase *PurchaseParams` shape
(`asset.go:129`). `EffectFilters` carries `Owner`, `Scope`, `Attribute`, `Phases`,
`Action`, `AssetTypes`, `Categories`, `Stealthy *bool` — a pointer, because an absent
filter differs from `false`. Each phase is a string-backed `EffectRollPhase` naming a
`hooks.RollPhase`, since `domain` must not import `hooks` (Effort 1 plan, Decision 17).
`Phases` is a list since the Effort 3 plan, Decision 17 (Panopticon needs attack and
defense).
The effect name is a `SpecialEffectType` with `Special*` constants, beside the A-side's
`AbilityEffectType` (Effort 1 plan, Decision 16). `EffectGates` carries `OncePer`, `NoStack`, and
`Contest *EffectContest`, where `EffectContest` embeds the existing A-side
`domain.Contest` (`asset.go:113`) and adds `Roller` naming which side tests.
`RelocateParams` grows `Categories []FactionStat`, `ExcludeTypes []AssetType` and
`CoinCostMode` (Planning Decision 4).

**How each trigger reads `owner`.** `owner` names whose side of the event the filter
matches, and the reading is per-trigger because the triggers describe different events.
Each registrar documents its own reading; the loader does not need to know it.

| Trigger | `owner` selects |
|---|---|
| `turn_start`, `on_purchase` | whose turn or purchase — always `own` in the SWN set |
| `own_attack_hit`, `own_counter_hit`, `own_roll`, `own_defense_win` | the side that acted; `rival` names the party acted upon |
| `enemy_roll_at_me`, `stealth_arrival` | who the incoming actor is |
| `coin_loss` | who imposed the loss |
| `*_query` | whose query it is — `rival` means the effect reaches a non-owner's query |

**`owner` and `scope` are stated on every entry.** They have no defaults, so a reader
never has to know one (D-4). Omitting either fails the load.

**Loud failure.** The loader already rejects an unknown asset flag rather than skipping
it (`toAssetFlags`, `rulebook.go:557`). The effects list follows the same discipline:
an unknown `trigger`, `effect`, `filter` or `gate` name fails the load (any unknown key
in an asset file does); a parameter block whose name does not match the declared effect,
or a missing block for an effect that takes parameters, fails the load; a missing
`owner` or `scope` fails the load; and an entry whose effect has no registrar, or whose
trigger, owner, scope, gates or filters its registrar does not accept, fails at engine
construction (Planning Decision 7; Effort 1 plan, Decisions 2, 4, 11, 12). An asset with
**no** effects list stays silently inert exactly as today. Silence is for absent data,
not for wrong data.

**The eight dispatch items, by effort.** Each was verified against source in the
discovery session and re-verified this session. Item numbers are the discovery's.

| Item | What it is | Effort |
|---|---|---|
| 1 | Cat 3 reactors reach only the acting faction | 2 |
| 2 | Cat 3 reactors never reach asset scope | 2 |
| 3 | A roll hook sees only the roller's own hooks | 3 |
| 4 | Cat 4 asset-cost lookup is buyer-scoped | 4 |
| 5 | `GrantedMovementAbilities` has no caller | 5 |
| 6 | `on_purchase` needs a second dispatch route | 5 |
| 7 | Roll hooks are wired into the Attack action only | 3 |
| 8 | No dispatch point asks whether one asset may fight another | 4 |
| *new* | No dispatch point exists for a turn-start event | 1 |

**Item 2 is cheaper than the discovery implies.** `MutationReactorsFor` already checks
the asset bucket (`registry.go:148`). The `""` that defeats it is hardcoded inside
`reactDispatch` (`mutations.go:41`), below the three orchestrator call sites. It is a
signature change plus call-site edits, not a registry gap.

**Determinism.** Every fan-out over factions or assets iterates in sorted ID order —
`domain.SortedAssets` for assets, an explicit sort for factions. Map iteration order must
never reach a mutation slice, because the history file records it.

**Test patterns.** Follow `ability/dispatch_test.go` and `hooks/dispatch/roll_test.go`:
a `seqRoller{values}` for deterministic rolls, a gomock `MockCollector` for prompts.
Each registrar gets a test that its hook fires when the filters match and stays silent
when they do not. Rule 4 checks only what a registrar *declares*: nothing makes it
honour a gate or filter it accepts. So every gate and filter in a registrar's
`Accepts` gets a test that proves it suppresses — a spent `once_per` budget stops the
second firing, a `no_stack` mark stops the second copy, a lost contest stops the
effect, a non-matching filter stops the match. A registrar commit without those tests
is incomplete; the pre-merge review checks it. Loud-failure rules 1–4 each get a loader
or construction test asserting the error, not a skip.

**Model discipline.** Efforts 1, 3 and 5 plan on Opus. Effort execution is Sonnet by
default; Effort 1's first commit (the domain reshape) and Effort 3's item-7 commit carry
enough design weight to warrant Opus. The pre-merge code review is Opus.

## Out of Scope

- **A.001.6's fourteen assets and six mechanisms.** A change of asset owner, state held
  on an asset, a prompt to a faction that is not acting, a purchase that targets another
  asset, a permission model for movement, and damage redirected into combat. Bounded by
  this initiative's discovery, designed in A.001.6's own Discovery. Marketers,
  Seditionists and Monopoly pass through to A.001.6 with the restriction group.
- **Migrating `[transport]` onto the effects list.** Deferred by D-3. The eight transport
  profiles are untouched. Filed as a backlog refactor, triggered once the S-flag model is
  fully built.
- **The Deep Strike Landers and Blockade Runners bypass clause.** Both move cargo past an
  objecting planetary government. That needs A.001.6's permission model.
- **Retiring the `S` flag.** It stays a human-facing label on the asset card. Nothing in
  the engine reads it today and nothing will after this work; registration is driven by
  the presence of an effects list (ACD-1).
- **Capital Fleet and Mercenaries.** Their whole special is a maintenance cost, which the
  `maintenance` field already carries. No handler is owed.
- **Informers, Seductress and Covert Shipping.** All three carry both `A` and `S`, and
  A.001.2 shipped the active mechanic for each. Their `S` text restates the same ability.
- **`B.002` (semantic table keys).** Untriggered — effects key on their own names.
- **H.A.W.K assets.** A.001.5 authors them against the vocabularies settled here.

## Reconciliation Items

The arc-plan routed three here and the discovery found two more. Each is resolved by the
effort that makes its claim true, not in a separate documentation pass.

| Item | Resolved by |
|---|---|
| Catalog hook-category labels drifted (Cat 4 = RuleModifier family, Cat 5 = TieResolver) | Effort 1 — renumber the catalog to match `hooks/doc.go` |
| Catalog files free `relocate-other` as Cat 3 | Effort 5 — it is a Cat 4 `AssetMovementGranter`; correct when Transit Web lands |
| `hooks.md:89-92` says the movement phase calls `GrantedMovementAbilities` | Effort 5 — correct the page when the call is wired |
| `hooks.md:74-75` says the effects engine registers asset-scoped hooks; `effect-mutation.md` carries the same claim in its heading at line 63 while its body is correct at line 76 | Effort 2 — fixed by item 2 |
| The catalog routes two S-flag effects to `coin_steal` and `coin_drain`; A.001.2 removed both `AbilityEffectType` values | Effort 1 — the catalog's "latent vocabulary" table is stale on the same point |
| `hooks.md` lists the Cat 4 resolvers as cost, tech level and movement; `ResolveAttackEligibility` is a fourth, and it refuses where the others chain | **Resolved** at the Effort 4 close-out — Decision Record 21 graduated into `hooks.md` Key Decisions; resolver list and diagram fixed |

---

## Work Breakdown

Efforts 1, 3 and 5 carry their Phase/Commit/Task detail in their own per-effort plan
files. The outlines below for those three are **scoping guidance for those sessions**,
not executable tasks. Efforts 2 and 4 are specified at commit granularity to execute
straight from this file.

### Effort 1 — Foundation *(own Plan session)*

Planned in full in [`s-flag-effect-registry-effort-1-plan.md`](./s-flag-effect-registry-effort-1-plan.md)
— four commits. The outline below is the summary; the effort plan is authoritative.
Nothing here ships an effect that needs a dispatch extension; `coin_gain` is the proof
that the model works end to end.

- `refactor: add the special effect model to the domain` — `AssetDefinition.Effects`,
  `EffectDefinition`, `EffectFilters`, `EffectGates`, `EffectContest`, the per-effect
  parameter structs for all 13 effects, the vocabulary types, and the `RelocateParams`
  growth (Planning Decision 4). Defines the shared structures every later effort cites
  (KEY RULE 2).
- `feat: parse the asset effects list` — grow `assetRecord` with `Effects []effectRecord`;
  parse `filters`, `gates` and the named parameter block for all 13 effects; loud-failure
  rules 1–3; reject unknown keys; `parseAmount` for amounts (Planning Decision 11).
- `feat: effect registrar contract with construction-time validation` — the registrar
  spec struct; the registry keyed on effect name; the composed per-definition handler;
  `Register` rejects a duplicate definition (Planning Decision 6); loud-failure rule 4;
  `NewWithRulebook` grows an error return and the four callers change (Planning
  Decision 7).
- `feat: turn-start effect route and coin_gain` — the direct definition read at the start
  of the bookkeeping phase (Planning Decision 5); the `coin_gain` registrar; author Party
  Machine `SWN-C4-001`; the catalog reconciliation items.

Party Machine's authored entry:

```toml
[[assets.PartyMachine.effects]]
trigger = "turn_start"
effect  = "coin_gain"
[assets.PartyMachine.effects.filters]
owner = "own"
scope = "self"
[assets.PartyMachine.effects.coin_gain]
amount = "1"
```

---

### Effort 2 — Cat 3 reach

Dispatch items 1 and 2, the mutation vocabulary they need, and the four effects that ride
Cat 3. Commits 1 and 2 are independent of Effort 1's API. Commits 3–6 bind to the
registrar contract and re-ground against it first.

#### Commit 1 — `feat: cat 3 dispatch reaches every faction and asset scope`

Today `reactDispatch` looks up reactors for one faction with an empty asset ID
(`mutations.go:41`), so a defender's reactor never runs and an asset-scoped reactor is
never found. `TransportHandler` works around this by registering at faction scope
(`effects/transport.go:25`). Per-asset effects are the normal case for the nineteen, so
the dispatch grows rather than the effects contorting.

##### Task 1 — `internal/faction/engine/hooks/registry.go`

`MutationReactorsFor` bundles three scopes into one lookup, which cannot express "the
faction bucket once, then each asset bucket". Replace it with three narrow accessors.

**Find:**
```go
// MutationReactorsFor returns all Cat 3 hooks matching the given faction and asset.
func (registry *Registry) MutationReactorsFor(factionID, assetInstanceID string) []RegisteredMutationReactor {
	var result []RegisteredMutationReactor
	result = append(result, registry.mutationReactors[GlobalScope()]...)
	result = append(result, registry.mutationReactors[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.mutationReactors[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}
```
**Replace with:**
```go
// MutationReactorsGlobal returns the Cat 3 hooks registered at global scope.
func (registry *Registry) MutationReactorsGlobal() []RegisteredMutationReactor {
	return registry.mutationReactors[GlobalScope()]
}

// MutationReactorsForFaction returns the Cat 3 hooks registered against one faction.
func (registry *Registry) MutationReactorsForFaction(factionID string) []RegisteredMutationReactor {
	return registry.mutationReactors[FactionScope(factionID)]
}

// MutationReactorsForAsset returns the Cat 3 hooks registered against one asset instance.
func (registry *Registry) MutationReactorsForAsset(factionID, assetInstanceID string) []RegisteredMutationReactor {
	return registry.mutationReactors[AssetScope(factionID, assetInstanceID)]
}
```

##### Task 2 — `internal/faction/engine/hooks/dispatch/mutations.go`

Dispatch every faction's reactors, not just the acting faction's, and reach asset scope.
The acting faction is still passed, because the recursion-cap error names it.

**Find:**
```go
// MutationReactors runs all MutationReactors registered for faction
// (faction-scoped and global) in registration order against combined.
// Returned mutations are appended and the process recurses until no
// new mutations are produced. Returns an error if the recursion depth
// exceeds maxHookDepth to prevent infinite loops.
func MutationReactors(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	return reactDispatch(registry, faction, combined, factionState, rulebook, 0)
}
```
**Replace with:**
```go
// MutationReactors runs every registered MutationReactor against combined —
// global scope, then each faction's scope, then each asset's scope, in sorted
// ID order so the emitted slice is deterministic. Reactors belonging to a
// faction other than the acting one do run: a defender reacts to what was done
// to it. Returned mutations are appended and the process recurses until no new
// mutations are produced. Returns an error if the recursion depth exceeds
// maxHookDepth to prevent infinite loops.
func MutationReactors(
	registry *hooks.Registry,
	faction *domain.Faction,
	combined []domain.Mutation,
	factionState *state.FactionState,
	rulebook *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	return reactDispatch(registry, faction, combined, factionState, rulebook, 0)
}

// allReactors collects every registered Cat 3 hook in deterministic order:
// global, then per faction (sorted by ID), then per asset within that faction
// (sorted by ID).
func allReactors(registry *hooks.Registry, factionState *state.FactionState) []hooks.RegisteredMutationReactor {
	reactors := append([]hooks.RegisteredMutationReactor(nil), registry.MutationReactorsGlobal()...)

	factionIDs := make([]string, 0, len(factionState.Factions))
	for id := range factionState.Factions {
		factionIDs = append(factionIDs, id)
	}
	sort.Strings(factionIDs)

	for _, factionID := range factionIDs {
		reactors = append(reactors, registry.MutationReactorsForFaction(factionID)...)
		for _, asset := range domain.SortedAssets(factionState.Factions[factionID]) {
			reactors = append(reactors, registry.MutationReactorsForAsset(factionID, asset.ID)...)
		}
	}
	return reactors
}
```

##### Task 3 — `internal/faction/engine/hooks/dispatch/mutations.go`

Point the recursion at the full reactor set.

**Find:**
```go
	reactors := registry.MutationReactorsFor(faction.ID, "")
```
**Replace with:**
```go
	reactors := allReactors(registry, factionState)
```

##### Task 4 — `internal/faction/engine/hooks/dispatch/mutations.go`

Add the `sort` import alongside the existing ones.

**Find:**
```go
import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
```
**Replace with:**
```go
import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
```

##### Task 5 — `internal/faction/engine/effect/effects/transport.go`

Transport registered at faction scope only because asset scope was unreachable. Now that
it is reachable, bind the reactor to its own asset so it is not consulted for every
mutation in the faction.

**Find:**
```go
	hookRegistry.RegisterMutationReactor(
		hooks.FactionScope(faction.ID),
		"transport:"+handler.def.ID,
```
**Replace with:**
```go
	hookRegistry.RegisterMutationReactor(
		hooks.AssetScope(faction.ID, asset.ID),
		"transport:"+handler.def.ID,
```

##### Task 6 — `internal/faction/engine/hooks/dispatch/mutations_test.go`

Add two cases to the existing table: a reactor registered against a non-acting faction
fires on the acting faction's mutation batch, and a reactor registered at asset scope
fires. Assert reactor order is stable across runs by registering reactors against two
factions whose map insertion order differs from their sorted ID order. The existing
cases pass a `nil` state; give them one via a `stateOf` helper. Move the other
`MutationReactorsFor` callers — the Cat 3 tests in `hooks/registry_test.go` and the
transport test in `effect/effect_test.go` (now asserting asset scope) — to the new
accessors (Decision Record 1).

##### Task 7 — `docs/architecture/engine/hooks.md`, `effect-mutation.md`

Note in `hooks.md` that Cat 3 exposes its buckets separately and that
`MutationReactors` runs every faction's and asset's reactors in sorted order. Change
`effect-mutation.md`'s transport paragraph from faction-scoped to asset-scoped
(Decision Record 2).

##### Commit message

```
feat: cat 3 dispatch reaches every faction and asset scope

- replace MutationReactorsFor with global/faction/asset accessors
- dispatch every faction's reactors, sorted by faction then asset ID
- bind the transport reactor to asset scope now that it is reachable
- cover non-acting-faction and asset-scope dispatch
- correct the hooks and effect-mutation arch pages on reactor scope
```

#### Commit 2 — `feat: record the causing asset on damage mutations`

A reactor sees the mutation batch, not the action that produced it. `AssetHPDelta`
names the damaged asset and the causing *faction*, so a Franchise cannot tell whether it
landed the hit. `own_attack_hit` cannot be filtered to the declaring asset without this.

##### Task 1 — `internal/faction/domain/mutation.go`

**Find:**
```go
// AssetHPDelta adjusts an asset's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds the asset's definition max HP.
type AssetHPDelta struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
}
```
**Replace with:**
```go
// AssetHPDelta adjusts an asset's CurrentHP by Delta. Resolve pre-caps the
// delta so the applied value never exceeds the asset's definition max HP.
// CausedByAssetID names the asset that dealt the damage when one did; it is
// empty for damage with no asset source (bookkeeping, goal effects).
type AssetHPDelta struct {
	FactionID         string `json:"faction_id"`
	AssetID           string `json:"asset_id"`
	Delta             int    `json:"delta"`
	Cause             string `json:"cause"`
	CausedByFactionID string `json:"caused_by_faction_id"`
	CausedByAssetID   string `json:"caused_by_asset_id,omitempty"`
}
```

##### Task 2 — `internal/faction/engine/action/actions/attack.go`

`applyAssetDamage` is the single emitter for both the attack and the counterattack, so
the causing asset threads through one parameter.

**Find:**
```go
func (attack *AttackAction) applyAssetDamage(factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker map[string]int) {
	attack.mutations = append(attack.mutations, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             "attack",
		CausedByFactionID: causedByFactionID,
	})
```
**Replace with:**
```go
func (attack *AttackAction) applyAssetDamage(factionID, causedByFactionID string, asset *domain.Asset, damage int, tracker map[string]int, causedByAsset *domain.Asset) {
	causedByAssetID := ""
	if causedByAsset != nil {
		causedByAssetID = causedByAsset.ID
	}
	attack.mutations = append(attack.mutations, domain.AssetHPDelta{
		FactionID:         factionID,
		AssetID:           asset.ID,
		Delta:             -damage,
		Cause:             "attack",
		CausedByFactionID: causedByFactionID,
		CausedByAssetID:   causedByAssetID,
	})
```

##### Task 3 — `internal/faction/engine/action/actions/attack.go`

Three call sites pass the asset that dealt the damage. The two attack-damage sites pass
the attacker; the counterattack site passes the defender.

**Find:**
```go
					attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker)
				}
			} else {
				attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker)
			}
```
**Replace with:**
```go
					attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker, attacker)
				}
			} else {
				attack.applyAssetDamage(defenderFaction.ID, faction.ID, defender, damage, assetHPTracker, attacker)
			}
```

**Find:**
```go
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, counterDamage, assetHPTracker)
```
**Replace with:**
```go
				counterDamage := defenderDef.Counter.Roll(attack.roller)
				attack.applyAssetDamage(faction.ID, defenderFaction.ID, attacker, counterDamage, assetHPTracker, defender)
```

##### Task 4 — `internal/faction/engine/action/actions/attack_test.go`

Extend the existing attack and counterattack assertions to check `CausedByAssetID` names
the attacker on attack damage and the defender on counterattack damage.

##### Commit message

```
feat: record the causing asset on damage mutations

- add CausedByAssetID to AssetHPDelta
- thread the dealing asset through applyAssetDamage
- attack damage names the attacker, counterattack names the defender
```

#### Commit 3 — `feat: coin_steal effect with a per-target turn budget`

Blockade Fleet `SWN-F5-001` steals `1d4` Coin from the target faction when it lands a
hit; Franchise `SWN-W1-001` takes a flat `1`. Both cap at once per target faction per
turn, however many copies attack. Both are Cat 3 reactors on `own_attack_hit`, reading
`AssetHPDelta` (with `Counter == false`) or `BaseHPDelta`, with `Cause == "attack"` and
`CausedByAssetID` equal to the bound asset.

The commit opens with the plumbing the attack-hit match needs (Execution Decisions 3–5):
`ApplyAll` takes the engine roller and `RegisterContext` carries it; `AssetHPDelta`
grows `Counter bool`, and `BaseHPDelta` grows `CausedByAssetID`;
`applyAssetDamage` takes a `counter` argument, and the Base-redirect branch fills
`CausedByAssetID`. A shared `attackHits` helper in `effects` yields the target faction
of each hit the bound asset dealt, for this commit and Commit 4.

This commit also builds the **Cat 3 budget gate**, the first consumer of Planning
Decision 3. The gate helper takes the owning faction, the effect name, the def ID and
the target faction ID, builds the key `effect:<name>:<defID>:<targetFactionID>`, and
reports whether the budget is spent — writing it when it is not. `once_per =
"owner_faction"` omits the target segment; `once_per = "asset"` puts the asset instance
ID in the target's place (`effect:<name>:<defID>:<assetID>`), with no target segment, so
one Bank absorbs one loss whichever rival caused it. The helper lives beside the registrars, not in `hooks`,
because `RollWithHooks` owns the roll-path budget and this is the non-roll path.

Franchise's amount is capped at the target's balance ("the enemy faction loses 1 Coin
*if available*"); Blockade Fleet's roll is capped the same way. The effect emits a paired
`CoinDelta` — negative against the target, positive to the owner — so the transfer is
visible on both sides of the history.

Fold in the A.001.2 D5 prose cleanup for Blockade Fleet while `force_assets.toml` is open
(Planning Decision 15): delete "As an action, this asset may also move itself to a world
within one hex." from its description.

Authored entries:

```toml
[[assets.BlockadeFleet.effects]]
trigger = "own_attack_hit"
effect  = "coin_steal"
[assets.BlockadeFleet.effects.filters]
owner = "rival"
scope = "world"
[assets.BlockadeFleet.effects.gates]
once_per = "target_faction"
[assets.BlockadeFleet.effects.coin_steal]
amount = "1d4"

[[assets.Franchise.effects]]
trigger = "own_attack_hit"
effect  = "coin_steal"
[assets.Franchise.effects.filters]
owner = "rival"
scope = "world"
[assets.Franchise.effects.gates]
once_per = "target_faction"
[assets.Franchise.effects.coin_steal]
amount = "1"
```

##### Commit message

```
feat: coin_steal effect with a per-target turn budget

- cat 3 budget gate keyed on effect, definition and target faction
- coin_steal registrar reacting to own_attack_hit
- amount capped at the target's balance
- author Blockade Fleet and Franchise
- drop the misleading move clause from Blockade Fleet's description
```

#### Commit 4 — `feat: self_damage effect`

Zealots `SWN-F3-001` take `1d4` damage every time they land a successful attack or a
counterattack. Two entries, one per trigger, because the two are separate events and the
engine's counterattack only fires when the defender wins (`attack.go:192-197`) — so
"perform a counterattack" already implies a hit. The registrar emits `AssetHPDelta`
against the Zealots themselves, plus `AssetRemoved` when the damage is lethal, matching
`applyAssetDamage`'s inline pattern. The two entries split on the `Counter` mark
(Execution Decision 4): `own_attack_hit` reads `Counter == false`, `own_counter_hit`
reads `Counter == true`, so one hit fires exactly one entry.

The counterattack half is what needs dispatch item 1: when Zealots counterattack, they
belong to the *defending* faction, whose reactors did not run before Commit 1.

Authored entries:

```toml
[[assets.Zealots.effects]]
trigger = "own_attack_hit"
effect  = "self_damage"
[assets.Zealots.effects.filters]
owner = "own"
scope = "self"
[assets.Zealots.effects.self_damage]
amount = "1d4"

[[assets.Zealots.effects]]
trigger = "own_counter_hit"
effect  = "self_damage"
[assets.Zealots.effects.filters]
owner = "own"
scope = "self"
[assets.Zealots.effects.self_damage]
amount = "1d4"
```

##### Commit message

```
feat: self_damage effect

- self_damage registrar on own_attack_hit and own_counter_hit
- emit AssetRemoved inline when the damage is lethal
- author Zealots
```

#### Commit 5 — `feat: ignore_coin_loss effect`

Bank `SWN-W4-004` lets its faction ignore one Coin cost or loss imposed by another
faction, once per turn per Bank. A Cat 3 reactor cannot cancel a mutation
(Planning Decision 10), so it emits a compensating `+CoinDelta` with cause
`bank_absorbed` matching the loss it found.

The reactor matches a `CoinDelta` where `FactionID` is the owner, `Delta` is negative,
and `CausedByFactionID` is neither empty nor the owner — that last clause is the "imposed
by another faction" filter, and it is why `owner = "rival"` reads as "who imposed the
loss" here (see Shared Context). `once_per = "asset"` gives each Bank its own budget, so
two Banks absorb two losses.

This needs dispatch item 1: the loss is usually imposed during a rival's turn, when the
Bank's owner is not the acting faction.

Authored entry:

```toml
[[assets.Bank.effects]]
trigger = "coin_loss"
effect  = "ignore_coin_loss"
[assets.Bank.effects.filters]
owner = "rival"
scope = "faction"
[assets.Bank.effects.gates]
once_per = "asset"
```

##### Commit message

```
feat: ignore_coin_loss effect

- ignore_coin_loss registrar emitting a compensating CoinDelta
- match losses imposed by another faction only
- per-asset budget so multiple Banks absorb multiple losses
- author Bank
```

#### Commit 6 — `feat: clear_stealth effect on stealth arrival`

Tripwire Cells `SWN-C4-003` make an immediate Cunning vs. Cunning attack against the
owning faction whenever a stealthed asset lands or is purchased on their world. On
success the asset loses its stealth.

The reactor matches `AssetMoved` into the Cells' world (an in-flight move, with no
`WorldID`, is not a landing) where the moved asset is stealthy and rival-owned. The
contest gate rolls with `roller = "self"` — the Cells attack — and the effect emits
`AssetStealthCleared` against the target.

This needs dispatch item 1: a rival's stealthed asset arrives during the *rival's* turn,
so the Cells' owner is not the acting faction. The purchase case is deferred to Effort 5
(Execution Decision 7): nothing purchases a stealthed asset until `apply_stealth` lands.

Authored entry:

```toml
[[assets.TripwireCells.effects]]
trigger = "stealth_arrival"
effect  = "clear_stealth"
[assets.TripwireCells.effects.filters]
owner    = "rival"
scope    = "world"
stealthy = true
[assets.TripwireCells.effects.gates.contest]
roller        = "self"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```

##### Commit message

```
feat: clear_stealth effect on stealth arrival

- clear_stealth registrar reacting to a stealthed asset landing
- contest gate rolled by the declaring asset
- author Tripwire Cells
```

---

### Effort 3 — Roll hooks *(own Plan session)*

Planned in full in [`s-flag-effect-registry-effort-3-plan.md`](./s-flag-effect-registry-effort-3-plan.md)
— six commits. The outline below is the summary; the effort plan is authoritative.

- `feat: roll hooks reach the opponent and the world's assets` — item 3. The dispatcher
  consults the asset buckets of both sides' assets on the roll's world, never the
  opponent's faction bucket, and spends each budget on the hook's owning faction.
- `feat: route every roll through RollWithHooks` — item 7. Expand Influence's faction
  test, rival rolls and free attack, and the A-side ability contest. `RollContext`
  gains `Action` and `OpponentAsset`.
- `feat: bonus_die effect` — Integral Protocols (die only), Surveyors (owner's
  expansion test; D5 prose cleanup), Panopticon Matrix's die (attacks and defenses).
  The `phase` filter becomes a list.
- `feat: remove_tag_dice effect` — Blackmail, through a new Cat 1 `RollOfferFilter` and
  a typed `ModifierOffer.FromTag`.
- `feat: reroll effect` — Book of Secrets, with separate own and forced budgets; the
  budget key gains the trigger.
- `feat: clear_stealth on turn start` — Panopticon Matrix's stealth test at its owner's
  turn start. `FireContext` and `FireTurnStart` gain `FactionState`.

Authored entries for this effort (literal blocks live in the effort plan):

```toml
[[assets.IntegralProtocols.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.IntegralProtocols.effects.filters]
owner     = "own"
scope     = "self"
phase     = ["defense"]
attribute = "Cunning"
[assets.IntegralProtocols.effects.bonus_die]
sides = 10
count = 1

[[assets.Surveyors.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.Surveyors.effects.filters]
owner  = "own"
scope  = "world"
phase  = ["faction_test"]
action = "Expand Influence"
[assets.Surveyors.effects.bonus_die]
sides = 10
count = 1

[[assets.PanopticonMatrix.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[assets.PanopticonMatrix.effects.filters]
owner     = "own"
scope     = "world"
phase     = ["attack", "defense"]
attribute = "Cunning"
[assets.PanopticonMatrix.effects.bonus_die]
sides = 10
count = 1

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

[[assets.Blackmail.effects]]
trigger = "enemy_roll_at_me"
effect  = "remove_tag_dice"
[assets.Blackmail.effects.filters]
owner = "rival"
scope = "self"

[[assets.BookOfSecrets.effects]]
trigger = "own_roll"
effect  = "reroll"
[assets.BookOfSecrets.effects.filters]
owner = "own"
scope = "world"
[assets.BookOfSecrets.effects.gates]
once_per = "owner_faction"
[assets.BookOfSecrets.effects.reroll]
dice     = 1
elective = true

[[assets.BookOfSecrets.effects]]
trigger = "enemy_roll_at_me"
effect  = "reroll"
[assets.BookOfSecrets.effects.filters]
owner = "rival"
scope = "world"
[assets.BookOfSecrets.effects.gates]
once_per = "owner_faction"
[assets.BookOfSecrets.effects.reroll]
dice     = 1
elective = true
```

Panopticon Matrix's `bonus_die` names both `attack` and `defense`, so Expand Influence
tests and ability contests do not get its die (Effort 3 plan, Decision 5).

---

### Effort 4 — Rule queries

Dispatch items 4 and 8, and the three effects that are rule queries with a filter. They
store nothing; each answers a question the engine already asks, or is about to ask.
Commits 1 and 4 are independent of Effort 1's API. Commits 2, 3 and 5 bind to the
registrar contract and re-ground against it first.

#### Commit 1 — `feat: asset-scoped cat 4 cost and tech-level lookup`

`ResolveAssetCost` reads modifiers for the buyer (`rules.go:17`), and
`ResolveWorldTechLevel` does the same (`rules.go:45`). Both pass an empty asset ID, so
neither finds a hook registered at `AssetScope` — the scope every effect registrar uses.
The two resolvers need different reach:

- **Cost** — Local Investments belongs to a rival and charges the buyer. The question
  is "who has an asset on *this world* with something to say about this purchase",
  which neither the buyer's faction scope nor global scope expresses. The chain walks
  the world's occupants, skipping the buyer's own.
- **Tech level** — every floor serves its owner only (Decision Record 13), and R&D
  Department reaches every world. The chain walks the buyer's own assets wherever they
  stand; each hook checks its own reach (Decision Record 14). The hook also needs the
  definition for its `category` filter, so the cap resolves per definition (Decision
  Record 17).

##### Task 1 — `internal/faction/engine/hooks/dispatch/rules.go`

`ResolveAssetCost` gains the world's occupants. The world index supplies them, so the
signature grows one parameter rather than the caller pre-filtering.

**Find:**
```go
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int) int {
	if registry == nil {
		return baseCost
	}
	cost := baseCost
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		cost = registered.Hook.ModifyAssetCost(buyer, def, world, cost)
	}
	return cost
}
```
**Replace with:**
```go
// ResolveAssetCost runs the buyer's own AssetCostModifiers, then those registered
// against every asset standing on world — including a rival's, because a cost
// hook may belong to a faction other than the buyer (Local Investments). Hooks
// run in registration order, chaining outputs so each sees the previous result.
// onWorld is the world index's asset slice for world; a nil slice adds nothing.
func ResolveAssetCost(registry *hooks.Registry, buyer *domain.Faction, def *domain.AssetDefinition, world string, baseCost int, onWorld []*domain.Asset) int {
	if registry == nil {
		return baseCost
	}
	cost := baseCost
	for _, registered := range registry.AssetCostModifiersFor(buyer.ID, "") {
		cost = registered.Hook.ModifyAssetCost(buyer, def, world, cost)
	}
	for _, asset := range sortedByID(onWorld) {
		if asset.OwnerID == buyer.ID {
			continue
		}
		for _, registered := range registry.AssetCostModifiersForAsset(asset.OwnerID, asset.ID) {
			cost = registered.Hook.ModifyAssetCost(buyer, def, world, cost)
		}
	}
	return cost
}

// sortedByID returns assets in ID order so a world-scoped hook chain is
// deterministic regardless of index insertion order.
func sortedByID(assets []*domain.Asset) []*domain.Asset {
	sorted := slices.Clone(assets)
	slices.SortFunc(sorted, func(a, b *domain.Asset) int {
		return strings.Compare(a.ID, b.ID)
	})
	return sorted
}
```

##### Task 2 — `internal/faction/engine/hooks/registry.go`

Add the asset-only accessor the world-scoped chain needs, mirroring Effort 2 Commit 1's
`MutationReactorsForAsset`.

Append after `AssetCostModifiersFor`:
```go
// AssetCostModifiersForAsset returns the Cat 4 asset-cost hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) AssetCostModifiersForAsset(factionID, assetInstanceID string) []RegisteredAssetCostModifier {
	return registry.assetCostModifiers[AssetScope(factionID, assetInstanceID)]
}

// WorldTechLevelModifiersForAsset returns the Cat 4 tech-level hooks registered
// against one asset instance, without the global or faction buckets.
func (registry *Registry) WorldTechLevelModifiersForAsset(factionID, assetInstanceID string) []RegisteredWorldTechLevelModifier {
	return registry.worldTechLevelModifiers[AssetScope(factionID, assetInstanceID)]
}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`, `dispatch/rules.go`

The tech-level hook gains the definition (Decision Record 17). No implementer exists, so
only the interface and the resolver change.

In `rule_modifier.go`, **Find:**
```go
	ModifyWorldTechLevel(faction *domain.Faction, world string, baseTL int) int
```
**Replace with:**
```go
	ModifyWorldTechLevel(faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int
```

In `dispatch/rules.go`, replace `ResolveWorldTechLevel` whole:
```go
// ResolveWorldTechLevel runs faction's own WorldTechLevelModifiers, then those
// registered against each asset faction owns, in asset ID order. A tech-level
// floor serves its owner only, so a rival's assets are never asked; each
// asset-scoped hook decides whether its reach covers world and def.
func ResolveWorldTechLevel(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int {
	if registry == nil {
		return baseTL
	}
	tl := baseTL
	for _, registered := range registry.WorldTechLevelModifiersFor(faction.ID, "") {
		tl = registered.Hook.ModifyWorldTechLevel(faction, def, world, tl)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.WorldTechLevelModifiersForAsset(faction.ID, asset.ID) {
			tl = registered.Hook.ModifyWorldTechLevel(faction, def, world, tl)
		}
	}
	return tl
}
```
Add `slices` and `strings` to the imports (Decision Record 18).

##### Task 4 — `internal/faction/engine/action/actions/buy_asset.go`

The cost call site gains the world's assets from the index.

**Find:**
```go
	ba.cost = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost)
```
**Replace with:**
```go
	var onWorld []*domain.Asset
	if ba.worldEngine != nil && ba.worldEngine.Index != nil {
		onWorld = ba.worldEngine.Index.AssetsByLocation[ba.buyOrder.World]
	}
	ba.cost = dispatch.ResolveAssetCost(ba.registry, faction, def, ba.buyOrder.World, def.Cost, onWorld)
```

The tech-level call in `purchasableDefinitions` moves into the def loop, so each
candidate gets its own cap. An unknown world keeps no ceiling and asks no hook, as now.

**Find:**
```go
	techCap := math.MaxInt
	holdsGovernment := false
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		techCap = dispatch.ResolveWorldTechLevel(registry, faction, fragmentID, loc.TechLevel())
		if worldEngine.Index != nil {
			holdsGovernment = domain.HasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID])
		}
	}

	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: def.Cost, TechCap: techCap, HoldsGovernment: holdsGovernment}) {
```
**Replace with:**
```go
	worldTechLevel := math.MaxInt
	holdsGovernment := false
	if worldEngine != nil {
		loc, ok := worldEngine.Location(fragmentID)
		if !ok {
			return nil
		}
		worldTechLevel = loc.TechLevel()
		if worldEngine.Index != nil {
			holdsGovernment = domain.HasPlanetaryGovernment(faction.ID, worldEngine.Index.BasesByLocation[fragmentID])
		}
	}

	var result []*domain.AssetDefinition
	for _, def := range rulebook.Assets {
		techCap := worldTechLevel
		if worldEngine != nil {
			techCap = dispatch.ResolveWorldTechLevel(registry, faction, def, fragmentID, worldTechLevel)
		}
		if domain.CanPurchase(faction, def, domain.PurchaseTerms{Cost: def.Cost, TechCap: techCap, HoldsGovernment: holdsGovernment}) {
```

Re-ground at execution: confirm `purchasableDefinitions` is still the only reader of the
tech cap on the purchase path. If a validate step checks tech level on its own, it takes
the same per-def call.

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules_test.go`

- **Cost, rival on world:** the cost hook is registered against a rival's asset on the
  world and the buyer has none. The surcharge applies.
- **Cost, own asset skipped:** a cost hook on the buyer's own asset on the world does
  not run.
- **Tech level, own asset off world:** a tech-level hook on one of the buyer's assets on
  another world runs. The hook, not the resolver, decides reach.
- **Tech level, rival asset not asked:** a tech-level hook on a rival's asset does not
  run.
- **Tech level, def reaches the hook:** a stub hook that records its `def` sees the
  definition passed to the resolver.

##### Commit message

```
feat: asset-scoped cat 4 cost and tech-level lookup

- chain cost hooks registered against rival assets on the world
- chain tech-level hooks registered against the buyer's own assets
- pass the asset definition to tech-level hooks; resolve the cap per definition
- add asset-only registry accessors for both Cat 4 families
- sort each asset-scoped chain by asset ID for determinism
```

#### Commit 2 — `feat: cost_surcharge effect`

Local Investments `SWN-W1-003`: any *other* faction buying an asset on that world pays 1
extra Coin. The money is lost, not gained by the owner, so the effect returns a raised
cost and emits nothing. `no_stack` caps the surcharge at one application however many
Local Investments stand on the world — the registrar returns the base cost unchanged if a
surcharge from the same effect name has already been applied in this chain.

`no_stack` is the first gate that is a property of the *query*, not of the turn, so it
cannot use `HookBudgets`. It is carried on the chain itself: the registrar marks its
contribution and later copies see the mark. The mechanism lands here as `no_stack`'s only
consumer (Decision Record 15).

In `hooks/rule_modifier.go`, the chain state becomes a struct the hooks share:
```go
// AssetCostQuery is the state an asset-cost chain carries from hook to hook.
// Surcharge is the part of Cost that surcharge hooks added, so a purchase
// discount can floor below it without removing it. Applied holds the effect
// names whose no_stack gate has already fired.
type AssetCostQuery struct {
	Cost      int
	Surcharge int
	Applied   map[string]bool
}

// AssetCostModifier adjusts the Coin cost of an asset purchase (Cat 4).
type AssetCostModifier interface {
	ModifyAssetCost(buyer *domain.Faction, def *domain.AssetDefinition, world string, query *AssetCostQuery)
}
```

`ResolveAssetCost` builds
`&hooks.AssetCostQuery{Cost: baseCost, Applied: map[string]bool{}}`, passes it to every
hook in both loops, and returns the query by value. The surcharge adds to `Cost` and
`Surcharge`. `dispatch.DiscountedCost` applies a pending discount on top. Buy Asset's
menu, Buy Asset's `Resolve` and `purchase_enabler` all price through this chain
(Decision Record 23). Preceptor Archive
(`tags/preceptor_archive.go:14`) is the only implementer: it sets `query.Cost` in place
of its return value.

The registrar honours every filter and gate it accepts: `owner = "rival"` (the buyer is
not the owner), `scope = "world"` (its asset stands on the purchase world), and
`active()` (Decision Record 11). The resolver walks only the world's rival occupants, but
the registrar does not rely on that. The hook reads its asset as
`owner.Assets[assetID]` at query time, because a Cat 4 hook gets no state (Decision
Record 19). Three test stubs implement `ModifyAssetCost` and move with Preceptor
Archive.

Authored entry:

```toml
[[assets.LocalInvestments.effects]]
trigger = "asset_cost_query"
effect  = "cost_surcharge"
[assets.LocalInvestments.effects.filters]
owner = "rival"
scope = "world"
[assets.LocalInvestments.effects.gates]
no_stack = true
[assets.LocalInvestments.effects.cost_surcharge]
amount = 1
```

##### Commit message

```
feat: cost_surcharge effect

- cost_surcharge registrar as a world-scoped AssetCostModifier
- AssetCostQuery carries the cost chain; Preceptor Archive moves to it
- no_stack gate carried on the resolution chain, not on HookBudgets
- author Local Investments
```

#### Commit 3 — `feat: tech_level_floor effect`

Three assets raise the effective tech level for a purchase, each with a different reach.
Laboratory `SWN-W3-002` treats its world as Tech Level 4 for any asset. Pretech
Researchers `SWN-W5-002` treats its world as Tech Level 5 for Cunning and Wealth assets
only. R&D Department `SWN-W6-002` treats *all* worlds as Tech Level 4 for Wealth assets —
`scope = "faction"`, not `"world"`, and the only entry in the initiative that uses it.

The registrar is a `WorldTechLevelModifier` returning `max(current, level)` when every
filter passes. Each floor serves its owner only (Decision Record 13). Commit 1's resolver
asks only the buyer's own assets, wherever they stand, so the hook decides its own reach
(Decision Record 14):

- `owner = "own"` — the querying faction owns the asset.
- `scope = "world"` — the asset stands on the purchase world. `scope = "faction"` — any
  world.
- `category` — the `def` passed to the hook is in the listed categories (Decision
  Record 17).
- `active()` — the asset is active (Decision Record 11).

The hook reads its asset through its owner, like `cost_surcharge` (Decision Record 19).
The commit also adds `TestNewWithRulebook_ShippedRulebook`, which builds the engine
from `rulebooks/swn` so that every authored entry passes registrar validation (Decision
Record 20).

Authored entries:

```toml
[[assets.Laboratory.effects]]
trigger = "tech_level_query"
effect  = "tech_level_floor"
[assets.Laboratory.effects.filters]
owner = "own"
scope = "world"
[assets.Laboratory.effects.tech_level_floor]
level = 4

[[assets.PretechResearchers.effects]]
trigger  = "tech_level_query"
effect   = "tech_level_floor"
[assets.PretechResearchers.effects.filters]
owner    = "own"
scope    = "world"
category = ["Cunning", "Wealth"]
[assets.PretechResearchers.effects.tech_level_floor]
level = 5

[[assets.RandDDepartment.effects]]
trigger  = "tech_level_query"
effect   = "tech_level_floor"
[assets.RandDDepartment.effects.filters]
owner    = "own"
scope    = "faction"
category = ["Wealth"]
[assets.RandDDepartment.effects.tech_level_floor]
level = 4
```

##### Commit message

```
feat: tech_level_floor effect

- tech_level_floor registrar raising the effective tech level to a floor
- category filter narrows which asset categories the floor serves
- author Laboratory, Pretech Researchers and R&D Department
- build the engine from the shipped rulebook in a test
- test that the purchase tech cap resolves per definition
```

#### Commit 4 — `feat: attack eligibility cat 4 query`

No dispatch point asks whether one asset may fight another. Eligibility is decided by
three plain functions with no hook consulted: `attackerHasTarget` (`attack.go:240`) and
`eligibleDefendersAtHex` (`eligibility.go:30`) choose who may fight, and the counterattack
branch reads `defenderDef.Counter` directly (`attack.go:196-201`).

Three assets restrict it, in three different directions. Planetary Defenses restricts who
may attack *it*. Lawyers restricts what it may attack **and** counterattack. Integral
Protocols restricts what it may defend against.

One interface serves all three (Decision Record 16). The query carries the striking
asset, the struck asset, both definitions and a `side`: `attack` from
`eligibleDefendersAtHex` (which `attackerHasTarget` and `Resolve` both call), `counter`
from the counterattack branch, where the original defender is the striker. The striker's
hooks answer for `attack` and `counter`. On an attack only, the target's hooks then
answer as `defense` (Decision Record 21). The authored `side` in Commit 5 (`defense`,
`attack_and_counter`) names which of these a restriction answers.

The new Cat 4 family member follows the pattern in `rule_modifier.go`: a named interface,
a `Registered*` struct, a registry map, a `Register*` method, a `*For` accessor, an
asset-only accessor, and a `Resolve*` dispatcher in `dispatch/rules.go`. The dispatcher is
consulted at all three sites above, and the counterattack site is the one most easily
missed.

##### Commit message

```
feat: attack eligibility cat 4 query

- new Cat 4 interface reporting whether one asset may fight another
- registry map, registration and asset-scoped accessor
- consult it in attackerHasTarget, eligibleDefendersAtHex and the counterattack branch
```

#### Commit 5 — `feat: restrict_attack effect`

The three restrictions authored against Commit 4's query. Each is a filter plus a
direction, and none stores anything.

Planetary Defenses `SWN-F6-002` may only defend against Starship-type attackers —
`asset_type`, allow-only. Lawyers `SWN-W2-002` may not attack or counterattack Force
assets — `category`, deny-listed, and note this corrects the discovery's filter table
(Planning Decision 12). Integral Protocols `SWN-F7-002` may only defend against attacks
versus Cunning — `attribute`, allow-only, matching the attack profile's `DefenderStat`.

A restriction applies only while its asset is active (Decision Record 11). An inactive
Lawyers or Planetary Defenses restricts nothing.

Authored entries:

```toml
[[assets.PlanetaryDefenses.effects]]
trigger    = "attack_eligibility_query"
effect     = "restrict_attack"
[assets.PlanetaryDefenses.effects.filters]
owner      = "rival"
scope      = "self"
asset_type = ["Starship"]
[assets.PlanetaryDefenses.effects.restrict_attack]
side  = "defense"
allow = "only_listed"

[[assets.Lawyers.effects]]
trigger  = "attack_eligibility_query"
effect   = "restrict_attack"
[assets.Lawyers.effects.filters]
owner    = "rival"
scope    = "self"
category = ["Force"]
[assets.Lawyers.effects.restrict_attack]
side  = "attack_and_counter"
allow = "except_listed"

[[assets.IntegralProtocols.effects]]
trigger   = "attack_eligibility_query"
effect    = "restrict_attack"
[assets.IntegralProtocols.effects.filters]
owner     = "rival"
scope     = "self"
attribute = "Cunning"
[assets.IntegralProtocols.effects.restrict_attack]
side  = "defense"
allow = "only_listed"
```

Integral Protocols is now complete: its `bonus_die` shipped in Effort 3 and its
`restrict_attack` ships here (Planning Decision 14).

##### Commit message

```
feat: restrict_attack effect

- restrict_attack registrar with allow-only and deny-listed directions
- filter on asset type, category or attack attribute per asset
- author Planetary Defenses, Lawyers and Integral Protocols
```

---

### Effort 5 — Movement and purchase *(own Plan session)*

Planned in full in [`s-flag-effect-registry-effort-5-plan.md`](./s-flag-effect-registry-effort-5-plan.md)
— five commits. The outline below was the scoping input; the effort plan is
authoritative. Where they differ, the effort plan wins: the movement grant is
faction-level (Decision 1), granted moves run after the decision batch applies
(Decision 6), and the purchase route fires from an orchestrator scan, not from each
purchase path (Decision 10).

Two dispatch items, one new package, one hook type reshaped, and the two remaining
assets.

- `refactor: extract the shared relocate orchestration` — Planning Decision 2. A new
  `relocate` package holding enumerate → filter → prompt → build, taking a one-method
  `OrderCollector` interface so both the action collector and the phase collector satisfy
  it. `ability/relocate.go` shrinks to a call. The world-engine primitive
  (`BuildRelocateMutations`, `WorldsWithin`, `WorldsInRegion`) is untouched — this
  extracts the orchestration above it, not the geometry. Resolves Open Question 1.
- `feat: relocate honours category, excluded type and flat cost` — the fields
  `Categories []FactionStat`, `ExcludeTypes []AssetType` and `CoinCostMode` already exist
  and parse (Effort 1 added them, Planning Decision 4); this commit makes the shared
  relocate code read them, and lifts the loader's rejection of them on an `[ability]`
  relocate (`convertAbilityStep`, Effort 1 plan Decision 14). Transit Web reaches
  "non-Starship Cunning or Wealth assets", which the single `AssetType` field cannot
  express. `TransportProfile` already carries the equivalent pair as `CargoTypes` and
  `ExcludeCategories`, so the shape has precedent.
- `feat: wire granted movement abilities into the movement phase` — item 5.
  `dispatch.GrantedMovementAbilities` has no caller; the only reference is its own
  definition (`rules.go:54`). `prepareMovementDecisions` (`orchestrator.go:637`) takes no
  registry, so it must be threaded in. That function already hardcodes a
  `def.Transport != nil` branch at `:664-668`; the granted-ability path must **not** add a
  second hardcoded branch beside it. Grow `hooks.MovementAbility` from its
  `{Name, Description}` placeholder (`rule_modifier.go:7`) to carry an effect
  discriminator and typed parameters, so the phase can execute what a granter offers.
- `feat: grant_movement effect` — Transit Web `SWN-W7-003`. Registers an
  `AssetMovementGranter`; the phase executes it through the shared `relocate` package
  without spending the action.
- `feat: on_purchase effect route and apply_stealth` — item 6.
  `EffectsEngine.ApplyAll` runs once per cycle (`effect.go:34`), so an asset bought
  mid-turn has no registered hook until the next cycle, but Psychic Assassins
  `SWN-F5-003` must be stealthy the moment it is bought. The purchase path reads the
  definition's effects list directly (D-5). **Both** purchase paths need the route:
  `buy_asset.go` and `purchase_enabler.go` both build through
  `domain.NewPurchasedAsset` (`asset.go:237`), and Psychic Assassins is a Force asset an
  enabler can buy. Note that `buy_asset.go:62` already hardcodes `SWN-C3-002` for a
  different stealth-on-purchase case; that asset belongs to A.001.6 and is not migrated
  here, but the new route should sit cleanly beside it rather than duplicating it.
  **This commit also wires Tripwire Cells' purchase case** (Execution Decision 7): the
  `clear_stealth` reactor from Effort 2 matches movement only, and must learn to see an
  asset purchased stealthed on its world, in whatever mutation shape this route emits.

Authored entries:

```toml
[[assets.TransitWeb.effects]]
trigger = "movement_options_query"
effect  = "grant_movement"
[assets.TransitWeb.effects.filters]
owner = "own"
scope = "faction"
[assets.TransitWeb.effects.grant_movement.relocate]
scope          = "other"
range          = "hexes"
max_hex        = 3
coin_cost      = 1
coin_cost_mode = "flat"
categories     = ["Cunning", "Wealth"]
exclude_types  = ["Starship"]

[[assets.PsychicAssassins.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"
[assets.PsychicAssassins.effects.filters]
owner = "own"
scope = "self"
```
