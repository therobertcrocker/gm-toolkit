# A-flag Ability Registry (A.001.2) — Effort 1 Plan: Pipeline + parser foundation

## Context / Goal

Build the per-step composition pipeline and its parser path that Efforts 2–4 all
stand on. After this effort an ability is a uniform `[]AbilityStep`; the engine
runs each step through *resolve-target → gate (Contest) → terminal effect*, keyed
on the effect. The inline stat-contest and target-selection that live inside
`revealStealth` today become generic pipeline stages; `revealStealth` shrinks to
"given the resolved target, emit the stealth-cleared mutations." Informers and
Seductress's reveal step are converted as proof. **No new effects.**

- Overview plan: [`a-flag-ability-registry-plan.md`](./a-flag-ability-registry-plan.md)
  (Effort 1 row in the Effort Summary; scoping outline in Work Breakdown).
- Discovery: [`a-flag-ability-registry-discovery.md`](../discovery/a-flag-ability-registry-discovery.md)
  (D2, D6, D7, D8 bind here).
- Architecture: [ability-catalog](../../architecture/engine/ability-catalog.md) (the
  A-side spine), [actions](../../architecture/engine/actions.md) (hosts `Use Asset Ability`).

## Decisions Ratified in Planning

1. **Halt-on-failure step semantics.** When a step's resolution gate (Contest)
   fails, the loop `break`s — remaining steps do not run; mutations already
   collected stand. Resolves the overview's "multi-step failure semantics" open
   question. **Consequence (authoring invariant):** ungated steps (self-relocate,
   roll-tables) must be ordered *before* gated ones, so e.g. Seductress's
   relocate (Effort 3) stands even when its later reveal contest loses. The loop's
   `break` carries a `// deliberate:` fence-sign recording this; Effort 3 repeats
   it when Seductress's relocate step lands.

2. **Generalize the collector interface now.** `SelectFactionTestTarget` is
   replaced by a single generic `SelectAbilityTarget(asset, step, candidates)
   (AbilityTarget, error)` across all five Collector implementations (mock,
   testharness, TUI adapter, smoke, overlay) this effort. Only the `faction`
   target kind is wired end-to-end; `asset`/`world`/`self`/`none` parse and route,
   but the asset and world kinds return an explicit "not supported yet" error
   until Efforts 3/4 add their candidate-gathering. This pulls the interface
   surgery forward (bigger Effort-1 diff) so Efforts 3/4 only add `TargetKind`
   cases against a method that already exists — they never re-touch the interface.

3. **`StepContext` handler contract.** The pipeline resolves the target and runs
   the gate, then calls a slim `EffectHandler func(StepContext) ([]Mutation,
   error)`. A handler that is called has, by halt-on-failure, already passed its
   gate, so it is terminal-only. `StepContext` carries the ambient inputs
   (`Faction`, `Asset`, `Def`, `Step`, `Target`, `FactionState`, `Rulebook`).
   `Roller` and `Collector` are deliberately **omitted** — reveal needs neither
   (the stages own them); Effort 2 adds the resolved roll result, Effort 4 adds
   `Collector` for its secondary prompt, each at first use (YAGNI).

4. **Per-effect param structs deferred to first use.** The overview's D6 names
   `Relocate`/`Attach`/`Purchase` sub-structs on `AbilityStep`. They are **not**
   added in Effort 1 — defining empty structs Efforts 3/4 will fill is speculative
   scaffolding. `AbilityStep` carries `Effect`, `Target`, `Contest`, `Roll` only;
   Effort 3 adds `Relocate *RelocateParams` and Effort 4 the rest, additively, when
   first consumed. (Deviation from a literal reading of overview D6, justified by
   YAGNI; the shape direction D6 ratifies is unchanged.)

5. **Transient dual-shape across the commits (buildability).** Reshaping
   `AbilityDefinition` from flat fields to `Steps` ripples to every reader. To keep
   each commit individually buildable with green tests, Commit 1 is **additive** —
   `Steps` is added *alongside* the flat `AttackerStat`/`DefenderStat`/`Effect`/
   `Die`/`Outcomes` fields. Commit 2's parser populates `Steps`. Commit 3 rewires
   dispatch + reveal to read `Steps` and only *then* removes the now-dead flat
   fields. No back-compat shim is needed: no test exercises the real-parser →
   dispatch path (verified — `rulebook_test` only parses; dispatch tests hand-build
   defs), so the brief C2→C4 window where real Informers data parses without a
   `target` is untested. Recommend executing C3 and C4 back-to-back to close it.

