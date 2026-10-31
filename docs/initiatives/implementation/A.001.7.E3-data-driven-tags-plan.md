# Data-Driven Tags (A.001.7) — Effort 3 Plan: Contest Rolls and Reactions

## In Plain Words

- **What:** five tag rules start to work, and three asset contests start to use the
  same dice rules as every other roll.
  - **Contests.** Tripwire Cells, Panopticon Matrix and Lobbyists roll bare dice
    today, so no tag or asset bonus reaches them. After this effort they roll
    through the hooked roll path, with no one asked: free extra dice are taken,
    and limited or optional ones wait for the roll card (`R.025`).
  - **Perimeter Agency** gets its extra die on every roll that tests for stealth.
  - **Psychic Academy** gets its engine part: a reroll forced on any rival's roll,
    even one the holder is not in. Its data entry waits for the roll card (D-7).
  - **Deep Rooted** is lost when its faction changes homeworld; **Exchange
    Consulate** rolls 1d6 for a bonus XP when it completes Peaceable Kingdom.
- **Why:** these rules do nothing in play today. The contests skip the roll hooks,
  and goal events never reach the rule engine's reactors.
- **How it fits:** Effort 3 of five in `A.001.7` (arc `A.001`), on branch
  `feature/data-driven-tags`. It builds on Efforts 1 and 2. Effort 4 (world and
  setting rules) and Effort 5 (A-side tag holder) follow; neither needs this one.

## Context / Goal

- Overview: [`A.001.7-data-driven-tags-overview-plan.md`](./A.001.7-data-driven-tags-overview-plan.md).
  Decisions 2 (contests with no collector), 6 (Perimeter Agency's three rolls) and
  7 (values in the entry) are built here. R-5 (this session) splits the old
  Effort 3 in two.
- Discovery: [`A.001.7-data-driven-tags-discovery.md`](../discovery/A.001.7-data-driven-tags-discovery.md),
  D-7, D-16, D-17, D-18, D-19.
- Effort 2: [`-effort-2-plan.md`](./A.001.7-data-driven-tags-effort-2-plan.md).
  Decision 3 there promised that the contests set `OpponentAttribute`.
- Architecture read: hooks, effect-mutation.

## Decisions Ratified in Planning

1. **The old Effort 3 splits at the roll/world seam.** Robert's call (overview
   R-5). This effort holds the roll path and the reactors. `settings.toml`, the
   Starship rule, Colonists, Technical Expertise and Pirates move to a new Effort 4
   with its own Plan session; the A-side holder becomes Effort 5. The two halves
   share no new engine part.

2. **The no-prompt rule lives in `RollWithHooks`.** Planning call (builds overview
   Decision 2). With a nil collector, Cat 1 takes every offer whose `BudgetKey` is
   empty (`freeOffers`) and skips the rest, and Cat 2 declines every elective
   reroll; forced rerolls apply. The offer filters and budget filter run first, as
   today. No caller changes until a contest passes nil.

3. **The three contests roll through `RollWithHooks`; each holder carries what it
   needs.** Robert's call (the rulebook route); planning call (the rest).
   - `contestPasses` takes one `contestRoll` value and rolls each side as a hooked
     roll: 1d10 plus the stat, the other side's stat as `OpponentAttribute`
     (Effort 2 Decision 3), both assets, the world.
   - The reactor (Tripwire Cells) and the contester (Lobbyists) keep the registry
     from `RegisterContext`, as a reactor keeps its roller. The registry is
     cleared in place each cycle, so the pointer stays good.
   - `FireContext` gains `Registry` and `Rulebook`; `FireTurnStart` and
     `FirePurchase` take the registry (Panopticon Matrix's turn-start route).
   - `WinsContest(query, factionState, rulebook)`: without state no roll hook
     passes its holder check. `RequestPermission` and `RequestArrivalPermission`
     take the rulebook, and their five callers pass it. Rejected: the rulebook held
     at registration (a Cat 4 hook holds no rulebook); a nil rulebook (tech-level
     filters would silently match nothing).
   - Each call site passes `collector: nil` under a fence-sign that names the
     roll-card follow-on row (overview Decision 2).

4. **Contest phases follow the book's words, carried by the entry.** Robert's call
   (reading); planning call (the key). Tripwire Cells is "a Cunning vs. Cunning
   attack": attack and defense. Panopticon Matrix and Lobbyists are "tests":
   faction test and contested, as the A-side gate rolls. One registrar
   (`clear_stealth`) serves both Tripwire and Panopticon, so the phase cannot live
   in Go (overview Decision 7). The contest gate gains a required key
   `roll = "attack" | "test"`; the loader refuses an entry without it.
   - What reaches Tripwire under attack/defense with no prompt: Fanatical's 1s,
     Deep Rooted's die when the target defends on its homeworld, and Panopticon's
     own Cunning die when its owner's Cells share its world. Every once-per-turn
     die waits.
   - An `attack` contest's tie asks the tie resolvers, so Fanatical's `lose_ties`
     reaches Tripwire too (Decision Record 2).
   - **Imperialists' die keeps to the Attack action.** Robert's call (rules
     reading). Effort 2's `goal` filter matches any attack on a rival on the
     Planetary Seizure target world, so Tripwire's attack would get the free die.
     The book says "attacks made as part of a Seize Planet action", so the entry
     gains `action = "Attack"`. The Expand Influence free attack loses it too.
   - Panopticon's own die does not reach its own test.

5. **A detection roll is marked on the detecting side only.** Robert's call (the
   flag); planning call (the side). `RollContext.Detection` is set on the roll of
   the side trying to detect: the actor of the reveal ability, and the declaring
   side of `clear_stealth` (Tripwire's attack, Panopticon's defense). The side
   resisting detection never sets it, so a Perimeter Agency faction that is itself
   the target of a reveal gets no die. Lobbyists sets nothing. `bonus_die` accepts
   `detection = true`. The die has no limit, so the free-die rule takes it in both
   contests (overview Decision 6).

6. **Goal mutations reach the reactors.** Robert's call. Two new passes of
   `dispatch.MutationReactors`: over the goal-lock mutations before they apply,
   and over the `UpdateProgress` output alone after the action batch's own pass.
   The second pass reads only the goal mutations, so no reactor sees the action
   batch twice. Rejected: `goal_completed` and `homeworld_changed` as phase
   triggers fired directly (two new routes, and goal events stay hidden from
   every other reactor).

7. **Psychic Academy's reroll is `rival_roll` on `reroll`, read from every
   faction's bucket.** Robert's call (the reach); planning call (the rest).
   - `rollScopes` adds each bystander faction's faction bucket after both sides',
     in ID order. Only tags register at faction scope, and every tag roll hook
     checks its owner in `rollMatches`, so no existing hook changes behaviour.
   - `rival_roll` matches any roll whose actor is not the owner, anywhere. It
     takes scope `faction` and no filters; a tag holds it only (`Validate`),
     because a bystander's asset buckets are never read.
   - The forced reroll picks the rival's highest die, as `enemy_roll_at_me` does,
     and is elective with a once-per-turn gate, so the free-die rule keeps it out
     of the contests (overview Decision 2). No `tags.toml` entry (D-7).

8. **Deep Rooted's loss is `remove_tag` on `homeworld_changed`.** Planning call
   (D-19). A reactor on the first `HomeworldChanged` of its owner in the batch
   emits `TagRemoved` for the tag that declared it, cause `remove_tag`. No
   parameters. Only a tag holds it: an asset has no tag of its own.

9. **Exchange Consulate's XP is `xp_roll` on `goal_completed`.** Planning call
   (D-18, overview Decision 7). The entry carries the die, the target and the XP
   (`die = "1d6"`, `target = 4`, `amount = 1`), and the `goal` filter names the
   completed goal, checked at load by the existing goal-reference check. On a
   goal trigger `goal` reads the completed goal; on a roll trigger it still reads
   the active goal's target world (Effort 2 Decision 5). Emits `XPAwarded`, cause
   `xp_roll`. Either holder may use it.

10. **Four phases.** Planning call. Phase 1: the no-prompt rule, and a refactor
    that carries the `roll` key, registry, state and rulebook to the contest call
    sites with no behaviour change. Phase 2: the contests on the roll path. The
    split runs at the cross-package seam, because one phase for both ran about
    1,090 lines against the 800 budget. Phase 3: the two roll features that need
    the contests (detection, `rival_roll`). Phase 4: the goal-event reactors and
    the two tags that use them.

## Decision Record — Execution

1. **Phase 4, Commit 7: the digest's goal line leaves out Exchange Consulate's
   bonus XP. Deferred, not missed.** `digest/build.go` adds every `XPAwarded` to
   the beat's `XPGained`, so the total is right. But it folds an `XPAwarded` into
   the goal line only when `GoalCompleted.XPAwarded` is zero, and `completeGoal`
   sets it to 1. So the line reads "completed Peaceable Kingdom (+1 XP)" when the
   faction gained 2. The TUI turn list shows only the `GoalCompleted` line too.
   For the effort-close review to decide: fold the `xp_roll` award into the goal
   line, show it as its own beat, or leave it. Settled in Decision Record 4.

2. **Effort close: an `attack` contest's tie asks the tie resolvers.** Robert's
   call (rules reading), from the effort review. Fanatical "always loses ties
   during attacks", and Tripwire Cells is an attack (Decision 4), so
   `contestPasses` asks `dispatch.ResolveTie` when a `roll = "attack"` contest
   ties. `TieAttackerWins` gives the tie to the attacker; any other answer keeps
   it with the defender, as before. A `test` contest never asks: a test is not an
   attack. The Phase 2 code spec below predates this. The A-side gate
   (`passesGate`) rolls a test, so it asks no tie resolver; Effort 5 checks
   whether the book words any ability contest as an attack.

3. **Effort close: `rival_roll` asks the wrong player. Deferred.** From the effort
   review. `RollWithHooks` confirms an elective reroll through the roller's
   collector, so the Academy's forced reroll on a rival is confirmed by that
   rival, not by the Academy's owner, and a contest with no collector always
   declines it. A non-elective `rival_roll` with a once-per-turn gate would fire
   on the first rival roll of the turn and spend the budget. No `tags.toml` entry
   uses `rival_roll` yet (Decision 7, D-7), so nothing plays wrong today. Asking
   the owner needs a design: a second collector on the roll, or the owner's
   collector found from the registry. The effort that gives Psychic Academy its
   entry settles it.

4. **Effort close: the XP lines tell the whole total.** Robert's call ("the XP line
   should be accurate"), settling Decision Record 1. The digest adds an
   `XPAwarded` with cause `xp_roll` to the latest `GoalCompleted` of the beat, so
   the goal line reads "+2 XP". The TUI turn list formats one mutation at a time
   and cannot fold, so it shows the bonus on its own line, "bonus XP +1", and
   still hides every other `XPAwarded`.

---

## Shared Context

### Branch, cadence, and checks

- Branch `feature/data-driven-tags`. Commit per session. At the effort boundary:
  `/gestalt:effort-close A.001.7.E3`.
