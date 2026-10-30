# Data-Driven Tags (A.001.7) — Effort 2 Plan: Bonus Dice and Simple Grants

## In Plain Words

- **What:** ten more tags work from rulebook data:
  - **Eight tag dice.** Machiavellian, Plutocratic, Deep Rooted, Theocratic,
    Exchange Consulate, Savage, Perimeter Agency and Imperialists each add an extra
    d10 to some rolls. Six need new ways to say *which* rolls: "on my homeworld",
    "against an attack made with Cunning", "with an asset that needs tech level 0",
    "against an asset that needs tech level 5", "on my Planetary Seizure target".
  - **Secretive.** Every asset its faction buys starts stealthed.
  - **Mercenary Group.** Every asset of the faction has a speed of at least 1, so
    a facility can move itself one hex a turn.
- **Why:** these ten tags do nothing in play today. After this effort, each one is
  a few lines of `tags.toml`, and a second rulebook can reuse the same filters.
- **How it fits:** Effort 2 of four in `A.001.7` (arc `A.001`), on branch
  `feature/data-driven-tags`. It builds on Effort 1's tag holder. Efforts 3 and 4
  depend on it: Effort 3 reuses the `homeworld` scope for Colonists.

## Context / Goal

- Overview: [`A.001.7-data-driven-tags-overview-plan.md`](./A.001.7-data-driven-tags-overview-plan.md).
  Decisions 1, 2, 4, 6 and 7 stand. R-4 (this session) replaces R-2's Mercenary
  Group clause. Under Decision 7, every value comes from the entry: the dice, the
  tech levels, the goal ID, the speed.
- Effort 1: [`-effort-1-plan.md`](./A.001.7-data-driven-tags-effort-1-plan.md).
  Its holder (`effect.Holder`, `holderRef`), the tag routes and the opponent's
  faction bucket on rolls are in place.
- Discovery: [`A.001.7-data-driven-tags-discovery.md`](../discovery/A.001.7-data-driven-tags-discovery.md),
  Design Summary, Effort 2; D-5, D-12, D-17.
- Architecture read: hooks, effect-mutation, world-movement.

## Decisions Ratified in Planning

1. **Theocratic and Exchange Consulate key on the attack's stat.** Robert's call
   (rules reading). "Defending against a Cunning attack" means an attack made with
   Cunning, as Machiavellian's "making a Cunning attack" does. The defense roll's
   `attribute` is the defender's own stat, which is right for Integral Protocols
   ("attacks versus Cunning") but not here. `hooks.RollContext` gains
   `OpponentAttribute`, and a new filter `opponent_attribute` reads it. A Cunning
   vs. Wealth attack gives Theocratic its die.

2. **The roll-filter vocabulary.** Robert's call (shape).
   - `opponent_attribute = "Cunning"`; `tech_level = 0`; `opponent_tech_level = 5`;
     `goal = "G-004"`.
   - Tech levels match exactly and are `*int`, as `stealthy` is `*bool`, because 0
     is a real value (Savage).
   - Planning call: each filter reads the roll, not the owner. `attribute` and
     `tech_level` read the side making the roll (`Attribute`, `Asset`);
     `opponent_attribute` and `opponent_tech_level` read the other side
     (`OpponentAttribute`, `OpponentAsset`). On `own_roll`, the only trigger
     `bonus_die` accepts, the rolling side is the owner.
   - A Base of Influence is no asset, so a Base defending gives Savage no die, and
     an attack on a Base gives Perimeter Agency none.

3. **`OpponentAttribute` is set on every roll that another roll opposes.**
   Planning call. A field set on some rolls and not others fails a filter with no
   error. Sites: the Attack action (both rolls), Expand Influence (the rival's
   contested roll, and both rolls of the free attack), and the A-side ability
   contest (both rolls). The faction's own Expand Influence roll has no opponent
   and stays empty. When Effort 3 moves the `R.023` contests onto
   `RollWithHooks`, it sets the field there too.

4. **`homeworld` is a scope.** Planning call (discovery group B; D-8 reuses it).
   It matches a roll on the owner's homeworld. Either holder may use it, because it
   reads the faction, not the declaring asset. Only `bonus_die` accepts it in this
   effort (overview R-3); Effort 3 adds it to `tech_level_floor` for Colonists.

5. **`goal` matches the target world of the owner's active goal, and is checked
   at load.** Planning call (D-5). The roll is on the world named in the active
   goal's `TargetWorld`, and the goal is the one named. A goal with no target yet
   matches nothing. `rulebook.Load` refuses an entry that names a goal the
   rulebook does not have, for assets and tags both.

6. **`rollMatches` takes the rulebook.** Planning call. The tech-level filters read
   the asset's definition. `bonus_die` and `reroll` pass the rulebook their hook
   receives. `remove_tag_dice`'s offer filter gets no rulebook, and its registrar
   accepts no tech-level filter, so it passes `nil` under a fence-sign. A
   tech-level filter with no rulebook matches nothing.

7. **Mercenary Group is a speed floor, not a movement grant.** Robert's call.
   Reverses overview R-2's "self-move mode" (overview R-4).
   - Moves in this engine are speed orders; the movement redesign replaced
     action-based self-moves with `Speed`. "Move itself to any world within one
     hex" becomes speed 1.
   - A new effect `speed_floor` on a new trigger `speed_query` (owner `own`, scope
     `faction`). It is a Cat 4 `AssetSpeedModifier`, read by a new
     `dispatch.ResolveAssetSpeed`, which walks the owner's global, faction and
     asset buckets as `ResolveWorldTechLevel` does. The entry carries the speed
     (`[speed_floor] speed = 1`, overview Decision 7). An asset with a higher speed
     keeps it.
   - `grant_movement` keeps refusing a tag; its fence-sign changes to say why.

8. **A floored asset still rides a transport.** Robert's call. Cargo eligibility
   keeps reading the definition's own speed, so a Mercenary facility may crawl one
   hex or take a transport. New guard: an asset that gets its own order in this
   decision batch is not offered as cargo. This amends the world-movement Key
   Decision "Only immobile assets can be cargo" at pre-merge.

9. **Two readers resolve speed: the tick and the movable list.** Planning call.
   `TickMovementOrders` takes the hook registry and advances each order by
   `ResolveAssetSpeed`; without that, a floored facility's order never moves.
   `eligibleMovableAssets` offers an asset whose resolved speed is above 0. `world`
   imports `hooks/dispatch`, as `relocate` does; there is no cycle.

10. **Secretive's second stealth line is accepted.** Robert's call. Psychic
    Assassins stealths itself on purchase. Bought by a Secretive faction, both
    entries fire before either applies, so history shows two `AssetStealthApplied`
    for one asset. The state is right. A fence-sign in `fireApplyStealth` names
    the case.

11. **Gates follow the text.** Once per turn (`once_per = "owner_faction"`):
    Machiavellian, Plutocratic, Theocratic, Exchange Consulate, Savage, Perimeter
    Agency. No limit: Deep Rooted (D-17), Imperialists (D-5). Phase `attack`
    covers the Expand Influence free attack and phase `defense` its defense, as
    Effort 1 Decision Record 5 read Fanatical. Every die is 1d10, keep highest.

12. **Four phases.** Planning call. Phase 1: the tags that need no new roll
    context (two data-only dice, the `homeworld` scope, Secretive). Phase 2:
    `OpponentAttribute`, which touches every opposed roll site. Phase 3: the two
    filters that read other records (definitions, goals). Phase 4: Mercenary
    Group's speed floor. Phase 4 runs about 14% past the 800-line budget; the
    excess is literal test code, and its two commits split at a seam that would
    leave a hook with no reader for a whole session.

## Decision Record — Execution

---

## Shared Context

### Branch, cadence, and checks

- Branch `feature/data-driven-tags`. Commit per session. At the effort boundary:
  interim `/code-review` of the effort's diff, `fix:` commits, then a squash into
  one `effort 2` commit (overview Shared Context).
- Every commit builds and passes: `go build ./... && go vet ./... && go test ./...`
  from the repo root. Run `gofmt -w` on each touched Go file; struct and literal
  alignment in the tasks is as gofmt leaves it.
- Edit files only with Edit/Write, never `sed` or scripts.
- No saved-state shape changes, so no sandbox re-seed.

### Pre-merge additions (beyond the overview's)

- **Graduate to Key Decisions:** Decisions 2 and 3 (hooks: roll filters read the
  roll's sides; `OpponentAttribute` on every opposed roll); Decision 7 (hooks:
  the Cat 4 speed query, owner-only walk); Decisions 8 and 9 (world-movement:
  amend "Only immobile assets can be cargo" to "by the definition's speed"; the
  tick reads the resolved speed).

## Out of Scope

- The other parts of three tags: Exchange Consulate's XP roll, Deep Rooted's
  self-removal and Perimeter Agency's detect die (Effort 3).
- `grant_movement` for a tag holder. No tag needs it after R-4.
- `faction` or `homeworld` scope on any registrar not named here (overview R-3).
- Robert's `campaigns/sandbox` files.

---

## Work Breakdown

**Overview**

- **Phase 1 — Data-only tags and the homeworld scope**
  - Commit 1 — `feat: Machiavellian and Plutocratic as rulebook data` (2 tasks)
  - Commit 2 — `feat: Deep Rooted's homeworld defense die` (6 tasks)
  - Commit 3 — `feat: Secretive as rulebook data` (3 tasks)
- **Phase 2 — The opposing stat**
  - Commit 4 — `feat: Theocratic and Exchange Consulate defend against the attack's stat` (9 tasks)
- **Phase 3 — Tech-level and goal dice**
  - Commit 5 — `feat: Savage and Perimeter Agency read each side's tech level` (7 tasks)
  - Commit 6 — `feat: Imperialists' die on its Planetary Seizure target` (7 tasks)
- **Phase 4 — Mercenary Group**
  - Commit 7 — `feat: a speed floor rule hook` (8 tasks)
  - Commit 8 — `feat: Mercenary Group as rulebook data` (5 tasks)

### Phase 1 — Data-only tags and the homeworld scope

**In plain words:** three tags get their extra die (two attack dice that need only
data, and Deep Rooted's die when defending on its homeworld), and a Secretive
faction's purchases start stealthed.

#### Commit 1 — `feat: Machiavellian and Plutocratic as rulebook data`

##### Task 1 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Once per turn, roll an additional d10 when making a Cunning attack."
```
**Replace with:**
```toml
effect = "Once per turn, roll an additional d10 when making a Cunning attack."

[[tags.Machiavellian.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Machiavellian.effects.filters]
owner     = "own"
scope     = "faction"
phase     = ["attack"]
attribute = "Cunning"
[tags.Machiavellian.effects.gates]
once_per = "owner_faction"
[tags.Machiavellian.effects.bonus_die]
sides = 10
count = 1
```

**Find:**
```toml
effect = "Once per turn, roll an additional d10 when making a Wealth attack."
```
**Replace with:**
```toml
effect = "Once per turn, roll an additional d10 when making a Wealth attack."

[[tags.Plutocratic.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Plutocratic.effects.filters]
owner     = "own"
scope     = "faction"
phase     = ["attack"]
attribute = "Wealth"
[tags.Plutocratic.effects.gates]
once_per = "owner_faction"
[tags.Plutocratic.effects.bonus_die]
sides = 10
count = 1
```

##### Task 2 — `internal/faction/rulebook/rulebook_test.go`

A table of the shipped tag dice. Commits 2, 4, 5 and 6 add rows. `TestNewWithRulebook_ShippedRulebook`
(`engine/core_test.go`) already checks that each entry passes its registrar.

**Find:**
```go
func TestLoad_TransportBypasses(t *testing.T) {
```
**Replace with:**
```go
// tagBonusDie is the shipped form of a tag's extra d10, keep highest.
func tagBonusDie(filters domain.EffectFilters, oncePer domain.OncePerKey) []domain.EffectDefinition {
	return []domain.EffectDefinition{{
		Trigger:  domain.TriggerOwnRoll,
		Effect:   domain.SpecialBonusDie,
		Filters:  filters,
		Gates:    domain.EffectGates{OncePer: oncePer},
		BonusDie: &domain.BonusDieParams{Sides: 10, Count: 1},
	}}
}

func TestLoad_TagBonusDice(t *testing.T) {
	rb := loadShippedRulebook(t)
	attack := []domain.EffectRollPhase{domain.RollPhaseAttack}
	cases := []struct {
		tagID string
		want  []domain.EffectDefinition
	}{
		{"SWN-T-007", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Attribute: domain.StatCunning}, domain.OncePerOwnerFaction)},
		{"SWN-T-012", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Attribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
	}
	for _, tc := range cases {
		if got := rb.Tags[tc.tagID].Effects; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s effects = %+v, want %+v", tc.tagID, got, tc.want)
		}
	}
}

func TestLoad_TransportBypasses(t *testing.T) {
```

##### Commit message

```
feat: Machiavellian and Plutocratic as rulebook data

- Machiavellian: +1d10 once per turn on its Cunning attacks
- Plutocratic: +1d10 once per turn on its Wealth attacks
- TestLoad_TagBonusDice checks the shipped tag dice
```

#### Commit 2 — `feat: Deep Rooted's homeworld defense die`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
const (
	ScopeSelf    EffectScope = "self"
	ScopeWorld   EffectScope = "world"
	ScopeFaction EffectScope = "faction"
)
```
**Replace with:**
```go
const (
	ScopeSelf      EffectScope = "self"
	ScopeWorld     EffectScope = "world"
	ScopeHomeworld EffectScope = "homeworld"
	ScopeFaction   EffectScope = "faction"
)
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction}
```
**Replace with:**
```go
	effectScopes     = []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeHomeworld, domain.ScopeFaction}
```

##### Task 3 — `internal/faction/engine/effect/effects/roll_match.go`

The trigger switch now also names the owner faction, which the `homeworld` scope
reads. Full contents:

```go
package effects

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)