6. **Keep the faction-target TUI plumbing as-is.** The `AskSelectFactionTestTarget`
   channel, `SelectFactionTestTargetPayload`, and `NewSelectFactionTestTarget`
   overlay are kept and reused as the `faction`-kind implementation behind the new
   generic method — they are the faction-target renderer. Efforts 3/4 add asset/
   world channels + overlays. Renaming them now is cosmetic churn for no gain.

## Decision Record — Execution

<!-- Empty at plan time. Execution-time decisions and reversals land here, logged
before the commit they ride. Append AND reconcile. -->

---

## Open Questions — To Ratify at Implementation Time

- **Multi-step ability mixing built + unbuilt effects** — Effort 1's no-handler
  fallthrough early-returns `confirmApplied`, which would discard a prior step's
  mutations if an unbuilt-effect step followed a built one. Effort 1 has no such
  ability (all authored abilities are single-step reveal), so it never fires.
  Revisit the loop's no-handler semantics in Effort 2, when the first
  description-only special is authored.

## Shared Context

**Model.** This Plan session runs on Opus. For execution: Commit 1 (`refactor:`)
and Commit 3 (pipeline + interface surgery) are design-heavy — suggest Opus;
Commits 2 and 4 are mechanical enough for Sonnet. Surface the recommendation at
the Plan → Execution transition.

**Test patterns.** Follow `ability/dispatch_test.go`: `seqRoller{values}` for
deterministic d10s, gomock `MockCollector` for prompts. The single-step
defender-wins test already covers the gate-fail → halt path (the loop breaks
before dispatch); a genuine *multi-step* halt test lands in Effort 3 with
Seductress's two real steps — Effort 1 does not add a synthetic one.

**Mock regeneration.** `MockCollector` is generated (`//go:generate mockgen` on
`action/collector.go:3`). The interface change is applied by regenerating, not
hand-editing the mock.

## Out of Scope

- All new effects: `coin_delta`, `purchase_discount` (Effort 2); `relocate`
  (Effort 3); `coin_or_disable`, `purchase_enabler`, `attach` (Effort 4).
- The Roll-stage *execution* (rolling the die, mapping `Outcomes`). Effort 1 builds
  the parser path for `Roll` (`Die` + `Outcomes`) but the gate treats a non-Contest
  step as auto-pass; Effort 2 implements roll resolution.
- Asset/world/self target candidate-gathering and their TUI overlays (Efforts 3/4).
- The `RelocateParams`/`AttachParams`/`PurchaseParams` sub-structs (Decision 4).

---

## Work Breakdown

### Phase 1 — Pipeline + parser foundation

Single phase, four commits, executed in order. Each commit builds with green tests
(Decision 5).

#### Commit 1 — `refactor: reshape ability domain into steps`

##### Task 1 — `internal/faction/domain/asset.go`

Reshape the ability vocabulary into the step model. Additive: `Steps` is added
alongside the flat fields (Decision 5); the flat fields are removed in Commit 3.
Coin enum values and their dead branches go now (D8). **This task defines the
shared domain contract every later task and effort cites.**

**Find:**
```go
type AbilityEffectType string

const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
	EffectCoinDrain     AbilityEffectType = "coin_drain"
	EffectCoinSteal     AbilityEffectType = "coin_steal"
)

type AbilityDefinition struct {
	Die          *DiceRoll
	Outcomes     []AbilityOutcome
	AttackerStat FactionStat
	DefenderStat FactionStat
	Effect       AbilityEffectType
}

type AbilityOutcome struct {
	From int
	To   int
	Coin int
}
```
**Replace with:**
```go
type AbilityEffectType string

const (
	EffectRevealStealth AbilityEffectType = "reveal_stealth"
)

type TargetKind string

const (
	TargetNone    TargetKind = "none"
	TargetSelf    TargetKind = "self"
	TargetFaction TargetKind = "faction"
	TargetAsset   TargetKind = "asset"
	TargetWorld   TargetKind = "world"
)

type Contest struct {
	AttackerStat FactionStat
	DefenderStat FactionStat
}

type RollTable struct {
	Die      DiceRoll
	Outcomes []AbilityOutcome
}

// AbilityDefinition is a uniform sequence of steps; a flat [ability] block parses
// to a one-step list. The flat AttackerStat/DefenderStat/Die/Outcomes/Effect
// fields are transitional (removed in Commit 3) — readers use Steps.
type AbilityDefinition struct {
	Steps []AbilityStep

	// Transitional flat fields (removed in Commit 3 once dispatch reads Steps).
	Die          *DiceRoll
	Outcomes     []AbilityOutcome
	AttackerStat FactionStat
	DefenderStat FactionStat
	Effect       AbilityEffectType
}

type AbilityStep struct {
	Effect  AbilityEffectType
	Target  TargetKind
	Contest *Contest
	Roll    *RollTable
}

type AbilityOutcome struct {
	From        int
	To          int
	Coin        int // signed: gain (positive) or loss (negative)
	DestroySelf bool
	Discount    int
}

// AbilityTarget is the step's resolved target, returned by the target stage and
// carried in the StepContext handed to an effect handler.
type AbilityTarget struct {
	Kind    TargetKind
	Faction *Faction
	Asset   *Asset
	World   string
}

// AbilityTargetCandidates bundles the selectable options for a target stage; the
// collector reads the slice matching the step's TargetKind.
type AbilityTargetCandidates struct {
	Factions []*Faction
	Assets   []*Asset
	Worlds   []string
}
```

