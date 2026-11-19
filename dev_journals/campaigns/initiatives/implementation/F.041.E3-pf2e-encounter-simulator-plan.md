# PF2e Encounter Simulator (F.041) — Effort 3 Plan: Effect Core

## In Plain Words

- **What:** the first of three rules efforts. Abilities stop being a list of Strikes. A step can now be a saving throw, a spell attack, plain damage, healing, temporary HP, a condition, or a lowered defense, and each degree of success can run more steps. The fight learns typed bonuses and penalties, the conditions that change numbers (frightened, sickened, clumsy, off-guard and the rest) with durations, persistent and splash damage, defenses keyed on a trait such as holy, and the Down row: doomed, drained, hero points and the dying limit.
- **Why:** the bosses' abilities and most of the party's kit are saves, conditions and damage riders. Without these the sim can only trade Strikes. Each rule is tested on its own, so gate 2 can close at the end of Effort 5.
- **How it fits:** Effort 3 of 8 in F.041, on branch `feature/pf2e-encounter-simulator`, after the rules split (overview Decisions 8 to 12). The [rules scope](./F.041-pf2e-encounter-simulator-rules-scope.md) says what is in; its "Last rule in E3" rows are this effort's checklist. Effort 4 adds triggers and limits on top, Effort 5 spells and tables. No real ability is converted here; Effort 6 does that.

## Context / Goal

- Discovery: [`F.041-pf2e-encounter-simulator-discovery.md`](../discovery/F.041-pf2e-encounter-simulator-discovery.md), section Rules coverage.
- Overview plan: [`F.041-pf2e-encounter-simulator-overview-plan.md`](./F.041-pf2e-encounter-simulator-overview-plan.md). Its Decisions 1 to 12 bind this plan; its Work Breakdown, Effort 3, sets the scope.
- Rules scope: [`F.041-pf2e-encounter-simulator-rules-scope.md`](./F.041-pf2e-encounter-simulator-rules-scope.md). Section 6 holds the rules calls this effort settled.
- Effort 1 and 2 plans: their Decision Records hand findings to this effort (E1 entries 4 and 5, E2 entries 2 and 5).
- Ends with: every rules-scope row whose last rule is in E3 has a test, the Fall of Luna import picks up the new import rules, and the wiki documents the effect core.

## Decisions Ratified in Planning