var rollPhaseNames = map[hooks.RollPhase]domain.EffectRollPhase{
	hooks.PhaseAttack:      domain.RollPhaseAttack,
	hooks.PhaseDefense:     domain.RollPhaseDefense,
	hooks.PhaseFactionTest: domain.RollPhaseFactionTest,
	hooks.PhaseContested:   domain.RollPhaseContested,
}

// rollMatches reports whether a roll fires an entry declared on asset. own_roll
// fires on the owner's own roll; enemy_roll_at_me on a rival's roll made against
// the owner. scope = "self" ties the roll to the declaring asset (rolling it for
// own_roll, rolled against for enemy_roll_at_me); scope = "world" to any such roll
// on the asset's world; scope = "homeworld" to any such roll on the owner's
// homeworld; scope = "faction" to any such roll, so asset is nil for a tag. The
// dispatcher reaches both sides, so this check, not the registration scope, is
// what keeps an entry to its own side.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters) bool {
	if ctx.Actor == nil {
		return false
	}
	var owner *domain.Faction
	var subject *domain.Asset
	switch trigger {
	case domain.TriggerOwnRoll:
		if ctx.Actor.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Actor, ctx.Asset
	case domain.TriggerEnemyRollAtMe:
		if ctx.Actor.ID == ownerFactionID || ctx.Opponent == nil || ctx.Opponent.ID != ownerFactionID {
			return false
		}
		owner, subject = ctx.Opponent, ctx.OpponentAsset
	default:
		return false
	}
	switch filters.Scope {
	case domain.ScopeSelf:
		if subject == nil || subject.ID != asset.ID {
			return false
		}
	case domain.ScopeWorld:
		if ctx.World == "" || ctx.World != asset.Location.WorldID {
			return false
		}
	case domain.ScopeHomeworld:
		if ctx.World == "" || ctx.World != owner.Homeworld.WorldID {
			return false
		}
	case domain.ScopeFaction:
	default:
		return false
	}
	if filters.Attribute != "" && ctx.Attribute != string(filters.Attribute) {
		return false
	}
	if len(filters.Phases) > 0 && !slices.Contains(filters.Phases, rollPhaseNames[ctx.Phase]) {
		return false
	}
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	return true
}
```

##### Task 4 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world;
// "faction" any owner roll anywhere. A tag's die is FromTag, so Blackmail strips it.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeFaction},
```
**Replace with:**
```go
// bonusDieRegistrar reads owner = "own" as "the owner is rolling". scope =
// "self" is a roll by the declaring asset; "world" any owner roll on its world;
// "homeworld" any owner roll on the owner's homeworld; "faction" any owner roll
// anywhere. A tag's die is FromTag, so Blackmail strips it.
var bonusDieRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOwnRoll: {
			Owners:  []domain.EffectOwner{domain.OwnerOwn},
			Scopes:  []domain.EffectScope{domain.ScopeSelf, domain.ScopeWorld, domain.ScopeHomeworld, domain.ScopeFaction},
```

##### Task 5 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Roll one additional d10 when defending against attacks on assets on their homeworld. Lost if the faction ever changes homeworlds."
```
**Replace with:**
```toml
effect = "Roll one additional d10 when defending against attacks on assets on their homeworld. Lost if the faction ever changes homeworlds."

[[tags.DeepRooted.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.DeepRooted.effects.filters]
owner = "own"
scope = "homeworld"
phase = ["defense"]
[tags.DeepRooted.effects.bonus_die]
sides = 10
count = 1
```

##### Task 6 — tests

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
	alpha := &domain.Faction{ID: "alpha"}
```
**Replace with:**
```go
	alpha := &domain.Faction{ID: "alpha", Homeworld: domain.Location{WorldID: "Krylos"}}
```

**Find:**
```go
	expansion := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, Action: "Expand Influence"}
```
**Replace with:**
```go
	expansion := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeWorld, Action: "Expand Influence"}
	homeworld := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld}
	rivalHomeworld := domain.EffectFilters{Owner: domain.OwnerRival, Scope: domain.ScopeHomeworld}
```

**Find:**
```go
		{"action mismatch", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Attack"}, false},
```
**Replace with:**
```go
		{"action mismatch", domain.TriggerOwnRoll, expansion, hooks.RollContext{Actor: alpha, World: "Krylos", Action: "Attack"}, false},
		{"own roll, homeworld", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: "Krylos"}, true},
		{"own roll, off homeworld", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: "Tartarus"}, false},
		{"own roll, homeworld scope in flight", domain.TriggerOwnRoll, homeworld, hooks.RollContext{Actor: alpha, World: ""}, false},
		{"enemy at me, the owner's homeworld", domain.TriggerEnemyRollAtMe, rivalHomeworld, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, true},
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
	attack := []domain.EffectRollPhase{domain.RollPhaseAttack}
```
**Replace with:**
```go
	attack := []domain.EffectRollPhase{domain.RollPhaseAttack}
	defense := []domain.EffectRollPhase{domain.RollPhaseDefense}
```

**Find:**
```go
		{"SWN-T-012", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Attribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
```
**Replace with:**
```go
		{"SWN-T-012", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Attribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
		{"SWN-T-002", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld, Phases: defense}, "")},
```

##### Commit message

```
feat: Deep Rooted's homeworld defense die

- New effect scope homeworld: a roll on the owner's homeworld
- rollMatches names the owner faction for the scope to read
- bonus_die accepts homeworld; Deep Rooted defends there with +1d10
```

#### Commit 3 — `feat: Secretive as rulebook data`

##### Task 1 — `internal/faction/engine/effect/effects/apply_stealth.go`

Full contents:

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// applyStealthRegistrar reads owner = "own" as "the holder's faction bought the
// asset". scope = "self" is the declaring asset's own purchase; scope = "faction"
// is every asset the faction buys, which serves a tag (Secretive): an asset's
// on_purchase entries fire on its own purchase only.
var applyStealthRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerOnPurchase: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeSelf, domain.ScopeFaction},
		},
	},
	Fire: fireApplyStealth,
}

func fireApplyStealth(ctx effect.FireContext) []domain.Mutation {
	// deliberate: a second apply_stealth on one purchase emits again, not nothing, because both fire before either applies and the repeat changes no state (Effort 2 Decision 10).
	if ctx.Asset.Stealthy {
		return nil
	}
	return []domain.Mutation{domain.AssetStealthApplied{
		FactionID:         ctx.Holder.Faction.ID,
		AssetID:           ctx.Asset.ID,
		Cause:             string(domain.SpecialApplyStealth),
		CausedByFactionID: ctx.Holder.Faction.ID,
	}}
}
```

##### Task 2 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "All assets purchased by this faction automatically begin Stealthed."
```
**Replace with:**
```toml
effect = "All assets purchased by this faction automatically begin Stealthed."

[[tags.Secretive.effects]]
trigger = "on_purchase"
effect  = "apply_stealth"
[tags.Secretive.effects.filters]
owner = "own"
scope = "faction"
```

##### Task 3 — tests

`internal/faction/engine/effect/effects/apply_stealth_test.go`:

**Find:**
```go
import (
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
)
```
**Replace with:**
```go
import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)
```

Appended at the end of the file:

```go
// TestApplyStealth_SecretiveStealthsEveryPurchase: a tag entry at faction scope
// passes the engine's check and stealths any asset its faction buys.
func TestApplyStealth_SecretiveStealthsEveryPurchase(t *testing.T) {
	secretive := &domain.Tag{ID: "SWN-T-017", Name: "Secretive", Effects: []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialApplyStealth,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
	}}}
	engine := effect.New(effects.Registrars(), slog.New(slog.DiscardHandler))
	if err := engine.CheckTagEffects(secretive); err != nil {
		t.Fatalf("CheckTagEffects: %v", err)
	}
	rb := &rulebook.Rulebook{
		Assets: map[string]*domain.AssetDefinition{"SWN-W1-001": {ID: "SWN-W1-001"}},
		Tags:   map[string]*domain.Tag{secretive.ID: secretive},
	}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{secretive.ID}}
	bought := &domain.Asset{ID: "franchise", DefinitionID: "SWN-W1-001", OwnerID: "alpha"}

	got := engine.FirePurchase(alpha, bought, nil, rb, nil)

	want := []domain.Mutation{domain.AssetStealthApplied{FactionID: "alpha", AssetID: "franchise", Cause: "apply_stealth", CausedByFactionID: "alpha"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutations = %+v, want %+v", got, want)
	}
}
```

`internal/faction/rulebook/rulebook_test.go`:

**Find:**
```go
func TestLoad_TransportBypasses(t *testing.T) {
```
**Replace with:**
```go
func TestLoad_SecretiveStealthsPurchases(t *testing.T) {
	rb := loadShippedRulebook(t)
	want := []domain.EffectDefinition{{
		Trigger: domain.TriggerOnPurchase,
		Effect:  domain.SpecialApplyStealth,
		Filters: domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
	}}
	if got := rb.Tags["SWN-T-017"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Secretive effects = %+v, want %+v", got, want)
	}
}

func TestLoad_TransportBypasses(t *testing.T) {
```

##### Commit message

```
feat: Secretive as rulebook data

- apply_stealth accepts faction scope, so a tag stealths every purchase
- Secretive: every asset its faction buys begins Stealthed (D-12)
- A second stealth line on one purchase is accepted (Effort 2 Decision 10)
```

### Phase 2 — The opposing stat

**In plain words:** every roll that another roll opposes now also says which stat
the other side used, so Theocratic and Exchange Consulate can add their die when
defending against a Cunning or a Wealth attack.

#### Commit 4 — `feat: Theocratic and Exchange Consulate defend against the attack's stat`

##### Task 1 — `internal/faction/engine/hooks/types.go`