##### Task 2 — `internal/faction/rulebook/rulebook.go`

Remove the dead coin cases from `toAbilityEffect` (D8). The parser still produces
the flat shape this commit — `convertAbility` is rewritten in Commit 2.

**Find:**
```go
func toAbilityEffect(s string) (domain.AbilityEffectType, error) {
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	case "coin_drain":
		return domain.EffectCoinDrain, nil
	case "coin_steal":
		return domain.EffectCoinSteal, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
}
```
**Replace with:**
```go
func toAbilityEffect(s string) (domain.AbilityEffectType, error) {
	switch s {
	case "reveal_stealth":
		return domain.EffectRevealStealth, nil
	default:
		return "", fmt.Errorf("unknown ability effect: %q", s)
	}
}
```

##### Task 3 — `internal/faction/engine/action/actions/ability/dispatch_test.go`

The stub test references the removed `EffectCoinDrain`. Repoint it at a fabricated
unregistered effect so it still exercises the no-handler fallthrough.

**Find:**
```go
	def := &domain.AssetDefinition{
		ID:    "SWN-W1-002",
		Name:  "Harvesters",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Effect: domain.EffectCoinDrain,
		},
	}
```
**Replace with:**
```go
	def := &domain.AssetDefinition{
		ID:    "SWN-W1-002",
		Name:  "Harvesters",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Effect: domain.AbilityEffectType("not_built"),
		},
	}
```

##### Commit message
```
refactor: reshape ability domain into steps

- add AbilityStep, Contest, RollTable, TargetKind, AbilityTarget,
  AbilityTargetCandidates to domain
- add AbilityDefinition.Steps alongside the (transitional) flat fields
- generalize AbilityOutcome with signed Coin, DestroySelf, Discount
- remove EffectCoinDrain/EffectCoinSteal and their toAbilityEffect cases
```

#### Commit 2 — `feat: parse ability steps with flat-block sugar`

##### Task 1 — `internal/faction/rulebook/rulebook.go`

Grow `abilityRecord` into the step-carrying form and rewrite `convertAbility` to
produce `Steps`. Flat `[ability]` → one-step list; `[[ability.steps]]` → explicit
list; reject both at once. Parse `Contest` (from the stat fields), `Roll` (`die` +
`outcomes`). The legacy `type` step field has no `toml` tag, so the decoder ignores
it (Decision 5 / overview Decision 5).

**Find:**
```go
type abilityRecord struct {
	AttackerStat string `toml:"attacker_stat"`
	DefenderStat string `toml:"defender_stat"`
	Effect       string `toml:"effect"`
}
```
**Replace with:**
```go
type abilityRecord struct {
	AttackerStat string              `toml:"attacker_stat"`
	DefenderStat string              `toml:"defender_stat"`
	Effect       string              `toml:"effect"`
	Target       string              `toml:"target"`
	Die          string              `toml:"die"`
	Outcomes     []outcomeRecord     `toml:"outcomes"`
	Steps        []abilityStepRecord `toml:"steps"`
}

type abilityStepRecord struct {
	Effect       string          `toml:"effect"`
	Target       string          `toml:"target"`
	AttackerStat string          `toml:"attacker_stat"`
	DefenderStat string          `toml:"defender_stat"`
	Die          string          `toml:"die"`
	Outcomes     []outcomeRecord `toml:"outcomes"`
}

type outcomeRecord struct {
	From        int  `toml:"from"`
	To          int  `toml:"to"`
	Coin        int  `toml:"coin"`
	DestroySelf bool `toml:"destroy_self"`
	Discount    int  `toml:"discount"`
}
```