1. **An area catches every standing foe until Effort 7.** Robert's call. Every creature already counts as engaged with every foe; the engine asks one function, `areaTargets`, and Effort 7 replaces it with range bands and the spread dial. Rejected: the decider picks (moving parts Effort 7 rebuilds); a count per ability (a guess thrown away later).
2. **Condition rules are a Go table built from shared parts.** Robert's call. Each condition names the numbers it lowers, the bonus type, a fixed or valued penalty, whether it decays at the end of the bearer's turn, and the conditions it brings with it. Code exists only where a part cannot express the rule (concealment flat checks, doomed, drained). The ruleset as data is later work (D-2); a table lifts into files then. Rejected: a campaign TOML file now; one handler per condition.
3. **Fight-time modifiers and degree shifts are new top-level tables, `[modifier.<id>]` and `[degree_shift.<id>]`.** Robert's call. They depend on the incoming effect's traits, so the build cannot compute them, and NPCs need them too (Luna's +2 against occult). Rejected: `against` on `[build.modifier]` (PCs only).
4. **Hero points come from a `hero_points` field, set to 1 in each PC's hand layer.** Robert's call. The sheets' counts are sheet state (D-26). Effort 8's scenario may override it. Rejected: a fixed 1 in the engine; the sheet's count.
5. **Rules calls (rules scope, section 6).** Robert's calls. Hammer critical specialization lands a flier prone, with fall damage as a named approximation; Tenderizer Grenade (Greater) lowers resistances by 3; Kas dies at dying 6; Destiny Spores' save is basic; Ancestral Suspicion counts only the controlled condition.
6. **A step is recursive: each degree of success holds a list of steps.** `strike`, `save`, `spell_attack`, `damage`, `heal`, `temp_hp`, `condition` and `weaken` are step kinds; a rolled step (`strike`, `save`, `spell_attack`) carries `critical_success`, `success`, `failure` and `critical_failure` lists. On an attack roll, a critical success with no `critical_success` list runs the `success` list; a save runs only the list for its degree, as PF2e writes each degree out. A Strike carries the same lists as riders. The shape mirrors PF2e text, so Effort 6's converter fills it degree by degree. Rejected: separate rider, condition and outcome formats (three shapes for one idea).
7. **The ability schema stays version 1.** Every key is an addition and nothing written so far changes meaning (E2 Decision 14). Effort 6 approves the first draft, so a bump is still free.
8. **What a condition does to a turn is Effort 4's.** E3 builds each condition's penalties, durations and decay. Slowed, stunned, quickened, confused, immobilized, deafened and the Stand action that prone costs come with Effort 4's action economy, so those names join the table there.
9. **Durations follow the rules: a duration in rounds counts down at the start of the source's turn; `until` ends at a named turn boundary of the source or the target.** An `until` at a turn's end, applied during that same creature's turn, skips the current end, so "until the end of your next turn" means the next one. "1 minute" is 10 rounds.
10. **Typed bonuses stack in the fight among fight-time sources; the base number from the sheet or the build counts as one untyped value.** A fight-time bonus of a type the build already applies (an item bonus on a save) is written as the difference. Rejected: carrying every build-time bonus into the fight (a second stacking pass for no case in the scope).
11. **Splash hits only the primary target until Effort 7; persistent damage doubles on a critical hit; the same persistent type keeps the larger.** Splash on other creatures needs position. A miss (failure) still deals the splash; a critical failure deals nothing.
12. **Concealment is a flat check before an attack roll and before a targeted single-target save at the top of an ability.** DC 5 when the target is concealed or the attacker dazzled, DC 11 when the attacker is blinded.
13. **Immunity to an effect trait skips ability steps.** `immune` entries match an ability's traits, with `death-effects` meaning the `death` trait, so Vampiric Exsanguination does nothing to the Avatar. Strikes are not checked: their damage types already meet immunities.
14. **Drained N loses level × N HP when gained and lowers maximum HP by the same; healing a dying PC removes dying and adds 1 wounded.** As written.
15. **Critical specialization is a Go table by weapon group, fed by the build.** `[build] critical_specialization` holds the lowest rank that grants it (Kas: `master`, from his `CriticalSpecialization` rule); a weapon at or above that rank adds its group's step to its critical riders. One row: hammer.
16. **E3's campaign data is the hero points and Kas's runes.** Greater Crushing (clumsy 2 and enfeebled 2 until the end of Kas's next turn) and Greater Fearsome (frightened 2) are critical riders in Kas's hand layer (overview Decision 11). Other passives, Lilith's Protean Anatomy included, come with Effort 6's conversion. Amends the signed-off breakdown, which put Protean Anatomy in Commit 8.
17. **Heal and temporary HP steps land in Commit 5, with the damage they interact with**, not in Commit 2 as the signed-off breakdown listed.
18. **The importer's reasons follow the split.** `ReasonEffort3` becomes `ReasonLaterRule` ("needs a rule from a later effort"), `ReasonAbility` names Effort 6, and a new `ReasonHandLayer` covers rune effects typed by hand.
19. **Four phases.** Each is about one session: modifiers and saves; conditions and the Down row; damage kinds; import and wiki.
20. **Two rules calls settled at sign-off.** Robert's calls. (a) A step's own shift (Disintegrate on a critical hit) applies before the saver's degree shifts (Juggernaut), so the attacker's critical hit always bites. Rejected: the saver's shifts first. (b) Heroic Recovery is spent when the dying value would reach death, as the script does. Rejected: at the first rise in dying (spends the point on a PC who might have recovered anyway).

## Decision Record — Execution

1. **Commit 2, Task 12: two test shapes made exact.** `TestLoadReadsSteps` gives the spell attack a one-step `success` list, not an empty one, so `For(3, true)` returning that list is something the test can see. `TestStepsFromALoadedLayer` makes the cone cost 3 actions and runs one round, so the caster uses it once and "one `SaveRolled` per PC" is an exact count.
2. **Commit 2: `show` prints every ability step as `strike <id>`.** A save, spell attack or damage step prints as a bare `strike `. No commit in this plan changes the step line, and no real creature carries such a step until Effort 6 converts abilities, so nothing in E3's data shows it. Deferred: a step-text printer, placed in a later E3 commit or in Effort 6.
3. **Commit 3: prone also lowers spell attacks.** Robert's call. The planned row listed only `attack`, which the engine reads for Strikes alone; spell attacks read `spell-attack`, so a prone caster's spell attacks took no penalty. PF2e's prone gives –2 to all attack rolls. Prone's selectors are now `attack` and `spell-attack`, and `TestConditionPenalties` checks the spell attack. Task 4 and Task 9 are amended to match.
4. **Commit 3: condition expiry is logged before `TurnStarted`.** `takeTurn` runs `startTurn` before `act`, which emits `TurnStarted`, so a condition that ends at a turn's start shows in the log ahead of the turn it belongs to. Harmless now; Effort 8's replay review may want the order flipped.
5. **Commit 4: `untyped` becoming a damage type changes the importer.** `TestImportPCSkipsWeaponsTheLoaderRejects` and `TestImportNPCSkipsDamageTheLoaderRejects` used `untyped` as their unsupported type; they now use `chaotic`, a legacy type the loader still rejects. The importer now keeps untyped damage it used to skip (the test actors avatar-of-lilith, tavish and arabesque carry some). Commit 8's Fall of Luna import shows the effect on the real layers.
6. **Commit 5: `TestWeaknessSkipsAZeroPart` made exact.** The planned text had the slashing part "raised to 1 by the minimum" beside a fire part that rolls. The minimum of 1 applies to the whole roll, so with fire rolling 3 the slashing part stays 0 and takes no weakness. The test is now two cases: a lone `1d4-5 slashing` part rolls 1, is raised to 1 and takes weakness 5 for 6; `1d4-5 slashing` beside `1d6 fire` leaves slashing at 0 (no weakness) and fire at 3. Task 9 is amended to match.
7. **Commit 5: two small gaps, deferred.** (a) `takePersistentDamage` still rolls the flat check for the instance whose damage killed the bearer, so the log shows one `FlatChecked` on a dead creature. Harmless; Effort 8's replay review may want it skipped. (b) A basic save keeps the roll's persistent parts on a success or a failure and does not double them on a critical failure. No E3 data has persistent damage on a basic save; Effort 6's conversion puts persistent damage in the degree lists, as PF2e writes it.
8. **Commit 6: three test edits the plan did not list.** (a) `TestUntypedDamageTakesResistAll` in `down_test.go` (Commit 4) also called the old `adjustedDamage`; it now builds a combatant like the `rules_test.go` pair. `TestAdjustedDamage` shares one defense across its cases, so it builds one combatant, not one per case. (b) Splitting "value and duration belong on a condition" into two messages changed the `value on a save` case of `TestLoadRejectsBadConditions` to `value belongs on a condition`. (c) `TestCriticalSpecializationFromBuild` adds an advanced hammer (untrained, so below master) beside the dagger, so the rank gate has a test as well as the group lookup. Task 9 is amended to match.
9. **Commit 6: the critical specialization row is shared by reference.** `riders` copies the hammer `Step` by value, but its `Failure` and `CriticalFailure` slices point at the package-level table, and `checkStep` validates rider steps in place. Safe now: a condition step writes nothing during validation. A later row whose steps carry damage, heal or temp HP would have its parsed rolls written into the shared table on every load. For the Effort 3 close review: clone the row deeply in `riders` if Effort 4 or 6 adds such a row.
10. **Commit 7: a conditional `AdjustModifier` is named by its selectors, not its slug.** Bands of Force (Greater) has no `slug`, so the planned text printed a bare `AdjustModifier `. The conditional skip now uses the same `AdjustModifier <selectors>` text as the unsupported skip below it; Calculated Splash reads `AdjustModifier alchemical-bomb-damage`. Task 4 is amended to match.
11. **Commit 7: three test edits the plan did not list.** (a) `TestImportPCRuleElementsTavish` expects the new `expanded-splash` build modifier. (b) `TestImportPCRuleElementsKas` checks that Bravery and Toughness are no longer in the report at all, not only that their old reason is gone. (c) `TestImportPCWeaponsAndQuantities` keeps its Alchemist's Fire row under `ReasonLaterRule`: the bomb is still reported for its untracked quantity. `TestImportBombDamage` checks that no persistent or splash skip remains.
12. **Commit 8: the wiki's stale effort numbers are fixed with the E3 pages.** The pf2e pages still used the numbers from before the rules split (overview Decision 8): the runner in Effort 6, tactics and position in Effort 5, the approval cache and `check` in Effort 4. They now read Effort 8, Effort 7 and Effort 6. Task 6 named only the two `overview.md` rows. `engine.md`'s Purpose and Not modelled yet sections were rewritten for the effect core, beyond the outline's new sections.
13. **Commit 8: the `data-files.md` example ability is generic.** The plan asked for an example with an area, a basic save and a `failure` list; it is named Spore Cloud, not Destiny Spores, so it does not pass for Luna's real ability, which Effort 6 converts. The example was loaded through `creature.Load` in a throwaway test before it went in the page.

---

## Shared Context

### Gate

Every commit passes `go build ./... && go vet ./... && go test ./...` before it lands. Commit 8 also runs the import by hand (its tasks).

### Test conventions

- `creature` tests stay black-box (`package creature_test`) and write layers with `writeLayers` (`layers_test.go`).
- `engine` tests stay white-box (`package engine`) and drive rules with the scripted roller (`helpers_test.go`). Commit 2 adds two helpers there, `damagePart` and `giveAbility`.
- Gate 2 for an E3 row means a test that exercises the rule. Abilities in engine tests are built in Go; each phase also loads one hand-written layer from `t.TempDir()` through `creature.Load` and runs it, so the TOML shape is proven end to end.
- `foundry` tests stay white-box and read only `testdata/`.

### Running the app

Claude runs the binary only against `hawk-claude` and passes `--campaign hawk-claude`. Build it into the scratchpad: `go build -o "$SCRATCH/gm-toolkit" ./cmd/gm-toolkit`. Robert runs the import on `hawk` (E2 Decision 11).

### Files this effort creates

```
internal/pf2e/sim/creature/
  fight.go, fight_test.go                    (Commit 1)
  steps.go, steps_test.go                    (Commit 2)
  conditions.go                              (Commit 3)
internal/pf2e/sim/engine/
  checks.go, steps.go, steps_test.go         (Commit 2)
  conditions.go, conditions_test.go          (Commit 3)
  down_test.go                               (Commit 4)
  persistent.go, healing.go, damage_test.go  (Commit 5)
  defenses_test.go                           (Commit 6)
internal/pf2e/sim/foundry/
  effects.go, effects_test.go                (Commit 7)
campaigns/hawk/pf2e/party/{kas,tavish,arabesque}/hand.toml (Commit 8)
```

### Effort close review addition

`/gestalt:effort-close F.041.E3` also checks that E3's shapes do not box in the later efforts:

- **Effort 4:** reactions need a trigger hook in `strike`, `saveStep` and `applyDamage`; the condition table takes slowed, stunned and quickened; `effectSource` can carry a fortune or misfortune flag.
- **Effort 5:** a spell is an ability with a rank; `spell_attack` and `dc_from = "spell"` already read `[spellcasting]`.
- **Effort 6:** every step kind and field has a short, unambiguous TOML spelling a converter can write.
- **Effort 7:** `areaTargets` and the splash rule are the two places position replaces.

A finding that needs a change now goes in the review-fix commit; one that can wait goes in the Decision Record for the effort that will need it.

## Out of Scope

- **Triggers and limits** (Effort 4): reactions, fast healing, recharge, frequency, slowed, stunned, quickened, confused, the Stand action, Sustain.
- **Spells** (Effort 5): slots, ranks, heightening, counteract, incapacitation, random tables, the stupefied flat check to cast.
- **Converting any real ability** (Effort 6), and every other passive's hand-layer data.
- **Position** (Effort 7): splash on other creatures, fall damage, reach, range.
- **The approximation lines in the run report** (Effort 8); each one named here goes in the rules scope until then.
- **`DamageDice`** rule elements: every one on the sheets belongs to an Effort 4 action (Vicious Swing, Brutal Finish, the relics).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — Modifiers and saves**
  - Commit 1 — `feat(E3): add fight modifiers and degree shifts to the creature format` (4 tasks)
  - Commit 2 — `feat(E3): run save, spell attack and damage steps` (12 tasks)
- [x] **Phase 2 — Conditions and the Down row**
  - Commit 3 — `feat(E3): apply conditions with durations` (9 tasks)
  - Commit 4 — `feat(E3): add doomed, drained, hero points and dying limits` (6 tasks)
- [x] **Phase 3 — Damage kinds**
  - Commit 5 — `feat(E3): add persistent and splash damage, healing and temp HP` (9 tasks)
  - Commit 6 — `feat(E3): key defenses on traits and add Strike riders` (9 tasks)
- [x] **Phase 4 — Import and wiki**
  - Commit 7 — `feat(E3): import degree shifts, dying limits and damage categories` (7 tasks)
  - Commit 8 — `docs(E3): document the effect core and add the party's hand layers` (7 tasks)

### Phase 1 — Modifiers and saves

**In plain words:** creatures can carry bonuses that depend on the incoming effect, and abilities can force saving throws, make spell attacks and deal damage with no roll. When the phase lands, an ability like Destiny Spores (a basic Will save in a cone) runs in a test.

#### Commit 1 — `feat(E3): add fight modifiers and degree shifts to the creature format`

##### Task 1 — `internal/pf2e/sim/creature/fight.go` (new)

Full contents:

```go
package creature

import (
	"maps"
	"slices"
	"strings"
)

var (
	DegreeWords      = []string{"critical-failure", "failure", "success", "critical-success"}
	CheckSelectors   = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "melee-attack", "ranged-attack", "damage", "melee-damage", "spell-attack", "spell-dc", "class-dc"}
	saveSelectors    = []string{"fortitude", "reflex", "will", "saving-throw"}
	againstSelectors = []string{"ac", "fortitude", "reflex", "will", "saving-throw"}
)

type CheckModifier struct {
	Selector string   `toml:"selector"`
	Type     string   `toml:"type,omitempty"`
	Value    int      `toml:"value"`
	Against  []string `toml:"against,omitempty"`
}

type DegreeShift struct {
	Selector string   `toml:"selector"`
	Against  []string `toml:"against,omitempty"`
	From     string   `toml:"from"`
	To       string   `toml:"to"`
}

func (modifier CheckModifier) Applies(selectors, traits []string) bool {
	return slices.Contains(selectors, modifier.Selector) && matchesAgainst(modifier.Against, traits)
}

func (shift DegreeShift) Applies(selectors, traits []string) bool {
	return slices.Contains(selectors, shift.Selector) && matchesAgainst(shift.Against, traits)
}

func matchesAgainst(against, traits []string) bool {
	return len(against) == 0 || slices.ContainsFunc(against, func(trait string) bool { return slices.Contains(traits, trait) })
}

func (creature *Creature) DegreeShiftIDs() []string {
	return slices.Sorted(maps.Keys(creature.DegreeShifts))
}

func (creature *Creature) checkFightRules(report func(string, ...any)) {
	for _, id := range slices.Sorted(maps.Keys(creature.Modifiers)) {
		modifier := creature.Modifiers[id]
		if !slices.Contains(CheckSelectors, modifier.Selector) {
			report("modifier.%s: selector %q: want %s", id, modifier.Selector, strings.Join(CheckSelectors, ", "))
		}
		if !slices.Contains(modifierTypes, modifier.Type) {
			report("modifier.%s: type %q: want item, status, circumstance or none", id, modifier.Type)
		}
		if modifier.Value == 0 {
			report("modifier.%s: value must not be 0", id)
		}
		if len(modifier.Against) > 0 && !slices.Contains(againstSelectors, modifier.Selector) {
			report("modifier.%s: against needs selector ac or a save", id)
		}
	}
	for _, id := range creature.DegreeShiftIDs() {
		shift := creature.DegreeShifts[id]
		if !slices.Contains(saveSelectors, shift.Selector) {
			report("degree_shift.%s: selector %q: want %s", id, shift.Selector, strings.Join(saveSelectors, ", "))
		}
		if !slices.Contains(DegreeWords, shift.From) || !slices.Contains(DegreeWords, shift.To) {
			report("degree_shift.%s: from %q and to %q must each be one of %s", id, shift.From, shift.To, strings.Join(DegreeWords, ", "))
		}
	}
}
```

##### Task 2 — `internal/pf2e/sim/creature/creature.go` and `validate.go`

In `Creature`, add the two tables after `Build`.

**Find:**
```go
	Build        *Build             `toml:"build,omitempty"`
	Strikes      map[string]Strike  `toml:"strike,omitempty"`
```
**Replace with:**
```go
	Build        *Build                   `toml:"build,omitempty"`
	Modifiers    map[string]CheckModifier `toml:"modifier,omitempty"`
	DegreeShifts map[string]DegreeShift   `toml:"degree_shift,omitempty"`
	Strikes      map[string]Strike        `toml:"strike,omitempty"`
```

In `prepare` (`validate.go`), check the new tables before the Strikes.

**Find:**
```go
	for _, id := range creature.StrikeIDs() {
		strike := creature.Strikes[id]
```
**Replace with:**
```go
	creature.checkFightRules(report)

	for _, id := range creature.StrikeIDs() {
		strike := creature.Strikes[id]
```

##### Task 3 — `cmd/gm-toolkit/pf2e/show.go`

In `printCreature`, print the new tables before the layers line.

**Find:**
```go
	fmt.Fprintf(out, "Layers %s\n", strings.Join(shown.Layers, " + "))
}
```
**Replace with:**
```go
	printFightRules(out, shown, bold, reset)
	fmt.Fprintf(out, "Layers %s\n", strings.Join(shown.Layers, " + "))
}

func printFightRules(out io.Writer, shown *creature.Creature, bold, reset string) {
	if len(shown.Modifiers) > 0 {
		fmt.Fprintf(out, "%sModifiers%s\n", bold, reset)
		for _, id := range slices.Sorted(maps.Keys(shown.Modifiers)) {
			modifier := shown.Modifiers[id]
			line := fmt.Sprintf("  %s  %+d", id, modifier.Value)
			if modifier.Type != "" {
				line += " " + modifier.Type
			}
			line += " " + modifier.Selector
			if len(modifier.Against) > 0 {
				line += " against " + strings.Join(modifier.Against, ", ")
			}
			fmt.Fprintln(out, line)
		}
	}
	if len(shown.DegreeShifts) > 0 {
		fmt.Fprintf(out, "%sDegree shifts%s\n", bold, reset)
		for _, id := range shown.DegreeShiftIDs() {
			shift := shown.DegreeShifts[id]
			line := fmt.Sprintf("  %s  %s %s -> %s", id, shift.Selector, shift.From, shift.To)
			if len(shift.Against) > 0 {
				line += " against " + strings.Join(shift.Against, ", ")
			}
			fmt.Fprintln(out, line)
		}
	}
}
```

##### Task 4 — tests

`internal/pf2e/sim/creature/fight_test.go` (new, `package creature_test`):

- **`TestLoadReadsFightRules`** — a hand layer (`minimalHand` plus a `[modifier.occult-saves]` with selector `saving-throw`, type `status`, value 2, against `["occult"]`, and a `[degree_shift.juggernaut]` with selector `fortitude`, from `success`, to `critical-success`) loads; both tables hold those values; `Applies([]string{"will", "saving-throw"}, []string{"occult"})` is true and `Applies(..., []string{"arcane"})` false.
- **`TestLoadRejectsBadFightRules`** — table cases, each failing with `ErrInvalid` and the quoted text: unknown selector (`selector "speed"`), unknown type (`type "luck"`), value 0 (`value must not be 0`), `against` on `attack` (`against needs selector ac or a save`), degree shift on `ac` (`degree_shift.x: selector "ac"`), bad degree word (`from "great"`).

`cmd/gm-toolkit/pf2e/show_test.go`: **`TestShowPrintsFightRules`** — a creature with one modifier and one degree shift prints `  occult-saves  +2 status saving-throw against occult` under `Modifiers` and `  juggernaut  fortitude success -> critical-success` under `Degree shifts`, before the `Layers` line.

##### Commit message

```
feat(E3): add fight modifiers and degree shifts to the creature format

- [modifier.<id>]: a typed bonus or penalty on a check or DC, optionally only against effects with given traits
- [degree_shift.<id>]: a save result that moves to another degree, optionally only against given traits
- load checks both tables; show prints them
```

#### Commit 2 — `feat(E3): run save, spell attack and damage steps`

##### Task 1 — `internal/pf2e/sim/creature/steps.go` (new)

Full contents:

```go
package creature

import (
	"fmt"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const (
	StepStrike      = "strike"
	StepSave        = "save"
	StepSpellAttack = "spell_attack"
	StepDamage      = "damage"
)

var (
	SaveNames   = []string{"fortitude", "reflex", "will"}
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage}
	rolledSteps = []string{StepStrike, StepSave, StepSpellAttack}
	areaShapes  = []string{"burst", "cone", "emanation", "line"}
	dcSources   = []string{"class", "spell"}
)

type Area struct {
	Shape string `toml:"shape"`
	Size  int    `toml:"size"`
}

type Outcomes struct {
	CriticalSuccess []Step `toml:"critical_success,omitempty"`
	Success         []Step `toml:"success,omitempty"`
	Failure         []Step `toml:"failure,omitempty"`
	CriticalFailure []Step `toml:"critical_failure,omitempty"`
}

type Step struct {
	Strike      string       `toml:"strike,omitempty"`
	Save        string       `toml:"save,omitempty"`
	SpellAttack bool         `toml:"spell_attack,omitempty"`
	Damage      []DamagePart `toml:"damage,omitempty"`
	Target      string       `toml:"target,omitempty"`
	DC          int          `toml:"dc,omitzero"`
	DCFrom      string       `toml:"dc_from,omitempty"`
	Basic       bool         `toml:"basic,omitempty"`
	Shift       int          `toml:"shift,omitzero"`
	Outcomes
}

// For takes a degree as the engine numbers it: 0 critical failure to 3 critical success.
func (outcomes Outcomes) For(degree int, attack bool) []Step {
	if attack && degree == 3 && len(outcomes.CriticalSuccess) == 0 {
		return outcomes.Success
	}
	return outcomes.lists()[degree]
}

func (outcomes Outcomes) lists() [][]Step {
	return [][]Step{outcomes.CriticalFailure, outcomes.Failure, outcomes.Success, outcomes.CriticalSuccess}
}

func (step Step) Kind() string {
	if heads := step.heads(); len(heads) == 1 {
		return heads[0]
	}
	return ""
}

func (step Step) heads() []string {
	var heads []string
	if step.Strike != "" {
		heads = append(heads, StepStrike)
	}
	if step.Save != "" {
		heads = append(heads, StepSave)
	}
	if step.SpellAttack {
		heads = append(heads, StepSpellAttack)
	}
	if len(step.Damage) > 0 && step.Save == "" && !step.SpellAttack {
		heads = append(heads, StepDamage)
	}
	return heads
}

func (creature *Creature) checkStep(path string, step *Step, report func(string, ...any)) {
	heads := step.heads()
	if len(heads) != 1 {
		report("%s: want exactly one of %s; got %d", path, strings.Join(stepKinds, ", "), len(heads))
		return
	}
	kind := heads[0]
	switch kind {
	case StepStrike:
		if _, found := creature.Strikes[step.Strike]; !found {
			report("%s: unknown strike %q", path, step.Strike)
		}
	case StepSave:
		if !slices.Contains(SaveNames, step.Save) {
			report("%s: save %q: want %s", path, step.Save, strings.Join(SaveNames, ", "))
		}
		creature.checkDC(path, step, report)
		if len(step.Damage) > 0 && !step.Basic {
			report("%s: damage on a save needs basic = true; put other damage in the degree lists", path)
		}
		if step.Basic && len(step.Damage) == 0 {
			report("%s: basic needs damage", path)
		}
	case StepSpellAttack:
		if creature.Spellcasting == nil {
			report("%s: spell_attack needs [spellcasting]", path)
		}
	}
	if kind != StepSave && (step.DC != 0 || step.DCFrom != "" || step.Basic || step.Shift != 0) {
		report("%s: dc, dc_from, basic and shift belong on a save", path)
	}
	if step.Shift < -3 || step.Shift > 3 {
		report("%s: shift %d: want -3 to 3", path, step.Shift)
	}
	if step.Target != "" && step.Target != "self" {
		report("%s: target %q: want self, or leave it unset", path, step.Target)
	}
	prepareDamage(path, step.Damage, report)
	creature.checkOutcomes(path, step.Outcomes, slices.Contains(rolledSteps, kind), report)
}

func (creature *Creature) checkOutcomes(path string, outcomes Outcomes, rolled bool, report func(string, ...any)) {
	for degree, list := range outcomes.lists() {
		word := strings.ReplaceAll(DegreeWords[degree], "-", "_")
		if len(list) > 0 && !rolled {
			report("%s: %s needs a strike, save or spell_attack", path, word)
		}
		for index := range list {
			creature.checkStep(fmt.Sprintf("%s.%s[%d]", path, word, index), &list[index], report)
		}
	}
}

func (creature *Creature) checkDC(path string, step *Step, report func(string, ...any)) {
	switch {
	case step.DC > 0 && step.DCFrom != "":
		report("%s: set dc or dc_from, not both", path)
	case step.DC > 0:
	case !slices.Contains(dcSources, step.DCFrom):
		report("%s: a save needs dc, or dc_from class or spell", path)
	case step.DCFrom == "class" && creature.ClassDC < 1:
		report("%s: dc_from class needs a class DC", path)
	case step.DCFrom == "spell" && creature.Spellcasting == nil:
		report("%s: dc_from spell needs [spellcasting]", path)
	}
}

func checkArea(path string, area *Area, report func(string, ...any)) {
	if area == nil {
		return
	}
	if !slices.Contains(areaShapes, area.Shape) {
		report("%s.area: shape %q: want %s", path, area.Shape, strings.Join(areaShapes, ", "))
	}
	if area.Size < 5 {
		report("%s.area: size must be at least 5", path)
	}
}

func prepareDamage(path string, parts []DamagePart, report func(string, ...any)) {
	for index := range parts {
		part := &parts[index]
		expression, err := dice.Parse(part.Dice)
		if err != nil {
			report("%s.damage[%d]: %v", path, index, err)
		}
		part.Roll = expression
		if !IsDamageType(part.Type) {
			report("%s.damage[%d]: unknown damage type %q", path, index, part.Type)
		}
	}
}
```

##### Task 2 — `internal/pf2e/sim/creature/creature.go`

`Ability` gains traits and an area; `Step` moves to `steps.go`.

**Find:**
```go
type Ability struct {
	ID      string `toml:"-"`
	Name    string `toml:"name,omitempty"`
	Actions int    `toml:"actions,omitzero"`
	Ignore  bool   `toml:"ignore,omitempty"`
	Steps   []Step `toml:"steps,omitempty"`
}

type Step struct {
	Strike string `toml:"strike"`
}
```
**Replace with:**
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

##### Task 3 — `internal/pf2e/sim/creature/validate.go`

Strike damage goes through `prepareDamage`; ability steps through `checkStep`.

**Find:**
```go
		if len(strike.Damage) == 0 {
			report("strike.%s: no damage", id)
		}
		for index := range strike.Damage {
			part := &strike.Damage[index]
			expression, err := dice.Parse(part.Dice)
			if err != nil {
				report("strike.%s.damage[%d]: %v", id, index, err)
			}
			part.Roll = expression
			if !IsDamageType(part.Type) {
				report("strike.%s.damage[%d]: unknown damage type %q", id, index, part.Type)
			}
		}
		creature.Strikes[id] = strike
```
**Replace with:**
```go
		if len(strike.Damage) == 0 {
			report("strike.%s: no damage", id)
		}
		prepareDamage("strike."+id, strike.Damage, report)
		creature.Strikes[id] = strike
```

**Find:**
```go
		if len(ability.Steps) == 0 {
			report("ability.%s: no steps", id)
		}
		for index, step := range ability.Steps {
			if _, found := creature.Strikes[step.Strike]; !found {
				report("ability.%s.steps[%d]: unknown strike %q", id, index, step.Strike)
			}
		}
```
**Replace with:**
```go
		if len(ability.Steps) == 0 {
			report("ability.%s: no steps", id)
		}
		checkArea("ability."+id, ability.Area, report)
		for index := range ability.Steps {
			creature.checkStep(fmt.Sprintf("ability.%s.steps[%d]", id, index), &ability.Steps[index], report)
		}
```

Remove the now-unused `dice` import from `validate.go`.

##### Task 4 — `internal/pf2e/sim/creature/build.go`

Export the bonus pool so the engine stacks fight bonuses the same way.

**Find:**
```go
type bonusPool struct {
	untyped   int
	bonuses   map[string]int
	penalties map[string]int
}

func newBonusPool() *bonusPool {
	return &bonusPool{bonuses: map[string]int{}, penalties: map[string]int{}}
}

func (pool *bonusPool) add(kind string, value int) {
```
**Replace with:**
```go
type BonusPool struct {
	untyped   int
	bonuses   map[string]int
	penalties map[string]int
}

func NewBonusPool() *BonusPool {
	return &BonusPool{bonuses: map[string]int{}, penalties: map[string]int{}}
}

func (pool *BonusPool) Add(kind string, value int) {
```

**Find:**
```go
func (pool *bonusPool) total() int {
```
**Replace with:**
```go
func (pool *BonusPool) Total() int {
```

In `Build.total`:

**Find:**
```go
	pool := newBonusPool()
	pool.add("item", itemBonus)
	for _, id := range slices.Sorted(maps.Keys(build.Modifiers)) {
		modifier := build.Modifiers[id]
		if !slices.Contains(selectors, modifier.Selector) || !containsAll(weaponTraits, modifier.WeaponTraits) {
			continue
		}
		pool.add(modifier.Type, modifier.Value+modifier.PerLevel*level)
	}
	return pool.total()
```
**Replace with:**
```go
	pool := NewBonusPool()
	pool.Add("item", itemBonus)
	for _, id := range slices.Sorted(maps.Keys(build.Modifiers)) {
		modifier := build.Modifiers[id]
		if !slices.Contains(selectors, modifier.Selector) || !containsAll(weaponTraits, modifier.WeaponTraits) {
			continue
		}
		pool.Add(modifier.Type, modifier.Value+modifier.PerLevel*level)
	}
	return pool.Total()
```

##### Task 5 — `internal/pf2e/sim/engine/checks.go` (new)

Full contents:

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var saveSelectors = map[string][]string{
	"fortitude": {"fortitude", "saving-throw"},
	"reflex":    {"reflex", "saving-throw"},
	"will":      {"will", "saving-throw"},
}

func (fight *Fight) modifierTotal(combatant *Combatant, selectors, traits []string) int {
	return fight.modifierPool(combatant, selectors, traits).Total()
}

func (fight *Fight) modifierPool(combatant *Combatant, selectors, traits []string) *creature.BonusPool {
	pool := creature.NewBonusPool()
	for _, modifier := range combatant.Creature.Modifiers {
		if modifier.Applies(selectors, traits) {
			pool.Add(modifier.Type, modifier.Value)
		}
	}
	return pool
}

func (fight *Fight) armorClass(target *Combatant, traits []string) int {
	return target.Creature.Defense.AC + fight.modifierTotal(target, []string{"ac"}, traits)
}

func saveBonus(defense creature.Defense, save string) int {
	switch save {
	case "fortitude":
		return defense.Fortitude
	case "reflex":
		return defense.Reflex
	default:
		return defense.Will
	}
}

func (fight *Fight) savingThrow(target *Combatant, save string, dc, shift int, source effectSource) Degree {
	selectors := saveSelectors[save]
	bonus := saveBonus(target.Creature.Defense, save) + fight.modifierTotal(target, selectors, source.traits)
	die := fight.roller.Die(20)
	total := die + bonus
	degree := shiftDegree(DegreeOf(die, total, dc), shift)
	degree = applyDegreeShifts(target, selectors, source.traits, degree)
	fight.observer.Observe(SaveRolled{
		Combatant: target.ID,
		Source:    source.actor.ID,
		Ability:   source.ability,
		Save:      save,
		Die:       die,
		Bonus:     bonus,
		Total:     total,
		DC:        dc,
		Degree:    degree,
	})
	return degree
}

func shiftDegree(degree Degree, steps int) Degree {
	return min(max(degree+Degree(steps), CriticalFailure), CriticalSuccess)
}

// deliberate: a step's own shift applies before the saver's degree shifts, not after, because Robert chose that order at the E3 plan sign-off (Decision 20a).
func applyDegreeShifts(combatant *Combatant, selectors, traits []string, degree Degree) Degree {
	for _, id := range combatant.Creature.DegreeShiftIDs() {
		shift := combatant.Creature.DegreeShifts[id]
		if shift.Applies(selectors, traits) && Degree(slices.Index(creature.DegreeWords, shift.From)) == degree {
			return Degree(slices.Index(creature.DegreeWords, shift.To))
		}
	}
	return degree
}
```

##### Task 6 — `internal/pf2e/sim/engine/steps.go` (new)

Full contents:

```go
package engine

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

type effectSource struct {
	actor   *Combatant
	ability string
	strike  string
	traits  []string
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

// deliberate: an area catches every standing foe, not a share of them, because every creature counts as engaged with every foe until Effort 7's position seam replaces this (Decision 1).
func (fight *Fight) areaTargets(actor *Combatant) []*Combatant {
	return fight.StandingFoes(actor)
}

func (fight *Fight) runSteps(source effectSource, steps []creature.Step, target *Combatant) {
	for _, step := range steps {
		fight.runStep(source, step, []*Combatant{target})
	}
}

func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	if step.Target == "self" {
		targets = []*Combatant{source.actor}
	}
	switch step.Kind() {
	case creature.StepStrike:
		strike := source.actor.Creature.Strikes[step.Strike]
		for _, target := range targets {
			if target.Standing() {
				fight.strike(source.actor, target, strike, source.ability)
			}
		}
	case creature.StepSave:
		fight.saveStep(source, step, targets)
	case creature.StepSpellAttack:
		for _, target := range targets {
			if target.Standing() {
				fight.spellAttack(source, step, target)
			}
		}
	case creature.StepDamage:
		rolled := fight.rollDamage(step.Damage, false, 0)
		for _, target := range targets {
			if !target.Dead {
				fight.applyDamage(source, target, rolled, false)
			}
		}
	}
}

func (fight *Fight) saveStep(source effectSource, step creature.Step, targets []*Combatant) {
	dc := fight.stepDC(source.actor, step)
	var rolled damageRoll
	if step.Basic {
		rolled = fight.rollDamage(step.Damage, false, 0)
	}
	for _, target := range targets {
		if target.Dead {
			continue
		}
		degree := fight.savingThrow(target, step.Save, dc, step.Shift, source)
		if step.Basic && degree < CriticalSuccess {
			fight.applyDamage(source, target, rolled.basic(degree), degree == CriticalFailure)
		}
		fight.runSteps(source, step.For(int(degree), false), target)
	}
}

func (fight *Fight) stepDC(actor *Combatant, step creature.Step) int {
	switch step.DCFrom {
	case "class":
		return actor.Creature.ClassDC + fight.modifierTotal(actor, []string{"class-dc"}, nil)
	case "spell":
		return actor.Creature.Spellcasting.DC + fight.modifierTotal(actor, []string{"spell-dc"}, nil)
	default:
		return step.DC + fight.modifierTotal(actor, []string{"class-dc"}, nil)
	}
}

func (fight *Fight) spellAttack(source effectSource, step creature.Step, target *Combatant) {
	attacker := source.actor
	penalty := attackPenalty(attacker.AttacksThisTurn, false)
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
	total := die + bonus + penalty
	armorClass := fight.armorClass(target, source.traits)
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(SpellAttackRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Ability:  source.ability,
		Die:      die,
		Bonus:    bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	critical := degree == CriticalSuccess
	if degree >= Success && len(step.Damage) > 0 {
		fight.applyDamage(source, target, fight.rollDamage(step.Damage, critical, 0), critical)
	}
	if !target.Dead {
		fight.runSteps(source, step.For(int(degree), true), target)
	}
}
```

##### Task 7 — `internal/pf2e/sim/engine/damage.go`

Damage is rolled once into a `damageRoll` and then applied, so an area rolls once and a basic save scales the same roll. Replace `dealDamage` whole.

**Find:**
```go
func (fight *Fight) dealDamage(attacker, target *Combatant, strike creature.Strike, critical bool) {
	rolledByType := map[string]int{}
	var types []string
	for _, part := range strike.Damage {
		amount := max(part.Roll.Roll(fight.roller), 0)
		if critical {
			amount *= 2
		}
		if _, seen := rolledByType[part.Type]; !seen {
			types = append(types, part.Type)
		}
		rolledByType[part.Type] += amount
	}

	parts := make([]DamageAmount, 0, len(types))
	total := 0
	for _, damageType := range types {
		rolled := rolledByType[damageType]
		taken := adjustedDamage(target.Creature.Defense, damageType, strike.Traits, rolled)
		parts = append(parts, DamageAmount{Type: damageType, Rolled: rolled, Taken: taken})
		total += taken
	}

	wasStanding := target.Standing()
	target.HP = max(target.HP-total, 0)
	fight.observer.Observe(DamageDealt{
		Source:   attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		HPAfter:  target.HP,
	})
	if total == 0 {
		return
	}
	switch {
	case wasStanding && target.HP == 0:
		fight.knockDown(target, attacker, critical)
	case !wasStanding:
		fight.hitWhileDown(target, critical)
	}
}
```
**Replace with:**
```go
type rolledPart struct {
	damageType string
	amount     int
}

type damageRoll struct {
	parts []rolledPart
}

func (fight *Fight) rollDamage(parts []creature.DamagePart, critical bool, bonus int) damageRoll {
	var roll damageRoll
	for index, part := range parts {
		amount := part.Roll.Roll(fight.roller)
		if index == 0 {
			amount += bonus
		}
		amount = max(amount, 0)
		if critical {
			amount *= 2
		}
		roll.parts = addRolled(roll.parts, part.Type, amount)
	}
	return roll
}

func addRolled(parts []rolledPart, damageType string, amount int) []rolledPart {
	for index := range parts {
		if parts[index].damageType == damageType {
			parts[index].amount += amount
			return parts
		}
	}
	return append(parts, rolledPart{damageType: damageType, amount: amount})
}

func (roll damageRoll) basic(degree Degree) damageRoll {
	scaled := damageRoll{parts: make([]rolledPart, len(roll.parts))}
	for index, part := range roll.parts {
		amount := part.amount
		switch degree {
		case Success:
			amount /= 2
		case CriticalFailure:
			amount *= 2
		}
		scaled.parts[index] = rolledPart{damageType: part.damageType, amount: amount}
	}
	return scaled
}

func (fight *Fight) applyDamage(source effectSource, target *Combatant, roll damageRoll, critical bool) {
	parts := make([]DamageAmount, 0, len(roll.parts))
	total := 0
	for _, part := range roll.parts {
		taken := adjustedDamage(target.Creature.Defense, part.damageType, source.traits, part.amount)
		parts = append(parts, DamageAmount{Type: part.damageType, Rolled: part.amount, Taken: taken})
		total += taken
	}

	wasStanding := target.Standing()
	target.HP = max(target.HP-total, 0)
	fight.observer.Observe(DamageDealt{
		Source:   source.actor.ID,
		Target:   target.ID,
		Strike:   source.strike,
		Ability:  source.ability,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		HPAfter:  target.HP,
	})
	if total == 0 {
		return
	}
	switch {
	case wasStanding && target.HP == 0:
		fight.knockDown(target, source.actor, critical)
	case !wasStanding:
		fight.hitWhileDown(target, critical)
	}
}
```

A save's critical failure counts as a critical hit for the dying value, as PF2e words it ("a critical hit or a critical failure on a save").

##### Task 8 — `internal/pf2e/sim/engine/strike.go`

`strike` takes the ability it belongs to, and fight modifiers on attack, damage and AC.

**Find:**
```go
func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike) {
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	total := die + strike.Bonus + penalty
	armorClass := target.Creature.Defense.AC
	degree := DegreeOf(die, total, armorClass)
	fight.observer.Observe(StrikeRolled{
		Attacker: attacker.ID,
		Target:   target.ID,
		Strike:   strike.ID,
		Die:      die,
		Bonus:    strike.Bonus,
		Penalty:  penalty,
		Total:    total,
		AC:       armorClass,
		Degree:   degree,
	})
	if degree < Success {
		return
	}
	fight.dealDamage(attacker, target, strike, degree == CriticalSuccess)
}
```
**Replace with:**
```go
func (fight *Fight) strike(attacker, target *Combatant, strike creature.Strike, ability string) {
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierTotal(attacker, []string{"attack"}, nil)
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
	if degree < Success {
		return
	}
	source := effectSource{actor: attacker, ability: ability, strike: strike.ID, traits: strike.Traits}
	critical := degree == CriticalSuccess
	damageBonus := fight.modifierTotal(attacker, []string{"damage"}, nil)
	fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
}
```

##### Task 9 — `internal/pf2e/sim/engine/run.go`

Initiative takes Perception and initiative modifiers.

**Find:**
```go
		combatant.Initiative = die + combatant.Creature.Perception
```
**Replace with:**
```go
		combatant.Initiative = die + combatant.Creature.Perception + fight.modifierTotal(combatant, []string{"perception", "initiative"}, nil)
```

A Strike choice names no ability; an ability choice runs its steps.

**Find:**
```go
		fight.strike(actor, choice.Target, strike)
		return nil
```
**Replace with:**
```go
		fight.strike(actor, choice.Target, strike, "")
		return nil
```

**Find:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: ability.Actions, Target: choice.Target.ID})
		target := choice.Target
		for _, step := range ability.Steps {
			if !target.Standing() {
				next, found := fight.decider.NextTarget(fight, actor)
				if !found {
					return nil
				}
				target = next
			}
			fight.strike(actor, target, actor.Creature.Strikes[step.Strike])
		}
		return nil
```
**Replace with:**
```go
		fight.observer.Observe(ActionUsed{Combatant: actor.ID, Action: choice.ID, Actions: ability.Actions, Target: choice.Target.ID})
		fight.useAbility(actor, ability, choice.Target)
		return nil
```

##### Task 10 — `internal/pf2e/sim/engine/events.go`

**Find:**
```go
type DamageDealt struct {
	Source   string
	Target   string
	Strike   string
	Critical bool
```
**Replace with:**
```go
type SaveRolled struct {
	Combatant string
	Source    string
	Ability   string
	Save      string
	Die       int
	Bonus     int
	Total     int
	DC        int
	Degree    Degree
}

type SpellAttackRolled struct {
	Attacker string
	Target   string
	Ability  string
	Die      int
	Bonus    int
	Penalty  int
	Total    int
	AC       int
	Degree   Degree
}

type DamageDealt struct {
	Source   string
	Target   string
	Strike   string
	Ability  string
	Critical bool
```

**Find:**
```go
func (StrikeRolled) isEvent()     {}
```
**Replace with:**
```go
func (StrikeRolled) isEvent()      {}
func (SaveRolled) isEvent()        {}
func (SpellAttackRolled) isEvent() {}
```

(`gofmt` realigns the block.)

##### Task 11 — `internal/pf2e/sim/engine/helpers_test.go` and `rules_test.go`

`testFight` sets a decider, and two helpers build abilities.

**Find:**
```go
	fight := &Fight{
		roller:     roller,
		observer:   events,
		combatants: combatants,
		order:      slices.Clone(combatants),
	}
	return fight, events
}
```
**Replace with:**
```go
	fight := &Fight{
		roller:     roller,
		observer:   events,
		decider:    FirstFitting{},
		combatants: combatants,
		order:      slices.Clone(combatants),
	}
	return fight, events
}

func damagePart(t *testing.T, text, damageType string) creature.DamagePart {
	t.Helper()
	expression, err := dice.Parse(text)
	if err != nil {
		t.Fatalf("damagePart(%q): %v", text, err)
	}
	return creature.DamagePart{Dice: text, Type: damageType, Roll: expression}
}

func giveAbility(combatant *Combatant, ability creature.Ability) {
	if combatant.Creature.Abilities == nil {
		combatant.Creature.Abilities = map[string]creature.Ability{}
	}
	combatant.Creature.Abilities[ability.ID] = ability
}
```

In `rules_test.go`, every `fight.strike(attacker, target, fist())` call gains a fourth argument: `fight.strike(attacker, target, fist(), "")`. Edit each call with `replace_all`.

##### Task 12 — tests

`internal/pf2e/sim/creature/steps_test.go` (new, `package creature_test`):

- **`TestLoadReadsSteps`** — a hand layer with `[spellcasting]` (attack 20, DC 30), `class_dc = 25`, and `[ability.spores]` (actions 2, traits `["mental", "occult"]`, area cone 40, one step: save `will`, DC 39, basic, damage `15d6 mental`, and a `failure` list holding one damage step `1d6 mental`) loads; the step's `Kind()` is `save`, its damage's `Roll` is parsed, and `For(1, false)` returns the one nested step with `Kind()` `damage`. A second ability (`spell_attack = true`, damage `2d6 fire`, `success` list empty) loads, and `For(3, true)` on it returns its `success` list.
- **`TestLoadRejectsBadSteps`** — table cases, each `ErrInvalid` with the quoted text: two heads (`strike` and `save`: `want exactly one of`); save with no DC (`a save needs dc, or dc_from class or spell`); `dc` and `dc_from` (`set dc or dc_from, not both`); `dc_from = "class"` with no class DC (`dc_from class needs a class DC`); damage on a non-basic save (`damage on a save needs basic = true`); `basic` with no damage (`basic needs damage`); a `failure` list on a damage step (`failure needs a strike, save or spell_attack`); an unknown strike inside a `success` list (`steps[0].success[0]: unknown strike "kick"`); area shape `square` (`shape "square"`); `target = "ally"` (`target "ally"`); `shift` on a strike (`belong on a save`).

`internal/pf2e/sim/engine/steps_test.go` (new, `package engine`):

- **`TestBasicSaveScalesDamage`** — a caster's ability with one basic Will save step (DC 20, damage `2d6 mental`) against a target with Will +10; scripted damage dice 3 and 4 (7), then the save die. Die 20 (total 30, critical success): no `DamageDealt`. Die 15 (25, success): 3 damage. Die 5 (15, failure): 7. Die 1 (11, a failure the natural 1 drops to critical failure): 14, and `Critical` is true.
- **`TestAreaCatchesEveryStandingFoe`** — an ability with `Area` cone 40 and a basic save step against three foes, one already at 0 HP; two `SaveRolled` events fire (the standing foes, in setup order), and the damage dice are rolled once (the scripted roller holds exactly one damage roll and two save dice).
- **`TestSaveDCFromClassAndSpell`** — `dc_from = "class"` uses `ClassDC` and `dc_from = "spell"` uses `Spellcasting.DC`; a `class-dc` modifier of +1 raises the class one; the `SaveRolled` DC shows each.
- **`TestSaveModifierAgainstTraits`** — a target with `[modifier]` +2 status `saving-throw` against `occult`: its `SaveRolled` bonus is Will + 2 against an ability with the `occult` trait and Will + 0 against one without.
- **`TestDegreeShiftUpgradesSave`** — a target with a `fortitude` shift from `success` to `critical-success` turns a rolled success into a critical success; a failure stays a failure.
- **`TestStepShiftWorsensSave`** — a save step with `Shift: -1` turns a rolled success into a failure, and the saver's `success` → `critical-success` shift does not apply to the shifted result (Decision 20a).
- **`TestSpellAttackUsesPenaltyAndOutcomes`** — a caster with `AttacksThisTurn = 1` makes a spell attack (attack +15) at AC 20 with die 12: the event shows penalty −5 and total 22, damage is dealt, and the step's `success` list (a damage step) runs; with die 20 and no `critical_success` list the `success` list runs again and the damage is doubled.
- **`TestAbilityRetargetsWhenTargetFalls`** — a three-Strike ability whose first Strike drops its target moves on to the next standing foe, as before Commit 2.
- **`TestStepsFromALoadedLayer`** — `creature.Load` reads a hand layer written to `t.TempDir()` with a basic-save cone ability; `Run` with `FirstFitting` and a seeded roller plays it against two dummy PCs, and the events hold one `SaveRolled` per PC.

##### Commit message

```
feat(E3): run save, spell attack and damage steps

- steps are recursive: strike, save, spell_attack or damage, each degree of success holding more steps
- abilities gain traits and an area; an area catches every standing foe until Effort 7
- saves: basic damage by degree, DC fixed or from class or spell DC, modifiers and degree shifts against the effect's traits
- damage is rolled once per step, then applied; Strikes take fight modifiers on attack, damage and AC
- new events SaveRolled and SpellAttackRolled; DamageDealt names the ability
```

### Phase 2 — Conditions and the Down row

**In plain words:** abilities can give conditions that lower numbers, last for a set time and wear off, and a PC's death follows doomed, drained, hero points and the dying limit from their sheet. When the phase lands, frightened from a failed save lowers a PC's AC and fades at the end of their turns, and Kas dies at dying 6.

#### Commit 3 — `feat(E3): apply conditions with durations`

##### Task 1 — `internal/pf2e/sim/creature/conditions.go` (new)

Full contents:

```go
package creature

import (
	"maps"
	"slices"
	"strings"
)

const StepCondition = "condition"

var (
	ConditionNames     = []string{"blinded", "clumsy", "concealed", "dazzled", "doomed", "drained", "encumbered", "enfeebled", "frightened", "off-guard", "prone", "sickened", "stupefied"}
	valuedConditions   = []string{"clumsy", "doomed", "drained", "enfeebled", "frightened", "sickened", "stupefied"}
	DurationBoundaries = []string{"source-turn-start", "source-turn-end", "target-turn-start", "target-turn-end"}
)

type Duration struct {
	Rounds int    `toml:"rounds,omitzero"`
	Until  string `toml:"until,omitempty"`
}

type ConditionAdjustment struct {
	Gain  int `toml:"gain,omitzero"`
	Decay int `toml:"decay,omitzero"`
}

func IsCondition(name string) bool {
	return slices.Contains(ConditionNames, name)
}

func IsValuedCondition(name string) bool {
	return slices.Contains(valuedConditions, name)
}

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

func checkDuration(path string, duration *Duration, report func(string, ...any)) {
	if duration == nil {
		return
	}
	switch {
	case duration.Rounds < 0:
		report("%s.duration: rounds must not be negative", path)
	case duration.Rounds > 0 && duration.Until != "":
		report("%s.duration: set rounds or until, not both", path)
	case duration.Rounds == 0 && !slices.Contains(DurationBoundaries, duration.Until):
		report("%s.duration: until %q: want %s", path, duration.Until, strings.Join(DurationBoundaries, ", "))
	}
}

func (creature *Creature) checkConditionAdjustments(report func(string, ...any)) {
	for _, name := range slices.Sorted(maps.Keys(creature.ConditionAdjustments)) {
		adjustment := creature.ConditionAdjustments[name]
		if !IsCondition(name) {
			report("condition_adjust.%s: not a condition", name)
		}
		if adjustment.Decay < 0 {
			report("condition_adjust.%s: decay must not be negative", name)
		}
		if adjustment.Gain == 0 && adjustment.Decay == 0 {
			report("condition_adjust.%s: sets neither gain nor decay", name)
		}
	}
}
```

##### Task 2 — `internal/pf2e/sim/creature/steps.go`

A step can give a condition, and any step can skip a target that already has one (Terrifying Touch's "isn't already frightened").

**Find:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage}
```
**Replace with:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepCondition}
```

**Find:**
```go
	Basic       bool         `toml:"basic,omitempty"`
	Shift       int          `toml:"shift,omitzero"`
	Outcomes
}
```
**Replace with:**
```go
	Basic       bool         `toml:"basic,omitempty"`
	Shift       int          `toml:"shift,omitzero"`
	Condition   string       `toml:"condition,omitempty"`
	Value       int          `toml:"value,omitzero"`
	Duration    *Duration    `toml:"duration,omitempty"`
	Unless      string       `toml:"unless,omitempty"`
	Outcomes
}
```

**Find:**
```go
	if len(step.Damage) > 0 && step.Save == "" && !step.SpellAttack {
		heads = append(heads, StepDamage)
	}
	return heads
```
**Replace with:**
```go
	if len(step.Damage) > 0 && step.Save == "" && !step.SpellAttack {
		heads = append(heads, StepDamage)
	}
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	return heads
```

**Find:**
```go
	case StepSpellAttack:
		if creature.Spellcasting == nil {
			report("%s: spell_attack needs [spellcasting]", path)
		}
	}
```
**Replace with:**
```go
	case StepSpellAttack:
		if creature.Spellcasting == nil {
			report("%s: spell_attack needs [spellcasting]", path)
		}
	case StepCondition:
		checkCondition(path, step, report)
	}
	if kind != StepCondition && (step.Value != 0 || step.Duration != nil) {
		report("%s: value and duration belong on a condition", path)
	}
	checkDuration(path, step.Duration, report)
	if step.Unless != "" && !IsCondition(step.Unless) {
		report("%s: unless %q: not a condition", path, step.Unless)
	}
```

##### Task 3 — `internal/pf2e/sim/creature/creature.go`, `fight.go` and `build.go`

`Creature` gains the condition adjustments; `Strike` gains its range, so a condition can tell melee from ranged.

**Find:**
```go
	DegreeShifts map[string]DegreeShift   `toml:"degree_shift,omitempty"`
```
**Replace with:**
```go
	DegreeShifts map[string]DegreeShift   `toml:"degree_shift,omitempty"`
	ConditionAdjustments map[string]ConditionAdjustment `toml:"condition_adjust,omitempty"`
```

**Find:**
```go
type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name,omitempty"`
	Bonus  int          `toml:"bonus,omitzero"`
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
}
```
**Replace with:**
```go
type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name,omitempty"`
	Bonus  int          `toml:"bonus,omitzero"`
	Range  int          `toml:"range,omitzero"`
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
}

