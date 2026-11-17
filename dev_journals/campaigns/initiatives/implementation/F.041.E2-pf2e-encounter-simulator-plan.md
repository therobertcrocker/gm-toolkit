# PF2e Encounter Simulator (F.041) — Effort 2 Plan: Foundry Import

## In Plain Words

- **What:** an importer that turns the Foundry exports into the simulator's own creature files. Bosses come across as numbers and Strikes. Player characters come across as their build: boosts, proficiency ranks, armor, weapons and item bonuses. The simulator works out their AC, saves, attacks and the rest when it loads them. A new `gm-toolkit pf2e sim import <fight>` writes the files and prints a report: each creature's stat block, and everything left out with the reason.
- **Why:** the old script's party numbers were typed in by hand and went stale. With the build stored and the math done at load, a level-up is a short layer of new choices, and the numbers follow. Effort 2 ends at gate 1: the computed level-14 numbers for Kas, Tavish and Arabesque match the hand-derived fixture that Robert has checked against Foundry.
- **How it fits:** Effort 2 of 6 in F.041, on branch `feature/pf2e-encounter-simulator`. It adds the `internal/pf2e/sim/foundry` package and a build section to the creature format. Boss abilities stay unconverted. Their rules come in Effort 3 and their description sentences in Effort 4, and the report lists every one.

## Context / Goal

- Discovery: [`F.041-pf2e-encounter-simulator-discovery.md`](../discovery/F.041-pf2e-encounter-simulator-discovery.md), sections Import as a separate step and PC import.
- Overview plan: [`F.041-pf2e-encounter-simulator-overview-plan.md`](./F.041-pf2e-encounter-simulator-overview-plan.md). Its Decisions 1 to 7 bind this plan. Decisions 4 and 5 below amend its Effort 2 and 3 scope (Commit 8 fixes the overview's text).
- Effort 1 plan: [`F.041.E1-pf2e-encounter-simulator-plan.md`](./F.041.E1-pf2e-encounter-simulator-plan.md). Its Decision Record entries 4 and 5 hand findings to this effort.
- Gate 1 fixture: [`F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`](./F.041.E2-pf2e-encounter-simulator-gate1-fixture.md).
- Ends with: `TestGateOneStatParity` passes, and the Fall of Luna import layers are committed in `campaigns/hawk/`.

## Decisions Ratified in Planning

1. **The gate 1 fixture is the hand-derived file, renamed and committed; Robert spot-checks it in Foundry before Phase 4.** Robert's call. The file was saved as `…-fixture-handoff.md`, so git ignored it and the process would delete it. It is now `F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`, committed with this plan. No Foundry instance was run to make it, and it differs from the script's `PARTY[14]` on 13 numbers. Robert checks the three sheets before Phase 4, and Commit 6's table follows the fixture as he corrects it. The gate covers the fixture's high- and medium-confidence numbers. Out of the gate: Kas's and Tavish's empty spell entries, Arabesque's class DC, the Desolation Locket DC, and the Spellsap Grenade. Rejected: the file in testdata (a derivation doc inside a Go package); the tables folded into this plan (loses the derivation trails).
2. **A PC's import layer holds plain build choices; the importer resolves Foundry's rule elements, and load does the core math.** Robert's call; fills in D-4 and D-10. `[build]` holds HP per ancestry and class, the key attribute, base speed, flaws, `weapon_specialization`, and keyed tables for boosts (one list per source, applied in a fixed order), ranks (as words, Foundry's proficiency names), armor, weapons and modifiers (selector, bonus type, value). Predicates, formulas and templates never reach the format. A formula of exactly `@actor.level` becomes `per_level = 1`, so a level-up layer gets Toughness right on its own; every other formula is worked out at the export's level. Rejected: final attribute modifiers instead of boosts (a level-up layer would restate them); raw rule elements evaluated at load (the rule-element engine D-4 rejected).
3. **A PC with a build takes choices only.** Robert's call. If any layer has `[build]`, a layer that sets `perception`, `class_dc`, `spellcasting`, any `[strike.*]`, `speed.land` or `defense.hp`, `ac`, `fortitude`, `reflex` or `will` fails the load with `ErrComputedKey`, naming the file and the key. Fixes go in as build choices or `[build.modifier.*]`. Resistances, immunities, weaknesses, fly speed and abilities stay hand-settable. Rejected: a hand number overrides the computed one (a stale override hides a level-up).
4. **The tag converter moves to Effort 3.** Robert's call; amends the overview's Effort 2 scope. Schema 1 has only Strike steps, and every tag on both bosses names a rule Effort 3 builds, so a parser here would convert nothing. NPC import writes numbers and Strikes; each ability is listed in the report. Effort 3 builds the parser along with each rule it targets.
5. **Effort 2 imports the rule elements the gate numbers need; `AdjustDegreeOfSuccess` and `DamageDice` move to Effort 3.** Robert's call; amends the overview's list. Handled: `FlatModifier` (stat selectors), `Resistance`, `ActiveEffectLike` rank upgrades, `MartialProficiency` for a weapon group (its simple and martial weapons; Decision Record 4), `AdjustModifier` suppressing the armor speed penalty, and `DexterityModifierCap`. Every other rule element is listed with its reason, or counted when the sim never uses its kind (`Note`, `GrantItem` and the like).
6. **Every toggle is imported off.** Robert's call. The fixture is a baseline: shields down, no conditions, toggles off. A toggle (`RollOption` with `toggleable`) is listed in the report as stored on or off. Sticky Bomb, Debilitating Bomb and Diverting Vortex are stored on.
7. **Tests read copies of the exports in `internal/pf2e/sim/foundry/testdata/`.** Robert's call. The five actor exports and the two loose items, about 1.4 MB, are copied as they are. This pins the exact files the fixture was derived from and keeps Effort 1's rule that no test reads campaign data. The copies are made with `cp`, the one place this effort writes files without Write; they are data, not edits.
8. **One new package, `internal/pf2e/sim/foundry`; the stat computation lives in `creature`.** `Load` calls the computation, and `creature` must not import the importer. `foundry` exports only `ImportFight`, `Report`, `Skip` and the reason strings; its tests are white-box (`package foundry`), like `engine`'s. `creature` exports what `foundry` needs: `ImportLayer`, `IsFolderName`, `IsProficiencyKey`, `AttributeNames`, `RankWords` and the two weapon key prefixes.
9. **A creature's folder is its export's file name; PCs go to the party folder, NPCs to the fight folder.** `source/actors/kas.json` writes `pf2e/party/kas/import.toml`; `source/actors/abyssal-luna.json` writes `encounters/fall-of-luna/abyssal-luna/import.toml`. The name must be a slug, and an NPC cannot be named `source` (Decision Record 4). The loose exports in `source/items/` belong to Lilith's Sin Vulnerability, which needs Effort 3's rules; the report lists them as not placed.
10. **The report shows each creature's merged stat block, then the skipped items grouped by reason.** The command loads each written creature (import plus any hand layer) and prints it as `show` does. A creature that fails to load is reported and makes the command exit with an error after the rest are printed.
11. **The Hawk import is run by Robert.** Claude never runs the app against `campaigns/hawk` (standing rule). Commit 7 runs the import on Claude's copy, `hawk-claude`, and checks it. Robert then runs it on `hawk`, and Claude commits the result.
12. **Effort 1's two leftovers stay deferred.** Robert's call. No export has a spaced dice formula (`3d6 + 14`), and only one fight exists, so `ErrCreatureAmbiguous` cannot fire yet. Each waits for the import or fight that hits it.
13. **Weapon specialization comes from the feat's slug, not its rule elements.** `weapon-specialization` gives `normal` and `greater-weapon-specialization` gives `greater`; the rules on those feats are not reported. Its rule elements need a rank-aware `AdjustModifier`, which the build format replaces with the 2/3/4 table. A two-entry table in the importer, not a handler per feat.
14. **Schema stays 1.** The build tables and the computed fields are additions; no layer written so far changes meaning. Effort 4 approves the first draft, so a bump is still free if one is needed.
15. **The importer skips what the format cannot hold yet, with the reason; it never writes a key that would fail the load.** Lilith's `holy` weakness (a trait, not a damage type), persistent and splash damage, property runes, attack effects and `precision` resistance are each listed as needing Effort 3. Immunities are written as they stand, as Effort 1 does for Luna.
16. **Weapons are imported whatever their carry state; a weapon with quantity 0 is not.** A stowed bomb is still a bomb Tavish can draw. Drawing costs are Effort 5's. Consumable quantities are listed as not tracked until Effort 3.
17. **The wiki is specified by outline and required claims, as Effort 1's Decision 10 set.** Commit 8 writes the prose after checking the code; its Key Decisions are given word for word.
18. **Five phases, not four.** Robert's call. The PC import as one phase came to about 1,900 lines, past the phase budget. It splits at Commit 4, which builds green on its own: Phase 3 is the build choices, and Phase 4 the rule elements and the gate test.

## Decision Record — Execution

1. **Commit 7, Task 1: a PC with no resistances is written with an empty `[defense]` table.** The encoder leaves out an empty `[speed]`, but it writes the `[defense]` header for Arabesque, whose `Defense` holds nothing; `Defense` has slice and map fields, so the encoder does not treat it as empty. The file loads, and a bare `[defense]` is not a computed key. Left as is; the effort close can pick a fix (a pointer field, or a custom check) if the noise matters.
2. **Commit 7, Task 1: report reasons that look off on the real import.** No line is wrong enough to block the commit; each is for Effort 3, which sorts these rules anyway.
   - **`AdjustModifier` is always "not supported by the importer"**, even when its target is an Effort 3 rule: Calculated Splash (splash damage), Sticky Bomb twice (persistent damage), Fearsome Brute (damage against a frightened target) and Unburdened Iron's second rule (other speed penalties). Effort 3 should report these with `ReasonEffort3` or apply them as it adds the rules they adjust.
   - **Magnetic Shot's `{item|_id}-attack` selector is "not a combat statistic".** `flatModifierFor` looks the selector up before resolving templates, so an item-scoped attack bonus on ammunition reads as a skill. Ammunition is Effort 3's; resolve the selector template there.
   - **Armor Expertise lists some resistances twice** (piercing, bludgeoning and precision), one line per rule on the feat. Cosmetic.
3. **Commit 8: three wiki changes beyond Task 1 to 3's lists, each to keep a page true to the source.**
   - **`data-files.md`, The file format.** The example quoted Luna's `hand.toml`, which Commit 7 cut to Abyssal Frenzy. It now shows her two layers written as one file (numbers and Strikes from `import.toml`, the ability from `hand.toml`) and adds her `[spellcasting]` table.
   - **`engine.md`, Not modelled yet.** "Bonus-type stacking" stays listed for Effort 3, now marked as in the fight: a PC's build already stacks its bonuses on load.
   - **`overview.md`, Page index.** The data files row's scope adds PC builds and the import.
4. **Effort close: four review fixes, one commit.** The code listings in Phases 1 to 4 stay as executed; this entry and the changed `data-files.md` supersede them.
   - **An advanced weapon skips the `weapon-group-<group>` rank.** Robert's call (option A). Foundry's fighter mastery gives a group's simple and martial weapons one rank and its advanced weapons a lower one, and the format holds only the first, so an advanced hammer in Kas's hands would have read legendary. `Build.strike` now applies the group rank to non-advanced weapons only; the advanced rule stays reported as not supported. Amends Decision 5's `MartialProficiency` line and the Phase 5 rank claim. Rejected: a `weapon-group-advanced-<group>` key (a schema addition no export needs yet); no change, logged for Effort 3 (a wrong number waiting for the first advanced weapon).
   - **A numeric comparison in a predicate reports as not supported.** `rollOptions.numbers` was never filled, so `gte`, `lt` and the like read as false (and true under `not`) with no report line. The field and the comparison code are gone; Fall of Luna's report does not change.
   - **The importer skips what the load would reject.** A PC weapon with dice below 1, a die below 2 or an unknown damage type, an NPC damage part whose dice do not parse or whose type is unknown, and an NPC Strike left with no damage are each reported as not supported. One bad item no longer fails the whole creature. Makes Decision 15 hold.
   - **An export's file name must be a slug, and an NPC cannot be named `source`.** Otherwise the import stops with `errBadFileName`. Narrows Decision 9.