- Every commit builds and passes: `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file; alignment in the
  tasks is as gofmt leaves it.
- Edit files only with Edit/Write, never `sed` or scripts.
- No saved-state shape changes, so no re-seed. Robert's `campaigns/sandbox`
  rulebook copy needs the `roll` key on its three contest gates (Commit 2) to load,
  and the new tag entries to use them (D-4). Claude does not touch it.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions (hooks):** a nil collector takes only free dice and
  forced rerolls (Decision 2); effect contests roll through the hooks, with the
  phase from the entry (Decisions 3, 4); `Detection` marks the detecting side
  (Decision 5); bystander faction buckets for `rival_roll`, amending the roll
  dispatcher's scope entry (Decision 7); a reactor keeps its registry as well as
  its roller (Decision 3).
- **Graduate to Key Decisions (effect-mutation, orchestrator):** goal mutations
  reach the reactors in two passes (Decision 6).

## Out of Scope

- `settings.toml`, the Starship rule, Colonists, Technical Expertise, Pirates
  (Effort 4).
- Psychic Academy's `[[effects]]` entry (D-7, `R.025`).
- Contests that ask the GM (overview Decision 2's follow-on row).
- The A-side tag holder (Effort 5).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — What the contests need**
  - Commit 1 — `feat: a roll with no collector takes only free dice` (2 tasks)
  - Commit 2 — `refactor: carry the roll key, registry and rulebook to the effect contests` (11 tasks)
- [x] **Phase 2 — Contests roll through the hooks**
  - Commit 3 — `feat: the three effect contests roll through the hooks` (5 tasks)
- [x] **Phase 3 — Detection and a reroll on any rival's roll**
  - Commit 4 — `feat: Perimeter Agency's detect die` (10 tasks)
  - Commit 5 — `feat: a reroll forced on any rival's roll` (6 tasks)
- [x] **Phase 4 — Goal events reach reactors**
  - Commit 6 — `feat: Deep Rooted is lost on a homeworld change` (7 tasks)
  - Commit 7 — `feat: Exchange Consulate's XP roll` (7 tasks)

### Phase 1 — What the contests need

**In plain words:** a roll with no one to ask takes only free extra dice, and the
contest call sites get the registry, state, rulebook and a `roll` key they will
need. The contests still roll bare dice when this phase lands.

#### Commit 1 — `feat: a roll with no collector takes only free dice`

##### Task 1 — `internal/faction/engine/hooks/dispatch/roll.go`

**Find:**
```go
// faction that registered it, and its budget is read and spent on that faction.
func RollWithHooks(
```
**Replace with:**
```go
// faction that registered it, and its budget is read and spent on that faction.
// A nil collector asks no one: Cat 1 takes every offer with no budget and skips
// the rest, and Cat 2 declines every elective reroll (overview Decision 2).
func RollWithHooks(
```

**Find:**
```go
		offers := offersBudgetFilter(allOffers, faction, factionState)
		if len(offers) > 0 {
			chosen = collector.SelectModifiers(offers)
		}
```
**Replace with:**
```go
		offers := offersBudgetFilter(allOffers, faction, factionState)
		switch {
		case len(offers) == 0:
		case collector == nil:
			chosen = freeOffers(offers)
		default:
			chosen = collector.SelectModifiers(offers)
		}
```

**Find:**
```go
			if directive.Elective && !collector.ConfirmReroll(directive) {
```
**Replace with:**
```go
			if directive.Elective && (collector == nil || !collector.ConfirmReroll(directive)) {
```

**Find:**
```go
func diceSum(dice []int) int {
```
**Replace with:**
```go
// freeOffers returns the offers with no budget: what a roll with no one to ask
// takes. An extra die where the highest is kept never hurts the roller, while a
// once-per-turn die spent on a roll no one chose could cost a later one.
func freeOffers(offers []hooks.ModifierOffer) []hooks.ModifierOffer {
	var free []hooks.ModifierOffer
	for _, offer := range offers {
		if offer.BudgetKey == "" {
			free = append(free, offer)
		}
	}
	return free
}

func diceSum(dice []int) int {
```

##### Task 2 — `internal/faction/engine/hooks/dispatch/roll_test.go`

Append:
```go
func TestRollWithHooks_NilCollectorTakesOnlyFreeOffers(t *testing.T) {
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "free", addDieModifier("free", ""))
	registry.RegisterRollModifier(hooks.GlobalScope(), "limited", addDieModifier("limited", "tag:Warlike"))
	faction := newFaction("A")

	result := RollWithHooks(rollCtx(), baseRoll(), registry, nil, &fixedRoller{values: []int{5, 7, 6}}, faction, nil, nil)

	if len(result.Dice) != 3 {
		t.Fatalf("expected 3 dice (2 base + the free die), got %d", len(result.Dice))
	}
	if faction.HookBudgets["tag:Warlike"] != 0 {
		t.Errorf("expected the limited offer skipped and its budget unspent, got %d", faction.HookBudgets["tag:Warlike"])
	}
}

func TestRollWithHooks_NilCollectorDeclinesElectiveRerolls(t *testing.T) {
	cases := []struct {
		name     string
		elective bool
		want     int
	}{
		{"a forced reroll applies", false, 9},
		{"an elective reroll is declined", true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry := hooks.NewRegistry()
			registry.RegisterRollResultHook(hooks.GlobalScope(), "reroll", stubResultHook{directive: hooks.RerollDirective{
				Source: "reroll", Indices: []int{0}, Elective: tc.elective,
			}})

			result := RollWithHooks(rollCtx(), baseRoll(), registry, nil, &fixedRoller{values: []int{1, 7, 9}}, newFaction("A"), nil, nil)

			if result.Dice[0] != tc.want {
				t.Errorf("die 0 = %d, want %d", result.Dice[0], tc.want)
			}
		})
	}
}
```

##### Commit message

```
feat: a roll with no collector takes only free dice

- RollWithHooks with a nil collector takes every Cat 1 offer with no budget
  and skips the rest
- Cat 2 declines elective rerolls with no collector; forced rerolls apply
```

#### Commit 2 — `refactor: carry the roll key, registry and rulebook to the effect contests`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
const (
	ContestRollerSelf   ContestRoller = "self"
	ContestRollerTarget ContestRoller = "target"
)
```
**Replace with:**
```go
const (
	ContestRollerSelf   ContestRoller = "self"
	ContestRollerTarget ContestRoller = "target"
)

// ContestRoll names the roll phases an effect contest rolls with, as the book
// words it: an attack and its defense (Tripwire Cells' "Cunning vs. Cunning
// attack"), or a faction test and its contested roll (Panopticon Matrix's and
// Lobbyists' "test"), as the A-side gate rolls.
type ContestRoll string

const (
	ContestRollAttack ContestRoll = "attack"
	ContestRollTest   ContestRoll = "test"
)
```

**Find:**
```go
type EffectContest struct {
	Contest
	Roller ContestRoller
}
```
**Replace with:**
```go
type EffectContest struct {
	Contest
	Roller ContestRoller
	Roll   ContestRoll
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
type effectContestRecord struct {
	Roller       string `toml:"roller"`
```
**Replace with:**
```go
type effectContestRecord struct {
	Roller       string `toml:"roller"`
	Roll         string `toml:"roll"`
```

**Find:**
```go
	contestRollers = []domain.ContestRoller{domain.ContestRollerSelf, domain.ContestRollerTarget}
```
**Replace with:**
```go
	contestRollers = []domain.ContestRoller{domain.ContestRollerSelf, domain.ContestRollerTarget}
	contestRolls   = []domain.ContestRoll{domain.ContestRollAttack, domain.ContestRollTest}
```

**Find:**
```go
		roller, err := toEnum("contest roller", contest.Roller, contestRollers)
		if err != nil {
			return domain.EffectGates{}, err
		}
```
**Replace with:**
```go
		roller, err := toEnum("contest roller", contest.Roller, contestRollers)
		if err != nil {
			return domain.EffectGates{}, err
		}
		roll, err := toEnum("contest roll", contest.Roll, contestRolls)
		if err != nil {
			return domain.EffectGates{}, err
		}
```

**Find:**
```go
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender},
			Roller:  roller,
		}
```
**Replace with:**
```go
			Contest: domain.Contest{AttackerStat: attacker, DefenderStat: defender},
			Roller:  roller,
			Roll:    roll,
		}