func (strike Strike) Melee() bool {
	return strike.Range == 0
}
```

In `checkFightRules` (`fight.go`), check the adjustments last.

**Find:**
```go
		if !slices.Contains(DegreeWords, shift.From) || !slices.Contains(DegreeWords, shift.To) {
			report("degree_shift.%s: from %q and to %q must each be one of %s", id, shift.From, shift.To, strings.Join(DegreeWords, ", "))
		}
	}
}
```
**Replace with:**
```go
		if !slices.Contains(DegreeWords, shift.From) || !slices.Contains(DegreeWords, shift.To) {
			report("degree_shift.%s: from %q and to %q must each be one of %s", id, shift.From, shift.To, strings.Join(DegreeWords, ", "))
		}
	}
	creature.checkConditionAdjustments(report)
}
```

In `Build.strike` (`build.go`), keep the weapon's range.

**Find:**
```go
	return Strike{
		Name:   weapon.Name,
		Bonus:  attack,
		Traits: slices.Clone(weapon.Traits),
```
**Replace with:**
```go
	return Strike{
		Name:   weapon.Name,
		Bonus:  attack,
		Range:  weapon.Range,
		Traits: slices.Clone(weapon.Traits),
```

In `prepare` (`validate.go`), a Strike's range is not negative.

**Find:**
```go
		if strike.Name == "" {
			report("strike.%s: name is empty", id)
		}
```
**Replace with:**
```go
		if strike.Name == "" {
			report("strike.%s: name is empty", id)
		}
		if strike.Range < 0 {
			report("strike.%s: range must not be negative", id)
		}
```

##### Task 4 — `internal/pf2e/sim/engine/conditions.go` (new)

Full contents:

```go
package engine

import (
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var allChecksAndDCs = []string{"ac", "fortitude", "reflex", "will", "saving-throw", "perception", "initiative", "attack", "spell-attack", "spell-dc", "class-dc"}

type conditionRule struct {
	selectors []string
	kind      string
	fixed     int
	decays    bool
	implies   map[string]int
}

// deliberate: each condition is a row of shared parts, not a handler, because the ruleset as data (D-2) will lift this table into files (Decision 2).
var conditionRules = map[string]conditionRule{
	"blinded":    {implies: map[string]int{"off-guard": 0}},
	"clumsy":     {selectors: []string{"ac", "reflex", "ranged-attack"}, kind: "status"},
	"concealed":  {},
	"dazzled":    {},
	"doomed":     {},
	"drained":    {selectors: []string{"fortitude"}, kind: "status"},
	"encumbered": {implies: map[string]int{"clumsy": 1}},
	"enfeebled":  {selectors: []string{"melee-attack", "melee-damage"}, kind: "status"},
	"frightened": {selectors: allChecksAndDCs, kind: "status", decays: true},
	"off-guard":  {selectors: []string{"ac"}, kind: "circumstance", fixed: 2},
	"prone":      {selectors: []string{"attack", "spell-attack"}, kind: "circumstance", fixed: 2, implies: map[string]int{"off-guard": 0}},
	"sickened":   {selectors: allChecksAndDCs, kind: "status"},
	"stupefied":  {selectors: []string{"will", "perception", "spell-attack", "spell-dc"}, kind: "status"},
}

type timing struct {
	source *Combatant
	rounds int
	until  string
	armed  bool
}

type conditionInstance struct {
	name  string
	value int
	clock timing
}

func (fight *Fight) newTiming(source, bearer *Combatant, duration *creature.Duration) timing {
	if duration == nil {
		return timing{source: source}
	}
	clock := timing{source: source, rounds: duration.Rounds, until: duration.Until}
	// deliberate: an end-of-turn boundary applied during the owner's own turn skips that turn's end, because "until the end of your next turn" means the next one (Decision 9).
	clock.armed = fight.current != clock.owner(bearer)
	return clock
}

func (clock timing) owner(bearer *Combatant) *Combatant {
	if strings.HasPrefix(clock.until, "target-") {
		return bearer
	}
	return clock.source
}

func (clock *timing) expires(actor, bearer *Combatant, boundary string) bool {
	if clock.rounds > 0 {
		if boundary != "start" || clock.source != actor {
			return false
		}
		clock.rounds--
		return clock.rounds == 0
	}
	if clock.until == "" || clock.owner(bearer) != actor || !strings.HasSuffix(clock.until, "-"+boundary) {
		return false
	}
	if boundary == "end" && !clock.armed {
		clock.armed = true
		return false
	}
	return true
}

func (combatant *Combatant) Condition(name string) int {
	value := 0
	for _, instance := range combatant.conditions {
		if instance.name == name {
			value = max(value, instance.value, 1)
		}
		if implied, found := conditionRules[instance.name].implies[name]; found {
			value = max(value, implied, 1)
		}
	}
	return value
}

func (fight *Fight) conditionPenalties(combatant *Combatant, selectors []string, pool *creature.BonusPool) {
	for _, name := range creature.ConditionNames {
		rule := conditionRules[name]
		value := combatant.Condition(name)
		if value == 0 || !overlaps(rule.selectors, selectors) {
			continue
		}
		if rule.fixed > 0 {
			value = rule.fixed
		}
		pool.Add(rule.kind, -value)
	}
}

func overlaps(first, second []string) bool {
	return slices.ContainsFunc(first, func(entry string) bool { return slices.Contains(second, entry) })
}

func (fight *Fight) applyCondition(source, target *Combatant, step creature.Step) {
	if target.Dead || slices.Contains(target.Creature.Defense.Immune, step.Condition) {
		return
	}
	value := step.Value
	if creature.IsValuedCondition(step.Condition) {
		value += target.Creature.ConditionAdjustments[step.Condition].Gain
		if value < 1 {
			return
		}
	}
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
}

func (fight *Fight) changeConditions(bearer *Combatant, change func()) {
	before := make(map[string]int, len(creature.ConditionNames))
	for _, name := range creature.ConditionNames {
		before[name] = bearer.Condition(name)
	}
	change()
	for _, name := range creature.ConditionNames {
		if after := bearer.Condition(name); after != before[name] {
			fight.observer.Observe(ConditionChanged{Combatant: bearer.ID, Condition: name, From: before[name], To: after})
		}
	}
}

func (fight *Fight) startTurn(actor *Combatant) {
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "start")
	}
}

func (fight *Fight) endTurn(actor *Combatant) {
	fight.decayConditions(actor)
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "end")
	}
}