5. **Effort close: review findings declined or deferred.**
   - **Stowed weapons are imported while their rules are off** — declined. Decision 16 imports them on purpose, and Foundry also turns off rules on stowed items.
   - **Deferred, none hit Fall of Luna:** an `or` term in `MartialProficiency` keeps only its last group and drops other terms unreported; `DexterityModifierCap` and the armor-speed-penalty suppression ignore their predicates; worn armor that is not invested is skipped whole instead of losing only its runes; boost lists above the export's level are applied; unknown `FlatModifier` selectors and bonus types, and `ActiveEffectLike` speed paths, are reported as "not a combat statistic". Each waits for the export or effort that hits it.
   - **Effort close review addition.** Efforts 3, 4 and 6 hold up: the Effort 3 rules and every ability appear in the report, and a level-15 layer of choices loads (Kas: HP 246, Reflex +24, Protector's Fury +30 `3d12+13`). Effort 5's missing range on `creature.Strike` is logged in the overview's Findings for the effort plans.

---

## Shared Context

### Gate

Every commit passes `go build ./... && go vet ./... && go test ./...` before it lands. Commit 7 also runs the import by hand (its tasks).

### Before Phase 4

Robert opens the three PC sheets in Foundry and confirms or corrects the fixture's summary tables. Corrections go into the fixture file, and Commit 6, Task 1's table follows the corrected file. Phase 4's execution session checks the fixture's status line first; if Robert has not confirmed it, the session stops before Commit 6.

### Files this effort creates

```
internal/pf2e/sim/creature/
  build.go, build_test.go                                (Commit 1)
internal/pf2e/sim/foundry/
  actor.go, slug.go, actor_test.go, testdata/            (Commit 2)
  report.go, npc.go, write.go, npc_test.go, write_test.go (Commit 3)
  pc.go, pc_test.go                                      (Commit 4)
  formula.go, predicate.go, rules.go, tests              (Commit 5)
  gate_test.go                                           (Commit 6)
cmd/gm-toolkit/pf2e/import.go, import_test.go            (Commit 3)
campaigns/hawk/pf2e/party/{kas,tavish,arabesque}/import.toml
campaigns/hawk/encounters/fall-of-luna/{abyssal-luna,avatar-of-lilith}/import.toml (Commit 7)
```

### Test conventions

- `creature` tests stay black-box (`package creature_test`) and use `writeLayers` from `layers_test.go`.
- `foundry` tests are white-box (`package foundry`): the formula, predicate and import steps are unexported. They read only `testdata/`, and write only to `t.TempDir()`.
- A test that checks the report looks for one `Skip` by item and reason with `hasSkip` (Commit 3, Task 4); it never compares the whole list.

### Running the app

Claude runs the binary only against `hawk-claude` and passes `--campaign hawk-claude`. Build it into the scratchpad: `go build -o "$SCRATCH/gm-toolkit" ./cmd/gm-toolkit`.

### Effort close review addition

`/gestalt:effort-close F.041.E2` also checks that Effort 2's shapes do not box in the later efforts:

- **Effort 3:** the tag converter has a home (`foundry`) and a report to clear; `AdjustDegreeOfSuccess`, `DamageDice`, persistent and splash damage, property runes, `holy` weaknesses and attack effects each appear in the report with `ReasonEffort3`; the import re-runs cleanly at Effort 3's end.
- **Effort 4:** every PC and NPC ability is in the report with `ReasonAbility`, which is the list D-8 works from.
- **Effort 5:** `creature.Weapon` has `range`, but `creature.Strike` does not; tactics will need it on the Strike.
- **Effort 6:** a level-15 variant is a layer of choices: `level = 15`, a `level-15` boost list, `weapon_specialization = "greater"` for Kas, Evasion as `reflex = "master"`.

A finding that needs a change now goes in the review-fix commit; one that can wait goes in the Decision Record for the effort that will need it.

## Out of Scope

- **Every ability's mechanics,** NPC and PC: tags (Effort 3), sentences (Effort 4).
- **Spells** beyond the spellcasting numbers: lists, slots used, focus points (Effort 3).
- **Rule elements outside Decision 5,** and every predicate that does not hold at baseline (toggles, targets, actions, conditions).
- **Persistent and splash damage, property runes, material effects (cold iron), critical specialization, shields** (Effort 3).
- **Skills, initiative modifiers and recovery-check changes** (Toughness's lower recovery DC is Effort 3's, with the Down row).
- **Advanced weapons of a group** (`MartialProficiency` with `item:category:advanced`); none of the three PCs carries one.
- **Unarmed attacks.** The sheets hide the fist (`showBasicUnarmed: false`), so none is imported.
- **A level-15 variant** (Effort 6, scenario variants).

---

## Work Breakdown

**Overview**

- [x] **Phase 1 — PC builds computed on load**
  - Commit 1 — `feat(E2): add PC build layers and compute their stats on load` (7 tasks)
- [x] **Phase 2 — Foundry reader and NPC import**
  - Commit 2 — `feat(E2): read Foundry actor exports` (4 tasks)
  - Commit 3 — `feat(E2): import NPCs and add pf2e sim import` (9 tasks)
- [x] **Phase 3 — PC build choices**
  - Commit 4 — `feat(E2): import PC build choices` (6 tasks)
- [x] **Phase 4 — PC rule elements and gate 1**
  - Commit 5 — `feat(E2): import PC rule elements` (7 tasks)
  - Commit 6 — `test(E2): check gate 1 stat parity` (2 tasks)
- [x] **Phase 5 — Hawk data and the wiki**
  - Commit 7 — `feat(E2): import Fall of Luna and trim Luna's hand layer` (4 tasks)
  - Commit 8 — `docs(E2): document PC builds and the Foundry import` (5 tasks)

### Phase 1 — PC builds computed on load

**In plain words:** the creature format gains a `[build]` section for player characters, and loading one works out its HP, AC, saves, Perception, speed, Strikes, class DC and spell numbers. `show` prints the attributes and the new numbers.

#### Commit 1 — `feat(E2): add PC build layers and compute their stats on load`

##### Task 1 — `internal/pf2e/sim/creature/creature.go`

Rewritten: new fields (`Attributes`, `ClassDC`, `Spellcasting`, `Build`) and encode options, so the importer (Commit 3) can write a `Creature` without zero values. Decoding ignores the options.

`internal/pf2e/sim/creature/creature.go` — full contents:

```go
package creature

import (
	"maps"
	"slices"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const SchemaVersion = 1

type Kind string

const (
	KindPC  Kind = "pc"
	KindNPC Kind = "npc"
)

type Creature struct {
	ID           string             `toml:"-"`
	Layers       []string           `toml:"-"`
	Attributes   map[string]int     `toml:"-"`
	Schema       int                `toml:"schema"`
	Name         string             `toml:"name,omitempty"`
	Kind         Kind               `toml:"kind,omitempty"`
	Level        int                `toml:"level,omitzero"`
	Traits       []string           `toml:"traits,omitempty"`
	Perception   int                `toml:"perception,omitzero"`
	ClassDC      int                `toml:"class_dc,omitzero"`
	Speed        Speed              `toml:"speed,omitempty"`
	Defense      Defense            `toml:"defense,omitempty"`
	Spellcasting *Spellcasting      `toml:"spellcasting,omitempty"`
	Build        *Build             `toml:"build,omitempty"`
	Strikes      map[string]Strike  `toml:"strike,omitempty"`
	Abilities    map[string]Ability `toml:"ability,omitempty"`
}

type Speed struct {
	Land int `toml:"land,omitzero"`
	Fly  int `toml:"fly,omitzero"`
}

type Defense struct {
	HP        int                   `toml:"hp,omitzero"`
	AC        int                   `toml:"ac,omitzero"`
	Fortitude int                   `toml:"fortitude,omitzero"`
	Reflex    int                   `toml:"reflex,omitzero"`
	Will      int                   `toml:"will,omitzero"`
	Immune    []string              `toml:"immune,omitempty"`
	Resist    map[string]Adjustment `toml:"resist,omitempty"`
	Weakness  map[string]Adjustment `toml:"weakness,omitempty"`
}

type Adjustment struct {
	Value  int      `toml:"value"`
	Except []string `toml:"except,omitempty"`
}

type Spellcasting struct {
	Attack  int `toml:"attack,omitzero"`
	DC      int `toml:"dc,omitzero"`
	TopRank int `toml:"top_rank,omitzero"`
}

type Strike struct {
	ID     string       `toml:"-"`
	Name   string       `toml:"name,omitempty"`
	Bonus  int          `toml:"bonus,omitzero"`
	Traits []string     `toml:"traits,omitempty"`
	Damage []DamagePart `toml:"damage,omitempty"`
}

type DamagePart struct {
	Dice string          `toml:"dice"`
	Type string          `toml:"type"`
	Roll dice.Expression `toml:"-"`
}

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

func (strike Strike) Agile() bool {
	return slices.Contains(strike.Traits, "agile")
}

func (creature *Creature) StrikeIDs() []string {
	return slices.Sorted(maps.Keys(creature.Strikes))
}

func (creature *Creature) AbilityIDs() []string {
	return slices.Sorted(maps.Keys(creature.Abilities))
}

var (
	damageTypes   = []string{"acid", "bleed", "bludgeoning", "cold", "electricity", "fire", "force", "mental", "piercing", "poison", "slashing", "sonic", "spirit", "vitality", "void"}
	physicalTypes = []string{"bludgeoning", "piercing", "slashing"}
)

func IsDamageType(name string) bool {
	return slices.Contains(damageTypes, name)
}

func IsPhysical(damageType string) bool {
	return slices.Contains(physicalTypes, damageType)
}
```

##### Task 2 — `internal/pf2e/sim/creature/build.go`

`internal/pf2e/sim/creature/build.go` (new) — full contents:

```go
package creature

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/dice"
)

const (
	WeaponGroupPrefix = "weapon-group-"
	WeaponBasePrefix  = "weapon-base-"
)

var ErrComputedKey = errors.New("key is computed from the build")

var (
	AttributeNames = []string{"str", "dex", "con", "int", "wis", "cha"}
	RankWords      = []string{"untrained", "trained", "expert", "master", "legendary"}

	fixedProficiencies   = []string{"perception", "fortitude", "reflex", "will", "class-dc", "spellcasting", "unarmored", "light", "medium", "heavy", "simple", "martial", "advanced", "unarmed"}
	armorCategories      = []string{"unarmored", "light", "medium", "heavy"}
	weaponCategories     = []string{"simple", "martial", "advanced", "unarmed"}
	modifierSelectors    = []string{"hp", "ac", "fortitude", "reflex", "will", "saving-throw", "perception", "land-speed", "attack", "damage"}
	weaponSelectors      = []string{"attack", "damage"}
	modifierTypes        = []string{"", "item", "status", "circumstance"}
	fixedBoostSources    = []string{"ancestry", "background", "class"}
	saveAttributes       = map[string]string{"fortitude": "con", "reflex": "dex", "will": "wis"}
	specializationDamage = map[string][]int{"normal": {0, 0, 2, 3, 4}, "greater": {0, 0, 4, 6, 8}}
	computedKeys         = [][]string{
		{"perception"}, {"class_dc"}, {"spellcasting"}, {"strike"}, {"speed", "land"},
		{"defense", "hp"}, {"defense", "ac"}, {"defense", "fortitude"}, {"defense", "reflex"}, {"defense", "will"},
	}
)

type Build struct {
	AncestryHP           int                 `toml:"ancestry_hp,omitzero"`
	ClassHP              int                 `toml:"class_hp,omitzero"`
	Key                  string              `toml:"key,omitempty"`
	Speed                int                 `toml:"speed,omitzero"`
	Flaws                []string            `toml:"flaws,omitempty"`
	WeaponSpecialization string              `toml:"weapon_specialization,omitempty"`
	Boosts               map[string][]string `toml:"boosts,omitempty"`
	Proficiency          map[string]string   `toml:"proficiency,omitempty"`
	Armor                *Armor              `toml:"armor,omitempty"`
	Weapons              map[string]Weapon   `toml:"weapon,omitempty"`
	Modifiers            map[string]Modifier `toml:"modifier,omitempty"`
	Spellcasting         *Casting            `toml:"spellcasting,omitempty"`
}

type Armor struct {
	Name         string `toml:"name,omitempty"`
	Category     string `toml:"category"`
	ACBonus      int    `toml:"ac_bonus,omitzero"`
	DexCap       *int   `toml:"dex_cap,omitempty"`
	Potency      int    `toml:"potency,omitzero"`
	Resilient    int    `toml:"resilient,omitzero"`
	SpeedPenalty int    `toml:"speed_penalty,omitzero"`
}

type Weapon struct {
	Name      string   `toml:"name"`
	Category  string   `toml:"category"`
	Group     string   `toml:"group,omitempty"`
	Base      string   `toml:"base,omitempty"`
	Dice      int      `toml:"dice"`
	Die       int      `toml:"die"`
	Type      string   `toml:"type"`
	Potency   int      `toml:"potency,omitzero"`
	Striking  int      `toml:"striking,omitzero"`
	ItemBonus int      `toml:"item_bonus,omitzero"`
	Range     int      `toml:"range,omitzero"`
	Traits    []string `toml:"traits,omitempty"`
}

type Modifier struct {
	Selector     string   `toml:"selector"`
	Type         string   `toml:"type,omitempty"`
	Value        int      `toml:"value,omitzero"`
	PerLevel     int      `toml:"per_level,omitzero"`
	WeaponTraits []string `toml:"weapon_traits,omitempty"`
}

type Casting struct {
	Attribute string `toml:"attribute"`
	TopRank   int    `toml:"top_rank,omitzero"`
}

func IsProficiencyKey(key string) bool {
	if slices.Contains(fixedProficiencies, key) {
		return true
	}
	for _, prefix := range []string{WeaponGroupPrefix, WeaponBasePrefix} {
		if rest, found := strings.CutPrefix(key, prefix); found && rest != "" {
			return true
		}
	}
	return false
}

func (build *Build) Attributes() (map[string]int, error) {
	sources, err := boostSources(build.Boosts)
	if err != nil {
		return nil, err
	}
	modifiers := map[string]int{}
	for _, name := range AttributeNames {
		modifiers[name] = 0
	}
	for _, flaw := range build.Flaws {
		if _, known := modifiers[flaw]; !known {
			return nil, fmt.Errorf("build.flaws: %q is not an attribute", flaw)
		}
		modifiers[flaw]--
	}
	partial := map[string]bool{}
	for _, source := range sources {
		for _, attribute := range build.Boosts[source] {
			if _, known := modifiers[attribute]; !known {
				return nil, fmt.Errorf("build.boosts.%s: %q is not an attribute", source, attribute)
			}
			switch {
			case modifiers[attribute] < 4:
				modifiers[attribute]++
			case partial[attribute]:
				modifiers[attribute]++
				partial[attribute] = false
			default:
				partial[attribute] = true
			}
		}
	}
	return modifiers, nil
}

func boostSources(boosts map[string][]string) ([]string, error) {
	type levelSource struct {
		level  int
		source string
	}
	var ordered []string
	for _, source := range fixedBoostSources {
		if _, found := boosts[source]; found {
			ordered = append(ordered, source)
		}
	}
	var levels []levelSource
	for _, source := range slices.Sorted(maps.Keys(boosts)) {
		if slices.Contains(fixedBoostSources, source) {
			continue
		}
		levelText, isLevel := strings.CutPrefix(source, "level-")
		level, err := strconv.Atoi(levelText)
		if !isLevel || err != nil || level < 1 {
			return nil, fmt.Errorf("build.boosts.%s: want ancestry, background, class or level-<n>", source)
		}
		levels = append(levels, levelSource{level: level, source: source})
	}
	slices.SortStableFunc(levels, func(first, second levelSource) int { return first.level - second.level })
	for _, entry := range levels {
		ordered = append(ordered, entry.source)
	}
	return ordered, nil
}

func checkComputedKeys(names []string, layers []map[string]any) error {
	hasBuild := slices.ContainsFunc(layers, func(layer map[string]any) bool {
		_, found := layer["build"]
		return found
	})
	if !hasBuild {
		return nil
	}
	var problems []string
	for index, layer := range layers {
		for _, path := range computedKeys {
			if hasPath(layer, path) {
				problems = append(problems, fmt.Sprintf("%s sets %s", names[index], strings.Join(path, ".")))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %s; change the build instead", ErrComputedKey, strings.Join(problems, ", "))
	}
	return nil
}

func hasPath(tree map[string]any, path []string) bool {
	for index, key := range path {
		value, found := tree[key]
		if !found {
			return false
		}
		if index == len(path)-1 {
			return true
		}
		next, isTable := value.(map[string]any)
		if !isTable {
			return false
		}
		tree = next
	}
	return false
}

func (creature *Creature) checkBuild() []string {
	build := creature.Build
	var problems []string
	report := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if creature.Kind != KindPC {
		report("build: kind %q: only a pc has a build", creature.Kind)
	}
	if creature.Level < 1 {
		report("build: level must be at least 1")
	}
	if !slices.Contains(AttributeNames, build.Key) {
		report("build.key %q: not an attribute", build.Key)
	}
	if build.AncestryHP < 0 {
		report("build.ancestry_hp must not be negative")
	}
	if build.ClassHP < 1 {
		report("build.class_hp must be at least 1")
	}
	if build.Speed < 0 {
		report("build.speed must not be negative")
	}
	if _, known := specializationDamage[build.WeaponSpecialization]; build.WeaponSpecialization != "" && !known {
		report("build.weapon_specialization %q: want normal or greater", build.WeaponSpecialization)
	}
	for _, key := range slices.Sorted(maps.Keys(build.Proficiency)) {
		if !IsProficiencyKey(key) {
			report("build.proficiency.%s: unknown proficiency", key)
		}
		if rank := build.Proficiency[key]; !slices.Contains(RankWords, rank) {
			report("build.proficiency.%s: rank %q: want %s", key, rank, strings.Join(RankWords, ", "))
		}
	}
	if armor := build.Armor; armor != nil {
		if !slices.Contains(armorCategories, armor.Category) {
			report("build.armor.category %q: want %s", armor.Category, strings.Join(armorCategories, ", "))
		}
		if armor.ACBonus < 0 || armor.Potency < 0 || armor.Resilient < 0 || (armor.DexCap != nil && *armor.DexCap < 0) {
			report("build.armor: ac_bonus, dex_cap, potency and resilient must not be negative")
		}
		if armor.SpeedPenalty > 0 {
			report("build.armor.speed_penalty must not be positive")
		}
	}
	for _, id := range slices.Sorted(maps.Keys(build.Weapons)) {
		weapon := build.Weapons[id]
		if weapon.Name == "" {
			report("build.weapon.%s: name is empty", id)
		}
		if !slices.Contains(weaponCategories, weapon.Category) {
			report("build.weapon.%s: category %q: want %s", id, weapon.Category, strings.Join(weaponCategories, ", "))
		}
		if weapon.Dice < 1 || weapon.Die < 2 {
			report("build.weapon.%s: dice must be at least 1 and die at least 2", id)
		}
		if !IsDamageType(weapon.Type) {
			report("build.weapon.%s: unknown damage type %q", id, weapon.Type)
		}
		if weapon.Potency < 0 || weapon.Striking < 0 || weapon.ItemBonus < 0 || weapon.Range < 0 {
			report("build.weapon.%s: potency, striking, item_bonus and range must not be negative", id)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(build.Modifiers)) {
		modifier := build.Modifiers[id]
		if !slices.Contains(modifierSelectors, modifier.Selector) {
			report("build.modifier.%s: selector %q: want %s", id, modifier.Selector, strings.Join(modifierSelectors, ", "))
		}
		if !slices.Contains(modifierTypes, modifier.Type) {
			report("build.modifier.%s: type %q: want item, status, circumstance or none", id, modifier.Type)
		}
		if len(modifier.WeaponTraits) > 0 && !slices.Contains(weaponSelectors, modifier.Selector) {
			report("build.modifier.%s: weapon_traits needs selector attack or damage", id)
		}
	}
	if casting := build.Spellcasting; casting != nil {
		if !slices.Contains(AttributeNames, casting.Attribute) {
			report("build.spellcasting.attribute %q: not an attribute", casting.Attribute)
		}
		if casting.TopRank < 0 || casting.TopRank > 10 {
			report("build.spellcasting.top_rank %d: want 0 to 10", casting.TopRank)
		}
	}
	return problems
}

func (creature *Creature) computeBuild() error {
	if problems := creature.checkBuild(); len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
	}
	build := creature.Build
	attributes, err := build.Attributes()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	level := creature.Level
	creature.Attributes = attributes

	armor := Armor{Category: "unarmored"}
	if build.Armor != nil {
		armor = *build.Armor
	}
	dexterity := attributes["dex"]
	if armor.DexCap != nil {
		dexterity = min(dexterity, *armor.DexCap)
	}
	saveTotal := func(save string) int {
		return build.proficiency(level, save) + attributes[saveAttributes[save]] + build.total(level, armor.Resilient, nil, save, "saving-throw")
	}

	creature.Defense.HP = build.AncestryHP + level*(build.ClassHP+attributes["con"]) + build.total(level, 0, nil, "hp")
	creature.Defense.AC = 10 + dexterity + build.proficiency(level, armor.Category) + build.total(level, armor.ACBonus+armor.Potency, nil, "ac")
	creature.Defense.Fortitude = saveTotal("fortitude")
	creature.Defense.Reflex = saveTotal("reflex")
	creature.Defense.Will = saveTotal("will")
	creature.Perception = build.proficiency(level, "perception") + attributes["wis"] + build.total(level, 0, nil, "perception")
	creature.Speed.Land = build.Speed + armor.SpeedPenalty + build.total(level, 0, nil, "land-speed")
	creature.ClassDC = 10 + build.proficiency(level, "class-dc") + attributes[build.Key]
	if casting := build.Spellcasting; casting != nil {
		attack := build.proficiency(level, "spellcasting") + attributes[casting.Attribute]
		creature.Spellcasting = &Spellcasting{Attack: attack, DC: 10 + attack, TopRank: casting.TopRank}
	}
	creature.Strikes = map[string]Strike{}
	for id, weapon := range build.Weapons {
		creature.Strikes[id] = build.strike(level, attributes, weapon)
	}
	return nil
}

func (build *Build) rank(key string) int {
	return max(slices.Index(RankWords, build.Proficiency[key]), 0)
}

func (build *Build) proficiency(level int, key string) int {
	return proficiencyBonus(level, build.rank(key))
}

func proficiencyBonus(level, rank int) int {
	if rank == 0 {
		return 0
	}
	return level + 2*rank
}

func (build *Build) total(level, itemBonus int, weaponTraits []string, selectors ...string) int {
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
}

func (build *Build) strike(level int, attributes map[string]int, weapon Weapon) Strike {
	rank := max(build.rank(weapon.Category), build.rank(WeaponGroupPrefix+weapon.Group), build.rank(WeaponBasePrefix+weapon.Base))
	attack := proficiencyBonus(level, rank) + attackAttribute(attributes, weapon) + build.total(level, max(weapon.Potency, weapon.ItemBonus), weapon.Traits, "attack")
	modifier := damageAttribute(attributes, weapon) + build.specialization(rank) + build.total(level, 0, weapon.Traits, "damage")
	expression := dice.Expression{Count: weapon.Dice + weapon.Striking, Sides: weapon.Die, Modifier: modifier}
	return Strike{
		Name:   weapon.Name,
		Bonus:  attack,
		Traits: slices.Clone(weapon.Traits),
		Damage: []DamagePart{{Dice: expression.String(), Type: weapon.Type}},
	}
}

func (build *Build) specialization(rank int) int {
	table, found := specializationDamage[build.WeaponSpecialization]
	if !found {
		return 0
	}
	return table[rank]
}

func attackAttribute(attributes map[string]int, weapon Weapon) int {
	switch {
	case weapon.Range > 0:
		return attributes["dex"]
	case slices.Contains(weapon.Traits, "finesse"):
		return max(attributes["str"], attributes["dex"])
	default:
		return attributes["str"]
	}
}

func damageAttribute(attributes map[string]int, weapon Weapon) int {
	strength := attributes["str"]
	switch {
	case weapon.Range == 0 || isThrown(weapon.Traits):
		return strength
	case slices.Contains(weapon.Traits, "propulsive") && strength > 0:
		return strength / 2
	case slices.Contains(weapon.Traits, "propulsive"):
		return strength
	default:
		return 0
	}
}

func isThrown(traits []string) bool {
	return slices.ContainsFunc(traits, func(trait string) bool {
		return trait == "thrown" || strings.HasPrefix(trait, "thrown-")
	})
}

func containsAll(have, want []string) bool {
	for _, entry := range want {
		if !slices.Contains(have, entry) {
			return false
		}
	}
	return true
}

type bonusPool struct {
	untyped   int
	bonuses   map[string]int
	penalties map[string]int
}

func newBonusPool() *bonusPool {
	return &bonusPool{bonuses: map[string]int{}, penalties: map[string]int{}}
}

func (pool *bonusPool) add(kind string, value int) {
	switch {
	case kind == "":
		pool.untyped += value
	case value >= 0:
		pool.bonuses[kind] = max(pool.bonuses[kind], value)
	default:
		pool.penalties[kind] = min(pool.penalties[kind], value)
	}
}

func (pool *bonusPool) total() int {
	total := pool.untyped
	for _, value := range pool.bonuses {
		total += value
	}
	for _, value := range pool.penalties {
		total += value
	}
	return total
}
```

##### Task 3 — `internal/pf2e/sim/creature/layers.go`

Rewritten: `ImportLayer` and `IsFolderName` are exported for the importer and the `import` command; every layer is read before any merge, so the computed-key check sees each file as written (the merge reuses the first layer's tables); a build is computed before `prepare`.

`internal/pf2e/sim/creature/layers.go` — full contents:

```go
package creature

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	ImportLayer = "import.toml"
	handLayer   = "hand.toml"
)

var (
	ErrNoLayers   = errors.New("creature has neither an import nor a hand layer")
	ErrInvalid    = errors.New("invalid creature")
	ErrBadVariant = errors.New("bad variant name")
)

func Load(directory string, variants ...string) (*Creature, error) {
	var names []string
	for _, name := range []string{ImportLayer, handLayer} {
		_, err := os.Stat(filepath.Join(directory, name))
		if err == nil {
			names = append(names, name)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("load creature %s: %w", directory, ErrNoLayers)
	}
	for _, variant := range variants {
		name := variant + ".toml"
		if !IsFolderName(variant) || name == ImportLayer || name == handLayer {
			return nil, fmt.Errorf("load creature %s: %w: %q", directory, ErrBadVariant, variant)
		}
		names = append(names, name)
	}

	layers := make([]map[string]any, len(names))
	for index, name := range names {
		layer, err := readLayer(filepath.Join(directory, name))
		if err != nil {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
		layers[index] = layer
	}
	if err := checkComputedKeys(names, layers); err != nil {
		return nil, fmt.Errorf("load creature %s: %w", directory, err)
	}
	merged := map[string]any{}
	for _, layer := range layers {
		mergeInto(merged, layer)
	}

	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(merged); err != nil {
		return nil, fmt.Errorf("load creature %s: encode merged layers: %w", directory, err)
	}
	var creature Creature
	if _, err := toml.Decode(buffer.String(), &creature); err != nil {
		return nil, fmt.Errorf("load creature %s: decode merged layers: %w", directory, err)
	}
	creature.ID = filepath.Base(directory)
	creature.Layers = names
	if creature.Build != nil {
		if err := creature.computeBuild(); err != nil {
			return nil, fmt.Errorf("load creature %s: %w", directory, err)
		}
	}
	if err := creature.prepare(); err != nil {
		return nil, fmt.Errorf("load creature %s: %w", directory, err)
	}
	return &creature, nil
}

func IsFolderName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

func readLayer(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	var typed Creature
	metadata, err := toml.Decode(string(content), &typed)
	if err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("layer %s: unknown keys: %s", path, joinKeys(undecoded))
	}
	// deliberate: a key with a capital letter is an error, not accepted, because the decoder matches fields without case but the merge does not, so `HP` and `hp` would both survive the merge and either could win.
	var capitalised []toml.Key
	for _, key := range metadata.Keys() {
		if text := key.String(); text != strings.ToLower(text) {
			capitalised = append(capitalised, key)
		}
	}
	if len(capitalised) > 0 {
		return nil, fmt.Errorf("layer %s: keys must be lower case: %s", path, joinKeys(capitalised))
	}
	if typed.Schema != SchemaVersion {
		return nil, fmt.Errorf("layer %s: schema %d, want %d", path, typed.Schema, SchemaVersion)
	}
	var layer map[string]any
	if _, err := toml.Decode(string(content), &layer); err != nil {
		return nil, fmt.Errorf("layer %s: %w", path, err)
	}
	return layer, nil
}

func joinKeys(keys []toml.Key) string {
	texts := make([]string, len(keys))
	for index, key := range keys {
		texts[index] = key.String()
	}
	return strings.Join(texts, ", ")
}

func mergeInto(base, overlay map[string]any) {
	for key, value := range overlay {
		overlayTable, overlayIsTable := value.(map[string]any)
		baseTable, baseIsTable := base[key].(map[string]any)
		if overlayIsTable && baseIsTable {
			mergeInto(baseTable, overlayTable)
			continue
		}
		base[key] = value
	}
}
```

In `internal/pf2e/sim/creature/locate.go`, `Locate` uses the renamed check.

**Find:**
```go
	if !isFolderName(id) {
```
**Replace with:**
```go
	if !IsFolderName(id) {
```

##### Task 4 — `internal/pf2e/sim/creature/build_test.go`

`fighterBuild` is Kas's build as Commit 4 and 5's importer will write it; its numbers are the fixture's. `scoutBuild` covers the weapon attributes and bonus stacking.

`internal/pf2e/sim/creature/build_test.go` (new) — full contents:

```go
package creature_test

import (
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

const fighterBuild = `schema = 1
name  = "Kas"
kind  = "pc"
level = 14

[build]
ancestry_hp = 6
class_hp    = 10
key         = "str"
speed       = 25
weapon_specialization = "normal"

[build.boosts]
ancestry   = ["str", "cha"]
background = ["con", "str"]
class      = ["str"]
level-1    = ["cha", "con", "str", "wis"]
level-5    = ["str", "con", "cha", "wis"]
level-10   = ["str", "con", "cha", "wis"]

[build.proficiency]
perception = "master"
fortitude  = "master"
reflex     = "expert"
will       = "expert"
class-dc   = "expert"
heavy      = "expert"
simple     = "master"
martial    = "master"
unarmed    = "master"
weapon-group-hammer = "legendary"

[build.armor]
name     = "Full Plate"
category = "heavy"
ac_bonus = 6
dex_cap  = 0
potency  = 2
resilient = 2

[build.weapon.protectors-fury]
name     = "Protector's Fury"
category = "martial"
group    = "hammer"
base     = "maul"
dice     = 1
die      = 12
type     = "bludgeoning"
potency  = 2
striking = 2
traits   = ["trip", "versatile-p"]

[build.modifier.toughness]
selector  = "hp"
per_level = 1

[build.modifier.mountains-stoutness]
selector  = "hp"
per_level = 1

[build.modifier.fleet]
selector = "land-speed"
value    = 5

[build.modifier.amulet-of-kinships-strength]
selector = "will"
type     = "item"
value    = 2

[defense.resist.slashing]
value = 4
`

const scoutBuild = `schema = 1
name  = "Scout"
kind  = "pc"
level = 5

[build]
ancestry_hp = 8
class_hp    = 8
key         = "dex"
speed       = 25
weapon_specialization = "normal"

[build.boosts]
ancestry   = ["dex"]
background = ["dex", "str"]
class      = ["dex"]
level-1    = ["dex", "str", "con", "wis"]
level-5    = ["dex", "str", "con", "wis"]

[build.proficiency]
unarmored = "trained"
simple    = "expert"
martial   = "trained"
weapon-base-alchemical-bomb = "expert"

[build.weapon.needle]
name     = "Needle"
category = "simple"
dice     = 1
die      = 6
type     = "piercing"
traits   = ["finesse"]

[build.weapon.longbow]
name     = "Longbow"
category = "martial"
dice     = 1
die      = 8
type     = "piercing"
range    = 100

[build.weapon.sling]
name     = "Sling"
category = "simple"
dice     = 1
die      = 6
type     = "bludgeoning"
range    = 50
traits   = ["propulsive"]

[build.weapon.javelin]
name     = "Javelin"
category = "simple"
dice     = 1
die      = 6
type     = "piercing"
range    = 30
traits   = ["thrown"]

[build.weapon.fire-bomb]
name       = "Fire Bomb"
category   = "martial"
base       = "alchemical-bomb"
dice       = 2
die        = 8
type       = "fire"
item_bonus = 2
range      = 20
traits     = ["alchemical", "bomb"]

[build.weapon.weak-bomb]
name     = "Weak Bomb"
category = "martial"
base     = "alchemical-bomb"
dice     = 2
die      = 8
type     = "fire"
range    = 20
traits   = ["alchemical", "bomb"]

[build.modifier.goggles]
selector      = "attack"
type          = "item"
value         = 1
weapon_traits = ["bomb"]

[build.modifier.bracers]
selector = "ac"
type     = "item"
value    = 1

[build.modifier.bands]
selector = "ac"
type     = "item"
value    = 2

[build.modifier.cover]
selector = "ac"
type     = "circumstance"
value    = 1

[build.modifier.frightened]
selector = "ac"
type     = "status"
value    = -1

[build.modifier.talisman]
selector = "ac"
value    = 1
`

func TestBuildComputesFighterStats(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantAttributes := map[string]int{"str": 5, "dex": 0, "con": 4, "int": 0, "wis": 3, "cha": 4}
	if !maps.Equal(loaded.Attributes, wantAttributes) {
		t.Errorf("Attributes = %v, want %v", loaded.Attributes, wantAttributes)
	}
	for _, check := range []struct {
		name      string
		got, want int
	}{
		{"hp", loaded.Defense.HP, 230},
		{"ac", loaded.Defense.AC, 36},
		{"fortitude", loaded.Defense.Fortitude, 26},
		{"reflex", loaded.Defense.Reflex, 20},
		{"will", loaded.Defense.Will, 23},
		{"perception", loaded.Perception, 23},
		{"land speed", loaded.Speed.Land, 30},
		{"class dc", loaded.ClassDC, 33},
		{"slashing resistance", loaded.Defense.Resist["slashing"].Value, 4},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	fury, found := loaded.Strikes["protectors-fury"]
	if !found {
		t.Fatalf("no protectors-fury Strike; have %v", loaded.StrikeIDs())
	}
	if fury.ID != "protectors-fury" || fury.Bonus != 29 {
		t.Errorf("Strike = %q %+d, want protectors-fury +29", fury.ID, fury.Bonus)
	}
	if got := fury.Damage[0].Roll.String(); got != "3d12+9" || fury.Damage[0].Type != "bludgeoning" {
		t.Errorf("damage = %s %s, want 3d12+9 bludgeoning", got, fury.Damage[0].Type)
	}
	if loaded.Spellcasting != nil {
		t.Errorf("Spellcasting = %+v, want none without a spellcasting choice", *loaded.Spellcasting)
	}
}

func TestBuildLevelUpLayerFollowsThrough(t *testing.T) {
	directory := writeLayers(t, map[string]string{
		"import.toml":   fighterBuild,
		"level-15.toml": "schema = 1\nlevel = 15\n",
	})
	loaded, err := creature.Load(directory, "level-15")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 246 {
		t.Errorf("hp = %d, want 246 (Toughness and Mountain's Stoutness grow per level)", loaded.Defense.HP)
	}
	if bonus := loaded.Strikes["protectors-fury"].Bonus; bonus != 30 {
		t.Errorf("attack = %+d, want +30", bonus)
	}
}

func TestBuildWeaponAttributesAndBonusStacking(t *testing.T) {
	loaded, err := creature.Load(writeLayers(t, map[string]string{"import.toml": scoutBuild}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.AC != 24 {
		t.Errorf("ac = %d, want 24 (item 2 of 1 and 2, circumstance +1, status -1, untyped +1)", loaded.Defense.AC)
	}
	if loaded.Defense.HP != 58 {
		t.Errorf("hp = %d, want 58", loaded.Defense.HP)
	}
	for _, want := range []struct {
		id     string
		bonus  int
		damage string
	}{
		{"needle", 13, "1d6+5"},
		{"longbow", 11, "1d8"},
		{"sling", 13, "1d6+3"},
		{"javelin", 13, "1d6+5"},
		{"fire-bomb", 15, "2d8+2"},
		{"weak-bomb", 14, "2d8+2"},
	} {
		strike := loaded.Strikes[want.id]
		if strike.Bonus != want.bonus || strike.Damage[0].Roll.String() != want.damage {
			t.Errorf("%s = %+d %s, want %+d %s", want.id, strike.Bonus, strike.Damage[0].Roll, want.bonus, want.damage)
		}
	}
}

func TestBuildAttributesTrackPartialBoosts(t *testing.T) {
	build := &creature.Build{
		Flaws: []string{"con"},
		Boosts: map[string][]string{
			"ancestry":   {"dex", "int", "cha"},
			"background": {"int", "cha"},
			"class":      {"cha"},
			"level-1":    {"int", "cha", "dex", "con"},
			"level-5":    {"cha", "int", "dex", "wis"},
			"level-10":   {"con", "dex", "int", "cha"},
		},
	}
	got, err := build.Attributes()
	if err != nil {
		t.Fatalf("Attributes: %v", err)
	}
	want := map[string]int{"str": 0, "dex": 4, "con": 1, "int": 4, "wis": 1, "cha": 5}
	if !maps.Equal(got, want) {
		t.Errorf("Attributes = %v, want %v", got, want)
	}

	build.Boosts["level-five"] = []string{"str"}
	if _, err := build.Attributes(); err == nil || !strings.Contains(err.Error(), "build.boosts.level-five") {
		t.Errorf("err = %v, want one naming build.boosts.level-five", err)
	}
}

func TestBuildRejectsComputedKeys(t *testing.T) {
	_, err := creature.Load(writeLayers(t, map[string]string{
		"import.toml": fighterBuild,
		"hand.toml":   "schema = 1\n\n[defense]\nac = 40\n",
	}))
	if !errors.Is(err, creature.ErrComputedKey) {
		t.Fatalf("err = %v, want ErrComputedKey", err)
	}
	if !strings.Contains(err.Error(), "hand.toml sets defense.ac") {
		t.Errorf("err = %v, want it to name hand.toml and defense.ac", err)
	}
}

func TestBuildRejectsBadChoices(t *testing.T) {
	hand := "schema = 1\n\n[build.proficiency]\nperception = \"grandmaster\"\n\n[build.modifier.lucky]\nselector = \"initiative\"\nvalue = 1\n"
	_, err := creature.Load(writeLayers(t, map[string]string{"import.toml": fighterBuild, "hand.toml": hand}))
	if !errors.Is(err, creature.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	for _, want := range []string{`build.proficiency.perception: rank "grandmaster"`, `build.modifier.lucky: selector "initiative"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to contain %q", err, want)
		}
	}
}
```

##### Task 5 — `cmd/gm-toolkit/pf2e/show.go`

In `printCreature`, print the attributes after Perception, and class DC and spell numbers after speed.

**Find:**
```go
	defense := shown.Defense
	fmt.Fprintf(out, "Perception %+d\n", shown.Perception)