```

##### Task 3 — `rulebooks/swn/assets/cunning_assets.toml`

**Find:**
```toml
[assets.Lobbyists.effects.gates.contest]
roller        = "self"
```
**Replace with:**
```toml
[assets.Lobbyists.effects.gates.contest]
roller        = "self"
roll          = "test"
```

**Find:**
```toml
[assets.TripwireCells.effects.gates.contest]
roller        = "self"
```
**Replace with:**
```toml
[assets.TripwireCells.effects.gates.contest]
roller        = "self"
roll          = "attack"
```

**Find:**
```toml
[assets.PanopticonMatrix.effects.gates.contest]
roller        = "target"
```
**Replace with:**
```toml
[assets.PanopticonMatrix.effects.gates.contest]
roller        = "target"
roll          = "test"
```

##### Task 4 — `internal/faction/engine/hooks/rule_modifier.go`

**Find:**
```go
import "github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
```
**Replace with:**
```go
import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)
```

**Find:**
```go
// rolls the hook's own test against the requester.
type PermissionContester interface {
	OfferContest(query PermissionQuery) *PermissionContest
	WinsContest(query PermissionQuery) bool
}
```
**Replace with:**
```go
// rolls the hook's own test against the requester. WinsContest gets the state and
// rulebook because the test rolls through the roll hooks (Effort 3 Decision 3).
type PermissionContester interface {
	OfferContest(query PermissionQuery) *PermissionContest
	WinsContest(query PermissionQuery, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool
}
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/permission.go`

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"` to the
import block, between `engine/hooks` and `state`.

**Find:**
```go
// requester's homeworld (Effort 4 Decision 15). cause names the purchase path on
// the record.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```
**Replace with:**
```go
// requester's homeworld (Effort 4 Decision 15). cause names the purchase path on
// the record. rulebook reaches the roll hooks of a Lobbyists contest.
func RequestPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, rulebook *rulebook.Rulebook, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```

**Find:**
```go
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
```
**Replace with:**
```go
	if worldID == requester.Homeworld.WorldID {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, rulebook, requester, def, worldID, cause)
```

**Find:**
```go
func RequestArrivalPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsArrivalPermission(asset) {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, requester, def, worldID, cause)
```
**Replace with:**
```go
func RequestArrivalPermission(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, rulebook *rulebook.Rulebook, requester *domain.Faction, asset *domain.Asset, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
	if !def.NeedsArrivalPermission(asset) {
		return nil, nil
	}
	return askGovernment(registry, collector, factionState, rulebook, requester, def, worldID, cause)
```

**Find:**
```go
func askGovernment(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```
**Replace with:**
```go
func askGovernment(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, rulebook *rulebook.Rulebook, requester *domain.Faction, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```

**Find:**
```go
	return contestGrant(registry, collector, factionState, requester, grantedBy, def, worldID, cause)
```
**Replace with:**
```go
	return contestGrant(registry, collector, factionState, rulebook, requester, grantedBy, def, worldID, cause)
```

**Find:**
```go
func contestGrant(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, requester *domain.Faction, grantedBy string, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```
**Replace with:**
```go
func contestGrant(registry *hooks.Registry, collector hooks.PermissionCollector, factionState *state.FactionState, rulebook *rulebook.Rulebook, requester *domain.Faction, grantedBy string, def *domain.AssetDefinition, worldID, cause string) (*domain.PermissionRefused, error) {
```

**Find:**
```go
				if !contest || !registered.Hook.WinsContest(query) {
```
**Replace with:**
```go
				if !contest || !registered.Hook.WinsContest(query, factionState, rulebook) {
```

##### Task 6 — the five permission callers

Each call gains the rulebook after the state. Every caller already holds it.

| File | Find | Replace with |
|---|---|---|
| `engine/action/actions/buy_asset.go` | `dispatch.RequestPermission(ba.registry, ba.collector, factionState, faction, order.Definition, order.World, "buy")` | `dispatch.RequestPermission(ba.registry, ba.collector, factionState, rulebook, faction, order.Definition, order.World, "buy")` |
| `engine/action/actions/refit_asset.go` | `dispatch.RequestPermission(ra.registry, ra.collector, factionState, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")` | `dispatch.RequestPermission(ra.registry, ra.collector, factionState, rulebook, faction, order.NewDefinition, order.OldAsset.Location.WorldID, "refit")` |
| `engine/action/actions/ability/purchase_enabler.go` | `dispatch.RequestPermission(ctx.Registry, ctx.Collector, ctx.FactionState, ctx.Faction, def, worldID, "ability")` | `dispatch.RequestPermission(ctx.Registry, ctx.Collector, ctx.FactionState, ctx.Rulebook, ctx.Faction, def, worldID, "ability")` |
| `engine/relocate/relocate.go` | `dispatch.RequestArrivalPermission(request.Registry, request.Collector, request.FactionState, request.Faction, order.Asset, def, order.DestinationWorld, domain.PermissionCauseRelocate)` | `dispatch.RequestArrivalPermission(request.Registry, request.Collector, request.FactionState, request.Rulebook, request.Faction, order.Asset, def, order.DestinationWorld, domain.PermissionCauseRelocate)` |
| `engine/arrival.go` | `dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)` | `dispatch.RequestArrivalPermission(e.Hooks, collector, factionState, e.Rulebook, requester, asset, def, order.Destination.WorldID, domain.PermissionCauseArrival)` |

##### Task 7 — `internal/faction/engine/effect/registrar.go`

**Find:**
```go
// the subject: the bought asset, a quality's target, or at turn start the
// holder's own asset.
type FireContext struct {
	Holder       Holder
	Asset        *domain.Asset
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
}
```
**Replace with:**
```go
// the subject: the bought asset, a quality's target, or at turn start the
// holder's own asset. Registry and Rulebook ride along because a contest gate
// rolls through the roll hooks (Effort 3 Decision 3); FireQuality leaves them
// nil, as no quality rolls.
type FireContext struct {
	Holder       Holder
	Asset        *domain.Asset
	Entry        domain.EffectDefinition
	Roller       domain.Roller
	FactionState *state.FactionState
	Registry     *hooks.Registry
	Rulebook     *rulebook.Rulebook
}
```

##### Task 8 — `internal/faction/engine/effect/effect.go`

**Find:**
```go
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
```
**Replace with:**
```go
func (e *EffectsEngine) FireTurnStart(faction *domain.Faction, factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller) []domain.Mutation {
```

**Find:**
```go
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerTurnStart, asset, factionState, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerTurnStart, nil, factionState, roller)...)
	}
```
**Replace with:**
```go
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerTurnStart, asset, factionState, rulebook, hookRegistry, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerTurnStart, nil, factionState, rulebook, hookRegistry, roller)...)
	}
```

**Find:**
```go
func (e *EffectsEngine) fireEntries(holder Holder, entries []domain.EffectDefinition, trigger domain.EffectTrigger, subject *domain.Asset, factionState *state.FactionState, roller domain.Roller) []domain.Mutation {
```
**Replace with:**
```go
func (e *EffectsEngine) fireEntries(holder Holder, entries []domain.EffectDefinition, trigger domain.EffectTrigger, subject *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller) []domain.Mutation {
```

**Find:**
```go
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
		})...)
```
**Replace with:**
```go
			Entry:        entry,
			Roller:       roller,
			FactionState: factionState,
			Registry:     hookRegistry,
			Rulebook:     rulebook,
		})...)
```

**Find:**
```go
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	if def := rulebook.Assets[asset.DefinitionID]; def != nil {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerOnPurchase, asset, factionState, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerOnPurchase, asset, factionState, roller)...)
	}
```
**Replace with:**
```go
func (e *EffectsEngine) FirePurchase(faction *domain.Faction, asset *domain.Asset, factionState *state.FactionState, rulebook *rulebook.Rulebook, hookRegistry *hooks.Registry, roller domain.Roller) []domain.Mutation {
	var mutations []domain.Mutation
	if def := rulebook.Assets[asset.DefinitionID]; def != nil {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Asset: asset, Def: def}, def.Effects, domain.TriggerOnPurchase, asset, factionState, rulebook, hookRegistry, roller)...)
	}
	for _, tag := range heldTags(faction, rulebook) {
		mutations = append(mutations, e.fireEntries(Holder{Faction: faction, Tag: tag}, tag.Effects, domain.TriggerOnPurchase, asset, factionState, rulebook, hookRegistry, roller)...)
	}
```

##### Task 9 — `internal/faction/engine/orchestrator.go`

**Find:**
```go
e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Rand)
```
**Replace with:**
```go
e.Effect.FireTurnStart(faction, factionState, e.Rulebook, e.Hooks, e.Rand)
```

**Find:**
```go
e.Effect.FirePurchase(faction, &added.Asset, factionState, e.Rulebook, e.Rand)
```
**Replace with:**
```go
e.Effect.FirePurchase(faction, &added.Asset, factionState, e.Rulebook, e.Hooks, e.Rand)
```

##### Task 10 — `internal/faction/engine/effect/effects/withdraw_permission.go`

The signature only; the contest stays bare until Commit 3.

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"` and
`"github.com/therobertcrocker/gm-toolkit/internal/faction/state"` to the import
block, after `engine/hooks`.

**Find:**
```go
// WinsContest rolls the bare contest, as every effect contest does (Effort 3
// Decision 12): the request path carries no rulebook for roll hooks.
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery) bool {
```
**Replace with:**
```go
// WinsContest rolls the bare contest, as every effect contest does (Effort 3
// Decision 12).
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery, _ *state.FactionState, _ *rulebook.Rulebook) bool {
```

##### Task 11 — tests

`internal/faction/engine/effect/effects/withdraw_permission_test.go`:

**Find:**
```go
			if got := contester.WinsContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos"}); got != tc.want {
```
**Replace with:**
```go
			if got := contester.WinsContest(hooks.PermissionQuery{Requester: bravo, WorldID: "Krylos"}, nil, nil); got != tc.want {
```

`internal/faction/engine/hooks/dispatch/permission_test.go`: add
`"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"` to the
imports, before `state`. Then:

**Find:**
```go
func (stub *stubContester) WinsContest(hooks.PermissionQuery) bool {
```
**Replace with:**
```go
func (stub *stubContester) WinsContest(hooks.PermissionQuery, *state.FactionState, *rulebook.Rulebook) bool {
```

Each of the nine calls gains `nil, ` after `factionState, `:

| Line | Find | Replace with |
|---|---|---|
| 81 | `RequestPermission(nil, collector, factionState, requester, tc.def, tc.world, "buy")` | `RequestPermission(nil, collector, factionState, nil, requester, tc.def, tc.world, "buy")` |
| 108 | `RequestPermission(registry, collector, factionState, requester, def, "Krylos", "buy")` | `RequestPermission(registry, collector, factionState, nil, requester, def, "Krylos", "buy")` |
| 161 | `RequestArrivalPermission(nil, collector, factionState, requester, tc.asset, tc.def, tc.world, domain.PermissionCauseArrival)` | `RequestArrivalPermission(nil, collector, factionState, nil, requester, tc.asset, tc.def, tc.world, domain.PermissionCauseArrival)` |
| 230, 261 | `RequestPermission(registry, collector, factionState, requester, permissionDef, "Tartarus", "buy")` | `RequestPermission(registry, collector, factionState, nil, requester, permissionDef, "Tartarus", "buy")` |
| 300 | `RequestPermission(registry, collector, factionState, requester, tc.def, tc.world, "buy")` | `RequestPermission(registry, collector, factionState, nil, requester, tc.def, tc.world, "buy")` |
| 332 | `RequestPermission(registry, collector, factionState, requester, permissionDef, "Krylos", "buy")` | `RequestPermission(registry, collector, factionState, nil, requester, permissionDef, "Krylos", "buy")` |
| 348 | `RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Krylos", domain.PermissionCauseArrival)` | `RequestArrivalPermission(nil, collector, factionState, nil, requester, &domain.Asset{}, permissionDef, "Krylos", domain.PermissionCauseArrival)` |
| 360 | `RequestArrivalPermission(nil, collector, factionState, requester, &domain.Asset{}, permissionDef, "Hadrian", domain.PermissionCauseArrival)` | `RequestArrivalPermission(nil, collector, factionState, nil, requester, &domain.Asset{}, permissionDef, "Hadrian", domain.PermissionCauseArrival)` |

The `FireTurnStart` and `FirePurchase` test calls gain `nil, ` before the last
argument (the registry, before the roller):

| File:line | Find | Replace with |
|---|---|---|
| `engine/effect/effect_test.go:491` | `FireTurnStart(faction, nil, rb, nil)` | `FireTurnStart(faction, nil, rb, nil, nil)` |
| `engine/effect/effect_test.go:526` | `engine.FirePurchase(faction, bought, nil, rb, nil)` | `engine.FirePurchase(faction, bought, nil, rb, nil, nil)` |
| `engine/effect/effect_test.go:535` | `engine.FirePurchase(faction, unknown, nil, rb, nil)` | `engine.FirePurchase(faction, unknown, nil, rb, nil, nil)` |
| `engine/effect/effect_test.go:561` | `FireTurnStart(faction, factionState, rb, nil)` | `FireTurnStart(faction, factionState, rb, nil, nil)` |
| `engine/effect/effect_test.go:685` | `engine.FireTurnStart(alpha, factionState, rb, nil)` | `engine.FireTurnStart(alpha, factionState, rb, nil, nil)` |
| `engine/effect/effect_test.go:686` | `factionState, rb, nil)` (the `FirePurchase` of `"unknown"`) | `factionState, rb, nil, nil)` |
| `engine/effect/effects/apply_stealth_test.go:72` | `engine.FirePurchase(alpha, bought, nil, rb, nil)` | `engine.FirePurchase(alpha, bought, nil, rb, nil, nil)` |

`internal/faction/rulebook/effects_test.go`:

**Find:**
```toml
[assets.Machine.effects.gates.contest]
roller        = "target"
```
**Replace with:**
```toml
[assets.Machine.effects.gates.contest]
roller        = "target"
roll          = "test"
```

**Find:**
```go
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
					Roller:  domain.ContestRollerTarget,
				},
```
**Replace with:**
```go
					Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
					Roller:  domain.ContestRollerTarget,
					Roll:    domain.ContestRollTest,
				},
```

**Find:**
```go
			wantErr: `unknown contest roller: ""`,
		},
```
**Replace with:**
```go
			wantErr: `unknown contest roller: ""`,
		},
		{
			name: "contest without a roll",
			body: `
[[assets.Machine.effects]]
trigger = "stealth_arrival"
effect  = "clear_stealth"` + filters + `
[assets.Machine.effects.gates.contest]
roller        = "self"
attacker_stat = "Cunning"
defender_stat = "Cunning"
`,
			wantErr: `unknown contest roll: ""`,
		},
```

`internal/faction/rulebook/rulebook_test.go`:

**Find:**
```go
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
		}},
	}}
	if got := rb.Assets["SWN-C2-001"].Effects; !reflect.DeepEqual(got, want) {
```
**Replace with:**
```go
			Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			Roller:  domain.ContestRollerSelf,
			Roll:    domain.ContestRollTest,
		}},
	}}
	if got := rb.Assets["SWN-C2-001"].Effects; !reflect.DeepEqual(got, want) {
```

Append:
```go
func TestLoad_ContestRolls(t *testing.T) {
	rb := loadShippedRulebook(t)
	cases := []struct {
		name    string
		assetID string
		want    domain.ContestRoll
	}{
		{"Tripwire Cells", "SWN-C4-003", domain.ContestRollAttack},
		{"Panopticon Matrix", "SWN-C8-001", domain.ContestRollTest},
		{"Lobbyists", "SWN-C2-001", domain.ContestRollTest},
	}
	for _, tc := range cases {
		var got domain.ContestRoll
		for _, entry := range rb.Assets[tc.assetID].Effects {
			if entry.Gates.Contest != nil {
				got = entry.Gates.Contest.Roll
			}
		}
		if got != tc.want {
			t.Errorf("%s contest roll = %q, want %q", tc.name, got, tc.want)
		}
	}
}
```

##### Commit message

```
refactor: carry the roll key, registry and rulebook to the effect contests

- the contest gate gains a required roll key: attack (Tripwire Cells) or
  test (Panopticon Matrix, Lobbyists)
- FireContext gains Registry and Rulebook; FireTurnStart and FirePurchase
  take the registry
- WinsContest takes the state and rulebook; RequestPermission and
  RequestArrivalPermission take the rulebook, and their five callers pass it
- no behaviour change: the contests still roll bare dice
```

### Phase 2 — Contests roll through the hooks

**In plain words:** the three asset contests roll through the same hooked path as
an Attack, with no one asked, so Fanatical's rerolls and other free dice reach
them and each roll names the stat on the other side.

#### Commit 3 — `feat: the three effect contests roll through the hooks`

##### Task 1 — `internal/faction/engine/effect/effects/contest.go` (rewritten)

Full contents:
```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// contestRoll is one effect contest, seen from the effect: the declaring faction
// and its asset, the affected faction and its asset (nil when no asset stands for
// it: Lobbyists' requester, or an asset bought in this batch), the world, and what
// the hooked rolls read. collector is nil until the roll-card follow-on row gives
// each contest one (overview Decision 2).
type contestRoll struct {
	contest        domain.EffectContest
	declaring      *domain.Faction
	declaringAsset *domain.Asset
	target         *domain.Faction
	targetAsset    *domain.Asset
	world          string
	collector      hooks.Collector
	registry       *hooks.Registry
	factionState   *state.FactionState
	rulebook       *rulebook.Rulebook
	roller         domain.Roller
}

// contestPasses rolls an effect's contest gate through the roll hooks and reports
// whether the effect applies. With roller "self" the declaring faction attacks
// and must win; with "target" the affected faction tests to keep what it has, and
// the effect applies when that test fails. Ties go to the defender, as in the
// A-side gate. roll "attack" rolls an attack and its defense; roll "test" a
// faction test and its contested roll, as the A-side gate does (Effort 3
// Decision 4).
func contestPasses(roll contestRoll) bool {
	attacker, attackerAsset, defender, defenderAsset := roll.declaring, roll.declaringAsset, roll.target, roll.targetAsset
	if roll.contest.Roller == domain.ContestRollerTarget {
		attacker, attackerAsset, defender, defenderAsset = roll.target, roll.targetAsset, roll.declaring, roll.declaringAsset
	}
	attackPhase, defensePhase := hooks.PhaseFactionTest, hooks.PhaseContested
	if roll.contest.Roll == domain.ContestRollAttack {
		attackPhase, defensePhase = hooks.PhaseAttack, hooks.PhaseDefense
	}
	attackerStat, defenderStat := roll.contest.AttackerStat, roll.contest.DefenderStat
	attack := roll.side(attackPhase, attacker, defender, attackerAsset, defenderAsset, attackerStat, defenderStat)
	defense := roll.side(defensePhase, defender, attacker, defenderAsset, attackerAsset, defenderStat, attackerStat)
	attackerWins := attack > defense
	if roll.contest.Roller == domain.ContestRollerTarget {
		return !attackerWins
	}
	return attackerWins
}

// side rolls one side of the contest: 1d10 plus actor's stat, with the stat the
// other side rolls as OpponentAttribute (Effort 2 Decision 3).
func (roll contestRoll) side(phase hooks.RollPhase, actor, opponent *domain.Faction, asset, opponentAsset *domain.Asset, stat, opponentStat domain.FactionStat) int {
	ctx := hooks.RollContext{
		Phase:             phase,
		Actor:             actor,
		Opponent:          opponent,
		Attribute:         string(stat),
		OpponentAttribute: string(opponentStat),
		Asset:             asset,
		OpponentAsset:     opponentAsset,
		World:             roll.world,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: actor.StatScore(stat)}
	return dispatch.RollWithHooks(ctx, base, roll.registry, roll.collector, roll.roller, actor, roll.factionState, roll.rulebook).Sum
}
```

##### Task 2 — `internal/faction/engine/effect/effects/clear_stealth.go`

Add `"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"` to
the import block, after `engine/effect`.

**Find:**
```go
			if ctx.Entry.Gates.Contest != nil && !contestPasses(*ctx.Entry.Gates.Contest, ctx.Holder.Faction, target, ctx.Roller) {
				continue
			}
```
**Replace with:**
```go
			if ctx.Entry.Gates.Contest != nil && !contestPasses(contestRoll{
				contest:        *ctx.Entry.Gates.Contest,
				declaring:      ctx.Holder.Faction,
				declaringAsset: ctx.Asset,
				target:         target,
				targetAsset:    asset,
				world:          worldID,
				// deliberate: a nil collector, not a prompt, because a phase route has none; the roll-card follow-on row gives it one (overview Decision 2).
				collector:    nil,
				registry:     ctx.Registry,
				factionState: ctx.FactionState,
				rulebook:     ctx.Rulebook,
				roller:       ctx.Roller,
			}) {
				continue
			}
```

**Find:**
```go
			contest:  ctx.Entry.Gates.Contest,
			roller:   ctx.Roller,
		},
	)
}

type clearStealthReactor struct {
	holder   holderRef
	stealthy *bool
	contest  *domain.EffectContest
	roller   domain.Roller
}

func (reactor *clearStealthReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
```
**Replace with:**
```go
			contest:  ctx.Entry.Gates.Contest,
			roller:   ctx.Roller,
			registry: ctx.Registry,
		},
	)
}

// clearStealthReactor keeps the roller and registry its registration carried:
// OnMutations takes neither, and its contest rolls through the roll hooks.
type clearStealthReactor struct {
	holder   holderRef
	stealthy *bool
	contest  *domain.EffectContest
	roller   domain.Roller
	registry *hooks.Registry
}

func (reactor *clearStealthReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, rulebook *rulebook.Rulebook) []domain.Mutation {
```

**Find:**
```go
		if reactor.contest != nil && !contestPasses(*reactor.contest, owner, target, reactor.roller) {
			continue
		}
```
**Replace with:**
```go
		if reactor.contest != nil && !contestPasses(contestRoll{
			contest:        *reactor.contest,
			declaring:      owner,
			declaringAsset: self,
			target:         target,
			targetAsset:    target.Assets[assetID],
			world:          worldID,
			// deliberate: a nil collector, not a prompt, because a reactor has none; the roll-card follow-on row gives it one (overview Decision 2).
			collector:    nil,
			registry:     reactor.registry,
			factionState: factionState,
			rulebook:     rulebook,
			roller:       reactor.roller,
		}) {
			continue
		}
```

##### Task 3 — `internal/faction/engine/effect/effects/withdraw_permission.go`

**Find:**
```go
		&permissionContester{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder), contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller its
// registration carried, as a reactor does.
type permissionContester struct {
	owner   *domain.Faction
	holder  holderRef
	contest domain.EffectContest
	roller  domain.Roller
}
```
**Replace with:**
```go
		&permissionContester{owner: ctx.Holder.Faction, holder: refOf(ctx.Holder), contest: *ctx.Entry.Gates.Contest, roller: ctx.Roller, registry: ctx.Registry},
	)
}

// permissionContester holds its owner, not its asset: a Cat 4 hook gets no
// FactionState (Effort 2 Decision Record 19). It keeps the roller and registry
// its registration carried, as a reactor does.
type permissionContester struct {
	owner    *domain.Faction
	holder   holderRef
	contest  domain.EffectContest
	roller   domain.Roller
	registry *hooks.Registry
}
```

**Find:**
```go
// WinsContest rolls the bare contest, as every effect contest does (Effort 3
// Decision 12).
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery, _ *state.FactionState, _ *rulebook.Rulebook) bool {
	return contestPasses(contester.contest, contester.owner, query.Requester, contester.roller)
}
```
**Replace with:**
```go
// WinsContest rolls the contest through the roll hooks (Effort 3 Decision 3),
// with the state and rulebook the permission dispatcher holds. The requester has
// no asset on the roll: it asks for one that is not there yet.
func (contester *permissionContester) WinsContest(query hooks.PermissionQuery, factionState *state.FactionState, rulebook *rulebook.Rulebook) bool {
	self, _ := contester.holder.liveIn(contester.owner)
	return contestPasses(contestRoll{
		contest:        contester.contest,
		declaring:      contester.owner,
		declaringAsset: self,
		target:         query.Requester,
		world:          query.WorldID,
		// deliberate: a nil collector, not a prompt, because the permission path holds only a PermissionCollector; the roll-card follow-on row gives it one (overview Decision 2).
		collector:    nil,
		registry:     contester.registry,
		factionState: factionState,
		rulebook:     rulebook,
		roller:       contester.roller,
	})
}
```

##### Task 4 — `rulebooks/swn/tags.toml`

Imperialists' die keeps to the Attack action (Decision 4).

**Find:**
```toml
[tags.Imperialists.effects.filters]
owner = "own"
scope = "faction"
phase = ["attack"]
goal  = "G-004"
```
**Replace with:**
```toml
[tags.Imperialists.effects.filters]
owner  = "own"
scope  = "faction"
phase  = ["attack"]
goal   = "G-004"
action = "Attack"
```

##### Task 5 — tests

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-006", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Goal: "G-004"}, "")},
```
**Replace with:**
```go
		{"SWN-T-006", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Goal: "G-004", Action: "Attack"}, "")},