func (fight *Fight) expire(bearer, actor *Combatant, boundary string) {
	if len(bearer.conditions) == 0 {
		return
	}
	fight.changeConditions(bearer, func() {
		var kept []*conditionInstance
		for _, instance := range bearer.conditions {
			if !instance.clock.expires(actor, bearer, boundary) {
				kept = append(kept, instance)
			}
		}
		bearer.conditions = kept
	})
}

func (fight *Fight) decayConditions(actor *Combatant) {
	if len(actor.conditions) == 0 {
		return
	}
	fight.changeConditions(actor, func() {
		var kept []*conditionInstance
		for _, instance := range actor.conditions {
			if conditionRules[instance.name].decays {
				decay := actor.Creature.ConditionAdjustments[instance.name].Decay
				if decay == 0 {
					decay = 1
				}
				instance.value -= decay
				if instance.value < 1 {
					continue
				}
			}
			kept = append(kept, instance)
		}
		actor.conditions = kept
	})
}

func immuneToEffect(target *Combatant, traits []string) bool {
	return slices.ContainsFunc(target.Creature.Defense.Immune, func(key string) bool {
		return slices.Contains(traits, strings.TrimSuffix(key, "-effects"))
	})
}

func (fight *Fight) affected(source effectSource, targets []*Combatant) []*Combatant {
	var affected []*Combatant
	for _, target := range targets {
		if immuneToEffect(target, source.traits) {
			fight.observer.Observe(ImmuneToEffect{Combatant: target.ID, Source: source.actor.ID, Ability: source.ability})
			continue
		}
		affected = append(affected, target)
	}
	return affected
}

func (fight *Fight) passesConcealment(attacker, target *Combatant) bool {
	dc := 0
	switch {
	case attacker.Condition("blinded") > 0:
		dc = 11
	case target.Condition("concealed") > 0 || attacker.Condition("dazzled") > 0:
		dc = 5
	default:
		return true
	}
	die := fight.roller.Die(20)
	fight.observer.Observe(FlatChecked{Combatant: attacker.ID, Reason: "concealment", Die: die, DC: dc})
	return die >= dc
}
```

##### Task 5 — `internal/pf2e/sim/engine/checks.go`

Conditions join the bonus pool for every check and DC.

**Find:**
```go
	for _, modifier := range combatant.Creature.Modifiers {
		if modifier.Applies(selectors, traits) {
			pool.Add(modifier.Type, modifier.Value)
		}
	}
	return pool
```
**Replace with:**
```go
	for _, modifier := range combatant.Creature.Modifiers {
		if modifier.Applies(selectors, traits) {
			pool.Add(modifier.Type, modifier.Value)
		}
	}
	fight.conditionPenalties(combatant, selectors, pool)
	return pool
```

##### Task 6 — `internal/pf2e/sim/engine/steps.go` and `strike.go`

`useAbility` makes the concealment check once for a targeted ability that opens with a save; `runStep` drops targets an `unless` names or an immunity covers, and runs condition steps.

**Find:**
```go
	for _, step := range ability.Steps {
		if step.Target != "self" && !target.Standing() {
```
**Replace with:**
```go
	if first := ability.Steps[0]; first.Kind() == creature.StepSave && first.Target != "self" && !fight.passesConcealment(actor, target) {
		return
	}
	for _, step := range ability.Steps {
		if step.Target != "self" && !target.Standing() {
```

**Find:**
```go
	if step.Target == "self" {
		targets = []*Combatant{source.actor}
	}
	switch step.Kind() {
```
**Replace with:**
```go
	if step.Target == "self" {
		targets = []*Combatant{source.actor}
	}
	if step.Unless != "" {
		targets = slices.DeleteFunc(slices.Clone(targets), func(target *Combatant) bool { return target.Condition(step.Unless) > 0 })
	}
	if source.strike == "" && step.Kind() != creature.StepStrike {
		targets = fight.affected(source, targets)
	}
	switch step.Kind() {
	case creature.StepCondition:
		for _, target := range targets {
			fight.applyCondition(source.actor, target, step)
		}
```

Add `"slices"` to the imports of `steps.go`.

In `spellAttack`, a concealed target can be missed before the roll.

**Find:**
```go
	penalty := attackPenalty(attacker.AttacksThisTurn, false)
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
```
**Replace with:**
```go
	penalty := attackPenalty(attacker.AttacksThisTurn, false)
	attacker.AttacksThisTurn++
	if !fight.passesConcealment(attacker, target) {
		return
	}
	die := fight.roller.Die(20)
	bonus := attacker.Creature.Spellcasting.Attack + fight.modifierTotal(attacker, []string{"spell-attack"}, nil)
```

In `strike`, the same check, and the Strike's reach picks its selectors.

**Find:**
```go
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierTotal(attacker, []string{"attack"}, nil)
```
**Replace with:**
```go
	penalty := attackPenalty(attacker.AttacksThisTurn, strike.Agile())
	attacker.AttacksThisTurn++
	if !fight.passesConcealment(attacker, target) {
		return
	}
	attackSelectors, damageSelectors := []string{"attack", "ranged-attack"}, []string{"damage"}
	if strike.Melee() {
		attackSelectors, damageSelectors = []string{"attack", "melee-attack"}, []string{"damage", "melee-damage"}
	}
	die := fight.roller.Die(20)
	bonus := strike.Bonus + fight.modifierTotal(attacker, attackSelectors, nil)
```

**Find:**
```go
	damageBonus := fight.modifierTotal(attacker, []string{"damage"}, nil)
```
**Replace with:**
```go
	damageBonus := fight.modifierTotal(attacker, damageSelectors, nil)
```

##### Task 7 — `internal/pf2e/sim/engine/run.go`, `fight.go` and `combatant.go`

The fight knows whose turn it is, and every turn opens and closes with its boundaries, dead combatants included, so durations keep counting.

**Find:**
```go
func (fight *Fight) takeTurn(actor *Combatant) error {
	if actor.Dead {
		return nil
	}
```
**Replace with:**
```go
func (fight *Fight) takeTurn(actor *Combatant) error {
	fight.current = actor
	fight.startTurn(actor)
	err := fight.act(actor)
	if err == nil {
		fight.endTurn(actor)
	}
	fight.current = nil
	return err
}

func (fight *Fight) act(actor *Combatant) error {
	if actor.Dead {
		return nil
	}
```

In `Fight` (`fight.go`):

**Find:**
```go
	order      []*Combatant
	round      int
}
```
**Replace with:**
```go
	order      []*Combatant
	round      int
	current    *Combatant
}
```

In `Combatant` (`combatant.go`):

**Find:**
```go
	AttacksThisTurn int
	ReactionUsed    bool
}
```
**Replace with:**
```go
	AttacksThisTurn int
	ReactionUsed    bool
	conditions      []*conditionInstance
}
```

##### Task 8 — `internal/pf2e/sim/engine/events.go` and `cmd/gm-toolkit/pf2e/show.go`

**Find:**
```go
type Died struct {
	Combatant string
}
```
**Replace with:**
```go
type Died struct {
	Combatant string
}

type ConditionChanged struct {
	Combatant string
	Condition string
	From      int
	To        int
}

type ImmuneToEffect struct {
	Combatant string
	Source    string
	Ability   string
}

type FlatChecked struct {
	Combatant string
	Reason    string
	Die       int
	DC        int
}
```

**Find:**
```go
func (Died) isEvent()             {}
```
**Replace with:**
```go
func (Died) isEvent()             {}
func (ConditionChanged) isEvent() {}
func (ImmuneToEffect) isEvent()   {}
func (FlatChecked) isEvent()      {}
```

In `show.go`, print a Strike's range and the condition adjustments.

**Find:**
```go
			if len(strike.Traits) > 0 {
				fmt.Fprintf(out, " (%s)", strings.Join(strike.Traits, ", "))
			}
```
**Replace with:**
```go
			if len(strike.Traits) > 0 {
				fmt.Fprintf(out, " (%s)", strings.Join(strike.Traits, ", "))
			}
			if strike.Range > 0 {
				fmt.Fprintf(out, " range %d ft", strike.Range)
			}
```

**Find:**
```go
			line := fmt.Sprintf("  %s  %s %s -> %s", id, shift.Selector, shift.From, shift.To)
			if len(shift.Against) > 0 {
				line += " against " + strings.Join(shift.Against, ", ")
			}
			fmt.Fprintln(out, line)
		}
	}
}
```
**Replace with:**
```go
			line := fmt.Sprintf("  %s  %s %s -> %s", id, shift.Selector, shift.From, shift.To)
			if len(shift.Against) > 0 {
				line += " against " + strings.Join(shift.Against, ", ")
			}
			fmt.Fprintln(out, line)
		}
	}
	if len(shown.ConditionAdjustments) > 0 {
		fmt.Fprintf(out, "%sCondition adjustments%s\n", bold, reset)
		for _, name := range slices.Sorted(maps.Keys(shown.ConditionAdjustments)) {
			adjustment := shown.ConditionAdjustments[name]
			fmt.Fprintf(out, "  %s  gain %+d  decay %d\n", name, adjustment.Gain, adjustment.Decay)
		}
	}
}
```

##### Task 9 — tests

`internal/pf2e/sim/creature/steps_test.go`:

- **`TestLoadReadsConditionSteps`** — a save step whose `failure` list gives `frightened` value 1 and whose `critical_failure` list gives `off-guard` with duration `until = "source-turn-end"` loads; a top-level `[condition_adjust.frightened]` with `gain = -1` loads; a PC weapon with `range = 20` gives a Strike whose `Melee()` is false.
- **`TestLoadRejectsBadConditions`** — `ErrInvalid` with the quoted text: unknown condition (`unknown condition "sad"`); frightened with no value (`frightened needs a value of at least 1`); off-guard with a value (`off-guard takes no value`); rounds and until (`set rounds or until, not both`); bad boundary (`until "tomorrow"`); a value on a save (`value and duration belong on a condition`); `unless = "sad"` (`unless "sad": not a condition`); `[condition_adjust.sad]` (`not a condition`); `[condition_adjust.frightened]` with neither key (`sets neither gain nor decay`).

`internal/pf2e/sim/engine/conditions_test.go` (new):

- **`TestConditionPenalties`** — table over one combatant with one condition: frightened 2 lowers AC, Will, a melee attack and the spell DC by 2; clumsy 1 lowers AC, Reflex and a ranged attack but not a melee attack; enfeebled 2 lowers a melee attack and melee damage but not a ranged attack; stupefied 1 lowers Will and the spell attack but not Fortitude; drained 1 lowers Fortitude; off-guard lowers AC by 2; prone lowers attacks and spell attacks by 2 and AC by 2; encumbered lowers AC by 1 (clumsy 1). Checked through `armorClass` and `modifierTotal`.
- **`TestConditionsStackByType`** — frightened 1 and sickened 2 lower AC by 2 (status, highest only); off-guard and frightened 1 lower it by 3 (two types).
- **`TestFrightenedDecays`** — frightened 2 is 1 after the bearer's `endTurn` and gone after the next; with `[condition_adjust.frightened] decay = 2` it is gone after one; a `ConditionChanged` event marks each step.
- **`TestConditionGain`** — with `gain = -1`, a frightened 1 step does nothing and a frightened 2 step gives frightened 1.
- **`TestDurationInRounds`** — clumsy with `rounds = 1`, applied while the source's turn is current, survives the target's turn and ends at the source's next `startTurn`.
- **`TestDurationUntilSourceTurnEnd`** — applied during the source's turn, it survives that `endTurn` and ends at the next one; applied during another creature's turn, it ends at the source's next `endTurn`.
- **`TestDurationUntilTargetTurnStart`** — ends at the target's next `startTurn`, not the source's.
- **`TestConditionImmunity`** — a target whose `immune` holds `frightened` gains nothing and no event fires.
- **`TestEffectTraitImmunity`** — a target immune to `death-effects` facing an ability with the `death` trait gets an `ImmuneToEffect` event and no save is rolled.
- **`TestUnlessSkipsTarget`** — a step with `unless = "frightened"` does not run against a frightened target.
- **`TestConcealmentFlatCheck`** — a Strike at a concealed target: flat die 4 stops it before the attack roll (no `StrikeRolled`, and `AttacksThisTurn` still rises); die 5 lets it roll. A dazzled attacker checks at DC 5; a blinded one at DC 11.
- **`TestConditionStepsFromALoadedLayer`** — `creature.Load` reads a hand layer with a basic-save cone whose `failure` list gives frightened 1; `Run` with a seeded roller produces at least one `ConditionChanged` event, and replaying the seed gives the same events.

##### Commit message

```
feat(E3): apply conditions with durations

- condition steps, durations in rounds or until a turn boundary, and end-of-turn decay
- a Go table of conditions built from shared parts: the numbers each lowers, the bonus type, fixed or valued, what it brings with it
- conditions join the bonus pool for every check and DC; Strikes split melee and ranged selectors by range
- condition immunity, immunity to effect traits (death-effects), unless gates and concealment flat checks
- condition adjustments for gains and decay (Bravery, Calm and Centered)
- every turn opens and closes with its boundaries; new events ConditionChanged, ImmuneToEffect and FlatChecked
```

#### Commit 4 — `feat(E3): add doomed, drained, hero points and dying limits`

##### Task 1 — `internal/pf2e/sim/creature/creature.go` and `validate.go`

**Find:**
```go
	ClassDC      int                `toml:"class_dc,omitzero"`
```
**Replace with:**
```go
	ClassDC      int                `toml:"class_dc,omitzero"`
	HeroPoints   int                `toml:"hero_points,omitzero"`
	DyingMax     int                `toml:"dying_max,omitzero"`
	RecoveryDC   int                `toml:"recovery_dc,omitzero"`
```

Damage with no type (Disintegrate) is `untyped`.

**Find:**
```go
	damageTypes   = []string{"acid", "bleed", "bludgeoning", "cold", "electricity", "fire", "force", "mental", "piercing", "poison", "slashing", "sonic", "spirit", "vitality", "void"}
```
**Replace with:**
```go
	damageTypes   = []string{"acid", "bleed", "bludgeoning", "cold", "electricity", "fire", "force", "mental", "piercing", "poison", "slashing", "sonic", "spirit", "untyped", "vitality", "void"}
```

In `prepare` (`validate.go`):

**Find:**
```go
	if creature.Defense.AC < 1 {
		report("defense.ac must be at least 1")
	}
```
**Replace with:**
```go
	if creature.Defense.AC < 1 {
		report("defense.ac must be at least 1")
	}
	if creature.HeroPoints < 0 || creature.RecoveryDC < 0 {
		report("hero_points and recovery_dc must not be negative")
	}
	if creature.DyingMax != 0 && creature.DyingMax < 2 {
		report("dying_max %d: want at least 2, or leave it unset for 4", creature.DyingMax)
	}
```

##### Task 2 — `internal/pf2e/sim/creature/steps.go`

A step's damage can kill outright at 0 HP.

**Find:**
```go
	Unless      string       `toml:"unless,omitempty"`
	Outcomes
}
```
**Replace with:**
```go
	Unless      string       `toml:"unless,omitempty"`
	Lethal      bool         `toml:"lethal,omitempty"`
	Outcomes
}
```

**Find:**
```go
	if step.Unless != "" && !IsCondition(step.Unless) {
		report("%s: unless %q: not a condition", path, step.Unless)
	}
```
**Replace with:**
```go
	if step.Unless != "" && !IsCondition(step.Unless) {
		report("%s: unless %q: not a condition", path, step.Unless)
	}
	if step.Lethal && len(step.Damage) == 0 {
		report("%s: lethal needs damage", path)
	}
```

##### Task 3 — `internal/pf2e/sim/engine/dying.go`

The death value is the dying maximum minus doomed; hero points are spent when a PC would die (Decision 20b); recovery checks use the sheet's base DC.

**Find:**
```go
const deathThreshold = 4
```
**Replace with:**
```go
const (
	defaultDyingMax   = 4
	defaultRecoveryDC = 10
)

func (combatant *Combatant) deathValue() int {
	dyingMax := combatant.Creature.DyingMax
	if dyingMax == 0 {
		dyingMax = defaultDyingMax
	}
	return dyingMax - combatant.Condition("doomed")
}

func (combatant *Combatant) MaxHP() int {
	return combatant.Creature.Defense.HP - combatant.Condition("drained")*max(combatant.Creature.Level, 1)
}
```

**Find:**
```go
func (fight *Fight) setDying(target *Combatant, value int) {
	from := target.Dying
	target.Dying = min(value, deathThreshold)
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: target.Dying})
	if target.Dying >= deathThreshold {
		fight.kill(target)
	}
}
```
**Replace with:**
```go
func (fight *Fight) setDying(target *Combatant, value int) {
	limit := target.deathValue()
	// deliberate: hero points are spent when the PC would die, not at the first rise in dying, because the script plays it that way and Robert kept it (Decision 20b).
	if value >= limit && target.HeroPoints > 0 {
		fight.heroicRecovery(target)
		return
	}
	from := target.Dying
	target.Dying = min(value, max(limit, 1))
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: target.Dying})
	if value >= limit {
		fight.kill(target)
	}
}