```
**Replace with:**
```go
	defense := shown.Defense
	fmt.Fprintf(out, "Perception %+d\n", shown.Perception)
	if len(shown.Attributes) > 0 {
		texts := make([]string, len(creature.AttributeNames))
		for index, name := range creature.AttributeNames {
			texts[index] = fmt.Sprintf("%s %+d", strings.ToUpper(name[:1])+name[1:], shown.Attributes[name])
		}
		fmt.Fprintln(out, strings.Join(texts, "  "))
	}
```

**Find:**
```go
	fmt.Fprintf(out, "Speed %s\n", speedText(shown.Speed))
```
**Replace with:**
```go
	fmt.Fprintf(out, "Speed %s\n", speedText(shown.Speed))
	if shown.ClassDC > 0 {
		fmt.Fprintf(out, "Class DC %d\n", shown.ClassDC)
	}
	if casting := shown.Spellcasting; casting != nil {
		fmt.Fprintf(out, "Spell attack %+d  DC %d", casting.Attack, casting.DC)
		if casting.TopRank > 0 {
			fmt.Fprintf(out, "  top rank %d", casting.TopRank)
		}
		fmt.Fprintln(out)
	}
```

##### Task 6 — `cmd/gm-toolkit/pf2e/show_test.go`

Append after `TestPrintCreature`:

```go
func TestPrintCreatureShowsComputedNumbers(t *testing.T) {
	shown := &creature.Creature{
		Name:         "Arabesque",
		Kind:         creature.KindPC,
		Level:        14,
		Perception:   19,
		Attributes:   map[string]int{"str": 0, "dex": 4, "con": 1, "int": 4, "wis": 1, "cha": 5},
		ClassDC:      31,
		Spellcasting: &creature.Spellcasting{Attack: 23, DC: 33, TopRank: 7},
		Speed:        creature.Speed{Land: 30},
		Defense:      creature.Defense{HP: 104, AC: 34, Fortitude: 21, Reflex: 24, Will: 21},
		Layers:       []string{"import.toml"},
	}
	var out bytes.Buffer
	printCreature(&out, shown, false)
	for _, want := range []string{
		"Str +0  Dex +4  Con +1  Int +4  Wis +1  Cha +5",
		"Class DC 31",
		"Spell attack +23  DC 33  top rank 7",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}
```

##### Task 7 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-plan.md`

In the Work Breakdown Overview, tick Phase 1.

**Find:**
```
- [ ] **Phase 1 — PC builds computed on load**
```
**Replace with:**
```
- [x] **Phase 1 — PC builds computed on load**
```

##### Commit message

```
feat(E2): add PC build layers and compute their stats on load

- creature: a [build] section of plain choices (boosts, ranks, armor,
  weapons, modifiers, spellcasting) for player characters
- Load computes HP, AC, saves, Perception, speed, Strikes, class DC and
  spell numbers from it; a layer that sets a computed key fails with
  ErrComputedKey
- show prints attributes, class DC and spell numbers
- encode options on the creature types so a layer can be written back
```

### Phase 2 — Foundry reader and NPC import

**In plain words:** the new `foundry` package reads Foundry actor exports, turns a boss into an import layer of numbers and Strikes, and `gm-toolkit pf2e sim import <fight>` writes those layers and prints the report. Player characters are listed but not imported until Phase 3.

#### Commit 2 — `feat(E2): read Foundry actor exports`

##### Task 1 — `internal/pf2e/sim/foundry/testdata/`

Copy the exports as they are (Decision 7). Run from the repository root:

```bash
mkdir -p internal/pf2e/sim/foundry/testdata/actors internal/pf2e/sim/foundry/testdata/items
cp campaigns/hawk/encounters/fall-of-luna/source/actors/abyssal-luna.json \
   campaigns/hawk/encounters/fall-of-luna/source/actors/avatar-of-lilith.json \
   internal/pf2e/sim/foundry/testdata/actors/
cp campaigns/hawk/encounters/fall-of-luna/source/items/effect-faithful-strike.json \
   campaigns/hawk/encounters/fall-of-luna/source/items/swear-fidelity.json \
   internal/pf2e/sim/foundry/testdata/items/
```

##### Task 2 — `internal/pf2e/sim/foundry/actor.go`

Item `system` blocks share field names with different shapes across item types (`range`, `damage`), so an item keeps its `system` raw and decodes the common fields once; each importer step decodes the type it needs. `damageRolls` keeps the export's order, which a Go map would lose.

`internal/pf2e/sim/foundry/actor.go` (new) — full contents:

```go
package foundry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var ErrNotAnActor = errors.New("not an actor export")

type valueOf[T any] struct {
	Value T `json:"value"`
}

type actor struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Raw   json.RawMessage `json:"system"`
	Items []item          `json:"items"`
}

type item struct {
	ID     string          `json:"_id"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Raw    json.RawMessage `json:"system"`
	System itemCommon      `json:"-"`
}

type itemCommon struct {
	Slug       string           `json:"slug"`
	Traits     itemTraits       `json:"traits"`
	Rules      []map[string]any `json:"rules"`
	Equipped   equipped         `json:"equipped"`
	Quantity   int              `json:"quantity"`
	ActionType valueOf[string]  `json:"actionType"`
}

type itemTraits struct {
	Value  []string `json:"value"`
	Rarity string   `json:"rarity"`
}

type equipped struct {
	CarryType string `json:"carryType"`
	Invested  *bool  `json:"invested"`
}

type damageRoll struct {
	Damage     string `json:"damage"`
	DamageType string `json:"damageType"`
	Category   string `json:"category"`
}

type damageRolls []damageRoll

func (rolls *damageRolls) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, isDelimiter := opening.(json.Delim); !isDelimiter || delimiter != '{' {
		return fmt.Errorf("damageRolls: want an object, got %v", opening)
	}
	for decoder.More() {
		if _, err := decoder.Token(); err != nil {
			return err
		}
		var roll damageRoll
		if err := decoder.Decode(&roll); err != nil {
			return err
		}
		*rolls = append(*rolls, roll)
	}
	_, err = decoder.Token()
	return err
}

func readActor(path string) (*actor, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read actor %s: %w", path, err)
	}
	var source actor
	if err := json.Unmarshal(content, &source); err != nil {
		return nil, fmt.Errorf("read actor %s: %w", path, err)
	}
	if source.Type != "character" && source.Type != "npc" {
		return nil, fmt.Errorf("read actor %s: %w: type %q", path, ErrNotAnActor, source.Type)
	}
	for index := range source.Items {
		if err := source.Items[index].decodeCommon(); err != nil {
			return nil, fmt.Errorf("read actor %s: %w", path, err)
		}
	}
	return &source, nil
}

func readItem(path string) (*item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	var loose item
	if err := json.Unmarshal(content, &loose); err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	if err := loose.decodeCommon(); err != nil {
		return nil, fmt.Errorf("read item %s: %w", path, err)
	}
	return &loose, nil
}

func (source *actor) decode(target any) error {
	if err := json.Unmarshal(source.Raw, target); err != nil {
		return fmt.Errorf("actor %s: %w", source.Name, err)
	}
	return nil
}

func (current *item) decodeCommon() error {
	return current.decode(&current.System)
}

func (current *item) decode(target any) error {
	if err := json.Unmarshal(current.Raw, target); err != nil {
		return fmt.Errorf("item %s (%s): %w", current.Name, current.Type, err)
	}
	return nil
}
```

##### Task 3 — `internal/pf2e/sim/foundry/slug.go`

`internal/pf2e/sim/foundry/slug.go` (new) — full contents:

```go
package foundry

import (
	"fmt"
	"strings"
)

func slugify(name string) string {
	lowered := strings.NewReplacer("'", "", "’", "").Replace(strings.ToLower(name))
	var builder strings.Builder
	pendingDash := false
	for _, character := range lowered {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			if pendingDash && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(character)
			pendingDash = false
			continue
		}
		pendingDash = true
	}
	return builder.String()
}