**Find:**
```go
// RollContext carries the full context of a roll being processed by hooks.
// Asset is the rolling asset and OpponentAsset the asset on the other side;
// either is nil when no asset stands there (a faction test, a Base defending).
// Action is the action's Name().
type RollContext struct {
	Phase         RollPhase
	Actor         *domain.Faction
	Opponent      *domain.Faction
	Attribute     string
	Asset         *domain.Asset
	OpponentAsset *domain.Asset
	World         string
	Action        string
}
```
**Replace with:**
```go
// RollContext carries the full context of a roll being processed by hooks.
// Attribute is the stat this roll uses and OpponentAttribute the stat the other
// side rolls against it; OpponentAttribute is empty when no roll opposes this one
// (a faction test). Asset is the rolling asset and OpponentAsset the asset on the
// other side; either is nil when no asset stands there (a faction test, a Base
// defending). Action is the action's Name().
type RollContext struct {
	Phase             RollPhase
	Actor             *domain.Faction
	Opponent          *domain.Faction
	Attribute         string
	OpponentAttribute string
	Asset             *domain.Asset
	OpponentAsset     *domain.Asset
	World             string
	Action            string
}
```

##### Task 2 — `internal/faction/engine/action/actions/attack.go`

**Find:**
```go
			hooks.RollContext{
				Phase:         hooks.PhaseAttack,
				Actor:         faction,
				Opponent:      defenderFaction,
				Attribute:     string(attackerDef.Attack.AttackerStat),
				Asset:         attacker,
				OpponentAsset: defender,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:             hooks.PhaseAttack,
				Actor:             faction,
				Opponent:          defenderFaction,
				Attribute:         string(attackerDef.Attack.AttackerStat),
				OpponentAttribute: string(attackerDef.Attack.DefenderStat),
				Asset:             attacker,
				OpponentAsset:     defender,
				World:             attacker.Location.WorldID,
				Action:            attack.Name(),
			},
```

**Find:**
```go
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         defenderFaction,
				Opponent:      faction,
				Attribute:     string(attackerDef.Attack.DefenderStat),
				Asset:         defender,
				OpponentAsset: attacker,
				World:         attacker.Location.WorldID,
				Action:        attack.Name(),
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:             hooks.PhaseDefense,
				Actor:             defenderFaction,
				Opponent:          faction,
				Attribute:         string(attackerDef.Attack.DefenderStat),
				OpponentAttribute: string(attackerDef.Attack.AttackerStat),
				Asset:             defender,
				OpponentAsset:     attacker,
				World:             attacker.Location.WorldID,
				Action:            attack.Name(),
			},
```

##### Task 3 — `internal/faction/engine/action/actions/expand_influence.go`

The rival's contested roll answers the faction's Cunning roll.

**Find:**
```go
			hooks.RollContext{
				Phase:     hooks.PhaseContested,
				Actor:     rival,
				Opponent:  faction,
				Attribute: string(domain.StatCunning),
				World:     ei.order.World,
				Action:    ei.Name(),
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:             hooks.PhaseContested,
				Actor:             rival,
				Opponent:          faction,
				Attribute:         string(domain.StatCunning),
				OpponentAttribute: string(domain.StatCunning),
				World:             ei.order.World,
				Action:            ei.Name(),
			},
```

The free attack: the Base defends with Cunning.

**Find:**
```go
			hooks.RollContext{
				Phase:     hooks.PhaseAttack,
				Actor:     rival,
				Opponent:  ownerFaction,
				Attribute: string(attackerDef.Attack.AttackerStat),
				Asset:     attacker,
				World:     worldID,
				Action:    ba.action,
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:             hooks.PhaseAttack,
				Actor:             rival,
				Opponent:          ownerFaction,
				Attribute:         string(attackerDef.Attack.AttackerStat),
				OpponentAttribute: string(domain.StatCunning),
				Asset:             attacker,
				World:             worldID,
				Action:            ba.action,
			},
```

**Find:**
```go
			hooks.RollContext{
				Phase:         hooks.PhaseDefense,
				Actor:         ownerFaction,
				Opponent:      rival,
				Attribute:     string(domain.StatCunning),
				OpponentAsset: attacker,
				World:         worldID,
				Action:        ba.action,
			},
```
**Replace with:**
```go
			hooks.RollContext{
				Phase:             hooks.PhaseDefense,
				Actor:             ownerFaction,
				Opponent:          rival,
				Attribute:         string(domain.StatCunning),
				OpponentAttribute: string(attackerDef.Attack.AttackerStat),
				OpponentAsset:     attacker,
				World:             worldID,
				Action:            ba.action,
			},
```

##### Task 4 — `internal/faction/engine/action/actions/ability/stage.go`

**Find:**
```go
		hooks.RollContext{
			Phase:         hooks.PhaseFactionTest,
			Actor:         faction,
			Opponent:      defender,
			Attribute:     string(step.Contest.AttackerStat),
			Asset:         asset,
			OpponentAsset: target.Asset,
			World:         worldID,
			Action:        actionName,
		},
```
**Replace with:**
```go
		hooks.RollContext{
			Phase:             hooks.PhaseFactionTest,
			Actor:             faction,
			Opponent:          defender,
			Attribute:         string(step.Contest.AttackerStat),
			OpponentAttribute: string(step.Contest.DefenderStat),
			Asset:             asset,
			OpponentAsset:     target.Asset,
			World:             worldID,
			Action:            actionName,
		},
```

**Find:**
```go
		hooks.RollContext{
			Phase:         hooks.PhaseContested,
			Actor:         defender,
			Opponent:      faction,
			Attribute:     string(step.Contest.DefenderStat),
			Asset:         target.Asset,
			OpponentAsset: asset,
			World:         worldID,
			Action:        actionName,
		},
```
**Replace with:**
```go
		hooks.RollContext{
			Phase:             hooks.PhaseContested,
			Actor:             defender,
			Opponent:          faction,
			Attribute:         string(step.Contest.DefenderStat),
			OpponentAttribute: string(step.Contest.AttackerStat),
			Asset:             target.Asset,
			OpponentAsset:     asset,
			World:             worldID,
			Action:            actionName,
		},
```

##### Task 5 — `internal/faction/domain/effect.go`

**Find:**
```go
const (
	FilterAttribute    EffectFilterName = "attribute"
	FilterPhase        EffectFilterName = "phase"
	FilterAction       EffectFilterName = "action"
	FilterAssetType    EffectFilterName = "asset_type"
	FilterCategory     EffectFilterName = "category"
	FilterStealthy     EffectFilterName = "stealthy"
	FilterMinTechLevel EffectFilterName = "min_tech_level"
	FilterCausedBy     EffectFilterName = "caused_by"
)
```
**Replace with:**
```go
const (
	FilterAttribute         EffectFilterName = "attribute"
	FilterOpponentAttribute EffectFilterName = "opponent_attribute"
	FilterPhase             EffectFilterName = "phase"
	FilterAction            EffectFilterName = "action"
	FilterAssetType         EffectFilterName = "asset_type"
	FilterCategory          EffectFilterName = "category"
	FilterStealthy          EffectFilterName = "stealthy"
	FilterMinTechLevel      EffectFilterName = "min_tech_level"
	FilterCausedBy          EffectFilterName = "caused_by"
)
```

**Find:**
```go
type EffectFilters struct {
	Owner        EffectOwner
	Scope        EffectScope
	Attribute    FactionStat
	Phases       []EffectRollPhase
	Action       string
	AssetTypes   []AssetType
	Categories   []FactionStat
	Stealthy     *bool
	MinTechLevel int
	CausedBy     EffectOwner
}

func (filters EffectFilters) Used() []EffectFilterName {
	var present []EffectFilterName
	if filters.Attribute != "" {
		present = append(present, FilterAttribute)
	}
```
**Replace with:**
```go
type EffectFilters struct {
	Owner             EffectOwner
	Scope             EffectScope
	Attribute         FactionStat
	OpponentAttribute FactionStat
	Phases            []EffectRollPhase
	Action            string
	AssetTypes        []AssetType
	Categories        []FactionStat
	Stealthy          *bool
	MinTechLevel      int
	CausedBy          EffectOwner
}

func (filters EffectFilters) Used() []EffectFilterName {
	var present []EffectFilterName
	if filters.Attribute != "" {
		present = append(present, FilterAttribute)
	}
	if filters.OpponentAttribute != "" {
		present = append(present, FilterOpponentAttribute)
	}
```

##### Task 6 — `internal/faction/rulebook/effects.go`

**Find:**
```go
type effectFiltersRecord struct {
	Owner        string   `toml:"owner"`
	Scope        string   `toml:"scope"`
	Attribute    string   `toml:"attribute"`
	Phase        []string `toml:"phase"`
	Action       string   `toml:"action"`
	AssetType    []string `toml:"asset_type"`
	Category     []string `toml:"category"`
	Stealthy     *bool    `toml:"stealthy"`
	MinTechLevel int      `toml:"min_tech_level"`
	CausedBy     string   `toml:"caused_by"`
}
```
**Replace with:**
```go
type effectFiltersRecord struct {
	Owner             string   `toml:"owner"`
	Scope             string   `toml:"scope"`
	Attribute         string   `toml:"attribute"`
	OpponentAttribute string   `toml:"opponent_attribute"`
	Phase             []string `toml:"phase"`
	Action            string   `toml:"action"`
	AssetType         []string `toml:"asset_type"`
	Category          []string `toml:"category"`
	Stealthy          *bool    `toml:"stealthy"`
	MinTechLevel      int      `toml:"min_tech_level"`
	CausedBy          string   `toml:"caused_by"`
}
```

In `convertEffectFilters`:

**Find:**
```go
		filters.Attribute = attribute
	}
```
**Replace with:**
```go
		filters.Attribute = attribute
	}
	if record.OpponentAttribute != "" {
		opponentAttribute, err := toFactionStat(record.OpponentAttribute)
		if err != nil {
			return domain.EffectFilters{}, fmt.Errorf("filters.opponent_attribute: %w", err)
		}
		filters.OpponentAttribute = opponentAttribute
	}
```

##### Task 7 — `internal/faction/engine/effect/effects/roll_match.go` and `bonus_die.go`

`roll_match.go`:

**Find:**
```go
// dispatcher reaches both sides, so this check, not the registration scope, is
// what keeps an entry to its own side.
```
**Replace with:**
```go
// dispatcher reaches both sides, so this check, not the registration scope, is
// what keeps an entry to its own side. attribute reads the stat this roll uses;
// opponent_attribute the stat the other side rolls against it.
```

**Find:**
```go
	if filters.Attribute != "" && ctx.Attribute != string(filters.Attribute) {
		return false
	}
```
**Replace with:**
```go
	if filters.Attribute != "" && ctx.Attribute != string(filters.Attribute) {
		return false
	}
	if filters.OpponentAttribute != "" && ctx.OpponentAttribute != string(filters.OpponentAttribute) {
		return false
	}
```

`bonus_die.go`:

**Find:**
```go
			Filters: []domain.EffectFilterName{domain.FilterAttribute, domain.FilterPhase, domain.FilterAction},
```
**Replace with:**
```go
			Filters: []domain.EffectFilterName{
				domain.FilterAttribute, domain.FilterOpponentAttribute, domain.FilterPhase, domain.FilterAction,
			},
```

##### Task 8 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "When completing a Peaceable Kingdom goal, roll 1d6; on 4+ gain a bonus XP. Once per turn, roll an extra d10 when defending against a Wealth attack."
```
**Replace with:**
```toml
effect = "When completing a Peaceable Kingdom goal, roll 1d6; on 4+ gain a bonus XP. Once per turn, roll an extra d10 when defending against a Wealth attack."