func (fight *Fight) heroicRecovery(target *Combatant) {
	from := target.Dying
	target.HeroPoints = 0
	target.Dying = 0
	fight.observer.Observe(HeroicRecovery{Combatant: target.ID})
	fight.observer.Observe(DyingChanged{Combatant: target.ID, From: from, To: 0})
}

func (fight *Fight) drain(source, target *Combatant, before int) {
	gained := target.Condition("drained") - before
	if gained <= 0 || target.Dead {
		return
	}
	loss := gained * max(target.Creature.Level, 1)
	wasStanding := target.Standing()
	target.HP = max(target.HP-loss, 0)
	fight.observer.Observe(HPLost{Combatant: target.ID, Reason: "drained", Amount: loss, HPAfter: target.HP})
	if wasStanding && target.HP == 0 {
		fight.knockDown(target, source, false)
	}
}
```

**Find:**
```go
	dc := 10 + target.Dying
```
**Replace with:**
```go
	base := target.Creature.RecoveryDC
	if base == 0 {
		base = defaultRecoveryDC
	}
	dc := base + target.Dying
```

##### Task 4 — `internal/pf2e/sim/engine/conditions.go`, `steps.go`, `damage.go` and `combatant.go`

Gaining drained loses HP.

**Find:**
```go
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
}
```
**Replace with:**
```go
	instance := &conditionInstance{name: step.Condition, value: value, clock: fight.newTiming(source, target, step.Duration)}
	drainedBefore := target.Condition("drained")
	fight.changeConditions(target, func() { target.conditions = append(target.conditions, instance) })
	if step.Condition == "drained" {
		fight.drain(source, target, drainedBefore)
	}
}
```

The source carries the step's `lethal` flag.

**Find:**
```go
type effectSource struct {
	actor   *Combatant
	ability string
	strike  string
	traits  []string
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
}
```

**Find:**
```go
func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	if step.Target == "self" {
```
**Replace with:**
```go
func (fight *Fight) runStep(source effectSource, step creature.Step, targets []*Combatant) {
	source.lethal = step.Lethal
	if step.Target == "self" {
```

In `applyDamage` (`damage.go`):

**Find:**
```go
	switch {
	case wasStanding && target.HP == 0:
		fight.knockDown(target, source.actor, critical)
```
**Replace with:**
```go
	switch {
	case wasStanding && target.HP == 0 && source.lethal:
		fight.kill(target)
	case wasStanding && target.HP == 0:
		fight.knockDown(target, source.actor, critical)
```

In `combatant.go`, a combatant holds its hero points:

**Find:**
```go
	ReactionUsed    bool
	conditions      []*conditionInstance
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	return &Combatant{ID: source.ID, Side: side, Creature: source, HP: source.Defense.HP}
}
```
**Replace with:**
```go
	ReactionUsed    bool
	HeroPoints      int
	conditions      []*conditionInstance
}

func NewCombatant(side Side, source *creature.Creature) *Combatant {
	return &Combatant{ID: source.ID, Side: side, Creature: source, HP: source.Defense.HP, HeroPoints: source.HeroPoints}
}
```

##### Task 5 — `internal/pf2e/sim/engine/events.go` and `cmd/gm-toolkit/pf2e/show.go`

**Find:**
```go
type ConditionChanged struct {
```
**Replace with:**
```go
type HeroicRecovery struct {
	Combatant string
}

type HPLost struct {
	Combatant string
	Reason    string
	Amount    int
	HPAfter   int
}

type ConditionChanged struct {
```

**Find:**
```go
func (ConditionChanged) isEvent() {}
```
**Replace with:**
```go
func (ConditionChanged) isEvent() {}
func (HeroicRecovery) isEvent()   {}
func (HPLost) isEvent()           {}
```

In `printCreature`, after the class DC line:

**Find:**
```go
	if shown.ClassDC > 0 {
		fmt.Fprintf(out, "Class DC %d\n", shown.ClassDC)
	}
```
**Replace with:**
```go
	if shown.ClassDC > 0 {
		fmt.Fprintf(out, "Class DC %d\n", shown.ClassDC)
	}
	if shown.HeroPoints > 0 || shown.DyingMax > 0 || shown.RecoveryDC > 0 {
		fmt.Fprintf(out, "Hero points %d  dying max %d  recovery DC %d\n", shown.HeroPoints, cmp.Or(shown.DyingMax, 4), cmp.Or(shown.RecoveryDC, 10))
	}
```

Add `"cmp"` to the imports of `show.go`.

##### Task 6 — tests

`internal/pf2e/sim/creature/fight_test.go`: **`TestLoadReadsDownFields`** — `hero_points = 1`, `dying_max = 6`, `recovery_dc = 6` load; `dying_max = 1` fails with `dying_max 1`; `hero_points = -1` fails; a damage step with `lethal = true` and type `untyped` loads, and `lethal` on a condition step fails with `lethal needs damage`.

`internal/pf2e/sim/engine/down_test.go` (new):

- **`TestDoomedLowersTheDeathValue`** — a PC with doomed 1 at dying 2 takes damage while down and dies at dying 3.
- **`TestDyingMax`** — with `DyingMax: 6`, a PC at dying 4 hit again (dying 5) stays alive; dying 6 kills.
- **`TestRecoveryDC`** — with `RecoveryDC: 6`, a PC at dying 1 checks against DC 7 (die 7 is a success).
- **`TestHeroicRecovery`** — a PC with 1 hero point at dying 3 that would reach dying 4 gets `HeroicRecovery`, dying 0, 0 HP, its wounded value unchanged and 0 hero points; knocked down and pushed to dying 4 again, it dies.
- **`TestDrainedLosesHP`** — a level-5 PC at 40 HP gains drained 2: an `HPLost` of 10, HP 30, `MaxHP()` lowered by 10; raising it to drained 3 loses 5 more.
- **`TestLethalDamageKills`** — a lethal damage step that takes a PC to 0 HP fires `Died`, not `DyingChanged`.
- **`TestUntypedDamageTakesResistAll`** — 20 untyped damage against `resist all 5` deals 15.

##### Commit message

```
feat(E3): add doomed, drained, hero points and dying limits

- a PC dies at its dying maximum (dying_max, 4 by default) minus doomed
- hero points: Heroic Recovery when a PC would die, leaving it stable at 0 HP with wounded unchanged
- recovery checks use recovery_dc (10 by default) plus the dying value
- drained loses level x value HP when gained and lowers maximum HP
- lethal damage kills outright at 0 HP; untyped damage for Disintegrate
```

### Phase 3 — Damage kinds

**In plain words:** damage gains its special forms: persistent damage that burns each turn until a flat check ends it, splash, healing, temporary HP and the minimum of 1; defenses learn traits (the Avatar's holy weakness, cold iron through her resistance), Protean Anatomy's shifting resistance and effects that lower a defense; and a Strike can carry riders on a hit or a critical hit. When the phase lands, Tavish's bombs burn and splash, and Kas's hammer knocks a foe prone on a critical hit.

#### Commit 5 — `feat(E3): add persistent and splash damage, healing and temp HP`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`, `steps.go` and `validate.go`

A damage part has a category.

**Find:**
```go
type DamagePart struct {
	Dice string          `toml:"dice"`
	Type string          `toml:"type"`
	Roll dice.Expression `toml:"-"`
}
```
**Replace with:**
```go
type DamagePart struct {
	Dice     string          `toml:"dice"`
	Type     string          `toml:"type"`
	Category string          `toml:"category,omitempty"`
	Roll     dice.Expression `toml:"-"`
}

var damageCategories = []string{"", "persistent", "splash"}
```

In `prepareDamage` (`steps.go`):

**Find:**
```go
		if !IsDamageType(part.Type) {
			report("%s.damage[%d]: unknown damage type %q", path, index, part.Type)
		}
	}
}
```
**Replace with:**
```go
		if !IsDamageType(part.Type) {
			report("%s.damage[%d]: unknown damage type %q", path, index, part.Type)
		}
		if !slices.Contains(damageCategories, part.Category) {
			report("%s.damage[%d]: category %q: want persistent, splash or none", path, index, part.Category)
		}
	}
}
```

Heal and temporary HP steps:

**Find:**
```go
const (
	StepStrike      = "strike"
	StepSave        = "save"
	StepSpellAttack = "spell_attack"
	StepDamage      = "damage"
)
```
**Replace with:**
```go
const (
	StepStrike      = "strike"
	StepSave        = "save"
	StepSpellAttack = "spell_attack"
	StepDamage      = "damage"
	StepHeal        = "heal"
	StepTempHP      = "temp_hp"
)
```

**Find:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepCondition}
```
**Replace with:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition}
```

**Find:**
```go
	Lethal      bool         `toml:"lethal,omitempty"`
	Outcomes
}
```
**Replace with:**
```go
	Lethal      bool            `toml:"lethal,omitempty"`
	Heal        string          `toml:"heal,omitempty"`
	TempHP      string          `toml:"temp_hp,omitempty"`
	HealRoll    dice.Expression `toml:"-"`
	TempHPRoll  dice.Expression `toml:"-"`
	Outcomes
}
```

**Find:**
```go
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	return heads
```
**Replace with:**
```go
	if step.Heal != "" {
		heads = append(heads, StepHeal)
	}
	if step.TempHP != "" {
		heads = append(heads, StepTempHP)
	}
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	return heads
```

**Find:**
```go
	case StepCondition:
		checkCondition(path, step, report)
	}
```
**Replace with:**
```go
	case StepCondition:
		checkCondition(path, step, report)
	case StepHeal:
		step.HealRoll = parseAmount(path+".heal", step.Heal, report)
	case StepTempHP:
		step.TempHPRoll = parseAmount(path+".temp_hp", step.TempHP, report)
	}
```

Add below `prepareDamage`:

```go
func parseAmount(path, text string, report func(string, ...any)) dice.Expression {
	expression, err := dice.Parse(text)
	if err != nil {
		report("%s: %v", path, err)
	}
	return expression
}
```

In `prepare` (`validate.go`), a Strike needs a part with no category:

**Find:**
```go
		prepareDamage("strike."+id, strike.Damage, report)
		creature.Strikes[id] = strike
```
**Replace with:**
```go
		prepareDamage("strike."+id, strike.Damage, report)
		if len(strike.Damage) > 0 && !slices.ContainsFunc(strike.Damage, func(part DamagePart) bool { return part.Category == "" }) {
			report("strike.%s: needs a damage part with no category", id)
		}
		creature.Strikes[id] = strike
```

##### Task 2 — `internal/pf2e/sim/creature/build.go`

A weapon carries persistent damage and splash; `splash` is a modifier selector.

**Find:**
```go
	modifierSelectors    = []string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage"}
	weaponSelectors      = []string{"attack", "damage"}
```
**Replace with:**
```go
	modifierSelectors    = []string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage", "splash"}
	weaponSelectors      = []string{"attack", "damage", "splash"}
```

**Find:**
```go
	Range     int      `toml:"range,omitzero"`
	Traits    []string `toml:"traits,omitempty"`
}
```
**Replace with:**
```go
	Range      int         `toml:"range,omitzero"`
	Traits     []string    `toml:"traits,omitempty"`
	Persistent *DamagePart `toml:"persistent,omitempty"`
	Splash     int         `toml:"splash,omitzero"`
}
```

In `checkBuild`:

**Find:**
```go
		if weapon.Potency < 0 || weapon.Striking < 0 || weapon.ItemBonus < 0 || weapon.Range < 0 {
			report("build.weapon.%s: potency, striking, item_bonus and range must not be negative", id)
		}
```
**Replace with:**
```go
		if weapon.Potency < 0 || weapon.Striking < 0 || weapon.ItemBonus < 0 || weapon.Range < 0 || weapon.Splash < 0 {
			report("build.weapon.%s: potency, striking, item_bonus, range and splash must not be negative", id)
		}
		if persistent := weapon.Persistent; persistent != nil {
			if _, err := dice.Parse(persistent.Dice); err != nil || !IsDamageType(persistent.Type) {
				report("build.weapon.%s.persistent: want parsable dice and a damage type", id)
			}
		}
```

In `Build.strike`:

**Find:**
```go
		Damage: []DamagePart{{Dice: expression.String(), Type: weapon.Type}},
	}
}
```
**Replace with:**
```go
		Damage: build.weaponDamage(level, weapon, expression),
	}
}

func (build *Build) weaponDamage(level int, weapon Weapon, expression dice.Expression) []DamagePart {
	damage := []DamagePart{{Dice: expression.String(), Type: weapon.Type}}
	if persistent := weapon.Persistent; persistent != nil {
		damage = append(damage, DamagePart{Dice: persistent.Dice, Type: persistent.Type, Category: "persistent"})
	}
	if weapon.Splash > 0 {
		splash := weapon.Splash + build.total(level, 0, weapon.Traits, "splash")
		damage = append(damage, DamagePart{Dice: strconv.Itoa(splash), Type: weapon.Type, Category: "splash"})
	}
	return damage
}
```

##### Task 3 — `internal/pf2e/sim/engine/damage.go`

The roll keeps persistent parts aside, doubles neither splash nor persistent at roll time, and deals at least 1. Replace `rollDamage` and `basic`.

**Find:**
```go
type damageRoll struct {
	parts []rolledPart
}

func (fight *Fight) rollDamage(parts []creature.DamagePart, critical bool, bonus int) damageRoll {
	var roll damageRoll
	for index, part := range parts {
		amount := part.Roll.Roll(fight.roller)
		if index == 0 {
			amount += bonus
		}
		amount = max(amount, 0)
		if critical {
			amount *= 2
		}
		roll.parts = addRolled(roll.parts, part.Type, amount)
	}
	return roll
}
```
**Replace with:**
```go
type damageRoll struct {
	parts      []rolledPart
	persistent []creature.DamagePart
	doubled    bool
}

func (fight *Fight) rollDamage(parts []creature.DamagePart, critical bool, bonus int) damageRoll {
	roll := damageRoll{doubled: critical}
	bonusAdded := false
	for _, part := range parts {
		if part.Category == "persistent" {
			roll.persistent = append(roll.persistent, part)
			continue
		}
		amount := part.Roll.Roll(fight.roller)
		if part.Category == "" && !bonusAdded {
			amount += bonus
			bonusAdded = true
		}
		amount = max(amount, 0)
		if critical && part.Category == "" {
			amount *= 2
		}
		roll.parts = addRolled(roll.parts, part.Type, amount)
	}
	// deliberate: a roll that comes to less than 1 deals 1, not 0, because penalties never take damage below 1; a save's halving comes after.
	if len(roll.parts) > 0 && roll.total() < 1 {
		roll.parts[0].amount = 1
	}
	return roll
}

func (roll damageRoll) total() int {
	total := 0
	for _, part := range roll.parts {
		total += part.amount
	}
	return total
}

func splashParts(parts []creature.DamagePart) []creature.DamagePart {
	var splash []creature.DamagePart
	for _, part := range parts {
		if part.Category == "splash" {
			splash = append(splash, part)
		}
	}
	return splash
}
```

**Find:**
```go
	scaled := damageRoll{parts: make([]rolledPart, len(roll.parts))}
```
**Replace with:**
```go
	scaled := damageRoll{parts: make([]rolledPart, len(roll.parts)), persistent: roll.persistent, doubled: roll.doubled}
```

`applyDamage`: temporary HP takes damage first, and persistent parts start after the hit lands.

**Find:**
```go
	wasStanding := target.Standing()
	target.HP = max(target.HP-total, 0)
	fight.observer.Observe(DamageDealt{
		Source:   source.actor.ID,
		Target:   target.ID,
		Strike:   source.strike,
		Ability:  source.ability,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		HPAfter:  target.HP,
	})
	if total == 0 {
		return
	}
	switch {
	case wasStanding && target.HP == 0 && source.lethal:
		fight.kill(target)
	case wasStanding && target.HP == 0:
		fight.knockDown(target, source.actor, critical)
	case !wasStanding:
		fight.hitWhileDown(target, critical)
	}
}
```
**Replace with:**
```go
	absorbed := min(target.TempHP, total)
	target.TempHP -= absorbed
	wasStanding := target.Standing()
	target.HP = max(target.HP-(total-absorbed), 0)
	fight.observer.Observe(DamageDealt{
		Source:   source.actor.ID,
		Target:   target.ID,
		Strike:   source.strike,
		Ability:  source.ability,
		Critical: critical,
		Parts:    parts,
		Total:    total,
		Absorbed: absorbed,
		HPAfter:  target.HP,
	})
	if total > absorbed {
		switch {
		case wasStanding && target.HP == 0 && source.lethal:
			fight.kill(target)
		case wasStanding && target.HP == 0:
			fight.knockDown(target, source.actor, critical)
		case !wasStanding:
			fight.hitWhileDown(target, critical)
		}
	}
	if !target.Dead {
		for _, part := range roll.persistent {
			fight.addPersistent(source, target, part, roll.doubled)
		}
	}
}
```

The E1 finding (Decision Record 4): a weakness is not added to a type that rolled 0.

**Find:**
```go
	if weakness, found := strongestAdjustment(defense.Weakness, damageType, traits); found {
		amount += weakness
	}
```
**Replace with:**
```go
	if weakness, found := strongestAdjustment(defense.Weakness, damageType, traits); found && amount > 0 {
		amount += weakness
	}
```

##### Task 4 — `internal/pf2e/sim/engine/strike.go`

A failed Strike (not a critical failure) still deals its splash.

**Find:**
```go
	if degree < Success {
		return
	}
	source := effectSource{actor: attacker, ability: ability, strike: strike.ID, traits: strike.Traits}
	critical := degree == CriticalSuccess
	damageBonus := fight.modifierTotal(attacker, damageSelectors, nil)
	fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
}
```
**Replace with:**
```go
	source := effectSource{actor: attacker, ability: ability, strike: strike.ID, traits: strike.Traits}
	critical := degree == CriticalSuccess
	switch {
	case degree >= Success:
		damageBonus := fight.modifierTotal(attacker, damageSelectors, nil)
		fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with Effort 7 (Decision 11).
		if splash := splashParts(strike.Damage); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
}
```

##### Task 5 — `internal/pf2e/sim/engine/persistent.go` (new)

Full contents:

```go
package engine

import (
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const persistentRecoveryDC = 15

type persistentInstance struct {
	source  effectSource
	part    creature.DamagePart
	doubled bool
}

func (instance *persistentInstance) twiceAverage() int {
	roll := instance.part.Roll
	value := roll.Count*(roll.Sides+1) + 2*roll.Modifier
	if instance.doubled {
		value *= 2
	}
	return value
}

func (instance *persistentInstance) dice() string {
	if instance.doubled {
		return instance.part.Roll.String() + " x2"
	}
	return instance.part.Roll.String()
}

func (fight *Fight) addPersistent(source effectSource, target *Combatant, part creature.DamagePart, doubled bool) {
	if slices.Contains(target.Creature.Defense.Immune, part.Type) {
		return
	}
	added := &persistentInstance{source: source, part: part, doubled: doubled}
	for index, existing := range target.persistent {
		if existing.part.Type != part.Type {
			continue
		}
		if added.twiceAverage() <= existing.twiceAverage() {
			return
		}
		target.persistent[index] = added
		fight.observer.Observe(PersistentChanged{Combatant: target.ID, Type: part.Type, Dice: added.dice()})
		return
	}
	target.persistent = append(target.persistent, added)
	fight.observer.Observe(PersistentChanged{Combatant: target.ID, Type: part.Type, Dice: added.dice()})
}

func (fight *Fight) takePersistentDamage(actor *Combatant) {
	var kept []*persistentInstance
	for _, instance := range actor.persistent {
		if actor.Dead {
			break
		}
		amount := max(instance.part.Roll.Roll(fight.roller), 0)
		if instance.doubled {
			amount *= 2
		}
		source := instance.source
		source.ability, source.strike, source.lethal = "persistent", "", false
		fight.applyDamage(source, actor, damageRoll{parts: []rolledPart{{damageType: instance.part.Type, amount: amount}}}, false)
		die := fight.roller.Die(20)
		fight.observer.Observe(FlatChecked{Combatant: actor.ID, Reason: "persistent " + instance.part.Type, Die: die, DC: persistentRecoveryDC})
		if die >= persistentRecoveryDC {
			fight.observer.Observe(PersistentEnded{Combatant: actor.ID, Type: instance.part.Type})
			continue
		}
		kept = append(kept, instance)
	}
	if actor.Dead {
		kept = nil
	}
	actor.persistent = kept
}
```

##### Task 6 — `internal/pf2e/sim/engine/healing.go` (new)

Full contents:

```go
package engine

func (fight *Fight) heal(source effectSource, target *Combatant, amount int) {
	if target.Dead {
		return
	}
	wasDown := target.HP == 0
	target.HP = min(target.HP+max(amount, 0), target.MaxHP())
	fight.observer.Observe(Healed{Source: source.actor.ID, Target: target.ID, Ability: source.ability, Amount: amount, HPAfter: target.HP})
	if wasDown && target.HP > 0 && target.Dying > 0 {
		from := target.Dying
		target.Dying = 0
		target.Wounded++
		fight.observer.Observe(Revived{Combatant: target.ID, DyingFrom: from, Wounded: target.Wounded})
	}
}

func (fight *Fight) gainTempHP(target *Combatant, amount int) {
	if target.Dead || amount <= target.TempHP {
		return
	}
	target.TempHP = amount
	fight.observer.Observe(TempHPGained{Combatant: target.ID, Amount: amount})
}
```

Healing an ally is a target the decider picks, which is Effort 4's; until then a heal step reaches an ally only through `target = "self"` on the ally's own ability.

##### Task 7 — `internal/pf2e/sim/engine/steps.go`, `conditions.go` and `combatant.go`

In `runStep`:

**Find:**
```go
	case creature.StepCondition:
		for _, target := range targets {
			fight.applyCondition(source.actor, target, step)
		}
```
**Replace with:**
```go
	case creature.StepCondition:
		for _, target := range targets {
			fight.applyCondition(source.actor, target, step)
		}
	case creature.StepHeal:
		amount := step.HealRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.heal(source, target, amount)
		}
	case creature.StepTempHP:
		amount := step.TempHPRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.gainTempHP(target, amount)
		}
```

In `endTurn` (`conditions.go`), persistent damage comes first:

**Find:**
```go
func (fight *Fight) endTurn(actor *Combatant) {
	fight.decayConditions(actor)
```
**Replace with:**
```go
func (fight *Fight) endTurn(actor *Combatant) {
	fight.takePersistentDamage(actor)
	fight.decayConditions(actor)
```

In `Combatant`:

**Find:**
```go
	HeroPoints      int
	conditions      []*conditionInstance
}
```
**Replace with:**
```go
	HeroPoints      int
	TempHP          int
	conditions      []*conditionInstance
	persistent      []*persistentInstance
}
```

##### Task 8 — `internal/pf2e/sim/engine/events.go` and `cmd/gm-toolkit/pf2e/show.go`

**Find:**
```go
	Parts    []DamageAmount
	Total    int
	HPAfter  int
}
```
**Replace with:**
```go
	Parts    []DamageAmount
	Total    int
	Absorbed int
	HPAfter  int
}

type Healed struct {
	Source  string
	Target  string
	Ability string
	Amount  int
	HPAfter int
}

type Revived struct {
	Combatant string
	DyingFrom int
	Wounded   int
}

type TempHPGained struct {
	Combatant string
	Amount    int
}

type PersistentChanged struct {
	Combatant string
	Type      string
	Dice      string
}

type PersistentEnded struct {
	Combatant string
	Type      string
}
```

**Find:**
```go
func (HPLost) isEvent()           {}
```
**Replace with:**
```go
func (HPLost) isEvent()            {}
func (Healed) isEvent()            {}
func (Revived) isEvent()           {}
func (TempHPGained) isEvent()      {}
func (PersistentChanged) isEvent() {}
func (PersistentEnded) isEvent()   {}
```

In `damageText` (`show.go`), name a part's category:

**Find:**
```go
		texts[index] = part.Roll.String() + " " + part.Type
```
**Replace with:**
```go
		texts[index] = part.Roll.String() + " " + part.Type
		if part.Category != "" {
			texts[index] += " " + part.Category
		}
```

##### Task 9 — tests

`internal/pf2e/sim/creature/steps_test.go`:

- **`TestLoadReadsDamageCategories`** — a PC build weapon with `persistent = { dice = "3", type = "fire" }`, `splash = 3` and a `[build.modifier.splashy]` with selector `splash`, value 1 and `weapon_traits = ["splash"]` (the weapon carries `splash`) gives a Strike whose damage is the weapon part, then `3 fire persistent`, then `4 fire splash`.
- **`TestLoadRejectsBadCategories`** — category `precision` (`category "precision"`); a Strike whose only part is splash (`needs a damage part with no category`); `heal = "xd6"` (`.heal:`); a weapon persistent part with type `sad` (`persistent: want parsable dice`).

`internal/pf2e/sim/engine/damage_test.go` (new):

- **`TestPersistentDamageOnAHit`** — a hit with a `3 fire persistent` part gives `PersistentChanged`; at the target's `endTurn` it takes 3 fire, then a flat check: die 14 keeps it, die 15 ends it with `PersistentEnded`.
- **`TestPersistentKeepsTheLarger`** — `1d6 fire` then `2d6 fire` replaces it; a later `1 fire` does not.
- **`TestPersistentDoublesOnACriticalHit`** — on a critical hit the persistent part is marked doubled, and its rolled 3 deals 6.
- **`TestPersistentImmunity`** — a target immune to `bleed` gains no bleed instance.
- **`TestSplash`** — a bomb Strike (`3d8 fire` plus `3 fire splash`): a hit deals the dice plus 3, combined into one fire amount; a critical hit doubles only the dice; a failure deals 3; a critical failure deals nothing.
- **`TestDamageMinimumOfOne`** — a `1d4-5` Strike that rolls 1 deals 1.
- **`TestWeaknessSkipsAZeroPart`** — against weakness 5 slashing: a lone `1d4-5 slashing` part that rolls 1 is raised to 1 by the minimum and takes the weakness for 6; `1d4-5 slashing` beside `1d6 fire` (rolling 1 and 3) leaves slashing at 0 with no weakness, and fire at 3 (Decision Record 6).
- **`TestTempHPAbsorbsFirst`** — 10 temp HP and a 15-damage hit: `Absorbed` 10, HP down by 5; gaining 5 temp HP over 8 keeps 8.
- **`TestHealingCapsAtMaxHP`** — a heal of 50 on a PC 10 below maximum restores 10.
- **`TestHealingRevivesADyingPC`** — a dying 2 PC healed to 5 HP gets `Revived`, dying 0 and wounded 1.

##### Commit message

```
feat(E3): add persistent and splash damage, healing and temp HP

- damage parts carry a category: persistent parts start on a hit, splash parts add to the hit and land alone on a miss
- persistent damage at the end of the bearer's turn, then a DC 15 flat check; the larger of one type is kept; it doubles on a critical hit
- PC weapons carry persistent damage and splash, raised by splash modifiers
- heal and temp_hp steps; temporary HP absorbs damage first; healing revives a dying PC with 1 more wounded
- every damage roll deals at least 1, and a type that rolled 0 takes no weakness
```

#### Commit 6 — `feat(E3): key defenses on traits and add Strike riders`

##### Task 1 — `internal/pf2e/sim/creature/creature.go` and `validate.go`

Weaknesses and resistances can name a trait or material of the source; Protean Anatomy is an adaptive resistance.

**Find:**
```go
	Immune    []string              `toml:"immune,omitempty"`
	Resist    map[string]Adjustment `toml:"resist,omitempty"`
	Weakness  map[string]Adjustment `toml:"weakness,omitempty"`
}
```
**Replace with:**
```go
	Immune    []string              `toml:"immune,omitempty"`
	Resist    map[string]Adjustment `toml:"resist,omitempty"`
	Weakness  map[string]Adjustment `toml:"weakness,omitempty"`
	Adaptive  *Adaptive             `toml:"adaptive,omitempty"`
}

type Adaptive struct {
	Value int      `toml:"value"`
	Types []string `toml:"types"`
}
```

**Find:**
```go
	physicalTypes = []string{"bludgeoning", "piercing", "slashing"}
)
```
**Replace with:**
```go
	physicalTypes = []string{"bludgeoning", "piercing", "slashing"}
	DefenseTraits = []string{"adamantine", "cold-iron", "holy", "silver", "unholy"}
)

func IsDefenseTrait(name string) bool {
	return slices.Contains(DefenseTraits, name)
}
```

A Strike carries riders by degree:

**Find:**
```go
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
}

func (strike Strike) Melee() bool {
```
**Replace with:**
```go
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
	Outcomes
}

func (strike Strike) Melee() bool {
```

In `prepare` (`validate.go`):

**Find:**
```go
			if key != "all" && key != "physical" && !IsDamageType(key) {
				report("defense.%s.%s: not a damage type, physical or all", table.name, key)
			}
```
**Replace with:**
```go
			if key != "all" && key != "physical" && !IsDamageType(key) && !IsDefenseTrait(key) {
				report("defense.%s.%s: not a damage type, physical, all or one of %s", table.name, key, strings.Join(DefenseTraits, ", "))
			}
```

**Find:**
```go
	creature.checkFightRules(report)
```
**Replace with:**
```go
	if adaptive := creature.Defense.Adaptive; adaptive != nil {
		if adaptive.Value < 1 || len(adaptive.Types) == 0 || !slices.ContainsFunc(adaptive.Types, IsDamageType) {
			report("defense.adaptive: want a value of at least 1 and damage types")
		}
		for _, damageType := range adaptive.Types {
			if !IsDamageType(damageType) {
				report("defense.adaptive: %q is not a damage type", damageType)
			}
		}
	}
	creature.checkFightRules(report)
```

**Find:**
```go
		if strike.Range < 0 {
			report("strike.%s: range must not be negative", id)
		}
```
**Replace with:**
```go
		if strike.Range < 0 {
			report("strike.%s: range must not be negative", id)
		}
		creature.checkOutcomes("strike."+id, strike.Outcomes, true, report)
```

##### Task 2 — `internal/pf2e/sim/creature/steps.go` and `fight.go`

A `weaken` step adds a weakness or lowers resistances for a duration (Inflammation Flask, Tenderizer Grenade).

**Find:**
```go
	StepHeal        = "heal"
	StepTempHP      = "temp_hp"
)
```
**Replace with:**
```go
	StepHeal        = "heal"
	StepTempHP      = "temp_hp"
	StepWeaken      = "weaken"
)

type DefenseChange struct {
	Weakness   int      `toml:"weakness,omitzero"`
	Resistance int      `toml:"resistance,omitzero"`
	Types      []string `toml:"types"`
}
```

**Find:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition}
```
**Replace with:**
```go
	stepKinds   = []string{StepStrike, StepSave, StepSpellAttack, StepDamage, StepHeal, StepTempHP, StepCondition, StepWeaken}
```

**Find:**
```go
	TempHPRoll  dice.Expression `toml:"-"`
	Outcomes
}
```
**Replace with:**
```go
	TempHPRoll  dice.Expression `toml:"-"`
	Weaken      *DefenseChange  `toml:"weaken,omitempty"`
	Outcomes
}
```

**Find:**
```go
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	return heads
```
**Replace with:**
```go
	if step.Condition != "" {
		heads = append(heads, StepCondition)
	}
	if step.Weaken != nil {
		heads = append(heads, StepWeaken)
	}
	return heads
```

**Find:**
```go
	if kind != StepCondition && (step.Value != 0 || step.Duration != nil) {
		report("%s: value and duration belong on a condition", path)
	}
```
**Replace with:**
```go
	if kind != StepCondition && step.Value != 0 {
		report("%s: value belongs on a condition", path)
	}
	if kind != StepCondition && kind != StepWeaken && step.Duration != nil {
		report("%s: duration belongs on a condition or weaken", path)
	}
	if kind == StepWeaken {
		change := step.Weaken
		if change.Weakness < 0 || change.Resistance < 0 || change.Weakness+change.Resistance == 0 {
			report("%s.weaken: want a weakness or a resistance above 0, and neither below", path)
		}
		if len(change.Types) == 0 || !slices.ContainsFunc(change.Types, IsDamageType) {
			report("%s.weaken: want damage types", path)
		}
	}
```

Modifiers keyed on the target's condition (Fearsome Brute). In `fight.go`:

**Find:**
```go
type CheckModifier struct {
	Selector string   `toml:"selector"`
	Type     string   `toml:"type,omitempty"`
	Value    int      `toml:"value"`
	Against  []string `toml:"against,omitempty"`
}
```
**Replace with:**
```go
type CheckModifier struct {
	Selector     string   `toml:"selector"`
	Type         string   `toml:"type,omitempty"`
	Value        int      `toml:"value"`
	Against      []string `toml:"against,omitempty"`
	PerCondition string   `toml:"per_condition,omitempty"`
}

var offenseSelectors = []string{"attack", "melee-attack", "ranged-attack", "damage", "melee-damage"}
```

**Find:**
```go
		if len(modifier.Against) > 0 && !slices.Contains(againstSelectors, modifier.Selector) {
			report("modifier.%s: against needs selector ac or a save", id)
		}
```
**Replace with:**
```go
		if len(modifier.Against) > 0 && !slices.Contains(againstSelectors, modifier.Selector) {
			report("modifier.%s: against needs selector ac or a save", id)
		}
		if modifier.PerCondition != "" && (!IsCondition(modifier.PerCondition) || !slices.Contains(offenseSelectors, modifier.Selector)) {
			report("modifier.%s: per_condition needs a condition and an attack or damage selector", id)
		}
```

##### Task 3 — `internal/pf2e/sim/creature/build.go`

Weapons carry riders, and a weapon at the critical specialization rank adds its group's rider (Decision 15).

**Find:**
```go
	Persistent *DamagePart `toml:"persistent,omitempty"`
	Splash     int         `toml:"splash,omitzero"`
}
```
**Replace with:**
```go
	Persistent *DamagePart `toml:"persistent,omitempty"`
	Splash     int         `toml:"splash,omitzero"`
	Outcomes
}