```

`internal/faction/engine/effect/effects/contest_test.go` (rewritten) — full
contents:
```go
package effects

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

type pairRoller struct {
	values []int
	next   int
}

func (roller *pairRoller) Roll(int) int {
	value := roller.values[roller.next]
	roller.next++
	return value
}

// rollRecorder is a Cat 1 hook that records each roll it sees and offers nothing.
type rollRecorder struct {
	seen []hooks.RollContext
}

func (recorder *rollRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.seen = append(recorder.seen, ctx)
	return nil
}

func cunningContest(roller domain.ContestRoller, roll domain.ContestRoll) domain.EffectContest {
	return domain.EffectContest{
		Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
		Roller:  roller,
		Roll:    roll,
	}
}

func TestContestPasses(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 3}
	target := &domain.Faction{ID: "beta", Cunning: 2}
	cases := []struct {
		name   string
		roller domain.ContestRoller
		rolls  []int
		want   bool
	}{
		{"self: the declaring faction wins", domain.ContestRollerSelf, []int{6, 2}, true},
		{"self: the declaring faction loses", domain.ContestRollerSelf, []int{1, 10}, false},
		{"self: a tie fails", domain.ContestRollerSelf, []int{5, 6}, false},
		{"target: the target fails its test", domain.ContestRollerTarget, []int{1, 10}, true},
		{"target: the target passes its test", domain.ContestRollerTarget, []int{10, 1}, false},
		{"target: a tie passes the effect", domain.ContestRollerTarget, []int{6, 5}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roll := contestRoll{
				contest:   cunningContest(tc.roller, domain.ContestRollTest),
				declaring: declaring,
				target:    target,
				roller:    &pairRoller{values: tc.rolls},
			}
			if got := contestPasses(roll); got != tc.want {
				t.Errorf("contestPasses = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContestPasses_RollPhasesAndSides(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 3, Wealth: 1}
	target := &domain.Faction{ID: "beta", Cunning: 2, Wealth: 4}
	cells := &domain.Asset{ID: "cells"}
	spy := &domain.Asset{ID: "spy"}
	cases := []struct {
		name            string
		roll            domain.ContestRoll
		attack, defense hooks.RollPhase
	}{
		{"an attack", domain.ContestRollAttack, hooks.PhaseAttack, hooks.PhaseDefense},
		{"a test", domain.ContestRollTest, hooks.PhaseFactionTest, hooks.PhaseContested},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &rollRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)
			contest := domain.EffectContest{
				Contest: domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatWealth},
				Roller:  domain.ContestRollerSelf,
				Roll:    tc.roll,
			}

			contestPasses(contestRoll{
				contest: contest, declaring: declaring, declaringAsset: cells, target: target, targetAsset: spy,
				world: "Krylos", registry: registry, roller: &pairRoller{values: []int{5, 5}},
			})

			want := []hooks.RollContext{
				{Phase: tc.attack, Actor: declaring, Opponent: target, Attribute: "Cunning", OpponentAttribute: "Wealth", Asset: cells, OpponentAsset: spy, World: "Krylos"},
				{Phase: tc.defense, Actor: target, Opponent: declaring, Attribute: "Wealth", OpponentAttribute: "Cunning", Asset: spy, OpponentAsset: cells, World: "Krylos"},
			}
			if !reflect.DeepEqual(recorder.seen, want) {
				t.Errorf("rolls = %+v, want %+v", recorder.seen, want)
			}
		})
	}
}

