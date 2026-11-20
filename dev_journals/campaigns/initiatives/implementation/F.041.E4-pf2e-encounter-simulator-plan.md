# PF2e Encounter Simulator (F.041) — Effort 4 Plan: Triggers and Limits

## In Plain Words

- **What:** the second of three rules efforts. Effort 3 taught the sim what an ability does to a creature. This effort teaches it when things happen and how often. A turn no longer has a fixed three actions: slowed, stunned and quickened change the count, and a knocked-out PC wakes prone and must Stand. An ability can be free, cost one to three actions, be used once per fight, draw on a shared count (potions, quick bombs) or recharge on a 1d4. An ability can fire on a trigger: a reaction, a free action or a passive effect. Rolls can be made twice (fortune and misfortune). A PC can heal an ally, and an act of loyalty hurts the Avatar. Skill checks arrive, computed from each PC's build.
- **Why:** the Fall of Luna fight turns on these rules: Luna's Challenge Fate and Destiny Spores' recharge, the Avatar's Vengeful Spite, Wail, Last Stand and Sin Vulnerability, Kas's Attack of Opportunity, the party's potions and Haste. Without them the sim trades Strikes and saves on a fixed clock.
- **How it fits:** Effort 4 of 8 in `F.041` (sim core), on branch `feature/pf2e-encounter-simulator`. The [rules scope](./F.041-pf2e-encounter-simulator-rules-scope.md) rows whose last rule is E4 are this effort's checklist. Effort 5 adds spells and tables. Effort 6 writes every kept ability by hand, starting from this effort's test files. Effort 7's rule lists answer the reaction question this effort adds to the decider. Position is another initiative (`A.004.2`), so every reach or movement rule here has a stated behaviour without it.

## Context / Goal

- Discovery: [`F.041-pf2e-encounter-simulator-discovery.md`](../discovery/F.041-pf2e-encounter-simulator-discovery.md). Read D-1 to D-26 as the [arc-discovery](../arcs/A.004-encounter-lab/A.004-encounter-lab-arc-discovery.md) sorts them.
- Arc plan: [`A.004-encounter-lab-arc-plan.md`](../arcs/A.004-encounter-lab/A.004-encounter-lab-arc-plan.md). AP-1 to AP-10 and ACD-1 to ACD-10 bind this plan.
- Overview plan: [`F.041-pf2e-encounter-simulator-overview-plan.md`](./F.041-pf2e-encounter-simulator-overview-plan.md). Decisions 1 to 13 bind this plan; its Work Breakdown, Effort 4, sets the scope.
- Rules scope: [`F.041-pf2e-encounter-simulator-rules-scope.md`](./F.041-pf2e-encounter-simulator-rules-scope.md). Commit 13 writes this session's rules calls into its sections 2, 3 and 6.
- Effort 3 plan: [`F.041.E3-pf2e-encounter-simulator-plan.md`](./F.041.E3-pf2e-encounter-simulator-plan.md). Its Decision Record 9 (the shared critical specialization row) is fixed in Commit 3.
- Ends with: every rules-scope row whose last rule is E4 has a test, each kept E4 item exists as an ability in the real format, and the wiki documents triggers and limits.

## Decisions Ratified in Planning

Robert's calls are marked. The rest are plan decisions.

**Rules calls (Robert).** Each goes into the rules scope, section 6, in Commit 13.