**Find:**
```go
func convertAbility(r *abilityRecord) (*domain.AbilityDefinition, error) {
	if r == nil {
		return nil, nil
	}
	def := &domain.AbilityDefinition{}
	if r.AttackerStat != "" {
		stat, err := toFactionStat(r.AttackerStat)
		if err != nil {
			return nil, fmt.Errorf("attacker_stat: %w", err)
		}
		def.AttackerStat = stat
	}
	if r.DefenderStat != "" {
		stat, err := toFactionStat(r.DefenderStat)
		if err != nil {
			return nil, fmt.Errorf("defender_stat: %w", err)
		}
		def.DefenderStat = stat
	}
	if r.Effect != "" {
		effect, err := toAbilityEffect(r.Effect)
		if err != nil {
			return nil, err
		}
		def.Effect = effect
	}
	return def, nil
}
```
**Replace with:**
```go
func convertAbility(r *abilityRecord) (*domain.AbilityDefinition, error) {
	if r == nil {
		return nil, nil
	}
	if len(r.Steps) > 0 {
		if r.Effect != "" || r.AttackerStat != "" || r.DefenderStat != "" || r.Die != "" {
			return nil, fmt.Errorf("use either a flat [ability] block or [[ability.steps]], not both")
		}
		steps := make([]domain.AbilityStep, 0, len(r.Steps))
		for i, sr := range r.Steps {
			step, err := convertAbilityStep(sr.Effect, sr.Target, sr.AttackerStat, sr.DefenderStat, sr.Die, sr.Outcomes)
			if err != nil {
				return nil, fmt.Errorf("step %d: %w", i, err)
			}
			steps = append(steps, step)
		}
		return &domain.AbilityDefinition{Steps: steps}, nil
	}
	step, err := convertAbilityStep(r.Effect, r.Target, r.AttackerStat, r.DefenderStat, r.Die, r.Outcomes)
	if err != nil {
		return nil, err
	}
	return &domain.AbilityDefinition{Steps: []domain.AbilityStep{step}}, nil
}

func convertAbilityStep(effectStr, targetStr, attackerStr, defenderStr, dieStr string, outcomes []outcomeRecord) (domain.AbilityStep, error) {
	var step domain.AbilityStep
	if effectStr != "" {
		effect, err := toAbilityEffect(effectStr)
		if err != nil {
			return domain.AbilityStep{}, err
		}
		step.Effect = effect
	}
	target, err := toTargetKind(targetStr)
	if err != nil {
		return domain.AbilityStep{}, err
	}
	step.Target = target
	if attackerStr != "" || defenderStr != "" {
		attacker, err := toFactionStat(attackerStr)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("attacker_stat: %w", err)
		}
		defender, err := toFactionStat(defenderStr)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("defender_stat: %w", err)
		}
		step.Contest = &domain.Contest{AttackerStat: attacker, DefenderStat: defender}
	}
	if dieStr != "" {
		die, err := parseDice(dieStr)
		if err != nil {
			return domain.AbilityStep{}, fmt.Errorf("die: %w", err)
		}
		roll := &domain.RollTable{Die: *die}
		for _, o := range outcomes {
			roll.Outcomes = append(roll.Outcomes, domain.AbilityOutcome{
				From:        o.From,
				To:          o.To,
				Coin:        o.Coin,
				DestroySelf: o.DestroySelf,
				Discount:    o.Discount,
			})
		}
		step.Roll = roll
	}
	if step.Contest != nil && step.Roll != nil {
		return domain.AbilityStep{}, fmt.Errorf("a step has both contest and roll; they are mutually exclusive")
	}
	return step, nil
}

func toTargetKind(s string) (domain.TargetKind, error) {
	switch s {
	case "", "none":
		return domain.TargetNone, nil
	case "self":
		return domain.TargetSelf, nil
	case "faction":
		return domain.TargetFaction, nil
	case "asset":
		return domain.TargetAsset, nil
	case "world":
		return domain.TargetWorld, nil
	default:
		return "", fmt.Errorf("unknown target kind: %q", s)
	}
}
```

##### Task 2 — `internal/faction/rulebook/rulebook_test.go`

The Informers parse assertions read the flat fields, which the parser no longer
populates. Repoint them at the one-step list. (Target is asserted in Commit 4
once the TOML carries it.)

**Find:**
```go
		if asset.Ability.AttackerStat != domain.StatCunning {
			t.Errorf("attacker_stat: got %q, want %q", asset.Ability.AttackerStat, domain.StatCunning)
		}
		if asset.Ability.DefenderStat != domain.StatCunning {
			t.Errorf("defender_stat: got %q, want %q", asset.Ability.DefenderStat, domain.StatCunning)
		}
		if asset.Ability.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", asset.Ability.Effect, domain.EffectRevealStealth)
		}
```
**Replace with:**
```go
		if len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %d, want 1", len(asset.Ability.Steps))
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectRevealStealth)
		}
		if step.Contest == nil {
			t.Fatal("expected a contest on the reveal step, got nil")
		}
		if step.Contest.AttackerStat != domain.StatCunning {
			t.Errorf("attacker_stat: got %q, want %q", step.Contest.AttackerStat, domain.StatCunning)
		}
		if step.Contest.DefenderStat != domain.StatCunning {
			t.Errorf("defender_stat: got %q, want %q", step.Contest.DefenderStat, domain.StatCunning)
		}
```