// deliberate: a group's critical specialization is a rider step in a table, not code per group, because the effect is data; one row exists because only Kas's hammer earns it in the scope.
var criticalSpecializations = map[string]Step{
	"hammer": {Save: "fortitude", DCFrom: "class", Outcomes: Outcomes{
		Failure:         []Step{{Condition: "prone"}},
		CriticalFailure: []Step{{Condition: "prone"}},
	}},
}
```

**Find:**
```go
	WeaponSpecialization string              `toml:"weapon_specialization,omitempty"`
```
**Replace with:**
```go
	WeaponSpecialization string              `toml:"weapon_specialization,omitempty"`
	CriticalSpecialization string            `toml:"critical_specialization,omitempty"`
```

In `checkBuild`:

**Find:**
```go
	if _, known := specializationDamage[build.WeaponSpecialization]; build.WeaponSpecialization != "" && !known {
		report("build.weapon_specialization %q: want normal or greater", build.WeaponSpecialization)
	}
```
**Replace with:**
```go
	if _, known := specializationDamage[build.WeaponSpecialization]; build.WeaponSpecialization != "" && !known {
		report("build.weapon_specialization %q: want normal or greater", build.WeaponSpecialization)
	}
	if build.CriticalSpecialization != "" && slices.Index(RankWords, build.CriticalSpecialization) < 1 {
		report("build.critical_specialization %q: want a rank from trained to legendary", build.CriticalSpecialization)
	}
```

In `Build.strike`:

**Find:**
```go
	return Strike{
		Name:   weapon.Name,
		Bonus:  attack,
		Range:  weapon.Range,
		Traits: slices.Clone(weapon.Traits),
		Damage: build.weaponDamage(level, weapon, expression),
	}
}
```
**Replace with:**
```go
	return Strike{
		Name:     weapon.Name,
		Bonus:    attack,
		Range:    weapon.Range,
		Traits:   slices.Clone(weapon.Traits),
		Damage:   build.weaponDamage(level, weapon, expression),
		Outcomes: build.riders(weapon, rank),
	}
}

func (build *Build) riders(weapon Weapon, rank int) Outcomes {
	outcomes := weapon.Outcomes
	minimum := slices.Index(RankWords, build.CriticalSpecialization)
	if specialization, found := criticalSpecializations[weapon.Group]; found && minimum > 0 && rank >= minimum {
		outcomes.CriticalSuccess = append(slices.Clone(outcomes.For(3, true)), specialization)
	}
	return outcomes
}
```

##### Task 4 — `internal/pf2e/sim/engine/damage.go`

Defenses read the combatant: applied weaknesses, lowered resistances and Protean Anatomy join the type-keyed adjustments, and a trait-keyed weakness or resistance applies once per hit. `adjustedDamage` becomes a method on `Combatant`.

**Find:**
```go
func adjustedDamage(defense creature.Defense, damageType string, traits []string, amount int) int {
	immune := slices.ContainsFunc(defense.Immune, func(key string) bool { return adjustmentCovers(key, damageType) })
	if immune {
		return 0
	}
	if weakness, found := strongestAdjustment(defense.Weakness, damageType, traits); found && amount > 0 {
		amount += weakness
	}
	if resistance, found := strongestAdjustment(defense.Resist, damageType, traits); found {
		amount -= resistance
	}
	return max(amount, 0)
}

func strongestAdjustment(adjustments map[string]creature.Adjustment, damageType string, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if !adjustmentCovers(key, damageType) {
			continue
		}
```
**Replace with:**
```go
func (target *Combatant) adjustedDamage(damageType string, traits []string, amount int) int {
	defense := target.Creature.Defense
	immune := slices.ContainsFunc(defense.Immune, func(key string) bool { return adjustmentCovers(key, damageType) })
	if immune {
		return 0
	}
	weakness, hasWeakness := strongestAdjustment(defense.Weakness, damageType, traits)
	for _, applied := range target.defenseChanges {
		if applied.change.Weakness > weakness && slices.Contains(applied.change.Types, damageType) {
			weakness, hasWeakness = applied.change.Weakness, true
		}
	}
	if hasWeakness && amount > 0 {
		amount += weakness
	}
	resistance, hasResistance := strongestAdjustment(defense.Resist, damageType, traits)
	if adaptive := defense.Adaptive; adaptive != nil && target.adaptiveType == damageType && adaptive.Value > resistance {
		resistance, hasResistance = adaptive.Value, true
	}
	if hasResistance {
		lowered := 0
		for _, applied := range target.defenseChanges {
			if slices.Contains(applied.change.Types, damageType) {
				lowered = max(lowered, applied.change.Resistance)
			}
		}
		amount -= max(resistance-lowered, 0)
	}
	return max(amount, 0)
}

func traitAdjustment(adjustments map[string]creature.Adjustment, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if !creature.IsDefenseTrait(key) || !slices.Contains(traits, key) {
			continue
		}
		if !found || adjustment.Value > strongest {
			strongest, found = adjustment.Value, true
		}
	}
	return strongest, found
}

func strongestAdjustment(adjustments map[string]creature.Adjustment, damageType string, traits []string) (int, bool) {
	strongest, found := 0, false
	for key, adjustment := range adjustments {
		if creature.IsDefenseTrait(key) || !adjustmentCovers(key, damageType) {
			continue
		}
```

In `applyDamage`, use the method, apply trait adjustments once, and adapt afterwards:

**Find:**
```go
	for _, part := range roll.parts {
		taken := adjustedDamage(target.Creature.Defense, part.damageType, source.traits, part.amount)
		parts = append(parts, DamageAmount{Type: part.damageType, Rolled: part.amount, Taken: taken})
		total += taken
	}
```
**Replace with:**
```go
	for _, part := range roll.parts {
		taken := target.adjustedDamage(part.damageType, source.traits, part.amount)
		parts = append(parts, DamageAmount{Type: part.damageType, Rolled: part.amount, Taken: taken})
		total += taken
	}
	traitChange := 0
	if weakness, found := traitAdjustment(target.Creature.Defense.Weakness, source.traits); found && total > 0 {
		traitChange += weakness
	}
	if resistance, found := traitAdjustment(target.Creature.Defense.Resist, source.traits); found {
		traitChange -= min(resistance, total+traitChange)
	}
	total += traitChange
```

**Find:**
```go
		Total:    total,
		Absorbed: absorbed,
		HPAfter:  target.HP,
	})
```
**Replace with:**
```go
		Total:    total,
		Absorbed: absorbed,
		HPAfter:  target.HP,
		Trait:    traitChange,
	})
	fight.adapt(target, parts)