[[tags.ExchangeConsulate.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.ExchangeConsulate.effects.filters]
owner              = "own"
scope              = "faction"
phase              = ["defense"]
opponent_attribute = "Wealth"
[tags.ExchangeConsulate.effects.gates]
once_per = "owner_faction"
[tags.ExchangeConsulate.effects.bonus_die]
sides = 10
count = 1
```

**Find:**
```toml
effect = "Once per turn, roll an extra d10 when defending against a Cunning attack."
```
**Replace with:**
```toml
effect = "Once per turn, roll an extra d10 when defending against a Cunning attack."

[[tags.Theocratic.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Theocratic.effects.filters]
owner              = "own"
scope              = "faction"
phase              = ["defense"]
opponent_attribute = "Cunning"
[tags.Theocratic.effects.gates]
once_per = "owner_faction"
[tags.Theocratic.effects.bonus_die]
sides = 10
count = 1
```

##### Task 9 — tests

`internal/faction/engine/action/actions/expand_influence_test.go`, in
`TestExpandInfluence_RollsCarryContext`:

**Find:**
```go
	want := []hooks.RollContext{
		{Phase: hooks.PhaseFactionTest, Actor: faction, Attribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseContested, Actor: rival, Opponent: faction, Attribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseAttack, Actor: rival, Opponent: faction, Attribute: string(domain.StatForce), Asset: f2Asset},
		{Phase: hooks.PhaseDefense, Actor: faction, Opponent: rival, Attribute: string(domain.StatCunning), OpponentAsset: f2Asset},
	}
```
**Replace with:**
```go
	want := []hooks.RollContext{
		{Phase: hooks.PhaseFactionTest, Actor: faction, Attribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseContested, Actor: rival, Opponent: faction, Attribute: string(domain.StatCunning), OpponentAttribute: string(domain.StatCunning)},
		{Phase: hooks.PhaseAttack, Actor: rival, Opponent: faction, Attribute: string(domain.StatForce), OpponentAttribute: string(domain.StatCunning), Asset: f2Asset},
		{Phase: hooks.PhaseDefense, Actor: faction, Opponent: rival, Attribute: string(domain.StatCunning), OpponentAttribute: string(domain.StatForce), OpponentAsset: f2Asset},
	}
```

`internal/faction/engine/action/actions/attack_test.go`, appended at the end of the
file. It reuses `contextRecorder` from `expand_influence_test.go` (same package).

```go
// TestAttack_RollsNameTheOpposingStat: each side's roll carries the stat the other
// side rolls. The attacker's profile is Force vs. Cunning so the two differ. Rolls
// [2, 9]: the defender wins and has no counter, so no third roll.
func TestAttack_RollsNameTheOpposingStat(t *testing.T) {
	rulebook := makeAttackRulebook()
	rulebook.Assets["force-attacker"].Attack.DefenderStat = domain.StatCunning
	factionState, attackerAsset, defenderAsset := makeAttackState(8, 8, "force-defender-nocounter", false)
	faction := factionState.Factions["f1"]

	recorder := &contextRecorder{}
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "test:recorder", recorder)

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAttackers(gomock.Any(), gomock.Any()).Return([]*domain.Asset{attackerAsset}, nil)
	collector.EXPECT().SelectDefender(gomock.Any(), gomock.Any(), gomock.Any()).Return(defenderAsset, nil)

	runAttack(t, collector, &fixedRoller{values: []int{2, 9}}, faction, factionState, rulebook, registry)

	if len(recorder.contexts) != 2 {
		t.Fatalf("recorded %d roll contexts, want 2: %+v", len(recorder.contexts), recorder.contexts)
	}
	attack, defense := recorder.contexts[0], recorder.contexts[1]
	if attack.Attribute != "Force" || attack.OpponentAttribute != "Cunning" {
		t.Errorf("attack roll stats = %q vs %q, want Force vs Cunning", attack.Attribute, attack.OpponentAttribute)
	}
	if defense.Attribute != "Cunning" || defense.OpponentAttribute != "Force" {
		t.Errorf("defense roll stats = %q vs %q, want Cunning vs Force", defense.Attribute, defense.OpponentAttribute)
	}
}
```

`internal/faction/engine/action/actions/ability/dispatch_test.go`, appended at the
end of the file. Rolls [3, 8] are `TestDispatch_RevealStealth_DefenderWins`'s.

```go
// contextLog records every RollContext a roll shows its hooks, and offers nothing.
type contextLog struct{ contexts []hooks.RollContext }

func (log *contextLog) OfferModifiers(ctx hooks.RollContext, _ *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	log.contexts = append(log.contexts, ctx)
	return nil
}

// TestDispatch_ContestRollsNameTheOpposingStat: each contest roll carries the stat
// the other side rolls.
func TestDispatch_ContestRollsNameTheOpposingStat(t *testing.T) {
	faction, asset, def, factionState := makeDispatchFixture()
	targetFaction := factionState.Factions["f2"]

	ctrl := gomock.NewController(t)
	collector := mocks.NewMockCollector(ctrl)
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)

	recorder := &contextLog{}
	registry := hooks.NewRegistry()
	registry.RegisterRollModifier(hooks.GlobalScope(), "test:recorder", recorder)

	if _, err := Dispatch(faction, asset, def, collector, &seqRoller{values: []int{3, 8}}, registry, factionState, NewView(factionState), nil, nil); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(recorder.contexts) != 2 {
		t.Fatalf("recorded %d roll contexts, want 2: %+v", len(recorder.contexts), recorder.contexts)
	}
	attack, defense := recorder.contexts[0], recorder.contexts[1]
	if attack.OpponentAttribute == "" || attack.OpponentAttribute != defense.Attribute || defense.OpponentAttribute != attack.Attribute {
		t.Errorf("contest stats: attack %q vs %q, defense %q vs %q; want each roll to name the other's stat",
			attack.Attribute, attack.OpponentAttribute, defense.Attribute, defense.OpponentAttribute)
	}
}
```

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
	homeworld := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld}
```
**Replace with:**
```go
	homeworld := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld}
	againstCunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentAttribute: domain.StatCunning}
```

**Find:**
```go
		{"enemy at me, the owner's homeworld", domain.TriggerEnemyRollAtMe, rivalHomeworld, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, true},
```
**Replace with:**
```go
		{"enemy at me, the owner's homeworld", domain.TriggerEnemyRollAtMe, rivalHomeworld, hooks.RollContext{Actor: beta, Opponent: alpha, World: "Krylos"}, true},
		{"opponent attribute match", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Wealth", OpponentAttribute: "Cunning"}, true},
		{"opponent attribute mismatch", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Cunning", OpponentAttribute: "Wealth"}, false},
```

`internal/faction/rulebook/effects_test.go`, appended at the end of the file:

```go
func TestConvertEffectFilters_OpponentAttribute(t *testing.T) {
	filters, err := convertEffectFilters(effectFiltersRecord{Owner: "own", Scope: "faction", OpponentAttribute: "Cunning"})
	if err != nil || filters.OpponentAttribute != domain.StatCunning {
		t.Errorf("filters = %+v, err = %v; want opponent attribute Cunning", filters, err)
	}
	if !reflect.DeepEqual(filters.Used(), []domain.EffectFilterName{domain.FilterOpponentAttribute}) {
		t.Errorf("Used() = %v, want [opponent_attribute]", filters.Used())
	}
	_, err = convertEffectFilters(effectFiltersRecord{Owner: "own", Scope: "faction", OpponentAttribute: "Charm"})
	if err == nil || !strings.Contains(err.Error(), "filters.opponent_attribute") {
		t.Errorf("err = %v, want a filters.opponent_attribute error", err)
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-002", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld, Phases: defense}, "")},
```
**Replace with:**
```go
		{"SWN-T-002", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeHomeworld, Phases: defense}, "")},
		{"SWN-T-019", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatCunning}, domain.OncePerOwnerFaction)},
		{"SWN-T-004", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
```

##### Commit message

```
feat: Theocratic and Exchange Consulate defend against the attack's stat

- RollContext.OpponentAttribute: the stat the other side rolls, set on every
  opposed roll (Attack, Expand Influence, the ability contest)
- New filter opponent_attribute, accepted by bonus_die
- Theocratic: +1d10 once per turn defending against a Cunning attack
- Exchange Consulate: +1d10 once per turn defending against a Wealth attack
```

### Phase 3 — Tech-level and goal dice

**In plain words:** three tags whose die depends on another record: Savage and
Perimeter Agency read an asset's required tech level from its definition, and
Imperialists reads its active goal. A goal that the rulebook does not have now
fails the load.

#### Commit 5 — `feat: Savage and Perimeter Agency read each side's tech level`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	FilterCausedBy          EffectFilterName = "caused_by"
)
```
**Replace with:**
```go
	FilterCausedBy          EffectFilterName = "caused_by"
	FilterTechLevel         EffectFilterName = "tech_level"
	FilterOpponentTechLevel EffectFilterName = "opponent_tech_level"
)
```

**Find:**
```go
// EffectFilters narrows which events fire an entry. A zero optional field is an
// absent filter. Stealthy is a pointer because "stealthy = false" is a real
// filter, distinct from no filter.
```
**Replace with:**
```go
// EffectFilters narrows which events fire an entry. A zero optional field is an
// absent filter. Stealthy, TechLevel and OpponentTechLevel are pointers because
// "stealthy = false" and "tech_level = 0" are real filters, distinct from no
// filter.
```

**Find:**
```go
	MinTechLevel      int
	CausedBy          EffectOwner
}
```
**Replace with:**
```go
	MinTechLevel      int
	TechLevel         *int
	OpponentTechLevel *int
	CausedBy          EffectOwner
}
```

**Find:**
```go
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
```
**Replace with:**
```go
	if filters.MinTechLevel > 0 {
		present = append(present, FilterMinTechLevel)
	}
	if filters.TechLevel != nil {
		present = append(present, FilterTechLevel)
	}
	if filters.OpponentTechLevel != nil {
		present = append(present, FilterOpponentTechLevel)
	}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	MinTechLevel      int      `toml:"min_tech_level"`
	CausedBy          string   `toml:"caused_by"`
}
```
**Replace with:**
```go
	MinTechLevel      int      `toml:"min_tech_level"`
	TechLevel         *int     `toml:"tech_level"`
	OpponentTechLevel *int     `toml:"opponent_tech_level"`
	CausedBy          string   `toml:"caused_by"`
}
```

**Find:**
```go
	if record.MinTechLevel < 0 {
		return domain.EffectFilters{}, fmt.Errorf("filters.min_tech_level must be >= 0, got %d", record.MinTechLevel)
	}
	filters := domain.EffectFilters{Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy, MinTechLevel: record.MinTechLevel}
```
**Replace with:**
```go
	if record.MinTechLevel < 0 {
		return domain.EffectFilters{}, fmt.Errorf("filters.min_tech_level must be >= 0, got %d", record.MinTechLevel)
	}
	if record.TechLevel != nil && *record.TechLevel < 0 {
		return domain.EffectFilters{}, fmt.Errorf("filters.tech_level must be >= 0, got %d", *record.TechLevel)
	}
	if record.OpponentTechLevel != nil && *record.OpponentTechLevel < 0 {
		return domain.EffectFilters{}, fmt.Errorf("filters.opponent_tech_level must be >= 0, got %d", *record.OpponentTechLevel)
	}
	filters := domain.EffectFilters{
		Owner: owner, Scope: scope, Action: record.Action, Stealthy: record.Stealthy,
		MinTechLevel: record.MinTechLevel, TechLevel: record.TechLevel, OpponentTechLevel: record.OpponentTechLevel,
	}
```

##### Task 3 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)
```

**Find:**
```go
// opponent_attribute the stat the other side rolls against it.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters) bool {
```
**Replace with:**
```go
// opponent_attribute the stat the other side rolls against it. tech_level reads
// the rolling asset's definition, opponent_tech_level the other side's; neither
// matches without a rulebook.
func rollMatches(ctx hooks.RollContext, trigger domain.EffectTrigger, ownerFactionID string, asset *domain.Asset, filters domain.EffectFilters, rulebook *rulebook.Rulebook) bool {
```

**Find:**
```go
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	return true
}
```
**Replace with:**
```go
	if filters.Action != "" && ctx.Action != filters.Action {
		return false
	}
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
	return true
}

// techLevelMatches reports whether asset's definition requires exactly want. A nil
// want is no filter; a missing asset, rulebook or definition matches nothing.
func techLevelMatches(asset *domain.Asset, want *int, rulebook *rulebook.Rulebook) bool {
	if want == nil {
		return true
	}
	if asset == nil || rulebook == nil {
		return false
	}
	def := rulebook.Assets[asset.DefinitionID]
	return def != nil && def.TechLevel == *want
}
```