##### Commit message
```
feat: parse ability steps with flat-block sugar

- grow abilityRecord into the step-carrying form (steps, target, die, outcomes)
- convertAbility produces Steps: flat block -> one-step list, [[steps]] explicit
- parse Contest from stat fields, Roll from die+outcomes; contest xor roll
- legacy `type` step field is ignored (no toml tag)
```

#### Commit 3 — `feat: per-step pipeline with generic stages`

##### Task 1 — `internal/faction/engine/action/collector.go`

Replace the reveal-specific target method with the generic one (overview/Plan
Decision 2).

**Find:**
```go
	SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error)
```
**Replace with:**
```go
	SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```

##### Task 2 — `internal/faction/engine/action/actions/ability/dispatch.go` (rewritten) — full contents:

Defines `StepContext` and `EffectHandler` (the contract Efforts 2–4 cite as
"Commit 3 Task 2"). The loop is halt-on-failure with the fence-sign from Decision 1.

```go
package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/rulebook"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

// StepContext carries the ambient inputs and the pipeline-resolved target into a
// terminal effect handler. The pipeline owns target-selection and the resolution
// gate; a handler that is called has, by halt-on-failure, already passed its gate.
type StepContext struct {
	Faction      *domain.Faction
	Asset        *domain.Asset
	Def          *domain.AssetDefinition
	Step         domain.AbilityStep
	Target       domain.AbilityTarget
	FactionState *state.FactionState
	Rulebook     *rulebook.Rulebook
}

// EffectHandler emits the mutations for one resolved step.
type EffectHandler func(StepContext) ([]domain.Mutation, error)

var handlers = map[domain.AbilityEffectType]EffectHandler{
	domain.EffectRevealStealth: revealStealth,
}

// Dispatch runs an asset's ability as a sequence of steps: resolve the step's
// target, run its gate, dispatch its terminal effect, collect the mutations.
// Assets with no ability block, or whose effect has no registered handler, fall
// through to confirmApplied (the not-yet-built effects).
func Dispatch(
	faction *domain.Faction,
	asset *domain.Asset,
	def *domain.AssetDefinition,
	collector action.Collector,
	roller domain.Roller,
	factionState *state.FactionState,
	rb *rulebook.Rulebook,
) ([]domain.Mutation, error) {
	if def.Ability == nil || len(def.Ability.Steps) == 0 {
		return confirmApplied(asset, def, collector)
	}
	var mutations []domain.Mutation
	for _, step := range def.Ability.Steps {
		handler, ok := handlers[step.Effect]
		if !ok {
			return confirmApplied(asset, def, collector)
		}
		target, err := resolveTarget(step, faction, asset, factionState, collector)
		if err != nil {
			return nil, err
		}
		passed, err := passesGate(step, faction, target, roller)
		if err != nil {
			return nil, err
		}
		if !passed {
			// deliberate: halt the chain on a failed gate, not collect-and-continue,
			// because a missed contest fizzles the rest of the ability. Multi-step
			// abilities must therefore author ungated steps (relocate, roll-table)
			// before gated ones (Effort 1 Plan, Decision 1).
			break
		}
		stepMutations, err := handler(StepContext{
			Faction:      faction,
			Asset:        asset,
			Def:          def,
			Step:         step,
			Target:       target,
			FactionState: factionState,
			Rulebook:     rb,
		})
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, stepMutations...)
	}
	return mutations, nil
}

func confirmApplied(asset *domain.Asset, def *domain.AssetDefinition, collector action.Collector) ([]domain.Mutation, error) {
	if _, err := collector.ConfirmAbilityApplied(asset, def); err != nil {
		return nil, err
	}
	return nil, nil
}
```

##### Task 3 — `internal/faction/engine/action/actions/ability/stage.go` (new) — full contents:

The generic target stage and the Contest gate, plus the `statScore`/candidate
helpers lifted out of `reveal_stealth.go`. Asset/world kinds error until 3/4.