func TestContestPasses_TakesOnlyFreeDice(t *testing.T) {
	declaring := &domain.Faction{ID: "alpha", Cunning: 1, Tags: []string{"T"}, HookBudgets: map[string]int{}}
	target := &domain.Faction{ID: "beta", Cunning: 1}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": declaring, "beta": target}}
	registry := hooks.NewRegistry()
	free := domain.EffectDefinition{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}
	limited := free
	limited.Gates = domain.EffectGates{OncePer: domain.OncePerOwnerFaction}
	for _, entry := range []domain.EffectDefinition{free, limited} {
		bonusDieRegistrar.Register(effect.RegisterContext{
			Holder: effect.Holder{Faction: declaring, Tag: &domain.Tag{ID: "T", Name: "Tag"}}, Entry: entry, Registry: registry,
		})
	}
	roll := contestRoll{
		contest:      cunningContest(domain.ContestRollerSelf, domain.ContestRollTest),
		declaring:    declaring,
		target:       target,
		registry:     registry,
		factionState: factionState,
		// alpha rolls 2 and its free die 9, keeping 9; beta rolls 5.
		roller: &pairRoller{values: []int{2, 9, 5}},
	}

	if !contestPasses(roll) {
		t.Error("contestPasses = false, want true: the free die's 9 beats beta's 5")
	}
	if len(declaring.HookBudgets) != 0 {
		t.Errorf("HookBudgets = %v, want none spent: the once-per-turn die waits", declaring.HookBudgets)
	}
}
```

##### Commit message

```
feat: the three effect contests roll through the hooks

- contestPasses rolls each side through RollWithHooks with a nil collector:
  free dice and forced rerolls reach Tripwire Cells, Panopticon Matrix and
  Lobbyists, and the rolls carry OpponentAttribute
- the roll key picks attack/defense or test/contested
- the Tripwire Cells reactor and the Lobbyists contester keep the registry
  their registration carried
- Imperialists' die keeps to the Attack action, so Tripwire's attack does
  not get it
```

### Phase 3 — Detection and a reroll on any rival's roll

**In plain words:** a roll that tests for stealth is marked, and Perimeter Agency's
data entry adds its die to those rolls; a tag can now force a reroll on any rival's
roll anywhere, which Psychic Academy will use once the roll card exists.

#### Commit 4 — `feat: Perimeter Agency's detect die`

##### Task 1 — `internal/faction/engine/hooks/types.go`

**Find:**
```go
// stands there (a faction test, a Base defending). Action is the action's Name().
type RollContext struct {
```
**Replace with:**
```go
// stands there (a faction test, a Base defending). Action is the action's Name().
// Detection marks a roll made by the side testing to detect Stealthed assets: the
// actor of a reveal ability, or the declaring side of a stealth-clearing contest.
// The side resisting detection never sets it (Effort 3 Decision 5).
type RollContext struct {
```

**Find:**
```go
	World             string
	Action            string
}
```
**Replace with:**
```go
	World             string
	Action            string
	Detection         bool
}
```

##### Task 2 — `internal/faction/domain/effect.go`

**Find:**
```go
	FilterGoal              EffectFilterName = "goal"
)
```
**Replace with:**
```go
	FilterGoal              EffectFilterName = "goal"
	FilterDetection         EffectFilterName = "detection"
)
```

**Find:**
```go
// filter.
type EffectFilters struct {
```
**Replace with:**
```go
// filter. Detection is a plain bool: only "detection = true" filters.
type EffectFilters struct {
```

**Find:**
```go
	CausedBy          EffectOwner
	Goal              string
}
```
**Replace with:**
```go
	CausedBy          EffectOwner
	Goal              string
	Detection         bool
}
```

**Find:**
```go
	if filters.Goal != "" {
		present = append(present, FilterGoal)
	}
	return present
```
**Replace with:**
```go
	if filters.Goal != "" {
		present = append(present, FilterGoal)
	}
	if filters.Detection {
		present = append(present, FilterDetection)
	}
	return present
```

##### Task 3 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	CausedBy          string   `toml:"caused_by"`
	Goal              string   `toml:"goal"`
}
```
**Replace with:**
```go
	CausedBy          string   `toml:"caused_by"`
	Goal              string   `toml:"goal"`
	Detection         bool     `toml:"detection"`
}
```

**Find:**
```go
		Goal: record.Goal,
	}
```
**Replace with:**
```go
		Goal: record.Goal, Detection: record.Detection,
	}
```

##### Task 4 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
// matches without a rulebook. goal matches a roll against a rival on the target
// world of the owner's active goal.
```
**Replace with:**
```go
// matches without a rulebook. goal matches a roll against a rival on the target
// world of the owner's active goal. detection matches only a roll made by the
// side testing to detect stealth.
```

**Find:**
```go
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
```
**Replace with:**
```go
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	if filters.Detection && !ctx.Detection {
		return false
	}
```

##### Task 5 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal,
```
**Replace with:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal, domain.FilterDetection,
```

##### Task 6 — `internal/faction/engine/effect/effects/contest.go`

Uses `contestRoll` and `side`, defined in Commit 3 Task 1.

**Find:**
```go
// the hooked rolls read. collector is nil until the roll-card follow-on row gives
// each contest one (overview Decision 2).
type contestRoll struct {
	contest        domain.EffectContest
```
**Replace with:**
```go
// the hooked rolls read. collector is nil until the roll-card follow-on row gives
// each contest one (overview Decision 2). declaringDetects marks the declaring
// side's roll as a detection roll (Effort 3 Decision 5).
type contestRoll struct {
	contest          domain.EffectContest
	declaringDetects bool
```

**Find:**
```go
		World:             roll.world,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: actor.StatScore(stat)}
```
**Replace with:**
```go
		World:             roll.world,
		Detection:         roll.declaringDetects && actor == roll.declaring,
	}
	base := domain.DiceRoll{NumDice: 1, Sides: 10, Modifier: actor.StatScore(stat)}
```

##### Task 7 — `internal/faction/engine/effect/effects/clear_stealth.go`

Both contest literals from Commit 3 Task 2 gain the flag.

**Find:**
```go
			if ctx.Entry.Gates.Contest != nil && !contestPasses(contestRoll{
				contest:        *ctx.Entry.Gates.Contest,
```
**Replace with:**
```go
			if ctx.Entry.Gates.Contest != nil && !contestPasses(contestRoll{
				contest:          *ctx.Entry.Gates.Contest,
				declaringDetects: true,
```

**Find:**
```go
		if reactor.contest != nil && !contestPasses(contestRoll{
			contest:        *reactor.contest,
```
**Replace with:**
```go
		if reactor.contest != nil && !contestPasses(contestRoll{
			contest:          *reactor.contest,
			declaringDetects: true,
```

##### Task 8 — `internal/faction/engine/action/actions/ability/stage.go`

In `passesGate`, the attacker's roll is the detecting side of a reveal.

**Find:**
```go
			Phase:             hooks.PhaseFactionTest,
			Actor:             faction,
```
**Replace with:**
```go
			Phase:             hooks.PhaseFactionTest,
			Detection:         step.Effect == domain.EffectRevealStealth,
			Actor:             faction,
```

##### Task 9 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.PerimeterAgency.effects.bonus_die]
sides = 10
count = 1
```
**Replace with:**
```toml
[tags.PerimeterAgency.effects.bonus_die]
sides = 10
count = 1