##### Task 4 — the three `rollMatches` callers

`internal/faction/engine/effect/effects/bonus_die.go`:

**Find:**
```go
func (modifier *bonusDieModifier) OfferModifiers(ctx hooks.RollContext, factionState *state.FactionState, _ *rulebook.Rulebook) []hooks.ModifierOffer {
	asset, live := modifier.holder.live(factionState)
	if !live || !rollMatches(ctx, modifier.trigger, modifier.holder.factionID, asset, modifier.filters) {
```
**Replace with:**
```go
func (modifier *bonusDieModifier) OfferModifiers(ctx hooks.RollContext, factionState *state.FactionState, rulebook *rulebook.Rulebook) []hooks.ModifierOffer {
	asset, live := modifier.holder.live(factionState)
	if !live || !rollMatches(ctx, modifier.trigger, modifier.holder.factionID, asset, modifier.filters, rulebook) {
```

`internal/faction/engine/effect/effects/reroll.go`:

**Find:**
```go
func (hook *rerollHook) OnRollResult(ctx hooks.RollContext, result hooks.RollResult, factionState *state.FactionState, _ *rulebook.Rulebook) hooks.RerollDirective {
	asset, live := hook.holder.live(factionState)
	if !live || !rollMatches(ctx, hook.trigger, hook.holder.factionID, asset, hook.filters) {
```
**Replace with:**
```go
func (hook *rerollHook) OnRollResult(ctx hooks.RollContext, result hooks.RollResult, factionState *state.FactionState, rulebook *rulebook.Rulebook) hooks.RerollDirective {
	asset, live := hook.holder.live(factionState)
	if !live || !rollMatches(ctx, hook.trigger, hook.holder.factionID, asset, hook.filters, rulebook) {
```

`internal/faction/engine/effect/effects/remove_tag_dice.go`:

**Find:**
```go
	asset, live := filter.holder.live(factionState)
	if !live || !rollMatches(ctx, filter.trigger, filter.holder.factionID, asset, filter.filters) {
```
**Replace with:**
```go
	asset, live := filter.holder.live(factionState)
	// deliberate: a nil rulebook, not a threaded one, because removeTagDiceRegistrar accepts no tech-level filter, the only filter that reads it (Effort 2 Decision 6).
	if !live || !rollMatches(ctx, filter.trigger, filter.holder.factionID, asset, filter.filters, nil) {
```

##### Task 5 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
				domain.FilterAttribute, domain.FilterOpponentAttribute, domain.FilterPhase, domain.FilterAction,
```
**Replace with:**
```go
				domain.FilterAttribute, domain.FilterOpponentAttribute, domain.FilterPhase, domain.FilterAction,
				domain.FilterTechLevel, domain.FilterOpponentTechLevel,
```

##### Task 6 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "Once per turn, roll an extra die when defending with an asset that requires tech level 0 to purchase."
```
**Replace with:**
```toml
effect = "Once per turn, roll an extra die when defending with an asset that requires tech level 0 to purchase."

[[tags.Savage.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Savage.effects.filters]
owner      = "own"
scope      = "faction"
phase      = ["defense"]
tech_level = 0
[tags.Savage.effects.gates]
once_per = "owner_faction"
[tags.Savage.effects.bonus_die]
sides = 10
count = 1
```

**Find:**
```toml
effect = "Once per turn, roll an additional d10 when attacking an asset that requires tech level 5. Roll an extra die when making a test to detect Stealthed assets."
```
**Replace with:**
```toml
effect = "Once per turn, roll an additional d10 when attacking an asset that requires tech level 5. Roll an extra die when making a test to detect Stealthed assets."

[[tags.PerimeterAgency.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.PerimeterAgency.effects.filters]
owner               = "own"
scope               = "faction"
phase               = ["attack"]
opponent_tech_level = 5
[tags.PerimeterAgency.effects.gates]
once_per = "owner_faction"
[tags.PerimeterAgency.effects.bonus_die]
sides = 10
count = 1
```

##### Task 7 — tests

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
)
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
)
```

**Find:**
```go
	other := &domain.Asset{ID: "y", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
```
**Replace with:**
```go
	other := &domain.Asset{ID: "y", OwnerID: "alpha", Location: domain.Location{WorldID: "Krylos"}}
	primitive := &domain.Asset{ID: "p", OwnerID: "alpha", DefinitionID: "tl0"}
	advanced := &domain.Asset{ID: "q", OwnerID: "beta", DefinitionID: "tl5"}
	book := &rulebook.Rulebook{Assets: map[string]*domain.AssetDefinition{
		"tl0": {ID: "tl0", TechLevel: 0},
		"tl5": {ID: "tl5", TechLevel: 5},
	}}
	zero, five := 0, 5
```

**Find:**
```go
	againstCunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentAttribute: domain.StatCunning}
```
**Replace with:**
```go
	againstCunning := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentAttribute: domain.StatCunning}
	withPrimitive := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, TechLevel: &zero}
	againstAdvanced := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentTechLevel: &five}
```

**Find:**
```go
		{"opponent attribute mismatch", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Cunning", OpponentAttribute: "Wealth"}, false},
```
**Replace with:**
```go
		{"opponent attribute mismatch", domain.TriggerOwnRoll, againstCunning, hooks.RollContext{Actor: alpha, Attribute: "Cunning", OpponentAttribute: "Wealth"}, false},
		{"tech level match", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha, Asset: primitive}, true},
		{"tech level, asset with no definition", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha, Asset: declaring}, false},
		{"tech level, no rolling asset", domain.TriggerOwnRoll, withPrimitive, hooks.RollContext{Actor: alpha}, false},
		{"opponent tech level match", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: advanced}, true},
		{"opponent tech level mismatch", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: primitive}, false},
```

**Find:**
```go
			if got := rollMatches(tc.ctx, tc.trigger, "alpha", declaring, tc.filters); got != tc.want {
```
**Replace with:**
```go
			if got := rollMatches(tc.ctx, tc.trigger, "alpha", declaring, tc.filters, book); got != tc.want {
```

**Find:**
```go
	if !rollMatches(hooks.RollContext{Actor: alpha, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction) {
```
**Replace with:**
```go
	if !rollMatches(hooks.RollContext{Actor: alpha, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction, nil) {
```

**Find:**
```go
	if rollMatches(hooks.RollContext{Actor: beta, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction) {
		t.Error("faction scope: a rival's roll matches")
	}
```
**Replace with:**
```go
	if rollMatches(hooks.RollContext{Actor: beta, World: "Tartarus"}, domain.TriggerOwnRoll, "alpha", nil, faction, nil) {
		t.Error("faction scope: a rival's roll matches")
	}
	if rollMatches(hooks.RollContext{Actor: alpha, Asset: primitive}, domain.TriggerOwnRoll, "alpha", nil, withPrimitive, nil) {
		t.Error("a tech-level filter with no rulebook matches")
	}
```

`internal/faction/rulebook/effects_test.go`, appended at the end of the file:

```go
func TestConvertEffectFilters_TechLevels(t *testing.T) {
	zero, five := 0, 5
	filters, err := convertEffectFilters(effectFiltersRecord{Owner: "own", Scope: "faction", TechLevel: &zero, OpponentTechLevel: &five})
	if err != nil || filters.TechLevel == nil || *filters.TechLevel != 0 || filters.OpponentTechLevel == nil || *filters.OpponentTechLevel != 5 {
		t.Errorf("filters = %+v, err = %v; want tech level 0 and opponent tech level 5", filters, err)
	}
	if want := []domain.EffectFilterName{domain.FilterTechLevel, domain.FilterOpponentTechLevel}; !reflect.DeepEqual(filters.Used(), want) {
		t.Errorf("Used() = %v, want %v: tech_level = 0 is a filter", filters.Used(), want)
	}
	negative := -1
	for _, record := range []effectFiltersRecord{
		{Owner: "own", Scope: "faction", TechLevel: &negative},
		{Owner: "own", Scope: "faction", OpponentTechLevel: &negative},
	} {
		if _, err := convertEffectFilters(record); err == nil || !strings.Contains(err.Error(), "must be >= 0") {
			t.Errorf("err = %v, want a negative tech level refused", err)
		}
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, in `TestLoad_TagBonusDice`:

**Find:**
```go
	defense := []domain.EffectRollPhase{domain.RollPhaseDefense}
```
**Replace with:**
```go
	defense := []domain.EffectRollPhase{domain.RollPhaseDefense}
	zero, five := 0, 5
```

**Find:**
```go
		{"SWN-T-004", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
```
**Replace with:**
```go
		{"SWN-T-004", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, OpponentAttribute: domain.StatWealth}, domain.OncePerOwnerFaction)},
		{"SWN-T-015", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: defense, TechLevel: &zero}, domain.OncePerOwnerFaction)},
		{"SWN-T-009", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, OpponentTechLevel: &five}, domain.OncePerOwnerFaction)},
```

##### Commit message

```
feat: Savage and Perimeter Agency read each side's tech level

- New filters tech_level and opponent_tech_level: an exact match on the
  rolling or the opposing asset's definition; 0 is a real value
- rollMatches takes the rulebook; remove_tag_dice passes nil
- Savage: +1d10 once per turn defending with a TL0 asset
- Perimeter Agency: +1d10 once per turn attacking a TL5 asset
```

#### Commit 6 — `feat: Imperialists' die on its Planetary Seizure target`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	FilterOpponentTechLevel EffectFilterName = "opponent_tech_level"
)
```
**Replace with:**
```go
	FilterOpponentTechLevel EffectFilterName = "opponent_tech_level"
	FilterGoal              EffectFilterName = "goal"
)
```

**Find:**
```go
	OpponentTechLevel *int
	CausedBy          EffectOwner
}
```
**Replace with:**
```go
	OpponentTechLevel *int
	CausedBy          EffectOwner
	Goal              string
}
```

**Find:**
```go
	if filters.CausedBy != "" {
		present = append(present, FilterCausedBy)
	}
	return present
```
**Replace with:**
```go
	if filters.CausedBy != "" {
		present = append(present, FilterCausedBy)
	}
	if filters.Goal != "" {
		present = append(present, FilterGoal)
	}
	return present
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	CausedBy          string   `toml:"caused_by"`
}
```
**Replace with:**
```go
	CausedBy          string   `toml:"caused_by"`
	Goal              string   `toml:"goal"`
}
```

**Find:**
```go
		MinTechLevel: record.MinTechLevel, TechLevel: record.TechLevel, OpponentTechLevel: record.OpponentTechLevel,
	}
```
**Replace with:**
```go
		MinTechLevel: record.MinTechLevel, TechLevel: record.TechLevel, OpponentTechLevel: record.OpponentTechLevel,
		Goal: record.Goal,
	}
```

##### Task 3 — `internal/faction/rulebook/rulebook.go`

**Find:**
```go
	"fmt"
	"path/filepath"
```
**Replace with:**
```go
	"fmt"
	"maps"
	"path/filepath"
```

In `Load`:

**Find:**
```go
	goals, err := loadGoals(filepath.Join(dataDir, "goals.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading goals: %w", err)
	}
```
**Replace with:**
```go
	goals, err := loadGoals(filepath.Join(dataDir, "goals.toml"))
	if err != nil {
		return nil, fmt.Errorf("loading goals: %w", err)
	}
	if err := checkGoalReferences(assets, tags, goals); err != nil {
		return nil, fmt.Errorf("checking goal references: %w", err)
	}
```

After `loadGoals`:

**Find:**
```go
		result[r.ID] = &domain.Goal{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Difficulty:  r.Difficulty,
		}
	}
	return result, nil
}
```
**Replace with:**
```go
		result[r.ID] = &domain.Goal{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Difficulty:  r.Difficulty,
		}
	}
	return result, nil
}