```go
package ability

import (
	"fmt"
	"sort"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/engine/action"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func resolveTarget(
	step domain.AbilityStep,
	faction *domain.Faction,
	asset *domain.Asset,
	factionState *state.FactionState,
	collector action.Collector,
) (domain.AbilityTarget, error) {
	switch step.Target {
	case domain.TargetNone:
		return domain.AbilityTarget{Kind: domain.TargetNone}, nil
	case domain.TargetSelf:
		return domain.AbilityTarget{Kind: domain.TargetSelf, Asset: asset, Faction: faction}, nil
	case domain.TargetFaction:
		candidates := otherFactions(factionState, faction.ID)
		target, err := collector.SelectAbilityTarget(asset, step, domain.AbilityTargetCandidates{Factions: candidates})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		if target.Faction == nil {
			return domain.AbilityTarget{}, fmt.Errorf("%s: %w", step.Effect, action.ErrNoSelection)
		}
		return target, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("ability target kind %q not supported yet", step.Target)
	}
}

func passesGate(step domain.AbilityStep, faction *domain.Faction, target domain.AbilityTarget, roller domain.Roller) (bool, error) {
	if step.Contest == nil {
		// Ungated (self-move, roll-table) — always proceeds. Roll-table outcome
		// resolution lands in Effort 2.
		return true, nil
	}
	if target.Faction == nil {
		return false, fmt.Errorf("contest: target faction not resolved")
	}
	attack := roller.Roll(10) + statScore(faction, step.Contest.AttackerStat)
	defense := roller.Roll(10) + statScore(target.Faction, step.Contest.DefenderStat)
	return attack > defense, nil
}

// otherFactions returns every faction but the acting one, name-sorted. For a
// faction-target ability every opponent is a valid target.
func otherFactions(factionState *state.FactionState, actingFactionID string) []*domain.Faction {
	var candidates []*domain.Faction
	for id, faction := range factionState.Factions {
		if id == actingFactionID {
			continue
		}
		candidates = append(candidates, faction)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

func statScore(faction *domain.Faction, stat domain.FactionStat) int {
	switch stat {
	case domain.StatForce:
		return faction.Force
	case domain.StatCunning:
		return faction.Cunning
	case domain.StatWealth:
		return faction.Wealth
	default:
		return 0
	}
}
```

##### Task 4 — `internal/faction/engine/action/actions/ability/reveal_stealth.go` (rewritten) — full contents:

Shrinks to the terminal effect — target and contest are now the pipeline's.

```go
package ability

import (
	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
)

func revealStealth(ctx StepContext) ([]domain.Mutation, error) {
	targetFaction := ctx.Target.Faction
	var mutations []domain.Mutation
	for _, targetAsset := range targetFaction.Assets {
		if targetAsset.Location.WorldID == ctx.Asset.Location.WorldID && targetAsset.Stealthy {
			mutations = append(mutations, domain.AssetStealthCleared{
				FactionID:         targetFaction.ID,
				AssetID:           targetAsset.ID,
				Cause:             "ability",
				CausedByFactionID: ctx.Faction.ID,
			})
		}
	}
	return mutations, nil
}
```

##### Task 5 — `internal/faction/domain/asset.go`

Remove the now-dead transitional flat fields (Decision 5) — nothing reads them
after Tasks 2–4.

**Find:**
```go
// AbilityDefinition is a uniform sequence of steps; a flat [ability] block parses
// to a one-step list. The flat AttackerStat/DefenderStat/Die/Outcomes/Effect
// fields are transitional (removed in Commit 3) — readers use Steps.
type AbilityDefinition struct {
	Steps []AbilityStep

	// Transitional flat fields (removed in Commit 3 once dispatch reads Steps).
	Die          *DiceRoll
	Outcomes     []AbilityOutcome
	AttackerStat FactionStat
	DefenderStat FactionStat
	Effect       AbilityEffectType
}
```
**Replace with:**
```go
// AbilityDefinition is a uniform sequence of steps; a flat [ability] block parses
// to a one-step list.
type AbilityDefinition struct {
	Steps []AbilityStep
}
```

##### Task 6 — `internal/faction/tui/adapter/action_collector.go`

Replace the faction-test method with the generic one, routing the faction kind to
the existing channel (Decision 6). Ensure `fmt` is imported.

**Find:**
```go
func (a *actionCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
		Asset:      asset,
		Effect:     effect,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}
	chosen, _ := raw.(*domain.Faction)
	return chosen, nil
}
```
**Replace with:**
```go
func (a *actionCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	switch step.Target {
	case domain.TargetFaction:
		raw, err := a.adapter.ask(AskSelectFactionTestTarget, nil, SelectFactionTestTargetPayload{
			Asset:      asset,
			Effect:     step.Effect,
			Candidates: candidates.Factions,
		})
		if err != nil {
			return domain.AbilityTarget{}, err
		}
		chosen, _ := raw.(*domain.Faction)
		return domain.AbilityTarget{Kind: domain.TargetFaction, Faction: chosen}, nil
	default:
		return domain.AbilityTarget{}, fmt.Errorf("SelectAbilityTarget: target kind %q not supported yet", step.Target)
	}
}
```