[[tags.PerimeterAgency.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.PerimeterAgency.effects.filters]
owner     = "own"
scope     = "faction"
detection = true
[tags.PerimeterAgency.effects.bonus_die]
sides = 10
count = 1
```

##### Task 10 — tests

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```
**Replace with:**
```go
		{"detection, a detecting roll", domain.TriggerOwnRoll, domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Detection: true}, hooks.RollContext{Actor: alpha, Detection: true}, true},
		{"detection, any other roll", domain.TriggerOwnRoll, domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Detection: true}, hooks.RollContext{Actor: alpha}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```

`internal/faction/engine/effect/effects/contest_test.go`, append:
```go
func TestContestPasses_DetectionMarksTheDeclaringSide(t *testing.T) {
	cases := []struct {
		name   string
		roller domain.ContestRoller
		want   []bool
	}{
		{"self: the declaring side attacks", domain.ContestRollerSelf, []bool{true, false}},
		{"target: the declaring side defends", domain.ContestRollerTarget, []bool{false, true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &rollRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)

			contestPasses(contestRoll{
				contest:          cunningContest(tc.roller, domain.ContestRollTest),
				declaringDetects: true,
				declaring:        &domain.Faction{ID: "alpha"},
				target:           &domain.Faction{ID: "beta"},
				registry:         registry,
				roller:           &pairRoller{values: []int{5, 5}},
			})

			var got []bool
			for _, ctx := range recorder.seen {
				got = append(got, ctx.Detection)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Detection per roll = %v, want %v", got, tc.want)
			}
		})
	}
}
```

`internal/faction/engine/action/actions/ability/detection_test.go` (new) — full
contents:
```go
package ability

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// detectionRecorder is a Cat 1 hook that records whether each roll it sees is a
// detection roll, and offers nothing.
type detectionRecorder struct {
	detection []bool
}

func (recorder *detectionRecorder) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	recorder.detection = append(recorder.detection, ctx.Detection)
	return nil
}

func TestPassesGate_RevealMarksTheActorsRollAsDetection(t *testing.T) {
	cases := []struct {
		name   string
		effect domain.AbilityEffectType
		want   []bool
	}{
		{"reveal_stealth", domain.EffectRevealStealth, []bool{true, false}},
		{"another contest", domain.EffectDisableAsset, []bool{false, false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			informers := &domain.Asset{ID: "inf", Location: domain.Location{WorldID: "Krylos"}}
			alpha := &domain.Faction{ID: "alpha", Cunning: 3, Assets: map[string]*domain.Asset{"inf": informers}}
			beta := &domain.Faction{ID: "beta", Cunning: 2}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": beta}}
			recorder := &detectionRecorder{}
			registry := hooks.NewRegistry()
			registry.RegisterRollModifier(hooks.GlobalScope(), "recorder", recorder)
			step := domain.AbilityStep{
				Effect:  tc.effect,
				Target:  domain.TargetFaction,
				Contest: &domain.Contest{AttackerStat: domain.StatCunning, DefenderStat: domain.StatCunning},
			}
			target := domain.AbilityTarget{Kind: domain.TargetFaction, Faction: beta}

			if _, err := passesGate(step, alpha, informers, target, registry, nil, &seqRoller{values: []int{5}}, factionState, nil); err != nil {
				t.Fatalf("passesGate: %v", err)
			}
			if !reflect.DeepEqual(recorder.detection, tc.want) {
				t.Errorf("Detection per roll = %v, want %v (attacker, then defender)", recorder.detection, tc.want)
			}
		})
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-009", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, OpponentTechLevel: &five}, domain.OncePerOwnerFaction)},
```
**Replace with:**
```go
		{"SWN-T-009", append(
			tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, OpponentTechLevel: &five}, domain.OncePerOwnerFaction),
			tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Detection: true}, "")...,
		)},
```

##### Commit message

```
feat: Perimeter Agency's detect die

- RollContext.Detection marks the roll of the side testing to detect
  stealth: the reveal ability's actor and clear_stealth's declaring side
- bonus_die accepts a detection filter
- Perimeter Agency's second tags.toml entry: an extra d10 on every
  detection roll, with no limit
```

#### Commit 5 — `feat: a reroll forced on any rival's roll`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	TriggerSpeedQuery             EffectTrigger = "speed_query"
)
```
**Replace with:**
```go
	TriggerSpeedQuery             EffectTrigger = "speed_query"
	TriggerRivalRoll              EffectTrigger = "rival_roll"
)
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
		domain.TriggerAssetDestroyed, domain.TriggerSpeedQuery,
	}
```
**Replace with:**
```go
		domain.TriggerAssetDestroyed, domain.TriggerSpeedQuery, domain.TriggerRivalRoll,
	}
```

##### Task 3 — `internal/faction/engine/hooks/dispatch/roll.go`

Add `"maps"` and `"slices"` to the import block, before `"sort"`.

**Find:**
```go
		scopes = rollScopes(ctx, faction)
```
**Replace with:**
```go
		scopes = rollScopes(ctx, faction, factionState)
```

**Find:**
```go
// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, the opponent's faction, then the
// opponent's assets. Every S hook checks its owner relation itself (rollMatches),
// so a hook in the opponent's bucket acts only on rolls made against its owner.
func rollScopes(ctx hooks.RollContext, roller *domain.Faction) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, hooks.FactionScope(ctx.Opponent.ID))
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	return scopes
}
```
**Replace with:**
```go
// rollScopes lists the registry buckets a roll consults, in order: global, the
// roller's faction, the roller's assets, the opponent's faction, the opponent's
// assets, then every other faction's faction bucket in ID order. Every S hook
// checks its owner relation itself (rollMatches), so a hook in the opponent's
// bucket acts only on rolls made against its owner, and a bystander's only
// through rival_roll (Effort 3 Decision 7). A bystander's asset buckets are never
// read: only a tag holds rival_roll.
func rollScopes(ctx hooks.RollContext, roller *domain.Faction, factionState *state.FactionState) []hooks.Scope {
	scopes := []hooks.Scope{hooks.GlobalScope(), hooks.FactionScope(roller.ID)}
	scopes = append(scopes, assetScopes(roller, ctx.World, ctx.Asset)...)
	if ctx.Opponent != nil {
		scopes = append(scopes, hooks.FactionScope(ctx.Opponent.ID))
		scopes = append(scopes, assetScopes(ctx.Opponent, ctx.World, ctx.OpponentAsset)...)
	}
	if factionState == nil {
		return scopes
	}
	for _, factionID := range slices.Sorted(maps.Keys(factionState.Factions)) {
		if factionID == roller.ID || (ctx.Opponent != nil && factionID == ctx.Opponent.ID) {
			continue
		}
		scopes = append(scopes, hooks.FactionScope(factionID))
	}
	return scopes
}
```

##### Task 4 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
```
**Replace with:**
```go
// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner; rival_roll on any rival's roll, whether or not the owner is in it.
// scope = "self" ties the roll to the declaring asset (rolling it for
```

**Find:**
```go
		owner, subject = ctx.Opponent, ctx.OpponentAsset
	default:
		return false
	}
```
**Replace with:**
```go
		owner, subject = ctx.Opponent, ctx.OpponentAsset
	case domain.TriggerRivalRoll:
		// deliberate: return here, not read scope or filters, because the owner takes no part in the roll and the registrar accepts only scope faction and no filters (Effort 3 Decision 7).
		return ctx.Actor.ID != ownerFactionID
	default:
		return false
	}
```

##### Task 5 — `internal/faction/engine/effect/effects/reroll.go`

Add `"fmt"` to the import block, before `"sort"`.

**Find:**
```go
// rerollRegistrar reads own_roll as the owner rerolling its own dice, and
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll.
// owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
```
**Replace with:**
```go
// rerollRegistrar reads own_roll as the owner rerolling its own dice,
// enemy_roll_at_me as the owner forcing a rival rolling against it to reroll, and
// rival_roll as the owner forcing any rival's roll, anywhere, to reroll
// (Psychic Academy, D-7). owner and scope read as in rollMatches.
var rerollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerRivalRoll: {
			Owners: []domain.EffectOwner{domain.OwnerRival},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
			Gates:  []domain.EffectGateName{domain.GateOncePer},
		},
```

**Find:**
```go
	Register: registerReroll,
}

func registerReroll(ctx effect.RegisterContext) {
	source := ctx.Holder.Name()
	if ctx.Entry.Trigger == domain.TriggerEnemyRollAtMe {
		source += " (forced)"
	}
```
**Replace with:**
```go
	Register: registerReroll,
	Validate: validateReroll,
}

// validateReroll keeps rival_roll to a tag: the dispatcher reads a bystander's
// faction bucket only, so an asset's rival_roll hook would never reach a roll its
// faction takes no part in (Effort 3 Decision 7).
func validateReroll(entry domain.EffectDefinition, holder effect.HolderKind) error {
	if entry.Trigger == domain.TriggerRivalRoll && holder != effect.HolderTag {
		return fmt.Errorf("only a tag can hold trigger %q", domain.TriggerRivalRoll)
	}
	return nil
}

// forcedTrigger reports whether trigger forces a rival's roll: one made against
// the owner, or any rival's roll at all.
func forcedTrigger(trigger domain.EffectTrigger) bool {
	return trigger == domain.TriggerEnemyRollAtMe || trigger == domain.TriggerRivalRoll
}

func registerReroll(ctx effect.RegisterContext) {
	source := ctx.Holder.Name()
	if forcedTrigger(ctx.Entry.Trigger) {
		source += " (forced)"
	}
```

**Find:**
```go
	forced := hook.trigger == domain.TriggerEnemyRollAtMe
```
**Replace with:**
```go
	forced := forcedTrigger(hook.trigger)
```

##### Task 6 — tests

`internal/faction/engine/hooks/dispatch/roll_test.go`, append:
```go
func TestRollWithHooks_BystanderFactionHookFires(t *testing.T) {
	alpha := factionWithAssets("alpha", nil)
	beta := factionWithAssets("beta", nil)
	gamma := factionWithAssets("gamma", map[string]string{"g-1": "Krylos"})
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.FactionScope("gamma"), "gamma-tag", addDieModifier("gamma-tag", ""))
	registry.RegisterRollModifier(hooks.AssetScope("gamma", "g-1"), "g-1", addDieModifier("g-1", ""))
	collector := &recordingCollector{}

	RollWithHooks(contestCtx(alpha, beta, "Krylos"), baseRoll(), registry, collector, &fixedRoller{values: []int{5}}, alpha, rollState(alpha, beta, gamma), nil)

	if len(collector.offers) != 1 || collector.offers[0].Source != "gamma-tag" || collector.offers[0].FactionID != "gamma" {
		t.Errorf("offers = %+v, want only gamma's faction hook, stamped gamma: a bystander's assets are not read", collector.offers)
	}
}
```