// checkGoalReferences refuses an [[effects]] entry whose goal filter names a goal
// the rulebook does not have: a data-to-data reference (D-5), checked once here.
func checkGoalReferences(assets map[string]*domain.AssetDefinition, tags map[string]*domain.Tag, goals map[string]*domain.Goal) error {
	check := func(holder string, entries []domain.EffectDefinition) error {
		for i, entry := range entries {
			if goal := entry.Filters.Goal; goal != "" && goals[goal] == nil {
				return fmt.Errorf("%s effects[%d]: unknown goal %q", holder, i, goal)
			}
		}
		return nil
	}
	for _, id := range slices.Sorted(maps.Keys(assets)) {
		if err := check(fmt.Sprintf("asset %q", id), assets[id].Effects); err != nil {
			return err
		}
	}
	for _, id := range slices.Sorted(maps.Keys(tags)) {
		if err := check(fmt.Sprintf("tag %q", id), tags[id].Effects); err != nil {
			return err
		}
	}
	return nil
}
```

##### Task 4 — `internal/faction/engine/effect/effects/roll_match.go`

**Find:**
```go
// matches without a rulebook.
```
**Replace with:**
```go
// matches without a rulebook. goal matches a roll on the target world of the
// owner's active goal.
```

**Find:**
```go
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
	return true
}
```
**Replace with:**
```go
	if !techLevelMatches(ctx.Asset, filters.TechLevel, rulebook) || !techLevelMatches(ctx.OpponentAsset, filters.OpponentTechLevel, rulebook) {
		return false
	}
	if filters.Goal != "" && !onGoalTarget(owner, filters.Goal, ctx.World) {
		return false
	}
	return true
}

// onGoalTarget reports whether owner's active goal is goalID and world is its
// target. A goal with no target chosen yet has none.
func onGoalTarget(owner *domain.Faction, goalID, world string) bool {
	goal := owner.ActiveGoal
	return goal != nil && goal.GoalID == goalID && world != "" && goal.TargetWorld.WorldID == world
}
```

##### Task 5 — `internal/faction/engine/effect/effects/bonus_die.go`

**Find:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel,
```
**Replace with:**
```go
				domain.FilterTechLevel, domain.FilterOpponentTechLevel, domain.FilterGoal,
```

##### Task 6 — `rulebooks/swn/tags.toml`

`A.001.8` renames the goal reference when it prefixes goal IDs (overview Out of Scope).

**Find:**
```toml
effect = "May roll an extra d10 for attacks made as part of a Seize Planet action."
```
**Replace with:**
```toml
effect = "May roll an extra d10 for attacks made as part of a Seize Planet action."

[[tags.Imperialists.effects]]
trigger = "own_roll"
effect  = "bonus_die"
[tags.Imperialists.effects.filters]
owner = "own"
scope = "faction"
phase = ["attack"]
goal  = "G-004"
[tags.Imperialists.effects.bonus_die]
sides = 10
count = 1
```

##### Task 7 — tests

`internal/faction/engine/effect/effects/roll_match_test.go`:

**Find:**
```go
	alpha := &domain.Faction{ID: "alpha", Homeworld: domain.Location{WorldID: "Krylos"}}
```
**Replace with:**
```go
	alpha := &domain.Faction{
		ID: "alpha", Homeworld: domain.Location{WorldID: "Krylos"},
		ActiveGoal: &domain.ActiveGoal{GoalID: "G-004", TargetWorld: domain.Location{WorldID: "Krylos"}},
	}
	planning := &domain.Faction{ID: "alpha", ActiveGoal: &domain.ActiveGoal{GoalID: "G-004"}}
```

**Find:**
```go
	againstAdvanced := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentTechLevel: &five}
```
**Replace with:**
```go
	againstAdvanced := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, OpponentTechLevel: &five}
	seizure := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Goal: "G-004"}
	otherGoal := domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Goal: "G-001"}
```

**Find:**
```go
		{"opponent tech level mismatch", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: primitive}, false},
```
**Replace with:**
```go
		{"opponent tech level mismatch", domain.TriggerOwnRoll, againstAdvanced, hooks.RollContext{Actor: alpha, OpponentAsset: primitive}, false},
		{"goal target world", domain.TriggerOwnRoll, seizure, hooks.RollContext{Actor: alpha, World: "Krylos"}, true},
		{"off the goal target", domain.TriggerOwnRoll, seizure, hooks.RollContext{Actor: alpha, World: "Tartarus"}, false},
		{"another goal", domain.TriggerOwnRoll, otherGoal, hooks.RollContext{Actor: alpha, World: "Krylos"}, false},
		{"goal target not yet chosen", domain.TriggerOwnRoll, seizure, hooks.RollContext{Actor: planning, World: ""}, false},
```

`internal/faction/rulebook/effects_test.go`, appended at the end of the file:

```go
func TestConvertEffectFilters_Goal(t *testing.T) {
	filters, err := convertEffectFilters(effectFiltersRecord{Owner: "own", Scope: "faction", Goal: "G-004"})
	if err != nil || filters.Goal != "G-004" {
		t.Errorf("filters = %+v, err = %v; want goal G-004", filters, err)
	}
	if !reflect.DeepEqual(filters.Used(), []domain.EffectFilterName{domain.FilterGoal}) {
		t.Errorf("Used() = %v, want [goal]", filters.Used())
	}
}
```

`internal/faction/rulebook/rulebook_test.go`, after `TestLoad_TagBonusDice`:

**Find:**
```go
func TestLoad_TransportBypasses(t *testing.T) {
```
**Replace with:**
```go
func TestCheckGoalReferences(t *testing.T) {
	goals := map[string]*domain.Goal{"G-004": {ID: "G-004"}}
	naming := func(goal string) []domain.EffectDefinition {
		return []domain.EffectDefinition{{Filters: domain.EffectFilters{Goal: goal}}}
	}
	if err := checkGoalReferences(nil, map[string]*domain.Tag{"T-1": {ID: "T-1", Effects: naming("G-004")}}, goals); err != nil {
		t.Errorf("known goal: err = %v, want nil", err)
	}
	err := checkGoalReferences(nil, map[string]*domain.Tag{"T-1": {ID: "T-1", Effects: naming("G-404")}}, goals)
	if err == nil || !strings.Contains(err.Error(), `tag "T-1" effects[0]: unknown goal "G-404"`) {
		t.Errorf("unknown goal on a tag: err = %v, want it named", err)
	}
	err = checkGoalReferences(map[string]*domain.AssetDefinition{"A-1": {ID: "A-1", Effects: naming("G-404")}}, nil, goals)
	if err == nil || !strings.Contains(err.Error(), `asset "A-1" effects[0]: unknown goal "G-404"`) {
		t.Errorf("unknown goal on an asset: err = %v, want it named", err)
	}
}

func TestLoad_TransportBypasses(t *testing.T) {
```

In `TestLoad_TagBonusDice`:

**Find:**
```go
		{"SWN-T-009", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, OpponentTechLevel: &five}, domain.OncePerOwnerFaction)},
```
**Replace with:**
```go
		{"SWN-T-009", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, OpponentTechLevel: &five}, domain.OncePerOwnerFaction)},
		{"SWN-T-006", tagBonusDie(domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction, Phases: attack, Goal: "G-004"}, "")},
```

##### Commit message

```
feat: Imperialists' die on its Planetary Seizure target

- New filter goal: a roll on the target world of the owner's active goal
- rulebook.Load refuses an entry that names an unknown goal
- Imperialists: +1d10 on its attacks on its World Seizure target (D-5)
```

### Phase 4 — Mercenary Group

**In plain words:** a Mercenary Group faction's assets all move at least one hex a
turn, facilities included; a facility can still ride a transport instead.

#### Commit 7 — `feat: a speed floor rule hook`

##### Task 1 — `internal/faction/domain/effect.go`

**Find:**
```go
	SpecialCostDiscount       SpecialEffectType = "cost_discount"
)
```
**Replace with:**
```go
	SpecialCostDiscount       SpecialEffectType = "cost_discount"
	SpecialSpeedFloor         SpecialEffectType = "speed_floor"
)
```

**Find:**
```go
	TriggerAssetDestroyed         EffectTrigger = "asset_destroyed"
)
```
**Replace with:**
```go
	TriggerAssetDestroyed         EffectTrigger = "asset_destroyed"
	TriggerSpeedQuery             EffectTrigger = "speed_query"
)
```

**Find:**
```go
type CostDiscountParams struct {
	Amount int
}
```
**Replace with:**
```go
type CostDiscountParams struct {
	Amount int
}

type SpeedFloorParams struct {
	Speed int
}
```

**Find:**
```go
	DiscountRepair *DiscountRepairParams
	CostDiscount   *CostDiscountParams
}
```
**Replace with:**
```go
	DiscountRepair *DiscountRepairParams
	CostDiscount   *CostDiscountParams
	SpeedFloor     *SpeedFloorParams
}
```

##### Task 2 — `internal/faction/rulebook/effects.go`

**Find:**
```go
	CostDiscount   *costDiscountRecord   `toml:"cost_discount"`
}
```
**Replace with:**
```go
	CostDiscount   *costDiscountRecord   `toml:"cost_discount"`
	SpeedFloor     *speedFloorRecord     `toml:"speed_floor"`
}
```

**Find:**
```go
type costDiscountRecord struct {
	Amount int `toml:"amount"`
}
```
**Replace with:**
```go
type costDiscountRecord struct {
	Amount int `toml:"amount"`
}

type speedFloorRecord struct {
	Speed int `toml:"speed"`
}
```

**Find:**
```go
		domain.TriggerAssetDestroyed,
	}
```
**Replace with:**
```go
		domain.TriggerAssetDestroyed, domain.TriggerSpeedQuery,
	}
```

**Find:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies, domain.SpecialCostDiscount,
```
**Replace with:**
```go
		domain.SpecialLockOut, domain.SpecialLoseTies, domain.SpecialCostDiscount, domain.SpecialSpeedFloor,
```

In `presentParamBlocks`:

**Find:**
```go
	if record.CostDiscount != nil {
		present = append(present, domain.SpecialCostDiscount)
	}
	return present
```
**Replace with:**
```go
	if record.CostDiscount != nil {
		present = append(present, domain.SpecialCostDiscount)
	}
	if record.SpeedFloor != nil {
		present = append(present, domain.SpecialSpeedFloor)
	}
	return present