func uniqueID(name string, taken map[string]bool) string {
	base := slugify(name)
	if base == "" {
		base = "item"
	}
	id := base
	for suffix := 2; taken[id]; suffix++ {
		id = fmt.Sprintf("%s-%d", base, suffix)
	}
	taken[id] = true
	return id
}
```

##### Task 4 — `internal/pf2e/sim/foundry/actor_test.go`

`internal/pf2e/sim/foundry/actor_test.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestReadActorKeepsDamageRollOrder(t *testing.T) {
	luna, err := readActor(filepath.Join("testdata", "actors", "abyssal-luna.json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	if luna.Name != "Abyssal Luna" || luna.Type != "npc" {
		t.Errorf("actor = %q %q, want Abyssal Luna npc", luna.Name, luna.Type)
	}
	index := slices.IndexFunc(luna.Items, func(current item) bool { return current.Name == "Slash" })
	if index < 0 {
		t.Fatal("no Slash item")
	}
	var system struct {
		DamageRolls damageRolls `json:"damageRolls"`
	}
	if err := luna.Items[index].decode(&system); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := damageRolls{{Damage: "3d6+14", DamageType: "slashing"}, {Damage: "1d8", DamageType: "mental"}}
	if !slices.Equal(system.DamageRolls, want) {
		t.Errorf("damage rolls = %+v, want %+v", system.DamageRolls, want)
	}
	if traits := luna.Items[index].System.Traits.Value; !slices.Contains(traits, "agile") {
		t.Errorf("Slash traits = %v, want agile among them", traits)
	}
}

func TestReadActorRejectsAnItemExport(t *testing.T) {
	_, err := readActor(filepath.Join("testdata", "items", "swear-fidelity.json"))
	if !errors.Is(err, ErrNotAnActor) {
		t.Errorf("err = %v, want ErrNotAnActor", err)
	}
}

func TestReadItem(t *testing.T) {
	loose, err := readItem(filepath.Join("testdata", "items", "swear-fidelity.json"))
	if err != nil {
		t.Fatalf("readItem: %v", err)
	}
	if loose.Name != "Swear Fidelity" || loose.Type != "action" {
		t.Errorf("item = %q %q, want Swear Fidelity action", loose.Name, loose.Type)
	}
}

func TestSlugify(t *testing.T) {
	for _, test := range []struct{ name, want string }{
		{"Protector's Fury", "protectors-fury"},
		{"Alchemist's Fire (Greater)", "alchemists-fire-greater"},
		{"Artisan's Mark™", "artisans-mark"},
		{"Bands of Force (Greater)", "bands-of-force-greater"},
		{"  ", ""},
	} {
		if got := slugify(test.name); got != test.want {
			t.Errorf("slugify(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestUniqueID(t *testing.T) {
	taken := map[string]bool{}
	for _, want := range []struct{ name, id string }{
		{"Slash", "slash"},
		{"Slash", "slash-2"},
		{"Slash", "slash-3"},
		{"™", "item"},
	} {
		if got := uniqueID(want.name, taken); got != want.id {
			t.Errorf("uniqueID(%q) = %q, want %q", want.name, got, want.id)
		}
	}
}
```

##### Commit message

```
feat(E2): read Foundry actor exports

- foundry: read actor and item exports; an item keeps its system block
  raw and decodes the shared fields once
- damage rolls keep the export's order
- slugify and uniqueID turn item names into keys
- testdata: copies of the two boss exports and the two loose items
```

#### Commit 3 — `feat(E2): import NPCs and add pf2e sim import`

##### Task 1 — `internal/pf2e/sim/foundry/report.go`

`internal/pf2e/sim/foundry/report.go` (new) — full contents:

```go
package foundry

import "github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"

const (
	ReasonAbility     = "ability not converted yet (rules in Effort 3, descriptions in Effort 4)"
	ReasonEffort3     = "needs a rule from Effort 3"
	ReasonConditional = "conditional: does not hold at baseline"
	ReasonNotCombat   = "not a combat statistic"
	ReasonState       = "sheet state, ignored (D-26)"
	ReasonInactive    = "item not worn, held or invested"
	ReasonToggle      = "toggle, imported off"
	ReasonQuantity    = "quantity 0"
	ReasonNotPlaced   = "loose item, not placed on a creature"
	ReasonEmpty       = "holds nothing the sim uses"
	ReasonUnsupported = "not supported by the importer"
)

type Skip struct {
	Item   string
	What   string
	Reason string
}

type Report struct {
	Source   string
	Creature string
	Kind     creature.Kind
	Folder   string
	Skipped  []Skip
	Unused   map[string]int
}

func (report *Report) skip(itemName, what, reason string) {
	report.Skipped = append(report.Skipped, Skip{Item: itemName, What: what, Reason: reason})
}

func (report *Report) unused(kind string) {
	if report.Unused == nil {
		report.Unused = map[string]int{}
	}
	report.Unused[kind]++
}
```

##### Task 2 — `internal/pf2e/sim/foundry/npc.go`

An NPC's numbers come across as stored (D-6). Anything the schema cannot hold yet is skipped with the reason (Decision 15). `spellcastingEntrySystem` is shared with the PC import (Commit 4).

`internal/pf2e/sim/foundry/npc.go` (new) — full contents:

```go
package foundry

import (
	"fmt"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type npcSystem struct {
	Details    levelDetails  `json:"details"`
	Attributes npcAttributes `json:"attributes"`
	Perception npcPerception `json:"perception"`
	Saves      npcSaves      `json:"saves"`
	Traits     itemTraits    `json:"traits"`
}

type levelDetails struct {
	Level valueOf[int] `json:"level"`
}

type npcAttributes struct {
	HP          npcHP           `json:"hp"`
	AC          valueOf[int]    `json:"ac"`
	AllSaves    valueOf[string] `json:"allSaves"`
	Speed       npcSpeed        `json:"speed"`
	Immunities  []npcAdjustment `json:"immunities"`
	Resistances []npcAdjustment `json:"resistances"`
	Weaknesses  []npcAdjustment `json:"weaknesses"`
}

type npcHP struct {
	Max int `json:"max"`
}

type npcSpeed struct {
	Value       *int            `json:"value"`
	OtherSpeeds []npcOtherSpeed `json:"otherSpeeds"`
}

type npcOtherSpeed struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type npcAdjustment struct {
	Type       string   `json:"type"`
	Value      int      `json:"value"`
	Exceptions []string `json:"exceptions"`
	DoubleVs   []string `json:"doubleVs"`
}

type npcPerception struct {
	Mod int `json:"mod"`
}

type npcSaves struct {
	Fortitude valueOf[int] `json:"fortitude"`
	Reflex    valueOf[int] `json:"reflex"`
	Will      valueOf[int] `json:"will"`
}

type meleeSystem struct {
	Bonus         valueOf[int]      `json:"bonus"`
	DamageRolls   damageRolls       `json:"damageRolls"`
	Traits        itemTraits        `json:"traits"`
	AttackEffects valueOf[[]string] `json:"attackEffects"`
}

type spellcastingEntrySystem struct {
	Ability valueOf[string]      `json:"ability"`
	SpellDC spellDC              `json:"spelldc"`
	Slots   map[string]slotEntry `json:"slots"`
}

type spellDC struct {
	Value int `json:"value"`
	DC    int `json:"dc"`
}

type slotEntry struct {
	Max int `json:"max"`
}

func importNPC(source *actor) (*creature.Creature, *Report, error) {
	var system npcSystem
	if err := source.decode(&system); err != nil {
		return nil, nil, err
	}
	report := &Report{Kind: creature.KindNPC}
	attributes := system.Attributes
	layer := &creature.Creature{
		Schema:     creature.SchemaVersion,
		Name:       source.Name,
		Kind:       creature.KindNPC,
		Level:      system.Details.Level.Value,
		Traits:     npcTraits(system.Traits),
		Perception: system.Perception.Mod,
		Speed:      npcSpeedOf(attributes.Speed),
		Defense: creature.Defense{
			HP:        attributes.HP.Max,
			AC:        attributes.AC.Value,
			Fortitude: system.Saves.Fortitude.Value,
			Reflex:    system.Saves.Reflex.Value,
			Will:      system.Saves.Will.Value,
		},
	}
	if text := attributes.AllSaves.Value; text != "" {
		report.skip(source.Name, "saves: "+text, ReasonConditional)
	}
	for _, immunity := range attributes.Immunities {
		layer.Defense.Immune = append(layer.Defense.Immune, immunity.Type)
		if len(immunity.Exceptions) > 0 {
			report.skip(source.Name, fmt.Sprintf("immunity %s except %s", immunity.Type, strings.Join(immunity.Exceptions, ", ")), ReasonEffort3)
		}
	}
	layer.Defense.Resist = npcAdjustments(attributes.Resistances, "resistance", source.Name, report)
	layer.Defense.Weakness = npcAdjustments(attributes.Weaknesses, "weakness", source.Name, report)

	strikeIDs := map[string]bool{}
	for index := range source.Items {
		current := &source.Items[index]
		var err error
		switch current.Type {
		case "melee":
			err = importNPCStrike(current, strikeIDs, layer, report)
		case "spellcastingEntry":
			err = importNPCSpellcasting(current, layer, report)
		case "spell":
			report.skip(current.Name, "spell", ReasonEffort3)
		case "action":
			report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
		case "lore":
			report.skip(current.Name, "lore", ReasonNotCombat)
		default:
			report.skip(current.Name, current.Type, ReasonUnsupported)
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return layer, report, nil
}

func npcTraits(source itemTraits) []string {
	var traits []string
	if source.Rarity != "" && source.Rarity != "common" {
		traits = append(traits, source.Rarity)
	}
	return append(traits, source.Value...)
}

func npcSpeedOf(source npcSpeed) creature.Speed {
	var speed creature.Speed
	if source.Value != nil {
		speed.Land = *source.Value
	}
	for _, other := range source.OtherSpeeds {
		if other.Type == "fly" {
			speed.Fly = other.Value
		}
	}
	return speed
}

func npcAdjustments(entries []npcAdjustment, label, owner string, report *Report) map[string]creature.Adjustment {
	adjustments := map[string]creature.Adjustment{}
	for _, entry := range entries {
		key := entry.Type
		if key == "all-damage" {
			key = "all"
		}
		if key != "all" && key != "physical" && !creature.IsDamageType(key) {
			report.skip(owner, fmt.Sprintf("%s %s %d", label, entry.Type, entry.Value), ReasonEffort3)
			continue
		}
		if len(entry.DoubleVs) > 0 {
			report.skip(owner, fmt.Sprintf("%s %s doubled against %s", label, key, strings.Join(entry.DoubleVs, ", ")), ReasonEffort3)
		}
		adjustments[key] = creature.Adjustment{Value: entry.Value, Except: entry.Exceptions}
	}
	if len(adjustments) == 0 {
		return nil
	}
	return adjustments
}

func importNPCStrike(current *item, taken map[string]bool, layer *creature.Creature, report *Report) error {
	var system meleeSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	strike := creature.Strike{Name: current.Name, Bonus: system.Bonus.Value, Traits: system.Traits.Value}
	for _, roll := range system.DamageRolls {
		if roll.Category != "" {
			report.skip(current.Name, fmt.Sprintf("%s %s %s damage", roll.Damage, roll.DamageType, roll.Category), ReasonEffort3)
			continue
		}
		strike.Damage = append(strike.Damage, creature.DamagePart{Dice: roll.Damage, Type: roll.DamageType})
	}
	for _, effect := range system.AttackEffects.Value {
		report.skip(current.Name, "attack effect "+effect, ReasonEffort3)
	}
	if layer.Strikes == nil {
		layer.Strikes = map[string]creature.Strike{}
	}
	layer.Strikes[uniqueID(current.Name, taken)] = strike
	return nil
}

func importNPCSpellcasting(current *item, layer *creature.Creature, report *Report) error {
	var system spellcastingEntrySystem
	if err := current.decode(&system); err != nil {
		return err
	}
	if layer.Spellcasting != nil {
		report.skip(current.Name, "second spellcasting entry", ReasonUnsupported)
		return nil
	}
	layer.Spellcasting = &creature.Spellcasting{Attack: system.SpellDC.Value, DC: system.SpellDC.DC}
	return nil
}
```

##### Task 3 — `internal/pf2e/sim/foundry/write.go`

Until Commit 4, a character export is reported and not written.

`internal/pf2e/sim/foundry/write.go` (new) — full contents:

```go
package foundry

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var ErrNoExports = errors.New("no actor exports in source/actors")

const layerHeader = "# Written by gm-toolkit pf2e sim import from %s.\n# Import rewrites this file. Put changes in hand.toml.\n\n"

func ImportFight(fightDirectory, partyDirectory string) ([]*Report, error) {
	sourceDirectory := filepath.Join(fightDirectory, "source")
	actorPaths, err := filepath.Glob(filepath.Join(sourceDirectory, "actors", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("import %s: %w", fightDirectory, err)
	}
	if len(actorPaths) == 0 {
		return nil, fmt.Errorf("import %s: %w", fightDirectory, ErrNoExports)
	}
	var reports []*Report
	for _, path := range actorPaths {
		report, err := importActorFile(path, sourceDirectory, fightDirectory, partyDirectory)
		if err != nil {
			return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
		}
		reports = append(reports, report)
	}
	itemPaths, err := filepath.Glob(filepath.Join(sourceDirectory, "items", "*.json"))
	if err != nil {
		return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
	}
	for _, path := range itemPaths {
		loose, err := readItem(path)
		if err != nil {
			return reports, fmt.Errorf("import %s: %w", fightDirectory, err)
		}
		report := &Report{Source: relativeSource(sourceDirectory, path)}
		report.skip(loose.Name, loose.Type, ReasonNotPlaced)
		reports = append(reports, report)
	}
	return reports, nil
}

func importActorFile(path, sourceDirectory, fightDirectory, partyDirectory string) (*Report, error) {
	source, err := readActor(path)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	if source.Type == "character" {
		report := &Report{Source: relativeSource(sourceDirectory, path), Creature: id, Kind: creature.KindPC}
		report.skip(source.Name, "build choices", ReasonUnsupported)
		return report, nil
	}
	layer, report, err := importNPC(source)
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(fightDirectory, id)
	report.Source = relativeSource(sourceDirectory, path)
	report.Creature = id
	report.Folder = directory
	if err := writeLayer(directory, report.Source, layer); err != nil {
		return nil, err
	}
	return report, nil
}

func writeLayer(directory, source string, layer *creature.Creature) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("write layer %s: %w", directory, err)
	}
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, layerHeader, source)
	if err := toml.NewEncoder(&buffer).Encode(layer); err != nil {
		return fmt.Errorf("write layer %s: %w", directory, err)
	}
	path := filepath.Join(directory, creature.ImportLayer)
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write layer %s: %w", path, err)
	}
	return nil
}

func relativeSource(sourceDirectory, path string) string {
	relative, err := filepath.Rel(sourceDirectory, path)
	if err != nil {
		return filepath.Base(path)
	}
	return filepath.ToSlash(relative)
}
```

##### Task 4 — `internal/pf2e/sim/foundry/npc_test.go`

`internal/pf2e/sim/foundry/npc_test.go` (new) — full contents:

```go
package foundry

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func hasSkip(report *Report, itemName, reason string) bool {
	return slices.ContainsFunc(report.Skipped, func(skip Skip) bool {
		return skip.Item == itemName && skip.Reason == reason
	})
}

func importTestNPC(t *testing.T, name string) (*creature.Creature, *Report) {
	t.Helper()
	source, err := readActor(filepath.Join("testdata", "actors", name+".json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	layer, report, err := importNPC(source)
	if err != nil {
		t.Fatalf("importNPC: %v", err)
	}
	return layer, report
}

func TestImportNPCLuna(t *testing.T) {
	layer, report := importTestNPC(t, "abyssal-luna")
	if layer.Name != "Abyssal Luna" || layer.Kind != creature.KindNPC || layer.Level != 16 || layer.Perception != 29 {
		t.Errorf("header = %q %s %d perception %d, want Abyssal Luna npc 16 perception 29", layer.Name, layer.Kind, layer.Level, layer.Perception)
	}
	if want := []string{"unique", "fey", "occult"}; !slices.Equal(layer.Traits, want) {
		t.Errorf("Traits = %v, want %v", layer.Traits, want)
	}
	if want := (creature.Speed{Land: 60, Fly: 180}); layer.Speed != want {
		t.Errorf("Speed = %+v, want %+v", layer.Speed, want)
	}
	defense := layer.Defense
	if defense.HP != 300 || defense.AC != 36 || defense.Fortitude != 27 || defense.Reflex != 28 || defense.Will != 29 {
		t.Errorf("defense = %+v, want HP 300 AC 36 saves 27/28/29", defense)
	}
	if want := []string{"confused", "doomed", "paralyzed", "sleep"}; !slices.Equal(defense.Immune, want) {
		t.Errorf("Immune = %v, want %v", defense.Immune, want)
	}
	slash := layer.Strikes["slash"]
	wantSlash := creature.Strike{
		Name:   "Slash",
		Bonus:  30,
		Traits: []string{"agile", "magical", "reach-10", "unarmed"},
		Damage: []creature.DamagePart{{Dice: "3d6+14", Type: "slashing"}, {Dice: "1d8", Type: "mental"}},
	}
	if !reflect.DeepEqual(slash, wantSlash) {
		t.Errorf("slash = %+v, want %+v", slash, wantSlash)
	}
	if ids := slices.Sorted(maps.Keys(layer.Strikes)); !slices.Equal(ids, []string{"slash", "stab"}) {
		t.Errorf("strikes = %v, want slash and stab", ids)
	}
	if want := (creature.Spellcasting{Attack: 31, DC: 39}); layer.Spellcasting == nil || *layer.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", layer.Spellcasting, want)
	}
	if len(layer.Abilities) != 0 {
		t.Errorf("Abilities = %v, want none (Decision 4)", layer.Abilities)
	}
	for _, want := range []struct{ item, reason string }{
		{"Destiny Spores", ReasonAbility},
		{"Challenge Fate", ReasonAbility},
		{"Slash", ReasonEffort3},
		{"Ill Omen (At Will)", ReasonEffort3},
		{"Abyssal Luna", ReasonConditional},
		{"Fortune-Telling Lore", ReasonNotCombat},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}

func TestImportNPCLilith(t *testing.T) {
	layer, report := importTestNPC(t, "avatar-of-lilith")
	if want := (creature.Speed{Land: 0, Fly: 60}); layer.Speed != want {
		t.Errorf("Speed = %+v, want %+v (no land speed on the sheet)", layer.Speed, want)
	}
	wantResist := map[string]creature.Adjustment{"all": {Value: 12, Except: []string{"holy", "cold-iron"}}}
	if !reflect.DeepEqual(layer.Defense.Resist, wantResist) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, wantResist)
	}
	if layer.Defense.Weakness != nil {
		t.Errorf("Weakness = %+v, want none: holy is a trait (Decision 15)", layer.Defense.Weakness)
	}
	if !hasSkip(report, "Avatar of Lilith", ReasonEffort3) {
		t.Error("report lacks the holy weakness as needing Effort 3")
	}
	if ids := slices.Sorted(maps.Keys(layer.Strikes)); !slices.Equal(ids, []string{"hand-of-chaos", "slash", "stinger"}) {
		t.Errorf("strikes = %v, want hand-of-chaos, slash, stinger", ids)
	}
	if want := (creature.Spellcasting{Attack: 32, DC: 38}); layer.Spellcasting == nil || *layer.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", layer.Spellcasting, want)
	}
	if !hasSkip(report, "Last Stand", ReasonAbility) {
		t.Error("report lacks Last Stand as an ability")
	}
}

func TestImportedNPCLoadsUnderItsHandLayer(t *testing.T) {
	layer, _ := importTestNPC(t, "abyssal-luna")
	directory := filepath.Join(t.TempDir(), "abyssal-luna")
	if err := writeLayer(directory, "actors/abyssal-luna.json", layer); err != nil {
		t.Fatalf("writeLayer: %v", err)
	}
	hand := "schema = 1\n\n[ability.abyssal-frenzy]\nname = \"Abyssal Frenzy\"\nactions = 2\nsteps = [{ strike = \"slash\" }, { strike = \"slash\" }, { strike = \"stab\" }]\n"
	if err := os.WriteFile(filepath.Join(directory, "hand.toml"), []byte(hand), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	loaded, err := creature.Load(directory)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 300 || loaded.Strikes["slash"].Bonus != 30 {
		t.Errorf("loaded HP %d, slash %+d; want 300 and +30", loaded.Defense.HP, loaded.Strikes["slash"].Bonus)
	}
	if steps := loaded.Abilities["abyssal-frenzy"].Steps; len(steps) != 3 {
		t.Errorf("abyssal-frenzy steps = %v, want 3", steps)
	}
	if want := []string{"import.toml", "hand.toml"}; !slices.Equal(loaded.Layers, want) {
		t.Errorf("Layers = %v, want %v", loaded.Layers, want)
	}
}
```

##### Task 5 — `internal/pf2e/sim/foundry/write_test.go`

`internal/pf2e/sim/foundry/write_test.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	content, err := os.ReadFile(from)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(to, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func TestImportFightWritesNPCLayers(t *testing.T) {
	fight := filepath.Join(t.TempDir(), "fall-of-luna")
	party := filepath.Join(t.TempDir(), "party")
	copyFile(t, filepath.Join("testdata", "actors", "abyssal-luna.json"), filepath.Join(fight, "source", "actors", "abyssal-luna.json"))
	copyFile(t, filepath.Join("testdata", "items", "swear-fidelity.json"), filepath.Join(fight, "source", "items", "swear-fidelity.json"))

	reports, err := ImportFight(fight, party)
	if err != nil {
		t.Fatalf("ImportFight: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("reports = %d, want 2", len(reports))
	}
	luna := reports[0]
	if luna.Source != "actors/abyssal-luna.json" || luna.Creature != "abyssal-luna" || luna.Folder != filepath.Join(fight, "abyssal-luna") {
		t.Errorf("luna report = %+v", luna)
	}
	content, err := os.ReadFile(filepath.Join(fight, "abyssal-luna", "import.toml"))
	if err != nil {
		t.Fatalf("read layer: %v", err)
	}
	if !strings.HasPrefix(string(content), "# Written by gm-toolkit pf2e sim import from actors/abyssal-luna.json.") {
		t.Errorf("layer header = %q", strings.SplitN(string(content), "\n", 2)[0])
	}
	loaded, err := creature.Load(luna.Folder)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Defense.HP != 300 {
		t.Errorf("HP = %d, want 300", loaded.Defense.HP)
	}
	loose := reports[1]
	if loose.Source != "items/swear-fidelity.json" || loose.Folder != "" || !hasSkip(loose, "Swear Fidelity", ReasonNotPlaced) {
		t.Errorf("loose item report = %+v", loose)
	}
}

func TestImportFightNeedsExports(t *testing.T) {
	_, err := ImportFight(t.TempDir(), t.TempDir())
	if !errors.Is(err, ErrNoExports) {
		t.Errorf("err = %v, want ErrNoExports", err)
	}
}
```

##### Task 6 — `cmd/gm-toolkit/pf2e/import.go`

`cmd/gm-toolkit/pf2e/import.go` (new) — full contents:

```go
package pf2e

import (
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/foundry"
)

func newImportCmd(campaignOverride *string) *cobra.Command {
	return &cobra.Command{
		Use:   "import <fight>",
		Short: "Import a fight's Foundry exports into creature files",
		Long:  "Reads encounters/<fight>/source/actors/*.json and writes each creature's import.toml: PCs under pf2e/party/<pc>/, NPCs under encounters/<fight>/<npc>/. For each creature it prints the merged stat block and everything it did not import, with the reason.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fight := args[0]
			if !creature.IsFolderName(fight) {
				return fmt.Errorf("pf2e sim import: bad fight name %q", fight)
			}
			campaign, err := resolveCampaign(*campaignOverride)
			if err != nil {
				return fmt.Errorf("pf2e sim import: %w", err)
			}
			paths := campaign.Paths()
			reports, importErr := foundry.ImportFight(filepath.Join(paths.EncountersDir, fight), paths.PartyDir)
			out := cmd.OutOrStdout()
			styled := isTerminal(out)
			failed := 0
			for _, report := range reports {
				var shown *creature.Creature
				var loadErr error
				if report.Folder != "" {
					shown, loadErr = creature.Load(report.Folder)
					if loadErr != nil {
						failed++
					}
				}
				printReport(out, report, shown, loadErr, styled)
			}
			if importErr != nil {
				return fmt.Errorf("pf2e sim import: %w", importErr)
			}
			if failed > 0 {
				return fmt.Errorf("pf2e sim import: %d creatures failed to load", failed)
			}
			return nil
		},
	}
}

func printReport(out io.Writer, report *foundry.Report, shown *creature.Creature, loadErr error, styled bool) {
	bold, reset := "", ""
	if styled {
		bold, reset = ansiBold, ansiReset
	}
	switch {
	case report.Folder == "":
		fmt.Fprintf(out, "%s%s%s  not imported\n", bold, report.Source, reset)
	case loadErr != nil:
		fmt.Fprintf(out, "%s%s%s  wrote %s\n", bold, report.Source, reset, filepath.Join(report.Folder, creature.ImportLayer))
		fmt.Fprintf(out, "Load failed: %v\n", loadErr)
	default:
		fmt.Fprintf(out, "%s%s%s  wrote %s\n", bold, report.Source, reset, filepath.Join(report.Folder, creature.ImportLayer))
		printCreature(out, shown, styled)
	}

	byReason := map[string][]foundry.Skip{}
	for _, skip := range report.Skipped {
		byReason[skip.Reason] = append(byReason[skip.Reason], skip)
	}
	for _, reason := range slices.Sorted(maps.Keys(byReason)) {
		fmt.Fprintf(out, "Skipped, %s:\n", reason)
		for _, skip := range byReason[reason] {
			fmt.Fprintf(out, "  %s: %s\n", skip.Item, skip.What)
		}
	}
	if len(report.Unused) > 0 {
		counts := make([]string, 0, len(report.Unused))
		for _, kind := range slices.Sorted(maps.Keys(report.Unused)) {
			counts = append(counts, fmt.Sprintf("%s ×%d", kind, report.Unused[kind]))
		}
		fmt.Fprintf(out, "Not used by the sim: %s\n", strings.Join(counts, ", "))
	}
	fmt.Fprintln(out)
}
```

##### Task 7 — `cmd/gm-toolkit/pf2e/import_test.go`

`cmd/gm-toolkit/pf2e/import_test.go` (new) — full contents:

```go
package pf2e

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/foundry"
)

func TestPrintReport(t *testing.T) {
	report := &foundry.Report{
		Source:   "actors/kas.json",
		Creature: "kas",
		Kind:     creature.KindPC,
		Folder:   "party/kas",
		Skipped: []foundry.Skip{
			{Item: "Keen Eyes", What: "FlatModifier perception", Reason: foundry.ReasonConditional},
			{Item: "Bravery", What: "AdjustDegreeOfSuccess", Reason: foundry.ReasonEffort3},
		},
		Unused: map[string]int{"Note": 6, "GrantItem": 4},
	}
	shown := &creature.Creature{
		Name:    "Kas Fringebramble",
		Kind:    creature.KindPC,
		Level:   14,
		Defense: creature.Defense{HP: 230, AC: 36},
		Layers:  []string{"import.toml"},
	}
	var out bytes.Buffer
	printReport(&out, report, shown, nil, false)
	for _, want := range []string{
		"actors/kas.json  wrote party/kas/import.toml",
		"Kas Fringebramble  pc 14",
		"Skipped, " + foundry.ReasonConditional + ":\n  Keen Eyes: FlatModifier perception",
		"Skipped, " + foundry.ReasonEffort3 + ":\n  Bravery: AdjustDegreeOfSuccess",
		"Not used by the sim: GrantItem ×4, Note ×6",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}

	out.Reset()
	printReport(&out, report, nil, errors.New("bad layer"), false)
	if !strings.Contains(out.String(), "Load failed: bad layer") {
		t.Errorf("output lacks the load error:\n%s", out.String())
	}

	out.Reset()
	printReport(&out, &foundry.Report{Source: "items/swear-fidelity.json"}, nil, nil, false)
	if !strings.Contains(out.String(), "items/swear-fidelity.json  not imported") {
		t.Errorf("output lacks the loose item line:\n%s", out.String())
	}
}
```

##### Task 8 — `cmd/gm-toolkit/pf2e/sim.go`

In `newSimCmd`, register `import`.

**Find:**
```go
	cmd.AddCommand(newShowCmd(campaignOverride))
```
**Replace with:**
```go
	cmd.AddCommand(newShowCmd(campaignOverride))
	cmd.AddCommand(newImportCmd(campaignOverride))
```

##### Task 9 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-plan.md`

In the Work Breakdown Overview, tick Phase 2.

**Find:**
```
- [ ] **Phase 2 — Foundry reader and NPC import**
```
**Replace with:**
```
- [x] **Phase 2 — Foundry reader and NPC import**
```

##### Commit message

```
feat(E2): import NPCs and add pf2e sim import

- foundry: an NPC's import layer holds its numbers, Strikes and spell
  numbers as stored; abilities, trait weaknesses, attack effects and
  damage categories are reported, not written
- ImportFight writes each export's import.toml and reports loose items
- pf2e sim import <fight> prints each creature's merged stat block and
  what was skipped, grouped by reason
- creature: export ImportLayer and IsFolderName
```

### Phase 3 — PC build choices

**In plain words:** the importer writes each player character's build from the choices stored on the class, ancestry, background, feats and gear, and `pf2e sim import` puts PCs in the party folder. Item bonuses and ranks held in Foundry's rule elements wait for Phase 4.

#### Commit 4 — `feat(E2): import PC build choices`

##### Task 1 — `internal/pf2e/sim/foundry/testdata/actors/`

Copy the three PC exports as they are (Decision 7). Run from the repository root:

```bash
cp campaigns/hawk/encounters/fall-of-luna/source/actors/kas.json \
   campaigns/hawk/encounters/fall-of-luna/source/actors/tavish.json \
   campaigns/hawk/encounters/fall-of-luna/source/actors/arabesque.json \
   internal/pf2e/sim/foundry/testdata/actors/
```

##### Task 2 — `internal/pf2e/sim/foundry/pc.go`

The class is read first, because a feat names the class DC by the class's slug. Ranks are kept as numbers while importing and written as words at the end. The class DC starts at trained, as pf2e system 8.6.1 gives every class. `pcImporter` already holds the fields Commit 5's rule elements fill (`resist`, `dexterityCaps`, `speedPenaltySuppressed`, `modifierIDs`); until then they stay empty.

`internal/pf2e/sim/foundry/pc.go` (new) — full contents:

```go
package foundry

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var (
	specializationSlugs = map[string]string{"weapon-specialization": "normal", "greater-weapon-specialization": "greater"}
	abilityActionTypes  = []string{"action", "reaction", "free"}
)

type characterSystem struct {
	Details levelDetails   `json:"details"`
	Build   characterBuild `json:"build"`
}

type characterBuild struct {
	Attributes buildAttributes `json:"attributes"`
}

type buildAttributes struct {
	Boosts map[string][]string `json:"boosts"`
}

type boostChoice struct {
	Value    []string `json:"value"`
	Selected string   `json:"selected"`
}

type classSystem struct {
	HP           int            `json:"hp"`
	KeyAbility   boostChoice    `json:"keyAbility"`
	Perception   int            `json:"perception"`
	SavingThrows map[string]int `json:"savingThrows"`
	Attacks      classAttacks   `json:"attacks"`
	Defenses     map[string]int `json:"defenses"`
	Spellcasting int            `json:"spellcasting"`
}

type classAttacks struct {
	Simple   int `json:"simple"`
	Martial  int `json:"martial"`
	Advanced int `json:"advanced"`
	Unarmed  int `json:"unarmed"`
}

type ancestrySystem struct {
	HP                      int                    `json:"hp"`
	Speed                   int                    `json:"speed"`
	Boosts                  map[string]boostChoice `json:"boosts"`
	Flaws                   map[string]boostChoice `json:"flaws"`
	AlternateAncestryBoosts []string               `json:"alternateAncestryBoosts"`
	Traits                  itemTraits             `json:"traits"`
}

type backgroundSystem struct {
	Boosts map[string]boostChoice `json:"boosts"`
}

type featSystem struct {
	Subfeatures subfeatures `json:"subfeatures"`
}

type subfeatures struct {
	Proficiencies map[string]rankEntry `json:"proficiencies"`
}

type rankEntry struct {
	Rank int `json:"rank"`
}

type armorSystem struct {
	Category     string     `json:"category"`
	Group        string     `json:"group"`
	ACBonus      int        `json:"acBonus"`
	DexCap       int        `json:"dexCap"`
	Strength     int        `json:"strength"`
	SpeedPenalty int        `json:"speedPenalty"`
	Runes        armorRunes `json:"runes"`
}

type armorRunes struct {
	Potency   int      `json:"potency"`
	Resilient int      `json:"resilient"`
	Property  []string `json:"property"`
}

type weaponSystem struct {
	Category     string       `json:"category"`
	Group        string       `json:"group"`
	BaseItem     string       `json:"baseItem"`
	Damage       weaponDamage `json:"damage"`
	Runes        weaponRunes  `json:"runes"`
	Range        *int         `json:"range"`
	Bonus        valueOf[int] `json:"bonus"`
	SplashDamage valueOf[int] `json:"splashDamage"`
}

type weaponDamage struct {
	Dice       int               `json:"dice"`
	Die        string            `json:"die"`
	DamageType string            `json:"damageType"`
	Persistent *persistentDamage `json:"persistent"`
}

type persistentDamage struct {
	Number int    `json:"number"`
	Type   string `json:"type"`
}

type weaponRunes struct {
	Potency  int      `json:"potency"`
	Striking int      `json:"striking"`
	Property []string `json:"property"`
}

type spellSystem struct {
	Location valueOf[string] `json:"location"`
}

type pcImporter struct {
	source                 *actor
	report                 *Report
	level                  int
	build                  *creature.Build
	traits                 []string
	classSlug              string
	ranks                  map[string]int
	resist                 map[string]creature.Adjustment
	armor                  *armorSystem
	armorName              string
	dexterityCaps          []int
	speedPenaltySuppressed bool
	weaponIDs              map[string]bool
	modifierIDs            map[string]bool
	entries                []*item
	spellsByEntry          map[string]int
}

func importPC(source *actor) (*creature.Creature, *Report, error) {
	var system characterSystem
	if err := source.decode(&system); err != nil {
		return nil, nil, err
	}
	importer := &pcImporter{
		source: source,
		report: &Report{Kind: creature.KindPC},
		level:  system.Details.Level.Value,
		build: &creature.Build{
			Boosts:      map[string][]string{},
			Proficiency: map[string]string{},
			Weapons:     map[string]creature.Weapon{},
			Modifiers:   map[string]creature.Modifier{},
		},
		ranks:         map[string]int{"class-dc": 1},
		resist:        map[string]creature.Adjustment{},
		weaponIDs:     map[string]bool{},
		modifierIDs:   map[string]bool{},
		spellsByEntry: map[string]int{},
	}
	for level, boosts := range system.Build.Attributes.Boosts {
		importer.build.Boosts["level-"+level] = boosts
	}
	for index := range source.Items {
		if current := &source.Items[index]; current.Type == "class" {
			if err := importer.class(current); err != nil {
				return nil, nil, err
			}
		}
	}
	for index := range source.Items {
		if err := importer.item(&source.Items[index]); err != nil {
			return nil, nil, err
		}
	}
	if err := importer.spellcasting(); err != nil {
		return nil, nil, err
	}
	return importer.finish(), importer.report, nil
}

func (importer *pcImporter) item(current *item) error {
	switch current.Type {
	case "ancestry":
		return importer.ancestry(current)
	case "background":
		return importer.background(current)
	case "feat":
		return importer.feat(current)
	case "armor":
		return importer.armorItem(current)
	case "weapon":
		return importer.weapon(current)
	case "spellcastingEntry":
		importer.entries = append(importer.entries, current)
	case "spell":
		var system spellSystem
		if err := current.decode(&system); err != nil {
			return err
		}
		importer.spellsByEntry[system.Location.Value]++
	case "action":
		importer.report.skip(current.Name, current.System.ActionType.Value, ReasonAbility)
	case "condition", "effect":
		importer.report.skip(current.Name, current.Type, ReasonState)
	case "shield":
		importer.report.skip(current.Name, "shield", ReasonEffort3)
	}
	return nil
}

func (importer *pcImporter) raise(key string, rank int) {
	importer.ranks[key] = max(importer.ranks[key], rank)
}

func (importer *pcImporter) class(current *item) error {
	var system classSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.classSlug = current.System.Slug
	importer.build.ClassHP = system.HP
	switch {
	case system.KeyAbility.Selected != "":
		importer.build.Key = system.KeyAbility.Selected
	case len(system.KeyAbility.Value) == 1:
		importer.build.Key = system.KeyAbility.Value[0]
	default:
		importer.report.skip(current.Name, "key attribute not chosen", ReasonUnsupported)
	}
	if importer.build.Key != "" {
		importer.build.Boosts["class"] = []string{importer.build.Key}
	}
	importer.raise("perception", system.Perception)
	for _, save := range slices.Sorted(maps.Keys(system.SavingThrows)) {
		importer.raise(save, system.SavingThrows[save])
	}
	importer.raise("simple", system.Attacks.Simple)
	importer.raise("martial", system.Attacks.Martial)
	importer.raise("advanced", system.Attacks.Advanced)
	importer.raise("unarmed", system.Attacks.Unarmed)
	for _, category := range slices.Sorted(maps.Keys(system.Defenses)) {
		importer.raise(category, system.Defenses[category])
	}
	importer.raise("spellcasting", system.Spellcasting)
	return nil
}

func (importer *pcImporter) ancestry(current *item) error {
	var system ancestrySystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.build.AncestryHP = system.HP
	importer.build.Speed = system.Speed
	importer.traits = system.Traits.Value
	if system.AlternateAncestryBoosts != nil {
		importer.build.Boosts["ancestry"] = system.AlternateAncestryBoosts
		return nil
	}
	importer.build.Boosts["ancestry"] = importer.choices(current.Name, "boost", system.Boosts)
	importer.build.Flaws = importer.choices(current.Name, "flaw", system.Flaws)
	return nil
}

func (importer *pcImporter) background(current *item) error {
	var system backgroundSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.build.Boosts["background"] = importer.choices(current.Name, "boost", system.Boosts)
	return nil
}

func (importer *pcImporter) choices(owner, label string, entries map[string]boostChoice) []string {
	var chosen []string
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		entry := entries[key]
		switch {
		case entry.Selected != "":
			chosen = append(chosen, entry.Selected)
		case len(entry.Value) == 1:
			chosen = append(chosen, entry.Value[0])
		default:
			importer.report.skip(owner, fmt.Sprintf("%s %s not chosen", label, key), ReasonUnsupported)
		}
	}
	return chosen
}

func (importer *pcImporter) feat(current *item) error {
	if level, found := specializationSlugs[current.System.Slug]; found && importer.build.WeaponSpecialization != "greater" {
		importer.build.WeaponSpecialization = level
	}
	if actionType := current.System.ActionType.Value; slices.Contains(abilityActionTypes, actionType) {
		importer.report.skip(current.Name, actionType, ReasonAbility)
	}
	var system featSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	for _, key := range slices.Sorted(maps.Keys(system.Subfeatures.Proficiencies)) {
		target := key
		if key == importer.classSlug {
			target = "class-dc"
		}
		if !creature.IsProficiencyKey(target) {
			importer.report.skip(current.Name, "proficiency "+key, ReasonUnsupported)
			continue
		}
		importer.raise(target, system.Subfeatures.Proficiencies[key].Rank)
	}
	return nil
}

func (importer *pcImporter) armorItem(current *item) error {
	equipped := current.System.Equipped
	if equipped.CarryType != "worn" || (equipped.Invested != nil && !*equipped.Invested) {
		importer.report.skip(current.Name, "armor", ReasonInactive)
		return nil
	}
	if importer.armor != nil {
		importer.report.skip(current.Name, "second worn armor", ReasonUnsupported)
		return nil
	}
	var system armorSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	importer.armor = &system
	importer.armorName = current.Name
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonEffort3)
	}
	return nil
}

func (importer *pcImporter) weapon(current *item) error {
	if current.System.Quantity < 1 {
		importer.report.skip(current.Name, "weapon", ReasonQuantity)
		return nil
	}
	var system weaponSystem
	if err := current.decode(&system); err != nil {
		return err
	}
	die, err := strconv.Atoi(strings.TrimPrefix(system.Damage.Die, "d"))
	if err != nil {
		importer.report.skip(current.Name, "damage die "+system.Damage.Die, ReasonUnsupported)
		return nil
	}
	weapon := creature.Weapon{
		Name:      current.Name,
		Category:  system.Category,
		Group:     system.Group,
		Base:      system.BaseItem,
		Dice:      system.Damage.Dice,
		Die:       die,
		Type:      system.Damage.DamageType,
		Potency:   system.Runes.Potency,
		Striking:  system.Runes.Striking,
		ItemBonus: system.Bonus.Value,
		Traits:    current.System.Traits.Value,
	}
	if system.Range != nil {
		weapon.Range = *system.Range
	}
	if persistent := system.Damage.Persistent; persistent != nil {
		importer.report.skip(current.Name, fmt.Sprintf("persistent %d %s", persistent.Number, persistent.Type), ReasonEffort3)
	}
	if system.SplashDamage.Value > 0 {
		importer.report.skip(current.Name, fmt.Sprintf("splash %d", system.SplashDamage.Value), ReasonEffort3)
	}
	if len(system.Runes.Property) > 0 {
		importer.report.skip(current.Name, "property runes "+strings.Join(system.Runes.Property, ", "), ReasonEffort3)
	}
	if slices.Contains(weapon.Traits, "consumable") {
		importer.report.skip(current.Name, fmt.Sprintf("quantity %d, not tracked", current.System.Quantity), ReasonEffort3)
	}
	importer.build.Weapons[uniqueID(current.Name, importer.weaponIDs)] = weapon
	return nil
}

func (importer *pcImporter) spellcasting() error {
	slices.SortStableFunc(importer.entries, func(first, second *item) int { return strings.Compare(first.Name, second.Name) })
	var casting *creature.Casting
	for _, entry := range importer.entries {
		var system spellcastingEntrySystem
		if err := entry.decode(&system); err != nil {
			return err
		}
		ability := system.Ability.Value
		switch {
		case importer.spellsByEntry[entry.ID] == 0:
			importer.report.skip(entry.Name, "spellcasting entry with no spells", ReasonEmpty)
			continue
		case ability == "":
			importer.report.skip(entry.Name, "spellcasting entry on a class statistic", ReasonUnsupported)
			continue
		case casting == nil:
			casting = &creature.Casting{Attribute: ability}
		case casting.Attribute != ability:
			importer.report.skip(entry.Name, "spellcasting attribute "+ability, ReasonUnsupported)
			continue
		}
		casting.TopRank = max(casting.TopRank, topSlotRank(system.Slots))
	}
	importer.build.Spellcasting = casting
	spells := 0
	for _, count := range importer.spellsByEntry {
		spells += count
	}
	if spells > 0 {
		importer.report.skip(importer.source.Name, fmt.Sprintf("%d spells", spells), ReasonEffort3)
	}
	return nil
}

func topSlotRank(slots map[string]slotEntry) int {
	top := 0
	for key, slot := range slots {
		rank, err := strconv.Atoi(strings.TrimPrefix(key, "slot"))
		if err == nil && slot.Max > 0 {
			top = max(top, rank)
		}
	}
	return top
}

func (importer *pcImporter) finish() *creature.Creature {
	build := importer.build
	for _, key := range slices.Sorted(maps.Keys(importer.ranks)) {
		if rank := importer.ranks[key]; rank > 0 {
			build.Proficiency[key] = creature.RankWords[rank]
		}
	}
	build.Armor = importer.finishArmor()
	layer := &creature.Creature{
		Schema: creature.SchemaVersion,
		Name:   importer.source.Name,
		Kind:   creature.KindPC,
		Level:  importer.level,
		Traits: importer.traits,
		Build:  build,
	}
	if len(importer.resist) > 0 {
		layer.Defense.Resist = importer.resist
	}
	return layer
}

func (importer *pcImporter) finishArmor() *creature.Armor {
	var lowestCap *int
	if len(importer.dexterityCaps) > 0 {
		lowest := slices.Min(importer.dexterityCaps)
		lowestCap = &lowest
	}
	if importer.armor == nil {
		if lowestCap == nil {
			return nil
		}
		return &creature.Armor{Category: "unarmored", DexCap: lowestCap}
	}
	system := importer.armor
	armorCap := system.DexCap
	if lowestCap != nil {
		armorCap = min(armorCap, *lowestCap)
	}
	return &creature.Armor{
		Name:         importer.armorName,
		Category:     system.Category,
		ACBonus:      system.ACBonus,
		DexCap:       &armorCap,
		Potency:      system.Runes.Potency,
		Resilient:    system.Runes.Resilient,
		SpeedPenalty: importer.speedPenalty(),
	}
}

func (importer *pcImporter) speedPenalty() int {
	system := importer.armor
	if importer.speedPenaltySuppressed || system.SpeedPenalty >= 0 {
		return 0
	}
	attributes, err := importer.build.Attributes()
	if err != nil {
		importer.report.skip(importer.armorName, "speed penalty: "+err.Error(), ReasonUnsupported)
		return system.SpeedPenalty
	}
	if attributes["str"] >= system.Strength {
		return min(system.SpeedPenalty+5, 0)
	}
	return system.SpeedPenalty
}
```

##### Task 3 — `internal/pf2e/sim/foundry/write.go`

In `importActorFile`, a character export now goes through `importPC` into the party folder.

**Find:**
```go
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	if source.Type == "character" {
		report := &Report{Source: relativeSource(sourceDirectory, path), Creature: id, Kind: creature.KindPC}
		report.skip(source.Name, "build choices", ReasonUnsupported)
		return report, nil
	}
	layer, report, err := importNPC(source)
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(fightDirectory, id)
```
**Replace with:**
```go
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	importer, directory := importNPC, filepath.Join(fightDirectory, id)
	if source.Type == "character" {
		importer, directory = importPC, filepath.Join(partyDirectory, id)
	}
	layer, report, err := importer(source)
	if err != nil {
		return nil, err
	}
```

##### Task 4 — `internal/pf2e/sim/foundry/pc_test.go`

These tests check only what the build choices alone give. Ranks and bonuses that come from rule elements are Commit 5's tests.

`internal/pf2e/sim/foundry/pc_test.go` (new) — full contents:

```go
package foundry

import (
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func importTestPC(t *testing.T, name string) (*creature.Creature, *Report) {
	t.Helper()
	source, err := readActor(filepath.Join("testdata", "actors", name+".json"))
	if err != nil {
		t.Fatalf("readActor: %v", err)
	}
	layer, report, err := importPC(source)
	if err != nil {
		t.Fatalf("importPC: %v", err)
	}
	return layer, report
}

func checkRanks(t *testing.T, build *creature.Build, want map[string]string) {
	t.Helper()
	for _, key := range slices.Sorted(maps.Keys(want)) {
		if build.Proficiency[key] != want[key] {
			t.Errorf("proficiency %s = %q, want %q", key, build.Proficiency[key], want[key])
		}
	}
}

func TestImportPCBuildChoices(t *testing.T) {
	layer, _ := importTestPC(t, "kas")
	if layer.Name != "Kas Fringebramble" || layer.Kind != creature.KindPC || layer.Level != 14 {
		t.Errorf("header = %q %s %d, want Kas Fringebramble pc 14", layer.Name, layer.Kind, layer.Level)
	}
	if want := []string{"halfling", "humanoid"}; !slices.Equal(layer.Traits, want) {
		t.Errorf("Traits = %v, want %v", layer.Traits, want)
	}
	build := layer.Build
	if build.AncestryHP != 6 || build.ClassHP != 10 || build.Key != "str" || build.Speed != 25 || build.WeaponSpecialization != "normal" {
		t.Errorf("build header = %+v", build)
	}
	if build.Flaws != nil {
		t.Errorf("Flaws = %v, want none: alternate ancestry boosts replace the flaw", build.Flaws)
	}
	wantBoosts := map[string][]string{
		"ancestry":   {"str", "cha"},
		"background": {"con", "str"},
		"class":      {"str"},
		"level-1":    {"cha", "con", "str", "wis"},
		"level-5":    {"str", "con", "cha", "wis"},
		"level-10":   {"str", "con", "cha", "wis"},
	}
	if !reflect.DeepEqual(build.Boosts, wantBoosts) {
		t.Errorf("Boosts = %v, want %v", build.Boosts, wantBoosts)
	}
	checkRanks(t, build, map[string]string{
		"fortitude": "master", "reflex": "expert", "class-dc": "expert", "heavy": "expert",
		"simple": "master", "martial": "master", "unarmed": "master",
	})
	armor := build.Armor
	if armor == nil || armor.Name != "Full Plate" || armor.Category != "heavy" || armor.ACBonus != 6 ||
		armor.DexCap == nil || *armor.DexCap != 0 || armor.Potency != 2 || armor.Resilient != 2 {
		t.Errorf("Armor = %+v, want Full Plate heavy +6, dex cap 0, potency 2, resilient 2", armor)
	}
	wantFury := creature.Weapon{
		Name: "Protector's Fury", Category: "martial", Group: "hammer", Base: "maul",
		Dice: 1, Die: 12, Type: "bludgeoning", Potency: 2, Striking: 2,
		Traits: []string{"gnome", "trip", "versatile-p"},
	}
	if got := build.Weapons["protectors-fury"]; !reflect.DeepEqual(got, wantFury) {
		t.Errorf("protectors-fury = %+v, want %+v", got, wantFury)
	}
}

func TestImportPCWeaponsAndQuantities(t *testing.T) {
	layer, report := importTestPC(t, "tavish")
	build := layer.Build
	if build.Key != "int" || !slices.Equal(build.Flaws, []string{"cha"}) {
		t.Errorf("key %q flaws %v, want int and [cha]", build.Key, build.Flaws)
	}
	if want := []string{"con", "wis", "cha"}; !slices.Equal(build.Boosts["ancestry"], want) {
		t.Errorf("ancestry boosts = %v, want %v", build.Boosts["ancestry"], want)
	}
	if len(build.Weapons) != 12 {
		t.Errorf("weapons = %v, want 12 (two bombs have quantity 0)", slices.Sorted(maps.Keys(build.Weapons)))
	}
	wantFire := creature.Weapon{
		Name: "Alchemist's Fire (Greater)", Category: "martial", Group: "bomb", Base: "alchemical-bomb",
		Dice: 3, Die: 8, Type: "fire", ItemBonus: 2, Range: 20,
		Traits: []string{"alchemical", "bomb", "consumable", "fire", "infused", "splash"},
	}
	if got := build.Weapons["alchemists-fire-greater"]; !reflect.DeepEqual(got, wantFire) {
		t.Errorf("alchemists-fire-greater = %+v, want %+v", got, wantFire)
	}
	if _, found := build.Weapons["frost-vial-moderate"]; found {
		t.Error("frost-vial-moderate imported, want it skipped at quantity 0")
	}
	armor := build.Armor
	if armor == nil || armor.Category != "light" || armor.ACBonus != 2 || *armor.DexCap != 3 || armor.Potency != 2 || armor.Resilient != 1 {
		t.Errorf("Armor = %+v, want light +2, dex cap 3, potency 2, resilient 1", armor)
	}
	if build.Spellcasting != nil {
		t.Errorf("Spellcasting = %+v, want none: the innate entry's spells sit nowhere", build.Spellcasting)
	}
	for _, want := range []struct{ item, reason string }{
		{"Frost Vial (Moderate)", ReasonQuantity},
		{"Alchemist's Fire (Greater)", ReasonEffort3},
		{"Arcane Innate Spells", ReasonEmpty},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}

func TestImportPCSpellcasting(t *testing.T) {
	layer, report := importTestPC(t, "arabesque")
	build := layer.Build
	if want := (creature.Casting{Attribute: "cha", TopRank: 7}); build.Spellcasting == nil || *build.Spellcasting != want {
		t.Errorf("Spellcasting = %+v, want %+v", build.Spellcasting, want)
	}
	if !slices.Equal(build.Flaws, []string{"con"}) {
		t.Errorf("Flaws = %v, want [con]", build.Flaws)
	}
	checkRanks(t, build, map[string]string{
		"spellcasting": "expert", "unarmored": "expert", "fortitude": "expert", "reflex": "expert", "will": "expert", "simple": "expert",
	})
	if blade := build.Weapons["blink-blade"]; blade.Striking != 1 || blade.Potency != 2 || blade.Die != 4 {
		t.Errorf("blink-blade = %+v, want a +2 striking d4", blade)
	}
	if !hasSkip(report, "Desolation Locket", ReasonUnsupported) {
		t.Error("report lacks the Desolation Locket entry")
	}
}

func TestImportPCIgnoresSheetState(t *testing.T) {
	_, kasReport := importTestPC(t, "kas")
	_, arabesqueReport := importTestPC(t, "arabesque")
	for _, want := range []struct {
		report       *Report
		item, reason string
	}{
		{kasReport, "Slowed", ReasonState},
		{kasReport, "Spell Effect: Haste", ReasonState},
		{kasReport, "Bramblecrest (Destroyed)", ReasonEffort3},
		{kasReport, "Arcane Charges Spells", ReasonEmpty},
		{arabesqueReport, "Frightened", ReasonState},
		{arabesqueReport, "Persistent Damage", ReasonState},
	} {
		if !hasSkip(want.report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}
```

##### Task 5 — `internal/pf2e/sim/foundry/write_test.go`

Append after `TestImportFightNeedsExports`:

```go
func TestImportFightPlacesPCsInParty(t *testing.T) {
	fight := filepath.Join(t.TempDir(), "fall-of-luna")
	party := filepath.Join(t.TempDir(), "party")
	copyFile(t, filepath.Join("testdata", "actors", "kas.json"), filepath.Join(fight, "source", "actors", "kas.json"))

	reports, err := ImportFight(fight, party)
	if err != nil {
		t.Fatalf("ImportFight: %v", err)
	}
	if len(reports) != 1 || reports[0].Kind != creature.KindPC || reports[0].Folder != filepath.Join(party, "kas") {
		t.Fatalf("reports = %+v, want one pc report for party/kas", reports)
	}
	loaded, err := creature.Load(reports[0].Folder)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Build == nil || loaded.Strikes["protectors-fury"].Name != "Protector's Fury" {
		t.Errorf("loaded = %+v, want Kas with a build and Protector's Fury", loaded)
	}
}
```

##### Task 6 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-plan.md`

In the Work Breakdown Overview, tick Phase 3.

**Find:**
```
- [ ] **Phase 3 — PC build choices**
```
**Replace with:**
```
- [x] **Phase 3 — PC build choices**
```

##### Commit message

```
feat(E2): import PC build choices

- foundry: a character's import layer holds its build: boosts by
  source, ranks from the class and feats, worn armor, weapons, and the
  spellcasting attribute and top rank
- sheet state, quantity-0 weapons, persistent and splash damage,
  property runes and empty spell entries are reported, not written
- ImportFight writes PCs into the party folder
- testdata: copies of the three PC exports
```

### Phase 4 — PC rule elements and gate 1

**In plain words:** the importer reads the item bonuses, rank upgrades and resistances held in Foundry's rule elements, so the three PCs load with the fixture's level-14 numbers, and a test holds them there.

Before the first commit, check the fixture's status line (Shared Context, Before Phase 4).

#### Commit 5 — `feat(E2): import PC rule elements`

##### Task 1 — `internal/pf2e/sim/foundry/formula.go`

A small evaluator for rule-element values: numbers, `+ - * /`, brackets, `max`, `min`, `floor`, `ceil`, and the variables the importer passes. Anything else is `errUnsupportedFormula`, which the importer reports.

`internal/pf2e/sim/foundry/formula.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var errUnsupportedFormula = errors.New("unsupported formula")

func evaluateFormula(text string, variables map[string]float64) (int, error) {
	tokens, err := tokenize(text)
	if err != nil {
		return 0, err
	}
	parser := &formulaParser{tokens: tokens, variables: variables}
	value, err := parser.expression()
	if err != nil {
		return 0, fmt.Errorf("%w: %q: %v", errUnsupportedFormula, text, err)
	}
	if parser.position != len(tokens) {
		return 0, fmt.Errorf("%w: %q: unexpected %q", errUnsupportedFormula, text, parser.peek())
	}
	if value != math.Trunc(value) {
		return 0, fmt.Errorf("%w: %q gives %v, not a whole number", errUnsupportedFormula, text, value)
	}
	return int(value), nil
}

func tokenize(text string) ([]string, error) {
	var tokens []string
	characters := []rune(text)
	for index := 0; index < len(characters); {
		character := characters[index]
		start := index
		switch {
		case unicode.IsSpace(character):
			index++
			continue
		case strings.ContainsRune("+-*/(),", character):
			index++
		case unicode.IsDigit(character) || character == '.':
			for index < len(characters) && (unicode.IsDigit(characters[index]) || characters[index] == '.') {
				index++
			}
		case character == '@' || unicode.IsLetter(character):
			for index < len(characters) && isNameCharacter(characters[index]) {
				index++
			}
		default:
			return nil, fmt.Errorf("%w: %q: unexpected %q", errUnsupportedFormula, text, character)
		}
		tokens = append(tokens, string(characters[start:index]))
	}
	return tokens, nil
}

func isNameCharacter(character rune) bool {
	return character == '@' || character == '.' || character == '_' || unicode.IsLetter(character) || unicode.IsDigit(character)
}

type formulaParser struct {
	tokens    []string
	position  int
	variables map[string]float64
}

func (parser *formulaParser) peek() string {
	if parser.position < len(parser.tokens) {
		return parser.tokens[parser.position]
	}
	return ""
}

func (parser *formulaParser) next() string {
	token := parser.peek()
	parser.position++
	return token
}

func (parser *formulaParser) expression() (float64, error) {
	value, err := parser.term()
	for err == nil && (parser.peek() == "+" || parser.peek() == "-") {
		operator := parser.next()
		var right float64
		right, err = parser.term()
		if operator == "+" {
			value += right
		} else {
			value -= right
		}
	}
	return value, err
}

func (parser *formulaParser) term() (float64, error) {
	value, err := parser.factor()
	for err == nil && (parser.peek() == "*" || parser.peek() == "/") {
		operator := parser.next()
		var right float64
		right, err = parser.factor()
		switch {
		case err != nil:
		case operator == "*":
			value *= right
		case right == 0:
			err = errors.New("division by zero")
		default:
			value /= right
		}
	}
	return value, err
}

func (parser *formulaParser) factor() (float64, error) {
	token := parser.next()
	switch {
	case token == "":
		return 0, errors.New("unexpected end")
	case token == "-":
		value, err := parser.factor()
		return -value, err
	case token == "(":
		value, err := parser.expression()
		if err == nil && parser.next() != ")" {
			err = errors.New("missing )")
		}
		return value, err
	case strings.HasPrefix(token, "@"):
		value, found := parser.variables[token]
		if !found {
			return 0, fmt.Errorf("unknown variable %s", token)
		}
		return value, nil
	case unicode.IsDigit(rune(token[0])) || token[0] == '.':
		return strconv.ParseFloat(token, 64)
	default:
		return parser.call(token)
	}
}

func (parser *formulaParser) call(name string) (float64, error) {
	if parser.next() != "(" {
		return 0, fmt.Errorf("unexpected %s", name)
	}
	var arguments []float64
	for {
		value, err := parser.expression()
		if err != nil {
			return 0, err
		}
		arguments = append(arguments, value)
		separator := parser.next()
		if separator == ")" {
			break
		}
		if separator != "," {
			return 0, fmt.Errorf("unexpected %q in %s()", separator, name)
		}
	}
	switch {
	case name == "max":
		return slices.Max(arguments), nil
	case name == "min":
		return slices.Min(arguments), nil
	case name == "floor" && len(arguments) == 1:
		return math.Floor(arguments[0]), nil
	case name == "ceil" && len(arguments) == 1:
		return math.Ceil(arguments[0]), nil
	}
	return 0, fmt.Errorf("unknown function %s", name)
}
```

##### Task 2 — `internal/pf2e/sim/foundry/predicate.go`

A predicate holds when every term holds. At baseline the only options set are the actor's class, feats and heritage and the worn armor's category and group, so a term about a target, an action, a toggle or a condition is false, and its rule is reported as conditional. A comparison on a number the importer does not know is false, not an error.

`internal/pf2e/sim/foundry/predicate.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	errUnsupportedPredicate = errors.New("unsupported predicate")
	errUnsupportedTemplate  = errors.New("unsupported template")
	templatePattern         = regexp.MustCompile(`\{(\w+)\|([^}]+)\}`)
)

type rollOptions struct {
	flags   map[string]bool
	numbers map[string]float64
}

func (options rollOptions) holds(predicate []any) (bool, error) {
	for _, term := range predicate {
		held, err := options.term(term)
		if err != nil || !held {
			return false, err
		}
	}
	return true, nil
}

func (options rollOptions) term(term any) (bool, error) {
	switch typed := term.(type) {
	case string:
		return options.flags[typed], nil
	case map[string]any:
		if len(typed) == 1 {
			for operator, operand := range typed {
				return options.compound(operator, operand)
			}
		}
	}
	return false, fmt.Errorf("%w: %v", errUnsupportedPredicate, term)
}

func (options rollOptions) compound(operator string, operand any) (bool, error) {
	switch operator {
	case "not":
		held, err := options.term(operand)
		return err == nil && !held, err
	case "and", "or", "nor":
		terms, isList := operand.([]any)
		if !isList {
			break
		}
		anyHeld := false
		for _, term := range terms {
			held, err := options.term(term)
			if err != nil {
				return false, err
			}
			if operator == "and" && !held {
				return false, nil
			}
			anyHeld = anyHeld || held
		}
		switch operator {
		case "and":
			return true, nil
		case "or":
			return anyHeld, nil
		default:
			return !anyHeld, nil
		}
	case "gte", "gt", "lte", "lt", "eq":
		return options.comparison(operator, operand)
	}
	return false, fmt.Errorf("%w: %s %v", errUnsupportedPredicate, operator, operand)
}

func (options rollOptions) comparison(operator string, operand any) (bool, error) {
	pair, isList := operand.([]any)
	if !isList || len(pair) != 2 {
		return false, fmt.Errorf("%w: %s %v", errUnsupportedPredicate, operator, operand)
	}
	name, isName := pair[0].(string)
	right, isNumber := pair[1].(float64)
	if !isName || !isNumber {
		return false, fmt.Errorf("%w: %s %v", errUnsupportedPredicate, operator, operand)
	}
	left, found := options.numbers[name]
	if !found {
		return false, nil
	}
	switch operator {
	case "gte":
		return left >= right, nil
	case "gt":
		return left > right, nil
	case "lte":
		return left <= right, nil
	case "lt":
		return left < right, nil
	default:
		return left == right, nil
	}
}

func resolveTemplates(text string, current *item) (string, error) {
	var failure error
	resolved := templatePattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := templatePattern.FindStringSubmatch(match)
		value, err := templateValue(parts[1], parts[2], current)
		if err != nil {
			failure = err
			return match
		}
		return value
	})
	return resolved, failure
}

func templateValue(scope, path string, current *item) (string, error) {
	if scope == "item" && path == "_id" {
		return current.ID, nil
	}
	if flag, found := strings.CutPrefix(path, "flags.pf2e.rulesSelections."); scope == "item" && found {
		for _, rule := range current.System.Rules {
			if rule["key"] == "ChoiceSet" && rule["flag"] == flag {
				if selection, isText := rule["selection"].(string); isText {
					return selection, nil
				}
			}
		}
	}
	return "", fmt.Errorf("%w: {%s|%s}", errUnsupportedTemplate, scope, path)
}
```

##### Task 3 — `internal/pf2e/sim/foundry/rules.go`

The rule elements of Decision 5. Rules on condition, effect and spell items are not read (sheet state, and spells act only when cast). A physical item's rules count only when it is worn or held and, if it takes investing, invested. Weapon specialization's own rules are skipped: the feat's slug already set the choice (Decision 13).

`internal/pf2e/sim/foundry/rules.go` (new) — full contents:

```go
package foundry

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

var (
	statSelectors = map[string]string{
		"hp": "hp", "ac": "ac", "fortitude": "fortitude", "reflex": "reflex", "will": "will",
		"saving-throw": "saving-throw", "perception": "perception", "land-speed": "land-speed",
	}
	weaponStatSelectors = map[string]string{
		"attack": "attack", "strike-attack-roll": "attack", "weapon-damage": "damage", "strike-damage": "damage",
	}
	laterSelectors = []string{"initiative", "unarmed-attack", "unarmed-damage", "alchemical-bomb-damage", "melee-strike-damage"}
	bonusTypes     = map[string]string{"": "", "untyped": "", "item": "item", "status": "status", "circumstance": "circumstance"}
	laterRuleKeys  = []string{"AdjustDegreeOfSuccess", "DamageDice", "CriticalSpecialization", "AdjustStrike", "Strike", "Weakness", "Immunity"}
	unusedRuleKeys = []string{"Note", "GrantItem", "ChoiceSet", "Sense", "TokenLight", "TokenEffectIcon", "ItemAlteration", "CraftingAbility"}
	physicalTypes  = []string{"weapon", "armor", "equipment", "consumable", "shield", "backpack", "treasure", "ammo", "book"}
	unreadTypes    = []string{"condition", "effect", "spell"}
)

func (importer *pcImporter) applyRules() {
	options := importer.actorOptions()
	for index := range importer.source.Items {
		current := &importer.source.Items[index]
		if len(current.System.Rules) == 0 || slices.Contains(unreadTypes, current.Type) {
			continue
		}
		if _, handled := specializationSlugs[current.System.Slug]; handled {
			continue
		}
		if !rulesActive(current) {
			importer.report.skip(current.Name, fmt.Sprintf("%d rule elements", len(current.System.Rules)), ReasonInactive)
			continue
		}
		for _, rule := range current.System.Rules {
			importer.applyRule(current, rule, options)
		}
	}
}

func rulesActive(current *item) bool {
	if !slices.Contains(physicalTypes, current.Type) {
		return true
	}
	equipped := current.System.Equipped
	carried := equipped.CarryType == "worn" || equipped.CarryType == "held"
	return carried && (equipped.Invested == nil || *equipped.Invested)
}

func (importer *pcImporter) actorOptions() rollOptions {
	options := rollOptions{flags: map[string]bool{}, numbers: map[string]float64{}}
	for _, current := range importer.source.Items {
		switch current.Type {
		case "class":
			options.flags["class:"+current.System.Slug] = true
		case "feat":
			options.flags["feat:"+current.System.Slug] = true
		case "heritage":
			options.flags["self:heritage:"+current.System.Slug] = true
		}
	}
	if armor := importer.armor; armor != nil {
		options.flags["armor:category:"+armor.Category] = true
		options.flags["armor:group:"+armor.Group] = true
	}
	return options
}

func (importer *pcImporter) applyRule(current *item, rule map[string]any, options rollOptions) {
	key, _ := rule["key"].(string)
	switch {
	case key == "FlatModifier":
		importer.flatModifier(current, rule, options)
	case key == "Resistance":
		importer.resistance(current, rule, options)
	case key == "ActiveEffectLike":
		importer.activeEffectLike(current, rule, options)
	case key == "MartialProficiency":
		importer.martialProficiency(current, rule)
	case key == "AdjustModifier":
		importer.adjustModifier(current, rule)
	case key == "DexterityModifierCap":
		importer.dexterityCap(current, rule)
	case key == "RollOption":
		importer.rollOption(current, rule)
	case slices.Contains(laterRuleKeys, key):
		importer.report.skip(current.Name, key, ReasonEffort3)
	case slices.Contains(unusedRuleKeys, key):
		importer.report.unused(key)
	default:
		importer.report.skip(current.Name, key, ReasonUnsupported)
	}
}

func (importer *pcImporter) flatModifier(current *item, rule map[string]any, options rollOptions) {
	for _, field := range []string{"damageCategory", "damageType", "critical"} {
		if _, found := rule[field]; found {
			importer.report.skip(current.Name, "FlatModifier with "+field, ReasonEffort3)
			return
		}
	}
	selectors := textList(rule["selector"])
	if len(selectors) == 0 {
		importer.report.skip(current.Name, "FlatModifier without a selector", ReasonUnsupported)
		return
	}
	for _, selector := range selectors {
		importer.flatModifierFor(current, rule, selector, options)
	}
}

func (importer *pcImporter) flatModifierFor(current *item, rule map[string]any, selector string, options rollOptions) {
	what := "FlatModifier " + selector
	stat, isStat := statSelectors[selector]
	weaponStat, isWeaponStat := weaponStatSelectors[selector]
	switch {
	case slices.Contains(laterSelectors, selector):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	case !isStat && !isWeaponStat:
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	kind, _ := rule["type"].(string)
	bonusType, isBonusType := bonusTypes[kind]
	if !isBonusType {
		importer.report.skip(current.Name, what+" ("+kind+")", ReasonNotCombat)
		return
	}
	modifier := creature.Modifier{Selector: stat, Type: bonusType}
	predicate, _ := rule["predicate"].([]any)
	if isWeaponStat {
		traits, applies := importer.weaponTraits(predicate)
		if !applies {
			importer.report.skip(current.Name, what, ReasonConditional)
			return
		}
		modifier.Selector = weaponStat
		modifier.WeaponTraits = traits
	} else {
		held, err := options.holds(predicate)
		if err != nil {
			importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
			return
		}
		if !held {
			importer.report.skip(current.Name, what, ReasonConditional)
			return
		}
	}
	if text, isText := rule["value"].(string); isText && strings.TrimSpace(text) == "@actor.level" {
		modifier.PerLevel = 1
	} else {
		value, err := importer.number(rule["value"])
		if err != nil {
			importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
			return
		}
		modifier.Value = value
	}
	importer.build.Modifiers[uniqueID(current.Name, importer.modifierIDs)] = modifier
}

func (importer *pcImporter) weaponTraits(predicate []any) ([]string, bool) {
	var traits []string
	for _, term := range predicate {
		text, isText := term.(string)
		if !isText {
			return nil, false
		}
		trait := strings.TrimPrefix(text, "item:trait:")
		carried := false
		for _, weapon := range importer.build.Weapons {
			carried = carried || slices.Contains(weapon.Traits, trait)
		}
		if !carried {
			return nil, false
		}
		traits = append(traits, trait)
	}
	return traits, true
}

func (importer *pcImporter) resistance(current *item, rule map[string]any, options rollOptions) {
	kindText, _ := rule["type"].(string)
	kind, err := resolveTemplates(kindText, current)
	what := "Resistance " + kind
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	predicate, _ := rule["predicate"].([]any)
	held, err := options.holds(predicate)
	switch {
	case err != nil:
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	case !held:
		importer.report.skip(current.Name, what, ReasonConditional)
		return
	case !creature.IsDamageType(kind):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	}
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, what+": "+err.Error(), ReasonUnsupported)
		return
	}
	if value > importer.resist[kind].Value {
		importer.resist[kind] = creature.Adjustment{Value: value}
	}
}

func (importer *pcImporter) activeEffectLike(current *item, rule map[string]any, options rollOptions) {
	path, _ := rule["path"].(string)
	what := "ActiveEffectLike " + path
	key, isRank := rankKey(path)
	switch {
	case strings.HasPrefix(path, "system.attributes.dying."):
		importer.report.skip(current.Name, what, ReasonEffort3)
		return
	case !isRank:
		importer.report.skip(current.Name, what, ReasonNotCombat)
		return
	}
	if mode, _ := rule["mode"].(string); mode != "upgrade" {
		importer.report.skip(current.Name, what+" mode "+mode, ReasonUnsupported)
		return
	}
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
	rank, err := importer.number(rule["value"])
	if err != nil || rank < 0 || rank >= len(creature.RankWords) {
		importer.report.skip(current.Name, fmt.Sprintf("%s value %v", what, rule["value"]), ReasonUnsupported)
		return
	}
	importer.raise(key, rank)
}

func rankKey(path string) (string, bool) {
	inner, isSystem := strings.CutPrefix(path, "system.")
	inner, isRank := strings.CutSuffix(inner, ".rank")
	if !isSystem || !isRank {
		return "", false
	}
	switch {
	case inner == "perception":
		return "perception", true
	case inner == "proficiencies.spellcasting":
		return "spellcasting", true
	case strings.HasPrefix(inner, "saves."):
		key := strings.TrimPrefix(inner, "saves.")
		return key, creature.IsProficiencyKey(key)
	}
	for _, prefix := range []string{"proficiencies.attacks.", "proficiencies.defenses."} {
		if key, found := strings.CutPrefix(inner, prefix); found {
			return key, creature.IsProficiencyKey(key)
		}
	}
	return "", false
}

func (importer *pcImporter) martialProficiency(current *item, rule map[string]any) {
	definition, _ := rule["definition"].([]any)
	var group string
	var categories []string
	collect := func(text string) {
		if rest, found := strings.CutPrefix(text, "item:group:"); found {
			group = rest
		}
		if rest, found := strings.CutPrefix(text, "item:category:"); found {
			categories = append(categories, rest)
		}
	}
	for _, term := range definition {
		switch typed := term.(type) {
		case string:
			resolved, err := resolveTemplates(typed, current)
			if err != nil {
				importer.report.skip(current.Name, "MartialProficiency: "+err.Error(), ReasonUnsupported)
				return
			}
			collect(resolved)
		case map[string]any:
			alternatives, _ := typed["or"].([]any)
			for _, alternative := range alternatives {
				if text, isText := alternative.(string); isText {
					collect(text)
				}
			}
		}
	}
	what := fmt.Sprintf("MartialProficiency group %q, categories %s", group, strings.Join(categories, ", "))
	rank, err := importer.number(rule["value"])
	if group == "" || !slices.Contains(categories, "martial") || err != nil || rank < 0 || rank >= len(creature.RankWords) {
		importer.report.skip(current.Name, what, ReasonUnsupported)
		return
	}
	importer.raise(creature.WeaponGroupPrefix+group, rank)
}

func (importer *pcImporter) adjustModifier(current *item, rule map[string]any) {
	slug, _ := rule["slug"].(string)
	if suppress, _ := rule["suppress"].(bool); suppress && slug == "armor-speed-penalty" {
		importer.speedPenaltySuppressed = true
		return
	}
	selectors := append(textList(rule["selector"]), textList(rule["selectors"])...)
	importer.report.skip(current.Name, "AdjustModifier "+strings.Join(selectors, ", "), ReasonUnsupported)
}

func (importer *pcImporter) dexterityCap(current *item, rule map[string]any) {
	value, err := importer.number(rule["value"])
	if err != nil {
		importer.report.skip(current.Name, "DexterityModifierCap: "+err.Error(), ReasonUnsupported)
		return
	}
	importer.dexterityCaps = append(importer.dexterityCaps, value)
}

func (importer *pcImporter) rollOption(current *item, rule map[string]any) {
	if toggleable, found := rule["toggleable"]; !found || toggleable == false {
		importer.report.unused("RollOption")
		return
	}
	option, _ := rule["option"].(string)
	state := "off"
	if stored, _ := rule["value"].(bool); stored {
		state = "stored on"
	}
	importer.report.skip(current.Name, fmt.Sprintf("toggle %s (%s)", option, state), ReasonToggle)
}

func (importer *pcImporter) number(value any) (int, error) {
	switch typed := value.(type) {
	case float64:
		if typed != math.Trunc(typed) {
			return 0, fmt.Errorf("%w: %v is not a whole number", errUnsupportedFormula, typed)
		}
		return int(typed), nil
	case string:
		return evaluateFormula(typed, importer.variables())
	}
	return 0, fmt.Errorf("%w: value %v", errUnsupportedFormula, value)
}

func (importer *pcImporter) variables() map[string]float64 {
	variables := map[string]float64{"@actor.level": float64(importer.level)}
	if importer.armor != nil {
		variables["@armor.system.runes.potency"] = float64(importer.armor.Runes.Potency)
	}
	return variables
}

func textList(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		var texts []string
		for _, entry := range typed {
			if text, isText := entry.(string); isText {
				texts = append(texts, text)
			}
		}
		return texts
	}
	return nil
}
```

##### Task 4 — `internal/pf2e/sim/foundry/pc.go`

In `importPC`, apply the rule elements once every item is read: the predicates need the worn armor, and weapon-scoped modifiers need the weapons' traits.

**Find:**
```go
	if err := importer.spellcasting(); err != nil {
		return nil, nil, err
	}
	return importer.finish(), importer.report, nil
```
**Replace with:**
```go
	importer.applyRules()
	if err := importer.spellcasting(); err != nil {
		return nil, nil, err
	}
	return importer.finish(), importer.report, nil
```

##### Task 5 — `internal/pf2e/sim/foundry/formula_test.go`

`internal/pf2e/sim/foundry/formula_test.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"testing"
)

func TestEvaluateFormula(t *testing.T) {
	variables := map[string]float64{"@actor.level": 14, "@armor.system.runes.potency": 2}
	for _, test := range []struct {
		text string
		want int
	}{
		{"@actor.level", 14},
		{"2 + @armor.system.runes.potency", 4},
		{"max(1,floor(@actor.level/2))", 7},
		{"ceil(@actor.level/2) + 1", 8},
		{"floor((@actor.level+3)/2)", 8},
		{"(max(floor((@actor.level)/2),1))", 7},
		{"-3 * 2", -6},
	} {
		got, err := evaluateFormula(test.text, variables)
		if err != nil || got != test.want {
			t.Errorf("evaluateFormula(%q) = %d, %v; want %d", test.text, got, err, test.want)
		}
	}
	for _, text := range []string{
		"@actor.abilities.int.mod",
		"match(when(btwn(@actor.level, 1, 8), 1))",
		"7/2",
		"2 +",
		"1/0",
		"3 $ 4",
	} {
		if _, err := evaluateFormula(text, variables); !errors.Is(err, errUnsupportedFormula) {
			t.Errorf("evaluateFormula(%q) err = %v, want errUnsupportedFormula", text, err)
		}
	}
}
```

##### Task 6 — `internal/pf2e/sim/foundry/predicate_test.go`

`internal/pf2e/sim/foundry/predicate_test.go` (new) — full contents:

```go
package foundry

import (
	"errors"
	"testing"
)

func TestRollOptionsHold(t *testing.T) {
	options := rollOptions{
		flags:   map[string]bool{"feat:toughness": true, "armor:category:heavy": true, "armor:group:plate": true},
		numbers: map[string]float64{"item:proficiency:rank": 3},
	}
	for _, test := range []struct {
		name      string
		predicate []any
		want      bool
	}{
		{"empty", nil, true},
		{"all flags", []any{"armor:category:heavy", "armor:group:plate"}, true},
		{"missing flag", []any{"armor:category:heavy", "armor:group:composite"}, false},
		{"not", []any{map[string]any{"not": "feat:expanded-splash"}}, true},
		{"or", []any{map[string]any{"or": []any{"target:hidden", "feat:toughness"}}}, true},
		{"and", []any{map[string]any{"and": []any{"target:hidden", "feat:toughness"}}}, false},
		{"nor", []any{map[string]any{"nor": []any{"target:hidden", "target:undetected"}}}, true},
		{"gte held", []any{map[string]any{"gte": []any{"item:proficiency:rank", 2.0}}}, true},
		{"lt not held", []any{map[string]any{"lt": []any{"item:proficiency:rank", 3.0}}}, false},
		{"unknown number", []any{map[string]any{"gte": []any{"skill:intimidation:rank", 3.0}}}, false},
	} {
		got, err := options.holds(test.predicate)
		if err != nil || got != test.want {
			t.Errorf("%s: holds = %v, %v; want %v", test.name, got, err, test.want)
		}
	}
	if _, err := options.holds([]any{map[string]any{"xor": []any{}}}); !errors.Is(err, errUnsupportedPredicate) {
		t.Errorf("xor err = %v, want errUnsupportedPredicate", err)
	}
}

func TestResolveTemplates(t *testing.T) {
	legend := &item{ID: "abc123", System: itemCommon{Rules: []map[string]any{
		{"key": "ChoiceSet", "flag": "weaponLegend", "selection": "hammer"},
	}}}
	for _, test := range []struct{ text, want string }{
		{"item:group:{item|flags.pf2e.rulesSelections.weaponLegend}", "item:group:hammer"},
		{"{item|_id}-attack", "abc123-attack"},
		{"no template", "no template"},
	} {
		got, err := resolveTemplates(test.text, legend)
		if err != nil || got != test.want {
			t.Errorf("resolveTemplates(%q) = %q, %v; want %q", test.text, got, err, test.want)
		}
	}
	if _, err := resolveTemplates("{actor|flags.pf2e.x}", legend); !errors.Is(err, errUnsupportedTemplate) {
		t.Errorf("actor template err = %v, want errUnsupportedTemplate", err)
	}
}
```

##### Task 7 — `internal/pf2e/sim/foundry/rules_test.go`

The modifier tables are exact: every other `FlatModifier` on the three sheets is conditional, a skill, or a damage category.

`internal/pf2e/sim/foundry/rules_test.go` (new) — full contents:

```go
package foundry

import (
	"reflect"
	"slices"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

func TestImportPCRuleElementsKas(t *testing.T) {
	layer, report := importTestPC(t, "kas")
	build := layer.Build
	checkRanks(t, build, map[string]string{"will": "expert", "perception": "master", "weapon-group-hammer": "legendary"})
	if build.Armor.SpeedPenalty != 0 {
		t.Errorf("armor speed penalty = %d, want 0 (Unburdened Iron)", build.Armor.SpeedPenalty)
	}
	wantModifiers := map[string]creature.Modifier{
		"toughness":                   {Selector: "hp", PerLevel: 1},
		"mountains-stoutness":         {Selector: "hp", PerLevel: 1},
		"fleet":                       {Selector: "land-speed", Value: 5},
		"amulet-of-kinships-strength": {Selector: "will", Type: "item", Value: 2},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	if want := map[string]creature.Adjustment{"slashing": {Value: 4}}; !reflect.DeepEqual(layer.Defense.Resist, want) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, want)
	}
	for _, want := range []struct{ item, reason string }{
		{"Keen Eyes", ReasonConditional},
		{"Bravery", ReasonEffort3},
		{"Toughness", ReasonEffort3},
		{"Vicious Swing", ReasonToggle},
		{"Demon Mask", ReasonNotCombat},
		{"Weapon Legend (Hammer)", ReasonUnsupported},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
	if report.Unused["Note"] == 0 {
		t.Errorf("Unused = %v, want Note counted", report.Unused)
	}
}

func TestImportPCRuleElementsTavish(t *testing.T) {
	layer, report := importTestPC(t, "tavish")
	build := layer.Build
	checkRanks(t, build, map[string]string{"will": "expert", "simple": "expert", "weapon-base-alchemical-bomb": "expert", "class-dc": "expert"})
	wantModifiers := map[string]creature.Modifier{
		"alchemist-goggles":         {Selector: "attack", Type: "item", Value: 1, WeaponTraits: []string{"alchemical", "bomb"}},
		"boots-of-bounding-greater": {Selector: "land-speed", Type: "item", Value: 10},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	wantResist := map[string]creature.Adjustment{"poison": {Value: 7}, "electricity": {Value: 5}}
	if !reflect.DeepEqual(layer.Defense.Resist, wantResist) {
		t.Errorf("Resist = %+v, want %+v", layer.Defense.Resist, wantResist)
	}
	if !slices.Contains(report.Skipped, Skip{Item: "Sticky Bomb", What: "toggle sticky-bomb (stored on)", Reason: ReasonToggle}) {
		t.Error("report lacks Sticky Bomb as a toggle stored on")
	}
	if !hasSkip(report, "Torch", ReasonInactive) {
		t.Error("report lacks the stowed Torch as inactive")
	}
}

func TestImportPCRuleElementsArabesque(t *testing.T) {
	layer, report := importTestPC(t, "arabesque")
	build := layer.Build
	wantModifiers := map[string]creature.Modifier{
		"bands-of-force-greater":   {Selector: "ac", Type: "item", Value: 2},
		"bands-of-force-greater-2": {Selector: "saving-throw", Type: "item", Value: 2},
	}
	if !reflect.DeepEqual(build.Modifiers, wantModifiers) {
		t.Errorf("Modifiers = %+v, want %+v", build.Modifiers, wantModifiers)
	}
	if armor := build.Armor; armor == nil || armor.Category != "unarmored" || armor.DexCap == nil || *armor.DexCap != 5 {
		t.Errorf("Armor = %+v, want unarmored with dex cap 5", armor)
	}
	if layer.Defense.Resist != nil {
		t.Errorf("Resist = %+v, want none", layer.Defense.Resist)
	}
	for _, want := range []struct{ item, reason string }{
		{"Bands of Force", ReasonInactive},
		{"Diverting Vortex", ReasonConditional},
		{"Diverting Vortex", ReasonToggle},
	} {
		if !hasSkip(report, want.item, want.reason) {
			t.Errorf("report lacks %s with reason %q", want.item, want.reason)
		}
	}
}
```

##### Commit message

```
feat(E2): import PC rule elements

- foundry: FlatModifier on stat selectors becomes a build modifier;
  @actor.level becomes per_level, other formulas are worked out
- Resistance, rank upgrades (ActiveEffectLike), MartialProficiency for
  a weapon group, the armor speed penalty suppression and
  DexterityModifierCap feed the build
- predicates hold only on the actor's class, feats, heritage and armor,
  so targets, actions and toggles leave a rule conditional
- every other rule element is reported with its reason, or counted
```

#### Commit 6 — `test(E2): check gate 1 stat parity`

##### Task 1 — `internal/pf2e/sim/foundry/gate_test.go`

The numbers are the fixture's summary tables ([`F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`](./F.041.E2-pf2e-encounter-simulator-gate1-fixture.md), section 1) as Robert confirmed them (Shared Context, Before Phase 4). If he corrected any, use the corrected value here. Left out (Decision 1): Arabesque's class DC, the spell numbers of Kas's and Tavish's empty entries, and the Spellsap Grenade. A `classDC` of 0 means not gated.

`internal/pf2e/sim/foundry/gate_test.go` (new) — full contents:

```go
package foundry

import (
	"maps"
	"path/filepath"
	"testing"

	"github.com/therobertcrocker/gm-toolkit/internal/pf2e/sim/creature"
)

type gateStrike struct {
	id     string
	bonus  int
	damage string
	kind   string
}

type gateCase struct {
	pc           string
	attributes   map[string]int
	hp, ac       int
	saves        [3]int
	perception   int
	speed        int
	classDC      int
	spellcasting *creature.Spellcasting
	strikes      []gateStrike
	resist       map[string]int
}

var gateOne = []gateCase{
	{
		pc:         "kas",
		attributes: map[string]int{"str": 5, "dex": 0, "con": 4, "int": 0, "wis": 3, "cha": 4},
		hp:         230, ac: 36, saves: [3]int{26, 20, 23}, perception: 23, speed: 30, classDC: 33,
		strikes: []gateStrike{{"protectors-fury", 29, "3d12+9", "bludgeoning"}},
		resist:  map[string]int{"slashing": 4},
	},
	{
		pc:         "tavish",
		attributes: map[string]int{"str": 0, "dex": 4, "con": 3, "int": 4, "wis": 2, "cha": 3},
		hp:         164, ac: 35, saves: [3]int{24, 23, 21}, perception: 20, speed: 30, classDC: 32,
		strikes: []gateStrike{
			{"dagger", 22, "1d4+2", "piercing"},
			{"sling", 22, "1d6+2", "bludgeoning"},
			{"alchemists-fire-greater", 24, "3d8+2", "fire"},
			{"bottled-lightning-greater", 24, "3d6+2", "electricity"},
			{"tenderizer-grenade-greater", 24, "3d6+2", "acid"},
			{"nail-bomb-greater", 24, "6d4+2", "piercing"},
			{"star-grenade-greater", 24, "3d8+2", "fire"},
			{"sticky-algae-bomb-greater", 24, "3d8+2", "poison"},
			{"inflammation-flask-greater", 24, "3d6+2", "acid"},
			{"lodestone-bomb", 24, "3d4+2", "force"},
			{"crystal-shards-moderate", 23, "2d4+2", "piercing"},
		},
		resist: map[string]int{"poison": 7, "electricity": 5},
	},
	{
		pc:           "arabesque",
		attributes:   map[string]int{"str": 0, "dex": 4, "con": 1, "int": 4, "wis": 1, "cha": 5},
		hp:           104, ac: 34, saves: [3]int{21, 24, 21}, perception: 19, speed: 30,
		spellcasting: &creature.Spellcasting{Attack: 23, DC: 33, TopRank: 7},
		strikes:      []gateStrike{{"blink-blade", 24, "2d4+2", "piercing"}},
		resist:       map[string]int{},
	},
}

func TestGateOneStatParity(t *testing.T) {
	for _, want := range gateOne {
		t.Run(want.pc, func(t *testing.T) {
			layer, _ := importTestPC(t, want.pc)
			directory := filepath.Join(t.TempDir(), want.pc)
			if err := writeLayer(directory, "actors/"+want.pc+".json", layer); err != nil {
				t.Fatalf("writeLayer: %v", err)
			}
			loaded, err := creature.Load(directory)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !maps.Equal(loaded.Attributes, want.attributes) {
				t.Errorf("attributes = %v, want %v", loaded.Attributes, want.attributes)
			}
			defense := loaded.Defense
			for _, check := range []struct {
				name      string
				got, want int
			}{
				{"hp", defense.HP, want.hp},
				{"ac", defense.AC, want.ac},
				{"fortitude", defense.Fortitude, want.saves[0]},
				{"reflex", defense.Reflex, want.saves[1]},
				{"will", defense.Will, want.saves[2]},
				{"perception", loaded.Perception, want.perception},
				{"land speed", loaded.Speed.Land, want.speed},
			} {
				if check.got != check.want {
					t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
				}
			}
			if want.classDC > 0 && loaded.ClassDC != want.classDC {
				t.Errorf("class DC = %d, want %d", loaded.ClassDC, want.classDC)
			}
			switch {
			case want.spellcasting == nil && loaded.Spellcasting != nil:
				t.Errorf("spellcasting = %+v, want none", *loaded.Spellcasting)
			case want.spellcasting != nil && (loaded.Spellcasting == nil || *loaded.Spellcasting != *want.spellcasting):
				t.Errorf("spellcasting = %+v, want %+v", loaded.Spellcasting, *want.spellcasting)
			}
			for _, strike := range want.strikes {
				got, found := loaded.Strikes[strike.id]
				if !found {
					t.Errorf("no Strike %s", strike.id)
					continue
				}
				part := got.Damage[0]
				if got.Bonus != strike.bonus || part.Roll.String() != strike.damage || part.Type != strike.kind {
					t.Errorf("%s = %+d %s %s, want %+d %s %s", strike.id, got.Bonus, part.Roll, part.Type, strike.bonus, strike.damage, strike.kind)
				}
			}
			resist := map[string]int{}
			for kind, adjustment := range defense.Resist {
				resist[kind] = adjustment.Value
			}
			if !maps.Equal(resist, want.resist) {
				t.Errorf("resistances = %v, want %v", resist, want.resist)
			}
		})
	}
}
```

##### Task 2 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-plan.md`

In the Work Breakdown Overview, tick Phase 4.

**Find:**
```
- [ ] **Phase 4 — PC rule elements and gate 1**
```
**Replace with:**
```
- [x] **Phase 4 — PC rule elements and gate 1**
```

##### Commit message

```
test(E2): check gate 1 stat parity

- the three PCs, imported from the copied exports and loaded, match the
  fixture's level-14 numbers: attributes, HP, AC, saves, Perception,
  speed, class DC, spell numbers, Strikes and resistances
```

### Phase 5 — Hawk data and the wiki

**In plain words:** the Fall of Luna exports are imported into the Hawk campaign, Luna's hand layer shrinks to the one ability the import cannot write, and the wiki describes PC builds and the importer.

#### Commit 7 — `feat(E2): import Fall of Luna and trim Luna's hand layer`

##### Task 1 — `campaigns/hawk-claude/` (Claude's copy, gitignored)

Refresh the fight in Claude's copy and run the import there (Decision 11). Run from the repository root:

```bash
echo "==== refresh the fight in hawk-claude ===="
rm -rf campaigns/hawk-claude/encounters/fall-of-luna
mkdir -p campaigns/hawk-claude/encounters
cp -r campaigns/hawk/encounters/fall-of-luna campaigns/hawk-claude/encounters/
echo "==== build and import ===="
go build -o "$SCRATCH/gm-toolkit" ./cmd/gm-toolkit
"$SCRATCH/gm-toolkit" pf2e sim import fall-of-luna --campaign hawk-claude
```

Check the output before going on:

- The command exits 0. Five creatures print a stat block (`arabesque`, `kas` and `tavish` under `pf2e/party/`; `abyssal-luna` and `avatar-of-lilith` under `encounters/fall-of-luna/`), and the two loose items print `not imported`.
- The three PC stat blocks match `TestGateOneStatParity`'s table.
- Every skipped line names a reason from `report.go`. Note any line whose reason looks wrong in the Decision Record before the commit.

##### Task 2 — `campaigns/hawk/` (Robert runs it)

Ask Robert to run the import on his campaign:

```
! go run ./cmd/gm-toolkit pf2e sim import fall-of-luna --campaign hawk
```

It writes `campaigns/hawk/pf2e/party/{arabesque,kas,tavish}/import.toml` and `campaigns/hawk/encounters/fall-of-luna/{abyssal-luna,avatar-of-lilith}/import.toml`. Read each written file once before the commit.

##### Task 3 — `campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml`

The import layer now carries Luna's numbers and Strikes (Effort 1, Decision 3). The hand layer keeps only the ability that sentences describe.

`campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml` — full contents:

```toml
schema = 1

[ability.abyssal-frenzy]
name    = "Abyssal Frenzy"
actions = 2
steps   = [{ strike = "slash" }, { strike = "slash" }, { strike = "stab" }]
```

##### Task 4 — check Luna on Claude's copy

```bash
cp campaigns/hawk/encounters/fall-of-luna/abyssal-luna/hand.toml campaigns/hawk-claude/encounters/fall-of-luna/abyssal-luna/hand.toml
"$SCRATCH/gm-toolkit" pf2e sim show abyssal-luna --campaign hawk-claude
```

The stat block shows HP 300, AC 36, Slash +30 and Stab +28, `Abyssal Frenzy [2]  strike slash, strike slash, strike stab`, and `Layers import.toml + hand.toml`. `git status` lists the five new `import.toml` files and the changed `hand.toml` under `campaigns/hawk/`, and nothing under `campaigns/hawk-claude/`.

##### Commit message

```
feat(E2): import Fall of Luna and trim Luna's hand layer

- campaigns/hawk: import layers for the three PCs and both bosses,
  written by pf2e sim import
- Luna's hand layer keeps only Abyssal Frenzy; her numbers and Strikes
  now come from the import layer
```

#### Commit 8 — `docs(E2): document PC builds and the Foundry import`

##### Task 1 — `docs/architecture/pf2e/data-files.md`

Load `/gestalt:style-guide` first. Check each claim against the code before writing it (D-23); the prose is execution's (Decision 17).

Changes to existing text:

- The **Code** line adds `internal/pf2e/sim/foundry/`.
- Layers and their order, item 1: `import.toml` is written by `gm-toolkit pf2e sim import` (no longer planned).
- The file format: the top level also holds `class_dc`, and a `[spellcasting]` table holds `attack`, `dc` and `top_rank`. NPCs set them; a PC's come from its build.
- What the load checks: a creature with `[build]` is checked and computed before these checks (link to the new section).
- Dependencies: **Used by** adds `foundry`, which builds `creature.Creature` values and writes them as `import.toml`; the `import` command then loads each one. `creature` does not import `foundry`.

New section **PC builds**, after What the load checks. Required claims:

- A PC whose layers hold `[build]` stores choices; `Load` computes the numbers after the merge and before `prepare`.
- A table of the build keys: `ancestry_hp`, `class_hp`, `key`, `speed`, `flaws`, `weapon_specialization`, `[build.boosts]`, `[build.proficiency]`, `[build.armor]`, `[build.weapon.<id>]`, `[build.modifier.<id>]`, `[build.spellcasting]`, each with what it holds.
- Attributes: flaws first, then boosts in the order ancestry, background, class, then `level-<n>` by number. A boost to an attribute already at +4 is a partial; two partials make +1.
- Ranks are words, untrained to legendary. The proficiency bonus is level plus twice the rank, and 0 when untrained. A weapon's rank is the highest of its category, `weapon-group-<group>` and `weapon-base-<base>`; an advanced weapon skips the group rank (Decision Record 4).
- How each number is computed: HP, AC (no armor table means unarmored with no cap; armor's item bonus is `ac_bonus` plus potency), the saves (resilient as the item bonus), Perception, land speed (with `speed_penalty`), class DC, spell attack and DC, and a Strike's attack (Dex at range, the higher of Str and Dex with finesse, Str otherwise; potency or the weapon's own `item_bonus`) and damage (dice plus striking; Str in melee and thrown, half Str when propulsive, none at other range; specialization 2, 3, 4 by rank, or 4, 6, 8 when greater).
- Bonus stacking: per type, the highest bonus and the lowest penalty count; untyped values all add. A modifier with `weapon_traits` applies only to weapons that carry all of them. `per_level` is multiplied by the level.
- Choices only: the computed keys, and `ErrComputedKey` naming each file and key.

New section **Import**, after Finding a creature. Required claims:

- `gm-toolkit pf2e sim import <fight>` reads `encounters/<fight>/source/actors/*.json`. The file name is the creature's folder name; PCs go to `pf2e/party/<pc>/`, NPCs to `encounters/<fight>/<npc>/`. `source/items/*.json` are reported as not placed.
- Each `import.toml` starts with a header naming its export, and the import rewrites it whole.
- NPCs: numbers as stored, Strikes in the export's damage order, spell attack and DC from the first spellcasting entry. `all-damage` becomes `all`. What is skipped and why (abilities, spells, trait weaknesses, attack effects, damage categories).
- PCs: where each build choice comes from (class with class DC trained at base, ancestry with alternate boosts replacing boosts and flaws, background, level boosts, feat subfeatures, worn armor, weapons with quantity above 0, spellcasting entries that hold spells and name an attribute).
- A table of the rule elements handled and what each feeds. A predicate sees only the actor's class, feats, heritage and worn armor, so rules about targets, actions, conditions and toggles are reported as conditional. Every toggle is imported off.
- The report: each creature's merged stat block, then skipped items grouped by reason, then counts of rule kinds the sim never uses. A creature that fails to load makes the command fail after the rest print.

Key Decisions to add, word for word:

```markdown
- **A PC's import layer holds build choices, and load computes the
  numbers.** A level-up layer states what changed and the numbers follow;
  `show`, the import report and the sim share one computation. (`F.041`
  D-10; `F.041.E2` Decision 2.)
- **The importer resolves Foundry's rule elements; the format never holds a
  predicate or a formula.** Only `@actor.level` survives, as `per_level`.
  Foundry's rule-element language stays out of the format and the engine.
  (`F.041` D-1, D-4; `F.041.E2` Decision 2.)
- **A PC with a build takes choices only.** A layer that sets a computed
  number fails the load, so no stale override hides a level-up.
  (`F.041.E2` Decision 3.)
- **The import is a baseline: toggles off, sheet state dropped, conditional
  rules skipped.** Every run starts from the same character, and the report
  lists each thing left out. (`F.041` D-26; `F.041.E2` Decisions 5 and 6.)
```

##### Task 2 — `docs/architecture/pf2e/overview.md`

Required changes, each checked against the code:

- The **Code** line adds `internal/pf2e/sim/foundry/`.
- Shape: four of the five stages exist; the pipeline line reads `import → files → load → engine → runner (planned)`. The import bullet describes `pf2e sim import` and links to [data files](data-files.md#import). The files bullet says the import and the GM write them.
- The diagram comment marks only the runner as planned.
- Packages: a row for `internal/pf2e/sim/foundry` ("Reads Foundry exports and writes import layers, with a report"); the `cmd/gm-toolkit/pf2e` row names `show` and `import`.
- The dependency sentence: `foundry` reads `creature`, and the commands read `foundry` too.
- Commands: one sentence on `import`.

##### Task 3 — `docs/architecture/pf2e/engine.md`

In Not modelled yet, the weapon-traits line now says where the traits are shown.

**Find:**
```
- Weapon traits other than agile (deadly, fatal, versatile and the rest) are
  not read. Effort 2's import report names the ones the PCs carry.
```
**Replace with:**
```
- Weapon traits other than agile (deadly, fatal, versatile and the rest) are
  not read. The import copies every trait onto the Strike, and `show` and
  the import report print them.
```

##### Task 4 — `dev_journals/campaigns/initiatives/implementation/F.041-pf2e-encounter-simulator-overview-plan.md`

Bring the overview in line with Decisions 1, 4 and 5, and tick Effort 2.

**Find:**
```
| [ ] | 2 | Foundry import | NPC import (numbers and tags), PC import as choices, load-time PC stats, import report, `import` | 1 — stat parity |
```
**Replace with:**
```
| [x] | 2 | Foundry import | NPC import (numbers and Strikes), PC import as choices, load-time PC stats, import report, `import` | 1 — stat parity |
```

**Find:**
```
- **Robert supplies the gate 1 fixture:** the level-14 numbers for Kas, Tavish and Arabesque read off their Foundry sheets (HP, AC, saves, Perception, attack bonuses, class and spell DCs). The exports do not hold them (discovery, What the Foundry exports hold). The script's `PARTY[14]` row (`hawk_finale_sim.py:46`) is a starting point if still current.
```
**Replace with:**
```
- **The gate 1 fixture** is [`F.041.E2-pf2e-encounter-simulator-gate1-fixture.md`](./F.041.E2-pf2e-encounter-simulator-gate1-fixture.md): the level-14 numbers derived by hand from the exports, checked by Robert against the Foundry sheets (E2 plan, Decision 1).
```

**Find:**
```
- **Tags before rules.** Tags convert in Go with no approval, but a tag that names a rule the engine lacks (a `Slowed 2` condition link before Effort 3) has nowhere to go. Effort 2 reports it as skipped; Effort 3 extends the tag converter with each primitive and re-runs the import at its end. A re-import costs nothing, so this gap needs no ordering fix. (Efforts 2 and 3)
```
**Replace with:**
```
- **Tags before rules.** Tags convert in Go with no approval, but every tag on the two bosses names a rule the engine lacks until Effort 3, so the tag converter itself moved to Effort 3 (E2 plan, Decision 4). Effort 2 lists each ability in the import report; Effort 3 builds the converter with each primitive and re-runs the import at its end. A re-import costs nothing. (Efforts 2 and 3)
```

**Find:**
```
- NPC import: numbers as stored, and abilities whose mechanics are in tags, converted exactly in Go. Sentences are left for Effort 4; tags that name a rule not built yet are skipped with the reason (Findings, Tags before rules).
- PC import as build choices (D-10), and the load-time computation: core math and the first rule elements (`FlatModifier`, `AdjustDegreeOfSuccess`, `Resistance`, `DamageDice`). Mixed editions follow the sheet (D-5).
```
**Replace with:**
```
- NPC import: numbers and Strikes as stored. Each ability is listed in the import report, not converted: the tag converter moved to Effort 3 (E2 plan, Decision 4), and sentences are Effort 4's.
- PC import as build choices (D-10), and the load-time computation: core math and the rule elements the gate numbers need (`FlatModifier`, `Resistance`, rank upgrades, `MartialProficiency`). `AdjustDegreeOfSuccess` and `DamageDice` moved to Effort 3 (E2 plan, Decision 5). Mixed editions follow the sheet (D-5).
```

**Find:**
```
- Each rule extends the ability schema and, where a tag names it, the tag converter. The import re-runs at the end so the import layers pick up the new rules.
```
**Replace with:**
```
- Effort 3 builds the tag converter (moved from Effort 2) and the `AdjustDegreeOfSuccess` and `DamageDice` imports. Each rule extends the ability schema and, where a tag names it, the converter. The import re-runs at the end so the import layers pick up the new rules.
```

##### Task 5 — `dev_journals/campaigns/initiatives/implementation/F.041.E2-pf2e-encounter-simulator-plan.md`

In the Work Breakdown Overview, tick Phase 5.

**Find:**
```
- [ ] **Phase 5 — Hawk data and the wiki**
```
**Replace with:**
```
- [x] **Phase 5 — Hawk data and the wiki**
```

##### Commit message

```
docs(E2): document PC builds and the Foundry import

- data-files: PC builds and how load computes them, the import, and
  four Key Decisions
- overview: the import stage exists; foundry package and import command
- engine: where weapon traits are shown
- overview plan: Effort 2 and 3 scope follow the E2 plan's Decisions 1,
  4 and 5; Effort 2 ticked
```