##### Task 7 — `internal/faction/tui/views/turn/overlay/ability.go`

`abilityLabel` reads the removed flat `Effect`. Read the first step's effect.

**Find:**
```go
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && def.Ability.Effect != "" {
		return fmt.Sprintf("%s · %s", name, def.Ability.Effect)
	}
```
**Replace with:**
```go
	if def, ok := rb.Assets[asset.DefinitionID]; ok && def.Ability != nil && len(def.Ability.Steps) > 0 {
		return fmt.Sprintf("%s · %s", name, def.Ability.Steps[0].Effect)
	}
```

##### Task 8 — `internal/faction/tui/adapter/smoke.go`

**Find:**
```go
func (SmokeActionCollector) SelectFactionTestTarget(_ *domain.Asset, _ domain.AbilityEffectType, _ []*domain.Faction) (*domain.Faction, error) {
	return nil, action.ErrActionUnavailable
}
```
**Replace with:**
```go
func (SmokeActionCollector) SelectAbilityTarget(_ *domain.Asset, _ domain.AbilityStep, _ domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	return domain.AbilityTarget{}, action.ErrActionUnavailable
}
```

##### Task 9 — `internal/faction/engine/testharness/collector.go`

Rename the scripted hook field and rewrite the method.

**Find:**
```go
	SelectFactionTestTargetFn     func(*domain.Asset, domain.AbilityEffectType, []*domain.Faction) (*domain.Faction, error)
```
**Replace with:**
```go
	SelectAbilityTargetFn         func(*domain.Asset, domain.AbilityStep, domain.AbilityTargetCandidates) (domain.AbilityTarget, error)
```

**Find:**
```go
func (c *ScriptedCollector) SelectFactionTestTarget(asset *domain.Asset, effect domain.AbilityEffectType, candidates []*domain.Faction) (*domain.Faction, error) {
	if c.SelectFactionTestTargetFn != nil {
		return c.SelectFactionTestTargetFn(asset, effect, candidates)
	}
	return nil, nil
}
```
**Replace with:**
```go
func (c *ScriptedCollector) SelectAbilityTarget(asset *domain.Asset, step domain.AbilityStep, candidates domain.AbilityTargetCandidates) (domain.AbilityTarget, error) {
	if c.SelectAbilityTargetFn != nil {
		return c.SelectAbilityTargetFn(asset, step, candidates)
	}
	return domain.AbilityTarget{}, nil
}
```

##### Task 10 — regenerate `internal/faction/engine/action/actions/mocks/mock_collector.go`

Run `go generate ./internal/faction/engine/action/...` to regenerate `MockCollector`
against the changed interface (replaces the `SelectFactionTestTarget` mock methods
with `SelectAbilityTarget`). Do not hand-edit.

##### Task 11 — `internal/faction/engine/action/actions/ability/dispatch_test.go`

Rebuild the reveal fixture on the step shape; update the prompt expectation to
`SelectAbilityTarget`; rebuild the stub on a step.

**Find:**
```go
	def := &domain.AssetDefinition{
		ID:    "SWN-C1-002",
		Name:  "Informers",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			AttackerStat: domain.StatCunning,
			DefenderStat: domain.StatCunning,
			Effect:       domain.EffectRevealStealth,
		},
	}
```
**Replace with:**
```go
	def := &domain.AssetDefinition{
		ID:    "SWN-C1-002",
		Name:  "Informers",
		Flags: []domain.AssetFlag{domain.FlagAction},
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{{
				Effect: domain.EffectRevealStealth,
				Target: domain.TargetFaction,
				Contest: &domain.Contest{
					AttackerStat: domain.StatCunning,
					DefenderStat: domain.StatCunning,
				},
			}},
		},
	}
```

In both reveal tests, update the prompt expectation. **Find** (appears twice — apply to each):
```go
	collector.EXPECT().SelectFactionTestTarget(gomock.Any(), gomock.Any(), gomock.Any()).Return(targetFaction, nil)
```
**Replace with:**
```go
	collector.EXPECT().SelectAbilityTarget(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.AbilityTarget{Kind: domain.TargetFaction, Faction: targetFaction}, nil)
```