1. **Wail follows Wails of the Damned as written, with the Avatar's own numbers.** Fortitude DC 38, 8d10 spirit: nothing on a critical success, full damage on a success, full damage and drained 1d4 on a failure, double damage and drained 4 on a critical failure. Not a basic save. The spell text is Player Core p. 367 (overview Decision 11). The 3-action version and the lingering round stay dropped (rules scope, section 2).
2. **Faithful Strike is spent by the next Strike or damaging spell against the Avatar, hit or miss.** The text, not the Foundry rule elements. It ends unused at the end of the holder's next turn.
3. **A recharge counts down at the start of the user's turn; a roll of 1 means ready next turn.** Destiny Spores, Wail and Searing Wave share it.
4. **At the start of the Avatar's turn, Last Stand is checked first, then Fast Healing.**
5. **Impending Fate and Fey Momentum fire on critical Strikes only**, not on other attack rolls.
6. **Vengeful Spite's frightened and stunned follow the save result alone.** Terrifying Touch's "not already frightened" limit does not carry over.
7. **Sin Vulnerability: once per creature per initiative round; fires on a `loyalty` ability, a revive of an ally, and a passed save against a boss effect marked for it.** Faithful Strike goes to the acting PC. No step-between ability in sim core; `A.004.2` adds it with position.
8. **Hands are not run.** Drinking or feeding a potion or elixir is 1 action. No draw, no re-grip, no dropped weapon. Robert's table rule; it beats rules-as-written wherever he states one.
9. **A knocked-out PC is prone; Stand costs 1 action.** Too Angry to Die is a reaction on that Stand, as written, and leaves the stand-ins.
10. **Starting counts come from the sheets: Kas 3 Moderate and 1 Greater healing potion; Tavish 1 Elixir of Life and 1 potion; Arabesque 5 Therapeutic Snap Peas; Tavish 6 quick bombs.** Each is one line in a hand layer and a dial for `A.004.2`.
11. **One relic gift per bomb. Searing Wave's cone fires on a hit or a miss**, not on a critical failure, and catches foes only.
12. **Attack of Opportunity fires on every manipulate action, ranged Strike and Stand by a foe.** Every creature counts as engaged with every foe. Other move actions wait for position. A critical hit disrupts a manipulate action: the actions and the ability's use are both lost.
13. **Position stand-ins.** Walk the Timelines is two Strikes and no move. Farabellus Flip is the +2 AC without the Step. Lash Out is the damage without the shove. Return Force waits for `A.004.2`. Healing Wave catches Kas and every ally, dying ones included. Ring of the Ram's push is dropped.
14. **Challenge Fate answers Strikes and spell attacks.**
15. **Comforting Presence is a reaction that helps the triggering ally only.**
16. **Skill modifiers are computed from the build.** The importer reads skill ranks; the load computes each modifier. Rejected: a number typed in the hand layer.
17. **Under Bewitching Bloom (Cherry Blossom) a PC rolls recovery against 10 or its own DC, whichever is lower.**
18. **"Once per day, hour or minute" is once per fight** (D-25's reading, applied to items). Slowed "for 1 round" is the existing `rounds = 1` duration.

**Shape decisions.**

19. **An ability's limits are one `limit` table: `per_fight`, `per_round`, `recharge` and `pool`.** A pool is a named count on the creature that several abilities spend, so "drink a potion" and "feed a potion" share one number. Once per turn is the `flourish` trait [now once per round; see Decision Record — Execution 10], and `press` needs an earlier attack this turn; both are traits the engine reads, as PF2e writes them. Rejected: a `requires` list beside the traits (two spellings of one rule).
20. **A triggered ability is an ordinary ability with a `trigger` table.** `on` names one of ten kinds; `cost` is `reaction`, `free` or unset. An unset cost is a passive effect and always fires. A reaction or a free action is a decision.
21. **The decider gains a third question, `React`.** The engine gives it the trigger and the IDs of the abilities that could answer; it returns one or none. `FirstFitting` takes the first. Effort 7's rule lists answer it behind the same seam (hand-off item 3). Limits are exported state: `Combatant.Blocked`, `UsesLeft`, `RechargeLeft` and `Pool` [`Blocked` sits on the combatant; see Decision Record — Execution 65].
22. **Simple riders stay step lists; triggers are for effects that need a decision or cross creatures.** Impending Fate and Fey Momentum are steps in the `critical_success` list of Luna's Strikes and in Destiny Spores' `critical_failure` list. Rejected: a trigger kind for the owner's own Strike result (a second way to say what a rider says).
23. **A relic gift or a bomb additive is its own ability, not an add-on to a Strike.** A Flare Bolt bomb is a 2-action ability whose Strike step carries `extra_damage`. One gift per bomb (Decision 11) makes each combination one ability. Rejected: an add-on cost on the choice (a new decision shape for three abilities).
24. **The multiple attack penalty applies only on the attacker's own turn.** A reaction Strike on another creature's turn takes no penalty and does not count, as the rules say. `counts = 2` makes a Strike count as two attacks (Vicious Swing without Furious Focus), and `penalty = "shared"` reuses the previous Strike's penalty without counting again (Walk the Timelines).
25. **Quickened's extra action is for a Strike only.** Haste and the Warpwave table both limit it to Strike and movement, and movement waits for position. Last Stand's extra action is a `gain_actions` step with no limit on its use.
26. **Events are added, never changed.** Each new thing has its own event that names who, which ability and what set it off (hand-off item 1): `Triggered`, `Rerolled`, `RechargeRolled`, `Recharged`, `UseSpent`, `PoolSpent`, `ActionsRegained`, `ActionsGained`, `ActionDisrupted`, `StoodUp`, `LoyaltyAct`, `TraitGranted`, `TraitSpent`, `SkillChecked` and the rest. `DamageDealt` and `Healed` gain one field, `TriggeredBy`, so damage from Sin Vulnerability can be credited to the PC who caused it.
27. **A Strike step's own degree lists now run.** Effort 3 validated them and never ran them: `runStep` called `strike` and dropped the step's lists. Brutal Finish's failure damage needs them. They run after the Strike's riders.
28. **An ability stops when its user falls.** With reactions, a user can drop in the middle of its own ability (Vengeful Spite on a critical hit). Every step loop checks that the user still stands.
29. **Commits 8 and 9 swap against the signed breakdown.** Ally targets and consumables come before loyalty, because an act of loyalty is an ally-targeted ability or a revive.
30. **Bellflower's last rule moves to Effort 5.** It helps recovery from confused, and confused arrives there with the spells that cause it. Confused, like immobilized, is not built here. Deafened is a name only.
31. **Demoralize's temporary immunity waits for Effort 5**, which builds temporary immunity. Until then a creature can be demoralized more than once; Effort 6 writes the real ability after Effort 5.
32. **Sustain is built here with no real user.** The overview puts it in this effort; Unfathomable Song and Devouring Void (Effort 5) use it. It is the smallest form: one Sustain per effect per turn, and the effect ends when a turn passes without one.
33. **The ability schema stays version 1.** Every key is an addition (E3 Decision 7).
34. **Six phases.** Turn economy; limits and Strike tweaks; triggers and reactions; rerolls, healing and loyalty; skills; data and wiki. The signed breakdown had five; Commit 3 alone is over 1,100 lines, so it is its own phase to keep each one inside the phase budget.
35. **An ability with two triggers is two abilities with the same steps.** Vengeful Spite fires on a critical hit or on a critically failed save. One reaction pays for either. Rejected: a list of triggers on one ability (every other ability would write a one-item list).
36. **Three small format words carry the stand-ins.** `target = "other"` on a step picks a second foe (Walk the Timelines); `success_damage = "full"` on a basic save keeps full damage on a success (Wail, Decision 1), so the damage is still rolled once for all targets; a relic's cone is a save in the Strike step's `success` and `failure` lists, so it reaches the struck foe only, which is every foe while each phase has one boss.

## Decision Record — Execution

1. **Commit 2: a quickened creature that loses actions loses the quickened action first.** Robert's call (2026-11-19); the plan carried it only in a fence-sign in `regainActions`. Slowed 1 with quickened gives 3 open actions and no Strike-only one; quickened with stunned 4 loses all 4 actions and clears stunned in one turn. `TestRegainActions` has a case for each.
2. **Commit 2: a knock-down logs off-guard before prone.** Shared Context said the two `ConditionChanged` events come as prone, then off-guard. `changeConditions` walks `ConditionNames` in order, so off-guard comes first, then prone, then `DyingChanged`. `TestKnockedOutPCIsProne` checks that order. No existing test compares a whole event list across a knock-down, so none changed. Shared Context is amended to match.
3. **Commit 2: two test shapes made exact.** (a) `TestSlowedForOneRound` slows through a Fortitude save that fails in round 1 and critically succeeds in round 2. Limits are not enforced until Commit 3, so the source uses the ability every turn; the save keeps the second slowed off. (b) `TestLoadReadsConditionValueDice` checks that a loaded condition step carries its parsed `ValueRoll` and `stack`, so the roll is on the creature and not only on a copy.
4. **Fix commit after Commit 2: a downed creature still regains actions.** Robert's call (2026-11-20) on the Player Core text: you still regain actions when you can't act unless an effect prevents it (pg. 415), regaining actions is the last step of starting a turn (pg. 435), and only death stops it (pg. 411). In `act`, a dying PC regains actions after its recovery check unless that check kills it, and a stable PC at 0 HP regains actions too. Neither acts. So stunned and slowed count down while a PC is down. `TestStunnedCountsDownWhileDown` covers the dying, stable and killed cases.
5. **Fix commit after Commit 2: stunned gained on the bearer's own turn takes its remaining actions at once.** Robert's call (2026-11-20) on Player Core pg. 442: a stunned creature with a value loses its remaining actions to lower the value, and does not wait for its next turn. `applyCondition` takes `min(stunned, actions left)` at once, the quickened action first (entry 1), and logs a new `ActionsLost` event. The ability in progress completes: its cost is paid and only the actions left change. `TestStunnedOnOwnTurn` covers it, with one case through `perform`.
6. **Fix commit after Commit 2: stunned can have a duration instead of a value.** Robert's call (2026-11-20) on Player Core pg. 446: "stunned for 1 minute" loses all your actions for the duration, and stunned overrides slowed. The loader takes `condition = "stunned"` with a `duration` and no value; with neither it reports "stunned needs a value or a duration", and `stack` needs a value. The engine holds it as a value-0 instance, which `regainActions` reads as the loss of every action (quickened included) and `reduceCondition` leaves for its clock to end. Gained on the bearer's own turn, it takes every action left. Inference, not quoted rule: a valueless stunned beside a valued one loses every action, and the valued one counts down by what it would have taken. `TestStunnedForDuration` and `TestLoadReadsStunnedForDuration` cover it.
7. **Commit 3: the shared penalty starts at 0 each turn.** `strikePenalty` sets `lastPenalty` and nothing reset it, so a `penalty = "shared"` Strike that is the first attack of a turn reused the last penalty of an earlier turn. `act` now sets `lastPenalty` to 0 where it sets `AttacksThisTurn` to 0. `TestSharedPenaltyStartsFreshEachTurn` runs a turn that ends at −10, then a shared Strike first in the next turn, which takes 0.
8. **Commit 3: an unset `Choice.Actions` spends the minimum, so it cannot be refused.** The test list said `Actions = 0` on a 2-action minimum fails; with the planned `perform`, 0 is the zero value and means "the ability's own cost". `TestVariableCost` instead refuses `Actions = 1` on a 2-to-3 ability and checks that an unset choice spends 2. The test list is amended to match.
9. **Commit 3: two tests sit where the source lets them.** `TestCriticalSpecializationRowIsCopied` is in `creature/build_test.go`, not the engine's `limits_test.go`, because the riders are built by the loader and the `creature` tests are black-box; it loads two hammers, changes one rider, and checks the other weapon and a fresh load still read prone. `TestFreeAbilityDoesNotSpinTheTurn` (not in the plan's list) runs a free, once-per-round ability through `Run` under `FirstFitting` and checks it is used once per turn and the fight ends normally.
10. **Fix commit after Commit 3: flourish and per-round uses share one clock, until the start of the user's next turn.** Robert's call (2026-11-20) on Player Core pg. 139, "You can use only one action with the flourish trait per round", and pg. 435, where a round runs from the top of the initiative order to its end. The reset stays in `tickLimits` at the start of the user's own turn, as a 1-round duration does; the strict top-of-initiative round is not built. `Blocked` now says "flourish already used this round", and a fence-sign in `tickLimits` names the reading. `TestFlourishAndPress` checks the new wording.
11. **Fix commit after Commit 3: `Recharged.By` names the Strike when no ability set it off.** Robert's call (2026-11-20). A recharge rider on a plain chosen Strike logged an empty `By`, the same as the turn-start countdown. The `recharge` step now passes the ability, or the Strike when the ability is empty. The Destiny Spores row in `TestScopeRows` checks `Recharged{Combatant: "abyssal-luna", Ability: "destiny-spores", By: "slash"}` after a critical Slash; `TestStrikeStepListsRun` checks `By: "fist"`; `TestRecharge` still checks an empty `By` at turn start and `By: "refresh"` for a step inside an ability.
12. **Fix commit after Commit 3: `turn_ends` clears the Strike-only action.** Robert's call (2026-11-20). `perform` set `ActionsLeft` to 0 and left `StrikeOnlyActions`, so a quickened creature had −1 open actions for the rest of the turn. It now sets both to 0. `TestTurnEnds` has a quickened case: `ActionsLeft` 0 and `OpenActions()` 0.
13. **Known gap: press outside your turn.** Player Core pg. 139: "You can't use a press action when it's not your turn." `AttacksThisTurn` keeps its value between turns, so a reaction with `press` would pass `Blocked`. No ability in scope is a press reaction. Robert's call (2026-11-20): logged, not built.
14. **Known gap: press on a success.** Player Core pg. 139: "If your press action succeeds, you can choose to apply the failure effect instead." Not modelled; it needs a decider question. Robert's call (2026-11-20): logged, not built.
15. **Known gap: steps in a Strike's degree lists skip the effect-immunity check.** `runStep` skips `affected` when `source.strike` is set (E3 behaviour, unchanged). Searing Wave's and Debilitating Bomb's saves rely on it; damage immunity and condition immunity by name still apply. Robert's call (2026-11-20): a design question for Effort 6, which writes those abilities for real.
16. **Known gap: `gain_actions` after a stun in the same ability.** If a creature is stunned on its own turn partway through an ability, the stun takes the actions left at that moment, and a later `gain_actions` step adds actions the stun never sees, even a duration stunned. Unreachable with today's data.
17. **Commit 4: extra reactions come back inside `regainActions`.** The plan set `extraReactions` in `act`, in the standing branch only. After entry 4 a dying or stable PC regains actions too, from two other call sites, so the line sits at the top of `regainActions` and all three get it. Tactical Reflexes: "At the start of each of your turns when you regain your actions, you gain an additional reaction" (Player Core pg. 148). `TestExtraReaction` checks a stable and a dying owner, and that the extra reaction is not there before the owner's first turn.
18. **Commit 4: a triggered ability's Strike names its cause.** `strike` builds its own `effectSource`, so a reaction Strike (Attack of Opportunity) logged `DamageDealt` with an empty `TriggeredBy`, against Decision 26. `strikeOptions` gains `triggeredBy`; the strike case of `runStep` passes `source.triggeredBy`, and `strike` puts it in its source. `TestTriggeredDamageNamesItsCause` has a damage step, a Strike step and a heal step.
19. **Commit 5: the Vengeful Spite row follows entry 5.** The test list said a critically failed save left Kas with 0 actions and stunned 1 on his next turn. Entry 5 makes stunned gained on his own turn take his remaining actions at once. In the row Kas uses Vicious Swing first, so 1 action is left: stunned 4 takes it (`ActionsLost{Lost: 1, Left: 0}`) and drops to 3; his next turn regains 0 actions with 3 lost (`ActionsRegained{Actions: 0, Lost: 3}`) and stunned ends; the turn after has 3 actions and no `ActionsRegained`. The test list is amended to match.
20. **Commit 6: a `reduce` step is not a stunned step with no value.** `condition = "stunned", reduce = 1` has no value and no duration, so the Phase 1 case "stunned needs a value or a duration" reported it before the planned `reduce` case was reached. `stunnedWithNoValue` now needs `reduce` unset, so the `reduce` case decides. `TestLoadReadsReduceDisruptsAndFastHealing` loads a stunned `reduce = 2` cleanly.
21. **Reactions before a creature's first turn are allowed.** Robert's table rule (2026-11-20). Player Core pg. 436 leaves it to the GM: "The GM determines whether you can use reactions before your first turn begins" (https://2e.aonprd.com/Rules.aspx?ID=2432). As built: the shared reaction is there from the start (`ReactionUsed` starts false); Tactical Reflexes' extra reaction arrives at the first turn, as its own text says. No code change. `TestExtraReaction`, "not there before the owner's first turn", spends the shared reaction before the owner's first turn and is offered nothing after.
22. **Fix commit after Commit 6: fast healing works on a downed creature.** Robert's call (2026-11-20) on Player Core pg. 410: a creature with fast healing "regains the listed amount of Hit Points each round at the beginning of its turn" (https://2e.aonprd.com/Rules.aspx?ID=2322), with no exception for 0 HP or dying. `fastHeal` now heals any creature that is not dead and is below its maximum HP. In `act`, a creature that starts its turn at 0 HP gets its fast healing after `tickLimits` and before the recovery check; healed above 0, it is revived (dying cleared, wounded raised) and takes the standing branch in the same turn, still prone. A local `healedWhileDown` keeps it to one heal a turn; a standing creature's order is unchanged. Two inferences, not quoted rules: (a) healing comes before the recovery check, because pg. 435 lets the start-of-turn steps run "in any order you choose" and healing first is never worse; (b) the healed creature acts this turn, because the book regains actions as the last start-of-turn step, after the healing. `TestFastHealingRevivesTheDowned` (dying 2: one `Healed` to 10, `Revived{DyingFrom: 2, Wounded: 1}`, no recovery check die, then Stand and two Strikes; stable: one heal, acts; dead: nothing) and `TestFastHealing` cover it; `TestTurnStartBeforeFastHealing` is unchanged.
23. **Fix commit after Commit 6: a trigger chain nested too deep ends the fight.** Robert's call (2026-11-20). A passive `ally-gains` ability that stacks a condition on the triggering ally fires itself without end, and two passive `damaged` abilities that deal damage bounce until one side falls. `fire` counts its nesting on `Fight.triggerDepth`; at `maxTriggerDepth` (16) it sets a fault that wraps the new `ErrTriggerLoop` and names the owner and the kind (`trigger loop: b: ally-gains nested more than 16 deep`) and fires nothing. `TestTriggerLoopEndsTheFight` runs both chains through `Run` to `NoOutcome` and the exact error; `TestReactionsDoNotLoop` passes unchanged.
24. **Fix commit after Commit 6: the first fault is kept.** `fire` overwrote `fight.fault`, so a second bad `React` in one turn replaced the first. Both places that set it now go through `setFault`, which sets it only when it is nil. `TestFirstFaultIsKept` gives two bad answers and checks that `run` returns the first.
25. **Known gap: start-of-turn order.** Player Core pg. 435 (https://2e.aonprd.com/Rules.aspx?ID=2428): "The last step of starting your turn is always the same. Regain your 3 actions and 1 reaction." The engine regains actions first, then fires turn-start triggers, then fast healing, because Last Stand's `gain_actions` needs the actions to exist. One edge differs: a creature whose stunned value ends at this regain can use a turn-start free action that the book would deny. Robert's call (2026-11-20): logged, not built.
26. **Known gap: one free action and one reaction per trigger.** Player Core pg. 15 (https://2e.aonprd.com/Rules.aspx?ID=2024): "you can use only one free action per trigger", beside the one reaction a round, so the book allows one of each on the same trigger. The engine offers one pick from both. No creature in scope has both for one trigger. Robert's call (2026-11-20): logged, not built.
27. **Known gap, for Effort 5: Reactive Strike on a ranged spell attack.** Player Core pg. 138 (https://2e.aonprd.com/Actions.aspx?ID=2256) triggers when a creature "makes a ranged attack"; ratified Decision 12 and the code cover ranged Strikes only, because a `spell_attack` step has no range yet. Effort 5 defines spells and takes this up.
28. **Data fact: Lash Out deals nothing to the Avatar.** Its 2d6 void rolls 12 at most, against her resistance 12 to all damage but holy and cold iron. Not a bug. Effort 7's rule lists should know it before they spend Arabesque's reaction; the Lash Out row in `TestScopeRows` asserts `Rolled 12, Taken 0`.
29. **Known gap: an ally's `reduce` on drained runs before the drain's HP loss.** `applyCondition` tells allies right after the new instance is added, so a `reduce` on drained would lower the value before `drain` works out the HP lost from `drainedBefore`. Unreachable with today's data: no `ally-gains` ability lists drained.
30. **Commit 7: of two fortune (or two misfortune) effects on one roll, the first gained applies.** Player Core pg. 401 (https://2e.aonprd.com/Traits.aspx?ID=654, https://2e.aonprd.com/Traits.aspx?ID=612): "You can never have more than one misfortune effect alter a single roll. If multiple misfortune effects would apply, the GM decides which is worse"; for fortune, "you have to pick which to use". `takeReroll` takes the first matching effect; the other keeps its use. When fortune and misfortune cancel, each still loses a use: an inference, since the text says only "you roll normally". No creature in scope holds two of one kind. `TestFortuneAndMisfortuneCancel` covers both cases. Robert's call (2026-11-20): logged, not built; both readings stay as known gaps.
31. **Known gap: a one-attack effect outlives an attack that never rolls.** Farabellus Flip's AC bonus and Challenge Fate's reroll come from a `targeted` reaction and are spent by the next roll of their kind. If the attacker falls between the `targeted` trigger and the roll, the effect waits for a later attack. Unreachable with today's data: nothing in scope answers `targeted` with damage. Robert's call (2026-11-20): logged, not built.
32. **Every d20 in the engine, accounted for.** After Commit 7, `fight.d20` rolls attack rolls (`strike`, `spellAttack`) and saves (`savingThrow`). Four `roller.Die(20)` calls stay outside, on purpose: the concealment flat check (`passesConcealment`), the persistent damage flat check (`persistent.go`), the recovery check (`recoveryCheck`) and initiative (`rollInitiative`). `TestRerollOnlyItsRolls` checks that an attack reroll leaves a save, a flat check and a recovery check alone.
33. **Commit 8: `perform` checks each choice kind in its own order.** The target check moved into the cases, merged onto the `perform` that the review fixes left (Stand's `provoke`, the `TurnEnds` block that zeroes `StrikeOnlyActions`). Stand: prone and an open action. Sustain: a waiting effect not yet sustained this turn and an open action. Strike: the Strike exists, then a standing target. Ability: it exists and is not ignored or triggered, then `checkTarget`, cost, open actions, `Blocked`. An unknown kind is now refused as unknown even with no target; before, the missing target was named first. No existing test read that message.
34. **Fix commit after Phase 4: the loader rejects `targets = "self"` with an area.** Robert's call (2026-11-20). `areaTargets` returns the foes for any `targets` but `ally` and `others`, so a self ability with an area would have hit foes. `prepare` now reports `ability.<id>: targets self takes no area`; `TestLoadRejectsBadTargetsAndSustain` has the case. Known gap, not built: a triggered ability may set `targets`; it then never moves to a new target and skips the concealment check. None does.
35. **Fix commit after Phase 4: a sustained effect lasts to the end of the user's next turn, even if the user drops.** Robert's call (2026-11-20), by the book. Duration (https://2e.aonprd.com/Rules.aspx?ID=2378): "An effect with the sustained duration lasts until the end of your next turn". `endSustains` no longer reads `actor.Standing()`: an effect sustained this turn carries over, and one not sustained this turn ends. A downed user cannot Sustain, so its effect ends at the end of its next turn. `TestSustain` asserts no `SustainEnded` at the end of the turn the user drops, then `SustainEnded` at the end of its next turn, spent dying. Two more readings from Sustain (pg. 419, https://2e.aonprd.com/Actions.aspx?ID=2317) are known gaps for Effort 5, which has the two real sustained spells (Unfathomable Song, Devouring Void): "Sustaining more than once in the same turn doesn't extend the duration to subsequent turns", so a second Sustain in a turn is allowed, and the engine refuses it (Decision 32); "If your Sustain action is disrupted, the ability ends", and the engine's Sustain does not `provoke`. Robert's call, 2026-11-20.
36. **Commit 9: a fed potion that revives an ally is one act of loyalty, and its reason is `revive`.** Feeding is two acts in one action: the `loyalty` tag and the revive. The revive runs first, inside `heal`, and sets the round mark; the tag check after `useAbility` in `perform` then finds the mark set. Sin Vulnerability fires once. The "Feeding a potion" row in `TestScopeRows` and `TestReviveIsAnActOfLoyalty` assert one `LoyaltyAct{Reason: "revive"}`.
37. **Commit 9: a disrupted consumable spends its pool and is no act of loyalty.** `perform` runs `spend` before `provoke`, so a fed potion lost to a critical Attack of Opportunity still costs its count, heals nothing and fires nothing (ratified Decision 12: the use is lost). `TestDisruptedAbilityIsNoAct` checks `PoolSpent{Left: 0}`, no `Healed` and no `LoyaltyAct`.
38. **Fix commit after Phase 4: Haste reaches the caster.** Robert's call (2026-11-20), by the book. Player Core Haste (https://2e.aonprd.com/Spells.aspx?ID=1553): "Target 1 creature", and "Heightened (7th) You can target up to 6 creatures"; nothing keeps the caster out. The scope data's `targets = "others"` from Commit 8 becomes `targets = "ally"`, so the emanation reaches Arabesque, Kas and Tavish. The Haste row in `TestScopeRows` checks quickened for 10 rounds from Arabesque on all three; the loyalty Haste row still has one `LoyaltyAct{arabesque, haste}`, and the granted holy goes to Arabesque, the acting PC.
39. **Known gap, for Effort 5: a sustained effect is not recorded when its user drops during the first use.** `perform` records the effect only if the user still stands after `useAbility`, so a caster dropped mid-cast gets no effect, where entry 35's reading says it lasts to the end of the user's next turn. No sustained ability exists until Effort 5, which takes this up with entry 35's two gaps.
40. **Before Commit 10: the campaign's PC exports are replaced; the test copies stay frozen.** Robert's call (2026-11-20). `campaigns/hawk/encounters/fall-of-luna/source/actors/` takes the exports of 2026-11-17 for Kas, Tavish and Arabesque. `internal/pf2e/sim/foundry/testdata/actors/` keeps the exports of 2026-11-10, the snapshot the gate 1 fixture names. The builds and skill ranks are the same in both; only sheet state differs (conditions, effects, Tavish's Nail Bomb count), and four importer tests are pinned to the old state. Rejected: refresh the test copies and rewrite the four tests (a fixture should not move each time a sheet is exported again).
41. **Before Commit 10: the skill fixture is in, and three importer additions join Commit 10.** Robert read all 48 skill modifiers from the sheets (gate 1 fixture, section 6). They need three things the planned importer lacks, each Robert's call (2026-11-20). (a) A background or class item trains skills on the item (`system.trainedSkills.value`), and the actor's stored rank stays 0: the importer raises each listed skill to at least trained (Tavish's Thievery, from Saboteur). (b) Untrained Improvisation is a `proficiency`-type bonus to every skill check: the build takes `proficiency` as a bonus type on skills, a skill uses the higher of it and its rank's own bonus, and the formula reader learns `match`, `when`, `lte` and `gte`. Rejected: a special case for the feat; a typed number in the hand layer. (c) A hand-entered modifier on the actor (`system.customModifiers`, Tavish's +1 Stealth) is read for skills; one on anything else is reported as unsupported.
42. **Resolved in Commit 12: a bomb with no stock that the PC can brew is imported.** Robert's call (2026-11-20). Nail Bomb (Greater) has quantity 0 on the new export, but Tavish brews bombs with Quick Alchemy, so the count on the sheet does not decide what he can throw; the quick-bombs pool (ratified Decision 10) limits how many. It is his cold-iron bomb, one of two ways through the Avatar's resistance. The rule: a weapon with quantity 0 is still imported when (1) it has the `bomb` trait, (2) the actor has an item with the slug `quick-alchemy`, and (3) the bomb's compendium origin (`_stats.compendiumSource`, or `flags.core.sourceId` on older data) equals the `uuid` of an entry in `system.crafting.formulas`. Any other quantity-0 weapon is skipped as `quantity 0`, as before. A quantity-0 bomb with Quick Alchemy but no matching formula, or no origin recorded, is skipped with the new reason `quantity 0 and no known formula to brew it`. A brewed bomb is reported under the hand-layer reason as "quantity 0, no stock: brewed with Quick Alchemy from the quick bombs pool", in place of Task 2's "typed as a pool" line. Rejected: Quick Alchemy only, with no formula check (it would import any bomb on the sheet, known or not); typing the bomb in the hand layer (a hand-written copy of an imported weapon that drifts from the sheet). In the frozen test data the rule brings in Frost Vial (Moderate) and Bottled Lightning (Moderate), both at quantity 0 with known formulas, so `TestImportPCWeaponsAndQuantities` counts 14 weapons, not 12. In the new export it brings in those two and Nail Bomb (Greater). `TestImportBrewedBombs` (known formula, with `cold-iron` kept; origin in `flags.core.sourceId`; formula removed; no origin; no Quick Alchemy) and `TestImportQuantityZeroNonBomb` cover it.
43. **Commit 10: a skill's proficiency part is the higher of its rank's bonus and its best `proficiency` modifier, never both.** Entry 41 (b). The book's rule that bonuses of one type do not stack, with the rank's own bonus counted as a proficiency bonus. `Build.skills` takes `max(proficiencyBonus(level, rank), proficiencyModifier(...))` under a fence-sign; `total` skips `proficiency` modifiers so none reaches the typed pool; the loader takes `type = "proficiency"` on a skill or `skill-check` selector only, and the importer the same. Untrained Improvisation imports as a flat +14 on `skill-check` (the formula is evaluated at import, like every other formula). `TestSkillsUseTheHigherProficiency`, `TestLoadRejectsBadSkills`, `TestImportProficiencyModifier` and `TestGateSkills` cover it.
44. **Commit 10: a background's trained lore is reported as not a combat statistic, not dropped silently.** Entry 41 (a). Lore skills stay out of the sim (Task 1), and the importer already reports a lore `FlatModifier` (Aethership Navigation) with that reason; a silent drop would break trap 4. `additional` needs nothing: the free picks are in the stored ranks. `TestImportTrainedSkills` checks Warrior's Warfare Lore and Saboteur's Engineering.
45. **Commit 10: a disabled or ignored hand-entered modifier is reported as "toggle, imported off".** Entry 41 (c). Of the existing reasons it is the nearest: the sheet holds the entry switched off, and the import keeps it off. An entry with `force`, traits, `ability`, `damageType`, `damageCategory` or `critical` set is reported as not supported, beside the adjustments and unknown types entry 41 names; none of these changes a plain skill bonus in our data, and each would change how it stacks or where it lands. `TestImportCustomModifiers`.
46. **Commit 10: two small names added to the plan's code.** `creature.IsSkillSelector` (a skill or `skill-check`) and `creature.ProficiencyModifier` (`"proficiency"`), used by the build check, `skills`, `flatModifierFor` and the hand-entered modifiers, so the test is written once. Task 1 and Task 3 are amended to match.
47. **Commit 10: existing importer tests gain the new skill modifiers.** Demon Mask's +2 Intimidation was reported as not a combat statistic and is now a build modifier, so `TestImportPCRuleElementsKas` checks it in the modifiers and checks Aethership Navigation for the not-a-combat-statistic reason instead. `TestImportPCRuleElementsTavish` gains Artisan's Mark, Untrained Improvisation and Shadow Property; `TestImportPCRuleElementsArabesque` gains Ring of Wizardry. `TestPrintCreature` gains a skills line for Task 4.
48. **Commit 11: no existing test changed for the split `dc` message.** Task 1 planned an update to the `steps_test.go` case that reads the old message, but "shift on a strike" checks only "belong on a save", and "dc_from, basic and shift belong on a save" still holds it. `TestLoadRejectsBadChecks` checks both new messages.
49. **Commit 11: Too Angry to Die fires on a Stand after a fast-healing revive.** The `downed` mark is set by the knock-down and cleared only by `stand`, so a creature lifted by its own fast healing (entry 22) and then Standing still fires `stood-up`. Too Angry to Die's trigger is "you stand up from prone after waking up from being unconscious" (Kingmaker Companion Guide pg. 9, https://2e.aonprd.com/Feats.aspx?ID=3899). The first Too Angry to Die row in `TestScopeRows` runs it through `takeTurn`.
50. **Commit 11: the Battle Cry row runs through `newFight` and `run`, not `Run`.** `Run` hides the combatants, and `act` resets `ReactionUsed` at the start of Kas's turn, so the row records his reaction when round 1 starts through a small observer (`roundWatch`), after the initiative trigger and before any turn.
51. **Fix commit after Phase 5: Demoralize is visual.** Robert's call (2026-11-20), by the book. Intimidating Glare, Player Core pg. 257 (https://2e.aonprd.com/Feats.aspx?ID=5162): "Demoralize loses the auditory trait and gains the visual trait". The scope data's `demoralize`, `battle-cry` and `too-angry-to-die` now carry `visual`, so a target immune to `visual` (an effect-immunity trait the engine already reads) is skipped. The Demoralize row in `TestScopeRows` has a case: the Avatar made immune to `visual` gives `ImmuneToEffect{avatar-of-lilith, kas, demoralize}`, no roll and no frightened.
52. **Known gap, for the Effort 5 plan session: a critical result with no list of its own.** Player Core pg. 401 (https://2e.aonprd.com/Rules.aspx?ID=2286): "If an effect doesn't list a critical success effect, the critical success effect is the same as the success effect, and the same goes for critical failures." The engine does this for attack rolls only (`Outcomes.For` with `attack` set); a save or a check with no list for its critical degree runs nothing, and `TestSkillCheckAgainstADefense` asserts that for a check. Going by the book needs a format answer first: how a hand layer says "a critical success does nothing", since an empty list does not survive the layer merge. All of today's data lists every degree. It would change the engine Key Decision "Abilities are recursive steps with a list of steps per degree of success" (`docs/architecture/pf2e/engine.md`, `F.041.E3` Decision 6). Robert's call (2026-11-20): logged, not built now.
53. **Known gap: the armor check penalty and the attack trait.** Player Core pg. 271 (https://2e.aonprd.com/Rules.aspx?ID=2173): "you take this penalty to Strength- and Dexterity-based skill checks, except for those that have the attack trait." The penalty is in the stored skill number, so a Trip, Shove or Grapple check would carry it. No such check exists yet. Robert's call (2026-11-20): logged.
54. **Known gap: Intimidating Prowess needs physical menace.** Player Core pg. 257 (https://2e.aonprd.com/Feats.aspx?ID=5163): "In situations where you can physically menace the target when you Coerce or Demoralize, you gain a +1 circumstance bonus to your Intimidation check". The scope data's `[modifier.intimidating-prowess]` is unconditional on Intimidation; only Demoralize-style checks exist. Robert's call (2026-11-20): logged.
55. **Known gap: Untrained Improvisation is the legacy text.** Core Rulebook pg. 268 (https://2e.aonprd.com/Feats.aspx?ID=861): "half your level … at 7th level or higher … your full level"; Player Core pg. 264 (https://2e.aonprd.com/Feats.aspx?ID=5233): "your level –2. This improves to your level –1 at 5th level and your full level at 7th level". Both give +14 at level 14; they differ at levels 5 and 6 (legacy 2 and 3, Remaster 4 and 5). The importer reads what the sheet's formula says, which is the legacy text. Robert's call (2026-11-20): logged.
56. **Note: an imported formula is a number fixed at import.** Untrained Improvisation is stored as 14, so a later layer that levels a PC would not move it. The same holds for every imported formula today.
57. **Commit 12: `show` and the import report print fast healing.** Task 4 checks that the Avatar shows fast healing 10, but `printCreature` had no line for it, so the check could not pass. It now prints `Fast healing 10` after the adaptive resistance line when the value is above 0. `TestPrintCreatureShowsComputedNumbers` checks the line, and that no line prints at 0.
58. **Commit 12: a brewed bomb's own rule elements now apply.** Entry 42 puts a quantity-0 bomb in `weaponByItem`, so its `AdjustStrike` materials rule finds it: the Nail Bomb keeps `cold-iron`. A skipped quantity-0 weapon still has its `AdjustStrike` reported as not supported, as before. `TestImportBrewedBombs` runs through `importPC`, not `weapon` alone, so the cold-iron case crosses this seam (trap 1). Known gap: a brewed bomb that is neither held nor worn has its rule elements reported as inactive, as for any carried weapon; every bomb in our data is held.
59. **Commit 12: the fast healing guard is tested through `importNPC`.** `TestImportNPCFastHealing` replaces the Avatar's Fast Healing rule with five copies: `type = "fast-healing"` (read), `type = "regeneration"` with `deactivatedBy` (not read), a non-empty predicate (not read), an empty predicate (read) and a formula value (not read). A rule that is not read leaves the item in the report as an ability. `TestImportNPCSkills` also checks the Avatar's three skills and that a non-skill key (`sailing-lore`) is reported as not supported.
60. **Commit 12: the `DamageDice` reason is checked by name and text.** The rule-element tests for Kas and Tavish gain `checkDamageDice`: Vicious Swing and Brutal Finish, and Tidal Crash Round, Magnetic Shot (Moderate), Relic: Tidal Crash and Relic: Flare Bolt, each with `What: "DamageDice"` and `ReasonAbilityPart`, and no `DamageDice` row with another reason. `TestImportBombDamage` needed no change: it reads no reason that moved.
61. **Commit 12: sheet stock does not set a pool; the pool is a dial.** Robert's call (2026-11-20). The new export holds 3 infused reagents (`system.resources.crafting.infusedReagents.value`), the frozen one 6. Legacy Quick Alchemy costs "1 batch of infused reagents" (Core Rulebook pg. 72, https://2e.aonprd.com/Actions.aspx?ID=2). The 3 is mid-adventure sheet state, like the spent Nail Bombs, so the quick-bombs pool stays 6 (ratified Decision 10). The importer does not read reagents.
62. **Known gap, for Effort 6: a bare bomb Strike spends nothing.** Robert's call (2026-11-20): logged, no code now. The quick-bombs pool limits only the `quick-bomb` abilities, which strike `alchemists-fire-greater`. A bare Strike with any imported bomb, the brewed Nail Bomb included, has no limit, so the sim can throw it every action. The import report's "quantity N, typed as a pool" line promises pools that no hand layer types; its wording stays. The fix is a design question for Effort 6: each bomb becomes an ability on a pool, and a bare bomb Strike stops being a choice. Also in the overview's Findings for the effort plans.
63. **Note: the Remaster's Quick Alchemy spends versatile vials.** Player Core 2 pg. 59 (https://2e.aonprd.com/Actions.aspx?ID=2801): "You expend one of your versatile vials to create a single alchemical consumable item of your level or lower that's in your formula book." The PCs are legacy builds; entry 42's formula rule matches the legacy text (Core Rulebook pg. 72: "You have the formula for the alchemical item you're creating"). No action.
64. **Note: an NPC skill is read from `base` only.** Both boss exports hold `base` and nothing else in each `system.skills` entry, and the importer reads only `base`. No action.
65. **Commit 13: two plan statements corrected against the source.** Ratified Decision 21 named `Fight.Blocked`; `Blocked` is a method on `Combatant`, beside `UsesLeft`, `RechargeLeft` and `Pool`. Task 1 said the step kinds grow to 16; there are 15 (`creature.stepKinds`), and the Strike step's tweaks are keys on the Strike step. Both plan lines are amended; `engine.md` follows the source.
66. **Commit 13: the wiki and the rules scope go a little past the task list.** `overview.md`'s position line named Effort 7; it now names `A.004.2`, which the engine page also uses. `engine.md` gains a short Known gaps list that points at this Decision Record. In the rules scope, the six named approximations this effort adds (Attack of Opportunity, Farabellus Flip, Lash Out, Ring of the Ram, Healing Wave, Searing Wave) also change their section 3 model to `approximation`, so the two sections agree; section 6 carries the execution-time rules calls too, with table rules and inferences marked.
67. **Commit 13: the hand layers load as copied.** The five real folders under `campaigns/hawk` took the scope files' ability, pool, modifier and extra-reaction tables, and Luna's two Strike `critical_success` lists, as the Task 5 table says; Luna keeps `abyssal-frenzy`. A throwaway test loaded all five through `creature.Load` with no error (Luna 5 abilities, the Avatar 5 and `fast_healing` 10 from the import, Kas 14, Tavish 12, Arabesque 6), then was deleted. Every Strike ID the abilities name exists in the import layer, so nothing needed fixing. The same test loaded the triggered-ability example in `data-files.md`.
68. **Known gap, a rules gap and not a decision: a triggered ability never provokes.** `provoke` runs for a chosen ability, a Stand and a ranged Strike; `runTriggered` never calls it, so a reaction or free action with the `manipulate` trait cannot be disrupted by Attack of Opportunity. Sustain's case is entry 35. No triggered ability in scope has `manipulate`. Robert's call (2026-11-20): logged, not built.
69. **Known gap: passive triggered abilities run first, in ID order.** `fire` walks the owner's abilities by ID and runs each matching passive (no `cost`) inline, then asks `React` about the reactions and free actions it gathered. A passive that changes state, such as one that drops the attacker, runs before the owner's reaction is offered. Robert's call (2026-11-20): logged, not built.
70. **Finding for Effort 8: an ordinary turn logs no action count.** `ActionsRegained` fires only when the creature is quickened or loses actions, so a plain 3-action turn has no event for its actions; the fight record cannot read a turn's action count from events alone. Also in the overview's Findings for the effort plans.
71. **Finding for Effort 6: the E4 abilities exist twice.** They sit in `internal/pf2e/sim/engine/testdata/scope/` (the scope tests) and in the real hand layers under `campaigns/hawk` (Commit 13), and the two copies can drift. Effort 6 names one as the source. Also in the overview's Findings for the effort plans.
72. **Fix commit after Phase 6: a spell attack records its penalty, and position points at `A.004.2`.** Robert's calls (2026-11-20). (a) `spellAttack` counted toward the multiple attack penalty on the attacker's own turn but did not set `lastPenalty`, so a `penalty = "shared"` Strike after a spell attack in the same turn reused an older Strike's penalty. It now sets `lastPenalty` in the same own-turn branch, as `strikePenalty` does. `TestSharedPenaltyAfterASpellAttack`: a Strike (0), then an ability with a spell attack (−5) and a shared Strike, which takes −5; it failed with [0 0] before the fix. `engine.md` says a shared Strike takes the penalty of the last attack, a Strike or a spell attack. (b) The rules scope's section 3 said position comes with E7 in 12 rows, and its basic-actions line said Effort 7; since the replan Effort 7 is rule lists and position is `A.004.2`, so all 13 now name `A.004.2`. `overview.md`'s page index no longer lists position under the sim page (Efforts 7 and 8), and `engine.md` names `A.004.2` beside immobilized.

---

## Open Questions — To Ratify at Implementation Time

| Question | Resolved |
|---|---|
| The skill fixture: the 16 skill modifiers of Kas, Tavish and Arabesque, read by Robert from the Foundry sheets at level 14 | Resolved 2026-11-20: Robert read all 48 from the sheets into the gate 1 fixture, section 6 (Decision Record 41) |
| Whether "damaging spell" for Faithful Strike needs more than a `spell` mark on the ability | Effort 5, which defines what a spell is. Until then Faithful Strike is spent by Strikes only |

## Shared Context

### Gate

Every commit passes `go build ./... && go vet ./... && go test ./...` before it lands. Commit 13 also runs `show` and `import` by hand (its tasks).

### Test conventions

- `creature` tests stay black-box (`package creature_test`) and write layers with `writeLayers` (`layers_test.go`).
- `engine` tests stay white-box (`package engine`) and drive rules with the scripted roller (`helpers_test.go`). Commit 3 adds a scripted decider there.
- `foundry` tests stay white-box and read only `testdata/`.
- **Scope rows.** Gate 2 for an E4 row means a test that runs the row's ability. The abilities live in `internal/pf2e/sim/engine/testdata/scope/<creature>/hand.toml`, one folder per Fall of Luna creature, written in the real format with small typed stat blocks. Each commit adds the abilities its rule completes and a case in `TestScopeRows` (`scope_test.go`). Commit 13 copies the ability tables into the real hand layers.
- **Traps.** Each commit's test list has at least one case from the plan-traps catalogue (the memory file `feedback_plan_traps.md`): a seam with older code, a state change with a follow-on check, or a rule's case that the Fall of Luna data does not hit. They are marked `(trap n)`.

### Whole-list event tests

Commit 2 makes a knocked-out PC prone. A test that compares a whole event list after a knock-down gains two `ConditionChanged` events (off-guard, then prone, in `ConditionNames` order) before `DyingChanged`. The gate 3 test compares a run with itself and needs no change.

### Running the app

Claude runs the binary only against `hawk-claude` and passes `--campaign hawk-claude`. Build it into the scratchpad: `go build -o "$SCRATCH/gm-toolkit" ./cmd/gm-toolkit`. Robert runs the import on `hawk` (E2 Decision 11).

### Files this effort creates

```
internal/pf2e/sim/creature/
  limits.go, limits_test.go                  (Commit 1)
  triggers.go, triggers_test.go              (Commit 4)
  skills.go, skills_test.go                  (Commit 10)
internal/pf2e/sim/engine/
  actions.go, actions_test.go                (Commit 2)
  limits.go, limits_test.go                  (Commit 3)
  scope_test.go, testdata/scope/             (Commit 3)
  triggers.go, triggers_test.go              (Commit 4)
  rolls.go, rolls_test.go                    (Commit 7)
  allies.go, allies_test.go                  (Commit 8)
  loyalty.go, loyalty_test.go                (Commit 9)
  skills_test.go                             (Commit 11)
internal/pf2e/sim/foundry/
  skills.go, skills_test.go                  (Commit 10)
campaigns/hawk/encounters/fall-of-luna/avatar-of-lilith/hand.toml (Commit 13)
```

### Effort close review addition

`/gestalt:effort-close F.041.E4` also checks that E4's shapes do not box in what follows:

- **Effort 5:** a spell is an ability with a rank and a slot; `limit` can take a slot pool, `Sustain` runs Unfathomable Song, and the `reroll` step runs True Target. Faithful Strike's "damaging spell" has a place to hook.
- **Effort 7:** `React`, `Blocked`, `UsesLeft`, `RechargeLeft` and `Pool` give a rule-list condition what it needs to ask.
- **Effort 8:** each new event names who, which ability and what set it off; damage share by source can be built from `DamageDealt` alone.
- **`A.004.2`:** every "all engaged" rule sits behind one function or one trigger field (`areaTargets`, `provoke`, the `traits` of a `foe-acts` trigger).

## Out of Scope

- **Spells** (Effort 5): slots, ranks, heightening, counteract, incapacitation, temporary immunity, random tables, confused, True Target, Quickened Casting, the reaction spells, Bellflower.
- **Writing every kept ability** (Effort 6). This effort writes the E4 rows only, as first drafts.
- **Rule lists** (Effort 7). `FirstFitting` still decides, and takes the first reaction offered.
- **Position** (`A.004.2`): reach, movement triggers other than Stand, shoves and pushes, the Step of Farabellus Flip, Return Force, Walk the Timelines' move, areas that catch a share of a side, immobilized.
- **Hands** (Decision 8): draw, re-grip and dropped items are not modelled, now or later.
- **The penalties of unconscious.** No creature can target a downed PC, so they never apply.
- **Untethered to Fate, Double Brew, Perpetual Infusions, Unstable Concoction:** dropped in the rules scope, sections 2 and 5.
- **A cap on a sustained effect's length.** No fight in scope lasts the 10 rounds a 1-minute spell allows.
- **Hero-point rerolls.** Hero points still pay only for Heroic Recovery (E3 Decision 20b).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Turn economy**
  - Commit 1 — `feat(E4): add action costs, limits and pools to the ability format` (5 tasks)
  - Commit 2 — `feat(E4): set a turn's actions from conditions and add Stand` (9 tasks)
- [x] **Phase 2 — Limits and Strike tweaks**
  - Commit 3 — `feat(E4): enforce limits and add Strike tweaks` (11 tasks)
- [x] **Phase 3 — Triggers and reactions**
  - Commit 4 — `feat(E4): add triggers, reaction slots and the decider's React` (7 tasks)
  - Commit 5 — `feat(E4): fire triggers on attacks, saves and damage` (5 tasks)
  - Commit 6 — `feat(E4): fire triggers on turns, initiative, actions and conditions` (8 tasks)
- [x] **Phase 4 — Rerolls, healing and loyalty**
  - Commit 7 — `feat(E4): add fortune, misfortune and granted bonuses` (6 tasks)
  - Commit 8 — `feat(E4): target allies, share consumables and sustain effects` (7 tasks)
  - Commit 9 — `feat(E4): add acts of loyalty and granted traits` (6 tasks)
- [x] **Phase 5 — Skills**
  - Commit 10 — `feat(E4): compute skills from the build and import their ranks` (7 tasks)
  - Commit 11 — `feat(E4): add the skill check step` (5 tasks)
- [x] **Phase 6 — Data and wiki**
  - Commit 12 — `feat(E4): import fast healing and NPC skills` (4 tasks)
  - Commit 13 — `docs(E4): document triggers and limits and write the E4 abilities` (6 tasks)

The execution session ticks a phase's box in that phase's last commit.

### Phase 1 — Turn economy

**In plain words:** the ability format learns what an ability costs and how often it can be used, and a turn's action count now follows the creature's conditions. When this lands, slowed, stunned and quickened change a turn, and a knocked-out PC wakes prone and must Stand.

#### Commit 1 — `feat(E4): add action costs, limits and pools to the ability format`

##### Task 1 — `internal/pf2e/sim/creature/limits.go`

`internal/pf2e/sim/creature/limits.go` (new) — full contents:

```go
package creature

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

type Limit struct {
	PerFight     int             `toml:"per_fight,omitzero"`
	PerRound     int             `toml:"per_round,omitzero"`
	Recharge     string          `toml:"recharge,omitempty"`
	Pool         string          `toml:"pool,omitempty"`
	RechargeRoll dice.Expression `toml:"-"`
}

func (ability Ability) CostRange() (int, int) {
	if ability.Free {
		return 0, 0
	}
	return ability.Actions, max(ability.Actions, ability.MaxActions)
}

func (ability Ability) HasTrait(trait string) bool {
	return slices.Contains(ability.Traits, trait)
}

func (creature *Creature) checkPools(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.Pools)) {
		if creature.Pools[name] < 0 {
			report("pool.%s: must not be negative", name)
		}
	}
}

func checkCost(path string, ability Ability, report func(string, ...any)) {
	switch {
	case ability.Free && (ability.Actions != 0 || ability.MaxActions != 0):
		report("%s: a free ability takes no actions", path)
	case ability.Free && ability.Limit == nil:
		report("%s: a free ability needs a limit", path)
	case !ability.Free && (ability.Actions < 1 || ability.Actions > 3):
		report("%s: actions %d: want 1 to 3", path, ability.Actions)
	case ability.MaxActions != 0 && (ability.MaxActions <= ability.Actions || ability.MaxActions > 3):
		report("%s: max_actions %d: want more than actions and at most 3", path, ability.MaxActions)
	}
}

func (creature *Creature) checkLimit(path string, limit *Limit, report func(string, ...any)) {
	if limit == nil {
		return
	}
	if limit.PerFight < 0 || limit.PerRound < 0 {
		report("%s.limit: per_fight and per_round must not be negative", path)
	}
	if limit.PerFight == 0 && limit.PerRound == 0 && limit.Recharge == "" && limit.Pool == "" {
		report("%s.limit: sets nothing", path)
	}
	if limit.Recharge != "" {
		limit.RechargeRoll = parseAmount(path+".limit.recharge", limit.Recharge, report)
	}
	if _, found := creature.Pools[limit.Pool]; limit.Pool != "" && !found {
		report("%s.limit: unknown pool %q", path, limit.Pool)
	}
}
```

##### Task 2 — `internal/pf2e/sim/creature/creature.go`

The creature gains its pools; the ability gains its cost range, the free mark and the limit.

**Find:**
```go
	Strikes              map[string]Strike              `toml:"strike,omitempty"`
	Abilities            map[string]Ability             `toml:"ability,omitempty"`
}
```
**Replace with:**
```go
	Pools                map[string]int                 `toml:"pool,omitempty"`
	Strikes              map[string]Strike              `toml:"strike,omitempty"`
	Abilities            map[string]Ability             `toml:"ability,omitempty"`
}
```

**Find:**
```go
type Ability struct {
	ID      string   `toml:"-"`
	Name    string   `toml:"name,omitempty"`
	Actions int      `toml:"actions,omitzero"`
	Ignore  bool     `toml:"ignore,omitempty"`
	Traits  []string `toml:"traits,omitempty"`
	Area    *Area    `toml:"area,omitempty"`
	Steps   []Step   `toml:"steps,omitempty"`
}
```
**Replace with:**
```go
type Ability struct {
	ID         string   `toml:"-"`
	Name       string   `toml:"name,omitempty"`
	Actions    int      `toml:"actions,omitzero"`
	MaxActions int      `toml:"max_actions,omitzero"`
	Free       bool     `toml:"free,omitempty"`
	Ignore     bool     `toml:"ignore,omitempty"`
	Traits     []string `toml:"traits,omitempty"`
	Area       *Area    `toml:"area,omitempty"`
	Limit      *Limit   `toml:"limit,omitempty"`
	Steps      []Step   `toml:"steps,omitempty"`
}
```

##### Task 3 — `internal/pf2e/sim/creature/validate.go`

In `prepare`, the pools are checked with the fight rules, and the ability's cost and limit replace the fixed 1-to-3 check.

**Find:**
```go
	creature.checkFightRules(report)
```
**Replace with:**
```go
	creature.checkFightRules(report)
	creature.checkPools(report)
```

**Find:**
```go
		if ability.Actions < 1 || ability.Actions > 3 {
			report("ability.%s: actions %d: want 1 to 3", id, ability.Actions)
		}
```
**Replace with:**
```go
		checkCost("ability."+id, ability, report)
		creature.checkLimit("ability."+id, ability.Limit, report)
```

##### Task 4 — `cmd/gm-toolkit/pf2e/show.go`

`show` prints the cost and the limit, and names each step by its kind (closes E3 Decision Record 2).

**Find:**
```go
			steps := make([]string, len(ability.Steps))
			for index, step := range ability.Steps {
				steps[index] = "strike " + step.Strike
			}
			fmt.Fprintf(out, "  %s [%d]  %s\n", ability.Name, ability.Actions, strings.Join(steps, ", "))
```
**Replace with:**
```go
			steps := make([]string, len(ability.Steps))
			for index, step := range ability.Steps {
				steps[index] = stepText(step)
			}
			fmt.Fprintf(out, "  %s [%s]%s  %s\n", ability.Name, costText(ability), limitText(ability.Limit), strings.Join(steps, ", "))
```

Add after `damageText`:

```go
func stepText(step creature.Step) string {
	if step.Kind() == creature.StepStrike {
		return "strike " + step.Strike
	}
	return step.Kind()
}

func costText(ability creature.Ability) string {
	low, high := ability.CostRange()
	switch {
	case ability.Free:
		return "free"
	case high > low:
		return fmt.Sprintf("%d-%d", low, high)
	default:
		return strconv.Itoa(low)
	}
}

func limitText(limit *creature.Limit) string {
	if limit == nil {
		return ""
	}
	var parts []string
	if limit.PerFight > 0 {
		parts = append(parts, fmt.Sprintf("%d per fight", limit.PerFight))
	}
	if limit.PerRound > 0 {
		parts = append(parts, fmt.Sprintf("%d per round", limit.PerRound))
	}
	if limit.Recharge != "" {
		parts = append(parts, "recharge "+limit.Recharge)
	}
	if limit.Pool != "" {
		parts = append(parts, "pool "+limit.Pool)
	}
	return " (" + strings.Join(parts, ", ") + ")"
}
```

Add `strconv` to the imports. After the Abilities block, print the pools:

**Find:**
```go
	printFightRules(out, shown, bold, reset)
```
**Replace with:**
```go
	if len(shown.Pools) > 0 {
		entries := make([]string, 0, len(shown.Pools))
		for _, name := range slices.Sorted(maps.Keys(shown.Pools)) {
			entries = append(entries, fmt.Sprintf("%s %d", name, shown.Pools[name]))
		}
		fmt.Fprintf(out, "Pools %s\n", strings.Join(entries, ", "))
	}
	printFightRules(out, shown, bold, reset)
```

##### Task 5 — `internal/pf2e/sim/creature/limits_test.go` and `cmd/gm-toolkit/pf2e/show_test.go`

`limits_test.go` (new, `package creature_test`), each case a hand layer written with `writeLayers` and loaded with `creature.Load`:

- **`TestLoadReadsLimits`** — an ability with `limit = { per_fight = 1, recharge = "1d4", pool = "potions" }` and `[pool] potions = 3` loads; `RechargeRoll` is 1d4; `CostRange` gives 2, 2. An ability with `actions = 1, max_actions = 3` gives 1, 3. A `free = true` ability with a limit gives 0, 0.
- **`TestLoadRejectsBadCosts`** — table: `actions = 0` without `free` ("want 1 to 3"); `free` with `actions = 1`; `free` with no limit; `max_actions = 1` beside `actions = 1`; `max_actions = 4`.
- **`TestLoadRejectsBadLimits`** — table: an empty `limit = {}` ("sets nothing"); `per_fight = -1`; `recharge = "soon"`; `pool = "missing"` (trap 2: a field whose producer is absent); `[pool] potions = -1`.
- **`TestIgnoredAbilitySkipsLimitChecks`** — an ignored ability with a bad limit still loads (trap 1: the existing `ignore` early exit still holds for the new checks).

`show_test.go`: extend the existing stat-block case with an ability that has a range cost, a limit and a save step, and a pool; expect `[1-3] (1 per fight, pool potions)`, the step printed as `save`, and the `Pools` line.

##### Commit message

```
feat(E4): add action costs, limits and pools to the ability format

- An ability can be free or cost a range of actions (max_actions)
- A limit table holds per_fight, per_round, recharge and pool
- A creature's [pool] table holds counts that several abilities share
- show prints the cost, the limit and each step's kind
```

#### Commit 2 — `feat(E4): set a turn's actions from conditions and add Stand`

##### Task 1 — `internal/pf2e/sim/creature/conditions.go`

Four names join the list, and a condition step can roll its value or add to the current value.

**Find:**
```go
	ConditionNames     = []string{"blinded", "clumsy", "concealed", "dazzled", "doomed", "drained", "encumbered", "enfeebled", "frightened", "off-guard", "prone", "sickened", "stupefied"}
	valuedConditions   = []string{"clumsy", "doomed", "drained", "enfeebled", "frightened", "sickened", "stupefied"}
```
**Replace with:**
```go
	ConditionNames     = []string{"blinded", "clumsy", "concealed", "dazzled", "deafened", "doomed", "drained", "encumbered", "enfeebled", "frightened", "off-guard", "prone", "quickened", "sickened", "slowed", "stunned", "stupefied"}
	valuedConditions   = []string{"clumsy", "doomed", "drained", "enfeebled", "frightened", "sickened", "slowed", "stunned", "stupefied"}
```

**Find:**
```go
func checkCondition(path string, step *Step, report func(string, ...any)) {
	switch {
	case !IsCondition(step.Condition):
		report("%s: unknown condition %q", path, step.Condition)
	case IsValuedCondition(step.Condition) && step.Value < 1:
		report("%s: %s needs a value of at least 1", path, step.Condition)
	case !IsValuedCondition(step.Condition) && step.Value != 0:
		report("%s: %s takes no value", path, step.Condition)
	}
}
```
**Replace with:**
```go
func checkCondition(path string, step *Step, report func(string, ...any)) {
	valued := IsValuedCondition(step.Condition)
	switch {
	case !IsCondition(step.Condition):
		report("%s: unknown condition %q", path, step.Condition)
	case valued && step.Value != 0 && step.ValueDice != "":
		report("%s: set value or value_dice, not both", path)
	case valued && step.Value < 1 && step.ValueDice == "":
		report("%s: %s needs a value of at least 1", path, step.Condition)
	case !valued && (step.Value != 0 || step.ValueDice != "" || step.Stack):
		report("%s: %s takes no value", path, step.Condition)
	}
	if step.ValueDice != "" {
		step.ValueRoll = parseAmount(path+".value_dice", step.ValueDice, report)
	}
}
```

##### Task 2 — `internal/pf2e/sim/creature/steps.go`

**Find:**
```go
	Value       int             `toml:"value,omitzero"`
```
**Replace with:**
```go
	Value       int             `toml:"value,omitzero"`
	ValueDice   string          `toml:"value_dice,omitempty"`
	ValueRoll   dice.Expression `toml:"-"`
	Stack       bool            `toml:"stack,omitempty"`
```

In `checkStep`, the two new keys belong on a condition, like `value`.

**Find:**
```go
	if kind != StepCondition && step.Value != 0 {
		report("%s: value belongs on a condition", path)
	}
```
**Replace with:**
```go
	if kind != StepCondition && (step.Value != 0 || step.ValueDice != "" || step.Stack) {
		report("%s: value belongs on a condition", path)
	}
```

##### Task 3 — `internal/pf2e/sim/engine/conditions.go`

The four names get rows with no penalties; what they do to a turn is in Task 5.

**Find:**
```go
	"dazzled":    {},
	"doomed":     {},
```
**Replace with:**
```go
	"dazzled":    {},
	"deafened":   {},
	"doomed":     {},
```

**Find:**
```go
	"prone":      {selectors: []string{"attack", "spell-attack"}, kind: "circumstance", fixed: 2, implies: map[string]int{"off-guard": 0}},
	"sickened":   {selectors: allChecksAndDCs, kind: "status"},
```
**Replace with:**
```go
	"prone":      {selectors: []string{"attack", "spell-attack"}, kind: "circumstance", fixed: 2, implies: map[string]int{"off-guard": 0}},
	"quickened":  {},
	"sickened":   {selectors: allChecksAndDCs, kind: "status"},
	"slowed":     {},
	"stunned":    {},
```

In `applyCondition`, the value can be rolled, and `stack` adds it to the bearer's current value.

**Find:**
```go
	value := step.Value
	if creature.IsValuedCondition(step.Condition) {
		value += target.Creature.ConditionAdjustments[step.Condition].Gain
		if value < 1 {
			return
		}
	}
```
**Replace with:**
```go
	value := step.Value
	if step.ValueDice != "" {
		value = step.ValueRoll.Roll(fight.roller)
	}
	if creature.IsValuedCondition(step.Condition) {
		value += target.Creature.ConditionAdjustments[step.Condition].Gain
		if value < 1 {
			return
		}
		if step.Stack {
			value += target.Condition(step.Condition)
		}
	}
```

Add after `decayConditions`:

```go
func (fight *Fight) reduceCondition(bearer *Combatant, name string, amount int) {
	fight.changeConditions(bearer, func() {
		var kept []*conditionInstance
		for _, instance := range bearer.conditions {
			if instance.name == name {
				instance.value -= amount
				if instance.value < 1 {
					continue
				}
			}
			kept = append(kept, instance)
		}
		bearer.conditions = kept
	})
}
```

##### Task 4 — `internal/pf2e/sim/engine/events.go`

Add after `TurnStarted`:

```go
type ActionsRegained struct {
	Combatant string
	Actions   int
	Quickened bool
	Lost      int
}

type StoodUp struct {
	Combatant string
}
```

Add to the `isEvent` block:

```go
func (ActionsRegained) isEvent()   {}
func (StoodUp) isEvent()           {}
```

##### Task 5 — `internal/pf2e/sim/engine/actions.go`

The fix commit after Commit 2 extends this code (Decision Record 4 to 6): `regainActions` reads stunned with a duration, and `loseActionsToStunned` joins the file.

`internal/pf2e/sim/engine/actions.go` (new) — full contents:

```go
package engine

import "slices"

func (combatant *Combatant) OpenActions() int {
	return combatant.ActionsLeft - combatant.StrikeOnlyActions
}

func (fight *Fight) regainActions(actor *Combatant) {
	actions := actionsPerTurn
	quickened := actor.Condition("quickened") > 0
	if quickened {
		actions++
	}
	stunned := min(actor.Condition("stunned"), actions)
	// deliberate: actions lost to stunned count toward slowed, not on top of it, because the stunned condition says they do.
	lost := min(max(stunned, actor.Condition("slowed")), actions)
	if stunned > 0 {
		fight.reduceCondition(actor, "stunned", stunned)
	}
	actor.ActionsLeft = actions - lost
	actor.StrikeOnlyActions = 0
	// deliberate: a creature that loses an action loses the quickened one first, not a free one, because it picks which action it gives up.
	if quickened && lost == 0 {
		actor.StrikeOnlyActions = 1
	}
	if quickened || lost > 0 {
		fight.observer.Observe(ActionsRegained{Combatant: actor.ID, Actions: actor.ActionsLeft, Quickened: quickened, Lost: lost})
	}
}

func (fight *Fight) stand(actor *Combatant) {
	fight.changeConditions(actor, func() {
		actor.conditions = slices.DeleteFunc(actor.conditions, func(instance *conditionInstance) bool {
			return instance.name == "prone"
		})
	})
	fight.observer.Observe(StoodUp{Combatant: actor.ID})
}
```

##### Task 6 — `internal/pf2e/sim/engine/combatant.go`

**Find:**
```go
	ActionsLeft     int
	AttacksThisTurn int
```
**Replace with:**
```go
	ActionsLeft       int
	StrikeOnlyActions int
	AttacksThisTurn   int
```

Run `gofmt` on the file so the field column lines up.

##### Task 7 — `internal/pf2e/sim/engine/decide.go` and `run.go`

The fix commit after Commit 2 extends `act` (Decision Record 4 to 6): a dying or stable PC regains actions before its turn ends.

`decide.go`: Stand is a third kind of choice, and `FirstFitting` Stands before anything else and counts only open actions for an ability.

**Find:**
```go
const (
	ChooseStrike ChoiceKind = iota
	ChooseAbility
)
```
**Replace with:**
```go
const (
	ChooseStrike ChoiceKind = iota
	ChooseAbility
	ChooseStand
)
```

**Find:**
```go
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Actions <= actor.ActionsLeft {
			return Choice{Kind: ChooseAbility, ID: id, Target: target}, true
		}
	}
```
**Replace with:**
```go
	if actor.Condition("prone") > 0 && actor.OpenActions() > 0 {
		return Choice{Kind: ChooseStand}, true
	}
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Actions <= actor.OpenActions() {
			return Choice{Kind: ChooseAbility, ID: id, Target: target}, true
		}
	}
```

`run.go`, in `act`: the reaction resets for every living creature, the actions come from `regainActions`, and the loop stops when the actor falls.

**Find:**
```go
	fight.observer.Observe(TurnStarted{Round: fight.round, Combatant: actor.ID})
	if actor.Dying > 0 {
		fight.recoveryCheck(actor)
		return nil
	}
	if !actor.Standing() {
		return nil
	}
	actor.ActionsLeft = actionsPerTurn
	actor.AttacksThisTurn = 0
	actor.ReactionUsed = false
	for actor.ActionsLeft > 0 {
```
**Replace with:**
```go
	fight.observer.Observe(TurnStarted{Round: fight.round, Combatant: actor.ID})
	actor.ReactionUsed = false
	if actor.Dying > 0 {
		fight.recoveryCheck(actor)
		return nil
	}
	if !actor.Standing() {
		return nil
	}
	actor.AttacksThisTurn = 0
	fight.regainActions(actor)
	for actor.ActionsLeft > 0 && actor.Standing() {
```

In `perform`, Stand is handled before the target check, a Strike spends the Strike-only action first, and an ability needs open actions.

**Find:**
```go
func (fight *Fight) perform(actor *Combatant, choice Choice) error {
	if choice.Target == nil || !choice.Target.Standing() {
```
**Replace with:**
```go
func (fight *Fight) perform(actor *Combatant, choice Choice) error {
	if choice.Kind == ChooseStand {
		if actor.Condition("prone") == 0 || actor.OpenActions() < 1 {
			return fmt.Errorf("%w: %s cannot Stand", ErrBadChoice, actor.ID)
		}
		actor.ActionsLeft--
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: "stand", Actions: 1})
		fight.stand(actor)
		return nil
	}
	if choice.Target == nil || !choice.Target.Standing() {
```

**Find:**
```go
		actor.ActionsLeft--
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: choice.Target.ID})
```
**Replace with:**
```go
		actor.ActionsLeft--
		actor.StrikeOnlyActions = max(actor.StrikeOnlyActions-1, 0)
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: choice.Target.ID})
```

**Find:**
```go
		if ability.Actions > actor.ActionsLeft {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d left", ErrBadChoice, actor.ID, choice.ID, ability.Actions, actor.ActionsLeft)
		}
```
**Replace with:**
```go
		if ability.Actions > actor.OpenActions() {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d open", ErrBadChoice, actor.ID, choice.ID, ability.Actions, actor.OpenActions())
		}
```

##### Task 8 — `internal/pf2e/sim/engine/dying.go`

A PC that is knocked out falls prone (Decision 9).

**Find:**
```go
	if target != fight.current {
		fight.moveBefore(target, attacker)
	}
	fight.setDying(target, dyingOnKnockDown(target, critical))
```
**Replace with:**
```go
	if target != fight.current {
		fight.moveBefore(target, attacker)
	}
	fight.applyCondition(attacker, target, creature.Step{Condition: "prone"})
	fight.setDying(target, dyingOnKnockDown(target, critical))
```

Seam check (trap 1): `applyCondition` returns early for a dead or immune target, builds a timing with no duration, and touches drained and doomed only for those names, so none of its side effects fire here. `knockDown` is also reached from `drain` and `takePersistentDamage`; a PC downed by either is prone too, which is right.

##### Task 9 — tests

`internal/pf2e/sim/engine/actions_test.go` (new):

- **`TestRegainActions`** — table over a standing combatant given conditions in Go, then `regainActions`: none gives 3 and no event; quickened gives 4 with 1 Strike-only; slowed 1 gives 2; slowed 2 with quickened gives 2 and no Strike-only action; stunned 4 gives 0 and leaves stunned 1, and a second call gives 2 and no stunned; stunned 1 with slowed 2 gives 1, not 0 (trap 3: the overlap rule, which no Fall of Luna sheet hits); stunned 2 with slowed 1 gives 1.
- **`TestStrikeOnlyAction`** — a quickened actor with a 3-action ability and a Strike: through `perform`, the ability leaves 1 action, a 1-action ability is then refused with `ErrBadChoice`, and a Strike is accepted and leaves 0.
- **`TestKnockedOutPCIsProne`** — a PC dropped by a Strike has prone and off-guard; healed above 0 it is still prone; under `FirstFitting` its next turn opens with `ActionUsed{Action: "stand"}` and `StoodUp`, then two more actions.
- **`TestStandNeedsProne`** — `ChooseStand` from a creature that is not prone fails with `ErrBadChoice`; so does a prone creature whose only action is the Strike-only one.
- **`TestSlowedForOneRound`** — a source applies slowed 1 with `rounds = 1` on its turn; in a two-round `Run` the target has 2 actions on its next turn and 3 on the one after.
- **`TestConditionValueDice`** — a `drained` step with `value_dice = "1d4"` at two targets rolls once per target; an immune target takes no die from the scripted roller (trap 4). A `doomed` step with `stack = true` takes doomed 1 to 2, and on a dying PC already one short of its death value it kills (trap 5: the E3 follow-on check still runs through the new path).

`internal/pf2e/sim/creature/steps_test.go`: add to the bad-conditions table `value` and `value_dice` together, `stack` on `off-guard`, and `value_dice` on a heal step.

Existing tests: see Shared Context, Whole-list event tests. A test that runs a prone creature under `FirstFitting` now sees it Stand first; give its scripted roller the changed rolls.

##### Commit message

```
feat(E4): set a turn's actions from conditions and add Stand

- Slowed, stunned and quickened change the actions a turn starts with
- Quickened's extra action is for a Strike only
- A knocked-out PC is prone, and Stand costs 1 action
- A condition step can roll its value (value_dice) or add to it (stack)
- Deafened joins the condition names with no effect yet
```

### Phase 2 — Limits and Strike tweaks

**In plain words:** the engine now spends and enforces every limit, and a Strike inside an ability can carry extra dice, extra damage and its own attack count. When this lands, Vicious Swing, Brutal Finish, Destiny Spores, Wail, Walk the Timelines, the relic bombs and the quick bombs all run from data.

#### Commit 3 — `feat(E4): enforce limits and add Strike tweaks`

##### Task 1 — `internal/pf2e/sim/creature/steps.go`

Two new step kinds (`gain_actions`, `recharge`), the Strike tweaks, a second target word and a save whose success takes full damage.

**Find:**
```go
	StepWeaken      = "weaken"
)
```
**Replace with:**
```go
	StepWeaken      = "weaken"
	StepGainActions = "gain_actions"
	StepRecharge    = "recharge"

	PenaltyShared = "shared"
	TargetSelf    = "self"
	TargetOther   = "other"
	SuccessFull   = "full"
)
```

**Find:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken}
```
**Replace with:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken, StepGainActions, StepRecharge}
	stepTargets = []string{"", TargetSelf, TargetOther}
```

**Find:**
```go
	Weaken      *DefenseChange  `toml:"weaken,omitempty"`
	Outcomes
}
```
**Replace with:**
```go
	Weaken        *DefenseChange `toml:"weaken,omitempty"`
	GainActions   int            `toml:"gain_actions,omitzero"`
	Recharge      string         `toml:"recharge,omitempty"`
	ExtraDice     int            `toml:"extra_dice,omitzero"`
	ExtraDamage   []DamagePart   `toml:"extra_damage,omitempty"`
	Counts        int            `toml:"counts,omitzero"`
	Penalty       string         `toml:"penalty,omitempty"`
	SuccessDamage string         `toml:"success_damage,omitempty"`
	Outcomes
}
```

In `heads`, before `return heads`:

```go
	if step.GainActions != 0 {
		heads = append(heads, StepGainActions)
	}
	if step.Recharge != "" {
		heads = append(heads, StepRecharge)
	}
```

In `checkStep`, add two cases to the `switch kind`:

```go
	case StepGainActions:
		if step.GainActions < 1 || step.GainActions > 3 {
			report("%s: gain_actions %d: want 1 to 3", path, step.GainActions)
		}
	case StepRecharge:
		if target, found := creature.Abilities[step.Recharge]; !found || target.Limit == nil || target.Limit.Recharge == "" {
			report("%s: recharge %q: want an ability with limit.recharge", path, step.Recharge)
		}
```

**Find:**
```go
	if step.Target != "" && step.Target != "self" {
		report("%s: target %q: want self, or leave it unset", path, step.Target)
	}
	prepareDamage(path, step.Damage, report)
```
**Replace with:**
```go
	if !slices.Contains(stepTargets, step.Target) {
		report("%s: target %q: want self, other, or leave it unset", path, step.Target)
	}
	if kind != StepStrike && (step.ExtraDice != 0 || len(step.ExtraDamage) > 0 || step.Counts != 0 || step.Penalty != "") {
		report("%s: extra_dice, extra_damage, counts and penalty belong on a strike", path)
	}
	if step.ExtraDice < 0 || step.Counts < 0 || step.Counts > 3 {
		report("%s: extra_dice must not be negative, and counts is 0 to 3", path)
	}
	if step.Penalty != "" && step.Penalty != PenaltyShared {
		report("%s: penalty %q: want shared, or leave it unset", path, step.Penalty)
	}
	if step.SuccessDamage != "" && (step.SuccessDamage != SuccessFull || !step.Basic) {
		report("%s: success_damage: want full, on a basic save", path)
	}
	prepareDamage(path, step.Damage, report)
	prepareDamage(path+".extra", step.ExtraDamage, report)
```

Add at the end of the file:

```go
func (step Step) clone() Step {
	step.Damage = slices.Clone(step.Damage)
	step.ExtraDamage = slices.Clone(step.ExtraDamage)
	step.CriticalSuccess = cloneSteps(step.CriticalSuccess)
	step.Success = cloneSteps(step.Success)
	step.Failure = cloneSteps(step.Failure)
	step.CriticalFailure = cloneSteps(step.CriticalFailure)
	return step
}

func cloneSteps(steps []Step) []Step {
	if steps == nil {
		return nil
	}
	cloned := make([]Step, len(steps))
	for index, step := range steps {
		cloned[index] = step.clone()
	}
	return cloned
}
```

##### Task 2 — `internal/pf2e/sim/creature/creature.go` and `build.go`

`creature.go`: a damage part can scale with the actions spent, and an ability can end the turn.

**Find:**
```go
	Category string          `toml:"category,omitempty"`
	Roll     dice.Expression `toml:"-"`
}
```
**Replace with:**
```go
	Category  string          `toml:"category,omitempty"`
	PerAction bool            `toml:"per_action,omitempty"`
	Roll      dice.Expression `toml:"-"`
}
```

**Find:**
```go
	Free       bool     `toml:"free,omitempty"`
```
**Replace with:**
```go
	Free       bool     `toml:"free,omitempty"`
	TurnEnds   bool     `toml:"turn_ends,omitempty"`
```

`build.go`, in `riders`: each weapon gets its own copy of the table row, so validation never writes parsed rolls into the shared table (E3 Decision Record 9).

**Find:**
```go
		outcomes.CriticalSuccess = append(slices.Clone(outcomes.For(3, true)), specialization)
```
**Replace with:**
```go
		outcomes.CriticalSuccess = append(slices.Clone(outcomes.For(3, true)), specialization.clone())
```

##### Task 3 — `internal/pf2e/sim/engine/combatant.go`

The combatant holds what it has spent. `NewCombatant` makes the maps, so every test combatant has them.

**Find:**
```go
	conditions      []*conditionInstance
	persistent      []*persistentInstance
	defenseChanges  []*defenseChange
	adaptiveType    string
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	return &Combatant{ID: source.ID, Side: side, Creature: source, HP: source.Defense.HP, HeroPoints: source.HeroPoints}
}
```
**Replace with:**
```go
	conditions      []*conditionInstance
	persistent      []*persistentInstance
	defenseChanges  []*defenseChange
	adaptiveType    string
	uses            map[string]int
	roundUses       map[string]int
	recharging      map[string]int
	pools           map[string]int
	flourished      bool
	lastPenalty     int
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	pools := map[string]int{}
	maps.Copy(pools, source.Pools)
	return &Combatant{
		ID:         source.ID,
		Side:       side,
		Creature:   source,
		HP:         source.Defense.HP,
		HeroPoints: source.HeroPoints,
		uses:       map[string]int{},
		roundUses:  map[string]int{},
		recharging: map[string]int{},
		pools:      pools,
	}
}
```

Add `maps` to the imports.

##### Task 4 — `internal/pf2e/sim/engine/limits.go`

`internal/pf2e/sim/engine/limits.go` (new) — full contents:

```go
package engine

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func (combatant *Combatant) UsesLeft(ability creature.Ability) int {
	if ability.Limit == nil || ability.Limit.PerFight == 0 {
		return -1
	}
	return ability.Limit.PerFight - combatant.uses[ability.ID]
}

func (combatant *Combatant) RechargeLeft(abilityID string) int {
	return combatant.recharging[abilityID]
}

func (combatant *Combatant) Pool(name string) int {
	return combatant.pools[name]
}

func (combatant *Combatant) Blocked(ability creature.Ability) string {
	if limit := ability.Limit; limit != nil {
		switch {
		case limit.PerFight > 0 && combatant.uses[ability.ID] >= limit.PerFight:
			return "no uses left"
		case limit.PerRound > 0 && combatant.roundUses[ability.ID] >= limit.PerRound:
			return "used this round"
		case combatant.recharging[ability.ID] > 0:
			return "recharging"
		case limit.Pool != "" && combatant.pools[limit.Pool] < 1:
			return "pool " + limit.Pool + " is empty"
		}
	}
	switch {
	case ability.HasTrait("flourish") && combatant.flourished:
		return "flourish already used this turn"
	case ability.HasTrait("press") && combatant.AttacksThisTurn == 0:
		return "press needs an earlier attack this turn"
	}
	return ""
}

func (fight *Fight) spend(actor *Combatant, ability creature.Ability) {
	if ability.HasTrait("flourish") {
		actor.flourished = true
	}
	limit := ability.Limit
	if limit == nil {
		return
	}
	if limit.PerFight > 0 {
		actor.uses[ability.ID]++
		fight.observer.Observe(UseSpent{Combatant: actor.ID, Ability: ability.ID, Left: limit.PerFight - actor.uses[ability.ID]})
	}
	if limit.PerRound > 0 {
		actor.roundUses[ability.ID]++
	}
	if limit.Pool != "" {
		actor.pools[limit.Pool]--
		fight.observer.Observe(PoolSpent{Combatant: actor.ID, Pool: limit.Pool, Ability: ability.ID, Left: actor.pools[limit.Pool]})
	}
	if limit.Recharge != "" {
		rounds := max(limit.RechargeRoll.Roll(fight.roller), 1)
		actor.recharging[ability.ID] = rounds
		fight.observer.Observe(RechargeRolled{Combatant: actor.ID, Ability: ability.ID, Rounds: rounds})
	}
}

// deliberate: a recharge counts down at the start of the user's turn, not at the end, because a roll of 1 must mean ready next turn (Decision 3).
func (fight *Fight) tickLimits(actor *Combatant) {
	clear(actor.roundUses)
	actor.flourished = false
	for _, id := range slices.Sorted(maps.Keys(actor.recharging)) {
		actor.recharging[id]--
		if actor.recharging[id] < 1 {
			delete(actor.recharging, id)
			fight.observer.Observe(Recharged{Combatant: actor.ID, Ability: id})
		}
	}
}

func (fight *Fight) recharge(actor *Combatant, abilityID, by string) {
	if actor.recharging[abilityID] < 1 {
		return
	}
	delete(actor.recharging, abilityID)
	fight.observer.Observe(Recharged{Combatant: actor.ID, Ability: abilityID, By: by})
}
```

##### Task 5 — `internal/pf2e/sim/engine/events.go`

Add after `ActionUsed`:

```go
type UseSpent struct {
	Combatant string
	Ability   string
	Left      int
}

type PoolSpent struct {
	Combatant string
	Pool      string
	Ability   string
	Left      int
}

type RechargeRolled struct {
	Combatant string
	Ability   string
	Rounds    int
}

type Recharged struct {
	Combatant string
	Ability   string
	By        string
}

type ActionsGained struct {
	Combatant string
	Ability   string
	Amount    int
}
```

Add to the `isEvent` block:

```go
func (UseSpent) isEvent()          {}
func (PoolSpent) isEvent()         {}
func (RechargeRolled) isEvent()    {}
func (Recharged) isEvent()         {}
func (ActionsGained) isEvent()     {}
```

##### Task 6 — `internal/pf2e/sim/engine/strike.go`

`strike` takes its tweaks as one options value, reports its degree, applies the penalty on the attacker's own turn only, and runs the step's own degree lists after the Strike's riders (Decisions 24 and 27). Replace everything from `func (fight *Fight) strike` to the end of the file with:

```go
type strikeOptions struct {
	ability     string
	extraDice   int
	extraDamage []creature.DamagePart
	counts      int
	penalty     string
	lists       creature.Outcomes
}

func (fight *Fight) strikePenalty(attacker *Combatant, strike creature.Strike, options strikeOptions) int {
	// deliberate: the multiple attack penalty applies on the attacker's own turn only, not to a reaction Strike, because the rules count attacks per turn (Decision 24).
	if attacker != fight.current {
		return 0
	}
	if options.penalty == creature.PenaltyShared {
		return attacker.lastPenalty
	}
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn += max(options.counts, 1)
	attacker.lastPenalty = penalty
	return penalty
}

func strikeDamage(parts []creature.DamagePart, options strikeOptions) []creature.DamagePart {
	if options.extraDice == 0 && len(options.extraDamage) == 0 {
		return parts
	}
	combined := slices.Clone(parts)
	for index := range combined {
		// deliberate: extra dice go on the first rolled part with no category, not on a flat number, because "an extra weapon damage die" needs a die size.
		if combined[index].Category == "" && combined[index].Roll.Sides > 0 {
			combined[index].Roll.Count += options.extraDice
			break
		}
	}
	return append(combined, options.extraDamage...)
}

func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, options strikeOptions) (Degree, bool) {
	penalty := fight.strikePenalty(attacker, strike, options)
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
	if strike.Melee() {
		attackSelectors, damageSelectors = []string{"attack", "melee-attack"}, []string{"damage", "melee-damage"}
	}
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, strike.Traits)
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(StrikeRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Die:      die,
		Bonus:    bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits}
	critical := degree == CriticalSuccess
	parts := strikeDamage(strike.Damage, options)
	switch {
	case degree >= Success:
		damageBonus := fight.modifierPool(attacker, damageSelectors, nil, target).Total()
		fight.applyDamage(source, target, fight.rollDamage(parts, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with A.004.2 (E3 Decision 11).
		if splash := splashParts(parts); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
	// deliberate: the lists run even when the target died, not only on a living target, because a list can hold a step for the attacker (a recharge); each step kind skips a dead target itself.
	fight.runSteps(source, strike.For(int(degree), true), target)
	fight.runSteps(source, options.lists.For(int(degree), true), target)
	return degree, true
}
```

Add `slices` to the imports. Seam check (trap 1) on dropping the old `if !target.Dead` gate: `applyCondition`, `heal`, `gainTempHP` and `weaken` return for a dead target; the strike and spell attack cases need a standing one; `saveStep` and the damage case skip a dead one. So a list on a dead target runs only its steps for the attacker.

##### Task 7 — `internal/pf2e/sim/engine/steps.go`

`useAbility` takes one value, knows the actions spent, stops when its user falls, and survives an ability with no steps. Steps for the user (`gain_actions`, `recharge`) run before any target is worked out.

**Find:**
```go
type effectSource struct {
	actor   *Combatant
	ability string
	strike  string
	traits  []string
	lethal  bool
}

func (fight *Fight) useAbility(actor *Combatant, ability creature.Ability, target *Combatant) {
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits}
	if ability.Area != nil {
		targets := fight.areaTargets(actor)
		for _, step := range ability.Steps {
			fight.runStep(source, step, targets)
		}
		return
	}
	if first := ability.Steps[0]; first.Kind() == creature.StepSave && first.Target != "self" && !fight.passesConcealment(actor, target) {
		return
	}
	for _, step := range ability.Steps {
		if step.Target != "self" && !target.Standing() {
			next, found := fight.decider.NextTarget(fight, actor)
			if !found {
				return
			}
			target = next
		}
		fight.runStep(source, step, []*Combatant{target})
	}
}
```
**Replace with:**
```go
type effectSource struct {
	actor   *Combatant
	ability string
	strike  string
	traits  []string
	lethal  bool
	actions int
}

type abilityUse struct {
	actor   *Combatant
	ability creature.Ability
	target  *Combatant
	actions int
}

func (fight *Fight) useAbility(use abilityUse) {
	actor, ability, target := use.actor, use.ability, use.target
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions}
	if ability.Area != nil {
		targets := fight.areaTargets(actor)
		for _, step := range ability.Steps {
			if !actor.Standing() {
				return
			}
			fight.runStep(source, step, targets)
		}
		return
	}
	if len(ability.Steps) > 0 && ability.Steps[0].Kind() == creature.StepSave && ability.Steps[0].Target != creature.TargetSelf && !fight.passesConcealment(actor, target) {
		return
	}
	for _, step := range ability.Steps {
		// deliberate: the ability stops when its user falls, not at its last step, because a reaction can drop the user partway through (Decision 28).
		if !actor.Standing() {
			return
		}
		if step.Target != creature.TargetSelf && !target.Standing() {
			next, found := fight.decider.NextTarget(fight, actor)
			if !found {
				return
			}
			target = next
		}
		stepTarget := target
		if step.Target == creature.TargetOther {
			stepTarget = fight.otherFoe(actor, target)
		}
		fight.runStep(source, step, []*Combatant{stepTarget})
	}
}

// deliberate: "other" is the first standing foe that is not the ability's target, not a decider pick, because Walk the Timelines is a stand-in (Decision 13); with one foe standing it is that foe.
func (fight *Fight) otherFoe(actor, target *Combatant) *Combatant {
	for _, foe := range fight.StandingFoes(actor) {
		if foe != target {
			return foe
		}
	}
	return target
}
```

In `runStep`, the user's own steps come first.

**Find:**
```go
func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	source.lethal = step.Lethal
	if step.Target == "self" {
```
**Replace with:**
```go
func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	switch step.Kind() {
	case creature.StepGainActions:
		source.actor.ActionsLeft += step.GainActions
		fight.observer.Observe(ActionsGained{Combatant: source.actor.ID, Ability: source.ability, Amount: step.GainActions})
		return
	case creature.StepRecharge:
		fight.recharge(source.actor, step.Recharge, source.ability)
		return
	}
	source.lethal = step.Lethal
	if step.Target == creature.TargetSelf {
```

**Find:**
```go
			if target.Standing() {
				fight.strike(source.actor, target, strike, source.ability)
			}
```
**Replace with:**
```go
			if target.Standing() {
				fight.strike(source.actor, target, strike, strikeOptions{
					ability:     source.ability,
					extraDice:   step.ExtraDice,
					extraDamage: step.ExtraDamage,
					counts:      step.Counts,
					penalty:     step.Penalty,
					lists:       step.Outcomes,
				})
			}
```

**Find:**
```go
	case creature.StepDamage:
		rolled := fight.rollDamage(step.Damage, false, 0)
```
**Replace with:**
```go
	case creature.StepDamage:
		rolled := fight.rollDamage(scaleDamage(step.Damage, source.actions), false, 0)
```

In `saveStep`, the damage scales with the actions, and a save marked `success_damage = "full"` does not halve.

**Find:**
```go
	if step.Basic {
		rolled = fight.rollDamage(step.Damage, false, 0)
	}
```
**Replace with:**
```go
	if step.Basic {
		rolled = fight.rollDamage(scaleDamage(step.Damage, source.actions), false, 0)
	}
```

**Find:**
```go
		if step.Basic && degree < CriticalSuccess {
			fight.applyDamage(source, target, rolled.basic(degree), degree == CriticalFailure)
		}
```
**Replace with:**
```go
		if step.Basic && degree < CriticalSuccess {
			taken := rolled.basic(degree)
			if degree == Success && step.SuccessDamage == creature.SuccessFull {
				taken = rolled
			}
			fight.applyDamage(source, target, taken, degree == CriticalFailure)
		}
```

In `spellAttack`:

**Find:**
```go
	penalty := attackPenalty(attacker.AttacksThisTurn, false)
	attacker.AttacksThisTurn++
```
**Replace with:**
```go
	penalty := 0
	if attacker == fight.current {
		penalty = attackPenalty(attacker.AttacksThisTurn, false)
		attacker.AttacksThisTurn++
	}
```

**Find:**
```go
		fight.applyDamage(source, target, fight.rollDamage(step.Damage, critical, 0), critical)
```
**Replace with:**
```go
		fight.applyDamage(source, target, fight.rollDamage(scaleDamage(step.Damage, source.actions), critical, 0), critical)
```

Add at the end of the file:

```go
func scaleDamage(parts []creature.DamagePart, actions int) []creature.DamagePart {
	if actions < 2 || !slices.ContainsFunc(parts, func(part creature.DamagePart) bool { return part.PerAction }) {
		return parts
	}
	scaled := slices.Clone(parts)
	for index := range scaled {
		if scaled[index].PerAction {
			scaled[index].Roll.Count *= actions
			scaled[index].Roll.Modifier *= actions
		}
	}
	return scaled
}
```

##### Task 8 — `internal/pf2e/sim/engine/decide.go`

A choice can state the actions it spends, and `FirstFitting` skips what is blocked.

**Find:**
```go
type Choice struct {
	Kind   ChoiceKind
	ID     string
	Target *Combatant
}
```
**Replace with:**
```go
type Choice struct {
	Kind    ChoiceKind
	ID      string
	Target  *Combatant
	Actions int
}
```

**Find:**
```go
		if !ability.Ignore && ability.Actions <= actor.OpenActions() {
```
**Replace with:**
```go
		if !ability.Ignore && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
```

##### Task 9 — `internal/pf2e/sim/engine/run.go`

In `act`, limits tick for every living creature, a dying one included.

**Find:**
```go
	actor.ReactionUsed = false
	if actor.Dying > 0 {
```
**Replace with:**
```go
	actor.ReactionUsed = false
	fight.tickLimits(actor)
	if actor.Dying > 0 {
```

Where `act` sets `actor.AttacksThisTurn = 0`, it also sets `actor.lastPenalty = 0` (Decision Record 7).

In `perform`:

**Find:**
```go
		fight.strike(actor, choice.Target, strike, "")
		return nil
```
**Replace with:**
```go
		fight.strike(actor, choice.Target, strike, strikeOptions{})
		return nil
```

**Find:**
```go
		if ability.Actions > actor.OpenActions() {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d open", ErrBadChoice, actor.ID, choice.ID, ability.Actions, actor.OpenActions())
		}
		actor.ActionsLeft -= ability.Actions
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: ability.Actions, Target: choice.Target.ID})
		fight.useAbility(actor, ability, choice.Target)
		return nil
```
**Replace with:**
```go
		cost := ability.Actions
		if choice.Actions != 0 {
			cost = choice.Actions
		}
		low, high := ability.CostRange()
		if cost < low || cost > high {
			return fmt.Errorf("%w: %s chose %s for %d actions; it takes %d to %d", ErrBadChoice, actor.ID, choice.ID, cost, low, high)
		}
		if cost > actor.OpenActions() {
			return fmt.Errorf("%w: %s chose %s (%d actions) with %d open", ErrBadChoice, actor.ID, choice.ID, cost, actor.OpenActions())
		}
		if reason := actor.Blocked(ability); reason != "" {
			return fmt.Errorf("%w: %s chose %s: %s", ErrBadChoice, actor.ID, choice.ID, reason)
		}
		actor.ActionsLeft -= cost
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: cost, Target: choice.Target.ID})
		fight.spend(actor, ability)
		fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
		if ability.TurnEnds {
			actor.ActionsLeft = 0
		}
		return nil
```

##### Task 10 — `internal/pf2e/sim/engine/testdata/scope/`

Five new folders, each one `hand.toml`: a small typed stat block and the E4 abilities this commit completes. The Strike IDs are the real ones, so Commit 13 can copy the ability tables as they are. Later commits add abilities to these files.

`testdata/scope/abyssal-luna/hand.toml`:

```toml
schema = 1

name       = "Abyssal Luna"
kind       = "npc"
level      = 16
perception = 29

[defense]
hp        = 300
ac        = 36
fortitude = 26
reflex    = 30
will      = 31
immune    = ["confused", "doomed", "paralyzed", "sleep"]

[spellcasting]
attack = 31
dc     = 39

# Impending Fate and Fey Momentum: a critical Strike gives doomed and recharges Destiny Spores.
[strike.slash]
name   = "Slash"
bonus  = 30
traits = ["agile", "magical", "reach-10", "unarmed"]
damage = [{ dice = "3d6+14", type = "slashing" }, { dice = "1d8", type = "mental" }]
critical_success = [
  { condition = "doomed", value = 1, stack = true },
  { recharge = "destiny-spores" },
]

[strike.stab]
name   = "Stab"
bonus  = 28
traits = ["magical", "reach-10", "unarmed"]
damage = [{ dice = "3d8+14", type = "piercing" }, { dice = "1d8", type = "mental" }]
critical_success = [
  { condition = "doomed", value = 1, stack = true },
  { recharge = "destiny-spores" },
]

[ability.destiny-spores]
name    = "Destiny Spores"
actions = 2
traits  = ["mental", "occult"]
area    = { shape = "cone", size = 40 }
limit   = { recharge = "1d4" }
steps   = [
  { save = "will", dc = 39, basic = true, damage = [{ dice = "15d6", type = "mental" }],
    failure = [{ condition = "slowed", value = 1, duration = { rounds = 1 } }],
    critical_failure = [
      { condition = "slowed", value = 2, duration = { rounds = 1 } },
      { condition = "doomed", value = 1, stack = true },
    ] },
]

# approximation: two Strikes at one penalty; the move out of reach waits for A.004.2.
[ability.walk-the-timelines]
name    = "Walk the Timelines"
actions = 2
traits  = ["occult"]
limit   = { per_fight = 1 }
steps   = [{ strike = "slash" }, { strike = "slash", penalty = "shared", target = "other" }]
```

`testdata/scope/avatar-of-lilith/hand.toml`:

```toml
schema = 1

name       = "Avatar of Lilith"
kind       = "npc"
level      = 17
perception = 32

[defense]
hp        = 390
ac        = 39
fortitude = 25
reflex    = 29
will      = 32
immune    = ["bleed", "death-effects", "disease", "paralyzed", "poison", "precision", "unconscious"]

[defense.resist.all]
value  = 12
except = ["holy", "cold-iron"]

[defense.weakness.holy]
value = 15

[spellcasting]
attack = 32
dc     = 40

[strike.slash]
name   = "Slash"
bonus  = 32
traits = ["agile", "magical", "reach-10", "unholy"]
damage = [{ dice = "3d10+15", type = "slashing" }]

# approximation: the burst only; the 3-action version and the lingering round are dropped.
[ability.wail]
name    = "Wail"
actions = 2
traits  = ["auditory", "concentrate", "occult", "unholy"]
area    = { shape = "emanation", size = 40 }
limit   = { recharge = "1d4" }
steps   = [
  { save = "fortitude", dc = 38, basic = true, success_damage = "full", damage = [{ dice = "8d10", type = "spirit" }],
    failure = [{ condition = "drained", value_dice = "1d4" }],
    critical_failure = [{ condition = "drained", value = 4 }] },
]
```

`testdata/scope/kas/hand.toml`:

```toml
schema = 1

name        = "Kas Fringebramble"
kind        = "pc"
level       = 14
perception  = 24
class_dc    = 34
hero_points = 1
dying_max   = 6
recovery_dc = 6

[defense]
hp        = 232
ac        = 36
fortitude = 28
reflex    = 23
will      = 24

[strike.protectors-fury]
name   = "Protector's Fury"
bonus  = 30
traits = ["magical", "shove"]
damage = [{ dice = "3d12+11", type = "bludgeoning" }]

# Furious Focus: with a two-handed weapon the swing counts as one attack, so no `counts = 2`.
[ability.vicious-swing]
name    = "Vicious Swing"
actions = 2
traits  = ["fighter", "flourish"]
steps   = [{ strike = "protectors-fury", extra_dice = 2 }]

[ability.brutal-finish]
name      = "Brutal Finish"
actions   = 1
traits    = ["fighter", "press"]
turn_ends = true
steps     = [
  { strike = "protectors-fury", extra_dice = 1,
    failure = [{ damage = [{ dice = "1d12", type = "bludgeoning" }] }] },
]

# approximation: the push is dropped; position is A.004.2's.
[ability.ring-of-the-ram]
name        = "Ring of the Ram (Greater)"
actions     = 1
max_actions = 3
traits      = ["force", "magical", "manipulate"]
limit       = { per_fight = 1 }
steps       = [
  { save = "fortitude", dc = 32, basic = true, damage = [{ dice = "3d6", type = "force", per_action = true }] },
]
```

`testdata/scope/tavish/hand.toml`:

```toml
schema = 1

name        = "Tavish Tyne"
kind        = "pc"
level       = 14
perception  = 22
class_dc    = 35
hero_points = 1

[defense]
hp        = 172
ac        = 34
fortitude = 26
reflex    = 26
will      = 22

# Quick Alchemy stand-in: a count of quick bombs, a dial for A.004.2.
[pool]
quick-bombs = 6

[strike.alchemists-fire-greater]
name   = "Alchemist's Fire (Greater)"
bonus  = 27
range  = 20
traits = ["alchemical", "bomb", "consumable", "fire", "splash"]
damage = [
  { dice = "3d8", type = "fire" },
  { dice = "3", type = "fire", category = "persistent" },
  { dice = "8", type = "fire", category = "splash" },
]

[ability.relic-flare-bolt]
name    = "Relic: Flare Bolt"
actions = 2
traits  = ["fire"]
steps   = [
  { strike = "alchemists-fire-greater", extra_damage = [{ dice = "7d6", type = "fire" }],
    critical_success = [{ damage = [{ dice = "14", type = "fire", category = "persistent" }] }] },
]

[ability.relic-tidal-crash]
name    = "Relic: Tidal Crash"
actions = 2
traits  = ["water"]
limit   = { per_fight = 1 }
steps   = [
  { strike = "alchemists-fire-greater",
    extra_damage = [{ dice = "8d10", type = "bludgeoning" }, { dice = "8", type = "bludgeoning", category = "splash" }] },
]

# approximation: the cone is a save by the struck foe; each phase has one boss.
[ability.relic-searing-wave]
name    = "Relic: Searing Wave"
actions = 2
traits  = ["fire"]
limit   = { recharge = "1d4" }
steps   = [
  { strike = "alchemists-fire-greater",
    success = [{ save = "reflex", dc_from = "class", basic = true, damage = [{ dice = "7d10", type = "fire" }] }],
    failure = [{ save = "reflex", dc_from = "class", basic = true, damage = [{ dice = "7d10", type = "fire" }] }] },
]

[ability.quick-bomb]
name    = "Quick Alchemy: bomb"
actions = 2
traits  = ["alchemist", "manipulate"]
limit   = { pool = "quick-bombs" }
steps   = [{ strike = "alchemists-fire-greater" }]

[ability.quick-bomb-sticky]
name    = "Quick Alchemy: Sticky Bomb"
actions = 2
traits  = ["additive", "alchemist", "manipulate"]
limit   = { pool = "quick-bombs", per_round = 1 }
steps   = [
  { strike = "alchemists-fire-greater", extra_damage = [{ dice = "8", type = "fire", category = "persistent" }] },
]

[ability.quick-bomb-debilitating]
name    = "Quick Alchemy: Debilitating Bomb"
actions = 2
traits  = ["additive", "alchemist", "manipulate"]
limit   = { pool = "quick-bombs", per_round = 1 }
steps   = [
  { strike = "alchemists-fire-greater",
    success = [{ save = "fortitude", dc_from = "class",
      failure = [{ condition = "dazzled", duration = { until = "source-turn-start" } }],
      critical_failure = [{ condition = "dazzled", duration = { until = "source-turn-start" } }] }] },
]
```

`testdata/scope/arabesque/hand.toml`:

```toml
schema = 1

name        = "Arabesque Opalheim"
kind        = "pc"
level       = 14
perception  = 20
hero_points = 1

[defense]
hp        = 150
ac        = 33
fortitude = 22
reflex    = 24
will      = 26

[spellcasting]
attack = 26
dc     = 36

[strike.blink-blade]
name   = "Blink Blade"
bonus  = 24
traits = ["agile", "finesse", "magical", "versatile-s"]
damage = [{ dice = "2d4+5", type = "piercing" }]
```

##### Task 11 — tests

`helpers_test.go`: add a scripted decider, and a loader for the scope folders.

```go
type scriptedDecider struct {
	choices []Choice
}

func (decider *scriptedDecider) Decide(fight *Fight, actor *Combatant) (Choice, bool) {
	if len(decider.choices) == 0 {
		return Choice{}, false
	}
	choice := decider.choices[0]
	decider.choices = decider.choices[1:]
	return choice, true
}

func (decider *scriptedDecider) NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool) {
	return FirstFitting{}.NextTarget(fight, actor)
}

func scopeCombatant(t *testing.T, id string, side Side) *Combatant {
	t.Helper()
	loaded, err := creature.Load(filepath.Join("testdata", "scope", id))
	if err != nil {
		t.Fatalf("load scope creature %s: %v", id, err)
	}
	return NewCombatant(side, loaded)
}
```

Add `path/filepath` to its imports.

Call sites: every `fight.strike(a, b, s, "")` in the test files becomes `fight.strike(a, b, s, strikeOptions{})`, and every `fight.useAbility(a, ability, b)` becomes `fight.useAbility(abilityUse{actor: a, ability: ability, target: b})`. A test that calls `strike` or a spell attack directly and expects a multiple attack penalty sets `fight.current` to the attacker first; `rules_test.go` has the penalty cases.

`internal/pf2e/sim/engine/limits_test.go` (new):

- **`TestPerFightLimit`** — an ability with `per_fight = 1`: the first `perform` emits `UseSpent{Left: 0}`; the second fails with `ErrBadChoice` and "no uses left"; `UsesLeft` goes 1, 0; an unlimited ability reports −1.
- **`TestPoolIsShared`** — two abilities on one pool of 2: one use of each empties it, both are then blocked, and `PoolSpent` carries 1 then 0.
- **`TestPerRoundLimit`** — blocked for the rest of the turn, open again after `tickLimits`.
- **`TestRecharge`** — the roller gives 1: `RechargeRolled{Rounds: 1}`, blocked this turn, `Recharged` at the start of the user's next turn. The roller gives 4: blocked on the next three turns. A `recharge` step for a ready ability emits nothing; for a recharging one it emits `Recharged{By: ...}`.
- **`TestLimitsTickWhileDying`** (trap 5) — a PC rolls a recharge of 2, drops, and is dying on its next two turns: the ability is ready when it is revived.
- **`TestFlourishAndPress`** — a second flourish ability in one turn is blocked, and open next turn; a press ability is blocked before any attack and open after one.
- **`TestFirstFittingSkipsBlocked`** — with its only ability used up, `FirstFitting` falls back to the Strike.
- **`TestVariableCost`** — a 1-to-3 ability with `per_action` damage: `Choice.Actions = 3` rolls three times the dice and spends 3; `Actions = 4` fails, and `Actions = 1` on a 2-action minimum fails; an unset `Actions` spends the minimum (Decision Record 8).
- **`TestStrikeTweaks`** — table: `extra_dice = 2` rolls 5d12 on a 3d12 Strike, doubled on a critical hit; `extra_damage` with a splash part still lands on a miss; `counts = 2` puts the next Strike at −10; `penalty = "shared"` gives the second Strike the first one's penalty and leaves the count at 1; a Strike made while `fight.current` is another creature takes no penalty and leaves `AttacksThisTurn` alone (trap 1: the old code would have counted it).
- **`TestStrikeStepListsRun`** — a strike step with a `failure` damage list deals it on a miss and not on a hit (Decision 27); a `critical_success` recharge step on a Strike rider runs even when the hit killed the target (trap 5).
- **`TestExtraDiceSkipsFlatParts`** (trap 3) — a Strike whose first part is a flat `5` gets no dice from `extra_dice`.
- **`TestTurnEnds`** — after a `turn_ends` ability the actor has 0 actions.
- **`TestOtherFoe`** — `target = "other"` picks the second standing foe, and the same foe when it stands alone.
- **`TestCriticalSpecializationRowIsCopied`** (in `creature/build_test.go`, Decision Record 9) — two weapons of the hammer group get riders that do not share a slice (E3 Decision Record 9).

`internal/pf2e/sim/engine/scope_test.go` (new) — `TestScopeRows`, one subtest per rules-scope row, each with scope combatants, a scripted roller and a scripted decider:

- **Vicious Swing, with Furious Focus** — 5d12+11 on a hit; the next Strike is at −5, not −10; a second Vicious Swing in the turn is refused (flourish).
- **Brutal Finish** — refused as the first action; after a Strike it deals 4d12+11 on a hit, 1d12 on a miss, nothing on a critical miss, and the turn ends.
- **Destiny Spores, Impending Fate, Fey Momentum** — three PCs save; a failure is slowed 1 and a critical failure slowed 2 and doomed 1; the recharge is rolled; a critical Slash recharges it and gives doomed; Luna (immune to doomed) is not involved. A PC already doomed 1 goes to 2.
- **Walk the Timelines** — two Slashes at the same penalty at two PCs, then the next Strike at −4 (agile, one attack counted); refused the second time.
- **Wail** — a success takes the full 8d10, a critical success nothing, a failure the same roll and drained 1d4 with its HP loss, a critical failure double and drained 4.
- **Ring of the Ram** — 9d6 at 3 actions with a basic save, once per fight.
- **Relic: Flare Bolt, Tidal Crash, Searing Wave** — the extra damage joins the bomb's damage in one `DamageDealt`; Flare Bolt's persistent fire lands only on a critical hit; Tidal Crash is refused the second time; Searing Wave's save follows a hit and a miss and not a critical miss, and its recharge is rolled each time.
- **Quick Alchemy group** — six quick bombs empty the pool; Sticky adds persistent fire 8; Debilitating dazzles on a failed save; each additive is refused a second time in one round.

Each later commit adds its rows to this test.

##### Commit message

```
feat(E4): enforce limits and add Strike tweaks

- Per-fight, per-round, pool and recharge limits are spent and enforced
- flourish is once per turn and press needs an earlier attack
- A Strike step takes extra dice, extra damage, counts and a shared penalty
- The multiple attack penalty applies on the attacker's own turn only
- A Strike step's own degree lists now run, after the Strike's riders
- gain_actions and recharge steps, turn_ends, per_action damage,
  target = "other" and success_damage = "full"
- An ability stops when its user falls
- First scope rows: Vicious Swing, Brutal Finish, Destiny Spores, Wail,
  Walk the Timelines, Ring of the Ram, the relics and the quick bombs
```

### Phase 3 — Triggers and reactions

**In plain words:** an ability can now fire on a trigger, as a reaction, a free action or a passive effect, and the decider is asked whether to react. When this lands, Vengeful Spite, Lash Out, Attack of Opportunity with its disrupt, Last Stand, Fast Healing and Comforting Presence run from data.

#### Commit 4 — `feat(E4): add triggers, reaction slots and the decider's React`

##### Task 1 — `internal/pf2e/sim/creature/triggers.go`

`internal/pf2e/sim/creature/triggers.go` (new) — full contents:

```go
package creature

import (
	"maps"
	"slices"
	"strings"
)

const (
	TriggerTargeted   = "targeted"
	TriggerAttacked   = "attacked"
	TriggerSaved      = "saved"
	TriggerDamaged    = "damaged"
	TriggerTurnStart  = "turn-start"
	TriggerInitiative = "initiative"
	TriggerFoeActs    = "foe-acts"
	TriggerAllyGains  = "ally-gains"
	TriggerStoodUp    = "stood-up"
	TriggerLoyalty    = "loyalty"

	CostReaction = "reaction"
	CostFree     = "free"

	AttackStrike      = "strike"
	AttackMeleeStrike = "melee-strike"
)

var (
	triggerKinds   = []string{TriggerTargeted, TriggerAttacked, TriggerSaved, TriggerDamaged, TriggerTurnStart, TriggerInitiative, TriggerFoeActs, TriggerAllyGains, TriggerStoodUp, TriggerLoyalty}
	triggerCosts   = []string{"", CostReaction, CostFree}
	triggerAttacks = []string{"", AttackStrike, AttackMeleeStrike}
)

type Trigger struct {
	On         string   `toml:"on"`
	Cost       string   `toml:"cost,omitempty"`
	Attack     string   `toml:"attack,omitempty"`
	Degrees    []string `toml:"degrees,omitempty"`
	Damaging   bool     `toml:"damaging,omitempty"`
	Melee      bool     `toml:"melee,omitempty"`
	Traits     []string `toml:"traits,omitempty"`
	Ranged     bool     `toml:"ranged,omitempty"`
	Conditions []string `toml:"conditions,omitempty"`
	HPAtMost   int      `toml:"hp_at_most,omitzero"`
}

func checkTrigger(path string, ability Ability, report func(string, ...any)) {
	trigger := ability.Trigger
	if !slices.Contains(triggerKinds, trigger.On) {
		report("%s.trigger: on %q: want %s", path, trigger.On, strings.Join(triggerKinds, ", "))
		return
	}
	if !slices.Contains(triggerCosts, trigger.Cost) {
		report("%s.trigger: cost %q: want reaction, free, or leave it unset", path, trigger.Cost)
	}
	if ability.Actions != 0 || ability.MaxActions != 0 || ability.Free || ability.TurnEnds {
		report("%s: a triggered ability sets no actions, max_actions, free or turn_ends; its cost is trigger.cost", path)
	}
	onAttack := trigger.On == TriggerTargeted || trigger.On == TriggerAttacked
	if !slices.Contains(triggerAttacks, trigger.Attack) || (trigger.Attack != "" && !onAttack) {
		report("%s.trigger: attack %q: want strike or melee-strike, on targeted or attacked", path, trigger.Attack)
	}
	if len(trigger.Degrees) > 0 && trigger.On != TriggerAttacked && trigger.On != TriggerSaved {
		report("%s.trigger: degrees belong on attacked or saved", path)
	}
	for _, degree := range trigger.Degrees {
		if !slices.Contains(DegreeWords, degree) {
			report("%s.trigger: degree %q: want %s", path, degree, strings.Join(DegreeWords, ", "))
		}
	}
	if trigger.Damaging && trigger.On != TriggerSaved {
		report("%s.trigger: damaging belongs on saved", path)
	}
	if trigger.Melee && trigger.On != TriggerDamaged {
		report("%s.trigger: melee belongs on damaged", path)
	}
	names := len(trigger.Traits) > 0 || trigger.Ranged
	if names != (trigger.On == TriggerFoeActs) {
		report("%s.trigger: foe-acts needs traits or ranged, and no other kind takes them", path)
	}
	if (len(trigger.Conditions) > 0) != (trigger.On == TriggerAllyGains) {
		report("%s.trigger: ally-gains needs conditions, and no other kind takes them", path)
	}
	for _, name := range trigger.Conditions {
		if !IsCondition(name) {
			report("%s.trigger: %q is not a condition", path, name)
		}
	}
	if trigger.HPAtMost < 0 || (trigger.HPAtMost != 0 && trigger.On != TriggerTurnStart) {
		report("%s.trigger: hp_at_most must be positive, on turn-start", path)
	}
}

func (creature *Creature) checkExtraReactions(report func(string, ...any)) {
	for _, id := range slices.Sorted(maps.Keys(creature.ExtraReactions)) {
		ability, found := creature.Abilities[id]
		if !found || ability.Trigger == nil || ability.Trigger.Cost != CostReaction {
			report("extra_reaction.%s: want an ability whose trigger costs a reaction", id)
		}
		if creature.ExtraReactions[id] < 1 {
			report("extra_reaction.%s: must be at least 1", id)
		}
	}
}
```

An ability with two triggers (Vengeful Spite: a critical hit, or a critically failed save) is written as two abilities with the same steps. One reaction pays for either.

##### Task 2 — `internal/pf2e/sim/creature/creature.go`, `limits.go` and `validate.go`

`creature.go`:

**Find:**
```go
	Pools                map[string]int                 `toml:"pool,omitempty"`
```
**Replace with:**
```go
	Pools                map[string]int                 `toml:"pool,omitempty"`
	ExtraReactions       map[string]int                 `toml:"extra_reaction,omitempty"`
```

**Find:**
```go
	Limit      *Limit   `toml:"limit,omitempty"`
	Steps      []Step   `toml:"steps,omitempty"`
```
**Replace with:**
```go
	Limit      *Limit   `toml:"limit,omitempty"`
	Trigger    *Trigger `toml:"trigger,omitempty"`
	Steps      []Step   `toml:"steps,omitempty"`
```

`limits.go`, in `CostRange`:

**Find:**
```go
	if ability.Free {
		return 0, 0
	}
```
**Replace with:**
```go
	if ability.Free || ability.Trigger != nil {
		return 0, 0
	}
```

`validate.go`, in `prepare`: a triggered ability is checked by `checkTrigger` and not by `checkCost`.

**Find:**
```go
		checkCost("ability."+id, ability, report)
```
**Replace with:**
```go
		if ability.Trigger != nil {
			checkTrigger("ability."+id, ability, report)
		} else {
			checkCost("ability."+id, ability, report)
		}
```

**Find:**
```go
	creature.checkPools(report)
```
**Replace with:**
```go
	creature.checkPools(report)
	creature.checkExtraReactions(report)
```

`show.go`, in `costText`, add a first case: `case ability.Trigger != nil: return triggerText(*ability.Trigger)`, and the helper:

```go
func triggerText(trigger creature.Trigger) string {
	cost := trigger.Cost
	if cost == "" {
		cost = "passive"
	}
	return cost + " on " + trigger.On
}
```

##### Task 3 — `internal/pf2e/sim/engine/events.go`

Add after `ActionsGained`:

```go
type Triggered struct {
	Combatant string
	Ability   string
	On        string
	Cost      string
	By        string
	ByAbility string
	ByStrike  string
}
```

Add `func (Triggered) isEvent() {}` to the `isEvent` block. `DamageDealt` and `Healed` each gain one last field, `TriggeredBy string` (Decision 26).

##### Task 4 — `internal/pf2e/sim/engine/decide.go`

**Find:**
```go
type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
}
```
**Replace with:**
```go
type Decider interface {
	Decide(fight *Fight, actor *Combatant) (Choice, bool)
	NextTarget(fight *Fight, actor *Combatant) (*Combatant, bool)
	React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool)
}
```

**Find:**
```go
		if !ability.Ignore && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
```
**Replace with:**
```go
		if !ability.Ignore && ability.Trigger == nil && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
```

Add at the end of the file:

```go
// deliberate: FirstFitting always reacts with the first ability offered, not by weighing the trigger, because it is a test behaviour; Effort 7's rule lists decide (Decision 21).
func (FirstFitting) React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool) {
	return offered[0], true
}
```

##### Task 5 — `internal/pf2e/sim/engine/triggers.go`

`internal/pf2e/sim/engine/triggers.go` (new) — full contents:

```go
package engine

import (
	"fmt"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type TriggerEvent struct {
	Kind      string
	By        *Combatant
	Ability   string
	Strike    string
	Traits    []string
	Degree    Degree
	IsStrike  bool
	Melee     bool
	Ranged    bool
	Damaging  bool
	Condition string
}

func matches(owner *Combatant, trigger creature.Trigger, event TriggerEvent) bool {
	if trigger.On != event.Kind {
		return false
	}
	switch event.Kind {
	case creature.TriggerTargeted, creature.TriggerAttacked:
		if trigger.Attack != "" && !event.IsStrike {
			return false
		}
		if trigger.Attack == creature.AttackMeleeStrike && !event.Melee {
			return false
		}
		return degreeListed(trigger.Degrees, event.Degree)
	case creature.TriggerSaved:
		return (!trigger.Damaging || event.Damaging) && degreeListed(trigger.Degrees, event.Degree)
	case creature.TriggerDamaged:
		return !trigger.Melee || event.Melee
	case creature.TriggerTurnStart:
		return trigger.HPAtMost == 0 || owner.HP <= trigger.HPAtMost
	case creature.TriggerFoeActs:
		return (trigger.Ranged && event.Ranged) || overlaps(trigger.Traits, event.Traits)
	case creature.TriggerAllyGains:
		return slices.Contains(trigger.Conditions, event.Condition)
	}
	return true
}

func degreeListed(words []string, degree Degree) bool {
	return len(words) == 0 || slices.Contains(words, creature.DegreeWords[degree])
}

func (combatant *Combatant) reactionFor(abilityID string) bool {
	return combatant.extraReactions[abilityID] > 0 || !combatant.ReactionUsed
}

// deliberate: an ability's own extra reaction is spent before the shared one, not after, because Tactical Reflexes exists to keep the shared reaction free.
func (combatant *Combatant) spendReaction(abilityID string) {
	if combatant.extraReactions[abilityID] > 0 {
		combatant.extraReactions[abilityID]--
		return
	}
	combatant.ReactionUsed = true
}

func (fight *Fight) fire(owner *Combatant, event TriggerEvent) {
	if event.By != nil && !event.By.Standing() {
		return
	}
	var offered []string
	for _, id := range owner.Creature.AbilityIDs() {
		if !owner.Standing() {
			return
		}
		ability := owner.Creature.Abilities[id]
		if ability.Ignore || ability.Trigger == nil || !matches(owner, *ability.Trigger, event) || owner.Blocked(ability) != "" {
			continue
		}
		switch {
		case ability.Trigger.Cost == "":
			fight.runTriggered(owner, ability, event)
		case owner.Condition("stunned") > 0:
		case ability.Trigger.Cost == creature.CostFree || owner.reactionFor(id):
			offered = append(offered, id)
		}
	}
	if len(offered) == 0 || !owner.Standing() {
		return
	}
	picked, reacting := fight.decider.React(fight, owner, event, offered)
	if !reacting {
		return
	}
	if !slices.Contains(offered, picked) {
		fight.fault = fmt.Errorf("%w: %s reacted with %q, which was not offered", ErrBadChoice, owner.ID, picked)
		return
	}
	ability := owner.Creature.Abilities[picked]
	if ability.Trigger.Cost == creature.CostReaction {
		owner.spendReaction(picked)
	}
	fight.runTriggered(owner, ability, event)
}

func (fight *Fight) runTriggered(owner *Combatant, ability creature.Ability, event TriggerEvent) {
	target, by := event.By, ""
	if target != nil {
		by = target.ID
	} else {
		next, found := fight.decider.NextTarget(fight, owner)
		if !found {
			return
		}
		target = next
	}
	fight.observer.Observe(Triggered{
		Combatant: owner.ID,
		Ability:   ability.ID,
		On:        event.Kind,
		Cost:      ability.Trigger.Cost,
		By:        by,
		ByAbility: event.Ability,
		ByStrike:  event.Strike,
	})
	fight.spend(owner, ability)
	fight.useAbility(abilityUse{actor: owner, ability: ability, target: target, fixed: event.By != nil, triggeredBy: by})
}
```

A trigger with no causing creature (a turn start, initiative, a Stand) aims its steps at the decider's next target; a step for the owner says `target = "self"`.

##### Task 6 — `internal/pf2e/sim/engine/fight.go`, `combatant.go`, `run.go`, `steps.go`, `damage.go`, `healing.go`

`fight.go`: the fight keeps the first fault a decider causes outside a turn's own choice.

**Find:**
```go
	round      int
	current    *Combatant
}
```
**Replace with:**
```go
	round      int
	current    *Combatant
	fault      error
}
```

`combatant.go`: add the field `extraReactions map[string]int` after `lastPenalty`.

`run.go`, in `run`: a fault ends the fight with its error.

**Find:**
```go
			if err := fight.takeTurn(actor); err != nil {
				return NoOutcome, err
			}
```
**Replace with:**
```go
			if err := fight.takeTurn(actor); err != nil {
				return NoOutcome, err
			}
			if fight.fault != nil {
				return NoOutcome, fight.fault
			}
```

`actions.go`, at the top of `regainActions`: the extra reactions come back when the actions do, for a standing, dying or stable creature alike (Decision Record — Execution 17).

```go
	actor.extraReactions = maps.Clone(actor.Creature.ExtraReactions)
```

Add `maps` to the imports of `actions.go`. In `run.go`, in `perform`, a triggered ability cannot be chosen:

**Find:**
```go
		if !found || ability.Ignore {
```
**Replace with:**
```go
		if !found || ability.Ignore || ability.Trigger != nil {
```

`steps.go`: the use carries whether its target is fixed and who set it off.

**Find:**
```go
	lethal  bool
	actions int
}

type abilityUse struct {
	actor   *Combatant
	ability creature.Ability
	target  *Combatant
	actions int
}
```
**Replace with:**
```go
	lethal      bool
	actions     int
	triggeredBy string
}

type abilityUse struct {
	actor       *Combatant
	ability     creature.Ability
	target      *Combatant
	actions     int
	fixed       bool
	triggeredBy string
}
```

**Find:**
```go
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions}
```
**Replace with:**
```go
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions, triggeredBy: use.triggeredBy}
```

**Find:**
```go
		if step.Target != creature.TargetSelf && !target.Standing() {
			next, found := fight.decider.NextTarget(fight, actor)
```
**Replace with:**
```go
		if step.Target != creature.TargetSelf && !target.Standing() {
			// deliberate: a triggered ability does not move to a new target, not even when its target fell, because it answers one creature.
			if use.fixed {
				continue
			}
			next, found := fight.decider.NextTarget(fight, actor)
```

`damage.go`, in `applyDamage`, and `healing.go`, in `heal`: set `TriggeredBy: source.triggeredBy` in the `DamageDealt` and `Healed` events.

`strike.go`: `strikeOptions` gains a last field, `triggeredBy string`, and `strike` puts it in its source: `source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, triggeredBy: options.triggeredBy}`. `steps.go`, in the strike case of `runStep`, adds `triggeredBy: source.triggeredBy,` to the `strikeOptions` literal (Decision Record — Execution 18).

##### Task 7 — tests

`helpers_test.go`: the scripted decider answers `React` from a list. An empty string declines; with the list used up it declines.

```go
func (decider *scriptedDecider) React(fight *Fight, actor *Combatant, event TriggerEvent, offered []string) (string, bool) {
	if len(decider.reactions) == 0 {
		return "", false
	}
	picked := decider.reactions[0]
	decider.reactions = decider.reactions[1:]
	return picked, picked != ""
}
```

Add the field `reactions []string` to `scriptedDecider`.

`internal/pf2e/sim/creature/triggers_test.go` (new):

- **`TestLoadReadsTriggers`** — one ability per kind loads with its qualifiers; a triggered ability's `CostRange` is 0, 0.
- **`TestLoadRejectsBadTriggers`** — table: an unknown `on`; an unknown `cost`; `actions = 1` beside a trigger; `attack` on `damaged`; an unknown degree word; `damaging` on `attacked`; `melee` on `saved`; `foe-acts` with neither traits nor `ranged`; `traits` on `turn-start`; `ally-gains` with no conditions; `hp_at_most` on `initiative`; `[extra_reaction]` naming a missing ability, a passive one, and a count of 0.

`internal/pf2e/sim/engine/triggers_test.go` (new), each test calling `fight.fire` with a hand-made `TriggerEvent`:

- **`TestPassiveTriggerAlwaysFires`** — a passive ability fires with no question to the decider, twice for two events, and emits `Triggered` with an empty `Cost`.
- **`TestReactionAsksTheDecider`** — the decider is offered the matching, usable abilities in ID order; a decline runs nothing and spends nothing; a pick emits `Triggered{Cost: "reaction"}` and sets `ReactionUsed`.
- **`TestOneReactionPerRound`** — after one reaction a second event offers nothing; after the owner's next `act` it offers again.
- **`TestExtraReaction`** — with `[extra_reaction]` for one ability, that ability can answer twice between turns and the shared reaction is still unspent after the first; another ability gets only the shared one. Before the owner's first turn the extra reaction is not there (trap 3: "at the start of each of your turns").
- **`TestStunnedCannotReact`** — a stunned owner is offered nothing, and its passive trigger still fires.
- **`TestDownedOwnerDoesNotFire`** (trap 5) — a dying or dead owner fires nothing, a passive included; an owner whose own passive drops it is offered no reaction afterwards.
- **`TestBlockedAbilityIsNotOffered`** — a triggered ability with `per_fight = 1` fires once.
- **`TestFreeTriggerNeedsNoReaction`** — a `cost = "free"` ability is offered with `ReactionUsed` already true.
- **`TestReactWithUnofferedAbilityFaults`** — a decider that returns an ID it was not offered ends `Run` with `ErrBadChoice`.
- **`TestTriggeredAbilityKeepsItsTarget`** (trap 1) — the causing creature falls before the reaction's second step: the step is skipped and `NextTarget` is not asked.
- **`TestTriggerWithNoCauseAimsAtAFoe`** — a `turn-start` event runs a damage step at the first standing foe and a `target = "self"` step at the owner.
- **`TestTriggeredDamageNamesItsCause`** — `DamageDealt.TriggeredBy` holds the causing creature's ID.
- **`TestMatches`** — table over `matches`: each kind with its qualifier met and unmet (attack kinds, degrees, damaging, melee, HP threshold at, above and below, traits, ranged, conditions).
- **`TestFirstFittingSkipsTriggered`** — `Decide` never returns a triggered ability, and `perform` refuses one.

##### Commit message

```
feat(E4): add triggers, reaction slots and the decider's React

- An ability with a trigger table fires on one of ten kinds of event
- A trigger costs a reaction, a free action or nothing (passive)
- The decider answers React: which offered ability, or none
- [extra_reaction] gives one ability reactions of its own
- A Triggered event names the owner, the ability, the kind and the cause
- DamageDealt and Healed name the creature that set a triggered ability off
```

#### Commit 5 — `feat(E4): fire triggers on attacks, saves and damage`

##### Task 1 — `internal/pf2e/sim/engine/strike.go`

A Strike fires `targeted` on its target before the roll and `attacked` after its lists. The source records whether the Strike was melee.

**Find:**
```go
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
```
**Replace with:**
```go
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
	// deliberate: the target is told after the concealment check, not before, because a reaction to an attack that never rolls would be spent on nothing.
	attack := TriggerEvent{By: attacker, Ability: options.ability, Strike: strike.ID, Traits: strike.Traits, IsStrike: true, Melee: strike.Melee()}
	attack.Kind = creature.TriggerTargeted
	fight.fire(target, attack)
	if !attacker.Standing() {
		return Failure, false
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
```

**Find:**
```go
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, triggeredBy: options.triggeredBy}
```
**Replace with:**
```go
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, melee: strike.Melee(), triggeredBy: options.triggeredBy}
```

**Find:**
```go
	fight.runSteps(source, options.lists.For(int(degree), true), target)
	return degree, true
```
**Replace with:**
```go
	fight.runSteps(source, options.lists.For(int(degree), true), target)
	attack.Kind, attack.Degree = creature.TriggerAttacked, degree
	fight.fire(target, attack)
	return degree, true
```

##### Task 2 — `internal/pf2e/sim/engine/steps.go`

Add `melee bool` to `effectSource`, after `lethal`.

In `spellAttack`, the same two events, with `IsStrike` unset.

**Find:**
```go
	if !fight.passesConcealment(attacker, target) {
		return
	}
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
```
**Replace with:**
```go
	if !fight.passesConcealment(attacker, target) {
		return
	}
	attack := TriggerEvent{Kind: creature.TriggerTargeted, By: attacker, Ability: source.ability, Traits: source.traits}
	fight.fire(target, attack)
	if !attacker.Standing() {
		return
	}
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
```

**Find:**
```go
	if !target.Dead {
		fight.runSteps(source, step.For(int(degree), true), target)
	}
}
```
**Replace with:**
```go
	if !target.Dead {
		fight.runSteps(source, step.For(int(degree), true), target)
	}
	attack.Kind, attack.Degree = creature.TriggerAttacked, degree
	fight.fire(target, attack)
}
```

In `saveStep`, a creature that saved against a foe's effect fires `saved`.

**Find:**
```go
		fight.runSteps(source, step.For(int(degree), false), target)
	}
}
```
**Replace with:**
```go
		fight.runSteps(source, step.For(int(degree), false), target)
		if source.actor.Side != target.Side {
			fight.fire(target, TriggerEvent{
				Kind:     creature.TriggerSaved,
				By:       source.actor,
				Ability:  source.ability,
				Traits:   source.traits,
				Degree:   degree,
				Damaging: dealsDamage(step),
			})
		}
	}
}

func dealsDamage(step creature.Step) bool {
	if len(step.Damage) > 0 {
		return true
	}
	for _, list := range [][]creature.Step{step.CriticalSuccess, step.Success, step.Failure, step.CriticalFailure} {
		if slices.ContainsFunc(list, dealsDamage) {
			return true
		}
	}
	return false
}
```

##### Task 3 — `internal/pf2e/sim/engine/damage.go` and `persistent.go`

`damage.go`, at the end of `applyDamage`: a creature that took damage from a foe fires `damaged`.

**Find:**
```go
	if !target.Dead {
		for _, part := range roll.persistent {
			fight.addPersistent(source, target, part, roll.doubled)
		}
	}
}
```
**Replace with:**
```go
	if !target.Dead {
		for _, part := range roll.persistent {
			fight.addPersistent(source, target, part, roll.doubled)
		}
	}
	if total > 0 && source.actor.Side != target.Side && source.ability != persistentAbility {
		fight.fire(target, TriggerEvent{
			Kind:    creature.TriggerDamaged,
			By:      source.actor,
			Ability: source.ability,
			Strike:  source.strike,
			Traits:  source.traits,
			Melee:   source.melee,
		})
	}
}
```

`persistent.go`: a persistent tick is not a new hit, and it must not keep the melee mark of the Strike that set it (trap 1).

**Find:**
```go
const persistentRecoveryDC = 15
```
**Replace with:**
```go
const (
	persistentRecoveryDC = 15
	persistentAbility    = "persistent"
)
```

**Find:**
```go
		source.ability, source.strike, source.lethal = "persistent", "", false
```
**Replace with:**
```go
		source.ability, source.strike, source.lethal, source.melee = persistentAbility, "", false, false
```

##### Task 4 — `testdata/scope/avatar-of-lilith/hand.toml` and `testdata/scope/arabesque/hand.toml`

Add to the Avatar's file:

```toml
# Vengeful Spite has two triggers, so it is two abilities with the same steps.
[ability.vengeful-spite]
name    = "Vengeful Spite"
traits  = ["occult"]
trigger = { on = "attacked", cost = "reaction", degrees = ["critical-success"] }
steps   = [
  { save = "will", dc = 38, basic = true, damage = [{ dice = "4d10+14", type = "mental" }],
    failure = [{ condition = "frightened", value = 2 }],
    critical_failure = [{ condition = "frightened", value = 2 }, { condition = "stunned", value = 4 }] },
]

[ability.vengeful-spite-save]
name    = "Vengeful Spite (failed save)"
traits  = ["occult"]
trigger = { on = "saved", cost = "reaction", degrees = ["critical-failure"], damaging = true }
steps   = [
  { save = "will", dc = 38, basic = true, damage = [{ dice = "4d10+14", type = "mental" }],
    failure = [{ condition = "frightened", value = 2 }],
    critical_failure = [{ condition = "frightened", value = 2 }, { condition = "stunned", value = 4 }] },
]
```

Add to Arabesque's file:

```toml
# approximation: the 2d6 void; the shove waits for A.004.2.
[ability.lash-out]
name    = "Lash Out"
traits  = ["occult", "void"]
trigger = { on = "damaged", cost = "reaction", melee = true }
steps   = [{ damage = [{ dice = "2d6", type = "void" }] }]
```

##### Task 5 — tests

`internal/pf2e/sim/engine/triggers_test.go`:

- **`TestTargetedFiresBeforeTheRoll`** — a passive `targeted` ability on the target shows its `Triggered` event before `StrikeRolled`; with `attack = "melee-strike"` it does not fire for a ranged Strike or a spell attack; with no `attack` it fires for all three.
- **`TestTargetedWaitsForConcealment`** — an attack that fails its concealment flat check fires nothing.
- **`TestAttackedFiresAfterDamage`** — a `degrees = ["critical-success"]` reaction fires after `DamageDealt` and the riders, on a critical hit only.
- **`TestSavedFiresForTheSaver`** — fires with the saver's degree; `damaging = true` skips a save step with no damage anywhere in it and takes one whose damage sits in a degree list; a save against an ally's effect fires nothing.
- **`TestDamagedFires`** — fires on damage from a foe; not on 0 damage after resistance; not on a persistent tick, even one set by a melee Strike (trap 1); `melee = true` skips a ranged Strike and a save's damage.
- **`TestReactionCanDropTheAttacker`** (trap 5) — an attacker using a three-Strike ability is dropped by the target's reaction after the first Strike: no second `StrikeRolled`, and the turn's action loop ends.
- **`TestReactionsDoNotLoop`** — two foes that each react to damage with damage exchange one reaction each and stop.
- **`TestKnockedOutTargetCannotReact`** — a PC dropped by the hit that would trigger its reaction does not react.

`scope_test.go`, new rows:

- **Vengeful Spite** — Kas critically hits the Avatar: the decider reacts, Kas saves against DC 38; a failure is full damage and frightened 2, a critical failure double damage, frightened 2 and stunned 4. Stunned on his own turn (Decision Record — Execution 5 and 19): Vicious Swing as his first action leaves 1 action, which stunned 4 takes at once (`ActionsLost` 1, left 0), leaving stunned 3; his next turn loses 3 (`ActionsRegained` with 0 actions, 3 lost) and ends at stunned 0; the turn after is a full 3. A PC already frightened 1 still goes to frightened 2 (Decision 6). The Avatar critically fails a damaging save from a PC: the second ability fires. With her reaction spent, neither fires.
- **Lash Out** — the Avatar's Slash damages Arabesque: 2d6 void back, less her resistance 12 (so 0 at most: 6 and 6 roll 12 and she takes 0); a ranged hit does not trigger it; a hit that drops Arabesque does not.

##### Commit message

```
feat(E4): fire triggers on attacks, saves and damage

- targeted fires before an attack roll, after the concealment check
- attacked fires with the roll's degree, after damage and riders
- saved fires for a creature that saved against a foe's effect
- damaged fires for damage from a foe; a persistent tick does not count
- Scope rows: Vengeful Spite, Lash Out
```

#### Commit 6 — `feat(E4): fire triggers on turns, initiative, actions and conditions`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`, `steps.go`, `conditions.go`, `validate.go`

`creature.go`: fast healing is a number on the creature.

**Find:**
```go
	RecoveryDC           int                            `toml:"recovery_dc,omitzero"`
```
**Replace with:**
```go
	RecoveryDC           int                            `toml:"recovery_dc,omitzero"`
	FastHealing          int                            `toml:"fast_healing,omitzero"`
```

`validate.go`:

**Find:**
```go
	if creature.HeroPoints < 0 || creature.RecoveryDC < 0 {
		report("hero_points and recovery_dc must not be negative")
	}
```
**Replace with:**
```go
	if creature.HeroPoints < 0 || creature.RecoveryDC < 0 || creature.FastHealing < 0 {
		report("hero_points, recovery_dc and fast_healing must not be negative")
	}
```

`steps.go`: a Strike step can disrupt the action that triggered it, and a condition step can lower a value.

**Find:**
```go
	Penalty       string         `toml:"penalty,omitempty"`
```
**Replace with:**
```go
	Penalty       string         `toml:"penalty,omitempty"`
	Disrupts      []string       `toml:"disrupts,omitempty"`
	Reduce        int            `toml:"reduce,omitzero"`
```

**Find:**
```go
	if kind != StepStrike && (step.ExtraDice != 0 || len(step.ExtraDamage) > 0 || step.Counts != 0 || step.Penalty != "") {
		report("%s: extra_dice, extra_damage, counts and penalty belong on a strike", path)
	}
```
**Replace with:**
```go
	if kind != StepStrike && (step.ExtraDice != 0 || len(step.ExtraDamage) > 0 || step.Counts != 0 || step.Penalty != "" || len(step.Disrupts) > 0) {
		report("%s: extra_dice, extra_damage, counts, penalty and disrupts belong on a strike", path)
	}
	if step.Reduce != 0 && (kind != StepCondition || step.Reduce < 0) {
		report("%s: reduce belongs on a condition and must be positive", path)
	}
```

`conditions.go`, in `checkCondition`: a step that lowers a value sets no value of its own, and is not a stunned step with no value (Decision Record — Execution 20).

**Find:**
```go
	stunnedWithNoValue := step.Condition == "stunned" && step.Value == 0 && step.ValueDice == ""
```
**Replace with:**
```go
	stunnedWithNoValue := step.Condition == "stunned" && step.Value == 0 && step.ValueDice == "" && step.Reduce == 0
```

**Find:**
```go
	case valued && step.Value < 1 && step.ValueDice == "":
		report("%s: %s needs a value of at least 1", path, step.Condition)
```
**Replace with:**
```go
	case step.Reduce != 0 && (!valued || step.Value != 0 || step.ValueDice != "" || step.Stack || step.Duration != nil):
		report("%s: reduce needs a valued condition and no value, stack or duration", path)
	case valued && step.Reduce == 0 && step.Value < 1 && step.ValueDice == "":
		report("%s: %s needs a value of at least 1", path, step.Condition)
```

##### Task 2 — `internal/pf2e/sim/engine/events.go` and `combatant.go`

`events.go`, add after `Triggered`:

```go
type ActionDisrupted struct {
	Combatant string
	Action    string
	By        string
	Ability   string
}
```

Add `func (ActionDisrupted) isEvent() {}` to the `isEvent` block.

`combatant.go`: add the field `downed bool` after `extraReactions`.

##### Task 3 — `internal/pf2e/sim/engine/triggers.go`

Add at the end of the file:

```go
type pendingAction struct {
	actor     *Combatant
	action    string
	traits    []string
	disrupted bool
}

// deliberate: every standing foe may answer, not only those in reach, because every creature counts as engaged with every foe until A.004.2's position (Decision 12).
func (fight *Fight) provoke(actor *Combatant, action string, event TriggerEvent) bool {
	outer := fight.pending
	pending := &pendingAction{actor: actor, action: action, traits: event.Traits}
	fight.pending = pending
	event.Kind, event.By = creature.TriggerFoeActs, actor
	for _, foe := range fight.StandingFoes(actor) {
		fight.fire(foe, event)
	}
	fight.pending = outer
	return !pending.disrupted && actor.Standing()
}

func (fight *Fight) disrupt(attacker *Combatant, ability string, traits []string) {
	pending := fight.pending
	if pending == nil || pending.disrupted || !overlaps(traits, pending.traits) {
		return
	}
	pending.disrupted = true
	fight.observer.Observe(ActionDisrupted{Combatant: pending.actor.ID, Action: pending.action, By: attacker.ID, Ability: ability})
}

func (fight *Fight) gained(bearer *Combatant, condition string) {
	for _, ally := range fight.combatants {
		if ally != bearer && ally.Side == bearer.Side {
			fight.fire(ally, TriggerEvent{Kind: creature.TriggerAllyGains, By: bearer, Condition: condition})
		}
	}
}

func (fight *Fight) fastHeal(actor *Combatant) {
	amount := actor.Creature.FastHealing
	if amount > 0 && actor.Standing() && actor.HP < actor.MaxHP() {
		fight.heal(effectSource{actor: actor, ability: "fast-healing"}, actor, amount)
	}
}
```

`fight.go`: add the field `pending *pendingAction` after `fault`.

The fix commit after Commit 6 changes this code: `fastHeal` guards on `!actor.Dead`, not `actor.Standing()` (Decision Record — Execution 22), and `fire` gains a depth guard and `setFault` (entries 23 and 24).

##### Task 4 — `internal/pf2e/sim/engine/run.go`

In `run`, each combatant's `initiative` triggers fire once, in initiative order, before round 1.

**Find:**
```go
func (fight *Fight) run() (Outcome, error) {
	for fight.round = 1; fight.round <= fight.maxRounds; fight.round++ {
```
**Replace with:**
```go
func (fight *Fight) run() (Outcome, error) {
	for _, combatant := range slices.Clone(fight.order) {
		fight.fire(combatant, TriggerEvent{Kind: creature.TriggerInitiative})
	}
	if fight.fault != nil {
		return NoOutcome, fight.fault
	}
	for fight.round = 1; fight.round <= fight.maxRounds; fight.round++ {
```

In `act`, the turn-start triggers fire first, then fast healing (Decision 4).

**Find:**
```go
	fight.regainActions(actor)
	for actor.ActionsLeft > 0 && actor.Standing() {
```
**Replace with:**
```go
	fight.regainActions(actor)
	// deliberate: turn-start triggers see the HP the turn starts with, not the HP after fast healing, because Robert ruled Last Stand first (Decision 4).
	fight.fire(actor, TriggerEvent{Kind: creature.TriggerTurnStart})
	fight.fastHeal(actor)
	for actor.ActionsLeft > 0 && actor.Standing() {
```

The fix commit after Commit 6 adds a downed step: a creature at 0 HP gets its fast healing after `tickLimits` and before the recovery check, and the standing branch's `fight.fastHeal(actor)` runs only when that step did not (Decision Record — Execution 22).

In `perform`, a Stand and an ability give every foe the chance to answer before they resolve. A disrupted ability has spent its actions and its use (Decision 12).

**Find:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: "stand", Actions: 1})
		fight.stand(actor)
		return nil
```
**Replace with:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: "stand", Actions: 1})
		if fight.provoke(actor, "stand", TriggerEvent{Traits: []string{"move"}}) {
			fight.stand(actor)
		}
		return nil
```

**Find:**
```go
		fight.spend(actor, ability)
		fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
```
**Replace with:**
```go
		fight.spend(actor, ability)
		if fight.provoke(actor, ability.ID, TriggerEvent{Ability: ability.ID, Traits: ability.Traits}) {
			fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
		}
```

##### Task 5 — `internal/pf2e/sim/engine/strike.go`

A ranged Strike provokes before anything else; a Strike step with `disrupts` disrupts on a critical hit.

**Find:**
```go
	counts      int
	penalty     string
	lists       creature.Outcomes
	triggeredBy string
}
```
**Replace with:**
```go
	counts      int
	penalty     string
	disrupts    []string
	lists       creature.Outcomes
	triggeredBy string
}
```

**Find:**
```go
func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, options strikeOptions) (Degree, bool) {
	penalty := fight.strikePenalty(attacker, strike, options)
```
**Replace with:**
```go
func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, options strikeOptions) (Degree, bool) {
	if !strike.Melee() && !fight.provoke(attacker, strike.ID, TriggerEvent{Ability: options.ability, Strike: strike.ID, Traits: strike.Traits, Ranged: true}) {
		return Failure, false
	}
	penalty := fight.strikePenalty(attacker, strike, options)
```

**Find:**
```go
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, melee: strike.Melee(), triggeredBy: options.triggeredBy}
	critical := degree == CriticalSuccess
```
**Replace with:**
```go
	source := effectSource{actor: attacker, ability: options.ability, strike: strike.ID, traits: strike.Traits, melee: strike.Melee(), triggeredBy: options.triggeredBy}
	critical := degree == CriticalSuccess
	if critical {
		fight.disrupt(attacker, options.ability, options.disrupts)
	}
```

`steps.go`, in the strike case of `runStep`, add `disrupts: step.Disrupts,` to the `strikeOptions` literal.

A ranged Strike inside a manipulate ability provokes twice, once for the ability and once for the Strike. That is right: Quick Alchemy and the throw are two actions.

##### Task 6 — `internal/pf2e/sim/engine/conditions.go`, `actions.go`, `dying.go`, `steps.go`

`conditions.go`, in `applyCondition`: a raised value tells the bearer's allies.

**Find:**
```go
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	drainedBefore := target.Condition("drained")
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
```
**Replace with:**
```go
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	drainedBefore, before := target.Condition("drained"), target.Condition(step.Condition)
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
	if target.Condition(step.Condition) > before {
		fight.gained(target, step.Condition)
	}
```

`steps.go`, in the condition case of `runStep`:

**Find:**
```go
		for _, target := range targets {
			fight.applyCondition(source.actor, target, step)
		}
```
**Replace with:**
```go
		for _, target := range targets {
			if step.Reduce > 0 {
				fight.reduceCondition(target, step.Condition, step.Reduce)
				continue
			}
			fight.applyCondition(source.actor, target, step)
		}
```

`dying.go`, in `knockDown`: mark the PC as downed, so its next Stand counts as getting up from being knocked out.

**Find:**
```go
	fight.applyCondition(attacker, target, creature.Step{Condition: "prone"})
```
**Replace with:**
```go
	fight.applyCondition(attacker, target, creature.Step{Condition: "prone"})
	target.downed = true
```

`actions.go`, in `stand`:

**Find:**
```go
	fight.observer.Observe(StoodUp{Combatant: actor.ID})
}
```
**Replace with:**
```go
	fight.observer.Observe(StoodUp{Combatant: actor.ID})
	if actor.downed {
		actor.downed = false
		fight.fire(actor, TriggerEvent{Kind: creature.TriggerStoodUp})
	}
}
```

Add the `creature` import to `actions.go`.

##### Task 7 — scope files

Add to `testdata/scope/avatar-of-lilith/hand.toml`, at the top level, `fast_healing = 10`, and:

```toml
# approximation: the heal, the Wail recharge and the extra action; the move to The Realm Between is dropped.
[ability.last-stand]
name    = "Last Stand"
trigger = { on = "turn-start", cost = "free", hp_at_most = 75 }
limit   = { per_fight = 1 }
steps   = [{ heal = "40", target = "self" }, { recharge = "wail" }, { gain_actions = 1 }]
```

Add to `testdata/scope/kas/hand.toml`:

```toml
# Tactical Reflexes: one more reaction each turn, for Attack of Opportunity only.
[extra_reaction]
attack-of-opportunity = 1

# approximation: every foe counts as in reach; only Stand carries `move` until A.004.2.
[ability.attack-of-opportunity]
name    = "Attack of Opportunity"
trigger = { on = "foe-acts", cost = "reaction", traits = ["manipulate", "move"], ranged = true }
steps   = [{ strike = "protectors-fury", disrupts = ["manipulate"] }]
```

Add to `testdata/scope/arabesque/hand.toml`:

```toml
# stand-in: the triggering ally only; one ability per condition. Adjacency is A.004.2's.
[ability.comforting-presence-frightened]
name    = "Comforting Presence (frightened)"
traits  = ["concentrate", "emotion", "mental"]
trigger = { on = "ally-gains", cost = "reaction", conditions = ["frightened"] }
steps   = [
  { condition = "frightened", reduce = 1 },
  { condition = "frightened", value = 1, stack = true, target = "self" },
]

[ability.comforting-presence-stupefied]
name    = "Comforting Presence (stupefied)"
traits  = ["concentrate", "emotion", "mental"]
trigger = { on = "ally-gains", cost = "reaction", conditions = ["stupefied"] }
steps   = [
  { condition = "stupefied", reduce = 1 },
  { condition = "stupefied", value = 1, stack = true, target = "self" },
]
```

No boss ability in the scope files carries `manipulate` yet (Ill Omen brings it in Commit 7), so the Attack of Opportunity row gives the boss a test ability with that trait in Go.

##### Task 8 — tests

`internal/pf2e/sim/engine/triggers_test.go`:

- **`TestTurnStartTrigger`** — fires after the actions are set, so a `gain_actions` step adds to them; `hp_at_most = 75` fires at 75 and not at 76.
- **`TestFastHealing`** — heals at the start of the turn, not above maximum HP; a `Healed` event with ability `fast-healing`. After the fix commit (Decision Record — Execution 22) a stable or dying creature at 0 HP is healed to 10, and a dead one is not.
- **`TestTurnStartBeforeFastHealing`** — a creature at 70 HP with `hp_at_most = 75` and `fast_healing = 10` fires its trigger, then heals (Decision 4).
- **`TestInitiativeTrigger`** — fires once per combatant, in initiative order, after every `InitiativeRolled` and before `RoundStarted`; a condition it applies with a duration still ends on time (trap 1: `newTiming` runs with no current actor).
- **`TestFoeActsOnTraits`** — a manipulate ability provokes before its steps and after `ActionUsed`; an ability with no listed trait does not; a ranged Strike provokes and a melee one does not; a Stand provokes a `move` trigger.
- **`TestDisrupt`** — a critical reaction Strike with `disrupts = ["manipulate"]` against a manipulate ability: no step of the ability runs, `ActionDisrupted` names both creatures, the actions are spent, a `per_fight` use is spent and a recharge is rolled. A normal hit does not disrupt. A critical hit against a Stand does not disrupt it (the Stand is `move`).
- **`TestProvokeCanDropTheActor`** (trap 5) — the reaction drops the acting creature: its ability does not run and its turn ends.
- **`TestReactionStrikeHasNoPenalty`** (trap 1) — the reactor made three Strikes on its own turn; its reaction Strike on the foe's turn rolls at penalty 0 and leaves `AttacksThisTurn` at 3.
- **`TestAllyGainsTrigger`** — fires for allies when a value rises, not when the same value is applied again, not for the bearer, not for foes; a `reduce` step lowers the ally's value and removes the condition at 0.
- **`TestStoodUpTrigger`** — fires on the first Stand after a knock-out and not on a Stand from an ordinary prone; not a second time.

`creature/steps_test.go`: bad cases for `disrupts` on a save, `reduce` on a heal, `reduce` with `value`, `reduce` on `off-guard`. `TestLoadReadsReduceDisruptsAndFastHealing` loads `condition = "stunned", reduce = 2` cleanly beside a `frightened` reduce, a Strike's `disrupts` and `fast_healing` (Decision Record — Execution 20). `creature/fight_test.go`: the negative hero points case reads the new message, and a negative `fast_healing` case joins it.

`scope_test.go`, new rows:

- **Fast Healing 10 and Last Stand** — the Avatar starts a turn at 70 HP: Last Stand heals 40, Wail (recharging) is ready, she has 4 actions, then Fast Healing takes her to 120. It never fires again. At 76 HP only Fast Healing runs.
- **Attack of Opportunity, Tactical Reflexes** — a boss uses a manipulate ability twice in a round: Kas Strikes at no penalty both times (one extra reaction, one shared) and not a third time. A critical hit disrupts the ability. A boss knocked prone by Kas Stands and takes a Strike (Decision 12).
- **Comforting Presence** — Kas gains frightened 2: with Arabesque's reaction he is at 1 and she at 1; with her reaction spent on Lash Out nothing changes.

##### Commit message

```
feat(E4): fire triggers on turns, initiative, actions and conditions

- turn-start fires after the actions are set and before fast healing
- initiative fires once per creature before round 1
- foe-acts fires on a foe's ability by trait, a ranged Strike and a Stand
- A critical reaction Strike can disrupt the action that triggered it
- ally-gains fires when an ally's condition rises; a condition step can reduce
- stood-up fires on the first Stand after a knock-out
- fast_healing on the creature
- Scope rows: Fast Healing, Last Stand, Attack of Opportunity, Tactical
  Reflexes, Comforting Presence
```

### Phase 4 — Rerolls, healing and loyalty

**In plain words:** a d20 can be rolled twice, a PC can aim an ability at itself or its allies, consumables share counts, an effect can be sustained, and an act of loyalty hurts the Avatar and makes the next Strike holy. When this lands, Challenge Fate, Ill Omen, Farabellus Flip, the potions, Healing Wave, Haste, Sin Vulnerability and Swear Fidelity run from data.

#### Commit 7 — `feat(E4): add fortune, misfortune and granted bonuses`

##### Task 1 — `internal/pf2e/sim/creature/steps.go`

Two step kinds: `reroll` puts fortune or misfortune on the target's rolls, and `bonus` gives the target a fight-time modifier. Both take `uses` and a duration.

**Find:**
```go
	StepRecharge    = "recharge"
```
**Replace with:**
```go
	StepRecharge    = "recharge"
	StepReroll      = "reroll"
	StepBonus       = "bonus"

	RerollWorse  = "worse"
	RerollBetter = "better"
	RollAttack   = "attack"
	RollSave     = "save"
	RollSkill    = "skill"
```

**Find:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken, StepGainActions, StepRecharge}
```
**Replace with:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken, StepGainActions, StepRecharge, StepReroll, StepBonus}
	timedSteps  = []string{StepCondition, StepWeaken, StepReroll, StepBonus}
	usedSteps   = []string{StepReroll, StepBonus}
	rollKinds   = []string{RollAttack, RollSave, RollSkill}
```

**Find:**
```go
	Reduce        int            `toml:"reduce,omitzero"`
```
**Replace with:**
```go
	Reduce        int            `toml:"reduce,omitzero"`
	Reroll        string         `toml:"reroll,omitempty"`
	Rolls         []string       `toml:"rolls,omitempty"`
	Bonus         *CheckModifier `toml:"bonus,omitempty"`
	Uses          int            `toml:"uses,omitzero"`
```

In `heads`, before `return heads`:

```go
	if step.Reroll != "" {
		heads = append(heads, StepReroll)
	}
	if step.Bonus != nil {
		heads = append(heads, StepBonus)
	}
```

In `checkStep`, add two cases to the `switch kind`:

```go
	case StepReroll:
		if step.Reroll != RerollWorse && step.Reroll != RerollBetter {
			report("%s: reroll %q: want worse or better", path, step.Reroll)
		}
		if len(step.Rolls) == 0 || slices.ContainsFunc(step.Rolls, func(kind string) bool { return !slices.Contains(rollKinds, kind) }) {
			report("%s: rolls: want one or more of %s", path, strings.Join(rollKinds, ", "))
		}
	case StepBonus:
		bonus := step.Bonus
		if !slices.Contains(CheckSelectors, bonus.Selector) || !slices.Contains(modifierTypes, bonus.Type) || bonus.Value == 0 || bonus.PerCondition != "" {
			report("%s.bonus: want a known selector and type, a value other than 0, and no per_condition", path)
		}
		if step.Uses > 0 && bonus.Selector != "ac" {
			report("%s: uses on a bonus needs selector ac", path)
		}
```

**Find:**
```go
	if kind != StepCondition && kind != StepWeaken && step.Duration != nil {
		report("%s: duration belongs on a condition or weaken", path)
	}
```
**Replace with:**
```go
	if !slices.Contains(timedSteps, kind) && step.Duration != nil {
		report("%s: duration belongs on one of %s", path, strings.Join(timedSteps, ", "))
	}
	if len(step.Rolls) > 0 && kind != StepReroll {
		report("%s: rolls belongs on a reroll", path)
	}
	switch used := slices.Contains(usedSteps, kind); {
	case step.Uses < 0 || (step.Uses != 0 && !used):
		report("%s: uses belongs on one of %s and must not be negative", path, strings.Join(usedSteps, ", "))
	case used && step.Uses == 0 && step.Duration == nil:
		report("%s: %s needs uses or a duration", path, kind)
	}
```

The duration message changed; update the case in `steps_test.go` that reads the old one.

##### Task 2 — `internal/pf2e/sim/engine/events.go`

Add after `SaveRolled`:

```go
type RerollGained struct {
	Combatant string
	Source    string
	Ability   string
	Keep      string
}

type Rerolled struct {
	Combatant string
	Roll      string
	Keep      string
	First     int
	Second    int
	Kept      int
	Ability   string
}

type RerollsCancelled struct {
	Combatant string
	Roll      string
}

type BonusGained struct {
	Combatant string
	Source    string
	Ability   string
	Selector  string
	Value     int
}
```

Add all four to the `isEvent` block.

##### Task 3 — `internal/pf2e/sim/engine/rolls.go`

`internal/pf2e/sim/engine/rolls.go` (new) — full contents:

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type rerollEffect struct {
	keep    string
	rolls   []string
	uses    int
	ability string
	clock   timing
}

type grantedBonus struct {
	modifier creature.CheckModifier
	uses     int
	clock    timing
}

func (fight *Fight) d20(roller *Combatant, kind string) int {
	die := fight.roller.Die(20)
	worse := roller.takeReroll(creature.RerollWorse, kind)
	better := roller.takeReroll(creature.RerollBetter, kind)
	switch {
	case worse == nil && better == nil:
		return die
	case worse != nil && better != nil:
		// deliberate: fortune and misfortune on one roll cancel, not stack, because the rules say the roll is then made once.
		fight.observer.Observe(RerollsCancelled{Combatant: roller.ID, Roll: kind})
		return die
	}
	second := fight.roller.Die(20)
	applied, kept := better, max(die, second)
	if worse != nil {
		applied, kept = worse, min(die, second)
	}
	fight.observer.Observe(Rerolled{Combatant: roller.ID, Roll: kind, Keep: applied.keep, First: die, Second: second, Kept: kept, Ability: applied.ability})
	return kept
}

func (combatant *Combatant) takeReroll(keep, kind string) *rerollEffect {
	for index, effect := range combatant.rerolls {
		if effect.keep != keep || !slices.Contains(effect.rolls, kind) {
			continue
		}
		if effect.uses > 0 {
			effect.uses--
			if effect.uses == 0 {
				combatant.rerolls = slices.Delete(combatant.rerolls, index, index+1)
			}
		}
		return effect
	}
	return nil
}

func (combatant *Combatant) spendBonuses(selector string) {
	combatant.bonuses = slices.DeleteFunc(combatant.bonuses, func(granted *grantedBonus) bool {
		if granted.uses == 0 || granted.modifier.Selector != selector {
			return false
		}
		granted.uses--
		return granted.uses == 0
	})
}

func (fight *Fight) addReroll(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.rerolls = append(target.rerolls, &rerollEffect{
		keep:    step.Reroll,
		rolls:   step.Rolls,
		uses:    step.Uses,
		ability: source.ability,
		clock:   fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(RerollGained{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability, Keep: step.Reroll})
}

func (fight *Fight) addBonus(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.bonuses = append(target.bonuses, &grantedBonus{
		modifier: *step.Bonus,
		uses:     step.Uses,
		clock:    fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(BonusGained{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability, Selector: step.Bonus.Selector, Value: step.Bonus.Value})
}

func (fight *Fight) expireEffects(bearer, actor *Combatant, boundary string) {
	bearer.rerolls = slices.DeleteFunc(bearer.rerolls, func(effect *rerollEffect) bool {
		return effect.clock.expires(actor, bearer, boundary)
	})
	bearer.bonuses = slices.DeleteFunc(bearer.bonuses, func(granted *grantedBonus) bool {
		return granted.clock.expires(actor, bearer, boundary)
	})
}
```

An effect with `uses` and no duration lasts until it is used. `combatant.go`: add the fields `rerolls []*rerollEffect` and `bonuses []*grantedBonus` after `downed`.

##### Task 4 — `internal/pf2e/sim/engine/conditions.go`, `checks.go`, `strike.go`, `steps.go`

`conditions.go`, in both `startTurn` and `endTurn`: the new effects end with the conditions.

**Find:**
```go
		fight.expire(bearer, actor, "start")
		fight.expireDefenseChanges(bearer, actor, "start")
```
**Replace with:**
```go
		fight.expire(bearer, actor, "start")
		fight.expireDefenseChanges(bearer, actor, "start")
		fight.expireEffects(bearer, actor, "start")
```

**Find:**
```go
		fight.expire(bearer, actor, "end")
		fight.expireDefenseChanges(bearer, actor, "end")
```
**Replace with:**
```go
		fight.expire(bearer, actor, "end")
		fight.expireDefenseChanges(bearer, actor, "end")
		fight.expireEffects(bearer, actor, "end")
```

`checks.go`, in `modifierPool`: granted bonuses join the creature's own modifiers.

**Find:**
```go
	fight.conditionPenalties(combatant, selectors, pool)
	return pool
```
**Replace with:**
```go
	for _, granted := range combatant.bonuses {
		if granted.modifier.Applies(selectors, traits) {
			pool.Add(granted.modifier.Type, granted.modifier.Value)
		}
	}
	fight.conditionPenalties(combatant, selectors, pool)
	return pool
```

In `savingThrow`:

**Find:**
```go
	die := fight.roller.Die(20)
	total := die + bonus
	degree := shiftDegree(DegreeOf(die, total, dc), shift)
```
**Replace with:**
```go
	die := fight.d20(target, creature.RollSave)
	total := die + bonus
	degree := shiftDegree(DegreeOf(die, total, dc), shift)
```

`strike.go`, in `strike`: the attack roll goes through `d20`, and an AC bonus with `uses` is spent by the attack that read it.

**Find:**
```go
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, strike.Traits)
```
**Replace with:**
```go
	die := fight.d20(attacker, creature.RollAttack)
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, strike.Traits)
	target.spendBonuses("ac")
```

`steps.go`, in `spellAttack`:

**Find:**
```go
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, source.traits)
```
**Replace with:**
```go
	die := fight.d20(attacker, creature.RollAttack)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, source.traits)
	target.spendBonuses("ac")
```

In `runStep`, add two cases to the kind switch at the bottom:

```go
	case creature.StepReroll:
		for _, target := range targets {
			fight.addReroll(source, target, step)
		}
	case creature.StepBonus:
		for _, target := range targets {
			fight.addBonus(source, target, step)
		}
```

Flat checks, recovery checks and initiative stay on `fight.roller.Die(20)`: fortune and misfortune do not touch them.

##### Task 5 — scope files

Add to `testdata/scope/abyssal-luna/hand.toml`:

```toml
[ability.challenge-fate]
name    = "Challenge Fate"
traits  = ["misfortune", "occult"]
trigger = { on = "targeted", cost = "reaction" }
steps   = [{ reroll = "worse", rolls = ["attack"], uses = 1 }]

[ability.ill-omen]
name    = "Ill Omen"
actions = 2
traits  = ["concentrate", "curse", "manipulate", "misfortune"]
steps   = [
  { save = "will", dc_from = "spell",
    failure = [{ reroll = "worse", rolls = ["attack", "skill"], uses = 1, duration = { rounds = 1 } }],
    critical_failure = [{ reroll = "worse", rolls = ["attack", "skill"], duration = { rounds = 1 } }] },
]
```

Add to `testdata/scope/kas/hand.toml`:

```toml
# approximation: the +2 AC; the Step waits for A.004.2.
[ability.farabellus-flip]
name    = "Farabellus Flip"
trigger = { on = "targeted", cost = "reaction", attack = "melee-strike" }
steps   = [{ bonus = { selector = "ac", type = "circumstance", value = 2 }, uses = 1, target = "self" }]
```

##### Task 6 — tests

`internal/pf2e/sim/engine/rolls_test.go` (new):

- **`TestMisfortuneKeepsTheWorse`** and **`TestFortuneKeepsTheBetter`** — two dice come off the scripted roller, `Rerolled` carries both and the kept one, and the degree follows the kept die, natural 1 and 20 included.
- **`TestFortuneAndMisfortuneCancel`** (trap 3) — both on one roll: one die, `RerollsCancelled`, and each effect loses a use.
- **`TestRerollUses`** — `uses = 1` is gone after one roll; no `uses` with `rounds = 1` covers every roll until the source's next turn and then ends; `uses = 1` with `rounds = 1` ends unused at the source's next turn.
- **`TestRerollOnlyItsRolls`** — an `attack` reroll leaves a save and a flat check alone; a `save` reroll leaves an attack alone.
- **`TestGrantedBonus`** — an AC bonus with `uses = 1` raises the AC of one attack and is gone for the next; a status bonus with a duration follows the stacking rule against the creature's own status bonus; it ends at its boundary.
- **`TestBonusIsSpentByAMiss`** — the attack that read the bonus spends it whatever its degree.

`creature/steps_test.go`: bad cases for `reroll = "twice"`, no `rolls`, an unknown roll kind, a reroll with neither `uses` nor a duration, `uses` on a bonus to a save, a bonus with value 0, `uses` on a damage step.

`scope_test.go`, new rows:

- **Challenge Fate** — Kas Strikes Luna and she reacts: two attack dice, the worse kept; his second Strike rolls once (one reaction a round). Arabesque's spell attack triggers it too (Decision 14). A Strike that fails its concealment check leaves her reaction unspent.
- **Ill Omen** — a failed Will save: the PC's next attack rolls twice and the one after once; a critical failure: every attack until Luna's next turn; a success: nothing. Kas's Attack of Opportunity answers the manipulate trait.
- **Farabellus Flip** — a Slash that would hit by 1 misses with the reaction; a spell attack offers no reaction; the second Slash of a frenzy sees his normal AC.

##### Commit message

```
feat(E4): add fortune, misfortune and granted bonuses

- A reroll step makes the target roll twice and keep the worse or better
- Fortune and misfortune on one roll cancel
- Attack rolls and saves go through one d20 function
- A bonus step gives a fight-time modifier; an AC bonus can last one attack
- Scope rows: Challenge Fate, Ill Omen, Farabellus Flip
```

#### Commit 8 — `feat(E4): target allies, share consumables and sustain effects`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`, `steps.go`, `validate.go`

`creature.go`: an ability says whom it is for, and what a Sustain does.

**Find:**
```go
	Trigger    *Trigger `toml:"trigger,omitempty"`
	Steps      []Step   `toml:"steps,omitempty"`
```
**Replace with:**
```go
	Trigger    *Trigger `toml:"trigger,omitempty"`
	Targets    string   `toml:"targets,omitempty"`
	Sustain    []Step   `toml:"sustain,omitempty"`
	Steps      []Step   `toml:"steps,omitempty"`
```

`steps.go`: the target words, and a step that eases recovery checks.

**Find:**
```go
	StepBonus       = "bonus"
```
**Replace with:**
```go
	StepBonus       = "bonus"
	StepRecovery    = "recovery_dc"

	TargetsSelf   = "self"
	TargetsAlly   = "ally"
	TargetsOthers = "others"
```

Add `StepRecovery` to `stepKinds` and to `timedSteps`. Add `abilityTargets = []string{"", TargetsSelf, TargetsAlly, TargetsOthers}` to the `var` block. Add the field `RecoveryDC int` with the tag `toml:"recovery_dc,omitzero"` after `Uses`. In `heads`, before `return heads`:

```go
	if step.RecoveryDC != 0 {
		heads = append(heads, StepRecovery)
	}
```

In `checkStep`, add a case:

```go
	case StepRecovery:
		if step.RecoveryDC < 1 || step.Duration == nil {
			report("%s: recovery_dc must be at least 1 and needs a duration", path)
		}
```

`validate.go`, in `prepare`, after the loop over `ability.Steps`:

```go
		if !slices.Contains(abilityTargets, ability.Targets) {
			report("ability.%s: targets %q: want self, ally, others, or leave it unset", id, ability.Targets)
		}
		if ability.Targets == TargetsOthers && ability.Area == nil {
			report("ability.%s: targets others needs an area", id)
		}
		if len(ability.Sustain) > 0 && ability.Trigger != nil {
			report("ability.%s: a triggered ability cannot be sustained", id)
		}
		for index := range ability.Sustain {
			creature.checkStep(fmt.Sprintf("ability.%s.sustain[%d]", id, index), &ability.Sustain[index], report)
		}
```

- **`targets = "self"`** — the user; the choice needs no target. It takes no area: the fix commit after Phase 4 reports `ability.<id>: targets self takes no area` (Decision Record — Execution 34).
- **`targets = "ally"`** — one living creature of the user's side, dying or not. With an area: every living creature of that side, the user included.
- **`targets = "others"`** — with an area: every living creature of that side but the user.

##### Task 2 — `internal/pf2e/sim/engine/events.go` and `combatant.go`

`events.go`, add after `RecoveryChecked`:

```go
type RecoveryEased struct {
	Combatant string
	Ability   string
	DC        int
}

type Sustained struct {
	Combatant string
	Ability   string
}

type SustainEnded struct {
	Combatant string
	Ability   string
}
```

Add all three to the `isEvent` block.

`combatant.go`: add the fields `recoveryCaps []*recoveryCap` and `sustained map[string]*sustainedEffect` after `bonuses`, and `sustained: map[string]*sustainedEffect{},` to the literal in `NewCombatant`.

##### Task 3 — `internal/pf2e/sim/engine/allies.go`

`internal/pf2e/sim/engine/allies.go` (new) — full contents:

```go
package engine

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type recoveryCap struct {
	dc    int
	clock timing
}

type sustainedEffect struct {
	target *Combatant
	done   bool
}

func (fight *Fight) Allies(actor *Combatant) []*Combatant {
	var allies []*Combatant
	for _, combatant := range fight.combatants {
		if combatant.Side == actor.Side && !combatant.Dead {
			allies = append(allies, combatant)
		}
	}
	return allies
}

// deliberate: an area catches the whole side it is for, not a share of it, because every creature counts as engaged with every other until A.004.2's position (E3 Decision 1, Decision 13).
func (fight *Fight) areaTargets(actor *Combatant, ability creature.Ability) []*Combatant {
	switch ability.Targets {
	case creature.TargetsAlly:
		return fight.Allies(actor)
	case creature.TargetsOthers:
		return slices.DeleteFunc(fight.Allies(actor), func(ally *Combatant) bool { return ally == actor })
	default:
		return fight.StandingFoes(actor)
	}
}

func targetID(target *Combatant) string {
	if target == nil {
		return ""
	}
	return target.ID
}

func (fight *Fight) checkTarget(actor *Combatant, ability creature.Ability, target *Combatant) bool {
	switch {
	case ability.Targets == creature.TargetsSelf, ability.Area != nil && ability.Targets != "":
		return true
	case ability.Targets == creature.TargetsAlly:
		return target != nil && !target.Dead && target.Side == actor.Side
	default:
		return target != nil && target.Standing()
	}
}

func (fight *Fight) easeRecovery(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.recoveryCaps = append(target.recoveryCaps, &recoveryCap{dc: step.RecoveryDC, clock: fight.newTiming(source.actor, target, step.Duration)})
	fight.observer.Observe(RecoveryEased{Combatant: target.ID, Ability: source.ability, DC: step.RecoveryDC})
}

func (combatant *Combatant) Sustaining() []string {
	var waiting []string
	for _, id := range slices.Sorted(maps.Keys(combatant.sustained)) {
		if !combatant.sustained[id].done {
			waiting = append(waiting, id)
		}
	}
	return waiting
}

// deliberate: an effect ends at the end of a turn with no Sustain, not when the next turn starts, because "sustained up to" lasts to the end of the caster's next turn only if it is sustained there.
func (fight *Fight) endSustains(actor *Combatant) {
	for _, id := range slices.Sorted(maps.Keys(actor.sustained)) {
		effect := actor.sustained[id]
		if effect.done {
			effect.done = false
			continue
		}
		delete(actor.sustained, id)
		fight.observer.Observe(SustainEnded{Combatant: actor.ID, Ability: id})
	}
}
```

##### Task 4 — `internal/pf2e/sim/engine/steps.go`

Delete the old `areaTargets` and its fence-sign from `steps.go`; `allies.go` has the new one. In `useAbility`:

**Find:**
```go
	if ability.Area != nil {
		targets := fight.areaTargets(actor)
```
**Replace with:**
```go
	friendly := ability.Targets != ""
	if ability.Targets == creature.TargetsSelf {
		target = actor
	}
	if ability.Area != nil {
		targets := fight.areaTargets(actor, ability)
```

**Find:**
```go
	if len(ability.Steps) > 0 && ability.Steps[0].Kind() == creature.StepSave && ability.Steps[0].Target != creature.TargetSelf && !fight.passesConcealment(actor, target) {
```
**Replace with:**
```go
	if !friendly && len(ability.Steps) > 0 && ability.Steps[0].Kind() == creature.StepSave && ability.Steps[0].Target != creature.TargetSelf && !fight.passesConcealment(actor, target) {
```

**Find:**
```go
		if step.Target != creature.TargetSelf && !target.Standing() {
			// deliberate: a triggered ability does not move to a new target, not even when its target fell, because it answers one creature.
			if use.fixed {
				continue
			}
			next, found := fight.decider.NextTarget(fight, actor)
```
**Replace with:**
```go
		if step.Target != creature.TargetSelf && !target.Standing() && !(friendly && !target.Dead) {
			// deliberate: a triggered or friendly ability does not move to a new target, not even when its target fell, because it is for one creature; a dying ally still counts.
			if use.fixed || friendly {
				continue
			}
			next, found := fight.decider.NextTarget(fight, actor)
```

In `runStep`, add a case to the kind switch at the bottom:

```go
	case creature.StepRecovery:
		for _, target := range targets {
			fight.easeRecovery(source, target, step)
		}
```

##### Task 5 — `internal/pf2e/sim/engine/run.go`, `decide.go`, `dying.go`, `conditions.go`, `rolls.go`

`decide.go`: add `ChooseSustain` after `ChooseStand`. `FirstFitting` Sustains what waits, and never picks an ability that is for its own side.

**Find:**
```go
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Trigger == nil && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
```
**Replace with:**
```go
	if waiting := actor.Sustaining(); len(waiting) > 0 && actor.OpenActions() > 0 {
		return Choice{Kind: ChooseSustain, ID: waiting[0]}, true
	}
	for _, id := range actor.Creature.AbilityIDs() {
		ability := actor.Creature.Abilities[id]
		if !ability.Ignore && ability.Trigger == nil && ability.Targets == "" && ability.Actions <= actor.OpenActions() && actor.Blocked(ability) == "" {
```

`run.go`, in `perform`: Sustain is handled beside Stand, and the target check depends on the choice.

**Find:**
```go
	if choice.Target == nil || !choice.Target.Standing() {
		return fmt.Errorf("%w: %s chose %q with no standing target", ErrBadChoice, actor.ID, choice.ID)
	}
	switch choice.Kind {
	case ChooseStrike:
		strike, found := actor.Creature.Strikes[choice.ID]
		if !found {
			return fmt.Errorf("%w: %s has no strike %q", ErrBadChoice, actor.ID, choice.ID)
		}
```
**Replace with:**
```go
	if choice.Kind == ChooseSustain {
		effect, found := actor.sustained[choice.ID]
		if !found || effect.done || actor.OpenActions() < 1 {
			return fmt.Errorf("%w: %s cannot Sustain %q", ErrBadChoice, actor.ID, choice.ID)
		}
		ability := actor.Creature.Abilities[choice.ID]
		actor.ActionsLeft--
		effect.done = true
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: 1, Target: targetID(effect.target)})
		fight.observer.Observe(Sustained{Combatant: actor.ID, Ability: choice.ID})
		ability.Steps = ability.Sustain
		fight.useAbility(abilityUse{actor: actor, ability: ability, target: effect.target, actions: 1, fixed: true})
		return nil
	}
	switch choice.Kind {
	case ChooseStrike:
		strike, found := actor.Creature.Strikes[choice.ID]
		if !found {
			return fmt.Errorf("%w: %s has no strike %q", ErrBadChoice, actor.ID, choice.ID)
		}
		if choice.Target == nil || !choice.Target.Standing() {
			return fmt.Errorf("%w: %s chose %q with no standing target", ErrBadChoice, actor.ID, choice.ID)
		}
```

In the ability case:

**Find:**
```go
		cost := ability.Actions
		if choice.Actions != 0 {
```
**Replace with:**
```go
		if !fight.checkTarget(actor, ability, choice.Target) {
			return fmt.Errorf("%w: %s chose %q with no fitting target", ErrBadChoice, actor.ID, choice.ID)
		}
		cost := ability.Actions
		if choice.Actions != 0 {
```

**Find:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: cost, Target: choice.Target.ID})
		fight.spend(actor, ability)
		if fight.provoke(actor, ability.ID, TriggerEvent{Ability: ability.ID, Traits: ability.Traits}) {
			fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
		}
```
**Replace with:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: cost, Target: targetID(choice.Target)})
		fight.spend(actor, ability)
		if fight.provoke(actor, ability.ID, TriggerEvent{Ability: ability.ID, Traits: ability.Traits}) {
			fight.useAbility(abilityUse{actor: actor, ability: ability, target: choice.Target, actions: cost})
			if len(ability.Sustain) > 0 && actor.Standing() {
				actor.sustained[ability.ID] = &sustainedEffect{target: choice.Target, done: true}
			}
		}
```

`conditions.go`, in `endTurn`, after `fight.decayConditions(actor)`: add `fight.endSustains(actor)`.

`rolls.go`, in `expireEffects`, add:

```go
	bearer.recoveryCaps = slices.DeleteFunc(bearer.recoveryCaps, func(eased *recoveryCap) bool {
		return eased.clock.expires(actor, bearer, boundary)
	})
```

`dying.go`, in `recoveryCheck`: an eased check uses the lower DC (Decision 17).

**Find:**
```go
	dc := base + target.Dying
```
**Replace with:**
```go
	dc := base + target.Dying
	for _, eased := range target.recoveryCaps {
		dc = min(dc, eased.dc)
	}
```

##### Task 6 — scope files

Add to `testdata/scope/kas/hand.toml`, in a new `[pool]` table before the first `[strike.*]`:

```toml
# Starting counts from the sheet; dials for A.004.2.
[pool]
healing-potion-moderate = 3
healing-potion-greater  = 1
```

and:

```toml
# approximation: the cone catches Kas and every ally; spread is A.004.2's.
[ability.healing-wave]
name    = "Healing Wave"
actions = 2
traits  = ["concentrate", "healing", "vitality"]
area    = { shape = "cone", size = 15 }
targets = "ally"
limit   = { per_fight = 1 }
steps   = [{ heal = "14d4" }]

[ability.drink-healing-potion-moderate]
name    = "Healing Potion (Moderate)"
actions = 1
traits  = ["manipulate"]
targets = "self"
limit   = { pool = "healing-potion-moderate" }
steps   = [{ heal = "3d8+10" }]

[ability.feed-healing-potion-moderate]
name    = "Healing Potion (Moderate), to an ally"
actions = 1
traits  = ["manipulate"]
targets = "ally"
limit   = { pool = "healing-potion-moderate" }
steps   = [{ heal = "3d8+10" }]

[ability.drink-healing-potion-greater]
name    = "Healing Potion (Greater)"
actions = 1
traits  = ["manipulate"]
targets = "self"
limit   = { pool = "healing-potion-greater" }
steps   = [{ heal = "6d8+20" }]

[ability.feed-healing-potion-greater]
name    = "Healing Potion (Greater), to an ally"
actions = 1
traits  = ["manipulate"]
targets = "ally"
limit   = { pool = "healing-potion-greater" }
steps   = [{ heal = "6d8+20" }]
```

Add to `testdata/scope/tavish/hand.toml`, in its `[pool]` table, `elixir-of-life-moderate = 1` and `healing-potion-moderate = 1`, and:

```toml
[ability.drink-elixir-of-life]
name    = "Elixir of Life (Moderate)"
actions = 1
traits  = ["manipulate"]
targets = "self"
limit   = { pool = "elixir-of-life-moderate" }
steps   = [{ heal = "5d6+12" }]

[ability.feed-elixir-of-life]
name    = "Elixir of Life (Moderate), to an ally"
actions = 1
traits  = ["manipulate"]
targets = "ally"
limit   = { pool = "elixir-of-life-moderate" }
steps   = [{ heal = "5d6+12" }]

[ability.drink-healing-potion-moderate]
name    = "Healing Potion (Moderate)"
actions = 1
traits  = ["manipulate"]
targets = "self"
limit   = { pool = "healing-potion-moderate" }
steps   = [{ heal = "3d8+10" }]

[ability.feed-healing-potion-moderate]
name    = "Healing Potion (Moderate), to an ally"
actions = 1
traits  = ["manipulate"]
targets = "ally"
limit   = { pool = "healing-potion-moderate" }
steps   = [{ heal = "3d8+10" }]

[ability.bewitching-bloom-cherry-blossom]
name    = "Bewitching Bloom (Cherry Blossom)"
actions = 2
traits  = ["concentrate"]
targets = "ally"
limit   = { per_fight = 1 }
steps   = [{ recovery_dc = 10, duration = { rounds = 10 } }]
```

Add to `testdata/scope/arabesque/hand.toml`:

```toml
# Starting count from the sheet; a dial for A.004.2.
[pool]
therapeutic-snap-peas = 5

[ability.therapeutic-snap-peas]
name    = "Therapeutic Snap Peas"
actions = 1
traits  = ["concentrate", "manipulate"]
area    = { shape = "emanation", size = 30 }
targets = "ally"
limit   = { pool = "therapeutic-snap-peas" }
steps   = [{ heal = "4d8+10" }]

# Rank 7: up to six creatures. Effort 5 adds the slot.
[ability.haste]
name    = "Haste"
actions = 2
traits  = ["concentrate", "manipulate"]
area    = { shape = "emanation", size = 30 }
targets = "ally"
steps   = [{ condition = "quickened", duration = { rounds = 10 } }]
```

Put the `[pool]` table before the first `[strike.*]` table in each file.

##### Task 7 — tests

`internal/pf2e/sim/engine/allies_test.go` (new):

- **`TestSelfAbilityNeedsNoTarget`** — a `targets = "self"` heal with a nil target heals the user; `ActionUsed.Target` is empty.
- **`TestAllyAbility`** — a heal aimed at a dying ally lifts it (`Revived`); aimed at a foe, a dead ally or nil it fails with `ErrBadChoice`; the retarget question is never asked.
- **`TestAllyArea`** — `targets = "ally"` with an area reaches every living creature of the side, the user and a dying ally included, and no foe; `targets = "others"` leaves the user out; one heal roll serves all.
- **`TestAreaStillCatchesFoes`** (trap 1) — an area ability with no `targets` catches every standing foe, as before.
- **`TestFirstFittingSkipsFriendlyAbilities`** — with only a self heal usable, it Strikes.
- **`TestRecoveryEased`** — under `recovery_dc = 10` a PC at dying 3 with the default DC rolls against 10, not 13; a PC whose own DC plus dying is 7 rolls against 7 (Decision 17); after the duration the DC is back.
- **`TestSustain`** — an ability with a `sustain` list: on the turn after it is used, `FirstFitting` Sustains first (`ActionUsed`, `Sustained`, the list's steps); a second Sustain in the same turn fails; a turn that ends with no Sustain emits `SustainEnded` and a later Sustain fails; an actor that is dying at the end of a turn in which it sustained keeps the effect through that turn's end and loses it at the end of its next turn, which it spends down (trap 5; Decision Record — Execution 35); a Sustain whose target fell runs no step for it.
- **`TestSustainNeedsAnOpenAction`** — refused with only the Strike-only action left.

`creature/steps_test.go` and `limits_test.go`: bad cases for an unknown `targets`, `others` without an area, `sustain` on a triggered ability, a bad step inside `sustain`, `recovery_dc` with no duration.

`scope_test.go`, new rows:

- **Healing potions** (Kas) — drinking costs 1 action and spends the pool; feeding heals a dying Tavish and shares the same count; the fourth Moderate potion is refused.
- **Elixir of Life, Healing Potion** (Tavish) — one of each, then both pools are empty.
- **Healing Wave** — one 14d4 roll heals Kas and both allies, a dying one included; refused the second time.
- **Therapeutic Snap Peas** — all three PCs heal the same amount; five uses.
- **Haste** — Arabesque, Kas and Tavish are quickened for 10 rounds (Decision Record — Execution 38); Kas's fourth action can be a Strike and cannot be Vicious Swing.
- **Bewitching Bloom (Cherry Blossom)** — a dying ally's recovery DC is capped at 10 for 10 rounds; once per fight.

##### Commit message

```
feat(E4): target allies, share consumables and sustain effects

- An ability's targets can be self, ally or others
- An area for the user's side reaches every living creature on it
- Consumables are abilities that spend a shared pool
- A recovery_dc step caps a PC's recovery DC for a time
- An ability with a sustain list can be Sustained once a turn
- Scope rows: healing potions, Elixir of Life, Healing Wave, Therapeutic
  Snap Peas, Haste, Bewitching Bloom (Cherry Blossom)
```

#### Commit 9 — `feat(E4): add acts of loyalty and granted traits`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`, `steps.go`, `validate.go`

`creature.go`: an ability carries tags, which mean something to the engine and are not PF2e traits.

**Find:**
```go
	Traits     []string `toml:"traits,omitempty"`
	Area       *Area    `toml:"area,omitempty"`
```
**Replace with:**
```go
	Traits     []string `toml:"traits,omitempty"`
	Tags       []string `toml:"tags,omitempty"`
	Area       *Area    `toml:"area,omitempty"`
```

`steps.go`:

**Find:**
```go
	StepRecovery    = "recovery_dc"
```
**Replace with:**
```go
	StepRecovery    = "recovery_dc"
	StepGrant       = "grant"

	TagLoyalty      = "loyalty"
	TagTestsLoyalty = "tests-loyalty"
```

Add `StepGrant` to `stepKinds`, `timedSteps` and `usedSteps`. Add `abilityTags = []string{TagLoyalty, TagTestsLoyalty}` to the `var` block. Add the field `Grant string` with the tag `toml:"grant,omitempty"` after `RecoveryDC`. In `heads`, before `return heads`:

```go
	if step.Grant != "" {
		heads = append(heads, StepGrant)
	}
```

In `checkStep`, add a case:

```go
	case StepGrant:
		if !IsDefenseTrait(step.Grant) || step.Uses < 1 {
			report("%s: grant %q: want one of %s, with uses of at least 1", path, step.Grant, strings.Join(DefenseTraits, ", "))
		}
```

`validate.go`, in `prepare`: a tagged ability may have no steps (Swear Fidelity is only an act of loyalty).

**Find:**
```go
		if len(ability.Steps) == 0 {
			report("ability.%s: no steps", id)
		}
```
**Replace with:**
```go
		if len(ability.Steps) == 0 && len(ability.Tags) == 0 {
			report("ability.%s: no steps", id)
		}
		for _, tag := range ability.Tags {
			if !slices.Contains(abilityTags, tag) {
				report("ability.%s: tag %q: want %s", id, tag, strings.Join(abilityTags, ", "))
			}
		}
```

- **`loyalty`** — using the ability is an act of loyalty.
- **`tests-loyalty`** — on a boss ability: a creature that succeeds at a save against it has done an act of loyalty. Effort 5 puts it on Confusion, Unfathomable Song and Warp Mind.

##### Task 2 — `internal/pf2e/sim/engine/events.go` and `combatant.go`

`events.go`, add after `Triggered`:

```go
type LoyaltyAct struct {
	Combatant string
	Reason    string
}

type TraitGranted struct {
	Combatant string
	Trait     string
	Against   string
	Ability   string
}

type TraitSpent struct {
	Combatant string
	Trait     string
	Against   string
}
```

Add all three to the `isEvent` block.

`combatant.go`: add the fields `grants []*grantedTrait` and `loyalRound int` after `sustained`.

##### Task 3 — `internal/pf2e/sim/engine/loyalty.go`

`internal/pf2e/sim/engine/loyalty.go` (new) — full contents:

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type grantedTrait struct {
	trait   string
	against *Combatant
	uses    int
	clock   timing
}

func answersLoyalty(foe *Combatant) bool {
	return slices.ContainsFunc(foe.Creature.AbilityIDs(), func(id string) bool {
		ability := foe.Creature.Abilities[id]
		return !ability.Ignore && ability.Trigger != nil && ability.Trigger.On == creature.TriggerLoyalty
	})
}

func (fight *Fight) loyalty(actor *Combatant, reason string) {
	// deliberate: once per creature per initiative round, not per turn of that creature, because Robert ruled the round (Decision 7). The mark is round+1 so the zero value means "never".
	if !actor.Standing() || actor.loyalRound == fight.round+1 {
		return
	}
	var answering []*Combatant
	for _, foe := range fight.StandingFoes(actor) {
		if answersLoyalty(foe) {
			answering = append(answering, foe)
		}
	}
	if len(answering) == 0 {
		return
	}
	actor.loyalRound = fight.round + 1
	fight.observer.Observe(LoyaltyAct{Combatant: actor.ID, Reason: reason})
	for _, foe := range answering {
		fight.fire(foe, TriggerEvent{Kind: creature.TriggerLoyalty, By: actor, Ability: reason})
	}
}

func (fight *Fight) grantTrait(source effectSource, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	target.grants = append(target.grants, &grantedTrait{
		trait:   step.Grant,
		against: source.actor,
		uses:    step.Uses,
		clock:   fight.newTiming(source.actor, target, step.Duration),
	})
	fight.observer.Observe(TraitGranted{Combatant: target.ID, Trait: step.Grant, Against: source.actor.ID, Ability: source.ability})
}

// deliberate: a granted trait is spent by the attack that carries it, hit or miss, not by the first hit, because Robert ruled the text over the rule elements (Decision 2).
func (fight *Fight) takeGrants(attacker, target *Combatant) []string {
	var traits []string
	attacker.grants = slices.DeleteFunc(attacker.grants, func(grant *grantedTrait) bool {
		if grant.against != target {
			return false
		}
		traits = append(traits, grant.trait)
		fight.observer.Observe(TraitSpent{Combatant: attacker.ID, Trait: grant.trait, Against: target.ID})
		grant.uses--
		return grant.uses < 1
	})
	return traits
}
```

A granted trait holds only against the creature whose ability granted it: Faithful Strike is "against the avatar".

##### Task 4 — `internal/pf2e/sim/engine/strike.go`, `steps.go`, `run.go`, `healing.go`, `rolls.go`

`strike.go`, in `strike`: the Strike carries its granted traits from the moment it is made.

**Find:**
```go
	penalty := fight.strikePenalty(attacker, strike, options)
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
```
**Replace with:**
```go
	penalty := fight.strikePenalty(attacker, strike, options)
	if granted := fight.takeGrants(attacker, target); len(granted) > 0 {
		strike.Traits = append(slices.Clone(strike.Traits), granted...)
	}
	if !fight.passesConcealment(attacker, target) {
		return Failure, false
	}
```

`strike` is a value, so the change lasts for this Strike only; every later line reads `strike.Traits`.

`steps.go`: the source knows the ability's tags. Add `tags []string` to `effectSource` after `triggeredBy`, and set it in `useAbility`:

**Find:**
```go
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions, triggeredBy: use.triggeredBy}
```
**Replace with:**
```go
	source := effectSource{actor: actor, ability: ability.ID, traits: ability.Traits, actions: use.actions, triggeredBy: use.triggeredBy, tags: ability.Tags}
```

In `saveStep`: a passed save against a marked ability is an act of loyalty.

**Find:**
```go
		fight.runSteps(source, step.For(int(degree), false), target)
		if source.actor.Side != target.Side {
```
**Replace with:**
```go
		fight.runSteps(source, step.For(int(degree), false), target)
		if degree >= Success && source.actor.Side != target.Side && slices.Contains(source.tags, creature.TagTestsLoyalty) {
			fight.loyalty(target, source.ability)
		}
		if source.actor.Side != target.Side {
```

In `runStep`, add a case to the kind switch at the bottom:

```go
	case creature.StepGrant:
		for _, target := range targets {
			fight.grantTrait(source, target, step)
		}
```

`run.go`, in `perform`: a tagged ability that went ahead is an act of loyalty.

**Find:**
```go
			if len(ability.Sustain) > 0 && actor.Standing() {
				actor.sustained[ability.ID] = &sustainedEffect{target: choice.Target, done: true}
			}
		}
```
**Replace with:**
```go
			if len(ability.Sustain) > 0 && actor.Standing() {
				actor.sustained[ability.ID] = &sustainedEffect{target: choice.Target, done: true}
			}
			if slices.Contains(ability.Tags, creature.TagLoyalty) {
				fight.loyalty(actor, ability.ID)
			}
		}
```

`healing.go`: lifting an ally from 0 HP, dying or stable, is an act of loyalty.

**Find:**
```go
	if wasDown && target.HP > 0 && target.Dying > 0 {
		from := target.Dying
		target.Dying = 0
		target.Wounded++
		fight.observer.Observe(Revived{Combatant: target.ID, DyingFrom: from, Wounded: target.Wounded})
	}
```
**Replace with:**
```go
	if !wasDown || target.HP == 0 {
		return
	}
	if target.Dying > 0 {
		from := target.Dying
		target.Dying = 0
		target.Wounded++
		fight.observer.Observe(Revived{Combatant: target.ID, DyingFrom: from, Wounded: target.Wounded})
	}
	if source.actor != target && source.actor.Side == target.Side {
		fight.loyalty(source.actor, "revive")
	}
```

`rolls.go`, in `expireEffects`, add:

```go
	bearer.grants = slices.DeleteFunc(bearer.grants, func(grant *grantedTrait) bool {
		return grant.clock.expires(actor, bearer, boundary)
	})
```

##### Task 5 — scope files

Add to `testdata/scope/avatar-of-lilith/hand.toml`:

```toml
# stand-in: tagged abilities, revives and passed saves fire it; the acting PC gets Faithful Strike.
# The +12 in the roll offsets her resistance 12, so the damage goes through her defenses as usual.
[ability.sin-vulnerability]
name    = "Sin Vulnerability (Betrayal)"
trigger = { on = "loyalty" }
steps   = [
  { damage = [{ dice = "4d6+12", type = "mental" }], target = "self" },
  { grant = "holy", uses = 1, duration = { until = "target-turn-end" } },
]
```

Add to each of the three PC scope files:

```toml
[ability.swear-fidelity]
name    = "Swear Fidelity"
actions = 1
traits  = ["auditory", "concentrate"]
tags    = ["loyalty"]
targets = "self"
```

Add `tags = ["loyalty"]` to every `feed-*` ability in the Kas and Tavish scope files, and to `[ability.haste]` in Arabesque's (a spell that benefits only allies).

##### Task 6 — tests

`internal/pf2e/sim/engine/loyalty_test.go` (new):

- **`TestLoyaltyNeedsAnAnswer`** — against a foe with no `loyalty` trigger an act emits nothing and leaves the actor's mark unset, so a later act in the same round against a foe that answers still fires (trap 3: phase 1 has no Avatar).
- **`TestLoyaltyOncePerRound`** — a second act by the same PC in one initiative round does nothing; another PC's act fires; the same PC fires again in the next round, even when a moved initiative place makes its turns adjacent (Decision 7).
- **`TestReviveIsAnActOfLoyalty`** — a heal that lifts a dying ally fires for the healer; one that lifts a stable ally at 0 HP fires too; a heal on a standing ally does not; fast healing on oneself does not. A `loyalty` ability that also revives fires once.
- **`TestPassedSaveIsAnActOfLoyalty`** — a success or critical success against a `tests-loyalty` ability fires for the saver, on the boss's turn; a failure does not; a save against an untagged ability does not.
- **`TestDisruptedAbilityIsNoAct`** — a `loyalty` ability that is disrupted fires nothing.
- **`TestGrantedTrait`** — a Strike against the granting creature carries the trait: its `DamageDealt` shows the weakness and no resistance; a Strike against another creature does not carry or spend it; it is spent by a miss, and by a Strike that fails its concealment check; `TraitSpent` fires once.
- **`TestGrantedTraitExpires`** — granted on the holder's own turn, it lasts through the end of its next turn; granted on another creature's turn, it ends at the end of the holder's next turn.
- **`TestLoyaltyDamageCanKill`** (trap 5) — an Avatar at 5 HP dies of the damage: no trait is granted, and the fight ends with the party's win after the action.

`creature/steps_test.go`: bad cases for an unknown tag, `grant = "fire"`, a grant with no uses, an untagged ability with no steps.

`scope_test.go`, new rows:

- **Swear Fidelity, Sin Vulnerability, Faithful Strike** — Kas uses Swear Fidelity: the Avatar takes 4d6+12 mental less her resistance 12, `DamageDealt.TriggeredBy` is Kas, and Kas holds holy. His next Strike ignores her resistance and adds weakness 15, and the one after does not. A second Swear Fidelity in the round does nothing; Tavish's still fires.
- **Feeding a potion** — Tavish feeds a dying Kas an elixir: one act of loyalty, for Tavish.
- **Haste** — Arabesque's Haste on her side, herself included, is an act of loyalty; she holds the granted holy.

##### Commit message

```
feat(E4): add acts of loyalty and granted traits

- A loyalty tag, a revive of an ally and a passed save against a
  tests-loyalty ability are acts of loyalty, once per creature per round
- A loyalty trigger answers them; a grant step gives a trait against the
  granting creature, spent by the next Strike, hit or miss
- An ability with a tag may have no steps
- Scope rows: Swear Fidelity, Sin Vulnerability, Faithful Strike
```

### Phase 5 — Skills

**In plain words:** a creature gets skill modifiers, computed from a PC's build, and an ability step can roll a skill against a foe's DC. When this lands, Demoralize, Battle Cry and Too Angry to Die run from data, and the three PCs' skill numbers match their sheets.

#### Commit 10 — `feat(E4): compute skills from the build and import their ranks`

Before this commit: Robert has added the skill numbers to the gate 1 fixture (Open Questions).

##### Task 1 — `internal/pf2e/sim/creature/skills.go`

`internal/pf2e/sim/creature/skills.go` (new) — full contents:

```go
package creature

import (
	"maps"
	"slices"
)

const (
	SkillCheckSelector  = "skill-check"
	ProficiencyModifier = "proficiency"
)

var (
	skillAttributes = map[string]string{
		"acrobatics": "dex", "arcana": "int", "athletics": "str", "crafting": "int",
		"deception": "cha", "diplomacy": "cha", "intimidation": "cha", "medicine": "wis",
		"nature": "wis", "occultism": "int", "performance": "cha", "religion": "wis",
		"society": "int", "stealth": "dex", "survival": "wis", "thievery": "dex",
	}
	SkillNames = slices.Sorted(maps.Keys(skillAttributes))
)

func IsSkill(name string) bool {
	_, found := skillAttributes[name]
	return found
}

func IsSkillSelector(selector string) bool {
	return IsSkill(selector) || selector == SkillCheckSelector
}

func SkillAttribute(name string) string {
	return skillAttributes[name]
}

func (build *Build) skills(level int, attributes map[string]int, armor Armor) map[string]int {
	skills := make(map[string]int, len(SkillNames))
	for _, name := range SkillNames {
		attribute := skillAttributes[name]
		rank := max(slices.Index(RankWords, build.Skills[name]), 0)
		// deliberate: the higher of the rank's bonus and a proficiency modifier counts, not both, because bonuses of one type do not stack and the rank's own bonus is a proficiency bonus.
		proficiency := max(proficiencyBonus(level, rank), build.proficiencyModifier(level, name, SkillCheckSelector))
		modifier := proficiency + attributes[attribute] + build.total(level, 0, nil, name, SkillCheckSelector)
		// deliberate: the armor check penalty lands on Strength and Dexterity skills only, not on all skills, because that is the armor rule.
		if attribute == "str" || attribute == "dex" {
			modifier += armor.CheckPenalty
		}
		skills[name] = modifier
	}
	return skills
}

func (build *Build) proficiencyModifier(level int, selectors ...string) int {
	highest := 0
	for _, modifier := range build.Modifiers {
		if modifier.Type == ProficiencyModifier && slices.Contains(selectors, modifier.Selector) {
			highest = max(highest, modifier.Value+modifier.PerLevel*level)
		}
	}
	return highest
}

func (creature *Creature) checkSkills(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.Skills)) {
		if !IsSkill(name) {
			report("skill.%s: not a skill", name)
		}
	}
}
```

Lore skills are not combat statistics and stay out. A `proficiency`-type build modifier (Untrained Improvisation) is valid on a skill or `skill-check` only; a skill takes the higher of it and its rank's own bonus (Decision Record 43).

##### Task 2 — `internal/pf2e/sim/creature/creature.go`, `build.go`, `fight.go`, `validate.go`

`creature.go`: the creature holds its skill modifiers. An NPC's are typed or imported; a PC's are computed.

**Find:**
```go
	Pools                map[string]int                 `toml:"pool,omitempty"`
```
**Replace with:**
```go
	Skills               map[string]int                 `toml:"skill,omitempty"`
	Pools                map[string]int                 `toml:"pool,omitempty"`
```

`build.go`: the build holds a rank per skill, the armor its check penalty, and a build modifier may name a skill.

**Find:**
```go
	Proficiency            map[string]string   `toml:"proficiency,omitempty"`
```
**Replace with:**
```go
	Proficiency            map[string]string   `toml:"proficiency,omitempty"`
	Skills                 map[string]string   `toml:"skill,omitempty"`
```

**Find:**
```go
	SpeedPenalty int    `toml:"speed_penalty,omitzero"`
}
```
**Replace with:**
```go
	SpeedPenalty int    `toml:"speed_penalty,omitzero"`
	CheckPenalty int    `toml:"check_penalty,omitzero"`
}
```

**Find:**
```go
	modifierSelectors    = []string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage", "splash"}
```
**Replace with:**
```go
	modifierSelectors    = append([]string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage", "splash", SkillCheckSelector}, SkillNames...)
```

In `computedKeys`, add `{"skill"}` after `{"strike"}`. In `checkBuild`, after the loop over `build.Proficiency`:

```go
	for _, name := range slices.Sorted(maps.Keys(build.Skills)) {
		if !IsSkill(name) {
			report("build.skill.%s: not a skill", name)
		}
		if rank := build.Skills[name]; !slices.Contains(RankWords, rank) {
			report("build.skill.%s: rank %q: want %s", name, rank, strings.Join(RankWords, ", "))
		}
	}
```

In the modifier loop of `checkBuild`, the type check becomes:

```go
		switch {
		case modifier.Type == ProficiencyModifier && !IsSkillSelector(modifier.Selector):
			report("build.modifier.%s: type proficiency needs a skill or skill-check selector", id)
		case modifier.Type != ProficiencyModifier && !slices.Contains(modifierTypes, modifier.Type):
			report("build.modifier.%s: type %q: want item, status, circumstance, proficiency or none", id, modifier.Type)
		}
```

In `total`, a `proficiency` modifier is skipped (`modifier.Type == ProficiencyModifier ||` leads the `continue` condition), so it never reaches the typed bonus pool.

In the armor check of `checkBuild`:

**Find:**
```go
		if armor.SpeedPenalty > 0 {
			report("build.armor.speed_penalty must not be positive")
		}
```
**Replace with:**
```go
		if armor.SpeedPenalty > 0 || armor.CheckPenalty > 0 {
			report("build.armor: speed_penalty and check_penalty must not be positive")
		}
```

In `computeBuild`:

**Find:**
```go
	creature.Strikes = map[string]Strike{}
```
**Replace with:**
```go
	creature.Skills = build.skills(level, attributes, armor)
	creature.Strikes = map[string]Strike{}
```

`fight.go`: a fight-time modifier may name a skill, so a conditional bonus (Intimidating Prowess) can sit in a hand layer.

**Find:**
```go
	CheckSelectors   = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "melee-attack", "ranged-attack", "damage", "melee-damage", "spell-attack", "spell-dc", "class-dc"}
```
**Replace with:**
```go
	CheckSelectors   = append([]string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "melee-attack", "ranged-attack", "damage", "melee-damage", "spell-attack", "spell-dc", "class-dc", SkillCheckSelector}, SkillNames...)
```

`validate.go`, in `prepare`: add `creature.checkSkills(report)` after `creature.checkPools(report)`.

##### Task 3 — `internal/pf2e/sim/foundry/skills.go`, `pc.go`, `rules.go`

`internal/pf2e/sim/foundry/skills.go` (new) — full contents:

```go
package foundry

import (
	"fmt"
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const skillRankPrefix = "skill:"

type skillRank struct {
	Rank int `json:"rank"`
}

func (importer *pcImporter) skills(source map[string]skillRank) {
	for _, name := range slices.Sorted(maps.Keys(source)) {
		rank := source[name].Rank
		if !creature.IsSkill(name) || rank < 0 || rank >= len(creature.RankWords) {
			importer.report.skip(importer.source.Name, fmt.Sprintf("skill %s rank %d", name, rank), ReasonUnsupported)
			continue
		}
		importer.raise(skillRankPrefix+name, rank)
	}
}

// deliberate: the check penalty is dropped when the Strength threshold is met, not reduced, because that is the armor rule; the speed penalty is the one that is only reduced.
func (importer *pcImporter) checkPenalty() int {
	system := importer.armor
	attributes, err := importer.build.Attributes()
	if system.CheckPenalty >= 0 || (err == nil && attributes["str"] >= system.Strength) {
		return 0
	}
	return system.CheckPenalty
}
```

Added in execution (Decision Record 41, 44, 45), in the same file: the shapes `trainedSkills` (`value`, `lore`, `additional`, `custom`) and `customModifier` (`label`, `modifier`, `type`, `enabled`, `ignored`, `force`, `predicate`, `adjustments`, `traits`, `ability`, `damageType`, `damageCategory`, `critical`), and two methods:

- `trainedSkills(owner, trained)` raises each skill in `value` to at least trained through `raise`; an unknown name and a non-empty `custom` are reported as unsupported; each `lore` entry is reported as not a combat statistic.
- `customModifiers(source)` walks `system.customModifiers` by sorted selector. An entry on a skill or `skill-check` that is enabled, not ignored, with no predicate, a known bonus type (or `proficiency`), and no `force`, adjustments, traits, `ability`, `damageType`, `damageCategory` or `critical` becomes a build modifier named from its label. Otherwise: a selector that is not a skill is unsupported; disabled or ignored is a toggle, imported off; a predicate is conditional; anything else is unsupported.

`pc.go`: read the ranks, and write them and the check penalty into the build.

**Find:**
```go
type characterSystem struct {
	Details levelDetails   `json:"details"`
	Build   characterBuild `json:"build"`
}
```
**Replace with:**
```go
type characterSystem struct {
	Details         levelDetails                `json:"details"`
	Build           characterBuild              `json:"build"`
	Skills          map[string]skillRank        `json:"skills"`
	CustomModifiers map[string][]customModifier `json:"customModifiers"`
}
```

In `armorSystem`, add the field `CheckPenalty int` with the tag `json:"checkPenalty"` after `SpeedPenalty`. In the `creature.Build` literal of `importPC`, add `Skills: map[string]string{},`. `classSystem` and `backgroundSystem` gain `TrainedSkills trainedSkills` (tag `json:"trainedSkills"`); `class` and `background` each end with `importer.trainedSkills(current.Name, system.TrainedSkills)`.

**Find:**
```go
	for level, boosts := range system.Build.Attributes.Boosts {
		importer.build.Boosts["level-"+level] = boosts
	}
```
**Replace with:**
```go
	for level, boosts := range system.Build.Attributes.Boosts {
		importer.build.Boosts["level-"+level] = boosts
	}
	importer.skills(system.Skills)
	importer.customModifiers(system.CustomModifiers)
```

In `finish`: a rank whose key is a skill goes to the skill table.

**Find:**
```go
		if rank := importer.ranks[key]; rank > 0 {
			build.Proficiency[key] = creature.RankWords[rank]
		}
```
**Replace with:**
```go
		rank := importer.ranks[key]
		if rank == 0 {
			continue
		}
		if name, isSkill := strings.CutPrefix(key, skillRankPrefix); isSkill {
			build.Skills[name] = creature.RankWords[rank]
			continue
		}
		build.Proficiency[key] = creature.RankWords[rank]
```

In `finishArmor`, add `CheckPenalty: importer.checkPenalty(),` to the returned `creature.Armor` literal for worn armor.

`rules.go`: a rule element can raise a skill's rank or add a bonus to a skill.

In `rankKey`, add a case before the `saves.` case:

```go
	case strings.HasPrefix(inner, "skills."):
		name := strings.TrimPrefix(inner, "skills.")
		return skillRankPrefix + name, creature.IsSkill(name)
```

In `flatModifierFor`:

**Find:**
```go
	stat, isStat := statSelectors[selector]
	weaponStat, isWeaponStat := weaponStatSelectors[selector]
```
**Replace with:**
```go
	stat, isStat := statSelectors[selector]
	if creature.IsSkillSelector(selector) {
		stat, isStat = selector, true
	}
	weaponStat, isWeaponStat := weaponStatSelectors[selector]
```

After the `bonusTypes` lookup in `flatModifierFor`, `type = "proficiency"` is taken on a skill or `skill-check` selector only:

```go
	if kind == creature.ProficiencyModifier && creature.IsSkillSelector(selector) {
		bonusType, isBonusType = kind, true
	}
```

`formula.go` learns `match`, `when`, `lte` and `gte` (Decision Record 41): `lte` and `gte` give 1 or 0; `when(condition, value)` gives its value when the condition is not 0 and the new `errConditionFailed` otherwise; `match` skips each argument that fails that way and gives the first that holds, and it is an error when none does. `btwn` and the rest stay unknown.

Import format check (trap 4), for each Foundry shape this task reads:

| Shape | Handled | Reported |
|---|---|---|
| `system.skills.<name>.rank` | Ranks 0 to 4 of the 16 skills | An unknown skill name or a rank outside 0 to 4 |
| `ActiveEffectLike` on `system.skills.<name>.rank` | Mode `upgrade`, through the existing rank path | Any other mode, as today |
| `FlatModifier` on a skill or `skill-check` | Item, status, circumstance, untyped and `proficiency`, with a predicate that holds at baseline | A predicate that needs a target or an action (Intimidating Prowess), as conditional; an `ability` type, as not a combat statistic; `proficiency` on any other selector, as not a combat statistic, as today |
| `AdjustModifier` on a skill | Nothing | As not supported, as today (Intimidating Prowess's upgrade, Bands of Force) |
| `armor.checkPenalty` | Dropped at the Strength threshold | — |
| `trainedSkills.value` on a background or class item | Each skill raised to at least trained | An unknown skill name, as not supported |
| `trainedSkills.lore` | Nothing | Each lore, as not a combat statistic |
| `trainedSkills.additional` | Nothing to do: a count of free picks, which the stored ranks already hold | — |
| `trainedSkills.custom` | Nothing | A non-empty string, as not supported |
| `customModifiers.<selector>` list | An entry on a skill or `skill-check` (below) | Any entry on another selector, as not supported; an empty list is silent |
| `customModifiers` entry: `enabled`, `ignored` | Enabled and not ignored | Disabled or ignored, as a toggle imported off |
| `customModifiers` entry: `predicate` | Empty | Non-empty, as conditional |
| `customModifiers` entry: `type` | Untyped, item, status, circumstance, proficiency | Any other type (`ability`), as not supported |
| `customModifiers` entry: `force`, `adjustments`, `traits`, `ability`, `damageType`, `damageCategory`, `critical` | False, empty or null | Anything else, as not supported |
| `customModifiers` entry: `modifier` | A number or a formula the reader knows | A formula it does not know, as not supported |
| `customModifiers` entry: `label` | The modifier's ID (slugged) | — |
| `customModifiers` entry: `slug`, `kind`, `source`, `custom`, `notes`, `hideIfDisabled` | Not read: none changes the number (`kind` follows the sign of `modifier`) | — |

##### Task 4 — `cmd/gm-toolkit/pf2e/show.go`

`show` prints the skills on one line, after the class DC.

**Find:**
```go
	if shown.HeroPoints > 0 || shown.DyingMax > 0 || shown.RecoveryDC > 0 {
```
**Replace with:**
```go
	if len(shown.Skills) > 0 {
		entries := make([]string, 0, len(shown.Skills))
		for _, name := range slices.Sorted(maps.Keys(shown.Skills)) {
			entries = append(entries, fmt.Sprintf("%s %+d", name, shown.Skills[name]))
		}
		fmt.Fprintf(out, "Skills %s\n", strings.Join(entries, ", "))
	}
	if shown.HeroPoints > 0 || shown.DyingMax > 0 || shown.RecoveryDC > 0 {
```

##### Task 5 — `internal/pf2e/sim/creature/skills_test.go`

New, `package creature_test`:

- **`TestSkillsFromBuild`** — a level 14 build with Cha +4, `intimidation = "master"` and an item +2 build modifier on `intimidation` gives 14 + 6 + 4 + 2 = 26; an untrained skill is the attribute alone; two item bonuses on one skill do not stack; a `skill-check` modifier reaches every skill.
- **`TestArmorCheckPenalty`** — `check_penalty = -2` lowers athletics, acrobatics, stealth and thievery and no other skill (trap 3: the rule is wider than the one skill this fight uses).
- **`TestLoadRejectsBadSkills`** — an unknown skill in `[build.skill]`, a bad rank word, a positive `check_penalty`, an unknown key in an NPC's `[skill]`.
- **`TestSkillIsAComputedKey`** — a hand layer that sets `[skill]` on a creature with a build fails with `ErrComputedKey`.
- **`TestNPCSkillsAreTyped`** — an NPC layer with `[skill] athletics = 30` loads with that number.
- **`TestSkillsUseTheHigherProficiency`** (added, Decision Record 43) — a `skill-check` proficiency +14 gives an untrained skill 14 plus its attribute, leaves trained and master skills on their rank's bonus, and is not added twice; a skill's own proficiency +20 beats the `skill-check` one; a proficiency +3 below a trained rank changes nothing. `TestLoadRejectsBadSkills` also rejects `type = "proficiency"` on `will`.

##### Task 6 — `internal/pf2e/sim/foundry/skills_test.go` and `gate_test.go`

`skills_test.go` (new):

- **`TestImportPCSkills`** — Kas's import has `intimidation`, `athletics` and `acrobatics` at `master` and no entry for an untrained skill; Tavish has `crafting`, `deception` and `stealth` at `master`; Arabesque has `performance` and `acrobatics` at `expert`.
- **`TestImportSkillModifiers`** — Kas's Demon Mask becomes an item +2 build modifier on `intimidation`; Intimidating Prowess's `FlatModifier` is reported as conditional and its `AdjustModifier` as not supported (trap 4).
- **`TestImportCheckPenalty`** — an actor whose Strength meets the armor's threshold imports `check_penalty` 0; one below it imports the armor's number. Build the second case by lowering the threshold's opposite in a copy of the armor item in Go, not by a new testdata file. [As built: the test data has both cases, Kas (Strength +5, Full Plate 4) and Tavish (Strength +0, Studded Leather 1, −1); the Go copies test the boundary, a threshold equal to Strength (0) and one above it (−1).]
- **Added for Decision Record 41:** `TestImportPCSkillsReportsBadRanks`, `TestImportTrainedSkills` (Saboteur trains Tavish's Thievery; the Alchemist's and the Warrior's training lowers nothing; lore reported as not a combat statistic; an unknown name and a custom string reported as not supported), `TestImportProficiencyModifier` (the Untrained Improvisation formula imports +14 on `skill-check`; `proficiency` on `perception` and `ability` on `arcana` stay reported), `TestImportCustomModifiers` (Shadow Property +1 item on Stealth, and one case per row of the format table), `TestEvaluateMatchWhen` (the formula at levels 1, 5, 6, 7 and 14, and five shapes that must fail).

`gate_test.go`: add **`TestGateSkills`** beside the gate 1 test. It imports the three PCs from `testdata/actors`, loads each written layer, and compares all 16 skill modifiers with the fixture's Skills section, typed into the test as a table. A mismatch is a finding to put to Robert before any number is changed: the sheet may hold a bonus the importer reports as conditional.

##### Task 7 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`

Robert's edit, made before this commit and committed with it: a `## Skills` section with one table, a row per skill and a column per PC. [Done before Commit 10, as section 6 — Skills (Decision Record 41); nothing to add in this commit.]

##### Commit message

```
feat(E4): compute skills from the build and import their ranks

- A build holds a rank per skill; the load computes all 16 modifiers
- The armor check penalty lands on Strength and Dexterity skills
- The importer reads skill ranks, rank upgrades and skill bonuses
- An NPC's skills are typed in a [skill] table
- show prints the skills; a gate test checks them against the sheets
- The importer trains the skills a background or class item lists
- A proficiency bonus on skills (Untrained Improvisation) counts when it
  beats the rank's own; the formula reader learns match and when
- Hand-entered skill modifiers on the actor are imported; others reported
```

#### Commit 11 — `feat(E4): add the skill check step`

##### Task 1 — `internal/pf2e/sim/creature/steps.go`

A `check` step rolls one of the user's skills against a DC of the target, or a fixed DC, and runs the list for its degree.

**Find:**
```go
	StepGrant       = "grant"
```
**Replace with:**
```go
	StepGrant       = "grant"
	StepCheck       = "check"
```

Add `StepCheck` to `stepKinds` and to `rolledSteps`. Add `checkDefenses = []string{"fortitude", "reflex", "will", "perception"}` to the `var` block. Add the fields `Check string` (tag `toml:"check,omitempty"`) and `Against string` (tag `toml:"against,omitempty"`) after `Grant`. In `heads`, before `return heads`:

```go
	if step.Check != "" {
		heads = append(heads, StepCheck)
	}
```

In `checkStep`, add a case:

```go
	case StepCheck:
		if _, found := creature.Skills[step.Check]; !IsSkill(step.Check) || !found {
			report("%s: check %q: want a skill the creature has a number for", path, step.Check)
		}
		if (step.Against == "") == (step.DC == 0) || (step.Against != "" && !slices.Contains(checkDefenses, step.Against)) {
			report("%s: a check needs one of dc and against (%s)", path, strings.Join(checkDefenses, ", "))
		}
```

**Find:**
```go
	if kind != StepSave && (step.DC != 0 || step.DCFrom != "" || step.Basic || step.Shift != 0) {
		report("%s: dc, dc_from, basic and shift belong on a save", path)
	}
```
**Replace with:**
```go
	if kind != StepSave && (step.DCFrom != "" || step.Basic || step.Shift != 0) {
		report("%s: dc_from, basic and shift belong on a save", path)
	}
	if kind != StepSave && kind != StepCheck && step.DC != 0 {
		report("%s: dc belongs on a save or a check", path)
	}
	if kind != StepCheck && step.Against != "" {
		report("%s: against belongs on a check", path)
	}
```

The old message is split; update the case in `steps_test.go` that reads it. [As built: no update needed; the case reads only "belong on a save", which the new message still holds (Decision Record 48).]

Producer check (trap 2): `creature.Skills` is filled by `computeBuild` for every PC with a build, all 16 skills, before `prepare` runs; an NPC or a PC with no build must type the skill, and the load says so when it is missing.

##### Task 2 — `internal/pf2e/sim/engine/events.go`

Add after `SaveRolled`:

```go
type SkillChecked struct {
	Combatant string
	Target    string
	Ability   string
	Skill     string
	Die       int
	Bonus     int
	Total     int
	DC        int
	Degree    Degree
}
```

Add `func (SkillChecked) isEvent() {}` to the `isEvent` block.

##### Task 3 — `internal/pf2e/sim/engine/checks.go`, `conditions.go`, `steps.go`

`checks.go`, add at the end:

```go
func (fight *Fight) defenseDC(target *Combatant, against string, traits []string) int {
	if against == "perception" {
		return 10 + target.Creature.Perception + fight.modifierTotal(target, []string{"perception"}, traits)
	}
	return 10 + saveBonus(target.Creature.Defense, against) + fight.modifierTotal(target, saveSelectors[against], traits)
}

func (fight *Fight) skillCheck(source effectSource, step creature.Step, target *Combatant) Degree {
	actor := source.actor
	selectors := []string{step.Check, creature.SkillCheckSelector, creature.SkillAttribute(step.Check) + "-skill"}
	bonus := actor.Creature.Skills[step.Check] + fight.modifierPool(actor, selectors, nil, target).Total()
	dc := step.DC
	if step.Against != "" {
		dc = fight.defenseDC(target, step.Against, source.traits)
	}
	die := fight.d20(actor, creature.RollSkill)
	total := die + bonus
	degree := DegreeOf(die, total, dc)
	fight.observer.Observe(SkillChecked{
		Combatant: actor.ID,
		Target:    target.ID,
		Ability:   source.ability,
		Skill:     step.Check,
		Die:       die,
		Bonus:     bonus,
		Total:     total,
		DC:        dc,
		Degree:    degree,
	})
	return degree
}
```

A defense DC is 10 plus the save, with the target's modifiers against the ability's traits and its condition penalties, so a frightened target is easier to demoralize again. The saver's degree shifts do not apply: the target rolls nothing.

`conditions.go`: the conditions that lower checks lower skill checks too.

**Find:**
```go
var allChecksAndDCs =[]string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "spell-attack", "spell-dc", "class-dc"}
```
**Replace with:**
```go
var allChecksAndDCs = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "spell-attack", "spell-dc", "class-dc", "skill-check"}
```

**Find:**
```go
	"clumsy":     {selectors: []string{"ac", "reflex", "ranged-attack"}, kind: "status"},
```
**Replace with:**
```go
	"clumsy":     {selectors: []string{"ac", "reflex", "ranged-attack", "dex-skill"}, kind: "status"},
```

**Find:**
```go
	"enfeebled":  {selectors: []string{"melee-attack", "melee-damage"}, kind: "status"},
```
**Replace with:**
```go
	"enfeebled":  {selectors: []string{"melee-attack", "melee-damage", "str-skill"}, kind: "status"},
```

**Find:**
```go
	"stupefied":  {selectors: []string{"will", "perception", "spell-attack", "spell-dc"}, kind: "status"},
```
**Replace with:**
```go
	"stupefied":  {selectors: []string{"will", "perception", "spell-attack", "spell-dc", "int-skill", "wis-skill", "cha-skill"}, kind: "status"},
```

`steps.go`, in `runStep`, add a case to the kind switch at the bottom:

```go
	case creature.StepCheck:
		for _, target := range targets {
			if target.Standing() {
				fight.runSteps(source, step.For(int(fight.skillCheck(source, step, target)), false), target)
			}
		}
```

##### Task 4 — `testdata/scope/kas/hand.toml`

Add, with the `[skill]` and `[modifier.*]` tables before the first `[strike.*]`:

```toml
[skill]
intimidation = 26

# Intimidating Prowess: +2 circumstance to Demoralize at Strength +5 and master Intimidation.
[modifier.intimidating-prowess]
selector = "intimidation"
type     = "circumstance"
value    = 2

# Temporary immunity to a repeat Demoralize comes with Effort 5.
[ability.demoralize]
name    = "Demoralize"
actions = 1
traits  = ["concentrate", "emotion", "fear", "mental", "visual"]
steps   = [
  { check = "intimidation", against = "will",
    success = [{ condition = "frightened", value = 1 }],
    critical_success = [{ condition = "frightened", value = 2 }] },
]

[ability.battle-cry]
name    = "Battle Cry"
traits  = ["concentrate", "emotion", "fear", "mental", "visual"]
trigger = { on = "initiative", cost = "free" }
steps   = [
  { check = "intimidation", against = "will",
    success = [{ condition = "frightened", value = 1 }],
    critical_success = [{ condition = "frightened", value = 2 }] },
]

[ability.too-angry-to-die]
name    = "Too Angry to Die"
traits  = ["concentrate", "emotion", "fear", "mental", "visual"]
trigger = { on = "stood-up", cost = "reaction" }
steps   = [
  { check = "intimidation", against = "will",
    success = [{ condition = "frightened", value = 1 }],
    critical_success = [{ condition = "frightened", value = 2 }] },
]
```

Intimidating Glare makes Demoralize visual, not auditory, so the ability carries `visual` and no `auditory` trait (Decision Record 51).

##### Task 5 — tests

`internal/pf2e/sim/engine/skills_test.go` (new):

- **`TestSkillCheckAgainstADefense`** — the DC is 10 plus the target's Will; the four degrees run their own lists, and a critical success with no list runs nothing (a check is not an attack roll); `SkillChecked` carries the numbers.
- **`TestSkillCheckFixedDC`** — `dc = 20` ignores the target's saves.
- **`TestSkillCheckModifiers`** — a fight-time circumstance bonus on the skill adds; frightened 2 on the user lowers the roll by 2; frightened 1 on the target lowers its DC by 1; enfeebled lowers athletics and not intimidation; stupefied lowers intimidation and not athletics; clumsy lowers stealth (trap 3: three attribute groups, one of which this fight uses).
- **`TestSkillCheckTakesMisfortune`** — a `skill` reroll applies; an `attack` reroll does not.
- **`TestSkillCheckSkipsImmune`** — a target immune to `fear-effects` is skipped by a fear ability, with `ImmuneToEffect` and no roll.
- **`TestTargetModifierAgainstTraits`** — a target with +3 to Will against `fear` has a DC 3 higher against a fear ability.

`creature/steps_test.go`: bad cases for a check with neither `dc` nor `against`, with both, `against = "ac"`, a skill the creature has no number for, `against` on a save, `dc` on a damage step. [As built: `TestLoadRejectsBadChecks` has these plus a name that is not a skill, a PC with no build and no `[skill]` (trap 2) and `dc_from` on a check; `TestLoadReadsCheckSteps` loads two check steps on a PC with a build.]

`scope_test.go`, new rows:

- **Demoralize, with Intimidating Glare and Intimidating Prowess** — Kas rolls d20 + 26 + 2 against the Avatar's Will DC 42: a success is frightened 1, a critical success frightened 2, and her frightened value then lowers her AC against his next Strike.
- **Battle Cry** — in a `Run`, after the last `InitiativeRolled` and before round 1, Kas demoralizes the first foe as a free action; his reaction is unspent.
- **Too Angry to Die** — Kas is dropped, healed, and Stands on his turn: the decider reacts and he demoralizes; a Stand from an ordinary prone offers nothing; with his reaction spent on Farabellus Flip it offers nothing. [As built: the heal is Kas's own fast healing at the start of his turn (Decision Record 22, 49).]

##### Commit message

```
feat(E4): add the skill check step

- A check step rolls a skill against a target's save or Perception DC,
  or a fixed DC, and runs the list for its degree
- Frightened and sickened lower skill checks; enfeebled, clumsy and
  stupefied lower the skills of their attributes
- Scope rows: Demoralize, Battle Cry, Too Angry to Die
```

### Phase 6 — Data and wiki

**In plain words:** the importer picks up the two new numbers an NPC export holds, the wiki describes everything this effort built, and the E4 abilities move from the test folders into the real creatures' hand layers. When this lands, Effort 4 is ready for its close.

#### Commit 12 — `feat(E4): import fast healing and NPC skills`

##### Task 1 — `internal/pf2e/sim/foundry/npc.go`

An NPC's skills come across as stored, and a `FastHealing` rule element on an ability becomes `fast_healing`.

**Find:**
```go
type npcSystem struct {
	Details    levelDetails  `json:"details"`
	Attributes npcAttributes `json:"attributes"`
	Perception npcPerception `json:"perception"`
	Saves      npcSaves      `json:"saves"`
	Traits     itemTraits    `json:"traits"`
}
```
**Replace with:**
```go
type npcSystem struct {
	Details    levelDetails        `json:"details"`
	Attributes npcAttributes       `json:"attributes"`
	Perception npcPerception       `json:"perception"`
	Saves      npcSaves            `json:"saves"`
	Traits     itemTraits          `json:"traits"`
	Skills     map[string]npcSkill `json:"skills"`
}

type npcSkill struct {
	Base int `json:"base"`
}
```

**Find:**
```go
	layer.Defense.Resist = npcAdjustments(attributes.Resistances, "resistance", source.Name, report)
```
**Replace with:**
```go
	for _, name := range slices.Sorted(maps.Keys(system.Skills)) {
		if !creature.IsSkill(name) {
			report.skip(source.Name, "skill "+name, ReasonUnsupported)
			continue
		}
		if layer.Skills == nil {
			layer.Skills = map[string]int{}
		}
		layer.Skills[name] = system.Skills[name].Base
	}
	layer.Defense.Resist = npcAdjustments(attributes.Resistances, "resistance", source.Name, report)
```

**Find:**
```go
		case "action":
			report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
```
**Replace with:**
```go
		case "action":
			if amount, found := fastHealing(current); found {
				layer.FastHealing = max(layer.FastHealing, amount)
				break
			}
			report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
```

Add at the end of the file, and `maps` to the imports:

```go
// deliberate: only a plain number is read, not a formula, because an NPC's fast healing is a fixed value; anything else stays in the report as an ability to write by hand.
func fastHealing(current *item) (int, bool) {
	for _, rule := range current.System.Rules {
		if rule["key"] != "FastHealing" {
			continue
		}
		if kind, isText := rule["type"].(string); isText && kind != "fast-healing" {
			continue
		}
		if predicate, isList := rule["predicate"].([]any); isList && len(predicate) > 0 {
			continue
		}
		if value, isNumber := rule["value"].(float64); isNumber && value > 0 {
			return int(value), true
		}
	}
	return 0, false
}
```

Import format check (trap 4): a `FastHealing` rule has `value`, and optionally `type` (`fast-healing` or `regeneration`), `deactivatedBy` and a predicate. Regeneration changes how the creature dies and is not fast healing, so a rule with `type` set to `regeneration`, or with a predicate, must not be read as plain fast healing. The two guards in the loop above do this: a rule whose `type` is a string other than `fast-healing`, or whose `predicate` is a non-empty list, is skipped, and the item stays in the report as an ability. An NPC skill entry in both exports holds `base` only; the importer reads `base`.

`cmd/gm-toolkit/pf2e/show.go`, in `printCreature`, after the adaptive resistance line (Decision Record 57):

```go
	if shown.FastHealing > 0 {
		fmt.Fprintf(out, "Fast healing %d\n", shown.FastHealing)
	}
```

##### Task 2 — `internal/pf2e/sim/foundry/report.go`, `rules.go`, `pc.go`, `actor.go`

The reasons follow overview Decision 13: abilities are written by hand in Effort 6, not converted.

**Find:**
```go
	ReasonAbility     = "ability not converted yet (Effort 6)"
```
**Replace with:**
```go
	ReasonAbility     = "ability written by hand (Effort 6)"
	ReasonAbilityPart = "part of an ability written by hand (Effort 6)"
```

`rules.go`: `DamageDice` rules belong to Vicious Swing, Brutal Finish and the relics, which are abilities now.

**Find:**
```go
	laterRuleKeys  = []string{"DamageDice", "Strike", "Weakness", "Immunity"}
```
**Replace with:**
```go
	laterRuleKeys  = []string{"Strike", "Weakness", "Immunity"}
```

In `applyRule`, add a case before the `laterRuleKeys` case:

```go
	case key == "DamageDice":
		importer.report.skip(current.Name, key, ReasonAbilityPart)
```

`pc.go`, in `weapon`: a consumable weapon's count is typed as a pool, and a bomb with no stock that the PC can brew is imported (Decision Record 42).

`report.go` gains `ReasonNoFormula = "quantity 0 and no known formula to brew it"`. `actor.go`: `item` gains `Stats itemStats` (`_stats`, with `CompendiumSource`) and `Flags itemFlags` (`flags.core`, with `SourceID` from `sourceId`). `pc.go`: `characterSystem` gains `Crafting crafting` (`crafting.formulas`, each with `UUID` from `uuid`); `pcImporter` gains `quickAlchemy bool` and `formulas map[string]bool`, filled in `importPC` before the item loop, from the formula book and from any item whose slug is `quick-alchemy` (`quickAlchemySlug`).

In `weapon`, the quantity check becomes:

```go
	brewed := false
	if current.System.Quantity < 1 {
		if brewed = importer.brewable(current); !brewed {
			return nil
		}
	}
```

and the consumable line becomes:

```go
	switch {
	case brewed:
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, no stock: brewed with Quick Alchemy from the quick bombs pool", current.System.Quantity), ReasonHandLayer)
	case slices.Contains(weapon.Traits, "consumable"):
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, typed as a pool", current.System.Quantity), ReasonHandLayer)
	}
```

with

```go
func (importer *pcImporter) brewable(current *item) bool {
	if !slices.Contains(current.System.Traits.Value, "bomb") || !importer.quickAlchemy {
		importer.report.skip(current.Name, "weapon", ReasonQuantity)
		return false
	}
	origin := cmp.Or(current.Stats.CompendiumSource, current.Flags.Core.SourceID)
	switch {
	case origin == "":
		importer.report.skip(current.Name, "bomb with no compendium origin to match a formula", ReasonNoFormula)
		return false
	case !importer.formulas[origin]:
		importer.report.skip(current.Name, "bomb whose formula is not in the formula book", ReasonNoFormula)
		return false
	}
	return true
}
```

Import format check (trap 4): a formula entry in `system.crafting.formulas` holds `uuid` and `deletable` in both exports. `uuid` is the compendium item the formula makes, and is read; `deletable` says only whether the sheet lets the entry be removed, and is not read. The origin is `_stats.compendiumSource`; `_stats.duplicateSource` (a copy of a world item, not a compendium origin) and the other `_stats` fields are not read. `flags.core.sourceId` is the older place for the origin and is read when `compendiumSource` is empty; in our data `flags.core` is present and empty on some items. Quick Alchemy is two items on Tavish, an `action` and a class-feature `feat`, both with the slug `quick-alchemy`; either one counts.

##### Task 3 — tests

`internal/pf2e/sim/foundry/npc_test.go`:

- **`TestImportNPCSkills`** — Abyssal Luna imports athletics 30 and occultism 33; her two lore items stay in the report as not combat statistics.
- **`TestImportNPCFastHealing`** — the Avatar of Lilith imports `fast_healing = 10`, and Fast Healing leaves the report; an action item with no such rule is still reported as an ability; a rule with `type = "regeneration"` and one with a predicate are not read (trap 4), built in Go from a copy of the item.

`internal/pf2e/sim/foundry/pc_test.go`:

- **`TestImportBrewedBombs`** — through `importPC` on the frozen Tavish with Nail Bomb (Greater) set to quantity 0: a known formula imports it with `cold-iron` and reports it as brewed; the origin moved to `flags.core.sourceId` imports it too; the formula removed, or no origin, skips it with `ReasonNoFormula`; Quick Alchemy removed skips it with `ReasonQuantity` (trap 1, trap 4).
- **`TestImportQuantityZeroNonBomb`** — the Dagger at quantity 0 is skipped with `ReasonQuantity`.

`cmd/gm-toolkit/pf2e/show_test.go`: `TestPrintCreatureShowsComputedNumbers` checks the fast healing line.

Update the tests that count skips by reason: `DamageDice` rows move to `ReasonAbilityPart`, and consumable weapons to `ReasonHandLayer` (`TestImportPCWeaponsAndQuantities`, `TestImportBombDamage` and the rule-element tests for Kas and Tavish). `TestImportPCWeaponsAndQuantities` counts 14 weapons: Frost Vial (Moderate) and Bottled Lightning (Moderate) are brewed (Decision Record 42).

##### Task 4 — the import, by hand

Build the binary into the scratchpad and run `gm-toolkit pf2e sim import fall-of-luna --campaign hawk-claude`. Check in the report: the Avatar shows fast healing 10 and no Fast Healing skip; both bosses and all three PCs show a `Skills` line; Kas's intimidation matches the fixture; Tavish has a Strike for Nail Bomb (Greater) with `cold-iron`, reported as brewed. Robert runs the same import on `hawk`; the rewritten `import.toml` files are committed with this commit.

##### Commit message

```
feat(E4): import fast healing and NPC skills

- An NPC's skills and its FastHealing rule element are imported
- show and the import report print fast healing
- Abilities and DamageDice rules are reported as written by hand
- A consumable weapon's quantity is reported as a pool to type
- A bomb with no stock is imported when the PC has Quick Alchemy
  and its formula
- The Fall of Luna import layers are rewritten
```

#### Commit 13 — `docs(E4): document triggers and limits and write the E4 abilities`

Load `/gestalt:style-guide` before these tasks. Every claim on a wiki page is checked against the source as it stands after Commit 12, not against this plan.

##### Task 1 — `docs/architecture/pf2e/engine.md`

- **Purpose** — Effort 4 is built; Effort 5 adds spells and tables.
- **Setup and Run** — the combatant's state list gains uses, pools, recharges, rerolls and granted effects.
- **The turn** — rewrite: the action count from quickened, slowed and stunned, and the Strike-only action; limits tick and the reaction returns for every living creature; turn-start triggers, then fast healing; Stand and Sustain as choices; a free ability and a cost range; the last paragraph about the unused reaction slot goes.
- **New sections**, after Strikes: **Limits** (the four limits, pools, `flourish` and `press`, when each resets, `Blocked` and the three queries); **Triggers and reactions** (a table of the ten kinds with when each fires and its qualifiers; the three costs; one answer per event; extra reactions; `provoke` and disrupt; what a triggered ability aims at); **Fortune and misfortune** (`d20`, which rolls go through it, cancel); **Allies and areas** (`targets`, `areaTargets`, revives); **Acts of loyalty** (the three sources, once per round, granted traits); **Skill checks** (the roll, the DC, the selectors); **Sustain**.
- **Abilities and steps** — the kinds table grows from 8 to 15: `gain_actions`, `recharge`, `reroll`, `bonus`, `recovery_dc`, `grant`, `check`; the Strike step's tweaks are keys on the Strike step, not a kind (Decision Record — Execution 65). Note that a Strike step's own lists run after the riders, and that an ability stops when its user falls.
- **Strikes and the multiple attack penalty** — the penalty applies on the attacker's own turn only; `counts` and `penalty = "shared"`.
- **Conditions** — four new rows (deafened, quickened, slowed, stunned); the skill selectors on clumsy, enfeebled and stupefied; replace the "wait for Effort 4" sentence with: confused arrives with Effort 5, immobilized with position.
- **Dying and death** — a knocked-out PC is prone; the eased recovery DC.
- **The decider** — the three-method interface, `ChooseStand`, `ChooseSustain`, `Choice.Actions`, and what `FirstFitting` does with each.
- **Events** — a row for every event this effort adds; `TriggeredBy` on `DamageDealt` and `Healed`.
- **Not modelled yet** — drop the Effort 4 bullet; position is `A.004.2`'s, with this effort's stand-ins listed (Decision 13); add hands and the unconscious penalties as not modelled by choice.
- **Key Decisions** — add: triggered abilities are ordinary abilities with a `trigger` table (Decision 20); reactions are decisions behind `React` (Decision 21); the multiple attack penalty is per own turn (Decision 24); events are added, never changed, and name their cause (Decision 26); every "all engaged" rule sits behind `areaTargets`, `provoke` or a trigger's `traits`.

##### Task 2 — `docs/architecture/pf2e/data-files.md`

- **The file format** — the top-level keys gain `fast_healing`; new tables `[skill]`, `[pool]` and `[extra_reaction]`; the ability gains `max_actions`, `free`, `turn_ends`, `tags`, `targets`, `limit`, `trigger` and `sustain`; a damage part gains `per_action`. Extend the step table with the new kinds and keys, and add a second example: one triggered ability with a limit, loaded through `creature.Load` in a throwaway test before it goes in the page.
- **What the load checks** — the new checks, grouped as the page does today.
- **PC builds** — `[build.skill]`, `check_penalty`, the skill computation, and `skill` as a computed key.
- **Import** — NPC skills and fast healing; PC skill ranks, rank upgrades and skill bonuses; the two changed report reasons.
- **Key Decisions** — add: limits are one table and pools are shared counts (Decision 19); skills are computed from the build (Decision 16).

##### Task 3 — `docs/architecture/pf2e/overview.md`

Only the lines this effort makes false: where the page lists what the engine does, add triggers, limits and skills. The stale "Effort 7's tactics" and "scores the results" lines stay for Effort 8 (hand-off, Known stale lines).

##### Task 4 — `dev_journals/campaigns/initiatives/implementation/F.041-pf2e-encounter-simulator-rules-scope.md`

- **Section 2** — reword four stand-ins to Decisions 7, 10, 13 and 15: Walk the Timelines (no move), Sin Vulnerability (tags, revives and passed saves; the acting PC; no step-between), Quick Alchemy (6 quick bombs, a dial; each additive its own ability), Comforting Presence (a reaction, the triggering ally). Remove the Too Angry to Die row. Add to the named approximations: Attack of Opportunity (every foe in reach), Searing Wave (the cone is a save by the struck foe), Healing Wave (the whole party), Ring of the Ram, Farabellus Flip and Lash Out (no movement).
- **Section 3** — Too Angry to Die becomes "as written"; Demoralize's last rule is E4; Bellflower's is E5 (split the Bewitching Bloom row); Return Force's is `A.004.2`; Wail's model names the spell's degrees.
- **Section 6** — close the two open calls and add a row for each of Decisions 1 to 18, with E4 as the effort.

##### Task 5 — the hand layers

Copy the ability, pool, modifier and extra-reaction tables from `internal/pf2e/sim/engine/testdata/scope/` into the real hand layers, leaving the typed stat blocks behind:

| Scope file | Real layer | Besides the abilities |
|---|---|---|
| `abyssal-luna` | `campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml` | The `critical_success` lists of `[strike.slash]` and `[strike.stab]` only; the import owns the rest of each Strike |
| `avatar-of-lilith` | `campaigns/hawk/encounters/fall-of-luna/avatar-of-lilith/hand.toml` (new) | Not `fast_healing`; the import writes it |
| `kas` | `campaigns/hawk/pf2e/party/kas/hand.toml` | `[pool]`, `[extra_reaction]`, `[modifier.intimidating-prowess]`; not `[skill]`, which his build computes |
| `tavish` | `campaigns/hawk/pf2e/party/tavish/hand.toml` | `[pool]` |
| `arabesque` | `campaigns/hawk/pf2e/party/arabesque/hand.toml` | `[pool]` |

Each stand-in and approximation keeps its comment line. Then load all five real folders through `creature.Load` in a throwaway test, fix what it reports (a Strike ID that differs from the scope file is the likely case), and delete the test.

##### Task 6 — `F.041-pf2e-encounter-simulator-overview-plan.md` and this plan

Tick Effort 4's Done box in the overview's Effort Summary and the Phase 6 box in this plan's Overview. In the overview's Findings, add one bullet: Demoralize was listed under Effort 3 and built in Effort 4, with the skill computation (Decision 16).

##### Commit message

```
docs(E4): document triggers and limits and write the E4 abilities

- engine.md: limits, triggers and reactions, fortune and misfortune,
  allies, acts of loyalty, skill checks, Sustain, the new events
- data-files.md: the ability's new keys, skills, pools, the import changes
- Rules scope: this effort's rules calls, stand-ins and approximations
- The E4 abilities move from the scope test folders to the five hand layers
- Effort 4 ticked in the overview plan
```