```

Add after `applyDamage`:

```go
func (fight *Fight) adapt(target *Combatant, parts []DamageAmount) {
	adaptive := target.Creature.Defense.Adaptive
	if adaptive == nil || target.Dead {
		return
	}
	for _, part := range parts {
		if part.Taken == 0 || !slices.Contains(adaptive.Types, part.Type) {
			continue
		}
		if part.Type != target.adaptiveType {
			target.adaptiveType = part.Type
			fight.observer.Observe(ResistanceAdapted{Combatant: target.ID, Type: part.Type, Value: adaptive.Value})
		}
		return
	}
}
```

##### Task 5 — `internal/pf2e/sim/engine/conditions.go`, `steps.go` and `combatant.go`

Applied defense changes keep the same clocks as conditions.

Add to `conditions.go`:

```go
type defenseChange struct {
	change creature.DefenseChange
	clock  timing
}

func (fight *Fight) weaken(source, target *Combatant, step creature.Step) {
	if target.Dead {
		return
	}
	change := *step.Weaken
	target.defenseChanges = append(target.defenseChanges, &defenseChange{change: change, clock: fight.newTiming(source, target, step.Duration)})
	fight.observer.Observe(DefenseChanged{Combatant: target.ID, Source: source.ID, Weakness: change.Weakness, Resistance: change.Resistance, Types: change.Types})
}

func (fight *Fight) expireDefenseChanges(bearer, actor *Combatant, boundary string) {
	var kept []*defenseChange
	for _, applied := range bearer.defenseChanges {
		if applied.clock.expires(actor, bearer, boundary) {
			fight.observer.Observe(DefenseRestored{Combatant: bearer.ID, Types: applied.change.Types})
			continue
		}
		kept = append(kept, applied)
	}
	bearer.defenseChanges = kept
}
```

**Find:**
```go
func (fight *Fight) startTurn(actor *Combatant) {
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "start")
	}
}
```
**Replace with:**
```go
func (fight *Fight) startTurn(actor *Combatant) {
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "start")
		fight.expireDefenseChanges(bearer, actor, "start")
	}
}
```

**Find:**
```go
	fight.decayConditions(actor)
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "end")
	}
}
```
**Replace with:**
```go
	fight.decayConditions(actor)
	for _, bearer := range fight.combatants {
		fight.expire(bearer, actor, "end")
		fight.expireDefenseChanges(bearer, actor, "end")
	}
}
```

In `runStep` (`steps.go`):

**Find:**
```go
	case creature.StepTempHP:
		amount := step.TempHPRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.gainTempHP(target, amount)
		}
```
**Replace with:**
```go
	case creature.StepTempHP:
		amount := step.TempHPRoll.Roll(fight.roller)
		for _, target := range targets {
			fight.gainTempHP(target, amount)
		}
	case creature.StepWeaken:
		for _, target := range targets {
			fight.weaken(source.actor, target, step)
		}
```

In `Combatant`:

**Find:**
```go
	conditions      []*conditionInstance
	persistent      []*persistentInstance
}
```
**Replace with:**
```go
	conditions      []*conditionInstance
	persistent      []*persistentInstance
	defenseChanges  []*defenseChange
	adaptiveType    string
}
```

##### Task 6 — `internal/pf2e/sim/engine/checks.go` and `strike.go`

A modifier keyed on a condition reads the target's value.

**Find:**
```go
func (fight *Fight) modifierTotal(combatant *Combatant, selectors, traits []string) int {
	return fight.modifierPool(combatant, selectors, traits).Total()
}

func (fight *Fight) modifierPool(combatant *Combatant, selectors, traits []string) *creature.BonusPool {
	pool := creature.NewBonusPool()
	for _, modifier := range combatant.Creature.Modifiers {
		if modifier.Applies(selectors, traits) {
			pool.Add(modifier.Type, modifier.Value)
		}
	}
```
**Replace with:**
```go
func (fight *Fight) modifierTotal(combatant *Combatant, selectors, traits []string) int {
	return fight.modifierPool(combatant, selectors, traits, nil).Total()
}

func (fight *Fight) modifierPool(combatant *Combatant, selectors, traits []string, target *Combatant) *creature.BonusPool {
	pool := creature.NewBonusPool()
	for _, modifier := range combatant.Creature.Modifiers {
		if !modifier.Applies(selectors, traits) {
			continue
		}
		value := modifier.Value
		if modifier.PerCondition != "" {
			if target == nil {
				continue
			}
			value *= target.Condition(modifier.PerCondition)
		}
		if value != 0 {
			pool.Add(modifier.Type, value)
		}
	}
```

In `strike`, the target counts for attack and damage, and the Strike's riders run for every degree:

**Find:**
```go
	bonus := strike.Bonus + fight.modifierTotal(attacker, attackSelectors, nil)
```
**Replace with:**
```go
	bonus := strike.Bonus + fight.modifierPool(attacker, attackSelectors, nil, target).Total()
```

**Find:**
```go
		damageBonus := fight.modifierTotal(attacker, damageSelectors, nil)
		fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with Effort 7 (Decision 11).
		if splash := splashParts(strike.Damage); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
}
```
**Replace with:**
```go
		damageBonus := fight.modifierPool(attacker, damageSelectors, nil, target).Total()
		fight.applyDamage(source, target, fight.rollDamage(strike.Damage, critical, damageBonus), critical)
	case degree == Failure:
		// deliberate: splash lands on the primary target only, not on its neighbours, because position comes with Effort 7 (Decision 11).
		if splash := splashParts(strike.Damage); len(splash) > 0 {
			fight.applyDamage(source, target, fight.rollDamage(splash, false, 0), false)
		}
	}
	if !target.Dead {
		fight.runSteps(source, strike.For(int(degree), true), target)
	}
}
```

##### Task 7 — `internal/pf2e/sim/engine/events.go`

**Find:**
```go
	Absorbed int
	HPAfter  int
}
```
**Replace with:**
```go
	Absorbed int
	HPAfter  int
	Trait    int
}

type ResistanceAdapted struct {
	Combatant string
	Type      string
	Value     int
}

type DefenseChanged struct {
	Combatant  string
	Source     string
	Weakness   int
	Resistance int
	Types      []string
}

type DefenseRestored struct {
	Combatant string
	Types     []string
}
```

**Find:**
```go
func (PersistentEnded) isEvent()   {}
```
**Replace with:**
```go
func (PersistentEnded) isEvent()   {}
func (ResistanceAdapted) isEvent() {}
func (DefenseChanged) isEvent()    {}
func (DefenseRestored) isEvent()   {}
```

##### Task 8 — `cmd/gm-toolkit/pf2e/show.go`

**Find:**
```go
	printAdjustments(out, "Resist", defense.Resist)
	printAdjustments(out, "Weakness", defense.Weakness)
```
**Replace with:**
```go
	printAdjustments(out, "Resist", defense.Resist)
	printAdjustments(out, "Weakness", defense.Weakness)
	if adaptive := defense.Adaptive; adaptive != nil {
		fmt.Fprintf(out, "Adaptive resistance %d (%s)\n", adaptive.Value, strings.Join(adaptive.Types, ", "))
	}
```

**Find:**
```go
			if len(modifier.Against) > 0 {
				line += " against " + strings.Join(modifier.Against, ", ")
			}
			fmt.Fprintln(out, line)
```
**Replace with:**
```go
			if len(modifier.Against) > 0 {
				line += " against " + strings.Join(modifier.Against, ", ")
			}
			if modifier.PerCondition != "" {
				line += " per " + modifier.PerCondition
			}
			fmt.Fprintln(out, line)
```

##### Task 9 — tests

`internal/pf2e/sim/engine/rules_test.go`: `TestAdjustedDamage` and `TestAdjustedDamageImmunityGroups` build a combatant around the `Defense` under test (`NewCombatant(SideEnemies, &creature.Creature{Defense: defense})`; one for `TestAdjustedDamage`'s shared defense) and call its `adjustedDamage(damageType, traits, amount)`. `TestUntypedDamageTakesResistAll` in `down_test.go` does the same. The expected values do not change. In `TestLoadRejectsBadConditions`, the `value on a save` case now wants `value belongs on a condition` (Decision Record 8).

`internal/pf2e/sim/creature/steps_test.go`:

- **`TestLoadReadsDefensesAndRiders`** — `[defense.weakness.holy]` value 15, `[defense.resist.all]` value 12 except `["holy", "cold-iron"]`, and `[defense.adaptive]` value 25 over acid, electricity and sonic load; a Strike with a `success` list holding a Will save (`unless = "frightened"`, failure gives frightened 2) loads; a `weaken` step with weakness 3 over acid, fire and slashing and `duration = { rounds = 3 }` loads; a modifier with `per_condition = "frightened"` on `damage` loads.
- **`TestLoadRejectsBadDefenses`** — weakness key `chaos` (`one of adamantine, cold-iron`); adaptive with type `sad` (`"sad" is not a damage type`); `weaken` with neither number (`want a weakness or a resistance above 0`); `duration` on a damage step (`duration belongs on a condition or weaken`); `per_condition` on `ac` (`per_condition needs a condition`).
- **`TestCriticalSpecializationFromBuild`** — a PC build with `critical_specialization = "master"`, a legendary hammer with its own `critical_success` list of one condition, an untrained advanced hammer and a dagger: the legendary hammer's Strike has two critical riders (its own, then the Fortitude save against the class DC); the advanced hammer's (below master) and the dagger's (no group row) have none.

`internal/pf2e/sim/engine/defenses_test.go` (new):

- **`TestTraitWeaknessAppliesOnce`** — a `holy` Strike with slashing and spirit parts against weakness holy 15: `Trait` is 15 and the total rises by 15, once.
- **`TestResistAllExceptTraits`** — resistance all 12 except cold iron: a `cold-iron` Strike takes no resistance; a plain one takes 12 off each type.
- **`TestAdaptiveResistance`** — the first acid hit is not resisted and fires `ResistanceAdapted`; the next acid hit loses 25; an electricity hit is not resisted, then switches it; against resistance all 12 the adaptive 25 wins for its type.
- **`TestAppliedWeakness`** — a `weaken` of weakness 3 fire for 3 rounds adds 3 to fire damage and ends at the source's third `startTurn` with `DefenseRestored`.
- **`TestLoweredResistance`** — a `weaken` of resistance 3 over bludgeoning, piercing and slashing lowers resistance all 12 to 9 for slashing and leaves fire at 12.
- **`TestStrikeRiders`** — a hit runs the `success` list; a critical hit runs the `critical_success` list alone when there is one, and the `success` list when there is not; a miss runs neither.
- **`TestRiderUnlessFrightened`** — a Terrifying Touch rider does not run its save against a frightened target.
- **`TestPerConditionDamage`** — Fearsome Brute (+3 circumstance damage per frightened) against frightened 2 adds 6; against a target with no frightened it adds nothing.
- **`TestRunesFromALoadedLayer`** — `creature.Load` reads a PC with a hammer weapon whose `critical_success` list gives clumsy 2 and enfeebled 2 until `source-turn-end` and frightened 2, plus `critical_specialization = "master"`; a scripted critical hit fires `ConditionChanged` for all three and a `SaveRolled` for Fortitude.

##### Commit message

```
feat(E3): key defenses on traits and add Strike riders

- weaknesses and resistances keyed on a trait or material (holy, cold iron) apply once per hit
- adaptive resistance (Protean Anatomy); weaken steps add a weakness or lower resistances for a duration
- Strikes and PC weapons carry riders by degree; critical specialization by weapon group (hammer)
- modifiers keyed on the target's condition (Fearsome Brute)
- new events ResistanceAdapted, DefenseChanged and DefenseRestored; DamageDealt shows the trait adjustment
```

### Phase 4 — Import and wiki

**In plain words:** the importer reads the sheet rules that E3 can now use (degree shifts, the dying limit and recovery DC, the initiative bonus, critical specialization, bomb splash and persistent damage, cold iron, the Avatar's holy weakness), the party gets hand layers with their hero points and Kas's runes, and the wiki describes the effect core. When the phase lands, `show kas` prints his degree shifts and dying 6, and the Fall of Luna import layers are committed again.

#### Commit 7 — `feat(E3): import degree shifts, dying limits and damage categories`

##### Task 1 — `internal/pf2e/sim/foundry/report.go`, then every use of `ReasonEffort3`

**Find:**
```go
	ReasonAbility     = "ability not converted yet (rules in Effort 3, descriptions in Effort 4)"
	ReasonEffort3     = "needs a rule from Effort 3"
```
**Replace with:**
```go
	ReasonAbility     = "ability not converted yet (Effort 6)"
	ReasonLaterRule   = "needs a rule from a later effort"
	ReasonHandLayer   = "typed into the hand layer (overview Decision 11)"
```

Then replace every `ReasonEffort3` with `ReasonLaterRule` (Edit with `replace_all`) in `foundry/pc.go`, `rules.go`, `npc.go`, `npc_test.go`, `pc_test.go`, `rules_test.go` and `cmd/gm-toolkit/pf2e/import_test.go`. The tasks below then change the uses that E3 resolves.

##### Task 2 — `internal/pf2e/sim/creature/build.go`

A build modifier can add an attribute's modifier to splash (Expanded Splash adds Intelligence).

**Find:**
```go
	PerLevel     int      `toml:"per_level,omitzero"`
	WeaponTraits []string `toml:"weapon_traits,omitempty"`
}
```
**Replace with:**
```go
	PerLevel     int      `toml:"per_level,omitzero"`
	Attribute    string   `toml:"attribute,omitempty"`
	WeaponTraits []string `toml:"weapon_traits,omitempty"`
}
```

In `checkBuild`:

**Find:**
```go
		if len(modifier.WeaponTraits) > 0 && !slices.Contains(weaponSelectors, modifier.Selector) {
			report("build.modifier.%s: weapon_traits needs selector attack or damage", id)
		}
```
**Replace with:**
```go
		if len(modifier.WeaponTraits) > 0 && !slices.Contains(weaponSelectors, modifier.Selector) {
			report("build.modifier.%s: weapon_traits needs selector attack, damage or splash", id)
		}
		if modifier.Attribute != "" && (modifier.Selector != "splash" || !slices.Contains(AttributeNames, modifier.Attribute)) {
			report("build.modifier.%s: attribute needs selector splash and an attribute name", id)
		}
```

**Find:**
```go
		Damage:   build.weaponDamage(level, weapon, expression),
```
**Replace with:**
```go
		Damage:   build.weaponDamage(level, attributes, weapon, expression),