In the stub test, put the unregistered effect on a step. **Find:**
```go
		Ability: &domain.AbilityDefinition{
			Effect: domain.AbilityEffectType("not_built"),
		},
```
**Replace with:**
```go
		Ability: &domain.AbilityDefinition{
			Steps: []domain.AbilityStep{{
				Effect: domain.AbilityEffectType("not_built"),
				Target: domain.TargetNone,
			}},
		},
```

##### Task 12 — `internal/faction/engine/action/actions/use_asset_ability_test.go`

The hand-built def uses flat fields. Move it to the step shape.

**Find:**
```go
			Ability: &domain.AbilityDefinition{
				AttackerStat: domain.StatCunning,
				DefenderStat: domain.StatCunning,
				Effect:       domain.EffectRevealStealth,
			},
```
**Replace with:**
```go
			Ability: &domain.AbilityDefinition{
				Steps: []domain.AbilityStep{{
					Effect: domain.EffectRevealStealth,
					Target: domain.TargetFaction,
					Contest: &domain.Contest{
						AttackerStat: domain.StatCunning,
						DefenderStat: domain.StatCunning,
					},
				}},
			},
```

<!-- Re-grounding note for execution: if use_asset_ability_test scripts the
SelectFactionTestTarget hook via ScriptedCollector, repoint it at
SelectAbilityTargetFn returning an AbilityTarget. Verify against the file at
execution time. -->

##### Commit message
```
feat: per-step ability pipeline with generic stages

- Dispatch loops steps: resolve target -> gate -> terminal effect, halt-on-failure
- extract generic target stage (SelectAbilityTarget) and Contest gate into stage.go
- add StepContext + EffectHandler; revealStealth shrinks to the terminal effect
- generalize Collector.SelectFactionTestTarget -> SelectAbilityTarget across all
  five implementations (faction kind wired; asset/world error until Efforts 3/4)
- remove the transitional flat AbilityDefinition fields
```

#### Commit 4 — `feat: author Informers and Seductress reveal step`

##### Task 1 — `rulebooks/swn/assets/cunning_assets.toml`

Informers: add the explicit faction target to the flat block.

**Find:**
```toml
[assets.Informers.ability]
attacker_stat = "Cunning"
defender_stat = "Cunning"
effect = "reveal_stealth"
```
**Replace with:**
```toml
[assets.Informers.ability]
attacker_stat = "Cunning"
defender_stat = "Cunning"
effect = "reveal_stealth"
target = "faction"
```

Seductress: drop the legacy `type` discriminator (Decision 5) and add the target.
Stays a one-step list — the relocate step lands in Effort 3.

**Find:**
```toml
[[assets.Seductress.ability.steps]]
type = "faction_test"
attacker_stat = "Cunning"
defender_stat = "Cunning"
effect = "reveal_stealth"
```
**Replace with:**
```toml
[[assets.Seductress.ability.steps]]
effect = "reveal_stealth"
target = "faction"
attacker_stat = "Cunning"
defender_stat = "Cunning"
```

##### Task 2 — `internal/faction/rulebook/rulebook_test.go`

Assert the Informers reveal step now carries the faction target, and add a
Seductress test proving the `[[steps]]` path parses (and the `type` field is
ignored).

In the existing Informers subtest, after the `Contest.DefenderStat` assertion,
**add:**
```go
		if step.Target != domain.TargetFaction {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetFaction)
		}
```

After the Informers subtest closes, **add** a new subtest:
```go
	t.Run("seductress reveal step parsed from [[steps]]", func(t *testing.T) {
		asset, ok := rb.Assets["SWN-C2-004"]
		if !ok {
			t.Fatal("asset SWN-C2-004 not found")
		}
		if asset.Ability == nil || len(asset.Ability.Steps) != 1 {
			t.Fatalf("steps: got %v, want 1", asset.Ability)
		}
		step := asset.Ability.Steps[0]
		if step.Effect != domain.EffectRevealStealth {
			t.Errorf("effect: got %q, want %q", step.Effect, domain.EffectRevealStealth)
		}
		if step.Target != domain.TargetFaction {
			t.Errorf("target: got %q, want %q", step.Target, domain.TargetFaction)
		}
		if step.Contest == nil || step.Contest.AttackerStat != domain.StatCunning {
			t.Errorf("contest: got %+v, want Cunning vs Cunning", step.Contest)
		}
	})
```

##### Commit message
```
feat: author Informers and Seductress reveal step

- Informers: explicit target = "faction" on the flat ability block
- Seductress: drop legacy `type` discriminator, add target; one-step reveal list
  (relocate step deferred to Effort 3)
- rulebook tests assert the faction target and the [[steps]] parse path
```