`internal/faction/engine/effect/effects/reroll_test.go`, append:
```go
func psychicAcademyEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger: domain.TriggerRivalRoll,
		Effect:  domain.SpecialReroll,
		Filters: domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction},
		Gates:   domain.EffectGates{OncePer: domain.OncePerOwnerFaction},
		Reroll:  &domain.RerollParams{Dice: 1, Elective: true},
	}
}

func TestReroll_RivalRollForcesABystandersHighestDie(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-014"}, HookBudgets: map[string]int{}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}, "gamma": {ID: "gamma"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialReroll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-014", Name: "Psychic Academy"}}, Entry: psychicAcademyEntry(), Registry: registry,
	})
	collector := &rollCollector{accept: true}

	result := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "beta", "gamma", "Krylos"), 8, 2)

	if result.Sum != 2 || collector.rerolls != 1 {
		t.Errorf("Sum = %d, rerolls asked = %d; want 2 and 1: alpha forces beta's 8 to be rerolled", result.Sum, collector.rerolls)
	}
	if alpha.HookBudgets["effect:reroll:rival_roll:SWN-T-014"] != 1 {
		t.Errorf("alpha's HookBudgets = %v, want the rival_roll budget spent", alpha.HookBudgets)
	}

	own := rollThroughDispatcher(factionState, registry, collector, rollOn(factionState, "alpha", "beta", "Krylos"), 8)
	if own.Sum != 8 || collector.rerolls != 1 {
		t.Errorf("alpha's own roll: Sum = %d, rerolls asked = %d; want 8 and still 1", own.Sum, collector.rerolls)
	}
}

func TestReroll_RivalRollIsForTagsOnly(t *testing.T) {
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	asset := &domain.AssetDefinition{ID: "SWN-X", Effects: []domain.EffectDefinition{psychicAcademyEntry()}}
	if err := engine.RegisterEffectsList(asset); err == nil || !strings.Contains(err.Error(), "only a tag") {
		t.Errorf("asset holder: err = %v, want an only-a-tag refusal", err)
	}
	tag := &domain.Tag{ID: "SWN-T-014", Effects: []domain.EffectDefinition{psychicAcademyEntry()}}
	if err := engine.CheckTagEffects(tag); err != nil {
		t.Errorf("tag holder: err = %v, want none", err)
	}
}
```
Add `"log/slog"` to the `reroll_test.go` imports, before `"reflect"`.

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```
**Replace with:**
```go
		{"rival roll, a bystander's roll", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: beta, Opponent: gamma, World: "Tartarus"}, true},
		{"rival roll, against the owner", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: beta, Opponent: alpha}, true},
		{"rival roll, the owner's own roll", domain.TriggerRivalRoll, domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeFaction}, hooks.RollContext{Actor: alpha, Opponent: beta}, false},
		{"unknown trigger", domain.TriggerTurnStart, self, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
```

##### Commit message

```
feat: a reroll forced on any rival's roll

- new trigger rival_roll on reroll: the holder forces any rival's roll,
  anywhere, to reroll its highest die
- rollScopes reads every bystander faction's faction bucket after both
  sides'; rollMatches matches rival_roll on any roll the owner does not make
- only a tag holds rival_roll; Psychic Academy's entry waits for the roll
  card (D-7)
```

### Phase 4 — Goal events reach reactors

**In plain words:** the rule engine's reactors now see goal completions and
homeworld changes, so Deep Rooted drops itself when its faction moves home and
Exchange Consulate rolls for a bonus XP when it completes Peaceable Kingdom.

#### Commit 6 — `feat: Deep Rooted is lost on a homeworld change`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialSpeedFloor         SpecialEffectType = "speed_floor"
)
```
**Replace with:**
```go
	SpecialSpeedFloor         SpecialEffectType = "speed_floor"
	SpecialRemoveTag          SpecialEffectType = "remove_tag"
)
```

**Find:**
```go
	TriggerRivalRoll              EffectTrigger = "rival_roll"
)
```
**Replace with:**
```go
	TriggerRivalRoll              EffectTrigger = "rival_roll"
	TriggerHomeworldChanged       EffectTrigger = "homeworld_changed"
)
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
		domain.TriggerAssetDestroyed, domain.TriggerSpeedQuery, domain.TriggerRivalRoll,
	}
```
**Replace with:**
```go
		domain.TriggerAssetDestroyed, domain.TriggerSpeedQuery, domain.TriggerRivalRoll,
		domain.TriggerHomeworldChanged,
	}
```

**Find:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies, domain.SpecialCostDiscount, domain.SpecialSpeedFloor,
	}
```
**Replace with:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies, domain.SpecialCostDiscount, domain.SpecialSpeedFloor,
		domain.SpecialRemoveTag,
	}
```

##### Task 3 — `internal/faction/engine/effect/effects/remove_tag.go` (new)

Full contents:
```go
package effects

import (
	"fmt"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// removeTagRegistrar is Deep Rooted's loss: when the owner's homeworld changes,
// the tag that declared the entry leaves the faction (D-19). owner = "own" reads
// as "the owner's homeworld changed". Only a tag holds it: an asset has no tag of
// its own to remove.
var removeTagRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerHomeworldChanged: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerRemoveTag,
	Validate: validateRemoveTag,
}

func validateRemoveTag(_ domain.EffectDefinition, holder effect.HolderKind) error {
	if holder != effect.HolderTag {
		return fmt.Errorf("only a tag can remove itself")
	}
	return nil
}

func registerRemoveTag(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialRemoveTag), &removeTagReactor{holder: refOf(ctx.Holder)})
}

type removeTagReactor struct {
	holder holderRef
}

// OnMutations removes the tag once, on the first HomeworldChanged of its owner in
// the batch, while the owner still holds it.
func (reactor *removeTagReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	for _, mutation := range mutations {
		changed, ok := mutation.(domain.HomeworldChanged)
		if !ok || changed.FactionID != reactor.holder.factionID {
			continue
		}
		return []domain.Mutation{domain.TagRemoved{
			FactionID:         reactor.holder.factionID,
			TagID:             reactor.holder.tagID,
			Cause:             string(domain.SpecialRemoveTag),
			CausedByFactionID: reactor.holder.factionID,
		}}
	}
	return nil
}
```

##### Task 4 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialSpeedFloor:         speedFloorRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialSpeedFloor:         speedFloorRegistrar,
		domain.SpecialRemoveTag:          removeTagRegistrar,
	}
```

##### Task 5 — `internal/faction/engine/orchestrator.go`

In `runGoalLockPhase`, the reactors read the lock's mutations before they apply.

**Find:**
```go
	lock, lockMutations := e.Goal.CheckLock(faction, factionState, e.Rulebook, phaseLog.With("engine", "goal"))
	observer.OnGoalLockApplied(faction, lock, lockMutations)
```
**Replace with:**
```go
	lock, lockMutations := e.Goal.CheckLock(faction, factionState, e.Rulebook, phaseLog.With("engine", "goal"))
	if len(lockMutations) > 0 {
		var err error
		// deliberate: reactors read the lock's mutations, not only action and movement batches, because a goal completion or homeworld change is an event a rule reacts to (Effort 3 Decision 6).
		lockMutations, err = dispatch.MutationReactors(e.Hooks, faction, lockMutations, factionState, e.Rulebook)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return lock, nil
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return lock, err
		}
	}
	observer.OnGoalLockApplied(faction, lock, lockMutations)
```

In `runActionPhase`, a second reactor pass reads the goal mutations alone.

**Find:**
```go
	combined = append(combined, e.Goal.UpdateProgress(faction.ID, combined, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))...)
```
**Replace with:**
```go
	goalMutations := e.Goal.UpdateProgress(faction.ID, combined, factionState, e.Rulebook, worldIndex, phaseLog.With("engine", "goal"))
	if len(goalMutations) > 0 {
		// deliberate: a second pass over the goal mutations alone, not the whole batch again, because the first pass already reacted to the rest (Effort 3 Decision 6).
		goalMutations, err = dispatch.MutationReactors(e.Hooks, faction, goalMutations, factionState, e.Rulebook)
		if err != nil {
			if factionErrors.IsRecoverable(err) {
				phaseLog.Warn("recoverable failure", "err", err)
				observer.OnError(faction, err)
				return nil // continue past this sub-engine call
			}
			phaseLog.Error("failure", "err", err)
			observer.OnError(faction, err)
			return err
		}
	}
	combined = append(combined, goalMutations...)
```

##### Task 6 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.DeepRooted.effects.bonus_die]
sides = 10
count = 1
```
**Replace with:**
```toml
[tags.DeepRooted.effects.bonus_die]
sides = 10
count = 1

[[tags.DeepRooted.effects]]
trigger = "homeworld_changed"
effect  = "remove_tag"
[tags.DeepRooted.effects.filters]
owner = "own"
scope = "faction"
```

##### Task 7 — tests

`internal/faction/engine/effect/effects/remove_tag_test.go` (new) — full contents:
```go
package effects_test

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func deepRootedEntry() domain.EffectDefinition {
	return domain.EffectDefinition{
		Trigger: domain.TriggerHomeworldChanged,
		Effect:  domain.SpecialRemoveTag,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
	}
}

func TestRemoveTag_DropsTheTagOnItsOwnersHomeworldChange(t *testing.T) {
	alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-002"}}
	factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialRemoveTag].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-002", Name: "Deep Rooted"}}, Entry: deepRootedEntry(), Registry: registry,
	})
	reactor := registry.MutationReactorsForFaction("alpha")[0].Hook
	moved := domain.HomeworldChanged{FactionID: "alpha", ToWorld: domain.Location{WorldID: "Krylos"}}

	got := reactor.OnMutations([]domain.Mutation{moved, moved}, factionState, nil)
	want := []domain.Mutation{domain.TagRemoved{FactionID: "alpha", TagID: "SWN-T-002", Cause: "remove_tag", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("own homeworld change: got %+v, want %+v (once)", got, want)
	}
	if got := reactor.OnMutations([]domain.Mutation{domain.HomeworldChanged{FactionID: "beta"}}, factionState, nil); got != nil {
		t.Errorf("rival's homeworld change: got %+v, want none", got)
	}
	alpha.Tags = nil
	if got := reactor.OnMutations([]domain.Mutation{moved}, factionState, nil); got != nil {
		t.Errorf("tag already gone: got %+v, want none", got)
	}
}

func TestRemoveTag_IsForTagsOnly(t *testing.T) {
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	asset := &domain.AssetDefinition{ID: "SWN-X", Effects: []domain.EffectDefinition{deepRootedEntry()}}
	if err := engine.RegisterEffectsList(asset); err == nil || !strings.Contains(err.Error(), "only a tag") {
		t.Errorf("asset holder: err = %v, want an only-a-tag refusal", err)
	}
}
```

`internal/faction/engine/testharness/scenarios/tags_test.go`, append:
```go
// --- scenario N+6: Deep Rooted tag — lost when the homeworld changes ---

func TestDeepRooted_LostOnHomeworldChange(t *testing.T) {
	h := testharness.NewHarness(t)
	alpha := h.AddFaction("alpha", "Tartarus", 4, 3, 2)
	alpha.Tags = []string{"SWN-T-002"}
	alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-012", TargetWorld: domain.Location{WorldID: "Krylos"}, TurnsRemaining: 1}
	h.Collector.SelectActionFn = skipAction

	if err := h.Engine.Turn.Start(h.FactionState); err != nil {
		t.Fatalf("Turn.Start: %v", err)
	}
	if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	testharness.CheckStep(t, "alpha moved home to Krylos", alpha.Homeworld.WorldID == "Krylos",
		fmt.Sprintf("alpha.Homeworld = %q, want Krylos", alpha.Homeworld.WorldID))
	testharness.CheckStep(t, "Deep Rooted is gone", !alpha.HasTag("SWN-T-002"),
		fmt.Sprintf("alpha.Tags = %v, want no SWN-T-002", alpha.Tags))
	records := testharness.ReadHistory(t, h.Paths.HistoryPath)
	_, found := testharness.FindMutationByTypeAndCause(records, "tag_removed", "remove_tag")
	testharness.CheckStep(t, "tag_removed(cause=remove_tag) in history", found, "no tag_removed found")
}
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-002", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld, Phases: defense}, "")},
```
**Replace with:**
```go
		{"SWN-T-002", append(tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld, Phases: defense}, ""), domain.EffectDefinition{
			Trigger: domain.TriggerHomeworldChanged,
			Effect:  domain.SpecialRemoveTag,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		})},
```

##### Commit message

```
feat: Deep Rooted is lost on a homeworld change

- the reactors run over the goal-lock mutations, and again over the
  UpdateProgress output alone
- new effect remove_tag on trigger homeworld_changed: the declaring tag
  leaves its faction; only a tag holds it
- Deep Rooted's second tags.toml entry
```

#### Commit 7 — `feat: Exchange Consulate's XP roll`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialRemoveTag          SpecialEffectType = "remove_tag"
)
```
**Replace with:**
```go
	SpecialRemoveTag          SpecialEffectType = "remove_tag"
	SpecialXPRoll             SpecialEffectType = "xp_roll"
)
```

**Find:**
```go
	TriggerHomeworldChanged       EffectTrigger = "homeworld_changed"
)
```
**Replace with:**
```go
	TriggerHomeworldChanged       EffectTrigger = "homeworld_changed"
	TriggerGoalCompleted          EffectTrigger = "goal_completed"
)
```

**Find:**
```go
type SpeedFloorParams struct {
	Speed int
}
```
**Replace with:**
```go
type SpeedFloorParams struct {
	Speed int
}