```

**Find:**
```go
func (build *Build) weaponDamage(level int, weapon Weapon, expression dice.Expression) []DamagePart {
```
**Replace with:**
```go
func (build *Build) weaponDamage(level int, attributes map[string]int, weapon Weapon, expression dice.Expression) []DamagePart {
```

**Find:**
```go
		splash := weapon.Splash + build.total(level, 0, weapon.Traits, "splash")
```
**Replace with:**
```go
		splash := weapon.Splash + build.total(level, 0, weapon.Traits, "splash") + build.attributeBonus(attributes, weapon.Traits, "splash")
```

Add below `weaponDamage`:

```go
func (build *Build) attributeBonus(attributes map[string]int, weaponTraits []string, selector string) int {
	bonus := 0
	for _, modifier := range build.Modifiers {
		if modifier.Selector == selector && modifier.Attribute != "" && containsAll(weaponTraits, modifier.WeaponTraits) {
			bonus += attributes[modifier.Attribute]
		}
	}
	return bonus
}
```

##### Task 3 — `internal/pf2e/sim/foundry/effects.go` (new)

Full contents:

```go
package foundry

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const baseRecoveryDC = 10

var (
	degreeKeys       = map[string]string{"criticalFailure": "critical-failure", "failure": "failure", "success": "success", "criticalSuccess": "critical-success"}
	degreeSteps      = map[string]int{"one-degree-better": 1, "one-degree-worse": -1}
	shiftSelectors   = []string{"saving-throw", "fortitude", "reflex", "will"}
	attributeFormula = regexp.MustCompile(`^@actor\.abilities\.(str|dex|con|int|wis|cha)\.mod$`)
)

func (importer *pcImporter) degreeShift(current *item, rule map[string]any, options rollOptions) {
	selector, _ := rule["selector"].(string)
	what := "AdjustDegreeOfSuccess " + selector
	if !slices.Contains(shiftSelectors, selector) {
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	adjustment, _ := rule["adjustment"].(map[string]any)
	if len(adjustment) == 0 {
		importer.report.skip(current.Name, what+" without an adjustment", ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	against, rest := splitTraitTerms(predicate)
	held, err := options.holds(rest)
	switch {
	case err != nil:
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	case len(against) > 1:
		importer.report.skip(current.Name, what+" against "+strings.Join(against, " and "), ReasonUnsupported)
		return
	}
	for _, key := range slices.Sorted(maps.Keys(adjustment)) {
		text, _ := adjustment[key].(string)
		from, knownDegree := degreeKeys[key]
		step, knownStep := degreeSteps[text]
		to := slices.Index(creature.DegreeWords, from) + step
		if !knownDegree || !knownStep || to < 0 || to >= len(creature.DegreeWords) {
			importer.report.skip(current.Name, fmt.Sprintf("%s %s %s", what, key, text), ReasonUnsupported)
			continue
		}
		importer.degreeShifts[uniqueID(current.Name, importer.fightIDs)] = creature.DegreeShift{Selector: selector, Against: against, From: from, To: creature.DegreeWords[to]}
	}
}

// deliberate: a bare trait term ("fear", "item:trait:emotion") names a trait of the effect, not an actor option, because Foundry tests degree adjustments against the incoming roll's traits.
func splitTraitTerms(predicate []any) ([]string, []any) {
	var traits []string
	var rest []any
	for _, term := range predicate {
		text, isText := term.(string)
		trait := strings.TrimPrefix(text, "item:trait:")
		if isText && !strings.Contains(trait, ":") {
			traits = append(traits, trait)
			continue
		}
		rest = append(rest, term)
	}
	return traits, rest
}

func (importer *pcImporter) criticalSpecialization(current *item, rule map[string]any, options rollOptions) {
	predicate, _ := rule["predicate"].([]any)
	minimum := 0
	var rest []any
	for _, term := range predicate {
		if rank, isRank := proficiencyRankAtLeast(term); isRank {
			minimum = rank
			continue
		}
		rest = append(rest, term)
	}
	held, err := options.holds(rest)
	switch {
	case err != nil:
		importer.report.skip(current.Name, "CriticalSpecialization: "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, "CriticalSpecialization", ReasonConditional)
		return
	case minimum < 1 || minimum >= len(creature.RankWords):
		importer.report.skip(current.Name, "CriticalSpecialization without a rank", ReasonUnsupported)
		return
	}
	existing := importer.build.CriticalSpecialization
	if existing == "" || minimum < slices.Index(creature.RankWords, existing) {
		importer.build.CriticalSpecialization = creature.RankWords[minimum]
	}
}

func proficiencyRankAtLeast(term any) (int, bool) {
	comparison, _ := term.(map[string]any)
	operands, found := comparison["gte"].([]any)
	if !found || len(operands) != 2 || operands[0] != "item:proficiency:rank" {
		return 0, false
	}
	rank, isNumber := operands[1].(float64)
	return int(rank), isNumber
}

func (importer *pcImporter) adjustStrike(current *item, rule map[string]any) {
	property, _ := rule["property"].(string)
	value, _ := rule["value"].(string)
	what := "AdjustStrike " + property
	if property != "materials" {
		importer.report.skip(current.Name, what, ReasonLaterRule)
		return
	}
	id, found := importer.weaponByItem[current.ID]
	if !found || !creature.IsDefenseTrait(value) {
		importer.report.skip(current.Name, what+" "+value, ReasonUnsupported)
		return
	}
	weapon := importer.build.Weapons[id]
	weapon.Traits = append(slices.Clone(weapon.Traits), value)
	importer.build.Weapons[id] = weapon
}

func (importer *pcImporter) initiativeModifier(current *item, rule map[string]any) {
	predicate, _ := rule["predicate"].([]any)
	for _, term := range predicate {
		if term != "perception" {
			importer.report.skip(current.Name, "FlatModifier initiative", ReasonConditional)
			return
		}
	}
	kind, _ := rule["type"].(string)
	bonusType, known := bonusTypes[kind]
	value, err := importer.number(rule["value"])
	if !known || err != nil || value == 0 {
		importer.report.skip(current.Name, fmt.Sprintf("FlatModifier initiative %v", rule["value"]), ReasonUnsupported)
		return
	}
	importer.fightModifiers[uniqueID(current.Name, importer.fightIDs)] = creature.CheckModifier{Selector: "initiative", Type: bonusType, Value: value}
}

func (importer *pcImporter) splashModifier(current *item, rule map[string]any) {
	value, _ := rule["value"].(string)
	match := attributeFormula.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		importer.report.skip(current.Name, "FlatModifier splash "+value, ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	traits, applies := importer.weaponTraits(predicate)
	if !applies {
		importer.report.skip(current.Name, "FlatModifier splash", ReasonConditional)
		return
	}
	importer.build.Modifiers[uniqueID(current.Name, importer.modifierIDs)] = creature.Modifier{Selector: "splash", Attribute: match[1], WeaponTraits: traits}
}

func (importer *pcImporter) dyingRule(current *item, rule map[string]any, options rollOptions, path string) {
	what := "ActiveEffectLike " + path
	predicate, _ := rule["predicate"].([]any)
	held, err := options.holds(predicate)
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	if !held {
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	}
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	mode, _ := rule["mode"].(string)
	switch {
	case path == "system.attributes.dying.max" && mode == "upgrade":
		importer.dyingMax = max(importer.dyingMax, value)
	case path == "system.attributes.dying.recoveryDC" && mode == "subtract":
		importer.recoverySubtract += value
	case path == "system.attributes.dying.recoveryDC" && mode == "downgrade":
		if importer.recoveryFloor == 0 || value < importer.recoveryFloor {
			importer.recoveryFloor = value
		}
	default:
		importer.report.skip(current.Name, what+" mode "+mode, ReasonUnsupported)
	}
}

// deliberate: subtractions apply before downgrades, not in item order, because Foundry applies ActiveEffectLike modes by priority (subtract before downgrade).
func (importer *pcImporter) recoveryDC() int {
	dc := baseRecoveryDC - importer.recoverySubtract
	if importer.recoveryFloor > 0 {
		dc = min(dc, importer.recoveryFloor)
	}
	return dc
}
```

##### Task 4 — `internal/pf2e/sim/foundry/rules.go`

**Find:**
```go
	laterSelectors = []string{"initiative", "unarmed-attack", "unarmed-damage", "alchemical-bomb-damage", "melee-strike-damage"}
```
**Replace with:**
```go
	laterSelectors = []string{"unarmed-attack", "unarmed-damage", "alchemical-bomb-damage", "melee-strike-damage"}
```

**Find:**
```go
	laterRuleKeys  = []string{"AdjustDegreeOfSuccess", "DamageDice", "CriticalSpecialization", "AdjustStrike", "Strike", "Weakness", "Immunity"}
```
**Replace with:**
```go
	laterRuleKeys  = []string{"DamageDice", "Strike", "Weakness", "Immunity"}
```

In `applyRule`:

**Find:**
```go
	case key == "AdjustModifier":
		importer.adjustModifier(current, rule)
```
**Replace with:**
```go
	case key == "AdjustModifier":
		importer.adjustModifier(current, rule, options)
	case key == "AdjustDegreeOfSuccess":
		importer.degreeShift(current, rule, options)
	case key == "CriticalSpecialization":
		importer.criticalSpecialization(current, rule, options)
	case key == "AdjustStrike":
		importer.adjustStrike(current, rule)
```

In `flatModifier`:

**Find:**
```go
	for _, field := range []string{"damageCategory", "damageType", "critical"} {
```
**Replace with:**
```go
	if category, _ := rule["damageCategory"].(string); category == "splash" {
		importer.splashModifier(current, rule)
		return
	}
	for _, field := range []string{"damageCategory", "damageType", "critical"} {
```

In `flatModifierFor`:

**Find:**
```go
	what := "FlatModifier " + selector
	stat, isStat := statSelectors[selector]
```
**Replace with:**
```go
	if selector == "initiative" {
		importer.initiativeModifier(current, rule)
		return
	}
	what := "FlatModifier " + selector
	stat, isStat := statSelectors[selector]
```

In `activeEffectLike`:

**Find:**
```go
	case strings.HasPrefix(path, "system.attributes.dying."):
		importer.report.skip(current.Name, what, ReasonLaterRule)
		return
```
**Replace with:**
```go
	case strings.HasPrefix(path, "system.attributes.dying."):
		importer.dyingRule(current, rule, options, path)
		return
```

`adjustModifier` reports a rule whose predicate does not hold as conditional (Calculated Splash, with Expanded Splash taken):

**Find:**
```go
func (importer *pcImporter) adjustModifier(current *item, rule map[string]any) {
	slug, _ := rule["slug"].(string)
```
**Replace with:**
```go
func (importer *pcImporter) adjustModifier(current *item, rule map[string]any, options rollOptions) {
	slug, _ := rule["slug"].(string)
	selectors := append(textList(rule["selector"]), textList(rule["selectors"])...)
	what := "AdjustModifier " + strings.Join(selectors, ", ")
	predicate, _ := rule["predicate"].([]any)
	if held, err := options.holds(predicate); err == nil && !held {
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	}
```

The unsupported skip at the end of the function reuses `what` (Decision Record 10).

##### Task 5 — `internal/pf2e/sim/foundry/pc.go`

**Find:**
```go
	weaponIDs              map[string]bool
	modifierIDs            map[string]bool
	entries                []*item
	spellsByEntry          map[string]int
}
```
**Replace with:**
```go
	weaponIDs              map[string]bool
	modifierIDs            map[string]bool
	fightIDs               map[string]bool
	weaponByItem           map[string]string
	fightModifiers         map[string]creature.CheckModifier
	degreeShifts           map[string]creature.DegreeShift
	dyingMax               int
	recoverySubtract       int
	recoveryFloor          int
	entries                []*item
	spellsByEntry          map[string]int
}
```

**Find:**
```go
		weaponIDs:     map[string]bool{},
		modifierIDs:   map[string]bool{},
		spellsByEntry: map[string]int{},
	}
```
**Replace with:**
```go
		weaponIDs:      map[string]bool{},
		modifierIDs:    map[string]bool{},
		fightIDs:       map[string]bool{},
		weaponByItem:   map[string]string{},
		fightModifiers: map[string]creature.CheckModifier{},
		degreeShifts:   map[string]creature.DegreeShift{},
		spellsByEntry:  map[string]int{},
	}
```

In `item`, a destroyed shield is not in any effort's scope:

**Find:**
```go
		importer.report.skip(current.Name, "shield", ReasonLaterRule)
```
**Replace with:**
```go
		importer.report.skip(current.Name, "shield", ReasonUnsupported)
```

In `armorItem`, property runes go in the hand layer:

**Find:**
```go
	importer.armor = &system
	importer.armorName = current.Name
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonLaterRule)
	}
```
**Replace with:**
```go
	importer.armor = &system
	importer.armorName = current.Name
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonHandLayer)
	}
```

In `weapon`, persistent damage and splash are read, and the weapon's ID is kept by item:

**Find:**
```go
	if persistent := system.Damage.Persistent; persistent != nil {
		importer.report.skip(current.Name, fmt.Sprintf("persistent %d %s", persistent.Number, persistent.Type), ReasonLaterRule)
	}
	if system.SplashDamage.Value > 0 {
		importer.report.skip(current.Name, fmt.Sprintf("splash %d", system.SplashDamage.Value), ReasonLaterRule)
	}
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonLaterRule)
	}
	if slices.Contains(weapon.Traits, "consumable") {
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, not tracked", current.System.Quantity), ReasonLaterRule)
	}
	importer.build.Weapons[uniqueID(current.Name, importer.weaponIDs)] = weapon
	return nil
```
**Replace with:**
```go
	if persistent := system.Damage.Persistent; persistent != nil {
		if creature.IsDamageType(persistent.Type) && persistent.Number > 0 {
			weapon.Persistent = &creature.DamagePart{Dice: strconv.Itoa(persistent.Number), Type: persistent.Type}
		} else {
			importer.report.skip(current.Name, fmt.Sprintf("persistent %d %s", persistent.Number, persistent.Type), ReasonUnsupported)
		}
	}
	weapon.Splash = max(system.SplashDamage.Value, 0)
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonHandLayer)
	}
	if slices.Contains(weapon.Traits, "consumable") {
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, not tracked", current.System.Quantity), ReasonLaterRule)
	}
	id := uniqueID(current.Name, importer.weaponIDs)
	importer.build.Weapons[id] = weapon
	importer.weaponByItem[current.ID] = id
	return nil
```

In `finish`:

**Find:**
```go
	if len(importer.resist) > 0 {
		layer.Defense.Resist = importer.resist
	}
	return layer
}
```
**Replace with:**
```go
	if len(importer.resist) > 0 {
		layer.Defense.Resist = importer.resist
	}
	layer.DyingMax = importer.dyingMax
	if dc := importer.recoveryDC(); dc != baseRecoveryDC {
		layer.RecoveryDC = dc
	}
	if len(importer.fightModifiers) > 0 {
		layer.Modifiers = importer.fightModifiers
	}
	if len(importer.degreeShifts) > 0 {
		layer.DegreeShifts = importer.degreeShifts
	}
	return layer
}
```

##### Task 6 — `internal/pf2e/sim/foundry/npc.go`

Weaknesses and resistances keyed on a trait come across (the Avatar's holy 15):

**Find:**
```go
		if key != "all" && key != "physical" && !creature.IsDamageType(key) {
```
**Replace with:**
```go
		if key != "all" && key != "physical" && !creature.IsDamageType(key) && !creature.IsDefenseTrait(key) {
```

Persistent and splash parts of an NPC Strike come across with their category:

**Find:**
```go
		if roll.Category != "" {
			report.skip(current.Name, what+" "+roll.Category, ReasonLaterRule)
			continue
		}
```
**Replace with:**
```go
		if roll.Category != "" && roll.Category != "persistent" && roll.Category != "splash" {
			report.skip(current.Name, what+" "+roll.Category, ReasonLaterRule)
			continue
		}
```

**Find:**
```go
		strike.Damage = append(strike.Damage, creature.DamagePart{Dice: roll.Damage, Type: roll.DamageType})
	}
```
**Replace with:**
```go
		strike.Damage = append(strike.Damage, creature.DamagePart{Dice: roll.Damage, Type: roll.DamageType, Category: roll.Category})
	}
```

**Find:**
```go
	if len(strike.Damage) == 0 {
		report.skip(current.Name, "strike with no damage the sim can roll", ReasonUnsupported)
		return nil
	}
```
**Replace with:**
```go
	if !slices.ContainsFunc(strike.Damage, func(part creature.DamagePart) bool { return part.Category == "" }) {
		report.skip(current.Name, "strike with no damage the sim can roll", ReasonUnsupported)
		return nil
	}
```

Add `"slices"` to the imports of `npc.go`.

##### Task 7 — tests

`internal/pf2e/sim/foundry/effects_test.go` (new), against the `testdata` exports:

- **`TestImportDegreeShifts`** — Kas's layer holds Juggernaut (`fortitude`, success to critical success), Bravery (`will` against `fear`), Gutsy Halfling (`saving-throw` against `emotion`) and Irrepressible's two shifts (success up, and critical failure to failure, both against `emotion`); Inextinguishable's rule is skipped as `AdjustDegreeOfSuccess saving-throw without an adjustment`, `ReasonUnsupported`.
- **`TestImportDyingLimits`** — Kas's layer has `DyingMax` 6 and `RecoveryDC` 6; Mountain's Stoutness's first rule is skipped as conditional; Tavish's and Arabesque's layers leave both at 0.
- **`TestImportInitiative`** — Kas's layer holds a `battlefield-surveyor` modifier: `initiative`, circumstance, +2.
- **`TestImportCriticalSpecialization`** — Kas's build has `critical_specialization = "master"`; Arabesque's has none, and her Weapon Expertise rule is skipped as conditional.
- **`TestImportBombDamage`** — Tavish's Alchemist's Fire (Greater) has `Persistent` `3 fire` and `Splash` 3; his build holds an Expanded Splash modifier (`splash`, attribute `int`, weapon traits `splash`); after `creature.Load` on the written folder the Strike's damage ends with `7 fire splash` (3 plus Intelligence +4). Calculated Splash is skipped as conditional.
- **`TestImportColdIron`** — Tavish's Nail Bomb (Greater) weapon carries the `cold-iron` trait.

Updates to existing tests for what E3 now imports: in `rules_test.go`, Bravery and Toughness are no longer skipped; in `npc_test.go`, the Avatar's `weakness holy 15` is no longer skipped and is in `Defense.Weakness`; in `pc_test.go`, Alchemist's Fire (Greater) is no longer skipped for persistent damage or splash; in `cmd/gm-toolkit/pf2e/import_test.go`, the expected report text follows the new reason strings.

##### Commit message

```
feat(E3): import degree shifts, dying limits and damage categories

- AdjustDegreeOfSuccess becomes degree shifts against the effect's traits
- the dying maximum and the recovery DC from ActiveEffectLike rules
- the initiative bonus (Battlefield Surveyor) and critical specialization by rank
- bomb persistent damage and splash; Expanded Splash adds Intelligence; cold iron from AdjustStrike
- NPC weaknesses keyed on a trait (holy) and Strike parts with a category
- report reasons follow the split: later rules, Effort 6 for abilities, hand layers for runes
```

#### Commit 8 — `docs(E3): document the effect core and add the party's hand layers`

##### Task 1 — `campaigns/hawk/pf2e/party/kas/hand.toml` (new)

Full contents (Decisions 4, 16):

```toml
schema = 1

hero_points = 1

# Greater Crushing: clumsy 2 and enfeebled 2 until the end of Kas's next turn.
# Greater Fearsome: frightened 2. Both on a critical hit.
[build.weapon.protectors-fury]
critical_success = [
  { condition = "clumsy", value = 2, duration = { until = "source-turn-end" } },
  { condition = "enfeebled", value = 2, duration = { until = "source-turn-end" } },
  { condition = "frightened", value = 2 },
]
```

##### Task 2 — `campaigns/hawk/pf2e/party/tavish/hand.toml` and `arabesque/hand.toml` (new)

Full contents of each:

```toml
schema = 1

hero_points = 1
```

##### Task 3 — the Fall of Luna import, on Claude's copy, then on `hawk`

1. Copy the three new hand layers into `campaigns/hawk-claude/pf2e/party/` (Claude's copy, gitignored).
2. Build and run: `go build -o "$SCRATCH/gm-toolkit" ./cmd/gm-toolkit`, then `NO_COLOR=1 "$SCRATCH/gm-toolkit" pf2e sim import fall-of-luna --campaign hawk-claude`. It exits 0, every creature loads, and the report names no `needs a rule from Effort 3`.
3. Check with `show` on `hawk-claude`: `kas` prints `Hero points 1  dying max 6  recovery DC 6`, the five degree shifts and the initiative modifier; `tavish` prints his bombs with persistent and splash parts; `avatar-of-lilith` prints `Weakness holy 15`.
4. Robert runs the same import on `hawk` (E2 Decision 11). Claude commits the rewritten `import.toml` files with this commit.

##### Task 4 — `docs/architecture/pf2e/engine.md`

Written after checking the code, as D-23 sets. Outline and required claims:

- **Purpose** — Efforts 1 and 3 have built the core and the effect core; Efforts 4 and 5 add triggers, limits and spells.
- **Checks and modifiers** (new) — one bonus pool per check: the sheet's number counts as an untyped base; fight-time modifiers, then conditions, add typed values; per type only the highest bonus and the lowest penalty count. The selectors, `against` traits, `per_condition`, and initiative (Perception plus `perception` and `initiative` modifiers).
- **Saves** (new) — DC from a number, the class DC or the spell DC, plus the source's DC modifiers; the degree, then the step's `shift`, then the saver's first matching degree shift by ID; `basic` damage at none, half, full or double.
- **Abilities and steps** (replaces The turn's ability paragraph) — the step kinds; outcome lists per degree; the attack fallback (a critical success with no list runs `success`); an area catches every standing foe; a targeted ability retargets when its target falls; a concealment flat check before an attack and before a targeted opening save; immunity to an effect's traits (`death-effects`) skips ability steps; `unless`.
- **Conditions** (new) — the table and what each row lowers; one value per condition, the highest instance; durations in rounds (counted at the start of the source's turn) or until a turn boundary, with the skipped current end; frightened's decay; `condition_adjust` gain and decay; immunity by name; what waits for Effort 4.
- **Damage** (rewritten) — rolled once per step, then applied; categories; the minimum of 1; temporary HP first; per type immunity, weakness (not on a type that rolled 0) and resistance, with applied weaknesses, lowered resistances and the adaptive resistance; trait-keyed weakness and resistance once per hit; persistent damage and its flat check; splash on the primary target only.
- **Healing** (new) — capped at maximum HP (lowered by drained); a dying PC healed above 0 loses dying and gains 1 wounded.
- **Dying and death** (extended) — the death value is `dying_max` minus doomed; `recovery_dc`; Heroic Recovery at the death value; drained; lethal damage.
- **Events** — the table gains every event added in Commits 2 to 6.
- **Not modelled yet** — Effort 4 (triggers, reactions, limits, action economy, the conditions that change a turn), Effort 5 (spells, counteract, random tables), Effort 7 (position: splash on neighbours, fall damage, reach).

Key Decisions to add, word for word:

- **Abilities are recursive steps with a list of steps per degree of success.** A save, a spell attack or a Strike runs the list for its degree, so a rider, a condition and a nested save share one shape, and the converter fills it the way PF2e text is written. (`F.041.E3` Decision 6.)
- **Condition rules are a Go table of shared parts.** Each row names what the condition lowers, the bonus type and whether it decays; code exists only where a part cannot say the rule. The ruleset as data will lift the table into files. (`F.041.E3` Decision 2.)
- **An area catches every standing foe until the position seam arrives.** One function, `areaTargets`, answers it, and Effort 7 replaces it. (`F.041.E3` Decision 1.)
- **The sheet's number is one untyped base; fight-time bonuses stack among themselves.** A fight-time bonus of a type the build already applies is written as the difference. (`F.041.E3` Decision 10.)

##### Task 5 — `docs/architecture/pf2e/data-files.md`

- **The file format** — the new top-level keys (`hero_points`, `dying_max`, `recovery_dc`), the `[modifier.*]`, `[degree_shift.*]` and `[condition_adjust.*]` tables, a Strike's `range` and riders, `defense.adaptive`, trait keys in `resist` and `weakness`, and a damage part's `category`. Replace "in schema 1 the only step is a Strike" with a table of step kinds and their keys; add one example ability with an area, a basic save and a `failure` list.
- **What the load checks** — the new checks: step kinds and their keys, DC sources, outcome lists only on rolled steps, conditions and their values, durations, condition adjustments, defense traits, the adaptive resistance, `per_condition`, and the build's `critical_specialization` and modifier `attribute`.
- **PC builds** — weapons gain `persistent`, `splash` and riders by degree; `splash` is a modifier selector, and a `splash` modifier may add an `attribute`; critical specialization adds the group's rider to weapons at or above the rank.
- **Import** — the rule-element table gains rows for `AdjustDegreeOfSuccess`, `ActiveEffectLike` on `dying.max` and `dying.recoveryDC`, `FlatModifier` on `initiative`, `FlatModifier` with `damageCategory` splash, `CriticalSpecialization` and `AdjustStrike` materials; NPCs keep trait-keyed weaknesses and Strike parts with a category; property runes are reported for the hand layer.

Key Decisions to add, word for word:

- **Fight-time modifiers and degree shifts are creature tables for PCs and NPCs alike.** They depend on the incoming effect, so a build cannot compute them, and a monster needs them too. (`F.041.E3` Decision 3.)
- **Critical specialization is a rider from a table by weapon group.** The build holds the rank that grants it, and the load adds the group's rider to each weapon at or above it. (`F.041.E3` Decision 15.)

##### Task 6 — `docs/architecture/pf2e/overview.md` and the overview plan

In `overview.md`, the page index's engine row reads: "The roller, checks and modifiers, saves, ability steps, conditions, damage and dying, the fight loop, the decider and the events". The engine stage under Shape adds "and the effect core: saves, conditions, damage kinds and the Down row".

In `F.041-pf2e-encounter-simulator-overview-plan.md`, tick Effort 3's Done box in the Effort Summary.

##### Task 7 — this plan

Tick Phase 4 in the Overview, and log any execution decisions in the Decision Record.

##### Commit message

```
docs(E3): document the effect core and add the party's hand layers

- engine.md: checks and modifiers, saves, ability steps, conditions, damage kinds, healing, dying and death
- data-files.md: the new keys and tables, step kinds, the load checks and the new import rules
- hand layers for Kas, Tavish and Arabesque: one hero point each, and Kas's two runes as critical riders
- the Fall of Luna import layers re-run with the new import rules
- the overview plan ticks Effort 3
```