```

In `convertEffectParams`:

**Find:**
```go
		definition.CostDiscount = &domain.CostDiscountParams{Amount: record.CostDiscount.Amount}
	}
	return nil
}
```
**Replace with:**
```go
		definition.CostDiscount = &domain.CostDiscountParams{Amount: record.CostDiscount.Amount}
	case domain.SpecialSpeedFloor:
		if record.SpeedFloor == nil {
			return missingParamBlock(definition.Effect)
		}
		if record.SpeedFloor.Speed <= 0 {
			return fmt.Errorf("speed_floor speed must be > 0, got %d", record.SpeedFloor.Speed)
		}
		definition.SpeedFloor = &domain.SpeedFloorParams{Speed: record.SpeedFloor.Speed}
	}
	return nil
}
```

##### Task 3 — `internal/faction/engine/hooks/rule_modifier.go`

**Find:**
```go
type WorldTechLevelModifier interface {
	ModifyWorldTechLevel(faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int
}
```
**Replace with:**
```go
type WorldTechLevelModifier interface {
	ModifyWorldTechLevel(faction *domain.Faction, def *domain.AssetDefinition, world string, baseTL int) int
}

// AssetSpeedModifier adjusts the speed, in hexes a turn, of an asset faction
// owns, read by the movement phase (Cat 4).
type AssetSpeedModifier interface {
	ModifyAssetSpeed(faction *domain.Faction, def *domain.AssetDefinition, speed int) int
}
```

##### Task 4 — `internal/faction/engine/hooks/registry.go`

**Find:**
```go
type RegisteredAttackEligibilityRule struct {
```
**Replace with:**
```go
type RegisteredAssetSpeedModifier struct {
	Scope  Scope
	Source string
	Hook   AssetSpeedModifier
}

type RegisteredAttackEligibilityRule struct {
```

**Find:**
```go
	worldTechLevelModifiers  map[Scope][]RegisteredWorldTechLevelModifier
	attackEligibilityRules   map[Scope][]RegisteredAttackEligibilityRule
```
**Replace with:**
```go
	worldTechLevelModifiers  map[Scope][]RegisteredWorldTechLevelModifier
	assetSpeedModifiers      map[Scope][]RegisteredAssetSpeedModifier
	attackEligibilityRules   map[Scope][]RegisteredAttackEligibilityRule
```

In `Reset`:

**Find:**
```go
	registry.worldTechLevelModifiers = make(map[Scope][]RegisteredWorldTechLevelModifier)
```
**Replace with:**
```go
	registry.worldTechLevelModifiers = make(map[Scope][]RegisteredWorldTechLevelModifier)
	registry.assetSpeedModifiers = make(map[Scope][]RegisteredAssetSpeedModifier)
```

In `ClearAsset`:

**Find:**
```go
	delete(registry.worldTechLevelModifiers, scope)
```
**Replace with:**
```go
	delete(registry.worldTechLevelModifiers, scope)
	delete(registry.assetSpeedModifiers, scope)
```

**Find:**
```go
// RegisterAttackEligibilityRule registers a Cat 4 attack-eligibility hook under the given scope.
```
**Replace with:**
```go
// RegisterAssetSpeedModifier registers a Cat 4 speed hook under the given scope.
func (registry *Registry) RegisterAssetSpeedModifier(scope Scope, source string, hook AssetSpeedModifier) {
	registry.assetSpeedModifiers[scope] = append(registry.assetSpeedModifiers[scope], RegisteredAssetSpeedModifier{Scope: scope, Source: source, Hook: hook})
}

// RegisterAttackEligibilityRule registers a Cat 4 attack-eligibility hook under the given scope.
```

**Find:**
```go
// AttackEligibilityRulesFor returns all Cat 4 attack-eligibility hooks matching the given faction and asset.
```
**Replace with:**
```go
// AssetSpeedModifiersFor returns all Cat 4 speed hooks matching the given faction and asset.
func (registry *Registry) AssetSpeedModifiersFor(factionID, assetInstanceID string) []RegisteredAssetSpeedModifier {
	var result []RegisteredAssetSpeedModifier
	result = append(result, registry.assetSpeedModifiers[GlobalScope()]...)
	result = append(result, registry.assetSpeedModifiers[FactionScope(factionID)]...)
	if assetInstanceID != "" {
		result = append(result, registry.assetSpeedModifiers[AssetScope(factionID, assetInstanceID)]...)
	}
	return result
}

// AssetSpeedModifiersForAsset returns the Cat 4 speed hooks registered against one
// asset instance, without the global or faction buckets.
func (registry *Registry) AssetSpeedModifiersForAsset(factionID, assetInstanceID string) []RegisteredAssetSpeedModifier {
	return registry.assetSpeedModifiers[AssetScope(factionID, assetInstanceID)]
}

// AttackEligibilityRulesFor returns all Cat 4 attack-eligibility hooks matching the given faction and asset.
```

##### Task 5 — `internal/faction/engine/hooks/dispatch/rules.go`

**Find:**
```go
// ResolveAttackEligibility reports whether query.Attacker may deal damage to
```
**Replace with:**
```go
// ResolveAssetSpeed returns the speed of a def asset that faction owns: the
// definition's own, run through faction's AssetSpeedModifiers, then those
// registered against each asset faction owns, in asset ID order. A speed floor
// serves its owner only, so a rival's assets are never asked. Returns def.Speed
// when registry is nil.
func ResolveAssetSpeed(registry *hooks.Registry, faction *domain.Faction, def *domain.AssetDefinition) int {
	if registry == nil {
		return def.Speed
	}
	speed := def.Speed
	for _, registered := range registry.AssetSpeedModifiersFor(faction.ID, "") {
		speed = registered.Hook.ModifyAssetSpeed(faction, def, speed)
	}
	for _, asset := range domain.SortedAssets(faction) {
		for _, registered := range registry.AssetSpeedModifiersForAsset(faction.ID, asset.ID) {
			speed = registered.Hook.ModifyAssetSpeed(faction, def, speed)
		}
	}
	return speed
}

// ResolveAttackEligibility reports whether query.Attacker may deal damage to
```

##### Task 6 — `internal/faction/engine/effect/effects/speed_floor.go` (new)

```go
package effects

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
)

// speedFloorRegistrar reads owner = "own" as "the moving asset is the owner's",
// and scope = "faction" as "every asset the owner holds" (Mercenary Group). The
// entry carries the floor; an asset already faster keeps its speed.
var speedFloorRegistrar = effect.Registrar{
	Triggers: map[domain.EffectTrigger]effect.Accepts{
		domain.TriggerSpeedQuery: {
			Owners: []domain.EffectOwner{domain.OwnerOwn},
			Scopes: []domain.EffectScope{domain.ScopeFaction},
		},
	},
	Register: registerSpeedFloor,
}

func registerSpeedFloor(ctx effect.RegisterContext) {
	ctx.Registry.RegisterAssetSpeedModifier(
		ctx.Holder.Scope(),
		ctx.Holder.HookID(domain.SpecialSpeedFloor),
		&speedFloorModifier{
			owner:  ctx.Holder.Faction,
			holder: refOf(ctx.Holder),
			speed:  ctx.Entry.SpeedFloor.Speed,
		},
	)
}

// speedFloorModifier holds its owner, not its holder's asset: a Cat 4 hook gets no
// FactionState (Decision Record 19).
type speedFloorModifier struct {
	owner  *domain.Faction
	holder holderRef
	speed  int
}

func (modifier *speedFloorModifier) ModifyAssetSpeed(faction *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	if _, live := modifier.holder.liveIn(modifier.owner); !live || faction.ID != modifier.owner.ID {
		return speed
	}
	return max(speed, modifier.speed)
}
```

##### Task 7 — `internal/faction/engine/effect/effects/registrars.go`

**Find:**
```go
		domain.SpecialCostDiscount:       costDiscountRegistrar,
```
**Replace with:**
```go
		domain.SpecialCostDiscount:       costDiscountRegistrar,
		domain.SpecialSpeedFloor:         speedFloorRegistrar,
```

##### Task 8 — tests

`internal/faction/engine/effect/effects/speed_floor_test.go` (new):

```go
package effects_test

import (
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/effect/effects"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
)

func TestSpeedFloor_RaisesTheOwnersSlowAssets(t *testing.T) {
	mercenaries := &domain.Tag{ID: "SWN-T-008", Name: "Mercenary Group"}
	alpha := &domain.Faction{ID: "alpha", Tags: []string{mercenaries.ID}}
	beta := &domain.Faction{ID: "beta"}
	registry := hooks.NewRegistry()
	effects.Registrars()[domain.SpecialSpeedFloor].Register(effect.RegisterContext{
		Holder: effect.Holder{Faction: alpha, Tag: mercenaries},
		Entry: domain.EffectDefinition{
			Trigger:    domain.TriggerSpeedQuery,
			Effect:     domain.SpecialSpeedFloor,
			Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
			SpeedFloor: &domain.SpeedFloorParams{Speed: 1},
		},
		Registry: registry,
	})
	facility := &domain.AssetDefinition{ID: "facility"}
	starship := &domain.AssetDefinition{ID: "starship", Speed: 3}

	cases := []struct {
		name    string
		faction *domain.Faction
		def     *domain.AssetDefinition
		want    int
	}{
		{"a speed-0 asset gains the floor", alpha, facility, 1},
		{"a faster asset keeps its speed", alpha, starship, 3},
		{"a rival's asset is not raised", beta, facility, 0},
	}
	for _, tc := range cases {
		if got := dispatch.ResolveAssetSpeed(registry, tc.faction, tc.def); got != tc.want {
			t.Errorf("%s: speed = %d, want %d", tc.name, got, tc.want)
		}
	}

	alpha.Tags = nil
	if got := dispatch.ResolveAssetSpeed(registry, alpha, facility); got != 0 {
		t.Errorf("speed after the tag is dropped = %d, want 0", got)
	}
}
```

`internal/faction/engine/hooks/dispatch/rules_test.go`, after
`TestResolveWorldTechLevel_PassesDefinition`:

**Find:**
```go
func TestResolveMaintenanceCost_NoModifiers(t *testing.T) {
```
**Replace with:**
```go
// stubAssetSpeedModifier raises speed to a fixed floor.
type stubAssetSpeedModifier struct{ floor int }

func (stub stubAssetSpeedModifier) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	return max(speed, stub.floor)
}

func TestResolveAssetSpeed(t *testing.T) {
	def := &domain.AssetDefinition{ID: "facility"}
	faction := &domain.Faction{ID: "f1", Assets: map[string]*domain.Asset{"a1": {ID: "a1", OwnerID: "f1"}}}
	if got := ResolveAssetSpeed(nil, faction, def); got != 0 {
		t.Errorf("nil registry: speed = %d, want the definition's 0", got)
	}

	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.AssetScope("f1", "a1"), "own asset", stubAssetSpeedModifier{floor: 1})
	registry.RegisterAssetSpeedModifier(hooks.AssetScope("rival", "r1"), "rival", stubAssetSpeedModifier{floor: 9})
	if got := ResolveAssetSpeed(registry, faction, def); got != 1 {
		t.Errorf("speed = %d, want 1: the own asset's floor is asked, the rival's never", got)
	}

	registry.RegisterAssetSpeedModifier(hooks.FactionScope("f1"), "faction", stubAssetSpeedModifier{floor: 2})
	if got := ResolveAssetSpeed(registry, faction, def); got != 2 {
		t.Errorf("speed = %d, want 2 from the faction's floor", got)
	}
}

func TestResolveMaintenanceCost_NoModifiers(t *testing.T) {
```

`internal/faction/rulebook/effects_test.go`, appended at the end of the file:

```go
func TestLoadEffects_SpeedFloor(t *testing.T) {
	entry := func(speed string) string {
		return effectTestAsset + `
[[assets.Machine.effects]]
trigger = "speed_query"
effect  = "speed_floor"
[assets.Machine.effects.filters]
owner = "own"
scope = "faction"
[assets.Machine.effects.speed_floor]
speed = ` + speed + "\n"
	}
	assets, err := loadAssetsFromTOML(t, entry("1"))
	if err != nil {
		t.Fatalf("loadAssets: %v", err)
	}
	if got := assets["T-001"].Effects[0].SpeedFloor; !reflect.DeepEqual(got, &domain.SpeedFloorParams{Speed: 1}) {
		t.Errorf("SpeedFloor = %+v, want speed 1", got)
	}
	if _, err := loadAssetsFromTOML(t, entry("0")); err == nil || !strings.Contains(err.Error(), "speed_floor speed must be > 0") {
		t.Errorf("err = %v, want speed 0 refused", err)
	}
}
```

##### Commit message

```
feat: a speed floor rule hook

- New effect speed_floor on a new trigger speed_query; the entry carries
  the speed
- Cat 4 AssetSpeedModifier and dispatch.ResolveAssetSpeed, an owner-only
  walk of global, faction and own asset buckets
- No reader yet: the movement phase reads it in the next commit
```

#### Commit 8 — `feat: Mercenary Group as rulebook data`

##### Task 1 — `internal/faction/engine/world/movement.go`

**Find:**
```go
import (
	"fmt"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)
```
**Replace with:**
```go
import (
	"fmt"
	"log/slog"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)
```

**Find:**
```go
func (engine *WorldEngine) TickMovementOrders(faction *domain.Faction, rulebook *rulebook.Rulebook, log *slog.Logger) ([]domain.Mutation, error) {
```
**Replace with:**
```go
// TickMovementOrders advances each open order of faction's assets by the asset's
// speed, as registry's speed floors resolve it (Mercenary Group).
func (engine *WorldEngine) TickMovementOrders(faction *domain.Faction, rulebook *rulebook.Rulebook, registry *hooks.Registry, log *slog.Logger) ([]domain.Mutation, error) {
```

**Find:**
```go
		newStepIdx := asset.CurrentOrder.StepIdx + def.Speed
```
**Replace with:**
```go
		newStepIdx := asset.CurrentOrder.StepIdx + dispatch.ResolveAssetSpeed(registry, faction, def)
```

##### Task 2 — `internal/faction/engine/orchestrator.go`

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks/dispatch"
```

**Find:**
```go
	tickMutations, err := e.World.TickMovementOrders(faction, e.Rulebook, worldLog)
```
**Replace with:**
```go
	tickMutations, err := e.World.TickMovementOrders(faction, e.Rulebook, e.Hooks, worldLog)
```

**Find:**
```go
	decisionMutations, err := prepareMovementDecisions(faction, collectors.Phase, e.World, e.Rulebook, worldLog)
```
**Replace with:**
```go
	decisionMutations, err := prepareMovementDecisions(faction, collectors.Phase, e.World, e.Rulebook, e.Hooks, worldLog)
```

**Find:**
```go
	rulebook *rulebook.Rulebook,
	worldLog *slog.Logger,
) ([]domain.Mutation, error) {
	eligible := eligibleMovableAssets(faction, rulebook)
	if len(eligible) == 0 {
		return nil, nil
	}
	decisions, err := collector.SelectMovementDecisions(faction, eligible)
	if err != nil {
		return nil, fmt.Errorf("selecting movement decisions: %w", err)
	}
	if len(decisions) == 0 {
		return nil, nil
	}

	for i, decision := range decisions {
```
**Replace with:**
```go
	rulebook *rulebook.Rulebook,
	registry *hooks.Registry,
	worldLog *slog.Logger,
) ([]domain.Mutation, error) {
	eligible := eligibleMovableAssets(faction, rulebook, registry)
	if len(eligible) == 0 {
		return nil, nil
	}
	decisions, err := collector.SelectMovementDecisions(faction, eligible)
	if err != nil {
		return nil, fmt.Errorf("selecting movement decisions: %w", err)
	}
	if len(decisions) == 0 {
		return nil, nil
	}
	ordered := make(map[string]bool, len(decisions))
	for _, decision := range decisions {
		ordered[decision.AssetID] = true
	}

	for i, decision := range decisions {
```

**Find:**
```go
		cargoEligible := eligibleCargoForTransport(faction, asset, def.Transport, rulebook)
```
**Replace with:**
```go
		// deliberate: an asset ordered in this batch is not offered as cargo, because a speed floor (Mercenary Group) lets an immobile asset take its own order (Effort 2 Decision 8).
		cargoEligible := slices.DeleteFunc(eligibleCargoForTransport(faction, asset, def.Transport, rulebook), func(cargo *domain.Asset) bool {
			return ordered[cargo.ID]
		})
```

**Find:**
```go
func eligibleMovableAssets(faction *domain.Faction, rulebook *rulebook.Rulebook) []*domain.Asset {
	var eligible []*domain.Asset
	for _, asset := range faction.Assets {
		def := rulebook.Assets[asset.DefinitionID]
		// deliberate: a sheltered asset stays put, not moves, because Boltholes make it unusable until repaired (Effort 2 Decision 11).
		if def != nil && def.Speed > 0 && !asset.HasCondition(domain.ConditionSheltered) {
```
**Replace with:**
```go
func eligibleMovableAssets(faction *domain.Faction, rulebook *rulebook.Rulebook, registry *hooks.Registry) []*domain.Asset {
	var eligible []*domain.Asset
	for _, asset := range faction.Assets {
		def := rulebook.Assets[asset.DefinitionID]
		// deliberate: a sheltered asset stays put, not moves, because Boltholes make it unusable until repaired (Effort 2 Decision 11).
		if def != nil && dispatch.ResolveAssetSpeed(registry, faction, def) > 0 && !asset.HasCondition(domain.ConditionSheltered) {
```

The Boltholes fence-sign names an older initiative's Effort 2; it stays as it is.

**Find:**
```go
		def := rulebook.Assets[asset.DefinitionID]
		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 || asset.HasCondition(domain.ConditionSheltered) {
```
**Replace with:**
```go
		def := rulebook.Assets[asset.DefinitionID]
		// deliberate: the definition's speed, not the resolved one, because a speed floor adds a way to move and does not take the ride away (Effort 2 Decision 8).
		if def == nil || asset.CurrentOrder != nil || def.Speed > 0 || asset.HasCondition(domain.ConditionSheltered) {
```

##### Task 3 — `internal/faction/engine/effect/effects/grant_movement.go`

**Find:**
```go
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset; Effort 2's self-move mode lifts this (Effort 1 Decision 2).
```
**Replace with:**
```go
	// deliberate: refuse a tag, not accept it, because the range is measured from the declaring asset and a tag has none; Mercenary Group is a speed floor, not a grant (Effort 2 Decision 7).
```

##### Task 4 — `rulebooks/swn/tags.toml`

**Find:**
```toml
effect = "All faction assets gain a special ability: as an action, the asset may move itself to any world within one hex."
```
**Replace with:**
```toml
effect = "All faction assets gain a special ability: as an action, the asset may move itself to any world within one hex."

[[tags.MercenaryGroup.effects]]
trigger = "speed_query"
effect  = "speed_floor"
[tags.MercenaryGroup.effects.filters]
owner = "own"
scope = "faction"
[tags.MercenaryGroup.effects.speed_floor]
speed = 1
```

##### Task 5 — tests

`internal/faction/engine/world/world_test.go`:

**Find:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
```
**Replace with:**
```go
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/hooks"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/world"
```

Replace all (two occurrences):

**Find:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(1), slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(1), nil, slog.New(slog.DiscardHandler))
```

**Find:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(2), slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(2), nil, slog.New(slog.DiscardHandler))
```

**Find:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(3), slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := we.TickMovementOrders(faction, testRulebook(3), nil, slog.New(slog.DiscardHandler))
```

Appended at the end of `world_test.go`:

```go
// speedFloor stands in for Mercenary Group's speed_floor hook.
type speedFloor struct{ floor int }

func (floor speedFloor) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	return max(speed, floor.floor)
}

func TestTickMovementOrders_ReadsTheResolvedSpeed(t *testing.T) {
	we := world.NewWithMap(&stubMap{}, slog.New(slog.DiscardHandler))
	faction := factionWithOrder(0, makePath(5), domain.Location{WorldID: "beta"})
	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.FactionScope("alpha"), "floor", speedFloor{floor: 1})

	mutations, err := we.TickMovementOrders(faction, testRulebook(0), registry, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	if progressed, ok := mutations[0].(domain.MovementOrderProgressed); !ok || progressed.NewStepIdx != 1 {
		t.Errorf("mutation = %+v, want progress to step 1: a speed-0 asset under a floor of 1", mutations[0])
	}
}
```

`internal/faction/engine/world/movement_test.go`:

**Find:**
```go
	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, reviseRulebook(), slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	mutations, err := regionWorldEngine(t).TickMovementOrders(faction, reviseRulebook(), nil, slog.New(slog.DiscardHandler))
```

`internal/faction/engine/orchestrator_test.go`:

**Find:**
```go
	eligible := eligibleMovableAssets(faction, transportRulebook())
```
**Replace with:**
```go
	eligible := eligibleMovableAssets(faction, transportRulebook(), nil)
```

**Find:**
```go
	_, err := prepareMovementDecisions(faction, collector, we, rb, slog.New(slog.DiscardHandler))
```
**Replace with:**
```go
	_, err := prepareMovementDecisions(faction, collector, we, rb, nil, slog.New(slog.DiscardHandler))
```

**Find:**
```go
func TestFilterAllowedActions(t *testing.T) {
```
**Replace with:**
```go
// speedFloor stands in for Mercenary Group's speed_floor hook.
type speedFloor struct{ floor int }

func (floor speedFloor) ModifyAssetSpeed(_ *domain.Faction, _ *domain.AssetDefinition, speed int) int {
	return max(speed, floor.floor)
}

func TestEligibleMovableAssets_ReadsTheResolvedSpeed(t *testing.T) {
	special := &domain.Asset{ID: "sf-1", DefinitionID: "sf-def", OwnerID: "alpha"}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{special.ID: special}}
	if eligible := eligibleMovableAssets(faction, transportRulebook(), nil); len(eligible) != 0 {
		t.Errorf("eligible = %v, want none: sf-def has speed 0", eligible)
	}

	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.FactionScope("alpha"), "floor", speedFloor{floor: 1})
	if eligible := eligibleMovableAssets(faction, transportRulebook(), registry); len(eligible) != 1 || eligible[0].ID != "sf-1" {
		t.Errorf("eligible = %v, want sf-1 under a speed floor", eligible)
	}
}

func TestPrepareMovementDecisions_AnOrderedAssetIsNotCargo(t *testing.T) {
	rb := transportRulebook()
	loc := domain.Location{WorldID: "world-a"}
	transport := &domain.Asset{ID: "transport-1", DefinitionID: "transport-def", OwnerID: "alpha", Location: loc}
	walker := &domain.Asset{ID: "sf-1", DefinitionID: "sf-def", OwnerID: "alpha", Location: loc}
	rider := &domain.Asset{ID: "sf-2", DefinitionID: "sf-def", OwnerID: "alpha", Location: loc}
	faction := &domain.Faction{ID: "alpha", Assets: map[string]*domain.Asset{transport.ID: transport, walker.ID: walker, rider.ID: rider}}
	registry := hooks.NewRegistry()
	registry.RegisterAssetSpeedModifier(hooks.FactionScope("alpha"), "floor", speedFloor{floor: 1})

	var offered []*domain.Asset
	collector := &minimalPhaseCollector{
		movementDecisionsFn: func(_ *domain.Faction, _ []*domain.Asset) ([]world.MovementDecision, error) {
			return []world.MovementDecision{
				{AssetID: "transport-1", Kind: world.MovementDecisionIssue, Destination: &domain.Location{WorldID: "world-c"}},
				{AssetID: "sf-1", Kind: world.MovementDecisionIssue, Destination: &domain.Location{WorldID: "world-c"}},
			}, nil
		},
		selectTransportCargo: func(_ *domain.Asset, eligible []*domain.Asset, _ *domain.TransportProfile) ([]*domain.Asset, error) {
			offered = eligible
			return nil, nil
		},
	}

	we := world.NewWithMap(&stubWorldMap{}, slog.New(slog.DiscardHandler))
	// The stub map cannot path the orders, so the error is not the subject: the
	// cargo offer is made before the orders are built.
	_, _ = prepareMovementDecisions(faction, collector, we, rb, registry, slog.New(slog.DiscardHandler))

	if len(offered) != 1 || offered[0].ID != "sf-2" {
		t.Errorf("offered cargo = %v, want only sf-2: sf-1 has its own order", offered)
	}
}

func TestFilterAllowedActions(t *testing.T) {
```

`internal/faction/rulebook/rulebook_test.go`:

**Find:**
```go
func TestLoad_TransportBypasses(t *testing.T) {
```
**Replace with:**
```go
func TestLoad_MercenaryGroupSpeedFloor(t *testing.T) {
	rb := loadShippedRulebook(t)
	want := []domain.EffectDefinition{{
		Trigger:    domain.TriggerSpeedQuery,
		Effect:     domain.SpecialSpeedFloor,
		Filters:    domain.EffectFilters{Owner: domain.OwnerOwn, Scope: domain.ScopeFaction},
		SpeedFloor: &domain.SpeedFloorParams{Speed: 1},
	}}
	if got := rb.Tags["SWN-T-008"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("Mercenary Group effects = %+v, want %+v", got, want)
	}
}

func TestLoad_TransportBypasses(t *testing.T) {
```

##### Commit message

```
feat: Mercenary Group as rulebook data

- The movement tick and the movable list read ResolveAssetSpeed, so a
  speed floor moves an asset the definition holds still
- An asset ordered in this batch is not offered as cargo; cargo
  eligibility keeps the definition's speed (Effort 2 Decision 8)
- Mercenary Group: every asset of the faction moves at least 1 hex a turn
- grant_movement's tag refusal names the new reason
```