// XPRollParams is a bonus-XP roll: roll Die, and on Target or more gain Amount XP.
type XPRollParams struct {
	Die    DiceRoll
	Target int
	Amount int
}
```

**Find:**
```go
	SpeedFloor     *SpeedFloorParams
}
```
**Replace with:**
```go
	SpeedFloor     *SpeedFloorParams
	XPRoll         *XPRollParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	SpeedFloor     *speedFloorRecord     `toml:"speed_floor"`
}
```
**Replace with:**
```go
	SpeedFloor     *speedFloorRecord     `toml:"speed_floor"`
	XPRoll         *xpRollRecord         `toml:"xp_roll"`
}
```

**Find:**
```go
type speedFloorRecord struct {
	Speed int `toml:"speed"`
}
```
**Replace with:**
```go
type speedFloorRecord struct {
	Speed int `toml:"speed"`
}

type xpRollRecord struct {
	Die    string `toml:"die"`
	Target int    `toml:"target"`
	Amount int    `toml:"amount"`
}
```

**Find:**
```go
		domain.TriggerHomeworldChanged,
	}
```
**Replace with:**
```go
		domain.TriggerHomeworldChanged, domain.TriggerGoalCompleted,
	}
```

**Find:**
```go
		domain.SpecialRemoveTag,
	}
```
**Replace with:**
```go
		domain.SpecialRemoveTag, domain.SpecialXPRoll,
	}
```

**Find:**
```go
	if record.SpeedFloor != nil {
		present = append(present, domain.SpecialSpeedFloor)
	}
	return present
```
**Replace with:**
```go
	if record.SpeedFloor != nil {
		present = append(present, domain.SpecialSpeedFloor)
	}
	if record.XPRoll != nil {
		present = append(present, domain.SpecialXPRoll)
	}
	return present
```

**Find:**
```go
		definition.SpeedFloor = &domain.SpeedFloorParams{Speed: record.SpeedFloor.Speed}
	}
	return nil
}
```
**Replace with:**
```go
		definition.SpeedFloor = &domain.SpeedFloorParams{Speed: record.SpeedFloor.Speed}
	case domain.SpecialXPRoll:
		if record.XPRoll == nil {
			return missingParamBlock(definition.Effect)
		}
		die, err := parseDice(record.XPRoll.Die)
		if err != nil {
			return fmt.Errorf("xp_roll die: %w", err)
		}
		if die == nil || die.NumDice == 0 {
			return fmt.Errorf("xp_roll die is required")
		}
		if record.XPRoll.Target <= 0 || record.XPRoll.Amount <= 0 {
			return fmt.Errorf("xp_roll target and amount must be > 0, got %d and %d", record.XPRoll.Target, record.XPRoll.Amount)
		}
		definition.XPRoll = &domain.XPRollParams{Die: *die, Target: record.XPRoll.Target, Amount: record.XPRoll.Amount}
	}
	return nil
}
```

##### Task 3 — `internal/faction/engine/effect/effects/xp_roll.go` (new)

Full contents:
```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// xpRollRegistrar is Exchange Consulate's bonus XP: when the owner completes a
// goal, it rolls the entry's die and gains the entry's XP on the target or more
// (D-18). owner = "own" reads as "the owner completed the goal"; goal names the
// completed goal, checked at load. No choice, so no prompt.
var xpRollRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerGoalCompleted: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeFaction},
			Filters: []domain.EffectFilterName{domain.FilterGoal},
		},
	},
	Register: registerXPRoll,
}

func registerXPRoll(ctx effect.RegisterContext) {
	ctx.Registry.RegisterMutationReactor(ctx.Holder.Scope(), ctx.Holder.HookID(domain.SpecialXPRoll), &xpRollReactor{
		holder: refOf(ctx.Holder),
		goal:   ctx.Entry.Filters.Goal,
		params: *ctx.Entry.XPRoll,
		roller: ctx.Roller,
	})
}

type xpRollReactor struct {
	holder holderRef
	goal   string
	params domain.XPRollParams
	roller domain.Roller
}

// OnMutations rolls once for each of the owner's GoalCompleted that names the
// entry's goal, or any goal when the entry names none.
func (reactor *xpRollReactor) OnMutations(mutations []domain.Mutation, factionState *state.FactionState, _ *rulebook.Rulebook) []domain.Mutation {
	if _, live := reactor.holder.live(factionState); !live {
		return nil
	}
	var awarded []domain.Mutation
	for _, mutation := range mutations {
		completed, ok := mutation.(domain.GoalCompleted)
		if !ok || completed.FactionID != reactor.holder.factionID || (reactor.goal != "" && completed.GoalID != reactor.goal) {
			continue
		}
		if reactor.params.Die.Roll(reactor.roller) < reactor.params.Target {
			continue
		}
		awarded = append(awarded, domain.XPAwarded{
			FactionID:         reactor.holder.factionID,
			Amount:            reactor.params.Amount,
			Cause:             string(domain.SpecialXPRoll),
			CausedByFactionID: reactor.holder.factionID,
		})
	}
	return awarded
}
```

##### Task 4 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialRemoveTag:          removeTagRegistrar,
	}
```
**Replace with:**
```go
		domain.SpecialRemoveTag:          removeTagRegistrar,
		domain.SpecialXPRoll:             xpRollRegistrar,
	}
```

##### Task 5 — `rulebooks/swn/tags.toml`

**Find:**
```toml
[tags.ExchangeConsulate.effects.bonus_die]
sides = 10
count = 1
```
**Replace with:**
```toml
[tags.ExchangeConsulate.effects.bonus_die]
sides = 10
count = 1

[[tags.ExchangeConsulate.effects]]
trigger = "goal_completed"
effect  = "xp_roll"
[tags.ExchangeConsulate.effects.filters]
owner = "own"
scope = "faction"
goal  = "G-007"
[tags.ExchangeConsulate.effects.xp_roll]
die    = "1d6"
target = 4
amount = 1
```

##### Task 6 — `internal/faction/rulebook/effects_test.go`

In `TestLoadEffects_LoudFailure`, after the "contest without a roll" case added in
Commit 2 Task 11:

**Find:**
```go
			wantErr: `unknown contest roll: ""`,
		},
```
**Replace with:**
```go
			wantErr: `unknown contest roll: ""`,
		},
		{
			name: "xp_roll with a flat die",
			body: `
[[assets.Machine.effects]]
trigger = "goal_completed"
effect  = "xp_roll"` + filters + `
[assets.Machine.effects.xp_roll]
die    = "4"
target = 4
amount = 1
`,
			wantErr: `xp_roll die: invalid dice notation: "4"`,
		},
		{
			name: "xp_roll with no target",
			body: `
[[assets.Machine.effects]]
trigger = "goal_completed"
effect  = "xp_roll"` + filters + `
[assets.Machine.effects.xp_roll]
die    = "1d6"
amount = 1
`,
			wantErr: "xp_roll target and amount must be > 0",
		},
```

##### Task 7 — tests

`internal/faction/engine/effect/effects/xp_roll_test.go` (new) — full contents:
```go
package effects_test

import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// exchangeConsulateReactor registers Exchange Consulate's XP roll for alpha with
// the given dice and returns its reactor.
func exchangeConsulateReactor(alpha *domain.Faction, dice ...int) hooks.MutationReactor {
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialXPRoll].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: &domain.Tag{ID: "SWN-T-004", Name: "Exchange Consulate"}},
		Entry: domain.EffectDefinition{
			Trigger: domain.TriggerGoalCompleted,
			Effect:  domain.SpecialXPRoll,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Goal: "G-007"},
			XPRoll:  &domain.XPRollParams{Die: domain.DiceRoll{NumDice: 1, Sides: 6}, Target: 4, Amount: 1},
		},
		Registry: registry,
		Roller:   &seqRoller{values: dice},
	})
	return registry.MutationReactorsForFaction("alpha")[0].Hook
}

func TestXPRoll_RollsOnTheNamedGoal(t *testing.T) {
	bonus := []domain.Mutation{domain.XPAwarded{FactionID: "alpha", Amount: 1, Cause: "xp_roll", CausedByFactionID: "alpha"}}
	cases := []struct {
		name      string
		completed domain.GoalCompleted
		dice      []int
		want      []domain.Mutation
	}{
		{"Peaceable Kingdom, 4", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-007"}, []int{4}, bonus},
		{"Peaceable Kingdom, 3", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-007"}, []int{3}, nil},
		{"another goal rolls nothing", domain.GoalCompleted{FactionID: "alpha", GoalID: "G-012"}, nil, nil},
		{"a rival's completion rolls nothing", domain.GoalCompleted{FactionID: "beta", GoalID: "G-007"}, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alpha := &domain.Faction{ID: "alpha", Tags: []string{"SWN-T-004"}}
			factionState := &state.FactionState{Factions: map[string]*domain.Faction{"alpha": alpha, "beta": {ID: "beta"}}}
			reactor := exchangeConsulateReactor(alpha, tc.dice...)

			if got := reactor.OnMutations([]domain.Mutation{tc.completed}, factionState, nil); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
```

`internal/faction/engine/testharness/scenarios/tags_test.go`, append:
```go
// --- scenario N+7: Exchange Consulate tag — bonus XP on Peaceable Kingdom ---

func TestExchangeConsulate_BonusXPOnPeaceableKingdom(t *testing.T) {
	cases := []struct {
		name   string
		roll   int
		wantXP int
	}{
		{"a 4 gains the bonus", 4, 2},
		{"a 3 does not", 3, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testharness.NewHarness(t)
			alpha := h.AddFaction("alpha", "Tartarus", 5, 3, 2)
			alpha.Tags = []string{"SWN-T-004"}
			alpha.Coin = 20
			alpha.ActiveGoal = &domain.ActiveGoal{GoalID: "G-007", Progress: 3}
			xpBefore := alpha.XP
			scriptPurchase(t, h, alpha, testharness.DefFranchise)
			h.Engine.Rand = &testharness.FixedRoller{Values: []int{tc.roll}}

			if err := h.Engine.Turn.Start(h.FactionState); err != nil {
				t.Fatalf("Turn.Start: %v", err)
			}
			if err := h.Engine.RunCycle(h.FactionState, h.Paths, h.Collectors, h.Observer); err != nil {
				t.Fatalf("RunCycle: %v", err)
			}

			testharness.CheckStep(t, "XP gained", alpha.XP-xpBefore == tc.wantXP,
				fmt.Sprintf("alpha gained %d XP, want %d (1 for the goal, 1 more on a 4+)", alpha.XP-xpBefore, tc.wantXP))
		})
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-004", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
```
**Replace with:**
```go
		{"SWN-T-004", append(tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatWealth}, domain.OncePerOwnerFaction), domain.EffectDefinition{
			Trigger: domain.TriggerGoalCompleted,
			Effect:  domain.SpecialXPRoll,
			Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Goal: "G-007"},
			XPRoll:  &domain.XPRollParams{Die: domain.DiceRoll{NumDice: 1, Sides: 6}, Target: 4, Amount: 1},
		})},
```

##### Commit message

```
feat: Exchange Consulate's XP roll

- new effect xp_roll on trigger goal_completed: roll the entry's die, and on
  its target or more gain its XP; the goal filter names the completed goal
- Exchange Consulate's second tags.toml entry: 1d6 on Peaceable Kingdom,
  4+ for 1 XP
```
